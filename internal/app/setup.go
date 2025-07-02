package app

import (
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/cache"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	weathh "github.com/velosypedno/genesis-weather-api/internal/handlers/weather"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	subr "github.com/velosypedno/genesis-weather-api/internal/repos/subscription"
	weathchain "github.com/velosypedno/genesis-weather-api/internal/repos/weather/chain"
	weathdecorator "github.com/velosypedno/genesis-weather-api/internal/repos/weather/decorator"
	weathprovider "github.com/velosypedno/genesis-weather-api/internal/repos/weather/provider"
	subsvc "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
	weathsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather"
	weathnotsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather_notification"
	"github.com/velosypedno/genesis-weather-api/pkg/cb"
)

const (
	freeWeathRName      = "weatherapi.com"
	tomorrowRName       = "tomorrow.io"
	visualCrossingRName = "visualcrossing.com"

	confirmSubTmplName    = "confirm_sub.html"
	weatherRequestTimeout = 5 * time.Second
	cacheTTL              = 5 * time.Minute

	// CB = CircuitBreaker
	weatherCBTimeout = 5 * time.Minute
	weatherCBLimit   = 10
	weatherCBRecover = 5
)

func (a *App) setupWeatherRepo() *weathdecorator.CacheDecorator {
	freeWeathR := weathprovider.NewFreeWeatherAPI(
		a.cfg.FreeWeather.Key,
		a.cfg.FreeWeather.URL,
		&http.Client{},
	)
	tomorrowWeathR := weathprovider.NewTomorrowAPI(
		a.cfg.TomorrowWeather.Key,
		a.cfg.TomorrowWeather.URL,
		&http.Client{},
	)
	vcWeathR := weathprovider.NewVisualCrossingAPI(
		a.cfg.VisualCrossing.Key,
		a.cfg.VisualCrossing.URL,
		&http.Client{},
	)

	logFreeWeathR := weathdecorator.NewLogDecorator(freeWeathR, freeWeathRName, a.reposLogger)
	logTomorrowR := weathdecorator.NewLogDecorator(tomorrowWeathR, tomorrowRName, a.reposLogger)
	logVcWeathR := weathdecorator.NewLogDecorator(vcWeathR, visualCrossingRName, a.reposLogger)

	breakerFreeWeathR := weathdecorator.NewBreakerDecorator(logFreeWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerTomorrowR := weathdecorator.NewBreakerDecorator(logTomorrowR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerVcWeathR := weathdecorator.NewBreakerDecorator(logVcWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))

	weathChain := weathchain.NewFirstFromChain(breakerFreeWeathR, breakerTomorrowR, breakerVcWeathR)

	redisBackend := cache.NewRedisBackend[domain.Weather](a.redisClient)
	cachedRepoChain := weathdecorator.NewCacheDecorator(weathChain, cacheTTL, redisBackend, a.metrics.weather)
	return cachedRepoChain
}

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()

	smtpEmailBackend := email.NewSMTPBackend(a.cfg.SMTP.Host, a.cfg.SMTP.Port, a.cfg.SMTP.User,
		a.cfg.SMTP.Pass, a.cfg.SMTP.EmailFrom)

	weatherRepo := a.setupWeatherRepo()
	weatherService := weathsvc.NewWeatherService(weatherRepo)

	subRepo := subr.NewDBRepo(a.db)
	confirmTmplPath := filepath.Join(a.cfg.Srv.TemplatesDir, confirmSubTmplName)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend, confirmTmplPath)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	api := router.Group("/api")
	{
		api.GET("/weather", weathh.NewWeatherGETHandler(weatherService, weatherRequestTimeout))
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subService))
	}

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}

func (a *App) setupCron() error {
	a.cron = cron.New()
	subRepo := subr.NewDBRepo(a.db)
	weatherRepoChain := a.setupWeatherRepo()
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := weathnotsvc.NewWeatherNotificationService(subRepo, weatherMailer, weatherRepoChain)

	_, err := a.cron.AddFunc("0 * * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqHourly)
	})
	if err != nil {
		return err
	}
	_, err = a.cron.AddFunc("0 7 * * *", func() {
		weatherMailerSrv.SendByFreq(domain.FreqDaily)
	})
	if err != nil {
		return err
	}
	return nil
}

package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	subh "github.com/velosypedno/genesis-weather-api/internal/handlers/subscription"
	weathh "github.com/velosypedno/genesis-weather-api/internal/handlers/weather"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	subr "github.com/velosypedno/genesis-weather-api/internal/repos/subscription"
	weathr "github.com/velosypedno/genesis-weather-api/internal/repos/weather"
	subsvc "github.com/velosypedno/genesis-weather-api/internal/services/subscription"
	weathsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather"
	weathnotsvc "github.com/velosypedno/genesis-weather-api/internal/services/weather_notification"
)

const (
	freeWeathRName     = "weatherapi.com"
	tomorrowWeathRName = "tomorrow.io"
)

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()

	smtpEmailBackend := email.NewSMTPBackend(a.cfg.SMTPHost, a.cfg.SMTPPort, a.cfg.SMTPUser, a.cfg.SMTPPass,
		a.cfg.EmailFrom)

	freeWeathR := weathr.NewWeatherAPIRepo(a.cfg.FreeWeatherAPIKey, &http.Client{})
	logFreeWeathR := weathr.NewLoggingWeatherRepo(freeWeathR, freeWeathRName, a.reposLogger)
	tomorrowWeathR := weathr.NewTomorrowAPIRepo(a.cfg.TomorrowWeatherAPIKey, &http.Client{})
	logTomorrowWeathR := weathr.NewLoggingWeatherRepo(tomorrowWeathR, tomorrowWeathRName, a.reposLogger)
	weatherRepoChain := weathr.NewWeatherRepoChain(logFreeWeathR, logTomorrowWeathR)
	weatherService := weathsvc.NewWeatherService(weatherRepoChain)

	subRepo := subr.NewSubscriptionDBRepo(a.db)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	api := router.Group("/api")
	{
		api.GET("/weather", weathh.NewWeatherGETHandler(weatherService))
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subService))
	}
	return router
}

func (a *App) setupCron() error {
	a.cron = cron.New()
	subRepo := subr.NewSubscriptionDBRepo(a.db)
	weatherRepo := weathr.NewWeatherAPIRepo(a.cfg.FreeWeatherAPIKey, &http.Client{})
	stdoutEmailBackend := email.NewStdoutBackend()
	weatherMailer := mailers.NewWeatherMailer(stdoutEmailBackend)
	weatherMailerSrv := weathnotsvc.NewWeatherNotificationService(subRepo, weatherMailer, weatherRepo)

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

package app

import (
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/cache"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/email"
	subgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/subscription"
	weathgrpc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/grpc/weather"
	subhttp "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/http/subscription"
	weathhttp "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/handlers/http/weather"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/mailers"
	subr "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/subscription"
	weathchain "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/chain"
	weathdecorator "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/decorator"
	weathprovider "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/provider"
	subsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	weathsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/weather"
	weathnotsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/weather_notification"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"
)

const (
	freeWeatherName    = "weatherapi.com"
	tomorrowIOName     = "tomorrow.io"
	visualCrossingName = "visualcrossing.com"

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
		weathprovider.APICfg{APIKey: a.cfg.FreeWeather.Key, APIURL: a.cfg.FreeWeather.URL},
		&http.Client{},
	)
	tomorrowWeathR := weathprovider.NewTomorrowAPI(
		weathprovider.APICfg{APIKey: a.cfg.TomorrowWeather.Key, APIURL: a.cfg.TomorrowWeather.URL},
		&http.Client{},
	)
	vcWeathR := weathprovider.NewVisualCrossingAPI(
		weathprovider.APICfg{APIKey: a.cfg.VisualCrossing.Key, APIURL: a.cfg.VisualCrossing.URL},
		&http.Client{},
	)

	logFreeWeathR := weathdecorator.NewLogDecorator(freeWeathR, freeWeatherName, a.reposLogger)
	logTomorrowR := weathdecorator.NewLogDecorator(tomorrowWeathR, tomorrowIOName, a.reposLogger)
	logVcWeathR := weathdecorator.NewLogDecorator(vcWeathR, visualCrossingName, a.reposLogger)

	breakerFreeWeathR := weathdecorator.NewBreakerDecorator(logFreeWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerTomorrowR := weathdecorator.NewBreakerDecorator(logTomorrowR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerVcWeathR := weathdecorator.NewBreakerDecorator(logVcWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))

	weathChain := weathchain.NewProvidersFallbackChain(breakerFreeWeathR, breakerTomorrowR, breakerVcWeathR)

	redisBackend := cache.NewRedisCacheClient[domain.Weather](a.redisClient, cacheTTL)
	cachedRepoChain := weathdecorator.NewCacheDecorator(weathChain, redisBackend, a.metrics.weather)
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
		api.GET("/weather", weathhttp.NewWeatherGETHandler(weatherService, weatherRequestTimeout))
		api.POST("/subscribe", subhttp.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subhttp.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subhttp.NewUnsubscribeGETHandler(subService))
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

func (a *App) setupGRPCListener() (net.Listener, error) {
	const port = 50100
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		return nil, err
	}
	return lis, nil
}

func (a *App) setupGRPCServer() *grpc.Server {
	weatherRepo := a.setupWeatherRepo()
	weatherService := weathsvc.NewWeatherService(weatherRepo)

	smtpEmailBackend := email.NewSMTPBackend(a.cfg.SMTP.Host, a.cfg.SMTP.Port, a.cfg.SMTP.User,
		a.cfg.SMTP.Pass, a.cfg.SMTP.EmailFrom)
	subRepo := subr.NewDBRepo(a.db)
	confirmTmplPath := filepath.Join(a.cfg.Srv.TemplatesDir, confirmSubTmplName)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend, confirmTmplPath)
	subService := subsvc.NewSubscriptionService(subRepo, subMailer)

	grpcServer := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcServer, subgrpc.NewSubGRPCServer(subService))
	pb.RegisterWeatherServiceServer(grpcServer, weathgrpc.NewWeatherGRPCServer(weatherService, weatherRequestTimeout))
	return grpcServer
}

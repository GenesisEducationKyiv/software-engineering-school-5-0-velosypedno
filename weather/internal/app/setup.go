package app

import (
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cache"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	grpch "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/handlers/grpc"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/chain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/decorator"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/repos/provider"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

const (
	freeWeatherName    = "weatherapi.com"
	tomorrowIOName     = "tomorrow.io"
	visualCrossingName = "visualcrossing.com"

	confirmSubTmplName    = "confirm_sub.html"
	weatherRequestTimeout = 10 * time.Second
	cacheTTL              = 5 * time.Minute

	// CB = CircuitBreaker
	weatherCBTimeout = 5 * time.Minute
	weatherCBLimit   = 10
	weatherCBRecover = 5
)

func (a *App) setupWeatherRepo() *decorator.CacheDecorator {
	_ = freeWeatherName
	_ = tomorrowIOName
	_ = visualCrossingName

	freeWeathR := provider.NewFreeWeatherAPI(
		provider.APICfg{APIKey: a.cfg.FreeWeather.Key, APIURL: a.cfg.FreeWeather.URL},
		&http.Client{},
	)
	tomorrowWeathR := provider.NewTomorrowAPI(
		provider.APICfg{APIKey: a.cfg.TomorrowWeather.Key, APIURL: a.cfg.TomorrowWeather.URL},
		&http.Client{},
	)
	vcWeathR := provider.NewVisualCrossingAPI(
		provider.APICfg{APIKey: a.cfg.VisualCrossing.Key, APIURL: a.cfg.VisualCrossing.URL},
		&http.Client{},
	)

	breakerFreeWeathR := decorator.NewBreakerDecorator(freeWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerTomorrowR := decorator.NewBreakerDecorator(tomorrowWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerVcWeathR := decorator.NewBreakerDecorator(vcWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))

	weathChain := chain.NewProvidersFallbackChain(breakerFreeWeathR, breakerTomorrowR, breakerVcWeathR)

	redisBackend := cache.NewRedisCacheClient[domain.Weather](a.redisClient, cacheTTL)
	cachedRepoChain := decorator.NewCacheDecorator(weathChain, redisBackend, a.metrics.weather)
	return cachedRepoChain
}

func (a *App) setupRouter() *gin.Engine {
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	return router
}

func (a *App) setupGRPCSrv() *grpc.Server {
	grpcServer := grpc.NewServer()
	logger := a.logFactory.ForPackage("handlers/grpc")
	weatherRepo := a.setupWeatherRepo()
	weatherService := services.NewWeatherService(weatherRepo)
	pb.RegisterWeatherServiceServer(grpcServer, grpch.NewWeatherGRPCServer(weatherService, weatherRequestTimeout, logger))
	return grpcServer
}

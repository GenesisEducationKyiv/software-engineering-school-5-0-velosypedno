package app

import (
	"net/http"
	"time"

	subh "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/handlers"
	subsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/services"
	weathh "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/weather/handlers"
	weathsvc "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/weather/services"
	"github.com/gin-gonic/gin"
)

const readTimeout = 15 * time.Second
const weatherRequestTimeout = 5 * time.Second

func (a *App) setupHTTPServer() *http.Server {
	router := gin.Default()
	router.Use(a.metrics.handler.Middleware())

	weathGRPCClientLogger := a.logFactory.ForPackage("weather/services/grpc_adapter")
	weathService := weathsvc.NewGRPCAdapter(weathGRPCClientLogger, a.weathGRPCClient)
	weathHandlerLogger := a.logFactory.ForPackage("weather/handlers")

	subGRPCClientLogger := a.logFactory.ForPackage("subscription/services/grpc_adapter")
	subService := subsvc.NewGRPCAdapter(subGRPCClientLogger, a.subGRPCClient)
	subHandlersLogger := a.logFactory.ForPackage("subscription/handlers")

	api := router.Group("/api")
	{
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subHandlersLogger, subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subHandlersLogger, subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subHandlersLogger, subService))
		api.GET("/weather", weathh.NewWeatherGETHandler(weathHandlerLogger, weathService, weatherRequestTimeout))
	}
	httpSrv := http.Server{
		Addr:        ":" + a.cfg.APIGatewayPort,
		Handler:     router,
		ReadTimeout: readTimeout,
	}

	return &httpSrv
}

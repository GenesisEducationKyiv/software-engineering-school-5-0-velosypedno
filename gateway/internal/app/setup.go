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

	subService := subsvc.NewGRPCAdapter(a.subSvc)
	weathService := weathsvc.NewGRPCAdapter(a.weathSvc)

	api := router.Group("/api")
	{
		api.POST("/subscribe", subh.NewSubscribePOSTHandler(subService))
		api.GET("/confirm/:token", subh.NewConfirmGETHandler(subService))
		api.GET("/unsubscribe/:token", subh.NewUnsubscribeGETHandler(subService))
		api.GET("/weather", weathh.NewWeatherGETHandler(weathService, weatherRequestTimeout))
	}
	httpSrv := http.Server{
		Addr:        ":" + a.cfg.APIGatewayPort,
		Handler:     router,
		ReadTimeout: readTimeout,
	}

	return &httpSrv
}

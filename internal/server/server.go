package server

import (
	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
)

func SetupRoutes(h *ioc.Handlers) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api")
	{
		api.GET("/weather", h.WeatherGETHandler)
		api.POST("/subscribe", h.SubscribePOSTHandler)
		api.GET("/confirm/:token", h.ConfirmGETHandler)
		api.GET("/unsubscribe/:token", h.UnsubscribeGETHandler)
	}
	return router
}

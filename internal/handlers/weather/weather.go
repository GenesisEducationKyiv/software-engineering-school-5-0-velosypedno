package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/models"
	weathsrv "github.com/velosypedno/genesis-weather-api/internal/services/weather"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (models.Weather, error)
}

func NewWeatherGETHandler(service weatherService) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weather, err := service.GetCurrent(c.Request.Context(), city)
		if errors.Is(err, weathsrv.ErrCityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		if errors.Is(err, weathsrv.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}
		c.JSON(http.StatusOK, weather)
	}
}

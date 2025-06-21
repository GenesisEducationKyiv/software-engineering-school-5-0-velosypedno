package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
	weathsrv "github.com/velosypedno/genesis-weather-api/internal/services/weather"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherResp struct {
	temperature float64
	humidity    float64
	description string
}

func NewWeatherGETHandler(service weatherService) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weatherEnt, err := service.GetCurrent(c.Request.Context(), city)
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
		weatherResp := weatherResp{
			temperature: weatherEnt.Temperature,
			humidity:    weatherEnt.Humidity,
			description: weatherEnt.Description,
		}
		c.JSON(http.StatusOK, weatherResp)
	}
}

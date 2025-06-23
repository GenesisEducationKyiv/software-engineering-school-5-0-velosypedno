package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherResp struct {
	temperature float64
	humidity    float64
	description string
}

func NewWeatherGETHandler(service weatherService, requestTimeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		ctxWithTimeout, cancel := context.WithTimeout(c.Request.Context(), requestTimeout)
		defer cancel()
		weatherEnt, err := service.GetCurrent(ctxWithTimeout, city)
		if errors.Is(err, domain.ErrCityNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		if errors.Is(err, domain.ErrInternal) {
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

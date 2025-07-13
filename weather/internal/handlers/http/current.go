package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"github.com/gin-gonic/gin"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherResp struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
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
		if errors.Is(err, domain.ErrWeatherUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sources are unavailable"})
			return
		}
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}
		weatherResp := weatherResp{
			Temperature: weatherEnt.Temperature,
			Humidity:    weatherEnt.Humidity,
			Description: weatherEnt.Description,
		}
		c.JSON(http.StatusOK, weatherResp)
	}
}

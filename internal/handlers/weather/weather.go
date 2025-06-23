package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherResp struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

func NewWeatherGETHandler(service weatherService) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weatherEnt, err := service.GetCurrent(c.Request.Context(), city)
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
			Temperature: weatherEnt.Temperature,
			Humidity:    weatherEnt.Humidity,
			Description: weatherEnt.Description,
		}
		c.JSON(http.StatusOK, weatherResp)
	}
}

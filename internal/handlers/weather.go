package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type WeatherRepo interface {
	GetCurrentWeather(ctx context.Context, city string) (models.Weather, error)
}

func NewWeatherHandler(repo WeatherRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weather, err := repo.GetCurrentWeather(c.Request.Context(), city)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		c.JSON(http.StatusOK, weather)
	}
}

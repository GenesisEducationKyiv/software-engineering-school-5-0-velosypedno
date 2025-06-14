package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

type WeatherRepo interface {
	GetCurrent(ctx context.Context, city string) (models.Weather, error)
}

func NewWeatherGETHandler(repo WeatherRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weather, err := repo.GetCurrent(c.Request.Context(), city)
		if errors.Is(err, services.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		if errors.Is(err, services.ErrInternal) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}
		c.JSON(http.StatusOK, weather)
	}
}

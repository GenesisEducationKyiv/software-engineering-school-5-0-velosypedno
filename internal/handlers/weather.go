package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

type WeatherRepo interface {
	GetCurrentWeather(ctx context.Context, city string) (models.Weather, error)
}

func NewWeatherGETHandler(repo WeatherRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		weather, err := repo.GetCurrentWeather(c.Request.Context(), city)
		if err != nil {
			if err == repos.ErrCityNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
				return
			}
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}
		c.JSON(http.StatusOK, weather)
	}
}

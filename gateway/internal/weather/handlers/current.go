package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/weather/domain"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherResp struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Description string  `json:"description"`
}

func NewWeatherGETHandler(
	logger *zap.Logger,
	service weatherService,
	requestTimeout time.Duration,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		city := c.Query("city")
		if city == "" {
			logger.Warn("City not provided in query")
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		ctxWithTimeout, cancel := context.WithTimeout(c.Request.Context(), requestTimeout)
		defer cancel()

		logger.Info("Fetching weather",
			zap.String("city", city),
		)

		weatherEnt, err := service.GetCurrent(ctxWithTimeout, city)

		switch {
		case errors.Is(err, domain.ErrCityNotFound):
			logger.Warn("City not found", zap.String("city", city))
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		case errors.Is(err, domain.ErrInternal):
			logger.Error("Internal error during weather fetch", zap.String("city", city), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		case errors.Is(err, domain.ErrWeatherUnavailable):
			logger.Warn("Weather sources unavailable", zap.String("city", city))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "sources are unavailable"})
			return
		case err != nil:
			logger.Error("Unexpected error during weather fetch", zap.String("city", city), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get weather for given city"})
			return
		}

		resp := weatherResp{
			Temperature: weatherEnt.Temperature,
			Humidity:    weatherEnt.Humidity,
			Description: weatherEnt.Description,
		}

		logger.Info("Successfully fetched weather",
			zap.String("city", city),
		)

		c.JSON(http.StatusOK, resp)
	}
}

package services

import (
	"context"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type WeatherRepo interface {
	GetCurrent(ctx context.Context, city string) (models.Weather, error)
}

type WeatherService struct {
	repo WeatherRepo
}

func NewWeatherService(repo WeatherRepo) *WeatherService {
	return &WeatherService{repo: repo}
}

func (s *WeatherService) GetCurrent(ctx context.Context, city string) (models.Weather, error) {
	return s.repo.GetCurrent(ctx, city)
}

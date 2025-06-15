package services

import (
	"context"
	"errors"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/models"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

var (
	ErrCityNotFound = errors.New("city not found")
	ErrInternal     = errors.New("internal error")
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
	w, err := s.repo.GetCurrent(ctx, city)
	if errors.Is(err, repos.ErrCityNotFound) {
		return models.Weather{}, ErrCityNotFound
	} else if err != nil {
		log.Println(err)
		return models.Weather{}, ErrInternal
	}
	return w, nil
}

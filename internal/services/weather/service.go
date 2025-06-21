package services

import (
	"context"
	"errors"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

var (
	ErrCityNotFound = errors.New("city not found")
	ErrInternal     = errors.New("internal error")
)

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type WeatherService struct {
	repo weatherRepo
}

func NewWeatherService(repo weatherRepo) *WeatherService {
	return &WeatherService{repo: repo}
}

func (s *WeatherService) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	w, err := s.repo.GetCurrent(ctx, city)
	if errors.Is(err, repos.ErrCityNotFound) {
		return domain.Weather{}, ErrCityNotFound
	} else if err != nil {
		log.Println(err)
		return domain.Weather{}, ErrInternal
	}
	return w, nil
}

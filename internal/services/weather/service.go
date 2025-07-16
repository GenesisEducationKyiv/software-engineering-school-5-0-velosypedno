package services

import (
	"context"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
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
	if err != nil {
		return w, fmt.Errorf("weather service: %w", err)
	}
	return w, nil
}

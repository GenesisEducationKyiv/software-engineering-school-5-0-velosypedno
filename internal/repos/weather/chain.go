package repos

import (
	"context"
	"fmt"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type WeatherRepoChain struct {
	Repos []weatherRepo
}

func NewWeatherRepoChain(repos ...weatherRepo) *WeatherRepoChain {
	return &WeatherRepoChain{Repos: repos}
}

func (c *WeatherRepoChain) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var lastError error
	for _, repo := range c.Repos {
		weather, err := repo.GetCurrent(ctx, city)
		if err == nil {
			return weather, nil
		}
		err = fmt.Errorf("chain: %w", err)
		log.Println(err)
		lastError = err

	}
	return domain.Weather{}, lastError
}

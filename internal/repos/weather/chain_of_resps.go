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

type Chain struct {
	Repos []weatherRepo
}

func NewChain(repos ...weatherRepo) *Chain {
	return &Chain{Repos: repos}
}

func (c *Chain) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	var lastError error
	for _, repo := range c.Repos {
		weather, err := repo.GetCurrent(ctx, city)
		if err != nil {
			err = fmt.Errorf("chain: %w", err)
			log.Println(err)
			lastError = err
			continue
		}
		return weather, nil
	}
	return domain.Weather{}, lastError
}

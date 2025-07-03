package chain

import (
	"context"
	"fmt"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type FirstFromChain struct {
	Repos []weatherRepo
}

func NewFirstFromChain(repos ...weatherRepo) *FirstFromChain {
	return &FirstFromChain{Repos: repos}
}

func (c *FirstFromChain) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
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

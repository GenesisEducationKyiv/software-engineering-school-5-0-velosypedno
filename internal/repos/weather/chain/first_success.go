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
		if err == nil {
			return weather, nil
		}
		err = fmt.Errorf("chain: %w", err)
		log.Println(err)
		lastError = err
	}
	return domain.Weather{}, lastError
}

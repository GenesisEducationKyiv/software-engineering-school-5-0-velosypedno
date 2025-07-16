package decorator

import (
	"context"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
)

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type LogDecorator struct {
	Inner    weatherRepo
	RepoName string
	Logger   *log.Logger
}

func NewLogDecorator(inner weatherRepo, repoName string, logger *log.Logger) *LogDecorator {
	return &LogDecorator{Inner: inner, RepoName: repoName, Logger: logger}
}

func (d *LogDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	weather, err := d.Inner.GetCurrent(ctx, city)
	if err != nil {
		d.Logger.Printf("%s - error for %s: %v\n", d.RepoName, city, err)
		return domain.Weather{}, err
	}
	d.Logger.Printf("%s - success for %s: %v\n", d.RepoName, city, weather)
	return weather, nil
}

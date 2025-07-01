package repos

import (
	"context"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type LogDecorator struct {
	Inner    weatherRepo
	RepoName string
	Logger   *log.Logger
}

func NewLogDecorator(inner weatherRepo, repoName string, logger *log.Logger) *LogDecorator {
	return &LogDecorator{Inner: inner, RepoName: repoName, Logger: logger}
}

func (f *LogDecorator) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	weather, err := f.Inner.GetCurrent(ctx, city)
	if err != nil {
		f.Logger.Printf("%s - error for %s: %v\n", f.RepoName, city, err)
		return domain.Weather{}, err
	}
	f.Logger.Printf("%s - success for %s: %v\n", f.RepoName, city, weather)
	return weather, nil
}

package repos

import (
	"context"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/domain"
)

type LoggingWeatherRepo struct {
	Inner    weatherRepo
	RepoName string
	Logger   *log.Logger
}

func NewLoggingWeatherRepo(inner weatherRepo, repoName string, logger *log.Logger) *LoggingWeatherRepo {
	return &LoggingWeatherRepo{Inner: inner, RepoName: repoName, Logger: logger}
}

func (f *LoggingWeatherRepo) GetCurrent(ctx context.Context, city string) (domain.Weather, error) {
	weather, err := f.Inner.GetCurrent(ctx, city)
	if err != nil {
		f.Logger.Printf("%s - error for %s: %v\n", f.RepoName, city, err)
		return domain.Weather{}, err
	}
	f.Logger.Printf("%s - success for %s: %v\n", f.RepoName, city, weather)
	return weather, nil
}

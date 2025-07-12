package app

import (
	"context"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subservice "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	weathservice "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/weather"
	weathnotify "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/weather_notification"
	"github.com/google/uuid"
)

type subscriptionService interface {
	Activate(token uuid.UUID) error
	Unsubscribe(token uuid.UUID) error
	Subscribe(subInput subservice.SubscriptionInput) error
}

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type weatherNotificationService interface {
	SendByFreq(freq domain.Frequency)
}

type BusinessContainer struct {
	SubService         subscriptionService
	WeathService       weatherService
	WeathNotifyService weatherNotificationService
}

func NewBusinessContainer(infraContainer *InfrastructureContainer) (*BusinessContainer, error) {
	weathService := weathservice.NewWeatherService(infraContainer.WeatherRepo)
	subService := subservice.NewSubscriptionService(
		infraContainer.SubRepo,
		infraContainer.SubMailer,
	)
	weathNotifyService := weathnotify.NewWeatherNotificationService(
		infraContainer.SubRepo,
		infraContainer.WeatherMailer,
		infraContainer.WeatherRepo,
	)

	return &BusinessContainer{
		SubService:         subService,
		WeathService:       weathService,
		WeathNotifyService: weathNotifyService,
	}, nil
}

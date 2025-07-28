package app

import (
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	subservice "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/services/subscription"
	weathnotify "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/services/weather_notification"
	"github.com/google/uuid"
)

type subscriptionService interface {
	Activate(token uuid.UUID) error
	Unsubscribe(token uuid.UUID) error
	Subscribe(subInput subservice.SubscriptionInput) error
}

type weatherNotificationService interface {
	SendByFreq(freq domain.Frequency)
}

type BusinessContainer struct {
	SubService         subscriptionService
	WeathNotifyService weatherNotificationService
}

func NewBusinessContainer(infraContainer *InfrastructureContainer) (*BusinessContainer, error) {
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
		WeathNotifyService: weathNotifyService,
	}, nil
}

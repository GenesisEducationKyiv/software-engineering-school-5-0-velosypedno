package notifiers

import (
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
)

type weatherNotifyCommandProducer interface {
	Produce(sub domain.Subscription, weath domain.Weather) error
}

type WeatherNotifyCommandNotifier struct {
	producer weatherNotifyCommandProducer
}

func NewWeatherNotifyCommandNotifier(producer weatherNotifyCommandProducer) *WeatherNotifyCommandNotifier {
	return &WeatherNotifyCommandNotifier{
		producer: producer,
	}
}

func (m *WeatherNotifyCommandNotifier) SendCurrent(subscription domain.Subscription, weather domain.Weather) error {
	err := m.producer.Produce(subscription, weather)
	if err != nil {
		return fmt.Errorf("weather notify command notifier: %w", err)
	}
	return nil
}

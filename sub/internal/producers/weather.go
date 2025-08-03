package producers

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type WeatherNotifyCommandProducer struct {
	logger *zap.Logger
	ch     *amqp.Channel
}

func NewWeatherNotifyCommandProducer(logger *zap.Logger, ch *amqp.Channel) *WeatherNotifyCommandProducer {
	return &WeatherNotifyCommandProducer{
		logger: logger.With(zap.String("producer", "WeatherNotifyCommandProducer")),
		ch:     ch,
	}
}

func (p *WeatherNotifyCommandProducer) Produce(sub domain.Subscription, weath domain.Weather) error {
	event := messaging.WeatherNotifyCommand{
		Email: sub.Email,
		Token: sub.Token.String(),
		Weather: messaging.Weather{
			Temperature: weath.Temperature,
			Humidity:    weath.Humidity,
			Description: weath.Description,
		},
	}
	body, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal weather notify command", zap.Error(err))
		return fmt.Errorf("weather notify command producer: %w", domain.ErrInternal)
	}
	err = p.ch.Publish(
		messaging.ExchangeName,
		messaging.WeatherRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish weather notify command", zap.Error(err))
		return fmt.Errorf("weather notify command producer: %w", domain.ErrInternal)
	}
	return nil
}

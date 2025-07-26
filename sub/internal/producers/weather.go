package producers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type WeatherNotifyCommandProducer struct {
	ch *amqp.Channel
}

func NewWeatherNotifyCommandProducer(ch *amqp.Channel) *WeatherNotifyCommandProducer {
	return &WeatherNotifyCommandProducer{
		ch: ch,
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
		log.Println(fmt.Errorf("weather notify command producer: %v", err))
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
		log.Println(fmt.Errorf("weather notify command producer: %v", err))
		return fmt.Errorf("weather notify command producer: %w", domain.ErrInternal)
	}
	return nil
}

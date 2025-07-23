package producers

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SubscribeEventProducer struct {
	ch *amqp.Channel
}

func NewSubscribeEventProducer(ch *amqp.Channel) *SubscribeEventProducer {
	return &SubscribeEventProducer{
		ch: ch,
	}
}

func (p *SubscribeEventProducer) Produce(sub domain.Subscription) error {
	event := messaging.SubscribeEvent{
		Email: sub.Email,
		Token: sub.Token.String(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("subscription producer: %w", err)
	}
	return p.ch.Publish(
		messaging.ExchangeName,
		messaging.SubscribeRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

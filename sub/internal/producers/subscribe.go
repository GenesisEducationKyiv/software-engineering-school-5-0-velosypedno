package producers

import (
	"encoding/json"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type SubscribeEventProducer struct {
	logger *zap.Logger
	ch     *amqp.Channel
}

func NewSubscribeEventProducer(logger *zap.Logger, ch *amqp.Channel) *SubscribeEventProducer {
	return &SubscribeEventProducer{
		logger: logger.With(zap.String("producer", "SubscribeEventProducer")),
		ch:     ch,
	}
}

func (p *SubscribeEventProducer) Produce(sub domain.Subscription) error {
	event := messaging.SubscribeEvent{
		Email: sub.Email,
		Token: sub.Token.String(),
	}
	body, err := json.Marshal(event)
	if err != nil {
		p.logger.Error("Failed to marshal subscribe event", zap.Error(err))
		return fmt.Errorf("subscription event producer: %w", domain.ErrInternal)
	}
	err = p.ch.Publish(
		messaging.ExchangeName,
		messaging.SubscribeRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		p.logger.Error("Failed to publish subscribe event", zap.Error(err))
		return fmt.Errorf("subscription event producer: %w", domain.ErrInternal)
	}
	return nil
}

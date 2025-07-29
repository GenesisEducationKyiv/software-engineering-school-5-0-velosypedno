package consumers

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type handler[T any] interface {
	Handle(T) error
}

type GenericConsumer[T any] struct {
	logger  *zap.Logger
	handler handler[T]
	msgs    <-chan amqp.Delivery
	name    string
}

func NewGenericConsumer[T any](logger *zap.Logger, handler handler[T], msgs <-chan amqp.Delivery, name string) *GenericConsumer[T] {
	return &GenericConsumer[T]{
		logger:  logger.With(zap.String("consumer", name)),
		handler: handler,
		msgs:    msgs,
		name:    name,
	}
}

func (c *GenericConsumer[T]) Consume(ctx context.Context) {
	c.logger.Info("Consumer started")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer stopped by context")
			return
		case rawMsg, ok := <-c.msgs:
			if !ok {
				c.logger.Info("Consumer stopped: channel closed")
				return
			}

			c.logger.Debug("Received new message")

			var msg T
			err := json.Unmarshal(rawMsg.Body, &msg)
			if err != nil {
				c.logger.Error("Failed to unmarshal message", zap.Error(err))
				if err := rawMsg.Reject(false); err != nil {
					c.logger.Error("Failed to reject message", zap.Error(err))
				}
				continue
			}

			if err := c.handler.Handle(msg); err != nil {
				c.logger.Error("Handler returned error", zap.Error(err))
				if err := rawMsg.Nack(false, true); err != nil {
					c.logger.Error("Failed to nack message", zap.Error(err))
				}
				continue
			}

			if err := rawMsg.Ack(false); err != nil {
				c.logger.Error("Failed to ack message", zap.Error(err))
			} else {
				c.logger.Debug("Message successfully acked")
			}
		}
	}
}

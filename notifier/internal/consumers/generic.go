package consumers

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type metrics interface {
	IncReceived(consumer string)
	IncAcked(consumer string)
	IncNotAcked(consumer string)
	ObserveHandleDuration(consumer string, seconds float64)
}

type handler[T any] interface {
	Handle(T) error
}

type GenericConsumer[T any] struct {
	logger  *zap.Logger
	handler handler[T]
	msgs    <-chan amqp.Delivery
	name    string
	metrics metrics
}

func NewGenericConsumer[T any](
	logger *zap.Logger,
	handler handler[T],
	msgs <-chan amqp.Delivery,
	name string,
	metrics metrics,
) *GenericConsumer[T] {
	return &GenericConsumer[T]{
		logger:  logger.With(zap.String("consumer", name)),
		handler: handler,
		msgs:    msgs,
		name:    name,
		metrics: metrics,
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

			c.metrics.IncReceived(c.name)
			c.logger.Debug("Received new message")

			var msg T
			if err := json.Unmarshal(rawMsg.Body, &msg); err != nil {
				c.logger.Error("Failed to unmarshal message", zap.Error(err))
				c.metrics.IncNotAcked(c.name)

				if err := rawMsg.Reject(false); err != nil {
					c.logger.Error("Failed to reject message", zap.Error(err))
					c.metrics.IncNotAcked(c.name)
				}
				continue
			}

			start := time.Now()
			if err := c.handler.Handle(msg); err != nil {
				c.logger.Error("Handler returned error", zap.Error(err))
				c.metrics.IncNotAcked(c.name)

				if err := rawMsg.Nack(false, true); err != nil {
					c.logger.Error("Failed to nack message", zap.Error(err))
					c.metrics.IncNotAcked(c.name)
				}
				continue
			}
			duration := time.Since(start).Seconds()
			c.metrics.ObserveHandleDuration(c.name, duration)

			if err := rawMsg.Ack(false); err != nil {
				c.logger.Error("Failed to ack message", zap.Error(err))
				c.metrics.IncNotAcked(c.name)
			} else {
				c.logger.Debug("Message successfully acked")
				c.metrics.IncAcked(c.name)
			}
		}
	}
}

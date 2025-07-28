package consumers

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type handler[T any] interface {
	Handle(T) error
}

type GenericConsumer[T any] struct {
	handler handler[T]
	msgs    <-chan amqp.Delivery
	name    string
}

func NewGenericConsumer[T any](handler handler[T], msgs <-chan amqp.Delivery, name string) *GenericConsumer[T] {
	return &GenericConsumer[T]{
		handler: handler,
		msgs:    msgs,
		name:    name,
	}
}

func (c *GenericConsumer[T]) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("%s consumer stopped\n", c.name)
			return
		case rawMsg, ok := <-c.msgs:
			if !ok {
				log.Printf("%s consumer stopped: channel closed\n", c.name)
				return
			}
			var msg T
			err := json.Unmarshal(rawMsg.Body, &msg)
			if err != nil {
				log.Printf("%s consumer: %v\n", c.name, err)
				err = rawMsg.Reject(false)
				if err != nil {
					log.Printf("%s consumer: %v\n", c.name, err)
				}
				continue
			}
			err = c.handler.Handle(msg)
			if err != nil {
				log.Printf("%s consumer: %v\n", c.name, err)
				err = rawMsg.Nack(false, true)
				if err != nil {
					log.Printf("%s consumer: %v\n", c.name, err)
				}
				continue
			}
			err = rawMsg.Ack(false)
			if err != nil {
				log.Printf("%s consumer: %v\n", c.name, err)
			}
		}
	}
}

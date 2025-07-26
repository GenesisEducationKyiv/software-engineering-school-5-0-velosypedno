package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type subEventHandler interface {
	Handle(event messaging.SubscribeEvent) error
}

type SubscribeEventConsumer struct {
	handler subEventHandler
	msgs    <-chan amqp.Delivery
}

func NewSubscribeEventConsumer(handler subEventHandler, msgs <-chan amqp.Delivery) *SubscribeEventConsumer {
	return &SubscribeEventConsumer{
		handler: handler,
		msgs:    msgs,
	}
}

func (c *SubscribeEventConsumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("subscribe event consumer stopped")
			return
		case msg, ok := <-c.msgs:
			if !ok {
				log.Println("subscribe event consumer: channel closed")
				return
			}
			var event messaging.SubscribeEvent
			err := json.Unmarshal(msg.Body, &event)
			if err != nil {
				log.Println(fmt.Errorf("subscribe event consumer: %v", err))
				err = msg.Reject(false)
				if err != nil {
					log.Println(fmt.Errorf("subscribe event consumer: %v", err))
				}
				continue
			}
			err = c.handler.Handle(event)
			if err != nil {
				log.Println(fmt.Errorf("subscribe event consumer: %v", err))
				err = msg.Nack(false, true)
				if err != nil {
					log.Println(fmt.Errorf("subscribe event consumer: %v", err))
				}
				continue
			}
			err = msg.Ack(false)
			if err != nil {
				log.Println(fmt.Errorf("subscribe event consumer: %v", err))
			}
		}
	}
}

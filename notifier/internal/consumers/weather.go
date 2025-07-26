package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type weatherCommandHandler interface {
	Handle(command messaging.WeatherNotifyCommand) error
}

type WeatherCommandConsumer struct {
	handler weatherCommandHandler
	msgs    <-chan amqp.Delivery
}

func NewWeatherCommandConsumer(handler weatherCommandHandler, msgs <-chan amqp.Delivery) *WeatherCommandConsumer {
	return &WeatherCommandConsumer{
		handler: handler,
		msgs:    msgs,
	}
}

func (c *WeatherCommandConsumer) Consume(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("weather notify command consumer stopped")
			return
		case msg, ok := <-c.msgs:
			if !ok {
				log.Println("weather notify command consumer: channel closed")
				return
			}
			var command messaging.WeatherNotifyCommand
			err := json.Unmarshal(msg.Body, &command)
			if err != nil {
				log.Println(fmt.Errorf("weather notify command consumer: %v", err))
				err = msg.Reject(false)
				if err != nil {
					log.Println(fmt.Errorf("weather notify command consumer: %v", err))
				}
				continue
			}
			err = c.handler.Handle(command)
			if err != nil {
				log.Println(fmt.Errorf("weather notify command consumer: %v", err))
				err = msg.Nack(false, true)
				if err != nil {
					log.Println(fmt.Errorf("weather notify command consumer: %v", err))
				}
				continue
			}
			err = msg.Ack(false)
			if err != nil {
				log.Println(fmt.Errorf("weather notify command consumer: %v", err))
			}
		}
	}
}

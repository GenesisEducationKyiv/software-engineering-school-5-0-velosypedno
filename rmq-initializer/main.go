package main

import (
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	"github.com/kelseyhightower/envconfig"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	Host string `envconfig:"RABBITMQ_HOST" required:"true"`
	Port string `envconfig:"RABBITMQ_PORT" required:"true"`
	User string `envconfig:"RABBITMQ_USER" required:"true"`
	Pass string `envconfig:"RABBITMQ_PASSWORD" required:"true"`
}

func (c RabbitMQConfig) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Pass, c.Host, c.Port)
}

func main() {
	cfg := RabbitMQConfig{}
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}
	conn, err := amqp.Dial(cfg.Addr())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		messaging.ExchangeName,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatal(err)
	}

	q, err := ch.QueueDeclare(
		messaging.SubscribeQueueName, // name
		true,                         // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,                        // queue name
		messaging.SubscribeRoutingKey, // routing key
		messaging.ExchangeName,        // exchange
		false,
		nil)
	if err != nil {
		log.Fatal(err)
	}

	q, err = ch.QueueDeclare(
		messaging.WeatherQueueName, // name
		true,                       // durable
		false,                      // delete when unused
		false,                      // exclusive
		false,                      // no-wait
		nil,                        // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,                      // queue name
		messaging.WeatherRoutingKey, // routing key
		messaging.ExchangeName,      // exchange
		false,
		nil)
	if err != nil {
		log.Fatal(err)
	}
}

//go:build integration

package consumers_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	RMQConnection *amqp.Connection
	RMQChannel    *amqp.Channel
)

func closeConnections() {
	if RMQChannel != nil {
		if err := RMQChannel.Close(); err != nil {
			log.Println("failed to close RMQ channel:", err)
		}
	}
	if RMQConnection != nil {
		if err := RMQConnection.Close(); err != nil {
			log.Println("failed to close RMQ conn:", err)
		}
	}
}

func TestMain(m *testing.M) {

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(cfg)

	// setup messaging
	RMQConnection, RMQChannel, err = setupRMQ(cfg.RabbitMQ)
	if err != nil {
		closeConnections()
		log.Panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// run app
	logger := zap.NewNop()
	logFactory := logging.NewFactory(logger, "notifier")
	app := app.New(cfg, logFactory)
	go func() {
		runErr := app.Run(ctx)
		if runErr != nil {
			closeConnections()
			log.Panic(runErr)
		}
	}()

	// run tests
	code := m.Run()
	cancel()
	closeConnections()
	os.Exit(code)
}

func clearRMQ() {
	_, err := RMQChannel.QueuePurge(messaging.SubscribeQueueName, false)
	if err != nil {
		log.Panic(err)
	}
}

func setupRMQ(cfg config.RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.Addr())
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return conn, nil, err
	}

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
		return conn, ch, err
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
		return conn, ch, err
	}

	err = ch.QueueBind(
		q.Name,                        // queue name
		messaging.SubscribeRoutingKey, // routing key
		messaging.ExchangeName,        // exchange
		false,
		nil)
	if err != nil {
		return conn, ch, err
	}

	return conn, ch, nil
}

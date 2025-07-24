package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	shutdownTimeout = 20 * time.Second
)

type App struct {
	cfg *config.Config

	rmqConn *amqp.Connection
	rmqCh   *amqp.Channel
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// rabbitmq
	a.rmqConn, err = amqp.Dial(a.cfg.RabbitMQ.Addr())
	if err != nil {
		return err
	}
	a.rmqCh, err = a.rmqConn.Channel()
	if err != nil {
		return err
	}

	err = a.rmqCh.ExchangeDeclare(
		messaging.ExchangeName, // name
		"direct",               // type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		return err
	}

	subEventConsumer, err := a.setupSubscribeEventConsumer()
	if err != nil {
		return err
	}
	go subEventConsumer.Consume(ctx)

	// wait on shutdown signal
	<-ctx.Done()

	// shutdown
	err = a.shutdown()
	return err
}

func (a *App) shutdown() error {
	var shutdownErr error

	// rabbitmq
	if a.rmqCh != nil {
		if err := a.rmqCh.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq channel: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("RabbitMQ channel closed")
		}
	}
	if a.rmqConn != nil {
		if err := a.rmqConn.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq connection: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("RabbitMQ connection closed")
		}
	}
	return shutdownErr
}

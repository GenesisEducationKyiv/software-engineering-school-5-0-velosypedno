package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const readTimeout = 10 * time.Second
const shutdownTimeout = 20 * time.Second

type App struct {
	cfg *config.Config

	rmqConn *amqp.Connection
	rmqCh   *amqp.Channel
	httpSrv *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// rabbitmq
	log.Println("Connecting to RabbitMQ...")
	a.rmqConn, err = amqp.Dial(a.cfg.RabbitMQ.Addr())
	if err != nil {
		return err
	}
	log.Println("RabbitMQ connected")
	log.Println("Creating RabbitMQ channel...")
	a.rmqCh, err = a.rmqConn.Channel()
	if err != nil {
		return err
	}
	log.Println("RabbitMQ channel created")

	subEventConsumer, err := a.setupSubscribeEventConsumer()
	if err != nil {
		return err
	}
	go subEventConsumer.Consume(ctx)
	log.Println("Subscribe event consumer started in background")

	weatherCommandConsumer, err := a.setupWeatherCommandConsumer()
	if err != nil {
		return err
	}
	log.Println("Weather command consumer started in background")
	go weatherCommandConsumer.Consume(ctx)

	// http api
	router := a.setupRouter()
	log.Println("Router created")
	a.httpSrv = &http.Server{
		Addr:        a.cfg.HTTPSrv.Addr(),
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	log.Println("HTTP server started in background")
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil {
			log.Printf("http srv: %v", err)
		}
	}()

	// wait on shutdown signal
	<-ctx.Done()
	log.Println("Context canceled, shutting down app...")

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = a.shutdown(timeoutCtx)
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// http api
	if a.httpSrv != nil {
		if err := a.httpSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown api server: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("APIServer Shutdown successfully")
		}
	}

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

package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/metrics"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/prometheus/client_golang/prometheus"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

const readTimeout = 10 * time.Second
const shutdownTimeout = 20 * time.Second

var (
	appMetricsRegister = prometheus.DefaultRegisterer
)

type appMetrics struct {
	consumers *metrics.ConsumerMetrics
}

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	logFactory *logging.LoggerFactory

	rmqConn *amqp.Connection
	rmqCh   *amqp.Channel
	httpSrv *http.Server
	metrics appMetrics
}

func New(cfg *config.Config, logFactory *logging.LoggerFactory) *App {
	return &App{
		cfg:        cfg,
		logFactory: logFactory,
		logger:     logFactory.ForPackage("app"),
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// metrics
	a.logger.Info("Initializing metrics...")
	a.metrics.consumers = metrics.NewConsumerMetrics(appMetricsRegister)

	// rabbitmq
	a.logger.Info("Connecting to RabbitMQ...", zap.String("addr", a.cfg.RabbitMQ.Addr()))
	a.rmqConn, err = amqp.Dial(a.cfg.RabbitMQ.Addr())
	if err != nil {
		a.logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return err
	}
	a.logger.Info("RabbitMQ connected")

	a.logger.Info("Creating RabbitMQ channel...")
	a.rmqCh, err = a.rmqConn.Channel()
	if err != nil {
		a.logger.Error("Failed to create RabbitMQ channel", zap.Error(err))
		return err
	}
	a.logger.Info("RabbitMQ channel created")

	subEventConsumer, err := a.setupSubscribeEventConsumer()
	if err != nil {
		a.logger.Error("Failed to setup subscribe event consumer", zap.Error(err))
		return err
	}
	go subEventConsumer.Consume(ctx)
	a.logger.Info("Subscribe event consumer started in background")

	weatherCommandConsumer, err := a.setupWeatherCommandConsumer()
	if err != nil {
		a.logger.Error("Failed to setup weather command consumer", zap.Error(err))
		return err
	}
	go weatherCommandConsumer.Consume(ctx)
	a.logger.Info("Weather command consumer started in background")

	// http api
	router := a.setupRouter()
	a.logger.Info("Router created")

	a.httpSrv = &http.Server{
		Addr:        a.cfg.HTTPSrv.Addr(),
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	a.logger.Info("Starting http server...", zap.String("addr", a.httpSrv.Addr))
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("Failed to start http server", zap.Error(err))
		}
	}()

	// wait on shutdown signal
	<-ctx.Done()
	a.logger.Info("Shutdown signal received")

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	a.logger.Info("Shutting down...")
	err = a.shutdown(timeoutCtx)
	if err != nil {
		a.logger.Error("Failed to shutdown", zap.Error(err))
	}
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// http api
	a.logger.Info("Shutting down HTTP server...")
	if a.httpSrv != nil {
		if err := a.httpSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown api server: %w", err)
			a.logger.Error("HTTP server shutdown failed", zap.Error(wrapped))
			shutdownErr = wrapped
		} else {
			a.logger.Info("HTTP server shutdown successful")
		}
	}

	// rabbitmq
	a.logger.Info("Closing RabbitMQ channel...")
	if a.rmqCh != nil {
		if err := a.rmqCh.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq channel: %w", err)
			a.logger.Error("RabbitMQ channel close failed", zap.Error(wrapped))
			shutdownErr = wrapped
		} else {
			a.logger.Info("RabbitMQ channel closed")
		}
	}

	a.logger.Info("Closing RabbitMQ connection...")
	if a.rmqConn != nil {
		if err := a.rmqConn.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq connection: %w", err)
			a.logger.Error("RabbitMQ connection close failed", zap.Error(wrapped))
			shutdownErr = wrapped
		} else {
			a.logger.Info("RabbitMQ connection closed")
		}
	}

	return shutdownErr
}

package app

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	pbweath "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	brokernotify "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/notifiers/broker"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/producers"
	subrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/subscription"
	weathrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/weather"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	confirmSubTmplName = "confirm_sub.html"
	WeatherQueue       = "weather_email_queue"
	SubscribeQueue     = "subscribe_email_queue"
)

type (
	weatherRepo interface {
		GetCurrent(ctx context.Context, city string) (domain.Weather, error)
	}

	subscriptionRepo interface {
		Create(subscription domain.Subscription) error
		Activate(token uuid.UUID) error
		DeleteByToken(token uuid.UUID) error
		GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error)
	}
)

type (
	emailBackend interface {
		Send(to, subject, body string) error
	}

	weatherNotifier interface {
		SendCurrent(subscription domain.Subscription, weather domain.Weather) error
	}

	subNotifier interface {
		SendConfirmation(subscription domain.Subscription) error
	}
)

type InfrastructureContainer struct {
	DB       *sql.DB
	GRPCConn *grpc.ClientConn

	RabbitMQConn   *amqp.Connection
	RabbitMQCh     *amqp.Channel
	WeatherQueue   *amqp.Queue
	SubscribeQueue *amqp.Queue

	WeatherRepo weatherRepo
	SubRepo     subscriptionRepo

	EmailBackend    emailBackend
	WeatherNotifier weatherNotifier
	SubNotifier     subNotifier
}

func NewInfrastructureContainer(
	cfg *config.Config,
	logger *zap.Logger,
	logFactory *logging.LoggerFactory,
) (*InfrastructureContainer, error) {
	// messaging
	logger.Info("Connecting to RabbitMQ...")
	conn, err := newRabbitMQConn(cfg.RabbitMQ)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to RabbitMQ")

	logger.Info("Opening RabbitMQ channel...")
	ch, err := newRabbitMQChannel(conn)
	if err != nil {
		logger.Error("Failed to open RabbitMQ channel", zap.Error(err))
		return nil, err
	}
	logger.Info("RabbitMQ channel opened")

	logger.Info("Declaring RabbitMQ exchange...")
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
		logger.Error("Failed to declare RabbitMQ exchange", zap.Error(err))
		return nil, err
	}
	logger.Info("RabbitMQ exchange declared")

	// db
	logger.Info("Connecting to database...")
	db, err := newDB(cfg.DB)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to database")

	// gRPC to Weather Service
	logger.Info("Connecting to Weather gRPC service...")
	grpcConn, err := newWeatherGRPCConn(cfg)
	if err != nil {
		logger.Error("Failed to connect to Weather gRPC service", zap.Error(err))
		return nil, err
	}
	logger.Info("Connected to Weather gRPC service")

	// repos
	subRepoLogger := logFactory.ForPackage("repos/subscription")
	subRepo := subrepo.NewDBRepo(subRepoLogger, db)
	weathGRPCClient := pbweath.NewWeatherServiceClient(grpcConn)
	weatherRepoLogger := logFactory.ForPackage("repos/weather")
	weathRepo := weathrepo.NewGRPCRepo(weatherRepoLogger, weathGRPCClient)

	// mailers
	producerLogger := logFactory.ForPackage("producers")
	weatherNotifyCommandProducer := producers.NewWeatherNotifyCommandProducer(producerLogger, ch)
	weatherNotifier := brokernotify.NewWeatherNotifyCommandNotifier(weatherNotifyCommandProducer)
	subEventProducer := producers.NewSubscribeEventProducer(producerLogger, ch)
	subNotifier := brokernotify.NewSubscriptionEmailNotifier(subEventProducer)

	return &InfrastructureContainer{
		DB:              db,
		GRPCConn:        grpcConn,
		RabbitMQConn:    conn,
		RabbitMQCh:      ch,
		WeatherRepo:     weathRepo,
		SubRepo:         subRepo,
		WeatherNotifier: weatherNotifier,
		SubNotifier:     subNotifier,
	}, nil
}

func (c *InfrastructureContainer) Shutdown(ctx context.Context, logger *zap.Logger) error {
	var shutdownErr error

	logger.Info("Starting infrastructure shutdown")

	// grpc
	if c.GRPCConn != nil {
		logger.Info("Closing gRPC connection...")
		if err := c.GRPCConn.Close(); err != nil {
			errWrapped := fmt.Errorf("shutdown gRPC connection: %w", err)
			logger.Error("Failed to close gRPC connection", zap.Error(errWrapped))
			shutdownErr = errWrapped
		} else {
			logger.Info("gRPC connection closed")
		}
	}

	// db
	if c.DB != nil {
		logger.Info("Closing database connection...")
		if err := c.DB.Close(); err != nil {
			errWrapped := fmt.Errorf("shutdown db: %w", err)
			logger.Error("Failed to close database connection", zap.Error(errWrapped))
			if shutdownErr == nil {
				shutdownErr = errWrapped
			}
		} else {
			logger.Info("Database connection closed")
		}
	}

	// messaging
	if c.RabbitMQCh != nil {
		logger.Info("Closing RabbitMQ channel...")
		if err := c.RabbitMQCh.Close(); err != nil {
			errWrapped := fmt.Errorf("shutdown RabbitMQ channel: %w", err)
			logger.Error("Failed to close RabbitMQ channel", zap.Error(errWrapped))
			if shutdownErr == nil {
				shutdownErr = errWrapped
			}
		} else {
			logger.Info("RabbitMQ channel closed")
		}
	}

	if c.RabbitMQConn != nil {
		logger.Info("Closing RabbitMQ connection...")
		if err := c.RabbitMQConn.Close(); err != nil {
			errWrapped := fmt.Errorf("shutdown RabbitMQ connection: %w", err)
			logger.Error("Failed to close RabbitMQ connection", zap.Error(errWrapped))
			if shutdownErr == nil {
				shutdownErr = errWrapped
			}
		} else {
			logger.Info("RabbitMQ connection closed")
		}
	}

	logger.Info("Infrastructure shutdown complete")
	return shutdownErr
}

func newWeatherGRPCConn(cfg *config.Config) (*grpc.ClientConn, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	grpcConn, err := grpc.NewClient(cfg.WeathSvc.Addr(), opt)
	if err != nil {
		return nil, err
	}
	return grpcConn, nil
}

func newDB(cfg config.DBConfig) (*sql.DB, error) {
	db, err := sql.Open(cfg.Driver, cfg.DSN())
	if err != nil {
		return nil, err
	}
	return db, nil
}

func newRabbitMQConn(cfg config.RabbitMQConfig) (*amqp.Connection, error) {
	conn, err := amqp.Dial(cfg.Addr())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newRabbitMQChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

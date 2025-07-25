package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	pbweath "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/email"
	brokernotify "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/notifiers/broker"
	emailnotify "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/notifiers/email"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/producers"
	subrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/subscription"
	weathrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/weather"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
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

func NewInfrastructureContainer(cfg config.Config) (*InfrastructureContainer, error) {
	// messaging
	conn, err := newRabbitMQConn(cfg.RabbitMQ)
	if err != nil {
		return nil, err
	}
	ch, err := newRabbitMQChannel(conn)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// db
	db, err := newDB(cfg.DB)
	if err != nil {
		return nil, err
	}

	// repos
	subRepo := subrepo.NewDBRepo(db)
	grpcConn, err := newWeatherGRPCConn(cfg)
	if err != nil {
		return nil, err
	}
	weathGRPCClient := pbweath.NewWeatherServiceClient(grpcConn)
	weathRepo := weathrepo.NewGRPCAdapter(weathGRPCClient)

	// mailers
	emailBackend := newSMTPEmailBackend(cfg.SMTP)
	weatherNotifier := emailnotify.NewWeatherEmailNotifier(emailBackend)

	subEventProducer := producers.NewSubscribeEventProducer(ch)
	subNotifier := brokernotify.NewSubscriptionEmailNotifier(subEventProducer)

	return &InfrastructureContainer{
		DB:       db,
		GRPCConn: grpcConn,

		RabbitMQConn: conn,
		RabbitMQCh:   ch,

		WeatherRepo: weathRepo,
		SubRepo:     subRepo,

		EmailBackend:    emailBackend,
		WeatherNotifier: weatherNotifier,
		SubNotifier:     subNotifier,
	}, nil
}

func (c *InfrastructureContainer) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// grpc
	if c.GRPCConn != nil {
		if err := c.GRPCConn.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown gRPC connection: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("gRPC connection closed")
		}
	}

	// db
	if c.DB != nil {
		if err := c.DB.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown db: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("DB closed")
		}
	}

	// messaging
	if c.RabbitMQCh != nil {
		if err := c.RabbitMQCh.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq channel: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("RabbitMQ channel closed")
		}
	}

	if c.RabbitMQConn != nil {
		if err := c.RabbitMQConn.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown rabbitmq connection: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("RabbitMQ connection closed")
		}
	}

	return shutdownErr
}

func newWeatherGRPCConn(cfg config.Config) (*grpc.ClientConn, error) {
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	grpcConn, err := grpc.NewClient(cfg.WeathSvc.Addr(), opt)
	if err != nil {
		return nil, err
	}
	return grpcConn, nil
}

func newSMTPEmailBackend(cfg config.SMTPConfig) *email.SMTPBackend {
	return email.NewSMTPBackend(cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.EmailFrom)
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

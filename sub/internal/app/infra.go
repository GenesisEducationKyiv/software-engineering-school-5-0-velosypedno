package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	pbweath "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/email"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/mailers"
	subrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/subscription"
	weathrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/repos/weather"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	confirmSubTmplName = "confirm_sub.html"
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

	weatherMailer interface {
		SendCurrent(subscription domain.Subscription, weather domain.Weather) error
	}

	subMailer interface {
		SendConfirmation(subscription domain.Subscription) error
	}
)

type InfrastructureContainer struct {
	DB       *sql.DB
	GRPCConn *grpc.ClientConn

	WeatherRepo weatherRepo
	SubRepo     subscriptionRepo

	EmailBackend  emailBackend
	WeatherMailer weatherMailer
	SubMailer     subMailer
}

func NewInfrastructureContainer(cfg config.Config) (*InfrastructureContainer, error) {
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
	weatherMailer := mailers.NewWeatherMailer(emailBackend)
	confirmTmplPath := filepath.Join(cfg.Srv.TemplatesDir, confirmSubTmplName)
	subMailer := mailers.NewSubscriptionMailer(emailBackend, confirmTmplPath)

	return &InfrastructureContainer{
		DB:       db,
		GRPCConn: grpcConn,

		WeatherRepo: weathRepo,
		SubRepo:     subRepo,

		EmailBackend:  emailBackend,
		WeatherMailer: weatherMailer,
		SubMailer:     subMailer,
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

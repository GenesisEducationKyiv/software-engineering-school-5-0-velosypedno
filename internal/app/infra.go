package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/cache"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/email"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/mailers"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/metrics"
	subrepo "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/subscription"
	weathchain "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/chain"
	weathdecorator "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/decorator"
	weathprovider "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/repos/weather/provider"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/cb"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

const (
	logPerm os.FileMode = 0644

	logFilepath        = "./log.log"
	confirmSubTmplName = "confirm_sub.html"

	freeWeatherName    = "weatherapi.com"
	tomorrowIOName     = "tomorrow.io"
	visualCrossingName = "visualcrossing.com"

	// CB = CircuitBreaker
	weatherCBTimeout = 5 * time.Minute
	weatherCBLimit   = 10
	weatherCBRecover = 5

	cacheTTL = 5 * time.Minute
)

var (
	appMetricsRegister = prometheus.DefaultRegisterer
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
	DB             *sql.DB
	Redis          *redis.Client
	WeatherMetrics *metrics.WeatherMetrics
	Logger         *log.Logger

	WeatherRepo weatherRepo
	SubRepo     subscriptionRepo

	EmailBackend  emailBackend
	WeatherMailer weatherMailer
	SubMailer     subMailer
}

func NewInfrastructureContainer(cfg config.Config) (*InfrastructureContainer, error) {
	// db, metrics, logger, redis
	db, err := newDB(cfg.DB)
	if err != nil {
		return nil, err
	}
	redis := newRedis(cfg.Redis)
	weatherMetrics := metrics.NewWeatherMetrics(appMetricsRegister)
	logger, err := newLogger(logFilepath)
	if err != nil {
		return nil, err
	}

	// repos
	weatherRepo := newWeatherRepo(cfg, redis, weatherMetrics, logger)
	subRepo := subrepo.NewDBRepo(db)

	// mailers
	emailBackend := newSMTPEmailBackend(cfg.SMTP)
	weatherMailer := mailers.NewWeatherMailer(emailBackend)
	confirmTmplPath := filepath.Join(cfg.Srv.TemplatesDir, confirmSubTmplName)
	subMailer := mailers.NewSubscriptionMailer(emailBackend, confirmTmplPath)

	return &InfrastructureContainer{
		DB:             db,
		Redis:          redis,
		WeatherMetrics: weatherMetrics,
		Logger:         logger,

		WeatherRepo: weatherRepo,
		SubRepo:     subRepo,

		EmailBackend:  emailBackend,
		WeatherMailer: weatherMailer,
		SubMailer:     subMailer,
	}, nil
}

func (c *InfrastructureContainer) Shutdown(ctx context.Context) error {
	var shutdownErr error

	// redis
	if c.Redis != nil {
		if err := c.Redis.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown redis: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("Redis closed")
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

func newWeatherRepo(cfg config.Config, redis *redis.Client,
	metrics *metrics.WeatherMetrics, logger *log.Logger) *weathdecorator.CacheDecorator {
	freeWeathR := weathprovider.NewFreeWeatherAPI(
		weathprovider.APICfg{APIKey: cfg.FreeWeather.Key, APIURL: cfg.FreeWeather.URL},
		&http.Client{},
	)
	tomorrowWeathR := weathprovider.NewTomorrowAPI(
		weathprovider.APICfg{APIKey: cfg.TomorrowWeather.Key, APIURL: cfg.TomorrowWeather.URL},
		&http.Client{},
	)
	vcWeathR := weathprovider.NewVisualCrossingAPI(
		weathprovider.APICfg{APIKey: cfg.VisualCrossing.Key, APIURL: cfg.VisualCrossing.URL},
		&http.Client{},
	)

	logFreeWeathR := weathdecorator.NewLogDecorator(freeWeathR, freeWeatherName, logger)
	logTomorrowR := weathdecorator.NewLogDecorator(tomorrowWeathR, tomorrowIOName, logger)
	logVcWeathR := weathdecorator.NewLogDecorator(vcWeathR, visualCrossingName, logger)

	breakerFreeWeathR := weathdecorator.NewBreakerDecorator(logFreeWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerTomorrowR := weathdecorator.NewBreakerDecorator(logTomorrowR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))
	breakerVcWeathR := weathdecorator.NewBreakerDecorator(logVcWeathR,
		cb.NewCircuitBreaker(weatherCBTimeout, weatherCBLimit, weatherCBRecover))

	weathChain := weathchain.NewProvidersFallbackChain(breakerFreeWeathR, breakerTomorrowR, breakerVcWeathR)

	redisBackend := cache.NewRedisCacheClient[domain.Weather](redis, cacheTTL)
	cachedRepoChain := weathdecorator.NewCacheDecorator(weathChain, redisBackend, metrics)
	return cachedRepoChain
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

func newRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Pass,
	})
}

func newLogger(path string) (*log.Logger, error) {
	// #nosec G304 -- logFilepath is a constant and not user-controlled
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logPerm)
	if err != nil {
		return nil, err
	}
	return log.New(f, "", log.LstdFlags), nil
}

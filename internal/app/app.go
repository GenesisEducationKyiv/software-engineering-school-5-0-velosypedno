package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/metrics"
)

const (
	readTimeout     = 15 * time.Second
	shutdownTimeout = 20 * time.Second
	logFilepath     = "log.log"

	logPerm os.FileMode = 0644
)

var (
	appMetricsRegister = prometheus.DefaultRegisterer
)

type appMetrics struct {
	weather *metrics.WeatherMetrics
}

type App struct {
	cfg         *config.Config
	db          *sql.DB
	redisClient *redis.Client
	cron        *cron.Cron
	apiSrv      *http.Server
	reposLogger *log.Logger
	metrics     appMetrics
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// logger
	f, err := os.OpenFile(logFilepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, logPerm)
	if err != nil {
		return err
	}
	a.reposLogger = log.New(f, "", log.LstdFlags)

	// metrics
	a.metrics.weather = metrics.NewWeatherMetrics(appMetricsRegister)

	// db
	a.db, err = sql.Open(a.cfg.DB.Driver, a.cfg.DB.DSN())
	if err != nil {
		return err
	}
	log.Println("DB connected")

	// redis
	a.redisClient = redis.NewClient(&redis.Options{
		Addr:     a.cfg.Redis.Addr(),
		Password: a.cfg.Redis.Pass,
	})
	log.Println("Redis connected")

	// cron
	err = a.setupCron()
	if err != nil {
		return err
	}
	a.cron.Start()
	log.Println("Cron tasks are scheduled")

	// api
	router := a.setupRouter()
	a.apiSrv = &http.Server{
		Addr:        ":" + a.cfg.Srv.Port,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	go func() {
		if err := a.apiSrv.ListenAndServe(); err != nil {
			log.Printf("api server: %v", err)
		}
	}()
	log.Printf("APIServer started on port %s", a.cfg.Srv.Port)

	// wait on shutdown signal
	<-ctx.Done()

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = a.shutdown(timeoutCtx)
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// api
	if a.apiSrv != nil {
		if err := a.apiSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown api server: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("APIServer Shutdown successfully")
		}
	}

	// cron
	if a.cron != nil {
		log.Println("Stopping cron scheduler")
		cronCtx := a.cron.Stop()
		select {
		case <-cronCtx.Done():
			log.Println("Cron scheduler stopped")
		case <-timeoutCtx.Done():
			wrapped := fmt.Errorf("shutdown cron scheduler: %w", timeoutCtx.Err())
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		}
	}

	// redis
	if a.redisClient != nil {
		if err := a.redisClient.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown redis: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("Redis closed")
		}
	}

	// db
	if a.db != nil {
		if err := a.db.Close(); err != nil {
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

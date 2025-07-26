package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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
	cfg *config.Config

	redisClient *redis.Client
	httpSrv     *http.Server
	grpcSrv     *grpc.Server
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

	// redis
	a.redisClient = redis.NewClient(&redis.Options{
		Addr:     a.cfg.Redis.Addr(),
		Password: a.cfg.Redis.Pass,
	})
	log.Println("Redis connected")

	// http api
	router := a.setupRouter()
	a.httpSrv = &http.Server{
		Addr:        a.cfg.GRPCSrv.Host + ":" + a.cfg.HTTPSrv.Port,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil {
			log.Printf("http api: %v", err)
		}
	}()
	log.Printf("HTTP api started on port %s", a.cfg.HTTPSrv.Port)

	// grpc api
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", a.cfg.GRPCSrv.Host, a.cfg.GRPCSrv.Port))
	if err != nil {
		return err
	}
	a.grpcSrv = a.setupGRPCSrv()
	go func() {
		err = a.grpcSrv.Serve(lis)
		if err != nil {
			log.Printf("grpc api: %v", err)
		}
	}()

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

	// grpc api
	if a.grpcSrv != nil {
		done := make(chan struct{})
		go func() {
			a.grpcSrv.GracefulStop()
			close(done)
		}()
		select {
		case <-timeoutCtx.Done():
			log.Printf("shutdown grpc timeout: %v", timeoutCtx.Err())
		case <-done:
			log.Println("gRPC server stopped")
		}
	}

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

	return shutdownErr
}

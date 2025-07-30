package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	readTimeout     = 15 * time.Second
	shutdownTimeout = 20 * time.Second
)

var (
	appMetricsRegister = prometheus.DefaultRegisterer
)

type appMetrics struct {
	weather *metrics.WeatherMetrics
}

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	logFactory *logging.LoggerFactory

	redisClient *redis.Client
	httpSrv     *http.Server
	grpcSrv     *grpc.Server
	metrics     appMetrics
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

	a.logger.Info("Initializing weather service...")

	// metrics
	a.logger.Info("Initializing metrics...")
	a.metrics.weather = metrics.NewWeatherMetrics(appMetricsRegister)

	// redis
	a.logger.Info("Connecting to Redis...", zap.String("addr", a.cfg.Redis.Addr()))
	a.redisClient = redis.NewClient(&redis.Options{
		Addr:     a.cfg.Redis.Addr(),
		Password: a.cfg.Redis.Pass,
	})
	a.logger.Info("Redis connected")

	// http API
	a.logger.Info("Setting up HTTP server...")
	router := a.setupRouter()
	a.httpSrv = &http.Server{
		Addr:        a.cfg.GRPCSrv.Host + ":" + a.cfg.HTTPSrv.Port,
		Handler:     router,
		ReadTimeout: readTimeout,
	}
	go func() {
		a.logger.Info("Starting HTTP server...", zap.String("addr", a.httpSrv.Addr))
		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// gRPC API
	a.logger.Info("Starting gRPC server...",
		zap.String("host", a.cfg.GRPCSrv.Host),
		zap.String("port", a.cfg.GRPCSrv.Port),
	)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", a.cfg.GRPCSrv.Host, a.cfg.GRPCSrv.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on grpc: %w", err)
	}
	a.grpcSrv = a.setupGRPCSrv()
	go func() {
		if err := a.grpcSrv.Serve(lis); err != nil {
			a.logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	// wait for shutdown signal
	<-ctx.Done()
	a.logger.Info("Shutdown signal received")

	// graceful shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	a.logger.Info("Shutting down...")
	err = a.shutdown(timeoutCtx)
	if err != nil {
		a.logger.Error("Shutdown error", zap.Error(err))
	}
	return err
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// gRPC shutdown
	if a.grpcSrv != nil {
		a.logger.Info("Shutting down gRPC server...")
		done := make(chan struct{})
		go func() {
			a.grpcSrv.GracefulStop()
			close(done)
		}()
		select {
		case <-timeoutCtx.Done():
			a.logger.Error("Timeout during gRPC shutdown", zap.Error(timeoutCtx.Err()))
		case <-done:
			a.logger.Info("gRPC server stopped")
		}
	}

	// HTTP shutdown
	if a.httpSrv != nil {
		a.logger.Info("Shutting down HTTP server...")
		if err := a.httpSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown HTTP server: %w", err)
			a.logger.Error("HTTP shutdown failed", zap.Error(wrapped))
			shutdownErr = wrapped
		} else {
			a.logger.Info("HTTP server stopped")
		}
	}

	// Redis shutdown
	if a.redisClient != nil {
		a.logger.Info("Closing Redis client...")
		if err := a.redisClient.Close(); err != nil {
			wrapped := fmt.Errorf("shutdown Redis: %w", err)
			a.logger.Error("Redis shutdown failed", zap.Error(wrapped))
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			a.logger.Info("Redis client closed")
		}
	}

	return shutdownErr
}

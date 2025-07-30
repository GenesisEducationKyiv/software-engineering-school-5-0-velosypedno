package app

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/robfig/cron/v3"
)

const (
	readTimeout     = 15 * time.Second
	shutdownTimeout = 20 * time.Second
)

type App struct {
	cfg        *config.Config
	logFactory *logging.LoggerFactory
	logger     *zap.Logger

	cron    *cron.Cron
	grpcAPI *grpc.Server

	infra    *InfrastructureContainer
	business *BusinessContainer
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

	a.logger.Info("Initializing infrastructure container")
	a.infra, err = NewInfrastructureContainer(*a.cfg, a.logger)
	if err != nil {
		a.logger.Error("Failed to initialize infrastructure container", zap.Error(err))
		return err
	}

	a.logger.Info("Initializing business container")
	a.business, err = NewBusinessContainer(a.infra)
	if err != nil {
		a.logger.Error("Failed to initialize business container", zap.Error(err))
		return err
	}

	a.logger.Info("Initializing presentation container")
	presentation, err := NewPresentationContainer(a.business)
	if err != nil {
		a.logger.Error("Failed to initialize presentation container", zap.Error(err))
		return err
	}

	// cron
	a.logger.Info("Starting cron scheduler")
	a.cron = presentation.Cron
	a.cron.Start()

	// grpc api
	a.logger.Info("Starting gRPC server", zap.String("host", a.cfg.GRPCSrv.Host), zap.String("port", a.cfg.GRPCSrv.Port))
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", a.cfg.GRPCSrv.Host, a.cfg.GRPCSrv.Port))
	if err != nil {
		a.logger.Error("Failed to listen on gRPC address", zap.Error(err))
		return err
	}

	a.grpcAPI = presentation.GRPCSrv
	go func() {
		a.logger.Info("GRPC server is now serving")
		if err := a.grpcAPI.Serve(lis); err != nil {
			a.logger.Error("GRPC server exited with error", zap.Error(err))
		}
	}()

	// wait on shutdown signal
	<-ctx.Done()
	a.logger.Info("Shutdown signal received")

	// shutdown
	timeoutCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	a.logger.Info("Shutting down...")
	if err := a.shutdown(timeoutCtx); err != nil {
		a.logger.Error("Shutdown failed", zap.Error(err))
		return err
	}

	a.logger.Info("Shutdown completed successfully")
	return nil
}

func (a *App) shutdown(timeoutCtx context.Context) error {
	var shutdownErr error

	// grpc api
	a.logger.Info("Shutting down gRPC server")
	if a.grpcAPI != nil {
		done := make(chan struct{})
		go func() {
			a.grpcAPI.GracefulStop()
			close(done)
		}()
		select {
		case <-timeoutCtx.Done():
			err := fmt.Errorf("Shutdown grpc timeout: %w", timeoutCtx.Err())
			a.logger.Error("Timeout during gRPC shutdown", zap.Error(err))
			shutdownErr = err
		case <-done:
			a.logger.Info("GRPC server stopped successfully")
		}
	}

	// cron
	a.logger.Info("Shutting down cron scheduler")
	if a.cron != nil {
		cronCtx := a.cron.Stop()
		select {
		case <-cronCtx.Done():
			a.logger.Info("Cron scheduler stopped successfully")
		case <-timeoutCtx.Done():
			err := fmt.Errorf("Shutdown cron scheduler: %w", timeoutCtx.Err())
			a.logger.Error("Timeout during cron shutdown", zap.Error(err))
			shutdownErr = err
		}
	}

	// infrastructure
	a.logger.Info("Shutting down infrastructure")
	if a.infra != nil {
		if err := a.infra.Shutdown(timeoutCtx, a.logger); err != nil {
			errWrapped := fmt.Errorf("shutdown infrastructure: %w", err)
			a.logger.Error("Error during infrastructure shutdown", zap.Error(errWrapped))
			if shutdownErr == nil {
				shutdownErr = errWrapped
			}
		} else {
			a.logger.Info("Infrastructure shut down successfully")
		}
	}

	return shutdownErr
}

package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	pbsub "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	pbweath "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownTimeout = 20 * time.Second

type App struct {
	cfg        *config.Config
	logger     *zap.Logger
	logFactory *logging.LoggerFactory

	subGRPCCon      *grpc.ClientConn
	subGRPCClient   pbsub.SubscriptionServiceClient
	weathGRPCCon    *grpc.ClientConn
	weathGRPCClient pbweath.WeatherServiceClient
	httpSrv         *http.Server
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

	// setup subscription grpc connection
	a.logger.Info("Setting up grpc connection to subscription service...", zap.String("addr", a.cfg.SubSvc.Addr()))
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	a.subGRPCCon, err = grpc.NewClient(a.cfg.SubSvc.Addr(), opt)
	if err != nil {
		return err
	}
	a.logger.Info("Grpc connection to subscription service established")
	a.subGRPCClient = pbsub.NewSubscriptionServiceClient(a.subGRPCCon)

	// setup weather grpc connection
	a.logger.Info("Setting up grpc connection to weather service...", zap.String("addr", a.cfg.WeatherSvc.Addr()))
	a.weathGRPCCon, err = grpc.NewClient(a.cfg.WeatherSvc.Addr(), opt)
	if err != nil {
		return err
	}
	a.logger.Info("Grpc connection to weather service established")
	a.weathGRPCClient = pbweath.NewWeatherServiceClient(a.weathGRPCCon)

	// setup http server
	a.httpSrv = a.setupHTTPServer()
	go func() {
		a.logger.Info("Starting http server...", zap.String("addr", a.httpSrv.Addr))
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

	// http server
	a.logger.Info("Shutting down http server...")
	if a.httpSrv != nil {
		if err := a.httpSrv.Shutdown(timeoutCtx); err != nil {
			a.logger.Error("Failed to shutdown http server", zap.Error(err))
			wrapped := fmt.Errorf("shutdown http server: %w", err)
			shutdownErr = wrapped
		} else {
			a.logger.Info("HTTP Server Shutdown successfully")
		}
	}

	// gRPC subscription connection
	a.logger.Info("Shutting down grpc connection to subscription service...")
	if a.subGRPCCon != nil {
		if err := a.subGRPCCon.Close(); err != nil {
			a.logger.Error("Failed to shutdown grpc connection to subscription service", zap.Error(err))
			wrapped := fmt.Errorf("close sub grpc connection: %w", err)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			a.logger.Info("Grpc connection to subscription service closed successfully")
		}
	}

	// gRPC weather connection
	a.logger.Info("Shutting down grpc connection to weather service...")
	if a.weathGRPCCon != nil {
		if err := a.weathGRPCCon.Close(); err != nil {
			a.logger.Error("Failed to shutdown grpc connection to weather service", zap.Error(err))
			wrapped := fmt.Errorf("close weather grpc connection: %w", err)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			a.logger.Info("Grpc connection to weather service closed successfully")
		}
	}

	return shutdownErr
}

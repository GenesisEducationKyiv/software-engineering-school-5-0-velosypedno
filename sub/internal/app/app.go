package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"google.golang.org/grpc"

	"github.com/robfig/cron/v3"
)

const (
	readTimeout     = 15 * time.Second
	shutdownTimeout = 20 * time.Second
)

type App struct {
	cfg *config.Config

	cron    *cron.Cron
	grpcAPI *grpc.Server

	infra    *InfrastructureContainer
	business *BusinessContainer
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	a.infra, err = NewInfrastructureContainer(*a.cfg)
	if err != nil {
		return err
	}
	a.business, err = NewBusinessContainer(a.infra)
	if err != nil {
		return err
	}
	presentation, err := NewPresentationContainer(a.business)
	if err != nil {
		return err
	}

	// cron
	a.cron = presentation.Cron
	a.cron.Start()

	// grpc api
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", a.cfg.GRPCSrv.Host, a.cfg.GRPCSrv.Port))
	if err != nil {
		return err
	}
	a.grpcAPI = presentation.GRPCSrv
	go func() {
		err = presentation.GRPCSrv.Serve(lis)
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
	if a.grpcAPI != nil {
		done := make(chan struct{})
		go func() {
			a.grpcAPI.GracefulStop()
			close(done)
		}()
		select {
		case <-timeoutCtx.Done():
			log.Printf("shutdown grpc timeout: %v", timeoutCtx.Err())
		case <-done:
			log.Println("gRPC server stopped")
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
			shutdownErr = wrapped
		}
	}

	// infrastructure
	if a.infra != nil {
		if err := a.infra.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown infrastructure: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		}
	}

	return shutdownErr
}

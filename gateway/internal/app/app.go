package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/config"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownTimeout = 20 * time.Second

type App struct {
	cfg *config.Config

	grpcCon  *grpc.ClientConn
	subSvc   pb.SubscriptionServiceClient
	weathSvc pb.WeatherServiceClient
	httpSrv  *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// setup grpc connection
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	a.grpcCon, err = grpc.NewClient(a.cfg.GRPCAddress(), opt)
	if err != nil {
		return err
	}

	// setup grpc clients
	a.subSvc = pb.NewSubscriptionServiceClient(a.grpcCon)
	a.weathSvc = pb.NewWeatherServiceClient(a.grpcCon)

	// setup http server
	a.httpSrv = a.setupHTTPServer()
	go func() {
		if err := a.httpSrv.ListenAndServe(); err != nil {
			log.Printf("http server: %v", err)
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

	// http server
	if a.httpSrv != nil {
		if err := a.httpSrv.Shutdown(timeoutCtx); err != nil {
			wrapped := fmt.Errorf("shutdown http server: %w", err)
			log.Println(wrapped)
			shutdownErr = wrapped
		} else {
			log.Println("HTTP Server Shutdown successfully")
		}
	}
	return shutdownErr
}

package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/config"
	pbsub "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	pbweath "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const shutdownTimeout = 20 * time.Second

type App struct {
	cfg *config.Config

	subGRPCCon    *grpc.ClientConn
	subGRPCClient pbsub.SubscriptionServiceClient

	weathGRPCCon    *grpc.ClientConn
	weathGRPCClient pbweath.WeatherServiceClient

	httpSrv *http.Server
}

func New(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

func (a *App) Run(ctx context.Context) error {
	var err error

	// setup subscription grpc connection
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	a.subGRPCCon, err = grpc.NewClient(a.cfg.SubSvc.Addr(), opt)
	if err != nil {
		return err
	}
	a.subGRPCClient = pbsub.NewSubscriptionServiceClient(a.subGRPCCon)

	// setup weather grpc connection
	a.weathGRPCCon, err = grpc.NewClient(a.cfg.WeatherSvc.Addr(), opt)
	if err != nil {
		return err
	}
	a.weathGRPCClient = pbweath.NewWeatherServiceClient(a.weathGRPCCon)

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

	// gRPC subscription connection
	if a.subGRPCCon != nil {
		if err := a.subGRPCCon.Close(); err != nil {
			wrapped := fmt.Errorf("close sub grpc connection: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("Subscription gRPC connection closed successfully")
		}
	}

	// gRPC weather connection
	if a.weathGRPCCon != nil {
		if err := a.weathGRPCCon.Close(); err != nil {
			wrapped := fmt.Errorf("close weather grpc connection: %w", err)
			log.Println(wrapped)
			if shutdownErr == nil {
				shutdownErr = wrapped
			}
		} else {
			log.Println("Weather gRPC connection closed successfully")
		}
	}

	return shutdownErr
}

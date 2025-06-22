package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/velosypedno/genesis-weather-api/internal/app"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

const shutdownTimeout = 20 * time.Second

func main() {
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	cfg := config.Load()
	a := app.New(cfg)
	err := a.Run()
	if err != nil {
		log.Fatal(err)
	}

	<-shutdownCtx.Done()

	timeoutCtx, cancel := context.WithTimeout(shutdownCtx, shutdownTimeout)
	defer cancel()
	err = a.Shutdown(timeoutCtx)
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/velosypedno/genesis-weather-api/internal/app"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	a := app.New(cfg)
	err = a.Run(ctx)
	if err != nil {
		log.Panic(err)
	}
}

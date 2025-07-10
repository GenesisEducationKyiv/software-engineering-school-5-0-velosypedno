package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/velosypedno/genesis-weather-api/gateway/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	app := app.New()
	err := app.Run(ctx)
	if err != nil {
		log.Panic(err)
	}
}

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	app := app.New(cfg)
	err = app.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}

}

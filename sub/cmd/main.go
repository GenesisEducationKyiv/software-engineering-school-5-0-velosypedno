package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
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

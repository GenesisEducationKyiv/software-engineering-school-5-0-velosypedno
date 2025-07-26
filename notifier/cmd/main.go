package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/notifier/internal/config"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigs {
			log.Printf("Received signal: %s\n", sig)
		}
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	app := app.New(cfg)
	err = app.Run(ctx)
	if err != nil {
		log.Panic(err)
	}
}

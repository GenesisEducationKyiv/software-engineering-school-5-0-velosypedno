package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/config"
	"go.uber.org/zap"
)

const service = "weather"

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		err := logger.Sync()
		if err != nil {
			panic(err)
		}
	}()
	logFactory := logging.NewFactory(logger, service)
	mainLogger := logFactory.ForPackage("main")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	mainLogger.Info("Loading config...")
	cfg, err := config.Load()
	if err != nil {
		mainLogger.Panic("Failed to load config", zap.Error(err))
	}
	mainLogger.Info("Config loaded")

	mainLogger.Info("Creating app...")
	app := app.New(cfg, logFactory)
	mainLogger.Info("App created")

	mainLogger.Info("Running app...")
	if err := app.Run(ctx); err != nil {
		mainLogger.Panic("Failed to run app", zap.Error(err))
	}
	mainLogger.Info("App stopped successfully")
}

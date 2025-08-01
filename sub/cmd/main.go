package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	"go.uber.org/zap"
)

const service = "sub"

func main() {
	logDir := os.Getenv("LOG_DIR")
	logFactory, err := logging.NewFactory(logDir, service)
	if err != nil {
		panic(err)
	}
	defer logFactory.Sync() //nolint:errcheck

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	mainLogger := logFactory.ForPackage("main")
	mainLogger.Info("Loading config...")
	cfg, err := config.Load()
	if err != nil {
		mainLogger.Panic("Failed to load config", zap.Error(err))
	}
	mainLogger.Info("Config loaded")

	mainLogger.Info("Creating app...")
	a := app.New(cfg, logFactory)
	mainLogger.Info("App created")

	mainLogger.Info("Running app...")
	if err := a.Run(ctx); err != nil {
		mainLogger.Panic("App failed", zap.Error(err))
	}
	mainLogger.Info("App stopped gracefully")
}

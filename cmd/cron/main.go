package main

import (
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/containers"
	"github.com/velosypedno/genesis-weather-api/internal/scheduler"
)

func main() {
	cfg := config.Load()
	taskContainer := containers.BuildTaskContainer(cfg)
	cron := scheduler.SetupScheduler(taskContainer)
	cron.Start()
	select {}
}

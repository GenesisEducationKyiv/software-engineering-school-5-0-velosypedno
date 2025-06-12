package main

import (
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
	"github.com/velosypedno/genesis-weather-api/internal/scheduler"
)

func main() {
	cfg := config.Load()
	taskContainer := ioc.BuildTaskContainer(cfg)
	cron := scheduler.SetupScheduler(taskContainer)
	cron.Start()
	log.Println("Cron tasks are scheduled")
	select {}
}

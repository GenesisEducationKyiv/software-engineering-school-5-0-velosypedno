package main

import (
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
	"github.com/velosypedno/genesis-weather-api/internal/scheduler"
)

func main() {
	cfg := config.Load()
	tasks := ioc.NewTasks(cfg)
	cron := scheduler.SetupScheduler(tasks)
	cron.Start()
	log.Println("Cron tasks are scheduled")
	select {}
}

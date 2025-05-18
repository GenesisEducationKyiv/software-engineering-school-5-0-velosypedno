package main

import (
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/container"
	"github.com/velosypedno/genesis-weather-api/internal/server"
)

func main() {
	cfg := config.Load()
	handlerContainer := container.BuildHandlerContainer(cfg)
	router := server.SetupRoutes(handlerContainer)
	err := router.Run(":" + cfg.PORT)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}

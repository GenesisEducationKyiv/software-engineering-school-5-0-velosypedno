package main

import (
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
	"github.com/velosypedno/genesis-weather-api/internal/server"
)

func main() {
	cfg := config.Load()
	handlers := ioc.NewHandlers(cfg)
	router := server.SetupRoutes(handlers)
	err := router.Run(":" + cfg.Port)
	if err != nil {
		log.Fatal("Server error:", err)
	}
}

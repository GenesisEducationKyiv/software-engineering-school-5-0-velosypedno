package main

import (
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/app"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	a := app.New(cfg)
	err = a.Run()
	if err != nil {
		log.Panic(err)
	}
}

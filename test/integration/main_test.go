//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/app"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/test/mock"
)

const apiURL = "http://127.0.0.1:8081"

var DB *sql.DB

func TestMain(m *testing.M) {
	// setup fake free weather API
	testWeatherAPI := mock.NewFreeWeatherAPI()
	defer testWeatherAPI.Close()

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg.FreeWeather.URL = testWeatherAPI.URL
	fmt.Println(cfg)

	// setup DB
	db, err := sql.Open(cfg.DB.Driver, cfg.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	// run app
	app := app.New(cfg)
	go app.Run(ctx)

	// run tests
	code := m.Run()
	cancel()

	os.Exit(code)
}

func clearDB() {
	_, err := DB.Exec("TRUNCATE subscriptions")
	if err != nil {
		log.Fatal(err)
	}
}

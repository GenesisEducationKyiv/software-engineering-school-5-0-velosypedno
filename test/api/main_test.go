//go:build integration

package api_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/config"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/test/mock"
	_ "github.com/lib/pq"
)

const apiURL = "http://127.0.0.1:8081"

var DB *sql.DB

func TestMain(m *testing.M) {
	// setup fake weather APIs
	freeWeatherAPI := mock.NewFreeWeatherAPI()
	defer freeWeatherAPI.Close()
	tomorrowAPI := mock.NewTomorrowAPI()
	defer tomorrowAPI.Close()
	vcAPI := mock.NewVisualCrossingAPI()
	defer vcAPI.Close()

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
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

//go:build integration
// +build integration

package integration_test

import (
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
	"github.com/velosypedno/genesis-weather-api/internal/server"
)

const serverTimeout = 10

var DB *sql.DB
var TestServer *httptest.Server

func TestMain(m *testing.M) {
	fmt.Println("Starting integration tests...")
	err := godotenv.Load("../.env.test")
	if err != nil {
		fmt.Println(err)
	}
	cfg := config.Load()
	fmt.Println(cfg)
	db, err := sql.Open(cfg.DbDriver, cfg.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	router := server.SetupRoutes(ioc.NewHandlers(cfg))

	testServer := httptest.NewServer(router)
	TestServer = testServer
	defer testServer.Close()

	code := m.Run()
	os.Exit(code)
}

func ClearDB() {
	_, err := DB.Exec("TRUNCATE subscriptions")
	if err != nil {
		log.Fatal(err)
	}
}

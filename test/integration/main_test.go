//go:build integration
// +build integration

package integration_test

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/app"
	"github.com/velosypedno/genesis-weather-api/internal/config"
)

const (
	invalidCity = "InvalidCity"
	apiURL      = "http://127.0.0.1:8081"
)

var DB *sql.DB

func TestMain(m *testing.M) {
	// setup fake weather API
	testWeatherAPI := startFakeWeatherAPI()
	defer testWeatherAPI.Close()

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	cfg.WeatherAPIBaseURL = testWeatherAPI.URL
	fmt.Println(cfg)

	// setup DB
	db, err := sql.Open(cfg.DbDriver, cfg.DSN())
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// run app
	app := app.New(cfg)
	app.Run()

	// run tests
	code := m.Run()
	app.Shutdown()

	os.Exit(code)
}

func clearDB() {
	_, err := DB.Exec("TRUNCATE subscriptions")
	if err != nil {
		log.Fatal(err)
	}
}

func startFakeWeatherAPI() *httptest.Server {
	handler := http.NewServeMux()

	handler.HandleFunc("/current.json", func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("q")
		if city == invalidCity {
			http.Error(w, `{"error": {"code": 1006, "message": "No matching location found."}}`, http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"current": {"temp_c": 20.0, "humidity": 80.0, "condition": {"text": "Sunny"}}}`))
	})

	httpServer := httptest.NewServer(handler)
	return httpServer
}

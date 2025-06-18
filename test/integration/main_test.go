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

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/ioc"
	"github.com/velosypedno/genesis-weather-api/internal/server"
)

const serverTimeout = 10
const invalidCity = "InvalidCity"

var DB *sql.DB
var TestServer *httptest.Server

func TestMain(m *testing.M) {
	fmt.Println("Starting integration tests...")
	err := godotenv.Load("../.env")
	if err != nil {
		fmt.Println(err)
	}

	testWeatherAPI := StartFakeWeatherAPI()
	defer testWeatherAPI.Close()

	err = os.Setenv("WEATHER_API_BASE_URL", testWeatherAPI.URL)
	if err != nil {
		log.Fatal(err)
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

func StartFakeWeatherAPI() *httptest.Server {
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

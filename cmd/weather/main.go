package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/velosypedno/genesis-weather-api/internal/handlers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open(os.Getenv("DB_DRIVER"), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	router := gin.Default()
	weatherHandler := handlers.NewWeatherHandler(repos.NewWeatherAPIRepo(os.Getenv("WEATHER_API_KEY")))
	router.GET("/api/weather", weatherHandler)

	API_PORT := os.Getenv("API_PORT")
	if API_PORT == "" {
		API_PORT = "8080"
	}
	router.Run(":" + API_PORT)
}

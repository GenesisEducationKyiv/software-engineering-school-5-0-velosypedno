package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB_DSN          string
	WEATHER_API_KEY string
	PORT            string
	DB_DRIVER       string
}

func Load() *Config {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	apiKey := os.Getenv("WEATHER_API_KEY")
	port := os.Getenv("PORT")
	driver := os.Getenv("DB_DRIVER")
	return &Config{
		DB_DSN:          dsn,
		WEATHER_API_KEY: apiKey,
		PORT:            port,
		DB_DRIVER:       driver,
	}
}

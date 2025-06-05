package config

import (
	"fmt"
	"os"
)

type Config struct {
	DB_DRIVER string
	DB_DSN    string
	PORT      string

	WEATHER_API_KEY string

	SMTP_HOST  string
	SMTP_PORT  string
	SMTP_USER  string
	SMTP_PASS  string
	EMAIL_FROM string
}

func Load() *Config {
	return &Config{
		DB_DSN: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		),
		DB_DRIVER: os.Getenv("DB_DRIVER"),
		PORT:      os.Getenv("PORT"),

		WEATHER_API_KEY: os.Getenv("WEATHER_API_KEY"),

		SMTP_HOST:  os.Getenv("SMTP_HOST"),
		SMTP_PORT:  os.Getenv("SMTP_PORT"),
		SMTP_USER:  os.Getenv("SMTP_USER"),
		SMTP_PASS:  os.Getenv("SMTP_PASS"),
		EMAIL_FROM: os.Getenv("EMAIL_FROM"),
	}
}

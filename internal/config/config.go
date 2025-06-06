package config

import (
	"fmt"
	"os"
)

type Config struct {
	DbDriver string
	DbDSN    string
	Port     string

	WeatherAPIKey string

	SMTPHost  string
	SMTPPort  string
	SMTPUser  string
	SMTPPass  string
	EmailFrom string
}

func Load() *Config {
	return &Config{
		DbDSN: fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
		),
		DbDriver: os.Getenv("DB_DRIVER"),
		Port:     os.Getenv("PORT"),

		WeatherAPIKey: os.Getenv("WEATHER_API_KEY"),

		SMTPHost:  os.Getenv("SMTP_HOST"),
		SMTPPort:  os.Getenv("SMTP_PORT"),
		SMTPUser:  os.Getenv("SMTP_USER"),
		SMTPPass:  os.Getenv("SMTP_PASS"),
		EmailFrom: os.Getenv("EMAIL_FROM"),
	}
}

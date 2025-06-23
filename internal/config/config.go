package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DbDriver string `envconfig:"DB_DRIVER" required:"true"`
	DbHost   string `envconfig:"DB_HOST" required:"true"`
	DbPort   string `envconfig:"DB_PORT" required:"true"`
	DbUser   string `envconfig:"DB_USER" required:"true"`
	DbPass   string `envconfig:"DB_PASSWORD" required:"true"`
	DbName   string `envconfig:"DB_NAME" required:"true"`

	Port string `envconfig:"PORT" required:"true"`

	FreeWeatherAPIKey     string `envconfig:"FREE_WEATHER_API_KEY" required:"true"`
	TomorrowWeatherAPIKey string `envconfig:"TOMORROW_WEATHER_API_KEY" required:"true"`

	SMTPHost  string `envconfig:"SMTP_HOST" required:"true"`
	SMTPPort  string `envconfig:"SMTP_PORT" required:"true"`
	SMTPUser  string `envconfig:"SMTP_USER" required:"true"`
	SMTPPass  string `envconfig:"SMTP_PASS" required:"true"`
	EmailFrom string `envconfig:"EMAIL_FROM" required:"true"`
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName,
	)
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

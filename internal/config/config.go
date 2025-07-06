package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type DBConfig struct {
	Driver string `envconfig:"DB_DRIVER" required:"true"`
	Host   string `envconfig:"DB_HOST" required:"true"`
	Port   string `envconfig:"DB_PORT" required:"true"`
	User   string `envconfig:"DB_USER" required:"true"`
	Pass   string `envconfig:"DB_PASSWORD" required:"true"`
	Name   string `envconfig:"DB_NAME" required:"true"`
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Pass, c.Name,
	)
}

type SMTPConfig struct {
	Host      string `envconfig:"SMTP_HOST" required:"true"`
	Port      string `envconfig:"SMTP_PORT" required:"true"`
	User      string `envconfig:"SMTP_USER" required:"true"`
	Pass      string `envconfig:"SMTP_PASS" required:"true"`
	EmailFrom string `envconfig:"EMAIL_FROM" required:"true"`
}

type RedisConfig struct {
	Host string `envconfig:"REDIS_HOST" required:"true"`
	Port string `envconfig:"REDIS_PORT" required:"true"`
	Pass string `envconfig:"REDIS_PASSWORD" required:"true"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type TomorrowWeatherConfig struct {
	Key string `envconfig:"TOMORROW_WEATHER_API_KEY" required:"true"`
	URL string `envconfig:"TOMORROW_API_BASE_URL" required:"true"`
}

type FreeWeatherConfig struct {
	Key string `envconfig:"FREE_WEATHER_API_KEY" required:"true"`
	URL string `envconfig:"WEATHER_API_BASE_URL" required:"true"`
}

type VisualCrossingConfig struct {
	Key string `envconfig:"VISUAL_CROSSING_API_KEY" required:"true"`
	URL string `envconfig:"VISUAL_CROSSING_API_BASE_URL" required:"true"`
}
type SrvConfig struct {
	Port         string `envconfig:"API_PORT" required:"true"`
	TemplatesDir string `envconfig:"TEMPLATES_DIR" required:"true"`
}

type Config struct {
	DB    DBConfig
	SMTP  SMTPConfig
	Srv   SrvConfig
	Redis RedisConfig

	TomorrowWeather TomorrowWeatherConfig
	FreeWeather     FreeWeatherConfig
	VisualCrossing  VisualCrossingConfig
}

func Load() (*Config, error) {
	var сfg Config
	if err := envconfig.Process("", &сfg); err != nil {
		return nil, err
	}

	return &сfg, nil
}

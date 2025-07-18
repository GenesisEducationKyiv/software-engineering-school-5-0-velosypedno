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

type SrvConfig struct {
	Port         string `envconfig:"API_PORT" required:"true"`
	TemplatesDir string `envconfig:"TEMPLATES_DIR" required:"true"`
}

type GRPCConfig struct {
	Port string `envconfig:"GRPC_PORT" required:"true"`
	Host string `envconfig:"GRPC_HOST" required:"true"`
}

type WeatherServiceConfig struct {
	Port string `envconfig:"WEATHER_SERVICE_PORT" required:"true"`
	Host string `envconfig:"WEATHER_SERVICE_HOST" required:"true"`
}

func (c WeatherServiceConfig) Addr() string {
	return c.Host + ":" + c.Port
}

type Config struct {
	DB       DBConfig
	SMTP     SMTPConfig
	Srv      SrvConfig
	GRPCSrv  GRPCConfig
	WeathSvc WeatherServiceConfig
}

func Load() (*Config, error) {
	var сfg Config
	if err := envconfig.Process("", &сfg); err != nil {
		return nil, err
	}

	return &сfg, nil
}

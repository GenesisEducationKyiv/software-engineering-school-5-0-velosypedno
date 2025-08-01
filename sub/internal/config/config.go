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

type RabbitMQConfig struct {
	Host string `envconfig:"RABBITMQ_HOST" required:"true"`
	Port string `envconfig:"RABBITMQ_PORT" required:"true"`
	User string `envconfig:"RABBITMQ_USER" required:"true"`
	Pass string `envconfig:"RABBITMQ_PASSWORD" required:"true"`
}

func (c RabbitMQConfig) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Pass, c.Host, c.Port)
}

type GRPCConfig struct {
	Port string `envconfig:"GRPC_PORT" required:"true"`
	Host string `envconfig:"GRPC_HOST" required:"true"`
}

func (c GRPCConfig) Addr() string {
	return c.Host + ":" + c.Port
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
	RabbitMQ RabbitMQConfig

	GRPCSrv  GRPCConfig
	WeathSvc WeatherServiceConfig
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

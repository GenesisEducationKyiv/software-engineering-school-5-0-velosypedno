package config

import "github.com/kelseyhightower/envconfig"

type WeatherServiceConfig struct {
	Port string `envconfig:"WEATHER_SERVICE_PORT" required:"true"`
	Host string `envconfig:"WEATHER_SERVICE_HOST" required:"true"`
}

func (c WeatherServiceConfig) Addr() string {
	return c.Host + ":" + c.Port
}

type SubServiceConfig struct {
	Port string `envconfig:"SUB_SERVICE_PORT" required:"true"`
	Host string `envconfig:"SUB_SERVICE_HOST" required:"true"`
}

func (c SubServiceConfig) Addr() string {
	return c.Host + ":" + c.Port
}

type Config struct {
	WeatherSvc WeatherServiceConfig
	SubSvc     SubServiceConfig

	APIGatewayPort string `envconfig:"API_GATEWAY_PORT" required:"true"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

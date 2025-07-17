package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

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
type HTTPConfig struct {
	Port string `envconfig:"HTTP_PORT" required:"true"`
	Host string `envconfig:"HTTP_HOST" required:"true"`
}

type GRPCConfig struct {
	Port string `envconfig:"GRPC_PORT" required:"true"`
	Host string `envconfig:"GRPC_HOST" required:"true"`
}

type Config struct {
	GRPCSrv GRPCConfig
	HTTPSrv HTTPConfig
	Redis   RedisConfig

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

package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	GRPCPort string `envconfig:"GRPC_PORT" required:"true"`
	GRPCHost string `envconfig:"GRPC_HOST" required:"true"`

	APIGatewayPort string `envconfig:"API_GATEWAY_PORT" required:"true"`
}

func (c *Config) GRPCAddress() string {
	return c.GRPCHost + ":" + c.GRPCPort
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

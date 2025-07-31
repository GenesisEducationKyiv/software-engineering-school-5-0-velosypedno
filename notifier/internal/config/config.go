package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type RabbitMQConfig struct {
	Host string `envconfig:"RABBITMQ_HOST" required:"true"`
	Port string `envconfig:"RABBITMQ_PORT" required:"true"`
	User string `envconfig:"RABBITMQ_USER" required:"true"`
	Pass string `envconfig:"RABBITMQ_PASSWORD" required:"true"`
}

func (c RabbitMQConfig) Addr() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.User, c.Pass, c.Host, c.Port)
}

type SMTPConfig struct {
	Host      string `envconfig:"SMTP_HOST" required:"true"`
	Port      string `envconfig:"SMTP_PORT" required:"true"`
	User      string `envconfig:"SMTP_USER" required:"true"`
	Pass      string `envconfig:"SMTP_PASS" required:"true"`
	EmailFrom string `envconfig:"EMAIL_FROM" required:"true"`
}

type HTTPSrvConfig struct {
	Port string `envconfig:"HTTP_PORT" required:"true"`
}

func (c HTTPSrvConfig) Addr() string {
	return ":" + c.Port
}

type Config struct {
	SMTP     SMTPConfig
	RabbitMQ RabbitMQConfig
	HTTPSrv  HTTPSrvConfig

	TemplatesDir string `envconfig:"TEMPLATES_DIR" required:"true"`
	LogDir       string `envconfig:"LOG_DIR" required:"true"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

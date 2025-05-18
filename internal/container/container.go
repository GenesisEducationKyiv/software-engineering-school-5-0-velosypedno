package container

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/handlers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

type HandlerContainer struct {
	WeatherGETHandler     gin.HandlerFunc
	SubscribePOSTHandler  gin.HandlerFunc
	ConfirmGETHandler     gin.HandlerFunc
	UnsubscribeGETHandler gin.HandlerFunc
}

func BuildHandlerContainer(c *config.Config) *HandlerContainer {
	db, err := sql.Open(c.DB_DRIVER, c.DB_DSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WEATHER_API_KEY)
	weatherService := services.NewWeatherService(weatherRepo)

	subRepo := repos.NewSubscriptionDBRepo(db)
	emailService := services.NewSmtpEmailService(c.SMTP_HOST, c.SMTP_PORT, c.SMTP_USER, c.SMTP_PASS, c.EMAIL_FROM)
	subService := services.NewSubscriptionService(subRepo, emailService)

	return &HandlerContainer{
		WeatherGETHandler:     handlers.NewWeatherGETHandler(weatherService),
		SubscribePOSTHandler:  handlers.NewSubscribePOSTHandler(subService),
		ConfirmGETHandler:     handlers.NewConfirmGETHandler(subService),
		UnsubscribeGETHandler: handlers.NewUnsubscribeGETHandler(subService),
	}
}

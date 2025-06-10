package ioc

import (
	"database/sql"
	"log"
	"net/http"

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
	db, err := sql.Open(c.DbDriver, c.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WeatherAPIKey, &http.Client{})
	weatherService := services.NewWeatherService(weatherRepo)

	subRepo := repos.NewSubscriptionDBRepo(db)
	emailService := services.NewSMTPEmailService(c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPass, c.EmailFrom)
	subService := services.NewSubscriptionService(subRepo, emailService)

	return &HandlerContainer{
		WeatherGETHandler:     handlers.NewWeatherGETHandler(weatherService),
		SubscribePOSTHandler:  handlers.NewSubscribePOSTHandler(subService),
		ConfirmGETHandler:     handlers.NewConfirmGETHandler(subService),
		UnsubscribeGETHandler: handlers.NewUnsubscribeGETHandler(subService),
	}
}

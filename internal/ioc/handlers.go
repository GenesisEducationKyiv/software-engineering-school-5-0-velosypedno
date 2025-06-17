package ioc

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/velosypedno/genesis-weather-api/internal/config"
	"github.com/velosypedno/genesis-weather-api/internal/email"
	"github.com/velosypedno/genesis-weather-api/internal/handlers"
	"github.com/velosypedno/genesis-weather-api/internal/mailers"
	"github.com/velosypedno/genesis-weather-api/internal/repos"
	"github.com/velosypedno/genesis-weather-api/internal/services"
)

const confirmTmpl = "confirm_sub.html"

type Handlers struct {
	WeatherGETHandler     gin.HandlerFunc
	SubscribePOSTHandler  gin.HandlerFunc
	ConfirmGETHandler     gin.HandlerFunc
	UnsubscribeGETHandler gin.HandlerFunc
}

func NewHandlers(c *config.Config) *Handlers {
	db, err := sql.Open(c.DbDriver, c.DbDSN)
	if err != nil {
		log.Fatal(err)
	}
	weatherRepo := repos.NewWeatherAPIRepo(c.WeatherAPIKey, &http.Client{})
	weatherService := services.NewWeatherService(weatherRepo)

	subRepo := repos.NewSubscriptionDBRepo(db)
	smtpEmailBackend := email.NewSMTPBackend(c.SMTPHost, c.SMTPPort, c.SMTPUser, c.SMTPPass, c.EmailFrom)
	subMailer := mailers.NewSubscriptionMailer(smtpEmailBackend, c.TemplatesDir, confirmTmpl)
	subService := services.NewSubscriptionService(subRepo, subMailer)

	return &Handlers{
		WeatherGETHandler:     handlers.NewWeatherGETHandler(weatherService),
		SubscribePOSTHandler:  handlers.NewSubscribePOSTHandler(subService),
		ConfirmGETHandler:     handlers.NewConfirmGETHandler(subService),
		UnsubscribeGETHandler: handlers.NewUnsubscribeGETHandler(subService),
	}
}

package services

import (
	"context"
	"fmt"
	"log"

	"github.com/velosypedno/genesis-weather-api/internal/models"
)

type activeSubsRepo interface {
	GetActivatedByFreq(freq models.Frequency) ([]models.Subscription, error)
}

type weatherMailer interface {
	SendCurrent(subscription models.Subscription, weather models.Weather) error
}

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (models.Weather, error)
}

type WeatherNotificationService struct {
	subRepo       activeSubsRepo
	weatherMailer weatherMailer
	weatherRepo   weatherRepo
}

func NewWeatherNotificationService(
	subRepo activeSubsRepo,
	weatherMailer weatherMailer,
	weatherRepo weatherRepo,
) *WeatherNotificationService {
	return &WeatherNotificationService{
		subRepo:       subRepo,
		weatherMailer: weatherMailer,
		weatherRepo:   weatherRepo,
	}
}

func (s *WeatherNotificationService) SendByFreq(freq models.Frequency) {
	subscriptions, err := s.subRepo.GetActivatedByFreq(freq)
	if err != nil {
		log.Println(fmt.Errorf("weather notification service: failed to get subscriptions, err:%v ", err))
		return
	}
	for _, sub := range subscriptions {
		weather, err := s.weatherRepo.GetCurrent(context.Background(), sub.City)
		if err != nil {
			err = fmt.Errorf("weather notification service: failed to get weather for %s, err:%v ", sub.City, err)
			log.Println(err)
			continue
		}
		if err := s.weatherMailer.SendCurrent(sub, weather); err != nil {
			err = fmt.Errorf("weather notification service: failed to send email to %s, err:%v ", sub.Email, err)
			log.Println(err)
			continue
		}
	}
}

package services

import (
	"context"
	"fmt"
	"log"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
)

type activeSubsRepo interface {
	GetActivatedByFreq(freq domain.Frequency) ([]domain.Subscription, error)
}

type weatherMailer interface {
	SendCurrent(subscription domain.Subscription, weather domain.Weather) error
}

type weatherRepo interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
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

func (s *WeatherNotificationService) SendByFreq(freq domain.Frequency) {
	subscriptions, err := s.subRepo.GetActivatedByFreq(freq)
	if err != nil {
		log.Println(fmt.Errorf("weather notification service: %v ", err))
		return
	}
	for _, sub := range subscriptions {
		weather, err := s.weatherRepo.GetCurrent(context.Background(), sub.City)
		if err != nil {
			err = fmt.Errorf("weather notification service: failed to get weather for %s, err:%v ",
				sub.City, err)
			log.Println(err)
			continue
		}
		if err := s.weatherMailer.SendCurrent(sub, weather); err != nil {
			err = fmt.Errorf("weather notification service: failed to send email to %s, err:%v ",
				sub.Email, err)
			log.Println(err)
			continue
		}
	}
}

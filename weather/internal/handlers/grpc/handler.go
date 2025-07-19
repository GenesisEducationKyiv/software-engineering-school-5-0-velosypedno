package handlers

import (
	"context"
	"time"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
)

type weatherService interface {
	GetCurrent(ctx context.Context, city string) (domain.Weather, error)
}

type WeathGRPCServer struct {
	pb.UnimplementedWeatherServiceServer

	weathSvc       weatherService
	requestTimeout time.Duration
}

func NewWeatherGRPCServer(weathSvc weatherService, requestTimeout time.Duration) *WeathGRPCServer {
	return &WeathGRPCServer{
		weathSvc:       weathSvc,
		requestTimeout: requestTimeout,
	}
}

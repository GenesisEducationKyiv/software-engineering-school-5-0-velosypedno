package handlers

import (
	"context"
	"time"

	"go.uber.org/zap"

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
	logger         *zap.Logger
}

func NewWeatherGRPCServer(
	weathSvc weatherService,
	requestTimeout time.Duration,
	logger *zap.Logger,
) *WeathGRPCServer {
	return &WeathGRPCServer{
		weathSvc:       weathSvc,
		requestTimeout: requestTimeout,
		logger:         logger,
	}
}

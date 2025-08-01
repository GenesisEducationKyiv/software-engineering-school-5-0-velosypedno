package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/weath/v1alpha1"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/weather/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *WeathGRPCServer) GetCurrent(ctx context.Context, req *pb.GetCurrentRequest) (*pb.GetCurrentResponse, error) {
	city := req.City
	if city == "" {
		return nil, status.Errorf(codes.InvalidArgument, "city is empty")
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.requestTimeout)
	defer cancel()
	weather, err := s.weathSvc.GetCurrent(ctxWithTimeout, city)
	if errors.Is(err, domain.ErrCityNotFound) {
		log.Println(fmt.Errorf("current weather grpc handler: %v", err))
		return nil, status.Errorf(codes.NotFound, "city not found")
	}
	if errors.Is(err, domain.ErrInternal) {
		log.Println(fmt.Errorf("current weather grpc handler: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get weather")
	}
	if errors.Is(err, domain.ErrWeatherUnavailable) {
		log.Println(fmt.Errorf("current weather grpc handler: %v", err))
		return nil, status.Errorf(codes.Unavailable, "weather unavailable")
	}
	if errors.Is(err, domain.ErrProviderUnreliable) {
		log.Println(fmt.Errorf("current weather grpc handler: %v", err))
		return nil, status.Errorf(codes.Unavailable, "weather provider is unreliable")
	}
	if err != nil {
		log.Println(fmt.Errorf("current weather grpc handler: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to get weather")
	}

	return &pb.GetCurrentResponse{
		Temperature: float32(weather.Temperature),
		Humidity:    float32(weather.Humidity),
		Description: weather.Description,
	}, nil
}

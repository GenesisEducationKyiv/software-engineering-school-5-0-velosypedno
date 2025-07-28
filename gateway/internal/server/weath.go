package server

import (
	"context"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WeathGrpcServer struct {
	pb.UnimplementedWeatherServiceServer
}

func NewWeathGrpcServer() *WeathGrpcServer {
	return &WeathGrpcServer{}
}

func (s *WeathGrpcServer) GetCurrent(_ context.Context, req *pb.GetCurrentRequest) (
	*pb.GetCurrentResponse, error,
) {
	if req.City == "InvalidCity" {
		return nil, status.Errorf(codes.NotFound, "city not found")
	}

	var humidity float32 = 46.8
	var temperature float32 = 25.9
	return &pb.GetCurrentResponse{
		Humidity:    humidity,
		Temperature: temperature,
		Description: "Sunny",
	}, nil
}

package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SubGRPCServer) Confirm(_ context.Context, req *pb.ConfirmRequest) (
	*pb.ConfirmResponse, error,
) {
	parsedToken, err := uuid.Parse(req.Token)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	err = s.subSvc.Activate(parsedToken)
	if errors.Is(err, domain.ErrSubNotFound) {
		log.Println(fmt.Errorf("confirm subscription grpc handler: %v", err))
		return nil, status.Errorf(codes.NotFound, "subscription with such token not found")
	}
	if errors.Is(err, domain.ErrInternal) {
		log.Println(fmt.Errorf("confirm subscription grpc handler: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to activate subscription")
	}
	if err != nil {
		log.Println(fmt.Errorf("confirm subscription grpc handler: %v", err))
		return nil, status.Errorf(codes.Internal, "failed to activate subscription")
	}
	return &pb.ConfirmResponse{
		Message: "successfully confirmed",
	}, nil
}

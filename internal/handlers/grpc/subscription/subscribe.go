package grpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/mail"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/domain"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/internal/services/subscription"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SubGrpcServer) Subscribe(_ context.Context, req *pb.SubscribeRequest) (
	*pb.SubscribeResponse, error,
) {
	err := validateSubscribeRequest(req)
	if err != nil {
		log.Println(fmt.Errorf("invalid subscribe request: %v", err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = s.subSvc.Subscribe(subsrv.SubscriptionInput{
		Email:     req.Email,
		Frequency: req.Frequency,
		City:      req.City,
	})
	if errors.Is(err, domain.ErrSubAlreadyExists) {
		return nil, status.Errorf(codes.AlreadyExists, "Email already subscribed")
	}
	if errors.Is(err, domain.ErrInternal) {
		return nil, status.Errorf(codes.Internal, "failed to create subscription")
	}
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "failed to create subscription")
	}
	return &pb.SubscribeResponse{
		Message: "Successfully subscribed",
	}, nil
}

func validateSubscribeRequest(req *pb.SubscribeRequest) error {
	if req.Email == "" || req.Frequency == "" || req.City == "" {
		return errors.New("all fields are required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email address")
	}

	switch domain.Frequency(req.Frequency) {
	case domain.FreqDaily, domain.FreqHourly:
		return nil
	default:
		return errors.New("frequency must be either 'daily' or 'hourly'")
	}
}

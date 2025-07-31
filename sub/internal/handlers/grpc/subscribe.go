package handlers

import (
	"context"
	"errors"
	"net/mail"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	subsrv "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/services/subscription"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SubGRPCServer) Subscribe(_ context.Context, req *pb.SubscribeRequest) (
	*pb.SubscribeResponse, error,
) {
	// step 1: logger setup
	logger := s.logger.With(
		zap.String("method", "Subscribe"),
	)
	logger.Info(
		"grpc method called",
		zap.String("email_hash", logging.HashEmail(req.Email)),
		zap.String("frequency", req.Frequency),
		zap.String("city", req.City),
	)

	// step 2: validate request
	err := validateSubscribeRequest(req)
	if err != nil {
		logger.Warn("invalid subscribe request", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	// step 3: create subscription
	err = s.subSvc.Subscribe(subsrv.SubscriptionInput{
		Email:     req.Email,
		Frequency: req.Frequency,
		City:      req.City,
	})

	// step 4: handle errors
	if errors.Is(err, domain.ErrSubAlreadyExists) {
		logger.Warn(
			"subscription already exists",
			zap.Error(err),
			zap.String("email_hash", logging.HashEmail(req.Email)),
			zap.String("city", req.City),
		)
		return nil, status.Errorf(codes.AlreadyExists, "Email already subscribed")
	}
	if errors.Is(err, domain.ErrInternal) || err != nil {
		logger.Error(
			"failed to create subscription",
			zap.Error(err),
			zap.String("email_hash", logging.HashEmail(req.Email)),
		)
		return nil, status.Errorf(codes.Internal, "failed to create subscription")
	}

	// step 5: return response
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

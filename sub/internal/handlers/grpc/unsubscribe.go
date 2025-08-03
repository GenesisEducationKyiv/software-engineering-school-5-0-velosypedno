package handlers

import (
	"context"
	"errors"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/domain"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *SubGRPCServer) Unsubscribe(_ context.Context, req *pb.UnsubscribeRequest) (
	*pb.UnsubscribeResponse, error,
) {
	// step 1: logger setup
	logger := s.logger.With(
		zap.String("method", "Unsubscribe"),
	)
	logger.Info(
		"grpc method called",
		zap.String("token", req.Token),
	)

	// step 2: validate request
	parsedToken, err := uuid.Parse(req.Token)
	if err != nil {
		logger.Warn("failed to parse token", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	// step 3: unsubscribe
	err = s.subSvc.Unsubscribe(parsedToken)

	// step 4: handle errors
	if errors.Is(err, domain.ErrSubNotFound) {
		logger.Warn(
			"subscription not found",
			zap.Error(err),
			zap.String("token", req.Token),
		)
		return nil, status.Errorf(codes.NotFound, "subscription with such token not found")
	}
	if errors.Is(err, domain.ErrInternal) || (err != nil) {
		logger.Error(
			"failed to unsubscribe subscription",
			zap.Error(err),
			zap.String("token", req.Token),
		)
		return nil, status.Errorf(codes.Internal, "failed to activate subscription")
	}

	// step 5: return response
	return &pb.UnsubscribeResponse{
		Message: "successfully unsubscribed",
	}, nil
}

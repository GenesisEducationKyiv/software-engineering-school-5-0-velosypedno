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

func (s *SubGRPCServer) Confirm(_ context.Context, req *pb.ConfirmRequest) (
	*pb.ConfirmResponse, error,
) {
	logger := s.logger.With(
		zap.String("method", "Confirm"),
	)
	logger.Info("grpc method called", zap.String("token", req.Token))
	parsedToken, err := uuid.Parse(req.Token)
	if err != nil {
		logger.Warn("failed to parse token", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	err = s.subSvc.Activate(parsedToken)
	if errors.Is(err, domain.ErrSubNotFound) {
		logger.Warn(
			"subscription not found",
			zap.Error(err),
			zap.String("token", req.Token),
		)
		return nil, status.Errorf(codes.NotFound, "subscription with such token not found")
	}
	if errors.Is(err, domain.ErrInternal) {
		logger.Error(
			"failed to activate subscription",
			zap.Error(err),
			zap.String("token", req.Token),
		)
		return nil, status.Errorf(codes.Internal, "failed to activate subscription")
	}
	if err != nil {
		logger.Error(
			"failed to activate subscription",
			zap.Error(err),
			zap.String("token", req.Token),
		)
		return nil, status.Errorf(codes.Internal, "failed to activate subscription")
	}
	return &pb.ConfirmResponse{
		Message: "successfully confirmed",
	}, nil
}

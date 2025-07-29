package services

import (
	"context"
	"fmt"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const timeout = 5 * time.Second

type SubscriptionInput struct {
	Email     string
	Frequency string
	City      string
}

type GRPCAdapter struct {
	client pb.SubscriptionServiceClient
	logger *zap.Logger
}

func NewGRPCAdapter(client pb.SubscriptionServiceClient, logger *zap.Logger) *GRPCAdapter {
	return &GRPCAdapter{
		client: client,
		logger: logger.With(zap.String("component", "GRPCAdapter")),
	}
}

func (a *GRPCAdapter) Subscribe(subInput SubscriptionInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger := a.logger.With(
		zap.String("method", "Subscribe"),
	)

	sub := pb.SubscribeRequest{
		Email:     subInput.Email,
		Frequency: subInput.Frequency,
		City:      subInput.City,
	}

	logger.Info("Sending Subscribe request",
		zap.String("email_hash", logging.HashEmail(sub.Email)),
	)
	_, err := a.client.Subscribe(ctx, &sub)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			logger.Error("unexpected gRPC error", zap.Error(err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		logger.Warn("gRPC Subscribe failed",
			zap.String("msg", st.Message()),
			zap.String("email_hash", logging.HashEmail(sub.Email)),
			zap.String("code", st.Code().String()),
		)

		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

	logger.Info("Subscribe successful", zap.String("email_hash", logging.HashEmail(sub.Email)))
	return nil
}

func (a *GRPCAdapter) Activate(token uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger := a.logger.With(
		zap.String("method", "Activate"),
	)

	logger.Info("Sending Confirm request", zap.String("token", token.String()))
	_, err := a.client.Confirm(ctx, &pb.ConfirmRequest{Token: token.String()})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			logger.Error("unexpected gRPC error", zap.Error(err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		logger.Warn("gRPC Confirm failed",
			zap.String("msg", st.Message()),
			zap.String("token", token.String()),
			zap.String("code", st.Code().String()),
		)

		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

	logger.Info("Confirm successful", zap.String("token", token.String()))
	return nil
}

func (a *GRPCAdapter) Unsubscribe(token uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger := a.logger.With(
		zap.String("method", "Unsubscribe"),
	)

	logger.Info("Sending Unsubscribe request", zap.String("token", token.String()))

	_, err := a.client.Unsubscribe(ctx, &pb.UnsubscribeRequest{Token: token.String()})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			logger.Error("unexpected gRPC error", zap.Error(err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		logger.Warn("gRPC Unsubscribe failed",
			zap.String("msg", st.Message()),
			zap.String("token", token.String()),
			zap.String("code", st.Code().String()),
		)

		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

	logger.Info("Unsubscribe successful", zap.String("token", token.String()))
	return nil
}

func gRPCToDomainError(code codes.Code) error {
	switch code {
	case codes.InvalidArgument:
		return domain.ErrSubInvalid
	case codes.AlreadyExists:
		return domain.ErrSubAlreadyExists
	case codes.NotFound:
		return domain.ErrSubNotFound
	case codes.Internal:
		return domain.ErrInternal
	default:
		return domain.ErrInternal
	}
}

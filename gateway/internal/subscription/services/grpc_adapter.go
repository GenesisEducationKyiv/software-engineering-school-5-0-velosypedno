package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/subscription/domain"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/google/uuid"
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
}

func NewGRPCAdapter(client pb.SubscriptionServiceClient) *GRPCAdapter {
	return &GRPCAdapter{client: client}
}

func (a *GRPCAdapter) Subscribe(subInput SubscriptionInput) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sub := pb.SubscribeRequest{
		Email:     subInput.Email,
		Frequency: subInput.Frequency,
		City:      subInput.City,
	}

	_, err := a.client.Subscribe(ctx, &sub)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			log.Println(fmt.Errorf("grpc adapter: %v", err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		log.Println(fmt.Errorf("grpc adapter: %s", st.Message()))
		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

	return nil
}

func (a *GRPCAdapter) Activate(token uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := a.client.Confirm(ctx, &pb.ConfirmRequest{Token: token.String()})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			log.Println(fmt.Errorf("grpc adapter: %v", err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		log.Println(fmt.Errorf("grpc adapter: %s", st.Message()))
		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}
	return nil
}

func (a *GRPCAdapter) Unsubscribe(token uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := a.client.Unsubscribe(ctx, &pb.UnsubscribeRequest{Token: token.String()})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			log.Println(fmt.Errorf("grpc adapter: %v", err))
			return fmt.Errorf("grpc adapter: %w", domain.ErrInternal)
		}

		log.Println(fmt.Errorf("grpc adapter: %s", st.Message()))
		return fmt.Errorf("grpc adapter: %w", gRPCToDomainError(st.Code()))
	}

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

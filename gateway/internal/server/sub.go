package server

import (
	"context"
	"log"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type subscription struct {
	email     string
	city      string
	freq      string
	confirmed bool
}

type SubGrpcServer struct {
	pb.UnimplementedSubscriptionServiceServer

	subscriptions map[string]subscription
}

func NewSubGrpcServer() *SubGrpcServer {
	return &SubGrpcServer{
		subscriptions: make(map[string]subscription),
	}
}

func (s *SubGrpcServer) Subscribe(_ context.Context, req *pb.SubscribeRequest) (
	*pb.SubscribeResponse, error,
) {
	token := uuid.New().String()
	for _, sub := range s.subscriptions {
		if sub.city == req.City && sub.email == req.Email {
			return nil, status.Errorf(codes.AlreadyExists,
				"subscription with such email and city already exists")
		}
	}
	s.subscriptions[token] = subscription{
		email:     req.Email,
		city:      req.City,
		freq:      req.Frequency,
		confirmed: false,
	}

	log.Printf("subscription: %s - %s - %s, token (%s)\n", req.Email, req.City, req.Frequency, token)
	return &pb.SubscribeResponse{
		Message: "successfully subscribed",
	}, nil
}

func (s *SubGrpcServer) Confirm(_ context.Context, req *pb.ConfirmRequest) (
	*pb.ConfirmResponse, error,
) {
	if uuid.Validate(req.Token) != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	for token, sub := range s.subscriptions {
		if sub.confirmed {
			continue
		}
		if token == req.Token {
			sub.confirmed = true
			return &pb.ConfirmResponse{
				Message: "successfully confirmed",
			}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "subscription with such token not found")
}

func (s *SubGrpcServer) Unsubscribe(_ context.Context, req *pb.UnsubscribeRequest) (
	*pb.UnsubscribeResponse, error,
) {
	if uuid.Validate(req.Token) != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token")
	}

	for token := range s.subscriptions {
		if token == req.Token {
			delete(s.subscriptions, token)
			return &pb.UnsubscribeResponse{
				Message: "successfully deleted",
			}, nil
		}
	}

	return nil, status.Errorf(codes.NotFound, "subscription with such token not found")
}

package main

import (
	"fmt"
	"log"
	"net"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/gateway/internal/server"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha1"
	"google.golang.org/grpc"
)

const port = 50100

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSubscriptionServiceServer(grpcServer, server.NewSubGrpcServer())
	pb.RegisterWeatherServiceServer(grpcServer, server.NewWeathGrpcServer())
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}

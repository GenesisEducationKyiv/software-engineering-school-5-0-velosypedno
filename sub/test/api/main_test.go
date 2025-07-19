//go:build integration

package api_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const apiURL = "http://127.0.0.1:8081"

var DB *sql.DB
var SubGRPCClient pb.SubscriptionServiceClient

func TestMain(m *testing.M) {
	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg)

	// setup DB
	db, err := sql.Open(cfg.DB.Driver, cfg.DB.DSN())
	if err != nil {
		log.Fatal(err)
	}
	DB = db
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	// run app
	app := app.New(cfg)
	go app.Run(ctx)

	// wait in grpc server start
	deadline := time.Now().Add(1 * time.Second)
	for {
		conn, err := net.Dial("tcp", cfg.GRPCSrv.Addr())
		if err == nil {
			_ = conn.Close()
			log.Println("gRPC server is ready")
			break
		}

		if time.Now().After(deadline) {
			log.Fatalf("gRPC server did not become ready within 1 second at %s", cfg.GRPCSrv.Addr())
		}
		time.Sleep(100 * time.Millisecond)
	}

	// setup grpc client
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.NewClient(cfg.GRPCSrv.Addr(), opt)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	SubGRPCClient = pb.NewSubscriptionServiceClient(conn)

	// run tests
	code := m.Run()
	cancel()

	os.Exit(code)
}

func clearDB() {
	_, err := DB.Exec("TRUNCATE subscriptions")
	if err != nil {
		log.Fatal(err)
	}
}

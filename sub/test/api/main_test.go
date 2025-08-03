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

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/logging"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/pkg/messaging"
	pb "github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/proto/sub/v1alpha2"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/app"
	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-velosypedno/sub/internal/config"
	_ "github.com/lib/pq"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const apiURL = "http://127.0.0.1:8081"

var SubGRPCClient pb.SubscriptionServiceClient

var (
	DB            *sql.DB
	RMQConnection *amqp.Connection
	RMQChannel    *amqp.Channel
	GRPCConn      *grpc.ClientConn
)

func closeConnections() {
	if GRPCConn != nil {
		if err := GRPCConn.Close(); err != nil {
			log.Println("failed to close gRPC conn:", err)
		}
	}
	if RMQChannel != nil {
		if err := RMQChannel.Close(); err != nil {
			log.Println("failed to close RMQ channel:", err)
		}
	}
	if RMQConnection != nil {
		if err := RMQConnection.Close(); err != nil {
			log.Println("failed to close RMQ conn:", err)
		}
	}
	if DB != nil {
		if err := DB.Close(); err != nil {
			log.Println("failed to close DB:", err)
		}
	}
}

func TestMain(m *testing.M) {

	// setup config
	cfg, err := config.Load()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(cfg)

	// setup messaging
	RMQConnection, RMQChannel, err = setupRMQ(cfg.RabbitMQ)
	if err != nil {
		closeConnections()
		log.Panic(err)
	}

	// setup DB
	DB, err = sql.Open(cfg.DB.Driver, cfg.DB.DSN())
	if err != nil {
		closeConnections()
		log.Panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// run app
	logFactory := logging.NewFakeFactory()
	app := app.New(cfg, logFactory)
	go func() {
		runErr := app.Run(ctx)
		if runErr != nil {
			closeConnections()
			log.Panic(runErr)
		}
	}()

	// wait on grpc server start
	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.Dial("tcp", cfg.GRPCSrv.Addr())
		if err == nil {
			_ = conn.Close()
			log.Println("gRPC server is ready")
			break
		}

		if time.Now().After(deadline) {
			log.Panicf("gRPC server did not become ready within 3 second at %s", cfg.GRPCSrv.Addr())
		}
		time.Sleep(100 * time.Millisecond)
	}

	// setup grpc client
	opt := grpc.WithTransportCredentials(insecure.NewCredentials())
	GRPCConn, err = grpc.NewClient(cfg.GRPCSrv.Addr(), opt)
	if err != nil {
		log.Panic(err)
	}
	SubGRPCClient = pb.NewSubscriptionServiceClient(GRPCConn)

	// run tests
	code := m.Run()
	cancel()
	closeConnections()
	os.Exit(code)
}

func clearDB() {
	_, err := DB.Exec("TRUNCATE subscriptions")
	if err != nil {
		log.Panic(err)
	}
}

func clearRMQ() {
	_, err := RMQChannel.QueuePurge(messaging.SubscribeQueueName, false)
	if err != nil {
		log.Panic(err)
	}
}

func setupRMQ(cfg config.RabbitMQConfig) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(cfg.Addr())
	if err != nil {
		return nil, nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return conn, nil, err
	}

	err = ch.ExchangeDeclare(
		messaging.ExchangeName,
		"direct",
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return conn, ch, err
	}

	q, err := ch.QueueDeclare(
		messaging.SubscribeQueueName, // name
		true,                         // durable
		false,                        // delete when unused
		false,                        // exclusive
		false,                        // no-wait
		nil,                          // arguments
	)
	if err != nil {
		return conn, ch, err
	}

	err = ch.QueueBind(
		q.Name,                        // queue name
		messaging.SubscribeRoutingKey, // routing key
		messaging.ExchangeName,        // exchange
		false,
		nil)
	if err != nil {
		return conn, ch, err
	}

	return conn, ch, nil
}

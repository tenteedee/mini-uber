package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenteedee/mini-uber/services/driver-service/internal/infrastructure/events"
	"github.com/tenteedee/mini-uber/services/driver-service/internal/infrastructure/grpc"
	"github.com/tenteedee/mini-uber/services/driver-service/internal/service"
	"github.com/tenteedee/mini-uber/shared/env"
	"github.com/tenteedee/mini-uber/shared/messaging"
	"github.com/tenteedee/mini-uber/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9092"

func main() {
	// initialize tracing
	tracerCfg := tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	}
	shutdown, err := tracing.InitTracer(tracerCfg)
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer shutdown(ctx)

	rabbitmqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
	driverService := service.NewService()

	// Handle OS signals for graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh
		cancel()
	}()

	lis, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("starting RabbitMQ connection on Driver service")

	// Initialize and start gRPC server
	grpcServer := grpcserver.NewServer()
	grpc.NewGrpcHandler(grpcServer, driverService)

	// initialize queue consumer
	consumer := events.NewTripEventConsumer(rabbitmq, driverService)
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("failed to listen to the message: %v", err)
		}
	}()

	log.Printf("Starting gRPC server Driver service on port %s", lis.Addr().String())

	// Start gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("failed to serve: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down the server...")
	grpcServer.GracefulStop()
}

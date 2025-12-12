package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/events"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/grpc"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/repository"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/service"
	"github.com/tenteedee/mini-uber/shared/env"
	"github.com/tenteedee/mini-uber/shared/messaging"
	"github.com/tenteedee/mini-uber/shared/tracing"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	// initialize tracing
	tracerCfg := tracing.Config{
		ServiceName:    "trip-service",
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
	inmemRepo := repository.NewInmemRepository()
	tripService := service.NewService(inmemRepo)

	// Handle OS signals for graceful shutdown
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
		<-signalChan
		cancel()
	}()

	listener, err := net.Listen("tcp", GrpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("starting RabbitMQ connection on Trip service")

	// Initialize TripEventPublisher
	publisher := events.NewTripEventPublisher(rabbitmq)

	// Initialize and start DriverEventConsumer
	consumer := events.NewTripEventConsumer(rabbitmq, tripService)
	go consumer.Listen()

	// Initialize and start gRPC server
	grpcServer := grpcserver.NewServer()
	grpc.NewgRPCHandler(grpcServer, tripService, publisher)

	log.Printf("starting Trip gRPC server on %s", listener.Addr().String())

	// Start gRPC server in a separate goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("failed to serve gRPC server: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Println("shutting down Trip gRPC server")
	grpcServer.GracefulStop()

}

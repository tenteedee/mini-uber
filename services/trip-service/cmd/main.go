package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/grpc"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/infrastructure/repository"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/service"
	grpcserver "google.golang.org/grpc"
)

var GrpcAddr = ":9093"

func main() {
	inmemRepo := repository.NewInmemRepository()
	service := service.NewService(inmemRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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
	grpcServer := grpcserver.NewServer()

	// initialize grpc server implementation
	grpc.NewgRPCHandler(grpcServer, service)

	log.Printf("starting Trip gRPC server on %s", listener.Addr().String())

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

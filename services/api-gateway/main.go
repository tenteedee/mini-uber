package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tenteedee/mini-uber/shared/env"
	"github.com/tenteedee/mini-uber/shared/messaging"
	"github.com/tenteedee/mini-uber/shared/tracing"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitmqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@localhost:5672/")
)

func main() {
	log.Println("Starting API Gateway")

	// initialize tracing
	tracerCfg := tracing.Config{
		ServiceName:    "api-gateway",
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

	mux := http.NewServeMux()

	// Initialize RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitmqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()
	log.Println("starting RabbitMQ connection on API Gateway")

	// initialize endpoints
	mux.Handle("POST /trip/preview", tracing.WrapHandlerFunc(enableCORS(handleTripPreview), "/trip/preview"))
	mux.Handle("POST /trip/start", tracing.WrapHandlerFunc(enableCORS(handleTripStart), "/trip/start"))
	mux.Handle("/ws/drivers", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDriverWebSocket(w, r, rabbitmq)
	}, "/ws/drivers"))
	mux.Handle("/ws/riders", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebSocket(w, r, rabbitmq)
	}, "/ws/riders"))
	mux.Handle("/webhook/stripe", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	}, "/webhook/stripe"))

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	serverError := make(chan error, 1)
	go func() {
		log.Printf("HTTP server listening on %s", httpAddr)
		serverError <- server.ListenAndServe()

	}()

	// Handle OS signals for graceful shutdown
	gfshutdown := make(chan os.Signal, 1)
	signal.Notify(gfshutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverError:
		log.Fatalf("could not start server: %v", err)
	case sig := <-gfshutdown:
		log.Printf("starting shutdown: %v", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("could not shutdown server: %v", err)
			server.Close()
		}
	}
}

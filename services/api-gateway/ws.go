package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	grpcclients "github.com/tenteedee/mini-uber/services/api-gateway/grpc_clients"
	"github.com/tenteedee/mini-uber/shared/contracts"
	"github.com/tenteedee/mini-uber/shared/messaging"
	pb "github.com/tenteedee/mini-uber/shared/proto/driver"
)

var (
	connManager = messaging.NewConnectionManager()
)

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userId := r.URL.Query().Get("userID")
	if userId == "" {
		log.Printf("Missing userID in query parameters")
		conn.Close()
		return
	}

	connManager.Add(userId, conn)
	defer connManager.Remove(userId)

	// queue consumers
	queues := []string{
		messaging.NotifyDriversNoDriversFoundQueue,
	}

	for _, qName := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, qName)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start queue consumer for queue %s: %v", qName, err)
			return
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Received message from rider %s: %s", userId, message)
	}

}

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userId := r.URL.Query().Get("userID")
	if userId == "" {
		log.Printf("Missing userID in query parameters")
		conn.Close()
		return
	}

	packageSlug := r.URL.Query().Get("packageSlug")
	if packageSlug == "" {
		log.Printf("Missing packageSlug in query parameters")
		conn.Close()
		return
	}

	connManager.Add(userId, conn)

	driverService, err := grpcclients.NewDriverServiceClient()
	if err != nil {
		log.Printf("Failed to create driver service client: %v", err)
		return
	}
	defer driverService.Close()

	// ensure driver is unregistered when the connection is closed
	defer func() {
		connManager.Remove(userId)

		_, err := driverService.Client.UnregisterDriver(r.Context(), &pb.RegisterDriverRequest{
			DriverId:    userId,
			PackageSlug: packageSlug,
		})
		if err != nil {
			log.Printf("Failed to unregister driver: %v", err)
		}

		log.Printf("Driver %s disconnected", userId)
	}()

	driverData, err := driverService.Client.RegisterDriver(r.Context(), &pb.RegisterDriverRequest{
		DriverId:    userId,
		PackageSlug: packageSlug,
	})
	if err != nil {
		log.Printf("Failed to register driver: %v", err)
		return
	}

	if err := connManager.SendMessage(userId, contracts.WSMessage{
		Type: contracts.DriverCmdRegister,
		Data: driverData.Driver,
	}); err != nil {
		log.Printf("Error reading message: %v", err)
		return
	}

	// queue consumers
	queues := []string{
		messaging.DriverCmdTripRequestQueue,
	}

	for _, qName := range queues {
		consumer := messaging.NewQueueConsumer(rb, connManager, qName)

		if err := consumer.Start(); err != nil {
			log.Printf("failed to start queue consumer for queue %s: %v", qName, err)
			return
		}
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		type driverMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var driverMsg driverMessage
		if err := json.Unmarshal(message, &driverMsg); err != nil {
			log.Printf("error unmarshalling driver message: %v", err)
			continue
		}

		log.Printf("Received message from driver %s: %s", userId, message)

		// handle different message types from driver
		switch driverMsg.Type {
		case contracts.DriverCmdLocation:
			if err := connManager.SendMessage(userId, contracts.WSMessage{
				Type: contracts.DriverCmdLocation,
				Data: driverMsg.Data,
			}); err != nil {
				log.Printf("Error sending update location message: %v", err)
			}
			continue
		case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
			if err := rb.PublishMessage(context.Background(), driverMsg.Type, contracts.AmqpMessage{
				OwnerID: userId,
				Data:    driverMsg.Data,
			}); err != nil {
				log.Printf("Error publishing driver trip response message: %v", err)
			}
		default:
			log.Printf("Unknown driver message type: %s", driverMsg.Type)
		}

	}
}

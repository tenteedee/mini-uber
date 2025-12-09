package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	grpcclients "github.com/tenteedee/mini-uber/services/api-gateway/grpc_clients"
	"github.com/tenteedee/mini-uber/shared/contracts"
	pb "github.com/tenteedee/mini-uber/shared/proto/driver"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleRidersWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocker upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	userId := r.URL.Query().Get("userID")
	if userId == "" {
		log.Printf("Missing userID in query parameters")
		conn.Close()
		return
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

func handleDriverWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocker upgrade failed: %v", err)
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

	driverService, err := grpcclients.NewDriverServiceClient()
	if err != nil {
		log.Printf("Failed to create driver service client: %v", err)
		return
	}
	defer driverService.Close()

	// ensure driver is unregistered when the connection is closed
	defer func() {
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

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: driverData.Driver,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Error reading message: %v", err)
		return
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

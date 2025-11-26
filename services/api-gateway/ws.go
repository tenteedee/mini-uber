package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/tenteedee/mini-uber/shared/contracts"
	"github.com/tenteedee/mini-uber/shared/util"
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
	if userId == "" {
		log.Printf("Missing packageSlug in query parameters")
		conn.Close()
		return
	}

	type Driver struct {
		Id             string `json:"id"`
		Name           string `json:"name"`
		ProfilePicture string `json:"profilePicture"`
		CarPlate       string `json:"carPlate"`
		PackageSlug    string `json:"packageSlug"`
	}

	msg := contracts.WSMessage{
		Type: "driver.cmd.register",
		Data: Driver{
			Id:             userId,
			Name:           "John Doe",
			ProfilePicture: util.GetRandomAvatar(1),
			CarPlate:       "XYZ-1234",
			PackageSlug:    packageSlug,
		},
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

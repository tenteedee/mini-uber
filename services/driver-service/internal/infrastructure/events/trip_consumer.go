package events

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tenteedee/mini-uber/services/driver-service/internal/domain"
	"github.com/tenteedee/mini-uber/shared/contracts"
	"github.com/tenteedee/mini-uber/shared/messaging"
)

type TripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.DriverService
}

func NewTripEventConsumer(rabbitmq *messaging.RabbitMQ, service domain.DriverService) *TripEventConsumer {
	return &TripEventConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *TripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(
		messaging.FindAvailableDriversQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var tripEvent contracts.AmqpMessage

			if err := json.Unmarshal(msg.Body, &tripEvent); err != nil {
				log.Printf("failed to unmarshal trip event message: %v", err)
				return err
			}

			var payload messaging.TripEventData
			if err := json.Unmarshal(tripEvent.Data, &payload); err != nil {
				log.Printf("failed to unmarshal trip event data: %v", err)
				return err
			}

			log.Printf("driver received message: %+v", payload)

			switch msg.RoutingKey {
			case contracts.TripEventCreated, contracts.TripEventDriverNotInterested:
				return c.handleFindAndNotifyDrivers(ctx, payload)
			}

			log.Printf("unknown trip event: %+v", payload)

			return nil
		})
}

func (c *TripEventConsumer) handleFindAndNotifyDrivers(ctx context.Context, payload messaging.TripEventData) error {
	suitableDrivers := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)
	log.Printf("found %v suitable drivers", len(suitableDrivers))

	if len(suitableDrivers) == 0 {
		if err := c.rabbitmq.PublishMessage(
			ctx,
			contracts.TripEventNoDriversFound,
			contracts.AmqpMessage{
				OwnerID: payload.Trip.UserID,
			},
		); err != nil {
			log.Printf("failed to publish message to exchange: %v", err)
			return err
		}

		log.Println("no suitable drivers found")
		return nil
	}

	randomIndex := rand.Intn(len(suitableDrivers))

	driver := suitableDrivers[randomIndex]

	marshalledEvent, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal trip event data: %v", err)
		return err
	}

	if err := c.rabbitmq.PublishMessage(
		ctx,
		contracts.DriverCmdTripRequest,
		contracts.AmqpMessage{
			OwnerID: driver,
			Data:    marshalledEvent,
		},
	); err != nil {
		log.Printf("failed to publish message to exchange: %v", err)
		return err
	}

	return nil
}

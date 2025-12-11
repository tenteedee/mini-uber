package events

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	"github.com/tenteedee/mini-uber/shared/contracts"
	"github.com/tenteedee/mini-uber/shared/messaging"
	pbd "github.com/tenteedee/mini-uber/shared/proto/driver"
)

type DriverEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewTripEventConsumer(rabbitmq *messaging.RabbitMQ, service domain.TripService) *DriverEventConsumer {
	return &DriverEventConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *DriverEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(
		messaging.DriverTripResponseQueue,
		func(ctx context.Context, msg amqp091.Delivery) error {
			var message contracts.AmqpMessage

			if err := json.Unmarshal(msg.Body, &message); err != nil {
				log.Printf("failed to unmarshal trip event message: %v", err)
				return err
			}

			var payload messaging.DriverTripResponseData
			if err := json.Unmarshal(message.Data, &payload); err != nil {
				log.Printf("failed to unmarshal trip event data: %v", err)
				return err
			}

			log.Printf("driver response received message: %+v", payload)

			switch msg.RoutingKey {
			case contracts.DriverCmdTripAccept:
				if err := c.handleTripAccepted(ctx, payload.TripID, payload.Driver); err != nil {
					log.Printf("failed to handle trip accept: %v", err)
					return err
				}
			case contracts.DriverCmdTripDecline:
				if err := c.handleTripDeclined(ctx, payload.TripID, payload.RiderId); err != nil {
					log.Printf("Failed to handle the trip decline: %v", err)
					return err
				}
				return nil
			}
			log.Printf("unknown trip event: %+v", payload)

			return nil
		})
}

func (c *DriverEventConsumer) handleTripAccepted(ctx context.Context, tripId string, driver *pbd.Driver) error {
	trip, err := c.service.GetTripById(ctx, tripId)
	if err != nil {
		return err
	}

	if trip == nil {
		log.Printf("trip not found: %s", tripId)
		return nil
	}

	if err := c.service.UpdateTrip(ctx, tripId, "accepted", driver); err != nil {
		log.Printf("failed to update trip: %v", err)
		return err
	}

	trip, err = c.service.GetTripById(ctx, tripId)
	if err != nil {
		return err
	}

	mashhalledTrip, err := json.Marshal(trip)
	if err != nil {
		log.Printf("failed to marshal trip: %v", err)
		return err
	}

	// notify the rider that the driver has been assigned
	if err := c.rabbitmq.PublishMessage(
		ctx,
		contracts.TripEventDriverAssigned,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    mashhalledTrip,
		},
	); err != nil {
		log.Printf("failed to publish trip driver assigned event: %v", err)
		return err
	}
	// TODO: notify the payment service to start a payment link
	return nil
}

func (c *DriverEventConsumer) handleTripDeclined(ctx context.Context, tripID string, riderID string) error {
	// When a driver declines, we should try to find another driver

	trip, err := c.service.GetTripById(ctx, tripID)
	if err != nil {
		return err
	}

	newPayload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := json.Marshal(newPayload)
	if err != nil {
		return err
	}

	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverNotInterested,
		contracts.AmqpMessage{
			OwnerID: riderID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return err
	}

	return nil
}

package events

import (
	"context"
	"encoding/json"

	"github.com/tenteedee/mini-uber/services/trip-service/internal/domain"
	"github.com/tenteedee/mini-uber/shared/contracts"
	"github.com/tenteedee/mini-uber/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rabbitmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rabbitmq,
	}
}

func (p *TripEventPublisher) PublishTripCreatedEvent(ctx context.Context, trip *domain.TripModel) error {
	payload := messaging.TripEventData{
		Trip: trip.ToProto(),
	}

	tripEventJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(
		ctx,
		contracts.TripEventCreated,
		contracts.AmqpMessage{
			OwnerID: trip.UserID,
			Data:    tripEventJSON,
		},
	)

}

package events

import (
	"context"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"github.com/tenteedee/mini-uber/shared/messaging"
)

type TripEventConsumer struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventConsumer(rabbitmq *messaging.RabbitMQ) *TripEventConsumer {
	return &TripEventConsumer{rabbitmq: rabbitmq}
}

func (c *TripEventConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages("hello", func(ctx context.Context, msg amqp091.Delivery) error {
		log.Printf("driver received message: %v", msg)
		return nil
	})
}

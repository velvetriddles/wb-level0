package nats

import (
	"encoding/json"
	"fmt"

	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/velvetriddles/wb-level0/internal/domain"
)

type Publisher struct {
	js     nats.JetStreamContext
	logger *slog.Logger
}

func NewPublisher(js nats.JetStreamContext, logger *slog.Logger) *Publisher {
	return &Publisher{
		js:     js,
		logger: logger,
	}
}

func (p *Publisher) PublishOrder(subject string, order *domain.Order) error {
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	_, err = p.js.Publish(subject, orderJSON)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Order published", slog.String("orderID", order.OrderUID))
	return nil
}

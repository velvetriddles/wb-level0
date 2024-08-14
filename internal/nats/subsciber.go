package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/nats.go"
	"github.com/velvetriddles/wb-level0/internal/domain"
	"github.com/velvetriddles/wb-level0/internal/service"
)

type Subscriber struct {
	js      nats.JetStreamContext
	logger  *slog.Logger
	service *service.OrderService
	sub     *nats.Subscription
	subject string
}

func NewSubscriber(js nats.JetStreamContext, logger *slog.Logger, service *service.OrderService) *Subscriber {
	return &Subscriber{
		js:      js,
		logger:  logger,
		service: service,
	}
}

func (s *Subscriber) Subscribe(subject string) error {
	sub, err := s.js.Subscribe(subject, func(msg *nats.Msg) {
		var order domain.Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			s.logger.Error("Failed to unmarshal order", slog.String("error", err.Error()))
			msg.Nak() //retry
			return
		}

		err = s.service.CreateOrder(&order)
		if err != nil {
			if _, ok := err.(validator.ValidationErrors); ok {
				s.logger.Error("Validation failed for order", slog.String("error", err.Error()))
				msg.Ack()
			} else {
				s.logger.Error("Failed to create order", slog.String("error", err.Error()))
				msg.Nak()
			}
			return
		}

		s.logger.Info("Order processed", slog.String("orderID", order.OrderUID))
		msg.Ack()
	}, nats.DeliverNew(), nats.ManualAck(), nats.AckWait(10*time.Second), nats.MaxDeliver(3)) // retries

	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	s.sub = sub
	s.subject = subject
	s.logger.Info("Subscribed to subject", slog.String("subject", subject))
	return nil
}

func (s *Subscriber) Unsubscribe() error {
	if s.sub == nil {
		return fmt.Errorf("not subscribed to any subject")
	}

	err := s.sub.Unsubscribe()
	if err != nil {
		return fmt.Errorf("failed to unsubscribe: %w", err)
	}

	s.logger.Info("Unsubscribed from subject", slog.String("subject", s.subject))
	s.sub = nil
	s.subject = ""
	return nil
}

package main

import (
	"github.com/nats-io/nats.go"
	"github.com/velvetriddles/wb-level0/internal/config"
	"github.com/velvetriddles/wb-level0/internal/domain/generator"
	"github.com/velvetriddles/wb-level0/internal/logger"
	natsclient "github.com/velvetriddles/wb-level0/internal/nats"
)

func main() {
	cfg, err := config.LoadConfig()
	logger := logger.NewLogger()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		return
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", "error", err)
		return
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.Error("Failed to create JetStream", "error", err)
		return
	}

	publisher := natsclient.NewPublisher(js, logger)

	order := generator.GenerateRandomOrder()

	if err := publisher.PublishOrder(cfg.NatsSubject, &order); err != nil {
		logger.Error("Failed to publish order", "error", err)
		return
	}

	logger.Info("Order published successfully", "subject", cfg.NatsSubject, "orderID", order.OrderUID)
}

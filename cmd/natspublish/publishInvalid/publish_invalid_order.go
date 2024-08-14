package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/velvetriddles/wb-level0/internal/config"
	"github.com/velvetriddles/wb-level0/internal/domain"
	"github.com/velvetriddles/wb-level0/internal/logger"
	natsclient "github.com/velvetriddles/wb-level0/internal/nats"
)

func generateInvalidOrder() *domain.Order {
	return &domain.Order{
		OrderUID:    "dsaddsadasdsasadsadsadsadsads",
		TrackNumber: "",
		Entry:       "",
		Delivery: domain.Delivery{
			Name:  "",
			Phone: "543543",
			Zip:   "432432432",
			City:  "",
			Email: "fdsfdsfdsf",
		},
		Payment: domain.Payment{
			Transaction:  "fdsfdsfds",
			RequestID:    "",
			Currency:     "fdsfdsfds",
			Provider:     "", 
			Amount:       -100,
			PaymentDt:    0,
			Bank:         "12332",
			DeliveryCost: -50,
			GoodsTotal:   0,
			CustomFee:    -10,
		},
		Items: []domain.Item{
			{
				ChrtID:      -1,
				TrackNumber: "",
				Price:       0,
				RID:         "321321",
				Name:        "",
				Sale:        200,
				Size:        "321321",
				TotalPrice:  -50,
				NmID:        0,
				Brand:       "",
				Status:      999,
			},
		},
		Locale:            "321321321",
		InternalSignature: "321321321",
		CustomerID:        "",
		DeliveryService:   "",
		Shardkey:          "321321312",
		SmID:              -1,
		DateCreated:       time.Now(),
		OofShard:          "321321",
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	l := logger.NewLogger()

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		l.Error("Failed to connect to NATS", "error", err)
		return
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		l.Error("Failed to create JetStream", "error", err)
		return
	}

	publisher := natsclient.NewPublisher(js, l)
	invalidOrder := generateInvalidOrder()

	if err := publisher.PublishOrder(cfg.NatsSubject, invalidOrder); err != nil {
		l.Error("Failed to publish", "error", err)
		return
	}

	l.Info("publish successfully", "subject", cfg.NatsSubject)
}

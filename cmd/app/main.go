package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/velvetriddles/wb-level0/internal/config"
	"github.com/velvetriddles/wb-level0/internal/delivery/rest/handlers"
	"github.com/velvetriddles/wb-level0/internal/logger"
	natsClient "github.com/velvetriddles/wb-level0/internal/nats"
	"github.com/velvetriddles/wb-level0/internal/repository/cache"
	"github.com/velvetriddles/wb-level0/internal/repository/postgres"
	"github.com/velvetriddles/wb-level0/internal/service"
)

func main() {

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.NewLogger()
	logger.Info("Config loaded", "config", cfg)

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	// pull for iptimization
	db.SetMaxOpenConns(10000)
	db.SetMaxIdleConns(500)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err != nil {
		logger.Error("Failed to connect to DB", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		logger.Error("Failed to ping DB", "error", err)
		os.Exit(1)
	}

	orderRepo := postgres.NewOrderRepository(db, logger)

	orderCache := cache.NewOrderCache(logger, orderRepo)
	if err := orderCache.Restore(); err != nil {
		logger.Error("Failed to restore cache", "error", err)
		os.Exit(1)
	}

	orderService := service.NewOrderService(orderRepo, orderCache, logger)

	orderHandler := handlers.NewOrderHandler(orderService, logger)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error("Failed to connect to NATS", "error", err)
		os.Exit(1)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.Error("Failed to create JetStream", "error", err)
		os.Exit(1)
	}

	streamConfig := &nats.StreamConfig{
		Name:     "ORDERS_STREAM",
		Subjects: []string{"orders.*"},
	}

	_, err = js.AddStream(streamConfig)
	if err != nil {
		logger.Error("Failed to create stream", slog.String("error", err.Error()))
		os.Exit(1)
	}

	subscriber := natsClient.NewSubscriber(js, logger, orderService)
	if err := subscriber.Subscribe(cfg.NatsSubject); err != nil {
		logger.Error("Failed to subscribe to NATS subject", "error", err)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.HandleFunc("/orders", orderHandler.ListOrders).Methods(http.MethodGet)
	r.HandleFunc("/orders/{id}", orderHandler.GetOrder).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:    cfg.HTTPPort,
		Handler: r,
	}

	go func() {
		logger.Info("Server starting", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
		}
	}()
	// publisher := natsClient.NewPublisher(js, logger)
	// order := generator.GenerateRandomOrder()
	// err = publisher.PublishOrder("orders.new", &order)

	//gracefull shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	if err := subscriber.Unsubscribe(); err != nil {
		logger.Error("Failed to unsubscribe from NATS", "error", err)
	}

	logger.Info("Server exited gracefully")
}

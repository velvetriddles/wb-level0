package cache

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/velvetriddles/wb-level0/internal/domain"
)

type OrderRepository interface {
	GetAllOrders() ([]*domain.Order, error)
}

type OrderCache struct {
	repo   OrderRepository
	cache  sync.Map
	logger *slog.Logger
}

func NewOrderCache(logger *slog.Logger, repo OrderRepository) *OrderCache {
	return &OrderCache{
		logger: logger,
		repo:   repo,
	}
}

func (c *OrderCache) Set(order *domain.Order) {
	c.cache.Store(order.OrderUID, order)
	c.logger.Info("Order added to cache",
		slog.String("orderID", order.OrderUID))
}

func (c *OrderCache) Get(id string) (*domain.Order, bool) {
	value, found := c.cache.Load(id)
	if found {
		c.logger.Info("Cache hit", slog.String("orderID", id))
		return value.(*domain.Order), true
	}
	c.logger.Info("Miss cache", slog.String("orderID", id))
	return nil, false
}

func (c *OrderCache) Restore() error {
	startTime := time.Now()
	c.logger.Info("Starting cache restore from DB")

	orders, err := c.repo.GetAllOrders()
	if err != nil {
		c.logger.Error("Failed to get orders for cache from DB",
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to restore cache: %w", err)
	}

	for _, order := range orders {
		c.cache.Store(order.OrderUID, order)
	}

	duration := time.Since(startTime)
	c.logger.Info("Cache restored",
		slog.Int("orderCount", len(orders)),
		slog.String("duration", duration.String()))
	return nil
}

func (c *OrderCache) Delete(id string) {
	c.cache.Delete(id)
	c.logger.Info("Order removed from cache",
		slog.String("orderID", id))
}

func (c *OrderCache) GetAll() []*domain.Order {
	var orders []*domain.Order
	c.cache.Range(func(key, value interface{}) bool {
		orders = append(orders, value.(*domain.Order))
		return true
	})
	return orders
}

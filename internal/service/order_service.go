package service

import (
	"fmt"
	"log/slog"

	"github.com/velvetriddles/wb-level0/internal/domain"
)

type OrderRepository interface {
	SaveOrder(order *domain.Order) error
	GetOrderByID(id string) (*domain.Order, error)
	GetAllOrders() ([]*domain.Order, error)
}

type OrderCache interface {
	Set(order *domain.Order)
	Get(id string) (*domain.Order, bool)
	GetAll() []*domain.Order
	Restore() error
}

type OrderService struct {
	repo   OrderRepository
	cache  OrderCache
	logger *slog.Logger
}

func NewOrderService(repo OrderRepository, cache OrderCache, logger *slog.Logger) *OrderService {
	return &OrderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *OrderService) GetAllOrders() ([]*domain.Order, error) {
	cachedOrders := s.cache.GetAll()
	if len(cachedOrders) > 0 {
		s.logger.Info("Retrieved all orders from cache",
			slog.Int("count", len(cachedOrders)))
		return cachedOrders, nil
	}


	orders, err := s.repo.GetAllOrders()
	if err != nil {
		s.logger.Error("Failed to get all orders from repository",
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}

	for _, order := range orders {
		s.cache.Set(order)
	}

	s.logger.Info("Retrieved and cached all orders from repository",
		slog.Int("count", len(orders)))
	return orders, nil
}

func (s *OrderService) GetOrder(id string) (*domain.Order, error) {
	if order, found := s.cache.Get(id); found {
		s.logger.Info("Order found in cache",
			slog.String("orderID", id))
		return order, nil
	}

	order, err := s.repo.GetOrderByID(id)
	if err != nil {
		s.logger.Error("Failed to get order from repository",
			slog.String("error", err.Error()),
			slog.String("orderID", id))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		s.logger.Info("Order not found",
			slog.String("orderID", id))
		return nil, nil
	}

	s.cache.Set(order)
	s.logger.Info("Order retrieved from repository and cached",
		slog.String("orderID", id))
	return order, nil
}

func (s *OrderService) CreateOrder(order *domain.Order) error {
	if err := order.Validate(); err != nil {
		s.logger.Error("Invalid order data",
			slog.String("error", err.Error()),
			slog.String("orderID", order.OrderUID))
		return fmt.Errorf("invalid order data: %w", err)
	}

	if err := s.repo.SaveOrder(order); err != nil {
		s.logger.Error("Failed to save order in repository",
			slog.String("error", err.Error()),
			slog.String("orderID", order.OrderUID))
		return fmt.Errorf("failed to save order: %w", err)
	}

	s.cache.Set(order)

	s.logger.Info("Order created and cached",
		slog.String("orderID", order.OrderUID))
	return nil
}

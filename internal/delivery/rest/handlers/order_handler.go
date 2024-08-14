package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/velvetriddles/wb-level0/internal/domain"
)

type OrderService interface {
	GetAllOrders() ([]*domain.Order, error)
	GetOrder(id string) (*domain.Order, error)
}

type OrderHandler struct {
	service   OrderService
	templates *template.Template
	logger    *slog.Logger
}

func NewOrderHandler(service OrderService, logger *slog.Logger) *OrderHandler {
	templates := template.Must(template.ParseGlob("./internal/templates/*.html"))
	return &OrderHandler{
		service:   service,
		templates: templates,
		logger:    logger,
	}
}

func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling request to list all orders",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path))

	orders, err := h.service.GetAllOrders()
	if err != nil {
		h.logger.Error("Failed to get all orders",
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Successfully retrieved orders",
		slog.Int("count", len(orders)))

	err = h.templates.ExecuteTemplate(w, "list.html", orders)
	if err != nil {
		h.logger.Error("Failed to execute template",
			slog.String("template", "list.html"),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	h.logger.Info("Handling request to get specific order",
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("orderID", id))

	order, err := h.service.GetOrder(id)
	if err != nil {
		h.logger.Error("Failed to get order",
			slog.String("orderID", id),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if order == nil {
		h.logger.Info("Order not found",
			slog.String("orderID", id))
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	h.logger.Info("Successfully retrieved order",
		slog.String("orderID", id))

	err = h.templates.ExecuteTemplate(w, "detail.html", order)
	if err != nil {
		h.logger.Error("Failed to execute template",
			slog.String("template", "detail.html"),
			slog.String("error", err.Error()))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

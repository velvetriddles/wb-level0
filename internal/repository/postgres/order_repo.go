package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "github.com/lib/pq"
	"github.com/velvetriddles/wb-level0/internal/domain"
)

type OrderRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewOrderRepository(db *sql.DB, logger *slog.Logger) *OrderRepository {
	return &OrderRepository{db: db, logger: logger}
}

func (r *OrderRepository) SaveOrder(order *domain.Order) error {
	r.logger.Info("Attempting to save order", slog.String("orderUID", order.OrderUID))

	tx, err := r.db.Begin()
	if err != nil {
		r.logger.Error("Failed to begin transaction", slog.String("error", err.Error()))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Order info
	_, err = tx.Exec(`
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		r.logger.Error("Failed to insert order info", slog.String("error", err.Error()))
		return fmt.Errorf("failed to insert order info: %w", err)
	}

	// Delivery
	_, err = tx.Exec(`
        INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		r.logger.Error("Failed to insert delivery info", slog.String("error", err.Error()))
		return fmt.Errorf("failed to insert delivery info: %w", err)
	}

	// Payment
	_, err = tx.Exec(`
        INSERT INTO payment (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		r.logger.Error("Failed to insert payment info", slog.String("error", err.Error()))
		return fmt.Errorf("failed to insert payment info: %w", err)
	}

	// Items
	for _, item := range order.Items {
		_, err = tx.Exec(`
            INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			r.logger.Error("Failed to insert item",
				slog.String("error", err.Error()),
				slog.Int("chrtID", item.ChrtID))
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction", slog.String("error", err.Error()))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Info("Successfully saved order", slog.String("orderUID", order.OrderUID))
	return nil
}

func (r *OrderRepository) GetOrderByID(orderUID string) (*domain.Order, error) {
	r.logger.Info("Attempting to get order by ID", slog.String("orderUID", orderUID))
	// Optimization with JOIN
	var order domain.Order
	err := r.db.QueryRow(`
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard, 
				d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
				p.transaction, p.request_id, p.currency, p.provider, p.amount, 
				p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        JOIN delivery d ON o.order_uid = d.order_uid
        JOIN payment p ON o.order_uid = p.order_uid
        WHERE o.order_uid = $1`, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
		&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
		&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
		&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank,
		&order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Info("Order not found", slog.String("orderUID", orderUID))
			return nil, nil
		}
		r.logger.Error("Failed to get order", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get Items
	rows, err := r.db.Query(`
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		r.logger.Error("Failed to query items", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Item
		err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			r.logger.Error("Failed to scan item", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	r.logger.Info("Successfully retrieved order", slog.String("orderUID", orderUID))
	return &order, nil
}

func (r *OrderRepository) GetAllOrders() ([]*domain.Order, error) {
	r.logger.Info("Attempting to get all orders")

	rows, err := r.db.Query(`
        SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, 
               o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
               d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
               p.transaction, p.request_id, p.currency, p.provider, p.amount, 
               p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
        FROM orders o
        JOIN delivery d ON o.order_uid = d.order_uid
        JOIN payment p ON o.order_uid = p.order_uid`)
	if err != nil {
		r.logger.Error("Failed to query all orders", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query all orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		var o domain.Order
		err := rows.Scan(
			&o.OrderUID, &o.TrackNumber, &o.Entry, &o.Locale, &o.InternalSignature,
			&o.CustomerID, &o.DeliveryService, &o.Shardkey, &o.SmID, &o.DateCreated, &o.OofShard,
			&o.Delivery.Name, &o.Delivery.Phone, &o.Delivery.Zip, &o.Delivery.City,
			&o.Delivery.Address, &o.Delivery.Region, &o.Delivery.Email,
			&o.Payment.Transaction, &o.Payment.RequestID, &o.Payment.Currency,
			&o.Payment.Provider, &o.Payment.Amount, &o.Payment.PaymentDt, &o.Payment.Bank,
			&o.Payment.DeliveryCost, &o.Payment.GoodsTotal, &o.Payment.CustomFee)
		if err != nil {
			r.logger.Error("Failed to scan order", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, &o)
	}
	// Get items singly
	itemRows, err := r.db.Query(`
        SELECT order_uid, chrt_id, track_number, price, rid, name, 
               sale, size, total_price, nm_id, brand, status
        FROM items`)
	if err != nil {
		r.logger.Error("Failed to query items", slog.String("error", err.Error()))
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer itemRows.Close()
	// put all items in map by orderuud
	itemMap := make(map[string][]domain.Item)
	for itemRows.Next() {
		var item domain.Item
		var orderUID string
		err := itemRows.Scan(
			&orderUID, &item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			r.logger.Error("Failed to scan item", slog.String("error", err.Error()))
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		itemMap[orderUID] = append(itemMap[orderUID], item)
	}
	// then we mapping items with orders by uuid
	for _, order := range orders {
		order.Items = itemMap[order.OrderUID]
	}

	r.logger.Info("Successfully retrieved all orders", slog.Int("count", len(orders)))
	return orders, nil
}

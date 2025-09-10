package repository

import (
	"context"
	"database/sql"
	"fmt"

	"order_service/internal/domain"

	"github.com/lib/pq"
)

// OrderRepository определяет интерфейс для работы с заказами в хранилище.
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, orderUID string) (*domain.Order, error)
}

// orderRepository — реализация OrderRepository с использованием PostgreSQL.
type orderRepository struct {
	db *sql.DB
}

// NewOrderRepository создает новый экземпляр orderRepository.
func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Create сохраняет заказ и связанные данные в базу данных.
func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	// Проверка уникальности order_uid
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM orders WHERE order_uid = $1)`, order.OrderUID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("postgres check exists error: %w", err)
	}
	if exists {
		return domain.ErrOrderUIDNotUnique
	}

	// Начать транзакцию
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("postgres begin tx error: %w", err)
	}
	defer tx.Rollback()

	// Вставка основного заказа
	_, err = tx.ExecContext(ctx, `
        INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrOrderUIDNotUnique // Дубликат на уровне БД
		}
		return fmt.Errorf("postgres insert order error: %w", err)
	}

	// Вставка данных доставки
	_, err = tx.ExecContext(ctx, `
        INSERT INTO deliveries (order_uid, name, phone, zip, city, address, region, email)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return fmt.Errorf("postgres insert delivery error: %w", err)
	}

	// Вставка данных оплаты
	_, err = tx.ExecContext(ctx, `
        INSERT INTO payments (order_uid, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return fmt.Errorf("postgres insert payment error: %w", err)
	}

	// Вставка товаров
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return fmt.Errorf("postgres insert item error: %w", err)
		}
	}

	// Подтвердить транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("postgres commit error: %w", err)
	}
	return nil
}

// GetByID получает заказ по order_uid.
func (r *orderRepository) GetByID(ctx context.Context, orderUID string) (*domain.Order, error) {
	order := &domain.Order{}

	// Получение основного заказа
	err := r.db.QueryRowContext(ctx, `
        SELECT order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        FROM orders WHERE order_uid = $1`, orderUID).
		Scan(&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("postgres order query error: %w", err)
	}

	// Получение данных доставки
	err = r.db.QueryRowContext(ctx, `
        SELECT name, phone, zip, city, address, region, email
        FROM deliveries WHERE order_uid = $1`, orderUID).
		Scan(&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
			&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("postgres delivery query error: %w", err)
	}

	// Получение данных оплаты
	err = r.db.QueryRowContext(ctx, `
        SELECT transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payments WHERE order_uid = $1`, orderUID).
		Scan(&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal, &order.Payment.CustomFee)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrOrderNotFound
		}
		return nil, fmt.Errorf("postgres payment query error: %w", err)
	}

	// Получение товаров
	rows, err := r.db.QueryContext(ctx, `
        SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
        FROM items WHERE order_uid = $1`, orderUID)
	if err != nil {
		return nil, fmt.Errorf("postgres items query error: %w", err)
	}
	defer rows.Close()

	order.Items = []domain.Item{}
	for rows.Next() {
		var item domain.Item
		err := rows.Scan(&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status)
		if err != nil {
			return nil, fmt.Errorf("postgres items scan error: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres items iteration error: %w", err)
	}

	return order, nil
}

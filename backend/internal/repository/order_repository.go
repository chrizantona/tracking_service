package repository

import (
	"database/sql"
	"errors"
	"time"
	"fmt"
	"backend/internal/entity"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrClientNotFound = errors.New("client not found")

type OrderRepository interface {
	Create(order *entity.Order) error
	GetByID(id uuid.UUID) (*entity.Order, error)
	GetAll() ([]*entity.Order, error)
	Update(order *entity.Order) error
	Delete(id uuid.UUID) error
}

type orderRepository struct {
	db     *sql.DB
	logger *zap.Logger 
}

func NewOrderRepository(db *sql.DB, logger *zap.Logger) OrderRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &orderRepository{
		db:     db,
		logger: logger, 
	}
}

func (r *orderRepository) Create(order *entity.Order) error {
	const op = "repository.OrderRepository.Create" 
	l := r.logger.With(zap.String("operation", op)) 
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM users WHERE id = $1", order.ClientID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Warn("client not found for new order", zap.String("client_id", order.ClientID.String()))
			return fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		l.Error("DB error checking client existence", zap.Error(err), zap.String("client_id", order.ClientID.String()))
		return fmt.Errorf("%s: failed checking client existence: %w", op, err)
	}

	if order.ID == uuid.Nil {
		order.ID = uuid.New()
		l.Debug("Generated new Order ID", zap.String("order_id", order.ID.String()))
	}

	now := time.Now().UTC() 
	order.CreatedAt = now
	order.UpdatedAt = now

	query := `
		INSERT INTO orders (id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.Exec(query,
		order.ID,
		order.ClientID,
		order.CourierID, 
		order.Status,
		order.DeliveryAddress,
		order.DeliveryCoords,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		l.Error("Failed to insert order", zap.Error(err), zap.String("order_id", order.ID.String()))
		return fmt.Errorf("%s: failed to insert order: %w", op, err) 
	}

	l.Info("Order created successfully", zap.String("order_id", order.ID.String()))
	return nil
}


func (r *orderRepository) GetByID(id uuid.UUID) (*entity.Order, error) {
	const op = "repository.OrderRepository.GetByID"
	l := r.logger.With(zap.String("operation", op), zap.String("order_id", id.String()))

	query := `
		SELECT id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	var order entity.Order
	err := row.Scan(
		&order.ID,
		&order.ClientID,
		&order.CourierID, 
		&order.Status,
		&order.DeliveryAddress,
		&order.DeliveryCoords,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Warn("Order not found") 
			return nil, ErrOrderNotFound 
		}
		l.Error("Failed to scan order", zap.Error(err))
		return nil, fmt.Errorf("%s: failed to scan order: %w", op, err) 
	}
	l.Debug("Order retrieved successfully") 
	return &order, nil
}

func (r *orderRepository) GetAll() ([]*entity.Order, error) {
	const op = "repository.OrderRepository.GetAll"
	l := r.logger.With(zap.String("operation", op))
	query := `
		SELECT id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC -- Example ordering
		-- Consider adding LIMIT and OFFSET for pagination here
	`
	rows, err := r.db.Query(query)
	if err != nil {
		l.Error("Failed to query all orders", zap.Error(err))
		return nil, fmt.Errorf("%s: failed to query all orders: %w", op, err)
	}
	defer rows.Close() 

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(
			&order.ID, &order.ClientID, &order.CourierID, &order.Status,
			&order.DeliveryAddress, &order.DeliveryCoords, &order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			l.Error("Failed to scan order row during GetAll", zap.Error(err))
			return nil, fmt.Errorf("%s: failed to scan order row: %w", op, err)
		}
		orders = append(orders, &order)
	}

	if err = rows.Err(); err != nil {
		l.Error("Error occurred during rows iteration for GetAll orders", zap.Error(err))
		return nil, fmt.Errorf("%s: error iterating order rows: %w", op, err)
	}

	l.Debug("Successfully retrieved all orders", zap.Int("count", len(orders)))
	return orders, nil
}

func (r *orderRepository) Update(order *entity.Order) error {
	const op = "repository.OrderRepository.Update"
	l := r.logger.With(zap.String("operation", op), zap.String("order_id", order.ID.String()))
	order.UpdatedAt = time.Now().UTC() 

	query := `
		UPDATE orders
		SET client_id = $2, courier_id = $3, status = $4, delivery_address = $5, delivery_coords = $6, updated_at = $7
		WHERE id = $1
	`
	res, err := r.db.Exec(query,
		order.ID,
		order.ClientID,
		order.CourierID, 
		order.Status,
		order.DeliveryAddress,
		order.DeliveryCoords,
		order.UpdatedAt,
	)
	if err != nil {
		l.Error("Failed to execute order update", zap.Error(err))
		return fmt.Errorf("%s: failed to execute order update: %w", op, err) 
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		l.Error("Failed to get rows affected after order update", zap.Error(err))
		return fmt.Errorf("%s: failed to get rows affected after update: %w", op, err) 
	}
	if rowsAffected == 0 {
		l.Warn("Attempted to update non-existent or unchanged order") 
		return ErrOrderNotFound 
	}
	l.Info("Order updated successfully") 
	return nil
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	const op = "repository.OrderRepository.Delete"
	l := r.logger.With(zap.String("operation", op), zap.String("order_id", id.String()))
	query := `DELETE FROM orders WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		l.Error("Failed to execute order delete", zap.Error(err))
		return fmt.Errorf("%s: failed to execute order delete: %w", op, err) 
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		l.Error("Failed to get rows affected after order delete", zap.Error(err))
		return fmt.Errorf("%s: failed to get rows affected after delete: %w", op, err) 
	}
	if rowsAffected == 0 {
		l.Warn("Attempted to delete non-existent order") 
		return ErrOrderNotFound 
	}
	l.Info("Order deleted successfully") 
	return nil
}
package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	return &orderRepository{db: db, logger: logger}
}

func (r *orderRepository) Create(order *entity.Order) error {
	const op = "OrderRepository.Create"
	l := r.logger.With(zap.String("op", op))

	// проверяем, что клиент есть
	var exists int
	if err := r.db.QueryRow("SELECT 1 FROM users WHERE id = $1", order.ClientID).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			l.Warn("client not found", zap.String("client_id", order.ClientID.String()))
			return fmt.Errorf("%s: %w", op, ErrClientNotFound)
		}
		l.Error("db error checking client", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	// генерируем ID, если нужно
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}

	now := time.Now().UTC()
	order.CreatedAt = now
	order.UpdatedAt = now

	// парсим "lat,lon"
	parts := strings.Split(order.DeliveryCoords, ",")
	if len(parts) != 2 {
		return fmt.Errorf("%s: invalid coords format", op)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("%s: parse lat: %w", op, err)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("%s: parse lon: %w", op, err)
	}

	// вставляем с помощью PostGIS-функции
	query := `
		INSERT INTO orders (
			id, client_id, courier_id, status,
			delivery_address,
			delivery_coords,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5,
			ST_SetSRID(ST_MakePoint($6, $7), 4326),
			$8, $9
		)
	`
	_, err = r.db.Exec(query,
		order.ID,
		order.ClientID,
		order.CourierID,
		order.Status,
		order.DeliveryAddress,
		lon, lat,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		l.Error("failed to insert order", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	l.Info("order created", zap.String("order_id", order.ID.String()))
	return nil
}

func (r *orderRepository) GetByID(id uuid.UUID) (*entity.Order, error) {
	const op = "OrderRepository.GetByID"
	l := r.logger.With(zap.String("op", op), zap.String("order_id", id.String()))

	query := `
		SELECT
			id, client_id, courier_id, status,
			delivery_address,
			ST_AsText(delivery_coords) AS delivery_coords,
			created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)

	var order entity.Order
	if err := row.Scan(
		&order.ID,
		&order.ClientID,
		&order.CourierID,
		&order.Status,
		&order.DeliveryAddress,
		&order.DeliveryCoords,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		l.Error("failed to scan order", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	l.Debug("order fetched", zap.String("order_id", order.ID.String()))
	return &order, nil
}

func (r *orderRepository) GetAll() ([]*entity.Order, error) {
	const op = "OrderRepository.GetAll"
	l := r.logger.With(zap.String("op", op))

	query := `
		SELECT
			id, client_id, courier_id, status,
			delivery_address,
			ST_AsText(delivery_coords) AS delivery_coords,
			created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		l.Error("failed to query orders", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var list []*entity.Order
	for rows.Next() {
		var order entity.Order
		if err := rows.Scan(
			&order.ID,
			&order.ClientID,
			&order.CourierID,
			&order.Status,
			&order.DeliveryAddress,
			&order.DeliveryCoords,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			l.Error("failed to scan order row", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		list = append(list, &order)
	}
	if err := rows.Err(); err != nil {
		l.Error("rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	l.Debug("all orders fetched", zap.Int("count", len(list)))
	return list, nil
}

func (r *orderRepository) Update(order *entity.Order) error {
	const op = "OrderRepository.Update"
	l := r.logger.With(zap.String("op", op), zap.String("order_id", order.ID.String()))

	order.UpdatedAt = time.Now().UTC()

	parts := strings.Split(order.DeliveryCoords, ",")
	if len(parts) != 2 {
		return fmt.Errorf("%s: invalid coords format", op)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return fmt.Errorf("%s: parse lat: %w", op, err)
	}
	lon, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return fmt.Errorf("%s: parse lon: %w", op, err)
	}

	query := `
		UPDATE orders SET
			client_id        = $2,
			courier_id       = $3,
			status           = $4,
			delivery_address = $5,
			delivery_coords  = ST_SetSRID(ST_MakePoint($6, $7), 4326),
			updated_at       = $8
		WHERE id = $1
	`
	res, err := r.db.Exec(query,
		order.ID,
		order.ClientID,
		order.CourierID,
		order.Status,
		order.DeliveryAddress,
		lon, lat,
		order.UpdatedAt,
	)
	if err != nil {
		l.Error("failed to update order", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if n == 0 {
		return ErrOrderNotFound
	}
	l.Info("order updated", zap.String("order_id", order.ID.String()))
	return nil
}

func (r *orderRepository) Delete(id uuid.UUID) error {
	const op = "OrderRepository.Delete"
	l := r.logger.With(zap.String("op", op), zap.String("order_id", id.String()))

	res, err := r.db.Exec("DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		l.Error("failed to delete order", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if n == 0 {
		return ErrOrderNotFound
	}
	l.Info("order deleted", zap.String("order_id", id.String()))
	return nil
}

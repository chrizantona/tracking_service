package repository

import (
	"database/sql"
	"errors"
	"time"

	"backend/internal/entity"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(order *entity.Order) error
	GetByID(id string) (*entity.Order, error)
	GetAll() ([]*entity.Order, error)
	Update(order *entity.Order) error
	Delete(id string) error
}

type orderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *entity.Order) error {
	var exists int
	err := r.db.QueryRow("SELECT 1 FROM users WHERE id = $1", order.ClientID).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("client does not exist")
		}
		return err
	}

	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now

	query := `
		INSERT INTO orders (id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.db.Exec(query, order.ID, order.ClientID, order.CourierID, order.Status, order.DeliveryAddress, order.DeliveryCoords, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *orderRepository) GetByID(id string) (*entity.Order, error) {
	query := `
		SELECT id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	row := r.db.QueryRow(query, id)
	var order entity.Order
	var courierID sql.NullString
	err := row.Scan(&order.ID, &order.ClientID, &courierID, &order.Status, &order.DeliveryAddress, &order.DeliveryCoords, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, err
	}
	if courierID.Valid {
		order.CourierID = &courierID.String
	}
	return &order, nil
}

func (r *orderRepository) GetAll() ([]*entity.Order, error) {
	query := `
		SELECT id, client_id, courier_id, status, delivery_address, delivery_coords, created_at, updated_at
		FROM orders
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		var order entity.Order
		var courierID sql.NullString
		if err := rows.Scan(&order.ID, &order.ClientID, &courierID, &order.Status, &order.DeliveryAddress, &order.DeliveryCoords, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		if courierID.Valid {
			order.CourierID = &courierID.String
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *orderRepository) Update(order *entity.Order) error {
	order.UpdatedAt = time.Now()
	query := `
		UPDATE orders
		SET client_id = $2, courier_id = $3, status = $4, delivery_address = $5, delivery_coords = $6, updated_at = $7
		WHERE id = $1
	`
	res, err := r.db.Exec(query, order.ID, order.ClientID, order.CourierID, order.Status, order.DeliveryAddress, order.DeliveryCoords, order.UpdatedAt)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *orderRepository) Delete(id string) error {
	query := `DELETE FROM orders WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("order not found")
	}
	return nil
}

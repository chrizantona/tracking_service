package entity

import (
	"time"
	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusAssigned  OrderStatus = "ASSIGNED"
	StatusInTransit OrderStatus = "IN_TRANSIT"
	StatusDelivered OrderStatus = "DELIVERED"
	StatusCanceled  OrderStatus = "CANCELED"
)

type Order struct {
	ID              uuid.UUID      `json:"id"`
	ClientID        uuid.UUID   `json:"client_id"`
	CourierID       *uuid.UUID   `json:"courier_id,omitempty"` 
	Status          OrderStatus `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	DeliveryCoords  string      `json:"delivery_coords"` 
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

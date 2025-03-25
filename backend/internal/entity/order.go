package entity

import (
	"time"
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
	ID              string      `json:"id"`
	ClientID        string      `json:"client_id"`
	CourierID       *string     `json:"courier_id,omitempty"` 
	Status          OrderStatus `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	DeliveryCoords  string      `json:"delivery_coords"` 
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

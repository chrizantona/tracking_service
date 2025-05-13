package entity

import (
	"fmt"
	"strconv"
	"strings"
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
	ID              uuid.UUID   `json:"id"`
	ClientID        uuid.UUID   `json:"client_id"`
	CourierID       *uuid.UUID  `json:"courier_id,omitempty"`
	Status          OrderStatus `json:"status"`
	DeliveryAddress string      `json:"delivery_address"`
	DeliveryCoords  string      `json:"delivery_coords"` 
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

func (o *Order) ParseCoords() (lat, lon float64, err error) {
	parts := strings.Split(o.DeliveryCoords, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid coords format: %q", o.DeliveryCoords)
	}
	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse latitude: %w", err)
	}
	lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parse longitude: %w", err)
	}
	return lat, lon, nil
}

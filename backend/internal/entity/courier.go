package entity

import (
	"github.com/google/uuid"
)

type CourierStatus string

const (
	CourierStatusAvailable CourierStatus = "AVAILABLE"
	CourierStatusBusy      CourierStatus = "BUSY"
	CourierStatusOffline   CourierStatus = "OFFLINE"
)

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

type Courier struct {
	UserID   uuid.UUID     `db:"user_id" json:"user_id"`
	Name     string        `db:"name" json:"name"`
	Status   CourierStatus `db:"status" json:"status"`
	Location *Coordinates  `db:"location" json:"location"`
	Rating   float64       `db:"rating" json:"rating"`
}
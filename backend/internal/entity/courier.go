package entity

import (
	"github.com/google/uuid"
)

type CourierStatus string

const (
	CourierAvailable CourierStatus = "AVAILABLE"
	CourierBusy      CourierStatus = "BUSY"
	CourierOffline   CourierStatus = "OFFLINE"
)


type Geometry string

type Courier struct {
	UserID   uuid.UUID     `json:"user_id"`   
	Name     string        `json:"name"`      
	Status   CourierStatus `json:"status"`    
	Location Geometry      `json:"location"`  
	Rating   float64       `json:"rating"`    
}
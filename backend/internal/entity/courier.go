package entity

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type CourierStatus string

const (
	CourierStatusAvailable CourierStatus = "AVAILABLE"
	CourierStatusBusy      CourierStatus = "BUSY"
	CourierStatusOffline   CourierStatus = "OFFLINE"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Courier struct {
	UserID   uuid.UUID     `db:"user_id" json:"user_id"`
	Name     string        `db:"name" json:"name"`
	Status   CourierStatus `db:"status" json:"status"`
	Location *Coordinates  `db:"location" json:"location"`
	Rating   float64       `db:"rating" json:"rating"`
}


func CoordinatesToEWKB(c *Coordinates) ([]byte, error) {
	if c == nil {
		return nil, nil
	}
	p := geom.NewPoint(geom.XY).SetSRID(4326)
	if _, err := p.SetCoords(geom.Coord{c.Longitude, c.Latitude}); err != nil {
		return nil, fmt.Errorf("set coords: %w", err)
	}
	return ewkb.Marshal(p, binary.LittleEndian)
}


func EWKBToCoordinates(b []byte) (*Coordinates, error) {
	if b == nil {
		return nil, nil
	}
	g, err := ewkb.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("unmarshal EWKB: %w", err)
	}
	pt, ok := g.(*geom.Point)
	if !ok {
		return nil, fmt.Errorf("expected Point, got %T", g)
	}
	coords := pt.Coords()
	return &Coordinates{Latitude: coords.Y(), Longitude: coords.X()}, nil
}

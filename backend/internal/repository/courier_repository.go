package repository

import (
	"backend/internal/entity"
	"database/sql" 
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type CourierRepository interface {
	GetByID(id uuid.UUID) (*entity.Courier, error)
	Update(courier *entity.Courier) error
	FindNearestAvailable(latitude, longitude float64, radius float64) ([]*entity.Courier, error)
}

type courierRepo struct {
	db     *sql.DB 
	logger *zap.Logger
}

func NewCourierRepository(db *sql.DB, logger *zap.Logger) CourierRepository { 
	return &courierRepo{
		db:     db,
		logger: logger,
	}
}

func coordinatesToEWKB(coords *entity.Coordinates) ([]byte, error) {
	if coords == nil {
		return nil, nil
	}
    p := geom.NewPoint(geom.XY).SetSRID(4326)
    p, err := p.SetCoords(geom.Coord{coords.Longitude, coords.Latitude})
    if err != nil {
        return nil, fmt.Errorf("failed to set coords for geom.Point: %w", err)
    }
	return ewkb.Marshal(p, binary.LittleEndian)
}

func ewkbToCoordinates(b []byte) (*entity.Coordinates, error) {
	if b == nil {
		return nil, nil
	}
	geo, err := ewkb.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal EWKB: %w", err)
	}
	point, ok := geo.(*geom.Point)
	if !ok {
		return nil, fmt.Errorf("expected Point, got %T", geo)
	}
    coords := point.Coords()
	return &entity.Coordinates{Latitude: coords.Y(), Longitude: coords.X()}, nil
}

func (r *courierRepo) GetByID(id uuid.UUID) (*entity.Courier, error) {
	query := `SELECT user_id, name, status, location, rating FROM couriers WHERE user_id = $1`
	courier := &entity.Courier{}
	var locationBytes []byte

	err := r.db.QueryRow(query, id).Scan(&courier.UserID, &courier.Name, &courier.Status, &locationBytes, &courier.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("courier not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("courier not found")
		}
		r.logger.Error("failed to get courier", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get courier from db: %w", err)
	}

	courier.Location, err = ewkbToCoordinates(locationBytes)
	if err != nil {
		r.logger.Error("failed to convert location from EWKB", zap.String("id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to process courier location: %w", err)
	}

	return courier, nil
}

func (r *courierRepo) Update(courier *entity.Courier) error {
	locationBytes, err := coordinatesToEWKB(courier.Location)
	if err != nil {
		r.logger.Error("failed to convert location to EWKB", zap.String("id", courier.UserID.String()), zap.Error(err))
		return fmt.Errorf("failed to process courier location for update: %w", err)
	}

	query := `UPDATE couriers SET name=$2, status=$3, location=$4, rating=$5 WHERE user_id=$1`
	res, err := r.db.Exec(query, courier.UserID, courier.Name, courier.Status, locationBytes, courier.Rating)
	if err != nil {
		r.logger.Error("failed to update courier", zap.String("id", courier.UserID.String()), zap.Error(err))
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("failed to get rows affected", zap.Error(err))
		return err
	}
	if affected == 0 {
		r.logger.Warn("no rows updated", zap.String("id", courier.UserID.String()))
		return fmt.Errorf("courier not found")
	}
	r.logger.Info("courier updated", zap.String("id", courier.UserID.String()))
	return nil
}

func (r *courierRepo) FindNearestAvailable(latitude, longitude float64, radius float64) ([]*entity.Courier, error) {
	query := `
		SELECT user_id, name, status, location, rating
		FROM couriers
		WHERE status = $1 AND ST_DWithin(location, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4)
	`

    rows, err := r.db.Query(query, entity.CourierStatusAvailable, longitude, latitude, radius)
    if err != nil {
        r.logger.Error("failed to find nearest available couriers", zap.Error(err))
        return nil, fmt.Errorf("failed to find nearest available couriers: %w", err)
    }
    defer rows.Close()

    var couriers []*entity.Courier
    for rows.Next() {
        courier := &entity.Courier{}
        var locationBytes []byte

        err := rows.Scan(&courier.UserID, &courier.Name, &courier.Status, &locationBytes, &courier.Rating)
        if err != nil {
            r.logger.Error("failed to scan courier row for nearest search", zap.Error(err))
            return nil, fmt.Errorf("failed to scan courier row: %w", err)
        }

        courier.Location, err = ewkbToCoordinates(locationBytes)
        if err != nil {
            r.logger.Error("failed to convert location from EWKB for nearest courier", zap.String("courier_id", courier.UserID.String()), zap.Error(err))
            return nil, fmt.Errorf("failed to process courier location for nearest search: %w", err)
        }

        couriers = append(couriers, courier)
    }

    if err = rows.Err(); err != nil {
        r.logger.Error("error during rows iteration for nearest search", zap.Error(err))
        return nil, fmt.Errorf("error during rows iteration: %w", err)
    }

	r.logger.Info("found nearest available couriers", zap.Int("count", len(couriers)), zap.Float64("latitude", latitude), zap.Float64("longitude", longitude), zap.Float64("radius", radius))

	return couriers, nil
}
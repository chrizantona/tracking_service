package repository

import (
	"database/sql"
	"fmt"

	"backend/internal/entity"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CourierRepository interface {
	GetByID(id uuid.UUID) (*entity.Courier, error)
	Update(c *entity.Courier) error
	FindNearestAvailable(lat, lon, radius float64) ([]*entity.Courier, error)
}

type courierRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewCourierRepository(db *sql.DB, logger *zap.Logger) CourierRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &courierRepo{db: db, logger: logger}
}

func (r *courierRepo) GetByID(id uuid.UUID) (*entity.Courier, error) {
	const op = "CourierRepository.GetByID"
	l := r.logger.With(zap.String("op", op), zap.String("courier_id", id.String()))

	const query = `
	SELECT user_id, name, status,
	       ST_X(location) AS lon,
	       ST_Y(location) AS lat,
	       rating
	  FROM couriers
	 WHERE user_id = $1
	`
	row := r.db.QueryRow(query, id)
	var c entity.Courier
	var lon, lat float64
	if err := row.Scan(&c.UserID, &c.Name, &c.Status, &lon, &lat, &c.Rating); err != nil {
		if err == sql.ErrNoRows {
			l.Warn("not found")
			return nil, fmt.Errorf("%s: courier not found", op)
		}
		l.Error("scan failed", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	c.Location = &entity.Coordinates{Latitude: lat, Longitude: lon}
	return &c, nil
}

func (r *courierRepo) Update(c *entity.Courier) error {
	const op = "CourierRepository.Update"
	l := r.logger.With(zap.String("op", op), zap.String("courier_id", c.UserID.String()))

	const query = `
	INSERT INTO couriers(user_id, name, status, location, rating)
	VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), $6)
	ON CONFLICT (user_id) DO UPDATE
	  SET name = EXCLUDED.name,
	      status = EXCLUDED.status,
	      location = EXCLUDED.location,
	      rating = EXCLUDED.rating
	`
	res, err := r.db.Exec(
		query,
		c.UserID,
		c.Name,
		c.Status,
		c.Location.Longitude, // X
		c.Location.Latitude,  // Y
		c.Rating,
	)
	if err != nil {
		l.Error("exec failed", zap.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		l.Warn("no rows affected")
	}
	return nil
}

func (r *courierRepo) FindNearestAvailable(lat, lon, radius float64) ([]*entity.Courier, error) {
	const op = "CourierRepository.FindNearestAvailable"
	l := r.logger.With(zap.String("op", op))

	const query = `
	SELECT user_id, name, status,
	       ST_X(location) AS lon,
	       ST_Y(location) AS lat,
	       rating
	  FROM couriers
	 WHERE status = $1
	   AND ST_DWithin(
	         location,
	         ST_SetSRID(ST_MakePoint($2, $3), 4326),
	         $4
	       )
	`
	rows, err := r.db.Query(query, entity.CourierStatusAvailable, lon, lat, radius)
	if err != nil {
		l.Error("query failed", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	var ret []*entity.Courier
	for rows.Next() {
		var c entity.Courier
		var lon2, lat2 float64
		if err := rows.Scan(&c.UserID, &c.Name, &c.Status, &lon2, &lat2, &c.Rating); err != nil {
			l.Error("scan failed", zap.Error(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		c.Location = &entity.Coordinates{Latitude: lat2, Longitude: lon2}
		ret = append(ret, &c)
	}
	if err := rows.Err(); err != nil {
		l.Error("rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	l.Info("found", zap.Int("count", len(ret)))
	return ret, nil
}

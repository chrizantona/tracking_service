package repository

import (
	"backend/internal/entity"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CourierRepository interface {
	GetByID(id uuid.UUID) (*entity.Courier, error)
	Update(courier *entity.Courier) error
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

func (r *courierRepo) GetByID(id uuid.UUID) (*entity.Courier, error) {
	query := `SELECT user_id, name, status, location, rating FROM couriers WHERE user_id = $1`
	courier := &entity.Courier{}
	err := r.db.QueryRow(query, id).Scan(&courier.UserID, &courier.Name, &courier.Status, &courier.Location, &courier.Rating)
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("courier not found", zap.String("id", id.String()))
			return nil, fmt.Errorf("courier not found")
		}
		r.logger.Error("failed to get courier", zap.Error(err))
		return nil, err
	}
	return courier, nil
}

func (r *courierRepo) Update(courier *entity.Courier) error {
	query := `UPDATE couriers SET name=$2, status=$3, location=$4, rating=$5 WHERE user_id=$1`
	res, err := r.db.Exec(query, courier.UserID, courier.Name, courier.Status, courier.Location, courier.Rating)
	if err != nil {
		r.logger.Error("failed to update courier", zap.Error(err))
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

package service

import (
	"fmt"

	"backend/internal/entity"
	"backend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CourierService interface {
	GetCourierByID(id uuid.UUID) (*entity.Courier, error)
	UpdateCourierStatus(id uuid.UUID, status entity.CourierStatus) error
	UpdateCourierLocation(id uuid.UUID, location *entity.Coordinates) error
	FindNearestAvailable(latitude, longitude float64, radius float64) ([]*entity.Courier, error)
}

type courierService struct {
	repo   repository.CourierRepository
	logger *zap.Logger
}

func NewCourierService(repo repository.CourierRepository, logger *zap.Logger) CourierService {
	return &courierService{
		repo:   repo,
		logger: logger,
	}
}

func (s *courierService) GetCourierByID(id uuid.UUID) (*entity.Courier, error) {
	s.logger.Info("Fetching courier", zap.String("courier_id", id.String()))
	courier, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Failed to get courier", zap.String("courier_id", id.String()), zap.Error(err))
		return nil, fmt.Errorf("failed to get courier: %w", err)
	}
	return courier, nil
}

func (s *courierService) UpdateCourierStatus(id uuid.UUID, status entity.CourierStatus) error {
	s.logger.Info("Updating courier status", zap.String("courier_id", id.String()), zap.String("status", string(status)))
	courier, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Courier not found for status update", zap.String("courier_id", id.String()), zap.Error(err))
		return fmt.Errorf("courier not found: %w", err)
	}
	courier.Status = status
	if err := s.repo.Update(courier); err != nil {
		s.logger.Error("Failed to update courier status", zap.String("courier_id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to update courier status: %w", err)
	}
	return nil
}

func (s *courierService) UpdateCourierLocation(id uuid.UUID, location *entity.Coordinates) error {
	logLat, logLon := float64(0), float64(0)
    if location != nil {
        logLat = location.Latitude
        logLon = location.Longitude
    }
	s.logger.Info("Updating courier location", zap.String("courier_id", id.String()), zap.Float64("lat", logLat), zap.Float64("lon", logLon))

	courier, err := s.repo.GetByID(id)
	if err != nil {
		s.logger.Error("Courier not found for location update", zap.String("courier_id", id.String()), zap.Error(err))
		return fmt.Errorf("courier not found: %w", err)
	}
	courier.Location = location
	if err := s.repo.Update(courier); err != nil {
		s.logger.Error("Failed to update courier location", zap.String("courier_id", id.String()), zap.Error(err))
		return fmt.Errorf("failed to update courier location: %w", err)
	}
	return nil
}

func (s *courierService) FindNearestAvailable(latitude, longitude float64, radius float64) ([]*entity.Courier, error) {
    s.logger.Info("Finding nearest available couriers", zap.Float64("lat", latitude), zap.Float64("lon", longitude), zap.Float64("radius", radius))

    couriers, err := s.repo.FindNearestAvailable(latitude, longitude, radius)
    if err != nil {
        s.logger.Error("Failed to find nearest available couriers from repository", zap.Error(err))
        return nil, fmt.Errorf("failed to find nearest available couriers: %w", err)
    }

    s.logger.Info("Successfully found nearest available couriers", zap.Int("count", len(couriers)))

    return couriers, nil
}
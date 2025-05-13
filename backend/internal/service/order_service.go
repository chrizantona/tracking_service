package service

import (
	"context"
	"fmt"

	"backend/internal/entity"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error)
	AssignCourierToOrder(ctx context.Context, orderID uuid.UUID) error
	GetOrderByID(id uuid.UUID) (*entity.Order, error)
	GetAllOrders() ([]*entity.Order, error)
	UpdateOrder(order *entity.Order) error
	DeleteOrder(id uuid.UUID) error
}

type orderService struct {
	orderRepo   repository.OrderRepository
	courierRepo repository.CourierRepository
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	courierRepo repository.CourierRepository,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		courierRepo: courierRepo,
	}
}


func (s *orderService) CreateOrder(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	if err := s.orderRepo.Create(order); err != nil {
		return nil, fmt.Errorf("create order in repository: %w", err)
	}
	return order, nil
}

func (s *orderService) GetOrderByID(id uuid.UUID) (*entity.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *orderService) GetAllOrders() ([]*entity.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *orderService) UpdateOrder(order *entity.Order) error {
	return s.orderRepo.Update(order)
}

func (s *orderService) DeleteOrder(id uuid.UUID) error {
	return s.orderRepo.Delete(id)
}



func (s *orderService) AssignCourierToOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return fmt.Errorf("get order: %w", err)
	}

	lat, lon, err := order.ParseCoords()
	if err != nil {
		return fmt.Errorf("parse delivery coords: %w", err)
	}

	couriers, err := s.courierRepo.FindNearestAvailable(lat, lon, 5_000)
	if err != nil {
		return fmt.Errorf("find nearest couriers: %w", err)
	}
	if len(couriers) == 0 {
		return fmt.Errorf("no available couriers found")
	}

	chosen := couriers[0]
	order.CourierID = &chosen.UserID
	order.Status = entity.StatusAssigned

	if err := s.orderRepo.Update(order); err != nil {
		return fmt.Errorf("assign courier to order: %w", err)
	}

	return nil
}

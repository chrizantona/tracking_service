package service

import (
	"errors"
	"fmt"

	"backend/internal/entity"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type OrderService interface {
	CreateOrder(order *entity.Order) (*entity.Order, error)
	GetOrderByID(id uuid.UUID) (*entity.Order, error)
	GetAllOrders() ([]*entity.Order, error)
	UpdateOrder(order *entity.Order) error
	DeleteOrder(id uuid.UUID) error
}

type orderService struct {
	repo repository.OrderRepository
}

func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &orderService{repo: orderRepo}
}

func (s *orderService) CreateOrder(order *entity.Order) (*entity.Order, error) {
	err := s.repo.Create(order)
	if err != nil {
		if errors.Is(err, repository.ErrClientNotFound) {
			return nil, fmt.Errorf("failed to create order: %w", err)
		}
		return nil, fmt.Errorf("failed to create order in repository: %w", err)
	}
	return order, nil
}

func (s *orderService) GetOrderByID(id uuid.UUID) (*entity.Order, error) {
	order, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return nil, fmt.Errorf("order with id %s not found: %w", id, err)
		}
		return nil, fmt.Errorf("failed to get order %s: %w", id, err)
	}
	return order, nil
}

func (s *orderService) GetAllOrders() ([]*entity.Order, error) {
	orders, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all orders: %w", err)
	}
	return orders, nil
}

func (s *orderService) UpdateOrder(order *entity.Order) error {
	err := s.repo.Update(order)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return fmt.Errorf("cannot update order %s: not found or no changes: %w", order.ID, err)
		}
		return fmt.Errorf("failed to update order %s: %w", order.ID, err)
	}
	return nil
}

func (s *orderService) DeleteOrder(id uuid.UUID) error {
	err := s.repo.Delete(id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			return fmt.Errorf("cannot delete order %s: not found: %w", id, err)
		}
		return fmt.Errorf("failed to delete order %s: %w", id, err)
	}
	return nil
}

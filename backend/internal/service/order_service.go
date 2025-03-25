package service

import (
	"backend/internal/entity"
	"backend/internal/repository"
)

type OrderService interface {
	CreateOrder(order *entity.Order) error
	GetOrderByID(id string) (*entity.Order, error)
	GetAllOrders() ([]*entity.Order, error)
	UpdateOrder(order *entity.Order) error
	DeleteOrder(id string) error
}

type orderService struct {
	repo repository.OrderRepository
	userRepo repository.UserRepository
}

func NewOrderService(orderRepo repository.OrderRepository, userRepo repository.UserRepository) OrderService {
	return &orderService{repo: orderRepo, userRepo: userRepo}
}

func (s *orderService) CreateOrder(order *entity.Order) error {
	// добавить проверку что клиент существует
	return s.repo.Create(order)
}

func (s *orderService) GetOrderByID(id string) (*entity.Order, error) {
	return s.repo.GetByID(id)
}

func (s *orderService) GetAllOrders() ([]*entity.Order, error) {
	return s.repo.GetAll()
}

func (s *orderService) UpdateOrder(order *entity.Order) error {
	return s.repo.Update(order)
}

func (s *orderService) DeleteOrder(id string) error {
	return s.repo.Delete(id)
}

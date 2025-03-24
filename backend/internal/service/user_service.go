package service

import (
	"errors"

	"backend/internal/entity"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(email, password string) (*entity.User, error)
	Login(email, password string) (*entity.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) Register(email, password string) (*entity.User, error) {
	_, err := s.repo.GetByEmail(email)
	if err == nil {
		return nil, errors.New("user already exists")
	}
	if err != nil && err.Error() != "user not found" {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: string(hashed),
		Role:         entity.RoleClient, 
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) Login(email, password string) (*entity.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}
	return user, nil
}

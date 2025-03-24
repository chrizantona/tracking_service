package entity

import (
	"time"
)

type Role string

const (
	RoleClient  Role = "CLIENT"
	RoleCourier Role = "COURIER"
	RoleAdmin   Role = "ADMIN"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` 
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

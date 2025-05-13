package config

import (
	"errors"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	JWTSecret   string
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	jwt := os.Getenv("JWT_SECRET")
	if jwt == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}
	return &Config{
		ServerPort:  port,
		DatabaseURL: dbURL,
		JWTSecret:   jwt,
	}, nil
}

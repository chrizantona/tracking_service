package config

import (
	"errors"
	"os"
)

type Config struct {
	ServerPort string 
	DBUrl      string 
	JWTSecret  string 
}

func LoadConfig() (*Config, error) {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" 
	}
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		return nil, errors.New("DATABASE_URL не установлен")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET не установлен")
	}
	return &Config{
		ServerPort: port,
		DBUrl:      dbUrl,
		JWTSecret:  jwtSecret,
	}, nil
}

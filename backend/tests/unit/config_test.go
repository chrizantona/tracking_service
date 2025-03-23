package config_test

import (
	"os"
	"testing"

	"backend/config"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("SERVER_PORT", "9000")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
	os.Setenv("JWT_SECRET", "secret")
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("ожидалась успешная загрузка конфигурации, получена ошибка: %v", err)
	}
	if cfg.ServerPort != "9000" {
		t.Errorf("ожидался порт 9000, получен %s", cfg.ServerPort)
	}
	if cfg.DBUrl == "" {
		t.Error("DATABASE_URL должен быть установлен")
	}
	if cfg.JWTSecret == "" {
		t.Error("JWT_SECRET должен быть установлен")
	}
}

package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"backend/internal/controller"
	"backend/internal/entity"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeUserService struct{}

func (f *fakeUserService) Register(email, password string) (*entity.User, error) {
	if email == "exists@example.com" {
		return nil, errors.New("user already exists")
	}
	return &entity.User{
		ID:    "123",
		Email: email,
		Role:  entity.RoleClient,
	}, nil
}

func (f *fakeUserService) Login(email, password string) (*entity.User, error) {
	if email == "test@example.com" && password == "password123" {
		return &entity.User{
			ID:    "123",
			Email: email,
			Role:  entity.RoleClient,
		}, nil
	}
	return nil, errors.New("invalid email or password")
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	fakeSvc := &fakeUserService{}
	uc := controller.NewUserController(fakeSvc, "testsecret")
	router.POST("/register", uc.Register)
	router.POST("/login", uc.Login)
	return router
}

func TestRegister_ValidData(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]string{
		"email":    "newuser@example.com",
		"password": "securepass",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "newuser@example.com", resp["email"])
	assert.Equal(t, "CLIENT", resp["role"])
	assert.NotEmpty(t, resp["id"])
}

func TestRegister_DuplicateEmail(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]string{
		"email":    "exists@example.com",
		"password": "securepass",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "user already exists")
}

func TestLogin_ValidData(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp["token"])
}

func TestLogin_InvalidPassword(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid email or password")
}

func TestLogin_NonExistentUser(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "password",
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid email or password")
}

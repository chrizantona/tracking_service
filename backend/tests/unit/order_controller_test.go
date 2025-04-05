package config_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/controller"
	"backend/internal/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type fakeOrderService struct {
	orders          map[uuid.UUID]*entity.Order
	failClientCheck bool
}

func newFakeOrderService() *fakeOrderService {
	return &fakeOrderService{
		orders: make(map[uuid.UUID]*entity.Order),
	}
}

func (f *fakeOrderService) CreateOrder(order *entity.Order) (*entity.Order, error) {
	if f.failClientCheck {
		return nil, errors.New("client does not exist")
	}
	order.ID = uuid.New()
	now := time.Now()
	order.CreatedAt = now
	order.UpdatedAt = now
	f.orders[order.ID] = order
	return order, nil
}

func (f *fakeOrderService) GetOrderByID(id uuid.UUID) (*entity.Order, error) {
	order, exists := f.orders[id]
	if !exists {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (f *fakeOrderService) GetAllOrders() ([]*entity.Order, error) {
	var orders []*entity.Order
	for _, order := range f.orders {
		orders = append(orders, order)
	}
	return orders, nil
}

func (f *fakeOrderService) UpdateOrder(order *entity.Order) error {
	_, exists := f.orders[order.ID]
	if !exists {
		return errors.New("order not found")
	}
	order.UpdatedAt = time.Now()
	f.orders[order.ID] = order
	return nil
}

func (f *fakeOrderService) DeleteOrder(id uuid.UUID) error {
	_, exists := f.orders[id]
	if !exists {
		return errors.New("order not found")
	}
	delete(f.orders, id)
	return nil
}

func setupOrderRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	fakeSvc := newFakeOrderService()
	oc := controller.NewOrderController(fakeSvc)
	router.POST("/orders", oc.CreateOrder)
	router.GET("/orders", oc.GetOrders)
	router.GET("/orders/:id", oc.GetOrder)
	router.PUT("/orders/:id", oc.UpdateOrder)
	router.DELETE("/orders/:id", oc.DeleteOrder)
	return router
}

func TestCreateOrder_ValidData(t *testing.T) {
	router := setupOrderRouter()
	clientID := uuid.New().String()
	reqBody := map[string]string{
		"client_id":        clientID,
		"delivery_address": "123 Main St",
		"delivery_coords":  "37.7749,-122.4194",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var order entity.Order
	err := json.Unmarshal(w.Body.Bytes(), &order)
	assert.NoError(t, err)
	assert.Equal(t, clientID, order.ClientID.String())
	assert.Equal(t, entity.StatusCreated, order.Status)
}

func TestCreateOrder_ClientNotFound(t *testing.T) {
	fakeSvc := newFakeOrderService()
	fakeSvc.failClientCheck = true
	gin.SetMode(gin.TestMode)
	router := gin.New()
	oc := controller.NewOrderController(fakeSvc)
	router.POST("/orders", oc.CreateOrder)
	reqBody := map[string]string{
		"client_id":        uuid.New().String(),
		"delivery_address": "123 Main St",
		"delivery_coords":  "37.7749,-122.4194",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetOrder_NotFound(t *testing.T) {
	router := setupOrderRouter()
	req, _ := http.NewRequest("GET", "/orders/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateOrder(t *testing.T) {
	router := setupOrderRouter()
	clientID := uuid.New().String()
	reqBody := map[string]string{
		"client_id":        clientID,
		"delivery_address": "123 Main St",
		"delivery_coords":  "37.7749,-122.4194",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var order entity.Order
	err := json.Unmarshal(w.Body.Bytes(), &order)
	assert.NoError(t, err)
	updateReq := map[string]interface{}{
		"status":           string(entity.StatusAssigned),
		"delivery_address": "456 Elm St",
		"delivery_coords":  "40.7128,-74.0060",
	}
	updateBody, _ := json.Marshal(updateReq)
	reqUpdate, _ := http.NewRequest("PUT", "/orders/"+order.ID.String(), bytes.NewBuffer(updateBody))
	reqUpdate.Header.Set("Content-Type", "application/json")
	wUpdate := httptest.NewRecorder()
	router.ServeHTTP(wUpdate, reqUpdate)
	assert.Equal(t, http.StatusOK, wUpdate.Code)
	var updatedOrder entity.Order
	err = json.Unmarshal(wUpdate.Body.Bytes(), &updatedOrder)
	assert.NoError(t, err)
	assert.Equal(t, entity.StatusAssigned, updatedOrder.Status)
	assert.Equal(t, "456 Elm St", updatedOrder.DeliveryAddress)
}

func TestDeleteOrder(t *testing.T) {
	router := setupOrderRouter()
	clientID := uuid.New().String()
	reqBody := map[string]string{
		"client_id":        clientID,
		"delivery_address": "123 Main St",
		"delivery_coords":  "37.7749,-122.4194",
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var order entity.Order
	err := json.Unmarshal(w.Body.Bytes(), &order)
	assert.NoError(t, err)
	reqDelete, _ := http.NewRequest("DELETE", "/orders/"+order.ID.String(), nil)
	wDelete := httptest.NewRecorder()
	router.ServeHTTP(wDelete, reqDelete)
	assert.Equal(t, http.StatusOK, wDelete.Code)
	reqGet, _ := http.NewRequest("GET", "/orders/"+order.ID.String(), nil)
	wGet := httptest.NewRecorder()
	router.ServeHTTP(wGet, reqGet)
	assert.Equal(t, http.StatusNotFound, wGet.Code)
}

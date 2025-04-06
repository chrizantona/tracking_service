package config_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"backend/internal/controller"
	"backend/internal/entity"
	"backend/tests/mocks"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func setupMockCourierRouter(ctrl *gomock.Controller) (*gin.Engine, *mocks.MockCourierService) {
	router := gin.New()
	mockSvc := mocks.NewMockCourierService(ctrl)
	cc := controller.NewCourierController(mockSvc)
	router.GET("/couriers/:id", cc.GetCourier)
	router.PUT("/couriers/:id/status", cc.UpdateStatus)
	router.PUT("/couriers/:id/location", cc.UpdateLocation)
	return router, mockSvc
}

func TestGetCourier_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router, mockSvc := setupMockCourierRouter(ctrl)

	courierID := uuid.New()
	expectedCourier := &entity.Courier{
		UserID:   courierID,
		Name:     "Test Courier",
		Status:   entity.CourierAvailable,
		Location: "0,0",
		Rating:   4.5,
	}

	mockSvc.EXPECT().
		GetCourierByID(courierID).
		Return(expectedCourier, nil)

	req, _ := http.NewRequest("GET", "/couriers/"+courierID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var courier entity.Courier
	err := json.Unmarshal(w.Body.Bytes(), &courier)
	assert.NoError(t, err)
	assert.Equal(t, expectedCourier.UserID, courier.UserID)
	assert.Equal(t, expectedCourier.Name, courier.Name)
}

func TestUpdateCourierStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router, mockSvc := setupMockCourierRouter(ctrl)
	courierID := uuid.New()
	mockSvc.EXPECT().
		UpdateCourierStatus(courierID, entity.CourierBusy).
		Return(nil)

	reqBody := map[string]string{
		"status": string(entity.CourierBusy),
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/couriers/"+courierID.String()+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateCourierLocation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	router, mockSvc := setupMockCourierRouter(ctrl)
	courierID := uuid.New()
	newLocation := "10.1234,20.5678"
	mockSvc.EXPECT().
		UpdateCourierLocation(courierID, entity.Geometry(newLocation)).
		Return(nil)

	reqBody := map[string]string{
		"location": newLocation,
	}
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/couriers/"+courierID.String()+"/location", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

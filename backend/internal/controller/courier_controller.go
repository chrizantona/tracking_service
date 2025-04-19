package controller

import (
	"net/http"
	"backend/internal/entity"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourierController struct {
	service service.CourierService
}

func NewCourierController(s service.CourierService) *CourierController {
	return &CourierController{
		service: s,
	}
}

func (cc *CourierController) GetCourier(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
		return
	}
	courier, err := cc.service.GetCourierByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, courier)
}

type UpdateCourierStatusRequest struct {
	Status entity.CourierStatus `json:"status" binding:"required"`
}

func (cc *CourierController) UpdateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
		return
	}
	var req UpdateCourierStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := cc.service.UpdateCourierStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "status updated"})
}

type UpdateCourierLocationRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
}

func (cc *CourierController) UpdateLocation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
		return
	}
	var req UpdateCourierLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	location := &entity.Coordinates{
        Latitude: req.Latitude,
        Longitude: req.Longitude,
    }

	if err := cc.service.UpdateCourierLocation(id, location); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "location updated"})
}

type FindNearestRequest struct {
	Latitude  float64 `form:"latitude" binding:"required"`
	Longitude float64 `form:"longitude" binding:"required"`
	Radius    float64 `form:"radius" binding:"required,gt=0"`
}

func (cc *CourierController) FindNearestCouriers(c *gin.Context) {
	var req FindNearestRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	couriers, err := cc.service.FindNearestAvailable(req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, couriers)
}
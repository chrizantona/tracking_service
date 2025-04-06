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
	Location entity.Geometry `json:"location" binding:"required"`
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
	if err := cc.service.UpdateCourierLocation(id, req.Location); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "location updated"})
}

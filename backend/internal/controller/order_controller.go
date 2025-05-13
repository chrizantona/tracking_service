package controller

import (
	"net/http"

	"backend/internal/entity"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OrderController struct {
	orderService service.OrderService
}

func NewOrderController(orderService service.OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

type CreateOrderRequest struct {
	ClientID        uuid.UUID `json:"client_id" binding:"required"`
	DeliveryAddress string    `json:"delivery_address" binding:"required"`
	DeliveryCoords  string    `json:"delivery_coords" binding:"required"`
}

func (oc *OrderController) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := &entity.Order{
		ID:              uuid.New(),
		ClientID:        req.ClientID,
		Status:          entity.StatusCreated,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryCoords:  req.DeliveryCoords,
	}

	created, err := oc.orderService.CreateOrder(c.Request.Context(), order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, created)
}

func (oc *OrderController) GetOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	order, err := oc.orderService.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (oc *OrderController) GetOrders(c *gin.Context) {
	list, err := oc.orderService.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

type UpdateOrderRequest struct {
	Status          entity.OrderStatus `json:"status" binding:"required"`
	DeliveryAddress string             `json:"delivery_address" binding:"required"`
	DeliveryCoords  string             `json:"delivery_coords" binding:"required"`
}

func (oc *OrderController) UpdateOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := oc.orderService.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	order.Status = req.Status
	order.DeliveryAddress = req.DeliveryAddress
	order.DeliveryCoords = req.DeliveryCoords

	if err := oc.orderService.UpdateOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (oc *OrderController) DeleteOrder(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	if err := oc.orderService.DeleteOrder(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order deleted"})
}

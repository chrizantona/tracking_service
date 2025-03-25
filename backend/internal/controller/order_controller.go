package controller

import (
	"net/http"

	"backend/internal/entity"
	"backend/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService service.OrderService
}

func NewOrderController(orderService service.OrderService) *OrderController {
	return &OrderController{orderService: orderService}
}

type CreateOrderRequest struct {
	ClientID        string `json:"client_id" binding:"required"`
	DeliveryAddress string `json:"delivery_address" binding:"required"`
	DeliveryCoords  string `json:"delivery_coords" binding:"required"`
}

func (oc *OrderController) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := &entity.Order{
		ClientID:        req.ClientID,
		Status:          entity.StatusCreated,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryCoords:  req.DeliveryCoords,
	}

	if err := oc.orderService.CreateOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (oc *OrderController) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := oc.orderService.GetOrderByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (oc *OrderController) GetOrders(c *gin.Context) {
	orders, err := oc.orderService.GetAllOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

type UpdateOrderRequest struct {
	Status          entity.OrderStatus `json:"status" binding:"required"`
	DeliveryAddress string             `json:"delivery_address" binding:"required"`
	DeliveryCoords  string             `json:"delivery_coords" binding:"required"`
	// подумать над тем чтобы айдишник курьера передавать
}

func (oc *OrderController) UpdateOrder(c *gin.Context) {
	id := c.Param("id")
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
	id := c.Param("id")
	if err := oc.orderService.DeleteOrder(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order deleted"})
}

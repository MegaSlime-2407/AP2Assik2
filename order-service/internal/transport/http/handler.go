package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"order-service/internal/domain"
	"order-service/internal/usecase"
)

type OrderHandler struct {
	uc *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{uc: uc}
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	orders := r.Group("/orders")
	{
		orders.GET("", h.getAllOrders)
		orders.POST("", h.createOrder)
		orders.GET("/:id", h.getOrder)
		orders.DELETE("/:id", h.deleteOrder)
		orders.PATCH("/:id/cancel", h.cancelOrder)
		orders.GET("/:id/task", h.task)
	}
}

type createOrderRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
	ItemName   string `json:"item_name" binding:"required"`
	Amount     int64  `json:"amount" binding:"required"`
}

type orderResponse struct {
	ID         string `json:"id"`
	CustomerID string `json:"customer_id"`
	ItemName   string `json:"item_name"`
	Amount     int64  `json:"amount"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func toOrderResponse(o *domain.Order) orderResponse {
	return orderResponse{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		ItemName:   o.ItemName,
		Amount:     o.Amount,
		Status:     o.Status,
		CreatedAt:  o.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *OrderHandler) createOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id, item_name and amount are required"})
		return
	}

	order, err := h.uc.CreateOrder(c.Request.Context(), usecase.CreateOrderInput{
		CustomerID: req.CustomerID,
		ItemName:   req.ItemName,
		Amount:     req.Amount,
	})
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, domain.ErrPaymentFailed) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment service unavailable"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":     order.ID,
		"status": order.Status,
	})
}

func (h *OrderHandler) getAllOrders(c *gin.Context) {
	orders, err := h.uc.GetAllOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	result := make([]orderResponse, 0, len(orders))
	for _, o := range orders {
		result = append(result, toOrderResponse(o))
	}
	c.JSON(http.StatusOK, result)
}

func (h *OrderHandler) getOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.GetOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) cancelOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.uc.CancelOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		if errors.Is(err, domain.ErrCancelNotAllowed) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) deleteOrder(c *gin.Context) {
	id := c.Param("id")
	err := h.uc.DeleteOrder(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "order deleted"})
}

func (h *OrderHandler) task(c *gin.Context) {
	customerID := c.Param("id")
	result, err := h.uc.Task(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"customer_id":  customerID,
		"total_amount": result.TotalAmount,
		"total_orders": result.TotalOrders,
	})
}

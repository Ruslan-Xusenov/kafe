package handlers

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"kafe/internal/models"
)

type BulkEditRequest struct {
	AddItems    []models.OrderItem `json:"add_items"`
	CancelItems []struct {
		ProductID int     `json:"product_id"`
		Quantity  float64 `json:"quantity"`
	} `json:"cancel_items"`
}

func (h *OrderHandler) BulkEditOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заказа"})
		return
	}

	var req BulkEditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	u := user.(*models.User)

	err = h.service.BulkEditOrder(orderID, req.AddItems, req.CancelItems, u.ID, u.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order bulk updated successfully"})
}

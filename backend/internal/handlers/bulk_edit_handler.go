package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
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

	userID, exists := c.Get("user_id")
	role, roleExists := c.Get("role")
	if !exists || !roleExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uid := userID.(int)
	userRole := role.(string)

	err = h.service.BulkEditOrder(orderID, req.AddItems, req.CancelItems, uid, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "order.bulk_edit", "order", &orderID, gin.H{
		"add_items_count":    len(req.AddItems),
		"cancel_items_count": len(req.CancelItems),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Order bulk updated successfully"})
}

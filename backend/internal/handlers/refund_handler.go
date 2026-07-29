package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type RefundHandler struct {
	refundRepo *repository.RefundRepository
	auditRepo  *repository.AuditRepository
}

func NewRefundHandler(refundRepo *repository.RefundRepository, auditRepo *repository.AuditRepository) *RefundHandler {
	return &RefundHandler{refundRepo: refundRepo, auditRepo: auditRepo}
}

// CreateRefund creates a new refund request
func (h *RefundHandler) CreateRefund(c *gin.Context) {
	var req struct {
		OrderID      int     `json:"order_id" binding:"required"`
		Amount       float64 `json:"amount" binding:"required"`
		Reason       string  `json:"reason" binding:"required"`
		ReasonDetail string  `json:"reason_detail"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	// Validate reason
	if _, ok := models.ValidRefundReasons[req.Reason]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri sabab kodi"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Summa musbat bo'lishi kerak"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int)

	// Get user name for display
	userName := "Unknown"
	if uname, exists := c.Get("user_name"); exists {
		userName = uname.(string)
	}

	refund := &models.Refund{
		OrderID:         req.OrderID,
		Amount:          req.Amount,
		Reason:          req.Reason,
		ReasonDetail:    req.ReasonDetail,
		RequestedBy:     &uid,
		RequestedByName: userName,
	}

	if err := h.refundRepo.CreateRefund(refund); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "refund.create", "refund", &refund.ID, gin.H{
		"order_id": req.OrderID,
		"amount":   req.Amount,
		"reason":   req.Reason,
	})

	c.JSON(http.StatusCreated, refund)
}

// GetPendingRefunds returns all pending refund requests
func (h *RefundHandler) GetPendingRefunds(c *gin.Context) {
	refunds, err := h.refundRepo.GetPendingRefunds()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, refunds)
}

// GetAllRefunds returns refunds with optional filtering
func (h *RefundHandler) GetAllRefunds(c *gin.Context) {
	status := c.Query("status")
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 100
	}

	refunds, err := h.refundRepo.GetAllRefunds(status, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, refunds)
}

// ApproveRefund approves a pending refund (admin only)
func (h *RefundHandler) ApproveRefund(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	var req struct {
		RefundMethod string `json:"refund_method"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.RefundMethod != "cash" && req.RefundMethod != "card" && req.RefundMethod != "click" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri refund usuli"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int)
	userName := "Admin"
	if uname, exists := c.Get("user_name"); exists {
		userName = uname.(string)
	}

	if err := h.refundRepo.ApproveRefund(id, uid, userName, req.RefundMethod); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "refund.approve", "refund", &id, gin.H{
		"refund_method": req.RefundMethod,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Refund tasdiqlandi"})
}

// RejectRefund rejects a pending refund (admin only)
func (h *RefundHandler) RejectRefund(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int)
	userName := "Admin"
	if uname, exists := c.Get("user_name"); exists {
		userName = uname.(string)
	}

	if err := h.refundRepo.RejectRefund(id, uid, userName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "refund.reject", "refund", &id, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Refund rad etildi"})
}

// MarkMoneyReturned confirms that money was physically returned
func (h *RefundHandler) MarkMoneyReturned(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	if err := h.refundRepo.MarkMoneyReturned(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "refund.money_returned", "refund", &id, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Pul qaytarildi deb belgilandi"})
}

// GetRefundsByOrder returns refunds for a specific order
func (h *RefundHandler) GetRefundsByOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri order ID"})
		return
	}

	refunds, err := h.refundRepo.GetRefundsByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, refunds)
}

// GetRefundReasons returns the list of valid refund reasons
func (h *RefundHandler) GetRefundReasons(c *gin.Context) {
	c.JSON(http.StatusOK, models.ValidRefundReasons)
}

// CountPending returns count of pending refunds (for badge)
func (h *RefundHandler) CountPending(c *gin.Context) {
	count, err := h.refundRepo.CountPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"count": count})
}

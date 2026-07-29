package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type CashierHandler struct {
	cashierRepo  *repository.CashierRepository
	orderService *service.OrderService
	auditRepo    *repository.AuditRepository
}

func NewCashierHandler(cashierRepo *repository.CashierRepository, orderService *service.OrderService, auditRepo *repository.AuditRepository) *CashierHandler {
	return &CashierHandler{
		cashierRepo:  cashierRepo,
		orderService: orderService,
		auditRepo:    auditRepo,
	}
}

// OpenShift starts a new cashier shift
func (h *CashierHandler) OpenShift(c *gin.Context) {
	var req struct {
		OpeningCash float64 `json:"opening_cash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	userID, _ := c.Get("user_id")
	shift, err := h.cashierRepo.OpenShift(userID.(int), req.OpeningCash)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "cashier.shift.open", "cashier_shift", &shift.ID, gin.H{
		"opening_cash": req.OpeningCash,
	})

	c.JSON(http.StatusCreated, shift)
}

// CloseShift closes the current shift with cash reconciliation
func (h *CashierHandler) CloseShift(c *gin.Context) {
	var req struct {
		ShiftID     int     `json:"shift_id" binding:"required"`
		ClosingCash float64 `json:"closing_cash"`
		Notes       string  `json:"notes"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	shift, statusCode, err := h.authorizedShift(c, req.ShiftID)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	if shift.Status != "open" {
		c.JSON(http.StatusConflict, gin.H{"error": "Smena allaqachon yopilgan"})
		return
	}

	shift, err = h.cashierRepo.CloseShift(req.ShiftID, req.ClosingCash, req.Notes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "cashier.shift.close", "cashier_shift", &req.ShiftID, gin.H{
		"closing_cash":  req.ClosingCash,
		"expected_cash": shift.ExpectedCash,
	})

	c.JSON(http.StatusOK, shift)
}

// GetCurrentShift returns the active shift for the logged-in cashier
func (h *CashierHandler) GetCurrentShift(c *gin.Context) {
	userID, _ := c.Get("user_id")
	shift, err := h.cashierRepo.GetActiveShift(userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Return null if no active shift (frontend will show "Open Shift" button)
	c.JSON(http.StatusOK, shift)
}

// AddCashOperation handles cash-in/cash-out during a shift
func (h *CashierHandler) AddCashOperation(c *gin.Context) {
	var req struct {
		ShiftID int     `json:"shift_id" binding:"required"`
		Type    string  `json:"type" binding:"required"`
		Amount  float64 `json:"amount" binding:"required"`
		Reason  string  `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	if req.Type != "cash_in" && req.Type != "cash_out" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type faqat cash_in yoki cash_out bo'lishi mumkin"})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "summa musbat bo'lishi kerak"})
		return
	}
	if shift, statusCode, err := h.authorizedShift(c, req.ShiftID); err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	} else if shift.Status != "open" {
		c.JSON(http.StatusConflict, gin.H{"error": "Yopilgan smenaga kassa operatsiyasi yozib bo'lmaydi"})
		return
	}

	userID, _ := c.Get("user_id")
	op := &models.CashOperation{
		ShiftID:   req.ShiftID,
		Type:      req.Type,
		Amount:    req.Amount,
		Reason:    req.Reason,
		CreatedBy: userID.(int),
	}

	if err := h.cashierRepo.AddCashOperation(op); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, fmt.Sprintf("cashier.%s", req.Type), "cash_operation", &op.ID, gin.H{
		"shift_id": req.ShiftID,
		"amount":   req.Amount,
		"reason":   req.Reason,
	})

	c.JSON(http.StatusCreated, op)
}

// GetShiftReport returns X-report (during shift) or Z-report (after closing)
func (h *CashierHandler) GetShiftReport(c *gin.Context) {
	shiftID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri shift ID"})
		return
	}
	if _, statusCode, err := h.authorizedShift(c, shiftID); err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	report, err := h.cashierRepo.GetShiftReport(shiftID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// authorizedShift prevents a cashier from operating on another cashier's shift.
// Admins may inspect or manage any shift.
func (h *CashierHandler) authorizedShift(c *gin.Context, shiftID int) (*models.CashierShift, int, error) {
	shift, err := h.cashierRepo.GetShiftByID(shiftID)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	if shift == nil {
		return nil, http.StatusNotFound, fmt.Errorf("Smena topilmadi")
	}

	if c.GetString("role") == string(models.RoleAdmin) {
		return shift, http.StatusOK, nil
	}

	userID, exists := c.Get("user_id")
	if !exists {
		return nil, http.StatusUnauthorized, fmt.Errorf("Avtorizatsiya talab qilinadi")
	}
	id, ok := userID.(int)
	if !ok || shift.CashierID != id {
		return nil, http.StatusForbidden, fmt.Errorf("Bu smenani boshqarish huquqi yo'q")
	}
	return shift, http.StatusOK, nil
}

// QuickSale creates a fast order without table (direct POS sale)
func (h *CashierHandler) QuickSale(c *gin.Context) {
	var req struct {
		Items    []models.OrderItem    `json:"items" binding:"required"`
		Payments []models.PaymentInput `json:"payments" binding:"required"`
		ShiftID  int                   `json:"shift_id" binding:"required"`
		Comment  string                `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Mahsulotlar ro'yxati bo'sh"})
		return
	}
	if len(req.Payments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "To'lov ma'lumotlari kerak"})
		return
	}

	userID, _ := c.Get("user_id")
	id := userID.(int)
	shift, err := h.cashierRepo.GetShiftByID(req.ShiftID)
	if err != nil || shift == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Smena topilmadi"})
		return
	}
	if shift.Status != "open" || (shift.CashierID != id && c.GetString("role") != "admin") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu smenada savdo qilish huquqi yo'q"})
		return
	}

	// Build order
	order := &models.Order{
		WaiterID: &id, // cashier acts as waiter for POS orders
		Items:    req.Items,
		Address:  "POS Savdo",
		Phone:    "POS",
		Comment:  req.Comment,
	}

	// Create order through service (handles stock, price calc, notifications)
	if err := h.orderService.CreatePaidOrder(order, req.Payments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Close order immediately (POS = instant delivery)
	if err := h.orderService.UpdateOrderStatus(order.ID, models.StatusDelivered, id, "cashier"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Buyurtmani yopishda xatolik: " + err.Error()})
		return
	}

	// Record sale in shift
	var cashAmt, cardAmt, clickAmt, nasiyaAmt float64
	for _, p := range req.Payments {
		switch p.Method {
		case "cash":
			cashAmt += p.Amount
		case "card":
			cardAmt += p.Amount
		case "click":
			clickAmt += p.Amount
		case "nasiya":
			nasiyaAmt += p.Amount
		}
	}
	if err := h.cashierRepo.RecordSale(req.ShiftID, order.TotalPrice, cashAmt, cardAmt, clickAmt, nasiyaAmt); err != nil {
		// Non-fatal: order is already created and closed. Log the error but don't fail the response.
		fmt.Printf("⚠️  [CASHIER] RecordSale failed for shift %d, order %d: %v\n", req.ShiftID, order.ID, err)
	}

	writeAudit(c, h.auditRepo, "cashier.quick_sale", "order", &order.ID, gin.H{
		"total":    order.TotalPrice,
		"payments": req.Payments,
		"shift_id": req.ShiftID,
	})

	c.JSON(http.StatusCreated, gin.H{
		"order":   order,
		"message": "Savdo muvaffaqiyatli yakunlandi",
	})
}

// GetAllShifts returns recent shifts for admin
func (h *CashierHandler) GetAllShifts(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	shifts, err := h.cashierRepo.GetAllShifts(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, shifts)
}

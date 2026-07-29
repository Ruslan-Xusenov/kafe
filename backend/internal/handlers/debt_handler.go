package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type DebtHandler struct {
	debtRepo  *repository.DebtRepository
	auditRepo *repository.AuditRepository
}

func NewDebtHandler(debtRepo *repository.DebtRepository, auditRepo *repository.AuditRepository) *DebtHandler {
	return &DebtHandler{debtRepo: debtRepo, auditRepo: auditRepo}
}

// GetAllDebtors returns all debtors
func (h *DebtHandler) GetAllDebtors(c *gin.Context) {
	onlyWithDebt := c.DefaultQuery("active", "false") == "true"
	debtors, err := h.debtRepo.GetAllDebtors(onlyWithDebt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, debtors)
}

// CreateDebtor creates a new debtor
func (h *DebtHandler) CreateDebtor(c *gin.Context) {
	var debtor models.Debtor
	if err := c.ShouldBindJSON(&debtor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	if debtor.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ism kiritilishi shart"})
		return
	}

	if err := h.debtRepo.CreateDebtor(&debtor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "debtor.create", "debtor", &debtor.ID, gin.H{
		"name":  debtor.Name,
		"phone": debtor.Phone,
	})

	c.JSON(http.StatusCreated, debtor)
}

// UpdateDebtor updates debtor info
func (h *DebtHandler) UpdateDebtor(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	var debtor models.Debtor
	if err := c.ShouldBindJSON(&debtor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}
	debtor.ID = id

	if err := h.debtRepo.UpdateDebtor(&debtor); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, debtor)
}

// GetDebtorHistory returns debt/payment history for a debtor
func (h *DebtHandler) GetDebtorHistory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	debtor, err := h.debtRepo.GetDebtorByID(id)
	if err != nil || debtor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Qarzdor topilmadi"})
		return
	}

	history, err := h.debtRepo.GetDebtHistory(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"debtor":  debtor,
		"history": history,
	})
}

// PayDebt records a payment against a debtor's balance
func (h *DebtHandler) PayDebt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ID"})
		return
	}

	var req struct {
		Amount      float64 `json:"amount" binding:"required"`
		Method      string  `json:"method"`
		Description string  `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Summa musbat bo'lishi kerak"})
		return
	}

	debtor, err := h.debtRepo.GetDebtorByID(id)
	if err != nil || debtor == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Qarzdor topilmadi"})
		return
	}

	if req.Amount > debtor.TotalDebt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "To'lov summasi qarzdan ko'p"})
		return
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int)

	record := &models.DebtRecord{
		DebtorID:      id,
		Amount:        req.Amount,
		Type:          "payment",
		PaymentMethod: req.Method,
		Description:   req.Description,
		CreatedBy:     &uid,
	}

	if err := h.debtRepo.AddDebtRecord(record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "debtor.payment", "debtor", &id, gin.H{
		"amount": req.Amount,
		"method": req.Method,
	})

	// Return updated debtor
	updatedDebtor, _ := h.debtRepo.GetDebtorByID(id)
	c.JSON(http.StatusOK, gin.H{
		"message": "To'lov qabul qilindi",
		"debtor":  updatedDebtor,
		"record":  record,
	})
}

// AddDebtRecord creates a new debt entry (used when closing order with nasiya)
func (h *DebtHandler) AddDebtRecord(c *gin.Context) {
	var record models.DebtRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri ma'lumot"})
		return
	}

	if record.DebtorID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Qarzdor tanlanmagan"})
		return
	}
	if record.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Summa musbat bo'lishi kerak"})
		return
	}
	if record.Type == "" {
		record.Type = "debt"
	}

	userID, _ := c.Get("user_id")
	uid := userID.(int)
	record.CreatedBy = &uid

	if err := h.debtRepo.AddDebtRecord(&record); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "debtor.add_debt", "debtor", &record.DebtorID, gin.H{
		"amount":   record.Amount,
		"order_id": record.OrderID,
	})

	c.JSON(http.StatusCreated, record)
}

// GetDebtSummary returns aggregate debt stats
func (h *DebtHandler) GetDebtSummary(c *gin.Context) {
	summary, err := h.debtRepo.GetDebtSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

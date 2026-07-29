package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type FiscalHandler struct {
	fiscalService *service.FiscalService
	auditRepo     *repository.AuditRepository
}

func NewFiscalHandler(fiscalService *service.FiscalService, auditRepo *repository.AuditRepository) *FiscalHandler {
	return &FiscalHandler{fiscalService: fiscalService, auditRepo: auditRepo}
}

// GetReceiptByOrder returns the fiscal receipt for a given order
func (h *FiscalHandler) GetReceiptByOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri buyurtma ID"})
		return
	}

	receipt, err := h.fiscalService.GetReceiptByOrderID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fiskal chek topilmadi"})
		return
	}

	c.JSON(http.StatusOK, receipt)
}

// GetAllReceipts returns all fiscal receipts with optional filtering
func (h *FiscalHandler) GetAllReceipts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	status := c.Query("status")

	receipts, err := h.fiscalService.GetAllReceipts(limit, offset, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if receipts == nil {
		receipts = []models.FiscalReceipt{}
	}

	c.JSON(http.StatusOK, receipts)
}

// GetSettings returns fiscal configuration
func (h *FiscalHandler) GetSettings(c *gin.Context) {
	settings, err := h.fiscalService.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// UpdateSettings updates fiscal configuration
func (h *FiscalHandler) UpdateSettings(c *gin.Context) {
	var settings models.FiscalSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.fiscalService.UpdateSettings(&settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "fiscal.settings_update", "fiscal", nil, gin.H{
		"enabled":      settings.Enabled,
		"vat_rate":     settings.VATRate,
		"inn":          settings.INN,
		"company_name": settings.CompanyName,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Fiskal sozlamalar saqlandi"})
}

// ResendToOFD attempts to resend a receipt to OFD
func (h *FiscalHandler) ResendToOFD(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("order_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Noto'g'ri buyurtma ID"})
		return
	}

	receipt, err := h.fiscalService.GetReceiptByOrderID(orderID)
	if err != nil || receipt == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Fiskal chek topilmadi"})
		return
	}

	if err := h.fiscalService.ResendToOFD(receipt.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "fiscal.resend_ofd", "fiscal_receipt", &receipt.ID, gin.H{
		"receipt_number": receipt.ReceiptNumber,
	})

	c.JSON(http.StatusOK, gin.H{"message": "OFD ga yuborish so'rovi qabul qilindi"})
}

// GetStats returns fiscal statistics
func (h *FiscalHandler) GetStats(c *gin.Context) {
	stats, err := h.fiscalService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

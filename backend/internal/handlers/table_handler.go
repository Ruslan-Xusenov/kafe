package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type TableHandler struct {
	tableRepo    *repository.TableRepository
	orderService *service.OrderService
	auditRepo    *repository.AuditRepository
}

func NewTableHandler(tableRepo *repository.TableRepository, orderService *service.OrderService, auditRepo *repository.AuditRepository) *TableHandler {
	return &TableHandler{tableRepo: tableRepo, orderService: orderService, auditRepo: auditRepo}
}

func (h *TableHandler) GetAll(c *gin.Context) {
	tables, err := h.tableRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var filtered []models.Table
	for _, t := range tables {
		if t.Status == "occupied" && role == "waiter" {
			if t.ActiveWaiterID != nil && *t.ActiveWaiterID != userID.(int) {
				continue // Hide occupied tables belonging to other waiters
			}
		}
		filtered = append(filtered, t)
	}

	c.JSON(http.StatusOK, filtered)
}

func (h *TableHandler) Create(c *gin.Context) {
	var table models.Table
	if err := c.ShouldBindJSON(&table); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if table.Status == "" {
		table.Status = "free"
	}

	if err := h.tableRepo.Create(&table); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "table.create", "table", &table.ID, gin.H{
		"name":     table.Name,
		"capacity": table.Capacity,
		"status":   table.Status,
	})

	c.JSON(http.StatusCreated, table)
}

func (h *TableHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	var payload struct {
		models.Table
		PaymentMethod string              `json:"payment_method"`
		Payments      []models.PaymentInput `json:"payments"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	table := payload.Table
	table.ID = id

	// Pre-validation: If table is being freed, check waiter ownership
	if table.Status == "free" && h.orderService != nil {
		activeOrders, err := h.orderService.GetActiveOrders()
		if err == nil {
			for _, o := range activeOrders {
				if o.TableID != nil && *o.TableID == table.ID {
					userID, _ := c.Get("user_id")
					role, _ := c.Get("role")
					if userID == nil {
						userID = 0
					}
					if role == nil {
						role = "admin"
					}

					// Security: Only the waiter who owns the order can free the table
					if role == "waiter" && o.WaiterID != nil && *o.WaiterID != userID.(int) {
						c.JSON(http.StatusForbidden, gin.H{"error": "Вы можете освобождать только свои столы."})
						return
					}
				}
			}
		}
	}

	if table.Status == "free" && h.orderService != nil {
		userID, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if userID == nil {
			userID = 0
		}
		if role == nil {
			role = "admin"
		}

		// Use mixed payments if provided, otherwise fall back to single method
		if len(payload.Payments) > 0 {
			if err := h.orderService.CloseTableWithPayments(table.ID, payload.Payments, userID.(int), role.(string)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			writeAudit(c, h.auditRepo, "table.close", "table", &table.ID, gin.H{
				"payments": payload.Payments,
			})
		} else {
			if err := h.orderService.CloseTable(table.ID, payload.PaymentMethod, userID.(int), role.(string)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			writeAudit(c, h.auditRepo, "table.close", "table", &table.ID, gin.H{
				"payment_method": payload.PaymentMethod,
			})
		}
	}

	if err := h.tableRepo.Update(&table); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "table.update", "table", &table.ID, gin.H{
		"name":     table.Name,
		"capacity": table.Capacity,
		"status":   table.Status,
	})

	c.JSON(http.StatusOK, table)
}

func (h *TableHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}

	if err := h.tableRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "table.delete", "table", &id, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}

// UpdateLayout batch-updates table positions from the floor plan editor
func (h *TableHandler) UpdateLayout(c *gin.Context) {
	var payload struct {
		Tables []models.Table `json:"tables"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tableRepo.UpdateLayout(payload.Tables); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "table.layout_update", "table", nil, gin.H{
		"count": len(payload.Tables),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Zal xaritasi saqlandi", "count": len(payload.Tables)})
}

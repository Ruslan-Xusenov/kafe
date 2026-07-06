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
}

func NewTableHandler(tableRepo *repository.TableRepository, orderService *service.OrderService) *TableHandler {
	return &TableHandler{tableRepo: tableRepo, orderService: orderService}
}

func (h *TableHandler) GetAll(c *gin.Context) {
	tables, err := h.tableRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tables)
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
		PaymentMethod string `json:"payment_method"`
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

	if err := h.tableRepo.Update(&table); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Post-update: Close the order now that we confirmed permission
	if table.Status == "free" && h.orderService != nil {
		activeOrders, _ := h.orderService.GetActiveOrders()
		for _, o := range activeOrders {
			if o.TableID != nil && *o.TableID == table.ID {
				userID, _ := c.Get("user_id")
				role, _ := c.Get("role")
				if userID == nil { userID = 0 }
				if role == nil { role = "admin" }
				
				// Set Payment Method if provided
				if payload.PaymentMethod != "" {
					_ = h.orderService.SetPaymentMethod(o.ID, payload.PaymentMethod)
				}
				
				_ = h.orderService.UpdateOrderStatus(o.ID, models.StatusDelivered, userID.(int), role.(string))
			}
		}
	}

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

	c.JSON(http.StatusOK, gin.H{"message": "Удалено"})
}

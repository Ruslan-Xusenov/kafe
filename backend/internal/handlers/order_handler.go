package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

func (h *OrderHandler) Service() *service.OrderService {
	return h.service
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")
	id := userID.(int)

	if role == string(models.RoleWaiter) {
		order.WaiterID = &id
	} else {
		order.CustomerID = &id
	}

	if err := h.service.CreateOrder(&order); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	order, err := h.service.GetOrderByID(id, userID.(int), role.(string))
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Заказ не найден"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetMyOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orders, err := h.service.GetCustomerOrders(userID.(int))
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders()
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetActiveOrders(c *gin.Context) {
	orders, err := h.service.GetActiveOrders()
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("API_DEBUG: Sending %d active orders to kitchen UI\n", len(orders))
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	if err := h.service.UpdateOrderStatus(id, req.Status, userID.(int), role.(string)); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Статус обновлен"})
}

func (h *OrderHandler) SubmitRating(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ratings []models.StaffRating
	if err := c.ShouldBindJSON(&ratings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.SubmitRating(id, ratings); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Оценки отправлены"})
}

func (h *OrderHandler) GetStaffPerformance(c *gin.Context) {
	performance, err := h.service.GetStaffPerformance()
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, performance)
}

func (h *OrderHandler) AssignCourier(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	courierID, _ := c.Get("user_id") // Use the logged in courier's ID

	if err := h.service.AssignCourier(id, courierID.(int)); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Курьер назначен"})
}
func (h *OrderHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	stats, err := h.service.GetStats(userID.(int), role.(string))
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (h *OrderHandler) GetOrderRatings(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	ratings, err := h.service.GetRatingsByOrderID(id)
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ratings)
}

func (h *OrderHandler) TestPrinter(c *gin.Context) {
	if err := h.service.TestPrinter(); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Тест отправлен"})
}

func (h *OrderHandler) ReprintOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.ReprintOrder(id); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Заказ отправлен на печать"})
}

func (h *OrderHandler) SetServiceFee(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Percentage float64 `json:"percentage"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.service.SetServiceFee(id, req.Percentage)
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetWaiterHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	orders, err := h.service.GetOrderHistoryByWaiter(userID.(int))
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetActiveOrdersByWaiter(c *gin.Context) {
	waiterID, err := strconv.Atoi(c.Param("waiterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID официанта"})
		return
	}
	orders, err := h.service.GetActiveOrdersByWaiter(waiterID)
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) GetOrderHistoryByWaiter(c *gin.Context) {
	waiterID, err := strconv.Atoi(c.Param("waiterID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID официанта"})
		return
	}
	orders, err := h.service.GetOrderHistoryByWaiter(waiterID)
	if err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

func (h *OrderHandler) CancelOrderItem(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заказа"})
		return
	}
	itemID, err := strconv.Atoi(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID продукта"})
		return
	}

	var req struct {
		Quantity float64 `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.service.CancelOrderItem(orderID, itemID, req.Quantity); err != nil {
		fmt.Printf("CREATE_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Продукт отменен из заказа"})
}

func (h *OrderHandler) GetActiveOrderByTable(c *gin.Context) {
	tableID, err := strconv.Atoi(c.Param("tableID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID стола"})
		return
	}
	order, err := h.service.GetActiveOrderByTable(tableID)
	if err != nil {
		fmt.Printf("GET_ACTIVE_ORDER_BY_TABLE_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) AddItemsToOrder(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID заказа"})
		return
	}

	var req struct {
		Items []models.OrderItem `json:"items"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Список продуктов пуст"})
		return
	}

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	updatedOrder, err := h.service.AddItemsToExistingOrder(orderID, req.Items, userID.(int), role.(string))
	if err != nil {
		fmt.Printf("ADD_ITEMS_TO_ORDER_ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedOrder)
}
package service

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	settingsRepo *repository.SettingsRepository
	inventoryRepo repository.InventoryRepository
	wsService    *WebsocketService
	botService   *BotService
	printerService *PrinterService
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, settingsRepo *repository.SettingsRepository, inventoryRepo repository.InventoryRepository, wsService *WebsocketService, botService *BotService, printerService *PrinterService) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		settingsRepo: settingsRepo,
		inventoryRepo: inventoryRepo,
		wsService:    wsService,
		botService:   botService,
		printerService: printerService,
	}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	// 1. Get container price and ID from settings
	containerPrice := 1000.0
	if cp, err := s.settingsRepo.Get("container_price"); err == nil && cp != "" {
		if f, err := strconv.ParseFloat(cp, 64); err == nil {
			containerPrice = f
		}
	}

	containerID := 7 
	if cid, err := s.settingsRepo.Get("container_product_id"); err == nil && cid != "" {
		if id, err := strconv.Atoi(cid); err == nil {
			containerID = id
		}
	}

	// Check for existing order from the same phone in the last 30 minutes (skip for cafe orders)
	if os.Getenv("APP_ENV") != "development" && order.TableID == nil {
		lastOrder, err := s.orderRepo.GetLastOrderByPhone(order.Phone)
		if err == nil && lastOrder != nil {
			if time.Since(lastOrder.CreatedAt) < 30*time.Minute {
				return fmt.Errorf("вы делали заказ в последние 30 минут. Пожалуйста, подождите.")
			}
		}
	}

	// Business logic for mandatory containers (only for delivery/takeaway, skip for dine-in tables)
	var itemsToAdd []models.OrderItem
	if order.TableID == nil {
	for _, item := range order.Items {
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil || prod == nil { continue }

		if prod.HasMandatoryContainer {
			totalPortions := 0.0
			if item.Unit == "gr" {
				totalPortions = item.Quantity / 100.0
			} else if item.Unit == "kg" {
				totalPortions = item.Quantity * 10.0
			} else {
				// For 'pors' and 'dona'
				totalPortions = item.Quantity
			}

			if totalPortions > 0 {
				numContainers := math.Ceil(totalPortions) 
				
				itemsToAdd = append(itemsToAdd, models.OrderItem{
					ProductID: containerID, // Use dynamic ID
					Quantity:  numContainers,
					Price:     containerPrice,
				})
			}
		}
	}

	// Add containers to the order items if not already present
	for _, newItem := range itemsToAdd {
		found := false
		for i, existingItem := range order.Items {
			if existingItem.ProductID == newItem.ProductID {
				order.Items[i].Quantity += newItem.Quantity
				found = true
				break
			}
		}
		if !found {
			order.Items = append(order.Items, newItem)
		}
	}
	} // end if order.TableID == nil

	var total float64
	for i := range order.Items {
		item := &order.Items[i]
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil || prod == nil {
			if item.ProductID == containerID {
				return fmt.Errorf("продукт 'контейнер' в настройках системы (ID: %d) не найден. Пожалуйста, проверьте настройки в админ панели.", containerID)
			}
			return fmt.Errorf("продукт не найден (ID: %d)", item.ProductID)
		}

		itemPrice := prod.Price
		if item.Unit == "dona" && prod.Unit == "pors" {
			itemPrice = prod.Price / 4.0
		}
		
		item.Price = itemPrice
		item.ProductName = prod.Name
		total += item.Price * item.Quantity
	}
	
	// Check minimum order value (40,000 UZS) for delivery customers
	if os.Getenv("APP_ENV") != "development" && order.TableID == nil {
		if total < 40000 {
			return fmt.Errorf("минимальная сумма заказа должна быть 40.000 сум")
		}
	}

	order.TotalPrice = total
	order.Status = models.StatusNew
	order.CreatedAt = time.Now()

	if err := s.orderRepo.Create(order); err != nil {
		return err
	}

	// Fetch fully populated order to get WaiterName and TableNumber
	populatedOrder, err := s.orderRepo.GetByID(order.ID)
	if err == nil && populatedOrder != nil {
		order = populatedOrder
	}

	// Enrich order items with product names for notification
	// And deduct inventory stock for each item
	for i := range order.Items {
		prod, _ := s.productRepo.GetByID(order.Items[i].ProductID)
		if prod != nil {
			order.Items[i].ProductName = prod.Name
		}
		
		// Deduct inventory stock
		_ = s.inventoryRepo.DeductStockForProduct(order.Items[i].ProductID, order.Items[i].Quantity)
	}

	// Trigger Task: Send Notification to Telegram
	var firstImageUrl string
	if len(order.Items) > 0 {
		prod, err := s.productRepo.GetByID(order.Items[0].ProductID)
		if err == nil && prod != nil {
			firstImageUrl = prod.ImageURL
		}
	}
	s.botService.SendNewOrderNotification(order, &firstImageUrl)
	
	// Real-time: Notify Cooks and Admin (Printer is notified separately via notifyAPI to avoid double printing for bot orders)
	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": order})
	s.wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": order})

	// Print Cafe orders (since they don't trigger the bot's notifyAPI)
	if order.TableID != nil {
		s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})
		go s.printerService.PrintOrder(order)
	}

	return nil
}

func (s *OrderService) GetOrderByID(id int, userID int, role string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, nil
	}

	// Permisson check: Admin/Staff or the Customer who placed the order
	isOwner := false
	if order.CustomerID != nil && *order.CustomerID == userID {
		isOwner = true
	}
	if role != "admin" && role != "cook" && role != "courier" && !isOwner {
		return nil, fmt.Errorf("unauthorized access to order")
	}

	return order, nil
}

func (s *OrderService) GetOrderWithItems(id int) (*models.Order, error) {
	// The repo GetByID already fetches items with product names
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order %d not found", id)
	}

	return order, nil
}

func (s *OrderService) GetOrderItems(orderID int) ([]models.OrderItem, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order %d not found", orderID)
	}
	return order.Items, nil
}

func (s *OrderService) GetCustomerOrders(customerID int) ([]models.Order, error) {
	return s.orderRepo.GetByCustomerID(customerID)
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *OrderService) GetActiveOrders() ([]models.Order, error) {
	var activeOrders []models.Order
	all, err := s.orderRepo.GetAll()
	if err != nil {
		return nil, err
	}

	for _, o := range all {
		if o.Status != models.StatusDelivered && o.Status != models.StatusCancelled {
			activeOrders = append(activeOrders, o)
		}
	}
	return activeOrders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID int, status models.OrderStatus, userID int, role string) error {
	order, _ := s.orderRepo.GetByID(orderID)
	if order != nil && role == "waiter" {
		if order.WaiterID != nil && *order.WaiterID != userID {
			return fmt.Errorf("Siz faqat o'zingizning buyurtmalaringizni o'zgartira olasiz")
		}
	}

	var cookID *int
	if role == "cook" && status == models.StatusPreparing {
		cookID = &userID
	}

	err := s.orderRepo.UpdateStatus(orderID, status, cookID)
	if err == nil {
		order, _ := s.orderRepo.GetByID(orderID)
		if order != nil {
			// If cancelled, restore stock
			if status == models.StatusCancelled {
				for _, item := range order.Items {
					_ = s.inventoryRepo.RestoreStockForProduct(item.ProductID, item.Quantity)
				}
			}

			// Notify Customer
			if order.CustomerID != nil {
				s.wsService.BroadcastToUser(*order.CustomerID, map[string]interface{}{"type": "status_update", "status": status, "order_id": orderID})
			}

			// Notify Roles
			updatePayload := map[string]interface{}{"type": "status_update", "status": status, "order_id": orderID}
			s.wsService.BroadcastToRole("admin", updatePayload)
			s.wsService.BroadcastToRole("cook", updatePayload)
			s.wsService.BroadcastToRole("courier", updatePayload)

			// Notify bot for specific statuses
			if status == models.StatusReady {
				s.botService.SendOrderStatusNotification(order, "ГОТОВО - Курьеры могут забрать")
			} else if status == models.StatusDelivered {
				s.botService.SendOrderStatusNotification(order, "ДОСТАВЛЕНО")
				
				// Print Final Receipt for Cafe orders when closed
				if order.TableID != nil {
					s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "close_order", "order": order})
					go s.printerService.PrintOrder(order)
				}
			}
		}
	}
	return err
}

func (s *OrderService) AssignCourier(orderID int, courierID int) error {
	err := s.orderRepo.AssignCourier(orderID, courierID)
	if err == nil {
		order, _ := s.orderRepo.GetByID(orderID)
		if order != nil {
			if order.CustomerID != nil {
				s.wsService.BroadcastToUser(*order.CustomerID, map[string]interface{}{"type": "status_update", "status": models.StatusOnWay, "order_id": orderID})
			}

			// Notify Roles
			updatePayload := map[string]interface{}{"type": "status_update", "status": models.StatusOnWay, "order_id": orderID}
			s.wsService.BroadcastToRole("admin", updatePayload)
			s.wsService.BroadcastToRole("cook", updatePayload)
			s.wsService.BroadcastToRole("courier", updatePayload)
		}
	}
	return err
}
func (s *OrderService) GetStats(userID int, role string) (*models.DeliveryStats, error) {
	if role == "admin" {
		return s.orderRepo.GetAdminStats()
	} else if role == "cook" {
		return s.orderRepo.GetKitchenStats()
	} else if role == "courier" {
		return s.orderRepo.GetCourierStats(userID)
	}
	return nil, fmt.Errorf("role %s not authorized for stats", role)
}

func (s *OrderService) SubmitRating(orderID int, ratings []models.StaffRating) error {
	for i := range ratings {
		ratings[i].OrderID = orderID
		if err := s.orderRepo.AddStaffRating(&ratings[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderService) GetStaffPerformance() ([]models.StaffPerformance, error) {
	return s.orderRepo.GetStaffPerformance()
}

func (s *OrderService) GetRatingsByOrderID(orderID int) ([]models.StaffRating, error) {
	return s.orderRepo.GetRatingsByOrderID(orderID)
}

func (s *OrderService) TestPrinter() error {
	testOrder := &models.Order{
		ID:         9999,
		TotalPrice: 50000,
		Address:    "ТЕСТОВЫЙ АДРЕС",
		Phone:      "998901234567",
		Items: []models.OrderItem{
			{ProductName: "ТЕСТ БЛЮДО 1", Quantity: 1, Price: 25000},
			{ProductName: "ТЕСТ БЛЮДО 2", Quantity: 1, Price: 25000},
		},
	}
	
	// Broadcast to roles
	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": testOrder})
	s.wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": testOrder})
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": testOrder})

	// Direct Print
	go s.printerService.PrintOrder(testOrder)

	return nil
}

func (s *OrderService) ReprintOrder(orderID int) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}
	
	// Notify WS bridge just in case they use it instead of TCP
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})
	
	// Direct TCP print
	go s.printerService.PrintOrder(order)
	return nil
}

func (s *OrderService) SetServiceFee(orderID int, percentage float64) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order %d not found", orderID)
	}

	// Calculate base price (total minus any previous service fee)
	basePrice := order.TotalPrice - order.ServiceFee

	// Calculate new service fee
	serviceFee := basePrice * (percentage / 100.0)
	newTotal := basePrice + serviceFee

	// Save to DB
	err = s.orderRepo.SetServiceFee(orderID, percentage, serviceFee, newTotal)
	if err != nil {
		return nil, err
	}

	// Refresh order data
	order, _ = s.orderRepo.GetByID(orderID)
	return order, nil
}

func (s *OrderService) SetPaymentMethod(orderID int, method string) error {
	validMethods := map[string]bool{"cash": true, "card": true, "click": true, "nasiya": true}
	if !validMethods[method] {
		return fmt.Errorf("недопустимый тип оплаты: %s", method)
	}
	return s.orderRepo.SetPaymentMethod(orderID, method)
}
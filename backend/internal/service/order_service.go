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

	// Always print orders created via API (both cafe and delivery)
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})
	go s.printerService.PrintOrder(order)

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
				// NOTE: For cafe/table orders, printing is handled by CloseTable() to produce ONE combined receipt.
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
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "reprint_order", "order": order})
	
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

	// Only allow service fee for internal orders (those with a table)
	if order.TableID == nil {
		return nil, fmt.Errorf("плата за обслуживание применяется только к внутренним заказам")
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

func (s *OrderService) GetWaiterHistory() ([]models.Order, error) {
	return s.orderRepo.GetWaiterHistory()
}

func (s *OrderService) GetActiveOrdersByWaiter(waiterID int) ([]models.Order, error) {
	return s.orderRepo.GetActiveByWaiterID(waiterID)
}

func (s *OrderService) GetOrderHistoryByWaiter(waiterID int) ([]models.Order, error) {
	return s.orderRepo.GetHistoryByWaiterID(waiterID)
}

func (s *OrderService) CancelOrderItem(orderID, itemID int, cancelQty float64) error {
	// First fetch the order and item to know what we are cancelling for the printer
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("order not found")
	}

	var cancelledItem *models.OrderItem
	for _, it := range order.Items {
		if it.ID == itemID {
			// Copy the item so we don't modify the original array by reference just in case
			itemCopy := it
			cancelledItem = &itemCopy
			break
		}
	}
	if cancelledItem == nil {
		return fmt.Errorf("item not found in order")
	}
	
	// Validate quantity
	if cancelQty <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if cancelQty > cancelledItem.Quantity {
		cancelQty = cancelledItem.Quantity
	}

	// Update DB
	if err := s.orderRepo.CancelItem(orderID, itemID, cancelQty); err != nil {
		return err
	}

	// Broadcast cancellation to printer with the specific cancelled quantity
	cancelledItem.Quantity = cancelQty
	cancelPayload := map[string]interface{}{
		"type": "cancel_item",
		"order_id": orderID,
		"item": cancelledItem,
		"waiter_name": order.WaiterName,
		"table_number": order.TableNumber,
	}
	s.wsService.BroadcastToRole("printer", cancelPayload)

	return nil
}

// CloseTable marks all active orders for a table as delivered and prints ONE combined receipt.
func (s *OrderService) CloseTable(tableID int, paymentMethod string, userID int, role string) error {
	activeOrders, err := s.orderRepo.GetAll()
	if err != nil {
		return err
	}

	// Filter orders belonging to this table that are still active
	var tableOrders []*models.Order
	for i := range activeOrders {
		o := &activeOrders[i]
		if o.TableID != nil && *o.TableID == tableID &&
			o.Status != models.StatusDelivered && o.Status != models.StatusCancelled {
			tableOrders = append(tableOrders, o)
		}
	}

	if len(tableOrders) == 0 {
		return nil // Nothing to close
	}

	// Mark all orders as delivered
	for _, o := range tableOrders {
		if paymentMethod != "" {
			_ = s.orderRepo.SetPaymentMethod(o.ID, paymentMethod)
		}
		_ = s.orderRepo.UpdateStatus(o.ID, models.StatusDelivered, nil)
	}

	// Build ONE combined receipt from the first order's metadata + all items merged
	firstOrder := tableOrders[0]

	// Re-fetch first order to get full waiter/table info
	populated, err := s.orderRepo.GetByID(firstOrder.ID)
	if err != nil || populated == nil {
		populated = firstOrder
	}

	// Merge all items and sum totals
	var allItems []models.OrderItem
	var grandTotal float64
	var grandServiceFee float64
	var servicePercentage float64

	for _, o := range tableOrders {
		full, err := s.orderRepo.GetByID(o.ID)
		if err != nil || full == nil {
			continue
		}
		allItems = append(allItems, full.Items...)
		grandTotal += full.TotalPrice
		grandServiceFee += full.ServiceFee
		if full.ServicePercentage > 0 {
			servicePercentage = full.ServicePercentage
		}
	}

	// Create combined order struct for printing
	combinedOrder := &models.Order{
		ID:                populated.ID,
		TableID:           populated.TableID,
		TableNumber:       populated.TableNumber,
		WaiterID:         populated.WaiterID,
		WaiterName:       populated.WaiterName,
		TotalPrice:       grandTotal,
		ServiceFee:       grandServiceFee,
		ServicePercentage: servicePercentage,
		Items:             allItems,
		CreatedAt:        populated.CreatedAt,
		Status:           models.StatusDelivered,
	}

	// Broadcast ONE combined receipt to printer
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "close_order", "order": combinedOrder})
	go s.printerService.PrintOrder(combinedOrder)

	return nil
}

// GetActiveOrderByTable returns the active order for a table (if any)
func (s *OrderService) GetActiveOrderByTable(tableID int) (*models.Order, error) {
	return s.orderRepo.FindActiveOrderByTableID(tableID)
}

// AddItemsToExistingOrder appends new items to an existing active order
func (s *OrderService) AddItemsToExistingOrder(orderID int, items []models.OrderItem, userID int, role string) (*models.Order, error) {
	// 1. Get existing order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil || order == nil {
		return nil, fmt.Errorf("buyurtma topilmadi (ID: %d)", orderID)
	}

	// 2. Check order is still active
	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return nil, fmt.Errorf("bu buyurtma allaqachon yopilgan")
	}

	// 3. Check waiter ownership (Admin can bypass)
	if role != "admin" {
		if order.WaiterID != nil && *order.WaiterID != userID {
			return nil, fmt.Errorf("вы можете добавлять товары только в свои заказы")
		}
	}

	// 4. Validate and set prices from DB
	for i := range items {
		item := &items[i]
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil || prod == nil {
			return nil, fmt.Errorf("mahsulot topilmadi (ID: %d)", item.ProductID)
		}

		itemPrice := prod.Price
		if item.Unit == "dona" && prod.Unit == "pors" {
			itemPrice = prod.Price / 4.0
		}

		item.Price = itemPrice
		item.ProductName = prod.Name
	}

	// 5. Add items to DB and recalculate totals
	if err := s.orderRepo.AddItemsToOrder(orderID, items); err != nil {
		return nil, err
	}

	// 6. Deduct inventory for new items
	for _, item := range items {
		_ = s.inventoryRepo.DeductStockForProduct(item.ProductID, item.Quantity)
	}

	// 7. Fetch updated order
	updatedOrder, err := s.orderRepo.GetByID(orderID)
	if err != nil || updatedOrder == nil {
		return nil, fmt.Errorf("yangilangan buyurtmani olishda xatolik")
	}

	// 8. Enrich new items with product names for notification
	for i := range items {
		prod, _ := s.productRepo.GetByID(items[i].ProductID)
		if prod != nil {
			items[i].ProductName = prod.Name
		}
	}

	// 9. Build a partial order with ONLY new items for printing to kitchen
	partialOrder := &models.Order{
		ID:          updatedOrder.ID,
		TableID:     updatedOrder.TableID,
		TableNumber: updatedOrder.TableNumber,
		WaiterID:    updatedOrder.WaiterID,
		WaiterName:  updatedOrder.WaiterName,
		TotalPrice:  updatedOrder.TotalPrice,
		Items:       items,
		CreatedAt:   updatedOrder.CreatedAt,
		Status:      updatedOrder.Status,
	}

	// 10. Notify kitchen and admin about new items
	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": partialOrder})
	s.wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": partialOrder})

	// 11. Print ONLY the new items (not the entire order)
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "is_dop": true, "order": partialOrder})
	go s.printerService.PrintOrder(partialOrder)

	// WORKAROUND: The local printer bridge (running at the cafe) has a 30s deduplication filter 
	// based on the order ID. Since 'dop' items share the same order ID, rapid consecutive additions 
	// get blocked. We send a dummy order with a fake ID to reset the bridge's lastOrderID state.
	dummyOrder := map[string]interface{}{
		"id": -9999,
		"items": []interface{}{},
	}
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": dummyOrder})

	return updatedOrder, nil
}
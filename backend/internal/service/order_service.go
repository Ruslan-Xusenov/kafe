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
	orderRepo      *repository.OrderRepository
	productRepo    *repository.ProductRepository
	settingsRepo   *repository.SettingsRepository
	inventoryRepo  repository.InventoryRepository
	tableRepo      *repository.TableRepository
	wsService      *WebsocketService
	botService     *BotService
	printerService *PrinterService
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, settingsRepo *repository.SettingsRepository, inventoryRepo repository.InventoryRepository, tableRepo *repository.TableRepository, wsService *WebsocketService, botService *BotService, printerService *PrinterService) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		productRepo:    productRepo,
		settingsRepo:   settingsRepo,
		inventoryRepo:  inventoryRepo,
		tableRepo:      tableRepo,
		wsService:      wsService,
		botService:     botService,
		printerService: printerService,
	}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	return s.createOrder(order, nil)
}

// CreatePaidOrder validates the total and persists the order with its
// payments atomically. It is intended for cashier/POS sales.
func (s *OrderService) CreatePaidOrder(order *models.Order, payments []models.PaymentInput) error {
	return s.createOrder(order, payments)
}

func (s *OrderService) createOrder(order *models.Order, payments []models.PaymentInput) error {
	// 0. Idempotency check: if a key is provided and an order already exists, return it.
	if order.IdempotencyKey != nil && *order.IdempotencyKey != "" {
		existing, err := s.orderRepo.FindByIdempotencyKey(*order.IdempotencyKey)
		if err != nil {
			return fmt.Errorf("idempotency tekshiruvida xatolik: %w", err)
		}
		if existing != nil {
			// Order already exists for this key — return the existing order silently.
			*order = *existing
			return nil
		}
	}
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
	if os.Getenv("APP_ENV") != "development" && order.TableID == nil && order.Phone != "POS" {
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
			if err != nil || prod == nil {
				continue
			}

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

		// Block orders for inactive products
		if !prod.IsActive && item.ProductID != containerID {
			return fmt.Errorf("mahsulot faol emas va buyurtma qilib bo'lmaydi: %s (ID: %d)", prod.Name, item.ProductID)
		}

		// Validate quantity against product rules
		if prod.MinQuantity > 0 && item.Quantity < prod.MinQuantity {
			return fmt.Errorf("%s: minimal miqdor %.2f, siz %.2f kiritdingiz", prod.Name, prod.MinQuantity, item.Quantity)
		}
		if prod.QuantityStep > 0 {
			// Check that quantity is a multiple of step (within floating point tolerance)
			remainder := math.Mod(item.Quantity-prod.MinQuantity, prod.QuantityStep)
			if remainder > 0.001 && (prod.QuantityStep-remainder) > 0.001 {
				return fmt.Errorf("%s: miqdor %.2f qadam bo'yicha kiritilishi kerak", prod.Name, prod.QuantityStep)
			}
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
	if os.Getenv("APP_ENV") != "development" && order.TableID == nil && order.Phone != "POS" {
		if total < 40000 {
			return fmt.Errorf("минимальная сумма заказа должна быть 40.000 сум")
		}
	}

	order.TotalPrice = total
	order.Status = models.StatusNew
	order.CreatedAt = time.Now()

	if len(payments) > 0 {
		validMethods := map[string]bool{"cash": true, "card": true, "click": true, "nasiya": true}
		var paymentTotal float64
		for _, payment := range payments {
			if !validMethods[payment.Method] {
				return fmt.Errorf("noto'g'ri to'lov usuli: %s", payment.Method)
			}
			if payment.Amount <= 0 {
				return fmt.Errorf("to'lov summasi musbat bo'lishi kerak")
			}
			paymentTotal += payment.Amount
		}
		if math.Abs(paymentTotal-order.TotalPrice) > 0.01 {
			return fmt.Errorf("to'lov summasi order summasiga teng bo'lishi kerak: %.2f / %.2f", paymentTotal, order.TotalPrice)
		}
	}

	var createErr error
	if len(payments) > 0 {
		createErr = s.orderRepo.CreateWithPayments(order, payments)
	} else {
		createErr = s.orderRepo.Create(order)
	}
	if createErr != nil {
		return createErr
	}

	// Fetch fully populated order to get WaiterName and TableNumber
	populatedOrder, err := s.orderRepo.GetByID(order.ID)
	if err == nil && populatedOrder != nil {
		order = populatedOrder
	}

	// Enrich order items with product names for notification.
	// Stock deduction is handled atomically inside orderRepo.Create.
	for i := range order.Items {
		prod, _ := s.productRepo.GetByID(order.Items[i].ProductID)
		if prod != nil {
			order.Items[i].ProductName = prod.Name
		}

		target, err := s.productRepo.GetPrinterTarget(order.Items[i].ProductID)
		if err == nil {
			order.Items[i].PrinterTarget = target
		} else {
			order.Items[i].PrinterTarget = "ALL"
		}
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
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %d not found", orderID)
	}

	if order.Status == status {
		return nil
	}
	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return fmt.Errorf("yopilgan buyurtma statusini o'zgartirib bo'lmaydi")
	}

	allowed := false
	switch role {
	case "admin":
		allowed = true
	case "cook":
		allowed = (order.Status == models.StatusNew && status == models.StatusPreparing) ||
			(order.Status == models.StatusPreparing && status == models.StatusReady)
	case "courier":
		allowed = (order.Status == models.StatusReady || order.Status == models.StatusOnWay) && status == models.StatusDelivered
	case "cashier":
		allowed = order.Status == models.StatusNew && status == models.StatusDelivered
	}
	if !allowed {
		return fmt.Errorf("role %s uchun %s statusiga o'tish mumkin emas", role, status)
	}

	if role == "waiter" {
		if order.WaiterID != nil && *order.WaiterID != userID {
			return fmt.Errorf("Siz faqat o'zingizning buyurtmalaringizni o'zgartira olasiz")
		}
	}

	var cookID *int
	if role == "cook" && status == models.StatusPreparing {
		cookID = &userID
	}

	err = s.orderRepo.UpdateStatus(orderID, status, cookID)
	if err == nil {
		order, _ := s.orderRepo.GetByID(orderID)
		if order != nil {
			// If cancelled, restore stock (log errors instead of silently ignoring)
			// MarkStockRestored is atomic — prevents double restoration if called twice.
			if status == models.StatusCancelled {
				should, markErr := s.orderRepo.MarkStockRestored(orderID)
				if markErr != nil {
					fmt.Printf("⚠️  [STOCK] Order #%d: MarkStockRestored failed: %v\n", orderID, markErr)
				} else if should {
					for _, item := range order.Items {
						if restoreErr := s.inventoryRepo.RestoreStockForProduct(item.ProductID, item.Quantity); restoreErr != nil {
							fmt.Printf("⚠️  [STOCK] Order #%d cancel: RestoreStock failed for product %d: %v\n", orderID, item.ProductID, restoreErr)
						}
					}
				} else {
					fmt.Printf("ℹ️  [STOCK] Order #%d stock already restored, skipping.\n", orderID)
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

func (s *OrderService) SubmitRating(orderID int, userID int, role string, ratings []models.StaffRating) error {
	if len(ratings) == 0 {
		return fmt.Errorf("ratinglar ro'yxati bo'sh")
	}

	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return fmt.Errorf("order %d not found", orderID)
	}
	if order.Status != models.StatusDelivered {
		return fmt.Errorf("faqat yetkazilgan buyurtmani baholash mumkin")
	}

	isOwner := order.CustomerID != nil && *order.CustomerID == userID
	if role != string(models.RoleAdmin) && !isOwner {
		return fmt.Errorf("bu buyurtmaga rating berish huquqi yo'q")
	}

	for i := range ratings {
		rating := &ratings[i]
		if rating.Rating < 1 || rating.Rating > 5 {
			return fmt.Errorf("rating 1 dan 5 gacha bo'lishi kerak")
		}
		if len(rating.Comment) > 1000 {
			return fmt.Errorf("izoh 1000 belgidan oshmasligi kerak")
		}
		if role != string(models.RoleAdmin) && !ratingMatchesOrder(order, rating) {
			return fmt.Errorf("rating faqat buyurtmaga biriktirilgan xodim uchun berilishi mumkin")
		}
		ratings[i].OrderID = orderID
		if err := s.orderRepo.AddStaffRating(rating); err != nil {
			return err
		}
	}
	return nil
}

func (s *OrderService) GetStaffPerformance() ([]models.StaffPerformance, error) {
	return s.orderRepo.GetStaffPerformance()
}

func (s *OrderService) GetRatingsByOrderID(orderID int, userID int, role string) ([]models.StaffRating, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, fmt.Errorf("order %d not found", orderID)
	}

	isOwner := order.CustomerID != nil && *order.CustomerID == userID
	if role != string(models.RoleAdmin) && !isOwner && !staffAssignedToOrder(order, userID, role) {
		return nil, fmt.Errorf("bu buyurtma ratinglarini ko'rish huquqi yo'q")
	}
	return s.orderRepo.GetRatingsByOrderID(orderID)
}

func ratingMatchesOrder(order *models.Order, rating *models.StaffRating) bool {
	switch rating.StaffRole {
	case "cook":
		return order.CookID != nil && *order.CookID == rating.StaffID
	case "courier":
		return order.CourierID != nil && *order.CourierID == rating.StaffID
	default:
		return false
	}
}

func staffAssignedToOrder(order *models.Order, userID int, role string) bool {
	switch role {
	case string(models.RoleCook):
		return order.CookID != nil && *order.CookID == userID
	case string(models.RoleCourier):
		return order.CourierID != nil && *order.CourierID == userID
	case string(models.RoleWaiter):
		return order.WaiterID != nil && *order.WaiterID == userID
	default:
		return false
	}
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
	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return fmt.Errorf("yopilgan buyurtmadan mahsulotni bekor qilib bo'lmaydi")
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

	// Validate
	if cancelQty <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if cancelQty > cancelledItem.Quantity {
		cancelQty = cancelledItem.Quantity
	}

	if err := s.orderRepo.CancelItem(orderID, itemID, cancelQty); err != nil {
		return err
	}

	cancelledItem.Quantity = cancelQty
	cancelPayload := map[string]interface{}{
		"type":        "cancel_item",
		"order_id":    orderID,
		"item":        cancelledItem,
		"waiter_name": order.WaiterName,
		"table_name":  order.TableName,
	}
	s.wsService.BroadcastToRole("printer", cancelPayload)

	return nil
}

func (s *OrderService) CloseTable(tableID int, paymentMethod string, userID int, role string) error {
	// Backward compat: convert single method to payments array
	if paymentMethod != "" {
		return s.CloseTableWithPayments(tableID, []models.PaymentInput{{Method: paymentMethod, Amount: 0}}, userID, role)
	}
	return s.CloseTableWithPayments(tableID, nil, userID, role)
}

func (s *OrderService) CloseTableWithPayments(tableID int, payments []models.PaymentInput, userID int, role string) error {
	// Validate payment methods
	validMethods := map[string]bool{"cash": true, "card": true, "click": true, "nasiya": true}
	for _, p := range payments {
		if !validMethods[p.Method] {
			return fmt.Errorf("недопустимый тип оплаты: %s", p.Method)
		}
	}

	activeOrders, err := s.orderRepo.GetAll()
	if err != nil {
		return err
	}

	var tableOrders []*models.Order
	for i := range activeOrders {
		o := &activeOrders[i]
		if o.TableID != nil && *o.TableID == tableID &&
			o.Status != models.StatusDelivered && o.Status != models.StatusCancelled {
			tableOrders = append(tableOrders, o)
		}
	}

	if len(tableOrders) == 0 {
		return nil
	}

	// Calculate grand total for payment allocation
	var grandTotal float64
	for _, o := range tableOrders {
		full, err := s.orderRepo.GetByID(o.ID)
		if err != nil || full == nil {
			return fmt.Errorf("order %d summasini olishda xatolik", o.ID)
		}
		grandTotal += full.TotalPrice
	}

	if len(payments) == 0 {
		return fmt.Errorf("stolni yopishdan oldin to'lov usulini tanlang")
	}
	if len(payments) == 1 && payments[0].Amount == 0 {
		// Backward compatibility for CloseTable(paymentMethod), which did not
		// carry an amount.
		payments[0].Amount = grandTotal
	}
	var paymentTotal float64
	for _, payment := range payments {
		if payment.Amount <= 0 {
			return fmt.Errorf("to'lov summasi musbat bo'lishi kerak")
		}
		paymentTotal += payment.Amount
	}
	if math.Abs(paymentTotal-grandTotal) > 0.01 {
		return fmt.Errorf("to'lov summasi order summasiga teng bo'lishi kerak: %.2f / %.2f", paymentTotal, grandTotal)
	}

	// Determine primary payment method
	primaryMethod := ""
	if len(payments) == 1 {
		primaryMethod = payments[0].Method
		// If amount is 0 (backward compat), set it to grand total
		if payments[0].Amount == 0 {
			payments[0].Amount = grandTotal
		}
	} else if len(payments) > 1 {
		primaryMethod = "mixed"
	}

	// Calculate total for proportional payment allocation across multiple orders
	var orderTotals []float64
	for _, o := range tableOrders {
		full, err := s.orderRepo.GetByID(o.ID)
		if err != nil || full == nil {
			orderTotals = append(orderTotals, o.TotalPrice)
		} else {
			orderTotals = append(orderTotals, full.TotalPrice)
		}
	}

	for idx, o := range tableOrders {
		if primaryMethod != "" {
			if err := s.orderRepo.SetPaymentMethod(o.ID, primaryMethod); err != nil {
				return fmt.Errorf("failed to set payment method for order %d: %w", o.ID, err)
			}
		}

		// Save payments to order_payments for EVERY order, proportionally distributed
		if len(payments) > 0 {
			orderShare := orderTotals[idx] / grandTotal // proportion of this order
			for _, p := range payments {
				proRataAmount := p.Amount * orderShare
				if proRataAmount < 0.01 {
					continue
				}
				if _, err := s.orderRepo.GetDB().Exec(`
						INSERT INTO order_payments (order_id, method, amount)
						VALUES ($1, $2, $3)
						ON CONFLICT DO NOTHING
					`, o.ID, p.Method, proRataAmount); err != nil {
					return fmt.Errorf("to'lovni saqlashda xatolik (order %d): %w", o.ID, err)
				}
			}
		}

		if err := s.orderRepo.UpdateStatus(o.ID, models.StatusDelivered, nil); err != nil {
			return fmt.Errorf("failed to close order %d: %w", o.ID, err)
		}
	}

	firstOrder := tableOrders[0]

	populated, err := s.orderRepo.GetByID(firstOrder.ID)
	if err != nil || populated == nil {
		populated = firstOrder
	}

	var allItems []models.OrderItem
	var grandServiceFee float64
	var servicePercentage float64
	grandTotal = 0

	for _, o := range tableOrders {
		full, err := s.orderRepo.GetByID(o.ID)
		if err != nil || full == nil {
			return fmt.Errorf("order %d ma'lumotlarini olishda xatolik", o.ID)
		}
		allItems = append(allItems, full.Items...)
		grandTotal += full.TotalPrice
		grandServiceFee += full.ServiceFee
		if full.ServicePercentage > 0 {
			servicePercentage = full.ServicePercentage
		}
	}
	combinedOrder := &models.Order{
		ID:                populated.ID,
		TableID:           populated.TableID,
		TableName:         populated.TableName,
		WaiterID:          populated.WaiterID,
		WaiterName:        populated.WaiterName,
		TotalPrice:        grandTotal,
		ServiceFee:        grandServiceFee,
		ServicePercentage: servicePercentage,
		Items:             allItems,
		CreatedAt:         populated.CreatedAt,
		Status:            models.StatusDelivered,
		PaymentMethod:     primaryMethod,
	}

	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "close_order", "order": combinedOrder})
	go s.printerService.PrintOrder(combinedOrder)

	return nil
}

func (s *OrderService) GetActiveOrderByTable(tableID int) (*models.Order, error) {
	return s.orderRepo.FindActiveOrderByTableID(tableID)
}
func (s *OrderService) AddItemsToExistingOrder(orderID int, items []models.OrderItem, userID int, role string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil || order == nil {
		return nil, fmt.Errorf("buyurtma topilmadi (ID: %d)", orderID)
	}

	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return nil, fmt.Errorf("bu buyurtma allaqachon yopilgan")
	}
	if role != "admin" {
		if order.WaiterID != nil && *order.WaiterID != userID {
			return nil, fmt.Errorf("вы можете добавлять товары только в свои заказы")
		}
	}
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

		target, err := s.productRepo.GetPrinterTarget(item.ProductID)
		if err == nil {
			item.PrinterTarget = target
		} else {
			item.PrinterTarget = "ALL"
		}
	}

	if err := s.orderRepo.AddItemsToOrder(orderID, items); err != nil {
		return nil, err
	}

	updatedOrder, err := s.orderRepo.GetByID(orderID)
	if err != nil || updatedOrder == nil {
		return nil, fmt.Errorf("yangilangan buyurtmani olishda xatolik")
	}

	for i := range items {
		prod, _ := s.productRepo.GetByID(items[i].ProductID)
		if prod != nil {
			items[i].ProductName = prod.Name
		}
	}

	partialOrder := &models.Order{
		ID:         updatedOrder.ID,
		TableID:    updatedOrder.TableID,
		TableName:  updatedOrder.TableName,
		WaiterID:   updatedOrder.WaiterID,
		WaiterName: updatedOrder.WaiterName,
		TotalPrice: updatedOrder.TotalPrice,
		Items:      items,
		CreatedAt:  updatedOrder.CreatedAt,
		Status:     updatedOrder.Status,
	}

	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": partialOrder})
	s.wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": partialOrder})

	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "is_dop": true, "order": partialOrder})
	go s.printerService.PrintOrder(partialOrder)

	dummyOrder := map[string]interface{}{
		"id":    -9999,
		"items": []interface{}{},
	}
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": dummyOrder})

	return updatedOrder, nil
}
func (s *OrderService) CancelProductFromOrder(orderID, productID int, cancelQty float64) error {
	// 1. Fetch order
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("order not found")
	}
	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return fmt.Errorf("yopilgan buyurtmadan mahsulotni bekor qilib bo'lmaydi")
	}

	if cancelQty <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}

	// 2. Find all items with this product_id
	var matchedItems []models.OrderItem
	var totalQty float64
	for _, it := range order.Items {
		if it.ProductID == productID {
			matchedItems = append(matchedItems, it)
			totalQty += it.Quantity
		}
	}

	if len(matchedItems) == 0 {
		return fmt.Errorf("product not found in order")
	}

	if cancelQty > totalQty {
		cancelQty = totalQty
	}

	// 3. Cancel from database incrementally
	remainingToCancel := cancelQty
	for i := len(matchedItems) - 1; i >= 0 && remainingToCancel > 0; i-- {
		it := matchedItems[i]
		qtyToCancelHere := remainingToCancel
		if qtyToCancelHere > it.Quantity {
			qtyToCancelHere = it.Quantity
		}

		err := s.orderRepo.CancelItem(orderID, it.ID, qtyToCancelHere)
		if err != nil {
			return err
		}
		remainingToCancel -= qtyToCancelHere
	}

	// 4. Broadcast single cancel event to printer
	// We use the first matched item as the base for the WS payload (it has the correct names and printer_target)
	cancelledItem := matchedItems[0]
	cancelledItem.Quantity = cancelQty

	cancelPayload := map[string]interface{}{
		"type":        "cancel_item",
		"order_id":    orderID,
		"item":        cancelledItem,
		"waiter_name": order.WaiterName,
		"table_name":  order.TableName,
	}
	s.wsService.BroadcastToRole("printer", cancelPayload)

	return nil
}

func (s *OrderService) TransferTable(fromTableID int, toTableID int) error {
	// 1. Get active order on fromTableID
	order, err := s.orderRepo.FindActiveOrderByTableID(fromTableID)
	if err != nil || order == nil {
		return fmt.Errorf("tanlangan stolda faol buyurtma topilmadi")
	}

	// 2. Check if toTableID is free
	toTable, err := s.tableRepo.GetByID(toTableID)
	if err != nil {
		return fmt.Errorf("yangi stolni topishda xatolik: %v", err)
	}
	if toTable.Status != "free" {
		return fmt.Errorf("tanlangan yangi stol band. Iltimos bo'sh stol tanlang")
	}

	// 3. Update order's table_id
	if err := s.orderRepo.TransferTable(order.ID, toTableID); err != nil {
		return fmt.Errorf("buyurtmani ko'chirishda xatolik: %v", err)
	}

	// 4. Update table statuses
	if err := s.tableRepo.UpdateStatus(fromTableID, "free"); err != nil {
		return fmt.Errorf("eski stol holatini yangilashda xatolik: %v", err)
	}
	if err := s.tableRepo.UpdateStatus(toTableID, "occupied"); err != nil {
		return fmt.Errorf("yangi stol holatini yangilashda xatolik: %v", err)
	}

	// 5. Notify clients (Waiters and Admins) to refresh
	s.wsService.BroadcastToRole("waiter", map[string]interface{}{"type": "tables_updated"})
	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "tables_updated"})

	return nil
}

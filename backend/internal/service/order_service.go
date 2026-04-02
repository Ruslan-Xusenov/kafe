package service

import (
	"fmt"

	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	productRepo *repository.ProductRepository
	wsService    *WebsocketService
	botService   *BotService
	printerService *PrinterService
}

func NewOrderService(orderRepo *repository.OrderRepository, productRepo *repository.ProductRepository, wsService *WebsocketService, botService *BotService, printerService *PrinterService) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		wsService:    wsService,
		botService:   botService,
		printerService: printerService,
	}
}

func (s *OrderService) CreateOrder(order *models.Order) error {
	var total float64
	for i := range order.Items {
		item := &order.Items[i]
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to get product %d: %w", item.ProductID, err)
		}
		if prod == nil {
			return fmt.Errorf("product %d not found", item.ProductID)
		}
		item.Price = prod.Price
		total += item.Price * float64(item.Quantity)
	}
	order.TotalPrice = total
	order.Status = models.StatusNew

	if err := s.orderRepo.Create(order); err != nil {
		return err
	}

	// Enrich order items with product names for notification
	for i := range order.Items {
		prod, _ := s.productRepo.GetByID(order.Items[i].ProductID)
		if prod != nil {
			order.Items[i].ProductName = prod.Name
		}
	}

	// Trigger Task: Send Notification to Telegram
	var firstImageUrl *string
	if len(order.Items) > 0 {
		prod, err := s.productRepo.GetByID(order.Items[0].ProductID)
		if err == nil && prod != nil {
			firstImageUrl = prod.ImageURL
		}
	}
	s.botService.SendNewOrderNotification(order, firstImageUrl)
	
	// Real-time: Notify Cooks, Admin and Printer
	s.wsService.BroadcastToRole("admin", map[string]interface{}{"type": "new_order", "order": order})
	s.wsService.BroadcastToRole("cook", map[string]interface{}{"type": "new_order", "order": order})
	s.wsService.BroadcastToRole("printer", map[string]interface{}{"type": "new_order", "order": order})

	// Direct Print
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
	if role != "admin" && role != "cook" && role != "courier" && order.CustomerID != userID {
		return nil, fmt.Errorf("unauthorized access to order")
	}

	return order, nil
}

func (s *OrderService) GetCustomerOrders(customerID int) ([]models.Order, error) {
	return s.orderRepo.GetByCustomerID(customerID)
}

func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *OrderService) GetActiveOrders() ([]models.Order, error) {
	// Active orders for Kitchen/Admin (status not delivered or cancelled)
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
	var cookID *int
	if role == "cook" && status == models.StatusPreparing {
		cookID = &userID
	}

	err := s.orderRepo.UpdateStatus(orderID, status, cookID)
	if err == nil {
		order, _ := s.orderRepo.GetByID(orderID)
		if order != nil {
			// Notify Customer
			s.wsService.BroadcastToUser(order.CustomerID, map[string]interface{}{"type": "status_update", "status": status, "order_id": orderID})

			// Notify Roles
			updatePayload := map[string]interface{}{"type": "status_update", "status": status, "order_id": orderID}
			s.wsService.BroadcastToRole("admin", updatePayload)
			s.wsService.BroadcastToRole("cook", updatePayload)
			s.wsService.BroadcastToRole("courier", updatePayload)

			// Notify bot for specific statuses
			if status == models.StatusReady {
				s.botService.SendOrderStatusNotification(order, "TAYYOR - Kuryerlar olib ketishi mumkin")
			} else if status == models.StatusDelivered {
				s.botService.SendOrderStatusNotification(order, "YETKAZIB BERILDI")
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
			s.wsService.BroadcastToUser(order.CustomerID, map[string]interface{}{"type": "status_update", "status": models.StatusOnWay, "order_id": orderID})

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

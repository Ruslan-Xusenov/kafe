package service

import (
	"fmt"
	"kafe/internal/models"
)

func (s *OrderService) BulkEditOrder(orderID int, addItems []models.OrderItem, cancelItems []struct {
	ProductID int     `json:"product_id"`
	Quantity  float64 `json:"quantity"`
}, userID int, role string) error {
	order, err := s.orderRepo.GetByID(orderID)
	if err != nil || order == nil {
		return fmt.Errorf("buyurtma topilmadi")
	}

	if order.Status == models.StatusDelivered || order.Status == models.StatusCancelled {
		return fmt.Errorf("bu buyurtma allaqachon yopilgan")
	}

	var addedItemsForPrint []models.OrderItem
	var cancelledItemsForPrint []models.OrderItem

	// 1. Process Additions
	if len(addItems) > 0 {
		for i := range addItems {
			item := &addItems[i]
			prod, err := s.productRepo.GetByID(item.ProductID)
			if err == nil && prod != nil {
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
				addedItemsForPrint = append(addedItemsForPrint, *item)
			}
		}
		if len(addedItemsForPrint) > 0 {
			err = s.orderRepo.AddItemsToOrder(orderID, addedItemsForPrint)
			if err != nil {
				return err
			}
		}
	}

	// 2. Process Cancellations
	if len(cancelItems) > 0 {
		// re-fetch order to get updated items
		currentOrder, _ := s.orderRepo.GetByID(orderID)
		if currentOrder != nil {
			for _, cancelReq := range cancelItems {
				var matchedItems []models.OrderItem
				var totalQty float64
				for _, it := range currentOrder.Items {
					if it.ProductID == cancelReq.ProductID {
						matchedItems = append(matchedItems, it)
						totalQty += it.Quantity
					}
				}
				if len(matchedItems) > 0 {
					cancelQty := cancelReq.Quantity
					if cancelQty > totalQty {
						cancelQty = totalQty
					}
					
					remainingToCancel := cancelQty
					for i := len(matchedItems) - 1; i >= 0 && remainingToCancel > 0; i-- {
						it := matchedItems[i]
						qtyToCancelHere := remainingToCancel
						if qtyToCancelHere > it.Quantity {
							qtyToCancelHere = it.Quantity
						}
						s.orderRepo.CancelItem(orderID, it.ID, qtyToCancelHere)
						remainingToCancel -= qtyToCancelHere
					}
					
					cancelledItem := matchedItems[0]
					cancelledItem.Quantity = cancelQty
					cancelledItemsForPrint = append(cancelledItemsForPrint, cancelledItem)
				}
			}
		}
	}

	// 3. Broadcast single bulk_edit event
	if len(addedItemsForPrint) > 0 || len(cancelledItemsForPrint) > 0 {
		bulkPayload := map[string]interface{}{
			"type":            "bulk_edit",
			"order_id":        orderID,
			"added_items":     addedItemsForPrint,
			"cancelled_items": cancelledItemsForPrint,
			"waiter_name":     order.WaiterName,
			"table_number":    order.TableNumber,
		}
		s.wsService.BroadcastToRole("printer", bulkPayload)
	}

	return nil
}

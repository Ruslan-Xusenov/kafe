package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db *sqlx.DB
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetDB() *sqlx.DB {
	return r.db
}

func (r *OrderRepository) Create(order *models.Order) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var defaultPercentage float64
	if order.WaiterID != nil && order.TableID != nil {
		_ = tx.Get(&defaultPercentage, `SELECT COALESCE(default_service_percentage, 0) FROM users WHERE id = $1`, *order.WaiterID)
	}
	if defaultPercentage > 0 {
		order.ServicePercentage = defaultPercentage
		order.ServiceFee = order.TotalPrice * defaultPercentage / 100
		order.TotalPrice += order.ServiceFee
	}

	// Insert Order
	query := `INSERT INTO orders (customer_id, total_price, status, address, phone, lat, lng, comment, table_id, waiter_id, service_percentage, service_fee) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(query, order.CustomerID, order.TotalPrice, order.Status, order.Address, order.Phone, order.Lat, order.Lng, order.Comment, order.TableID, order.WaiterID, order.ServicePercentage, order.ServiceFee).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Order Items
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, price, unit, comment) 
                  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	for i := range order.Items {
		item := &order.Items[i]
		err = tx.QueryRow(itemQuery, order.ID, item.ProductID, item.Quantity, item.Price, item.Unit, item.Comment).
			Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(id int) (*models.Order, error) {
	var order models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.id = $1
	`
	err := r.db.Get(&order, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Get Items
	var items []models.OrderItem
	itemQuery := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit, 
			   COALESCE(oi.comment, '') as comment, oi.created_at,
			   COALESCE(p.name, 'Noma''lum') as product_name,
			   COALESCE(c.printer_target, 'ALL') as printer_target
		FROM order_items oi
		LEFT JOIN products p ON oi.product_id = p.id
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE oi.order_id = $1
	`
	err = r.db.Select(&items, itemQuery, id)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return &order, nil
}

func (r *OrderRepository) GetByCustomerID(customerID int) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.customer_id = $1 
		ORDER BY o.created_at DESC
	`
	err := r.db.Select(&orders, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("GetByCustomerID error: %w", err)
	}

	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit, 
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		_ = r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderRepository) GetAll() ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		ORDER BY o.created_at DESC
	`
	err := r.db.Select(&orders, query)
	if err != nil {
		fmt.Printf("DATABASE_DEBUG: GetAll select error: %v\n", err)
		return nil, fmt.Errorf("GetAll error: %w", err)
	}

	fmt.Printf("DATABASE_DEBUG: GetAll returned %d orders from DB\n", len(orders))
	
	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit, 
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		_ = r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}
	
	return orders, nil
}

func (r *OrderRepository) GetByStatus(status models.OrderStatus) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.status = $1 
		ORDER BY o.created_at ASC
	`
	err := r.db.Select(&orders, query, status)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit,
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderRepository) UpdateStatus(orderID int, status models.OrderStatus, cookID *int) error {
	var query string
	if cookID != nil {
		query = `UPDATE orders SET status = $1, cook_id = $2, updated_at = NOW() WHERE id = $3`
		_, err := r.db.Exec(query, status, *cookID, orderID)
		return err
	}
	query = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, orderID)
	return err
}

func (r *OrderRepository) AddStaffRating(rating *models.StaffRating) error {
	query := `
		INSERT INTO staff_ratings (order_id, staff_id, staff_role, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (order_id, staff_id, staff_role) 
		DO UPDATE SET 
			rating = EXCLUDED.rating,
			comment = EXCLUDED.comment,
			created_at = NOW()
	`
	_, err := r.db.Exec(query, rating.OrderID, rating.StaffID, rating.StaffRole, rating.Rating, rating.Comment)
	return err
}

func (r *OrderRepository) GetStaffPerformance() ([]models.StaffPerformance, error) {
	var performance []models.StaffPerformance
	query := `
		SELECT 
			u.id as staff_id,
			u.full_name,
			u.role,
			COALESCE(AVG(sr.rating), 0) as avg_rating,
			COUNT(sr.id) as total_reviews,
			COUNT(sr.id) FILTER (WHERE sr.rating >= 4) as good_reviews,
			COUNT(sr.id) FILTER (WHERE sr.rating <= 2) as bad_reviews
		FROM users u
		JOIN staff_ratings sr ON u.id = sr.staff_id
		GROUP BY u.id, u.full_name, u.role
		ORDER BY avg_rating DESC
	`
	err := r.db.Select(&performance, query)
	return performance, err
}

func (r *OrderRepository) AssignCourier(orderID int, courierID int) error {
	query := `UPDATE orders SET courier_id = $1, status = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, courierID, models.StatusOnWay, orderID)
	return err
}

func (r *OrderRepository) GetCourierStats(courierID int) (*models.DeliveryStats, error) {
	stats := &models.DeliveryStats{}
	query := `
	SELECT 
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE) as today,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '3 days') as three_day,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 week') as week,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 month') as month,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 year') as year
	FROM orders 
	WHERE courier_id = $1 AND status = 'delivered'
	`
	err := r.db.Get(stats, query, courierID)
	return stats, err
}

func (r *OrderRepository) GetKitchenStats() (*models.DeliveryStats, error) {
	stats := &models.DeliveryStats{}
	query := `
	SELECT 
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE) as today,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '3 days') as three_day,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 week') as week,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 month') as month,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 year') as year
	FROM orders 
	WHERE created_at >= CURRENT_DATE AND status IN ('new', 'preparing', 'ready', 'on_way', 'delivered')
	`
	err := r.db.Get(stats, query)
	return stats, err
}
func (r *OrderRepository) GetAdminStats() (*models.DeliveryStats, error) {
	stats := &models.DeliveryStats{}
	query := `
	SELECT 
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE) as today,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '3 days') as three_day,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 week') as week,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 month') as month,
		COUNT(*) FILTER (WHERE updated_at >= CURRENT_DATE - INTERVAL '1 year') as year
	FROM orders 
	WHERE status = 'delivered'
	`
	err := r.db.Get(stats, query)
	return stats, err
}

func (r *OrderRepository) GetRatingsByOrderID(orderID int) ([]models.StaffRating, error) {
	var ratings []models.StaffRating
	query := `SELECT * FROM staff_ratings WHERE order_id = $1`
	err := r.db.Select(&ratings, query, orderID)
	return ratings, err
}

func (r *OrderRepository) GetLastOrderByPhone(phone string) (*models.Order, error) {
	var order models.Order
	query := `SELECT * FROM orders WHERE phone = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.Get(&order, query, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) SetServiceFee(orderID int, percentage float64, fee float64, newTotal float64) error {
	query := `UPDATE orders SET service_percentage = $1, service_fee = $2, total_price = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(query, percentage, fee, newTotal, orderID)
	return err
}

func (r *OrderRepository) SetPaymentMethod(orderID int, method string) error {
	query := `UPDATE orders SET payment_method = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, method, orderID)
	return err
}

// GetWaiterHistory returns all cafe (table-based) orders, newest first, with waiter info
func (r *OrderRepository) GetWaiterHistory() ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.table_id IS NOT NULL
		ORDER BY o.created_at DESC
		LIMIT 200
	`
	err := r.db.Select(&orders, query)
	if err != nil {
		return nil, fmt.Errorf("GetWaiterHistory error: %w", err)
	}

	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit, 
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		_ = r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}

	return orders, nil
}

// GetActiveByWaiterID returns all active (non-delivered, non-cancelled) table orders for a waiter
func (r *OrderRepository) GetActiveByWaiterID(waiterID int) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.waiter_id = $1 
		  AND o.table_id IS NOT NULL
		  AND o.status NOT IN ('delivered', 'cancelled')
		ORDER BY o.created_at DESC
	`
	err := r.db.Select(&orders, query, waiterID)
	if err != nil {
		return nil, fmt.Errorf("GetActiveByWaiterID error: %w", err)
	}
	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit,
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		_ = r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}
	return orders, nil
}

// GetHistoryByWaiterID returns all completed table orders for a specific waiter
func (r *OrderRepository) GetHistoryByWaiterID(waiterID int) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id, COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.number as table_number
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.waiter_id = $1
		  AND o.table_id IS NOT NULL
		ORDER BY o.created_at DESC
		LIMIT 300
	`
	err := r.db.Select(&orders, query, waiterID)
	if err != nil {
		return nil, fmt.Errorf("GetHistoryByWaiterID error: %w", err)
	}
	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.unit,
				   COALESCE(oi.comment, '') as comment, oi.created_at,
				   COALESCE(p.name, 'Noma''lum') as product_name,
				   COALESCE(c.printer_target, 'ALL') as printer_target
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE oi.order_id = $1
		`
		_ = r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}
	return orders, nil
}

// RemoveItem deletes an item from an order and recalculates the total price and service fee
func (r *OrderRepository) RemoveItem(orderID, itemID int) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete item
	res, err := tx.Exec(`DELETE FROM order_items WHERE id = $1 AND order_id = $2`, itemID, orderID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("item not found")
	}

	// Recalculate order total price
	var newTotal float64
	err = tx.Get(&newTotal, `SELECT COALESCE(SUM(price * quantity), 0) FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return err
	}

	// Fetch current order to recalculate service fee
	var servicePercentage float64
	err = tx.Get(&servicePercentage, `SELECT COALESCE(service_percentage, 0) FROM orders WHERE id = $1`, orderID)
	if err != nil {
		return err
	}

	serviceFee := newTotal * servicePercentage / 100
	finalTotal := newTotal + serviceFee

	_, err = tx.Exec(`UPDATE orders SET total_price = $1, service_fee = $2, updated_at = NOW() WHERE id = $3`, finalTotal, serviceFee, orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type OrderRepository struct {
	db *sqlx.DB
}

type ingredientStockRow struct {
	IngredientID   int     `db:"ingredient_id"`
	IngredientName string  `db:"ingredient_name"`
	Stock          float64 `db:"stock"`
	RequiredQty    float64 `db:"required_qty"`
}

func NewOrderRepository(db *sqlx.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetDB() *sqlx.DB {
	return r.db
}

func deductStockForProductTx(tx *sqlx.Tx, productID int, productQuantity float64) error {
	if productQuantity <= 0 {
		return fmt.Errorf("product quantity must be greater than zero")
	}

	var ingredients []ingredientStockRow
	err := tx.Select(&ingredients, `
		SELECT
			i.id as ingredient_id,
			i.name as ingredient_name,
			i.stock,
			(pi.quantity * $2) as required_qty
		FROM product_ingredients pi
		JOIN ingredients i ON i.id = pi.ingredient_id
		WHERE pi.product_id = $1
		FOR UPDATE OF i
	`, productID, productQuantity)
	if err != nil {
		return err
	}

	for _, ing := range ingredients {
		if ing.Stock < ing.RequiredQty {
			return fmt.Errorf("skladda %s yetarli emas: kerak %.3f, mavjud %.3f", ing.IngredientName, ing.RequiredQty, ing.Stock)
		}
	}

	for _, ing := range ingredients {
		if _, err := tx.Exec(`
			UPDATE ingredients
			SET stock = stock - $1, updated_at = NOW()
			WHERE id = $2
		`, ing.RequiredQty, ing.IngredientID); err != nil {
			return err
		}
	}

	return nil
}

func restoreStockForProductTx(tx *sqlx.Tx, productID int, productQuantity float64) error {
	if productQuantity <= 0 {
		return nil
	}

	_, err := tx.Exec(`
		UPDATE ingredients
		SET stock = stock + (pi.quantity * $2), updated_at = NOW()
		FROM product_ingredients pi
		WHERE ingredients.id = pi.ingredient_id AND pi.product_id = $1
	`, productID, productQuantity)
	return err
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.create(order, nil)
}

// CreateWithPayments creates an order, deducts stock and records all payments
// in one transaction. This is used by POS sales so a failed payment cannot
// leave behind an unpaid order or a partially deducted stock balance.
func (r *OrderRepository) CreateWithPayments(order *models.Order, payments []models.PaymentInput) error {
	return r.create(order, payments)
}

func (r *OrderRepository) create(order *models.Order, payments []models.PaymentInput) error {
	if len(payments) > 0 {
		validMethods := map[string]bool{"cash": true, "card": true, "click": true, "nasiya": true}
		for _, payment := range payments {
			if !validMethods[payment.Method] {
				return fmt.Errorf("noto'g'ri to'lov usuli: %s", payment.Method)
			}
			if payment.Amount <= 0 {
				return fmt.Errorf("to'lov summasi musbat bo'lishi kerak")
			}
		}
	}

	// Fetch table service percentage from settings for table orders
	var defaultPercentage float64
	if order.TableID != nil {
		var percStr string
		err := r.db.Get(&percStr, `SELECT COALESCE((SELECT value FROM settings WHERE key = 'table_service_percentage'), '10')`)
		if err == nil && percStr != "" {
			if f, parseErr := strconv.ParseFloat(percStr, 64); parseErr == nil {
				defaultPercentage = f
			}
		}
	}
	if defaultPercentage > 0 {
		order.ServicePercentage = defaultPercentage
		order.ServiceFee = order.TotalPrice * defaultPercentage / 100
		order.TotalPrice += order.ServiceFee
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert Order
	query := `INSERT INTO orders (customer_id, total_price, status, address, phone, lat, lng, comment, table_id, waiter_id, service_percentage, service_fee, shift_id, idempotency_key) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(query, order.CustomerID, order.TotalPrice, order.Status, order.Address, order.Phone, order.Lat, order.Lng, order.Comment, order.TableID, order.WaiterID, order.ServicePercentage, order.ServiceFee, order.ShiftID, order.IdempotencyKey).
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
		if err := deductStockForProductTx(tx, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, err)
		}
	}

	if len(payments) > 0 {
		primaryMethod := payments[0].Method
		if len(payments) > 1 {
			primaryMethod = "mixed"
		}
		for _, payment := range payments {
			if _, err := tx.Exec(`
				INSERT INTO order_payments (order_id, method, amount)
				VALUES ($1, $2, $3)
			`, order.ID, payment.Method, payment.Amount); err != nil {
				return fmt.Errorf("to'lov saqlashda xatolik: %w", err)
			}
		}
		if _, err := tx.Exec(`UPDATE orders SET payment_method = $1, updated_at = NOW() WHERE id = $2`, primaryMethod, order.ID); err != nil {
			return fmt.Errorf("to'lov usulini saqlashda xatolik: %w", err)
		}
	}

	return tx.Commit()
}

func (r *OrderRepository) GetByID(id int) (*models.Order, error) {
	var order models.Order
	query := `
		SELECT o.id, o.customer_id, o.total_price, o.status, o.address, o.phone, 
			   o.lat, o.lng, o.courier_id, o.cook_id, o.table_id, o.waiter_id,
			   o.shift_id, o.idempotency_key, COALESCE(o.stock_restored, false) as stock_restored,
			   COALESCE(o.comment, '') as comment, 
			   COALESCE(o.service_percentage, 0) as service_percentage, COALESCE(o.service_fee, 0) as service_fee,
			   COALESCE(o.payment_method, '') as payment_method,
			   o.created_at, o.updated_at,
			   COALESCE(u1.full_name, '') as courier_name, 
			   COALESCE(u2.full_name, '') as cook_name,
			   COALESCE(u3.full_name, '') as waiter_name,
			   t.name as table_name
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
			   t.name as table_name
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
			   t.name as table_name
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
			   t.name as table_name
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
		query = `UPDATE orders SET status = $1, cook_id = $2, updated_at = NOW() WHERE id = $3 AND status NOT IN ('delivered', 'cancelled')`
		result, err := r.db.Exec(query, status, *cookID, orderID)
		if err == nil {
			rows, rowsErr := result.RowsAffected()
			if rowsErr == nil && rows == 0 {
				return errors.New("order yopilgan yoki topilmadi")
			}
		}
		return err
	}
	query = `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2 AND status NOT IN ('delivered', 'cancelled')`
	result, err := r.db.Exec(query, status, orderID)
	if err == nil {
		rows, rowsErr := result.RowsAffected()
		if rowsErr == nil && rows == 0 {
			return errors.New("order yopilgan yoki topilmadi")
		}
	}
	return err
}

// MarkStockRestored atomically marks an order's stock as restored.
// Returns false if the stock was already restored (idempotent).
func (r *OrderRepository) MarkStockRestored(orderID int) (bool, error) {
	result, err := r.db.Exec(
		`UPDATE orders SET stock_restored = TRUE WHERE id = $1 AND stock_restored = FALSE`,
		orderID,
	)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// FindByIdempotencyKey looks up an existing order by idempotency key.
// Returns nil (not an error) when no order is found.
func (r *OrderRepository) FindByIdempotencyKey(key string) (*models.Order, error) {
	if key == "" {
		return nil, nil
	}
	var id int
	err := r.db.Get(&id, `SELECT id FROM orders WHERE idempotency_key = $1 LIMIT 1`, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.GetByID(id)
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
			   t.name as table_name
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
			   t.name as table_name
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
			   t.name as table_name
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

// CancelItem cancels a specific quantity of an item. If cancelQty >= current qty, it deletes the item.
func (r *OrderRepository) CancelItem(orderID, itemID int, cancelQty float64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current quantity
	var item struct {
		ProductID int     `db:"product_id"`
		Quantity  float64 `db:"quantity"`
	}
	err = tx.Get(&item, `
		SELECT oi.product_id, oi.quantity
		FROM order_items oi
		JOIN orders o ON o.id = oi.order_id
		WHERE oi.id = $1 AND oi.order_id = $2
		  AND o.status NOT IN ('delivered', 'cancelled')
	`, itemID, orderID)
	if err != nil {
		return errors.New("item not found")
	}

	restoreQty := cancelQty
	if restoreQty > item.Quantity {
		restoreQty = item.Quantity
	}

	if cancelQty >= item.Quantity {
		// Delete item
		_, err := tx.Exec(`DELETE FROM order_items WHERE id = $1`, itemID)
		if err != nil {
			return err
		}
	} else {
		// Update item
		_, err := tx.Exec(`UPDATE order_items SET quantity = quantity - $1 WHERE id = $2`, cancelQty, itemID)
		if err != nil {
			return err
		}
	}

	if err := restoreStockForProductTx(tx, item.ProductID, restoreQty); err != nil {
		return fmt.Errorf("failed to restore stock for product %d: %w", item.ProductID, err)
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

// FindActiveOrderByTableID returns the first active (non-delivered, non-cancelled) order for a table
func (r *OrderRepository) FindActiveOrderByTableID(tableID int) (*models.Order, error) {
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
			   t.name as table_name
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		LEFT JOIN users u3 ON o.waiter_id = u3.id
		LEFT JOIN tables t ON o.table_id = t.id
		WHERE o.table_id = $1 
		  AND o.status NOT IN ('delivered', 'cancelled')
		ORDER BY o.created_at DESC
		LIMIT 1
	`
	err := r.db.Get(&order, query, tableID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Fetch items
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
	_ = r.db.Select(&items, itemQuery, order.ID)
	order.Items = items

	return &order, nil
}

// AddItemsToOrder appends new items to an existing order and recalculates totals
func (r *OrderRepository) AddItemsToOrder(orderID int, items []models.OrderItem) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert new items
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, price, unit, comment) 
                  VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	for i := range items {
		item := &items[i]
		err = tx.QueryRow(itemQuery, orderID, item.ProductID, item.Quantity, item.Price, item.Unit, item.Comment).
			Scan(&item.ID, &item.CreatedAt)
		if err != nil {
			return fmt.Errorf("failed to insert order item: %w", err)
		}
		if err := deductStockForProductTx(tx, item.ProductID, item.Quantity); err != nil {
			return fmt.Errorf("failed to deduct stock for product %d: %w", item.ProductID, err)
		}
	}

	// Recalculate order total from ALL items (old + new)
	var newBaseTotal float64
	err = tx.Get(&newBaseTotal, `SELECT COALESCE(SUM(price * quantity), 0) FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return err
	}

	// Get service percentage to recalculate service fee
	var servicePercentage float64
	err = tx.Get(&servicePercentage, `SELECT COALESCE(service_percentage, 0) FROM orders WHERE id = $1`, orderID)
	if err != nil {
		return err
	}

	serviceFee := newBaseTotal * servicePercentage / 100
	finalTotal := newBaseTotal + serviceFee

	_, err = tx.Exec(`UPDATE orders SET total_price = $1, service_fee = $2, updated_at = NOW() WHERE id = $3`, finalTotal, serviceFee, orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *OrderRepository) TransferTable(orderID int, newTableID int) error {
	query := `UPDATE orders SET table_id = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, newTableID, orderID)
	return err
}

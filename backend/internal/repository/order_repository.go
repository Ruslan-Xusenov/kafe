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

func (r *OrderRepository) Create(order *models.Order) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert Order
	query := `INSERT INTO orders (customer_id, total_price, status, address, phone, lat, lng, comment) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at, updated_at`
	err = tx.QueryRow(query, order.CustomerID, order.TotalPrice, order.Status, order.Address, order.Phone, order.Lat, order.Lng, order.Comment).
		Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Order Items
	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, price, comment) 
                  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	for i := range order.Items {
		item := &order.Items[i]
		err = tx.QueryRow(itemQuery, order.ID, item.ProductID, item.Quantity, item.Price, item.Comment).
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
		SELECT o.*, 
			   u1.full_name as courier_name, 
			   u2.full_name as cook_name
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
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
		SELECT oi.*, p.name as product_name 
		FROM order_items oi
		LEFT JOIN products p ON oi.product_id = p.id
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
		SELECT o.*, 
			   u1.full_name as courier_name, 
			   u2.full_name as cook_name
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		WHERE o.customer_id = $1 
		ORDER BY o.created_at DESC
	`
	err := r.db.Select(&orders, query, customerID)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.*, p.name as product_name 
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = $1
		`
		r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}

	return orders, nil
}

func (r *OrderRepository) GetAll() ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.*, 
			   u1.full_name as courier_name, 
			   u2.full_name as cook_name
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
		ORDER BY o.created_at DESC
	`
	err := r.db.Select(&orders, query)
	if err != nil {
		return nil, err
	}
	
	// Preload items (simple approach for small quantities)
	for i := range orders {
		var items []models.OrderItem
		itemQuery := `
			SELECT oi.*, p.name as product_name 
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
			WHERE oi.order_id = $1
		`
		r.db.Select(&items, itemQuery, orders[i].ID)
		orders[i].Items = items
	}
	
	return orders, nil
}

func (r *OrderRepository) GetByStatus(status models.OrderStatus) ([]models.Order, error) {
	var orders []models.Order
	query := `
		SELECT o.*, 
			   u1.full_name as courier_name, 
			   u2.full_name as cook_name
		FROM orders o
		LEFT JOIN users u1 ON o.courier_id = u1.id
		LEFT JOIN users u2 ON o.cook_id = u2.id
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
			SELECT oi.*, p.name as product_name 
			FROM order_items oi
			LEFT JOIN products p ON oi.product_id = p.id
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

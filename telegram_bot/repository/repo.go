package repository

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"telegram_bot/models"
)

type BotRepository struct {
	db *sqlx.DB
}

func NewBotRepository(db *sqlx.DB) *BotRepository {
	return &BotRepository{db: db}
}

func (r *BotRepository) GetDB() *sqlx.DB {
	return r.db
}

// Categories
func (r *BotRepository) GetAllCategories() ([]models.Category, error) {
	var categories []models.Category
	query := `SELECT * FROM categories ORDER BY id ASC`
	err := r.db.Select(&categories, query)
	return categories, err
}

func (r *BotRepository) GetCategoryByID(id int) (*models.Category, error) {
	var cat models.Category
	query := `SELECT * FROM categories WHERE id = $1`
	err := r.db.Get(&cat, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

// Products
func (r *BotRepository) GetByCategoryID(categoryID int) ([]models.Product, error) {
	var products []models.Product
	query := `SELECT * FROM products WHERE category_id = $1 AND is_active = true ORDER BY id ASC`
	err := r.db.Select(&products, query, categoryID)
	return products, err
}

func (r *BotRepository) GetAllProducts() ([]models.Product, error) {
	var products []models.Product
	query := `SELECT * FROM products ORDER BY id ASC`
	err := r.db.Select(&products, query)
	return products, err
}

func (r *BotRepository) GetProductByID(id int) (*models.Product, error) {
	var prod models.Product
	query := `SELECT * FROM products WHERE id = $1`
	err := r.db.Get(&prod, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &prod, nil
}

func (r *BotRepository) CreateProduct(p *models.Product) error {
	query := `INSERT INTO products (category_id, name, description, price, image_url, is_active) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	return r.db.QueryRow(query, p.CategoryID, p.Name, p.Description, p.Price, p.ImageURL, p.IsActive).Scan(&p.ID)
}

func (r *BotRepository) UpdateProduct(p *models.Product) error {
	query := `UPDATE products SET category_id = $1, name = $2, description = $3, price = $4, image_url = $5, is_active = $6, updated_at = NOW() 
              WHERE id = $7`
	_, err := r.db.Exec(query, p.CategoryID, p.Name, p.Description, p.Price, p.ImageURL, p.IsActive, p.ID)
	return err
}

func (r *BotRepository) DeleteProduct(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Orders
func (r *BotRepository) CreateOrder(order *models.Order) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO orders (customer_id, total_price, status, address, phone, comment) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`
	err = tx.QueryRow(query, order.CustomerID, order.TotalPrice, order.Status, order.Address, order.Phone, order.Comment).
		Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return err
	}

	for i := range order.Items {
		item := &order.Items[i]
		_, err = tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity, price, comment) VALUES ($1, $2, $3, $4, $5)`,
			order.ID, item.ProductID, item.Quantity, item.Price, item.Comment)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BotRepository) GetOrderByID(id int) (*models.Order, error) {
	var order models.Order
	query := `
		SELECT o.*, u1.full_name as courier_name, u2.full_name as cook_name
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

	var items []models.OrderItem
	err = r.db.Select(&items, `SELECT oi.*, p.name as product_name FROM order_items oi LEFT JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`, id)
	if err == nil {
		order.Items = items
	}

	return &order, nil
}

func (r *BotRepository) GetAllOrders() ([]models.Order, error) {
	var orders []models.Order
	query := `SELECT o.*, u1.full_name as courier_name, u2.full_name as cook_name FROM orders o 
              LEFT JOIN users u1 ON o.courier_id = u1.id LEFT JOIN users u2 ON o.cook_id = u2.id ORDER BY o.created_at DESC`
	err := r.db.Select(&orders, query)
	if err != nil {
		return nil, err
	}
	
	for i := range orders {
		var items []models.OrderItem
		_ = r.db.Select(&items, `SELECT oi.*, p.name as product_name FROM order_items oi LEFT JOIN products p ON oi.product_id = p.id WHERE oi.order_id = $1`, orders[i].ID)
		orders[i].Items = items
	}
	return orders, nil
}

func (r *BotRepository) UpdateStatus(orderID int, status models.OrderStatus, cookID *int) error {
	if cookID != nil {
		_, err := r.db.Exec(`UPDATE orders SET status = $1, cook_id = $2, updated_at = NOW() WHERE id = $3`, status, *cookID, orderID)
		return err
	}
	_, err := r.db.Exec(`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`, status, orderID)
	return err
}

func (r *BotRepository) AddStaffRating(rating *models.StaffRating) error {
	query := `INSERT INTO staff_ratings (order_id, staff_id, staff_role, rating, comment) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, rating.OrderID, rating.StaffID, rating.StaffRole, rating.Rating, rating.Comment)
	return err
}

// Bot Staff Management (Persistence)
func (r *BotRepository) EnsureStaffTable() error {
	query := `CREATE TABLE IF NOT EXISTS bot_staff (
		telegram_id BIGINT PRIMARY KEY,
		role VARCHAR(20) NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
	)`
	_, err := r.db.Exec(query)
	return err
}

func (r *BotRepository) GetStaffByRole(role string) ([]int64, error) {
	var ids []int64
	query := `SELECT telegram_id FROM bot_staff WHERE role = $1`
	err := r.db.Select(&ids, query, role)
	return ids, err
}

func (r *BotRepository) AddStaff(telegramID int64, role string) error {
	query := `INSERT INTO bot_staff (telegram_id, role) VALUES ($1, $2) ON CONFLICT (telegram_id) DO UPDATE SET role = EXCLUDED.role`
	_, err := r.db.Exec(query, telegramID, role)
	return err
}

func (r *BotRepository) RemoveStaff(telegramID int64) error {
	query := `DELETE FROM bot_staff WHERE telegram_id = $1`
	_, err := r.db.Exec(query, telegramID)
	return err
}

func (r *BotRepository) EnsureCourierColumn() error {
	// Add courier_telegram_id column to orders if it doesn't exist
	query := `ALTER TABLE orders ADD COLUMN IF NOT EXISTS courier_telegram_id BIGINT`
	_, err := r.db.Exec(query)
	return err
}

func (r *BotRepository) AssignCourierByTelegramID(orderID int, telegramID int64) error {
	_ = r.EnsureCourierColumn() // Ensure column exists
	query := `UPDATE orders SET courier_telegram_id = $1, status = $2, updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(query, telegramID, models.StatusOnWay, orderID)
	return err
}

package repository

import (
	"database/sql"
	"errors"

	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(product *models.Product) error {
	query := `INSERT INTO products (category_id, name, description, price, image_url, is_active, unit, min_quantity, quantity_step, has_mandatory_container) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, product.CategoryID, product.Name, product.Description, product.Price, product.ImageURL, product.IsActive, product.Unit, product.MinQuantity, product.QuantityStep, product.HasMandatoryContainer).
		Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT p.id, p.category_id, p.name, COALESCE(p.description, '') as description, p.price, COALESCE(p.image_url, '') as image_url, p.is_active, COALESCE(p.unit, 'dona') as unit, COALESCE(p.min_quantity, 1) as min_quantity, COALESCE(p.quantity_step, 1) as quantity_step, COALESCE(p.has_mandatory_container, false) as has_mandatory_container, p.created_at, p.updated_at,
		COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0) as cost_price,
		(p.price - COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0)) as profit_margin
		FROM products p ORDER BY p.id ASC
	`
	err := r.db.Select(&products, query)
	return products, err
}

func (r *ProductRepository) GetByCategoryID(categoryID int) ([]models.Product, error) {
	var products []models.Product
	query := `
		SELECT p.id, p.category_id, p.name, COALESCE(p.description, '') as description, p.price, COALESCE(p.image_url, '') as image_url, p.is_active, COALESCE(p.unit, 'dona') as unit, COALESCE(p.min_quantity, 1) as min_quantity, COALESCE(p.quantity_step, 1) as quantity_step, COALESCE(p.has_mandatory_container, false) as has_mandatory_container, p.created_at, p.updated_at,
		COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0) as cost_price,
		(p.price - COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0)) as profit_margin
		FROM products p WHERE p.category_id = $1 AND p.is_active = true ORDER BY p.id ASC
	`
	err := r.db.Select(&products, query, categoryID)
	return products, err
}

func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT p.id, p.category_id, p.name, COALESCE(p.description, '') as description, p.price, COALESCE(p.image_url, '') as image_url, p.is_active, COALESCE(p.unit, 'dona') as unit, COALESCE(p.min_quantity, 1) as min_quantity, COALESCE(p.quantity_step, 1) as quantity_step, COALESCE(p.has_mandatory_container, false) as has_mandatory_container, p.created_at, p.updated_at,
		COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0) as cost_price,
		(p.price - COALESCE((
			SELECT SUM(
				CASE 
					WHEN pi.unit = i.unit THEN pi.quantity * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'gr' AND i.unit = 'kg' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'kg' AND i.unit = 'gr' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'ml' AND i.unit = 'l' THEN (pi.quantity / 1000.0) * COALESCE(i.cost_price, 0)
					WHEN pi.unit = 'l' AND i.unit = 'ml' THEN (pi.quantity * 1000.0) * COALESCE(i.cost_price, 0)
					ELSE pi.quantity * COALESCE(i.cost_price, 0)
				END
			) FROM product_ingredients pi JOIN ingredients i ON pi.ingredient_id = i.id WHERE pi.product_id = p.id
		), 0)) as profit_margin
		FROM products p WHERE p.id = $1
	`
	err := r.db.Get(&product, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) Update(product *models.Product) error {
	query := `UPDATE products SET category_id = $1, name = $2, description = $3, price = $4, image_url = $5, is_active = $6, 
              unit = $7, min_quantity = $8, quantity_step = $9, has_mandatory_container = $10, updated_at = NOW() 
              WHERE id = $11 RETURNING updated_at`
	return r.db.QueryRow(query, product.CategoryID, product.Name, product.Description, product.Price, product.ImageURL, product.IsActive, 
		product.Unit, product.MinQuantity, product.QuantityStep, product.HasMandatoryContainer, product.ID).
		Scan(&product.UpdatedAt)
}

func (r *ProductRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *ProductRepository) GetPrinterTarget(productID int) (string, error) {
	var target string
	query := `SELECT COALESCE(c.printer_target, 'ALL') FROM categories c JOIN products p ON c.id = p.category_id WHERE p.id = $1`
	err := r.db.Get(&target, query, productID)
	if err != nil {
		return "ALL", err
	}
	return target, nil
}


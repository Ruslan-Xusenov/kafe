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
	query := `INSERT INTO products (category_id, name, description, price, image_url, is_active) 
              VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, product.CategoryID, product.Name, product.Description, product.Price, product.ImageURL, product.IsActive).
		Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
}

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product
	query := `SELECT * FROM products ORDER BY id ASC`
	err := r.db.Select(&products, query)
	return products, err
}

func (r *ProductRepository) GetByCategoryID(categoryID int) ([]models.Product, error) {
	var products []models.Product
	query := `SELECT * FROM products WHERE category_id = $1 AND is_active = true ORDER BY id ASC`
	err := r.db.Select(&products, query, categoryID)
	return products, err
}

func (r *ProductRepository) GetByID(id int) (*models.Product, error) {
	var product models.Product
	query := `SELECT * FROM products WHERE id = $1`
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
	query := `UPDATE products SET category_id = $1, name = $2, description = $3, price = $4, image_url = $5, is_active = $6, updated_at = NOW() 
              WHERE id = $7 RETURNING updated_at`
	return r.db.QueryRow(query, product.CategoryID, product.Name, product.Description, product.Price, product.ImageURL, product.IsActive, product.ID).
		Scan(&product.UpdatedAt)
}

func (r *ProductRepository) Delete(id int) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

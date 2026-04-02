package repository

import (
	"database/sql"
	"errors"

	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type CategoryRepository struct {
	db *sqlx.DB
}

func NewCategoryRepository(db *sqlx.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(category *models.Category) error {
	query := `INSERT INTO categories (name, image_url) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, category.Name, category.ImageURL).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

func (r *CategoryRepository) GetAll() ([]models.Category, error) {
	var categories []models.Category
	query := `SELECT * FROM categories ORDER BY id ASC`
	err := r.db.Select(&categories, query)
	return categories, err
}

func (r *CategoryRepository) GetByID(id int) (*models.Category, error) {
	var category models.Category
	query := `SELECT * FROM categories WHERE id = $1`
	err := r.db.Get(&category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *CategoryRepository) Update(category *models.Category) error {
	query := `UPDATE categories SET name = $1, image_url = $2, updated_at = NOW() WHERE id = $3 RETURNING updated_at`
	return r.db.QueryRow(query, category.Name, category.ImageURL, category.ID).Scan(&category.UpdatedAt)
}

func (r *CategoryRepository) Delete(id int) error {
	query := `DELETE FROM categories WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

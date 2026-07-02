package repository

import (
	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type InventoryRepository interface {
	// Ingredients
	CreateIngredient(ing *models.Ingredient) error
	GetIngredients() ([]models.Ingredient, error)
	UpdateIngredient(ing *models.Ingredient) error
	DeleteIngredient(id int) error

	// Product Ingredients (Recipes)
	AddProductIngredient(pi *models.ProductIngredient) error
	GetProductIngredients(productID int) ([]models.ProductIngredient, error)
	DeleteProductIngredient(id int) error

	// Stock Operations
	DeductStockForProduct(productID int, productQuantity float64) error
	RestoreStockForProduct(productID int, productQuantity float64) error
}

type inventoryRepository struct {
	db *sqlx.DB
}

func NewInventoryRepository(db *sqlx.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) CreateIngredient(ing *models.Ingredient) error {
	query := `INSERT INTO ingredients (name, stock, unit, min_stock) VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.QueryRow(query, ing.Name, ing.Stock, ing.Unit, ing.MinStock).Scan(&ing.ID, &ing.CreatedAt, &ing.UpdatedAt)
}

func (r *inventoryRepository) GetIngredients() ([]models.Ingredient, error) {
	query := `SELECT id, name, stock, unit, min_stock, created_at, updated_at FROM ingredients ORDER BY name ASC`
	var ingredients []models.Ingredient
	err := r.db.Select(&ingredients, query)
	return ingredients, err
}

func (r *inventoryRepository) UpdateIngredient(ing *models.Ingredient) error {
	query := `UPDATE ingredients SET name=$1, stock=$2, unit=$3, min_stock=$4, updated_at=NOW() WHERE id=$5`
	_, err := r.db.Exec(query, ing.Name, ing.Stock, ing.Unit, ing.MinStock, ing.ID)
	return err
}

func (r *inventoryRepository) DeleteIngredient(id int) error {
	query := `DELETE FROM ingredients WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *inventoryRepository) AddProductIngredient(pi *models.ProductIngredient) error {
	query := `INSERT INTO product_ingredients (product_id, ingredient_id, quantity, unit) VALUES ($1, $2, $3, $4) RETURNING id, created_at`
	return r.db.QueryRow(query, pi.ProductID, pi.IngredientID, pi.Quantity, pi.Unit).Scan(&pi.ID, &pi.CreatedAt)
}

func (r *inventoryRepository) GetProductIngredients(productID int) ([]models.ProductIngredient, error) {
	query := `
		SELECT pi.id, pi.product_id, pi.ingredient_id, i.name as ingredient_name, pi.quantity, pi.unit, pi.created_at
		FROM product_ingredients pi
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE pi.product_id = $1
	`
	var pis []models.ProductIngredient
	err := r.db.Select(&pis, query, productID)
	return pis, err
}

func (r *inventoryRepository) DeleteProductIngredient(id int) error {
	query := `DELETE FROM product_ingredients WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *inventoryRepository) DeductStockForProduct(productID int, productQuantity float64) error {
	// Reduce stock for each ingredient in the product's recipe
	query := `
		UPDATE ingredients
		SET stock = stock - (pi.quantity * $2)
		FROM product_ingredients pi
		WHERE ingredients.id = pi.ingredient_id AND pi.product_id = $1
	`
	_, err := r.db.Exec(query, productID, productQuantity)
	return err
}

func (r *inventoryRepository) RestoreStockForProduct(productID int, productQuantity float64) error {
	// Restore stock for each ingredient in the product's recipe
	query := `
		UPDATE ingredients
		SET stock = stock + (pi.quantity * $2)
		FROM product_ingredients pi
		WHERE ingredients.id = pi.ingredient_id AND pi.product_id = $1
	`
	_, err := r.db.Exec(query, productID, productQuantity)
	return err
}

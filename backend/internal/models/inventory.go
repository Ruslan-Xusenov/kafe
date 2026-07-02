package models

import "time"

type Ingredient struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Stock     float64   `json:"stock" db:"stock"`
	Unit      string    `json:"unit" db:"unit"`
	MinStock  float64   `json:"min_stock" db:"min_stock"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type ProductIngredient struct {
	ID             int       `json:"id" db:"id"`
	ProductID      int       `json:"product_id" db:"product_id"`
	IngredientID   int       `json:"ingredient_id" db:"ingredient_id"`
	IngredientName string    `json:"ingredient_name,omitempty" db:"ingredient_name"` // For joining
	Quantity       float64   `json:"quantity" db:"quantity"`
	Unit           string    `json:"unit" db:"unit"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
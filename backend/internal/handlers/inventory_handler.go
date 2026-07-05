package handlers

import (
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	repo repository.InventoryRepository
}

func NewInventoryHandler(repo repository.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

// Ingredients
func (h *InventoryHandler) GetIngredients(c *gin.Context) {
	ingredients, err := h.repo.GetIngredients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get ingredients"})
		return
	}
	if ingredients == nil {
		ingredients = []models.Ingredient{}
	}
	c.JSON(http.StatusOK, ingredients)
}

func (h *InventoryHandler) CreateIngredient(c *gin.Context) {
	var input models.Ingredient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.repo.CreateIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ingredient"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *InventoryHandler) UpdateIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var input models.Ingredient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	input.ID = id
	if err := h.repo.UpdateIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ingredient"})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *InventoryHandler) DeleteIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.repo.DeleteIngredient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ingredient"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

// Product Ingredients (Recipes)
func (h *InventoryHandler) GetProductIngredients(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}
	pis, err := h.repo.GetProductIngredients(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get product ingredients"})
		return
	}
	if pis == nil {
		pis = []models.ProductIngredient{}
	}
	c.JSON(http.StatusOK, pis)
}

func (h *InventoryHandler) AddProductIngredient(c *gin.Context) {
	var input models.ProductIngredient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := h.repo.AddProductIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product ingredient"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *InventoryHandler) DeleteProductIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := h.repo.DeleteProductIngredient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product ingredient"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
}

func (h *InventoryHandler) RestockIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}
	if err := h.repo.RestockIngredient(id, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restock ingredient"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Restocked successfully"})
}

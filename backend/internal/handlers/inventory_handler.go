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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении ингредиентов"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := h.repo.CreateIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при создании ингредиента"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *InventoryHandler) UpdateIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var input models.Ingredient
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	input.ID = id
	if err := h.repo.UpdateIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении ингредиента"})
		return
	}
	c.JSON(http.StatusOK, input)
}

func (h *InventoryHandler) DeleteIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	if err := h.repo.DeleteIngredient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении ингредиента"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешно удалено"})
}

// Product Ingredients (Recipes)
func (h *InventoryHandler) GetProductIngredients(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID продукта"})
		return
	}
	pis, err := h.repo.GetProductIngredients(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при получении ингредиентов продукта"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	if err := h.repo.AddProductIngredient(&input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при добавлении ингредиента продукта"})
		return
	}
	c.JSON(http.StatusCreated, input)
}

func (h *InventoryHandler) DeleteProductIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	if err := h.repo.DeleteProductIngredient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении ингредиента продукта"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешно удалено"})
}

func (h *InventoryHandler) RestockIngredient(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID"})
		return
	}
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество должно быть положительным"})
		return
	}
	if err := h.repo.RestockIngredient(id, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при пополнении ингредиента"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Успешно пополнено"})
}

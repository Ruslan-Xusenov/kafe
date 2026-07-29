package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
	"github.com/username/kafe-backend/internal/service"
)

type CatalogHandler struct {
	service   *service.CatalogService
	auditRepo *repository.AuditRepository
}

func NewCatalogHandler(s *service.CatalogService, auditRepo *repository.AuditRepository) *CatalogHandler {
	return &CatalogHandler{service: s, auditRepo: auditRepo}
}

// Category Handlers
func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "category.create", "category", &cat.ID, gin.H{
		"name":           cat.Name,
		"printer_target": cat.PrinterTarget,
	})

	c.JSON(http.StatusCreated, cat)
}

func (h *CatalogHandler) GetAllCategories(c *gin.Context) {
	cats, err := h.service.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *CatalogHandler) UpdateCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var cat models.Category
	if err := c.ShouldBindJSON(&cat); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat.ID = id

	if err := h.service.UpdateCategory(&cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "category.update", "category", &cat.ID, gin.H{
		"name":           cat.Name,
		"printer_target": cat.PrinterTarget,
	})

	c.JSON(http.StatusOK, cat)
}

func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteCategory(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(c, h.auditRepo, "category.delete", "category", &id, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Категория удалена"})
}

// Product Handlers
func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	var prod models.Product
	if err := c.ShouldBindJSON(&prod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateProduct(&prod); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "product.create", "product", &prod.ID, gin.H{
		"name":        prod.Name,
		"category_id": prod.CategoryID,
		"price":       prod.Price,
	})

	c.JSON(http.StatusCreated, prod)
}

func (h *CatalogHandler) GetAllProducts(c *gin.Context) {
	prods, err := h.service.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prods)
}

func (h *CatalogHandler) GetProductsByCategory(c *gin.Context) {
	catID, _ := strconv.Atoi(c.Param("cat_id"))
	prods, err := h.service.GetProductsByCategory(catID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prods)
}

func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var prod models.Product
	if err := c.ShouldBindJSON(&prod); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prod.ID = id

	if err := h.service.UpdateProduct(&prod); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	writeAudit(c, h.auditRepo, "product.update", "product", &prod.ID, gin.H{
		"name":        prod.Name,
		"category_id": prod.CategoryID,
		"price":       prod.Price,
		"is_active":   prod.IsActive,
	})

	c.JSON(http.StatusOK, prod)
}

func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteProduct(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	writeAudit(c, h.auditRepo, "product.delete", "product", &id, nil)
	c.JSON(http.StatusOK, gin.H{"message": "Продукт удален"})
}

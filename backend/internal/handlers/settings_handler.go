package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/username/kafe-backend/internal/repository"
)

type SettingsHandler struct {
	repo *repository.SettingsRepository
}

func NewSettingsHandler(repo *repository.SettingsRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

func (h *SettingsHandler) GetSettings(c *gin.Context) {
	containerPrice, err := h.repo.Get("container_price")
	if err != nil {
		containerPrice = "1000" // Fallback
	}
	
	containerID, err := h.repo.Get("container_product_id")
	if err != nil {
		containerID = "7" // Fallback
	}
	
	tableServicePercentage, err := h.repo.Get("table_service_percentage")
	if err != nil {
		tableServicePercentage = "10" // Fallback
	}
	
	c.JSON(http.StatusOK, gin.H{
		"container_price":      containerPrice,
		"container_product_id": containerID,
		"table_service_percentage": tableServicePercentage,
	})
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var body struct {
		ContainerPrice         string `json:"container_price"`
		ContainerProductID     string `json:"container_product_id"`
		TableServicePercentage string `json:"table_service_percentage"`
	}
	
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}
	
	if body.ContainerPrice != "" {
		if err := h.repo.Set("container_price", body.ContainerPrice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении цены"})
			return
		}
	}
	
	if body.ContainerProductID != "" {
		if err := h.repo.Set("container_product_id", body.ContainerProductID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении ID"})
			return
		}
	}
	
	if body.TableServicePercentage != "" {
		if err := h.repo.Set("table_service_percentage", body.TableServicePercentage); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при обновлении процента обслуживания"})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"message": "Настройки обновлены"})
}

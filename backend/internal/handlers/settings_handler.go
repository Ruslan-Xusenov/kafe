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
	
	c.JSON(http.StatusOK, gin.H{
		"container_price":      containerPrice,
		"container_product_id": containerID,
	})
}

func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	var body struct {
		ContainerPrice     string `json:"container_price"`
		ContainerProductID string `json:"container_product_id"`
	}
	
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}
	
	if body.ContainerPrice != "" {
		if err := h.repo.Set("container_price", body.ContainerPrice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update price"})
			return
		}
	}
	
	if body.ContainerProductID != "" {
		if err := h.repo.Set("container_product_id", body.ContainerProductID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ID"})
			return
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated"})
}

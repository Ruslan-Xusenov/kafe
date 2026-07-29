package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не предоставлен"})
		return
	}

	const maxImageSize = 5 * 1024 * 1024
	if file.Size <= 0 || file.Size > maxImageSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Rasm hajmi 5 MB dan oshmasligi kerak"})
		return
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Faylni o'qib bo'lmadi"})
		return
	}
	header, readErr := io.ReadAll(io.LimitReader(opened, 512))
	opened.Close()
	if readErr != nil || len(header) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fayl formati noto'g'ri"})
		return
	}

	contentType := http.DetectContentType(header)
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}
	ext, ok := extensions[contentType]
	if !ok {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Faqat JPG, PNG yoki WEBP rasmlar qabul qilinadi"})
		return
	}

	// Ensure uploads directory exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать директорию для загрузки"})
		return
	}

	newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join(uploadDir, newFilename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить файл"})
		return
	}

	// Return the relative URL
	c.JSON(http.StatusOK, gin.H{
		"url": fmt.Sprintf("/%s", dst),
	})
}

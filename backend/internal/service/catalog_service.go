package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type CatalogService struct {
	categoryRepo *repository.CategoryRepository
	productRepo  *repository.ProductRepository
}

func NewCatalogService(catRepo *repository.CategoryRepository, prodRepo *repository.ProductRepository) *CatalogService {
	return &CatalogService{
		categoryRepo: catRepo,
		productRepo:  prodRepo,
	}
}

// Category Methods
func (s *CatalogService) CreateCategory(cat *models.Category) error {
	return s.categoryRepo.Create(cat)
}

func (s *CatalogService) GetAllCategories() ([]models.Category, error) {
	return s.categoryRepo.GetAll()
}

func (s *CatalogService) GetCategoryByID(id int) (*models.Category, error) {
	return s.categoryRepo.GetByID(id)
}

func (s *CatalogService) UpdateCategory(cat *models.Category) error {
	return s.categoryRepo.Update(cat)
}

// Helper to delete physical image file
func (s *CatalogService) deleteImageFile(imageURL string) {
	if imageURL == "" {
		return
	}

	// Assuming imageURL is like "/uploads/filename.png" or "https://.../uploads/filename.png"
	// We need the part after "/uploads/"
	var filename string
	if parts := strings.Split(imageURL, "/uploads/"); len(parts) > 1 {
		filename = parts[1]
	} else if strings.HasPrefix(imageURL, "uploads/") {
		filename = strings.TrimPrefix(imageURL, "uploads/")
	} else {
		// Just filename?
		filename = imageURL
	}
	
	if filename != "" {
		path := filepath.Join("uploads", filename)
		_ = os.Remove(path) // Ignore error if file doesn't exist
	}
}

func (s *CatalogService) DeleteCategory(id int) error {
	// 1. Get Category info to get its image
	cat, err := s.categoryRepo.GetByID(id)
	if err == nil && cat != nil && cat.ImageURL != nil {
		s.deleteImageFile(*cat.ImageURL)
	}

	// 2. Get all products in this category to delete their images
	prods, err := s.productRepo.GetByCategoryID(id)
	if err == nil {
		for _, p := range prods {
			s.deleteImageFile(p.ImageURL)
		}
	}

	// 3. Delete from DB (cascade handles record removal)
	return s.categoryRepo.Delete(id)
}

// Product Methods
func (s *CatalogService) CreateProduct(prod *models.Product) error {
	return s.productRepo.Create(prod)
}

func (s *CatalogService) GetAllProducts() ([]models.Product, error) {
	return s.productRepo.GetAll()
}

func (s *CatalogService) GetProductsByCategory(categoryID int) ([]models.Product, error) {
	return s.productRepo.GetByCategoryID(categoryID)
}

func (s *CatalogService) GetProductByID(id int) (*models.Product, error) {
	return s.productRepo.GetByID(id)
}

func (s *CatalogService) UpdateProduct(prod *models.Product) error {
	return s.productRepo.Update(prod)
}

func (s *CatalogService) DeleteProduct(id int) error {
	// 1. Get product info to get its image
	prod, err := s.productRepo.GetByID(id)
	if err == nil && prod != nil {
		s.deleteImageFile(prod.ImageURL)
	}
	
	// 2. Delete from DB
	return s.productRepo.Delete(id)
}

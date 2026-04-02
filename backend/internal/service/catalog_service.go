package service

import (
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

func (s *CatalogService) DeleteCategory(id int) error {
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
	return s.productRepo.Delete(id)
}

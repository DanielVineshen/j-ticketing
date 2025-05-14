package services

import (
	"j-ticketing/internal/db/repositories"
)

// ProductService handles product-related business logic
type ProductService struct {
	repo *repositories.ProductRepository
}

// NewProductService creates a new ProductService
func NewProductService(repo *repositories.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

// GetProductByID gets a product by ID
func (s *ProductService) GetProductByID(id int64) (*repositories.Product, error) {
	return s.repo.FindByID(id)
}

// GetAllProducts gets all products
func (s *ProductService) GetAllProducts() ([]*repositories.Product, error) {
	return s.repo.FindAll()
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *repositories.Product) error {
	// Add any business logic/validation here
	return s.repo.Create(product)
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(product *repositories.Product) error {
	// Add any business logic/validation here
	return s.repo.Update(product)
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(id int64) error {
	return s.repo.Delete(id)
}

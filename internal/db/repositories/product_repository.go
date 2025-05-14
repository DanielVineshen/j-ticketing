package repositories

import (
	"database/sql"
	"fmt"
)

// Product represents a product entity
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       float64
	// Add other fields as needed
}

// ProductRepository handles product data access
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

// FindByID finds a product by ID
func (r *ProductRepository) FindByID(id int64) (*Product, error) {
	product := &Product{}

	query := "SELECT id, name, description, price FROM products WHERE id = ?"
	err := r.db.QueryRow(query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Price)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Product not found
		}
		return nil, fmt.Errorf("error finding product by ID: %w", err)
	}

	return product, nil
}

// FindAll finds all products
func (r *ProductRepository) FindAll() ([]*Product, error) {
	query := "SELECT id, name, description, price FROM products"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error finding all products: %w", err)
	}
	defer rows.Close()

	var products []*Product
	for rows.Next() {
		product := &Product{}
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.Price); err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// Create creates a new product
func (r *ProductRepository) Create(product *Product) error {
	query := "INSERT INTO products (name, description, price) VALUES (?, ?, ?)"
	result, err := r.db.Exec(query, product.Name, product.Description, product.Price)
	if err != nil {
		return fmt.Errorf("error creating product: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting inserted ID: %w", err)
	}

	product.ID = id
	return nil
}

// Update updates a product
func (r *ProductRepository) Update(product *Product) error {
	query := "UPDATE products SET name = ?, description = ?, price = ? WHERE id = ?"
	_, err := r.db.Exec(query, product.Name, product.Description, product.Price, product.ID)
	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	return nil
}

// Delete deletes a product
func (r *ProductRepository) Delete(id int64) error {
	query := "DELETE FROM products WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

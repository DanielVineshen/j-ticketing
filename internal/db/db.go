package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // MariaDB/MySQL driver

	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/config"
)

// Database represents a database connection
type Database struct {
	db *sql.DB
	// Repositories
	userRepo    *repositories.UserRepository
	productRepo *repositories.ProductRepository
	// Add other repositories as needed
}

// Connect establishes a connection to the MariaDB database
func Connect(dbConfig config.DatabaseConfig) (*Database, error) {
	db, err := sql.Open("mysql", dbConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool parameters
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to MariaDB database")

	// Initialize repositories
	database := &Database{
		db: db,
	}

	database.initRepositories()

	return database, nil
}

// initRepositories initializes all repositories
func (d *Database) initRepositories() {
	d.userRepo = repositories.NewUserRepository(d.db)
	d.productRepo = repositories.NewProductRepository(d.db)
	// Initialize other repositories
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// UserRepository returns the user repository
func (d *Database) UserRepository() *repositories.UserRepository {
	return d.userRepo
}

// ProductRepository returns the product repository
func (d *Database) ProductRepository() *repositories.ProductRepository {
	return d.productRepo
}

// DB returns the underlying database connection
func (d *Database) DB() *sql.DB {
	return d.db
}

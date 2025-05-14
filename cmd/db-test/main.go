package main

import (
	"fmt"
	"log"

	"j-ticketing/internal/db"
	"j-ticketing/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Try to connect to the database
	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// If we get here, connection was successful
	fmt.Println("Successfully connected to MariaDB!")

	// Test a simple query
	var version string
	err = database.DB().QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to query MariaDB version: %v", err)
	}

	fmt.Printf("Connected to MariaDB version: %s\n", version)
}

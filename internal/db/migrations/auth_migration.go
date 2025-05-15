package migrations

import (
	"database/sql"
	"fmt"
)

// AddPasswordColumnToAdmin adds the password column to the Admin table
func AddPasswordColumnToAdmin(db *sql.DB) error {
	// Check if the column already exists
	var columnExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM information_schema.COLUMNS
		WHERE TABLE_NAME = 'Admin'
		AND COLUMN_NAME = 'password'
	`).Scan(&columnExists)

	if err != nil {
		return fmt.Errorf("error checking if password column exists: %w", err)
	}

	// If the column doesn't exist, add it
	if !columnExists {
		_, err = db.Exec(`
			ALTER TABLE Admin
			ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT '';
		`)
		if err != nil {
			return fmt.Errorf("error adding password column to Admin table: %w", err)
		}
		fmt.Println("Added password column to Admin table")
	}

	return nil
}

// CreateDefaultAdminUser creates a default admin user
func CreateDefaultAdminUser(db *sql.DB) error {
	// Check if the admin user already exists
	var userExists bool
	err := db.QueryRow(`
		SELECT COUNT(*) > 0
		FROM Admin
		WHERE username = 'admin'
	`).Scan(&userExists)

	if err != nil {
		return fmt.Errorf("error checking if admin user exists: %w", err)
	}

	// If the admin user doesn't exist, create it
	if !userExists {
		// Create the admin user with a default password (hashed)
		// In a production environment, you would use bcrypt to hash the password
		// For this example, we'll use a placeholder hash
		_, err = db.Exec(`
			INSERT INTO Admin (username, password, full_name, role, created_at, updated_at)
			VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'System Administrator', 'SYSADMIN', NOW(), NOW());
		`)
		if err != nil {
			return fmt.Errorf("error creating default admin user: %w", err)
		}
		fmt.Println("Created default admin user with username 'admin' and password 'password'")
	}

	return nil
}

// These functions should be called in your main migration function
func RunAuthMigrations(db *sql.DB) error {
	if err := AddPasswordColumnToAdmin(db); err != nil {
		return err
	}

	if err := CreateDefaultAdminUser(db); err != nil {
		return err
	}

	return nil
}

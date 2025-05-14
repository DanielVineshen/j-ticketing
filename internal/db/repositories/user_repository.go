package repositories

import (
	"database/sql"
	"fmt"
)

// User represents a user entity
type User struct {
	ID       int64
	Username string
	Email    string
	// Add other fields as needed
}

// UserRepository handles user data access
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int64) (*User, error) {
	user := &User{}

	query := "SELECT id, username, email FROM users WHERE id = ?"
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Username, &user.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("error finding user by ID: %w", err)
	}

	return user, nil
}

// FindAll finds all users
func (r *UserRepository) FindAll() ([]*User, error) {
	query := "SELECT id, username, email FROM users"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error finding all users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user := &User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// Create creates a new user
func (r *UserRepository) Create(user *User) error {
	query := "INSERT INTO users (username, email) VALUES (?, ?)"
	result, err := r.db.Exec(query, user.Username, user.Email)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting inserted ID: %w", err)
	}

	user.ID = id
	return nil
}

// Update updates a user
func (r *UserRepository) Update(user *User) error {
	query := "UPDATE users SET username = ?, email = ? WHERE id = ?"
	_, err := r.db.Exec(query, user.Username, user.Email, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(id int64) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

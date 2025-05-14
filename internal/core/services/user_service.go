package services

import (
	"j-ticketing/internal/db/repositories"
)

// UserService handles user-related business logic
type UserService struct {
	repo *repositories.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(id int64) (*repositories.User, error) {
	return s.repo.FindByID(id)
}

// GetAllUsers gets all users
func (s *UserService) GetAllUsers() ([]*repositories.User, error) {
	return s.repo.FindAll()
}

// CreateUser creates a new user
func (s *UserService) CreateUser(user *repositories.User) error {
	// Add any business logic/validation here
	return s.repo.Create(user)
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(user *repositories.User) error {
	// Add any business logic/validation here
	return s.repo.Update(user)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(id int64) error {
	return s.repo.Delete(id)
}

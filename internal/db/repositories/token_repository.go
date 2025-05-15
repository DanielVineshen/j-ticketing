// FILE: internal/repositories/token_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TokenRepository is the interface for token database operations
type TokenRepository interface {
	Create(token *models.Token) error
	FindByUserIdAndRefreshToken(userID, refreshToken string) (*models.Token, error)
	DeleteByUserIdAndRefreshToken(userID, refreshToken string) error
}

type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

// Create creates a new token in the database
func (r *tokenRepository) Create(token *models.Token) error {
	return r.db.Create(token).Error
}

// FindByUserIdAndRefreshToken finds a token by user ID and refresh token
func (r *tokenRepository) FindByUserIdAndRefreshToken(userID, refreshToken string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteByUserIdAndRefreshToken deletes a token by user ID and refresh token
func (r *tokenRepository) DeleteByUserIdAndRefreshToken(userID, refreshToken string) error {
	return r.db.Where("user_id = ? AND refresh_token = ?", userID, refreshToken).Delete(&models.Token{}).Error
}

// FILE: internal/repositories/token_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TokenRepository is the interface for token database operations
type TokenRepository interface {
	Create(token *models.Token) error
	FindByUserIdAndRefreshToken(userID, access string) (*models.Token, error)
	FindByUserIdAndAccessToken(userID, access string) (*models.Token, error)
	DeleteByUserIdAndAccessToken(userID, accessToken string) error
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

// FindByUserIdAndAccessToken finds a token by user ID and access token
func (r *tokenRepository) FindByUserIdAndAccessToken(userID, accessToken string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("user_id = ? AND access_token = ?", userID, accessToken).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteByUserIdAndAccessToken deletes a token by user ID and access token
func (r *tokenRepository) DeleteByUserIdAndAccessToken(userID, accessToken string) error {
	return r.db.Where("user_id = ? AND access_token = ?", userID, accessToken).Delete(&models.Token{}).Error
}

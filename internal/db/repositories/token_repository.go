// File: j-ticketing/internal/db/repositories/token_repository.go
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
	CountByUserId(userID string) (int64, error)
	FindOldestByUserId(userID string) (*models.Token, error)
	DeleteToken(token *models.Token) error
	UpdateToken(token *models.Token) error
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

// CountByUserId counts the number of tokens for a user
func (r *tokenRepository) CountByUserId(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Token{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// FindOldestByUserId finds the oldest token for a user
func (r *tokenRepository) FindOldestByUserId(userID string) (*models.Token, error) {
	var token models.Token
	err := r.db.Where("user_id = ?", userID).Order("created_at ASC").First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// DeleteToken deletes a token
func (r *tokenRepository) DeleteToken(token *models.Token) error {
	return r.db.Delete(token).Error
}

// UpdateToken updates a token in the database
func (r *tokenRepository) UpdateToken(token *models.Token) error {
	return r.db.Save(token).Error
}

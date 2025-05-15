// FILE: internal/auth/service/auth_service.go
package service

import (
	"errors"
	"j-ticketing/internal/auth/jwt"
	"j-ticketing/internal/auth/models"
	"j-ticketing/internal/db/repositories"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService is the interface for authentication operations
type AuthService interface {
	LoginAdmin(username, password string) (*models.TokenResponse, error)
	LoginCustomer(email, password string) (*models.TokenResponse, error)
	ValidateToken(token string) (*models.UserClaims, error)
	RefreshToken(refreshToken string) (*models.TokenResponse, error)
	SaveToken(userID, userType, accessToken, refreshToken, ipAddress, userAgent string) error
	RevokeToken(userID, refreshToken string) error
}

type authService struct {
	jwtService      jwt.JWTService
	adminRepo       repositories.AdminRepository
	customerRepo    repositories.CustomerRepository
	tokenRepo       repositories.TokenRepository
	accessTokenTTL  int64
	refreshTokenTTL int64
}

// NewAuthService creates a new authentication service
func NewAuthService(
	jwtService jwt.JWTService,
	adminRepo repositories.AdminRepository,
	customerRepo repositories.CustomerRepository,
	tokenRepo repositories.TokenRepository,
	accessTokenTTL, refreshTokenTTL int64,
) AuthService {
	return &authService{
		jwtService:      jwtService,
		adminRepo:       adminRepo,
		customerRepo:    customerRepo,
		tokenRepo:       tokenRepo,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// LoginAdmin handles admin login authentication
func (s *authService) LoginAdmin(username, password string) (*models.TokenResponse, error) {
	// Find admin by username
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Validate password (assuming Admin model has a Password field)
	// In a real application, you should use bcrypt to compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Create user claims
	userClaims := &models.UserClaims{
		UserID:   string(admin.AdminId), // Convert uint to string
		Username: admin.Username,
		UserType: "admin",
		Role:     admin.Role,
		Roles:    []string{admin.Role}, // If you need multiple roles
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateToken(userClaims, false)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateToken(userClaims, true)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenTTL,
	}, nil
}

// LoginCustomer handles customer login authentication
func (s *authService) LoginCustomer(email, password string) (*models.TokenResponse, error) {
	// Find customer by email
	customer, err := s.customerRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if customer is disabled
	if customer.IsDisabled {
		return nil, errors.New("account is disabled")
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Create user claims
	userClaims := &models.UserClaims{
		UserID:   customer.CustId,
		Username: customer.Email, // Using email as username for customers
		UserType: "customer",
		Role:     "CUSTOMER",
		Roles:    []string{"CUSTOMER"},
	}

	// Generate tokens
	accessToken, err := s.jwtService.GenerateToken(userClaims, false)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtService.GenerateToken(userClaims, true)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenTTL,
	}, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(token string) (*models.UserClaims, error) {
	return s.jwtService.ValidateToken(token)
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(refreshToken string) (*models.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if refresh token exists in database
	_, err = s.tokenRepo.FindByUserIdAndRefreshToken(claims.UserID, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateToken(claims, false)
	if err != nil {
		return nil, err
	}

	// Return new tokens
	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Keep the same refresh token
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenTTL,
	}, nil
}

// SaveToken saves a token to the database
func (s *authService) SaveToken(userID, userType, accessToken, refreshToken, ipAddress, userAgent string) error {
	token := &coremodels.Token{
		UserId:       userID,
		UserType:     userType,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IpAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.tokenRepo.Create(token)
}

// RevokeToken revokes a token (for logout)
func (s *authService) RevokeToken(userID, refreshToken string) error {
	return s.tokenRepo.DeleteByUserIdAndRefreshToken(userID, refreshToken)
}

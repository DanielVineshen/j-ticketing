// FILE: internal/auth/service/auth_service.go (fixed import)
package service

import (
	"database/sql"
	"errors"
	dto "j-ticketing/internal/core/dto/auth"
	coremodels "j-ticketing/internal/db/models" // Add this import for core models
	"j-ticketing/internal/db/repositories"
	jwt "j-ticketing/pkg/jwt"
	"log"
	"strconv" // Add this import for string conversion
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService is the interface for authentication operations
type AuthService interface {
	LoginAdmin(username, password string) (*dto.TokenResponse, error)
	LoginCustomer(email, password string) (*dto.TokenResponse, error)
	ValidateToken(token string) (bool, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
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
func (s *authService) LoginAdmin(username, password string) (*dto.TokenResponse, error) {
	// Find admin by username
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials -> name")
	}

	// Validate password (assuming Admin model has a Password field)
	// In a real application, you should use bcrypt to compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials -> password")
	}

	// Create user claims
	userClaims := &dto.UserClaims{
		UserID:   strconv.FormatUint(uint64(admin.AdminId), 10), // Correctly convert uint to string
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

	// Create user info for response
	userInfo := dto.UserInfo{
		AdminID:  admin.AdminId,
		Username: admin.Username,
		Role:     admin.Role,
		FullName: admin.FullName,
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userInfo,
	}, nil
}

// LoginCustomer handles customer login authentication
func (s *authService) LoginCustomer(email, password string) (*dto.TokenResponse, error) {
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
	if err := bcrypt.CompareHashAndPassword([]byte(customer.Password.String), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Create user claims
	userClaims := &dto.UserClaims{
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

	// Create user info for response
	userInfo := dto.UserInfo{
		CustID:           customer.CustId,
		Email:            customer.Email,
		IdentificationNo: customer.IdentificationNo,
		IsDisabled:       customer.IsDisabled,
		ContactNo:        customer.ContactNo.String,
		Role:             userClaims.UserType,
		FullName:         customer.FullName,
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userInfo,
	}, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(token string) (bool, error) {
	claims, err := s.jwtService.ValidateToken(token)

	// Get the identifier to use for token lookup
	var lookupID string = claims.Username
	// Check if refresh token exists in database
	_, err = s.tokenRepo.FindByUserIdAndAccessToken(lookupID, token)
	if err != nil {
		return false, nil
	}

	// return s.jwtService.ValidateToken(token)
	return true, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	// Validate refresh token
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get the identifier to use for token lookup
	var lookupID string = claims.Username

	// Check if refresh token exists in database
	_, err = s.tokenRepo.FindByUserIdAndRefreshToken(lookupID, refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateToken(claims, false)
	if err != nil {
		return nil, err
	}

	if claims.UserType == "admin" {
		// Find admin by username
		admin, err := s.adminRepo.FindByUsername(claims.Username)
		if err != nil {
			return nil, errors.New("invalid credentials -> name")
		}

		// Create user info for response
		userInfo := dto.UserInfo{
			AdminID:  admin.AdminId,
			Username: admin.Username,
			Role:     admin.Role,
			FullName: admin.FullName,
		}

		return &dto.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         userInfo,
		}, nil
	} else {
		// Find customer by email
		customer, err := s.customerRepo.FindByEmail(claims.Username)
		if err != nil {
			return nil, errors.New("invalid credentials")
		}

		// Check if customer is disabled
		if customer.IsDisabled {
			return nil, errors.New("account is disabled")
		}

		// Create user info for response
		userInfo := dto.UserInfo{
			CustID:           customer.CustId,
			Email:            customer.Email,
			IdentificationNo: customer.IdentificationNo,
			IsDisabled:       customer.IsDisabled,
			ContactNo:        customer.ContactNo.String,
			Role:             claims.UserType,
			FullName:         customer.FullName,
		}

		return &dto.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			User:         userInfo,
		}, nil
	}
}

// SaveToken saves a token to the database
func (s *authService) SaveToken(userID, userType, accessToken, refreshToken, ipAddress, userAgent string) error {
	token := &coremodels.Token{
		UserId:       userID,
		UserType:     userType,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IpAddress:    sql.NullString{String: ipAddress, Valid: ipAddress != ""},
		UserAgent:    sql.NullString{String: userAgent, Valid: userAgent != ""},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return s.tokenRepo.Create(token)
}

// RevokeToken revokes a token (for logout)
func (s *authService) RevokeToken(userID, accessToken string) error {
	log.Printf("UserId: %v", userID)
	return s.tokenRepo.DeleteByUserIdAndAccessToken(userID, accessToken)
}

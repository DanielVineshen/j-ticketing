// File: j-ticketing/internal/core/services/auth_service.go
package service

import (
	"database/sql"
	"errors"
	"fmt"
	dto "j-ticketing/internal/core/dto/auth"
	coremodels "j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/email"
	jwt "j-ticketing/pkg/jwt"
	util "j-ticketing/pkg/utils"
	"log"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AuthService is the interface for authentication operations
type AuthService interface {
	// Authentication
	LoginAdmin(username, password string) (*dto.TokenResponse, error)
	LoginCustomer(email, password string) (*dto.TokenResponse, error)

	// Token management
	ValidateToken(token string) (bool, error)
	RefreshToken(refreshToken string) (*dto.TokenResponse, error)
	SaveToken(userID, userType, accessToken, refreshToken, ipAddress, userAgent string) error
	RevokeToken(userID, refreshToken string) error

	// Customer management
	CreateCustomer(req *dto.CreateCustomerRequest) (*coremodels.Customer, error)
	ResetPassword(email string) (*coremodels.Customer, *dto.PasswordChangeResult, error)
}

type authService struct {
	jwtService      jwt.JWTService
	adminRepo       repositories.AdminRepository
	customerRepo    repositories.CustomerRepository
	tokenRepo       repositories.TokenRepository
	emailService    email.EmailService
	accessTokenTTL  int64
	refreshTokenTTL int64
}

// NewAuthService creates a new authentication service
func NewAuthService(
	jwtService jwt.JWTService,
	adminRepo repositories.AdminRepository,
	customerRepo repositories.CustomerRepository,
	tokenRepo repositories.TokenRepository,
	emailService email.EmailService,
	accessTokenTTL, refreshTokenTTL int64,
) AuthService {
	return &authService{
		jwtService:      jwtService,
		adminRepo:       adminRepo,
		customerRepo:    customerRepo,
		tokenRepo:       tokenRepo,
		emailService:    emailService,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// LoginAdmin handles admin login authentication
func (s *authService) LoginAdmin(username, password string) (*dto.TokenResponse, error) {
	// Find admin by username
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Create user claims
	userClaims := &dto.UserClaims{
		UserID:   strconv.FormatUint(uint64(admin.AdminId), 10),
		Username: admin.Username,
		UserType: "admin",
		Role:     admin.Role,
		Roles:    []string{admin.Role},
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
		AdminID:    admin.AdminId,
		Username:   admin.Username,
		Role:       admin.Role,
		FullName:   admin.FullName,
		Email:      admin.Email,
		ContactNo:  admin.ContactNo,
		IsDisabled: admin.IsDisabled,
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenTTL,
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
		Role:             userClaims.Role,
		FullName:         customer.FullName,
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.accessTokenTTL,
		User:         userInfo,
	}, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(token string) (bool, error) {
	// First validate the token cryptographically
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return false, err
	}

	// Get the identifier to use for token lookup
	lookupID := claims.Username

	// Check if token exists in database
	_, err = s.tokenRepo.FindByUserIdAndAccessToken(lookupID, token)
	if err != nil {
		return false, err
	}

	return true, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(refreshToken string) (*dto.TokenResponse, error) {
	// Validate refresh token cryptographically
	claims, err := s.jwtService.ValidateToken(refreshToken)
	if err != nil {
		fmt.Printf("Refresh token failed JWT validation: %v\n", err)
		return nil, err
	}

	// Get the identifier to use for token lookup
	lookupID := claims.Username
	fmt.Printf("Refreshing token for user: %s\n", lookupID)

	// Check if refresh token exists in database
	dbToken, err := s.tokenRepo.FindByUserIdAndRefreshToken(lookupID, refreshToken)
	if err != nil {
		fmt.Printf("Refresh token not found in database for user %s: %v\n", lookupID, err)
		return nil, errors.New("invalid refresh token")
	}

	// Generate new access token
	accessToken, err := s.jwtService.GenerateToken(claims, false)
	if err != nil {
		fmt.Printf("Error generating new access token: %v\n", err)
		return nil, err
	}

	// Update the token in the database
	dbToken.AccessToken = accessToken
	dbToken.UpdatedAt = time.Now()
	if err := s.tokenRepo.UpdateToken(dbToken); err != nil {
		fmt.Printf("Warning: Failed to update token in database: %v\n", err)
		// Continue anyway since we have generated the new token
	} else {
		fmt.Printf("Successfully updated access token for user %s\n", lookupID)
	}

	// Create the appropriate response based on user type
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
			TokenType:    "Bearer",
			ExpiresIn:    s.accessTokenTTL,
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
			TokenType:    "Bearer",
			ExpiresIn:    s.accessTokenTTL,
			User:         userInfo,
		}, nil
	}
}

// SaveToken saves a token to the database
func (s *authService) SaveToken(userID, userType, accessToken, refreshToken, ipAddress, userAgent string) error {
	// Check how many tokens the user already has
	count, err := s.tokenRepo.CountByUserId(userID)
	if err != nil {
		fmt.Printf("Error counting tokens for user %s: %v\n", userID, err)
		// Continue anyway since this isn't a critical error
	}

	fmt.Printf("User %s has %d existing tokens\n", userID, count)

	// If user has 3 or more tokens, delete the oldest one
	if count >= 3 {
		oldestToken, err := s.tokenRepo.FindOldestByUserId(userID)
		if err != nil {
			fmt.Printf("Error finding oldest token for user %s: %v\n", userID, err)
			// Continue anyway
		} else {
			// Delete the oldest token
			if err := s.tokenRepo.DeleteToken(oldestToken); err != nil {
				fmt.Printf("Error deleting oldest token for user %s: %v\n", userID, err)
				// Continue anyway
			} else {
				fmt.Printf("Successfully deleted oldest token (ID: %d) for user %s\n", oldestToken.TokenId, userID)
			}
		}
	}

	// Create the new token
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

	// Create the token in the database
	err = s.tokenRepo.Create(token)
	if err != nil {
		fmt.Printf("Error creating token for user %s: %v\n", userID, err)
		return err
	}

	fmt.Printf("Token created successfully for user %s\n", userID)
	return nil
}

// RevokeToken revokes a token (for logout)
func (s *authService) RevokeToken(userID, accessToken string) error {
	return s.tokenRepo.DeleteByUserIdAndAccessToken(userID, accessToken)
}

// CreateCustomer creates a new customer
func (s *authService) CreateCustomer(req *dto.CreateCustomerRequest) (*coremodels.Customer, error) {
	// Check if email already exists
	existingCustomer, err := s.customerRepo.FindByEmail(req.Email)
	if err == nil && existingCustomer != nil {
		return nil, errors.New("email already exists")
	}

	// Generate a unique customer ID
	custID, err := util.GenerateCustomerID("CUST")
	if err != nil {
		return nil, fmt.Errorf("failed to generate customer ID: %w", err)
	}

	// Hash the password if provided
	var hashedPassword sql.NullString
	if req.Password != "" {
		pwHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		hashedPassword = sql.NullString{String: string(pwHash), Valid: true}
	}

	// Create customer model
	customer := &coremodels.Customer{
		CustId:           custID,
		Email:            req.Email,
		Password:         hashedPassword,
		IdentificationNo: req.IdentificationNo,
		FullName:         req.FullName,
		ContactNo:        sql.NullString{String: req.ContactNo, Valid: req.ContactNo != ""},
		IsDisabled:       false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	if err := s.customerRepo.Create(customer); err != nil {
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Mask password in the returned object for security
	customer.Password = sql.NullString{String: "", Valid: false}

	return customer, nil
}

// ResetPassword resets a customer's password
func (s *authService) ResetPassword(email string) (*coremodels.Customer, *dto.PasswordChangeResult, error) {
	// Find customer by email
	customer, err := s.customerRepo.FindByEmail(email)

	// If customer doesn't exist, return success anyway (security measure)
	if err != nil {
		return customer,
			&dto.PasswordChangeResult{
				Success: true,
				Message: "If your email exists in our system, you will receive a password reset email shortly.",
			}, nil
	}

	// Generate a new random password (12 characters)
	newPassword, err := util.GenerateRandomPassword(12)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Update customer's password
	customer.Password = sql.NullString{String: string(hashedPassword), Valid: true}
	customer.UpdatedAt = time.Now()

	if err := s.customerRepo.Update(customer); err != nil {
		return nil, nil, fmt.Errorf("failed to update customer password: %w", err)
	}

	// Send email with the new password
	err = s.emailService.SendPasswordResetEmail(email, newPassword)
	if err != nil {
		log.Printf("Failed to send password reset email to %s: %v", email, err)
		// Continue anyway since the password has been reset
	}

	return customer,
		&dto.PasswordChangeResult{
			Success: true,
			Message: "If your email exists in our system, you will receive a password reset email shortly.",
		}, nil
}

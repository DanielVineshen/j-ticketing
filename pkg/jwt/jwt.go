// FILE: internal/auth/jwt/jwt.go
package jwt

import (
	"errors"
	"j-ticketing/internal/core/dto"
	"j-ticketing/pkg/config"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService is the interface for JWT operations
type JWTService interface {
	GenerateToken(userClaims *dto.UserClaims, isRefreshToken bool) (string, error)
	ValidateToken(tokenString string) (*dto.UserClaims, error)
	ExtractTokenFromHeader(authHeader string) string
}

type jwtService struct {
	secretKey       string
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.Config) JWTService {
	return &jwtService{
		secretKey:       cfg.JWT.SecretKey,
		accessTokenTTL:  time.Duration(cfg.JWT.AccessTokenTTL) * time.Minute,
		refreshTokenTTL: time.Duration(cfg.JWT.RefreshTokenTTL) * time.Hour,
	}
}

// GenerateToken generates a new JWT token
func (s *jwtService) GenerateToken(userClaims *dto.UserClaims, isRefreshToken bool) (string, error) {
	// Determine token expiration time
	var expiration time.Duration
	if isRefreshToken {
		expiration = s.refreshTokenTTL
	} else {
		expiration = s.accessTokenTTL
	}

	// Create claims with user info and expiration time
	claims := jwt.MapClaims{
		"userId":   userClaims.UserID,
		"username": userClaims.Username,
		"userType": userClaims.UserType,
		"role":     userClaims.Role,
		"roles":    userClaims.Roles,
		"exp":      time.Now().Add(expiration).Unix(),
		"iat":      time.Now().Unix(),
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate signed token
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *jwtService) ValidateToken(tokenString string) (*dto.UserClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	// Extract user info from claims
	userID, _ := claims["userId"].(string)
	username, _ := claims["username"].(string)
	userType, _ := claims["userType"].(string)
	role, _ := claims["role"].(string)

	// Extract roles array
	var roles []string
	if rolesArr, ok := claims["roles"].([]interface{}); ok {
		for _, r := range rolesArr {
			if role, ok := r.(string); ok {
				roles = append(roles, role)
			}
		}
	}

	// Create user claims object
	userClaims := &dto.UserClaims{
		UserID:   userID,
		Username: username,
		UserType: userType,
		Role:     role,
		Roles:    roles,
	}

	return userClaims, nil
}

// ExtractTokenFromHeader extracts the token from the Authorization header
func (s *jwtService) ExtractTokenFromHeader(authHeader string) string {
	// Check if the header starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	// Extract the token
	return strings.TrimPrefix(authHeader, "Bearer ")
}

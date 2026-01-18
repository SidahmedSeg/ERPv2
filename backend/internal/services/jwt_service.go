package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/config"
)

// JWTService handles JWT token generation and validation
type JWTService struct {
	config *config.JWTConfig
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg *config.JWTConfig) *JWTService {
	return &JWTService{
		config: cfg,
	}
}

// Claims represents the JWT claims
type Claims struct {
	UserID     uuid.UUID `json:"user_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantSlug string    `json:"tenant_slug"`
	Email      string    `json:"email"`
	TokenType  string    `json:"token_type"` // access | refresh | 2fa
	jwt.RegisteredClaims
}

// TokenType constants
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
	TokenType2FA     = "2fa"
)

// GenerateAccessToken generates an access token for a user
func (s *JWTService) GenerateAccessToken(userID, tenantID uuid.UUID, tenantSlug, email string, rememberMe bool) (string, int64, error) {
	// Determine expiry based on rememberMe
	var expiresIn time.Duration
	if rememberMe {
		expiresIn = s.config.RememberMeExpiry
	} else {
		expiresIn = s.config.AccessTokenExpiry
	}

	expiresAt := time.Now().Add(expiresIn)

	claims := Claims{
		UserID:     userID,
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
		Email:      email,
		TokenType:  TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, int64(expiresIn.Seconds()), nil
}

// GenerateRefreshToken generates a refresh token for a user
func (s *JWTService) GenerateRefreshToken(userID, tenantID uuid.UUID, tenantSlug, email string) (string, error) {
	expiresAt := time.Now().Add(s.config.RefreshTokenExpiry)

	claims := Claims{
		UserID:     userID,
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
		Email:      email,
		TokenType:  TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.RefreshSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// Generate2FAToken generates a temporary token for 2FA verification
func (s *JWTService) Generate2FAToken(userID, tenantID uuid.UUID, tenantSlug, email string) (string, error) {
	// 2FA tokens are short-lived (5 minutes)
	expiresAt := time.Now().Add(5 * time.Minute)

	claims := Claims{
		UserID:     userID,
		TenantID:   tenantID,
		TenantSlug: tenantSlug,
		Email:      email,
		TokenType:  TokenType2FA,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.Issuer,
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.Secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign 2FA token: %w", err)
	}

	return tokenString, nil
}

// ValidateAccessToken validates an access token and returns the claims
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.Secret, TokenTypeAccess)
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.RefreshSecret, TokenTypeRefresh)
}

// Validate2FAToken validates a 2FA token and returns the claims
func (s *JWTService) Validate2FAToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString, s.config.Secret, TokenType2FA)
}

// validateToken validates a token with the specified secret and token type
func (s *JWTService) validateToken(tokenString, secret, expectedType string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Verify token type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Verify token is not expired
	if time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string, rememberMe bool) (string, int64, error) {
	// Validate refresh token
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", 0, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token
	return s.GenerateAccessToken(claims.UserID, claims.TenantID, claims.TenantSlug, claims.Email, rememberMe)
}

// ExtractClaims extracts claims from a token without full validation (for debugging)
func (s *JWTService) ExtractClaims(tokenString string) (*Claims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

// GetTokenExpiry returns the expiry time from a token string
func (s *JWTService) GetTokenExpiry(tokenString string) (time.Time, error) {
	claims, err := s.ExtractClaims(tokenString)
	if err != nil {
		return time.Time{}, err
	}

	return claims.ExpiresAt.Time, nil
}

// IsTokenExpired checks if a token is expired without full validation
func (s *JWTService) IsTokenExpired(tokenString string) bool {
	expiry, err := s.GetTokenExpiry(tokenString)
	if err != nil {
		return true
	}

	return time.Now().After(expiry)
}

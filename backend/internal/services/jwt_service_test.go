package services

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAccessToken(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()
	email := "test@example.com"
	tenantSlug := "test-tenant"
	roles := []string{"admin", "user"}

	// Test: Generate access token
	token, err := service.GenerateAccessToken(tenantID, userID, email, tenantSlug, roles)

	require.NoError(t, err, "GenerateAccessToken should not return error")
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.Greater(t, len(token), 50, "Token should be a valid JWT string")
}

func TestJWTService_ValidateToken(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()
	email := "test@example.com"
	tenantSlug := "test-tenant"
	roles := []string{"admin"}

	// Generate token
	token, err := service.GenerateAccessToken(tenantID, userID, email, tenantSlug, roles)
	require.NoError(t, err)

	// Test: Validate valid token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err, "ValidateToken should not return error for valid token")

	assert.Equal(t, userID, claims.UserID, "UserID should match")
	assert.Equal(t, tenantID, claims.TenantID, "TenantID should match")
	assert.Equal(t, email, claims.Email, "Email should match")
	assert.Equal(t, tenantSlug, claims.TenantSlug, "TenantSlug should match")
	assert.Equal(t, roles, claims.Roles, "Roles should match")
	assert.Equal(t, "access", claims.TokenType, "TokenType should be 'access'")

	// Test: Validate invalid token
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.signature"
	_, err = service.ValidateToken(invalidToken)
	assert.Error(t, err, "ValidateToken should return error for invalid token")
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()

	// Test: Generate refresh token
	token, err := service.GenerateRefreshToken(tenantID, userID)

	require.NoError(t, err, "GenerateRefreshToken should not return error")
	assert.NotEmpty(t, token, "Refresh token should not be empty")

	// Validate token claims
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, "refresh", claims.TokenType, "TokenType should be 'refresh'")
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)
}

func TestJWTService_Generate2FAToken(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()

	// Test: Generate 2FA token
	token, err := service.Generate2FAToken(tenantID, userID)

	require.NoError(t, err, "Generate2FAToken should not return error")
	assert.NotEmpty(t, token, "2FA token should not be empty")

	// Validate token claims
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)

	assert.Equal(t, "2fa", claims.TokenType, "TokenType should be '2fa'")
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, tenantID, claims.TenantID)

	// Verify expiration is short (5 minutes)
	expectedExpiry := time.Now().Add(5 * time.Minute)
	assert.InDelta(t, expectedExpiry.Unix(), claims.ExpiresAt.Unix(), 10,
		"2FA token should expire in approximately 5 minutes")
}

func TestJWTService_ExpiredToken(t *testing.T) {
	// This test would require mocking time or generating a token with negative expiry
	// For now, we'll skip detailed implementation
	t.Skip("Expired token test requires time mocking")
}

func TestJWTService_InvalidSecret(t *testing.T) {
	// Setup with different secrets
	jwtSecret1 := "test-secret-key-minimum-32-characters-long-for-security-1"
	jwtSecret2 := "test-secret-key-minimum-32-characters-long-for-security-2"

	service1 := NewJWTService(jwtSecret1, nil)
	service2 := NewJWTService(jwtSecret2, nil)

	tenantID := uuid.New()
	userID := uuid.New()

	// Generate token with service1
	token, err := service1.GenerateAccessToken(tenantID, userID, "test@example.com", "test-tenant", []string{"user"})
	require.NoError(t, err)

	// Test: Validate with different secret should fail
	_, err = service2.ValidateToken(token)
	assert.Error(t, err, "ValidateToken should fail when using different secret")
}

func TestJWTService_EmptyRoles(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()

	// Test: Generate token with empty roles
	token, err := service.GenerateAccessToken(tenantID, userID, "test@example.com", "test-tenant", []string{})
	require.NoError(t, err)

	// Validate
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Empty(t, claims.Roles, "Roles should be empty array")
}

func TestJWTService_MultipleRoles(t *testing.T) {
	// Setup
	jwtSecret := "test-secret-key-minimum-32-characters-long-for-security"
	service := NewJWTService(jwtSecret, nil)

	tenantID := uuid.New()
	userID := uuid.New()
	roles := []string{"admin", "manager", "user", "developer"}

	// Test: Generate token with multiple roles
	token, err := service.GenerateAccessToken(tenantID, userID, "test@example.com", "test-tenant", roles)
	require.NoError(t, err)

	// Validate
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.ElementsMatch(t, roles, claims.Roles, "All roles should be present")
}

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	authService *services.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate validates JWT token and adds user/tenant to context
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.Unauthorized(w, "Missing authorization header")
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Unauthorized(w, "Invalid authorization header format")
			return
		}

		accessToken := parts[1]

		// Validate session and get user/tenant
		user, tenant, err := m.authService.ValidateSession(r.Context(), accessToken)
		if err != nil {
			utils.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Add user, tenant, and token to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "tenant_id", tenant.ID)
		ctx = context.WithValue(ctx, "tenant_slug", tenant.Slug)
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "tenant", tenant)
		ctx = context.WithValue(ctx, "access_token", accessToken)

		// Continue with authenticated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Optional makes authentication optional (adds context if token is present)
func (m *AuthMiddleware) Optional(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		accessToken := parts[1]

		// Validate session and get user/tenant
		user, tenant, err := m.authService.ValidateSession(r.Context(), accessToken)
		if err != nil {
			// Invalid token, continue without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Add user, tenant, and token to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "tenant_id", tenant.ID)
		ctx = context.WithValue(ctx, "tenant_slug", tenant.Slug)
		ctx = context.WithValue(ctx, "user", user)
		ctx = context.WithValue(ctx, "tenant", tenant)
		ctx = context.WithValue(ctx, "access_token", accessToken)

		// Continue with authenticated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireStatus ensures user has a specific status
func (m *AuthMiddleware) RequireStatus(allowedStatuses ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := r.Context().Value("user").(*models.User)
			if !ok || user == nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			// Check if user has one of the allowed statuses
			for _, status := range allowedStatuses {
				if user.Status == status {
					next.ServeHTTP(w, r)
					return
				}
			}

			utils.Forbidden(w, "User does not have required status")
		})
	}
}

// RequireEmailVerified ensures user's email is verified
func (m *AuthMiddleware) RequireEmailVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value("user").(*models.User)
		if !ok || user == nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		if !user.EmailVerified {
			utils.Forbidden(w, "Email verification required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireTenantActive ensures tenant is active
func (m *AuthMiddleware) RequireTenantActive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, ok := r.Context().Value("tenant").(*models.Tenant)
		if !ok || tenant == nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		if !tenant.CanAccess() {
			utils.Forbidden(w, "Tenant account is not active")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext extracts user from context
func GetUserFromContext(ctx context.Context) (*models.User, error) {
	user, ok := ctx.Value("user").(*models.User)
	if !ok || user == nil {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

// GetTenantFromContext extracts tenant from context
func GetTenantFromContext(ctx context.Context) (*models.Tenant, error) {
	tenant, ok := ctx.Value("tenant").(*models.Tenant)
	if !ok || tenant == nil {
		return nil, fmt.Errorf("tenant not found in context")
	}
	return tenant, nil
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value("user_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value("tenant_id").(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("tenant ID not found in context")
	}
	return tenantID, nil
}

// GetTenantSlugFromContext extracts tenant slug from context
func GetTenantSlugFromContext(ctx context.Context) (string, error) {
	slug, ok := ctx.Value("tenant_slug").(string)
	if !ok {
		return "", fmt.Errorf("tenant slug not found in context")
	}
	return slug, nil
}

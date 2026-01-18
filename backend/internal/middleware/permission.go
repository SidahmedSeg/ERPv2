package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// PermissionMiddleware handles authorization
type PermissionMiddleware struct {
	permissionService *services.PermissionService
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(permissionService *services.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequirePermission checks if user has a specific permission
func (m *PermissionMiddleware) RequirePermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and tenant from context (set by auth middleware)
			userID, err := GetUserIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			tenantID, err := GetTenantIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			// Check permission
			hasPermission, err := m.permissionService.HasPermission(r.Context(), tenantID, userID, resource, action)
			if err != nil {
				utils.InternalServerError(w, "Failed to check permissions")
				return
			}

			if !hasPermission {
				utils.Forbidden(w, fmt.Sprintf("Missing required permission: %s.%s", resource, action))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission checks if user has any of the specified permissions
func (m *PermissionMiddleware) RequireAnyPermission(checks []services.PermissionCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and tenant from context
			userID, err := GetUserIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			tenantID, err := GetTenantIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			// Check if user has any of the permissions
			hasAny, err := m.permissionService.HasAnyPermission(r.Context(), tenantID, userID, checks)
			if err != nil {
				utils.InternalServerError(w, "Failed to check permissions")
				return
			}

			if !hasAny {
				utils.Forbidden(w, "Missing required permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions checks if user has all of the specified permissions
func (m *PermissionMiddleware) RequireAllPermissions(checks []services.PermissionCheck) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and tenant from context
			userID, err := GetUserIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			tenantID, err := GetTenantIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			// Check if user has all permissions
			hasAll, err := m.permissionService.HasAllPermissions(r.Context(), tenantID, userID, checks)
			if err != nil {
				utils.InternalServerError(w, "Failed to check permissions")
				return
			}

			if !hasAll {
				utils.Forbidden(w, "Missing required permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if user has a specific role
func (m *PermissionMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and tenant from context
			userID, err := GetUserIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			tenantID, err := GetTenantIDFromContext(r.Context())
			if err != nil {
				utils.Unauthorized(w, "Authentication required")
				return
			}

			// Check role
			hasRole, err := m.permissionService.CheckUserRole(r.Context(), tenantID, userID, roleName)
			if err != nil {
				utils.InternalServerError(w, "Failed to check role")
				return
			}

			if !hasRole {
				utils.Forbidden(w, fmt.Sprintf("Missing required role: %s", roleName))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOwner ensures user is an owner
func (m *PermissionMiddleware) RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user and tenant from context
		userID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		tenantID, err := GetTenantIDFromContext(r.Context())
		if err != nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		// Check if user is owner
		isOwner, err := m.permissionService.IsOwner(r.Context(), tenantID, userID)
		if err != nil {
			utils.InternalServerError(w, "Failed to check ownership")
			return
		}

		if !isOwner {
			utils.Forbidden(w, "Owner access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAdmin ensures user is an admin or owner
func (m *PermissionMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user and tenant from context
		userID, err := GetUserIDFromContext(r.Context())
		if err != nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		tenantID, err := GetTenantIDFromContext(r.Context())
		if err != nil {
			utils.Unauthorized(w, "Authentication required")
			return
		}

		// Check if user is admin or owner
		isAdmin, err := m.permissionService.IsAdmin(r.Context(), tenantID, userID)
		if err != nil {
			utils.InternalServerError(w, "Failed to check admin status")
			return
		}

		if !isAdmin {
			utils.Forbidden(w, "Admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// OptionalPermission checks permission but doesn't block if missing (adds to context)
func (m *PermissionMiddleware) OptionalPermission(resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user and tenant from context
			userID, err := GetUserIDFromContext(r.Context())
			if err != nil {
				// No auth, continue
				next.ServeHTTP(w, r)
				return
			}

			tenantID, err := GetTenantIDFromContext(r.Context())
			if err != nil {
				// No tenant, continue
				next.ServeHTTP(w, r)
				return
			}

			// Check permission
			hasPermission, err := m.permissionService.HasPermission(r.Context(), tenantID, userID, resource, action)
			if err != nil {
				// Error checking, continue
				next.ServeHTTP(w, r)
				return
			}

			// Add permission check result to context
			ctx := context.WithValue(r.Context(), fmt.Sprintf("has_permission:%s.%s", resource, action), hasPermission)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helper to check if permission is available from context (set by OptionalPermission)
func HasPermissionFromContext(ctx context.Context, resource, action string) bool {
	key := fmt.Sprintf("has_permission:%s.%s", resource, action)
	hasPerm, ok := ctx.Value(key).(bool)
	return ok && hasPerm
}

// CheckPermission is a helper to check permissions in handlers without middleware
func (m *PermissionMiddleware) CheckPermission(ctx context.Context, tenantID, userID uuid.UUID, resource, action string) (bool, error) {
	return m.permissionService.HasPermission(ctx, tenantID, userID, resource, action)
}

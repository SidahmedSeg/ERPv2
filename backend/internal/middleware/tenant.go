package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/utils"
)

var (
	ErrNoTenantContext = errors.New("no tenant context found")
)

// TenantMiddleware resolves tenant from request
type TenantMiddleware struct {
	tenantRepo *repository.TenantRepository
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(tenantRepo *repository.TenantRepository) *TenantMiddleware {
	return &TenantMiddleware{
		tenantRepo: tenantRepo,
	}
}

// ResolveTenant resolves tenant from subdomain or X-Tenant-Slug header
func (m *TenantMiddleware) ResolveTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tenantSlug string

		// 1. Try X-Tenant-Slug header (for testing/development)
		if slug := r.Header.Get("X-Tenant-Slug"); slug != "" {
			tenantSlug = slug
		} else {
			// 2. Try subdomain (e.g., acme.myerp.local -> "acme")
			host := r.Host
			// Remove port if present
			if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
				host = host[:colonIndex]
			}

			// Split by dots and check for subdomain
			parts := strings.Split(host, ".")
			if len(parts) >= 2 {
				// If host is like "acme.myerp.local", tenant is "acme"
				// If host is "myerp.local" or "localhost", no tenant
				if parts[0] != "localhost" && parts[0] != "myerp" {
					tenantSlug = parts[0]
				}
			}
		}

		// If no tenant slug found, continue without setting context
		// (some endpoints like /register don't need tenant)
		if tenantSlug == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Find tenant by slug
		tenant, err := m.tenantRepo.FindBySlug(r.Context(), tenantSlug)
		if err != nil {
			utils.NotFound(w, "Tenant not found")
			return
		}

		// Check tenant status
		if !tenant.CanAccess() {
			utils.Forbidden(w, "Tenant account is not active")
			return
		}

		// Add tenant to context (use same keys as auth.go)
		ctx := context.WithValue(r.Context(), "tenant_id", tenant.ID)
		ctx = context.WithValue(ctx, "tenant_slug", tenant.Slug)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireTenant is a middleware that requires tenant context
func (m *TenantMiddleware) RequireTenant(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, ok := r.Context().Value("tenant_id").(uuid.UUID)
		if !ok {
			utils.BadRequest(w, "Tenant context required. Please provide X-Tenant-Slug header or use tenant subdomain.")
			return
		}

		// Verify tenant ID is not zero
		if tenantID == uuid.Nil {
			utils.BadRequest(w, "Invalid tenant context")
			return
		}

		next.ServeHTTP(w, r)
	})
}


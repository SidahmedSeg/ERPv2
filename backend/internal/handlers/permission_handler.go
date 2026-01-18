package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// PermissionHandler handles permission endpoints
type PermissionHandler struct {
	permissionService *services.PermissionService
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(permissionService *services.PermissionService) *PermissionHandler {
	return &PermissionHandler{
		permissionService: permissionService,
	}
}

// List retrieves all permissions
// GET /api/permissions
func (h *PermissionHandler) List(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.permissionService.ListAllPermissions(r.Context())
	if err != nil {
		utils.InternalServerError(w, "Failed to list permissions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// ListByCategory retrieves permissions grouped by category
// GET /api/permissions/by-category
func (h *PermissionHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	groups, err := h.permissionService.ListPermissionsByCategory(r.Context())
	if err != nil {
		utils.InternalServerError(w, "Failed to list permissions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"groups": groups,
		"count":  len(groups),
	})
}

// Search searches for permissions
// GET /api/permissions/search?q=keyword
func (h *PermissionHandler) Search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		utils.BadRequest(w, "Search term is required")
		return
	}

	permissions, err := h.permissionService.SearchPermissions(r.Context(), searchTerm)
	if err != nil {
		utils.InternalServerError(w, "Failed to search permissions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"permissions": permissions,
		"count":       len(permissions),
		"query":       searchTerm,
	})
}

// GetStats retrieves permission statistics
// GET /api/permissions/stats
func (h *PermissionHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.permissionService.GetPermissionStats(r.Context())
	if err != nil {
		utils.InternalServerError(w, "Failed to get stats")
		return
	}

	utils.Success(w, stats)
}

// GetMyPermissions retrieves permissions for the current user
// GET /api/permissions/me
func (h *PermissionHandler) GetMyPermissions(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	permissions, err := h.permissionService.GetUserPermissions(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get user permissions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// CheckPermission checks if current user has a specific permission
// POST /api/permissions/check
func (h *PermissionHandler) CheckPermission(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if req.Resource == "" || req.Action == "" {
		utils.BadRequest(w, "Resource and action are required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	hasPermission, err := h.permissionService.HasPermission(r.Context(), tenantID, userID, req.Resource, req.Action)
	if err != nil {
		utils.InternalServerError(w, "Failed to check permission")
		return
	}

	utils.Success(w, map[string]interface{}{
		"has_permission": hasPermission,
		"resource":       req.Resource,
		"action":         req.Action,
	})
}

// RegisterRoutes registers all permission routes
func (h *PermissionHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/permissions", func(r chi.Router) {
		// All permission routes require authentication
		r.Use(authMiddleware.Authenticate)

		// List all permissions - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/", h.List)

		// List by category - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/by-category", h.ListByCategory)

		// Search permissions - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/search", h.Search)

		// Permission stats - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/stats", h.GetStats)

		// Get current user's permissions - no special permission needed
		r.Get("/me", h.GetMyPermissions)

		// Check permission - no special permission needed
		r.Post("/check", h.CheckPermission)
	})
}

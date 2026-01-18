package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// RoleHandler handles role management endpoints
type RoleHandler struct {
	roleRepo         *repository.RoleRepository
	userRoleRepo     *repository.UserRoleRepository
	permissionService *services.PermissionService
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(
	roleRepo *repository.RoleRepository,
	userRoleRepo *repository.UserRoleRepository,
	permissionService *services.PermissionService,
) *RoleHandler {
	return &RoleHandler{
		roleRepo:          roleRepo,
		userRoleRepo:      userRoleRepo,
		permissionService: permissionService,
	}
}

// List retrieves all roles
// GET /api/roles
func (h *RoleHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Query parameter to include details
	includeDetails := r.URL.Query().Get("include_details") == "true"

	var roles []models.Role

	if includeDetails {
		roles, err = h.roleRepo.ListWithDetails(r.Context(), tenantID)
	} else {
		roles, err = h.roleRepo.List(r.Context(), tenantID)
	}

	if err != nil {
		utils.InternalServerError(w, "Failed to list roles")
		return
	}

	utils.Success(w, map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	})
}

// Get retrieves a single role by ID
// GET /api/roles/{id}
func (h *RoleHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	// Get role with details
	role, err := h.roleRepo.GetRoleWithDetails(r.Context(), tenantID, roleID)
	if err != nil {
		utils.NotFound(w, "Role not found")
		return
	}

	utils.Success(w, map[string]interface{}{
		"role": role,
	})
}

// Create creates a new role
// POST /api/roles
func (h *RoleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.RoleCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("name", req.Name, "Role name", &errors)
	utils.ValidateRequired("display_name", req.DisplayName, "Display name", &errors)
	utils.ValidateStringLength("name", req.Name, 2, 100, "Role name", &errors)
	utils.ValidateStringLength("display_name", req.DisplayName, 2, 255, "Display name", &errors)

	if len(req.PermissionIDs) == 0 {
		errors.Add("permission_ids", "At least one permission is required")
	}

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Check if role name already exists
	exists, err := h.roleRepo.CheckNameExists(r.Context(), tenantID, req.Name, nil)
	if err != nil {
		utils.InternalServerError(w, "Failed to check role name")
		return
	}
	if exists {
		utils.Conflict(w, "Role name already exists")
		return
	}

	// Validate permission IDs
	valid, _, err := h.permissionService.ValidatePermissionIDs(r.Context(), req.PermissionIDs)
	if err != nil {
		utils.InternalServerError(w, "Failed to validate permissions")
		return
	}
	if !valid {
		utils.BadRequest(w, "Some permission IDs do not exist")
		return
	}

	// Create role
	role := &models.Role{
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  &req.Description,
		ParentRoleID: req.ParentRoleID,
		Level:        0, // Will be calculated based on parent if needed
		IsSystem:     false,
		CreatedBy:    &userID,
	}

	if err := h.roleRepo.Create(r.Context(), tenantID, role); err != nil {
		utils.InternalServerError(w, "Failed to create role")
		return
	}

	// Assign permissions
	if err := h.roleRepo.AssignPermissions(r.Context(), tenantID, role.ID, req.PermissionIDs, userID); err != nil {
		utils.InternalServerError(w, "Failed to assign permissions")
		return
	}

	// Get role with details
	createdRole, _ := h.roleRepo.GetRoleWithDetails(r.Context(), tenantID, role.ID)

	utils.Created(w, map[string]interface{}{
		"role":    createdRole,
		"message": "Role created successfully",
	})
}

// Update updates an existing role
// PUT /api/roles/{id}
func (h *RoleHandler) Update(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	var req models.RoleUpdateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get existing role
	role, err := h.roleRepo.FindByID(r.Context(), tenantID, roleID)
	if err != nil {
		utils.NotFound(w, "Role not found")
		return
	}

	// Cannot update system roles
	if role.IsSystem {
		utils.Forbidden(w, "Cannot update system roles")
		return
	}

	// Update fields
	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = req.Description
	}
	if req.ParentRoleID != nil {
		role.ParentRoleID = req.ParentRoleID
	}

	// Update role
	if err := h.roleRepo.Update(r.Context(), tenantID, role); err != nil {
		utils.InternalServerError(w, "Failed to update role")
		return
	}

	// Update permissions if provided
	if req.PermissionIDs != nil && len(req.PermissionIDs) > 0 {
		// Validate permission IDs
		valid, _, err := h.permissionService.ValidatePermissionIDs(r.Context(), req.PermissionIDs)
		if err != nil {
			utils.InternalServerError(w, "Failed to validate permissions")
			return
		}
		if !valid {
			utils.BadRequest(w, "Some permission IDs do not exist")
			return
		}

		if err := h.roleRepo.AssignPermissions(r.Context(), tenantID, roleID, req.PermissionIDs, userID); err != nil {
			utils.InternalServerError(w, "Failed to update permissions")
			return
		}

		// Invalidate permission cache for all users with this role
		h.permissionService.InvalidateRolePermissions(r.Context(), tenantID, roleID)
	}

	// Get updated role with details
	updatedRole, _ := h.roleRepo.GetRoleWithDetails(r.Context(), tenantID, roleID)

	utils.Success(w, map[string]interface{}{
		"role":    updatedRole,
		"message": "Role updated successfully",
	})
}

// Delete deletes a role
// DELETE /api/roles/{id}
func (h *RoleHandler) Delete(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Check if role has users
	userCount, err := h.roleRepo.CountUsers(r.Context(), tenantID, roleID)
	if err != nil {
		utils.InternalServerError(w, "Failed to check role usage")
		return
	}

	if userCount > 0 {
		utils.BadRequest(w, "Cannot delete role with assigned users")
		return
	}

	// Delete role
	if err := h.roleRepo.Delete(r.Context(), tenantID, roleID); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	// Invalidate permission cache for users with this role
	h.permissionService.InvalidateRolePermissions(r.Context(), tenantID, roleID)

	utils.Success(w, map[string]interface{}{
		"message": "Role deleted successfully",
	})
}

// GetPermissions retrieves permissions for a role
// GET /api/roles/{id}/permissions
func (h *RoleHandler) GetPermissions(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	permissions, err := h.roleRepo.GetPermissions(r.Context(), tenantID, roleID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get permissions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"permissions": permissions,
		"count":       len(permissions),
	})
}

// GetUsers retrieves users assigned to a role
// GET /api/roles/{id}/users
func (h *RoleHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	users, err := h.userRoleRepo.GetUsersByRole(r.Context(), tenantID, roleID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get users")
		return
	}

	utils.Success(w, map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// AssignToUsers assigns a role to multiple users
// POST /api/roles/{id}/assign
func (h *RoleHandler) AssignToUsers(w http.ResponseWriter, r *http.Request) {
	roleIDStr := chi.URLParam(r, "id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid role ID")
		return
	}

	var req struct {
		UserIDs []uuid.UUID `json:"user_ids"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.UserIDs) == 0 {
		utils.BadRequest(w, "At least one user ID is required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Bulk assign role
	if err := h.userRoleRepo.BulkAssignRole(r.Context(), tenantID, req.UserIDs, roleID, userID); err != nil {
		utils.InternalServerError(w, "Failed to assign role")
		return
	}

	// Invalidate cache for all affected users
	for _, targetUserID := range req.UserIDs {
		h.permissionService.InvalidateUserPermissions(r.Context(), tenantID, targetUserID)
	}

	utils.Success(w, map[string]interface{}{
		"message": "Role assigned successfully",
		"count":   len(req.UserIDs),
	})
}

// RegisterRoutes registers all role routes
func (h *RoleHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/roles", func(r chi.Router) {
		// All role routes require authentication
		r.Use(authMiddleware.Authenticate)

		// List roles - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/", h.List)

		// Get single role - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/{id}", h.Get)

		// Create role - requires create permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionCreate)).Post("/", h.Create)

		// Update role - requires edit permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionEdit)).Put("/{id}", h.Update)

		// Delete role - requires delete permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionDelete)).Delete("/{id}", h.Delete)

		// Role permissions - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/{id}/permissions", h.GetPermissions)

		// Role users - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionView)).Get("/{id}/users", h.GetUsers)

		// Assign role to users - requires assign permission
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionAssign)).Post("/{id}/assign", h.AssignToUsers)
	})
}

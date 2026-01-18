package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// UserHandler handles user management endpoints
type UserHandler struct {
	userRepo          *repository.UserRepository
	userRoleRepo      *repository.UserRoleRepository
	permissionService *services.PermissionService
	config            interface{} // Will be *config.Config
}

// NewUserHandler creates a new user handler
func NewUserHandler(
	userRepo *repository.UserRepository,
	userRoleRepo *repository.UserRoleRepository,
	permissionService *services.PermissionService,
) *UserHandler {
	return &UserHandler{
		userRepo:          userRepo,
		userRoleRepo:      userRoleRepo,
		permissionService: permissionService,
	}
}

// List retrieves all users with pagination
// GET /api/users?page=1&page_size=20
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	users, totalCount, err := h.userRepo.List(r.Context(), tenantID, pageSize, offset)
	if err != nil {
		utils.InternalServerError(w, "Failed to list users")
		return
	}

	// Enrich users with roles
	for i := range users {
		roles, _ := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, users[i].ID)
		users[i].Roles = roles
	}

	meta := utils.NewMeta(page, pageSize, totalCount)

	utils.SuccessWithMeta(w, map[string]interface{}{
		"users": users,
	}, meta)
}

// Get retrieves a single user by ID
// GET /api/users/{id}
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	user, err := h.userRepo.FindByID(r.Context(), tenantID, userID)
	if err != nil {
		utils.NotFound(w, "User not found")
		return
	}

	// Get user roles
	roles, _ := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, userID)
	user.Roles = roles

	// Get user permissions
	permissions, _ := h.permissionService.GetUserPermissions(r.Context(), tenantID, userID)
	user.Permissions = permissions

	utils.Success(w, map[string]interface{}{
		"user": user,
	})
}

// Create creates a new user
// POST /api/users
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("email", req.Email, "Email", &errors)
	utils.ValidateEmail("email", req.Email, &errors)
	utils.ValidateRequired("password", req.Password, "Password", &errors)
	utils.ValidatePassword("password", req.Password, &errors)
	utils.ValidateRequired("first_name", req.FirstName, "First name", &errors)
	utils.ValidateRequired("last_name", req.LastName, "Last name", &errors)
	utils.ValidateName("first_name", req.FirstName, "First name", &errors)
	utils.ValidateName("last_name", req.LastName, "Last name", &errors)

	if req.Phone != "" && !utils.IsValidPhone(req.Phone) {
		errors.Add("phone", "Invalid phone number")
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

	creatorID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Check if email already exists
	exists, err := h.userRepo.CheckEmailExists(r.Context(), tenantID, req.Email)
	if err != nil {
		utils.InternalServerError(w, "Failed to check email")
		return
	}
	if exists {
		utils.Conflict(w, "Email already registered")
		return
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password, 10)
	if err != nil {
		utils.InternalServerError(w, "Failed to hash password")
		return
	}

	// Create user
	user := &models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        &req.Phone,
		Status:       models.UserStatusActive,
		Timezone:     "UTC",
		Language:     "en",
		Preferences:  []byte("{}"),
		CreatedBy:    &creatorID,
	}

	if err := h.userRepo.Create(r.Context(), tenantID, user); err != nil {
		utils.InternalServerError(w, "Failed to create user")
		return
	}

	// Assign roles if provided
	if len(req.RoleIDs) > 0 {
		if err := h.userRoleRepo.AssignRoles(r.Context(), tenantID, user.ID, req.RoleIDs, creatorID); err != nil {
			// Log error but don't fail user creation
		}
	}

	// Get user with roles
	roles, _ := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, user.ID)
	user.Roles = roles

	utils.Created(w, map[string]interface{}{
		"user":    user,
		"message": "User created successfully",
	})
}

// Update updates an existing user
// PUT /api/users/{id}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	var req models.UserUpdateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get existing user
	user, err := h.userRepo.FindByID(r.Context(), tenantID, userID)
	if err != nil {
		utils.NotFound(w, "User not found")
		return
	}

	// Update fields
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}
	if req.Language != nil {
		user.Language = *req.Language
	}

	// Update user
	if err := h.userRepo.Update(r.Context(), tenantID, user); err != nil {
		utils.InternalServerError(w, "Failed to update user")
		return
	}

	// Get user with roles
	roles, _ := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, userID)
	user.Roles = roles

	utils.Success(w, map[string]interface{}{
		"user":    user,
		"message": "User updated successfully",
	})
}

// Delete deletes a user
// DELETE /api/users/{id}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	currentUserID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Cannot delete self
	if userID == currentUserID {
		utils.BadRequest(w, "Cannot delete your own account")
		return
	}

	// Delete user
	if err := h.userRepo.Delete(r.Context(), tenantID, userID); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	// Invalidate permission cache
	h.permissionService.InvalidateUserPermissions(r.Context(), tenantID, userID)

	utils.Success(w, map[string]interface{}{
		"message": "User deleted successfully",
	})
}

// UpdateStatus updates a user's status
// PATCH /api/users/{id}/status
func (h *UserHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	var req struct {
		Status string `json:"status"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate status
	validStatuses := []string{models.UserStatusActive, models.UserStatusSuspended, models.UserStatusDeactivated}
	utils.ValidateEnum("status", req.Status, validStatuses, "Status", &utils.ValidationErrors{})

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Update status
	if err := h.userRepo.UpdateStatus(r.Context(), tenantID, userID, req.Status); err != nil {
		utils.InternalServerError(w, "Failed to update status")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "User status updated successfully",
		"status":  req.Status,
	})
}

// GetRoles retrieves roles for a user
// GET /api/users/{id}/roles
func (h *UserHandler) GetRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	roles, err := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get user roles")
		return
	}

	utils.Success(w, map[string]interface{}{
		"roles": roles,
		"count": len(roles),
	})
}

// AssignRoles assigns roles to a user
// POST /api/users/{id}/roles
func (h *UserHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid user ID")
		return
	}

	var req struct {
		RoleIDs []uuid.UUID `json:"role_ids"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.RoleIDs) == 0 {
		utils.BadRequest(w, "At least one role ID is required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	assignerID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Assign roles
	if err := h.userRoleRepo.AssignRoles(r.Context(), tenantID, userID, req.RoleIDs, assignerID); err != nil {
		utils.InternalServerError(w, "Failed to assign roles")
		return
	}

	// Invalidate permission cache
	h.permissionService.InvalidateUserPermissions(r.Context(), tenantID, userID)

	// Get updated roles
	roles, _ := h.userRoleRepo.GetUserRoles(r.Context(), tenantID, userID)

	utils.Success(w, map[string]interface{}{
		"message": "Roles assigned successfully",
		"roles":   roles,
	})
}

// Search searches for users
// GET /api/users/search?q=keyword
func (h *UserHandler) Search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		utils.BadRequest(w, "Search term is required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	users, err := h.userRepo.Search(r.Context(), tenantID, searchTerm, 50)
	if err != nil {
		utils.InternalServerError(w, "Failed to search users")
		return
	}

	utils.Success(w, map[string]interface{}{
		"users": users,
		"count": len(users),
		"query": searchTerm,
	})
}

// RegisterRoutes registers all user routes
func (h *UserHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/users", func(r chi.Router) {
		// All user routes require authentication
		r.Use(authMiddleware.Authenticate)

		// List users - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/", h.List)

		// Search users - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/search", h.Search)

		// Get single user - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/{id}", h.Get)

		// Create user - requires create permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionCreate)).Post("/", h.Create)

		// Update user - requires edit permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionEdit)).Put("/{id}", h.Update)

		// Delete user - requires delete permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionDelete)).Delete("/{id}", h.Delete)

		// Update status - requires manage_status permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionManageStatus)).Patch("/{id}/status", h.UpdateStatus)

		// Get user roles - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/{id}/roles", h.GetRoles)

		// Assign roles - requires assign permission from roles resource
		r.With(permMiddleware.RequirePermission(models.ResourceRoles, models.ActionAssign)).Post("/{id}/roles", h.AssignRoles)
	})
}

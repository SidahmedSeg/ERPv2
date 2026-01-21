package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
	"myerp-v2/internal/utils"
)

// DepartmentHandler handles department management endpoints
type DepartmentHandler struct {
	departmentRepo *repository.DepartmentRepository
	userRepo       *repository.UserRepository
}

// NewDepartmentHandler creates a new department handler
func NewDepartmentHandler(
	departmentRepo *repository.DepartmentRepository,
	userRepo *repository.UserRepository,
) *DepartmentHandler {
	return &DepartmentHandler{
		departmentRepo: departmentRepo,
		userRepo:       userRepo,
	}
}

// List retrieves all departments
// GET /departments
func (h *DepartmentHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	departments, err := h.departmentRepo.List(r.Context(), tenantID)
	if err != nil {
		utils.InternalServerError(w, "Failed to list departments")
		return
	}

	utils.Success(w, map[string]interface{}{
		"departments": departments,
		"count":       len(departments),
	})
}

// Get retrieves a single department by ID
// GET /departments/{id}
func (h *DepartmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "id")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid department ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	department, err := h.departmentRepo.GetWithDetails(r.Context(), tenantID, deptID)
	if err != nil {
		utils.NotFound(w, "Department not found")
		return
	}

	utils.Success(w, map[string]interface{}{
		"department": department,
	})
}

// Create creates a new department
// POST /departments
func (h *DepartmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.DepartmentCreateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("name", req.Name, "Department name", &errors)
	utils.ValidateStringLength("name", req.Name, 2, 255, "Department name", &errors)
	utils.ValidateRequired("color", req.Color, "Color", &errors)
	utils.ValidateRequired("icon", req.Icon, "Icon", &errors)

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

	// Check if department name already exists
	exists, err := h.departmentRepo.CheckNameExists(r.Context(), tenantID, req.Name, nil)
	if err != nil {
		utils.InternalServerError(w, "Failed to check department name")
		return
	}
	if exists {
		utils.Conflict(w, "Department name already exists")
		return
	}

	// Validate head_user_id if provided
	if req.HeadUserID != nil {
		_, err := h.userRepo.FindByID(r.Context(), tenantID, *req.HeadUserID)
		if err != nil {
			utils.BadRequest(w, "Invalid head user ID")
			return
		}
	}

	// Create department
	department := &models.Department{
		Name:        req.Name,
		Description: req.Description,
		HeadUserID:  req.HeadUserID,
		Color:       req.Color,
		Icon:        req.Icon,
		Status:      models.DepartmentStatusActive,
		CreatedBy:   &creatorID,
	}

	if err := h.departmentRepo.Create(r.Context(), tenantID, department); err != nil {
		utils.InternalServerError(w, "Failed to create department")
		return
	}

	// Get department with details
	createdDept, _ := h.departmentRepo.GetWithDetails(r.Context(), tenantID, department.ID)

	utils.Created(w, map[string]interface{}{
		"department": createdDept,
		"message":    "Department created successfully",
	})
}

// Update updates an existing department
// PUT /departments/{id}
func (h *DepartmentHandler) Update(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "id")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid department ID")
		return
	}

	var req models.DepartmentUpdateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get existing department
	department, err := h.departmentRepo.FindByID(r.Context(), tenantID, deptID)
	if err != nil {
		utils.NotFound(w, "Department not found")
		return
	}

	// Update fields
	if req.Name != nil {
		// Check name uniqueness
		exists, err := h.departmentRepo.CheckNameExists(r.Context(), tenantID, *req.Name, &deptID)
		if err != nil {
			utils.InternalServerError(w, "Failed to check department name")
			return
		}
		if exists {
			utils.Conflict(w, "Department name already exists")
			return
		}
		department.Name = *req.Name
	}
	if req.Description != nil {
		department.Description = req.Description
	}
	if req.HeadUserID != nil {
		// Validate head_user_id
		_, err := h.userRepo.FindByID(r.Context(), tenantID, *req.HeadUserID)
		if err != nil {
			utils.BadRequest(w, "Invalid head user ID")
			return
		}
		department.HeadUserID = req.HeadUserID
	}
	if req.Color != nil {
		department.Color = *req.Color
	}
	if req.Icon != nil {
		department.Icon = *req.Icon
	}
	if req.Status != nil {
		department.Status = *req.Status
	}

	// Update department
	if err := h.departmentRepo.Update(r.Context(), tenantID, department); err != nil {
		utils.InternalServerError(w, "Failed to update department")
		return
	}

	// Get updated department with details
	updatedDept, _ := h.departmentRepo.GetWithDetails(r.Context(), tenantID, deptID)

	utils.Success(w, map[string]interface{}{
		"department": updatedDept,
		"message":    "Department updated successfully",
	})
}

// Delete deletes a department
// DELETE /departments/{id}
func (h *DepartmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "id")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid department ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Note: Members will have their department_id set to NULL due to ON DELETE SET NULL
	// Frontend already confirms with user about member count impact

	// Delete department
	if err := h.departmentRepo.Delete(r.Context(), tenantID, deptID); err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Department deleted successfully",
	})
}

// GetMembers retrieves members of a department
// GET /departments/{id}/members
func (h *DepartmentHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "id")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid department ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	members, err := h.departmentRepo.GetMembers(r.Context(), tenantID, deptID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get department members")
		return
	}

	utils.Success(w, map[string]interface{}{
		"members": members,
		"count":   len(members),
	})
}

// RegisterRoutes registers all department routes
func (h *DepartmentHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/departments", func(r chi.Router) {
		// All department routes require authentication
		r.Use(authMiddleware.Authenticate)

		// List departments - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionView)).Get("/", h.List)

		// Get single department - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionView)).Get("/{id}", h.Get)

		// Create department - requires create permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionCreate)).Post("/", h.Create)

		// Update department - requires edit permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionEdit)).Put("/{id}", h.Update)

		// Delete department - requires delete permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionDelete)).Delete("/{id}", h.Delete)

		// Get department members - requires view permission
		r.With(permMiddleware.RequirePermission(models.ResourceDepartments, models.ActionView)).Get("/{id}/members", h.GetMembers)
	})
}

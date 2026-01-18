package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// InvitationHandler handles team invitation endpoints
type InvitationHandler struct {
	invitationService *services.InvitationService
}

// NewInvitationHandler creates a new invitation handler
func NewInvitationHandler(invitationService *services.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
	}
}

// CreateInvitation sends a new team invitation
// POST /api/invitations
func (h *InvitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email   string      `json:"email"`
		RoleIDs []uuid.UUID `json:"role_ids"`
		Message string      `json:"message"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("email", req.Email, "Email", &errors)
	utils.ValidateEmail("email", req.Email, &errors)

	if len(req.RoleIDs) == 0 {
		errors.Add("role_ids", "At least one role is required")
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

	// Create invitation
	invitation, err := h.invitationService.CreateInvitation(
		r.Context(),
		tenantID,
		userID,
		req.Email,
		req.RoleIDs,
		req.Message,
	)
	if err != nil {
		if err.Error() == "user with this email already exists" {
			utils.Conflict(w, err.Error())
			return
		}
		utils.InternalServerError(w, "Failed to create invitation")
		return
	}

	utils.Created(w, map[string]interface{}{
		"invitation": invitation,
		"message":    "Invitation sent successfully",
	})
}

// AcceptInvitation accepts an invitation and creates user account
// POST /api/invitations/accept
func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token     string `json:"token"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	// Validate
	errors := utils.ValidationErrors{}
	utils.ValidateRequired("token", req.Token, "Token", &errors)
	utils.ValidateRequired("password", req.Password, "Password", &errors)
	utils.ValidatePassword("password", req.Password, &errors)
	utils.ValidateRequired("first_name", req.FirstName, "First name", &errors)
	utils.ValidateRequired("last_name", req.LastName, "Last name", &errors)
	utils.ValidateName("first_name", req.FirstName, "First name", &errors)
	utils.ValidateName("last_name", req.LastName, "Last name", &errors)

	if errors.HasErrors() {
		utils.UnprocessableEntity(w, "Validation failed", errors.ToMap())
		return
	}

	// Accept invitation
	user, err := h.invitationService.AcceptInvitation(
		r.Context(),
		req.Token,
		req.Password,
		req.FirstName,
		req.LastName,
	)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"user":    user,
		"message": "Invitation accepted successfully. You can now log in.",
	})
}

// ListInvitations lists all invitations with filtering
// GET /api/invitations?status=pending&page=1&page_size=20
func (h *InvitationHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get query parameters
	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// List invitations
	invitations, totalCount, err := h.invitationService.ListInvitations(
		r.Context(),
		tenantID,
		status,
		pageSize,
		offset,
	)
	if err != nil {
		utils.InternalServerError(w, "Failed to list invitations")
		return
	}

	meta := utils.NewMeta(page, pageSize, totalCount)

	utils.SuccessWithMeta(w, map[string]interface{}{
		"invitations": invitations,
	}, meta)
}

// GetInvitation retrieves a single invitation
// GET /api/invitations/{id}
func (h *InvitationHandler) GetInvitation(w http.ResponseWriter, r *http.Request) {
	invitationIDStr := chi.URLParam(r, "id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid invitation ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	invitation, err := h.invitationService.GetInvitation(r.Context(), tenantID, invitationID)
	if err != nil {
		utils.NotFound(w, "Invitation not found")
		return
	}

	utils.Success(w, map[string]interface{}{
		"invitation": invitation,
	})
}

// RevokeInvitation revokes a pending invitation
// DELETE /api/invitations/{id}
func (h *InvitationHandler) RevokeInvitation(w http.ResponseWriter, r *http.Request) {
	invitationIDStr := chi.URLParam(r, "id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid invitation ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	err = h.invitationService.RevokeInvitation(r.Context(), tenantID, invitationID)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Invitation revoked successfully",
	})
}

// ResendInvitation resends an invitation email
// POST /api/invitations/{id}/resend
func (h *InvitationHandler) ResendInvitation(w http.ResponseWriter, r *http.Request) {
	invitationIDStr := chi.URLParam(r, "id")
	invitationID, err := uuid.Parse(invitationIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid invitation ID")
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

	err = h.invitationService.ResendInvitation(r.Context(), tenantID, invitationID, userID)
	if err != nil {
		utils.BadRequest(w, err.Error())
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Invitation resent successfully",
	})
}

// RegisterRoutes registers all invitation routes
func (h *InvitationHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/invitations", func(r chi.Router) {
		// Accept invitation (public endpoint - no auth required)
		r.Post("/accept", h.AcceptInvitation)

		// All other routes require authentication
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// List invitations - requires view permission
			r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/", h.ListInvitations)

			// Get single invitation - requires view permission
			r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionView)).Get("/{id}", h.GetInvitation)

			// Create invitation - requires create permission
			r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionCreate)).Post("/", h.CreateInvitation)

			// Revoke invitation - requires delete permission
			r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionDelete)).Delete("/{id}", h.RevokeInvitation)

			// Resend invitation - requires create permission
			r.With(permMiddleware.RequirePermission(models.ResourceUsers, models.ActionCreate)).Post("/{id}/resend", h.ResendInvitation)
		})
	})
}

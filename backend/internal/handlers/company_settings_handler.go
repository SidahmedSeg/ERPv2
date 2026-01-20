package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// CompanySettingsHandler handles company settings endpoints
type CompanySettingsHandler struct {
	service *services.CompanySettingsService
}

// NewCompanySettingsHandler creates a new company settings handler
func NewCompanySettingsHandler(service *services.CompanySettingsService) *CompanySettingsHandler {
	return &CompanySettingsHandler{service: service}
}

// GetSettings returns company settings for the authenticated tenant
// GET /api/settings/company
func (h *CompanySettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	settings, err := h.service.GetSettings(r.Context(), tenantID)
	if err != nil {
		utils.InternalServerError(w, "Failed to fetch settings")
		return
	}

	utils.Success(w, settings)
}

// UpdateSettings updates company settings (partial update)
// PUT /api/settings/company
func (h *CompanySettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
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

	var req models.CompanySettingsUpdateRequest
	if err := utils.ParseJSONBody(r, &req); err != nil {
		utils.BadRequest(w, "Invalid request body")
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), tenantID, userID, &req)
	if err != nil {
		utils.InternalServerError(w, "Failed to update settings")
		return
	}

	utils.Success(w, settings)
}

// RegisterRoutes registers company settings routes
func (h *CompanySettingsHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/settings/company", func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/", h.GetSettings)
		r.Put("/", h.UpdateSettings)
	})
}

package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// AuditHandler handles audit log endpoints
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// ListAuditLogs retrieves audit logs with filtering
// GET /api/audit-logs?user_id=xxx&action=login&status=success&start_date=...&end_date=...&page=1&page_size=20
func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Parse query parameters
	filters := services.AuditFilters{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filters.UserID = &userID
		}
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filters.Action = action
	}

	if resourceType := r.URL.Query().Get("resource_type"); resourceType != "" {
		filters.ResourceType = resourceType
	}

	if resourceIDStr := r.URL.Query().Get("resource_id"); resourceIDStr != "" {
		resourceID, err := uuid.Parse(resourceIDStr)
		if err == nil {
			filters.ResourceID = &resourceID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		filters.Status = status
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err == nil {
			filters.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err == nil {
			filters.EndDate = &endDate
		}
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Query audit logs
	logs, totalCount, err := h.auditService.Query(r.Context(), tenantID, filters, pageSize, offset)
	if err != nil {
		utils.InternalServerError(w, "Failed to query audit logs")
		return
	}

	meta := utils.NewMeta(page, pageSize, totalCount)

	utils.SuccessWithMeta(w, map[string]interface{}{
		"logs": logs,
	}, meta)
}

// GetUserActivity retrieves recent activity for a specific user
// GET /api/audit-logs/user/{user_id}?limit=50
func (h *AuditHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "user_id")
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

	// Get limit
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 200 {
				limit = parsedLimit
			}
		}
	}

	logs, err := h.auditService.GetUserActivity(r.Context(), tenantID, userID, limit)
	if err != nil {
		utils.InternalServerError(w, "Failed to get user activity")
		return
	}

	utils.Success(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// GetActionStats retrieves statistics by action type
// GET /api/audit-logs/stats?start_date=...&end_date=...
func (h *AuditHandler) GetActionStats(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Parse date range (default to last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = parsed
		}
	}

	stats, err := h.auditService.GetActionStats(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		utils.InternalServerError(w, "Failed to get action stats")
		return
	}

	utils.Success(w, map[string]interface{}{
		"stats":      stats,
		"start_date": startDate,
		"end_date":   endDate,
	})
}

// GetFailedAttempts retrieves failed login/auth attempts
// GET /api/audit-logs/failed-attempts?user_id=xxx&since=...
func (h *AuditHandler) GetFailedAttempts(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	var userID *uuid.UUID
	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err == nil {
			userID = &parsed
		}
	}

	// Default to last 24 hours
	since := time.Now().Add(-24 * time.Hour)
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsed
		}
	}

	logs, err := h.auditService.GetFailedAttempts(r.Context(), tenantID, userID, since)
	if err != nil {
		utils.InternalServerError(w, "Failed to get failed attempts")
		return
	}

	utils.Success(w, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
		"since": since,
	})
}

// GetResourceActivity retrieves recent actions for a specific resource
// GET /api/audit-logs/resource/{resource_type}/{resource_id}?limit=50
func (h *AuditHandler) GetResourceActivity(w http.ResponseWriter, r *http.Request) {
	resourceType := chi.URLParam(r, "resource_type")
	resourceIDStr := chi.URLParam(r, "resource_id")

	resourceID, err := uuid.Parse(resourceIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid resource ID")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Get limit
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 200 {
				limit = parsedLimit
			}
		}
	}

	logs, err := h.auditService.GetRecentActions(r.Context(), tenantID, resourceType, resourceID, limit)
	if err != nil {
		utils.InternalServerError(w, "Failed to get resource activity")
		return
	}

	utils.Success(w, map[string]interface{}{
		"logs":          logs,
		"count":         len(logs),
		"resource_type": resourceType,
		"resource_id":   resourceID,
	})
}

// Search searches audit logs by keyword
// GET /api/audit-logs/search?q=keyword&page=1&page_size=20
func (h *AuditHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("q")
	if keyword == "" {
		utils.BadRequest(w, "Search keyword is required")
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	logs, totalCount, err := h.auditService.Search(r.Context(), tenantID, keyword, pageSize, offset)
	if err != nil {
		utils.InternalServerError(w, "Failed to search audit logs")
		return
	}

	meta := utils.NewMeta(page, pageSize, totalCount)

	utils.SuccessWithMeta(w, map[string]interface{}{
		"logs":  logs,
		"query": keyword,
	}, meta)
}

// RegisterRoutes registers all audit log routes
func (h *AuditHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/audit-logs", func(r chi.Router) {
		// All audit log routes require authentication
		r.Use(authMiddleware.Authenticate)

		// Require security.view_logs permission for all audit log endpoints
		r.Use(func(next http.Handler) http.Handler {
			return permMiddleware.RequirePermission(models.ResourceSecurity, models.ActionViewLogs)(next)
		})

		// List audit logs with filters
		r.Get("/", h.ListAuditLogs)

		// Search audit logs
		r.Get("/search", h.Search)

		// Action statistics
		r.Get("/stats", h.GetActionStats)

		// Failed attempts
		r.Get("/failed-attempts", h.GetFailedAttempts)

		// User activity
		r.Get("/user/{user_id}", h.GetUserActivity)

		// Resource activity
		r.Get("/resource/{resource_type}/{resource_id}", h.GetResourceActivity)
	})
}

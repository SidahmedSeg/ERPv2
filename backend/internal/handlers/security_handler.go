package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/models"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// SecurityHandler handles security monitoring endpoints
type SecurityHandler struct {
	auditService   *services.AuditService
	sessionService *services.SessionService
	twoFactorService *services.TwoFactorService
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(
	auditService *services.AuditService,
	sessionService *services.SessionService,
	twoFactorService *services.TwoFactorService,
) *SecurityHandler {
	return &SecurityHandler{
		auditService:   auditService,
		sessionService: sessionService,
		twoFactorService: twoFactorService,
	}
}

// GetSecurityOverview returns a security dashboard overview
// GET /api/security/overview
func (h *SecurityHandler) GetSecurityOverview(w http.ResponseWriter, r *http.Request) {
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

	// Get session stats
	sessionStats, err := h.sessionService.GetSessionStats(r.Context(), tenantID, userID)
	if err != nil {
		sessionStats = map[string]interface{}{
			"active_sessions": 0,
		}
	}

	// Get recent login attempts (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	failedAttempts, err := h.auditService.GetFailedAttempts(r.Context(), tenantID, &userID, since)
	if err != nil {
		failedAttempts = []models.AuditLog{}
	}

	// Get recent activity (last 20 events)
	recentActivity, err := h.auditService.GetUserActivity(r.Context(), tenantID, userID, 20)
	if err != nil {
		recentActivity = []models.AuditLog{}
	}

	// Check 2FA status
	// This would normally come from user repo, but for now we'll skip it

	// Get action stats for last 7 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7)
	actionStats, err := h.auditService.GetActionStats(r.Context(), tenantID, startDate, endDate)
	if err != nil {
		actionStats = map[string]int{}
	}

	overview := map[string]interface{}{
		"session_stats":       sessionStats,
		"failed_attempts":     len(failedAttempts),
		"recent_activity":     recentActivity,
		"action_stats":        actionStats,
		"last_24h_failures":   failedAttempts,
	}

	utils.Success(w, overview)
}

// GetSuspiciousActivity returns potentially suspicious activity patterns
// GET /api/security/suspicious-activity?since=...
func (h *SecurityHandler) GetSuspiciousActivity(w http.ResponseWriter, r *http.Request) {
	tenantID, err := middleware.GetTenantIDFromContext(r.Context())
	if err != nil {
		utils.Unauthorized(w, "Authentication required")
		return
	}

	// Parse since parameter (default to last 7 days)
	since := time.Now().AddDate(0, 0, -7)
	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsed
		}
	}

	// Get suspicious activity patterns
	activities, err := h.auditService.GetSuspiciousActivity(r.Context(), tenantID, since)
	if err != nil {
		utils.InternalServerError(w, "Failed to get suspicious activity")
		return
	}

	utils.Success(w, map[string]interface{}{
		"activities": activities,
		"count":      len(activities),
		"since":      since,
	})
}

// GetSecurityRecommendations returns security recommendations for the user
// GET /api/security/recommendations
func (h *SecurityHandler) GetSecurityRecommendations(w http.ResponseWriter, r *http.Request) {
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

	recommendations := []map[string]interface{}{}

	// Check if 2FA is enabled (would need user repo)
	// For now, we'll simulate this check
	twoFactorEnabled := false // This would come from user data

	if !twoFactorEnabled {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "enable_2fa",
			"severity":    "high",
			"title":       "Enable Two-Factor Authentication",
			"description": "Protect your account with an extra layer of security by enabling 2FA.",
			"action":      "/api/2fa/setup",
		})
	}

	// Check for old sessions (more than 30 days)
	sessionStats, _ := h.sessionService.GetSessionStats(r.Context(), tenantID, userID)
	if activeCount, ok := sessionStats["active_sessions"].(int); ok && activeCount > 5 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "review_sessions",
			"severity":    "medium",
			"title":       "Review Active Sessions",
			"description": "You have multiple active sessions. Review and revoke any unfamiliar devices.",
			"action":      "/api/sessions",
		})
	}

	// Check for recent failed login attempts
	since := time.Now().Add(-7 * 24 * time.Hour)
	failedAttempts, _ := h.auditService.GetFailedAttempts(r.Context(), tenantID, &userID, since)
	if len(failedAttempts) > 3 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "failed_logins",
			"severity":    "high",
			"title":       "Multiple Failed Login Attempts Detected",
			"description": "There have been multiple failed login attempts on your account. Consider changing your password.",
			"action":      "/api/auth/change-password",
			"count":       len(failedAttempts),
		})
	}

	// If no issues, add a positive message
	if len(recommendations) == 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "all_good",
			"severity":    "info",
			"title":       "Security Status: Good",
			"description": "No security issues detected. Keep following best practices!",
		})
	}

	utils.Success(w, map[string]interface{}{
		"recommendations": recommendations,
		"count":           len(recommendations),
	})
}

// GetLoginHistory returns login history for the current user
// GET /api/security/login-history?limit=50
func (h *SecurityHandler) GetLoginHistory(w http.ResponseWriter, r *http.Request) {
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

	// Get limit from query
	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		// Parse limit
	}

	// Get recent logins
	logins, err := h.sessionService.GetRecentLogins(r.Context(), tenantID, userID, limit)
	if err != nil {
		utils.InternalServerError(w, "Failed to get login history")
		return
	}

	utils.Success(w, map[string]interface{}{
		"logins": logins,
		"count":  len(logins),
	})
}

// RegisterRoutes registers all security routes
func (h *SecurityHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
	r.Route("/security", func(r chi.Router) {
		// All security routes require authentication
		r.Use(authMiddleware.Authenticate)

		// Security overview - available to all authenticated users
		r.Get("/overview", h.GetSecurityOverview)

		// Recommendations - available to all authenticated users
		r.Get("/recommendations", h.GetSecurityRecommendations)

		// Login history - available to all authenticated users
		r.Get("/login-history", h.GetLoginHistory)

		// Suspicious activity - requires security.view_logs permission
		r.With(permMiddleware.RequirePermission(models.ResourceSecurity, models.ActionViewLogs)).Get("/suspicious-activity", h.GetSuspiciousActivity)
	})
}

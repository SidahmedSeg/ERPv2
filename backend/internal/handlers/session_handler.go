package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"myerp-v2/internal/middleware"
	"myerp-v2/internal/services"
	"myerp-v2/internal/utils"
)

// SessionHandler handles session management endpoints
type SessionHandler struct {
	sessionService *services.SessionService
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionService *services.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// ListSessions retrieves all active sessions for the current user
// GET /api/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
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

	// Get current token hash from Authorization header
	authHeader := r.Header.Get("Authorization")
	var currentTokenHash string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]
		hash := sha256.Sum256([]byte(token))
		currentTokenHash = hex.EncodeToString(hash[:])
	}

	// List sessions
	sessions, err := h.sessionService.ListUserSessions(r.Context(), tenantID, userID, currentTokenHash)
	if err != nil {
		utils.InternalServerError(w, "Failed to list sessions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeSession revokes a specific session
// DELETE /api/sessions/{id}
func (h *SessionHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	sessionIDStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		utils.BadRequest(w, "Invalid session ID")
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

	// Revoke session
	err = h.sessionService.RevokeSession(r.Context(), tenantID, userID, sessionID)
	if err != nil {
		if err.Error() == "session not found or already revoked" {
			utils.NotFound(w, "Session not found")
			return
		}
		utils.InternalServerError(w, "Failed to revoke session")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions except the current one
// POST /api/sessions/revoke-all
func (h *SessionHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
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

	// Get current token hash
	authHeader := r.Header.Get("Authorization")
	var currentTokenHash string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token := authHeader[7:]
		hash := sha256.Sum256([]byte(token))
		currentTokenHash = hex.EncodeToString(hash[:])
	}

	// Revoke all sessions except current
	revokedCount, err := h.sessionService.RevokeAllSessions(r.Context(), tenantID, userID, currentTokenHash)
	if err != nil {
		utils.InternalServerError(w, "Failed to revoke sessions")
		return
	}

	utils.Success(w, map[string]interface{}{
		"message":       "All other sessions revoked successfully",
		"revoked_count": revokedCount,
	})
}

// GetSessionStats returns session statistics for the current user
// GET /api/sessions/stats
func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
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

	// Get stats
	stats, err := h.sessionService.GetSessionStats(r.Context(), tenantID, userID)
	if err != nil {
		utils.InternalServerError(w, "Failed to get session stats")
		return
	}

	utils.Success(w, stats)
}

// GetRecentLogins returns recent login attempts
// GET /api/sessions/recent-logins?limit=10
func (h *SessionHandler) GetRecentLogins(w http.ResponseWriter, r *http.Request) {
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

	// Get limit from query param (default 10, max 50)
	limit := 10
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 50 {
				limit = parsedLimit
			}
		}
	}

	// Get recent logins
	logins, err := h.sessionService.GetRecentLogins(r.Context(), tenantID, userID, limit)
	if err != nil {
		utils.InternalServerError(w, "Failed to get recent logins")
		return
	}

	utils.Success(w, map[string]interface{}{
		"logins": logins,
		"count":  len(logins),
	})
}

// RegisterRoutes registers all session management routes
func (h *SessionHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware) {
	r.Route("/sessions", func(r chi.Router) {
		// All session routes require authentication
		r.Use(authMiddleware.Authenticate)

		// List active sessions
		r.Get("/", h.ListSessions)

		// Session statistics
		r.Get("/stats", h.GetSessionStats)

		// Recent logins
		r.Get("/recent-logins", h.GetRecentLogins)

		// Revoke specific session
		r.Delete("/{id}", h.RevokeSession)

		// Revoke all other sessions
		r.Post("/revoke-all", h.RevokeAllSessions)
	})
}

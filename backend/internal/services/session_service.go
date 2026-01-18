package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// SessionService handles session management operations
type SessionService struct {
	db *sqlx.DB
}

// NewSessionService creates a new session service
func NewSessionService(db *sqlx.DB) *SessionService {
	return &SessionService{
		db: db,
	}
}

// SessionInfo contains detailed session information
type SessionInfo struct {
	models.Session
	IsCurrent bool `json:"is_current" db:"-"`
}

// ListUserSessions retrieves all active sessions for a user
func (s *SessionService) ListUserSessions(ctx context.Context, tenantID, userID uuid.UUID, currentTokenHash string) ([]SessionInfo, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id,
			tenant_id,
			user_id,
			token_hash,
			device_type,
			browser,
			os,
			ip_address,
			user_agent,
			country_code,
			city,
			last_activity_at,
			expires_at,
			created_at
		FROM sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`

	var sessions []models.Session
	err = tx.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	// Mark current session
	sessionInfos := make([]SessionInfo, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = SessionInfo{
			Session:   session,
			IsCurrent: session.TokenHash == currentTokenHash,
		}
	}

	return sessionInfos, tx.Commit()
}

// RevokeSession revokes a specific session
func (s *SessionService) RevokeSession(ctx context.Context, tenantID, userID, sessionID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM sessions
		WHERE id = $1
		  AND user_id = $2
	`

	result, err := tx.ExecContext(ctx, query, sessionID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already revoked")
	}

	return tx.Commit()
}

// RevokeAllSessions revokes all sessions except the current one
func (s *SessionService) RevokeAllSessions(ctx context.Context, tenantID, userID uuid.UUID, exceptTokenHash string) (int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM sessions
		WHERE user_id = $1
		  AND token_hash != $2
	`

	result, err := tx.ExecContext(ctx, query, userID, exceptTokenHash)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return int(rowsAffected), tx.Commit()
}

// RevokeAllUserSessions revokes all sessions for a user (admin operation)
func (s *SessionService) RevokeAllUserSessions(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `DELETE FROM sessions WHERE user_id = $1`

	result, err := tx.ExecContext(ctx, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to revoke all sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return int(rowsAffected), tx.Commit()
}

// GetSessionStats returns session statistics for a user
func (s *SessionService) GetSessionStats(ctx context.Context, tenantID, userID uuid.UUID) (map[string]interface{}, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Active sessions count
	var activeCount int
	countQuery := `
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
	`
	err = tx.GetContext(ctx, &activeCount, countQuery, userID)
	if err != nil {
		return nil, err
	}

	// Sessions by device type
	var deviceStats []struct {
		DeviceType string `db:"device_type"`
		Count      int    `db:"count"`
	}
	deviceQuery := `
		SELECT
			COALESCE(device_type, 'Unknown') as device_type,
			COUNT(*) as count
		FROM sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
		GROUP BY device_type
	`
	err = tx.SelectContext(ctx, &deviceStats, deviceQuery, userID)
	if err != nil {
		return nil, err
	}

	// Last login info
	var lastLogin struct {
		LastActivityAt *time.Time `db:"last_activity_at"`
		IPAddress      *string    `db:"ip_address"`
		DeviceType     *string    `db:"device_type"`
		Browser        *string    `db:"browser"`
	}
	lastLoginQuery := `
		SELECT
			last_activity_at,
			ip_address,
			device_type,
			browser
		FROM sessions
		WHERE user_id = $1
		ORDER BY last_activity_at DESC
		LIMIT 1
	`
	err = tx.GetContext(ctx, &lastLogin, lastLoginQuery, userID)
	if err != nil {
		// No sessions found is not an error
		lastLogin.LastActivityAt = nil
	}

	stats := map[string]interface{}{
		"active_sessions":   activeCount,
		"device_breakdown":  deviceStats,
		"last_activity_at":  lastLogin.LastActivityAt,
		"last_ip_address":   lastLogin.IPAddress,
		"last_device_type":  lastLogin.DeviceType,
		"last_browser":      lastLogin.Browser,
	}

	return stats, tx.Commit()
}

// CleanupExpiredSessions removes expired sessions (should be run periodically)
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int, error) {
	// Use bypass RLS for cleanup task (system operation)
	tx, err := database.WithBypassRLS(ctx, s.db)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		DELETE FROM sessions
		WHERE expires_at < NOW()
	`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	return int(rowsAffected), tx.Commit()
}

// GetSessionByToken retrieves a session by token hash
func (s *SessionService) GetSessionByToken(ctx context.Context, tokenHash string) (*models.Session, error) {
	// Use bypass RLS for token validation (can't know tenant yet)
	tx, err := database.WithBypassRLS(ctx, s.db)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var session models.Session
	query := `
		SELECT
			id,
			tenant_id,
			user_id,
			token_hash,
			device_type,
			browser,
			os,
			ip_address,
			user_agent,
			country_code,
			city,
			last_activity_at,
			expires_at,
			created_at
		FROM sessions
		WHERE token_hash = $1
		  AND expires_at > NOW()
		LIMIT 1
	`

	err = tx.GetContext(ctx, &session, query, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	return &session, tx.Commit()
}

// UpdateSessionActivity updates the last activity timestamp
func (s *SessionService) UpdateSessionActivity(ctx context.Context, tenantID, sessionID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE sessions
		SET last_activity_at = NOW()
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}

	return tx.Commit()
}

// GetRecentLogins retrieves recent login attempts (for security monitoring)
func (s *SessionService) GetRecentLogins(ctx context.Context, tenantID, userID uuid.UUID, limit int) ([]models.Session, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id,
			tenant_id,
			user_id,
			token_hash,
			device_type,
			browser,
			os,
			ip_address,
			user_agent,
			country_code,
			city,
			last_activity_at,
			expires_at,
			created_at
		FROM sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var sessions []models.Session
	err = tx.SelectContext(ctx, &sessions, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent logins: %w", err)
	}

	return sessions, tx.Commit()
}

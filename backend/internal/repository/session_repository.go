package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// SessionRepository handles database operations for sessions
type SessionRepository struct {
	db *sqlx.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create creates a new session with RLS
func (r *SessionRepository) Create(ctx context.Context, req *models.SessionCreateRequest) (*models.Session, error) {
	tx, err := database.WithTenantContext(ctx, r.db, req.TenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	session := &models.Session{
		TenantID:       req.TenantID,
		UserID:         req.UserID,
		TokenHash:      req.Token, // Should be hashed before passing to this function
		DeviceType:     req.DeviceType,
		Browser:        req.Browser,
		OS:             req.OS,
		IPAddress:      &req.IPAddress,
		UserAgent:      req.UserAgent,
		LastActivityAt: time.Now(),
		ExpiresAt:      req.ExpiresAt,
	}

	query := `
		INSERT INTO sessions (
			tenant_id, user_id, token_hash, device_type, browser,
			os, ip_address, user_agent, last_activity_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	err = tx.QueryRowContext(
		ctx, query,
		session.TenantID,
		session.UserID,
		session.TokenHash,
		session.DeviceType,
		session.Browser,
		session.OS,
		session.IPAddress,
		session.UserAgent,
		session.LastActivityAt,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return session, nil
}

// FindByTokenHash retrieves a session by token hash
func (r *SessionRepository) FindByTokenHash(ctx context.Context, tenantID uuid.UUID, tokenHash string) (*models.Session, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var session models.Session
	query := `
		SELECT * FROM sessions
		WHERE token_hash = $1
		  AND expires_at > NOW()
		LIMIT 1
	`

	err = tx.GetContext(ctx, &session, query, tokenHash)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return &session, nil
}

// FindByID retrieves a session by ID
func (r *SessionRepository) FindByID(ctx context.Context, tenantID, sessionID uuid.UUID) (*models.Session, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var session models.Session
	query := `SELECT * FROM sessions WHERE id = $1 LIMIT 1`

	err = tx.GetContext(ctx, &session, query, sessionID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return &session, nil
}

// ListByUser retrieves all sessions for a user
func (r *SessionRepository) ListByUser(ctx context.Context, tenantID, userID uuid.UUID) ([]models.Session, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var sessions []models.Session
	query := `
		SELECT * FROM sessions
		WHERE user_id = $1
		ORDER BY last_activity_at DESC
	`

	err = tx.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

// ListActiveSessions retrieves all active (non-expired) sessions for a user
func (r *SessionRepository) ListActiveSessions(ctx context.Context, tenantID, userID uuid.UUID) ([]models.Session, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var sessions []models.Session
	query := `
		SELECT * FROM sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
		ORDER BY last_activity_at DESC
	`

	err = tx.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active sessions: %w", err)
	}

	return sessions, nil
}

// UpdateActivity updates the last activity time of a session
func (r *SessionRepository) UpdateActivity(ctx context.Context, tenantID, sessionID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
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
		return fmt.Errorf("failed to update activity: %w", err)
	}

	return tx.Commit()
}

// Delete deletes a specific session (logout)
func (r *SessionRepository) Delete(ctx context.Context, tenantID, sessionID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM sessions WHERE id = $1`

	result, err := tx.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return tx.Commit()
}

// DeleteByTokenHash deletes a session by token hash
func (r *SessionRepository) DeleteByTokenHash(ctx context.Context, tenantID uuid.UUID, tokenHash string) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM sessions WHERE token_hash = $1`

	_, err = tx.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return tx.Commit()
}

// DeleteAllByUser deletes all sessions for a user (logout from all devices)
func (r *SessionRepository) DeleteAllByUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err = tx.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all sessions: %w", err)
	}

	return tx.Commit()
}

// DeleteExpiredSessions deletes all expired sessions
func (r *SessionRepository) DeleteExpiredSessions(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := tx.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// DeleteInactiveSessions deletes sessions that have been inactive for too long
func (r *SessionRepository) DeleteInactiveSessions(ctx context.Context, tenantID uuid.UUID, inactivityDuration time.Duration) (int64, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	inactivityThreshold := time.Now().Add(-inactivityDuration)

	query := `DELETE FROM sessions WHERE last_activity_at < $1`

	result, err := tx.ExecContext(ctx, query, inactivityThreshold)
	if err != nil {
		return 0, fmt.Errorf("failed to delete inactive sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// CountActiveSessions counts active sessions for a user
func (r *SessionRepository) CountActiveSessions(ctx context.Context, tenantID, userID uuid.UUID) (int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var count int
	query := `
		SELECT COUNT(*) FROM sessions
		WHERE user_id = $1
		  AND expires_at > NOW()
	`

	err = tx.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count active sessions: %w", err)
	}

	return count, nil
}

// GetSessionStats retrieves session statistics for a tenant
func (r *SessionRepository) GetSessionStats(ctx context.Context, tenantID uuid.UUID) (map[string]int, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stats := make(map[string]int)

	// Total sessions
	var totalSessions int
	err = tx.GetContext(ctx, &totalSessions, `SELECT COUNT(*) FROM sessions`)
	if err != nil {
		return nil, fmt.Errorf("failed to count total sessions: %w", err)
	}
	stats["total"] = totalSessions

	// Active sessions
	var activeSessions int
	err = tx.GetContext(ctx, &activeSessions, `SELECT COUNT(*) FROM sessions WHERE expires_at > NOW()`)
	if err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}
	stats["active"] = activeSessions

	// Expired sessions
	stats["expired"] = totalSessions - activeSessions

	return stats, nil
}

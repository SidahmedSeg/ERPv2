package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// AuditService handles security audit logging and querying
type AuditService struct {
	db *sqlx.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *sqlx.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// AuditFilters contains filtering options for audit log queries
type AuditFilters struct {
	UserID       *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	Status       string
	StartDate    *time.Time
	EndDate      *time.Time
}

// LogEvent creates an audit log entry
func (s *AuditService) LogEvent(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	action, resourceType string,
	resourceID uuid.UUID,
	status string,
	ipAddress, userAgent string,
	metadata map[string]interface{},
) error {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO audit_logs (
			tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = tx.ExecContext(ctx, query,
		tenantID, userID, action, resourceType, resourceID,
		status, ipAddress, userAgent, metadataJSON,
	)
	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return tx.Commit()
}

// Query retrieves audit logs with filters
func (s *AuditService) Query(
	ctx context.Context,
	tenantID uuid.UUID,
	filters AuditFilters,
	limit, offset int,
) ([]models.AuditLog, int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	// Build query
	baseQuery := `FROM audit_logs WHERE 1=1`
	var args []interface{}
	argIndex := 1

	// Apply filters
	if filters.UserID != nil {
		baseQuery += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.Action != "" {
		baseQuery += fmt.Sprintf(" AND action = $%d", argIndex)
		args = append(args, filters.Action)
		argIndex++
	}

	if filters.ResourceType != "" {
		baseQuery += fmt.Sprintf(" AND resource_type = $%d", argIndex)
		args = append(args, filters.ResourceType)
		argIndex++
	}

	if filters.ResourceID != nil {
		baseQuery += fmt.Sprintf(" AND resource_id = $%d", argIndex)
		args = append(args, *filters.ResourceID)
		argIndex++
	}

	if filters.Status != "" {
		baseQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filters.Status)
		argIndex++
	}

	if filters.StartDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filters.StartDate)
		argIndex++
	}

	if filters.EndDate != nil {
		baseQuery += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filters.EndDate)
		argIndex++
	}

	// Get total count
	var totalCount int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err = tx.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	// Get audit logs
	selectQuery := fmt.Sprintf(`
		SELECT
			id, tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata, created_at
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIndex, argIndex+1)

	args = append(args, limit, offset)

	var logs []models.AuditLog
	err = tx.SelectContext(ctx, &logs, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, tx.Commit()
}

// GetUserActivity retrieves recent activity for a specific user
func (s *AuditService) GetUserActivity(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	limit int,
) ([]models.AuditLog, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id, tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	var logs []models.AuditLog
	err = tx.SelectContext(ctx, &logs, query, userID, limit)
	if err != nil {
		return nil, err
	}

	return logs, tx.Commit()
}

// GetActionStats retrieves statistics by action type
func (s *AuditService) GetActionStats(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT action, COUNT(*) as count
		FROM audit_logs
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY action
		ORDER BY count DESC
	`

	type ActionStat struct {
		Action string `db:"action"`
		Count  int    `db:"count"`
	}

	var stats []ActionStat
	err = tx.SelectContext(ctx, &stats, query, startDate, endDate)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, stat := range stats {
		result[stat.Action] = stat.Count
	}

	return result, tx.Commit()
}

// GetFailedAttempts retrieves failed login/auth attempts
func (s *AuditService) GetFailedAttempts(
	ctx context.Context,
	tenantID uuid.UUID,
	userID *uuid.UUID,
	since time.Time,
) ([]models.AuditLog, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	baseQuery := `
		SELECT
			id, tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE status = 'failure'
		  AND created_at >= $1
		  AND action IN ('login', '2fa.verify', 'password.reset')
	`

	var args []interface{}
	args = append(args, since)

	if userID != nil {
		baseQuery += " AND user_id = $2"
		args = append(args, *userID)
	}

	baseQuery += " ORDER BY created_at DESC LIMIT 100"

	var logs []models.AuditLog
	err = tx.SelectContext(ctx, &logs, baseQuery, args...)
	if err != nil {
		return nil, err
	}

	return logs, tx.Commit()
}

// GetSuspiciousActivity detects potentially suspicious activity patterns
func (s *AuditService) GetSuspiciousActivity(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]map[string]interface{}, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Multiple failed login attempts from same IP
	failedLoginsQuery := `
		SELECT
			ip_address,
			user_id,
			COUNT(*) as attempt_count,
			MAX(created_at) as last_attempt
		FROM audit_logs
		WHERE status = 'failure'
		  AND action = 'login'
		  AND created_at >= $1
		GROUP BY ip_address, user_id
		HAVING COUNT(*) >= 3
		ORDER BY attempt_count DESC
		LIMIT 20
	`

	type FailedLogin struct {
		IPAddress    string     `db:"ip_address"`
		UserID       *uuid.UUID `db:"user_id"`
		AttemptCount int        `db:"attempt_count"`
		LastAttempt  time.Time  `db:"last_attempt"`
	}

	var failedLogins []FailedLogin
	err = tx.SelectContext(ctx, &failedLogins, failedLoginsQuery, since)
	if err != nil {
		return nil, err
	}

	// Unusual activity patterns (e.g., multiple locations in short time)
	locationChangesQuery := `
		WITH location_changes AS (
			SELECT
				user_id,
				country_code,
				LAG(country_code) OVER (PARTITION BY user_id ORDER BY created_at) as prev_country,
				created_at,
				LAG(created_at) OVER (PARTITION BY user_id ORDER BY created_at) as prev_created_at
			FROM audit_logs
			WHERE created_at >= $1
			  AND country_code IS NOT NULL
		)
		SELECT
			user_id,
			country_code,
			prev_country,
			created_at,
			prev_created_at,
			EXTRACT(EPOCH FROM (created_at - prev_created_at)) as time_diff_seconds
		FROM location_changes
		WHERE prev_country IS NOT NULL
		  AND country_code != prev_country
		  AND EXTRACT(EPOCH FROM (created_at - prev_created_at)) < 3600
		ORDER BY created_at DESC
		LIMIT 20
	`

	type LocationChange struct {
		UserID           uuid.UUID  `db:"user_id"`
		CountryCode      string     `db:"country_code"`
		PrevCountry      string     `db:"prev_country"`
		CreatedAt        time.Time  `db:"created_at"`
		PrevCreatedAt    *time.Time `db:"prev_created_at"`
		TimeDiffSeconds  float64    `db:"time_diff_seconds"`
	}

	var locationChanges []LocationChange
	err = tx.SelectContext(ctx, &locationChanges, locationChangesQuery, since)
	if err != nil {
		// Non-critical error, continue
		locationChanges = []LocationChange{}
	}

	// Compile results
	results := make([]map[string]interface{}, 0)

	for _, fl := range failedLogins {
		results = append(results, map[string]interface{}{
			"type":          "failed_logins",
			"ip_address":    fl.IPAddress,
			"user_id":       fl.UserID,
			"attempt_count": fl.AttemptCount,
			"last_attempt":  fl.LastAttempt,
			"severity":      getSeverityLevel(fl.AttemptCount),
		})
	}

	for _, lc := range locationChanges {
		results = append(results, map[string]interface{}{
			"type":             "unusual_location",
			"user_id":          lc.UserID,
			"country_code":     lc.CountryCode,
			"prev_country":     lc.PrevCountry,
			"time_diff_minutes": lc.TimeDiffSeconds / 60,
			"last_seen":        lc.CreatedAt,
			"severity":         "medium",
		})
	}

	return results, tx.Commit()
}

// GetRecentActions retrieves recent actions for specific resource
func (s *AuditService) GetRecentActions(
	ctx context.Context,
	tenantID uuid.UUID,
	resourceType string,
	resourceID uuid.UUID,
	limit int,
) ([]models.AuditLog, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT
			id, tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE resource_type = $1
		  AND resource_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`

	var logs []models.AuditLog
	err = tx.SelectContext(ctx, &logs, query, resourceType, resourceID, limit)
	if err != nil {
		return nil, err
	}

	return logs, tx.Commit()
}

// Search searches audit logs by keyword in action or metadata
func (s *AuditService) Search(
	ctx context.Context,
	tenantID uuid.UUID,
	keyword string,
	limit, offset int,
) ([]models.AuditLog, int, error) {
	tx, err := database.WithTenantContext(ctx, s.db, tenantID)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	keyword = strings.ToLower(keyword)
	searchPattern := "%" + keyword + "%"

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM audit_logs
		WHERE LOWER(action) LIKE $1
		   OR LOWER(resource_type) LIKE $1
	`

	var totalCount int
	err = tx.GetContext(ctx, &totalCount, countQuery, searchPattern)
	if err != nil {
		return nil, 0, err
	}

	// Get logs
	selectQuery := `
		SELECT
			id, tenant_id, user_id, action, resource_type, resource_id,
			status, ip_address, user_agent, metadata, created_at
		FROM audit_logs
		WHERE LOWER(action) LIKE $1
		   OR LOWER(resource_type) LIKE $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var logs []models.AuditLog
	err = tx.SelectContext(ctx, &logs, selectQuery, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	return logs, totalCount, tx.Commit()
}

// Helper function to determine severity level
func getSeverityLevel(attemptCount int) string {
	if attemptCount >= 10 {
		return "critical"
	} else if attemptCount >= 5 {
		return "high"
	} else if attemptCount >= 3 {
		return "medium"
	}
	return "low"
}

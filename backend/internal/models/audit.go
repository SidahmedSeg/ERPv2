package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents a security audit log entry
type AuditLog struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	TenantID uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID   *uuid.UUID `json:"user_id,omitempty" db:"user_id"` // Can be NULL for system events

	// Event details
	Action       string     `json:"action" db:"action"`               // e.g., user.login, user.created, role.assigned
	ResourceType *string    `json:"resource_type,omitempty" db:"resource_type"` // e.g., user, role, session
	ResourceID   *uuid.UUID `json:"resource_id,omitempty" db:"resource_id"`

	// Request context
	IPAddress *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string `json:"user_agent,omitempty" db:"user_agent"`

	// Status: success | failure
	Status string `json:"status" db:"status"`

	// Additional context (flexible JSONB field)
	Metadata map[string]interface{} `json:"metadata,omitempty" db:"metadata"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Computed fields
	User *User `json:"user,omitempty" db:"-"`
}

// AuditLog action constants
const (
	// Authentication events
	ActionUserLogin        = "user.login"
	ActionUserLoginFailed  = "user.login.failed"
	ActionUserLogout       = "user.logout"
	ActionUserRegistered   = "user.registered"
	ActionUserVerified     = "user.verified"

	// User management events
	ActionUserCreated      = "user.created"
	ActionUserUpdated      = "user.updated"
	ActionUserDeleted      = "user.deleted"
	ActionUserSuspended    = "user.suspended"
	ActionUserActivated    = "user.activated"
	ActionUserPasswordReset = "user.password_reset"
	ActionUserPasswordChanged = "user.password_changed"

	// Role events
	ActionRoleCreated      = "role.created"
	ActionRoleUpdated      = "role.updated"
	ActionRoleDeleted      = "role.deleted"
	ActionRoleAssigned     = "role.assigned"
	ActionRoleUnassigned   = "role.unassigned"

	// Permission events
	ActionPermissionGranted = "permission.granted"
	ActionPermissionRevoked = "permission.revoked"

	// 2FA events
	Action2FAEnabled       = "2fa.enabled"
	Action2FADisabled      = "2fa.disabled"
	Action2FAVerified      = "2fa.verified"
	Action2FAFailed        = "2fa.failed"
	Action2FABackupCodeUsed = "2fa.backup_code_used"

	// Session events
	ActionSessionCreated   = "session.created"
	ActionSessionRevoked   = "session.revoked"
	ActionSessionExpired   = "session.expired"

	// Security events
	ActionUnauthorizedAccess = "security.unauthorized_access"
	ActionRateLimitExceeded  = "security.rate_limit_exceeded"
	ActionSuspiciousActivity = "security.suspicious_activity"
)

// AuditLog status constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailure = "failure"
)

// AuditLogCreateRequest represents a request to create an audit log entry
type AuditLogCreateRequest struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType *string
	ResourceID   *uuid.UUID
	IPAddress    *string
	UserAgent    *string
	Status       string
	Metadata     map[string]interface{}
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	Status       *string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

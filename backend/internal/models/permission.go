package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Resource    string    `json:"resource" db:"resource"` // e.g., users, roles, settings
	Action      string    `json:"action" db:"action"`     // e.g., view, create, edit, delete, *
	DisplayName string    `json:"display_name" db:"display_name"`
	Description *string   `json:"description,omitempty" db:"description"`
	Category    *string   `json:"category,omitempty" db:"category"` // For UI grouping
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Permission categories
const (
	CategoryUserManagement = "User Management"
	CategoryAccessControl  = "Access Control"
	CategorySettings       = "Settings"
	CategorySecurity       = "Security"
)

// Common resources
const (
	ResourceUsers    = "users"
	ResourceRoles    = "roles"
	ResourceSettings = "settings"
	ResourceSecurity = "security"
)

// Common actions
const (
	ActionView          = "view"
	ActionCreate        = "create"
	ActionEdit          = "edit"
	ActionDelete        = "delete"
	ActionAll           = "*"
	ActionManageStatus  = "manage_status"
	ActionAssign        = "assign"
	ActionViewLogs      = "view_logs"
	ActionViewSessions  = "view_sessions"
	ActionManageSessions = "manage_sessions"
)

// String returns the permission in resource.action format
func (p *Permission) String() string {
	return p.Resource + "." + p.Action
}

// IsWildcard returns true if this is a wildcard permission (action = *)
func (p *Permission) IsWildcard() bool {
	return p.Action == ActionAll
}

// Matches checks if this permission matches the given resource and action
func (p *Permission) Matches(resource, action string) bool {
	if p.Resource != resource {
		return false
	}
	return p.Action == action || p.Action == ActionAll
}

// PermissionGroup represents a group of permissions for UI display
type PermissionGroup struct {
	Category    string       `json:"category"`
	Permissions []Permission `json:"permissions"`
}

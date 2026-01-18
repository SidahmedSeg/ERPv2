package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a role with permissions
type Role struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"` // Unique per tenant
	DisplayName string     `json:"display_name" db:"display_name"`
	Description *string    `json:"description,omitempty" db:"description"`

	// Hierarchy support
	ParentRoleID *uuid.UUID `json:"parent_role_id,omitempty" db:"parent_role_id"`
	Level        int        `json:"level" db:"level"` // 0 = root, higher = more restrictive

	// System roles cannot be deleted/modified
	IsSystem bool `json:"is_system" db:"is_system"`

	// Metadata
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`

	// Computed fields
	Permissions     []Permission `json:"permissions,omitempty" db:"-"`
	PermissionCount int          `json:"permission_count,omitempty" db:"-"`
	UserCount       int          `json:"user_count,omitempty" db:"-"`
}

// System role names
const (
	RoleOwner   = "owner"
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
)

// IsOwner returns true if this is the owner role
func (r *Role) IsOwner() bool {
	return r.Name == RoleOwner
}

// IsAdmin returns true if this is the admin role
func (r *Role) IsAdmin() bool {
	return r.Name == RoleAdmin
}

// CanDelete returns true if the role can be deleted
func (r *Role) CanDelete() bool {
	return !r.IsSystem
}

// CanEdit returns true if the role can be edited
func (r *Role) CanEdit() bool {
	return !r.IsSystem
}

// RoleCreateRequest represents a request to create a new role
type RoleCreateRequest struct {
	Name          string      `json:"name" validate:"required,min=2,max=100"`
	DisplayName   string      `json:"display_name" validate:"required,min=2,max=255"`
	Description   string      `json:"description,omitempty"`
	ParentRoleID  *uuid.UUID  `json:"parent_role_id,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids" validate:"required,min=1"`
}

// RoleUpdateRequest represents a request to update a role
type RoleUpdateRequest struct {
	DisplayName   *string     `json:"display_name,omitempty" validate:"omitempty,min=2,max=255"`
	Description   *string     `json:"description,omitempty"`
	ParentRoleID  *uuid.UUID  `json:"parent_role_id,omitempty"`
	PermissionIDs []uuid.UUID `json:"permission_ids,omitempty"`
}

// RoleAssignRequest represents a request to assign roles to a user
type RoleAssignRequest struct {
	UserID  uuid.UUID   `json:"user_id" validate:"required"`
	RoleIDs []uuid.UUID `json:"role_ids" validate:"required,min=1"`
}

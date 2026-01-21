package models

import (
	"time"

	"github.com/google/uuid"
)

// Department represents an organizational department
type Department struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`

	// Basic Info
	Name        string  `json:"name" db:"name"`
	Description *string `json:"description,omitempty" db:"description"`

	// Leadership
	HeadUserID *uuid.UUID `json:"head_user_id,omitempty" db:"head_user_id"`

	// Visual Customization
	Color string `json:"color" db:"color"`
	Icon  string `json:"icon" db:"icon"`

	// Status
	Status string `json:"status" db:"status"`

	// Metadata
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`

	// Computed fields (not in database)
	HeadUserName *string `json:"head_user_name,omitempty" db:"head_user_name"`
	MemberCount  int     `json:"member_count" db:"member_count"`
}

// Department status constants
const (
	DepartmentStatusActive   = "active"
	DepartmentStatusInactive = "inactive"
)

// Permission resource constant
const (
	ResourceDepartments = "departments"
)

// Permission actions for departments
const (
	ActionManageMembers = "manage_members"
)

// IsActive returns true if the department is active
func (d *Department) IsActive() bool {
	return d.Status == DepartmentStatusActive
}

// DepartmentCreateRequest represents a request to create a new department
type DepartmentCreateRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	HeadUserID  *uuid.UUID `json:"head_user_id,omitempty"`
	Color       string     `json:"color"`
	Icon        string     `json:"icon"`
}

// DepartmentUpdateRequest represents a request to update a department
type DepartmentUpdateRequest struct {
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
	HeadUserID  *uuid.UUID `json:"head_user_id,omitempty"`
	Color       *string    `json:"color,omitempty"`
	Icon        *string    `json:"icon,omitempty"`
	Status      *string    `json:"status,omitempty"`
}

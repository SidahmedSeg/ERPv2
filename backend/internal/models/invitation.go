package models

import (
	"time"

	"github.com/google/uuid"
)

// Invitation represents a team invitation
type Invitation struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`

	Email string    `json:"email" db:"email"`
	Token uuid.UUID `json:"-" db:"token"` // Don't expose in JSON

	// Roles to assign upon acceptance
	RoleIDs []uuid.UUID `json:"role_ids" db:"role_ids"`

	// Status: pending | accepted | expired | revoked
	Status string `json:"status" db:"status"`

	// Optional welcome message
	Message *string `json:"message,omitempty" db:"message"`

	// Metadata
	InvitedBy  uuid.UUID  `json:"invited_by" db:"invited_by"`
	InvitedAt  time.Time  `json:"invited_at" db:"invited_at"`
	AcceptedAt *time.Time `json:"accepted_at,omitempty" db:"accepted_at"`
	ExpiresAt  time.Time  `json:"expires_at" db:"expires_at"`

	// Computed fields
	InvitedByUser *User  `json:"invited_by_user,omitempty" db:"-"`
	Roles         []Role `json:"roles,omitempty" db:"-"`
}

// InvitationStatus constants
const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusExpired  = "expired"
	InvitationStatusRevoked  = "revoked"
)

// IsExpired returns true if the invitation has expired
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending returns true if the invitation is pending
func (i *Invitation) IsPending() bool {
	return i.Status == InvitationStatusPending
}

// CanAccept returns true if the invitation can be accepted
func (i *Invitation) CanAccept() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// InvitationCreateRequest represents a request to create an invitation
type InvitationCreateRequest struct {
	Email   string      `json:"email" validate:"required,email"`
	RoleIDs []uuid.UUID `json:"role_ids" validate:"required,min=1"`
	Message string      `json:"message,omitempty" validate:"omitempty,max=500"`
}

// InvitationAcceptRequest represents a request to accept an invitation
type InvitationAcceptRequest struct {
	Token     string `json:"token" validate:"required"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Password  string `json:"password" validate:"required,min=8"`
}

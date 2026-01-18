package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant (company/organization) in the system
type Tenant struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Slug      string    `json:"slug" db:"slug"`
	CompanyName string  `json:"company_name" db:"company_name"`

	// Status: pending_verification | active | suspended | canceled
	Status string `json:"status" db:"status"`

	// Email verification
	Email           string     `json:"email" db:"email"`
	EmailVerified   bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`

	VerificationToken          *uuid.UUID `json:"-" db:"verification_token"`
	VerificationTokenExpiresAt *time.Time `json:"-" db:"verification_token_expires_at"`

	// Subscription
	PlanTier    string     `json:"plan_tier" db:"plan_tier"` // free | starter | professional | enterprise
	TrialEndsAt *time.Time `json:"trial_ends_at,omitempty" db:"trial_ends_at"`

	// Settings (stored as JSONB in database)
	Settings json.RawMessage `json:"settings,omitempty" db:"settings"`

	// Metadata
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty" db:"activated_at"`
	SuspendedAt *time.Time `json:"suspended_at,omitempty" db:"suspended_at"`
}

// TenantStatus constants
const (
	TenantStatusPendingVerification = "pending_verification"
	TenantStatusActive              = "active"
	TenantStatusSuspended           = "suspended"
	TenantStatusCanceled            = "canceled"
)

// PlanTier constants
const (
	PlanTierFree         = "free"
	PlanTierStarter      = "starter"
	PlanTierProfessional = "professional"
	PlanTierEnterprise   = "enterprise"
)

// IsActive returns true if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

// IsPendingVerification returns true if the tenant is pending email verification
func (t *Tenant) IsPendingVerification() bool {
	return t.Status == TenantStatusPendingVerification
}

// IsSuspended returns true if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

// CanAccess returns true if the tenant can access the system
func (t *Tenant) CanAccess() bool {
	return t.Status == TenantStatusActive
}

// TenantCreateRequest represents a request to create a new tenant
type TenantCreateRequest struct {
	CompanyName string `json:"company_name" validate:"required,min=2,max=255"`
	Email       string `json:"email" validate:"required,email"`
	Slug        string `json:"slug,omitempty"` // Optional, will be generated from company_name if not provided

	// Initial admin user details
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
}

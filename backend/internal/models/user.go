package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// User represents a user account in the system
type User struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`

	// Authentication
	Email         string `json:"email" db:"email"`
	EmailVerified bool   `json:"email_verified" db:"email_verified"`
	PasswordHash  string `json:"-" db:"password_hash"` // Never expose in JSON

	// Profile
	FirstName string  `json:"first_name" db:"first_name"`
	LastName  string  `json:"last_name" db:"last_name"`
	Phone     *string `json:"phone,omitempty" db:"phone"`
	AvatarURL *string `json:"avatar_url,omitempty" db:"avatar_url"`

	// Organization
	DepartmentID *uuid.UUID `json:"department_id,omitempty" db:"department_id"`

	// Status: active | suspended | deactivated | pending
	Status string `json:"status" db:"status"`

	// Password reset
	ResetToken          *uuid.UUID `json:"-" db:"reset_token"`
	ResetTokenExpiresAt *time.Time `json:"-" db:"reset_token_expires_at"`

	// Two-Factor Authentication
	TwoFactorEnabled       bool           `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret        *string        `json:"-" db:"two_factor_secret"` // Encrypted
	TwoFactorBackupCodes   pq.StringArray `json:"-" db:"two_factor_backup_codes"` // Encrypted array
	TwoFactorEnabledAt     *time.Time     `json:"two_factor_enabled_at,omitempty" db:"two_factor_enabled_at"`
	TwoFactorRecoveryEmail *string        `json:"two_factor_recovery_email,omitempty" db:"two_factor_recovery_email"`

	// Activity tracking
	LastLoginAt   *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	LastLoginIP   *string    `json:"last_login_ip,omitempty" db:"last_login_ip"`
	LastActiveAt  *time.Time `json:"last_active_at,omitempty" db:"last_active_at"`

	// Preferences
	Timezone    string          `json:"timezone" db:"timezone"`
	Language    string          `json:"language" db:"language"`
	Preferences json.RawMessage `json:"preferences,omitempty" db:"preferences"`

	// Metadata
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy *uuid.UUID `json:"created_by,omitempty" db:"created_by"`

	// Computed fields (not in database)
	Roles       []Role       `json:"roles,omitempty" db:"-"`
	Permissions []Permission `json:"permissions,omitempty" db:"-"`
}

// UserStatus constants
const (
	UserStatusActive      = "active"
	UserStatusSuspended   = "suspended"
	UserStatusDeactivated = "deactivated"
	UserStatusPending     = "pending"
)

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsSuspended returns true if the user is suspended
func (u *User) IsSuspended() bool {
	return u.Status == UserStatusSuspended
}

// CanLogin returns true if the user can log in
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive && u.EmailVerified
}

// FullName returns the user's full name
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// UserCreateRequest represents a request to create a new user
type UserCreateRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,phone"`
	RoleIDs   []uuid.UUID `json:"role_ids,omitempty"` // Roles to assign
}

// UserUpdateRequest represents a request to update a user
type UserUpdateRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,phone"`
	AvatarURL *string `json:"avatar_url,omitempty"`
	Timezone  *string `json:"timezone,omitempty"`
	Language  *string `json:"language,omitempty"`
}

// UserLoginRequest represents a login request
type UserLoginRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required"`
	RememberMe bool   `json:"remember_me,omitempty"`
	TenantID   string `json:"tenant_id,omitempty"` // Optional: For multi-tenant selection
}

// UserLoginResponse represents a login response
type UserLoginResponse struct {
	User              *User      `json:"user,omitempty"`
	Tenant            *Tenant    `json:"tenant,omitempty"`
	Tenants           []*Tenant  `json:"tenants,omitempty"` // For multi-tenant selection
	AccessToken       string     `json:"access_token,omitempty"`
	RefreshToken      string     `json:"refresh_token,omitempty"`
	ExpiresIn         int64      `json:"expires_in,omitempty"` // Seconds until expiration
	RequiresTwoFactor bool       `json:"requires_two_factor,omitempty"`
	TwoFactorToken    string     `json:"two_factor_token,omitempty"` // Temporary token for 2FA verification
}

// UserPasswordResetRequest represents a password reset request
type UserPasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// UserPasswordResetConfirm represents a password reset confirmation
type UserPasswordResetConfirm struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserChangePasswordRequest represents a password change request
type UserChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

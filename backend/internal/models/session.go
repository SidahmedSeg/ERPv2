package models

import (
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with device information
type Session struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`

	// Token (hashed for security)
	TokenHash string `json:"-" db:"token_hash"`

	// Device information
	DeviceType string  `json:"device_type" db:"device_type"` // Desktop | Mobile | Tablet
	Browser    string  `json:"browser" db:"browser"`
	OS         string  `json:"os" db:"os"`
	IPAddress  *string `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent  string  `json:"user_agent,omitempty" db:"user_agent"`

	// Location (GeoIP)
	CountryCode *string `json:"country_code,omitempty" db:"country_code"`
	City        *string `json:"city,omitempty" db:"city"`

	// Activity
	LastActivityAt time.Time `json:"last_activity_at" db:"last_activity_at"`

	// Expiry
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`

	// Metadata
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// Computed fields
	IsActive  bool   `json:"is_active" db:"-"`
	IsCurrent bool   `json:"is_current" db:"-"` // True if this is the current session
}

// IsExpired returns true if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsInactive returns true if the session has been inactive for too long
func (s *Session) IsInactive(maxInactivity time.Duration) bool {
	return time.Since(s.LastActivityAt) > maxInactivity
}

// DeviceString returns a human-readable device description
func (s *Session) DeviceString() string {
	return s.Browser + " on " + s.OS + " (" + s.DeviceType + ")"
}

// SessionCreateRequest represents a request to create a new session
type SessionCreateRequest struct {
	UserID     uuid.UUID
	TenantID   uuid.UUID
	Token      string // Unhashed token
	DeviceType string
	Browser    string
	OS         string
	IPAddress  string
	UserAgent  string
	ExpiresAt  time.Time
}

// SessionListResponse represents a list of user sessions for display
type SessionListResponse struct {
	Sessions      []Session `json:"sessions"`
	CurrentCount  int       `json:"current_count"`
	TotalCount    int       `json:"total_count"`
}

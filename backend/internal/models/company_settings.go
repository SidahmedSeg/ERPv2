package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CompanySettings represents company-wide settings
type CompanySettings struct {
	ID       uuid.UUID `db:"id" json:"id"`
	TenantID uuid.UUID `db:"tenant_id" json:"tenant_id"`

	// Company Information
	CompanyName       string  `db:"company_name" json:"company_name"`
	LegalBusinessName *string `db:"legal_business_name" json:"legal_business_name,omitempty"`
	Industry          *string `db:"industry" json:"industry,omitempty"`
	Speciality        *string `db:"speciality" json:"speciality,omitempty"`
	CompanySize       *string `db:"company_size" json:"company_size,omitempty"`
	FoundedDate       *string `db:"founded_date" json:"founded_date,omitempty"`
	WebsiteURL        *string `db:"website_url" json:"website_url,omitempty"`
	LogoURL           *string `db:"logo_url" json:"logo_url,omitempty"`

	// Contact Details
	PrimaryEmail *string `db:"primary_email" json:"primary_email,omitempty"`
	SupportEmail *string `db:"support_email" json:"support_email,omitempty"`
	PhoneNumber  *string `db:"phone_number" json:"phone_number,omitempty"`
	Fax          *string `db:"fax" json:"fax,omitempty"`

	// Address
	StreetAddress *string `db:"street_address" json:"street_address,omitempty"`
	City          *string `db:"city" json:"city,omitempty"`
	State         *string `db:"state" json:"state,omitempty"`
	PostalCode    *string `db:"postal_code" json:"postal_code,omitempty"`
	Country       *string `db:"country" json:"country,omitempty"`

	// Business Hours
	Timezone          string          `db:"timezone" json:"timezone"`
	WorkingDays       json.RawMessage `db:"working_days" json:"working_days"`
	WorkingHoursStart string          `db:"working_hours_start" json:"working_hours_start"`
	WorkingHoursEnd   string          `db:"working_hours_end" json:"working_hours_end"`

	// Fiscal Settings
	FiscalYearStart *string  `db:"fiscal_year_start" json:"fiscal_year_start,omitempty"`
	DefaultCurrency string   `db:"default_currency" json:"default_currency"`
	DateFormat      string   `db:"date_format" json:"date_format"`
	NumberFormat    string   `db:"number_format" json:"number_format"`
	RCNumber        *string  `db:"rc_number" json:"rc_number,omitempty"`
	NIFNumber       *string  `db:"nif_number" json:"nif_number,omitempty"`
	NISNumber       *string  `db:"nis_number" json:"nis_number,omitempty"`
	AINumber        *string  `db:"ai_number" json:"ai_number,omitempty"`
	CapitalSocial   *float64 `db:"capital_social" json:"capital_social,omitempty"`

	// Metadata
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy *uuid.UUID `db:"created_by" json:"created_by,omitempty"`
	UpdatedBy *uuid.UUID `db:"updated_by" json:"updated_by,omitempty"`
}

// CompanySettingsUpdateRequest represents a request to update company settings
type CompanySettingsUpdateRequest struct {
	// All fields optional for partial updates
	CompanyName       *string          `json:"company_name,omitempty"`
	LegalBusinessName *string          `json:"legal_business_name,omitempty"`
	Industry          *string          `json:"industry,omitempty"`
	Speciality        *string          `json:"speciality,omitempty"`
	CompanySize       *string          `json:"company_size,omitempty"`
	FoundedDate       *string          `json:"founded_date,omitempty"`
	WebsiteURL        *string          `json:"website_url,omitempty"`
	LogoURL           *string          `json:"logo_url,omitempty"`
	PrimaryEmail      *string          `json:"primary_email,omitempty"`
	SupportEmail      *string          `json:"support_email,omitempty"`
	PhoneNumber       *string          `json:"phone_number,omitempty"`
	Fax               *string          `json:"fax,omitempty"`
	StreetAddress     *string          `json:"street_address,omitempty"`
	City              *string          `json:"city,omitempty"`
	State             *string          `json:"state,omitempty"`
	PostalCode        *string          `json:"postal_code,omitempty"`
	Country           *string          `json:"country,omitempty"`
	Timezone          *string          `json:"timezone,omitempty"`
	WorkingDays       *json.RawMessage `json:"working_days,omitempty"`
	WorkingHoursStart *string          `json:"working_hours_start,omitempty"`
	WorkingHoursEnd   *string          `json:"working_hours_end,omitempty"`
	FiscalYearStart   *string          `json:"fiscal_year_start,omitempty"`
	DefaultCurrency   *string          `json:"default_currency,omitempty"`
	DateFormat        *string          `json:"date_format,omitempty"`
	NumberFormat      *string          `json:"number_format,omitempty"`
	RCNumber          *string          `json:"rc_number,omitempty"`
	NIFNumber         *string          `json:"nif_number,omitempty"`
	NISNumber         *string          `json:"nis_number,omitempty"`
	AINumber          *string          `json:"ai_number,omitempty"`
	CapitalSocial     *float64         `json:"capital_social,omitempty"`
}

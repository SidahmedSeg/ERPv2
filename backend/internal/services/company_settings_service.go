package services

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"myerp-v2/internal/models"
	"myerp-v2/internal/repository"
)

// CompanySettingsService handles company settings business logic
type CompanySettingsService struct {
	repo         *repository.CompanySettingsRepository
	tenantRepo   *repository.TenantRepository
	auditService *AuditService
}

// NewCompanySettingsService creates a new company settings service
func NewCompanySettingsService(
	repo *repository.CompanySettingsRepository,
	tenantRepo *repository.TenantRepository,
	auditService *AuditService,
) *CompanySettingsService {
	return &CompanySettingsService{
		repo:         repo,
		tenantRepo:   tenantRepo,
		auditService: auditService,
	}
}

// GetSettings retrieves company settings for a tenant
func (s *CompanySettingsService) GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.CompanySettings, error) {
	settings, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Return nil if no settings exist (frontend will create defaults)
	return settings, nil
}

// UpdateSettings updates company settings (creates if doesn't exist)
func (s *CompanySettingsService) UpdateSettings(
	ctx context.Context,
	tenantID, userID uuid.UUID,
	req *models.CompanySettingsUpdateRequest,
) (*models.CompanySettings, error) {
	// Check if settings exist
	existing, err := s.repo.GetByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// If settings don't exist, create initial settings
	if existing == nil {
		tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
		if err != nil {
			return nil, err
		}

		// Default working days
		workingDays := map[string]bool{
			"monday":    true,
			"tuesday":   true,
			"wednesday": true,
			"thursday":  true,
			"friday":    true,
			"saturday":  false,
			"sunday":    false,
		}
		workingDaysJSON, _ := json.Marshal(workingDays)

		// Create initial settings with tenant name
		companyName := tenant.CompanyName
		if req.CompanyName != nil {
			companyName = *req.CompanyName
		}

		newSettings := &models.CompanySettings{
			TenantID:          tenantID,
			CompanyName:       companyName,
			Timezone:          "UTC",
			WorkingDays:       workingDaysJSON,
			WorkingHoursStart: "09:00",
			WorkingHoursEnd:   "17:00",
			DefaultCurrency:   "USD",
			DateFormat:        "DD/MM/YYYY",
			NumberFormat:      "1,000.00",
			CreatedBy:         &userID,
		}

		if err := s.repo.Create(ctx, tenantID, newSettings); err != nil {
			return nil, err
		}

		existing = newSettings
	}

	// Build update map from request
	updates := make(map[string]interface{})

	if req.CompanyName != nil {
		updates["company_name"] = *req.CompanyName
	}
	if req.LegalBusinessName != nil {
		updates["legal_business_name"] = *req.LegalBusinessName
	}
	if req.Industry != nil {
		updates["industry"] = *req.Industry
	}
	if req.Speciality != nil {
		updates["speciality"] = *req.Speciality
	}
	if req.CompanySize != nil {
		updates["company_size"] = *req.CompanySize
	}
	if req.FoundedDate != nil {
		updates["founded_date"] = *req.FoundedDate
	}
	if req.WebsiteURL != nil {
		updates["website_url"] = *req.WebsiteURL
	}
	if req.LogoURL != nil {
		updates["logo_url"] = *req.LogoURL
	}
	if req.PrimaryEmail != nil {
		updates["primary_email"] = *req.PrimaryEmail
	}
	if req.SupportEmail != nil {
		updates["support_email"] = *req.SupportEmail
	}
	if req.PhoneNumber != nil {
		updates["phone_number"] = *req.PhoneNumber
	}
	if req.Fax != nil {
		updates["fax"] = *req.Fax
	}
	if req.StreetAddress != nil {
		updates["street_address"] = *req.StreetAddress
	}
	if req.City != nil {
		updates["city"] = *req.City
	}
	if req.State != nil {
		updates["state"] = *req.State
	}
	if req.PostalCode != nil {
		updates["postal_code"] = *req.PostalCode
	}
	if req.Country != nil {
		updates["country"] = *req.Country
	}
	if req.Timezone != nil {
		updates["timezone"] = *req.Timezone
	}
	if req.WorkingDays != nil {
		updates["working_days"] = *req.WorkingDays
	}
	if req.WorkingHoursStart != nil {
		updates["working_hours_start"] = *req.WorkingHoursStart
	}
	if req.WorkingHoursEnd != nil {
		updates["working_hours_end"] = *req.WorkingHoursEnd
	}
	if req.FiscalYearStart != nil {
		updates["fiscal_year_start"] = *req.FiscalYearStart
	}
	if req.DefaultCurrency != nil {
		updates["default_currency"] = *req.DefaultCurrency
	}
	if req.DateFormat != nil {
		updates["date_format"] = *req.DateFormat
	}
	if req.NumberFormat != nil {
		updates["number_format"] = *req.NumberFormat
	}
	if req.RCNumber != nil {
		updates["rc_number"] = *req.RCNumber
	}
	if req.NIFNumber != nil {
		updates["nif_number"] = *req.NIFNumber
	}
	if req.NISNumber != nil {
		updates["nis_number"] = *req.NISNumber
	}
	if req.AINumber != nil {
		updates["ai_number"] = *req.AINumber
	}
	if req.CapitalSocial != nil {
		updates["capital_social"] = *req.CapitalSocial
	}

	// If no updates, return existing settings
	if len(updates) == 0 {
		return existing, nil
	}

	// Update settings
	updated, err := s.repo.Update(ctx, tenantID, updates, userID)
	if err != nil {
		return nil, err
	}

	// Audit log
	s.auditService.LogEvent(ctx, tenantID, userID, "company_settings.updated", "company_settings", updated.ID, "success", "", "", updates)

	return updated, nil
}

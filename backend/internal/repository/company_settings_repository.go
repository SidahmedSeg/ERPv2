package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"myerp-v2/internal/database"
	"myerp-v2/internal/models"
)

// CompanySettingsRepository handles company settings data access
type CompanySettingsRepository struct {
	db *sqlx.DB
}

// NewCompanySettingsRepository creates a new company settings repository
func NewCompanySettingsRepository(db *sqlx.DB) *CompanySettingsRepository {
	return &CompanySettingsRepository{db: db}
}

// GetByTenantID retrieves company settings for a tenant
func (r *CompanySettingsRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*models.CompanySettings, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var settings models.CompanySettings
	query := `SELECT * FROM company_settings LIMIT 1`

	err = tx.GetContext(ctx, &settings, query)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No settings exist yet
		}
		return nil, err
	}

	return &settings, nil
}

// Create creates new company settings
func (r *CompanySettingsRepository) Create(ctx context.Context, tenantID uuid.UUID, settings *models.CompanySettings) error {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO company_settings (
			tenant_id, company_name, timezone, working_days,
			working_hours_start, working_hours_end, default_currency,
			date_format, number_format, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	err = tx.QueryRowContext(ctx, query,
		tenantID, settings.CompanyName, settings.Timezone, settings.WorkingDays,
		settings.WorkingHoursStart, settings.WorkingHoursEnd, settings.DefaultCurrency,
		settings.DateFormat, settings.NumberFormat, settings.CreatedBy,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		return err
	}

	return tx.Commit()
}

// Update updates company settings with dynamic fields
func (r *CompanySettingsRepository) Update(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}, updatedBy uuid.UUID) (*models.CompanySettings, error) {
	tx, err := database.WithTenantContext(ctx, r.db, tenantID)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if len(updates) == 0 {
		// No updates provided, just return current settings
		var settings models.CompanySettings
		err := tx.GetContext(ctx, &settings, "SELECT * FROM company_settings LIMIT 1")
		if err != nil {
			return nil, err
		}
		return &settings, nil
	}

	// Whitelist of allowed column names to prevent SQL injection
	allowedColumns := map[string]bool{
		"company_name": true, "legal_name": true, "industry": true,
		"employee_count": true, "founded_date": true, "website": true,
		"description": true, "email": true, "phone": true, "fax": true,
		"street_address": true, "city": true, "state": true,
		"postal_code": true, "country": true, "fiscal_year_start": true,
		"default_currency": true, "date_format": true, "time_zone": true,
		"language": true, "tax_id": true, "registration_number": true,
		"vat_number": true, "nif_number": true, "ai_number": true,
		"logo_url": true, "preferences": true,
	}

	// Build dynamic UPDATE query
	query := "UPDATE company_settings SET updated_at = NOW(), updated_by = $1"
	args := []interface{}{updatedBy}
	argIndex := 2

	for field, value := range updates {
		// Security: Only allow whitelisted column names
		if !allowedColumns[field] {
			continue
		}
		query += fmt.Sprintf(", %s = $%d", field, argIndex)
		args = append(args, value)
		argIndex++
	}

	query += " RETURNING *"

	var settings models.CompanySettings
	err = tx.GetContext(ctx, &settings, query, args...)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &settings, nil
}

# Company Settings Implementation Plan

## Overview
Implement comprehensive company settings page matching the original MyERP with Google Places API integration for address autocomplete.

---

## 📋 Features Summary

### 1. Company Information Card
- Company logo upload (via ParaDrive)
- Company name (editable)
- Legal business name
- **Industry (cascading comboboxes):**
  - **Primary Industry** (15 core industries)
  - **Speciality** (5-10 options per industry, auto-populated based on selected industry)
- Company size (1-10, 11-50, 51-200, 201-500, 500+)
- Founded date (date picker)
- Website URL

### 2. Contact Details Card
- Primary email
- Support email
- Phone number (with country code selector: +213 Algeria default)
- Fax (with country code selector)

### 3. Address Card
- **Google Places Autocomplete** for address search
- Street address (textarea)
- City
- State/Province
- Postal code
- Country

### 4. Business Hours Card
- Timezone selector (UTC, ET, CT, MT, PT, GMT, CET, Algiers)
- Working days checkboxes (Mon-Sun)
- Working hours start time
- Working hours end time

### 5. Fiscal Settings Card
- Fiscal year start date
- Default currency (USD, EUR, GBP, DZD, CAD)
- Date format (DD/MM/YYYY, MM/DD/YYYY, YYYY-MM-DD)
- Number format (1,000.00, 1.000,00, 1 000.00)
- RC Number (alphanumeric - Commerce Registry)
- NIF Number (numeric - Tax ID)
- NIS Number (numeric - Social Security ID)
- AI Number (numeric - Statistical ID)
- Capital Social (decimal)

---

## 🔑 Google Places API Configuration

### API Key
```
AIzaSyAxOFMLNk2NuAf0fojr6oRnM-MD6oM8zpA
```

### Integration Points
1. **Script Loading** (lines 102-126 in old version):
   - Load Google Maps JavaScript API with Places library
   - Script: `https://maps.googleapis.com/maps/api/js?key=${apiKey}&libraries=places`
   - Async + defer loading
   - Check if already loaded to avoid duplicates

2. **Autocomplete Initialization** (lines 128-275):
   - Initialize when Address dialog opens and Google is loaded
   - Options: `{ types: ["address"], fields: ["address_components", "formatted_address"] }`
   - Parse address components: street_number, route, locality, administrative_area_level_1, postal_code, country
   - Handle pac-container dropdown clicks (prevent blur)
   - Clear search input after selection

3. **Security**:
   - Environment variable: `NEXT_PUBLIC_GOOGLE_PLACES_API_KEY`
   - Client-side only (NEXT_PUBLIC_ prefix)
   - Should be restricted to specific domains in Google Cloud Console

### API Restrictions (Recommended)
1. Go to Google Cloud Console → APIs & Services → Credentials
2. Edit API key → Application restrictions:
   - HTTP referrers
   - Add: `https://app.infold.app/*`, `https://localhost:*` (for dev)
3. API restrictions:
   - Restrict key to: Maps JavaScript API, Places API

---

## 🗄️ Backend Database Schema

### `company_settings` Table
```sql
CREATE TABLE company_settings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Company Information
    company_name VARCHAR(255) NOT NULL,
    legal_business_name VARCHAR(255),
    industry VARCHAR(100),
    speciality VARCHAR(100),
    company_size VARCHAR(50),
    founded_date DATE,
    website_url TEXT,
    logo_url TEXT,

    -- Contact Details
    primary_email VARCHAR(255),
    support_email VARCHAR(255),
    phone_number VARCHAR(50),
    fax VARCHAR(50),

    -- Address
    street_address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100),

    -- Business Hours
    timezone VARCHAR(100) NOT NULL DEFAULT 'UTC',
    working_days JSONB NOT NULL DEFAULT '{"monday": true, "tuesday": true, "wednesday": true, "thursday": true, "friday": true, "saturday": false, "sunday": false}',
    working_hours_start TIME NOT NULL DEFAULT '09:00',
    working_hours_end TIME NOT NULL DEFAULT '17:00',

    -- Fiscal Settings
    fiscal_year_start VARCHAR(10),
    default_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    date_format VARCHAR(20) NOT NULL DEFAULT 'DD/MM/YYYY',
    number_format VARCHAR(20) NOT NULL DEFAULT '1,000.00',

    -- Tax/Legal IDs (Algeria specific)
    rc_number VARCHAR(50),      -- Commerce Registry
    nif_number VARCHAR(50),     -- Tax Identification
    nis_number VARCHAR(50),     -- Social Security ID
    ai_number VARCHAR(50),      -- Statistical ID
    capital_social DECIMAL(15,2),

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,
    updated_by UUID,

    CONSTRAINT unique_settings_per_tenant UNIQUE(tenant_id)
);

CREATE INDEX idx_company_settings_tenant ON company_settings(tenant_id);

-- Enable RLS
ALTER TABLE company_settings ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON company_settings
    USING (tenant_id = current_setting('app.current_tenant_id')::uuid);

-- Trigger for updated_at
CREATE TRIGGER update_company_settings_updated_at
    BEFORE UPDATE ON company_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
```

### Migration File
**File:** `backend/migrations/015_create_company_settings.up.sql`
**Down:** `backend/migrations/015_create_company_settings.down.sql`

---

## 🔧 Backend Implementation

### 1. Create Model
**File:** `backend/internal/models/company_settings.go`

```go
package models

import (
    "encoding/json"
    "time"
    "github.com/google/uuid"
)

type CompanySettings struct {
    ID        uuid.UUID `db:"id" json:"id"`
    TenantID  uuid.UUID `db:"tenant_id" json:"tenant_id"`

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

type CompanySettingsUpdateRequest struct {
    // All fields optional for partial updates
    CompanyName       *string  `json:"company_name,omitempty"`
    LegalBusinessName *string  `json:"legal_business_name,omitempty"`
    Industry          *string  `json:"industry,omitempty"`
    Speciality        *string  `json:"speciality,omitempty"`
    CompanySize       *string  `json:"company_size,omitempty"`
    FoundedDate       *string  `json:"founded_date,omitempty"`
    WebsiteURL        *string  `json:"website_url,omitempty"`
    LogoURL           *string  `json:"logo_url,omitempty"`
    PrimaryEmail      *string  `json:"primary_email,omitempty"`
    SupportEmail      *string  `json:"support_email,omitempty"`
    PhoneNumber       *string  `json:"phone_number,omitempty"`
    Fax               *string  `json:"fax,omitempty"`
    StreetAddress     *string  `json:"street_address,omitempty"`
    City              *string  `json:"city,omitempty"`
    State             *string  `json:"state,omitempty"`
    PostalCode        *string  `json:"postal_code,omitempty"`
    Country           *string  `json:"country,omitempty"`
    Timezone          *string  `json:"timezone,omitempty"`
    WorkingDays       *json.RawMessage `json:"working_days,omitempty"`
    WorkingHoursStart *string  `json:"working_hours_start,omitempty"`
    WorkingHoursEnd   *string  `json:"working_hours_end,omitempty"`
    FiscalYearStart   *string  `json:"fiscal_year_start,omitempty"`
    DefaultCurrency   *string  `json:"default_currency,omitempty"`
    DateFormat        *string  `json:"date_format,omitempty"`
    NumberFormat      *string  `json:"number_format,omitempty"`
    RCNumber          *string  `json:"rc_number,omitempty"`
    NIFNumber         *string  `json:"nif_number,omitempty"`
    NISNumber         *string  `json:"nis_number,omitempty"`
    AINumber          *string  `json:"ai_number,omitempty"`
    CapitalSocial     *float64 `json:"capital_social,omitempty"`
}
```

### 2. Create Repository
**File:** `backend/internal/repository/company_settings_repository.go`

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "myerp-v2/internal/database"
    "myerp-v2/internal/models"
)

type CompanySettingsRepository struct {
    db *sqlx.DB
}

func NewCompanySettingsRepository(db *sqlx.DB) *CompanySettingsRepository {
    return &CompanySettingsRepository{db: db}
}

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
        return nil, err
    }

    return &settings, nil
}

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

func (r *CompanySettingsRepository) Update(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}, updatedBy uuid.UUID) (*models.CompanySettings, error) {
    tx, err := database.WithTenantContext(ctx, r.db, tenantID)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // Build dynamic UPDATE query based on provided fields
    query := "UPDATE company_settings SET updated_at = NOW(), updated_by = $1"
    args := []interface{}{updatedBy}
    argIndex := 2

    for field, value := range updates {
        query += fmt.Sprintf(", %s = $%d", field, argIndex)
        args = append(args, value)
        argIndex++
    }

    query += " WHERE tenant_id = $" + fmt.Sprintf("%d", argIndex) + " RETURNING *"
    args = append(args, tenantID)

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
```

### 3. Create Service
**File:** `backend/internal/services/company_settings_service.go`

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/google/uuid"
    "myerp-v2/internal/models"
    "myerp-v2/internal/repository"
)

type CompanySettingsService struct {
    repo         *repository.CompanySettingsRepository
    tenantRepo   *repository.TenantRepository
    auditService *AuditService
}

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

func (s *CompanySettingsService) GetSettings(ctx context.Context, tenantID uuid.UUID) (*models.CompanySettings, error) {
    settings, err := s.repo.GetByTenantID(ctx, tenantID)
    if err != nil {
        // If no settings exist, return nil (frontend will create defaults)
        return nil, nil
    }
    return settings, nil
}

func (s *CompanySettingsService) UpdateSettings(
    ctx context.Context,
    tenantID, userID uuid.UUID,
    req *models.CompanySettingsUpdateRequest,
) (*models.CompanySettings, error) {
    // Check if settings exist
    existing, _ := s.repo.GetByTenantID(ctx, tenantID)

    if existing == nil {
        // Create initial settings
        tenant, err := s.tenantRepo.FindByID(ctx, tenantID)
        if err != nil {
            return nil, err
        }

        workingDays := map[string]bool{
            "monday": true, "tuesday": true, "wednesday": true,
            "thursday": true, "friday": true, "saturday": false, "sunday": false,
        }
        workingDaysJSON, _ := json.Marshal(workingDays)

        newSettings := &models.CompanySettings{
            TenantID:          tenantID,
            CompanyName:       tenant.CompanyName,
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

    if req.CompanyName != nil { updates["company_name"] = *req.CompanyName }
    if req.LegalBusinessName != nil { updates["legal_business_name"] = *req.LegalBusinessName }
    if req.Industry != nil { updates["industry"] = *req.Industry }
    if req.Speciality != nil { updates["speciality"] = *req.Speciality }
    if req.CompanySize != nil { updates["company_size"] = *req.CompanySize }
    if req.FoundedDate != nil { updates["founded_date"] = *req.FoundedDate }
    if req.WebsiteURL != nil { updates["website_url"] = *req.WebsiteURL }
    if req.LogoURL != nil { updates["logo_url"] = *req.LogoURL }
    if req.PrimaryEmail != nil { updates["primary_email"] = *req.PrimaryEmail }
    if req.SupportEmail != nil { updates["support_email"] = *req.SupportEmail }
    if req.PhoneNumber != nil { updates["phone_number"] = *req.PhoneNumber }
    if req.Fax != nil { updates["fax"] = *req.Fax }
    if req.StreetAddress != nil { updates["street_address"] = *req.StreetAddress }
    if req.City != nil { updates["city"] = *req.City }
    if req.State != nil { updates["state"] = *req.State }
    if req.PostalCode != nil { updates["postal_code"] = *req.PostalCode }
    if req.Country != nil { updates["country"] = *req.Country }
    if req.Timezone != nil { updates["timezone"] = *req.Timezone }
    if req.WorkingDays != nil { updates["working_days"] = *req.WorkingDays }
    if req.WorkingHoursStart != nil { updates["working_hours_start"] = *req.WorkingHoursStart }
    if req.WorkingHoursEnd != nil { updates["working_hours_end"] = *req.WorkingHoursEnd }
    if req.FiscalYearStart != nil { updates["fiscal_year_start"] = *req.FiscalYearStart }
    if req.DefaultCurrency != nil { updates["default_currency"] = *req.DefaultCurrency }
    if req.DateFormat != nil { updates["date_format"] = *req.DateFormat }
    if req.NumberFormat != nil { updates["number_format"] = *req.NumberFormat }
    if req.RCNumber != nil { updates["rc_number"] = *req.RCNumber }
    if req.NIFNumber != nil { updates["nif_number"] = *req.NIFNumber }
    if req.NISNumber != nil { updates["nis_number"] = *req.NISNumber }
    if req.AINumber != nil { updates["ai_number"] = *req.AINumber }
    if req.CapitalSocial != nil { updates["capital_social"] = *req.CapitalSocial }

    if len(updates) == 0 {
        return existing, nil
    }

    // Update settings
    updated, err := s.repo.Update(ctx, tenantID, updates, userID)
    if err != nil {
        return nil, err
    }

    // Audit log
    s.auditService.Log(ctx, tenantID, userID, "company_settings.updated", "company_settings", updated.ID, "success", updates)

    return updated, nil
}
```

### 4. Create Handler
**File:** `backend/internal/handlers/company_settings_handler.go`

```go
package handlers

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "myerp-v2/internal/middleware"
    "myerp-v2/internal/models"
    "myerp-v2/internal/services"
    "myerp-v2/internal/utils"
)

type CompanySettingsHandler struct {
    service *services.CompanySettingsService
}

func NewCompanySettingsHandler(service *services.CompanySettingsService) *CompanySettingsHandler {
    return &CompanySettingsHandler{service: service}
}

// GetSettings returns company settings
// GET /api/settings/company
func (h *CompanySettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
    tenantID, _ := middleware.GetTenantIDFromContext(r.Context())

    settings, err := h.service.GetSettings(r.Context(), tenantID)
    if err != nil {
        utils.InternalServerError(w, "Failed to fetch settings")
        return
    }

    utils.Success(w, settings)
}

// UpdateSettings updates company settings (partial update)
// PUT /api/settings/company
func (h *CompanySettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
    tenantID, _ := middleware.GetTenantIDFromContext(r.Context())
    userID, _ := middleware.GetUserIDFromContext(r.Context())

    var req models.CompanySettingsUpdateRequest
    if err := utils.ParseJSONBody(r, &req); err != nil {
        utils.BadRequest(w, "Invalid request body")
        return
    }

    settings, err := h.service.UpdateSettings(r.Context(), tenantID, userID, &req)
    if err != nil {
        utils.InternalServerError(w, "Failed to update settings")
        return
    }

    utils.Success(w, settings)
}

// RegisterRoutes registers company settings routes
func (h *CompanySettingsHandler) RegisterRoutes(r chi.Router, authMiddleware *middleware.AuthMiddleware, permMiddleware *middleware.PermissionMiddleware) {
    r.Route("/settings/company", func(r chi.Router) {
        r.Use(authMiddleware.Authenticate)
        r.Get("/", h.GetSettings)
        r.Put("/", h.UpdateSettings) // Could add permission check: permMiddleware.RequirePermission("settings", "edit")
    })
}
```

### 5. Register in Router
**File:** `backend/internal/server/router.go`

```go
// Add to initialization
companySettingsRepo := repository.NewCompanySettingsRepository(db)
companySettingsService := services.NewCompanySettingsService(companySettingsRepo, tenantRepo, auditService)
companySettingsHandler := handlers.NewCompanySettingsHandler(companySettingsService)

// Register routes
companySettingsHandler.RegisterRoutes(r, authMiddleware, permMiddleware)
```

---

## 🎨 Frontend Implementation

### Step 1: Add Google Places API Key to Environment
**File:** `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/.env.local`

```bash
NEXT_PUBLIC_API_URL=https://app.infold.app/api
NEXT_PUBLIC_APP_URL=https://app.infold.app
NEXT_PUBLIC_BASE_DOMAIN=infold.app
NEXT_PUBLIC_GOOGLE_PLACES_API_KEY=AIzaSyAxOFMLNk2NuAf0fojr6oRnM-MD6oM8zpA
```

### Step 2: Create Lib Files

**File 1:** `frontend/src/lib/industries-specialities.ts` ✅ ALREADY CREATED
```typescript
// This file contains:
// - INDUSTRIES_SPECIALITIES: Array of 15 industries with their specialities
// - INDUSTRIES: Extracted list of just industry names
// - getSpecialities(industry): Helper function to get specialities for selected industry
//
// Structure:
// [
//   { industry: "Technology & Software", specialities: ["Software Development", "IT Consulting", ...] },
//   { industry: "Manufacturing & Industrial", specialities: [...] },
//   ...
// ]
```

Location: `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/src/lib/industries-specialities.ts`

**File 2:** `frontend/src/lib/country-codes.ts`
```typescript
export const COUNTRY_CODES = [
  { code: "+213", country: "Algeria", flag: "🇩🇿" },
  { code: "+1", country: "United States", flag: "🇺🇸" },
  { code: "+44", country: "United Kingdom", flag: "🇬🇧" },
  { code: "+33", country: "France", flag: "🇫🇷" },
  // ... (copy all 54 entries from old version)
] as const;

export type CountryCode = typeof COUNTRY_CODES[number];
```

### Step 3: Implement Cascading Industry-Speciality Comboboxes

**Update CompanySettings Interface:**
```typescript
interface CompanySettings {
  // ... other fields
  industry?: string;        // Primary industry (15 options)
  speciality?: string;      // Speciality (5-10 options based on industry)
  // ... other fields
}
```

**Update Company Info Dialog (in general-settings.tsx):**

```typescript
// Add state for selected industry
const [selectedIndustry, setSelectedIndustry] = useState<string>("");
const [availableSpecialities, setAvailableSpecialities] = useState<string[]>([]);

// Import at top
import { INDUSTRIES, getSpecialities } from "@/lib/industries-specialities";

// In openCompanyInfoDialog function:
const openCompanyInfoDialog = () => {
  const industry = settings?.industry || "";
  setSelectedIndustry(industry);
  setAvailableSpecialities(industry ? getSpecialities(industry) : []);

  setCompanyInfoForm({
    company_name: settings?.company_name,
    legal_business_name: settings?.legal_business_name,
    industry: industry,
    speciality: settings?.speciality,
    company_size: settings?.company_size,
    founded_date: settings?.founded_date,
    website_url: settings?.website_url,
  });
  setIsCompanyInfoDialogOpen(true);
};

// Handle industry change
const handleIndustryChange = (industry: string) => {
  setSelectedIndustry(industry);
  setAvailableSpecialities(getSpecialities(industry));
  setCompanyInfoForm({
    ...companyInfoForm,
    industry,
    speciality: "" // Reset speciality when industry changes
  });
};

// In the dialog JSX:
<div className="grid grid-cols-2 gap-4">
  <div className="space-y-2">
    <Label htmlFor="industry">Industry *</Label>
    <Combobox
      value={companyInfoForm.industry || ""}
      onChange={handleIndustryChange}
      options={INDUSTRIES}
      placeholder="Select industry..."
      className="w-full bg-white"
    />
  </div>
  <div className="space-y-2">
    <Label htmlFor="speciality">Speciality</Label>
    <Combobox
      value={companyInfoForm.speciality || ""}
      onChange={(value) => setCompanyInfoForm({ ...companyInfoForm, speciality: value })}
      options={availableSpecialities}
      placeholder={selectedIndustry ? "Select speciality..." : "Select industry first"}
      className="w-full bg-white"
      disabled={!selectedIndustry}
    />
  </div>
</div>
```

**Display in Card (Read-Only):**
```typescript
<div>
  <p className="text-gray-500">Industry</p>
  <p className="font-medium">{settings.industry || "-"}</p>
</div>
<div>
  <p className="text-gray-500">Speciality</p>
  <p className="font-medium">{settings.speciality || "-"}</p>
</div>
```

### Step 4: Copy Exact Page from Old Version

**Files to Copy:**
1. `/Users/intelifoxdz/myerp-project/frontend/src/app/dashboard/settings/company/page.tsx`
   → `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/src/app/dashboard/settings/company/page.tsx`

2. `/Users/intelifoxdz/myerp-project/frontend/src/app/dashboard/settings/company/_components/general-settings.tsx`
   → `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/src/app/dashboard/settings/company/_components/general-settings.tsx`

### Step 4: Update API URLs in Frontend
In `general-settings.tsx`, change:
- `http://localhost:8080/api/settings/company` → `https://app.infold.app/api/settings/company`
- Or use environment variable: `process.env.NEXT_PUBLIC_API_URL + '/settings/company'`

### Step 5: Remove ParaDrive Dependency (Temporary)
Since ParaDrive is not implemented yet:
1. Remove logo upload functionality temporarily
2. Or add simple file upload input instead
3. Remove `FileSelectorDialog` import and usage

---

## ✅ Validation Plan

### Backend Validation

#### 1. Database Migration
```bash
# SSH into VPS
ssh root@167.86.117.179

# Run migration
cd /opt/myerp-v2/backend
docker exec myerp_postgres psql -U myerp -d myerp_v2 -f /root/migrations/015_create_company_settings.up.sql

# Verify table created
docker exec myerp_postgres psql -U myerp -d myerp_v2 -c "\d company_settings"
```

Expected output: Table structure with all columns

#### 2. API Testing

**Test 1: GET Settings (Empty)**
```bash
curl -X GET https://app.infold.app/api/settings/company \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```
Expected: `{"success": true, "data": null}` (no settings yet)

**Test 2: PUT Settings (Create)**
```bash
curl -X PUT https://app.infold.app/api/settings/company \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "MOZG",
    "industry": "Information Technology",
    "company_size": "1-10",
    "primary_email": "info@mozg.dz",
    "phone_number": "+213 555 123 456",
    "timezone": "Africa/Algiers",
    "default_currency": "DZD"
  }'
```
Expected: `{"success": true, "data": {...}}` with created settings

**Test 3: PUT Settings (Update)**
```bash
curl -X PUT https://app.infold.app/api/settings/company \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "street_address": "123 Rue Test",
    "city": "Algiers",
    "postal_code": "16000",
    "country": "Algeria"
  }'
```
Expected: Previous settings + new address fields

**Test 4: GET Settings (Populated)**
```bash
curl -X GET https://app.infold.app/api/settings/company \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```
Expected: All settings returned

#### 3. RLS Verification
```sql
-- As different tenant, try to access MOZG settings
-- Should return 0 rows (RLS blocking)
SET LOCAL app.current_tenant_id = 'different-tenant-id';
SELECT * FROM company_settings WHERE company_name = 'MOZG';
```

### Frontend Validation

#### 1. Google Places API Loading
1. Open DevTools → Network tab
2. Navigate to https://app.infold.app/dashboard/settings/company
3. Look for request to `maps.googleapis.com/maps/api/js`
4. Status should be 200
5. Check console for "Google Maps API loaded successfully"

#### 2. Google Places Autocomplete
1. Click Edit button on Address card
2. Type in "Search Address" field
3. Google dropdown should appear with suggestions
4. Select an address
5. Verify all fields auto-populate (street, city, state, postal, country)
6. Clear search input should empty
7. Save should update backend

#### 3. Form Validation
Test each dialog:

**Company Information:**
- Upload logo (if ParaDrive available)
- **Test cascading industry-speciality:**
  - Select industry from first combobox (15 options)
  - Verify speciality combobox becomes enabled
  - Verify speciality options match selected industry (5-10 options)
  - Change industry → verify speciality resets and new options appear
  - Save and verify both fields persist
- Select company size dropdown
- Pick founded date
- Enter website URL (validate format)

**Contact Details:**
- Enter emails (validate format)
- Select country code +213
- Enter phone number
- Select fax country code
- Enter fax number

**Address:**
- Search with Google Places
- Manually edit fields
- Save and verify

**Business Hours:**
- Change timezone to "Africa/Algiers"
- Toggle working days
- Change start/end times

**Fiscal Settings:**
- Pick fiscal year start date
- Change currency to DZD
- Change date format
- Change number format
- Enter RC, NIF, NIS, AI numbers
- Enter capital social

#### 4. API Integration
1. Make changes in any dialog
2. Open DevTools → Network
3. Click Update
4. Verify PUT request to `/api/settings/company`
5. Check response: `{"success": true, "data": {...}}`
6. Verify toast notification "Settings updated successfully"
7. Refresh page
8. Verify data persists

#### 5. Error Handling
1. Stop backend container
2. Try to save settings
3. Verify error toast appears
4. Start backend
5. Retry save
6. Verify success

---

## 📝 Implementation Checklist

### Backend Tasks
- [ ] Create migration file `015_create_company_settings.up/down.sql`
- [ ] Run migration on VPS
- [ ] Create `models/company_settings.go`
- [ ] Create `repository/company_settings_repository.go`
- [ ] Create `services/company_settings_service.go`
- [ ] Create `handlers/company_settings_handler.go`
- [ ] Register routes in `server/router.go`
- [ ] Test API endpoints with curl
- [ ] Verify RLS enforcement

### Frontend Tasks
- [ ] Add Google Places API key to `.env.local`
- [x] Create `lib/industries-specialities.ts` (15 industries + specialities) ✅
- [ ] Create `lib/country-codes.ts`
- [ ] Copy `settings/company/page.tsx` from old version
- [ ] Copy `settings/company/_components/general-settings.tsx`
- [ ] Implement cascading industry-speciality comboboxes:
  - [ ] Update CompanySettings interface (add speciality field)
  - [ ] Add selectedIndustry and availableSpecialities state
  - [ ] Import INDUSTRIES and getSpecialities helper
  - [ ] Update openCompanyInfoDialog to initialize specialities
  - [ ] Add handleIndustryChange function
  - [ ] Update dialog JSX with two comboboxes (industry + speciality)
  - [ ] Update read-only display card to show both fields
- [ ] Update API URLs to production
- [ ] Handle ParaDrive dependency (remove or stub)
- [ ] Test Google Places loading in browser
- [ ] Test autocomplete functionality
- [ ] Test all 5 dialogs (save/cancel)
- [ ] Test form validation
- [ ] Test API integration
- [ ] Test error handling
- [ ] Verify data persistence after refresh

### Deployment Tasks
- [ ] Build backend with new code
- [ ] Deploy backend to VPS
- [ ] Build frontend with new code
- [ ] Deploy frontend to VPS
- [ ] Verify Google Places API key restrictions in Google Cloud Console
- [ ] Test end-to-end on production

---

## 🔒 Google Cloud Console Setup

1. Go to: https://console.cloud.google.com/apis/credentials
2. Select project (or create new one)
3. Click on API key `AIzaSyAxOFMLNk2NuAf0fojr6oRnM-MD6oM8zpA`
4. Application restrictions:
   - Type: HTTP referrers
   - Add: `https://app.infold.app/*`
   - Add: `https://localhost:*` (for development)
5. API restrictions:
   - Restrict key
   - Select: Maps JavaScript API, Places API
6. Save

---

## 🚀 Next Steps

1. **Review this plan** with user
2. **Implement backend** (2-3 hours)
3. **Implement frontend** (1-2 hours)
4. **Test thoroughly** (1 hour)
5. **Deploy to production** (30 minutes)
6. **Validate with real data** (30 minutes)

**Total estimated time:** 5-7 hours

---

**Created:** 2026-01-19
**Status:** Ready for implementation
**Dependencies:** None (ParaDrive is optional)

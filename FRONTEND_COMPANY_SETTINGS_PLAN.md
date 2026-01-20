# Frontend Company Settings Implementation Plan

## 📊 Current Status Analysis

### ✅ What's Already Implemented

**Frontend (`general-settings.tsx`):**
- Complete UI with 5 card sections (Company Info, Contact, Address, Business Hours, Fiscal)
- Read-only display cards with edit buttons
- 5 modal dialogs for editing each section
- Google Places API integration for address autocomplete
- ParaDrive integration for logo upload
- Phone number country code selector
- All shadcn/ui components (Card, Dialog, Input, Select, etc.)
- API calls to `/api/settings/company` (GET and PUT)

**Backend (Just Created):**
- ✅ Database migration `015_create_company_settings.up.sql`
- ✅ Model `internal/models/company_settings.go`
- ✅ Repository `internal/repository/company_settings_repository.go`
- ✅ Service `internal/services/company_settings_service.go`
- ✅ Handler `internal/handlers/company_settings_handler.go`
- ✅ Routes registered in `internal/server/router.go`
- ✅ Endpoints: GET/PUT `/api/settings/company`

### ⚠️ Issues & Mismatches to Fix

| Issue | Frontend Field | Backend Field | Status |
|-------|---------------|---------------|--------|
| Working days JSON | `working_days: Record<string, boolean>` | `working_days: json.RawMessage` | ✅ Compatible |
| Company size | `company_size` | `company_size` | ✅ Matches |
| Founded date | `founded_date` | `founded_date` | ✅ Matches |
| Website | `website_url` | `website_url` | ✅ Matches |
| Primary email | `primary_email` | `primary_email` | ✅ Matches |
| Support email | `support_email` | `support_email` | ✅ Matches |

**Good news:** The frontend and backend schemas are already aligned! No field mismatches.

---

## 🎯 Implementation Plan

### Phase 1: Backend Setup & Testing (30 minutes)

**Step 1.1: Run Database Migration**
```bash
cd backend
go run cmd/migrate/main.go up
```
Expected: Creates `company_settings` table with RLS policies.

**Step 1.2: Compile and Test Backend**
```bash
cd backend
go build -o bin/server cmd/server/main.go
go run cmd/server/main.go
```
Expected: No compilation errors, server starts on port 8080.

**Step 1.3: Test API Endpoints**
```bash
# Get settings (should return empty or default settings)
curl -X GET http://localhost:8080/api/settings/company \
  -H "Authorization: Bearer <TOKEN>"

# Update settings
curl -X PUT http://localhost:8080/api/settings/company \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Test Company",
    "industry": "Technology",
    "timezone": "UTC"
  }'
```

---

### Phase 2: Frontend Type Definitions (15 minutes)

**Step 2.1: Update TypeScript Types**

File: `frontend/src/types/index.ts`

Add company settings types (already drafted, need to apply):
```typescript
export interface CompanySettings {
  id: string;
  tenant_id: string;
  company_name: string;
  legal_business_name?: string;
  industry?: string;
  speciality?: string;
  company_size?: string;
  founded_date?: string;
  website_url?: string;
  logo_url?: string;
  primary_email?: string;
  support_email?: string;
  phone_number?: string;
  fax?: string;
  street_address?: string;
  city?: string;
  state?: string;
  postal_code?: string;
  country?: string;
  timezone: string;
  working_days: Record<string, boolean>;
  working_hours_start: string;
  working_hours_end: string;
  fiscal_year_start?: string;
  default_currency: string;
  date_format: string;
  number_format: string;
  rc_number?: string;
  nif_number?: string;
  nis_number?: string;
  ai_number?: string;
  capital_social?: number;
  created_at: string;
  updated_at: string;
}
```

**Step 2.2: Create API Client Functions**

File: `frontend/src/lib/api/company-settings.ts` (NEW FILE)

```typescript
import { api } from '@/lib/api';
import type { ApiResponse, CompanySettings } from '@/types';

export const companySettingsApi = {
  // Get company settings
  getSettings: () =>
    api.get<ApiResponse<CompanySettings>>('/settings/company'),

  // Update company settings (partial update)
  updateSettings: (data: Partial<CompanySettings>) =>
    api.put<ApiResponse<CompanySettings>>('/settings/company', data),
};
```

---

### Phase 3: Google Places API Setup (20 minutes)

**Step 3.1: Get Google Places API Key**
1. Go to: https://console.cloud.google.com/apis/credentials
2. Create or select a project
3. Enable "Places API" and "Maps JavaScript API"
4. Create API key with restrictions:
   - Application restrictions: HTTP referrers
   - Website restrictions:
     - `http://localhost:13000/*`
     - `https://app.infold.app/*`
   - API restrictions:
     - ✓ Places API
     - ✓ Maps JavaScript API

**Step 3.2: Add API Key to Environment**

File: `frontend/.env.local`
```bash
NEXT_PUBLIC_GOOGLE_PLACES_API_KEY=AIza...your-key-here
```

File: `frontend/.env.production` (for VPS deployment)
```bash
NEXT_PUBLIC_GOOGLE_PLACES_API_KEY=AIza...your-key-here
```

**Step 3.3: Verify Google Places Integration**

The frontend already has the integration code in `general-settings.tsx` (lines 102-275). Just verify:
- ✅ Script loads on dialog open
- ✅ Autocomplete initializes
- ✅ Address components parse correctly
- ✅ pac-container doesn't cause dialog blur issues

---

### Phase 4: Frontend Refactoring (Optional - 1 hour)

**Current Status:** The frontend component works but could be cleaner.

**Recommended Refactoring (Optional):**

1. **Extract Dialog Components**
   - Create `_components/company-info-dialog.tsx`
   - Create `_components/contact-dialog.tsx`
   - Create `_components/address-dialog.tsx`
   - Create `_components/business-hours-dialog.tsx`
   - Create `_components/fiscal-dialog.tsx`

2. **Use New API Client**
   Replace direct fetch calls with:
   ```typescript
   import { companySettingsApi } from '@/lib/api/company-settings';

   const fetchSettings = async () => {
     const { data } = await companySettingsApi.getSettings();
     setSettings(data.data);
   };
   ```

3. **Add Zod Validation**
   Create `frontend/src/lib/validations/company-settings.ts`:
   ```typescript
   import { z } from 'zod';

   export const companyInfoSchema = z.object({
     company_name: z.string().min(1, 'Company name is required'),
     legal_business_name: z.string().optional(),
     industry: z.string().optional(),
     // ... etc
   });
   ```

**Decision:** Skip refactoring for now, focus on getting it working first.

---

### Phase 5: Integration Testing (30 minutes)

**Test Case 1: First-Time Setup**
1. Login as tenant owner
2. Navigate to Settings > Company
3. Verify default values appear
4. Edit each section (Company Info, Contact, Address, Business Hours, Fiscal)
5. Upload logo via ParaDrive
6. Verify all changes save successfully
7. Refresh page and verify data persists

**Test Case 2: Google Places Autocomplete**
1. Open Address dialog
2. Click on address search field
3. Type "1600 Amphitheatre Parkway"
4. Select suggestion
5. Verify fields auto-populate:
   - Street: "1600 Amphitheatre Parkway"
   - City: "Mountain View"
   - State: "California"
   - Postal Code: "94043"
   - Country: "United States"

**Test Case 3: Validation**
1. Try to clear company name (required field)
2. Verify validation error
3. Enter invalid email format
4. Verify validation error
5. Enter invalid phone number
6. Verify validation works

**Test Case 4: RLS Testing**
1. Create two tenants
2. Set company settings for Tenant A
3. Login as Tenant B
4. Verify Tenant B cannot see Tenant A's settings
5. Set different settings for Tenant B
6. Verify isolation

---

### Phase 6: VPS Deployment (45 minutes)

**Step 6.1: Prepare Backend**
```bash
cd backend
go build -o bin/server cmd/server/main.go
```

**Step 6.2: Build Frontend**
```bash
cd frontend
npm run build
```

**Step 6.3: Copy to VPS**
```bash
# Backend binary
scp backend/bin/server root@167.86.117.179:/opt/myerp-v2/backend/

# Frontend build
scp -r frontend/.next root@167.86.117.179:/opt/myerp-v2/frontend/
scp -r frontend/public root@167.86.117.179:/opt/myerp-v2/frontend/
scp frontend/.env.production root@167.86.117.179:/opt/myerp-v2/frontend/.env
```

**Step 6.4: Run Migration on VPS**
```bash
ssh root@167.86.117.179
cd /opt/myerp-v2/backend
./migrate up
```

**Step 6.5: Restart Services**
```bash
ssh root@167.86.117.179
cd /opt/myerp-v2
docker compose -f docker-compose.prod.yml restart backend frontend
```

**Step 6.6: Verify Deployment**
```bash
# Check logs
ssh root@167.86.117.179 "docker logs myerp-backend -n 50"
ssh root@167.86.117.179 "docker logs myerp-frontend -n 50"

# Test API
curl https://app.infold.app/api/health
```

---

## 📋 Complete Checklist

### Backend
- [x] Create database migration (`015_create_company_settings`)
- [x] Create model (`models/company_settings.go`)
- [x] Create repository with RLS (`repository/company_settings_repository.go`)
- [x] Create service (`services/company_settings_service.go`)
- [x] Create handler (`handlers/company_settings_handler.go`)
- [x] Register routes (`server/router.go`)
- [ ] Run migration locally (`go run cmd/migrate/main.go up`)
- [ ] Test backend compilation (`go build`)
- [ ] Test GET endpoint
- [ ] Test PUT endpoint

### Frontend
- [x] UI component exists (`general-settings.tsx`)
- [x] Google Places integration exists
- [x] ParaDrive integration exists
- [ ] Add TypeScript types to `types/index.ts`
- [ ] Create API client (`lib/api/company-settings.ts`)
- [ ] Update component to use API client (optional)
- [ ] Get Google Places API key
- [ ] Add API key to `.env.local`
- [ ] Add API key to `.env.production`
- [ ] Test Google Places autocomplete locally

### Testing
- [ ] Test first-time setup flow
- [ ] Test all 5 dialog sections
- [ ] Test logo upload
- [ ] Test Google Places autocomplete
- [ ] Test data persistence
- [ ] Test validation
- [ ] Test RLS isolation (multi-tenant)

### Deployment
- [ ] Build backend binary
- [ ] Build frontend
- [ ] Copy files to VPS
- [ ] Run migration on VPS
- [ ] Restart containers
- [ ] Test on production URL
- [ ] Verify logs

---

## 🚨 Critical Notes

1. **Google Places API Key:**
   - MUST restrict by HTTP referrer
   - MUST enable only Places API and Maps JavaScript API
   - DO NOT commit API key to Git

2. **RLS Context:**
   - Backend already uses `WithTenantContext()` correctly ✅
   - Each tenant's settings are isolated automatically

3. **Partial Updates:**
   - Backend supports partial updates (only send changed fields)
   - Frontend sends full section updates (safe, no issue)

4. **Default Values:**
   - Backend creates default settings on first access
   - Frontend handles `null` settings gracefully

5. **Working Days JSON:**
   - Frontend: `{ monday: true, tuesday: true, ... }`
   - Backend: Stored as JSONB, works seamlessly

---

## 🎯 Estimated Time

| Phase | Time | Status |
|-------|------|--------|
| Backend Setup & Testing | 30 min | ⏳ Next |
| Frontend Types & API | 15 min | ⏳ Pending |
| Google Places Setup | 20 min | ⏳ Pending |
| Integration Testing | 30 min | ⏳ Pending |
| VPS Deployment | 45 min | ⏳ Pending |
| **Total** | **~2.5 hours** | |

---

## 🔗 Reference Files

**Backend:**
- Migration: `backend/migrations/015_create_company_settings.up.sql`
- Model: `backend/internal/models/company_settings.go`
- Repository: `backend/internal/repository/company_settings_repository.go`
- Service: `backend/internal/services/company_settings_service.go`
- Handler: `backend/internal/handlers/company_settings_handler.go`

**Frontend:**
- Main Component: `frontend/src/app/dashboard/settings/company/_components/general-settings.tsx`
- Types: `frontend/src/types/index.ts`
- API Client: `frontend/src/lib/api.ts` (existing), `frontend/src/lib/api/company-settings.ts` (new)

**External:**
- Google Places API: https://console.cloud.google.com/apis/credentials
- Original Reference: `/Users/intelifoxdz/myerp-project/frontend/src/app/dashboard/settings/company/general-settings.tsx`

---

**Last Updated:** 2026-01-19
**Status:** Ready to implement - Backend complete, Frontend 90% complete
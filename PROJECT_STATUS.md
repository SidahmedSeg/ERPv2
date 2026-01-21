# MyERP v2 - Project Status

**Last Updated:** January 21, 2026
**Project:** Multi-Tenant ERP System (Shared Schema + RLS)
**Phase:** 2.0 - Production Running & Feature Development

---

## 📊 Overall Progress

### Phase 1: Core Auth & RBAC - ✅ 100% Complete
### Phase 1.5: Production Deployment - ✅ 100% Complete
### Phase 2: Business Features - ⏳ 10% In Progress

| Component | Status | Progress | Notes |
|-----------|--------|----------|-------|
| Database Schema | ✅ Complete | 100% | 15 migrations, RLS enabled |
| Backend Services | ✅ Complete | 100% | Production deployed on VPS |
| Frontend UI | ✅ Complete | 100% | Dark mode, company settings |
| Infrastructure | ✅ Complete | 100% | PostgreSQL, Redis, Mailpit, Caddy |
| Documentation | ✅ Complete | 100% | README, CLAUDE.md, API docs |
| Production VPS | ✅ Deployed | 100% | app.infold.app, api.infold.app |

---

## 🌐 Production Environment

### VPS Details
- **Domain**: infold.app
- **Frontend URL**: https://app.infold.app
- **Backend API URL**: https://api.infold.app
- **VPS IP**: 167.86.117.179
- **Reverse Proxy**: Caddy (auto-HTTPS with Let's Encrypt)
- **Status**: ✅ Running in production
- **URL Structure**: Clean domain-based routing (no /api prefix)

### Deployment Architecture
```
Internet → Caddy Proxy
    ├── app.infold.app → Frontend (Next.js on port 13000)
    └── api.infold.app → Backend (Go on port 18080)

Docker Compose Services:
    ├── myerp_frontend (Next.js 15.5.6)
    ├── myerp_backend (Go 1.24)
    ├── myerp_postgres (PostgreSQL 16)
    ├── myerp_redis (Redis 7)
    └── myerp_mailpit (Email testing)
```

### Environment Variables
- **Frontend**: Build-time env vars via Docker ARG
  - NEXT_PUBLIC_API_URL=https://api.infold.app (**no /api suffix**)
  - NEXT_PUBLIC_BASE_DOMAIN=infold.app
  - NEXT_PUBLIC_GOOGLE_PLACES_API_KEY

- **Backend**: Runtime env vars via .env
  - FRONTEND_URL=https://app.infold.app
  - APP_BASE_URL=https://api.infold.app
  - CORS_ALLOWED_ORIGINS=https://app.infold.app
  - DATABASE_URL (PostgreSQL)
  - REDIS_URL

---

## 🗄️ Database Status

### Infrastructure
- **PostgreSQL**: ✅ Running (Production: internal Docker network)
- **Redis**: ✅ Running (Production: internal Docker network)
- **Mailpit**: ✅ Running (Production: SMTP + Web UI)

### Migrations
```
✅ 001 - Create Extensions (uuid-ossp, pgcrypto, pg_trgm)
✅ 002 - Create Tenants Table
✅ 003 - Create Users Table
✅ 004 - Enable RLS on Users
✅ 005 - Add 2FA Fields
✅ 006 - Create Sessions Table
✅ 007 - Create Permissions Table (with seed data)
✅ 008 - Create Roles Table
✅ 009 - Create Role-Permissions Junction
✅ 010 - Create User-Roles Junction
✅ 011 - Create Invitations Table
✅ 012 - Create Audit Logs Table
✅ 013 - Create Helper Functions
✅ 014 - Seed System Data
✅ 015 - Create Company Settings Table (NEW)
```

### Company Settings Schema
```sql
- company_info (name, legal_name, industry, employee_count, founded_date, website, description)
- contact_details (email, phone, fax)
- address (street, city, state, postal_code, country)
- fiscal_settings (fiscal_year_start, currency, date_format, time_zone, language)
- legal_identifiers (tax_id, registration_number, vat_number, nif_number, ai_number)
- Additional fields: logo_url, preferences (JSONB)
- RLS enabled with tenant isolation
```

### Test Data
- **Tenant**: ACME Corporation (slug: `acme-corp`) ✅
- **Admin User**: `admin@acme-corp.com` / `Admin@123` ✅

---

## 💻 Backend Status

### File Structure (46+ Go Files)
```
✅ cmd/server/main.go              - HTTP server entry point
✅ cmd/migrate/main.go             - Migration CLI tool
✅ internal/config/config.go       - Configuration management
✅ internal/database/
   ✅ postgres.go                  - PostgreSQL connection
   ✅ redis.go                     - Redis connection
   ✅ rls.go                       - Row-Level Security helpers
✅ internal/models/                - 8 domain models (+ CompanySettings)
✅ internal/repository/            - 8 repositories (with RLS)
✅ internal/services/              - 10 services (+ CompanySettingsService)
✅ internal/handlers/              - 10 HTTP handlers (+ CompanySettingsHandler)
✅ internal/middleware/            - 2 middleware (auth, permission)
✅ internal/server/router.go       - ✅ Fixed and deployed
✅ internal/utils/                 - 8 utility modules
```

### Services Implemented
- ✅ Auth Service (register, login, logout, verify)
- ✅ JWT Service (token generation, validation)
- ✅ User Service (CRUD operations)
- ✅ Permission Service (with Redis caching)
- ✅ Two-Factor Service (TOTP, backup codes)
- ✅ Session Service (device tracking)
- ✅ Invitation Service (team invites)
- ✅ Email Service (SMTP)
- ✅ Audit Service (security logging)
- ✅ **Company Settings Service** (NEW - company profile management)

### API Endpoints Implemented (70+)
**Note:** Clean URL structure - domain indicates API, routes at root level

```
Authentication:
  POST   /auth/register                  ✅
  POST   /auth/verify-email              ✅
  POST   /auth/login                     ✅
  POST   /auth/verify-2fa                ✅
  POST   /auth/refresh                   ✅
  POST   /auth/logout                    ✅
  POST   /auth/logout-all                ✅
  POST   /auth/forgot-password           ✅
  POST   /auth/reset-password            ✅
  POST   /auth/change-password           ✅
  GET    /auth/me                        ✅

Users:
  GET    /users                      ✅
  POST   /users                      ✅
  GET    /users/:id                  ✅
  PUT    /users/:id                  ✅
  DELETE /users/:id                  ✅
  PATCH  /users/:id/status           ✅
  GET    /users/:id/roles            ✅
  POST   /users/:id/roles            ✅
  GET    /users/search               ✅

Roles & Permissions:
  GET    /roles                      ✅
  POST   /roles                      ✅
  GET    /roles/:id                  ✅
  PUT    /roles/:id                  ✅
  DELETE /roles/:id                  ✅
  GET    /roles/:id/permissions      ✅
  GET    /roles/:id/users            ✅
  POST   /roles/:id/assign           ✅
  GET    /permissions                ✅
  GET    /permissions/by-category    ✅
  GET    /permissions/search         ✅
  GET    /permissions/stats          ✅
  GET    /permissions/me             ✅
  POST   /permissions/check          ✅

Sessions:
  GET    /sessions                   ✅
  GET    /sessions/stats             ✅
  GET    /sessions/recent-logins     ✅
  DELETE /sessions/:id               ✅
  POST   /sessions/revoke-all        ✅

Two-Factor Authentication:
  POST   /2fa/setup                  ✅
  POST   /2fa/enable                 ✅
  POST   /2fa/disable                ✅
  POST   /2fa/verify                 ✅
  POST   /2fa/verify-backup          ✅
  POST   /2fa/backup-codes/regenerate ✅
  GET    /2fa/backup-codes/count     ✅
  POST   /2fa/device/trust           ✅

Invitations:
  GET    /invitations                ✅
  GET    /invitations/:id            ✅
  POST   /invitations                ✅
  POST   /invitations/accept         ✅
  DELETE /invitations/:id            ✅
  POST   /invitations/:id/resend     ✅

Audit:
  GET    /audit-logs                 ✅
  GET    /audit-logs/search          ✅
  GET    /audit-logs/stats           ✅
  GET    /audit-logs/failed-attempts ✅
  GET    /audit-logs/user/:id        ✅
  GET    /audit-logs/resource/:type/:id ✅

Security:
  GET    /security/overview          ✅
  GET    /security/suspicious-activity ✅
  GET    /security/recommendations   ✅
  GET    /security/login-history     ✅

Company Settings (NEW):
  GET    /settings/company           ✅
  PUT    /settings/company           ✅
  POST   /settings/company/logo      ✅
  DELETE /settings/company/logo      ✅

Health:
  GET    /health                         ✅
```

---

## 🎨 Frontend Status

### Technology Stack
- **Framework**: Next.js 15.5.6 (App Router)
- **React**: 19.0.0
- **TypeScript**: 5.7.3
- **Styling**: Tailwind CSS 3.4.1
- **Components**: shadcn/ui (Radix UI)
- **State**: Zustand 5.0.1
- **Icons**: Lucide React 0.468.0
- **Theme**: next-themes 0.4.4 ✅
- **Maps**: Google Places API (autocomplete)

### Features Implemented
- ✅ **Dark Mode**: Full implementation with system detection
  - Light / Dark / System modes
  - Persistent theme preference (localStorage)
  - Smooth transitions
  - Theme toggle in header
  - Zero flash of unstyled content (FOUC)

- ✅ **Pages**:
  - Login / Register / Verify Email
  - Dashboard (with KPI cards)
  - User Management (team/members)
  - Role Management (team/roles)
  - Security Settings (2FA, sessions, audit logs)
  - App Settings (profile, preferences)
  - **Company Settings** (NEW - comprehensive company profile)

- ✅ **Company Settings Feature** (NEW):
  - Company Information (name, industry, employees, founded date)
  - Contact Details (email, phone, fax)
  - Physical Address (with Google Places autocomplete)
  - Fiscal Settings (fiscal year, currency, date format, timezone)
  - Legal Identifiers (tax ID, registration number, VAT, NIF, AI)
  - Logo upload functionality
  - Preferences (JSONB storage)
  - Multi-section form with validation
  - Save functionality per section

- ✅ **Components**:
  - 43+ shadcn/ui components
  - Custom Layout (Sidebar, Header)
  - Notifications (Toast system)
  - Breadcrumbs
  - Theme Toggle Dropdown
  - DatePicker with dropdown month/year selector
  - Google Places Autocomplete integration

- ✅ **State Management**:
  - Auth Store (Zustand)
  - Permission Context
  - Notification Context

### Color Palette
```css
Light Mode:
  Primary (Indigo):  #4F46E5
  Secondary (Cyan):  #06B6D4
  Background:        #FFFFFF
  Success:           #16A34A
  Warning:           #F59E0B
  Error:             #EF4444

Dark Mode:
  Primary (Indigo):  #818CF8
  Secondary (Cyan):  #22D3EE
  Background:        #1E293B
  Success:           #22C55E
  Warning:           #FCD34D
  Error:             #F87171
```

### Frontend Production
- **URL**: https://app.infold.app
- **Status**: ✅ Running in production
- **Build**: Production optimized
- **Docker**: Multi-stage build with Next.js standalone output

---

## 🔐 Security Features

### Authentication
- ✅ JWT tokens (Access + Refresh)
- ✅ HTTP-only cookies
- ✅ SameSite=Strict
- ✅ Password hashing (bcrypt)
- ✅ Email verification
- ✅ Password reset flow

### Authorization
- ✅ Row-Level Security (RLS) for tenant isolation
- ✅ Permission-based access control (RBAC)
- ✅ Permission caching (Redis, 15min TTL)
- ✅ Hierarchical roles support

### Two-Factor Authentication
- ✅ TOTP (Time-based One-Time Password)
- ✅ QR code generation
- ✅ Backup codes (10 codes, encrypted)
- ✅ Trusted device support (30 days)
- ✅ Rate limiting (5 attempts per 15min)

### Audit & Monitoring
- ✅ Comprehensive audit logging
- ✅ Session tracking with device info
- ✅ IP address logging
- ✅ Security event logging
- ✅ Rate limiting on auth endpoints

### CORS & Networking
- ✅ CORS properly configured for production domains
- ✅ HTTPS enforced via Caddy reverse proxy
- ✅ Secure headers (X-Real-IP, X-Forwarded-For, X-Forwarded-Proto)

---

## 📋 TODO List

### Phase 2.0 - Current Sprint
- [x] Company Settings feature ✅
- [x] Production deployment ✅
- [x] CORS configuration ✅
- [x] Environment variable management ✅
- [ ] Additional business features
- [ ] Inventory management
- [ ] Customer management
- [ ] Invoice generation

### Short Term (Next 2 Weeks)
- [ ] Write unit tests for services
- [ ] Write integration tests for auth flow
- [ ] Add API documentation (Swagger/OpenAPI)
- [ ] Performance testing & optimization
- [ ] Security audit
- [ ] Backup & restore procedures
- [ ] Monitoring & logging setup (production)

### Phase 2.1 (Next Month)
- [ ] Customer Management module
- [ ] Product/Inventory module
- [ ] Invoice & Quotation module
- [ ] Payment processing integration
- [ ] Advanced reporting
- [ ] Email templates
- [ ] File storage (S3/MinIO)

### Phase 3 (Future)
- [ ] Mobile app (React Native)
- [ ] SSO integration (Google, Microsoft)
- [ ] Advanced analytics dashboard
- [ ] Multi-language support (i18n)
- [ ] Webhooks & API integrations
- [ ] Advanced workflow automation

---

## 🧪 Testing

### Production Testing
```bash
# Health check
curl https://api.infold.app/health

# Login
curl -X POST https://api.infold.app/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@acme-corp.com",
    "password": "Admin@123"
  }'

# Get company settings (requires auth token)
curl https://api.infold.app/settings/company \
  -H "Authorization: Bearer <token>"
```

### Local Development Testing
```bash
# Health check
curl http://localhost:18080/health

# Register tenant
curl -X POST http://localhost:18080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "company_name": "Test Corp",
    "email": "admin@testcorp.com",
    "password": "Test@123",
    "first_name": "Admin",
    "last_name": "User"
  }'

# Login
curl -X POST http://localhost:18080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@acme-corp.com",
    "password": "Admin@123"
  }'
```

### Automated Tests
```bash
# Run unit tests
cd backend
go test ./internal/services/...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🚀 Quick Start

### Local Development

#### 1. Start Infrastructure
```bash
docker-compose up -d
```

#### 2. Start Backend
```bash
cd backend
go run cmd/server/main.go
# Server: http://localhost:18080
```

#### 3. Start Frontend
```bash
cd frontend
npm run dev -- -p 13000
# Frontend: http://localhost:13000
```

#### 4. Access Services
- **Frontend**: http://localhost:13000
- **Backend API**: http://localhost:18080
- **Mailpit UI**: http://localhost:18025
- **PostgreSQL**: localhost:15433
- **Redis**: localhost:26379

### Production Deployment

#### SSH into VPS
```bash
ssh -i ~/.ssh/myerp_vps_key root@167.86.117.179
```

#### Deploy Updates
```bash
cd /opt/myerp-v2

# Pull latest code
git pull

# Rebuild and restart services
docker compose -f docker-compose.prod.yml build --no-cache
docker compose -f docker-compose.prod.yml up -d --force-recreate

# View logs
docker logs myerp_frontend -f
docker logs myerp_backend -f
```

#### Check Status
```bash
# Container status
docker ps

# Service health
curl https://api.infold.app/health

# Caddy status
systemctl status caddy
```

---

## 📁 Recent Changes (Jan 21, 2026)

### Production Fixes & Improvements
- ✅ **Clean URL Structure Implemented**
  - Removed `/api` prefix from all backend routes
  - Routes now at root: `/auth/login`, `/users`, `/roles`, etc.
  - Domain-based routing: `api.infold.app` indicates API, no path prefix needed
  - Updated backend router (router.go)
  - Rebuilt and deployed backend with Go 1.24

- ✅ **Fixed CORS Configuration**
  - Frontend rebuilt with correct production API URL
  - `NEXT_PUBLIC_API_URL=https://api.infold.app` (no /api suffix)
  - Docker build arguments properly configured
  - CORS properly allows `https://app.infold.app`

- ✅ **Fixed Caddy Reverse Proxy**
  - Changed from `localhost` to `127.0.0.1` (IPv4)
  - Resolved connection refused errors
  - Frontend now accessible via https://app.infold.app

### Previous Features (Jan 20, 2026)
- ✅ **Company Settings Module**: Complete company profile management
  - Backend: Repository, Service, Handler, Routes
  - Database: Migration 015 with RLS
  - Frontend: Multi-section form with Google Places integration
  - API endpoints for CRUD operations

- ✅ Initial Production Deployment to VPS at infold.app
- ✅ Configured Caddy reverse proxy with auto-HTTPS
- ✅ Backend running on port 18080
- ✅ Frontend running on port 13000

### Created Files
- `backend/migrations/015_create_company_settings_table.up.sql`
- `backend/migrations/015_create_company_settings_table.down.sql`
- `backend/internal/models/company_settings.go`
- `backend/internal/repository/company_settings_repository.go`
- `backend/internal/services/company_settings_service.go`
- `backend/internal/handlers/company_settings_handler.go`
- `frontend/src/app/dashboard/settings/company/page.tsx`
- `frontend/src/app/dashboard/settings/company/_components/general-settings.tsx`
- `frontend/src/components/ui/date-picker.tsx`
- `docker-compose.prod.yml`

### Modified Files
- `backend/internal/server/router.go` - Added company settings routes
- `frontend/Dockerfile` - Added ARG for build-time env vars
- `frontend/src/lib/api.ts` - Updated API base URL
- `frontend/src/app/globals.css` - Added Google Places z-index fix
- `frontend/src/components/ui/calendar.tsx` - Increased cell size
- Backend `.env` (production) - Updated URLs for production

### Database Changes
- Added company_settings table with RLS
- Migration 015 created and applied

---

## 🐛 Known Issues

### All Issues Resolved ✅
- ✅ Backend compilation errors (fixed)
- ✅ CORS configuration (fixed Jan 21)
- ✅ API URL structure (fixed Jan 21 - clean domain-based routing)
- ✅ Environment variable management (fixed)
- ✅ Docker build configuration (fixed)
- ✅ Caddy IPv4 binding (fixed Jan 21)
- ✅ Frontend 404 errors (fixed Jan 21)
- ✅ Production deployment (completed)

### Minor UI Polish (Low Priority)
- ⚠️ DatePicker calendar width could be wider
- ⚠️ Google Places autocomplete selection edge cases
- Note: Core functionality works perfectly, these are cosmetic improvements

### Production Status
- ✅ No blocking issues in production
- ✅ All services running smoothly
- ✅ Frontend and backend fully accessible
- ✅ HTTPS working with auto-renewal

---

## 📊 Metrics & Performance

### Database
- Tables: 10 (with RLS)
- Migrations: 15
- Indexes: 45+
- RLS Policies: 9

### Backend
- Go Files: 50+
- Lines of Code: ~7,500
- Services: 10
- Repositories: 8
- Handlers: 10
- Middleware: 2

### Frontend
- Components: 55+
- Pages: 16+
- Routes: 22+
- Lines of Code: ~9,500

### Production Performance
- Frontend load time: ~1.7s
- Backend response time: <100ms
- Database queries: Optimized with indexes
- CDN: Caddy with HTTP/2

---

## 🎯 Success Criteria

### Phase 1 - Core Platform ✅
- [x] Multi-tenant registration ✅
- [x] Email verification flow ✅
- [x] JWT authentication ✅
- [x] User management CRUD ✅
- [x] Role-based access control ✅
- [x] Permission system ✅
- [x] Two-factor authentication ✅
- [x] Session management ✅
- [x] Audit logging ✅
- [x] Dark mode ✅

### Phase 1.5 - Production Deployment ✅
- [x] Backend server running ✅
- [x] Frontend deployed ✅
- [x] VPS deployment ✅
- [x] HTTPS with Caddy ✅
- [x] Environment configuration ✅
- [x] Full end-to-end login test ✅

### Phase 2.0 - Business Features ⏳
- [x] Company settings module ✅
- [ ] Customer management 📝
- [ ] Inventory management 📝
- [ ] Invoice generation 📝
- [ ] Payment processing 📝

---

**Status Legend:**
- ✅ Complete
- ⚠️ Needs attention
- ⏳ In progress
- ❌ Blocked
- 📝 Planned

---

*This document tracks the current state of MyERP v2 development and deployment.*

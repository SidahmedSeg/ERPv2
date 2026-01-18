# MyERP v2 - Claude Code Guide

**Project:** Multi-Tenant ERP System (Shared Schema + Row-Level Security)
**Status:** Phase 1.5 - UI Enhancement in Progress
**Phase 1 Complete:** Backend (Auth + RBAC) ✅ | Frontend (Basic UI) ✅

---

## 🎯 Big Picture Architecture

### Multi-Tenancy: Shared Schema + Row-Level Security (RLS)

Unlike traditional schema-per-tenant approaches, MyERP v2 uses **PostgreSQL Row-Level Security** for tenant isolation:

- **One shared schema** for all tenants (not N schemas)
- **RLS policies** automatically filter data by `tenant_id`
- **Transaction-level context** set via `SET LOCAL app.current_tenant_id`
- **Critical pattern:** Every tenant-scoped query MUST use `WithTenantContext()`

**Why this matters:**
- ✅ 10x faster migrations (run once, not per tenant)
- ✅ 91% lower memory overhead
- ✅ Simpler backups (one DB)
- ⚠️ Security-critical: RLS bypass = data leakage across tenants

---

## 🔑 Critical RLS Pattern

**NEVER query tenant-scoped tables directly.** Always use the RLS context wrapper:

```go
// ✅ CORRECT - RLS enforced automatically
tx, err := database.WithTenantContext(ctx, db, tenantID)
if err != nil {
    return err
}
defer tx.Rollback()

// All queries within this transaction are filtered by tenant_id
_, err = tx.ExecContext(ctx, "INSERT INTO users (email, ...) VALUES ($1, ...)", email)
if err != nil {
    return err
}

return tx.Commit()
```

```go
// ❌ WRONG - No RLS context, will fail or leak data
_, err := db.ExecContext(ctx, "INSERT INTO users ...")
```

**Key files:**
- `backend/internal/database/rls.go` - RLS helpers (`WithTenantContext`, `WithBypassRLS`)
- `backend/internal/middleware/auth.go` - Sets `tenant_id` in request context from JWT

**RLS-enabled tables:** users, roles, role_permissions, user_roles, sessions, invitations, audit_logs
**No RLS:** tenants (global registry), permissions (global catalog)

---

## 🔐 Authentication & Authorization Flow

### 1. JWT Token Chain

```
Login → Access Token (15min) + Refresh Token (7 days)
       ↓
   2FA Required? → 2FA Token (5min) → Verify TOTP → Access Token
       ↓
   Set Cookie (HTTP-only, SameSite=Strict)
```

**Token types (see `internal/services/jwt_service.go`):**
- **Access Token:** Short-lived (15min), carries `tenant_id`, `user_id`, `tenant_slug`
- **Refresh Token:** Long-lived (7 days), stored in Redis with session hash
- **2FA Token:** Temporary (5min), issued when 2FA required before full access

### 2. Permission Checking with Redis Cache

```
Request → Auth Middleware → Permission Middleware → Handler
          ↓                  ↓
     Extract JWT         Check Redis cache (15min TTL)
     Set context         ↓           ↓
     (tenant_id,     Cache HIT   Cache MISS
      user_id)        Return      ↓
                                 Query DB (has_permission function)
                                 Cache result
```

**Key files:**
- `backend/internal/middleware/auth.go` - Validates JWT, injects tenant/user into context
- `backend/internal/middleware/permission.go` - Checks permissions (cached)
- `backend/internal/services/permission_service.go` - Permission logic with Redis caching

**Cache invalidation:** When roles or permissions change, flush `perms:{tenant_id}:{user_id}` from Redis

---

## 📂 Backend Architecture Layers

```
HTTP Request
    ↓
Middleware Chain (auth → permission → handler)
    ↓
Handler (validates input, calls service)
    ↓
Service (business logic, transactions)
    ↓
Repository (database access with RLS)
    ↓
PostgreSQL (RLS policies auto-filter)
```

### Directory Structure

```
backend/
├── cmd/
│   ├── server/main.go          # HTTP server entry point
│   └── migrate/main.go         # Migration CLI
├── internal/
│   ├── config/                 # Environment config loader
│   ├── database/               # PostgreSQL + Redis + RLS helpers
│   ├── middleware/             # Auth, permission, CORS, rate limiting
│   ├── models/                 # Domain models (User, Role, Permission, etc.)
│   ├── repository/             # Data access with RLS context
│   ├── services/               # Business logic (Auth, JWT, 2FA, Email, etc.)
│   ├── handlers/               # HTTP handlers (auth, user, role, etc.)
│   ├── utils/                  # Response, validation, crypto, slug, device
│   └── server/                 # Router setup
└── migrations/                 # SQL migrations (14 files)
```

### Creating a New Repository (with RLS)

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "github.com/jmoiron/sqlx"
    "myerp-v2/internal/database"
    "myerp-v2/internal/models"
)

type ExampleRepository struct {
    db *sqlx.DB
}

func (r *ExampleRepository) Create(ctx context.Context, tenantID uuid.UUID, item *models.Example) error {
    // CRITICAL: Start transaction with RLS context
    tx, err := database.WithTenantContext(ctx, r.db, tenantID)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    query := `INSERT INTO examples (tenant_id, name) VALUES ($1, $2) RETURNING id`
    err = tx.QueryRowContext(ctx, query, tenantID, item.Name).Scan(&item.ID)
    if err != nil {
        return err
    }

    return tx.Commit()
}

func (r *ExampleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]models.Example, error) {
    tx, err := database.WithTenantContext(ctx, r.db, tenantID)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    var items []models.Example
    // RLS automatically filters by tenant_id
    err = tx.SelectContext(ctx, &items, "SELECT * FROM examples ORDER BY created_at DESC")
    return items, err
}
```

### Adding a New API Endpoint with Permission Check

```go
// 1. Define route in server/router.go
r.Route("/api/examples", func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    r.Use(permMiddleware.RequirePermission("examples", "view"))
    r.Get("/", exampleHandler.List)

    r.With(permMiddleware.RequirePermission("examples", "create")).Post("/", exampleHandler.Create)
})

// 2. Create handler in handlers/example_handler.go
func (h *ExampleHandler) List(w http.ResponseWriter, r *http.Request) {
    tenantID, _ := database.GetTenantIDFromContext(r.Context())

    items, err := h.exampleService.List(r.Context(), tenantID)
    if err != nil {
        utils.InternalServerError(w, "Failed to fetch examples")
        return
    }

    utils.Success(w, items)
}
```

---

## 🗄️ Database Migrations

**Tool:** golang-migrate
**Location:** `backend/migrations/`
**Count:** 14 migrations (extensions → tenants → users → RLS → roles → permissions → audit)

### Commands

```bash
# Run all migrations up
go run cmd/migrate/main.go up

# Rollback one migration
go run cmd/migrate/main.go down

# Rollback N migrations
go run cmd/migrate/main.go down 3

# Create new migration
migrate create -ext sql -dir migrations -seq add_new_table
```

### Key Migrations

- **001-003:** Extensions, Tenants, Users with composite PK (`tenant_id`, `id`)
- **004:** Enable RLS on users table with policies
- **005:** Add 2FA fields (TOTP secret, backup codes, trusted devices)
- **006:** Sessions table with device tracking
- **007:** Permissions catalog with seed data (users.view, users.create, etc.)
- **008-010:** Roles, role_permissions, user_roles with RLS
- **011:** Invitations system
- **012:** Audit logs with JSONB metadata
- **013:** Helper functions (`get_user_permissions()`, `has_permission()`)
- **014:** Seed system roles (owner, admin, manager, user)

---

## 🧪 Testing

```bash
# Backend tests
cd backend
go test ./...                          # All tests
go test -v ./internal/services/...     # Specific package
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend tests
cd frontend
npm test                               # Run tests
npm run build                          # Production build
```

**Integration tests location:** `backend/tests/integration/`
**Security tests:** RLS enforcement, permission checks, rate limiting

---

## 🚀 Development Workflow

### 1. Start Infrastructure

```bash
# From project root
docker-compose up -d

# Services:
# - PostgreSQL 16: localhost:5432 (user: myerp, pass: myerp_password, db: myerp_v2)
# - Redis 7: localhost:6379 (pass: redis_password)
# - Mailpit: localhost:8025 (email UI), localhost:1025 (SMTP)
```

### 2. Run Backend

```bash
cd backend

# Copy environment file (first time only)
cp .env.example .env

# Install dependencies
go mod download

# Run migrations
go run cmd/migrate/main.go up

# Start server (http://localhost:8080)
go run cmd/server/main.go
```

### 3. Run Frontend

```bash
cd frontend

# Install dependencies (first time only)
npm install

# Start dev server (http://localhost:13000)
npm run dev -- -p 13000

# Production build
npm run build
npm run start
```

### 4. Check Emails (Mailpit)

Open http://localhost:8025 to see all emails sent by the application (verification, password reset, invitations).

---

## 🎨 Frontend: Phase 1.5 UI Enhancement (In Progress)

### Current Status: Recreating UI to Match Original MyERP

**Original MyERP:** `/Users/intelifoxdz/myerp-project/frontend/`
**New MyERP v2:** `/Users/intelifoxdz/Zyndra/myerp-v2/frontend/`

**Goal:** Replace basic Tailwind UI with shadcn/ui component library matching original MyERP's polished design.

### Key UI Patterns from Original MyERP

1. **shadcn/ui Components** (43 components):
   - Radix UI primitives + Tailwind styling
   - `components/ui/` - button, input, card, dialog, sheet, dropdown-menu, etc.

2. **DataTable Pattern** (NOT plain HTML tables):
   - Column sorting with indicators (ASC/DESC)
   - Bulk selection (row checkboxes)
   - Bulk actions toolbar (shows when items selected)
   - Custom cell renderers
   - Pagination controls
   - Empty states, loading skeletons

3. **Sheet Components** (NOT modal dialogs):
   - Slide-out side panels for forms (Add User, Edit Role, etc.)
   - Better UX than center modals for forms

4. **Layout:**
   - Fixed sidebar: 256px width, company selector, nav menu, settings at bottom
   - Sticky header: 64px height, breadcrumbs, search, notifications, user dropdown
   - Content area: Proper spacing, responsive

5. **Toast System:**
   - Sonner library for global toast notifications
   - Success, error, warning, info variants

### Design Tokens (Tailwind Config)

```typescript
// Primary Blue: #2563EB (600)
// Secondary Slate: #1E293B (800)
// Success Green: #22C55E (500)
// Warning Yellow: #EAB308 (500)
// Error Red: #ED4343 (600)
// Background: #FAFAFA
// Border: #E4E4E7
```

### Phase 1.5 Progress Checklist

**Phase 1: Foundation** (60% Complete)
- [x] Install Radix UI deps (43 packages)
- [x] Configure Tailwind with exact colors
- [ ] Set up `lib/utils.ts` (cn helper)
- [x] Configure next-themes (dark mode) ✅
- [x] Create ThemeToggle component ✅
- [x] Integrate theme toggle in Header ✅
- [ ] Set up Sonner (toast system)

**Phase 2: Core UI Components** (~20 components, 0% Complete)
- [ ] button, input, label, card, dialog, sheet, dropdown-menu, select, tabs, accordion, badge, separator, scroll-area, checkbox, switch, popover, alert-dialog, skeleton, avatar, progress

**Phase 3: Advanced Components** (0% Complete)
- [ ] data-table.tsx - Advanced table with sorting, bulk selection, bulk actions
- [ ] status-badge.tsx, confirm-dialog.tsx, table-pagination.tsx, table-filter.tsx

**Phase 4: Layout Components** (0% Complete)
- [ ] sidebar.tsx, header.tsx, dashboard-layout.tsx

**Phase 5: Page Recreation** (0% Complete)
- [ ] Users page (DataTable + Sheet forms)
- [ ] Roles page (Card layout + Permission groups)
- [ ] Security page (Tabs: Overview, 2FA, Sessions, Audit)
- [ ] Settings page (Tabs: Account, Security, Preferences)
- [ ] Dashboard homepage (KPI cards + Quick actions)

**Phase 6: Polish** (0% Complete)
- [ ] Toast notifications, confirmations, skeletons, empty states, responsive design, dark mode, transitions

**Estimated time:** 2-3 days full implementation

---

## 📚 API Endpoints (64+ endpoints)

### Auth (`/api/auth/*`)
- POST `/register` - Tenant registration
- POST `/verify-email` - Email verification
- POST `/login` - User login (returns 2FA token if enabled)
- POST `/verify-2fa` - Verify TOTP code
- POST `/refresh` - Refresh access token
- POST `/logout` - Logout and invalidate session
- POST `/forgot-password` - Request password reset
- POST `/reset-password` - Reset password with token

### Users (`/api/users/*`)
- GET `/` - List users (paginated, filtered)
- POST `/` - Create user (requires `users.create`)
- GET `/:id` - Get user by ID
- PUT `/:id` - Update user (requires `users.edit`)
- DELETE `/:id` - Delete user (requires `users.delete`)
- PUT `/:id/status` - Update user status (requires `users.manage_status`)
- GET `/:id/roles` - Get user's roles
- PUT `/:id/roles` - Assign roles to user (requires `roles.assign`)

### Roles (`/api/roles/*`)
- GET `/` - List roles
- POST `/` - Create role (requires `roles.create`)
- GET `/:id` - Get role by ID
- PUT `/:id` - Update role (requires `roles.edit`)
- DELETE `/:id` - Delete role (requires `roles.delete`)
- GET `/:id/permissions` - Get role permissions
- PUT `/:id/permissions` - Update role permissions (requires `roles.edit`)
- GET `/:id/users` - Get users with role

### Permissions (`/api/permissions/*`)
- GET `/` - List all permissions
- GET `/categories` - List permission categories
- GET `/check` - Check if user has permission

### Sessions (`/api/sessions/*`)
- GET `/` - List active sessions
- DELETE `/:id` - Revoke session
- GET `/stats` - Session statistics

### 2FA (`/api/2fa/*`)
- POST `/setup` - Generate TOTP secret + QR code
- POST `/enable` - Enable 2FA with verification
- POST `/disable` - Disable 2FA
- POST `/verify-backup-code` - Verify backup code

**Full API docs:** See `backend/docs/API.md`

---

## 🔒 Security Checklist

When adding new features:

- [ ] **RLS Context:** All tenant-scoped queries use `WithTenantContext()`
- [ ] **Permission Check:** Sensitive endpoints have `RequirePermission()` middleware
- [ ] **Input Validation:** All user inputs validated (email, password strength, etc.)
- [ ] **SQL Injection:** Use parameterized queries (`$1, $2, ...`)
- [ ] **XSS Prevention:** Never render unsanitized user input
- [ ] **Audit Logging:** Log sensitive operations (create user, delete role, etc.)
- [ ] **Rate Limiting:** Auth endpoints protected (login, 2FA, password reset)
- [ ] **Error Messages:** Don't leak sensitive info (e.g., "user not found" vs "invalid credentials")

---

## 🐛 Common Issues & Solutions

### "RLS policy violation" error
**Cause:** Forgot to set tenant context before query
**Fix:** Wrap query in `WithTenantContext(ctx, db, tenantID)`

### "tenant_id not found in context" error
**Cause:** Missing auth middleware or JWT doesn't contain tenant_id
**Fix:** Ensure route uses `authMiddleware.Authenticate` and JWT is valid

### Permission cache not invalidating
**Cause:** Didn't flush Redis cache after role/permission change
**Fix:** Call `redis.Del(ctx, fmt.Sprintf("perms:%s:%s", tenantID, userID))`

### Migration fails with "already exists"
**Cause:** Database state doesn't match migration version
**Fix:** Check `schema_migrations` table, manually fix state, or use `migrate force VERSION`

### Can't login after 2FA enabled
**Cause:** Device fingerprint changed or TOTP clock drift
**Fix:** Use backup code or reset 2FA (admin operation)

---

## 📊 Performance Notes

### Database Indexes
All tenant-scoped tables have:
- Composite index on (`tenant_id`, `id`) for primary key
- Individual index on `tenant_id` for RLS performance
- Additional indexes for common queries (email, status, etc.)

### Redis Caching Strategy
- **Permissions:** 15-min TTL, invalidate on role/permission changes
- **Sessions:** 7-day expiry (or 30-day with "remember me")
- **Rate Limiting:** 5min-15min windows depending on endpoint
- **Trusted Devices (2FA):** 30-day expiry

### Connection Pooling
```go
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(time.Hour)
db.SetConnMaxIdleTime(10 * time.Minute)
```

---

## 🔗 Related Documents

- **README.md** - Getting started guide, environment setup
- **PROJECT_STATUS.md** - Current phase, progress tracking
- **backend/docs/API.md** - Complete API reference (all 64+ endpoints)
- **DEPLOYMENT.md** - Production deployment guide
- **Plan Document** - `/Users/intelifoxdz/.claude/plans/scalable-exploring-reef.md`
- **Original MyERP** - `/Users/intelifoxdz/myerp-project/` (reference for UI patterns)

---

## 🎯 Current Work: Phase 1.5 UI Enhancement

**What's happening:** Recreating all frontend pages to match the original MyERP's polished UI with shadcn/ui components, DataTable, Sheet forms, and advanced features.

**What's done (Phase 1 & 1.5):**
- ✅ Backend complete (Auth, RBAC, 2FA, Sessions, Audit)
- ✅ Frontend basic UI (login, dashboard, users, roles, security, settings)
- ✅ Dependencies installed (Radix UI packages + next-themes)
- ✅ Tailwind configured with exact color system
- ✅ Dark mode implementation (next-themes + ThemeToggle) 🎨

**What's next:**
1. Create `lib/utils.ts` with `cn()` helper
2. Set up Sonner (toast system)
3. Create 20+ shadcn/ui components
4. Build advanced DataTable component
5. Recreate all pages with new components

**Reference:** Original MyERP at `/Users/intelifoxdz/myerp-project/frontend/` has the exact UI patterns to replicate.

---

**Last Updated:** January 17, 2026
**Version:** 2.0.0
**Backend Status:** Phase 1 Complete ✅
**Frontend Status:** Phase 1.5 In Progress (60% - Dark Mode ✅)

# MyERP v2 - Multi-Tenant ERP System

A modern, scalable multi-tenant ERP system built with Go, PostgreSQL (RLS), Redis, and Next.js.

## 🏗️ Architecture

### Multi-Tenancy Strategy: Shared Schema + Row-Level Security (RLS)

This project uses **PostgreSQL Row-Level Security (RLS)** for tenant isolation instead of the traditional schema-per-tenant approach. This provides:

- ✅ **Better Performance**: Single query plan cache, faster cross-tenant analytics
- ✅ **Easier Migrations**: Run migrations once, not N times per tenant
- ✅ **Simpler Backups**: One database instead of hundreds of schemas
- ✅ **Cost Efficient**: Lower memory overhead, better resource utilization
- ✅ **Horizontal Scaling**: Easier sharding by tenant_id

### Technology Stack

**Backend:**
- Go 1.24+
- PostgreSQL 16+ (with RLS)
- Redis 7+ (caching & sessions)
- Chi Router (HTTP routing)
- sqlx (database toolkit)
- golang-migrate (database migrations)

**Frontend:**
- Next.js 15+ (App Router)
- React 19
- TypeScript 5.7+
- Tailwind CSS
- shadcn/ui (component library)
- Zustand (state management)

**Infrastructure:**
- Docker & Docker Compose
- Mailpit (local email testing)

## 📋 Features (Phase 1: Core Auth & RBAC)

### Authentication & Authorization
- ✅ Multi-tenant registration with email verification
- ✅ JWT authentication with Redis sessions
- ✅ Enhanced 2FA (TOTP + backup codes + trusted devices)
- ✅ Password reset flow
- ✅ Session management with device tracking

### User Management
- ✅ User CRUD operations
- ✅ User roles and permissions
- ✅ Team invitations
- ✅ User status management (active, suspended, deactivated)

### RBAC (Role-Based Access Control)
- ✅ Hierarchical roles
- ✅ Dynamic permissions (resource.action pattern)
- ✅ Permission middleware with Redis caching
- ✅ System roles (Owner, Admin, Manager, User)
- ✅ Custom role creation

### Security
- ✅ Row-Level Security (RLS) for tenant isolation
- ✅ Comprehensive audit logging
- ✅ Rate limiting (login, 2FA)
- ✅ bcrypt password hashing
- ✅ AES-256-GCM encryption for sensitive data
- ✅ CSRF protection (SameSite cookies)

## 🚀 Getting Started

### Prerequisites

- Go 1.24 or higher
- Node.js 20+ and npm/yarn/pnpm
- Docker and Docker Compose
- PostgreSQL 16+ (via Docker)
- Redis 7+ (via Docker)

### Installation

1. **Clone the repository:**
   ```bash
   cd myerp-v2
   ```

2. **Start infrastructure services:**
   ```bash
   docker-compose up -d
   ```
   This starts PostgreSQL, Redis, and Mailpit.

3. **Set up backend:**
   ```bash
   cd backend

   # Copy environment file
   cp .env.example .env

   # Install Go dependencies
   go mod download

   # Run database migrations
   go run cmd/migrate/main.go up

   # Start the backend server
   go run cmd/server/main.go
   ```
   Backend will run on `http://localhost:8080`

4. **Set up frontend:**
   ```bash
   cd frontend

   # Install dependencies
   npm install
   # or
   yarn install
   # or
   pnpm install

   # Start the development server
   npm run dev
   ```
   Frontend will run on `http://localhost:3000`

5. **Access services:**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Mailpit (Email UI): http://localhost:8025
   - PostgreSQL: localhost:5432
   - Redis: localhost:6379

## 📁 Project Structure

```
myerp-v2/
├── backend/
│   ├── cmd/
│   │   ├── server/           # Main application entry point
│   │   └── migrate/          # Database migration CLI
│   ├── internal/
│   │   ├── config/           # Configuration management
│   │   ├── database/         # Database & RLS helpers
│   │   ├── middleware/       # HTTP middleware
│   │   ├── models/           # Domain models
│   │   ├── repository/       # Data access layer
│   │   ├── services/         # Business logic
│   │   ├── handlers/         # HTTP handlers
│   │   ├── utils/            # Utilities
│   │   └── server/           # HTTP server setup
│   ├── migrations/           # SQL migrations
│   ├── tests/                # Tests
│   ├── go.mod
│   └── .env.example
│
├── frontend/
│   ├── src/
│   │   ├── app/              # Next.js App Router
│   │   ├── components/       # React components
│   │   ├── lib/              # Utilities
│   │   ├── store/            # State management
│   │   └── types/            # TypeScript types
│   ├── public/
│   └── package.json
│
└── docker-compose.yml
```

## 🗄️ Database Schema

The database uses a **shared schema with Row-Level Security** for tenant isolation.

### Key Tables:

- **tenants**: Central tenant registry (no RLS)
- **users**: User accounts (RLS enabled)
- **sessions**: User sessions with device tracking (RLS enabled)
- **roles**: Tenant-specific roles (RLS enabled)
- **permissions**: Global permission catalog (no RLS)
- **role_permissions**: Role-permission mapping (RLS enabled)
- **user_roles**: User-role assignments (RLS enabled)
- **invitations**: Team invitation system (RLS enabled)
- **audit_logs**: Security event logging (RLS enabled)

### RLS Implementation:

Every tenant-scoped table includes:
```sql
-- Enable RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policy
CREATE POLICY tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Bypass for admin operations
CREATE POLICY bypass_rls_for_superuser ON users
    USING (current_setting('app.bypass_rls', true) = 'true')
    WITH CHECK (current_setting('app.bypass_rls', true) = 'true');
```

## 🔑 Environment Variables

### Backend (.env)

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=myerp
DB_PASSWORD=myerp_password
DB_NAME=myerp_v2
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=redis_password

# JWT
JWT_SECRET=your-jwt-secret-key-change-in-production
JWT_REFRESH_SECRET=your-jwt-refresh-secret-key-change-in-production

# Email (Mailpit for local development)
SMTP_HOST=localhost
SMTP_PORT=1025
EMAIL_FROM=noreply@myerp.local

# Security
ENCRYPTION_KEY=change-this-to-a-32-byte-key!!
BCRYPT_COST=10

# Application
APP_NAME=MyERP v2
APP_BASE_URL=http://localhost:8080
FRONTEND_URL=http://localhost:3000
```

### Frontend (.env.local)

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_APP_NAME=MyERP v2
```

## 🔒 Security Best Practices

### Row-Level Security (RLS)

Always set tenant context before queries:

```go
import "myerp-v2/internal/database"

// In your repository
tx, err := database.WithTenantContext(ctx, db, tenantID)
if err != nil {
    return err
}
defer tx.Rollback()

// All queries are automatically filtered by tenant_id
_, err = tx.ExecContext(ctx, "INSERT INTO users (...) VALUES (...)")
if err != nil {
    return err
}

return tx.Commit()
```

### Password Security
- bcrypt with cost factor 10
- Minimum 8 characters (12+ recommended)
- Must contain uppercase, lowercase, number, special character

### Session Security
- HTTP-only cookies
- SameSite=Strict
- Secure flag in production
- 7-day default expiry (30 days with "remember me")

### Rate Limiting
- Login: 5 attempts per 5 minutes
- 2FA: 5 attempts per 15 minutes
- Password reset: 3 attempts per hour

## 🧪 Testing

```bash
# Backend tests
cd backend
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Frontend tests
cd frontend
npm test
```

## 📚 API Documentation

API documentation will be available at `/api/docs` when Swagger is enabled:

```bash
ENABLE_SWAGGER=true go run cmd/server/main.go
```

Visit: http://localhost:8080/api/docs

## 🔄 Database Migrations

### Create a new migration:

```bash
cd backend
migrate create -ext sql -dir migrations -seq your_migration_name
```

### Run migrations:

```bash
# Up
go run cmd/migrate/main.go up

# Down
go run cmd/migrate/main.go down

# Down N steps
go run cmd/migrate/main.go down 2
```

## 🎯 Development Workflow

1. **Start infrastructure:**
   ```bash
   docker-compose up -d
   ```

2. **Run backend (terminal 1):**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

3. **Run frontend (terminal 2):**
   ```bash
   cd frontend
   npm run dev
   ```

4. **Check emails (Mailpit):**
   Open http://localhost:8025

5. **Stop infrastructure:**
   ```bash
   docker-compose down
   ```

## 📊 Performance Targets

- Login response time: <200ms (p95)
- Permission check (cached): <5ms
- Permission check (uncached): <50ms
- Database query response: <100ms (p95)
- API response time: <300ms (p95)

## 🗺️ Roadmap

### Phase 1: Core Auth & RBAC (Current)
- ✅ Multi-tenant registration
- ✅ Authentication & 2FA
- ✅ User management
- ✅ RBAC system
- ✅ Session management
- ✅ Audit logging

### Phase 2: Business Modules (Future)
- ⏳ Customer management
- ⏳ Product catalog
- ⏳ Invoice generation
- ⏳ Inventory tracking

### Phase 3: Advanced Features (Future)
- ⏳ File storage (ParaDrive integration)
- ⏳ Analytics & reporting
- ⏳ Payment processing
- ⏳ Email templates
- ⏳ API rate limiting

## 🐛 Troubleshooting

### Database Connection Issues

```bash
# Check if PostgreSQL is running
docker-compose ps

# View PostgreSQL logs
docker-compose logs postgres

# Restart PostgreSQL
docker-compose restart postgres
```

### Redis Connection Issues

```bash
# Check if Redis is running
docker-compose ps

# Test Redis connection
docker-compose exec redis redis-cli -a redis_password ping

# View Redis logs
docker-compose logs redis
```

### Migration Errors

```bash
# Check migration version
psql -h localhost -U myerp -d myerp_v2 -c "SELECT version FROM schema_migrations;"

# Force version (use with caution)
go run cmd/migrate/main.go force VERSION_NUMBER
```

## 📝 License

[Your License Here]

## 👥 Contributing

[Your Contributing Guidelines Here]

## 📧 Support

For issues and questions, please open an issue on GitHub or contact [your-email@example.com].

---

**Note:** This is Phase 1 focusing on core authentication and RBAC. Business modules (Customers, Products, Invoices, etc.) will be added in Phase 2.

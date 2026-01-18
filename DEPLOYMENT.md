# MyERP v2 Deployment Guide

**Version:** 1.0.0
**Last Updated:** January 17, 2026

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Environment Setup](#environment-setup)
3. [Database Setup](#database-setup)
4. [Backend Deployment](#backend-deployment)
5. [Frontend Deployment](#frontend-deployment)
6. [Docker Deployment](#docker-deployment)
7. [Production Checklist](#production-checklist)
8. [Monitoring & Logging](#monitoring--logging)
9. [Backup & Recovery](#backup--recovery)
10. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Software
- **Go** 1.24 or higher
- **Node.js** 20.x or higher
- **PostgreSQL** 16 or higher
- **Redis** 7 or higher
- **Docker** & **Docker Compose** (for containerized deployment)
- **Git**

### Required Tools
- `golang-migrate` for database migrations
- `npm` or `pnpm` for frontend dependencies

---

## Environment Setup

### 1. Clone the Repository

```bash
git clone https://github.com/your-org/myerp-v2.git
cd myerp-v2
```

### 2. Create Environment Files

**Backend** (`backend/.env`):
```bash
# Database
DATABASE_URL=postgresql://myerp:YOUR_PASSWORD@localhost:5432/myerp_v2
POSTGRES_USER=myerp
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD
POSTGRES_DB=myerp_v2

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=YOUR_REDIS_PASSWORD

# JWT
JWT_SECRET=YOUR_64_CHAR_RANDOM_STRING_HERE_MINIMUM_32_CHARACTERS_REQUIRED

# Email (Production - use real SMTP)
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=YOUR_SENDGRID_API_KEY
SMTP_FROM=noreply@yourdomain.com

# Application
PORT=8080
ENVIRONMENT=production
BASE_DOMAIN=yourdomain.com
CORS_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Security
BCRYPT_COST=12
RATE_LIMIT_ENABLED=true
```

**Frontend** (`frontend/.env.production`):
```bash
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

### 3. Generate Secrets

```bash
# Generate JWT secret (64 characters)
openssl rand -hex 32

# Generate PostgreSQL password
openssl rand -base64 32

# Generate Redis password
openssl rand -base64 24
```

---

## Database Setup

### 1. Install PostgreSQL

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql-16 postgresql-contrib
```

**macOS:**
```bash
brew install postgresql@16
brew services start postgresql@16
```

### 2. Create Database and User

```bash
sudo -u postgres psql

# In PostgreSQL prompt:
CREATE USER myerp WITH PASSWORD 'YOUR_SECURE_PASSWORD';
CREATE DATABASE myerp_v2 OWNER myerp;
GRANT ALL PRIVILEGES ON DATABASE myerp_v2 TO myerp;
\q
```

### 3. Install golang-migrate

```bash
# macOS
brew install golang-migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

### 4. Run Database Migrations

```bash
cd backend
export DATABASE_URL="postgresql://myerp:YOUR_PASSWORD@localhost:5432/myerp_v2?sslmode=disable"

# Run all migrations
migrate -path migrations -database "$DATABASE_URL" up

# Verify migrations
migrate -path migrations -database "$DATABASE_URL" version
```

### 5. Verify Database Setup

```bash
psql -U myerp -d myerp_v2

# Check tables
\dt

# Verify RLS is enabled
SELECT tablename, rowsecurity FROM pg_tables WHERE schemaname = 'public';

# Exit
\q
```

---

## Backend Deployment

### 1. Install Dependencies

```bash
cd backend
go mod download
go mod verify
```

### 2. Build Application

```bash
# Development build
go build -o bin/server cmd/server/main.go

# Production build (optimized)
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bin/server cmd/server/main.go
```

### 3. Run Tests

```bash
# Unit tests
go test ./internal/...

# Integration tests (requires test database)
go test -tags=integration ./tests/integration/...

# With coverage
go test -coverprofile=coverage.out ./internal/...
go tool cover -html=coverage.out
```

### 4. Start Backend Server

**Development:**
```bash
./bin/server
```

**Production (with systemd):**

Create `/etc/systemd/system/myerp-backend.service`:
```ini
[Unit]
Description=MyERP v2 Backend API
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=myerp
WorkingDirectory=/opt/myerp-v2/backend
EnvironmentFile=/opt/myerp-v2/backend/.env
ExecStart=/opt/myerp-v2/backend/bin/server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

Enable and start service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable myerp-backend
sudo systemctl start myerp-backend
sudo systemctl status myerp-backend
```

---

## Frontend Deployment

### 1. Install Dependencies

```bash
cd frontend
npm install
# or
pnpm install
```

### 2. Build Production Bundle

```bash
npm run build
# or
pnpm build
```

This creates an optimized production build in `.next/` directory.

### 3. Deploy with PM2 (Recommended)

```bash
# Install PM2 globally
npm install -g pm2

# Start application
pm2 start npm --name "myerp-frontend" -- start

# Save PM2 process list
pm2 save

# Setup startup script
pm2 startup systemd
```

### 4. Deploy with Static Hosting

**For Vercel:**
```bash
npm install -g vercel
vercel --prod
```

**For Netlify:**
```bash
npm install -g netlify-cli
netlify deploy --prod
```

### 5. Nginx Configuration

Create `/etc/nginx/sites-available/myerp`:
```nginx
# Frontend
server {
    listen 80;
    server_name yourdomain.com;

    # Redirect to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Backend API
server {
    listen 80;
    server_name api.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/api.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/api.yourdomain.com/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable site:
```bash
sudo ln -s /etc/nginx/sites-available/myerp /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 6. SSL Certificates (Let's Encrypt)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificates
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com -d api.yourdomain.com

# Auto-renewal is set up automatically
# Verify with:
sudo certbot renew --dry-run
```

---

## Docker Deployment

### 1. Build Docker Images

**Backend Dockerfile** (`backend/Dockerfile`):
```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080
CMD ["./server"]
```

**Frontend Dockerfile** (`frontend/Dockerfile`):
```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000
CMD ["node", "server.js"]
```

### 2. Docker Compose Production

**`docker-compose.prod.yml`:**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: myerp-postgres
    environment:
      POSTGRES_USER: ${POSTGRES_USER}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
      POSTGRES_DB: ${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - myerp-network
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: myerp-redis
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    networks:
      - myerp-network
    restart: unless-stopped

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: myerp-backend
    env_file:
      - ./backend/.env
    depends_on:
      - postgres
      - redis
    networks:
      - myerp-network
    ports:
      - "8080:8080"
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: myerp-frontend
    environment:
      NEXT_PUBLIC_API_URL: https://api.yourdomain.com
    depends_on:
      - backend
    networks:
      - myerp-network
    ports:
      - "3000:3000"
    restart: unless-stopped

networks:
  myerp-network:
    driver: bridge

volumes:
  postgres_data:
  redis_data:
```

### 3. Deploy with Docker Compose

```bash
# Build images
docker-compose -f docker-compose.prod.yml build

# Start services
docker-compose -f docker-compose.prod.yml up -d

# View logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop services
docker-compose -f docker-compose.prod.yml down
```

---

## Production Checklist

### Security
- [ ] Change all default passwords
- [ ] Generate new JWT secret (minimum 32 characters)
- [ ] Enable HTTPS with valid SSL certificates
- [ ] Configure CORS with production domains only
- [ ] Enable rate limiting
- [ ] Set secure cookie flags (HttpOnly, Secure, SameSite)
- [ ] Configure CSP headers
- [ ] Enable database connection encryption (SSL/TLS)
- [ ] Restrict database access to application servers only
- [ ] Disable debug mode and verbose logging
- [ ] Set up firewall rules
- [ ] Regular security audits

### Performance
- [ ] Enable database query caching
- [ ] Configure Redis connection pooling
- [ ] Set up CDN for static assets
- [ ] Enable Gzip/Brotli compression
- [ ] Configure database indexes
- [ ] Set up database connection pooling
- [ ] Enable HTTP/2
- [ ] Optimize Next.js bundle size

### Monitoring
- [ ] Set up application logging
- [ ] Configure error tracking (Sentry, Rollbar)
- [ ] Set up uptime monitoring
- [ ] Configure database performance monitoring
- [ ] Set up alerts for critical errors
- [ ] Monitor disk space and memory usage
- [ ] Track API response times

### Backup
- [ ] Configure automated database backups
- [ ] Test backup restoration procedure
- [ ] Set up Redis persistence
- [ ] Configure log rotation
- [ ] Store backups in separate location

---

## Monitoring & Logging

### Application Logging

Configure structured logging in production:

```go
// backend/internal/utils/logger.go
import "github.com/sirupsen/logrus"

var log = logrus.New()

func InitLogger() {
    log.SetFormatter(&logrus.JSONFormatter{})
    log.SetLevel(logrus.InfoLevel)

    if os.Getenv("ENVIRONMENT") == "development" {
        log.SetLevel(logrus.DebugLevel)
    }
}
```

### Health Check Endpoint

```bash
curl https://api.yourdomain.com/health
```

Expected response:
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "version": "1.0.0",
  "uptime": "72h15m30s"
}
```

### Log Aggregation

**Using Loki + Grafana:**
```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  loki:
    image: grafana/loki:latest
    ports:
      - "3100:3100"
    volumes:
      - ./loki-config.yaml:/etc/loki/config.yaml
      - loki_data:/loki

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  loki_data:
  grafana_data:
```

---

## Backup & Recovery

### Database Backup

**Automated daily backup script** (`/opt/myerp-v2/scripts/backup-db.sh`):
```bash
#!/bin/bash
BACKUP_DIR="/opt/myerp-v2/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/myerp_v2_$DATE.sql.gz"

# Create backup
pg_dump -U myerp myerp_v2 | gzip > "$BACKUP_FILE"

# Keep only last 30 days
find "$BACKUP_DIR" -name "*.sql.gz" -mtime +30 -delete

echo "Backup completed: $BACKUP_FILE"
```

**Cron job** (add to crontab):
```bash
# Daily backup at 2 AM
0 2 * * * /opt/myerp-v2/scripts/backup-db.sh >> /var/log/myerp-backup.log 2>&1
```

### Database Restore

```bash
# Restore from backup
gunzip < /opt/myerp-v2/backups/myerp_v2_20260117_020000.sql.gz | psql -U myerp myerp_v2
```

---

## Troubleshooting

### Backend Not Starting

**Check logs:**
```bash
sudo journalctl -u myerp-backend -n 100 -f
```

**Common issues:**
- Database connection failed → Verify `DATABASE_URL` in `.env`
- Redis connection failed → Check Redis is running: `redis-cli ping`
- Port already in use → `sudo lsof -i :8080`

### Database Connection Issues

```bash
# Test database connection
psql -U myerp -d myerp_v2 -h localhost

# Check PostgreSQL status
sudo systemctl status postgresql

# View PostgreSQL logs
sudo journalctl -u postgresql -n 100
```

### Frontend Build Errors

```bash
# Clear cache and rebuild
rm -rf .next
rm -rf node_modules
npm install
npm run build
```

### Migration Failures

```bash
# Check current version
migrate -path migrations -database "$DATABASE_URL" version

# Rollback last migration
migrate -path migrations -database "$DATABASE_URL" down 1

# Force version (use with caution)
migrate -path migrations -database "$DATABASE_URL" force <version>
```

---

## Scaling Considerations

### Horizontal Scaling

1. **Load Balancer**: Use Nginx or HAProxy
2. **Shared Session Store**: Redis for session management
3. **Database Read Replicas**: PostgreSQL streaming replication
4. **CDN**: CloudFlare or AWS CloudFront for static assets

### Database Optimization

```sql
-- Add indexes for common queries
CREATE INDEX CONCURRENTLY idx_users_email ON users(tenant_id, email);
CREATE INDEX CONCURRENTLY idx_sessions_user ON sessions(tenant_id, user_id, expires_at);
CREATE INDEX CONCURRENTLY idx_audit_created ON audit_logs(tenant_id, created_at DESC);

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM users WHERE tenant_id = 'xxx' AND email = 'xxx';
```

---

**For additional support, please contact the development team or create an issue on GitHub.**

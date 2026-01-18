# MyERP v2 API Documentation

**Version:** 1.0.0
**Base URL:** `http://localhost:8080/api`
**Authentication:** JWT Bearer Token

---

## Table of Contents

1. [Authentication](#authentication)
2. [Users](#users)
3. [Roles](#roles)
4. [Permissions](#permissions)
5. [Two-Factor Authentication](#two-factor-authentication)
6. [Sessions](#sessions)
7. [Invitations](#invitations)
8. [Audit Logs](#audit-logs)
9. [Security](#security)
10. [Error Responses](#error-responses)

---

## Authentication

### POST /auth/register
Register a new tenant and create the owner account.

**Request Body:**
```json
{
  "company_name": "My Company",
  "slug": "my-company",
  "email": "owner@mycompany.com",
  "first_name": "John",
  "last_name": "Doe",
  "password": "SecurePassword123!"
}
```

**Response (201 Created):**
```json
{
  "status": "success",
  "message": "Tenant registered successfully. Please check your email to verify your account.",
  "data": {
    "tenant_id": "uuid",
    "user_id": "uuid"
  }
}
```

---

### POST /auth/verify-email
Verify email address using the token sent via email.

**Request Body:**
```json
{
  "token": "verification-token-from-email"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Email verified successfully"
}
```

---

### POST /auth/login
Authenticate user and receive access/refresh tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "password123",
  "tenant_slug": "my-company",
  "remember_me": false
}
```

**Response (200 OK):**

**Without 2FA:**
```json
{
  "status": "success",
  "data": {
    "access_token": "jwt.access.token",
    "refresh_token": "jwt.refresh.token",
    "expires_in": 900,
    "user": {
      "id": "uuid",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "two_factor_enabled": false,
      "roles": [...]
    },
    "tenant": {
      "id": "uuid",
      "slug": "my-company",
      "company_name": "My Company"
    }
  }
}
```

**With 2FA Enabled:**
```json
{
  "status": "success",
  "data": {
    "requires_2fa": true,
    "two_factor_token": "temporary-2fa-token"
  }
}
```

---

### POST /auth/verify-2fa
Verify 2FA code after login.

**Request Body:**
```json
{
  "two_factor_token": "token-from-login",
  "code": "123456",
  "trust_device": false
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "access_token": "jwt.access.token",
    "refresh_token": "jwt.refresh.token",
    "expires_in": 900,
    "device_token": "device-trust-token",
    "user": {...},
    "tenant": {...}
  }
}
```

---

### POST /auth/logout
Logout current session (revoke token).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Logged out successfully"
}
```

---

### POST /auth/logout-all
Logout from all sessions (revoke all user tokens).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Logged out from all devices successfully"
}
```

---

### POST /auth/refresh
Refresh access token using refresh token.

**Request Body:**
```json
{
  "refresh_token": "jwt.refresh.token"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "access_token": "new.jwt.access.token",
    "expires_in": 900
  }
}
```

---

### POST /auth/forgot-password
Request password reset email.

**Request Body:**
```json
{
  "email": "user@example.com",
  "tenant_slug": "my-company"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "If an account exists with this email, a password reset link has been sent."
}
```

---

### POST /auth/reset-password
Reset password using token from email.

**Request Body:**
```json
{
  "token": "reset-token-from-email",
  "new_password": "NewSecurePassword123!"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Password reset successfully"
}
```

---

### POST /auth/change-password
Change password (requires authentication).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "current_password": "OldPassword123",
  "new_password": "NewPassword123!"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Password changed successfully"
}
```

---

### GET /auth/me
Get current authenticated user information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "two_factor_enabled": true,
    "status": "active",
    "roles": [
      {
        "id": "uuid",
        "name": "admin",
        "display_name": "Administrator"
      }
    ],
    "permissions": [
      {
        "resource": "users",
        "action": "view"
      }
    ]
  }
}
```

---

## Users

### GET /users
List users with pagination.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10, max: 100)

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "users": [
      {
        "id": "uuid",
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "status": "active",
        "two_factor_enabled": true,
        "last_login_at": "2026-01-17T10:30:00Z",
        "created_at": "2026-01-01T00:00:00Z",
        "roles": [...]
      }
    ]
  },
  "meta": {
    "current_page": 1,
    "page_size": 10,
    "total_count": 25,
    "total_pages": 3
  }
}
```

---

### GET /users/:id
Get user by ID.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "status": "active",
    "two_factor_enabled": true,
    "roles": [...],
    "permissions": [...]
  }
}
```

---

### POST /users
Create user (send invitation).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "email": "newuser@example.com",
  "role_ids": ["role-uuid-1", "role-uuid-2"],
  "message": "Welcome to the team!"
}
```

**Response (201 Created):**
```json
{
  "status": "success",
  "message": "Invitation sent successfully",
  "data": {
    "invitation_id": "uuid",
    "email": "newuser@example.com",
    "expires_at": "2026-01-24T10:30:00Z"
  }
}
```

---

### PUT /users/:id
Update user information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+1234567890"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "Jane",
    "last_name": "Smith",
    ...
  }
}
```

---

### DELETE /users/:id
Delete user (soft delete).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "User deleted successfully"
}
```

---

### PUT /users/:id/status
Update user status (activate/suspend).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "status": "suspended"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "User status updated successfully"
}
```

---

### GET /users/:id/roles
Get user's roles.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "roles": [
      {
        "id": "uuid",
        "name": "admin",
        "display_name": "Administrator",
        "is_system": true
      }
    ]
  }
}
```

---

### PUT /users/:id/roles
Assign roles to user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "role_ids": ["role-uuid-1", "role-uuid-2"]
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Roles assigned successfully"
}
```

---

### GET /users/search
Search users by name or email.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `query` (required): Search query
- `page` (optional): Page number
- `page_size` (optional): Items per page

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "users": [...],
    "query": "john"
  },
  "meta": {...}
}
```

---

## Roles

### GET /roles
List all roles.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `include_user_count` (optional): Include user count (default: false)

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "roles": [
      {
        "id": "uuid",
        "name": "admin",
        "display_name": "Administrator",
        "description": "Full system access",
        "is_system": true,
        "user_count": 5,
        "permissions": [...]
      }
    ]
  }
}
```

---

### POST /roles
Create custom role.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "name": "project-manager",
  "display_name": "Project Manager",
  "description": "Manages projects and team members",
  "permission_ids": ["perm-uuid-1", "perm-uuid-2"]
}
```

**Response (201 Created):**
```json
{
  "status": "success",
  "data": {
    "id": "uuid",
    "name": "project-manager",
    ...
  }
}
```

---

### PUT /roles/:id
Update role details.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "display_name": "Senior Project Manager",
  "description": "Updated description"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {...}
}
```

---

### DELETE /roles/:id
Delete role (cannot delete system roles).

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Role deleted successfully"
}
```

---

### PUT /roles/:id/permissions
Update role permissions.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "permission_ids": ["perm-uuid-1", "perm-uuid-2"]
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Permissions updated successfully"
}
```

---

## Permissions

### GET /permissions
List all available permissions.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "permissions": [
      {
        "id": "uuid",
        "resource": "users",
        "action": "view",
        "display_name": "View Users",
        "description": "View user list and details",
        "category": "User Management"
      }
    ]
  }
}
```

---

### GET /permissions/by-category
List permissions grouped by category.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "User Management": [
      {
        "id": "uuid",
        "resource": "users",
        "action": "view",
        "display_name": "View Users"
      }
    ],
    "Access Control": [...]
  }
}
```

---

## Two-Factor Authentication

### POST /2fa/setup
Generate QR code and backup codes for 2FA setup.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "qr_code": "base64-encoded-png-image",
    "secret": "JBSWY3DPEHPK3PXP",
    "backup_codes": [
      "1234-5678",
      "2345-6789",
      ...
    ]
  }
}
```

---

### POST /2fa/enable
Enable 2FA after verifying code.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "code": "123456"
}
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Two-factor authentication enabled successfully"
}
```

---

### POST /2fa/disable
Disable 2FA.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Two-factor authentication disabled"
}
```

---

## Sessions

### GET /sessions
List active sessions.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "sessions": [
      {
        "id": "uuid",
        "device_type": "desktop",
        "browser": "Chrome",
        "os": "macOS",
        "ip_address": "192.168.1.1",
        "city": "New York",
        "country_code": "US",
        "last_activity_at": "2026-01-17T10:30:00Z",
        "created_at": "2026-01-15T09:00:00Z"
      }
    ]
  }
}
```

---

### DELETE /sessions/:id
Revoke specific session.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "Session revoked successfully"
}
```

---

### DELETE /sessions/all
Revoke all sessions except current.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "status": "success",
  "message": "All sessions revoked successfully"
}
```

---

## Audit Logs

### GET /audit
Query audit logs with filters.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `action` (optional): Filter by action type
- `status` (optional): Filter by status (success/failure)
- `resource_type` (optional): Filter by resource type
- `start_date` (optional): Start date (ISO 8601)
- `end_date` (optional): End date (ISO 8601)
- `page` (optional): Page number
- `page_size` (optional): Items per page

**Response (200 OK):**
```json
{
  "status": "success",
  "data": {
    "logs": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "action": "user.login",
        "resource_type": "user",
        "resource_id": "uuid",
        "ip_address": "192.168.1.1",
        "status": "success",
        "metadata": {},
        "created_at": "2026-01-17T10:30:00Z"
      }
    ]
  },
  "meta": {...}
}
```

---

## Error Responses

All error responses follow this format:

**400 Bad Request:**
```json
{
  "status": "error",
  "message": "Validation failed",
  "errors": {
    "email": "Invalid email format",
    "password": "Password too short"
  }
}
```

**401 Unauthorized:**
```json
{
  "status": "error",
  "message": "Unauthorized: Invalid or expired token"
}
```

**403 Forbidden:**
```json
{
  "status": "error",
  "message": "Forbidden: Insufficient permissions"
}
```

**404 Not Found:**
```json
{
  "status": "error",
  "message": "Resource not found"
}
```

**500 Internal Server Error:**
```json
{
  "status": "error",
  "message": "Internal server error occurred"
}
```

---

## Rate Limiting

- **Login attempts:** 5 requests per 5 minutes per IP
- **2FA verification:** 5 requests per 15 minutes per user
- **Password reset:** 3 requests per hour per email
- **General API:** 100 requests per minute per user

Rate limit headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642435200
```

---

## Pagination

Paginated responses include metadata:

```json
{
  "meta": {
    "current_page": 1,
    "page_size": 10,
    "total_count": 45,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

---

## Authentication Flow

1. **Register**: POST `/auth/register` → Receive verification email
2. **Verify Email**: POST `/auth/verify-email` with token
3. **Login**: POST `/auth/login` → Receive tokens (or 2FA challenge)
4. **2FA (if enabled)**: POST `/auth/verify-2fa` → Receive tokens
5. **Access Protected Routes**: Include `Authorization: Bearer <access_token>` header
6. **Refresh Token**: POST `/auth/refresh` when access token expires
7. **Logout**: POST `/auth/logout`

---

**For additional support or questions, please refer to the GitHub repository or contact the development team.**

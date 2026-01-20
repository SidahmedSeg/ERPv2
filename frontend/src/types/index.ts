// User types
export interface User {
  id: string;
  tenant_id: string;
  email: string;
  email_verified: boolean;
  first_name: string;
  last_name: string;
  phone?: string;
  avatar_url?: string;
  status: 'active' | 'suspended' | 'deactivated' | 'pending';
  two_factor_enabled: boolean;
  timezone: string;
  language: string;
  preferences: Record<string, any>;
  last_login_at?: string;
  last_login_ip?: string;
  last_active_at?: string;
  created_at: string;
  updated_at: string;
  roles?: Role[];
  permissions?: Permission[];
}

// Tenant types
export interface Tenant {
  id: string;
  slug: string;
  company_name: string;
  status: 'pending_verification' | 'active' | 'suspended' | 'canceled';
  email: string;
  email_verified: boolean;
  plan_tier: 'free' | 'starter' | 'professional' | 'enterprise';
  settings: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// Role types
export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  parent_role_id?: string;
  level: number;
  is_system: boolean;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
  user_count?: number;
}

// Permission types
export interface Permission {
  id: string;
  resource: string;
  action: string;
  display_name: string;
  description?: string;
  category?: string;
  created_at: string;
}

export interface PermissionGroup {
  category: string;
  permissions: Permission[];
}

// Session types
export interface Session {
  id: string;
  tenant_id: string;
  user_id: string;
  device_type?: string;
  browser?: string;
  os?: string;
  ip_address?: string;
  user_agent?: string;
  country_code?: string;
  city?: string;
  last_activity_at: string;
  expires_at: string;
  created_at: string;
  is_current?: boolean;
}

// Invitation types
export interface Invitation {
  id: string;
  tenant_id: string;
  email: string;
  token: string;
  role_ids: string[];
  status: 'pending' | 'accepted' | 'expired' | 'revoked';
  message?: string;
  invited_by: string;
  invited_at: string;
  accepted_at?: string;
  expires_at: string;
}

// Audit log types
export interface AuditLog {
  id: string;
  tenant_id: string;
  user_id?: string;
  action: string;
  resource_type?: string;
  resource_id?: string;
  status: 'success' | 'failure';
  ip_address?: string;
  user_agent?: string;
  metadata: Record<string, any>;
  created_at: string;
}

// API Response types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
  meta?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  page_size: number;
  total_count: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
  tenant_id?: string; // Optional: For multi-tenant selection
  remember_me?: boolean;
}

export interface LoginResponse {
  user?: User;
  tenant?: Tenant;
  tenants?: Tenant[]; // For multi-tenant selection
  access_token?: string;
  refresh_token?: string;
  requires_2fa?: boolean;
  two_factor_token?: string;
}

export interface RegisterRequest {
  company_name: string;
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
}

export interface Verify2FARequest {
  two_factor_token: string;
  code: string;
  trust_device?: boolean;
}

// Form validation types
export interface ValidationErrors {
  [key: string]: string;
}

// 2FA types
export interface TwoFactorSetup {
  secret: string;
  qr_code_url: string;
  qr_code_image: string;
  backup_codes: string[];
}

// Security types
export interface SecurityOverview {
  session_stats: {
    active_sessions: number;
    device_breakdown: Array<{
      device_type: string;
      count: number;
    }>;
    last_activity_at?: string;
    last_ip_address?: string;
  };
  failed_attempts: number;
  recent_activity: AuditLog[];
  action_stats: Record<string, number>;
  last_24h_failures: AuditLog[];
}

export interface SecurityRecommendation {
  type: string;
  severity: 'info' | 'low' | 'medium' | 'high' | 'critical';
  title: string;
  description: string;
  action?: string;
  count?: number;
}

// Dashboard stats types
export interface DashboardStats {
  total_users: number;
  active_users: number;
  total_roles: number;
  pending_invitations: number;
}

// Company Settings types
export interface CompanySettings {
  id: string;
  tenant_id: string;
  company_name: string;
  legal_business_name?: string;
  industry?: string;
  specialty?: string;
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

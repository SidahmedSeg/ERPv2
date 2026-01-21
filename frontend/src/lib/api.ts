import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios';
import type {
  ApiResponse,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  User,
  Role,
  Permission,
  Session,
  Invitation,
  AuditLog,
  TwoFactorSetup,
  SecurityOverview,
  PermissionGroup,
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:18080/api';

// Create axios instance
const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - add auth token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    if (typeof window !== 'undefined') {
      const authData = localStorage.getItem('auth-storage');
      if (authData) {
        try {
          const { state } = JSON.parse(authData);
          if (state?.accessToken) {
            config.headers.Authorization = `Bearer ${state.accessToken}`;
          }
        } catch (error) {
          console.error('Failed to parse auth data:', error);
        }
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - handle errors and refresh tokens
api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    // Handle 401 Unauthorized - try to refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      if (typeof window !== 'undefined') {
        const authData = localStorage.getItem('auth-storage');
        if (authData) {
          try {
            const { state } = JSON.parse(authData);
            if (state?.refreshToken) {
              // Try to refresh the token
              const response = await axios.post(
                `${API_BASE_URL}/auth/refresh`,
                { refresh_token: state.refreshToken }
              );

              const { access_token, refresh_token } = response.data.data;

              // Update stored tokens
              const updatedState = {
                ...state,
                accessToken: access_token,
                refreshToken: refresh_token,
              };
              localStorage.setItem('auth-storage', JSON.stringify({ state: updatedState }));

              // Retry original request with new token
              originalRequest.headers.Authorization = `Bearer ${access_token}`;
              return api(originalRequest);
            }
          } catch (refreshError) {
            // Refresh failed - logout user
            localStorage.removeItem('auth-storage');
            if (typeof window !== 'undefined') {
              window.location.href = '/auth/login';
            }
          }
        }
      }
    }

    return Promise.reject(error);
  }
);

// Authentication API
export const authApi = {
  register: (data: RegisterRequest) =>
    api.post<ApiResponse<{ tenant: any; message: string }>>('/auth/register', data),

  login: (data: LoginRequest) =>
    api.post<ApiResponse<LoginResponse>>('/auth/login', data),

  verify2FA: (twoFactorToken: string, code: string, trustDevice?: boolean) =>
    api.post<ApiResponse<LoginResponse>>('/auth/verify-2fa', {
      two_factor_token: twoFactorToken,
      code,
      trust_device: trustDevice,
    }),

  verifyEmail: (token: string) =>
    api.post<ApiResponse>('/auth/verify-email', { token }),

  logout: () => api.post<ApiResponse>('/auth/logout'),

  logoutAll: () => api.post<ApiResponse>('/auth/logout-all'),

  refreshToken: (refreshToken: string) =>
    api.post<ApiResponse<{ access_token: string; refresh_token: string }>>(
      '/auth/refresh',
      { refresh_token: refreshToken }
    ),

  forgotPassword: (email: string, tenantSlug: string) =>
    api.post<ApiResponse>('/auth/forgot-password', { email, tenant_slug: tenantSlug }),

  resetPassword: (token: string, newPassword: string) =>
    api.post<ApiResponse>('/auth/reset-password', { token, new_password: newPassword }),

  changePassword: (currentPassword: string, newPassword: string) =>
    api.post<ApiResponse>('/auth/change-password', {
      current_password: currentPassword,
      new_password: newPassword,
    }),

  getMe: () => api.get<ApiResponse<{ user: User }>>('/auth/me'),
};

// User API
export const userApi = {
  list: (page = 1, pageSize = 20) =>
    api.get<ApiResponse<{ users: User[] }>>(`/users?page=${page}&page_size=${pageSize}`),

  get: (id: string) => api.get<ApiResponse<{ user: User }>>(`/users/${id}`),

  create: (data: Partial<User> & { password: string; role_ids: string[] }) =>
    api.post<ApiResponse<{ user: User }>>('/users', data),

  update: (id: string, data: Partial<User>) =>
    api.put<ApiResponse<{ user: User }>>(`/users/${id}`, data),

  delete: (id: string) => api.delete<ApiResponse>(`/users/${id}`),

  updateStatus: (id: string, status: string) =>
    api.patch<ApiResponse>(`/users/${id}/status`, { status }),

  getRoles: (id: string) =>
    api.get<ApiResponse<{ roles: Role[] }>>(`/users/${id}/roles`),

  assignRoles: (id: string, roleIds: string[]) =>
    api.post<ApiResponse<{ roles: Role[] }>>(`/users/${id}/roles`, { role_ids: roleIds }),

  search: (query: string) =>
    api.get<ApiResponse<{ users: User[] }>>(`/users/search?q=${encodeURIComponent(query)}`),
};

// Role API
export const roleApi = {
  list: (includeDetails = true) =>
    api.get<ApiResponse<{ roles: Role[] }>>(`/roles?include_details=${includeDetails}`),

  get: (id: string) => api.get<ApiResponse<{ role: Role }>>(`/roles/${id}`),

  create: (data: { name: string; display_name: string; description?: string; permission_ids: string[] }) =>
    api.post<ApiResponse<{ role: Role }>>('/roles', data),

  update: (id: string, data: { display_name?: string; description?: string; permission_ids?: string[] }) =>
    api.put<ApiResponse<{ role: Role }>>(`/roles/${id}`, data),

  delete: (id: string) => api.delete<ApiResponse>(`/roles/${id}`),

  getPermissions: (id: string) =>
    api.get<ApiResponse<{ permissions: Permission[] }>>(`/roles/${id}/permissions`),

  getUsers: (id: string) =>
    api.get<ApiResponse<{ users: User[] }>>(`/roles/${id}/users`),

  assignToUsers: (id: string, userIds: string[]) =>
    api.post<ApiResponse>(`/roles/${id}/assign`, { user_ids: userIds }),
};

// Permission API
export const permissionApi = {
  list: () => api.get<ApiResponse<{ permissions: Permission[] }>>('/permissions'),

  listByCategory: () =>
    api.get<ApiResponse<{ groups: PermissionGroup[] }>>('/permissions/by-category'),

  search: (query: string) =>
    api.get<ApiResponse<{ permissions: Permission[] }>>(`/permissions/search?q=${encodeURIComponent(query)}`),

  getStats: () => api.get<ApiResponse>('/permissions/stats'),

  getMyPermissions: () =>
    api.get<ApiResponse<{ permissions: Permission[] }>>('/permissions/me'),

  checkPermission: (resource: string, action: string) =>
    api.post<ApiResponse<{ has_permission: boolean }>>('/permissions/check', { resource, action }),
};

// 2FA API
export const twoFactorApi = {
  setup: () => api.post<ApiResponse<TwoFactorSetup>>('/2fa/setup'),

  enable: (secret: string, verificationCode: string, backupCodes: string[]) =>
    api.post<ApiResponse>('/2fa/enable', {
      secret,
      verification_code: verificationCode,
      backup_codes: backupCodes,
    }),

  disable: (password: string) => api.post<ApiResponse>('/2fa/disable', { password }),

  verify: (code: string) => api.post<ApiResponse>('/2fa/verify', { code }),

  verifyBackup: (code: string) =>
    api.post<ApiResponse<{ remaining_backup_codes: number }>>('/2fa/verify-backup', { code }),

  regenerateBackupCodes: (verificationCode: string) =>
    api.post<ApiResponse<{ backup_codes: string[] }>>('/2fa/backup-codes/regenerate', {
      verification_code: verificationCode,
    }),

  getBackupCodesCount: () =>
    api.get<ApiResponse<{ count: number }>>('/2fa/backup-codes/count'),

  trustDevice: (deviceFingerprint: string) =>
    api.post<ApiResponse<{ device_token: string }>>('/2fa/device/trust', {
      device_fingerprint: deviceFingerprint,
    }),
};

// Session API
export const sessionApi = {
  list: () => api.get<ApiResponse<{ sessions: Session[] }>>('/sessions'),

  getStats: () => api.get<ApiResponse>('/sessions/stats'),

  getRecentLogins: (limit = 10) =>
    api.get<ApiResponse<{ logins: Session[] }>>(`/sessions/recent-logins?limit=${limit}`),

  revoke: (id: string) => api.delete<ApiResponse>(`/sessions/${id}`),

  revokeAll: () => api.post<ApiResponse<{ revoked_count: number }>>('/sessions/revoke-all'),
};

// Invitation API
export const invitationApi = {
  list: (status = '', page = 1, pageSize = 20) =>
    api.get<ApiResponse<{ invitations: Invitation[] }>>(
      `/invitations?status=${status}&page=${page}&page_size=${pageSize}`
    ),

  get: (id: string) => api.get<ApiResponse<{ invitation: Invitation }>>(`/invitations/${id}`),

  create: (data: { email: string; role_ids: string[]; message?: string }) =>
    api.post<ApiResponse<{ invitation: Invitation }>>('/invitations', data),

  accept: (token: string, password: string, firstName: string, lastName: string) =>
    api.post<ApiResponse<{ user: User }>>('/invitations/accept', {
      token,
      password,
      first_name: firstName,
      last_name: lastName,
    }),

  revoke: (id: string) => api.delete<ApiResponse>(`/invitations/${id}`),

  resend: (id: string) => api.post<ApiResponse>(`/invitations/${id}/resend`),
};

// Audit Log API
export const auditApi = {
  list: (filters: any, page = 1, pageSize = 20) =>
    api.get<ApiResponse<{ logs: AuditLog[] }>>('/audit-logs', {
      params: { ...filters, page, page_size: pageSize },
    }),

  search: (query: string, page = 1, pageSize = 20) =>
    api.get<ApiResponse<{ logs: AuditLog[] }>>(
      `/audit-logs/search?q=${encodeURIComponent(query)}&page=${page}&page_size=${pageSize}`
    ),

  getStats: (startDate?: string, endDate?: string) =>
    api.get<ApiResponse>('/audit-logs/stats', {
      params: { start_date: startDate, end_date: endDate },
    }),

  getFailedAttempts: (userId?: string, since?: string) =>
    api.get<ApiResponse<{ logs: AuditLog[] }>>('/audit-logs/failed-attempts', {
      params: { user_id: userId, since },
    }),

  getUserActivity: (userId: string, limit = 50) =>
    api.get<ApiResponse<{ logs: AuditLog[] }>>(`/audit-logs/user/${userId}?limit=${limit}`),

  getResourceActivity: (resourceType: string, resourceId: string, limit = 50) =>
    api.get<ApiResponse<{ logs: AuditLog[] }>>(
      `/audit-logs/resource/${resourceType}/${resourceId}?limit=${limit}`
    ),
};

// Security API
export const securityApi = {
  getOverview: () => api.get<ApiResponse<SecurityOverview>>('/security/overview'),

  getSuspiciousActivity: (since?: string) =>
    api.get<ApiResponse<{ activities: any[] }>>('/security/suspicious-activity', {
      params: { since },
    }),

  getRecommendations: () =>
    api.get<ApiResponse<{ recommendations: any[] }>>('/security/recommendations'),

  getLoginHistory: (limit = 50) =>
    api.get<ApiResponse<{ logins: Session[] }>>(`/security/login-history?limit=${limit}`),
};

export { api };
export default api;

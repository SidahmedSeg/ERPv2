import { api } from './api';

// ============================================
// TYPES
// ============================================

export interface Permission {
  id: string;
  resource: string;
  action: string;
  display_name: string;
  description?: string;
  category?: string;
  created_at: string;
}

export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  display_name: string;
  description?: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
  created_by?: string;
}

export interface RoleWithPermissions extends Role {
  permissions: Permission[];
}

export interface Department {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  head_user_id?: string;
  color: string;
  icon: string;
  status: string;
  created_at: string;
  updated_at: string;
  created_by?: string;
}

export interface DepartmentWithDetails extends Department {
  head_user_name?: string;
  member_count: number;
}

export interface TeamMember {
  id: string;
  email: string;
  name: string;
  role: string; // owner, admin, manager, user
  job_title?: string;
  phone?: string;
  avatar_url?: string;
  department_id?: string;
  department_name?: string;
  status: string;
  last_active_at?: string;
  created_at: string;
}

export interface TeamMemberWithRoles extends TeamMember {
  roles: string[];
}

export interface Invitation {
  id: string;
  tenant_id: string;
  email: string;
  token: string;
  role_ids: string[];
  department_id?: string;
  invited_by?: string;
  welcome_message?: string;
  status: string;
  expires_at: string;
  accepted_at?: string;
  accepted_by?: string;
  created_at: string;
  updated_at: string;
}

export interface InvitationWithDetails extends Invitation {
  invited_by_name?: string;
  department_name?: string;
  is_expired: boolean;
}

// Request types
export interface CreateRoleRequest {
  name: string;
  display_name: string;
  description?: string;
  permission_ids: string[];
}

export interface UpdateRoleRequest {
  display_name?: string;
  description?: string;
  permission_ids?: string[];
}

export interface CreateDepartmentRequest {
  name: string;
  description?: string;
  head_user_id?: string;
  color: string;
  icon: string;
}

export interface UpdateDepartmentRequest {
  name?: string;
  description?: string;
  head_user_id?: string;
  color?: string;
  icon?: string;
  status?: string;
}

export interface InviteMemberRequest {
  email: string;
  role_ids: string[];
  department_id?: string;
  welcome_message?: string;
}

export interface UpdateMemberRequest {
  name?: string;
  job_title?: string;
  phone?: string;
  department_id?: string;
}

export interface UpdateMemberStatusRequest {
  status: string;
}

export interface AssignRolesRequest {
  role_ids: string[];
}

// ============================================
// PERMISSIONS API
// ============================================

export const permissionsApi = {
  async listAll(grouped?: boolean): Promise<Permission[] | Record<string, Permission[]>> {
    const url = grouped ? '/api/team/permissions?grouped=true' : '/api/team/permissions';
    const response = await api.get(url);
    return response.data;
  },

  async search(query: string): Promise<Permission[]> {
    const response = await api.get(`/api/team/permissions/search?q=${encodeURIComponent(query)}`);
    return response.data;
  },
};

// ============================================
// ROLES API
// ============================================

export const rolesApi = {
  async list(): Promise<Role[]> {
    const response = await api.get('/api/team/roles');
    return response.data;
  },

  async get(id: string): Promise<RoleWithPermissions> {
    const response = await api.get(`/api/team/roles/${id}`);
    return response.data;
  },

  async create(data: CreateRoleRequest): Promise<RoleWithPermissions> {
    const response = await api.post('/api/team/roles', data);
    return response.data;
  },

  async update(id: string, data: UpdateRoleRequest): Promise<RoleWithPermissions> {
    const response = await api.put(`/api/team/roles/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/team/roles/${id}`);
  },

  async getUserRoles(userId: string): Promise<Role[]> {
    const response = await api.get(`/api/team/members/${userId}/roles`);
    return response.data;
  },

  async assignRoles(userId: string, data: AssignRolesRequest): Promise<void> {
    await api.post(`/api/team/members/${userId}/roles`, data);
  },
};

// ============================================
// DEPARTMENTS API
// ============================================

export const departmentsApi = {
  async list(): Promise<DepartmentWithDetails[]> {
    const response = await api.get('/api/team/departments');
    return response.data;
  },

  async get(id: string): Promise<DepartmentWithDetails> {
    const response = await api.get(`/api/team/departments/${id}`);
    return response.data;
  },

  async create(data: CreateDepartmentRequest): Promise<DepartmentWithDetails> {
    const response = await api.post('/api/team/departments', data);
    return response.data;
  },

  async update(id: string, data: UpdateDepartmentRequest): Promise<DepartmentWithDetails> {
    const response = await api.put(`/api/team/departments/${id}`, data);
    return response.data;
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/team/departments/${id}`);
  },

  async getMembers(id: string): Promise<TeamMember[]> {
    const response = await api.get(`/api/team/departments/${id}/members`);
    return response.data;
  },
};

// ============================================
// TEAM MEMBERS API
// ============================================

export const teamMembersApi = {
  async list(params?: {
    status?: string;
    department_id?: string;
    with_roles?: boolean;
  }): Promise<TeamMember[] | TeamMemberWithRoles[]> {
    const searchParams = new URLSearchParams();
    if (params?.status) searchParams.append('status', params.status);
    if (params?.department_id) searchParams.append('department_id', params.department_id);
    if (params?.with_roles) searchParams.append('with_roles', 'true');

    const url = `/api/team/members${searchParams.toString() ? '?' + searchParams.toString() : ''}`;
    const response = await api.get(url);
    return response.data;
  },

  async get(id: string, withRoles?: boolean): Promise<TeamMember | TeamMemberWithRoles> {
    const url = `/api/team/members/${id}${withRoles ? '?with_roles=true' : ''}`;
    const response = await api.get(url);
    return response.data;
  },

  async update(id: string, data: UpdateMemberRequest): Promise<TeamMember> {
    const response = await api.put(`/api/team/members/${id}`, data);
    return response.data;
  },

  async updateStatus(id: string, data: UpdateMemberStatusRequest): Promise<void> {
    await api.patch(`/api/team/members/${id}/status`, data);
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/team/members/${id}`);
  },

  async updateAvatar(id: string, avatarUrl: string): Promise<void> {
    await api.put(`/api/team/members/${id}/avatar`, { avatar_url: avatarUrl });
  },

  async getStats(): Promise<Record<string, number>> {
    const response = await api.get('/api/team/members/stats');
    return response.data;
  },
};

// ============================================
// INVITATIONS API
// ============================================

export const invitationsApi = {
  async list(status?: string): Promise<InvitationWithDetails[]> {
    const url = status ? `/api/team/invitations?status=${status}` : '/api/team/invitations';
    const response = await api.get(url);
    return response.data;
  },

  async get(id: string): Promise<Invitation> {
    const response = await api.get(`/api/team/invitations/${id}`);
    return response.data;
  },

  async create(data: InviteMemberRequest): Promise<Invitation> {
    const response = await api.post('/api/team/invitations', data);
    return response.data;
  },

  async resend(id: string): Promise<void> {
    await api.post(`/api/team/invitations/${id}/resend`);
  },

  async revoke(id: string): Promise<void> {
    await api.post(`/api/team/invitations/${id}/revoke`);
  },

  async delete(id: string): Promise<void> {
    await api.delete(`/api/team/invitations/${id}`);
  },
};

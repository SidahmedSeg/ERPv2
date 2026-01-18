import { api } from '../api';

export interface ProfileResponse {
    id: string;
    email: string;
    first_name: string;
    last_name: string;
    phone?: string;
    job_title?: string;
    avatar_url?: string;
    department_id?: string;
    role: string;
    status: string;
    is_active: boolean;
    timezone: string;
    language: string;
    created_at: string;
    last_login_at?: string;
}

export interface UpdateProfileRequest {
    first_name: string;
    last_name: string;
    phone?: string;
    job_title?: string;
}

export interface ChangePasswordRequest {
    current_password: string;
    new_password: string;
    confirm_password: string;
}

export interface UpdatePreferencesRequest {
    timezone: string;
    language: string;
}

export const userProfileApi = {
    // Get current user profile
    getProfile: async (token: string): Promise<ProfileResponse> => {
        const response = await api.get('/api/users/me/profile', {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Update profile information
    updateProfile: async (data: UpdateProfileRequest, token: string): Promise<ProfileResponse> => {
        const response = await api.put('/api/users/me/profile', data, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Change password
    changePassword: async (data: ChangePasswordRequest, token: string): Promise<{ message: string }> => {
        const response = await api.put('/api/users/me/password', data, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Update preferences
    updatePreferences: async (data: UpdatePreferencesRequest, token: string): Promise<{ message: string; timezone: string; language: string }> => {
        const response = await api.put('/api/users/me/preferences', data, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Upload avatar
    uploadAvatar: async (file: File, token: string): Promise<{ message: string; avatar_url: string }> => {
        const formData = new FormData();
        formData.append('avatar', file);

        const response = await api.post('/api/users/me/avatar', formData, {
            headers: {
                Authorization: `Bearer ${token}`,
                'Content-Type': 'multipart/form-data',
            },
        });
        return response.data;
    },

    // Delete avatar
    deleteAvatar: async (token: string): Promise<{ message: string }> => {
        const response = await api.delete('/api/users/me/avatar', {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },
};

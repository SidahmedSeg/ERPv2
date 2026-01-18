import { api } from '../api';

export interface TwoFactorStatus {
    user_id: string;
    enabled: boolean;
    enabled_at?: string;
}

export interface TwoFactorSetupResponse {
    secret: string;
    qr_code: string; // base64 encoded PNG
    backup_codes: string[];
}

export interface EnableTwoFactorRequest {
    secret: string;
    code: string;
    backup_codes: string[];
}

export interface VerifyTOTPRequest {
    code: string;
}

export interface UserSession {
    id: string;
    user_id: string;
    device_info: Record<string, any>;
    ip_address: string;
    user_agent: string;
    last_activity_at: string;
    created_at: string;
    expires_at: string;
    is_current: boolean;
}

export const twoFactorApi = {
    // Get 2FA status for current user
    getStatus: async (token: string): Promise<TwoFactorStatus> => {
        const response = await api.get('/api/2fa/status', {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data.data;
    },

    // Setup 2FA (generate secret and QR code)
    setup: async (token: string): Promise<TwoFactorSetupResponse> => {
        const response = await api.post('/api/2fa/setup', {}, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data.data;
    },

    // Enable 2FA after verifying the initial code
    enable: async (data: EnableTwoFactorRequest, token: string): Promise<void> => {
        const response = await api.post('/api/2fa/enable', data, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Disable 2FA
    disable: async (password: string, token: string): Promise<void> => {
        const response = await api.post('/api/2fa/disable', { password }, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Verify TOTP code (used during login)
    verify: async (data: VerifyTOTPRequest, token: string): Promise<{ valid: boolean }> => {
        const response = await api.post('/api/2fa/verify', data, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },
};

export const sessionApi = {
    // Get all active sessions for current user
    getSessions: async (token: string): Promise<UserSession[]> => {
        const response = await api.get('/api/sessions', {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data.data.sessions;
    },

    // Revoke a specific session
    revokeSession: async (sessionId: string, token: string): Promise<void> => {
        const response = await api.delete(`/api/sessions/${sessionId}`, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },

    // Revoke all other sessions (keep current one)
    revokeAllOtherSessions: async (token: string): Promise<void> => {
        const response = await api.post('/api/sessions/revoke-all', {}, {
            headers: { Authorization: `Bearer ${token}` },
        });
        return response.data;
    },
};

import { api } from '@/lib/api';
import type { ApiResponse, CompanySettings } from '@/types';

export const companySettingsApi = {
  // Get company settings
  getSettings: () =>
    api.get<ApiResponse<CompanySettings>>('/settings/company'),

  // Update company settings (partial update)
  updateSettings: (data: Partial<CompanySettings>) =>
    api.put<ApiResponse<CompanySettings>>('/settings/company', data),
};
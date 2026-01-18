'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useAuthStore } from '@/store/auth-store';
import { authApi } from '@/lib/api';
import { handleApiError, validateEmail } from '@/lib/utils';

export default function LoginPage() {
  const router = useRouter();
  const setAuth = useAuthStore((state) => state.setAuth);

  const [formData, setFormData] = useState({
    email: '',
    password: '',
    tenantSlug: '',
    rememberMe: false,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(false);
  const [requires2FA, setRequires2FA] = useState(false);
  const [twoFactorToken, setTwoFactorToken] = useState('');
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [trustDevice, setTrustDevice] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === 'checkbox' ? checked : value,
    }));
    // Clear error when user types
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: '' }));
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!validateEmail(formData.email)) {
      newErrors.email = 'Invalid email format';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    }

    if (!formData.tenantSlug) {
      newErrors.tenantSlug = 'Organization slug is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) return;

    setLoading(true);
    setErrors({});

    try {
      const response = await authApi.login({
        email: formData.email,
        password: formData.password,
        tenant_slug: formData.tenantSlug,
        remember_me: formData.rememberMe,
      });

      const data = response.data.data;

      if (data?.requires_2fa) {
        // 2FA required
        setRequires2FA(true);
        setTwoFactorToken(data.two_factor_token || '');
      } else if (data) {
        // Login successful
        setAuth({
          user: data.user,
          tenant: data.tenant,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
        });
        router.push('/dashboard');
      }
    } catch (error) {
      const errorMessage = handleApiError(error);
      setErrors({ submit: errorMessage });
    } finally {
      setLoading(false);
    }
  };

  const handleVerify2FA = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!twoFactorCode) {
      setErrors({ twoFactorCode: 'Verification code is required' });
      return;
    }

    setLoading(true);
    setErrors({});

    try {
      const response = await authApi.verify2FA(twoFactorToken, twoFactorCode, trustDevice);
      const data = response.data.data;

      if (data) {
        setAuth({
          user: data.user,
          tenant: data.tenant,
          accessToken: data.access_token,
          refreshToken: data.refresh_token,
        });
        router.push('/dashboard');
      }
    } catch (error) {
      const errorMessage = handleApiError(error);
      setErrors({ twoFactorCode: errorMessage });
    } finally {
      setLoading(false);
    }
  };

  if (requires2FA) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div>
            <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
              Two-Factor Authentication
            </h2>
            <p className="mt-2 text-center text-sm text-gray-600">
              Enter the 6-digit code from your authenticator app
            </p>
          </div>

          <form className="mt-8 space-y-6" onSubmit={handleVerify2FA}>
            {errors.submit && (
              <div className="rounded-md bg-red-50 p-4">
                <p className="text-sm text-red-800">{errors.submit}</p>
              </div>
            )}

            <div>
              <label htmlFor="twoFactorCode" className="block text-sm font-medium text-gray-700">
                Verification Code
              </label>
              <input
                id="twoFactorCode"
                name="twoFactorCode"
                type="text"
                autoComplete="one-time-code"
                required
                maxLength={6}
                className={`mt-1 appearance-none relative block w-full px-3 py-2 border ${
                  errors.twoFactorCode ? 'border-red-300' : 'border-gray-300'
                } placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm`}
                placeholder="000000"
                value={twoFactorCode}
                onChange={(e) => {
                  setTwoFactorCode(e.target.value.replace(/\D/g, '').slice(0, 6));
                  if (errors.twoFactorCode) {
                    setErrors((prev) => ({ ...prev, twoFactorCode: '' }));
                  }
                }}
              />
              {errors.twoFactorCode && (
                <p className="mt-1 text-sm text-red-600">{errors.twoFactorCode}</p>
              )}
            </div>

            <div className="flex items-center">
              <input
                id="trustDevice"
                name="trustDevice"
                type="checkbox"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                checked={trustDevice}
                onChange={(e) => setTrustDevice(e.target.checked)}
              />
              <label htmlFor="trustDevice" className="ml-2 block text-sm text-gray-900">
                Trust this device for 30 days
              </label>
            </div>

            <div>
              <button
                type="submit"
                disabled={loading}
                className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading ? 'Verifying...' : 'Verify'}
              </button>
            </div>

            <div className="text-center">
              <button
                type="button"
                onClick={() => {
                  setRequires2FA(false);
                  setTwoFactorCode('');
                  setTwoFactorToken('');
                }}
                className="text-sm text-blue-600 hover:text-blue-500"
              >
                Back to login
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Sign in to MyERP v2
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600">
            Or{' '}
            <Link href="/auth/signup" className="font-medium text-blue-600 hover:text-blue-500">
              create a new account
            </Link>
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          {errors.submit && (
            <div className="rounded-md bg-red-50 p-4">
              <p className="text-sm text-red-800">{errors.submit}</p>
            </div>
          )}

          <div className="rounded-md shadow-sm space-y-4">
            <div>
              <label htmlFor="tenantSlug" className="block text-sm font-medium text-gray-700">
                Organization Slug
              </label>
              <input
                id="tenantSlug"
                name="tenantSlug"
                type="text"
                autoComplete="organization"
                required
                className={`mt-1 appearance-none relative block w-full px-3 py-2 border ${
                  errors.tenantSlug ? 'border-red-300' : 'border-gray-300'
                } placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm`}
                placeholder="your-company"
                value={formData.tenantSlug}
                onChange={handleChange}
              />
              {errors.tenantSlug && (
                <p className="mt-1 text-sm text-red-600">{errors.tenantSlug}</p>
              )}
            </div>

            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                Email address
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                className={`mt-1 appearance-none relative block w-full px-3 py-2 border ${
                  errors.email ? 'border-red-300' : 'border-gray-300'
                } placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm`}
                placeholder="you@example.com"
                value={formData.email}
                onChange={handleChange}
              />
              {errors.email && <p className="mt-1 text-sm text-red-600">{errors.email}</p>}
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                className={`mt-1 appearance-none relative block w-full px-3 py-2 border ${
                  errors.password ? 'border-red-300' : 'border-gray-300'
                } placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm`}
                placeholder="••••••••"
                value={formData.password}
                onChange={handleChange}
              />
              {errors.password && <p className="mt-1 text-sm text-red-600">{errors.password}</p>}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <input
                id="rememberMe"
                name="rememberMe"
                type="checkbox"
                className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                checked={formData.rememberMe}
                onChange={handleChange}
              />
              <label htmlFor="rememberMe" className="ml-2 block text-sm text-gray-900">
                Remember me
              </label>
            </div>

            <div className="text-sm">
              <Link
                href="/auth/forgot-password"
                className="font-medium text-blue-600 hover:text-blue-500"
              >
                Forgot your password?
              </Link>
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Signing in...' : 'Sign in'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

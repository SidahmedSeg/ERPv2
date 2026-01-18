'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { LoadingButton } from '@/components/ui/loading-button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Checkbox } from '@/components/ui/checkbox';
import { authApi } from '@/lib/api';
import { useAuthStore } from '@/store/auth-store';

interface Tenant {
    id: string;
    company_name: string;
    slug: string;
}

export default function LoginPage() {
    const router = useRouter();
    const setAuth = useAuthStore((state) => state.setAuth);
    const [loading, setLoading] = useState(false);
    const [isNavigating, setIsNavigating] = useState(false);
    const [error, setError] = useState('');

    const [formData, setFormData] = useState({
        email: '',
        password: '',
        remember: false,
        totpCode: '',
        tenant_id: '',
    });

    const [requires2FA, setRequires2FA] = useState(false);
    const [tempToken, setTempToken] = useState('');
    const [availableTenants, setAvailableTenants] = useState<Tenant[]>([]);
    const [showTenantSelector, setShowTenantSelector] = useState(false);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value,
        }));
    };

    const handleTenantSelect = async (tenantId: string) => {
        setError('');
        setLoading(true);
        setFormData(prev => ({ ...prev, tenant_id: tenantId }));

        try {
            const response = await authApi.login({
                email: formData.email,
                password: formData.password,
                remember_me: formData.remember,
                tenant_id: tenantId,
            });

            if (response.data.success && response.data.data) {
                const responseData = response.data.data;

                // Check if 2FA is required
                if (responseData.requires_2fa && responseData.two_factor_token) {
                    setRequires2FA(true);
                    setTempToken(responseData.two_factor_token);
                    setShowTenantSelector(false);
                    setLoading(false);
                    return;
                }

                // Ensure all required fields are present
                if (responseData.user && responseData.tenant && responseData.access_token && responseData.refresh_token) {
                    setAuth({
                        user: responseData.user,
                        tenant: responseData.tenant,
                        accessToken: responseData.access_token,
                        refreshToken: responseData.refresh_token,
                    });

                    setLoading(false);
                    setIsNavigating(true);
                    router.push('/dashboard');
                } else {
                    setError('Invalid response from server');
                    setLoading(false);
                }
            }
        } catch (err: any) {
            setError(err.response?.data?.error || 'Login failed. Please try again.');
            setLoading(false);
            setShowTenantSelector(false);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setLoading(true);

        try {
            const response = await authApi.login({
                email: formData.email,
                password: formData.password,
                remember_me: formData.remember,
            });

            if (response.data.success && response.data.data) {
                const responseData = response.data.data;

                // Check if user belongs to multiple tenants
                if (responseData.tenants && responseData.tenants.length > 1) {
                    setAvailableTenants(responseData.tenants);
                    setShowTenantSelector(true);
                    setLoading(false);
                    return;
                }

                // Check if 2FA is required
                if (responseData.requires_2fa && responseData.two_factor_token) {
                    setRequires2FA(true);
                    setTempToken(responseData.two_factor_token);
                    setLoading(false);
                    return;
                }

                // Ensure all required fields are present
                if (responseData.user && responseData.tenant && responseData.access_token && responseData.refresh_token) {
                    setAuth({
                        user: responseData.user,
                        tenant: responseData.tenant,
                        accessToken: responseData.access_token,
                        refreshToken: responseData.refresh_token,
                    });

                    setLoading(false);
                    setIsNavigating(true);
                    router.push('/dashboard');
                } else {
                    setError('Invalid response from server');
                    setLoading(false);
                }
            }
        } catch (err: any) {
            setError(err.response?.data?.error || 'Login failed. Please check your credentials.');
            setLoading(false);
            setIsNavigating(false);
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
            <Card className="w-full max-w-md">
                <CardHeader>
                    <CardTitle className="text-2xl">Welcome Back</CardTitle>
                    <CardDescription>
                        {showTenantSelector
                            ? 'Select your company to continue'
                            : requires2FA
                            ? 'Enter your two-factor authentication code'
                            : 'Sign in to your account'
                        }
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    {showTenantSelector ? (
                        <div className="space-y-3">
                            <p className="text-sm text-gray-600 mb-4">
                                Your account has access to multiple companies. Please select one:
                            </p>
                            {availableTenants.map((tenant) => (
                                <button
                                    key={tenant.id}
                                    onClick={() => handleTenantSelect(tenant.id)}
                                    disabled={loading}
                                    className="w-full p-4 border rounded-lg hover:border-indigo-500 hover:bg-indigo-50 transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    <div className="font-semibold text-gray-900">{tenant.company_name}</div>
                                    <div className="text-sm text-gray-500">app.infold.app</div>
                                </button>
                            ))}
                            {error && (
                                <Alert variant="destructive" className="mt-4">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}
                        </div>
                    ) : (
                        <form onSubmit={handleSubmit} className="space-y-4">
                            {error && (
                                <Alert variant="destructive">
                                    <AlertDescription>{error}</AlertDescription>
                                </Alert>
                            )}

                            {!requires2FA && (
                                <>
                                    <div className="space-y-2">
                                        <Label htmlFor="email">Email Address *</Label>
                                        <Input
                                            id="email"
                                            name="email"
                                            type="email"
                                            placeholder="john@company.com"
                                            value={formData.email}
                                            onChange={handleChange}
                                            required
                                            autoComplete="email"
                                        />
                                    </div>

                                    <div className="space-y-2">
                                        <Label htmlFor="password">Password *</Label>
                                        <Input
                                            id="password"
                                            name="password"
                                            type="password"
                                            placeholder="••••••••"
                                            value={formData.password}
                                            onChange={handleChange}
                                            required
                                            autoComplete="current-password"
                                        />
                                    </div>

                                    <div className="flex items-center space-x-2">
                                        <Checkbox
                                            id="remember"
                                            name="remember"
                                            checked={formData.remember}
                                            onCheckedChange={(checked) =>
                                                setFormData(prev => ({ ...prev, remember: checked as boolean }))
                                            }
                                        />
                                        <Label htmlFor="remember" className="text-sm font-normal cursor-pointer">
                                            Remember me for 30 days
                                        </Label>
                                    </div>
                                </>
                            )}

                            {requires2FA && (
                                <div className="space-y-2">
                                    <Label htmlFor="totpCode">Two-Factor Authentication Code *</Label>
                                    <Input
                                        id="totpCode"
                                        name="totpCode"
                                        type="text"
                                        inputMode="numeric"
                                        maxLength={6}
                                        placeholder="000000"
                                        value={formData.totpCode}
                                        onChange={handleChange}
                                        required
                                        className="text-center text-2xl tracking-widest font-mono"
                                        autoFocus
                                    />
                                    <p className="text-sm text-gray-600">
                                        Enter the 6-digit code from your authenticator app
                                    </p>
                                </div>
                            )}

                            <LoadingButton
                                type="submit"
                                className="w-full"
                                loading={loading || isNavigating}
                                loadingText={loading ? 'Authenticating...' : 'Redirecting to dashboard...'}
                            >
                                Sign In
                            </LoadingButton>

                            <div className="text-center space-y-2">
                                <p className="text-sm text-gray-600">
                                    Don't have an account?{' '}
                                    <Link
                                        href="/auth/signup"
                                        className="text-indigo-600 hover:underline font-semibold"
                                    >
                                        Sign up
                                    </Link>
                                </p>
                                <Link href="/auth/forgot-password" className="text-sm text-indigo-600 hover:underline block">
                                    Forgot password?
                                </Link>
                            </div>
                        </form>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}

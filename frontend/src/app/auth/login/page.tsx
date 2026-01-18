'use client';

import { useState, useEffect } from 'react';
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

export default function LoginPage() {
    const router = useRouter();
    const setAuth = useAuthStore((state) => state.setAuth);
    const [loading, setLoading] = useState(false);
    const [isNavigating, setIsNavigating] = useState(false);
    const [error, setError] = useState('');
    const [currentSubdomain, setCurrentSubdomain] = useState<string | null>(null);

    const [formData, setFormData] = useState({
        slug: '',
        email: '',
        password: '',
        remember: false,
        totpCode: '',
    });
    const [requires2FA, setRequires2FA] = useState(false);
    const [tempToken, setTempToken] = useState('');

    useEffect(() => {
        // Extract subdomain from current hostname
        const hostname = window.location.hostname;
        const subdomain = extractSubdomain(hostname);

        if (subdomain) {
            setCurrentSubdomain(subdomain);
            setFormData(prev => ({ ...prev, slug: subdomain }));
        }
    }, []);

    const extractSubdomain = (hostname: string): string | null => {
        try {
            if (!hostname) return null;

            const parts = hostname.split('.');
            const baseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || 'myerp.local';

            // For localhost, return null
            if (hostname.includes('localhost')) {
                return null;
            }

            if (parts.length >= 3) {
                const domain = parts.slice(-2).join('.');
                if (domain === baseDomain) {
                    return parts[0];
                }
            }
            return null;
        } catch (error) {
            console.error('Error extracting subdomain:', error);
            return null;
        }
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: type === 'checkbox' ? checked : value,
        }));
    };

    const handleSubdomainRedirect = () => {
        if (formData.slug && !currentSubdomain) {
            // Redirect to subdomain
            const baseDomain = process.env.NEXT_PUBLIC_BASE_DOMAIN || 'myerp.local';
            const port = window.location.port ? `:${window.location.port}` : '';
            window.location.href = `http://${formData.slug}.${baseDomain}${port}/auth/login`;
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        // If no subdomain in URL, redirect to subdomain first
        if (!currentSubdomain && formData.slug) {
            handleSubdomainRedirect();
            return;
        }

        setLoading(true);

        try {
            const response = await authApi.login({
                email: formData.email,
                password: formData.password,
                tenant_slug: formData.slug,
                remember_me: formData.remember,
            });

            if (response.data.success) {
                const responseData = response.data.data;

                // Check if 2FA is required
                if (responseData.requires_2fa) {
                    setRequires2FA(true);
                    setTempToken(responseData.two_factor_token);
                    setLoading(false);
                    return;
                }

                setAuth({
                    user: responseData.user,
                    tenant: responseData.tenant,
                    accessToken: responseData.access_token,
                    refreshToken: responseData.refresh_token,
                });

                // Show navigating state
                setLoading(false);
                setIsNavigating(true);

                // Redirect to dashboard
                router.push('/dashboard');
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
                        {currentSubdomain ? (
                            <>Sign in to <span className="font-semibold text-primary">{currentSubdomain}.myerp.com</span></>
                        ) : (
                            'Enter your subdomain and credentials to sign in'
                        )}
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <Alert variant="destructive">
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}

                        {/* Show subdomain field only if not on subdomain */}
                        {!currentSubdomain && (
                            <div className="space-y-2">
                                <Label htmlFor="slug">Company Subdomain *</Label>
                                <div className="flex items-center">
                                    <Input
                                        id="slug"
                                        name="slug"
                                        placeholder="acme"
                                        value={formData.slug}
                                        onChange={handleChange}
                                        required
                                        className="rounded-r-none"
                                    />
                                    <span className="bg-gray-100 border border-l-0 px-3 py-2 rounded-r-md text-sm text-gray-600">
                    .myerp.com
                  </span>
                                </div>
                            </div>
                        )}

                        {/* Show email/password only if on subdomain */}
                        {currentSubdomain && !requires2FA && (
                            <>
                                <div className="space-y-2">
                                    <Label htmlFor="email">Email Address *</Label>
                                    <Input
                                        id="email"
                                        name="email"
                                        type="email"
                                        placeholder="john@acme.com"
                                        value={formData.email}
                                        onChange={handleChange}
                                        required
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

                        {/* Show 2FA code input when 2FA is required */}
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
                            {currentSubdomain ? 'Sign In' : 'Continue'}
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
                            {currentSubdomain && (
                                <Link href="/auth/forgot-password" className="text-sm text-indigo-600 hover:underline block">
                                    Forgot password?
                                </Link>
                            )}
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

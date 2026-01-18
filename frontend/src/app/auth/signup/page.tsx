'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { authApi } from '@/lib/api';

export default function SignupPage() {
    const router = useRouter();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const [formData, setFormData] = useState({
        company_name: '',
        email: '',
        password: '',
        confirm_password: '',
        first_name: '',
        last_name: '',
    });

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        setFormData(prev => ({ ...prev, [name]: value }));
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        // Validation
        if (formData.password !== formData.confirm_password) {
            setError('Passwords do not match');
            return;
        }

        if (formData.password.length < 8) {
            setError('Password must be at least 8 characters');
            return;
        }

        setLoading(true);

        try {
            const response = await authApi.register({
                company_name: formData.company_name,
                email: formData.email,
                password: formData.password,
                first_name: formData.first_name,
                last_name: formData.last_name,
            });

            if (response.data.success) {
                router.push('/auth/verify?status=sent');
            }
        } catch (err: any) {
            const errorMessage = err.response?.data?.error?.message
                || err.response?.data?.error
                || err.response?.data?.message
                || err.message
                || 'Registration failed. Please try again.';
            setError(typeof errorMessage === 'string' ? errorMessage : 'Registration failed. Please try again.');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
            <Card className="w-full max-w-2xl">
                <CardHeader>
                    <CardTitle className="text-2xl">Create Your Account</CardTitle>
                    <CardDescription>
                        Get started with MyERP in minutes. Secure, isolated workspace for your company.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <Alert variant="destructive">
                                <AlertDescription>{error}</AlertDescription>
                            </Alert>
                        )}

                        {/* Company Name */}
                        <div className="space-y-2">
                            <Label htmlFor="company_name">Company Name *</Label>
                            <Input
                                id="company_name"
                                name="company_name"
                                placeholder="Acme Corporation"
                                value={formData.company_name}
                                onChange={handleChange}
                                required
                            />
                        </div>

                        {/* Name Fields */}
                        <div className="grid md:grid-cols-2 gap-4">
                            <div className="space-y-2">
                                <Label htmlFor="first_name">First Name *</Label>
                                <Input
                                    id="first_name"
                                    name="first_name"
                                    placeholder="John"
                                    value={formData.first_name}
                                    onChange={handleChange}
                                    required
                                />
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="last_name">Last Name *</Label>
                                <Input
                                    id="last_name"
                                    name="last_name"
                                    placeholder="Doe"
                                    value={formData.last_name}
                                    onChange={handleChange}
                                    required
                                />
                            </div>
                        </div>

                        {/* Email */}
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

                        {/* Password Fields */}
                        <div className="grid md:grid-cols-2 gap-4">
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
                                    minLength={8}
                                />
                                <p className="text-xs text-gray-500">Minimum 8 characters</p>
                            </div>

                            <div className="space-y-2">
                                <Label htmlFor="confirm_password">Confirm Password *</Label>
                                <Input
                                    id="confirm_password"
                                    name="confirm_password"
                                    type="password"
                                    placeholder="••••••••"
                                    value={formData.confirm_password}
                                    onChange={handleChange}
                                    required
                                    minLength={8}
                                />
                            </div>
                        </div>

                        <Button type="submit" className="w-full" disabled={loading}>
                            {loading ? 'Creating Account...' : 'Create Account'}
                        </Button>

                        <p className="text-center text-sm text-gray-600">
                            Already have an account?{' '}
                            <Link href="/auth/login" className="text-indigo-600 hover:underline font-semibold">
                                Sign in
                            </Link>
                        </p>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}

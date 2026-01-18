'use client';

import { useState } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { userProfileApi } from '@/lib/api/user-profile';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Key, Loader2, Eye, EyeOff } from 'lucide-react';
import { toast } from 'sonner';

export function PasswordSection() {
    const { accessToken } = useAuthStore();
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [processing, setProcessing] = useState(false);
    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);

    const [formData, setFormData] = useState({
        current_password: '',
        new_password: '',
        confirm_password: '',
    });

    const handleChange = (field: string) => (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData(prev => ({ ...prev, [field]: e.target.value }));
    };

    const validateForm = (): boolean => {
        if (!formData.current_password) {
            toast.error('Please enter your current password');
            return false;
        }

        if (!formData.new_password) {
            toast.error('Please enter a new password');
            return false;
        }

        if (formData.new_password.length < 8) {
            toast.error('Password must be at least 8 characters long');
            return false;
        }

        if (formData.new_password === formData.current_password) {
            toast.error('New password must be different from current password');
            return false;
        }

        if (formData.new_password !== formData.confirm_password) {
            toast.error('Passwords do not match');
            return false;
        }

        return true;
    };

    const handleSubmit = async () => {
        if (!accessToken || !validateForm()) return;

        try {
            setProcessing(true);
            await userProfileApi.changePassword({
                current_password: formData.current_password,
                new_password: formData.new_password,
                confirm_password: formData.confirm_password,
            }, accessToken);
            toast.success('Password changed successfully');
            setIsDialogOpen(false);
            setFormData({
                current_password: '',
                new_password: '',
                confirm_password: '',
            });
        } catch (error: any) {
            console.error('Error changing password:', error);
            toast.error(error?.response?.data?.error || 'Failed to change password');
        } finally {
            setProcessing(false);
        }
    };

    const handleDialogClose = (open: boolean) => {
        if (!processing) {
            setIsDialogOpen(open);
            if (!open) {
                setFormData({
                    current_password: '',
                    new_password: '',
                    confirm_password: '',
                });
                setShowCurrentPassword(false);
                setShowNewPassword(false);
                setShowConfirmPassword(false);
            }
        }
    };

    return (
        <>
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Key className="h-5 w-5" />
                        Password
                    </CardTitle>
                    <CardDescription>
                        Change your password to keep your account secure
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <Button onClick={() => setIsDialogOpen(true)}>
                        Change Password
                    </Button>
                </CardContent>
            </Card>

            <Dialog open={isDialogOpen} onOpenChange={handleDialogClose}>
                <DialogContent className="max-w-md">
                    <DialogHeader>
                        <DialogTitle>Change Password</DialogTitle>
                        <DialogDescription>
                            Enter your current password and choose a new one. Password must be at least 8 characters long.
                        </DialogDescription>
                    </DialogHeader>

                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="current-password">Current Password</Label>
                            <div className="relative">
                                <Input
                                    id="current-password"
                                    type={showCurrentPassword ? 'text' : 'password'}
                                    placeholder="Enter your current password"
                                    value={formData.current_password}
                                    onChange={handleChange('current_password')}
                                    disabled={processing}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    disabled={processing}
                                >
                                    {showCurrentPassword ? (
                                        <EyeOff className="h-4 w-4" />
                                    ) : (
                                        <Eye className="h-4 w-4" />
                                    )}
                                </button>
                            </div>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="new-password">New Password</Label>
                            <div className="relative">
                                <Input
                                    id="new-password"
                                    type={showNewPassword ? 'text' : 'password'}
                                    placeholder="Enter your new password"
                                    value={formData.new_password}
                                    onChange={handleChange('new_password')}
                                    disabled={processing}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowNewPassword(!showNewPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    disabled={processing}
                                >
                                    {showNewPassword ? (
                                        <EyeOff className="h-4 w-4" />
                                    ) : (
                                        <Eye className="h-4 w-4" />
                                    )}
                                </button>
                            </div>
                            <p className="text-xs text-muted-foreground">
                                Must be at least 8 characters long
                            </p>
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="confirm-password">Confirm New Password</Label>
                            <div className="relative">
                                <Input
                                    id="confirm-password"
                                    type={showConfirmPassword ? 'text' : 'password'}
                                    placeholder="Confirm your new password"
                                    value={formData.confirm_password}
                                    onChange={handleChange('confirm_password')}
                                    disabled={processing}
                                    className="pr-10"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                                    disabled={processing}
                                >
                                    {showConfirmPassword ? (
                                        <EyeOff className="h-4 w-4" />
                                    ) : (
                                        <Eye className="h-4 w-4" />
                                    )}
                                </button>
                            </div>
                        </div>
                    </div>

                    <DialogFooter>
                        <Button
                            variant="outline"
                            onClick={() => handleDialogClose(false)}
                            disabled={processing}
                        >
                            Cancel
                        </Button>
                        <Button
                            onClick={handleSubmit}
                            disabled={processing}
                        >
                            {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Change Password
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}

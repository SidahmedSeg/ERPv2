'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { twoFactorApi, TwoFactorStatus, TwoFactorSetupResponse } from '@/lib/api/two-factor';
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
import {
    Alert,
    AlertDescription,
} from '@/components/ui/alert';
import { Shield, Loader2, Check, Copy, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';

export function SecuritySection() {
    const { accessToken } = useAuthStore();
    const [status, setStatus] = useState<TwoFactorStatus | null>(null);
    const [loading, setLoading] = useState(true);
    const [isSetupDialogOpen, setIsSetupDialogOpen] = useState(false);
    const [isDisableDialogOpen, setIsDisableDialogOpen] = useState(false);
    const [setupData, setSetupData] = useState<TwoFactorSetupResponse | null>(null);
    const [verificationCode, setVerificationCode] = useState('');
    const [processing, setProcessing] = useState(false);
    const [setupStep, setSetupStep] = useState<'qr' | 'verify' | 'backup'>('qr');
    const [copiedCodes, setCopiedCodes] = useState(false);
    const [disablePassword, setDisablePassword] = useState('');

    useEffect(() => {
        loadStatus();
    }, []);

    const loadStatus = async () => {
        if (!accessToken) return;
        try {
            setLoading(true);
            const data = await twoFactorApi.getStatus(accessToken);
            setStatus(data);
        } catch (error: any) {
            console.error('Error loading 2FA status:', error);
            toast.error(error?.response?.data?.error || 'Failed to load 2FA status');
        } finally {
            setLoading(false);
        }
    };

    const handleSetupStart = async () => {
        if (!accessToken) return;
        try {
            setProcessing(true);
            const data = await twoFactorApi.setup(accessToken);
            setSetupData(data);
            setSetupStep('qr');
            setIsSetupDialogOpen(true);
        } catch (error: any) {
            console.error('Error starting 2FA setup:', error);
            toast.error(error?.response?.data?.error || 'Failed to start 2FA setup');
        } finally {
            setProcessing(false);
        }
    };

    const handleVerifyAndEnable = async () => {
        if (!accessToken || !setupData) return;

        if (!verificationCode || verificationCode.length !== 6) {
            toast.error('Please enter a valid 6-digit code');
            return;
        }

        try {
            setProcessing(true);
            await twoFactorApi.enable({
                secret: setupData.secret,
                code: verificationCode,
                backup_codes: setupData.backup_codes,
            }, accessToken);
            toast.success('Two-factor authentication enabled successfully');
            setSetupStep('backup');
        } catch (error: any) {
            console.error('Error enabling 2FA:', error);
            toast.error(error?.response?.data?.error || 'Invalid verification code');
        } finally {
            setProcessing(false);
        }
    };

    const handleFinishSetup = () => {
        setIsSetupDialogOpen(false);
        setSetupData(null);
        setVerificationCode('');
        setSetupStep('qr');
        setCopiedCodes(false);
        loadStatus();
    };

    const handleDisable = async () => {
        if (!accessToken) return;

        if (!disablePassword) {
            toast.error('Please enter your password');
            return;
        }

        try {
            setProcessing(true);
            await twoFactorApi.disable(disablePassword, accessToken);
            toast.success('Two-factor authentication disabled');
            setIsDisableDialogOpen(false);
            setDisablePassword('');
            loadStatus();
        } catch (error: any) {
            console.error('Error disabling 2FA:', error);
            toast.error(error?.response?.data?.error || 'Failed to disable 2FA');
        } finally {
            setProcessing(false);
        }
    };

    const copyBackupCodes = async () => {
        if (!setupData) return;
        try {
            const codesText = setupData.backup_codes.join('\n');
            await navigator.clipboard.writeText(codesText);
            setCopiedCodes(true);
            toast.success('Backup codes copied to clipboard');
            setTimeout(() => setCopiedCodes(false), 3000);
        } catch (err) {
            console.error('Failed to copy backup codes:', err);
            toast.error('Failed to copy to clipboard');
        }
    };

    if (loading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Shield className="h-5 w-5" />
                        Two-Factor Authentication
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex items-center justify-center py-8">
                        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                    </div>
                </CardContent>
            </Card>
        );
    }

    return (
        <>
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Shield className="h-5 w-5" />
                        Two-Factor Authentication
                    </CardTitle>
                    <CardDescription>
                        Add an extra layer of security to your account
                    </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <p className="font-medium">
                                Status: {status?.enabled ? (
                                    <span className="text-green-600">Enabled</span>
                                ) : (
                                    <span className="text-muted-foreground">Disabled</span>
                                )}
                            </p>
                            {status?.enabled && status.enabled_at && (
                                <p className="text-sm text-muted-foreground">
                                    Enabled on {new Date(status.enabled_at).toLocaleDateString()}
                                </p>
                            )}
                        </div>
                        {status?.enabled ? (
                            <Button
                                variant="default"
                                onClick={() => setIsDisableDialogOpen(true)}
                            >
                                Disable 2FA
                            </Button>
                        ) : (
                            <Button
                                onClick={handleSetupStart}
                                disabled={processing}
                            >
                                {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                Enable 2FA
                            </Button>
                        )}
                    </div>

                    {!status?.enabled && (
                        <Alert>
                            <AlertCircle className="h-4 w-4" />
                            <AlertDescription>
                                Two-factor authentication adds an extra layer of security to your account by requiring a code from your authenticator app when signing in.
                            </AlertDescription>
                        </Alert>
                    )}
                </CardContent>
            </Card>

            {/* Setup Dialog */}
            <Dialog open={isSetupDialogOpen} onOpenChange={setIsSetupDialogOpen}>
                <DialogContent className="max-w-md">
                    <DialogHeader>
                        <DialogTitle>
                            {setupStep === 'qr' && 'Scan QR Code'}
                            {setupStep === 'verify' && 'Verify Code'}
                            {setupStep === 'backup' && 'Save Backup Codes'}
                        </DialogTitle>
                        <DialogDescription>
                            {setupStep === 'qr' && 'Scan this QR code with your authenticator app'}
                            {setupStep === 'verify' && 'Enter the 6-digit code from your authenticator app'}
                            {setupStep === 'backup' && 'Save these backup codes in a safe place'}
                        </DialogDescription>
                    </DialogHeader>

                    {setupStep === 'qr' && setupData && (
                        <div className="space-y-4">
                            <div className="flex justify-center bg-white p-4 rounded-lg">
                                <img
                                    src={`data:image/png;base64,${setupData.qr_code}`}
                                    alt="2FA QR Code"
                                    width={256}
                                    height={256}
                                    className="rounded"
                                />
                            </div>
                            <div className="space-y-2">
                                <Label>Or enter this code manually:</Label>
                                <div className="flex items-center gap-2">
                                    <Input
                                        value={setupData.secret}
                                        readOnly
                                        className="font-mono text-sm"
                                    />
                                    <Button
                                        variant="outline"
                                        size="icon"
                                        onClick={async () => {
                                            try {
                                                await navigator.clipboard.writeText(setupData.secret);
                                                toast.success('Secret copied to clipboard');
                                            } catch (err) {
                                                console.error('Failed to copy secret:', err);
                                                toast.error('Failed to copy to clipboard');
                                            }
                                        }}
                                    >
                                        <Copy className="h-4 w-4" />
                                    </Button>
                                </div>
                            </div>
                        </div>
                    )}

                    {setupStep === 'verify' && (
                        <div className="space-y-4">
                            <div className="space-y-2">
                                <Label htmlFor="verification-code">Verification Code</Label>
                                <Input
                                    id="verification-code"
                                    type="text"
                                    inputMode="numeric"
                                    pattern="[0-9]*"
                                    maxLength={6}
                                    placeholder="000000"
                                    value={verificationCode}
                                    onChange={(e) => setVerificationCode(e.target.value.replace(/\D/g, ''))}
                                    className="text-center text-2xl tracking-widest font-mono"
                                />
                            </div>
                        </div>
                    )}

                    {setupStep === 'backup' && setupData && (
                        <div className="space-y-4">
                            <Alert>
                                <AlertCircle className="h-4 w-4" />
                                <AlertDescription>
                                    Store these codes securely. Each code can only be used once to access your account if you lose access to your authenticator app.
                                </AlertDescription>
                            </Alert>
                            <div className="bg-muted p-4 rounded-lg space-y-2">
                                {setupData.backup_codes.map((code, index) => (
                                    <div key={index} className="font-mono text-sm">
                                        {code}
                                    </div>
                                ))}
                            </div>
                            <Button
                                variant="outline"
                                className="w-full"
                                onClick={copyBackupCodes}
                            >
                                {copiedCodes ? (
                                    <>
                                        <Check className="mr-2 h-4 w-4" />
                                        Copied!
                                    </>
                                ) : (
                                    <>
                                        <Copy className="mr-2 h-4 w-4" />
                                        Copy All Codes
                                    </>
                                )}
                            </Button>
                        </div>
                    )}

                    <DialogFooter>
                        {setupStep === 'qr' && (
                            <>
                                <Button variant="outline" onClick={() => setIsSetupDialogOpen(false)}>
                                    Cancel
                                </Button>
                                <Button onClick={() => setSetupStep('verify')}>
                                    Next
                                </Button>
                            </>
                        )}
                        {setupStep === 'verify' && (
                            <>
                                <Button variant="outline" onClick={() => setSetupStep('qr')}>
                                    Back
                                </Button>
                                <Button onClick={handleVerifyAndEnable} disabled={processing}>
                                    {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                                    Verify & Enable
                                </Button>
                            </>
                        )}
                        {setupStep === 'backup' && (
                            <Button onClick={handleFinishSetup} className="w-full">
                                Done
                            </Button>
                        )}
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Disable Dialog */}
            <Dialog open={isDisableDialogOpen} onOpenChange={(open) => {
                setIsDisableDialogOpen(open);
                if (!open) setDisablePassword('');
            }}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Disable Two-Factor Authentication</DialogTitle>
                        <DialogDescription>
                            Please enter your password to confirm disabling 2FA. This will make your account less secure.
                        </DialogDescription>
                    </DialogHeader>
                    <div className="space-y-4 py-4">
                        <div className="space-y-2">
                            <Label htmlFor="disable-password">Password</Label>
                            <Input
                                id="disable-password"
                                type="password"
                                placeholder="Enter your password"
                                value={disablePassword}
                                onChange={(e) => setDisablePassword(e.target.value)}
                            />
                        </div>
                    </div>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setIsDisableDialogOpen(false)}>
                            Cancel
                        </Button>
                        <Button
                            variant="default"
                            onClick={handleDisable}
                            disabled={processing}
                        >
                            {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Disable 2FA
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}

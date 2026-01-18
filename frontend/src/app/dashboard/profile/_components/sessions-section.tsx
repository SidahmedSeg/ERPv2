'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { sessionApi, UserSession } from '@/lib/api/two-factor';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Laptop, Smartphone, Tablet, Monitor, Loader2, LogOut, AlertCircle } from 'lucide-react';
import { toast } from 'sonner';
import { format } from 'date-fns';

export function SessionsSection() {
    const { token } = useAuthStore();
    const [sessions, setSessions] = useState<UserSession[]>([]);
    const [loading, setLoading] = useState(true);
    const [processing, setProcessing] = useState(false);
    const [revokeDialogOpen, setRevokeDialogOpen] = useState(false);
    const [revokeAllDialogOpen, setRevokeAllDialogOpen] = useState(false);
    const [sessionToRevoke, setSessionToRevoke] = useState<string | null>(null);

    useEffect(() => {
        loadSessions();
    }, []);

    const loadSessions = async () => {
        if (!token) return;
        try {
            setLoading(true);
            const data = await sessionApi.getSessions(token);
            setSessions(data);
        } catch (error: any) {
            console.error('Error loading sessions:', error);
            toast.error(error?.response?.data?.error || 'Failed to load sessions');
            setSessions([]);
        } finally {
            setLoading(false);
        }
    };

    const handleRevokeSession = async (sessionId: string) => {
        if (!token) return;
        try {
            setProcessing(true);
            await sessionApi.revokeSession(sessionId, token);
            toast.success('Session revoked successfully');
            setRevokeDialogOpen(false);
            setSessionToRevoke(null);
            loadSessions();
        } catch (error: any) {
            console.error('Error revoking session:', error);
            toast.error(error?.response?.data?.error || 'Failed to revoke session');
        } finally {
            setProcessing(false);
        }
    };

    const handleRevokeAllOtherSessions = async () => {
        if (!token) return;
        try {
            setProcessing(true);
            await sessionApi.revokeAllOtherSessions(token);
            toast.success('All other sessions revoked successfully');
            setRevokeAllDialogOpen(false);
            loadSessions();
        } catch (error: any) {
            console.error('Error revoking sessions:', error);
            toast.error(error?.response?.data?.error || 'Failed to revoke sessions');
        } finally {
            setProcessing(false);
        }
    };

    const getDeviceIcon = (userAgent: string) => {
        const ua = userAgent.toLowerCase();
        if (ua.includes('mobile') || ua.includes('android') || ua.includes('iphone')) {
            return <Smartphone className="h-5 w-5" />;
        }
        if (ua.includes('tablet') || ua.includes('ipad')) {
            return <Tablet className="h-5 w-5" />;
        }
        return <Monitor className="h-5 w-5" />;
    };

    const getDeviceName = (userAgent: string) => {
        const ua = userAgent.toLowerCase();
        if (ua.includes('chrome')) return 'Chrome';
        if (ua.includes('firefox')) return 'Firefox';
        if (ua.includes('safari')) return 'Safari';
        if (ua.includes('edge')) return 'Edge';
        return 'Unknown Browser';
    };

    const getDeviceType = (userAgent: string) => {
        const ua = userAgent.toLowerCase();
        if (ua.includes('windows')) return 'Windows';
        if (ua.includes('mac')) return 'macOS';
        if (ua.includes('linux')) return 'Linux';
        if (ua.includes('android')) return 'Android';
        if (ua.includes('iphone') || ua.includes('ipad')) return 'iOS';
        return 'Unknown OS';
    };

    if (loading) {
        return (
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Laptop className="h-5 w-5" />
                        Active Sessions
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

    // Ensure sessions is always an array before filtering
    const sessionsList = Array.isArray(sessions) ? sessions : [];
    const otherSessions = sessionsList.filter(s => !s.is_current);

    return (
        <>
            <Card>
                <CardHeader>
                    <div className="flex items-center justify-between">
                        <div>
                            <CardTitle className="flex items-center gap-2">
                                <Laptop className="h-5 w-5" />
                                Active Sessions
                            </CardTitle>
                            <CardDescription>
                                Manage devices where you're currently signed in
                            </CardDescription>
                        </div>
                        {otherSessions.length > 0 && (
                            <Button
                                variant="destructive"
                                size="sm"
                                onClick={() => setRevokeAllDialogOpen(true)}
                            >
                                <LogOut className="mr-2 h-4 w-4" />
                                Revoke All Others
                            </Button>
                        )}
                    </div>
                </CardHeader>
                <CardContent className="space-y-4">
                    {sessionsList.length === 0 ? (
                        <div className="text-center py-8 text-muted-foreground">
                            <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
                            <p>No active sessions found</p>
                        </div>
                    ) : (
                        <>
                            {sessionsList.map((session) => (
                                <div
                                    key={session.id}
                                    className="flex items-start justify-between p-4 border rounded-lg"
                                >
                                    <div className="flex items-start gap-3">
                                        <div className="mt-1">
                                            {getDeviceIcon(session.user_agent)}
                                        </div>
                                        <div className="space-y-1">
                                            <div className="flex items-center gap-2">
                                                <p className="font-medium">
                                                    {getDeviceName(session.user_agent)} on {getDeviceType(session.user_agent)}
                                                </p>
                                                {session.is_current && (
                                                    <span className="text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded-full font-medium">
                                                        Current
                                                    </span>
                                                )}
                                            </div>
                                            <p className="text-sm text-muted-foreground">
                                                IP: {session.ip_address}
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                Last active: {format(new Date(session.last_activity_at), 'PPp')}
                                            </p>
                                            <p className="text-xs text-muted-foreground">
                                                Created: {format(new Date(session.created_at), 'PPp')}
                                            </p>
                                        </div>
                                    </div>
                                    {!session.is_current && (
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={() => {
                                                setSessionToRevoke(session.id);
                                                setRevokeDialogOpen(true);
                                            }}
                                        >
                                            <LogOut className="mr-2 h-4 w-4" />
                                            Revoke
                                        </Button>
                                    )}
                                </div>
                            ))}
                        </>
                    )}
                </CardContent>
            </Card>

            {/* Revoke Single Session Dialog */}
            <Dialog open={revokeDialogOpen} onOpenChange={setRevokeDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Revoke Session</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to revoke this session? The device will be signed out immediately.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRevokeDialogOpen(false)}>
                            Cancel
                        </Button>
                        <Button
                            variant="destructive"
                            onClick={() => sessionToRevoke && handleRevokeSession(sessionToRevoke)}
                            disabled={processing}
                        >
                            {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Revoke Session
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>

            {/* Revoke All Other Sessions Dialog */}
            <Dialog open={revokeAllDialogOpen} onOpenChange={setRevokeAllDialogOpen}>
                <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Revoke All Other Sessions</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to revoke all other sessions? All other devices will be signed out immediately. Your current session will remain active.
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={() => setRevokeAllDialogOpen(false)}>
                            Cancel
                        </Button>
                        <Button
                            variant="destructive"
                            onClick={handleRevokeAllOtherSessions}
                            disabled={processing}
                        >
                            {processing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                            Revoke All Others
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </>
    );
}

'use client';

import { PasswordSection } from '../profile/_components/password-section';
import { SecuritySection } from '../profile/_components/security-section';
import { SessionsSection } from '../profile/_components/sessions-section';

// Force dynamic rendering for this page
export const dynamic = 'force-dynamic';

export default function SecurityPage() {
    return (
        <div className="p-6 space-y-6 max-w-5xl mx-auto">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Security Settings</h1>
                <p className="text-muted-foreground mt-1">
                    Manage your account security, password, and active sessions
                </p>
            </div>

            <div className="grid gap-6">
                <PasswordSection />
                <SecuritySection />
                <SessionsSection />
            </div>
        </div>
    );
}

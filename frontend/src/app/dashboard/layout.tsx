'use client';

import { useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { useAuthStore } from '@/store/auth-store';
import { Sidebar } from '@/components/layout/sidebar';
import { Header } from '@/components/layout/header';
import { NotificationProvider } from '@/contexts/NotificationContext';
import { PermissionProvider } from '@/contexts/PermissionContext';
import { ToastContainer } from '@/components/notifications/toast-container';

// Force dynamic rendering for all dashboard pages
export const dynamic = 'force-dynamic';

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const router = useRouter();
  const { user, isAuthenticated } = useAuthStore();

  useEffect(() => {
    // Redirect to login if not authenticated
    if (!isAuthenticated || !user) {
      router.push('/auth/login');
      return;
    }
  }, [isAuthenticated, user, router]);

  // Don't render until auth is checked
  if (!isAuthenticated || !user) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <NotificationProvider>
      <PermissionProvider>
        <div className="min-h-screen bg-background">
          {/* Sidebar - Desktop Only */}
          <div className="hidden md:block">
            <Sidebar />
          </div>

          {/* Main Content Area */}
          <div className="md:pl-64">
            {/* Header */}
            <Header />

            {/* Page Content */}
            <main className="bg-background">
              {children}
            </main>
          </div>
        </div>

        {/* Toast Notifications */}
        <ToastContainer />
      </PermissionProvider>
    </NotificationProvider>
  );
}

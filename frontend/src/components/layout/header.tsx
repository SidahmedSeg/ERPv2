'use client';

import React from 'react';
import { UserDropdown } from './user-dropdown';
import { Notifications } from './notifications';
import { MobileSidebar } from './mobile-sidebar';
import { QuickActions } from './quick-actions';
import { ThemeToggle } from '@/components/ui/theme-toggle';
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from '@/components/ui/breadcrumb';
import { useBreadcrumbs } from '@/hooks/use-breadcrumbs';
import Link from 'next/link';

export function Header() {
    const breadcrumbs = useBreadcrumbs();

    return (
        <header className="sticky top-0 z-40 bg-card">
            <div className="flex h-16 items-center justify-between gap-4 px-4 md:px-6 border-b border-border">
                {/* Left Section - Mobile Menu + Breadcrumb */}
                <div className="flex items-center gap-4">
                    <MobileSidebar />

                    {/* Breadcrumb - Hidden on mobile */}
                    <Breadcrumb className="hidden md:block">
                        <BreadcrumbList>
                            {breadcrumbs.map((crumb, index) => (
                                <React.Fragment key={index}>
                                    {index > 0 && <BreadcrumbSeparator />}
                                    <BreadcrumbItem>
                                        {crumb.href ? (
                                            <Link href={crumb.href}>
                                                <BreadcrumbLink>{crumb.label}</BreadcrumbLink>
                                            </Link>
                                        ) : (
                                            <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                                        )}
                                    </BreadcrumbItem>
                                </React.Fragment>
                            ))}
                        </BreadcrumbList>
                    </Breadcrumb>
                </div>

                {/* Right Section - Notifications + Quick Actions + Theme + User */}
                <div className="flex items-center gap-2 ml-auto">
                    {/* Notifications */}
                    <Notifications />

                    {/* Quick Actions */}
                    <QuickActions />

                    {/* Theme Toggle */}
                    <ThemeToggle />

                    {/* User Dropdown */}
                    <UserDropdown />
                </div>
            </div>
        </header>
    );
}
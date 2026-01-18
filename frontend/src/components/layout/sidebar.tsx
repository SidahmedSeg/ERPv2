'use client';

import { useState, useEffect } from 'react';
import { usePathname } from 'next/navigation';
import Link from 'next/link';
import { Settings, ChevronDown, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import { NAVIGATION_ITEMS } from '@/lib/constants';
import { CompanyDropdown } from './company-dropdown';
import { usePermissions } from '@/hooks/usePermissions';
import { useOptimisticNavigation } from '@/hooks/useOptimisticNavigation';
import { useAuthStore } from '@/store/auth-store';

export function Sidebar() {
    const pathname = usePathname();
    const [openMenus, setOpenMenus] = useState<string[]>([]); // All menus closed by default
    const { hasPermission, loading } = usePermissions();
    const { navigate, isPending, pendingPath } = useOptimisticNavigation();
    // Removed: unread email count (Stalwart inbox feature removed)

    const toggleMenu = (title: string) => {
        setOpenMenus((prev) =>
            prev.includes(title) ? prev.filter((t) => t !== title) : [...prev, title]
        );
    };

    const isMenuActive = (href: string, submenu?: any[]) => {
        if (submenu) {
            return submenu.some((sub) => pathname === sub.href || pathname.startsWith(sub.href + '/'));
        }
        return pathname === href;
    };

    // Filter navigation items based on permissions
    // If permissions are still loading, show items without permission requirements
    const filteredNavigationItems = NAVIGATION_ITEMS.filter((item) => {
        // If no permission required, always show
        if (!item.permission) return true;

        // If still loading, hide items that require permissions
        if (loading) return false;

        // If item has submenu, check if user has access to at least one submenu item
        if (item.submenu && item.submenu.length > 0) {
            const hasSubmenuAccess = item.submenu.some((subitem) => {
                if (!subitem.permission) return true;
                return hasPermission(subitem.permission.resource, subitem.permission.action);
            });
            return hasSubmenuAccess;
        }

        // For regular items, check permission
        return hasPermission(item.permission.resource, item.permission.action);
    }).map((item) => {
        // If item has submenu, filter submenu items based on permissions
        if (item.submenu && item.submenu.length > 0) {
            return {
                ...item,
                submenu: item.submenu.filter((subitem: any) => {
                    if (!subitem.permission) return true;
                    // If still loading, hide items that require permissions
                    if (loading) return false;
                    return hasPermission(subitem.permission.resource, subitem.permission.action);
                }),
            };
        }
        return item;
    });

    return (
        <aside className="fixed left-0 top-0 h-screen w-64 border-r border-border bg-card flex flex-col">
            {/* Company Section - Same height as header (h-16) */}
            <div className="h-16 flex items-center px-4 border-b border-border">
                <CompanyDropdown />
            </div>

            {/* Navigation */}
            <nav className="flex-1 overflow-y-auto py-4 px-3">
                <ul className="space-y-0.5">
                    {filteredNavigationItems.map((item) => {
                        const Icon = item.icon;
                        const hasSubmenu = item.submenu && item.submenu.length > 0;
                        const isOpen = openMenus.includes(item.title);
                        const isActive = isMenuActive(item.href, item.submenu);

                        return (
                            <li key={item.href}>
                                {hasSubmenu ? (
                                    <>
                                        <button
                                            onClick={() => toggleMenu(item.title)}
                                            className={cn(
                                                'flex items-center justify-between w-full gap-3 rounded-lg px-3 py-1.5 text-sm text-foreground transition-all',
                                                isActive
                                                    ? 'bg-accent font-semibold'
                                                    : 'font-normal hover:bg-accent/50 hover:font-medium'
                                            )}
                                        >
                                            <div className="flex items-center gap-3">
                                                <Icon className="h-4 w-4" />
                                                <span>{item.title}</span>
                                            </div>
                                            <ChevronDown
                                                className={cn(
                                                    'h-4 w-4 transition-transform',
                                                    isOpen ? 'rotate-180' : ''
                                                )}
                                            />
                                        </button>
                                        {isOpen && (
                                            <ul className="mt-1 ml-6 space-y-0.5">
                                                {item.submenu.map((subitem) => {
                                                    const SubIcon = subitem.icon;
                                                    const isSubActive = pathname === subitem.href;
                                                    const isPendingNav = isPending && pendingPath === subitem.href;

                                                    return (
                                                        <li key={subitem.href}>
                                                            <Link
                                                                href={subitem.href}
                                                                onClick={(e) => {
                                                                    e.preventDefault();
                                                                    navigate(subitem.href);
                                                                }}
                                                                prefetch={true}
                                                                className={cn(
                                                                    'flex items-center gap-2.5 rounded-md px-3 py-1.5 text-xs text-foreground transition-all w-full text-left',
                                                                    isSubActive || isPendingNav
                                                                        ? 'bg-accent font-semibold'
                                                                        : 'font-normal hover:bg-accent/50 hover:font-medium',
                                                                    isPendingNav && 'opacity-60'
                                                                )}
                                                            >
                                                                <SubIcon className={cn('h-3.5 w-3.5', isPendingNav && 'animate-pulse')} />
                                                                <span>{subitem.title}</span>
                                                                {isPendingNav && <Loader2 className="ml-auto h-3 w-3 animate-spin" />}
                                                            </Link>
                                                        </li>
                                                    );
                                                })}
                                            </ul>
                                        )}
                                    </>
                                ) : (
                                    (() => {
                                        const isPendingNav = isPending && pendingPath === item.href;
                                        return (
                                            <Link
                                                href={item.href}
                                                onClick={(e) => {
                                                    e.preventDefault();
                                                    navigate(item.href);
                                                }}
                                                prefetch={true}
                                                className={cn(
                                                    'flex items-center gap-3 rounded-lg px-3 py-1.5 text-sm text-foreground transition-all w-full text-left',
                                                    isActive || isPendingNav
                                                        ? 'bg-accent font-semibold'
                                                        : 'font-normal hover:bg-accent/50 hover:font-medium',
                                                    isPendingNav && 'opacity-60'
                                                )}
                                            >
                                                <Icon className={cn('h-4 w-4', isPendingNav && 'animate-pulse')} />
                                                <span>{item.title}</span>
                                                {/* Removed: Inbox unread badge (Stalwart inbox feature removed) */}
                                                {isPendingNav && <Loader2 className="ml-auto h-4 w-4 animate-spin" />}
                                            </Link>
                                        );
                                    })()
                                )}
                            </li>
                        );
                    })}
                </ul>
            </nav>

            {/* Settings at Bottom */}
            <div className="p-3 border-t border-border">
                <Link
                    href="/dashboard/settings"
                    className={cn(
                        'flex items-center gap-3 rounded-lg px-3 py-1.5 text-sm text-foreground transition-all',
                        pathname.startsWith('/dashboard/settings')
                            ? 'bg-accent font-semibold'
                            : 'font-normal hover:bg-accent/50 hover:font-medium'
                    )}
                >
                    <Settings className="h-4 w-4" />
                    <span>Settings</span>
                </Link>
            </div>
        </aside>
    );
}
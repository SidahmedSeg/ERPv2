import { usePathname } from 'next/navigation';
import { NAVIGATION_ITEMS } from '@/lib/constants';

interface BreadcrumbItem {
    label: string;
    href?: string;
}

export function useBreadcrumbs(): BreadcrumbItem[] {
    const pathname = usePathname();

    // Always start with Dashboard
    const breadcrumbs: BreadcrumbItem[] = [
        { label: 'Dashboard', href: '/dashboard' }
    ];

    // If we're just on dashboard, return early
    if (pathname === '/dashboard') {
        return breadcrumbs;
    }

    // First pass: Check items WITH submenus (higher priority)
    for (const navItem of NAVIGATION_ITEMS) {
        if (navItem.submenu && navItem.submenu.length > 0) {
            // Sort submenu items by path length (longest first) to match most specific paths first
            const sortedSubmenu = [...navItem.submenu].sort((a, b) => b.href.length - a.href.length);

            // Check if current path matches any submenu item
            const submenuItem = sortedSubmenu.find(
                sub => pathname === sub.href || pathname.startsWith(sub.href + '/')
            );

            if (submenuItem) {
                // Add parent menu item
                breadcrumbs.push({
                    label: navItem.title,
                    href: navItem.href
                });

                // Add current submenu item (no href for current page)
                breadcrumbs.push({
                    label: submenuItem.title
                });

                return breadcrumbs;
            }
        }
    }

    // Second pass: Check top-level items WITHOUT submenus
    for (const navItem of NAVIGATION_ITEMS) {
        // Skip items with submenus (already checked above)
        if (navItem.submenu && navItem.submenu.length > 0) {
            continue;
        }

        // Skip Dashboard item (already in breadcrumbs)
        if (navItem.href === '/dashboard') {
            continue;
        }

        // Check if current path matches top-level nav item
        if (pathname === navItem.href || pathname.startsWith(navItem.href + '/')) {
            breadcrumbs.push({
                label: navItem.title
            });
            return breadcrumbs;
        }
    }

    // Fallback: parse pathname if no match found
    const segments = pathname.split('/').filter(Boolean);
    const pathSegments = segments.slice(1); // Remove 'dashboard'

    pathSegments.forEach((segment, index) => {
        // Skip UUID segments
        const isUUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment);
        if (isUUID) {
            return;
        }

        // Capitalize and format segment
        const label = segment
            .split('-')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1))
            .join(' ');

        // Build href for intermediate segments
        const href = index < pathSegments.length - 1
            ? `/dashboard/${pathSegments.slice(0, index + 1).join('/')}`
            : undefined;

        breadcrumbs.push({ label, href });
    });

    return breadcrumbs;
}
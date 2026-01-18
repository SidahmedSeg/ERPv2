// src/hooks/usePermission.ts
'use client';

import { useAuthStore } from '@/store/auth-store';

/**
 * Hook to check if the current user has a specific permission
 * @param resource - The resource name (e.g., 'purchase_receipts', 'purchase_orders')
 * @param action - The action name (e.g., 'view', 'create', 'edit', 'delete', 'approve')
 * @returns boolean indicating if user has the permission
 */
export function usePermission(resource: string, action: string): boolean {
    const { user } = useAuthStore();

    if (!user || !(user as any).permissions) {
        return false;
    }

    return (user as any).permissions.some(
        (permission: any) =>
            permission.resource === resource &&
            permission.action === action
    );
}

/**
 * Hook to check multiple permissions at once
 * @param resource - The resource name
 * @param actions - Array of action names to check
 * @returns Object with action names as keys and boolean permission status as values
 */
export function usePermissions(resource: string, actions: string[]): Record<string, boolean> {
    const { user } = useAuthStore();

    if (!user || !(user as any).permissions) {
        return actions.reduce((acc, action) => ({ ...acc, [action]: false }), {});
    }

    return actions.reduce((acc, action) => ({
        ...acc,
        [action]: (user as any).permissions.some(
            (permission: any) =>
                permission.resource === resource &&
                permission.action === action
        )
    }), {});
}

/**
 * Hook to get all permissions for a specific resource
 * @param resource - The resource name
 * @returns Array of action names the user has permission for
 */
export function useResourcePermissions(resource: string): string[] {
    const { user } = useAuthStore();

    if (!user || !(user as any).permissions) {
        return [];
    }

    return (user as any).permissions
        .filter((permission: any) => permission.resource === resource)
        .map((permission: any) => permission.action);
}

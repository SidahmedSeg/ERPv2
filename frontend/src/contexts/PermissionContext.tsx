'use client';

import React, { createContext, useContext, useEffect, useState, useMemo, useCallback, useRef } from 'react';

interface Permission {
  resource: string;
  action: string;
}

interface UserPermissions {
  permissions: Permission[];
}

interface PermissionContextType {
  permissions: UserPermissions | null;
  loading: boolean;
  error: string | null;
  hasPermission: (resource: string, action: string) => boolean;
  refetch: () => Promise<void>;
}

const PermissionContext = createContext<PermissionContextType | undefined>(undefined);

const PERMISSIONS_CACHE_KEY = 'user_permissions_cache';

export function PermissionProvider({ children }: { children: React.ReactNode }) {
  const [permissions, setPermissions] = useState<UserPermissions | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const permissionCache = useRef<Map<string, boolean>>(new Map());

  const fetchPermissions = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // Try to load from sessionStorage first
      const cachedPermissions = sessionStorage.getItem(PERMISSIONS_CACHE_KEY);
      if (cachedPermissions) {
        try {
          const parsed = JSON.parse(cachedPermissions);
          setPermissions(parsed);
          setLoading(false);
          // Still fetch fresh data in background but don't block
          fetchFreshPermissions();
          return;
        } catch (e) {
          // Invalid cache, continue to fetch
          sessionStorage.removeItem(PERMISSIONS_CACHE_KEY);
        }
      }

      await fetchFreshPermissions();
    } catch (err) {
      console.error('Error fetching permissions:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch permissions');
      setLoading(false);
    }
  }, []);

  const fetchFreshPermissions = async () => {
    // Get token from cookies (auth_token) or localStorage (token) as fallback
    const getCookie = (name: string) => {
      const value = `; ${document.cookie}`;
      const parts = value.split(`; ${name}=`);
      if (parts.length === 2) return parts.pop()?.split(';').shift();
      return null;
    };

    const token = getCookie('auth_token') || localStorage.getItem('token');
    if (!token) {
      setError('No authentication token found');
      setLoading(false);
      return;
    }

    const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/auth/me/permissions`, {
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch permissions: ${response.statusText}`);
    }

    const data = await response.json();
    setPermissions(data.data);
    // Cache in sessionStorage
    sessionStorage.setItem(PERMISSIONS_CACHE_KEY, JSON.stringify(data.data));
    // Clear permission cache when permissions change
    permissionCache.current = new Map();
    setLoading(false);
  };

  useEffect(() => {
    fetchPermissions();
  }, [fetchPermissions]);

  // Memoized hasPermission function with caching
  const hasPermission = useCallback((resource: string, action: string): boolean => {
    if (!permissions || !permissions.permissions) {
      return false;
    }

    // Create cache key
    const cacheKey = `${resource}:${action}`;

    // Check cache first
    if (permissionCache.current.has(cacheKey)) {
      return permissionCache.current.get(cacheKey)!;
    }

    // Compute permission
    const result = permissions.permissions.some(
      (perm) =>
        perm.resource === resource &&
        (perm.action === action || perm.action === '*')
    );

    // Store in cache (this won't trigger a re-render)
    permissionCache.current.set(cacheKey, result);

    return result;
  }, [permissions]);

  const value: PermissionContextType = {
    permissions,
    loading,
    error,
    hasPermission,
    refetch: fetchPermissions,
  };

  return (
    <PermissionContext.Provider value={value}>
      {children}
    </PermissionContext.Provider>
  );
}

export function usePermissionContext() {
  const context = useContext(PermissionContext);
  if (context === undefined) {
    throw new Error('usePermissionContext must be used within a PermissionProvider');
  }
  return context;
}

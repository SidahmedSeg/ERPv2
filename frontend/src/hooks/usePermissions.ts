import { usePermissionContext } from '@/contexts/PermissionContext';

export function usePermissions() {
  const { hasPermission, permissions, loading, error, refetch } = usePermissionContext();

  return {
    hasPermission,
    permissions,
    loading,
    error,
    refetch,
    // Convenience methods for common checks
    canView: (resource: string) => hasPermission(resource, 'view'),
    canCreate: (resource: string) => hasPermission(resource, 'create'),
    canEdit: (resource: string) => hasPermission(resource, 'edit'),
    canDelete: (resource: string) => hasPermission(resource, 'delete'),
    canExport: (resource: string) => hasPermission(resource, 'export'),
    canImport: (resource: string) => hasPermission(resource, 'import'),
  };
}

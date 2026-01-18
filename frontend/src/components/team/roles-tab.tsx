'use client';

import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import {
  Shield,
  Plus,
  Search,
  MoreVertical,
  Pencil,
  Trash2,
  Lock,
  Users,
  CheckSquare,
  Square,
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  rolesApi,
  permissionsApi,
  Role,
  RoleWithPermissions,
  Permission,
} from '@/lib/team-api';

export function RolesTab() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [permissions, setPermissions] = useState<Record<string, Permission[]>>({});
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<RoleWithPermissions | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    display_name: '',
    description: '',
    permission_ids: [] as string[],
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [rolesData, permissionsData] = await Promise.all([
        rolesApi.list(),
        permissionsApi.listAll(true) as Promise<Record<string, Permission[]>>,
      ]);
      setRoles(rolesData);
      setPermissions(permissionsData);
    } catch (error) {
      console.error('Failed to fetch roles data:', error);
      toast.error('Failed to load roles');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingRole(null);
    setFormData({
      name: '',
      display_name: '',
      description: '',
      permission_ids: [],
    });
    setDialogOpen(true);
  };

  const handleEdit = async (role: Role) => {
    try {
      const roleWithPerms = await rolesApi.get(role.id);
      setEditingRole(roleWithPerms);
      setFormData({
        name: roleWithPerms.name,
        display_name: roleWithPerms.display_name,
        description: roleWithPerms.description || '',
        permission_ids: roleWithPerms.permissions.map((p) => p.id),
      });
      setDialogOpen(true);
    } catch (error) {
      toast.error('Failed to load role details');
    }
  };

  const handleSubmit = async () => {
    try {
      if (!formData.display_name.trim()) {
        toast.error('Display name is required');
        return;
      }

      if (editingRole) {
        await rolesApi.update(editingRole.id, {
          display_name: formData.display_name,
          description: formData.description,
          permission_ids: formData.permission_ids,
        });
        toast.success('Role updated successfully');
      } else {
        if (!formData.name.trim()) {
          toast.error('Role name is required');
          return;
        }
        await rolesApi.create(formData);
        toast.success('Role created successfully');
      }

      setDialogOpen(false);
      fetchData();
    } catch (error) {
      toast.error(editingRole ? 'Failed to update role' : 'Failed to create role');
    }
  };

  const handleDelete = async (role: Role) => {
    if (role.is_system) {
      toast.error('Cannot delete system roles');
      return;
    }

    if (!confirm(`Are you sure you want to delete the role "${role.display_name}"?`)) return;

    try {
      await rolesApi.delete(role.id);
      toast.success('Role deleted successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to delete role');
    }
  };

  const togglePermission = (permissionId: string) => {
    setFormData((prev) => ({
      ...prev,
      permission_ids: prev.permission_ids.includes(permissionId)
        ? prev.permission_ids.filter((id) => id !== permissionId)
        : [...prev.permission_ids, permissionId],
    }));
  };

  const toggleCategory = (category: string) => {
    const categoryPermissions = permissions[category] || [];
    const categoryPermissionIds = categoryPermissions.map((p) => p.id);
    const allSelected = categoryPermissionIds.every((id) =>
      formData.permission_ids.includes(id)
    );

    setFormData((prev) => ({
      ...prev,
      permission_ids: allSelected
        ? prev.permission_ids.filter((id) => !categoryPermissionIds.includes(id))
        : [...new Set([...prev.permission_ids, ...categoryPermissionIds])],
    }));
  };

  const filteredRoles = (roles || []).filter((role) =>
    role.display_name.toLowerCase().includes(search.toLowerCase()) ||
    role.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b bg-white">
        {/* Search and Actions */}
        <div className="flex gap-3 items-center">
          <div className="relative w-80">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search roles..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="ml-auto">
            <Button onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-2" />
              Create Role
            </Button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="space-y-4">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-24 w-full" />
            ))}
          </div>
        ) : filteredRoles.length === 0 ? (
          <div className="text-center py-12">
            <Shield className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No roles found</h3>
            <p className="text-gray-600">Create a custom role to get started</p>
          </div>
        ) : (
          <div className="bg-white rounded-lg border">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Role
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Description
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Type
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredRoles.map((role) => (
                  <tr key={role.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-lg bg-blue-100 flex items-center justify-center">
                          <Shield className="h-5 w-5 text-blue-600" />
                        </div>
                        <div>
                          <div className="font-medium text-gray-900">{role.display_name}</div>
                          <div className="text-xs text-gray-500">{role.name}</div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-gray-700">
                        {role.description || <span className="text-gray-400">No description</span>}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      {role.is_system ? (
                        <Badge variant="secondary" className="text-xs">
                          <Lock className="h-3 w-3 mr-1" />
                          System
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-xs">
                          Custom
                        </Badge>
                      )}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {!role.is_system && (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => handleEdit(role)}>
                              <Pencil className="h-4 w-4 mr-2" />
                              Edit
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() => handleDelete(role)}
                              className="text-red-600"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              Delete
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Create/Edit Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-3xl max-h-[80vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>
              {editingRole ? `Edit Role: ${editingRole.display_name}` : 'Create New Role'}
            </DialogTitle>
            <DialogDescription>
              {editingRole
                ? 'Update the role details and permissions'
                : 'Create a new role and assign permissions'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            {!editingRole && (
              <div className="space-y-2">
                <Label htmlFor="name">Role Name (lowercase, no spaces)</Label>
                <Input
                  id="name"
                  placeholder="e.g., sales_manager"
                  value={formData.name}
                  onChange={(e) =>
                    setFormData({ ...formData, name: e.target.value.toLowerCase().replace(/\s/g, '_') })
                  }
                />
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="display_name">Display Name</Label>
              <Input
                id="display_name"
                placeholder="e.g., Sales Manager"
                value={formData.display_name}
                onChange={(e) => setFormData({ ...formData, display_name: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe the role and its responsibilities"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label>Permissions ({formData.permission_ids.length} selected)</Label>
              <Accordion type="multiple" className="w-full">
                {Object.entries(permissions).map(([category, perms]) => {
                  const categoryPermissionIds = perms.map((p) => p.id);
                  const allSelected = categoryPermissionIds.every((id) =>
                    formData.permission_ids.includes(id)
                  );
                  const someSelected = categoryPermissionIds.some((id) =>
                    formData.permission_ids.includes(id)
                  );

                  return (
                    <AccordionItem key={category} value={category}>
                      <AccordionTrigger className="hover:no-underline">
                        <div className="flex items-center gap-2">
                          <button
                            type="button"
                            onClick={(e) => {
                              e.stopPropagation();
                              toggleCategory(category);
                            }}
                            className="p-1 hover:bg-gray-100 rounded"
                          >
                            {allSelected ? (
                              <CheckSquare className="h-4 w-4 text-blue-600" />
                            ) : someSelected ? (
                              <CheckSquare className="h-4 w-4 text-gray-400" />
                            ) : (
                              <Square className="h-4 w-4 text-gray-400" />
                            )}
                          </button>
                          <span className="font-medium">{category}</span>
                          <Badge variant="secondary" className="text-xs">
                            {perms.length}
                          </Badge>
                        </div>
                      </AccordionTrigger>
                      <AccordionContent>
                        <div className="border rounded-md">
                          <Table>
                            <TableHeader>
                              <TableRow className="bg-gray-50">
                                <TableHead className="w-[50%]">Permission Name</TableHead>
                                <TableHead className="w-[35%]">Resource.Action</TableHead>
                                <TableHead className="w-[15%] text-center">Select</TableHead>
                              </TableRow>
                            </TableHeader>
                            <TableBody>
                              {perms.map((permission) => (
                                <TableRow key={permission.id} className="hover:bg-gray-50">
                                  <TableCell>
                                    <div>
                                      <div className="font-medium text-sm">{permission.display_name}</div>
                                      {permission.description && (
                                        <div className="text-xs text-gray-500 mt-1">{permission.description}</div>
                                      )}
                                    </div>
                                  </TableCell>
                                  <TableCell>
                                    <code className="text-xs text-gray-600 bg-gray-100 px-2 py-1 rounded">
                                      {permission.resource}.{permission.action}
                                    </code>
                                  </TableCell>
                                  <TableCell className="text-center">
                                    <input
                                      type="checkbox"
                                      checked={formData.permission_ids.includes(permission.id)}
                                      onChange={() => togglePermission(permission.id)}
                                      className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer"
                                    />
                                  </TableCell>
                                </TableRow>
                              ))}
                            </TableBody>
                          </Table>
                        </div>
                      </AccordionContent>
                    </AccordionItem>
                  );
                })}
              </Accordion>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit}>
              {editingRole ? 'Update Role' : 'Create Role'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

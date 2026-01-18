'use client';

import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import {
  Building2,
  Plus,
  Search,
  MoreVertical,
  Pencil,
  Trash2,
  Users,
  Archive,
  CheckCircle,
  LayoutGrid,
  Table,
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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Skeleton } from '@/components/ui/skeleton';
import {
  departmentsApi,
  teamMembersApi,
  DepartmentWithDetails,
  TeamMember,
} from '@/lib/team-api';

const DEPARTMENT_COLORS = [
  '#3B82F6', // Blue
  '#8B5CF6', // Purple
  '#EC4899', // Pink
  '#F59E0B', // Amber
  '#10B981', // Green
  '#14B8A6', // Teal
  '#EF4444', // Red
  '#6366F1', // Indigo
];

const DEPARTMENT_ICONS = [
  'building-2',
  'users',
  'briefcase',
  'code',
  'megaphone',
  'dollar-sign',
  'package',
  'headphones',
];

export function DepartmentsTab() {
  const [departments, setDepartments] = useState<DepartmentWithDetails[]>([]);
  const [teamMembers, setTeamMembers] = useState<TeamMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingDepartment, setEditingDepartment] = useState<DepartmentWithDetails | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    head_user_id: '',
    color: '#3B82F6',
    icon: 'users',
  });

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [departmentsData, membersData] = await Promise.all([
        departmentsApi.list(),
        teamMembersApi.list({ status: 'active' }),
      ]);
      setDepartments(departmentsData);
      setTeamMembers(membersData as TeamMember[]);
    } catch (error) {
      console.error('Failed to fetch departments data:', error);
      toast.error('Failed to load departments');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingDepartment(null);
    setFormData({
      name: '',
      description: '',
      head_user_id: '',
      color: '#3B82F6',
      icon: 'users',
    });
    setDialogOpen(true);
  };

  const handleEdit = (department: DepartmentWithDetails) => {
    setEditingDepartment(department);
    setFormData({
      name: department.name,
      description: department.description || '',
      head_user_id: department.head_user_id || '',
      color: department.color,
      icon: department.icon,
    });
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    try {
      if (!formData.name.trim()) {
        toast.error('Department name is required');
        return;
      }

      const data = {
        ...formData,
        head_user_id: formData.head_user_id || undefined,
      };

      if (editingDepartment) {
        await departmentsApi.update(editingDepartment.id, data);
        toast.success('Department updated successfully');
      } else {
        await departmentsApi.create(data);
        toast.success('Department created successfully');
      }

      setDialogOpen(false);
      fetchData();
    } catch (error) {
      toast.error(editingDepartment ? 'Failed to update department' : 'Failed to create department');
    }
  };

  const handleDelete = async (department: DepartmentWithDetails) => {
    if (!confirm(`Are you sure you want to delete "${department.name}"? This will remove ${department.member_count} members from this department.`)) {
      return;
    }

    try {
      await departmentsApi.delete(department.id);
      toast.success('Department deleted successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to delete department');
    }
  };

  const handleArchive = async (department: DepartmentWithDetails) => {
    try {
      await departmentsApi.update(department.id, {
        status: department.status === 'active' ? 'inactive' : 'active',
      });
      toast.success(
        department.status === 'active'
          ? 'Department archived successfully'
          : 'Department activated successfully'
      );
      fetchData();
    } catch (error) {
      toast.error('Failed to update department status');
    }
  };

  const filteredDepartments = (departments || []).filter((dept) =>
    dept.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b bg-white">
        {/* Search and Actions */}
        <div className="flex gap-3 items-center">
          <div className="relative w-80">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search departments..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="flex gap-2 ml-auto">
            <div className="flex items-center gap-0 border rounded-md">
              <Button
                variant={viewMode === 'grid' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('grid')}
                className="rounded-r-none h-9"
              >
                <LayoutGrid className="h-4 w-4" />
              </Button>
              <Button
                variant={viewMode === 'table' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('table')}
                className="rounded-l-none h-9"
              >
                <Table className="h-4 w-4" />
              </Button>
            </div>
            <Button onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-2" />
              Create Department
            </Button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="space-y-4">
            {[...Array(4)].map((_, i) => (
              <Skeleton key={i} className="h-32 w-full" />
            ))}
          </div>
        ) : filteredDepartments.length === 0 ? (
          <div className="text-center py-12">
            <Building2 className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No departments found</h3>
            <p className="text-gray-600">Create a department to organize your team</p>
          </div>
        ) : viewMode === 'grid' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredDepartments.map((department) => (
              <div
                key={department.id}
                className="bg-white rounded-lg border p-6 hover:shadow-md transition-shadow"
              >
                <div className="flex items-start justify-between mb-4">
                  <div className="flex items-center gap-3">
                    <div
                      className="h-12 w-12 rounded-lg flex items-center justify-center"
                      style={{ backgroundColor: `${department.color}20` }}
                    >
                      <Building2 className="h-6 w-6" style={{ color: department.color }} />
                    </div>
                    <div>
                      <h3 className="font-semibold text-gray-900">{department.name}</h3>
                      <div className="flex items-center gap-2 mt-1">
                        <Users className="h-3 w-3 text-gray-400" />
                        <span className="text-xs text-gray-500">
                          {department.member_count} {department.member_count === 1 ? 'member' : 'members'}
                        </span>
                      </div>
                    </div>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => handleEdit(department)}>
                        <Pencil className="h-4 w-4 mr-2" />
                        Edit
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => handleArchive(department)}>
                        {department.status === 'active' ? (
                          <>
                            <Archive className="h-4 w-4 mr-2" />
                            Archive
                          </>
                        ) : (
                          <>
                            <CheckCircle className="h-4 w-4 mr-2" />
                            Activate
                          </>
                        )}
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        onClick={() => handleDelete(department)}
                        className="text-red-600"
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>

                {department.description && (
                  <p className="text-sm text-gray-600 mb-4">{department.description}</p>
                )}

                {department.head_user_name && (
                  <div className="flex items-center gap-2 text-sm text-gray-600">
                    <Users className="h-4 w-4 text-gray-400" />
                    <span>Head: {department.head_user_name}</span>
                  </div>
                )}

                <div className="flex items-center gap-2 mt-4">
                  <Badge variant={department.status === 'active' ? 'default' : 'secondary'}>
                    {department.status}
                  </Badge>
                </div>
              </div>
            ))}
          </div>
        ) : (
          <div className="bg-white rounded-lg border">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Department
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Head
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Members
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredDepartments.map((department) => (
                  <tr key={department.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div
                          className="h-10 w-10 rounded-lg flex items-center justify-center"
                          style={{ backgroundColor: `${department.color}20` }}
                        >
                          <Building2 className="h-5 w-5" style={{ color: department.color }} />
                        </div>
                        <div>
                          <div className="font-medium text-gray-900">{department.name}</div>
                          {department.description && (
                            <div className="text-sm text-gray-500">{department.description}</div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {department.head_user_name ? (
                        <div className="flex items-center gap-2">
                          <Users className="h-4 w-4 text-gray-400" />
                          <span className="text-sm text-gray-700">{department.head_user_name}</span>
                        </div>
                      ) : (
                        <span className="text-sm text-gray-400">No head assigned</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <span className="text-sm text-gray-700">
                        {department.member_count} {department.member_count === 1 ? 'member' : 'members'}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <Badge variant={department.status === 'active' ? 'default' : 'secondary'}>
                        {department.status}
                      </Badge>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEdit(department)}>
                            <Pencil className="h-4 w-4 mr-2" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleArchive(department)}>
                            {department.status === 'active' ? (
                              <>
                                <Archive className="h-4 w-4 mr-2" />
                                Archive
                              </>
                            ) : (
                              <>
                                <CheckCircle className="h-4 w-4 mr-2" />
                                Activate
                              </>
                            )}
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleDelete(department)}
                            className="text-red-600"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
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
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingDepartment ? `Edit Department: ${editingDepartment.name}` : 'Create New Department'}
            </DialogTitle>
            <DialogDescription>
              {editingDepartment
                ? 'Update the department details'
                : 'Create a new department to organize your team'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Department Name</Label>
              <Input
                id="name"
                placeholder="e.g., Engineering, Sales, Marketing"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                placeholder="Describe the department's responsibilities"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={3}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="head_user_id">Department Head (Optional)</Label>
              <Select
                value={formData.head_user_id || "none"}
                onValueChange={(value) => setFormData({ ...formData, head_user_id: value === "none" ? "" : value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a department head" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No department head</SelectItem>
                  {teamMembers.map((member) => (
                    <SelectItem key={member.id} value={member.id}>
                      {member.name} ({member.email})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>Department Color</Label>
              <div className="flex gap-2">
                {DEPARTMENT_COLORS.map((color) => (
                  <button
                    key={color}
                    type="button"
                    className={`h-10 w-10 rounded-lg border-2 transition-all ${
                      formData.color === color ? 'border-gray-900 scale-110' : 'border-gray-200'
                    }`}
                    style={{ backgroundColor: color }}
                    onClick={() => setFormData({ ...formData, color })}
                  />
                ))}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit}>
              {editingDepartment ? 'Update Department' : 'Create Department'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

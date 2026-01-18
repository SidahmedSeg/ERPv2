'use client';

import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import {
  Users,
  Plus,
  Search,
  MoreVertical,
  Pencil,
  Ban,
  Trash2,
  Mail,
  Shield,
  Building2,
  UserCog,
} from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
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
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { teamMembersApi, TeamMemberWithRoles, departmentsApi, DepartmentWithDetails, rolesApi, Role } from '@/lib/team-api';
import { Skeleton } from '@/components/ui/skeleton';

interface MembersTabProps {
  onSwitchToInvitations?: () => void;
}

export function MembersTab({ onSwitchToInvitations }: MembersTabProps) {
  const [members, setMembers] = useState<TeamMemberWithRoles[]>([]);
  const [departments, setDepartments] = useState<DepartmentWithDetails[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterDepartment, setFilterDepartment] = useState('');

  // Dialog states
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [manageRolesDialogOpen, setManageRolesDialogOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState<TeamMemberWithRoles | null>(null);
  const [editFormData, setEditFormData] = useState({
    name: '',
    job_title: '',
    phone: '',
    department_id: '',
  });
  const [selectedRoleIds, setSelectedRoleIds] = useState<string[]>([]);

  useEffect(() => {
    fetchData();
  }, [filterStatus, filterDepartment]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [membersData, departmentsData, rolesData] = await Promise.all([
        teamMembersApi.list({
          status: filterStatus || undefined,
          department_id: filterDepartment || undefined,
          with_roles: true,
        }),
        departmentsApi.list(),
        rolesApi.list(),
      ]);
      setMembers(membersData as TeamMemberWithRoles[]);
      setDepartments(departmentsData);
      setRoles(rolesData);
    } catch (error) {
      console.error('Failed to fetch team data:', error);
      toast.error('Failed to load team members');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateStatus = async (id: string, status: string) => {
    try {
      await teamMembersApi.updateStatus(id, { status });
      toast.success(`Member ${status === 'suspended' ? 'suspended' : 'activated'} successfully`);
      fetchData();
    } catch (error) {
      toast.error('Failed to update member status');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to deactivate this member?')) return;
    try {
      await teamMembersApi.delete(id);
      toast.success('Member deactivated successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to deactivate member');
    }
  };

  const handleEditMember = (member: TeamMemberWithRoles) => {
    setSelectedMember(member);
    setEditFormData({
      name: member.name,
      job_title: member.job_title || '',
      phone: member.phone || '',
      department_id: member.department_id || '',
    });
    setEditDialogOpen(true);
  };

  const handleEditSubmit = async () => {
    if (!selectedMember) return;

    try {
      if (!editFormData.name.trim()) {
        toast.error('Name is required');
        return;
      }

      await teamMembersApi.update(selectedMember.id, {
        name: editFormData.name,
        job_title: editFormData.job_title || undefined,
        phone: editFormData.phone || undefined,
        department_id: editFormData.department_id || undefined,
      });

      toast.success('Member updated successfully');
      setEditDialogOpen(false);
      fetchData();
    } catch (error) {
      toast.error('Failed to update member');
    }
  };

  const handleManageRoles = async (member: TeamMemberWithRoles) => {
    setSelectedMember(member);

    try {
      // Get role IDs from role names
      const memberRoleIds = roles
        .filter(role => member.roles.includes(role.display_name))
        .map(role => role.id);

      setSelectedRoleIds(memberRoleIds);
      setManageRolesDialogOpen(true);
    } catch (error) {
      toast.error('Failed to load member roles');
    }
  };

  const handleRolesSubmit = async () => {
    if (!selectedMember) return;

    try {
      await rolesApi.assignRoles(selectedMember.id, {
        role_ids: selectedRoleIds,
      });

      toast.success('Roles updated successfully');
      setManageRolesDialogOpen(false);
      fetchData();
    } catch (error) {
      toast.error('Failed to update roles');
    }
  };

  const toggleRole = (roleId: string) => {
    setSelectedRoleIds((prev) =>
      prev.includes(roleId)
        ? prev.filter((id) => id !== roleId)
        : [...prev, roleId]
    );
  };

  const getStatusBadge = (status: string) => {
    const variants: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
      active: { label: 'Active', variant: 'default' },
      suspended: { label: 'Suspended', variant: 'secondary' },
      deactivated: { label: 'Deactivated', variant: 'destructive' },
    };
    const config = variants[status] || variants.active;
    return <Badge variant={config.variant}>{config.label}</Badge>;
  };

  const filteredMembers = members.filter((member) =>
    member.name.toLowerCase().includes(search.toLowerCase()) ||
    member.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b bg-white">
        {/* Filters */}
        <div className="flex gap-3 items-center">
          <div className="relative w-80">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search members..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select value={filterStatus || "all"} onValueChange={(val) => setFilterStatus(val === "all" ? "" : val)}>
            <SelectTrigger className="w-[180px]">
              <SelectValue placeholder="All Statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Statuses</SelectItem>
              <SelectItem value="active">Active</SelectItem>
              <SelectItem value="suspended">Suspended</SelectItem>
              <SelectItem value="deactivated">Deactivated</SelectItem>
            </SelectContent>
          </Select>
          <Select value={filterDepartment || "all"} onValueChange={(val) => setFilterDepartment(val === "all" ? "" : val)}>
            <SelectTrigger className="w-[200px]">
              <SelectValue placeholder="All Departments" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Departments</SelectItem>
              {departments?.map((dept) => (
                <SelectItem key={dept.id} value={dept.id}>
                  {dept.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <div className="ml-auto">
            <Button onClick={onSwitchToInvitations}>
              <Mail className="h-4 w-4 mr-2" />
              Invite Member
            </Button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex-1 overflow-auto p-6">
        {loading ? (
          <div className="space-y-4">
            {[...Array(5)].map((_, i) => (
              <Skeleton key={i} className="h-20 w-full" />
            ))}
          </div>
        ) : filteredMembers.length === 0 ? (
          <div className="text-center py-12">
            <Users className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No team members found</h3>
            <p className="text-gray-600">Start by inviting your first team member</p>
          </div>
        ) : (
          <div className="bg-white rounded-lg border">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Member
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Department
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Roles
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Joined
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredMembers.map((member) => (
                  <tr key={member.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
                          <span className="text-blue-600 font-medium">
                            {member.name.charAt(0).toUpperCase()}
                          </span>
                        </div>
                        <div>
                          <div className="font-medium text-gray-900">{member.name}</div>
                          <div className="text-sm text-gray-500">{member.email}</div>
                          {member.job_title && (
                            <div className="text-xs text-gray-400">{member.job_title}</div>
                          )}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {member.department_name ? (
                        <div className="flex items-center gap-2">
                          <Building2 className="h-4 w-4 text-gray-400" />
                          <span className="text-sm text-gray-700">{member.department_name}</span>
                        </div>
                      ) : (
                        <span className="text-sm text-gray-400">No department</span>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {member.roles && member.roles.length > 0 ? (
                          member.roles.map((role) => (
                            <Badge key={role} variant="outline" className="text-xs">
                              <Shield className="h-3 w-3 mr-1" />
                              {role}
                            </Badge>
                          ))
                        ) : (
                          <span className="text-sm text-gray-400">No roles</span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">{getStatusBadge(member.status)}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {new Date(member.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => handleEditMember(member)}>
                            <Pencil className="h-4 w-4 mr-2" />
                            Edit Details
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => handleManageRoles(member)}>
                            <UserCog className="h-4 w-4 mr-2" />
                            Manage Roles
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          {member.status === 'active' ? (
                            <DropdownMenuItem onClick={() => handleUpdateStatus(member.id, 'suspended')}>
                              <Ban className="h-4 w-4 mr-2" />
                              Suspend
                            </DropdownMenuItem>
                          ) : (
                            <DropdownMenuItem onClick={() => handleUpdateStatus(member.id, 'active')}>
                              <Shield className="h-4 w-4 mr-2" />
                              Activate
                            </DropdownMenuItem>
                          )}
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => handleDelete(member.id)}
                            className="text-red-600"
                          >
                            <Trash2 className="h-4 w-4 mr-2" />
                            Deactivate
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

      {/* Edit Member Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Member Details</DialogTitle>
            <DialogDescription>
              Update member information and department assignment
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="edit-name">Full Name *</Label>
              <Input
                id="edit-name"
                placeholder="John Doe"
                value={editFormData.name}
                onChange={(e) => setEditFormData({ ...editFormData, name: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-job-title">Job Title</Label>
              <Input
                id="edit-job-title"
                placeholder="Software Engineer"
                value={editFormData.job_title}
                onChange={(e) => setEditFormData({ ...editFormData, job_title: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-phone">Phone</Label>
              <Input
                id="edit-phone"
                placeholder="+1 (555) 000-0000"
                value={editFormData.phone}
                onChange={(e) => setEditFormData({ ...editFormData, phone: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="edit-department">Department</Label>
              <Select
                value={editFormData.department_id || "none"}
                onValueChange={(value) => setEditFormData({ ...editFormData, department_id: value === "none" ? "" : value })}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select a department" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">No department</SelectItem>
                  {departments?.map((dept) => (
                    <SelectItem key={dept.id} value={dept.id}>
                      {dept.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleEditSubmit}>
              Save Changes
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Manage Roles Dialog */}
      <Dialog open={manageRolesDialogOpen} onOpenChange={setManageRolesDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Manage Roles</DialogTitle>
            <DialogDescription>
              Assign roles to {selectedMember?.name}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Roles ({selectedRoleIds.length} selected)</Label>
              <div className="border rounded-lg p-4 space-y-2 max-h-96 overflow-y-auto">
                {roles?.map((role) => (
                  <label
                    key={role.id}
                    className="flex items-center gap-3 p-2 rounded hover:bg-gray-50 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedRoleIds.includes(role.id)}
                      onChange={() => toggleRole(role.id)}
                      className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary cursor-pointer"
                    />
                    <div className="flex-1">
                      <div className="font-medium text-sm">{role.display_name}</div>
                      {role.description && (
                        <div className="text-xs text-gray-500">{role.description}</div>
                      )}
                    </div>
                    {role.is_system && (
                      <Badge variant="secondary" className="text-xs">
                        System
                      </Badge>
                    )}
                  </label>
                ))}
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setManageRolesDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleRolesSubmit}>
              <Shield className="h-4 w-4 mr-2" />
              Update Roles
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

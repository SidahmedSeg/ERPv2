'use client';

import { useState, useEffect } from 'react';
import { toast } from 'sonner';
import {
  Mail,
  Plus,
  Search,
  MoreVertical,
  Send,
  X,
  Trash2,
  Clock,
  CheckCircle,
  XCircle,
  Shield,
  Building2,
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
  invitationsApi,
  rolesApi,
  departmentsApi,
  InvitationWithDetails,
  Role,
  DepartmentWithDetails,
} from '@/lib/team-api';

export function InvitationsTab() {
  const [invitations, setInvitations] = useState<InvitationWithDetails[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [departments, setDepartments] = useState<DepartmentWithDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [dialogOpen, setDialogOpen] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    role_ids: [] as string[],
    department_id: '',
    welcome_message: '',
  });

  useEffect(() => {
    fetchData();
  }, [filterStatus]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [invitationsData, rolesData, departmentsData] = await Promise.all([
        invitationsApi.list(filterStatus || undefined),
        rolesApi.list(),
        departmentsApi.list(),
      ]);
      setInvitations(invitationsData);
      setRoles(rolesData);
      setDepartments(departmentsData);
    } catch (error) {
      console.error('Failed to fetch invitations data:', error);
      toast.error('Failed to load invitations');
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setFormData({
      email: '',
      role_ids: [],
      department_id: '',
      welcome_message: '',
    });
    setDialogOpen(true);
  };

  const handleSubmit = async () => {
    try {
      if (!formData.email.trim()) {
        toast.error('Email is required');
        return;
      }

      if (formData.role_ids.length === 0) {
        toast.error('At least one role is required');
        return;
      }

      await invitationsApi.create({
        ...formData,
        department_id: formData.department_id || undefined,
      });
      toast.success('Invitation sent successfully');
      setDialogOpen(false);
      fetchData();
    } catch (error: any) {
      const errorMessage = error.response?.data?.error || error.response?.data?.message || 'Failed to send invitation';
      toast.error(errorMessage);
    }
  };

  const handleResend = async (id: string) => {
    try {
      await invitationsApi.resend(id);
      toast.success('Invitation resent successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to resend invitation');
    }
  };

  const handleRevoke = async (id: string) => {
    if (!confirm('Are you sure you want to revoke this invitation?')) return;

    try {
      await invitationsApi.revoke(id);
      toast.success('Invitation revoked successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to revoke invitation');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this invitation?')) return;

    try {
      await invitationsApi.delete(id);
      toast.success('Invitation deleted successfully');
      fetchData();
    } catch (error) {
      toast.error('Failed to delete invitation');
    }
  };

  const getStatusBadge = (invitation: InvitationWithDetails) => {
    if (invitation.status === 'accepted') {
      return (
        <Badge variant="default">
          <CheckCircle className="h-3 w-3 mr-1" />
          Accepted
        </Badge>
      );
    }
    if (invitation.status === 'revoked') {
      return (
        <Badge variant="destructive">
          <XCircle className="h-3 w-3 mr-1" />
          Revoked
        </Badge>
      );
    }
    if (invitation.is_expired) {
      return (
        <Badge variant="secondary">
          <Clock className="h-3 w-3 mr-1" />
          Expired
        </Badge>
      );
    }
    return (
      <Badge variant="secondary" className="bg-amber-100 text-amber-800 border-amber-200">
        <Clock className="h-3 w-3 mr-1" />
        Pending
      </Badge>
    );
  };

  const toggleRole = (roleId: string) => {
    setFormData((prev) => ({
      ...prev,
      role_ids: prev.role_ids.includes(roleId)
        ? prev.role_ids.filter((id) => id !== roleId)
        : [...prev.role_ids, roleId],
    }));
  };

  const filteredInvitations = (invitations || []).filter((invitation) =>
    invitation.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="h-full flex flex-col">
      <div className="p-6 border-b bg-white">
        {/* Search and Actions */}
        <div className="flex gap-3 items-center">
          <div className="relative w-80">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
            <Input
              placeholder="Search invitations..."
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
              <SelectItem value="pending">Pending</SelectItem>
              <SelectItem value="accepted">Accepted</SelectItem>
              <SelectItem value="revoked">Revoked</SelectItem>
            </SelectContent>
          </Select>
          <div className="ml-auto">
            <Button onClick={handleCreate}>
              <Plus className="h-4 w-4 mr-2" />
              Send Invitation
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
        ) : filteredInvitations.length === 0 ? (
          <div className="text-center py-12">
            <Mail className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No invitations found</h3>
            <p className="text-gray-600">Send an invitation to add team members</p>
          </div>
        ) : (
          <div className="bg-white rounded-lg border">
            <table className="w-full">
              <thead className="bg-gray-50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Email
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Roles
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Department
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Invited By
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Expires
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {filteredInvitations.map((invitation) => (
                  <tr key={invitation.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <Mail className="h-4 w-4 text-gray-400" />
                        <span className="text-sm text-gray-900">{invitation.email}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex flex-wrap gap-1">
                        {invitation.role_ids.map((roleId) => {
                          const role = roles?.find((r) => r.id === roleId);
                          return role ? (
                            <Badge key={roleId} variant="secondary" className="text-xs bg-blue-100 text-blue-800 border-blue-200">
                              <Shield className="h-3 w-3 mr-1" />
                              {role.display_name}
                            </Badge>
                          ) : null;
                        })}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      {invitation.department_name ? (
                        <div className="flex items-center gap-2">
                          <Building2 className="h-4 w-4 text-gray-400" />
                          <span className="text-sm text-gray-700">{invitation.department_name}</span>
                        </div>
                      ) : (
                        <span className="text-sm text-gray-400">No department</span>
                      )}
                    </td>
                    <td className="px-6 py-4">{getStatusBadge(invitation)}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {invitation.invited_by_name || 'System'}
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {new Date(invitation.expires_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          {invitation.status === 'pending' && !invitation.is_expired && (
                            <>
                              <DropdownMenuItem onClick={() => handleResend(invitation.id)}>
                                <Send className="h-4 w-4 mr-2" />
                                Resend
                              </DropdownMenuItem>
                              <DropdownMenuItem onClick={() => handleRevoke(invitation.id)}>
                                <X className="h-4 w-4 mr-2" />
                                Revoke
                              </DropdownMenuItem>
                              <DropdownMenuSeparator />
                            </>
                          )}
                          <DropdownMenuItem
                            onClick={() => handleDelete(invitation.id)}
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

      {/* Send Invitation Dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Send Team Invitation</DialogTitle>
            <DialogDescription>
              Invite a new team member by email. They'll receive a link to create their account.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email Address</Label>
              <Input
                id="email"
                type="email"
                placeholder="colleague@company.com"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <Label>Roles (Select at least one)</Label>
              <div className="border rounded-lg p-4 space-y-2 max-h-48 overflow-y-auto">
                {roles?.map((role) => (
                  <label
                    key={role.id}
                    className="flex items-center gap-3 p-2 rounded hover:bg-gray-50 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={formData.role_ids.includes(role.id)}
                      onChange={() => toggleRole(role.id)}
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

            <div className="space-y-2">
              <Label htmlFor="department_id">Department (Optional)</Label>
              <Select
                value={formData.department_id || "none"}
                onValueChange={(value) => setFormData({ ...formData, department_id: value === "none" ? "" : value })}
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

            <div className="space-y-2">
              <Label htmlFor="welcome_message">Welcome Message (Optional)</Label>
              <Textarea
                id="welcome_message"
                placeholder="Welcome to our team! We're excited to have you join us."
                value={formData.welcome_message}
                onChange={(e) => setFormData({ ...formData, welcome_message: e.target.value })}
                rows={3}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSubmit}>
              <Send className="h-4 w-4 mr-2" />
              Send Invitation
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { useRouter } from 'next/navigation';
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
  Filter,
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
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { teamMembersApi, TeamMemberWithRoles, departmentsApi, DepartmentWithDetails } from '@/lib/team-api';
import { Skeleton } from '@/components/ui/skeleton';

export default function TeamMembersPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  const [members, setMembers] = useState<TeamMemberWithRoles[]>([]);
  const [departments, setDepartments] = useState<DepartmentWithDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [mounted, setMounted] = useState(false);
  const [search, setSearch] = useState('');
  const [filterStatus, setFilterStatus] = useState('');
  const [filterDepartment, setFilterDepartment] = useState('');

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [mounted, isAuthenticated, router]);

  useEffect(() => {
    if (mounted && isAuthenticated) {
      fetchData();
    }
  }, [mounted, isAuthenticated, filterStatus, filterDepartment]);

  const fetchData = async () => {
    try {
      setLoading(true);
      const [membersData, departmentsData] = await Promise.all([
        teamMembersApi.list({
          status: filterStatus || undefined,
          department_id: filterDepartment || undefined,
          with_roles: true,
        }),
        departmentsApi.list(),
      ]);
      setMembers(membersData as TeamMemberWithRoles[]);
      setDepartments(departmentsData);
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

  if (!mounted) return null;

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b bg-white">
        <div className="p-6">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 flex items-center gap-2">
                <Users className="h-8 w-8" />
                Team Members
              </h1>
              <p className="text-gray-600 mt-1">Manage your team members and their roles</p>
            </div>
            <Button onClick={() => router.push('/dashboard/team/invitations')}>
              <Mail className="h-4 w-4 mr-2" />
              Invite Member
            </Button>
          </div>

          {/* Filters */}
          <div className="flex gap-4 items-center">
            <div className="flex-1 relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
              <Input
                placeholder="Search members..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-10"
              />
            </div>
            <Select value={filterStatus} onValueChange={setFilterStatus}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="All Statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All Statuses</SelectItem>
                <SelectItem value="active">Active</SelectItem>
                <SelectItem value="suspended">Suspended</SelectItem>
                <SelectItem value="deactivated">Deactivated</SelectItem>
              </SelectContent>
            </Select>
            <Select value={filterDepartment} onValueChange={setFilterDepartment}>
              <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="All Departments" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="">All Departments</SelectItem>
                {departments.map((dept) => (
                  <SelectItem key={dept.id} value={dept.id}>
                    {dept.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
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
                        {member.role === 'owner' && (
                          <Badge variant="default" className="text-xs bg-purple-600 hover:bg-purple-700">
                            <Shield className="h-3 w-3 mr-1" />
                            Owner
                          </Badge>
                        )}
                        {member.roles.length > 0 ? (
                          member.roles.map((role) => (
                            <Badge key={role} variant="outline" className="text-xs">
                              <Shield className="h-3 w-3 mr-1" />
                              {role}
                            </Badge>
                          ))
                        ) : (
                          !member.role || member.role === 'owner' ? null : (
                            <span className="text-sm text-gray-400">No roles</span>
                          )
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">{getStatusBadge(member.status)}</td>
                    <td className="px-6 py-4 text-sm text-gray-600">
                      {new Date(member.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      {member.role === 'owner' ? (
                        <span className="text-xs text-gray-400">Owner</span>
                      ) : (
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreVertical className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => router.push(`/dashboard/team/members/${member.id}`)}>
                              <Pencil className="h-4 w-4 mr-2" />
                              Edit Details
                            </DropdownMenuItem>
                            <DropdownMenuItem onClick={() => router.push(`/dashboard/team/members/${member.id}/roles`)}>
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
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}

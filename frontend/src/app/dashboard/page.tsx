'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { Users, CheckCircle, Shield, Mail, Activity, Clock, Globe, Calendar } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { userApi, roleApi, invitationApi, sessionApi } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import type { DashboardStats } from '@/types';
import { formatDate } from '@/lib/utils';

export default function DashboardPage() {
  const { user } = useAuthStore();
  const [stats, setStats] = useState<DashboardStats>({
    total_users: 0,
    active_users: 0,
    total_roles: 0,
    pending_invitations: 0,
  });
  const [sessionStats, setSessionStats] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchDashboardData = async () => {
      try {
        const [usersRes, rolesRes, invitationsRes, sessionsRes] = await Promise.all([
          userApi.list(1, 1),
          roleApi.list(false),
          invitationApi.list('pending', 1, 1),
          sessionApi.getStats(),
        ]);

        setStats({
          total_users: usersRes.data.meta?.total_count || 0,
          active_users: usersRes.data.meta?.total_count || 0,
          total_roles: rolesRes.data.data?.roles?.length || 0,
          pending_invitations: invitationsRes.data.meta?.total_count || 0,
        });

        setSessionStats(sessionsRes.data.data);
      } catch (error) {
        console.error('Failed to fetch dashboard data:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-foreground">
          Welcome back, {user?.first_name}!
        </h1>
        <p className="mt-2 text-muted-foreground">
          Here's what's happening with your organization today.
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Users</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_users}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Active Users</CardTitle>
            <CheckCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.active_users}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Roles</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.total_roles}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Pending Invites</CardTitle>
            <Mail className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{stats.pending_invitations}</div>
          </CardContent>
        </Card>
      </div>

      {/* Session & Security Info */}
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-2">
        {/* Session Information */}
        <Card>
          <CardHeader>
            <CardTitle>Your Active Sessions</CardTitle>
            <CardDescription>Information about your current sessions</CardDescription>
          </CardHeader>
          <CardContent>
            {sessionStats ? (
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Activity className="h-4 w-4 text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Active sessions</span>
                  </div>
                  <span className="text-sm font-medium">
                    {sessionStats.active_sessions || 0}
                  </span>
                </div>
                {sessionStats.last_activity_at && (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Clock className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">Last active</span>
                    </div>
                    <span className="text-sm font-medium">
                      {formatDate(sessionStats.last_activity_at, true)}
                    </span>
                  </div>
                )}
                {sessionStats.last_ip_address && (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4 text-muted-foreground" />
                      <span className="text-sm text-muted-foreground">Last IP</span>
                    </div>
                    <span className="text-sm font-medium">
                      {sessionStats.last_ip_address}
                    </span>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">No session data available</p>
            )}
          </CardContent>
        </Card>

        {/* Account Security */}
        <Card>
          <CardHeader>
            <CardTitle>Account Security</CardTitle>
            <CardDescription>Your security settings overview</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Two-Factor Authentication</span>
                <Badge variant={user?.two_factor_enabled ? "success" : "warning"}>
                  {user?.two_factor_enabled ? "Enabled" : "Not Enabled"}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Account Status</span>
                <Badge variant="success" className="capitalize">
                  {user?.status}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Member since</span>
                </div>
                <span className="text-sm font-medium">
                  {user?.created_at ? formatDate(user.created_at) : 'N/A'}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
          <CardDescription>Common tasks and shortcuts</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
            <Link
              href="/dashboard/team/members"
              className="flex items-center gap-4 p-4 rounded-lg border hover:border-primary hover:bg-accent transition-colors"
            >
              <Users className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-medium">Manage Users</h3>
                <p className="text-sm text-muted-foreground">View and manage team members</p>
              </div>
            </Link>

            <Link
              href="/dashboard/team/roles"
              className="flex items-center gap-4 p-4 rounded-lg border hover:border-primary hover:bg-accent transition-colors"
            >
              <Shield className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-medium">Manage Roles</h3>
                <p className="text-sm text-muted-foreground">Configure permissions</p>
              </div>
            </Link>

            <Link
              href="/dashboard/security"
              className="flex items-center gap-4 p-4 rounded-lg border hover:border-primary hover:bg-accent transition-colors"
            >
              <Shield className="h-8 w-8 text-primary" />
              <div>
                <h3 className="font-medium">Security</h3>
                <p className="text-sm text-muted-foreground">Review security settings</p>
              </div>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

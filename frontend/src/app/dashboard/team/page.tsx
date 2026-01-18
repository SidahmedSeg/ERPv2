'use client';

import { useState, useEffect } from 'react';
import { useAuthStore } from '@/store/auth-store';
import { useRouter } from 'next/navigation';
import { Users } from 'lucide-react';
import { MembersTab } from '@/components/team/members-tab';
import { RolesTab } from '@/components/team/roles-tab';
import { DepartmentsTab } from '@/components/team/departments-tab';
import { InvitationsTab } from '@/components/team/invitations-tab';

export default function TeamPage() {
  const { isAuthenticated } = useAuthStore();
  const router = useRouter();
  const [mounted, setMounted] = useState(false);
  const [activeTab, setActiveTab] = useState('members');

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (mounted && !isAuthenticated) {
      router.push('/auth/login');
    }
  }, [mounted, isAuthenticated, router]);

  if (!mounted) return null;

  const tabs = [
    { id: 'members', label: 'Members' },
    { id: 'roles', label: 'Roles' },
    { id: 'departments', label: 'Departments' },
    { id: 'invitations', label: 'Invitations' },
  ];

  return (
    <div className="h-full w-full flex flex-col bg-white">
      {/* Tabs */}
      <div className="border-b border-border px-6">
        <div className="flex gap-6">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`px-1 py-3 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-primary text-primary'
                  : 'border-transparent text-text-secondary hover:text-text-primary'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Tab Content */}
      <div className="flex-1 overflow-y-auto">
        {activeTab === 'members' && <MembersTab onSwitchToInvitations={() => setActiveTab('invitations')} />}
        {activeTab === 'roles' && <RolesTab />}
        {activeTab === 'departments' && <DepartmentsTab />}
        {activeTab === 'invitations' && <InvitationsTab />}
      </div>
    </div>
  );
}

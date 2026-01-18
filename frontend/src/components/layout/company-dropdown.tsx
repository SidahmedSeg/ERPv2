'use client';

import { useState, useEffect } from 'react';
import { ChevronDown, Building2, SlidersHorizontal, LogOut, Users } from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuthStore } from '@/store/auth-store';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import Image from 'next/image';

export function CompanyDropdown() {
    const router = useRouter();
    const { user, tenantSlug, clearAuth, token } = useAuthStore();
    const [open, setOpen] = useState(false);
    const [companyLogo, setCompanyLogo] = useState<string | null>(null);

    useEffect(() => {
        const fetchCompanyLogo = async () => {
            if (!token) return;

            try {
                const response = await fetch("http://localhost:8080/api/settings/company", {
                    headers: {
                        Authorization: `Bearer ${token}`,
                    },
                });

                const data = await response.json();
                if (data.success && data.data?.logo_url) {
                    setCompanyLogo(data.data.logo_url);
                }
            } catch (error) {
                console.error('Failed to fetch company logo:', error);
            }
        };

        fetchCompanyLogo();

        // Poll for logo changes every 10 seconds when dropdown is open
        const interval = open ? setInterval(fetchCompanyLogo, 10000) : null;

        return () => {
            if (interval) clearInterval(interval);
        };
    }, [token, open]);

    const handleLogout = async () => {
        try {
            await api.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            clearAuth();
            router.push('/auth/login');
        }
    };

    // Convert slug to company name (capitalize first letter)
    const companyName = tenantSlug
        ? tenantSlug.charAt(0).toUpperCase() + tenantSlug.slice(1)
        : 'Company';

    return (
        <DropdownMenu open={open} onOpenChange={setOpen}>
            <DropdownMenuTrigger className="w-full outline-none">
                <div className="flex items-center gap-3 px-3 py-2 rounded-lg cursor-pointer">
                    {companyLogo ? (
                        <div className="h-10 w-10 rounded-lg overflow-hidden flex items-center justify-center border border-border">
                            <Image
                                src={companyLogo}
                                alt={companyName}
                                width={40}
                                height={40}
                                className="object-cover"
                            />
                        </div>
                    ) : (
                        <div className="h-10 w-10 rounded-lg bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-white font-bold">
                            {companyName[0]}
                        </div>
                    )}
                    <div className="flex-1 text-left">
                        <p className="text-sm font-semibold text-text-primary line-clamp-1">
                            {companyName}
                        </p>
                        <p className="text-xs text-text-secondary line-clamp-1">
                            {tenantSlug}.myerp.com
                        </p>
                    </div>
                    <ChevronDown className={`h-4 w-4 text-text-secondary transition-transform ${open ? 'rotate-180' : ''}`} />
                </div>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start" className="w-56">
                <DropdownMenuItem onClick={() => router.push('/dashboard/settings/company')}>
                    <SlidersHorizontal className="mr-2 h-4 w-4" />
                    <span>Company Settings</span>
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => router.push('/dashboard/team')}>
                    <Users className="mr-2 h-4 w-4" />
                    <span>Team Management</span>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout} className="text-error">
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>Logout</span>
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
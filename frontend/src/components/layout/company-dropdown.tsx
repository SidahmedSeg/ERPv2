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
import { api, authApi } from '@/lib/api';
import Image from 'next/image';

export function CompanyDropdown() {
    const router = useRouter();
    const { user, tenant, logout, accessToken } = useAuthStore();
    const [open, setOpen] = useState(false);
    const [companyLogo, setCompanyLogo] = useState<string | null>(null);

    useEffect(() => {
        const fetchCompanyLogo = async () => {
            if (!accessToken) return;

            try {
                const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:18080/api';
                const response = await fetch(`${apiUrl}/settings/company`, {
                    headers: {
                        Authorization: `Bearer ${accessToken}`,
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
    }, [accessToken, open]);

    const handleLogout = async () => {
        try {
            await authApi.logout();
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            logout();
            router.push('/auth/login');
        }
    };

    // Convert slug to company name (capitalize first letter)
    const companyName = tenant?.slug
        ? tenant.slug.charAt(0).toUpperCase() + tenant.slug.slice(1)
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
                            {tenant?.slug}.myerp.com
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
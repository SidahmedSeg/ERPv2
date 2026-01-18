'use client';

import { useState } from 'react';
import { User, Moon, Sun, Globe, FileText, LogOut, Shield } from 'lucide-react';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
} from '@/components/ui/dropdown-menu';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { useAuthStore } from '@/store/auth-store';
import { useRouter } from 'next/navigation';
import { useTheme } from 'next-themes';
import { api } from '@/lib/api';
import { LANGUAGES } from '@/lib/constants';

export function UserDropdown() {
    const router = useRouter();
    const { user, clearAuth } = useAuthStore();
    const { theme, setTheme } = useTheme();
    const [language, setLanguage] = useState('en');

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

    return (
        <DropdownMenu>
            <DropdownMenuTrigger className="outline-none">
                <Avatar className="h-9 w-9 cursor-pointer ring-2 ring-transparent hover:ring-primary transition-all">
                    <AvatarImage
                        src={user?.avatar_url || undefined}
                        alt={`${user?.first_name} ${user?.last_name}`}
                    />
                    <AvatarFallback className="bg-gradient-to-br from-indigo-500 to-purple-600 text-white font-semibold">
                        {user?.first_name?.[0]}{user?.last_name?.[0]}
                    </AvatarFallback>
                </Avatar>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel>
                    <div className="flex flex-col space-y-1">
                        <p className="text-sm font-medium text-text-primary">{user?.first_name} {user?.last_name}</p>
                        <p className="text-xs text-text-secondary">{user?.email}</p>
                    </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />

                <DropdownMenuItem onClick={() => router.push('/dashboard/profile')}>
                    <User className="mr-2 h-4 w-4" />
                    <span>My Profile</span>
                </DropdownMenuItem>

                <DropdownMenuItem onClick={() => router.push('/dashboard/security')}>
                    <Shield className="mr-2 h-4 w-4" />
                    <span>Security</span>
                </DropdownMenuItem>

                <DropdownMenuSeparator />

                {/* Appearance submenu */}
                <DropdownMenuSub>
                    <DropdownMenuSubTrigger>
                        {theme === 'dark' ? (
                            <Moon className="mr-2 h-4 w-4" />
                        ) : (
                            <Sun className="mr-2 h-4 w-4" />
                        )}
                        <span>Appearance</span>
                    </DropdownMenuSubTrigger>
                    <DropdownMenuSubContent>
                        <DropdownMenuRadioGroup value={theme} onValueChange={setTheme}>
                            <DropdownMenuRadioItem value="light">
                                <Sun className="mr-2 h-4 w-4" />
                                Light
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value="dark">
                                <Moon className="mr-2 h-4 w-4" />
                                Dark
                            </DropdownMenuRadioItem>
                            <DropdownMenuRadioItem value="system">
                                <Globe className="mr-2 h-4 w-4" />
                                System
                            </DropdownMenuRadioItem>
                        </DropdownMenuRadioGroup>
                    </DropdownMenuSubContent>
                </DropdownMenuSub>

                {/* Language submenu */}
                <DropdownMenuSub>
                    <DropdownMenuSubTrigger>
                        <Globe className="mr-2 h-4 w-4" />
                        <span>Language</span>
                    </DropdownMenuSubTrigger>
                    <DropdownMenuSubContent>
                        <DropdownMenuRadioGroup value={language} onValueChange={setLanguage}>
                            {LANGUAGES.map((lang) => (
                                <DropdownMenuRadioItem key={lang.code} value={lang.code}>
                                    <span className="mr-2">{lang.flag}</span>
                                    {lang.name}
                                </DropdownMenuRadioItem>
                            ))}
                        </DropdownMenuRadioGroup>
                    </DropdownMenuSubContent>
                </DropdownMenuSub>

                <DropdownMenuItem onClick={() => window.open('/docs', '_blank')}>
                    <FileText className="mr-2 h-4 w-4" />
                    <span>Documentation</span>
                </DropdownMenuItem>

                <DropdownMenuSeparator />

                <DropdownMenuItem onClick={handleLogout} className="text-error hover:text-error">
                    <LogOut className="mr-2 h-4 w-4" />
                    <span>Logout</span>
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
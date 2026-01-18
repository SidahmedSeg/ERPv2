'use client';

import { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Menu, X, Settings } from 'lucide-react';
import { cn } from '@/lib/utils';
import { NAVIGATION_ITEMS } from '@/lib/constants';
import { CompanyDropdown } from './company-dropdown';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetTrigger } from '@/components/ui/sheet';
import { Separator } from '@/components/ui/separator';

export function MobileSidebar() {
    const pathname = usePathname();
    const [open, setOpen] = useState(false);

    return (
        <Sheet open={open} onOpenChange={setOpen}>
            <SheetTrigger asChild>
                <Button variant="ghost" size="icon" className="md:hidden">
                    <Menu className="h-6 w-6" />
                </Button>
            </SheetTrigger>
            <SheetContent side="left" className="w-64 p-0">
                <div className="flex flex-col h-full">
                    {/* Company Section */}
                    <div className="p-4">
                        <CompanyDropdown />
                    </div>

                    <Separator />

                    {/* Navigation */}
                    <nav className="flex-1 overflow-y-auto py-4 px-3">
                        <ul className="space-y-1">
                            {NAVIGATION_ITEMS.map((item) => {
                                const Icon = item.icon;
                                const isActive = pathname === item.href;

                                return (
                                    <li key={item.href}>
                                        <Link
                                            href={item.href}
                                            onClick={() => setOpen(false)}
                                            className={cn(
                                                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:bg-accent',
                                                isActive
                                                    ? 'bg-accent text-accent-foreground'
                                                    : 'text-muted-foreground hover:text-foreground'
                                            )}
                                        >
                                            <Icon className="h-5 w-5" />
                                            <span>{item.title}</span>
                                        </Link>
                                    </li>
                                );
                            })}
                        </ul>
                    </nav>

                    {/* Settings at Bottom */}
                    <div className="p-3 border-t">
                        <Link
                            href="/dashboard/settings"
                            onClick={() => setOpen(false)}
                            className={cn(
                                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all hover:bg-accent',
                                pathname.startsWith('/dashboard/settings')
                                    ? 'bg-accent text-accent-foreground'
                                    : 'text-muted-foreground hover:text-foreground'
                            )}
                        >
                            <Settings className="h-5 w-5" />
                            <span>Settings</span>
                        </Link>
                    </div>
                </div>
            </SheetContent>
        </Sheet>
    );
}
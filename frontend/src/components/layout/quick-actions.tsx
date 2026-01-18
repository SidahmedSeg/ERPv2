'use client';

import { useState, useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { Plus, Package, Users, Building2, TrendingUp, UserPlus, FileCheck, Receipt, ReceiptText } from 'lucide-react';
import '@/Loader/Redirect_loader.css';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { useAuthStore } from '@/store/auth-store';

export function QuickActions() {
    const router = useRouter();
    const pathname = usePathname();
    const { user } = useAuthStore();
    const [showAddProductDialog, setShowAddProductDialog] = useState(false);
    const [showAddCustomerDialog, setShowAddCustomerDialog] = useState(false);
    const [showAddSupplierDialog, setShowAddSupplierDialog] = useState(false);
    const [showStockMovementDialog, setShowStockMovementDialog] = useState(false);
    const [showInviteMemberDialog, setShowInviteMemberDialog] = useState(false);
    const [navigating, setNavigating] = useState(false);

    // Reset navigating state when pathname changes (route navigation completes)
    useEffect(() => {
        setNavigating(false);
    }, [pathname]);

    // Fallback: Auto-hide loader after 2 seconds if pathname doesn't change
    useEffect(() => {
        if (navigating) {
            const timeout = setTimeout(() => {
                setNavigating(false);
            }, 2000);
            return () => clearTimeout(timeout);
        }
    }, [navigating]);

    const handleCreateProduct = () => {
        setNavigating(true);
        router.push('/dashboard/products/new');
    };

    const handleCreateCustomer = () => {
        setNavigating(true);
        router.push('/dashboard/customers/new');
    };

    const handleCreateSupplier = () => {
        setNavigating(true);
        router.push('/dashboard/suppliers?action=add');
    };

    const handleAddStockMovement = () => {
        setNavigating(true);
        router.push('/dashboard/stock/movements?action=add');
    };

    const handleInviteMember = () => {
        setNavigating(true);
        router.push('/dashboard/team?action=invite');
    };

    const handleCreateEstimate = () => {
        setNavigating(true);
        router.push('/dashboard/estimates/new');
    };

    const handleCreateInvoice = () => {
        setNavigating(true);
        router.push('/dashboard/invoices/new');
    };

    const handleCreatePurchaseOrder = () => {
        setNavigating(true);
        router.push('/dashboard/purchase-orders/new');
    };

    // Check if user is admin
    const isAdmin = user?.roles?.some(role => role.name === 'admin' || role.name === 'superadmin') || false;

    return (
        <>
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="relative bg-gray-50 border border-gray-200 hover:bg-gray-200 h-8 w-auto px-3 gap-2">
                    <span className="text-sm font-medium">Add</span>
                    <Plus className="h-4 w-4" />
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
                <DropdownMenuLabel>Quick Actions</DropdownMenuLabel>
                <DropdownMenuSeparator />

                <DropdownMenuItem onClick={handleCreateProduct} className="cursor-pointer">
                    <Package className="mr-2 h-4 w-4" />
                    Create new product
                </DropdownMenuItem>

                <DropdownMenuItem onClick={handleCreateCustomer} className="cursor-pointer">
                    <Users className="mr-2 h-4 w-4" />
                    Create new customer
                </DropdownMenuItem>

                <DropdownMenuItem onClick={handleCreateSupplier} className="cursor-pointer">
                    <Building2 className="mr-2 h-4 w-4" />
                    Create new supplier
                </DropdownMenuItem>

                <DropdownMenuItem onClick={handleAddStockMovement} className="cursor-pointer">
                    <TrendingUp className="mr-2 h-4 w-4" />
                    Add stock movement
                </DropdownMenuItem>

                <DropdownMenuSeparator />

                <DropdownMenuItem onClick={handleCreateEstimate} className="cursor-pointer">
                    <FileCheck className="mr-2 h-4 w-4" />
                    Create new estimate
                </DropdownMenuItem>

                <DropdownMenuItem onClick={handleCreateInvoice} className="cursor-pointer">
                    <Receipt className="mr-2 h-4 w-4" />
                    Create new invoice
                </DropdownMenuItem>

                <DropdownMenuItem onClick={handleCreatePurchaseOrder} className="cursor-pointer">
                    <ReceiptText className="mr-2 h-4 w-4" />
                    Create new purchase order
                </DropdownMenuItem>

                {isAdmin && (
                    <>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={handleInviteMember} className="cursor-pointer">
                            <UserPlus className="mr-2 h-4 w-4" />
                            Invite new member
                        </DropdownMenuItem>
                    </>
                )}
            </DropdownMenuContent>
        </DropdownMenu>

        {/* Navigation Loader Overlay */}
        {navigating && (
            <div className="fixed inset-0 bg-black/50 backdrop-blur-sm z-[9999] flex items-center justify-center">
                <div className="flex flex-col items-center gap-4">
                    <div className="loader"></div>
                    <p className="text-sm text-white font-medium">Redirecting...</p>
                </div>
            </div>
        )}
        </>
    );
}

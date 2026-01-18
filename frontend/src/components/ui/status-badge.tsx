'use client';

import { CheckCircle2, XCircle, AlertTriangle, Clock, Package, TrendingUp, Ban, RotateCcw, AlertOctagon } from 'lucide-react';

interface StatusBadgeProps {
    status: string;
    type?: 'product' | 'customer' | 'serial_number' | 'order' | 'invoice' | 'default';
    showIcon?: boolean;
    showText?: boolean;
    className?: string;
}

export function StatusBadge({
    status,
    type = 'default',
    showIcon = true,
    showText = true,
    className = ''
}: StatusBadgeProps) {
    const getStatusConfig = () => {
        const lowerStatus = status.toLowerCase().replace('_', ' ');

        // Product statuses
        if (type === 'product') {
            switch (lowerStatus) {
                case 'active':
                    return {
                        icon: CheckCircle2,
                        color: 'text-green-600',
                        bgColor: 'bg-green-100',
                        label: 'Active'
                    };
                case 'discontinued':
                    return {
                        icon: XCircle,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Discontinued'
                    };
                case 'out of stock':
                case 'out_of_stock':
                    return {
                        icon: AlertTriangle,
                        color: 'text-orange-600',
                        bgColor: 'bg-orange-100',
                        label: 'Out of Stock'
                    };
                default:
                    return {
                        icon: Package,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: formatStatus(status)
                    };
            }
        }

        // Customer statuses
        if (type === 'customer') {
            switch (lowerStatus) {
                case 'active':
                    return {
                        icon: CheckCircle2,
                        color: 'text-green-600',
                        bgColor: 'bg-green-100',
                        label: 'Active'
                    };
                case 'inactive':
                    return {
                        icon: XCircle,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: 'Inactive'
                    };
                case 'blacklist':
                    return {
                        icon: Ban,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Blacklisted'
                    };
                default:
                    return {
                        icon: AlertTriangle,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: formatStatus(status)
                    };
            }
        }

        // Serial number statuses
        if (type === 'serial_number') {
            switch (lowerStatus) {
                case 'in stock':
                case 'in_stock':
                    return {
                        icon: Package,
                        color: 'text-green-600',
                        bgColor: 'bg-green-100',
                        label: 'In Stock'
                    };
                case 'reserved':
                    return {
                        icon: Clock,
                        color: 'text-blue-600',
                        bgColor: 'bg-blue-100',
                        label: 'Reserved'
                    };
                case 'sold':
                    return {
                        icon: CheckCircle2,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: 'Sold'
                    };
                case 'returned':
                    return {
                        icon: RotateCcw,
                        color: 'text-orange-600',
                        bgColor: 'bg-orange-100',
                        label: 'Returned'
                    };
                case 'defective':
                    return {
                        icon: AlertOctagon,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Defective'
                    };
                case 'expired':
                    return {
                        icon: XCircle,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Expired'
                    };
                case 'lost':
                    return {
                        icon: AlertTriangle,
                        color: 'text-orange-600',
                        bgColor: 'bg-orange-100',
                        label: 'Lost'
                    };
                case 'transferred':
                    return {
                        icon: TrendingUp,
                        color: 'text-purple-600',
                        bgColor: 'bg-purple-100',
                        label: 'Transferred'
                    };
                default:
                    return {
                        icon: Package,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: formatStatus(status)
                    };
            }
        }

        // Order statuses
        if (type === 'order') {
            switch (lowerStatus) {
                case 'pending':
                    return {
                        icon: Clock,
                        color: 'text-yellow-600',
                        bgColor: 'bg-yellow-100',
                        label: 'Pending'
                    };
                case 'confirmed':
                    return {
                        icon: CheckCircle2,
                        color: 'text-blue-600',
                        bgColor: 'bg-blue-100',
                        label: 'Confirmed'
                    };
                case 'delivered':
                    return {
                        icon: CheckCircle2,
                        color: 'text-green-600',
                        bgColor: 'bg-green-100',
                        label: 'Delivered'
                    };
                case 'cancelled':
                    return {
                        icon: XCircle,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Cancelled'
                    };
                default:
                    return {
                        icon: Package,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: formatStatus(status)
                    };
            }
        }

        // Invoice statuses
        if (type === 'invoice') {
            switch (lowerStatus) {
                case 'draft':
                    return {
                        icon: Clock,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: 'Draft'
                    };
                case 'sent':
                    return {
                        icon: TrendingUp,
                        color: 'text-blue-600',
                        bgColor: 'bg-blue-100',
                        label: 'Sent'
                    };
                case 'paid':
                    return {
                        icon: CheckCircle2,
                        color: 'text-green-600',
                        bgColor: 'bg-green-100',
                        label: 'Paid'
                    };
                case 'overdue':
                    return {
                        icon: AlertTriangle,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Overdue'
                    };
                case 'cancelled':
                    return {
                        icon: XCircle,
                        color: 'text-red-600',
                        bgColor: 'bg-red-100',
                        label: 'Cancelled'
                    };
                default:
                    return {
                        icon: Package,
                        color: 'text-gray-600',
                        bgColor: 'bg-gray-100',
                        label: formatStatus(status)
                    };
            }
        }

        // Default (generic) statuses
        switch (lowerStatus) {
            case 'active':
                return {
                    icon: CheckCircle2,
                    color: 'text-green-600',
                    bgColor: 'bg-green-100',
                    label: 'Active'
                };
            case 'inactive':
                return {
                    icon: XCircle,
                    color: 'text-gray-600',
                    bgColor: 'bg-gray-100',
                    label: 'Inactive'
                };
            case 'pending':
                return {
                    icon: Clock,
                    color: 'text-yellow-600',
                    bgColor: 'bg-yellow-100',
                    label: 'Pending'
                };
            case 'completed':
                return {
                    icon: CheckCircle2,
                    color: 'text-green-600',
                    bgColor: 'bg-green-100',
                    label: 'Completed'
                };
            default:
                return {
                    icon: Package,
                    color: 'text-gray-600',
                    bgColor: 'bg-gray-100',
                    label: formatStatus(status)
                };
        }
    };

    const formatStatus = (status: string) => {
        return status
            .replace(/_/g, ' ')
            .split(' ')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
            .join(' ');
    };

    const config = getStatusConfig();
    const Icon = config.icon;

    return (
        <div className={`flex items-center gap-2 ${className}`}>
            {showIcon && <Icon className={`h-4 w-4 ${config.color}`} />}
            {showText && (
                <span className={`text-sm font-medium ${config.color}`}>
                    {config.label}
                </span>
            )}
        </div>
    );
}

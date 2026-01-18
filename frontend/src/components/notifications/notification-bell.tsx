'use client';

import { useState, useEffect, useRef } from 'react';
import { Bell, Download, X, CheckCheck, FileText, CheckCircle } from 'lucide-react';
import { useNotifications } from '@/contexts/NotificationContext';
import { useRouter } from 'next/navigation';
import { toast } from 'sonner';

export function NotificationBell() {
    const { notifications, unreadCount, clearNotifications, isConnected } = useNotifications();
    const [isOpen, setIsOpen] = useState(false);
    const dropdownRef = useRef<HTMLDivElement>(null);
    const router = useRouter();

    // Close dropdown when clicking outside
    useEffect(() => {
        function handleClickOutside(event: MouseEvent) {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        }

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
            return () => document.removeEventListener('mousedown', handleClickOutside);
        }
    }, [isOpen]);

    const handleApprovePO = async (e: React.MouseEvent, poId: string) => {
        e.stopPropagation();

        try {
            const token = localStorage.getItem('token');
            const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/api/purchase-orders/${poId}/approve`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ notes: '' }),
            });

            if (!response.ok) {
                throw new Error('Failed to approve purchase order');
            }

            toast.success('Purchase order approved successfully');
            setIsOpen(false);
        } catch (error) {
            console.error('Error approving PO:', error);
            toast.error('Failed to approve purchase order');
        }
    };

    const handleNotificationClick = (notification: any) => {
        if (notification.type === 'export_completed' && notification.data?.job_id) {
            // Navigate to download
            const hostname = window.location.hostname;
            const subdomain = hostname.split('.')[0];
            const token = localStorage.getItem('token');

            const downloadUrl = `http://${subdomain}.myerp.local:8080/api/export/jobs/${notification.data.job_id}/download`;

            // Create temporary link and trigger download
            const a = document.createElement('a');
            a.href = downloadUrl;
            a.download = `customers_export_${new Date().toISOString().split('T')[0]}.xlsx`;

            // Add authorization header via fetch
            fetch(downloadUrl, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            })
                .then((response) => response.blob())
                .then((blob) => {
                    const url = window.URL.createObjectURL(blob);
                    a.href = url;
                    document.body.appendChild(a);
                    a.click();
                    window.URL.revokeObjectURL(url);
                    document.body.removeChild(a);
                })
                .catch((error) => {
                    console.error('Download failed:', error);
                    alert('Failed to download file');
                });

            setIsOpen(false);
        }
    };

    const formatTime = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffMins = Math.floor(diffMs / 60000);

        if (diffMins < 1) return 'Just now';
        if (diffMins < 60) return `${diffMins}m ago`;
        const diffHours = Math.floor(diffMins / 60);
        if (diffHours < 24) return `${diffHours}h ago`;
        const diffDays = Math.floor(diffHours / 24);
        return `${diffDays}d ago`;
    };

    const getNotificationIcon = (type: string) => {
        switch (type) {
            case 'export_completed':
                return <Download className="h-4 w-4 text-green-500" />;
            case 'export_failed':
                return <X className="h-4 w-4 text-red-500" />;
            case 'export_processing':
                return <div className="h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />;
            case 'po_approval_request':
                return <FileText className="h-4 w-4 text-orange-500" />;
            default:
                return <Bell className="h-4 w-4 text-blue-500" />;
        }
    };

    return (
        <div className="relative" ref={dropdownRef}>
            {/* Bell Button */}
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="relative p-2 text-text-secondary hover:text-text-primary hover:bg-gray-100 rounded-lg transition-colors"
            >
                <Bell className="h-5 w-5" />

                {/* Badge */}
                {unreadCount > 0 && (
                    <span className="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-primary text-[10px] font-bold text-white">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
                )}

                {/* Connection indicator */}
                {isConnected && (
                    <span className="absolute bottom-1 right-1 h-2 w-2 rounded-full bg-green-500 border border-white" />
                )}
            </button>

            {/* Dropdown */}
            {isOpen && (
                <div className="absolute right-0 mt-2 w-96 bg-white rounded-lg shadow-lg border border-border z-50 max-h-[600px] flex flex-col">
                    {/* Header */}
                    <div className="flex items-center justify-between p-4 border-b border-border">
                        <h3 className="font-semibold text-text-primary">Notifications</h3>
                        {notifications.length > 0 && (
                            <button
                                onClick={() => clearNotifications()}
                                className="text-xs text-text-secondary hover:text-text-primary flex items-center gap-1"
                            >
                                <CheckCheck className="h-3 w-3" />
                                Clear all
                            </button>
                        )}
                    </div>

                    {/* Notifications List */}
                    <div className="overflow-y-auto flex-1">
                        {notifications.length === 0 ? (
                            <div className="p-8 text-center">
                                <Bell className="h-12 w-12 text-gray-300 mx-auto mb-3" />
                                <p className="text-sm text-text-secondary">No notifications yet</p>
                            </div>
                        ) : (
                            <div className="divide-y divide-border">
                                {notifications.slice(0, 10).map((notification) => (
                                    <div
                                        key={notification.id}
                                        onClick={() => handleNotificationClick(notification)}
                                        className={`
                      p-4 hover:bg-gray-50 transition-colors cursor-pointer
                      ${!notification.read ? 'bg-blue-50' : ''}
                    `}
                                    >
                                        <div className="flex items-start gap-3">
                                            <div className="flex-shrink-0 mt-1">
                                                {getNotificationIcon(notification.type)}
                                            </div>

                                            <div className="flex-1 min-w-0">
                                                <div className="flex items-start justify-between gap-2">
                                                    <h4 className="text-sm font-medium text-text-primary">
                                                        {notification.title}
                                                    </h4>
                                                    <span className="text-xs text-text-secondary whitespace-nowrap">
                            {formatTime(notification.created_at)}
                          </span>
                                                </div>

                                                <p className="text-sm text-text-secondary mt-1">
                                                    {notification.message}
                                                </p>

                                                {/* Progress bar for processing */}
                                                {notification.type === 'export_processing' && notification.data?.progress && (
                                                    <div className="mt-2">
                                                        <div className="flex items-center justify-between text-xs text-text-secondary mb-1">
                                                            <span>Progress</span>
                                                            <span>{notification.data.progress}%</span>
                                                        </div>
                                                        <div className="w-full bg-gray-200 rounded-full h-1.5">
                                                            <div
                                                                className="bg-blue-500 h-1.5 rounded-full transition-all duration-300"
                                                                style={{ width: `${notification.data.progress}%` }}
                                                            />
                                                        </div>
                                                    </div>
                                                )}

                                                {/* Download button for completed exports */}
                                                {notification.type === 'export_completed' && (
                                                    <button className="mt-2 text-xs text-primary hover:text-primary-700 font-medium flex items-center gap-1">
                                                        <Download className="h-3 w-3" />
                                                        Download file
                                                    </button>
                                                )}

                                                {/* Action buttons for PO approval requests */}
                                                {notification.type === 'po_approval_request' && (notification as any).actions && (
                                                    <div className="mt-3 flex gap-2">
                                                        {(notification as any).actions.map((action: any, index: number) => {
                                                            if (action.type === 'approve') {
                                                                return (
                                                                    <button
                                                                        key={index}
                                                                        onClick={(e) => handleApprovePO(e, action.po_id)}
                                                                        className="px-3 py-1.5 text-xs font-medium text-white bg-green-600 hover:bg-green-700 rounded transition-colors flex items-center gap-1"
                                                                    >
                                                                        <CheckCircle className="h-3 w-3" />
                                                                        {action.label}
                                                                    </button>
                                                                );
                                                            } else if (action.type === 'view') {
                                                                return (
                                                                    <button
                                                                        key={index}
                                                                        onClick={(e) => {
                                                                            e.stopPropagation();
                                                                            router.push(action.url);
                                                                            setIsOpen(false);
                                                                        }}
                                                                        className="px-3 py-1.5 text-xs font-medium text-primary border border-primary hover:bg-primary hover:text-white rounded transition-colors flex items-center gap-1"
                                                                    >
                                                                        <FileText className="h-3 w-3" />
                                                                        {action.label}
                                                                    </button>
                                                                );
                                                            }
                                                            return null;
                                                        })}
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    {/* Footer */}
                    {notifications.length > 10 && (
                        <div className="p-3 border-t border-border text-center">
                            <button className="text-sm text-primary hover:text-primary-700 font-medium">
                                View all notifications
                            </button>
                        </div>
                    )}
                </div>
            )}
        </div>
    );
}
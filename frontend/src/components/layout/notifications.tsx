'use client';

import { useState, useEffect, useRef } from 'react';
import { Bell, Download, X, CheckCheck } from 'lucide-react';
import { useNotifications } from '@/contexts/NotificationContext';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';

export function Notifications() {
    const { notifications, unreadCount, clearNotifications, isConnected } = useNotifications();

    const handleNotificationClick = (notification: any) => {
        if (notification.type === 'export_completed' && notification.data?.job_id) {
            // Navigate to download
            const hostname = window.location.hostname;
            const subdomain = hostname.split('.')[0];
            const token = localStorage.getItem('token');

            const downloadUrl = `http://${subdomain}.myerp.local:8080/api/export/jobs/${notification.data.job_id}/download`;

            // Download file
            fetch(downloadUrl, {
                headers: {
                    Authorization: `Bearer ${token}`,
                },
            })
                .then((response) => response.blob())
                .then((blob) => {
                    const url = window.URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `customers_export_${new Date().toISOString().split('T')[0]}.xlsx`;
                    document.body.appendChild(a);
                    a.click();
                    window.URL.revokeObjectURL(url);
                    document.body.removeChild(a);
                })
                .catch((error) => {
                    console.error('Download failed:', error);
                    alert('Failed to download file');
                });
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
            default:
                return <Bell className="h-4 w-4 text-blue-500" />;
        }
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="relative bg-gray-50 border border-gray-200 hover:bg-gray-200 h-8 w-8">
                    <Bell className="h-4 w-4" />
                    {unreadCount > 0 && (
                        <Badge
                            className="absolute -top-0.5 -right-0.5 h-4 w-4 flex items-center justify-center p-0 text-[10px] bg-primary text-white"
                        >
                            {unreadCount > 9 ? '9+' : unreadCount}
                        </Badge>
                    )}
                    {/* Connection indicator */}
                    {isConnected && (
                        <span className="absolute bottom-0 right-0 h-2 w-2 rounded-full bg-green-500 border border-white" />
                    )}
                </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-96 max-h-[600px] overflow-y-auto">
                <div className="flex items-center justify-between px-2 py-1.5">
                    <DropdownMenuLabel>Notifications</DropdownMenuLabel>
                    {notifications.length > 0 && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                clearNotifications();
                            }}
                            className="text-xs text-muted-foreground hover:text-foreground flex items-center gap-1"
                        >
                            <CheckCheck className="h-3 w-3" />
                            Clear all
                        </button>
                    )}
                </div>
                <DropdownMenuSeparator />

                {notifications.length === 0 ? (
                    <div className="p-8 text-center">
                        <Bell className="h-12 w-12 text-muted-foreground/30 mx-auto mb-3" />
                        <p className="text-sm text-muted-foreground">No notifications yet</p>
                        <p className="text-xs text-muted-foreground mt-1">
                            {isConnected ? 'Connected - waiting for updates' : 'Connecting...'}
                        </p>
                    </div>
                ) : (
                    notifications.slice(0, 10).map((notification) => (
                        <DropdownMenuItem
                            key={notification.id}
                            onClick={() => handleNotificationClick(notification)}
                            className={`p-3 cursor-pointer ${!notification.read ? 'bg-blue-50' : ''}`}
                        >
                            <div className="flex items-start gap-3 w-full">
                                <div className="flex-shrink-0 mt-1">
                                    {getNotificationIcon(notification.type)}
                                </div>

                                <div className="flex-1 min-w-0">
                                    <div className="flex items-start justify-between gap-2">
                                        <p className="text-sm font-medium">{notification.title}</p>
                                        <span className="text-xs text-muted-foreground whitespace-nowrap">
                                            {formatTime(notification.created_at)}
                                        </span>
                                    </div>

                                    <p className="text-xs text-muted-foreground mt-1">
                                        {notification.message}
                                    </p>

                                    {/* Progress bar for processing */}
                                    {notification.type === 'export_processing' && notification.data?.progress && (
                                        <div className="mt-2">
                                            <div className="flex items-center justify-between text-xs text-muted-foreground mb-1">
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
                                        <button className="mt-2 text-xs text-primary hover:text-primary/80 font-medium flex items-center gap-1">
                                            <Download className="h-3 w-3" />
                                            Download file
                                        </button>
                                    )}
                                </div>
                            </div>
                        </DropdownMenuItem>
                    ))
                )}

                {notifications.length > 10 && (
                    <>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem className="text-center justify-center text-sm text-primary cursor-pointer">
                            View all notifications
                        </DropdownMenuItem>
                    </>
                )}
            </DropdownMenuContent>
        </DropdownMenu>
    );
}
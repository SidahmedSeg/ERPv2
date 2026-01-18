'use client';

import { useEffect, useState, useRef } from 'react';
import { Toast } from './toast';
import { useNotifications } from '@/contexts/NotificationContext';

interface ToastData {
    id: string;
    type: 'success' | 'error' | 'info';
    title: string;
    message: string;
}

export function ToastContainer() {
    const { notifications } = useNotifications();
    const [toasts, setToasts] = useState<ToastData[]>([]);
    const processedIdsRef = useRef<Set<string>>(new Set());

    useEffect(() => {
        // Only show toasts for PO approval requests
        const newNotifications = notifications.filter(
            (notification) =>
                notification.type === 'po_approval_request' &&
                !processedIdsRef.current.has(notification.id)
        );

        newNotifications.forEach((notification) => {
            // Mark as processed
            processedIdsRef.current.add(notification.id);

            setToasts((prev) => [
                ...prev,
                {
                    id: notification.id,
                    type: 'info',
                    title: notification.title,
                    message: notification.message,
                },
            ]);
        });
    }, [notifications]);

    const handleClose = (id: string) => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
    };

    return (
        <div className="fixed top-4 right-4 z-50 flex flex-col gap-2">
            {toasts.map((toast) => (
                <Toast
                    key={toast.id}
                    id={toast.id}
                    type={toast.type}
                    title={toast.title}
                    message={toast.message}
                    onClose={handleClose}
                />
            ))}
        </div>
    );
}
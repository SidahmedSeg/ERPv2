'use client';

import React, { createContext, useContext, ReactNode, useEffect, useState } from 'react';
import { useWebSocket } from '@/hooks/useWebSocket';
import Cookies from 'js-cookie';

interface WebSocketMessage {
    id: string;
    type: string;
    title: string;
    message: string;
    data?: any;
    created_at: string;
    read: boolean;
}

interface NotificationContextType {
    isConnected: boolean;
    notifications: WebSocketMessage[];
    clearNotifications: () => void;
    unreadCount: number;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

interface NotificationProviderProps {
    children: ReactNode;
}

export function NotificationProvider({ children }: NotificationProviderProps) {
    const [token, setToken] = useState<string | null>(null);

    useEffect(() => {
        // Get token from cookies (where auth store saves it)
        const storedToken = Cookies.get('auth_token');

        console.log('📍 NotificationProvider: Token loaded:', storedToken ? 'Yes ✅' : 'No ❌');
        if (storedToken) {
            console.log('📍 Token preview:', storedToken.substring(0, 30) + '...');
        }
        setToken(storedToken || null);
    }, []);

    const { isConnected, notifications, clearNotifications } = useWebSocket(token);

    const unreadCount = notifications.filter(n => !n.read).length;

    return (
        <NotificationContext.Provider
            value={{
                isConnected,
                notifications,
                clearNotifications,
                unreadCount,
            }}
        >
            {children}
        </NotificationContext.Provider>
    );
}

export function useNotifications() {
    const context = useContext(NotificationContext);
    if (context === undefined) {
        throw new Error('useNotifications must be used within a NotificationProvider');
    }
    return context;
}
import { useEffect, useRef, useState, useCallback } from 'react';

interface WebSocketMessage {
    id: string;
    type: string;
    title: string;
    message: string;
    data?: any;
    actions?: any[];
    created_at: string;
    read: boolean;
}

interface UseWebSocketReturn {
    isConnected: boolean;
    notifications: WebSocketMessage[];
    clearNotifications: () => void;
}

export function useWebSocket(token: string | null): UseWebSocketReturn {
    const [isConnected, setIsConnected] = useState(false);
    const [notifications, setNotifications] = useState<WebSocketMessage[]>(() => {
        // Load notifications from localStorage on mount
        if (typeof window !== 'undefined') {
            try {
                const stored = localStorage.getItem('notifications');
                return stored ? JSON.parse(stored) : [];
            } catch (error) {
                console.error('Failed to load notifications from localStorage:', error);
                return [];
            }
        }
        return [];
    });
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
    const reconnectAttemptsRef = useRef(0);

    const connect = useCallback(() => {
        if (!token) {
            console.log('🔌 No token, skipping WebSocket connection');
            return;
        }

        // Get subdomain
        const hostname = window.location.hostname;
        const subdomain = hostname.split('.')[0];

        // WebSocket URL
        const wsUrl = `ws://${subdomain}.myerp.local:8080/api/ws/notifications?token=${encodeURIComponent(token)}`;

        console.log('🔌 Connecting to WebSocket:', wsUrl);
        console.log('🔑 Token present:', token ? 'Yes' : 'No');

        try {
            const ws = new WebSocket(wsUrl);
            wsRef.current = ws;

            ws.onopen = () => {
                console.log('✅ WebSocket connected');
                setIsConnected(true);
                reconnectAttemptsRef.current = 0;

                // Send authentication
                ws.send(JSON.stringify({
                    type: 'auth',
                    token: token,
                }));
            };

            ws.onmessage = (event) => {
                try {
                    const notification: WebSocketMessage = JSON.parse(event.data);
                    console.log('📨 Notification received:', notification);

                    setNotifications((prev) => {
                        const updated = [notification, ...prev];
                        // Save to localStorage
                        if (typeof window !== 'undefined') {
                            localStorage.setItem('notifications', JSON.stringify(updated));
                        }
                        return updated;
                    });

                    // Show browser notification if permitted
                    if (Notification.permission === 'granted') {
                        new Notification(notification.title, {
                            body: notification.message,
                            icon: '/favicon.ico',
                        });
                    }
                } catch (error) {
                    console.error('Failed to parse notification:', error);
                }
            };

            ws.onerror = (error) => {
                console.error('❌ WebSocket error:', error);
            };

            ws.onclose = () => {
                console.log('🔌 WebSocket disconnected');
                setIsConnected(false);
                wsRef.current = null;

                // Reconnect with exponential backoff
                if (token && reconnectAttemptsRef.current < 10) {
                    const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
                    console.log(`🔄 Reconnecting in ${delay}ms...`);

                    reconnectTimeoutRef.current = setTimeout(() => {
                        reconnectAttemptsRef.current++;
                        connect();
                    }, delay);
                }
            };
        } catch (error) {
            console.error('Failed to create WebSocket:', error);
        }
    }, [token]);

    useEffect(() => {
        connect();

        // Request notification permission
        if (Notification.permission === 'default') {
            Notification.requestPermission();
        }

        return () => {
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }
            if (wsRef.current) {
                wsRef.current.close();
            }
        };
    }, [connect]);

    const clearNotifications = useCallback(() => {
        setNotifications([]);
        // Clear from localStorage
        if (typeof window !== 'undefined') {
            localStorage.removeItem('notifications');
        }
    }, []);

    return {
        isConnected,
        notifications,
        clearNotifications,
    };
}
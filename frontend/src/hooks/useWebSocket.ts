import { useRef, useCallback } from 'react';
import { useAuthStore } from '@/store/authStore';

interface UseWebSocketOptions {
  onMessage?: (data: unknown) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
  onError?: (error: Event) => void;
}

export function useWebSocket(channel: string, options: UseWebSocketOptions = {}) {
  const wsRef = useRef<WebSocket | null>(null);
  const { token } = useAuthStore();
  const { onMessage, onConnect, onDisconnect, onError } = options;

  const connect = useCallback(() => {
    if (!token) {
      console.error('No auth token available');
      return;
    }

    const wsUrl = `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/ws/chat?token=${token}&channel=${channel}`;
    
    try {
      wsRef.current = new WebSocket(wsUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connected to channel:', channel);
        onConnect?.();
      };

      wsRef.current.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          onMessage?.(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      wsRef.current.onclose = () => {
        console.log('WebSocket disconnected from channel:', channel);
        onDisconnect?.();
      };

      wsRef.current.onerror = (error) => {
        console.error('WebSocket error occurred');
        console.error('Error details:', {
          type: error.type || 'unknown',
          readyState: wsRef.current?.readyState,
          url: wsUrl,
          timestamp: new Date().toISOString(),
          errorEvent: error.constructor.name
        });
        
        // WebSocketの状態も確認
        const readyStateNames = ['CONNECTING', 'OPEN', 'CLOSING', 'CLOSED'];
        console.error('WebSocket state:', readyStateNames[wsRef.current?.readyState || 0]);
        
        onError?.(error);
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
    }
  }, [token, channel, onMessage, onConnect, onDisconnect, onError]);

  const disconnect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const sendMessage = useCallback((data: unknown) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data));
      return true;
    }
    return false;
  }, []);

  const isConnected = wsRef.current?.readyState === WebSocket.OPEN;

  return {
    connect,
    disconnect,
    sendMessage,
    isConnected,
  };
}
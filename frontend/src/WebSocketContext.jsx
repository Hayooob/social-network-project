import React, { createContext, useContext, useEffect, useRef, useState, useCallback } from 'react';
import { useAuth } from './VerifyAuth';

const WebSocketContext = createContext(null);

export function WebSocketProvider({ children }) {
  const { user } = useAuth();
  const wsRef = useRef(null);
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState(null);
  const [badgeRefreshTrigger, setBadgeRefreshTrigger] = useState(0);
  const reconnectTimeoutRef = useRef(null);
  const reconnectAttemptsRef = useRef(0);
  const maxReconnectAttempts = 5;
  const mountedRef = useRef(true);

  // Call this to trigger badge refresh in Layout
  const refreshBadges = useCallback(() => {
    setBadgeRefreshTrigger(prev => prev + 1);
  }, []);

  const connect = useCallback(() => {
    if (!user || wsRef.current?.readyState === WebSocket.OPEN) {
      return;
    }

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    console.log('WebSocket: Connecting...');
    const ws = new WebSocket('ws://localhost:8080/ws');
    wsRef.current = ws;

    ws.onopen = () => {
      if (!mountedRef.current) return;
      console.log('WebSocket: Connected');
      setIsConnected(true);
      reconnectAttemptsRef.current = 0;
    };

    ws.onmessage = (event) => {
      if (!mountedRef.current) return;
      try {
        const data = JSON.parse(event.data);
        console.log('WebSocket: Message received', data);
        setLastMessage(data);
        
        // Auto-refresh badges when new message or notification arrives
        if (data.type === 'chat_message' || data.type === 'notification') {
          refreshBadges();
        }
      } catch (err) {
        console.error('WebSocket: Failed to parse message', err);
      }
    };

    ws.onerror = (error) => {
      console.error('WebSocket: Error', error);
    };

    ws.onclose = (event) => {
      if (!mountedRef.current) return;
      console.log('WebSocket: Disconnected', event.code, event.reason);
      setIsConnected(false);
      wsRef.current = null;

      if (user && reconnectAttemptsRef.current < maxReconnectAttempts) {
        const delay = Math.min(1000 * Math.pow(2, reconnectAttemptsRef.current), 30000);
        console.log(`WebSocket: Reconnecting in ${delay}ms (attempt ${reconnectAttemptsRef.current + 1})`);
        reconnectTimeoutRef.current = setTimeout(() => {
          reconnectAttemptsRef.current++;
          connect();
        }, delay);
      }
    };
  }, [user, refreshBadges]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    setIsConnected(false);
  }, []);

  const sendMessage = useCallback((message) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      console.log('WebSocket: Sending', message);
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    console.warn('WebSocket: Cannot send - not connected');
    return false;
  }, []);

  useEffect(() => {
    mountedRef.current = true;

    if (user) {
      const timer = setTimeout(connect, 100);
      return () => clearTimeout(timer);
    } else {
      disconnect();
    }

    return () => {
      mountedRef.current = false;
    };
  }, [user, connect, disconnect]);

  useEffect(() => {
    return () => {
      mountedRef.current = false;
      disconnect();
    };
  }, [disconnect]);

  const value = {
    isConnected,
    lastMessage,
    sendMessage,
    reconnect: connect,
    refreshBadges,
    badgeRefreshTrigger
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}

export function useWebSocket() {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
}
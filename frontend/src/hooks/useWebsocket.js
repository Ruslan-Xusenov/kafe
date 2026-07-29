import { useEffect, useRef, useCallback } from 'react';
import { useAuthStore } from '../store/authStore';

export const useWebsocket = (onMessage) => {
  const socketRef = useRef(null);
  const onMessageRef = useRef(onMessage);
  const { token, isAuthenticated } = useAuthStore();

  // Keep the callback ref up to date without triggering reconnects
  useEffect(() => {
    onMessageRef.current = onMessage;
  });

  const getSocket = useCallback(() => socketRef.current, []);

  useEffect(() => {
    if (!isAuthenticated || !token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${protocol}//${window.location.host}/api/ws`, [`auth.${token}`]);

    socket.onopen = () => {
      console.log('WS Shared Connection Opened');
    };

    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (onMessageRef.current) onMessageRef.current(data);
      } catch {
        console.error('WS: failed to parse message');
      }
    };

    socket.onclose = () => {
      console.log('WS Connection Closed');
    };

    socketRef.current = socket;

    return () => {
      socket.close();
      socketRef.current = null;
    };
  }, [isAuthenticated, token]);

  return getSocket;
};

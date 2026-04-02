import { useEffect, useRef } from 'react';
import { useAuthStore } from '../store/authStore';

export const useWebsocket = (onMessage) => {
  const socketRef = useRef(null);
  const { token, isAuthenticated } = useAuthStore();

  useEffect(() => {
    if (!isAuthenticated || !token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const socket = new WebSocket(`${protocol}//${window.location.host}/api/ws?token=${token}`);
    
    socket.onopen = () => {
      console.log('WS Shared Connection Opened');
    };

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      if (onMessage) onMessage(data);
    };

    socket.onclose = () => {
      console.log('WS Connection Closed');
    };

    socketRef.current = socket;

    return () => {
      socket.close();
    };
  }, [isAuthenticated, token]);

  return socketRef.current;
};
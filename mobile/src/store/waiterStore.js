import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Alert } from 'react-native';
import * as api from '../api';

export const useWaiterStore = create((set, get) => ({
  tables: [],
  categories: [],
  products: [],
  activeOrder: null,
  loadingTables: false,
  loadingMenu: false,
  ws: null,
  // FIX #7: Parallel ulanishning oldini olish uchun flag
  _wsConnecting: false,

  // ─── WebSocket Connection ──────────────────────────────────────────
  connectWS: async (token) => {
    // FIX #7: Allaqachon ulanayotgan yoki ulanilgan bo'lsa — to'xtatish
    if (get().ws || get()._wsConnecting) return;

    set({ _wsConnecting: true });

    const wsUrl = process.env.EXPO_PUBLIC_API_URL
      ? process.env.EXPO_PUBLIC_API_URL.replace(/^http/, 'ws') + '/ws'
      : 'wss://kafe.securehub.uz/api/ws';

    const ws = new WebSocket(wsUrl, ['auth.' + token]);

    ws.onopen = () => {
      console.log('Waiter WS Connected');
      set({ _wsConnecting: false });
    };

    ws.onmessage = (e) => {
      try {
        const data = JSON.parse(e.data);
        if (data.type === 'tables_updated' || data.type === 'new_order') {
          get().fetchTables();
        }
      } catch (err) {
        console.error('WS Parse Error', err);
      }
    };

    ws.onclose = () => {
      console.log('Waiter WS Disconnected');
      set({ ws: null, _wsConnecting: false });
      // FIX #2: Reconnect da fresh token olish (eski closure token emas)
      setTimeout(async () => {
        try {
          const { useAuthStore } = require('./authStore');
          const authState = useAuthStore.getState();
          if (authState.isAuthenticated && authState.user?.role === 'waiter') {
            const freshToken = await AsyncStorage.getItem('token');
            if (freshToken) {
              get().connectWS(freshToken);
            }
          }
        } catch (e) {
          console.warn('WS reconnect error:', e);
        }
      }, 5000);
    };

    ws.onerror = (e) => {
      console.error('WS Error:', e.message);
      set({ _wsConnecting: false });
    };

    set({ ws });
  },

  disconnectWS: () => {
    const ws = get().ws;
    if (ws) {
      ws.close();
      set({ ws: null, _wsConnecting: false });
    }
  },

  // ─── Load all tables ────────────────────────────────────────────────
  fetchTables: async () => {
    set({ loadingTables: true });
    try {
      const res = await api.getTables();
      set({ tables: res.data || [] });
    } catch (e) {
      console.error('fetchTables error:', e);
      Alert.alert('Xatolik', e.response?.data?.error || e.message);
    } finally {
      set({ loadingTables: false });
    }
  },

  // ─── Load menu (categories + products) ─────────────────────────────
  fetchMenu: async () => {
    set({ loadingMenu: true });
    try {
      const [cRes, pRes] = await Promise.all([
        api.getCategories(),
        api.getProducts(),
      ]);
      set({
        categories: cRes.data || [],
        products: (pRes.data || []).filter((p) => p.is_active),
      });
    } catch (e) {
      console.error('fetchMenu error:', e);
      Alert.alert('Xatolik', e.response?.data?.error || e.message);
    } finally {
      set({ loadingMenu: false });
    }
  },

  // ─── Fetch active order for a table ────────────────────────────────
  fetchActiveOrder: async (tableID) => {
    try {
      const res = await api.getActiveOrderByTable(tableID);
      set({ activeOrder: res.data || null });
      return res.data || null;
    } catch (e) {
      set({ activeOrder: null });
      return null;
    }
  },

  clearActiveOrder: () => set({ activeOrder: null }),
}));

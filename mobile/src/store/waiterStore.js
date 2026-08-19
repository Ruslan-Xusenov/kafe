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

  // ─── WebSocket Connection ──────────────────────────────────────────
  connectWS: (token) => {
    if (get().ws) return; // Already connected

    const wsUrl = process.env.EXPO_PUBLIC_API_URL 
      ? process.env.EXPO_PUBLIC_API_URL.replace('http', 'ws') + '/ws'
      : 'wss://kafe.securehub.uz/api/ws';

    const ws = new WebSocket(wsUrl, ['auth.' + token]);

    ws.onopen = () => console.log('Waiter WS Connected');

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
      set({ ws: null });
      // Reconnect after 5 seconds if still authenticated
      setTimeout(() => {
        const authStore = require('./authStore').useAuthStore.getState();
        if (authStore.isAuthenticated && authStore.user?.role === 'waiter') {
          get().connectWS(token);
        }
      }, 5000);
    };

    set({ ws });
  },

  disconnectWS: () => {
    const ws = get().ws;
    if (ws) {
      ws.close();
      set({ ws: null });
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

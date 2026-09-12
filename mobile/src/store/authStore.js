import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import api, { getMe } from '../api';

export const useAuthStore = create((set) => ({
  user: null,
  isAuthenticated: false,
  loading: true,
  error: null,

  login: async (phone, password) => {
    set({ loading: true, error: null });
    try {
      const res = await api.post('/auth/login', { phone, password });
      const { token, user } = res.data;
      await AsyncStorage.setItem('token', token);
      await AsyncStorage.setItem('user', JSON.stringify(user));
      set({ user, isAuthenticated: true, loading: false });
      return { success: true, role: user.role };
    } catch (err) {
      set({
        error: err.response?.data?.error || 'Kirish xatosi',
        loading: false,
      });
      return { success: false };
    }
  },

  logout: async () => {
    await AsyncStorage.removeItem('token');
    await AsyncStorage.removeItem('user');
    set({ user: null, isAuthenticated: false, error: null });
  },

  // FIX #9: checkAuth endi serverni ham tekshiradi (expired token holati)
  checkAuth: async () => {
    try {
      const token = await AsyncStorage.getItem('token');
      const userStr = await AsyncStorage.getItem('user');

      if (!token || !userStr) {
        set({ loading: false, isAuthenticated: false });
        return;
      }

      // Avval local user ma'lumotini yuklash (tezroq UI uchun)
      const localUser = JSON.parse(userStr);
      set({ user: localUser, isAuthenticated: true });

      // Keyin server bilan tokenni tekshirish
      try {
        const res = await getMe();
        const serverUser = res.data;
        await AsyncStorage.setItem('user', JSON.stringify(serverUser));
        set({ user: serverUser, isAuthenticated: true, loading: false });
      } catch (serverErr) {
        // Token yaroqsiz — logout qilinadi
        console.warn('Token yaroqsiz, tizimdan chiqilmoqda...');
        await AsyncStorage.removeItem('token');
        await AsyncStorage.removeItem('user');
        set({ user: null, isAuthenticated: false, loading: false });
      }
    } catch (err) {
      set({ loading: false, isAuthenticated: false });
    }
  },
}));

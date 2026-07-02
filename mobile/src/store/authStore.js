import { create } from 'zustand';
import AsyncStorage from '@react-native-async-storage/async-storage';
import api from '../api';

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
        error: err.response?.data?.error || 'Kirishda xatolik yuz berdi',
        loading: false 
      });
      return { success: false };
    }
  },

  logout: async () => {
    await AsyncStorage.removeItem('token');
    await AsyncStorage.removeItem('user');
    set({ user: null, isAuthenticated: false });
  },

  checkAuth: async () => {
    try {
      const token = await AsyncStorage.getItem('token');
      const userStr = await AsyncStorage.getItem('user');
      
      if (!token || !userStr) {
        set({ loading: false, isAuthenticated: false });
        return;
      }
      
      const user = JSON.parse(userStr);
      set({ user, isAuthenticated: true, loading: false });
    } catch (err) {
      set({ loading: false, isAuthenticated: false });
    }
  }
}));

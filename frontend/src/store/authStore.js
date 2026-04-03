import { create } from 'zustand';
import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
});

// Interceptor to add token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Backend xato xabarlarini o'zbek tiliga tarjima
const translateError = (err) => {
  const serverMsg = err.response?.data?.error || '';
  const map = {
    'user already exists':       'Bu telefon raqam allaqachon ro\'yxatdan o\'tgan',
    'invalid phone or password':  'Telefon raqam yoki parol noto\'g\'ri',
    'unauthorized':               'Kirish huquqi yo\'q',
    'invalid credentials':        'Telefon raqam yoki parol noto\'g\'ri',
    'phone already taken':        'Bu telefon raqam band',
    'user not found':             'Foydalanuvchi topilmadi',
  };
  return map[serverMsg] || serverMsg || null;
};

export const useAuthStore = create((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  isAuthenticated: !!localStorage.getItem('token'),
  loading: false,
  error: null,

  login: async (phone, password) => {
    set({ loading: true, error: null });
    try {
      const res = await api.post('/auth/login', { phone, password });
      const { user, token } = res.data;
      localStorage.setItem('token', token);
      set({ user, token, isAuthenticated: true, loading: false });
      return { success: true, role: user.role };
    } catch (err) {
      const msg = translateError(err) || 'Kirishda xatolik yuz berdi';
      set({ error: msg, loading: false });
      return { success: false, error: msg };
    }
  },

  register: async (fullName, phone, password) => {
    set({ loading: true, error: null });
    try {
      const res = await api.post('/auth/register', { full_name: fullName, phone, password });
      const { user, token } = res.data;
      localStorage.setItem('token', token);
      set({ user, token, isAuthenticated: true, loading: false });
      return { success: true };
    } catch (err) {
      const msg = translateError(err) || 'Ro\'yxatdan o\'tishda xatolik yuz berdi';
      set({ error: msg, loading: false });
      return { success: false, error: msg };
    }
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null, isAuthenticated: false });
  },

  checkAuth: async () => {
    const token = localStorage.getItem('token');
    if (!token) return;
    try {
      const res = await api.get('/auth/me');
      set({ user: res.data, isAuthenticated: true });
    } catch (err) {
      localStorage.removeItem('token');
      set({ user: null, token: null, isAuthenticated: false });
    }
  }
}));

export default api;

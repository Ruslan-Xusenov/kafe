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

// Backend xato xabarlarini O'zbekchaga tarjima
const translateError = (err) => {
  // Tarmoq xatosi (backend o'chiq yoki ulanish yo'q)
  if (!err.response) {
    return 'Server bilan aloqa yo\'q. Internet yoki server holatini tekshiring.';
  }

  const serverMsg = (err.response?.data?.error || '').toLowerCase().trim();
  const map = {
    'user already exists':              'Bu telefon raqam allaqachon ro\'yxatdan o\'tgan',
    'invalid phone or password':         'Telefon raqam yoki parol noto\'g\'ri',
    'invalid credentials':               'Telefon raqam yoki parol noto\'g\'ri',
    'unauthorized':                      'Kirish huquqi yo\'q',
    'phone already taken':               'Bu telefon raqam band',
    'user not found':                    'Foydalanuvchi topilmadi',
    'этот номер телефона уже зарегистрирован': 'Bu telefon raqam allaqachon ro\'yxatdan o\'tgan',
    'wrong password':                    'Parol noto\'g\'ri',
    'account not found':                 'Bunday hisob topilmadi',
  };
  return map[serverMsg] || err.response?.data?.error || 'Kirish paytida xatolik yuz berdi';
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
      const msg = translateError(err) || 'Kirish paytida xatolik yuz berdi';
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
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      localStorage.removeItem('token');
      set({ user: null, token: null, isAuthenticated: false });
    }
  }
}));

export default api;

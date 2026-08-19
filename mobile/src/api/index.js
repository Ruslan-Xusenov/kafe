import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Platform } from 'react-native';

const BASE_URL =
  process.env.EXPO_PUBLIC_API_URL ||
  (Platform.OS === 'web'
    ? 'https://kafe.securehub.uz/api'
    : 'https://kafe.securehub.uz/api');

const api = axios.create({ baseURL: BASE_URL, timeout: 15000 });

import { useAuthStore } from '../store/authStore';

api.interceptors.request.use(async (config) => {
  const token = await AsyncStorage.getItem('token');
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      useAuthStore.getState().logout();
    }
    return Promise.reject(error);
  }
);

// ─── AUTH ──────────────────────────────────────────────────────────────────
export const login = (phone, password) =>
  api.post('/auth/login', { phone, password });

export const getMe = () => api.get('/auth/me');

// ─── TABLES ────────────────────────────────────────────────────────────────
export const getTables = () => api.get(`/tables/?_t=${Date.now()}`);

export const updateTable = (id, payload) => api.put(`/tables/${id}`, payload);

// ─── CATALOG ───────────────────────────────────────────────────────────────
export const getCategories = () => api.get('/catalog/categories');

export const getProducts = () => api.get('/catalog/products');

// ─── ORDERS ────────────────────────────────────────────────────────────────
export const createOrder = (payload) => api.post('/orders', payload);

export const getOrderById = (id) => api.get(`/orders/${id}`);

export const getActiveOrderByTable = (tableID) =>
  api.get(`/orders/active-by-table/${tableID}`);

export const getWaiterHistory = () => api.get('/orders/waiter/history');

export const updateOrderStatus = (id, status) =>
  api.put(`/orders/${id}/status`, { status });

export const addItemsToOrder = (id, items) =>
  api.post(`/orders/${id}/add-items`, { items });

export const cancelOrderItem = (orderId, itemId, quantity) =>
  api.post(`/orders/${orderId}/items/${itemId}/cancel`, { quantity });

export const cancelProductFromOrder = (orderId, productId, quantity) =>
  api.post(`/orders/${orderId}/cancel-product/${productId}`, { quantity });

export const setServiceFee = (id, percentage) =>
  api.put(`/orders/${id}/service-fee`, { percentage });

export const transferOrderTable = (from_table_id, to_table_id) =>
  api.post('/orders/transfer', { from_table_id, to_table_id });

export default api;

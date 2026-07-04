import axios from 'axios';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Platform } from 'react-native';

const baseURL = process.env.EXPO_PUBLIC_API_URL || (Platform.OS === 'web'
  ? 'http://localhost:8080/api'
  : 'http://10.0.2.2:8080/api');

const api = axios.create({
  baseURL, 
});

api.interceptors.request.use(async (config) => {
  const token = await AsyncStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export default api;

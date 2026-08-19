import React, { useEffect } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import AsyncStorage from '@react-native-async-storage/async-storage';

import { useAuthStore } from '../store/authStore';
import { useLangStore } from '../store/langStore';
import { useWaiterStore } from '../store/waiterStore';

// Screens
import LoginScreen from '../screens/LoginScreen';
import WaiterDashboardScreen from '../screens/WaiterDashboardScreen';
import TableDetailScreen from '../screens/TableDetailScreen';
import NewOrderScreen from '../screens/NewOrderScreen';
import OrderHistoryScreen from '../screens/OrderHistoryScreen';
import ProfileScreen from '../screens/ProfileScreen';

// Other roles (unchanged)
import KitchenScreen from '../screens/KitchenScreen';
import DeliveryScreen from '../screens/DeliveryScreen';
import HomeScreen from '../screens/HomeScreen';
import CartScreen from '../screens/CartScreen';

const Stack = createNativeStackNavigator();

const RootNavigator = () => {
  const { user, isAuthenticated, checkAuth, loading } = useAuthStore();
  const { loadLang } = useLangStore();
  const { connectWS, disconnectWS } = useWaiterStore();

  useEffect(() => {
    checkAuth();
    loadLang();
  }, []);

  useEffect(() => {
    if (isAuthenticated && user?.role === 'waiter') {
      AsyncStorage.getItem('token').then(token => {
        if (token) connectWS(token);
      });
    } else {
      disconnectWS();
    }
  }, [isAuthenticated, user]);

  if (loading) return null;

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false, animation: 'slide_from_right' }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Login" component={LoginScreen} />
        ) : user?.role === 'waiter' ? (
          // ─── WAITER STACK ───────────────────────────────────────
          <>
            <Stack.Screen name="WaiterDashboard" component={WaiterDashboardScreen} />
            <Stack.Screen name="TableDetail" component={TableDetailScreen} />
            <Stack.Screen name="NewOrder" component={NewOrderScreen} />
            <Stack.Screen name="History" component={OrderHistoryScreen} />
            <Stack.Screen name="Profile" component={ProfileScreen} />
          </>
        ) : user?.role === 'cook' ? (
          <Stack.Screen name="Kitchen" component={KitchenScreen} />
        ) : user?.role === 'courier' ? (
          <Stack.Screen name="Delivery" component={DeliveryScreen} />
        ) : (
          // Customer / Admin
          <>
            <Stack.Screen name="Home" component={HomeScreen} />
            <Stack.Screen name="Cart" component={CartScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default RootNavigator;

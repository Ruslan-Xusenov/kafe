import React, { useEffect } from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Utensils, ShoppingCart, User, ClipboardList, ChefHat, Truck } from 'lucide-react-native';

import { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';

// Screens
import HomeScreen from '../screens/HomeScreen';
import CartScreen from '../screens/CartScreen';
import ProfileScreen from '../screens/ProfileScreen';
import LoginScreen from '../screens/LoginScreen';
import WaiterScreen from '../screens/WaiterScreen';
import KitchenScreen from '../screens/KitchenScreen';
import DeliveryScreen from '../screens/DeliveryScreen';

const Stack = createNativeStackNavigator();
const Tab = createBottomTabNavigator();

const CustomerTabs = () => {
  const items = useCartStore(state => state.items);
  const cartCount = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarStyle: {
          backgroundColor: '#0d0d0f',
          borderTopColor: '#222',
        },
        tabBarActiveTintColor: '#f97316',
        tabBarInactiveTintColor: '#666',
      }}
    >
      <Tab.Screen 
        name="Home" 
        component={HomeScreen} 
        options={{
          tabBarIcon: ({ color, size }) => <Utensils color={color} size={size} />,
          tabBarLabel: 'Menyu'
        }}
      />
      <Tab.Screen 
        name="Cart" 
        component={CartScreen} 
        options={{
          tabBarIcon: ({ color, size }) => <ShoppingCart color={color} size={size} />,
          tabBarLabel: 'Savat',
          tabBarBadge: cartCount > 0 ? cartCount : null,
          tabBarBadgeStyle: { backgroundColor: '#f97316' }
        }}
      />
      <Tab.Screen 
        name="Profile" 
        component={ProfileScreen} 
        options={{
          tabBarIcon: ({ color, size }) => <User color={color} size={size} />,
          tabBarLabel: 'Profil'
        }}
      />
    </Tab.Navigator>
  );
};

const RootNavigator = () => {
  const { user, isAuthenticated, checkAuth, loading } = useAuthStore();

  useEffect(() => {
    checkAuth();
  }, []);

  if (loading) {
    return null; // or a splash screen
  }

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {isAuthenticated ? (
          // Role-based routing
          user.role === 'waiter' ? (
            <Stack.Screen name="WaiterScreen" component={WaiterScreen} />
          ) : user.role === 'cook' ? (
            <Stack.Screen name="KitchenScreen" component={KitchenScreen} />
          ) : user.role === 'courier' ? (
            <Stack.Screen name="DeliveryScreen" component={DeliveryScreen} />
          ) : (
            // Customer or Admin routes
            <Stack.Screen name="CustomerMain" component={CustomerTabs} />
          )
        ) : (
          // Not authenticated
          <>
            <Stack.Screen name="CustomerMain" component={CustomerTabs} />
            <Stack.Screen name="Login" component={LoginScreen} />
          </>
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default RootNavigator;

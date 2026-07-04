import React from 'react';
import { View, Text, StyleSheet, Button } from 'react-native';
import { useAuthStore } from '../store/authStore';

const ProfileScreen = ({ navigation }) => {
  const { user, isAuthenticated, logout } = useAuthStore();

  return (
    <View style={styles.container}>
      <Text style={styles.text}>Профиль</Text>
      {isAuthenticated ? (
        <>
          <Text style={styles.subtext}>Добро пожаловать, {user.full_name}</Text>
          <Button title="Выйти" onPress={logout} color="#ef4444" />
        </>
      ) : (
        <Button title="Войти в систему" onPress={() => navigation.navigate('Login')} color="#f97316" />
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f', alignItems: 'center', justifyContent: 'center', gap: 20 },
  text: { color: '#fff', fontSize: 20, fontWeight: 'bold' },
  subtext: { color: '#aaa', fontSize: 16 }
});

export default ProfileScreen;

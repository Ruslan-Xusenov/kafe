import React, { useState } from 'react';
import { View, Text, TextInput, StyleSheet, TouchableOpacity, ActivityIndicator } from 'react-native';
import { useAuthStore } from '../store/authStore';

const LoginScreen = ({ navigation }) => {
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const { login, loading, error } = useAuthStore();

  const handleLogin = async () => {
    if (!phone || !password) return;
    await login(phone, password);
  };

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Kirish</Text>
      {error && <Text style={styles.error}>{error}</Text>}
      
      <TextInput 
        style={styles.input}
        placeholder="+998901234567"
        placeholderTextColor="#666"
        value={phone}
        onChangeText={setPhone}
      />
      <TextInput 
        style={styles.input}
        placeholder="Parol"
        placeholderTextColor="#666"
        secureTextEntry
        value={password}
        onChangeText={setPassword}
      />
      
      <TouchableOpacity style={styles.button} onPress={handleLogin} disabled={loading}>
        {loading ? <ActivityIndicator color="#fff" /> : <Text style={styles.buttonText}>Tizimga kirish</Text>}
      </TouchableOpacity>
      
      <TouchableOpacity onPress={() => navigation.navigate('CustomerMain')} style={{marginTop: 20}}>
        <Text style={{color: '#f97316'}}>Menyuga qaytish (Mehmon)</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f', alignItems: 'center', justifyContent: 'center', padding: 20 },
  title: { color: '#fff', fontSize: 28, fontWeight: 'bold', marginBottom: 30 },
  input: { width: '100%', backgroundColor: '#1a1a1f', color: '#fff', padding: 15, borderRadius: 10, marginBottom: 15, borderWidth: 1, borderColor: '#333' },
  button: { width: '100%', backgroundColor: '#f97316', padding: 15, borderRadius: 10, alignItems: 'center' },
  buttonText: { color: '#fff', fontSize: 16, fontWeight: 'bold' },
  error: { color: '#ef4444', marginBottom: 15 }
});

export default LoginScreen;

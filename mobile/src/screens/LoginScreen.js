import React, { useState } from 'react';
import {
  View, Text, TextInput, StyleSheet, TouchableOpacity,
  ActivityIndicator, KeyboardAvoidingView, Platform, ScrollView,
} from 'react-native';
import { useAuthStore } from '../store/authStore';
import { useLangStore } from '../store/langStore';

const LoginScreen = () => {
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const { login, loading, error } = useAuthStore();
  const { t, lang, setLang } = useLangStore();

  const handleLogin = async () => {
    const cleanPhone = phone.replace(/\s+/g, '');
    if (!cleanPhone || !password.trim()) return;
    await login(cleanPhone, password);
  };

  return (
    <KeyboardAvoidingView
      style={styles.root}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scroll} keyboardShouldPersistTaps="handled">
        {/* Logo / Header */}
        <View style={styles.header}>
          <View style={styles.logoCircle}>
            <Text style={styles.logoEmoji}>🍽️</Text>
          </View>
          <Text style={styles.appName}>Kafe</Text>
          <Text style={styles.appSub}>{t.waiterPanel}</Text>
        </View>

        {/* Lang Toggle */}
        <View style={styles.langRow}>
          <TouchableOpacity
            style={[styles.langBtn, lang === 'uz' && styles.langBtnActive]}
            onPress={() => setLang('uz')}
          >
            <Text style={[styles.langText, lang === 'uz' && styles.langTextActive]}>UZ</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[styles.langBtn, lang === 'ru' && styles.langBtnActive]}
            onPress={() => setLang('ru')}
          >
            <Text style={[styles.langText, lang === 'ru' && styles.langTextActive]}>RU</Text>
          </TouchableOpacity>
        </View>

        {/* Card */}
        <View style={styles.card}>
          <Text style={styles.cardTitle}>{t.login}</Text>

          {error ? <Text style={styles.errorMsg}>{error}</Text> : null}

          <Text style={styles.label}>{t.phone}</Text>
          <TextInput
            style={styles.input}
            placeholder="+998 90 123 45 67"
            placeholderTextColor="#555"
            value={phone}
            onChangeText={setPhone}
            keyboardType="phone-pad"
            autoCapitalize="none"
          />

          <Text style={styles.label}>{t.password}</Text>
          <TextInput
            style={styles.input}
            placeholder="••••••"
            placeholderTextColor="#555"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />

          <TouchableOpacity
            style={[styles.loginBtn, loading && styles.loginBtnDisabled]}
            onPress={handleLogin}
            disabled={loading}
            activeOpacity={0.8}
          >
            {loading ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <Text style={styles.loginBtnText}>{t.loginBtn}</Text>
            )}
          </TouchableOpacity>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },
  scroll: { flexGrow: 1, justifyContent: 'center', padding: 24 },

  header: { alignItems: 'center', marginBottom: 32 },
  logoCircle: {
    width: 80, height: 80, borderRadius: 40,
    backgroundColor: 'rgba(249,115,22,0.15)',
    borderWidth: 2, borderColor: 'rgba(249,115,22,0.4)',
    justifyContent: 'center', alignItems: 'center', marginBottom: 12,
  },
  logoEmoji: { fontSize: 36 },
  appName: { fontSize: 32, fontWeight: '800', color: '#fff', letterSpacing: 1 },
  appSub: { color: '#888', marginTop: 4, fontSize: 14 },

  langRow: { flexDirection: 'row', justifyContent: 'center', gap: 10, marginBottom: 28 },
  langBtn: {
    paddingHorizontal: 22, paddingVertical: 8, borderRadius: 20,
    backgroundColor: '#1a1a1f', borderWidth: 1, borderColor: '#333',
  },
  langBtnActive: { backgroundColor: '#f97316', borderColor: '#f97316' },
  langText: { color: '#888', fontWeight: '700', fontSize: 13 },
  langTextActive: { color: '#fff' },

  card: {
    backgroundColor: '#141417', borderRadius: 20, padding: 24,
    borderWidth: 1, borderColor: '#252530',
  },
  cardTitle: { color: '#fff', fontSize: 22, fontWeight: '700', marginBottom: 20 },

  errorMsg: {
    backgroundColor: 'rgba(239,68,68,0.12)', color: '#f87171',
    padding: 12, borderRadius: 10, marginBottom: 16, fontSize: 13,
  },

  label: { color: '#888', fontSize: 12, fontWeight: '600', marginBottom: 6, letterSpacing: 0.5 },
  input: {
    backgroundColor: '#1a1a1f', color: '#fff', padding: 14,
    borderRadius: 12, marginBottom: 16, borderWidth: 1, borderColor: '#2a2a35', fontSize: 15,
  },

  loginBtn: {
    backgroundColor: '#f97316', padding: 16, borderRadius: 14,
    alignItems: 'center', marginTop: 4,
    shadowColor: '#f97316', shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3, shadowRadius: 8, elevation: 6,
  },
  loginBtnDisabled: { opacity: 0.6 },
  loginBtnText: { color: '#fff', fontSize: 16, fontWeight: '800', letterSpacing: 0.5 },
});

export default LoginScreen;

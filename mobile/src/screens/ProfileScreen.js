import React from 'react';
import {
  View, Text, StyleSheet, TouchableOpacity, Alert,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LogOut, Globe } from 'lucide-react-native';
import { useAuthStore } from '../store/authStore';
import { useLangStore } from '../store/langStore';

const ProfileScreen = ({ navigation }) => {
  const { user, logout } = useAuthStore();
  const { t, lang, setLang } = useLangStore();

  const handleLogout = () => {
    Alert.alert(t.logout, t.logout + '?', [
      { text: t.cancel, style: 'cancel' },
      { text: t.confirm, style: 'destructive', onPress: logout },
    ]);
  };

  return (
    <SafeAreaView style={styles.root}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>{t.profile}</Text>
      </View>

      {/* Avatar */}
      <View style={styles.avatarSection}>
        <View style={styles.avatarCircle}>
          <Text style={styles.avatarText}>{user?.full_name?.[0]?.toUpperCase() || '?'}</Text>
        </View>
        <Text style={styles.userName}>{user?.full_name}</Text>
        <View style={styles.roleBadge}>
          <Text style={styles.roleText}>Ofitsant</Text>
        </View>
        <Text style={styles.userPhone}>{user?.phone}</Text>
      </View>

      {/* Lang section */}
      <View style={styles.section}>
        <View style={styles.sectionHeader}>
          <Globe color="#888" size={16} />
          <Text style={styles.sectionTitle}>{t.chooseLang}</Text>
        </View>
        <View style={styles.langRow}>
          <TouchableOpacity
            style={[styles.langBtn, lang === 'uz' && styles.langBtnActive]}
            onPress={() => setLang('uz')}
            activeOpacity={0.8}
          >
            <Text style={styles.langFlag}>🇺🇿</Text>
            <Text style={[styles.langLabel, lang === 'uz' && styles.langLabelActive]}>
              O'zbek
            </Text>
            {lang === 'uz' && <View style={styles.checkDot} />}
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.langBtn, lang === 'ru' && styles.langBtnActive]}
            onPress={() => setLang('ru')}
            activeOpacity={0.8}
          >
            <Text style={styles.langFlag}>🇷🇺</Text>
            <Text style={[styles.langLabel, lang === 'ru' && styles.langLabelActive]}>
              Русский
            </Text>
            {lang === 'ru' && <View style={styles.checkDot} />}
          </TouchableOpacity>
        </View>
      </View>

      {/* App info */}
      <View style={styles.section}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>{t.version}</Text>
          <Text style={styles.infoValue}>1.0.0</Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>Server</Text>
          <Text style={styles.infoValue}>kafe.securehub.uz</Text>
        </View>
      </View>

      {/* Logout */}
      <TouchableOpacity style={styles.logoutBtn} onPress={handleLogout} activeOpacity={0.8}>
        <LogOut color="#ef4444" size={18} />
        <Text style={styles.logoutText}>{t.logout}</Text>
      </TouchableOpacity>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },
  header: {
    paddingHorizontal: 20, paddingVertical: 16,
    borderBottomWidth: 1, borderBottomColor: '#1a1a22',
  },
  headerTitle: { color: '#fff', fontSize: 22, fontWeight: '800' },

  avatarSection: { alignItems: 'center', paddingVertical: 36 },
  avatarCircle: {
    width: 90, height: 90, borderRadius: 45,
    backgroundColor: 'rgba(249,115,22,0.15)',
    borderWidth: 2, borderColor: 'rgba(249,115,22,0.4)',
    justifyContent: 'center', alignItems: 'center', marginBottom: 14,
  },
  avatarText: { fontSize: 36, fontWeight: '800', color: '#f97316' },
  userName: { color: '#fff', fontSize: 22, fontWeight: '800' },
  roleBadge: {
    backgroundColor: 'rgba(249,115,22,0.15)', paddingHorizontal: 14, paddingVertical: 4,
    borderRadius: 20, marginTop: 8, borderWidth: 1, borderColor: 'rgba(249,115,22,0.3)',
  },
  roleText: { color: '#f97316', fontWeight: '700', fontSize: 12 },
  userPhone: { color: '#666', fontSize: 14, marginTop: 8 },

  section: {
    marginHorizontal: 20, marginBottom: 16,
    backgroundColor: '#141417', borderRadius: 16,
    padding: 16, borderWidth: 1, borderColor: '#1e1e26',
  },
  sectionHeader: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 14 },
  sectionTitle: { color: '#888', fontSize: 13, fontWeight: '700' },

  langRow: { flexDirection: 'row', gap: 10 },
  langBtn: {
    flex: 1, borderRadius: 14, padding: 14,
    backgroundColor: '#1a1a22', borderWidth: 1.5, borderColor: '#2a2a35',
    alignItems: 'center', gap: 6,
  },
  langBtnActive: { borderColor: '#f97316', backgroundColor: 'rgba(249,115,22,0.08)' },
  langFlag: { fontSize: 28 },
  langLabel: { color: '#888', fontWeight: '700', fontSize: 14 },
  langLabelActive: { color: '#f97316' },
  checkDot: {
    width: 8, height: 8, borderRadius: 4, backgroundColor: '#f97316',
  },

  infoRow: {
    flexDirection: 'row', justifyContent: 'space-between',
    paddingVertical: 10, borderBottomWidth: 1, borderBottomColor: '#1e1e26',
  },
  infoLabel: { color: '#666', fontSize: 14 },
  infoValue: { color: '#ccc', fontSize: 14, fontWeight: '600' },

  logoutBtn: {
    marginHorizontal: 20, flexDirection: 'row', alignItems: 'center', justifyContent: 'center',
    gap: 10, backgroundColor: 'rgba(239,68,68,0.1)', padding: 16, borderRadius: 16,
    borderWidth: 1, borderColor: 'rgba(239,68,68,0.3)',
  },
  logoutText: { color: '#ef4444', fontSize: 16, fontWeight: '800' },
});

export default ProfileScreen;

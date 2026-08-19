import React, { useEffect } from 'react';
import {
  View, Text, StyleSheet, FlatList, TouchableOpacity,
  ActivityIndicator, RefreshControl,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuthStore } from '../store/authStore';
import { useWaiterStore } from '../store/waiterStore';
import { useLangStore } from '../store/langStore';
import { RefreshCw, LogOut, ClipboardList, Clock } from 'lucide-react-native';

const WaiterDashboardScreen = ({ navigation }) => {
  const { logout, user } = useAuthStore();
  const { tables, fetchTables, fetchMenu, loadingTables } = useWaiterStore();
  const { t } = useLangStore();

  useEffect(() => {
    fetchTables();
    fetchMenu();
  }, []);

  const handleTablePress = (table) => {
    navigation.navigate('TableDetail', { table });
  };

  const freeTables = tables.filter((t) => t.status === 'free').length;
  const occupiedTables = tables.filter((t) => t.status === 'occupied').length;

  const renderTable = ({ item }) => {
    const isFree = item.status === 'free';
    return (
      <TouchableOpacity
        style={[styles.tableCard, isFree ? styles.cardFree : styles.cardOccupied]}
        onPress={() => handleTablePress(item)}
        activeOpacity={0.75}
      >
        <View style={[styles.tableIconBg, isFree ? styles.iconBgFree : styles.iconBgOccupied]}>
          <Text style={styles.tableIcon}>🪑</Text>
        </View>
        <Text style={[styles.tableName, isFree ? styles.textFree : styles.textOccupied]}>
          {item.name || `${t.table} ${item.number || item.id}`}
        </Text>
        <View style={[styles.badge, isFree ? styles.badgeFree : styles.badgeOccupied]}>
          <Text style={[styles.badgeText, isFree ? styles.badgeTextFree : styles.badgeTextOccupied]}>
            {isFree ? t.free : t.occupied}
          </Text>
        </View>
        {item.capacity && (
          <Text style={styles.capacity}>👥 {item.capacity}</Text>
        )}
      </TouchableOpacity>
    );
  };

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <View>
          <Text style={styles.greeting}>👋 {user?.full_name?.split(' ')[0]}</Text>
          <Text style={styles.headerTitle}>{t.waiterPanel}</Text>
        </View>
        <View style={styles.headerActions}>
          <TouchableOpacity
            style={styles.iconBtn}
            onPress={() => navigation.navigate('History')}
          >
            <Clock color="#f97316" size={20} />
          </TouchableOpacity>
          <TouchableOpacity style={styles.iconBtn} onPress={fetchTables}>
            <RefreshCw color="#888" size={20} />
          </TouchableOpacity>
          <TouchableOpacity
            style={styles.iconBtn}
            onPress={() => navigation.navigate('Profile')}
          >
            <Text style={styles.avatarText}>
              {user?.full_name?.[0]?.toUpperCase() || 'U'}
            </Text>
          </TouchableOpacity>
          <TouchableOpacity style={styles.logoutBtn} onPress={logout}>
            <LogOut color="#ef4444" size={20} />
          </TouchableOpacity>
        </View>
      </View>

      {/* Stats */}
      <View style={styles.statsRow}>
        <View style={styles.statCard}>
          <Text style={styles.statNum}>{tables.length}</Text>
          <Text style={styles.statLabel}>{t.tables}</Text>
        </View>
        <View style={[styles.statCard, styles.statCardGreen]}>
          <Text style={[styles.statNum, { color: '#10b981' }]}>{freeTables}</Text>
          <Text style={styles.statLabel}>{t.free}</Text>
        </View>
        <View style={[styles.statCard, styles.statCardRed]}>
          <Text style={[styles.statNum, { color: '#ef4444' }]}>{occupiedTables}</Text>
          <Text style={styles.statLabel}>{t.occupied}</Text>
        </View>
      </View>

      {/* Tables Grid */}
      {loadingTables && tables.length === 0 ? (
        <View style={styles.center}>
          <ActivityIndicator size="large" color="#f97316" />
        </View>
      ) : (
        <FlatList
          data={tables}
          renderItem={renderTable}
          keyExtractor={(item) => item.id.toString()}
          numColumns={2}
          columnWrapperStyle={styles.columnWrapper}
          contentContainerStyle={styles.listContent}
          refreshControl={
            <RefreshControl
              refreshing={loadingTables}
              onRefresh={fetchTables}
              tintColor="#f97316"
              colors={['#f97316']}
            />
          }
          ListEmptyComponent={
            <View style={styles.center}>
              <Text style={styles.emptyText}>🪑</Text>
              <Text style={styles.emptyLabel}>{t.tables}</Text>
            </View>
          }
        />
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },

  header: {
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    paddingHorizontal: 20, paddingTop: 16, paddingBottom: 12,
    borderBottomWidth: 1, borderBottomColor: '#1a1a22',
  },
  greeting: { color: '#888', fontSize: 13 },
  headerTitle: { color: '#fff', fontSize: 22, fontWeight: '800', marginTop: 2 },
  headerActions: { flexDirection: 'row', gap: 8, alignItems: 'center' },
  iconBtn: {
    width: 38, height: 38, borderRadius: 19,
    backgroundColor: '#1a1a22', justifyContent: 'center', alignItems: 'center',
  },
  avatarText: { color: '#f97316', fontWeight: '800', fontSize: 15 },
  logoutBtn: {
    width: 38, height: 38, borderRadius: 19,
    backgroundColor: 'rgba(239,68,68,0.12)', justifyContent: 'center', alignItems: 'center',
  },

  statsRow: { flexDirection: 'row', gap: 8, paddingHorizontal: 16, paddingVertical: 14 },
  statCard: {
    flex: 1, backgroundColor: '#141417', borderRadius: 14, paddingVertical: 12, paddingHorizontal: 6,
    alignItems: 'center', borderWidth: 1, borderColor: '#222',
  },
  statCardGreen: { borderColor: 'rgba(16,185,129,0.2)', backgroundColor: 'rgba(16,185,129,0.05)' },
  statCardRed: { borderColor: 'rgba(239,68,68,0.2)', backgroundColor: 'rgba(239,68,68,0.05)' },
  statNum: { fontSize: 24, fontWeight: '800', color: '#fff' },
  statLabel: { color: '#666', fontSize: 11, marginTop: 2, textAlign: 'center' },

  listContent: { paddingHorizontal: 16, paddingBottom: 24 },
  columnWrapper: { justifyContent: 'space-between', marginBottom: 16 },

  tableCard: {
    width: '48%', borderRadius: 18, padding: 12, alignItems: 'center', justifyContent: 'center',
    borderWidth: 1.5, aspectRatio: 0.85,
  },
  cardFree: {
    backgroundColor: 'rgba(16,185,129,0.07)',
    borderColor: 'rgba(16,185,129,0.3)',
  },
  cardOccupied: {
    backgroundColor: 'rgba(239,68,68,0.07)',
    borderColor: 'rgba(239,68,68,0.3)',
  },
  tableIconBg: {
    width: 48, height: 48, borderRadius: 24,
    justifyContent: 'center', alignItems: 'center', marginBottom: 10,
  },
  iconBgFree: { backgroundColor: 'rgba(16,185,129,0.15)' },
  iconBgOccupied: { backgroundColor: 'rgba(239,68,68,0.15)' },
  tableIcon: { fontSize: 22 },
  tableName: { fontSize: 16, fontWeight: '800', marginBottom: 8, textAlign: 'center' },
  textFree: { color: '#10b981' },
  textOccupied: { color: '#ef4444' },
  badge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 20 },
  badgeFree: { backgroundColor: 'rgba(16,185,129,0.15)' },
  badgeOccupied: { backgroundColor: 'rgba(239,68,68,0.15)' },
  badgeText: { fontSize: 11, fontWeight: '700' },
  badgeTextFree: { color: '#10b981' },
  badgeTextOccupied: { color: '#ef4444' },
  capacity: { color: '#555', fontSize: 12, marginTop: 8 },

  center: { flex: 1, justifyContent: 'center', alignItems: 'center', paddingTop: 60 },
  emptyText: { fontSize: 48 },
  emptyLabel: { color: '#555', marginTop: 8 },
});

export default WaiterDashboardScreen;

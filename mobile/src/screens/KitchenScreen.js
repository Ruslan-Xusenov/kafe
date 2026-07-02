import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator, SafeAreaView, Alert, useWindowDimensions } from 'react-native';
import { useAuthStore } from '../store/authStore';
import api from '../api';
import { ChefHat, Flame, CheckCircle2, Clock, RefreshCw } from 'lucide-react-native';

const STATUS_CONFIG = {
  new:       { label: 'Yangi',          color: '#818cf8', bg: 'rgba(99,102,241,0.15)' },
  preparing: { label: 'Tayyorlanmoqda', color: '#fbbf24', bg: 'rgba(251,191,36,0.15)' },
  ready:     { label: 'Tayyor',         color: '#34d399', bg: 'rgba(16,185,129,0.15)' },
};

const KitchenScreen = () => {
  const { width } = useWindowDimensions();
  const isDesktop = width > 768;

  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const logout = useAuthStore(s => s.logout);

  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const res = await api.get('/orders/active');
      const kitchenOrders = (res.data || []).filter(o =>
        o.status === 'new' || o.status === 'preparing' || o.status === 'ready'
      );
      setOrders(kitchenOrders);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleRefresh = () => {
    setRefreshing(true);
    fetchOrders();
  };

  const updateStatus = async (orderId, newStatus) => {
    try {
      await api.put(`/orders/${orderId}/status`, { status: newStatus });
      fetchOrders();
    } catch (err) {
      Alert.alert('Xato', 'Statusni yangilashda xatolik');
    }
  };

  const renderOrderCard = (order) => {
    const isNew = order.status === 'new';
    const isPrep = order.status === 'preparing';
    const isReady = order.status === 'ready';

    return (
      <View key={order.id} style={styles.orderCard}>
        <View style={styles.orderTop}>
          <Text style={styles.orderNum}>#{order.id}</Text>
          <View style={styles.timeWrap}>
            <Clock size={12} color="#888" />
            <Text style={styles.timeText}>
              {new Date(order.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
            </Text>
          </View>
        </View>

        <View style={styles.itemsList}>
          {(order.items || []).map((item, idx) => (
            <View key={idx} style={styles.itemRow}>
              <Text style={styles.itemQty}>{item.quantity} {item.unit || 'x'}</Text>
              <Text style={styles.itemName}>{item.product_name}</Text>
            </View>
          ))}
        </View>

        {order.comment ? (
          <View style={styles.noteBox}>
            <Text style={styles.noteText}>💬 {order.comment}</Text>
          </View>
        ) : null}

        <View style={styles.actions}>
          {isNew && (
            <TouchableOpacity style={[styles.btn, { backgroundColor: STATUS_CONFIG.preparing.bg }]} onPress={() => updateStatus(order.id, 'preparing')}>
              <Flame size={16} color={STATUS_CONFIG.preparing.color} />
              <Text style={[styles.btnText, { color: STATUS_CONFIG.preparing.color }]}>Pishirishni boshlash</Text>
            </TouchableOpacity>
          )}
          {isPrep && (
            <TouchableOpacity style={[styles.btn, { backgroundColor: STATUS_CONFIG.ready.bg }]} onPress={() => updateStatus(order.id, 'ready')}>
              <CheckCircle2 size={16} color={STATUS_CONFIG.ready.color} />
              <Text style={[styles.btnText, { color: STATUS_CONFIG.ready.color }]}>Tayyor!</Text>
            </TouchableOpacity>
          )}
          {isReady && (
            <View style={styles.waitingCourier}>
              <Text style={{color: STATUS_CONFIG.ready.color, fontWeight: 'bold'}}>Kuryerni kutmoqda...</Text>
            </View>
          )}
        </View>
      </View>
    );
  };

  const renderColumn = (status) => {
    const colOrders = orders.filter(o => o.status === status);
    const cfg = STATUS_CONFIG[status];
    
    // On mobile, column takes 100% width and auto height (grows with content)
    // On desktop, column takes 320px width and 100% height
    return (
      <View style={[styles.column, { width: isDesktop ? 320 : '100%', height: isDesktop ? '100%' : 'auto' }]} key={status}>
        <View style={[styles.colHeader, { borderBottomColor: cfg.color }]}>
          <Text style={[styles.colLabel, { color: cfg.color }]}>{cfg.label}</Text>
          <View style={[styles.badge, { backgroundColor: cfg.bg }]}>
            <Text style={[styles.badgeText, { color: cfg.color }]}>{colOrders.length}</Text>
          </View>
        </View>
        <View style={isDesktop ? styles.colBodyDesktop : styles.colBodyMobile}>
          {isDesktop ? (
            <ScrollView showsVerticalScrollIndicator={false}>
              {colOrders.length === 0 ? <Text style={styles.emptyText}>Bo'sh</Text> : colOrders.map(renderOrderCard)}
            </ScrollView>
          ) : (
            <View>
              {colOrders.length === 0 ? <Text style={styles.emptyText}>Bo'sh</Text> : colOrders.map(renderOrderCard)}
            </View>
          )}
        </View>
      </View>
    );
  };

  if (loading) return (
    <View style={styles.center}>
      <ActivityIndicator size="large" color="#f97316" />
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <View style={styles.headerLeft}>
          <ChefHat color="#fff" size={24} />
          <View>
            <Text style={styles.headerTitle}>Oshxona</Text>
            <Text style={styles.headerSubtitle}>{orders.length} ta aktiv buyurtma</Text>
          </View>
        </View>
        <View style={{flexDirection: 'row', gap: 10}}>
          <TouchableOpacity onPress={handleRefresh} style={styles.iconBtn}>
            <RefreshCw color="#888" size={20} />
          </TouchableOpacity>
          <TouchableOpacity onPress={logout} style={styles.logoutBtn}>
            <Text style={styles.logoutText}>Chiqish</Text>
          </TouchableOpacity>
        </View>
      </View>

      <ScrollView 
        horizontal={isDesktop}
        showsHorizontalScrollIndicator={isDesktop} 
        showsVerticalScrollIndicator={!isDesktop}
        style={styles.board}
        contentContainerStyle={isDesktop ? { paddingHorizontal: 15, gap: 15 } : { padding: 15, gap: 20 }}
      >
        {renderColumn('new')}
        {renderColumn('preparing')}
        {renderColumn('ready')}
      </ScrollView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f' },
  center: { flex: 1, backgroundColor: '#0d0d0f', justifyContent: 'center', alignItems: 'center' },
  header: { padding: 20, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', borderBottomWidth: 1, borderBottomColor: '#222' },
  headerLeft: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  headerTitle: { color: '#fff', fontSize: 20, fontWeight: 'bold' },
  headerSubtitle: { color: '#888', fontSize: 12 },
  iconBtn: { padding: 8, backgroundColor: '#1a1a1f', borderRadius: 8 },
  logoutBtn: { padding: 8, paddingHorizontal: 12, backgroundColor: 'rgba(239,68,68,0.15)', borderRadius: 8, justifyContent: 'center' },
  logoutText: { color: '#ef4444', fontWeight: 'bold', fontSize: 12 },
  
  board: { flex: 1 },
  column: { padding: 0 },
  colHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', paddingBottom: 10, borderBottomWidth: 2, marginBottom: 15 },
  colLabel: { fontSize: 16, fontWeight: 'bold', textTransform: 'uppercase' },
  badge: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: 12 },
  badgeText: { fontWeight: 'bold', fontSize: 12 },
  colBodyDesktop: { flex: 1 },
  colBodyMobile: { paddingBottom: 10 },
  emptyText: { color: '#444', textAlign: 'center', marginTop: 50, fontSize: 16 },

  orderCard: { backgroundColor: '#1a1a1f', padding: 15, borderRadius: 12, marginBottom: 15, borderWidth: 1, borderColor: '#222' },
  orderTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 },
  orderNum: { color: '#f97316', fontSize: 18, fontWeight: 'bold' },
  timeWrap: { flexDirection: 'row', alignItems: 'center', gap: 4 },
  timeText: { color: '#888', fontSize: 12 },
  
  itemsList: { marginBottom: 10 },
  itemRow: { flexDirection: 'row', alignItems: 'flex-start', gap: 8, marginBottom: 6 },
  itemQty: { color: '#f97316', fontWeight: 'bold', minWidth: 30 },
  itemName: { color: '#fff', flex: 1 },
  
  noteBox: { backgroundColor: 'rgba(255,255,255,0.05)', padding: 10, borderRadius: 8, marginBottom: 10 },
  noteText: { color: '#aaa', fontStyle: 'italic', fontSize: 12 },
  
  actions: { marginTop: 5 },
  btn: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 8, padding: 12, borderRadius: 8 },
  btnText: { fontWeight: 'bold', fontSize: 14 },
  waitingCourier: { padding: 12, alignItems: 'center', backgroundColor: 'rgba(255,255,255,0.05)', borderRadius: 8 }
});

export default KitchenScreen;

import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ScrollView, TouchableOpacity, ActivityIndicator, SafeAreaView, Alert, Linking } from 'react-native';
import { useAuthStore } from '../store/authStore';
import api from '../api';
import { Truck, MapPin, Phone, CheckCircle, Navigation, RefreshCw, Package } from 'lucide-react-native';

const DeliveryScreen = () => {
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
      setOrders((res.data || []).filter(o => o.status === 'ready' || o.status === 'on_way'));
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

  const pickUp = async (orderId) => {
    try {
      await api.post(`/orders/${orderId}/assign`);
      fetchOrders();
    } catch (err) {
      Alert.alert('Ошибка', 'Ошибка при принятии заказа');
    }
  };

  const deliver = async (orderId) => {
    try {
      await api.put(`/orders/${orderId}/status`, { status: 'delivered' });
      fetchOrders();
    } catch (err) {
      Alert.alert('Ошибка', 'Ошибка при обновлении статуса');
    }
  };

  const readyOrders = orders.filter(o => o.status === 'ready');
  const activeOrders = orders.filter(o => o.status === 'on_way');

  const renderReadyOrder = (order) => (
    <View key={order.id} style={[styles.card, styles.readyCard]}>
      <View style={styles.cardTop}>
        <Text style={styles.cardId}>#{order.id}</Text>
        <Text style={styles.cardPrice}>{order.total_price?.toLocaleString()} сум</Text>
      </View>
      <View style={styles.infoRow}>
        <MapPin size={16} color="#888" />
        <Text style={styles.infoText}>{order.address}</Text>
      </View>
      <TouchableOpacity style={styles.btnPrimary} onPress={() => pickUp(order.id)}>
        <Navigation size={18} color="#fff" />
        <Text style={styles.btnTextPrimary}>Забрать</Text>
      </TouchableOpacity>
    </View>
  );

  const renderActiveOrder = (order) => (
    <View key={order.id} style={[styles.card, styles.activeCard]}>
      <View style={styles.cardTop}>
        <Text style={styles.cardId}>#{order.id}</Text>
        <View style={styles.onWayBadge}>
          <Text style={styles.onWayText}>🚴 В пути</Text>
        </View>
      </View>
      
      <View style={styles.details}>
        <View style={styles.infoRow}>
          <MapPin size={16} color="#888" />
          <Text style={styles.infoText}>{order.address}</Text>
        </View>
        <TouchableOpacity style={styles.infoRow} onPress={() => Linking.openURL(`tel:${order.phone}`)}>
          <Phone size={16} color="#f97316" />
          <Text style={[styles.infoText, { color: '#f97316' }]}>{order.phone}</Text>
        </TouchableOpacity>
      </View>

      <View style={styles.itemsWrap}>
        {(order.items || []).map((item, i) => (
          <View key={i} style={styles.itemChip}>
            <Text style={styles.itemChipText}>{item.quantity}x {item.product_name}</Text>
          </View>
        ))}
      </View>

      <TouchableOpacity style={styles.btnSuccess} onPress={() => deliver(order.id)}>
        <CheckCircle size={18} color="#34d399" />
        <Text style={styles.btnTextSuccess}>Доставлено</Text>
      </TouchableOpacity>
    </View>
  );

  if (loading) return (
    <View style={styles.center}>
      <ActivityIndicator size="large" color="#f97316" />
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <View style={styles.headerLeft}>
          <Truck color="#fff" size={24} />
          <View>
            <Text style={styles.headerTitle}>Панель Курьера</Text>
            <Text style={styles.headerSubtitle}>{readyOrders.length} готово · {activeOrders.length} в пути</Text>
          </View>
        </View>
        <View style={{flexDirection: 'row', gap: 10}}>
          <TouchableOpacity onPress={handleRefresh} style={styles.iconBtn}>
            <RefreshCw color="#888" size={20} />
          </TouchableOpacity>
          <TouchableOpacity onPress={logout} style={styles.logoutBtn}>
            <Text style={styles.logoutText}>Выйти</Text>
          </TouchableOpacity>
        </View>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Ready Orders */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <View style={styles.sectionPill}>
              <Package size={16} color="#fff" />
              <Text style={styles.sectionPillText}>Доступно для взятия</Text>
            </View>
            <Text style={styles.sectionCount}>{readyOrders.length} шт</Text>
          </View>
          
          {readyOrders.length === 0 ? (
            <View style={styles.emptyState}>
              <Text style={styles.emptyEmoji}>📦</Text>
              <Text style={styles.emptyText}>Пока нет готовых заказов</Text>
            </View>
          ) : (
            readyOrders.map(renderReadyOrder)
          )}
        </View>

        {/* Active Orders */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <View style={[styles.sectionPill, { backgroundColor: 'rgba(249,115,22,0.15)' }]}>
              <Truck size={16} color="#f97316" />
              <Text style={[styles.sectionPillText, { color: '#f97316' }]}>Активные доставки</Text>
            </View>
            <Text style={styles.sectionCount}>{activeOrders.length} шт</Text>
          </View>
          
          {activeOrders.length === 0 ? (
            <View style={styles.emptyState}>
              <Text style={styles.emptyEmoji}>🚴</Text>
              <Text style={styles.emptyText}>Пока нет активных заказов</Text>
            </View>
          ) : (
            activeOrders.map(renderActiveOrder)
          )}
        </View>
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
  
  content: { flex: 1, padding: 15 },
  section: { marginBottom: 30 },
  sectionHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 15 },
  sectionPill: { flexDirection: 'row', alignItems: 'center', gap: 6, backgroundColor: '#333', paddingHorizontal: 10, paddingVertical: 5, borderRadius: 20 },
  sectionPillText: { color: '#fff', fontWeight: 'bold', fontSize: 12 },
  sectionCount: { color: '#888', fontWeight: 'bold' },
  
  emptyState: { padding: 30, alignItems: 'center', justifyContent: 'center', backgroundColor: '#1a1a1f', borderRadius: 12, borderWidth: 1, borderColor: '#222' },
  emptyEmoji: { fontSize: 40, opacity: 0.5, marginBottom: 10 },
  emptyText: { color: '#666', fontWeight: 'bold' },
  
  card: { backgroundColor: '#1a1a1f', padding: 15, borderRadius: 12, marginBottom: 15, borderWidth: 1, borderColor: '#222' },
  readyCard: { borderTopWidth: 3, borderTopColor: '#10b981' },
  activeCard: { borderTopWidth: 3, borderTopColor: '#f97316' },
  
  cardTop: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 },
  cardId: { color: '#f97316', fontSize: 20, fontWeight: 'bold' },
  cardPrice: { color: '#10b981', fontWeight: 'bold', fontSize: 16 },
  
  onWayBadge: { backgroundColor: 'rgba(249,115,22,0.15)', paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12, borderWidth: 1, borderColor: 'rgba(249,115,22,0.3)' },
  onWayText: { color: '#f97316', fontWeight: 'bold', fontSize: 12 },
  
  details: { gap: 8, marginBottom: 12 },
  infoRow: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 8 },
  infoText: { color: '#ccc', flex: 1 },
  
  itemsWrap: { flexDirection: 'row', flexWrap: 'wrap', gap: 8, marginBottom: 15 },
  itemChip: { backgroundColor: 'rgba(255,255,255,0.05)', paddingHorizontal: 8, paddingVertical: 4, borderRadius: 8, borderWidth: 1, borderColor: '#333' },
  itemChipText: { color: '#aaa', fontSize: 12 },
  
  btnPrimary: { backgroundColor: '#f97316', padding: 12, borderRadius: 8, flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: 8 },
  btnTextPrimary: { color: '#fff', fontWeight: 'bold', fontSize: 16 },
  
  btnSuccess: { backgroundColor: 'rgba(16,185,129,0.15)', borderWidth: 1, borderColor: 'rgba(16,185,129,0.3)', padding: 12, borderRadius: 8, flexDirection: 'row', justifyContent: 'center', alignItems: 'center', gap: 8 },
  btnTextSuccess: { color: '#34d399', fontWeight: 'bold', fontSize: 16 }
});

export default DeliveryScreen;

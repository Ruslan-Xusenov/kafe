import React, { useEffect, useState } from 'react';
import {
  View, Text, StyleSheet, FlatList,
  TouchableOpacity, ActivityIndicator, RefreshControl,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { ArrowLeft } from 'lucide-react-native';
import { useLangStore } from '../store/langStore';
import * as api from '../api';

const STATUS_COLOR = {
  new: '#f97316',
  preparing: '#f59e0b',
  ready: '#10b981',
  on_way: '#6366f1',
  delivered: '#22c55e',
  cancelled: '#ef4444',
};

const OrderHistoryScreen = ({ navigation }) => {
  const { t } = useLangStore();
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadHistory = async () => {
    try {
      const res = await api.getWaiterHistory();
      setOrders(res.data || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  useEffect(() => { loadHistory(); }, []);

  const statusLabel = (status) => {
    const map = {
      new: t.status_new,
      preparing: t.status_preparing,
      ready: t.status_ready,
      on_way: t.status_on_way,
      delivered: t.status_delivered,
      cancelled: t.status_cancelled,
    };
    return map[status] || status;
  };

  const formatDate = (str) => {
    if (!str) return '';
    const d = new Date(str);
    return `${d.getDate().toString().padStart(2,'0')}.${(d.getMonth()+1).toString().padStart(2,'0')}.${d.getFullYear()} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`;
  };

  const renderOrder = ({ item }) => {
    const color = STATUS_COLOR[item.status] || '#888';
    return (
      <View style={styles.orderCard}>
        <View style={styles.orderTop}>
          <View>
            <Text style={styles.orderId}>#{item.id}</Text>
            <Text style={styles.orderDate}>{formatDate(item.created_at)}</Text>
          </View>
          <View style={[styles.statusBadge, { backgroundColor: color + '22' }]}>
            <Text style={[styles.statusText, { color }]}>{statusLabel(item.status)}</Text>
          </View>
        </View>

        {item.table_name && (
          <Text style={styles.tableName}>🪑 {item.table_name}</Text>
        )}

        {/* Items preview */}
        {(item.items || []).slice(0, 3).map((oi, idx) => (
          <View key={idx} style={styles.itemRow}>
            <Text style={styles.itemName}>{oi.product_name}</Text>
            <Text style={styles.itemQty}>{oi.quantity} {oi.unit}</Text>
            <Text style={styles.itemPrice}>{(oi.price * oi.quantity).toLocaleString()} {t.sum}</Text>
          </View>
        ))}
        {(item.items || []).length > 3 && (
          <Text style={styles.moreItems}>+{item.items.length - 3} ta mahsulot...</Text>
        )}

        <View style={styles.orderBottom}>
          {item.payment_method && (
            <Text style={styles.payment}>💳 {item.payment_method}</Text>
          )}
          <Text style={styles.orderTotal}>{item.total_price?.toLocaleString()} {t.sum}</Text>
        </View>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.root}>
      <View style={styles.header}>
        <TouchableOpacity style={styles.backBtn} onPress={() => navigation.goBack()}>
          <ArrowLeft color="#fff" size={22} />
        </TouchableOpacity>
        <Text style={styles.headerTitle}>{t.ordersHistory}</Text>
      </View>

      {loading ? (
        <View style={styles.center}>
          <ActivityIndicator color="#f97316" size="large" />
        </View>
      ) : (
        <FlatList
          data={orders}
          keyExtractor={(item) => item.id.toString()}
          renderItem={renderOrder}
          contentContainerStyle={styles.list}
          refreshControl={
            <RefreshControl
              refreshing={refreshing}
              onRefresh={() => { setRefreshing(true); loadHistory(); }}
              tintColor="#f97316"
              colors={['#f97316']}
            />
          }
          ListEmptyComponent={
            <View style={styles.center}>
              <Text style={{ fontSize: 48 }}>📋</Text>
              <Text style={styles.emptyText}>{t.noHistory}</Text>
            </View>
          }
        />
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', paddingTop: 60 },

  header: {
    flexDirection: 'row', alignItems: 'center', gap: 14,
    paddingHorizontal: 16, paddingVertical: 14,
    borderBottomWidth: 1, borderBottomColor: '#1a1a22',
  },
  backBtn: {
    width: 38, height: 38, borderRadius: 19,
    backgroundColor: '#1a1a22', justifyContent: 'center', alignItems: 'center',
  },
  headerTitle: { color: '#fff', fontSize: 20, fontWeight: '800' },

  list: { padding: 16, paddingBottom: 40 },
  orderCard: {
    backgroundColor: '#141417', borderRadius: 16, padding: 16,
    marginBottom: 12, borderWidth: 1, borderColor: '#1e1e26',
  },
  orderTop: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 10 },
  orderId: { color: '#fff', fontSize: 17, fontWeight: '800' },
  orderDate: { color: '#555', fontSize: 12, marginTop: 2 },
  statusBadge: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 20, alignSelf: 'flex-start' },
  statusText: { fontWeight: '700', fontSize: 12 },
  tableName: { color: '#888', fontSize: 13, marginBottom: 10 },

  itemRow: { flexDirection: 'row', alignItems: 'center', paddingVertical: 4, borderTopWidth: 1, borderTopColor: '#1e1e26' },
  itemName: { flex: 1, color: '#ccc', fontSize: 13 },
  itemQty: { color: '#666', fontSize: 12, marginHorizontal: 8 },
  itemPrice: { color: '#f97316', fontWeight: '700', fontSize: 13 },
  moreItems: { color: '#555', fontSize: 12, marginTop: 4 },

  orderBottom: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: 12 },
  payment: { color: '#666', fontSize: 12 },
  orderTotal: { color: '#f97316', fontSize: 18, fontWeight: '900' },
  emptyText: { color: '#555', fontSize: 15, marginTop: 12 },
});

export default OrderHistoryScreen;

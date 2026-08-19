import React, { useState, useCallback } from 'react';
import { useFocusEffect } from '@react-navigation/native';
import {
  View, Text, StyleSheet, ScrollView, TouchableOpacity,
  ActivityIndicator, Alert, Modal, TextInput, FlatList,
  KeyboardAvoidingView, Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { ArrowLeft, Plus, Trash2, MoveRight, Percent, CreditCard } from 'lucide-react-native';
import { useLangStore } from '../store/langStore';
import { useWaiterStore } from '../store/waiterStore';
import * as api from '../api';

const PAYMENT_METHODS = ['cash', 'card', 'click', 'nasiya'];

const TableDetailScreen = ({ route, navigation }) => {
  const { t } = useLangStore();
  const { tables, fetchTables, fetchActiveOrder, activeOrder, clearActiveOrder } = useWaiterStore();
  const table = tables.find((t) => t.id === route.params.table?.id) || route.params.table;

  const [loading, setLoading] = useState(true);
  const [order, setOrder] = useState(null);

  // Modals
  const [closeModal, setCloseModal] = useState(false);
  const [transferModal, setTransferModal] = useState(false);
  const [serviceFeeModal, setServiceFeeModal] = useState(false);
  const [cancelItemModal, setCancelItemModal] = useState(null); // item object

  // Close table state
  const [payments, setPayments] = useState([{ method: 'cash', amount: '' }]);

  // Service fee state
  const [feePercent, setFeePercent] = useState('');

  // Transfer target table
  const [targetTable, setTargetTable] = useState(null);

  // Cancel item qty
  const [cancelQty, setCancelQty] = useState('');

  const loadOrder = useCallback(async () => {
    setLoading(true);
    try {
      const o = await fetchActiveOrder(table.id);
      setOrder(o);
    } finally {
      setLoading(false);
    }
  }, [table.id]);

  useFocusEffect(
    useCallback(() => {
      fetchTables();
      loadOrder();
      return () => clearActiveOrder();
    }, [table.id])
  );

  // ─── STATUS LABEL ──────────────────────────────────────────────────────────
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

  const statusColor = (status) => {
    if (status === 'new') return '#f97316';
    if (status === 'preparing') return '#f59e0b';
    if (status === 'ready') return '#10b981';
    if (status === 'delivered') return '#6366f1';
    if (status === 'cancelled') return '#ef4444';
    return '#888';
  };

  // ─── CLOSE TABLE (payment) ─────────────────────────────────────────────────
  const handleCloseTable = async () => {
    const totalPaid = payments.reduce((s, p) => s + (parseFloat(p.amount) || 0), 0);
    if (totalPaid <= 0) {
      Alert.alert(t.errorTitle, t.enterAmount);
      return;
    }

    try {
      const payload = {
        status: 'free',
        payments: payments
          .filter((p) => parseFloat(p.amount) > 0)
          .map((p) => ({ method: p.method, amount: parseFloat(p.amount) })),
      };
      await api.updateTable(table.id, payload);
      setCloseModal(false);
      await fetchTables();
      navigation.goBack();
    } catch (e) {
      Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
    }
  };

  // ─── SET SERVICE FEE ────────────────────────────────────────────────────────
  const handleSetServiceFee = async () => {
    const pct = parseFloat(feePercent);
    if (isNaN(pct) || pct < 0 || pct > 100) {
      Alert.alert(t.errorTitle, t.enterValidPercent);
      return;
    }
    try {
      const res = await api.setServiceFee(order.id, pct);
      setOrder(res.data);
      setServiceFeeModal(false);
      setFeePercent('');
    } catch (e) {
      Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
    }
  };

  // ─── TRANSFER TABLE ────────────────────────────────────────────────────────
  const handleTransfer = async () => {
    if (!targetTable) return;
    try {
      await api.transferOrderTable(table.id, targetTable.id);
      setTransferModal(false);
      await fetchTables();
      navigation.goBack();
    } catch (e) {
      Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
    }
  };

  // ─── CANCEL ITEM ────────────────────────────────────────────────────────────
  const handleCancelItem = async () => {
    const qty = parseFloat(cancelQty);
    if (!cancelItemModal || isNaN(qty) || qty <= 0) return;
    try {
      await api.cancelProductFromOrder(order.id, cancelItemModal.product_id, qty);
      setCancelItemModal(null);
      setCancelQty('');
      loadOrder();
    } catch (e) {
      Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
    }
  };

  // ─── FREE TABLE (no payment) ───────────────────────────────────────────────
  const handleFreeTable = () => {
    Alert.alert(t.confirm, t.closeTable + '?', [
      { text: t.cancel, style: 'cancel' },
      {
        text: t.confirm, style: 'destructive',
        onPress: async () => {
          try {
            await api.updateTable(table.id, { ...table, status: 'free' });
            await fetchTables();
            navigation.goBack();
          } catch (e) {
            Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
          }
        },
      },
    ]);
  };

  if (loading) {
    return (
      <SafeAreaView style={styles.root}>
        <View style={styles.center}><ActivityIndicator color="#f97316" size="large" /></View>
      </SafeAreaView>
    );
  }

  const isFree = table.status === 'free';

  return (
    <SafeAreaView style={styles.root}>
      {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity style={styles.backBtn} onPress={() => navigation.goBack()}>
          <ArrowLeft color="#fff" size={22} />
        </TouchableOpacity>
        <View style={{ flex: 1, marginLeft: 12 }}>
          <Text style={styles.headerTitle}>
            {t.tableDetail}: {table.name || `${t.table} ${table.id}`}
          </Text>
          <View style={[styles.statusBadge, { backgroundColor: isFree ? 'rgba(16,185,129,0.15)' : 'rgba(239,68,68,0.15)' }]}>
            <Text style={{ color: isFree ? '#10b981' : '#ef4444', fontWeight: '700', fontSize: 12 }}>
              {isFree ? t.free : t.occupied}
            </Text>
          </View>
        </View>
      </View>

      <ScrollView contentContainerStyle={{ padding: 16, paddingBottom: 40 }}>
        {/* Active Order */}
        {order ? (
          <>
            <View style={styles.orderCard}>
              <View style={styles.orderCardHeader}>
                <Text style={styles.orderTitle}>{t.activeOrder} #{order.id}</Text>
                <View style={[styles.orderStatus, { backgroundColor: statusColor(order.status) + '22' }]}>
                  <Text style={{ color: statusColor(order.status), fontWeight: '700', fontSize: 12 }}>
                    {statusLabel(order.status)}
                  </Text>
                </View>
              </View>

              {/* Items */}
              {(order.items || []).map((item) => (
                <View key={item.id} style={styles.itemRow}>
                  <View style={{ flex: 1 }}>
                    <Text style={styles.itemName}>{item.product_name}</Text>
                    <Text style={styles.itemMeta}>
                      {item.quantity} {item.unit} × {item.price.toLocaleString()} {t.sum}
                    </Text>
                    {item.comment ? <Text style={styles.itemComment}>💬 {item.comment}</Text> : null}
                  </View>
                  <View style={styles.itemRight}>
                    <Text style={styles.itemTotal}>
                      {(item.price * item.quantity).toLocaleString()} {t.sum}
                    </Text>
                    <TouchableOpacity
                      style={styles.cancelItemBtn}
                      onPress={() => { setCancelItemModal(item); setCancelQty(String(item.quantity)); }}
                    >
                      <Trash2 color="#ef4444" size={16} />
                    </TouchableOpacity>
                  </View>
                </View>
              ))}

              {/* Totals */}
              <View style={styles.separator} />
              {order.service_fee > 0 && (
                <>
                  <View style={styles.totalRow}>
                    <Text style={styles.totalLabel}>{t.subtotal}</Text>
                    <Text style={styles.totalValue}>
                      {(order.total_price - order.service_fee).toLocaleString()} {t.sum}
                    </Text>
                  </View>
                  <View style={styles.totalRow}>
                    <Text style={styles.totalLabel}>
                      {t.serviceFee} ({order.service_percentage}%)
                    </Text>
                    <Text style={styles.totalValue}>
                      {order.service_fee.toLocaleString()} {t.sum}
                    </Text>
                  </View>
                </>
              )}
              <View style={[styles.totalRow, { marginTop: 4 }]}>
                <Text style={styles.grandTotalLabel}>{t.total}</Text>
                <Text style={styles.grandTotal}>
                  {order.total_price.toLocaleString()} {t.sum}
                </Text>
              </View>
            </View>

            {/* Action Buttons */}
            <View style={styles.actionsGrid}>
              <TouchableOpacity
                style={styles.actionBtn}
                onPress={() => navigation.navigate('NewOrder', { table, existingOrderId: order.id })}
              >
                <Plus color="#f97316" size={20} />
                <Text style={styles.actionLabel}>{t.addItems}</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.actionBtn}
                onPress={() => { setFeePercent(String(order.service_percentage || '')); setServiceFeeModal(true); }}
              >
                <Percent color="#6366f1" size={20} />
                <Text style={styles.actionLabel}>{t.serviceFee}</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.actionBtn}
                onPress={() => setTransferModal(true)}
              >
                <MoveRight color="#10b981" size={20} />
                <Text style={styles.actionLabel}>{t.transferTable}</Text>
              </TouchableOpacity>

              <TouchableOpacity
                style={[styles.actionBtn, styles.actionBtnClose]}
                onPress={() => setCloseModal(true)}
              >
                <CreditCard color="#fff" size={20} />
                <Text style={[styles.actionLabel, { color: '#fff' }]}>{t.closeTable}</Text>
              </TouchableOpacity>
            </View>
          </>
        ) : (
          // No active order
          <View style={styles.noOrderCard}>
            <Text style={styles.noOrderEmoji}>🪑</Text>
            <Text style={styles.noOrderText}>{t.noActiveOrder}</Text>
            <TouchableOpacity
              style={styles.newOrderBtn}
              onPress={() => navigation.navigate('NewOrder', { table })}
            >
              <Plus color="#fff" size={18} />
              <Text style={styles.newOrderBtnText}>{t.newOrder}</Text>
            </TouchableOpacity>
          </View>
        )}
      </ScrollView>

      {/* ─── CLOSE TABLE MODAL ─────────────────────────────────────────── */}
      <Modal visible={closeModal} transparent animationType="slide" onRequestClose={() => setCloseModal(false)}>
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={styles.modalOverlay}>
          <View style={styles.modal}>
            <Text style={styles.modalTitle}>{t.closeTable}</Text>
            <Text style={styles.modalSub}>
              {t.total}: {order?.total_price?.toLocaleString()} {t.sum}
            </Text>

            {payments.map((p, i) => (
              <View key={i} style={styles.paymentRow}>
                <View style={styles.methodBtns}>
                  {PAYMENT_METHODS.map((m) => (
                    <TouchableOpacity
                      key={m}
                      style={[styles.methodBtn, p.method === m && styles.methodBtnActive]}
                      onPress={() => {
                        const next = [...payments];
                        next[i] = { ...next[i], method: m };
                        setPayments(next);
                      }}
                    >
                      <Text style={[styles.methodBtnText, p.method === m && styles.methodBtnTextActive]}>
                        {t[m]}
                      </Text>
                    </TouchableOpacity>
                  ))}
                </View>
                <View style={styles.amountRow}>
                  <TextInput
                    style={styles.amountInput}
                    value={p.amount}
                    onChangeText={(v) => {
                      const next = [...payments];
                      next[i] = { ...next[i], amount: v };
                      setPayments(next);
                    }}
                    keyboardType="numeric"
                    placeholder="Summa"
                    placeholderTextColor="#555"
                  />
                  {payments.length > 1 && (
                    <TouchableOpacity
                      onPress={() => setPayments(payments.filter((_, j) => j !== i))}
                      style={{ padding: 8 }}
                    >
                      <Trash2 color="#ef4444" size={18} />
                    </TouchableOpacity>
                  )}
                </View>
              </View>
            ))}

            <TouchableOpacity
              style={styles.addPaymentBtn}
              onPress={() => setPayments([...payments, { method: 'card', amount: '' }])}
            >
              <Plus color="#f97316" size={16} />
              <Text style={{ color: '#f97316', fontWeight: '700', marginLeft: 6 }}>{t.mixed}</Text>
            </TouchableOpacity>

            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.modalCancel} onPress={() => setCloseModal(false)}>
                <Text style={{ color: '#888', fontWeight: '700' }}>{t.cancel}</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.modalConfirm} onPress={handleCloseTable}>
                <Text style={{ color: '#fff', fontWeight: '700' }}>{t.confirm}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>

      {/* ─── SERVICE FEE MODAL ─────────────────────────────────────────── */}
      <Modal visible={serviceFeeModal} transparent animationType="fade" onRequestClose={() => setServiceFeeModal(false)}>
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={styles.modalOverlay}>
          <View style={[styles.modal, { maxHeight: 250 }]}>
            <Text style={styles.modalTitle}>{t.serviceFeeTitle}</Text>
            <TextInput
              style={styles.amountInput}
              value={feePercent}
              onChangeText={setFeePercent}
              keyboardType="numeric"
              placeholder="0"
              placeholderTextColor="#555"
              autoFocus
            />
            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.modalCancel} onPress={() => setServiceFeeModal(false)}>
                <Text style={{ color: '#888', fontWeight: '700' }}>{t.cancel}</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.modalConfirm} onPress={handleSetServiceFee}>
                <Text style={{ color: '#fff', fontWeight: '700' }}>{t.serviceFeeSet}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>

      {/* ─── TRANSFER TABLE MODAL ──────────────────────────────────────── */}
      <Modal visible={transferModal} transparent animationType="slide" onRequestClose={() => setTransferModal(false)}>
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={styles.modalOverlay}>
          <View style={[styles.modal, { maxHeight: '60%' }]}>
            <Text style={styles.modalTitle}>{t.transferTitle}</Text>
            <Text style={styles.modalSub}>{t.selectTargetTable}</Text>
            <FlatList
              data={tables.filter((tb) => tb.id !== table.id && tb.status === 'free')}
              keyExtractor={(item) => item.id.toString()}
              renderItem={({ item }) => (
                <TouchableOpacity
                  style={[
                    styles.transferTableItem,
                    targetTable?.id === item.id && styles.transferTableItemActive,
                  ]}
                  onPress={() => setTargetTable(item)}
                >
                  <Text style={{
                    color: targetTable?.id === item.id ? '#f97316' : '#fff',
                    fontWeight: '700',
                  }}>
                    {item.name || `${t.table} ${item.id}`}
                  </Text>
                </TouchableOpacity>
              )}
              style={{ maxHeight: 200 }}
            />
            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.modalCancel} onPress={() => setTransferModal(false)}>
                <Text style={{ color: '#888', fontWeight: '700' }}>{t.cancel}</Text>
              </TouchableOpacity>
              <TouchableOpacity style={styles.modalConfirm} onPress={handleTransfer}>
                <Text style={{ color: '#fff', fontWeight: '700' }}>{t.confirm}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>

      {/* ─── CANCEL ITEM MODAL ─────────────────────────────────────────── */}
      <Modal visible={!!cancelItemModal} transparent animationType="fade" onRequestClose={() => setCancelItemModal(null)}>
        <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={styles.modalOverlay}>
          <View style={[styles.modal, { maxHeight: 280 }]}>
            <Text style={styles.modalTitle}>{t.cancelItem}</Text>
            <Text style={styles.modalSub}>{cancelItemModal?.product_name}</Text>
            <TextInput
              style={styles.amountInput}
              value={cancelQty}
              onChangeText={setCancelQty}
              keyboardType="numeric"
              placeholder={t.cancelQty}
              placeholderTextColor="#555"
              autoFocus
            />
            <View style={styles.modalActions}>
              <TouchableOpacity style={styles.modalCancel} onPress={() => setCancelItemModal(null)}>
                <Text style={{ color: '#888', fontWeight: '700' }}>{t.cancel}</Text>
              </TouchableOpacity>
              <TouchableOpacity style={[styles.modalConfirm, { backgroundColor: '#ef4444' }]} onPress={handleCancelItem}>
                <Text style={{ color: '#fff', fontWeight: '700' }}>{t.cancel}</Text>
              </TouchableOpacity>
            </View>
          </View>
        </KeyboardAvoidingView>
      </Modal>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center' },

  header: {
    flexDirection: 'row', alignItems: 'center',
    paddingHorizontal: 16, paddingVertical: 14,
    borderBottomWidth: 1, borderBottomColor: '#1a1a22',
  },
  backBtn: {
    width: 38, height: 38, borderRadius: 19,
    backgroundColor: '#1a1a22', justifyContent: 'center', alignItems: 'center',
  },
  headerTitle: { color: '#fff', fontSize: 18, fontWeight: '800' },
  statusBadge: { alignSelf: 'flex-start', paddingHorizontal: 10, paddingVertical: 3, borderRadius: 12, marginTop: 4 },

  orderCard: {
    backgroundColor: '#141417', borderRadius: 18, padding: 18,
    borderWidth: 1, borderColor: '#222', marginBottom: 16,
  },
  orderCardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 },
  orderTitle: { color: '#fff', fontSize: 17, fontWeight: '800' },
  orderStatus: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 12 },

  itemRow: {
    flexDirection: 'row', alignItems: 'flex-start',
    paddingVertical: 10, borderTopWidth: 1, borderTopColor: '#1e1e26',
  },
  itemName: { color: '#fff', fontWeight: '700', fontSize: 14 },
  itemMeta: { color: '#777', fontSize: 13, marginTop: 2 },
  itemComment: { color: '#555', fontSize: 12, marginTop: 2 },
  itemRight: { alignItems: 'flex-end', gap: 6 },
  itemTotal: { color: '#f97316', fontWeight: '800', fontSize: 14 },
  cancelItemBtn: {
    width: 32, height: 32, borderRadius: 16,
    backgroundColor: 'rgba(239,68,68,0.12)', justifyContent: 'center', alignItems: 'center',
  },

  separator: { height: 1, backgroundColor: '#222', marginVertical: 12 },
  totalRow: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 4 },
  totalLabel: { color: '#888', fontSize: 14 },
  totalValue: { color: '#ccc', fontSize: 14, fontWeight: '600' },
  grandTotalLabel: { color: '#fff', fontSize: 17, fontWeight: '800' },
  grandTotal: { color: '#f97316', fontSize: 20, fontWeight: '900' },

  actionsGrid: { flexDirection: 'row', flexWrap: 'wrap', gap: 10 },
  actionBtn: {
    flex: 1, minWidth: '45%', backgroundColor: '#141417',
    borderRadius: 14, padding: 16, alignItems: 'center', gap: 8,
    borderWidth: 1, borderColor: '#222',
  },
  actionBtnClose: { backgroundColor: '#f97316', borderColor: '#f97316', flexBasis: '100%', flex: 0, flexDirection: 'row' },
  actionLabel: { color: '#ccc', fontSize: 13, fontWeight: '700' },

  noOrderCard: {
    backgroundColor: '#141417', borderRadius: 18, padding: 40,
    alignItems: 'center', borderWidth: 1, borderColor: '#222',
  },
  noOrderEmoji: { fontSize: 48, marginBottom: 12 },
  noOrderText: { color: '#666', fontSize: 15, marginBottom: 24 },
  newOrderBtn: {
    flexDirection: 'row', gap: 8, alignItems: 'center',
    backgroundColor: '#f97316', paddingHorizontal: 24, paddingVertical: 12, borderRadius: 14,
  },
  newOrderBtnText: { color: '#fff', fontWeight: '800', fontSize: 15 },

  // Modals
  modalOverlay: {
    flex: 1, backgroundColor: 'rgba(0,0,0,0.75)',
    justifyContent: 'flex-end',
  },
  modal: {
    backgroundColor: '#141417', borderTopLeftRadius: 24, borderTopRightRadius: 24,
    padding: 24, borderTopWidth: 1, borderColor: '#222',
  },
  modalTitle: { color: '#fff', fontSize: 20, fontWeight: '800', marginBottom: 6 },
  modalSub: { color: '#888', fontSize: 14, marginBottom: 16 },

  paymentRow: { marginBottom: 14 },
  methodBtns: { flexDirection: 'row', flexWrap: 'wrap', gap: 6, marginBottom: 10 },
  methodBtn: {
    paddingHorizontal: 12, paddingVertical: 6, borderRadius: 10,
    backgroundColor: '#1e1e26', borderWidth: 1, borderColor: '#333',
  },
  methodBtnActive: { backgroundColor: '#f97316', borderColor: '#f97316' },
  methodBtnText: { color: '#888', fontWeight: '700', fontSize: 12 },
  methodBtnTextActive: { color: '#fff' },
  amountRow: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  amountInput: {
    flex: 1, backgroundColor: '#1a1a22', color: '#fff',
    padding: 13, borderRadius: 12, borderWidth: 1, borderColor: '#2a2a35', fontSize: 15,
  },
  addPaymentBtn: {
    flexDirection: 'row', alignItems: 'center', justifyContent: 'center',
    padding: 10, borderRadius: 10, borderWidth: 1, borderColor: 'rgba(249,115,22,0.3)',
    marginBottom: 16,
  },

  modalActions: { flexDirection: 'row', gap: 10, marginTop: 16 },
  modalCancel: {
    flex: 1, padding: 14, borderRadius: 12, backgroundColor: '#1e1e26',
    alignItems: 'center', borderWidth: 1, borderColor: '#333',
  },
  modalConfirm: {
    flex: 1, padding: 14, borderRadius: 12, backgroundColor: '#f97316', alignItems: 'center',
  },

  transferTableItem: {
    padding: 14, borderRadius: 12, backgroundColor: '#1e1e26',
    marginBottom: 8, borderWidth: 1, borderColor: '#2a2a35',
  },
  transferTableItemActive: { borderColor: '#f97316', backgroundColor: 'rgba(249,115,22,0.1)' },
});

export default TableDetailScreen;

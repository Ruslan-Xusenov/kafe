import React, { useState, useEffect, useMemo } from 'react';
import {
  View, Text, StyleSheet, FlatList, TouchableOpacity,
  SafeAreaView, TextInput, Alert, ActivityIndicator, Image,
  KeyboardAvoidingView, Platform,
} from 'react-native';
import { ArrowLeft, Search, Minus, Plus, ShoppingCart, Send } from 'lucide-react-native';
import { useLangStore } from '../store/langStore';
import { useWaiterStore } from '../store/waiterStore';
import { useCartStore } from '../store/cartStore';
import * as api from '../api';

const NewOrderScreen = ({ route, navigation }) => {
  const { table, existingOrderId } = route.params;
  const { t } = useLangStore();
  const { categories, products, loadingMenu, fetchMenu } = useWaiterStore();

  const [selectedCat, setSelectedCat] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const { items: cart, addItem, updateQuantity, removeItem, clearCart, getTotalPrice } = useCartStore();
  const [sending, setSending] = useState(false);
  const [showCart, setShowCart] = useState(false);

  useEffect(() => {
    if (products.length === 0) fetchMenu();
  }, []);

  // ─── Filtered products ─────────────────────────────────────────────────────
  const filtered = useMemo(() => {
    let list = products;
    if (selectedCat) list = list.filter((p) => p.category_id === selectedCat);
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      list = list.filter((p) => p.name.toLowerCase().includes(q));
    }
    return list;
  }, [products, selectedCat, searchQuery]);

  // ─── Cart helpers ───────────────────────────────────────────────────────────
  const cartCount = cart.reduce((s, i) => s + i.quantity, 0);
  const cartTotal = getTotalPrice();

  const getCartItem = (productId) => cart.find((i) => i.product_id === productId);

  const addToCart = (product) => addItem(product);
  const removeFromCart = (product) => updateQuantity(product.id, -(product.quantity_step || 1));
  const removeAll = (productId) => removeItem(productId);

  // ─── Submit order ───────────────────────────────────────────────────────────
  const handleSubmit = async () => {
    if (cart.length === 0) return;
    setSending(true);
    try {
      if (existingOrderId) {
        // Add to existing order
        await api.addItemsToOrder(existingOrderId, cart.map((i) => ({
          product_id: i.product_id,
          quantity: i.quantity,
          price: i.price,
          unit: i.unit,
        })));
      } else {
        // Create new order
        const totalPrice = cart.reduce((s, i) => s + i.price * i.quantity, 0);
        await api.createOrder({
          table_id: table.id,
          items: cart.map((i) => ({
            product_id: i.product_id,
            quantity: i.quantity,
            price: i.price,
            unit: i.unit,
          })),
          total_price: totalPrice,
          address: table.name || `Stol ${table.id}`,
          phone: 'ofitsant',
          delivery_method: 'dine_in',
        });
      }

      Alert.alert(t.orderSent, t.orderSentMsg, [
        { text: 'OK', onPress: () => { clearCart(); navigation.goBack(); } },
      ]);
    } catch (e) {
      Alert.alert(t.errorTitle, e.response?.data?.error || e.message);
    } finally {
      setSending(false);
    }
  };

  // ─── Render product card ───────────────────────────────────────────────────
  const renderProduct = ({ item }) => {
    const inCart = getCartItem(item.id);
    return (
      <View style={styles.productCard}>
        {item.image_url ? (
          <Image source={{ uri: item.image_url }} style={styles.productImg} />
        ) : (
          <View style={[styles.productImg, styles.productImgPlaceholder]}>
            <Text style={{ fontSize: 24 }}>🍽️</Text>
          </View>
        )}
        <View style={styles.productInfo}>
          <Text style={styles.productName} numberOfLines={2}>{item.name}</Text>
          <Text style={styles.productPrice}>{item.price.toLocaleString()} {t.sum}</Text>
          {item.unit && <Text style={styles.productUnit}>{item.unit}</Text>}
        </View>
        <View style={styles.productActions}>
          {inCart ? (
            <View style={styles.qtyControl}>
              <TouchableOpacity
                style={styles.qtyBtn}
                onPress={() => removeFromCart(item)}
              >
                <Minus size={14} color="#fff" />
              </TouchableOpacity>
              <Text style={styles.qtyText}>{inCart.quantity}</Text>
              <TouchableOpacity
                style={styles.qtyBtn}
                onPress={() => addToCart(item)}
              >
                <Plus size={14} color="#fff" />
              </TouchableOpacity>
            </View>
          ) : (
            <TouchableOpacity style={styles.addBtn} onPress={() => addToCart(item)}>
              <Plus size={18} color="#fff" />
            </TouchableOpacity>
          )}
        </View>
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.root}>
      <KeyboardAvoidingView 
        style={{ flex: 1 }} 
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        {/* Header */}
      <View style={styles.header}>
        <TouchableOpacity style={styles.backBtn} onPress={() => navigation.goBack()}>
          <ArrowLeft color="#fff" size={22} />
        </TouchableOpacity>
        <View style={{ flex: 1, marginHorizontal: 12 }}>
          <Text style={styles.headerTitle}>{t.newOrder}</Text>
          <Text style={styles.headerSub}>{table.name || `${t.table} ${table.id}`}</Text>
        </View>
        <TouchableOpacity style={styles.cartToggle} onPress={() => setShowCart(!showCart)}>
          <ShoppingCart color={cartCount > 0 ? '#f97316' : '#888'} size={22} />
          {cartCount > 0 && (
            <View style={styles.cartBadge}>
              <Text style={styles.cartBadgeText}>{cartCount}</Text>
            </View>
          )}
        </TouchableOpacity>
      </View>

      {/* Search */}
      <View style={styles.searchRow}>
        <Search color="#555" size={16} style={{ marginLeft: 12 }} />
        <TextInput
          style={styles.searchInput}
          value={searchQuery}
          onChangeText={setSearchQuery}
          placeholder={t.search}
          placeholderTextColor="#555"
        />
      </View>

      {/* Categories */}
      <FlatList
        data={[{ id: null, name: t.allCategories }, ...categories]}
        keyExtractor={(item) => String(item.id)}
        renderItem={({ item }) => (
          <TouchableOpacity
            style={[styles.catChip, selectedCat === item.id && styles.catChipActive]}
            onPress={() => setSelectedCat(item.id)}
          >
            <Text style={[styles.catChipText, selectedCat === item.id && styles.catChipTextActive]}>
              {item.name}
            </Text>
          </TouchableOpacity>
        )}
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={styles.catList}
        style={styles.catScroll}
      />

      {/* Products or Cart view */}
      {showCart ? (
        <FlatList
          data={cart}
          keyExtractor={(item) => item.product_id.toString()}
          contentContainerStyle={{ padding: 16, paddingBottom: 120 }}
          ListEmptyComponent={
            <View style={styles.center}>
              <Text style={{ fontSize: 40 }}>🛒</Text>
              <Text style={styles.emptyText}>{t.cartEmpty}</Text>
            </View>
          }
          renderItem={({ item }) => (
            <View style={styles.cartItem}>
              <View style={{ flex: 1 }}>
                <Text style={styles.cartItemName}>{item.name}</Text>
                <Text style={styles.cartItemPrice}>{item.price.toLocaleString()} {t.sum}</Text>
              </View>
              <View style={styles.qtyControl}>
                <TouchableOpacity
                  style={styles.qtyBtn}
                  onPress={() => removeAll(item.product_id)}
                >
                  <Minus size={14} color="#fff" />
                </TouchableOpacity>
                <Text style={styles.qtyText}>{item.quantity}</Text>
                <TouchableOpacity
                  style={styles.qtyBtn}
                  onPress={() => addToCart({ id: item.product_id, price: item.price, unit: item.unit, quantity_step: item.step, min_quantity: item.step })}
                >
                  <Plus size={14} color="#fff" />
                </TouchableOpacity>
              </View>
              <Text style={styles.cartItemTotal}>
                {(item.price * item.quantity).toLocaleString()}
              </Text>
            </View>
          )}
        />
      ) : (
        loadingMenu && filtered.length === 0 ? (
          <View style={styles.center}>
            <ActivityIndicator color="#f97316" size="large" />
          </View>
        ) : (
          <FlatList
            data={filtered}
            renderItem={renderProduct}
            keyExtractor={(p) => p.id.toString()}
            contentContainerStyle={styles.productList}
            ListEmptyComponent={
              <View style={styles.center}>
                <Text style={{ color: '#555' }}>🔍 {searchQuery}</Text>
              </View>
            }
          />
        )
      )}

      {/* Submit bar */}
      {cart.length > 0 && (
        <View style={styles.submitBar}>
          <View>
            <Text style={{ color: '#aaa', fontSize: 12 }}>{t.orderTotal}</Text>
            <Text style={styles.submitTotal}>{cartTotal.toLocaleString()} {t.sum}</Text>
          </View>
          <TouchableOpacity style={styles.submitBtn} onPress={handleSubmit} disabled={sending}>
            {sending ? (
              <ActivityIndicator color="#fff" />
            ) : (
              <>
                <Send size={18} color="#fff" />
                <Text style={styles.submitBtnText}>{t.sendOrder}</Text>
              </>
            )}
          </TouchableOpacity>
        </View>
      )}
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  root: { flex: 1, backgroundColor: '#0b0b0e' },
  center: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 40 },

  header: {
    flexDirection: 'row', alignItems: 'center',
    paddingHorizontal: 16, paddingVertical: 14,
    borderBottomWidth: 1, borderBottomColor: '#1a1a22',
  },
  backBtn: {
    width: 38, height: 38, borderRadius: 19,
    backgroundColor: '#1a1a22', justifyContent: 'center', alignItems: 'center',
  },
  headerTitle: { color: '#fff', fontSize: 17, fontWeight: '800' },
  headerSub: { color: '#888', fontSize: 13 },
  cartToggle: { position: 'relative', padding: 8 },
  cartBadge: {
    position: 'absolute', top: 0, right: 0,
    width: 18, height: 18, borderRadius: 9,
    backgroundColor: '#f97316', justifyContent: 'center', alignItems: 'center',
  },
  cartBadgeText: { color: '#fff', fontSize: 10, fontWeight: '900' },

  searchRow: {
    flexDirection: 'row', alignItems: 'center',
    backgroundColor: '#141417', margin: 12, borderRadius: 12,
    borderWidth: 1, borderColor: '#222',
  },
  searchInput: { flex: 1, color: '#fff', padding: 12, fontSize: 14 },

  catScroll: { maxHeight: 44 },
  catList: { paddingHorizontal: 12, gap: 8, alignItems: 'center' },
  catChip: {
    paddingHorizontal: 14, paddingVertical: 7, borderRadius: 20,
    backgroundColor: '#1a1a22', borderWidth: 1, borderColor: '#2a2a35',
  },
  catChipActive: { backgroundColor: '#f97316', borderColor: '#f97316' },
  catChipText: { color: '#888', fontWeight: '700', fontSize: 12 },
  catChipTextActive: { color: '#fff' },

  productList: { padding: 12, paddingBottom: 120 },
  productCard: {
    flexDirection: 'row', alignItems: 'center',
    backgroundColor: '#141417', borderRadius: 14, marginBottom: 10, padding: 12,
    borderWidth: 1, borderColor: '#1e1e26',
  },
  productImg: { width: 56, height: 56, borderRadius: 10 },
  productImgPlaceholder: {
    backgroundColor: '#1e1e26', justifyContent: 'center', alignItems: 'center',
  },
  productInfo: { flex: 1, marginHorizontal: 12 },
  productName: { color: '#fff', fontWeight: '700', fontSize: 14 },
  productPrice: { color: '#f97316', fontWeight: '800', fontSize: 13, marginTop: 4 },
  productUnit: { color: '#666', fontSize: 11, marginTop: 2 },
  productActions: { alignItems: 'flex-end' },
  addBtn: {
    width: 36, height: 36, borderRadius: 18,
    backgroundColor: '#f97316', justifyContent: 'center', alignItems: 'center',
  },
  qtyControl: { flexDirection: 'row', alignItems: 'center', gap: 8 },
  qtyBtn: {
    width: 28, height: 28, borderRadius: 14,
    backgroundColor: '#2a2a35', justifyContent: 'center', alignItems: 'center',
  },
  qtyText: { color: '#fff', fontWeight: '800', minWidth: 28, textAlign: 'center', fontSize: 14 },

  cartItem: {
    flexDirection: 'row', alignItems: 'center', gap: 10,
    backgroundColor: '#141417', borderRadius: 14, padding: 14, marginBottom: 10,
    borderWidth: 1, borderColor: '#1e1e26',
  },
  cartItemName: { color: '#fff', fontWeight: '700', fontSize: 14 },
  cartItemPrice: { color: '#888', fontSize: 12, marginTop: 2 },
  cartItemTotal: { color: '#f97316', fontWeight: '800', fontSize: 14, minWidth: 60, textAlign: 'right' },
  emptyText: { color: '#555', marginTop: 12, fontSize: 15 },

  submitBar: {
    position: 'absolute', bottom: 0, left: 0, right: 0,
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    backgroundColor: '#141417', padding: 20,
    borderTopWidth: 1, borderTopColor: '#222',
  },
  submitTotal: { color: '#fff', fontSize: 20, fontWeight: '900' },
  submitBtn: {
    flexDirection: 'row', gap: 8, alignItems: 'center',
    backgroundColor: '#f97316', paddingHorizontal: 22, paddingVertical: 14, borderRadius: 14,
    shadowColor: '#f97316', shadowOffset: { width: 0, height: 4 }, shadowOpacity: 0.4, elevation: 6,
  },
  submitBtnText: { color: '#fff', fontWeight: '800', fontSize: 15 },
});

export default NewOrderScreen;

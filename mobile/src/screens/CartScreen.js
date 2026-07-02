import React from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, SafeAreaView, Alert } from 'react-native';
import { useCartStore } from '../store/cartStore';
import { Plus, Minus, Trash2, ShoppingCart } from 'lucide-react-native';
import api from '../api';

const CartScreen = ({ navigation }) => {
  const { items, updateQuantity, clearCart, getTotalPrice } = useCartStore();
  const totalPrice = getTotalPrice();

  const handleCheckout = async () => {
    if (items.length === 0) return;
    try {
      const payload = {
        items: items.map(i => ({
          product_id: i.product_id,
          quantity: i.quantity,
          price: i.price,
          unit: i.unit
        })),
        total_price: totalPrice,
        address: 'Mobile App',
        phone: 'Belgilanmagan', // We'll need a real modal for this later
        delivery_method: 'pickup'
      };

      await api.post('/orders', payload);
      Alert.alert('Muvaffaqiyatli', 'Buyurtma oshxonaga yuborildi!');
      clearCart();
      navigation.navigate('Home');
    } catch (err) {
      console.error(err);
      Alert.alert('Xatolik', 'Buyurtma yuborishda xatolik yuz berdi.');
    }
  };

  const renderItem = ({ item }) => (
    <View style={styles.cartItem}>
      <View style={styles.itemInfo}>
        <Text style={styles.itemName}>{item.name}</Text>
        <Text style={styles.itemPrice}>{(item.price * item.quantity).toLocaleString()} so'm</Text>
      </View>
      <View style={styles.qtyControls}>
        <TouchableOpacity style={styles.qtyBtn} onPress={() => updateQuantity(item.product_id, -item.step)}>
          <Minus size={16} color="#fff" />
        </TouchableOpacity>
        <Text style={styles.qtyText}>{item.quantity} {item.unit}</Text>
        <TouchableOpacity style={styles.qtyBtn} onPress={() => updateQuantity(item.product_id, item.step)}>
          <Plus size={16} color="#fff" />
        </TouchableOpacity>
      </View>
    </View>
  );

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Savatcha</Text>
        {items.length > 0 && (
          <TouchableOpacity onPress={clearCart}>
            <Trash2 color="#ef4444" size={24} />
          </TouchableOpacity>
        )}
      </View>

      {items.length === 0 ? (
        <View style={styles.emptyState}>
          <ShoppingCart color="#333" size={64} style={{ marginBottom: 20 }} />
          <Text style={styles.emptyText}>Savatchangiz bo'sh</Text>
        </View>
      ) : (
        <>
          <FlatList
            data={items}
            renderItem={renderItem}
            keyExtractor={item => item.product_id.toString()}
            contentContainerStyle={styles.list}
          />
          <View style={styles.footer}>
            <View style={styles.totalRow}>
              <Text style={styles.totalLabel}>Jami summa:</Text>
              <Text style={styles.totalPrice}>{totalPrice.toLocaleString()} so'm</Text>
            </View>
            <TouchableOpacity style={styles.checkoutBtn} onPress={handleCheckout}>
              <Text style={styles.checkoutBtnText}>Buyurtma Berish</Text>
            </TouchableOpacity>
          </View>
        </>
      )}
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f' },
  header: { padding: 20, paddingTop: 40, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  title: { color: '#fff', fontSize: 24, fontWeight: 'bold' },
  emptyState: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  emptyText: { color: '#666', fontSize: 18 },
  list: { padding: 20 },
  cartItem: {
    backgroundColor: '#1a1a1f', padding: 15, borderRadius: 12,
    flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center',
    marginBottom: 15, borderWidth: 1, borderColor: '#222'
  },
  itemInfo: { flex: 1 },
  itemName: { color: '#fff', fontSize: 16, fontWeight: 'bold', marginBottom: 4 },
  itemPrice: { color: '#f97316', fontSize: 14, fontWeight: '600' },
  qtyControls: { flexDirection: 'row', alignItems: 'center', gap: 10, backgroundColor: '#0d0d0f', padding: 5, borderRadius: 20 },
  qtyBtn: { width: 30, height: 30, borderRadius: 15, backgroundColor: '#333', justifyContent: 'center', alignItems: 'center' },
  qtyText: { color: '#fff', fontWeight: 'bold', minWidth: 30, textAlign: 'center' },
  footer: { padding: 20, backgroundColor: '#1a1a1f', borderTopWidth: 1, borderColor: '#222' },
  totalRow: { flexDirection: 'row', justifyContent: 'space-between', marginBottom: 15 },
  totalLabel: { color: '#888', fontSize: 16 },
  totalPrice: { color: '#fff', fontSize: 20, fontWeight: 'bold' },
  checkoutBtn: { backgroundColor: '#f97316', padding: 16, borderRadius: 12, alignItems: 'center' },
  checkoutBtnText: { color: '#fff', fontSize: 16, fontWeight: 'bold' }
});

export default CartScreen;

import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, SafeAreaView, ActivityIndicator, Image, Alert } from 'react-native';
import { useAuthStore } from '../store/authStore';
import api from '../api';
import { LogOut, ArrowLeft, Plus, Minus, ShoppingCart, RefreshCw } from 'lucide-react-native';

const WaiterScreen = () => {
  const [tables, setTables] = useState([]);
  const [products, setProducts] = useState([]);
  const [selectedTable, setSelectedTable] = useState(null);
  const [cart, setCart] = useState([]);
  const [loading, setLoading] = useState(true);

  const logout = useAuthStore(s => s.logout);

  useEffect(() => {
    fetchInitialData();
  }, []);

  const fetchInitialData = async () => {
    setLoading(true);
    try {
      const [tRes, pRes] = await Promise.all([
        api.get(`/tables/?_t=${Date.now()}`),
        api.get('/catalog/products')
      ]);
      setTables(tRes.data || []);
      setProducts(pRes.data || []);
    } catch (err) {
      console.error(err);
      Alert.alert("Xato", "Ma'lumotlarni yuklashda xatolik: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const addToCart = (product) => {
    setCart(prev => {
      const existing = prev.find(i => i.product_id === product.id);
      if (existing) {
        return prev.map(i => i.product_id === product.id ? { ...i, quantity: i.quantity + product.quantity_step } : i);
      }
      return [...prev, { 
        product_id: product.id, 
        name: product.name, 
        price: product.price, 
        quantity: product.min_quantity || 1,
        unit: product.unit,
        step: product.quantity_step || 1
      }];
    });
  };

  const updateQuantity = (productId, delta) => {
    setCart(prev => {
      return prev.map(i => {
        if (i.product_id === productId) {
          const newQ = i.quantity + delta;
          return newQ > 0 ? { ...i, quantity: newQ } : i;
        }
        return i;
      }).filter(i => i.quantity > 0);
    });
  };

  const submitOrder = async () => {
    if (cart.length === 0) return Alert.alert("Xato", "Savatcha bo'sh!");
    
    const totalPrice = cart.reduce((acc, item) => acc + (item.price * item.quantity), 0);
    
    try {
      const payload = {
        table_id: selectedTable.id,
        items: cart.map(i => ({
          product_id: i.product_id,
          quantity: i.quantity,
          price: i.price,
          unit: i.unit
        })),
        total_price: totalPrice,
        address: `Stol: ${selectedTable.number}`,
        phone: 'Ichki buyurtma',
        delivery_method: 'dine_in'
      };

      await api.post('/orders', payload);
      
      if (selectedTable.status === 'free') {
        await api.put(`/tables/${selectedTable.id}`, { ...selectedTable, status: 'occupied' });
      }

      Alert.alert('Muvaffaqiyatli', 'Buyurtma oshxonaga yuborildi!');
      setSelectedTable(null);
      setCart([]);
      fetchInitialData();
    } catch (err) {
      Alert.alert('Xato', 'Buyurtma yuborishda xatolik');
    }
  };

  const freeTable = () => {
    Alert.alert(
      "Tasdiqlash",
      "Stolni bo'shatishni xohlaysizmi?",
      [
        { text: "Yo'q", style: "cancel" },
        { 
          text: "Ha", 
          onPress: async () => {
            try {
              await api.put(`/tables/${selectedTable.id}`, { ...selectedTable, status: 'free' });
              Alert.alert("Bajarildi", "Stol bo'shatildi!");
              setSelectedTable(null);
              fetchInitialData();
            } catch (err) {
              Alert.alert("Xato", "Xatolik yuz berdi");
            }
          }
        }
      ]
    );
  };

  const renderTable = ({ item }) => {
    const isFree = item.status === 'free';
    return (
      <TouchableOpacity 
        style={[styles.tableCard, isFree ? styles.tableFree : styles.tableOccupied]}
        onPress={() => setSelectedTable(item)}
      >
        <Text style={[styles.tableNum, isFree ? styles.textFree : styles.textOccupied]}>
          Stol {item.number}
        </Text>
        <Text style={[styles.tableStatus, isFree ? styles.textFree : styles.textOccupied]}>
          {isFree ? "Bo'sh" : "Band"}
        </Text>
      </TouchableOpacity>
    );
  };

  const renderProduct = ({ item }) => {
    const inCart = cart.find(c => c.product_id === item.id);
    return (
      <View style={styles.productCard}>
        <View style={{flex: 1}}>
          <Text style={styles.productName}>{item.name}</Text>
          <Text style={styles.productPrice}>{item.price.toLocaleString()} so'm</Text>
        </View>
        
        {inCart ? (
          <View style={styles.qtyWrap}>
            <TouchableOpacity style={styles.qtyBtn} onPress={() => updateQuantity(item.id, -item.quantity_step)}>
              <Minus size={16} color="#fff" />
            </TouchableOpacity>
            <Text style={styles.qtyText}>{inCart.quantity} {item.unit}</Text>
            <TouchableOpacity style={styles.qtyBtn} onPress={() => updateQuantity(item.id, item.quantity_step)}>
              <Plus size={16} color="#fff" />
            </TouchableOpacity>
          </View>
        ) : (
          <TouchableOpacity style={styles.addBtn} onPress={() => addToCart(item)}>
            <Plus size={20} color="#fff" />
          </TouchableOpacity>
        )}
      </View>
    );
  };

  if (loading) return (
    <View style={styles.center}>
      <ActivityIndicator size="large" color="#f97316" />
    </View>
  );

  const cartTotal = cart.reduce((acc, item) => acc + (item.price * item.quantity), 0);

  if (selectedTable) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.orderHeader}>
          <TouchableOpacity onPress={() => { setSelectedTable(null); setCart([]); }} style={{padding: 10}}>
            <ArrowLeft color="#fff" size={24} />
          </TouchableOpacity>
          <View style={{alignItems: 'center'}}>
            <Text style={styles.orderTitle}>Stol № {selectedTable.number}</Text>
            <Text style={{color: selectedTable.status === 'free' ? '#10b981' : '#ef4444', fontWeight: 'bold'}}>
              {selectedTable.status === 'free' ? "Bo'sh" : "Band"}
            </Text>
          </View>
          {selectedTable.status !== 'free' ? (
            <TouchableOpacity onPress={freeTable} style={styles.freeBtn}>
              <Text style={styles.freeBtnText}>Bo'shatish</Text>
            </TouchableOpacity>
          ) : (
            <View style={{width: 60}} /> // Placeholder for alignment
          )}
        </View>

        <FlatList
          key="products-list"
          data={products.filter(p => p.is_active)}
          renderItem={renderProduct}
          keyExtractor={p => p.id.toString()}
          contentContainerStyle={{padding: 15, paddingBottom: 120}}
        />

        {cart.length > 0 && (
          <View style={styles.cartPanel}>
            <View>
              <Text style={{color: '#fff', fontSize: 12}}>Jami summa</Text>
              <Text style={{color: '#fff', fontSize: 20, fontWeight: 'bold'}}>{cartTotal.toLocaleString()} so'm</Text>
            </View>
            <TouchableOpacity style={styles.submitBtn} onPress={submitOrder}>
              <Text style={{color: '#fff', fontWeight: 'bold', fontSize: 16}}>Yuborish</Text>
            </TouchableOpacity>
          </View>
        )}
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.headerTitle}>Ofitsant Paneli</Text>
        <View style={{flexDirection: 'row', gap: 10}}>
          <TouchableOpacity onPress={fetchInitialData} style={styles.iconBtn}>
            <RefreshCw color="#888" size={20} />
          </TouchableOpacity>
          <TouchableOpacity onPress={logout} style={styles.logoutBtn}>
            <LogOut color="#ef4444" size={20} />
          </TouchableOpacity>
        </View>
      </View>
      
      <FlatList
        key="tables-list"
        data={tables}
        renderItem={renderTable}
        keyExtractor={t => t.id.toString()}
        numColumns={2}
        contentContainerStyle={{padding: 15}}
        columnWrapperStyle={{justifyContent: 'space-between'}}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f' },
  center: { flex: 1, backgroundColor: '#0d0d0f', justifyContent: 'center', alignItems: 'center' },
  
  header: { padding: 20, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', borderBottomWidth: 1, borderBottomColor: '#222' },
  headerTitle: { color: '#fff', fontSize: 24, fontWeight: 'bold' },
  iconBtn: { padding: 8, backgroundColor: '#1a1a1f', borderRadius: 8 },
  logoutBtn: { padding: 8, backgroundColor: 'rgba(239,68,68,0.15)', borderRadius: 8 },
  
  tableCard: { width: '48%', padding: 20, borderRadius: 12, marginBottom: 15, alignItems: 'center', justifyContent: 'center', borderWidth: 2 },
  tableFree: { backgroundColor: 'rgba(16,185,129,0.1)', borderColor: 'rgba(16,185,129,0.3)' },
  tableOccupied: { backgroundColor: 'rgba(239,68,68,0.1)', borderColor: 'rgba(239,68,68,0.3)' },
  tableNum: { fontSize: 24, fontWeight: 'bold', marginBottom: 5 },
  tableStatus: { fontSize: 14, fontWeight: 'bold', textTransform: 'uppercase' },
  textFree: { color: '#10b981' },
  textOccupied: { color: '#ef4444' },

  orderHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', padding: 15, backgroundColor: '#1a1a1f', borderBottomWidth: 1, borderColor: '#333' },
  orderTitle: { color: '#fff', fontSize: 18, fontWeight: 'bold' },
  freeBtn: { backgroundColor: 'rgba(16,185,129,0.2)', paddingHorizontal: 12, paddingVertical: 6, borderRadius: 8, borderWidth: 1, borderColor: '#10b981' },
  freeBtnText: { color: '#10b981', fontWeight: 'bold', fontSize: 12 },

  productCard: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', backgroundColor: '#1a1a1f', padding: 15, borderRadius: 12, marginBottom: 10, borderWidth: 1, borderColor: '#222' },
  productName: { color: '#fff', fontSize: 16, fontWeight: 'bold', marginBottom: 4 },
  productPrice: { color: '#f97316', fontWeight: 'bold' },
  addBtn: { backgroundColor: '#333', padding: 10, borderRadius: 8 },
  
  qtyWrap: { flexDirection: 'row', alignItems: 'center', gap: 10, backgroundColor: '#0d0d0f', padding: 4, borderRadius: 8 },
  qtyBtn: { padding: 6, backgroundColor: '#333', borderRadius: 6 },
  qtyText: { color: '#fff', fontWeight: 'bold', minWidth: 25, textAlign: 'center' },

  cartPanel: { position: 'absolute', bottom: 20, left: 15, right: 15, backgroundColor: '#f97316', borderRadius: 16, padding: 20, flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', shadowColor: '#f97316', shadowOffset: { width: 0, height: 4 }, shadowOpacity: 0.3, shadowRadius: 8, elevation: 5 },
  submitBtn: { backgroundColor: 'rgba(255,255,255,0.2)', paddingHorizontal: 20, paddingVertical: 10, borderRadius: 8 }
});

class ErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { hasError: false, error: null };
  }
  static getDerivedStateFromError(error) {
    return { hasError: true, error };
  }
  componentDidCatch(error, errorInfo) {
    console.error("WaiterScreen Crash:", error, errorInfo);
  }
  render() {
    if (this.state.hasError) {
      return (
        <SafeAreaView style={{flex: 1, backgroundColor: '#0d0d0f', padding: 20, justifyContent: 'center'}}>
          <Text style={{color: '#ef4444', fontSize: 20, fontWeight: 'bold'}}>Kutilmagan xatolik yuz berdi!</Text>
          <Text style={{color: '#fff', marginTop: 10}}>{this.state.error?.toString()}</Text>
          <TouchableOpacity onPress={() => this.setState({hasError: false, error: null})} style={{marginTop: 20, padding: 10, backgroundColor: '#333', borderRadius: 8}}>
            <Text style={{color: '#fff', textAlign: 'center'}}>Qayta urinish</Text>
          </TouchableOpacity>
        </SafeAreaView>
      );
    }
    return this.props.children;
  }
}

const WaiterScreenWrapper = (props) => (
  <ErrorBoundary>
    <WaiterScreen {...props} />
  </ErrorBoundary>
);

export default WaiterScreenWrapper;

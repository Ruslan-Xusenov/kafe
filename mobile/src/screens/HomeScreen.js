import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, FlatList, TouchableOpacity, Image, ActivityIndicator, SafeAreaView } from 'react-native';
import api from '../api';
import { useCartStore } from '../store/cartStore';
import { ShoppingCart } from 'lucide-react-native';

const HomeScreen = ({ navigation }) => {
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [activeCategory, setActiveCategory] = useState(null);
  const [loading, setLoading] = useState(true);

  const addItem = useCartStore(s => s.addItem);
  const cartItems = useCartStore(s => s.items);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [catRes, prodRes] = await Promise.all([
        api.get('/catalog/categories'),
        api.get('/catalog/products')
      ]);
      const cats = catRes.data || [];
      setCategories(cats);
      setProducts(prodRes.data || []);
      if (cats.length > 0) setActiveCategory(cats[0].id);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const filteredProducts = products.filter(p => p.category_id === activeCategory && p.is_active);

  const renderCategory = ({ item }) => (
    <TouchableOpacity 
      style={[styles.categoryPill, activeCategory === item.id && styles.categoryPillActive]}
      onPress={() => setActiveCategory(item.id)}
    >
      <Text style={[styles.categoryText, activeCategory === item.id && styles.categoryTextActive]}>
        {item.name}
      </Text>
    </TouchableOpacity>
  );

  const getImageUrl = (path) => {
    if (!path) return null;
    return path.startsWith('/') ? api.defaults.baseURL.replace('/api', '') + path : path;
  };

  const renderProduct = ({ item }) => {
    const inCart = cartItems.find(c => c.product_id === item.id);
    return (
      <View style={styles.productCard}>
        <Image source={{ uri: getImageUrl(item.image_url) }} style={styles.productImage} />
        <View style={styles.productInfo}>
          <Text style={styles.productName}>{item.name}</Text>
          <Text style={styles.productDesc} numberOfLines={2}>{item.description}</Text>
          <View style={styles.productBottom}>
            <Text style={styles.productPrice}>{item.price.toLocaleString()} сум</Text>
            <TouchableOpacity 
              style={[styles.addButton, inCart && styles.addButtonActive]} 
              onPress={() => addItem(item)}
            >
              <ShoppingCart size={16} color="#fff" />
              <Text style={styles.addButtonText}>{inCart ? `+${item.quantity_step}` : "Добавить"}</Text>
            </TouchableOpacity>
          </View>
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
        <Text style={styles.headerTitle}>Kafe<Text style={{color: '#f97316'}}>Plat</Text></Text>
        <Text style={styles.headerSub}>Список самых вкусных блюд</Text>
      </View>

      <View style={styles.categoriesContainer}>
        <FlatList
          horizontal
          data={categories}
          renderItem={renderCategory}
          keyExtractor={item => item.id.toString()}
          showsHorizontalScrollIndicator={false}
          contentContainerStyle={{ paddingHorizontal: 20 }}
        />
      </View>

      <FlatList
        data={filteredProducts}
        renderItem={renderProduct}
        keyExtractor={item => item.id.toString()}
        contentContainerStyle={{ padding: 20, paddingBottom: 100 }}
        showsVerticalScrollIndicator={false}
      />
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#0d0d0f' },
  center: { flex: 1, backgroundColor: '#0d0d0f', justifyContent: 'center', alignItems: 'center' },
  header: { padding: 20, paddingTop: 40 },
  headerTitle: { color: '#fff', fontSize: 28, fontWeight: 'bold' },
  headerSub: { color: '#888', fontSize: 14, marginTop: 4 },
  categoriesContainer: { paddingBottom: 15 },
  categoryPill: { 
    paddingHorizontal: 20, paddingVertical: 10, 
    borderRadius: 20, backgroundColor: '#1a1a1f', 
    marginRight: 10, borderWidth: 1, borderColor: '#222' 
  },
  categoryPillActive: { backgroundColor: '#f97316', borderColor: '#f97316' },
  categoryText: { color: '#888', fontWeight: '600' },
  categoryTextActive: { color: '#fff' },
  
  productCard: {
    backgroundColor: '#1a1a1f', borderRadius: 16, 
    marginBottom: 20, overflow: 'hidden',
    borderWidth: 1, borderColor: '#222'
  },
  productImage: { width: '100%', height: 180, resizeMode: 'cover' },
  productInfo: { padding: 15 },
  productName: { color: '#fff', fontSize: 18, fontWeight: 'bold', marginBottom: 5 },
  productDesc: { color: '#888', fontSize: 13, marginBottom: 15 },
  productBottom: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  productPrice: { color: '#f97316', fontSize: 18, fontWeight: '900' },
  addButton: { 
    flexDirection: 'row', alignItems: 'center', gap: 6,
    backgroundColor: '#333', paddingHorizontal: 15, paddingVertical: 8, 
    borderRadius: 8 
  },
  addButtonActive: { backgroundColor: '#f97316' },
  addButtonText: { color: '#fff', fontWeight: 'bold', fontSize: 13 }
});

export default HomeScreen;

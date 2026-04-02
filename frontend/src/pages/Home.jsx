import React, { useState, useEffect } from 'react';
import api, { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import { Plus, ShoppingCart, Loader2 } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

const Home = () => {
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [selectedCat, setSelectedCat] = useState(null);
  const [loading, setLoading] = useState(true);
  const addItem = useCartStore(state => state.addItem);

  useEffect(() => {
    fetchData();
  }, []);

  const getImageUrl = (url) => {
    if (!url) return 'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=500&q=80';
    if (url.startsWith('/')) return url; // Use relative path for production
    return url;
  };

  const fetchData = async () => {
    try {
      const [catRes, prodRes] = await Promise.all([
        api.get('/catalog/categories'),
        api.get('/catalog/products')
      ]);
      setCategories(catRes.data || []);
      setProducts(prodRes.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const filteredProducts = selectedCat 
    ? products.filter(p => p.category_id === selectedCat)
    : products;

  if (loading) return (
    <div className="flex-center h-full">
      <Loader2 className="animate-spin text-primary" size={40} />
    </div>
  );

  return (
    <div className="home-page">
      <section className="hero animate-fade">
        <h1>Mazali taomlar yetkazib beramiz</h1>
        <p>Sevimli kafengizdan eng sara taomlar to'plami</p>
      </section>

      <section className="categories-scroll animate-fade">
        <button 
          className={`cat-tab ${!selectedCat ? 'active' : ''}`}
          onClick={() => setSelectedCat(null)}
        >
          Barchasi
        </button>
        {categories.map(cat => (
          <button 
            key={cat.id}
            className={`cat-tab ${selectedCat === cat.id ? 'active' : ''}`}
            onClick={() => setSelectedCat(cat.id)}
          >
            {cat.name}
          </button>
        ))}
      </section>

      <section className="product-grid">
        <AnimatePresence mode='popLayout'>
          {filteredProducts.length > 0 ? (
            filteredProducts.map(prod => (
              <motion.div 
                key={prod.id} 
                layout
                initial={{ opacity: 0, scale: 0.9 }}
                animate={{ opacity: 1, scale: 1 }}
                exit={{ opacity: 0, scale: 0.9 }}
                className="premium-card product-card"
              >
                <div className="prod-img-wrapper">
                  <img 
                    src={getImageUrl(prod.image_url)} 
                    alt={prod.name} 
                    onError={(e) => {
                      e.target.onerror = null;
                      e.target.src = 'https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=500&q=80';
                    }}
                  />
                </div>
                <div className="prod-info">
                  <h3>{prod.name}</h3>
                  <p>{prod.description}</p>
                  <div className="prod-footer">
                    <span className="price">{prod.price.toLocaleString()} so'm</span>
                    <button className="add-btn" onClick={() => addItem(prod)}>
                      <Plus size={20} />
                    </button>
                  </div>
                </div>
              </motion.div>
            ))
          ) : (
            <div className="empty-state">
              <p>Hozircha mahsulotlar mavjud emas.</p>
              <p className="hint">Admin panel orqali yangi mahsulotlar qo'shishingiz mumkin.</p>
              {useAuthStore.getState().user?.role === 'admin' && (
                <button 
                  className="btn-primary mt-2" 
                  style={{margin: '1.5rem auto 0'}} 
                  onClick={() => window.location.href='/admin'}
                >
                  Dashboardga o'tish
                </button>
              )}
            </div>
          )}
        </AnimatePresence>
      </section>

      <style>{`
        .empty-state {
          grid-column: 1 / -1;
          text-align: center;
          padding: 4rem;
          background: rgba(255,255,255,0.05);
          border-radius: 20px;
          border: 1px dashed var(--border);
        }

        .empty-state p {
          font-size: 1.2rem;
          color: var(--text-dim);
        }

        .empty-state .hint {
          font-size: 0.9rem;
          margin-top: 0.5rem;
          opacity: 0.7;
        }
        .home-page {
          padding-bottom: 4rem;
        }

        .hero {
          text-align: center;
          margin-bottom: 3rem;
          padding: 2rem 0;
        }

        .hero h1 {
          font-size: 3rem;
          margin-bottom: 1rem;
          background: linear-gradient(to right, #fff, var(--primary));
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }

        .hero p {
          color: var(--text-dim);
          font-size: 1.2rem;
        }

        .categories-scroll {
          display: flex;
          gap: 1rem;
          overflow-x: auto;
          padding-bottom: 1rem;
          margin-bottom: 2rem;
          scrollbar-width: none;
        }

        .categories-scroll::-webkit-scrollbar { display: none; }

        .cat-tab {
          padding: 0.75rem 1.5rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 25px;
          color: var(--text-dim);
          white-space: nowrap;
          transition: var(--transition);
        }

        .cat-tab.active {
          background: var(--primary);
          color: white;
          border-color: var(--primary);
        }

        .product-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 2rem;
        }

        .product-card {
          padding: 0;
          overflow: hidden;
          display: flex;
          flex-direction: column;
        }

        .prod-img-wrapper {
          width: 100%;
          height: 200px;
          overflow: hidden;
        }

        .prod-img-wrapper img {
          width: 100%;
          height: 100%;
          object-fit: cover;
          transition: var(--transition);
        }

        .product-card:hover .prod-img-wrapper img {
          transform: scale(1.1);
        }

        .prod-info {
          padding: 1.5rem;
          flex: 1;
          display: flex;
          flex-direction: column;
        }

        .prod-info h3 {
          margin-bottom: 0.5rem;
          font-size: 1.25rem;
        }

        .prod-info p {
          color: var(--text-dim);
          font-size: 0.9rem;
          margin-bottom: 1.5rem;
          flex: 1;
        }

        .prod-footer {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .price {
          font-weight: 700;
          font-size: 1.1rem;
          color: var(--primary);
        }

        .add-btn {
          background: var(--primary);
          color: white;
          width: 40px;
          height: 40px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .flex-center {
          display: flex;
          align-items: center;
          justify-content: center;
        }
      `}</style>
    </div>
  );
};

export default Home;

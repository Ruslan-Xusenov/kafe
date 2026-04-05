import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api, { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import { Plus, ShoppingCart, Loader2, Search, SlidersHorizontal, Eye } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

const Home = () => {
  const { user } = useAuthStore();
  const { addItem } = useCartStore();
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedCat, setSelectedCat] = useState(null);
  const [search, setSearch] = useState('');
  const [selectedUnits, setSelectedUnits] = useState({}); // { productId: 'dona' | 'pors' }

  const getImageUrl = (url) => {
    if (!url) return null;
    if (url.startsWith('/')) return url;
    return url;
  };

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [catRes, prodRes] = await Promise.all([
        api.get('/catalog/categories'),
        api.get('/catalog/products')
      ]);
      setCategories(Array.isArray(catRes.data) ? catRes.data : []);
      setProducts(Array.isArray(prodRes.data) ? prodRes.data : []);
      
      // Initialize default units
      const initialUnits = {};
      (prodRes.data || []).forEach(p => {
        initialUnits[p.id] = p.unit || 'dona';
      });
      setSelectedUnits(initialUnits);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const toggleUnit = (productId) => {
    setSelectedUnits(prev => ({
      ...prev,
      [productId]: prev[productId] === 'pors' ? 'dona' : 'pors'
    }));
  };

  const filteredProducts = products.filter(p => {
    const matchCat = !selectedCat || p.category_id === selectedCat;
    const matchSearch = !search || p.name.toLowerCase().includes(search.toLowerCase());
    return matchCat && matchSearch;
  });

  if (loading) return (
    <div className="loading-screen">
      <div className="spinner" />
      <p style={{ color: 'var(--text-secondary)', fontWeight: 600 }}>Yuklanmoqda...</p>
    </div>
  );

  const getProductPrice = (prod) => {
    const unit = selectedUnits[prod.id] || prod.unit;
    if (unit === 'dona' && prod.unit === 'pors') {
      return prod.price / 4;
    }
    return prod.price;
  };

  return (
    <div className="home-page">
      {/* ── HERO ── */}
      <section className="hero-section animate-fade">
        <div className="hero-badge">
          <span>🔥</span>
          <span>Eng mashhur taomlar</span>
        </div>
        <h1 className="hero-title">
          Mazali taomlar<br />
          <span className="text-gradient">yetkazib beramiz</span>
        </h1>
        <p className="hero-sub">
          Sevimli kafengizdan eng sara taomlar — tez, issiq va ajoyib!
        </p>

        {/* Search */}
        <div className="search-bar-wrap">
          <div className="search-bar">
            <Search size={18} className="search-icon" />
            <input
              type="text"
              placeholder="Taom qidiring..."
              value={search}
              onChange={e => setSearch(e.target.value)}
              className="search-input"
            />
          </div>
        </div>
      </section>

      {/* ── CATEGORY TABS ── */}
      <div className="cat-section">
        <div className="scroll-x">
          <button
            className={`cat-chip ${!selectedCat ? 'active' : ''}`}
            onClick={() => setSelectedCat(null)}
          >
            🍽 Barchasi
          </button>
          {categories.map(cat => (
            <button
              key={cat.id}
              className={`cat-chip ${selectedCat === cat.id ? 'active' : ''}`}
              onClick={() => setSelectedCat(selectedCat === cat.id ? null : cat.id)}
            >
              {cat.name}
            </button>
          ))}
        </div>
      </div>

      {/* ── PRODUCTS GRID ── */}
      <section className="product-grid">
        <AnimatePresence mode="popLayout">
          {filteredProducts.length > 0 ? (
            filteredProducts.map((prod, i) => {
              const cat = categories.find(c => c.id === prod.category_id);
              const isUserControlled = cat?.is_user_controlled;
              const currentUnit = selectedUnits[prod.id] || prod.unit;
              const currentPrice = getProductPrice(prod);

              return (
                <motion.div
                  key={prod.id}
                  layout
                  initial={{ opacity: 0, y: 24 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.9 }}
                  transition={{ duration: 0.35, delay: i * 0.04 }}
                  className="prod-card"
                >
                  <Link to={`/product/${prod.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
                    {/* Image */}
                    <div className="prod-img-wrap">
                      {getImageUrl(prod.image_url) ? (
                        <img
                          src={getImageUrl(prod.image_url)}
                          alt={prod.name}
                          className="prod-img"
                          onError={e => {
                            e.target.onerror = null;
                            e.target.parentElement.innerHTML = '<div class="prod-img-placeholder">🍽</div>';
                          }}
                        />
                      ) : (
                        <div className="prod-img-placeholder">🍽</div>
                      )}
                      <div className="prod-img-overlay">
                        <div className="view-details">
                          <Eye size={18} />
                          <span>Batafsil</span>
                        </div>
                      </div>
                    </div>

                    {/* Info */}
                    <div className="prod-body">
                      <p className="prod-cat-tag">
                        {cat?.name || ''}
                      </p>
                      <h3 className="prod-name">{prod.name}</h3>
                      {prod.description && (
                        <p className="prod-desc">{prod.description}</p>
                      )}

                      {/* Unit Selector if User Controlled */}
                      {isUserControlled && prod.unit === 'pors' && (
                        <div className="unit-selector-mini" onClick={(e) => { e.preventDefault(); e.stopPropagation(); }}>
                          <button 
                            className={currentUnit === 'pors' ? 'active' : ''} 
                            onClick={() => setSelectedUnits(prev => ({...prev, [prod.id]: 'pors'}))}
                          >
                            Pors
                          </button>
                          <button 
                            className={currentUnit === 'dona' ? 'active' : ''} 
                            onClick={() => setSelectedUnits(prev => ({...prev, [prod.id]: 'dona'}))}
                          >
                            Dona
                          </button>
                        </div>
                      )}

                      <div className="prod-footer">
                        <span className="prod-price">
                          {currentPrice.toLocaleString()} so'm 
                          <span className="prod-unit"> / {currentUnit}</span>
                        </span>
                        
                        {currentUnit === 'pors' && (
                          <div className="portion-info">
                            (1 pors = 4 dona)
                          </div>
                        )}
                        
                        <button
                          className="add-btn"
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            addItem({
                              ...prod,
                              price: currentPrice,
                              unit: currentUnit
                            });
                          }}
                          title="Savatga qo'shish"
                        >
                          <Plus size={20} />
                        </button>
                      </div>
                    </div>
                  </Link>
                </motion.div>
              );
            })
          ) : (
            <div className="empty-products">
              <div className="empty-emoji">🍳</div>
              <h3>Mahsulot topilmadi</h3>
              <p>Boshqa kategoriya yoki qidiruvni sinab ko'ring.</p>
              {user?.role === 'admin' && (
                <button
                  className="btn-primary mt-2"
                  onClick={() => window.location.href='/admin'}
                >
                  Mahsulot qo'shish
                </button>
              )}
            </div>
          )}
        </AnimatePresence>
      </section>

      <style>{`
        /* Existing Home Styles ... */
        .prod-img-overlay {
          position: absolute;
          inset: 0;
          background: linear-gradient(to top, rgba(13,13,15,0.7) 0%, transparent 60%);
          display: flex; align-items: flex-end; justify-content: center;
          padding-bottom: 1.5rem; opacity: 0; transition: 0.3s;
        }

        .prod-card:hover .prod-img-overlay { opacity: 1; }

        .view-details {
          display: flex; align-items: center; gap: 0.5rem;
          background: rgba(255,255,255,0.15); backdrop-filter: blur(8px);
          padding: 0.4rem 1rem; border-radius: 99px; font-size: 0.75rem;
          font-weight: 700; color: white; border: 1px solid rgba(255,255,255,0.25);
        }
        .home-page {
          padding-bottom: 4rem;
        }

        /* HERO */
        .hero-section {
          text-align: center;
          padding: 3.5rem 1rem 2.5rem;
          position: relative;
        }

        .hero-badge {
          display: inline-flex;
          align-items: center;
          gap: 0.4rem;
          background: rgba(249,115,22,0.12);
          border: 1px solid rgba(249,115,22,0.30);
          border-radius: 99px;
          padding: 0.4rem 1rem;
          font-size: 0.82rem;
          font-weight: 700;
          color: var(--primary-light);
          letter-spacing: 0.04em;
          margin-bottom: 1.5rem;
          text-transform: uppercase;
        }

        .hero-title {
          font-size: clamp(2.4rem, 6vw, 4rem);
          margin-bottom: 1rem;
          letter-spacing: -0.03em;
          line-height: 1.1;
        }

        .hero-sub {
          color: var(--text-secondary);
          font-size: 1.05rem;
          max-width: 480px;
          margin: 0 auto 2.5rem;
        }

        /* SEARCH */
        .search-bar-wrap {
          max-width: 480px;
          margin: 0 auto;
        }

        .search-bar {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: var(--radius-xl);
          padding: 0 1.25rem;
          transition: var(--transition);
          box-shadow: var(--shadow-sm);
        }

        .search-bar:focus-within {
          border-color: var(--border-focus);
          box-shadow: 0 0 0 3px rgba(249,115,22,0.12), var(--shadow-sm);
        }

        .search-icon { color: var(--text-muted); flex-shrink: 0; }

        .search-input {
          background: none;
          border: none;
          padding: 0.85rem 0;
          color: var(--text-primary);
          font-size: 0.95rem;
          width: 100%;
        }

        .search-input::placeholder { color: var(--text-muted); }
        .search-input:focus { box-shadow: none; }

        /* CATEGORIES */
        .cat-section {
          margin-bottom: 2rem;
        }

        .cat-chip {
          display: inline-flex;
          align-items: center;
          gap: 0.35rem;
          padding: 0.55rem 1.2rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 99px;
          font-size: 0.875rem;
          font-weight: 600;
          color: var(--text-secondary);
          white-space: nowrap;
          transition: var(--transition);
          flex-shrink: 0;
        }

        .cat-chip:hover {
          border-color: rgba(249,115,22,0.40);
          color: var(--primary);
          background: rgba(249,115,22,0.08);
        }

        .cat-chip.active {
          background: var(--grad-brand);
          border-color: transparent;
          color: white;
          box-shadow: 0 4px 12px rgba(249,115,22,0.35);
        }

        /* PRODUCT GRID */
        .product-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
          gap: 1.5rem;
        }

        /* PRODUCT CARD */
        .prod-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius-lg);
          overflow: hidden;
          transition: var(--transition);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
          display: flex;
          flex-direction: column;
        }

        .prod-card:hover {
          border-color: rgba(249,115,22,0.30);
          transform: translateY(-4px);
          box-shadow: 0 12px 32px rgba(0,0,0,0.4), 0 0 0 1px rgba(249,115,22,0.15);
        }

        .prod-img-wrap {
          width: 100%;
          height: 190px;
          overflow: hidden;
          position: relative;
          background: rgba(255,255,255,0.03);
        }

        .prod-img {
          width: 100%;
          height: 100%;
          object-fit: cover;
          transition: var(--transition-slow);
        }

        .prod-card:hover .prod-img { transform: scale(1.07); }

        .prod-img-overlay {
          position: absolute;
          inset: 0;
          background: linear-gradient(to top, rgba(13,13,15,0.6) 0%, transparent 50%);
          pointer-events: none;
        }

        .prod-img-placeholder {
          width: 100%; height: 100%;
          display: flex; align-items: center; justify-content: center;
          font-size: 4rem;
          opacity: 0.3;
        }

        .prod-body {
          padding: 1.25rem 1.25rem 1rem;
          flex: 1;
          display: flex;
          flex-direction: column;
          gap: 0.4rem;
        }

        .prod-cat-tag {
          font-size: 0.72rem;
          font-weight: 700;
          color: var(--primary);
          text-transform: uppercase;
          letter-spacing: 0.06em;
        }

        .prod-name {
          font-family: var(--font);
          font-size: 1rem;
          font-weight: 700;
          color: var(--text-primary);
          line-height: 1.3;
        }

        .prod-desc {
          font-size: 0.82rem;
          color: var(--text-secondary);
          line-height: 1.5;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
          flex: 1;
        }

        .prod-footer {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-top: 0.75rem;
        }

        .prod-price {
          font-size: 1.1rem;
          font-weight: 800;
          color: var(--primary);
        }

        .prod-unit {
          font-size: 0.75rem;
          color: var(--text-secondary);
          font-weight: 600;
          margin-left: 2px;
          -webkit-text-fill-color: var(--text-secondary);
        }

        .add-btn {
          width: 38px; height: 38px;
          background: var(--grad-brand);
          color: white;
          border-radius: 10px;
          display: flex;
          align-items: center;
          justify-content: center;
          box-shadow: 0 4px 12px rgba(249,115,22,0.35);
          transition: var(--transition);
          flex-shrink: 0;
        }

        .add-btn:hover {
          transform: scale(1.12) rotate(8deg);
          box-shadow: 0 6px 20px rgba(249,115,22,0.50);
        }

        .add-btn:active { transform: scale(0.94); }

        /* EMPTY */
        .empty-products {
          grid-column: 1 / -1;
          text-align: center;
          padding: 5rem 2rem;
        }

        .empty-emoji {
          font-size: 4rem;
          margin-bottom: 1.25rem;
          animation: fadeUp 0.5s ease;
        }

        .empty-products h3 {
          font-size: 1.5rem;
          margin-bottom: 0.5rem;
          color: var(--text-primary);
        }

        .empty-products p {
          color: var(--text-secondary);
          margin-bottom: 1.5rem;
        }

        @media (max-width: 900px) {
          .hero-section { padding: 2.5rem 1rem 2rem; }
          .hero-title { font-size: 2.4rem; }
          .product-grid { grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; }
          .prod-img-wrap { height: 160px; }
        }

        @media (max-width: 640px) {
          .hero-section { padding: 2rem 0.5rem 1.75rem; }
          .hero-title { font-size: 2rem; }
          .hero-sub { font-size: 0.9rem; margin-bottom: 1.75rem; }
          .hero-badge { font-size: 0.75rem; padding: 0.3rem 0.75rem; }
          .product-grid { grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
          .prod-img-wrap { height: 130px; }
          .prod-body { padding: 0.9rem 0.9rem 0.75rem; }
          .prod-name { font-size: 0.88rem; }
          .prod-price { font-size: 0.95rem; }
          .add-btn { width: 32px; height: 32px; border-radius: 8px; }
          .search-bar { padding: 0 1rem; }
        }

        @media (max-width: 380px) {
          .product-grid { grid-template-columns: 1fr; }
          .prod-img-wrap { height: 180px; }
        }
      `}</style>
    </div>
  );
};

export default Home;

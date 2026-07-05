import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { motion, AnimatePresence } from 'framer-motion';
import { LayoutDashboard, ShoppingCart, Plus, Minus, ArrowLeft, Send, CheckCircle2, Coffee, UtensilsCrossed, Check, Clock, X } from 'lucide-react';

const Waiter = () => {
  const [tables, setTables] = useState([]);
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  
  const [selectedTable, setSelectedTable] = useState(null);
  const [activeCategory, setActiveCategory] = useState(null);
  const [cart, setCart] = useState([]);
  
  const [history, setHistory] = useState([]);
  const [showHistory, setShowHistory] = useState(false);
  
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchInitialData();
  }, []);

  const STATUS_MAP = {
    new: 'Yangi',
    preparing: 'Tayyorlanmoqda',
    ready: 'Tayyor',
    on_way: 'Yo\'lda',
    delivered: 'Yopilgan',
    cancelled: 'Bekor qilingan'
  };

  const fetchInitialData = async () => {
    setLoading(true);
    try {
      const [tableRes, catRes, prodRes] = await Promise.all([
        api.get(`/tables/?_t=${Date.now()}`),
        api.get('/catalog/categories'),
        api.get('/catalog/products')
      ]);
      setTables(tableRes.data || []);
      setCategories(catRes.data || []);
      setProducts(prodRes.data || []);
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      console.error("fetchInitialData ERROR:", err.message, err.response?.data);
      alert("Ma'lumotlarni yuklashda xatolik: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const fetchHistory = async () => {
    try {
      setLoading(true);
      const res = await api.get('/orders/waiter-history');
      setHistory(res.data || []);
      setShowHistory(true);
    } catch (err) {
      console.error(err);
      alert("Tarixni yuklashda xatolik: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const handleTableSelect = (table) => {
    setSelectedTable(table);
    setCart([]);
    setActiveCategory(categories.length > 0 ? categories[0].id : null);
  };

  const addToCart = (product) => {
    setCart(prev => {
      const existing = prev.find(item => item.product_id === product.id);
      if (existing) {
        return prev.map(item => 
          item.product_id === product.id 
            ? { ...item, quantity: item.quantity + product.quantity_step }
            : item
        );
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
      return prev.map(item => {
        if (item.product_id === productId) {
          const newQ = item.quantity + delta;
          return newQ > 0 ? { ...item, quantity: newQ } : item;
        }
        return item;
      }).filter(item => item.quantity > 0);
    });
  };

  const submitOrder = async () => {
    if (cart.length === 0) return alert("Savatcha bo'sh!");
    if (!selectedTable) return alert("Stol tanlanmagan!");

    const totalPrice = cart.reduce((acc, item) => acc + (item.price * item.quantity), 0);
    
    try {
      const payload = {
        table_id: selectedTable.id,
        items: cart.map(item => ({
          product_id: item.product_id,
          quantity: item.quantity,
          price: item.price,
          unit: item.unit
        })),
        total_price: totalPrice,
        address: `Stol: ${selectedTable.number}`,
        phone: 'Ichki buyurtma'
      };

      await api.post('/orders', payload);
      
      if (selectedTable.status === 'free') {
         await api.put(`/tables/${selectedTable.id}`, { ...selectedTable, status: 'occupied' });
      }

      alert('Buyurtma oshxonaga yuborildi!');
      setSelectedTable(null);
      setCart([]);
      fetchInitialData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      console.error(err);
      alert('Buyurtma yuborishda xatolik');
    }
  };

  const freeTable = async () => {
    if (!window.confirm("Stolni bo'shatishni xohlaysizmi?")) return;
    
    const paymentMethodMap = {
      '1': 'cash',
      '2': 'card',
      '3': 'click',
      '4': 'nasiya'
    };
    
    let method = null;
    while (!method) {
      const choice = window.prompt("To'lov turini tanlang (Raqamni kiriting):\n1 - 💵 Naqd\n2 - 💳 Terminal (Karta)\n3 - 📱 Click/Payme\n4 - 📒 Qarzga (Nasiya)");
      
      if (choice === null) return; // User cancelled
      
      method = paymentMethodMap[choice];
      if (!method) {
        alert("Iltimos, 1 dan 4 gacha bo'lgan raqamni kiriting.");
      }
    }

    try {
      await api.put(`/tables/${selectedTable.id}`, { ...selectedTable, status: 'free', payment_method: method });
      alert("Stol bo'shatildi!");
      setSelectedTable(null);
      fetchInitialData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      console.error(err);
      const msg = err.response?.data?.error || "Xatolik yuz berdi";
      alert(msg);
    }
  };

  if (loading) return (
    <div className="flex-center h-screen w-full">
      <div className="spinner-glow"></div>
    </div>
  );

  const cartTotal = cart.reduce((acc, item) => acc + (item.price * item.quantity), 0);
  const freeCount = tables.filter(t => t.status === 'free').length;
  const occCount = tables.filter(t => t.status !== 'free').length;

  return (
    <div className="waiter-wrapper">
      <AnimatePresence mode="wait">
        {!selectedTable ? (
          <motion.div 
            key="tables-view"
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -20 }}
            className="tables-container"
          >
            <header className="waiter-header glass">
              <div className="header-info">
                <UtensilsCrossed size={32} className="text-primary" />
                <div>
                  <h1 className="waiter-title">Zal Boshqaruvi</h1>
                  <p className="waiter-subtitle">Xizmat ko'rsatish uchun stolni tanlang</p>
                </div>
              </div>
              <div className="table-stats">
                <button className="btn-secondary" style={{ padding: '0.5rem 1rem', borderRadius: '99px', fontSize: '0.85rem' }} onClick={fetchHistory}>
                  <Clock size={16} style={{ display: 'inline', marginRight: '4px', verticalAlign: 'text-bottom' }}/> Tarix
                </button>
                <div className="stat-pill free-pill">
                  <span className="dot"></span>
                  {freeCount} Bo'sh
                </div>
                <div className="stat-pill occ-pill">
                  <span className="dot"></span>
                  {occCount} Band
                </div>
              </div>
            </header>

            <div className="tables-grid">
              {tables.map((table, i) => (
                <motion.div 
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: i * 0.05 }}
                  key={table.id} 
                  className={`premium-table-card ${table.status}`}
                  onClick={() => handleTableSelect(table)}
                >
                  <div className="table-card-inner">
                    <div className="table-icon-wrap">
                      <Coffee size={24} />
                    </div>
                    <div className="table-details">
                      <h3>№ {table.number}</h3>
                      <span>{table.capacity} kishi</span>
                    </div>
                  </div>
                  <div className="table-status-bar">
                    {table.status === 'free' ? (
                      <><CheckCircle2 size={14} /> Bo'sh</>
                    ) : (
                      <><LayoutDashboard size={14} /> Band</>
                    )}
                  </div>
                </motion.div>
              ))}
              {tables.length === 0 && (
                <div className="no-tables">
                  <p>Hozircha stollar yo'q</p>
                </div>
              )}
            </div>
          </motion.div>
        ) : (
          <motion.div 
            key="order-view"
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: 20 }}
            className="order-container"
          >
            <header className="order-header glass sticky top-0 z-20">
              <button className="back-button" onClick={() => setSelectedTable(null)}>
                <ArrowLeft size={20} />
                <span>Ortga</span>
              </button>
              <div className="header-center">
                <h2>Stol № {selectedTable.number}</h2>
                <span className={`status-badge-small ${selectedTable.status}`}>
                  {selectedTable.status === 'free' ? 'Bo\'sh' : 'Band'}
                </span>
              </div>
              <div className="header-right">
                {selectedTable.status !== 'free' && (
                  <button onClick={freeTable} className="btn-free-table">
                    Stolni bo'shatish
                  </button>
                )}
                <div className="order-total-pill">
                  {cartTotal.toLocaleString()} <small>so'm</small>
                </div>
              </div>
            </header>

            <div className="menu-area">
              <div className="categories-slider no-scrollbar">
                {categories.map(c => (
                  <button 
                    key={c.id} 
                    className={`modern-cat-btn ${activeCategory === c.id ? 'active' : ''}`}
                    onClick={() => setActiveCategory(c.id)}
                  >
                    {c.name}
                  </button>
                ))}
              </div>

              <div className="menu-grid pb-[300px]">
                {products.filter(p => p.category_id === activeCategory && p.is_active).map((p, i) => {
                  const cartItem = cart.find(c => c.product_id === p.id);
                  return (
                    <motion.div 
                      initial={{ opacity: 0, y: 10 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: i * 0.03 }}
                      key={p.id} 
                      className={`menu-card ${cartItem ? 'selected' : ''}`}
                      onClick={() => addToCart(p)}
                    >
                      <div className="menu-card-img" style={{backgroundImage: `url(${p.image_url.startsWith('/') ? api.defaults.baseURL.replace('/api', '') + p.image_url : p.image_url})`}}>
                        {cartItem && (
                          <div className="selected-overlay">
                            <Check size={24} />
                          </div>
                        )}
                      </div>
                      <div className="menu-card-info">
                        <h4>{p.name}</h4>
                        <div className="price-row">
                          <span className="price">{p.price.toLocaleString()} so'm</span>
                        </div>
                      </div>
                    </motion.div>
                  );
                })}
              </div>
            </div>

            <AnimatePresence>
              {cart.length > 0 && (
                <motion.div 
                  initial={{ y: 100, opacity: 0 }}
                  animate={{ y: 0, opacity: 1 }}
                  exit={{ y: 100, opacity: 0 }}
                  className="smart-cart-panel glass"
                >
                  <div className="cart-panel-header">
                    <h3>Savatcha ({cart.length})</h3>
                    <span>Jami: {cartTotal.toLocaleString()} so'm</span>
                  </div>
                  <div className="smart-cart-items no-scrollbar">
                    {cart.map(item => (
                      <div key={item.product_id} className="smart-cart-item">
                        <div className="item-details">
                          <span className="name">{item.name}</span>
                          <span className="price">{(item.price * item.quantity).toLocaleString()} so'm</span>
                        </div>
                        <div className="item-actions">
                          <button onClick={() => updateQuantity(item.product_id, -item.step)}>
                            <Minus size={14} />
                          </button>
                          <span className="qty">{item.quantity} <small>{item.unit}</small></span>
                          <button onClick={() => updateQuantity(item.product_id, item.step)}>
                            <Plus size={14} />
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                  <button className="submit-order-btn" onClick={submitOrder}>
                    <Send size={18} />
                    Oshxonaga Yuborish
                  </button>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        )}
      </AnimatePresence>

      {/* History Modal */}
      {showHistory && (
        <div className="modal-overlay" style={{ zIndex: 100 }}>
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content" style={{maxWidth: '800px', background: '#121212', maxHeight: '90vh', display: 'flex', flexDirection: 'column'}}>
            <div className="modal-header" style={{ padding: '1.5rem', borderBottom: '1px solid var(--border)' }}>
              <h3>Mening Buyurtmalarim (Tarix)</h3>
              <button onClick={() => setShowHistory(false)}><X size={20} /></button>
            </div>
            <div className="order-details-body" style={{ padding: '1.5rem', overflowY: 'auto', flex: 1 }}>
              {history.length === 0 ? (
                <div className="text-center text-muted py-8">Hali buyurtmalar yo'q</div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  {history.map(order => (
                    <div key={order.id} style={{ background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border)', borderRadius: '12px', padding: '1rem' }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                        <span className="font-700 text-primary">Chek #{order.id}</span>
                        <span className="text-muted" style={{ fontSize: '0.85rem' }}>{new Date(order.created_at).toLocaleString()}</span>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem', fontSize: '0.9rem' }}>
                        <span>Stol: <b>№{order.table_number || order.table_id || '-'}</b></span>
                        <span className={`status-badge-small ${order.status}`}>
                          {STATUS_MAP[order.status] || order.status}
                        </span>
                      </div>
                      <div style={{ fontSize: '0.9rem', marginBottom: '0.5rem' }}>
                        <span className="text-muted">To'lov: </span>
                        {order.payment_method === 'cash' ? '💵 Naqd' :
                         order.payment_method === 'card' ? '💳 Terminal (Karta)' :
                         order.payment_method === 'click' ? '📱 Click/Payme' :
                         order.payment_method === 'nasiya' ? '📒 Qarzga' : (order.payment_method || '-')}
                      </div>
                      <div style={{ marginTop: '0.5rem', paddingTop: '0.5rem', borderTop: '1px dashed var(--border)', display: 'flex', justifyContent: 'space-between' }}>
                        <span className="text-muted">Jami summasi:</span>
                        <span className="font-700">{(order.total_price || 0).toLocaleString()} so'm</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </motion.div>
        </div>
      )}

      <style>{`
        .waiter-wrapper {
          min-height: 100vh;
          background: var(--bg-main);
          color: var(--text-main);
        }

        /* Loading Spinner */
        .spinner-glow {
          width: 40px; height: 40px;
          border-radius: 50%;
          border: 3px solid rgba(249,115,22,0.2);
          border-top-color: var(--primary);
          animation: spin 1s linear infinite;
        }

        /* ── Tables View ── */
        .tables-container { padding: 1.5rem; max-width: 1200px; margin: 0 auto; padding-bottom: 100px; }
        
        .waiter-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1.5rem;
          border-radius: var(--radius-lg);
          margin-bottom: 2rem;
          flex-wrap: wrap;
          gap: 1rem;
        }

        .header-info { display: flex; align-items: center; gap: 1rem; }
        .waiter-title { font-size: 1.5rem; font-weight: 800; font-family: var(--font-display); background: var(--grad-brand); -webkit-background-clip: text; -webkit-text-fill-color: transparent; margin: 0; }
        .waiter-subtitle { font-size: 0.9rem; color: var(--text-muted); margin: 0; }

        .table-stats { display: flex; gap: 0.75rem; }
        .stat-pill { display: flex; align-items: center; gap: 0.5rem; padding: 0.5rem 1rem; border-radius: 99px; font-size: 0.85rem; font-weight: 600; }
        .stat-pill .dot { width: 8px; height: 8px; border-radius: 50%; }
        .free-pill { background: rgba(16,185,129,0.1); color: #10b981; }
        .free-pill .dot { background: #10b981; box-shadow: 0 0 10px #10b981; }
        .occ-pill { background: rgba(239,68,68,0.1); color: #ef4444; }
        .occ-pill .dot { background: #ef4444; box-shadow: 0 0 10px #ef4444; }

        .tables-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
          gap: 1.25rem;
        }

        .premium-table-card {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: var(--radius-lg);
          overflow: hidden;
          cursor: pointer;
          transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
          position: relative;
        }

        .premium-table-card:hover { transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.2); }
        .premium-table-card:active { transform: scale(0.96); }

        .premium-table-card.free { border-color: rgba(16,185,129,0.2); }
        .premium-table-card.free:hover { border-color: rgba(16,185,129,0.5); box-shadow: 0 8px 24px rgba(16,185,129,0.15); }
        
        .premium-table-card.occupied { border-color: rgba(239,68,68,0.2); }
        .premium-table-card.occupied:hover { border-color: rgba(239,68,68,0.5); box-shadow: 0 8px 24px rgba(239,68,68,0.15); }

        .table-card-inner { padding: 1.5rem; text-align: center; }
        
        .table-icon-wrap {
          width: 50px; height: 50px;
          margin: 0 auto 1rem;
          border-radius: 50%;
          display: flex; align-items: center; justify-content: center;
          background: rgba(255,255,255,0.05);
          color: var(--text-muted);
          transition: var(--transition);
        }

        .free .table-icon-wrap { color: #10b981; background: rgba(16,185,129,0.1); }
        .occupied .table-icon-wrap { color: #ef4444; background: rgba(239,68,68,0.1); }

        .table-details h3 { margin: 0; font-size: 1.5rem; font-weight: 800; color: var(--text-primary); }
        .table-details span { font-size: 0.85rem; color: var(--text-muted); }

        .table-status-bar {
          padding: 0.6rem;
          display: flex; align-items: center; justify-content: center; gap: 0.4rem;
          font-size: 0.8rem; font-weight: 700; text-transform: uppercase; letter-spacing: 0.05em;
        }

        .free .table-status-bar { background: rgba(16,185,129,0.1); color: #10b981; }
        .occupied .table-status-bar { background: rgba(239,68,68,0.1); color: #ef4444; }


        /* ── Order View ── */
        .order-container { position: relative; min-height: 100vh; }
        
        .order-header {
          display: flex; justify-content: space-between; align-items: center;
          padding: 1rem 1.5rem;
          border-bottom: 1px solid var(--border);
        }

        .back-button {
          display: flex; align-items: center; gap: 0.5rem;
          background: rgba(255,255,255,0.05); border: 1px solid var(--border);
          color: var(--text-primary); padding: 0.5rem 1rem; border-radius: 8px;
          font-weight: 600; cursor: pointer; transition: var(--transition);
        }
        .back-button:hover { background: rgba(255,255,255,0.1); }

        .header-center { text-align: center; display: flex; flex-direction: column; align-items: center; gap: 0.2rem; }
        .header-center h2 { margin: 0; font-size: 1.5rem; font-weight: 800; font-family: var(--font-display); }
        .status-badge-small { font-size: 0.75rem; padding: 4px 12px; border-radius: 99px; font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase; }
        .status-badge-small.free { background: rgba(16,185,129,0.15); color: #10b981; border: 1px solid rgba(16,185,129,0.3); }
        .status-badge-small.occupied { background: rgba(239,68,68,0.15); color: #ef4444; border: 1px solid rgba(239,68,68,0.3); }

        .header-right { display: flex; align-items: center; gap: 1rem; }

        .btn-free-table {
          background: rgba(16,185,129,0.1);
          color: #10b981;
          border: 1px solid rgba(16,185,129,0.3);
          padding: 0.5rem 1rem;
          border-radius: 8px;
          font-weight: 700;
          font-size: 0.85rem;
          cursor: pointer;
          transition: all 0.3s ease;
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }
        .btn-free-table:hover {
          background: rgba(16,185,129,0.2);
          border-color: rgba(16,185,129,0.5);
          box-shadow: 0 4px 12px rgba(16,185,129,0.2);
          transform: translateY(-1px);
        }
        .btn-free-table:active { transform: translateY(1px); }

        .order-total-pill {
          background: var(--grad-brand);
          color: white; padding: 0.5rem 1.25rem; border-radius: 8px;
          font-weight: 800; font-size: 1.2rem;
          box-shadow: 0 4px 12px rgba(249,115,22,0.3);
          display: flex; align-items: baseline; gap: 0.3rem;
        }
        .order-total-pill small { font-size: 0.8rem; font-weight: 600; opacity: 0.9; }

        .menu-area { padding: 1.5rem; max-width: 1200px; margin: 0 auto; }

        .categories-slider {
          display: flex; gap: 0.75rem; overflow-x: auto; padding-bottom: 1rem;
          margin-bottom: 1rem; scroll-behavior: smooth;
        }

        .modern-cat-btn {
          padding: 0.7rem 1.5rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 99px;
          color: var(--text-secondary);
          font-weight: 600; font-size: 0.9rem;
          white-space: nowrap; cursor: pointer;
          transition: all 0.2s;
        }
        .modern-cat-btn:hover { border-color: rgba(249,115,22,0.4); color: var(--text-primary); }
        .modern-cat-btn.active { background: var(--primary); color: white; border-color: var(--primary); box-shadow: 0 4px 12px rgba(249,115,22,0.3); }

        .menu-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
          gap: 1.25rem;
        }

        .menu-card {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: var(--radius-lg);
          overflow: hidden;
          cursor: pointer;
          transition: all 0.2s;
          position: relative;
        }
        .menu-card:hover { transform: translateY(-4px); border-color: var(--primary); box-shadow: 0 8px 24px rgba(0,0,0,0.3); }
        .menu-card.selected { border-color: var(--primary); box-shadow: 0 0 0 2px var(--primary); }

        .menu-card-img {
          height: 140px;
          background-size: cover; background-position: center;
          position: relative;
        }

        .selected-overlay {
          position: absolute; inset: 0;
          background: rgba(249,115,22,0.4); backdrop-filter: blur(2px);
          display: flex; align-items: center; justify-content: center;
          color: white;
        }

        .menu-card-info { padding: 1rem; }
        .menu-card-info h4 { margin: 0 0 0.5rem; font-size: 1rem; font-weight: 700; line-height: 1.3; }
        .price-row { display: flex; justify-content: space-between; align-items: center; }
        .price-row .price { color: var(--primary); font-weight: 800; font-size: 1.1rem; }

        /* ── Smart Cart Panel ── */
        .smart-cart-panel {
          position: fixed;
          bottom: 1.5rem; 
          right: 1.5rem;
          width: 450px;
          max-width: calc(100vw - 3rem);
          background: rgba(13,13,15,0.95) !important;
          border: 1px solid rgba(255,255,255,0.1);
          border-radius: 24px;
          padding: 1.5rem;
          box-shadow: 0 10px 50px rgba(0,0,0,0.6);
          z-index: 50;
          max-height: calc(100vh - 3rem);
          display: flex; flex-direction: column;
        }

        .cart-panel-header {
          display: flex; justify-content: space-between; align-items: center;
          margin-bottom: 1rem; padding-bottom: 1rem; border-bottom: 1px solid var(--border);
        }
        .cart-panel-header h3 { margin: 0; font-size: 1.2rem; }
        .cart-panel-header span { color: var(--primary); font-weight: 800; font-size: 1.2rem; }

        .smart-cart-items {
          flex: 1; overflow-y: auto;
          display: flex; flex-direction: column; gap: 0.75rem;
          margin-bottom: 1.5rem;
        }

        .smart-cart-item {
          display: flex; justify-content: space-between; align-items: center;
          background: rgba(255,255,255,0.03);
          border: 1px solid rgba(255,255,255,0.05);
          padding: 0.75rem 1rem;
          border-radius: 12px;
        }

        .item-details { display: flex; flex-direction: column; gap: 0.2rem; }
        .item-details .name { font-weight: 600; font-size: 1rem; }
        .item-details .price { color: var(--text-muted); font-size: 0.85rem; }

        .item-actions {
          display: flex; align-items: center; gap: 0.75rem;
          background: var(--bg-surface); padding: 0.3rem; border-radius: 99px;
          border: 1px solid var(--border);
        }

        .item-actions button {
          width: 32px; height: 32px; border-radius: 50%;
          border: none; background: rgba(255,255,255,0.05); color: var(--text-main);
          display: flex; align-items: center; justify-content: center; cursor: pointer;
          transition: var(--transition);
        }
        .item-actions button:hover { background: rgba(255,255,255,0.1); color: var(--primary); }
        .item-actions .qty { width: 40px; text-align: center; font-weight: 700; font-size: 0.9rem; }
        .item-actions .qty small { font-size: 0.7rem; color: var(--text-muted); font-weight: normal; }

        .submit-order-btn {
          background: var(--grad-brand);
          color: white; border: none; padding: 1.2rem; border-radius: 16px;
          font-size: 1.1rem; font-weight: 800;
          display: flex; align-items: center; justify-content: center; gap: 0.75rem;
          cursor: pointer; box-shadow: 0 8px 24px rgba(249,115,22,0.4);
          transition: all 0.2s; width: 100%;
        }
        .submit-order-btn:hover { transform: translateY(-2px); box-shadow: 0 12px 32px rgba(249,115,22,0.5); }
        .submit-order-btn:active { transform: scale(0.98); }

        /* Utilities */
        .no-scrollbar::-webkit-scrollbar { display: none; }
        .no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }

        @media (max-width: 600px) {
          .waiter-header { flex-direction: column; align-items: flex-start; }
          .table-stats { width: 100%; justify-content: space-between; }
          .stat-pill { flex: 1; justify-content: center; }
          
          .order-header { padding: 1rem; flex-wrap: wrap; gap: 1rem; justify-content: center; }
          .header-right { width: 100%; justify-content: space-between; }
          .back-button { position: absolute; left: 1rem; top: 1rem; }
          .back-button span { display: none; }
          .order-total-pill { font-size: 1.1rem; padding: 0.5rem 1rem; }
          .btn-free-table { padding: 0.5rem 0.8rem; font-size: 0.8rem; }
          
          .menu-grid { grid-template-columns: repeat(2, 1fr); gap: 0.75rem; }
          .menu-card-img { height: 110px; }
          .menu-card-info { padding: 0.75rem; }
          .menu-card-info h4 { font-size: 0.9rem; }
          .price-row .price { font-size: 1rem; }
          
          .smart-cart-panel { 
            position: fixed;
            bottom: 0; left: 0; right: 0;
            width: 100%;
            max-width: none;
            border-radius: 0;
            border-top-left-radius: 20px; 
            border-top-right-radius: 20px; 
            padding: 1.25rem 1rem; 
            max-height: 80vh;
          }
          .item-actions { gap: 0.5rem; }
          .item-actions button { width: 28px; height: 28px; }
          .item-actions .qty { width: 30px; }
        }
      `}</style>
    </div>
  );
};

export default Waiter;

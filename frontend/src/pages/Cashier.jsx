import React, { useState, useEffect, useRef, useCallback } from 'react';
import api, { useAuthStore } from '../store/authStore';
import { motion, AnimatePresence } from 'framer-motion';
import { ShoppingCart, Plus, Minus, Trash2, Printer, Check, Search, Lock, Power, RefreshCw, Banknote, History, ChevronRight } from 'lucide-react';
import PaymentModal from '../components/PaymentModal';
import ShiftModal from '../components/ShiftModal';

const Cashier = () => {
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [activeCategory, setActiveCategory] = useState(null);
  
  const [cart, setCart] = useState([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [loading, setLoading] = useState(true);
  
  const [shift, setShift] = useState(null);
  const [shiftModalState, setShiftModalState] = useState({ isOpen: false, type: 'open' }); // open, close, operation
  
  const [paymentModalOpen, setPaymentModalOpen] = useState(false);
  const [debtors, setDebtors] = useState([]);

  const { user, logout } = useAuthStore();
  const searchInputRef = useRef(null);

  useEffect(() => {
    fetchInitialData();
  }, []);

  const fetchInitialData = async () => {
    setLoading(true);
    try {
      // Parallel requests
      const [catRes, prodRes, shiftRes, debtorsRes] = await Promise.all([
        api.get('/catalog/categories'),
        api.get('/catalog/products'),
        api.get('/cashier/shift/current'),
        api.get('/debts/debtors')
      ]);
      
      setCategories(catRes.data || []);
      setProducts(prodRes.data || []);
      if (catRes.data && catRes.data.length > 0) {
        setActiveCategory(catRes.data[0].id);
      }
      
      if (shiftRes.data && shiftRes.data.id) {
        setShift(shiftRes.data);
      } else {
        setShift(null);
      }
      
      setDebtors(debtorsRes.data || []);
    } catch (err) {
      console.error("fetchInitialData ERROR:", err);
      alert("Ma'lumotlarni yuklashda xatolik: " + (err.response?.data?.error || err.message));
    } finally {
      setLoading(false);
    }
  };

  const refreshShift = async () => {
    try {
      const res = await api.get('/cashier/shift/current');
      if (res.data && res.data.id) setShift(res.data);
      else setShift(null);
    } catch (err) {
      console.error("Error refreshing shift:", err);
    }
  };

  // --- Cart Operations ---
  const addToCart = (product) => {
    if (!shift) return alert("Avval smenani oching!");
    
    setCart(prev => {
      const existing = prev.find(item => item.product_id === product.id);
      if (existing) {
        return prev.map(item => 
          item.product_id === product.id 
            ? { ...item, quantity: item.quantity + (product.quantity_step || 1) }
            : item
        );
      }
      return [{ 
        product_id: product.id, 
        name: product.name, 
        price: product.price, 
        quantity: product.min_quantity || 1,
        unit: product.unit,
        step: product.quantity_step || 1,
        image: product.image_url
      }, ...prev]; // Add to top for visibility
    });
  };

  const updateQuantity = (productId, delta) => {
    setCart(prev => prev.map(item => {
      if (item.product_id === productId) {
        const newQ = item.quantity + delta;
        return newQ > 0 ? { ...item, quantity: newQ } : item;
      }
      return item;
    }).filter(item => item.quantity > 0));
  };

  const removeFromCart = (productId) => {
    setCart(prev => prev.filter(item => item.product_id !== productId));
  };

  const clearCart = () => setCart([]);

  const cartTotal = cart.reduce((acc, item) => acc + (item.price * item.quantity), 0);

  // --- Shift Operations ---
  const handleShiftConfirm = async (data) => {
    try {
      if (shiftModalState.type === 'open') {
        const res = await api.post('/cashier/shift/open', { opening_cash: data.opening_cash });
        setShift(res.data);
        alert("Smena ochildi!");
      } else if (shiftModalState.type === 'close') {
        await api.post('/cashier/shift/close', { shift_id: shift.id, closing_cash: data.closing_cash, notes: data.notes });
        setShift(null); // Clear shift
        alert("Smena yopildi!");
      } else if (shiftModalState.type === 'operation') {
        await api.post('/cashier/shift/cash-operation', { shift_id: shift.id, type: data.type, amount: data.amount, reason: data.reason });
        alert("Kassa operatsiyasi saqlandi!");
        refreshShift();
      }
      setShiftModalState({ isOpen: false, type: 'open' });
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    }
  };

  // --- Payment Operations ---
  const handlePaymentClick = useCallback(() => {
    if (!shift) return alert("Avval smenani oching!");
    if (cart.length === 0) return alert("Korzina bo'sh!");
    setPaymentModalOpen(true);
  }, [cart, shift]);

  const handleCheckout = async (payments, selectedDebtor) => {
    try {
      const items = cart.map(item => ({
        product_id: item.product_id,
        quantity: item.quantity,
        price: item.price,
        unit: item.unit
      }));

      const payload = {
        items,
        payments,
        shift_id: shift.id,
        comment: "POS Tezkor Savdo"
      };

      const res = await api.post('/cashier/quick-sale', payload);
      
      // If there's nasiya, record the debt
      const nasiyaPayment = payments.find(p => p.method === 'nasiya');
      if (nasiyaPayment && selectedDebtor) {
        await api.post('/debts/record', {
          debtor_id: selectedDebtor.id,
          order_id: res.data.order.id,
          amount: nasiyaPayment.amount,
          type: 'debt',
          description: `POS Savdo #${res.data.order.id}`
        });
      }

      setPaymentModalOpen(false);
      clearCart();
      refreshShift();
      
      // Print check
      try {
        await api.post(`/orders/${res.data.order.id}/print`);
      } catch (err) {
        console.error("Failed to print check:", err);
      }
      
    } catch (err) {
      alert('Xatolik: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleCreateDebtor = async (debtorData) => {
    try {
      const res = await api.post('/debts/debtors', debtorData);
      setDebtors([...debtors, res.data]);
      return res.data;
    } catch (err) {
      alert("Qarzdor yaratishda xatolik: " + (err.response?.data?.error || err.message));
      return null;
    }
  };

  // --- Keyboard Shortcuts ---
  useEffect(() => {
    const handleKeyDown = (e) => {
      // Focus search on CTRL+F or F3
      if ((e.ctrlKey && e.key === 'f') || e.key === 'F3') {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
      // Open payment on F12
      if (e.key === 'F12') {
        e.preventDefault();
        if (cart.length > 0) handlePaymentClick();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
		return () => window.removeEventListener('keydown', handleKeyDown);
	}, [cart, shift, handlePaymentClick]);

  if (loading) return (
    <div className="flex-center h-screen w-full" style={{ background: '#0f172a' }}>
      <div className="spinner-glow"></div>
    </div>
  );

  return (
    <div className="pos-layout">
      {/* Top Header */}
      <header className="pos-header">
        <div className="pos-header-left">
          <div className="pos-logo">MILANO POS</div>
          {shift ? (
            <div className="pos-shift-badge active">
              <span className="dot"></span>
              Smena ochiq: {new Date(shift.opened_at).toLocaleTimeString('ru-RU')}
            </div>
          ) : (
            <div className="pos-shift-badge inactive">
              <span className="dot"></span>
              Smena yopiq
            </div>
          )}
        </div>

        <div className="pos-header-center">
          <div className="pos-search">
            <Search size={20} className="search-icon" />
            <input 
              ref={searchInputRef}
              type="text" 
              placeholder="Mahsulot qidirish... (Ctrl+F)" 
              value={searchQuery}
              onChange={e => setSearchQuery(e.target.value)}
            />
            {searchQuery && (
              <button className="clear-search" onClick={() => setSearchQuery('')}><X size={16}/></button>
            )}
          </div>
        </div>

        <div className="pos-header-right">
          <span className="pos-cashier-name">👤 {user?.full_name || 'Kassir'}</span>
          <button className="pos-icon-btn" onClick={() => setShiftModalState({ isOpen: true, type: 'operation' })} title="Kassa Operatsiyasi" disabled={!shift}>
            <Banknote size={20} />
          </button>
          {shift ? (
            <button className="pos-btn-danger" onClick={() => setShiftModalState({ isOpen: true, type: 'close' })}>
              <Power size={18} /> Smenani Yopish
            </button>
          ) : (
            <button className="pos-btn-success" onClick={() => setShiftModalState({ isOpen: true, type: 'open' })}>
              <Power size={18} /> Smenani Ochish
            </button>
          )}
          <button className="pos-icon-btn logout" onClick={logout} title="Chiqish"><Lock size={20} /></button>
        </div>
      </header>

      <div className="pos-body">
        {/* Left Side: Products (70%) */}
        <div className="pos-products-area">
          {!shift && (
            <div className="pos-overlay">
              <div className="pos-overlay-card">
                <Lock size={48} className="text-gray-400 mb-4" />
                <h2>Smena yopiq</h2>
                <p>Savdoni boshlash uchun avval smenani oching</p>
                <button className="pos-btn-success mt-4 w-full justify-center text-lg py-3" onClick={() => setShiftModalState({ isOpen: true, type: 'open' })}>
                  Smenani Ochish
                </button>
              </div>
            </div>
          )}

          {!searchQuery && (
            <div className="pos-categories no-scrollbar">
              {categories.map(c => (
                <button 
                  key={c.id} 
                  className={`pos-cat-btn ${activeCategory === c.id ? 'active' : ''}`}
                  onClick={() => setActiveCategory(c.id)}
                >
                  {c.name}
                </button>
              ))}
            </div>
          )}

          <div className="pos-grid">
            {products.filter(p => {
              if (!p.is_active) return false;
              if (searchQuery) return p.name.toLowerCase().includes(searchQuery.toLowerCase());
              return p.category_id === activeCategory;
            }).map(p => {
              const inCart = cart.find(c => c.product_id === p.id);
              const imgUrl = p.image_url.startsWith('/') ? api.defaults.baseURL.replace('/api', '') + p.image_url : p.image_url;
              return (
                <div 
                  key={p.id} 
                  className={`pos-item-card ${inCart ? 'selected' : ''}`}
                  onClick={() => addToCart(p)}
                >
                  <div className="pos-item-img" style={{ backgroundImage: `url(${imgUrl})` }}>
                    {inCart && <div className="pos-item-badge">{inCart.quantity}</div>}
                  </div>
                  <div className="pos-item-info">
                    <h4>{p.name}</h4>
                    <span>{p.price.toLocaleString()} so'm</span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>

        {/* Right Side: Cart (30%) */}
        <div className="pos-cart-area">
          <div className="pos-cart-header">
            <h3>🛒 Joriy Chek</h3>
            {cart.length > 0 && (
              <button className="pos-clear-cart" onClick={clearCart}>Tozalash</button>
            )}
          </div>

          <div className="pos-cart-items no-scrollbar">
            {cart.length === 0 ? (
              <div className="pos-cart-empty">
                <ShoppingCart size={48} />
                <p>Korzina bo'sh<br/>Mahsulot tanlang</p>
              </div>
            ) : (
              <AnimatePresence>
                {cart.map(item => (
                  <motion.div 
                    layout
                    initial={{ opacity: 0, x: 20 }}
                    animate={{ opacity: 1, x: 0 }}
                    exit={{ opacity: 0, height: 0, marginBottom: 0, overflow: 'hidden' }}
                    key={item.product_id} 
                    className="pos-cart-item"
                  >
                    <div className="pos-cart-item-info">
                      <div className="name" title={item.name}>{item.name}</div>
                      <div className="price">{(item.price * item.quantity).toLocaleString()} so'm</div>
                    </div>
                    <div className="pos-cart-item-actions">
                      <button className="qty-btn" onClick={() => updateQuantity(item.product_id, -item.step)}><Minus size={14}/></button>
                      <span className="qty-val">{item.quantity}</span>
                      <button className="qty-btn" onClick={() => updateQuantity(item.product_id, item.step)}><Plus size={14}/></button>
                      <button className="del-btn" onClick={() => removeFromCart(item.product_id)}><Trash2 size={16}/></button>
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            )}
          </div>

          <div className="pos-cart-footer">
            <div className="pos-cart-total">
              <span>Jami:</span>
              <strong>{cartTotal.toLocaleString()} so'm</strong>
            </div>
            
            <button 
              className={`pos-checkout-btn ${cart.length === 0 || !shift ? 'disabled' : ''}`}
              onClick={handlePaymentClick}
              disabled={cart.length === 0 || !shift}
            >
              <Banknote size={24} />
              <span>To'lov (F12)</span>
            </button>
          </div>
        </div>
      </div>

      <ShiftModal 
        isOpen={shiftModalState.isOpen}
        type={shiftModalState.type}
        currentShift={shift}
        onClose={() => setShiftModalState({ isOpen: false, type: 'open' })}
        onConfirm={handleShiftConfirm}
      />

      <PaymentModal 
        isOpen={paymentModalOpen}
        onClose={() => setPaymentModalOpen(false)}
        totalAmount={cartTotal}
        onConfirm={handleCheckout}
        debtors={debtors}
        onCreateDebtor={handleCreateDebtor}
      />

      {/* Embedded CSS for POS Layout */}
      <style>{`
        .pos-layout { display: flex; flex-direction: column; height: 100vh; background: #0f172a; overflow: hidden; font-family: 'Inter', sans-serif; }
        .pos-header { height: 70px; background: #1e293b; display: flex; align-items: center; justify-content: space-between; padding: 0 1.5rem; border-bottom: 1px solid #334155; }
        .pos-header-left, .pos-header-right { display: flex; align-items: center; gap: 1rem; }
        .pos-logo { font-size: 1.25rem; font-weight: 800; color: #fff; letter-spacing: 1px; }
        .pos-shift-badge { display: flex; align-items: center; gap: 6px; padding: 6px 12px; border-radius: 20px; font-size: 0.85rem; font-weight: 600; }
        .pos-shift-badge.active { background: rgba(16,185,129,0.15); color: #34d399; border: 1px solid rgba(16,185,129,0.3); }
        .pos-shift-badge.inactive { background: rgba(239,68,68,0.15); color: #f87171; border: 1px solid rgba(239,68,68,0.3); }
        .pos-shift-badge .dot { width: 8px; height: 8px; border-radius: 50%; background: currentColor; }
        
        .pos-header-center { flex: 1; max-width: 500px; margin: 0 2rem; }
        .pos-search { position: relative; width: 100%; }
        .pos-search .search-icon { position: absolute; left: 1rem; top: 50%; transform: translateY(-50%); color: #64748b; }
        .pos-search input { width: 100%; background: #0f172a; border: 1px solid #334155; color: #f8fafc; padding: 0.75rem 1rem 0.75rem 2.75rem; border-radius: 12px; font-size: 1rem; outline: none; transition: all 0.2s; }
        .pos-search input:focus { border-color: #3b82f6; box-shadow: 0 0 0 2px rgba(59,130,246,0.2); }
        .pos-search .clear-search { position: absolute; right: 1rem; top: 50%; transform: translateY(-50%); background: none; border: none; color: #64748b; cursor: pointer; }
        
        .pos-cashier-name { color: #cbd5e1; font-weight: 500; font-size: 0.95rem; }
        .pos-icon-btn { width: 40px; height: 40px; border-radius: 10px; background: #334155; border: 1px solid #475569; color: #f8fafc; display: flex; align-items: center; justify-content: center; cursor: pointer; transition: all 0.2s; }
        .pos-icon-btn:hover { background: #475569; }
        .pos-icon-btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .pos-icon-btn.logout:hover { background: #ef4444; border-color: #ef4444; }
        
        .pos-btn-success { display: flex; align-items: center; gap: 8px; background: #10b981; color: white; padding: 0.6rem 1.25rem; border-radius: 10px; border: none; font-weight: 600; cursor: pointer; transition: all 0.2s; }
        .pos-btn-success:hover { background: #059669; }
        .pos-btn-danger { display: flex; align-items: center; gap: 8px; background: #ef4444; color: white; padding: 0.6rem 1.25rem; border-radius: 10px; border: none; font-weight: 600; cursor: pointer; transition: all 0.2s; }
        .pos-btn-danger:hover { background: #dc2626; }
        
        .pos-body { display: flex; flex: 1; overflow: hidden; }
        
        /* Left Side */
        .pos-products-area { flex: 7; display: flex; flex-direction: column; position: relative; padding: 1rem 0 1rem 1.5rem; }
        .pos-overlay { position: absolute; inset: 0; background: rgba(15,23,42,0.8); backdrop-filter: blur(8px); z-index: 10; display: flex; align-items: center; justify-content: center; }
        .pos-overlay-card { background: #1e293b; padding: 2.5rem; border-radius: 20px; border: 1px solid #334155; text-align: center; max-width: 400px; box-shadow: 0 25px 50px -12px rgba(0,0,0,0.5); }
        .pos-overlay-card h2 { color: #f8fafc; font-size: 1.5rem; margin-bottom: 0.5rem; }
        .pos-overlay-card p { color: #94a3b8; margin-bottom: 1.5rem; }
        
        .pos-categories { display: flex; gap: 0.75rem; overflow-x: auto; padding-bottom: 1rem; margin-bottom: 0.5rem; }
        .pos-cat-btn { white-space: nowrap; padding: 0.75rem 1.5rem; background: #1e293b; color: #cbd5e1; border: 1px solid #334155; border-radius: 12px; font-weight: 600; font-size: 0.95rem; cursor: pointer; transition: all 0.2s; }
        .pos-cat-btn:hover { background: #334155; }
        .pos-cat-btn.active { background: #3b82f6; color: white; border-color: #3b82f6; box-shadow: 0 4px 12px rgba(59,130,246,0.3); }
        
        .pos-grid { flex: 1; overflow-y: auto; display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 1rem; padding-right: 1.5rem; padding-bottom: 2rem; align-content: start; }
        .pos-item-card { background: #1e293b; border: 1px solid #334155; border-radius: 16px; overflow: hidden; cursor: pointer; transition: all 0.2s; position: relative; user-select: none; }
        .pos-item-card:hover { transform: translateY(-2px); border-color: #475569; box-shadow: 0 10px 20px rgba(0,0,0,0.2); }
        .pos-item-card:active { transform: scale(0.98); }
        .pos-item-card.selected { border-color: #3b82f6; }
        .pos-item-img { height: 120px; background-size: cover; background-position: center; position: relative; }
        .pos-item-badge { position: absolute; top: 0.5rem; right: 0.5rem; background: #3b82f6; color: white; width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 0.9rem; box-shadow: 0 2px 8px rgba(0,0,0,0.3); }
        .pos-item-info { padding: 1rem; }
        .pos-item-info h4 { margin: 0 0 0.5rem 0; color: #f8fafc; font-size: 0.95rem; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden; line-height: 1.3; }
        .pos-item-info span { color: #34d399; font-weight: 700; font-size: 1.05rem; }
        
        /* Right Side (Cart) */
        .pos-cart-area { flex: 3; background: #ffffff; display: flex; flex-direction: column; border-left: 1px solid #e2e8f0; min-width: 350px; max-width: 450px; }
        .pos-cart-header { padding: 1.25rem 1.5rem; border-bottom: 1px solid #e2e8f0; display: flex; justify-content: space-between; align-items: center; background: #f8fafc; }
        .pos-cart-header h3 { margin: 0; color: #0f172a; font-size: 1.2rem; font-weight: 800; }
        .pos-clear-cart { background: transparent; border: none; color: #ef4444; font-weight: 600; cursor: pointer; font-size: 0.9rem; padding: 4px 8px; border-radius: 6px; }
        .pos-clear-cart:hover { background: #fef2f2; }
        
        .pos-cart-items { flex: 1; overflow-y: auto; padding: 1rem; display: flex; flex-direction: column; gap: 0.75rem; }
        .pos-cart-empty { height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; color: #94a3b8; text-align: center; gap: 1rem; }
        .pos-cart-item { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; padding: 1rem; display: flex; flex-direction: column; gap: 0.75rem; box-shadow: 0 2px 4px rgba(0,0,0,0.02); }
        .pos-cart-item-info { display: flex; justify-content: space-between; align-items: flex-start; gap: 1rem; }
        .pos-cart-item-info .name { color: #1e293b; font-weight: 600; font-size: 0.95rem; line-height: 1.3; }
        .pos-cart-item-info .price { color: #3b82f6; font-weight: 700; white-space: nowrap; }
        .pos-cart-item-actions { display: flex; align-items: center; justify-content: space-between; background: #f8fafc; padding: 6px; border-radius: 8px; }
        .qty-btn { width: 32px; height: 32px; border-radius: 8px; background: #fff; border: 1px solid #cbd5e1; display: flex; align-items: center; justify-content: center; cursor: pointer; color: #475569; transition: all 0.1s; }
        .qty-btn:hover { border-color: #3b82f6; color: #3b82f6; }
        .qty-btn:active { background: #eff6ff; transform: scale(0.95); }
        .qty-val { font-weight: 700; color: #0f172a; width: 40px; text-align: center; font-size: 1.1rem; }
        .del-btn { width: 32px; height: 32px; border-radius: 8px; background: #fef2f2; border: 1px solid #fecaca; color: #ef4444; display: flex; align-items: center; justify-content: center; cursor: pointer; margin-left: auto; transition: all 0.1s; }
        .del-btn:hover { background: #fee2e2; }
        
        .pos-cart-footer { padding: 1.5rem; background: #fff; border-top: 1px solid #e2e8f0; box-shadow: 0 -4px 20px rgba(0,0,0,0.05); }
        .pos-cart-total { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.25rem; }
        .pos-cart-total span { color: #64748b; font-size: 1.1rem; font-weight: 600; }
        .pos-cart-total strong { color: #0f172a; font-size: 1.75rem; font-weight: 800; }
        .pos-checkout-btn { width: 100%; padding: 1.25rem; border-radius: 16px; background: linear-gradient(135deg, #3b82f6, #2563eb); color: white; border: none; font-size: 1.25rem; font-weight: 800; display: flex; align-items: center; justify-content: center; gap: 0.75rem; cursor: pointer; transition: all 0.2s; box-shadow: 0 10px 25px rgba(59,130,246,0.4); }
        .pos-checkout-btn:hover:not(.disabled) { transform: translateY(-2px); box-shadow: 0 15px 30px rgba(59,130,246,0.5); }
        .pos-checkout-btn:active:not(.disabled) { transform: translateY(0); }
        .pos-checkout-btn.disabled { background: #cbd5e1; color: #94a3b8; box-shadow: none; cursor: not-allowed; }
      `}</style>
    </div>
  );
};

export default Cashier;

import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useCartStore } from '../store/cartStore';
import { motion, AnimatePresence } from 'framer-motion';
import { Trash2, Plus, Minus, ArrowRight, ShoppingBag, ArrowLeft } from 'lucide-react';

const Cart = () => {
  const { items, removeItem, updateQuantity, getTotal } = useCartStore();
  const navigate = useNavigate();

  const getImageUrl = (url) => {
    if (!url) return null;
    if (url.startsWith('/')) return url;
    return url;
  };

  if (items.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="empty-cart-wrap"
      >
        <div className="empty-cart-icon">🛒</div>
        <h2>Savatingiz bo'sh</h2>
        <p>Sevimli taomlaringizni tanlang va buyurtma bering!</p>
        <Link to="/" className="btn-primary empty-cart-btn">
          <ArrowLeft size={18} /> Menyuga o'tish
        </Link>
      </motion.div>
    );
  }

  return (
    <div className="cart-page animate-fade">
      {/* Header */}
      <div className="cart-head">
        <div>
          <h1>Savat</h1>
          <p className="cart-subtitle">{items.length} xil mahsulot</p>
        </div>
        <button className="clear-all-btn" onClick={() => navigate('/')}>
          <ArrowLeft size={16} /> Davom etish
        </button>
      </div>

      <div className="cart-layout">
        {/* Items */}
        <div className="cart-items-list">
          <AnimatePresence mode="popLayout">
            {items.map((item, i) => (
              <motion.div
                key={item.id}
                layout
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0, transition: { delay: i * 0.05 } }}
                exit={{ opacity: 0, x: 20, scale: 0.95 }}
                className="cart-item-card"
              >
                {/* Image */}
                <div className="cart-item-img">
                  {getImageUrl(item.image_url) ? (
                    <img
                      src={getImageUrl(item.image_url)}
                      alt={item.name}
                      onError={e => {
                        e.target.onerror = null;
                        e.target.parentElement.innerHTML = '<div class="cart-img-ph">🍽</div>';
                      }}
                    />
                  ) : (
                    <div className="cart-img-ph">🍽</div>
                  )}
                </div>

                {/* Info */}
                <div className="cart-item-info">
                  <h3>{item.name}</h3>
                  <span className="cart-unit-price">{item.price.toLocaleString()} so'm / dona</span>
                </div>

                {/* Qty controls */}
                <div className="qty-ctrl">
                  <button className="qty-btn" onClick={() => updateQuantity(item.id, -1)}>
                    <Minus size={14} />
                  </button>
                  <span className="qty-num">{item.quantity}</span>
                  <button className="qty-btn plus" onClick={() => updateQuantity(item.id, 1)}>
                    <Plus size={14} />
                  </button>
                </div>

                {/* Line total */}
                <div className="line-total">
                  {(item.price * item.quantity).toLocaleString()} <small>so'm</small>
                </div>

                {/* Remove */}
                <button className="remove-btn" onClick={() => removeItem(item.id)} title="O'chirish">
                  <Trash2 size={16} />
                </button>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        {/* Summary */}
        <div className="cart-summary">
          <div className="summary-card">
            <h2 className="summary-title">Buyurtma xulosasi</h2>
            
            <div className="divider" />

            <div className="summary-rows">
              <div className="summary-row">
                <span>Mahsulotlar ({items.reduce((s,i) => s + i.quantity, 0)} ta)</span>
                <span>{getTotal().toLocaleString()} so'm</span>
              </div>
              <div className="summary-row">
                <span>Yetkazib berish</span>
                <span style={{ color: 'var(--success)' }}>15 000 so'm</span>
              </div>
            </div>

            <div className="divider" />

            <div className="summary-total">
              <span>Jami</span>
              <span className="total-price">{(getTotal() + 15000).toLocaleString()} so'm</span>
            </div>

            <button
              className="btn-primary checkout-btn"
              onClick={() => navigate('/checkout')}
            >
              Buyurtma berish <ArrowRight size={18} />
            </button>

            <Link to="/" className="continue-btn">
              ← Xarid qilishni davom eting
            </Link>
          </div>
        </div>
      </div>

      <style>{`
        .cart-page { padding-bottom: 2rem; }

        .cart-head {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 2rem;
        }

        .cart-head h1 { font-size: 2rem; }

        .cart-subtitle {
          color: var(--text-secondary);
          font-size: 0.9rem;
          margin-top: 0.2rem;
        }

        .clear-all-btn {
          background: var(--bg-surface);
          border: 1px solid var(--border);
          color: var(--text-secondary);
          padding: 0.55rem 1rem;
          font-size: 0.85rem;
          display: flex;
          align-items: center;
          gap: 0.4rem;
          border-radius: var(--radius-sm);
        }

        .clear-all-btn:hover { border-color: var(--primary); color: var(--primary); }

        /* Layout */
        .cart-layout {
          display: grid;
          grid-template-columns: 1fr 340px;
          gap: 1.75rem;
          align-items: start;
        }

        /* Items */
        .cart-items-list {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
        }

        .cart-item-card {
          display: flex;
          align-items: center;
          gap: 1rem;
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius);
          padding: 0.9rem 1.1rem;
          transition: var(--transition);
          backdrop-filter: blur(12px);
        }

        .cart-item-card:hover {
          border-color: rgba(249,115,22,0.25);
          background: var(--bg-card-hover);
        }

        .cart-item-img {
          width: 68px; height: 68px;
          border-radius: 10px;
          overflow: hidden;
          flex-shrink: 0;
          background: rgba(255,255,255,0.04);
        }

        .cart-item-img img {
          width: 100%; height: 100%;
          object-fit: cover;
        }

        .cart-img-ph {
          width: 100%; height: 100%;
          display: flex; align-items: center; justify-content: center;
          font-size: 2rem;
          opacity: 0.4;
        }

        .cart-item-info { flex: 1; min-width: 0; }

        .cart-item-info h3 {
          font-size: 0.95rem;
          font-family: var(--font);
          font-weight: 700;
          margin-bottom: 0.25rem;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .cart-unit-price {
          font-size: 0.78rem;
          color: var(--text-secondary);
        }

        /* Qty */
        .qty-ctrl {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: rgba(255,255,255,0.05);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 0.35rem 0.5rem;
          flex-shrink: 0;
        }

        .qty-btn {
          background: none;
          color: var(--text-secondary);
          width: 22px; height: 22px;
          display: flex; align-items: center; justify-content: center;
          border-radius: 4px;
          transition: var(--transition);
        }

        .qty-btn:hover { background: rgba(249,115,22,0.15); color: var(--primary); }
        .qty-btn.plus { color: var(--primary); }

        .qty-num {
          min-width: 22px;
          text-align: center;
          font-weight: 800;
          font-size: 0.95rem;
        }

        /* Line total */
        .line-total {
          min-width: 110px;
          text-align: right;
          font-weight: 800;
          font-size: 0.95rem;
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
          flex-shrink: 0;
        }

        .line-total small {
          font-size: 0.72rem;
          -webkit-text-fill-color: var(--text-secondary);
          color: var(--text-secondary);
        }

        /* Remove */
        .remove-btn {
          background: none;
          color: var(--text-muted);
          padding: 0.4rem;
          border-radius: 6px;
          flex-shrink: 0;
          transition: var(--transition);
        }

        .remove-btn:hover { color: var(--danger); background: rgba(239,68,68,0.10); }

        /* Summary */
        .summary-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius-lg);
          padding: 1.75rem;
          position: sticky;
          top: 88px;
          backdrop-filter: blur(16px);
          -webkit-backdrop-filter: blur(16px);
          box-shadow: var(--shadow-card);
        }

        .summary-title {
          font-size: 1.15rem;
          margin-bottom: 1rem;
          color: var(--text-primary);
        }

        .summary-rows { display: flex; flex-direction: column; gap: 0.85rem; margin: 1rem 0; }

        .summary-row {
          display: flex;
          justify-content: space-between;
          font-size: 0.9rem;
          color: var(--text-secondary);
        }

        .summary-total {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin: 1rem 0 1.5rem;
          font-weight: 700;
        }

        .total-price {
          font-size: 1.25rem;
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        .checkout-btn {
          width: 100%;
          padding: 0.9rem;
          display: flex; align-items: center; justify-content: center; gap: 0.6rem;
          font-size: 0.95rem;
          margin-bottom: 1rem;
        }

        .continue-btn {
          display: block;
          text-align: center;
          font-size: 0.82rem;
          color: var(--text-secondary);
          transition: var(--transition);
          padding: 0.5rem;
        }

        .continue-btn:hover { color: var(--primary); }

        /* Empty */
        .empty-cart-wrap {
          text-align: center;
          padding: 5rem 2rem;
          max-width: 420px;
          margin: 0 auto;
        }

        .empty-cart-icon {
          font-size: 5rem;
          margin-bottom: 1.5rem;
          animation: float 3s ease-in-out infinite;
        }

        @keyframes float {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-12px); }
        }

        .empty-cart-wrap h2 { font-size: 1.75rem; margin-bottom: 0.75rem; }
        .empty-cart-wrap p { color: var(--text-secondary); margin-bottom: 2rem; }

        .empty-cart-btn {
          display: inline-flex;
          align-items: center;
          gap: 0.5rem;
          text-decoration: none;
          padding: 0.8rem 1.75rem;
          border-radius: var(--radius-sm);
          font-size: 0.95rem;
        }

        @media (max-width: 900px) {
          .cart-layout { grid-template-columns: 1fr; }
          .summary-card { position: static; }
        }

        @media (max-width: 640px) {
          .cart-head { flex-wrap: wrap; gap: 0.75rem; }
          .cart-head h1 { font-size: 1.5rem; }
          .cart-item-card {
            flex-wrap: wrap;
            gap: 0.75rem;
          }
          .cart-item-img { width: 56px; height: 56px; }
          .cart-item-info { min-width: 0; flex: 1; }
          .line-total { min-width: auto; }
          .qty-ctrl { padding: 0.25rem 0.4rem; }
          .summary-card { padding: 1.25rem; }
          .total-price { font-size: 1.1rem; }
          .checkout-btn { padding: 0.8rem; font-size: 0.9rem; }
        }

        @media (max-width: 420px) {
          .cart-item-card { padding: 0.75rem; }
          .cart-item-img { width: 48px; height: 48px; border-radius: 8px; }
          .remove-btn { padding: 0.3rem; }
          .line-total { font-size: 0.85rem; }
        }
      `}</style>
    </div>
  );
};

export default Cart;

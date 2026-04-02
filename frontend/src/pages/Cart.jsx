import React from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useCartStore } from '../store/cartStore';
import { motion, AnimatePresence } from 'framer-motion';
import { Trash2, Plus, Minus, ArrowRight, ShoppingBag } from 'lucide-react';

const Cart = () => {
  const { items, removeItem, updateQuantity, getTotal } = useCartStore();
  const navigate = useNavigate();

  const getImageUrl = (url) => {
    if (!url) return 'https://via.placeholder.com/100x100?text=Taom';
    if (url.startsWith('/')) return `http://localhost:8080${url}`;
    return url;
  };

  if (items.length === 0) {
    return (
      <motion.div 
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        className="premium-card empty-cart glass"
      >
        <div className="empty-illustration">
          <div className="icon-glow"></div>
          <ShoppingBag size={80} className="empty-icon-modern" />
        </div>
        <h2>Savatingiz bo'sh</h2>
        <p className="empty-desc">Sevimli taomlaringizni tanlang va tasavvuringizdagi mazani kashf eting. Biz ularni tezda yetkazib beramiz!</p>
        <Link to="/" className="btn-modern-primary back-btn">Menyuga o'tish <ArrowRight size={20} /></Link>
      </motion.div>
    );
  }

  return (
    <div className="cart-page animate-fade">
      <div className="cart-header">
        <h1>Savat</h1>
        <p>{items.length} xil mahsulot</p>
      </div>

      <div className="cart-content">
        <div className="cart-items">
          <AnimatePresence mode='popLayout'>
            {items.map(item => (
              <motion.div 
                key={item.id}
                layout
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: 20 }}
                className="premium-card cart-item"
              >
                <div className="item-img">
                  <img 
                    src={getImageUrl(item.image_url)} 
                    alt={item.name} 
                    onError={(e) => {
                      e.target.onerror = null;
                      e.target.src = 'https://via.placeholder.com/100x100?text=Taom';
                    }}
                  />
                </div>
                <div className="item-info">
                  <h3>{item.name}</h3>
                  <div className="item-price">{item.price.toLocaleString()} so'm</div>
                </div>
                <div className="item-controls">
                  <button onClick={() => updateQuantity(item.id, -1)} className="control-btn"><Minus size={16} /></button>
                  <span className="qty">{item.quantity}</span>
                  <button onClick={() => updateQuantity(item.id, 1)} className="control-btn"><Plus size={16} /></button>
                </div>
                <div className="item-total">
                  {(item.price * item.quantity).toLocaleString()} so'm
                </div>
                <button onClick={() => removeItem(item.id)} className="remove-btn">
                  <Trash2 size={20} />
                </button>
              </motion.div>
            ))}
          </AnimatePresence>
        </div>

        <div className="cart-summary premium-card glass">
          <h2>Umumiy</h2>
          <div className="summary-row">
            <span>Mahsulotlar</span>
            <span>{getTotal().toLocaleString()} so'm</span>
          </div>
          <div className="summary-row">
            <span>Yetkazib berish</span>
            <span>15,000 so'm</span>
          </div>
          <div className="summary-total">
            <span>Jami</span>
            <span>{(getTotal() + 15000).toLocaleString()} so'm</span>
          </div>
          <button className="btn-primary checkout-btn" onClick={() => navigate('/checkout')}>
            Buyurtma berish <ArrowRight size={20} />
          </button>
        </div>
      </div>

      <style>{`
        .cart-page {
          padding-top: 1rem;
        }

        .cart-header {
          margin-bottom: 2rem;
        }

        .cart-content {
          display: grid;
          grid-template-columns: 1fr 350px;
          gap: 2rem;
          align-items: start;
        }

        .cart-items {
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }

        .cart-item {
          display: flex;
          align-items: center;
          gap: 1.5rem;
          padding: 1rem;
        }

        .item-img img {
          width: 80px;
          height: 80px;
          border-radius: 12px;
          object-fit: cover;
        }

        .item-info {
          flex: 1;
        }

        .item-info h3 {
          font-size: 1.1rem;
          margin-bottom: 0.25rem;
        }

        .item-price {
          color: var(--text-dim);
          font-size: 0.9rem;
        }

        .item-controls {
          display: flex;
          align-items: center;
          gap: 1rem;
          background: var(--bg-dark);
          padding: 0.5rem;
          border-radius: 10px;
        }

        .control-btn {
          background: none;
          color: var(--text-main);
          width: 24px;
          height: 24px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        .qty {
          width: 20px;
          text-align: center;
          font-weight: 600;
        }

        .item-total {
          width: 120px;
          text-align: right;
          font-weight: 700;
          color: var(--primary);
        }

        .remove-btn {
          background: none;
          color: #ef4444;
          padding: 0.5rem;
        }

        .cart-summary {
          position: sticky;
          top: 90px;
          padding: 2rem;
        }

        .cart-summary h2 {
          margin-bottom: 1.5rem;
        }

        .summary-row {
          display: flex;
          justify-content: space-between;
          color: var(--text-dim);
          margin-bottom: 1rem;
        }

        .summary-total {
          display: flex;
          justify-content: space-between;
          border-top: 1px solid var(--border);
          padding-top: 1.5rem;
          margin-top: 1.5rem;
          font-size: 1.25rem;
          font-weight: 700;
          color: white;
          margin-bottom: 2rem;
        }

        .checkout-btn {
          width: 100%;
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 1rem;
          padding: 1rem;
          font-size: 1.1rem;
        }

        .empty-cart {
          text-align: center;
          padding: 5rem 3rem;
          max-width: 500px;
          margin: 4rem auto;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          border: 1px solid rgba(255, 255, 255, 0.05);
          background: linear-gradient(145deg, rgba(30, 41, 59, 0.4), rgba(15, 23, 42, 0.6));
          box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
          border-radius: 24px;
        }

        .empty-illustration {
          position: relative;
          margin-bottom: 2.5rem;
          display: flex;
          justify-content: center;
          align-items: center;
        }

        .icon-glow {
          position: absolute;
          width: 120px;
          height: 120px;
          background: var(--primary);
          filter: blur(50px);
          opacity: 0.2;
          border-radius: 50%;
          animation: pulse 3s infinite ease-in-out;
        }

        .empty-icon-modern {
          color: white;
          z-index: 1;
          filter: drop-shadow(0 10px 15px rgba(0,0,0,0.3));
          animation: float 4s ease-in-out infinite;
        }

        .empty-cart h2 {
          font-size: 2rem;
          font-weight: 800;
          margin-bottom: 1rem;
          background: linear-gradient(to right, #ffffff, #94a3b8);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }

        .empty-desc {
          font-size: 1.1rem;
          color: var(--text-dim);
          margin-bottom: 2.5rem;
          line-height: 1.6;
        }

        .btn-modern-primary {
          background: linear-gradient(135deg, var(--primary), #4f46e5);
          color: white;
          padding: 1rem 2rem;
          border-radius: 50px;
          text-decoration: none;
          font-weight: 700;
          font-size: 1.1rem;
          display: flex;
          align-items: center;
          gap: 0.75rem;
          transition: all 0.3s ease;
          box-shadow: 0 10px 25px -5px rgba(99, 102, 241, 0.4);
        }

        .btn-modern-primary:hover {
          transform: translateY(-3px);
          box-shadow: 0 15px 35px -5px rgba(99, 102, 241, 0.6);
        }

        @keyframes float {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-10px); }
        }

        @keyframes pulse {
          0%, 100% { opacity: 0.2; transform: scale(1); }
          50% { opacity: 0.4; transform: scale(1.2); }
        }

        @media (max-width: 992px) {
          .cart-content {
            grid-template-columns: 1fr;
          }
        }
      `}</style>
    </div>
  );
};

export default Cart;

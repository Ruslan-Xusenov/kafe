import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Package, Clock, CheckCircle2, Star, History, RefreshCcw, Truck, XCircle, ChefHat } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import RatingModal from '../components/RatingModal';
import { Link } from 'react-router-dom';

const STATUS_MAP = {
  new:       { label: 'Yangi',          icon: <Package size={18} />,     color: '#818cf8', bg: 'rgba(99,102,241,0.15)' },
  preparing: { label: 'Tayyorlanmoqda', icon: <ChefHat size={18} />,     color: '#fbbf24', bg: 'rgba(251,191,36,0.12)' },
  ready:     { label: 'Tayyor',         icon: <CheckCircle2 size={18} />, color: '#34d399', bg: 'rgba(16,185,129,0.12)' },
  on_way:    { label: 'Yo\'lda',        icon: <Truck size={18} />,        color: '#fb923c', bg: 'rgba(249,115,22,0.15)' },
  delivered: { label: 'Yetkazildi',     icon: <CheckCircle2 size={18} />, color: '#a3e635', bg: 'rgba(163,230,53,0.12)' },
  cancelled: { label: 'Bekor qilindi',  icon: <XCircle size={18} />,      color: '#fca5a5', bg: 'rgba(239,68,68,0.12)'  },
};

const MyOrders = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [isRatingOpen, setIsRatingOpen] = useState(false);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await api.get('/orders/my');
      setOrders(Array.isArray(res.data) ? res.data : []);
    } catch (err) {
      console.error(err);
      setOrders([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchOrders(); }, []);

  if (loading) return (
    <div className="loading-screen">
      <div className="spinner" />
      <p style={{ color: 'var(--text-secondary)' }}>Yuklanmoqda...</p>
    </div>
  );

  return (
    <div className="myorders-page animate-fade">
      {/* Header */}
      <div className="myorders-header">
        <div>
          <h1>Buyurtmalar tarixi</h1>
          <p className="myorders-sub">Sizning barcha buyurtmalaringiz va holati</p>
        </div>
        <button className="mo-refresh-btn" onClick={fetchOrders}>
          <RefreshCcw size={16} /> Yangilash
        </button>
      </div>

      {/* Orders list */}
      {orders.length === 0 ? (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mo-empty"
        >
          <div className="mo-empty-icon">📦</div>
          <h2>Buyurtmalar yo'q</h2>
          <p>Hali birorta buyurtma bermagansiz. Menyudan o'zingizga yoqqan taomni tanlang!</p>
          <Link to="/" className="btn-primary mo-empty-btn">
            Menyuga o'tish
          </Link>
        </motion.div>
      ) : (
        <div className="mo-list">
          <AnimatePresence>
            {orders.map((order, idx) => {
              const cfg = STATUS_MAP[order.status] || STATUS_MAP.new;
              return (
                <motion.div
                  key={order.id}
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: idx * 0.05 }}
                  className="mo-card"
                >
                  {/* Top row */}
                  <div className="mo-card-top">
                    <div className="mo-left">
                      <div className="mo-status-icon" style={{ background: cfg.bg, color: cfg.color }}>
                        {cfg.icon}
                      </div>
                      <div>
                        <h3 className="mo-order-id">Buyurtma #{order.id}</h3>
                        <p className="mo-time">
                          {new Date(order.created_at).toLocaleString('uz-UZ', {
                            day: '2-digit', month: 'short', year: 'numeric',
                            hour: '2-digit', minute: '2-digit'
                          })}
                        </p>
                      </div>
                    </div>

                    <div className="mo-right">
                      <span className="mo-badge" style={{ background: cfg.bg, color: cfg.color, borderColor: cfg.color + '55' }}>
                        {cfg.label}
                      </span>
                      <div className="mo-price">{(order.total_price || 0).toLocaleString()} so'm</div>
                    </div>
                  </div>

                  {/* Items */}
                  {order.items && order.items.length > 0 && (
                    <div className="mo-items">
                      {order.items.map((item, i) => (
                        <span key={i} className="mo-item-chip">
                          <span className="mo-item-qty">{item.quantity}×</span>
                          {item.product_name}
                        </span>
                      ))}
                    </div>
                  )}

                  {/* Address */}
                  {order.address && (
                    <div className="mo-address">
                      📍 {order.address}
                    </div>
                  )}

                  {/* Rating button */}
                  {order.status === 'delivered' && (
                    <div className="mo-footer">
                      <button
                        className="mo-rate-btn"
                        onClick={() => { setSelectedOrder(order); setIsRatingOpen(true); }}
                      >
                        <Star size={15} /> Baholash
                      </button>
                    </div>
                  )}
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      )}

      {/* Rating Modal */}
      {selectedOrder && (
        <RatingModal
          isOpen={isRatingOpen}
          onClose={() => setIsRatingOpen(false)}
          order={selectedOrder}
          onSuccess={() => {
            fetchOrders();
            setIsRatingOpen(false);
          }}
        />
      )}

      <style>{`
        .myorders-page { padding-bottom: 3rem; }

        .myorders-header {
          display: flex;
          align-items: flex-start;
          justify-content: space-between;
          gap: 1rem;
          margin-bottom: 2rem;
        }

        .myorders-header h1 { font-size: 2rem; }

        .myorders-sub {
          color: var(--text-secondary);
          font-size: 0.9rem;
          margin-top: 0.3rem;
        }

        .mo-refresh-btn {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          color: var(--text-secondary);
          padding: 0.55rem 1rem;
          border-radius: var(--radius-sm);
          font-size: 0.85rem;
          font-weight: 600;
          transition: var(--transition);
          white-space: nowrap;
          flex-shrink: 0;
        }

        .mo-refresh-btn:hover { border-color: var(--primary); color: var(--primary); }

        /* Empty */
        .mo-empty {
          text-align: center;
          padding: 5rem 2rem;
          max-width: 440px;
          margin: 0 auto;
        }

        .mo-empty-icon {
          font-size: 5rem;
          margin-bottom: 1.5rem;
          animation: float 3s ease-in-out infinite;
        }

        @keyframes float {
          0%, 100% { transform: translateY(0); }
          50% { transform: translateY(-10px); }
        }

        .mo-empty h2 { font-size: 1.6rem; margin-bottom: 0.75rem; }
        .mo-empty p { color: var(--text-secondary); margin-bottom: 2rem; line-height: 1.6; }

        .mo-empty-btn {
          display: inline-flex;
          align-items: center;
          gap: 0.5rem;
          text-decoration: none;
          padding: 0.75rem 1.75rem;
          border-radius: var(--radius-sm);
          font-size: 0.95rem;
        }

        /* List */
        .mo-list {
          display: flex;
          flex-direction: column;
          gap: 0.9rem;
        }

        /* Card */
        .mo-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius);
          padding: 1.2rem 1.4rem;
          display: flex;
          flex-direction: column;
          gap: 0.9rem;
          transition: var(--transition);
          backdrop-filter: blur(12px);
          -webkit-backdrop-filter: blur(12px);
        }

        .mo-card:hover {
          border-color: rgba(249,115,22,0.25);
          background: var(--bg-card-hover);
        }

        .mo-card-top {
          display: flex;
          align-items: center;
          justify-content: space-between;
          gap: 1rem;
        }

        .mo-left {
          display: flex;
          align-items: center;
          gap: 0.85rem;
        }

        .mo-status-icon {
          width: 42px; height: 42px;
          border-radius: 12px;
          display: flex; align-items: center; justify-content: center;
          flex-shrink: 0;
        }

        .mo-order-id {
          font-size: 1rem;
          font-weight: 700;
          font-family: var(--font);
          margin-bottom: 0.15rem;
        }

        .mo-time {
          font-size: 0.78rem;
          color: var(--text-muted);
        }

        .mo-right {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 0.4rem;
          flex-shrink: 0;
        }

        .mo-badge {
          display: inline-flex;
          align-items: center;
          padding: 0.2rem 0.75rem;
          border-radius: 99px;
          font-size: 0.75rem;
          font-weight: 700;
          letter-spacing: 0.04em;
          border: 1px solid;
          text-transform: uppercase;
        }

        .mo-price {
          font-size: 1.05rem;
          font-weight: 800;
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        /* Items */
        .mo-items {
          display: flex;
          flex-wrap: wrap;
          gap: 0.4rem;
        }

        .mo-item-chip {
          display: inline-flex;
          align-items: center;
          gap: 0.3rem;
          background: rgba(255,255,255,0.05);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 0.25rem 0.65rem;
          font-size: 0.8rem;
          color: var(--text-secondary);
        }

        .mo-item-qty {
          color: var(--primary);
          font-weight: 700;
        }

        /* Address */
        .mo-address {
          font-size: 0.82rem;
          color: var(--text-secondary);
          background: rgba(255,255,255,0.03);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 0.45rem 0.75rem;
        }

        /* Footer */
        .mo-footer {
          border-top: 1px solid var(--border);
          padding-top: 0.85rem;
          display: flex;
          justify-content: flex-end;
        }

        .mo-rate-btn {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          background: rgba(251,191,36,0.10);
          border: 1px solid rgba(251,191,36,0.25);
          color: #fbbf24;
          padding: 0.5rem 1rem;
          border-radius: var(--radius-sm);
          font-size: 0.85rem;
          font-weight: 700;
          transition: var(--transition);
        }

        .mo-rate-btn:hover {
          background: rgba(251,191,36,0.20);
          transform: translateY(-1px);
        }

        @media (max-width: 600px) {
          .mo-card-top { flex-direction: column; align-items: flex-start; }
          .mo-right { align-items: flex-start; flex-direction: row; gap: 0.75rem; }
        }
      `}</style>
    </div>
  );
};

export default MyOrders;

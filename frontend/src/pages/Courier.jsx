import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion, AnimatePresence } from 'framer-motion';
import { Truck, MapPin, Phone, CheckCircle, Navigation, RefreshCw, Package } from 'lucide-react';
import StatsSection from '../components/StatsSection';

const Courier = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  useWebsocket((data) => {
    if (data.type === 'status_update' || data.type === 'new_order') fetchOrders();
  });

  useEffect(() => { fetchOrders(); }, []);

  const fetchOrders = async () => {
    try {
      const res = await api.get('/orders/active');
      setOrders((res.data || []).filter(o => o.status === 'ready' || o.status === 'on_way'));
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const pickUp = async (orderId) => {
    try {
      await api.post(`/orders/${orderId}/assign`);
      fetchOrders();
    } catch (err) {
      alert('Buyurtmani qabul qilishda xatolik');
    }
  };

  const deliver = async (orderId) => {
    const prev = [...orders];
    setOrders(o => o.filter(x => x.id !== orderId));
    try {
      await api.put(`/orders/${orderId}/status`, { status: 'delivered' });
      fetchOrders();
    } catch (err) {
      setOrders(prev);
      alert('Statusni yangilashda xatolik');
    }
  };

  const readyOrders = orders.filter(o => o.status === 'ready');
  const activeOrders = orders.filter(o => o.status === 'on_way');

  if (loading) return (
    <div className="loading-screen">
      <div className="spinner" />
      <p>Yuklanmoqda...</p>
    </div>
  );

  return (
    <div className="courier-page animate-fade">
      {/* Header */}
      <div className="courier-header">
        <div className="courier-title-wrap">
          <div className="courier-icon">
            <Truck size={24} />
          </div>
          <div>
            <h1>Kuryer Paneli</h1>
            <p className="courier-subtitle">
              {readyOrders.length} tayyor · {activeOrders.length} yo'lda
            </p>
          </div>
        </div>
        <button
          className={`refresh-btn ${refreshing ? 'refreshing' : ''}`}
          onClick={() => { setRefreshing(true); fetchOrders(); }}
        >
          <RefreshCw size={16} />
        </button>
      </div>

      <StatsSection role="courier" />

      {/* Ready for pickup */}
      <section className="courier-section">
        <div className="section-header">
          <div className="section-pill">
            <Package size={14} />
            Olish mumkin
          </div>
          <span className="section-count">{readyOrders.length} ta</span>
        </div>

        {readyOrders.length === 0 ? (
          <div className="empty-section">
            <span>📦</span>
            <p>Hozircha tayyor buyurtmalar yo'q</p>
          </div>
        ) : (
          <div className="courier-grid">
            <AnimatePresence>
              {readyOrders.map(order => (
                <motion.div
                  key={order.id}
                  layout
                  initial={{ opacity: 0, y: 16 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.9 }}
                  className="courier-card ready-card"
                >
                  <div className="ccard-top">
                    <span className="ccard-id">#{order.id}</span>
                    <span className="ccard-price">{order.total_price?.toLocaleString()} so'm</span>
                  </div>
                  <div className="ccard-addr">
                    <MapPin size={15} />
                    <span>{order.address}</span>
                  </div>
                  <button className="btn-primary ccard-btn" onClick={() => pickUp(order.id)}>
                    <Navigation size={16} /> Olib ketish
                  </button>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </section>

      {/* Active deliveries */}
      <section className="courier-section">
        <div className="section-header">
          <div className="section-pill" style={{ background: 'rgba(249,115,22,0.12)', borderColor: 'rgba(249,115,22,0.25)', color: 'var(--primary)' }}>
            <Truck size={14} />
            Faol yetkazishlar
          </div>
          <span className="section-count">{activeOrders.length} ta</span>
        </div>

        {activeOrders.length === 0 ? (
          <div className="empty-section">
            <span>🚴</span>
            <p>Hozircha faol buyurtmalar yo'q</p>
          </div>
        ) : (
          <div className="courier-grid">
            <AnimatePresence>
              {activeOrders.map(order => (
                <motion.div
                  key={order.id}
                  layout
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="courier-card active-card"
                >
                  <div className="ccard-top">
                    <span className="ccard-id">#{order.id}</span>
                    <span className="on-way-badge">🚴 Yo'lda</span>
                  </div>

                  <div className="ccard-details">
                    <div className="ccard-addr">
                      <MapPin size={15} />
                      <span>{order.address}</span>
                    </div>
                    <div className="ccard-addr">
                      <Phone size={15} />
                      <a href={`tel:${order.phone}`} style={{ color: 'var(--primary)' }}>{order.phone}</a>
                    </div>
                  </div>

                  {order.items && order.items.length > 0 && (
                    <div className="ccard-items">
                      {order.items.map((item, i) => (
                        <span key={i} className="item-chip">
                          {item.quantity}× {item.product_name}
                        </span>
                      ))}
                    </div>
                  )}

                  <button className="ccard-delivered-btn ccard-btn" onClick={() => deliver(order.id)}>
                    <CheckCircle size={16} /> Yetkazib berildi
                  </button>
                </motion.div>
              ))}
            </AnimatePresence>
          </div>
        )}
      </section>

      <style>{`
        .courier-page { padding-bottom: 3rem; }

        .courier-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 2rem;
        }

        .courier-title-wrap {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .courier-icon {
          width: 52px; height: 52px;
          background: var(--grad-brand);
          border-radius: 14px;
          display: flex; align-items: center; justify-content: center;
          color: white;
          box-shadow: 0 6px 20px rgba(249,115,22,0.35);
          flex-shrink: 0;
        }

        .courier-header h1 { font-size: 1.75rem; }

        .courier-subtitle {
          color: var(--text-secondary);
          font-size: 0.85rem;
          margin-top: 0.2rem;
        }

        .refresh-btn {
          width: 40px; height: 40px;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          color: var(--text-secondary);
          display: flex; align-items: center; justify-content: center;
          border-radius: 10px;
          transition: var(--transition);
        }

        .refresh-btn:hover { border-color: var(--primary); color: var(--primary); }
        .refresh-btn.refreshing svg { animation: spin 1s linear infinite; }

        .courier-section { margin-bottom: 2.5rem; }

        .section-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 1.1rem;
        }

        .section-count {
          font-size: 0.82rem;
          color: var(--text-secondary);
          font-weight: 600;
        }

        .courier-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 1rem;
        }

        /* Cards */
        .courier-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius);
          padding: 1.1rem;
          display: flex;
          flex-direction: column;
          gap: 0.75rem;
          transition: var(--transition);
          backdrop-filter: blur(12px);
        }

        .ready-card {
          border-top: 3px solid rgba(16,185,129,0.5);
        }

        .ready-card:hover {
          border-color: rgba(16,185,129,0.4);
          box-shadow: 0 8px 24px rgba(16,185,129,0.12);
        }

        .active-card {
          border-top: 3px solid var(--primary);
        }

        .active-card:hover {
          border-color: rgba(249,115,22,0.4);
          box-shadow: 0 8px 24px rgba(249,115,22,0.12);
        }

        .ccard-top {
          display: flex;
          align-items: center;
          justify-content: space-between;
        }

        .ccard-id {
          font-size: 1.15rem;
          font-weight: 800;
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        .ccard-price {
          font-weight: 700;
          color: var(--success);
          font-size: 0.9rem;
        }

        .on-way-badge {
          background: rgba(249,115,22,0.15);
          color: var(--primary);
          border: 1px solid rgba(249,115,22,0.30);
          padding: 0.2rem 0.7rem;
          border-radius: 99px;
          font-size: 0.78rem;
          font-weight: 700;
          letter-spacing: 0.03em;
          animation: pulse-orange 2s ease-in-out infinite;
        }

        @keyframes pulse-orange {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.6; }
        }

        .ccard-addr {
          display: flex;
          gap: 0.6rem;
          color: var(--text-secondary);
          font-size: 0.88rem;
          align-items: flex-start;
        }

        .ccard-addr svg { flex-shrink: 0; margin-top: 2px; color: var(--primary-light); }
        .ccard-details { display: flex; flex-direction: column; gap: 0.5rem; }

        .ccard-items {
          display: flex;
          flex-wrap: wrap;
          gap: 0.35rem;
        }

        .item-chip {
          background: rgba(255,255,255,0.06);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 0.2rem 0.6rem;
          font-size: 0.75rem;
          color: var(--text-secondary);
        }

        .ccard-btn {
          width: 100%;
          padding: 0.7rem;
          font-size: 0.88rem;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.4rem;
          border-radius: var(--radius-sm);
          font-weight: 700;
        }

        .ccard-delivered-btn {
          background: rgba(16,185,129,0.15);
          color: #34d399;
          border: 1px solid rgba(16,185,129,0.30);
        }

        .ccard-delivered-btn:hover {
          background: rgba(16,185,129,0.25);
          transform: translateY(-1px);
        }

        .empty-section {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 0.5rem;
          padding: 2.5rem 1rem;
          color: var(--text-muted);
          font-size: 0.88rem;
          font-weight: 600;
        }

        .empty-section span { font-size: 2rem; opacity: 0.5; }
      `}</style>
    </div>
  );
};

export default Courier;

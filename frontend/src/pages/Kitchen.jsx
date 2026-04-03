import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion, AnimatePresence } from 'framer-motion';
import { ChefHat, Clock, CheckCircle2, RefreshCw, Flame, Zap } from 'lucide-react';
import StatsSection from '../components/StatsSection';

const STATUS_CONFIG = {
  new:       { label: 'Yangi',          emoji: '🆕', color: '#818cf8', bg: 'rgba(99,102,241,0.15)',  border: 'rgba(99,102,241,0.30)'  },
  preparing: { label: 'Tayyorlanmoqda', emoji: '👨‍🍳', color: '#fbbf24', bg: 'rgba(251,191,36,0.12)', border: 'rgba(251,191,36,0.30)'  },
  ready:     { label: 'Tayyor',         emoji: '✅', color: '#34d399', bg: 'rgba(16,185,129,0.12)', border: 'rgba(16,185,129,0.30)'  },
};

const Kitchen = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const onWSMessage = (data) => {
    if (data.type === 'new_order') {
      setOrders(prev => [data.order, ...prev]);
    } else if (data.type === 'status_update') {
      fetchOrders();
    }
  };

  useWebsocket(onWSMessage);
  useEffect(() => { fetchOrders(); }, []);

  const fetchOrders = async () => {
    try {
      const res = await api.get('/orders/active');
      const kitchenOrders = (res.data || []).filter(o =>
        o.status === 'new' || o.status === 'preparing' || o.status === 'ready'
      );
      setOrders(kitchenOrders);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const updateStatus = async (orderId, newStatus) => {
    try {
      await api.put(`/orders/${orderId}/status`, { status: newStatus });
      fetchOrders();
    } catch (err) {
      alert('Statusni yangilashda xatolik');
    }
  };

  const handleRefresh = () => {
    setRefreshing(true);
    fetchOrders();
  };

  if (loading) return (
    <div className="loading-screen">
      <div className="spinner" />
      <p>Buyurtmalar yuklanmoqda...</p>
    </div>
  );

  return (
    <div className="kitchen-page animate-fade">
      {/* Header */}
      <div className="kitchen-header">
        <div className="kitchen-title-wrap">
          <div className="kitchen-icon">
            <ChefHat size={24} />
          </div>
          <div>
            <h1>Oshxona Paneli</h1>
            <p className="kitchen-subtitle">
              {orders.length} ta aktiv buyurtma
            </p>
          </div>
        </div>
        <button
          className={`refresh-btn ${refreshing ? 'refreshing' : ''}`}
          onClick={handleRefresh}
          title="Yangilash"
        >
          <RefreshCw size={16} />
        </button>
      </div>

      {/* Stats */}
      <StatsSection role="cook" />

      {/* Kanban Board */}
      <div className="kanban-board">
        {Object.entries(STATUS_CONFIG).map(([status, cfg]) => {
          const colOrders = orders.filter(o => o.status === status);
          return (
            <div key={status} className="kanban-col">
              {/* Column Header */}
              <div className="col-header" style={{ borderColor: cfg.border }}>
                <div className="col-header-left">
                  <span className="col-emoji">{cfg.emoji}</span>
                  <span className="col-label" style={{ color: cfg.color }}>{cfg.label}</span>
                </div>
                <span
                  className="col-count"
                  style={{ background: cfg.bg, color: cfg.color, border: `1px solid ${cfg.border}` }}
                >
                  {colOrders.length}
                </span>
              </div>

              {/* Orders */}
              <div className="col-body">
                <AnimatePresence mode="popLayout">
                  {colOrders.length === 0 ? (
                    <motion.div
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      className="col-empty"
                    >
                      <span style={{ fontSize: '2rem', opacity: 0.3 }}>{cfg.emoji}</span>
                      <p>Bo'sh</p>
                    </motion.div>
                  ) : (
                    colOrders.map(order => (
                      <motion.div
                        key={order.id}
                        layout
                        initial={{ opacity: 0, scale: 0.95, y: 10 }}
                        animate={{ opacity: 1, scale: 1, y: 0 }}
                        exit={{ opacity: 0, scale: 0.9 }}
                        transition={{ duration: 0.25 }}
                        className="order-card"
                        style={{ borderTopColor: cfg.color }}
                      >
                        {/* Order top */}
                        <div className="order-top">
                          <span className="order-num">#{order.id}</span>
                          <span className="order-time">
                            <Clock size={12} />
                            {new Date(order.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                          </span>
                        </div>

                        {/* Items */}
                        <div className="order-items">
                          {(order.items || []).map((item, idx) => (
                            <div key={idx} className="order-item-row">
                              <span className="item-qty">{item.quantity}×</span>
                              <span className="item-name">{item.product_name || 'Noma\'lum'}</span>
                            </div>
                          ))}
                        </div>

                        {order.comment && (
                          <div className="order-note">
                            💬 {order.comment}
                          </div>
                        )}

                        {/* Actions */}
                        <div className="order-actions">
                          {status === 'new' && (
                            <button
                              className="action-btn prepare-btn"
                              onClick={() => updateStatus(order.id, 'preparing')}
                            >
                              <Flame size={14} /> Pishirishni boshlash
                            </button>
                          )}
                          {status === 'preparing' && (
                            <button
                              className="action-btn ready-btn"
                              onClick={() => updateStatus(order.id, 'ready')}
                            >
                              <CheckCircle2 size={14} /> Tayyor!
                            </button>
                          )}
                          {status === 'ready' && (
                            <div className="waiting-courier">
                              <Zap size={14} />
                              Kuryerni kutmoqda
                            </div>
                          )}
                        </div>
                      </motion.div>
                    ))
                  )}
                </AnimatePresence>
              </div>
            </div>
          );
        })}
      </div>

      <style>{`
        .kitchen-page { padding-bottom: 3rem; }

        .kitchen-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          margin-bottom: 2rem;
        }

        .kitchen-title-wrap {
          display: flex;
          align-items: center;
          gap: 1rem;
        }

        .kitchen-icon {
          width: 52px; height: 52px;
          background: var(--grad-brand);
          border-radius: 14px;
          display: flex; align-items: center; justify-content: center;
          color: white;
          box-shadow: 0 6px 20px rgba(249,115,22,0.35);
          flex-shrink: 0;
        }

        .kitchen-header h1 { font-size: 1.75rem; }

        .kitchen-subtitle {
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

        /* KANBAN */
        .kanban-board {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 1.25rem;
          align-items: start;
        }

        .kanban-col {
          background: rgba(255,255,255,0.02);
          border: 1px solid var(--border);
          border-radius: var(--radius-lg);
          overflow: hidden;
          display: flex;
          flex-direction: column;
        }

        .col-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 0.9rem 1rem;
          border-bottom: 2px solid;
          background: rgba(255,255,255,0.02);
        }

        .col-header-left {
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }

        .col-emoji { font-size: 1.1rem; }

        .col-label {
          font-size: 0.85rem;
          font-weight: 700;
          text-transform: uppercase;
          letter-spacing: 0.04em;
        }

        .col-count {
          font-size: 0.8rem;
          font-weight: 800;
          padding: 0.2rem 0.6rem;
          border-radius: 99px;
          min-width: 24px;
          text-align: center;
        }

        .col-body {
          padding: 0.75rem;
          display: flex;
          flex-direction: column;
          gap: 0.65rem;
          min-height: 200px;
        }

        .col-empty {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 2.5rem 1rem;
          gap: 0.5rem;
          color: var(--text-muted);
          font-size: 0.85rem;
          font-weight: 600;
        }

        /* ORDER CARD */
        .order-card {
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-top: 3px solid;
          border-radius: var(--radius);
          padding: 0.9rem;
          display: flex;
          flex-direction: column;
          gap: 0.65rem;
          transition: var(--transition);
          backdrop-filter: blur(12px);
        }

        .order-card:hover {
          background: var(--bg-card-hover);
          box-shadow: var(--shadow-md);
        }

        .order-top {
          display: flex;
          align-items: center;
          justify-content: space-between;
        }

        .order-num {
          font-size: 1.1rem;
          font-weight: 800;
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
        }

        .order-time {
          display: flex;
          align-items: center;
          gap: 0.3rem;
          font-size: 0.78rem;
          color: var(--text-secondary);
        }

        .order-items { display: flex; flex-direction: column; gap: 0.35rem; }

        .order-item-row {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          font-size: 0.88rem;
        }

        .item-qty {
          font-weight: 800;
          color: var(--primary);
          font-size: 0.82rem;
          min-width: 20px;
        }

        .item-name { color: var(--text-primary); font-weight: 500; }

        .order-note {
          background: rgba(255,255,255,0.04);
          border: 1px solid var(--border);
          border-radius: 6px;
          padding: 0.5rem 0.65rem;
          font-size: 0.8rem;
          color: var(--text-secondary);
          font-style: italic;
        }

        .order-actions { margin-top: 0.25rem; }

        .action-btn {
          width: 100%;
          padding: 0.65rem;
          border-radius: var(--radius-sm);
          font-size: 0.85rem;
          display: flex; align-items: center; justify-content: center; gap: 0.4rem;
          font-weight: 700;
          transition: var(--transition);
        }

        .prepare-btn {
          background: rgba(99,102,241,0.15);
          color: #818cf8;
          border: 1px solid rgba(99,102,241,0.30);
        }

        .prepare-btn:hover {
          background: rgba(99,102,241,0.25);
          transform: translateY(-1px);
        }

        .ready-btn {
          background: rgba(16,185,129,0.15);
          color: #34d399;
          border: 1px solid rgba(16,185,129,0.30);
        }

        .ready-btn:hover {
          background: rgba(16,185,129,0.25);
          transform: translateY(-1px);
        }

        .waiting-courier {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.4rem;
          color: #fbbf24;
          font-size: 0.82rem;
          font-weight: 700;
          padding: 0.65rem;
          background: rgba(251,191,36,0.10);
          border: 1px solid rgba(251,191,36,0.25);
          border-radius: var(--radius-sm);
          animation: pulse-yellow 2s ease-in-out infinite;
        }

        @keyframes pulse-yellow {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.7; }
        }

        @media (max-width: 1100px) {
          .kanban-board { grid-template-columns: repeat(2, 1fr); }
        }

        @media (max-width: 720px) {
          .kanban-board { grid-template-columns: 1fr; }
          .kitchen-header { flex-wrap: wrap; gap: 0.75rem; }
          .kitchen-header h1 { font-size: 1.35rem; }
          .kitchen-icon { width: 42px; height: 42px; }
          .col-body { min-height: 100px; }
        }

        @media (max-width: 480px) {
          .kitchen-title-wrap { gap: 0.6rem; }
          .kitchen-icon { display: none; }
          .order-card { padding: 0.75rem; }
        }
      `}</style>
    </div>
  );
};

export default Kitchen;

import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion, AnimatePresence } from 'framer-motion';
import { ChefHat, Clock, CheckCircle2, Loader2, RefreshCw } from 'lucide-react';
import StatsSection from '../components/StatsSection';

const Kitchen = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);

  const onWSMessage = (data) => {
    if (data.type === 'new_order') {
      setOrders(prev => [data.order, ...prev]);
    } else if (data.type === 'status_update') {
      fetchOrders(); // Refresh all to be safe
    }
  };

  useWebsocket(onWSMessage);

  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const res = await api.get('/orders/active');
      // Only show orders that are NOT on way or delivered
      const kitchenOrders = res.data.filter(o => 
        o.status === 'new' || o.status === 'preparing' || o.status === 'ready'
      );
      setOrders(kitchenOrders);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
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

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" size={40} /></div>;

  return (
    <div className="kitchen-page animate-fade">
      <div className="page-header">
        <div className="flex items-center gap-4">
          <ChefHat size={32} className="text-primary" />
          <h1>Oshxona Paneli</h1>
        </div>
        <button className="refresh-btn" onClick={() => window.location.reload()} title="Sahifani yangilash"><RefreshCw size={20} /></button>
      </div>

      <StatsSection role="cook" />

      <div className="orders-board">
        {['new', 'preparing', 'ready'].map(status => (
          <div key={status} className="status-column">
            <div className="column-header">
              <h2 className="capitalize">{status === 'new' ? 'Yangi' : status === 'preparing' ? 'Tayyorlanmoqda' : 'Tayyor'}</h2>
              <span className="count">{orders.filter(o => o.status === status).length}</span>
            </div>

            <div className="column-content">
              <AnimatePresence mode='popLayout'>
                {orders.filter(o => o.status === status).map(order => (
                  <motion.div 
                    key={order.id}
                    layout
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.9 }}
                    className="premium-card order-card"
                  >
                    <div className="order-top">
                      <span className="order-id">#{order.id}</span>
                      <span className="order-time">
                        <Clock size={14} /> {new Date(order.created_at).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}
                      </span>
                    </div>

                    <div className="order-items">
                      {order.items?.map((item, idx) => (
                        <div key={idx} className="item">
                          <span className="qty">{item.quantity}x</span>
                          <span className="name">{item.product_name || 'Ma\'lumot yo\'q'}</span>
                        </div>
                      ))}
                    </div>

                    {order.comment && <div className="order-comment">"{order.comment}"</div>}

                    <div className="order-actions">
                      {status === 'new' && (
                        <button 
                          className="btn-status prepare-btn"
                          onClick={() => updateStatus(order.id, 'preparing')}
                        >
                          Pishirishni boshlash
                        </button>
                      )}
                      {status === 'preparing' && (
                        <button 
                          className="btn-status ready-btn"
                          onClick={() => updateStatus(order.id, 'ready')}
                        >
                          Tayyor
                        </button>
                      )}
                      {status === 'ready' && (
                        <div className="ready-indicator">
                          <CheckCircle2 size={16} /> Kuryerni kutmoqda
                        </div>
                      )}
                    </div>
                  </motion.div>
                ))}
              </AnimatePresence>
            </div>
          </div>
        ))}
      </div>

      <style>{`
        .kitchen-page {
          padding-top: 1rem;
        }

        .page-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
        }

        .orders-board {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 1.5rem;
          min-height: calc(100vh - 250px);
        }

        .status-column {
          background: rgba(15, 23, 42, 0.3);
          border-radius: 16px;
          padding: 1rem;
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }

        .column-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 0.5rem;
          margin-bottom: 1rem;
          border-bottom: 2px solid var(--border);
        }

        .column-header h2 {
          font-size: 1.1rem;
          color: var(--text-dim);
        }

        .count {
          background: var(--bg-surface);
          padding: 2px 10px;
          border-radius: 12px;
          font-weight: 700;
          font-size: 0.9rem;
        }

        .column-content {
          display: flex;
          flex-direction: column;
          gap: 1rem;
          flex: 1;
          overflow-y: auto;
        }

        .order-card {
          border-top: 4px solid var(--primary);
        }

        .order-top {
          display: flex;
          justify-content: space-between;
          margin-bottom: 1rem;
        }

        .order-id {
          font-weight: 800;
          font-size: 1.2rem;
        }

        .order-time {
          color: var(--text-dim);
          font-size: 0.85rem;
          display: flex;
          align-items: center;
          gap: 4px;
        }

        .order-items {
          margin-bottom: 1rem;
        }

        .item {
          display: flex;
          gap: 0.5rem;
          margin-bottom: 0.25rem;
          font-size: 1rem;
        }

        .qty { font-weight: 700; color: var(--primary); }

        .order-comment {
          font-style: italic;
          color: var(--text-dim);
          font-size: 0.9rem;
          padding: 0.5rem;
          background: var(--bg-dark);
          border-radius: 6px;
          margin-bottom: 1rem;
        }

        .btn-status {
          width: 100%;
          padding: 0.75rem;
          font-weight: 700;
        }

        .prepare-btn { background: #3b82f6; color: white; }
        .ready-btn { background: #10b981; color: white; }

        .ready-indicator {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.5rem;
          color: #10b981;
          font-weight: 600;
          font-size: 0.9rem;
          padding: 0.5rem;
        }

        .refresh-btn {
          background: var(--bg-surface);
          color: var(--text-dim);
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
        }

        @media (max-width: 1200px) {
          .orders-board { grid-template-columns: 1fr; }
        }
      `}</style>
    </div>
  );
};

export default Kitchen;

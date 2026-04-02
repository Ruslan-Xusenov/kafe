import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion, AnimatePresence } from 'framer-motion';
import { Truck, MapPin, Phone, CheckCircle, Navigation, Loader2, RefreshCw, BarChart3 } from 'lucide-react';
import StatsSection from '../components/StatsSection';

const Courier = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);

  const onWSMessage = (data) => {
    if (data.type === 'status_update' || data.type === 'new_order') {
      fetchOrders();
    }
  };

  useWebsocket(onWSMessage);

  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      const res = await api.get('/orders/active');
      // Show orders that are 'ready' (available to pick up) or 'on_way' (already picked up by this courier)
      setOrders(res.data.filter(o => o.status === 'ready' || o.status === 'on_way'));
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
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
    // Optimistic Update: remove from view immediately
    const originalOrders = [...orders];
    setOrders(prev => prev.filter(o => o.id !== orderId));

    try {
      await api.put(`/orders/${orderId}/status`, { status: 'delivered' });
      // Re-fetch to ensure sync, but optimistic update already handled the UI
      fetchOrders();
    } catch (err) {
      // Rollback on error
      setOrders(originalOrders);
      alert('Statusni yangilashda xatolik');
    }
  };

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" size={40} /></div>;

  return (
    <div className="courier-page animate-fade">
      <div className="page-header">
        <div className="flex items-center gap-4">
          <Truck size={32} className="text-primary" />
          <h1>Kuryer Paneli</h1>
        </div>
        <button className="refresh-btn" onClick={() => window.location.reload()} title="Sahifani yangilash"><RefreshCw size={20} /></button>
      </div>

      <StatsSection role="courier" />

      <div className="orders-list">
        <div className="section">
          <h2>Tayyor buyurtmalar (Olish mumkin)</h2>
          <div className="order-cards">
            <AnimatePresence>
              {orders.filter(o => o.status === 'ready').map(order => (
                <motion.div 
                  key={order.id}
                  layout
                  initial={{ opacity: 0, x: -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: 20 }}
                  className="premium-card courier-card ready"
                >
                  <div className="card-top">
                    <span className="order-id">#{order.id}</span>
                    <span className="price">{(order.total_price).toLocaleString()} so'm</span>
                  </div>
                  
                  <div className="card-address">
                    <MapPin size={18} />
                    <span>{order.address}</span>
                  </div>
                  
                  <button className="btn-primary action-btn" onClick={() => pickUp(order.id)}>
                    <Navigation size={18} /> Olib ketish
                  </button>
                </motion.div>
              ))}
              {orders.filter(o => o.status === 'ready').length === 0 && (
                <p className="empty-msg">Hozircha tayyor buyurtmalar yo'q.</p>
              )}
            </AnimatePresence>
          </div>
        </div>

        <div className="section mt-4">
          <h2>Sizning faol buyurtmalaringiz</h2>
          <div className="order-cards">
            <AnimatePresence>
              {orders.filter(o => o.status === 'on_way').map(order => (
                <motion.div 
                  key={order.id}
                  layout
                  initial={{ opacity: 0, scale: 0.9 }}
                  animate={{ opacity: 1, scale: 1 }}
                  className="premium-card courier-card active"
                >
                  <div className="card-top">
                    <span className="order-id">#{order.id}</span>
                    <span className="status-label">Yo'lda</span>
                  </div>
                  
                  <div className="card-body">
                    <div className="info-row">
                      <MapPin size={18} /> <span>{order.address}</span>
                    </div>
                    <div className="info-row">
                      <Phone size={18} /> <span>{order.phone}</span>
                    </div>
                  </div>
                  
                  <div className="order-items-mini">
                    {order.items?.map((item, idx) => (
                      <span key={idx}>{item.quantity}x {item.product_name}</span>
                    ))}
                  </div>

                  <button className="btn-delivered action-btn" onClick={() => deliver(order.id)}>
                    <CheckCircle size={18} /> Yetkazib berildi
                  </button>
                </motion.div>
              ))}
              {orders.filter(o => o.status === 'on_way').length === 0 && (
                <p className="empty-msg">Sizda hozircha faol buyurtmalar yo'q.</p>
              )}
            </AnimatePresence>
          </div>
        </div>
      </div>

      <style>{`
        .courier-page {
          padding-top: 1rem;
        }

        .page-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 3rem;
        }

        .section h2 {
          font-size: 1.25rem;
          margin-bottom: 1.5rem;
          color: var(--text-dim);
          border-left: 4px solid var(--primary);
          padding-left: 1rem;
        }

        .order-cards {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
          gap: 1.5rem;
          margin-bottom: 3rem;
        }

        .courier-card {
          display: flex;
          flex-direction: column;
          gap: 1.25rem;
        }

        .card-top {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .order-id {
          font-weight: 800;
          font-size: 1.25rem;
        }

        .price {
          font-weight: 700;
          color: var(--primary);
        }

        .card-address, .info-row {
          display: flex;
          gap: 1rem;
          color: var(--text-main);
          font-size: 1rem;
          line-height: 1.4;
        }

        .card-address span, .info-row span {
          flex: 1;
        }

        .action-btn {
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 0.75rem;
          padding: 1rem;
          font-weight: 700;
          width: 100%;
        }

        .btn-delivered {
          background: #10b981;
          color: white;
        }

        .btn-delivered:hover {
          background: #059669;
        }

        .status-label {
          background: #3b82f6;
          color: white;
          padding: 2px 10px;
          border-radius: 12px;
          font-size: 0.8rem;
          font-weight: 700;
        }

        .order-items-mini {
          font-size: 0.85rem;
          color: var(--text-dim);
          display: flex;
          flex-wrap: wrap;
          gap: 0.5rem;
        }

        .order-items-mini span {
          background: var(--bg-dark);
          padding: 2px 8px;
          border-radius: 4px;
        }

        .empty-msg {
          color: var(--text-dim);
          font-style: italic;
        }

        .mt-4 { margin-top: 4rem; }

        .refresh-btn {
          background: var(--bg-surface);
          color: var(--text-dim);
          width: 40px;
          height: 40px;
          display: flex;
          align-items: center;
          justify-content: center;
        }
      `}</style>
    </div>
  );
};

export default Courier;

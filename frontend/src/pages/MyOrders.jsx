import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Package, Clock, CheckCircle2, Star, History, Loader2, RefreshCcw } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import RatingModal from '../components/RatingModal';

const MyOrders = () => {
  const [orders, setOrders] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedOrder, setSelectedOrder] = useState(null);
  const [isRatingOpen, setIsRatingOpen] = useState(false);

  const fetchOrders = async () => {
    setLoading(true);
    try {
      const res = await api.get('/orders/my');
      setOrders(res.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrders();
  }, []);

  const getStatusIcon = (status) => {
    switch (status) {
      case 'new': return <Package className="text-blue-400" />;
      case 'preparing': return <Clock className="animate-pulse text-orange-400" />;
      case 'ready': return <CheckCircle2 className="text-green-400" />;
      case 'on_way': return <RefreshCcw className="animate-spin text-purple-400" />;
      case 'delivered': return <CheckCircle2 className="text-primary" />;
      default: return <Package />;
    }
  };

  if (loading) return (
    <div className="flex-center py-20">
      <Loader2 className="animate-spin text-primary" size={40} />
    </div>
  );

  return (
    <div className="my-orders-page container py-10">
      <div className="page-header mb-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold flex items-center gap-3">
            <History className="text-primary" /> Buyurtmalar Tarixi
          </h1>
          <p className="text-dim mt-1">Sizning barcha buyurtmalaringiz va ularning holati</p>
        </div>
        <button className="refresh-btn" onClick={fetchOrders}><RefreshCcw size={18} /> Yangilash</button>
      </div>

      <div className="orders-list">
        <AnimatePresence>
          {orders.length > 0 ? (
            orders.map((order, idx) => (
              <motion.div 
                key={order.id}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: idx * 0.1 }}
                className="premium-card order-card mb-4"
              >
                <div className="order-main">
                  <div className="order-left">
                    <div className={`status-icon-wrapper ${order.status}`}>
                      {getStatusIcon(order.status)}
                    </div>
                    <div className="order-info-text">
                      <h3>Buyurtma #{order.id}</h3>
                      <p className="text-dim">{new Date(order.created_at).toLocaleString()}</p>
                    </div>
                  </div>

                  <div className="order-middle">
                    <div className="items-summary">
                      {order.items?.map(i => (
                        <span key={i.id} className="item-tag">{i.quantity}x {i.product_name}</span>
                      ))}
                    </div>
                  </div>

                  <div className="order-right text-right">
                    <div className="order-price">{order.total_price.toLocaleString()} so'm</div>
                    <div className={`status-text ${order.status}`}>{order.status}</div>
                  </div>
                </div>

                {order.status === 'delivered' && (
                  <div className="order-actions">
                    <button 
                      className="rate-btn"
                      onClick={() => {
                        setSelectedOrder(order);
                        setIsRatingOpen(true);
                      }}
                    >
                      <Star size={16} /> Baholash va Fikr bildirish
                    </button>
                  </div>
                )}
              </motion.div>
            ))
          ) : (
            <div className="empty-orders text-center py-20 premium-card">
              <Package size={48} className="mx-auto text-dim mb-4" />
              <h3>Sizda hali buyurtmalar mavjud emas</h3>
              <p className="text-dim mt-2">Hoziroq birinchi buyurtmangizni berishni xohlaysizmi?</p>
              <button className="btn-primary mt-6" onClick={() => window.location.href='/'}>Menyuga o'tish</button>
            </div>
          )}
        </AnimatePresence>
      </div>

      {selectedOrder && (
        <RatingModal 
          isOpen={isRatingOpen}
          onClose={() => setIsRatingOpen(false)}
          order={selectedOrder}
          onSuccess={() => {
            fetchOrders();
            alert('Baho va fikringiz uchun rahmat!');
          }}
        />
      )}

      <style jsx>{`
        .order-card {
          padding: 1.5rem;
          border-radius: 20px;
          transition: 0.3s;
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }
        .order-main {
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 1.5rem;
        }
        .order-left { display: flex; align-items: center; gap: 1.5rem; }
        .status-icon-wrapper {
          width: 50px; height: 50px;
          border-radius: 14px;
          display: flex; align-items: center; justify-content: center;
          background: rgba(255,255,255,0.05);
        }
        .status-icon-wrapper.delivered { background: rgba(255, 107, 107, 0.1); }
        .order-info-text h3 { font-size: 1.1rem; font-weight: 700; }
        
        .order-middle { flex: 1; }
        .items-summary { display: flex; flex-wrap: wrap; gap: 0.5rem; }
        .item-tag {
          padding: 4px 10px;
          background: rgba(255,255,255,0.05);
          border-radius: 8px;
          font-size: 0.8rem;
          color: var(--text-dim);
          border: 1px solid var(--border);
        }

        .order-price { font-size: 1.25rem; font-weight: 700; color: var(--primary); }
        .status-text { text-transform: uppercase; font-size: 0.75rem; font-weight: 800; letter-spacing: 1px; margin-top: 4px; }
        .status-text.delivered { color: var(--primary); }
        .status-text.on_way { color: var(--secondary); }

        .order-actions {
          border-top: 1px solid var(--border);
          padding-top: 1rem;
          display: flex;
          justify-content: flex-end;
        }
        .rate-btn {
          display: flex; align-items: center; gap: 0.5rem;
          background: rgba(255, 255, 255, 0.05);
          color: var(--text-dim);
          padding: 8px 16px;
          border-radius: 12px;
          font-size: 0.9rem;
          transition: 0.2s;
        }
        .rate-btn:hover { background: var(--primary); color: white; transform: translateY(-2px); }

        .refresh-btn {
          display: flex; align-items: center; gap: 0.5rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          padding: 8px 16px;
          border-radius: 14px;
          color: var(--text-dim);
          transition: 0.2s;
        }
        .refresh-btn:hover { border-color: var(--primary); color: #fff; }

        @media (max-width: 768px) {
          .order-main { flex-direction: column; align-items: flex-start; gap: 1rem; }
          .order-right { text-align: left; }
          .order-actions { justify-content: flex-start; }
        }
      `}</style>
    </div>
  );
};

export default MyOrders;

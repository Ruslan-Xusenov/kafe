import React, { useState } from 'react';
import { Star, X, Send, Loader2 } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import api from '../store/authStore';

const RatingModal = ({ isOpen, onClose, order, onSuccess }) => {
  const [cookRating, setCookRating] = useState(0);
  const [courierRating, setCourierRating] = useState(0);
  const [cookComment, setCookComment] = useState('');
  const [courierComment, setCourierComment] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async () => {
    // Validate only for assigned staff
    if (order.cook_id && cookRating === 0) {
      alert('Iltimos, oshpazga baho bering');
      return;
    }
    if (order.courier_id && courierRating === 0) {
      alert('Iltimos, kuryerga baho bering');
      return;
    }

    setLoading(true);
    try {
      const ratings = [];
      if (order.cook_id) {
        ratings.push({
          staff_id: order.cook_id,
          staff_role: 'cook',
          rating: cookRating,
          comment: cookComment
        });
      }
      if (order.courier_id) {
        ratings.push({
          staff_id: order.courier_id,
          staff_role: 'courier',
          rating: courierRating,
          comment: courierComment
        });
      }

      await api.post(`/orders/${order.id}/rate`, ratings);
      onSuccess();
      onClose();
    } catch (err) {
      console.error(err);
      alert('Xatolik yuz berdi. Iltimos qaytadan urunib ko\'ring.');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <AnimatePresence>
      <div className="modal-overlay">
        <motion.div 
          initial={{ opacity: 0, scale: 0.9, y: 20 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.9, y: 20 }}
          className="premium-card feedback-modal"
        >
          <div className="modal-header">
            <h3>Xizmat ko'rsatish sifatini baholang</h3>
            <button className="close-btn" onClick={onClose}><X size={20} /></button>
          </div>

          <div className="order-summary-mini">
            <span>Buyurtma #{order.id}</span>
            <span>{new Date(order.created_at).toLocaleDateString()}</span>
          </div>

          <div className="rating-sections">
            {/* Cook Rating */}
            {order.cook_id && (
              <div className="rating-group">
                <div className="group-header">
                  <span className="role-tag">Oshpaz</span>
                  <h4>Taom sifati va tayyorlanishi</h4>
                </div>
                <div className="stars">
                  {[1, 2, 3, 4, 5].map(star => (
                    <Star 
                      key={star}
                      size={32}
                      className={star <= cookRating ? 'star filled' : 'star'}
                      onClick={() => setCookRating(star)}
                      fill={star <= cookRating ? 'var(--primary)' : 'none'}
                    />
                  ))}
                </div>
                <textarea 
                  placeholder="Oshpaz haqida izohingiz..."
                  value={cookComment}
                  onChange={(e) => setCookComment(e.target.value)}
                />
              </div>
            )}

            {order.cook_id && order.courier_id && <div className="divider" />}

            {/* Courier Rating */}
            {order.courier_id && (
              <div className="rating-group">
                <div className="group-header">
                  <span className="role-tag courier">Kuryer</span>
                  <h4>Yetkazib berish xizmati</h4>
                </div>
                <div className="stars">
                  {[1, 2, 3, 4, 5].map(star => (
                    <Star 
                      key={star}
                      size={32}
                      className={star <= courierRating ? 'star filled' : 'star'}
                      onClick={() => setCourierRating(star)}
                      fill={star <= courierRating ? 'var(--secondary)' : 'none'}
                    />
                  ))}
                </div>
                <textarea 
                  placeholder="Kuryer haqida izohingiz..."
                  value={courierComment}
                  onChange={(e) => setCourierComment(e.target.value)}
                />
              </div>
            )}
          </div>

          <button 
            className="submit-rating-btn" 
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? <Loader2 className="animate-spin" /> : <><Send size={18} /> Bahoni yuborish</>}
          </button>
        </motion.div>

        <style jsx>{`
          .modal-overlay {
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            background: rgba(0,0,0,0.8);
            backdrop-filter: blur(8px);
            z-index: 1000;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 1rem;
          }
          .feedback-modal {
            width: 100%;
            max-width: 500px;
            background: var(--bg-surface);
            border: 1px solid var(--border);
            border-radius: 24px;
            padding: 2rem;
            position: relative;
          }
          .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
          }
          .modal-header h3 { font-size: 1.5rem; font-weight: 700; color: #fff; }
          .close-btn { color: var(--text-dim); transition: 0.2s; }
          .close-btn:hover { color: #fff; transform: rotate(90deg); }
          
          .order-summary-mini {
            display: flex;
            justify-content: space-between;
            font-size: 0.9rem;
            color: var(--text-dim);
            margin-bottom: 1.5rem;
          }

          .rating-sections {
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
          }
          .rating-group {
            display: flex;
            flex-direction: column;
            gap: 1rem;
          }
          .group-header {
            display: flex;
            align-items: center;
            gap: 1rem;
          }
          .role-tag {
            font-size: 0.7rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            padding: 4px 10px;
            background: rgba(255, 107, 107, 0.1);
            color: var(--primary);
            border-radius: 8px;
            font-weight: 700;
          }
          .role-tag.courier {
            background: rgba(77, 171, 245, 0.1);
            color: var(--secondary);
          }
          .stars {
            display: flex;
            justify-content: center;
            gap: 0.5rem;
          }
          .star {
            cursor: pointer;
            color: var(--border);
            transition: 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
          }
          .star:hover { transform: scale(1.2); color: var(--primary); }
          .star.filled { color: var(--primary); }
          .rating-group:nth-child(3) .star.filled { color: var(--secondary); }
          
          textarea {
            background: rgba(255,255,255,0.03);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 1rem;
            color: #fff;
            font-size: 0.9rem;
            min-height: 80px;
            resize: none;
            transition: 0.2s;
          }
          textarea:focus { border-color: var(--primary); outline: none; background: rgba(255,255,255,0.06); }
          
          .divider { height: 1px; background: var(--border); opacity: 0.3; }

          .submit-rating-btn {
            width: 100%;
            margin-top: 2rem;
            background: linear-gradient(135deg, var(--primary), #ff8787);
            color: #fff;
            padding: 1rem;
            border-radius: 14px;
            font-weight: 700;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.75rem;
            transition: 0.3s;
            box-shadow: 0 10px 20px rgba(255, 107, 107, 0.2);
          }
          .submit-rating-btn:hover { transform: translateY(-3px); box-shadow: 0 15px 30px rgba(255, 107, 107, 0.3); }
          .submit-rating-btn:disabled { opacity: 0.5; transform: none; cursor: not-allowed; }
        `}</style>
      </div>
    </AnimatePresence>
  );
};

export default RatingModal;

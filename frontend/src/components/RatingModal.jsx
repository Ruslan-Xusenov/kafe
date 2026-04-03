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
      <div className="rm-overlay" onClick={(e) => e.target === e.currentTarget && onClose()}>
        <motion.div 
          initial={{ opacity: 0, y: 60 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: 60 }}
          transition={{ type: 'spring', damping: 28, stiffness: 300 }}
          className="feedback-modal"
        >
          {/* Drag handle (mobile) */}
          <div className="rm-handle" />
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

        <style>{`
          /* Overlay */
          .rm-overlay {
            position: fixed;
            inset: 0;
            background: rgba(0,0,0,0.75);
            backdrop-filter: blur(8px);
            -webkit-backdrop-filter: blur(8px);
            z-index: 2000;
            display: flex;
            align-items: flex-end;
            justify-content: center;
          }

          /* Modal card */
          .feedback-modal {
            width: 100%;
            max-width: 520px;
            background: var(--bg-surface);
            border: 1px solid var(--border);
            border-radius: 24px 24px 0 0;
            padding: 0.5rem 2rem 2rem;
            position: relative;
            max-height: 92vh;
            overflow-y: auto;
          }

          /* Drag handle */
          .rm-handle {
            width: 40px; height: 4px;
            background: rgba(255,255,255,0.15);
            border-radius: 99px;
            margin: 0.75rem auto 1.25rem;
          }

          /* Desktop: center it */
          @media (min-width: 640px) {
            .rm-overlay { align-items: center; }
            .feedback-modal {
              border-radius: 24px;
              padding: 2rem;
              width: 90%;
              max-width: 500px;
            }
            .rm-handle { display: none; }
          }

          .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
          }
          .modal-header h3 { font-size: 1.2rem; font-weight: 700; color: #fff; }
          .close-btn { color: var(--text-muted); transition: 0.2s; }
          .close-btn:hover { color: #fff; transform: rotate(90deg); }
          
          .order-summary-mini {
            display: flex;
            justify-content: space-between;
            font-size: 0.85rem;
            color: var(--text-muted);
            margin-bottom: 1.25rem;
          }

          .rating-sections { display: flex; flex-direction: column; gap: 1.25rem; }
          .rating-group { display: flex; flex-direction: column; gap: 0.75rem; }
          .group-header { display: flex; align-items: center; gap: 0.75rem; }
          .role-tag {
            font-size: 0.68rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            padding: 3px 8px;
            background: rgba(249,115,22,0.12);
            color: var(--primary);
            border-radius: 6px;
            font-weight: 700;
          }
          .role-tag.courier {
            background: rgba(99,102,241,0.12);
            color: #818cf8;
          }
          .stars {
            display: flex;
            justify-content: center;
            gap: 0.4rem;
          }
          .star {
            cursor: pointer;
            color: var(--border);
            transition: 0.2s cubic-bezier(0.175, 0.885, 0.32, 1.275);
          }
          .star:hover { transform: scale(1.2); color: var(--primary); }
          .star.filled { color: #fbbf24; }
          
          textarea {
            background: rgba(255,255,255,0.03);
            border: 1px solid var(--border);
            border-radius: 10px;
            padding: 0.85rem;
            color: #fff;
            font-size: 0.88rem;
            min-height: 70px;
            resize: none;
            transition: 0.2s;
          }
          textarea:focus { border-color: var(--primary); outline: none; }
          
          .divider { height: 1px; background: var(--border); opacity: 0.3; }

          .submit-rating-btn {
            width: 100%;
            margin-top: 1.5rem;
            background: var(--grad-brand);
            color: #fff;
            padding: 0.9rem;
            border-radius: 14px;
            font-weight: 700;
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.75rem;
            transition: 0.3s;
            box-shadow: 0 8px 24px rgba(249,115,22,0.3);
          }
          .submit-rating-btn:hover { transform: translateY(-2px); box-shadow: 0 12px 32px rgba(249,115,22,0.45); }
          .submit-rating-btn:disabled { opacity: 0.5; transform: none; cursor: not-allowed; }
        `}</style>
      </div>
    </AnimatePresence>
  );
};

export default RatingModal;

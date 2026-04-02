import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useCartStore } from '../store/cartStore';
import { useAuthStore } from '../store/authStore';
import api from '../store/authStore';
import { validatePhone, validateAddress } from '../utils/validation';
import { motion } from 'framer-motion';
import { MapPin, Phone, CreditCard, Send, Loader2 } from 'lucide-react';

const Checkout = () => {
  const { items, getTotal, clearCart } = useCartStore();
  const { user } = useAuthStore();
  const navigate = useNavigate();
  
  const [loading, setLoading] = useState(false);
  const [address, setAddress] = useState('');
  const [phone, setPhone] = useState(user?.phone || '');
  const [comment, setComment] = useState('');
  const [errors, setErrors] = useState({});

  const handleValidate = () => {
    const newErrors = {
      address: validateAddress(address),
      phone: validatePhone(phone)
    };
    setErrors(newErrors);
    return !newErrors.address && !newErrors.phone;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!handleValidate()) return;

    setLoading(true);

    try {
      const orderData = {
        items: items.map(i => ({
          product_id: i.id,
          quantity: i.quantity,
          comment: ''
        })),
        address,
        phone,
        comment,
        total_price: getTotal() + 15000
      };

      const res = await api.post('/orders/', orderData);
      if (res.status === 201) {
        clearCart();
        navigate('/success', { state: { orderId: res.data.id } });
      }
    } catch (err) {
      alert(err.response?.data?.error || 'Buyurtma berishda xatolik yuz berdi');
    } finally {
      setLoading(false);
    }
  };

  if (items.length === 0) {
    navigate('/cart');
    return null;
  }

  return (
    <div className="checkout-page animate-fade">
      <h1>Rasmiylashtirish</h1>
      
      <div className="checkout-content">
        <form onSubmit={handleSubmit} className="checkout-form">
          <div className="premium-card">
            <div className="section-title">
              <MapPin size={24} /> <h2>Yetkazib berish</h2>
            </div>
            
            <div className={`input-group ${errors.address ? 'has-error' : ''}`}>
              <label>Manzil</label>
              <textarea 
                placeholder="Shahar, tuman, ko'cha, uy raqami..." 
                value={address}
                onChange={(e) => {
                  setAddress(e.target.value);
                  if (errors.address) setErrors({...errors, address: null});
                }}
              />
              {errors.address && <span className="field-error">{errors.address}</span>}
            </div>

            <div className={`input-group ${errors.phone ? 'has-error' : ''}`}>
              <label><Phone size={16} /> Bog'lanish uchun telefon</label>
              <input 
                type="text" 
                value={phone}
                onChange={(e) => {
                  setPhone(e.target.value);
                  if (errors.phone) setErrors({...errors, phone: null});
                }}
              />
              {errors.phone && <span className="field-error">{errors.phone}</span>}
            </div>

            <div className="input-group">
              <label>Izoh (ixtiyoriy)</label>
              <textarea 
                placeholder="Podyezd kodi, mo'ljal..." 
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              />
            </div>
          </div>

          <div className="premium-card mt-2">
            <div className="section-title">
              <CreditCard size={24} /> <h2>To'lov turi</h2>
            </div>
            <div className="payment-options">
              <label className="radio-group active">
                <input type="radio" name="payment" defaultChecked />
                <span>Naqd pul / Terminal</span>
              </label>
              <p className="payment-hint">To'lov taom yetkazilganda amalga oshiriladi</p>
            </div>
          </div>
        </form>

        <div className="checkout-summary premium-card glass">
          <h2>Xulosa</h2>
          <div className="summary-list">
            {items.map(item => (
              <div key={item.id} className="summary-item">
                <span>{item.name} x {item.quantity}</span>
                <span>{(item.price * item.quantity).toLocaleString()} so'm</span>
              </div>
            ))}
          </div>
          
          <div className="total-breakdown">
            <div className="row">
              <span>Mahsulotlar</span>
              <span>{getTotal().toLocaleString()} so'm</span>
            </div>
            <div className="row">
              <span>Yetkazib berish</span>
              <span>15,000 so'm</span>
            </div>
            <div className="final-total">
              <span>Jami</span>
              <span>{(getTotal() + 15000).toLocaleString()} so'm</span>
            </div>
          </div>

          <button 
            type="submit" 
            className="btn-primary confirm-btn" 
            onClick={handleSubmit}
            disabled={loading}
          >
            {loading ? <Loader2 className="animate-spin" /> : <><Send size={20} /> Tasdiqlash</>}
          </button>
        </div>
      </div>

      <style>{`
        .checkout-page h1 {
          margin-bottom: 2rem;
        }

        .checkout-content {
          display: grid;
          grid-template-columns: 1fr 400px;
          gap: 2rem;
          align-items: start;
        }

        .section-title {
          display: flex;
          align-items: center;
          gap: 1rem;
          margin-bottom: 2rem;
          color: var(--primary);
        }

        .section-title h2 {
          color: white;
          font-size: 1.25rem;
        }

        .mt-2 { margin-top: 2rem; }

        textarea {
          background: rgba(15, 23, 42, 0.5);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 1rem;
          color: white;
          min-height: 100px;
          resize: vertical;
          outline: none;
          transition: var(--transition);
        }

        textarea:focus { border-color: var(--primary); }

        .field-error {
          color: #ef4444;
          font-size: 0.75rem;
          margin-top: 0.25rem;
          display: block;
        }

        .input-group.has-error input, .input-group.has-error textarea {
          border-color: #ef4444;
        }

        .payment-options {
          padding: 1rem;
          background: var(--bg-dark);
          border-radius: 12px;
          border: 1px solid var(--primary);
        }

        .radio-group {
          display: flex;
          align-items: center;
          gap: 1rem;
          font-weight: 600;
        }

        .payment-hint {
          font-size: 0.85rem;
          color: var(--text-dim);
          margin-top: 0.5rem;
          margin-left: 1.75rem;
        }

        .checkout-summary {
          position: sticky;
          top: 90px;
          padding: 2rem;
        }

        .summary-list {
          margin-bottom: 1.5rem;
          max-height: 200px;
          overflow-y: auto;
        }

        .summary-item {
          display: flex;
          justify-content: space-between;
          font-size: 0.9rem;
          color: var(--text-dim);
          margin-bottom: 0.5rem;
        }

        .total-breakdown {
          border-top: 1px solid var(--border);
          padding-top: 1.5rem;
        }

        .row {
          display: flex;
          justify-content: space-between;
          margin-bottom: 0.75rem;
          color: var(--text-dim);
        }

        .final-total {
          display: flex;
          justify-content: space-between;
          font-size: 1.5rem;
          font-weight: 800;
          color: white;
          margin-top: 1rem;
          margin-bottom: 2rem;
        }

        .confirm-btn {
          width: 100%;
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 1rem;
          padding: 1rem;
          font-size: 1.1rem;
        }

        @media (max-width: 992px) {
          .checkout-content {
            grid-template-columns: 1fr;
          }
        }
      `}</style>
    </div>
  );
};

export default Checkout;

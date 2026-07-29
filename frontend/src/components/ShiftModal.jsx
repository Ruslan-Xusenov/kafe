import React, { useState } from 'react';
import { X, Wallet, ArrowRightLeft } from 'lucide-react';

const ShiftModal = ({ isOpen, onClose, type, currentShift, onConfirm }) => {
  const [amount, setAmount] = useState('');
  const [notes, setNotes] = useState('');
  const [operationType, setOperationType] = useState('cash_in'); // For cash operations

  if (!isOpen) return null;

  const handleSubmit = () => {
    const val = parseFloat(amount);
    if (isNaN(val) || val < 0) return alert("Iltimos, to'g'ri summa kiriting");

    if (type === 'open') {
      onConfirm({ opening_cash: val });
    } else if (type === 'close') {
      onConfirm({ closing_cash: val, notes });
    } else if (type === 'operation') {
      if (!notes) return alert("Sabab kiritilishi shart");
      onConfirm({ type: operationType, amount: val, reason: notes });
    }
  };

  return (
    <div className="payment-modal-overlay" onClick={onClose}>
      <div className="payment-modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '400px' }}>
        <div className="payment-modal-header">
          <h2>
            {type === 'open' ? 'Smena Ochish' : 
             type === 'close' ? 'Smena Yopish' : 'Kassa Operatsiyasi'}
          </h2>
          <button className="pm-close" onClick={onClose}><X size={22} /></button>
        </div>

        <div className="pm-rows" style={{ marginTop: '1rem' }}>
          {type === 'close' && currentShift && (
            <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '12px', marginBottom: '1rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
                <span style={{ color: '#64748b' }}>Kutilayotgan naqd pul:</span>
                <strong style={{ fontSize: '1.1rem' }}>{currentShift.expected_cash?.toLocaleString() || 0} so'm</strong>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span style={{ color: '#64748b' }}>Umumiy savdo:</span>
                <strong>{currentShift.total_sales?.toLocaleString() || 0} so'm</strong>
              </div>
            </div>
          )}

          {type === 'operation' && (
            <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
              <button 
                style={{ flex: 1, padding: '0.75rem', borderRadius: '8px', border: '1px solid',
                         background: operationType === 'cash_in' ? '#ecfdf5' : '#fff',
                         borderColor: operationType === 'cash_in' ? '#10b981' : '#e2e8f0',
                         color: operationType === 'cash_in' ? '#059669' : '#64748b',
                         fontWeight: 600, cursor: 'pointer' }}
                onClick={() => setOperationType('cash_in')}
              >
                Kirim (Cash In)
              </button>
              <button 
                style={{ flex: 1, padding: '0.75rem', borderRadius: '8px', border: '1px solid',
                         background: operationType === 'cash_out' ? '#fef2f2' : '#fff',
                         borderColor: operationType === 'cash_out' ? '#ef4444' : '#e2e8f0',
                         color: operationType === 'cash_out' ? '#dc2626' : '#64748b',
                         fontWeight: 600, cursor: 'pointer' }}
                onClick={() => setOperationType('cash_out')}
              >
                Chiqim (Cash Out)
              </button>
            </div>
          )}

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1rem' }}>
            <label style={{ fontSize: '0.9rem', color: '#475569', fontWeight: 600 }}>
              {type === 'open' ? "Kassadagi boshlang'ich pul" : 
               type === 'close' ? "Kassadagi oxirgi pul" : "Summa"}
            </label>
            <div style={{ position: 'relative' }}>
              <input
                type="number"
                value={amount}
                onChange={e => setAmount(e.target.value)}
                placeholder="0"
                style={{ width: '100%', padding: '0.9rem 1rem', paddingLeft: '2.5rem', 
                         borderRadius: '12px', border: '1.5px solid #e2e8f0', fontSize: '1.1rem', outline: 'none' }}
                autoFocus
              />
              <Wallet size={18} style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', color: '#94a3b8' }} />
            </div>
          </div>

          {(type === 'close' || type === 'operation') && (
            <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1.5rem' }}>
              <label style={{ fontSize: '0.9rem', color: '#475569', fontWeight: 600 }}>
                {type === 'close' ? "Izoh (ixtiyoriy)" : "Sabab *"}
              </label>
              <input
                type="text"
                value={notes}
                onChange={e => setNotes(e.target.value)}
                placeholder={type === 'close' ? "Kun bo'yicha izohlar..." : "Masalan: Xo'jalik xarajatlari"}
                style={{ width: '100%', padding: '0.9rem 1rem', borderRadius: '12px', border: '1.5px solid #e2e8f0', fontSize: '1rem', outline: 'none' }}
              />
            </div>
          )}
        </div>

        <button 
          onClick={handleSubmit}
          style={{ width: '100%', padding: '1rem', borderRadius: '12px', background: '#3b82f6', color: '#fff', 
                   border: 'none', fontWeight: 700, fontSize: '1.05rem', cursor: 'pointer', transition: 'all 0.2s' }}
        >
          {type === 'open' ? 'Smenani Boshlash' : 
           type === 'close' ? 'Smenani Yopish' : 'Tasdiqlash'}
        </button>
      </div>
    </div>
  );
};

export default ShiftModal;

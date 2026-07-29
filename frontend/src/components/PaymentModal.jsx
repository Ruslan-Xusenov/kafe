import React, { useState, useEffect } from 'react';
import { X, Plus, Trash2, CreditCard, Banknote, Smartphone, BookOpen } from 'lucide-react';

const PAYMENT_METHODS = [
  { key: 'cash', label: 'Naqd', icon: Banknote, color: '#22c55e' },
  { key: 'card', label: 'Karta', icon: CreditCard, color: '#3b82f6' },
  { key: 'click', label: 'Click/Payme', icon: Smartphone, color: '#8b5cf6' },
  { key: 'nasiya', label: 'Nasiya', icon: BookOpen, color: '#f59e0b' },
];

const PaymentModal = ({ isOpen, onClose, totalAmount, onConfirm, debtors = [], onCreateDebtor }) => {
  const [payments, setPayments] = useState([{ method: 'cash', amount: totalAmount || 0 }]);
  const [selectedDebtor, setSelectedDebtor] = useState(null);
  const [showDebtorForm, setShowDebtorForm] = useState(false);
  const [newDebtorName, setNewDebtorName] = useState('');
  const [newDebtorPhone, setNewDebtorPhone] = useState('');

  useEffect(() => {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setPayments([{ method: 'cash', amount: totalAmount || 0 }]);
      setSelectedDebtor(null);
      setShowDebtorForm(false);
  }, [isOpen, totalAmount]);

  if (!isOpen) return null;

  const paymentSum = payments.reduce((s, p) => s + (parseFloat(p.amount) || 0), 0);
  const remaining = totalAmount - paymentSum;
  const isValid = Math.abs(remaining) < 1;
  const hasNasiya = payments.some(p => p.method === 'nasiya');

  const addPaymentRow = () => {
    setPayments([...payments, { method: 'cash', amount: 0 }]);
  };

  const removePaymentRow = (idx) => {
    if (payments.length <= 1) return;
    setPayments(payments.filter((_, i) => i !== idx));
  };

  const updatePayment = (idx, field, value) => {
    const updated = [...payments];
    updated[idx] = { ...updated[idx], [field]: field === 'amount' ? parseFloat(value) || 0 : value };
    setPayments(updated);
  };

  const setFullAmount = (method) => {
    setPayments([{ method, amount: totalAmount }]);
  };

  const handleConfirm = () => {
    if (!isValid) return;
    onConfirm(payments, hasNasiya ? selectedDebtor : null);
  };

  const handleCreateDebtor = async () => {
    if (!newDebtorName.trim()) return;
    if (onCreateDebtor) {
      const debtor = await onCreateDebtor({ name: newDebtorName, phone: newDebtorPhone });
      if (debtor) {
        setSelectedDebtor(debtor);
        setShowDebtorForm(false);
        setNewDebtorName('');
        setNewDebtorPhone('');
      }
    }
  };

  return (
    <div className="payment-modal-overlay" onClick={onClose}>
      <div className="payment-modal" onClick={e => e.stopPropagation()}>
        <div className="payment-modal-header">
          <h2>💳 To'lov</h2>
          <button className="pm-close" onClick={onClose}><X size={22} /></button>
        </div>

        <div className="payment-modal-total">
          <span>Umumiy summa:</span>
          <strong>{totalAmount?.toLocaleString()} so'm</strong>
        </div>

        {/* Quick payment buttons */}
        <div className="pm-quick-buttons">
          {PAYMENT_METHODS.map(m => (
            <button
              key={m.key}
              className="pm-quick-btn"
              style={{ '--btn-color': m.color }}
              onClick={() => setFullAmount(m.key)}
            >
              <m.icon size={20} />
              <span>{m.label}</span>
            </button>
          ))}
        </div>

        <div className="pm-divider">
          <span>yoki aralash to'lov</span>
        </div>

        {/* Payment rows */}
        <div className="pm-rows">
          {payments.map((p, idx) => (
            <div key={idx} className="pm-row">
              <select
                value={p.method}
                onChange={e => updatePayment(idx, 'method', e.target.value)}
                className="pm-select"
              >
                {PAYMENT_METHODS.map(m => (
                  <option key={m.key} value={m.key}>{m.label}</option>
                ))}
              </select>
              <input
                type="number"
                value={p.amount || ''}
                onChange={e => updatePayment(idx, 'amount', e.target.value)}
                placeholder="Summa"
                className="pm-input"
              />
              <span className="pm-currency">so'm</span>
              {payments.length > 1 && (
                <button className="pm-remove" onClick={() => removePaymentRow(idx)}>
                  <Trash2 size={16} />
                </button>
              )}
            </div>
          ))}
          <button className="pm-add-row" onClick={addPaymentRow}>
            <Plus size={16} /> Qo'shish
          </button>
        </div>

        {/* Remaining balance indicator */}
        <div className={`pm-balance ${isValid ? 'valid' : remaining > 0 ? 'short' : 'over'}`}>
          {isValid ? (
            <span>✅ To'lov to'g'ri</span>
          ) : remaining > 0 ? (
            <span>⚠️ Yana {remaining.toLocaleString()} so'm kerak</span>
          ) : (
            <span>⚠️ {Math.abs(remaining).toLocaleString()} so'm ortiqcha</span>
          )}
        </div>

        {/* Nasiya section - debtor selection */}
        {hasNasiya && (
          <div className="pm-nasiya-section">
            <h4>📒 Qarzdor tanlang:</h4>
            {!showDebtorForm ? (
              <>
                <select
                  value={selectedDebtor?.id || ''}
                  onChange={e => {
                    const d = debtors.find(d => d.id === parseInt(e.target.value));
                    setSelectedDebtor(d || null);
                  }}
                  className="pm-select pm-debtor-select"
                >
                  <option value="">-- Qarzdor tanlang --</option>
                  {debtors.map(d => (
                    <option key={d.id} value={d.id}>
                      {d.name} {d.phone ? `(${d.phone})` : ''} — Qarz: {d.total_debt?.toLocaleString()} so'm
                    </option>
                  ))}
                </select>
                <button className="pm-new-debtor-btn" onClick={() => setShowDebtorForm(true)}>
                  + Yangi qarzdor
                </button>
              </>
            ) : (
              <div className="pm-debtor-form">
                <input
                  type="text"
                  placeholder="Ism *"
                  value={newDebtorName}
                  onChange={e => setNewDebtorName(e.target.value)}
                  className="pm-input"
                />
                <input
                  type="text"
                  placeholder="Telefon"
                  value={newDebtorPhone}
                  onChange={e => setNewDebtorPhone(e.target.value)}
                  className="pm-input"
                />
                <div style={{ display: 'flex', gap: '0.5rem' }}>
                  <button className="pm-confirm-btn" style={{ flex: 1 }} onClick={handleCreateDebtor}>Saqlash</button>
                  <button className="pm-cancel-btn" onClick={() => setShowDebtorForm(false)}>Bekor</button>
                </div>
              </div>
            )}
          </div>
        )}

        <button
          className={`pm-confirm-btn ${!isValid || (hasNasiya && !selectedDebtor) ? 'disabled' : ''}`}
          onClick={handleConfirm}
          disabled={!isValid || (hasNasiya && !selectedDebtor)}
        >
          ✅ To'lovni tasdiqlash
        </button>

        <style>{`
          .payment-modal-overlay {
            position: fixed; inset: 0; z-index: 9999;
            background: rgba(0,0,0,0.6); backdrop-filter: blur(4px);
            display: flex; align-items: center; justify-content: center;
            padding: 1rem;
          }
          .payment-modal {
            background: #fff; border-radius: 20px; width: 100%; max-width: 480px;
            max-height: 90vh; overflow-y: auto; box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            padding: 1.5rem;
          }
          .payment-modal-header {
            display: flex; justify-content: space-between; align-items: center;
            margin-bottom: 1rem;
          }
          .payment-modal-header h2 { font-size: 1.3rem; font-weight: 800; margin: 0; }
          .pm-close {
            background: #f1f5f9; border: none; border-radius: 10px;
            width: 36px; height: 36px; display: flex; align-items: center;
            justify-content: center; cursor: pointer;
          }
          .payment-modal-total {
            background: linear-gradient(135deg, #f0fdf4, #dcfce7);
            border: 1px solid #bbf7d0; border-radius: 14px;
            padding: 1rem 1.25rem; display: flex; justify-content: space-between;
            align-items: center; margin-bottom: 1.25rem;
          }
          .payment-modal-total span { color: #166534; font-size: 0.95rem; }
          .payment-modal-total strong { font-size: 1.25rem; color: #15803d; }
          .pm-quick-buttons {
            display: grid; grid-template-columns: repeat(4, 1fr);
            gap: 0.5rem; margin-bottom: 1rem;
          }
          .pm-quick-btn {
            display: flex; flex-direction: column; align-items: center; gap: 4px;
            padding: 0.75rem 0.5rem; border-radius: 12px; border: 2px solid #e2e8f0;
            background: #fff; cursor: pointer; transition: all 0.2s;
            font-size: 0.75rem; font-weight: 600; color: #334155;
          }
          .pm-quick-btn:hover {
            border-color: var(--btn-color); color: var(--btn-color);
            background: color-mix(in srgb, var(--btn-color) 8%, white);
            transform: translateY(-1px);
          }
          .pm-divider {
            text-align: center; margin: 0.75rem 0; position: relative;
          }
          .pm-divider::before, .pm-divider::after {
            content: ''; position: absolute; top: 50%; height: 1px;
            background: #e2e8f0; width: calc(50% - 80px);
          }
          .pm-divider::before { left: 0; }
          .pm-divider::after { right: 0; }
          .pm-divider span { 
            background: #fff; padding: 0 0.75rem; font-size: 0.8rem; 
            color: #94a3b8; position: relative; 
          }
          .pm-rows { display: flex; flex-direction: column; gap: 0.5rem; margin-bottom: 1rem; }
          .pm-row {
            display: flex; gap: 0.5rem; align-items: center;
          }
          .pm-select {
            padding: 0.65rem 0.75rem; border-radius: 10px; border: 1.5px solid #e2e8f0;
            font-size: 0.9rem; background: #f8fafc; min-width: 130px; outline: none;
          }
          .pm-select:focus { border-color: #3b82f6; }
          .pm-input {
            padding: 0.65rem 0.75rem; border-radius: 10px; border: 1.5px solid #e2e8f0;
            font-size: 0.95rem; flex: 1; outline: none; min-width: 0;
          }
          .pm-input:focus { border-color: #3b82f6; }
          .pm-currency { color: #94a3b8; font-size: 0.8rem; white-space: nowrap; }
          .pm-remove {
            background: #fef2f2; border: 1px solid #fecaca; border-radius: 8px;
            color: #ef4444; cursor: pointer; padding: 6px; display: flex;
          }
          .pm-add-row {
            display: flex; align-items: center; justify-content: center; gap: 4px;
            padding: 0.5rem; border-radius: 10px; border: 2px dashed #cbd5e1;
            background: transparent; color: #64748b; cursor: pointer;
            font-size: 0.85rem; font-weight: 600; transition: all 0.2s;
          }
          .pm-add-row:hover { border-color: #3b82f6; color: #3b82f6; }
          .pm-balance {
            text-align: center; padding: 0.6rem; border-radius: 10px;
            font-weight: 600; font-size: 0.9rem; margin-bottom: 1rem;
          }
          .pm-balance.valid { background: #f0fdf4; color: #16a34a; border: 1px solid #bbf7d0; }
          .pm-balance.short { background: #fefce8; color: #ca8a04; border: 1px solid #fde68a; }
          .pm-balance.over { background: #fef2f2; color: #dc2626; border: 1px solid #fecaca; }
          .pm-nasiya-section {
            background: #fffbeb; border: 1px solid #fde68a; border-radius: 12px;
            padding: 1rem; margin-bottom: 1rem;
          }
          .pm-nasiya-section h4 { margin: 0 0 0.75rem 0; font-size: 0.95rem; }
          .pm-debtor-select { width: 100%; margin-bottom: 0.5rem; }
          .pm-new-debtor-btn {
            background: transparent; border: 1.5px dashed #f59e0b;
            border-radius: 8px; padding: 0.5rem; width: 100%;
            color: #92400e; cursor: pointer; font-weight: 600; font-size: 0.85rem;
          }
          .pm-debtor-form { display: flex; flex-direction: column; gap: 0.5rem; }
          .pm-confirm-btn {
            width: 100%; padding: 0.9rem; border-radius: 14px;
            background: linear-gradient(135deg, #22c55e, #16a34a);
            color: white; border: none; font-size: 1.05rem; font-weight: 700;
            cursor: pointer; transition: all 0.2s;
          }
          .pm-confirm-btn:hover:not(.disabled) { transform: translateY(-1px); box-shadow: 0 4px 12px rgba(34,197,94,0.4); }
          .pm-confirm-btn.disabled { opacity: 0.5; cursor: not-allowed; }
          .pm-cancel-btn {
            padding: 0.5rem 1rem; border-radius: 10px; background: #f1f5f9;
            border: 1px solid #e2e8f0; cursor: pointer; font-weight: 600;
          }
        `}</style>
      </div>
    </div>
  );
};

export default PaymentModal;

import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Loader2, Plus, ArrowRight, Save, History, ChevronDown, ChevronUp, User, Phone, CheckCircle, X } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';

const DebtsSection = () => {
  const [debtors, setDebtors] = useState([]);
  const [summary, setSummary] = useState({ total_debt: 0, debtor_count: 0 });
  const [loading, setLoading] = useState(true);
  
  const [showPayModal, setShowPayModal] = useState(false);
  const [selectedDebtor, setSelectedDebtor] = useState(null);
  const [payAmount, setPayAmount] = useState('');
  const [payMethod, setPayMethod] = useState('cash');
  const [payDesc, setPayDesc] = useState('');
  
  const [history, setHistory] = useState([]);
  const [historyLoading, setHistoryLoading] = useState(false);
  const [expandedDebtorId, setExpandedDebtorId] = useState(null);

  useEffect(() => {
    fetchDebts();
  }, []);

  const fetchDebts = async () => {
    setLoading(true);
    try {
      const [debtsRes, sumRes] = await Promise.all([
        api.get('/debts/debtors'),
        api.get('/debts/summary')
      ]);
      setDebtors(debtsRes.data || []);
      setSummary(sumRes.data || { total_debt: 0, debtor_count: 0 });
	} catch {
		alert("Qarzdorlarni yuklashda xatolik");
    } finally {
      setLoading(false);
    }
  };

  const fetchHistory = async (id) => {
    if (expandedDebtorId === id) {
      setExpandedDebtorId(null);
      return;
    }
    
    setExpandedDebtorId(id);
    setHistoryLoading(true);
    try {
      const res = await api.get(`/debts/debtors/${id}/history`);
      setHistory(res.data.history || []);
	} catch {
		alert("Tarixni yuklashda xatolik");
    } finally {
      setHistoryLoading(false);
    }
  };

  const openPayModal = (debtor) => {
    setSelectedDebtor(debtor);
    setPayAmount('');
    setPayDesc('');
    setShowPayModal(true);
  };

  const handlePay = async (e) => {
    e.preventDefault();
    if (!payAmount || isNaN(parseFloat(payAmount)) || parseFloat(payAmount) <= 0) {
      return alert("To'g'ri summa kiriting");
    }
    if (parseFloat(payAmount) > selectedDebtor.total_debt) {
      return alert("Summa qarzdan ko'p bo'lishi mumkin emas");
    }

    try {
      await api.post(`/debts/debtors/${selectedDebtor.id}/pay`, {
        amount: parseFloat(payAmount),
        method: payMethod,
        description: payDesc
      });
      setShowPayModal(false);
      fetchDebts();
      alert("To'lov muvaffaqiyatli qabul qilindi!");
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    }
  };

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" /></div>;

  return (
    <div className="animate-fade">
      <div className="section-header mb-6">
        <h2>Qarzdorlar (Nasiya)</h2>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', gap: '1rem', marginBottom: '1.5rem' }}>
        <div className="premium-card flex items-center justify-between" style={{ background: 'linear-gradient(135deg, #fef3c7, #fde68a)', border: '1px solid #fcd34d' }}>
          <div>
            <div style={{ color: '#92400e', fontSize: '0.9rem', fontWeight: 600 }}>Umumiy qarz miqdori</div>
            <div style={{ color: '#78350f', fontSize: '1.8rem', fontWeight: 800 }}>{summary.total_debt?.toLocaleString()} so'm</div>
          </div>
          <div style={{ background: '#f59e0b', color: 'white', padding: '1rem', borderRadius: '50%' }}>
            <History size={24} />
          </div>
        </div>
        <div className="premium-card flex items-center justify-between" style={{ background: '#f8fafc' }}>
          <div>
            <div style={{ color: 'var(--text-muted)', fontSize: '0.9rem', fontWeight: 600 }}>Aktiv qarzdorlar soni</div>
            <div style={{ color: 'var(--text-primary)', fontSize: '1.8rem', fontWeight: 800 }}>{summary.debtor_count} ta</div>
          </div>
          <div style={{ background: '#e2e8f0', color: '#64748b', padding: '1rem', borderRadius: '50%' }}>
            <User size={24} />
          </div>
        </div>
      </div>

      <div className="premium-card">
        <div className="table-responsive">
          <table className="admin-table w-full">
            <thead>
              <tr>
                <th>Mijoz Ismi</th>
                <th>Telefon</th>
                <th>Qolgan Qarz</th>
                <th>Oxirgi o'zgarish</th>
                <th>Deyarli</th>
              </tr>
            </thead>
            <tbody>
              {debtors.map(d => (
                <React.Fragment key={d.id}>
                  <tr style={{ background: expandedDebtorId === d.id ? '#f8fafc' : 'transparent' }}>
                    <td style={{ fontWeight: 600 }}>{d.name}</td>
                    <td>{d.phone || '—'}</td>
                    <td style={{ color: d.total_debt > 0 ? '#ef4444' : '#10b981', fontWeight: 700 }}>
                      {d.total_debt?.toLocaleString()} so'm
                    </td>
                    <td style={{ color: 'var(--text-muted)', fontSize: '0.9rem' }}>
                      {new Date(d.updated_at).toLocaleString('ru-RU')}
                    </td>
                    <td>
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        {d.total_debt > 0 && (
                          <button className="btn-primary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }} onClick={() => openPayModal(d)}>
                            Qarz uzish
                          </button>
                        )}
                        <button className="btn-secondary" style={{ padding: '0.4rem' }} onClick={() => fetchHistory(d.id)}>
                          {expandedDebtorId === d.id ? <ChevronUp size={18} /> : <ChevronDown size={18} />}
                        </button>
                      </div>
                    </td>
                  </tr>
                  <AnimatePresence>
                    {expandedDebtorId === d.id && (
                      <motion.tr
                        initial={{ opacity: 0, height: 0 }}
                        animate={{ opacity: 1, height: 'auto' }}
                        exit={{ opacity: 0, height: 0 }}
                      >
                        <td colSpan="5" style={{ padding: 0, border: 'none' }}>
                          <div style={{ background: '#f8fafc', padding: '1.5rem', borderBottom: '1px solid var(--border)' }}>
                            <h4 style={{ marginBottom: '1rem', color: '#475569' }}>Qarz Tarixi ({d.name})</h4>
                            {historyLoading ? (
                              <div className="flex-center"><Loader2 className="animate-spin text-primary" /></div>
                            ) : history.length === 0 ? (
                              <p className="text-muted">Tarix bo'sh</p>
                            ) : (
                              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
                                {history.map(h => (
                                  <div key={h.id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.75rem 1rem', background: '#fff', borderRadius: '8px', borderLeft: `4px solid ${h.type === 'debt' ? '#ef4444' : '#10b981'}` }}>
                                    <div>
                                      <div style={{ fontWeight: 600, color: h.type === 'debt' ? '#ef4444' : '#10b981' }}>
                                        {h.type === 'debt' ? 'Qarz olindi' : 'To\'lov qilindi'} 
                                        {h.order_id && <span style={{ color: '#94a3b8', fontSize: '0.8rem', marginLeft: '0.5rem' }}>(Chek #{h.order_id})</span>}
                                      </div>
                                      <div style={{ fontSize: '0.85rem', color: '#64748b' }}>
                                        {new Date(h.created_at).toLocaleString('ru-RU')} • Xodim: {h.created_by_name || 'Tizim'}
                                      </div>
                                      {h.description && <div style={{ fontSize: '0.85rem', color: '#475569', marginTop: '4px' }}>Izoh: {h.description}</div>}
                                    </div>
                                    <div style={{ textAlign: 'right' }}>
                                      <div style={{ fontWeight: 800, fontSize: '1.1rem', color: h.type === 'debt' ? '#ef4444' : '#10b981' }}>
                                        {h.type === 'debt' ? '+' : '-'}{h.amount?.toLocaleString()} so'm
                                      </div>
                                      {h.payment_method && <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{h.payment_method}</div>}
                                    </div>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        </td>
                      </motion.tr>
                    )}
                  </AnimatePresence>
                </React.Fragment>
              ))}
              {debtors.length === 0 && (
                <tr><td colSpan="5" className="text-center p-4">Qarzdorlar topilmadi</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showPayModal && selectedDebtor && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content" style={{ maxWidth: '400px' }}>
            <div className="modal-header">
              <h3>Qarz Uzish</h3>
              <button onClick={() => setShowPayModal(false)}><X size={20} /></button>
            </div>
            <form onSubmit={handlePay}>
              <div style={{ background: '#fef2f2', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem', border: '1px solid #fecaca' }}>
                <div style={{ fontSize: '0.9rem', color: '#991b1b' }}>Mijoz: {selectedDebtor.name}</div>
                <div style={{ fontSize: '1.2rem', fontWeight: 800, color: '#b91c1c' }}>Qolgan qarz: {selectedDebtor.total_debt?.toLocaleString()} so'm</div>
              </div>

              <div className="input-group">
                <label>To'lanayotgan summa</label>
                <input 
                  type="number" 
                  value={payAmount} 
                  onChange={e => setPayAmount(e.target.value)} 
                  placeholder="0" 
                  autoFocus 
                  required
                />
              </div>

              <div className="input-group">
                <label>To'lov usuli</label>
                <select value={payMethod} onChange={e => setPayMethod(e.target.value)}>
                  <option value="cash">Naqd</option>
                  <option value="card">Karta</option>
                  <option value="click">Click/Payme</option>
                </select>
              </div>

              <div className="input-group">
                <label>Izoh (ixtiyoriy)</label>
                <input 
                  type="text" 
                  value={payDesc} 
                  onChange={e => setPayDesc(e.target.value)} 
                  placeholder="Masalan: qisman to'lov" 
                />
              </div>

              <button type="submit" className="btn-primary w-full mt-4" style={{ padding: '0.9rem', fontSize: '1.05rem' }}>
                <CheckCircle size={18} /> Tasdiqlash
              </button>
            </form>
          </motion.div>
        </div>
      )}
    </div>
  );
};

export default DebtsSection;

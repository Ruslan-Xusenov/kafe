import React, { useState, useEffect, useCallback } from 'react';
import api from '../store/authStore';
import { Loader2, Check, X, RefreshCw, Filter, Banknote, AlertCircle } from 'lucide-react';
import { motion } from 'framer-motion';

const STATUS_MAP = {
  pending: { label: 'Kutilmoqda', color: '#f59e0b', bg: '#fef3c7' },
  approved: { label: 'Tasdiqlangan', color: '#10b981', bg: '#d1fae5' },
  rejected: { label: 'Rad etilgan', color: '#ef4444', bg: '#fee2e2' }
};

const RefundsSection = () => {
  const [refunds, setRefunds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('pending'); // pending, all
  
  const [showApproveModal, setShowApproveModal] = useState(false);
  const [selectedRefund, setSelectedRefund] = useState(null);
  const [refundMethod, setRefundMethod] = useState('cash');

	const fetchRefunds = useCallback(async () => {
		setLoading(true);
    try {
      const endpoint = filter === 'pending' ? '/refunds/pending' : '/refunds/all';
      const res = await api.get(endpoint);
      setRefunds(res.data || []);
	} catch {
		alert("Qaytarish so'rovlarini yuklashda xatolik");
		} finally {
			setLoading(false);
		}
	}, [filter]);

	useEffect(() => {
		fetchRefunds();
	}, [fetchRefunds]);

  const openApproveModal = (refund) => {
    setSelectedRefund(refund);
    setRefundMethod(refund.refund_method || 'cash');
    setShowApproveModal(true);
  };

  const handleApprove = async () => {
    try {
      await api.put(`/refunds/${selectedRefund.id}/approve`, { refund_method: refundMethod });
      setShowApproveModal(false);
      fetchRefunds();
      alert("Qaytarish so'rovi tasdiqlandi!");
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    }
  };

  const handleReject = async (id) => {
    if (!window.confirm("Rostdan ham rad etmoqchimisiz?")) return;
    try {
      await api.put(`/refunds/${id}/reject`);
      fetchRefunds();
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    }
  };

  const markMoneyReturned = async (id) => {
    try {
      await api.put(`/refunds/${id}/money-returned`);
      fetchRefunds();
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    }
  };

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" /></div>;

  return (
    <div className="animate-fade">
      <div className="section-header mb-6" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>Qaytarish (Refund) Operatsiyalari</h2>
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button 
            className={`pos-cat-btn ${filter === 'pending' ? 'active' : ''}`} 
            style={{ padding: '0.5rem 1rem' }}
            onClick={() => setFilter('pending')}
          >
            Kutilmoqda
          </button>
          <button 
            className={`pos-cat-btn ${filter === 'all' ? 'active' : ''}`} 
            style={{ padding: '0.5rem 1rem' }}
            onClick={() => setFilter('all')}
          >
            Barchasi
          </button>
        </div>
      </div>

      <div className="premium-card">
        <div className="table-responsive">
          <table className="admin-table w-full">
            <thead>
              <tr>
                <th>ID</th>
                <th>Chek №</th>
                <th>Summa</th>
                <th>Sabab</th>
                <th>So'radi</th>
                <th>Holat</th>
                <th>Amallar</th>
              </tr>
            </thead>
            <tbody>
              {refunds.map(r => (
                <tr key={r.id}>
                  <td>#{r.id}</td>
                  <td>
                    <span style={{ background: '#f1f5f9', padding: '4px 8px', borderRadius: '6px', fontWeight: 600 }}>
                      #{r.order_id}
                    </span>
                  </td>
                  <td style={{ fontWeight: 700, color: '#0f172a' }}>{r.amount?.toLocaleString()} so'm</td>
                  <td>
                    <div style={{ fontWeight: 600 }}>{r.reason}</div>
                    {r.reason_detail && <div style={{ fontSize: '0.8rem', color: '#64748b' }}>{r.reason_detail}</div>}
                  </td>
                  <td>
                    <div>{r.requested_by_name}</div>
                    <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>{new Date(r.created_at).toLocaleString('ru-RU')}</div>
                  </td>
                  <td>
                    <span style={{ 
                      background: STATUS_MAP[r.status]?.bg, color: STATUS_MAP[r.status]?.color,
                      padding: '4px 8px', borderRadius: '20px', fontSize: '0.85rem', fontWeight: 600
                    }}>
                      {STATUS_MAP[r.status]?.label}
                    </span>
                    {r.status === 'approved' && !r.money_returned && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.75rem', color: '#ef4444', marginTop: '4px' }}>
                        <AlertCircle size={12} /> Pul qaytarilmagan
                      </div>
                    )}
                    {r.status === 'approved' && r.money_returned && (
                      <div style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.75rem', color: '#10b981', marginTop: '4px' }}>
                        <Check size={12} /> Pul qaytarildi
                      </div>
                    )}
                  </td>
                  <td>
                    {r.status === 'pending' ? (
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        <button className="btn-primary" style={{ padding: '0.4rem', background: '#10b981', borderColor: '#10b981' }} onClick={() => openApproveModal(r)} title="Tasdiqlash">
                          <Check size={16} />
                        </button>
                        <button className="btn-secondary" style={{ padding: '0.4rem', color: '#ef4444', borderColor: '#fecaca', background: '#fef2f2' }} onClick={() => handleReject(r.id)} title="Rad etish">
                          <X size={16} />
                        </button>
                      </div>
                    ) : (
                      <div style={{ display: 'flex', gap: '0.5rem' }}>
                        {r.status === 'approved' && !r.money_returned && (
                          <button className="btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.8rem' }} onClick={() => markMoneyReturned(r.id)}>
                            <Banknote size={14} style={{ marginRight: '4px', display: 'inline' }} /> Pul berildi
                          </button>
                        )}
                        {r.approved_by_name && (
                          <span style={{ fontSize: '0.8rem', color: '#94a3b8' }}>
                            {r.status === 'approved' ? 'Tasdiqladi' : 'Rad etdi'}: {r.approved_by_name}
                          </span>
                        )}
                      </div>
                    )}
                  </td>
                </tr>
              ))}
              {refunds.length === 0 && (
                <tr><td colSpan="7" className="text-center p-4">So'rovlar topilmadi</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>

      {showApproveModal && selectedRefund && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content" style={{ maxWidth: '400px' }}>
            <div className="modal-header">
              <h3>Refund Tasdiqlash</h3>
              <button onClick={() => setShowApproveModal(false)}><X size={20} /></button>
            </div>
            
            <div style={{ background: '#f8fafc', padding: '1rem', borderRadius: '8px', marginBottom: '1.5rem', border: '1px solid #e2e8f0' }}>
              <div style={{ fontSize: '0.9rem', color: '#64748b', marginBottom: '4px' }}>Chek #{selectedRefund.order_id}</div>
              <div style={{ fontSize: '1.4rem', fontWeight: 800, color: '#0f172a', marginBottom: '8px' }}>{selectedRefund.amount?.toLocaleString()} so'm</div>
              <div style={{ fontSize: '0.9rem', color: '#475569' }}><strong>Sabab:</strong> {selectedRefund.reason}</div>
            </div>

            <div className="input-group">
              <label>Qaytarish usuli</label>
              <select value={refundMethod} onChange={e => setRefundMethod(e.target.value)}>
                <option value="cash">Naqd</option>
                <option value="card">Karta</option>
                <option value="click">Click/Payme</option>
              </select>
            </div>

            <button className="btn-primary w-full mt-4" style={{ padding: '0.9rem', fontSize: '1.05rem', background: '#10b981', borderColor: '#10b981' }} onClick={handleApprove}>
              <Check size={18} style={{ marginRight: '8px' }} /> Tasdiqlash
            </button>
          </motion.div>
        </div>
      )}
    </div>
  );
};

export default RefundsSection;

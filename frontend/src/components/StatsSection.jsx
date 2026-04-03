import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion } from 'framer-motion';
import { Calendar, Clock, BarChart3, TrendingUp, Award } from 'lucide-react';

const StatsSection = ({ role }) => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchStats = async () => {
    try {
      const res = await api.get('/orders/stats');
      setStats(res.data);
    } catch (err) {
      console.error('Failed to fetch stats', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchStats(); }, []);

  useWebsocket((data) => {
    if (data.type === 'status_update' || data.type === 'new_order') fetchStats();
  });

  const getRoleTitle = () => {
    switch (role) {
      case 'admin': return 'Umumiy Statistika';
      case 'cook': return 'Oshxona Statistikasi';
      case 'courier': return 'Sizning Statistikangiz';
      default: return 'Statistika';
    }
  };

  const statItems = stats ? [
    { label: 'Bugun', value: stats.today, icon: <Clock size={18} />, bg: 'rgba(99,102,241,0.15)', color: '#818cf8' },
    { label: '3 kunlik', value: stats.three_day, icon: <Calendar size={18} />, bg: 'rgba(139,92,246,0.15)', color: '#a78bfa' },
    { label: 'Haftalik', value: stats.week, icon: <BarChart3 size={18} />, bg: 'rgba(16,185,129,0.15)', color: '#34d399' },
    { label: 'Oylik', value: stats.month, icon: <TrendingUp size={18} />, bg: 'rgba(249,115,22,0.15)', color: '#fb923c' },
    { label: 'Yillik', value: stats.year, icon: <Award size={18} />, bg: 'rgba(239,68,68,0.15)', color: '#fca5a5' },
  ] : [];

  return (
    <div className="ss-wrap">
      <div className="ss-header">
        <BarChart3 size={18} className="ss-icon" />
        <span className="ss-title">{getRoleTitle()}</span>
      </div>

      {loading ? (
        <div className="ss-grid">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="ss-card ss-skeleton" />
          ))}
        </div>
      ) : (
        <div className="ss-grid">
          {statItems.map((item, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 16 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.07 }}
              className="ss-card"
            >
              <div className="ss-icon-box" style={{ background: item.bg, color: item.color }}>
                {item.icon}
              </div>
              <div>
                <div className="ss-value">{(item.value || 0).toLocaleString()}</div>
                <div className="ss-label">{item.label}</div>
              </div>
            </motion.div>
          ))}
        </div>
      )}

      <style>{`
        .ss-wrap { margin-bottom: 2rem; }

        .ss-header {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          margin-bottom: 1rem;
        }

        .ss-icon { color: var(--primary); }

        .ss-title {
          font-size: 1rem;
          font-weight: 700;
          color: var(--text-secondary);
          text-transform: uppercase;
          letter-spacing: 0.05em;
          font-size: 0.82rem;
        }

        .ss-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
          gap: 0.75rem;
        }

        .ss-card {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          background: var(--bg-card);
          border: 1px solid var(--border);
          border-radius: var(--radius);
          padding: 0.9rem 1rem;
          transition: var(--transition);
          backdrop-filter: blur(12px);
        }

        .ss-card:hover {
          border-color: rgba(249,115,22,0.25);
          transform: translateY(-2px);
        }

        .ss-skeleton {
          height: 70px;
          background: linear-gradient(90deg, var(--bg-card) 25%, rgba(255,255,255,0.06) 50%, var(--bg-card) 75%);
          background-size: 200% 100%;
          animation: shimmer 1.5s infinite;
          border-radius: var(--radius);
        }

        .ss-icon-box {
          width: 38px; height: 38px;
          border-radius: 10px;
          display: flex; align-items: center; justify-content: center;
          flex-shrink: 0;
        }

        .ss-value {
          font-size: 1.5rem;
          font-weight: 800;
          font-family: var(--font-display);
          background: var(--grad-brand);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
          background-clip: text;
          line-height: 1;
        }

        .ss-label {
          font-size: 0.72rem;
          color: var(--text-secondary);
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.04em;
          margin-top: 0.2rem;
        }

        @media (max-width: 640px) {
          .ss-grid { grid-template-columns: repeat(2, 1fr); }
        }
      `}</style>
    </div>
  );
};

export default StatsSection;

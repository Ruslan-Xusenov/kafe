import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { useWebsocket } from '../hooks/useWebsocket';
import { motion } from 'framer-motion';
import { Calendar, Clock, BarChart3, TrendingUp, Award, Loader2 } from 'lucide-react';

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

  useEffect(() => {
    fetchStats();
  }, []);

  useWebsocket((data) => {
    if (data.type === 'status_update' || data.type === 'new_order') {
      fetchStats();
    }
  });

  if (loading) return (
    <div className="flex justify-center p-8">
      <Loader2 className="animate-spin text-primary" size={32} />
    </div>
  );

  if (!stats) return null;

  const statItems = [
    { label: 'Bugun', value: stats.today, icon: <Clock size={20} />, color: '#3b82f6' },
    { label: '3 kunlik', value: stats.three_day, icon: <Calendar size={20} />, color: '#8b5cf6' },
    { label: 'Haftalik', value: stats.week, icon: <BarChart3 size={20} />, color: '#10b981' },
    { label: 'Oylik', value: stats.month, icon: <TrendingUp size={20} />, color: '#f59e0b' },
    { label: 'Yillik', value: stats.year, icon: <Award size={20} />, color: '#ef4444' },
  ];

  const getRoleTitle = () => {
    switch (role) {
      case 'admin': return 'Umumiy Statistika';
      case 'cook': return 'Oshxona Statistikasi';
      case 'courier': return 'Sizning Statistikangiz';
      default: return 'Statistika';
    }
  };

  return (
    <div className="stats-container mb-8">
      <h2 className="text-xl font-bold mb-4 flex items-center gap-2">
        <BarChart3 className="text-primary" />
        {getRoleTitle()}
      </h2>
      <div className="stats-grid">
        {statItems.map((item, index) => (
          <motion.div
            key={index}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className="stats-card premium-card"
          >
            <div className="stats-icon" style={{ backgroundColor: `${item.color}20`, color: item.color }}>
              {item.icon}
            </div>
            <div className="stats-info">
              <span className="stats-label">{item.label}</span>
              <span className="stats-value">{item.value.toLocaleString()}</span>
            </div>
          </motion.div>
        ))}
      </div>

      <style>{`
        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
          gap: 1rem;
        }
        .stats-card {
          display: flex;
          align-items: center;
          gap: 1rem;
          padding: 1.25rem;
          background: var(--bg-surface);
          border: 1px solid rgba(255, 255, 255, 0.05);
        }
        .stats-icon {
          width: 48px;
          height: 48px;
          border-radius: 12px;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .stats-info {
          display: flex;
          flex-direction: column;
        }
        .stats-label {
          font-size: 0.85rem;
          color: var(--text-dim);
          font-weight: 500;
        }
        .stats-value {
          font-size: 1.5rem;
          font-weight: 800;
          color: var(--text-main);
        }
        @media (max-width: 640px) {
          .stats-grid {
            grid-template-columns: repeat(2, 1fr);
          }
        }
      `}</style>
    </div>
  );
};

export default StatsSection;

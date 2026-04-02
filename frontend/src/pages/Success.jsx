import React from 'react';
import { useLocation, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { CheckCircle, Package, ArrowLeft, Ticket } from 'lucide-react';

const Success = () => {
  const location = useLocation();
  const orderId = location.state?.orderId || '#8213';

  return (
    <div className="success-page animate-fade">
      <motion.div 
        initial={{ scale: 0.8, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        className="premium-card success-card"
      >
        <div className="success-header">
          <CheckCircle size={80} className="success-icon" />
          <h1>Buyurtma qabul qilindi!</h1>
          <p>Sizning buyurtmangiz muvaffaqiyatli rasmiylashtirildi.</p>
        </div>

        <div className="order-details glass">
          <div className="detail-item">
            <span className="label">Buyurtma ID</span>
            <span className="value">#{orderId}</span>
          </div>
          <div className="detail-item">
            <span className="label">Holat</span>
            <span className="value status-badge">Yangi</span>
          </div>
        </div>

        <div className="success-actions">
          <Link to="/" className="btn-primary home-btn">
            <ArrowLeft size={20} /> Menyuga qaytish
          </Link>
        </div>

        <div className="success-footer">
          <Package size={20} />
          <span>Oshpazlarimiz taomingizni tayyorlashni boshladilar!</span>
        </div>
      </motion.div>

      <style>{`
        .success-page {
          display: flex;
          justify-content: center;
          align-items: center;
          min-height: calc(100vh - 150px);
          padding: 2rem;
        }

        .success-card {
          width: 100%;
          max-width: 500px;
          text-align: center;
          padding: 3rem;
        }

        .success-icon {
          color: #10b981;
          margin-bottom: 2rem;
        }

        .success-header h1 {
          font-size: 2rem;
          margin-bottom: 0.75rem;
        }

        .success-header p {
          color: var(--text-dim);
          margin-bottom: 3rem;
        }

        .order-details {
          padding: 1.5rem;
          border-radius: 12px;
          margin-bottom: 3rem;
          display: flex;
          flex-direction: column;
          gap: 1rem;
        }

        .detail-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
        }

        .label {
          color: var(--text-dim);
          font-size: 0.9rem;
        }

        .value {
          font-weight: 700;
          font-size: 1.1rem;
        }

        .status-badge {
          background: var(--primary);
          padding: 4px 12px;
          border-radius: 20px;
          font-size: 0.8rem;
          color: white;
        }

        .success-actions {
          margin-bottom: 2.5rem;
        }

        .home-btn {
          display: flex;
          justify-content: center;
          align-items: center;
          gap: 1rem;
          padding: 1rem 2rem;
          text-decoration: none;
        }

        .success-footer {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 1rem;
          color: var(--text-dim);
          font-size: 0.9rem;
        }
      `}</style>
    </div>
  );
};

export default Success;

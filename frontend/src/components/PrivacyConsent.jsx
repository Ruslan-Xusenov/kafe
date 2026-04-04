import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Cookie, X, Check } from 'lucide-react';

const PrivacyConsent = () => {
  const [show, setShow] = useState(false);

  useEffect(() => {
    const consent = localStorage.getItem('privacy-consent');
    if (!consent) {
      // Show after a short delay
      const timer = setTimeout(() => setShow(true), 1500);
      return () => clearTimeout(timer);
    }
  }, []);

  const handleAccept = () => {
    localStorage.setItem('privacy-consent', 'true');
    setShow(false);
  };

  return (
    <AnimatePresence>
      {show && (
        <motion.div
          initial={{ y: 100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: 100, opacity: 0 }}
          transition={{ type: 'spring', damping: 25, stiffness: 200 }}
          className="cookie-consent-fixed"
        >
          <div className="cookie-card">
            <div className="cookie-icon-wrap">
              <Cookie size={24} className="cookie-icon" />
            </div>
            <div className="cookie-content">
              <h3>Cookie-dan foydalanish</h3>
              <p>
                Saytimizda eng yaxshi tajribani taqdim etish uchun cookie-fayllardan foydalanamiz. 
                Davom etish orqali siz bizning shartlarimizga rozilik bildirasiz.
              </p>
            </div>
            <div className="cookie-actions">
              <button className="cookie-btn accept" onClick={handleAccept}>
                <Check size={18} />
                Tushunarli
              </button>
            </div>
          </div>

          <style>{`
            .cookie-consent-fixed {
              position: fixed;
              bottom: 5.5rem;
              left: 1rem;
              right: 1rem;
              z-index: 1000;
              display: flex;
              justify-content: center;
              pointer-events: none;
            }
            .cookie-card {
              pointer-events: auto;
              background: rgba(23, 23, 26, 0.85);
              backdrop-filter: blur(20px);
              -webkit-backdrop-filter: blur(20px);
              border: 1px solid rgba(255, 255, 255, 0.1);
              border-radius: 24px;
              padding: 1.25rem;
              max-width: 500px;
              width: 100%;
              display: flex;
              align-items: center;
              gap: 1.25rem;
              box-shadow: 0 20px 50px rgba(0, 0, 0, 0.4);
            }
            .cookie-icon-wrap {
              width: 48px;
              height: 48px;
              background: rgba(249, 115, 22, 0.15);
              border-radius: 16px;
              display: flex;
              align-items: center;
              justify-content: center;
              flex-shrink: 0;
            }
            .cookie-icon {
              color: #f97316;
            }
            .cookie-content {
              flex: 1;
            }
            .cookie-content h3 {
              color: white;
              font-size: 1rem;
              font-weight: 700;
              margin-bottom: 0.25rem;
            }
            .cookie-content p {
              color: rgba(255, 255, 255, 0.6);
              font-size: 0.85rem;
              line-height: 1.4;
              margin: 0;
            }
            .cookie-actions {
              display: flex;
              gap: 0.75rem;
            }
            .cookie-btn {
              display: flex;
              align-items: center;
              gap: 0.5rem;
              padding: 0.6rem 1.1rem;
              border-radius: 12px;
              font-size: 0.85rem;
              font-weight: 700;
              cursor: pointer;
              transition: all 0.2s;
              border: none;
            }
            .cookie-btn.accept {
              background: #f97316;
              color: white;
              box-shadow: 0 4px 12px rgba(249, 115, 22, 0.3);
            }
            .cookie-btn.accept:hover {
              background: #fb923c;
              transform: translateY(-2px);
              box-shadow: 0 6px 15px rgba(249, 115, 22, 0.4);
            }
            .cookie-btn.accept:active {
              transform: translateY(0);
            }
            
            @media (max-width: 600px) {
              .cookie-card {
                flex-direction: column;
                text-align: center;
                gap: 1rem;
                padding: 1.5rem;
              }
              .cookie-actions {
                width: 100%;
              }
              .cookie-btn {
                width: 100%;
                justify-content: center;
              }
            }
          `}</style>
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default PrivacyConsent;

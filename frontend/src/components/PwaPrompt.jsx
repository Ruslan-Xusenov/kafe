import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Download, Share, PlusSquare } from 'lucide-react';

const CAFE_NAME = import.meta.env.VITE_CAFE_NAME || 'KafePlat';

const isIOSDevice = () => /iphone|ipad|ipod/.test(window.navigator.userAgent.toLowerCase());

const PwaPrompt = () => {
  const [deferredPrompt, setDeferredPrompt] = useState(null);
  const [showPrompt, setShowPrompt] = useState(false);
  const [isIOS] = useState(isIOSDevice);

  useEffect(() => {
    // Check if the app is already installed (standalone mode)
    const isStandalone = window.matchMedia('(display-mode: standalone)').matches || window.navigator.standalone;
    
    // Check if the user already dismissed the prompt recently (e.g. 3 days)
    const lastDismissed = localStorage.getItem('pwaPromptDismissed');
    if (lastDismissed) {
      const daysSinceDismissed = (new Date().getTime() - new Date(lastDismissed).getTime()) / (1000 * 3600 * 24);
      if (daysSinceDismissed < 3) return; // Don't show if dismissed within last 3 days
    }

    if (isStandalone) {
      return; // Already installed
    }

    // iOS detection
    if (isIOS) {
      // iOS doesn't support beforeinstallprompt, show custom instruction immediately
      setTimeout(() => setShowPrompt(true), 3000); // 3 sec delay
    }

    // Android / Desktop detection
    const handleBeforeInstallPrompt = (e) => {
      e.preventDefault();
      setDeferredPrompt(e);
      setTimeout(() => setShowPrompt(true), 3000); // 3 sec delay
    };

    window.addEventListener('beforeinstallprompt', handleBeforeInstallPrompt);

    return () => {
      window.removeEventListener('beforeinstallprompt', handleBeforeInstallPrompt);
    };
  }, [isIOS]);

  const handleInstallClick = async () => {
    if (deferredPrompt) {
      deferredPrompt.prompt();
      const { outcome } = await deferredPrompt.userChoice;
      if (outcome === 'accepted') {
        console.log('User accepted the install prompt');
        setShowPrompt(false);
      } else {
        console.log('User dismissed the install prompt');
      }
      setDeferredPrompt(null);
    }
  };

  const handleDismiss = () => {
    setShowPrompt(false);
    localStorage.setItem('pwaPromptDismissed', new Date().toISOString());
  };

  return (
    <AnimatePresence>
      {showPrompt && (
        <motion.div
          initial={{ y: -100, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: -100, opacity: 0 }}
          transition={{ type: 'spring', stiffness: 200, damping: 20 }}
          style={{
            position: 'fixed',
            top: 16,
            left: '50%',
            transform: 'translateX(-50%)',
            zIndex: 9999,
            width: '90%',
            maxWidth: '400px',
            backgroundColor: '#ffffff',
            color: '#1a1a1a',
            borderRadius: '16px',
            boxShadow: '0 8px 32px rgba(0,0,0,0.12)',
            border: '1px solid #e5e7eb',
            padding: '16px',
            display: 'flex',
            flexDirection: 'column',
            gap: '12px'
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
              <img src="/icon-192x192.png" alt={CAFE_NAME} style={{ width: '48px', height: '48px', borderRadius: '10px' }} />
              <div>
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 'bold', color: '#1a1a1a' }}>{CAFE_NAME}</h3>
                <p style={{ margin: 0, fontSize: '13px', color: '#6b7280' }}>Ilovani o'rnating va tezroq kiring</p>
              </div>
            </div>
            <button 
              onClick={handleDismiss}
              style={{ background: 'none', border: 'none', color: '#9ca3af', cursor: 'pointer', padding: '4px' }}
            >
              <X size={20} />
            </button>
          </div>

          {isIOS ? (
            <div style={{ backgroundColor: '#f3f4f6', padding: '12px', borderRadius: '8px', fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '8px', color: '#4b5563' }}>
              <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                1. <Share size={16} /> Ulashish (Share) tugmasini bosing
              </span>
              <span style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                2. <PlusSquare size={16} /> "Add to Home Screen" ni tanlang
              </span>
            </div>
          ) : (
            <button
              onClick={handleInstallClick}
              style={{
                backgroundColor: '#f97316',
                color: 'white',
                border: 'none',
                borderRadius: '8px',
                padding: '10px',
                fontWeight: 'bold',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                width: '100%',
                boxShadow: '0 4px 12px rgba(249,115,22,0.35)'
              }}
            >
              <Download size={18} /> O'rnatish
            </button>
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
};

export default PwaPrompt;

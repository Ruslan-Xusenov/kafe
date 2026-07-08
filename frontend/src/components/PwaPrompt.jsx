import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Download, Share, PlusSquare } from 'lucide-react';

const PwaPrompt = () => {
  const [deferredPrompt, setDeferredPrompt] = useState(null);
  const [showPrompt, setShowPrompt] = useState(false);
  const [isIOS, setIsIOS] = useState(false);

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
    const userAgent = window.navigator.userAgent.toLowerCase();
    const isIosDevice = /iphone|ipad|ipod/.test(userAgent);
    setIsIOS(isIosDevice);

    if (isIosDevice) {
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
  }, []);

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
            backgroundColor: '#1e1e1e',
            color: '#fff',
            borderRadius: '16px',
            boxShadow: '0 10px 25px rgba(0,0,0,0.5)',
            border: '1px solid rgba(255,255,255,0.1)',
            padding: '16px',
            display: 'flex',
            flexDirection: 'column',
            gap: '12px'
          }}
        >
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
              <img src="/icon-192x192.png" alt="KafePlat" style={{ width: '48px', height: '48px', borderRadius: '10px' }} />
              <div>
                <h3 style={{ margin: 0, fontSize: '16px', fontWeight: 'bold' }}>KafePlat</h3>
                <p style={{ margin: 0, fontSize: '13px', color: '#aaa' }}>Ilovani o'rnating va tezroq kiring</p>
              </div>
            </div>
            <button 
              onClick={handleDismiss}
              style={{ background: 'none', border: 'none', color: '#888', cursor: 'pointer', padding: '4px' }}
            >
              <X size={20} />
            </button>
          </div>

          {isIOS ? (
            <div style={{ backgroundColor: 'rgba(255,255,255,0.05)', padding: '12px', borderRadius: '8px', fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '8px' }}>
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
                backgroundColor: '#ff7b00',
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
                width: '100%'
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

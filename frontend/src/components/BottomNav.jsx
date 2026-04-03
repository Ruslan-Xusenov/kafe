import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import { Utensils, ShoppingCart, Package, ChefHat, Truck, LayoutDashboard } from 'lucide-react';

const BottomNav = () => {
  const { user, isAuthenticated } = useAuthStore();
  const { items } = useCartStore();
  const location = useLocation();
  const cartCount = items.reduce((s, i) => s + i.quantity, 0);

  const isActive = (path) => location.pathname === path;

  const tabs = [
    { to: '/', icon: <Utensils size={22} />, label: 'Menyu', always: true },
    { to: '/cart', icon: <ShoppingCart size={22} />, label: 'Savat', badge: cartCount > 0 ? cartCount : null, always: true },
    { to: '/my-orders', icon: <Package size={22} />, label: 'Buyurtmalar', auth: true },
    { to: '/kitchen', icon: <ChefHat size={22} />, label: 'Oshxona', roles: ['admin','cook'] },
    { to: '/delivery', icon: <Truck size={22} />, label: 'Yetkazish', roles: ['admin','courier'] },
    { to: '/admin', icon: <LayoutDashboard size={22} />, label: 'Admin', roles: ['admin'] },
  ].filter(tab => {
    if (tab.always) return true;
    if (tab.auth && !isAuthenticated) return false;
    if (tab.roles && (!user || !tab.roles.includes(user.role))) return false;
    if (tab.auth) return isAuthenticated;
    return true;
  }).slice(0, 5); // max 5 tab

  return (
    <>
      <nav className="bottom-nav">
        {tabs.map(tab => (
          <Link
            key={tab.to}
            to={tab.to}
            className={`bnav-item ${isActive(tab.to) ? 'active' : ''}`}
          >
            <div className="bnav-icon-wrap">
              {tab.icon}
              {tab.badge && (
                <span className="bnav-badge">{tab.badge > 9 ? '9+' : tab.badge}</span>
              )}
            </div>
            <span className="bnav-label">{tab.label}</span>
          </Link>
        ))}
      </nav>

      <style>{`
        .bottom-nav {
          display: none;
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          z-index: 999;
          background: rgba(13,13,15,0.92);
          backdrop-filter: blur(24px) saturate(180%);
          -webkit-backdrop-filter: blur(24px) saturate(180%);
          border-top: 1px solid rgba(255,255,255,0.08);
          padding: 0.4rem 0.5rem calc(0.4rem + env(safe-area-inset-bottom));
          justify-content: space-around;
          align-items: center;
          box-shadow: 0 -8px 32px rgba(0,0,0,0.4);
        }

        .bnav-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          gap: 0.25rem;
          padding: 0.4rem 0.6rem;
          border-radius: 10px;
          text-decoration: none;
          color: var(--text-muted);
          transition: var(--transition);
          flex: 1;
          position: relative;
          min-width: 0;
        }

        .bnav-item:hover {
          color: var(--text-secondary);
        }

        .bnav-item.active {
          color: var(--primary);
        }

        .bnav-icon-wrap {
          position: relative;
          display: flex;
          align-items: center;
          justify-content: center;
          width: 40px;
          height: 32px;
          border-radius: 10px;
          transition: var(--transition);
        }

        .bnav-item.active .bnav-icon-wrap {
          background: rgba(249,115,22,0.15);
        }

        .bnav-badge {
          position: absolute;
          top: -4px;
          right: -4px;
          background: var(--grad-brand);
          color: white;
          font-size: 0.6rem;
          font-weight: 800;
          min-width: 16px;
          height: 16px;
          border-radius: 99px;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 0 3px;
          box-shadow: 0 2px 6px rgba(249,115,22,0.5);
        }

        .bnav-label {
          font-size: 0.65rem;
          font-weight: 600;
          letter-spacing: 0.01em;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
          max-width: 100%;
        }

        /* Only show on mobile */
        @media (max-width: 900px) {
          .bottom-nav { display: flex; }
          /* Add bottom padding to main content so it's not hidden behind nav */
          .main-content { padding-bottom: calc(70px + 1rem) !important; }
        }

        @media (max-width: 380px) {
          .bnav-label { display: none; }
          .bnav-item { padding: 0.4rem 0.3rem; }
          .bnav-icon-wrap { width: 36px; height: 36px; }
        }
      `}</style>
    </>
  );
};

export default BottomNav;

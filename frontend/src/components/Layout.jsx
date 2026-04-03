import React from 'react';
import { Outlet, Link, useNavigate, useLocation } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import {
  ShoppingCart, LogOut, Utensils, LayoutDashboard,
  Truck, ChefHat, Package, User, Menu, X
} from 'lucide-react';
import { useState } from 'react';

const Layout = () => {
  const { user, isAuthenticated, logout } = useAuthStore();
  const { items } = useCartStore();
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);

  const cartCount = items.reduce((sum, item) => sum + item.quantity, 0);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const isActive = (path) => location.pathname === path;

  const navLinks = [
    { to: '/', label: 'Menyu', icon: <Utensils size={16} />, always: true },
    { to: '/my-orders', label: 'Buyurtmalarim', icon: <Package size={16} />, auth: true },
    { to: '/admin', label: 'Dashboard', icon: <LayoutDashboard size={16} />, roles: ['admin'] },
    { to: '/kitchen', label: 'Oshxona', icon: <ChefHat size={16} />, roles: ['admin','cook'] },
    { to: '/delivery', label: 'Yetkazish', icon: <Truck size={16} />, roles: ['admin','courier'] },
  ].filter(link => {
    if (link.always) return true;
    if (link.auth && !isAuthenticated) return false;
    if (link.roles && (!user || !link.roles.includes(user.role))) return false;
    if (link.auth) return isAuthenticated;
    return true;
  });

  return (
    <div className="layout-root">
      {/* ── NAVBAR ── */}
      <header className="navbar glass">
        <div className="navbar-inner">
          {/* Logo */}
          <Link to="/" className="navbar-logo">
            <div className="logo-icon">
              <Utensils size={18} />
            </div>
            <span className="logo-text">Kafe<span>Plat</span></span>
          </Link>

          {/* Desktop nav links */}
          <nav className="navbar-links">
            {navLinks.map(link => (
              <Link
                key={link.to}
                to={link.to}
                className={`nav-link ${isActive(link.to) ? 'active' : ''}`}
              >
                {link.icon}
                {link.label}
              </Link>
            ))}
          </nav>

          {/* Actions */}
          <div className="navbar-actions">
            {/* Cart */}
            <button className="icon-btn cart-btn" onClick={() => navigate('/cart')}>
              <ShoppingCart size={20} />
              {cartCount > 0 && <span className="cart-badge">{cartCount}</span>}
            </button>

            {isAuthenticated ? (
              <div className="user-chip">
                <div className="user-avatar">
                  {(user?.full_name || 'U')[0].toUpperCase()}
                </div>
                <span className="user-chip-name">{user?.full_name?.split(' ')[0]}</span>
                <button className="icon-btn logout-btn" onClick={handleLogout} title="Chiqish">
                  <LogOut size={16} />
                </button>
              </div>
            ) : (
              <Link to="/login" className="btn-primary nav-login-btn">Kirish</Link>
            )}

            {/* Mobile burger */}
            <button className="icon-btn mobile-menu-btn" onClick={() => setMobileOpen(!mobileOpen)}>
              {mobileOpen ? <X size={22} /> : <Menu size={22} />}
            </button>
          </div>
        </div>

        {/* Mobile nav */}
        {mobileOpen && (
          <nav className="mobile-nav">
            {navLinks.map(link => (
              <Link
                key={link.to}
                to={link.to}
                className={`mobile-nav-link ${isActive(link.to) ? 'active' : ''}`}
                onClick={() => setMobileOpen(false)}
              >
                {link.icon}
                {link.label}
              </Link>
            ))}
          </nav>
        )}
      </header>

      {/* ── MAIN ── */}
      <main className="main-content">
        <Outlet />
      </main>

      <style>{`
        .layout-root { min-height: 100vh; display: flex; flex-direction: column; }

        /* ── NAVBAR ── */
        .navbar {
          position: fixed;
          top: 0; left: 0; right: 0;
          z-index: 1000;
          border-bottom: 1px solid var(--border);
        }

        .navbar-inner {
          display: flex;
          align-items: center;
          justify-content: space-between;
          height: 68px;
          max-width: 1280px;
          margin: 0 auto;
          padding: 0 1.5rem;
          gap: 1.5rem;
        }

        /* Logo */
        .navbar-logo {
          display: flex;
          align-items: center;
          gap: 0.65rem;
          text-decoration: none;
          flex-shrink: 0;
        }

        .logo-icon {
          width: 36px; height: 36px;
          background: var(--grad-brand);
          border-radius: 10px;
          display: flex; align-items: center; justify-content: center;
          color: white;
          box-shadow: 0 4px 12px rgba(249,115,22,0.40);
          flex-shrink: 0;
        }

        .logo-text {
          font-family: var(--font-display);
          font-size: 1.25rem;
          font-weight: 700;
          color: var(--text-primary);
        }

        .logo-text span { color: var(--primary); }

        /* Nav links */
        .navbar-links {
          display: flex;
          align-items: center;
          gap: 0.25rem;
          flex: 1;
          justify-content: center;
        }

        .nav-link {
          display: flex;
          align-items: center;
          gap: 0.4rem;
          padding: 0.5rem 0.9rem;
          border-radius: 8px;
          font-size: 0.88rem;
          font-weight: 600;
          color: var(--text-secondary);
          text-decoration: none;
          transition: var(--transition);
        }

        .nav-link:hover {
          color: var(--text-primary);
          background: var(--bg-surface);
        }

        .nav-link.active {
          color: var(--primary);
          background: rgba(249,115,22,0.10);
        }

        /* Actions */
        .navbar-actions {
          display: flex;
          align-items: center;
          gap: 0.75rem;
          flex-shrink: 0;
        }

        .icon-btn {
          background: var(--bg-surface);
          color: var(--text-primary);
          width: 38px; height: 38px;
          border-radius: 10px;
          border: 1px solid var(--border);
          display: flex; align-items: center; justify-content: center;
          transition: var(--transition);
        }

        .icon-btn:hover {
          border-color: var(--primary);
          color: var(--primary);
          background: rgba(249,115,22,0.08);
        }

        .cart-btn { position: relative; }

        .cart-badge {
          position: absolute;
          top: -6px; right: -6px;
          background: var(--grad-brand);
          color: white;
          font-size: 0.68rem;
          font-weight: 800;
          width: 18px; height: 18px;
          border-radius: 50%;
          display: flex; align-items: center; justify-content: center;
          box-shadow: 0 2px 8px rgba(249,115,22,0.5);
          animation: scaleIn 0.3s ease;
        }

        /* User chip */
        .user-chip {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 10px;
          padding: 0.35rem 0.75rem 0.35rem 0.4rem;
          transition: var(--transition);
        }

        .user-avatar {
          width: 28px; height: 28px;
          background: var(--grad-brand);
          border-radius: 8px;
          display: flex; align-items: center; justify-content: center;
          font-size: 0.8rem;
          font-weight: 800;
          color: white;
          flex-shrink: 0;
        }

        .user-chip-name {
          font-size: 0.875rem;
          font-weight: 600;
          color: var(--text-primary);
        }

        .logout-btn {
          background: none;
          border: none;
          color: var(--text-secondary);
          width: auto; height: auto;
          padding: 0.2rem;
        }

        .logout-btn:hover { color: var(--danger); }

        .nav-login-btn {
          font-size: 0.85rem;
          padding: 0.5rem 1.1rem;
          border-radius: 8px;
        }

        /* Mobile */
        .mobile-menu-btn { display: none; }

        .mobile-nav {
          display: flex;
          flex-direction: column;
          gap: 0.25rem;
          padding: 0.75rem 1rem 1rem;
          border-top: 1px solid var(--border);
          animation: fadeUp 0.25s ease;
        }

        .mobile-nav-link {
          display: flex;
          align-items: center;
          gap: 0.6rem;
          padding: 0.7rem 0.9rem;
          border-radius: 8px;
          font-weight: 600;
          font-size: 0.9rem;
          color: var(--text-secondary);
          transition: var(--transition);
        }

        .mobile-nav-link:hover,
        .mobile-nav-link.active {
          color: var(--primary);
          background: rgba(249,115,22,0.10);
        }

        /* Main content */
        .main-content {
          margin-top: 68px;
          flex: 1;
          max-width: 1280px;
          width: 100%;
          margin-left: auto;
          margin-right: auto;
          padding: 2rem 1.5rem 4rem;
        }

        @media (max-width: 768px) {
          .navbar-links { display: none; }
          .user-chip-name { display: none; }
          .mobile-menu-btn { display: flex; }
          .main-content { padding: 1.25rem 1rem 3rem; }
        }
      `}</style>
    </div>
  );
};

export default Layout;

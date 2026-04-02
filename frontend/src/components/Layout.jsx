import React from 'react';
import { Outlet, Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { useCartStore } from '../store/cartStore';
import { ShoppingCart, User, LogOut, Utensils, LayoutDashboard, Truck, ChefHat, Package } from 'lucide-react';

const Layout = () => {
  const { user, isAuthenticated, logout } = useAuthStore();
  const { items } = useCartStore();
  const navigate = useNavigate();
  
  const cartCount = items.reduce((sum, item) => sum + item.quantity, 0);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="layout-container">
      <header className="glass header animate-fade">
        <nav className="nav-bar">
          <Link to="/" className="logo">
            <Utensils /> <span>Kafe Plat</span>
          </Link>
          
          <div className="nav-links">
            <Link to="/">Menyu</Link>
            {isAuthenticated && (
              <Link to="/my-orders" className="nav-icon-link"><Package size={20} /> Buyurtmalarim</Link>
            )}
            
            {isAuthenticated && user?.role === 'admin' && (
              <Link to="/admin" className="nav-icon-link"><LayoutDashboard size={20} /> Dashboard</Link>
            )}
            
            {isAuthenticated && (user?.role === 'admin' || user?.role === 'cook') && (
              <Link to="/kitchen" className="nav-icon-link"><ChefHat size={20} /> Oshxona</Link>
            )}
            
            {isAuthenticated && (user?.role === 'admin' || user?.role === 'courier') && (
              <Link to="/delivery" className="nav-icon-link"><Truck size={20} /> Yetkazish</Link>
            )}
          </div>
          
          <div className="nav-actions">
            <button className="cart-btn" onClick={() => navigate('/cart')}>
              <ShoppingCart size={20} />
              {cartCount > 0 && <span className="cart-badge">{cartCount}</span>}
            </button>
            
            {isAuthenticated ? (
              <div className="user-profile">
                <span className="user-name">{user?.full_name || 'Mijoz'}</span>
                <button className="logout-btn" onClick={handleLogout}>
                  <LogOut size={18} />
                </button>
              </div>
            ) : (
              <Link to="/login" className="btn-primary login-btn">Kirish</Link>
            )}
          </div>
        </nav>
      </header>

      <main className="main-content">
        <Outlet />
      </main>

      <style>{`
        .header {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          height: 70px;
          display: flex;
          align-items: center;
          padding: 0 2rem;
          z-index: 1000;
        }

        .nav-bar {
          display: flex;
          justify-content: space-between;
          align-items: center;
          width: 100%;
          max-width: 1200px;
          margin: 0 auto;
        }

        .logo {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          color: var(--primary);
          font-weight: 700;
          font-family: 'Outfit', sans-serif;
          font-size: 1.25rem;
          text-decoration: none;
        }

        .nav-links {
          display: flex;
          gap: 2rem;
        }

        .nav-links a {
          color: var(--text-dim);
          text-decoration: none;
          font-weight: 500;
          transition: var(--transition);
          display: flex;
          align-items: center;
          gap: 0.5rem;
        }

        .nav-links a:hover {
          color: var(--text-main);
        }

        .nav-actions {
          display: flex;
          align-items: center;
          gap: 1.5rem;
        }

        .cart-btn {
          background: none;
          color: var(--text-main);
          position: relative;
        }

        .cart-badge {
          position: absolute;
          top: -8px;
          right: -8px;
          background: var(--primary);
          font-size: 0.75rem;
          padding: 2px 6px;
          border-radius: 10px;
        }

        .user-profile {
          display: flex;
          align-items: center;
          gap: 1rem;
          background: var(--bg-surface);
          padding: 0.5rem 1rem;
          border-radius: 20px;
          border: 1px solid var(--border);
        }

        .logout-btn {
          background: none;
          color: #ef4444;
          display: flex;
          align-items: center;
        }

        .main-content {
          margin-top: 90px;
          padding: 2rem;
          max-width: 1200px;
          margin-left: auto;
          margin-right: auto;
        }

        @media (max-width: 768px) {
          .nav-links { display: none; }
          .header { padding: 0 1rem; }
        }
      `}</style>
    </div>
  );
};

export default Layout;

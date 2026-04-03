import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { validateFullName, validatePhone, validatePassword } from '../utils/validation';
import { motion } from 'framer-motion';
import { UserPlus, User, Phone, Lock, ChefHat, Eye, EyeOff } from 'lucide-react';

const Register = () => {
  const [fullName, setFullName] = useState('');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [showPass, setShowPass] = useState(false);
  const [errors, setErrors] = useState({});
  const { register, loading, error: serverError } = useAuthStore();
  const navigate = useNavigate();

  const handleValidate = () => {
    const newErrors = {
      fullName: validateFullName(fullName),
      phone: validatePhone(phone),
      password: validatePassword(password)
    };
    setErrors(newErrors);
    return !newErrors.fullName && !newErrors.phone && !newErrors.password;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!handleValidate()) return;
    const res = await register(fullName, phone, password);
    if (res.success) navigate('/');
  };

  return (
    <div className="auth-page">
      <div className="auth-bg-blob" />

      <motion.div
        initial={{ opacity: 0, y: 32 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: [0.4, 0, 0.2, 1] }}
        className="auth-card"
      >
        {/* Header */}
        <div className="auth-header">
          <div className="auth-brand-icon">
            <ChefHat size={28} />
          </div>
          <h1 className="auth-title">Hisob yaratish</h1>
          <p className="auth-subtitle">Tezkor buyurtma uchun ro'yxatdan o'ting</p>
        </div>

        {/* Error */}
        {serverError && (
          <motion.div
            initial={{ opacity: 0, x: -10 }}
            animate={{ opacity: 1, x: 0 }}
            className="auth-error"
          >
            ⚠️ {serverError}
          </motion.div>
        )}

        <form onSubmit={handleSubmit} className="auth-form">
          {/* Full Name */}
          <div className={`auth-field ${errors.fullName ? 'has-error' : ''}`}>
            <label>
              <User size={15} /> To'liq ismingiz
            </label>
            <div className="field-input-wrap">
              <input
                type="text"
                placeholder="Ali Valiyev"
                value={fullName}
                onChange={e => {
                  setFullName(e.target.value);
                  if (errors.fullName) setErrors({ ...errors, fullName: null });
                }}
              />
            </div>
            {errors.fullName && <span className="field-error">{errors.fullName}</span>}
          </div>

          {/* Phone */}
          <div className={`auth-field ${errors.phone ? 'has-error' : ''}`}>
            <label>
              <Phone size={15} /> Telefon raqami
            </label>
            <div className="field-input-wrap">
              <input
                type="text"
                placeholder="+998 90 123 45 67"
                value={phone}
                onChange={e => {
                  setPhone(e.target.value);
                  if (errors.phone) setErrors({ ...errors, phone: null });
                }}
              />
            </div>
            {errors.phone && <span className="field-error">{errors.phone}</span>}
          </div>

          {/* Password */}
          <div className={`auth-field ${errors.password ? 'has-error' : ''}`}>
            <label>
              <Lock size={15} /> Parol
            </label>
            <div className="field-input-wrap pass-wrap">
              <input
                type={showPass ? 'text' : 'password'}
                placeholder="••••••••"
                value={password}
                onChange={e => {
                  setPassword(e.target.value);
                  if (errors.password) setErrors({ ...errors, password: null });
                }}
              />
              <button type="button" className="pass-toggle" onClick={() => setShowPass(!showPass)}>
                {showPass ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
            </div>
            {errors.password && <span className="field-error">{errors.password}</span>}
          </div>

          <button type="submit" className="auth-submit btn-primary" disabled={loading}>
            {loading ? (
              <span className="btn-loading">
                <span className="spinner-sm" /> Yaratilmoqda...
              </span>
            ) : (
              <>
                <UserPlus size={18} /> Ro'yxatdan o'tish
              </>
            )}
          </button>
        </form>

        <p className="auth-footer-text">
          Akkauntingiz bormi?{' '}
          <Link to="/login" className="auth-link">Tizimga kiring</Link>
        </p>
      </motion.div>

      <style>{`
        .auth-page {
          min-height: calc(100vh - 68px);
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 2rem 1rem;
          position: relative;
          overflow: hidden;
        }

        .auth-bg-blob {
          position: fixed;
          width: 600px; height: 600px;
          background: radial-gradient(circle, rgba(249,115,22,0.12) 0%, transparent 70%);
          border-radius: 50%;
          top: 50%; left: 50%;
          transform: translate(-50%, -50%);
          pointer-events: none;
          animation: pulse 4s ease-in-out infinite;
        }

        @keyframes pulse {
          0%, 100% { opacity: 0.7; transform: translate(-50%,-50%) scale(1); }
          50% { opacity: 1; transform: translate(-50%,-50%) scale(1.1); }
        }

        .auth-card {
          width: 100%;
          max-width: 420px;
          background: rgba(255,255,255,0.04);
          border: 1px solid var(--border);
          border-radius: var(--radius-xl);
          padding: 2.5rem 2rem;
          backdrop-filter: blur(24px);
          -webkit-backdrop-filter: blur(24px);
          box-shadow: 0 24px 64px rgba(0,0,0,0.5), inset 0 1px 0 rgba(255,255,255,0.08);
          position: relative;
          z-index: 1;
        }

        .auth-header {
          text-align: center;
          margin-bottom: 2rem;
        }

        .auth-brand-icon {
          width: 64px; height: 64px;
          background: var(--grad-brand);
          border-radius: 18px;
          display: flex; align-items: center; justify-content: center;
          color: white;
          margin: 0 auto 1.25rem;
          box-shadow: 0 8px 24px rgba(249,115,22,0.40);
        }

        .auth-title {
          font-size: 1.75rem;
          margin-bottom: 0.4rem;
          color: var(--text-primary);
        }

        .auth-subtitle {
          color: var(--text-secondary);
          font-size: 0.9rem;
        }

        .auth-error {
          background: rgba(239,68,68,0.10);
          border: 1px solid rgba(239,68,68,0.25);
          color: #fca5a5;
          padding: 0.8rem 1rem;
          border-radius: var(--radius-sm);
          font-size: 0.875rem;
          margin-bottom: 1.25rem;
        }

        .auth-form { display: flex; flex-direction: column; gap: 0.15rem; }

        .auth-field {
          display: flex;
          flex-direction: column;
          gap: 0.45rem;
          margin-bottom: 0.85rem;
        }

        .auth-field label {
          font-size: 0.82rem;
          font-weight: 700;
          color: var(--text-secondary);
          display: flex;
          align-items: center;
          gap: 0.4rem;
          text-transform: uppercase;
          letter-spacing: 0.04em;
        }

        .field-input-wrap { position: relative; }

        .pass-wrap input { padding-right: 2.8rem; }

        .pass-toggle {
          position: absolute;
          right: 0.9rem; top: 50%;
          transform: translateY(-50%);
          background: none;
          border: none;
          color: var(--text-muted);
          padding: 0.2rem;
          transition: var(--transition);
        }

        .pass-toggle:hover { color: var(--primary); }

        .auth-field.has-error input {
          border-color: rgba(239,68,68,0.5);
          background: rgba(239,68,68,0.05);
        }

        .field-error {
          font-size: 0.78rem;
          color: #fca5a5;
          display: flex;
          align-items: center;
          gap: 0.3rem;
        }

        .auth-submit {
          width: 100%;
          margin-top: 0.5rem;
          padding: 0.9rem;
          font-size: 0.95rem;
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 0.5rem;
          border-radius: var(--radius-sm);
        }

        .btn-loading {
          display: flex;
          align-items: center;
          gap: 0.6rem;
        }

        .spinner-sm {
          width: 16px; height: 16px;
          border: 2px solid rgba(255,255,255,0.3);
          border-top-color: white;
          border-radius: 50%;
          animation: spin 0.7s linear infinite;
          display: inline-block;
        }

        .auth-footer-text {
          text-align: center;
          color: var(--text-secondary);
          font-size: 0.875rem;
          margin-top: 1.5rem;
        }

        .auth-link {
          color: var(--primary);
          font-weight: 700;
          text-decoration: none;
          transition: var(--transition);
        }

        .auth-link:hover {
          color: var(--primary-light);
          text-decoration: underline;
        }

        @media (max-width: 640px) {
          .auth-card {
            padding: 2rem 1.4rem;
            border-radius: var(--radius-lg);
          }
          .auth-brand-icon { width: 54px; height: 54px; border-radius: 14px; }
          .auth-title { font-size: 1.5rem; }
        }

        @media (max-width: 380px) {
          .auth-card { padding: 1.75rem 1.1rem; margin: 0 0.25rem; }
          .auth-brand-icon { width: 46px; height: 46px; }
        }
      `}</style>
    </div>
  );
};

export default Register;

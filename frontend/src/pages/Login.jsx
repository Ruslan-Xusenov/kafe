import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { validatePhone, validatePassword } from '../utils/validation';
import { motion } from 'framer-motion';
import { LogIn, Phone, Lock } from 'lucide-react';

const Login = () => {
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [errors, setErrors] = useState({});
  const { login, loading, error: serverError } = useAuthStore();
  const navigate = useNavigate();

  const handleValidate = () => {
    const newErrors = {
      phone: validatePhone(phone),
      password: validatePassword(password)
    };
    setErrors(newErrors);
    return !newErrors.phone && !newErrors.password;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!handleValidate()) return;

    const res = await login(phone, password);
    if (res.success) {
      if (res.role === 'admin') navigate('/admin');
      else if (res.role === 'cook') navigate('/kitchen');
      else if (res.role === 'courier') navigate('/delivery');
      else navigate('/');
    }
  };

  return (
    <div className="login-page">
      <motion.div 
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="premium-card auth-card"
      >
        <div className="auth-header">
          <LogIn size={40} className="auth-icon" />
          <h2>Xush kelibsiz</h2>
          <p>Tizimga kirish uchun ma'lumotlarni kiriting</p>
        </div>

        {serverError && <div className="error-msg">{serverError}</div>}

        <form onSubmit={handleSubmit}>
          <div className={`input-group ${errors.phone ? 'has-error' : ''}`}>
            <label><Phone size={16} /> Telefon raqami</label>
            <input 
              type="text" 
              placeholder="+998 90 123 45 67" 
              value={phone} 
              onChange={(e) => {
                setPhone(e.target.value);
                if (errors.phone) setErrors({...errors, phone: null});
              }}
            />
            {errors.phone && <span className="field-error">{errors.phone}</span>}
          </div>

          <div className={`input-group ${errors.password ? 'has-error' : ''}`}>
            <label><Lock size={16} /> Parol</label>
            <input 
              type="password" 
              placeholder="••••••••" 
              value={password}
              onChange={(e) => {
                setPassword(e.target.value);
                if (errors.password) setErrors({...errors, password: null});
              }}
            />
            {errors.password && <span className="field-error">{errors.password}</span>}
          </div>

          <button type="submit" className="btn-primary auth-submit" disabled={loading}>
            {loading ? 'Kirilmoqda...' : 'Tizimga kirish'}
          </button>
        </form>

        <div className="auth-footer">
          Akkauntingiz yo'qmi? <Link to="/register">Ro'yxatdan o'ting</Link>
        </div>
      </motion.div>

      <style>{`
        .login-page {
          display: flex;
          justify-content: center;
          align-items: center;
          min-height: calc(100vh - 150px);
        }

        .auth-card {
          width: 100%;
          max-width: 400px;
          text-align: center;
        }

        .auth-header {
          margin-bottom: 2rem;
        }

        .auth-icon {
          color: var(--primary);
          margin-bottom: 1rem;
        }

        .auth-header h2 {
          font-size: 1.75rem;
          margin-bottom: 0.5rem;
        }

        .auth-header p {
          color: var(--text-dim);
          font-size: 0.9rem;
        }

        .input-group label {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          font-size: 0.9rem;
          color: var(--text-dim);
        }

        .auth-submit {
          width: 100%;
          margin-top: 1rem;
          font-size: 1rem;
        }

        .error-msg {
          background: rgba(239, 68, 68, 0.1);
          color: #ef4444;
          padding: 0.75rem;
          border-radius: 8px;
          margin-bottom: 1.5rem;
          border: 1px solid rgba(239, 68, 68, 0.2);
          font-size: 0.9rem;
        }

        .field-error {
          color: #ef4444;
          font-size: 0.75rem;
          margin-top: 0.25rem;
          text-align: left;
          display: block;
        }

        .input-group.has-error input {
          border-color: #ef4444;
        }

        .auth-footer {
          margin-top: 1.5rem;
          color: var(--text-dim);
          font-size: 0.9rem;
        }

        .auth-footer a {
          color: var(--primary);
          text-decoration: none;
          font-weight: 500;
        }
      `}</style>
    </div>
  );
};

export default Login;

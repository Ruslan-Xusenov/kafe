import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from './store/authStore';

// Layouts & Pages
import Layout from './components/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Home from './pages/Home';
import Cart from './pages/Cart';
import Checkout from './pages/Checkout';
import Success from './pages/Success';
import Kitchen from './pages/Kitchen';
import Courier from './pages/Courier';
import Admin from './pages/Admin';
import Printer from './pages/Printer';
import MyOrders from './pages/MyOrders';
import ProductDetail from './pages/ProductDetail';
import PrivacyConsent from './components/PrivacyConsent';

const ProtectedRoute = ({ children, allowedRoles }) => {
  const { isAuthenticated, user, loading } = useAuthStore();
  
  if (loading) return <div>Yuklanmoqda...</div>;
  if (!isAuthenticated) return <Navigate to="/login" />;
  if (allowedRoles && !allowedRoles.includes(user?.role)) {
    return <Navigate to="/" />;
  }
  return children;
};

function App() {
  const checkAuth = useAuthStore(state => state.checkAuth);

  useEffect(() => {
    checkAuth();
  }, []);

  return (
    <BrowserRouter>
      <PrivacyConsent />
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Home />} />
          <Route path="login" element={<Login />} />
          <Route path="register" element={<Register />} />
          <Route path="cart" element={<Cart />} />
          <Route path="checkout" element={<ProtectedRoute><Checkout /></ProtectedRoute>} />
          <Route path="success" element={<ProtectedRoute><Success /></ProtectedRoute>} />
          <Route path="my-orders" element={<ProtectedRoute><MyOrders /></ProtectedRoute>} />
          <Route path="product/:id" element={<ProductDetail />} />
          
          <Route path="admin" element={
            <ProtectedRoute allowedRoles={['admin']}>
              <Admin />
            </ProtectedRoute>
          } />
          
          <Route path="printer" element={
            <ProtectedRoute allowedRoles={['admin']}>
              <Printer />
            </ProtectedRoute>
          } />
          
          <Route path="kitchen" element={
            <ProtectedRoute allowedRoles={['admin', 'cook']}>
              <Kitchen />
            </ProtectedRoute>
          } />
          
          <Route path="delivery" element={
            <ProtectedRoute allowedRoles={['admin', 'courier']}>
              <Courier />
            </ProtectedRoute>
          } />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;

import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import StatsSection from '../components/StatsSection';
import InventorySection from '../components/InventorySection';
import { validateNotEmpty, validatePrice, validatePhone, validatePassword } from '../utils/validation';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  LayoutDashboard, ShoppingBag, Users, Plus, Edit2, Trash2, 
  CheckCircle, XCircle, Clock, Loader2, Save, X, ChefHat, Truck, Star, RefreshCw, Settings, Wallet, TrendingUp, Package, Printer, CreditCard
} from 'lucide-react';

const STATUS_MAP = {
  new: 'Новый',
  preparing: 'Готовится',
  ready: 'Готов',
  on_way: 'В пути',
  delivered: 'Доставлен',
  cancelled: 'Отменен'
};

const Admin = () => {
  const [activeTab, setActiveTab] = useState('orders');
  const [orders, setOrders] = useState([]);
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [staff, setStaff] = useState([]);
  const [tables, setTables] = useState([]);
  const [performance, setPerformance] = useState([]);
  const [loading, setLoading] = useState(true);
  const [containerPrice, setContainerPrice] = useState('1000');
  const [containerId, setContainerId] = useState('7');
  const [tableServicePercentage, setTableServicePercentage] = useState('10');

  const [expenses, setExpenses] = useState([]);
  const [financeStats, setFinanceStats] = useState({ total_revenue: 0, total_expenses: 0, net_profit: 0 });
  const [newExpense, setNewExpense] = useState({ amount: '', category: 'mahsulot', description: '' });

  const [showCatModal, setShowCatModal] = useState(false);
  const [newCat, setNewCat] = useState({ name: '', image_url: '', is_user_controlled: false, printer_target: 'ALL' });
  const [editCatId, setEditCatId] = useState(null);
  const [showProdModal, setShowProdModal] = useState(false);
  const [newProd, setNewProd] = useState({ 
    name: '', description: '', price: '', category_id: '', image_url: '',
    unit: 'шт', min_quantity: 1, quantity_step: 1, has_mandatory_container: false,
    is_active: true
  });
  const [editProdId, setEditProdId] = useState(null);
  const [showStaffModal, setShowStaffModal] = useState(false);
  const [newStaff, setNewStaff] = useState({ full_name: '', phone: '', password: '', role: 'cook' });
  const [showOrderModal, setShowOrderModal] = useState(false);
  const [selectedOrderDetails, setSelectedOrderDetails] = useState(null);
  const [serviceFeePercent, setServiceFeePercent] = useState(10);
  const [showTableModal, setShowTableModal] = useState(false);
  const [newTable, setNewTable] = useState({ name: '', capacity: '' });
  const [editingTableId, setEditingTableId] = useState(null);

  const [showTransferModal, setShowTransferModal] = useState(false);
  const [transferFromTable, setTransferFromTable] = useState(null);
  const [transferToTableId, setTransferToTableId] = useState('');

  const [selectedWaiter, setSelectedWaiter] = useState(null);
  const [waiterOrders, setWaiterOrders] = useState([]);
  const [waiterHistory, setWaiterHistory] = useState([]);
  const [showWaiterHistory, setShowWaiterHistory] = useState(false);
  const [waiterDefaultFees, setWaiterDefaultFees] = useState({}); 
  const [waiterOrderFees, setWaiterOrderFees] = useState({});
  const [waiterOrdersLoading, setWaiterOrdersLoading] = useState(false);

  const [waiterSalaries, setWaiterSalaries] = useState([]);
  const [salaryStartDate, setSalaryStartDate] = useState(new Date().toISOString().split('T')[0]);
  const [salaryEndDate, setSalaryEndDate] = useState(new Date().toISOString().split('T')[0]);
  const [salariesLoading, setSalariesLoading] = useState(false);

  const [errors, setErrors] = useState({});

  useEffect(() => {
    fetchData();
  }, [activeTab]);

  const handleCloseShift = async () => {
    if (!window.confirm("Вы уверены, что хотите закрыть смену? Данные за сегодня будут отправлены в Telegram, а текущая статистика обнулится.")) return;
    try {
      await api.post('/finance/close-shift');
      alert("Смена успешно закрыта. Отчет отправлен в Telegram и распечатан!");
      fetchData();
    } catch (err) {
      alert("Ошибка при закрытии смены: " + (err.response?.data?.error || err.message));
    }
  };

  const handleSendRealProfit = async () => {
    try {
      const res = await api.post('/finance/send-real-profit');
      alert(res.data.message || "Реальная прибыль отправлена!");
      fetchData();
    } catch (err) {
      alert("Ошибка при отправке: " + (err.response?.data?.error || err.message));
    }
  };

  const fetchData = async () => {
    setLoading(true);
    try {
      if (activeTab === 'orders') {
        const res = await api.get('/orders/all');
        setOrders(res.data || []);
      } else if (activeTab === 'menu') {
        const [catRes, prodRes] = await Promise.all([
          api.get('/catalog/categories'),
          api.get('/catalog/products')
        ]);
        setCategories(catRes.data || []);
        setProducts(prodRes.data || []);
      } else if (activeTab === 'staff') {
        const res = await api.get('/catalog/staff');
        setStaff(res.data || []);
      } else if (activeTab === 'performance') {
        const res = await api.get('/catalog/performance');
        setPerformance(Array.isArray(res.data) ? res.data : []);
      } else if (activeTab === 'settings') {
        const res = await api.get('/catalog/settings');
        setContainerPrice(res.data.container_price || '1000');
        setContainerId(res.data.container_product_id || '7');
        setTableServicePercentage(res.data.table_service_percentage || '10');
      } else if (activeTab === 'inventory') {
        const res = await api.get('/catalog/products');
        setProducts(res.data || []);
      } else if (activeTab === 'finance') {
        const [statsRes, expRes] = await Promise.all([
          api.get('/finance/stats'),
          api.get('/finance/expenses')
        ]);
        setFinanceStats(statsRes.data || { total_revenue: 0, total_expenses: 0, net_profit: 0 });
        setExpenses(expRes.data || []);
      } else if (activeTab === 'tables') {
        const res = await api.get('/tables');
        setTables(res.data || []);
      } else if (activeTab === 'waiterMgmt') {
        const res = await api.get('/catalog/staff');
        const waiters = (res.data || []).filter(s => s.role === 'waiter');
        setStaff(waiters);
        const defaults = {};
        waiters.forEach(w => { defaults[w.id] = w.default_service_percentage || 0; });
        setWaiterDefaultFees(defaults);
      } else if (activeTab === 'salary') {
        fetchWaiterSalaries();
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const fetchWaiterSalaries = async () => {
    setSalariesLoading(true);
    try {
      const res = await api.get(`/finance/waiter-salaries?start_date=${salaryStartDate}&end_date=${salaryEndDate}`);
      setWaiterSalaries(res.data || []);
    } catch (err) {
      alert("Ошибка при загрузке зарплат: " + (err.response?.data?.error || err.message));
    } finally {
      setSalariesLoading(false);
    }
  };

  const fetchWaiterOrders = async (waiter) => {
    setSelectedWaiter(waiter);
    setWaiterOrders([]);
    setWaiterHistory([]);
    setShowWaiterHistory(false);
    setWaiterOrdersLoading(true);
    try {
      const res = await api.get(`/orders/waiter-active/${waiter.id}`);
      const orders = res.data || [];
      setWaiterOrders(orders);
      const fees = {};
      orders.forEach(o => { fees[o.id] = o.service_percentage || 0; });
      setWaiterOrderFees(fees);
    } catch (err) {
      alert("Ошибка при загрузке заказов официанта");
    } finally {
      setWaiterOrdersLoading(false);
    }
  };

  const fetchWaiterHistory = async (waiter) => {
    setSelectedWaiter(waiter);
    setWaiterOrders([]);
    setWaiterOrdersLoading(true);
    try {
      const res = await api.get(`/orders/waiter-hist/${waiter.id}`);
      setWaiterHistory(res.data || []);
      setShowWaiterHistory(true);
    } catch (err) {
      alert("Ошибка при загрузке истории: " + (err.response?.data?.error || err.message));
    } finally {
      setWaiterOrdersLoading(false);
    }
  };

  const handleWaiterDefaultFee = async (waiterId, fee) => {
    try {
      await api.put(`/catalog/staff/${waiterId}/default-fee`, { percentage: parseFloat(fee || 0) });
      setStaff(prev => prev.map(s => s.id === waiterId ? { ...s, default_service_percentage: parseFloat(fee || 0) } : s));
      alert(`Процент по умолчанию для официанта сохранен!`);
    } catch (err) {
      alert("Ошибка при сохранении: " + (err.response?.data?.error || err.message));
    }
  };

  const handleWaiterOrderFee = async (orderId) => {
    const pct = parseFloat(waiterOrderFees[orderId] || 0);
    try {
      const res = await api.put(`/orders/${orderId}/service-fee`, { percentage: pct });
      setWaiterOrders(prev => prev.map(o => o.id === orderId ? res.data : o));
      alert(`Заказ #${orderId} на ${pct}% плата за обслуживание добавлена!`);
    } catch (err) {
      alert("Ошибка при добавлении процента: " + (err.response?.data?.error || err.message));
    }
  };

  const handleWaiterOrderFeeAndPrint = async (orderId) => {
    const pct = parseFloat(waiterOrderFees[orderId] || 0);
    try {
      const res = await api.put(`/orders/${orderId}/service-fee`, { percentage: pct });
      setWaiterOrders(prev => prev.map(o => o.id === orderId ? res.data : o));
      await api.post(`/orders/${orderId}/print`);
      alert(`Чек #${orderId} отправлен на принтер!`);
    } catch (err) {
      alert("Ошибка: " + (err.response?.data?.error || err.message));
    }
  };

  const handleCreateCat = async (e) => {
    e.preventDefault();
    const err = validateNotEmpty(newCat.name, 'Название');
    if (err) { setErrors({ cat: err }); return; }

    try {
      if (editCatId) {
        await api.put(`/catalog/categories/${editCatId}`, newCat);
      } else {
        await api.post('/catalog/categories', newCat);
      }
      setShowCatModal(false);
      setNewCat({ name: '', image_url: '', is_user_controlled: false, printer_target: 'ALL' });
      setEditCatId(null);
      setErrors({});
      fetchData();
    } catch (err) { alert(err.response?.data?.error || 'Ошибка при сохранении категории'); }
  };

  const openEditCat = (cat) => {
    setNewCat({
      name: cat.name,
      image_url: cat.image_url || '',
      is_user_controlled: cat.is_user_controlled || false,
      printer_target: cat.printer_target || 'ALL'
    });
    setEditCatId(cat.id);
    setShowCatModal(true);
  };

  const handleCreateProd = async (e) => {
    e.preventDefault();
    const prodErrors = {
      name: validateNotEmpty(newProd.name, 'Название'),
      price: validatePrice(newProd.price),
      category: validateNotEmpty(newProd.category_id, 'Категория')
    };
    if (prodErrors.name || prodErrors.price || prodErrors.category) {
      setErrors({ prod: prodErrors });
      return;
    }

    try {
      const prodData = { 
        ...newProd, 
        price: parseFloat(newProd.price), 
        category_id: parseInt(newProd.category_id),
        min_quantity: parseFloat(newProd.min_quantity),
        quantity_step: parseFloat(newProd.quantity_step)
      };
      if (editProdId) {
        await api.put(`/catalog/products/${editProdId}`, prodData);
      } else {
        await api.post('/catalog/products', prodData);
      }
      setShowProdModal(false);
      setNewProd({ 
        name: '', description: '', price: '', category_id: '', image_url: '',
        unit: 'шт', min_quantity: 1, quantity_step: 1, has_mandatory_container: false,
        is_active: true
      });
      setEditProdId(null);
      setErrors({});
      fetchData();
    } catch (err) { alert(err.response?.data?.error || 'Ошибка при сохранении продукта'); }
  };

  const openEditProd = (p) => {
    setEditProdId(p.id);
    setNewProd({
      name: p.name,
      description: p.description || '',
      price: p.price,
      category_id: p.category_id,
      image_url: p.image_url || '',
      unit: p.unit || 'шт',
      min_quantity: p.min_quantity || 1,
      quantity_step: p.quantity_step || 1,
      has_mandatory_container: p.has_mandatory_container || false,
      is_active: p.is_active
    });
    setShowProdModal(true);
  };

  const handleCreateStaff = async (e) => {
    e.preventDefault();
    const staffErrors = {
      name: validateNotEmpty(newStaff.full_name, 'Имя'),
      phone: validatePhone(newStaff.phone),
      password: validatePassword(newStaff.password)
    };
    if (staffErrors.name || staffErrors.phone || staffErrors.password) {
      setErrors({ staff: staffErrors });
      return;
    }

    try {
      await api.post('/catalog/staff', newStaff);
      setShowStaffModal(false);
      setNewStaff({ full_name: '', phone: '', password: '', role: 'cook' });
      setErrors({});
      fetchData();
    } catch (err) {
      alert(err.response?.data?.error || 'Ошибка при добавлении сотрудника');
    }
  };

  const handleUpload = async (e, setter) => {
    const file = e.target.files[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    try {
      setLoading(true);
      const res = await api.post('/catalog/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      setter(prev => ({ ...prev, image_url: res.data.url }));
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Ошибка при загрузке изображения');
    } finally {
      setLoading(false);
    }
  };

  const deleteStaff = async (id) => {
    if (!window.confirm('Haqiqatan ham bu xodimni o\'chirmoqchimisiz?')) return;
    try {
      await api.delete(`/catalog/staff/${id}`);
      fetchData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert(err.response?.data?.error || 'Ошибка при удалении сотрудника');
    }
  };

  const deleteCat = async (id) => {
    if (!window.confirm('Вы действительно хотите удалить эту категорию? Вложенные продукты также могут быть удалены!')) return;
    try {
      await api.delete(`/catalog/categories/${id}`);
      fetchData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert(err.response?.data?.error || 'Ошибка при удалении категории. Сначала удалите продукты в ней.');
    }
  };

  const deleteProd = async (id) => {
    if (window.confirm('Вы уверены, что хотите удалить?')) {
      try {
        await api.delete(`/catalog/products/${id}`);
        fetchData();
      // eslint-disable-next-line no-unused-vars
      } catch (err) { alert('Ошибка'); }
    }
  };

  const handleUpdateSettings = async (e) => {
    e.preventDefault();
    try {
      setLoading(true);
      await api.put('/catalog/settings', { 
        container_price: containerPrice,
        container_product_id: containerId,
        table_service_percentage: tableServicePercentage
      });
      alert('Настройки сохранены');
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Ошибка: ' + (err.response?.data?.error || 'Не удалось сохранить'));
    } finally {
      setLoading(false);
    }
  };

  const handleCreateExpense = async (e) => {
    e.preventDefault();
    if (!newExpense.amount) {
      setErrors({ expense: 'Необходимо ввести сумму' });
      return;
    }
    try {
      await api.post('/finance/expenses', {
        amount: parseFloat(newExpense.amount),
        category: newExpense.category,
        description: newExpense.description
      });
      setNewExpense({ amount: '', category: 'mahsulot', description: '' });
      setErrors({});
      fetchData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Ошибка при добавлении расхода: ' + (err.response?.data?.error || ''));
    }
  };

  const handleCreateTable = async (e) => {
    e.preventDefault();
    try {
      if (editingTableId) {
        await api.put(`/tables/${editingTableId}`, { 
          name: newTable.name,
          capacity: parseInt(newTable.capacity) || 4 
        });
      } else {
        await api.post('/tables', { 
          name: newTable.name,
          capacity: parseInt(newTable.capacity) || 4 
        });
      }
      setShowTableModal(false);
      setNewTable({ name: '', capacity: '' });
      setEditingTableId(null);
      fetchData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert(err.response?.data?.error || 'Ошибка при сохранении стола');
    }
  };

  const openEditTable = (t) => {
    setNewTable({ name: t.name, capacity: t.capacity || 4 });
    setEditingTableId(t.id);
    setShowTableModal(true);
  };

  const deleteTable = async (id) => {
    if (!window.confirm('Rostdan ham bu stolni o\'chirmoqchimisiz?')) return;
    try {
      await api.delete(`/tables/${id}`);
      fetchData();
      // eslint-disable-next-line no-unused-vars
    } catch (err) { alert('Ошибка'); }
  };

  const handleTransferTable = async (e) => {
    e.preventDefault();
    if (!transferToTableId) return alert("Iltimos, ko'chirish uchun bo'sh stolni tanlang");
    try {
      await api.post('/orders/transfer', {
        from_table_id: transferFromTable.id,
        to_table_id: parseInt(transferToTableId)
      });
      setShowTransferModal(false);
      setTransferFromTable(null);
      setTransferToTableId('');
      fetchData();
      alert("Stol muvaffaqiyatli ko'chirildi!");
    } catch (err) {
      alert(err.response?.data?.error || "Stolni ko'chirishda xatolik yuz berdi");
    }
  };

  const openOrderModal = async (order) => {
    try {
      const res = await api.get(`/orders/${order.id}`);
      setSelectedOrderDetails(res.data);
      setShowOrderModal(true);
      setServiceFeePercent(res.data.service_percentage || 10);
    } catch (err) {
      alert("Ошибка при загрузке данных заказа");
    }
  };

  const handleReprintOrder = async (id) => {
    try {
      await api.post(`/orders/${id}/print`);
      alert("Чек отправлен на принтер!");
    } catch (err) {
      alert("Ошибка при печати чека");
    }
  };

  const handleRemoveOrderItem = async (orderId, item) => {
    const qtyInput = window.prompt(`Сколько вы хотите отменить?\nМаксимум: ${item.quantity}`, item.quantity);
    if (qtyInput === null) return; // User cancelled
    
    const qty = parseFloat(qtyInput);
    if (isNaN(qty) || qty <= 0) {
      alert("Iltimos, to'g'ri raqam kiriting!");
      return;
    }
    if (qty > item.quantity) {
      alert(`Kiritilgan miqdor joriy miqdordan oshmasligi kerak (Maks: ${item.quantity})!`);
      return;
    }

    try {
      await api.post(`/orders/${orderId}/items/${item.id}/cancel`, { quantity: qty });
      // Refresh order details
      const res = await api.get(`/orders/${orderId}`);
      setSelectedOrderDetails(res.data);
      fetchData();
    } catch (err) {
      alert("Произошла ошибка: " + (err.response?.data?.error || err.message));
    }
  };

  const handleSetServiceFee = async (id) => {
    try {
      const res = await api.put(`/orders/${id}/service-fee`, { percentage: parseFloat(serviceFeePercent) });
      setSelectedOrderDetails(res.data);
      fetchData();
    } catch (err) {
      alert("Ошибка при обновлении платы за обслуживание: " + (err.response?.data?.error || err.message));
    }
  };

  const handleSetServiceFeeAndPrint = async (id) => {
    try {
      await api.put(`/orders/${id}/service-fee`, { percentage: parseFloat(serviceFeePercent) });
      await api.post(`/orders/${id}/print`);
      alert("Плата за обслуживание сохранена и чек отправлен на принтер!");
      setShowOrderModal(false);
      fetchData();
    } catch (err) {
      alert("Произошла ошибка");
    }
  };

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" /></div>;

  return (
    <div className="admin-page">
      <aside className="admin-sidebar glass animate-fade">
        <div className="sidebar-header">
          <LayoutDashboard className="text-primary" />
          <span>Панель Администратора</span>
        </div>
        <nav className="sidebar-nav">
          <button className={activeTab === 'orders' ? 'active' : ''} onClick={() => setActiveTab('orders')}>
            <ShoppingBag size={20} /> Заказы
          </button>
          <button className={activeTab === 'menu' ? 'active' : ''} onClick={() => setActiveTab('menu')}>
            <Edit2 size={20} /> Меню (CRUD)
          </button>
          <button className={activeTab === 'staff' ? 'active' : ''} onClick={() => setActiveTab('staff')}>
            <Users size={20} /> Сотрудники
          </button>
          <button className={activeTab === 'tables' ? 'active' : ''} onClick={() => setActiveTab('tables')}>
            <LayoutDashboard size={20} /> Столы
          </button>
          <button className={activeTab === 'performance' ? 'active' : ''} onClick={() => setActiveTab('performance')}>
            <Star size={20} /> Рейтинги
          </button>
          <button className={activeTab === 'finance' ? 'active' : ''} onClick={() => setActiveTab('finance')}>
            <Wallet size={20} /> Финансы (Прибыль)
          </button>
          <button className={activeTab === 'inventory' ? 'active' : ''} onClick={() => setActiveTab('inventory')}>
            <Package size={20} /> Склад
          </button>
          <button className={activeTab === 'waiterMgmt' ? 'active' : ''} onClick={() => setActiveTab('waiterMgmt')}>
            <Users size={20} /> Процент официанта
          </button>
          <button className={activeTab === 'salary' ? 'active' : ''} onClick={() => setActiveTab('salary')}>
            <CreditCard size={20} /> Зарплата (Ойлик)
          </button>
          <button className={activeTab === 'settings' ? 'active' : ''} onClick={() => setActiveTab('settings')}>
            <Settings size={20} /> Настройки
          </button>
        </nav>
      </aside>

      <main className="admin-main animate-fade">
        {activeTab === 'inventory' && (
          <InventorySection products={products} />
        )}

        {activeTab === 'waiterMgmt' && (
          <div className="animate-fade">
            <div className="section-header mb-6">
              <h2>Заказы официантов и управление процентами</h2>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '280px 1fr', gap: '1.5rem', alignItems: 'start' }}>
              {/* Waiters list */}
              <div className="premium-card">
                <h3 className="mb-4">Официанты</h3>
                {staff.length === 0 && <p className="text-muted">Официант не найден</p>}
                {staff.map(w => (
                  <div
                    key={w.id}
                    style={{
                      padding: '0.75rem 1rem', borderRadius: '10px',
                      marginBottom: '0.5rem',
                      background: selectedWaiter?.id === w.id ? 'rgba(249,115,22,0.15)' : 'rgba(255,255,255,0.03)',
                      border: selectedWaiter?.id === w.id ? '1px solid rgba(249,115,22,0.5)' : '1px solid var(--border)',
                      transition: 'all 0.2s'
                    }}
                  >
                    <div style={{ fontWeight: 700, marginBottom: '0.4rem' }}>{w.full_name}</div>
                    <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)', marginBottom: '0.5rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <span>{w.phone}</span>
                      <div style={{ display: 'flex', alignItems: 'center', gap: '0.3rem' }}>
                        <span style={{ fontSize: '0.75rem' }}>По ум. %:</span>
                        <input
                          type="number" min="0" max="100" step="1"
                          value={waiterDefaultFees[w.id] ?? 0}
                          onChange={e => setWaiterDefaultFees(prev => ({ ...prev, [w.id]: e.target.value }))}
                          onBlur={e => handleWaiterDefaultFee(w.id, e.target.value)}
                          style={{ width: '45px', padding: '0.1rem 0.3rem', borderRadius: '4px', background: 'var(--bg-surface)', border: '1px solid var(--border)', color: 'var(--text-primary)', textAlign: 'center', fontSize: '0.8rem' }}
                        />
                      </div>
                    </div>
                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <button
                        className="btn-primary"
                        style={{ fontSize: '0.75rem', padding: '0.3rem 0.7rem', flex: 1 }}
                        onClick={() => fetchWaiterOrders(w)}
                      >
                        ⚡ Активные
                      </button>
                      <button
                        className="btn-secondary"
                        style={{ fontSize: '0.75rem', padding: '0.3rem 0.7rem', flex: 1 }}
                        onClick={() => fetchWaiterHistory(w)}
                      >
                        📋 История
                      </button>
                    </div>
                  </div>
                ))}
              </div>

              {/* Waiter orders / history panel */}
              <div>
                {!selectedWaiter && (
                  <div className="premium-card flex-center" style={{ minHeight: '200px', color: 'var(--text-muted)' }}>
                    Выберите официанта слева
                  </div>
                )}
                {selectedWaiter && !showWaiterHistory && (
                  <div className="premium-card">
                    <h3 className="mb-4">{selectedWaiter.full_name} — Активные заказы</h3>
                    {waiterOrdersLoading && <div className="flex-center"><Loader2 className="animate-spin" /></div>}
                    {!waiterOrdersLoading && waiterOrders.length === 0 && (
                      <p className="text-muted">На данный момент активных заказов нет</p>
                    )}
                    {waiterOrders.map(order => (
                      <div key={order.id} style={{
                        background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border)',
                        borderRadius: '12px', padding: '1.25rem', marginBottom: '1rem'
                      }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.75rem' }}>
                          <div>
                            <span style={{ fontWeight: 700, fontSize: '1.1rem', color: 'var(--primary)' }}>Чек #{order.id}</span>
                            <span style={{ marginLeft: '1rem', color: 'var(--text-muted)', fontSize: '0.85rem' }}>Стол №{order.table_name || order.table_id}</span>
                          </div>
                          <span style={{ fontWeight: 800, fontSize: '1.1rem' }}>{(order.total_price || 0).toLocaleString()} сум</span>
                        </div>

                        {/* Items summary */}
                        <div style={{ fontSize: '0.85rem', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
                          {order.items?.slice(0, 3).map(item => (
                            <span key={item.id} style={{ marginRight: '0.5rem' }}>
                              {item.product_name} ×{item.quantity}{item.unit};
                            </span>
                          ))}
                          {(order.items?.length || 0) > 3 && <span>+{order.items.length - 3} других</span>}
                        </div>

                        {/* Service fee control */}
                        <div style={{ background: 'rgba(249,115,22,0.05)', border: '1px solid rgba(249,115,22,0.2)', borderRadius: '10px', padding: '1rem' }}>
                          {order.service_percentage > 0 && (
                            <div style={{ marginBottom: '0.5rem', fontSize: '0.85rem', color: 'var(--text-muted)' }}>
                              Текущий процент: <strong style={{ color: 'var(--primary)' }}>{order.service_percentage}%</strong>
                              {' '}(+{(order.service_fee || 0).toLocaleString()} сум)
                            </div>
                          )}
                          <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center', flexWrap: 'wrap' }}>
                            <label style={{ fontSize: '0.9rem', fontWeight: 600, whiteSpace: 'nowrap' }}>Обслуживание %:</label>
                            <input
                              type="number" min="0" max="100" step="1"
                              value={waiterOrderFees[order.id] ?? 0}
                              onChange={e => setWaiterOrderFees(prev => ({ ...prev, [order.id]: e.target.value }))}
                              style={{ width: '80px', padding: '0.4rem 0.6rem', borderRadius: '8px', background: 'var(--bg-surface)', border: '1px solid var(--border)', color: 'var(--text-primary)', textAlign: 'center', fontSize: '1rem', fontWeight: 700 }}
                            />
                            <span style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                              = {((order.total_price || 0) * (1 + (parseFloat(waiterOrderFees[order.id]) || 0) / 100)).toLocaleString()} сум
                            </span>
                            <button className="btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem' }} onClick={() => handleWaiterOrderFee(order.id)}>
                              Сохранить
                            </button>
                            <button className="btn-primary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.85rem', display: 'flex', alignItems: 'center', gap: '0.3rem' }} onClick={() => handleWaiterOrderFeeAndPrint(order.id)}>
                              <Printer size={14} /> Сохранить + Чек
                            </button>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}

                {/* Waiter history panel */}
                {selectedWaiter && showWaiterHistory && (
                  <div className="premium-card">
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
                      <h3>{selectedWaiter.full_name} — История заказов</h3>
                      <button className="btn-secondary" style={{ padding: '0.3rem 0.8rem', fontSize: '0.85rem' }} onClick={() => setShowWaiterHistory(false)}>
                        ← К активным
                      </button>
                    </div>
                    {waiterOrdersLoading && <div className="flex-center"><Loader2 className="animate-spin" /></div>}
                    {!waiterOrdersLoading && waiterHistory.length === 0 && (
                      <p className="text-muted">История пуста</p>
                    )}

                    {/* Summary row */}
                    {waiterHistory.length > 0 && (
                      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
                        <div style={{ background: 'rgba(16,185,129,0.1)', border: '1px solid rgba(16,185,129,0.3)', borderRadius: '10px', padding: '0.75rem 1.25rem' }}>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Всего заказов</div>
                          <div style={{ fontWeight: 800, fontSize: '1.25rem' }}>{waiterHistory.length}</div>
                        </div>
                        <div style={{ background: 'rgba(249,115,22,0.1)', border: '1px solid rgba(249,115,22,0.3)', borderRadius: '10px', padding: '0.75rem 1.25rem' }}>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Общая выручка</div>
                          <div style={{ fontWeight: 800, fontSize: '1.25rem', color: 'var(--primary)' }}>
                            {waiterHistory.reduce((acc, o) => acc + (o.total_price || 0), 0).toLocaleString()} сум
                          </div>
                        </div>
                        <div style={{ background: 'rgba(59,130,246,0.1)', border: '1px solid rgba(59,130,246,0.3)', borderRadius: '10px', padding: '0.75rem 1.25rem' }}>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Закрытые (доставлен)</div>
                          <div style={{ fontWeight: 800, fontSize: '1.25rem' }}>{waiterHistory.filter(o => o.status === 'delivered').length}</div>
                        </div>
                      </div>
                    )}

                    <div style={{ maxHeight: '500px', overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
                      {waiterHistory.map(order => (
                        <div key={order.id} style={{
                          background: 'rgba(255,255,255,0.03)', border: '1px solid var(--border)',
                          borderRadius: '12px', padding: '1rem'
                        }}>
                          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: '0.4rem' }}>
                            <span style={{ fontWeight: 700, color: 'var(--primary)' }}>#{order.id}</span>
                            <span style={{ fontWeight: 800 }}>{(order.total_price || 0).toLocaleString()} сум</span>
                          </div>
                          <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.3rem' }}>
                            <span className="text-muted">Стол №{order.table_name || order.table_id}</span>
                            <span className={`status-badge ${order.status}`}>{STATUS_MAP[order.status] || order.status}</span>
                          </div>
                          <div style={{ fontSize: '0.85rem' }}>
                            <span className="text-muted">Оплата: </span>
                            {order.payment_method === 'cash' ? '💵 Наличные' :
                             order.payment_method === 'card' ? '💳 Карта' :
                             order.payment_method === 'click' ? '📱 Click/Payme' :
                             order.payment_method === 'nasiya' ? '📒 В долг' : (order.payment_method || '—')}
                            {order.service_percentage > 0 && (
                              <span className="text-muted" style={{ marginLeft: '1rem' }}>Плата за обслуживание: {order.service_percentage}%</span>
                            )}
                          </div>
                          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginTop: '0.25rem' }}>
                            {new Date(order.created_at).toLocaleString()}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'salary' && (
          <div className="animate-fade">
            <div className="section-header mb-6">
              <h2>Зарплата официантов</h2>
              <button className="btn-primary" onClick={fetchWaiterSalaries} disabled={salariesLoading}>
                {salariesLoading ? <Loader2 className="animate-spin" size={16} /> : <RefreshCw size={16} />}
                Обновить
              </button>
            </div>

            <div className="premium-card mb-6" style={{ display: 'flex', gap: '1rem', alignItems: 'flex-end', flexWrap: 'wrap' }}>
              <div className="input-group">
                <label>От (Start Date)</label>
                <input 
                  type="date" 
                  value={salaryStartDate}
                  onChange={e => setSalaryStartDate(e.target.value)}
                  style={{ width: 'auto', background: 'var(--bg-main)' }}
                />
              </div>
              <div className="input-group">
                <label>До (End Date)</label>
                <input 
                  type="date" 
                  value={salaryEndDate}
                  onChange={e => setSalaryEndDate(e.target.value)}
                  style={{ width: 'auto', background: 'var(--bg-main)' }}
                />
              </div>
              <button className="btn-primary" onClick={fetchWaiterSalaries} style={{ padding: '0.8rem 1.5rem' }}>
                Показать
              </button>
            </div>

            <div className="premium-card">
              <h3 className="mb-4">Начислено за выбранный период</h3>
              {salariesLoading ? (
                <div className="flex-center py-8"><Loader2 className="animate-spin" /></div>
              ) : waiterSalaries.length === 0 ? (
                <p className="text-muted text-center py-8">Нет данных для отображения</p>
              ) : (
                <div className="table-container">
                  <table className="modern-table">
                    <thead>
                      <tr>
                        <th>Официант</th>
                        <th>Кол-во заказов</th>
                        <th>Начисленная зарплата (Сумма %)</th>
                      </tr>
                    </thead>
                    <tbody>
                      {waiterSalaries.map(s => (
                        <tr key={s.waiter_id}>
                          <td style={{ fontWeight: 600 }}>{s.waiter_name}</td>
                          <td>{s.total_orders}</td>
                          <td style={{ fontWeight: 800, color: 'var(--primary)' }}>
                            {s.total_salary.toLocaleString()} сум
                          </td>
                        </tr>
                      ))}
                      <tr style={{ background: 'rgba(255,255,255,0.05)' }}>
                        <td style={{ fontWeight: 800 }}>ИТОГО</td>
                        <td style={{ fontWeight: 800 }}>{waiterSalaries.reduce((acc, s) => acc + s.total_orders, 0)}</td>
                        <td style={{ fontWeight: 800, color: 'var(--primary)', fontSize: '1.2rem' }}>
                          {waiterSalaries.reduce((acc, s) => acc + s.total_salary, 0).toLocaleString()} сум
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'finance' && (
          <div className="finance-mgmt animate-fade">
            <div className="flex justify-between items-center mb-6">
              <h2>Финансы и Расходы</h2>
              <div>
                <button className="btn-success" onClick={handleSendRealProfit} style={{ background: 'var(--primary)', marginRight: '10px' }}>
                  <TrendingUp size={16} /> Отправить Реал. Прибыль
                </button>
                <button className="btn-success" onClick={handleCloseShift} style={{ background: 'var(--danger)', marginRight: '10px' }}>
                  <Clock size={16} /> Закрыть смену (Отчет)
                </button>
                <button className="refresh-btn" onClick={fetchData}><RefreshCw size={16} /> Обновить</button>
              </div>
            </div>

            {/* Stats Cards */}
            <div className="stats-grid mb-6">
              <div className="stat-card">
                <div className="stat-icon-wrap" style={{ background: 'rgba(16,185,129,0.1)', color: 'var(--success)' }}>
                  <TrendingUp size={24} />
                </div>
                <div className="stat-info">
                  <span className="stat-label">Общая выручка</span>
                  <span className="stat-value">{financeStats.total_revenue.toLocaleString()} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card">
                <div className="stat-icon-wrap" style={{ background: 'rgba(239,68,68,0.1)', color: 'var(--danger)' }}>
                  <Wallet size={24} />
                </div>
                <div className="stat-info">
                  <span className="stat-label">Общие расходы</span>
                  <span className="stat-value">{financeStats.total_expenses.toLocaleString()} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card">
                <div className="stat-icon-wrap" style={{ background: 'rgba(59,130,246,0.1)', color: '#3b82f6' }}>
                  <LayoutDashboard size={24} />
                </div>
                <div className="stat-info">
                  <span className="stat-label">Чистая прибыль (без учета з/п)</span>
                  <span className="stat-value">{financeStats.net_profit.toLocaleString()} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card" style={{ background: 'linear-gradient(135deg, rgba(249,115,22,0.1) 0%, rgba(249,115,22,0) 100%)', border: '1px solid var(--primary)' }}>
                <div className="stat-icon-wrap" style={{ background: 'var(--primary)', color: '#fff' }}>
                  <TrendingUp size={24} />
                </div>
                <div className="stat-info">
                  <span className="stat-label">Реальная прибыль</span>
                  <span className="stat-value" style={{ color: 'var(--primary)' }}>{(financeStats.real_profit || 0).toLocaleString()} <small>so'm</small></span>
                </div>
              </div>
            </div>

            <h3 className="mb-4">Выручка по видам оплат</h3>
            <div className="stats-grid mb-6">
              <div className="stat-card" style={{ background: 'linear-gradient(to right, #10b98122, #10b98111)', borderColor: '#10b98155' }}>
                <div className="stat-info">
                  <span className="stat-label">💵 Наличные pul</span>
                  <span className="stat-value text-success" style={{ fontSize: '1.25rem' }}>{financeStats.cash_revenue?.toLocaleString() || 0} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card" style={{ background: 'linear-gradient(to right, #3b82f622, #3b82f611)', borderColor: '#3b82f655' }}>
                <div className="stat-info">
                  <span className="stat-label">💳 Терминал (Карта)</span>
                  <span className="stat-value text-primary" style={{ fontSize: '1.25rem' }}>{financeStats.card_revenue?.toLocaleString() || 0} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card" style={{ background: 'linear-gradient(to right, #8b5cf622, #8b5cf611)', borderColor: '#8b5cf655' }}>
                <div className="stat-info">
                  <span className="stat-label">📱 Click/Payme</span>
                  <span className="stat-value" style={{ color: '#8b5cf6', fontSize: '1.25rem' }}>{financeStats.click_revenue?.toLocaleString() || 0} <small>so'm</small></span>
                </div>
              </div>
              <div className="stat-card" style={{ background: 'linear-gradient(to right, #f59e0b22, #f59e0b11)', borderColor: '#f59e0b55' }}>
                <div className="stat-info">
                  <span className="stat-label">📒 В долг</span>
                  <span className="stat-value" style={{ color: '#f59e0b', fontSize: '1.25rem' }}>{financeStats.nasiya_revenue?.toLocaleString() || 0} <small>so'm</small></span>
                </div>
              </div>
            </div>

            <div className="grid" style={{ gridTemplateColumns: '1fr 2fr', gap: '2rem' }}>
              {/* Add Expense Form */}
              <div className="premium-card">
                <h3 className="mb-4">Добавить расход</h3>
                <form onSubmit={handleCreateExpense}>
                  <div className="input-group mb-4">
                    <label>Сумма (сум)</label>
                    <input 
                      type="number" 
                      value={newExpense.amount} 
                      onChange={e => setNewExpense({...newExpense, amount: e.target.value})} 
                      placeholder="0"
                    />
                    {errors.expense && <span className="error-text">{errors.expense}</span>}
                  </div>
                  <div className="input-group mb-4">
                    <label>Категория</label>
                    <select 
                      value={newExpense.category} 
                      onChange={e => setNewExpense({...newExpense, category: e.target.value})}
                    >
                      <option value="mahsulot">Продукт (Склад)</option>
                      <option value="oylik">Зарплата</option>
                      <option value="arenda">Аренда</option>
                      <option value="kommunal">Коммунальные платежи</option>
                      <option value="boshqa">Другие расходы</option>
                    </select>
                  </div>
                  <div className="input-group mb-4">
                    <label>Комментарий</label>
                    <textarea 
                      value={newExpense.description} 
                      onChange={e => setNewExpense({...newExpense, description: e.target.value})}
                      placeholder="Для чего использовалось?"
                      rows="2"
                    ></textarea>
                  </div>
                  <button type="submit" className="btn-primary w-full">
                    <Plus size={18} /> Добавить
                  </button>
                </form>
              </div>

              {/* Expenses List */}
              <div className="premium-card">
                <h3 className="mb-4">История расходов</h3>
                <div className="orders-table-wrapper" style={{ maxHeight: '400px', overflowY: 'auto' }}>
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Дата</th>
                        <th>Категория</th>
                        <th>Summa</th>
                        <th>Комментарий</th>
                      </tr>
                    </thead>
                    <tbody>
                      {expenses.length > 0 ? expenses.map(exp => (
                        <tr key={exp.id}>
                          <td>{new Date(exp.created_at).toLocaleString()}</td>
                          <td>
                            <span className="status-badge preparing">{exp.category}</span>
                          </td>
                          <td style={{ color: 'var(--danger)', fontWeight: 'bold' }}>
                            -{exp.amount.toLocaleString()} сум
                          </td>
                          <td>{exp.description || '-'}</td>
                        </tr>
                      )) : (
                        <tr><td colSpan="4" className="text-center text-muted py-4">Нет расходов</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="settings-mgmt animate-fade">
            <div className="flex justify-between items-center mb-6">
              <h2>Настройки системы</h2>
            </div>
            <div className="premium-card" style={{ maxWidth: '400px' }}>
              <form onSubmit={handleUpdateSettings}>
                <div className="input-group">
                  <label>Цена одноразовой посуды (сум)</label>
                  <input 
                    type="number" 
                    value={containerPrice} 
                    onChange={e => setContainerPrice(e.target.value)} 
                  />
                </div>
                <div className="input-group mt-4">
                  <label>ID продукта одноразовой посуды (для информации)</label>
                  <input 
                    type="number" 
                    value={containerId} 
                    onChange={e => setContainerId(e.target.value)} 
                  />
                  <p className="hint-text" style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginTop: '5px' }}>
                    * Если вы создаете новый продукт для посуды в базе данных, введите его ID здесь.
                  </p>
                </div>
                <div className="input-group mt-4">
                  <label>Процент за обслуживание столов (для клиентов) %</label>
                  <input 
                    type="number" 
                    value={tableServicePercentage} 
                    onChange={e => setTableServicePercentage(e.target.value)} 
                  />
                  <p className="hint-text" style={{ fontSize: '0.8rem', color: 'var(--text-dim)', marginTop: '5px' }}>
                    * Этот процент будет добавляться к счету клиента по умолчанию.
                  </p>
                </div>
                <button type="submit" className="btn-primary w-full mt-4">
                  <Save size={18} /> Сохранить настройки
                </button>
              </form>
            </div>
          </div>
        )}

        {activeTab === 'orders' && (
          <div className="orders-mgmt">
            <StatsSection role="admin" />
            <div className="flex justify-between items-center mb-4">
              <h2>Мониторинг заказов</h2>
              <div>
                <button className="btn-success" onClick={handleSendRealProfit} style={{ background: 'var(--primary)', marginRight: '10px' }}>
                  <TrendingUp size={16} /> Отправить Реал. Прибыль
                </button>
                <button className="btn-success" onClick={handleCloseShift} style={{ background: 'var(--danger)', marginRight: '10px' }}>
                  <Clock size={16} /> Закрыть смену (Отчет)
                </button>
                <button className="refresh-btn" onClick={fetchData}>
                  <RefreshCw size={18} /> Обновить
                </button>
              </div>
            </div>
            <div className="orders-table-wrapper premium-card">
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Клиент</th>
                    <th>Итого</th>
                    <th>Статус</th>
                    <th>Дата</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map(o => (
                    <tr key={o.id}>
                      <td>#{o.id}</td>
                      <td>
                        {o.table_name ? (
                          <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                            <span style={{ fontWeight: 'bold' }}>Стол: {o.table_name}</span>
                            <span style={{ fontSize: '0.85rem', color: '#666' }}>Официант: {o.waiter_name || 'Неизвестно'}</span>
                          </div>
                        ) : (
                          o.phone
                        )}
                      </td>
                      <td>{o.total_price.toLocaleString()}</td>
                      <td><span className={`status-badge ${o.status}`}>{o.status}</span></td>
                      <td>{new Date(o.created_at).toLocaleDateString()}</td>
                      <td>
                        <button className="action-btn-small" onClick={() => openOrderModal(o)}>Просмотр</button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === 'performance' && (
          <div className="performance-mgmt animate-fade">
             <div className="flex justify-between items-center mb-6">
                <h2>Рейтинг и качество работы сотрудников</h2>
                <button className="refresh-btn" onClick={fetchData}><RefreshCw size={18} /> Обновить</button>
             </div>
             <div className="performance-grid">
               {performance.map(p => (
                 <div key={p.staff_id} className="premium-card perf-card">
                    <div className="perf-header">
                       <div className="perf-user">
                          <ChefHat className={p.role === 'cook' ? 'text-primary' : 'text-secondary'} size={32} />
                          <div>
                             <h3>{p.full_name}</h3>
                             <span className="role-chip">{p.role === 'cook' ? 'Повар' : 'Курьер'}</span>
                          </div>
                       </div>
                       <div className="perf-score">
                          <Star size={20} fill="#f1c40f" className="text-yellow-400" />
                          <span>{p.avg_rating.toFixed(1)}</span>
                       </div>
                    </div>
                    
                    <div className="perf-stats">
                       <div className="stat-item">
                          <label>Итого baholar</label>
                          <span>{p.total_reviews}</span>
                       </div>
                       <div className="stat-item positive">
                          <label>Хорошо (4-5)</label>
                          <span>{p.good_reviews}</span>
                       </div>
                       <div className="stat-item negative">
                          <label>Плохо (1-2)</label>
                          <span>{p.bad_reviews}</span>
                       </div>
                    </div>
                 </div>
               ))}
             </div>
             
             {performance.length === 0 && (
               <div className="empty-state">
                 <p>Пока ни одному сотруднику не была выставлена оценка.</p>
               </div>
             )}
          </div>
        )}

        {activeTab === 'menu' && (
          <div className="menu-mgmt">
            <div className="flex-header">
              <h2>Управление меню</h2>
              <div className="actions">
                <button className="btn-primary" onClick={() => { setEditCatId(null); setNewCat({ name: '', image_url: '', is_user_controlled: false, printer_target: 'ALL' }); setShowCatModal(true); }}><Plus size={18} /> Категория</button>
                <button className="btn-primary" onClick={() => { setEditProdId(null); setNewProd({ name: '', description: '', price: '', category_id: '', image_url: '', unit: 'шт', min_quantity: 1, quantity_step: 1, has_mandatory_container: false, is_active: true }); setShowProdModal(true); }}><Plus size={18} /> Продукт</button>
              </div>
            </div>

            <div className="menu-sections">
              <section className="cat-section mb-4">
                <h3>Категории</h3>
                <div className="cat-grid">
                  {categories.map(c => (
                    <div key={c.id} className="premium-card cat-card-admin">
                      <span>{c.name}</span>
                      <div className="flex gap-2">
                        <button className="edit-btn" onClick={() => openEditCat(c)}><Edit2 size={16} /></button>
                        <button className="delete-btn-ico" onClick={() => deleteCat(c.id)}><Trash2 size={16} /></button>
                      </div>
                    </div>
                  ))}
                </div>
              </section>

              <section className="prod-section">
                <h3>Продукты</h3>
                <div className="admin-table-wrapper premium-card">
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>ID</th>
                        <th>Название</th>
                        <th>Категория</th>
                        <th>Цена</th>
                        <th>Себестоимость</th>
                        <th>Прибыль</th>
                        <th>Статус</th>
                        <th>Действия</th>
                      </tr>
                    </thead>
                    <tbody>
                      {products.map(p => (
                        <tr key={p.id}>
                          <td>{p.id}</td>
                          <td>{p.name}</td>
                          <td>{categories.find(c => c.id === p.category_id)?.name || 'Категория topilmadi'}</td>
                          <td className="font-700">{p.price.toLocaleString()} сум</td>
                          <td style={{ color: 'var(--danger)' }}>{p.cost_price ? p.cost_price.toLocaleString() : 0} сум</td>
                          <td style={{ color: 'var(--success)' }}>{p.profit_margin ? p.profit_margin.toLocaleString() : p.price.toLocaleString()} сум</td>
                          <td>{p.is_active ? 'Активен' : 'Неактивен'}</td>
                          <td>
                            <div className="flex gap-2">
                              <button className="edit-btn" onClick={() => openEditProd(p)}><Edit2 size={16} /></button>
                              {parseInt(p.id) !== parseInt(containerId) ? (
                                <button className="delete-btn" onClick={() => deleteProd(p.id)}>
                                  <Trash2 size={16} />
                                </button>
                              ) : (
                                <span className="locked-badge" title="Системный продукт, удаление невозможно">🔒</span>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </section>
            </div>
          </div>
        )}

        {activeTab === 'staff' && (
          <div className="staff-mgmt">
            <div className="flex-header">
              <h2>Управление сотрудниками</h2>
              <button className="btn-primary" onClick={() => setShowStaffModal(true)}>
                <Plus size={18} /> Добавить сотрудника
              </button>
            </div>
            <div className="premium-card">
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>Имя</th>
                    <th>Телефон</th>
                    <th>Роль</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {staff.map(s => (
                    <tr key={s.id}>
                      <td>{s.full_name}</td>
                      <td>{s.phone}</td>
                      <td><span className={`status-badge ${s.role}`}>{s.role}</span></td>
                      <td>
                        <button className="delete-btn" onClick={() => deleteStaff(s.id)}><Trash2 size={16} /></button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === 'tables' && (
          <div className="tables-mgmt animate-fade">
            <div className="flex-header">
              <h2>Управление столами</h2>
              <button className="btn-primary" onClick={() => { setNewTable({ name: '', capacity: '' }); setEditingTableId(null); setShowTableModal(true); }}>
                <Plus size={18} /> Добавить стол
              </button>
            </div>
            <div className="premium-card">
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>Название стола</th>
                    <th>Кол-во человек (Capacity)</th>
                    <th>Статус</th>
                    <th>Действия</th>
                  </tr>
                </thead>
                <tbody>
                  {tables.map(t => (
                    <tr key={t.id}>
                      <td>{t.name}</td>
                      <td>{t.capacity || 4} чел.</td>
                      <td>
                        <span className={`status-badge ${t.status === 'free' ? 'ready' : 'cancelled'}`}>
                          {t.status === 'free' ? 'Bo\'sh' : 'Занят'}
                        </span>
                      </td>
                      <td>
                        <div style={{ display: 'flex', gap: '8px' }}>
                          <button className="btn-secondary" style={{ padding: '6px' }} onClick={() => openEditTable(t)}><Edit2 size={16} /></button>
                          <button className="delete-btn" onClick={() => deleteTable(t.id)}><Trash2 size={16} /></button>
                          {t.status === 'occupied' && (
                            <button className="btn-primary" style={{ padding: '6px', fontSize: '12px' }} onClick={() => {
                              setTransferFromTable(t);
                              setShowTransferModal(true);
                            }}>🔄 Ko'chirish</button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
              {tables.length === 0 && <div className="text-center p-4">Столы yo'q</div>}
            </div>
          </div>
        )}
      </main>

      {/* Transfer Table Modal */}
      {showTransferModal && transferFromTable && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>Stolni ko'chirish</h3>
              <button onClick={() => { setShowTransferModal(false); setTransferFromTable(null); setTransferToTableId(''); }}><X size={20} /></button>
            </div>
            <form onSubmit={handleTransferTable}>
              <div className="input-group">
                <label>Qaysi stoldan:</label>
                <input type="text" value={transferFromTable.name} disabled />
              </div>
              <div className="input-group">
                <label>Qaysi stolga (Bo'sh stollar):</label>
                <select 
                  value={transferToTableId} 
                  onChange={e => setTransferToTableId(e.target.value)}
                  required
                >
                  <option value="">-- Stol tanlang --</option>
                  {tables.filter(t => t.status === 'free').map(t => (
                    <option key={t.id} value={t.id}>{t.name} ({t.capacity || 4} kishi)</option>
                  ))}
                </select>
              </div>
              <button type="submit" className="btn-primary w-full mt-4">Ko'chirish</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Staff Modal */}
      {showStaffModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>Новый Xodim</h3>
              <button onClick={() => setShowStaffModal(false)}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateStaff}>
              <div className={`input-group ${errors.staff?.name ? 'has-error' : ''}`}>
                <label>ФИО</label>
                <input 
                  value={newStaff.full_name} 
                  onChange={e => {
                    setNewStaff({...newStaff, full_name: e.target.value});
                    if (errors.staff?.name) setErrors({...errors, staff: {...errors.staff, name: null}});
                  }} 
                />
                {errors.staff?.name && <span className="field-error">{errors.staff.name}</span>}
              </div>
              <div className={`input-group ${errors.staff?.phone ? 'has-error' : ''}`}>
                <label>Телефон</label>
                <input 
                  value={newStaff.phone} 
                  onChange={e => {
                    setNewStaff({...newStaff, phone: e.target.value});
                    if (errors.staff?.phone) setErrors({...errors, staff: {...errors.staff, phone: null}});
                  }} 
                />
                {errors.staff?.phone && <span className="field-error">{errors.staff.phone}</span>}
              </div>
              <div className={`input-group ${errors.staff?.password ? 'has-error' : ''}`}>
                <label>Пароль</label>
                <input 
                  type="password" 
                  value={newStaff.password} 
                  onChange={e => {
                    setNewStaff({...newStaff, password: e.target.value});
                    if (errors.staff?.password) setErrors({...errors, staff: {...errors.staff, password: null}});
                  }} 
                />
                {errors.staff?.password && <span className="field-error">{errors.staff.password}</span>}
              </div>
              <div className="input-group">
                <label>Роль</label>
                <select value={newStaff.role} onChange={e => setNewStaff({...newStaff, role: e.target.value})}>
                  <option value="cook">Повар (Cook)</option>
                  <option value="courier">Курьер (Courier)</option>
                  <option value="waiter">Официант (Waiter)</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Сохранить</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Category Modal */}
      {showCatModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>{editCatId ? 'Редактирование категории' : 'Новая категория'}</h3>
              <button onClick={() => { setShowCatModal(false); setEditCatId(null); }}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateCat}>
              <div className={`input-group ${errors.cat ? 'has-error' : ''}`}>
                <label>Название</label>
                <input 
                  value={newCat.name} 
                  onChange={e => {
                    setNewCat({...newCat, name: e.target.value});
                    if (errors.cat) setErrors({...errors, cat: null});
                  }} 
                />
                {errors.cat && <span className="field-error">{errors.cat}</span>}
              </div>
              <div className="input-group">
                <label>Загрузить изображение или URL</label>
                <input 
                  type="file" 
                  accept="image/*" 
                  onChange={e => handleUpload(e, setNewCat)} 
                  style={{ marginBottom: '10px' }}
                />
                <input 
                  value={newCat.image_url} 
                  onChange={e => setNewCat({...newCat, image_url: e.target.value})} 
                  placeholder="Или введите прямой URL"
                />
                {newCat.image_url && (
                  <div style={{marginTop: 10}}>
                    <img src={newCat.image_url.startsWith('/') ? `${api.defaults.baseURL.replace('/api', '')}${newCat.image_url}` : newCat.image_url} style={{height: 60, borderRadius: 8}} alt="Preview" />
                  </div>
                )}
              </div>
              
              <div className="input-group flex-row gap-2 mt-2 mb-4">
                <input 
                  type="checkbox" 
                  id="cat_user_controlled"
                  checked={newCat.is_user_controlled} 
                  onChange={e => setNewCat({...newCat, is_user_controlled: e.target.checked})} 
                />
                <label htmlFor="cat_user_controlled" style={{ cursor: 'pointer' }}>Пользовательский выбор порций (Порция/Шт)</label>
              </div>

              <div className="input-group mb-4">
                <label>Маршрутизация принтера (Откуда выйдет чек?)</label>
                <select 
                  style={{ padding: '0.8rem', borderRadius: '12px', border: '1px solid var(--border)', background: 'var(--bg-surface)', color: 'var(--text-primary)', width: '100%', outline: 'none' }}
                  value={newCat.printer_target}
                  onChange={e => setNewCat({...newCat, printer_target: e.target.value})}
                >
                  <option value="ALL">Барча принтерлардан (По умолчанию)</option>
                  <option value="USB">Kassa printeri (USB)</option>
                  <option value="192.168.1.11:9100">Kuxnya (LAN 192.168.1.11)</option>
                  <option value="192.168.1.10:9100">Salat (LAN 192.168.1.10)</option>
                </select>
              </div>

              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Сохранить</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Product Modal */}
      {showProdModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>{editProdId ? 'Редактирование продукта' : 'Новый продукт'}</h3>
              <button onClick={() => { setShowProdModal(false); setEditProdId(null); }}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateProd}>
              <div className={`input-group ${errors.prod?.name ? 'has-error' : ''}`}>
                <label>Название</label>
                <input 
                  value={newProd.name} 
                  onChange={e => {
                    setNewProd({...newProd, name: e.target.value});
                    if (errors.prod?.name) setErrors({...errors, prod: {...errors.prod, name: null}});
                  }} 
                />
                {errors.prod?.name && <span className="field-error">{errors.prod.name}</span>}
              </div>
              <div className="input-group">
                <label>Описание</label>
                <input value={newProd.description} onChange={e => setNewProd({...newProd, description: e.target.value})} />
              </div>
              <div className="input-group">
                <label>Загрузить изображение или URL</label>
                <input 
                  type="file" 
                  accept="image/*" 
                  onChange={e => handleUpload(e, setNewProd)} 
                  style={{ marginBottom: '10px' }}
                />
                <input 
                  value={newProd.image_url} 
                  onChange={e => setNewProd({...newProd, image_url: e.target.value})} 
                  placeholder="Или введите прямой URL"
                />
                {newProd.image_url && (
                  <div style={{marginTop: 10}}>
                    <img src={newProd.image_url.startsWith('/') ? `${api.defaults.baseURL.replace('/api', '')}${newProd.image_url}` : newProd.image_url} style={{height: 60, borderRadius: 8}} alt="Preview" />
                  </div>
                )}
              </div>
              <div className={`input-group ${errors.prod?.price ? 'has-error' : ''}`}>
                <label>Цена</label>
                <input 
                  type="number" 
                  value={newProd.price} 
                  onChange={e => {
                    setNewProd({...newProd, price: e.target.value});
                    if (errors.prod?.price) setErrors({...errors, prod: {...errors.prod, price: null}});
                  }} 
                />
                {errors.prod?.price && <span className="field-error">{errors.prod.price}</span>}
              </div>
              <div className={`input-group ${errors.prod?.category ? 'has-error' : ''}`}>
                <label>Категория</label>
                <select 
                  value={newProd.category_id} 
                  onChange={e => {
                    setNewProd({...newProd, category_id: e.target.value});
                    if (errors.prod?.category) setErrors({...errors, prod: {...errors.prod, category: null}});
                  }}
                >
                  <option value="">Выберите...</option>
                  {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
                {errors.prod?.category && <span className="field-error">{errors.prod.category}</span>}
              </div>

              <div className="admin-form-row">
                <div className="input-group">
                  <label>Единица измерения</label>
                  <select value={newProd.unit} onChange={e => setNewProd({...newProd, unit: e.target.value})}>
                    <option value="шт">шт</option>
                    <option value="порц">порц</option>
                    <option value="kg">kg</option>
                    <option value="gr">gr</option>
                  </select>
                </div>
                <div className="input-group">
                  <label>Мин. количество</label>
                  <input type="number" step="0.1" value={newProd.min_quantity} onChange={e => setNewProd({...newProd, min_quantity: e.target.value})} />
                </div>
              </div>

              <div className="admin-form-row">
                <div className="input-group">
                  <label>Шаг (Step)</label>
                  <input type="number" step="0.1" value={newProd.quantity_step} onChange={e => setNewProd({...newProd, quantity_step: e.target.value})} />
                </div>
                <div className="input-group flex-row gap-2 mt-4">
                  <input 
                    type="checkbox" 
                    id="mandatory_container"
                    checked={newProd.has_mandatory_container} 
                    onChange={e => setNewProd({...newProd, has_mandatory_container: e.target.checked})} 
                  />
                  <label htmlFor="mandatory_container" style={{ cursor: 'pointer' }}>Обязательное добавление посуды</label>
                </div>
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Сохранить</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Table Modal */}
      {showTableModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content" style={{maxWidth: '400px'}}>
            <div className="modal-header">
              <h3>{editingTableId ? 'Редактировать стол' : 'Новый стол'}</h3>
              <button onClick={() => setShowTableModal(false)}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateTable}>
              <div className="input-group mb-4">
                <label>Название стола</label>
                <input 
                  type="text"
                  required
                  value={newTable.name} 
                  onChange={e => setNewTable({...newTable, name: e.target.value})} 
                />
              </div>
              <div className="input-group mb-4">
                <label>Количество человек (Необязательно)</label>
                <input 
                  type="number"
                  value={newTable.capacity} 
                  onChange={e => setNewTable({...newTable, capacity: e.target.value})} 
                />
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Сохранить</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Order Details Modal */}
      {showOrderModal && selectedOrderDetails && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content" style={{maxWidth: '600px', background: '#ffffff'}}>
            <div className="modal-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                <h3>Детали заказа #{selectedOrderDetails.id}</h3>
                <button className="btn-secondary" style={{ padding: '0.4rem 0.8rem', fontSize: '0.9rem' }} onClick={() => handleReprintOrder(selectedOrderDetails.id)}>
                  <Printer size={16} /> Распечатать чек
                </button>
              </div>
              <button onClick={() => setShowOrderModal(false)}><X size={20} /></button>
            </div>
            <div className="order-details-body" style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginTop: '1rem' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span className="text-muted">Статус:</span>
                <span className={`status-badge ${selectedOrderDetails.status}`}>{STATUS_MAP[selectedOrderDetails.status] || selectedOrderDetails.status}</span>
              </div>
              {selectedOrderDetails.payment_method && (
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <span className="text-muted">Тип оплаты:</span>
                  <span className="font-700">
                    {selectedOrderDetails.payment_method === 'cash' ? '💵 Наличные' :
                     selectedOrderDetails.payment_method === 'card' ? '💳 Терминал (Карта)' :
                     selectedOrderDetails.payment_method === 'click' ? '📱 Click/Payme' :
                     selectedOrderDetails.payment_method === 'nasiya' ? '📒 В долг' : selectedOrderDetails.payment_method}
                  </span>
                </div>
              )}
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span className="text-muted">Клиент:</span>
                <span className="font-700">{selectedOrderDetails.phone}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span className="text-muted">Адрес / Стол:</span>
                <span className="font-700">{selectedOrderDetails.table_name ? `Стол: ${selectedOrderDetails.table_name}` : selectedOrderDetails.address || '-'}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                <span className="text-muted">Официант:</span>
                <span className="font-700">{selectedOrderDetails.waiter_name || '-'}</span>
              </div>
              {selectedOrderDetails.comment && (
                <div style={{ background: 'var(--bg-surface)', padding: '0.75rem', borderRadius: '8px' }}>
                  <span className="text-muted">Комментарий:</span>
                  <p style={{ marginTop: '0.25rem' }}>{selectedOrderDetails.comment}</p>
                </div>
              )}
              
              <div style={{ marginTop: '1rem', borderTop: '1px solid var(--border)', paddingTop: '1rem' }}>
                <h4 style={{ marginBottom: '0.75rem' }}>Список продуктов</h4>
                <div style={{ maxHeight: '250px', overflowY: 'auto', paddingRight: '0.5rem' }}>
                  {(() => {
                    // Group items by product_id, summing quantities
                    const grouped = (selectedOrderDetails.items || []).reduce((acc, item) => {
                      const key = item.product_id;
                      if (acc[key]) {
                        acc[key].quantity += item.quantity;
                        acc[key].totalPrice += (item.price || 0) * item.quantity;
                        acc[key]._origItems.push(item);
                      } else {
                        acc[key] = { ...item, totalPrice: (item.price || 0) * item.quantity, _origItems: [item] };
                      }
                      return acc;
                    }, {});
                    return Object.values(grouped).map((item) => (
                      <div key={item.product_id} style={{ display: 'flex', justifyContent: 'space-between', padding: '0.5rem 0', borderBottom: '1px solid var(--border)' }}>
                        <div>
                          <div className="font-700">{item.product_name || "Noma'lum"}</div>
                          <div className="text-muted" style={{ fontSize: '0.85rem' }}>{item.quantity} {item.unit || 'x'}</div>
                        </div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
                          <div className="font-700 text-primary">{item.totalPrice.toLocaleString()} сум</div>
                          <button 
                            className="delete-btn-ico" 
                            style={{ padding: '0.2rem', color: '#ef4444', background: 'transparent', border: 'none', cursor: 'pointer' }}
                            onClick={() => handleRemoveOrderItem(selectedOrderDetails.id, item._origItems[item._origItems.length - 1])}
                            title="Отменить"
                          >
                            <Trash2 size={16} />
                          </button>
                        </div>
                      </div>
                    ));
                  })()}
                </div>
              </div>
              
              {selectedOrderDetails.service_fee > 0 && (
                <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '0.5rem', color: '#f59e0b' }}>
                  <span className="text-muted">Плата за обслуживание ({selectedOrderDetails.service_percentage}%):</span>
                  <span className="font-700">{selectedOrderDetails.service_fee?.toLocaleString()} сум</span>
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: '1rem', paddingTop: '1rem', borderTop: '1px solid var(--border)' }}>
                <span className="font-700" style={{ fontSize: '1.2rem' }}>Итого:</span>
                <span className="font-800 text-primary" style={{ fontSize: '1.4rem' }}>{selectedOrderDetails.total_price?.toLocaleString()} сум</span>
              </div>

              {/* Service Fee Controls */}
              {(selectedOrderDetails.table_id || selectedOrderDetails.delivery_method === 'dine_in') && (
                <div style={{ marginTop: '1.5rem', paddingTop: '1rem', borderTop: '1px solid var(--border)' }}>
                  <h4 style={{ marginBottom: '0.75rem', color: '#f59e0b' }}>Процент обслуживания (%)</h4>
                  <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                    <input
                      type="number"
                      min="0"
                      max="100"
                      step="1"
                      value={serviceFeePercent}
                      onChange={e => setServiceFeePercent(e.target.value)}
                      style={{ width: '80px', padding: '0.5rem', borderRadius: '8px', border: '1px solid var(--border)', background: 'var(--bg-surface)', color: 'var(--text-primary)' }}
                    />
                    <span className="text-muted">%</span>
                    <button className="btn-secondary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }} onClick={() => handleSetServiceFee(selectedOrderDetails.id)}>
                      Применить
                    </button>
                    <button className="btn-primary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }} onClick={() => handleSetServiceFeeAndPrint(selectedOrderDetails.id)}>
                      <Printer size={14} /> Сохранить и распечатать чек
                    </button>
                  </div>
                </div>
              )}
            </div>
          </motion.div>
        </div>
      )}

      <style>{`
        .admin-page {
          display: flex;
          min-height: calc(100vh - 100px);
          gap: 2rem;
        }

        .admin-sidebar {
          width: 280px;
          border-radius: 24px;
          padding: 1.5rem;
          display: flex;
          flex-direction: column;
          gap: 2rem;
          height: fit-content;
          position: sticky;
          top: 90px;
        }

        .sidebar-header {
          display: flex;
          align-items: center;
          gap: 1rem;
          font-weight: 700;
          font-size: 1.2rem;
        }

        .sidebar-nav {
          display: flex;
          flex-direction: column;
          gap: 0.5rem;
        }

        .sidebar-nav button {
          display: flex;
          align-items: center;
          gap: 1rem;
          padding: 0.75rem 1rem;
          background: none;
          color: var(--text-dim);
          text-align: left;
          width: 100%;
        }

        .sidebar-nav button.active {
          background: var(--bg-surface);
          color: var(--primary);
          border-right: 4px solid var(--primary);
        }

        .admin-main {
          padding-top: 0.5rem;
          flex: 1;
        }
        
        .admin-form-row {
          display: grid;
          grid-template-columns: 1fr 1fr;
          gap: 1rem;
        }
        
        .flex-row {
          display: flex;
          align-items: center;
        }

        .gap-2 { gap: 0.5rem; }
        .mt-4 { margin-top: 1rem; }

        .flex-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 2rem;
        }

        .actions {
          display: flex;
          gap: 1rem;
        }

        .admin-table {
          width: 100%;
          border-collapse: collapse;
          text-align: left;
        }

        .admin-table th, .admin-table td {
          padding: 1rem;
          border-bottom: 1px solid var(--border);
        }

        .admin-table th {
          color: var(--text-dim);
          font-weight: 600;
          font-size: 0.9rem;
        }

        .status-badge {
          padding: 4px 10px;
          border-radius: 20px;
          font-size: 0.8rem;
          font-weight: 700;
          text-transform: capitalize;
        }

        .status-badge.new { background: #3b82f6; }
        .status-badge.delivered { background: #10b981; }

        .cat-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
          gap: 1rem;
          margin-top: 1rem;
        }

        .cat-card-admin {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 1rem;
        }

        .delete-btn-ico { color: #ef4444; background: none; }
        .delete-btn { color: #ef4444; background: rgba(239, 68, 68, 0.1); padding: 6px; }
        .edit-btn { color: var(--primary); background: rgba(99, 102, 241, 0.1); padding: 6px; }

        .modal-overlay {
          position: fixed;
          top: 0; left: 0; right: 0; bottom: 0;
          background: rgba(0,0,0,0.8);
          backdrop-filter: blur(4px);
          display: flex;
          justify-content: center;
          align-items: center;
          z-index: 2000;
        }

        .modal-content {
          width: 90%;
          max-width: 500px;
        }

        .modal-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 1.5rem;
        }

        .w-full { width: 100%; }
        .mb-4 { margin-bottom: 2rem; }
        select {
          width: 100%;
          background: var(--bg-input);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 0.75rem;
          color: var(--text-primary);
          outline: none;
        }

        .field-error {
          color: #ef4444;
          font-size: 0.75rem;
          margin-top: 0.25rem;
          display: block;
        }

        /* ── PERFORMANCE GRID ── */
        .performance-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
          gap: 1.25rem;
          margin-top: 1rem;
        }

        .perf-card { padding: 1.5rem; }

        .perf-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 1.25rem;
        }

        .perf-user {
          display: flex;
          align-items: center;
          gap: 0.85rem;
        }

        .perf-user h3 { font-size: 1rem; font-weight: 700; margin-bottom: 0.2rem; }

        .role-chip {
          display: inline-block;
          font-size: 0.7rem;
          font-weight: 700;
          text-transform: uppercase;
          letter-spacing: 0.06em;
          padding: 2px 8px;
          background: rgba(249,115,22,0.12);
          color: var(--primary);
          border-radius: 6px;
        }

        .perf-score {
          display: flex;
          align-items: center;
          gap: 0.35rem;
          font-size: 1.5rem;
          font-weight: 800;
          color: #fbbf24;
        }

        .perf-stats {
          display: flex;
          gap: 0.75rem;
        }

        .stat-item {
          flex: 1;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          border-radius: 10px;
          padding: 0.65rem 0.75rem;
          text-align: center;
        }

        .stat-item label {
          display: block;
          font-size: 0.7rem;
          color: var(--text-muted);
          margin-bottom: 0.3rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 0.04em;
        }

        .stat-item span {
          font-size: 1.2rem;
          font-weight: 800;
          color: var(--text-primary);
        }

        .stat-item.positive span { color: #34d399; }
        .stat-item.negative span { color: #f87171; }

        .empty-state {
          text-align: center;
          padding: 3rem;
          color: var(--text-muted);
        }

        .refresh-btn {
          display: flex;
          align-items: center;
          gap: 0.5rem;
          background: var(--bg-surface);
          border: 1px solid var(--border);
          color: var(--text-secondary);
          padding: 0.55rem 1rem;
          border-radius: var(--radius-sm);
          font-size: 0.85rem;
          font-weight: 600;
          transition: var(--transition);
        }

        .refresh-btn:hover { border-color: var(--primary); color: var(--primary); }

        /* ── RESPONSIVE ── */
        @media (max-width: 900px) {
          .admin-page { flex-direction: column; gap: 1rem; }
          .admin-sidebar {
            width: 100%;
            position: static;
            padding: 1rem;
            gap: 0.5rem;
            border-radius: var(--radius);
          }
          .sidebar-header { display: none; }
          .sidebar-nav { flex-direction: row; gap: 0.35rem; overflow-x: auto; }
          .sidebar-nav button {
            border-right: none !important;
            border-bottom: 2px solid transparent;
            padding: 0.55rem 0.9rem;
            font-size: 0.8rem;
            flex-shrink: 0;
            border-radius: 8px;
          }
          .sidebar-nav button.active {
            border-bottom-color: var(--primary);
          }
          .admin-main { flex: 1; }
          .flex-header { flex-wrap: wrap; gap: 0.75rem; }
          .actions { flex-wrap: wrap; gap: 0.5rem; }
          .orders-table-wrapper, .admin-table-wrapper { overflow-x: auto; -webkit-overflow-scrolling: touch; }
          .admin-table { min-width: 700px; }
          .performance-grid { grid-template-columns: 1fr !important; }
          .admin-form-row { grid-template-columns: 1fr; }
        }

        @media (max-width: 600px) {
          .modal-content {
            width: 95%;
            max-width: 100%;
            margin: 1rem;
            border-radius: var(--radius);
          }
          .sidebar-nav button { font-size: 0.75rem; padding: 0.45rem 0.7rem; }
        }
      `}</style>
    </div>
  );
};

export default Admin;


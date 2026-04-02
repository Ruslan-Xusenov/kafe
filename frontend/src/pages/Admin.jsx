import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import StatsSection from '../components/StatsSection';
import { validateNotEmpty, validatePrice, validatePhone, validatePassword } from '../utils/validation';
import { motion, AnimatePresence } from 'framer-motion';
import { 
  LayoutDashboard, ShoppingBag, Users, Plus, Edit2, Trash2, 
  CheckCircle, XCircle, Clock, Loader2, Save, X, ChefHat, Truck, Star, RefreshCw
} from 'lucide-react';

const Admin = () => {
  const [activeTab, setActiveTab] = useState('orders');
  const [orders, setOrders] = useState([]);
  const [categories, setCategories] = useState([]);
  const [products, setProducts] = useState([]);
  const [staff, setStaff] = useState([]);
  const [performance, setPerformance] = useState([]);
  const [loading, setLoading] = useState(true);

  // States for Modals/Forms
  const [showCatModal, setShowCatModal] = useState(false);
  const [newCat, setNewCat] = useState({ name: '', image_url: '' });
  const [showProdModal, setShowProdModal] = useState(false);
  const [newProd, setNewProd] = useState({ name: '', description: '', price: '', category_id: '', image_url: '' });
  const [editProdId, setEditProdId] = useState(null);
  const [showStaffModal, setShowStaffModal] = useState(false);
  const [newStaff, setNewStaff] = useState({ full_name: '', phone: '', password: '', role: 'cook' });
  const [showOrderModal, setShowOrderModal] = useState(false);
  const [selectedOrderDetails, setSelectedOrderDetails] = useState(null);
  
  const [errors, setErrors] = useState({});

  useEffect(() => {
    fetchData();
  }, [activeTab]);

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
        const res = await api.get('/performance');
        setPerformance(res.data || []);
      }
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateCat = async (e) => {
    e.preventDefault();
    const err = validateNotEmpty(newCat.name, 'Nomi');
    if (err) { setErrors({ cat: err }); return; }

    try {
      await api.post('/catalog/categories', newCat);
      setShowCatModal(false);
      setNewCat({ name: '', image_url: '' });
      setErrors({});
      fetchData();
    } catch (err) { alert('Xatolik'); }
  };

  const handleCreateProd = async (e) => {
    e.preventDefault();
    const prodErrors = {
      name: validateNotEmpty(newProd.name, 'Nomi'),
      price: validatePrice(newProd.price),
      category: validateNotEmpty(newProd.category_id, 'Kategoriya')
    };
    if (prodErrors.name || prodErrors.price || prodErrors.category) {
      setErrors({ prod: prodErrors });
      return;
    }

    try {
      if (editProdId) {
        await api.put(`/catalog/products/${editProdId}`, { ...newProd, price: parseFloat(newProd.price), category_id: parseInt(newProd.category_id) });
      } else {
        await api.post('/catalog/products', { ...newProd, price: parseFloat(newProd.price), category_id: parseInt(newProd.category_id) });
      }
      setShowProdModal(false);
      setNewProd({ name: '', description: '', price: '', category_id: '', image_url: '' });
      setEditProdId(null);
      setErrors({});
      fetchData();
    } catch (err) { alert('Xatolik'); }
  };

  const openEditProd = (p) => {
    setEditProdId(p.id);
    setNewProd({
      name: p.name,
      description: p.description || '',
      price: p.price,
      category_id: p.category_id,
      image_url: p.image_url || ''
    });
    setShowProdModal(true);
  };

  const handleCreateStaff = async (e) => {
    e.preventDefault();
    const staffErrors = {
      name: validateNotEmpty(newStaff.full_name, 'Ism'),
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
      alert(err.response?.data?.error || 'Xodim qo\'shishda xatolik');
    }
  };

  const handleUpload = async (e, setter, stateRef) => {
    const file = e.target.files[0];
    if (!file) return;
    const formData = new FormData();
    formData.append('file', file);
    try {
      setLoading(true);
      const res = await api.post('/catalog/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' }
      });
      setter({ ...stateRef, image_url: res.data.url });
    } catch (err) {
      alert('Rasm yuklashda xatolik');
    } finally {
      setLoading(false);
    }
  };

  const deleteProd = async (id) => {
    if (window.confirm('Olib tashlashga ishonchingiz komilmi?')) {
      try {
        await api.delete(`/catalog/products/${id}`);
        fetchData();
      } catch (err) { alert('Xatolik'); }
    }
  };

  if (loading) return <div className="flex-center h-full"><Loader2 className="animate-spin" /></div>;

  return (
    <div className="admin-page animate-fade">
      <aside className="admin-sidebar glass">
        <div className="sidebar-header">
          <LayoutDashboard className="text-primary" />
          <span>Admin Boshqaruvi</span>
        </div>
        <nav className="sidebar-nav">
          <button className={activeTab === 'orders' ? 'active' : ''} onClick={() => setActiveTab('orders')}>
            <ShoppingBag size={20} /> Buyurtmalar
          </button>
          <button className={activeTab === 'menu' ? 'active' : ''} onClick={() => setActiveTab('menu')}>
            <Edit2 size={20} /> Menyu (CRUD)
          </button>
          <button className={activeTab === 'staff' ? 'active' : ''} onClick={() => setActiveTab('staff')}>
            <Users size={20} /> Xodimlar
          </button>
          <button className={activeTab === 'performance' ? 'active' : ''} onClick={() => setActiveTab('performance')}>
            <Star size={20} /> Reytinglar
          </button>
        </nav>
      </aside>

      <main className="admin-main">
        {activeTab === 'orders' && (
          <div className="orders-mgmt">
            <StatsSection role="admin" />
            <div className="flex justify-between items-center mb-4">
              <h2>Buyurtmalar Monitoringi</h2>
              <button className="refresh-btn" onClick={fetchData}><Clock size={18} /> Yangilash</button>
            </div>
            <div className="orders-table-wrapper premium-card">
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>ID</th>
                    <th>Mijoz</th>
                    <th>Jami</th>
                    <th>Holat</th>
                    <th>Sana</th>
                    <th>Amallar</th>
                  </tr>
                </thead>
                <tbody>
                  {orders.map(o => (
                    <tr key={o.id}>
                      <td>#{o.id}</td>
                      <td>{o.phone}</td>
                      <td>{o.total_price.toLocaleString()}</td>
                      <td><span className={`status-badge ${o.status}`}>{o.status}</span></td>
                      <td>{new Date(o.created_at).toLocaleDateString()}</td>
                      <td>
                        <button className="action-btn-small" onClick={() => { setSelectedOrderDetails(o); setShowOrderModal(true); }}>Ko'rish</button>
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
                <h2>Xodimlar Reytingi va Ish Sifati</h2>
                <button className="refresh-btn" onClick={fetchData}><RefreshCw size={18} /> Yangilash</button>
             </div>
             <div className="performance-grid">
               {performance.map(p => (
                 <div key={p.staff_id} className="premium-card perf-card">
                    <div className="perf-header">
                       <div className="perf-user">
                          <ChefHat className={p.role === 'cook' ? 'text-primary' : 'text-secondary'} size={32} />
                          <div>
                             <h3>{p.full_name}</h3>
                             <span className="role-chip">{p.role === 'cook' ? 'Oshpaz' : 'Kuryer'}</span>
                          </div>
                       </div>
                       <div className="perf-score">
                          <Star size={20} fill="#f1c40f" className="text-yellow-400" />
                          <span>{p.avg_rating.toFixed(1)}</span>
                       </div>
                    </div>
                    
                    <div className="perf-stats">
                       <div className="stat-item">
                          <label>Jami baholar</label>
                          <span>{p.total_reviews}</span>
                       </div>
                       <div className="stat-item positive">
                          <label>Yaxshi (4-5)</label>
                          <span>{p.good_reviews}</span>
                       </div>
                       <div className="stat-item negative">
                          <label>Yomon (1-2)</label>
                          <span>{p.bad_reviews}</span>
                       </div>
                    </div>
                 </div>
               ))}
             </div>
             
             {performance.length === 0 && (
               <div className="empty-state">
                 <p>Hali birorta ham xodimga baho berilmagan.</p>
               </div>
             )}
          </div>
        )}

        {activeTab === 'menu' && (
          <div className="menu-mgmt">
            <div className="flex-header">
              <h2>Menyu Boshqaruvi</h2>
              <div className="actions">
                <button className="btn-primary" onClick={() => { setEditProdId(null); setNewProd({ name: '', description: '', price: '', category_id: '', image_url: '' }); setShowCatModal(true); }}><Plus size={18} /> Kategoriya</button>
                <button className="btn-primary" onClick={() => { setEditProdId(null); setNewProd({ name: '', description: '', price: '', category_id: '', image_url: '' }); setShowProdModal(true); }}><Plus size={18} /> Mahsulot</button>
              </div>
            </div>

            <div className="menu-sections">
              <section className="cat-section mb-4">
                <h3>Kategoriyalar</h3>
                <div className="cat-grid">
                  {categories.map(c => (
                    <div key={c.id} className="premium-card cat-card-admin">
                      <span>{c.name}</span>
                      <button className="delete-btn-ico"><Trash2 size={16} /></button>
                    </div>
                  ))}
                </div>
              </section>

              <section className="prod-section">
                <h3>Mahsulotlar</h3>
                <div className="admin-table-wrapper premium-card">
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Nomi</th>
                        <th>Kategoriya</th>
                        <th>Narxi</th>
                        <th>Holat</th>
                        <th>Amallar</th>
                      </tr>
                    </thead>
                    <tbody>
                      {products.map(p => (
                        <tr key={p.id}>
                          <td>{p.name}</td>
                          <td>{categories.find(c => c.id === p.category_id)?.name || 'Kategoriya topilmadi'}</td>
                          <td>{p.price.toLocaleString()} so'm</td>
                          <td>{p.is_active ? 'Faol' : 'Nofaol'}</td>
                          <td>
                            <div className="flex gap-2">
                              <button className="edit-btn" onClick={() => openEditProd(p)}><Edit2 size={16} /></button>
                              <button className="delete-btn" onClick={() => deleteProd(p.id)}><Trash2 size={16} /></button>
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
              <h2>Xodimlar Boshqaruvi</h2>
              <button className="btn-primary" onClick={() => setShowStaffModal(true)}>
                <Plus size={18} /> Xodim qo'shish
              </button>
            </div>
            <div className="premium-card">
              <table className="admin-table">
                <thead>
                  <tr>
                    <th>Ism</th>
                    <th>Telefon</th>
                    <th>Vazifa (Role)</th>
                    <th>Amallar</th>
                  </tr>
                </thead>
                <tbody>
                  {staff.map(s => (
                    <tr key={s.id}>
                      <td>{s.full_name}</td>
                      <td>{s.phone}</td>
                      <td><span className={`status-badge ${s.role}`}>{s.role}</span></td>
                      <td>
                        <button className="delete-btn"><Trash2 size={16} /></button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </main>

      {/* Staff Modal */}
      {showStaffModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>Yangi Xodim</h3>
              <button onClick={() => setShowStaffModal(false)}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateStaff}>
              <div className={`input-group ${errors.staff?.name ? 'has-error' : ''}`}>
                <label>Ismi Sharif</label>
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
                <label>Telefon</label>
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
                <label>Parol</label>
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
                <label>Vazifasi</label>
                <select value={newStaff.role} onChange={e => setNewStaff({...newStaff, role: e.target.value})}>
                  <option value="cook">Oshpaz (Cook)</option>
                  <option value="courier">Kuryer (Courier)</option>
                  <option value="admin">Admin</option>
                </select>
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Saqlash</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Category Modal */}
      {showCatModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>Yangi Kategoriya</h3>
              <button onClick={() => setShowCatModal(false)}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateCat}>
              <div className={`input-group ${errors.cat ? 'has-error' : ''}`}>
                <label>Nomi</label>
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
                <label>Rasm yuklash yoki URL</label>
                <input 
                  type="file" 
                  accept="image/*" 
                  onChange={e => handleUpload(e, setNewCat, newCat)} 
                  style={{ marginBottom: '10px' }}
                />
                <input 
                  value={newCat.image_url} 
                  onChange={e => setNewCat({...newCat, image_url: e.target.value})} 
                  placeholder="Yoki to'g'ridan-to'g'ri URL kiriting"
                />
                {newCat.image_url && (
                  <div style={{marginTop: 10}}>
                    <img src={newCat.image_url.startsWith('/') ? `http://localhost:8080${newCat.image_url}` : newCat.image_url} style={{height: 60, borderRadius: 8}} alt="Preview" />
                  </div>
                )}
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Saqlash</button>
            </form>
          </motion.div>
        </div>
      )}

      {/* Product Modal */}
      {showProdModal && (
        <div className="modal-overlay">
          <motion.div initial={{ scale: 0.9, opacity: 0 }} animate={{ scale: 1, opacity: 1 }} className="premium-card modal-content">
            <div className="modal-header">
              <h3>{editProdId ? 'Mahsulotni tahrirlash' : 'Yangi Mahsulot'}</h3>
              <button onClick={() => { setShowProdModal(false); setEditProdId(null); }}><X size={20} /></button>
            </div>
            <form onSubmit={handleCreateProd}>
              <div className={`input-group ${errors.prod?.name ? 'has-error' : ''}`}>
                <label>Nomi</label>
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
                <label>Tavsif</label>
                <input value={newProd.description} onChange={e => setNewProd({...newProd, description: e.target.value})} />
              </div>
              <div className="input-group">
                <label>Rasm yuklash yoki URL</label>
                <input 
                  type="file" 
                  accept="image/*" 
                  onChange={e => handleUpload(e, setNewProd, newProd)} 
                  style={{ marginBottom: '10px' }}
                />
                <input 
                  value={newProd.image_url} 
                  onChange={e => setNewProd({...newProd, image_url: e.target.value})} 
                  placeholder="Yoki to'g'ridan-to'g'ri URL kiriting"
                />
                {newProd.image_url && (
                  <div style={{marginTop: 10}}>
                    <img src={newProd.image_url.startsWith('/') ? `http://localhost:8080${newProd.image_url}` : newProd.image_url} style={{height: 60, borderRadius: 8}} alt="Preview" />
                  </div>
                )}
              </div>
              <div className={`input-group ${errors.prod?.price ? 'has-error' : ''}`}>
                <label>Narxi</label>
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
                <label>Kategoriya</label>
                <select 
                  value={newProd.category_id} 
                  onChange={e => {
                    setNewProd({...newProd, category_id: e.target.value});
                    if (errors.prod?.category) setErrors({...errors, prod: {...errors.prod, category: null}});
                  }}
                >
                  <option value="">Tanlang...</option>
                  {categories.map(c => <option key={c.id} value={c.id}>{c.name}</option>)}
                </select>
                {errors.prod?.category && <span className="field-error">{errors.prod.category}</span>}
              </div>
              <button type="submit" className="btn-primary w-full mt-2"><Save size={18} /> Saqlash</button>
            </form>
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
        }

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
          background: var(--bg-dark);
          border: 1px solid var(--border);
          border-radius: 8px;
          padding: 0.75rem;
          color: white;
          outline: none;
        }

        .field-error {
          color: #ef4444;
          font-size: 0.75rem;
          margin-top: 0.25rem;
          display: block;
        }

        .input-group.has-error input, .input-group.has-error select {
          border-color: #ef4444;
        }
      `}</style>
    </div>
  );
};

export default Admin;

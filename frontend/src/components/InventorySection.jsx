import toast from 'react-hot-toast';
import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Loader2, Plus, AlertTriangle, Trash2, Link, Edit2 } from 'lucide-react';

const InventorySection = ({ products }) => {
  const [ingredients, setIngredients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newIngredient, setNewIngredient] = useState({ name: '', stock: '', unit: 'gr', min_stock: '', cost_price: '' });
  const [restockAmounts, setRestockAmounts] = useState({}); // {ingredientId: amount}
  const [editingIngredientId, setEditingIngredientId] = useState(null);
  const [editFormData, setEditFormData] = useState({ name: '', stock: '', unit: 'gr', min_stock: '', cost_price: '' });
  
  // Recipe linking state
  const [selectedProduct, setSelectedProduct] = useState('');
  const [recipes, setRecipes] = useState([]);
  const [newRecipe, setNewRecipe] = useState({ ingredient_id: '', quantity: '', unit: 'gr' });

  useEffect(() => {
    fetchIngredients();
  }, []);

  useEffect(() => {
    if (selectedProduct) {
      fetchRecipes(selectedProduct);
    } else {
      setRecipes([]);
    }
  }, [selectedProduct]);

  const fetchIngredients = async () => {
    try {
      setLoading(true);
      const res = await api.get('/inventory/ingredients');
      setIngredients(res.data || []);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const fetchRecipes = async (productId) => {
    try {
      const res = await api.get(`/inventory/recipes/${productId}`);
      setRecipes(res.data || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateIngredient = async (e) => {
    e.preventDefault();
    if (!newIngredient.name || !newIngredient.stock) return toast.success('Заполните все поля');
    try {
      await api.post('/inventory/ingredients', {
        name: newIngredient.name,
        stock: parseFloat(newIngredient.stock),
        unit: newIngredient.unit,
        min_stock: parseFloat(newIngredient.min_stock || 0),
        cost_price: parseFloat(newIngredient.cost_price || 0)
      });
      setNewIngredient({ name: '', stock: '', unit: 'gr', min_stock: '', cost_price: '' });
      fetchIngredients();
    } catch (err) {
      toast.error('Ошибка: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteIngredient = async (id) => {
    if (!window.confirm("Вы действительно хотите удалить?")) return;
    try {
      await api.delete(`/inventory/ingredients/${id}`);
      fetchIngredients();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      toast.error('Ошибка при удалении');
    }
  };

  const handleAddRecipe = async (e) => {
    e.preventDefault();
    if (!selectedProduct) return toast.success('Сначала выберите продукт');
    if (!newRecipe.ingredient_id || !newRecipe.quantity) return toast.success('Выберите ингредиент и количество');
    
    try {
      await api.post('/inventory/recipes', {
        product_id: parseInt(selectedProduct),
        ingredient_id: parseInt(newRecipe.ingredient_id),
        quantity: parseFloat(newRecipe.quantity),
        unit: newRecipe.unit
      });
      setNewRecipe({ ingredient_id: '', quantity: '', unit: 'gr' });
      fetchRecipes(selectedProduct);
    } catch (err) {
      toast.error('Ошибка: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteRecipe = async (id) => {
    if (!window.confirm("Удалить?")) return;
    try {
      await api.delete(`/inventory/recipes/${id}`);
      fetchRecipes(selectedProduct);
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      toast.error('Ошибка');
    }
  };

  const handleRestock = async (ingId) => {
    const amount = parseFloat(restockAmounts[ingId] || 0);
    if (!amount || amount <= 0) return toast.success("Введите количество");
    try {
      await api.post(`/inventory/ingredients/${ingId}/restock`, { amount });
      setRestockAmounts(prev => ({ ...prev, [ingId]: '' }));
      fetchIngredients();
    } catch (err) {
      toast.error('Ошибка при пополнении запаса: ' + (err.response?.data?.error || err.message));
    }
  };

  const handleEditClick = (ing) => {
    setEditingIngredientId(ing.id);
    setEditFormData({
      name: ing.name,
      stock: ing.stock,
      unit: ing.unit,
      min_stock: ing.min_stock || 0,
      cost_price: ing.cost_price || 0
    });
  };

  const handleSaveEdit = async (id) => {
    try {
      await api.put(`/inventory/ingredients/${id}`, {
        name: editFormData.name,
        stock: parseFloat(editFormData.stock),
        unit: editFormData.unit,
        min_stock: parseFloat(editFormData.min_stock || 0),
        cost_price: parseFloat(editFormData.cost_price || 0)
      });
      setEditingIngredientId(null);
      fetchIngredients();
      toast.success("Обновлено");
    } catch (err) {
      toast.error('Ошибка: ' + (err.response?.data?.error || ''));
    }
  };

  return (
    <div className="inventory-section animate-fade">
      <div className="flex justify-between items-center mb-6">
        <h2>Склад (Инвентарь)</h2>
      </div>

      <div className="grid" style={{ gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
        
        {/* Ingredients Management */}
        <div className="premium-card">
          <h3 className="mb-4">База ингредиентов</h3>
          <form onSubmit={handleCreateIngredient} className="flex gap-2 mb-4">
            <input type="text" placeholder="Название (Например: Мясо)" value={newIngredient.name} onChange={e=>setNewIngredient({...newIngredient, name: e.target.value})} className="flex-1" />
            <input type="number" placeholder="Кол-во" value={newIngredient.stock} onChange={e=>setNewIngredient({...newIngredient, stock: e.target.value})} style={{width: '100px'}} />
            <select value={newIngredient.unit} onChange={e=>setNewIngredient({...newIngredient, unit: e.target.value})}>
              <option value="gr">gr</option>
              <option value="kg">kg</option>
              <option value="litr">litr</option>
              <option value="ml">ml</option>
              <option value="dona">шт</option>
            </select>
            <input type="number" placeholder="Мин. запас" value={newIngredient.min_stock} onChange={e=>setNewIngredient({...newIngredient, min_stock: e.target.value})} style={{width: '100px'}} />
            <input type="number" placeholder="Себест-ть сум" value={newIngredient.cost_price} onChange={e=>setNewIngredient({...newIngredient, cost_price: e.target.value})} style={{width: '120px'}} />
            <button type="submit" className="btn-primary p-2"><Plus size={18}/></button>
          </form>

          {loading ? <div className="text-center"><Loader2 className="animate-spin" /></div> : (
            <div style={{maxHeight: '500px', overflowY: 'auto'}}>
              {/* Low stock warning banner */}
              {ingredients.filter(i => i.stock <= i.min_stock).length > 0 && (
                <div style={{ background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)', borderRadius: '10px', padding: '0.75rem 1rem', marginBottom: '1rem', display: 'flex', alignItems: 'center', gap: '0.5rem', color: '#ef4444' }}>
                  <AlertTriangle size={16} />
                  <strong>{ingredients.filter(i => i.stock <= i.min_stock).length} ингредиентов заканчиваются!</strong>
                </div>
              )}
              <table className="admin-table">
                <thead><tr><th>Название</th><th>Запас</th><th>Себестоимость</th><th>Статус / Добавить</th><th></th></tr></thead>
                <tbody>
                  {ingredients.map(ing => {
                    const isLow = ing.min_stock > 0 && ing.stock <= ing.min_stock;
                    const isEditing = editingIngredientId === ing.id;
                    return (
                      <tr key={ing.id} style={{background: isLow ? 'rgba(239, 68, 68, 0.07)' : ''}}>
                        {isEditing ? (
                          <>
                            <td>
                              <input type="text" value={editFormData.name} onChange={e => setEditFormData({...editFormData, name: e.target.value})} style={{width: '100%', minWidth: '120px', padding: '0.4rem 0.6rem', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: 'var(--bg-surface)', color: 'var(--text-primary)', outline: 'none', transition: 'border-color 0.2s'}} onFocus={e => e.target.style.borderColor = 'var(--primary)'} onBlur={e => e.target.style.borderColor = 'rgba(0,0,0,0.1)'} />
                            </td>
                            <td>
                              <div style={{display: 'flex', gap: '0.4rem', marginBottom: '0.4rem', alignItems: 'center'}}>
                                <input type="number" step="0.01" value={editFormData.stock} onChange={e => setEditFormData({...editFormData, stock: e.target.value})} style={{width: '75px', padding: '0.4rem', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: 'var(--bg-surface)', color: 'var(--text-primary)', outline: 'none'}} />
                                <select value={editFormData.unit} onChange={e => setEditFormData({...editFormData, unit: e.target.value})} style={{padding: '0.4rem', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: 'var(--bg-surface)', color: 'var(--text-primary)', outline: 'none', cursor: 'pointer'}}>
                                  <option value="gr">gr</option>
                                  <option value="kg">kg</option>
                                  <option value="litr">litr</option>
                                  <option value="ml">ml</option>
                                  <option value="dona">шт</option>
                                </select>
                              </div>
                              <div style={{display: 'flex', alignItems: 'center', gap: '0.4rem'}}>
                                <span style={{fontSize: '0.75rem', color: 'var(--text-muted)', fontWeight: '500'}}>min:</span>
                                <input type="number" step="0.01" value={editFormData.min_stock} onChange={e => setEditFormData({...editFormData, min_stock: e.target.value})} style={{width: '75px', padding: '0.3rem 0.4rem', borderRadius: '6px', fontSize: '0.8rem', border: '1px solid rgba(0,0,0,0.1)', background: 'var(--bg-surface)', color: 'var(--text-primary)', outline: 'none'}} />
                              </div>
                            </td>
                            <td>
                              <div style={{display: 'flex', alignItems: 'center', gap: '0.2rem'}}>
                                <input type="number" step="0.01" value={editFormData.cost_price} onChange={e => setEditFormData({...editFormData, cost_price: e.target.value})} style={{width: '90px', padding: '0.4rem 0.6rem', borderRadius: '6px', border: '1px solid rgba(0,0,0,0.1)', background: 'var(--bg-surface)', color: 'var(--text-primary)', outline: 'none'}} />
                                <span style={{fontSize: '0.8rem', color: 'var(--text-muted)'}}>so'm</span>
                              </div>
                            </td>
                            <td>
                              <div style={{display: 'flex', gap: '0.4rem'}}>
                                <button onClick={() => handleSaveEdit(ing.id)} className="btn-primary" style={{padding: '0.4rem 0.8rem', fontSize: '0.85rem', borderRadius: '6px', fontWeight: '500'}}>Сохранить</button>
                                <button onClick={() => setEditingIngredientId(null)} className="btn-secondary" style={{padding: '0.4rem 0.8rem', fontSize: '0.85rem', borderRadius: '6px', fontWeight: '500', background: '#f1f5f9', color: '#475569', border: 'none'}}>Отмена</button>
                              </div>
                            </td>
                            <td></td>
                          </>
                        ) : (
                          <>
                            <td style={{ fontWeight: 600 }}>{ing.name}</td>
                            <td className="font-bold" style={{ color: isLow ? '#ef4444' : 'var(--success)' }}>
                              {ing.stock} {ing.unit}
                              {ing.min_stock > 0 && <span style={{ color: 'var(--text-muted)', fontWeight: 400, fontSize: '0.8rem' }}> / min: {ing.min_stock}</span>}
                            </td>
                            <td style={{ color: 'var(--text-muted)', fontSize: '0.85rem' }}>
                              {ing.cost_price > 0 ? `${ing.cost_price.toLocaleString()} so'm` : '—'}
                            </td>
                            <td>
                              {isLow ? (
                                <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                                  <span style={{ color: '#ef4444', fontSize: '0.8rem', display: 'flex', alignItems: 'center', gap: '3px' }}>
                                    <AlertTriangle size={12}/> Мало!
                                  </span>
                                  <input
                                    type="number" min="0" step="0.1"
                                    placeholder="+кол-во"
                                    value={restockAmounts[ing.id] || ''}
                                    onChange={e => setRestockAmounts(prev => ({ ...prev, [ing.id]: e.target.value }))}
                                    style={{ width: '75px', padding: '0.25rem 0.4rem', borderRadius: '6px', background: 'var(--bg-surface)', border: '1px solid rgba(239,68,68,0.4)', color: 'var(--text-primary)', fontSize: '0.85rem' }}
                                  />
                                  <button onClick={() => handleRestock(ing.id)} className="btn-primary" style={{ padding: '0.25rem 0.6rem', fontSize: '0.8rem' }}>+</button>
                                </div>
                              ) : <span className="text-success" style={{ fontSize: '0.85rem' }}>Достаточно</span>}
                            </td>
                            <td>
                              <div style={{display: 'flex', gap: '0.5rem', alignItems: 'center'}}>
                                <button onClick={() => handleEditClick(ing)} className="icon-btn text-primary"><Edit2 size={16}/></button>
                                <button onClick={() => handleDeleteIngredient(ing.id)} className="icon-btn text-danger"><Trash2 size={16}/></button>
                              </div>
                            </td>
                          </>
                        )}
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Recipes Management */}
        <div className="premium-card">
          <h3 className="mb-4">Рецепты (Состав блюда)</h3>
          
          <div className="input-group mb-4">
            <label>Выберите блюдо из меню:</label>
            <select value={selectedProduct} onChange={e=>setSelectedProduct(e.target.value)}>
              <option value="">-- Выберите --</option>
              {products.map(p => (
                <option key={p.id} value={p.id}>{p.name}</option>
              ))}
            </select>
          </div>

          {selectedProduct && (
            <div className="animate-fade">
              <form onSubmit={handleAddRecipe} className="flex flex-wrap items-center gap-2 mb-4 bg-gray-50 dark:bg-gray-800 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <select 
                  value={newRecipe.ingredient_id} 
                  onChange={e => {
                    const ingId = e.target.value;
                    const ing = ingredients.find(i => i.id == ingId);
                    setNewRecipe({
                      ...newRecipe, 
                      ingredient_id: ingId,
                      unit: ing ? ing.unit : 'gr'
                    });
                  }}
                  className="flex-1 min-w-[200px]"
                >
                  <option value="">Выберите ингредиент...</option>
                  {ingredients.map(i => <option key={i.id} value={i.id}>{i.name} ({i.unit})</option>)}
                </select>
                <div className="flex items-center gap-2">
                  <input 
                    type="number" 
                    step="0.01"
                    placeholder="Кол-во" 
                    value={newRecipe.quantity} 
                    onChange={e => setNewRecipe({...newRecipe, quantity: e.target.value})} 
                    style={{width: '90px'}} 
                  />
                  <select 
                    value={newRecipe.unit} 
                    onChange={e => setNewRecipe({...newRecipe, unit: e.target.value})}
                    style={{width: '80px', minWidth: '80px'}}
                  >
                    <option value="gr">gr</option>
                    <option value="kg">kg</option>
                    <option value="litr">litr</option>
                    <option value="ml">ml</option>
                    <option value="dona">шт</option>
                  </select>
                  <button type="submit" className="btn-primary p-2"><Link size={18}/></button>
                </div>
              </form>

              <table className="admin-table">
                <thead><tr><th>Ингредиент</th><th>Количество</th><th></th></tr></thead>
                <tbody>
                  {recipes.length > 0 ? recipes.map(r => (
                    <tr key={r.id}>
                      <td>{r.ingredient_name}</td>
                      <td className="font-bold">{r.quantity} {r.unit === 'dona' ? 'шт' : r.unit}</td>
                      <td>
                        <button onClick={() => handleDeleteRecipe(r.id)} className="icon-btn text-danger"><Trash2 size={16}/></button>
                      </td>
                    </tr>
                  )) : (
                    <tr><td colSpan="3" className="text-center py-4 text-muted">Пока ничего не добавлено</td></tr>
                  )}
                </tbody>
              </table>
              
              <div className="mt-4 hint-text" style={{fontSize: '0.8rem', color: 'var(--text-dim)'}}>
                * При заказе одной порции (или штуки) вышеуказанные количества автоматически списываются со склада.
              </div>
            </div>
          )}
        </div>

      </div>
    </div>
  );
};

export default InventorySection;
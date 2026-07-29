import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Loader2, Plus, AlertTriangle, Trash2, Link } from 'lucide-react';

const InventorySection = ({ products }) => {
  const [ingredients, setIngredients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newIngredient, setNewIngredient] = useState({ name: '', stock: '', unit: 'gr', min_stock: '', cost_price: '' });
  const [restockAmounts, setRestockAmounts] = useState({}); // {ingredientId: amount}
  
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
    if (!newIngredient.name || !newIngredient.stock) return alert('Заполните все поля');
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
      alert('Ошибка: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteIngredient = async (id) => {
    if (!window.confirm("Вы действительно хотите удалить?")) return;
    try {
      await api.delete(`/inventory/ingredients/${id}`);
      fetchIngredients();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Ошибка при удалении');
    }
  };

  const handleAddRecipe = async (e) => {
    e.preventDefault();
    if (!selectedProduct) return alert('Сначала выберите продукт');
    if (!newRecipe.ingredient_id || !newRecipe.quantity) return alert('Выберите ингредиент и количество');
    
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
      alert('Ошибка: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteRecipe = async (id) => {
    if (!window.confirm("Удалить?")) return;
    try {
      await api.delete(`/inventory/recipes/${id}`);
      fetchRecipes(selectedProduct);
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Ошибка');
    }
  };

  const handleRestock = async (ingId) => {
    const amount = parseFloat(restockAmounts[ingId] || 0);
    if (!amount || amount <= 0) return alert("Введите количество");
    try {
      await api.post(`/inventory/ingredients/${ingId}/restock`, { amount });
      setRestockAmounts(prev => ({ ...prev, [ingId]: '' }));
      fetchIngredients();
    } catch (err) {
      alert('Ошибка при пополнении запаса: ' + (err.response?.data?.error || err.message));
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
                    return (
                      <tr key={ing.id} style={{background: isLow ? 'rgba(239, 68, 68, 0.07)' : ''}}>
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
                          <button onClick={() => handleDeleteIngredient(ing.id)} className="icon-btn text-danger"><Trash2 size={16}/></button>
                        </td>
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
              <form onSubmit={handleAddRecipe} className="flex gap-2 mb-4 bg-gray-50 dark:bg-gray-800 p-3 rounded-lg border border-gray-200 dark:border-gray-700">
                <select 
                  value={newRecipe.ingredient_id} 
                  onChange={e => setNewRecipe({...newRecipe, ingredient_id: e.target.value})}
                  className="flex-1"
                >
                  <option value="">Выберите ингредиент...</option>
                  {ingredients.map(i => <option key={i.id} value={i.id}>{i.name} ({i.unit})</option>)}
                </select>
                <input 
                  type="number" 
                  step="0.01"
                  placeholder="Кол-во" 
                  value={newRecipe.quantity} 
                  onChange={e => setNewRecipe({...newRecipe, quantity: e.target.value})} 
                  style={{width: '90px'}} 
                />
                <button type="submit" className="btn-primary p-2"><Link size={18}/></button>
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

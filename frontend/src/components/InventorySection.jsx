import React, { useState, useEffect } from 'react';
import api from '../store/authStore';
import { Loader2, Plus, AlertTriangle, Trash2, Link } from 'lucide-react';

const InventorySection = ({ products }) => {
  const [ingredients, setIngredients] = useState([]);
  const [loading, setLoading] = useState(false);
  const [newIngredient, setNewIngredient] = useState({ name: '', stock: '', unit: 'gr', min_stock: '', cost_price: '' });
  
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
      // eslint-disable-next-line no-unused-vars
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
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      console.error(err);
    }
  };

  const handleCreateIngredient = async (e) => {
    e.preventDefault();
    if (!newIngredient.name || !newIngredient.stock) return alert('Barcha maydonlarni to\'ldiring');
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
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Xatolik: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteIngredient = async (id) => {
    if (!window.confirm("Rostdan o'chirmoqchimisiz?")) return;
    try {
      await api.delete(`/inventory/ingredients/${id}`);
      fetchIngredients();
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('O\'chirishda xatolik');
    }
  };

  const handleAddRecipe = async (e) => {
    e.preventDefault();
    if (!selectedProduct) return alert('Avval mahsulot tanlang');
    if (!newRecipe.ingredient_id || !newRecipe.quantity) return alert('Masalliq va miqdorini tanlang');
    
    try {
      await api.post('/inventory/recipes', {
        product_id: parseInt(selectedProduct),
        ingredient_id: parseInt(newRecipe.ingredient_id),
        quantity: parseFloat(newRecipe.quantity),
        unit: newRecipe.unit
      });
      setNewRecipe({ ingredient_id: '', quantity: '', unit: 'gr' });
      fetchRecipes(selectedProduct);
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Xatolik: ' + (err.response?.data?.error || ''));
    }
  };

  const handleDeleteRecipe = async (id) => {
    if (!window.confirm("O'chirmoqchimisiz?")) return;
    try {
      await api.delete(`/inventory/recipes/${id}`);
      fetchRecipes(selectedProduct);
      // eslint-disable-next-line no-unused-vars
    } catch (err) {
      alert('Xatolik');
    }
  };

  return (
    <div className="inventory-section animate-fade">
      <div className="flex justify-between items-center mb-6">
        <h2>Omborxona (Sklad)</h2>
      </div>

      <div className="grid" style={{ gridTemplateColumns: '1fr 1fr', gap: '2rem' }}>
        
        {/* Ingredients Management */}
        <div className="premium-card">
          <h3 className="mb-4">Masalliqlar Bazasi</h3>
          <form onSubmit={handleCreateIngredient} className="flex gap-2 mb-4">
            <input type="text" placeholder="Nomi (Masalan: Go'sht)" value={newIngredient.name} onChange={e=>setNewIngredient({...newIngredient, name: e.target.value})} className="flex-1" />
            <input type="number" placeholder="Mavjud miqdor" value={newIngredient.stock} onChange={e=>setNewIngredient({...newIngredient, stock: e.target.value})} style={{width: '100px'}} />
            <select value={newIngredient.unit} onChange={e=>setNewIngredient({...newIngredient, unit: e.target.value})}>
              <option value="gr">gr</option>
              <option value="kg">kg</option>
              <option value="litr">litr</option>
              <option value="ml">ml</option>
              <option value="dona">dona</option>
            </select>
            <input type="number" placeholder="Min. chegara" value={newIngredient.min_stock} onChange={e=>setNewIngredient({...newIngredient, min_stock: e.target.value})} style={{width: '100px'}} />
            <input type="number" placeholder="Tannarxi so'm" value={newIngredient.cost_price} onChange={e=>setNewIngredient({...newIngredient, cost_price: e.target.value})} style={{width: '120px'}} />
            <button type="submit" className="btn-primary p-2"><Plus size={18}/></button>
          </form>

          {loading ? <div className="text-center"><Loader2 className="animate-spin" /></div> : (
            <div style={{maxHeight: '400px', overflowY: 'auto'}}>
              <table className="admin-table">
                <thead><tr><th>Nomi</th><th>Zahira</th><th>Holat</th><th></th></tr></thead>
                <tbody>
                  {ingredients.map(ing => {
                    const isLow = ing.stock <= ing.min_stock;
                    return (
                      <tr key={ing.id} style={{background: isLow ? 'rgba(239, 68, 68, 0.05)' : ''}}>
                        <td>{ing.name}</td>
                        <td className="font-bold">{ing.stock} {ing.unit}</td>
                        <td>
                          {isLow ? <span className="text-danger flex items-center gap-1"><AlertTriangle size={14}/> Kam qoldi</span> : <span className="text-success">Yetarli</span>}
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
          <h3 className="mb-4">Retseptlar (Taom tarkibi)</h3>
          
          <div className="input-group mb-4">
            <label>Menyudan taomni tanlang:</label>
            <select value={selectedProduct} onChange={e=>setSelectedProduct(e.target.value)}>
              <option value="">-- Tanlang --</option>
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
                  <option value="">Masalliqni tanlang...</option>
                  {ingredients.map(i => <option key={i.id} value={i.id}>{i.name} ({i.unit})</option>)}
                </select>
                <input 
                  type="number" 
                  step="0.01"
                  placeholder="Miqdori" 
                  value={newRecipe.quantity} 
                  onChange={e => setNewRecipe({...newRecipe, quantity: e.target.value})} 
                  style={{width: '90px'}} 
                />
                <button type="submit" className="btn-primary p-2"><Link size={18}/></button>
              </form>

              <table className="admin-table">
                <thead><tr><th>Masalliq</th><th>Biriktirilgan miqdor</th><th></th></tr></thead>
                <tbody>
                  {recipes.length > 0 ? recipes.map(r => (
                    <tr key={r.id}>
                      <td>{r.ingredient_name}</td>
                      <td className="font-bold">{r.quantity} {r.unit}</td>
                      <td>
                        <button onClick={() => handleDeleteRecipe(r.id)} className="icon-btn text-danger"><Trash2 size={16}/></button>
                      </td>
                    </tr>
                  )) : (
                    <tr><td colSpan="3" className="text-center py-4 text-muted">Hali hech narsa biriktirilmagan</td></tr>
                  )}
                </tbody>
              </table>
              
              <div className="mt-4 hint-text" style={{fontSize: '0.8rem', color: 'var(--text-dim)'}}>
                * Bitta pors (yoki bitta) taom buyurtma qilinganda avtomatik ravishda yuqoridagi miqdorlar ombordan ayriladi.
              </div>
            </div>
          )}
        </div>

      </div>
    </div>
  );
};

export default InventorySection;

import { create } from 'zustand';

export const useCartStore = create((set, get) => ({
  items: [],
  
  addItem: (product) => {
    const items = get().items;
    const existing = items.find(i => i.id === product.id);
    
    if (existing) {
      set({
        items: items.map(i => i.id === product.id 
          ? { ...i, quantity: i.quantity + 1 } 
          : i
        )
      });
    } else {
      set({ items: [...items, { ...product, quantity: 1 }] });
    }
  },
  
  removeItem: (productId) => {
    set({ items: get().items.filter(i => i.id !== productId) });
  },
  
  updateQuantity: (productId, delta) => {
    const items = get().items.map(i => {
      if (i.id === productId) {
        const newQty = Math.max(1, i.quantity + delta);
        return { ...i, quantity: newQty };
      }
      return i;
    });
    set({ items });
  },
  
  clearCart: () => set({ items: [] }),
  
  getTotal: () => {
    return get().items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  }
}));

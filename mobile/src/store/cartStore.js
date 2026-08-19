import { create } from 'zustand';

export const useCartStore = create((set, get) => ({
  items: [],
  
  addItem: (product) => {
    set((state) => {
      const existing = state.items.find(i => i.product_id === product.id);
      if (existing) {
        return {
          items: state.items.map(i => 
            i.product_id === product.id 
              ? { ...i, quantity: i.quantity + product.quantity_step }
              : i
          )
        };
      }
      return {
        items: [...state.items, {
          product_id: product.id,
          name: product.name,
          price: product.price,
          unit: product.unit,
          quantity: product.min_quantity || 1,
          step: product.quantity_step || 1
        }]
      };
    });
  },

  updateQuantity: (productId, delta) => {
    set((state) => ({
      items: state.items.map(i => {
        if (i.product_id === productId) {
          const newQ = i.quantity + delta;
          return newQ > 0 ? { ...i, quantity: newQ } : i;
        }
        return i;
      }).filter(i => i.quantity > 0)
    }));
  },

  removeItem: (productId) => {
    set((state) => ({
      items: state.items.filter((i) => i.product_id !== productId)
    }));
  },

  clearCart: () => set({ items: [] }),

  getTotalPrice: () => {
    const items = get().items;
    return items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  }
}));

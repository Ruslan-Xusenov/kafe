import { create } from 'zustand';

// ID for 'Bir martalik idish' (Disposable Container)
const CONTAINER_PRODUCT_ID = 7;

export const useCartStore = create((set, get) => ({
  items: [],
  
  addItem: (product) => {
    const items = get().items;
    const itemUnit = product.unit || 'dona';
    const existing = items.find(i => i.id === product.id && i.unit === itemUnit);
    
    // Determine initial quantity: use min_quantity or default to 1
    const initialQty = product.min_quantity || 1;
    
    let newItems;
    if (existing) {
      const step = product.quantity_step || 1;
      newItems = items.map(i => (i.id === product.id && i.unit === itemUnit)
        ? { ...i, quantity: i.quantity + step } 
        : i
      );
    } else {
      newItems = [...items, { ...product, quantity: initialQty, unit: itemUnit }];
    }
    
    // Handle mandatory container logic
    if (product.has_mandatory_container && product.id !== CONTAINER_PRODUCT_ID) {
      const containerItem = newItems.find(i => i.id === CONTAINER_PRODUCT_ID);
      const step = product.quantity_step || 1;
      const qtyToAdd = existing ? step : initialQty;
      
      if (containerItem) {
        newItems = newItems.map(i => i.id === CONTAINER_PRODUCT_ID 
          ? { ...i, quantity: i.quantity + qtyToAdd } 
          : i
        );
      } else {
        const containerInfo = {
          id: CONTAINER_PRODUCT_ID,
          name: 'Bir martalik idish',
          price: 1000,
          quantity: qtyToAdd,
          unit: 'dona'
        };
        newItems.push(containerInfo);
      }
    }
    
    set({ items: newItems });
  },
  
  removeItem: (productId, unit) => {
    const items = get().items;
    const itemToRemove = items.find(i => i.id === productId && i.unit === unit);
    let newItems = items.filter(i => !(i.id === productId && i.unit === unit));
    
    if (itemToRemove && itemToRemove.has_mandatory_container && productId !== CONTAINER_PRODUCT_ID) {
        const containerItem = newItems.find(i => i.id === CONTAINER_PRODUCT_ID);
        if (containerItem) {
            const newQty = Math.max(0, containerItem.quantity - itemToRemove.quantity);
            if (newQty === 0) {
                newItems = newItems.filter(i => i.id !== CONTAINER_PRODUCT_ID);
            } else {
                newItems = newItems.map(i => i.id === CONTAINER_PRODUCT_ID ? { ...i, quantity: newQty } : i);
            }
        }
    }
    
    set({ items: newItems });
  },
  
  updateQuantity: (productId, unit, delta) => {
    const items = get().items;
    const item = items.find(i => i.id === productId && i.unit === unit);
    if (!item) return;

    const step = item.quantity_step || 1;
    const minQty = item.min_quantity || 1;
    const newQty = Math.max(minQty, item.quantity + (delta > 0 ? step : -step));
    
    const actualDelta = newQty - item.quantity;

    let newItems = items.map(i => {
      if (i.id === productId && i.unit === unit) {
        return { ...i, quantity: newQty };
      }
      return i;
    });

    // Sync containers if needed
    if (item.has_mandatory_container && productId !== CONTAINER_PRODUCT_ID) {
      const containerItem = newItems.find(i => i.id === CONTAINER_PRODUCT_ID);
      if (containerItem) {
        newItems = newItems.map(i => i.id === CONTAINER_PRODUCT_ID 
            ? { ...i, quantity: Math.max(0, i.quantity + actualDelta) } 
            : i
        ).filter(i => i.id !== CONTAINER_PRODUCT_ID || i.quantity > 0);
      }
    }
    
    set({ items: newItems });
  },
  
  clearCart: () => set({ items: [] }),
  
  getTotal: () => {
    return get().items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  }
}));

import { create } from 'zustand';

// ID for 'Bir martalik idish' (Disposable Container)
let CONTAINER_PRODUCT_ID = 7; // Default, will be updated by fetchSettings

export const useCartStore = create((set, get) => ({
  items: [],
  containerPrice: 1000,
  containerId: 7,
  
  fetchSettings: async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL || 'https://kafe.ruslandev.uz'}/api/catalog/settings`);
      const data = await response.json();
      if (data.container_price) {
        set({ containerPrice: parseInt(data.container_price) });
      }
      if (data.container_product_id) {
        const newId = parseInt(data.container_product_id);
        set({ containerId: newId });
        CONTAINER_PRODUCT_ID = newId; // Update local variable for easier access
      }
      
      // Sync cart items with new settings
      const items = get().items;
      if (items.length > 0) {
          const newItems = items.map(i => {
              if (i.id === CONTAINER_PRODUCT_ID || i.name === 'Bir martalik idish') {
                  return { ...i, id: get().containerId, price: get().containerPrice };
              }
              return i;
          });
          set({ items: newItems });
      }
    } catch (err) { console.error("Failed to fetch settings", err); }
  },

  addItem: (product) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    const itemUnit = product.unit || 'dona';
    const existing = items.find(i => i.id === product.id && i.unit === itemUnit);
    
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
    
    // Handle mandatory container logic (4 dona = 1 portion)
    if (product.has_mandatory_container && product.id !== currentContainerId) {
      const containerItem = newItems.find(i => i.id === currentContainerId);
      const step = product.quantity_step || 1;
      const addedQty = existing ? step : initialQty;
      
      let containerDelta = addedQty;
      if (itemUnit === 'dona') {
        containerDelta = addedQty / 4.0;
      }
      
      if (containerItem) {
        newItems = newItems.map(i => i.id === currentContainerId 
          ? { ...i, quantity: i.quantity + containerDelta, price: get().containerPrice } 
          : i
        );
      } else {
        newItems.push({
          id: currentContainerId,
          name: 'Bir martalik idish',
          price: get().containerPrice,
          quantity: containerDelta,
          unit: 'dona'
        });
      }
    }
    
    set({ items: newItems });
  },
  
  removeItem: (productId, unit) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    
    // If we're removing the container itself, remove ALL items that require a container!
    if (productId === currentContainerId) {
        const newItems = items.filter(i => !i.has_mandatory_container && i.id !== currentContainerId);
        set({ items: newItems });
        return;
    }

    const itemToRemove = items.find(i => i.id === productId && i.unit === unit);
    let newItems = items.filter(i => !(i.id === productId && i.unit === unit));
    
    if (itemToRemove && itemToRemove.has_mandatory_container && productId !== currentContainerId) {
        const containerItem = newItems.find(i => i.id === currentContainerId);
        if (containerItem) {
            let containerSub = itemToRemove.quantity;
            if (unit === 'dona') containerSub = itemToRemove.quantity / 4.0;
            
            const newQty = Math.max(0, containerItem.quantity - containerSub);
            if (newQty <= 0.001) { // Floating point safety
                newItems = newItems.filter(i => i.id !== currentContainerId);
            } else {
                newItems = newItems.map(i => i.id === currentContainerId ? { ...i, quantity: newQty } : i);
            }
        }
    }
    
    set({ items: newItems });
  },
  
  updateQuantity: (productId, unit, delta) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    const item = items.find(i => i.id === productId && i.unit === unit);
    if (!item) return;

    const step = item.quantity_step || 1;
    const minQty = item.min_quantity || 1;
    const newQty = Math.max(minQty, item.quantity + (delta > 0 ? step : -step));
    
    const actualDelta = newQty - item.quantity;
    let containerDelta = actualDelta;
    if (unit === 'dona') containerDelta = actualDelta / 4.0;

    let newItems = items.map(i => {
      if (i.id === productId && i.unit === unit) {
        return { ...i, quantity: newQty };
      }
      return i;
    });

    if (item.has_mandatory_container && productId !== currentContainerId) {
      const containerItem = newItems.find(i => i.id === currentContainerId);
      if (containerItem) {
        newItems = newItems.map(i => i.id === currentContainerId 
            ? { ...i, quantity: Math.max(0, i.quantity + containerDelta) } 
            : i
        ).filter(i => i.id !== currentContainerId || i.quantity > 0.001);
      }
    }
    
    set({ items: newItems });
  },
  
  clearCart: () => set({ items: [] }),
  
  getTotal: () => {
    return get().items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  }
}));

import { create } from 'zustand';
import api from './authStore';

// ID for 'Bir martalik idish' (Disposable Container)
let CONTAINER_PRODUCT_ID = 7; // Default, will be updated by fetchSettings

const consolidateItems = (items) => {
    const newItems = [];
    const groups = {};
    items.forEach(i => {
        if (!groups[i.id]) groups[i.id] = [];
        groups[i.id].push(i);
    });

    for (let id in groups) {
        const productItems = groups[id];
        let porsItem = productItems.find(i => i.unit === 'pors');
        let donaItem = productItems.find(i => i.unit === 'dona');

        if (donaItem && donaItem.base_unit === 'pors') {
             const portions = Math.floor(donaItem.quantity / 4);
             if (portions > 0) {
                 if (porsItem) {
                     porsItem.quantity += portions;
                 } else {
                     porsItem = { ...donaItem, unit: 'pors', price: donaItem.price * 4, quantity: portions };
                     productItems.push(porsItem);
                 }
                 donaItem.quantity -= (portions * 4);
             }
        }
        
        productItems.forEach(i => {
            if (i.quantity > 0) newItems.push(i);
        });
    }
    return newItems;
};

const recalculateContainers = (items, currentContainerId, containerPrice) => {
    let totalPortions = 0;
    const filteredItems = items.filter(i => i.id !== currentContainerId); 
    
    filteredItems.forEach(i => {
       if (i.has_mandatory_container) {
          if (i.unit === 'gr') {
              totalPortions += (i.quantity / 100.0);
          } else if (i.unit === 'kg') {
              totalPortions += (i.quantity * 10.0);
          } else {
              // for 'pors' and 'dona', 1 quantity = 1 container
              totalPortions += i.quantity;
          }
       }
    });

    const neededContainers = Math.ceil(totalPortions);
    
    if (neededContainers > 0) {
        filteredItems.push({
            id: currentContainerId,
            name: 'Bir martalik idish',
            price: containerPrice,
            quantity: neededContainers,
            unit: 'dona'
        });
    }
    return filteredItems;
};

export const useCartStore = create((set, get) => ({
  items: [],
  containerPrice: 1000,
  containerId: 7,
  
  fetchSettings: async () => {
    try {
      const response = await api.get('/catalog/settings');
      const data = response.data;
      
      let price = 1000;
      let newId = 7;

      if (data.container_price) {
        price = parseInt(data.container_price);
        set({ containerPrice: price });
      }
      if (data.container_product_id) {
        newId = parseInt(data.container_product_id);
        set({ containerId: newId });
        CONTAINER_PRODUCT_ID = newId; 
      }
      
      // Update existing items in cart to match new settings (Price and ID)
      const currentItems = get().items;
      if (currentItems.length > 0) {
          const updatedItems = currentItems.map(i => {
              // Match by current known ID, the new ID, or name
              if (i.id === 7 || i.id === newId || i.name === 'Bir martalik idish') {
                  return { ...i, id: newId, price: price };
              }
              return i;
          });
          set({ items: updatedItems });
      }
      // eslint-disable-next-line no-unused-vars
    } catch (err) { console.error("Failed to fetch settings", err); }
  },

  addItem: (product) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    const itemUnit = product.unit || 'dona';
    const existing = items.find(i => i.id === product.id && i.unit === itemUnit);
    
    // Use product.quantity if provided, otherwise use step/min_quantity
    const qtyToAdd = product.quantity !== undefined ? product.quantity : (product.quantity_step || 1);
    const initialQty = product.quantity !== undefined ? product.quantity : (product.min_quantity || 1);

    let newItems;
    if (existing) {
      newItems = items.map(i => (i.id === product.id && i.unit === itemUnit)
        ? { ...i, quantity: i.quantity + qtyToAdd } 
        : i
      );
    } else {
      newItems = [...items, { ...product, quantity: initialQty, unit: itemUnit }];
    }

    
    set({ items: recalculateContainers(consolidateItems(newItems), currentContainerId, get().containerPrice) });
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

    let newItems = items.filter(i => !(i.id === productId && i.unit === unit));
    set({ items: recalculateContainers(consolidateItems(newItems), currentContainerId, get().containerPrice) });
  },
  
  updateQuantity: (productId, unit, delta) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    
    if (productId === currentContainerId) {
       // Allow manual modification of container qty? The user might just want to change it.
       // But to keep it simple, we just ignore manual changes if it's below the mandatory limit,
       // actually, let's just let it recalculate normally if they touch other items.
    }

    const item = items.find(i => i.id === productId && i.unit === unit);
    if (!item) return;

    const step = item.quantity_step || 1;
    const minQty = item.min_quantity || 1;
    const newQty = Math.max(minQty, Math.round((item.quantity + (delta > 0 ? step : -step)) * 1000) / 1000);
    
    let newItems = items.map(i => {
      if (i.id === productId && i.unit === unit) {
        return { ...i, quantity: newQty };
      }
      return i;
    });

    set({ items: recalculateContainers(consolidateItems(newItems), currentContainerId, get().containerPrice) });
  },

  setQuantity: (productId, unit, quantity) => {
    const items = get().items;
    const currentContainerId = get().containerId;
    const item = items.find(i => i.id === productId && i.unit === unit);
    if (!item) return;

    const minQty = item.min_quantity || 1;
    const validQty = Math.max(minQty, quantity || minQty);

    let newItems = items.map(i => {
      if (i.id === productId && i.unit === unit) {
         return { ...i, quantity: validQty };
      }
      return i;
    });
    set({ items: recalculateContainers(consolidateItems(newItems), currentContainerId, get().containerPrice) });
  },
  
  clearCart: () => set({ items: [] }),
  
  getTotal: () => {
    return get().items.reduce((sum, item) => sum + (item.price * item.quantity), 0);
  }
}));

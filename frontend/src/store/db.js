import Dexie from 'dexie';

export const db = new Dexie('KafePOSDB');

db.version(2).stores({
  posTables: 'id, name, floor, status',
  categories: 'id, name',
  products: 'id, category_id, name',
  offlineOrders: 'id, table_id, status, created_at',
  syncQueue: '++id, type, payload, status, created_at, order_id' // type: CREATE_ORDER, ADD_ITEMS, CLOSE_TABLE
});

// Helper for temporary offline IDs (negative numbers to avoid conflict with backend serial IDs)
export const generateOfflineId = () => {
  return -Math.floor(Date.now() / 1000);
};

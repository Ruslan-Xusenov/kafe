import React, { useEffect, useState } from 'react';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';
import { db } from '../store/db';
import api from '../store/authStore';

const SyncManager = () => {
  const [isOnline, setIsOnline] = useState(navigator.onLine);
  const [isSyncing, setIsSyncing] = useState(false);
  const [pendingCount, setPendingCount] = useState(0);

  const updatePendingCount = async () => {
    const count = await db.syncQueue.where('status').equals('pending').count();
    setPendingCount(count);
  };

  const processSyncQueue = async () => {
    if (isSyncing || !navigator.onLine) return;
    
    const pendingItems = await db.syncQueue.where('status').equals('pending').toArray();
    if (pendingItems.length === 0) return;

    setIsSyncing(true);
    let idMap = {}; // Maps offline negative IDs to real backend IDs

    for (const item of pendingItems) {
      try {
        if (item.type === 'CREATE_ORDER') {
          const payload = { ...item.payload };
          // If table was also created offline, we might need mapping, but tables are mostly read-only
          
          const res = await api.post('/orders', payload);
          if (res.data && res.data.id) {
            idMap[item.order_id] = res.data.id;

            // Persist the mapping on dependent queue items so a page reload
            // cannot leave ADD_ITEMS stuck with a negative offline ID.
            const dependentItems = await db.syncQueue
              .where('status').equals('pending')
              .filter(queueItem => queueItem.type === 'ADD_ITEMS' && queueItem.order_id === item.order_id)
              .toArray();
            for (const dependentItem of dependentItems) {
              await db.syncQueue.update(dependentItem.id, { real_order_id: res.data.id });
            }
            
            // Delete the offline order from local DB
            await db.offlineOrders.delete(item.order_id);
            
            // Mark queue item as synced
            await db.syncQueue.update(item.id, { status: 'synced', real_id: res.data.id });
          }
        }
        else if (item.type === 'ADD_ITEMS') {
          let realOrderId = item.real_order_id || item.order_id;
          // If it references an offline order, translate it
          if (realOrderId < 0 && idMap[realOrderId]) {
            realOrderId = idMap[realOrderId];
          }
          
          if (realOrderId < 0) {
            console.error('Cannot sync items for unsynced offline order', item);
            continue; // Skip for now, maybe retry later
          }

		  await api.post(`/orders/${realOrderId}/add-items`, item.payload);
          await db.syncQueue.update(item.id, { status: 'synced' });
        }
      } catch (err) {
        console.error('Sync error for item', item, err);
      }
    }

    // Clean up synced items
    await db.syncQueue.where('status').equals('synced').delete();
    
    setIsSyncing(false);
    updatePendingCount();
  };

  useEffect(() => {
    const handleOnline = () => {
      setIsOnline(true);
      processSyncQueue();
    };
    const handleOffline = () => setIsOnline(false);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    // Initial check
    updatePendingCount();

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);



  // Listen to dexie changes to update badge
  useEffect(() => {
    const interval = setInterval(updatePendingCount, 5000);
    return () => clearInterval(interval);
  }, []);

  if (isOnline && pendingCount === 0) return null;

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {!isOnline && (
        <div className="bg-red-500 text-white px-4 py-2 rounded-full shadow-lg flex items-center gap-2 animate-pulse">
          <WifiOff size={18} />
          <span className="font-medium text-sm">Offline</span>
        </div>
      )}
      
      {pendingCount > 0 && (
        <div className="bg-yellow-500 text-white px-4 py-2 rounded-full shadow-lg flex items-center gap-2 cursor-pointer" onClick={processSyncQueue}>
          <RefreshCw size={18} className={isSyncing ? "animate-spin" : ""} />
          <span className="font-medium text-sm">
            {isSyncing ? 'Синхронизация...' : `Ожидают: ${pendingCount}`}
          </span>
        </div>
      )}
    </div>
  );
};

export default SyncManager;

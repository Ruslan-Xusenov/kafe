import React, { useState, useEffect } from 'react';
import { useWebsocket } from '../hooks/useWebsocket';
import { Printer as PrinterIcon, Receipt } from 'lucide-react';

const Printer = () => {
  const [latestOrder, setLatestOrder] = useState(null);
  const [history, setHistory] = useState([]);

  const onWSMessage = (data) => {
    if (data.type === 'new_order') {
      setLatestOrder(data.order);
      setHistory(prev => [data.order, ...prev].slice(0, 50));
      
      // Avtomatik print
      setTimeout(() => {
        window.print();
      }, 500);
    }
  };

  useWebsocket(onWSMessage);

  const printOrder = (order) => {
    setLatestOrder(order);
    setTimeout(() => {
      window.print();
    }, 100);
  };

  return (
    <div className="printer-page">
      <div className="no-print printer-dashboard glass">
        <div className="dashboard-header">
          <PrinterIcon size={32} className="text-primary" />
          <h1>Avtomatik Kassa Printeri</h1>
          <p className="status-badge new">Kutmoqda...</p>
        </div>
        
        <p className="help-text">
          Yangi buyurtma tushganda chek avtomatik ravishda printerga yuboriladi. Ushbu sahifani yopmang.
        </p>

        <div className="history-list">
          <h3>Oxirgi buyurtmalar</h3>
          {history.length === 0 ? <p className="text-dim">Hozircha buyurtma yo'q</p> : null}
          {history.map(order => (
            <div key={order.id} className="history-item">
              <span>Buyurtma #{order.id} - {new Date(order.created_at).toLocaleTimeString()}</span>
              <button className="btn-primary-sm" onClick={() => printOrder(order)}>Qayta print qolish</button>
            </div>
          ))}
        </div>
      </div>

      {/* Printing Area: Only visible in browser print preview */}
      {latestOrder && (
        <div className="print-area">
          <div className="receipt">
            <h2 className="cafe-name">KAFE NOMi</h2>
            <div className="receipt-divider">======================</div>
            <p className="receipt-text">Buyurtma #{latestOrder.id}</p>
            <p className="receipt-text">Sana: {new Date(latestOrder.created_at).toLocaleString()}</p>
            <p className="receipt-text">Mijoz: {latestOrder.phone}</p>
            {latestOrder.address && <p className="receipt-text">Manzil: {latestOrder.address}</p>}
            <div className="receipt-divider">======================</div>
            
            <table className="receipt-table">
              <thead>
                <tr>
                  <th align="left">Nomi</th>
                  <th align="center">Miqdori</th>
                  <th align="right">Narx</th>
                </tr>
              </thead>
              <tbody>
                {latestOrder.items?.map((item, idx) => (
                  <tr key={idx}>
                    <td>{item.product_name}</td>
                    <td align="center">{item.quantity}</td>
                    <td align="right">{(item.price * item.quantity).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            
            <div className="receipt-divider">======================</div>
            <h3 className="receipt-total">JAMI: {latestOrder.total_price?.toLocaleString()} so'm</h3>
            <div className="receipt-divider">======================</div>
            <p className="receipt-footer">Xaridingiz uchun rahmat!</p>
            <p className="receipt-footer">Yoqimli ishtaha!</p>
          </div>
        </div>
      )}

      <style>{`
        /* Dashboard styling (hidden when printing) */
        .printer-page {
          padding: 2rem;
          min-height: calc(100vh - 100px);
        }

        .printer-dashboard {
          max-width: 600px;
          margin: 0 auto;
          padding: 2rem;
          border-radius: 16px;
        }

        .dashboard-header {
          display: flex;
          align-items: center;
          gap: 1rem;
          margin-bottom: 1rem;
        }

        .help-text {
          color: var(--text-dim);
          background: rgba(255,255,255,0.05);
          padding: 1rem;
          border-radius: 8px;
          margin-bottom: 2rem;
        }

        .history-list h3 { margin-bottom: 1rem; border-bottom: 1px solid var(--border); padding-bottom: 0.5rem; }
        .history-item { display: flex; justify-content: space-between; align-items: center; padding: 0.5rem 0; border-bottom: 1px dashed var(--border); }
        .history-item span { color: var(--text-main); }
        
        .btn-primary-sm {
          background: var(--primary); color: white; padding: 4px 10px; border-radius: 6px; font-size: 0.8rem; font-weight: 600;
        }

        /* Default hidden area for print */
        .print-area { display: none; }

        @media print {
          /* Hide everything except the receipt */
          body * {
            visibility: hidden;
            margin: 0;
            padding: 0;
          }

          /* Force background colors and remove margins for receipt printing */
          @page { margin: 0; size: auto; }

          body { background: white !important; }

          .print-area {
            display: block;
            position: absolute;
            top: 0;
            left: 0;
            width: 100%;
            visibility: visible;
          }

          .print-area * {
            visibility: visible;
            color: black !important; /* Ensure black ink */
            font-family: 'Courier New', Courier, monospace !important; /* Monospace for receipts */
          }

          .receipt {
            width: 280px; /* Standard 80mm thermal printer approximate width */
            margin: 0;
            padding: 10px;
            font-size: 14px;
            line-height: 1.2;
          }

          .cafe-name {
            text-align: center;
            font-size: 18px;
            font-weight: bold;
            margin-bottom: 5px;
          }

          .receipt-divider {
            text-align: center;
            margin: 5px 0;
            font-size: 12px;
          }

          .receipt-text {
            margin: 2px 0;
            font-size: 12px;
          }

          .receipt-table {
            width: 100%;
            margin: 10px 0;
            border-collapse: collapse;
          }

          .receipt-table th, .receipt-table td {
            font-size: 12px;
            padding: 2px 0;
          }

          .receipt-table th { border-bottom: 1px dashed black; }

          .receipt-total {
            text-align: right;
            font-size: 16px;
            font-weight: bold;
            margin: 5px 0;
          }

          .receipt-footer {
            text-align: center;
            font-size: 12px;
            font-style: italic;
            margin-top: 2px;
          }
        }
      `}</style>
    </div>
  );
};

export default Printer;

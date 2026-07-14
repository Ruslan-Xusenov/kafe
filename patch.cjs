const fs = require('fs');
const file = '/home/kali/Desktop/projects/Django/Kafe/frontend/src/pages/Waiter.jsx';
let content = fs.readFileSync(file, 'utf8');

const reprintFunc = `
  const handleReprintOrder = async (orderId) => {
    try {
      await api.post(\`/orders/\${orderId}/reprint\`);
      alert('Chek printerga yuborildi!');
    } catch (err) {
      alert('Xatolik: ' + (err.response?.data?.error || err.message));
    }
  };
`;

content = content.replace('const stageIncrease =', reprintFunc + '\n  const stageIncrease =');

const reprintButton = `
                      <span style={{ fontWeight: 800, fontSize: '1.05rem', display: 'flex', alignItems: 'center', gap: '10px' }}>
                        {(existingOrder.total_price || 0).toLocaleString()} сум
                        <button onClick={() => handleReprintOrder(existingOrder.id)} style={{ padding: '4px 8px', background: 'var(--primary)', color: 'white', border: 'none', borderRadius: '4px', cursor: 'pointer', fontSize: '0.8rem' }}>
                          🖨️ Chek chiqarish
                        </button>
                      </span>
`;

content = content.replace('<span style={{ fontWeight: 800, fontSize: \'1.05rem\' }}>{(existingOrder.total_price || 0).toLocaleString()} сум</span>', reprintButton);

fs.writeFileSync(file, content);

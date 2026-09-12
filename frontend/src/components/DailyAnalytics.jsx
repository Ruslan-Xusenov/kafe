import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import api from '../store/authStore';
import {
  TrendingUp, ShoppingBag, Wallet, BarChart3,
  ChevronLeft, ChevronRight, RefreshCw, Calendar,
  CreditCard, Smartphone, BookOpen, Banknote, Package,
  Award, QrCode
} from 'lucide-react';

const fmt = (n) => Math.round(n || 0).toLocaleString('uz-UZ');

/* ─── Thin Donut (pure SVG) ─── */
const DonutChart = ({ slices, total }) => {
  const R = 42, cx = 50, cy = 50, sw = 10;
  const circ = 2 * Math.PI * R;
  let off = circ * 0.25; // start from top
  return (
    <svg viewBox="0 0 100 100" width="110" height="110" style={{ flexShrink: 0 }}>
      <circle cx={cx} cy={cy} r={R} fill="none" stroke="rgba(0,0,0,0.06)" strokeWidth={sw} />
      {slices.map((s, i) => {
        const pct = total > 0 ? s.value / total : 0;
        const dash = pct * circ;
        const el = (
          <circle
            key={i}
            cx={cx} cy={cy} r={R}
            fill="none"
            stroke={s.color}
            strokeWidth={sw}
            strokeDasharray={`${dash} ${circ - dash}`}
            strokeDashoffset={off}
            strokeLinecap="butt"
            style={{ transition: 'stroke-dasharray 0.8s cubic-bezier(.4,0,.2,1)' }}
          />
        );
        off -= dash;
        return el;
      })}
      <text x={cx} y={cy - 6} textAnchor="middle" fontSize="7" fill="#9ca3af" fontFamily="inherit" fontWeight="600" letterSpacing="0.5">JAMI</text>
      <text x={cx} y={cy + 7} textAnchor="middle" fontSize="8.5" fill="#374151" fontFamily="inherit" fontWeight="800">{fmt(total)}</text>
    </svg>
  );
};

/* ─── Slim progress bar ─── */
const ProgressBar = ({ value, max, color }) => {
  const pct = max > 0 ? Math.max(3, (value / max) * 100) : 3;
  return (
    <div style={{ background: 'rgba(0,0,0,0.06)', borderRadius: 99, height: 5, overflow: 'hidden', marginTop: 6 }}>
      <motion.div
        initial={{ width: 0 }}
        animate={{ width: `${pct}%` }}
        transition={{ duration: 0.7, ease: 'easeOut' }}
        style={{ height: '100%', borderRadius: 99, background: color }}
      />
    </div>
  );
};

/* ─── Date helpers ─── */
const todayStr = () => new Date().toISOString().slice(0, 10);
const addDays = (s, n) => { const d = new Date(s); d.setDate(d.getDate() + n); return d.toISOString().slice(0, 10); };
const fmtDate = (s) => {
  const months = ['Yan','Fev','Mar','Apr','May','Iyn','Iyl','Avg','Sen','Okt','Noy','Dek'];
  const d = new Date(s);
  return `${d.getDate()} ${months[d.getMonth()]} ${d.getFullYear()}`;
};

const PRODUCT_COLORS = ['#f97316','#6366f1','#10b981','#3b82f6','#ec4899','#f59e0b','#06b6d4','#8b5cf6'];

const DailyAnalytics = ({ refreshTrigger, onDateChange }) => {
  const [date, setDate]     = useState(todayStr());
  const [report, setReport] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetch = useCallback(async (d) => {
    setLoading(true);
    try {
      const res = await api.get(`/finance/daily-report?date=${d}`);
      setReport(res.data);
    } catch (e) { console.error(e); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { fetch(date); }, [date, fetch, refreshTrigger]);



  const isToday = date === todayStr();
  const payments = report ? [
    { label: 'Naqd pul',    value: report.cash_revenue,    color: '#10b981', icon: <Banknote size={13}/> },
    { label: 'Karta',       value: report.card_revenue,    color: '#3b82f6', icon: <CreditCard size={13}/> },
    { label: 'Click/Payme', value: report.click_revenue,   color: '#8b5cf6', icon: <Smartphone size={13}/> },
    { label: 'Nasiya',      value: report.nasiya_revenue,  color: '#f59e0b', icon: <BookOpen size={13}/> },
    { label: 'QR Kod',      value: report.qr_revenue,      color: '#ec4899', icon: <QrCode size={13}/> },
  ] : [];

  const kpis = report ? [
    { label: "Daromad",     val: fmt(report.revenue) + " so'm",    color: '#10b981', bg: 'rgba(16,185,129,0.08)',  border: 'rgba(16,185,129,0.18)', icon: <TrendingUp size={16}/> },
    { label: "Заказlar", val: report.orders_count,               color: '#6366f1', bg: 'rgba(99,102,241,0.08)',  border: 'rgba(99,102,241,0.18)', icon: <ShoppingBag size={16}/> },
    { label: "Opr. Xarajatlar",  val: fmt(report.expenses) + " so'm",   color: '#ef4444', bg: 'rgba(239,68,68,0.08)',   border: 'rgba(239,68,68,0.18)',  icon: <Wallet size={16}/> },
    { label: "Sof foyda",   val: fmt(report.net_profit) + " so'm", color: '#f97316', bg: 'rgba(249,115,22,0.08)',  border: 'rgba(249,115,22,0.18)', icon: <Award, QrCode size={16}/> },
  ] : [];

  return (
    <div style={{ marginBottom: '1.5rem', fontFamily: 'var(--font-body, inherit)' }}>

      {/* ── Header ── */}
      <div style={{ display:'flex', alignItems:'center', justifyContent:'space-between', marginBottom:'1rem', flexWrap:'wrap', gap:'0.5rem' }}>
        <div style={{ display:'flex', alignItems:'center', gap:'0.5rem' }}>
          <div style={{ width:32, height:32, borderRadius:10, background:'rgba(249,115,22,0.1)', display:'flex', alignItems:'center', justifyContent:'center', color:'#f97316' }}>
            <BarChart3 size={16}/>
          </div>
          <span style={{ fontSize:'0.8rem', fontWeight:700, textTransform:'uppercase', letterSpacing:'0.08em', color:'var(--text-secondary,#6b7280)' }}>Kunlik tahlil</span>
        </div>

        <div style={{ display:'flex', alignItems:'center', gap:'0.35rem' }}>
          <button className="da2-nav" onClick={() => {
            const newD = addDays(date,-1);
            setDate(newD);
            if (onDateChange) onDateChange(newD);
          }}><ChevronLeft size={14}/></button>

          <label style={{ position:'relative', cursor:'pointer' }}>
            <div className="da2-pill">
              <Calendar size={12}/>
              <span>{isToday ? 'Bugun' : fmtDate(date)}</span>
            </div>
            <input type="date" value={date} max={todayStr()} onChange={e=> {
              setDate(e.target.value);
              if (onDateChange) onDateChange(e.target.value);
            }}
              style={{ position:'absolute', opacity:0, width:0, height:0, pointerEvents:'none' }}/>
          </label>

          <button className="da2-nav" onClick={() => {
            const newD = addDays(date,1);
            setDate(newD);
            if (onDateChange) onDateChange(newD);
          }} disabled={isToday} style={{ opacity: isToday ? 0.3:1 }}>
            <ChevronRight size={14}/>
          </button>
          <button className="da2-nav" onClick={() => fetch(date)}>
            <RefreshCw size={13} style={{ animation: loading ? 'da2spin 0.8s linear infinite' : 'none' }}/>
          </button>
        </div>
      </div>

      <AnimatePresence mode="wait">
        {loading ? (
          <motion.div key="sk" initial={{opacity:0}} animate={{opacity:1}} exit={{opacity:0}}>
            <div style={{ display:'grid', gridTemplateColumns:'repeat(auto-fill,minmax(180px,1fr))', gap:'0.6rem', marginBottom:'0.75rem' }}>
              {[...Array(4)].map((_,i)=><div key={i} style={{ height:72, borderRadius:14, background:'rgba(0,0,0,0.04)', animation:'da2shimmer 1.5s ease infinite' }}/>)}
            </div>
            <div style={{ height:160, borderRadius:14, background:'rgba(0,0,0,0.04)', animation:'da2shimmer 1.5s ease infinite' }}/>
          </motion.div>
        ) : report ? (
          <motion.div key={date} initial={{opacity:0,y:8}} animate={{opacity:1,y:0}} exit={{opacity:0}} transition={{duration:0.25}}>

            {/* ── KPI Cards ── */}
            <div style={{ display:'grid', gridTemplateColumns:'repeat(auto-fill,minmax(180px,1fr))', gap:'0.6rem', marginBottom:'0.75rem' }}>
              {kpis.map((k,i) => (
                <motion.div key={i} initial={{opacity:0,y:10}} animate={{opacity:1,y:0}} transition={{delay:i*0.05}}
                  style={{
                    background: k.bg,
                    border: `1px solid ${k.border}`,
                    borderRadius:14,
                    padding:'0.85rem 1rem',
                    display:'flex', alignItems:'center', gap:'0.7rem',
                    transition:'transform 0.15s',
                    cursor:'default',
                  }}
                  whileHover={{ y:-2 }}
                >
                  <div style={{ width:36, height:36, borderRadius:10, background:`${k.color}18`, display:'flex', alignItems:'center', justifyContent:'center', color:k.color, flexShrink:0 }}>
                    {k.icon}
                  </div>
                  <div>
                    <div style={{ fontSize:'1.05rem', fontWeight:800, color:k.color, lineHeight:1.1 }}>{k.val}</div>
                    <div style={{ fontSize:'0.68rem', color:'var(--text-muted,#9ca3af)', fontWeight:600, textTransform:'uppercase', letterSpacing:'0.04em', marginTop:2 }}>{k.label}</div>
                  </div>
                </motion.div>
              ))}
            </div>

            {report.revenue > 0 && (
              <>
                {/* ── Payments + Metrics ── */}
                <div style={{ display:'grid', gridTemplateColumns:'1fr 1fr', gap:'0.6rem', marginBottom:'0.75rem' }}>

                  {/* Payment breakdown */}
                  <div style={{ background:'var(--bg-card,#fff)', border:'1px solid var(--border,#e5e7eb)', borderRadius:14, padding:'1rem' }}>
                    <div style={{ fontSize:'0.7rem', fontWeight:700, textTransform:'uppercase', letterSpacing:'0.06em', color:'var(--text-secondary,#6b7280)', marginBottom:'0.85rem', display:'flex', alignItems:'center', gap:'0.35rem' }}>
                      <CreditCard size={13}/> Оплата turlari
                    </div>
                    <div style={{ display:'flex', alignItems:'center', gap:'1rem' }}>
                      <DonutChart slices={payments.filter(p=>p.value>0)} total={report.revenue}/>
                      <div style={{ flex:1, minWidth:0 }}>
                        {payments.map(p => (
                          <div key={p.label} style={{ marginBottom:'0.55rem' }}>
                            <div style={{ display:'flex', justifyContent:'space-between', alignItems:'center' }}>
                              <div style={{ display:'flex', alignItems:'center', gap:'0.3rem', color:'var(--text-muted,#9ca3af)', fontSize:'0.72rem' }}>
                                <span style={{ color:p.color }}>{p.icon}</span>{p.label}
                              </div>
                              <span style={{ fontSize:'0.78rem', fontWeight:700, color:p.color }}>
                                {report.revenue > 0 ? Math.round((p.value/report.revenue)*100) : 0}%
                              </span>
                            </div>
                            <ProgressBar value={p.value} max={report.revenue} color={p.color}/>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>

                  {/* Metrics */}
                  <div style={{ background:'var(--bg-card,#fff)', border:'1px solid var(--border,#e5e7eb)', borderRadius:14, padding:'1rem' }}>
                    <div style={{ fontSize:'0.7rem', fontWeight:700, textTransform:'uppercase', letterSpacing:'0.06em', color:'var(--text-secondary,#6b7280)', marginBottom:'0.85rem', display:'flex', alignItems:'center', gap:'0.35rem' }}>
                      <Package size={13}/> Ko'rsatkichlar
                    </div>
                    {[
                      { label: "O'rtacha заказ", val: fmt(report.orders_count > 0 ? report.revenue/report.orders_count : 0) + " so'm" },
                      { label: "Foyda ulushi",       val: (report.revenue>0?Math.round((report.net_profit/report.revenue)*100):0)+"%", color:'#10b981' },
                      { label: "Opr. Xarajat ulushi",     val: (report.revenue>0?Math.round((report.expenses/report.revenue)*100):0)+"%",   color:'#ef4444' },
                      { label: "Naqd ulushi",        val: (report.revenue>0?Math.round((report.cash_revenue/report.revenue)*100):0)+"%",color:'#10b981' },
                    ].map((m,i) => (
                      <div key={i} style={{ display:'flex', justifyContent:'space-between', alignItems:'center', padding:'0.4rem 0.6rem', borderRadius:8, background:'rgba(0,0,0,0.025)', marginBottom:'0.35rem' }}>
                        <span style={{ fontSize:'0.75rem', color:'var(--text-muted,#9ca3af)' }}>{m.label}</span>
                        <span style={{ fontSize:'0.8rem', fontWeight:700, color: m.color || 'var(--text-main,#111827)' }}>{m.val}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* ── Top Products ── */}
                {report.top_products?.length > 0 && (
                  <div style={{ background:'var(--bg-card,#fff)', border:'1px solid var(--border,#e5e7eb)', borderRadius:14, padding:'1rem' }}>
                    <div style={{ fontSize:'0.7rem', fontWeight:700, textTransform:'uppercase', letterSpacing:'0.06em', color:'var(--text-secondary,#6b7280)', marginBottom:'1rem', display:'flex', alignItems:'center', gap:'0.35rem' }}>
                      <TrendingUp size={13}/> Eng ko'p sotilgan mahsulotlar
                    </div>

                    {/* Top 3 highlight chips */}
                    {report.top_products.length >= 1 && (
                      <div style={{ display:'flex', gap:'0.5rem', marginBottom:'1rem', flexWrap:'wrap' }}>
                        {report.top_products.slice(0,3).map((p,i) => (
                          <div key={i} style={{
                            display:'flex', alignItems:'center', gap:'0.4rem',
                            padding:'0.4rem 0.75rem', borderRadius:99,
                            background: ['rgba(249,115,22,0.1)','rgba(99,102,241,0.1)','rgba(16,185,129,0.1)'][i],
                            border: `1px solid ${['rgba(249,115,22,0.2)','rgba(99,102,241,0.2)','rgba(16,185,129,0.2)'][i]}`,
                          }}>
                            <span style={{ fontSize:'0.9rem' }}>{['🥇','🥈','🥉'][i]}</span>
                            <span style={{ fontSize:'0.78rem', fontWeight:700, color:'var(--text-main,#111827)', maxWidth:90, overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>{p.product_name}</span>
                            <span style={{ fontSize:'0.7rem', color:['#f97316','#6366f1','#10b981'][i], fontWeight:600 }}>{fmt(p.total_amount)} so'm</span>
                          </div>
                        ))}
                      </div>
                    )}

                    {/* Bar list */}
                    <div>
                      {report.top_products.map((p,i) => {
                        const maxAmt = report.top_products[0].total_amount || 1;
                        const pct = Math.max(3, (p.total_amount/maxAmt)*100);
                        return (
                          <div key={i} style={{ marginBottom:'0.65rem' }}>
                            <div style={{ display:'flex', justifyContent:'space-between', marginBottom:4 }}>
                              <span style={{ fontSize:'0.78rem', fontWeight:600, color:'var(--text-main,#111827)', maxWidth:'60%', overflow:'hidden', textOverflow:'ellipsis', whiteSpace:'nowrap' }}>
                                {p.product_name}
                              </span>
                              <span style={{ fontSize:'0.72rem', color:'var(--text-muted,#9ca3af)' }}>
                                {fmt(p.total_amount)} so'm &nbsp;·&nbsp; ×{p.quantity%1===0?p.quantity:p.quantity.toFixed(2)}
                              </span>
                            </div>
                            <div style={{ background:'rgba(0,0,0,0.05)', borderRadius:99, height:6, overflow:'hidden' }}>
                              <motion.div
                                initial={{ width:0 }}
                                animate={{ width:`${pct}%` }}
                                transition={{ duration:0.6, delay:i*0.04, ease:'easeOut' }}
                                style={{ height:'100%', borderRadius:99, background:PRODUCT_COLORS[i%PRODUCT_COLORS.length], opacity:0.85 }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>
                )}
              </>
            )}

            {/* ── Expenses List ── */}
            {report.expense_list?.length > 0 && (
              <div style={{ background:'var(--bg-card,#fff)', border:'1px solid var(--border,#e5e7eb)', borderRadius:14, padding:'1rem', marginTop:'0.75rem' }}>
                <div style={{ fontSize:'0.7rem', fontWeight:700, textTransform:'uppercase', letterSpacing:'0.06em', color:'var(--text-secondary,#6b7280)', marginBottom:'1rem', display:'flex', alignItems:'center', gap:'0.35rem' }}>
                  <Wallet size={13}/> Shu kundagi xarajatlar tarixi
                </div>
                
                <div className="orders-table-wrapper" style={{ maxHeight: '300px', overflowY: 'auto' }}>
                  <table className="admin-table">
                    <thead>
                      <tr>
                        <th>Sana</th>
                        <th>Kategoriya</th>
                        <th>Summa</th>
                        <th>Izoh</th>
                      </tr>
                    </thead>
                    <tbody>
                      {report.expense_list.map((exp) => (
                        <tr key={exp.id}>
                          <td>{new Date(exp.created_at).toLocaleString('uz-UZ')}</td>
                          <td>{exp.category}</td>
                          <td style={{ color: '#ef4444', fontWeight: 600 }}>-{fmt(exp.amount)} so'm</td>
                          <td>{exp.description || '-'}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {report.revenue === 0 && (
              <div style={{ textAlign:'center', padding:'2.5rem 1rem', color:'var(--text-muted,#9ca3af)' }}>
                <BarChart3 size={32} style={{ opacity:0.2, margin:'0 auto 0.5rem' }}/>
                <p style={{ fontSize:'0.85rem' }}>{fmtDate(date)} kuni заказ topilmadi</p>
              </div>
            )}

          </motion.div>
        ) : null}
      </AnimatePresence>

      <style>{`
        .da2-nav {
          width:30px; height:30px; border-radius:8px;
          background:var(--bg-surface,#f9fafb);
          border:1px solid var(--border,#e5e7eb);
          display:flex; align-items:center; justify-content:center;
          cursor:pointer; color:var(--text-main,#374151);
          transition:border-color 0.15s, color 0.15s;
        }
        .da2-nav:hover { border-color:#f97316; color:#f97316; }

        .da2-pill {
          display:inline-flex; align-items:center; gap:0.35rem;
          background:var(--bg-surface,#f9fafb);
          border:1px solid var(--border,#e5e7eb);
          border-radius:99px; padding:0.3rem 0.75rem;
          font-size:0.78rem; font-weight:700; color:#f97316;
          cursor:pointer; white-space:nowrap;
          transition:background 0.15s;
        }
        .da2-pill:hover { background:rgba(249,115,22,0.06); }

        @keyframes da2spin { to { transform:rotate(360deg); } }
        @keyframes da2shimmer {
          0%   { opacity:1; }
          50%  { opacity:0.5; }
          100% { opacity:1; }
        }

        @media(max-width:600px){
          .da2-mid { grid-template-columns:1fr !important; }
        }
      `}</style>
    </div>
  );
};

export default DailyAnalytics;

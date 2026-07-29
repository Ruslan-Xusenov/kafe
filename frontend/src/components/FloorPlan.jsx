import React, { useState, useRef, useEffect, useCallback } from 'react';
import { motion } from 'framer-motion';
import { Map, Save, RotateCw, Maximize2, Circle, Square, RectangleHorizontal, ZoomIn, ZoomOut, Move, GripVertical } from 'lucide-react';
import api from '../store/authStore';

const SHAPES = {
  round: { label: 'Dumaloq', icon: Circle },
  square: { label: 'Kvadrat', icon: Square },
  rectangle: { label: "To'rtburchak", icon: RectangleHorizontal },
};

const FloorPlan = ({ tables, onTableSelect, isAdmin, onTablesUpdate }) => {
  const svgRef = useRef(null);
  const containerRef = useRef(null);

  const [editMode, setEditMode] = useState(false);
  const [localTables, setLocalTables] = useState([]);
  const [draggingId, setDraggingId] = useState(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [selectedEditTable, setSelectedEditTable] = useState(null);
  const [zoom, setZoom] = useState(1);
  const [saving, setSaving] = useState(false);
  const [activeFloor, setActiveFloor] = useState(1);

  useEffect(() => {
    setLocalTables(tables.map(t => ({
      ...t,
      pos_x: t.pos_x || Math.random() * 600 + 50,
      pos_y: t.pos_y || Math.random() * 400 + 50,
      shape: t.shape || 'round',
      width: t.width || 80,
      height: t.height || 80,
      floor: t.floor || 1,
      rotation: t.rotation || 0,
    })));
  }, [tables]);

  const floors = [...new Set(localTables.map(t => t.floor))].sort();
  const visibleTables = localTables.filter(t => t.floor === activeFloor);

  const getTableColor = (table) => {
    if (table.status === 'free') return { fill: '#10b981', stroke: '#059669', text: '#ffffff' };
    return { fill: '#f97316', stroke: '#ea580c', text: '#ffffff' };
  };

  const getSVGCoords = (e) => {
    const svg = svgRef.current;
    if (!svg) return { x: 0, y: 0 };
    const CTM = svg.getScreenCTM();
    const clientX = e.touches ? e.touches[0].clientX : e.clientX;
    const clientY = e.touches ? e.touches[0].clientY : e.clientY;
    return {
      x: (clientX - CTM.e) / CTM.a,
      y: (clientY - CTM.f) / CTM.d,
    };
  };

  const handlePointerDown = (e, table) => {
    if (!editMode) {
      onTableSelect(table);
      return;
    }
    e.preventDefault();
    e.stopPropagation();
    const coords = getSVGCoords(e);
    setDraggingId(table.id);
    setDragOffset({ x: coords.x - table.pos_x, y: coords.y - table.pos_y });
    setSelectedEditTable(table.id);
  };

  const handlePointerMove = useCallback((e) => {
    if (!draggingId) return;
    e.preventDefault();
    const coords = getSVGCoords(e);
    setLocalTables(prev => prev.map(t =>
      t.id === draggingId
        ? { ...t, pos_x: Math.max(0, coords.x - dragOffset.x), pos_y: Math.max(0, coords.y - dragOffset.y) }
        : t
    ));
  }, [draggingId, dragOffset]);

  const handlePointerUp = useCallback(() => {
    setDraggingId(null);
  }, []);

  useEffect(() => {
    if (!editMode) return;
    const svg = svgRef.current;
    if (!svg) return;
    svg.addEventListener('mousemove', handlePointerMove);
    svg.addEventListener('mouseup', handlePointerUp);
    svg.addEventListener('touchmove', handlePointerMove, { passive: false });
    svg.addEventListener('touchend', handlePointerUp);
    return () => {
      svg.removeEventListener('mousemove', handlePointerMove);
      svg.removeEventListener('mouseup', handlePointerUp);
      svg.removeEventListener('touchmove', handlePointerMove);
      svg.removeEventListener('touchend', handlePointerUp);
    };
  }, [editMode, handlePointerMove, handlePointerUp]);

  const handleSaveLayout = async () => {
    setSaving(true);
    try {
      await api.put('/tables/layout/batch', { tables: localTables });
      if (onTablesUpdate) onTablesUpdate();
      setEditMode(false);
    } catch (err) {
      alert("Xatolik: " + (err.response?.data?.error || err.message));
    } finally {
      setSaving(false);
    }
  };

  const updateTableProp = (id, key, value) => {
    setLocalTables(prev => prev.map(t =>
      t.id === id ? { ...t, [key]: value } : t
    ));
  };

  const renderTable = (table) => {
    const colors = getTableColor(table);
    const { pos_x, pos_y, width, height, shape, rotation, name, capacity, status } = table;
    const isSelected = selectedEditTable === table.id;
    const isDragging = draggingId === table.id;

    return (
      <g
        key={table.id}
        transform={`translate(${pos_x}, ${pos_y}) rotate(${rotation}, ${width/2}, ${height/2})`}
        style={{ cursor: editMode ? (isDragging ? 'grabbing' : 'grab') : 'pointer' }}
        onMouseDown={(e) => handlePointerDown(e, table)}
        onTouchStart={(e) => handlePointerDown(e, table)}
      >
        {/* Shadow */}
        {shape === 'round' ? (
          <ellipse
            cx={width/2} cy={height/2+3} rx={width/2} ry={height/2}
            fill="rgba(0,0,0,0.08)" filter="url(#shadow)"
          />
        ) : (
          <rect
            x={0} y={3} width={width} height={height}
            rx={shape === 'square' ? 8 : 12}
            fill="rgba(0,0,0,0.08)" filter="url(#shadow)"
          />
        )}

        {/* Table Shape */}
        {shape === 'round' ? (
          <ellipse
            cx={width/2} cy={height/2} rx={width/2} ry={height/2}
            fill={colors.fill}
            stroke={isSelected && editMode ? '#3b82f6' : colors.stroke}
            strokeWidth={isSelected && editMode ? 3 : 1.5}
            opacity={isDragging ? 0.8 : 1}
          />
        ) : (
          <rect
            x={0} y={0} width={width} height={height}
            rx={shape === 'square' ? 8 : 12}
            fill={colors.fill}
            stroke={isSelected && editMode ? '#3b82f6' : colors.stroke}
            strokeWidth={isSelected && editMode ? 3 : 1.5}
            opacity={isDragging ? 0.8 : 1}
          />
        )}

        {/* Inner gradient overlay */}
        {shape === 'round' ? (
          <ellipse
            cx={width/2} cy={height/2 - 4} rx={width/2 - 6} ry={height/2 - 8}
            fill="rgba(255,255,255,0.15)"
          />
        ) : (
          <rect
            x={4} y={2} width={width - 8} height={height/2 - 4}
            rx={6}
            fill="rgba(255,255,255,0.15)"
          />
        )}

        {/* Table Name */}
        <text
          x={width/2} y={height/2 - 6}
          textAnchor="middle"
          fill={colors.text}
          fontSize="15"
          fontWeight="800"
          fontFamily="'Plus Jakarta Sans', sans-serif"
        >
          {name}
        </text>

        {/* Capacity */}
        <text
          x={width/2} y={height/2 + 12}
          textAnchor="middle"
          fill="rgba(255,255,255,0.85)"
          fontSize="10"
          fontWeight="600"
          fontFamily="'Plus Jakarta Sans', sans-serif"
        >
          {capacity || 4} kishi
        </text>

        {/* Status dot */}
        <circle
          cx={width - 8} cy={8}
          r={5}
          fill={status === 'free' ? '#34d399' : '#fbbf24'}
          stroke="#fff"
          strokeWidth={1.5}
        />

        {/* Edit mode drag handle */}
        {editMode && isSelected && (
          <g>
            <circle cx={width/2} cy={-8} r={6} fill="#3b82f6" stroke="#fff" strokeWidth={1.5} />
            <text x={width/2} y={-4} textAnchor="middle" fill="#fff" fontSize="8" fontWeight="bold">✥</text>
          </g>
        )}
      </g>
    );
  };

  return (
    <div className="floor-plan-wrapper">
      {/* Toolbar */}
      <div className="floor-plan-toolbar glass">
        <div className="fp-toolbar-left">
          <Map size={20} style={{ color: 'var(--primary)' }} />
          <span className="fp-title">Zal Xaritasi</span>
          {floors.length > 1 && (
            <div className="fp-floor-tabs">
              {floors.map(f => (
                <button
                  key={f}
                  className={`fp-floor-btn ${activeFloor === f ? 'active' : ''}`}
                  onClick={() => setActiveFloor(f)}
                >
                  {f}-qavat
                </button>
              ))}
            </div>
          )}
        </div>
        <div className="fp-toolbar-right">
          <div className="fp-legend">
            <span className="fp-legend-item"><span className="fp-dot free"></span>Bo'sh</span>
            <span className="fp-legend-item"><span className="fp-dot occupied"></span>Band</span>
          </div>
          <button className="fp-zoom-btn" onClick={() => setZoom(z => Math.min(z + 0.15, 2.5))} title="Kattalashtirish">
            <ZoomIn size={16} />
          </button>
          <button className="fp-zoom-btn" onClick={() => setZoom(z => Math.max(z - 0.15, 0.5))} title="Kichiklashtirish">
            <ZoomOut size={16} />
          </button>
          {isAdmin && (
            <>
              {editMode ? (
                <>
                  <button className="fp-save-btn" onClick={handleSaveLayout} disabled={saving}>
                    <Save size={16} />
                    {saving ? 'Saqlanmoqda...' : 'Saqlash'}
                  </button>
                  <button className="fp-cancel-btn" onClick={() => { setEditMode(false); setLocalTables(tables); setSelectedEditTable(null); }}>
                    Bekor
                  </button>
                </>
              ) : (
                <button className="fp-edit-btn" onClick={() => setEditMode(true)}>
                  <Move size={16} />
                  Tahrirlash
                </button>
              )}
            </>
          )}
        </div>
      </div>

      {/* Edit Panel for selected table */}
      {editMode && selectedEditTable && (() => {
        const t = localTables.find(x => x.id === selectedEditTable);
        if (!t) return null;
        return (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="fp-edit-panel glass"
          >
            <span className="fp-edit-label">Stol: <strong>{t.name}</strong></span>
            <div className="fp-edit-controls">
              <div className="fp-control-group">
                <label>Shakl:</label>
                <div className="fp-shape-btns">
                  {Object.entries(SHAPES).map(([key, { label, icon: Icon }]) => (
                    <button
                      key={key}
                      className={`fp-shape-btn ${t.shape === key ? 'active' : ''}`}
                      onClick={() => updateTableProp(t.id, 'shape', key)}
                      title={label}
                    >
                      <Icon size={14} />
                    </button>
                  ))}
                </div>
              </div>
              <div className="fp-control-group">
                <label>Kattalik:</label>
                <input type="range" min="50" max="150" value={t.width}
                  onChange={(e) => {
                    const v = Number(e.target.value);
                    updateTableProp(t.id, 'width', v);
                    if (t.shape !== 'rectangle') updateTableProp(t.id, 'height', v);
                  }}
                />
              </div>
              {t.shape === 'rectangle' && (
                <div className="fp-control-group">
                  <label>Balandlik:</label>
                  <input type="range" min="40" max="150" value={t.height}
                    onChange={(e) => updateTableProp(t.id, 'height', Number(e.target.value))}
                  />
                </div>
              )}
              <div className="fp-control-group">
                <label>Burilish:</label>
                <input type="range" min="0" max="360" step="15" value={t.rotation}
                  onChange={(e) => updateTableProp(t.id, 'rotation', Number(e.target.value))}
                />
                <span className="fp-rotation-label">{t.rotation}°</span>
              </div>
            </div>
          </motion.div>
        );
      })()}

      {/* SVG Canvas */}
      <div className="floor-plan-canvas-wrap" ref={containerRef}>
        <svg
          ref={svgRef}
          className="floor-plan-svg"
          viewBox={`0 0 ${900 / zoom} ${600 / zoom}`}
          preserveAspectRatio="xMidYMid meet"
          style={{ touchAction: editMode ? 'none' : 'auto' }}
        >
          <defs>
            <filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">
              <feDropShadow dx="0" dy="2" stdDeviation="3" floodOpacity="0.15" />
            </filter>
            <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
              <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(0,0,0,0.04)" strokeWidth="0.5" />
            </pattern>
          </defs>

          {/* Background grid */}
          <rect width="100%" height="100%" fill="url(#grid)" />

          {/* Render tables */}
          {visibleTables.map(renderTable)}

          {/* Empty state */}
          {visibleTables.length === 0 && (
            <text x="50%" y="50%" textAnchor="middle" fill="#999" fontSize="16" fontFamily="'Plus Jakarta Sans', sans-serif">
              Bu qavatda stollar yo'q
            </text>
          )}
        </svg>
      </div>
    </div>
  );
};

export default FloorPlan;

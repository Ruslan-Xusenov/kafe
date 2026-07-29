-- =====================================================
-- Migration 007: Professional Refund System
-- Full refund workflow with approval and tracking
-- =====================================================

CREATE TABLE IF NOT EXISTS refunds (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    amount DECIMAL(12,2) NOT NULL,
    reason VARCHAR(50) NOT NULL,
    reason_detail TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    refund_method VARCHAR(20),
    money_returned BOOLEAN DEFAULT FALSE,
    requested_by INTEGER REFERENCES users(id),
    requested_by_name VARCHAR(255),
    approved_by INTEGER REFERENCES users(id),
    approved_by_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX IF NOT EXISTS idx_refunds_order ON refunds(order_id);
CREATE INDEX IF NOT EXISTS idx_refunds_status ON refunds(status);
CREATE INDEX IF NOT EXISTS idx_refunds_created ON refunds(created_at DESC);

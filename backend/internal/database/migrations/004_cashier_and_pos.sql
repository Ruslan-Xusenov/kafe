-- =====================================================
-- Migration 004: Cashier POS System
-- Adds cashier role, shift management, cash operations
-- =====================================================

-- 1. Add cashier role to user_role enum
DO $$
BEGIN
    ALTER TYPE user_role ADD VALUE IF NOT EXISTS 'cashier';
EXCEPTION WHEN duplicate_object THEN
    NULL;
END$$;

-- 2. Cashier Shifts Table
CREATE TABLE IF NOT EXISTS cashier_shifts (
    id SERIAL PRIMARY KEY,
    cashier_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    opened_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    opening_cash DECIMAL(12,2) NOT NULL DEFAULT 0,
    closing_cash DECIMAL(12,2),
    expected_cash DECIMAL(12,2),
    total_sales DECIMAL(12,2) DEFAULT 0,
    total_cash_sales DECIMAL(12,2) DEFAULT 0,
    total_card_sales DECIMAL(12,2) DEFAULT 0,
    total_click_sales DECIMAL(12,2) DEFAULT 0,
    total_nasiya_sales DECIMAL(12,2) DEFAULT 0,
    total_orders INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'open' CHECK (status IN ('open', 'closed')),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_cashier ON cashier_shifts(cashier_id);
CREATE INDEX IF NOT EXISTS idx_cashier_shifts_status ON cashier_shifts(status);

-- 3. Cash In/Out Operations
CREATE TABLE IF NOT EXISTS cash_operations (
    id SERIAL PRIMARY KEY,
    shift_id INTEGER REFERENCES cashier_shifts(id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('cash_in', 'cash_out')),
    amount DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    reason TEXT NOT NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_cash_operations_shift ON cash_operations(shift_id);

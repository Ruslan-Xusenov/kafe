-- =====================================================
-- Migration 006: Debt/Nasiya System
-- Customer debt tracking with partial payments
-- =====================================================

-- Debtors Table
CREATE TABLE IF NOT EXISTS debtors (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    total_debt DECIMAL(12,2) NOT NULL DEFAULT 0,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_debtors_phone ON debtors(phone);

-- Debt Records (both debt entries and payments)
CREATE TABLE IF NOT EXISTS debt_records (
    id SERIAL PRIMARY KEY,
    debtor_id INTEGER REFERENCES debtors(id) ON DELETE CASCADE,
    order_id INTEGER REFERENCES orders(id) ON DELETE SET NULL,
    amount DECIMAL(12,2) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('debt', 'payment')),
    payment_method VARCHAR(20),
    description TEXT,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_debt_records_debtor ON debt_records(debtor_id);
CREATE INDEX IF NOT EXISTS idx_debt_records_type ON debt_records(type);

-- Trigger to update debtors.updated_at
CREATE TRIGGER update_debtors_updated_at 
BEFORE UPDATE ON debtors 
FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

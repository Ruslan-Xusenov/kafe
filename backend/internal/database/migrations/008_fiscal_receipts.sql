-- Fiscal Receipts (Fiskal cheklar) uchun jadval
CREATE TABLE IF NOT EXISTS fiscal_receipts (
    id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(id) ON DELETE CASCADE,
    receipt_number VARCHAR(50) NOT NULL,
    fiscal_sign VARCHAR(100),
    total_amount DECIMAL(12,2) NOT NULL,
    vat_rate DECIMAL(5,2) DEFAULT 12.00,
    vat_amount DECIMAL(12,2) NOT NULL,
    subtotal DECIMAL(12,2) NOT NULL,
    payment_method VARCHAR(20),
    inn VARCHAR(20),
    company_name VARCHAR(255),
    cashier_name VARCHAR(255),
    status VARCHAR(20) DEFAULT 'local',  -- local, sent, confirmed, error
    ofd_response JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fiscal_receipts_order ON fiscal_receipts(order_id);
CREATE INDEX IF NOT EXISTS idx_fiscal_receipts_status ON fiscal_receipts(status);
CREATE INDEX IF NOT EXISTS idx_fiscal_receipts_number ON fiscal_receipts(receipt_number);

-- Fiskal sozlamalar
INSERT INTO settings (key, value) VALUES ('fiscal_enabled', 'false') ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value) VALUES ('fiscal_vat_rate', '12') ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value) VALUES ('fiscal_inn', '') ON CONFLICT (key) DO NOTHING;
INSERT INTO settings (key, value) VALUES ('fiscal_company_name', '') ON CONFLICT (key) DO NOTHING;

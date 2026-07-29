-- =====================================================
-- Migration 005: Mixed Payment System
-- Migrates existing payment data to order_payments table
-- =====================================================

-- Migrate existing single-method payments to order_payments table
-- Only for orders that have a payment_method set but no entries in order_payments yet
INSERT INTO order_payments (order_id, method, amount, created_at)
SELECT o.id, o.payment_method, o.total_price, COALESCE(o.updated_at, o.created_at)
FROM orders o
WHERE o.payment_method IS NOT NULL 
  AND o.payment_method != ''
  AND o.status = 'delivered'
  AND NOT EXISTS (SELECT 1 FROM order_payments op WHERE op.order_id = o.id)
ON CONFLICT DO NOTHING;

package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type PaymentRepository struct {
	db *sqlx.DB
}

func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// AddPayments inserts multiple payment records for an order within a transaction
func (r *PaymentRepository) AddPayments(orderID int, payments []models.PaymentInput) error {
	if len(payments) == 0 {
		return fmt.Errorf("to'lov ma'lumotlari bo'sh")
	}

	validMethods := map[string]bool{"cash": true, "card": true, "click": true, "nasiya": true}
	for _, p := range payments {
		if !validMethods[p.Method] {
			return fmt.Errorf("noto'g'ri to'lov usuli: %s", p.Method)
		}
		if p.Amount <= 0 {
			return fmt.Errorf("to'lov summasi musbat bo'lishi kerak")
		}
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete any existing payments for this order (in case of re-payment)
	_, _ = tx.Exec(`DELETE FROM order_payments WHERE order_id = $1`, orderID)

	for _, p := range payments {
		_, err := tx.Exec(`
			INSERT INTO order_payments (order_id, method, amount)
			VALUES ($1, $2, $3)
		`, orderID, p.Method, p.Amount)
		if err != nil {
			return fmt.Errorf("to'lov saqlashda xatolik: %w", err)
		}
	}

	// Update the orders.payment_method field for backward compatibility
	primaryMethod := payments[0].Method
	if len(payments) > 1 {
		primaryMethod = "mixed"
	}
	_, err = tx.Exec(`UPDATE orders SET payment_method = $1, updated_at = NOW() WHERE id = $2`, primaryMethod, orderID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetPaymentsByOrderID returns all payment records for an order
func (r *PaymentRepository) GetPaymentsByOrderID(orderID int) ([]models.OrderPayment, error) {
	var payments []models.OrderPayment
	err := r.db.Select(&payments, `
		SELECT id, order_id, method, amount, created_at
		FROM order_payments
		WHERE order_id = $1
		ORDER BY created_at
	`, orderID)
	return payments, err
}

// GetPaymentSummary returns aggregated payment breakdown for an order
func (r *PaymentRepository) GetPaymentSummary(orderID int) (map[string]float64, error) {
	type row struct {
		Method string  `db:"method"`
		Total  float64 `db:"total"`
	}
	var rows []row
	err := r.db.Select(&rows, `
		SELECT method, SUM(amount) as total
		FROM order_payments
		WHERE order_id = $1
		GROUP BY method
	`, orderID)
	if err != nil {
		return nil, err
	}

	summary := make(map[string]float64)
	for _, r := range rows {
		summary[r.Method] = r.Total
	}
	return summary, nil
}

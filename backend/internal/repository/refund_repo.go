package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type RefundRepository struct {
	db *sqlx.DB
}

func NewRefundRepository(db *sqlx.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

// CreateRefund creates a new refund request
func (r *RefundRepository) CreateRefund(refund *models.Refund) error {
	err := r.db.QueryRow(`
		INSERT INTO refunds (order_id, amount, reason, reason_detail, status, requested_by, requested_by_name)
		SELECT $1, $2, $3, $4, 'pending', $5, $6
		WHERE EXISTS (SELECT 1 FROM orders WHERE id = $1)
		  AND $2 <= (
			SELECT o.total_price - COALESCE((
				SELECT SUM(r.amount)
				FROM refunds r
				WHERE r.order_id = o.id
				  AND r.status IN ('pending', 'approved')
			), 0)
			FROM orders o
			WHERE o.id = $1
		  )
		RETURNING id, created_at
	`, refund.OrderID, refund.Amount, refund.Reason, refund.ReasonDetail,
		refund.RequestedBy, refund.RequestedByName,
	).Scan(&refund.ID, &refund.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("buyurtma topilmadi yoki refund summasi qaytarilmagan summadan oshib ketdi")
	}
	return err
}

// ApproveRefund approves a pending refund and records it as a cash_out in the active cashier shift.
func (r *RefundRepository) ApproveRefund(refundID int, approvedBy int, approvedByName string, refundMethod string) error {
	now := time.Now()

	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Approve the refund
	var refundAmount float64
	err = tx.QueryRow(`
		UPDATE refunds 
		SET status = 'approved', approved_by = $1, approved_by_name = $2, 
		    refund_method = $3, resolved_at = $4
		WHERE id = $5 AND status = 'pending'
		RETURNING amount
	`, approvedBy, approvedByName, refundMethod, now, refundID).Scan(&refundAmount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("refund topilmadi yoki allaqachon qayta ishlangan")
		}
		return err
	}

	// 2. If refund is cash, record it as cash_out in the active shift (if any)
	if refundMethod == "cash" {
		var shiftID int
		shiftErr := tx.Get(&shiftID, `
			SELECT id FROM cashier_shifts 
			WHERE cashier_id = $1 AND status = 'open' 
			ORDER BY opened_at DESC LIMIT 1
		`, approvedBy)
		if shiftErr == nil && shiftID > 0 {
			reason := fmt.Sprintf("Refund #%d qaytarildi", refundID)
			_, _ = tx.Exec(`
				INSERT INTO cash_operations (shift_id, type, amount, reason, created_by)
				VALUES ($1, 'cash_out', $2, $3, $4)
			`, shiftID, refundAmount, reason, approvedBy)
		}
	}

	return tx.Commit()
}

// RejectRefund rejects a pending refund
func (r *RefundRepository) RejectRefund(refundID int, rejectedBy int, rejectedByName string) error {
	now := time.Now()
	result, err := r.db.Exec(`
		UPDATE refunds 
		SET status = 'rejected', approved_by = $1, approved_by_name = $2, resolved_at = $3
		WHERE id = $4 AND status = 'pending'
	`, rejectedBy, rejectedByName, now, refundID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("refund topilmadi yoki allaqachon qayta ishlangan")
	}
	return nil
}

// MarkMoneyReturned marks that money has been physically returned
func (r *RefundRepository) MarkMoneyReturned(refundID int) error {
	result, err := r.db.Exec(`
		UPDATE refunds
		SET money_returned = TRUE
		WHERE id = $1 AND status = 'approved' AND money_returned = FALSE
	`, refundID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("approved refund topilmadi yoki pul allaqachon qaytarilgan")
	}
	return nil
}

// GetPendingRefunds returns all pending refund requests
func (r *RefundRepository) GetPendingRefunds() ([]models.Refund, error) {
	var refunds []models.Refund
	err := r.db.Select(&refunds, `
		SELECT id, order_id, amount, reason, 
		       COALESCE(reason_detail, '') as reason_detail,
		       status, COALESCE(refund_method, '') as refund_method,
		       money_returned,
		       requested_by, COALESCE(requested_by_name, '') as requested_by_name,
		       approved_by, COALESCE(approved_by_name, '') as approved_by_name,
		       created_at, resolved_at
		FROM refunds 
		WHERE status = 'pending'
		ORDER BY created_at DESC
	`)
	if refunds == nil {
		refunds = []models.Refund{}
	}
	return refunds, err
}

// GetAllRefunds returns all refunds with optional status filter
func (r *RefundRepository) GetAllRefunds(status string, limit int) ([]models.Refund, error) {
	var refunds []models.Refund
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := `
		SELECT id, order_id, amount, reason,
		       COALESCE(reason_detail, '') as reason_detail,
		       status, COALESCE(refund_method, '') as refund_method,
		       money_returned,
		       requested_by, COALESCE(requested_by_name, '') as requested_by_name,
		       approved_by, COALESCE(approved_by_name, '') as approved_by_name,
		       created_at, resolved_at
		FROM refunds
	`
	args := make([]interface{}, 0, 2)
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d`, len(args)+1)
	args = append(args, limit)

	err := r.db.Select(&refunds, query, args...)
	if refunds == nil {
		refunds = []models.Refund{}
	}
	return refunds, err
}

// GetRefundsByOrderID returns all refunds for an order
func (r *RefundRepository) GetRefundsByOrderID(orderID int) ([]models.Refund, error) {
	var refunds []models.Refund
	err := r.db.Select(&refunds, `
		SELECT id, order_id, amount, reason,
		       COALESCE(reason_detail, '') as reason_detail,
		       status, COALESCE(refund_method, '') as refund_method,
		       money_returned,
		       requested_by, COALESCE(requested_by_name, '') as requested_by_name,
		       approved_by, COALESCE(approved_by_name, '') as approved_by_name,
		       created_at, resolved_at
		FROM refunds
		WHERE order_id = $1
		ORDER BY created_at DESC
	`, orderID)
	if refunds == nil {
		refunds = []models.Refund{}
	}
	return refunds, err
}

// GetRefundByID returns a single refund
func (r *RefundRepository) GetRefundByID(id int) (*models.Refund, error) {
	refund := &models.Refund{}
	err := r.db.Get(refund, `
		SELECT id, order_id, amount, reason,
		       COALESCE(reason_detail, '') as reason_detail,
		       status, COALESCE(refund_method, '') as refund_method,
		       money_returned,
		       requested_by, COALESCE(requested_by_name, '') as requested_by_name,
		       approved_by, COALESCE(approved_by_name, '') as approved_by_name,
		       created_at, resolved_at
		FROM refunds WHERE id = $1
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return refund, nil
}

// CountPending returns the number of pending refunds
func (r *RefundRepository) CountPending() (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM refunds WHERE status = 'pending'`)
	return count, err
}

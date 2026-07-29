package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type CashierRepository struct {
	db *sqlx.DB
}

func NewCashierRepository(db *sqlx.DB) *CashierRepository {
	return &CashierRepository{db: db}
}

// OpenShift creates a new cashier shift
func (r *CashierRepository) OpenShift(cashierID int, openingCash float64) (*models.CashierShift, error) {
	// Check if there's already an open shift for this cashier
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM cashier_shifts WHERE cashier_id = $1 AND status = 'open'`, cashierID)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("sizda allaqachon ochiq smena mavjud")
	}

	shift := &models.CashierShift{}
	err = r.db.QueryRow(`
		INSERT INTO cashier_shifts (cashier_id, opening_cash, status)
		VALUES ($1, $2, 'open')
		RETURNING id, cashier_id, opened_at, opening_cash, status, 
		          COALESCE(total_sales, 0) as total_sales, 
		          COALESCE(total_cash_sales, 0) as total_cash_sales,
		          COALESCE(total_card_sales, 0) as total_card_sales,
		          COALESCE(total_click_sales, 0) as total_click_sales,
		          COALESCE(total_nasiya_sales, 0) as total_nasiya_sales,
		          COALESCE(total_orders, 0) as total_orders,
		          created_at
	`, cashierID, openingCash).Scan(
		&shift.ID, &shift.CashierID, &shift.OpenedAt, &shift.OpeningCash, &shift.Status,
		&shift.TotalSales, &shift.TotalCashSales, &shift.TotalCardSales,
		&shift.TotalClickSales, &shift.TotalNasiyaSales, &shift.TotalOrders,
		&shift.OpenedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("smena ochishda xatolik: %w", err)
	}

	return shift, nil
}

// GetActiveShift returns the current open shift for a cashier
func (r *CashierRepository) GetActiveShift(cashierID int) (*models.CashierShift, error) {
	shift := &models.CashierShift{}
	err := r.db.Get(shift, `
		SELECT cs.id, cs.cashier_id, cs.opened_at, cs.closed_at, 
		       cs.opening_cash, cs.closing_cash, cs.expected_cash,
		       COALESCE(cs.total_sales, 0) as total_sales,
		       COALESCE(cs.total_cash_sales, 0) as total_cash_sales,
		       COALESCE(cs.total_card_sales, 0) as total_card_sales,
		       COALESCE(cs.total_click_sales, 0) as total_click_sales,
		       COALESCE(cs.total_nasiya_sales, 0) as total_nasiya_sales,
		       COALESCE(cs.total_orders, 0) as total_orders,
		       cs.status, COALESCE(cs.notes, '') as notes,
		       COALESCE(u.full_name, '') as cashier_name
		FROM cashier_shifts cs
		LEFT JOIN users u ON cs.cashier_id = u.id
		WHERE cs.cashier_id = $1 AND cs.status = 'open'
		ORDER BY cs.opened_at DESC LIMIT 1
	`, cashierID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Load cash operations
	var ops []models.CashOperation
	_ = r.db.Select(&ops, `SELECT * FROM cash_operations WHERE shift_id = $1 ORDER BY created_at`, shift.ID)
	shift.CashOperations = ops

	return shift, nil
}

// CloseShift closes an active shift with reconciliation
func (r *CashierRepository) CloseShift(shiftID int, closingCash float64, notes string) (*models.CashierShift, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Calculate expected cash: opening + cash_sales + cash_in - cash_out
	var shift models.CashierShift
	err = tx.Get(&shift, `
		SELECT id, cashier_id, opening_cash, 
		       COALESCE(total_cash_sales, 0) as total_cash_sales,
		       COALESCE(total_sales, 0) as total_sales,
		       COALESCE(total_card_sales, 0) as total_card_sales,
		       COALESCE(total_click_sales, 0) as total_click_sales,
		       COALESCE(total_nasiya_sales, 0) as total_nasiya_sales,
		       COALESCE(total_orders, 0) as total_orders
		FROM cashier_shifts WHERE id = $1 AND status = 'open'
	`, shiftID)
	if err != nil {
		return nil, fmt.Errorf("ochiq smena topilmadi")
	}

	// Sum cash operations
	var totalCashIn, totalCashOut float64
	_ = tx.Get(&totalCashIn, `SELECT COALESCE(SUM(amount), 0) FROM cash_operations WHERE shift_id = $1 AND type = 'cash_in'`, shiftID)
	_ = tx.Get(&totalCashOut, `SELECT COALESCE(SUM(amount), 0) FROM cash_operations WHERE shift_id = $1 AND type = 'cash_out'`, shiftID)

	expectedCash := shift.OpeningCash + shift.TotalCashSales + totalCashIn - totalCashOut
	now := time.Now()

	_, err = tx.Exec(`
		UPDATE cashier_shifts 
		SET status = 'closed', closed_at = $1, closing_cash = $2, expected_cash = $3, notes = $4
		WHERE id = $5
	`, now, closingCash, expectedCash, notes, shiftID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Return updated shift
	return r.GetShiftByID(shiftID)
}

// GetShiftByID returns a shift by ID with cash operations
func (r *CashierRepository) GetShiftByID(shiftID int) (*models.CashierShift, error) {
	shift := &models.CashierShift{}
	err := r.db.Get(shift, `
		SELECT cs.id, cs.cashier_id, cs.opened_at, cs.closed_at,
		       cs.opening_cash, cs.closing_cash, cs.expected_cash,
		       COALESCE(cs.total_sales, 0) as total_sales,
		       COALESCE(cs.total_cash_sales, 0) as total_cash_sales,
		       COALESCE(cs.total_card_sales, 0) as total_card_sales,
		       COALESCE(cs.total_click_sales, 0) as total_click_sales,
		       COALESCE(cs.total_nasiya_sales, 0) as total_nasiya_sales,
		       COALESCE(cs.total_orders, 0) as total_orders,
		       cs.status, COALESCE(cs.notes, '') as notes,
		       COALESCE(u.full_name, '') as cashier_name
		FROM cashier_shifts cs
		LEFT JOIN users u ON cs.cashier_id = u.id
		WHERE cs.id = $1
	`, shiftID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var ops []models.CashOperation
	_ = r.db.Select(&ops, `SELECT * FROM cash_operations WHERE shift_id = $1 ORDER BY created_at`, shiftID)
	shift.CashOperations = ops

	return shift, nil
}

// AddCashOperation adds a cash-in or cash-out operation
func (r *CashierRepository) AddCashOperation(op *models.CashOperation) error {
	return r.db.QueryRow(`
		INSERT INTO cash_operations (shift_id, type, amount, reason, created_by)
		SELECT $1, $2, $3, $4, $5
		WHERE EXISTS (
			SELECT 1 FROM cashier_shifts WHERE id = $1 AND status = 'open'
		)
		RETURNING id, created_at
	`, op.ShiftID, op.Type, op.Amount, op.Reason, op.CreatedBy).Scan(&op.ID, &op.CreatedAt)
}

// RecordSale updates shift totals when a sale is made
func (r *CashierRepository) RecordSale(shiftID int, totalAmount float64, cashAmount, cardAmount, clickAmount, nasiyaAmount float64) error {
	result, err := r.db.Exec(`
		UPDATE cashier_shifts 
		SET total_sales = total_sales + $1,
		    total_cash_sales = total_cash_sales + $2,
		    total_card_sales = total_card_sales + $3,
		    total_click_sales = total_click_sales + $4,
		    total_nasiya_sales = total_nasiya_sales + $5,
		    total_orders = total_orders + 1
		WHERE id = $6 AND status = 'open'
	`, totalAmount, cashAmount, cardAmount, clickAmount, nasiyaAmount, shiftID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("ochiq smena topilmadi")
	}
	return nil
}

// GetShiftReport generates X-report (during shift) or Z-report (after closing)
func (r *CashierRepository) GetShiftReport(shiftID int) (*models.ShiftReport, error) {
	shift, err := r.GetShiftByID(shiftID)
	if err != nil || shift == nil {
		return nil, fmt.Errorf("smena topilmadi")
	}

	var totalCashIn, totalCashOut float64
	_ = r.db.Get(&totalCashIn, `SELECT COALESCE(SUM(amount), 0) FROM cash_operations WHERE shift_id = $1 AND type = 'cash_in'`, shiftID)
	_ = r.db.Get(&totalCashOut, `SELECT COALESCE(SUM(amount), 0) FROM cash_operations WHERE shift_id = $1 AND type = 'cash_out'`, shiftID)

	expectedCash := shift.OpeningCash + shift.TotalCashSales + totalCashIn - totalCashOut
	var cashDiff float64
	if shift.ClosingCash != nil {
		cashDiff = expectedCash - *shift.ClosingCash
	}

	report := &models.ShiftReport{
		Shift:          *shift,
		CashOperations: shift.CashOperations,
		TotalCashIn:    totalCashIn,
		TotalCashOut:   totalCashOut,
		ExpectedCash:   expectedCash,
		CashDifference: cashDiff,
	}

	return report, nil
}

// GetAllShifts returns recent shifts for admin view
func (r *CashierRepository) GetAllShifts(limit int) ([]models.CashierShift, error) {
	var shifts []models.CashierShift
	err := r.db.Select(&shifts, `
		SELECT cs.id, cs.cashier_id, cs.opened_at, cs.closed_at,
		       cs.opening_cash, cs.closing_cash, cs.expected_cash,
		       COALESCE(cs.total_sales, 0) as total_sales,
		       COALESCE(cs.total_cash_sales, 0) as total_cash_sales,
		       COALESCE(cs.total_card_sales, 0) as total_card_sales,
		       COALESCE(cs.total_click_sales, 0) as total_click_sales,
		       COALESCE(cs.total_nasiya_sales, 0) as total_nasiya_sales,
		       COALESCE(cs.total_orders, 0) as total_orders,
		       cs.status, COALESCE(cs.notes, '') as notes,
		       COALESCE(u.full_name, '') as cashier_name
		FROM cashier_shifts cs
		LEFT JOIN users u ON cs.cashier_id = u.id
		ORDER BY cs.opened_at DESC
		LIMIT $1
	`, limit)
	return shifts, err
}

package repository

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type FiscalRepository struct {
	db *sqlx.DB
}

func NewFiscalRepository(db *sqlx.DB) *FiscalRepository {
	return &FiscalRepository{db: db}
}

// CreateReceipt inserts a new fiscal receipt
func (r *FiscalRepository) CreateReceipt(receipt *models.FiscalReceipt) error {
	query := `INSERT INTO fiscal_receipts 
		(order_id, receipt_number, fiscal_sign, total_amount, vat_rate, vat_amount, subtotal, payment_method, inn, company_name, cashier_name, status, ofd_response)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at`
	return r.db.QueryRow(query,
		receipt.OrderID, receipt.ReceiptNumber, receipt.FiscalSign,
		receipt.TotalAmount, receipt.VATRate, receipt.VATAmount, receipt.Subtotal,
		receipt.PaymentMethod, receipt.INN, receipt.CompanyName, receipt.CashierName,
		receipt.Status, receipt.OFDResponse,
	).Scan(&receipt.ID, &receipt.CreatedAt)
}

// GetByOrderID retrieves fiscal receipt for a specific order
func (r *FiscalRepository) GetByOrderID(orderID int) (*models.FiscalReceipt, error) {
	var receipt models.FiscalReceipt
	err := r.db.Get(&receipt, `SELECT * FROM fiscal_receipts WHERE order_id = $1 ORDER BY id DESC LIMIT 1`, orderID)
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

// GetAll retrieves all fiscal receipts with optional pagination
func (r *FiscalRepository) GetAll(limit, offset int, status string) ([]models.FiscalReceipt, error) {
	var receipts []models.FiscalReceipt
	query := `SELECT * FROM fiscal_receipts`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(` WHERE status = $%d`, argIdx)
		args = append(args, status)
		argIdx++
	}

	query += ` ORDER BY created_at DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIdx)
		args = append(args, limit)
		argIdx++
	}
	if offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argIdx)
		args = append(args, offset)
	}

	err := r.db.Select(&receipts, query, args...)
	if err != nil {
		return nil, err
	}
	return receipts, nil
}

// UpdateStatus updates the OFD status and response of a fiscal receipt
func (r *FiscalRepository) UpdateStatus(id int, status string, fiscalSign string, ofdResponse string) error {
	query := `UPDATE fiscal_receipts SET status = $1, fiscal_sign = $2, ofd_response = $3 WHERE id = $4`
	_, err := r.db.Exec(query, status, fiscalSign, ofdResponse, id)
	return err
}

// GenerateReceiptNumber creates a unique receipt number in YYMMDD-NNNNN format
func (r *FiscalRepository) GenerateReceiptNumber() (string, error) {
	now := time.Now()
	datePrefix := now.Format("060102") // YYMMDD

	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM fiscal_receipts WHERE receipt_number LIKE $1`, datePrefix+"-%")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%05d", datePrefix, count+1), nil
}

// GetStats returns fiscal receipt statistics
func (r *FiscalRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	var totalReceipts int
	r.db.Get(&totalReceipts, `SELECT COUNT(*) FROM fiscal_receipts`)
	stats["total_receipts"] = totalReceipts

	var totalVAT float64
	r.db.Get(&totalVAT, `SELECT COALESCE(SUM(vat_amount), 0) FROM fiscal_receipts WHERE status != 'error'`)
	stats["total_vat"] = totalVAT

	var totalAmount float64
	r.db.Get(&totalAmount, `SELECT COALESCE(SUM(total_amount), 0) FROM fiscal_receipts WHERE status != 'error'`)
	stats["total_amount"] = totalAmount

	var pendingCount int
	r.db.Get(&pendingCount, `SELECT COUNT(*) FROM fiscal_receipts WHERE status = 'local'`)
	stats["pending_sync"] = pendingCount

	return stats, nil
}

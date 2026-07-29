package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type DebtRepository struct {
	db *sqlx.DB
}

func NewDebtRepository(db *sqlx.DB) *DebtRepository {
	return &DebtRepository{db: db}
}

// CreateDebtor creates a new debtor record
func (r *DebtRepository) CreateDebtor(debtor *models.Debtor) error {
	return r.db.QueryRow(`
		INSERT INTO debtors (name, phone, notes)
		VALUES ($1, $2, $3)
		RETURNING id, total_debt, created_at, updated_at
	`, debtor.Name, debtor.Phone, debtor.Notes).Scan(
		&debtor.ID, &debtor.TotalDebt, &debtor.CreatedAt, &debtor.UpdatedAt,
	)
}

// GetAllDebtors returns all debtors, optionally filtering by active debt
func (r *DebtRepository) GetAllDebtors(onlyWithDebt bool) ([]models.Debtor, error) {
	var debtors []models.Debtor
	query := `
		SELECT id, name, COALESCE(phone, '') as phone, total_debt, 
		       COALESCE(notes, '') as notes, created_at, updated_at
		FROM debtors
	`
	if onlyWithDebt {
		query += ` WHERE total_debt > 0`
	}
	query += ` ORDER BY total_debt DESC, name ASC`

	err := r.db.Select(&debtors, query)
	if debtors == nil {
		debtors = []models.Debtor{}
	}
	return debtors, err
}

// GetDebtorByID returns a single debtor
func (r *DebtRepository) GetDebtorByID(id int) (*models.Debtor, error) {
	debtor := &models.Debtor{}
	err := r.db.Get(debtor, `
		SELECT id, name, COALESCE(phone, '') as phone, total_debt,
		       COALESCE(notes, '') as notes, created_at, updated_at
		FROM debtors WHERE id = $1
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return debtor, nil
}

// UpdateDebtor updates a debtor's info
func (r *DebtRepository) UpdateDebtor(debtor *models.Debtor) error {
	_, err := r.db.Exec(`
		UPDATE debtors SET name = $1, phone = $2, notes = $3 WHERE id = $4
	`, debtor.Name, debtor.Phone, debtor.Notes, debtor.ID)
	return err
}

// AddDebtRecord records a debt or payment and updates the debtor's total
func (r *DebtRepository) AddDebtRecord(record *models.DebtRecord) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert the record
	err = tx.QueryRow(`
		INSERT INTO debt_records (debtor_id, order_id, amount, type, payment_method, description, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, record.DebtorID, record.OrderID, record.Amount, record.Type,
		record.PaymentMethod, record.Description, record.CreatedBy,
	).Scan(&record.ID, &record.CreatedAt)
	if err != nil {
		return fmt.Errorf("qarz yozuvini saqlashda xatolik: %w", err)
	}

	// Update total_debt on debtor
	if record.Type == "debt" {
		_, err = tx.Exec(`UPDATE debtors SET total_debt = total_debt + $1 WHERE id = $2`, record.Amount, record.DebtorID)
	} else if record.Type == "payment" {
		_, err = tx.Exec(`UPDATE debtors SET total_debt = total_debt - $1 WHERE id = $2`, record.Amount, record.DebtorID)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetDebtHistory returns all records for a debtor
func (r *DebtRepository) GetDebtHistory(debtorID int) ([]models.DebtRecord, error) {
	var records []models.DebtRecord
	err := r.db.Select(&records, `
		SELECT dr.id, dr.debtor_id, dr.order_id, dr.amount, dr.type,
		       COALESCE(dr.payment_method, '') as payment_method,
		       COALESCE(dr.description, '') as description,
		       dr.created_by, dr.created_at,
		       COALESCE(u.full_name, '') as created_by_name
		FROM debt_records dr
		LEFT JOIN users u ON dr.created_by = u.id
		WHERE dr.debtor_id = $1
		ORDER BY dr.created_at DESC
	`, debtorID)
	if records == nil {
		records = []models.DebtRecord{}
	}
	return records, err
}

// RecalculateDebt recalculates total_debt from all records
func (r *DebtRepository) RecalculateDebt(debtorID int) error {
	_, err := r.db.Exec(`
		UPDATE debtors SET total_debt = (
			SELECT COALESCE(
				SUM(CASE WHEN type = 'debt' THEN amount ELSE -amount END), 0
			) FROM debt_records WHERE debtor_id = $1
		) WHERE id = $1
	`, debtorID)
	return err
}

// GetDebtSummary returns aggregate debt statistics
func (r *DebtRepository) GetDebtSummary() (map[string]interface{}, error) {
	var totalDebt float64
	var debtorCount int
	_ = r.db.Get(&totalDebt, `SELECT COALESCE(SUM(total_debt), 0) FROM debtors WHERE total_debt > 0`)
	_ = r.db.Get(&debtorCount, `SELECT COUNT(*) FROM debtors WHERE total_debt > 0`)

	return map[string]interface{}{
		"total_debt":   totalDebt,
		"debtor_count": debtorCount,
	}, nil
}

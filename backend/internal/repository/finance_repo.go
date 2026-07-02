package repository

import (
	"database/sql"
	"fmt"
	"github.com/username/kafe-backend/internal/models"
	"github.com/jmoiron/sqlx"
)

type FinanceRepository interface {
	CreateExpense(expense *models.Expense) error
	GetExpenses() ([]models.Expense, error)
	GetStats() (*models.FinanceStats, error)
}

type financeRepository struct {
	db *sqlx.DB
}

func NewFinanceRepository(db *sqlx.DB) FinanceRepository {
	return &financeRepository{db: db}
}

func (r *financeRepository) CreateExpense(expense *models.Expense) error {
	query := `
		INSERT INTO expenses (amount, category, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, expense.Amount, expense.Category, expense.Description).
		Scan(&expense.ID, &expense.CreatedAt)
}

func (r *financeRepository) GetExpenses() ([]models.Expense, error) {
	query := `SELECT id, amount, category, description, created_at FROM expenses ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense
	for rows.Next() {
		var e models.Expense
		var desc sql.NullString
		if err := rows.Scan(&e.ID, &e.Amount, &e.Category, &desc, &e.CreatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			e.Description = desc.String
		}
		expenses = append(expenses, e)
	}
	return expenses, nil
}

func (r *financeRepository) GetStats() (*models.FinanceStats, error) {
	stats := &models.FinanceStats{}

	// Calculate total revenue from delivered orders
	// Using 'delivered' status to only count completed orders
	revenueQuery := `SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'delivered'`
	if err := r.db.QueryRow(revenueQuery).Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("failed to get revenue: %v", err)
	}

	// Calculate total expenses
	expensesQuery := `SELECT COALESCE(SUM(amount), 0) FROM expenses`
	if err := r.db.QueryRow(expensesQuery).Scan(&stats.TotalExpenses); err != nil {
		return nil, fmt.Errorf("failed to get expenses: %v", err)
	}

	stats.NetProfit = stats.TotalRevenue - stats.TotalExpenses

	return stats, nil
}

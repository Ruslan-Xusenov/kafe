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
	GetWaiterSalaries(startDate, endDate string) ([]models.WaiterSalary, error)
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

	// Payment method breakdown
	paymentQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total_price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total_price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'click' THEN total_price ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN payment_method = 'nasiya' THEN total_price ELSE 0 END), 0)
		FROM orders WHERE status = 'delivered'
	`
	if err := r.db.QueryRow(paymentQuery).Scan(&stats.CashRevenue, &stats.CardRevenue, &stats.ClickRevenue, &stats.NasiyaRevenue); err != nil {
		// Ignore error for backwards compatibility
		stats.CashRevenue = 0
		stats.CardRevenue = 0
		stats.ClickRevenue = 0
		stats.NasiyaRevenue = 0
	}

	return stats, nil
}

func (r *financeRepository) GetWaiterSalaries(startDate, endDate string) ([]models.WaiterSalary, error) {
	query := `
		SELECT 
			u.id as waiter_id,
			u.full_name as waiter_name,
			COUNT(o.id) as total_orders,
			COALESCE(SUM(o.service_fee), 0) as total_salary
		FROM users u
		LEFT JOIN orders o ON u.id = o.waiter_id 
			AND o.status = 'delivered' 
			AND o.table_id IS NOT NULL
			AND DATE(o.created_at) >= $1 
			AND DATE(o.created_at) <= $2
		WHERE u.role = 'waiter'
		GROUP BY u.id, u.full_name
		ORDER BY total_salary DESC
	`
	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var salaries []models.WaiterSalary
	for rows.Next() {
		var s models.WaiterSalary
		if err := rows.Scan(&s.WaiterID, &s.WaiterName, &s.TotalOrders, &s.TotalSalary); err != nil {
			return nil, err
		}
		salaries = append(salaries, s)
	}
	
	if salaries == nil {
		salaries = []models.WaiterSalary{}
	}
	
	return salaries, nil
}

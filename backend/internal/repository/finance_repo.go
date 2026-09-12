package repository

import (
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/username/kafe-backend/internal/models"
)

type FinanceRepository interface {
	CreateExpense(expense *models.Expense) error
	GetExpenses() ([]models.Expense, error)
	GetStats() (*models.FinanceStats, error)
	GetWaiterSalaries(startDate, endDate string) ([]models.WaiterSalary, error)
	CountActiveOrders() (int, error)
	ForceCloseActiveOrders() error
	CloseShift() error
	GetAllTimeRevenue() (float64, error)
	GetShiftSoldItems() ([]models.ShiftSoldItem, error)
	GetDailyReport(date string) (*models.DailyReport, error)
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

	query := `
		SELECT id, amount, category, description, created_at 
		FROM expenses 
		WHERE created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
		ORDER BY created_at DESC
	`
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
	revenueQuery := `
		SELECT COALESCE(SUM(total_price), 0) 
		FROM orders 
		WHERE status = 'delivered' AND created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(revenueQuery).Scan(&stats.TotalRevenue); err != nil {
		return nil, fmt.Errorf("failed to get revenue: %v", err)
	}

	// Calculate total operating expenses (excluding mahsulot)
	expensesQuery := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses
		WHERE category != 'mahsulot' AND created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(expensesQuery).Scan(&stats.TotalExpenses); err != nil {
		return nil, fmt.Errorf("failed to get expenses: %v", err)
	}

	// Calculate total inventory purchases (mahsulot)
	invQuery := `
		SELECT COALESCE(SUM(amount), 0) 
		FROM expenses
		WHERE category = 'mahsulot' AND created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(invQuery).Scan(&stats.TotalInventoryPurchase); err != nil {
		stats.TotalInventoryPurchase = 0
	}

	// Calculate Total Cost of Goods Sold (TotalCostPrice)
	costQuery := `
		SELECT COALESCE(SUM(oi.quantity * pi.quantity * i.cost_price), 0)
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN product_ingredients pi ON oi.product_id = pi.product_id
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE o.status = 'delivered' AND o.created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(costQuery).Scan(&stats.TotalCostPrice); err != nil {
		stats.TotalCostPrice = 0
	}

	stats.NetProfit = stats.TotalRevenue - stats.TotalExpenses - stats.TotalCostPrice

	// Payment method breakdown from order_payments table (supports mixed payments)
	paymentQuery := `
		SELECT 
			COALESCE(SUM(CASE WHEN op.method = 'cash' THEN op.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN op.method = 'card' THEN op.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN op.method = 'click' THEN op.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN op.method = 'nasiya' THEN op.amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN op.method = 'qr' THEN op.amount ELSE 0 END), 0)
		FROM order_payments op
		JOIN orders o ON op.order_id = o.id
		WHERE o.status = 'delivered' AND o.created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(paymentQuery).Scan(&stats.CashRevenue, &stats.CardRevenue, &stats.ClickRevenue, &stats.NasiyaRevenue, &stats.QrRevenue); err != nil {
		// Fallback to old method for backward compat
		fallbackQuery := `
			SELECT 
				COALESCE(SUM(CASE WHEN payment_method = 'cash' THEN total_price ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN payment_method = 'card' THEN total_price ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN payment_method = 'click' THEN total_price ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN payment_method = 'nasiya' THEN total_price ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN payment_method = 'qr' THEN total_price ELSE 0 END), 0)
			FROM orders 
			WHERE status = 'delivered' AND created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
		`
		if err2 := r.db.QueryRow(fallbackQuery).Scan(&stats.CashRevenue, &stats.CardRevenue, &stats.ClickRevenue, &stats.NasiyaRevenue, &stats.QrRevenue); err2 != nil {
			stats.CashRevenue = 0
			stats.CardRevenue = 0
			stats.ClickRevenue = 0
			stats.NasiyaRevenue = 0
		}
	}

	// Calculate Waiter Salaries for the current shift
	salariesQuery := `
		SELECT COALESCE(SUM((o.total_price - COALESCE(o.service_fee, 0)) * (COALESCE(u.default_service_percentage, 0) / 100.0)), 0)
		FROM orders o
		JOIN users u ON u.id = o.waiter_id
		WHERE o.status = 'delivered' 
		  AND o.table_id IS NOT NULL 
		  AND u.role = 'waiter'
		  AND o.created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
	`
	if err := r.db.QueryRow(salariesQuery).Scan(&stats.TotalSalaries); err != nil {
		stats.TotalSalaries = 0
	}

	stats.RealProfit = stats.NetProfit - stats.TotalSalaries

	// Calculate Today's Stats
	todayRevenueQuery := `SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'delivered' AND DATE(created_at) = CURRENT_DATE`
	_ = r.db.QueryRow(todayRevenueQuery).Scan(&stats.TodayRevenue)

	todayExpensesQuery := `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category != 'mahsulot' AND DATE(created_at) = CURRENT_DATE`
	_ = r.db.QueryRow(todayExpensesQuery).Scan(&stats.TodayExpenses)

	todayInvQuery := `SELECT COALESCE(SUM(amount), 0) FROM expenses WHERE category = 'mahsulot' AND DATE(created_at) = CURRENT_DATE`
	_ = r.db.QueryRow(todayInvQuery).Scan(&stats.TodayInventoryPurchase)
	
	todayCostQuery := `
		SELECT COALESCE(SUM(i.cost_price * pi.quantity * oi.quantity), 0)
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN product_ingredients pi ON oi.product_id = pi.product_id
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE o.status = 'delivered' AND DATE(o.created_at) = CURRENT_DATE
	`
	_ = r.db.QueryRow(todayCostQuery).Scan(&stats.TodayCostPrice)

	todaySalariesQuery := `
		SELECT COALESCE(SUM((o.total_price - COALESCE(o.service_fee, 0)) * (COALESCE(u.default_service_percentage, 0) / 100.0)), 0)
		FROM orders o
		JOIN users u ON u.id = o.waiter_id
		WHERE o.status = 'delivered' 
		  AND o.table_id IS NOT NULL 
		  AND u.role = 'waiter'
		  AND DATE(o.created_at) = CURRENT_DATE
	`
	_ = r.db.QueryRow(todaySalariesQuery).Scan(&stats.TodaySalaries)

	stats.TodayNetProfit = stats.TodayRevenue - stats.TodayExpenses - stats.TodayCostPrice
	stats.TodayRealProfit = stats.TodayNetProfit - stats.TodaySalaries

	return stats, nil
}

func (r *financeRepository) GetWaiterSalaries(startDate, endDate string) ([]models.WaiterSalary, error) {
	query := `
		SELECT 
			u.id as waiter_id,
			u.full_name as waiter_name,
			COUNT(o.id) as total_orders,
			COALESCE(SUM((o.total_price - COALESCE(o.service_fee, 0)) * (COALESCE(u.default_service_percentage, 0) / 100.0)), 0) as total_salary
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

func (r *financeRepository) CountActiveOrders() (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM orders
		WHERE status IN ('new', 'preparing', 'ready', 'on_way')
	`).Scan(&count)
	return count, err
}

func (r *financeRepository) CloseShift() error {
	query := `
		INSERT INTO settings (key, value, updated_at) 
		VALUES ('last_shift_closed_at', NOW()::text, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *financeRepository) ForceCloseActiveOrders() error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update active orders
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = 'delivered', payment_method = 'cash', updated_at = NOW() 
		WHERE status IN ('new', 'preparing', 'ready')
	`)
	if err != nil {
		return err
	}

	// Update all tables to free
	_, err = tx.Exec(`
		UPDATE tables 
		SET status = 'free'
	`)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *financeRepository) GetAllTimeRevenue() (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'delivered'`
	err := r.db.QueryRow(query).Scan(&total)
	return total, err
}

func (r *financeRepository) GetShiftSoldItems() ([]models.ShiftSoldItem, error) {
	query := `
		SELECT 
			p.name AS product_name, 
			SUM(oi.quantity) AS quantity, 
			SUM(oi.quantity * oi.price) AS total_amount
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE o.status = 'delivered' 
		  AND o.created_at > COALESCE((SELECT value::timestamp FROM settings WHERE key = 'last_shift_closed_at'), '1970-01-01'::timestamp)
		GROUP BY p.id, p.name
		ORDER BY quantity DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ShiftSoldItem
	for rows.Next() {
		var item models.ShiftSoldItem
		if err := rows.Scan(&item.ProductName, &item.Quantity, &item.TotalAmount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *financeRepository) GetDailyReport(date string) (*models.DailyReport, error) {
	report := &models.DailyReport{Date: date}

	baseFilter := `DATE(o.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Tashkent') = $1`

	// Revenue & orders count
	err := r.db.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(total_price), 0), COUNT(*)
		FROM orders o
		WHERE status = 'delivered' AND %s`, baseFilter), date).
		Scan(&report.Revenue, &report.OrdersCount)
	if err != nil {
		return nil, fmt.Errorf("daily revenue: %v", err)
	}

	// Expenses
	_ = r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0) FROM expenses
		WHERE category != 'mahsulot' AND DATE(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Tashkent') = $1`, date).
		Scan(&report.Expenses)

	// COGS
	_ = r.db.QueryRow(fmt.Sprintf(`
		SELECT COALESCE(SUM(i.cost_price * pi.quantity * oi.quantity), 0)
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN product_ingredients pi ON oi.product_id = pi.product_id
		JOIN ingredients i ON pi.ingredient_id = i.id
		WHERE o.status = 'delivered' AND %s`, baseFilter), date).
		Scan(&report.TotalCostPrice)

	report.NetProfit = report.Revenue - report.Expenses - report.TotalCostPrice

	// Payment breakdown
	payQ := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN op.method='cash'   THEN op.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN op.method='card'   THEN op.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN op.method='click'  THEN op.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN op.method='nasiya' THEN op.amount ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN op.method='qr'     THEN op.amount ELSE 0 END),0)
		FROM order_payments op
		JOIN orders o ON op.order_id = o.id
		WHERE o.status = 'delivered' AND %s`, baseFilter)
	err = r.db.QueryRow(payQ, date).Scan(
		&report.CashRevenue, &report.CardRevenue, &report.ClickRevenue, &report.NasiyaRevenue, &report.QrRevenue)
	if err != nil {
		// fallback
		_ = r.db.QueryRow(fmt.Sprintf(`
			SELECT
				COALESCE(SUM(CASE WHEN payment_method='cash'   THEN total_price ELSE 0 END),0),
				COALESCE(SUM(CASE WHEN payment_method='card'   THEN total_price ELSE 0 END),0),
				COALESCE(SUM(CASE WHEN payment_method='click'  THEN total_price ELSE 0 END),0),
				COALESCE(SUM(CASE WHEN payment_method='nasiya' THEN total_price ELSE 0 END),0),
				COALESCE(SUM(CASE WHEN payment_method='qr'     THEN total_price ELSE 0 END),0)
			FROM orders o WHERE status='delivered' AND %s`, baseFilter), date).
			Scan(&report.CashRevenue, &report.CardRevenue, &report.ClickRevenue, &report.NasiyaRevenue, &report.QrRevenue)
	}

	// Top products
	rows, err := r.db.Query(fmt.Sprintf(`
		SELECT p.name, SUM(oi.quantity), SUM(oi.quantity * oi.price)
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		JOIN products p ON oi.product_id = p.id
		WHERE o.status = 'delivered' AND %s
		GROUP BY p.id, p.name
		ORDER BY SUM(oi.quantity * oi.price) DESC
		LIMIT 20`, baseFilter), date)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var item models.ShiftSoldItem
			if err := rows.Scan(&item.ProductName, &item.Quantity, &item.TotalAmount); err == nil {
				report.TopProducts = append(report.TopProducts, item)
			}
		}
	}
	if report.TopProducts == nil {
		report.TopProducts = []models.ShiftSoldItem{}
	}

	// Fetch expenses for the day
	expRows, err := r.db.Query(`
		SELECT id, amount, category, description, created_at
		FROM expenses
		WHERE DATE(created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Tashkent') = $1
		ORDER BY created_at DESC`, date)
	if err == nil {
		defer expRows.Close()
		for expRows.Next() {
			var e models.Expense
			var desc sql.NullString
			if err := expRows.Scan(&e.ID, &e.Amount, &e.Category, &desc, &e.CreatedAt); err == nil {
				if desc.Valid {
					e.Description = desc.String
				}
				report.ExpenseList = append(report.ExpenseList, e)
			}
		}
	}
	if report.ExpenseList == nil {
		report.ExpenseList = []models.Expense{}
	}

	return report, nil
}

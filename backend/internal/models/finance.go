package models

import "time"

type Expense struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentInput struct {
	Method string  `json:"method"`
	Amount float64 `json:"amount"`
}

type OrderPayment struct {
	ID        int       `json:"id" db:"id"`
	OrderID   int       `json:"order_id" db:"order_id"`
	Method    string    `json:"method" db:"method"`
	Amount    float64   `json:"amount" db:"amount"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type FinanceStats struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	NetProfit     float64 `json:"net_profit"`
	CashRevenue   float64 `json:"cash_revenue"`
	CardRevenue   float64 `json:"card_revenue"`
	ClickRevenue  float64 `json:"click_revenue"`
	NasiyaRevenue float64 `json:"nasiya_revenue"`
	TotalSalaries float64 `json:"total_salaries"`
	RealProfit    float64 `json:"real_profit"`
}

type WaiterSalary struct {
	WaiterID    int     `json:"waiter_id"`
	WaiterName  string  `json:"waiter_name"`
	TotalOrders int     `json:"total_orders"`
	TotalSalary float64 `json:"total_salary"`
}

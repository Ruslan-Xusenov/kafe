package models

import "time"

type Expense struct {
	ID          int       `json:"id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type FinanceStats struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	NetProfit     float64 `json:"net_profit"`
	CashRevenue   float64 `json:"cash_revenue"`
	CardRevenue   float64 `json:"card_revenue"`
	ClickRevenue  float64 `json:"click_revenue"`
	NasiyaRevenue float64 `json:"nasiya_revenue"`
}

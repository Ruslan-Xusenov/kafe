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
	TotalRevenue           float64 `json:"total_revenue"`
	TotalExpenses          float64 `json:"total_expenses"`
	TotalInventoryPurchase float64 `json:"total_inventory_purchase"`
	TotalCostPrice         float64 `json:"total_cost_price"`
	NetProfit              float64 `json:"net_profit"`
	CashRevenue            float64 `json:"cash_revenue"`
	CardRevenue            float64 `json:"card_revenue"`
	ClickRevenue           float64 `json:"click_revenue"`
	NasiyaRevenue          float64 `json:"nasiya_revenue"`
	QrRevenue              float64 `json:"qr_revenue"`
	TotalSalaries          float64 `json:"total_salaries"`
	RealProfit             float64 `json:"real_profit"`

	// Stats for today (irrespective of shift closures)
	TodayRevenue           float64 `json:"today_revenue"`
	TodayExpenses          float64 `json:"today_expenses"`
	TodayInventoryPurchase float64 `json:"today_inventory_purchase"`
	TodayCostPrice         float64 `json:"today_cost_price"`
	TodayNetProfit         float64 `json:"today_net_profit"`
	TodaySalaries          float64 `json:"today_salaries"`
	TodayRealProfit        float64 `json:"today_real_profit"`
}

type WaiterSalary struct {
	WaiterID    int     `json:"waiter_id"`
	WaiterName  string  `json:"waiter_name"`
	TotalOrders int     `json:"total_orders"`
	TotalSalary float64 `json:"total_salary"`
}

type ShiftSoldItem struct {
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

type DailyReport struct {
	Date          string          `json:"date"`
	Revenue       float64         `json:"revenue"`
	Expenses      float64         `json:"expenses"`
	TotalCostPrice float64        `json:"total_cost_price"`
	NetProfit     float64         `json:"net_profit"`
	OrdersCount   int             `json:"orders_count"`
	CashRevenue   float64         `json:"cash_revenue"`
	CardRevenue   float64         `json:"card_revenue"`
	ClickRevenue  float64         `json:"click_revenue"`
	NasiyaRevenue float64         `json:"nasiya_revenue"`
	QrRevenue     float64         `json:"qr_revenue"`
	TopProducts   []ShiftSoldItem `json:"top_products"`
	ExpenseList   []Expense       `json:"expense_list"`
}

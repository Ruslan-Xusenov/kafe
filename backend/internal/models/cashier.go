package models

import "time"

// CashierShift represents a cashier work session
type CashierShift struct {
	ID              int        `json:"id" db:"id"`
	CashierID       int        `json:"cashier_id" db:"cashier_id"`
	CashierName     string     `json:"cashier_name,omitempty" db:"cashier_name"` // Joined field
	OpenedAt        time.Time  `json:"opened_at" db:"opened_at"`
	ClosedAt        *time.Time `json:"closed_at" db:"closed_at"`
	OpeningCash     float64    `json:"opening_cash" db:"opening_cash"`
	ClosingCash     *float64   `json:"closing_cash" db:"closing_cash"`
	ExpectedCash    *float64   `json:"expected_cash" db:"expected_cash"`
	TotalSales      float64    `json:"total_sales" db:"total_sales"`
	TotalCashSales  float64    `json:"total_cash_sales" db:"total_cash_sales"`
	TotalCardSales  float64    `json:"total_card_sales" db:"total_card_sales"`
	TotalClickSales float64    `json:"total_click_sales" db:"total_click_sales"`
	TotalNasiyaSales float64   `json:"total_nasiya_sales" db:"total_nasiya_sales"`
	TotalOrders     int        `json:"total_orders" db:"total_orders"`
	Status          string     `json:"status" db:"status"` // open, closed
	Notes           string     `json:"notes" db:"notes"`
	CashOperations  []CashOperation `json:"cash_operations,omitempty" db:"-"`
}

// CashOperation represents a cash-in or cash-out entry during a shift
type CashOperation struct {
	ID        int       `json:"id" db:"id"`
	ShiftID   int       `json:"shift_id" db:"shift_id"`
	Type      string    `json:"type" db:"type"` // cash_in, cash_out
	Amount    float64   `json:"amount" db:"amount"`
	Reason    string    `json:"reason" db:"reason"`
	CreatedBy int       `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ShiftReport is the X/Z-report data for a shift
type ShiftReport struct {
	Shift           CashierShift    `json:"shift"`
	CashOperations  []CashOperation `json:"cash_operations"`
	TotalCashIn     float64         `json:"total_cash_in"`
	TotalCashOut    float64         `json:"total_cash_out"`
	ExpectedCash    float64         `json:"expected_cash"`
	CashDifference  float64         `json:"cash_difference"` // expected - closing (only on Z-report)
}

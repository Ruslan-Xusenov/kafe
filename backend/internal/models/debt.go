package models

import "time"

// Debtor represents a customer who has outstanding debt
type Debtor struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Phone     string    `json:"phone" db:"phone"`
	TotalDebt float64   `json:"total_debt" db:"total_debt"`
	Notes     string    `json:"notes" db:"notes"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DebtRecord represents a single debt entry or payment
type DebtRecord struct {
	ID            int       `json:"id" db:"id"`
	DebtorID      int       `json:"debtor_id" db:"debtor_id"`
	OrderID       *int      `json:"order_id" db:"order_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Type          string    `json:"type" db:"type"` // "debt" or "payment"
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	Description   string    `json:"description" db:"description"`
	CreatedBy     *int      `json:"created_by" db:"created_by"`
	CreatedByName string    `json:"created_by_name,omitempty" db:"created_by_name"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

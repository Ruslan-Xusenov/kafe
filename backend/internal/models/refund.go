package models

import "time"

type RefundStatus string

const (
	RefundPending  RefundStatus = "pending"
	RefundApproved RefundStatus = "approved"
	RefundRejected RefundStatus = "rejected"
)

// Refund represents a refund request with approval workflow
type Refund struct {
	ID              int          `json:"id" db:"id"`
	OrderID         int          `json:"order_id" db:"order_id"`
	Amount          float64      `json:"amount" db:"amount"`
	Reason          string       `json:"reason" db:"reason"`
	ReasonDetail    string       `json:"reason_detail" db:"reason_detail"`
	Status          RefundStatus `json:"status" db:"status"`
	RefundMethod    string       `json:"refund_method" db:"refund_method"`
	MoneyReturned   bool         `json:"money_returned" db:"money_returned"`
	RequestedBy     *int         `json:"requested_by" db:"requested_by"`
	RequestedByName string       `json:"requested_by_name" db:"requested_by_name"`
	ApprovedBy      *int         `json:"approved_by" db:"approved_by"`
	ApprovedByName  string       `json:"approved_by_name" db:"approved_by_name"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	ResolvedAt      *time.Time   `json:"resolved_at" db:"resolved_at"`
}

// Valid refund reasons
var ValidRefundReasons = map[string]string{
	"customer_complaint": "Mijoz shikoyati",
	"wrong_order":        "Noto'g'ri buyurtma",
	"quality_issue":      "Sifat muammosi",
	"overcharge":         "Ortiqcha hisoblangan",
	"cancelled":          "Bekor qilingan",
	"other":              "Boshqa",
}

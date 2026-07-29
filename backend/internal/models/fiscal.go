package models

import "time"

// FiscalReceipt represents a fiscal receipt (fiskal chek) for tax compliance
type FiscalReceipt struct {
	ID            int       `json:"id" db:"id"`
	OrderID       int       `json:"order_id" db:"order_id"`
	ReceiptNumber string    `json:"receipt_number" db:"receipt_number"`
	FiscalSign    string    `json:"fiscal_sign" db:"fiscal_sign"`
	TotalAmount   float64   `json:"total_amount" db:"total_amount"`
	VATRate       float64   `json:"vat_rate" db:"vat_rate"`
	VATAmount     float64   `json:"vat_amount" db:"vat_amount"`
	Subtotal      float64   `json:"subtotal" db:"subtotal"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"`
	INN           string    `json:"inn" db:"inn"`
	CompanyName   string    `json:"company_name" db:"company_name"`
	CashierName   string    `json:"cashier_name" db:"cashier_name"`
	Status        string    `json:"status" db:"status"` // local, sent, confirmed, error
	OFDResponse   string    `json:"ofd_response" db:"ofd_response"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// FiscalSettings holds fiscal configuration
type FiscalSettings struct {
	Enabled     bool    `json:"enabled"`
	VATRate     float64 `json:"vat_rate"`
	INN         string  `json:"inn"`
	CompanyName string  `json:"company_name"`
}

// FiscalReceiptItem is used for printing fiscal receipt details
type FiscalReceiptItem struct {
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`
	Total    float64 `json:"total"`
	VAT      float64 `json:"vat"`
}

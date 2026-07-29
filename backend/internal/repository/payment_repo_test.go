package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/username/kafe-backend/internal/models"
)

func setupPaymentTestDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, nil
}

func TestAddPayments(t *testing.T) {
	db, mock, err := setupPaymentTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPaymentRepository(db)
	orderID := 1
	payments := []models.PaymentInput{
		{Method: "cash", Amount: 50000},
		{Method: "card", Amount: 20000},
	}

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM order_payments WHERE order_id = \\$1").
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec("INSERT INTO order_payments \\(order_id, method, amount\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs(orderID, "cash", 50000.0).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO order_payments \\(order_id, method, amount\\) VALUES \\(\\$1, \\$2, \\$3\\)").
		WithArgs(orderID, "card", 20000.0).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectExec("UPDATE orders SET payment_method = \\$1, updated_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs("mixed", orderID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	mock.ExpectCommit()

	err = repo.AddPayments(orderID, payments)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAddPayments_ValidationErrors(t *testing.T) {
	db, _, _ := setupPaymentTestDB()
	defer db.Close()
	repo := NewPaymentRepository(db)

	err := repo.AddPayments(1, []models.PaymentInput{})
	assert.ErrorContains(t, err, "to'lov ma'lumotlari bo'sh")

	err = repo.AddPayments(1, []models.PaymentInput{{Method: "invalid", Amount: 100}})
	assert.ErrorContains(t, err, "noto'g'ri to'lov usuli")

	err = repo.AddPayments(1, []models.PaymentInput{{Method: "cash", Amount: -10}})
	assert.ErrorContains(t, err, "to'lov summasi musbat bo'lishi kerak")
}

func TestGetPaymentsByOrderID(t *testing.T) {
	db, mock, err := setupPaymentTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPaymentRepository(db)
	orderID := 1
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "order_id", "method", "amount", "created_at"}).
		AddRow(1, orderID, "cash", 50000, now).
		AddRow(2, orderID, "card", 20000, now)

	mock.ExpectQuery("SELECT id, order_id, method, amount, created_at FROM order_payments WHERE order_id = \\$1 ORDER BY created_at").
		WithArgs(orderID).
		WillReturnRows(rows)

	payments, err := repo.GetPaymentsByOrderID(orderID)
	assert.NoError(t, err)
	assert.Len(t, payments, 2)
	assert.Equal(t, "cash", payments[0].Method)
	assert.Equal(t, 50000.0, payments[0].Amount)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetPaymentSummary(t *testing.T) {
	db, mock, err := setupPaymentTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPaymentRepository(db)
	orderID := 1

	rows := sqlmock.NewRows([]string{"method", "total"}).
		AddRow("cash", 70000.0).
		AddRow("card", 30000.0)

	mock.ExpectQuery("SELECT method, SUM\\(amount\\) as total FROM order_payments WHERE order_id = \\$1 GROUP BY method").
		WithArgs(orderID).
		WillReturnRows(rows)

	summary, err := repo.GetPaymentSummary(orderID)
	assert.NoError(t, err)
	assert.Equal(t, 70000.0, summary["cash"])
	assert.Equal(t, 30000.0, summary["card"])

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/username/kafe-backend/internal/models"
)

func setupRefundTestDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, nil
}

func TestCreateRefund(t *testing.T) {
	db, mock, err := setupRefundTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRefundRepository(db)
	now := time.Now()

	refund := &models.Refund{
		OrderID:         100,
		Amount:          50000,
		Reason:          "Mahsulot sifati yomon",
		RequestedBy:     func() *int { id := 2; return &id }(),
		RequestedByName: "Waiter A",
	}

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now)

	mock.ExpectQuery("INSERT INTO refunds .* RETURNING id, created_at").
		WithArgs(refund.OrderID, refund.Amount, refund.Reason, refund.ReasonDetail, refund.RequestedBy, refund.RequestedByName).
		WillReturnRows(rows)

	err = repo.CreateRefund(refund)
	assert.NoError(t, err)
	assert.Equal(t, 1, refund.ID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestApproveRefund(t *testing.T) {
	db, mock, err := setupRefundTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRefundRepository(db)

	// ApproveRefund now uses a transaction
	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE refunds SET status = 'approved', approved_by = \\$1, approved_by_name = \\$2, refund_method = \\$3, resolved_at = \\$4 WHERE id = \\$5 AND status = 'pending' RETURNING amount").
		WithArgs(1, "Admin X", "cash", sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows([]string{"amount"}).AddRow(500.0))
	// Expect shift lookup (no open shift found)
	mock.ExpectQuery("SELECT id FROM cashier_shifts WHERE cashier_id = .* AND status = 'open'").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	err = repo.ApproveRefund(10, 1, "Admin X", "cash")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestApproveRefund_NotFound(t *testing.T) {
	db, mock, err := setupRefundTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRefundRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery("UPDATE refunds SET status = 'approved', approved_by = \\$1, approved_by_name = \\$2, refund_method = \\$3, resolved_at = \\$4 WHERE id = \\$5 AND status = 'pending' RETURNING amount").
		WithArgs(1, "Admin X", "cash", sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows([]string{"amount"}))
	mock.ExpectRollback()

	err = repo.ApproveRefund(10, 1, "Admin X", "cash")
	assert.ErrorContains(t, err, "refund topilmadi yoki allaqachon qayta ishlangan")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestRejectRefund(t *testing.T) {
	db, mock, err := setupRefundTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRefundRepository(db)

	mock.ExpectExec("UPDATE refunds SET status = 'rejected', approved_by = \\$1, approved_by_name = \\$2, resolved_at = \\$3 WHERE id = \\$4 AND status = 'pending'").
		WithArgs(1, "Admin X", sqlmock.AnyArg(), 10).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.RejectRefund(10, 1, "Admin X")
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetPendingRefunds(t *testing.T) {
	db, mock, err := setupRefundTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewRefundRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{
		"id", "order_id", "amount", "reason", "reason_detail", "status", "refund_method",
		"money_returned", "requested_by", "requested_by_name", "approved_by", "approved_by_name",
		"created_at", "resolved_at",
	}).AddRow(1, 100, 50000, "Xato buyurtma", "", "pending", "", false, 2, "Waiter", nil, "", now, nil)

	mock.ExpectQuery("SELECT id, order_id, amount, reason, COALESCE\\(reason_detail, ''\\) as reason_detail, status, COALESCE\\(refund_method, ''\\) as refund_method, money_returned, requested_by, COALESCE\\(requested_by_name, ''\\) as requested_by_name, approved_by, COALESCE\\(approved_by_name, ''\\) as approved_by_name, created_at, resolved_at FROM refunds WHERE status = 'pending' ORDER BY created_at DESC").
		WillReturnRows(rows)

	refunds, err := repo.GetPendingRefunds()
	assert.NoError(t, err)
	assert.Len(t, refunds, 1)
	assert.Equal(t, models.RefundStatus("pending"), refunds[0].Status)
	assert.Equal(t, 50000.0, refunds[0].Amount)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

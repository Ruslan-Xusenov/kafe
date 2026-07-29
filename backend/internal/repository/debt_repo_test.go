package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/username/kafe-backend/internal/models"
)

func setupDebtTestDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, nil
}

func TestCreateDebtor(t *testing.T) {
	db, mock, err := setupDebtTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewDebtRepository(db)
	now := time.Now()

	debtor := &models.Debtor{
		Name:  "John Doe",
		Phone: "123456789",
		Notes: "Test note",
	}

	rows := sqlmock.NewRows([]string{"id", "total_debt", "created_at", "updated_at"}).
		AddRow(1, 0.0, now, now)

	mock.ExpectQuery("INSERT INTO debtors \\(name, phone, notes\\) VALUES \\(\\$1, \\$2, \\$3\\) RETURNING id, total_debt, created_at, updated_at").
		WithArgs(debtor.Name, debtor.Phone, debtor.Notes).
		WillReturnRows(rows)

	err = repo.CreateDebtor(debtor)
	assert.NoError(t, err)
	assert.Equal(t, 1, debtor.ID)
	assert.Equal(t, 0.0, debtor.TotalDebt)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestGetAllDebtors(t *testing.T) {
	db, mock, err := setupDebtTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewDebtRepository(db)
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "name", "phone", "total_debt", "notes", "created_at", "updated_at"}).
		AddRow(1, "John", "111", 50000, "", now, now).
		AddRow(2, "Jane", "222", 0, "", now, now)

	// Test with onlyWithDebt = false
	mock.ExpectQuery("SELECT id, name, COALESCE\\(phone, ''\\) as phone, total_debt, COALESCE\\(notes, ''\\) as notes, created_at, updated_at FROM debtors ORDER BY total_debt DESC, name ASC").
		WillReturnRows(rows)

	debtors, err := repo.GetAllDebtors(false)
	assert.NoError(t, err)
	assert.Len(t, debtors, 2)

	// Test with onlyWithDebt = true
	rowsWithDebt := sqlmock.NewRows([]string{"id", "name", "phone", "total_debt", "notes", "created_at", "updated_at"}).
		AddRow(1, "John", "111", 50000, "", now, now)

	mock.ExpectQuery("SELECT id, name, COALESCE\\(phone, ''\\) as phone, total_debt, COALESCE\\(notes, ''\\) as notes, created_at, updated_at FROM debtors WHERE total_debt > 0 ORDER BY total_debt DESC, name ASC").
		WillReturnRows(rowsWithDebt)

	debtorsDebt, err := repo.GetAllDebtors(true)
	assert.NoError(t, err)
	assert.Len(t, debtorsDebt, 1)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAddDebtRecord_Debt(t *testing.T) {
	db, mock, err := setupDebtTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewDebtRepository(db)
	now := time.Now()

	record := &models.DebtRecord{
		DebtorID:      1,
		OrderID:       func() *int { id := 10; return &id }(),
		Amount:        30000,
		Type:          "debt",
		PaymentMethod: "nasiya",
		CreatedBy:     func() *int { id := 1; return &id }(),
	}

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(1, now)

	mock.ExpectQuery("INSERT INTO debt_records \\(debtor_id, order_id, amount, type, payment_method, description, created_by\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7\\) RETURNING id, created_at").
		WithArgs(record.DebtorID, record.OrderID, record.Amount, record.Type, record.PaymentMethod, record.Description, record.CreatedBy).
		WillReturnRows(rows)

	mock.ExpectExec("UPDATE debtors SET total_debt = total_debt \\+ \\$1 WHERE id = \\$2").
		WithArgs(record.Amount, record.DebtorID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.AddDebtRecord(record)
	assert.NoError(t, err)
	assert.Equal(t, 1, record.ID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestAddDebtRecord_Payment(t *testing.T) {
	db, mock, err := setupDebtTestDB()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewDebtRepository(db)
	now := time.Now()

	record := &models.DebtRecord{
		DebtorID:      1,
		Amount:        20000,
		Type:          "payment",
		PaymentMethod: "cash",
		CreatedBy:     func() *int { id := 1; return &id }(),
	}

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"id", "created_at"}).AddRow(2, now)

	mock.ExpectQuery("INSERT INTO debt_records \\(debtor_id, order_id, amount, type, payment_method, description, created_by\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6, \\$7\\) RETURNING id, created_at").
		WithArgs(record.DebtorID, record.OrderID, record.Amount, record.Type, record.PaymentMethod, record.Description, record.CreatedBy).
		WillReturnRows(rows)

	mock.ExpectExec("UPDATE debtors SET total_debt = total_debt - \\$1 WHERE id = \\$2").
		WithArgs(record.Amount, record.DebtorID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.AddDebtRecord(record)
	assert.NoError(t, err)
	assert.Equal(t, 2, record.ID)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

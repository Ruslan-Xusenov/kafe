package service

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

func setupOrderServiceTestDB() (*sqlx.DB, sqlmock.Sqlmock, error) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	return sqlxDB, mock, nil
}

func setupOrderService(db *sqlx.DB) *OrderService {
	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	tableRepo := repository.NewTableRepository(db)

	wsService := NewWebsocketService()
	botService := NewBotService()
	printerService := NewPrinterService()

	return NewOrderService(orderRepo, productRepo, settingsRepo, inventoryRepo, tableRepo, wsService, botService, printerService)
}

func TestSetServiceFee(t *testing.T) {
	db, mock, err := setupOrderServiceTestDB()
	assert.NoError(t, err)
	defer db.Close()

	svc := setupOrderService(db)

	orderID := 1
	tableID := 5

	// Mock GetByID for fetching order
	orderRows := sqlmock.NewRows([]string{
		"id", "table_id", "total_price", "service_fee", "service_percentage", "created_at", "status",
	}).AddRow(orderID, tableID, 100000.0, 0.0, 0.0, time.Now(), "new")

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(orderRows)
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(10, orderID))

	// Mock SetServiceFee
	mock.ExpectExec(`UPDATE orders SET service_percentage = \$1, service_fee = \$2, total_price = \$3, updated_at = NOW\(\) WHERE id = \$4`).
		WithArgs(15.0, 15000.0, 115000.0, orderID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Mock GetByID for returning updated order
	updatedRows := sqlmock.NewRows([]string{
		"id", "table_id", "total_price", "service_fee", "service_percentage", "created_at", "status",
	}).AddRow(orderID, tableID, 115000.0, 15000.0, 15.0, time.Now(), "new")

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(updatedRows)
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(10, orderID))

	updatedOrder, err := svc.SetServiceFee(orderID, 15.0)
	assert.NoError(t, err)
	assert.NotNil(t, updatedOrder)
	if updatedOrder != nil {
		assert.Equal(t, 115000.0, updatedOrder.TotalPrice)
		assert.Equal(t, 15000.0, updatedOrder.ServiceFee)
	}

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestSetServiceFee_DeliveryOrderError(t *testing.T) {
	db, mock, err := setupOrderServiceTestDB()
	assert.NoError(t, err)
	defer db.Close()

	svc := setupOrderService(db)

	orderID := 2

	// Mock GetByID for delivery order (table_id = null)
	orderRows := sqlmock.NewRows([]string{
		"id", "table_id", "total_price", "service_fee", "service_percentage", "created_at", "status",
	}).AddRow(orderID, nil, 100000.0, 0.0, 0.0, time.Now(), "new")

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(orderRows)
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(10, orderID))

	_, err = svc.SetServiceFee(orderID, 15.0)
	assert.ErrorContains(t, err, "плата за обслуживание применяется только к внутренним заказам")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

func TestCloseTableWithPayments(t *testing.T) {
	db, mock, err := setupOrderServiceTestDB()
	assert.NoError(t, err)
	defer db.Close()

	svc := setupOrderService(db)

	tableID := 5
	userID := 1
	role := "admin"

	// Mock GetAll to find active orders for table
	allOrdersRows := sqlmock.NewRows([]string{
		"id", "table_id", "total_price", "status",
	}).AddRow(100, tableID, 50000.0, "ready").
		AddRow(101, tableID, 30000.0, "ready").
		AddRow(102, 6, 20000.0, "ready") // Different table

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*ORDER BY.*`).
		WillReturnRows(allOrdersRows)

	// In CloseTableWithPayments, it fetches full order details to calc grand total
	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "total_price", "service_fee"}).AddRow(100, 50000.0, 0.0))
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(1, 100))
	
	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(101).
		WillReturnRows(sqlmock.NewRows([]string{"id", "total_price", "service_fee"}).AddRow(101, 30000.0, 0.0))
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(101).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(2, 101))

	// For o := range tableOrders loop
	// grandTotal = 80000, order 100 share = 50000/80000 = 62.5%, order 101 share = 37.5%
	// Order 100
	mock.ExpectExec("UPDATE orders SET payment_method = \\$1, updated_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs("mixed", 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Proportional payments for order 100: cash=50000*0.625=31250, card=30000*0.625=18750
	mock.ExpectExec("INSERT INTO order_payments").
		WithArgs(100, "cash", 31250.0).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO order_payments").
		WithArgs(100, "card", 18750.0).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = NOW\(\) WHERE id = \$2`).
		WithArgs("delivered", 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Order 101
	mock.ExpectExec("UPDATE orders SET payment_method = \\$1, updated_at = NOW\\(\\) WHERE id = \\$2").
		WithArgs("mixed", 101).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Proportional payments for order 101: cash=50000*0.375=18750, card=30000*0.375=11250
	mock.ExpectExec("INSERT INTO order_payments").
		WithArgs(101, "cash", 18750.0).
		WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectExec("INSERT INTO order_payments").
		WithArgs(101, "card", 11250.0).
		WillReturnResult(sqlmock.NewResult(4, 1))

	mock.ExpectExec(`UPDATE orders SET status = \$1, updated_at = NOW\(\) WHERE id = \$2`).
		WithArgs("delivered", 101).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Finally it gets populated orders for WS/Printer
	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "table_id", "total_price", "service_fee"}).AddRow(100, tableID, 50000.0, 0.0))
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(1, 100))

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "table_id", "total_price", "service_fee"}).AddRow(100, tableID, 50000.0, 0.0))
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(100).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(1, 100))

	mock.ExpectQuery(`(?i)SELECT .* FROM orders.*WHERE.*\$1`).
		WithArgs(101).
		WillReturnRows(sqlmock.NewRows([]string{"id", "table_id", "total_price", "service_fee"}).AddRow(101, tableID, 30000.0, 0.0))
	mock.ExpectQuery(`(?i)SELECT .* FROM order_items.*WHERE.*\$1`).
		WithArgs(101).
		WillReturnRows(sqlmock.NewRows([]string{"id", "order_id"}).AddRow(2, 101))

	err = svc.CloseTableWithPayments(tableID, []models.PaymentInput{
		{Method: "cash", Amount: 50000},
		{Method: "card", Amount: 30000},
	}, userID, role)
	assert.NoError(t, err)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}

package service

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/repository"
)

type FiscalService struct {
	fiscalRepo *repository.FiscalRepository
	settingsRepo *repository.SettingsRepository
}

func NewFiscalService(fiscalRepo *repository.FiscalRepository, settingsRepo *repository.SettingsRepository) *FiscalService {
	return &FiscalService{
		fiscalRepo:   fiscalRepo,
		settingsRepo: settingsRepo,
	}
}

// GetSettings loads fiscal settings from the settings table
func (s *FiscalService) GetSettings() (*models.FiscalSettings, error) {
	settings := &models.FiscalSettings{}

	enabledStr, _ := s.settingsRepo.Get("fiscal_enabled")
	settings.Enabled = enabledStr == "true"

	vatStr, _ := s.settingsRepo.Get("fiscal_vat_rate")
	if vatStr != "" {
		if v, err := strconv.ParseFloat(vatStr, 64); err == nil {
			settings.VATRate = v
		}
	} else {
		settings.VATRate = 12.0
	}

	settings.INN, _ = s.settingsRepo.Get("fiscal_inn")
	settings.CompanyName, _ = s.settingsRepo.Get("fiscal_company_name")

	// Fallback to env vars
	if settings.CompanyName == "" {
		settings.CompanyName = os.Getenv("CAFE_FULL_NAME")
	}

	return settings, nil
}

// UpdateSettings saves fiscal settings
func (s *FiscalService) UpdateSettings(fs *models.FiscalSettings) error {
	enabledStr := "false"
	if fs.Enabled {
		enabledStr = "true"
	}
	if err := s.settingsRepo.Set("fiscal_enabled", enabledStr); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("fiscal_vat_rate", fmt.Sprintf("%.2f", fs.VATRate)); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("fiscal_inn", fs.INN); err != nil {
		return err
	}
	if err := s.settingsRepo.Set("fiscal_company_name", fs.CompanyName); err != nil {
		return err
	}
	return nil
}

// CreateReceiptForOrder generates a fiscal receipt when an order is closed
func (s *FiscalService) CreateReceiptForOrder(order *models.Order, paymentMethod string, cashierName string) (*models.FiscalReceipt, error) {
	settings, err := s.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get fiscal settings: %w", err)
	}

	if !settings.Enabled {
		log.Println("Fiscal receipts are disabled, skipping...")
		return nil, nil
	}

	// Generate receipt number
	receiptNumber, err := s.fiscalRepo.GenerateReceiptNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate receipt number: %w", err)
	}

	// Calculate VAT
	vatRate := settings.VATRate
	if vatRate <= 0 {
		vatRate = 12.0
	}

	totalAmount := order.TotalPrice
	// VAT is inclusive in Uzbekistan: subtotal = total / (1 + vatRate/100)
	subtotal := totalAmount / (1 + vatRate/100)
	vatAmount := totalAmount - subtotal

	receipt := &models.FiscalReceipt{
		OrderID:       order.ID,
		ReceiptNumber: receiptNumber,
		TotalAmount:   totalAmount,
		VATRate:       vatRate,
		VATAmount:     vatAmount,
		Subtotal:      subtotal,
		PaymentMethod: paymentMethod,
		INN:           settings.INN,
		CompanyName:   settings.CompanyName,
		CashierName:   cashierName,
		Status:        "local",
		OFDResponse:   "{}",
	}

	if err := s.fiscalRepo.CreateReceipt(receipt); err != nil {
		return nil, fmt.Errorf("failed to create fiscal receipt: %w", err)
	}

	log.Printf("✅ Fiscal receipt %s created for order #%d (VAT: %.0f, Total: %.0f)",
		receiptNumber, order.ID, vatAmount, totalAmount)

	return receipt, nil
}

// GetReceiptByOrderID returns fiscal receipt for a specific order
func (s *FiscalService) GetReceiptByOrderID(orderID int) (*models.FiscalReceipt, error) {
	return s.fiscalRepo.GetByOrderID(orderID)
}

// GetAllReceipts returns all fiscal receipts
func (s *FiscalService) GetAllReceipts(limit, offset int, status string) ([]models.FiscalReceipt, error) {
	return s.fiscalRepo.GetAll(limit, offset, status)
}

// GetStats returns fiscal statistics
func (s *FiscalService) GetStats() (map[string]interface{}, error) {
	return s.fiscalRepo.GetStats()
}

// ResendToOFD will attempt to send a receipt to OFD (adapter pattern for future OFD integration)
func (s *FiscalService) ResendToOFD(receiptID int) error {
	// This is a placeholder for OFD integration.
	// When a specific OFD provider is chosen (Billz, E-POS, Soliq.uz),
	// implement the adapter here.
	log.Printf("OFD integration not configured yet. Receipt #%d marked for manual sync.", receiptID)
	return s.fiscalRepo.UpdateStatus(receiptID, "local", "", "{\"note\": \"OFD not configured\"}")
}

package service

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/username/kafe-backend/internal/models"
)

// printJob is a single item in the printer queue.
type printJob struct {
	order   *models.Order
	receipt *models.FiscalReceipt // nil for regular orders
}

// PrinterService manages a background TCP printer with retry and deduplication.
type PrinterService struct {
	IP      string
	Port    string
	Enabled bool

	queue chan printJob

	// recentPrints tracks recently printed order IDs to prevent duplicates.
	recentMu     sync.Mutex
	recentPrints map[int]time.Time
}

// NewPrinterService creates the printer service and starts the background worker.
func NewPrinterService() *PrinterService {
	port := os.Getenv("PRINTER_PORT")
	if port == "" {
		port = "9100"
	}
	s := &PrinterService{
		IP:           os.Getenv("PRINTER_IP"),
		Port:         port,
		Enabled:      os.Getenv("PRINTER_ENABLED") == "true",
		queue:        make(chan printJob, 64),
		recentPrints: make(map[int]time.Time),
	}
	go s.worker()
	return s
}

// PrintOrder enqueues an order for printing. Non-blocking.
func (s *PrinterService) PrintOrder(order *models.Order) {
	if !s.Enabled || s.IP == "" {
		return
	}
	select {
	case s.queue <- printJob{order: order}:
	default:
		fmt.Printf("⚠️  [PRINTER] Queue full, dropping print job for order #%d\n", order.ID)
	}
}

// PrintFiscalReceipt enqueues a fiscal receipt for printing.
func (s *PrinterService) PrintFiscalReceipt(order *models.Order, receipt *models.FiscalReceipt) {
	if !s.Enabled || s.IP == "" {
		return
	}
	select {
	case s.queue <- printJob{order: order, receipt: receipt}:
	default:
		fmt.Printf("⚠️  [PRINTER] Queue full, dropping fiscal receipt for order #%d\n", order.ID)
	}
}

// worker processes the print queue sequentially.
func (s *PrinterService) worker() {
	for job := range s.queue {
		if job.receipt != nil {
			s.printFiscalWithRetry(job.order, job.receipt)
		} else {
			s.printOrderWithRetry(job.order)
		}
	}
}

// printOrderWithRetry attempts to print up to 3 times with exponential backoff.
func (s *PrinterService) printOrderWithRetry(order *models.Order) {
	// Deduplication: skip if same order was printed within the last 5 seconds.
	if order.ID > 0 {
		s.recentMu.Lock()
		if last, ok := s.recentPrints[order.ID]; ok && time.Since(last) < 5*time.Second {
			s.recentMu.Unlock()
			fmt.Printf("🔁 [PRINTER] Skipping duplicate print for order #%d (printed %.1fs ago)\n", order.ID, time.Since(last).Seconds())
			return
		}
		s.recentPrints[order.ID] = time.Now()
		// Clean up old entries
		for id, t := range s.recentPrints {
			if time.Since(t) > 30*time.Second {
				delete(s.recentPrints, id)
			}
		}
		s.recentMu.Unlock()
	}

	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := s.doPrintOrder(order)
		if err == nil {
			if attempt > 1 {
				fmt.Printf("✅ [PRINTER] Order #%d printed successfully on attempt %d\n", order.ID, attempt)
			}
			return
		}
		fmt.Printf("❌ [PRINTER] Attempt %d/%d failed for order #%d: %v\n", attempt, maxAttempts, order.ID, err)
		if attempt < maxAttempts {
			backoff := time.Duration(attempt*attempt) * time.Second
			time.Sleep(backoff)
		}
	}
	fmt.Printf("🚫 [PRINTER] All %d attempts failed for order #%d. Order not printed.\n", maxAttempts, order.ID)
}

// printFiscalWithRetry attempts to print fiscal receipt up to 3 times.
func (s *PrinterService) printFiscalWithRetry(order *models.Order, receipt *models.FiscalReceipt) {
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := s.doPrintFiscal(order, receipt)
		if err == nil {
			fmt.Printf("🧾 [PRINTER] Fiscal receipt %s printed for order #%d\n", receipt.ReceiptNumber, order.ID)
			return
		}
		fmt.Printf("❌ [PRINTER] Fiscal attempt %d/%d failed for order #%d: %v\n", attempt, maxAttempts, order.ID, err)
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}
	}
}

// doPrintOrder opens a TCP connection and sends ESC/POS data for a regular order.
func (s *PrinterService) doPrintOrder(order *models.Order) error {
	address := net.JoinHostPort(s.IP, s.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("TCP connection failed (%s): %w", address, err)
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	// Initialize
	conn.Write(ESC_INIT)
	conn.Write(BEEP)

	// Header
	cafeName := os.Getenv("CAFE_NAME")
	if cafeName == "" {
		cafeName = "KAFE"
	}
	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte(cafeName + " ONLINE ZAKAZ\n"))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("------------------------------------------------\n"))

	// Details
	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Chek №: %d\n", order.ID)))

	if order.TableName != nil && *order.TableName != "" {
		conn.Write([]byte(fmt.Sprintf("Stol: %s\n", *order.TableName)))
	} else {
		conn.Write([]byte("Stol: online\n"))
	}
	if order.WaiterName != "" {
		conn.Write([]byte(fmt.Sprintf("Ofitsiant: %s\n", s.transliterate(order.WaiterName))))
	}

	cafeFullName := os.Getenv("CAFE_FULL_NAME")
	if cafeFullName == "" {
		cafeFullName = cafeName
	}
	conn.Write([]byte(fmt.Sprintf("Kafe: %s\n", s.transliterate(cafeFullName))))

	createdAt := order.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	conn.Write([]byte(fmt.Sprintf("Ochilgan: %s\n", createdAt.Format("02.01.2006 15:04:05"))))
	conn.Write([]byte("------------------------------------------------\n"))

	// Items Table
	conn.Write([]byte("Nomi                   Soni   Narxi      Jami\n"))
	conn.Write([]byte("------------------------------------------------\n"))

	for _, item := range order.Items {
		name := s.transliterate(item.ProductName)
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		line := fmt.Sprintf("%-22s %-6.1f %-10.0f %-10.0f\n",
			name, item.Quantity, item.Price, item.Price*item.Quantity)
		conn.Write([]byte(line))

		if item.Comment != "" {
			conn.Write([]byte(fmt.Sprintf("  * Izoh: %s\n", s.transliterate(item.Comment))))
		}
	}
	conn.Write([]byte("------------------------------------------------\n"))

	// Footer
	conn.Write(ALIGN_RIGHT)

	if order.ServiceFee > 0 {
		subtotal := order.TotalPrice - order.ServiceFee
		conn.Write([]byte(fmt.Sprintf("Oraliq: %.0f\n", subtotal)))
		conn.Write([]byte(fmt.Sprintf("Xizmat(%.0f%%): %.0f\n", order.ServicePercentage, order.ServiceFee)))
	} else {
		conn.Write([]byte(fmt.Sprintf("Oraliq: %.0f\n", order.TotalPrice)))
		conn.Write([]byte("Xizmat(0%): 0\n"))
	}
	conn.Write([]byte("Chegirma(0%): 0\n\n"))

	conn.Write(FONT_DOUBLE_W)
	conn.Write([]byte(fmt.Sprintf("Jami: %.0f\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n\n\n"))

	// Payment method if available
	if order.PaymentMethod != "" {
		conn.Write(ALIGN_CENTER)
		conn.Write([]byte(fmt.Sprintf("To'lov: %s\n\n", s.paymentMethodLabel(order.PaymentMethod))))
	}

	conn.Write(PAPER_CUT)
	return nil
}

// doPrintFiscal sends a fiscal receipt over TCP.
func (s *PrinterService) doPrintFiscal(order *models.Order, receipt *models.FiscalReceipt) error {
	address := net.JoinHostPort(s.IP, s.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("TCP connection failed (%s): %w", address, err)
	}
	defer conn.Close()
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	conn.Write(ESC_INIT)
	conn.Write(BEEP)

	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	companyName := receipt.CompanyName
	if companyName == "" {
		companyName = os.Getenv("CAFE_NAME")
	}
	conn.Write([]byte(companyName + "\n"))
	conn.Write(FONT_NORMAL)

	if receipt.INN != "" {
		conn.Write([]byte(fmt.Sprintf("INN: %s\n", receipt.INN)))
	}
	cafeAddress := os.Getenv("CAFE_ADDRESS")
	if cafeAddress != "" {
		conn.Write([]byte(s.transliterate(cafeAddress) + "\n"))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_LEFT)
	conn.Write(FONT_DOUBLE_H)
	conn.Write([]byte("FISKAL CHEK\n"))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte(fmt.Sprintf("Chek raqami: %s\n", receipt.ReceiptNumber)))
	conn.Write([]byte(fmt.Sprintf("Buyurtma: #%d\n", order.ID)))
	conn.Write([]byte(fmt.Sprintf("Kassir: %s\n", s.transliterate(receipt.CashierName))))
	conn.Write([]byte(fmt.Sprintf("Sana: %s\n", receipt.CreatedAt.Format("02.01.2006 15:04:05"))))
	conn.Write([]byte(fmt.Sprintf("To'lov usuli: %s\n", s.paymentMethodLabel(receipt.PaymentMethod))))
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write([]byte("Nomi                   Soni  Narxi      Jami\n"))
	conn.Write([]byte("------------------------------------------------\n"))

	for _, item := range order.Items {
		name := s.transliterate(item.ProductName)
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		line := fmt.Sprintf("%-22s %-5.1f %-10.0f %-10.0f\n",
			name, item.Quantity, item.Price, item.Price*item.Quantity)
		conn.Write([]byte(line))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_RIGHT)
	conn.Write([]byte(fmt.Sprintf("QQSsiz summa:    %.0f\n", receipt.Subtotal)))
	conn.Write([]byte(fmt.Sprintf("QQS (%.0f%%):       %.0f\n", receipt.VATRate, receipt.VATAmount)))

	if order.ServiceFee > 0 {
		conn.Write([]byte(fmt.Sprintf("Xizmat haqi:     %.0f\n", order.ServiceFee)))
	}

	conn.Write([]byte("------------------------------------------------\n"))
	conn.Write(FONT_BIG)
	conn.Write([]byte(fmt.Sprintf("JAMI: %.0f\n", receipt.TotalAmount)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_CENTER)
	if receipt.FiscalSign != "" {
		conn.Write([]byte(fmt.Sprintf("Fiskal belgi: %s\n", receipt.FiscalSign)))
	}
	conn.Write([]byte(fmt.Sprintf("Status: %s\n", receipt.Status)))
	conn.Write([]byte("\nXaridingiz uchun rahmat!\n\n\n\n"))

	conn.Write(PAPER_CUT)
	return nil
}

// ESC/POS Commands
var (
	ESC_INIT     = []byte{0x1B, 0x40}            // ESC @
	ALIGN_LEFT   = []byte{0x1B, 0x61, 0x00}      // ESC a 0
	ALIGN_CENTER = []byte{0x1B, 0x61, 0x01}      // ESC a 1
	ALIGN_RIGHT  = []byte{0x1B, 0x61, 0x02}      // ESC a 2
	FONT_NORMAL  = []byte{0x1D, 0x21, 0x00}      // GS ! 0
	FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}     // GS ! 1
	FONT_DOUBLE_W = []byte{0x1D, 0x21, 0x10}     // GS ! 16
	FONT_BIG     = []byte{0x1D, 0x21, 0x11}      // GS ! 17 (Double width + height)
	PAPER_CUT    = []byte{0x1D, 0x56, 0x42, 0x00} // GS V B 0
	BEEP         = []byte{0x1B, 0x42, 0x02, 0x02} // ESC B n t (Beep 2 times)
)

func (s *PrinterService) paymentMethodLabel(method string) string {
	switch method {
	case "cash":
		return "Naqd"
	case "card":
		return "Karta"
	case "click":
		return "Click/Payme"
	case "nasiya":
		return "Nasiya"
	case "mixed":
		return "Aralash"
	default:
		return method
	}
}

func (s *PrinterService) transliterate(text string) string {
	r := strings.NewReplacer(
		"ў", "o'", "Ў", "O'",
		"қ", "q", "Қ", "Q",
		"ғ", "g'", "Ғ", "G'",
		"ҳ", "h", "Ҳ", "H",
		"\u2018", "'", "\u2019", "'",
	)
	return r.Replace(text)
}
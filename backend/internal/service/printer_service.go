package service

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/username/kafe-backend/internal/models"
)

type PrinterService struct {
	IP      string
	Port    string
	Enabled bool
}

func NewPrinterService() *PrinterService {
	port := os.Getenv("PRINTER_PORT")
	if port == "" {
		port = "9100"
	}
	return &PrinterService{
		IP:      os.Getenv("PRINTER_IP"),
		Port:    port,
		Enabled: os.Getenv("PRINTER_ENABLED") == "true",
	}
}

// ESC/POS Commands
var (
	ESC_INIT      = []byte{0x1B, 0x40}           // ESC @
	ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}      // ESC a 0
	ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}      // ESC a 1
	ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}      // ESC a 2
	FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}      // GS ! 0
	FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}      // GS ! 1
	FONT_DOUBLE_W = []byte{0x1D, 0x21, 0x10}      // GS ! 16
	FONT_BIG      = []byte{0x1D, 0x21, 0x11}      // GS ! 17 (Double width + height)
	PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00} // GS V B 0
	BEEP          = []byte{0x1B, 0x42, 0x02, 0x02} // ESC B n t (Beep 2 times)
)

func (s *PrinterService) PrintOrder(order *models.Order) {
	if !s.Enabled || s.IP == "" {
		return
	}

	address := net.JoinHostPort(s.IP, s.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		fmt.Printf("Printer connection error (%s): %v\n", address, err)
		return
	}
	defer conn.Close()

	// Initialize
	conn.Write(ESC_INIT)
	conn.Write(BEEP)

	// Header - YETUK ONLAYN ZAKAZ (Centered and Big)
	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte("YETUK ONLAYN ZAKAZ\n"))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("------------------------------------------------\n"))
	
	// Details
	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Чек №: %d\n", order.ID)))
	conn.Write([]byte(fmt.Sprintf("Зал: Доставка №%d\n", order.ID)))
	conn.Write([]byte("Стол: onlayn\n"))
	conn.Write([]byte("Обслужил: YETUK KAFE onlayn\n"))
	
	createdAt := order.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	conn.Write([]byte(fmt.Sprintf("Время открытия: %s\n", createdAt.Format("02.01.2006 15:04:05"))))
	conn.Write([]byte("Время закрытия: -\n"))
	conn.Write([]byte("------------------------------------------------\n"))

	// Items Table
	// Name: 22, Qty: 6, Price: 10, Total: 10
	conn.Write([]byte("Наименование           Soni   Narxi      Jami\n"))
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

	// Footer Summary
	conn.Write(ALIGN_RIGHT)
	conn.Write([]byte(fmt.Sprintf("Подитог: %.0f\n", order.TotalPrice)))
	conn.Write([]byte("Обслуживание(0.0%): 0\n"))
	conn.Write([]byte("Скидка(0%): 0\n"))
	conn.Write([]byte("\n"))
	
	conn.Write(FONT_DOUBLE_W)
	conn.Write([]byte(fmt.Sprintf("Итого: %.0f\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n\n\n"))

	// Cut
	conn.Write(PAPER_CUT)
}

func (s *PrinterService) transliterate(text string) string {
	r := strings.NewReplacer(
		"ў", "o'", "Ў", "O'",
		"қ", "q", "Қ", "Q",
		"ғ", "g'", "Ғ", "G'",
		"ҳ", "h", "Ҳ", "H",
		"‘", "'", "’", "'",
	)
	return r.Replace(text)
}
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

	// Header
	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte("KAFE\n")) // General name, can be customized later
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("------------------------------------------------\n"))
	
	conn.Write(FONT_DOUBLE_H)
	conn.Write([]byte(fmt.Sprintf("BUYURTMA #%d\n", order.ID)))
	conn.Write(FONT_NORMAL)
	
	createdAt := order.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	conn.Write([]byte(fmt.Sprintf("Sana: %s\n", createdAt.Format("02.01.2006 15:04"))))
	conn.Write([]byte("------------------------------------------------\n"))

	// Customer Info
	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Mijoz: %s\n", order.Phone)))
	conn.Write([]byte(fmt.Sprintf("Manzil: %s\n", s.transliterate(order.Address))))
	if order.Comment != "" {
		conn.Write([]byte(fmt.Sprintf("Izoh: %s\n", s.transliterate(order.Comment))))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	// Items Header
	// Total 48 chars
	// Name: 22, Qty: 5, Price: 10, Total: 10
	// 22 + 1 + 5 + 1 + 9 + 1 + 9 = 48
	conn.Write([]byte("Mahsulot               Soni   Narxi      Jami\n"))
	for _, item := range order.Items {
		name := s.transliterate(item.ProductName)
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		
		line := fmt.Sprintf("%-22s %-6d %-10.0f %-10.0f\n", 
			name, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		conn.Write([]byte(line))
		
		if item.Comment != "" {
			conn.Write([]byte(fmt.Sprintf("  * Izoh: %s\n", s.transliterate(item.Comment))))
		}
	}
	conn.Write([]byte("------------------------------------------------\n"))

	// Total
	conn.Write(ALIGN_RIGHT)
	conn.Write(FONT_DOUBLE_W)
	conn.Write([]byte(fmt.Sprintf("JAMI: %.0f so'm\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n"))

	// Footer
	conn.Write(ALIGN_CENTER)
	conn.Write([]byte("Xaridingiz uchun rahmat!\n"))
	conn.Write([]byte("\n\n\n\n\n"))

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

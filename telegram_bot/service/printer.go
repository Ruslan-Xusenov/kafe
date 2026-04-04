package service

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"telegram_bot/models"
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
	ESC_INIT      = []byte{0x1B, 0x40}
	ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}
	ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}
	ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}
	FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}
	FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}
	FONT_DOUBLE_W = []byte{0x1D, 0x21, 0x10}
	FONT_BIG      = []byte{0x1D, 0x21, 0x11}
	PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00}
	BEEP          = []byte{0x1B, 0x42, 0x02, 0x02}
)

func (s *PrinterService) PrintOrder(order *models.Order) error {
	if !s.Enabled {
		return fmt.Errorf("printer is disabled")
	}
	if s.IP == "" {
		return fmt.Errorf("printer IP is not set")
	}

	address := net.JoinHostPort(s.IP, s.Port)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("connection error (%s): %v", address, err)
	}
	defer conn.Close()

	if _, err := conn.Write(ESC_INIT); err != nil { return err }
	conn.Write(BEEP)

	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte("KAFE\n"))
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

	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Mijoz: %s\n", order.Phone)))
	conn.Write([]byte(fmt.Sprintf("Manzil: %s\n", s.transliterate(order.Address))))
	if order.Comment != nil && *order.Comment != "" {
		conn.Write([]byte(fmt.Sprintf("Izoh: %s\n", s.transliterate(*order.Comment))))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write([]byte("Mahsulot               Soni   Narxi      Jami\n"))
	for _, item := range order.Items {
		name := ""
		if item.ProductName != nil {
			name = s.transliterate(*item.ProductName)
		}
		if len(name) > 15 {
			name = name[:12] + "..."
		}
		
		line := fmt.Sprintf("%-22s %-6d %-10.0f %-10.0f\n", 
			name, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		conn.Write([]byte(line))
	}
	conn.Write([]byte("------------------------------------------------\n"))

	conn.Write(ALIGN_RIGHT)
	conn.Write(FONT_DOUBLE_W)
	conn.Write([]byte(fmt.Sprintf("JAMI: %.0f so'm\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n"))

	conn.Write(ALIGN_CENTER)
	conn.Write([]byte("Xaridingiz uchun rahmat!\n"))
	conn.Write([]byte("\n\n\n\n\n"))
	_, err = conn.Write(PAPER_CUT)
	return err
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
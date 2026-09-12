package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// Order status and models (same as backend)
type Order struct {
	ID                int        `json:"id"`
	TotalPrice        float64    `json:"total_price"`
	Items             []OrderItem `json:"items"`
	Address           string     `json:"address"`
	Status            string     `json:"status"`
	Phone             string     `json:"phone"`
	Comment           string     `json:"comment"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
	WaiterName        string     `json:"waiter_name"`
	TableName         *string    `json:"table_name"`
	ServiceFee        float64    `json:"service_fee"`
	ServicePercentage float64    `json:"service_percentage"`
}

type OrderItem struct {
	ProductName   string  `json:"product_name"`
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	PrinterTarget string  `json:"printer_target"`
}

type WSMessage struct {
	Type        string             `json:"type"`
	Order       Order              `json:"order,omitempty"`
	ShiftReport ShiftReportPayload `json:"-"` // Handled manually or flat
}

type ShiftSoldItem struct {
	ProductName string  `json:"product_name"`
	Quantity    float64 `json:"quantity"`
	TotalAmount float64 `json:"total_amount"`
}

type ShiftReportPayload struct {
	Timestamp      string          `json:"timestamp"`
	TotalRevenue   float64         `json:"total_revenue"`
	TotalExpenses  float64         `json:"total_expenses"`
	NetProfit      float64         `json:"net_profit"`
	Cash           float64         `json:"cash"`
	Card           float64         `json:"card"`
	Click          float64         `json:"click"`
	Nasiya         float64         `json:"nasiya"`
	AllTimeRevenue float64         `json:"all_time_revenue"`
	ItemsSold      []ShiftSoldItem `json:"items_sold"`
}

func main() {
	// Load environment variables (from local .env or from app folder)
	godotenv.Load()

	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = "kafe.securehub.uz"
	}
	printerKey := os.Getenv("PRINTER_KEY")
	if printerKey == "" {
		printerKey = os.Getenv("PRINTER_SECRET")
	}
	if printerKey == "" {
		printerKey = "KAFE_PRINTER_SECRET_2026"
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	scheme := "ws"
	if os.Getenv("USE_SSL") != "false" {
		scheme = "wss"
	}
	u := url.URL{Scheme: scheme, Host: apiHost, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}

	log.Printf("🔌 Connecting to WebSocket: %s", u.String())

	for {
		var conn *websocket.Conn
		var err error

		conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil && strings.Contains(err.Error(), "getaddrinfo") {
			// DNS failed, fallback to direct IP and skip TLS
			log.Printf("⚠️ DNS lookup failed, falling back to direct IP connection...")
			uFallback := u
			uFallback.Host = "157.180.118.115"
			fallbackDialer := &websocket.Dialer{
				Proxy:            http.ProxyFromEnvironment,
				HandshakeTimeout: 45 * time.Second,
				TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
			}
			conn, _, err = fallbackDialer.Dial(uFallback.String(), nil)
		}

		if err != nil {
			log.Printf("❌ Connection failed: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Bridge Connected to Server!")

		done := make(chan struct{})

		// Message handling loop
		go func() {
			defer close(done)
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					log.Printf("❌ Read error: %v", err)
					return
				}

				var wsMsg WSMessage
				if err := json.Unmarshal(message, &wsMsg); err != nil {
					log.Printf("❌ Message parse error: %v", err)
					continue
				}

				if wsMsg.Type == "new_order" {
					log.Printf("🔔 New Order #%d", wsMsg.Order.ID)
					if wsMsg.Order.ID <= 0 {
						log.Printf("⏭ Skipping empty/dummy order event")
						continue
					}
					// Route each item to its own printer (kitchen vs bar)
					printOrderRouted(wsMsg.Order)
				} else if wsMsg.Type == "close_order" || wsMsg.Type == "reprint_order" {
					log.Printf("🔔 Receipt Event (%s) #%d → Bar/Kassa only", wsMsg.Type, wsMsg.Order.ID)
					if wsMsg.Order.ID <= 0 {
						log.Printf("⏭ Skipping empty/dummy order event")
						continue
					}
					// Full receipt: always ONLY to Bar/Kassa (default printer)
					printOrderToDefault(wsMsg.Order)
				} else if wsMsg.Type == "shift_report" {
					log.Printf("🔔 Shift Report → Bar/Kassa only")
					var shiftReport ShiftReportPayload
					if err := json.Unmarshal(message, &shiftReport); err == nil {
						printShiftReport(shiftReport)
					}
				} else {
					log.Printf("ℹ️ Ignoring event type: %s", wsMsg.Type)
				}
			}
		}()

		// Keep-alive/Exit loop
		ticker := time.NewTicker(30 * time.Second)
		select {
		case <-done:
			log.Println("🔄 Reconnecting...")
		case <-interrupt:
			log.Println("🛑 Shutting down...")
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		case <-ticker.C:
			// Just a heartbeat
		}
		ticker.Stop()
		conn.Close()
		time.Sleep(2 * time.Second)
	}
}

// printOrderRouted routes each item to its designated printer (for new_order events).
// Kitchen items → kitchen printer, Bar/Kassa items → default printer.
func printOrderRouted(order Order) {
	defaultIP := os.Getenv("PRINTER_IP")
	if defaultIP == "" {
		defaultIP = "192.168.123.102"
	}
	printerPort := os.Getenv("PRINTER_PORT")
	if printerPort == "" {
		printerPort = "9100"
	}

	// Group items by PrinterTarget
	itemsByPrinter := make(map[string][]OrderItem)
	for _, item := range order.Items {
		ip := item.PrinterTarget
		// Treat "USB", "ALL", empty string, or any non-IP value as default printer
		if ip == "" || strings.EqualFold(ip, "ALL") || strings.EqualFold(ip, "USB") || net.ParseIP(ip) == nil {
			ip = defaultIP
		}
		itemsByPrinter[ip] = append(itemsByPrinter[ip], item)
	}

	for ip, items := range itemsByPrinter {
		printOrderToPrinter(order, items, ip, printerPort, false)
	}
}

// printOrderToDefault sends the FULL receipt (all items) to Bar/Kassa only.
// Used for close_order, reprint_order — customer-facing receipts.
func printOrderToDefault(order Order) {
	defaultIP := os.Getenv("PRINTER_IP")
	if defaultIP == "" {
		defaultIP = "192.168.123.102"
	}
	printerPort := os.Getenv("PRINTER_PORT")
	if printerPort == "" {
		printerPort = "9100"
	}
	// Always print ALL items to the default (Bar/Kassa) printer only
	printOrderToPrinter(order, order.Items, defaultIP, printerPort, true)
}

func printOrderToPrinter(order Order, items []OrderItem, printerIP, printerPort string, printPrice bool) {
	address := net.JoinHostPort(printerIP, printerPort)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Printf("❌ Printer connection failed (%s): %v", address, err)
		return
	}
	defer conn.Close()

	// ESC/POS Commands
	var (
		ESC_INIT      = []byte{0x1B, 0x40}
		CODE_PAGE_LAT = []byte{0x1B, 0x74, 0x00} // ESC t 0 — PC437 Latin (Standard Europe)
		ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}
		ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}
		ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}
		FONT_BIG      = []byte{0x1D, 0x21, 0x11} // Double Width & Height
		FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01} // Double Height
		FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}
		BOLD_ON       = []byte{0x1B, 0x45, 0x01}
		BOLD_OFF      = []byte{0x1B, 0x45, 0x00}
		PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00}
		BEEP          = []byte{0x1B, 0x42, 0x02, 0x02}
	)

	conn.Write(ESC_INIT)
	conn.Write(CODE_PAGE_LAT) // Set Latin code page
	conn.Write(BEEP)

	// Header: Cafe Name or Table Name (if kitchen ticket)
	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write(BOLD_ON)
	if !printPrice && order.TableName != nil && *order.TableName != "" {
		// New order (kitchen) -> HUGE TABLE NUMBER at top
		conn.Write([]byte("\n" + *order.TableName + "\n\n"))
	} else {
		cafeName := os.Getenv("CAFE_NAME")
		if cafeName == "" {
			cafeName = "Steak House"
		}
		conn.Write([]byte("\n" + cafeName + "\n\n"))
	}
	conn.Write(BOLD_OFF)

	// Meta Info
	conn.Write(FONT_DOUBLE_H)
	conn.Write(ALIGN_LEFT)

	waiter := transliterate(order.WaiterName)
	if waiter == "" {
		waiter = "Noma'lum"
	}
	conn.Write([]byte(fmt.Sprintf("Ofitsiant: %s\n", waiter)))

	if order.TableName != nil && *order.TableName != "" {
		if printPrice {
			conn.Write(BOLD_ON)
			conn.Write([]byte(fmt.Sprintf("Stol: %s\n", *order.TableName)))
			conn.Write(BOLD_OFF)
		}
	}

	uzt := time.FixedZone("UZT", 5*3600)
	conn.Write([]byte(fmt.Sprintf("Ochildi: %s\n", order.CreatedAt.In(uzt).Format("02.01.2006 15:04"))))
	if order.UpdatedAt != nil && !order.UpdatedAt.IsZero() && order.Status == "delivered" {
		conn.Write([]byte(fmt.Sprintf("Yopildi: %s\n", order.UpdatedAt.In(uzt).Format("02.01.2006 15:04"))))
	}

	conn.Write([]byte("================================\n"))
	
	conn.Write(BOLD_ON)
	if printPrice {
		conn.Write([]byte("Mahsulot       Soni Narx  Jami\n"))
	} else {
		conn.Write([]byte("Mahsulot                 Soni\n"))
	}
	conn.Write(BOLD_OFF)
	conn.Write([]byte("--------------------------------\n"))

	var subTotal float64
	for _, item := range items {
		name := transliterate(item.ProductName)
		
		if printPrice {
			if len(name) > 14 {
				conn.Write([]byte(name + "\n"))
				name = ""
			}
			itemTotal := item.Price * float64(item.Quantity)
			subTotal += itemTotal
			line := fmt.Sprintf("%-14s %-4g %-5.0f %-6.0f\n", name, item.Quantity, item.Price, itemTotal)
			conn.Write([]byte(line))
		} else {
			if len(name) > 24 {
				conn.Write([]byte(name + "\n"))
				name = ""
			}
			conn.Write(BOLD_ON)
			line := fmt.Sprintf("%-24s %-6g\n", name, item.Quantity)
			conn.Write([]byte(line))
			conn.Write(BOLD_OFF)
		}
	}
	conn.Write([]byte("================================\n"))

	if printPrice {
		conn.Write(ALIGN_RIGHT)
		conn.Write(BOLD_ON)

		if order.ServiceFee > 0 && order.ServicePercentage > 0 {
			conn.Write([]byte(fmt.Sprintf("Oraliq: %.0f sum\n", subTotal)))
			conn.Write([]byte(fmt.Sprintf("Xizmat (%.0f%%): %.0f sum\n", order.ServicePercentage, order.ServiceFee)))
			conn.Write(FONT_BIG)
			conn.Write([]byte(fmt.Sprintf("JAMI: %.0f sum\n", subTotal+order.ServiceFee)))
		} else {
			conn.Write(FONT_BIG)
			conn.Write([]byte(fmt.Sprintf("JAMI: %.0f sum\n", subTotal)))
		}
		conn.Write(BOLD_OFF)
	}

	// Footer
	if printPrice {
		conn.Write(ALIGN_CENTER)
		conn.Write(FONT_DOUBLE_H)
		conn.Write([]byte("\nXaridingiz uchun rahmat!\n"))
	}
	
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n\n\n\n"))
	conn.Write(PAPER_CUT)

	log.Printf("🖨 Order #%d printed successfully to %s!", order.ID, address)
}

func transliterate(text string) string {
	r := strings.NewReplacer(
		// Uzbek Cyrillic → Latin
		"ў", "o'", "Ў", "O'",
		"қ", "q", "Қ", "Q",
		"ғ", "g'", "Ғ", "G'",
		"ҳ", "h", "Ҳ", "H",
		"ъ", "'", "ь", "'",
		// Cyrillic vowels → Latin
		"а", "a", "А", "A",
		"б", "b", "Б", "B",
		"в", "v", "В", "V",
		"г", "g", "Г", "G",
		"д", "d", "Д", "D",
		"е", "e", "Е", "E",
		"ё", "yo", "Ё", "Yo",
		"ж", "j", "Ж", "J",
		"з", "z", "З", "Z",
		"и", "i", "И", "I",
		"й", "y", "Й", "Y",
		"к", "k", "К", "K",
		"л", "l", "Л", "L",
		"м", "m", "М", "M",
		"н", "n", "Н", "N",
		"о", "o", "О", "O",
		"п", "p", "П", "P",
		"р", "r", "Р", "R",
		"с", "s", "С", "S",
		"т", "t", "Т", "T",
		"у", "u", "У", "U",
		"ф", "f", "Ф", "F",
		"х", "x", "Х", "X",
		"ц", "ts", "Ц", "Ts",
		"ч", "ch", "Ч", "Ch",
		"ш", "sh", "Ш", "Sh",
		"щ", "sh", "Щ", "Sh",
		"э", "e", "Э", "E",
		"ю", "yu", "Ю", "Yu",
		"я", "ya", "Я", "Ya",
		"ы", "i", "Ы", "I",
		// Unicode apostrophes → ASCII
		"\u02BC", "'", "\u2019", "'", "\u2018", "'",
		"\u02B9", "'", "\u0060", "'",
		// Uzbek Latin special chars
		"O\u02BC", "O'", "o\u02BC", "o'",
		"G\u02BC", "G'", "g\u02BC", "g'",
	)
	return r.Replace(text)
}

func formatLine(title string, amount float64) string {
	return fmt.Sprintf("%-18s %12s", title, fmt.Sprintf("%.0f", amount))
}

func printShiftReport(report ShiftReportPayload) {
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" {
		printerIP = "192.168.123.102"
	}
	printerPort := os.Getenv("PRINTER_PORT")
	if printerPort == "" {
		printerPort = "9100"
	}

	address := net.JoinHostPort(printerIP, printerPort)
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		log.Printf("❌ Printer connection failed (%s): %v", address, err)
		return
	}
	defer conn.Close()

	// ESC/POS Commands
	var (
		ESC_INIT     = []byte{0x1B, 0x40}
		ALIGN_CENTER = []byte{0x1B, 0x61, 0x01}
		ALIGN_LEFT   = []byte{0x1B, 0x61, 0x00}
		FONT_BIG     = []byte{0x1D, 0x21, 0x11}
		FONT_NORMAL  = []byte{0x1D, 0x21, 0x00}
		PAPER_CUT    = []byte{0x1D, 0x56, 0x42, 0x00}
	)

	// Initialize printer
	conn.Write(ESC_INIT)
	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)

	cafeName := os.Getenv("CAFE_NAME")
	if cafeName == "" {
		cafeName = "Steak House"
	}
	conn.Write([]byte(cafeName + "\n"))

	conn.Write(FONT_NORMAL)
	conn.Write([]byte("--------------------------------\n"))
	conn.Write(FONT_BIG)
	conn.Write([]byte("Z-REPORT (SMENA YOPILDI)\n"))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("--------------------------------\n"))

	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Sana/Vaqt: %s\n", report.Timestamp)))
	conn.Write([]byte("--------------------------------\n"))

	conn.Write([]byte(formatLine("Umumiy savdo:", report.TotalRevenue) + "\n"))
	conn.Write([]byte(formatLine("Xarajatlar:", report.TotalExpenses) + "\n"))
	conn.Write([]byte(formatLine("Sof foyda:", report.NetProfit) + "\n"))

	conn.Write([]byte("--------------------------------\n"))
	conn.Write([]byte("Tolov turlari bo'yicha:\n"))
	if report.Cash > 0 {
		conn.Write([]byte(formatLine("Naqd:", report.Cash) + "\n"))
	}
	if report.Card > 0 {
		conn.Write([]byte(formatLine("Karta/Terminal:", report.Card) + "\n"))
	}
	if report.Click > 0 {
		conn.Write([]byte(formatLine("Click/Payme:", report.Click) + "\n"))
	}
	if report.Nasiya > 0 {
		conn.Write([]byte(formatLine("Nasiya (Qarz):", report.Nasiya) + "\n"))
	}

	conn.Write([]byte("--------------------------------\n"))
	conn.Write(ALIGN_CENTER)
	conn.Write([]byte("SOTILGAN TOVARLAR\n"))
	conn.Write(ALIGN_LEFT)
	conn.Write([]byte("--------------------------------\n"))

	for _, item := range report.ItemsSold {
		name := transliterate(item.ProductName)
		if len(name) > 16 {
			name = name[:16]
		}
		qtyStr := fmt.Sprintf("%.1f", item.Quantity)
		sumStr := fmt.Sprintf("%.0f", item.TotalAmount)
		line := fmt.Sprintf("%-16s %4s x %7s\n", name, qtyStr, sumStr)
		conn.Write([]byte(line))
	}

	conn.Write([]byte("--------------------------------\n"))
	conn.Write(FONT_BIG)
	conn.Write([]byte(fmt.Sprintf("JAMI TARIXIY SAVDO:\n%.0f so'm\n", report.AllTimeRevenue)))
	conn.Write(FONT_NORMAL)

	conn.Write([]byte("\n\n\n\n\n"))
	conn.Write(PAPER_CUT)

	log.Printf("🖨️ Shift Report printed successfully!")
}
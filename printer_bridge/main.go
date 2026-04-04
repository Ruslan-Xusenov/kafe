package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
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
	ID         int         `json:"id"`
	TotalPrice float64     `json:"total_price"`
	Items      []OrderItem `json:"items"`
	Address    string      `json:"address"`
	Phone      string      `json:"phone"`
	Comment    string      `json:"comment"`
	CreatedAt  time.Time   `json:"created_at"`
}

type OrderItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type WSMessage struct {
	Type  string `json:"type"`
	Order Order  `json:"order"`
}

func main() {
	// Load environment variables (from local .env or from app folder)
	godotenv.Load()

	apiHost := os.Getenv("API_HOST") // e.g. "api.kafe.uz"
	if apiHost == "" {
		apiHost = "localhost:8080"
	}
	printerKey := os.Getenv("PRINTER_KEY")
	if printerKey == "" {
		printerKey = "KAFE_PRINTER_SECRET_2026"
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: apiHost, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}
	
	log.Printf("🔌 Connecting to WebSocket: %s", u.String())

	for {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
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
					log.Printf("🔔 Received New Order #%d", wsMsg.Order.ID)
					printOrder(wsMsg.Order)
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

func printOrder(order Order) {
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" {
		log.Println("⚠️ PRINTER_IP not set in .env. Skipping print.")
		return
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
		ESC_INIT      = []byte{0x1B, 0x40}
		ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}
		ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}
		ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}
		FONT_BIG      = []byte{0x1D, 0x21, 0x11}
		FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}
		FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}
		PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00}
		BEEP          = []byte{0x1B, 0x42, 0x02, 0x02}
	)

	conn.Write(ESC_INIT)
	conn.Write(BEEP)

	conn.Write(ALIGN_CENTER)
	conn.Write(FONT_BIG)
	conn.Write([]byte("KAFE\n"))
	
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("--------------------------------\n"))
	conn.Write(FONT_DOUBLE_H)
	conn.Write([]byte(fmt.Sprintf("BUYURTMA #%d\n", order.ID)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte(fmt.Sprintf("Sana: %s\n", time.Now().Format("02.01.2006 15:04"))))
	conn.Write([]byte("--------------------------------\n"))

	conn.Write(ALIGN_LEFT)
	conn.Write([]byte(fmt.Sprintf("Mijoz: %s\n", order.Phone)))
	conn.Write([]byte(fmt.Sprintf("Manzil: %s\n", transliterate(order.Address))))
	if order.Comment != "" {
		conn.Write([]byte(fmt.Sprintf("Izoh: %s\n", transliterate(order.Comment))))
	}
	conn.Write([]byte("--------------------------------\n"))

	conn.Write([]byte("Mahsulot        Soni  Narxi   Jami\n"))
	for _, item := range order.Items {
		name := transliterate(item.ProductName)
		if len(name) > 15 {
			name = name[:12] + "..."
		}
		line := fmt.Sprintf("%-15s %-4d %-6.0f %-6.0f\n", 
			name, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		conn.Write([]byte(line))
	}
	conn.Write([]byte("--------------------------------\n"))

	conn.Write(ALIGN_RIGHT)
	conn.Write(FONT_DOUBLE_H)
	conn.Write([]byte(fmt.Sprintf("JAMI: %.0f so'm\n", order.TotalPrice)))
	conn.Write(FONT_NORMAL)
	conn.Write([]byte("\n\n\n"))
	
	conn.Write(PAPER_CUT)
	
	log.Printf("🖨 Order #%d printed successfully to %s!", order.ID, address)
}

func transliterate(text string) string {
	r := strings.NewReplacer(
		"ў", "o'", "Ў", "O'",
		"қ", "q", "Қ", "Q",
		"ғ", "g'", "Ғ", "G'",
		"ҳ", "h", "Ҳ", "H",
		"‘", "'", "’", "'",
	)
	return r.Replace(text)
}

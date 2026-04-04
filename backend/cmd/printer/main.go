package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

// ESC/POS Commands
var (
	ESC_INIT      = []byte{0x1B, 0x40}
	PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00}
	BEEP          = []byte{0x1B, 0x42, 0x02, 0x02}
	ALIGN_CTR     = []byte{0x1B, 0x61, 0x01}
	ALIGN_LFT     = []byte{0x1B, 0x61, 0x00}
	CHAR_CYR      = []byte{0x1B, 0x74, 0x11} // CP866 (Russian)
	TXT_NORMAL    = []byte{0x1D, 0x21, 0x00}
	TXT_DOUBLE_H  = []byte{0x1D, 0x21, 0x01}
	TXT_DOUBLE_W  = []byte{0x1D, 0x21, 0x10}
	TXT_BIG       = []byte{0x1D, 0x21, 0x11}
)

type OrderPrint struct {
	ID         int      `json:"id"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
	Comment    string   `json:"comment"`
	Items      []Item   `json:"items"`
	TotalPrice float64  `json:"total_price"`
}

type Item struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

func main() {
	_ = godotenv.Load()

	serverHost := os.Getenv("API_HOST")
	if serverHost == "" { serverHost = "46.224.133.140:8080" }
	
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" { printerIP = "\\\\localhost\\shashlik" }

	u := url.URL{
		Scheme: "ws", 
		Host: serverHost, 
		Path: "/api/ws",
		RawQuery: "printer_key=KAFE_PRINTER_SECRET_2026",
	}
	
	for {
		log.Printf("Connecting to Server: %s...", u.String())
		dialer := websocket.Dialer{ HandshakeTimeout: 10 * time.Second }
		c, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Failed: %v. Retrying...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Connected! Waiting for orders...")
		err = handleMessages(c, printerIP)
		if err != nil {
			log.Printf("⚠️  Lost connection: %v", err)
			c.Close()
			time.Sleep(2 * time.Second)
		}
	}
}

func handleMessages(c *websocket.Conn, printerIP string) error {
	for {
		_, message, err := c.ReadMessage()
		if err != nil { return err }

		var event struct {
			Type  string     `json:"type"`
			Order OrderPrint `json:"order"`
		}

		if err := json.Unmarshal(message, &event); err == nil && event.Type == "new_order" {
			log.Printf("📦 New order #%d received", event.Order.ID)
			go sendToPrinter(event.Order, printerIP)
		}
	}
}

func sendToPrinter(o OrderPrint, ip string) {
	var p []byte // Payload
	p = append(p, ESC_INIT...)
	p = append(p, CHAR_CYR...)
	p = append(p, BEEP...)

	// Header - ПРЕДВАРИТЕЛЬНЫЙ СЧЁТ
	p = append(p, ALIGN_CTR...)
	p = append(p, []byte("--------------------------------\n")...)
	p = append(p, TXT_DOUBLE_H...)
	p = append(p, []byte("ПРЕДВАРИТЕЛЬНЫЙ СЧЁТ\n")...)
	p = append(p, TXT_NORMAL...)
	p = append(p, []byte("--------------------------------\n")...)

	// Info section
	p = append(p, ALIGN_LFT...)
	p = append(p, []byte(fmt.Sprintf("Чек №: %d\n", o.ID))...)
	p = append(p, []byte("Зал: Доставка\n")...) // Default as the system focuses on delivery
	p = append(p, []byte(fmt.Sprintf("Тел: %s\n", o.Phone))...)
	p = append(p, []byte("Обслужил: Administrator\n")...)
	p = append(p, []byte(fmt.Sprintf("Время открытия: %s\n", time.Now().Format("02.01.2006 15:04:05")))...)
	p = append(p, []byte("Время закрытия: -\n")...)
	p = append(p, []byte("\n")...)

	// Table Header
	p = append(p, []byte("Наименование |Кол-во| Цена | Сумма\n")...)
	p = append(p, []byte("--------------------------------\n")...)
	
	// Items
	for _, item := range o.Items {
		name := strings.ToLower(item.ProductName)
		if len(name) > 13 { name = name[:10] + "..." }
		
		line := fmt.Sprintf("%-13s| %-4d |%-6.0f| %-6.0f\n", 
			name, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		p = append(p, []byte(line)...)
	}
	
	p = append(p, []byte("--------------------------------\n")...)

	// Note / Comment
	if o.Comment != "" {
		p = append(p, []byte("Примечание:\n")...)
		p = append(p, TXT_DOUBLE_W...)
		p = append(p, []byte(fmt.Sprintf("%s\n", strings.ToUpper(o.Comment)))...)
		p = append(p, TXT_NORMAL...)
	}

	// Totals
	p = append(p, []byte(fmt.Sprintf("\nПодитог: %.0f\n", o.TotalPrice))...)
	p = append(p, []byte("Обслуживание(0.0%): 0\n")...)
	p = append(p, []byte("Скидка(0%): 0\n")...)
	p = append(p, []byte("\n")...)

	// Final Total
	p = append(p, TXT_BIG...)
	p = append(p, []byte(fmt.Sprintf("Итого: %.0f\n", o.TotalPrice))...)
	p = append(p, TXT_NORMAL...)
	
	p = append(p, []byte("\n\n\n\n\n")...)
	p = append(p, PAPER_CUT...)

	// Execute Print
	if strings.HasPrefix(ip, "\\\\") {
		_ = os.WriteFile(ip, p, 0666)
		log.Printf("✅ Order #%d printed (Style: Preliminary Bill)", o.ID)
	} else {
		conn, err := net.DialTimeout("tcp", ip+":9100", 5*time.Second)
		if err == nil {
			_, _ = conn.Write(p)
			conn.Close()
			log.Printf("✅ Order #%d printed", o.ID)
		}
	}
}
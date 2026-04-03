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
	ESC_INIT   = []byte{0x1B, 0x40}
	PAPER_CUT  = []byte{0x1D, 0x56, 0x42, 0x00}
	BEEP       = []byte{0x1B, 0x42, 0x02, 0x02}
	ALIGN_CTR  = []byte{0x1B, 0x61, 0x01}
	ALIGN_LFT  = []byte{0x1B, 0x61, 0x00}
	CHAR_CYR   = []byte{0x1B, 0x74, 0x11} // CP866 (Russian)
)

type OrderPrint struct {
	ID         int      `json:"id"`
	Phone      string   `json:"phone"`
	Address    string   `json:"address"`
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
	p = append(p, CHAR_CYR...) // Kirill alifbosi rejimini yoqish
	p = append(p, BEEP...)
	p = append(p, ALIGN_CTR...)
	p = append(p, []byte(fmt.Sprintf("\n*** yangi buyurtma #%d ***\n\n", o.ID))...)
	p = append(p, ALIGN_LFT...)
	p = append(p, []byte(strings.ToLower(fmt.Sprintf("mijoz:  %s\n", o.Phone)))...)
	p = append(p, []byte(strings.ToLower(fmt.Sprintf("vaqt:   %s\n", time.Now().Format("15:04 02.01.2006"))))...)
	p = append(p, []byte(strings.ToLower(fmt.Sprintf("manzil: %s\n", o.Address)))...)
	p = append(p, []byte("--------------------------------\n")...)
	
	for _, item := range o.Items {
		name := strings.ToLower(item.ProductName) // Hammasi kichik harfda!
		if len(name) > 20 { name = name[:17] + "..." }
		p = append(p, []byte(fmt.Sprintf("%-20s x%-4d\n", name, item.Quantity))...)
	}
	
	p = append(p, []byte("--------------------------------\n")...)
	p = append(p, []byte(strings.ToLower(fmt.Sprintf("jami:   %.0f so'm\n\n", o.TotalPrice)))...)
	p = append(p, []byte("\n\n\n\n\n")...)
	p = append(p, PAPER_CUT...)

	// Windows USB yoki LAN orqali chiqarish
	if strings.HasPrefix(ip, "\\\\") {
		_ = os.WriteFile(ip, p, 0666)
		log.Printf("✅ Order #%d printed (USB)", o.ID)
	} else {
		conn, err := net.DialTimeout("tcp", ip+":9100", 5*time.Second)
		if err == nil {
			_, _ = conn.Write(p)
			conn.Close()
			log.Printf("✅ Order #%d printed (LAN)", o.ID)
		}
	}
}
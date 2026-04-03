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
	ESC_INIT  = []byte{0x1B, 0x40}
	PAPER_CUT = []byte{0x1D, 0x56, 0x42, 0x00}
	BEEP      = []byte{0x1B, 0x42, 0x02, 0x02}
	ALIGN_CTR = []byte{0x1B, 0x61, 0x01}
	ALIGN_LFT = []byte{0x1B, 0x61, 0x00}
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
	// 1. .env faylini yuklash
	_ = godotenv.Load()

	// Configuration
	serverHost := os.Getenv("API_HOST")
	if serverHost == "" {
		serverHost = "46.224.133.140:8080"
	}
	
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" {
		printerIP = "192.168.123.10"
	}

	printerPort := os.Getenv("PRINTER_PORT")
	if printerPort == "" {
		printerPort = "9100"
	}

	u := url.URL{
		Scheme: "ws", 
		Host: serverHost, 
		Path: "/api/ws",
		RawQuery: "printer_key=KAFE_PRINTER_SECRET_2026",
	}
	
	for {
		log.Printf("Connecting to Server: %s...", u.String())
		log.Printf("Current Target Printer: %s (%s)", printerIP, printerPort)
		
		dialer := websocket.Dialer{ HandshakeTimeout: 10 * time.Second }
		c, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Connection failed: %v. Retrying...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Connected to Server! Waiting for orders...")
		err = handleMessages(c, printerIP, printerPort)
		if err != nil {
			log.Printf("⚠️  Connection lost: %v. Reconnecting...", err)
			c.Close()
			time.Sleep(2 * time.Second)
		}
	}
}

func handleMessages(c *websocket.Conn, printerIP, printerPort string) error {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		var event struct {
			Type  string     `json:"type"`
			Order OrderPrint `json:"order"`
		}

		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("JSON Error: %v", err)
			continue
		}

		if event.Type == "new_order" {
			log.Printf("📦 New order received: #%d", event.Order.ID)
			go sendToPrinter(event.Order, printerIP, printerPort)
		}
	}
}

func sendToPrinter(o OrderPrint, ip, port string) {
	var payload []byte
	payload = append(payload, ESC_INIT...)
	payload = append(payload, BEEP...)
	payload = append(payload, ALIGN_CTR...)
	payload = append(payload, []byte(fmt.Sprintf("\n*** YANGI BUYURTMA #%d ***\n\n", o.ID))...)
	payload = append(payload, ALIGN_LFT...)
	payload = append(payload, []byte(fmt.Sprintf("Mijoz:  %s\n", o.Phone))...)
	payload = append(payload, []byte(fmt.Sprintf("Vaqt:   %s\n", time.Now().Format("02.01.2006 15:04")))...)
	payload = append(payload, []byte(fmt.Sprintf("Manzil: %s\n", o.Address))...)
	payload = append(payload, []byte("--------------------------------\n")...)
	
	for _, item := range o.Items {
		name := item.ProductName
		if len(name) > 20 { name = name[:17] + "..." }
		payload = append(payload, []byte(fmt.Sprintf("%-20s x%-4d\n", name, item.Quantity))...)
	}
	
	payload = append(payload, []byte("--------------------------------\n")...)
	payload = append(payload, []byte(fmt.Sprintf("JAMI:   %.0f so'm\n\n", o.TotalPrice))...)
	payload = append(payload, []byte("\n\n\n\n\n")...)
	payload = append(payload, PAPER_CUT...)

	if strings.HasPrefix(ip, "\\\\") {
		// WINDOWS USB
		log.Printf("🔌 Printing via USB Shared: %s", ip)
		err := os.WriteFile(ip, payload, 0666)
		if err != nil {
			log.Printf("🚨 USB Error: %v", err)
		} else {
			log.Printf("✅ Order #%d printed via USB!", o.ID)
		}
	} else {
		// NETWORK
		address := net.JoinHostPort(ip, port)
		log.Printf("🌐 Printing via LAN: %s", address)
		conn, err := net.DialTimeout("tcp", address, 5*time.Second)
		if err != nil {
			log.Printf("🚨 LAN Error: %v", err)
			return
		}
		defer conn.Close()
		_, _ = conn.Write(payload)
		log.Printf("✅ Order #%d printed via LAN!", o.ID)
	}
}
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/gorilla/websocket"
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
	// Configuration (Can be overridden by ENV)
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
	
	// Infinite connection loop (Auto-Reconnect)
	for {
		log.Printf("Connecting to Server: %s...", u.String())
		
		// Use a dialer with timeout
		dialer := websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}
		
		c, _, err := dialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Failed to connect to server: %v. Retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Connected to Server! Waiting for orders...")

		// Handle incoming messages
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
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if event.Type == "new_order" {
			log.Printf("📦 New order received: #%d. Printing to %s...", event.Order.ID, printerIP)
			go sendToPrinter(event.Order, printerIP, printerPort)
		}
	}
}

func sendToPrinter(o OrderPrint, ip, port string) {
	address := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		log.Printf("🚨 Printer Connection Error (%s): %v", address, err)
		return
	}
	defer conn.Close()

	// Format Receipt (Simplified ESC/POS)
	conn.Write(ESC_INIT)
	conn.Write(BEEP)
	conn.Write(ALIGN_CTR)
	conn.Write([]byte(fmt.Sprintf("\n*** YANGI BUYURTMA #%d ***\n", o.ID)))
	conn.Write(ALIGN_LFT)
	conn.Write([]byte(fmt.Sprintf("Mijoz:   %s\n", o.Phone)))
	conn.Write([]byte(fmt.Sprintf("Vaqt:    %s\n", time.Now().Format("02.01.2006 15:04"))))
	conn.Write([]byte("--------------------------------\n"))
	
	for _, item := range o.Items {
		name := item.ProductName
		if len(name) > 18 { name = name[:15] + "..." }
		conn.Write([]byte(fmt.Sprintf("%-18s x%-4d\n", name, item.Quantity)))
	}
	
	conn.Write([]byte("--------------------------------\n"))
	conn.Write([]byte(fmt.Sprintf("JAMI:    %.0f so'm\n\n", o.TotalPrice)))
	conn.Write([]byte("\n\n\n"))
	conn.Write(PAPER_CUT)
	
	log.Printf("✅ Order #%d successfully sent to printer!", o.ID)
}
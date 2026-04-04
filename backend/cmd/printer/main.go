package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type OrderItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Order struct {
	ID         int         `json:"id"`
	TotalPrice float64     `json:"total_price"`
	Address    string      `json:"address"`
	Phone      string      `json:"phone"`
	Comment    string      `json:"comment"`
	Items      []OrderItem `json:"items"`
}

type WSEvent struct {
	Type  string `json:"type"`
	Order Order  `json:"order"`
}

func main() {
	serverHost := os.Getenv("API_HOST")
	if serverHost == "" { serverHost = "46.224.133.140:8080" }
	
	printerIP := os.Getenv("PRINTER_IP")
	if printerIP == "" { printerIP = "\\\\localhost\\XP-80C" }

	for {
		log.Printf("Connecting to Server: ws://%s/api/ws?printer_key=KAFE_PRINTER_SECRET_2026...", serverHost)
		
		u := url.URL{
			Scheme: "ws", 
			Host: serverHost, 
			Path: "/api/ws",
			RawQuery: "printer_key=KAFE_PRINTER_SECRET_2026",
		}

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Connection failed: %v. Retrying in 5s...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("✅ Connected! Waiting for orders...")
		
		err = handleMessages(c, printerIP)
		if err != nil {
			log.Printf("📉 Connection lost: %v. Reconnecting in 5s...", err)
			c.Close()
			time.Sleep(5 * time.Second)
			continue
		}
	}
}

func handleMessages(c *websocket.Conn, printerIP string) error {
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			return err
		}

		var event WSEvent
		if err := json.Unmarshal(message, &event); err == nil && event.Type == "new_order" {
			log.Printf("📦 New order #%d received", event.Order.ID)
			printReceipt(event.Order, printerIP)
		}
	}
}

func printReceipt(order Order, printerIP string) {
	var b strings.Builder

	// ESC/POS Commands
	b.Write([]byte{0x1B, 0x40}) // Initialize
	b.Write([]byte{0x1B, 0x61, 0x01}) // Center
	b.Write([]byte{0x1D, 0x21, 0x11}) // Double size
	b.WriteString("ONLAYN BUYURTMA\n")
	b.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
	b.WriteString(fmt.Sprintf("Buyurtma #%d\n", order.ID))
	b.WriteString(time.Now().Format("02.01.2006 15:04:35") + "\n")
	b.WriteString("--------------------------------\n")
	b.Write([]byte{0x1B, 0x61, 0x00}) // Left align

	for _, item := range order.Items {
		name := item.ProductName
		if len(name) > 20 { name = name[:17] + "..." }
		b.WriteString(fmt.Sprintf("%-20s %2dx %6.0f\n", name, item.Quantity, item.Price))
	}

	b.WriteString("--------------------------------\n")
	b.Write([]byte{0x1B, 0x61, 0x02}) // Right align
	b.WriteString(fmt.Sprintf("JAMI: %.0f so'm\n", order.TotalPrice))
	b.Write([]byte{0x1B, 0x61, 0x00}) // Left align
	b.WriteString("--------------------------------\n")
	b.WriteString(fmt.Sprintf("Manzil: %s\n", order.Address))
	b.WriteString(fmt.Sprintf("Tel: %s\n", order.Phone))
	if order.Comment != "" {
		b.WriteString(fmt.Sprintf("Izoh: %s\n", order.Comment))
	}
	b.WriteString("\n\n\n\n\n")
	b.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Cut

	err := os.WriteFile(printerIP, []byte(b.String()), 0666)
	if err != nil {
		log.Printf("❌ Print Error: %v", err)
	} else {
		log.Printf("✅ Order #%d printed (Style: Onlayn Buyurtma)", order.ID)
	}
}
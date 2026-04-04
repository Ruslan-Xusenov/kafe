package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	serverAddr = "46.224.133.140:8080"
	printerKey = "KAFE_PRINTER_SECRET_2026"
	printerDevice = "\\\\localhost\\XP-80C"
)

func main() {
	log.Println("🚀 Kafe Printer Bridge Final Pro (v3.5) ishga tushdi...")
	log.Printf("📍 Server: %s\n", serverAddr)
	log.Printf("📍 Printer: %s\n\n", printerDevice)

	for {
		// 1. Health Check (ws-test)
		testURL := fmt.Sprintf("http://%s/api/ws-test", serverAddr)
		resp, err := http.Get(testURL)
		if err != nil {
			log.Printf("❌ Server topilmadi: %v. Qayta urinish (10s)...\n", err)
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 404 {
			log.Println("⚠️  OGOHLANTIRISH: Serverda eski kod turibdi (404 on ws-test)! Iltimos 'git pull' qiling.")
		} else {
			log.Println("✅ Server ulanishga tayyor (200 OK).")
		}

		// 2. WebSocket Handshake
		u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Ulanishda xatolik: %v. Qayta urinish (5s)...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		log.Println("✅ Connected! Buyurtmalar kutilmoqda...")
		
		handleMessages(c)
		c.Close()
		log.Println("⚠️  Ulanish uzildi. Qayta tiklanmoqda...")
		time.Sleep(2 * time.Second)
	}
}

func handleMessages(c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("❌ Read error:", err)
			return
		}

		var m map[string]interface{}
		if err := json.Unmarshal(message, &m); err != nil {
			continue
		}

		if m["type"] == "new_order" {
			orderData, _ := m["order"].(map[string]interface{})
			id := int(orderData["id"].(float64))
			log.Printf("🔔 Yangi buyurtma #%d qabul qilindi!\n", id)
			
			// Auto-print
			printOrder(orderData)
		}
	}
}

func printOrder(order map[string]interface{}) {
	id := int(order["id"].(float64))
	
	f, err := os.CreateTemp("", fmt.Sprintf("order_%d_*.bin", id))
	if err != nil {
		log.Println("❌ Fayl yaratishda xato:", err)
		return
	}
	defer os.Remove(f.Name())

	// --- ESC/POS Commands ---
	const (
		initialize     = "\x1b\x40"
		center         = "\x1b\x61\x01"
		left           = "\x1b\x61\x00"
		boldOn         = "\x1b\x45\x01"
		boldOff        = "\x1b\x45\x00"
		doubleSize     = "\x1d\x21\x11" // Double height and width
		regularSize    = "\x1d\x21\x00"
		cut            = "\x1d\x56\x42\x00" // Feed paper and cut
	)

	// Build the Binary Raw Data
	var b strings.Builder
	b.WriteString(initialize)
	
	// Header: Cafe Name
	b.WriteString(center + doubleSize + boldOn + "YETUK KAFE\n" + regularSize + boldOff)
	b.WriteString("--------------------------------\n")
	
	// Order ID
	b.WriteString(doubleSize + boldOn + fmt.Sprintf("BUYURTMA #%d\n", id) + regularSize + boldOff)
	b.WriteString(center + fmt.Sprintf("Vaqt: %s\n", time.Now().Format("15:04:05 02.01.2006")))
	b.WriteString(left + "--------------------------------\n")
	
	// Items Table
	b.WriteString(boldOn + fmt.Sprintf("%-18s %-4s %-8s\n", "Mahsulot", "Soni", "Narxi") + boldOff)
	items, _ := order["items"].([]interface{})
	for _, it := range items {
		item := it.(map[string]interface{})
		name := item["product_name"].(string)
		qty := int(item["quantity"].(float64))
		price := item["price"].(float64)
		
		// If name too long, truncate
		shortName := name
		if len(shortName) > 17 { shortName = shortName[:17] }
		
		b.WriteString(fmt.Sprintf("%-18s x%-3d %-8.0f\n", shortName, qty, price))
	}
	
	b.WriteString("--------------------------------\n")
	b.WriteString(doubleSize + boldOn + fmt.Sprintf("JAMI: %.0f so'm\n", order["total_price"].(float64)) + regularSize + boldOff)
	b.WriteString("--------------------------------\n")
	
	// Customer Details
	if addr, ok := order["address"].(string); ok && addr != "" {
		b.WriteString(boldOn + "Manzil: " + boldOff + addr + "\n")
	}
	if phone, ok := order["phone"].(string); ok && phone != "" {
		b.WriteString(boldOn + "Tel: " + boldOff + phone + "\n")
	}
	if comment, ok := order["comment"].(string); ok && comment != "" {
		b.WriteString(boldOn + "Izoh: " + boldOff + comment + "\n")
	}
	
	b.WriteString("\n" + center + "Xaridingiz uchun rahmat!\n")
	b.WriteString("\n\n\n\n\n") // Feed paper
	b.WriteString(cut) // Auto-cut!

	f.WriteString(b.String())
	f.Close()

	// Windows print command (Direct binary copy)
	cmd := exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice)
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Chiqarishda xatolik (#%d): %v\n", id, err)
	} else {
		log.Printf("✅ Professional chek #%d qirqildi.\n", id)
	}
}
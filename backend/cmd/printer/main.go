package main

import (
	"encoding/json"
	"fmt"
	"io"
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
			time.Sleep(10 * time.Sleep)
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
	
	f, err := os.CreateTemp("", fmt.Sprintf("order_%d_*.txt", id))
	if err != nil {
		log.Println("❌ Fayl yaratishda xato:", err)
		return
	}
	defer os.Remove(f.Name())

	// Format Order Text
	text := "--- KAFE BUYURTMA ---\n"
	text += fmt.Sprintf("ID: #%d\n", id)
	text += fmt.Sprintf("Vaqt: %s\n", time.Now().Format("15:04:05"))
	text += "----------------------\n"
	
	items, _ := order["items"].([]interface{})
	for _, it := range items {
		item := it.(map[string]interface{})
		name := item["product_name"].(string)
		qty := int(item["quantity"].(float64))
		price := item["price"].(float64)
		text += fmt.Sprintf("%-15s x%d  %.0f\n", name, qty, price)
	}
	
	text += "----------------------\n"
	text += fmt.Sprintf("JAMI: %.0f so'm\n", order["total_price"].(float64))
	text += "\n\n\n\n\n"

	f.WriteString(text)
	f.Close()

	// Windows print command
	cmd := exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice)
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Chiqarishda xatolik (#%d): %v\n", id, err)
	} else {
		log.Printf("✅ Buyurtma #%d printerdan chiqarildi.\n", id)
	}
}
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

var networkPrinters = []string{
	"192.168.1.10:9100",
	"192.168.1.11:9100",
}

const (
	serverAddr    = "kafe.securehub.uz"
	printerDevice = "\\\\localhost\\XP-80C"
)

var printerKey = os.Getenv("PRINTER_SECRET")

func init() {
	if printerKey == "" {
		printerKey = "YetukKafe2026"
	}
}

// ESC/POS Commands
var (
	ESC_INIT      = []byte{0x1B, 0x40}           // ESC @
	CODE_PAGE     = []byte{0x1B, 0x74, 0x11}      // ESC t 17 (PC866)
	ALIGN_LEFT    = []byte{0x1B, 0x61, 0x00}      // ESC a 0
	ALIGN_CENTER  = []byte{0x1B, 0x61, 0x01}      // ESC a 1
	ALIGN_RIGHT   = []byte{0x1B, 0x61, 0x02}      // ESC a 2
	FONT_NORMAL   = []byte{0x1D, 0x21, 0x00}      // GS ! 0
	FONT_DOUBLE_H = []byte{0x1D, 0x21, 0x01}      // GS ! 1
	FONT_DOUBLE_W = []byte{0x1D, 0x21, 0x10}      // GS ! 16
	FONT_BIG      = []byte{0x1D, 0x21, 0x11}      // GS ! 17 (Double width + height)
	PAPER_CUT     = []byte{0x1D, 0x56, 0x42, 0x00} // GS V B 0
	BEEP          = []byte{0x1B, 0x42, 0x02, 0x02} // ESC B n t (Beep 2 times)
)

func main() {
	log.Println("🚀 Kafe Printer Bridge Master v8.5 (Ultra Latin) ishga tushdi...")
	log.Printf("📍 Server: %s\n", serverAddr)
	log.Printf("📍 Printer: %s\n\n", printerDevice)

	for {
		// 1. Health Check (using https:// because server uses HTTPS)
		testURL := fmt.Sprintf("https://%s/api/ws-test", serverAddr)
		resp, err := http.Get(testURL)
		if err != nil {
			log.Printf("❌ Server topilmadi: %v. Qayta urinish (10s)...\n", err)
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		// 2. WebSocket Handshake (using wss:// because server uses HTTPS)
		u := url.URL{Scheme: "wss", Host: serverAddr, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Printf("❌ Ulanishda xatolik: %v. Qayta urinish (5s)...\n", err)
			time.Sleep(5 * time.Second)
			continue
		}
		
		log.Println("✅ Connected! Professional zakazlar kutilmoqda...")
		
		handleMessages(c)
		c.Close()
		log.Println("⚠️  Ulanish uzildi. Qayta tiklanmoqda...")
		time.Sleep(2 * time.Second)
	}
}

var (
	lastOrderID   int
	lastPrintTime time.Time
)

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

		if m["type"] == "new_order" || m["type"] == "close_order" || m["type"] == "reprint_order" {
			orderData, _ := m["order"].(map[string]interface{})
			id := int(orderData["id"].(float64))

			isReprint := (m["type"] == "reprint_order")

			// Deduplication Logic: Aggressive 30-second filter for v8.5 Final
			now := time.Now()
			if !isReprint && id == lastOrderID && now.Sub(lastPrintTime) < 30*time.Second {
				log.Printf("🚫 DUB_RAD: Buyurtma #%d allaqachon chiqarilgan (Rad etildi).\n", id)
				continue
			}

			lastOrderID = id
			lastPrintTime = now

			if isReprint {
				log.Printf("🔔 MASTER v8.5: QAYTA CHOP ETISH buyrug'i #%d qabul qilindi (Faqat USB)!\n", id)
			} else {
				log.Printf("🔔 MASTER v8.5: Yangi buyurtma #%d qabul qilindi!\n", id)
			}
			
			printOrder(orderData, isReprint)
		}
	}
}

func printOrder(order map[string]interface{}, onlyUSB bool) {
	id := int(order["id"].(float64))
	
	targets := []string{"USB"}
	if !onlyUSB {
		targets = append(targets, networkPrinters...)
	}

	for _, target := range targets {
		var itemsForTarget []interface{}
		items, _ := order["items"].([]interface{})
		for _, it := range items {
			item := it.(map[string]interface{})
			targetVal, _ := item["printer_target"].(string)
			if targetVal == "" {
				targetVal = "ALL"
			}
			
			// Always print all items if it's a manual reprint (onlyUSB)
			if targetVal == "ALL" || targetVal == target || onlyUSB {
				itemsForTarget = append(itemsForTarget, item)
			}
		}

		if len(itemsForTarget) == 0 {
			continue // Skip printing if no items are routed to this printer
		}

		generateAndPrintReceipt(target, id, order, itemsForTarget)
	}
}

func generateAndPrintReceipt(target string, id int, order map[string]interface{}, items []interface{}) {
	f, err := os.CreateTemp("", fmt.Sprintf("order_%d_*.bin", id))
	if err != nil {
		log.Println("❌ Fayl yaratishda xato:", err)
		return
	}
	defer os.Remove(f.Name())

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(CODE_PAGE)
	f.Write(BEEP)

	// Header - MANGAL ОНЛАЙН ЗАКАЗ
	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write(toCP866("MANGAL ОНЛАЙН ЗАКАЗ\n"))
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n"))
	
	// Details
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("Чек №: %d\n", id)))
	f.Write(toCP866(fmt.Sprintf("Тип: Доставка %d\n", id)))
	f.Write(toCP866("Стол: онлайн\n"))
	f.Write(toCP866("Обслужил: MANGAL онлайн\n"))
	
	f.Write(toCP866(fmt.Sprintf("Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	f.Write(toCP866("Закрытие: -\n"))
	f.Write(toCP866("------------------------------------------------\n"))

	// Items Table
	f.Write(toCP866("Наименование           Кол-во Цена      Итого\n"))
	f.Write(toCP866("------------------------------------------------\n"))
	
	var targetTotal float64 = 0

	for _, it := range items {
		item := it.(map[string]interface{})
		name := item["product_name"].(string)
		qty := item["quantity"].(float64)
		price := item["price"].(float64)
		
		targetTotal += price * qty

		if len([]rune(name)) > 22 {
			name = string([]rune(name)[:19]) + "..."
		}
		
		nameRunes := []rune(name)
		paddedName := string(nameRunes)
		for i := len(nameRunes); i < 22; i++ {
			paddedName += " "
		}
		
		line := fmt.Sprintf("%s %-6.1f %-10.0f %-10.0f\n", 
			paddedName, qty, price, price*qty)
		f.Write(toCP866(line))
		
		if comment, ok := item["comment"].(string); ok && comment != "" {
			f.Write(toCP866(fmt.Sprintf("  * Коммент: %s\n", comment)))
		}
	}
	f.Write(toCP866("------------------------------------------------\n"))

	// Footer Summary (Calculate based on routed items only)
	f.Write(ALIGN_RIGHT)
	f.Write(toCP866(fmt.Sprintf("Подитог: %.0f\n", targetTotal)))
	f.Write(toCP866("Обслуживание (0.0%): 0\n"))
	f.Write(toCP866("Скидка (0%): 0\n"))
	f.Write([]byte("\n"))
	
	f.Write(FONT_DOUBLE_W)
	f.Write(toCP866(fmt.Sprintf("  ИТОГО: %.0f\n", targetTotal)))
	f.Write(FONT_NORMAL)
	f.Write([]byte("\n\n\n\n"))

	// Cut
	f.Write(PAPER_CUT)
	f.Close()

	if target == "USB" {
		cmd := exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice)
		if err := cmd.Run(); err != nil {
			log.Printf("❌ Chiqarishda xatolik (#%d - USB): %v\n", id, err)
		} else {
			log.Printf("✅ Professional chek #%d qirqildi (USB).\n", id)
		}
	} else {
		data, err := os.ReadFile(f.Name())
		if err == nil {
			conn, err := net.DialTimeout("tcp", target, 5*time.Second)
			if err != nil {
				log.Printf("❌ Tarmoq printeri xatosi (%s): %v\n", target, err)
				return
			}
			defer conn.Close()
			conn.Write(data)
			log.Printf("✅ Chek qirqildi (LAN: %s)\n", target)
		}
	}
}

func toCP866(text string) []byte {
	runes := []rune(text)
	res := make([]byte, 0, len(runes))
	for _, r := range runes {
		if r >= 0x0410 && r <= 0x042F { // А-Я
			res = append(res, byte(r-0x0410+0x80))
		} else if r >= 0x0430 && r <= 0x043F { // а-п
			res = append(res, byte(r-0x0430+0xA0))
		} else if r >= 0x0440 && r <= 0x044F { // р-я
			res = append(res, byte(r-0x0440+0xE0))
		} else if r == 0x0401 { // Ё
			res = append(res, 0xF0)
		} else if r == 0x0451 { // ё
			res = append(res, 0xF1)
		} else if r == 0x2116 { // №
			res = append(res, 0xFC)
		} else if r == 0x049A { // Қ
			res = append(res, byte('K'))
		} else if r == 0x049B { // қ
			res = append(res, byte('k'))
		} else if r == 0x0492 { // Ғ
			res = append(res, byte('G'))
		} else if r == 0x0493 { // ғ
			res = append(res, byte('g'))
		} else if r == 0x04B2 { // Ҳ
			res = append(res, byte('H'))
		} else if r == 0x04B3 { // ҳ
			res = append(res, byte('h'))
		} else if r == 0x040E { // Ў
			res = append(res, byte('O'))
		} else if r == 0x045E { // ў
			res = append(res, byte('o'))
		} else if r < 128 { // ASCII
			res = append(res, byte(r))
		} else { // Unknown
			res = append(res, '?')
		}
	}
	return res
}
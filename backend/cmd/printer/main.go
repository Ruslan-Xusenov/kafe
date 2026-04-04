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
	serverAddr    = "46.224.133.140:8080"
	printerKey    = "KAFE_PRINTER_SECRET_2026"
	printerDevice = "\\\\localhost\\XP-80C"
)

// ESC/POS Commands
var (
	ESC_INIT      = []byte{0x1B, 0x40}           // ESC @
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
	log.Println("🚀 Kafe Printer Bridge Artistic Pro (v4.0) ishga tushdi...")
	log.Printf("📍 Server: %s\n", serverAddr)
	log.Printf("📍 Printer: %s\n\n", printerDevice)

	for {
		// 1. Health Check
		testURL := fmt.Sprintf("http://%s/api/ws-test", serverAddr)
		resp, err := http.Get(testURL)
		if err != nil {
			log.Printf("❌ Server topilmadi: %v. Qayta urinish (10s)...\n", err)
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		// 2. WebSocket Handshake
		u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}
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
			log.Printf("🔔 Yangi buyurtma #%d (Artistic Pro) qabul qilindi!\n", id)
			
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

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(BEEP)

	// Header - YETUK ONLAYN ZAKAZ
	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write([]byte("YETUK ONLAYN ZAKAZ\n"))
	f.Write(FONT_NORMAL)
	f.Write([]byte("------------------------------------------------\n"))
	
	// Details
	f.Write(ALIGN_LEFT)
	f.Write([]byte(fmt.Sprintf("Check No / Chek raqami: %d\n", id)))
	f.Write([]byte(fmt.Sprintf("Tur / Tip: Dostavka / Доставка %d\n", id)))
	f.Write([]byte("Stol / Стол: onlayn\n"))
	f.Write([]byte("Xizmat / Сервис: YETUK KAFE onlayn\n"))
	
	f.Write([]byte(fmt.Sprintf("Ochilish vaqti / Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	f.Write([]byte("Yopilish vaqti / Закрытие: -\n"))
	f.Write([]byte("------------------------------------------------\n"))

	// Items Table
	f.Write([]byte("Mahsulot nomi          Soni   Narxi      Jami\n"))
	f.Write([]byte("Наименование           Кол.   Цена       Итого\n"))
	f.Write([]byte("------------------------------------------------\n"))
	
	items, _ := order["items"].([]interface{})
	for _, it := range items {
		item := it.(map[string]interface{})
		name := transliterate(item["product_name"].(string))
		qty := item["quantity"].(float64)
		price := item["price"].(float64)
		
		if len(name) > 22 {
			name = name[:19] + "..."
		}
		
		line := fmt.Sprintf("%-22s %-6.1f %-10.0f %-10.0f\n", 
			name, qty, price, price*qty)
		f.Write([]byte(line))
		
		if comment, ok := item["comment"].(string); ok && comment != "" {
			f.Write([]byte(fmt.Sprintf("  * Izoh / Изречение: %s\n", transliterate(comment))))
		}
	}
	f.Write([]byte("------------------------------------------------\n"))

	// Footer Summary
	f.Write(ALIGN_RIGHT)
	f.Write([]byte(fmt.Sprintf("Podditog / Подитог: %.0f\n", order["total_price"].(float64))))
	f.Write([]byte("Xizmat / Сервис (0.0%%): 0\n"))
	f.Write([]byte("Chegirma / Скидка (0%%): 0\n"))
	f.Write([]byte("\n"))
	
	f.Write(FONT_DOUBLE_W)
	f.Write([]byte(fmt.Sprintf("  JAMI / ИТОГО: %.0f\n", order["total_price"].(float64))))
	f.Write(FONT_NORMAL)
	f.Write([]byte("\n\n\n\n"))

	// Cut
	f.Write(PAPER_CUT)
	f.Close()

	// Windows print command (Direct binary copy)
	cmd := exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice)
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Chiqarishda xatolik (#%d): %v\n", id, err)
	} else {
		log.Printf("✅ Professional chek #%d qirqildi.\n", id)
	}
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
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
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var networkPrinters []string

var (
	serverAddr    string
	printerDevice string
)

var printerKey = os.Getenv("PRINTER_SECRET")

func init() {
	serverAddr = os.Getenv("API_HOST")
	if serverAddr == "" {
		serverAddr = "localhost:8080"
	}

	printerDevice = os.Getenv("PRINTER_DEVICE")
	if printerDevice == "" {
		printerDevice = "\\\\localhost\\XP-80C"
	}

	if printerKey == "" {
		printerKey = os.Getenv("PRINTER_KEY")
	}

	// Parse NETWORK_PRINTERS from env (comma separated, e.g. "192.168.1.10:9100,192.168.1.11:9100")
	np := os.Getenv("NETWORK_PRINTERS")
	if np != "" {
		for _, p := range strings.Split(np, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				networkPrinters = append(networkPrinters, p)
			}
		}
	}
}

// ESC/POS Commands
var (
	ESC_INIT        = []byte{0x1B, 0x40}           // ESC @
	DISABLE_CHINESE = []byte{0x1C, 0x2E}           // FS . (Disable Chinese character mode)
	CODE_PAGE       = []byte{0x1B, 0x74, 0x11}      // ESC t 17 (PC866)
	ALIGN_LEFT      = []byte{0x1B, 0x61, 0x00}      // ESC a 0
	ALIGN_CENTER    = []byte{0x1B, 0x61, 0x01}      // ESC a 1
	ALIGN_RIGHT     = []byte{0x1B, 0x61, 0x02}      // ESC a 2
	FONT_NORMAL     = []byte{0x1D, 0x21, 0x00}      // GS ! 0
	FONT_DOUBLE_H   = []byte{0x1D, 0x21, 0x01}      // GS ! 1
	FONT_DOUBLE_W   = []byte{0x1D, 0x21, 0x10}      // GS ! 16
	FONT_BIG        = []byte{0x1D, 0x21, 0x11}      // GS ! 17 (Double width + height)
	PAPER_CUT       = []byte{0x1D, 0x56, 0x42, 0x00} // GS V B 0
	BEEP            = []byte{0x1B, 0x42, 0x02, 0x02} // ESC B n t (Beep 2 times)
)

func main() {
	godotenv.Load("../backend/.env")
	godotenv.Load()

	log.Println("🚀 Kafe Printer Bridge Master v8.5 (Ultra Latin) ishga tushdi...")
	log.Printf("📍 Server: %s\n", serverAddr)
	log.Printf("📍 Printer: %s\n\n", printerDevice)

	for {
		// 1. Health Check
		scheme := "https"
		if os.Getenv("USE_SSL") == "false" {
			scheme = "http"
		}
		testURL := fmt.Sprintf("%s://%s/api/ws-test", scheme, serverAddr)
		resp, err := http.Get(testURL)
		if err != nil {
			log.Printf("❌ Server topilmadi: %v. Qayta urinish (10s)...\n", err)
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		// 2. WebSocket Handshake
		wsScheme := "wss"
		if os.Getenv("USE_SSL") == "false" {
			wsScheme = "ws"
		}
		u := url.URL{Scheme: wsScheme, Host: serverAddr, Path: "/api/ws", RawQuery: "printer_key=" + printerKey}
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

		if m["type"] == "shift_report" {
			log.Printf("🔔 MASTER v8.9: Z-OTCHYOT qabul qilindi!\n")
			printShiftReport(m)
			continue
		}

		if m["type"] == "bulk_edit" {
			orderID := int(m["order_id"].(float64))
			addedItems, _ := m["added_items"].([]interface{})
			cancelledItems, _ := m["cancelled_items"].([]interface{})
			waiterName, _ := m["waiter_name"].(string)
			tableNumber, _ := m["table_name"].(string)

			log.Printf("🔔 MASTER v8.9: O'ZGARISHLAR (#%d buyurtmadan) qabul qilindi!\n", orderID)
			printBulkEditItems(orderID, addedItems, cancelledItems, waiterName, tableNumber)
			continue
		}

		if m["type"] == "cancel_item" {
			orderID := int(m["order_id"].(float64))
			item := m["item"].(map[string]interface{})
			waiterName, _ := m["waiter_name"].(string)
			tableNumber, _ := m["table_name"].(string)

			log.Printf("🔔 MASTER v8.5: BEKOR QILISH (#%d buyurtmadan) qabul qilindi!\n", orderID)
			printCancelItem(orderID, item, waiterName, tableNumber)
			continue
		}

		if m["type"] == "new_order" || m["type"] == "close_order" || m["type"] == "reprint_order" {
			orderData, _ := m["order"].(map[string]interface{})
			id := int(orderData["id"].(float64))

			isReprint := (m["type"] == "reprint_order")
			isDop, _ := m["is_dop"].(bool)

			// Deduplication Logic: Aggressive 30-second filter for v8.5 Final
			now := time.Now()
			if !isReprint && !isDop && id == lastOrderID && now.Sub(lastPrintTime) < 30*time.Second {
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
			
			eventType := m["type"].(string)
			printOrder(orderData, isReprint, eventType)
		}
	}
}

func printCancelItem(orderID int, item map[string]interface{}, waiterName string, tableNumber string) {
	targetVal, _ := item["printer_target"].(string)
	if targetVal == "" {
		targetVal = "ALL"
	}

	targets := []string{"USB"}
	targets = append(targets, networkPrinters...)

	for _, target := range targets {
		if targetVal == "ALL" || targetVal == target {
			generateAndPrintCancelReceipt(target, orderID, item, waiterName, tableNumber)
		}
	}
}

func generateAndPrintCancelReceipt(target string, orderID int, item map[string]interface{}, waiterName string, tableNumber string) {
	f, err := os.CreateTemp("", fmt.Sprintf("cancel_%d_*.bin", orderID))
	if err != nil {
		return
	}
	defer os.Remove(f.Name())

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(DISABLE_CHINESE)
	f.Write(CODE_PAGE)
	f.Write(BEEP)
	f.Write(BEEP)

	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write(toCP866("ОТМЕНА / БЕКОР\n"))
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("Чек №: %d\n", orderID)))
	f.Write(toCP866(fmt.Sprintf("Стол: %s\n", tableNumber)))
	f.Write(toCP866(fmt.Sprintf("Официант: %s\n", waiterName)))
	f.Write(toCP866(fmt.Sprintf("Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	f.Write(toCP866("------------------------------------------------\n"))

	name := item["product_name"].(string)
	qty := item["quantity"].(float64)
	
	f.Write(FONT_DOUBLE_H)
	f.Write(toCP866(fmt.Sprintf("- %s x %.1f\n", name, qty)))
	
	if comment, ok := item["comment"].(string); ok && comment != "" {
		f.Write(FONT_NORMAL)
		f.Write(toCP866(fmt.Sprintf("  * %s\n", comment)))
	}
	
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n\n\n\n"))
	f.Write(PAPER_CUT)
	f.Close()

	if target == "USB" {
		exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice).Run()
	} else {
		data, err := os.ReadFile(f.Name())
		if err == nil {
			conn, err := net.DialTimeout("tcp", target, 5*time.Second)
			if err == nil {
				defer conn.Close()
				conn.Write(data)
			}
		}
	}
}

func printOrder(order map[string]interface{}, onlyUSB bool, eventType string) {
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

		generateAndPrintReceipt(target, id, order, itemsForTarget, eventType)
	}
}

func generateAndPrintReceipt(target string, id int, order map[string]interface{}, items []interface{}, eventType string) {
	f, err := os.CreateTemp("", fmt.Sprintf("order_%d_*.bin", id))
	if err != nil {
		log.Println("❌ Fayl yaratishda xato:", err)
		return
	}
	defer os.Remove(f.Name())

	isFinal := eventType == "close_order" || eventType == "reprint_order"

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(DISABLE_CHINESE)
	f.Write(CODE_PAGE)
	f.Write(BEEP)

	// Header
	if isFinal {
		f.Write(ALIGN_CENTER)
		f.Write(FONT_BIG)
		cafeFullName := os.Getenv("CAFE_FULL_NAME")
		if cafeFullName == "" {
			cafeFullName = os.Getenv("CAFE_NAME")
		}
		if cafeFullName == "" {
			cafeFullName = "Kafe"
		}
		f.Write(toCP866(cafeFullName + "\n"))
		f.Write(FONT_NORMAL)
		f.Write(toCP866("------------------------------------------------\n"))
	}
	
	// Details
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("Чек №: %d\n", id)))

	// Determine order type
	tableID, hasTable := order["table_id"]
	isTableOrder := hasTable && tableID != nil

	if isFinal {
		if isTableOrder {
			f.Write(toCP866("Тип: В заведении\n"))
		} else {
			f.Write(toCP866("Тип: Доставка\n"))
		}
	}

	// Table number
	tableNumber := "-"
	if tn, ok := order["table_name"]; ok && tn != nil {
		switch v := tn.(type) {
		case float64:
			if v > 0 {
				tableNumber = fmt.Sprintf("%.0f", v)
			}
		case string:
			if v != "" && v != "0" {
				tableNumber = v
			}
		}
	}
	f.Write(toCP866(fmt.Sprintf("Стол: %s\n", tableNumber)))

	// Waiter name
	waiterName := os.Getenv("CAFE_FULL_NAME")
	if waiterName == "" {
		waiterName = os.Getenv("CAFE_NAME")
	}
	if waiterName == "" {
		waiterName = "Kafe"
	}
	if wn, ok := order["waiter_name"]; ok && wn != nil {
		if s, ok := wn.(string); ok && s != "" {
			waiterName = s
		}
	}
	f.Write(toCP866(fmt.Sprintf("Обслужил: %s\n", waiterName)))
	
	f.Write(toCP866(fmt.Sprintf("Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	if isFinal {
		f.Write(toCP866("Закрытие: -\n"))
	}
	f.Write(toCP866("------------------------------------------------\n"))

	// Items Table
	if isFinal {
		f.Write(toCP866("Наименование           Кол-во Цена      Итого\n"))
	} else {
		f.Write(toCP866("Наименование           Кол-во\n"))
	}
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
		
		if isFinal {
			line := fmt.Sprintf("%s %-6.1f %-10.0f %-10.0f\n", 
				paddedName, qty, price, price*qty)
			f.Write(toCP866(line))
		} else {
			line := fmt.Sprintf("%s %-6.1f\n", paddedName, qty)
			// Double height and width for kitchen receipt items to make them clear
			f.Write(FONT_DOUBLE_H)
			f.Write(toCP866(line))
			f.Write(FONT_NORMAL)
		}
		
		if comment, ok := item["comment"].(string); ok && comment != "" {
			f.Write(toCP866(fmt.Sprintf("  * Коммент: %s\n", comment)))
		}
	}
	f.Write(toCP866("------------------------------------------------\n"))

	// Footer Summary (Calculate based on routed items only)
	if isFinal {
		var servicePercentage float64 = 0
		if sp, ok := order["service_percentage"].(float64); ok {
			servicePercentage = sp
		}
		
		serviceFee := targetTotal * servicePercentage / 100
		finalTotal := targetTotal + serviceFee

		f.Write(ALIGN_RIGHT)
		f.Write(toCP866(fmt.Sprintf("Подитог: %.0f\n", targetTotal)))
		f.Write(toCP866(fmt.Sprintf("Обслуживание (%.1f%%): %.0f\n", servicePercentage, serviceFee)))
		f.Write(toCP866("Скидка (0%): 0\n"))
		f.Write([]byte("\n"))
		
		f.Write(FONT_DOUBLE_W)
		f.Write(toCP866(fmt.Sprintf("  ИТОГО: %.0f\n", finalTotal)))
		f.Write(FONT_NORMAL)
	}
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










func printBulkEditItems(orderID int, addedItems, cancelledItems []interface{}, waiterName, tableNumber string) {
	// Group items by printer target
	targetsMap := make(map[string]struct{})
	
	// Default target if none specified
	for _, it := range addedItems {
		item := it.(map[string]interface{})
		t, _ := item["printer_target"].(string)
		if t == "" { t = "ALL" }
		targetsMap[t] = struct{}{}
	}
	for _, it := range cancelledItems {
		item := it.(map[string]interface{})
		t, _ := item["printer_target"].(string)
		if t == "" { t = "ALL" }
		targetsMap[t] = struct{}{}
	}

	targets := []string{"USB"}
	targets = append(targets, networkPrinters...)

	for _, target := range targets {
		var addedForTarget []interface{}
		var cancelledForTarget []interface{}

		for _, it := range addedItems {
			item := it.(map[string]interface{})
			t, _ := item["printer_target"].(string)
			if t == "" { t = "ALL" }
			if t == "ALL" || t == target {
				addedForTarget = append(addedForTarget, item)
			}
		}

		for _, it := range cancelledItems {
			item := it.(map[string]interface{})
			t, _ := item["printer_target"].(string)
			if t == "" { t = "ALL" }
			if t == "ALL" || t == target {
				cancelledForTarget = append(cancelledForTarget, item)
			}
		}

		if len(addedForTarget) == 0 && len(cancelledForTarget) == 0 {
			continue
		}

		generateAndPrintBulkEditReceipt(target, orderID, addedForTarget, cancelledForTarget, waiterName, tableNumber)
	}
}

func generateAndPrintBulkEditReceipt(target string, orderID int, addedItems, cancelledItems []interface{}, waiterName, tableNumber string) {
	f, err := os.CreateTemp("", fmt.Sprintf("bulk_%d_*.bin", orderID))
	if err != nil {
		return
	}
	defer os.Remove(f.Name())

	// Build Binary Sequence
	f.Write(ESC_INIT)
	f.Write(DISABLE_CHINESE)
	f.Write(CODE_PAGE)
	f.Write(BEEP)
	f.Write(BEEP)

	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write(toCP866("ИЗМЕНЕНИЯ / O'ZGARISHLAR\n"))
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("Чек №: %d\n", orderID)))
	f.Write(toCP866(fmt.Sprintf("Стол: %s\n", tableNumber)))
	f.Write(toCP866(fmt.Sprintf("Официант: %s\n", waiterName)))
	f.Write(toCP866(fmt.Sprintf("Время: %s\n", time.Now().Format("02.01.2006 15:04:05"))))
	f.Write(toCP866("------------------------------------------------\n\n"))

	f.Write(FONT_DOUBLE_H)
	if len(addedItems) > 0 {
		f.Write(toCP866("[+] ДОБАВЛЕНО:\n"))
		for _, it := range addedItems {
			item := it.(map[string]interface{})
			name := item["product_name"].(string)
			qty := item["quantity"].(float64)
			f.Write(toCP866(fmt.Sprintf("  + %s x %.1f\n", name, qty)))
		}
		f.Write([]byte("\n"))
	}

	if len(cancelledItems) > 0 {
		f.Write(toCP866("[-] ОТМЕНЕНО:\n"))
		for _, it := range cancelledItems {
			item := it.(map[string]interface{})
			name := item["product_name"].(string)
			qty := item["quantity"].(float64)
			f.Write(toCP866(fmt.Sprintf("  - %s x %.1f\n", name, qty)))
		}
		f.Write([]byte("\n"))
	}
	
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n\n\n\n"))
	f.Write(PAPER_CUT)
	f.Close()

	if target == "USB" {
		exec.Command("cmd", "/c", "copy", "/b", f.Name(), printerDevice).Run()
	} else {
		data, err := os.ReadFile(f.Name())
		if err == nil {
			conn, err := net.DialTimeout("tcp", target, 5*time.Second)
			if err == nil {
				defer conn.Close()
				conn.Write(data)
			}
		}
	}
}

func printShiftReport(m map[string]interface{}) {
	target := os.Getenv("PRINTER_USB")
	if target == "" {
		target = "USB"
	}
	f, err := os.CreateTemp("", "shift_report_*.bin")
	if err != nil {
		return
	}
	defer os.Remove(f.Name())

	// Extract values safely
	timestamp, _ := m["timestamp"].(string)
	totalRev := safeFloat(m["total_revenue"])
	totalExp := safeFloat(m["total_expenses"])
	netProfit := safeFloat(m["net_profit"])
	cash := safeFloat(m["cash"])
	card := safeFloat(m["card"])
	click := safeFloat(m["click"])
	nasiya := safeFloat(m["nasiya"])

	f.Write(ESC_INIT)
	f.Write(DISABLE_CHINESE)
	f.Write(CODE_PAGE)
	f.Write(BEEP)
	f.Write(BEEP)

	f.Write(ALIGN_CENTER)
	f.Write(FONT_BIG)
	f.Write(toCP866("ИТОГИ СМЕНЫ\n(Z-ОТЧЕТ)\n"))
	f.Write(FONT_NORMAL)
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("ВРЕМЯ ЗАКРЫТИЯ: %s\n", timestamp)))
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(FONT_BIG)
	f.Write(toCP866(fmt.Sprintf("ОБЩАЯ ВЫРУЧКА: %.0f sum\n", totalRev)))
	f.Write(FONT_NORMAL)
	f.Write(toCP866(fmt.Sprintf("ОБЩИЕ РАСХОДЫ: %.0f sum\n", totalExp)))
	f.Write(toCP866(fmt.Sprintf("ЧИСТАЯ ПРИБЫЛЬ: %.0f sum\n", netProfit)))
	f.Write(toCP866("------------------------------------------------\n"))
	
	f.Write(ALIGN_CENTER)
	f.Write(toCP866("СПОСОБЫ ОПЛАТЫ\n"))
	f.Write(ALIGN_LEFT)
	f.Write(toCP866(fmt.Sprintf("НАЛИЧНЫЕ:       %.0f sum\n", cash)))
	f.Write(toCP866(fmt.Sprintf("ТЕРМИНАЛ:       %.0f sum\n", card)))
	f.Write(toCP866(fmt.Sprintf("CLICK/PAYME:    %.0f sum\n", click)))
	f.Write(toCP866(fmt.Sprintf("В ДОЛГ:         %.0f sum\n", nasiya)))
	
	f.Write(toCP866("------------------------------------------------\n\n\n\n\n"))
	f.Write(PAPER_CUT)
	f.Close()

	cmd := exec.Command("cmd", "/c", fmt.Sprintf("copy /b %s %s", f.Name(), target))
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Z-OTCHYOT chop etishda xatolik: %v\n", err)
	} else {
		log.Printf("✅ Z-OTCHYOT muvaffaqiyatli chop etildi!\n")
	}
}

func safeFloat(val interface{}) float64 {
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}

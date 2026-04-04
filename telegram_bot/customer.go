package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram_bot/models"
	"net/http"
	"os"
	"path/filepath"
)

func (b *Bot) handleStart(chatID int64) {
	text := "🍽 <b>Kafe Botiga Xush Kelibsiz!</b>\n\nMenyu ko'rish, savatchaga yig'ish va buyurtma berish uchun quyidagi tugmalardan foydalaning."
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍕 Menyu", "show_categories"), tgbotapi.NewInlineKeyboardButtonData("🛒 Savatcha", "show_cart")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📦 Buyurtmalarim", "show_orders")),
	)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) showMainMenu(chatID int64) {
	text := "🏠 <b>Asosiy Menyu</b>"
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍕 Menyu", "show_categories"), tgbotapi.NewInlineKeyboardButtonData("🛒 Savatcha", "show_cart")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📦 Buyurtmalarim", "show_orders")),
	}
	if b.IsCook(chatID) { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍳 Oshxona Paneli", "cook_active_orders"))) }
	if b.IsAdmin(chatID) { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⚙️ Admin Panel", "admin_panel"))) }
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) showCategories(chatID int64) {
	categories, err := b.repo.GetAllCategories()
	if err != nil || len(categories) == 0 { b.sendMessage(chatID, "📭 Hozircha kategoriyalar yo'q"); return }
	text := "📂 <b>Kategoriyalar</b>"
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, cat := range categories { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("📁 %s", cat.Name), fmt.Sprintf("cat_%d", cat.ID)))) }
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏠 Asosiy Menyu", "main_menu")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) showCategoryProducts(chatID int64, data string) {
	catID, _ := strconv.Atoi(strings.TrimPrefix(data, "cat_"))
	products, err := b.repo.GetByCategoryID(catID)
	if err != nil || len(products) == 0 { b.sendMessage(chatID, "📭 Bu kategoriyada mahsulotlar yo'q"); return }
	text := "📂 <b>Mahsulotlar</b>"
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, prod := range products { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🍽 %s — %.0f so'm", prod.Name, prod.Price), fmt.Sprintf("prod_%d", prod.ID)))) }
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Kategoriyalar", "show_categories")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) showProductDetail(chatID int64, data string) {
	prodID, _ := strconv.Atoi(strings.TrimPrefix(data, "prod_"))
	prod, err := b.repo.GetProductByID(prodID)
	if err != nil || prod == nil { b.sendMessage(chatID, "❌ Mahsulot topilmadi"); return }
	text := fmt.Sprintf("🍽 <b>%s</b>\n\n📝 %s\n💰 <b>%.0f so'm</b>", prod.Name, prod.Description, prod.Price)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Savatchaga qo'shish", fmt.Sprintf("add_%d", prod.ID))),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Orqaga", fmt.Sprintf("cat_%d", prod.CategoryID)), tgbotapi.NewInlineKeyboardButtonData("🛒 Savatcha", "show_cart")),
	)
	if prod.ImageURL != "" {
		imgURL := prod.ImageURL
		localPath := ""
		if !strings.HasPrefix(imgURL, "http") {
			// Try to find local path
			path := strings.TrimPrefix(imgURL, "/")
			localPath = filepath.Join("../backend", path)
		}

		if localPath != "" {
			if _, err := os.Stat(localPath); err == nil {
				photo := tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(localPath))
				photo.Caption = text; photo.ReplyMarkup = keyboard; photo.ParseMode = "HTML"
				if _, err := b.api.Send(photo); err == nil { return }
			}
		}

		// Fallback to URL if local failed
		if !strings.HasPrefix(imgURL, "http") {
			baseURL := strings.TrimSuffix(b.apiBaseURL, "/")
			path := strings.TrimPrefix(imgURL, "/")
			imgURL = fmt.Sprintf("%s/%s", baseURL, path)
		}
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imgURL))
		photo.Caption = text; photo.ReplyMarkup = keyboard; photo.ParseMode = "HTML"
		if _, err := b.api.Send(photo); err == nil { return }
	}
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) addToCart(chatID int64, data string) {
	prodID, _ := strconv.Atoi(strings.TrimPrefix(data, "add_"))
	prod, err := b.repo.GetProductByID(prodID)
	if err != nil || prod == nil { b.sendMessage(chatID, "❌ Mahsulot topilmadi"); return }
	session := b.getSession(chatID)
	found := false
	for i := range session.Cart { if session.Cart[i].ProductID == prodID { session.Cart[i].Quantity++; found = true; break } }
	if !found { session.Cart = append(session.Cart, CartItem{ProductID: prodID, ProductName: prod.Name, Price: prod.Price, Quantity: 1}) }
	text := fmt.Sprintf("✅ <b>%s</b> savatchaga qo'shildi!", prod.Name)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍕 Menyu", fmt.Sprintf("cat_%d", prod.CategoryID)), tgbotapi.NewInlineKeyboardButtonData("🛒 Savatcha", "show_cart")))
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) showCart(chatID int64) {
	session := b.getSession(chatID)
	if len(session.Cart) == 0 { b.sendMessage(chatID, "🛒 Savatcha bo'sh"); return }
	text := "🛒 <b>Savatcha</b>\n\n"; var total float64; var rows [][]tgbotapi.InlineKeyboardButton
	for i, item := range session.Cart {
		itotal := item.Price * float64(item.Quantity); total += itotal
		text += fmt.Sprintf("%d. <b>%s</b>\n   %d x %.0f = <b>%.0f so'm</b>\n", i+1, item.ProductName, item.Quantity, item.Price, itotal)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➖", fmt.Sprintf("qty_dec_%d", item.ProductID)), tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", item.Quantity), "qty_info"), tgbotapi.NewInlineKeyboardButtonData("➕", fmt.Sprintf("qty_inc_%d", item.ProductID))))
	}
	text += fmt.Sprintf("\n💰 <b>Jami: %.0f so'm</b>", total)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✅ Buyurtma berish", "checkout"), tgbotapi.NewInlineKeyboardButtonData("🗑 Tozalash", "clear_cart")))
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏠 Menyu", "main_menu")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) changeQuantity(chatID int64, data string) {
	session := b.getSession(chatID)
	if strings.HasPrefix(data, "qty_inc_") {
		id, _ := strconv.Atoi(strings.TrimPrefix(data, "qty_inc_"))
		for i := range session.Cart { if session.Cart[i].ProductID == id { session.Cart[i].Quantity++; break } }
	} else {
		id, _ := strconv.Atoi(strings.TrimPrefix(data, "qty_dec_"))
		for i := range session.Cart { if session.Cart[i].ProductID == id { session.Cart[i].Quantity--; if session.Cart[i].Quantity <= 0 { session.Cart = append(session.Cart[:i], session.Cart[i+1:]...) }; break } }
	}
	b.showCart(chatID)
}

func (b *Bot) clearCart(chatID int64) { b.getSession(chatID).Cart = []CartItem{}; b.sendMessage(chatID, "🗑 Savatcha tozalandi!"); b.showMainMenu(chatID) }

func (b *Bot) startCheckout(chatID int64) {
	if len(b.getSession(chatID).Cart) == 0 { b.sendMessage(chatID, "🛒 Savatcha bo'sh"); return }
	b.getSession(chatID).State = "awaiting_order_address"
	b.sendMessage(chatID, "📍 Yetkazib berish manzilini kiriting:")
}

func (b *Bot) confirmOrder(chatID int64) {
	s := b.getSession(chatID)
	var total float64; for _, it := range s.Cart { total += it.Price * float64(it.Quantity) }
	comment := ""
	if s.Comment != "" { comment = "\n💬 Izoh: " + s.Comment }
	text := fmt.Sprintf("📋 <b>Buyurtmani tasdiqlang:</b>\n📍 Manzil: %s\n📞 Telefon: %s%s\n\n💰 Jami: <b>%.0f so'm</b>", s.Address, s.OrderPhone, comment, total)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✅ Tasdiqlash", "confirm_order"), tgbotapi.NewInlineKeyboardButtonData("❌ Bekor qilish", "cancel_order")))
	b.sendMessageWithKeyboard(chatID, text, keyboard)
}

func (b *Bot) placeOrder(chatID int64) {
	s := b.getSession(chatID)
	comment := s.Comment
	order := &models.Order{CustomerID: nil, Address: s.Address, Phone: s.OrderPhone, Comment: &comment, Status: models.StatusNew}
	var total float64
	for _, it := range s.Cart { order.Items = append(order.Items, models.OrderItem{ProductID: it.ProductID, Quantity: it.Quantity, Price: it.Price}); total += it.Price * float64(it.Quantity) }
	order.TotalPrice = total
	if err := b.repo.CreateOrder(order); err != nil { b.sendMessage(chatID, "❌ Buyurtmada xatolik"); return }
	
	// Notify Backend for Printer and Real-time UI
	go b.notifyAPI(order.ID)
	
	b.notifyCooks(order)

	s.Cart = []CartItem{}; s.Address = ""; s.OrderPhone = ""; s.Comment = ""
	b.sendMessage(chatID, fmt.Sprintf("✅ <b>Buyurtma #%d qabul qilindi!</b>\n\nTez orada tayyor bo'ladi.", order.ID))
}

func (b *Bot) notifyAPI(orderID int) {
	key := "KAFE_PRINTER_SECRET_2026" // Standard key as per backend config
	url := fmt.Sprintf("http://localhost:8080/api/notify-order/%d?key=%s", orderID, key)
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("TELEGRAM_DEBUG: Failed to notify printer for order %d: %v\n", orderID, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("TELEGRAM_DEBUG: Printer notified for order %d. Status: %s\n", orderID, resp.Status)
}

func (b *Bot) showOrderHistory(chatID int64) {
	orders, _ := b.repo.GetAllOrders()
	if len(orders) == 0 { b.sendMessage(chatID, "📦 Hali buyurtma berilmagan"); return }
	text := "📦 <b>Oxirgi Buyurtmalar:</b>"
	var rows [][]tgbotapi.InlineKeyboardButton
	for i, o := range orders { if i > 10 { break }; rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("#%d | %.0f so'm | %s", o.ID, o.TotalPrice, o.Status), fmt.Sprintf("order_detail_%d", o.ID)))) }
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏠 Asosiy Menyu", "main_menu")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) showOrderDetail(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "order_detail_"))
	o, _ := b.repo.GetOrderByID(id)
	if o == nil { b.sendMessage(chatID, "❌ Buyurtma topilmadi"); return }
	text := fmt.Sprintf("📋 <b>Buyurtma #%d</b>\nHolati: %s\nManzil: %s\nTelefon: %s\n\n", o.ID, o.Status, o.Address, o.Phone)
	for _, it := range o.Items {
		name := "Noma'lum"
		if it.ProductName != nil { name = *it.ProductName }
		text += fmt.Sprintf("• %s x %d = %.0f so'm\n", name, it.Quantity, it.Price*float64(it.Quantity))
	}
	text += fmt.Sprintf("\n💰 Jami: %.0f so'm", o.TotalPrice)
	rows := [][]tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Orqaga", "show_orders"))}
	if o.Status == models.StatusDelivered { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⭐ Baho berish", fmt.Sprintf("rate_order_%d", o.ID)))) }
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) startRating(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "rate_order_"))
	s := b.getSession(chatID); s.RatingOrderID = id; s.State = "awaiting_rating_cook"
	b.sendMessage(chatID, "⭐ <b>Baho bering</b>\n\n🍳 Oshpazga baho bering (1-5):")
}

func (b *Bot) submitRating(chatID int64) {
	s := b.getSession(chatID); o, _ := b.repo.GetOrderByID(s.RatingOrderID)
	if o != nil {
		if o.CookID != nil { b.repo.AddStaffRating(&models.StaffRating{OrderID: s.RatingOrderID, StaffID: *o.CookID, StaffRole: "cook", Rating: s.RatingCookScore}) }
		if o.CourierID != nil { b.repo.AddStaffRating(&models.StaffRating{OrderID: s.RatingOrderID, StaffID: *o.CourierID, StaffRole: "courier", Rating: s.RatingCourierScore}) }
		
		// Notify Courier
		cid := int64(0)
		if o.CourierTelegramID != nil { cid = *o.CourierTelegramID }
		if cid == 0 {
			b.ordersMu.RLock(); cid = b.ordersToCouriers[s.RatingOrderID]; b.ordersMu.RUnlock()
		}
		if cid != 0 {
			b.sendMessage(cid, fmt.Sprintf("🌟 <b>Buyurtma #%d uchun yangi baho!</b>\n\nMijoz sizga <b>%d ⭐</b> baho berdi.", s.RatingOrderID, s.RatingCourierScore))
		}
	}
	b.sendMessage(chatID, "🌟 Rahmat! Sizning bahoingiz qabul qilindi.")
	s.RatingOrderID = 0; b.showMainMenu(chatID)
}

func (b *Bot) sendHelp(chatID int64) {
	b.sendMessage(chatID, "📋 <b>Yordam</b>\n\n/start - Boshlash\n/menu - Menyu\n/cart - Savatcha\n/orders - Tarix\n/admin - Admin\n/cook - Oshpaz")
}

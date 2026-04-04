package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram_bot/models"
)

func (b *Bot) handleAdminCommand(chatID int64) {
	if b.superAdminID == 0 { b.superAdminID = chatID; b.sendMessage(chatID, "👑 Siz asosiy admin bitasiz!") }
	if !b.IsAdmin(chatID) { b.sendMessage(chatID, "❌ Admin huquqi yo'q"); return }
	b.showAdminPanel(chatID)
}

func (b *Bot) showAdminPanel(chatID int64) {
	t := "⚙️ <b>Admin Panel</b>"
	r := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📊 Statistika", "admin_stats"), tgbotapi.NewInlineKeyboardButtonData("🍽 Mahsulotlar", "admin_products")),
	}
	if b.IsSuperAdmin(chatID) { r = append(r, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("👥 Admin Boshqaruvi", "admin_manage_admins"))) }
	r = append(r, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🏠 Menyu", "main_menu")))
	b.sendMessageWithKeyboard(chatID, t, tgbotapi.NewInlineKeyboardMarkup(r...))
}

func (b *Bot) showAdminStats(chatID int64) {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📅 1 kun", "admin_stats_1d"), tgbotapi.NewInlineKeyboardButtonData("📅 3 kun", "admin_stats_3d")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("📅 1 hafta", "admin_stats_1w"), tgbotapi.NewInlineKeyboardButtonData("📅 1 yil", "admin_stats_1y")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Admin Panel", "admin_panel")),
	)
	b.sendMessageWithKeyboard(chatID, "📊 <b>Statistika</b>\nDavrni tanlang:", kb)
}

func (b *Bot) showAdminStatsPeriod(chatID int64, data string) {
	p := strings.TrimPrefix(data, "admin_stats_")
	var sqlPeriod string
	switch p {
	case "1d": sqlPeriod = "CURRENT_DATE"
	case "3d": sqlPeriod = "CURRENT_DATE - INTERVAL '3 days'"
	case "1w": sqlPeriod = "CURRENT_DATE - INTERVAL '1 week'"
	case "1y": sqlPeriod = "CURRENT_DATE - INTERVAL '1 year'"
	default: sqlPeriod = "CURRENT_DATE"
	}
	type Stats struct { Total int `db:"total"`; Completed int `db:"comp"`; Sales float64 `db:"sales"`; Churned int `db:"churned"`; TotalU int `db:"totalu"` }
	var s Stats
	q := fmt.Sprintf(`SELECT 
		(SELECT COUNT(*) FROM orders WHERE created_at >= %s) as total,
		(SELECT COUNT(*) FROM orders WHERE status = 'delivered' AND created_at >= %s) as comp,
		(SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'delivered' AND created_at >= %s) as sales,
		(SELECT COUNT(*) FROM users) as totalu,
		(SELECT COUNT(*) FROM users u WHERE NOT EXISTS (SELECT 1 FROM orders o WHERE o.customer_id = u.id AND o.created_at >= CURRENT_DATE - INTERVAL '30 days')) as churned
	`, sqlPeriod, sqlPeriod, sqlPeriod)
	_ = b.repo.GetDB().Get(&s, q)
	text := fmt.Sprintf("📊 <b>Statistika (%s)</b>\n\n📦 Buyurtmalar: %d\n✅ Bajarilgan: %d\n💰 Savdo: %.0f so'm\n\n👥 Foydalanuvchilar: %d\n📉 Chiqib ketgan: %d", p, s.Total, s.Completed, s.Sales, s.TotalU, s.Churned)
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Statistika", "admin_stats"))))
}

func (b *Bot) showAdminProducts(chatID int64) {
	products, _ := b.repo.GetAllProducts()
	text := fmt.Sprintf("🍽 <b>Mahsulotlar</b> (%d)\n", len(products))
	var rows [][]tgbotapi.InlineKeyboardButton
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Yangi mahsulot", "admin_add_product")))
	for _, p := range products {
		icon := "✅"; if !p.IsActive { icon = "❌" }
		text += fmt.Sprintf("%s <b>%s</b> — %.0f so'm\n", icon, p.Name, p.Price)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✏️ "+p.Name, fmt.Sprintf("admin_edit_prod_%d", p.ID)), tgbotapi.NewInlineKeyboardButtonData("🗑", fmt.Sprintf("admin_del_prod_%d", p.ID))))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Admin Panel", "admin_panel")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) startAddProduct(chatID int64) {
	cats, _ := b.repo.GetAllCategories()
	if len(cats) == 0 { b.sendMessage(chatID, "❌ Kategoriyalar topilmadi (sayt orqali qo'shing)"); return }
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, cat := range cats { rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(cat.Name, fmt.Sprintf("admin_prod_cat_%d", cat.ID)))) }
	b.sendMessageWithKeyboard(chatID, "➕ <b>Yangi mahsulot</b>\nKategoriyani tanlang:", tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) selectProductCategory(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_prod_cat_"))
	s := b.getSession(chatID); s.AdminProductEdit = &AdminProductData{Action: "create", CategoryID: id, IsActive: true}; s.State = "admin_product_name"
	b.sendMessage(chatID, "📝 Nomini kiriting:")
}

func (b *Bot) startEditProduct(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_edit_prod_"))
	p, _ := b.repo.GetProductByID(id)
	if p != nil { s := b.getSession(chatID); s.AdminProductEdit = &AdminProductData{Action: "edit", ID: p.ID, CategoryID: p.CategoryID, Name: p.Name, Description: p.Description, Price: p.Price, ImageURL: p.ImageURL, IsActive: p.IsActive}; s.State = "admin_product_name"; b.sendMessage(chatID, "✏️ Tahrirlash: Nomini yuboring (yoki eskisinni):") }
}

func (b *Bot) saveProduct(chatID int64) {
	pd := b.getSession(chatID).AdminProductEdit
	p := &models.Product{ID: pd.ID, CategoryID: pd.CategoryID, Name: pd.Name, Description: pd.Description, Price: pd.Price, ImageURL: pd.ImageURL, IsActive: pd.IsActive}
	var err error
	if pd.Action == "create" { err = b.repo.CreateProduct(p) } else { err = b.repo.UpdateProduct(p) }
	if err == nil { b.sendMessage(chatID, "✅ Saqlandi!"); b.getSession(chatID).AdminProductEdit = nil; b.showAdminProducts(chatID) }
}

func (b *Bot) deleteProduct(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_del_prod_"))
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Ha, o'chirish", fmt.Sprintf("admin_confirm_del_%d", id)), tgbotapi.NewInlineKeyboardButtonData("Yo'q", "admin_products")))
	b.sendMessageWithKeyboard(chatID, "⚠️ O'chirishga ishonchingiz komilmi?", kb)
}

func (b *Bot) confirmDeleteProduct(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_confirm_del_"))
	if err := b.repo.DeleteProduct(id); err == nil { b.sendMessage(chatID, "✅ O'chirildi!"); b.showAdminProducts(chatID) }
}

func (b *Bot) showAdminManagement(chatID int64) {
	t := fmt.Sprintf("👥 <b>Adminlar:</b>\n👑 Asosiy: %d\n", b.superAdminID)
	b.adminsMu.RLock(); for id := range b.admins { t += fmt.Sprintf("• %d\n", id) }; b.adminsMu.RUnlock()
	t += "\n<b>Oshpazlar:</b>\n"
	b.cooksMu.RLock(); for id := range b.cooks { t += fmt.Sprintf("• %d\n", id) }; b.cooksMu.RUnlock()
	t += "\n<b>Kuryerlar:</b>\n"
	b.couriersMu.RLock(); for id := range b.couriers { t += fmt.Sprintf("• %d\n", id) }; b.couriersMu.RUnlock()
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("+Admin", "admin_add_admin"), tgbotapi.NewInlineKeyboardButtonData("-Admin", "admin_remove_admin")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("+Oshpaz", "admin_add_cook"), tgbotapi.NewInlineKeyboardButtonData("-Oshpaz", "admin_remove_cook")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("+Kuryer", "admin_add_courier"), tgbotapi.NewInlineKeyboardButtonData("-Kuryer", "admin_remove_courier")),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔙 Admin Panel", "admin_panel")),
	)
	b.sendMessageWithKeyboard(chatID, t, kb)
}

func (b *Bot) startAddAdmin(chatID int64) { s := b.getSession(chatID); s.AdminAction = "add_admin"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "Admin Telegram ID ni yuboring:") }
func (b *Bot) startRemoveAdmin(chatID int64) { s := b.getSession(chatID); s.AdminAction = "remove_admin"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "O'chirish uchun ID ni yuboring:") }
func (b *Bot) startAddCook(chatID int64) { s := b.getSession(chatID); s.AdminAction = "add_cook"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "Oshpaz Telegram ID ni yuboring:") }
func (b *Bot) startAddCourier(chatID int64) { s := b.getSession(chatID); s.AdminAction = "add_courier"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "Kuryer Telegram ID ni yuboring:") }
func (b *Bot) startRemoveCook(chatID int64) { s := b.getSession(chatID); s.AdminAction = "remove_cook"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "O'chirish uchun Oshpaz ID ni yuboring:") }
func (b *Bot) startRemoveCourier(chatID int64) { s := b.getSession(chatID); s.AdminAction = "remove_courier"; s.State = "awaiting_admin_id"; b.sendMessage(chatID, "O'chirish uchun Kuryer ID ni yuboring:") }

func (b *Bot) processAdminAction(chatID int64, input string) {
	s := b.getSession(chatID); id, _ := strconv.ParseInt(strings.TrimPrefix(strings.TrimSpace(input), "@"), 10, 64)
	if id == 0 { b.sendMessage(chatID, "❌ Noto'g'ri ID"); return }
	switch s.AdminAction {
	case "add_admin": b.AddAdmin(id); b.sendMessage(id, "👑 Siz admin etib tayinlandingiz!")
	case "remove_admin": if id != b.superAdminID { b.RemoveAdmin(id); b.sendMessage(id, "⚠️ Admin huquqlari olib tashlandi") }
	case "add_cook": b.AddCook(id); b.sendMessage(id, "👨‍🍳 Siz oshpaz etib tayinlandingiz!")
	case "remove_cook": b.RemoveCook(id); b.sendMessage(id, "⚠️ Oshpaz huquqlari olib tashlandi")
	case "add_courier": b.AddCourier(id); b.sendMessage(id, "🚚 Siz kuryer etib tayinlandingiz!")
	case "remove_courier": b.RemoveCourier(id); b.sendMessage(id, "⚠️ Kuryer huquqlari olib tashlandi")
	}
	s.AdminAction = ""; s.State = ""; b.showAdminManagement(chatID)
}

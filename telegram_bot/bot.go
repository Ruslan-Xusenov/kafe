package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram_bot/db"
	"telegram_bot/repository"
	"telegram_bot/service"
)

type Bot struct {
	api            *tgbotapi.BotAPI
	repo           *repository.BotRepository
	printerService *service.PrinterService
	sessions       map[int64]*UserSession
	sessionsMu     sync.RWMutex
	superAdminID   int64
	admins         map[int64]bool
	adminsMu       sync.RWMutex
	cooks          map[int64]bool
	cooksMu        sync.RWMutex
	couriers       map[int64]bool
	couriersMu     sync.RWMutex
	ordersToCouriers map[int]int64
	ordersMu       sync.RWMutex
	apiBaseURL     string
}

type UserSession struct {
	ChatID   int64
	State    string
	Cart     []CartItem
	Phone    string
	FullName string
	Address  string
	Comment  string
	OrderPhone string
	RatingOrderID   int
	RatingCookScore int
	RatingCourierScore int
	AdminProductEdit *AdminProductData
	AdminAction string
}

type CartItem struct {
	ProductID   int
	ProductName string
	Price       float64
	Quantity    int
}

type AdminProductData struct {
	Action      string
	Step        string
	ID          int
	CategoryID  int
	Name        string
	Description string
	Price       float64
	ImageURL    string
	IsActive    bool
}

func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	if !strings.HasPrefix(apiURL, "http") {
		apiURL = "http://" + apiURL
	}

	var superAdminID int64
	fmt.Sscanf(os.Getenv("SUPER_ADMIN_ID"), "%d", &superAdminID)

	repo := repository.NewBotRepository(db.DB)
	_ = repo.EnsureStaffTable()

	bot := &Bot{
		api:            api,
		repo:           repo,
		printerService: service.NewPrinterService(),
		sessions:       make(map[int64]*UserSession),
		admins:         make(map[int64]bool),
		cooks:          make(map[int64]bool),
		couriers:       make(map[int64]bool),
		ordersToCouriers: make(map[int]int64),
		apiBaseURL:     apiURL,
		superAdminID:   superAdminID,
	}

	// Load staff from DB
	if ids, err := repo.GetStaffByRole("admin"); err == nil {
		for _, id := range ids { bot.admins[id] = true }
	}
	if ids, err := repo.GetStaffByRole("cook"); err == nil {
		for _, id := range ids { bot.cooks[id] = true }
	}
	if ids, err := repo.GetStaffByRole("courier"); err == nil {
		for _, id := range ids { bot.couriers[id] = true }
	}

	return bot, nil
}

func (b *Bot) getSession(chatID int64) *UserSession {
	b.sessionsMu.Lock()
	defer b.sessionsMu.Unlock()
	if s, ok := b.sessions[chatID]; ok {
		return s
	}
	s := &UserSession{ChatID: chatID, Cart: []CartItem{}}
	b.sessions[chatID] = s
	return s
}

func (b *Bot) IsAdmin(chatID int64) bool {
	b.adminsMu.RLock()
	defer b.adminsMu.RUnlock()
	return b.superAdminID == chatID || b.admins[chatID]
}

func (b *Bot) IsSuperAdmin(chatID int64) bool {
	return b.superAdminID == chatID
}

func (b *Bot) IsCook(chatID int64) bool {
	b.cooksMu.RLock()
	defer b.cooksMu.RUnlock()
	return b.cooks[chatID]
}

func (b *Bot) AddAdmin(chatID int64) {
	b.adminsMu.Lock()
	defer b.adminsMu.Unlock()
	b.admins[chatID] = true
	_ = b.repo.AddStaff(chatID, "admin")
}

func (b *Bot) RemoveAdmin(chatID int64) {
	b.adminsMu.Lock()
	defer b.adminsMu.Unlock()
	delete(b.admins, chatID)
	_ = b.repo.RemoveStaff(chatID)
}

func (b *Bot) AddCook(chatID int64) {
	b.cooksMu.Lock()
	defer b.cooksMu.Unlock()
	b.cooks[chatID] = true
	_ = b.repo.AddStaff(chatID, "cook")
}

func (b *Bot) RemoveCook(chatID int64) {
	b.cooksMu.Lock()
	defer b.cooksMu.Unlock()
	delete(b.cooks, chatID)
	_ = b.repo.RemoveStaff(chatID)
}

func (b *Bot) IsCourier(chatID int64) bool {
	b.couriersMu.RLock()
	defer b.couriersMu.RUnlock()
	return b.couriers[chatID]
}

func (b *Bot) AddCourier(chatID int64) {
	b.couriersMu.Lock()
	defer b.couriersMu.Unlock()
	b.couriers[chatID] = true
	_ = b.repo.AddStaff(chatID, "courier")
}

func (b *Bot) RemoveCourier(chatID int64) {
	b.couriersMu.Lock()
	defer b.couriersMu.Unlock()
	delete(b.couriers, chatID)
	_ = b.repo.RemoveStaff(chatID)
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		if update.Message != nil {
			b.handleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
		}
	}
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	session := b.getSession(chatID)

	if msg.IsCommand() {
		switch msg.Command() {
		case "start": b.handleStart(chatID); return
		case "menu": b.showMainMenu(chatID); return
		case "cart": b.showCart(chatID); return
		case "orders": b.showOrderHistory(chatID); return
		case "admin": b.handleAdminCommand(chatID); return
		case "cook": b.handleCookCommand(chatID); return
		case "courier": b.handleCourierCommand(chatID); return
		case "help": b.sendHelp(chatID); return
		}
	}

	switch session.State {
	case "awaiting_order_address":
		session.Address = msg.Text
		session.State = "awaiting_order_phone"; b.sendMessage(chatID, "📞 Telefon raqamingizni kiriting:")
	case "awaiting_order_phone":
		session.OrderPhone = msg.Text
		session.State = "awaiting_order_comment"
		keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⏩ O'tkazib yuborish", "skip_comment")))
		reply := tgbotapi.NewMessage(chatID, "💬 Buyurtmaga izoh qo'shmoqchimisiz?")
		reply.ReplyMarkup = keyboard; b.api.Send(reply)
	case "awaiting_order_comment": session.Comment = msg.Text; session.State = ""; b.confirmOrder(chatID)
	case "awaiting_rating_cook":
		rating := parseRating(msg.Text)
		if rating == 0 { b.sendMessage(chatID, "❌ Iltimos, 1 dan 5 gacha baho bering"); return }
		session.RatingCookScore = rating; session.State = "awaiting_rating_courier"; b.sendMessage(chatID, "🚗 Kuryerga baho bering (1-5):")
	case "awaiting_rating_courier":
		rating := parseRating(msg.Text)
		if rating == 0 { b.sendMessage(chatID, "❌ Iltimos, 1 dan 5 gacha baho bering"); return }
		session.RatingCourierScore = rating; session.State = ""; b.submitRating(chatID)
	case "awaiting_admin_id": b.processAdminAction(chatID, msg.Text)
	case "admin_product_name":
		if session.AdminProductEdit != nil { session.AdminProductEdit.Name = msg.Text; session.AdminProductEdit.Step = "description"; session.State = "admin_product_description"; b.sendMessage(chatID, "📝 Mahsulot tavsifini kiriting:") }
	case "admin_product_description":
		if session.AdminProductEdit != nil { session.AdminProductEdit.Description = msg.Text; session.AdminProductEdit.Step = "price"; session.State = "admin_product_price"; b.sendMessage(chatID, "💰 Narxini kiriting (masalan: 25000):") }
	case "admin_product_price":
		if session.AdminProductEdit != nil {
			price := parsePrice(msg.Text)
			if price <= 0 { b.sendMessage(chatID, "❌ Noto'g'ri narx. Faqat raqam kiriting:"); return }
			session.AdminProductEdit.Price = price; session.AdminProductEdit.Step = "image"; session.State = "admin_product_image"
			keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⏩ O'tkazib yuborish", "skip_product_image")))
			reply := tgbotapi.NewMessage(chatID, "🖼 Mahsulot rasmining URL manzilini kiriting:"); reply.ReplyMarkup = keyboard; b.api.Send(reply)
		}
	case "admin_product_image":
		if session.AdminProductEdit != nil { session.AdminProductEdit.ImageURL = msg.Text; session.State = ""; b.saveProduct(chatID) }
	}
}

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	data := cb.Data
	b.api.Request(tgbotapi.NewCallback(cb.ID, ""))
	switch {
	case data == "main_menu": b.showMainMenu(chatID)
	case data == "show_categories": b.showCategories(chatID)
	case strings.HasPrefix(data, "cat_"): b.showCategoryProducts(chatID, data)
	case strings.HasPrefix(data, "prod_"): b.showProductDetail(chatID, data)
	case strings.HasPrefix(data, "add_"): b.addToCart(chatID, data)
	case strings.HasPrefix(data, "qty_"): b.changeQuantity(chatID, data)
	case data == "show_cart": b.showCart(chatID)
	case data == "clear_cart": b.clearCart(chatID)
	case data == "checkout": b.startCheckout(chatID)
	case data == "skip_comment": session := b.getSession(chatID); session.Comment = ""; session.State = ""; b.confirmOrder(chatID)
	case data == "confirm_order": b.placeOrder(chatID)
	case data == "cancel_order": session := b.getSession(chatID); session.State = ""; b.sendMessage(chatID, "❌ Buyurtma bekor qilindi"); b.showMainMenu(chatID)
	case data == "show_orders": b.showOrderHistory(chatID)
	case strings.HasPrefix(data, "order_detail_"): b.showOrderDetail(chatID, data)
	case strings.HasPrefix(data, "rate_order_"): b.startRating(chatID, data)
	case data == "cook_active_orders": b.showCookActiveOrders(chatID)
	case strings.HasPrefix(data, "cook_accept_"): b.cookAcceptOrder(chatID, data)
	case strings.HasPrefix(data, "cook_ready_"): b.cookReadyOrder(chatID, data)
	case strings.HasPrefix(data, "cook_print_"): b.cookPrintOrder(chatID, data)
	case data == "courier_active_orders": b.showCourierActiveOrders(chatID)
	case strings.HasPrefix(data, "courier_accept_"): b.courierAcceptOrder(chatID, data)
	case strings.HasPrefix(data, "courier_delivered_"): b.courierDeliveredOrder(chatID, data)
	case data == "admin_panel": b.showAdminPanel(chatID)
	case data == "admin_stats": b.showAdminStats(chatID)
	case strings.HasPrefix(data, "admin_stats_"): b.showAdminStatsPeriod(chatID, data)
	case data == "admin_products": b.showAdminProducts(chatID)
	case data == "admin_add_product": b.startAddProduct(chatID)
	case strings.HasPrefix(data, "admin_prod_cat_"): b.selectProductCategory(chatID, data)
	case strings.HasPrefix(data, "admin_edit_prod_"): b.startEditProduct(chatID, data)
	case strings.HasPrefix(data, "admin_del_prod_"): b.deleteProduct(chatID, data)
	case strings.HasPrefix(data, "admin_confirm_del_"): b.confirmDeleteProduct(chatID, data)
	case data == "admin_manage_admins": b.showAdminManagement(chatID)
	case data == "admin_add_admin": b.startAddAdmin(chatID)
	case data == "admin_remove_admin": b.startRemoveAdmin(chatID)
	case data == "admin_add_cook": b.startAddCook(chatID)
	case data == "admin_remove_cook": b.startRemoveCook(chatID)
	case data == "admin_add_courier": b.startAddCourier(chatID)
	case data == "admin_remove_courier": b.startRemoveCourier(chatID)
	case data == "skip_product_image": session := b.getSession(chatID); if session.AdminProductEdit != nil { session.AdminProductEdit.ImageURL = ""; session.State = ""; b.saveProduct(chatID) }
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"; b.api.Send(msg)
}

func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text); msg.ParseMode = "HTML"; msg.ReplyMarkup = keyboard; b.api.Send(msg)
}

func parseRating(text string) int {
	switch strings.TrimSpace(text) {
	case "1": return 1; case "2": return 2; case "3": return 3; case "4": return 4; case "5": return 5; default: return 0
	}
}

func parsePrice(text string) float64 {
	var price float64; fmt.Sscanf(strings.ReplaceAll(text, " ", ""), "%f", &price); return price
}

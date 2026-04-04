package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram_bot/models"
)

func (b *Bot) handleCourierCommand(chatID int64) {
	if !b.IsCourier(chatID) && !b.IsAdmin(chatID) {
		b.sendMessage(chatID, "❌ Sizda kuryer huquqlari yo'q")
		return
	}
	b.showCourierActiveOrders(chatID)
}

func (b *Bot) showCourierActiveOrders(chatID int64) {
	orders, err := b.repo.GetAllOrders()
	if err != nil {
		b.sendMessage(chatID, "❌ Xato")
		return
	}

	var available []models.Order
	var myOrders []models.Order
	for _, o := range orders {
		if o.Status == models.StatusReady {
			available = append(available, o)
		} else if o.Status == models.StatusOnWay && o.CourierTelegramID != nil && *o.CourierTelegramID == chatID {
			myOrders = append(myOrders, o)
		}
	}

	text := "🚗 <b>Kuryer Paneli</b>\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton

	if len(available) > 0 {
		text += "📦 <b>Tayyor buyurtmalar (Olib ketish uchun):</b>\n"
		for _, o := range available {
			text += fmt.Sprintf("<b>#%d</b> | %.0f so'm\n📍 %s\n\n", o.ID, o.TotalPrice, o.Address)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("🚚 Olish #%d", o.ID), fmt.Sprintf("courier_accept_%d", o.ID))))
		}
	}

	if len(myOrders) > 0 {
		text += "🚴 <b>Mening buyurtmalarim (Yo'lda):</b>\n"
		for _, o := range myOrders {
			text += fmt.Sprintf("<b>#%d</b> | 📞 %s\n📍 %s\n\n", o.ID, o.Phone, o.Address)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✅ Topshirildi #%d", o.ID), fmt.Sprintf("courier_delivered_%d", o.ID))))
		}
	}

	if len(available) == 0 && len(myOrders) == 0 {
		text += "📭 Hozircha yangi buyurtmalar yo'q."
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🔄 Yangilash", "courier_active_orders"), tgbotapi.NewInlineKeyboardButtonData("🏠 Menyu", "main_menu")))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) courierAcceptOrder(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "courier_accept_"))
	
	// Bazada kuryerni biriktirish
	if err := b.repo.AssignCourierByTelegramID(id, chatID); err != nil {
		b.sendMessage(chatID, "❌ Xato: Buyurtmani qabul qilib bo'lmadi")
		return
	}
	
	b.ordersMu.Lock()
	b.ordersToCouriers[id] = chatID
	b.ordersMu.Unlock()
	
	b.sendMessage(chatID, fmt.Sprintf("🚚 Buyurtma #%d qabul qilindi. Yo'lga tushdingiz!", id))
	b.showCourierActiveOrders(chatID)
}

func (b *Bot) courierDeliveredOrder(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "courier_delivered_"))
	if err := b.repo.UpdateStatus(id, models.StatusDelivered, nil); err != nil {
		b.sendMessage(chatID, "❌ Xato")
		return
	}
	b.sendMessage(chatID, fmt.Sprintf("✅ Buyurtma #%d topshirildi. Baraka toping!", id))
	b.showCourierActiveOrders(chatID)
}

func (b *Bot) notifyCouriers(order *models.Order) {
	text := fmt.Sprintf("🚗 <b>TAYYOR BUYURTMA #%d!</b>\n\nTayyor bo'ldi, olib ketishingiz mumkin.\n📍 %s", order.ID, order.Address)
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🚚 Olish", fmt.Sprintf("courier_accept_%d", order.ID))))
	
	var targets []int64
	b.couriersMu.RLock()
	for id := range b.couriers { targets = append(targets, id) }
	b.couriersMu.RUnlock()
	
	for _, id := range targets {
		b.sendMessageWithKeyboard(id, text, kb)
	}
}
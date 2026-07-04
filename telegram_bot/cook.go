package main

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"telegram_bot/models"
)

func (b *Bot) handleCookCommand(chatID int64) {
	if !b.IsCook(chatID) && !b.IsAdmin(chatID) { b.sendMessage(chatID, "❌ У вас нет прав повара"); return }
	b.showCookActiveOrders(chatID)
}

func (b *Bot) showCookActiveOrders(chatID int64) {
	orders, err := b.repo.GetAllOrders()
	if err != nil { b.sendMessage(chatID, "❌ Ошибка"); return }
	var active []models.Order
	for _, o := range orders { 
		// Oshpaz faqat Yangi va Tayyorlanayotgan zakazlarni ko'rishi kerak
		if o.Status == models.StatusNew || o.Status == models.StatusPreparing { 
			active = append(active, o) 
		} 
	}
	if len(active) == 0 { b.sendMessage(chatID, "✅ Нет активных заказов"); return }
	
	text := "🍳 <b>Кухня</b>\n\n"
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, o := range active {
		statusIcon := "🆕"; if o.Status == models.StatusPreparing { statusIcon = "👨‍🍳" }
		text += fmt.Sprintf("<b>#%d</b> | %s %s\n📍 %s\n📞 %s\n💰 %.0f сум\n\n", o.ID, statusIcon, o.Status, o.Address, o.Phone, o.TotalPrice)
		
		switch o.Status {
		case models.StatusNew:
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("👨‍🍳 Принять #%d", o.ID), fmt.Sprintf("cook_accept_%d", o.ID)),
			))
		case models.StatusPreparing:
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("✅ Готово #%d", o.ID), fmt.Sprintf("cook_ready_%d", o.ID)),
			))
		}
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔄 Обновить", "cook_active_orders"), 
		tgbotapi.NewInlineKeyboardButtonData("🏠 Меню", "main_menu"),
	))
	b.sendMessageWithKeyboard(chatID, text, tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) cookAcceptOrder(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "cook_accept_"))
	if err := b.repo.UpdateStatus(id, models.StatusPreparing, nil); err != nil { b.sendMessage(chatID, "❌ Ошибка"); return }
	b.sendMessage(chatID, fmt.Sprintf("🍳 Заказ #%d принят!", id)); b.showCookActiveOrders(chatID)
}

func (b *Bot) cookReadyOrder(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "cook_ready_"))
	if err := b.repo.UpdateStatus(id, models.StatusReady, nil); err != nil {
		b.sendMessage(chatID, "❌ Ошибка")
		return
	}
	b.sendMessage(chatID, fmt.Sprintf("✅ Заказ #%d готов!", id))
	b.showCookActiveOrders(chatID)

	// Notify Couriers
	o, _ := b.repo.GetOrderByID(id)
	if o != nil {
		go b.notifyCouriers(o)
	}
}

func (b *Bot) cookPrintOrder(chatID int64, data string) {
	id, _ := strconv.Atoi(strings.TrimPrefix(data, "cook_print_"))
	o, _ := b.repo.GetOrderByID(id)
	if o != nil {
		err := b.printerService.PrintOrder(o)
		if err != nil {
			b.sendMessage(chatID, fmt.Sprintf("❌ Ошибка принтера: %v", err))
		} else {
			b.sendMessage(chatID, "🖨 Печать...")
		}
	}
}

func (b *Bot) notifyCooks(order *models.Order) {
	text := fmt.Sprintf("🔔 <b>НОВЫЙ ЗАКАЗ #%d!</b>\n\n💰 %.0f сум\n📍 %s", order.ID, order.TotalPrice, order.Address)
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("🍳 Принять", fmt.Sprintf("cook_accept_%d", order.ID))))
	var targets []int64
	b.cooksMu.RLock(); for id := range b.cooks { targets = append(targets, id) }; b.cooksMu.RUnlock()
	b.adminsMu.RLock(); for id := range b.admins { targets = append(targets, id) }; b.adminsMu.RUnlock()
	if b.superAdminID != 0 { targets = append(targets, b.superAdminID) }
	for _, id := range targets { msg := tgbotapi.NewMessage(id, text); msg.ParseMode = "HTML"; msg.ReplyMarkup = kb; b.api.Send(msg) }
}

package service

import (
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/username/kafe-backend/internal/models"
)

type BotService struct {
	bot     *tgbotapi.BotAPI
	chatIDs []int64
}

func NewBotService() *BotService {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("Warning: TELEGRAM_BOT_TOKEN not set")
		return &BotService{}
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("Failed to initialize Telegram Bot: %v", err)
		return &BotService{}
	}

	// Get Chat IDs from env (comma-separated group or admin IDs)
	raw := os.Getenv("TELEGRAM_REPORT_CHAT_IDS")
	if raw == "" {
		raw = os.Getenv("TELEGRAM_CHAT_ID")
	}

	var chatIDs []int64
	for _, idStr := range strings.Split(raw, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr != "" {
			var id int64
			if _, err := fmt.Sscanf(idStr, "%d", &id); err == nil && id != 0 {
				chatIDs = append(chatIDs, id)
			}
		}
	}

	return &BotService{
		bot:     bot,
		chatIDs: chatIDs,
	}
}

func (s *BotService) SendNewOrderNotification(order *models.Order, imageUrl *string) {
	if s.bot == nil || len(s.chatIDs) == 0 {
		return
	}

	msgText := fmt.Sprintf("🔔 *Новый заказ! #%d*\n\n", order.ID)
	msgText += fmt.Sprintf("👤 Клиент: %s\n", order.Phone)
	msgText += fmt.Sprintf("📍 Адрес: %s\n\n", order.Address)
	msgText += "*Продукты:*\n"
	
	for _, item := range order.Items {
		msgText += fmt.Sprintf("- %.1fx %s\n", item.Quantity, item.ProductName)
	}
	
	msgText += fmt.Sprintf("\n💰 *Итого: %s сум*", fmt.Sprintf("%.0f", order.TotalPrice))

	for _, chatID := range s.chatIDs {
		var msg tgbotapi.Chattable
		if imageUrl != nil && *imageUrl != "" {
			// Attempt to send as photo
			var photo tgbotapi.PhotoConfig
			if strings.HasPrefix(*imageUrl, "http://") || strings.HasPrefix(*imageUrl, "https://") {
				photo = tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(*imageUrl))
			} else {
				imgPath := *imageUrl
				if strings.HasPrefix(imgPath, "/") {
					imgPath = "." + imgPath
				}
				photo = tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
			}
			photo.Caption = msgText
			photo.ParseMode = "Markdown"
			msg = photo
		} else {
			// Fallback to text message
			textMsg := tgbotapi.NewMessage(chatID, msgText)
			textMsg.ParseMode = "Markdown"
			msg = textMsg
		}

		_, err := s.bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send Telegram notification to %d: %v", chatID, err)
			
			// If photo sending failed (e.g., bad URL), retry as plain text
			if imageUrl != nil && *imageUrl != "" {
				textMsg := tgbotapi.NewMessage(chatID, msgText)
				textMsg.ParseMode = "Markdown"
				s.bot.Send(textMsg)
			}
		}
	}
}

func (s *BotService) SendOrderStatusNotification(order *models.Order, statusMsg string) {
	if s.bot == nil || len(s.chatIDs) == 0 {
		return
	}

	msgText := fmt.Sprintf("📦 *Статус заказа изменен: #%d*\n\n", order.ID)
	msgText += fmt.Sprintf("Статус: *%s*\n", statusMsg)
	msgText += fmt.Sprintf("Клиент: %s\n", order.Phone)
	msgText += fmt.Sprintf("Адрес: %s\n", order.Address)

	for _, chatID := range s.chatIDs {
		textMsg := tgbotapi.NewMessage(chatID, msgText)
		textMsg.ParseMode = "Markdown"
		
		_, err := s.bot.Send(textMsg)
		if err != nil {
			log.Printf("Failed to send order status notification to %d: %v", chatID, err)
		}
	}
}
func (s *BotService) SendNotificationToAll(msgText string, imageUrl string) {
	if s.bot == nil || len(s.chatIDs) == 0 {
		return
	}

	for _, chatID := range s.chatIDs {
		var msg tgbotapi.Chattable

		if imageUrl != "" {
			var photo tgbotapi.PhotoConfig
			if strings.HasPrefix(imageUrl, "http") {
				photo = tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(imageUrl))
			} else {
				imgPath := imageUrl
				if strings.HasPrefix(imgPath, "/") {
					imgPath = "." + imgPath
				}
				photo = tgbotapi.NewPhoto(chatID, tgbotapi.FilePath(imgPath))
			}
			photo.Caption = msgText
			photo.ParseMode = "Markdown"
			msg = photo
		} else {
			textMsg := tgbotapi.NewMessage(chatID, msgText)
			textMsg.ParseMode = "Markdown"
			msg = textMsg
		}

		_, err := s.bot.Send(msg)
		if err != nil {
			log.Printf("Failed to send Telegram notification to %d: %v", chatID, err)
		}
	}
}

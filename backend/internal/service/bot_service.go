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
	bot    *tgbotapi.BotAPI
	chatID int64
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

	// Get Chat ID from env (group or admin ID)
	var chatID int64
	fmt.Sscanf(os.Getenv("TELEGRAM_CHAT_ID"), "%d", &chatID)

	return &BotService{
		bot:    bot,
		chatID: chatID,
	}
}

func (s *BotService) SendNewOrderNotification(order *models.Order, imageUrl *string) {
	if s.bot == nil || s.chatID == 0 {
		return
	}

	msgText := fmt.Sprintf("🔔 *Yangi Buyurtma! #%d*\n\n", order.ID)
	msgText += fmt.Sprintf("👤 Mijoz: %s\n", order.Phone)
	msgText += fmt.Sprintf("📍 Manzil: %s\n\n", order.Address)
	msgText += "*Mahsulotlar:*\n"
	
	for _, item := range order.Items {
		msgText += fmt.Sprintf("- %.1fx %s\n", item.Quantity, item.ProductName)
	}
	
	msgText += fmt.Sprintf("\n💰 *Jami: %s so'm*", fmt.Sprintf("%.0f", order.TotalPrice))

	var msg tgbotapi.Chattable
	if imageUrl != nil && *imageUrl != "" {
		// Attempt to send as photo
		var photo tgbotapi.PhotoConfig
		if strings.HasPrefix(*imageUrl, "http://") || strings.HasPrefix(*imageUrl, "https://") {
			photo = tgbotapi.NewPhoto(s.chatID, tgbotapi.FileURL(*imageUrl))
		} else {
			imgPath := *imageUrl
			if strings.HasPrefix(imgPath, "/") {
				imgPath = "." + imgPath
			}
			photo = tgbotapi.NewPhoto(s.chatID, tgbotapi.FilePath(imgPath))
		}
		photo.Caption = msgText
		photo.ParseMode = "Markdown"
		msg = photo
	} else {
		// Fallback to text message
		textMsg := tgbotapi.NewMessage(s.chatID, msgText)
		textMsg.ParseMode = "Markdown"
		msg = textMsg
	}

	_, err := s.bot.Send(msg)
	if err != nil {
		log.Printf("Failed to send Telegram notification: %v", err)
		
		// If photo sending failed (e.g., bad URL), retry as plain text
		if imageUrl != nil && *imageUrl != "" {
			textMsg := tgbotapi.NewMessage(s.chatID, msgText)
			textMsg.ParseMode = "Markdown"
			s.bot.Send(textMsg)
		}
	}
}

func (s *BotService) SendOrderStatusNotification(order *models.Order, statusMsg string) {
	if s.bot == nil || s.chatID == 0 {
		return
	}

	msgText := fmt.Sprintf("📦 *Buyurtma Holati O'zgardi: #%d*\n\n", order.ID)
	msgText += fmt.Sprintf("Holati: *%s*\n", statusMsg)
	msgText += fmt.Sprintf("Mijoz: %s\n", order.Phone)
	msgText += fmt.Sprintf("Manzil: %s\n", order.Address)

	textMsg := tgbotapi.NewMessage(s.chatID, msgText)
	textMsg.ParseMode = "Markdown"
	
	_, err := s.bot.Send(textMsg)
	if err != nil {
		log.Printf("Failed to send order status notification: %v", err)
	}
}
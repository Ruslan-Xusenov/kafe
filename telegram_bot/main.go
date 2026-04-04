package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"telegram_bot/db"
)

func main() {
	if err := godotenv.Load("../backend/.env"); err != nil {
		log.Printf("Warning: ../backend/.env file not found, trying .env")
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: .env file not found")
		}
	}

	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}

	bot, err := NewBot(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	log.Println("🤖 Standalone Kafe Telegram Bot started!")
	bot.Start()
}

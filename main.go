package main

import (
	"log"
	"os"
	"tele-uni/config"
	"tele-uni/routes"

	"github.com/joho/godotenv"
	"gopkg.in/telebot.v3"
)

var (
	SheetID               = "1Y6RMu5Ceb2ZCSpmthYHlv-g7r4DCrUSdRP0PrK32k_Y" // Replace with your Sheet ID
	ConditionResponseData [][]string                                       // Global cache for condition-response data
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	botToken := os.Getenv("BOT_TOKEN")

	db := config.ConnectDB()

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// Create a bot instance
	bot, err := telebot.NewBot(telebot.Settings{
		Token: botToken,
	})
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	handler := routes.Handler(bot, db)

	log.Println("Bot is running...")
	handler.Start()
}

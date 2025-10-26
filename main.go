package main

import (
	"log"
	"os"
	"telecust/api"
	"telecust/bot"
	"telecust/database"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	err = database.InitDB("telecust.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Get bot token from environment (support both variable names)
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		botToken = os.Getenv("TELE_BOT_TOKEN")
	}
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN or TELE_BOT_TOKEN environment variable is required")
	}

	// Check OpenAI API key
	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		log.Println("Warning: OPENAI_API_KEY not set. Bot will not be able to respond intelligently.")
	} else {
		log.Println("OpenAI API configured successfully")
	}

	// Initialize bot
	err = bot.InitBot(botToken)
	if err != nil {
		log.Fatalf("Failed to initialize bot: %v", err)
	}

	// Start bot in goroutine
	go bot.GlobalBot.Start()

	// Start API server (blocking)
	api.StartServer()
}

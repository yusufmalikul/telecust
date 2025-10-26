package bot

import (
	"log"
	"telecust/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	API *tgbotapi.BotAPI
}

var GlobalBot *Bot

func InitBot(token string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return err
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	GlobalBot = &Bot{API: bot}
	return nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		go b.handleMessage(update.Message)
	}
}

func (b *Bot) handleMessage(message *tgbotapi.Message) {
	// Get or create conversation
	conv, err := database.GetOrCreateConversation(
		message.Chat.ID,
		message.From.UserName,
		message.From.FirstName,
	)
	if err != nil {
		log.Printf("Error getting conversation: %v", err)
		return
	}

	// Save user message
	err = database.SaveMessage(conv.ID, "user", message.Text)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		return
	}

	// Handle /start and /help commands
	if message.IsCommand() {
		switch message.Command() {
		case "start", "help":
			b.sendMessage(message.Chat.ID, "Halo! Saya siap membantu Anda. Silakan tanyakan apa saja!")
			database.SaveMessage(conv.ID, "bot", "Halo! Saya siap membantu Anda. Silakan tanyakan apa saja!")
			return
		}
	}

	// Check if bot is active for this conversation
	if !conv.IsBotActive {
		// Bot is in takeover mode, don't respond
		log.Printf("Bot inactive for chat %d, admin mode", message.Chat.ID)
		return
	}

	// Query knowledge base
	kb, err := database.GetKnowledgeBase()
	if err != nil {
		log.Printf("Error getting knowledge base: %v", err)
		kb = ""
	}

	response := QueryKnowledgeBase(message.Text, kb)

	// Send response
	b.sendMessage(message.Chat.ID, response)

	// Save bot response
	err = database.SaveMessage(conv.ID, "bot", response)
	if err != nil {
		log.Printf("Error saving bot response: %v", err)
	}
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.API.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}

// SendMessageAsAdmin sends a message from admin to user
func (b *Bot) SendMessageAsAdmin(chatID int64, text string, conversationID int) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.API.Send(msg)
	if err != nil {
		return err
	}

	// Save admin message
	return database.SaveMessage(conversationID, "admin", text)
}

package database

import "time"

type Conversation struct {
	ID               int       `json:"id"`
	TelegramChatID   int64     `json:"telegram_chat_id"`
	TelegramUsername string    `json:"telegram_username"`
	TelegramFirstName string   `json:"telegram_first_name"`
	IsBotActive      bool      `json:"is_bot_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	LastMessage      string    `json:"last_message,omitempty"`
	LastMessageTime  time.Time `json:"last_message_time,omitempty"`
}

type Message struct {
	ID             int       `json:"id"`
	ConversationID int       `json:"conversation_id"`
	SenderType     string    `json:"sender_type"` // 'user', 'bot', 'admin'
	MessageText    string    `json:"message_text"`
	CreatedAt      time.Time `json:"created_at"`
}

type KnowledgeBase struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	UpdatedAt time.Time `json:"updated_at"`
}

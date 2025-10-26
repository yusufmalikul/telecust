package database

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create tables
	schema := `
	CREATE TABLE IF NOT EXISTS conversations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegram_chat_id INTEGER UNIQUE NOT NULL,
		telegram_username TEXT,
		telegram_first_name TEXT,
		is_bot_active BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		conversation_id INTEGER NOT NULL,
		sender_type TEXT NOT NULL,
		message_text TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (conversation_id) REFERENCES conversations(id)
	);

	CREATE TABLE IF NOT EXISTS knowledge_base (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_messages_created ON messages(created_at);
	`

	_, err = DB.Exec(schema)
	if err != nil {
		return err
	}

	// Insert default knowledge base if empty
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM knowledge_base").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		defaultKB := `Harga kentang Rp5ribu perbungkus.
Pesan diatas 10 harga 4rb.
Jika pesan 10 Rp40ribu.
Jika pesan 20 Rp80ribu.
Jika pesan di atas 100 bungkus harga Rp3ribu.`
		_, err = DB.Exec("INSERT INTO knowledge_base (content) VALUES (?)", defaultKB)
		if err != nil {
			return err
		}
	}

	log.Println("Database initialized successfully")
	return nil
}

// GetOrCreateConversation finds or creates a conversation for a Telegram chat
func GetOrCreateConversation(chatID int64, username, firstName string) (*Conversation, error) {
	var conv Conversation
	err := DB.QueryRow(`
		SELECT id, telegram_chat_id, telegram_username, telegram_first_name, is_bot_active, created_at, updated_at
		FROM conversations WHERE telegram_chat_id = ?
	`, chatID).Scan(&conv.ID, &conv.TelegramChatID, &conv.TelegramUsername, &conv.TelegramFirstName, &conv.IsBotActive, &conv.CreatedAt, &conv.UpdatedAt)

	if err == sql.ErrNoRows {
		// Create new conversation
		result, err := DB.Exec(`
			INSERT INTO conversations (telegram_chat_id, telegram_username, telegram_first_name)
			VALUES (?, ?, ?)
		`, chatID, username, firstName)
		if err != nil {
			return nil, err
		}

		id, _ := result.LastInsertId()
		conv.ID = int(id)
		conv.TelegramChatID = chatID
		conv.TelegramUsername = username
		conv.TelegramFirstName = firstName
		conv.IsBotActive = true
		conv.CreatedAt = time.Now()
		conv.UpdatedAt = time.Now()
		return &conv, nil
	}

	if err != nil {
		return nil, err
	}

	return &conv, nil
}

// SaveMessage saves a message to the database
func SaveMessage(conversationID int, senderType, messageText string) error {
	_, err := DB.Exec(`
		INSERT INTO messages (conversation_id, sender_type, message_text)
		VALUES (?, ?, ?)
	`, conversationID, senderType, messageText)

	if err != nil {
		return err
	}

	// Update conversation updated_at
	_, err = DB.Exec(`
		UPDATE conversations SET updated_at = CURRENT_TIMESTAMP WHERE id = ?
	`, conversationID)

	return err
}

// GetAllConversations returns all conversations with their last message
func GetAllConversations() ([]Conversation, error) {
	rows, err := DB.Query(`
		SELECT c.id, c.telegram_chat_id, c.telegram_username, c.telegram_first_name,
		       c.is_bot_active, c.created_at, c.updated_at,
		       COALESCE(m.message_text, '') as last_message,
		       COALESCE(m.created_at, c.created_at) as last_message_time
		FROM conversations c
		LEFT JOIN (
			SELECT m1.conversation_id, m1.message_text, m1.created_at
			FROM messages m1
			INNER JOIN (
				SELECT conversation_id, MAX(created_at) as max_time
				FROM messages
				GROUP BY conversation_id
			) m2 ON m1.conversation_id = m2.conversation_id AND m1.created_at = m2.max_time
		) m ON c.id = m.conversation_id
		ORDER BY COALESCE(m.created_at, c.created_at) DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		err := rows.Scan(&conv.ID, &conv.TelegramChatID, &conv.TelegramUsername, &conv.TelegramFirstName,
			&conv.IsBotActive, &conv.CreatedAt, &conv.UpdatedAt, &conv.LastMessage, &conv.LastMessageTime)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// GetMessages returns all messages for a conversation
func GetMessages(conversationID int) ([]Message, error) {
	rows, err := DB.Query(`
		SELECT id, conversation_id, sender_type, message_text, created_at
		FROM messages
		WHERE conversation_id = ?
		ORDER BY created_at ASC
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.SenderType, &msg.MessageText, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// SetBotActive sets the is_bot_active flag for a conversation
func SetBotActive(conversationID int, active bool) error {
	_, err := DB.Exec(`
		UPDATE conversations SET is_bot_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, active, conversationID)
	return err
}

// GetKnowledgeBase returns the current knowledge base content
func GetKnowledgeBase() (string, error) {
	var content string
	err := DB.QueryRow("SELECT content FROM knowledge_base ORDER BY id DESC LIMIT 1").Scan(&content)
	return content, err
}

// UpdateKnowledgeBase updates the knowledge base content
func UpdateKnowledgeBase(content string) error {
	_, err := DB.Exec(`
		UPDATE knowledge_base SET content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = (SELECT id FROM knowledge_base ORDER BY id DESC LIMIT 1)
	`)
	if err == nil {
		_, err = DB.Exec("INSERT INTO knowledge_base (content) VALUES (?)", content)
	}
	return err
}

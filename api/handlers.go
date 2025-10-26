package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"telecust/bot"
	"telecust/database"

	"github.com/go-chi/chi/v5"
)

// GetConversations returns all conversations
func GetConversations(w http.ResponseWriter, r *http.Request) {
	conversations, err := database.GetAllConversations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

// GetConversationMessages returns all messages for a conversation
func GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	messages, err := database.GetMessages(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// TakeOverConversation disables bot for a conversation
func TakeOverConversation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	err = database.SetBotActive(id, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// ActivateBot enables bot for a conversation
func ActivateBot(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	err = database.SetBotActive(id, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// SendMessage sends a message from admin to user
func SendMessage(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Message string `json:"message"`
	}

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Message == "" {
		http.Error(w, "Message cannot be empty", http.StatusBadRequest)
		return
	}

	// Get conversation to find chat ID
	conversations, err := database.GetAllConversations()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var chatID int64
	found := false
	for _, conv := range conversations {
		if conv.ID == id {
			chatID = conv.TelegramChatID
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}

	// Send message via bot
	if bot.GlobalBot == nil {
		http.Error(w, "Bot not initialized", http.StatusInternalServerError)
		return
	}

	err = bot.GlobalBot.SendMessageAsAdmin(chatID, req.Message, id)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// GetKnowledgeBase returns the current knowledge base
func GetKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	content, err := database.GetKnowledgeBase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"content": content})
}

// UpdateKnowledgeBase updates the knowledge base
func UpdateKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Content string `json:"content"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = database.UpdateKnowledgeBase(req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

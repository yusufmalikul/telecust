package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"telecust/database"
)

type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// QueryKnowledgeBase uses OpenAI to answer user queries based on knowledge base and conversation history
func QueryKnowledgeBase(userQuery, knowledgeBase string, conversationID int) string {
	log.Printf("[AI] Received query: %s (conversation ID: %d)", userQuery, conversationID)

	// Check for simple greetings first
	greetings := []string{"halo", "hai", "hi", "hello", "hey", "selamat"}
	queryLower := strings.ToLower(userQuery)

	for _, greeting := range greetings {
		if strings.Contains(queryLower, greeting) && len(userQuery) < 20 {
			log.Printf("[AI] Detected greeting, returning instant response")
			return "Apa yang bisa saya bantu, kak?"
		}
	}

	// Use OpenAI for other queries
	apiKey := os.Getenv("OPENAI_API_KEY")
	apiBase := os.Getenv("OPENAI_API_BASE")

	if apiKey == "" {
		log.Printf("[AI] ERROR: OPENAI_API_KEY not configured")
		return "Maaf, sistem AI belum dikonfigurasi. Silakan hubungi admin."
	}

	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

	log.Printf("[AI] Using OpenAI API: %s", apiBase)

	// Build the system prompt with knowledge base
	systemPrompt := fmt.Sprintf(`Kamu adalah asisten customer service yang ramah dan membantu.
Jawab pertanyaan customer berdasarkan knowledge base berikut:

%s

Instruksi:
- Jawab dengan bahasa Indonesia yang sopan dan ramah
- Gunakan sapaan "kak" untuk customer
- PENTING: Perhatikan riwayat percakapan dengan baik. Jika customer bertanya tentang pesanan mereka sebelumnya, lihat di riwayat chat apa yang mereka pesan
- Jika pertanyaan tidak bisa dijawab dari knowledge base, beritahu dengan sopan bahwa kamu tidak memiliki informasi tersebut
- Jawab singkat dan jelas
- Jangan mengarang informasi yang tidak ada di knowledge base atau riwayat percakapan`, knowledgeBase)

	log.Printf("[AI] Knowledge base length: %d characters", len(knowledgeBase))

	// Get history limit from env or use default
	historyLimit := 10 // Default: last 10 messages (5 back-and-forth)
	if envLimit := os.Getenv("CONVERSATION_HISTORY_LIMIT"); envLimit != "" {
		if limit, err := strconv.Atoi(envLimit); err == nil && limit > 0 {
			historyLimit = limit
		}
	}

	// Get recent conversation history (we'll filter out the current one)
	log.Printf("[AI] Loading conversation history (limit: %d)...", historyLimit)
	history, err := database.GetRecentMessages(conversationID, historyLimit)
	if err != nil {
		log.Printf("[AI] Warning: Could not load conversation history: %v", err)
		history = []database.Message{}
	}
	log.Printf("[AI] Loaded %d messages from database", len(history))

	// Convert history to OpenAI message format, excluding the current message
	// Check if the last message is the current one (which we just saved)
	skipLastMessage := false
	if len(history) > 0 {
		lastMsg := history[len(history)-1]
		if lastMsg.MessageText == userQuery && lastMsg.SenderType == "user" {
			skipLastMessage = true
			log.Printf("[AI] Detected current message in history, will exclude it")
		}
	}

	var conversationHistory []Message
	for i, msg := range history {
		// Skip the last message if it's the current one
		if skipLastMessage && i == len(history)-1 {
			log.Printf("[AI] Skipping current message from history (last position)")
			continue
		}

		role := "user"
		if msg.SenderType == "bot" || msg.SenderType == "assistant" {
			role = "assistant"
		} else if msg.SenderType == "admin" {
			// Admin messages are also from assistant perspective
			role = "assistant"
		}

		conversationHistory = append(conversationHistory, Message{
			Role:    role,
			Content: msg.MessageText,
		})

		log.Printf("[AI] History[%d]: %s said: %s", i, role, msg.MessageText)
	}

	log.Printf("[AI] Current user query: %s", userQuery)
	log.Printf("[AI] Calling OpenAI API with %d history messages...", len(conversationHistory))

	// Call OpenAI API with conversation history
	response, err := callOpenAI(apiBase, apiKey, systemPrompt, userQuery, conversationHistory)
	if err != nil {
		log.Printf("[AI] ERROR: OpenAI API failed: %v", err)
		// Fallback to simple response
		return "Maaf, saya sedang mengalami kendala. Bisa ulangi pertanyaannya?"
	}

	log.Printf("[AI] SUCCESS: Received response from OpenAI (length: %d chars)", len(response))
	log.Printf("[AI] Response: %s", response)

	return response
}

func callOpenAI(apiBase, apiKey, systemPrompt, userMessage string, conversationHistory []Message) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(apiBase, "/"))
	log.Printf("[OpenAI] POST %s", url)

	// Get model from env, default to gpt-3.5-turbo
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	log.Printf("[OpenAI] Using model: %s", model)

	// Build messages array: system prompt + conversation history + current user message
	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Add conversation history
	messages = append(messages, conversationHistory...)

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	log.Printf("[OpenAI] Total messages in request: %d (1 system + %d history + 1 current)", len(messages), len(conversationHistory))

	requestBody := OpenAIRequest{
		Model:    model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("[OpenAI] Failed to marshal request: %v", err)
		return "", err
	}

	log.Printf("[OpenAI] Request body size: %d bytes", len(jsonData))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("[OpenAI] Failed to create request: %v", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	log.Printf("[OpenAI] Using API key: %s", maskKey(apiKey))

	client := &http.Client{}
	log.Printf("[OpenAI] Sending request...")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[OpenAI] HTTP request failed: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	log.Printf("[OpenAI] Response status: %d %s", resp.StatusCode, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[OpenAI] Failed to read response body: %v", err)
		return "", err
	}

	log.Printf("[OpenAI] Response body size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("[OpenAI] API returned error: %s", string(body))
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		log.Printf("[OpenAI] Failed to parse response JSON: %v", err)
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		log.Printf("[OpenAI] No choices in response")
		return "", fmt.Errorf("no response from OpenAI")
	}

	log.Printf("[OpenAI] Successfully parsed response, choices: %d", len(openAIResp.Choices))
	return openAIResp.Choices[0].Message.Content, nil
}

// maskKey masks the API key for logging (shows only first and last 4 chars)
func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
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

// QueryKnowledgeBase uses OpenAI to answer user queries based on knowledge base
func QueryKnowledgeBase(userQuery, knowledgeBase string) string {
	log.Printf("[AI] Received query: %s", userQuery)

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
- Jika pertanyaan tidak bisa dijawab dari knowledge base, beritahu dengan sopan bahwa kamu tidak memiliki informasi tersebut
- Jawab singkat dan jelas
- Jangan mengarang informasi yang tidak ada di knowledge base`, knowledgeBase)

	log.Printf("[AI] Knowledge base length: %d characters", len(knowledgeBase))
	log.Printf("[AI] Calling OpenAI API...")

	// Call OpenAI API
	response, err := callOpenAI(apiBase, apiKey, systemPrompt, userQuery)
	if err != nil {
		log.Printf("[AI] ERROR: OpenAI API failed: %v", err)
		// Fallback to simple response
		return "Maaf, saya sedang mengalami kendala. Bisa ulangi pertanyaannya?"
	}

	log.Printf("[AI] SUCCESS: Received response from OpenAI (length: %d chars)", len(response))
	log.Printf("[AI] Response: %s", response)

	return response
}

func callOpenAI(apiBase, apiKey, systemPrompt, userMessage string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(apiBase, "/"))
	log.Printf("[OpenAI] POST %s", url)

	// Get model from env, default to gpt-3.5-turbo
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	log.Printf("[OpenAI] Using model: %s", model)

	requestBody := OpenAIRequest{
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
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

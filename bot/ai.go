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
	// Check for simple greetings first
	greetings := []string{"halo", "hai", "hi", "hello", "hey", "selamat"}
	queryLower := strings.ToLower(userQuery)

	for _, greeting := range greetings {
		if strings.Contains(queryLower, greeting) && len(userQuery) < 20 {
			return "Apa yang bisa saya bantu, kak?"
		}
	}

	// Use OpenAI for other queries
	apiKey := os.Getenv("OPENAI_API_KEY")
	apiBase := os.Getenv("OPENAI_API_BASE")

	if apiKey == "" {
		return "Maaf, sistem AI belum dikonfigurasi. Silakan hubungi admin."
	}

	if apiBase == "" {
		apiBase = "https://api.openai.com/v1"
	}

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

	// Call OpenAI API
	response, err := callOpenAI(apiBase, apiKey, systemPrompt, userQuery)
	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		// Fallback to simple response
		return "Maaf, saya sedang mengalami kendala. Bisa ulangi pertanyaannya?"
	}

	return response
}

func callOpenAI(apiBase, apiKey, systemPrompt, userMessage string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", strings.TrimSuffix(apiBase, "/"))

	requestBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
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
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var openAIResp OpenAIResponse
	err = json.Unmarshal(body, &openAIResp)
	if err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

# Implementation Details

## Architecture Overview
Simple monolithic application with:
- Telegram bot handler (polling)
- HTTP API server for admin dashboard
- SQLite database for persistence
- OpenAI-powered AI for intelligent responses

## Tech Stack
- **Backend**: Go 1.21+
- **Database**: SQLite3
- **Telegram**: telegram-bot-api library
- **AI**: OpenAI API (GPT-3.5-turbo)
- **Frontend**: Plain HTML + CSS + Vanilla JS
- **HTTP Router**: Chi (lightweight)
- **Config**: godotenv for .env file support

## Database Schema

### Tables

**conversations**
- id (INTEGER PRIMARY KEY)
- telegram_chat_id (INTEGER UNIQUE)
- telegram_username (TEXT)
- telegram_first_name (TEXT)
- is_bot_active (BOOLEAN) - true if bot handling, false if admin took over
- created_at (DATETIME)
- updated_at (DATETIME)

**messages**
- id (INTEGER PRIMARY KEY)
- conversation_id (INTEGER FK)
- sender_type (TEXT) - 'user' or 'bot' or 'admin'
- message_text (TEXT)
- created_at (DATETIME)

**knowledge_base**
- id (INTEGER PRIMARY KEY)
- content (TEXT) - the knowledge base content
- updated_at (DATETIME)

## Project Structure
```
telecust/
├── main.go                 # Entry point, starts bot & web server
├── go.mod
├── go.sum
├── database/
│   ├── db.go              # SQLite connection & setup
│   └── models.go          # Database models & queries
├── bot/
│   ├── handler.go         # Telegram message handler
│   └── ai.go              # OpenAI integration
├── api/
│   ├── server.go          # HTTP server setup
│   ├── handlers.go        # API endpoints
│   └── middleware.go      # CORS, logging, etc
├── web/
│   ├── index.html         # Admin dashboard
│   ├── style.css          # Blue & white theme
│   └── app.js             # Frontend logic
└── telecust.db            # SQLite database (auto-created)
```

## AI/Knowledge Base Strategy
Uses OpenAI GPT-3.5-turbo for intelligent responses:
1. Simple greetings (halo, hai, etc.) get instant response without API call
2. Other queries are sent to OpenAI with knowledge base as context
3. System prompt instructs AI to:
   - Answer in polite Indonesian
   - Use "kak" to address customers
   - Only use information from knowledge base
   - Admit when information is not available
   - Keep responses short and clear
4. Fallback to error message if API fails

Supports custom OpenAI API base URL for alternative providers.

## API Endpoints

**GET /api/conversations**
- Returns all conversations with latest message

**GET /api/conversations/:id/messages**
- Returns all messages for a conversation

**POST /api/conversations/:id/takeover**
- Sets is_bot_active = false

**POST /api/conversations/:id/activate-bot**
- Sets is_bot_active = true

**POST /api/conversations/:id/send**
- Admin sends message to user via bot
- Body: { "message": "text" }

**GET /api/knowledge-base**
- Returns current knowledge base content

**PUT /api/knowledge-base**
- Updates knowledge base
- Body: { "content": "text" }

## Bot Flow
1. Receive message from Telegram
2. Check if conversation exists, create if not
3. Check is_bot_active flag
4. If bot active: query AI, send response, save both messages
5. If admin mode: just save user message, notify admin
6. Handle commands: /start, /help

## Frontend Features
- Real-time-ish updates (polling every 3 seconds)
- Chat list on left, conversation on right
- Blue (#0088cc) and white theme (Telegram colors)
- Take Over / Activate Bot toggle button
- Knowledge base editor in settings modal
- Simple, clean UI - no framework needed

## Environment Variables
- `TELEGRAM_BOT_TOKEN` or `TELE_BOT_TOKEN` - Bot token from @BotFather
- `OPENAI_API_KEY` - OpenAI API key (required for AI responses)
- `OPENAI_API_BASE` - Custom OpenAI API base URL (optional, defaults to https://api.openai.com/v1)
- `OPENAI_MODEL` - OpenAI model to use (optional, defaults to gpt-3.5-turbo, can use gpt-4 for better accuracy)
- `CONVERSATION_HISTORY_LIMIT` - Number of recent messages to include in conversation context (optional, default: 10)
- `PORT` - HTTP server port (default: 8080)

## Run Instructions
1. Create `.env` file with required variables (see .env example)
2. Run: `go run main.go`
3. Access dashboard: http://localhost:8080

## Deployment Notes
- Single binary, easy to deploy
- SQLite file persists data
- Can run on any VPS with Go installed
- Requires OpenAI API key (or compatible API)
- .env file or environment variables for configuration

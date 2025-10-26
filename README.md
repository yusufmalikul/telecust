# Telecust - Telegram Customer Service Bot

A simple AI-powered customer service bot for Telegram with an admin dashboard.

## Features

- Telegram bot that answers customer questions using a knowledge base
- OpenAI-powered intelligent responses (GPT-3.5-turbo)
- Admin dashboard to view all conversations
- Take over feature to stop bot and reply manually
- Knowledge base editor
- Clean UI with Telegram-style blue and white theme
- Supports custom OpenAI API endpoints

## Tech Stack

- **Backend**: Go 1.21+
- **Database**: SQLite3
- **AI**: OpenAI API (GPT-3.5-turbo)
- **Frontend**: HTML, CSS, Vanilla JavaScript
- **Telegram Bot**: telegram-bot-api

## Prerequisites

1. Go 1.21 or higher
2. A Telegram Bot Token (get it from [@BotFather](https://t.me/botfather))
3. An OpenAI API key (or compatible API service)

## Setup

### 1. Get Telegram Bot Token

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Send `/newbot` command
3. Follow the instructions to create your bot
4. Copy the bot token

### 2. Get OpenAI API Key

Get your API key from [OpenAI](https://platform.openai.com/api-keys) or use a compatible service.

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Configure Environment Variables

Create a `.env` file in the project root:

```env
TELE_BOT_TOKEN=your-telegram-bot-token-here
OPENAI_API_KEY=your-openai-api-key-here
OPENAI_API_BASE=https://api.openai.com/v1
PORT=8080
```

**Environment Variables:**
- `TELE_BOT_TOKEN` or `TELEGRAM_BOT_TOKEN` - Your Telegram bot token from @BotFather
- `OPENAI_API_KEY` - Your OpenAI API key (required)
- `OPENAI_API_BASE` - OpenAI API endpoint (optional, defaults to https://api.openai.com/v1)
- `PORT` - HTTP server port (optional, defaults to 8080)

**Note:** You can also use alternative OpenAI-compatible services by changing `OPENAI_API_BASE`.

### 5. Run the Application

```bash
go run main.go
```

The application will:
- Initialize the SQLite database (`telecust.db`)
- Start the Telegram bot
- Start the web server on port 8080

### 6. Access the Dashboard

Open your browser and go to:
```
http://localhost:8080
```

## Usage

### Customer Side (Telegram)

1. Users can start a chat with your bot
2. They can ask questions like "Halo kak" or "Harga keripik kentang?"
3. The bot will respond based on the knowledge base

### Admin Side (Dashboard)

1. Open the dashboard at `http://localhost:8080`
2. View all customer conversations in the left panel
3. Click on a conversation to see messages
4. Use "Take Over" button to disable the bot and reply manually
5. Use "Activate Bot" to re-enable automatic responses
6. Click "Knowledge Base Settings" to edit the knowledge base

## Knowledge Base

The bot uses OpenAI GPT-3.5-turbo to provide intelligent responses based on your knowledge base:

**How it works:**
- Simple greetings (halo, hai, hello) get instant responses without API calls
- Other queries are sent to OpenAI with your knowledge base as context
- The AI is instructed to:
  - Answer in polite Indonesian
  - Use "kak" to address customers
  - Only provide information from the knowledge base
  - Admit when information is not available
  - Keep responses short and clear

**Default knowledge base** (Indonesian example for potato chips):
```
Harga kentang Rp5ribu perbungkus.
Pesan diatas 10 harga 4rb.
Jika pesan 10 Rp40ribu.
Jika pesan 20 Rp80ribu.
Jika pesan di atas 100 bungkus harga Rp3ribu.
```

You can edit the knowledge base through the dashboard settings. The AI will use this information to answer customer questions intelligently.

## Project Structure

```
telecust/
├── main.go                 # Entry point
├── database/
│   ├── db.go              # Database operations
│   └── models.go          # Data models
├── bot/
│   ├── handler.go         # Telegram message handler
│   └── ai.go              # Keyword matching AI
├── api/
│   ├── server.go          # HTTP server
│   └── handlers.go        # API endpoints
├── web/
│   ├── index.html         # Admin dashboard
│   ├── style.css          # Styles
│   └── app.js             # Frontend logic
└── telecust.db            # SQLite database
```

## API Endpoints

- `GET /api/conversations` - Get all conversations
- `GET /api/conversations/:id/messages` - Get messages for a conversation
- `POST /api/conversations/:id/takeover` - Disable bot for conversation
- `POST /api/conversations/:id/activate-bot` - Re-enable bot
- `POST /api/conversations/:id/send` - Send message as admin
- `GET /api/knowledge-base` - Get knowledge base content
- `PUT /api/knowledge-base` - Update knowledge base

## Configuration

**Using .env file (recommended):**
```env
TELE_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
OPENAI_API_KEY=sk-your-api-key-here
OPENAI_API_BASE=https://api.openai.com/v1
PORT=8080
```

**Using environment variables:**
```bash
export TELEGRAM_BOT_TOKEN="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
export OPENAI_API_KEY="sk-your-api-key-here"
export PORT="3000"
go run main.go
```

**Available variables:**
- `TELEGRAM_BOT_TOKEN` or `TELE_BOT_TOKEN` - Your Telegram bot token (required)
- `OPENAI_API_KEY` - Your OpenAI API key (required)
- `OPENAI_API_BASE` - OpenAI API endpoint (optional, defaults to https://api.openai.com/v1)
- `PORT` - HTTP server port (optional, default: 8080)

## Deployment

### Simple VPS Deployment

1. Copy files to your server
2. Create `.env` file with your configuration
3. Run: `go build && ./telecust`

### Using systemd (Linux)

Create `/etc/systemd/system/telecust.service`:
```ini
[Unit]
Description=Telecust Bot
After=network.target

[Service]
Type=simple
User=youruser
WorkingDirectory=/path/to/telecust
Environment="TELEGRAM_BOT_TOKEN=your-token"
Environment="OPENAI_API_KEY=your-api-key"
Environment="OPENAI_API_BASE=https://api.openai.com/v1"
ExecStart=/path/to/telecust/telecust
Restart=always

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl daemon-reload
sudo systemctl enable telecust
sudo systemctl start telecust
```

## Development

To run in development mode with auto-reload, you can use tools like:
- `air` - Live reload for Go apps
- `nodemon` - Monitor for changes

Example with air:
```bash
go install github.com/cosmtrek/air@latest
air
```

## Troubleshooting

**Bot not responding:**
- Check if `TELEGRAM_BOT_TOKEN` is set correctly
- Verify the bot is running: check console logs
- Test the bot token using Telegram's Bot API

**Dashboard not loading:**
- Check if port 8080 is available
- Verify the server is running
- Check browser console for errors

**Messages not saving:**
- Check database file permissions
- Look for errors in console logs
- Verify SQLite is working: `sqlite3 telecust.db ".tables"`

## License

MIT

## Contributing

Feel free to submit issues and pull requests!

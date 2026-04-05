# Pocky Ops Bot

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A modular Telegram bot for personal workflows — AI-powered chat, Binance portfolio tracking (spot + futures), and an extensible tool-calling framework. Built in Go with clean architecture.

## Features

- **AI Chat** — Multi-provider LLM support (Google Gemini, Anthropic Claude, OpenAI, Qwen) with per-chat conversation history
- **Binance Portfolio** — Real-time spot balances + futures positions, orders, and P&L via `/dautu`
- **Tool Calling** — AI automatically invokes registered tools to fetch live data
- **Long-Polling** — Reliable update retrieval with exponential backoff and automatic retry
- **Per-Chat Isolation** — Each conversation runs in its own goroutine with local history; no shared state
- **Vietnamese Support** — Configurable to respond in Vietnamese (`AI_VIETNAMESE=true`)
- **Structured Logging** — `log/slog` throughout with configurable log level
- **Graceful Shutdown** — Context cancellation + WaitGroup for clean exit

## Quick Start

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- A Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
- An AI provider API key (Gemini, Claude, OpenAI, or Qwen)
- *(Optional)* Binance API key + secret for portfolio tracking

### Setup

```bash
git clone <repo>
cd pocky-ops-bot
go mod download
cp .env.example .env
# Edit .env — set TELEGRAM_BOT_TOKEN, AI_API_KEY, and optionally BINANCE_API_KEY
go run ./cmd/bot
```

## Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and overview |
| `/dautu` | Binance portfolio summary (spot + futures) |
| `/xoa` | Clear current conversation history |
| `/trogiup` | Full help and usage guide |

Any other text is sent to the AI as a chat message, with full conversation context.

## Configuration

All settings via environment variables or `.env` file. Environment variables take precedence.

### Telegram

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token from @BotFather |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `POLL_INTERVAL` | `1s` | Minimum time between polls |
| `TIMEOUT` | `30s` | Long-polling timeout (max 50s) |

### AI

| Variable | Default | Description |
|----------|---------|-------------|
| `AI_PROVIDER` | `gemini` | `gemini` / `claude` / `openai` / `qwen` |
| `AI_API_KEY` | (required) | API key for chosen provider |
| `AI_MODEL` | `gemini-2.0-flash` | Model name |
| `AI_BASE_URL` | — | Override default API endpoint |
| `AI_MAX_TOKENS` | `1024` | Max tokens per response |
| `AI_TIMEOUT` | `60s` | AI request timeout |
| `AI_SYSTEM_PROMPT` | `You are Pocky...` | System prompt |
| `AI_VIETNAMESE` | `true` | Force Vietnamese responses |

### Conversation

| Variable | Default | Description |
|----------|---------|-------------|
| `CONVERSATION_MAX_TURNS` | `20` | Max message pairs kept in history |
| `CONVERSATION_TTL` | `30m` | Idle timeout before history is reset |

### Binance *(optional — tools disabled if not set)*

| Variable | Default | Description |
|----------|---------|-------------|
| `BINANCE_API_KEY` | — | Binance API key |
| `BINANCE_SECRET_KEY` | — | Binance secret (HMAC-SHA256 signing) |
| `BINANCE_BASE_URL` | — | Override spot API URL (testnet) |
| `BINANCE_FUTURES_BASE_URL` | — | Override futures API URL (testnet) |

## Project Structure

```
pocky-ops-bot/
├── cmd/bot/                        # Application entry point
├── internal/
│   ├── bot/                        # Dispatcher, Router, Command handlers
│   │   ├── dispatcher.go           # Per-chat goroutine routing + history
│   │   ├── router.go               # Command routing
│   │   └── handlers/command.go     # /start, /trogiup
│   ├── clients/
│   │   ├── telegram/               # Poller, Sender, backoff
│   │   ├── llm/                    # Multi-provider LLM client
│   │   └── binance/                # Spot + Futures REST client
│   ├── services/chat.go            # Stateless AI chat with tool loop
│   ├── tools/                      # Tool registry + executor interface
│   │   └── binance/                # 8 Binance tools (3 spot, 5 futures)
│   └── config/config.go            # Configuration loading
├── docs/ARCHITECTURE.md            # Detailed architecture docs
└── .env.example                    # Environment variable template
```

## Development

```bash
go test ./...                   # run all tests
go test ./... -cover            # with coverage
go test ./... -race             # race detector
go build -o bot ./cmd/bot       # build binary
gofmt -w .                      # format code
golangci-lint run               # lint
```

## Architecture

The bot uses a **per-chat goroutine** model: each conversation runs in its own goroutine owning its history as local state. No locks or shared mutable state between chats.

```
Telegram → Poller → Dispatcher → per-chat worker goroutine
                                       ↓
                                 ChatService (stateless)
                                       ↓
                              LLM client + Tool registry
                                       ↓
                              Binance API (if configured)
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for full details: component reference, data flows, interface boundaries, and design patterns.

## Creating a Telegram Bot

1. Open Telegram and message [@BotFather](https://t.me/BotFather)
2. Send `/newbot` and follow prompts (name + username ending in `bot`)
3. Copy the token and set `TELEGRAM_BOT_TOKEN` in `.env`

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | v1.5.1 | Load `.env` files |

All other functionality uses the Go standard library.

## Security

- Never commit tokens or API keys — use environment variables only
- Validate all user inputs (commands and callbacks)
- Parameterized queries for any database access
- Apply least-privilege for bot permissions and integrations

## Roadmap

- [ ] Webhook mode support
- [ ] Middleware chain (auth, rate limiting, logging)
- [ ] Repository layer (SQLite/PostgreSQL) for persistent history
- [ ] More tool integrations

## License

MIT License — see [LICENSE](LICENSE) for details.

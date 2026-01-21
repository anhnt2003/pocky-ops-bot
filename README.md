# Pocky Ops Bot

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A modular Telegram bot for personal workflows, featuring commands for automation, notes, and lightweight integrations. Built with Go using clean architecture principles.

## Features

- 🚀 **Long-Polling Support** - Efficient update retrieval with configurable timeouts
- 🔄 **Automatic Retry Logic** - Exponential backoff for handling transient failures
- ⚙️ **Flexible Configuration** - Environment-based settings with sensible defaults
- 📝 **Structured Logging** - Built-in `log/slog` integration for debugging
- 🧩 **Modular Architecture** - Clean separation of concerns (MVC-style)
- ✅ **Well-Tested** - Comprehensive unit tests with mock HTTP clients

## Quick Start

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- A Telegram Bot Token (see [Creating a Telegram Bot](#creating-a-telegram-bot) below)

### Installation

1. Clone the repository
2. Run `go mod download` to install dependencies

### Configuration

1. Copy `.env.example` to `.env`
2. Set your `TELEGRAM_BOT_TOKEN` in the `.env` file

### Running the Bot

Run `go run ./cmd/bot` to start the bot.

## Creating a Telegram Bot

To use this project, you need to create a Telegram bot and obtain an API token. Follow these steps:

### Step 1: Start a Chat with BotFather

1. Open Telegram and search for **@BotFather** or click this link: [https://t.me/BotFather](https://t.me/BotFather)
2. Start a conversation by clicking **Start** or sending `/start`

### Step 2: Create a New Bot

1. Send the command `/newbot` to BotFather
2. BotFather will ask you for a **name** for your bot (this is the display name, e.g., "My Awesome Bot")
3. Next, provide a **username** for your bot. It must:
   - End with `bot` (e.g., `my_awesome_bot` or `MyAwesomeBot`)
   - Be unique across all Telegram bots
   - Contain only letters, numbers, and underscores

### Step 3: Get Your Bot Token

1. After successfully creating your bot, BotFather will send you a message containing your **HTTP API token**
2. The token looks like: `123456789:ABCdefGHIjklMNOpqrsTUVwxyz`
3. **Keep this token secret!** Anyone with this token can control your bot

### Step 4: Configure Your Bot (Optional)

You can customize your bot by sending these commands to BotFather:

| Command | Description |
|---------|-------------|
| `/setdescription` | Set the bot's description (shown in the bot's profile) |
| `/setabouttext` | Set the "About" text (shown when users open the bot) |
| `/setuserpic` | Upload a profile picture for your bot |
| `/setcommands` | Define the command list (shown in the menu) |
| `/setprivacy` | Set whether the bot can see all messages in groups |

### Step 5: Add Token to Your Project

1. Copy your token from the BotFather message
2. Add it to your `.env` file:
   ```
   TELEGRAM_BOT_TOKEN=123456789:ABCdefGHIjklMNOpqrsTUVwxyz
   ```

## Project Structure

```
pocky-ops-bot/
├── cmd/bot/                     # Application entry point
├── internal/
│   ├── bot/types/               # Telegram Bot API type definitions
│   ├── clients/telegram/        # Telegram API client & polling
│   └── config/                  # Configuration loading
├── docs/                        # Architecture documentation
├── .env.example                 # Environment variable template
├── AGENTS.md                    # AI agent & code guidelines
└── CHANGELOG.md                 # Version changelog
```

## Configuration Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token from BotFather |
| `LOG_LEVEL` | `info` | Logging level (`debug`, `info`, `warn`, `error`) |
| `POLL_INTERVAL` | `1s` | Minimum time between polling requests |
| `TIMEOUT` | `30s` | Long-polling timeout (max 50s per Telegram API) |

## Development

### Commands

| Command | Description |
|---------|-------------|
| `go test ./...` | Run all tests |
| `go test ./... -cover` | Run tests with coverage |
| `gofmt -w .` | Format code |
| `golangci-lint run` | Run linter |
| `go build -o bot ./cmd/bot` | Build binary |

## Architecture

The bot follows MVC-style layering adapted for Go:

| Layer | Description |
|-------|-------------|
| **Clients** | Telegram API client, polling, backoff strategies |
| **Bot Types** | Comprehensive Telegram Bot API type definitions |
| **Config** | Environment-based configuration loading |

For detailed architecture documentation, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| [github.com/joho/godotenv](https://github.com/joho/godotenv) | v1.5.1 | Load environment from `.env` files |

## Contributing

1. Fork the repository
2. Create a feature branch
3. Follow the [AGENTS.md](AGENTS.md) guidelines for code style
4. Write tests for new functionality (maintain >80% coverage on services)
5. Open a Pull Request

## Security

- **Never commit tokens or API keys** - use environment variables
- **Validate all user inputs** - especially commands and callbacks
- **Use parameterized queries** - for any database access
- **Apply least-privilege** - for bot permissions and integrations

## Roadmap

- [ ] Command router with prefix matching
- [ ] Middleware support (auth, logging, rate limiting)
- [ ] Conversation state management
- [ ] Service layer for business logic
- [ ] Repository layer with SQLite/PostgreSQL support
- [ ] Webhook mode support

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Telegram Bot API](https://core.telegram.org/bots/api) - Official API documentation
- [BotFather](https://t.me/BotFather) - Official bot creation tool
- [godotenv](https://github.com/joho/godotenv) - Environment variable loading

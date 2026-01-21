# Pocky Ops Bot - Architecture Documentation

## Overview

**Pocky Ops Bot** is a modular Telegram bot built in Go for personal workflows, featuring commands for automation, notes, and lightweight integrations. The architecture follows MVC-style layering adapted for Go with clean separation of concerns.

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture Diagram](#architecture-diagram)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Package Details](#package-details)
- [Configuration](#configuration)
- [Design Patterns](#design-patterns)
- [Testing Strategy](#testing-strategy)
- [Future Considerations](#future-considerations)

---

## Project Structure

```
pocky-ops-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Application entry point
├── internal/
│   ├── bot/
│   │   ├── router.go            # Update routing (placeholder)
│   │   ├── handlers/
│   │   │   └── connect.go       # Handler implementations (placeholder)
│   │   └── types/
│   │       ├── chat.go          # Chat-related types
│   │       ├── common.go        # Common utility types
│   │       ├── forum.go         # Forum/topic types
│   │       ├── inline.go        # Inline query & keyboard types
│   │       ├── media.go         # Media file types
│   │       ├── member.go        # Chat member types
│   │       ├── message.go       # Message types
│   │       ├── messaging.go     # Messaging utility types
│   │       ├── payment.go       # Payment/invoice types
│   │       ├── update.go        # Update types
│   │       └── user.go          # User types
│   ├── clients/
│   │   └── telegram/
│   │       ├── api.go           # API response types & interfaces
│   │       ├── backoff.go       # Exponential backoff strategy
│   │       ├── poller.go        # Long-polling implementation
│   │       ├── poller_test.go   # Poller unit tests
│   │       └── update_types.go  # Update type constants & helpers
│   └── config/
│       └── config.go            # Configuration loading
├── docs/
│   └── ARCHITECTURE.md          # This file
├── .env.example                 # Environment variable template
├── .gitignore                   # Git ignore rules
├── AGENTS.md                    # AI agent guidelines
├── CHANGELOG.md                 # Version changelog
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── README.md                    # Project readme
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              TELEGRAM API                                │
│                     https://api.telegram.org/bot<TOKEN>                 │
└───────────────────────────────────┬─────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           CLIENTS LAYER                                  │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                  internal/clients/telegram                        │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐ │   │
│  │  │   Poller     │  │  APIResponse │  │  ExponentialBackoff    │ │   │
│  │  │ - Start()    │  │  - OK        │  │  - NextBackoff()       │ │   │
│  │  │ - Stop()     │  │  - Result    │  │  - Reset()             │ │   │
│  │  │ - Updates()  │  │  - ErrorCode │  └────────────────────────┘ │   │
│  │  │ - GetMe()    │  └──────────────┘                             │   │
│  │  └──────────────┘                                               │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└───────────────────────────────────┬─────────────────────────────────────┘
                                    │ types.Update
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                             BOT LAYER                                    │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    internal/bot/types                              │ │
│  │  ┌────────┐ ┌─────────┐ ┌──────┐ ┌──────────────┐ ┌──────────┐   │ │
│  │  │ Update │ │ Message │ │ Chat │ │ CallbackQuery│ │   User   │   │ │
│  │  └────────┘ └─────────┘ └──────┘ └──────────────┘ └──────────┘   │ │
│  │  ┌───────────┐ ┌───────┐ ┌─────────┐ ┌──────────┐ ┌──────────┐   │ │
│  │  │InlineQuery│ │ Media │ │ Payment │ │  Member  │ │  Forum   │   │ │
│  │  └───────────┘ └───────┘ └─────────┘ └──────────┘ └──────────┘   │ │
│  └────────────────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    internal/bot/handlers (TBD)                     │ │
│  │           Router → Command Handlers → Service Layer                │ │
│  └────────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────┬─────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                          CONFIGURATION                                   │
│  ┌────────────────────────────────────────────────────────────────────┐ │
│  │                    internal/config                                 │ │
│  │  Config { TelegramToken, LogLevel, PollInterval, Timeout, ... }   │ │
│  │  - Load() reads from .env + environment variables                 │ │
│  └────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Core Components

### 1. Entry Point ([`cmd/bot/main.go`](../cmd/bot/main.go))

The application entry point handles:
- Configuration loading via [`config.Load()`](../internal/config/config.go:31)
- Logger setup with structured logging (`log/slog`)
- Telegram poller initialization via [`telegram.NewPollerWithOptions()`](../internal/clients/telegram/poller.go:485)
- Bot connection verification via [`poller.GetMe()`](../internal/clients/telegram/poller.go:384)
- Graceful shutdown handling (SIGINT/SIGTERM)
- Update handler registration

### 2. Telegram Client ([`internal/clients/telegram/`](../internal/clients/telegram/))

#### Poller ([`poller.go`](../internal/clients/telegram/poller.go))

The core polling mechanism for receiving updates:

| Component | Description |
|-----------|-------------|
| [`Poller`](../internal/clients/telegram/poller.go:63) | Main struct managing long-polling lifecycle |
| [`PollerConfig`](../internal/clients/telegram/poller.go:21) | Configuration options (token, timeout, retries, etc.) |
| [`UpdateHandler`](../internal/clients/telegram/poller.go:498) | Function type for processing updates |

**Key Methods:**
- [`Start(ctx)`](../internal/clients/telegram/poller.go:140) - Begins polling in a goroutine
- [`Stop()`](../internal/clients/telegram/poller.go:162) - Gracefully shuts down polling
- [`Updates()`](../internal/clients/telegram/poller.go:134) - Returns read-only update channel
- [`StartWithHandler(ctx, handler)`](../internal/clients/telegram/poller.go:502) - Convenience method with handler

**Functional Options:**
- [`WithTimeout(d)`](../internal/clients/telegram/poller.go:436)
- [`WithPollInterval(d)`](../internal/clients/telegram/poller.go:429)
- [`WithMaxRetries(n)`](../internal/clients/telegram/poller.go:443)
- [`WithBackoff(b)`](../internal/clients/telegram/poller.go:450)
- [`WithLogger(logger)`](../internal/clients/telegram/poller.go:471)
- [`WithHTTPClient(client)`](../internal/clients/telegram/poller.go:464)

#### API Types ([`api.go`](../internal/clients/telegram/api.go))

| Type | Purpose |
|------|---------|
| [`APIResponse`](../internal/clients/telegram/api.go:12) | Generic Telegram API response wrapper |
| [`APIError`](../internal/clients/telegram/api.go:27) | Error type with retry support |
| [`BackoffStrategy`](../internal/clients/telegram/api.go:48) | Interface for retry backoff calculation |
| [`HTTPClient`](../internal/clients/telegram/api.go:57) | Interface for HTTP operations (testable) |

#### Backoff Strategy ([`backoff.go`](../internal/clients/telegram/backoff.go))

Implements exponential backoff for retry logic:
- Initial interval: 1 second
- Maximum interval: 60 seconds
- Multiplier: 2.0

### 3. Bot Types ([`internal/bot/types/`](../internal/bot/types/))

Comprehensive Telegram Bot API type definitions organized by domain:

| File | Types | Description |
|------|-------|-------------|
| [`update.go`](../internal/bot/types/update.go) | `Update`, `WebhookInfo`, `BusinessConnection` | Incoming updates |
| [`message.go`](../internal/bot/types/message.go) | `Message` | Message with 100+ fields |
| [`user.go`](../internal/bot/types/user.go) | `User` | Telegram user info |
| [`chat.go`](../internal/bot/types/chat.go) | `Chat`, `ChatFullInfo`, `ChatPermissions` | Chat metadata |
| [`inline.go`](../internal/bot/types/inline.go) | `InlineQuery`, `CallbackQuery`, keyboard types | Inline features |
| [`media.go`](../internal/bot/types/media.go) | `Photo`, `Video`, `Audio`, `Document`, `Sticker` | Media files |
| [`member.go`](../internal/bot/types/member.go) | `ChatMember`, `ChatInviteLink`, `BotCommand` | Membership |
| [`payment.go`](../internal/bot/types/payment.go) | `Invoice`, `SuccessfulPayment`, `Gift` | Payments |
| [`forum.go`](../internal/bot/types/forum.go) | `ForumTopic`, `Giveaway`, `PassportData` | Forums |
| [`common.go`](../internal/bot/types/common.go) | `Location`, `Contact`, `Poll`, `Dice` | Utilities |
| [`messaging.go`](../internal/bot/types/messaging.go) | `MessageEntity`, `TextQuote`, `MessageOrigin` | Messaging |

### 4. Configuration ([`internal/config/config.go`](../internal/config/config.go))

```go
type Config struct {
    TelegramToken string        // Bot API token from BotFather
    LogLevel      string        // debug, info, warn, error
    PollInterval  time.Duration // Minimum time between polls
    Timeout       time.Duration // Long-polling timeout
    MaxRetries    int           // Max retry attempts
}
```

**Loading precedence:** Environment variables override `.env` file values.

---

## Data Flow

```
1. STARTUP
   main.go
     └─► config.Load()
           └─► .env file + ENV vars
     └─► telegram.NewPollerWithOptions(token, opts...)
           └─► Validate config, create channels
     └─► poller.GetMe(ctx)
           └─► HTTP GET /getMe → Verify bot token

2. POLLING LOOP
   poller.Start(ctx)
     └─► pollLoop goroutine
           ├─► getUpdates(ctx)
           │     └─► HTTP GET /getUpdates?timeout=30&offset=N
           │           └─► Parse APIResponse → []types.Update
           ├─► On success: Reset backoff, send to channel
           │     └─► updates chan <- update
           └─► On error: handleError()
                 ├─► Retryable? Apply backoff, retry
                 └─► Non-retryable? Stop poller

3. UPDATE HANDLING
   poller.StartWithHandler(ctx, handler)
     └─► goroutine: for update := range Updates()
           └─► handler(ctx, update)
                 ├─► update.Message → Handle text/command
                 ├─► update.CallbackQuery → Handle button clicks
                 └─► update.InlineQuery → Handle inline mode

4. SHUTDOWN
   Signal received (SIGINT/SIGTERM)
     └─► cancel() context
     └─► poller.Stop()
           └─► Close stopCh, wait for goroutine
           └─► Close updates channel
```

---

## Package Details

### Package Dependencies

```
cmd/bot/main
    ├── internal/config
    ├── internal/clients/telegram
    └── internal/bot/types

internal/clients/telegram
    ├── encoding/json
    ├── net/http
    ├── log/slog
    └── internal/bot/types

internal/config
    └── github.com/joho/godotenv

internal/bot/types
    └── (no dependencies - pure data types)
```

### Interface Boundaries

```go
// HTTPClient - allows mocking HTTP layer in tests
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// BackoffStrategy - pluggable retry strategies
type BackoffStrategy interface {
    NextBackoff(attempt int) time.Duration
    Reset()
}

// UpdateHandler - user-defined update processing
type UpdateHandler func(ctx context.Context, update types.Update) error
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token from @BotFather |
| `LOG_LEVEL` | `info` | Logging level (debug/info/warn/error) |
| `POLL_INTERVAL` | `1s` | Minimum time between polling requests |
| `TIMEOUT` | `30s` | Long-polling timeout (max 50s) |

### Example `.env`

```bash
TELEGRAM_BOT_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
LOG_LEVEL=debug
POLL_INTERVAL=1s
TIMEOUT=30s
```

---

## Design Patterns

### 1. Functional Options Pattern

Used for [`Poller`](../internal/clients/telegram/poller.go:485) configuration:

```go
poller, err := telegram.NewPollerWithOptions(
    token,
    telegram.WithTimeout(30*time.Second),
    telegram.WithMaxRetries(5),
    telegram.WithLogger(logger),
)
```

### 2. Dependency Injection

All dependencies passed via constructors:
- HTTPClient interface for testability
- Logger passed explicitly
- BackoffStrategy is pluggable

### 3. Graceful Shutdown

Context-based cancellation with WaitGroup synchronization:

```go
ctx, cancel := context.WithCancel(ctx)
// ...
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
<-sigCh
cancel()
poller.Stop()
```

### 4. Error Classification

Errors categorized for smart retry behavior:

```go
func (e *APIError) IsRetryable() bool {
    return e.RetryAfter > 0 || e.Code >= 500
}
```

---

## Testing Strategy

### Current Coverage

- **Poller tests** ([`poller_test.go`](../internal/clients/telegram/poller_test.go))
  - Configuration validation
  - Start/Stop lifecycle
  - Update receiving
  - Context cancellation
  - Error handling & backoff
  - Functional options

### Test Patterns

```go
// Mock HTTP client for isolation
type mockHTTPClient struct {
    responses []mockResponse
    requests  []*http.Request
}

// Table-driven tests for edge cases
tests := []struct {
    name    string
    config  PollerConfig
    wantErr bool
}{...}
```

### Running Tests

```bash
go test ./internal/clients/telegram/... -v
go test ./... -cover
```

---

## Future Considerations

### Planned Architecture Extensions

1. **Handler Layer** (`internal/bot/handlers/`)
   - Command router with prefix matching
   - Middleware support (auth, logging, rate limiting)
   - Conversation state management

2. **Service Layer** (`internal/services/`)
   - Business logic separation
   - Integration services (notes, automation)

3. **Repository Layer** (`internal/repository/`)
   - Data persistence abstraction
   - SQLite/PostgreSQL support

4. **Webhook Support**
   - HTTP server for webhook mode
   - SSL/TLS certificate handling

### Scalability Notes

- Current design: Single-instance, long-polling
- Future: Consider webhook mode for higher throughput
- Updates channel buffered (100) for burst handling
- Exponential backoff prevents API rate limiting

---

## Quick Reference

### Starting the Bot

```bash
# Set up environment
cp .env.example .env
# Edit .env with your bot token

# Run
go run ./cmd/bot
```

### Adding a New Update Handler

```go
handler := func(ctx context.Context, update types.Update) error {
    if update.Message != nil && update.Message.Text == "/start" {
        // Handle /start command
    }
    return nil
}

poller.StartWithHandler(ctx, handler)
```

### Creating Custom Backoff

```go
type ConstantBackoff struct {
    Interval time.Duration
}

func (b *ConstantBackoff) NextBackoff(attempt int) time.Duration {
    return b.Interval
}

func (b *ConstantBackoff) Reset() {}

// Usage
poller, _ := telegram.NewPollerWithOptions(
    token,
    telegram.WithBackoff(&ConstantBackoff{Interval: 5 * time.Second}),
)
```

---

*Last updated: 2026-01-21*

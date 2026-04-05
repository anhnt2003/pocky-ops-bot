# Pocky Ops Bot - Architecture Documentation

## Overview

**Pocky Ops Bot** is a modular Telegram bot built in Go for personal workflows, featuring AI-powered chat (multi-provider), Binance portfolio tracking (spot + futures), and a tool-calling framework. The architecture follows clean layered design with no shared mutable state between conversations.

## Table of Contents

- [Project Structure](#project-structure)
- [Architecture Diagram](#architecture-diagram)
- [Core Components](#core-components)
- [Data Flow](#data-flow)
- [Package Details](#package-details)
- [Configuration](#configuration)
- [Design Patterns](#design-patterns)
- [Testing Strategy](#testing-strategy)

---

## Project Structure

```
pocky-ops-bot/
├── cmd/
│   └── bot/
│       └── main.go                    # Application entry point
├── internal/
│   ├── bot/
│   │   ├── dispatcher.go              # Per-chat goroutine routing
│   │   ├── dispatcher_test.go
│   │   ├── router.go                  # Command routing
│   │   ├── router_test.go
│   │   └── handlers/
│   │       ├── command.go             # /start, /trogiup handlers
│   │       └── command_test.go
│   │   └── types/
│   │       ├── chat.go                # Chat-related types
│   │       ├── common.go              # Common utility types
│   │       ├── forum.go               # Forum/topic types
│   │       ├── inline.go              # Inline query & keyboard types
│   │       ├── media.go               # Media file types
│   │       ├── member.go              # Chat member types
│   │       ├── message.go             # Message types
│   │       ├── messaging.go           # Messaging utility types
│   │       ├── payment.go             # Payment/invoice types
│   │       ├── update.go              # Update types
│   │       └── user.go                # User types
│   ├── clients/
│   │   ├── telegram/
│   │   │   ├── api.go                 # API response types & interfaces
│   │   │   ├── backoff.go             # Exponential backoff strategy
│   │   │   ├── poller.go              # Long-polling implementation
│   │   │   ├── poller_test.go
│   │   │   ├── sender.go              # Message/action sending
│   │   │   ├── sender_test.go
│   │   │   └── update_types.go        # Update type constants & helpers
│   │   ├── llm/
│   │   │   ├── client.go              # Multi-provider LLM client
│   │   │   ├── client_test.go
│   │   │   ├── types.go               # ChatMessage, ToolCall, ToolDefinition
│   │   │   └── errors.go              # LLM error types
│   │   └── binance/
│   │       ├── client.go              # Base HTTP client + signing
│   │       ├── client_test.go
│   │       ├── types.go               # Shared types
│   │       ├── errors.go
│   │       ├── account.go             # Spot account endpoints
│   │       ├── account_test.go
│   │       ├── market.go              # Market data endpoints
│   │       ├── market_test.go
│   │       ├── signer.go              # HMAC-SHA256 request signing
│   │       ├── signer_test.go
│   │       ├── futures_client.go      # Futures HTTP client
│   │       ├── futures_types.go       # Futures-specific types
│   │       ├── futures_account.go     # Futures account endpoints
│   │       ├── futures_orders.go      # Futures order endpoints
│   │       └── futures_trades.go      # Futures trade endpoints
│   ├── config/
│   │   └── config.go                  # Configuration loading (30+ env vars)
│   ├── services/
│   │   ├── chat.go                    # Stateless ChatService with tool loop
│   │   └── chat_test.go
│   └── tools/
│       ├── types.go                   # ToolResult type
│       ├── registry.go                # Tool registry
│       ├── registry_test.go
│       ├── executor.go                # ToolExecutor interface
│       └── binance/
│           ├── tools.go               # Spot tools (balances, prices, 24hr stats)
│           ├── futures_tools.go       # Futures tools (account, positions, orders, trades, income)
│           └── tools_test.go
├── docs/
│   └── ARCHITECTURE.md                # This file
├── .env.example                       # Environment variable template
├── .gitignore
├── AGENTS.md                          # AI agent guidelines
├── CHANGELOG.md
├── go.mod
├── go.sum
└── README.md
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                          TELEGRAM API                                │
│               https://api.telegram.org/bot<TOKEN>                   │
└──────────────────────────────┬──────────────────────────────────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENTS LAYER                                │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │               internal/clients/telegram                       │   │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌─────────────┐  │   │
│  │  │  Poller  │  │  Sender  │  │APIResponse│  │ExponentialB.│  │   │
│  │  │ Start()  │  │SendText()│  │ - OK      │  │ NextBackoff │  │   │
│  │  │ Stop()   │  │SendAction│  │ - Result  │  │ Reset()     │  │   │
│  │  │ GetMe()  │  └──────────┘  └──────────┘  └─────────────┘  │   │
│  │  └──────────┘                                                 │   │
│  └──────────────────────────────────────────────────────────────┘   │
└──────────────────────────────┬──────────────────────────────────────┘
                               │ types.Update
                               ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           BOT LAYER                                  │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                    internal/bot                               │   │
│  │  ┌────────────┐  ┌──────────────────────────────────────┐   │   │
│  │  │   Router   │  │           Dispatcher                  │   │   │
│  │  │ RegisterCmd│  │  sync.Map: chatID → chatWorker        │   │   │
│  │  │ Handle()   │  │  per-chat goroutine + history         │   │   │
│  │  └────────────┘  │  /xoa → clear history                │   │   │
│  │                   │  /dautu → portfolio prompt           │   │   │
│  │                   └──────────────────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────────┘   │
└──────────┬───────────────────────────────────────────────────────────┘
           │ GenerateResponse(history, text)
           ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        SERVICES LAYER                                │
│  ┌──────────────────────────────────────────────────────────────┐   │
│  │                  internal/services                            │   │
│  │  ChatService { aiClient, systemPrompt, tools, vietnamese }   │   │
│  │  Tool call loop → max rounds safety limit                    │   │
│  └──────────────────────────────────────────────────────────────┘   │
└──────────┬─────────────────────────┬───────────────────────────────┘
           │ Complete()              │ Execute(toolName, args)
           ▼                         ▼
┌──────────────────────┐   ┌────────────────────────────────────────┐
│  internal/clients/   │   │        internal/tools/                  │
│       llm/           │   │  Registry { tools map }                 │
│  Gemini / Claude /   │   │  ┌─────────────────────────────────┐   │
│  OpenAI / Qwen       │   │  │    internal/tools/binance/       │   │
│  Tool calling        │   │  │  Spot: balances, prices, 24hr    │   │
└──────────────────────┘   │  │  Futures: account, positions,    │   │
                            │  │          orders, trades, income  │   │
                            │  └─────────────────────────────────┘   │
                            └──────────────┬─────────────────────────┘
                                           │
                                           ▼
                            ┌─────────────────────────────────────────┐
                            │       internal/clients/binance/          │
                            │  Spot: account, market endpoints         │
                            │  Futures: account, orders, trades        │
                            │  HMAC-SHA256 request signing             │
                            └─────────────────────────────────────────┘
```

---

## Core Components

### 1. Entry Point ([cmd/bot/main.go](../cmd/bot/main.go))

Wires all components together:
1. Load config → setup logger
2. Create Telegram poller + sender
3. Register bot command menu with Telegram (`/start`, `/dautu`, `/xoa`, `/trogiup`)
4. Create LLM client with provider/model from config
5. Optionally create Binance clients (spot + futures) and register 8 tools
6. Create stateless `ChatService`
7. Build `Router` + `CommandHandler`
8. Create `Dispatcher` (per-chat goroutines)
9. Start polling → graceful shutdown on SIGINT/SIGTERM

### 2. Telegram Client ([internal/clients/telegram/](../internal/clients/telegram/))

#### Poller ([poller.go](../internal/clients/telegram/poller.go))

Long-polling implementation with retry and backoff:

| Component | Description |
|-----------|-------------|
| `Poller` | Main struct managing long-polling lifecycle |
| `PollerConfig` | Options (token, timeout, retries, backoff, logger) |
| `UpdateHandler` | `func(ctx, update) error` — registered via `StartWithHandler` |

**Functional Options:** `WithTimeout`, `WithPollInterval`, `WithMaxRetries`, `WithBackoff`, `WithLogger`, `WithHTTPClient`

#### Sender ([sender.go](../internal/clients/telegram/sender.go))

Outbound message sending:
- `SendText(ctx, chatID, text)` — sends a plain text message
- `SendChatAction(ctx, chatID, action)` — sends "typing…" indicator
- `SetMyCommands(ctx, commands)` — registers bot command menu

#### API Types / Backoff ([api.go](../internal/clients/telegram/api.go), [backoff.go](../internal/clients/telegram/backoff.go))

| Type | Purpose |
|------|---------|
| `APIResponse` | Generic Telegram API response wrapper |
| `APIError` | Error type; `IsRetryable()` returns true for 5xx / rate-limit |
| `BackoffStrategy` | Interface for pluggable retry backoff |
| `HTTPClient` | Interface for HTTP operations (enables test mocking) |

Exponential backoff: initial 1s, max 60s, multiplier 2.0.

### 3. Bot Layer ([internal/bot/](../internal/bot/))

#### Dispatcher ([dispatcher.go](../internal/bot/dispatcher.go))

Routes updates to **per-chat goroutines** — the core concurrency primitive:

```
Dispatcher.Dispatch(update)
  └─► sync.Map lookup by chatID
        ├── Existing worker? → enqueue update (non-blocking)
        └── New chat? → spawn goroutine, own local history []ChatMessage
              └─► runWorker loop:
                    ├── /xoa → clear history
                    ├── /dautu → inject portfolio prompt → AI
                    ├── /start, /trogiup → delegate to Router
                    └── text → SendChatAction("typing") → GenerateResponse → append history
```

**No shared mutable state** — each goroutine owns its conversation history. `sync.Map` holds only channel pointers.

**Functional Options:** `WithBufferSize(n)`, `WithIdleTTL(d)`, `WithMaxTurns(n)`

#### Router ([router.go](../internal/bot/router.go))

Lightweight command dispatcher:
- `RegisterCommand(cmd, handler)` — registers `/cmd` handler
- `SetChatHandler(handler)` — fallback for non-command text
- `Handle(ctx, update)` — routes by command prefix, strips `@botname` suffix

#### Command Handlers ([handlers/command.go](../internal/bot/handlers/command.go))

| Handler | Command | Description |
|---------|---------|-------------|
| `Start` | `/start` | Welcome message with command overview |
| `Help` | `/trogiup` | Full help/usage guide |

Uses `MessageSender` interface (injected, mockable).

### 4. LLM Client ([internal/clients/llm/](../internal/clients/llm/))

Provider-agnostic completion client supporting four AI backends:

| Provider | Default Model | Notes |
|----------|--------------|-------|
| `gemini` | `gemini-2.0-flash` | Google Generative AI |
| `claude` | — | Anthropic Claude API |
| `openai` | — | OpenAI Chat Completions |
| `qwen` | — | Alibaba Qwen (OpenAI-compatible) |

**Key types** (`types.go`):

```go
type ChatMessage struct {
    Role    Role   // RoleUser / RoleAssistant / RoleSystem / RoleTool
    Content string
    // tool call fields...
}

type ToolDefinition struct {
    Name        string
    Description string
    Parameters  map[string]any // JSON Schema
}
```

**Functional Options:** `WithProvider`, `WithModel`, `WithMaxTokens`, `WithBaseURL`, `WithLLMTimeout`, `WithLLMLogger`

### 5. Services Layer ([internal/services/](../internal/services/))

#### ChatService ([chat.go](../internal/services/chat.go))

Stateless service — history is passed in by the caller (Dispatcher owns it):

```
GenerateResponse(ctx, history, userText)
  └─► build messages: system + history + userText
  └─► loop (max rounds):
        ├── aiClient.Complete(messages, tools)
        ├── no tool calls? → return reply
        └── tool calls → registry.Execute(name, args) → append tool result → continue
```

**Interfaces defined at consumer side:**
```go
type AICompleter interface {
    Complete(ctx, messages, tools) (ChatMessage, error)
}
type ToolExecutor interface {
    Execute(ctx, name, args) (ToolResult, error)
    Definitions() []llm.ToolDefinition
}
```

**Functional Options:** `WithTools(executor)`, `WithVietnamese()`

### 6. Tool Framework ([internal/tools/](../internal/tools/))

#### Registry ([registry.go](../internal/tools/registry.go))

```go
type Registry struct { tools map[string]Tool }
func (r *Registry) Register(t Tool)
func (r *Registry) Execute(ctx, name, argsJSON) (ToolResult, error)
func (r *Registry) Definitions() []llm.ToolDefinition
```

#### Binance Tools

**Spot tools** ([tools.go](../internal/tools/binance/tools.go)):

| Tool | Description |
|------|-------------|
| `get_spot_balances` | Fetch non-zero spot asset balances |
| `get_ticker_prices` | Current price(s) for symbol(s) |
| `get_24hr_ticker_stats` | 24-hour price change stats |

**Futures tools** ([futures_tools.go](../internal/tools/binance/futures_tools.go)):

| Tool | Description |
|------|-------------|
| `get_futures_account` | Futures wallet summary (balance, PnL, margin) |
| `get_futures_positions` | All open futures positions |
| `get_futures_open_orders` | Pending futures orders |
| `get_futures_trades` | Recent futures trade history |
| `get_futures_income` | Futures income/funding history |

### 7. Binance Client ([internal/clients/binance/](../internal/clients/binance/))

REST client with HMAC-SHA256 signing:

| Package | Endpoints |
|---------|-----------|
| `account.go` | `GetAccount` (spot balances) |
| `market.go` | `GetTickerPrice`, `GetTicker24hr` |
| `futures_account.go` | `GetFuturesAccount` |
| `futures_orders.go` | `GetFuturesOpenOrders` |
| `futures_trades.go` | `GetFuturesTrades`, `GetFuturesIncome`, `GetFuturesPositions` |

Separate `NewClient` (spot) and `NewFuturesClient` (futures). Both accept `WithBaseURL` for testnet.

### 8. Configuration ([internal/config/config.go](../internal/config/config.go))

```go
type Config struct {
    // Telegram
    TelegramToken string
    LogLevel      string
    PollInterval  time.Duration
    Timeout       time.Duration
    MaxRetries    int

    // AI
    AIProvider     string        // gemini | claude | openai | qwen
    AIAPIKey       string
    AIModel        string
    AIBaseURL      string        // optional override
    AIMaxTokens    int
    AITimeout      time.Duration
    AISystemPrompt string
    AIVietnamese   bool          // force Vietnamese responses

    // Conversation
    ConversationMaxTurns int
    ConversationTTL      time.Duration

    // Binance (optional — tools disabled if empty)
    BinanceAPIKey         string
    BinanceSecretKey      string
    BinanceBaseURL        string
    BinanceFuturesBaseURL string
}
```

**Loading precedence:** Environment variables → `.env` file.

---

## Data Flow

```
1. STARTUP
   main.go
     └─► config.Load()                          # .env + ENV vars
     └─► telegram.NewPollerWithOptions(...)
     └─► telegram.NewSender(...)
     └─► sender.SetMyCommands(...)              # register /start /dautu /xoa /trogiup
     └─► llm.NewClient(...)
     └─► (optional) binance.NewClient(...)
     └─► tools.NewRegistry() + Register(8 tools)
     └─► services.NewChatService(aiClient, prompt, opts...)
     └─► bot.NewRouter() + RegisterCommand(...)
     └─► bot.NewDispatcher(router, chatService, sender, ...)
     └─► poller.StartWithHandler(ctx, dispatcher.Dispatch)

2. POLLING LOOP
   poller.Start(ctx)
     └─► pollLoop goroutine
           ├─► GET /getUpdates?timeout=30&offset=N
           ├─► On success: send to updates chan, reset backoff
           └─► On error: retryable → backoff; non-retryable → stop

3. UPDATE ROUTING
   dispatcher.Dispatch(ctx, update)
     └─► extract chatID
     └─► sync.Map LoadOrStore(chatID, &chatWorker{ch})
     └─► new worker? → go runWorker(ctx, chatID, ch)
     └─► non-blocking send to worker.ch

4. PER-CHAT WORKER
   runWorker(ctx, chatID, ch)
     └─► owns: history []llm.ChatMessage
     └─► idle timer (ConversationTTL) → goroutine exits
     └─► on update:
           ├─► /xoa → clear history
           ├─► /dautu → inject portfolio prompt → fall through to AI
           ├─► /start, /trogiup → router.Handle()
           └─► text / portfolio prompt:
                 └─► sender.SendChatAction("typing")
                 └─► chatService.GenerateResponse(history, text)
                       └─► llm.Complete(messages, tool defs)
                       └─► tool call? → registry.Execute() → loop
                       └─► return reply
                 └─► append {user, assistant} to history, trim to maxTurns
                 └─► sender.SendText(reply)

5. SHUTDOWN
   SIGINT/SIGTERM
     └─► cancel() context
     └─► poller.Stop()           # close stopCh, drain updates chan
     └─► dispatcher.Shutdown()   # wg.Wait() for all workers
```

---

## Package Details

### Package Dependencies

```
cmd/bot/main
    ├── internal/config
    ├── internal/clients/telegram
    ├── internal/clients/llm
    ├── internal/clients/binance
    ├── internal/bot
    ├── internal/bot/handlers
    ├── internal/services
    ├── internal/tools
    └── internal/tools/binance

internal/bot
    ├── internal/bot/types
    └── internal/clients/llm         (for ChatMessage history type)

internal/services
    ├── internal/clients/llm
    └── internal/tools               (ToolExecutor interface)

internal/clients/telegram
    ├── internal/bot/types
    └── net/http, log/slog, encoding/json

internal/clients/llm
    └── net/http, log/slog, encoding/json

internal/clients/binance
    └── crypto/hmac, net/http, log/slog, encoding/json

internal/tools/binance
    ├── internal/clients/binance
    └── internal/tools

internal/config
    └── github.com/joho/godotenv

internal/bot/types
    └── (no dependencies — pure data types)
```

### Interface Boundaries

```go
// HTTPClient — mocks HTTP layer in tests
type HTTPClient interface {
    Do(req *http.Request) (*http.Response, error)
}

// BackoffStrategy — pluggable retry strategies
type BackoffStrategy interface {
    NextBackoff(attempt int) time.Duration
    Reset()
}

// UpdateHandler — user-defined update processing
type UpdateHandler func(ctx context.Context, update types.Update) error

// MessageSender — Telegram outbound operations (bot layer)
type MessageSender interface {
    SendText(ctx context.Context, chatID int64, text string) error
    SendChatAction(ctx context.Context, chatID int64, action string) error
}

// ChatCompleter — AI generation (dispatcher → services)
type ChatCompleter interface {
    GenerateResponse(ctx context.Context, history []llm.ChatMessage, userText string) (string, error)
}

// AICompleter — LLM completion (services → clients/llm)
type AICompleter interface {
    Complete(ctx context.Context, messages []ChatMessage, tools []ToolDefinition) (ChatMessage, error)
}

// ToolExecutor — tool execution (services → tools)
type ToolExecutor interface {
    Execute(ctx context.Context, name string, args string) (ToolResult, error)
    Definitions() []llm.ToolDefinition
}
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `TELEGRAM_BOT_TOKEN` | (required) | Bot token from @BotFather |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `POLL_INTERVAL` | `1s` | Minimum time between polling requests |
| `TIMEOUT` | `30s` | Long-polling timeout (max 50s) |
| `AI_PROVIDER` | `gemini` | `gemini` / `claude` / `openai` / `qwen` |
| `AI_API_KEY` | (required) | API key for chosen AI provider |
| `AI_MODEL` | `gemini-2.0-flash` | Model name |
| `AI_BASE_URL` | — | Override default API endpoint |
| `AI_MAX_TOKENS` | `1024` | Max tokens in AI response |
| `AI_TIMEOUT` | `60s` | AI request timeout |
| `AI_SYSTEM_PROMPT` | `You are Pocky...` | System prompt |
| `AI_VIETNAMESE` | `true` | Force Vietnamese responses |
| `CONVERSATION_MAX_TURNS` | `20` | Max message pairs kept in history |
| `CONVERSATION_TTL` | `30m` | Idle timeout before history reset |
| `BINANCE_API_KEY` | — | Binance API key (tools disabled if empty) |
| `BINANCE_SECRET_KEY` | — | Binance secret for HMAC signing |
| `BINANCE_BASE_URL` | — | Override Binance spot API URL (testnet) |
| `BINANCE_FUTURES_BASE_URL` | — | Override Binance futures API URL (testnet) |

---

## Design Patterns

### 1. Functional Options

Used throughout for configurable components:

```go
poller, _ := telegram.NewPollerWithOptions(token,
    telegram.WithTimeout(30*time.Second),
    telegram.WithMaxRetries(5),
    telegram.WithLogger(logger),
)

dispatcher := bot.NewDispatcher(router, chatService, sender, logger,
    bot.WithBufferSize(5),
    bot.WithIdleTTL(30*time.Minute),
    bot.WithMaxTurns(20),
)
```

### 2. Per-Chat Worker Isolation

Each chat gets its own goroutine and owns its conversation history — no locks, no shared state:

```go
// Dispatcher: sync.Map keyed by chatID
val, loaded := d.workers.LoadOrStore(chatID, &chatWorker{ch: make(chan Update, bufSize)})
if !loaded {
    go d.runWorker(ctx, chatID, worker.ch)
}
```

### 3. Stateless Services

`ChatService.GenerateResponse` is pure: takes history in, returns reply + does not store state. History lives in the per-chat goroutine stack.

### 4. Small Interfaces at Consumer Side

Each layer defines the interface it needs, not the concrete type it gets — enabling test mocking without import cycles.

### 5. Error Classification

```go
func (e *APIError) IsRetryable() bool {
    return e.RetryAfter > 0 || e.Code >= 500
}
```

---

## Testing Strategy

### Coverage Areas

| Package | Test File | Coverage Focus |
|---------|-----------|---------------|
| `clients/telegram` | `poller_test.go`, `sender_test.go` | Lifecycle, retry, mock HTTP |
| `bot` | `dispatcher_test.go`, `router_test.go` | Routing, history management |
| `bot/handlers` | `command_test.go` | Command responses |
| `services` | `chat_test.go` | Tool loop, history handling |
| `clients/binance` | `*_test.go` | API parsing, signing |
| `tools` | `registry_test.go`, `tools_test.go` | Tool dispatch |

### Test Patterns

```go
// Mock HTTP client for isolation
type mockHTTPClient struct {
    responses []mockResponse
    requests  []*http.Request
}

// Table-driven tests
tests := []struct {
    name    string
    input   string
    wantErr bool
}{...}

// Interface mocks for service layer
type mockAI struct{ reply string }
func (m *mockAI) Complete(...) (llm.ChatMessage, error) { ... }
```

### Running Tests

```bash
go test ./...                          # all tests
go test ./... -cover                   # with coverage
go test ./... -race                    # race detector
go test ./internal/bot/... -v          # bot layer verbose
```

---

*Last updated: 2026-04-05*

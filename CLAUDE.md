# Pocky Ops Bot

A modular Telegram bot for personal workflows (Go).

## Build & Run

```bash
go mod download
cp .env.example .env          # fill in TELEGRAM_BOT_TOKEN
go run ./cmd/bot
go build -o bot ./cmd/bot
```

## Test

```bash
go test ./...
go test ./... -v -cover       # with coverage
go test ./... -race            # race detection
```

## Code Style

- Go only — format with `gofmt` + `goimports`, lint with `golangci-lint`
- Keep functions small; split files at ~300 lines or when mixing concerns
- No globals — pass dependencies via constructor functions

## Conventions

- **Functional options** for configurable components (e.g. `WithTimeout`, `WithLogger`)
- **Small interfaces** defined at the consumer side for testability
- **Error classification**: retryable (5xx, rate-limit) vs non-retryable (4xx)
- **Structured logging** via `log/slog` with key-value pairs
- **Graceful shutdown**: context cancellation + `sync.WaitGroup`
- Table-driven tests with `testify`; mock via interfaces, not concrete types
- >80% coverage target on service layer

## Security

- Never commit secrets — use env vars / `.env` (already in `.gitignore`)
- Validate all user inputs (commands, callbacks)
- Parameterized queries for any database access

## Architecture

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for full details: project structure, data flows, design patterns, and component reference.

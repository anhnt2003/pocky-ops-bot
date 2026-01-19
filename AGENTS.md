# Project Name

A modular Telegram bot for personal workflows, featuring commands for automation, notes, and lightweight integrations.

## Code Style

- Use Go for all new files
- Follow `gofmt` and `goimports`
- Prefer `golangci-lint` defaults (keep rules consistent across the repo)
- Keep functions small and readable; avoid unnecessary abstractions

## Architecture

- Follow MVC-style layering adapted for Go:
  - **Handlers/Controllers**: Telegram update handlers and routing
  - **Services**: business logic and orchestration
  - **Repositories**: data access and persistence
- Keep packages cohesive (by feature or domain), avoid circular dependencies
- Use dependency injection via constructor functions (explicit wiring, no magic)
- Keep files/modules focused; split when a file grows too large or mixes concerns

## Testing

- Write unit tests for all business logic (services and pure functions)
- Maintain >80% code coverage on service layer
- Use Go’s built-in testing (`testing`) and `testify` for assertions/mocks when helpful
- Prefer table-driven tests for business rules and edge cases

## Security

- Never commit tokens, API keys, or secrets (use env vars / secret managers)
- Validate and sanitize all user inputs (especially commands and callbacks)
- Use parameterized queries / prepared statements for database access
- Apply least-privilege for bot permissions and external integrations
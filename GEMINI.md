# Kambing Cup Backend: Development Rules

## Tech Stack & Dependencies
- **Go Version**: 1.24.0 (Utilize 'tool' directives in go.mod for CLI tools)
- **Router**: github.com/go-chi/chi/v5 (Use idiomatic middleware and sub-routers)
- **Database**: github.com/jackc/pgx/v5 (Prefer pgxpool for concurrency)
- **Migrations**: github.com/golang-migrate/migrate/v4
- **Auth**: github.com/golang-jwt/jwt/v5

## Coding Standards
- **Routing**: Always use Chi sub-routers for different modules (e.g., /users, /matches).
- **Middleware**: Leverage standard Chi middlewares: `Logger`, `Recoverer`, and `Timeout`.
- **Error Handling**: Wrap errors with context where appropriate, but prefer `pgx` specific error checking for DB constraints.
- **JSON**: Since we are on Go 1.24, use the `omitzero` tag for JSON structs where applicable.

## Architecture
- Keep `main.go` in `cmd/api/` minimal; it should only initialize dependencies and start the Chi router.
- Logic should reside in `internal/` to prevent external package leakage.
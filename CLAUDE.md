# cmdSnippetVault

CLI snippet manager for saving, tagging, searching, and executing shell commands. Built in Go as a portfolio project demonstrating senior-level patterns.

## Commands

```bash
just build          # go build -o bin/cv ./cmd/cmdSnippetVault
just test           # go test -race -coverprofile=coverage.out ./...
just lint           # golangci-lint run ./...
just coverage       # go tool cover -html=coverage.out
just install        # go install ./cmd/cmdSnippetVault
just clean          # rm -rf bin/ coverage.out
```

## Architecture

Strict layered architecture. Dependencies flow one direction only: `main.go → cli → service → storage interface → sqlite implementation`. No layer imports upward. The domain package imports nothing from this project.

```
cmd/cmdSnippetVault/main.go   # Entry point. Wires dependencies, calls root command.
internal/cli/                  # Cobra command definitions. One file per command.
internal/domain/               # Models (Snippet, ListFilter), interfaces, custom errors. Zero external deps.
internal/service/              # Business logic. Depends on domain interfaces, never on concrete storage.
internal/storage/sqlite/       # SQLite implementation of SnippetRepository.
internal/storage/sqlite/migrations/  # Embedded SQL migration files.
internal/config/               # Viper-based config loading (flags > env > file > defaults).
internal/clipboard/            # Clipboard abstraction with system and noop implementations.
internal/export/               # JSON/YAML import and export logic.
pkg/templatevar/               # Template variable parser for exec command. Reusable outside this project.
testdata/                      # Golden files and test fixtures.
```

## Code Style

- **Error handling**: Use custom error types from `internal/domain/errors.go`. Sentinel errors (`ErrNotFound`, `ErrDuplicate`, `ErrValidation`) for category checks with `errors.Is()`. Typed errors (`ValidationError`) with `Unwrap()` for detail extraction with `errors.As()`. Wrap errors with context at each layer: `fmt.Errorf("sqlite: get snippet %d: %w", id, err)`.
- **Interfaces**: Define in domain/service packages, not in storage. Accept interfaces, return structs.
- **Dependency injection**: Constructor-based only. No globals, no init(), no service locators. `main.go` is the only composition root.
- **Context**: Every repository and service method takes `context.Context` as first parameter.
- **Naming**: Follow Go conventions. No stuttering (`storage.StorageRepository` → `storage.Repository`). Acronyms fully capitalized (`ID`, not `Id`).
- **Comments**: Go doc comments on all exported types, methods, and functions. No `// TODO` or `// FIXME` in committed code.
- **Commits**: Conventional commits format (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`).

## Testing

- **Table-driven tests** for every test function with more than one case. No exceptions.
- **Test files** live next to source files: `snippet.go` → `snippet_test.go`, same package.
- **Test helpers** use `t.Helper()` and `t.Cleanup()`, not `defer`. Shared helpers go in `test_helpers_test.go`.
- **Integration tests** (storage layer) use in-memory SQLite (`:memory:`). Each test gets a fresh DB via `setupTestDB(t *testing.T)`.
- **Golden file tests** (CLI output) store expected output in `testdata/`. Regenerate with `-update` flag.
- **Mocks**: Hand-written structs implementing domain interfaces. No mock frameworks.
- **Coverage targets**: domain 95%+, service 85%+, storage 80%+, cli 70%+.
- **Assertions**: Use `testify/assert` (continues on failure) and `testify/require` (stops on failure for preconditions). No other test dependencies.

## Database

- **Driver**: `modernc.org/sqlite` (pure Go, no CGO).
- **DSN format**: `file:<path>?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)`
- **Migrations**: Managed by `golang-migrate/migrate`. SQL files embedded via `go:embed`, applied automatically on startup. CLI available via `migrate` for manual up/down during development.
- **Transactions**: Required for any operation touching multiple tables (Create, Update, Delete snippets with tags).

## Domain Model

```go
type Snippet struct {
    ID          int64
    Command     string
    Description string
    Tags        []string    // stored as junction table, exposed as []string
    UseCount    int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

Tags are `[]string` in the domain. The three-table schema (snippets, tags, snippet_tags) is a storage concern only.

## Key Interfaces

```go
// In domain package — implemented by sqlite, mocked in service tests
type SnippetRepository interface {
    Create(ctx context.Context, snippet *Snippet) error
    GetByID(ctx context.Context, id int64) (*Snippet, error)
    List(ctx context.Context, filter ListFilter) ([]Snippet, error)
    Update(ctx context.Context, snippet *Snippet) error
    Delete(ctx context.Context, id int64) error
    Search(ctx context.Context, query string) ([]Snippet, error)
}

type Clipboard interface {
    Copy(text string) error
    IsAvailable() bool
}
```

## Things to Watch Out For

- Never call `defer repo.Close()` before checking the error from `New()`.
- `sql.Open` with `modernc.org/sqlite` creates the DB file but not the parent directory — create it with `os.MkdirAll` first.
- `ON CONFLICT DO NOTHING` means `LastInsertId()` returns garbage if no insert happened. Always follow with a SELECT.
- The driver name for `modernc.org/sqlite` is `"sqlite"`, not `"sqlite3"`.

- Template variables in exec command (`{{host}}`) must prompt the user before execution. Never execute without explicit confirmation.

# Contributing to cmdSnippetVault

Thanks for your interest in contributing! This document covers the conventions and workflow for this project.

## Getting Started

1. Fork and clone the repository
2. Install prerequisites:
   - [Go 1.25+](https://go.dev/dl/)
   - [just](https://github.com/casey/just)
   - [golangci-lint](https://golangci-lint.run/) (for linting)
3. Run the tests to confirm everything works:
   ```bash
   just test
   ```

## Development Workflow

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feat/your-feature
   ```
2. Make your changes
3. Run tests and linter:
   ```bash
   just test
   just lint
   ```
4. Commit using [Conventional Commits](#commit-messages)
5. Open a pull request against `main`

## Project Structure

```
cmd/cmdSnippetVault/main.go       Entry point, dependency wiring (composition root)
internal/cli/                     Cobra command definitions (one file per command)
internal/domain/                  Models, interfaces, errors (zero external deps)
internal/service/                 Business logic
internal/storage/sqlite/          SQLite repository implementation
internal/config/                  Viper-based configuration
internal/clipboard/               Clipboard abstraction
internal/importer/                JSON/YAML import
internal/exporter/                JSON/YAML export
pkg/                              Template variable parser
```

**Layered architecture** - dependencies flow one direction only:

```
main.go -> cli -> service -> domain <- storage
```

No layer imports upward. The `domain` package imports nothing from this project.

## Code Style

### General

- Follow standard Go conventions and `golangci-lint` rules
- No `// TODO` or `// FIXME` in committed code
- Go doc comments on all exported types, methods, and functions

### Naming

- **Files**: `snake_case.go` (e.g., `snippet_service.go`, `json_exporter.go`)
- **Exported identifiers**: `PascalCase` (e.g., `SnippetService`, `NewJSONImporter`)
- **Unexported identifiers**: `camelCase` (e.g., `dbPath`, `tagID`)
- **Acronyms**: fully capitalized (e.g., `ID`, `DSN`, `JSON`, `URL`)
- **No stuttering**: `storage.Repository`, not `storage.StorageRepository`

### Error Handling

- Use custom error types from `internal/domain/errors.go`
- Sentinel errors (`ErrNotFound`, `ErrDuplicate`, `ErrValidation`) for category checks with `errors.Is()`
- Wrap errors with context at each layer:
  ```go
  fmt.Errorf("sqlite: get snippet %d: %w", id, err)
  ```

### Dependency Injection

- Constructor-based only. No globals, no `init()`, no service locators
- `main.go` is the only composition root
- Define interfaces in `domain`/`service` packages, not in implementations
- Every repository and service method takes `context.Context` as first parameter

## Testing

### Conventions

- **Table-driven tests** for every test function with more than one case
- **Test files** live next to source files (`snippet.go` -> `snippet_test.go`)
- **Test helpers** use `t.Helper()` and `t.Cleanup()` (not `defer`)
- **Assertions**: use `testify/assert` (continues on failure) and `testify/require` (stops on failure)
- **Mocks**: hand-written structs implementing domain interfaces, no mock frameworks

### Running Tests

```bash
just test        # Run all tests with race detector and coverage
just coverage    # Open coverage report in browser
```

### Integration Tests

Storage layer tests use in-memory SQLite (`:memory:`). Each test gets a fresh database via `setupTestDB(t)`.

### Coverage Targets

| Layer   | Target |
|---------|--------|
| domain  | 95%+   |
| service | 85%+   |
| storage | 80%+   |
| cli     | 70%+   |

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat: add tag autocomplete to add command
fix: handle empty search query without panic
docs: update installation instructions
refactor: extract template parsing into pkg
test: add edge cases for batch import
chore: update golangci-lint to v2.10
```

## Pull Requests

- Keep PRs focused on a single change
- Include tests for new functionality
- Ensure `just test` and `just lint` pass before opening a PR
- CI runs lint and tests automatically on all PRs against `main`

## Database

- **Driver**: `modernc.org/sqlite` (pure Go, no CGO)
- **Migrations**: SQL files in `internal/storage/sqlite/migrations/`, embedded via `go:embed` and applied automatically on startup using `golang-migrate/migrate`
- **Transactions**: required for any operation touching multiple tables

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).

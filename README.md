[![CI](https://github.com/LucasJaiser/cmdSnipperVault/actions/workflows/ci.yml/badge.svg)](https://github.com/LucasJaiser/cmdSnipperVault/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)

# cmdvault

A fast, local CLI snippet manager for saving, tagging, searching, and executing shell commands. Built in Go with a layered architecture, SQLite storage, and a polished terminal UI.

## Features

- **Save commands** with descriptions and tags for quick retrieval
- **Search** across commands, descriptions, and tags with relevance ranking
- **Execute snippets** directly, with support for template variables (`{{host}}`, `{{port}}`)
- **Edit snippets** in your preferred editor with a diff preview before saving
- **Import/Export** collections as JSON or YAML for backup and sharing
- **Clipboard integration** - automatically copies commands on retrieval
- **Shell completions** for bash, zsh, fish, and PowerShell
- **Local-first** - all data stored in a local SQLite database, no network required

## Installation

### From source

Requires [Go 1.25+](https://go.dev/dl/) and [just](https://github.com/casey/just).

```bash
git clone https://github.com/LucasJaiser/cmdSnipperVault.git
cd cmdSnipperVault
just install
```

### Build locally

```bash
just build
./bin/csv --help
```

## Quick Start

```bash
# Add a snippet interactively
cmdvault add

# Add a snippet with flags
cmdvault add -c "docker ps -a --format 'table {{.Names}}\t{{.Status}}'" \
             -d "List all containers with name and status" \
             -t docker,devops

# Search your collection
cmdvault search docker

# Get a snippet by ID (copies to clipboard)
cmdvault get 1

# Execute a snippet directly
cmdvault exec 1

# List all snippets
cmdvault list

# List snippets filtered by tag
cmdvault list -t docker

# List all tags
cmdvault list tags
```

## Commands

| Command      | Description                                           |
|--------------|-------------------------------------------------------|
| `add`        | Add a new snippet (interactive or via flags)          |
| `get <id>`   | Retrieve a snippet and copy it to the clipboard       |
| `exec <id>`  | Execute a snippet in your shell                       |
| `edit <id>`  | Edit a snippet in your configured editor              |
| `delete <id>`| Delete a snippet (with confirmation)                  |
| `list`       | List snippets with optional tag filter and pagination |
| `list tags`  | List all tags with snippet counts                     |
| `search`     | Search snippets by command, description, or tag       |
| `import`     | Import snippets from a JSON or YAML file              |
| `export`     | Export snippets to a JSON or YAML file                |
| `completion` | Generate shell completion scripts                     |

## Template Variables

Snippets can contain template variables using `{{name}}` syntax. When you execute a snippet with template variables, you'll be prompted to fill in the values:

```bash
# Save a parameterized command
cmdvault add -c "ssh {{user}}@{{host}} -p {{port}}" -d "SSH into a server" -t ssh,remote

# When you run `cmdvault exec <id>`, it will prompt:
#   user> admin
#   host> 192.168.1.100
#   port> 22
# Then execute: ssh admin@192.168.1.100 -p 22
```

## Configuration

Configuration is stored at `$XDG_CONFIG_HOME/cmdvault/config.yaml` (defaults to `~/.config/cmdvault/config.yaml`). A default config file is created on first run.

```yaml
clipboard: true           # Copy commands to clipboard on `get`
editor: nano              # Editor for `edit` command
database_path: ""         # Custom database path (default: ~/.local/share/cmdvault/cmdvault.db)
color: auto               # Color output: auto, always, never
confirm_execute: true     # Prompt before executing snippets
default_format: yaml      # Default import/export format: json, yaml
```

All options can also be set via environment variables with the `CMDVAULT_` prefix (e.g., `CMDVAULT_EDITOR=vim`) or via CLI flags.

## Import / Export

Back up your collection or share snippets with others:

```bash
# Export all snippets to YAML
cmdvault export backup.yaml -f yaml

# Export only docker-related snippets
cmdvault export docker.json -t docker

# Import from a file (duplicates are skipped)
cmdvault import snippets.yaml

# Preview an import without saving
cmdvault import snippets.yaml --dry-run
```

### Import file format

**JSON**
```json
[
  {
    "Command": "kubectl get pods -n {{namespace}}",
    "Description": "List pods in a namespace",
    "Tags": ["kubernetes", "devops"]
  }
]
```

**YAML**
```yaml
- command: kubectl get pods -n {{namespace}}
  description: List pods in a namespace
  tags:
    - kubernetes
    - devops
```

## Shell Completions

```bash
# Bash
source <(cmdvault completion bash)

# Zsh
cmdvault completion zsh > "${fpath[1]}/_cmdvault"

# Fish
cmdvault completion fish | source

# PowerShell
cmdvault completion powershell | Out-String | Invoke-Expression
```

## Architecture

```
cmd/cmdvault/main.go              Entry point, dependency wiring
internal/cli/                     Cobra command definitions
internal/domain/                  Models, interfaces, errors (zero external deps)
internal/service/                 Business logic
internal/storage/sqlite/          SQLite repository implementation
internal/config/                  Viper-based configuration
internal/clipboard/               Clipboard abstraction
internal/importer/                JSON/YAML import
internal/exporter/                JSON/YAML export
pkg/                              Template variable parser
```

Dependencies flow one direction: `main -> cli -> service -> domain <- storage`. No layer imports upward.

## Development

Requires [Go 1.25+](https://go.dev/dl/), [just](https://github.com/casey/just), and optionally [golangci-lint](https://golangci-lint.run/).

```bash
just build       # Build binary to ./bin/csv
just test        # Run tests with race detector
just lint        # Run golangci-lint
just coverage    # Open HTML coverage report
just clean       # Remove build artifacts
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute.

## License

[MIT](LICENSE) - Copyright (c) 2026 Lucas Jaiser

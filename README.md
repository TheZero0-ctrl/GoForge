# GoForge

GoForge is a Rails-inspired CLI for building Go APIs quickly.

The generated app layout and code organization are inspired by patterns from the book _Let's Go Further_.

## Implemented Command

### `goforge new <app-name>`

Creates a new Go API project skeleton, including:

- `cmd/api` application entrypoints and handlers
- `internal/data` and `internal/validator`
- `migrations/.keep`
- project files like `go.mod`, `README.md`, `.gitignore`, and `Makefile`

Supported flags for `new`:

- `--module <path>`: set an explicit Go module path
- `--skip-git`: skip `git init`
- `--skip-tidy`: skip `go mod tidy`

Global flags available across commands:

- `--dry-run`
- `--force`
- `--skip`

## Coming Soon

Planned next steps include:

- `goforge generate migration <name>`
- `goforge generate resource <name> <field:type>...`
- `goforge generate scaffold <name> <field:type>...`
- concrete `destroy` subcommands that reverse generated artifacts

## Quick Start

```bash
go build -o bin/goforge ./cmd/goforge
./bin/goforge new demo-api
```

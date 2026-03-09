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

## Install

### Binary Release (Linux/macOS/Windows)

You can manually download a prebuilt binary from the Releases page:

- `https://github.com/TheZero0-ctrl/GoForge/releases/latest`

Download the archive for your OS/CPU, extract it, and place `goforge` in your `PATH`.

### Automated install/update (Linux)

Always inspect scripts before piping into `bash`.

```bash
curl -fsSL https://raw.githubusercontent.com/TheZero0-ctrl/GoForge/main/scripts/install_update_linux.sh | bash
```

The installer defaults to `$HOME/.local/bin`. You can override with `DIR`:

```bash
curl -fsSL https://raw.githubusercontent.com/TheZero0-ctrl/GoForge/main/scripts/install_update_linux.sh | DIR=/usr/local/bin bash
```

Install a specific version by setting `VERSION` (for example `v0.1.0`):

```bash
curl -fsSL https://raw.githubusercontent.com/TheZero0-ctrl/GoForge/main/scripts/install_update_linux.sh | VERSION=v0.1.0 bash
```

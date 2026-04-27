# GoForge

GoForge is a Rails-inspired CLI for building Go APIs quickly.

The generated app layout and code organization are inspired by patterns from the book _Let's Go Further_.

## Implemented Commands

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

### `goforge db:create`, `goforge db:drop`, `goforge db:migrate`, `goforge db:rollback`, and `goforge db:migrate:force`

Creates/drops the configured PostgreSQL database, applies migrations, rolls back steps, and recovers dirty migration state

Supported flags:

- `--dsn <postgres-dsn>`: use an explicit PostgreSQL connection string
- `--env <name>`: load the DSN from `config/database.toml` for the selected environment

If `--dsn` is omitted, GoForge reads DSN from `config/database.toml` (`development` by default).

If `db:migrate` reports a dirty version error, recover with:

- `goforge db:migrate:force <version> --dsn <postgres-dsn>`
- rerun `goforge db:migrate`

### `goforge generate migration <name> [field:type ...]`

Generates a timestamped migration pair in `migrations/`:

- `migrations/<timestamp>_<name>.up.sql`
- `migrations/<timestamp>_<name>.down.sql`

Alias form is supported:

- `goforge g migration <name> [field:type ...]`

Common Rails-style names like `create_<table>`, `add_<column>_to_<table>`, and `remove_<column>_from_<table>` generate starter SQL; custom names fall back to empty files.

For `create_<table>` migrations, GoForge includes implicit `created_at` and `version` columns by default.

### `goforge generate resource <name> <field:type>...`

Generates CRUD resource files and wiring in an existing GoForge app, plus a timestamped `create_<resources>` migration pair:

- `internal/data/<resources>.go`
- `cmd/api/<resources>.go`
- updates `cmd/api/routes.go`
- updates `internal/data/models.go`
- `migrations/<timestamp>_create_<resources>.up.sql`
- `migrations/<timestamp>_create_<resources>.down.sql`

Alias form is supported:

- `goforge g resource <name> <field:type>...`


Global flags available across commands:

- `--dry-run`
- `--force`
- `--skip`

## Coming Soon

Planned next steps include:

- `goforge generate scaffold <name> <field:type>...`
- concrete `destroy` subcommands that reverse generated artifacts

## Quick Start

```bash
go build -o bin/goforge ./cmd/goforge
./bin/goforge new demo-api
./bin/goforge db:create --dsn "postgres://localhost:5432/demo_api?sslmode=disable"
./bin/goforge db:migrate --dsn "postgres://localhost:5432/demo_api?sslmode=disable"
./bin/goforge db:rollback 1 --dsn "postgres://localhost:5432/demo_api?sslmode=disable"
./bin/goforge db:migrate:force 20260420034348 --dsn "postgres://localhost:5432/demo_api?sslmode=disable"
./bin/goforge db:drop --dsn "postgres://localhost:5432/demo_api?sslmode=disable"
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

## Architecture
<img width="1240" height="644" alt="image" src="https://github.com/user-attachments/assets/6d5b4539-2143-46d4-ab49-ce25fcfb44ec" />


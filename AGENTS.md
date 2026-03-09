# AGENTS.md

Guidance for coding agents working in `goforge`.

## Purpose

- Keep changes aligned with the current architecture.
- Prefer small, deterministic, test-backed edits.
- Preserve package boundaries (`cli -> app -> domain -> infra`).

## Repository Layout

- `cmd/goforge`: binary entrypoint.
- `internal/cli`: Cobra wiring, flag parsing, terminal output.
- `internal/app`: execution pipeline and exit code handling.
- `internal/domain`: command specs, planners, and pure domain logic.
- `internal/infra`: OS adapters (filesystem and process runner).
- `test/e2e`: end-to-end tests (build-tagged).
- `docs/`: architecture and phased implementation notes.

## E2E Test Commands

- E2E tests are guarded by build tags (`//go:build e2e`).
- Run all e2e tests:
  - `go test -tags=e2e ./test/e2e`
- Run a single e2e test:
  - `go test -tags=e2e ./test/e2e -run '^TestNewCommandE2EScaffoldsApp$'`

## Lint / Formatting Commands

- Format all Go code:
  - `gofmt -w ./cmd ./internal ./test`
- Vet static issues:
  - `go vet ./...`
- Recommended pre-PR quality gate:
  - `gofmt -w ./cmd ./internal ./test && go vet ./... && go test ./...`

## Architecture Constraints

- Keep Cobra types inside `internal/cli`.
- `internal/domain` should stay pure and OS-agnostic.
- `internal/infra` implements side-effecting ports/interfaces.
- New commands should be registry-driven, not added via giant switch blocks.
- Follow plan-then-apply flow:
  1. Validate input.
  2. Build a `plan.Plan` (no side effects).
  3. Execute operations via `app.Executor`.
- Reuse shared flags behavior (`--dry-run`, `--force`, `--skip`) consistently.

## Go Style Conventions (Observed)

- Always run `gofmt` after edits.
- Keep imports grouped in this order:
  1. standard library
  2. external modules
  3. internal imports (`goforge/...`)
- Prefer small structs and focused functions.
- Use constructor helpers named `NewX` (for example, `NewExecutor`, `NewRegistry`).
- Keep interfaces narrow (for example, `Params`, `Runner`, `FS`).
- Prefer value-free constants for command IDs and operation types where practical.
- Keep exported identifiers in `CamelCase`; unexported in `camelCase`.
- Use descriptive names over abbreviations (`registry`, `executor`, `planned`).

## Types and Data Modeling

- Favor explicit structs over `map[string]any` for core domain data.
- Use dedicated enum-like string types for finite sets (`OperationType`, exit codes).
- Keep command input shape stable through `command.Input`.
- Add fields only when needed across layers; avoid speculative fields.

## Error Handling Conventions

- Return errors; avoid panic in production paths.
- Wrap with context using `fmt.Errorf("...: %w", err)` when propagating.
- Use plain `fmt.Errorf("...")` for validation failures without wrapped causes.
- Distinguish conflict/validation/execution outcomes via `app.ExitCode`.
- In tests, fail with clear context using `t.Fatalf(...)`.

## CLI and Output Conventions

- Keep user-facing output structured as status + message pairs.
- Write non-error output to stdout; errors to stderr.
- Preserve existing status vocabulary where possible (`INFO`, `PLAN`, `CREATE`, `UPDATE`, `SKIP`, `RUN`, `DONE`, `ERROR`).

## Testing Conventions

- Prefer table-driven tests when checking many input/output variants.
- Use `t.Parallel()` for pure unit tests that do not share mutable global state.
- Keep e2e tests in `test/e2e` and utilities in `test/testutil/e2e`.
- E2E tests should build the binary, run in `t.TempDir()`, and assert artifacts.
- Assert both behavior and critical generated file contents when relevant.

## Templates and Generation

- Template rendering is manifest-driven (`manifest.json` in template pack).
- Keep generated output deterministic (sort lists before rendering where needed).
- Use `text/template` with `Option("missingkey=error")`.
- Keep template variable names explicit and minimal.

## When Adding or Changing Commands

- Register commands through `NewDefaultRegistry()`.
- Implement `Spec()`, `Validate(...)`, and `Plan(...)` coherently.
- Keep command alias collisions impossible (registry already enforces this).
- Add/adjust unit tests for validation + planning behavior.
- If command has side effects, ensure dry-run emits plan entries without writing.

## Agent Checklist Before Finishing

- Run `gofmt -w` on changed Go files.
- Run targeted package tests for touched areas.
- Run `go test ./...` unless intentionally scoped.
- If e2e behavior changed, run `go test -tags=e2e ./test/e2e`.
- Ensure no layer-boundary violations were introduced.

# internal/config — Module Conventions

## Purpose

`internal/config` loads, validates, and exposes the layered configuration tree under `.moai/config/` — `config.yaml` (main) plus `sections/*.yaml` (quality, language, user, workflow, harness, design, constitution, system, git-convention). It also owns the canonical env-var name constants (`envkeys.go`) consumed by every other package.

This package is the deterministic source of truth for "what does the user want". Loader changes cascade everywhere — every CLI subcommand reads from it.

## Conventions

- **Section-file layout (CLAUDE.local.md §9)**: Each concern is a separate `sections/<topic>.yaml` file with a top-level key matching the topic name (e.g., `quality:`, `language:`, `user:`). Main `config.yaml` aggregates references. New section files MUST have a corresponding loader struct in `internal/config/` and unit tests in `_test.go`.
- **Configuration priority order**: (1) Environment variables (`MOAI_USER_NAME`, `MOAI_CONVERSATION_LANG`, ...) override file values. (2) User config in `.moai/config/sections/*.yaml`. (3) Template defaults from `internal/template/templates/.moai/config/`. Tests MUST verify this priority via `t.Setenv` + fixture file combinations.
- **Env var naming (CLAUDE.local.md §14 [HARD])**: All env var names live as constants in `envkeys.go` (e.g., `EnvUserName = "MOAI_USER_NAME"`). NEVER inline `os.Getenv("MOAI_*")` strings in consuming code — they MUST reference these constants. Hardcoded env names elsewhere are §14 violations subject to lint enforcement.
- **YAML frontmatter format**: `tools:`/`allowed-tools:` are CSV strings (space-separated), `skills:` is a YAML array — this is the only documented exception (CLAUDE.local.md §12). `metadata.*` fields are quoted strings. The section loaders use `gopkg.in/yaml.v3` in **non-strict** mode (no `KnownFields`): unknown keys are silently ignored and missing keys fall back to zero-value/defaults (see `resolver.go` `strictUnmarshalSection`). Only genuine type mismatches (e.g. a sequence where a string is expected, a string where an int is expected) are reported as `ConfigTypeError` — unknown keys do NOT fail loud.
- **`teammateMode` runtime ownership**: The `teammateMode` field in `.claude/settings.local.json` (NOT `.moai/config/`) is runtime-managed by `moai cc`/`moai glm`/`moai cg` commands. This module does NOT own that field. Reading it must go through the `internal/cli/settings.go` helpers, not `internal/config/`.
- **Cross-platform path conventions**: Default config search paths use `os.UserHomeDir()` + `filepath.Join` — never hardcode `/Users/` or `/home/`. Windows MUST resolve to `%USERPROFILE%\.moai\` correctly. Verify with `GOOS=windows go build ./internal/config/...`.

## Key Patterns

- **`Loader` struct + `Load()` method**: New section loaders implement `func (l *Loader) LoadXxx(path string) (*XxxConfig, error)`. Return concrete typed structs, never `map[string]interface{}` — type safety catches drift between schema and consumers.
- **`Defaults()` factory**: Each section provides a `XxxConfig.Defaults() *XxxConfig` factory returning the canonical defaults. Used by `moai init` to seed new projects and by tests to construct baseline fixtures. The defaults MUST match the template-shipped `sections/<topic>.yaml`.
- **Validation hook**: Each section struct may implement `Validate() error` for cross-field invariants (e.g., `coverage_threshold` ∈ [0, 1]). Loader calls `Validate()` immediately after unmarshal. Validation errors include the offending file path + line number when available.
- **`envkeys.go` constant catalog**: When a new env var enters the codebase, add the constant to `envkeys.go` in the alphabetized block, then `git grep` for any pre-existing inline literal and replace. CI lint catches `os.Getenv("RAW_STRING")` outside `envkeys.go` per CLAUDE.local.md §14.
- **Defensive `nil` config handling**: `LoadXxx` returning `(nil, nil)` is reserved for "file legitimately absent, use defaults". Callers MUST check `cfg == nil` and call `XxxConfig.Defaults()`. Returning `(nil, err)` means parse failure — callers MUST propagate the error, never silently fall back.

## Cross-References

- Root CLAUDE.md §9 (Configuration Reference)
- CLAUDE.local.md §2 (settings.local.json runtime ownership separation), §9 (Configuration System), §12 (YAML Frontmatter quick reference), §14 (Hardcoding prevention), §22 (Dev settings intent for `defaultMode`, `enableAllProjectMcpServers`, `teammateMode`, `env.PATH`)
- `internal/cli/settings.go` — runtime mutation helpers for `.claude/settings.local.json` (owns `teammateMode`, not this package)
- `internal/template/templates/.moai/config/` — canonical defaults source
- `internal/config/envkeys.go` — SSOT for env var names (all `os.Getenv` MUST reference these)

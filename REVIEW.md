# Code Review Guidelines

Review-specific rules for Claude Code Review. These apply only during code reviews, not general Claude Code sessions.

## Always Check

### Go Code Quality
- All errors must be wrapped with context: `fmt.Errorf("operation: %w", err)`
- No bare `panic()` in library code — only in `main()` or test helpers
- Goroutines must have proper lifecycle management (context cancellation, WaitGroup)
- Exported functions must have godoc comments in English
- No `//nolint` without an explanation comment

### Testing
- New exported functions must have corresponding test cases
- Tests must use `t.TempDir()` for temporary files — never write to project root
- Table-driven test pattern required for multiple test cases
- Race detector compatibility: no shared mutable state without synchronization

### Security
- No hardcoded credentials, API keys, or tokens
- No `filepath.Join(cwd, userPath)` when `userPath` can be absolute — use `filepath.Abs()`
- Environment variable reads must have fallback or validation
- No `os.Exit()` in library code

### Template System
- Changes to `internal/template/templates/` must be accompanied by `make build` verification
- Template variables (`.HomeDir`, `.GoBinPath`) must not be used for fallback paths in shell scripts — use `$HOME`
- `$CLAUDE_PROJECT_DIR` must always be quoted in hook commands

### Configuration
- YAML frontmatter `tools` and `allowed-tools` fields must use CSV string format, not YAML arrays
- `metadata` values must be quoted strings
- Agent `model` field: only `inherit`, `opus`, `sonnet`, `haiku`

## Style

- Prefer early returns over deeply nested conditionals
- Use named return values only when they improve godoc clarity
- Constants over magic numbers — define with meaningful names
- Prefer `errors.Is()` / `errors.As()` over string comparison

## Skip

- Generated files: `internal/template/embedded.go`
- Lock files: `go.sum` (formatting-only changes)
- Vendored dependencies: `vendor/`
- Binary artifacts: `*.exe`, `*.bin`
- IDE configuration: `.idea/`, `.vscode/`

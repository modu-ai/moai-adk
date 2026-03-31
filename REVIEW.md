# Code Review Guidelines

Review-only rules for Claude Code Review. These are additive on top of default correctness checks and apply exclusively during code reviews, not general Claude Code sessions. Rules already enforced by linters or defined in CLAUDE.md are intentionally excluded.

## Always check

- New exported functions or CLI commands have corresponding test cases
- New handler/route registrations have integration tests
- Goroutines have proper lifecycle management (context cancellation or WaitGroup)
- No bare `panic()` in library code — only in `main()` or test helpers
- No `os.Exit()` in library code — return errors to callers instead
- Error messages do not leak internal paths, stack traces, or implementation details to users
- Breaking changes to exported functions or types are documented in commit message
- Changes to `internal/template/templates/` include `make build` to regenerate embedded files
- Hook timeout values are appropriate for the operation (default 5s, long operations up to 60s)
- Cross-platform compatibility: no OS-specific assumptions without build tags

## Style

- Prefer early returns over deeply nested conditionals
- Use `errors.Is()` / `errors.As()` over string comparison for error checking
- Constants with meaningful names over magic numbers

## Skip

- `internal/template/embedded.go` (generated)
- `go.sum` (formatting-only changes)
- `vendor/` (vendored dependencies)
- `*.exe`, `*.bin` (binary artifacts)
- `.idea/`, `.vscode/` (IDE configuration)

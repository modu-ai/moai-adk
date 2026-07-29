# MoAI-ADK (Go Edition) - Technology Stack Document

## Primary Language

**Go 1.26**

Go is the implementation language for the MoAI-ADK rewrite. The project uses Go 1.26, which provides Green Tea GC (10-40% GC overhead reduction), enhanced routing patterns in `net/http`, range-over-int iterators, and the latest `log/slog` structured logging capabilities.

## Claude Code Integration

**Minimum Recommended Version: Claude Code v2.1.110+** (April 2026). v3.0 development tracks the v2.1.219+ subagent-nesting defaults and the v2.1.198+ background-subagent defaults.

**Opus 5 / 4.8 Model Matrix** (v3.0):
- Anchor model: `claude-opus-5` (1M-context, the default Opus as of Claude Code 2.1.219); `claude-opus-4-8` and `claude-opus-4-7` still supported
- Effort levels: `low` / `medium` / `high` (default) / `xhigh` / `max`. Opus 5 carries a previously-set effort level across sessions (no hold)
- Adaptive Thinking: enabled via `thinking: {type: "adaptive"}` -- Opus 4.7+ rejects fixed `budget_tokens` with HTTP 400, so fixed thinking budgets are prohibited
- 33-cell profile matrix: 11 retained agents x 3 model tiers (`low` / `medium` / `high`) materialized by `internal/template/profile_matrix.go` and rendered into agent frontmatter at deploy time
- GLM tier-models table (`internal/config/defaults.go`): `DefaultGLMHigh = "glm-5.2"` (NO `[1m]` suffix -- the `[1m]` is added at the launcher layer in `internal/cli/launcher.go` only when the 1M-context variant is requested); `DefaultGLMMedium = "glm-4.7"`; `DefaultGLMLow = "glm-4.5-air"`; `DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"`

## Go Module

```
module github.com/modu-ai/moai-adk
```

The module path follows Go conventions with the GitHub organization and repository name. All internal imports use this module path prefix.

---

## Technology Stack

### Core Dependencies (v3.0)

| Category | Package | Version | Purpose |
|----------|---------|---------|---------|
| CLI Framework | `github.com/spf13/cobra` | v1.10.2 | Command-line interface with subcommands, flags, and shell completion |
| CLI UX Wrapper | `charm.land/fang/v2` | v2.0.1 | Charm-branded Cobra wrapper (help rendering, themes) |
| YAML Parsing | `gopkg.in/yaml.v3` | v3.0.1 | YAML marshaling/unmarshaling for configuration and SPEC documents |
| TUI Components | `charm.land/bubbles/v2` | v2.1.1 | Reusable TUI components (spinners, progress bars, viewport, text input) |
| Terminal UI | `charm.land/bubbletea/v2` | v2.0.8 | Interactive TUI framework with Elm-architecture patterns |
| Terminal Forms | `charm.land/huh/v2` | v2.0.3 | Modern form components (Select, MultiSelect, Input, Confirm, Form) |
| Terminal Styling | `charm.land/lipgloss/v2` | v2.0.5 | Terminal layout and styling for statusline and UI components |
| Markdown Rendering | `github.com/charmbracelet/glamour` | v1.0.0 | Terminal markdown rendering with syntax highlighting and auto dark/light detection |
| HTML Templating | `github.com/a-h/templ` | v0.3.1020 | Type-safe HTML templating for `internal/web` (loopback HTTP console, Templ + HTMX) |
| Shell Parsing | `mvdan.cc/sh/v3` | v3.13.1 | POSIX shell parser used by hook and sandbox script analysis |
| Tree-sitter | `github.com/smacker/go-tree-sitter` | master | Tree-sitter AST bindings for code analysis |
| LSP Client | `github.com/charmbracelet/x/powernap` | v0.1.6 | Multi-language LSP client foundation. Wraps `sourcegraph/jsonrpc2` with VSCode-compatible codec. Used by `internal/lsp/core/` and `internal/lsp/transport/` |
| Struct Validation | `github.com/go-playground/validator/v10` | v10.30.3 | Config struct validation |
| TTY Detection | `github.com/mattn/go-isatty` | v0.0.24 | Terminal detection for headless mode support |
| Runewidth | `github.com/mattn/go-runewidth` | v0.0.27 | East-Asian width-aware text rendering |
| Terminal Env | `github.com/muesli/termenv` | v0.16.0 | Cross-platform terminal environment detection |
| Color Profile | `github.com/charmbracelet/colorprofile` | v0.4.3 | Correct color rendering across terminal capabilities |
| Text Processing | `golang.org/x/text` | v0.40.0 | Unicode normalization, language tag processing, and text utilities |
| Sync Primitives | `golang.org/x/sync` | v0.22.0 | `errgroup` / `semaphore` for parallel quality gates and LSP queries |
| System Calls | `golang.org/x/sys` | v0.47.0 | Low-level OS primitives (signal, syscall) |
| Test Assertions | `github.com/stretchr/testify` | v1.11.1 | Replaces the earlier "stdlib-only" stance at v3.0 |
| Goroutine Leak Detection | `go.uber.org/goleak` | v1.3.0 | Goroutine-leak assertions in tests |
| Configuration | Custom YAML loader | -- | 14 `loader_*.go` files composing 32 YAML (Viper was not used) |
| Git Operations | System Git via `exec.Command` | -- | All Git operations use system Git binary (go-git was not used) |
| Logging | `log/slog` (stdlib) | Go 1.26 | Structured, leveled logging with JSON and text handlers |
| Testing | `testing` (stdlib) | Go 1.26 | Standard test framework with benchmarks and fuzzing |
| HTTP Client | `net/http` (stdlib) | Go 1.26 | HTTP client for update checking and `moai web` loopback console |
| Concurrency | goroutines + channels (stdlib) | Go 1.26 | Native concurrent execution for LSP, quality gates, and parallel operations |
| File Embedding | `embed` (stdlib) | Go 1.26 | Compile-time template embedding via `//go:embed all:templates` in `internal/template/embed.go` |
| Context | `context` (stdlib) | Go 1.26 | Cancellation, timeouts, and request-scoped values |
| URI Encoding | `net/url` (stdlib) | Go 1.26 | RFC 3986 file:// URI encoding for LSP (`internal/lsp/gopls/uri.go`) |

### Security Patterns (v2.10.4)

Cross-cutting security patterns established by the 3-perspective review bundle:

| Pattern | Where | Purpose |
|---------|-------|---------|
| Path traversal rejection | `internal/evolution/safety.go` | `CheckFrozenGuard` rejects absolute paths (`filepath.IsAbs` + leading `/` for Windows) and `..` components |
| Binary allowlist | `internal/lsp/gopls/config.go`, `internal/astgrep/scanner.go` | Trusted prefix list (`/usr/bin`, `/usr/local/bin`, `/opt/homebrew/bin`, `$HOME/{go,.local,.cargo}/bin`) + shell metachar rejection (`;&|\`$`) |
| Cross-platform path normalization | Both above | `filepath.ToSlash` on both sides before HasPrefix comparison (Windows `filepath.Clean` produces backslashes) |
| ID validation | `internal/evolution/learning.go` | `validateLearningID` regex enforcement prevents path injection via `LearningEntry.ID` |

### Cross-Platform CI Compatibility (v2.10.4)

Established patterns for `Test (windows-latest)` parity:

1. **Path handling**: Use `filepath.ToSlash` for all prefix/equality comparisons. `filepath.IsAbs` returns false for Unix-style `/foo` on Windows — check leading `/` explicitly when Unix-style inputs are expected.
2. **Test paths**: Use `t.TempDir()` + explicit URI/path conversion helpers. Never hardcode Unix-style test paths that will be compared against OS-generated paths.
3. **Shell metachar lists**: Do NOT include backslash `\`. Go's `exec.Command` does not invoke a shell, so backslashes in Windows paths are safe and must not be rejected.
4. **Timing thresholds**: Use generous thresholds (100ms+) with small allowances (5%+) for async operations on Windows CI — scheduler granularity is coarser than Linux/macOS.

### Language Support

MoAI-ADK provides built-in internationalization with 4 supported languages:

| Code | Language |
|------|----------|
| `ko` | Korean (default for wizard UI) |
| `en` | English (default for configuration) |
| `ja` | Japanese |
| `zh` | Chinese |

**Single Source of Truth**: All language mappings are centralized in `pkg/models/lang.go`:

- `LangNameMap`: Map of language codes to display names
- `SupportedLanguages`: Ordered slice of language codes (Korean-first for wizard)
- `GetLanguageName(code)`: Returns display name for a language code
- `IsValidLanguageCode(code)`: Validates language codes

**DRY Principle**: This shared module is used by:
- `internal/cli/wizard/`: Language selection in init wizard
- `internal/template/`: Template context language resolution
- Configuration files: Language settings validation

### Documentation Site Stack

Official documentation site running at `https://adk.mo.ai.kr`:

| Component | Technology | Version |
|-----------|-----------|---------|
| Static site generator | Hugo Extended | v0.160.1+ |
| Theme | Hextra | v0.12.2+ (Hugo module) |
| Content format | Hugo Markdown + shortcodes | -- |
| Diagrams | Mermaid (Hextra built-in, client-side) | v11+ |
| Multilingual | Hugo multilingual (file-based) | 4 locales: ko / en / ja / zh |
| Search | FlexSearch | Hextra built-in |
| Hosting | Vercel | Framework Preset = Hugo |
| Edge runtime | Vercel Edge Function | `@runtime: 'edge'` for Accept-Language locale detection |
| Release automation | Go + git archive | `scripts/docs-version-snapshot` |

**Bun/Node.js removed**: the Phase-2 Nextra (Next.js + Bun) stack was fully replaced, so `docs-site/` builds with zero Node runtime dependency. Local development requires only the Hugo single binary.

### LSP Dependencies

| Category | Package | Purpose |
|----------|---------|---------|
| JSON-RPC transport | `github.com/charmbracelet/x/powernap` | VSCode-compatible JSON-RPC 2.0 codec |
| LSP Types | In-tree types under `internal/lsp/` (8 sub-packages) | MoAI-specific abstractions over raw protocol types |
| Multi-server management | `internal/lsp/core/`, `internal/lsp/aggregator/`, `internal/lsp/gopls/` | Lifecycle, parallel diagnostic collection, gopls bridge |

**Decision**: The `go.lsp.dev` packages were not used. The v3.0 LSP client (8 sub-packages) is built on `powernap` and wraps it with MoAI-specific abstractions (multi-server management, diagnostic aggregation, cache, circuit breaker). 16-language auto-detection via project_markers; `lsp.client_impl` feature flag selects `gopls_bridge` (legacy, Go-only) or `powernap_core` (default, 16 languages).

### Planned Dependencies Not Used

The following dependencies were considered during planning but replaced with simpler solutions during implementation:

| Planned Package | Replacement | Rationale |
|----------------|-------------|-----------|
| `github.com/spf13/viper` | Custom YAML loader (14 `loader_*.go` files) | Simpler, type-safe configuration without Viper's complexity |
| `github.com/go-git/go-git/v5` | System Git via `exec.Command` | Full Git feature coverage including worktrees without library limitations |
| `go.lsp.dev/protocol` | `github.com/charmbracelet/x/powernap` + in-tree types in `internal/lsp/` | Multi-language LSP client without `go.lsp.dev` coupling |
| `go.lsp.dev/jsonrpc2` | `powernap` JSON-RPC codec in `internal/lsp/transport/` | Lightweight implementation tailored to MoAI's needs |

> The "no testify" stance was reversed at v3.0 -- `github.com/stretchr/testify v1.11.1` is now a direct dependency. The reversal is recorded to prevent stale "testify was not used" claims from propagating back.

### Development Dependencies

| Category | Package | Purpose |
|----------|---------|---------|
| Linter | `github.com/golangci/golangci-lint` | Comprehensive Go linter aggregator (staticcheck, gosec, ineffassign, etc.) |
| PR Reviewer | CodeRabbit (`.coderabbit.yaml`) | Automated PR review; the `claude-pr-review` GitHub App was disabled at v3.0 in favor of CodeRabbit |
| Mock Generation | Manual test doubles | Interface-based test doubles written manually (mockery was not used) |
| Release | `github.com/goreleaser/goreleaser` | Cross-platform binary builds and release automation |
| Code Generation | `go generate` (stdlib) | Driving `go:embed` directives and templ source generation |
| CI Platform | GitHub Actions | Workflow files under `.github/workflows/` |
| Vulnerability Scan | `govulncheck` (stdlib golang.org/x/vuln) | Known vulnerability detection in dependency tree |

---

## Build System

### Primary Build

The canonical targets are defined in the project `Makefile` under `##` help banners -- run `make help` to see them. The Template-First cycle is: edit `internal/template/templates/` -> `make build` (recompiles the binary, re-embedding templates via `//go:embed all:templates`) -> run tests -> commit. There is NO separate `embedded.go` generation step.

```makefile
# Rebuild the binary (re-embeds templates at compile time)
build:
    go build -ldflags "-X github.com/modu-ai/moai-adk/pkg/version.Version=$(VERSION) -X github.com/modu-ai/moai-adk/pkg/version.Commit=$(COMMIT)" -o bin/moai ./cmd/moai

# Run all tests with coverage and race detector
test:
    go test -race -coverprofile=coverage.out ./...

# Run linters
lint:
    golangci-lint run ./...

# Cross-compile release binaries for all 6 platform targets
release:
    goreleaser release --clean
```

### Build Pipeline

1. **go generate**: Generate mocks, embed directives, and any code generation
2. **go vet**: Static analysis for common mistakes
3. **golangci-lint**: Comprehensive linting (staticcheck, gosec, ineffassign, gocritic, gofumpt)
4. **go test -race**: Run all tests with race detector enabled
5. **go build**: Compile the binary with version metadata via ldflags

### Release Pipeline (goreleaser)

```yaml
# .goreleaser.yml targets
builds:
  - goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X pkg/version.Version={{.Version}}
      - -X pkg/version.Commit={{.Commit}}
      - -X pkg/version.Date={{.Date}}
```

**Output**: Six binaries (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64, windows/arm64) plus checksums and a changelog.

---

## Development Environment

### Required Tools

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.26 | Compiler and standard toolchain |
| gopls | Latest | Go language server for IDE integration |
| golangci-lint | v1.62+ | Linter aggregator |
| goreleaser | v2.5+ | Release automation |
| mockery | -- | Not used; test doubles written manually |
| ast-grep | Latest | Structural code search (runtime dependency) |
| Git | 2.30+ | Version control (system install for worktree operations) |

### IDE Configuration

Recommended gopls settings for the project:

- **gofumpt**: Enabled (stricter formatting than gofmt)
- **staticcheck**: Enabled (advanced static analysis)
- **analyses**: All enabled by default
- **build tags**: None required for standard development

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `MOAI_CONFIG_DIR` | Override `.moai/` directory location | `.moai/` in project root |
| `MOAI_LOG_LEVEL` | Log verbosity (debug, info, warn, error) | `info` |
| `MOAI_LOG_FORMAT` | Log output format (text, json) | `text` |
| `MOAI_NO_COLOR` | Disable terminal colors | `false` |
| `MOAI_DEVELOPMENT_MODE` | Override `quality.yaml` development mode | (unset -- file value wins) |
| `MOAI_BRANCH_GUARD_EXEMPT` | Exempt a spawn from the main-checkout branch guard (set to `1`) | (unset -- guard active) |
| `MOAI_SYNC_GATE_BLOCKING` | Make the sync-phase quality gate blocking (set to `1`) | advisory (`systemMessage` only) |
| `CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH` | Cap subagent nesting depth (Claude Code runtime) | `3` on v2.1.219+ |
| `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` | Cap concurrent subagents (Claude Code runtime) | `20` |

> `MOAI_USER_NAME` and `MOAI_CONVERSATION_LANG` are NOT implemented in v3.0 -- user / language come from `.moai/config/sections/user.yaml` and `language.yaml` only. The retired `MOAI_RANK_API_URL` is also gone (ranking subsystem retired).

---

## Testing Strategy

### Test Framework

**Standard `testing` package** as the foundation, augmented at v3.0 by `github.com/stretchr/testify v1.11.1` for assertions. The test suite spans ~1000 `*_test.go` files (table-driven, `t.Parallel()` where safe, `t.TempDir()` for filesystem isolation, `-race` enforced in CI).

```
go test -race -coverprofile=coverage.out -covermode=atomic ./...
```

### Test Categories

| Category | Location | Purpose | Coverage Target |
|----------|----------|---------|-----------------|
| Unit Tests | `*_test.go` (same package) | Individual function/method testing | 85%+ |
| Integration Tests | `*_integration_test.go` | Cross-package interaction testing | Key paths |
| Benchmark Tests | `*_bench_test.go` | Performance regression detection | Critical paths |
| Fuzz Tests | `*_fuzz_test.go` | Input boundary discovery | Config parsing, CLI args |
| Hook Contract Tests | `internal/hook/*_test.go` | Hook execution contract verification in non-interactive shell | All hook events |
| JSON Safety Tests | `internal/template/*_test.go` | settings.json generation validity and path normalization | 100% of JSON output |

### Test Conventions

- **Table-driven tests**: Use Go's idiomatic table-driven test pattern for parameterized testing
- **Parallel execution**: Mark independent tests with `t.Parallel()` for faster execution
- **Test fixtures**: Place test data in `testdata/` directories (Go convention, ignored by build)
- **Manual test doubles**: Test doubles are written manually without mock generation frameworks
- **Race detection**: All CI runs include `-race` flag for goroutine safety verification

### Hook Execution Contract Testing

The Go edition introduces formal contract testing for Claude Code hook integration, addressing 4 regression cycles from the Python predecessor.

**Contract Test Strategy**:

| Test Type | Purpose | Environment |
|-----------|---------|-------------|
| Minimal PATH test | Verify `moai hook <event>` works with PATH containing only /usr/bin:/bin | exec.Command with clean env |
| JSON round-trip test | Verify settings.json generation → parse → re-serialize produces identical output | json.Marshal → json.Valid → json.Unmarshal |
| Non-interactive shell test | Verify hooks work without .bashrc/.zshrc loaded | exec.Command without shell wrapper |
| Path normalization test | Verify all deployed file paths use correct separators and no trailing slash issues | filepath.Clean + string validation |
| Cross-platform test | Verify hook behavior on darwin, linux, windows | CI matrix (goreleaser targets) |

**Why this matters**: The Python predecessor had no hook contract tests. Each regression was discovered by users in production, leading to 41+ emergency commits over 5 months. Contract tests catch regressions at CI time.

### Coverage Requirements

- **Overall project**: 85% minimum (aligned with TRUST 5 "Tested" principle)
- **Core domains** (`internal/core/`): 90% minimum
- **CLI layer** (`internal/cli/`): 70% minimum (UI-heavy, harder to unit test)
- **Public API** (`pkg/`): 95% minimum

---

## CI/CD and Deployment

### Continuous Integration

| Stage | Tool | Purpose |
|-------|------|---------|
| Build | `go build` | Compilation verification |
| Lint | `golangci-lint` | Code quality enforcement |
| Test | `go test -race` | Correctness and race condition detection |
| Coverage | `go tool cover` | Coverage threshold enforcement |
| Security | `gosec` (via golangci-lint) | Security vulnerability scanning |
| Vulnerability | `govulncheck` | Known vulnerability detection in dependencies |
| Hook Contract | `go test ./internal/hook/...` | Hook execution contract verification |
| JSON Safety | `go test ./internal/template/...` | Settings.json generation and validation |
| Path Integrity | Custom validator | Deployed file path normalization check |

### Deployment Model

**Single binary distribution** -- no container, no package manager, no runtime.

| Channel | Method | Target |
|---------|--------|--------|
| GitHub Releases | goreleaser | macOS (arm64, amd64), Linux (arm64, amd64), Windows (amd64, arm64) |
| Homebrew | Tap formula | macOS and Linux via `brew install modu-ai/tap/moai` |
| Go Install | `go install` | Developers with Go toolchain: `go install github.com/modu-ai/moai-adk/cmd/moai@latest` |
| Self-Update | Built-in `moai update` | In-place binary replacement with checksum verification |

### Release Process

1. Tag the release commit: `git tag v1.0.0`
2. goreleaser builds all platform binaries
3. Checksums and signatures generated
4. GitHub Release created with binaries and changelog
5. Homebrew formula updated automatically

---

## Performance Requirements

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Binary cold start | < 50ms | `time moai version` |
| Config load (cold) | < 10ms | Benchmark test |
| Config load (cached) | < 1ms | Benchmark test |
| CLI command P95 latency | < 200ms | End-to-end benchmark |
| LSP server startup (single) | < 500ms | Integration test |
| LSP diagnostic collection (16 servers) | < 2s | Parallel benchmark |
| Quality gate (full TRUST 5) | < 5s | Integration benchmark |
| Binary size (stripped) | < 30MB | Build output measurement |
| Memory usage (idle) | < 20MB | Runtime profiling |
| Memory usage (peak, 16 LSP servers) | < 200MB | Load testing |

### Performance Optimization Strategies

1. **Lazy initialization**: Language servers start on-demand, not at CLI startup
2. **Connection pooling**: Reuse LSP connections across quality gate iterations
3. **Parallel execution**: Goroutines for concurrent LSP queries, Git operations, and file I/O
4. **Caching**: Configuration values cached with `sync.Once`; LSP diagnostics cached with TTL
5. **Binary optimization**: Build with `-ldflags "-s -w"` to strip debug symbols (30-40% size reduction)

---

## Security Requirements

### Code-Level Security

| Requirement | Implementation |
|-------------|----------------|
| Input validation | All CLI arguments validated before processing |
| Path traversal prevention | `filepath.Clean()` + base directory containment checks |
| Secret detection | Pre-commit hook scanning via ast-grep rules |
| Dependency scanning | `govulncheck` in CI pipeline |
| OWASP compliance | gosec rules enabled in golangci-lint |
| JSON injection prevention | All JSON generated via `json.Marshal()` from Go structs, never string concatenation |
| Template expansion prohibition | No `${VAR}` or `{{VAR}}` tokens in generated JSON/YAML files at rest |
| Path normalization | `filepath.Clean()` + directory containment on all generated paths |

### Credential Management

| Credential | Storage | Access Method |
|------------|---------|---------------|
| Git credentials | System Git credential helper | System Git credential store |
| LSP server tokens | Environment variables | `os.Getenv()` with validation |
| GLM API token | `~/.moai/.env.glm` (local file) | `internal/cli/glm.go` `loadGLMKey()` |
| Hook skip audit trail | `.moai/logs/hook-skip.log` (text log) | `MOAI_BRANCH_GUARD_EXEMPT=1` opt-out writes a record |

### Supply Chain Security

- **go.sum verification**: Cryptographic checksums for all dependencies
- **govulncheck**: Known vulnerability detection in dependency tree
- **Minimal dependencies**: Prefer standard library over external packages
- **Dependency review**: New dependencies require explicit justification

---

## Architecture Decisions

### ADR-001: Why Go over Python

| Factor | Python (Current) | Go (New) |
|--------|-------------------|----------|
| Distribution | pip install + venv + Python runtime | Single binary, zero dependencies |
| Startup time | 200-500ms (interpreter + imports) | < 50ms (compiled) |
| Concurrency | asyncio/threading (GIL limited) | Goroutines + channels (native) |
| Type safety | Runtime (mypy optional) | Compile-time (enforced) |
| Memory | 50-100MB baseline | 10-20MB baseline |
| Cross-platform | Requires Python on target | Pre-compiled per platform |

**Decision**: Go eliminates the primary friction point of Python distribution (requiring Python runtime and virtual environments) while delivering 5-10x faster startup and native concurrency.

### ADR-002: internal/ vs pkg/ Boundary

**Rule**: A package goes in `pkg/` only if external tools need to import it. Everything else goes in `internal/`.

- `pkg/version/`: External tools may query MoAI-ADK version
- `pkg/models/`: External tools may need to parse MoAI-ADK data structures
- Everything else: `internal/` (CLI commands, domain logic, LSP client, etc.)

> v3.0 dropped the earlier `pkg/utils/` draft -- the v3.0 `pkg/` tree contains only `models` and `version`. General utilities (logging, path resolution, atomic file writes, validation) live in `internal/` packages or the standard library.

**Impact**: Aggressive internalization allows breaking changes to implementation details without semver bumps on the module.

### ADR-003: embed Package for Template Distribution

**Decision**: Use Go's `//go:embed` directive to bundle all Claude Code templates into the binary. Templates are located at `internal/template/templates/` (not a root-level `templates/` directory).

```go
//go:embed templates/*
var templateFS embed.FS
```

**Benefits**:
- No external file paths to resolve at runtime
- Template version always matches binary version
- Single file distribution includes all scaffolding
- Templates are read-only (prevents accidental modification)

**Trade-off**: Binary size increases by the template payload size (estimated 2-5MB). This is acceptable given the 30MB target.

### ADR-004: Go Interfaces for DDD Boundaries

**Decision**: Define Go interfaces at domain boundaries. Each domain package exports an interface that consumers depend on.

```go
// internal/core/git/manager.go
type Repository interface {
    CurrentBranch() (string, error)
    CreateBranch(name string) error
    HasConflicts(target string) (bool, error)
}
```

**Benefits**:
- Compile-time verification of domain contracts
- Easy mock generation for testing (mockery)
- Swap implementations without changing consumers
- Go's implicit interface satisfaction means no tight coupling

### ADR-005: log/slog for Structured Logging

**Decision**: Use Go 1.21+'s standard `log/slog` package instead of external logging libraries (zap, zerolog, logrus).

**Rationale**:
- Zero external dependency for logging
- Structured key-value pairs natively
- JSON and text output handlers built-in
- Performance comparable to zap for common use cases
- Standard library ensures long-term stability

### ADR-006: Charmbracelet for Terminal UI

**Decision**: Use Bubble Tea (TUI framework) and Lip Gloss (styling) from the Charmbracelet ecosystem for all terminal UI.

**Rationale**:
- Elm-architecture model aligns with Go's explicit state management
- Rich component library (spinners, progress bars, tables, text input)
- Active maintenance and large community
- Cross-platform terminal compatibility
- Lip Gloss provides CSS-like styling without ANSI escape code management

### ADR-007: System Git via exec (Replaced go-git)

**Decision**: Use system Git exclusively via `exec.Command("git", ...)` instead of go-git.

**Rationale**:
- System Git provides full feature coverage including worktrees, submodules, and all advanced operations
- No library limitations or version mismatches between go-git and system Git behavior
- Simpler implementation with consistent behavior across all Git operations
- Error wrapping in `internal/core/git/errors.go` ensures consistent error types

### ADR-008: Programmatic JSON Generation with Validation

**Decision**: Generate all JSON configuration files (settings.json, manifest.json) via Go struct serialization (`json.MarshalIndent()`), followed by `json.Valid()` verification. Never construct JSON via string concatenation or template variable substitution.

**Rationale**: The Python predecessor's settings.json was generated via template variable substitution (`{{HOOK_SHELL_PREFIX}}`, `${SHELL:-/bin/bash}`). This caused 4 regression cycles because:
1. Template variables containing quotes broke JSON syntax
2. Shell variable syntax (${SHELL}) was stored as literal strings in JSON
3. Platform-specific path separators were incorrectly escaped
4. Each fix introduced new edge cases in different platforms

Go's `json.Marshal()` produces valid JSON by construction — it's impossible to generate malformed JSON from a valid Go struct.

**Implementation**: `internal/template/settings.go` defines Go structs mirroring Claude Code's settings.json schema, then calls `json.MarshalIndent()` to produce the file.

---

## Technical Constraints

### Go-Specific Constraints

1. **No generics abuse**: Use generics only where type parameterization genuinely reduces code duplication (collections, result types). Prefer interfaces for behavioral polymorphism.
2. **Error handling**: Always return errors explicitly. Never use `panic()` for recoverable errors. Use `fmt.Errorf("context: %w", err)` for error wrapping.
3. **Context propagation**: All long-running operations accept `context.Context` as the first parameter for cancellation and timeout support.
4. **Package naming**: Packages use short, lowercase names without underscores. Package name should not repeat the directory structure (e.g., `git`, not `core_git`).
5. **No string-based JSON/YAML generation**: All structured data files (JSON, YAML) must be generated via Go struct serialization (`json.Marshal`, `yaml.Marshal`). String concatenation or `fmt.Sprintf` for structured file generation is prohibited. This prevents template expansion and escaping issues that caused 41+ regression commits in the Python predecessor.

### Compatibility Constraints

1. **Configuration compatibility**: YAML configuration files under `.moai/config/sections/` must remain format-compatible with the Python ADK to allow gradual migration.
2. **Template compatibility**: Templates in `templates/.claude/` must produce identical output as the Python ADK for the same inputs.
3. **CLI compatibility**: Command names and flag semantics must match the Python ADK where possible. New Go-specific flags are allowed as additions.
4. **LSP protocol compliance**: LSP client must conform to LSP 3.17 specification for all supported operations.

### Dependency Constraints

1. **Minimal external dependencies**: Prefer standard library solutions. Each new dependency requires justification.
2. **No CGo dependencies**: The binary must compile without CGo for maximum cross-platform portability (`CGO_ENABLED=0`).
3. **Version pinning**: All dependencies pinned to specific versions in `go.sum`. No floating version references.
4. **License compliance**: All dependencies must use permissive licenses (MIT, Apache-2.0, BSD). No GPL dependencies in the binary.

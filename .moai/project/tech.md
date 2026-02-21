# MoAI-ADK (Go Edition) - Technology Stack Document

## Primary Language

**Go 1.26**

Go is the implementation language for the MoAI-ADK rewrite. The project uses Go 1.26, which provides Green Tea GC (10-40% GC overhead reduction), enhanced routing patterns in `net/http`, range-over-int iterators, and the latest `log/slog` structured logging capabilities.

## Go Module

```
module github.com/modu-ai/moai-adk
```

The module path follows Go conventions with the GitHub organization and repository name. All internal imports use this module path prefix.

---

## Technology Stack

### Core Dependencies

| Category | Package | Version | Purpose |
|----------|---------|---------|---------|
| CLI Framework | `github.com/spf13/cobra` | v1.10.2 | Command-line interface with subcommands, flags, and shell completion |
| YAML Parsing | `gopkg.in/yaml.v3` | v3.0.1 | YAML marshaling/unmarshaling for configuration and SPEC documents |
| TUI Components | `github.com/charmbracelet/bubbles` | v1.0.0 | Reusable TUI components (spinners, progress bars, viewport, text input) |
| Terminal UI | `github.com/charmbracelet/bubbletea` | v1.3.10 | Interactive TUI framework with Elm-architecture patterns |
| Terminal Forms | `github.com/charmbracelet/huh` | v0.8.0 | Modern form components (Select, MultiSelect, Input, Confirm, Form) with themes and accessibility |
| Terminal Styling | `github.com/charmbracelet/lipgloss` | v1.1.1+ | Terminal layout and styling for statusline and UI components |
| Markdown Rendering | `github.com/charmbracelet/glamour` | v0.10.0 | Terminal markdown rendering with syntax highlighting and auto dark/light detection |
| TTY Detection | `github.com/mattn/go-isatty` | v0.0.20 | Terminal detection for headless mode support |
| Text Processing | `golang.org/x/text` | v0.34.0 | Unicode normalization, language tag processing, and text utilities |
| Configuration | Custom YAML loader | -- | Custom implementation in `internal/config/loader.go` (Viper was not used) |
| Git Operations | System Git via `exec.Command` | -- | All Git operations use system Git binary (go-git was not used) |
| Logging | `log/slog` (stdlib) | Go 1.26 | Structured, leveled logging with JSON and text handlers |
| Testing | `testing` (stdlib) | Go 1.26 | Standard test framework with benchmarks and fuzzing (testify was not used) |
| HTTP Client | `net/http` (stdlib) | Go 1.26 | HTTP client for ranking API and update checking |
| Concurrency | goroutines + channels (stdlib) | Go 1.26 | Native concurrent execution for LSP, quality gates, and parallel operations |
| File Embedding | `embed` (stdlib) | Go 1.26 | Compile-time template embedding into the binary |
| Context | `context` (stdlib) | Go 1.26 | Cancellation, timeouts, and request-scoped values |

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

### LSP Dependencies

| Category | Package | Version | Purpose |
|----------|---------|---------|---------|
| JSON-RPC | Custom implementation | -- | Lightweight JSON-RPC 2.0 client over stdio and TCP |
| LSP Types | Custom implementation | -- | LSP protocol type definitions in `internal/lsp/models.go` |
| LSP Protocol | Custom implementation | -- | JSON-RPC 2.0 transport in `internal/lsp/protocol.go` |

**Decision**: The `go.lsp.dev` packages were not used. Instead, a fully custom LSP client was implemented in `internal/lsp/` providing MoAI-specific abstractions (multi-server management, diagnostic aggregation) without external dependencies.

### Planned Dependencies Not Used

The following dependencies were considered during planning but replaced with simpler solutions during implementation:

| Planned Package | Replacement | Rationale |
|----------------|-------------|-----------|
| `github.com/spf13/viper` | Custom YAML loader | Simpler, type-safe configuration without Viper's complexity |
| `github.com/go-git/go-git/v5` | System Git via `exec.Command` | Full Git feature coverage including worktrees without library limitations |
| `github.com/stretchr/testify` | Standard `testing` package | Reduced external dependencies; standard assertions sufficient |
| `go.lsp.dev/protocol` | Custom types in `internal/lsp/` | Full control over LSP type definitions without external dependency |
| `go.lsp.dev/jsonrpc2` | Custom protocol in `internal/lsp/` | Lightweight implementation tailored to MoAI's needs |

### Development Dependencies

| Category | Package | Purpose |
|----------|---------|---------|
| Linter | `github.com/golangci/golangci-lint` | Comprehensive Go linter aggregator (staticcheck, gosec, ineffassign, etc.) |
| Mock Generation | Manual test doubles | Interface-based test doubles written manually (mockery was not used) |
| Release | `github.com/goreleaser/goreleaser` | Cross-platform binary builds and release automation |
| Code Generation | `go generate` (stdlib) | Driving mockery and embed directives |

---

## Build System

### Primary Build

```makefile
# Build the binary
build:
    go build -ldflags "-X pkg/version.Version=$(VERSION) -X pkg/version.Commit=$(COMMIT)" -o bin/moai ./cmd/moai

# Run all tests with coverage
test:
    go test -race -coverprofile=coverage.out ./...

# Run linters
lint:
    golangci-lint run ./...

# Generate mocks and embedded resources
generate:
    go generate ./...

# Cross-compile for all platforms
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
| `MOAI_RANK_API_URL` | Ranking API endpoint | Production URL |

---

## Testing Strategy

### Test Framework

**Standard `testing` package** only (testify was not used). All 20 test packages use Go's built-in test assertions.

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
| Ranking API token | System keyring (macOS Keychain, Linux secret-service) | `internal/rank/auth.go` |
| Git credentials | System Git credential helper | System Git credential store |
| LSP server tokens | Environment variables | `os.Getenv()` with validation |

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
- `pkg/utils/`: General utilities that are stable and useful outside the project
- Everything else: `internal/` (CLI commands, domain logic, LSP client, etc.)

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

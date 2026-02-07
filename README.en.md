# MoAI-ADK (Go Edition)

**[English](README.en.md)** | **[한국어](README.md)**

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Copyleft%203.0-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-30%20packages-brightgreen)](./internal/)
[![Coverage](https://img.shields.io/badge/Coverage-85--100%25-brightgreen)](#test-coverage)
[![Version](https://img.shields.io/badge/Version-2.0.0-success)](CHANGELOG.md)

High-performance Agentic Development Kit for Claude Code -- a complete rewrite of the Python-based MoAI-ADK (~73,000+ lines) in Go.

**Module:** `github.com/modu-ai/moai-adk`

---

## Overview

MoAI-ADK (Go Edition) is a compiled development toolkit that serves as the runtime backbone for the MoAI framework within Claude Code. It provides CLI tooling, configuration management, LSP integration, Git operations, quality gates, and autonomous development loop capabilities -- all distributed as a single binary with zero runtime dependencies.

### Why Go?

| Concern | Python Edition | Go Edition |
|---------|---------------|------------|
| Distribution | pip install + venv + dependencies | Single binary, zero dependencies |
| Startup | ~800ms interpreter boot | ~5ms native launch |
| Concurrency | asyncio / threading | Native goroutines |
| Type Safety | Runtime (mypy optional) | Compile-time enforcement |
| Cross-platform | Requires Python runtime | Pre-built binaries for all platforms |
| Hook Execution | Shell wrappers + Python interpreter | Compiled binary, direct JSON protocol |

### Key Characteristics

- **32,977 lines** of Go code across 30 packages
- **85-100% test coverage** per package (30 test packages)
- **Native concurrency** via goroutines for parallel LSP, quality checks, and Git operations
- **Embedded templates** using `go:embed` for bundled resources
- **Cross-platform** builds for macOS (arm64, amd64), Linux (arm64, amd64), and Windows
- **64 moai-* skills** optimized with progressive disclosure system
- **Agent Teams v2.0** (experimental) - Dual-mode execution engine

---

## What's New (v2.0)

### Agent Teams v2.0 (Experimental Feature)

Dual-mode execution engine supporting both sub-agent and Agent Teams workflows:

- **Sub-agent Mode**: Sequential specialized agent delegation (traditional)
- **Agent Teams Mode**: Parallel team-based development with shared task list
- **Auto Mode**: Intelligent mode selection based on complexity analysis

```
User Request
    |
    v
MoAI Orchestrator
    |
    +-- Mode Selector
        |
        +-- Sub-Agent Mode (Task() single agent)
        +-- Agent Teams Mode (TeamCreate + SendMessage)
        +-- Auto Mode (complexity-based automatic selection)
```

**Team Composition Patterns:**
- Plan team: researcher + analyst + architect (parallel exploration)
- Run team: backend-dev + frontend-dev + tester (parallel implementation)
- Sync phase: Always sub-agent mode

**Activation:**
1. Set `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` environment variable
2. Set `workflow.team.enabled: true` in configuration
3. Use `--team` or `--auto` flags

See [workflow-v2.md](.moai/docs/workflow-v2.md) for details.

### Auto-Update and Caching

- **Automatic Update Notifications**: Automatic version check on session start
- **Update Result Caching**: Minimizes unnecessary API calls
- **Rollback Support**: Automatic recovery on update failure

### StatusLine Improvements

- **Enhanced Version Display**: Clearer version information rendering
- **Extended Test Coverage**: Improved stability

### Hybrid Methodology (Default)

Recommended methodology for new projects and ongoing development:
- **New code**: TDD (RED-GREEN-REFACTOR)
- **Existing code**: DDD (ANALYZE-PRESERVE-IMPROVE)

### 64 moai-* Skills

Comprehensive skill collection optimized with progressive disclosure:

**Foundation Skills:**
- moai-foundation-claude, moai-foundation-core, moai-foundation-philosopher, moai-foundation-quality, moai-foundation-context

**Workflow Skills:**
- moai-workflow-spec, moai-workflow-project, moai-workflow-jit-docs, moai-workflow-templates, moai-workloop, moai-workflow-ddd, moai-workflow-tdd, moai-workflow-testing, moai-workflow-worktree, moai-workflow-thinking

**Domain Skills:**
- moai-domain-backend, moai-domain-frontend, moai-domain-database, moai-domain-uiux, moai-formats-data

**Language Skills (18 languages):**
- Go, Python, TypeScript, JavaScript, Java, Kotlin, Swift, C++, C#, Ruby, PHP, Rust, Elixir, Scala, R, Flutter

**Library Skills:**
- moai-library-shadcn, moai-library-nextra, moai-library-mermaid

**Platform Skills:**
- Vercel, Supabase, Neon, Convex, Firebase, Firestore, Auth0, Clerk, Railway

**Tool Skills:**
- moai-tool-ast-grep, moai-tool-svg

**Specialist Skills:**
- Figma integration, Flutter development, Rust engineering, Performance optimization, Guidelines

---

## Installation

### Quick Install (Recommended)

Install with a single command. The script automatically detects your platform and downloads the appropriate binary.

#### macOS, Linux, WSL

```bash
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash
```

#### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1 | iex
```

Or:

```powershell
Invoke-Expression (Invoke-RestMethod -Uri "https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.ps1")
```

#### Windows CMD

```batch
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat -o install.bat
install.bat
```

If curl is not available:
```batch
powershell -Command "Invoke-WebRequest -Uri 'https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.bat' -OutFile 'install.bat'"
install.bat
```

### From Source

Requires Go 1.25 or later.

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
make build
```

The compiled binary is placed at `bin/moai`.

### Install to GOPATH

```bash
make install
```

### Pre-built Binaries

Download platform-specific binaries from the [Releases](https://github.com/modu-ai/moai-adk/releases) page. Archives are available for:

- `darwin_arm64` (macOS Apple Silicon)
- `darwin_amd64` (macOS Intel)
- `linux_arm64`
- `linux_amd64`
- `windows_amd64`

---

## Quick Start

### Initialize a Project

```bash
moai init
```

Runs an interactive project setup wizard that detects your language, framework, and methodology, then generates the appropriate configuration and Claude Code integration files. Default methodology is **Hybrid**.

### Check System Health

```bash
moai doctor
```

Diagnoses your development environment, verifying tool availability, configuration integrity, and LSP server readiness.

### View Project Status

```bash
moai status
```

Displays a summary of project state including Git branch, quality metrics, and configuration status.

### Check for Updates

```bash
# Check for updates only
moai update --check

# Update to latest version
moai update

# Sync project templates only
moai update --project
```

### Manage Git Worktrees

```bash
moai worktree new feature/auth
moai worktree list
moai worktree switch feature/auth
moai worktree sync
moai worktree remove feature/auth
moai worktree clean
```

Full worktree lifecycle management for parallel branch development.

---

## Architecture

```
moai-adk/
+-- cmd/moai/             # Application entry point
|   +-- main.go
+-- internal/             # Private application packages
|   +-- astgrep/          # AST-based code analysis
|   +-- cli/              # Cobra command definitions
|   +-- config/           # YAML configuration management
|   +-- core/
|   |   +-- git/          # Git operations
|   |   +-- project/      # Project initialization & detection
|   |   +-- quality/      # TRUST 5 quality gates
|   +-- foundation/       # EARS patterns, TRUST 5, language defs
|   +-- git/              # Git operations wrapper
|   +-- hook/             # Claude Code hook system
|   |   +-- lifecycle/    # Hook lifecycle management
|   |   +-- quality/      # Hook-based quality checks
|   |   +-- security/     # Hook-based security scanning
|   +-- loop/             # Ralph feedback loop & state machine
|   +-- lsp/              # LSP client (16+ languages)
|   |   +-- hook/         # LSP hook integration
|   +-- manifest/         # File provenance tracking (SHA-256)
|   +-- merge/            # 3-way merge engine
|   +-- ralph/            # Convergence decision engine
|   +-- rank/             # Session ranking (HMAC-SHA256)
|   +-- resilience/       # Resilience patterns (circuit breaker, retry)
|   +-- shell/            # Shell execution and environment detection
|   +-- statusline/       # Claude Code statusline integration
|   +-- template/         # Template deployment & security
|   +-- ui/               # Interactive TUI components
|   +-- update/           # Self-update with rollback
+-- pkg/                  # Public library packages
|   +-- models/           # Shared data models
|   +-- utils/            # Common utilities
|   +-- version/          # Build version metadata
+-- templates/            # Embedded project templates
+-- Makefile              # Build automation
+-- .goreleaser.yml       # Release configuration
```

### Package Overview

| Package | Purpose | Coverage |
|---------|---------|----------|
| `config` | Modular YAML configuration with thread-safe concurrent access | 94.1% |
| `foundation` | EARS patterns, TRUST 5 principles, 18 language definitions, methodology engine | 98.4% |
| `core/git` | Git operations (branch, worktree, conflict, event detection) | 88.1% |
| `core/project` | Project initialization, language/framework detection, methodology auto-detection | 89.2% |
| `core/quality` | TRUST 5 quality gates with parallel validators and phase gates | 96.8% |
| `hook` | Compiled hook system for Claude Code (6 event types, JSON protocol) | 90.0% |
| `lsp` | LSP client supporting 16+ languages, parallel server management | 91.3% |
| `template` | Template deployment, settings generation, path security | 85.7% |
| `manifest` | File provenance tracking with SHA-256 integrity verification | 88.0% |
| `ui` | Interactive TUI (selector, checkbox, prompt, progress, wizard) | 96.8% |
| `statusline` | Claude Code statusline with git/memory/quality metrics | 100% |
| `astgrep` | AST-based code analysis and pattern matching | 89.4% |
| `rank` | Session ranking with HMAC-SHA256 authentication | 85.1% |
| `update` | Self-update with SHA-256 verification and auto-rollback | 87.6% |
| `merge` | 3-way merge engine with 6 strategies and conflict markers | 90.3% |
| `loop` | Ralph feedback loop with state machine and convergence detection | 92.7% |
| `ralph` | Convergence decision engine for autonomous iteration | 100% |
| `cli` | Cobra commands (init, doctor, status, version, worktree) | 92.0% |
| `cli/worktree` | Git worktree subcommands (new, list, switch, sync, remove, clean) | 100% |

### Core Concepts

**TRUST 5 Quality Framework** -- Every code change is validated against five pillars:

- **Tested**: 85%+ coverage, characterization tests for existing code
- **Readable**: Clear naming conventions, consistent code style
- **Unified**: Consistent formatting, import ordering
- **Secured**: OWASP compliance, input validation
- **Trackable**: Conventional commits, issue references

**Hook Execution Contract** -- Compiled binary hooks replace shell wrappers, supporting 6 Claude Code event types (PreToolUse, PostToolUse, SessionStart, SessionEnd, PreCompact, Notification) via a JSON protocol. All hook outputs must include the `hookEventName` field in `hookSpecificOutput` for proper protocol compliance.

**Zero-Touch Template Updates** -- 3-way merge engine with file provenance tracking enables automatic template updates without losing user customizations.

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `moai init` | Interactive project setup with language/framework detection |
| `moai doctor` | System health diagnostics and environment verification |
| `moai status` | Project state overview with git and quality metrics |
| `moai version` | Version, commit hash, and build date information |
| `moai hook <event>` | Hook dispatcher for Claude Code integration |
| `moai worktree new <name>` | Create a new Git worktree |
| `moai worktree list` | List all active worktrees |
| `moai worktree switch <name>` | Switch to an existing worktree |
| `moai worktree sync` | Synchronize worktree with upstream |
| `moai worktree remove <name>` | Remove a worktree |
| `moai worktree clean` | Clean up stale worktrees |
| `moai update` | Update to the latest version (with auto-rollback) |
| `moai update --check` | Check for updates without installing |
| `moai update --project` | Update project templates without updating binary |

### Update Command

The `moai update` command checks for and installs the latest release. It supports:

- **Dev versions**: Automatically checks for `go-v*` tagged releases (Go edition)
- **Production versions**: Checks for latest stable releases
- **Environment override**: Use `MOAI_UPDATE_URL` to check a different repository

```bash
# Check for updates
moai update --check

# Update to latest version
moai update

# Update project templates only (no binary update)
moai update --project

# Use custom repository (environment variable)
export MOAI_UPDATE_URL="https://api.github.com/repos/owner/repo/releases/latest"
moai update
```

#### Automatic Update Notification

A session-start hook automatically checks for new versions. When an update is available, a notification appears in the statusline. Update check results are cached to minimize API calls.

#### Release Tagging

Release tags use the `v*` format. The install script recognizes both `v*` and `go-v*` tags:

```bash
# Create a release tag (recommended)
git tag v2.0.0
git push origin v2.0.0
```

The install script automatically strips tag prefixes to construct the correct download URL.

---

## Development

### Prerequisites

- Go 1.25 or later
- `golangci-lint` (for linting)
- `gofumpt` (for formatting)

### Build

```bash
# Build binary to bin/moai
make build

# Build with version info from git tags
make build VERSION=v1.0.0

# Build and run
make run
```

### Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary to `bin/moai` |
| `make install` | Install binary to `$GOPATH/bin` |
| `make test` | Run tests with race detection and coverage |
| `make test-verbose` | Run tests with verbose output |
| `make coverage` | Generate HTML coverage report |
| `make lint` | Run golangci-lint |
| `make vet` | Run go vet |
| `make fmt` | Format code with gofumpt |
| `make tidy` | Tidy go modules |
| `make clean` | Remove build artifacts |
| `make all` | Run lint, test, and build |

### Build Flags

Version metadata is injected at build time via `-ldflags`:

```bash
go build -ldflags "-s -w \
  -X github.com/modu-ai/moai-adk/pkg/version.Version=v1.0.0 \
  -X github.com/modu-ai/moai-adk/pkg/version.Commit=$(git rev-parse --short HEAD) \
  -X github.com/modu-ai/moai-adk/pkg/version.Date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o bin/moai ./cmd/moai
```

---

## Testing

### Run All Tests

```bash
# Standard run with race detection
go test -race ./... -count=1

# Via Makefile (includes coverage output)
make test
```

### Test Coverage

Generate an HTML coverage report:

```bash
make coverage
# Opens coverage.html
```

### Test Coverage by Package

| Package | Coverage |
|---------|----------|
| `config` | 94.1% |
| `foundation` | 98.4% |
| `core/quality` | 96.8% |
| `ui` | 96.8% |
| `loop` | 92.7% |
| `cli` | 92.0% |
| `lsp` | 91.3% |
| `merge` | 90.3% |
| `hook` | 90.0% |
| `astgrep` | 89.4% |
| `core/project` | 89.2% |
| `manifest` | 88.0% |
| `core/git` | 88.1% |
| `update` | 87.6% |
| `template` | 85.7% |
| `rank` | 85.1% |
| `ralph` | 100% |
| `statusline` | 100% |
| `cli/worktree` | 100% |

### Development Methodology

The project uses **Hybrid** methodology as the default:

- **Hybrid (default)**: Recommended for new projects and ongoing development. TDD for new code, DDD for existing code
- **DDD (Domain-Driven Development)**: For existing/brownfield projects only. ANALYZE existing behavior, PRESERVE with characterization tests, IMPROVE incrementally
- **TDD (Test-Driven Development)**: For isolated new modules only. RED (write failing test), GREEN (make it pass), REFACTOR (clean up)

Methodology is auto-selected during `moai init` based on project state and can be changed in `.moai/config/sections/quality.yaml`.

---

## Releases

Releases are automated with [GoReleaser](https://goreleaser.com/). Each release produces:

- Statically linked binaries (`CGO_ENABLED=0`) for all supported platforms
- `tar.gz` archives (Linux, macOS) and `zip` archives (Windows)
- SHA-256 checksums in `checksums.txt`

---

## Contributing

Want to contribute? Check out [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

### Quick Start

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Write tests first (TDD for new code, characterization tests for existing code)
4. Ensure all tests pass: `make test`
5. Ensure code passes linting: `make lint`
6. Format your code: `make fmt`
7. Commit with conventional commit messages
8. Open a pull request

### Code Quality Requirements

- All packages must maintain 85%+ test coverage
- Zero lint errors, zero type errors
- Follow existing package structure and naming conventions
- Include table-driven tests where appropriate

### Contributors

All contributors are welcome! From small bug fixes to major features, every contribution is valued.

---

## License

Copyleft 3.0 - See [LICENSE](LICENSE) for details.

---

## Related Projects

- [MoAI-ADK (Python)](https://github.com/modu-ai/moai-adk) -- The original Python implementation (~73,000+ lines)
- [Claude Code](https://docs.anthropic.com/en/docs/claude-code) -- The AI development environment that MoAI-ADK extends

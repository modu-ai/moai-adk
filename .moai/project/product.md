# MoAI-ADK (Go Edition) - Product Document

## Project Overview

### Project Name

MoAI-ADK (Go Edition) -- Agentic Development Kit for Claude Code

### Description

MoAI-ADK is a high-performance, compiled development toolkit that serves as the runtime backbone for the MoAI framework within Claude Code. This Go edition is a complete rewrite of the existing Python-based MoAI-ADK (~73,000+ lines, 220+ source files), redesigned to leverage Go's strengths: compiled single-binary distribution, native concurrency via goroutines, strong static typing, and minimal runtime dependencies.

The toolkit provides CLI tooling, configuration management, LSP integration, Git operations, quality gates, and autonomous development loop capabilities -- all orchestrated through Claude Code's agent and skill system.

### Mission

Deliver a production-grade, zero-dependency binary that empowers Claude Code developers to initialize, manage, and maintain MoAI-powered projects with sub-second responsiveness and rock-solid reliability.

### Vision

Become the definitive development toolkit for AI-assisted software engineering -- enabling developers to focus on creative problem-solving while MoAI-ADK handles project scaffolding, quality enforcement, code analysis, and workflow automation at native speed.

---

## Target Audience

### Primary Users

1. **Claude Code Developers**: Engineers using Claude Code as their primary development environment who need project initialization, diagnostics, and quality enforcement tooling.

2. **MoAI Framework Users**: Teams and individuals adopting the MoAI agent-based development workflow (Plan-Run-Sync) who require the ADK as their local runtime.

3. **DevOps Engineers**: Operations personnel integrating MoAI-ADK into CI/CD pipelines via headless mode for automated quality checks and documentation generation.

### Secondary Users

4. **Open Source Contributors**: Developers extending MoAI-ADK with custom skills, agents, and plugins.

5. **Enterprise Teams**: Organizations deploying MoAI across multiple projects needing consistent tooling, configuration management, and quality standards.

---

## Core Features

The following features are mapped from the existing Python implementation, redesigned for idiomatic Go.

### 1. CLI Tool

Command-line interface providing project lifecycle management.

- **Project Initialization** (`moai init`): Interactive project setup with template selection, language configuration, and directory scaffolding
- **Diagnostics** (`moai doctor`): System health checks including Claude Code configuration validation, dependency verification, and environment diagnostics
- **Status** (`moai status`): Project state overview showing SPEC progress, quality metrics, and configuration summary
- **Updates** (`moai update`): Self-update mechanism and template version management
- **Branch Switching** (`moai switch`): Context-aware branch switching with worktree integration
- **Worktree Management** (`moai worktree`): Git worktree operations for parallel SPEC development

### 2. Configuration Management

Modular YAML-based configuration system with thread-safe concurrent access.

- **Sectioned Configuration**: Modular YAML files under `.moai/config/sections/` (user, language, quality, workflow)
- **Thread-Safe Access**: Go's `sync.RWMutex` for concurrent read/write safety across goroutines
- **Version Migration**: Automatic configuration migration across ADK versions with backup and rollback
- **Validation**: Schema-based configuration validation with actionable error messages
- **Environment Overlay**: Environment variable overrides for CI/CD and container deployments

### 3. LSP Integration

Language Server Protocol client supporting 16+ programming languages.

- **Multi-Language Support**: Go, Python, TypeScript, Java, Rust, C/C++, Ruby, PHP, Kotlin, Swift, Dart, Elixir, Scala, Haskell, Zig, and more
- **Diagnostic Collection**: Real-time error, warning, and hint aggregation from language servers
- **Quality Gate Integration**: LSP diagnostics feed directly into TRUST 5 quality validation
- **Server Lifecycle Management**: Automatic language server startup, health monitoring, and graceful shutdown
- **JSON-RPC Protocol**: Standards-compliant LSP communication over stdio and TCP transports

### 4. Git Operations

Comprehensive Git domain powered by system Git via exec for reliable operations.

- **Branch Management**: Create, switch, merge, and delete branches with MoAI naming conventions
- **Conflict Detection**: Pre-merge conflict analysis with resolution suggestions
- **Worktree Management**: Parallel development via Git worktrees with automatic registry tracking
- **Event Detection**: Git hook integration and event-driven automation
- **Hook Execution Contract**: Formal specification of hook runtime guarantees -- stdin JSON format, exit code semantics, timeout behavior, config access. Explicitly documents what is NOT guaranteed (user's PATH, shell environment, Python availability). Contract tests verify behavior across all 6 platform targets in CI, preventing regression.
- **Checkpoint System**: Automatic save points before risky operations with rollback capability

### 5. Quality Gates (TRUST 5)

Validation framework enforcing five quality principles across all code changes.

- **Tested**: 85%+ code coverage requirement, characterization test enforcement, mutation testing support
- **Readable**: Naming convention validation, code complexity analysis, LSP lint error checking
- **Unified**: Consistent style enforcement via language-specific formatters and linters
- **Secured**: OWASP compliance scanning, secret detection, input validation verification
- **Trackable**: Conventional commit enforcement, SPEC reference linking, audit trail maintenance

### 6. Statusline

Claude Code statusline rendering with real-time project metrics.

- **Git Status Display**: Branch, modified files, ahead/behind counts
- **Memory Metrics**: Context window utilization and token budget tracking
- **Quality Indicators**: TRUST 5 compliance status at a glance
- **Version Information**: ADK version, update availability notifications
- **Custom Themes**: Configurable output styles and color schemes

### 7. EARS Methodology

Enhanced Actor-Role-Scenario methodology for requirements specification.

- **Requirement Templates**: Structured EARS patterns (ubiquitous, event-driven, unwanted behavior, state-driven, optional)
- **SPEC Generation**: Automated SPEC document scaffolding from EARS requirements
- **Traceability**: Requirement-to-implementation linking through SPEC identifiers
- **Validation**: Completeness and consistency checking of requirement specifications

### 8. Loop Controller (Ralph)

Autonomous feedback loop engine for iterative development cycles.

- **Feedback Collection**: Automated gathering of build, test, and lint results
- **State Machine**: Configurable loop states (analyze, implement, test, review) with transition rules
- **Persistence**: Loop state storage for session resumption across Claude Code restarts
- **Convergence Detection**: Automatic detection of diminishing returns to prevent infinite loops
- **Human-in-the-Loop**: Configurable breakpoints for human review and approval

### 9. Performance Ranking

Session ranking and community submission system.

- **Session Metrics**: Token efficiency, task completion rate, quality scores
- **Leaderboard Integration**: Anonymous submission to community ranking boards
- **Authentication**: Secure credential management for ranking API access
- **Hook Integration**: Automatic metric collection via Git and session hooks

### 10. AST-Grep Integration

Code analysis via structural AST (Abstract Syntax Tree) pattern matching.

- **Pattern Matching**: Structural code search across multiple languages
- **Custom Rules**: User-defined AST patterns for project-specific analysis
- **Refactoring Support**: Pattern-based code transformations with preview
- **Integration**: AST analysis results feed into quality gates and SPEC validation

### 11. Multi-Model Architecture

Support for multiple LLM providers and hybrid cost-optimization modes.

- **Claude Mode** (`moai cc`): Full Claude model stack with per-agent model assignment (opus, sonnet, haiku) controlled by model policy
- **GLM Mode** (`moai glm`): Switch all agents to Z.AI's GLM models for cost reduction
- **Hybrid CG Mode** (`moai cg`): Claude leader with GLM workers via worktree-based environment isolation for 60-70% cost reduction on implementation tasks
- **Model Policy** (`moai init --model-policy`): Apply high/medium/low model distribution across all agent definitions based on role-specific mappings

### 12. Agent Teams Integration (Experimental)

Support for Claude Code's experimental Agent Teams API for parallel phase execution.

- **Team Orchestration**: Spawn and manage multiple specialized teammates (researcher, analyst, architect, backend-dev, frontend-dev, tester, quality)
- **Worktree Isolation**: Implementation agents run in isolated git worktrees to prevent file conflicts (`isolation: worktree`)
- **Background Execution**: Non-blocking parallel execution for implementation teammates (`background: true`)
- **Hook Integration**: WorktreeCreate/WorktreeRemove lifecycle hooks for worktree tracking
- **Graceful Fallback**: Automatically falls back to sub-agent mode when Agent Teams prerequisites are not met

---

## Use Cases

### Primary Use Cases

1. **MoAI Project Initialization**: Developers run `moai init` to scaffold a new project with Claude Code integration, including agents, skills, commands, hooks, and rules.

2. **Code Quality Validation**: During the Run phase (`moai run SPEC-XXX`), the ADK enforces TRUST 5 quality gates, collecting LSP diagnostics and running validation checks.

3. **Autonomous Development Workflow**: The Ralph loop controller enables iterative development cycles where Claude Code agents implement, test, and refine code with minimal human intervention.

4. **Parallel SPEC Development**: Using Git worktrees, developers work on multiple SPECs simultaneously in isolated environments without branch switching overhead.

5. **CI/CD Integration**: In headless mode, MoAI-ADK runs quality checks, generates documentation, and validates SPEC compliance as part of automated pipelines.

### Secondary Use Cases

6. **Legacy Project Onboarding**: Existing projects adopt MoAI-ADK incrementally by running diagnostics and generating initial project documentation.

7. **Team Standardization**: Enterprise teams deploy MoAI-ADK to enforce consistent project structure, quality standards, and development workflows across repositories.

8. **Documentation Synchronization**: The Sync phase (`moai sync SPEC-XXX`) generates and updates API documentation, changelogs, and README files from implementation changes.

---

## Success Indicators

| Metric | Target | Measurement |
|--------|--------|-------------|
| Binary startup time | < 50ms | Benchmark on cold start |
| CLI command response | < 200ms (P95) | End-to-end command execution |
| LSP diagnostic collection | < 2s for 16 languages | Parallel server startup |
| Configuration load time | < 10ms | Hot path profiling |
| Binary size | < 30MB | Release build with embedded templates |
| Memory usage (idle) | < 20MB | Runtime profiling |
| Test coverage | >= 85% | Go test coverage tooling |
| Installation steps | 1 (single binary) | User experience |
| Hook regression commits | 0 per release cycle | Git log analysis (vs Python's 41+ over 5 months) |
| settings.json generation failures | 0 | Contract test suite (vs Python's 4 regression cycles) |
| Destructive update overwrites | 0 | Manifest-based update verification (vs Python's 6 overwrite issues) |

---

## Competitive Advantages

1. **Single Binary Distribution**: No runtime dependencies, no virtual environments, no package managers -- one binary handles everything.

2. **Native Concurrency**: Go's goroutines enable parallel LSP server communication, concurrent quality checks, and non-blocking Git operations.

3. **Type Safety**: Go's static type system catches configuration and protocol errors at compile time rather than runtime.

4. **Cross-Platform**: Single codebase compiles to macOS (arm64, amd64), Linux (arm64, amd64), and Windows binaries via goreleaser.

5. **Embedded Templates**: Go's `embed` package bundles all Claude Code templates (hooks, skills, rules, agents, output-styles) directly into the binary.

6. **Performance**: Compiled binary eliminates Python interpreter startup overhead, delivering sub-50ms cold start times.

7. **Hook Execution Reliability**: The Python predecessor suffered 41+ PATH-related commits over 5 months with 4 regression cycles, trying 7 different approaches to make hooks work in Claude Code's non-interactive shell environment. The Go edition eliminates this entire problem class: hooks are compiled binary subcommands (`moai hook <event>`), requiring no shell wrappers, no path resolution, no JSON escaping, and no platform-specific shell detection. A formal Hook Execution Contract guarantees behavior across all platforms.

8. **Zero-Touch Template Updates**: A file manifest system (`.moai/manifest.json`) tracks every deployed file's provenance (template-managed, user-modified, user-created). Updates use a Git-like 3-way merge engine to safely apply template changes while preserving user customizations. This resolves the most chronic pain point in the Python predecessor, where `update.py` was modified 38 times and still caused destructive overwrites.

---

## Implementation Status

All 14 SPEC documents have been fully implemented. The Go codebase contains 40,000+ lines of Go code across 20+ test packages with 85-100% test coverage. Module path: `github.com/modu-ai/moai-adk`.

### Feature Completion

| Feature | Status | Notes |
|---------|--------|-------|
| CLI Tool | Complete | All commands implemented (init, doctor, status, update, hook, rank, worktree) |
| Configuration Management | Complete | Custom YAML loader with thread-safe access (no Viper dependency) |
| LSP Integration | Complete | Custom LSP client implementation (no go.lsp.dev dependency) |
| Git Operations | Complete | Pure Go git operations via exec (no go-git dependency) |
| Quality Gates (TRUST 5) | Complete | All five principles validated |
| Statusline | Complete | Real-time project metrics rendering |
| EARS Methodology | Complete | Requirement templates and validation |
| Loop Controller (Ralph) | Complete | Autonomous feedback loop engine |
| Performance Ranking | Complete | Session ranking and submission |
| AST-Grep Integration | Complete | Structural code analysis |
| Hook System | Complete | Compiled binary subcommands replacing 46 Python scripts |
| Manifest System | Complete | File provenance tracking |
| Merge Engine | Complete | 3-way merge for template updates |
| Self-Update | Complete | Binary self-replacement with rollback |

### Runtime Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/charmbracelet/bubbles` | v1.0.0 | TUI components (spinners, progress bars, text input) |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | Interactive TUI framework |
| `github.com/charmbracelet/lipgloss` | v1.1.1+ | Terminal layout and styling |
| `github.com/mattn/go-isatty` | v0.0.20 | TTY detection for headless mode |
| `github.com/spf13/cobra` | v1.10.2 | CLI framework with subcommands |
| `golang.org/x/text` | v0.34.0 | Text processing utilities |
| `gopkg.in/yaml.v3` | v3.0.1 | YAML configuration parsing |

### Design Decisions Made During Implementation

Several planned dependencies were replaced with simpler, purpose-built solutions:

- **No go-git**: Git operations use `exec.Command("git", ...)` for reliability and full feature coverage
- **No Viper**: Custom YAML loader in `internal/config/loader.go` provides simpler, type-safe configuration
- **No testify**: Standard `testing` package used throughout, reducing external dependencies
- **No go.lsp.dev packages**: Custom LSP protocol implementation in `internal/lsp/` for full control
- **Go 1.26**: Final Go version used is significantly newer than the originally planned Go 1.22+; includes Green Tea GC for 10-40% GC overhead reduction

---

## Constraints and Dependencies

### Hard Constraints

- Must maintain full feature parity with the Python MoAI-ADK
- Must be backward-compatible with existing `.moai/` directory structures
- Must support all 16+ languages currently handled by the LSP integration
- Configuration file format (YAML sections) must remain unchanged
- Generated configuration files (settings.json, manifest.json) must never contain unexpanded dynamic tokens ($VAR, {{VAR}}, ${SHELL}) -- all values must be resolved at generation time via Go struct serialization
- All file paths in deployed artifacts must be normalized via filepath.Clean() before writing

### External Dependencies

- Claude Code runtime environment
- Git (system installation for all Git operations via exec)
- Language servers (installed separately per language)
- ast-grep binary (for AST pattern matching)

### Development Constraints

- Go 1.26 (actual version used in implementation)
- All public APIs must follow Go conventions (exported names, godoc comments)
- Internal packages must use the `internal/` directory for encapsulation

# MoAI-ADK (Go Edition) - Comprehensive System Design

Version: 1.0.0
Date: 2026-02-03
Status: DRAFT
Cross-references: product.md, structure.md, tech.md, redesign-report.md

---

## 1. Executive Summary

### Project Scope

MoAI-ADK (Go Edition) is a complete rewrite of the Python-based MoAI-ADK (~73,000 LOC, 220+ files, 4,174 commits) into idiomatic Go. The project targets single-binary distribution with zero external runtime dependencies, replacing a complex Python ecosystem (pip/uv/pipx, virtual environments, 46 hook scripts, Rich/InquirerPy TUI) with a compiled, cross-platform binary.

### Key Architectural Decisions

| ADR | Decision | Rationale |
|-----|----------|-----------|
| ADR-001 | Modular Monolithic | Single binary, clear domain boundaries via Go packages |
| ADR-003 | Interface-Based DDD | Compile-time contracts, mockable dependencies, swappable implementations |
| ADR-006 | Hooks as Binary Subcommands | Eliminates 28 Python hook issues (PATH, encoding, SIGALRM) |
| ADR-007 | File Manifest Provenance | Prevents destructive updates via 3-way merge |
| ADR-011 | Zero Runtime Template Expansion | All JSON/YAML generated via struct serialization |
| ADR-012 | Hook Execution Contract | Formal guarantees prevent regression across 6 platform targets |

### Primary Goals

1. **Zero-dependency binary**: Single `moai` binary, no Python/Node.js/pip/uv required
2. **Sub-50ms startup**: Compiled binary eliminates interpreter overhead
3. **Hook reliability**: 28 hook issues resolved by compiled subcommand architecture
4. **Safe updates**: File manifest + 3-way merge prevents destructive overwrites (6 issues resolved)
5. **Cross-platform**: 6 platform targets (darwin/linux/windows x amd64/arm64), CGO_ENABLED=0
6. **75% LOC reduction**: ~18,000 Go LOC vs 73,000 Python LOC

### Current Implementation State

The project has a minimal scaffold in place:

- `cmd/moai/main.go`: Thin entry point delegating to `cli.Execute()`
- `internal/cli/`: Root command + stub commands (init, doctor, status, version)
- `pkg/version/`: Build-time version injection via ldflags
- `pkg/utils/`: Logger (slog) and project root discovery
- `pkg/models/`: Basic config and project type definitions
- `go.mod`: Module `github.com/modu-ai/moai-adk-go`, Go 1.25.6, cobra v1.10.2

---

## 2. Module Architecture (22 Modules)

### Phase 1 -- Foundation (~4,000 LOC)

#### `internal/config/` -- Configuration Management

Viper-backed YAML configuration with typed Go structs and sync.RWMutex for concurrent access. Replaces Python's UnifiedConfigManager singleton.

Files: `manager.go`, `types.go`, `defaults.go`, `migration.go`, `validation.go`
Resolves: 25 config/init issues (#315, #283, #206, #245, #243, #304)

#### `internal/hook/` -- Compiled Hook System

Replaces 46 Python hook scripts with compiled binary subcommands (`moai hook <event>`). Each hook handler is a Go function registered in a type-safe registry.

Files: `registry.go`, `protocol.go`, `contract.go`, `session_start.go`, `pre_tool.go`, `post_tool.go`, `session_end.go`, `stop.go`, `compact.go`
Resolves: 28 hook issues (#129, #259, #265, #269, #278, #288)

#### `internal/template/` -- Template Deployment

Extracts embedded templates (go:embed) to project directories with manifest tracking. Generates settings.json programmatically via json.MarshalIndent (ADR-011).

Files: `deployer.go`, `renderer.go`, `settings.go`, `validator.go`
Resolves: 6 template substitution issues (#304, #308, #309)

#### `internal/manifest/` -- File Provenance Tracking

Tracks every deployed file's origin, template hash, and current hash in `.moai/manifest.json`. Enables safe update decisions (overwrite, merge, skip).

Files: `manifest.go`, `hasher.go`, `types.go`
Resolves: 6 destructive overwrite issues (#162, #187, #236, #246, #318, #319)

### Phase 2 -- Core Domains (~5,000 LOC)

#### `internal/core/git/` -- Git Operations

Pure Go Git via go-git with system Git fallback for worktree operations. Branch management, conflict detection, event-driven automation.

Files: `manager.go`, `branch.go`, `conflict.go`, `event.go`

#### `internal/core/quality/` -- TRUST 5 Quality Gates

Orchestrates five quality principles (Tested, Readable, Unified, Secured, Trackable) with parallel validation. Aggregates LSP diagnostics, AST analysis, and Git metadata.

Files: `trust.go`, `validators.go`

#### `internal/lsp/` -- Language Server Protocol Client

Custom LSP client supporting 16+ languages via JSON-RPC 2.0 over stdio and TCP. Manages server lifecycle, concurrent diagnostic collection.

Files: `client.go`, `server.go`, `protocol.go`, `models.go`

#### `internal/core/project/` -- Project Initialization and Detection

Interactive project setup (bubbletea wizard), language/framework detection, project structure validation.

Files: `initializer.go`, `detector.go`, `validator.go`, `phase.go`

### Phase 3 -- Automation (~3,000 LOC)

#### `internal/loop/` -- Ralph Feedback Loop

State machine controller for iterative development cycles (analyze, implement, test, review). Persistence for session resumption, convergence detection.

Files: `controller.go`, `feedback.go`, `state.go`, `storage.go`

#### `internal/ralph/` -- Decision Engine

Determines next loop iteration action based on current state and feedback. Implements convergence heuristics.

Files: `engine.go`

#### `internal/update/` -- Self-Update System

Downloads new binary from GitHub Releases, orchestrates template merge via manifest, atomic binary self-replacement with rollback.

Files: `checker.go`, `updater.go`, `rollback.go`, `orchestrator.go`
Resolves: 4 package manager issues (#253, #296, #159, #312)

#### `internal/merge/` -- 3-Way Merge Engine

Git-like 3-way merge for template updates. Per-filetype strategies (line merge, YAML deep merge, JSON merge, section merge). Conflict detection with `.conflict` file generation.

Files: `three_way.go`, `strategies.go`, `conflict.go`, `differ.go`
Resolves: 15 update/migration issues (#246, #187, #318)

### Phase 4 -- UI and Integration (~3,500 LOC)

#### `internal/ui/` -- Charmbracelet TUI

Modern terminal UI using bubbletea (Elm architecture) and lipgloss (styling). Replaces Rich + InquirerPy.

Files: `wizard.go`, `selector.go`, `checkbox.go`, `progress.go`, `theme.go`, `prompt.go`
Resolves: 8 TUI issues (#268, #249, #286)

#### `internal/statusline/` -- Statusline Rendering

Claude Code statusline with real-time Git status, memory metrics, quality indicators, and version information.

Files: `builder.go`, `git.go`, `metrics.go`, `memory.go`, `renderer.go`, `update.go`

#### `internal/astgrep/` -- AST-Grep Integration

Structural code analysis via ast-grep CLI subprocess. Pattern matching, custom rules, refactoring support.

Files: `analyzer.go`, `models.go`, `rules.go`

#### `internal/rank/` -- Performance Ranking

Session metrics collection and community leaderboard submission. Secure credential management.

Files: `client.go`, `auth.go`, `config.go`, `hook.go`

### Phase 5 -- Knowledge and CLI (~2,500 LOC)

#### `internal/foundation/` -- Foundation Methodologies

EARS requirement patterns, language ecosystem definitions (16+ languages), domain architecture patterns, TRUST 5 principle definitions.

Files: `ears.go`, `langs.go`, `backend.go`, `frontend.go`, `database.go`, `testing.go`, `devops.go`, `trust/principles.go`, `trust/checklist.go`

#### `internal/core/integration/` -- Integration Testing Engine

Cross-package integration test execution and result reporting.

Files: `engine.go`, `models.go`

#### `internal/core/migration/` -- Version Migration

Automatic configuration migration across ADK versions with backup and rollback.

Files: `migrator.go`, `backup.go`

#### `internal/cli/` -- Cobra Command Definitions (Composition Root)

All CLI commands as Cobra commands. Also serves as the dependency injection wiring point.

Files: `root.go`, `init.go`, `doctor.go`, `status.go`, `update.go`, `hook.go`, `switch.go`, `rank.go`, `worktree/*.go`

#### `pkg/` -- Public Packages

- `pkg/version/`: Build-time version constants via ldflags
- `pkg/models/`: Shared data structures (ProjectConfig, UserConfig, LanguageConfig, QualityConfig, SpecDocument)
- `pkg/utils/`: Logger (slog), file I/O helpers, path resolution, timeout utilities, input validation

#### `cmd/moai/` -- Entry Point

Thin main.go that calls `cli.Execute()`. All logic lives in `internal/` and `pkg/`.

### Estimated LOC by Phase

| Phase | Modules | Estimated LOC |
|-------|---------|---------------|
| Phase 1: Foundation | config, hook, template, manifest | ~4,000 |
| Phase 2: Core Domains | git, quality, lsp, project | ~5,000 |
| Phase 3: Automation | loop, ralph, update, merge | ~3,000 |
| Phase 4: UI and Integration | ui, statusline, astgrep, rank | ~3,500 |
| Phase 5: Knowledge and CLI | foundation, integration, migration, cli, pkg | ~2,500 |
| **Total** | **22 modules** | **~18,000** |

This represents a 75% reduction from the Python codebase (73,000 LOC).

---

## 3. Complete Interface Catalog

### 3.1 Config Module (`internal/config/`)

```go
// ConfigManager provides thread-safe configuration management.
type ConfigManager interface {
    // Load reads configuration from the project root's .moai/config/sections/ directory.
    Load(projectRoot string) (*Config, error)

    // Get returns the current in-memory configuration. Thread-safe via RWMutex.
    Get() *Config

    // GetSection returns a named configuration section (user, language, quality, etc.).
    GetSection(name string) (interface{}, error)

    // SetSection updates a named configuration section in memory and persists to disk.
    SetSection(name string, value interface{}) error

    // Save persists the current configuration to disk atomically (temp + rename).
    Save() error

    // Watch registers a callback invoked when configuration files change on disk.
    Watch(callback func(Config)) error

    // Reload forces a re-read from disk, acquiring write lock.
    Reload() error
}

// Config is the root configuration aggregate containing all sections.
type Config struct {
    User        UserConfig        `yaml:"user"`
    Language    LanguageConfig    `yaml:"language"`
    Project     ProjectConfig     `yaml:"project"`
    Quality     QualityConfig     `yaml:"quality"`
    GitStrategy GitStrategyConfig `yaml:"git_strategy"`
    System      SystemConfig      `yaml:"system"`
    LLM         LLMConfig         `yaml:"llm"`
    Pricing     PricingConfig     `yaml:"pricing"`
    Ralph       RalphConfig       `yaml:"ralph"`
    Workflow    WorkflowConfig    `yaml:"workflow"`
}

// UserConfig represents the user identity section.
type UserConfig struct {
    Name string `yaml:"name" validate:"required"`
}

// LanguageConfig represents language preferences for all output channels.
type LanguageConfig struct {
    ConversationLanguage     string `yaml:"conversation_language" default:"en"`
    ConversationLanguageName string `yaml:"conversation_language_name" default:"English"`
    AgentPromptLanguage      string `yaml:"agent_prompt_language" default:"en"`
    GitCommitMessages        string `yaml:"git_commit_messages" default:"en"`
    CodeComments             string `yaml:"code_comments" default:"en"`
    Documentation            string `yaml:"documentation" default:"en"`
    ErrorMessages            string `yaml:"error_messages" default:"en"`
}

// QualityConfig represents TRUST 5 quality gate configuration.
type QualityConfig struct {
    DevelopmentMode    string           `yaml:"development_mode" default:"hybrid"`
    EnforceQuality     bool             `yaml:"enforce_quality" default:"true"`
    TestCoverageTarget int              `yaml:"test_coverage_target" default:"85"`
    DDDSettings        DDDSettings      `yaml:"ddd_settings"`
    LSPQualityGates    LSPQualityGates  `yaml:"lsp_quality_gates"`
}

// DDDSettings configures the DDD development methodology.
type DDDSettings struct {
    RequireExistingTests   bool `yaml:"require_existing_tests" default:"true"`
    CharacterizationTests  bool `yaml:"characterization_tests" default:"true"`
    BehaviorSnapshots      bool `yaml:"behavior_snapshots" default:"true"`
    MaxTransformationSize  string `yaml:"max_transformation_size" default:"small"`
}

// LSPQualityGates configures per-phase LSP diagnostic thresholds.
type LSPQualityGates struct {
    Enabled         bool            `yaml:"enabled" default:"true"`
    Plan            PlanGate        `yaml:"plan"`
    Run             RunGate         `yaml:"run"`
    Sync            SyncGate        `yaml:"sync"`
    CacheTTLSeconds int             `yaml:"cache_ttl_seconds" default:"5"`
    TimeoutSeconds  int             `yaml:"timeout_seconds" default:"3"`
}

type PlanGate struct {
    RequireBaseline bool `yaml:"require_baseline" default:"true"`
}

type RunGate struct {
    MaxErrors      int  `yaml:"max_errors" default:"0"`
    MaxTypeErrors  int  `yaml:"max_type_errors" default:"0"`
    MaxLintErrors  int  `yaml:"max_lint_errors" default:"0"`
    AllowRegression bool `yaml:"allow_regression" default:"false"`
}

type SyncGate struct {
    MaxErrors       int  `yaml:"max_errors" default:"0"`
    MaxWarnings     int  `yaml:"max_warnings" default:"10"`
    RequireCleanLSP bool `yaml:"require_clean_lsp" default:"true"`
}

// GitStrategyConfig configures Git workflow behavior.
type GitStrategyConfig struct {
    AutoBranch     bool   `yaml:"auto_branch" default:"false"`
    BranchPrefix   string `yaml:"branch_prefix" default:"moai/"`
    CommitStyle    string `yaml:"commit_style" default:"conventional"`
    WorktreeRoot   string `yaml:"worktree_root"`
}

// SystemConfig holds system-level settings.
type SystemConfig struct {
    Version     string `yaml:"version"`
    LogLevel    string `yaml:"log_level" default:"info"`
    LogFormat   string `yaml:"log_format" default:"text"`
    NoColor     bool   `yaml:"no_color" default:"false"`
    NonInteractive bool `yaml:"non_interactive" default:"false"`
}

// LLMConfig configures language model preferences.
type LLMConfig struct {
    DefaultModel   string `yaml:"default_model" default:"sonnet"`
    QualityModel   string `yaml:"quality_model" default:"opus"`
    SpeedModel     string `yaml:"speed_model" default:"haiku"`
}

// PricingConfig holds pricing and budget settings.
type PricingConfig struct {
    TokenBudget    int    `yaml:"token_budget" default:"250000"`
    CostTracking   bool   `yaml:"cost_tracking" default:"false"`
}

// RalphConfig configures the Ralph feedback loop engine.
type RalphConfig struct {
    MaxIterations  int  `yaml:"max_iterations" default:"5"`
    AutoConverge   bool `yaml:"auto_converge" default:"true"`
    HumanReview    bool `yaml:"human_review" default:"true"`
}

// WorkflowConfig configures the Plan-Run-Sync workflow.
type WorkflowConfig struct {
    AutoClear      bool `yaml:"auto_clear" default:"true"`
    PlanTokens     int  `yaml:"plan_tokens" default:"30000"`
    RunTokens      int  `yaml:"run_tokens" default:"180000"`
    SyncTokens     int  `yaml:"sync_tokens" default:"40000"`
}
```

### 3.2 Hook Module (`internal/hook/`)

```go
// Handler processes a single hook event type.
type Handler interface {
    // Handle executes the hook logic for a given input.
    Handle(ctx context.Context, input *HookInput) (*HookOutput, error)

    // EventType returns which event this handler processes.
    EventType() EventType
}

// Registry manages handler registration and event dispatch.
type Registry interface {
    // Register adds a handler to the registry.
    Register(handler Handler)

    // Dispatch routes an event to all registered handlers for that event type.
    Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error)

    // Handlers returns all handlers registered for a given event type.
    Handlers(event EventType) []Handler
}

// Protocol handles Claude Code's JSON stdin/stdout communication.
type Protocol interface {
    // ReadInput parses JSON hook input from a reader (typically os.Stdin).
    ReadInput(r io.Reader) (*HookInput, error)

    // WriteOutput serializes a hook output as JSON to a writer (typically os.Stdout).
    WriteOutput(w io.Writer, output *HookOutput) error
}

// Contract defines the formal Hook Execution Contract (ADR-012).
type Contract interface {
    // Validate verifies that the execution environment meets contract guarantees.
    Validate(ctx context.Context) error

    // Guarantees returns the list of environment guarantees.
    Guarantees() []string

    // NonGuarantees returns the list of explicitly unsupported assumptions.
    NonGuarantees() []string
}

// EventType represents a Claude Code hook lifecycle event.
type EventType string

const (
    EventSessionStart EventType = "SessionStart"
    EventPreToolUse   EventType = "PreToolUse"
    EventPostToolUse  EventType = "PostToolUse"
    EventSessionEnd   EventType = "SessionEnd"
    EventStop         EventType = "Stop"
    EventPreCompact   EventType = "PreCompact"
)

// HookInput is the JSON payload received from Claude Code via stdin.
type HookInput struct {
    SessionID     string          `json:"session_id"`
    CWD           string          `json:"cwd"`
    HookEventName string          `json:"hook_event_name"`
    ToolName      string          `json:"tool_name,omitempty"`
    ToolInput     json.RawMessage `json:"tool_input,omitempty"`
    ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
    ProjectDir    string          `json:"project_dir"`
}

// HookOutput is the JSON payload written to stdout for Claude Code consumption.
type HookOutput struct {
    Decision string          `json:"decision,omitempty"` // "allow", "block", "skip"
    Reason   string          `json:"reason,omitempty"`
    Data     json.RawMessage `json:"data,omitempty"`
}
```

### 3.3 Template Module (`internal/template/`)

```go
// Deployer extracts and deploys templates from the embedded filesystem.
type Deployer interface {
    // Deploy extracts all templates to the project root, updating the manifest.
    Deploy(ctx context.Context, projectRoot string, manifest manifest.Manager) error

    // ExtractTemplate returns the contents of a single embedded template file.
    ExtractTemplate(name string) ([]byte, error)

    // ListTemplates returns all template file paths in the embedded filesystem.
    ListTemplates() []string
}

// Renderer processes Go text/template files with strict mode.
type Renderer interface {
    // Render processes a named template with the given data context.
    // Uses template.Option("missingkey=error") for strict mode.
    Render(templateName string, data interface{}) ([]byte, error)
}

// SettingsGenerator produces platform-aware settings.json via struct serialization.
type SettingsGenerator interface {
    // Generate creates a settings.json byte payload for the given platform.
    // Uses json.MarshalIndent -- never string concatenation (ADR-011).
    Generate(cfg *config.Config, platform string) ([]byte, error)
}

// Validator performs post-deployment integrity checks.
type Validator interface {
    // ValidateJSON checks that data is valid JSON via json.Valid.
    ValidateJSON(data []byte) error

    // ValidatePaths verifies all paths use correct separators and pass containment checks.
    ValidatePaths(projectRoot string, files []string) []PathError

    // ValidateDeployment runs a full integrity check on deployed templates.
    ValidateDeployment(projectRoot string) *ValidationReport
}

// PathError represents a path validation failure.
type PathError struct {
    Path    string
    Message string
}

// ValidationReport summarizes deployment validation results.
type ValidationReport struct {
    Valid       bool
    Errors      []PathError
    Warnings    []string
    FilesChecked int
}

// Settings mirrors the Claude Code settings.json schema as Go structs.
type Settings struct {
    Hooks       map[string][]HookGroup `json:"hooks,omitempty"`
    OutputStyle string                 `json:"output_style,omitempty"`
}

// HookGroup represents a matcher-hooks pair in settings.json.
type HookGroup struct {
    Matcher string `json:"matcher"`
    Hooks   []Hook `json:"hooks"`
}

// Hook represents a single hook entry in settings.json.
type Hook struct {
    Type    string `json:"type"`
    Command string `json:"command"`
    Timeout int    `json:"timeout,omitempty"`
    Async   bool   `json:"async,omitempty"`
}
```

### 3.4 Manifest Module (`internal/manifest/`)

```go
// Manager provides CRUD operations for the file provenance manifest.
type Manager interface {
    // Load reads .moai/manifest.json from the project root.
    Load(projectRoot string) (*Manifest, error)

    // Save persists the current manifest to disk atomically.
    Save() error

    // Track records or updates a file entry in the manifest.
    Track(path string, provenance Provenance, templateHash string) error

    // GetEntry returns the manifest entry for a path, if it exists.
    GetEntry(path string) (*FileEntry, bool)

    // DetectChanges compares current file hashes against manifest records.
    DetectChanges() ([]ChangedFile, error)

    // Remove deletes a file entry from the manifest.
    Remove(path string) error
}

// Provenance classifies a file's origin and ownership.
type Provenance string

const (
    // TemplateManaged: deployed from template, no user changes. Safe to overwrite.
    TemplateManaged Provenance = "template_managed"

    // UserModified: template base with user edits detected. Requires 3-way merge.
    UserModified Provenance = "user_modified"

    // UserCreated: user's own file, not from template. Never touch.
    UserCreated Provenance = "user_created"

    // Deprecated: removed from new template version. Notify user, keep file.
    Deprecated Provenance = "deprecated"
)

// Manifest is the root data structure persisted as .moai/manifest.json.
type Manifest struct {
    Version    string               `json:"version"`
    DeployedAt string               `json:"deployed_at"`
    Files      map[string]FileEntry `json:"files"`
}

// FileEntry records provenance and hash state for a single deployed file.
type FileEntry struct {
    Provenance   Provenance `json:"provenance"`
    TemplateHash string     `json:"template_hash"`
    DeployedHash string     `json:"deployed_hash"`
    CurrentHash  string     `json:"current_hash"`
}

// ChangedFile represents a file whose current hash differs from the manifest.
type ChangedFile struct {
    Path         string
    OldHash      string
    NewHash      string
    Provenance   Provenance
}
```

### 3.5 Merge Module (`internal/merge/`)

```go
// Engine performs 3-way merge operations (ADR-008).
type Engine interface {
    // ThreeWayMerge merges base, current, and updated byte slices.
    ThreeWayMerge(base, current, updated []byte) (*MergeResult, error)

    // MergeFile performs a context-aware merge on a file path, selecting
    // the appropriate strategy based on file type.
    MergeFile(ctx context.Context, path string, base, current, updated []byte) (*MergeResult, error)
}

// StrategySelector determines the merge strategy for a given file path.
type StrategySelector interface {
    // SelectStrategy returns the appropriate merge strategy for a file.
    SelectStrategy(path string) MergeStrategy
}

// MergeStrategy identifies the merge algorithm to apply.
type MergeStrategy string

const (
    // LineMerge: line-by-line diff merge (default for .md, .txt).
    LineMerge MergeStrategy = "line_merge"

    // YAMLDeep: deep merge preserving YAML structure.
    YAMLDeep MergeStrategy = "yaml_deep"

    // JSONMerge: JSON-aware merge preserving object structure.
    JSONMerge MergeStrategy = "json_merge"

    // SectionMerge: section-aware merge for CLAUDE.md (preserves user sections).
    SectionMerge MergeStrategy = "section_merge"

    // EntryMerge: entry-based merge for .gitignore (append new, keep existing).
    EntryMerge MergeStrategy = "entry_merge"

    // Overwrite: full replacement (for binary or non-mergeable files).
    Overwrite MergeStrategy = "overwrite"
)

// MergeResult holds the output of a merge operation.
type MergeResult struct {
    Content     []byte
    HasConflict bool
    Conflicts   []Conflict
    Strategy    MergeStrategy
}

// Conflict describes a specific merge conflict region.
type Conflict struct {
    StartLine int
    EndLine   int
    Base      string
    Current   string
    Updated   string
}
```

### 3.6 Update Module (`internal/update/`)

```go
// Checker queries GitHub Releases for version information.
type Checker interface {
    // CheckLatest fetches the latest release metadata from GitHub.
    CheckLatest(ctx context.Context) (*VersionInfo, error)

    // IsUpdateAvailable compares the current version against the latest release.
    IsUpdateAvailable(current string) (bool, *VersionInfo, error)
}

// Updater handles binary download and self-replacement.
type Updater interface {
    // Download fetches the platform binary to a temporary location.
    Download(ctx context.Context, version *VersionInfo) (string, error)

    // Replace performs atomic binary self-replacement (write temp, rename).
    Replace(ctx context.Context, newBinaryPath string) error
}

// Rollback provides backup and restore for safe binary updates.
type Rollback interface {
    // CreateBackup copies the current binary to a backup location.
    CreateBackup() (string, error)

    // Restore replaces the current binary with the backup.
    Restore(backupPath string) error
}

// Orchestrator coordinates the full update workflow.
type Orchestrator interface {
    // Update runs the complete update pipeline: check, download, merge, replace, verify.
    Update(ctx context.Context) (*UpdateResult, error)
}

// VersionInfo holds release metadata from GitHub Releases API.
type VersionInfo struct {
    Version  string    `json:"version"`
    URL      string    `json:"url"`
    Checksum string    `json:"checksum"`
    Date     time.Time `json:"date"`
}

// UpdateResult summarizes the outcome of an update operation.
type UpdateResult struct {
    PreviousVersion string
    NewVersion      string
    FilesUpdated    int
    FilesMerged     int
    FilesConflicted int
    FilesSkipped    int
    RollbackPath    string
}
```

### 3.7 LSP Module (`internal/lsp/`)

```go
// Client communicates with a single Language Server Protocol server.
type Client interface {
    // Initialize sends the initialize request with the project root URI.
    Initialize(ctx context.Context, rootURI string) error

    // Diagnostics returns current diagnostics for a document URI.
    Diagnostics(ctx context.Context, uri string) ([]Diagnostic, error)

    // References finds all references to the symbol at the given position.
    References(ctx context.Context, uri string, pos Position) ([]Location, error)

    // Hover returns hover information for the symbol at the given position.
    Hover(ctx context.Context, uri string, pos Position) (*HoverResult, error)

    // Definition returns the definition location(s) for the symbol at the given position.
    Definition(ctx context.Context, uri string, pos Position) ([]Location, error)

    // Symbols returns all document symbols for the given URI.
    Symbols(ctx context.Context, uri string) ([]DocumentSymbol, error)

    // Shutdown sends the shutdown request and cleans up the connection.
    Shutdown(ctx context.Context) error
}

// ServerManager manages the lifecycle of multiple language servers.
type ServerManager interface {
    // StartServer launches a language server for the given language identifier.
    StartServer(ctx context.Context, lang string) error

    // StopServer gracefully shuts down a language server.
    StopServer(ctx context.Context, lang string) error

    // GetClient returns the LSP client for a running language server.
    GetClient(lang string) (Client, error)

    // ActiveServers returns the list of currently running server language identifiers.
    ActiveServers() []string

    // HealthCheck returns the health status of all active servers.
    HealthCheck(ctx context.Context) map[string]error
}

// Diagnostic represents an LSP diagnostic message.
type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity"`
    Code     string             `json:"code,omitempty"`
    Source   string             `json:"source,omitempty"`
    Message  string             `json:"message"`
}

// DiagnosticSeverity levels per LSP specification.
type DiagnosticSeverity int

const (
    SeverityError   DiagnosticSeverity = 1
    SeverityWarning DiagnosticSeverity = 2
    SeverityInfo    DiagnosticSeverity = 3
    SeverityHint    DiagnosticSeverity = 4
)

// Position identifies a location within a text document.
type Position struct {
    Line      int `json:"line"`
    Character int `json:"character"`
}

// Range defines a span within a text document.
type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}

// Location identifies a position within a specific document.
type Location struct {
    URI   string `json:"uri"`
    Range Range  `json:"range"`
}

// HoverResult holds the response from a hover request.
type HoverResult struct {
    Contents string `json:"contents"`
    Range    *Range `json:"range,omitempty"`
}

// DocumentSymbol represents a symbol found in a document.
type DocumentSymbol struct {
    Name     string           `json:"name"`
    Kind     int              `json:"kind"`
    Range    Range            `json:"range"`
    Children []DocumentSymbol `json:"children,omitempty"`
}
```

### 3.8 Git Module (`internal/core/git/`)

```go
// Repository provides read operations on a Git repository.
type Repository interface {
    // CurrentBranch returns the name of the currently checked-out branch.
    CurrentBranch() (string, error)

    // Status returns the working tree status (staged, modified, untracked files).
    Status() (*GitStatus, error)

    // Log returns the last n commits from HEAD.
    Log(n int) ([]Commit, error)

    // Diff returns the diff output between two references.
    Diff(ref1, ref2 string) (string, error)

    // IsClean returns true if the working tree has no uncommitted changes.
    IsClean() (bool, error)

    // Root returns the absolute path to the repository root.
    Root() string
}

// BranchManager provides branch lifecycle operations.
type BranchManager interface {
    // Create creates a new branch from the current HEAD.
    Create(name string) error

    // Switch checks out an existing branch.
    Switch(name string) error

    // Delete removes a branch (local only).
    Delete(name string) error

    // List returns all local branches.
    List() ([]Branch, error)

    // HasConflicts detects whether merging the target branch would produce conflicts.
    HasConflicts(target string) (bool, error)

    // MergeBase returns the common ancestor commit of two branches.
    MergeBase(branch1, branch2 string) (string, error)
}

// WorktreeManager manages Git worktrees for parallel development.
type WorktreeManager interface {
    // Add creates a new worktree at the given path for the specified branch.
    Add(path, branch string) error

    // List returns all active worktrees.
    List() ([]Worktree, error)

    // Remove removes a worktree directory.
    Remove(path string) error

    // Prune cleans up stale worktree references.
    Prune() error
}

// GitStatus holds the working tree state.
type GitStatus struct {
    Staged    []string
    Modified  []string
    Untracked []string
    Ahead     int
    Behind    int
}

// Commit represents a Git commit record.
type Commit struct {
    Hash    string
    Author  string
    Date    time.Time
    Message string
}

// Branch represents a Git branch.
type Branch struct {
    Name     string
    IsRemote bool
    IsCurrent bool
}

// Worktree represents a Git worktree entry.
type Worktree struct {
    Path   string
    Branch string
    HEAD   string
}
```

### 3.9 Quality Module (`internal/core/quality/`)

```go
// Gate orchestrates TRUST 5 quality validation.
type Gate interface {
    // Validate runs all five quality principles and returns an aggregate report.
    Validate(ctx context.Context) (*Report, error)

    // ValidatePrinciple runs a single named quality principle.
    ValidatePrinciple(ctx context.Context, principle string) (*PrincipleResult, error)
}

// Report aggregates the results of all TRUST 5 quality principles.
type Report struct {
    Passed     bool                       `json:"passed"`
    Score      float64                    `json:"score"`
    Principles map[string]PrincipleResult `json:"principles"`
    Timestamp  time.Time                  `json:"timestamp"`
}

// PrincipleResult holds the validation result for a single TRUST 5 principle.
type PrincipleResult struct {
    Name   string  `json:"name"`
    Passed bool    `json:"passed"`
    Score  float64 `json:"score"`
    Issues []Issue `json:"issues"`
}

// Issue represents a single quality violation.
type Issue struct {
    File     string `json:"file"`
    Line     int    `json:"line"`
    Severity string `json:"severity"` // "error", "warning", "info"
    Message  string `json:"message"`
    Rule     string `json:"rule"`
}
```

### 3.10 Loop and Ralph Modules (`internal/loop/`, `internal/ralph/`)

```go
// Controller orchestrates the Ralph feedback loop lifecycle.
type Controller interface {
    // Start begins a new feedback loop for a given SPEC.
    Start(ctx context.Context, specID string) error

    // Pause suspends the current loop iteration.
    Pause() error

    // Resume continues a paused loop from the last saved state.
    Resume(ctx context.Context) error

    // Cancel terminates the loop and cleans up state.
    Cancel() error

    // Status returns the current loop state and progress.
    Status() *LoopStatus

    // RecordFeedback adds feedback from the latest iteration.
    RecordFeedback(feedback Feedback) error
}

// Storage persists loop state for session resumption.
type Storage interface {
    // SaveState persists the loop state to disk.
    SaveState(state *LoopState) error

    // LoadState reads the loop state for a given SPEC.
    LoadState(specID string) (*LoopState, error)

    // DeleteState removes the persisted state for a given SPEC.
    DeleteState(specID string) error
}

// FeedbackGenerator collects feedback from build, test, and lint results.
type FeedbackGenerator interface {
    // Collect gathers feedback from the current development iteration.
    Collect(ctx context.Context) (*Feedback, error)
}

// LoopPhase represents a stage in the feedback loop.
type LoopPhase string

const (
    PhaseAnalyze   LoopPhase = "analyze"
    PhaseImplement LoopPhase = "implement"
    PhaseTest      LoopPhase = "test"
    PhaseReview    LoopPhase = "review"
)

// LoopState captures the full state of a feedback loop for persistence.
type LoopState struct {
    SpecID    string     `json:"spec_id"`
    Phase     LoopPhase  `json:"phase"`
    Iteration int        `json:"iteration"`
    MaxIter   int        `json:"max_iterations"`
    Feedback  []Feedback `json:"feedback"`
    StartedAt time.Time  `json:"started_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}

// LoopStatus is a read-only snapshot of the current loop state.
type LoopStatus struct {
    SpecID    string
    Phase     LoopPhase
    Iteration int
    MaxIter   int
    Converged bool
    Running   bool
}

// Feedback holds the results of a single development iteration.
type Feedback struct {
    Phase        LoopPhase     `json:"phase"`
    Iteration    int           `json:"iteration"`
    TestsPassed  int           `json:"tests_passed"`
    TestsFailed  int           `json:"tests_failed"`
    LintErrors   int           `json:"lint_errors"`
    BuildSuccess bool          `json:"build_success"`
    Coverage     float64       `json:"coverage"`
    Duration     time.Duration `json:"duration"`
    Notes        string        `json:"notes"`
}

// DecisionEngine determines the next loop action based on state and feedback.
type DecisionEngine interface {
    // Decide evaluates the current state and latest feedback to produce a decision.
    Decide(ctx context.Context, state *LoopState, feedback *Feedback) (*Decision, error)
}

// Decision represents the engine's recommendation for the next loop action.
type Decision struct {
    Action    string    `json:"action"`
    NextPhase LoopPhase `json:"next_phase"`
    Converged bool      `json:"converged"`
    Reason    string    `json:"reason"`
}
```

### 3.11 AST-Grep Module (`internal/astgrep/`)

```go
// Analyzer provides structural code analysis via ast-grep.
type Analyzer interface {
    // Scan runs multiple patterns against multiple file paths.
    Scan(ctx context.Context, patterns []string, paths []string) (*ScanResult, error)

    // FindPattern searches for a single pattern in a given language.
    FindPattern(ctx context.Context, pattern string, lang string) ([]Match, error)

    // Replace applies a pattern-based replacement across files.
    Replace(ctx context.Context, pattern, replacement, lang string, paths []string) ([]FileChange, error)
}

// Match represents a single AST pattern match.
type Match struct {
    File   string `json:"file"`
    Line   int    `json:"line"`
    Column int    `json:"column"`
    Text   string `json:"text"`
    Rule   string `json:"rule"`
}

// ScanResult aggregates match results from a scan operation.
type ScanResult struct {
    Matches  []Match       `json:"matches"`
    Duration time.Duration `json:"duration"`
    Files    int           `json:"files_scanned"`
}

// FileChange records a replacement applied to a file.
type FileChange struct {
    Path        string
    Line        int
    OldText     string
    NewText     string
    Applied     bool
}
```

### 3.12 UI Module (`internal/ui/`)

```go
// Wizard runs the interactive project initialization flow.
type Wizard interface {
    // Run starts the init wizard and returns user selections.
    Run(ctx context.Context) (*WizardResult, error)
}

// WizardResult holds the user's choices from the init wizard.
type WizardResult struct {
    ProjectName string
    Language    string
    Framework   string
    Features    []string
    UserName    string
    ConvLang    string
}

// Selector provides single-selection with optional fuzzy search.
type Selector interface {
    // Select presents a list and returns the chosen item's value.
    Select(label string, items []SelectItem) (string, error)
}

// Checkbox provides multi-selection with optional search.
type Checkbox interface {
    // MultiSelect presents items with checkboxes and returns selected values.
    MultiSelect(label string, items []SelectItem) ([]string, error)
}

// SelectItem represents a selectable option.
type SelectItem struct {
    Label string
    Value string
    Desc  string
}

// Progress provides progress bars and spinners.
type Progress interface {
    // Start creates a determinate progress bar.
    Start(title string, total int) ProgressBar

    // Spinner creates an indeterminate spinner.
    Spinner(title string) Spinner
}

// ProgressBar is a determinate progress indicator.
type ProgressBar interface {
    Increment(n int)
    SetTitle(title string)
    Done()
}

// Spinner is an indeterminate progress indicator.
type Spinner interface {
    SetTitle(title string)
    Stop()
}

// NonInteractive provides headless mode support.
type NonInteractive interface {
    // SetDefaults pre-fills all wizard prompts with given values.
    SetDefaults(defaults map[string]string)

    // IsHeadless returns true if running without a TTY.
    IsHeadless() bool
}
```

### 3.13 Statusline Module (`internal/statusline/`)

```go
// Builder composes the statusline output from collected data.
type Builder interface {
    // Build generates the formatted statusline string.
    Build(ctx context.Context) (string, error)

    // SetMode switches between statusline display modes.
    SetMode(mode StatuslineMode)
}

// StatuslineMode controls the level of detail in the statusline.
type StatuslineMode string

const (
    ModeMinimal StatuslineMode = "minimal"
    ModeDefault StatuslineMode = "default"
    ModeVerbose StatuslineMode = "verbose"
)

// Collector gathers all data needed for statusline rendering.
type Collector interface {
    // Collect retrieves current status data from all sources.
    Collect(ctx context.Context) (*StatusData, error)
}

// StatusData aggregates all statusline data sources.
type StatusData struct {
    Git     GitStatusData     `json:"git"`
    Memory  MemoryData        `json:"memory"`
    Quality QualityData       `json:"quality"`
    Version VersionData       `json:"version"`
}

// GitStatusData holds Git information for the statusline.
type GitStatusData struct {
    Branch    string `json:"branch"`
    Modified  int    `json:"modified"`
    Staged    int    `json:"staged"`
    Untracked int    `json:"untracked"`
    Ahead     int    `json:"ahead"`
    Behind    int    `json:"behind"`
}

// MemoryData holds context window and token budget information.
type MemoryData struct {
    TokensUsed  int `json:"tokens_used"`
    TokenBudget int `json:"token_budget"`
}

// QualityData holds TRUST 5 compliance summary.
type QualityData struct {
    Score  float64 `json:"score"`
    Passed bool    `json:"passed"`
}

// VersionData holds ADK version and update availability.
type VersionData struct {
    Current       string `json:"current"`
    Latest        string `json:"latest"`
    UpdateAvailable bool `json:"update_available"`
}
```

### 3.14 Project Module (`internal/core/project/`)

```go
// Initializer handles project scaffolding and setup.
type Initializer interface {
    // Init creates a new MoAI project with the given options.
    Init(ctx context.Context, opts InitOptions) error
}

// InitOptions configures the project initialization.
type InitOptions struct {
    ProjectRoot string
    ProjectName string
    Language    string
    Framework   string
    Features    []string
    UserName    string
    ConvLang    string
    NonInteractive bool
}

// Detector identifies project characteristics from the filesystem.
type Detector interface {
    // DetectLanguages scans the project root and returns detected languages.
    DetectLanguages(root string) ([]Language, error)

    // DetectFrameworks scans for known framework configuration files.
    DetectFrameworks(root string) ([]Framework, error)

    // DetectProjectType classifies the project based on structure and files.
    DetectProjectType(root string) (ProjectType, error)
}

// Language represents a detected programming language.
type Language struct {
    Name       string
    Confidence float64
    FileCount  int
}

// Framework represents a detected development framework.
type Framework struct {
    Name    string
    Version string
    ConfigFile string
}

// ProjectType classifies a project (web-app, api, cli, library).
type ProjectType = models.ProjectType

// ProjectValidator checks project structure integrity.
type ProjectValidator interface {
    // Validate checks the overall project structure.
    Validate(root string) (*ValidationResult, error)

    // ValidateMoAI checks MoAI-specific configuration and file integrity.
    ValidateMoAI(root string) (*ValidationResult, error)
}

// ValidationResult holds project validation outcomes.
type ValidationResult struct {
    Valid    bool
    Errors   []string
    Warnings []string
}
```

### 3.15 Rank Module (`internal/rank/`)

```go
// RankClient communicates with the ranking API.
type RankClient interface {
    // Submit sends session metrics to the ranking service.
    Submit(ctx context.Context, metrics *SessionMetrics) error

    // Leaderboard fetches the current community leaderboard.
    Leaderboard(ctx context.Context) ([]RankEntry, error)

    // MyRank fetches the authenticated user's current rank.
    MyRank(ctx context.Context) (*RankEntry, error)
}

// SessionMetrics holds performance data for a development session.
type SessionMetrics struct {
    TokensUsed     int           `json:"tokens_used"`
    TasksCompleted int           `json:"tasks_completed"`
    QualityScore   float64       `json:"quality_score"`
    Duration       time.Duration `json:"duration"`
}

// RankEntry represents a single entry on the leaderboard.
type RankEntry struct {
    Rank       int     `json:"rank"`
    UserHash   string  `json:"user_hash"`
    Score      float64 `json:"score"`
    Sessions   int     `json:"sessions"`
}
```

### 3.16 Foundation Module (`internal/foundation/`)

```go
// EARSGenerator produces EARS-format requirement templates.
type EARSGenerator interface {
    // GenerateUbiquitous creates an always-active requirement.
    GenerateUbiquitous(description string) string

    // GenerateEventDriven creates a trigger-response requirement.
    GenerateEventDriven(trigger, response string) string

    // GenerateStateDriven creates a conditional behavior requirement.
    GenerateStateDriven(condition, behavior string) string

    // GenerateUnwanted creates a prohibition requirement.
    GenerateUnwanted(prohibition string) string

    // GenerateOptional creates a nice-to-have requirement.
    GenerateOptional(feature string) string
}

// LanguageRegistry provides language ecosystem metadata for 16+ languages.
type LanguageRegistry interface {
    // GetLanguage returns the definition for a language by identifier.
    GetLanguage(id string) (*LanguageDef, bool)

    // ListLanguages returns all supported language identifiers.
    ListLanguages() []string
}

// LanguageDef describes a supported programming language ecosystem.
type LanguageDef struct {
    ID            string
    Name          string
    LSPServer     string
    LSPArgs       []string
    FilePatterns  []string
    PackageFile   string
    Linter        string
    Formatter     string
}
```

### 3.17 Migration Module (`internal/core/migration/`)

```go
// Migrator handles configuration version upgrades.
type Migrator interface {
    // NeedsMigration checks if the current config version requires migration.
    NeedsMigration(currentVersion, targetVersion string) (bool, error)

    // Migrate upgrades configuration from currentVersion to targetVersion.
    Migrate(ctx context.Context, currentVersion, targetVersion string) error
}

// BackupManager creates and restores configuration backups.
type BackupManager interface {
    // CreateBackup saves the current .moai/ directory to a timestamped backup.
    CreateBackup(projectRoot string) (string, error)

    // RestoreBackup restores a .moai/ directory from a backup path.
    RestoreBackup(backupPath, projectRoot string) error

    // ListBackups returns available backup paths.
    ListBackups(projectRoot string) ([]string, error)
}
```

---

## 4. Dependency Graph

### Text-Based Dependency Diagram

```
Layer 0: Entry Point
======================
    cmd/moai/main.go
         |
         v
Layer 1: CLI (Composition Root)
=================================
    internal/cli/
    |    |    |    |    |    |    |    |
    v    v    v    v    v    v    v    v

Layer 2: Domain Modules
=========================
    internal/          internal/core/          internal/
    hook/              project/                update/
    config/            git/                    loop/
    template/          quality/                ralph/
    manifest/          integration/            merge/
    lsp/               migration/              ui/
    statusline/                                astgrep/
    rank/                                      foundation/

         |
         v
Layer 3: Public Packages
=========================
    pkg/version/
    pkg/models/
    pkg/utils/
```

### Detailed Import Direction

```
cmd/moai/ ---------> internal/cli/

internal/cli/ -----> internal/config/
                  -> internal/hook/
                  -> internal/core/project/
                  -> internal/core/git/
                  -> internal/core/quality/
                  -> internal/update/
                  -> internal/statusline/
                  -> internal/rank/
                  -> internal/ui/
                  -> pkg/version/
                  -> pkg/models/

internal/hook/ ----> internal/config/
                  -> internal/core/quality/
                  -> internal/lsp/
                  -> internal/loop/
                  -> pkg/models/
                  -> pkg/utils/

internal/update/ --> internal/manifest/
                  -> internal/merge/
                  -> internal/template/
                  -> internal/config/
                  -> internal/ui/
                  -> pkg/version/

internal/template/ -> internal/manifest/
                   -> internal/config/
                   -> pkg/utils/

internal/core/
  project/ --------> internal/config/
                  -> internal/core/git/
                  -> internal/template/
                  -> internal/foundation/
                  -> internal/ui/
                  -> pkg/models/

  quality/ --------> internal/lsp/
                  -> internal/astgrep/
                  -> internal/core/git/
                  -> internal/config/
                  -> pkg/models/

  git/ ------------> pkg/models/
                  -> pkg/utils/

  migration/ ------> internal/config/
                  -> pkg/version/

  integration/ ----> internal/core/quality/
                  -> internal/core/git/

internal/loop/ ----> internal/ralph/
                  -> internal/core/quality/
                  -> internal/core/git/
                  -> internal/config/
                  -> pkg/utils/

internal/ralph/ ---> internal/config/
                  -> pkg/models/

internal/statusline/ -> internal/core/git/
                     -> internal/config/
                     -> pkg/version/

internal/rank/ ----> internal/config/
                  -> internal/core/git/
                  -> pkg/utils/

internal/foundation/ -> pkg/models/

internal/lsp/ -----> pkg/models/
                  -> pkg/utils/

internal/astgrep/ -> pkg/models/
                  -> pkg/utils/

internal/manifest/ -> pkg/utils/

internal/merge/ ---> pkg/utils/

internal/ui/ ------> pkg/models/

pkg/version/ ------> (no internal imports)
pkg/models/ -------> (no internal imports)
pkg/utils/ --------> (no internal imports)
```

### Circular Dependency Prevention Rules

1. **cli/ imports domain packages; domain packages NEVER import cli/**: The CLI layer is the composition root. Domain packages are unaware of the CLI layer.

2. **Domain packages import only pkg/ and lower-layer internal packages**: A domain module may import other domain modules only if there is no reverse dependency. For example, `quality/` imports `lsp/`, but `lsp/` never imports `quality/`.

3. **pkg/ imports NOTHING from internal/**: Public packages are self-contained. They provide shared types and utilities used by all layers.

4. **foundation/ imports only pkg/models/**: The foundation module provides static knowledge (EARS patterns, language definitions) and depends on no other internal packages.

5. **No cross-imports within the same layer without clear ordering**: When two domain modules both need each other, extract a shared interface into pkg/models/ or introduce a mediator in a parent package.

---

## 5. Dependency Injection and Wiring

### Constructor Injection Pattern

Every module exposes a `New*()` constructor that accepts interface dependencies. No global state, no init() side effects for domain logic, no DI framework (wire, dig, fx).

```go
// internal/config/manager.go
func NewConfigManager() ConfigManager {
    return &configManager{
        mu: sync.RWMutex{},
    }
}

// internal/hook/registry.go
func NewRegistry(
    cfg config.ConfigManager,
    quality quality.Gate,
    lspMgr lsp.ServerManager,
) Registry {
    return &registry{
        cfg:     cfg,
        quality: quality,
        lspMgr:  lspMgr,
        handlers: make(map[EventType][]Handler),
    }
}

// internal/core/quality/trust.go
func NewGate(
    lspMgr lsp.ServerManager,
    analyzer astgrep.Analyzer,
    repo git.Repository,
    cfg config.ConfigManager,
) Gate {
    return &gate{
        lspMgr:   lspMgr,
        analyzer: analyzer,
        repo:     repo,
        cfg:      cfg,
    }
}

// internal/update/orchestrator.go
func NewOrchestrator(
    checker Checker,
    updater Updater,
    rollback Rollback,
    manifest manifest.Manager,
    merge merge.Engine,
    deployer template.Deployer,
    ui ui.Progress,
) Orchestrator {
    return &orchestrator{
        checker:  checker,
        updater:  updater,
        rollback: rollback,
        manifest: manifest,
        merge:    merge,
        deployer: deployer,
        ui:       ui,
    }
}
```

### Wiring in main.go

All dependency construction happens in `cmd/moai/main.go`. This is the only place where concrete types are instantiated and wired together.

```go
// cmd/moai/main.go -- dependency wiring (composition root)
package main

import (
    "os"

    "github.com/modu-ai/moai-adk-go/internal/astgrep"
    "github.com/modu-ai/moai-adk-go/internal/cli"
    "github.com/modu-ai/moai-adk-go/internal/config"
    gitpkg "github.com/modu-ai/moai-adk-go/internal/core/git"
    "github.com/modu-ai/moai-adk-go/internal/core/quality"
    "github.com/modu-ai/moai-adk-go/internal/hook"
    "github.com/modu-ai/moai-adk-go/internal/lsp"
    "github.com/modu-ai/moai-adk-go/internal/manifest"
    "github.com/modu-ai/moai-adk-go/internal/merge"
    "github.com/modu-ai/moai-adk-go/internal/template"
    "github.com/modu-ai/moai-adk-go/internal/ui"
    "github.com/modu-ai/moai-adk-go/internal/update"
    "github.com/modu-ai/moai-adk-go/pkg/utils"
)

func main() {
    // Initialize logger
    utils.InitLogger()

    // Layer 0: Foundation (no dependencies)
    cfgMgr := config.NewConfigManager()
    manifestMgr := manifest.NewManager()
    mergeEngine := merge.NewEngine()
    uiProgress := ui.NewProgress()

    // Layer 1: Infrastructure
    lspMgr := lsp.NewServerManager(cfgMgr)
    repo := gitpkg.NewRepository()
    analyzer := astgrep.NewAnalyzer()
    deployer := template.NewDeployer(manifestMgr)

    // Layer 2: Domain
    qualityGate := quality.NewGate(lspMgr, analyzer, repo, cfgMgr)
    hookRegistry := hook.NewRegistry(cfgMgr, qualityGate, lspMgr)

    // Layer 3: Orchestration
    checker := update.NewChecker()
    updater := update.NewUpdater()
    rollback := update.NewRollback()
    updateOrch := update.NewOrchestrator(
        checker, updater, rollback,
        manifestMgr, mergeEngine, deployer, uiProgress,
    )

    // Wire CLI with all dependencies
    if err := cli.Execute(cli.Dependencies{
        Config:   cfgMgr,
        Hook:     hookRegistry,
        Quality:  qualityGate,
        Update:   updateOrch,
        Git:      repo,
        LSP:      lspMgr,
        UI:       uiProgress,
        Manifest: manifestMgr,
        Template: deployer,
    }); err != nil {
        os.Exit(1)
    }
}
```

### Design Rationale

- **No DI framework**: Go's explicit construction makes DI frameworks unnecessary. Constructor injection is the idiomatic Go approach.
- **Single wiring point**: All dependency graphs are visible in one file, making the system easy to understand and debug.
- **Lazy initialization**: Heavy resources (LSP servers, Git repository) are initialized lazily on first use within their constructors, not at startup.
- **Interface boundaries**: Consumers depend on interfaces only, enabling test mocks via mockery.

---

## 6. Error Handling Patterns

### Sentinel Errors

Each module defines package-level sentinel errors for expected failure conditions.

```go
// internal/config/
var (
    ErrConfigNotFound    = errors.New("config: configuration directory not found")
    ErrConfigInvalid     = errors.New("config: validation failed")
    ErrSectionNotFound   = errors.New("config: section not found")
    ErrSectionReadOnly   = errors.New("config: section is read-only")
)

// internal/hook/
var (
    ErrHookTimeout       = errors.New("hook: execution timed out")
    ErrHookContractFail  = errors.New("hook: execution contract violated")
    ErrHookInvalidInput  = errors.New("hook: invalid JSON input")
    ErrHookBlocked       = errors.New("hook: action blocked by hook")
)

// internal/manifest/
var (
    ErrManifestNotFound  = errors.New("manifest: file not found")
    ErrManifestCorrupt   = errors.New("manifest: JSON parse error")
    ErrEntryNotFound     = errors.New("manifest: entry not found")
)

// internal/lsp/
var (
    ErrServerNotRunning  = errors.New("lsp: server not running")
    ErrServerStartFailed = errors.New("lsp: server failed to start")
    ErrInitializeFailed  = errors.New("lsp: initialization failed")
    ErrConnectionClosed  = errors.New("lsp: connection closed")
)

// internal/merge/
var (
    ErrMergeConflict     = errors.New("merge: unresolvable conflict detected")
    ErrMergeUnsupported  = errors.New("merge: file type not supported for merge")
)

// internal/update/
var (
    ErrUpdateNotAvail    = errors.New("update: no update available")
    ErrDownloadFailed    = errors.New("update: binary download failed")
    ErrChecksumMismatch  = errors.New("update: checksum verification failed")
    ErrReplaceFailed     = errors.New("update: binary replacement failed")
)
```

### Context Wrapping

All errors are wrapped with context using `fmt.Errorf` and the `%w` verb.

```go
func (m *configManager) Load(projectRoot string) (*Config, error) {
    path := filepath.Join(projectRoot, ".moai", "config", "sections")
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return nil, fmt.Errorf("loading config from %s: %w", path, ErrConfigNotFound)
    }

    data, err := os.ReadFile(filepath.Join(path, "user.yaml"))
    if err != nil {
        return nil, fmt.Errorf("reading user config: %w", err)
    }

    var userCfg UserConfig
    if err := yaml.Unmarshal(data, &userCfg); err != nil {
        return nil, fmt.Errorf("parsing user config: %w", err)
    }
    // ...
    return cfg, nil
}
```

### Custom Error Types

Complex validation failures use structured error types.

```go
// internal/config/validation.go
type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation: field %q: %s", e.Field, e.Message)
}

// internal/template/validator.go
type DeploymentError struct {
    Path    string `json:"path"`
    Reason  string `json:"reason"`
    Details string `json:"details,omitempty"`
}

func (e *DeploymentError) Error() string {
    return fmt.Sprintf("deployment: %s: %s", e.Path, e.Reason)
}
```

### Error Handling Rules

1. **Never `panic()` for recoverable errors**: All functions return `error` as the last return value. Panics are reserved for programmer errors (nil pointer dereference in test setup, impossible states).

2. **Always wrap with context**: Raw `return err` is prohibited. Use `fmt.Errorf("context: %w", err)` to build an error chain.

3. **Sentinel errors enable `errors.Is()`**: Callers check for specific conditions via `errors.Is(err, ErrConfigNotFound)`.

4. **Custom types enable `errors.As()`**: Callers extract structured details via `errors.As(err, &validationErr)`.

5. **Log at boundaries**: Errors are logged at the CLI layer (the outermost boundary), not in domain code. Domain functions return errors for the caller to handle.

---

## 7. Concurrency Model

### Goroutine Usage by Module

| Module | Concurrency Pattern | Mechanism | Timeout |
|--------|-------------------|-----------|---------|
| **lsp/** | Parallel server startup | `errgroup.Group` | 500ms per server |
| **lsp/** | Concurrent diagnostic collection | `errgroup.Group` with limit | 2s total |
| **quality/** | Parallel principle validation | `errgroup.Group` (5 goroutines) | 5s total |
| **hook/** | Hook execution with timeout | `context.WithTimeout` | 30s default |
| **config/** | Concurrent read/write access | `sync.RWMutex` | N/A |
| **update/** | Background version checking | Single goroutine + channel | 2s |
| **statusline/** | Parallel data collection | `errgroup.Group` (4 collectors) | 1s |
| **git/** | Parallel worktree operations | `errgroup.Group` | 10s per op |
| **astgrep/** | Parallel file scanning | Worker pool via channels | 5s total |

### LSP Concurrency

```go
// internal/lsp/server.go -- parallel server startup
func (m *serverManager) StartAll(ctx context.Context, languages []string) error {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(4) // Max 4 concurrent server startups

    for _, lang := range languages {
        lang := lang // capture loop variable
        g.Go(func() error {
            ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
            defer cancel()
            return m.StartServer(ctx, lang)
        })
    }
    return g.Wait()
}

// Concurrent diagnostic collection across 16+ servers
func (m *serverManager) CollectAllDiagnostics(ctx context.Context, uri string) ([]Diagnostic, error) {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    var mu sync.Mutex
    var allDiags []Diagnostic

    g, ctx := errgroup.WithContext(ctx)
    for _, lang := range m.ActiveServers() {
        client, err := m.GetClient(lang)
        if err != nil {
            continue
        }
        g.Go(func() error {
            diags, err := client.Diagnostics(ctx, uri)
            if err != nil {
                return nil // Non-fatal: log and continue
            }
            mu.Lock()
            allDiags = append(allDiags, diags...)
            mu.Unlock()
            return nil
        })
    }
    _ = g.Wait()
    return allDiags, nil
}
```

### Quality Gate Concurrency

```go
// internal/core/quality/trust.go -- parallel principle validation
func (g *gate) Validate(ctx context.Context) (*Report, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    principles := []string{"tested", "readable", "unified", "secured", "trackable"}
    results := make(map[string]PrincipleResult, len(principles))
    var mu sync.Mutex

    eg, ctx := errgroup.WithContext(ctx)
    for _, p := range principles {
        p := p
        eg.Go(func() error {
            result, err := g.ValidatePrinciple(ctx, p)
            if err != nil {
                return fmt.Errorf("validating %s: %w", p, err)
            }
            mu.Lock()
            results[p] = *result
            mu.Unlock()
            return nil
        })
    }

    if err := eg.Wait(); err != nil {
        return nil, err
    }
    // Aggregate results...
    return &Report{Principles: results}, nil
}
```

### Config Thread Safety

```go
// internal/config/manager.go
type configManager struct {
    mu     sync.RWMutex
    config *Config
    root   string
}

func (m *configManager) Get() *Config {
    m.mu.RLock()
    defer m.mu.RUnlock()
    // Return a copy to prevent external mutation
    cfg := *m.config
    return &cfg
}

func (m *configManager) SetSection(name string, value interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    // ... update section ...
    return m.save()
}
```

### Hook Timeout

```go
// internal/hook/registry.go
func (r *registry) Dispatch(ctx context.Context, event EventType, input *HookInput) (*HookOutput, error) {
    timeout := 30 * time.Second
    if cfg := r.cfg.Get(); cfg.System.HookTimeout > 0 {
        timeout = time.Duration(cfg.System.HookTimeout) * time.Second
    }

    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    handlers := r.Handlers(event)
    for _, h := range handlers {
        output, err := h.Handle(ctx, input)
        if err != nil {
            if errors.Is(err, context.DeadlineExceeded) {
                return nil, fmt.Errorf("hook %s: %w", event, ErrHookTimeout)
            }
            return nil, err
        }
        if output != nil && output.Decision == "block" {
            return output, nil
        }
    }
    return &HookOutput{Decision: "allow"}, nil
}
```

---

## 8. Testing Strategy

### Per-Module Test Approach

| Module | Primary Test Type | Mock Targets | Coverage Goal |
|--------|------------------|--------------|---------------|
| config | Unit + Fuzz | filesystem (os.DirFS) | 90% |
| hook | Contract + Unit | stdin/stdout, exec.Command | 95% |
| template | Unit + JSON Safety | filesystem, embed.FS | 95% |
| manifest | Unit | filesystem | 90% |
| merge | Unit + Property | filesystem | 90% |
| lsp | Integration + Unit | JSON-RPC connection | 85% |
| core/git | Integration + Unit | go-git Repository | 85% |
| core/quality | Unit | lsp.Client, astgrep.Analyzer | 90% |
| core/project | Unit + Integration | config, git, template | 85% |
| loop | Unit + State Machine | ralph, quality, storage | 90% |
| ralph | Unit | config | 85% |
| update | Integration + Unit | HTTP client, filesystem | 85% |
| ui | Snapshot + Unit | bubbletea model | 70% |
| statusline | Unit | collectors | 85% |
| astgrep | Integration | ast-grep CLI | 80% |
| rank | Unit + Integration | HTTP client | 80% |
| foundation | Unit | (pure logic, no mocks) | 95% |
| core/migration | Unit | config, filesystem | 90% |
| core/integration | Integration | quality, git | 80% |
| pkg/version | Unit | (pure logic) | 100% |
| pkg/models | Unit | (pure structs) | 100% |
| pkg/utils | Unit + Fuzz | filesystem, environment | 95% |

### Test Conventions

```go
// Table-driven tests (idiomatic Go pattern)
func TestConfigManager_GetSection(t *testing.T) {
    tests := []struct {
        name    string
        section string
        want    interface{}
        wantErr error
    }{
        {
            name:    "valid user section",
            section: "user",
            want:    &UserConfig{Name: "GOOS"},
            wantErr: nil,
        },
        {
            name:    "unknown section",
            section: "nonexistent",
            want:    nil,
            wantErr: ErrSectionNotFound,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            mgr := setupTestConfig(t) // test helper
            got, err := mgr.GetSection(tt.section)
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
        })
    }
}
```

### Hook Contract Tests (ADR-012)

```go
// internal/hook/contract_test.go
func TestContract_MinimalPATH(t *testing.T) {
    // Verify hooks work with minimal PATH (simulating Claude Code environment)
    binary := buildTestBinary(t)
    cmd := exec.Command(binary, "hook", "session-start")
    cmd.Env = []string{
        "PATH=/usr/bin:/bin",
        "HOME=" + t.TempDir(),
    }
    cmd.Stdin = strings.NewReader(`{"session_id":"test","cwd":"/tmp"}`)

    out, err := cmd.Output()
    require.NoError(t, err)
    assert.True(t, json.Valid(out))
}

func TestContract_JSONRoundTrip(t *testing.T) {
    // Verify settings.json generation produces valid JSON
    cfg := testConfig()
    gen := template.NewSettingsGenerator()

    data, err := gen.Generate(cfg, runtime.GOOS)
    require.NoError(t, err)
    assert.True(t, json.Valid(data))

    // Round-trip: marshal -> unmarshal -> re-marshal produces identical output
    var parsed template.Settings
    require.NoError(t, json.Unmarshal(data, &parsed))
    reData, err := json.MarshalIndent(parsed, "", "  ")
    require.NoError(t, err)
    assert.JSONEq(t, string(data), string(reData))
}

func TestContract_NonInteractiveShell(t *testing.T) {
    // Verify hooks work without .bashrc/.zshrc loaded
    binary := buildTestBinary(t)
    cmd := exec.Command(binary, "hook", "post-tool")
    cmd.Env = []string{
        "PATH=/usr/bin:/bin:" + filepath.Dir(binary),
    }
    // No SHELL, no HOME, no USER -- simulating non-interactive environment
    cmd.Stdin = strings.NewReader(`{
        "session_id": "test",
        "cwd": "/tmp",
        "hook_event_name": "PostToolUse",
        "tool_name": "Write"
    }`)

    out, err := cmd.CombinedOutput()
    require.NoError(t, err, "stderr: %s", out)
}
```

### Fuzz Tests

```go
// internal/config/manager_fuzz_test.go
func FuzzConfigParse(f *testing.F) {
    f.Add([]byte(`user:\n  name: test`))
    f.Add([]byte(`language:\n  conversation_language: ko`))
    f.Add([]byte(`{}`))
    f.Add([]byte(``))

    f.Fuzz(func(t *testing.T, data []byte) {
        var cfg Config
        // Should never panic, regardless of input
        _ = yaml.Unmarshal(data, &cfg)
    })
}

// pkg/utils/path_fuzz_test.go
func FuzzFindProjectRoot(f *testing.F) {
    f.Add("/home/user/project")
    f.Add("../../relative/path")
    f.Add("/")

    f.Fuzz(func(t *testing.T, path string) {
        // Should never panic
        _, _ = utils.FindProjectRoot()
    })
}
```

### Coverage Requirements

| Scope | Minimum | Enforcement |
|-------|---------|-------------|
| Overall project | 85% | CI gate: `go test -coverprofile=coverage.out ./...` |
| `internal/core/` | 90% | CI gate: module-specific coverage check |
| `internal/cli/` | 70% | CI gate: UI-heavy, relaxed threshold |
| `pkg/` | 95% | CI gate: public API must be thoroughly tested |
| Hook contract tests | 100% event coverage | CI gate: all 6 event types tested |
| JSON safety tests | 100% output coverage | CI gate: all generated JSON validated |

---

## 9. Security Architecture

### Hook Execution Contract (ADR-012)

The Hook Execution Contract is the primary security boundary between Claude Code and MoAI-ADK.

**Guarantees** (tested in CI on all 6 platforms):

| Guarantee | Implementation | Test |
|-----------|---------------|------|
| stdin is valid JSON | `json.Unmarshal` with error handling | `TestContract_JSONRoundTrip` |
| Exit codes: 0=allow, 2=block | Explicit exit code mapping | `TestContract_ExitCodes` |
| Timeout enforcement | `context.WithTimeout` (default 30s) | `TestContract_Timeout` |
| Config access via same binary | Shared `config.ConfigManager` | `TestContract_ConfigAccess` |
| Working directory is project root | `os.Getwd()` verification | `TestContract_WorkingDir` |

**Non-Guarantees** (explicitly documented):

| Non-Guarantee | Reason | Mitigation |
|---------------|--------|------------|
| User's PATH | Claude Code spawns with minimal env | Binary must be in system PATH |
| Shell environment (.bashrc) | Non-interactive execution | No shell-dependent features |
| Shell functions/aliases | No shell wrapper used | Direct binary invocation |
| Python/Node.js availability | Zero runtime dependencies | Compiled binary |

### Zero Runtime Template Expansion (ADR-011)

All configuration files are generated via Go struct serialization. No dynamic tokens exist at rest.

```go
// PROHIBITED: string concatenation for JSON
cmd := fmt.Sprintf(`{"command": "%s hook session-start"}`, binaryPath) // WRONG

// REQUIRED: struct serialization
hook := Hook{
    Type:    "command",
    Command: binaryPath + " hook session-start",
}
data, _ := json.MarshalIndent(hook, "", "  ") // CORRECT: valid JSON by construction
```

**Validation pipeline**:

1. `json.MarshalIndent()` generates JSON from Go structs
2. `json.Valid()` verifies output validity
3. `json.Unmarshal()` round-trip test in CI
4. No `${VAR}`, `{{VAR}}`, or `$()` patterns in output (regex check in CI)

### JSON Injection Prevention

All JSON output uses `json.Marshal()` exclusively. String interpolation into JSON is prohibited at the linter level.

```go
// gosec rule: G203 (template injection)
// Custom golangci-lint rule: no fmt.Sprintf with JSON-like patterns
```

### Path Normalization

All file paths pass through `filepath.Clean()` and directory containment checks before deployment.

```go
// pkg/utils/path.go
func SafePath(base, path string) (string, error) {
    cleaned := filepath.Clean(filepath.Join(base, path))
    if !strings.HasPrefix(cleaned, filepath.Clean(base)+string(os.PathSeparator)) {
        return "", fmt.Errorf("path traversal detected: %s escapes %s", path, base)
    }
    return cleaned, nil
}
```

### Credential Management

| Credential | Storage | Access |
|------------|---------|--------|
| Ranking API token | System keyring (macOS Keychain, Linux secret-service, Windows Credential Manager) | `internal/rank/auth.go` via `zalando/go-keyring` |
| Git credentials | System Git credential helper | go-git credential callback |
| LSP server tokens | Environment variables | `os.Getenv()` with validation |

### OWASP Compliance

- **gosec** enabled in golangci-lint (security-focused static analysis)
- **govulncheck** in CI pipeline (known vulnerability detection in dependency tree)
- Input validation on all CLI arguments (Cobra flag validation + custom validators)
- No shell wrapping, no PATH manipulation, no `os/exec` with shell expansion
- Subprocess execution (ast-grep) uses `exec.Command` with explicit arguments, never shell interpolation

### Supply Chain Security

- `go.sum`: Cryptographic checksums for all dependencies
- Minimal dependency count (prefer stdlib)
- License audit: All dependencies must use permissive licenses (MIT, Apache-2.0, BSD)
- `CGO_ENABLED=0`: No C dependencies, pure Go compilation

---

## 10. Implementation Phases and Estimated Scope

### Phase 1: Foundation (~4,000 LOC)

**Modules**: `internal/config/`, `internal/hook/`, `internal/template/`, `internal/manifest/`

**Deliverables**:

- Typed configuration management with Viper integration
- Thread-safe config access via sync.RWMutex
- Config validation via struct tags
- Legacy format migration
- Hook registry and dispatch system
- Claude Code JSON stdin/stdout protocol handler
- Hook execution contract with contract tests
- All 6 hook event handlers (session-start, pre-tool, post-tool, session-end, stop, compact)
- Template deployment from go:embed
- Platform-aware settings.json generation (ADR-011)
- Post-deployment validation (JSON validity, path normalization)
- File provenance manifest (.moai/manifest.json)
- SHA-256 file hashing for change detection

**Issues Resolved**: 28 hook + 25 config/init + 6 overwrite = 59 issues

**Testing Focus**: Hook contract tests, JSON safety tests, config fuzz tests

### Phase 2: Core Domains (~5,000 LOC)

**Modules**: `internal/core/git/`, `internal/core/quality/`, `internal/lsp/`, `internal/core/project/`

**Deliverables**:

- Git operations via go-git with system Git fallback
- Branch management with MoAI naming conventions
- Conflict detection and worktree management
- TRUST 5 quality gate framework (5 principles)
- Parallel quality validation
- LSP client with JSON-RPC 2.0 over stdio/TCP
- Multi-server lifecycle management (16+ languages)
- Concurrent diagnostic collection
- Project initialization with TTY detection
- Language and framework detection
- Project structure validation

**Issues Resolved**: 11 Windows (via cross-compile) + 7 performance = 18 issues

**Testing Focus**: LSP integration tests, Git worktree integration tests, quality gate unit tests

### Phase 3: Automation (~3,000 LOC)

**Modules**: `internal/loop/`, `internal/ralph/`, `internal/update/`, `internal/merge/`

**Deliverables**:

- Ralph feedback loop controller (state machine)
- Loop state persistence for session resumption
- Convergence detection and human-in-the-loop breakpoints
- Decision engine for loop iterations
- Self-update via GitHub Releases API
- Binary download with checksum verification
- Atomic binary self-replacement with rollback
- 3-way merge engine with per-filetype strategies
- Conflict detection and `.conflict` file generation

**Issues Resolved**: 15 update/migration + 4 package manager = 19 issues

**Testing Focus**: Merge property tests, update integration tests, state machine coverage

### Phase 4: UI and Integration (~3,500 LOC)

**Modules**: `internal/ui/`, `internal/statusline/`, `internal/astgrep/`, `internal/rank/`

**Deliverables**:

- Init wizard (bubbletea Elm-architecture model)
- Fuzzy single-select and multi-select components
- Progress bars and spinners (bubbles)
- MoAI color theme (lipgloss)
- Headless mode support (--non-interactive flag)
- Statusline builder with parallel data collection
- Git status, memory, quality, and version collectors
- AST-grep integration for structural code analysis
- Ranking API client with authentication
- Session metrics collection and submission

**Issues Resolved**: 8 TUI issues (#268, #249, #286)

**Testing Focus**: bubbletea model snapshot tests, statusline unit tests, ast-grep integration tests

### Phase 5: Knowledge and CLI (~2,500 LOC)

**Modules**: `internal/foundation/`, `internal/core/integration/`, `internal/core/migration/`, `internal/cli/`, `pkg/`

**Deliverables**:

- EARS requirement pattern templates
- Language ecosystem definitions (16+ languages)
- Domain architecture patterns (backend, frontend, database, testing, devops)
- TRUST 5 principle definitions and checklists
- Integration test execution engine
- Version migration orchestrator with backup/restore
- Full CLI wiring with all Cobra commands
- Public API packages (version, models, utils)

**Testing Focus**: Foundation unit tests (pure logic), CLI integration tests, migration rollback tests

### Total Estimated Scope

| Phase | LOC | Cumulative | Issues Resolved |
|-------|-----|-----------|----------------|
| Phase 1: Foundation | ~4,000 | ~4,000 | 59 |
| Phase 2: Core Domains | ~5,000 | ~9,000 | 77 |
| Phase 3: Automation | ~3,000 | ~12,000 | 96 |
| Phase 4: UI and Integration | ~3,500 | ~15,500 | 104 |
| Phase 5: Knowledge and CLI | ~2,500 | ~18,000 | 104+ |
| **Total** | **~18,000** | | **95% of 173 issues** |

Reduction from Python: 73,000 LOC to ~18,000 LOC (75% reduction).

---

## 11. Migration Path (Python to Go)

### Strategy: Incremental Replacement

The Go binary replaces the Python implementation incrementally, module by module, ensuring backward compatibility at each step. A separate repository (`moai-adk-go`) avoids the module path conflicts that caused the previous Go rewrite failure.

### Phase 1: Hybrid Mode (Hook Commands Only)

```
User installs Go binary alongside Python:
  pip install moai-adk       # Python (existing)
  brew install moai          # Go (new)

settings.json uses Go binary for hooks:
  "command": "moai hook session-start"   # Go binary (fast, reliable)

All other commands still use Python:
  moai-adk init              # Python
  moai-adk doctor            # Python
```

**Compatibility Requirements**:

- Go binary reads `.moai/config/sections/*.yaml` (same format as Python)
- Go binary reads `.claude/settings.json` (same schema)
- Go binary writes hook output in Claude Code's expected JSON format
- Python and Go can coexist in PATH without conflict (different binary names)

### Phase 2: Go Binary Handles All CLI Commands

```
Go binary replaces Python for all CLI commands:
  moai init                  # Go (bubbletea wizard)
  moai doctor                # Go (compiled diagnostics)
  moai status                # Go (compiled statusline)
  moai update                # Go (self-update, no pip)
  moai worktree new          # Go (go-git + system Git)
  moai hook <event>          # Go (already replaced in Phase 1)

Python package becomes optional:
  pip uninstall moai-adk     # No longer needed
```

**Compatibility Requirements**:

- Go binary generates `.moai/manifest.json` (new, additive)
- Go binary preserves all existing `.moai/` directory structures
- Go binary can read Python-generated config (no format changes)
- `moai update` replaces pip/uv/pipx entirely

### Phase 3: Full Go Replacement

```
Go binary is the sole MoAI-ADK distribution:
  brew install moai          # Primary installation
  go install github.com/...  # Alternative for Go developers
  moai update                # Self-update via GitHub Releases

Python package deprecated:
  PyPI package archived
  Documentation updated to reference Go binary exclusively
```

**Compatibility Requirements**:

- Full feature parity with Python MoAI-ADK
- All 16+ languages supported by LSP integration
- Statusline rendering parity
- Ralph feedback loop parity
- Ranking system parity

### Coexistence Strategy

The Go binary is designed to read and write the same `.moai/` directory structure as the Python predecessor:

| File/Directory | Read by Go | Written by Go | Format Change |
|---------------|-----------|--------------|---------------|
| `.moai/config/sections/*.yaml` | Yes | Yes | None (same YAML format) |
| `.moai/manifest.json` | Yes | Yes | New file (additive) |
| `.moai/memory/*.json` | Yes | Yes | None |
| `.moai/specs/SPEC-*/*.md` | Yes | No (agent responsibility) | None |
| `.claude/settings.json` | Yes | Yes | Hook commands change from script paths to binary subcommands |
| `.claude/agents/moai/*.md` | Yes | Yes (deploy) | None |
| `.claude/skills/**/*.md` | Yes | Yes (deploy) | None |
| `.claude/commands/moai/*.md` | Yes | Yes (deploy) | None |
| `.claude/rules/moai/**/*.md` | Yes | Yes (deploy) | None |

### Rollback Strategy

If the Go binary introduces regressions:

1. Users can revert to Python by reinstalling: `pip install moai-adk`
2. The `.moai/` directory remains compatible (no destructive format changes)
3. `settings.json` hook commands can be manually reverted to Python script paths
4. The `.moai/manifest.json` file (Go-only) is ignored by the Python implementation

### Migration Validation Checklist

- [ ] Go binary reads all Python-generated `.moai/config/sections/*.yaml` files correctly
- [ ] Go binary generates valid `.claude/settings.json` that Claude Code accepts
- [ ] Hook events produce identical functional outcomes (allow/block decisions)
- [ ] Template deployment produces identical file structure
- [ ] CLI commands accept the same flags and arguments
- [ ] Statusline output is visually consistent
- [ ] Worktree operations interoperate with Python-created worktrees
- [ ] Config migration handles all legacy JSON format variants

---

## Appendix A: ADR Cross-Reference Index

| ADR | Title | Document | Section |
|-----|-------|----------|---------|
| ADR-001 | Modular Monolithic over Microservices | structure.md | Architecture Decisions |
| ADR-002 | internal/ for Domain Encapsulation | structure.md | Architecture Decisions |
| ADR-003 | Interface-Based Domain Boundaries | structure.md, tech.md | Architecture Decisions |
| ADR-004 | Embedded Templates via go:embed | structure.md, tech.md | Architecture Decisions |
| ADR-005 | Structured Logging via log/slog | structure.md, tech.md | Architecture Decisions |
| ADR-006 | Hooks as Binary Subcommands | structure.md | New Modules |
| ADR-007 | File Manifest for Provenance Tracking | structure.md | New Modules |
| ADR-008 | 3-Way Merge for Template Updates | structure.md, tech.md | Architecture Decisions |
| ADR-009 | Self-Update via Binary Replacement | structure.md | New Modules |
| ADR-010 | Charmbracelet for Terminal UI | structure.md, tech.md | Architecture Decisions |
| ADR-011 | Zero Runtime Template Expansion | structure.md | Architecture Decisions |
| ADR-012 | Hook Execution Contract | structure.md | Architecture Decisions |

## Appendix B: External Dependency Matrix

| Package | Version | License | Purpose | Justification |
|---------|---------|---------|---------|---------------|
| spf13/cobra | v1.10+ | Apache-2.0 | CLI framework | De facto standard for Go CLIs |
| spf13/viper | v1.18+ | MIT | Config management | Integrates with Cobra, supports YAML, env vars |
| go-git/go-git | v5.12+ | Apache-2.0 | Git operations | Pure Go, no CGO, portable |
| yaml.v3 | v3.0+ | MIT | YAML parsing | Standard Go YAML library |
| charmbracelet/bubbletea | v1.2+ | MIT | TUI framework | Elm architecture, cross-platform |
| charmbracelet/lipgloss | v1.0+ | MIT | Terminal styling | CSS-like styling, no ANSI management |
| charmbracelet/bubbles | v0.20+ | MIT | TUI components | Spinner, progress, textinput, list |
| charmbracelet/huh | v0.6+ | MIT | Form framework | Multi-step form wizard |
| stretchr/testify | v1.9+ | MIT | Test assertions | Assert, require, mock helpers |
| go.lsp.dev/protocol | v0.12+ | BSD-3 | LSP types | Community-standard LSP type definitions |
| go.lsp.dev/jsonrpc2 | v0.10+ | BSD-3 | JSON-RPC transport | LSP communication layer |
| zalando/go-keyring | v0.2+ | MIT | System keyring | Secure credential storage |

All dependencies use permissive licenses (MIT, Apache-2.0, BSD). No GPL dependencies.

## Appendix C: Performance Budget

| Operation | Target | Measurement | Budget Allocation |
|-----------|--------|-------------|-------------------|
| Binary cold start | < 50ms | `time moai version` | Process startup only |
| Config load (cold) | < 10ms | Benchmark test | Viper + YAML parse |
| Config load (cached) | < 1ms | Benchmark test | RWMutex read lock |
| CLI command P95 | < 200ms | End-to-end benchmark | Full command execution |
| Hook execution | < 100ms | Contract test | JSON parse + logic + JSON output |
| LSP server startup | < 500ms | Integration test | Single server |
| LSP diagnostics (16 servers) | < 2s | Parallel benchmark | Concurrent collection |
| Quality gate (full TRUST 5) | < 5s | Integration benchmark | 5 principles in parallel |
| Binary size (stripped) | < 30MB | Build output | -ldflags "-s -w" |
| Memory (idle) | < 20MB | Runtime profiling | No LSP servers |
| Memory (peak, 16 LSP) | < 200MB | Load testing | All servers active |
| settings.json generation | < 5ms | Benchmark test | json.MarshalIndent |
| Manifest load | < 5ms | Benchmark test | JSON parse |
| 3-way merge (100KB file) | < 50ms | Benchmark test | Line-level diff |

## Appendix D: File Listing Summary

```
cmd/moai/
    main.go                           Entry point, dependency wiring

internal/cli/
    root.go                           Root Cobra command, global flags
    init.go                           moai init (bubbletea wizard)
    doctor.go                         moai doctor (diagnostics)
    status.go                         moai status (project overview)
    update.go                         moai update (self-update)
    hook.go                           moai hook <event> (dispatcher)
    switch.go                         moai switch (branch switching)
    rank.go                           moai rank (performance ranking)
    version.go                        moai version (build info)
    worktree/
        new.go list.go switch.go sync.go remove.go clean.go

internal/config/
    manager.go                        ConfigManager with sync.RWMutex
    types.go                          Config struct hierarchy
    defaults.go                       Compiled default values
    migration.go                      Legacy format migration
    validation.go                     Struct tag validation

internal/hook/
    registry.go                       Handler registration and dispatch
    protocol.go                       JSON stdin/stdout protocol
    contract.go                       Execution contract (ADR-012)
    session_start.go                  SessionStart handler
    pre_tool.go                       PreToolUse handler
    post_tool.go                      PostToolUse handler
    session_end.go                    SessionEnd handler
    stop.go                           Stop handler
    compact.go                        PreCompact handler

internal/template/
    deployer.go                       go:embed extraction
    renderer.go                       text/template strict mode
    settings.go                       settings.json generation (ADR-011)
    validator.go                      Post-deployment integrity checks

internal/manifest/
    manifest.go                       Manifest CRUD operations
    hasher.go                         SHA-256 file hashing
    types.go                          Provenance enum, data structures

internal/merge/
    three_way.go                      3-way merge algorithm
    strategies.go                     Per-filetype merge strategies
    conflict.go                       Conflict detection and reporting
    differ.go                         Line-level diff generation

internal/update/
    checker.go                        GitHub Releases version check
    updater.go                        Binary download and replacement
    rollback.go                       Atomic rollback on failure
    orchestrator.go                   Full update workflow

internal/core/
    git/
        manager.go branch.go conflict.go event.go
    quality/
        trust.go validators.go
    project/
        initializer.go detector.go validator.go phase.go
    integration/
        engine.go models.go
    migration/
        migrator.go backup.go

internal/lsp/
    client.go server.go protocol.go models.go

internal/loop/
    controller.go feedback.go state.go storage.go

internal/ralph/
    engine.go

internal/ui/
    wizard.go selector.go checkbox.go progress.go theme.go prompt.go

internal/statusline/
    builder.go git.go metrics.go memory.go renderer.go update.go

internal/astgrep/
    analyzer.go models.go rules.go

internal/rank/
    client.go auth.go config.go hook.go

internal/foundation/
    ears.go langs.go backend.go frontend.go database.go
    testing.go devops.go
    trust/
        principles.go checklist.go

pkg/version/
    version.go

pkg/models/
    project.go spec.go config.go

pkg/utils/
    logger.go file.go path.go timeout.go validator.go

templates/                            (go:embed source)
    .claude/
        settings.json.tmpl
        agents/moai/
        skills/
        commands/moai/
        rules/moai/
        output-styles/
    .moai/config/sections/
    CLAUDE.md.tmpl
    .gitignore.tmpl
```

---

Document Version: 1.0.0
Created: 2026-02-03
Module Count: 22
Interface Count: 37
Estimated Total LOC: ~18,000
Python LOC Replaced: ~73,000
Reduction: 75%

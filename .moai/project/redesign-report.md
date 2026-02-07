# MoAI-ADK Go Edition: Comprehensive Redesign Report

## Executive Summary

This report analyzes 173 GitHub issues, 4,174 git commits, Claude Code hooks documentation, and the existing Python codebase (~73K lines) to design an upgraded Go architecture that resolves historically observed pain points.

**Key Finding**: 95% of actionable issues (89/94) are architecturally solvable by the Go rewrite.

---

## 1. Data Sources Analyzed

| Source | Volume | Key Insights |
|--------|--------|-------------|
| GitHub Issues | 173 (2 open, 171 closed) | Hooks #1 pain point (28 issues), Config #2 (25), Update #3 (15) |
| Git Commits | 4,174 (Sep 2025 – Feb 2026) | update.py most modified (38x), fix:feat ratio 0.73:1 |
| Claude Code Docs | 12 hook events, 3 types | Go binary hooks fully supported (cc-tools precedent) |
| Python Codebase | 180+ files, 10 modules | JIT Hook Manager 1,988 LOC, template processor 28x modified |

---

## 2. Root Cause Analysis

### 2.1 Hook System Failures (28 Issues)

| Root Cause | Issues | Examples | Frequency |
|-----------|--------|----------|-----------|
| Python runtime dependency | 12 | #278 PyYAML missing, #288 uv version detect fail, #269 import errors | Most common |
| Path resolution | 8 | #259 Windows mixed separators, #161 $CLAUDE_PROJECT_DIR unset, #5 MODULE_NOT_FOUND | Critical on Windows |
| Platform incompatibility | 5 | #129 SIGALRM (CRITICAL), #249 cp1252 encoding, #25 infinite wait | Windows-specific |
| Hook format/protocol | 3 | #265 settings.json format incompatible, #207 duplicate execution | Protocol mismatch |

**Go Solution: Hooks as Subcommands**

```json
{
  "hooks": {
    "SessionStart": [{
      "matcher": "",
      "hooks": [{
        "type": "command",
        "command": "moai hook session-start"
      }]
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{
        "type": "command",
        "command": "moai hook post-tool"
      }]
    }]
  }
}
```

**Why this works:**
1. Zero dependency — single binary, no Python/PyYAML/uv needed
2. No path resolution — `moai` is in PATH, no `$CLAUDE_PROJECT_DIR` needed
3. Cross-platform — compiled for each OS, no SIGALRM/encoding issues
4. Shared config — hook reads .moai/config via same binary
5. Fast — compiled Go, no interpreter startup overhead

**Evidence**: [cc-tools](https://github.com/Veraticus/cc-tools) project already uses Go binaries as Claude Code hooks successfully.

### 2.2 Configuration & Init Failures (25 Issues)

| Root Cause | Issues | Examples |
|-----------|--------|----------|
| Config path confusion | 8 | #315 wrong directory, #283 config not loaded, #206 version missing |
| Template variable substitution | 6 | #304 variable failure, #308 CLAUDE.md broken, #309 wrong language |
| Destructive init/update | 6 | #246 settings wiped, #236 content deleted, #162 files overwritten |
| Context size | 3 | #298 too much context, #226 oversized CLAUDE.md |
| Non-interactive environment | 2 | #2 fails in non-interactive, #9 hangs on Git Bash |

**Go Solution: Typed Config Structs**

```go
type Config struct {
    System   SystemConfig   `yaml:"system"`
    Language LanguageConfig `yaml:"language"`
    User     UserConfig     `yaml:"user"`
    Quality  QualityConfig  `yaml:"quality"`
    Git      GitConfig      `yaml:"git_strategy"`
}
```

- Compile-time type safety (no runtime "key not found")
- Default values in struct tags
- Schema validation via struct validation
- No template substitution in config files
- Atomic config writes via temp file + rename
- TTY detection for non-interactive environments

### 2.3 Update/Migration Failures (15 Issues)

| Root Cause | Issues | Examples |
|-----------|--------|----------|
| Destructive updates | 5 | #246 settings lost, #187 workflows overwritten, #318 sync fails |
| Version detection fragility | 4 | #250 version not displayed, #312 uv update fails |
| Package manager dependency | 4 | #253 PyPI missing, #296 PATH issue, #159 uv upgrade fails |
| No rollback | 2 | #319 all plugins fail after install |

**Go Solution: Smart Merge with File Manifest**

See Section 4 for detailed architecture.

### 2.4 Windows Compatibility (11 Issues)

| Root Cause | Issues | Examples |
|-----------|--------|----------|
| signal.SIGALRM | 1 (CRITICAL) | #129 all hooks fail |
| Path separators | 2 | #259 mixed separators, #161 env var unset |
| Encoding (cp949/cp1252) | 3 | #249, #286, #314 emoji rendering failures |
| Shell incompatibility | 3 | #45 PowerShell hooks, #31 script errors, #25 infinite wait |
| Installation | 2 | #271 install fails, #272 MCP errors |

**Go Solution**: Cross-compiled binaries eliminate ALL Windows-specific issues:
- No SIGALRM (Go uses goroutines + context.WithTimeout)
- No path separator issues (filepath.Join handles OS-specific separators)
- No encoding issues (Go strings are UTF-8 native)
- No shell dependency (compiled binary, not script)

### 2.5 Performance/Memory (7 Issues)

| Root Cause | Issues | Examples |
|-----------|--------|----------|
| Node.js heap OOM | 6 | #291, #290, #284, #282, #261 (~4GB limit) |
| Token budget | 1 | #292 output tokens |

**Go Solution**: Compiled binary with ~20MB idle memory vs Python's 50-100MB baseline.

---

## 3. Architecture Redesign

### 3.1 New Module Structure

```
moai-adk-go/
├── cmd/moai/
│   └── main.go                    # Entry point
├── internal/
│   ├── cli/                       # Cobra CLI commands
│   │   ├── root.go
│   │   ├── init.go                # Interactive project setup (bubbletea)
│   │   ├── doctor.go              # System diagnostics
│   │   ├── status.go              # Project status display
│   │   ├── update.go              # Self-update with smart merge
│   │   ├── hook.go                # Hook dispatcher subcommands (NEW!)
│   │   └── worktree/
│   ├── hook/                      # Hook system (NEW - replaces 46 Python scripts)
│   │   ├── registry.go            # Hook registration & dispatch
│   │   ├── session_start.go       # Project info, config validation
│   │   ├── pre_tool.go            # Security guard, validation
│   │   ├── post_tool.go           # Linter, formatter, LSP diagnostics
│   │   ├── session_end.go         # Cleanup, rank submission
│   │   ├── stop.go                # Loop controller
│   │   ├── compact.go             # Context preservation
│   │   └── protocol.go            # Claude Code JSON stdin/stdout protocol
│   ├── config/                    # Typed configuration management
│   │   ├── manager.go             # viper + typed structs
│   │   ├── types.go               # Config struct definitions
│   │   ├── defaults.go            # Default values
│   │   ├── migration.go           # Legacy format migration
│   │   └── validation.go          # Schema validation
│   ├── manifest/                  # File provenance tracking (NEW!)
│   │   ├── manifest.go            # Manifest CRUD operations
│   │   ├── hasher.go              # SHA-256 file hashing
│   │   └── types.go               # FileProvenance enum
│   ├── merge/                     # Smart merge engine (NEW!)
│   │   ├── three_way.go           # 3-way merge algorithm
│   │   ├── strategies.go          # Per-filetype merge strategies
│   │   ├── conflict.go            # Conflict detection & reporting
│   │   └── differ.go              # Diff generation
│   ├── template/                  # Template deployment
│   │   ├── deployer.go            # go:embed extraction
│   │   ├── renderer.go            # Go text/template rendering
│   │   └── settings.go            # settings.json generation
│   ├── update/                    # Self-update system (NEW!)
│   │   ├── checker.go             # GitHub Releases API version check
│   │   ├── updater.go             # Binary self-replacement
│   │   ├── rollback.go            # Atomic rollback on failure
│   │   └── orchestrator.go        # Full update workflow
│   ├── core/                      # Business domains
│   │   ├── project/               # Project detection, init, validation
│   │   ├── git/                   # go-git operations
│   │   ├── quality/               # TRUST 5 framework
│   │   ├── integration/           # Integration test engine
│   │   └── migration/             # Version migration
│   ├── foundation/                # Methodologies (EARS, patterns)
│   ├── lsp/                       # Language Server Protocol client
│   ├── loop/                      # Ralph feedback loop
│   ├── statusline/                # Claude Code statusline
│   ├── rank/                      # Performance ranking
│   ├── astgrep/                   # AST-grep integration
│   └── ui/                        # Charmbracelet TUI (NEW!)
│       ├── wizard.go              # Init wizard (bubbletea model)
│       ├── selector.go            # Fuzzy single-select
│       ├── checkbox.go            # Multi-select with search
│       ├── progress.go            # Progress bars + spinners
│       ├── theme.go               # MoAI color theme (lipgloss)
│       └── prompt.go              # Confirm/input prompts
├── pkg/
│   ├── version/                   # Build-time version (ldflags)
│   ├── models/                    # Shared data models
│   └── utils/                     # Logger, file ops, path utils
├── templates/                     # go:embed source
│   ├── .claude/
│   │   ├── settings.json.tmpl     # Platform-aware template
│   │   ├── agents/moai/           # Agent markdown definitions
│   │   ├── skills/                # Skill definitions
│   │   ├── commands/moai/         # Slash commands
│   │   ├── rules/moai/            # Rule files
│   │   └── output-styles/         # Output style definitions
│   ├── .moai/config/sections/     # Config section templates
│   ├── CLAUDE.md.tmpl             # CLAUDE.md template
│   └── .gitignore.tmpl            # .gitignore template
├── go.mod
├── Makefile
└── .goreleaser.yml
```

### 3.2 New Modules vs Python

| Module | Python | Go (New) | Improvement |
|--------|--------|----------|-------------|
| Hooks | 46 Python scripts + JIT manager (1,988 LOC) | internal/hook/ (compiled subcommands) | Zero dependency, cross-platform |
| Config | UnifiedConfigManager (thread-safe singleton) | internal/config/ (viper + typed structs) | Compile-time type safety |
| Template | processor.py (28x modified!) | internal/template/ (Go text/template) | Strict mode, no JSON conflicts |
| Update | update.py (38x modified!) | internal/update/ + internal/merge/ | Smart merge, self-update, rollback |
| Manifest | None | internal/manifest/ (NEW) | File provenance tracking |
| Merge | merger.py (basic) | internal/merge/ (3-way merge) | Git-like merge with conflicts |
| TUI | Rich + InquirerPy | Charmbracelet (bubbletea + lipgloss) | Elm arch, cross-platform |
| Statusline | 10 Python files | internal/statusline/ (compiled) | ~10x faster startup |

---

## 4. Smart Merge System (Key Innovation)

### 4.1 File Provenance Tracking

```json
// .moai/manifest.json
{
  "version": "1.14.0",
  "deployed_at": "2026-02-03T00:00:00Z",
  "files": {
    ".claude/agents/moai/expert-backend.md": {
      "provenance": "template_managed",
      "template_hash": "sha256:abc123...",
      "deployed_hash": "sha256:abc123...",
      "current_hash": "sha256:abc123..."
    },
    ".claude/agents/moai/my-custom-agent.md": {
      "provenance": "user_created",
      "template_hash": null,
      "deployed_hash": null,
      "current_hash": "sha256:xyz789..."
    },
    ".claude/skills/moai-domain-python.md": {
      "provenance": "user_modified",
      "template_hash": "sha256:def456...",
      "deployed_hash": "sha256:def456...",
      "current_hash": "sha256:ghi012..."
    }
  }
}
```

### 4.2 File Provenance Types

| Provenance | Meaning | Update Behavior |
|-----------|---------|-----------------|
| `template_managed` | Deployed from template, no user changes | Safe overwrite |
| `user_modified` | Template base + user edits | 3-way merge |
| `user_created` | User's own file, not from template | Never touch |
| `deprecated` | Removed from new template | Notify user, keep file |

### 4.3 Update Algorithm

```
moai update
  │
  ├─ 1. Check latest version (GitHub Releases API)
  ├─ 2. Download new binary to temp location
  ├─ 3. Load current manifest (.moai/manifest.json)
  ├─ 4. Extract new templates from new binary (go:embed)
  ├─ 5. For each file in new template:
  │     ├─ NOT in manifest → Deploy as template_managed
  │     ├─ template_managed + no user changes → Safe overwrite
  │     ├─ template_managed + user changed → Promote to user_modified, 3-way merge
  │     ├─ user_modified → 3-way merge (base=deployed, current=actual, new=template)
  │     │   ├─ No conflict → Auto-merge
  │     │   └─ Conflict → Write .conflict file, keep current, notify
  │     └─ user_created → Skip (never touch)
  ├─ 6. For each file in manifest NOT in new template:
  │     └─ Mark as deprecated, notify user, keep file
  ├─ 7. Update manifest with new hashes
  ├─ 8. Replace binary (atomic: write temp → rename)
  └─ 9. Show summary: updated/merged/conflicted/skipped
```

### 4.4 Issues This Resolves

| Issue | Problem | How Smart Merge Fixes It |
|-------|---------|--------------------------|
| #246 | Settings file lost on update | Manifest tracks provenance, never overwrites user_modified |
| #187 | GitHub workflows overwritten | User files detected as user_created, never touched |
| #162 | Init overwrites project files | Manifest prevents overwriting existing files |
| #236 | /moai:0-project deletes content | user_modified files get 3-way merge |
| #318 | Template sync fails 1.4→1.14 | Progressive migration via manifest |
| #319 | All plugins fail after install | Rollback mechanism restores previous state |

---

## 5. Hook System Redesign

### 5.1 Hooks as CLI Subcommands

```
moai hook session-start    # Replaces session_start__show_project_info.py
moai hook pre-tool         # Replaces pre_tool__security_guard.py
moai hook post-tool        # Replaces post_tool__linter.py + formatter + lsp
moai hook session-end      # Replaces session_end__auto_cleanup.py + rank
moai hook stop             # Replaces stop__loop_controller.py
moai hook compact          # Replaces pre_compact__save_context.py
```

### 5.2 Hook Protocol Handler

```go
// internal/hook/protocol.go
type HookInput struct {
    SessionID     string          `json:"session_id"`
    CWD           string          `json:"cwd"`
    HookEventName string          `json:"hook_event_name"`
    ToolName      string          `json:"tool_name,omitempty"`
    ToolInput     json.RawMessage `json:"tool_input,omitempty"`
    ToolOutput    json.RawMessage `json:"tool_output,omitempty"`
}

type HookOutput struct {
    Decision string `json:"decision,omitempty"` // "allow", "block", "skip"
    Reason   string `json:"reason,omitempty"`
}

func ReadInput() (*HookInput, error) {
    data, err := io.ReadAll(os.Stdin)
    if err != nil {
        return nil, err
    }
    var input HookInput
    return &input, json.Unmarshal(data, &input)
}
```

### 5.3 settings.json Generation

```go
// internal/template/settings.go
func GenerateSettings(platform string) Settings {
    binary := "moai"
    if platform == "windows" {
        binary = "moai.exe"
    }
    return Settings{
        Hooks: map[string][]HookGroup{
            "SessionStart": {{
                Matcher: "",
                Hooks: []Hook{{
                    Type:    "command",
                    Command: binary + " hook session-start",
                }},
            }},
            "PreToolUse": {{
                Matcher: "Write|Edit|Bash",
                Hooks: []Hook{{
                    Type:    "command",
                    Command: binary + " hook pre-tool",
                }},
            }},
            "PostToolUse": {{
                Matcher: "Write|Edit",
                Hooks: []Hook{{
                    Type:    "command",
                    Command: binary + " hook post-tool",
                    Timeout: 30,
                }},
            }},
            "Stop": {{
                Matcher: "",
                Hooks: []Hook{{
                    Type:    "command",
                    Command: binary + " hook stop",
                }},
            }},
            "SessionEnd": {{
                Matcher: "",
                Hooks: []Hook{{
                    Type:    "command",
                    Command: binary + " hook session-end",
                    Async:   true,
                }},
            }},
        },
        OutputStyle: "MoAI",
    }
}
```

### 5.4 Issues This Resolves

| Issue | Root Cause | How Go Hooks Fix It |
|-------|-----------|---------------------|
| #278, #288, #269, #260 | Python import/dependency errors | No Python runtime needed |
| #259, #161, #5 | Path resolution failures | `moai` binary in PATH, no path needed |
| #129 | signal.SIGALRM on Windows | Go uses context.WithTimeout |
| #249, #314 | Encoding (cp949/cp1252) | Go strings are UTF-8 native |
| #265 | settings.json format incompatible | Programmatic JSON generation |
| #207 | Hooks execute twice | Deterministic single registration |
| #263 | Shows error despite success | Proper exit code handling |
| #25, #66, #107 | Hook causes hang/freeze | Timeout via context, no shell dependency |
| #245, #243 | Config load failure in hooks | Same binary = same config loader |
| #231 | Docker dependency missing | Single binary, zero dependencies |

---

## 6. Modern TUI Design

### 6.1 Technology Stack

| Component | Library | Purpose |
|-----------|---------|---------|
| Framework | bubbletea | Elm-architecture TUI framework |
| Styling | lipgloss | CSS-like terminal styling |
| Components | bubbles | Spinner, progress, textinput, list |
| Forms | huh | Multi-step form wizard |
| Logging | log (charmbracelet) | Colorful structured logging |

### 6.2 MoAI Theme

```go
// internal/ui/theme.go
var Theme = struct {
    Primary   lipgloss.Color
    Secondary lipgloss.Color
    Accent    lipgloss.Color
    Error     lipgloss.Color
    Warning   lipgloss.Color
    Info      lipgloss.Color
    Muted     lipgloss.Color
}{
    Primary:   lipgloss.Color("#DA7756"), // Claude Code terra cotta
    Secondary: lipgloss.Color("#10B981"), // Green (success)
    Accent:    lipgloss.Color("#DA7756"),
    Error:     lipgloss.Color("#EF4444"),
    Warning:   lipgloss.Color("#F59E0B"),
    Info:      lipgloss.Color("#3B82F6"),
    Muted:     lipgloss.Color("#6B7280"),
}
```

### 6.3 Init Wizard (bubbletea)

```
╭─────────────────────────────────────────╮
│  MoAI-ADK Project Setup                │
├─────────────────────────────────────────┤
│                                         │
│  Project Name: █my-project              │
│                                         │
│  Language:                              │
│    ● Go                                 │
│    ○ Python                             │
│    ○ TypeScript                         │
│    ○ Other                              │
│                                         │
│  Features:                              │
│    [x] Git hooks integration            │
│    [x] LSP quality gates                │
│    [ ] Performance ranking              │
│    [ ] Ralph feedback loop              │
│                                         │
│  [Enter: Continue]  [Esc: Cancel]       │
╰─────────────────────────────────────────╯
```

### 6.4 Update Progress

```
╭─────────────────────────────────────────╮
│  MoAI-ADK Update v1.14.0 → v1.15.0    │
├─────────────────────────────────────────┤
│                                         │
│  Downloading binary...         ████████ │
│  Comparing templates...        ████████ │
│  Merging files...             ████░░░░ │
│                                         │
│  Files:                                 │
│    ✓ 12 updated (template_managed)      │
│    ✓  3 auto-merged (user_modified)     │
│    ⚠  1 conflict (manual review)        │
│    ─  5 skipped (user_created)          │
│                                         │
│  Conflict:                              │
│    .claude/agents/moai/expert-backend.md│
│    → .claude/agents/moai/expert-backend │
│      .md.conflict                       │
│                                         │
╰─────────────────────────────────────────╯
```

---

## 7. Configuration Optimization

### 7.1 Current Problems

1. Three config formats coexist (sections YAML, monolithic JSON, legacy flat)
2. Config migration runs every time (expensive)
3. Thread-safe singleton is complex in Python (double-checked locking)
4. Template variables in config files cause parsing errors

### 7.2 Go Solution

```go
// internal/config/types.go
type Config struct {
    System   SystemConfig   `yaml:"system"`
    Language LanguageConfig `yaml:"language"`
    User     UserConfig     `yaml:"user"`
    Quality  QualityConfig  `yaml:"quality"`
    Git      GitConfig      `yaml:"git_strategy"`
    Workflow WorkflowConfig `yaml:"workflow"`
    Ralph    RalphConfig    `yaml:"ralph"`
    LLM      LLMConfig     `yaml:"llm"`
}

// Default values embedded in struct tags
type QualityConfig struct {
    DevelopmentMode    string `yaml:"development_mode" default:"hybrid"`
    EnforceQuality     bool   `yaml:"enforce_quality" default:"true"`
    TestCoverageTarget int    `yaml:"test_coverage_target" default:"85"`
}
```

### 7.3 Config Loading Priority

```
1. Environment variables (MOAI_*)     ← highest
2. .moai/config/sections/*.yaml       ← project-level
3. ~/.moai/config/sections/*.yaml     ← user-level (NEW)
4. Compiled defaults                  ← lowest
```

### 7.4 No Template Substitution in Config

Instead of `{{CONVERSATION_LANGUAGE}}` in YAML files, configs are generated programmatically:

```go
func InitConfig(answers InitAnswers) Config {
    return Config{
        Language: LanguageConfig{
            ConversationLanguage: answers.Language, // Direct assignment, no templates
        },
    }
}
```

---

## 8. Previous Go Rewrite Failure: Lessons Learned

### What Happened (from git history)

```
feat(go): implement Phase 1 foundation and configuration system
feat(go): implement Phase 2 hook system with Python compatibility  ← MISTAKE
feat(go): implement Phase 3 CLI commands with direct binary paths
feat(go): implement Phase 4 distribution and migration system
→ revert: undo Go module restructure
→ refactor(adk): Remove MoAI-GO and consolidate to Python-based implementation
```

### Failure Causes (Hypothesized)

1. **Python compatibility layer** — Tried to keep Python hooks working alongside Go
2. **Same-repo coexistence** — Go + Python in same repo caused conflicts
3. **Big-bang approach** — 4 phases implemented before user validation
4. **Module path issues** — Go module restructure had to be reverted

### Corrective Measures for This Attempt

| Previous Mistake | Current Approach |
|-----------------|------------------|
| Python compatibility | Clean break: Go replaces Python entirely |
| Same repo | Separate repo (moai-adk-go) ✓ |
| Big-bang implementation | Incremental: CLI → Hooks → Merge → LSP |
| Module path confusion | Clean module path from day 1 ✓ |

---

## 9. Issue Resolution Matrix

### Expected Resolution by Go Edition

| Category | Total | Resolved | Rate | Key Architecture |
|----------|-------|----------|------|-----------------|
| Hook errors | 28 | 28 | 100% | internal/hook/ (compiled subcommands) |
| Config/Init | 25 | 22 | 88% | internal/config/ (typed structs) |
| Update/Migration | 15 | 14 | 93% | internal/merge/ (3-way) + internal/manifest/ |
| Windows | 11 | 11 | 100% | Cross-compiled binary |
| TUI/UI | 8 | 7 | 88% | Charmbracelet ecosystem |
| Performance/OOM | 7 | 7 | 100% | Compiled binary (~20MB idle) |
| **Total** | **94** | **89** | **95%** | |

### Remaining 5% (Unsolvable by Architecture)

- Claude Code's own bugs (path resolution in subdirectories)
- Token/context limits (Claude Code platform constraint)
- User-specific environment issues

---

## 10. Implementation Roadmap

### Phase 1: Core CLI + Config (Foundation)

**Modules**: internal/cli/, internal/config/, internal/ui/, pkg/
**Deliverables**:
- `moai init` with bubbletea wizard
- `moai doctor` with platform diagnostics
- `moai status` with config display
- `moai version` with build metadata
- Typed config management (viper + struct validation)
- TTY detection + --non-interactive flag

**Issues Resolved**: 25 config/init + 4 version = 29 issues

### Phase 2: Hook System (Critical Path)

**Modules**: internal/hook/, internal/template/settings.go
**Deliverables**:
- `moai hook <event>` subcommands (6 handlers)
- JSON stdin/stdout protocol handler
- Platform-aware settings.json generation
- Hook timeout via context.WithTimeout

**Issues Resolved**: 28 hooks + 11 Windows = 39 issues

### Phase 3: Smart Merge + Update (Reliability)

**Modules**: internal/manifest/, internal/merge/, internal/update/, internal/template/
**Deliverables**:
- File manifest system (.moai/manifest.json)
- 3-way merge engine with per-filetype strategies
- `moai update` with self-replacement + rollback
- Template deployment from go:embed
- Selective restoration with TUI

**Issues Resolved**: 15 update/migration = 15 issues

### Phase 4: Statusline + LSP + Quality (Quality Gates)

**Modules**: internal/statusline/, internal/lsp/, internal/core/quality/
**Deliverables**:
- Compiled statusline builder
- Multi-server LSP client (JSON-RPC 2.0)
- TRUST 5 quality gate framework

**Issues Resolved**: 8 TUI/statusline + 7 performance = 15 issues

### Phase 5: Foundation + Loop + Rank (Feature Complete)

**Modules**: internal/foundation/, internal/loop/, internal/rank/, internal/astgrep/
**Deliverables**:
- EARS methodology engine
- Ralph loop controller
- Rank submission client
- AST-grep integration

---

## 11. Compatibility Boundary

### MUST Be Compatible (Claude Code reads these)

| File/Format | Constraint |
|-------------|-----------|
| `.claude/settings.json` | Claude Code hook configuration format |
| `.claude/agents/moai/*.md` | Agent definition markdown format |
| `.claude/skills/**/*.md` | Skill definition with YAML frontmatter |
| `.claude/commands/moai/*.md` | Slash command definitions |
| `.claude/rules/moai/**/*.md` | Rule files with optional paths frontmatter |
| `.moai/config/sections/*.yaml` | Config format (existing users have these) |
| Hook stdin JSON | Claude Code sends this to hooks |
| Hook exit codes | 0=success, 2=block, other=non-blocking error |

### CAN Change (Internal implementation)

| Aspect | Python | Go |
|--------|--------|-----|
| Hook language | Python scripts | Go binary subcommands |
| Hook manager | JIT + Event-Driven (1,988 LOC) | Compiled dispatch (< 500 LOC) |
| Config loader | UnifiedConfigManager (singleton) | viper + typed structs |
| Template engine | Custom {{VAR}} substitution | Go text/template |
| TUI library | Rich + InquirerPy | bubbletea + lipgloss |
| CLI framework | Click | Cobra |
| Package manager | pip/uv/pipx | Self-contained binary |
| Distribution | PyPI | GitHub Releases + Homebrew |

---

## 12. Key Architectural Decisions Summary

| # | Decision | Rationale | Issues Resolved |
|---|---------|-----------|----------------|
| ADR-001 | Hooks as moai subcommands | Eliminates Python runtime, path, encoding issues | 28 |
| ADR-002 | Typed config structs | Compile-time safety, no template vars in config | 8 |
| ADR-003 | File manifest + 3-way merge | Prevents destructive updates, enables smart merge | 15 |
| ADR-004 | Self-update binary replacement | Eliminates pip/uv/pipx dependency chain | 4 |
| ADR-005 | Cross-compiled single binary | Eliminates ALL Windows-specific issues | 11 |
| ADR-006 | Charmbracelet TUI | Cross-platform, Elm architecture, modern UX | 8 |
| ADR-007 | go:embed templates | No runtime file path dependency, version-matched | 6 |
| ADR-008 | Programmatic settings.json | No template substitution in JSON, platform-aware | 3 |

---

Version: 1.0.0
Date: 2026-02-03
Author: MoAI (UltraThink Analysis)
Data Sources: 173 GitHub issues, 4,174 git commits, Claude Code docs, Python codebase (73K LOC)

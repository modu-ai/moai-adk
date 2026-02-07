# MoAI-GO Template, Hook & Configuration Redesign Plan

Version: 1.0.0
Date: 2026-02-03
Status: DRAFT
Cross-references: design.md (ADR-001 through ADR-012), structure.md, tech.md, redesign-report.md

---

## Overview

This document defines the implementation plan for five interconnected redesign efforts that transform MoAI-ADK from a Python-based multi-file ecosystem into a Go compiled single-binary architecture. Each part includes current state analysis, target design, file-level migration mappings, and measurable metrics.

The five parts are ordered by dependency: configuration foundation first, hook system second, directive optimization third, skills optimization fourth, and Go template embedding last.

---

## PART 1: HOOK SYSTEM REDESIGN (Python to Go Compiled Subcommands)

### 1.1 Current State

The Python hook system consists of 12 hook scripts and 20 library modules totaling 21,535 lines of code across 32 files. This system suffers from 28 documented GitHub issues spanning Python runtime dependencies, path resolution failures, platform incompatibilities, and hook protocol mismatches.

**Hook Scripts Inventory (12 scripts, 12,722 LOC)**

| Hook Script | LOC | Event | Key Function |
|-------------|-----|-------|-------------|
| `session_start__show_project_info.py` | 1,116 | SessionStart | Project status display, Git info, SPEC progress, config validation |
| `session_end__auto_cleanup.py` | 901 | SessionEnd | Resource cleanup, state persistence, temp file removal |
| `session_end__rank_submit.py` | 76 | SessionEnd | Performance metrics submission to ranking API |
| `stop__loop_controller.py` | 628 | Stop | Ralph loop state management, convergence detection |
| `pre_tool__security_guard.py` | 436 | PreToolUse | Tool permission validation, path containment checks |
| `post_tool__lsp_diagnostic.py` | 515 | PostToolUse | LSP type/lint error collection, diagnostic aggregation |
| `post_tool__code_formatter.py` | 262 | PostToolUse | Ruff/Black/Prettier auto-formatting on Write/Edit |
| `post_tool__linter.py` | 323 | PostToolUse | Lint error detection and reporting |
| `post_tool__ast_grep_scan.py` | 283 | PostToolUse | Structural code pattern scanning via ast-grep |
| `pre_compact__save_context.py` | 65 | PreCompact | Context state persistence before compaction |
| `quality_gate_with_lsp.py` | 296 | QualityGate | TRUST 5 validation with LSP diagnostic integration |
| `notification__statusline.py` | ~200 | Notification | Statusline update rendering |

**Library Modules Inventory (20 modules, 8,813 LOC)**

| Library Module | LOC | Purpose |
|---------------|-----|---------|
| `jit_enhanced_hook_manager.py` | 1,988 | Hook discovery, loading, registration, dispatch |
| `tool_registry.py` | 896 | Runtime tool permission discovery and caching |
| `project.py` | 786 | Project metadata, type detection, language detection |
| `version_reader.py` | 749 | Version file reading from multiple formats |
| `unified_timeout_manager.py` | 658 | Cross-platform timeout with SIGALRM fallback |
| `config.py` | 344 | Configuration file reading |
| `config_manager.py` | 300 | Configuration management and validation |
| `git_operations_manager.py` | 592 | Git operations wrapper (subprocess-based) |
| `language_validator.py` | 417 | Language code validation and normalization |
| `renderer.py` | 419 | Terminal output rendering with Rich |
| `enhanced_output_style_detector.py` | 372 | Output style detection from CLAUDE.md |
| `path_utils.py` | 180 | Path resolution and normalization |
| `file_utils.py` | 134 | File I/O helpers, atomic writes |
| `exceptions.py` | 171 | Custom exception hierarchy |
| `memory_collector.py` | 268 | Memory and token usage collection |
| `git_collector.py` | 190 | Git status data collection for statusline |
| `update_checker.py` | 129 | PyPI/GitHub version check |
| `checkpoint.py` | 100 | Session checkpoint management |
| `models.py` | 104 | Shared data models |
| `common.py` | ~100 | Common utilities |

### 1.2 Go Handler Design

Six compiled handlers replace 12 Python scripts. Each handler corresponds to a single Claude Code hook event type and internally dispatches to sub-handlers for each concern (formatting, linting, LSP diagnostics, etc.).

The settings.json entry format changes from Python script paths to binary subcommands:

```
"command": "moai hook session-start"
```

**Go Handler Mapping (6 handlers replacing 12 scripts)**

| Event | Go Handler File | Consolidates | Est. LOC |
|-------|----------------|-------------|----------|
| SessionStart | `internal/hook/session_start.go` | `session_start__show_project_info.py` | ~400 |
| PreToolUse | `internal/hook/pre_tool.go` | `pre_tool__security_guard.py` | ~200 |
| PostToolUse | `internal/hook/post_tool.go` | `post_tool__lsp_diagnostic.py` + `code_formatter.py` + `linter.py` + `ast_grep_scan.py` | ~600 |
| SessionEnd | `internal/hook/session_end.go` | `session_end__auto_cleanup.py` + `rank_submit.py` | ~300 |
| Stop | `internal/hook/stop.go` | `stop__loop_controller.py` | ~250 |
| PreCompact | `internal/hook/compact.go` | `pre_compact__save_context.py` | ~100 |

**Supporting Infrastructure Files**

| Go File | Purpose | Est. LOC |
|---------|---------|----------|
| `internal/hook/registry.go` | Handler registration and event dispatch | ~200 |
| `internal/hook/protocol.go` | Claude Code JSON stdin/stdout protocol | ~150 |
| `internal/hook/contract.go` | Hook execution contract (ADR-012) | ~150 |

### 1.3 Python Library Elimination

All 20 Python library modules are eliminated. Their functionality maps to Go stdlib, existing Go packages, or purpose-built Go modules.

| Python Module | LOC | Go Replacement | Rationale |
|--------------|-----|---------------|-----------|
| `jit_enhanced_hook_manager.py` | 1,988 | Eliminated | Compiled binary has no discovery/loading phase |
| `tool_registry.py` | 896 | Eliminated | Compiled binary; permissions in `internal/hook/pre_tool.go` |
| `project.py` | 786 | `internal/core/project/detector.go` | Project detection compiled into binary |
| `version_reader.py` | 749 | `pkg/version/version.go` | Build-time `ldflags` injection, no runtime file reading |
| `unified_timeout_manager.py` | 658 | `context.WithTimeout` (Go stdlib) | Go context cancellation replaces SIGALRM hack |
| `config.py` + `config_manager.py` | 644 | `internal/config/manager.go` | Viper + typed structs (ADR-011) |
| `git_operations_manager.py` | 592 | `internal/core/git/manager.go` | go-git in-process + system Git fallback |
| `language_validator.py` | 417 | `internal/config/validation.go` | Struct tag validation |
| `renderer.py` | 419 | `internal/statusline/renderer.go` | lipgloss terminal rendering |
| `enhanced_output_style_detector.py` | 372 | Eliminated | Output style embedded in CLAUDE.md, no runtime detection |
| `memory_collector.py` | 268 | `internal/statusline/memory.go` | Direct OS API calls |
| `git_collector.py` | 190 | `internal/statusline/git.go` | go-git in-process |
| `path_utils.py` | 180 | `pkg/utils/path.go` | `filepath.Clean()` + containment checks |
| `exceptions.py` | 171 | Module-level sentinel errors | `var ErrNotFound = errors.New(...)` |
| `file_utils.py` | 134 | `pkg/utils/file.go` | Atomic write via temp + rename |
| `update_checker.py` | 129 | `internal/update/checker.go` | GitHub Releases API |
| `moai_detector.py` | 105 | `internal/core/project/detector.go` | Absorbed into project detection |
| `models.py` | 104 | `pkg/models/` | Already exists in Go scaffold |
| `checkpoint.py` + `atomic_write.py` | 160 | `pkg/utils/file.go` | Go atomic file operations |
| `common.py` | ~100 | Absorbed into respective modules | No standalone common module needed |

### 1.4 settings.json Generation (ADR-011)

The settings.json file is generated programmatically via Go struct serialization. No template variables, no string concatenation, no runtime token expansion.

**Go Structs**

```go
// Settings mirrors the Claude Code settings.json schema.
type Settings struct {
    Hooks map[string][]HookGroup `json:"hooks,omitempty"`
}

// HookGroup pairs a tool matcher with its hook entries.
type HookGroup struct {
    Matcher string `json:"matcher"`
    Hooks   []Hook `json:"hooks"`
}

// Hook defines a single hook command entry.
type Hook struct {
    Type    string `json:"type"`
    Command string `json:"command"`
    Timeout int    `json:"timeout,omitempty"`
}
```

**Before (Python) -- settings.json hook entries**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__code_formatter.py\"'",
            "timeout": 30000
          },
          {
            "type": "command",
            "command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__linter.py\"'",
            "timeout": 15000
          },
          {
            "type": "command",
            "command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__lsp_diagnostic.py\"'",
            "timeout": 20000
          },
          {
            "type": "command",
            "command": "${SHELL:-/bin/bash} -c 'python3 \"$CLAUDE_PROJECT_DIR/.claude/hooks/post_tool__ast_grep_scan.py\"'",
            "timeout": 15000
          }
        ]
      }
    ]
  }
}
```

**After (Go) -- settings.json hook entries**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "moai hook post-tool",
            "timeout": 30000
          }
        ]
      }
    ]
  }
}
```

Key differences:
- 4 Python hook entries consolidated into 1 Go subcommand
- No `${SHELL:-/bin/bash}` shell wrapper (resolves 5 platform issues)
- No `$CLAUDE_PROJECT_DIR` path variable (resolves 8 path issues)
- No Python runtime dependency (resolves 12 import/PyYAML issues)
- Generated by `json.MarshalIndent()`, never string concatenation

### 1.5 Metrics

| Metric | Python (Before) | Go (After) | Change |
|--------|----------------|-----------|--------|
| Total LOC | 21,535 (32 files) | ~2,500 (9 files) | -88% |
| Hook scripts | 12 files | 0 files | -100% |
| Hook handlers | -- | 6 Go files | -- |
| Library modules | 20 files | 0 files | -100% |
| Infrastructure | -- | 3 Go files (registry, protocol, contract) | -- |
| settings.json entries per event | 1-4 Python commands | 1 Go subcommand | Up to -75% |
| Runtime dependencies | Python 3.13+, PyYAML, Rich, uv | None (compiled binary) | -100% |
| PATH issues | 8 documented | 0 (binary in PATH) | -100% |
| SIGALRM issues | 1 critical (#129) | 0 (`context.WithTimeout`) | -100% |
| Encoding issues | 3 documented | 0 (Go native UTF-8) | -100% |

---

## PART 2: SKILLS OPTIMIZATION

### 2.1 Three-Tier Classification

Skills are classified into three deployment tiers based on usage frequency and project relevance.

**Tier 1: CORE (Always Deployed) -- 10 skills**

These skills are deployed to every MoAI project regardless of detection results. They contain essential orchestration instructions.

| Skill | Current Tokens | Target Tokens | Action |
|-------|---------------|--------------|--------|
| `moai` | ~15,000 (body) | ~3,000 | Extract modules, trim redundancy, reference rules |
| `moai-foundation-core` | ~5,000 | ~2,000 | Remove code examples, delegate to Context7 for docs |
| `moai-foundation-claude` | ~5,000 | ~2,000 | Keep essential patterns only, move refs to Level 3 |
| `moai-foundation-quality` | ~5,000 | ~2,000 | Consolidate with `.claude/rules/moai/core/` rules |
| `moai-foundation-context` | ~3,000 | ~1,500 | Streamline to trigger keywords and core instructions |
| `moai-workflow-spec` | ~5,000 | ~2,000 | Focus on EARS templates, move migration guide to modules |
| `moai-workflow-ddd` | ~5,000 | ~2,000 | Keep ANALYZE-PRESERVE-IMPROVE cycle, trim examples |
| `moai-workflow-loop` | ~3,000 | ~1,500 | State machine definition only, examples to Level 3 |
| `moai-workflow-project` | ~5,000 | ~2,000 | Init wizard instructions, move docs management to modules |
| `moai-workflow-thinking` | ~3,000 | ~1,500 | Sequential Thinking MCP patterns, activation triggers |

**Tier 1 Total**: ~56,000 current tokens reduced to ~19,500 target tokens (65% reduction)

**Tier 2: PROJECT-SPECIFIC (Deployed by Detection) -- 35 skills**

Deployed selectively based on project language, framework, and platform detection results.

| Category | Count | Detection Method | Examples |
|----------|-------|-----------------|----------|
| Language skills | 16 | File extension analysis (`*.py`, `*.go`, `*.ts`) | `moai-lang-python`, `moai-lang-go`, `moai-lang-typescript` |
| Domain skills | 4 | Project type + framework detection | `moai-domain-backend`, `moai-domain-frontend`, `moai-domain-database`, `moai-domain-mobile` |
| Platform skills | 10 | Config files + dependency detection | `moai-platform-docker`, `moai-platform-vercel`, `moai-platform-aws` |
| Workflow/tool skills | 5 | Feature flag + config detection | `moai-workflow-worktree`, `moai-workflow-testing`, `moai-tool-ast-grep` |

Deployment rules:
- Language: Deploy only detected languages (e.g., Go project gets `moai-lang-go` only)
- Domain: Deploy based on `project.yaml` type field
- Platform: Deploy based on presence of config files (Dockerfile, vercel.json, etc.)
- Workflow: Deploy based on `.moai/config/` feature flags

**Tier 3: ON-DEMAND (Install via CLI) -- 11 skills**

Not deployed by default. Available via `moai skill install <name>`.

| Skill | Reason for On-Demand |
|-------|---------------------|
| `moai-platform-electron` | Niche desktop framework |
| `moai-platform-tauri` | Niche desktop framework |
| `moai-platform-cloudflare` | Platform-specific |
| `moai-platform-railway` | Platform-specific |
| `moai-platform-fly` | Platform-specific |
| `moai-tool-mermaid` | Specialized diagramming |
| `moai-tool-svg` | Specialized graphics |
| `moai-tool-pencil` | UI prototyping |
| `moai-foundation-philosopher` | Optional strategic framework |
| `moai-library-shadcn` | Framework-specific component library |
| `moai-library-nextra` | Framework-specific documentation |

### 2.2 Size Reduction Strategy

| Target | Current | After | Reduction | Method |
|--------|---------|-------|-----------|--------|
| `moai` SKILL.md | 15,777 lines | ~2,000 lines | 87% | Extract to `modules/`, trim redundant content |
| Language skills (avg) | ~1,000 lines | ~500 lines | 50% | Use Context7 MCP for live docs, keep only patterns |
| Domain skills (avg) | ~800 lines | ~400 lines | 50% | Reference Context7 for framework docs |
| Platform skills (avg) | ~600 lines | ~300 lines | 50% | Reference Context7 for platform docs |
| Overall disk size | 5.6 MB | ~1.5 MB | 73% | All reduction methods combined |

Primary reduction methods:
1. **Context7 delegation**: Remove embedded library documentation; reference via `context7-libraries` metadata
2. **Module extraction**: Move advanced content from SKILL.md body to `modules/` directory (Level 3)
3. **Example consolidation**: Move inline examples to `examples.md` files
4. **Redundancy elimination**: Remove content duplicated across skills or available in rules

### 2.3 Progressive Disclosure Enforcement

| Level | Max Tokens | Content | Loading Trigger |
|-------|-----------|---------|----------------|
| Level 1 | 50 | `name`, `description`, `triggers` (frontmatter only) | Always loaded for skills listed in agent definitions |
| Level 2 | 3,000 | Core instructions (SKILL.md body) | When trigger keyword or agent name matches |
| Level 3 | Unlimited | `modules/`, `reference.md`, `examples.md` | On-demand by Claude when deeper context needed |

Enforcement rules:
- SKILL.md body MUST NOT exceed 500 lines (soft limit) or 3,000 tokens (hard limit)
- Any content exceeding Level 2 budget moves to `modules/` directory
- Level 1 frontmatter MUST include `triggers.keywords` for accurate matching
- Skills without progressive disclosure configuration default to Level 2 full-load

### 2.4 Context7 Integration

Instead of embedding full library documentation in skill files, skills reference Context7 MCP for live documentation retrieval.

**Skill Frontmatter Pattern**

```yaml
metadata:
  context7-libraries: "react/react,vercel/next.js"
```

**Division of Responsibility**

| Concern | Skill Provides | Context7 Provides |
|---------|---------------|-------------------|
| Architecture patterns | Yes | No |
| Trigger keywords | Yes | No |
| Project conventions | Yes | No |
| API documentation | No | Yes |
| Code examples | Minimal (patterns only) | Yes (live, versioned) |
| Version-specific info | No | Yes |

**Impact on Token Budget**

| Scenario | Without Context7 | With Context7 | Savings |
|----------|-----------------|--------------|---------|
| React + Next.js project | ~8,000 tokens in skills | ~2,000 tokens + on-demand | 75% idle reduction |
| Python + FastAPI project | ~6,000 tokens in skills | ~1,500 tokens + on-demand | 75% idle reduction |
| Go project | ~4,000 tokens in skills | ~1,000 tokens + on-demand | 75% idle reduction |

---

## PART 3: .CLAUDE DIRECTIVES OPTIMIZATION

### 3.1 Agent Definitions Streamlining

**Current State**: 20 agent definitions averaging 729 lines each, totaling 14,578 lines. Most agents contain duplicated instructions, embedded examples, and verbose explanations that overlap with skill content.

**Target State**: 20 agent definitions averaging 150 lines each, totaling ~3,000 lines. Agents reference skills for detailed knowledge and follow a standardized template.

**Standardized Agent Template**

```markdown
---
name: {agent-name}
model: inherit
tools: [Read, Write, Edit, Grep, Glob, Bash]
---

# {agent-name}

## Identity

{2-3 sentences defining the agent's role, expertise, and primary responsibility.}

## Capabilities

- {Capability 1: What the agent can do}
- {Capability 2: What the agent can do}
- {Capability 3: What the agent can do}
- {Capability 4: What the agent can do}

## Loaded Skills

- Skill("{skill-name-1}"): {brief reason}
- Skill("{skill-name-2}"): {brief reason}

## Working Protocol

1. {Step 1: How the agent begins work}
2. {Step 2: Core execution phase}
3. {Step 3: Validation and quality checks}
4. {Step 4: Output and reporting}

## Quality Standards

- {Standard 1: Key quality requirement}
- {Standard 2: Key quality requirement}
- {Standard 3: Key quality requirement}
```

**Size Reduction for Top 5 Largest Agents**

| Agent | Before (lines) | After (lines) | Reduction |
|-------|---------------|--------------|-----------|
| `manager-git.md` | 1,206 | ~200 | 83% |
| `manager-spec.md` | 1,002 | ~180 | 82% |
| `manager-project.md` | 972 | ~170 | 83% |
| `expert-backend.md` | 964 | ~160 | 83% |
| `expert-frontend.md` | 864 | ~150 | 83% |

**Where Does the Content Go?**

| Content Type | Current Location | New Location |
|-------------|-----------------|-------------|
| Detailed methodology | Agent .md body | Skill `modules/` directory |
| Code examples | Agent .md body | Skill `examples.md` |
| Tool usage patterns | Agent .md body | Skill Level 2 body |
| Quality checklists | Agent .md body | `.claude/rules/moai/core/` |
| Language-specific guidance | Agent .md body | Language skills (Tier 2) |

### 3.2 Rules Consolidation

**Before**: 22 rule files deployed to every project.

**After**: 6 core rules (always deployed) + 16 conditional language rules (1 per detected language).

| Action | Files | Result |
|--------|-------|--------|
| MERGE | `moai-constitution.md` + `coding-standards.md` | `moai-core.md` (single core reference) |
| KEEP | `skill-authoring.md` | Unchanged (skill creation reference) |
| KEEP | `spec-workflow.md` | Unchanged (Plan-Run-Sync workflow) |
| KEEP | `workflow-modes.md` | Unchanged (phase definitions) |
| KEEP | `file-reading-optimization.md` | Unchanged (token efficiency) |
| NEW | `token-budget.md` | Extracted from multiple sources (budget allocation reference) |
| CONDITIONAL | 16 language rules (`go.md`, `python.md`, `typescript.md`, etc.) | Deploy 1 per detected language |

**Net Change**: 22 always-deployed files reduced to 6 always-deployed + 1-2 conditional = 7-8 total per project (64% reduction).

### 3.3 Output Styles

**Before**: 3 output style files deployed to `.claude/output-styles/` directory (72 KB total).

**After**: 0 output style files deployed.

| Change | Details |
|--------|---------|
| Default MoAI style | Embedded directly in CLAUDE.md template (no separate file) |
| Alternative styles (Yoda, R2D2) | Available via `moai style install yoda` CLI command |
| Deployment behavior | CLI command copies style file from embedded templates on demand |

This eliminates 3 files and 72 KB from every project initialization.

### 3.4 CLAUDE.md Template

**Before**: 310 lines with embedded content for all features.

**After**: ~200 lines with `@import` references to rules and focused core content.

**Structure**

| Section | Lines | Content |
|---------|-------|---------|
| Core Identity | 20 | MoAI name, version, orchestrator role |
| HARD Rules | 30 | Language, parallel execution, no XML, markdown output |
| Command Reference | 40 | `/moai` subcommands with brief descriptions |
| Agent Catalog | 40 | Agent names and one-line descriptions (details in agent .md files) |
| Quality Gates | 20 | TRUST 5 summary + reference to `moai-core.md` |
| Configuration | 20 | `@import` references to config and rule files |
| Error Handling | 15 | Recovery patterns |
| Progressive Disclosure | 15 | System overview |
| **Total** | **~200** | |

**CLAUDE.md.tmpl Template with Go Variables**

```go-template
# MoAI Execution Directive

## 1. Core Identity

MoAI is the Strategic Orchestrator for Claude Code. All tasks must be delegated to specialized agents.

Version: {{ .Version }}
Project: {{ .ProjectName }}

### HARD Rules (Mandatory)

- [HARD] Language-Aware Responses: All user-facing responses MUST be in {{ .ConversationLanguageName }}
- [HARD] Parallel Execution: Execute all independent tool calls in parallel when no dependencies exist
- [HARD] No XML in User Responses: Never display XML tags in user-facing responses
- [HARD] Markdown Output: Use Markdown for all user-facing communication

---

## 2. Command Reference

### Unified Skill: /moai

Subcommands: plan, run, sync, project, fix, loop, feedback

---

## 3. Agent Catalog

### Manager Agents
{{ range .ManagerAgents }}
- {{ .Name }}: {{ .Description }}
{{ end }}

### Expert Agents
{{ range .ExpertAgents }}
- {{ .Name }}: {{ .Description }}
{{ end }}

---

## 4. Quality Gates

@.claude/rules/moai/core/moai-core.md

---

## 5. Configuration Reference

User language: {{ .ConversationLanguageName }} ({{ .ConversationLanguage }})
Code comments: {{ .CodeComments }}

@.moai/config/sections/user.yaml
@.moai/config/sections/language.yaml

---

## 6. Language Rules

- User Responses: Always in {{ .ConversationLanguageName }}
- Internal Agent Communication: English
- Code Comments: {{ .CodeComments }}
```

### 3.5 settings.json

The settings.json file is NOT a template file. It is generated programmatically by `internal/template/settings.go` via Go struct serialization (ADR-011).

| Aspect | Implementation |
|--------|---------------|
| Generation method | `json.MarshalIndent(settings, "", "  ")` |
| Platform awareness | `runtime.GOOS` check: `moai` (macOS/Linux) vs `moai.exe` (Windows) |
| Template variables | None (zero runtime expansion) |
| Validation | `json.Valid()` check after generation |
| Hook entries | Generated from `internal/hook/registry.go` handler list |

---

## PART 4: .MOAI CONFIGURATION REDESIGN

### 4.1 Directory Structure Change

**Before**

```
.moai/
├── config/
│   ├── config.yaml                    # Root config (version, init status)
│   └── sections/
│       ├── user.yaml                  # User identity
│       ├── language.yaml              # Language preferences
│       ├── project.yaml               # Project metadata
│       ├── quality.yaml               # TRUST 5 config
│       ├── git-strategy.yaml          # Git workflow config
│       ├── system.yaml                # System settings
│       ├── llm.yaml                   # LLM model preferences
│       ├── pricing.yaml               # Token budget and costs
│       ├── ralph.yaml                 # Loop config
│       └── workflow.yaml              # Phase token budgets
├── project/                           # Design documents (5 files)
│   ├── product.md
│   ├── structure.md
│   ├── tech.md
│   ├── design.md
│   └── redesign-report.md
├── announcements/                     # Version announcements (4 JSON files)
│   ├── v1.12.0.json
│   ├── v1.13.0.json
│   ├── v1.14.0.json
│   └── v1.15.0.json
├── llm-configs/                       # LLM provider configs (1 JSON file)
│   └── glm.json
├── memory/                            # Session state files
│   └── last-session-state.json
├── cache/                             # Temporary cache
└── logs/                              # Hook execution logs
```

**After**

```
.moai/
├── config.yaml                        # Root: version, init status, config_version
├── config/                            # User-editable configuration
│   ├── user.yaml                      # name, timezone
│   ├── language.yaml                  # conversation_language, code_comments
│   ├── project.yaml                   # name, description, type
│   └── preferences.yaml              # NEW: merged from llm + pricing + system
├── system/                            # System-managed (not user-editable)
│   ├── quality.yaml                   # TRUST 5, LSP gates
│   ├── workflow.yaml                  # Token budgets, phases
│   ├── git-strategy.yaml              # Branch naming, merge rules
│   └── ralph.yaml                     # Loop config, convergence
├── manifest.json                      # NEW: File provenance tracking (ADR-007)
├── project/                           # Design documents (unchanged)
│   ├── product.md
│   ├── structure.md
│   ├── tech.md
│   └── design.md
├── specs/                             # SPEC documents (unchanged)
│   └── SPEC-XXX/
│       ├── spec.md
│       ├── plan.md
│       └── acceptance.md
└── state/                             # NEW: Runtime state (not version-controlled)
    ├── session.json                   # Current session state
    ├── loop.json                      # Ralph loop state
    └── lsp-baseline.json              # LSP diagnostic baseline
```

Key structural changes:
1. **User vs System split**: `config/` for user-editable, `system/` for system-managed
2. **State directory**: Runtime state separated from configuration
3. **Root config.yaml**: Moved to `.moai/` root for easy discovery
4. **manifest.json**: New file for provenance tracking (ADR-007)

### 4.2 Files Eliminated

| File | Reason | Replacement |
|------|--------|------------|
| ~~`multilingual-triggers.yaml`~~ | **Removed (legacy Python config)** | Language triggers embedded in Go binary (`internal/foundation/`) |
| `llm.yaml` | Merged into preferences | `config/preferences.yaml` |
| `pricing.yaml` | Merged into preferences | `config/preferences.yaml` |
| `system.yaml` | Split and absorbed | Relevant fields in `config/preferences.yaml` and `system/workflow.yaml` |
| `announcements/*.json` (4 files) | Embedded in Go binary | `internal/template/deployer.go` via `embed.FS` |
| `llm-configs/glm.json` | Embedded in Go binary | `internal/foundation/` |
| `memory/*` | Relocated | `state/session.json` |
| `logs/*` | Moved to global location | `$HOME/.moai/logs/` (per-user, not per-project) |
| `redesign-report.md` | Development artifact | Not deployed to user projects |

### 4.3 Root config.yaml

Moved from `.moai/config/config.yaml` to `.moai/config.yaml` for consistent discovery.

```yaml
moai:
  version: "3.0.0"
  initialized: true
  initialized_at: "2026-02-03T00:00:00Z"
  config_version: 2
```

Fields:
- `version`: MoAI-ADK version that initialized the project
- `initialized`: Boolean flag for initialization status
- `initialized_at`: ISO 8601 timestamp
- `config_version`: Schema version for migration support

### 4.4 New preferences.yaml

Merges content from the eliminated `llm.yaml`, `pricing.yaml`, and relevant parts of `system.yaml`.

```yaml
preferences:
  output_style: moai
  default_model: sonnet
  quality_model: opus
  speed_model: haiku
  token_budget: 250000
  cost_tracking: false
  log_level: info
  non_interactive: false
```

**Merge Mapping**

| Field | Previous Source | New Location |
|-------|---------------|-------------|
| `default_model` | `llm.yaml` | `preferences.yaml` |
| `quality_model` | `llm.yaml` | `preferences.yaml` |
| `speed_model` | `llm.yaml` | `preferences.yaml` |
| `token_budget` | `pricing.yaml` | `preferences.yaml` |
| `cost_tracking` | `pricing.yaml` | `preferences.yaml` |
| `output_style` | `system.yaml` | `preferences.yaml` |
| `log_level` | `system.yaml` | `preferences.yaml` |
| `non_interactive` | `system.yaml` | `preferences.yaml` |
| `version` | `system.yaml` | `config.yaml` (root) |

### 4.5 manifest.json (NEW -- ADR-007)

Tracks every deployed file's provenance for safe updates via 3-way merge (ADR-008).

```json
{
  "version": "1.0.0",
  "deployed_at": "2026-02-03T00:00:00Z",
  "files": {
    ".claude/agents/moai/expert-backend.md": {
      "provenance": "template_managed",
      "template_hash": "sha256:a1b2c3d4e5f6...",
      "deployed_hash": "sha256:a1b2c3d4e5f6...",
      "current_hash": "sha256:a1b2c3d4e5f6..."
    },
    ".claude/skills/moai-lang-go/SKILL.md": {
      "provenance": "user_modified",
      "template_hash": "sha256:f6e5d4c3b2a1...",
      "deployed_hash": "sha256:f6e5d4c3b2a1...",
      "current_hash": "sha256:1a2b3c4d5e6f..."
    },
    "CLAUDE.md": {
      "provenance": "template_managed",
      "template_hash": "sha256:9876543210ab...",
      "deployed_hash": "sha256:9876543210ab...",
      "current_hash": "sha256:9876543210ab..."
    }
  }
}
```

**Provenance Types and Update Behavior**

| Provenance | Meaning | On Update |
|-----------|---------|-----------|
| `template_managed` | Deployed from template, no user edits | Safe to overwrite |
| `user_modified` | Template base + user edits detected | 3-way merge (ADR-008) |
| `user_created` | User's own file, not from template | Never touch |
| `deprecated` | Removed from new template version | Notify user, keep file |

**Change Detection**: Compare `current_hash` (SHA-256 of file on disk) against `deployed_hash` (hash at time of last deployment). If they differ, promote provenance from `template_managed` to `user_modified`.

### 4.6 Metrics

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Config section files | 10 YAML | 8 YAML (4 user + 4 system) | -20% |
| Eliminated files | -- | 8 files removed | -- |
| New files | -- | 3 added (manifest.json, preferences.yaml, state/) | -- |
| Total .moai/ files | ~25 files | ~15 files | -40% |
| Embedded in binary | 0 | 6 files (triggers, announcements, llm-configs) | -- |
| Runtime state location | Mixed in config/ and memory/ | Consolidated in state/ | Clean separation |
| Logs location | Per-project `.moai/logs/` | Global `$HOME/.moai/logs/` | Reduced project clutter |

---

## PART 5: GO TEMPLATE EMBEDDING (go:embed)

### 5.1 templates/ Directory Structure

All template files are bundled into the Go binary via `//go:embed` directive. The `templates/` directory is the source of truth for all deployable content.

```
templates/
├── CLAUDE.md.tmpl                     # CLAUDE.md template (Go text/template)
├── .gitignore.tmpl                    # .gitignore template
├── .mcp.json.tmpl                     # MCP configuration template
├── .claude/
│   ├── agents/
│   │   └── moai/                      # 20 agent definitions
│   │       ├── manager-spec.md
│   │       ├── manager-ddd.md
│   │       ├── manager-docs.md
│   │       ├── manager-quality.md
│   │       ├── manager-project.md
│   │       ├── manager-strategy.md
│   │       ├── manager-git.md
│   │       ├── expert-backend.md
│   │       ├── expert-frontend.md
│   │       ├── expert-security.md
│   │       ├── expert-devops.md
│   │       ├── expert-performance.md
│   │       ├── expert-debug.md
│   │       ├── expert-testing.md
│   │       ├── expert-refactoring.md
│   │       ├── builder-agent.md
│   │       ├── builder-command.md
│   │       ├── builder-skill.md
│   │       ├── builder-plugin.md
│   │       └── explore.md
│   ├── skills/
│   │   ├── moai/                      # Tier 1: Core skills (10)
│   │   │   └── SKILL.md
│   │   ├── moai-foundation-core/
│   │   │   ├── SKILL.md
│   │   │   └── modules/
│   │   ├── moai-foundation-claude/
│   │   │   ├── SKILL.md
│   │   │   ├── modules/
│   │   │   └── reference/
│   │   ├── moai-foundation-quality/
│   │   │   └── SKILL.md
│   │   ├── moai-foundation-context/
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-spec/
│   │   │   ├── SKILL.md
│   │   │   └── modules/
│   │   ├── moai-workflow-ddd/
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-loop/
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-project/
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-thinking/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-go/              # Tier 2: Language skills (16)
│   │   │   └── SKILL.md
│   │   ├── moai-lang-python/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-typescript/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-java/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-rust/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-swift/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-kotlin/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-csharp/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-cpp/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-ruby/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-php/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-dart/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-elixir/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-scala/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-lua/
│   │   │   └── SKILL.md
│   │   ├── moai-lang-zig/
│   │   │   └── SKILL.md
│   │   ├── moai-domain-backend/       # Tier 2: Domain skills (4)
│   │   │   └── SKILL.md
│   │   ├── moai-domain-frontend/
│   │   │   └── SKILL.md
│   │   ├── moai-domain-database/
│   │   │   └── SKILL.md
│   │   ├── moai-domain-mobile/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-docker/      # Tier 2: Platform skills (10)
│   │   │   └── SKILL.md
│   │   ├── moai-platform-vercel/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-aws/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-gcp/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-azure/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-supabase/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-firebase/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-netlify/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-github-actions/
│   │   │   └── SKILL.md
│   │   ├── moai-platform-gitlab-ci/
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-worktree/    # Tier 2: Workflow/tool skills (5)
│   │   │   └── SKILL.md
│   │   ├── moai-workflow-testing/
│   │   │   └── SKILL.md
│   │   ├── moai-tool-ast-grep/
│   │   │   └── SKILL.md
│   │   ├── moai-essentials-perf/
│   │   │   └── SKILL.md
│   │   └── moai-essentials-debug/
│   │       └── SKILL.md
│   ├── commands/
│   │   └── moai/                      # Slash command definitions
│   │       └── moai.md
│   └── rules/
│       └── moai/
│           ├── core/                  # Core rules (always deployed)
│           │   └── moai-core.md
│           ├── workflow/              # Workflow rules (always deployed)
│           │   ├── spec-workflow.md
│           │   ├── workflow-modes.md
│           │   └── file-reading-optimization.md
│           ├── development/           # Development rules (always deployed)
│           │   ├── skill-authoring.md
│           │   └── token-budget.md
│           └── language/              # Language rules (conditional)
│               ├── go.md
│               ├── python.md
│               ├── typescript.md
│               ├── java.md
│               ├── rust.md
│               ├── swift.md
│               ├── kotlin.md
│               ├── csharp.md
│               ├── cpp.md
│               ├── ruby.md
│               ├── php.md
│               ├── dart.md
│               ├── elixir.md
│               ├── scala.md
│               ├── lua.md
│               └── zig.md
├── .moai/
│   ├── config.yaml.tmpl              # Root config template
│   ├── config/
│   │   ├── user.yaml.tmpl
│   │   ├── language.yaml.tmpl
│   │   ├── project.yaml.tmpl
│   │   └── preferences.yaml.tmpl
│   └── system/
│       ├── quality.yaml               # Static (no template variables)
│       ├── workflow.yaml               # Static
│       ├── git-strategy.yaml           # Static
│       └── ralph.yaml                  # Static
└── hooks/
    └── .gitignore                     # Placeholder for hook directory
```

### 5.2 Deployment Matrix

| Category | Always Deploy | Conditional Deploy | Never Deploy (Runtime) |
|----------|--------------|-------------------|----------------------|
| `CLAUDE.md` | Rendered from `.tmpl` | | |
| `settings.json` | Generated programmatically | | |
| `manifest.json` | Generated programmatically | | |
| `.gitignore` | Rendered from `.tmpl` | | |
| `.mcp.json` | Rendered from `.tmpl` | | |
| Agents (20) | All 20 definitions | | |
| Tier 1 Skills (10) | All 10 skill directories | | |
| Tier 2 Skills (35) | | By detection result | |
| Tier 3 Skills (11) | | | Install via CLI |
| Core Rules (6) | All 6 files | | |
| Language Rules (16) | | 1 per detected language | |
| Config user/ (4) | All 4 rendered from `.tmpl` | | |
| Config system/ (4) | All 4 static files | | |
| Root config.yaml | Rendered from `.tmpl` | | |
| Output Styles | | | Embedded in CLAUDE.md |
| `state/` | | | Created at runtime |
| `cache/` | | | Created at runtime |
| `specs/` | | | User creates via `/moai plan` |
| `project/` | | | User creates via `/moai project` |

### 5.3 Embedded Size Budget

| Category | File Count | Estimated Size |
|----------|-----------|---------------|
| Agent definitions | 20 files | ~100 KB |
| Tier 1 Core skills (with modules) | ~30 files across 10 dirs | ~150 KB |
| Tier 2 Optional skills | ~50 files across 35 dirs | ~690 KB |
| Tier 3 On-demand skills | ~15 files across 11 dirs | ~100 KB |
| Core rules | 6 files | ~30 KB |
| Language rules | 16 files | ~20 KB |
| Config templates | 8 files | ~10 KB |
| CLAUDE.md.tmpl + misc templates | 5 files | ~20 KB |
| **Total embedded** | **~150 files** | **~1.1 MB** |

The 30 MB binary size budget (set in design.md) allows ample room. At ~1.1 MB of embedded templates, the templates consume less than 4% of the binary budget.

### 5.4 Go Implementation

**Embed Declaration**

```go
package template

import "embed"

//go:embed templates
var templateFS embed.FS
```

**Deployer Interface**

```go
// Deployer extracts and deploys templates from the embedded filesystem.
type Deployer interface {
    // Deploy extracts templates to the project root based on deployment matrix.
    Deploy(ctx context.Context, opts DeployOptions) (*DeployResult, error)

    // DeploySelective deploys only specified categories of templates.
    DeploySelective(ctx context.Context, categories []Category, opts DeployOptions) (*DeployResult, error)

    // ExtractTemplate returns raw bytes of a single embedded template.
    ExtractTemplate(name string) ([]byte, error)

    // ListTemplates returns all embedded template paths.
    ListTemplates() []string

    // ListByCategory returns template paths filtered by deployment category.
    ListByCategory(cat Category) []string
}

// DeployOptions configures deployment behavior.
type DeployOptions struct {
    ProjectRoot    string
    Config         *config.Config
    Manifest       manifest.Manager
    DetectedLangs  []string          // Detected project languages
    DetectedType   string            // Detected project type
    ForceOverwrite bool              // Skip manifest checks (dangerous)
}

// DeployResult summarizes what was deployed.
type DeployResult struct {
    Deployed    []string  // Files written
    Merged      []string  // Files 3-way merged
    Skipped     []string  // Files skipped (user_created)
    Conflicted  []string  // Files with conflicts (.conflict generated)
    Errors      []error   // Non-fatal errors
}

// Category classifies template deployment behavior.
type Category int

const (
    CategoryAlways      Category = iota  // Deploy to every project
    CategoryConditional                  // Deploy based on detection
    CategoryOnDemand                     // Deploy only via CLI command
)
```

**Deployment Flow**

```
moai init (or moai update)
    |
    v
1. Load embedded templates (embed.FS)
    |
    v
2. Detect project languages and type
    |  - Scan file extensions (*.go, *.py, *.ts)
    |  - Check config files (go.mod, pyproject.toml, package.json)
    |  - Read .moai/config/project.yaml if exists
    |
    v
3. Build deployment plan
    |  - Always: CLAUDE.md, agents, Tier 1 skills, core rules, config
    |  - Conditional: Tier 2 skills matching detected languages/platforms
    |  - Skip: Tier 3 skills, output styles, runtime directories
    |
    v
4. For each file in deployment plan:
    |
    |  Is .tmpl file?
    |  ├── Yes: Render via text/template with config data
    |  └── No: Use raw bytes
    |
    |  Is settings.json?
    |  ├── Yes: Generate via json.MarshalIndent (ADR-011)
    |  └── No: Continue
    |
    |  Check manifest provenance:
    |  ├── New file (not in manifest): Write + track
    |  ├── template_managed + unchanged: Overwrite + update hash
    |  ├── template_managed + changed: Promote to user_modified + 3-way merge
    |  ├── user_modified: 3-way merge
    |  ├── user_created: Skip
    |  └── deprecated: Notify user, keep file
    |
    v
5. Update manifest.json with new hashes
    |
    v
6. Validate deployment
    |  - json.Valid() on all JSON files
    |  - Path containment checks
    |  - Agent file integrity verification
    |
    v
7. Report results to user
```

---

## IMPLEMENTATION ORDER

| Phase | Part | Description | Dependencies | Estimated Work |
|-------|------|-------------|-------------|---------------|
| 1 | Part 4 | .moai config redesign | None | ~500 LOC (config struct changes, migration logic) |
| 2 | Part 1 | Hook system (Go handlers) | Part 4 (reads config) | ~2,500 LOC (9 Go handler files) |
| 3 | Part 3 | .claude directives optimization | Part 1 (hooks in settings.json) | ~3,000 lines (agent rewrites, rules merge) |
| 4 | Part 2 | Skills optimization | Part 3 (agents reference skills) | ~5,000 lines (content edits across 56 dirs) |
| 5 | Part 5 | Go template embedding | All above complete | ~1,000 LOC (deployer, embed, validation) |

**Phase 1 rationale**: Configuration is the foundation. Hook handlers read config, templates deploy config files, and skills reference config values. Stabilize config structure first.

**Phase 2 rationale**: Hook handlers are the highest-impact change (resolving 28 issues). They depend on config but not on template content. Build and test independently.

**Phase 3 rationale**: Agent definitions and rules reference hook commands in settings.json (from Phase 2) and need stable config paths (from Phase 1). Streamline before skill optimization.

**Phase 4 rationale**: Skills reference agent names (from Phase 3) and config values (from Phase 1). Skills are the largest content volume -- optimize after structure is stable.

**Phase 5 rationale**: Template embedding bundles all artifacts from Phases 1-4 into the binary. This is the integration phase that packages everything.

---

## OVERALL METRICS

| Metric | Before (Python) | After (Go) | Change |
|--------|----------------|-----------|--------|
| Hook scripts + libraries | 21,535 LOC (32 files) | ~2,500 LOC (9 files) | -88% |
| Agent definitions | 14,578 lines (20 files) | ~3,000 lines (20 files) | -79% |
| Skills total size | 5.6 MB (56 dirs) | ~1.5 MB (56 dirs) | -73% |
| Rules files deployed | 22 files (always) | 6 core + 1-2 conditional | -64% |
| Output styles deployed | 3 files (72 KB) | 0 files (embedded in CLAUDE.md) | -100% |
| Config files | 12 YAML files | 8 YAML files | -33% |
| Total .moai/ files | ~25 files | ~15 files | -40% |
| Template files embedded | N/A (Python scripts) | ~150 files | -- |
| .claude/ total size | 7.2 MB | ~2 MB | -72% |
| CLAUDE.md | 310 lines | ~200 lines | -35% |
| Binary embed size | N/A | ~1.1 MB | Within 30 MB budget |
| Idle token consumption | ~280,000 tokens | ~1,600 tokens | -99.4% |
| Runtime dependencies | Python 3.13+, pip/uv, PyYAML, Rich | None (single binary) | -100% |
| Platform targets | 1 (Python required) | 6 (darwin/linux/windows x amd64/arm64) | +500% |
| Hook issues resolved | 28 open | 0 | -100% |
| settings.json regressions | 4 cycles (41+ commits) | 0 (struct serialization) | -100% |

---

## RISK ASSESSMENT

| Risk | Impact | Probability | Mitigation |
|------|--------|------------|------------|
| Skills content loss during optimization | High | Low | Git-tracked; review each reduction before commit |
| Agent instructions too minimal after streamlining | Medium | Medium | Test agent behavior with reduced definitions; iterate |
| Config migration breaks existing projects | High | Medium | Write migration logic in Phase 1; test with real projects |
| Binary size exceeds budget | Low | Low | Current estimate 1.1 MB vs 30 MB budget (96% headroom) |
| Context7 MCP unavailability | Medium | Low | Skills retain minimal patterns; Context7 is enhancement, not dependency |
| Template rendering errors in CLAUDE.md | High | Low | `text/template` strict mode + post-validation (ADR-011) |
| 3-way merge produces incorrect results | Medium | Medium | Comprehensive test suite for merge engine; fallback to `.conflict` files |

---

## VALIDATION CRITERIA

Each phase must pass the following gates before proceeding to the next:

**Phase 1 (Config Redesign)**:
- All existing config values load correctly from new structure
- Migration from old layout to new layout succeeds on 3+ test projects
- `config_version` field correctly triggers migration

**Phase 2 (Hook System)**:
- All 6 hook handlers respond correctly to Claude Code JSON protocol
- Hook contract tests pass on macOS, Linux, and Windows
- `moai hook <event>` returns valid JSON within 100ms timeout

**Phase 3 (Directives)**:
- All 20 agents are under 200 lines
- Agent behavior unchanged when tested with representative prompts
- Rules consolidation produces no missing references

**Phase 4 (Skills)**:
- All Tier 1 skills under 3,000 tokens
- Context7 references resolve correctly for all library mappings
- Progressive disclosure levels correctly configured

**Phase 5 (Embedding)**:
- `go build` succeeds with all templates embedded
- `moai init` deploys correct files for Go, Python, and TypeScript projects
- Manifest tracking correctly identifies all deployed files
- `json.Valid()` passes for all generated JSON files
- Binary size under 30 MB

---

## CROSS-REFERENCE INDEX

| ADR | Relevant Parts | Key Sections |
|-----|---------------|-------------|
| ADR-001 (Modular Monolithic) | All | Single binary architecture |
| ADR-006 (Hooks as Subcommands) | Part 1 | Section 1.2 Go Handler Design |
| ADR-007 (File Manifest) | Part 4, Part 5 | Section 4.5 manifest.json, Section 5.4 Deployment Flow |
| ADR-008 (3-Way Merge) | Part 4, Part 5 | Section 4.5 Provenance Types, Section 5.4 Deployment Flow |
| ADR-011 (Zero Runtime Expansion) | Part 1, Part 3 | Section 1.4 settings.json, Section 3.5 settings.json |
| ADR-012 (Hook Execution Contract) | Part 1 | Section 1.2 Supporting Infrastructure |

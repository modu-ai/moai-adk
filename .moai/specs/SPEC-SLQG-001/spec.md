# SPEC-SLQG-001: Self-Learning Quality Guard

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-SLQG-001 |
| Title | Self-Learning Quality Guard |
| Status | Draft |
| Created | 2026-04-11 |
| Priority | High |

## Problem Statement

When domain-specific code quality issues (e.g., hardcoded URLs, duplicated constants, raw environment variable strings) are discovered and fixed, **there is no mechanism to prevent the same patterns from recurring**. The fix-learn-prevent cycle is broken:

1. **Detection**: Manual discovery or code review (reactive, not proactive)
2. **Fix**: Developer/agent fixes the issue
3. **Learn**: Lesson written manually to lessons.md (aspirational, rarely done)
4. **Prevent**: No automated check exists to block recurrence

## Goal

Build a 3-layer self-learning system that **automatically detects, blocks, and learns** from domain-specific code quality patterns:

```
Layer 1: DETECT  — ast-grep structural patterns catch known bad patterns
Layer 2: LEARN   — Lessons Protocol captures fix→rule mapping automatically
Layer 3: BLOCK   — Quality Gate enforces learned rules on new code
```

---

## Requirements (EARS Format)

### Layer 1: ast-grep Domain Rules

**REQ-SLQG-001**: When a developer adds a new Go source file or modifies existing Go source, the system SHALL scan for known hardcoding patterns using ast-grep structural rules.

**REQ-SLQG-002**: The system SHALL load domain-specific rules from `.moai/config/astgrep-rules/` directory using the existing `RuleLoader.LoadFromDirectory()` API.

**REQ-SLQG-003**: The system SHALL include initial rules for:
- `no-raw-getenv`: `os.Getenv("$LITERAL")` where $LITERAL is a string literal (should use `config.Env*` constants)
- `no-hardcoded-url`: HTTP/HTTPS URL literals in non-test Go files (should be extracted to constants)
- `no-duplicate-const`: Same numeric/string literal defined as const in multiple packages

**REQ-SLQG-004**: When ast-grep detects a pattern violation in PostToolUse, the system SHALL inject an instruction message describing the violation and the recommended fix pattern.

**REQ-SLQG-005**: When ast-grep detects a pattern violation during quality gate (pre-commit), the system SHALL block the commit with a clear error message listing all violations.

### Layer 2: Lessons Protocol Enhancement

**REQ-SLQG-010**: When a bug fix commit is completed that includes the pattern "refactor: " or "fix: " in the commit message AND modifies code patterns matching an existing ast-grep rule category, the system SHALL propose a new lesson entry.

**REQ-SLQG-011**: The system SHALL auto-generate lesson entries with:
- `id`: Auto-incremented `#N`
- `category`: Derived from modified file paths and change type
- `incorrect_pattern`: Extracted from the removed/changed code
- `correct_approach`: Extracted from the replacement code
- `date`: Commit date
- `tags`: Derived from package names and change scope

**REQ-SLQG-012**: Before proposing a lesson, the system SHALL check for duplicate or superseding entries in existing lessons.md.

**REQ-SLQG-013**: The Lessons Protocol SHALL be referenced in run.md Phase 1 (Context Loading) to inject relevant lessons into the agent context before implementation begins.

**REQ-SLQG-014**: Lesson injection SHALL filter by domain relevance:
- Match lesson tags against current SPEC domain keywords
- Match lesson categories against modified file paths
- Limit to top 5 most recent matching lessons
- Maximum 2000 tokens for lesson injection

### Layer 3: Quality Gate Extension

**REQ-SLQG-020**: The quality gate (pre-commit hook) SHALL include an ast-grep scan phase after the existing lint phase.

**REQ-SLQG-021**: The ast-grep scan phase SHALL:
- Load rules from `.moai/config/astgrep-rules/` directory
- Scan only staged files (not entire project)
- Report violations with file:line, rule ID, and fix suggestion
- Block commit if any `error` severity rules are violated
- Allow commit with warning for `warning` severity rules

**REQ-SLQG-022**: The system SHALL support a rule graduation mechanism:
- New rules start as `warning` severity (observation period)
- After 3 consecutive clean scans, rules MAY be promoted to `error` severity
- Promotion requires explicit configuration change (not automatic)

**REQ-SLQG-023**: The quality gate configuration SHALL be in `.moai/config/sections/quality.yaml`:
```yaml
quality:
  ast_grep_gate:
    enabled: true
    rules_dir: ".moai/config/astgrep-rules"
    block_on_error: true
    warn_only_mode: false
```

### Cross-Layer Integration

**REQ-SLQG-030**: The system SHALL support a "Fix → Rule → Lesson" pipeline:
1. Developer fixes a pattern issue (e.g., hardcoded URL)
2. System proposes an ast-grep rule to detect similar patterns
3. System proposes a lesson entry documenting the fix
4. Rule is added to `.moai/config/astgrep-rules/` as `warning`
5. Lesson is added to `lessons.md`
6. Future code changes are scanned against the new rule

**REQ-SLQG-031**: The /moai fix and /moai loop workflows SHALL, upon successful fix completion, check if the fix pattern matches a known rule category and propose rule/lesson creation.

---

## Technical Design

### Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Fix → Learn → Prevent                    │
│                                                             │
│  ┌─────────┐    ┌──────────┐    ┌────────────┐            │
│  │  DETECT  │───>│  LEARN   │───>│   BLOCK    │            │
│  │ ast-grep │    │ lessons  │    │ quality    │            │
│  │ rules    │    │ protocol │    │ gate       │            │
│  └────┬─────┘    └─────┬────┘    └─────┬──────┘            │
│       │                │               │                    │
│       v                v               v                    │
│  .moai/config/   lessons.md     pre-commit hook             │
│  astgrep-rules/  (auto-memory)  (block violations)          │
│                                                             │
│  PostToolUse     SessionStart   PreToolUse                  │
│  (observe)       (inject)       (enforce)                   │
└────────────────────────────────────────────────────────────┘
```

### File Changes

| File | Change Type | Description |
|------|------------|-------------|
| `.moai/config/astgrep-rules/*.yml` | New | Domain-specific ast-grep rule YAML files |
| `.moai/config/sections/quality.yaml` | Modify | Add `ast_grep_gate` configuration section |
| `internal/hook/quality/gate.go` | Modify | Add ast-grep scan phase after lint phase |
| `internal/hook/quality/astgrep_gate.go` | New | ast-grep gate implementation for pre-commit |
| `internal/hook/post_tool.go` | Modify | Add domain rule instruction injection |
| `.claude/skills/moai/workflows/run.md` | Modify | Add lessons loading in Phase 1 context |
| `.claude/rules/moai/core/moai-constitution.md` | Modify | Enhance Lessons Protocol with auto-capture |

### Initial ast-grep Rules

```yaml
# .moai/config/astgrep-rules/go-no-raw-getenv.yml
id: go-no-raw-getenv
language: go
severity: warning
message: "Use config.Env* constants instead of raw os.Getenv() string literals. See internal/config/envkeys.go."
pattern: 'os.Getenv("$LITERAL")'
constraints:
  LITERAL:
    regex: "^[A-Z]"
```

```yaml
# .moai/config/astgrep-rules/go-no-hardcoded-url.yml
id: go-no-hardcoded-url
language: go
severity: warning
message: "Extract URL to a package-level constant. Hardcoded URLs reduce maintainability."
pattern: '"https://$$$URL"'
constraints:
  URL:
    regex: "^(api\\.|raw\\.)"
```

### Lessons Auto-Capture Flow

```
1. User completes fix (e.g., /moai fix → commit)
2. PostCommit analysis:
   - Parse commit message for "refactor:", "fix:" prefix
   - Diff analysis: identify removed patterns vs replacements
   - Match against known categories (hardcoding, duplication, etc.)
3. Propose lesson entry via AskUserQuestion:
   "This fix removed hardcoded URLs. Save as a lesson?"
   - Yes: Append to lessons.md with auto-generated content
   - No: Skip
4. Optionally propose ast-grep rule:
   "Create an ast-grep rule to catch similar patterns?"
   - Yes: Generate YAML rule, save to astgrep-rules/
   - No: Skip
```

---

## Acceptance Criteria

### AC-1: ast-grep Domain Rules
- [ ] `.moai/config/astgrep-rules/` directory exists with at least 2 initial rules
- [ ] `go-no-raw-getenv` rule detects `os.Getenv("LITERAL")` patterns
- [ ] `go-no-hardcoded-url` rule detects hardcoded API URLs
- [ ] PostToolUse hook injects instruction message for violations
- [ ] Rules load via existing RuleLoader.LoadFromDirectory()

### AC-2: Quality Gate ast-grep Phase
- [ ] Pre-commit quality gate includes ast-grep scan phase
- [ ] Scans only staged Go files (not entire project)
- [ ] `error` severity violations block commit
- [ ] `warning` severity violations show warning but allow commit
- [ ] Configuration via `quality.yaml` `ast_grep_gate` section
- [ ] Graceful degradation when `sg` CLI not available

### AC-3: Lessons Protocol Enhancement
- [ ] run.md Phase 1 loads relevant lessons from lessons.md
- [ ] Lesson injection filtered by domain relevance (tags + categories)
- [ ] Maximum 5 lessons, 2000 tokens injected per session
- [ ] Constitution rules updated with auto-capture trigger

### AC-4: Fix → Rule → Lesson Pipeline
- [ ] After fix commit, system proposes lesson entry (user approval required)
- [ ] Lesson auto-generation includes category, patterns, date, tags
- [ ] Duplicate lesson detection prevents redundant entries
- [ ] Pipeline integrates with /moai fix and /moai loop workflows

### AC-5: Configuration
- [ ] `quality.yaml` has `ast_grep_gate` section with enable/disable
- [ ] Rules directory path configurable (default: `.moai/config/astgrep-rules/`)
- [ ] `warn_only_mode` allows observation without blocking

---

## Implementation Phases

### Phase 1: ast-grep Domain Rules (REQ-001~005)
- Create rule YAML files
- Integrate PostToolUse instruction injection
- Test rule loading and pattern matching

### Phase 2: Quality Gate Extension (REQ-020~023)
- Add ast-grep scan to pre-commit gate
- Implement staged-file-only scanning
- Add configuration to quality.yaml

### Phase 3: Lessons Protocol Enhancement (REQ-010~014)
- Add lessons loading to run.md context
- Implement domain matching filter
- Update constitution rules

### Phase 4: Cross-Layer Integration (REQ-030~031)
- Implement fix→rule→lesson proposal pipeline
- Integrate with /moai fix and /moai loop
- Add lesson auto-generation

---

## Out of Scope

- Custom rules for languages other than Go (future SPEC)
- Auto-promotion of warning→error severity (manual only)
- MCP integration for external rule sources
- Team mode lesson synchronization

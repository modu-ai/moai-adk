---
name: moai-workflow-mx
description: >
  Scan codebase and add @MX code-level annotations for AI agent context.
  Implements 3-Pass scan with priority queue for efficient tag insertion.
  Supports all 16 MoAI-ADK languages with language-aware comment syntax.
  Use when scanning code for MX tags or annotating codebase for AI context.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-22"
  tags: "mx, annotation, code-context, scan, tagging"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["mx", "annotation", "code context", "tag scan", "mx tag"]
  agents: ["expert-backend", "expert-frontend"]
  phases: ["mx", "sync"]
---
# Workflow: MX Tag Scan and Annotation

Purpose: Scan codebase and add @MX code-level annotations for AI agent context.

For tag types, lifecycle rules, mandatory fields, and per-file limits, see: @.claude/rules/moai/workflow/mx-tag-protocol.md

## When to Use

- Legacy codebase without @MX tags
- Before major refactoring to mark danger zones
- After significant code changes to update annotations
- During `/moai sync` for MX validation

## Command

```
/moai mx [options]
```

## Flags

| Flag | Description |
|------|-------------|
| `--all` | Scan entire codebase (all languages, all P1+P2 files) |
| `--dry` | Preview only - show tags to add without modifying files |
| `--priority P1-P4` | Filter by priority level (default: all) |
| `--force` | Overwrite existing @MX tags |
| `--exclude pattern` | Additional exclude patterns (comma-separated) |
| `--lang go,py,ts` | Scan only specified languages (default: auto-detect) |
| `--threshold N` | Override fan_in threshold (default: 3) |
| `--no-discovery` | Skip Phase 0 codebase discovery |
| `--team` | Parallel scan by language (Agent Teams mode) |

## Priority Levels

| Priority | Condition | Tag Type |
|----------|-----------|----------|
| P1 | fan_in >= 3 callers | `@MX:ANCHOR` |
| P2 | goroutine/async, complexity >= 15 | `@MX:WARN` |
| P3 | magic constant, missing docstring | `@MX:NOTE` |
| P4 | missing test | `@MX:TODO` |

## Workflow Phases

### Phase 0: Codebase Discovery

**Purpose**: Detect project languages and load context before scanning.

**Steps**:
1. **Language Detection** (16 languages supported)

   Check indicator files in priority order:

   | Language | Indicator Files | Comment Prefix |
   |----------|----------------|----------------|
   | Go | go.mod, go.sum | `//` |
   | Python | pyproject.toml, setup.py, requirements.txt | `#` |
   | TypeScript | tsconfig.json, package.json (with typescript) | `//` |
   | JavaScript | package.json (without tsconfig) | `//` |
   | Rust | Cargo.toml, Cargo.lock | `//` |
   | Java | pom.xml, build.gradle, build.gradle.kts | `//` |
   | Kotlin | build.gradle.kts (with kotlin plugin) | `//` |
   | C# | .csproj, .sln, .fsproj | `//` |
   | Ruby | Gemfile, .ruby-version, Rakefile | `#` |
   | PHP | composer.json, composer.lock | `//` |
   | Elixir | mix.exs | `#` |
   | C++ | CMakeLists.txt, Makefile (with C++) | `//` |
   | Scala | build.sbt, build.sc | `//` |
   | R | DESCRIPTION, .Rproj, renv.lock | `#` |
   | Flutter/Dart | pubspec.yaml | `//` |
   | Swift | Package.swift, .xcodeproj | `//` |

2. **Project Context Loading**
   - Read `.moai/project/tech.md` for tech stack context
   - Read `.moai/project/structure.md` for architecture context
   - Read `.moai/project/product.md` for feature context
   - Read `README.md` for project overview

3. **Scan Scope Calculation**
   - Count files per language
   - Estimate token budget
   - Apply exclude patterns

### Pass 1: Full File Scan

**Purpose**: Scan all source files and generate priority queue.

**Steps**:
1. For each enabled language:
   - Glob all source files using language-specific patterns
   - Fan-in analysis: Count function/method references across files
   - Complexity detection: Lines, branches, nesting depth
   - Pattern detection: Language-specific danger patterns (goroutines, async, threading, unsafe)
2. Build priority queue (ALL files included, ranked by score)
3. Output: Priority list P1-P4

### Pass 2: Selective Deep Read

**Purpose**: Read P1 + P2 files and generate accurate tag descriptions.

**Steps**:
1. For each P1 and P2 file:
   - Full file Read with context
   - Analyze function signatures and call patterns
   - Generate tag descriptions using project context
   - Use language-specific comment syntax

**Project Context Integration**:
- Tech stack information from `tech.md`
- Architecture patterns from `structure.md`
- Business domain from `product.md`

### Pass 3: Batch Edit

**Purpose**: Insert tags into files.

**Steps**:
1. One Edit call per file
2. All tags for a given file inserted in single operation
3. Preserve existing @MX tags (unless --force)
4. Generate final report

## Batch Checkpoint Pattern

Large scanning operations (50+ files) MUST use batch processing to prevent progress loss from rate limits or session interruptions.

### Batch Configuration

- **Batch size**: 50 files per iteration (configurable via `mx.yaml` scanning.batch_size)
- **Auto-commit**: After each batch, commit intermediate results
- **Commit message**: `chore(mx): batch {N}/{M} - add MX tags to {lang} files`
- **Progress tracking**: Maintain `.moai/cache/mx-scan-progress.json`

### Progress File Format

```json
{
  "started_at": "2026-02-22T14:00:00Z",
  "total_files": 521,
  "completed_files": ["path/to/file1.go", "path/to/file2.ts"],
  "current_batch": 3,
  "total_batches": 11,
  "tags_added": 87,
  "last_language": "go"
}
```

### Resume Capability

When resuming a previously interrupted scan:
1. Read `.moai/cache/mx-scan-progress.json`
2. Skip files already listed in `completed_files`
3. Continue from `current_batch` number
4. Append new results to progress file

### Rate Limit Detection

If 3 consecutive tool failures occur during a batch:
1. Commit current batch results immediately
2. Update progress file with current state
3. Stop gracefully with message: "Rate limit detected. Progress saved. Resume with `/moai mx --resume`"
4. Do NOT retry indefinitely

### Workflow

```
For each batch (50 files):
  1. Select next 50 unscanned files from priority queue
  2. Execute Pass 2 (deep read) + Pass 3 (batch edit) for this batch
  3. Update .moai/cache/mx-scan-progress.json
  4. If auto_commit enabled: git commit with batch message
  5. If rate limit detected: save progress and stop
  6. Continue to next batch
```

## Output

After completion, generates report:

```markdown
## @MX Tag Report

### Discovery Summary
- Languages detected: Go (45), Python (23), TypeScript (67)
- Project context: Loaded from .moai/project/tech.md
- Scan scope: 135 files, 35,000 estimated tokens

### Summary
- Files scanned: 135
- Tags added: 87
- Tags updated: 23
- Tags skipped (existing): 12

### Tags by Type
- @MX:ANCHOR: 32 (P1) - High fan_in functions (>= 3 callers)
- @MX:WARN: 18 (P2) - Complex/dangerous patterns
- @MX:NOTE: 28 (P3) - Context annotations
- @MX:TODO: 9 (P4) - Missing tests

### Files Modified
- internal/core/handler.go: +5 tags
- src/api/server.ts: +4 tags
- lib/utils/helper.py: +3 tags

### Attention Required
- High fan_in functions (>= 10 callers): handler.go:ProcessRequest
```

## Integration with Other Workflows

### With /moai sync

During sync phase, MX validation runs automatically:
1. Scan files modified since last sync (all languages)
2. Check for missing @MX tags in modified functions
3. Add tags if `--skip-mx` flag not provided
4. Include tag changes in sync report

### With /moai run

During DDD ANALYZE phase:
1. If codebase has zero @MX tags, 3-Pass auto-triggers
2. Existing tags are validated and updated
3. New tags added for new code

## Examples

```bash
# Scan entire codebase (all 16 languages)
/moai mx --all

# Preview tags without modifying files
/moai mx --dry

# Only P1 priority (high fan_in functions)
/moai mx --priority P1

# Force overwrite existing tags
/moai mx --all --force

# Scan only Go and Python
/moai mx --all --lang go,python

# Lower threshold for more coverage
/moai mx --all --threshold 2
```

---

Version: 2.5.0
Last Updated: 2026-02-22
Source: SPEC-MX-001

---
description: "Agentic auto-fix - Parallel scan with autonomous correction"
argument-hint: "[--dry] [--sequential] [--level N] [file_path]"
type: utility
allowed-tools: Task, AskUserQuestion, TodoWrite, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:fix - Agentic Auto-Fix

## Core Principle: Fully Autonomous Fixing

AI autonomously finds and fixes issues.

```
START: Issue Detection
  ↓
AI: Parallel Scan → Classify → Fix → Verify
  ↓
AI: Add Completion Marker
```

## Command Purpose

Autonomously fix LSP errors and linting issues:

1. **Parallel Scan** (LSP + AST-grep + Linters simultaneously)
2. **Auto Classification** (Level 1-4)
3. **Auto Fix** (Level 1-2)
4. **Verification**
5. **Report**

Target: $ARGUMENTS

## Quick Start

```bash
# Default fix (parallel scan)
/moai:fix

# Sequential scan (for debugging)
/moai:fix --sequential

# Preview only
/moai:fix --dry

# Specific file
/moai:fix src/auth.py

# Limit fix level
/moai:fix --level 2
```

## Command Options

| Option | Alias | Description | Default |
|--------|-------|-------------|---------|
| `--dry` | --dry-run | Preview only | Apply |
| `--sequential` | --seq | Sequential scan (for debugging) | Parallel |
| `--level N` | - | Maximum fix level | 3 |
| `--errors` | --errors-only | Fix errors only | All |
| `--security` | --include-security | Include security issues | Exclude |
| `--no-fmt` | --no-format | Skip formatting | Include |

## Parallel Scan

```bash
# Sequential (30s)
LSP → AST → Linter

# Parallel (8s)
LSP   ├─┐
      ├─→ Merge (3.75x faster)
AST   ├─┤
     ├─┘
Linter
```

### Parallel Scan Implementation

By default, execute all diagnostic tools simultaneously for optimal performance:

Step 1 - Launch Background Tasks:

1. LSP Diagnostics: Use Bash tool with run_in_background set to true for LSP diagnostic command based on detected language
2. AST-grep Scan: Use Bash tool with run_in_background set to true for ast-grep with security rules from sgconfig.yml
3. Linter Scan: Use Bash tool with run_in_background set to true for appropriate linter (ruff, eslint, etc.)

Step 2 - Collect Results:

1. Use TaskOutput tool to collect results from all three background tasks
2. Wait for all tasks to complete (timeout: 120 seconds per task)
3. Handle partial failures gracefully - continue with available results

Step 3 - Aggregate and Deduplicate:

1. Parse output from each tool into structured issue list
2. Remove duplicate issues appearing in multiple scanners
3. Sort by severity: Critical, High, Medium, Low
4. Group by file path for efficient fixing

Step 4 - Proceed to Classification:

1. Classify aggregated issues into Levels 1-4
2. Call TodoWrite with all issues as pending items
3. Begin fix execution

Language-Specific Commands:

Python: ruff check --output-format json for linter, mypy --output json for types
TypeScript: eslint --format json for linter, tsc --noEmit for types
Go: golangci-lint run --out-format json for linter
Rust: cargo clippy --message-format json for linter

## Auto-Fix Levels

| Level | Description | Approval | Examples |
|-------|-------------|----------|----------|
| 1 | Immediate | Not required | import, whitespace |
| 2 | Safe | Log only | rename var, add type |
| 3 | Review | Required | logic, API |
| 4 | Manual | Not allowed | security, architecture |

## TODO-Obsessive Rule

[HARD] TodoWrite Tool Mandatory Usage:

1. Immediate Creation: When issues are discovered, call TodoWrite tool to add items with pending status
2. Immediate Progress: Before starting work, call TodoWrite tool to change item to in_progress
3. Immediate Completion: After completing work, call TodoWrite tool to change item to completed
4. Prohibited: Output TODO lists as text (MUST use TodoWrite tool)

WHY: Using TodoWrite tool allows users to track progress in real-time.

## Output Format

### Preview

```markdown
## Fix: Dry Run

### Scan (0.8s, parallel)
- LSP: 12 issues
- AST-grep: 0 security
- Linter: 5 issues

### Level 1 (12 items)
- src/auth.py: import, formatting
- src/api/routes.py: import order
- tests/test_auth.py: whitespace

### Level 2 (3 items)
- src/auth.py:45 - 'usr' → 'user'
- src/api/routes.py:78 - add type
- src/models.py:23 - dataclass?

### Level 4 (2 items)
- src/auth.py:67 - logic error
- src/api/routes.py:112 - SQL injection

No changes (--dry).
```

### Complete

```markdown
## Fix: Complete

### Applied
- Level 1: 12 issues
- Level 2: 3 issues
- Level 3: 0 issues

### Evidence
**src/auth.py:5** - Removed unused `os`, `sys`
**src/auth.py:23** - Fixed whitespace
**src/api/routes.py:12** - Sorted imports

### Remaining (Level 4)
1. src/auth.py:67 - logic error
2. src/api/routes.py:112 - SQL injection

### Next
/moai:loop  # Continue with loop
```

## Quick Reference

```bash
# Fix (default parallel)
/moai:fix

# Sequential scan
/moai:fix --sequential

# Preview only
/moai:fix --dry

# Errors only
/moai:fix --errors

# Specific file
/moai:fix src/auth.py
```

---

## EXECUTION DIRECTIVE

1. Parse $ARGUMENTS (extract --sequential, --dry, --level, --errors, --security flags)

2. Detect project language from indicator files (pyproject.toml, package.json, go.mod, Cargo.toml)

3. Execute diagnostic scan:

   IF --sequential flag is specified:

   3a. Run LSP diagnostics, then AST-grep, then Linter sequentially

   ELSE (default parallel mode):

   3b. Launch all three diagnostic tools in parallel using Bash with run_in_background:
       - Task 1: LSP diagnostics for detected language
       - Task 2: AST-grep scan with sgconfig.yml rules
       - Task 3: Linter for detected language

   3c. Collect results using TaskOutput for each background task

   3d. Aggregate results, remove duplicates, sort by severity

4. Classify aggregated issues into Level 1-4

5. [HARD] Call TodoWrite tool to add all discovered issues with pending status

6. IF --dry flag: Display preview and exit

7. [HARD] Before each fix, call TodoWrite to change item to in_progress

8. Execute Level 1-2 fixes automatically

9. [HARD] After each fix completion, call TodoWrite to change item to completed

10. Request approval for Level 3 fixes via AskUserQuestion

11. Verify fixes by re-running affected diagnostics

12. Report with evidence (file:line changes made)

---

Version: 2.1.0
Last Updated: 2026-01-11
Core: Agentic AI Auto-Fix

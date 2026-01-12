---
description: "Agentic autonomous loop - Auto-fix until completion marker"
argument-hint: "[--max N] [--auto] [--seq] | --resume snapshot"
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

# /moai:loop - Agentic Autonomous Loop

## Core Principle: Fully Autonomous Iterative Fixing

AI autonomously finds issues, fixes them, and repeats until completion.

```
START: Issue Detection
  ↓
AI: Fix → Verify → Repeat
  ↓
AI: Add Completion Marker
  ↓
<moai>DONE</moai>
```

## Command Purpose

Autonomously fix LSP errors, test failures, and coverage issues:

1. **Parallel Diagnostics** (LSP + AST-grep + Tests simultaneously)
2. **Auto TODO Generation**
3. **Autonomous Fixing** (Level 1-3)
4. **Iterative Verification**
5. **Completion Marker Detection**

Arguments: $ARGUMENTS

## Quick Start

```bash
# Default autonomous loop (parallel diagnostics)
/moai:loop

# Maximum 50 iterations
/moai:loop --max 50

# Sequential diagnostics + auto fix
/moai:loop --sequential --auto

# Restore from snapshot
/moai:loop --resume latest
```

## Command Options

| Option | Alias | Description | Default |
|--------|-------|-------------|---------|
| `--max N` | --max-iterations | Maximum iteration count | 100 |
| `--auto` | --auto-fix | Enable auto-fix | Level 1 |
| `--sequential` | --seq | Sequential diagnostics (for debugging) | Parallel |
| `--errors` | --errors-only | Fix errors only | All |
| `--coverage` | --include-coverage | Include coverage | 85% |
| `--resume ID` | --resume-from | Restore from snapshot | - |

## Completion Promise

AI adds a marker when all work is complete:

```markdown
## Loop Complete

Resolved 5 errors, 3 warnings in 7 iterations. <moai>DONE</moai>
```

**Marker Types**:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion
- `<moai:done />` - XML format

Without marker, loop continues.

## Autonomous Loop Flow

```
START: /moai:loop

PARALLEL Diagnostics (default)
  ├── LSP: Errors/Warnings
  ├── AST-grep: Security
  ├── Tests: Test results
  └── Coverage: Coverage metrics
  ↓
Integrated Results
  ↓
Completion Marker Detected?
  ├── YES → COMPLETE
  ↓
Conditions Met?
  ├── YES → "Add marker or continue?"
  ↓
TODO Generation (immediate)
  ↓
Fix Execution (autonomous)
  ├── Level 1: Immediate fix (import, formatting)
  ├── Level 2: Safe fix (rename, type)
  └── Level 3: Approval required (logic, api)
  ↓
Verification
  ↓
Max reached? → STOP
  ↓
Repeat
```

## Parallel Diagnostics

```bash
# Sequential (--sequential)
LSP → AST-grep → Tests → Coverage
Total 30s

# Parallel (default)
LSP ├─┐
     ├─→ Merge → 8s (3.75x faster)
AST ├─┤
    ├─┘
Tests ┤
       └─→ 3-4x speed improvement
Coverage
```

### Parallel Diagnostics Implementation

By default, execute all four diagnostic tools simultaneously each iteration for optimal performance:

Step 1 - Launch Background Tasks:

1. LSP Diagnostics: Use Bash tool with run_in_background set to true for language-specific LSP diagnostic command
2. AST-grep Scan: Use Bash tool with run_in_background set to true for ast-grep with security and quality rules
3. Test Runner: Use Bash tool with run_in_background set to true for language-specific test framework (pytest, jest, go test)
4. Coverage Check: Use Bash tool with run_in_background set to true for coverage measurement (coverage.py, c8, go test -cover)

Step 2 - Collect Results:

1. Use TaskOutput tool to collect results from all four background tasks
2. Wait for all tasks to complete (timeout: 120 seconds per task)
3. Handle partial failures gracefully - continue with available results

Step 3 - Aggregate Diagnostics:

1. Parse output from each tool into structured diagnostic report
2. Calculate metrics: error count, warning count, test pass rate, coverage percentage
3. Detect completion conditions: zero errors AND tests passing AND coverage meets threshold

Step 4 - Completion Check:

1. Check for completion marker in previous iteration response
2. If marker found: Exit loop with success
3. If all conditions met and no new issues: Prompt for completion marker or continue

Language-Specific Commands:

Python: pytest --tb=short for tests, coverage run -m pytest for coverage
TypeScript: npm test or jest for tests, npm run coverage for coverage
Go: go test ./... for tests, go test -cover ./... for coverage
Rust: cargo test for tests, cargo tarpaulin for coverage

## TODO-Obsessive Rule

[HARD] TodoWrite Tool Mandatory Usage:

Call TodoWrite tool on every iteration to manage tasks:

1. Immediate Creation: When issues are discovered, call TodoWrite to add items with pending status
2. Immediate Progress: Before starting work, call TodoWrite to change item to in_progress
3. Immediate Completion: After completing work, call TodoWrite to change item to completed
4. Prohibited: Output TODO lists as text (MUST use TodoWrite tool)

WHY: Using TodoWrite tool allows users to track progress in real-time.

## Auto-Fix Levels

| Level | Description | Approval | Examples |
|-------|-------------|----------|----------|
| 1 | Immediate fix | Not required | import sort, whitespace |
| 2 | Safe fix | Log only | rename var, add type |
| 3 | Approval needed | Required | logic change, API modify |
| 4 | Manual required | Not allowed | security, architecture |

## Output Format

### Running

```markdown
## Loop: 3/100 (parallel)

### Diagnostics (0.8s)
- LSP: 2 errors, 5 warnings
- AST-grep: 0 security issues
- Tests: 23/25 passing
- Coverage: 82%

### TODO
1. [x] src/auth.py:45 - undefined 'jwt_token'
2. [in_progress] src/auth.py:67 - missing return
3. [ ] tests/test_auth.py:12 - unused 'result'

Fixing...
```

### Complete (Marker Detected)

```markdown
## Loop: COMPLETE

### Summary
- Iterations: 7
- Errors fixed: 5
- Warnings fixed: 3
- Tests: 25/25 passing
- Coverage: 87%

### Files Modified
- src/auth.py (7 fixes)
- tests/test_auth.py (3 fixes)
- src/api/routes.py (2 fixes)

<moai>DONE</moai>
```

### Max Iterations Reached

```markdown
## Loop: MAX REACHED (100/100)

### Remaining
- Errors: 1
- Warnings: 2

### Options
1. /moai:loop --max 200  # Continue
2. /moai:fix             # Single run
3. Manual fix
```

## State & Snapshot

```bash
# State storage
.moai/cache/.moai_loop_state.json

# Snapshots
.moai/cache/ralph-snapshots/
├── iteration-001.json
├── iteration-002.json
└── latest.json

# Restore
/moai:loop --resume iteration-002
/moai:loop --resume latest
```

## Cancellation

```bash
# Cancel (save snapshot)
/moai:cancel-loop --snapshot

# Force cancel
/moai:cancel-loop --force
```

## Quick Reference

```bash
# Autonomous loop (default parallel)
/moai:loop

# Sequential + auto
/moai:loop --sequential --auto

# Max iterations
/moai:loop --max 50

# Errors only
/moai:loop --errors

# Restore
/moai:loop --resume latest
```

---

## EXECUTION DIRECTIVE

1. Parse $ARGUMENTS (extract --max, --auto, --sequential, --errors, --coverage, --resume flags)

2. IF --resume flag: Load state from specified snapshot and continue from saved iteration

3. Detect project language from indicator files (pyproject.toml, package.json, go.mod, Cargo.toml)

4. Initialize iteration counter to 0

5. LOOP START (while iteration less than max):

   5a. Check for completion marker in previous response:
       - If DONE, COMPLETE, or done marker found: Exit loop with success

   5b. Execute diagnostic scan:

       IF --sequential flag is specified:

       - Run LSP, then AST-grep, then Tests, then Coverage sequentially

       ELSE (default parallel mode):

       - Launch all four diagnostic tools in parallel using Bash with run_in_background:
         - Task 1: LSP diagnostics for detected language
         - Task 2: AST-grep scan with sgconfig.yml rules
         - Task 3: Test runner for detected language
         - Task 4: Coverage measurement for detected language

       - Collect results using TaskOutput for each background task

       - Aggregate results into unified diagnostic report

   5c. Check completion conditions:
       - Zero errors AND all tests passing AND coverage meets threshold
       - If all conditions met: Prompt user to add completion marker or continue

   5d. [HARD] Call TodoWrite tool to add newly discovered issues with pending status

   5e. [HARD] Before each fix, call TodoWrite to change item to in_progress

   5f. Execute fixes based on --auto level (Level 1-3)

   5g. [HARD] After each fix completion, call TodoWrite to change item to completed

   5h. Save iteration snapshot to .moai/cache/ralph-snapshots/

   5i. Increment iteration counter

6. LOOP END

7. IF max iterations reached without completion: Display remaining issues and options

8. Report final summary with evidence

---

Version: 2.1.0
Last Updated: 2026-01-11
Core: Agentic AI Autonomous Loop

---
name: moai-workflow-loop
description: >
  Iterative autonomous fixing workflow that scans, fixes, verifies, and
  repeats until all issues are resolved or max iterations reached.
  Includes memory pressure detection and snapshot-based resume.
  Use when iterative error resolution or continuous fixing is needed.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "2.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-07"
  tags: "loop, iterative, auto-fix, diagnostics, testing, coverage"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["loop", "iterate", "repeat", "until done", "keep fixing", "all errors"]
  agents: ["expert-debug", "expert-backend", "expert-frontend", "expert-testing"]
  phases: ["loop"]
---

# Workflow: Loop - Iterative Autonomous Fixing

Purpose: Iterative autonomous fixing until all issues resolved. AI scans, fixes, verifies, and repeats until completion conditions met or max iterations reached.

Flow: Check Completion -> Memory Check -> Diagnose -> Fix -> Verify -> Repeat

## Supported Flags

- --max N (alias --max-iterations): Maximum iteration count (default 100)
- --auto-fix: Enable auto-fix (default Level 1)
- --sequential (alias --seq): Sequential diagnostics instead of parallel
- --errors (alias --errors-only): Fix errors only, skip warnings
- --coverage (alias --include-coverage): Include coverage threshold (default 85%)
- --memory-check: Enable memory pressure detection
- --resume ID (alias --resume-from): Restore from snapshot

## Per-Iteration Cycle

Each iteration executes the following steps in order:

Step 1 - Completion Check:
- Check for completion marker in previous iteration response
- Marker types: `<moai>DONE</moai>`, `<moai>COMPLETE</moai>`
- If marker found: Exit loop with success

Step 2 - Memory Pressure Check (if --memory-check enabled):
- Calculate session duration from start time
- Monitor iteration time for GC pressure signs (doubling iteration time)
- If session duration exceeds 25 minutes OR iteration time doubling:
  - Save proactive checkpoint to $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/memory-pressure.json
  - Warn user about memory pressure
  - Suggest resuming with /moai:loop --resume memory-pressure
- If memory-safe limit reached (50 iterations): Exit with checkpoint

Step 3 - Parallel Diagnostics:
- Launch four diagnostic tools simultaneously using Bash with run_in_background
- Tool 1: LSP diagnostics for detected language
- Tool 2: AST-grep scan with sgconfig.yml rules
- Tool 3: Test runner for detected language (pytest, jest, go test, cargo test)
- Tool 4: Coverage measurement (coverage.py, c8, go test -cover, cargo tarpaulin)
- Collect results using TaskOutput for each background task
- Aggregate into unified diagnostic report with metrics: error count, warning count, test pass rate, coverage percentage

If --sequential flag: Run LSP, then AST-grep, then Tests, then Coverage sequentially.

Step 4 - Completion Condition Check:
- Conditions: Zero errors AND all tests passing AND coverage meets threshold
- If all conditions met: Prompt user to add completion marker or continue

Step 5 - Task Generation:
- [HARD] TaskCreate for all newly discovered issues with pending status

Step 6 - Fix Execution:
- [HARD] Before each fix: TaskUpdate to change item to in_progress
- [HARD] Agent delegation mandate: ALL fix tasks MUST be delegated to specialized agents. NEVER execute fixes directly.

Agent selection by issue type:
- Type errors, logic bugs: expert-debug subagent
- Import/module issues: expert-backend or expert-frontend subagent
- Test failures: expert-testing subagent
- Security issues: expert-security subagent
- Performance issues: expert-performance subagent

Fix levels applied per --auto setting:
- Level 1 (Immediate): No approval. Import sorting, whitespace
- Level 2 (Safe): Log only. Rename variable, add type
- Level 3 (Approval): AskUserQuestion required. Logic change, API modify
- Level 4 (Manual): Not auto-fixed. Security, architecture

Step 7 - Verification:
- [HARD] After each fix: TaskUpdate to change item to completed

Step 8 - Snapshot Save:
- Save iteration snapshot to $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/
- Increment iteration counter

Step 9 - Repeat or Exit:
- If max iterations reached: Display remaining issues and options
- Otherwise: Return to Step 1

## Completion Conditions

The loop exits when any of these conditions are met:
- Completion marker detected in response
- All conditions met: zero errors + tests passing + coverage threshold
- Max iterations reached (displays remaining issues)
- Memory pressure threshold exceeded (saves checkpoint)
- User interruption (state auto-saved)

## Snapshot Management

Snapshot location: $CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/

Files:
- iteration-001.json, iteration-002.json, etc. (per-iteration snapshots)
- latest.json (symlink to most recent)
- memory-pressure.json (proactive checkpoint on memory pressure)

Loop state file: $CLAUDE_PROJECT_DIR/.moai/cache/.moai_loop_state.json

Resume commands:
- /moai:loop --resume latest
- /moai:loop --resume iteration-002
- /moai:loop --resume memory-pressure

## Language-Specific Commands

Python: pytest --tb=short (tests), coverage run -m pytest (coverage)
TypeScript: npm test or jest (tests), npm run coverage (coverage)
Go: go test ./... (tests), go test -cover ./... (coverage)
Rust: cargo test (tests), cargo tarpaulin (coverage)

Language detection: pyproject.toml (Python), package.json (TypeScript/JavaScript), go.mod (Go), Cargo.toml (Rust)

## Cancellation

Send any message to interrupt the loop. State is automatically saved via session_end hook.

## Safe Development Protocol

All fixes within the loop follow CLAUDE.md Section 7 Safe Development Protocol:
- Reproduction-first: Write failing tests before fixing bugs
- Post-fix review: List potential side effects after each fix cycle
- Maximum 3 retries per individual operation (per CLAUDE.md constitution)

## Execution Summary

1. Parse arguments (extract flags: --max, --auto-fix, --sequential, --errors, --coverage, --memory-check, --resume)
2. If --resume: Load state from specified snapshot and continue
3. Detect project language from indicator files
4. Initialize iteration counter and memory tracking (start time)
5. Loop: Execute per-iteration cycle (Steps 1-9)
6. On exit: Report final summary with evidence
7. If memory checkpoint created: Display resume instructions

---

Version: 2.0.0
Source: loop.md command v2.2.0

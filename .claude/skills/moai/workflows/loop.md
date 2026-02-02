# Workflow: Loop - Iterative Autonomous Fixing

Purpose: Iterative autonomous fixing until all issues resolved. AI scans, fixes, verifies, and repeats until completion conditions met or max iterations reached.

Flow: Check Completion -> Memory Check -> Diagnose -> Fix -> Verify -> Repeat

## Supported Flags

- --max N (alias --max-iterations): Maximum iteration count (default 100)
- --auto (alias --auto-fix): Enable auto-fix (default Level 1)
- --sequential (alias --seq): Sequential diagnostics instead of parallel
- --errors (alias --errors-only): Fix errors only, skip warnings
- --coverage (alias --include-coverage): Include coverage threshold (from quality.yaml test_coverage_target, default 85%)
- --memory-check: Enable memory pressure detection
- --resume ID (alias --resume-from): Restore from snapshot

## Context Loading

Before execution, load these essential files:

- .moai/config/sections/quality.yaml (LSP thresholds, coverage targets, TRUST 5 settings)
- .moai/config/sections/language.yaml (conversation_language)
- .moai/config/sections/ralph.yaml (loop settings, iteration defaults)

Pre-execution commands: git status, git diff.

## Per-Iteration Cycle

Each iteration executes the following steps in order:

Step 1 - Completion Check:
- Check for completion marker in previous iteration response
- Marker types: `<moai>DONE</moai>`, `<moai>COMPLETE</moai>`, `<moai:done />`
- If marker found: Exit loop with success

Step 2 - Memory Pressure Check (if --memory-check enabled):
- Calculate session duration from start time
- Monitor iteration time for GC pressure signs (doubling iteration time)
- If session duration exceeds 25 minutes OR iteration time doubling:
  - Save proactive checkpoint to .moai/cache/ralph-snapshots/memory-pressure.json
  - Warn user about memory pressure
  - Suggest resuming with /moai loop --resume memory-pressure
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
- Conditions: LSP errors meet quality.yaml thresholds (max_errors, max_type_errors, max_lint_errors) AND all tests passing AND coverage meets quality.yaml test_coverage_target
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
- Save iteration snapshot to .moai/cache/ralph-snapshots/
- Increment iteration counter

Step 9 - Repeat or Exit:
- If max iterations reached: Display remaining issues and options
- Otherwise: Return to Step 1

## Completion Conditions

The loop exits when any of these conditions are met:
- Completion marker detected in response
- All conditions met: LSP quality gates passed + tests passing + coverage meets quality.yaml target
- Max iterations reached (displays remaining issues)
- Memory pressure threshold exceeded (saves checkpoint)
- User interruption (state auto-saved)

## Snapshot Management

Snapshot location: .moai/cache/ralph-snapshots/

Files:
- iteration-001.json, iteration-002.json, etc. (per-iteration snapshots)
- latest.json (symlink to most recent)
- memory-pressure.json (proactive checkpoint on memory pressure)

Loop state file: .moai/cache/.moai_loop_state.json

Resume commands:
- /moai loop --resume latest
- /moai loop --resume iteration-002
- /moai loop --resume memory-pressure

## Language-Specific Commands

Python: pytest --tb=short (tests), coverage run -m pytest (coverage)
TypeScript: npm test or jest (tests), npm run coverage (coverage)
Go: go test ./... (tests), go test -cover ./... (coverage)
Rust: cargo test (tests), cargo tarpaulin (coverage)

Language detection: pyproject.toml (Python), package.json (TypeScript/JavaScript), go.mod (Go), Cargo.toml (Rust)

## Cancellation

Send any message to interrupt the loop. State is automatically saved via session_end hook.

## Loop Completion: Next Steps

Tool: AskUserQuestion (at orchestrator level)

Display loop summary: iterations completed, issues fixed, remaining issues, coverage percentage.

Options:

- Commit Changes (recommended): Stage and commit all fixed files via manager-git subagent
- Run Sync: Execute /moai sync to synchronize documentation
- Review Changes: Examine all modifications before committing
- Finish: Session complete

## Graceful Exit

When user aborts or loop exits:

- Current state saved to .moai/cache/ralph-snapshots/
- No uncommitted changes lost (snapshot preserves progress)
- Display retry command: /moai loop --resume latest
- Exit with code 0

## Execution Summary

1. Parse arguments (extract flags: --max, --auto, --sequential, --errors, --coverage, --memory-check, --resume)
2. If --resume: Load state from specified snapshot and continue
3. Load context (quality.yaml, language.yaml, ralph.yaml)
4. Detect project language from indicator files
5. Initialize iteration counter and memory tracking (start time)
6. Loop: Execute per-iteration cycle (Steps 1-9)
7. On exit: Report final summary with evidence
8. If memory checkpoint created: Display resume instructions
9. Present next steps via AskUserQuestion

## Completion Criteria

All of the following must be verified:

- Context loaded: quality.yaml, language.yaml, ralph.yaml read
- Per-iteration: All diagnostic tools executed (LSP, AST-grep, Tests, Coverage)
- Task tracking: All issues tracked via TaskCreate/TaskUpdate
- Agent delegation: All fixes delegated to specialized agents
- Completion condition: LSP quality gates + tests passing + coverage target met (or max iterations reached)
- Snapshot: Final state saved for potential resume
- Next steps: Presented to user

## Agent Chain Summary

- Diagnostics: Bash (parallel background tasks)
- Fixes: expert-debug, expert-backend, expert-frontend, expert-testing, expert-security, expert-performance subagents
- Completion: manager-git subagent (optional commit)

---

Version: 1.1.0
Source: loop.md command v2.2.0. Added context loading, quality.yaml threshold references, next steps, graceful exit, completion criteria.

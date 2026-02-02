# Workflow: Fix - One-Shot Auto-Fix

Purpose: One-shot autonomous fix with parallel scanning and classification. AI finds issues, classifies by severity, applies safe fixes, and reports results.

Flow: Parallel Scan -> Classify -> Fix -> Verify -> Report

## Supported Flags

- --dry (alias --dry-run): Preview only, no changes applied
- --sequential (alias --seq): Sequential scan instead of parallel
- --level N: Maximum fix level to apply (default 3)
- --errors (alias --errors-only): Fix errors only, skip warnings
- --security (alias --include-security): Include security issues in scan
- --no-fmt (alias --no-format): Skip formatting fixes
- --resume [ID] (alias --resume-from): Resume from snapshot (latest if no ID)

## Context Loading

Before execution, load these essential files:

- .moai/config/sections/quality.yaml (LSP thresholds, coverage targets)
- .moai/config/sections/language.yaml (conversation_language, code_comments)

Pre-execution commands: git status, git diff.

## Phase 1: Parallel Scan

Launch three diagnostic tools simultaneously using Bash with run_in_background for 3-4x speedup (8s vs 30s).

Scanner 1 - LSP Diagnostics:
- Language-specific type checking and error detection
- Python: mypy --output json
- TypeScript: tsc --noEmit
- Go: go vet ./...

Scanner 2 - AST-grep Scan:
- Structural pattern matching with sgconfig.yml rules
- Security patterns and code quality rules

Scanner 3 - Linter:
- Language-specific linting
- Python: ruff check --output-format json
- TypeScript: eslint --format json
- Go: golangci-lint run --out-format json
- Rust: cargo clippy --message-format json

After all scanners complete:
- Parse output from each tool into structured issue list
- Remove duplicate issues appearing in multiple scanners
- Sort by severity: Critical, High, Medium, Low
- Group by file path for efficient fixing

Language auto-detection uses indicator files: pyproject.toml (Python), package.json (TypeScript/JavaScript), go.mod (Go), Cargo.toml (Rust). Supports 16 languages.

If --sequential flag: Run LSP, then AST-grep, then Linter sequentially.

## Phase 2: Classification

Issues classified into four levels:

- Level 1 (Immediate): No approval required. Examples: import sorting, whitespace, formatting
- Level 2 (Safe): Log only, no approval. Examples: rename variable, add type annotation
- Level 3 (Review): User approval required. Examples: logic changes, API modifications
- Level 4 (Manual): Auto-fix not allowed. Examples: security vulnerabilities, architecture changes

## Phase 3: Auto-Fix

[HARD] Agent delegation mandate: ALL fix tasks MUST be delegated to specialized agents. NEVER execute fixes directly.

Agent selection by fix level:
- Level 1 (import, formatting): expert-backend or expert-frontend subagent
- Level 2 (rename, type): expert-refactoring subagent
- Level 3 (logic, API): expert-debug or expert-backend subagent (after user approval)

Execution order:
- Level 1 fixes applied automatically via agent delegation
- Level 2 fixes applied automatically with logging
- Level 3 fixes require AskUserQuestion approval, then delegated to agent
- Level 4 fixes listed in report as manual action items

If --dry flag: Display preview of all classified issues and exit without changes.

## Phase 4: Verification

- Re-run affected diagnostics on modified files
- Confirm fixes resolved the targeted issues
- Detect any regressions introduced by fixes
- Verify against LSP quality gate thresholds from quality.yaml (max_errors, max_type_errors, max_lint_errors)

## Phase 5: Next Steps

Tool: AskUserQuestion (at orchestrator level)

Display fix summary: issues found, fixed, remaining, regressions detected.

Options:

- Commit Changes (recommended): Stage and commit fixed files via manager-git subagent
- Review Changes: Examine modified files before committing
- Continue Fixing: Apply remaining Level 3-4 fixes
- Finish: Session complete, no commit

## Task Tracking

[HARD] Task management tools mandatory:
- All discovered issues added as pending via TaskCreate
- Before each fix: change to in_progress via TaskUpdate
- After each fix: change to completed via TaskUpdate

## Snapshot Save/Resume

Snapshot location: .moai/cache/fix-snapshots/

Snapshot contents:
- Timestamp
- Target path
- Issues found, fixed, and pending counts
- Current fix level
- TODO state
- Scan results

Resume commands:
- /moai fix --resume (uses latest snapshot)
- /moai fix --resume fix-20260119-143052 (uses specific snapshot)

## Graceful Exit

When user aborts at any decision point:

- No changes made to files or Git history
- Snapshot saved to .moai/cache/fix-snapshots/ for later resume
- Display retry command: /moai fix --resume
- Exit with code 0

## Execution Summary

1. Parse arguments (extract flags: --dry, --sequential, --level, --errors, --security, --resume)
2. If --resume: Load snapshot and continue from saved state
3. Load context (quality.yaml, language.yaml)
4. Detect project language from indicator files
5. Execute parallel scan (LSP + AST-grep + Linter)
6. Aggregate results and remove duplicates
7. Classify into Levels 1-4
8. TaskCreate for all discovered issues
9. If --dry: Display preview and exit
10. Apply Level 1-2 fixes via agent delegation
11. Request approval for Level 3 fixes via AskUserQuestion
12. Verify fixes against LSP quality gate thresholds
13. Save snapshot to .moai/cache/fix-snapshots/
14. Present next steps via AskUserQuestion (commit, review, continue, finish)

## Completion Criteria

All of the following must be verified:

- Context loaded: quality.yaml and language.yaml read
- Phase 1: All scanners executed (LSP, AST-grep, Linter)
- Phase 2: Issues classified into Levels 1-4
- Task tracking: All issues tracked via TaskCreate/TaskUpdate
- Phase 3: Level 1-2 fixes applied via agent delegation
- Phase 3: Level 3 fixes applied only after user approval
- Phase 4: Verification confirms no regressions
- Phase 5: Next steps presented to user
- Snapshot saved for potential resume

## Agent Chain Summary

- Phase 1: Bash (parallel diagnostic scans)
- Phase 3: expert-backend, expert-frontend, expert-refactoring, expert-debug subagents (fix execution)
- Phase 5: manager-git subagent (optional commit)

---

Version: 1.1.0
Source: fix.md command v2.2.0. Added context loading, LSP quality gate verification, next steps, graceful exit, completion criteria.

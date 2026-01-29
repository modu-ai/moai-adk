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
- /moai:fix --resume (uses latest snapshot)
- /moai:fix --resume fix-20260119-143052 (uses specific snapshot)

## Execution Summary

1. Parse arguments (extract flags: --dry, --sequential, --level, --errors, --security, --resume)
2. If --resume: Load snapshot and continue from saved state
3. Detect project language from indicator files
4. Execute parallel scan (LSP + AST-grep + Linter)
5. Aggregate results and remove duplicates
6. Classify into Levels 1-4
7. TaskCreate for all discovered issues
8. If --dry: Display preview and exit
9. Apply Level 1-2 fixes via agent delegation
10. Request approval for Level 3 fixes via AskUserQuestion
11. Verify fixes by re-running diagnostics
12. Save snapshot to .moai/cache/fix-snapshots/
13. Report with evidence (file:line changes)

---

Version: 1.0.0
Source: fix.md command v2.2.0

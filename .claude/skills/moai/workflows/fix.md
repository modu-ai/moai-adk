---
name: moai-workflow-fix
description: >
  One-shot autonomous fix workflow with parallel scanning and classification.
  Finds LSP errors, linting issues, and type errors, classifies by severity,
  applies safe fixes via agent delegation, and reports results.
  Use when fixing errors, linting issues, or running diagnostics.
user-invocable: false
metadata:
  version: "2.5.0"
  category: "workflow"
  status: "active"
  updated: "2026-02-21"
  tags: "fix, auto-fix, lsp, linting, diagnostics, errors, type-check"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["fix", "auto-fix", "error", "lint", "diagnostic", "lsp", "type error"]
  agents: ["expert-debug", "expert-backend", "expert-frontend", "expert-refactoring"]
  phases: ["fix"]
---

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
- --team: Enable team-based debugging (see team-debug.md for competing hypothesis investigation)

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

Error handling: If any scanner fails, continue with results from successful scanners. Note the failed scanner in the report.

If --sequential flag: Run LSP, then AST-grep, then Linter sequentially.

## Phase 2: Classification

Issues classified into four levels:

- Level 1 (Immediate): No approval required. Examples: import sorting, whitespace, formatting
- Level 2 (Safe): Log only, no approval. Examples: rename variable, add type annotation
- Level 3 (Review): User approval required. Examples: logic changes, API modifications
- Level 4 (Manual): Auto-fix not allowed. Examples: security vulnerabilities, architecture changes

## Phase 2.5: Pre-Fix MX Context Scan

Before applying fixes, scan target files for existing @MX tags to understand context and constraints:

**Scan Target:** All files with classified issues (from Phase 2 results).

**MX Context Extraction:**
- @MX:ANCHOR functions: Flag as critical path. Pass fan_in context to fix agent. Warn that signature changes may break multiple callers.
- @MX:WARN zones: Pass danger context to fix agent. Ensure fix does not worsen the warned condition.
- @MX:NOTE context: Pass business logic context to fix agent to prevent fixing symptoms while breaking intent.
- @MX:TODO items: Check if any classified issues match existing TODOs (enables removal upon fix).

**Output:** MX context map passed to Phase 3 agents as part of the fix prompt. Each fix agent receives:
- List of @MX:ANCHOR functions in the target file (do not break these contracts)
- List of @MX:WARN zones (approach with caution)
- Relevant @MX:NOTE context (understand before modifying)

**Skip Condition:** If no @MX tags found in target files, proceed directly to Phase 3.

See @.claude/rules/moai/workflow/mx-tag-protocol.md for tag type definitions.

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

## Phase 4.5: MX Tag Update

After fixes are verified, update @MX tags for modified files:

**Tag Actions by Fix Level:**
| Fix Level | MX Action |
|-----------|-----------|
| Level 1 (formatting) | No tag changes typically needed |
| Level 2 (rename, type) | Update @MX:NOTE if signature changed |
| Level 3 (logic, API) | Add @MX:NOTE for new logic, re-evaluate ANCHOR |
| Level 4 (manual) | Requires @MX:WARN with @MX:REASON if security-related |

**Specific Actions:**
- Bug fix applied: Remove corresponding @MX:TODO if exists
- New code introduced: Add appropriate @MX tags per protocol
- Function signature changed: Re-evaluate @MX:ANCHOR (fan_in may change)
- Complexity increased: Add @MX:WARN if cyclomatic complexity >= 15
- Dangerous pattern introduced: Add @MX:WARN with @MX:REASON

**MX Tag Report Generation:**
Generate MX_TAG_REPORT section in fix report:
```markdown
## MX Tag Report

### Tags Added (N)
- file:line: @MX:NOTE: [description]

### Tags Removed (N)
- file:line: @MX:TODO (resolved)

### Tags Updated (N)
- file:line: @MX:ANCHOR (fan_in updated)

### Attention Required
- Files with new @MX:WARN requiring review
```

See @.claude/rules/moai/workflow/mx-tag-protocol.md for complete tag rules.

## Phase 4.6: Dead Code Cleanup (Optional)

After fixes are applied and verified, scan for dead code exposed by the fixes:

- Delegate to clean workflow (workflows/clean.md) for comprehensive dead code analysis
- Targets: Files modified during fix phase that may now have unused imports, orphaned functions, or unreferenced variables
- Skip condition: --errors flag was set (errors-only mode skips cleanup) or no dead code detected
- Clean workflow applies safe removal with test verification

## Task Tracking

[HARD] Task management tools mandatory:
- All discovered issues added as pending via TaskCreate
- Before each fix: change to in_progress via TaskUpdate
- After each fix: change to completed via TaskUpdate

## Safe Development Protocol

All fixes follow CLAUDE.md Section 7 Safe Development Protocol:
- Reproduction-first: Write a failing test that reproduces the bug before fixing
- Approach-first: For Level 3+ fixes, explain approach before applying
- Post-fix review: List potential side effects after each fix

## Snapshot Save/Resume

Snapshot location: $CLAUDE_PROJECT_DIR/.moai/cache/fix-snapshots/

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

## Team Mode

When --team flag is provided, fix delegates to a team-based debugging workflow using competing hypotheses.

Team composition: 3 hypothesis agents (haiku) exploring different root causes in parallel.

For detailed team orchestration steps, see team/debug.md.

Fallback: If team mode is unavailable, standard single-agent fix workflow continues.

Team Prerequisites:
- workflow.team.enabled: true in .moai/config/sections/workflow.yaml
- CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 in environment
- If prerequisites not met: Falls back to standard single-agent fix workflow

## Execution Summary

1. Parse arguments (extract flags: --dry, --sequential, --level, --errors, --security, --resume)
2. If --resume: Load snapshot and continue from saved state
3. Detect project language from indicator files
4. Execute parallel scan (LSP + AST-grep + Linter)
5. Aggregate results and remove duplicates
6. Classify into Levels 1-4
7. Scan target files for @MX tags (Phase 2.5: Pre-Fix MX Context Scan)
8. TaskCreate for all discovered issues
9. If --dry: Display preview and exit
10. Apply Level 1-2 fixes via agent delegation (with MX context)
11. Request approval for Level 3 fixes via AskUserQuestion
12. Verify fixes by re-running diagnostics
13. Update @MX tags for modified files (Phase 4.5)
14. Save snapshot to $CLAUDE_PROJECT_DIR/.moai/cache/fix-snapshots/
15. Report with evidence (file:line changes)

---

Version: 2.1.0
Source: fix.md command v2.3.0. Added Phase 2.5 Pre-Fix MX Context Scan for context-aware fixing.

---
id: SPEC-HOOK-CONFIG-SAFETY-001
title: "moai update Configuration Safety /hooks Review Guidance"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec (P5 of hooks-hardening Epic)
priority: P3
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hooks, cli, configuration-safety, ux, discoverability"
---

# acceptance.md — SPEC-HOOK-CONFIG-SAFETY-001

> Characterization tests (DDD per `moai-workflow-ddd`). Every AC MUST be backed by a real `go test` invocation and observed output, per `feedback_coverage_audit_table_not_actually_run` and `.claude/rules/moai/core/verification-claim-integrity.md §1.1 surface 2`.

## §A Acceptance Criteria Overview

This SPEC modifies existing CLI output behavior (adds a stdout guidance line to the `moai update` "Template sync complete" branch). Because the change is behavior-preserving except for the added line, characterization tests (DDD PRESERVE-phase style) are the correct verification mode — they capture the post-change observable behavior and pin it against regression.

## §B AC Matrix

| AC ID | Description | Severity | Verification Method |
|-------|-------------|----------|---------------------|
| AC-CS-001 | Guidance emitted on "Template sync complete" branch | MUST | Characterization test: substring `/hooks` present in captured stdout |
| AC-CS-002 | Guidance NOT emitted on 3 no-op branches | MUST | Characterization tests (3): substring `/hooks` absent on each no-op branch |
| AC-CS-003 | Guidance string contains literal `/hooks` token | MUST | Characterization test: literal substring `/hooks` (not regex) matches |
| AC-CS-004 | No `AskUserQuestion` / `mcp__askuser__*` invocation | MUST | Static guard test mirroring `TestNew_NoAskUserQuestion` pattern (`internal/cli/worktree/new_test.go`) |
| AC-CS-005 | Implementation located by content token, not line number | SHOULD | Code review note in progress.md §E.2 (run-phase) |
| AC-CS-006 | Full test suite passes with no regressions | MUST | `go test ./...` observed output (CLAUDE.local.md §6 HARD rule) |
| AC-CS-007 | Lint clean on internal/cli/ | MUST | `golangci-lint run ./internal/cli/...` observed output |

## §C Detailed Acceptance Criteria (Given-When-Then)

### AC-CS-001 — Guidance emitted on "Template sync complete" branch

**Given** the `moai update` command is invoked in a project where the deployed template version differs from the current template version (template sync WILL run).

**When** the update flow reaches the "Template sync complete" branch (`internal/cli/update.go`, located by content token `Label: "Template sync complete"`).

**Then** the completion output SHALL include a guidance line whose substring contains the literal token `/hooks`.

**Evidence requirement**: a characterization test in `internal/cli/update_test.go` (or sibling test file following project convention) that captures stdout when the template-sync-complete branch runs and asserts the literal substring `/hooks` is present. Suggested test name: `TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance`.

### AC-CS-002 — Guidance NOT emitted on no-op branches

**Given** the `moai update` command is invoked in any of three no-op states:
- (a) template already up to date (`Label: "Already up to date"` branch)
- (b) template version up-to-date, sync skipped (`Label: "Template version up-to-date · Skipping sync"` branch)
- (c) `--binary` flag set, template sync skipped (`Label: "Binary updated (template sync skipped)"` branch)

**When** the update flow takes any of these three no-op branches.

**Then** the completion output SHALL NOT contain the `/hooks` guidance substring.

**Evidence requirement**: three characterization tests (one per no-op branch) asserting the literal substring `/hooks` is absent from captured stdout. Suggested test names:
- `TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance`
- `TestUpdate_Characterize_SkippingSync_NoHooksGuidance`
- `TestUpdate_Characterize_BinaryOnly_NoHooksGuidance`

### AC-CS-003 — Guidance string contains literal `/hooks` token

**Given** the guidance line is emitted (AC-CS-001 path).

**When** the guidance string is inspected.

**Then** the string SHALL contain the literal token `/hooks` (the CC menu name, lowercase, slash-prefixed — exactly matching what the user types into CC).

**Evidence requirement**: AC-CS-001's characterization test asserts the literal substring `/hooks` (NOT a regex, NOT case-insensitive — literal). This makes the guidance actionable: the user can copy-paste `/hooks` directly into CC.

### AC-CS-004 — No subagent boundary violation

**Given** the implementation adds a stdout guidance line to `internal/cli/update.go`.

**When** the static guard test scans `internal/cli/update.go` for `AskUserQuestion` or `mcp__askuser__*` references.

**Then** the static guard SHALL find 0 matches in non-comment, non-string-literal lines, mirroring the `TestNew_NoAskUserQuestion` pattern from `internal/cli/worktree/new_test.go` (the canonical C-HRA-008 / REQ-PGN-012 static guard).

**Evidence requirement**: a `TestUpdate_NoAskUserQuestion` static guard test in `internal/cli/update_test.go` (or a dedicated `_test.go` file if the project convention is one-test-file-per-static-guard). Test logic: read `update.go`, exclude comment lines and string literals, assert no `AskUserQuestion` / `mcp__askuser__*` substrings.

### AC-CS-005 — Implementation located by content token (SHOULD)

**Given** run-phase implementation adds the guidance line.

**When** the implementation is code-reviewed.

**Then** the insertion point SHALL be adjacent to the `Label: "Template sync complete"` pill (the canonical content token anchor), NOT at a hardcoded line number.

**Evidence requirement**: a note in progress.md §E.2 (run-phase, populated by manager-develop) confirming the insertion is content-token-anchored. This is a SHOULD (line drift is acceptable if content token is used; hardcoded line numbers are the anti-pattern per `feedback_line_number_drift_asymmetry`).

### AC-CS-006 — Full test suite passes

**Given** the M1 implementation is complete.

**When** `go test ./...` is invoked from the project root.

**Then** the full test suite SHALL pass with 0 failures.

**Evidence requirement**: verbatim `go test ./...` output observed in run-phase, captured in progress.md §E.2. Per CLAUDE.local.md §6 HARD rule: "After fixing ANY test, run the FULL test suite to catch cascading failures."

### AC-CS-007 — Lint clean

**Given** the M1 implementation is complete.

**When** `golangci-lint run ./internal/cli/...` and `go vet ./internal/cli/...` are invoked.

**Then** both SHALL return 0 errors / 0 issues on `internal/cli/`.

**Evidence requirement**: verbatim command output observed in run-phase, captured in progress.md §E.2.

## §D Edge Cases

### §D.1 Empty hooks section
If the template-rendered `.claude/settings.json` has an empty `hooks` section (e.g., `{}` or `hooks: {}`), the guidance STILL fires under Option B (template sync ran). This is acceptable — the user reviewing `/hooks` sees an empty menu, no harm done. Option A's JSON-diff would have suppressed this case; Option B does not, and that's the accepted trade-off (spec.md §E.2).

### §D.2 Concurrent `moai update` and running CC session
The guidance is advisory. CC's auto-warn on external modification remains the functional safety net. If the user closes the terminal without reading the guidance, the next CC interaction still shows the auto-warn. The guidance is NOT the safety net — it's the discoverability nudge pointing at the cause (the `moai update` that just ran).

### §D.3 `--json` flag interaction
If `moai update --json` is supported, the guidance MUST NOT corrupt the JSON output. Run-phase MUST characterize the `--json` behavior on `moai update` before finalizing. If `--json` is not supported on `moai update`, this edge case is moot and run-phase records that observation. (See plan.md §B.4.)

### §D.4 `moai update` invoked from outside a CC session
The guidance still fires (Option B trigger is "Template sync complete" branch, independent of whether a CC session is running). The user sees the guidance but takes no `/hooks` action (they have no running session). This is acceptable — the cost is one terminal line, and the user may start a CC session later, at which point the guidance is contextually useful.

### §D.5 `moai update` invoked via subprocess (non-interactive)
If `moai update` is invoked via `subprocess.run` or equivalent from another tool (e.g., a CI script, a wrapper), the guidance appears on stdout. The calling tool can ignore it. No corruption of machine-readable output (assuming §D.3 `--json` handling is correct).

## §E Quality Gate Criteria

### §E.1 Coverage
- The new characterization tests MUST achieve coverage of the modified branch (`update.go` "Template sync complete" path).
- Per-package coverage target: 85% minimum on `internal/cli/` (per CLAUDE.local.md §6).
- Coverage MUST be measured by actual `go test -cover ./internal/cli/...` invocation, NOT asserted from a table (per `feedback_coverage_audit_table_not_actually_run`).

### §E.2 Lint
- `golangci-lint run` MUST return 0 errors on `internal/cli/`.
- `go vet ./internal/cli/...` MUST return 0 issues.

### §E.3 Full test suite
- `go test ./...` MUST pass with 0 failures (CLAUDE.local.md §6 HARD rule: after fixing ANY test, run the FULL test suite).

## §F Definition of Done

- [ ] AC-CS-001 characterization test PASSES — real `go test` invocation observed and captured in progress.md §E.2.
- [ ] AC-CS-002 three characterization tests PASS (3 no-op branches) — real invocations observed.
- [ ] AC-CS-003 literal `/hooks` substring assertion PASSES.
- [ ] AC-CS-004 static guard test PASSES — 0 `AskUserQuestion` references in `update.go` non-comment lines.
- [ ] AC-CS-005 code review confirms content-token-anchored insertion (SHOULD) — note in progress.md §E.2.
- [ ] AC-CS-006 `go test ./...` full suite 0 failures — verbatim output captured in progress.md §E.2.
- [ ] AC-CS-007 `golangci-lint run` + `go vet` 0 errors on `internal/cli/` — verbatim output captured.
- [ ] Coverage measured by actual `go test -cover` — verbatim output captured in progress.md §E.2.
- [ ] progress.md §E.2 run-phase evidence populated by manager-develop.
- [ ] progress.md §E.3 run-phase audit-ready signal set by manager-develop.
- [ ] progress.md §E.4 sync-phase audit-ready signal set by manager-docs (after sync commit).

## §G Forward-Looking Checks (post-implementation, pre-close)

- Has the guidance string been reviewed for clarity? (The substring `/hooks` is asserted, but the surrounding wording should be human-readable and actionable.)
- Has the guidance been tested on at least one real `moai update` invocation in a development checkout? (Characterization tests capture stdout structure; a real invocation confirms actual UX.)
- Are there any adjacent code paths (e.g., `moai cc --update`, `moai glm --update`) that also need the guidance? (Out of scope for this SPEC, but worth flagging for a follow-up if discovered.)

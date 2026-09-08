---
id: SPEC-CODEX-DISABLE-EXIT-001
title: "acceptance — moai skills disable --codex exit-code contract"
version: "0.1.0"
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
---

# acceptance.md — SPEC-CODEX-DISABLE-EXIT-001

Every AC is binary-testable. Baseline tree: `a4855f0b2`. Where an AC is conditional on the M2 adjudication outcome, the condition is stated inside the AC.

## §D. AC Matrix

### AC-CDE-001 — Branch table completeness (both outcomes) — maps REQ-CDE-001

**Given** the measured branch table in spec.md §B.1 (B1-B10), which realizes REQ-CDE-001's three-class partition (performed / refused / absent-input)
**When** each row's `file:line` address is grepped on the implementation tree
**Then** every address resolves and carries the stated stdout marker string (`Nothing to disable`, `Refusing to write`, `Unchanged:`, `Skipped:`, `[dry-run]`, `Disabled`) — verified by one grep per marker, each returning ≥1 hit in `internal/cli/codex_skills_disable.go`.

### AC-CDE-002 — Unchanged and Skipped are distinct outcomes — maps REQ-CDE-002, REQ-CDE-003, REQ-CDE-006

**Outcome B (recommended, REQ-CDE-002)**:
- RED-now cell (REQ-CDE-006): on baseline `a4855f0b2`, the Skipped branch returns nil (`codex_skills_disable.go:422-424`, read this tree) — any test asserting non-zero on a guard-refusal fixture fails. Command once the test exists: `go test ./internal/cli/ -run TestRunCodexSkillDisableSkippedExitsNonZero -count=1` → observed RED on the pre-change tree (E8 carries the verbatim output).
- Green path: M3 flips the branch, M4's test goes GREEN: same command → `ok` with exit 0.

**Outcome A (REQ-CDE-003)**:
- The exit-code table documents Unchanged (performed-class) and Skipped (refused-class) as different named outcomes, and the rule-2 waiver is present in spec.md §E.
- Mechanical check: `grep -c "consciously accepted deviation" .moai/specs/SPEC-CODEX-DISABLE-EXIT-001/spec.md` ≥ 1.

### AC-CDE-003 — Absent-input branches stay zero (both outcomes) — maps REQ-CDE-004

**Given** fixtures for mirror-absent, unresolved-codex-home, and absent-config
**When** `runCodexSkillDisable` runs against each
**Then** each returns nil. Verified by `go test ./internal/cli/ -run TestRunCodexSkillDisable -count=1` → `ok`; the existing absent-input tests keep passing unchanged.

### AC-CDE-004 — Name-resolution stays non-zero (both outcomes) — maps REQ-CDE-005

**Given** unresolved and ambiguous name fixtures
**When** the verb runs
**Then** both return non-zero errors. Verified by the existing name-resolution tests in `codex_skills_disable_test.go` passing unchanged: `go test ./internal/cli/ -run TestRunCodexSkillDisable -count=1` → `ok`.

### AC-CDE-005 — Scope decision recorded as verb-local (both outcomes) — maps REQ-CDE-007

**Given** spec.md §D
**When** `grep -c "verb-local" .moai/specs/SPEC-CODEX-DISABLE-EXIT-001/spec.md` runs
**Then** the count is ≥ 1 AND `internal/cli/codex_skills_prune.go` is byte-unchanged on the implementation branch (`git diff --stat a4855f0b2 -- internal/cli/codex_skills_prune.go` → empty).

### AC-CDE-006 — Contract documented on the verb surface (both outcomes) — maps REQ-CDE-001, REQ-CDE-008

**Given** the adjudicated exit-code contract
**When** `go run . skills disable --help` (or the built binary's equivalent) executes
**Then** the Long help text carries the per-class exit-code statement (performed / refused / absent-input with their codes), greppable from the help output.

### AC-CDE-007 — Adjudication recorded before implementation (both outcomes) — maps REQ-CDE-002

**Given** the run phase
**When** the first implementation commit lands
**Then** progress.md §E.1 already carries the M2 decision record (outcome picked + rationale + date), verifiable by reading §E.1 — the decision record precedes the commit in the file's history.

## Quality Gates

- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (E2).
- `go test ./internal/cli/ -count=1` green on the touched scope; `golangci-lint run` with zero NEW findings (E5).
- Coverage on `internal/cli` not below the 85% package target (E3).

## Definition of Done

- M2 decision recorded in progress.md §E.1 with both outcomes' prescriptions intact (the unpicked outcome stays documented as the rejected alternative).
- All seven ACs PASS with attributed evidence (command + verbatim output + baseline SHA).
- spec.md frontmatter advanced `draft → in-progress` by manager-develop at the first run-phase commit; no other status transition performed at plan phase.

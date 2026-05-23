---
id: SPEC-V3R6-LEGACY-CLEANUP-003
title: "SPEC-V3R6-LEGACY-CLEANUP-003 — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
tags: "cleanup, legacy, terminology, sprint-2, acceptance"
issue_number: null
tier: M
phase: "v3.0.0"
module: "internal/runtime"
lifecycle: spec-anchored
---

# SPEC-V3R6-LEGACY-CLEANUP-003 — Acceptance Criteria

Each AC is binary (PASS / FAIL) with a verifiable shell command. All commands assume cwd at project root (`/Users/goos/MoAI/moai-adk-go`).

## AC-LCL-001 — Comment-only Wave→Round renames applied to 14 files (REQ-LCL-001)

**Given** the 14 production Go files listed in plan.md §M1 edit map,
**When** I grep for word-boundary `Wave`/`wave` keyword in those 14 files excluding the **3 immutable exemptions** (per REQ-LCL-001 decision rule §a/b/c): (a) `SPEC-V3R3-CI-AUTONOMY-001 Wave 5` historical SPEC-ID in `state_guard.go:15` and `doc.go:11`; (b) `strategy-wave5.md` file reference in `state_guard.go:39` (file exists on disk verbatim); (c) surrounding text rewritten,
**Then** the count shall be 0.

**Verification**:

```bash
# Expected: empty output (0 lines) — word-boundary match avoids false positives on "microwave", "wavelength", etc.
grep -nE "\b[Ww]ave\b" \
  internal/ciwatch/classifier.go \
  internal/ciwatch/handoff.go \
  internal/cli/hook.go \
  internal/cli/pr/watch.go \
  internal/cli/worktree/guard.go \
  internal/cli/worktree/new.go \
  internal/config/required_checks.go \
  internal/harness/types.go \
  internal/hook/session_start.go \
  internal/hook/spec_status.go \
  internal/runtime/budget.go \
  internal/spec/lint.go \
  internal/worktree/doc.go \
  internal/worktree/state_guard.go \
  2>/dev/null \
  | grep -v "SPEC-V3R3-CI-AUTONOMY-001 Wave 5" \
  | grep -v "strategy-wave5"

# Expected: state_guard.go line 15 retains the historical SPEC-ID reference
grep -n "SPEC-V3R3-CI-AUTONOMY-001 Wave 5" internal/worktree/state_guard.go

# Expected: state_guard.go line 39 retains the file reference verbatim (file exists on disk)
grep -n "strategy-wave5" internal/worktree/state_guard.go
```

**PASS** when first command returns empty AND second command returns exactly 1 matching line AND third command returns exactly 1 matching line.

## AC-LCL-002 — PersistProgress API parameter renamed to roundLabel (REQ-LCL-002)

**Given** `internal/runtime/persist.go`,
**When** I inspect the function signatures and internals,
**Then** zero `waveLabel` / `wave_label` / `- Wave:` references shall remain, and `roundLabel` / `round_label` / `- Round:` references shall be present.

**Verification**:

```bash
# Expected: empty
grep -n "waveLabel\|wave_label\|- Wave:" internal/runtime/persist.go

# Expected: 6 matches (4 roundLabel + 1 round_label + 1 - Round:)
grep -cE "roundLabel|round_label|- Round:" internal/runtime/persist.go

# Function signature verification
grep -n "func .* PersistProgress\|func buildResumeMessage" internal/runtime/persist.go
```

**PASS** when first grep returns 0 lines AND second grep returns count ≥ 6 AND third grep shows both functions with `roundLabel` parameter.

## AC-LCL-003 — ResumeMessageFormat default uses {round_label} placeholder (REQ-LCL-003)

**Given** `internal/runtime/config.go`,
**When** I inspect the `ResumeMessageFormat` default value (line ~136),
**Then** the placeholder shall be `{round_label}` not `{wave_label}`, and the remainder of the format string (broader template + Korean prose) shall be preserved.

**Verification**:

```bash
# Expected: 0
grep -c "{wave_label}" internal/runtime/config.go

# Expected: 1
grep -c "{round_label}" internal/runtime/config.go

# Expected: full default line printed with {round_label}
grep -n "ResumeMessageFormat:.*round_label" internal/runtime/config.go
```

**PASS** when `{wave_label}` count is 0, `{round_label}` count is 1, and the third grep returns 1 line containing the renamed default value.

## AC-LCL-004 — DefaultFallback renamed to "split_into_rounds" + downstream consumers updated (REQ-LCL-004)

**Given** `internal/runtime/config.go` line 28 + `internal/runtime/budget.go` line 166 + `internal/runtime/budget_test.go`,
**When** I grep for `split_into_waves` and `smaller waves` across the runtime package,
**Then** zero occurrences shall remain.

**Verification**:

```bash
# Expected: 0
grep -rn "split_into_waves\|smaller waves" internal/runtime/

# Expected: ≥ 5 matches (DefaultFallback + 2 budget_test.go assertions + 1 YAML fixture + budget.go message)
grep -rcE "split_into_rounds|smaller rounds" internal/runtime/
```

**PASS** when first grep returns 0 lines AND second grep returns total count ≥ 5.

## AC-LCL-005 — Test file budget_test.go aligned with production rename (REQ-LCL-005)

**Given** `internal/runtime/budget_test.go`,
**When** I grep for stale Wave-terminology string literals,
**Then** zero `"Wave [0-9]"`, zero `"split_into_waves"`, zero `"{wave_label}"` shall remain.

**Verification**:

```bash
# Expected: 0
grep -nE '"Wave [0-9]"|split_into_waves|\{wave_label\}' internal/runtime/budget_test.go

# Expected: ≥ 5 matches (3 "Round N" + 1 "{round_label}" + ≥1 "split_into_rounds")
grep -ncE '"Round [0-9]"|split_into_rounds|\{round_label\}' internal/runtime/budget_test.go
```

**PASS** when first grep returns 0 lines AND second grep returns count ≥ 5.

## AC-LCL-006 — Zero regression: full test suite passes (REQ-LCL-006 enforcement)

**Given** all M1-M4 edits applied,
**When** I run the full Go test suite with the race detector,
**Then** all tests shall pass with no failures, no skips except pre-existing skips, no race conditions.

**Verification**:

```bash
# Expected: PASS, exit code 0
go test ./...

# Expected: PASS, exit code 0
go test -race ./internal/runtime/...
```

**PASS** when both commands exit 0 with no FAIL lines in output.

## AC-LCL-007 — Quality gates clean (REQ-LCL-006 + LSP gate)

**Given** all M1-M4 edits applied,
**When** I run go vet and golangci-lint,
**Then** zero errors shall be reported.

**Verification**:

```bash
# Expected: empty output, exit 0
go vet ./...

# Expected: 0 errors (warnings allowed if pre-existing)
golangci-lint run --timeout=2m
```

**PASS** when both commands exit 0.

## AC-LCL-008 — [Unwanted] §A.6 retention categories byte-identical (REQ-LCL-007 + REQ-LCL-008)

**Given** the [Unwanted] §A.6 retention list (migrate_agency cluster + Copywriter/Designer fields + frozen.go + handle-harness-observe references + immutable SPEC-V3R3-CI-AUTONOMY-001 Wave 5 historical reference),
**When** I check each category for unintended modification,
**Then** all retention category files shall be byte-identical to HEAD (`dd321ac6c`) prior to this SPEC's run-phase commits, EXCEPT for the historical `Wave 5` reference in `internal/worktree/state_guard.go:15` which is preserved verbatim within an otherwise-modified file.

**Verification**:

```bash
# 1. migrate_agency cluster byte-identical (compare against base SHA dd321ac6c)
git diff dd321ac6c -- internal/cli/migrate_agency*.go
# Expected: empty output

# 2. Copywriter/Designer fields untouched
git diff dd321ac6c -- internal/config/types.go internal/config/defaults.go | grep -E "Copywriter|Designer"
# Expected: empty output (these specific lines untouched)

# 3. handle-harness-observe documentation references preserved (count must be ≥ 9 across 6 files)
grep -rn "handle-harness-observe" internal/ --include="*.go" 2>/dev/null | wc -l
# Expected: ≥ 9 (pre-flight count was 11 including hook.go LIVE code)

# 4. state_guard.go line 15 retains immutable SPEC-ID
grep -n "SPEC-V3R3-CI-AUTONOMY-001 Wave 5" internal/worktree/state_guard.go
# Expected: 1 matching line (line 15)

# 5. Self-compliance: no NEW Wave terminology in SPEC artifacts (excluding canonical citations + [Unwanted] retention + historical SPEC-ID)
grep -ni "wave" .moai/specs/SPEC-V3R6-LEGACY-CLEANUP-003/*.md | \
  grep -v "AP-SRN-004" | \
  grep -v "Wave→Round" | \
  grep -v "Wave to Round" | \
  grep -v "wave-to-round" | \
  grep -v "SPEC-V3R3-CI-AUTONOMY-001 Wave 5" | \
  grep -v "strategy-wave5" | \
  grep -v "{wave_label}" | \
  grep -v '"Wave [0-9]"' | \
  grep -v "Wave keyword" | \
  grep -v "Wave terminology" | \
  grep -v "Wave references" | \
  grep -v "Wave A" | grep -v "Wave 2" | grep -v "Wave 3" | grep -v "Wave 5" | grep -v "Wave 7" | grep -v "Wave C" | grep -v "Wave 1" | \
  grep -v "split_into_waves" | grep -v "smaller waves" | grep -v "waveLabel" | grep -v "wave_label"
# Expected: empty output (no contraband NEW Wave usage)
```

**PASS** when (1) empty, (2) empty, (3) ≥ 9, (4) exactly 1, (5) empty.

## Definition of Done Summary

All 8 ACs above MUST PASS for the run phase to be considered complete. Per L33 standing rule, the orchestrator runs all verifications as a single-turn multi-Bash parallel batch (read-only, independent) after `manager-develop` reports completion, before invoking `/moai sync`.

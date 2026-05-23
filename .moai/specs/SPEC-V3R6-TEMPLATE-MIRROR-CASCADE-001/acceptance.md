---
id: SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001
title: "Template mirror cascade: acceptance criteria"
version: "0.1.0"
status: draft
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai/workflows/plan"
lifecycle: spec-anchored
tags: "template-mirror, cascade, drift-fix, tier-s, sprint-2-p4-3"
---

# Acceptance Criteria — SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001

## SSOT for Acceptance Verification

5 ACs total. Each is binary PASS/FAIL with explicit verification command. acceptance.md is the canonical AC enumeration; spec.md §F-REQ rationale references AC IDs but does not duplicate AC body text.

## AC Matrix

| AC ID | Verification command | Expected outcome | REQs Covered |
|-------|----------------------|------------------|--------------|
| **AC-TMC-001** | `go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v` | Output contains `--- PASS: TestLateBranchTemplateMirror/spec-assembly.md`; no `--- FAIL` lines; exit code 0 | REQ-TMC-001, REQ-TMC-003 |
| **AC-TMC-002** | `wc -c internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md \| awk '{print $1}'` | Output is exactly `28423` | REQ-TMC-001 |
| **AC-TMC-003** | `diff .claude/skills/moai/workflows/plan/spec-assembly.md internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md \| wc -l` | Output is exactly `0` | REQ-TMC-001 |
| **AC-TMC-004** | `git diff HEAD~1..HEAD -- .claude/skills/moai/workflows/plan/spec-assembly.md \| wc -l` (post-commit) | Output is exactly `0` (source byte-identical pre/post-commit) | REQ-TMC-002 |
| **AC-TMC-005** | `go vet ./... 2>&1 \| wc -l` AND `golangci-lint run --timeout=2m 2>&1 \| tail -1` | First command output `0`; second command output `0 issues.` | REQ-TMC-005 |

> **REQ-TMC-004 (PRESERVE working tree hygiene)** and **REQ-TMC-006 (L46 sibling baseline attribution discipline)** are HARD discipline requirements but intentionally have no direct AC coverage. They are verified by the orchestrator's independent post-commit batch verify per L49 — analogous to SARM-001 REQ-SARM-007 Optional treatment and IVB-001's PRESERVE/L46 discipline patterns. PRESERVE verification: `git status --porcelain -- <PRESERVE-list>` shows unchanged. L46 verification: `go test ./internal/template/... -count=1 2>&1 | grep -cE '^--- FAIL'` returns sibling baseline count = expected-baseline minus 1 (the spec-assembly.md row cleared by this SPEC).
>
> **REQ-TMC-007 (Optional path-specific `git add` discipline)** is also intentionally without AC coverage — it documents preferred discipline (path-specific staging) but does not gate completion. MAY-clause semantics.

## Given-When-Then Scenarios

### Scenario 1: Template-mirror invariant test passes post-fix

**Given** the mirror file `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` has been overwritten from source per plan.md M1 step 4 and the change is committed.

**When** the developer runs:
```bash
go test ./internal/template/ -run 'TestLateBranchTemplateMirror/spec-assembly' -v
```

**Then** the output shows:
```
=== RUN   TestLateBranchTemplateMirror
=== RUN   TestLateBranchTemplateMirror/spec-assembly.md
--- PASS: TestLateBranchTemplateMirror/spec-assembly.md (0.00s)
--- PASS: TestLateBranchTemplateMirror (0.00s)
PASS
ok      github.com/modu-ai/moai-adk/internal/template    <duration>s
```
and exit code is `0`.

### Scenario 2: Mirror byte parity verified

**Given** the same edit applied.

**When** the developer runs:
```bash
wc -c internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md | awk '{print $1}'
```

**Then** the output is exactly `28423` (matches the source byte count from §A.2 drift evidence).

### Scenario 3: Source untouched

**Given** M1 is complete and committed (HEAD reflects the fix commit).

**When** the developer runs:
```bash
git diff HEAD~1..HEAD -- .claude/skills/moai/workflows/plan/spec-assembly.md | wc -l
```

**Then** the output is exactly `0` (zero changes to the operational source in the fix commit; only the mirror was modified per REQ-TMC-002).

### Scenario 4: Baseline quality preserved

**Given** the edit is complete and committed.

**When** the developer runs `go vet ./...` followed by `golangci-lint run --timeout=2m`.

**Then** both commands exit `0`; `go vet` produces no output (line count `0`); `golangci-lint` final line is `0 issues.` (matches pre-edit baseline; no new findings introduced per REQ-TMC-005).

### Scenario 5: Sibling baseline failures persist (L46 discipline)

**Given** the fix is committed and `spec-assembly.md` is no longer in the failing test set.

**When** the developer runs `go test ./internal/template/... -count=1` to verify the full template package.

**Then** the failing test count is decreased by exactly 1 (the `TestLateBranchTemplateMirror/spec-assembly.md` row only); all other pre-existing baseline failures (TEMPLATE-MIRROR-DRIFT-001 family + catalog/agent-folder drifts + retirement assertion mismatches enumerated in spec.md §B.2) **persist exactly as before** per REQ-TMC-006 attribution discipline. This SPEC scope is the `spec-assembly.md` cascade only; other drifts are deferred.

## Edge Cases

| Edge Case | Expected Behavior |
|-----------|-------------------|
| Source file has uncommitted edits at fix time | M1 pre-flight step 1 `Read` source captures actual content; step 7 `git diff -- .claude/skills/moai/workflows/plan/spec-assembly.md` = 0 verifies source state is clean pre-copy AND post-copy. If source has uncommitted dirty state, M1 SHOULD abort and surface to orchestrator (per L48 discipline — spec.md REQ-TMC-002 forbids any side-effect modification of source). |
| Mirror file deleted before fix | `cp` recreates the mirror file at the canonical path. `wc -c` reports `28423`. AC-TMC-001/002/003 PASS. |
| Source file deleted before fix | `cp` fails with `No such file or directory`. M1 aborts; manager-develop reports as blocker (this state should never occur — operational source is canonical and required for `/moai plan` workflow). |
| Sub-test name change in future SPEC | Out of scope for this SPEC. Future SPEC editing `TestLateBranchTemplateMirror` test runner would surface via separate test failure. |
| 32-line drift content has unicode or special characters | `cp` is byte-faithful; `wc -c` byte-count matches; `diff` reports 0; test passes. Byte-fidelity is the strictest invariant and is satisfied by mechanical `cp`. |

## Definition of Done (DoD)

- [ ] AC-TMC-001 PASS verified (post-fix test passes)
- [ ] AC-TMC-002 PASS verified (mirror byte count = 28423)
- [ ] AC-TMC-003 PASS verified (diff source vs mirror = 0)
- [ ] AC-TMC-004 PASS verified (source untouched in fix commit)
- [ ] AC-TMC-005 PASS verified (lint+vet baseline preserved)
- [ ] 4 SPEC artifacts (spec.md/plan.md/acceptance.md/progress.md) frontmatter status `draft → implemented` (sync phase)
- [ ] CHANGELOG `[Unreleased]` `### Fixed` entry added (sync phase)
- [ ] `progress.md` §Sync-phase Evidence row populated with B12 self-test PASS evidence (9th consecutive)
- [ ] `progress.md` §Mx-phase Evidence row populated with SKIP-justified per `.claude/rules/moai/workflow/mx-tag-protocol.md` §a (template-only .md edit; no @MX category trigger)

## Quality Gate Thresholds

| Gate | Threshold | This SPEC |
|------|-----------|-----------|
| plan-auditor score | ≥ 0.75 (Tier S, canonical SSOT per `.claude/agents/meta/plan-auditor.md` § Tier-differentiated PASS threshold + `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier) | Target |
| BLOCKING findings | 0 | Target |
| Code coverage | Baseline preserved (no new code) | Target |
| Lint/vet findings | 0 new | Target |
| Test pass rate | 100% (5/5 ACs) + cleared `TestLateBranchTemplateMirror/spec-assembly.md` row | Target |
| Sibling baseline failures (L46 attribution) | Net delta -1 (only `spec-assembly.md` row cleared; all others persist per REQ-TMC-006) | Target |

## Out of scope (no AC)

- Other TEMPLATE-MIRROR-DRIFT-family failures (deferred to future master `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001`)
- Catalog/agent-folder baseline failures (orthogonal subsystem)
- Operational source modification (forbidden by REQ-TMC-002)
- Test runner re-architecture (`TestLateBranchTemplateMirror` continues to operate as designed)

---
id: SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001
title: "Template mirror drift cleanup: 4-file mechanical mirror parity — Acceptance Criteria"
version: "0.1.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: GOOS행님
priority: P3
phase: "v3.0.0"
module: "internal/template/templates/.claude"
lifecycle: spec-anchored
tags: "template-mirror, drift-fix, sprint-7-entry, tier-s, mechanical-cleanup"
---

# SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001 — Acceptance Criteria

## §D. AC Matrix (canonical PASS/FAIL gates)

This file is the **canonical AC SSOT**. Every AC has a single verifiable command and a single expected output. progress.md §Run-phase Evidence references this matrix verbatim.

| AC ID | Severity | REQs Covered | Verification Command | Expected Output | Status |
|-------|----------|--------------|----------------------|-----------------|--------|
| **AC-TMD-001** | [HARD] | REQ-TMD-001 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/spec-workflow.md' -v 2>&1 \| grep -E '(PASS\|FAIL): TestRuleTemplateMirrorDrift/spec-workflow.md'` | `--- PASS: TestRuleTemplateMirrorDrift/spec-workflow.md (0.00s)` | TBD |
| **AC-TMD-002** | [HARD] | REQ-TMD-002 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/agent-common-protocol.md' -v 2>&1 \| grep -E '(PASS\|FAIL): TestRuleTemplateMirrorDrift/agent-common-protocol.md'` | `--- PASS: TestRuleTemplateMirrorDrift/agent-common-protocol.md (0.00s)` | TBD |
| **AC-TMD-003** | [HARD] | REQ-TMD-003 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/plan-auditor.md' -v 2>&1 \| grep -E '(PASS\|FAIL): TestRuleTemplateMirrorDrift/plan-auditor.md'` | `--- PASS: TestRuleTemplateMirrorDrift/plan-auditor.md (0.00s)` | TBD |
| **AC-TMD-004** | [HARD] | REQ-TMD-004, REQ-TMD-005 | `go test ./internal/template/ -run 'TestRuleTemplateMirrorDrift/hooks-system.md' -v 2>&1 \| grep -E '(PASS\|FAIL): TestRuleTemplateMirrorDrift/hooks-system.md'` | `--- PASS: TestRuleTemplateMirrorDrift/hooks-system.md (0.00s)` (NEW subtest activated by registry add) | TBD |
| **AC-TMD-005** | [HARD] | REQ-TMD-009 | `go vet ./... 2>&1; echo "vet_exit=$?"; golangci-lint run --timeout=2m 2>&1 \| tail -1` | `vet_exit=0` AND `0 issues.` | TBD |

## §D.1 Severity convention

- **[HARD]**: Must-pass. FAIL on any [HARD] AC blocks SPEC closure; manager-develop returns blocker report.
- **[SHOULD]**: Should-pass. FAIL with documented PASS-WITH-DEBT acceptable in rare cases (none in this SPEC).
- **[MAY]**: Optional verification, no closure blocking.

All 5 ACs in this SPEC are [HARD]. PASS-WITH-DEBT is NOT acceptable.

## §D.2 Status enum

- **TBD**: Pre-run-phase placeholder (default in plan-phase).
- **PASS**: Verification command produced expected output.
- **FAIL**: Verification command did NOT produce expected output → blocker report obligation.
- **PASS-WITH-DEBT**: Documented baseline failure attributable to sibling SPEC per L46 (NOT applicable to this SPEC's 5 ACs).

## §D.3 Coverage to REQ-TMD-001..011 traceability

| REQ-TMD | Covered by AC | Notes |
|---------|---------------|-------|
| REQ-TMD-001 (spec-workflow.md mirror parity) | AC-TMD-001 | Direct 1:1 |
| REQ-TMD-002 (agent-common-protocol.md mirror parity) | AC-TMD-002 | Direct 1:1 |
| REQ-TMD-003 (plan-auditor.md mirror parity) | AC-TMD-003 | Direct 1:1 |
| REQ-TMD-004 (hooks-system.md mirror parity) | AC-TMD-004 | Direct 1:1 (depends on REQ-TMD-005 registry add) |
| REQ-TMD-005 (rule_template_mirror_test.go registry add) | AC-TMD-004 | Indirect — registry add activates hooks-system.md subtest |
| REQ-TMD-006 (sources untouched — Unwanted Behavior) | (verified by §D.4 indirect) | No AC mapping per §C.1 decision rule (MAY semantics not applicable — but verified by git diff post-commit) |
| REQ-TMD-007 (PRESERVE list 11 entries — Unwanted Behavior) | (verified by §D.4 indirect) | No AC mapping — verified by `git status --porcelain` snapshot diff pre/post |
| REQ-TMD-008 (TestRuleTemplateMirrorDrift all 4 PASS) | AC-TMD-001..004 aggregate | Combined coverage |
| REQ-TMD-009 (go vet + golangci-lint baseline preserved) | AC-TMD-005 | Direct 1:1 |
| REQ-TMD-010 (other baseline failures persist exactly) | (verified by §D.4 indirect) | L46 attribution discipline; no AC mapping |
| REQ-TMD-011 (path-specific git add — MAY) | (verified by §D.4 indirect) | MAY semantics — no AC mapping per §C.1 decision rule (matches L48 precedent: REQ-SARM-007 / REQ-TMC-007) |

## §D.4 Indirect verification (no AC mapping but required for closure)

These checks are NOT in the AC matrix because they are **invariants** rather than test outcomes. manager-develop verifies them as part of M1 self-verification per plan.md §E:

| Invariant | Verification | Expected |
|-----------|-------------|----------|
| Sources untouched (REQ-TMD-006) | `git diff HEAD..HEAD~1 -- <4 sources>` | 0 lines |
| PRESERVE list unchanged (REQ-TMD-007) | `git status --porcelain` pre vs post M1 | 11 entries identical status |
| Baseline failures persist (REQ-TMD-010) | `go test ./internal/template/ -v 2>&1 \| grep -c '^--- FAIL:'` | unchanged or decreased only by the 4 fixed subtests + 1 NEW subtest = net -3 (4 newly PASS minus 1 NEW activation; subtests = -3) |
| Path-specific staging (REQ-TMD-011, MAY) | `git diff --cached --name-only` post-add | exactly 5 paths from §A.2 EXTEND list |

## §D.5 Closure gate

SPEC closure (status transition `draft → implemented`) requires:
1. All 5 [HARD] ACs in §D matrix = PASS
2. All 4 invariants in §D.4 verified
3. `🗿 MoAI` trailer present in M1 commit
4. `git push origin main` succeeded (L44 post-push fetch = `0 0`)
5. manager-develop returned structured self-verification report per plan.md §E

If any of the 5 gates fail, SPEC remains `draft` and manager-develop returns blocker report to orchestrator.

## §D.6 Verification command reproducibility

All verification commands in §D are reproducible from project root `/Users/goos/MoAI/moai-adk-go`. No environment variables required. Commands assume `go test`, `go vet`, `golangci-lint`, `git`, `grep` are on PATH.

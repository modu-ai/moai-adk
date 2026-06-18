---
id: SPEC-V3R6-LIFECYCLE-REDESIGN-001
progress_version: "0.2.0"
spec_version: "0.2.0"
status: in-progress
created: 2026-06-18
updated: 2026-06-19
author: manager-spec
tier: L
---

# Progress — SPEC-V3R6-LIFECYCLE-REDESIGN-001

> Plan-phase artifact. The §E section skeleton below carries placeholder headings only; run/sync/Mx evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4) per the canonical agent responsibility realignment.

## §E.1 Plan-phase Audit-Ready Signal

- spec.md: 12 canonical frontmatter fields present; 21 GEARS REQs (REQ-LR-001..021); Exclusions section (§J) with `### Out of Scope` H3.
- plan.md: 9 milestones (M1-M9); risk register (R1-R5); anti-pattern catalogue (AP-LR-P-001..005).
- acceptance.md: 13 ACs (AC-LR-001..013); 10 MUST-PASS, 3 SHOULD-PASS; Given-When-Then format.
- research.md: drift surface measured (Axis A = 14 files; Axis B = 102 files); era migration impact (V3R6 moving baseline, re-derived H-6 at-risk set = empty — §D.4); corrected era-reclassification trace (§D.3); Spec Kit citation verified (fetched 2026-06-18).
- design.md: H-4 reclassification strategy (corrected H-5 fall-through + S1 auto-migrate + narrowed dual-predicate window); all-three-findings drift update (§B.4 incl. `Y_N_N_Y`); close-infix reconciliation with DRIFT-LEGACY-CONVENTION-001 (§B.6); Epic taxonomy mapping (4 canonical terms).

Plan-phase revision: v0.2.0 (2026-06-19) — plan-audit iter-1 FAIL 0.71/0.85 → 7 defects fixed (D1 era mechanism / D2 Y_N_N_Y / D3 moving baseline / D4 close-infix reconciliation / D5 doc-comment+§E.5 scope / D6 file count / D7 §I summary). All ground-truth verified by direct source inspection.

Plan-phase audit-ready: _(pending plan-auditor iter-2 verdict)_

## §E.2 Run-phase Evidence

### Pre-flight (captured at M1 start, tree HEAD f2907ba4c)

- **PF-1 (D3, baseline N)**: `moai spec audit --json` → total_specs=353, grandfathered=272, modern_era_clean=78, **V3R6 count N=50** (moving baseline; NOT a frozen literal — AC-LR-003 asserts invariance post-M1 == post-M3 == this N). Breakdown: Y_N_N_Y=0, Y_Y_N_Y=4, Y_Y_Y_Y_StatusDrift=3.
- **PF-1b (D1, H-6 at-risk re-derivation)**: research.md §D.4 reproduction command → V3R6 total=50, **genuine H-6 at-risk=0** (empty set). Every current V3R6 SPEC is caught by H-5's `created >= 2026-04-01` / modern-`phase:` heuristic. REQ-LR-006 dual-predicate window is defense-in-depth + classification-rationale precision, not misclassification-prevention. **No blocker.**
- **PF-2 (regression baseline)**: `go test ./internal/spec/...` → 2 PRE-EXISTING failures in `lint_test.go` (`TestLinter_AC08_DanglingRuleReference`, `TestLinter_AC11_StrictMode`) — both in the linter domain (DanglingRuleReference / strict-mode warning escalation), OUT of M1-M3 scope (era.go/audit.go/transitions.go). These are the regression baseline; M1-M3 must not introduce NEW failures and must not touch these.
- **Build baseline**: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **Git**: branch=worktree-agent-a669304f677a4add1 (fast-forwarded to main HEAD f2907ba4c to acquire SPEC artifacts + current source); `internal/spec/` clean.

### M1 — era.go H-4 3-phase reclassification + dual-predicate window

_<pending M1 commit>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

sync_commit_sha: _(pending sync-phase)_

## §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — NOTE: this section is slated for removal per REQ-LR-004 / REQ-LR-007 of this very SPEC. The redesign merges §E.5 into §E.4. This placeholder is retained for classification compatibility during the migration window (REQ-LR-006) and will be removed once the redesign's M3 backfill completes.>_

mx_commit_sha: _(not applicable — this SPEC removes the Mx-phase concept)_

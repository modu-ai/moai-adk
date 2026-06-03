# Progress — SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001
era: V3R6
tier: M
plan_complete_at: 2026-06-03
plan_status: audit-ready
plan_audit_verdict: PASS-WITH-DEBT 0.86 (GATE-2 approved)
plan_audit_remediation: D1/D2 (SHOULD-FIX) + D3/D4 (MINOR) all addressed in this commit
artifacts:
  - spec.md          # 12-field frontmatter + era:V3R6, GEARS REQ-HAW-001..016 (+013b), Exclusions EX-1..EX-8
  - plan.md          # Tier M justified, M1..M6 milestones, template-first ordering, D1 prefix disambiguation
  - acceptance.md    # AC-HAW-001..015 + AC-HAW-PROC-1..2, all grep/test-verifiable, DoD, edge cases
  - design.md        # wiring-mechanism decision (Option A recommended), main.md additive router, 3-check smoke gate
  - progress.md      # this file
authored_by: manager-spec
plan_commit_sha: (this commit)
```

### Plan-audit remediation log (2026-06-03)

| Defect | Severity | Resolution |
|--------|----------|------------|
| D1 | SHOULD-FIX | plan.md M4 + AC-HAW-PROC-2 disambiguated: §6.4 correction target is the code-side `my-harness-*` (NOT the EX-1 `harness-*` migration); cites `meta-harness SKILL.md:168` doctrine-vs-code drift; AC PROC-2 now asserts `my-harness-` equality + no bare `harness-*` directive (impossible to read as endorsing migration) |
| D2 | SHOULD-FIX | (preferred option) REQ-HAW-013b + AC-HAW-015 added: Phase-6 smoke gate FAILs when a generated agent OMITS `skills:` — runtime enforcement of REQ-HAW-008, closing the silent auto-discovery gap; spec.md REQ-HAW-008 binding updated; plan.md M5 + design.md §C updated |
| D3 | MINOR | design.md §B corrected: `mainMD()` already emits Domain Summary + Linked Files; only the `## Task-Shape Routing` table is additive (not a from-scratch rewrite) |
| D4 | MINOR | spec.md EX-1 future SPEC-ID `SPEC-V3R6-HARNESS-NAMESPACE-V2-001` marked **(planned — not yet created)** |

## §E.0 Ground-Truth Diagnosis Anchors (verified during plan-phase)

- `InjectMarker` (`internal/harness/layer3.go`) — **0 non-test callers** (orphaned, but works + tested)
- `ScaffoldHarnessDir` (`internal/harness/layer5.go`, emits `main.md`) — **0 non-test callers** (orphaned)
- This repo CLAUDE.md + template CLAUDE.md — **0 `moai:harness-start` markers** each
- `project/meta-harness.md` — Phase 7 (5-Layer Activation) referenced but **body absent** (file ends at Phase 6.5)
- L4 import lines (`@.moai/harness/`) present in all of plan/run/sync/design workflows — L4 intact
- `doctor harness` L3 (marker) + L5 (`main.md`) checks already exist — smoke gate reuses them
- B3 (empty agent descriptions) REFUTED per diagnosis — codified as REQ-HAW-009 for the gate to assert

## §E.2 Run-phase Evidence
_(populated by manager-develop at run-phase)_

## §E.3 Run-phase Audit-Ready Signal
_(populated by manager-develop at run-phase)_

## §E.4 Sync-phase Audit-Ready Signal
_(populated by manager-docs at sync-phase)_

## §E.5 Mx-phase Audit-Ready Signal
_(populated at close)_

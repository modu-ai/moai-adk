# Progress: SPEC-FACTORY-DEFAULT-MODE-001

Card t1737 · Tier M · plan-phase artifact set (spec.md, plan.md, acceptance.md, progress.md,
decision-index.md). This file is lane-owned for §E.2–§E.4; manager-spec populates §E.1 only.

## §E.1 Plan-phase Audit-Ready Signal

- audit_ready: true
- plan_status: audit-ready
- plan_complete_at: 2026-10-19T03:20:00Z (UTC)
- artifacts: spec.md (7 REQ), acceptance.md (11 AC: 8 release-blocking, 3 regression-guard),
  plan.md, decision-index.md (6 rows: 3 DECIDED, 0 POLICY-COVERED, 0 EVIDENCE-NEEDED, 3
  FOUNDER — two implementation-level DEFAULT-APPLIED, one product-level operator-gated)
- lint: `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` — (pending first run; the
  observed result is recorded here after the run)
- predecessor note: C1 (t1735) skipped by operator exception; safety rationale at plan.md §A.2.

### Plan repair — round-1 audit FAIL (2026-10-19, this session)

Round-1 cold plan-audit: FAIL, overall 0.875, MP-8 must-pass failure (verdict:
`.moai/reports/t1737/plan-audit-cold.md`, audited_sha `f6742080a6`, receipt
`rcpt-c9f6d2f8c1f45eb7e134711c`). Repairs applied by manager-spec:

- **D1 (critical)** — acceptance.md two-cell state re-authored as observed-failure evidence:
  new §D.0 evidence ledger (EV-FDM-01…-15, -A1) carrying command + verbatim stdout + exit
  code + tree SHA per release-blocking AC; AC-FDM-004 demoted release-blocking →
  regression-guard (undecidable disposition — its fail-safe-serial RED cannot be honestly
  observed; none recorded); swept-set guard added (`go test -v` + named `--- PASS:` line
  required; a `[no tests to run]` result is never green); §F DoD updated. Observation
  channels: tree greps + file reads at `f6742080a6`; installed binary `moai todo classify` →
  `Unknown command "classify" for "moai todo"`, exit 1 (build `7705b365e2`, predates HEAD —
  lag stated per tool-provenance attribution); the codex zero-sweep output cited attributed
  to the verdict receipt. PENDING-observation: the three existing parent tests' passing
  outputs (EV-FDM-12/-13/-14) — the Go-restraint window was active at repair time; the lane
  records them at run entry alongside plan §C P2.
- **D2 (major)** — REQ-FDM-004 extended with the serial-slot conflict refusal (in-flight
  target ∧ requested `serial` ∧ another distinct in-flight slot holder → refuse with
  recovery advice, write nothing), grounded in `factorySerialInFlightExcluding`
  (`factory_card.go:324-340`) and the `:943`/`:1406` gates; new AC-FDM-011 +
  `TestTodoClassifySerialSlotConflictRefusal` (M2); decision-index Q6 appended (FOUNDER,
  implementation-level, DEFAULT-APPLIED: refuse).
- **D3 (minor)** — SPEC-FACTORY-CONFLICT-SLOTS-001 disposition stated in-body (spec.md §E;
  plan.md §H): proposed split not executed on this lineage; frontmatter reference kept as
  disposition-only context.
- **D4 (minor)** — plan.md §C P1 expected observation restated as measured (zero hits,
  exit 1).
- **D5 (minor)** — plan.md §F M3 relabeled guard-addition (green-now cells kept green), body
  aligned; M2 coverage line extended with AC-FDM-011.
- REQ/AC ids otherwise unchanged (no renumbering, no renaming); all other content kept stable
  for the round-2 delta re-audit (scope: acceptance.md#two-cell-adoption-state,
  spec.md#req-fdm-004, plus the AC-FDM-011 addition).
- **Round-2 repair (delta FAIL 0.9375, receipt `rcpt-ed1e059d6960d148efbfb510`, audited_sha
  `67ac361675`)** — R2-D1: EV-FDM-05/-06/-07/-09/-10/-15 re-observed with complete stdout at
  `67ac361675` (`internal/` byte-identical across all three trees); R2-D2: AC-FDM-001 bullet +
  EV-FDM-12 note reworded to the expected-green/PENDING form; R2-D3: new EV-FDM-A2 (attributed
  overlap reproduction, receipt `rcpt-c9f6d2f8c1f45eb7e134711c`) with AC-FDM-011 citations
  repointed from EV-FDM-A1; R2-D4: EV-FDM-13/-14 recording path corrected to plan §E E2;
  R2-D5: plan §C P4 enumeration extended to Q1–Q6. Single new id EV-FDM-A2; all REQ/AC/entry
  ids otherwise unchanged.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

# progress.md — SPEC-LOOP-SWEEP-001

> Canonical §E section skeleton. Plan-phase populates §E.1 only; §E.2/§E.3 are
> owned by manager-develop (run-phase), §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-12
- tier: M
- artifacts: spec.md, plan.md, acceptance.md, progress.md (research.md shared from SPEC-ANALYZE-FIRST-ROUTING-001)
- REQ count: 14 (REQ-LSW-001..014)
- AC count: 20 (AC-LSW-001..014 with 001b/002b/004b/008b/009b/013b split-outs; 001b + 013b added iteration-3 per plan-audit D1+D2 fix)
- depends_on: SPEC-GOAL-ENGINE-001
- open decisions: 0 remaining — both resolved iteration-2 (run --mode loop alias to KEEP; sweep exit_kind to sweep-residue additive). See plan.md Settled Decisions.
- plan-audit iter-1: FAIL 0.84 (Tier M threshold 0.85). D1 (1 BLOCKING) + D2 (3 should-fix) addressed iteration-3 → v0.2.0. D3 (4 items) DEFERRED to run-phase (see below).
- plan-audit iter-2 (re-audit of v0.2.0): PASS 0.91 (Tier M threshold 0.85, clears by 0.06). All 7 must-pass criteria PASS; D1 (AC-LSW-001b composition-reachability) + D2a/b/c fixes verified LANDED and non-vacuous (13 baseline-0 + 7 mirror baselines grep-verified TRUE on live tree; empty-diff guard exit 0). 1 minor residual noted for run-phase (AC-001b weak 'goal state' OR alternative). D3 four items confirmed correctly deferred, not re-raised.

### Deferred to run-phase (plan-auditor D3 — do NOT fix at plan-phase)

The following 4 plan-auditor D3 findings are should-fix items deferred to the
run-phase per orchestrator direction. They are noted here for run-phase awareness
and MUST NOT be addressed at plan-phase:

- **D3a — AC-LSW-002b split per-lens**: split the review-lens queue-additions check
  so security and `@MX` are enumerated separately (currently a single OR). Run-phase
  may split when the lenses are actually added to loop.md.
- **D3b — AC-LSW-001 coverage-conditional clause**: the "[+ coverage when enabled]"
  conditional in REQ-LSW-001/AC-LSW-001 is not independently pinned. Run-phase may
  add a coverage-conditional sub-check once the loop.md rewrite lands.
- **D3c — spec.md §B D4 stale "keep/retire" wording**: §B D4 still says "keep/retire
  the `run --mode loop` alias — decide + justify"; the decision was settled KEEP in
  iteration-2 (plan.md § Settled Decisions). Run-phase should align §B D4 wording to
  the settled KEEP decision (cosmetic prose fix, no AC impact).
- **D3d — AC-006/002b independence**: AC-LSW-006 ("review lens" consumer verb in
  loop.md) and AC-LSW-002b ("review lens" supplier noun in loop.md) share the
  "review lens" token and could theoretically both pass on one mention. Run-phase
  may tighten the two ACs to pin distinct procedure-verb vs supplier-noun forms.

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

# SPEC-DRIFT-CACHE-FILL-001 — Progress

Card t871. Branch `WT-drift-cache-fill`, base local `develop` `023664665`.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`.
SPEC ID regex pre-write check: `PASS`. ID uniqueness confirmed against
`.moai/specs/`.

**iter1** (v0.1.0): plan-auditor FAIL 0.55 against the Tier M 0.80 threshold; no
must-pass criterion failed. Report: `.moai/reports/t871/plan-audit.md`.

**iter2** (v0.2.0): revised in place. Requirements 22 → **16**, acceptance
criteria 15 → **16** (14 release-blocking, 2 regression-guard) — both exactly at
the Tier M ceiling, reached by consolidation. Tier stays **M**. Every
requirement is now referenced by at least one criterion (verified:
`comm -23` over the REQ id sets of `spec.md` and `acceptance.md` prints
nothing). Evidence ledger grew to five single-invocation RED cells (EL-5 added
for the missing typed config field); the former PENDING EL-5 promise is gone.

**iter3** (v0.3.0): scoped revision, four items only (N1/N2/N3/N5). plan-auditor
iter2 scored 0.81 (clears the Tier M 0.80 threshold) but returned FAIL under the
retry contract because one iter1 defect stayed open — the single-flight race had
moved, not closed. REQ-DCF-009 **amended** (not replaced): the suppression write
becomes an exclusive claim via `internal/atomicfile.Claim`. **Tier gate did not
fire** — the amendment adds no actor, no artifact and no new obligation, so no
REQ-017 was needed and the count stays 16 requirements / 16 acceptance criteria
at Tier M. All five evidence cells re-executed and reproduced on the new base pin
`71532427dadf89063f78c35ac5c4f33bde4ebf10`; the document pin was moved and the
old pin retained only as iteration-1/2 provenance.

**iter4** (v0.4.0): scoped revision, three items. plan-auditor iter3 scored 0.84
(above threshold) but FAILed: N2/N3/N5 closed, N1 only half-closed — the design
was right, the criterion meant to prove it was not. The audit **disproved the
v0.3.0 claim** that AC-DCF-006 clause (b) excluded the in-process-mutex mutant,
by writing the counter-mutant (a `sync.Mutex`-guarded `os.Stat` returning an
error wrapping `fs.ErrExist`). Clause (a) is now **cross-process** — N ≥ 8
re-executed child processes on the `internal/cli/gate_lock_cli_test.go` shape,
self-bounding with `-test.timeout` capping from outside — plus a new clause (c)
choke-point guard. REQ-DCF-009 **re-amended** to claim a separate
`<record>.lock` and perform the reclaim under it (NEW-1 TOCTOU, NEW-2 empty
record). The operator-preserved sentence — "an error value is not evidence of a
syscall — `fs.ErrExist` is a sentinel any code can wrap." — is carried verbatim
in `acceptance.md` § AC-DCF-006 mutant note and `plan.md` §G. AC-DCF-003 left
untouched; its observation-method question is recorded as a run-phase M2 item.
**Tier gate did not fire** — count stays 16 requirements / 16 acceptance criteria
at Tier M. All five evidence cells re-executed and reproduced on the new base pin
`881aa4bb86878b2f401244819c8ce73edba8d052`. Awaiting plan-audit iteration 4 and
Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI

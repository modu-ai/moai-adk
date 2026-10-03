# plan.md — SPEC-FACTORY-DECISION-AUTO-001

## §A Context

Card t1481, Tier L, release target v3.2.0. Eight capabilities that turn the leader's recurring
rulings into rules or disk records (spec.md §B). Probes measured at `ba2033d22`. Methodology per
`quality.yaml`: TDD for new Go code, DDD/characterization for the three verdict sites and the hook
path. Revision 0.2.0 answers plan-audit iter1 (`.moai/reports/t1481/plan-audit-iter1.md`).

## §B Known Issues

- Doctrine and code disagree on PASS-WITH-DEBT (research F2) and on factory Kickoff (F4). This plan
  makes both follow one phase-scoped predicate.
- T7 and T13 share `guardVerdictPass` (P9b). Plain PASS at T7 gains score, must-pass, and blocking
  checks; existing tests may encode label-only behavior and are characterized first (M2).
- The FOUNDER Default amends a [HARD] manager-spec clause (decision-index Q2, leader-decided).

## §C Pre-flight

- `go test ./internal/homestate/... ./internal/contract/... ./internal/hook/...` green at base
  (record exit codes in progress §E.2 before M1).
- `make agents-emit-check` clean at base.

## §D Constraints

Template-first; template neutrality; keep-set untouched (REQ-FDA-024); lane refusal relaxed only
for REQ-FDA-016; the run-state probe runs on every bind; no local full-suite runs.

## §E Self-Verification

Each milestone closes with its ACs' green cells (acceptance.md §D), scoped package tests, `go vet`,
`golangci-lint run` on changed packages, and for template milestones `make build` +
`make agents-emit` + `make agents-emit-check`.

## §F Milestones (ordered by decision-reversibility — highest change likelihood first)

| M | Priority | Scope | REQs | ACs |
|---|---|---|---|---|
| M0 | High | **Measurement baseline, committed before any implementation commit**: (a) degraded-notice rate under load on the current bind path, re-measured after M7 with the cache (Q5); (b) cache-write cost of a 5-minute one-shot recheck vs. longer delays (Q4), which may raise the default but never below 5. Both feed acceptance.md § Measurement notes | 019, 020 | — (measurement notes) |
| M1 | High | **Board data model + CLI**: record shape incl. `resolves`, kinds, location, append/lock, `record`/`read`, lane refusal, statuses | 001-003 | 001-003 |
| M2 | High | **Admission predicate**: characterize T7/T13 tests; shared `AdmitVerdict(phase)`; three call sites; auditor verdict-block fields (`must_pass_failed`, `blocking_count`, `scope`, `fix_scope`, `defect_class`, `reread_hunks`, `debts` with `dispose_in`); `decision-index.md` added to the plan-artifact digest inputs; T13 label-only check kept and characterized; §9.1/§9.2 edit | 006, 007, 009 | 006, 007, 009 |
| M3 | High | **Audit kickoff decider**: `DeciderAudit`; edge T8a kickoff→run that leases the card to its record owner in the same transaction (kickoff holds no lease today, `fr_lease_test.go:148-162`); guard re-checks predicate, `audited_sha`, digest incl. decision-index, audit-ready, blocker/hold, any empty FOUNDER verdict; `factory decide` lane admission when the lane is the record owner; Amendments row on SPEC-FACTORY-SELF-DISPATCH-001 | 014-016 | 014-016 |
| M4 | High | **Ceiling policy**: config key; auditor/spec-workflow text; procedure for every session (mechanical delta eligibility, final hit = ceiling + rounds, non-lane notice); release-blocking AC-wording exception | 010-013 | 010-013 |
| M5 | Medium | **Authority register + FOUNDER defaults**: manager-spec (C2 first, C1, `make agents-emit` for C3) incl. the clause narrowed to judgment calls; kickoff step; pin format; Class/Default/Alternate; product-level definition; DEFAULT-APPLIED | 005, 017, 018 | 005, 017, 018 |
| M6 | Medium | **Watchdog + doctrine wiring**: board read at step ②, record-before-message; binding run conditions copy + re-read by `sync-audit-4dim.js` and sync-auditor (undisposed = FAIL); wait ids + `resolves`; one-shot recheck; MCP intake comparison | 004, 008, 019, 023 | 004, 008, 019, 023 |
| M7 | Medium | **Hook messaging hygiene**: bind cache behind the run-state probe, retirement invalidation, degraded warn log + rate-limited notice | 020-022 | 020-022 |
| M8 | Low | **Mirror + regression sweep**: template mirrors, `make build`, `make agents-emit`, keep-set regression guard | 024, 025 | 024, 025 |

Decisions Q1-Q23 were settled by the leader on 2026-10-03 (decision-index.md).

Dependencies: M0(a) before M7; M0(b) before M6 fixes the default delay. M2 before M3 (the guard uses
the predicate), M4 (verdict fields), and M6 (sync re-read). M1 before M5 (pins cite board records) and
M6. M7 is independent of M1-M6.

## §G Risks

| Risk | Mitigation |
|---|---|
| A lane self-approves Kickoff on a forged verdict file or a post-audit decision-index edit | the guard binds `audited_sha` to the evidence SHA, recomputes the digest (decision-index included), and refuses on any open blocker, hold, or empty FOUNDER verdict; the sync audit re-reads the decision record |
| T8a lease write diverges from the normal lease path | T8a reuses the `guardLeaseAcquire` writes and registered-worker check; AC-FDA-015 runs the real chain |
| T7 behavior change breaks factory tests | characterize first (M2); the change is deliberate and named in REQ-FDA-009 |
| Board becomes a second queue | closed kind enum; no card-creating kind; queue verbs untouched |
| One-shot recheck multiplies cache writes | re-armed only while the wait is open; 5-minute floor; M0(b) measures the cost |
| Bind cache serves a stale binding | the probe runs on every bind; a retired or changed run invalidates and rebinds in the same invocation |
| Default read as a recommendation | the published rule is stated on the row; unrankable rows carry no Default and block |
| The exception becomes a routine bypass | three mechanical conditions plus a leader board record; any missing piece keeps the hold |

## §H Anti-Patterns

- Judging delta eligibility by reading prose instead of the `fix_scope` diff and id sets.
- Writing a ruling only in chat (REQ-FDA-004).
- Relaxing FAIL/INCONCLUSIVE anywhere.
- Skipping the run-state probe on a cache hit.
- Hand-editing `internal/template/templates/.codex/agents/moai/*.toml`.

## §I Cross-References

spec.md §C, design.md §1-8, research.md probe ledger, acceptance.md §D, decision-index.md,
`.moai/reports/t1481/plan-audit-iter1.md`.

# plan.md — SPEC-FACTORY-DECISION-AUTO-001

## §A Context

Card t1481, Tier L. Eight capabilities that turn the leader's recurring rulings into rules or disk
records (spec.md §B). Base tree `d7112d005`. Methodology per `quality.yaml` (TDD for new Go code;
DDD/characterization for the three verdict sites and the hook path).

## §B Known Issues

- Doctrine and code disagree on PASS-WITH-DEBT (research F2) and on factory Kickoff (F4); either
  side may be "right" today — this plan makes both follow one predicate.
- The FOUNDER Default marker amends a [HARD] manager-spec clause (decision-index Q2).

## §C Pre-flight

- `go test ./internal/homestate/... ./internal/contract/... ./internal/hook/...` green at base
  (record exit codes in progress §E.2 before M1).
- `make agents-emit-check` clean at base.

## §D Constraints

Template-first; template neutrality; keep-set untouched (REQ-FDA-024); lane refusal relaxed only
for REQ-FDA-017; no local full-suite runs.

## §E Self-Verification

Each milestone closes with its ACs' green cells (acceptance.md §D), scoped package tests, `go vet`,
`golangci-lint run` on changed packages, and for template milestones `make build` +
`make agents-emit` + `make agents-emit-check`.

## §F Milestones (ordered by decision-reversibility — highest change likelihood first)

| M | Priority | Scope | REQs | ACs |
|---|---|---|---|---|
| M1 | High | **Board data model + CLI**: record shape, kinds, location, append/lock, `record`/`read`, lane refusal, statuses | 001-004 | 001-004 |
| M2 | High | **Verdict admission predicate**: shared function, three call sites, auditor verdict-block fields (`blocking_findings`, `debts`, `delta_eligible`), §9.1/§9.2 doctrine edit | 008, 009, 011 | 008-010 |
| M3 | High | **Audit kickoff decider**: `DeciderAudit`, edge kickoff→run keeping lease, hash recompute, `factory decide` lane admission | 015-017 | 014-016 |
| M4 | High | **Ceiling policy**: config key + defaults, auditor/spec-workflow text reconciliation, lane procedure (delta round, hold + split proposal) | 012-014 | 011-013 |
| M5 | Medium | **Authority register + FOUNDER defaults**: manager-spec contract (C1/C2), plan/spec-assembly kickoff step, pin format, Class/Default/Alternate rows, DEFAULT-APPLIED | 007, 018, 019 | 007, 017, 018 |
| M6 | Medium | **Watchdog + doctrine wiring**: watchdog reads board at step ②, messages-as-nudges rule, binding run conditions copy + sync-auditor re-read, short recheck carrier, MCP intake comparison | 005, 006, 010, 020, 023 | 005, 006, 019, 022, 023 |
| M7 | Medium | **Hook messaging hygiene**: bind cache, degraded warn log + rate-limited notice | 021, 022 | 020, 021 |
| M8 | Low | **Mirror + regression sweep**: template mirrors, `make build`, `make agents-emit`, keep-set regression guard, codex emit check | 024, 025 | 024, 025 |

Dependencies: M2 before M3 (guard uses the predicate) and before M6's sync re-read; M1 before M5
(pin cites board records) and M6. M4 needs M2's verdict fields. M7 is independent.

## §G Risks

| Risk | Mitigation |
|---|---|
| A lane self-approves Kickoff on a forged verdict file | the guard recomputes the plan-artifact hash and reads the verdict file from the committed card evidence path; the sync audit re-reads the decision record |
| Board becomes a second queue | closed kind enum; no card-creating kind; queue verbs untouched |
| Short recheck multiplies cache writes | carrier exists only while a wait-on-leader is open; cadence configurable |
| Bind cache serves a stale binding after run retirement | full four-field match + run-state check on miss; any bind error deletes the cache file |
| Default marker read as a recommendation | reversibility rule stated on the row; unrankable rows carry no Default and block |

## §H Anti-Patterns

- Lane deciding delta eligibility itself instead of reading the auditor's field.
- Writing a ruling only in chat (REQ-FDA-006).
- Relaxing FAIL/INCONCLUSIVE anywhere.
- Hand-editing `internal/template/templates/.codex/agents/moai/*.toml`.

## §I Cross-References

spec.md §C, design.md §1-8, research.md probe ledger, acceptance.md §D, decision-index.md.

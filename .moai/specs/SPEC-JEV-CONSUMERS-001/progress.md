# SPEC-JEV-CONSUMERS-001 — Progress

Card: t1020 · Tier L · plan-phase artifacts authored 2026-09-20; revised 2026-09-20 (revision 0.2.0). Split from `SPEC-JEV-INTEGRATION-001` on the M4+M5+M6 seam.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 15 / ceiling 25; AC 15 / ceiling 25 |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 15 (`REQ-JEVN-001` … `REQ-JEVN-015`) |
| Acceptance criteria | 15 (`AC-JEVN-001` … `AC-JEVN-015`) |
| Revision | 0.2.0 — answers the iteration-1 plan-audit verdict |
| Audit verdict answered | `.moai/reports/t1020/plan-audit-SPEC-JEV-CONSUMERS-001.md` (iteration 1/3, FAIL 0.64 against the Tier L 0.85 threshold) |
| Coordinate baseline | every coordinate cited across the six artifacts re-verified at HEAD `7e1ed63b9` |
| Predecessor | `SPEC-JEV-OPTIN-MEASURE-001` |
| Successor | `SPEC-JEV-GOAL-DIST-001` |
| Status transition | (none) → draft (unchanged by this revision) |

### What revision 0.2.0 changed

- **N1 (dedup precedence) is SETTLED**, confirmed by the operator exactly as proposed, and is recorded as a decision in `plan.md` §B1 and `design.md` §2 rather than as an open question. It leaves the Kickoff-gate carry list.
- `REQ-JEVN-006` restated: the rule's two halves are separated, half (a) is shown to require a **new source-agnostic unordered predicate** (`HasFindingForPairAnySource`) that M4 now budgets, and the claim that `AppendFindingOnce` / `SamePairAs` enforce it is withdrawn as false.
- `REQ-JEVN-015` added: the disposition for a consumer whose gate **cannot be run**, distinct from one whose gate ran and failed. `AC-JEVN-015` asserts the distinction.
- `REQ-JEVN-001` scoped explicitly to card admission, excluding the `moai todo analyze` re-sweep; `AC-JEVN-013` asserts it.
- `AC-JEVN-012` added (write-path `Source: agent` absence with a positive control), `AC-JEVN-014` added (the ranked-signal presentation half of `REQ-JEVN-013`).
- `AC-JEVN-003` widened to all four source combinations in both arrival directions; `AC-JEVN-008` / `AC-JEVN-011` given call-path-exists preconditions; `AC-JEVN-007` / `AC-JEVN-010` gated on the fitted threshold's existence with a stated blocked disposition; the Definition of Done scoped per consumer; every AC now names its REQ ids.
- Three write-path coordinates corrected as **miscitations, not drift** — they resolved at neither HEAD `7e1ed63b9` nor the declared baseline `fd75cf692`, and no commit touched the files between the two trees.

### Open questions carried to the Implementation Kickoff Approval gate

- **N2 — which code path hosts Consumer A, and what verification surface covers Consumer B.** OPEN and **blocking M5 and M6**, which are declared blocked in `plan.md` §F. Not covered by the 2026-09-20 operator decisions. This revision deliberately does not name a host for either consumer: an unmeasured host would be exactly the unverified-premise defect this chain is bounded against. N2 replaces N1 on this list.
- **Q3 — labelled-set size per consumer.** OPEN, owned by `SPEC-JEV-OPTIN-MEASURE-001`.

Not an open question, recorded here because it shapes what run-phase will produce: with live measurement excluded from this batch, `Report.Verdict()` (`internal/jevmeasure/measure.go:193`) withholds every consumer, so M4's expected terminal state is the `REQ-JEVN-015` recorded decision rather than a shipped consumer.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — plan (v0.1.0)

Priority-ordered, no time estimates. Development mode per `quality.yaml`: TDD (RED first). Each milestone is its own commit with subject `feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> <what> (card t1872)` (M0 uses `docs(...)`), carrying the `🗿 MoAI` trailer; the card id in every message is a traceability carrier.

## A. Review-focus order (decision reversibility)

A reviewer should spend attention in this order; it is not the build order (§C).

1. **The record and verifier contracts** (manifest, attempt, reuse-record field sets; REQ-LVM-001..006, 016..020) — highest change likelihood, data-model changes, the one place a wrong choice weakens landing evidence. Review before any code.
2. **Defaults and the enabling decisions** (REQ-LVM-007, 011/012; decision-index Q7, Q8, Q9) — operator-visible landing policy.
3. **The ready view's predicates** (REQ-LVM-014/015) — a user-facing flow for the leader.
4. **Slot record shape** (capacity, `holders[]`) — a persisted-format change, backward-compatible by construction.
5. **Instrumentation events** — additive, fail-open, easily revised.
6. **Rules and mirrors** (M7) — mechanical.

## B. Preconditions and constraints

- Git-flow window path only (spec C1). Heavy suites run under the `go-test-heavy` lease with one acquire, one command, immediate release; scoped `-run` and package-scoped lint otherwise (the box's measured budget). No full `internal/cli` coverage run.
- Work in the card worktree; commit staging by explicit pathspec; no push; no file under `.moai/reports/` is ever force-added (the operator directive) — the M0 baseline is a tracked file under the SPEC directory.
- Gate enabling, capacity raising, and reuse opening in this repository's own config are operator acts and are not part of any milestone (decision-index Q7, Q8, Q9).
- Docs and rule files that are loaded into the session prefix are edited last (M7), per the cache-aware batching directive.

## C. Milestones in dependency order

| M | Scope | REQ | Flips | Depends on |
|---|---|---|---|---|
| M0 | Baseline commit: re-run the 28 RED-now cells and the landing-interval and eligibility measurements on the run tree; commit `evidence/m0-baseline.md` (tracked, never under `.moai/reports/`) | — | baseline (AC-LVM-008 ordering leg) | — |
| M1 | **Completeness gate**: manifest generator, attempt records, abort/restart classes, whole-manifest verifier, reject classes, single-verifier wiring into the merge step, `complete`, condition (d); `workflow.landing_gate.manifest_required` (false) and the narrowing allowance list (`internal/cli` only) | 001-007 | AC-LVM-001..007 | M0 |
| M2 | Dead-owner auto-release in `slot status`, `slot acquire`, the guard | 010 | AC-LVM-010 | M1 |
| M3 | Slot capacity, headroom admission, wall-clock-bound exclusion; config keys | 011-013 | AC-LVM-011..013 | M2 |
| M4 | Landing-event log and metrics verb, including the shadow reuse-eligibility counter | 009 | AC-LVM-009 | M1 |
| M5 | Ready view verb (derived), audit-HEAD predicate | 014, 015 | AC-LVM-014, 015, and the ready leg of 006 | M1, M4 |
| M6 | **Reuse (opens last)**: scope inventory, undecidable classes, reuse record, consumer recompute, no chaining, safety invariants, inertness without the gate | 008, 016-021 | AC-LVM-008, 016..021 | M1..M5 |
| M7 | Rules and mirrors (`factory-dispatch-mechanics.md`, `resource-slot-lease.md`) edited in the template tree first, `make build`, local copies regenerated | 022 | AC-LVM-022 | M1..M6 |

The ordering M1 < M2..M6 and M5 < M6 is enforced twice: by the commit-graph check of AC-LVM-008, and functionally by REQ-LVM-008 (reuse refuses without a gate-complete source), so a merge-order mistake cannot open the relaxation.

M2/M3 (slot) and M4 (metrics) are independent of each other and of the gate's internals; they follow M1 only to honour the leader's ordering. If the SPEC is split (design §9), M2/M3 move to SPEC B unchanged.

## D. Per-milestone method

Each milestone: (1) write the AC's tests first and observe them red for the stated reason (the names exist, the behaviour does not) — the RED output goes into the milestone's evidence per `tdd-result-contract`; (2) implement the minimum; (3) refactor; (4) run the milestone's AC commands with the empty-sweep guard; (5) one commit.

- **M1 specifics.** Manifest derivation uses `go list` over the card's changed-package set against the merge-base, with importers from `Imports`, `TestImports`, `XTestImports`. The verb that runs a manifest call extends `moai integration remeasure` (RunRemeasure keeps its clean-start and clean-finish checks, `integration_remeasure.go:634-720`). The verifier is a new function consulted where `ValidateRemeasureRecord` is consulted today (five call sites, `research.md` §2); with the key false it is not called, which is how REQ-LVM-007 holds byte-for-byte. Fixtures use temporary repositories and canned `go test -json` streams; no fixture touches this repository's refs.
- **M2 specifics.** The probe of `research.md` §4 becomes the regression test: a dead-pid record, `status`, record gone, audit row present.
- **M3 specifics.** Load and memory readers are injected; the production readers are `sysctl vm.loadavg`/`/proc/loadavg` and the platform memory source with a fail-closed unreadable path (including Windows, where the readers return unreadable and capacity stays 1).
- **M4 specifics.** Writes use the existing atomic-append helpers; a write error is swallowed with a one-line stderr advisory (fail-open).
- **M5 specifics.** The view is read-only: it takes no lock, writes no file, opens no window.
- **M6 specifics.** `go list -deps -test` runs with the source record's fingerprint; failures are classified, never retried silently.
- **M7 specifics.** Template-First: edit `internal/template/templates/.claude/rules/moai/workflow/*.md`, `make build`, confirm the local copies match (`cmp`), commit once.

## E. Risks

| Risk | Mitigation |
|---|---|
| The gate refuses landings the lanes currently complete (script-wrapped calls) once enabled | ships off; enabling is operator-held (Q7); lanes get a documented per-call invocation form in M7 |
| Reuse eligibility is near zero (measured 1/37 strict) so M6 buys little | M4 ships first and records the shadow counter; M6 is gated by Q8 and may be deferred or dropped without affecting M1-M5 |
| Hermetic declarations are wrong (R3) | default-deny; batch CI backstop; declaration list reviewed per entry |
| Capacity 2 raises machine load and flakes tests | headroom admission fail-closed; wall-clock-bound exclusion; capacity ships 1 |
| Plan audit hits the Tier L ceiling | split recommendation (design §9) |
| `internal/cli` is a 1,400-file package; new CLI tests compile slowly | verbs are thin; logic lives in `internal/factory`; CLI tests assert wiring only |

## F. Cross-references

`spec.md`, `acceptance.md` (RED-now ledger), `design.md` (threat analysis, split), `research.md`, `decision-index.md`. Rules touched at M7: `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing and § Serializing a heavy run across lanes; `.claude/rules/moai/workflow/resource-slot-lease.md`.

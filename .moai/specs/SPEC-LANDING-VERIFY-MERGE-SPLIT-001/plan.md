# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — plan (v0.2.0, piece A)

Priority-ordered, no time estimates. Development mode TDD (RED first). Each milestone is its own commit with subject `feat(SPEC-LANDING-VERIFY-MERGE-SPLIT-001): M<k> <what> (card t1872)` (M0 uses `docs(...)`), carrying the `🗿 MoAI` trailer; the card id in every message is a traceability carrier. Scope is piece A only (board `9dbe`); follow-ups are listed in `spec.md` §7 and not planned here.

## A. Review-focus order (decision reversibility)

1. **The record and verifier contracts** (manifest, attempt, verdict; REQ-LVM-001..006) — data-model decisions, highest change likelihood, and the place a wrong choice weakens landing evidence. Review before any code.
2. **The audit binding** (REQ-LVM-008/009): what "current HEAD" and "receipt at that HEAD" mean against the existing verdict and receipt stores.
3. **The switch** (REQ-LVM-011): operator-visible landing policy and its transition cost on the 16 merge-ready cards.
4. **Lane guidance text** (REQ-LVM-010).
5. Wiring edits in the merge step, `complete` and condition (d) — mechanical.

## B. Preconditions and constraints

- Git-flow window path only. Heavy suites run under the `go-test-heavy` lease with one acquire, one command, immediate release; scoped `-run` and package-scoped lint otherwise. No full `internal/cli` coverage run.
- Work in the card worktree; explicit pathspecs; no push; nothing under `.moai/reports/` is ever force-added — the M0 baseline is a tracked file under the SPEC directory.
- Rule files loaded into the session prefix are edited last before the switch (M3), per the cache-aware batching directive.
- Gate tests use temporary repositories and canned streams.

## C. Milestones in dependency order

| M | Scope | REQ | Flips | Depends on |
|---|---|---|---|---|
| M0 | **Baseline, alone and first.** Re-run the 14 RED-now observations and the landing-interval measurement on the run tree; commit `evidence/m0-baseline.md` (tracked, never under `.moai/reports/`) as the only file of the commit | — | AC-LVM-011 (M0 leg) | — |
| M1 | **Gate core.** Manifest generator (b2cd derivation, reader search, 35e9 narrowing only for `internal/cli`), attempt records and the structured form of the re-measure verb, abort/restart classes, the whole-manifest verifier and its eleven reject classes, wiring into the merge step, `complete` and condition (d); config key `workflow.landing_gate.enabled` (false) and the narrowing allowance list. Negative tests and the positive control are written first and end green | 001-007, 012 | AC-LVM-001..007, 012 | M0 |
| M2 | **Audit binding (Q13).** `complete` binds the sync-audit verdict to the card's current HEAD and requires the required backend's receipt at that HEAD to pass, both under the gate flag | 008, 009 | AC-LVM-008, 009 | M1 |
| M3 | **Lane guidance.** `factory-dispatch-mechanics.md` (and its mirror) describe the gate, the structured form and the per-call procedure; edited in the template tree first, `make build`, local copies regenerated | 010 | AC-LVM-010 | M1, M2 |
| M4 | **The switch.** `.moai/config/sections/workflow.yaml` gains `landing_gate: enabled: true` — the only path in the commit | 011 | AC-LVM-011 (full) | M0..M3 |

The order M0 < M1 < M2 < M3 < M4 is the leader's order inside A (structured form and guidance first, tests green, then the key). It is machine-checkable twice: by the commit-graph and diff commands of AC-LVM-011, and by the S3→S4 diff being configuration only, which makes the tests green on the M4 tree witness them green before the switch.

## D. Per-milestone method

Each milestone: (1) write the AC's tests first and observe them red for the stated reason (names exist, behaviour does not), keeping the RED output as milestone evidence per the TDD result contract; (2) implement the minimum; (3) refactor; (4) run the milestone's AC commands with the empty-sweep guard; (5) one commit.

- **M1.** Manifest derivation: changed packages from the diff against the merge-base (a changed file inside a package directory, embed and testdata included, marks that package changed), importers from `go list` `Imports`, `TestImports`, `XTestImports`; the expected package set is resolved by `go list` at derivation so "zero missing" has a denominator. The structured form extends `moai integration remeasure` (keeping `RunRemeasure`'s clean-start and clean-finish checks, `integration_remeasure.go:634-720`). The verifier replaces the decisions of `ValidateRemeasureRecord` where it is consulted (five call sites, `research.md` §2) only while the key is true, which is how REQ-LVM-007 holds byte-for-byte. A package with no test files is named in the manifest with its reason rather than skipped silently.
- **M2.** Reuse the verdict and receipt readers that already exist (`factorylane/merge_specless.go` for the audited-SHA bind and its evidence-directory allowance, `internal/auditreceipt` and `runtime.ResolveRequiredBackends` for the receipt); the run phase first confirms how receipts are keyed to a commit and records it in progress before coding (Gap in `design.md` §4).
- **M3.** Guidance states: derive the manifest; run each call as its own structured invocation; an aborted call is restarted as a new attempt; verify the verdict before `complete`; a script-wrapped record no longer satisfies the gate.
- **M4.** A one-file commit. The card's own landing is the first live use of the enabled gate.

## E. Risks

| Risk | Mitigation |
|---|---|
| Enabling refuses the 16 merge-ready cards until their calls are re-run in the structured form | decided cost (Q7); M3 guidance precedes the switch; fail-closed is the intent |
| `go list` derivation misses a reader of a changed non-Go file | the reader search is recorded with its command; a computed-path read stays an undetectable residual (`design.md` §3) |
| Receipt keying differs from the assumed commit key | confirmed at the start of M2; if it differs, the AC wording is amended through manager-spec before code |
| The gate lengthens the lane's landing sequence by one call per manifest entry | intended; wall-clock relief is follow-up work, not this piece |
| Plan audit holds at the ceiling | the narrowed scope halves the requirement count; split further only through the leader |

## F. Cross-references

`spec.md`, `acceptance.md` (RED-now ledger), `design.md`, `research.md`, `decision-index.md`. Rules touched at M3: `.claude/rules/moai/workflow/factory-dispatch-mechanics.md` § The lane's standard landing and its template mirror.

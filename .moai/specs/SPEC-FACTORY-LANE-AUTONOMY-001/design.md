# design.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

> Design decisions D1..D8. Card-inherent choices take the dispatch's recommended option; each
> records the rejected alternative and why. The author owns the final wording.

## D1 — Fallback detection surface: Go probe + doctrine wiring (RECOMMENDED, adopted)

**Chosen**: a small Go probe surface (a `factory`/sessionmsg subcommand exposing channel liveness:
registered-peer listing via the session registry, lead heartbeat age, and a bounded no-response
timer) plus doctrine wiring naming when a lane declares fallback. The switch ACT itself is the lane
invoking `/moai:todo --auto` with a recorded transition event (REQ-FLA-003).

**Why**: the failure is quiet today — `internal/sessionmsg/` is a poll-based store with NO
availability API (research.md §R2), and the doctrine's answer is "surface the constraint to the
operator", which is exactly the human round-trip the fallback exists to remove. A mechanical probe
makes "unavailable" a checkable predicate instead of an agent's impression, and keeps REQ-FLA-001
binary-testable.

**Rejected**: doctrine-only (no Go) — a lane would declare fallback by text-judgment, which is not
mechanically testable, not auditable, and contradicts the record-is-the-source-of-truth direction
of F1. Also rejected: making `--auto` itself probe — that couples the t1306 foreman's internals to
messaging state this SPEC must not own (spec.md §F exclusion).

**Boundary note**: the probe answers "is the channel available"; it does NOT send anything. The
bounded no-response timer (REQ-FLA-002) bounds the second failure shape — channel nominally up,
lead unresponsive — and joins t1241's stall determination for the card-resumption side (D6).

## D2 — Jev classification consumption: consumer-only, milestone-gated on t1332 (RECOMMENDED, adopted)

**Chosen**: this SPEC defines CONSUMPTION ONLY (REQ-FLA-008). The consumer is implemented against a
declared minimal consumption interface (read priority + execution axis; absent ⇒ fallback
REQ-FLA-007). M2 is milestone-gated: if t1332's SPEC does not exist at M2 entry, the consumer lands
behind the Where-fallback default with the interface stub documented — never the producer.

**Why**: `BacklogItem` today carries NO priority field and NO sequential/parallel axis field
(research.md §R2, `internal/kanban/backlog_store.go:83`); defining the field schema here would
silently become t1332's producer and collide with its SPEC. Tolerant consumption
(absent-metadata ⇒ fallback) makes ordering safe in both sequences.

**Rejected**: joint minimal field contract defined here (would pre-empt t1332's design and create
two schema SSOTs); blocking M2 entirely on t1332 (needlessly serializes implementable consumer
logic that is metadata-optional by construction).

## D3 — Lane-direct merge: condition triple as lane-executed checks over the existing window (RECOMMENDED, adopted)

**Chosen**: adopt t1241's condition triple — sync-audit PASS · no unresolved conflict · tree
identity `HEAD^{tree} == HEAD^2^{tree}` — as the lane-direct merge precondition, executed BY the
lane as a check sequence over the existing `moai integration acquire`/`release` window
(`internal/cli/integration.go:271/325/445`). No new serialization mechanism (REQ-FLA-010/011).

**Why**: the window exists and is the recorded serialization point; the triple exists as t1241's
normative interface (research.md §R5). Lane-executed checks deliver fragment 3 without the F3 M3
controller, honoring the umbrella non-overlap.

**Rejected**: waiting for F3 M3 machinery (scope explosion — that is t1241's build, explicitly
excluded); a lane-local lock (would create a second serialization point and break the serial
invariant the card demands be maintained).

**Run-phase ordering**: M3 entry requires t1240 merged (card predecessor) AND t1241's SPEC
landed-enough that the triple is normative; the triple is treated as an interface and re-pinned
from t1241's landed SPEC text at M3 pre-flight (plan.md §F).

## D4 — Post-push disposal: machine check on the `worktree done` path; CI-green EXCLUDED (RECOMMENDED, adopted)

**Chosen**: extend the `worktree done` path (or a thin wrapper verb) with a machine check —
`git fetch origin develop` + `git rev-list --count --left-right origin/develop...<merge-commit>`
confirming the card's merge commit reached origin — refusing disposal before it, allowing
unattended (`--auto`) disposal after (REQ-FLA-012..015). **CI-green is NOT in the machine
precondition.**

**Why excluded (the plan decision the card delegated)**: CI judgment is lead-side under the
operating doctrine (`CLAUDE.local.md` §4.1 — lanes read early signals; the lead reads CI), and
`worktree done` has zero origin awareness today (research.md §R2), so origin-landing is the
machine-checkable boundary this path can own without inventing a lane-side CI reader. The L1 guard
(`isL1WorktreePath` :192) and anchored-session guard (`LiveAnchoredSessions`) remain untouched
(REQ-FLA-013) — this SPEC adds a third precondition, it does not weaken the first two.

**Rejected**: including CI-green (duplicates lead-side judgment inside a disposal verb and needs a
`gh` dependency in `worktree done`); a separate disposal verb bypassing `done` (a parallel
disposal path is exactly the guard-bypass shape REQ-FLA-013 forbids).

## D5 — Operator-gate reduction table: enumerated, and bounded (REQUIRED by card constraint)

The full table lives in plan.md § Operator-Gate Reduction Table. Shrink exactly three
involvements — lead window grant → lane-direct merge under the triple; lead disposal approval →
machine-gated `--auto` disposal; dispatch messaging → lane self-service `--auto` under declared
fallback. Never shrink: card admission (the queue stays the channel), Implementation Kickoff
Approval, irreversible/destructive confirmations. Reasoning: each shrink replaces a HUMAN
coordination act with a MECHANICAL check (probe, triple, rev-list), never with another agent's
judgment — the verification-claim-integrity direction.

## D6 — Resumed ownership for picked-in-progress stalled cards (fragment 1 sub-clause)

**Chosen**: t1241's stall/re-alert determination is the trigger interface (this SPEC does not
define "stalled" — that is t1241 M5/M6's concern, pending); the resuming lane's obligations are
REQ-FLA-004/005 — read `progress.md` + recorded evidence FIRST, continue from the recorded phase,
append-only records. The lease machinery it rides on exists (F1:
`applyLeaseExpiry`/`RenewLease`, `internal/homestate/card_transition.go:339/368`) and the F1 verbs
`factory handoff recover-resume / abandon-lane` are the resume surface.

**Why**: silent restart is the destructive case — the previous lane's unpushed work and recorded
evidence would be discarded. Resume-from-record matches the F1 evidence-gate philosophy.

**Rejected**: fresh-pickup semantics (discards work; contradicts the "원격 착지 전 폐기 금지"
family of invariants this SPEC exists to mechanize).

## D7 — Discovery note routing (lead's 부고)

The `moai todo add "-f lane …"` flag-parse failure (research.md §R-D1) is recorded as a discovery
item with a routing suggestion (t1330 arg-parsing scope or a one-line Class A card) and is Out of
Scope here (spec.md §F). Not actionable by this SPEC.

## D8 — Run-entry predecessor expressed as a milestone gate, NOT `depends_on` (card constraint + dispatch decision 8)

**Chosen**: `depends_on` is OMITTED entirely. SPEC-FACTORY-SELF-DISPATCH-001 (and equally
SPEC-FACTORY-CONTROLLER-001, which has the same problem) is expressed as: an explicit run-entry
precondition (plan.md §F Milestone Gate M0: t1240's develop merge confirmed mechanically) plus
`related_specs` entries.

**Why (verified this run)**: both SPEC directories are absent from develop's tree (`.moai/specs/`
ls in this worktree, this run) — a `depends_on` entry naming either would make the run-gate
preflight attempt to read a SPEC dir that does not exist there and fail. `related_specs` is
non-blocking by definition and carries the reference safely. The predecessor remains binding
through the M0 gate, so the card's "선행: t1240 병합" constraint is enforced, not relaxed.

**Rejected**: putting SPEC-FACTORY-SELF-DISPATCH-001 in `depends_on` (preflight failure by
construction); dropping the predecessor entirely (violates the card).

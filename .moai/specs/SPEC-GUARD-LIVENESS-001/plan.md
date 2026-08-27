# SPEC-GUARD-LIVENESS-001 — Implementation Plan (card t333, surfacing model)

Baseline tree for every measurement: **`091966c55`** @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`).

Milestones are ordered by **decision reversibility**. The invocation contract leads, because it is what §D.1's answer to self-observation rests on and a wrong choice there cannot be repaired without redoing the surface.

## §A Context

See `spec.md` §A. Six empirical instances ground the problem; this SPEC owns the **surfacing half** of the event-history axis after the iter-3 scope reduction.

### A.1 The sibling SPEC

The state model — the manifest, the querying, the classifications, and the set comparison — is `SPEC-GUARD-STATE-MODEL-001`, card **t347**, authored alongside this SPEC as plan-phase artifacts.

This SPEC does **not** `depends_on` it. The seam is a consumed contract (`spec.md` §B.1), so the two are independently implementable — and they have different destinations: **this artifact set is finishable on its own** and goes to the Implementation Kickoff Approval gate, while t347's dispatch is a separate decision. Nothing in this set requires a t347 artifact to exist: the only cross-references are to the sibling SPEC's ID and to the contract clauses this SPEC declares for itself in REQ-GDL-001.

## §B Known issues and constraints carried in

- **B-0 — the defect is absent execution, not suppressed failure.** Nothing runs, so nothing can go red, and a mechanism reports success accurately about a set that had silently become the wrong set (`spec.md` §A.0). Do not implement or describe any part of this as "the check should have failed and did not" — there is no failure to suppress. A hidden-failure framing leads to building surfacing and routing for a signal that does not exist.
- **B-1 — the self-observation constraint.** A watcher that cannot prove its own firing is forbidden. Answered by making the evaluator pull-based rather than scheduled, and the answer's load-bearing half is REQ-GDL-003's unconditional invocation (`spec.md` §D.1).
- **B-2 — the unprompted-discoverability constraint.** A targeted query can only be issued by someone who already suspects the answer, so a design answering only when queried relocates the problem into whoever is expected to already know (`spec.md` §A.4, §D.2). Independent of B-1: a design can pass one and fail the other.
- **B-3 — the trigger must not enumerate.** Three audit iterations failed on one shape: the trigger listed symptoms and a run reached the same silence down an unenumerated branch (D1, then N1, then T3). The contract seam makes enumeration unrepresentable; the implementation must not reintroduce it by hardcoding value names (AC-GDL-002).
- **B-5 — the t326 citations are pinned to a different tree, and reading the wrong one inverts the finding.** `origin/develop` at `ec15ec2cd` is **diverged** from this SPEC's baseline (diverged, `merge-base --is-ancestor` false; `git merge-base --is-ancestor` returns false). t326's session-start advisory and its comparison package exist there and are **absent from the baseline tree**, so a check run here reports a landed feature as missing. Every t326 citation names its tree inline (`spec.md` §A.10); RED-now cells stay pinned to `091966c55` because they measure this deliverable's absence.
- **B-6 — the host surface is latency-bounded, so render and refresh are separate acts.** t326's comparable path bounds itself at 250 ms and can afford an inline comparison because it is two short local `git` invocations. This SPEC's evaluator issues one forge query per subject — 18 here — and no sequence of network round-trips fits that bound. The advisory reads a persisted result (REQ-GDL-011); wiring the queries into the host surface is a latency defect that only appears on a slow network.
- **B-4 — the always-red neighbour is real and is not repaired here.** `Graph Freshness` fails on every `develop` push (card t322). A new advisory rendered beside it inherits its trained filter; REQ-GDL-007 is the mitigation and it is partial.

## §C Pre-flight (measured on `091966c55`)

| Check | Command | Result |
|---|---|---|
| No advisory or evaluator wiring | `grep -rln "guard-liveness\|guardLiveness" .claude/hooks/ internal/` | no output, rc=1 |
| No `moai guard liveness` verb | `grep -n '"guard"' internal/cli/*.go \| grep -v _test.go` | `internal/cli/constitution.go:49` only — the unrelated `moai constitution guard` verb. Unnarrowed, the glob also returns `constitution_guard_test.go:93`, same unrelated verb |
| Rule carries no continued-firing clause | `grep -nE "last.fired\|continued.firing\|stopped firing\|liveness\|stale guard" .claude/rules/moai/development/verification-completeness.md` | no output, rc=1 |
| Scheduled-workflow baseline | `grep -l '^  schedule:' .github/workflows/* \| wc -l` | `3` |
| Evidence artifact tracked | `git log --oneline -1 -- .moai/reports/t333/trigger-axis-observation.md` | `c30f761dd` — the SPEC's citations resolve from the branch |

## §D Constraints

- No scheduled workflow is added (REQ-GDL-002; the baseline of 3 is asserted at merge).
- The advisory path performs no forge mutation and no working-tree write (REQ-GDL-008); result persistence lives outside the working tree.
- No existing text in `verification-completeness.md` is modified (REQ-GDL-009).
- No classification value name appears anywhere in this deliverable's source (AC-GDL-001(b), AC-GDL-002(b)).
- No file under `internal/template/templates/` is touched — this is repository-specific wiring, not template content.

## §E Self-verification

Every criterion in `acceptance.md` carries a RED-now cell pinned to `091966c55` and a green-path cell naming its flipping milestone. **No cell was carried across the scope reduction without re-running its command** — the audit's D7 finding is that an unreproducible measurement inside a RED-now cell is precisely the evidence-integrity failure the two-cell discipline exists to prevent.

## §F Milestones

### M1 — the invocation contract (least reversible)

The decision §D.1 rests on, and the one worth the most review attention.

- Wire the evaluator's invocation into an already-attended surface, pull-based, no scheduled workflow (REQ-GDL-002).
- **Invocation is unconditional on that surface's activation** — no path filter, no changed-file test, no subject-matter condition (REQ-GDL-003). This is the clause the audit found verified by nothing; the surviving mutant was an evaluator gated on whether the session diff touched `.github/workflows/`, which is `docs-i18n-check.yml` rebuilt inside the deliverable.
- Consume the classification contract without naming any value (REQ-GDL-001).

Flips: AC-GDL-001, AC-GDL-003, AC-GDL-004.

### M2 — the trigger and its arrival

- Render when any entry is non-clean — the partition, not a list (REQ-GDL-004). If implementation appears to need a value name, that is B-3 resurfacing; re-derive rather than enumerate.
- The advisory arrives with no operator-supplied guard identifier or query (REQ-GDL-005). A documented CLI verb is not an acceptable sole answer — it is still a question someone must know to ask.
- The stated age is derived from the persisted result's own timestamp (REQ-GDL-006).

- Join the **existing** session-start additional-context block by calling the established contributor helper from an already-registered handler; register no new handler (REQ-GDL-010). The baseline to hold: **7** non-test files under `internal/hook/` register a session-start handler, measured on `origin/develop` at `ec15ec2cd`.
- Read a **persisted** result at the host surface; issue no forge query inline; **declare a render join bound ≤ 250 ms as a named constant** (REQ-GDL-011, B-6). No shared bound exists on that surface — each contributor carries its own, so an undeclared bound is an unbounded one.
- Initiate the refresh on every activation and **never await it**; persist its result for a later activation (REQ-GDL-012). This is what makes REQ-GDL-003's unconditional binding compatible with the latency bound.
- Identify clean entries by reading the result's carried designator, never by a value literal and never by the surface fold (REQ-GDL-001 (iii)/(iv)).

Flips: AC-GDL-002, AC-GDL-005, AC-GDL-006, AC-GDL-010, AC-GDL-011, AC-GDL-012.

### M3 — legibility and safety

- Lead with changed classifications; carry unchanged non-clean entries as a compact standing count (REQ-GDL-007). Requires persisting the previous result to diff against, **outside the working tree** — AC-GDL-008(b) asserts a byte-identical tree across a run.
- No forge mutation, no working-tree write (REQ-GDL-008).

Flips: AC-GDL-007, AC-GDL-008.

### M4 — doctrine (most mechanical)

- Add the continued-firing clause to `verification-completeness.md` as new text only; verify by diff that no existing line changed (REQ-GDL-009).

Flips: AC-GDL-009.

**Milestone map check.** Every criterion appears in exactly one flip list and each listed milestone can actually deliver it — the audit's T4 finding was that a union count answers "is every criterion in some list?" and is structurally blind to "can the listed milestone deliver it?". AC-GDL-001 is at M1 because the contract is consumed at the invocation site; AC-GDL-003 and AC-GDL-004 are M1 for the same reason.

## §G Anti-patterns to avoid

- **Gating the evaluator's invocation on anything.** A path filter, a changed-file test, or a subject-matter condition rebuilds §A.3's guard inside the deliverable. This is the single most consequential mutant in the SPEC.
- **Hardcoding a classification value name.** It works today and breaks the moment the state SPEC adds a value; more importantly it reintroduces the enumeration shape that failed three audits (B-3).
- **Shipping a CLI verb as the whole answer to arrival.** A verb the operator must know to run reproduces §A.4 exactly: the lead session could run the right query the instant it was handed the workflow's name, and could not know a query was owed.
- **Reprinting the full standing list every session.** How a new advisory inherits the filter an always-red neighbour has already trained (B-4).
- **Persisting the previous result inside the working tree.** Violates AC-GDL-008(b) and creates drift for the next reader.
- **Opening a second advisory surface.** Two channels for one concern is §A.8's filtering mechanism applied twice. Join the existing block (REQ-GDL-010).
- **Querying the forge inline at the host surface.** It passes every test on a fast network and stalls session start on a slow one — a latency defect that appears only where it hurts (B-6).
- **Filtering the refresh anywhere, including inside it.** An evaluator invoked on every activation that returns early on a subject-matter test satisfies a call-site reading and is the same defect one frame inward. AC-GDL-003 counts at the query layer for this reason.
- **Identifying clean entries by the surface fold.** It under-fires: more than one classification folds to the clean surface value while only one classification is clean. Read the carried designator (REQ-GDL-001 (iii)).
- **Awaiting the refresh, or discarding it when it overruns.** Awaiting blocks the host; discarding leaves the persisted result frozen while its disclosed age grows. Initiate, abandon for this turn, persist on completion (REQ-GDL-012).
- **Checking for a t326 surface in this tree.** It is absent here and present on `origin/develop`; the check returns "No such file or directory" for a landed feature (B-5).
- **Describing any instance as a hidden failure.** B-0. The wording determines what gets built.

## §H Cross-references

- `.moai/reports/t333/trigger-axis-observation.md` — the primary evidence artifact (tracked at `c30f761dd`).
- `.moai/reports/plan-audit/SPEC-GUARD-LIVENESS-001-review-{1,2,3}.md` — the three audit iterations; review-3 carries the FAIL + STOP and the scope-reduction recommendation this split implements.
- `.moai/specs/SPEC-GUARD-STATE-MODEL-001/` — the sibling SPEC holding the state model.
- `.claude/rules/moai/development/verification-completeness.md` — the landed rule this SPEC extends additively.
- `SPEC-BINARY-LAG-VISIBILITY-001` (card t326, in flight) — the binary-state axis, out of scope here.

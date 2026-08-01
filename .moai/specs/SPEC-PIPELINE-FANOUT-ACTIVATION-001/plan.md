# Plan — SPEC-PIPELINE-FANOUT-ACTIVATION-001

Milestones are ordered by decision-reversibility: the choices most likely to change on review come
first, the mechanical mirroring comes last.

---

## A. Tier assignment and its justification

**Assigned tier: M.** Applying the very budget this SPEC introduces (REQ-PFA-008): Tier M permits
≤ 16 requirements and ≤ 16 acceptance criteria. This SPEC carries **10 requirements and 15
acceptance criteria**, both within budget.

The assignment is a judgment call between M and L, and the counterargument is recorded rather than
suppressed:

| Tier criterion | Measurement | Points to |
|---|---|---|
| Scope (LOC) | roughly 150-200 lines added, no code | S / M |
| Files affected | 20 | L (> 15) |
| Constitutional? | amends a `[HARD]` clause in the SPEC-workflow rule and a binding agent contract | L |

**Why M was chosen over L.** The 20-file count is an artifact of the mandatory 1:1 template
mirroring — it is 10 logical surfaces, each edited twice with the same semantic change. Counted by
logical surface the scope is 10, inside Tier M's 5-15 band. The change adds no code, no
behaviour, and no new mechanism; the two doctrine edits (D-1 fix, tier budget) each amount to a
single paragraph, and the D-1 edit removes a contradiction rather than introducing new policy.

> **Count correction (audit iteration 1, D3).** An earlier draft of this section stated 18 files /
> 9 surfaces and omitted `run.md` from the surface enumeration. The true figure is **10 surfaces /
> 20 files** — `run.md` hosts a Fan-Out Index table (§B M1.1) even though all four run-phase sites
> live in its sub-skills, which is exactly the D-2 discoverability problem this SPEC addresses. The
> Tier M conclusion survives the correction on the corrected premise: 10 logical surfaces is still
> inside the 5-15 band, and 20 files is still an artifact of doubling, not of genuine breadth. The
> constitutional-amendment criterion remains the stronger argument for L, and is answered in the
> paragraph below rather than by the file count.

**Residual risk of this call.** An auditor may reasonably read "amends a `[HARD]` rule" as
constitutional and argue for L. If that reading prevails, the remedy is cheap: Tier L requires adding
`design.md`, and the requirement and criterion counts already fit the wider Tier L budget of 25.
`research.md` is already present above the Tier M minimum, because the audit evidence needs a
version-controlled home (the originating report is gitignored).

---

## B. Milestone M1 — the only milestone in scope

M2 (verify-snapshot activation), M3 (refutation step / D-3), and M4 (audit-stream unification) are
named in `research.md` for continuity and are explicitly excluded from this SPEC.

### M1.1 — Fan-Out Index tables in the three routers (do this first)

This is the highest-reversibility decision in the SPEC: it introduces a new documentation structure
and a new identifier namespace, and it is the item most likely to be reshaped on review.

Add one `## Fan-Out Index` table to each router. Each table lists that phase's fan-out sites with
four columns: Fan-Out ID, trigger condition, target sub-skill file, and what is parallelised.

The tables exist to close D-2: sub-skills are Read on demand, so before this change the orchestrator
could not know a fan-out existed until it had already entered the phase serially. The index puts
that knowledge in the file that is read at phase entry.

**Ordering is load-bearing (risk R-1).** M1.1 lands before M1.2. Promoting ten sites to conditional
obligations without first making them discoverable would increase the orchestrator's judgment load
with no corresponding aid.

### M1.2 — Fan-Out ID assignment and discretionary-to-conditional promotion

Each of the ten sites gains its canonical ID inline and has its discretionary clause promoted.

| Fan-Out ID | File (local) | Content anchor | Discretionary token now |
|---|---|---|---|
| FO-PLAN-1 | `.claude/skills/moai/workflows/plan.md` | `plan-research-fanout.js` … `orchestrator MAY launch it` | `orchestrator MAY` |
| FO-PLAN-2 | `.claude/skills/moai/workflows/plan/spec-assembly.md` | `Parallel Review Lenses` … harness level `standard` or `thorough` | `orchestrator MAY` |
| FO-RUN-1 | `.claude/skills/moai/workflows/run/phase-execution.md` | `Sharding (read-only, optional)` … `shard the scan` | `orchestrator MAY` |
| FO-RUN-2 | `.claude/skills/moai/workflows/run/task-decomposition.md` | `RED-stage Drafter Pool` … `spans several independent test targets` | `orchestrator MAY` |
| FO-RUN-3 | `.claude/skills/moai/workflows/run/task-decomposition.md` | `sync-audit-4dim.js` … `launch it once across the Phase 13 / 16 / 17 quality band` | `orchestrator MAY` |
| FO-RUN-4 | `.claude/skills/moai/workflows/run/task-decomposition.md` | `**Sharding (optional).**` … `the *scan* MAY be sharded` | `MAY be sharded` |
| FO-SYNC-1 | `.claude/skills/moai/workflows/sync.md` | `sync-audit-4dim.js` … `launch it at Phase 7` | `orchestrator MAY` |
| FO-SYNC-2 | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `Sharding (read-only, optional)` … `this scan MAY be sharded` | `MAY be sharded` |
| FO-SYNC-3 | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | `Per-package fan-out (read-only drafting, optional)` | `orchestrator MAY` |
| FO-SYNC-4 | `.claude/skills/moai/workflows/sync/doc-execution.md` | `Drafter / Applier Structure` … `spans several independent document families` | `orchestrator MAY` |

**Promotion shape.** The promoted clause keeps the site's existing precondition and turns the
discretionary verb into a conditional obligation, in the compound GEARS form already used across
these files:

> **Where** *(capability precondition)* **While** *(scope condition)*, the orchestrator shall
> *(run the fan-out)*.

Three properties are preserved verbatim in meaning at every site:

1. **The fail-open fallback sentence stays.** Script absent, runtime without dynamic-workflow
   support, or concurrency ceiling reached, all still route to the serial path with no error and no
   warning (REQ-PFA-004).
   **One exception requires an addition, not a preservation.** FO-SYNC-2 is the only site with no
   fallback sentence today — its block ends at "scaling, not subagent nesting" with nothing
   describing the skipped-or-unavailable path. Its promotion must therefore *add* one, matching the
   shape used by its sibling FO-SYNC-3 in the same file. This is a small scope addition beyond the
   original milestone statement; it is required because promoting a site to a conditional obligation
   while leaving it without an escape hatch is the exact hazard R-4 names. See `research.md` § E.8.
2. **No unconditional obligation.** The promoted verb stays bound to its stated condition; it must
   never read as a bare MUST detached from that condition (REQ-PFA-005).
3. **Read-only and single-writer properties are untouched.** Every fan-out agent still writes no
   file, still returns a structured blocker report rather than prompting the user, and every applier
   step still has exactly one writer.

**Do not touch.** `spec-assembly.md` carries a second `orchestrator MAY` at its tier-judgment skip
condition, which is not a fan-out and stays as it is. The codemaps extraction fan-out uses the same
phrasing and is not one of the ten sites.

### M1.3 — Resolve the D-1 contradiction

The `plan-auditor` Retry Loop Contract currently states that iteration 2 or later performs a full
audit plus a regression check. Two other surfaces — the defect-list clause in the same file, and the
run-phase execution sub-skill — both state that a confirming re-audit is scoped to the enumerated
defect delta. Rewrite the Retry Loop Contract clause to agree: an iteration-2-or-later re-audit is
scoped to the enumerated defect delta, plus the regression check over prior-iteration defects.

Everything else in that contract is preserved: the three-iteration ceiling, the automatic FAIL on an
unresolved prior defect, the iteration-3 escalation report, and stagnation detection. The verdict-
authority clause is preserved explicitly — the delta scope reduces re-audit cost and never lets an
orchestrator self-assessment substitute for an auditor verdict (REQ-PFA-007).

### M1.4 — Tier size budget

**SSOT choice: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.**

That section is the single place where the tier taxonomy is defined: it already owns the per-tier
LOC guidance, the file-count band, the artifact set, and the plan-auditor PASS threshold, and every
other surface that mentions tiers cross-references it rather than restating it. A size budget is
another per-tier dimension of exactly that kind, so it belongs in the same table's neighbourhood.

The two alternatives were rejected: `spec-frontmatter-schema.md` documents the `tier` *field*, not
the tier *semantics*, and it already defers to the workflow rule for those; `manager-spec`'s own
agent body would bind only one authoring agent rather than the taxonomy itself.

Budget statement, resolving the sum-versus-each ambiguity explicitly:

| Tier | Requirement ceiling | Acceptance-criterion ceiling |
|---|---|---|
| S | 8 | 8 |
| M | 16 | 16 |
| L | 25 | 25 |

The ceilings apply **independently** to each count, not to their sum. Motivation, per `research.md`
§ E.8: the largest observed requirement count in the catalogue is 76, and a SPEC declaring Tier M
carries 66. There is currently no ceiling at all.

### M1.5 — Template mirroring (mechanical, do this last)

Every local surface above has a mirror at `internal/template/templates/<same-path>`; all ten pairs
were confirmed present.

**[HARD] The mirrors are not byte-identical, and the differences are intentional.** Apply the *same
semantic change* to each mirror by hand. Do not copy the local file over the template.

Four known neutralization divergences must survive the edit unchanged:

| Local-only content | File | Template form |
|---|---|---|
| the pre-v3 SPEC count and its accompanying date | `.claude/skills/moai/workflows/plan.md` | genericized |
| the `Updated:` footer date line | `.claude/skills/moai/workflows/plan.md` | line absent |
| the lint-engine source path | `.claude/agents/moai/plan-auditor.md` | genericized |
| the parenthetical internal date | `.claude/agents/moai/plan-auditor.md` | date stripped |

**[HARD] All new text written into a template mirror must itself be neutral**: no SPEC identifiers,
no requirement or acceptance-criterion tokens, no internal working dates, no commit hashes, and no
internal source-code paths. The Fan-Out ID namespace is neutral by construction — the IDs name
pipeline phases, not this SPEC — so the index tables and the inline IDs are safe to mirror verbatim.

After the mirror edits, run `make build` so the embedded template filesystem is regenerated.

---

## C. File-by-file change map

Ten logical surfaces, twenty files. Every row below has a template mirror, so the file count is
exactly twice the surface count.

| # | Local file | Mirror | Changes |
|---|---|---|---|
| 1 | `.claude/skills/moai/workflows/plan.md` | yes | Fan-Out Index table; FO-PLAN-1 ID + promotion |
| 2 | `.claude/skills/moai/workflows/plan/spec-assembly.md` | yes | FO-PLAN-2 ID + promotion |
| 3 | `.claude/skills/moai/workflows/run.md` | yes | Fan-Out Index table |
| 4 | `.claude/skills/moai/workflows/run/phase-execution.md` | yes | FO-RUN-1 ID + promotion |
| 5 | `.claude/skills/moai/workflows/run/task-decomposition.md` | yes | FO-RUN-2/3/4 IDs + promotions |
| 6 | `.claude/skills/moai/workflows/sync.md` | yes | Fan-Out Index table; FO-SYNC-1 ID + promotion |
| 7 | `.claude/skills/moai/workflows/sync/quality-gates-quality.md` | yes | FO-SYNC-2/3 IDs + promotions |
| 8 | `.claude/skills/moai/workflows/sync/doc-execution.md` | yes | FO-SYNC-4 ID + promotion |
| 9 | `.claude/agents/moai/plan-auditor.md` | yes | D-1 Retry Loop Contract fix |
| 10 | `.claude/rules/moai/workflow/spec-workflow.md` | yes | Tier size budget |

Note: surface 3 (`run.md`) hosts an index table but no fan-out site of its own — all four run-phase
sites live in its sub-skills, which is precisely the discoverability problem D-2 names.

---

## D. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R-1 | Promotion raises orchestrator judgment load per phase | Apply M1.1 (index tables) strictly before M1.2 (promotion), so the obligation arrives with the aid that makes it discoverable |
| R-2 | A blind mirror copy reintroduces forbidden content and trips the CI neutrality guard | M1.5 forbids copying; apply the semantic change by hand and preserve the four listed divergences |
| R-3 | The adjacent stale hook description in local `sync.md` gets swept into this diff | Recorded as an explicit out-of-scope exclusion in `spec.md` § 4 and in `research.md` § E.7; the acceptance criteria assert it is unchanged |
| R-4 | Promotion is over-applied into an unconditional obligation, removing the fail-open guarantee | Two criteria verify this independently: **AC-PFA-008** asserts the fail-open marker survives at every site (REQ-PFA-004), and **AC-PFA-015** asserts no promoted obligation is detached from its GEARS condition (REQ-PFA-005) |
| R-5 | A criterion passes vacuously because its selector matches nothing before or after | Every marker introduced by this SPEC returns 0 at baseline (`research.md` § E.3), and every criterion records a baseline that differs from its expected post-state |
| R-6 | Tier assignment is contested at audit (M vs L) | The counterargument is recorded in § A; escalation to L costs one added artifact and no requirement or criterion rework |

---

## E. Self-verification

- All ten fan-out sites confirmed present in this worktree, with content anchors recorded.
- Three of ten sites drifted line position between the audit and this plan, which is why no
  criterion anchors on a line number.
- All ten template mirrors confirmed to exist.
- All four intentional neutralization divergences confirmed present with local=1 / template=0.
- The D-1 contradiction confirmed live on both sides of the mirror.
- Every new marker token confirmed absent at baseline.

---

## F. Cross-references

- `research.md` — the tracked evidence record, including the re-anchored inventory and every
  baseline measurement.
- `acceptance.md` — the fifteen acceptance criteria with runnable judging commands.
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the SSOT amended by M1.4.
- `.claude/agents/moai/plan-auditor.md` § Retry Loop Contract — the surface corrected by M1.3.

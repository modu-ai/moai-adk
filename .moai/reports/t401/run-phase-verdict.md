# t401 run-phase verdict (lane-5, 2026-09-07)

card: t401 · SPEC-JUDGMENT-FIRST-MODE-001 v0.2.4 · Tier L
worktree `.claude/worktrees/t401` · branch `WT-analysis-pull` · HEAD `a06999d5f` · base `ad272be20`

Scope of this record: the run-phase decisions taken after the Implementation Kickoff Approval
gate passed. The plan-phase verdict is a separate file (`plan-phase-verdict.md`) and is not
restated here.

## 1. Accepted scope overrun — carrier form (bounded; NOT a general precedent)

`manager-spec` was instructed to redefine AC-JFM-013's sweep window (line → clause/paragraph)
and nothing more. It **also** wrote down a **carrier form**: a clause this SPEC forbids editing
counts as `conditioned` when the ledger row names the coordinate the conditioning actually lives
at. That exceeded the literal instruction. It self-reported the overrun rather than folding it in
silently, and the orchestrator accepted it.

**Why it was accepted.** The operator's approval rested on a stated reason — *a Frozen line cannot
satisfy a line-local criterion even in principle*. Redefining the window alone does not close that
half: `branch-origin-protocol.md:25` and `:26` are separate `- ` list items, so they remain
separate blocks under a block window and `:25` stays permanently red. The overrun closes the
unresolved half the approval's own reason named. It is inside the approved intent, past the
approving sentence's literal wording.

**[HARD] The conditions this acceptance is bounded by.** Cited without them, this row reads as
"scope overruns are permitted", which is not what happened. All four held:

1. It closed an unresolved part **explicitly named in the operator's approval reason** — not a
   part the executor judged worth adding.
2. It moved in the **strengthening** direction: required `conditioned` coordinates went 5 → 6.
   A narrowing overrun would not have been accepted on these grounds.
3. Its **revert surface was stated and small** — one `acceptance.md` bullet, one `plan.md`
   paragraph, one ledger row.
4. It was **surfaced before adoption**, not discovered afterwards. A silent widening is a
   different act and is not covered by this row.

Discipline this composes with (lead, to lane-3): a self-change that **narrows** is reported and
proceeds; a self-change that **widens** is always escalated. This one widened and was escalated.

## 2. F3 — an unverified-coordinate finding that MUST carry its tree SHA

`manager-spec` found that of the four coordinates `spec.md` §B.1 lists as S1, two — `:293` and
`:301` at this tree — carry `(Recommended)` but not `first` / `첫` / `먼저`, so the AC-JFM-013
sweep selector never reaches them, and AC-JFM-004 is section-scoped. **Two coordinates the SPEC
names as requiring conditioning are verified by no criterion at all.** Left untouched: out of this
card's scope. Raised as a follow-up card candidate.

**[HARD] The finding is only checkable against a named tree.** The lead attempted to reproduce it
and could not:

```
primary (main)   sed -n '293p;301p' askuser-protocol.md  → empty
develop          sed -n '293p;301p' askuser-protocol.md  → empty
```

That is **failure to reproduce, not refutation** — an empty read from a shell text pipeline is not
a zero, and a verdict taken outside the measured tree is not a verdict on it. The coordinates are
post-M1 line numbers in **this** tree; the lead's trees do not carry M1's insert, and this base sits
730 commits behind `origin/develop`.

Therefore the follow-up card MUST state: **tree `a06999d5f`, worktree `.claude/worktrees/t401`,
branch `WT-analysis-pull`**, and MUST cite the coordinates by their anchor text rather than by line
number alone.

## 3. A defective instrument manufactured a category that did not exist

Row 4 (`zone-registry.md:869`) was classed `unconditioned-by-design` on an explicit-SPEC-exclusion
ground that the ledger's two-class contract does not describe. The run phase reported the third
ground rather than quietly using it; the answer was that **the ground should not have existed**.

`:869` and the already-`conditioned` `branch-origin-protocol.md:25` are the same shape — both
Frozen, both with the conditioning carried at another coordinate. They received different classes
only because the **line window** could see one carrier and not the other. Under the block window
plus the carrier form the distinction disappears and `:869` is simply `conditioned`. No
`zone-registry.md` edit is implied, so AC-JFM-012's empty-diff invariant is untouched.

Recorded because the general shape is worth keeping: **a defective measuring instrument can
manufacture a category in the classification built on top of it.** The category looked real from
inside the ledger; it was an artifact of the window.

## 4. Gaps carried into sync — not closed by this verdict

1. **The re-sweep has not been run.** 26 rows were classified under the OLD (line) window. The
   claim that the 8 already-`conditioned` rows survive the new window is a **deduction** from
   line ⊆ block monotonicity, **not a measurement**. The run-phase re-sweep is what decides it;
   if it does not return 8, that is recorded as a difference, not reconciled away.
2. **9 of the 18 `unconditioned-by-design` rows were classified from the matched line alone** —
   the enclosing paragraph was never read: `ci-watch-protocol.md:98`, `moai.md:205`, `run.md:97`,
   `harness-build-entry.md:75`, `harness-build-entry.md:120`, `harness.md:131`, `feedback.md:124`,
   `mode-orchestration.md:79`, `hns-workflow-ci-loop/SKILL.md:198`. A `[HARD]` MUST sitting in an
   unread lead-in would not have been seen — which is exactly the shape row 12 (`SKILL.md:350`)
   turned out to have. Most suspect: `harness-build-entry.md:75`, a numbered procedure step (so it
   has a lead-in by construction) whose sibling `harness.md:190` does carry an explicit
   `MUST … per askuser-protocol.md`. These 9 are folded into the re-sweep rather than read twice.
3. **AC-JFM-010 and AC-JFM-004 are reasoned PASSes, not measured ones.** Both rest on comparing
   hunk headers against separately-measured section boundaries by hand. No single command asserts
   either containment.
4. **`spec-assembly.md:212`'s paragraph boundary (`:208`–`:214`) was read by eye**, and the SPEC
   records it as "roughly".
5. **The 18 `unconditioned-by-design` reasons are judgments, not measurements**, and could be
   wrong in the same direction — toward under-conditioning. That failure is silent: an
   under-conditioned row emits no signal at all.
6. **M1 makes the tree look compliant without proving behavior.** The clauses are conditioned by
   prose reference; nothing enforces that a composer reads the mode. Enforcement arrives with M0's
   observer and M4's CI guard. This is the D4 mutant the SPEC names.
7. **No test ran.** `go build ./...` (exit 0) was the only toolchain command; `make build` was not
   run. `design.md` / `research.md` were grep-checked for coordinates, not read through.
8. **Nothing here was measured on a merged tree.** `origin/develop` is 730 ahead of this base, so
   every measurement in this verdict requires re-taking at the integration window.

## 5. Commits (this worktree, unpushed)

| SHA | What |
|---|---|
| `129fe8b88` | M0 ratification record + `calls_issued` mitigation choice (records only) |
| `56a21342c` | M1 — pull-mode convention in doctrine (5 files + mirrors) |
| `82edb9109` | M1 — row-12 scope expansion (`SKILL.md`) + row-6 strike |
| `a06999d5f` | AC-JFM-013 window unit + stale coordinates (SPEC body, 0.2.3 → 0.2.4) |

No push, no merge, no worktree disposed. Integration and re-measurement happen in the lead's
window.

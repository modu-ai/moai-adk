# SPEC-BACKLOG-HYGIENE-001 — Implementation Plan (card t332)

## §A. Tier proposal — M (proposed, not decided)

**Proposed: Tier M**, and the SPEC now fits inside it. v0.1.0 justified Tier M while carrying 23
requirements, citing "L threshold is ~25" — which reads the **Tier L** row. The real budget, from
`.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier, is:

| Tier | Requirement ceiling | AC ceiling | PASS threshold |
|---|---|---|---|
| S | 8 | 8 | 0.75 |
| **M** | **16** | **16** | **0.80** |
| L | 25 | 25 | 0.85 |

The ceilings apply **independently** to each count, never to their sum.

| Signal | Measurement | Reads as |
|---|---|---|
| Requirement count | **16** (was 23) | at the M ceiling |
| Acceptance-criterion count | **16** (was 14) | at the M ceiling |
| Milestone count | 5 | M |
| Artifact set | spec.md + plan.md + acceptance.md + progress.md | M (no `design.md`: the deliverable is a reading, not a mechanism — there is nothing to design) |
| Files written | 0 source files; N report files under `.moai/reports/t332/` | below L |
| Decision-reversibility | one genuine decision (M1's landing method); the rest is mechanical reading | below L |

**Not tiered up to L to legalize the count.** Tier L would buy a `design.md` and a `research.md`
this card has no content for, plus a stricter 0.85 threshold, purely to accommodate a requirement
count that was inflated rather than substantive: REQ-BH-003 was a special case of 001,
REQ-BH-015/016/017 were three faces of one verification-claim-integrity obligation, 020 restated 006
in the overlap context, and 021/022/023 restated the same report obligation at three layers.
Consolidation recovered the budget without losing an obligation — recorded in spec.md HISTORY as a
merge so the reduction does not read as a deletion.

Not S either: 67 cards, a six-way fan-out, and a tooling-baseline question gating every landing
verdict is more than one milestone's work.

## §B. Known Issues carried in from measurement

- **B1 — The installed binary cannot answer the landed question for this project.** v3.1.2
  (`343399d2f`, built 2026-08-27) predates `260ea5369`, and `strings ~/go/bin/moai | grep -c
  'worktree_base_branch'` returns 0. Every landing verdict depends on resolving this, which is why
  it is M1 and not a footnote.
- **B2 — A dispatch premise was already falsified in plan phase.** "All 100 rows return `no-link`"
  is false; 5 rows report `landed`, and 5 + 91 = 96 is exactly the card count (spec.md §B.4). The
  run phase inherits the corrected figure and must not re-cite the dispatch's.
- **B3 — `moai todo pr <absent-id>` prints `queue is empty`.** A misleading message about the
  filtered result set. It is not evidence about the card, and the run phase must not read it as one.
- **B4 — Queued ≠ not started.** Several queued cards hold live, locked worktrees (spec.md §B.5).
- **B5 — Worktree evidence is gitignored.** Files under `.moai/reports/t332/` written inside this
  worktree do not survive its disposal; recovery to the primary checkout is an explicit step (M5).
- **B6 — Remote-tracking refs go stale silently** (plan-audit D1). `origin/develop` and
  `origin/main` in this worktree are whatever the last fetch left behind, and a card that landed
  since then reads `not-landed` with no error — the same silent-wrong-answer shape B1 describes,
  one layer up. They also *move*: an unpinned ref means two cards read minutes apart are measured
  against different trees. M1 fetches once and pins.
- **B7 — `git status` cannot decide the no-mutation question.** The queue store is a gitignored
  state dir, and `moai todo relate` legitimately writes it (`internal/cli/todo_relate.go:66`). The
  invariant is behavioural — no card changed — and the deciding observable is the card-row digest,
  not the working tree.

## §C. Pre-flight

1. `git rev-parse --show-toplevel` → the t332 worktree; `git rev-parse --short HEAD` recorded.
2. `.moai/reports/t332/queue-snapshot.tsv` exists, 100 rows, capture time recorded.
3. Snapshot queued count re-derived; if it differs from 68, REQ-BH-003 fires before any card work.
4. The 10 picked ids and the 18 dropped ids are loaded as exclusion lists before the first read, not
   filtered afterwards.

## §D. Constraints

- Read-only `moai todo` verbs only; `moai todo relate` is the single permitted recording verb, and
  it appends a finding rather than changing a card.
- One output file per fan-out worker; no shared write path.
- No source file, no template, no queue schema is modified.

## §E. Self-Verification

Each milestone closes by citing the command it ran and that command's output into
`progress.md` §E.2. A milestone with no cited command is not closed.

## §F. Milestones

Ordered by decision-reversibility: the decision most likely to change comes first, mechanical
reading last.

### M1 — Tooling baseline: pin the refs and settle how landing is determined (**gates everything downstream**)

The one real decision in this card. Every landing verdict inherits it, so it is settled first and
recorded, not re-litigated per card.

**Step 1 — refresh and pin the refs, once, before any card is read** (closes D1/B6):

```
git fetch origin develop main
git rev-parse origin/develop origin/main
```

Record the fetch time and both SHAs in `00-tooling-baseline.md`. Every landing query downstream runs
against those **pinned SHAs**, never against the branch names, so all 67 verdicts are attributable to
one pair of named trees.

**Step 2 — choose the landing method**, by measurement rather than preference:

| Path | Step | Cost | Risk |
|---|---|---|---|
| **A — direct git** (default) | Query the two pinned SHAs per card with `--perl-regexp --grep='\b<id>\b'`, then `merge-base --is-ancestor` | 2 cheap commands per card | none installed-state-dependent; verdict is self-citing |
| **B — rebuild** | `make build`, reinstall (`rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai`), re-verify `strings … \| grep -c 'worktree_base_branch'` **≥ 1**, then use `moai todo pr` | one build; touches the installed binary | a reinstall in the middle of a live batch affects every other session on this machine |

Path A is the default precisely because B mutates shared state for a read-only card. B is taken only
if A proves insufficient, and the choice is recorded with its grounds **and its post-rebuild
`strings` count** — AC-BH-008 guards on the measured value, not on the declared path, so declaring B
without rebuilding does not buy an exemption.

**Step 3 — capture the live worktree list** (`git worktree list`) into the baseline file, so
M3's `in-flight-unlanded` classification is decided against a recorded state rather than a re-run.

Closes: REQ-BH-008..012. Output: `$R/00-tooling-baseline.md` carrying the fetch time, both pinned
SHAs, the `strings` count, the chosen path, the worktree list, and the two t342 control queries
(non-empty against pinned `develop`, empty against pinned `main`).

### M2 — Snapshot integrity, scope derivation, and the opening digest

Re-derive queued / picked / dropped counts and the in-scope id list from the snapshot alone. Record
capture time and HEAD. Record the comparison against spec.md §B.1's 68 — "no delta" is itself the
record, per REQ-BH-003.

Capture the **card-row digest** here with AC-BH-006's exact **two-step** procedure: first
`git rev-parse --path-format=absolute --git-common-dir`, whose parent directory is the primary
checkout, then `jq -S -c '[.items[] | {id, state, text}] | sort_by(.id)' <that-path>/.moai/state/todo/backlog.json | shasum -a 256`
with the resolved path passed as a literal. M5 re-captures it with the same two-step form, and each
capture records the Step 1 output it resolved.

Two reasons it is two steps rather than one, both measured in the t332 worktree: `--show-toplevel`
resolves to the *worktree* root, where no queue store exists (the queue is primary-checkout-only),
and the one-liner that computes the path inline is refused outright by the worktree-isolation guard.
A run phase that collapses this back into a one-liner will be refused; one that uses
`--show-toplevel` will read nothing.

The `.items[]` projection structurally excludes the top-level `findings` array, so M4's `relate`
invocations between the two captures leave the digest unmoved — measured, not assumed.
This is the observable that survives a `moai todo edit`, which AC-BH-005's count comparison cannot
see.

Closes: REQ-BH-001..003, and the opening half of REQ-BH-007.
Output: `$R/01-scope.md`.

### M3 — Per-card sweep (read-only fan-out)

The per-card work is independent and read-only, so it fans out. **Fan-out is explicit here**:
spawn read-only investigators over card batches, each writing only its own file.

- **Partition**: the 67 in-scope ids in snapshot order, split into 6 batches by id range —
  `B1` t90..t224, `B2` t231..t248, `B3` t252..t287, `B4` t288..t305, `B5` t313..t339,
  `B6` t343..t362. **Measured sizes: B1=10, B2=11, B3=13, B4=10, B5=13, B6=10, total 67.**
  (v0.1.0 stated "6 batches of 11-12", which two batches exceed; corrected per plan-audit D9.) The
  exact membership is re-derived from M2's list, never retyped from here.
- **Write isolation [HARD]**: worker *k* writes `$R/cards/batch-<k>.md` and nothing else. No worker
  appends to a shared file; no two workers share a path. This is what makes the fan-out safe, and it
  is the property to check if two workers ever produce one file.
- **Per card**, each worker produces: the premise restated in one sentence; a `holds` / `falsified` /
  `unverified` verdict, with a reason on `unverified`; the landing verdict from M1's method citing
  the pinned ref SHA, the commit SHA and the `--is-ancestor` result; the `in-flight-unlanded`
  classification where M1's captured worktree list shows an unmerged branch; and the five evidence
  sections **on every entry**, not only the falsified ones.
- **Stagger the spawn**: start one worker, let it produce output, then start the rest — concurrent
  identical-prefix spawns all pay the cold cache write.

Closes: REQ-BH-013, REQ-BH-014. Output: `$R/cards/batch-1..6.md`.

### M4 — Overlap detection and relation recording

Cross-compare the 67 cards after M3, when each premise is already restated in one sentence — the
comparison is over the restatements, not over the raw card text. Confirmed overlaps are recorded
with `moai todo relate … --relation … --note "<text>"`; the note names the shared artifact.

Known starting points from the snapshot's 4 existing relation rows:

- `t313 ↔ t295` (`contains`) — both queued, both in scope.
- `t318 ↔ t256` (`absorbs`) — **`t256` measured `dropped`**, not merely absent from the queued set.
  A dropped card is already decided, so this relation is **reported as a reading only and never
  carried into the disposition proposal**; no un-drop is proposed and no fold across the boundary is
  suggested. (Corrected per plan-audit D11, which noted v0.1.0's "verify before re-recording" left
  open whether a decided card could be re-related at all.)

Closes: REQ-BH-006, REQ-BH-015. Output: `$R/02-overlaps.md`.

### M5 — Consolidated report, closing digest, disposition proposal, evidence recovery

Assemble the batch files into one reading report; **re-capture the card-row digest** and compare it
byte-for-byte against M2's; append the disposition proposal list; state that no card was mutated and
cite all three observables that decide it. Then recover the evidence tree to the primary checkout
(`ExitWorktree keep` → copy → `cmp`), because B5 makes it otherwise lost.

Closes: REQ-BH-004, REQ-BH-005, the closing half of REQ-BH-007, REQ-BH-016.
Output: `$R/report.md`, `$R/invocations.log`, the closing digest in `$R/01-scope.md`.

## §G. Anti-Patterns

- **Querying `origin/develop` without fetching, or citing the branch name instead of the pinned
  SHA.** Both produce confidently-wrong landing verdicts with no error (B6).
- **Citing the installed `moai todo pr` landed column** while B1 stands. It answers about
  `origin/main` and is silent about every `develop` landing. Declaring path B without actually
  rebuilding does not change this.
- **Reading `queue is empty` as a fact about a card** (B3).
- **Matching a card to a branch by name.** The t342 → `WT-check-must-fail` misattribution came from
  exactly this.
- **Reporting `not-landed` for an unanswerable query.** `unknown` and `not-landed` are different
  facts (REQ-BH-011).
- **Treating a clean `git status` as evidence of no mutation.** The queue store is gitignored (B7).
- **Dropping a card whose premise is "obviously" dead.** The verdict is the deliverable; the drop
  is the operator's.
- **Letting a worker "just fix" a defect it found while reading.** Out of scope by spec.md §C.
- **Re-citing the dispatch's `all 100 no-link` figure.** It is measured false (B2).

## §H. Cross-References

- spec.md §B — the measured ground truth every milestone is attributed against
- acceptance.md §B — the REQ↔AC traceability map each milestone closes against
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the 16/16 budget of §A
- `.claude/rules/moai/workflow/cache-aware-execution.md` directive 2 — the M3 stagger-spawn rule
- `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution — read-only batching
- `.moai/reports/t332/plan-audit-iter1.md` — the iter-1 verdict this revision answers

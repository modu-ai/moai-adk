---
id: SPEC-BACKLOG-HYGIENE-001
title: "Backlog hygiene sweep: read the live queue, falsify each card's premise, record relations — and mutate no card"
version: "0.2.0"
status: completed
created: 2026-08-29
updated: 2026-08-30
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: ".moai/reports/t332, internal/kanban (read-only)"
lifecycle: spec-anchored
tags: "backlog, kanban, queue, hygiene, audit, read-only, evidence, landing-state, t332"
era: V3R6
tier: M
related_specs: [SPEC-TODO-LANDING-STATE-001, SPEC-KANBAN-QUEUE-PR-SYNC-001, SPEC-TODO-ANALYSIS-001]
---

# SPEC-BACKLOG-HYGIENE-001 — Backlog hygiene sweep (card t332)

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.2.0 | 2026-08-29 | Plan-audit iter-1 repair (FAIL 0.74 against the Tier M threshold 0.80, `.moai/reports/t332/plan-audit-iter1.md`). **MP-3**: `lifecycle` was `spec-first`, which is not a member of the schema enum (`spec-anchored | spec-lite | exploratory`) — corrected to `spec-anchored`. The scoped lint had reported this file clean, because `internal/spec/lint.go:765` tests the field for presence and never for membership; the repair is verified against the schema SSOT by eye, not by the linter. **D2 — requirement set consolidated 23 → 16**, the Tier M ceiling, which v0.1.0 breached while citing the Tier L row (25) as its justification. This is a **merge, not a deletion**: REQ-BH-003 folded into 001 (truncation is a special case of "work from the snapshot"); 015/016/017 merged into one verification-claim-integrity obligation; 020 folded into the no-mutation prohibition it restated in the overlap context; 021/022/023 merged into one report-shape requirement. Every obligation of v0.1.0 survives; the count was inflated, not substantive. **D1**: the landing method queried `origin/develop` and `origin/main` as branch names, never fetched and never pinned — `grep -n fetch` over all three artifacts returned nothing, so a card that landed after this tree's last fetch would read `not-landed` with no error, and two cards read minutes apart would be measured against different trees. REQ-BH-009 now requires an explicit refresh and pinned ref SHAs. **D3**: §E's byte-level write boundary contradicted the `relate` requirement — `internal/cli/todo_relate.go:66` writes `backlog.json` through the same `newTodoStore().Mutate(...)` path a card mutation uses. The invariant is restated **behaviourally** (no card dropped, edited, closed, reordered, unpicked, or picked; appending a finding is permitted), and §E now records that the state dir is gitignored so `git status` cannot adjudicate it in either direction. **D10**: §B.4's embedded-newline caveat was an unverified hedge, and the arithmetic in this SPEC's own evidence root refutes it — 5 + 91 = 96, exactly the card count, which could not coincide if newlines were inflating the figure; replaced with that arithmetic. **D7** (`Where` → `When`/`While` on the two runtime-condition requirements) and **D8** (REQ-BH-009 given `the sweep` as its acting subject) also applied. | manager-spec |
| 0.1.0 | 2026-08-29 | Initial plan-phase authoring for card **t332**, in worktree `.claude/worktrees/t332` at HEAD `15453140a` (branch `WT-backlog-hygiene`). Every figure in §B was measured in this tree in this session rather than carried from the dispatch. Two dispatch premises were re-measured and **one was falsified** — the claim that bare `moai todo pr` returns all rows `no-link` is false: 5 rows report `landed` (§B.4). The card's subject therefore reproduced inside its own plan phase, which is recorded rather than smoothed over. | manager-spec |

## §A. Problem Statement

The backlog queue accumulates cards faster than anything re-reads them. A card is written from a
measurement taken on the day it was issued; the tree then moves. Three failure shapes follow, and
all three are silent:

- **The premise rots.** The defect the card names is repaired by unrelated work, or the cause it
  names was never the cause. The card still reads as actionable.
- **Two cards describe one defect.** Dispatched in parallel they overwrite each other; dispatched
  in sequence the second one re-derives what the first already settled.
- **The work already landed.** The commit is in the integration branch's history and nobody closed
  the card, so it is dispatched a second time.

Three cards in the current batch had their stated premise falsified by measurement during the
batch, not before it:

| Card | Stated premise | What measurement showed |
|---|---|---|
| t294 | "Fixing the workflow filter closes the double-fire" | `pull_request.branches` filters the PR's **base**, so a `develop→main` PR fires under any filter. The real cause is `concurrency.group` keyed on `github.ref`, present in all 6 workflows. The card's scope claim was also wrong — it classified `lsel-leak-guard` as carrying `branches: [main]`; that workflow carries only `paths:`. |
| t299 | "The allowlist is closed" | Not closed — **the allowlist and the doctrine disagree**. Expected 12 items; measured 5 FLAGGED / 0 non-advisory. |
| t343 | "Threshold adjustment" | Not a threshold — **a new criterion is needed**. `plan-auditor.md:482` has 0 criteria that read the RED column. |

Each was caught because a lane happened to open the card and re-measure. Nothing in the system
performs that re-measurement, and nothing records the result when it happens.

This SPEC defines that sweep as a **reading** task. Its output is a report plus recorded
relations. It changes no card, and it decides no card's fate — the operator does.

## §B. Measured Ground Truth

All figures below were measured in worktree `.claude/worktrees/t332` at HEAD `15453140a`
(`git rev-parse --short HEAD`), 2026-08-29.

### §B.1 Queue composition

Baseline artifact: `.moai/reports/t332/queue-snapshot.tsv`, 100 rows, captured 2026-08-29 20:36
from `moai todo` (3 tab-separated columns: id, status, text).

`cut -f2 .moai/reports/t332/queue-snapshot.tsv | sort | uniq -c`:

```
  18 dropped
  10 picked
  68 queued
   4 (relation rows: "↳ absorbs …" / "↳ contains …")
```

The 4 relation rows are the recorded findings of a prior `moai todo analyze` pass:
`t313 ↔ t295` (`contains`) and `t318 ↔ t256` (`absorbs`), each appearing twice (once per side).

**Queued ids (68)** — `awk -F'\t' '$2=="queued"{print $1}'`:

```
t90 t125 t154 t191 t196 t201 t204 t216 t223 t224 t231 t233 t236 t237 t239 t240
t242 t243 t244 t247 t248 t252 t253 t254 t255 t260 t262 t263 t264 t280 t281 t284
t286 t287 t288 t294 t295 t296 t297 t299 t300 t302 t304 t305 t313 t315 t318 t319
t320 t323 t324 t325 t327 t329 t332 t336 t337 t339 t343 t344 t345 t347 t348 t353
t359 t360 t361 t362
```

**Picked ids (10)** — live owning lanes, out of scope:

```
t278 t333 t338 t341 t346 t350 t354 t356 t357 t358
```

**Audit scope is 67 cards**: the 68 queued minus `t332` itself.

### §B.2 The landed question is asked against the integration branch

`.moai/config/sections/git-strategy.yaml:5` sets `worktree_base_branch: develop`.

`internal/kanban/prlink_landed.go` resolves the landed ref through `LandedRefFor(projectRoot)`
(line 74), which reads that key; `DefaultLandedRef = "origin/main"` is the fallback for a project
that configures no integration branch. The file's own header states the failure mode this closes:

> The SAME silence has a second entrance … asking the question about the WRONG ref. A project that
> integrates on a branch other than the default answers "not landed" for every card that shipped,
> with no error and no empty-output warning.

`LandedRefFor` was introduced by `260ea5369` — *"feat(SPEC-TODO-LANDING-STATE-001): M1 resolve the
landed ref from the integration branch (t331)"*.

### §B.3 The installed binary predates that fix

- MCP server banner and `moai version`: **v3.1.2, commit `343399d2f`, built 2026-08-27T14:07:38Z**.
- `strings ~/go/bin/moai | grep -c 'worktree_base_branch'` → **0**.

The installed binary therefore cannot read the integration-branch key and answers the landed
question against `origin/main`.

Consequence, measured directly:

```
$ git log origin/develop --perl-regexp --grep='\bt342\b' --oneline   # 3+ commits, tip 15453140a
$ git log origin/main    --perl-regexp --grep='\bt342\b' --oneline   # (empty)
```

t342 landed on `develop` and is invisible to the installed binary's landed check.

### §B.4 A dispatch premise falsified in this plan phase

The dispatch stated that bare `moai todo pr` "returns all 100 rows as `no-link`". Re-measured:

```
$ moai todo pr | grep -c "<TAB>landed"      → 5      (t201 t204 t237 t278 t312)
$ moai todo pr | grep -c "no-link"          → 91
```

**5 + 91 = 96, which is exactly the card count** (100 snapshot rows − 4 relation rows = 68 queued +
10 picked + 18 dropped = 96). The two figures partition the cards, so 91 is a row count and every
card is accounted for. The direction of the concern holds — every `develop`-only card reads
`no-link` — but the literal "all rows `no-link`" is false, and a run phase that asserted it would be
repeating exactly the error this card exists to catch. The 5 `landed` rows are cards whose commits
are on `origin/main`.

*(v0.1.0 annotated the 91 as a line count inflated by embedded newlines in card text. That was an
unverified hedge and the arithmetic above refutes it: if newlines were inflating the count the two
figures could not sum to the card count. Corrected per plan-audit D10.)*

**Second measured gap**: `moai todo pr t999` (an id absent from the queue) prints `queue is empty`.
That message is about the *filtered* result set, not about the queue, and it is not evidence about
the card.

### §B.5 A queued card may still have live work

`git worktree list` shows worktrees for queued cards (t294, t299, t318, t336, t337, t343, t362
among them), several `locked`. **Queued** therefore does not mean **not started**: work may exist
on an unmerged branch. The sweep must classify this state rather than infer from it.

## §C. Scope

### In scope

The 67 queued cards of §B.1 (68 queued minus t332). For each: premise re-measurement, landing
determination, and overlap detection against the other 66. Output: a per-card reading report under
`.moai/reports/t332/`, a consolidated disposition **proposal**, and `moai todo relate` records for
confirmed overlaps.

### Out of scope

The exclusions below are what this SPEC will NOT build. Each is excluded on a stated ground, not
because it was overlooked.

### Out of Scope — card mutation

- Any invocation of `moai todo drop`, `edit`, `done`, `undrop`, `move`, `unpick`, or
  `next <id>`. Disposition is the operator's act; this card produces the reading it is decided from.
- Closing, merging, or re-texting a card, however clearly falsified its premise reads.

### Out of Scope — picked cards

- The 10 `picked` cards (t278 t333 t338 t341 t346 t350 t354 t356 t357 t358) have live owning lanes.
  They are not read, not re-verified, and not proposed for disposition. A lane's card is the lane's.
- Cross-checking a queued card against a picked card for overlap is permitted **as a read**; the
  proposal that follows may name only the queued side.

### Out of Scope — dropped cards

- The 18 `dropped` rows are already decided. They are not re-litigated, and an un-drop is never
  proposed.
- A relation whose counterpart is `dropped` (the existing `t318 ↔ t256` record — `t256` measured
  `dropped`) is reported as a reading only and never carried into the disposition proposal.

### Out of Scope — repairing what the sweep finds

- No defect discovered during a premise re-measurement is fixed here. A falsified premise produces
  a finding; the repair is a separate card the operator issues.
- The binary-lag condition of §B.3 is **worked around** (§D.3) or resolved by a local rebuild; the
  release that ships the fix to users is not this card's work.

### Out of Scope — tooling changes

- No change to `internal/kanban`, to the `moai todo` surface, or to any queue schema. This card
  reads the tooling; it does not amend it.

## §D. Requirements

Sixteen requirements, the Tier M ceiling. v0.1.0 carried 23; the consolidation recorded in HISTORY
merged restatements and removed no obligation.

### §D.1 Queue-snapshot integrity

**REQ-BH-001** (Ubiquitous) — The sweep shall work from a single captured queue snapshot file
under `.moai/reports/t332/`, read in full, and shall not read the queue through a live per-card
invocation or through a truncating filter (`head`, `tail`, a `grep` that discards rows).
Truncation has previously hidden the newest cards.

**REQ-BH-002** (Ubiquitous) — The sweep shall record the snapshot's capture time and the tree HEAD
it was captured at, so every later count is attributable to one queue state.

**REQ-BH-003** (event-detected) — When the snapshot's queued-row count differs from the count
recorded in §B.1, the sweep shall re-derive the in-scope set from the snapshot and record the delta,
rather than proceeding against the §B.1 list. The in-scope set is the `queued` rows minus `t332`,
and excludes every `picked` and `dropped` row.

### §D.2 The no-mutation invariant

The invariant is **behavioural, not byte-level**: what is forbidden is changing a card, not writing
to disk. `moai todo relate` writes `backlog.json` (§E) and is permitted.

**REQ-BH-004** (Ubiquitous) — The sweep shall invoke only read-only `moai todo` verbs —
`moai todo`, `moai todo list`, `moai todo why`, `moai todo pr`, and `moai todo next` with **no
argument** — plus the recording verb of REQ-BH-006.

**REQ-BH-005** (unwanted) — The sweep shall not drop, edit, close, reorder, unpick, or pick any
card, and shall not invoke `moai todo drop`, `edit`, `done`, `undrop`, `move`, `unpick`, or
`next <id>` — under any finding, including a card whose premise is conclusively falsified and
including a card confirmed to overlap another.

**REQ-BH-006** (event-detected) — When an overlap between two in-scope cards is confirmed by
measurement, the sweep shall record it with
`moai todo relate <a> <b> --relation <contains|absorbs|replaces|conflicts> --note "<text>"`, which
appends a finding and changes no card.

**REQ-BH-007** (Ubiquitous) — The sweep shall produce two independent observables of the
no-mutation invariant: a verbatim log of every `moai todo` invocation it issued, and a digest over
the queue's card rows (`id`, `state`, `text` — the projection `.items[]`, which structurally
excludes the top-level `findings` array REQ-BH-006 legitimately appends to) captured before the
sweep and re-captured after it. The deciding command and its measured negative control are
AC-BH-006's.

### §D.3 Landing determination

**REQ-BH-008** (State-driven) — While the installed binary lacks the `worktree_base_branch` string
(the §B.3 condition), the sweep shall not cite the installed `moai todo pr` landed column as
evidence for any card.

**REQ-BH-009** (Ubiquitous) — The sweep shall refresh both integration refs once, pin the resulting
SHAs, and establish every `landed` verdict against those pinned SHAs:

```
git fetch origin develop main
git rev-parse origin/develop origin/main          # pinned; recorded once
git log <pinned-sha> --perl-regexp --grep='\b<id>\b' --oneline
git merge-base --is-ancestor <commit-sha> <pinned-sha>
```

and shall cite, for each verdict, the pinned ref SHA, the commit SHA, and the `--is-ancestor` exit
code. An unrefreshed remote-tracking ref answers `not-landed` for work that landed after this
tree's last fetch, with no error; an unpinned one moves under the sweep, so two cards read at
different moments are measured against different trees.

**REQ-BH-010** (unwanted) — The sweep shall not determine landing from a branch name. Branch-name
auto-matching produced a misattribution in a prior sweep (t342 matched to `WT-check-must-fail`); a
card↔branch pair is usable only when independently confirmed by commit-message id.

**REQ-BH-011** (event-detected) — When the landing query cannot be answered (no such ref, git
unavailable, an ambiguous id), the sweep shall record `unknown` and state why, and shall not
collapse the answer into `not-landed`. An unanswerable query is not a negative answer.

**REQ-BH-012** (State-driven) — While an in-scope card has a live worktree or an unmerged branch
(the §B.5 state), the sweep shall classify it `in-flight-unlanded` and shall record the branch and
its tip SHA, rather than reporting it as not started.

### §D.4 Per-card premise verification

**REQ-BH-013** (Ubiquitous) — For each in-scope card the sweep shall restate the card's central
premise in one sentence and shall decide it as `holds`, `falsified`, or `unverified`, where an
`unverified` verdict carries the reason it could not be decided. A plausible reading shall not be
promoted to a verdict.

**REQ-BH-014** (Ubiquitous) — Every card entry shall carry the five evidence sections — Claim,
Evidence, Baseline-attribution, Gaps, Residual-risk — and additionally: a `falsified` verdict shall
carry the command that falsified the premise together with that command's verbatim output, and any
claim that a card is already resolved, already landed, or no longer reproducible shall name the
files, commands, or refs scanned to reach it. An absence claim whose scanned scope is unnamed is a
Gap, not a finding. (`.claude/rules/moai/core/verification-claim-integrity.md` §2.)

### §D.5 Overlap and duplication

**REQ-BH-015** (Ubiquitous) — The sweep shall compare each in-scope card against the other 66 and
shall report every overlap candidate with the specific shared artifact (file, mechanism, or ref)
that grounds it. A merge or absorption is proposed, never performed — the fold is the operator's.

### §D.6 Report shape and disposition proposal

**REQ-BH-016** (Ubiquitous) — The report shall contain exactly one entry per in-scope card, with an
entry count equal to the snapshot's queued count minus one (t332); shall close with a disposition
**proposal** list whose rows each name the card, the proposed disposition (`keep` / `drop` /
`fold-into <id>` / `already-landed` / `needs-operator-decision`) and the single piece of evidence it
rests on; and shall state that no card was mutated and that the list awaits the operator.

## §E. Constraints

- **Write boundary (behavioural).** The sweep writes files under
  `.moai/specs/SPEC-BACKLOG-HYGIENE-001/` and `.moai/reports/t332/`, and additionally writes the
  queue store `backlog.json` — but **only** through `moai todo relate`, which appends a finding and
  changes no card. `internal/cli/todo_relate.go:66` reaches that file through the same
  `newTodoStore().Mutate(...)` path a card mutation uses, so a byte-level "writes nothing outside
  two directories" rule would forbid the very verb REQ-BH-006 mandates. The invariant that binds is
  REQ-BH-005: no card dropped, edited, closed, reordered, unpicked, or picked.
- **`git status` cannot adjudicate this.** The queue state dir is gitignored, so the working tree
  shows nothing whichever way a worker resolved it. The deciding observables are REQ-BH-007's
  invocation log and card-row digest, not the working tree.
- No source file, no template, and no queue schema is modified.
- Read-only fan-out workers write only their own file; two workers never share an output path.
- Evidence written under `.moai/reports/t332/` inside a worktree is gitignored and is lost on
  disposal — it is recovered to the primary checkout before the worktree closes.

## §F. Open Questions

- **Q1** — Resolve the binary lag by rebuilding (`make build` + `strings` re-verification) or by
  querying git directly for every card? plan.md M1 decides this by measurement; the default is the
  direct-git path, because it needs no install and is verifiable per card. Either path is bound by
  REQ-BH-008 and by AC-BH-008, which guards on the measured value rather than on the declared path.
- **Q2** — Tier proposal is **M**, now within the 16/16 budget (plan.md §A). Operator confirms.

## §G. Cross-references

- `.claude/rules/moai/core/verification-claim-integrity.md` — the 5-section evidence format
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — the 16/16 Tier M budget
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act
- `.claude/skills/moai/workflows/todo.md` — the queue surface and its operator-only mutation rule
- `internal/kanban/prlink_landed.go` — the landed check and its three-valued answer
- `internal/cli/todo_relate.go` — the `relate` write path the §E carve-out names
- SPEC-TODO-LANDING-STATE-001 (card t331) — the integration-branch ref resolution of §B.2

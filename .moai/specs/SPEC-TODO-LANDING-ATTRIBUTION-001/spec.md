---
id: SPEC-TODO-LANDING-ATTRIBUTION-001
title: "The landed verdict: an attribution-position predicate, and a ref chain that asks the branch this repository actually integrates on"
version: "0.1.0"
status: draft
created: 2026-09-03
updated: 2026-09-03
author: manager-spec (card t472)
priority: P1
phase: "v3.2.0 target"
module: "internal/kanban, internal/cli"
lifecycle: spec-anchored
tags: "kanban, backlog-queue, landing-state, attribution, git-flow, integration-branch, false-positive"
tier: M
related_specs:
  - SPEC-TODO-LANDING-STATE-001
  - SPEC-KANBAN-QUEUE-PR-SYNC-001
  - SPEC-WORKTREE-BASEREF-001
---

# SPEC: The landed verdict — attribution position, then the right ref

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-03 | Initial plan-phase authoring (card t472), measured in worktree `.claude/worktrees/t472` at HEAD `4bcac7079` (branch `WT-landed-drift-detect`). Two axes, merged by lead verdict into one SPEC: the attribution predicate (axis F) and the ref chain plus its disclosure (axes A+B). Every figure below is carried from this lane's own committed measurements at `.moai/reports/t472/premise-recheck.md` and `.moai/reports/t472/axis-bf-measurement.md`, or re-run in this tree and cited beside the command. |

> **Provenance discipline.** Every `file:line` citation is measured at tree `4bcac7079`. Every figure
> **about** a moving ref (`origin/main`, `origin/develop`, `refs/remotes/origin/HEAD`) travels with the
> command that produced it and is re-measured rather than re-cited by a later reader
> (`verification-claim-integrity.md` §2.1, remedy R4).

---

## §A Context

### A.1 What is wrong, in one sentence per axis

**Axis F — the predicate answers a question nobody asked.** `LandedGrepArgs`
(`internal/kanban/prlink_landed.go:96-109`, tree `4bcac7079`) builds

    git log <ref> --perl-regexp --grep=\btNNN\b --oneline

and `--grep` matches the **whole commit message**. A commit that merely *mentions* a card in its
body therefore reads as that card having landed. The file's own comment already knows the hazard —
"a card's first matching commit may be another card's report commit that merely mentions it"
(`prlink_landed.go:139-140`) — but that knowledge was spent suppressing SHA output, not on the
match itself.

**Axes A+B — the verdict asks the wrong branch, and never says which.** `LandedRefFor`
(`prlink_landed.go:74-80`) reads `git_strategy.worktree_base_branch` from the root the queue hangs
from, which `todoLandedRef` (`internal/cli/todo.go:82-90`) deliberately resolves to the **primary
checkout** — "the queue and the integration branch are properties of one repository, not of
whichever worktree the command happens to run in". That primary checkout is parked on `main`, whose
copy of the key is empty, so resolution falls to the hardcoded `DefaultLandedRef = "origin/main"`
(`prlink_landed.go:41`). The fallback is taken with no run-time notice, and the verdict line
`done <id> landing=<verdict>` (`internal/cli/todo.go:508`) does not name the ref that answered.

### A.2 The measured size of axis F

Measured in this tree with a binary built from it (`go build ./cmd/moai`), not the installed build
— VCI §2.2.

| Population | Landed verdicts | True positives | False positives |
|---|---|---|---|
| Live queue against `origin/main` (today's ref), 57 rows | 2 (t237, t312) | 0 | **2** |
| Live cards against `origin/develop` (the ref axis A would select), 31 cards | 9 | 2 (t401, t440) | **7** |

Precision of the shipped predicate on the population it currently marks landed: **0/2**.

Reproduction of one false positive, re-run in this tree at `4bcac7079`:

    git log origin/main --perl-regexp --grep='\bt237\b' --oneline
      539349c5b docs(t230): t230 sync-audit evidence ... (#1649)
      32d2221fa feat(cli): back up and disclose user-modified pre-commit hooks (t230) (#1647)

Both commits are **t230's**. `32d2221fa`'s body reads "the card about to change the hook body is
t237/#1641. Re-measured: the issue is OPEN" — a statement that t237 has *not* landed, read as
evidence that it has. Commits whose subject attributes t237: zero.

Control (the comparison is not vacuous): on `origin/develop`, ids attributed in a subject = 291,
ids mentioned anywhere in a message = 379. Both operands are non-empty.

### A.3 Why matching the subject is not enough

Two of the seven `develop` false positives survive a "match the subject only" repair, because they
are **other cards' subjects mentioning this card**. Re-run in this tree:

    git log origin/develop --perl-regexp --grep='\bt216\b' --oneline
      673d3d8a0 docs(t263): die-at-exit reproduced (0/5) — remedy sequenced behind t216

    git log origin/develop --perl-regexp --grep='\bt443\b' --format='%h %s'
      0d26f8a00 chore(catalog): revert sync-auditor hash to develop value — t443 jurisdiction (t461)

`673d3d8a0` is attributed to **t263**; `0d26f8a00` is attributed to **t461**. In both, the queried
card appears in the subject and is not the card the commit belongs to. The discriminator is
therefore not *occurrence*, nor *occurrence in the subject*, but **attribution position**.

### A.4 The three attributing positions

The two true positives (t401, t440) are caught by exactly these, and nothing else in the measured
population is:

| Form | Shape | Example |
|---|---|---|
| Conventional-commit scope | `<type>(<card>):` at subject start | `docs(t440): ...` |
| Trailing parenthetical | `(<card>)` or `(card <card>)` closing the subject | `... refresh catalog moai whole-tree hash (t447)` |
| Merge subject | a merge subject naming the card | `Merge branch 'WT-...' into develop (card t263)` |

Control for the trailing form, at `4bcac7079`: `git log refs/remotes/origin/develop --format=%s`
piped to `grep -cE '\(t[0-9]+\)'` returns 963, and `grep -cE '\(card t[0-9]+\)'` returns 214 — the
convention is alive and dominant, not a handful of stragglers.

### A.5 The ref: the repository already knows the answer

    git symbolic-ref refs/remotes/origin/HEAD   → refs/remotes/origin/develop
    primary .git/HEAD                            → ref: refs/heads/main
    primary .moai/config/sections/git-strategy.yaml:7 → worktree_base_branch: ""

The git-level answer is already `develop`. The landed check ignores it and uses a compiled-in
constant, because the only level it consults — a config file in a checkout parked on `main` — is
empty there. The key's value **is** `develop` on the integration branch (`a9c61cf56`), but that
commit is not an ancestor of `origin/main`, so the value cannot take effect until a release carries
it across.

### A.6 What the four consumers of the key want

All four consumers of `git_strategy.worktree_base_branch` want the **integration** meaning; none
wants the release meaning. One value can therefore answer all four.

| Consumer | Location (`4bcac7079`) | Wanted meaning |
|---|---|---|
| Card-worktree base | `internal/cli/session_worktree.go:215` | integration |
| doctor check | `internal/cli/doctor_worktree_base.go:43` | integration |
| Landed ref | `internal/kanban/prlink_landed.go:75` | integration |
| SessionStart alignment | `internal/hook/worktree_base_branch.go:156` | integration |

The fourth **writes** `refs/remotes/origin/HEAD` to match the setting. The key is therefore not
inert, and no requirement below may be justified on the premise that changing its value is
mechanism-free.

### A.7 The ordering decision — [HARD]

**Axis F lands before axis A.** Correcting the ref first does not fix a symptom; it **grows the
false-positive population from 2 to 9** (§A.2). Nine cards would then read landed on a ref the
operator trusts, seven of them wrongly. The predicate is repaired first, and the ref chain is laid
on top of a predicate that can bear it. This decision binds the milestone order in `plan.md`.

---

## §B Requirements

Notation: GEARS. Requirement IDs are stable; milestone assignment is in `plan.md`.

### B.1 Milestone 1 — the attribution predicate (axis F)

- **REQ-TLA-001** (Ubiquitous) — The landed predicate shall decide `landed` on **attribution
  position** within a commit's subject line, and shall not decide it on the presence of the card
  token anywhere in the commit message.

- **REQ-TLA-002** (Ubiquitous) — The set of attributing positions shall be exactly the three forms
  enumerated in §A.4: conventional-commit scope, trailing parenthetical (bare and `card`-prefixed),
  and merge subject. The enumeration shall live in one named place in the implementation, so a
  fourth form is a reviewable one-place diff.

- **REQ-TLA-003** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  only occurrence of the queried card token lies in the commit body.

- **REQ-TLA-004** (unwanted) — The landed predicate shall not report `landed` for a commit whose
  subject contains the queried card token in a **non-attributing** position, that is, a commit
  attributed to a different card.

- **REQ-TLA-005** (Ubiquitous) — The argv the landed check runs shall be constructed by a single
  exported builder, so a tripwire test asserts against the implementation's own construction rather
  than a transcription of it.

- **When** the landed query cannot be evaluated — no git, no such ref, a query error — the querier
  shall answer `unknown` rather than `not-landed` (**REQ-TLA-006**, event-detected; preserves the
  three-valued contract of `SPEC-TODO-LANDING-STATE-001`).

### B.2 Milestone 2 — the ref chain and its disclosure (axes A+B)

- **REQ-TLA-007** (Ubiquitous) — The landed ref resolver shall resolve through an ordered
  three-level chain: (1) the configured `git_strategy.worktree_base_branch`, (2)
  `refs/remotes/origin/HEAD`, (3) `DefaultLandedRef`.

- **Where** level 1 yields an empty value, the resolver shall consult `refs/remotes/origin/HEAD`
  before reaching the constant (**REQ-TLA-008**, capability gate).

- **When** `refs/remotes/origin/HEAD` cannot be resolved, the resolver shall yield
  `DefaultLandedRef` and shall not fail the invocation (**REQ-TLA-009**, event-detected).

- **REQ-TLA-010** (Ubiquitous) — The `todo done` landing verdict line shall name the ref that
  answered, appended so that the existing `done <id> ` prefix every reader keys off is preserved.

- **While** the answering ref came from a level below the configured key, **when** a landing verdict
  is emitted, the command shall disclose on stderr which chain level supplied the ref
  (**REQ-TLA-011**, compound).

- **REQ-TLA-012** (unwanted) — The landed ref resolution shall not write any repository ref, and in
  particular shall not set `refs/remotes/origin/HEAD`.

---

## §C Constraints

- **C-1** — The three-valued `LandingAnswer` contract (`landed` / `not-landed` / `unknown`) is
  inherited from `SPEC-TODO-LANDING-STATE-001` and is not renegotiated here.
- **C-2** — `--require-landed` remains opt-in, and remains asymmetric: it refuses on `not-landed`
  only, and proceeds on `unknown`. The measured behaviour at `internal/cli/todo.go:555-578` is the
  baseline.
- **C-3** — The landed query performs no network I/O. Level 2 of the chain reads a local ref.
- **C-4** — A project that configures the key keeps its current behaviour exactly: level 1 still
  answers first.
- **C-5** — The queue root resolution (`resolveTodoQueueRoot`, one repository one queue) is not
  changed. The chain is added below the config read, not moved to a different tree.

---

## §D Exclusions

This section states what is **out of scope** for this SPEC and why, so no successor re-derives the
boundary.

### Out of Scope — axis C, landing evidence storage

- `archived_items` carries no landing column. Its measured schema is
  `seq, id, text, added_at, spec_id, state, position` — no landing SHA, no answering ref, no verdict
  instant. The `landing=` line `todo done` prints is stdout-only and is never persisted; all 126
  archived rows are closed with no stored evidence.
- This is recorded here as **measured context only**. The surface belongs to card **t359**, which is
  picked and carries plan-audit iteration-1 redesign items. Authoring requirements against it here
  would duplicate a live card.

### Out of Scope — axis D, the blank card-to-SPEC link

- Re-measured in this tree: `items` 57/57 rows with a blank `spec_id`; `archived_items` 126/126
  blank. The attaching path (`todo next <n> --spec <SPEC-ID>`) exists and is used zero times.
- The card excludes this axis explicitly. It is an operational gap, not a schema defect, and it is
  recorded as context only.

### Out of Scope — axis E, a new detection subcommand

- The detection surface already ships: `moai todo pr` reads the whole live queue (57 rows), emits a
  five-value outcome column, and writes nothing.
- What is missing is that column's **precision**, which REQ-TLA-001..004 supply, and an operational
  routine for running it, which is not code.
- Representative mutant, declared a non-goal: building a new query subcommand. An implementation
  that adds one has not delivered this axis; it has re-delivered `todo pr` under a second name.

### Out of Scope — release-wait as the axis-A remedy

- Waiting for `develop` to reach `main` was rejected by lead verdict: its timing is bound to the
  t204 deploy gate, and the closed-but-picked backlog recurs on every release cycle regardless.

### Out of Scope — semantic landing ("has this card's last step landed")

- The predicate answers "has a commit attributed to this card landed on the ref", not "has this
  card's sync step landed". That limit is inherited verbatim from `SPEC-TODO-LANDING-STATE-001` and
  is not narrowed here.

---

## §E Cross-references

- `SPEC-TODO-LANDING-STATE-001` — the ancestor: the ref became resolvable, and the answer became
  three-valued. This SPEC corrects the predicate that ref is asked with, and completes the
  resolution chain.
- `.moai/reports/t472/premise-recheck.md` — the card's original axis-A premise, re-checked in this
  tree and found false.
- `.moai/reports/t472/axis-bf-measurement.md` — the axis B–F measurements and the merged axis A+B
  design section.
- `internal/kanban/prlink_landed.go`, `internal/cli/todo.go`, `internal/cli/todo_pr.go` — the
  implementation surface.

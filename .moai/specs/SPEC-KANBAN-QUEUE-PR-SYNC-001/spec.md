---
id: SPEC-KANBAN-QUEUE-PR-SYNC-001
title: "Read-only card-to-PR link surface for the kanban backlog queue"
version: "0.2.1"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, todo, github, observability, read-only"
tier: M
---

## HISTORY

- 2026-08-24 — v0.1.0 — plan-phase authoring. Grounded in the t210 measurement
  record `.moai/reports/t210/measurement.md` (M1..M4).
- 2026-08-24 — v0.1.1 — M5 (`gh` latency 0.878s) and M6 (merged-title convention
  already emergent) folded in.
- 2026-08-24 — v0.2.0 — audit iteration 1 FAIL (0.60 harmonic, MP-3 fail);
  `.moai/reports/t210/verdict.md`. Repairs: D1 (`tags` retyped to string — the
  document had never parsed), D4 (AC-002 / AC-003 fixtures were refuted by the
  carrier data and are rebuilt), D2+D16+D17 (landed-commit carrier adopted as a
  second, separately-scoped question; REQ-1.5's over-generalized ban narrowed),
  D14 (the former REQ-3 lifted into `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`),
  D3, D5, D6, D7, D9, D10, D11, D12, D13, D15.
- 2026-08-24 — v0.2.1 — audit iteration 2 PASS (0.801 against the Tier M
  threshold of 0.80); `.moai/reports/t210/verdict-2.md`. Lead-verified repairs
  applied without a further audit round (the Tier M iteration ceiling is
  reached): N2 (AC-014 added — none of the four NFRs previously had a criterion,
  and NFR-1 is the sole justification for REQ-2.5), N1 (`acceptance.md` §D.2's
  self-claim restated honestly; AC-013 mapped to the `plan.md` §D Template-First
  constraint). N5-N9 are recorded as run-phase debt in `progress.md`. Requirement
  count unchanged at 16/16.

## A. Context

The kanban backlog queue (`.moai/state/kanban/backlog.json`, surfaced by
`moai todo`) records exactly two things: what the operator admitted, and what
the operator picked. It has no knowledge of work that has already been
implemented — whether it is sitting in an open pull request or has already
landed on `origin/main`.

The originating card records **two** failure shapes, and both cost real work:

1. **The open-PR divergence.** Per **M1**, five live cards — t200, t201, t202,
   t203, t205 — sat in the queue as `queued` while each carried an open PR. The
   lead reads the queue, sees five cards awaiting dispatch, and dispatches work
   already in flight. (A sixth, t88, had been picked before the measurement, so
   the live divergence count is five.)

2. **The already-landed card.** Card t199 was `queued` while its fix commit
   `d9899f437` was *already an ancestor of `origin/main`*. It was discovered only
   after a lane had started — **one full lane wasted**, the only sub-case in the
   record with a quantified cost.

The second shape is the harder one, and it is why this SPEC carries two
independent questions rather than one. t199 has **no pull request of its own**:
its fix rode into `origin/main` inside the v3.1.3 batch PR #1602, whose title
carries no card token and whose body carries 26 card tokens *not including
t199*. Extending a PR-based resolver to merged PRs therefore does not reach it —
neither the title carrier nor the body carrier finds it. Only a query against
landed commits does. §C.2 records that measurement.

**M4** establishes that no card-to-PR mapping exists anywhere in the codebase
today. `gh` is used for PR *state* (`internal/github/gh.go`,
`internal/statusline/forge.go`, `session_worktree_prmerge.go`,
`branch_protection.go`), never for card attribution. The index would be new.

## B. The design ruling — the mechanism is strictly read-only

[HARD] **The mechanism observes and reports. It never writes to
`backlog.json`.** This ruling is stated first because it constrains every
requirement that follows.

The dispatch flagged a [HARD] conflict against
`.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an
operator act, which states *"The lead is the queue's sole producer"* (line 27)
and *"Promotion is the operator's act, always"* (line 29). Three grounds settle
it:

1. **Those clauses bind queue *mutation*, and an observation mutates nothing.**
   The first governs *admission*; the second governs *the pick*. A surface that
   computes a link and prints it does neither.

   The strongest precedent is the **third [HARD] clause of that same section**
   (line 31): *"The lead may attach a finding; it may not act on one… Analysis
   changes exactly one thing on its own authority — it refuses the admission of
   a card whose normalized text is identical to one already queued or picked,
   which creates no card and **leaves the queue file byte-identical**."* That is
   the boundary this SPEC proposes to sit inside, drawn in the very file it
   proposes to amend, and its byte-identity language is where REQ-2.1's wording
   comes from. `backlog.json`'s `findings[]` array is the storage form of the
   same idea: observations held ABOUT cards without touching them
   (`.claude/skills/moai/workflows/todo.md` § What the analyser may do restates
   the clause).

2. **Auto-updating card state would make the queue a derived artifact.** The
   queue's value is that every entry has visible operator provenance — someone
   asked for this, someone picked it. A state change nobody made destroys that
   property. It also fails asymmetrically: a wrong "this card is done" is
   contradicted by no later signal, so the error is silent and permanent,
   whereas a wrong read-only label is visible on the next render.

3. **It would inherit the carrier ambiguity measured in M2.** A mislinked PR
   under write mode silently mutates the wrong card. Read-only ambiguity is a
   label the operator can disbelieve; write-mode ambiguity is corruption.

### Rejected alternative — auto-update card state on PR merge or close

Considered and rejected: a hook or poll that flips a card to `done` when its
PR merges, and to `dropped` when its PR closes unmerged. Rejected on all three
grounds above — it is precisely the queue mutation clause (1) forbids, precisely
the derived-artifact failure (2) describes, and per (3) it would act on links
that **M2** shows are not reliably precise. Named here so a later reader sees the
option was weighed rather than overlooked.

## C. Two questions, two carriers

The SPEC answers two questions that look alike and behave differently.
Conflating them is what produced the audit's D16 finding, so they are separated
here before any requirement is stated.

| | Question | Carrier | Cost |
|---|---|---|---|
| **Q1 — attribution** | "Which card does this open PR deliver?" | PR title, then PR body | one `gh pr list` (network) |
| **Q2 — landed** | "Is this card's work already in `origin/main`?" | landed commit messages | local `git log` (no network) |

### C.1 — Q1: no PR carrier is both complete and precise (M2)

Across 11 open PRs, scanning for `\bt[0-9]{1,4}\b`:

| carrier | recall | precision | verdict |
|---|---|---|---|
| PR title | 7/11 (64%) | 7/7 — every token present is the delivering card | precise, incomplete |
| PR body | 11/11 (100%) | poor — 5 PRs carry extra tokens; #1614 carries 5 tokens for 1 card | complete, noisy |
| commit messages | 10/11, and **wrong** on #1600 | worst of the three | unusable **for Q1** |

The commit-message carrier deserves its own note because it is the one
`kanban-dispatch.md` § Isolation already makes [HARD]. PR #1600 carries commit
tokens for a dozen-plus cards, and its own delivering card (t184) is not among
them: a branch that merges the release branch inherits every other card's
commits, so the noise scales with integration rather than with the card. Being
mandated does not make it a usable attribution index.

No single carrier is both complete and precise **for Q1**. That is a
measurement, not a preference, and it is what forces the confidence labelling in
REQ-1.2 through REQ-1.4.

### C.2 — Q2: the commit carrier is the only one that works, and it is sound here

The noise that ruins the commit carrier for Q1 is **harmless for Q2**, because
the two questions read the same data in opposite directions. Q1 asks the commit
set to name a card exclusively; an inherited commit breaks that. Q2 asks only
whether `origin/main` names the card at all — and a commit that rode in on
another branch is still genuinely landed, so the "noise" *is* the property being
tested.

Measured in this worktree, both directions:

```
$ git log origin/main --perl-regexp --grep='\bt199\b' --oneline
b4b8bdfbe docs: update CHANGELOG for v3.1.3
711bfdbba merge(t199): internal/web 자기-SIGTERM TOCTOU — 시그널 등록을 바인드 앞으로
d9899f437 fix(web): register signal handling before binding the listener (t199)

$ git log origin/main --perl-regexp --grep='\bt205\b' --oneline
                                        (empty — t205 is queued with an OPEN PR)
```

The query returns the landed card and stays empty for the not-landed one. It is
not noise-free — a card's first hit can be another card's report commit merely
mentioning it — which is exactly why REQ-1.10 makes the answer a boolean and
forbids naming a delivering commit.

**The regex-engine trap.** `\b` is not POSIX ERE, and git fails *silently*:

```
$ git log origin/main -E --grep='\bt199\b' --oneline
                                        (empty — indistinguishable from "not landed")
```

An `-E` regression would render every card "not landed" and pass any acceptance
criterion that only checks for a clean result. REQ-1.9 mandates `--perl-regexp`
and AC-011 carries a positive control so the failure cannot masquerade as a
clean run.

## D. Requirements

16 leaf requirements (Tier M ceiling: 16). The doctrine changes that were
formerly REQ-3 now live in `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`.

### REQ-1 — Link resolver with honest confidence

**REQ-1.1** — The resolver shall accept a card id, the open pull-request set,
and the landed-commit history, and shall return exactly one outcome record per
card carrying the outcome kind, the PR number(s) where one applies, the PR
state where one applies, and a confidence label drawn from the closed set
`exact | inferred`.

**REQ-1.2** — **While** the PR title contains the card id token, the resolver
shall return a `linked` outcome with confidence `exact`.

**REQ-1.3** — **While** no PR title contains the card id token and exactly one
open PR body contains it, the resolver shall return a `linked` outcome with
confidence `inferred`.

**REQ-1.4** — **While** no PR title contains the card id token and more than one
open PR body contains it, the resolver shall return an `ambiguous` outcome and
shall enumerate every candidate PR number.

**REQ-1.5** — The resolver shall not consult commit messages **when attributing
an open pull request to a delivering card** (question Q1).

This prohibition is scoped deliberately. **M2** measured the commit carrier
against Q1 only, where #1600's inherited tokens are ruinous; §C.2 measures the
same carrier against Q2 and finds it sound. A blanket ban would forbid the only
carrier that answers Q2 — and would have left the t199 incident uncovered. The
ban binds Q1 and does not prejudge Q2, which REQ-1.9 governs.

**REQ-1.6** — The resolver shall not collapse an `ambiguous` outcome to a single
best candidate. An ambiguous outcome is reported as ambiguous.

**REQ-1.7** — The resolver shall return exactly one of four mutually
distinguishable outcome kinds — `linked`, `ambiguous`, `landed`, `no-link` — so
that a consumer can tell "already in `origin/main`" from "nobody has started
this" from "several candidates" without inspecting any other field.

**REQ-1.8** — The resolver shall match the card id as a whole token, so that
`t20` never matches a `t200` token.

**REQ-1.9** — **While** no open pull request carries the card id token in its
title or body, the resolver shall query landed history with
`git log origin/main --perl-regexp --grep='\b<card-id>\b'`, shall return a
`landed` outcome when that query returns at least one commit, and shall return a
`no-link` outcome when it returns none. The resolver shall use `--perl-regexp`
and shall not use `-E` or any other POSIX-ERE mode, because `\b` is unsupported
there and the resulting empty output is indistinguishable from a genuine
`no-link`.

**REQ-1.10** — The resolver shall not name, return, or otherwise claim which
commit delivered a card. The `landed` outcome is a boolean fact about
`origin/main` and nothing more. (Grounds: §C.2 — a card's first matching commit
may be another card's report commit that merely mentions it, so any
"first match is the delivering commit" reading attributes wrongly.)

### REQ-2 — A read surface that persists nothing

**REQ-2.1** — [HARD] The read surface shall leave
`.moai/state/kanban/backlog.json` byte-identical across an invocation, and shall
write no field, no `findings[]` entry, and no timestamp.

**REQ-2.2** — The read surface shall compute every outcome live at invocation
time and shall not cache any result into `backlog.json` or into any other
queue-owned file, sidecar, lock, or index under `.moai/state/kanban/`.

**REQ-2.3** — **When** the `gh` executable is absent, unauthenticated, offline,
or exits non-zero, the read surface shall degrade fail-open: it shall render the
queue with an empty link column, shall exit successfully, and shall report the
degradation as a note rather than an error. The landed check (REQ-1.9) is local
git and shall continue to run when `gh` is unavailable.

**REQ-2.4** — The read surface shall not block, delay, or fail an ordinary queue
read, and `moai todo list` shall remain lock-free, network-free, and free of any
`git` or `gh` subprocess.

**REQ-2.5** — [DESIGN DECISION] The link view shall be a **separate read-only
verb**, `moai todo pr [<id>]`, and shall not be a column on `moai todo list`.

Rationale, and it is a measurement rather than a judgement: per **M5**,
`gh pr list --state open --limit 40 --json number,title,body` costs **0.878s**,
of which 0.20s is user+sys and the rest is round-trip. `todo.md` documents
`moai todo list` as lock-free and cheap, and it is rendered on every operator
glance and by the foreman loop; a default-on column would make every one of
those callers pay ~0.9s of network to serve the one caller who wanted the link.
A separate verb keeps the cheap path cheap and makes the network cost an
explicit operator act. **No `--pr` flag on `moai todo list` is in scope** —
see §H.

**REQ-2.6** — The read surface shall render the outcome kind, and for a `linked`
outcome its confidence label, in both its human form and its `--json` form; the
`--json` form shall carry the card id, the outcome kind, the PR number(s) where
applicable, the PR state where applicable, and the confidence label where
applicable, so that a consumer's cross-check can be mechanical.

## E. Non-functional constraints

- **NFR-1** — The verb's network wall-time is bounded by **one** `gh pr list`
  invocation — measured at 0.878s (**M5**) — with no per-card network call. A
  per-card query would multiply that figure by the queue length.
- **NFR-2** — The landed check (REQ-1.9) is a local `git log` against an
  existing ref. It performs no network I/O and sits outside NFR-1's budget
  entirely.
- **NFR-3** — No new third-party dependency. The existing `gh` invocation path
  (`internal/github/gh.go`) is reused.
- **NFR-4** — The resolver is a pure function of (card id, PR record set, landed
  commit set), so it is testable against fixtures with no network and no repo.

## F. Exclusions

### Out of Scope — queue mutation

- No change to any card's state, text, queue position, or `spec_id`.
- Zero writes to `.moai/state/kanban/backlog.json`, including its `findings[]`
  array, and zero writes anywhere under `.moai/state/kanban/`.
- No automatic promotion, drop, or completion of a card on any PR event.

### Out of Scope — doctrine changes

- The pre-dispatch cross-check clause, the [HARD] PR-title naming clause, the
  non-contradiction note against the branch-name rule, and their template
  mirror are **not** in this SPEC. They are
  `SPEC-KANBAN-PR-CARD-TRACEABILITY-001`, a sibling Tier S SPEC that lands
  first. This SPEC ships the tooling that clause relies on.

### Out of Scope — merged and closed pull requests as an attribution carrier

- Q1 (REQ-1.2 through REQ-1.4) queries `--state open` only. **M2** scored the
  open set; **M6** sampled 15 merged PRs for the title-convention argument but
  did not score the merged side for precision the way M2 scores the open side.
- This exclusion is narrower than it was in v0.1.1, and the narrowing is the
  point: the t199 case is now covered by Q2 (REQ-1.9), which reaches it where a
  merged-PR extension measurably does not (§A, §C.2). What remains excluded is
  *attributing a merged PR to a delivering card* — scoring that carrier is
  deferred to a follow-up SPEC.

### Out of Scope — non-GitHub forges

- No GitLab or other forge support for Q1. The `glab` path contemplated by
  `internal/statusline/forge.go` is unmeasured for card-token carriage. Q2 is
  forge-independent by construction — it reads git, not a forge API.

### Out of Scope — retroactive title relabelling

- The currently-open PRs are not retitled. The sibling SPEC's title clause binds
  pull requests opened after it lands.

### Out of Scope — latency optimization

- `gh` latency is measured (**M5**: 0.878s, essentially all round-trip), and
  that measurement is what settles REQ-2.5. What stays out of scope is making
  that query *faster* — no caching layer, no background prefetch, no persisted
  link index. The cost is paid, visibly, only by the operator who invokes the
  verb.

## G. Cross-references

- `.moai/reports/t210/measurement.md` — M1..M6, the measurement record.
- `.moai/reports/t210/verdict.md` — audit iteration 1 (FAIL), D1-D18.
- `SPEC-KANBAN-PR-CARD-TRACEABILITY-001` — the sibling Tier S SPEC carrying the
  doctrine changes this SPEC's tooling serves.
- `.claude/rules/moai/workflow/kanban-dispatch.md` — § Entry into the board is an
  operator act (three [HARD] clauses at lines 27 / 29 / 31), § Isolation,
  § Completion is read, never trusted.
- `.claude/skills/moai/workflows/todo.md` — § What the analyser may do,
  § Reading the records (`findings[]`).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the
  read-don't-trust invariant the sibling SPEC's pre-dispatch clause
  operationalizes.

## H. Possible follow-ups (non-normative — not requirements)

Recorded so a later reader sees these were considered. Nothing here is in scope,
and nothing here has an acceptance criterion.

- **A `--pr` flag on `moai todo list`.** Deliberately not specified: it would put
  a network path behind the queue's cheapest read, and it would turn AC-009's
  zero-subprocess assertion into a no-flag-only claim. If it is ever wanted, it
  needs its own requirement plus a companion acceptance criterion asserting that
  the flag — and only the flag — gates the subprocess.
- **Scoring the merged-PR attribution carrier**, which would close the remaining
  half of the merged-PR exclusion above.
- **Forge-independent Q1** via the `glab` path, once a carrier measurement
  exists.

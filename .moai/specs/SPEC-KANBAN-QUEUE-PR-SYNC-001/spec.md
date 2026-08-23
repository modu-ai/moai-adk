---
id: SPEC-KANBAN-QUEUE-PR-SYNC-001
title: Read-only card-to-PR link surface for the kanban backlog queue
version: 0.1.0
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: kanban
lifecycle: spec-anchored
tags: [kanban, todo, github, observability, read-only]
tier: M
---

## HISTORY

- 2026-08-24 — v0.1.0 — plan-phase authoring. Grounded in the t210 measurement
  record `.moai/reports/t210/measurement.md` (M1..M4). `status: draft`.

## A. Context

The kanban backlog queue (`.moai/state/kanban/backlog.json`, surfaced by
`moai todo`) records exactly two things: what the operator admitted, and what
the operator picked. It has no knowledge of work that has already been
implemented and opened as a pull request.

The consequence is measured, not assumed. Per **M1**, five live cards — t200,
t201, t202, t203, t205 — sat in the queue as `queued` while each carried an
open PR. The lead reads the queue, sees five cards awaiting dispatch, and
dispatches work that is already in flight. A sixth card (t88) had been picked
before the measurement ran, so the live divergence count is five.

The queue is not wrong; it is *uninformed*. Nothing in it is false. What is
missing is a second fact — the PR state — that lives entirely outside the file.

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
operator act, which states *"The lead is the queue's sole producer"* and
*"Promotion is the operator's act, always"*. Three grounds settle it:

1. **Those clauses bind queue *mutation*, and an observation mutates nothing.**
   A read surface that computes a link and prints it engages neither clause —
   it produces no card, promotes no card, and leaves the file byte-identical.
   Per **M3**, this subsystem has already drawn exactly this line once:
   `.claude/skills/moai/workflows/todo.md` § What the analyser may do carries a
   [HARD] clause that analysis *"never folds one card into another, never
   reorders the queue, never drops a card, and never edits one… Acting on a
   record is the operator's act"*, and `backlog.json` already carries a
   `findings[]` array holding observations ABOUT cards without touching them.
   The precedent is not an analogy; it is the same file and the same boundary.

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
PR merges, and to `dropped` when its PR closes unmerged. It was rejected on all
three grounds above — it is precisely the queue mutation clause (1) forbids, it
is precisely the derived-artifact failure (2) describes, and per (3) it would
act on links that **M2** shows are not reliably precise. It is named here so a
later reader sees the option was weighed rather than overlooked.

## C. The carrier problem (M2)

Across 11 open PRs, scanning for `\bt[0-9]{1,4}\b`:

| carrier | recall | precision | verdict |
|---|---|---|---|
| PR title | 7/11 (64%) | 7/7 — every token present is the delivering card | precise, incomplete |
| PR body | 11/11 (100%) | poor — 5 PRs carry extra tokens; #1614 carries 5 tokens for 1 card | complete, noisy |
| commit messages | 10/11, and **wrong** on #1600 | worst of the three | unusable |

The commit-message carrier deserves its own note because it is the one
`kanban-dispatch.md` § Isolation already makes [HARD]. PR #1600 carries 15
commit tokens, and its own delivering card (t184) is not among them: a branch
that merges the release branch inherits every other card's commits, so the
noise scales with integration rather than with the card. Being mandated does
not make it a usable index.

No single carrier is both complete and precise. That is a measurement, not a
preference, and it is what forces both the confidence labelling in REQ-1 and
the naming convention in REQ-3.

## D. Requirements

### REQ-1 — Link resolver with honest confidence

**REQ-1.1** — The resolver shall accept a card id and the open pull-request set
and return at most one link record per card, carrying the PR number, the PR
state, and a confidence label drawn from the closed set
`exact | inferred | ambiguous`.

**REQ-1.2** — **Where** the PR title contains the card id token, the resolver
shall label the link `exact`.

**REQ-1.3** — **Where** no PR title contains the card id token **While** exactly
one open PR body contains it, the resolver shall label the link `inferred`.

**REQ-1.4** — **Where** no PR title contains the card id token **While** more
than one open PR body contains it, the resolver shall label the result
`ambiguous` and shall enumerate every candidate PR number.

**REQ-1.5** — The resolver shall not consult commit messages as a linking
carrier. (Grounds: **M2** — the carrier is wrong on #1600, where the delivering
card is absent from 15 inherited tokens.)

**REQ-1.6** — The resolver shall not resolve an `ambiguous` result to a
single best candidate. An ambiguous link is reported as ambiguous.

**REQ-1.7** — **When** the card id token appears in no title and no body of any
open PR, the resolver shall return no link for that card, distinguishable by
the consumer from an ambiguous result.

**REQ-1.8** — The token match shall be a whole-token match on the card id, so
that `t20` never matches a `t200` token.

### REQ-2 — A read surface that persists nothing

**REQ-2.1** — [HARD] The read surface shall leave
`.moai/state/kanban/backlog.json` byte-identical across an invocation. It writes
no field, no `findings[]` entry, and no timestamp.

**REQ-2.2** — The read surface shall compute the link set live at invocation
time and shall not cache the result into any queue-owned file.

**REQ-2.3** — **When** the `gh` executable is absent, unauthenticated, offline,
or exits non-zero, the read surface shall degrade fail-open: it renders the
queue with the link column blank or absent, exits successfully, and reports the
degradation as a note rather than an error.

**REQ-2.4** — The read surface shall not block, delay, or fail an ordinary queue
read. `moai todo list` remains lock-free and network-free unless the operator
opts in.

**REQ-2.5** — [DESIGN DECISION] The link view shall be a **separate read-only
verb** (`moai todo pr [<id>]`), not an always-on column on `moai todo list`.

Rationale, and it is a measurement rather than a judgement: per **M5**,
`gh pr list --state open --limit 40 --json number,title,body` costs **0.878s**,
of which 0.20s is user+sys and the rest is round-trip. `todo.md` documents
`moai todo list` as lock-free and cheap, and it is rendered on every operator
glance and by the foreman loop; a default-on column would make every one of
those callers pay ~0.9s of network to serve the one caller who wanted the link.
A separate verb keeps the cheap path cheap and makes the network cost an
explicit operator act. An opt-in `moai todo list --pr` flag MAY additionally
render the column, and if provided it inherits REQ-2.1 through REQ-2.4
unchanged.

**REQ-2.6** — The read surface shall render the confidence label alongside every
link, so an `inferred` or `ambiguous` link is never presented as an `exact` one.

**REQ-2.7** — The read surface shall emit a structured form (`--json`) carrying
card id, PR number(s), PR state, and confidence label, so the lead's
pre-dispatch cross-check (REQ-3.1) can be mechanical.

### REQ-3 — The doctrine change

**REQ-3.1** — [HARD] `.claude/rules/moai/workflow/kanban-dispatch.md` shall
require the lead, **before** dispatching a card out of `backlog`, to read that
card's PR state and report it in the same turn. **When** the card carries an
open PR, the lead shall surface that fact to the operator rather than
dispatching. (This is the step whose absence produced the M1 incident.)

**REQ-3.2** — [HARD] `kanban-dispatch.md` shall require every pull request
delivering a card to carry that card's id in its **PR title**. This converts the
M2 title carrier from 64% recall to 100%, which turns REQ-1's resolver from a
heuristic into an exact lookup — the fix for ambiguous parsing is a naming
convention, not a smarter parser.

**REQ-3.3** — The doctrine shall state explicitly that REQ-3.2 does not
contradict the existing [HARD] rule that a card worktree's **branch** name must
exclude the card id. The branch carries a descriptive slug for human
readability; `kanban-dispatch.md` already assigns traceability to three other
carriers (the dispatch `card:` field, the commit message, the evidence path).
This SPEC adds the PR title as the fourth, and the only machine-readable one.

**REQ-3.4** — The doctrine change shall be mirrored into
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md`
per the Template-First rule, subject to the template neutrality catalogue.

## E. Non-functional constraints

- **NFR-1** — The separate verb's wall-time is bounded by one `gh pr list`
  invocation; no per-card network call.
- **NFR-2** — No new third-party dependency. The existing `gh` invocation path
  (`internal/github/gh.go`) is reused.
- **NFR-3** — The resolver is a pure function of (card id set, PR record set),
  so it is testable against a fixture PR set with no network.

## F. Exclusions

### Out of Scope — queue mutation

- No change to any card's state, text, queue position, or `spec_id`.
- Zero writes to `.moai/state/kanban/backlog.json`, including its `findings[]`
  array.
- No automatic promotion, drop, or completion of a card on any PR event.

### Out of Scope — merged and closed pull requests

- Only `--state open` PRs are linked. **M2** measured open PRs only; the
  measurement record's gaps section states that a card whose PR has already
  merged (the t199 case) is a different query whose carrier statistics are
  unmeasured. Deferred to a follow-up SPEC that first measures the closed set.

### Out of Scope — non-GitHub forges

- No GitLab or other forge support. The `glab` path contemplated by
  `internal/statusline/forge.go` is unmeasured for card-token carriage.

### Out of Scope — retroactive title relabelling

- The 11 currently-open PRs are not retitled. REQ-3.2 binds pull requests opened
  after the doctrine change lands.

### Out of Scope — latency budgeting

- `gh` latency was not measured (measurement record, gaps). REQ-2.5's
  separate-verb choice removes the need for a latency budget on the hot path;
  no budget is specified here.

## G. Cross-references

- `.moai/reports/t210/measurement.md` — M1..M4, the measurement record this SPEC
  cites throughout.
- `.claude/rules/moai/workflow/kanban-dispatch.md` — § Entry into the board is an
  operator act, § Isolation, § Completion is read, never trusted.
- `.claude/skills/moai/workflows/todo.md` — § What the analyser may do (the
  observe-never-mutate precedent), § Reading the records (`findings[]`).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the read-don't-trust
  invariant REQ-3.1 operationalizes at the dispatch boundary.

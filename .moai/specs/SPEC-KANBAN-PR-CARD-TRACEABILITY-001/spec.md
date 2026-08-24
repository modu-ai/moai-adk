---
id: SPEC-KANBAN-PR-CARD-TRACEABILITY-001
title: "Pre-dispatch PR cross-check and the PR-title card-id convention"
version: "0.1.1"
status: in-progress
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "kanban, dispatch, doctrine, traceability, template-mirror"
tier: S
---

## HISTORY

- 2026-08-24 — v0.1.0 — plan-phase authoring. Split out of
  `SPEC-KANBAN-QUEUE-PR-SYNC-001` per audit finding D14
  (`.moai/reports/t210/verdict.md`): that SPEC carried 19 leaf requirements
  against a Tier M ceiling of 16, and its four doctrine requirements are
  code-free and land on a different schedule from the tooling.
- 2026-08-24 — v0.1.1 — audit iteration 2 PASS (0.775 against the Tier S
  threshold of 0.75); `.moai/reports/t210/verdict-2.md`. Lead-verified repairs:
  N3 (REQ-005 restored — the template-mirror obligation did not survive the
  split into the requirement layer, leaving AC-004 orphaned) and N4 (a judgement
  removed from AC-003's `**Mechanical.**` block, where it was a relapse of the
  very defect the split-AC pattern exists to prevent). 5 requirements against a
  Tier S ceiling of 8.

## A. Context

A kanban lead dispatches a card out of `backlog` without any check on whether
the work already exists. Two measured incidents follow from that, both recorded
in `.moai/reports/t210/measurement.md` and re-verified during the t210 audit:

1. Five cards (t200, t201, t202, t203, t205) sat `queued` while each carried an
   open PR (**M1**).
2. Card t199 sat `queued` while its fix commit was already an ancestor of
   `origin/main` — discovered only after a lane had started, costing one full
   lane.

Neither is a tooling gap alone. The lead has no obligation to look, so a lead
with perfect tooling available would still dispatch blind. That obligation is
what this SPEC adds.

This SPEC is **doctrine only — no code**. The tooling it relies on is
`SPEC-KANBAN-QUEUE-PR-SYNC-001` (`moai todo pr`, the resolver, the landed
check). The two are sequenced deliberately, and this one lands first: the
PR-title convention in REQ-2 is what raises the title carrier's recall, and the
resolver is materially more useful once titles are reliable.

**Doctrine before tooling is a real, accepted cost.** Between this SPEC landing
and the sibling shipping, the [HARD] cross-check clause is live and the operator
satisfies it by hand (`gh pr list`, `git log`). That interval is why the split
is honest rather than cosmetic: the previous single-SPEC plan left the same
interval implicit inside one milestone sequence.

## B. Requirements

5 leaf requirements (Tier S ceiling: 8).

**REQ-001** — **When** the lead is about to dispatch a card out of `backlog`,
`.claude/rules/moai/workflow/kanban-dispatch.md` shall require the lead to read
that card's pull-request and landed state first and to report what it read in
the same turn.

**REQ-002** — **While** a card carries an open pull request or is already landed
on `origin/main`, the lead shall report that fact to the operator and shall not
dispatch the card until the operator, so informed, confirms or withdraws it.

The wording is deliberate and the distinction is load-bearing. Under
`kanban-dispatch.md` line 29, *"Promotion is the operator's act, always"* — the
operator has **already picked** this card. A lead that then withholds dispatch
on its own authority has overridden an operator act, which is precisely the
de-facto-authority hazard the sibling SPEC's read-only ruling exists to prevent.
So the clause is **report-and-re-decide, never a lead-side veto**: the lead
surfaces what it read and hands the decision back. No mechanical check can tell
the two readings apart after the fact, so the wording is the only control.

**REQ-003** — Every pull request delivering a card shall carry that card's id in
its **title**.

The clause binds only card-delivering pull requests. A release PR, a batch PR,
or a `chore(release-update)` PR delivers no card and carries no obligation.

**REQ-004** — The doctrine shall state explicitly that REQ-003 does not
contradict the existing [HARD] rule that a card worktree's **branch** name must
exclude the card id, and shall name the four traceability carriers so a reader
sees the division rather than a conflict.

**REQ-005** — Every doctrine change this SPEC makes to
`.claude/rules/moai/workflow/kanban-dispatch.md` shall be mirrored into
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` per
the Template-First rule, subject to the template neutrality catalogue.

Restored per audit N3. This requirement existed in the pre-split SPEC (as
REQ-3.4) and did not survive the split into the requirement layer, which left
AC-004 exercising no requirement. Tier S is at 5 of a ceiling of 8, so unlike
the sibling SPEC's N1 there is no budget obstacle to stating it normatively.

## C. Why these two clauses, and why they are cheap

**REQ-001/002 close the gap that cost the lane.** The measured failure was not
that the lead looked and misread; it was that nothing required it to look.

**REQ-003 is codification, not imposition.** Per **M6**, 8 of 15 merged PR titles
already carry a card token, and most of the 7 that do not deliver no card at all
(`release: v3.1.3`, the v3.1.3 batch PR, three `chore(release-update)` entries).
Among merged PRs delivering a single card, the title token is close to
universal. The rule names a convention contributors already mostly follow.

**REQ-003 is also what makes the tooling exact.** Per **M2**, the title carrier
has precision 7/7 but recall 7/11 (64%); the body carrier has recall 11/11 but
carries up to 5 tokens for 1 card. Making the title mandatory takes the precise
carrier to full recall, which converts the sibling SPEC's resolver from a
heuristic into a lookup. **The fix for ambiguous parsing is a naming convention,
not a smarter parser.**

**REQ-004 forestalls a reading that looks like a contradiction.**
`kanban-dispatch.md` already forbids the card id in the branch name — the branch
carries a descriptive slug (`WT-branch-naming`, not `WT-t0`) so a reader of
`git branch` learns what changed. That rule assigns traceability to three
carriers: the dispatch `card:` field, the commit message, and the evidence path.
REQ-003 adds a fourth — the PR title — and it is the only one a resolver can
read off the pull-request surface itself. The dispatch `card:` field is
machine-readable too, but it lives in a session message rather than on the
pull request, so nothing reading a pull request can reach it.
Branch names are for humans scanning a list; PR titles are for a resolver. There
is no conflict, but a reader meeting both [HARD] clauses cold will suspect one,
so the doctrine says so outright.

## D. Acceptance criteria (inline — Tier S)

4 criteria (Tier S ceiling: 8). Each grep-shaped criterion records the
zero-hit pre-implementation baseline that makes it non-vacuous, and each is
split into a mechanical half and a reviewer-judgement half so the SPEC does not
claim mechanical coverage of a human judgement (audit D12).

### AC-001 — the pre-dispatch cross-check clause exists and is marked [HARD]

**Pre-implementation baseline (required, confirmed independently during the
t210 audit):**

```
$ grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

**Mechanical.** **Given** the doctrine file after M1 **When** the following run
**Then** both return at least 1:

```
grep -c 'pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
grep -c '\[HARD\].*pre-dispatch PR cross-check' .claude/rules/moai/workflow/kanban-dispatch.md
```

The second grep is what makes the `[HARD]` marker mechanically observable rather
than asserted.

**Reviewer judgement (recorded as such).** A reviewer confirms the clause
actually requires the lead to *read and report* the card's PR and landed state
before dispatching. No command verifies this, and this SPEC does not claim one
does.

### AC-002 — the clause is report-and-re-decide, not a lead-side veto

**Mechanical.** The clause contains the phrase `confirms or withdraws`:

```
grep -c 'confirms or withdraws' .claude/rules/moai/workflow/kanban-dispatch.md
```

Baseline 0 before M1; at least 1 after.

**Reviewer judgement (recorded as such).** A reviewer confirms the clause reads
as report-and-re-decide and cannot be read as authorizing the lead to refuse a
picked card on its own authority (REQ-002).

This criterion exists because the ambiguity it guards against is invisible to
the AC-001 greps: a clause satisfying both of those could still be written as a
veto.

### AC-003 — the PR-title clause and its non-contradiction note exist

**Pre-implementation baseline (required, confirmed independently during the
t210 audit):**

```
$ grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
0
```

**Mechanical.** After M1 that grep returns at least 1:

```
grep -c 'PR title MUST carry the delivering card id' .claude/rules/moai/workflow/kanban-dispatch.md
```

> Per audit N4, the clause *"and the same section names all four traceability
> carriers"* was removed from this Mechanical block. The shown grep covers only
> the first clause, so the four-carrier claim was a judgement wearing a
> mechanical label — a relapse of the exact defect this SPEC's split-AC pattern
> exists to prevent. Nothing is lost: the claim is stated correctly in the
> judgement half below.

**Reviewer judgement (recorded as such).** A reviewer confirms the section
explicitly states the non-contradiction against the branch-name exclusion rule,
names the branch slug and the three existing carriers, and scopes the obligation
to card-delivering pull requests only (REQ-003, REQ-004).

### AC-004 — template mirror parity (REQ-005)

**Given** the doctrine edits in `.claude/rules/moai/workflow/kanban-dispatch.md`
**When** the mirror at
`internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` is
compared
**Then** every clause added by AC-001, AC-002, and AC-003 is present in the
mirror,
**And** `MOAI_TEMPLATE_LEAK_STRICT=1 go test ./internal/template/... -v` passes,
**And** `make build` has been run.

Both mirror targets were confirmed present during the t210 audit.

### Traceability

| REQ | AC |
|---|---|
| REQ-001 | AC-001 |
| REQ-002 | AC-002 |
| REQ-003 | AC-003 |
| REQ-004 | AC-003 (judgement half) |
| REQ-005 | AC-004 |

Every requirement has a criterion and every criterion exercises a requirement.
AC-004's mapping to REQ-005 is what audit N3 restored — before the restoration
it exercised nothing.

## E. Exclusions

### Out of Scope — all tooling

- No Go code, no CLI verb, no resolver, no `gh` or `git` invocation ships in
  this SPEC. Every mechanical capability the doctrine references belongs to
  `SPEC-KANBAN-QUEUE-PR-SYNC-001`.
- The clause is satisfiable by hand in the interval before that SPEC lands
  (§A), which is what makes the ordering safe.

### Out of Scope — queue mutation

- Nothing here writes to `.moai/state/kanban/backlog.json`. The lead's
  obligation is to read and report; the operator still decides. The three [HARD]
  clauses of `kanban-dispatch.md` § Entry into the board is an operator act are
  unchanged.

### Out of Scope — retroactive relabelling

- Currently-open and already-merged pull requests are not retitled. REQ-003
  binds pull requests opened after this SPEC lands.

### Out of Scope — enforcement automation

- No CI check, hook, or lint rule enforces the PR-title convention. The
  obligation is doctrinal. Mechanical enforcement — a PR-title check in CI — is
  a plausible follow-up once the convention has settled, and is not specified
  here.

## F. Cross-references

- `SPEC-KANBAN-QUEUE-PR-SYNC-001` — the sibling Tier M SPEC carrying the
  tooling; lands second.
- `.moai/reports/t210/measurement.md` — M1 (the five divergent cards), M2 (the
  carrier scorecard), M6 (the title convention already emergent).
- `.moai/reports/t210/verdict.md` — audit iteration 1; D14 (the split), D3 (the
  REQ-002 wording), D12 (the mechanical/judgement AC split), D15 (mirror
  parity).
- `.claude/rules/moai/workflow/kanban-dispatch.md` — the file being amended:
  § Entry into the board is an operator act (three [HARD] clauses at lines 27 /
  29 / 31), § Isolation (the branch-naming rule and the three carriers),
  § Completion is read, never trusted.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the
  read-don't-trust invariant REQ-001 operationalizes at the dispatch boundary.

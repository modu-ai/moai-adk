---
description: "Card-lifecycle companion for factory-dispatch.md — report milestones ↔ queue cards, card classes, the /clear handoff message structure, isolation rationale, the pre-dispatch cross-check rationale, and the PR-title carrier measurements"
paths: "**/factory-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.claude/skills/moai/workflows/gtd.md"
---

# Factory Dispatch — Card Lifecycle and Traceability

> Companion of `factory-dispatch.md` (the always-loaded stub), split from `factory-dispatch-detail.md` by its per-file budget. The stub keeps every [HARD] rule; this file owns the card-lifecycle rationale: why reports carry Card Cross-Check tables, how cards are classified, the `/clear` message structure, the isolation rationale, the pre-dispatch cross-check wording, and the PR-title carrier measurements. Load when classifying a card, issuing or admitting one, or tracing a pull request back to its card. Siblings: `factory-dispatch-detail.md` (dispatch cycle and coordination) and `factory-dispatch-gates.md` (sync-gate review lenses and the integration-gate measurements). A bare `§` section reference that does not resolve in this file resolves in `factory-dispatch-detail.md`.

## Report milestones ↔ queue cards

The stub's [HARD] rule — a `## Card Cross-Check` section per milestone-bearing report, mapping claims verified against the queue — exists because report→card linkage used to live in one person's memory: a report declared milestones, cards were issued separately, and nothing reconciled the two — milestones surfaced with no card, cards claimed milestones already landed.

The mechanical check runs the same comparison the leader states by hand:

```
moai graph build && moai graph query --milestones-no-card
```

It writes the report-milestone and milestone-card edges from each report's Card Cross-Check table and lists every milestone whose claimed card is missing from the live queue (queued/picked; dropped does not qualify). "Not in live queue" covers both completed and never-issued cards — resolve each flag with `git log --oneline --grep 'merge: <card-id>'` before issuing a new card.

## Card classes — not every card needs every stage

Most of what accumulates in the backlog is chores: a one-line fix, a stale reference, a renamed flag. Sending those through `plan → run → sync` costs more in ceremony than the change is worth. The leader classifies each card as it leaves `backlog` and names the entry stage in the dispatch.

| Class | Shape | Path |
|---|---|---|
| A — direct close | The change is one file and one line, there is no design judgement in it, and CI catches the regression | One lane carries the card through to a pull request; `plan` is skipped |
| B — defect, cause unknown | Something is wrong and the cause has not been established | `run → sync`; `plan` is skipped, so no SPEC exists |
| C — design change | The change contains a decision, or spans subsystems | All three stages |

The Class-A evidence rule is the same shape as the CodeRabbit section of the stub: a class that skips review on a claim nobody checked is exactly the unobserved-claim hazard this rule forbids everywhere else; writing the justification down is not the same as verifying it.

The parallelism comes from the across-cards axis: each lane carries a whole card, so three lanes put three cards in flight (§ Sub-agent execution within a lane for the within-card axis; § Factory in-lane 3-stage for the stages). Research fan-out during `plan` is Class-C-only.

## The `/clear` handoff between cards — message structure

The leader's message to the operator states three things, in this order:

1. **What closed** — the card and the evidence that was read.
2. **Which session to `/clear`** — by name, so the operator clears the right terminal.
3. **What happens next** — the next card the lane is routed, and which lane will be instructed once the clear is done.

## Isolation rationale

Two properties make the shared checkout the wrong place for a card:

- Several sessions read it at once, so a branch switch, a `git stash`, or a `git add -A` there sweeps another session's uncommitted work into a commit never meant to carry it.
- A card outlives a stage — its worktree spans plan through sync, which is why disposal triggers on the merge rather than the stage finishing.

## The pre-dispatch cross-check

The stub's two [HARD] clauses are the rule; this section is why each is worded the way it is.

**The failure was not a misread — it was that nothing required looking.** Several cards sat `queued` while each already carried an open pull request, and one sat `queued` while its fix was already an ancestor of the integration branch. The second was discovered only after a lane had started work on it, which cost the whole lane — the only sub-case in the record with a quantified cost. A leader with perfect tooling available would still have dispatched blind, because reading was optional.

**Why "reports, never vetoes" is a separate clause with its own criterion.** The obligation to look and the prohibition on acting are different rules, and the second is the fragile one. The operator has already picked the card; a leader that then withholds the dispatch because it found an open pull request has overridden an operator act rather than informing one — the same de-facto-authority hazard the read-only ruling on the queue tooling exists to prevent.

The hazard is invisible after the fact. A clause requiring the leader to read and report, and a clause authorizing the leader to refuse, produce identical transcripts up to the moment the card does not move; nothing downstream distinguishes "the operator withdrew it" from "the leader declined to send it". No mechanical check can separate the two readings, so the wording is the only control there is — which is why the literal `confirms or withdraws` is pinned by its own acceptance criterion rather than folded into the obligation clause.

**The tooling is a convenience, not the obligation.** `moai gtd pr <id>` answers both halves in one read, but the clause is satisfiable by hand (`gh pr list`, then `git log` against the integration branch) and was written to be: the doctrine landed before the tooling, and the interval between them is a real operating condition rather than a paper one.

**Why the read can be a skip input for one card and report-only for another.** Put plainly, a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card: the operator already chose the second, and passing it over would override that act; the session chose the first itself, so skipping it overrides nobody. The two clauses govern different cards, so "reports, never vetoes" is untouched.

## The PR-title carrier

**Why the id leaves the branch name and lands on the PR title.** Neither name can serve both readers. The branch name is read by a human scanning `git branch` or a pull-request list, who learns nothing from an opaque card id and everything from a descriptive slug. The PR title is read by a resolver mapping pull requests back to cards, which needs a token it can match exactly. The branch-name rule and the title rule therefore assign different jobs to different names, and a reader meeting both [HARD] clauses cold will suspect a contradiction where there is none — which is why the stub states the non-contradiction outright instead of leaving it to be inferred.

**The carrier measurements.** Scanning a set of open pull requests for card tokens, three carriers behave differently:

| carrier | recall | precision | verdict |
|---|---|---|---|
| PR title | ~64% | every token present named the delivering card | precise, incomplete |
| PR body | complete | poor — one pull request carried five tokens for one card | complete, noisy |
| commit messages | high, and **wrong** on the worst case | worst of the three | unusable for attribution |

The commit carrier deserves its own note, because it is the one the isolation rule already makes [HARD]. A branch that merges the integration branch inherits every other card's commits, so its token set scales with integration rather than with the card: in the measured worst case a pull request carried a dozen-plus card tokens and its own delivering card was **not among them**. Being mandated as a traceability carrier does not make it a usable attribution index.

**The fix for ambiguous parsing is a naming convention, not a smarter parser.** The title carrier is already the precise one — where a token is present, it names the delivering card. It is incomplete only because nothing required it. Requiring it is what turns a resolver from a heuristic into a lookup, and no amount of parser sophistication substitutes for the missing token.

**Why a token in a batch pull request's title would be worse than none.** A release or batch pull request delivers no single card. A card token in such a title would name a card the pull request does not deliver, and a resolver reading titles would attribute confidently and wrongly — a silent error, where the absent token merely leaves the card unresolved and visible as such. The scope restriction is therefore load-bearing, not politeness.

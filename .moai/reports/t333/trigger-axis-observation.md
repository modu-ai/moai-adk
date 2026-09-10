# t333 — trigger-axis observation (card t333, lane-17)

Measured 2026-08-27T14:5x UTC in worktree `.claude/worktrees/t333`, tree `d34a789a4`
(= `origin/develop` at measurement time). Read-only; no CI run was triggered by this
measurement itself.

## Claim

Card t314 rewired two guards for `develop` push and left their first firing as a **pending
opportunistic observation**. Both have now fired on `develop` and both succeeded — but the
two observations are **not of equal strength, and the difference is itself this card's
subject**. State them separately; do not fold them into one sentence.

**Claim A (strong) — `spec-lint`.** Fired on `develop` push, success, on three consecutive
pushes. Visible in an **unfiltered** `gh run list --branch develop`. Independently
reproduced by the lead session from its own query.

**Claim B (weaker) — `docs-i18n-check`.** Fired once on `develop` push, success. **Not
visible in the unfiltered listing**; recovering it required a query targeted at that
workflow by name. This lane ran that targeted query first-hand, so the firing is a direct
observation here — but it is NOT reproducible from the default listing, and the lead
session, querying the default way, could not see it.

t314's residual #3 is closable for `spec-lint` on Claim A. For `docs-i18n-check` it is
closable only on a measurement a reader must know to go looking for.

Nothing announced either firing. Both were found only because this lane went looking.

## Evidence

`spec-lint.yml` — fired on the three most recent `develop` pushes:

```
$ gh run list --branch develop --limit 15 \
    --json workflowName,event,status,conclusion,headSha,createdAt
08-27T14:51  SPEC Lint  push  success  d34a789a4
08-27T14:46  SPEC Lint  push  success  0c7457f8d
08-27T14:23  SPEC Lint  push  success  812ee01fc
```

`d34a789a4` is this lane's own t298 integration push, which touched `.moai/specs/**` —
the path filter t314 added. That push is what satisfied the opportunistic condition.

`docs-i18n-check.yml` — fired once on `develop` push:

```
$ gh run list --branch develop --workflow docs-i18n-check.yml --limit 5 \
    --json event,conclusion,headSha,createdAt
[{"conclusion":"success","createdAt":"2026-08-27T14:02:53Z","event":"push",
  "headSha":"343399d2f6c040fd4f7997bbfc757b8870fecc6e"}]
```

It is absent from the 15-run listing above because its `paths:` filter
(`docs-site/content/**`, `scripts/docs-i18n-check.sh`) matched none of the three most
recent pushes — absence there is a path-filter miss, not a trigger defect.

## Baseline-attribution

Commands run in this run, against this tree (`d34a789a4`), via the `gh` CLI against
`modu-ai/moai-adk`. The t314 residual being closed is quoted from
`.moai/reports/t314/verdict.md` lines 111-117 ("관측 대기 — 이 카드가 배선한 두
트리거의 첫 발화 … 운영자 판정(2026-08-27): 기회주의 관측으로 간다").

## Why this belongs to t333

This is the card's second empirical instance, and its closure demonstrates the thesis
rather than refuting it. The guards were rewired, they did fire, and they were correct —
and none of that reached anyone. The operator's decision was explicitly "opportunistic
observation", which names the gap precisely: the plan for confirming a guard fires was
*someone might notice later*. Had they NOT fired, the same plan would have produced the
same silence, and the absence would have read as green.

The distinction the card draws holds here in the sharpest possible form: a firing and a
non-firing were, until this measurement, indistinguishable from outside.

### A third instance, produced by measuring the first two

The gap between Claim A and Claim B is not a reporting blemish — it is another occurrence
of the defect, observed live while documenting the other two.

`docs-i18n-check` is absent from the default `gh run list --branch develop` listing. That
absence has (at least) two causes, and the listing does not separate them:

1. it did not run this round because its `paths:` filter did not match, or
2. its trigger is broken and it cannot run at all.

Both render as **not in the list**. The path-filter explanation above is an inference from
reading the workflow file; a coherent explanation is not an observation. Answering "when
did this guard last fire?" required a human to know which targeted query to run — and
knowing to run it depends on already suspecting the answer.

The lead session, querying the default way, reproduced Claim A and could not reproduce
Claim B. Two competent readers of the same repository, minutes apart, reached different
pictures of which guards are alive — not because either measured badly, but because the
surface does not carry the answer.

**And then the lead reproduced Claim B — but only after this lane handed it the query.**
It ran `gh run list --branch develop --workflow docs-i18n-check.yml` and got output
matching this file's, and it recorded why it could: the workflow's name reached it in a
message from this lane. It did not suspect the answer; it was told the question.

That is the complete form of this instance, and it is stronger than the two-readers
version. The gap did not close because a second observer looked harder. It closed because
one observer **handed the other the question**. Absent that hand-off, the lead's picture
would have stayed as it was, and — this is the load-bearing part — **nothing anywhere
would have told it the picture was wrong.** There is no signal for "your view of which
guards are alive is incomplete"; the incompleteness is silent by the same mechanism the
non-firing is.

A targeted query can only be issued by someone who already suspects the answer. Any design
that relies on a reader knowing which question to ask has therefore not solved this
problem — it has relocated it into whoever is expected to already know.

### The always-red variant (lead observation, out of scope, recorded)

`Graph Freshness` failed on every `develop` push in the window measured
(`d34a789a4`, `0c7457f8d`, `812ee01fc` — 3/3). It is card t322's subject and no part of
this card's scope.

It is recorded here because it is a second route to the same end state. This card's (c)
asks who sees the silence; a guard that is red on every single run stops being read just
as thoroughly as one that never runs — the signal is present, carries no information, and
is filtered out by every reader. Silence and constant noise are different mechanisms
arriving at the same place: nobody looks.

## Gaps (explicitly NOT observed)

- Whether either guard fires on a `develop` push whose paths do NOT match its filter —
  not measured, and by design it should not.
- Whether `spec-lint` would have caught a real defect in this run — only its exit status
  was read, not its findings.
- The release-PR filter question (t314 residual #1) and the `spec-status-auto-sync`
  hypothesis (t314 residual #2) — untouched by this measurement.
- Whether any OTHER guard in `.github/workflows/` has stopped firing. This measurement
  covered exactly the two guards t314 named.

## Residual risk

`gh run list` reports the runs GitHub retained; a run aged out of retention would read as
absence here. The window measured is hours old, so retention is not a live concern for
this observation — but a cadence check built on `gh run list` inherits that limit, and
the SPEC must state it rather than assume the API is a complete history.

Separately: Graph Freshness failed on all three `develop` pushes above
(`d34a789a4`, `0c7457f8d`, `812ee01fc`). That is a known inherited red, out of this
card's scope, and is recorded here only so a later reader does not mistake it for
something this measurement introduced.

# t332 plan-phase — tooling baseline (landed-determination)

Measured in the t332 worktree at `origin/develop` = `15453140a`, 2026-08-29.

## Claim

`moai todo pr`, as installed on this machine, asks `origin/main` rather than this
project's integration branch `develop`. Its `landed` column therefore answers a
question nobody posed: a card whose work landed on `develop` but not yet on `main`
reads `no-link`, which is indistinguishable from "nobody has started this".

## Evidence

```
$ grep -n 'worktree_base_branch' .moai/config/sections/git-strategy.yaml
5:    worktree_base_branch: develop

$ git log --oneline -S'LandedRefFor' -- internal/kanban/prlink_landed.go | tail -1
260ea5369 feat(SPEC-TODO-LANDING-STATE-001): M1 resolve the landed ref from the integration branch (t331)

$ moai version | grep -iE 'commit|built'
 v3.1.2   343399d2f   built 2026-08-27T14:07:38Z

$ strings ~/go/bin/moai | grep -c 'worktree_base_branch'
0

$ git log origin/develop --perl-regexp --grep='\bt342\b' --oneline | head -3
15453140a merge(WT-moving-ref-guard): SPEC-MOVING-REF-GUARD-001 — moving-ref invariant guard, advisory emission (t342)
38f937a4f docs(SPEC-MOVING-REF-GUARD-001): re-record MERGE_BASELINE_SHA after the second develop absorption (t342)
de5cc7b08 fix(SPEC-MOVING-REF-GUARD-001): emit MovingRefUnpinned advisory so --strict stops gating on the corpus (t342)

$ git log origin/main --perl-regexp --grep='\bt342\b' --oneline | head -3
(no output)

$ moai todo pr 2>/dev/null | awk -F'\t' '{print $2}' | sort | uniq -c
   5 landed
  91 no-link

$ moai todo pr 2>/dev/null | awk -F'\t' '$2=="landed"{print $1}' | tr '\n' ' '
t201 t204 t237 t278 t312
```

(96 rows carry a link column; the 4 recorded-relation lines have no second field.)

## Baseline-attribution

`internal/kanban/prlink_landed.go` in this tree (`origin/develop` 15453140a) defines
`DefaultLandedRef = "origin/main"` and resolves `origin/<worktree_base_branch>` via
`LandedRefFor`. That resolution arrived in `260ea5369`; the installed binary is
`343399d2f`, which predates it — hence the `strings` count of 0. The binary therefore
asks `origin/main`, while this project integrates on `develop`.

## Gaps

- Not observed: whether `make build` + reinstall makes the landed column correct. The
  fix is inferred from the source, not measured end-to-end.
- Not observed: how `moai todo pr` behaves for a queued card that IS landed on develop.
  The t342 cross-check used a card already removed from the queue by `done`.
- `moai todo pr <id>` for an id absent from the queue prints `queue is empty`. That is a
  message about the lookup, not evidence about the card.
- Not observed: whether `gh` is degraded here. The documented degradation notice goes to
  stderr with exit 0; this run did not isolate stderr cleanly, so the notice's presence or
  absence was NOT established.
- The 5 rows reading `landed` (t201 t204 t237 t278 t312) are landed against `origin/main`
  and were not re-checked against `origin/develop`; t278 is a `picked` card, out of the
  audit's scope.

## Residual-risk

A `no-link` reading is consistent with two different causes — the ref lag above, and a
degraded `gh` (which empties the link column and notes the degradation on stderr while
still exiting 0). Neither was excluded, and both may be in play. Symmetrically, a
`landed` reading proves landing on `main` only; it says nothing about `develop`. A run-phase landing verdict must therefore come
from a direct `git log <ref> --grep` plus `git merge-base --is-ancestor`, against BOTH
`origin/develop` and `origin/main`, and never from this column while the lag stands.

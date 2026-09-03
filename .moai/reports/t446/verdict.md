# t446 — card premise REFUTED, no SPEC authored

**Verdict: the work this card asks for is already landed.** Both the main item and the
side item were delivered by card t305 itself, in a commit that is an ancestor of both
this tree and `origin/develop`. No plan-phase SPEC was written; doing so would have
specified work that exists.

Measured in worktree `.claude/worktrees/t446`, branch `WT-long-consumer-guard`, at
`4e4607abe` (local `develop` tip; the tree was fast-forwarded from `origin/develop`
`d592b0551` so the premise is re-measured against current state, not against the state
the card was written from).

---

## Claim

Card t446 asserts: the `.Long` deferral introduced by t305 created a coupling that
nothing binds — the deferral is safe only while the help path is `.Long`'s single
reader, and if someone deletes `fang.WithoutManpage()` the `todo pr` / `todo done`
man pages ship with empty descriptions, silently. It also asserts a side item: the
R2 help test compares against `todoLandedRefOnce()` and would pass vacuously if both
sides were the empty string.

Both assertions are false as of this tree.

## Evidence

**The binding guard exists.** `internal/cli/todo_lazy_landedref_test.go:115-141`,
`TestLandedRefDeferralHasExactlyOneConsumerSurface`. Its doc comment states the
coupling in the card's own terms, and its failure message names the consequence rather
than the change:

```
fang.WithoutManpage() is gone from fang.go, which admits a SECOND reader of .Long.
`todo pr` and `todo done` defer their .Long until the help function runs (todo.go,
withResolvedLandedRef), so cobra's man generator — which reads .Long directly and
runs no help function — would emit them EMPTY.
Enabling man pages therefore means populating .Long on that path too, not just
deleting this option.
```

It also records its own limit in the comment: it reads source text because
`fang.Option` values are functions and cannot be compared, so a behaviour-preserving
rename would trip it and a differently-spelled equivalent would not.

**It is green, and it is not vacuous.** GREEN:

```
go test ./internal/cli/ -run 'TestLandedRef' -v -count=1
--- PASS: TestLandedRefNotMaterializedAtConstruction (0.00s)
--- PASS: TestLandedRefDeferralHasExactlyOneConsumerSurface (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.293s
```

RED under mutation — `sed -i '92d' internal/cli/fang.go` (deleting the
`fang.WithoutManpage(),` line, which is exactly the act the card fears):

```
--- FAIL: TestLandedRefDeferralHasExactlyOneConsumerSurface (0.00s)
    todo_lazy_landedref_test.go:137: fang.WithoutManpage() is gone from fang.go, ...
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.316s
```

The tree was restored from the `sed` backup and re-verified clean
(`git status --porcelain` → 0 lines) before this file was written.

**The side item is closed too.** `todo_lazy_landedref_test.go:62-64` and `:91-97` both
assert `ref != ""` before the `strings.Contains` comparison. The second carries the
measurement in a comment rather than an assumption:

```
strings.Contains(anything, "") is true, so without this the assertion below passes
for every possible usage string. Measured, not supposed: forcing todoLandedRefOnce
to return "" left this test green while its sibling above caught the same mutant.
```

**The other premise numbers still hold.** `GenManTree` / `GenMan` → 0 hits across
`*.go` (`rc=1`). `fang.WithoutManpage()` appears at `internal/cli/fang.go:92`, and its
only other mention is the guard test asserting its presence.

## Baseline-attribution

Delivering commit: `e6d964ff2` — *"test(cli): bind the deferral's manpage coupling, and
de-vacuum one of its guards (card t305)"*.

```
git merge-base --is-ancestor e6d964ff2 HEAD           → rc=0
git merge-base --is-ancestor e6d964ff2 origin/develop → rc=0
```

So the guard is not merely landed locally — it is already on the remote. The card was
raised as R1 during t305 and t305 then absorbed it; the card records the state before
that absorption.

## Gaps

- The guard's own stated limit was not independently probed: a *rename* of
  `WithoutManpage` that preserved behaviour would trip it (false positive), and an
  equivalent option under another spelling would not (false negative). Both are
  properties of source-text matching, acknowledged in the test's comment; neither was
  exercised here.
- No check was made for a THIRD `.Long` reader arriving by some path other than the
  man generator. The card scoped the risk to the man surface and this verdict stays
  inside that scope.
- The full `internal/cli` package was not run — only the `TestLandedRef` selector.
  Package-wide state is CI's verdict, not this one's.

## Residual risk

The guard binds the *option's presence in source*, not the *property* (exactly one
`.Long` consumer surface). Someone enabling man pages by a route that does not touch
that literal would not be stopped. That is a weaker binding than the card asked for —
but it is a binding, and closing the remaining distance is a different card with a
different premise, not this one.

## Recommendation

Drop t446 as already-delivered, or re-scope it to the residual above. Card mutation is
the lead's act; nothing was changed in the queue from this lane.

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

---

# t446 — RE-SCOPED (lead ruling, 2026-09-03): the residual is now the card

The verdict above stands: the card's ORIGINAL premise is refuted, and nothing above
is withdrawn. The lead adopted that refutation and re-scoped the card to the residual
this lane's own report raised — **the guard bound the presence of an option literal,
not the property "`.Long` has exactly one consumer surface".**

Measured in worktree `.claude/worktrees/t446`, branch `WT-long-consumer-guard`, at
`30802a5ed`. `origin/develop` had moved to `7835148d3` by the time of this pass; per
the lead's instruction this was a read-only re-confirmation with no re-absorption,
so nothing below is measured against that tip.

## Claim

1. The premise re-confirmation holds in THIS tree: `fang.WithoutManpage()` has exactly
   one production call site, and no man/markdown generator consumes `.Long` anywhere.
2. `TestLandedRefDeferralHasExactlyOneConsumerSurface` now binds the PROPERTY
   behaviourally — it runs the real `fangOptions()` through the real `fang.Execute`
   and asks whether fang registered its `man` command — instead of matching the string
   `"fang.WithoutManpage()"` in `fang.go`.
3. The guard carries a mandatory control that fails loudly if the probe ever stops
   detecting fang's man surface. The source-text version could not have this.
4. The rewrite is a strict improvement, not a lateral move: it removes the old guard's
   false-positive limit, demonstrated by mutation.
5. The R2 vacuity side item is closed at both call sites, demonstrated by mutation.
6. Production code is unchanged. The only diff is one test file.

## Evidence

**Premise re-confirmation (lead item 1).** In this tree, read-only:

```
$ grep -rn "GenManTree\|GenMan\|cobra/doc" --include="*.go" .
rc=1                                   # no man/markdown generator anywhere

$ grep -rn "WithoutManpage" --include="*.go" .
internal/cli/todo_lazy_landedref_test.go:136:  (the old guard's own assertion)
internal/cli/todo_lazy_landedref_test.go:137:  (its failure message)
internal/cli/fang.go:83:                       (doc comment)
internal/cli/fang.go:92:                       fang.WithoutManpage(),   <- sole production call site
```

**The mechanism the new probe rests on.** `charm.land/fang/v2@v2.0.1`, `fang.go:111-160`:
`Execute` starts from `settings{manpages: true}`, applies the options, and — while
`opts.manpages` — calls `root.AddCommand(&cobra.Command{Use: "man", ... RunE: ...
mango.NewManPage(1, cmd.Root())})` BEFORE `root.ExecuteContext`. `mango` reads `.Long`
directly and runs no help function. So "a second `.Long` consumer exists" has an
observable runtime form: a subcommand named `man` on the built root.

**Mutant A — the act the card fears (property genuinely broken).** `sed -i '' '92d'`
on `fang.go`, deleting `fang.WithoutManpage(),`:

```
=== RUN   TestLandedRefDeferralHasExactlyOneConsumerSurface
    todo_lazy_landedref_test.go:175: fangOptions() now yields a fang `man` command, which admits a SECOND reader of .Long.
        `todo pr` and `todo done` defer their .Long until the help function runs (todo.go, withResolvedLandedRef),
        so cobra's man generator — which reads .Long directly and runs no help function — would emit them EMPTY.
        Enabling man pages therefore means populating .Long on that path too, not just flipping this option.
--- FAIL: TestLandedRefDeferralHasExactlyOneConsumerSurface (0.00s)
```

RED. The control is a `t.Fatal` placed BEFORE this assertion, so reaching the `t.Error`
at line 175 is itself proof the control passed — the probe still detects fang's man
surface, and the failure is the property, not the detector.

**Mutant C — behaviour-preserving rename (the old guard's false positive).** This
mutant was NOT run in the original pass and is what shows the rewrite is an improvement
rather than a substitution. `fang.go:92` rewritten to a differently-spelled alias with
identical behaviour:

```
		withoutMan(),
...
var withoutMan = fang.WithoutManpage
```

```
$ grep -c 'fang.WithoutManpage()' internal/cli/fang.go
0                                       # rc=1 — the OLD guard would have gone RED here

=== RUN   TestLandedRefDeferralHasExactlyOneConsumerSurface
--- PASS                                # the NEW guard correctly stays GREEN
```

The property still holds under this mutant, and the new guard says so. The old guard
would have failed a change that broke nothing. That limit is now gone.

**Mutant B — R2 vacuity (lead item 2, second half).** `todoLandedRefOnce` forced to
return `""`:

```
=== RUN   TestHelpNamesTheResolvedLandedRef
    todo_lazy_landedref_test.go:66: todoLandedRefOnce resolved to the empty string
--- FAIL: TestHelpNamesTheResolvedLandedRef (0.00s)
=== RUN   TestUsageNamesTheResolvedLandedRef
    todo_lazy_landedref_test.go:99: todoLandedRefOnce resolved to the empty string
--- FAIL: TestUsageNamesTheResolvedLandedRef (0.00s)
```

Both sides go RED at their `ref == ""` fatal. Neither passes vacuously on the
`strings.Contains(anything, "")` path. The guard test itself is unaffected, correctly —
it does not read the ref.

**Restored state and package verdict.**

```
$ git status --porcelain
 M internal/cli/todo_lazy_landedref_test.go     # every mutant restored; only the intended edit

$ go test ./internal/cli/ -run 'TestLandedRef|TestHelpNamesTheResolvedLandedRef|TestUsageNamesTheResolvedLandedRef' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.489s

$ go vet ./internal/cli/        -> rc=0
$ gofmt -l internal/cli/        -> 28 files listed, NONE of them this file
```

The `gofmt` list is inherited red on this tree, not introduced here; it is the subject
of separate cards and was deliberately not repaired (scope discipline).

Full-package `go test ./internal/cli/ -count=1` was run once by the implementing agent
and reported `ok ... 1002.210s` — the package's wall-time exceeds a 600s ceiling on
this machine, which is a duration property, not a defect. This lane did not re-run it.

## Baseline-attribution

Every command above was run in this session, in worktree `.claude/worktrees/t446`,
branch `WT-long-consumer-guard`, HEAD `30802a5ed`, with the working tree clean apart
from the intended test-file edit. Dependency coordinates read from this tree's
`go.mod`: `charm.land/fang/v2 v2.0.1`, and `github.com/muesli/mango-cobra v1.3.0`
resolved from the module cache. `origin/develop` at read time: `7835148d3` (not merged
in; nothing here is measured against it).

Mutants A and C were run and restored by this lane directly. Mutant B was run by the
implementing agent; this lane read its output rather than re-running it, and says so
here rather than presenting it as its own measurement.

## Gaps

- Mutant B was not independently re-run by this lane (see above).
- The full `internal/cli` package was not re-run after the mutant round-trip by this
  lane; the selector subset was. Package-wide and cross-platform verdicts are CI's.
- `golangci-lint` was not run; `-race` was not run.
- No man page was actually generated and inspected. The chain
  `man` command -> `mango.NewManPage` -> `WithLongDescription(c.Long)` was read in
  fang's and mango's source, not executed end-to-end.
- The probe's isolation from package-level state was not proved beyond the package
  passing; it operates on a throwaway `&cobra.Command`.
- No check was made for whether the rewritten guard changes the package's run time.

## Residual risk

- **The named remaining hole.** The guard binds fang's man surface only. A second
  `.Long` reader arriving by another route — a `cobra/doc` generator wired up
  elsewhere, a hand-rolled emitter, or a future fang feature that reads `.Long`
  without registering a command named `man` — is not bound. The doc comment states
  this rather than leaving it to be discovered. Measured today that route has zero
  occupants (`grep` above), so this is an open door, not an open defect.
- **The control's failure is ambiguous.** If fang renamed `man` to something else, the
  control fires — correct, but a human must then distinguish "renamed" from "removed"
  before re-deriving the probe. The message says so.
- `fangOptions()` calls `version.GetVersion()`, so the probe inherits any build
  configuration in which that misbehaves; it would surface as a confusing probe
  failure rather than a version failure.

## Disposition

Re-scoped work is complete. The card's guard now binds the property the card names.
Card mutation and the `done` transition remain the lead's act; nothing was changed in
the queue from this lane, and nothing was pushed.

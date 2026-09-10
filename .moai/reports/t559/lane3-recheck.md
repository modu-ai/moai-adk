# t559 — lane-3 re-dispatch re-check (GH #1680)

Card t559 was re-dispatched to lane-3 after its work had already landed on
local `develop`. This document records what lane-3 verified, and the one
measured finding the re-check produced.

## Claim

1. The card's work is already on local `develop`, has landed on
   `origin/develop`, and is green in this tree.
2. The landed change fixes the issue's **T2** shape (marker exists ONLY below
   the project top) but NOT its **T0/T1** shape (a marker at the project top
   AND a module nested below it) — in the T0 shape the heavy gate still
   reaches zero Go coverage, which is GH #1680's headline reproduction.

## Evidence

Commands run in this worktree (`.claude/worktrees/t559`, branch
`WT-nested-lang-marker`), against this tree.

Landing ancestry — `d060e0d13` is the local `develop` tip this worktree was
fast-forwarded to:

```
$ git merge-base --is-ancestor 88cdf473d HEAD && echo ANCESTOR-OF-WT-HEAD=yes
ANCESTOR-OF-WT-HEAD=yes
$ git merge-base --is-ancestor 810b933b0 HEAD && echo VERDICT-COMMIT-ANCESTOR=yes
VERDICT-COMMIT-ANCESTOR=yes
$ git merge-base --is-ancestor 88cdf473d refs/heads/develop && echo ON-LOCAL-DEVELOP=yes
ON-LOCAL-DEVELOP=yes
$ git merge-base --is-ancestor 88cdf473d refs/remotes/origin/develop || echo ON-ORIGIN-DEVELOP=no
ON-ORIGIN-DEVELOP=no          # WRONG — read a stale remote-tracking ref
$ git fetch origin develop && git rev-parse --short refs/remotes/origin/develop
d060e0d13
$ git merge-base --is-ancestor 88cdf473d refs/remotes/origin/develop && echo ON-ORIGIN-DEVELOP=yes
ON-ORIGIN-DEVELOP=yes         # the attributable reading
```

The first origin reading is retained deliberately rather than deleted: it was
taken **before** `git fetch`, against a remote-tracking ref this worktree had
never updated, so it reported the state of a ref rather than the state of the
remote. `origin/develop` is `d060e0d13` — byte-identical to local `develop` —
and both card commits are on it. The card's work is fully landed, remotely
included. The lesson is the standing one: re-fetch before any ancestry
verdict; an un-fetched remote-tracking ref answers a question nobody asked.

The two landed commits: `88cdf473d` (`fix(gate): detect language markers below
the project top with a bounded recursive scan (t559, #1680)`) and `810b933b0`
(`docs(t559): record GH #1680 verdict evidence with before/after detection
counts (#1680)`).

Affected-package suites, re-run by lane-3 (not carried from the prior run):

```
$ unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR \
        MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/hook/quality/ ./internal/config/
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	50.392s
ok  	github.com/modu-ai/moai-adk/internal/config	25.512s
```

The T0-shape measurement (throwaway probe, deleted after the run — the tree is
byte-unchanged; the fixture is root `package.json` + `apps/id/go.mod` +
`apps/id/main.go`, the shape the issue reproduces with):

```
$ go test ./internal/hook/quality/ -run TestT559ProbeLayer2 -v
=== RUN   TestT559ProbeLayer2
    zz_t559_probe_layer2_test.go:35: MEASUREMENT t0-detected-markers=[package.json] t0-detected-root-relative=.
    zz_t559_probe_layer2_test.go:37: MEASUREMENT t0-bound-step-dir-relative=.
    zz_t559_probe_layer2_test.go:38: MEASUREMENT t0-go-toolchain-reached=false
--- PASS: TestT559ProbeLayer2 (0.01s)
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.752s
```

`t0-go-toolchain-reached=false` is the finding: the nested Go module is never
reached, so `go vet` / `go test` never run and a Go defect in `apps/id/` still
passes the heavy gate silently.

The mechanism, at source level (`internal/hook/quality/gate.go`
`detectToolchain`): a root-level match returns immediately, so the recursive
scan (`detectNestedToolchain`) is only ever consulted when the project top
carries no marker at all.

```go
if d, _ := g.matchToolchainAt(dir); d != nil {
    g.detectedRoot = d.root
    return d                      // root-level match short-circuits
}
d := g.detectNestedToolchain(dir) // reached only when the top has no marker
```

## Baseline-attribution

Every figure above was produced in this run, in this worktree, against
`HEAD = d060e0d13`. The ancestry checks and both package suites were invoked
from the worktree root. No number is carried from the prior session's
`verdict.md` — the suites were re-run rather than cited, and the T0
measurement did not exist before this re-check.

One figure was corrected in-run: the first `origin/develop` ancestry reading
was taken before this worktree fetched, and is superseded by the post-fetch
reading. The corrected value (`yes`) is the attributable one; the superseded
one is kept above with its cause named.

## Gaps

- The prior session's own `verdict.md` § Gaps already declared issue layer 2
  unimplemented. This re-check does not discover that gap; it **measures** it,
  turning a stated scope boundary into an observed behavior.
- No end-to-end `moai gate` run against a nested fixture on an installed
  binary (the issue's shim-based T0'/T1' matrix). The measurement here is
  unit-level, at the `detectToolchain` seam.
- Whether layer 2 (per-toolchain fan-out in `Run`) belongs to t559 or to a
  follow-up card was a scope decision, not a technical finding. Surfaced to
  the lead and **resolved there**: layer 2 is card **t596**, and t559 closed
  on the landed change. The deciding fact was the landing itself — a card
  already on `origin/develop` has nothing left to extend, so separation was
  the only available shape once the un-landed premise fell.

## Residual-risk

- Closing t559 on the landed change leaves GH #1680's headline reproduction
  live. A reader of the issue who tests T0 after the fix ships will observe
  the defect and reasonably conclude the fix did not work, even though the
  T2 half is genuinely repaired. The mitigation is a card-text obligation on
  t596, not a code change here: **#1680 must not be closed until t596
  lands.** Closing it on t559 alone would close an issue whose headline
  reproduction still reproduces.
- The most common monorepo shape in practice is precisely T0 (a JS/TS
  workspace at the top with services nested below), so the unfixed half is
  likely the more frequently encountered one.

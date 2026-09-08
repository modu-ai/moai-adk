# t359 — develop absorb evidence

Card `t359` / `SPEC-TODO-LANDING-EVIDENCE-001`, lane-8, integration window
`lane-8` acquired 2026-09-08T05:20:57Z.

Every figure below was measured in THIS tree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch
`WT-landing-evidence`) at merge commit `8a589d9df`, after absorbing local
`develop` at `768306d27`. The decisive lines are reproduced here rather than
whole logs; what was not exported is named under Residual-risk.

## Claim

The card branch absorbed local `develop`, the two conflicts were resolved, the
merge tree was re-measured, and it is green. The absorb surfaced TWO real
integration failures that neither branch showed alone.

## Evidence

### The absorb

```
$ git rev-parse --short refs/heads/develop        → 768306d27   (re-measured, not inherited)
$ git rev-list --count --left-right refs/heads/develop...HEAD  → 1253  36
$ git merge --no-edit refs/heads/develop          → rc=1, CONFLICT in 2 files
$ git rev-parse -q --verify MERGE_HEAD            → 768306d2779aae7bce52ac6a78ec0839a498a203
$ git diff --name-only --diff-filter=U | wc -l    → 2
```

Resolved, then:

```
$ git commit --no-edit                            → merge commit 8a589d9df
$ git status --porcelain | wc -l                  → 0
$ git rev-parse -q --verify MERGE_HEAD            → (empty — merge complete)
$ git merge-base --is-ancestor refs/heads/develop HEAD → rc=0
```

### Conflict 1 — `CHANGELOG.md`, both sides added

Both branches appended to `### Added` under `[Unreleased]`. Both kept.

```
$ grep -c '^<<<<<<<\|^=======\|^>>>>>>>' CHANGELOG.md   → 0
$ grep -c 'SPEC-TODO-LANDING-EVIDENCE-001\|SPEC-SPEC-LINT-BLIND-AXES-001\|SPEC-LLMCFG-PRESERVE-001' CHANGELOG.md → 3
```

### Conflict 2 — `internal/template/catalog.yaml`, and NEITHER side was correct

The conflicted line is the content hash of the `moai` skill directory. Both
branches edited `todo.md` inside that directory, so each side's hash is the
hash of ITS OWN input; neither is the hash of the merged input. Picking a side
would have been picking between two stale answers, and it would have been
silent: the merge completes, the YAML parses, and the drift check that would
have caught it runs in a later commit or not at all.

```
HEAD side     : 2e2cda9450b7210e9441e6e6a10efc83d02ebfe90f0d09656f981b7f0bc4cb98
develop side  : b1e0bce67e959d045a079a6aa31b8e848cd746144a5ba5ebe9a08657d49e83fd
$ go run ./internal/template/scripts/gen-catalog-hashes.go --all   → rc=0
regenerated   : 5514ce677c396541ab97ade05028be6e976c3643e755f5bc187985df4b31b9dc
```

The discriminant, stated so the next reader does not have to re-derive it: a
file a BUILD STEP OR A CHECK regenerates (`make build`, `*-emit-check`,
`gen-*`) has no correct side in a conflict — regenerate it.

### F8 — the inherited `agents-emit-check` blocker, measured resolved

```
$ make build                             → rc=0   (agents-emit-check ran as a prerequisite)
$ git status --porcelain | wc -l         → 0      (after the build)
```

The second line is the load-bearing one: `make build` regenerates
`catalog.yaml`, and a clean tree afterwards is what shows the regeneration
above produced the same bytes. F8 is measured resolved, not assumed to be.

### Integration failure 1 — `internal/kanban` did not compile

```
internal/kanban/prlink_landed_attribution_test.go:117:15: undefined: LandedGrepArgs
FAIL	github.com/modu-ai/moai-adk/internal/kanban [build failed]
```

Cause: develop's SPEC-TODO-LANDING-ATTRIBUTION-001 renamed
`LandedGrepArgs(ref, cardID)` to `LandedSubjectArgs(ref)`, removing the card id
from the landing query's argv — the query became a subject stream and the
attribution predicate moved into Go.

Ported WITHOUT weakening: the old single query carried two properties (the
implementation's own argv, and SHAs for the identity check) that the new API
cannot carry together, because `LandedSubjectArgs` streams subjects only and
names no SHA. The control is now two queries — the implementation's builder
asserting the stream is non-empty (so a silent-empty regression in it is still
caught here), and a fixture-premise query built in the test asserting exactly
three commits name the card and each fixture SHA is among them.

Second failure, at `todo_landed_attribution_test.go:170`
(`outcome kind = "no-link", want "landed"`): the fixture constant was
`t359fixture`, and develop's tokenizer defines a card as `\bt[0-9]+\b`, so a
suffixed id is not a card token to it at all. Changed to `t9359`, with the
reason recorded in the comment.

```
$ go test ./internal/kanban/... -count=1 -timeout 900s   → rc=0
$ /usr/bin/grep -c -- '--- FAIL\|^FAIL' kanban3.txt      → 0
ok  	github.com/modu-ai/moai-adk/internal/kanban	143.160s
```

### Integration failure 2 — `internal/cli` did not compile

```
internal/cli/todo_landed_doc_test.go:130:27: unknown field landedFor in struct literal of type spyRunner
internal/cli/todo_pr_landing_test.go:112:49: unknown field landedFor in struct literal of type spyRunner
internal/cli/todo_pr_landing_test.go:235:43: unknown field landedFor in struct literal of type spyRunner
internal/cli/todo_pr_landing_test.go:299:41: unknown field landedFor in struct literal of type spyRunner
internal/cli/todo_pr_landing_test.go:344:49: unknown field landedFor in struct literal of type spyRunner
```

Same rename, one layer out: with no card in the argv the stub can no longer key
its answer on the argv, so develop replaced the card-keyed
`landedFor map[string]bool` with the positional `logPlan []spyLogAnswer` — call
N answers the Nth card that REACHES the landed query.

Delegated to `manager-develop` (`t359-mergefix`) rather than inferred, because a
positional plan mapped onto the wrong cards yields a test that PASSES while
asserting the wrong thing — the signature defect this card has been repairing.
The agent established the mapping with a throwaway probe (deleted after use)
that installed, at each position in turn, a subject stream naming every seeded
card, and read back which card rendered `landed`. Two facts it settled that
inference would have gotten wrong:

- `logCalls` is CUMULATIVE across renders. Four of the five sites render twice
  (`pr`, then `pr --json`), and an exhausted plan falls through to
  `return "", nil` — so a plan sized for one render makes the second render's
  cards answer "not landed" silently.
- At `:344` the first seeded card never reaches the query at all — it matches
  `pinnedPRJSON` and short-circuits — so position 0 answers `ids[1]`. A plan
  written on the seeding order would have put the landed answer on the wrong
  card.

### The merged tree, re-measured by the lane orchestrator

Independently re-run after the delegation returned; the agent's own figures are
not the basis of this verdict.

```
$ go test -c -o /dev/null ./internal/cli/      → rc=0, 0 bytes of output
$ go test -c -o /dev/null ./internal/kanban/   → rc=0, 0 bytes of output
$ go vet ./internal/cli/... ./internal/kanban/...  → rc=0, 0 lines
$ golangci-lint run ./internal/cli/... ./internal/kanban/...  → rc=0, "0 issues."
$ go test ./internal/cli/... -count=1 -timeout 1800s  → rc=0
    unfiltered FAIL scan (/usr/bin/grep -c -- '--- FAIL\|^FAIL')  → 0
    empty-sweep guard  (grep -c 'no test files\|no tests to run') → 0
    ok lines                                                      → 17
    ok  	github.com/modu-ai/moai-adk/internal/cli	580.933s
```

The empty-sweep guard is recorded because a package with no test files, and a
selector matching no tests, both print `ok`; a green run whose swept set was
empty asserts nothing.

The compile check exists because the editor's diagnostics reported
`logPlan` / `spyLogAnswer` / `landedLogLine` as undefined while `go vet` was
clean. Two readings disagreed, so neither was assumed: compiling the test
binaries settled it, and the diagnostics were stale.

## Baseline-attribution

Tree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t359`, branch
`WT-landing-evidence`, merge commit `8a589d9df` (re-read immediately before
each commit, never inherited). Absorb source: local `develop` at `768306d27`,
re-measured at absorb time rather than carried from the dispatch. Every figure
above is from this run against this tree.

## Gaps

- `-race` was NOT run. The lead reports a `Race Test` red on `origin/develop`
  at `a4855f0b2` in `TestSyncGitSpecStatuses_…` (a TempDir cleanup race), which
  lives in `internal/cli/spec_status_test.go` — the very package measured here.
  This card touches no part of that file, and a non-race run can neither
  confirm nor deny that race. It is named, not resolved.
- No cross-platform build or test: darwin/arm64 only.
- Packages outside `internal/cli` and `internal/kanban` were not measured. The
  full-suite verdict is CI's, on the head that actually merges.
- No coverage measurement.
- The sync-audit findings F2 and F5-F9 were neither touched nor re-measured by
  this absorb.
- The full `go test` logs live in session scratch and were NOT exported; only
  the decisive lines above are exported, and the scratch copies are expected to
  vanish. They are named here as a known loss and are not offered as the basis
  of any claim.

## Residual-risk

- The `internal/cli` port is verified by the package suite passing, not by a
  mutant per ported site. A plan that happens to align correctly for a reason
  other than the one the probe established would pass identically.
- The probe that established the call-to-card mapping was deleted after use, so
  the mapping is reproducible only by rebuilding it; the probe's verbatim
  output is recorded in the delegation report rather than in a runnable form.
- The catalog hash is correct for THIS merged input. Any later edit inside the
  `moai` skill directory changes it again, and a merge that does not regenerate
  reintroduces exactly the defect described above.

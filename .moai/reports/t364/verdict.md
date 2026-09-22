# t364 — `internal/lockfile` runs zero tests on windows

Card t364 · lane-12 · worktree `.claude/worktrees/t364`, branch `WT-lockfile-windows`.
Base: local `develop` `b7462203ac2d81a7e4b5716b30e9a426e5e946de` (fast-forwarded from
`origin/develop` `f7cabfc29`, which is 32 commits behind).

Verdict: **REPAIRED.** The card's premise held under investigation — the cause is a
build-constraint exclusion, and it is a defect rather than an intentional exclusion.

---

## Claim

1. `internal/lockfile` executes no tests on windows because its **only** test file,
   `lockfile_unix_test.go`, carries `//go:build !windows`. Not a `t.Skip`, not a
   `TestMain`, not a compile failure — the file is excluded from the build, leaving
   the package with zero test files on that GOOS.
2. The exclusion is **not intentional**. `lockfile_windows.go` ships a real
   `Lock`/`Unlock` pair with `fan_in >= 3` recorded in its own `@MX:ANCHOR`
   (`taskledger.ClaimTask`, `settings.go`, `glm_tools.go`). That implementation has
   had zero executed coverage.
3. The windows `skipped=332` vs ubuntu `skipped=100` gap is **unrelated to this
   card**. `internal/lockfile` contributes **0** skipped rows on windows.
4. After the repair, windows has one test file compiling three tests, so the
   census's `nothing-ran` count for windows drops from 4 to 3 and the three test
   names appear.

## Evidence

### C1 — the cause

`go list -json ./internal/lockfile`, run in this worktree at `b7462203a`:

| GOOS | GoFiles | TestGoFiles | IgnoredGoFiles |
|---|---|---|---|
| windows | `lockfile_windows.go` | `null` | `lockfile_unix.go`, `lockfile_unix_test.go` |
| darwin | `lockfile_unix.go` | `lockfile_unix_test.go` | `lockfile_windows.go` |
| linux | `lockfile_unix.go` | `lockfile_unix_test.go` | `lockfile_windows.go` |

The CI stream states it directly. From the windows leg of run `33308057570`
(artifact `test-stream-release-verify-windows-latest`, not expired, downloaded
2026-09-02) — the package's complete event set is three events:

```
{"Action":"start","Package":".../internal/lockfile"}
{"Action":"output","Package":".../internal/lockfile","Output":"?   \tgithub.com/modu-ai/moai-adk/internal/lockfile\t[no test files]\n"}
{"Action":"skip","Package":".../internal/lockfile","Elapsed":0}
```

`[no test files]` is the cause in the runner's own words. The same package on the
ubuntu leg emits `run`/`pass` pairs for `TestFlock` and `TestLockUnlockRoundTrip`.

### C2 — not an intentional exclusion

`lockfile_windows.go` is production code, not a stub: it maintains a
`map[string]*sync.Mutex` keyed by absolute path and is reached through the same
exported `Lock`/`Unlock` symbols as the Unix path. Its own comment documents the
cross-process limitation as a *deliberate behaviour*, which is precisely the kind of
contract that wants a test. The Unix test file's header explains why **`TestFlock`**
is Unix-only (it asserts `flock(2)` kernel semantics via `syscall` directly) — that
justification is sound and is left in place. Nothing was ever written for the
windows path.

### C3 — the skipped-count decomposition

Both legs' streams were re-counted here from the raw artifacts, using the census's
own rule (`Action=="skip"` **with** a `.Test` field):

```
win total: 332      ubu total: 100
```

Both reproduce the census report's published totals exactly, which attributes this
extraction to the same measurement the card cites.

`grep -c lockfile windows-skips-per-package.txt` → `0`. The package contributes no
skipped rows on windows, because a package with no test binary emits no per-test
events at all.

Largest contributors to the +232 (win − ubu), full tables committed alongside this
file as `windows-skips-per-package.txt` / `ubuntu-skips-per-package.txt`:

```
59 win=92 ubu=33  internal/cli
44 win=50 ubu=6   internal/hook
26 win=27 ubu=1   internal/kanban
14 win=15 ubu=1   internal/cli/update/deploy
14 win=14 ubu=0   internal/lsp/subprocess
11 win=13 ubu=2   internal/statusline
11 win=11 ubu=0   internal/cli/update/backup
```

This is a separate axis — windows-conditional `t.Skip` calls across the CLI, hook,
and kanban packages. **Not repaired here**; it is not this card's subject and would
be a different card.

### C4 — the repair

One new file, `internal/lockfile/lockfile_contract_test.go`, carrying **no** build
constraint. No existing file was modified — `md5 lockfile_unix.go` is
`81d5688a77794cf8776daca9fc9ebbb3` before and after, and `git status --short` shows
exactly one added path.

It asserts the contract the two implementations genuinely share — mutual exclusion
keyed by file path within one process, and release on `Unlock` — through the
exported API only, so it compiles and runs on every GOOS. Choosing the shared
contract over a `//go:build windows` file is what makes the tests locally
mutation-provable on darwin; a windows-tagged file could only have been
compile-checked from here.

After the change:

| GOOS | TestGoFiles |
|---|---|
| windows | `lockfile_contract_test.go` |

- `GOOS=windows go vet ./internal/lockfile` → rc 0
- `GOOS=windows go test -c -o …` → rc 0, a 4,175,872-byte windows test binary
  (stronger than vet: the test binary itself links)
- `go test -race -count=1 ./internal/lockfile/` on darwin → `ok … 1.443s`, 5 tests
- `gofmt -l internal/lockfile/` → no output; `go vet` → rc 0

### C5 — mutation: what the tests actually see

Both mutants were applied to `lockfile_unix.go` on darwin and reverted (md5 verified
above).

| Mutant | Change | Result |
|---|---|---|
| A | `Lock` → `return nil` | `FAIL … second handle acquired the lock while the first held it` (0.35s) |
| B | `Unlock` → `return nil` | `FAIL … second handle did not acquire the lock within 5s of release` (5.6s) |

`TestLockExcludesSecondHandleWhileHeld` catches both directions on its own, which is
why the round-trip test was reduced to a plain success-path check rather than
duplicating the exclusion machinery.

### C6 — a defect found in the test itself

The first draft of this test file **passed** its mutation runs but took **123.96s**
under mutant B, where it should have reported in ~5s. Cause: the draft deferred
`f2.Close()`, and closing a handle whose `Lock` is still blocked in flight wedges
the test binary until the `go test` timeout. In CI that shape burns a full job
timeout instead of failing fast, on any OS.

Repaired by closing the second handle only on paths where its `Lock` has
demonstrably returned, and deliberately leaving it open on the timeout path. Re-run
after the fix: mutant B fails in **5.6s** (from 124s), mutant A in 0.35s, baseline
green. The reasoning is recorded in the test file's own comment so the next reader
does not "simplify" the missing `defer` back in.

## Baseline-attribution

Every figure above was measured in this worktree at `b7462203a`, except the CI
census totals, which come from the artifacts of run `33308057570` (head `09dd1bee9`)
and were re-derived here rather than quoted — the re-derivation reproducing the
published 332/100 is what ties the two trees together.

## Gaps — what was NOT observed

- **Windows execution.** No windows runner was available. `TestGoFiles` becoming
  non-empty and the test binary linking are local, deterministic facts; the actual
  execution of the three tests on windows — and therefore the census `nothing-ran`
  dropping 4 → 3 with the three names present — is a **CI observation that has not
  yet happened**. It lands on the lead's batch push. The card's regression [HARD]
  is satisfied in prediction, not in observation, and must not be reported as
  observed until that run is read.
- **The `+232` skipped gap** was decomposed but not investigated. No claim is made
  about whether those windows skips are correct.
- **`internal/lockfile` on windows under contention.** The in-process-mutex
  implementation's documented cross-process limitation is unchanged and untested by
  design — testing it would require multiple processes, which is out of scope here.
- **The other three `nothing-ran` packages** (`cmd/moai`,
  `internal/template/scripts`, `scripts/convert-nextra-to-hextra`) are the expected
  zero-test-file baseline and were not touched.

## Residual risk

The negative window in `TestLockExcludesSecondHandleWhileHeld` is 250ms of wall
clock. Its failure direction is safe — a false FAIL requires the lock to have been
genuinely granted while held — but on a heavily loaded runner the goroutine may not
reach `Lock` within the window, making that half vacuous for that run. The positive
half (acquire within 5s of release) still asserts, and mutant A is caught by the
negative half at 0.35s locally, so the pair is not simultaneously vacuous under any
timing observed here. If windows CI shows this test flaking, the negative window is
the first thing to look at.

If the windows `Lock` implementation turns out to be broken in a way the shared
contract does not cover, this repair closes the *observability* gap without closing
a *correctness* gap — which is the correct scope for this card, but should not be
read as "windows locking is now verified correct".

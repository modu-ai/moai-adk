# t352 — reproduction record: `t.TempDir` cleanup race class

Card: t352 · Lane worktree: `.claude/worktrees/t352` · Branch: `WT-tempdir-cleanup-race`
Base: `origin/develop` @ `77b2bcae6` · Measured: 2026-08-28

The card's precondition was reproduction. This file records what was reproduced, what was not,
and the mechanism established for each of the two observations the card carries.

The lane entered on `e08d5e55c`; `origin/develop` advanced 22 commits during the investigation and
the branch was fast-forwarded to `77b2bcae6` before the figures below were taken. Every cited
source line was re-read on the new base and is unchanged; every measurement below was re-run there.

---

## Claim

- **Observation 2 (`TestBinaryLag_OneSeamServesBothSurfaces`, `.moai/state: directory not empty`)
  is reproduced, and its mechanism is established.** `Handle` dispatches a durable write into the
  caller's directory **after** the join that releases it, so the write can land after the test body
  has returned and `t.TempDir` cleanup has begun. Measured late in 3 of 4 configurations.
- **The intermittency is explained, not merely observed.** The write's lateness scales with the
  scan cost. On an empty directory — the condition of the actual failing test — the scan sometimes
  finishes inside the join budget and sometimes does not, which is the shape of an intermittent
  failure rather than a deterministic one.

  **Correction (2026-08-28, after plan audit).** An earlier revision of this file, and the lane's
  first report to the lead, described that shape as matching "1 CI failure in 5 runs". That figure
  is **not observation 2's**. `.moai/reports/t322/verdict.md:281` attributes it to
  `TestGitDiffNameCount_Predicate` — observation **1**, the instance this lane did NOT reproduce.
  Observation 2 (`TestBinaryLag_OneSeamServesBothSurfaces`) has **one** CI observation and no
  measured frequency at all (`t322/verdict.md:282`, "appeared", count 1). The mechanism finding is
  unaffected — it rests on the source ordering and the probe, not on any CI frequency — but the
  inference that the mechanism "matches the observed 1-in-5 shape" was unfounded and is withdrawn.
- **Observation 1 (`TestGitDiffNameCount_Predicate`, `.git/objects: directory not empty`) is NOT
  reproduced.** No post-command writer was found in the fixture tree on either macOS or Linux. Its
  cause remains unestablished, and the "one class, two instances" reading in the card is therefore
  **supported for one instance only**.

## Evidence

### Observation 2 — mechanism, read from source (`77b2bcae6`)

`internal/cli/binary_lag_test.go:57` calls `hook.NewSessionStartHandler(nil).Handle(...)` with
`ProjectDir` set to a `t.TempDir()`. Inside `Handle`:

- `internal/hook/session_start.go:1460` — `var deferredScansAsync = true`.
- `internal/hook/main_test.go:47` flips it to `false`, but that is the **`internal/hook` test
  binary's** `TestMain`, and the variable is package-private. A test in `internal/cli` runs a
  different binary, so it sees the production value `true`.
- `internal/hook/session_start.go:240` therefore takes the async branch:
  `spawnDeferredAdvisoryScans(...)`, joined with `deferredScanJoinBound` (250 ms).
- `internal/hook/session_start.go:608-611` — the goroutine sends the advisory into the buffered
  channel **first**, and only then runs `runMXColdStartScan(projectDir)`. The surrounding comment
  states the ordering is deliberate ("dispatched AFTER the advisory result is sent so it never
  delays advisory keys").
- `internal/hook/session_start.go:1606-1613` — `runMXColdStartScan` writes
  `<projectDir>/.moai/state/mx-index.json`.
- `mxIndexNeedsRebuild` returns `true` for any fresh `t.TempDir()` (index absent), so the write is
  always attempted.

The join releases `Handle` as soon as the advisory arrives, while a durable write into the caller's
directory is still ahead of it. Whether that write completes before `Handle` returns is a race
between the remaining join budget and the scan.

### Observation 2 — measured

A probe (`internal/cli/zz_t352_probe_test.go`, throwaway) creates a directory, calls `Handle`,
checks for the index at the instant `Handle` returns, then polls. `nFiles` pads the tree to vary
`ScanDir` cost:

```
$ go test ./internal/cli/ -run TestT352Probe_LateWrite -v -count=1 -timeout 600s
late=false nFiles=0    Handle=59.713209ms: index already present at return
late=true  nFiles=200  Handle=61.565500ms: index ABSENT at return, appeared 11.010417ms later
late=true  nFiles=2000 Handle=60.051625ms: index ABSENT at return, appeared 43.680791ms later
late=true  nFiles=8000 Handle=59.593792ms: index ABSENT at return, appeared 222.742875ms later
--- PASS: TestT352Probe_LateWrite (1.61s)
```

3 of 4. The lateness grows with scan cost, which is why a loaded CI runner under `-race` widens the
collision window.

The same probe on the lane's entry base `e08d5e55c`, before the fast-forward, returned `late=true`
on all four rows including `nFiles=0` (`index ABSENT at return, appeared 6.162125ms later`). The
`nFiles=0` row is therefore **not stable across runs** — it is itself the race, and reporting it as
deterministic would have been wrong. The three padded rows reproduced identically on both bases.

Cross-package caller inventory — **test** callers, one call site:

```
$ grep -rn 'hook\.NewSessionStartHandler' --include='*_test.go' internal/ | grep -v '/hook/'
internal/cli/binary_lag_test.go:57
internal/cli/zz_t352_probe_test.go:41    (the throwaway probe itself)
```

**Correction (2026-08-28, after plan audit).** That grep was scoped to `_test.go`, so calling its
result "the blast radius" was wrong. There is also a **production** cross-package caller:

```
$ grep -rn 'NewSessionStartHandler' --include='*.go' internal/ cmd/ | grep -v '_test.go'
internal/cli/deps.go:221:  deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))
internal/hook/session_start.go:41:func NewSessionStartHandler(cfg ConfigProvider) Handler {
```

The constructor is **non-variadic** — one parameter. Any seam added to it as an extra argument
therefore touches a production call site unless it is introduced as a variadic option parameter,
which keeps `deps.go:221` compiling unchanged. This is a constraint on the fix, not a finding about
the race.

### Observation 2 — the symptom alone does not reproduce locally

```
$ go test ./internal/cli/ -run TestBinaryLag_OneSeamServesBothSurfaces -race -count=50 -timeout 900s
ok  github.com/modu-ai/moai-adk/internal/cli  5.601s   (rc=0)
```

50 iterations on `e08d5e55c`, no failure. This is why the mechanism, not the symptom, is the
reproduction: the collision needs a slow cleanup or a slow scan, and a fast local disk supplies
neither.

### Observation 1 — not reproduced, on either platform

A probe replicating `newCheckFixture` + `TestGitDiffNameCount_Predicate`'s churn (39 tracked
fixtures, two commits) snapshots the tree when the last command returns, waits 750 ms, snapshots
again, then runs the `RemoveAll` that cleanup performs:

macOS (host, `git version 2.50.1 (Apple Git-155)`):

```
iter=0..4  postChanges=0 []  removeAllErr=""
--- PASS: TestT352Probe_FixtureLateWrite (6.05s)
```

Linux container (`golang:1.26`, aarch64, worktree bind-mounted):

```
$ docker run --rm -v <worktree>:/src -w /src -v /Users/goos/go/pkg/mod:/go/pkg/mod golang:1.26 \
    go test ./internal/graph/ -run TestT352Probe_FixtureLateWrite -v -count=1 -timeout 300s
iter=0..4  postChanges=0 []  removeAllErr=""
--- PASS: TestT352Probe_FixtureLateWrite (3.87s)
ok  github.com/modu-ai/moai-adk/internal/graph  3.873s
```

Zero post-command changes in 10 iterations across two operating systems. The whole package under
the race detector is likewise clean:

```
$ docker run --rm ... golang:1.26 go test ./internal/graph/... -race -count=5 -timeout 1800s
ok  github.com/modu-ai/moai-adk/internal/graph         2.798s
ok  github.com/modu-ai/moai-adk/internal/graph/symbol  1.173s
```

The `internal/graph` test files contain no goroutines and no `t.Parallel()`; the only subprocess is
`gitFix`, which uses `cmd.Output()` and waits:

```
$ grep -rn 'go func\|t.Parallel()\|exec.Command\|hook\.' internal/graph/*_test.go
internal/graph/codequery_test.go:145:  return exec.Command("mkfifo", path).Run()
internal/graph/check_test.go:22:       cmd := exec.Command("git", ...)
```

The in-process writer hypothesis is therefore ruled out for observation 1, and the detached-child
hypothesis the t322 verdict named ("a detached git background child is the obvious suspect") is
**not supported** by these measurements: no background writer touched the tree in the 750 ms after
the last command returned, on either platform.

## Baseline-attribution

- Tree: `.claude/worktrees/t352`, HEAD `77b2bcae6`, `git rev-list --count --left-right
  origin/develop...HEAD` → `0 0` after a fetch taken immediately before the figures above; working
  tree carrying only the two untracked paths listed under Gaps.
- Host: darwin/arm64, `go 1.26.4` (per `go.mod`), `git version 2.50.1 (Apple Git-155)`.
- Container: `golang:1.26` (docker 29.4.0), linux/aarch64, worktree bind-mounted at `/src`, host
  module cache at `/go/pkg/mod`.
- Every figure is the verbatim output of the command shown. The three graph-probe blocks and the
  `-count=50` block were measured on `e08d5e55c`; the `internal/cli` probe block and every source
  line citation were re-measured on `77b2bcae6`. Which base each figure belongs to is stated where
  it differs.
- Probe sources: `internal/cli/zz_t352_probe_test.go` (retained for the run phase) and
  `internal/graph/zz_t352_probe_test.go` (removed after measurement; the output quoted above is the
  only surviving record of it).

## Gaps

- **CI itself was not re-run.** Every measurement here is local (host or local container). The CI
  figures in the card are carried from `.moai/reports/t322/verdict.md` and were not re-measured.
  Note the two observations carry different evidence weight there: observation 1 has a frequency
  (1 of 5 post-landing runs, `t322/verdict.md:281`); observation 2 has a single appearance and no
  frequency (`t322/verdict.md:282`). See the Correction under Claim.
- **Observation 2's CI frequency is unknown, not low.** One appearance is not a rate. Nothing here
  establishes how often the race actually bites in CI, so the cost of leaving it unfixed is
  likewise unquantified.
- **The graph-probe figures were not re-taken on `77b2bcae6`.** They were measured on `e08d5e55c`
  and are reported as such; the 22 intervening commits were not examined for changes to
  `internal/graph` fixtures.
- **The container is aarch64, CI is x86-64.** The Linux measurements were not repeated under
  emulation on `linux/amd64`, so an architecture-specific writer is not excluded for observation 1.
- **`internal/cli`'s full package was not run.** Only the two named `-run` selectors executed.
- **No fix has been attempted or measured.** This file records reproduction only.
- **Observation 1's cause is unestablished.** Ruled out: an in-process goroutine in the package, a
  parallel test in the package, and a background writer detectable within 750 ms on macOS or Linux
  aarch64. Not ruled out: an x86-64-specific or runner-specific writer, a writer appearing later
  than 750 ms, or a cleanup interaction that is not a writer at all.
- **The "single class" reading is one instance short.** Two observations share a symptom string;
  only one has a demonstrated mechanism. They are not shown to share a cause.
- **Untracked at the time of writing:** `.moai/reports/t352/` and
  `internal/cli/zz_t352_probe_test.go`.

## Residual risk

- **A fix scoped to observation 2 leaves observation 1's flake in `develop`.** The card's stated
  cost — a flake surfacing as red under an unrelated card's name at integration — is only halved by
  the work this reproduction justifies.
- **The established mechanism has a production caller, and future ones inherit the race.**
  `Handle`'s deferred writer is fire-and-forget by design because the production hook process
  exits; any in-process caller that owns and then deletes its project directory inherits the same
  race. The known caller `internal/cli/deps.go:221` registers the handler in a long-lived process
  but does not delete a project directory, so it is not exposed today — that is a property of what
  it does, not a guarantee about what a later caller will do.

  *(This bullet superseded an earlier one reading "the inventory above covers `_test.go` callers
  only", which was true of the first inventory and false after the Correction under Evidence added
  the production caller. Recorded rather than silently replaced.)*
- **The probe measures the write, not the collision.** It establishes that the write lands after
  `Handle` returns; it does not itself produce the `unlinkat ... directory not empty` error, so the
  final link from mechanism to the observed CI failure string remains an inference — a
  well-supported one, but an inference.
- **The `nFiles=0` row flipped between the two bases.** Nothing here identifies what changed; the
  honest reading is that the empty-directory case sits on the boundary of the join budget, so its
  outcome is machine- and load-dependent rather than base-dependent.

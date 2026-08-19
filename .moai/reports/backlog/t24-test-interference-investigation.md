# t24 — `TestNavigatorEnrich_AtomicWriteBarrier` flake investigation

Read-only investigation. No source was edited.

- Repo: `/Users/goos/MoAI/moai-adk-go`
- Branch: `main`, HEAD `3b9b3bf99`
- Date: 2026-08-15
- Platform: darwin (Darwin 25.5.0)

---

## Claim

**There is no cross-test interference. The card's premise is wrong.**

`TestNavigatorEnrich_AtomicWriteBarrier` (`internal/cli/navigator_enrich_test.go:65`) is an
independently flaky test. It fails at a measured rate of **3/60 (5%)** when run entirely
alone with `-run TestNavigatorEnrich_AtomicWriteBarrier`, with no other test in the package
executing. Under CPU contention the rate rises (1/40 observed with 8 busy-loop hogs; the
rate is load-sensitive, not load-gated).

The cause is a **race inside the test's own synchronization handshake**, not shared state,
not `os.Chdir`, not env pollution, not a config cache, and not parallel tests.

### Mechanism

`atomicWrite` (`internal/cli/navigator_enrich.go:128-149`) performs, in order, inside the
test's goroutine:

1. `os.WriteFile(tmp, …)` — creates `capability-symbols.md.tmp` (line 130)
2. `os.Unsetenv("NAVIGATOR_PRE_RENAME_BARRIER")` (line 139)
3. `os.WriteFile(barrier, []byte("ready"), …)` — creates the barrier flag (line 140)
4. spin-wait until the barrier flag is removed (lines 141-145)
5. `os.Rename(tmp, path)` (line 148)

The test (`navigator_enrich_test.go:82-92`) waits for **step 1** and then asserts **step 3**
has already happened:

```go
tmp := filepath.Join(root, ".moai", "project", "codemaps", "capability-symbols.md.tmp")
deadline := time.Now().Add(2 * time.Second)
for time.Now().Before(deadline) {
        if _, err := os.Stat(tmp); err == nil {
                break                       // ← breaks as soon as step 1 lands
        }
        time.Sleep(5 * time.Millisecond)
}
if _, err := os.Stat(barrier); err != nil {
        t.Fatalf("barrier file not created (goroutine did not reach barrier)")   // ← step 3
}
```

The window between step 1 and step 3 is unsynchronized. Whenever the poll loop's 5 ms tick
happens to land inside that window — i.e. the goroutine is descheduled between the two
`os.WriteFile` calls — the test observes `tmp` present and `barrier` absent and fails
immediately.

The test is waiting on the **wrong file**: `tmp` is not the signal that the goroutine
reached the barrier; the barrier flag itself is.

**This is not a timeout.** Every observed failure completed in 0.14 s – 0.34 s, an order of
magnitude below the 2 s deadline. The deadline never expired in any failing run.

---

## Evidence

### E1 — Isolated run, 5 iterations (the card's baseline) — PASS 5/5

```
$ go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=5 -v
=== RUN   TestNavigatorEnrich_AtomicWriteBarrier
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.07s)
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.06s)
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.06s)
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.07s)
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.06s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.200s
EXIT=0
```

5 iterations is too small a sample to see a 5% failure rate (P(all pass) ≈ 0.77). The
card's "passes 5/5 in isolation" is consistent with an independently flaky test.

### E2 — Full package, run 1 — barrier test PASSED

```
$ go test ./internal/cli/ -count=1 -timeout 900s
EXIT=1
FAIL	github.com/modu-ai/moai-adk/internal/cli	228.160s
```

Only failure in the run:

```
--- FAIL: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (5.75s)
```

Barrier test line in the same log:

```
navigator-astx: enriched 1 row(s) -> /var/folders/.../TestNavigatorEnrich_AtomicWriteBarrier2121499165/001/.moai/project/codemaps
```

(reached completion → passed)

### E3 — Full package, run 2, verbose — barrier test PASSED

```
$ go test ./internal/cli/ -count=1 -timeout 900s -v
EXIT=1
--- FAIL: TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey (7.06s)   ← only failure
...
=== RUN   TestNavigatorEnrich_AtomicWriteBarrier
navigator-astx: enriched 1 row(s) -> /var/folders/.../TestNavigatorEnrich_AtomicWriteBarrier2385120118/001/.moai/project/codemaps
--- PASS: TestNavigatorEnrich_AtomicWriteBarrier (0.23s)
```

2293 test results in the run. The barrier test passed in **both** full-package runs, so the
card's "failed twice in a full-suite run" did not reproduce as a full-suite-specific effect.
Note the runtime under full-suite load (0.23 s) vs isolated (0.06 s) — load slows it ~4x but
stays far under the 2 s deadline.

Tests immediately preceding it in the sequential run (from the verbose log) — none of these
touch the navigator barrier path:

```
--- PASS: TestNewMxScanCmd_PathScansSubtree (0.01s)
--- PASS: TestNavigatorEnrich_EmitsFilesWhenCapabilityMapExists (0.34s)
--- PASS: TestNavigatorEnrich_AbsenceIsGraceful (0.09s)
=== RUN   TestNavigatorEnrich_AtomicWriteBarrier
```

### E4 — REPRODUCTION: isolated, under CPU contention — FAIL 1/40

```
$ for i in 1..8; do (while :; do :; done) & done
$ GOMAXPROCS=1 go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=40
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.14s)
    navigator_enrich_test.go:91: barrier file not created (goroutine did not reach barrier)
    testing.go:1464: TempDir RemoveAll cleanup: unlinkat /var/folders/.../TestNavigatorEnrich_AtomicWriteBarrier1254619968/001/.moai: directory not empty
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	9.339s
FAIL
```

**Decisive**: the `-run` filter means no other test in the package executed. The failure
therefore cannot be caused by another test.

The trailing `TempDir RemoveAll cleanup: ... directory not empty` is a *consequence*, not a
second defect: `t.Fatalf` ends the test while the goroutine is still spinning on the barrier,
so cleanup races the still-live goroutine's `os.Rename`.

### E5 — CONTROL: isolated, NO artificial load — FAIL 4/40

```
$ go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=40
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.16s)
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.33s)
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.20s)
--- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.34s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	8.943s
FAIL
```

The flake needs no artificial load at all.

### E6 — Rate + failure-site tally, 60 isolated iterations

```
$ go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=60 -v \
    | grep -E "navigator_enrich_test.go:[0-9]+:|^--- (PASS|FAIL)" | sort | uniq -c | sort -rn
      …
      3     navigator_enrich_test.go:91: barrier file not created (goroutine did not reach barrier)
      1 --- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.22s)
      1 --- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.17s)
      1 --- FAIL: TestNavigatorEnrich_AtomicWriteBarrier (0.14s)
```

- Failure rate: **3/60 = 5.0%**, isolated, unloaded.
- **Every** failure is at `navigator_enrich_test.go:91` — the `barrier`-absent assert.
- **Zero** failures at line 95 (`final file created before barrier removed`) or line 107.
- All failures complete in 0.14–0.22 s; the 2 s deadline never fires.

### E7 — Shared-state candidates enumerated and ruled out

```
$ grep -rn "os.Setenv\|os.Unsetenv\|os.Clearenv\|os.Chdir" internal/cli/*_test.go
internal/cli/cc_test.go:498,499,571:  os.Unsetenv(config.EnvMoaiKanban / EnvMoaiKanbanSpec)
internal/cli/coverage_test.go:430-928:            os.Chdir (×13, all with defer/restore)
internal/cli/coverage_improvement_test.go:1061-2344: os.Chdir (×10+, all t.Cleanup-restored)
```

```
$ grep -rn "atomicWrite(" internal/cli/
internal/cli/navigator_enrich.go:110,114   ← only callers of the barrier-aware atomicWrite
internal/cli/navigator_enrich.go:128       ← definition
internal/cli/preference/{decay,filestore}.go  ← a DIFFERENT function in a different package;
                                              no NAVIGATOR_PRE_RENAME_BARRIER handling
```

| Candidate | Verdict | Reason |
|---|---|---|
| `os.Chdir` mutation | **ruled out** | `runNavigatorEnrich` is called with an explicit absolute `root` (`t.TempDir()`); cwd is never consulted on this path (`navigator_enrich.go:61-71` only falls back to `os.Getwd` when `projectRoot == ""`). And E4/E5 reproduce with no other test running. |
| Env var set without `t.Setenv` | **ruled out** | The only raw `os.Unsetenv` in cli tests is `EnvMoaiKanban*` (unrelated key). The barrier test uses `t.Setenv`. E4/E5 reproduce in isolation. |
| Shared temp path not under `t.TempDir()` | **ruled out** | All three navigator-enrich tests use `t.TempDir()`; observed paths are per-test (`…/TestNavigatorEnrich_AtomicWriteBarrier<nonce>/001/`). |
| Singleton config/disk cache (SPEC-HOOK-PRETOOL-PERF-001) | **ruled out** | The barrier path reads no config; reproduces in isolation. |
| Parallel tests racing | **ruled out** | The barrier test calls `t.Setenv`, so it is necessarily non-parallel; Go resumes `t.Parallel()` siblings only after the sequential batch finishes. And E4/E5 reproduce with a `-run` filter admitting one test. |
| Another test stealing the barrier env (the `os.Unsetenv` at `navigator_enrich.go:139` is a process-global mutation) | **not the observed cause** | Structurally real — the barrier is a process-global one-shot — but only `navigator_enrich.go:110/114` can consume it, and no failure required a second test. Kept as residual risk. |

---

## Baseline-attribution

Every measurement above was taken in this run, against this tree:

- Tree: `/Users/goos/MoAI/moai-adk-go`, branch `main`, `git rev-parse --short HEAD` → `3b9b3bf99`
- Working tree at measurement time (`git status --short`): only untracked report/asset files
  (`.moai/reports/diagram-*.html`, `.moai/reports/webredesign/`, `.playwright-mcp/`,
  `settings-identity.png`) — no modified tracked source.
- Toolchain: system `go` on darwin; no build tags, no `-race`, `-count` as shown per command.
- Full-package runs used `-timeout 900s`; both completed in ~228 s (no hang, no timeout).
- Total isolated iterations of the barrier test executed: 5 (E1) + 40 (E4) + 40 (E5) + 60 (E6) = 145.
- Aggregate isolated, unloaded failure rate (E5 + E6): 7/100 = 7%. E6 alone: 3/60 = 5%.

---

## Gaps

Explicitly **not** observed:

- **The card's original full-suite failure was not reproduced.** Both full-package runs
  (E2, E3) passed the barrier test. The claim that it "failed twice in a full-suite run" is
  taken from the card, not measured here. The isolated flake I did reproduce is a sufficient
  explanation for those failures, but I did not witness them.
- **No `-race` run was performed** on either the isolated test or the full package.
- **No `go test -cpu` sweep** (e.g. `-cpu 1,2,8`) was run to characterize the rate vs GOMAXPROCS.
- **The known statusline/CwdGuard hang cluster was not observed at all today.** Both full-package
  runs completed in ~228 s without hanging, so I have no live data on those tests; my
  ruling-out below is on mechanism, not on a reproduced hang.
- **`TestHandleCodexReviewGate_LiveCodexBlocksInjectionAndKey` failed in both full-package runs**
  (5.75 s / 7.06 s). I did not investigate it — it is outside t24's scope, but it means the
  `internal/cli` package is currently red on `main` for an unrelated reason.
- **`astx.EnrichRows` internals were not read**; I did not verify whether it consults the cwd
  or any global. Not needed for the diagnosis (the failure precedes any dependence on it),
  but it is unexamined.
- **I did not instrument the step-1→step-3 window** (that would require editing source, which
  this investigation forbids). The mechanism is established by code reading plus the fact that
  every failure lands on exactly the assert that closes that window.

---

## Residual-risk

- **The one-shot global barrier is a genuine latent hazard even though it did not cause this
  failure.** `atomicWrite` calls `os.Unsetenv` (`navigator_enrich.go:139`) on the process-wide
  environment, so the barrier is consumed by whichever `atomicWrite` call reaches it first,
  process-wide. If a future test in package `cli` calls `runNavigatorEnrich` concurrently while
  the env is set, it will steal the barrier and the barrier test will then fail at line 95
  (`final file created before barrier removed — atomic-write broken`) — a *misleading* message
  suggesting the atomic-write implementation is broken when it is not. No failure of that shape
  was observed in 145 iterations.
- **The `t.Fatalf` at line 91 leaves the goroutine spinning.** On failure the goroutine is still
  in the unbounded `for { os.Stat(barrier) }` loop (no sleep, no yield, no timeout) while
  `t.TempDir` cleanup deletes the tree underneath it. Observed as
  `TempDir RemoveAll cleanup: … directory not empty` in E4. In a full-suite run that leaked
  spinner burns a CPU for the remainder of the package run, which plausibly *contributes to*
  the known statusline/CwdGuard slowness cluster even though it does not share its cause.
- **The failure rate is machine- and load-dependent.** 5–10% here; it could be materially
  higher on a busier CI runner or under `-race`, and correspondingly lower on an idle fast box.
  A CI green run is not evidence the flake is gone.
- **Fix-verification will need a large sample.** At a 5% base rate, a 20-iteration post-fix run
  passes 36% of the time by luck. Verify any fix with `-count=200` or more.

### Ruling on the statusline / CwdGuard hang cluster

**Ruled OUT as the same cause.** Three independent grounds:

1. Different failure mode — that cluster *hangs* (never returns); this test *asserts and
   returns* in 0.14–0.34 s.
2. This failure reproduces with a `-run` filter admitting exactly one test, so it needs no
   other test present; a shared cause with a cluster of other tests is excluded by construction.
3. The failing assert is on a file written by this test's own goroutine, within a window this
   test's own poll loop fails to cover.

A weak *one-directional* coupling remains plausible in the other direction (previous bullet):
the leaked spinner this test strands on failure adds CPU pressure to the rest of the package
run. That is this test affecting others, not others affecting it.

---

## Recommended fix (described, NOT applied)

**Primary — wait on the barrier flag, not on the `.tmp` file.** The barrier flag is the
goroutine's own "I have reached the barrier" signal and is written *after* the `.tmp` file, so
polling it closes the window completely. In `internal/cli/navigator_enrich_test.go:82-92`,
poll `barrier` in the deadline loop instead of `tmp`, and treat deadline expiry (rather than a
single post-loop `os.Stat`) as the failure. The `.tmp` file can still be asserted afterwards
if the test wants to state that a partial file exists — by then it provably does.

This is a **test-only change**. The production `atomicWrite` ordering is correct as written:
`.tmp` before the barrier flag before the rename is exactly the sequence the test is meant to
verify. Do not reorder `navigator_enrich.go` to satisfy the test.

**Secondary — bound the spin loop so a failing test cannot strand a goroutine.**
`navigator_enrich.go:141-145` spins with no sleep and no ceiling. Adding a small
`time.Sleep` per iteration and an overall timeout (after which it proceeds to the rename)
would stop a `t.Fatalf` from leaving a CPU-burning goroutine behind for the rest of the
package run. This one does touch production code, but only a test-hook path.

**Tertiary — note the one-shot global.** The `os.Unsetenv` self-consumption
(`navigator_enrich.go:139`) makes the barrier a process-global single-use latch. Worth a
comment, or a per-call mechanism, before a second concurrent caller is ever introduced.

**Verification for any fix**: `go test ./internal/cli/ -run TestNavigatorEnrich_AtomicWriteBarrier -count=200`,
then repeat with 8 background CPU hogs and `GOMAXPROCS=1`. Expect 0 failures in both.
Baseline for comparison: 7/100 failures on `3b9b3bf99` (E5 + E6).

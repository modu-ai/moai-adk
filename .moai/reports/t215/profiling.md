# t215 — `moai statusline` warm-path decomposition (internal measurement)

Date: 2026-08-27 · Tree: branch `WT-statusline-cost` (develop tip 22df80e90), worktree `.claude/worktrees/t215` · Machine: Apple M4 Max, 16 P-cores, serial runs (minimal concurrent load)

## Claim

The statusline warm-path wall time is **~236 ms per render on this machine/tree**, and **≈100% of it is mechanically attributed**: ~93% is the five serialized git subprocess spawns on the render path, ~7% is Go process boot. Every in-process pipeline stage other than the git collectors totals under 0.3 ms. There is no unknown residual compute cost; the earlier hook-audit figures (~0.55 s) measured a different load/window and are not reproduced here.

## Evidence

### 1. External wall timing — deployed chain vs components (N=40 each, fixture stdin redirect)

Harness: `bin/timeit_harness.py` (2 untimed priming runs per candidate, then N timed). Payload: Claude Code v2.1.246-shaped JSON (rate_limits present → usage collector gated off at `builder.go:371`; forge `none` in this repo's statusline.yaml → no detached refresh child, t293 gates active).

```
$ python3 bin/timeit_harness.py bin/timing_spec.json 40
true_floor(fork+exec+harness):        N=40 median=1.7ms   p95=2.0ms   min=1.4ms max=2.1ms
local_bin_version(startup floor):     N=40 median=16.2ms  p95=17.1ms  min=14.7ms max=19.1ms
installed_bin_version(startup floor): N=40 median=16.7ms  p95=20.2ms  min=15.5ms max=22.2ms
local_bin_statusline(render):         N=40 median=240.2ms p95=269.6ms min=222.0ms max=295.6ms
installed_bin_statusline(render):     N=40 median=235.8ms p95=262.8ms min=224.4ms max=312.6ms
sh_wrapper_chain(deployed):           N=40 median=235.5ms p95=250.9ms min=223.9ms max=260.2ms
```

Installed binary = v3.1.3-rc.5 built from commit **22df80e90** — byte-equivalent code to this tree minus ldflags; all three forms agree within ±3 ms.

Baseline attribution: this run, this tree, this day. The 2026-08-24 audit numbers (`cc-statusline.sh` median 1.03 s / standalone 0.61–0.66 s) were NOT re-run and are carried as reference only.

### 2. Per-phase in-process decomposition (`internal/statusline/profile_bench_test.go`)

```
$ MOAI_PROFILE_PHASES=1 go test ./internal/statusline/ -run TestProfilePhaseDistributions -v
build_end_to_end       N=120 median=167.245ms p95=192.860ms max=268.124ms
builder_init_new       N=120 median=58.703ms  p95=63.313ms  max=74.582ms
stdin_parse            N=120 median=0.008ms   p95=0.013ms   max=0.035ms
git_collect_status     N=120 median=161.759ms p95=189.820ms max=222.226ms
version_check_update   N=120 median=0.053ms   p95=0.066ms   max=0.398ms
instant_collectors     N=120 median=0.279ms   p95=0.427ms   max=0.613ms
snapshot_write         N=120 median=0.000ms   p95=0.000ms   max=0.000ms
render                 N=120 median=0.009ms   p95=0.010ms   max=0.015ms
```

Deployed per-render model = process start (~16 ms) + `New(opts)` with repo auto-open (58.7 ms) + parallel-collect critical path (= `git_collect_status`, 161.8 ms; version check 0.05 ms) + instant+parse+snapshot+render (< 0.35 ms) ≈ **237 ms** — matches the external medians above within measurement noise.

Note: `build_end_to_end` (167 ms) reuses one Builder, so it excludes per-render builder init that production pays every process; the deployed model adds it separately.

### 3. Where the git time goes — direct child command timing (N=30)

```
$ python3 bin/timeit_harness.py bin/git_timing_spec.json 30
git rev-parse --git-dir:                          median=29.5ms  ← NewRepository spawn 1/2 (manager.go:43)
git rev-parse --show-toplevel:                    median=29.1ms  ← NewRepository spawn 2/2 (manager.go:48)
git symbolic-ref --short HEAD:                    median=36.8ms  ← CollectGitStatus via CurrentBranch (manager.go:64)
git status --porcelain:                           median=86.5ms  ← CollectGitStatus via Status (manager.go:82)
git rev-list --count --left-right @{upstream}...HEAD: median=36.4ms  ← Status ahead/behind (manager.go:113)
git --version(no repo access):                    median=19.3ms  ← pure git binary startup floor
true(exec floor):                                 median=1.8ms
```

Arithmetic closure: init spawns 29.5+29.1 = 58.6 ms ↔ measured phase 58.7 ms. Status phase 36.8+86.5+36.4 = 159.7 ms ↔ measured phase 161.8 ms (remainder ≈ 2 ms/run of Go-side `exec.LookPath` + environ copy in `execGit`, manager.go:258).

Every bare git invocation carries a **19.3 ms binary-startup floor** before any repo work; repo-sized work adds the rest (`status --porcelain` dominates at 86.5 ms).

### 4. CPU profile confirms wait-bound, not compute-bound

```
$ go test ./internal/statusline/ -run '^$' -bench 'BenchmarkStatuslineWarmPath$' -benchmem -benchtime=60x \
    -cpuprofile bin/prof/cpu.out -o bin/prof/statusline.test
BenchmarkStatuslineWarmPath-16   60   162020357 ns/op   195920 B/op   2347 allocs/op

$ go tool pprof -top -nodecount=25 ...
Duration: 10.10s, Total samples = 200ms (1.98%)
   70ms 35.00%  syscall.rawsyscalln          ← blocked in child-wait syscalls
   20ms 10.00%  runtime.pthread_cond_wait
  50ms 25.00% (cum) core/git.(*gitManager).Status
  60ms 30.00% (cum) statusline.(*defaultBuilder).Build → collectAll.func1
  (full list: bin/prof/pprof_top25.txt, gitignored scratch)
```

The Go process spends **only 1.98% of benchmark wall on-CPU**; ~98% is waiting on child processes. All captured CPU samples sit on the `Build → collectAll → execGit` edge. Renderer, theme/gradient, ANSI processing, config load, stdin parse are each sub-0.01 ms phases — none of them is a hotspot.

## Named conclusion

| Component | Median | Share of ~236 ms external wall |
|---|---|---|
| Go process boot + cobra dispatch | 16.2 ms | ~7% |
| Builder init: 2× `git rev-parse` | 58.7 ms | ~25% |
| Git status collection: 3 spawns (`symbolic-ref`, `status --porcelain`, upstream `rev-list`) | 161.8 ms | ~68% |
| Everything else in-process (parse, version, instant collectors, snapshot write, render) | < 0.35 ms | < 0.2% |
| **Explained total** | **≈237 ms** | **≈100%** |

**Unexplained residual: none measurable on this run** (sub-0.35 ms aggregate outside the attributions above). The historical ~0.55–0.61 s standalone figures from the 2026-08-24 audit correspond to a different machine-load window; since even `git --version` costs 19.3 ms here, five git spawns scale multiplicatively under load — but that scaling direction is attribution, not a fresh measurement (marked as hypothesis, not observed today).

Candidates from the dispatch checklist, checked explicitly:
- renderer wide-string/ANSI/gradient: measured 0.009 ms median — cleared.
- update-check probes: cached YAML read, 0.053 ms — cleared.
- OTEL init: no otel imports in cmd/moai, cli/root.go, or statusline package (grep, static) — absent.
- sleep/backoff reaching sync render: no `time.Sleep` in any non-test statusline file (grep, static) — absent.
- telemetry writes: snapshot write 0.000–0.02 ms — cleared.
- refresh child spawning into the synchronous path: cannot happen — spawn is detached and gates on stale cache + segment enabled + `forge != none`; this repo has `forge: "none"` (t293 M2/M3/M7).
- **usage OAuth path** (only when stdin LACKS `rate_limits`): statically bounded at `usageFetchTimeout = 3s` (usage.go:40) + keychain spawn (usage.go:405). Never exercised live here (would touch network/keychain). This is the only plausible source of multi-second *tail* latencies (audit max 3.10 s), irrelevant to the warm median because Claude Code ≥ 2.1.80 always supplies `rate_limits`.

Fix-shape implication (diagnosis only, not implemented): the 5 spawns split as 2×(always-duplicated `rev-parse`) + 3×(git status family); collapsing/skipping idle-work spawns and dropping the ahead/behind probe are the levers with measured ceilings attached above. Reducing render FREQUENCY (task #1/#3) attacks the same wall-time × frequency product.

## Gaps

- Primary-checkout renders were not measured (guard forbids git operations there from this worktree session); all git timings are against this worktree. Absolute `git status` cost varies with worktree size/dirtiness.
- The rate_limits-absent OAuth/keychain path was analyzed statically (line citations), never executed.
- Audit-window conditions of 2026-08-24 (machine load, concurrent sessions) are not reproducible; its numbers were not re-measured.
- CPU profile attributes samples only; wall shares came from the phase timers. GPU/compositor-side effects were not observable from here.

## Residual-risk

Percentages drift on machines where git startup or `status --porcelain` differ; the composition (5 spawns ≈ dominant share; in-process stages negligible) holds structurally. The installed-binary comparison assumes binary equivalence at commit 22df80e90 (version string verified; ldflags build metadata differs only).

Reproduce:

```bash
go build -o bin/moai ./cmd/moai
python3 bin/timeit_harness.py bin/git_floor_spec.json 30      # fixtures/specs recreated from this doc
MOAI_PROFILE_PHASES=1 go test ./internal/statusline/ -run TestProfilePhaseDistributions -v
go test ./internal/statusline/ -run '^$' -bench . -benchmem
```

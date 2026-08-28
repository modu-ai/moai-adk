# SPEC-TEMPDIR-CLEANUP-RACE-001 — progress

Card: t352 · Branch: `WT-tempdir-cleanup-race` · Tier S

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
- Evidence base: `.moai/reports/t352/reproduction.md` (base `origin/develop` @ `77b2bcae6`).
- SPEC ID regex self-check executed as Bash; output `PASS`.
- Mechanism rung chosen in `plan.md` §D.1: rung 1 (exported synchronous-deferred-scan option),
  constrained to a **variadic** option parameter so `internal/cli/deps.go:221` keeps compiling
  unchanged (`plan.md` §D.3).
- Plan audit iter-1: PASS 0.88; iter-2: PASS 0.93 (monotonic). Repairs D1-D8 landed at spec version
  0.1.1, D9-D10 at 0.1.2 (`.moai/reports/t352/plan-audit.md`); the Tier S budget overrun (9 ACs, separate `acceptance.md`) is recorded and
  justified in `plan.md` §D.4 rather than resolved by deleting a criterion.
- AC-TCR-002b base SHA at plan time: `git merge-base origin/develop HEAD` → `77b2bcae6`
  (`origin/develop` tip had already moved to `c6aa61346`; the three-dot form pins the fork point).
- Status: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Measured on HEAD `410f6241d` (branch `WT-tempdir-cleanup-race`), host darwin/arm64, go 1.26.4.
Run-phase entry HEAD was `05678676e` (a merge of `origin/develop` taken at entry).

**Base-SHA attribution, with the mid-card merge noted (`acceptance.md` § Base-SHA attribution).**
`git merge-base origin/develop HEAD` → `48d8ef4bee768645d1a14a53eb5e4ba85170d447`. The plan-time
base was `77b2bcae6`; the run-phase-entry merge of `origin/develop` into this branch moved the
merge-base forward to `48d8ef4be`, which is the semantically correct baseline for this card's diff
and is recorded here because the change is otherwise silent.

### AC PASS/FAIL matrix

| AC | Status | Verifying command | Observed output |
|----|--------|-------------------|-----------------|
| AC-TCR-001 | PASS | `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=5 -timeout 600s` | `ok  github.com/modu-ai/moai-adk/internal/cli  13.405s`, rc=0 (`.moai/reports/t352/ac001.txt`) |
| AC-TCR-002a | PASS | `grep -n 'var deferredScansAsync = true' internal/hook/session_start.go` | `1546:var deferredScansAsync = true` — exactly one line, rc=0 |
| AC-TCR-002b(i) | PASS | `git diff origin/develop...HEAD -- internal/hook/session_start.go` | 5 hunks (`@@ -35,11 +35,70 @@`, `-82,7 +141,7`, `-251,7 +310,7`, `-490,7 +549,7`, `-503,7 +562,7`): one block adding the Option type + variadic constructor + `asyncDeferredScans`, and four single-line call-site changes. A grep of the diff body for `joinTimer`, `deferredScanJoinBound`, `spawnDeferredAdvisoryScans`, `advisoryCh` on `+`/`-` lines returned **nothing** — the async branch body and the join bound are unmodified (`.moai/reports/t352/ac002b-session-start.diff`) |
| AC-TCR-002b(ii) | PASS | `git diff --name-only origin/develop...HEAD -- internal/cli/deps.go` | prints **nothing** (assertion). Positive control `git diff --name-only origin/develop...HEAD -- internal/cli/binary_lag_test.go` prints `internal/cli/binary_lag_test.go` — non-empty, so the pathspec form is sound and the emptiness is a genuine unchanged-file result |
| AC-TCR-003 | PASS | `go test ./internal/cli/ -run TestBinaryLag_OneSeamServesBothSurfaces -race -count=20 -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/cli  3.552s`, rc=0; `grep -c 'directory not empty'` → `0` (`.moai/reports/t352/ac003-binlag.txt`). **Non-regression only** — 50 iterations passed on the pre-fix tree too, so this is NOT evidence the defect is fixed; AC-TCR-004 is |
| AC-TCR-004a | PASS (guard observed RED) | mutation, then `go test ./internal/cli/ -run TestSessionStartDeferredWriteDoesNotOutliveHandle -count=1 -timeout 600s` | rc=1, verbatim below (`.moai/reports/t352/ac004a-red.txt`) |
| AC-TCR-004b | PASS | same command, mutation reverted | `ok  github.com/modu-ai/moai-adk/internal/cli  2.800s`, rc=0 (`.moai/reports/t352/ac004b-green.txt`) |
| AC-TCR-005 | PASS | `go test ./internal/hook/ -race -count=1 -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/hook  34.813s`, rc=0; `found unexpected goroutines` absent (`.moai/reports/t352/ac005-hook-race.txt`) |
| AC-TCR-006 | PASS (compile only) | `GOOS=windows GOARCH=amd64 go vet ./internal/hook/... ./internal/cli/...` and the same with `GOOS=linux` | no output, rc=0 for each (`.moai/reports/t352/ac006-windows.txt`, `ac006-linux.txt`). Per the AC's own scope note this proves **compilation**, not behaviour, on those platforms |
| AC-TCR-007 | PASS | `go test ./internal/cli/ -count=1 -timeout 900s` | `ok  github.com/modu-ai/moai-adk/internal/cli  250.260s`, rc=0; `grep -c 'RESIDUE GUARD FAIL'` → `0`; `ls -d internal/cli/.moai` → `No such file or directory` (`.moai/reports/t352/ac007-cli-package.txt`) |

**AC-TCR-007 wall-clock (`spec.md` §D, CI headroom).** The whole `internal/cli` package reports
**250.260s** on this host without `-race`. The guard contributes roughly 2.7s per run (measured:
13.405s for `-count=5`). CI's per-package default is 600s. The repo comment at `ci.yml:232` citing
`internal/cli ~379s/70%` under `-race` remains an unverified carry — it was not re-measured here,
and this figure is a non-race host run, so the two are not comparable. What is measured is the
addition: seconds, not minutes.

### RED evidence

**(a) Step-1 RED — the guard run against the tree BEFORE the seam existed** (HEAD `05678676e`,
guard's `Handle` call had no option because none existed yet). Verbatim
(`.moai/reports/t352/red-step1.txt`):

```
--- FAIL: TestSessionStartDeferredWriteDoesNotOutliveHandle (1.92s)
    session_start_deferred_write_test.go:131: a durable write outlived Handle: 1 entr(ies) appeared under the caller's directory during the 1s settle window after Handle returned: [.moai/state/mx-index.json]
        REQ-TCR-001 requires every durable write into the caller's ProjectDir to complete before Handle returns.
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.764s
```

**(b) AC-TCR-004a mutation RED — the seam present, removed at the guard's own call site.** The
mutation, applied and then reverted:

```diff
-	if _, err := hook.NewSessionStartHandler(nil, hook.WithSynchronousDeferredScans()).Handle(context.Background(), &hook.HookInput{
+	if _, err := hook.NewSessionStartHandler(nil).Handle(context.Background(), &hook.HookInput{
```

Verbatim result (`.moai/reports/t352/ac004a-red.txt`):

```
--- FAIL: TestSessionStartDeferredWriteDoesNotOutliveHandle (2.17s)
    session_start_deferred_write_test.go:131: a durable write outlived Handle: 1 entr(ies) appeared under the caller's directory during the 1s settle window after Handle returned: [.moai/state/mx-index.json]
        REQ-TCR-001 requires every durable write into the caller's ProjectDir to complete before Handle returns.
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.010s
```

Both reds name the entry that appeared, and it is the writer the reproduction established. The
guard has therefore been seen to fail **for the stated reason**, not merely to fail
(`verification-completeness.md` §2).

### Scope decision — the guard-liveness second unawaited path

`plan.md` was written before card t333 (`SPEC-GUARD-LIVENESS-001`) merged into this branch at
run-phase entry. That work added a SECOND unawaited path to `Handle`
(`guardLivenessRefresh` at `session_start.go:85`, "never awaited" by design), branching on the SAME
private seam via `deferredScansAsyncEnabled()`. Including `binaryLagAdvisory` and
`guardLivenessAdvisory`, the seam now has **four** readers.

**Decision: the option covers all four.** Reasoning, stated rather than assumed:

- Only ONE of the four writes into the caller's `ProjectDir` — the deferred advisory scan's
  `runMXColdStartScan`. `guardLivenessRefresh` and `guardLivenessAdvisory` persist under the
  `~/.moai` state tree (`internal/guardliveness/store.go:16,61`), explicitly outside every evaluated
  working tree per REQ-GDL-008; `binaryLagAdvisory` runs only `git rev-parse` /
  `git merge-base --is-ancestor` (`internal/binlag/binlag.go:101,111`) and writes nothing. So the
  other three cannot reproduce this SPEC's defect, and AC-TCR-001's ProjectDir entry-set comparison
  would not catch them.
- They are covered anyway, for a different reason: an option named
  `WithSynchronousDeferredScans` that still left three goroutines running past `Handle` would be
  misnamed, and those goroutines are a genuine leak for any caller that checks for one
  (`internal/cli` has no goleak `TestMain`, so nothing there would notice).
- The cost is exactly what `plan.md` §D.3 pre-authorized for `binaryLagAdvisory` — "do it only if it
  costs one argument" — applied by symmetry to the pair that landed later. Each of the three
  package-level helpers takes one `async bool`; their six in-package test call sites pass
  `deferredScansAsyncEnabled()`, which is byte-equivalent to what the function computed internally
  before. `session_start_guard_liveness_render_test.go:160`'s source-literal grep for
  `"appendAdditionalContext(out, guardLivenessAdvisory("` still matches.

Files touched by the decision: `session_start_binary_lag.go`, `session_start_guard_liveness.go`, and
the two guard-liveness test files (mechanical call-site updates only).

### Probe disposition

`internal/cli/zz_t352_probe_test.go` is **deleted**. It measured the same write the guard now
measures, but asserted nothing (it only `t.Logf`'d), so retaining it alongside the guard would keep
an always-green instrument next to the assertion that replaces it. Stated in the M1 commit message.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: "410f6241d"
run_status: audit-ready
ac_pass_count: 9
ac_fail_count: 0
preserve_list_post_run_count: 1   # internal/cli/deps.go — verified absent from the diff (AC-TCR-002b(ii))
l44_pre_commit_fetch: "git fetch origin develop; git rev-list --count --left-right origin/develop...HEAD -> 0 3"
l44_post_push_fetch: "not performed — the lane does not push; integration is owned by the lead"
new_warnings_or_lints_introduced: none
cross_platform_build:
  windows_amd64_vet: pass
  linux_amd64_vet: pass
  host_build: pass
total_run_phase_files: 7
m1_to_mN_commit_strategy: "single milestone, single commit 410f6241d"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

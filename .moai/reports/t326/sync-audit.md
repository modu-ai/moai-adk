# SPEC-BINARY-LAG-VISIBILITY-001 — Sync-phase Independent Audit

- **Card**: t326 · **Tree**: `.claude/worktrees/t326` · **Branch**: `WT-binary-lag-visibility` · **HEAD**: `bfd09fcfc`
- **Audit scope**: sync-phase commits `d3454f1e6` · `f9c96c381` · `bfd09fcfc`, weighted toward what §E.4 asserts; run-phase (`c70c6aed9`, merged `ec15ec2cd`) re-tested but already independently audited at plan time.
- **Auditor**: sync-auditor (fresh judgment; no prior involvement in this card)
- **Verbatim evidence**: `.moai/reports/t326/sync-audit-evidence/` (`vet.txt`, `test-3pkg.txt`, `test-cli.txt`, `race-hook.txt`, `race-hook-full.txt`)

---

## Overall Verdict: **PASS-WITH-DEBT — 0.86** (weighted harmonic mean)

Must-pass firewall: **Functionality 0.95 PASS · Security 1.00 PASS** — both clear independently, so the firewall does not force FAIL.

**One blocking finding (F1) should land before this branch merges.** It is a five-line test fix, not a redesign; merging it converts that into a red `-race` gate the next lane inherits.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 0.95 | PASS (must-pass) | 17/17 AC tests PASS, none skipped, across `internal/binlag` · `internal/hook` · `internal/cli` (verbose runs below) |
| Security (25%) | 1.00 | PASS (must-pass) | argv-form `exec.CommandContext` (no shell), five lenient paths preserved, never promotes to `Fail`, buffered channel + `recover()` |
| Craft (20%) | 0.65 | FAIL-on-dimension | `go test -race ./internal/hook/` → **EXIT=1, 1 DATA RACE**, and it is this SPEC's own new test (F1) |
| Consistency (15%) | 0.80 | PASS | 3-phase close correct on all four artifacts; @MX commit mechanically comment-only; divergence scan incomplete (F2, F3) |

Weighted harmonic: `1 / (0.40/0.95 + 0.25/1.00 + 0.20/0.65 + 0.15/0.80)` = **0.8575 → 0.86**

---

## 1. Claim

Nine numbered assertions from §E.4 / §E.3 were tested rather than accepted:

1. AC-BLV-001..009 = 9 PASS / 0 FAIL
2. The @MX annotation commit changed no executable line
3. `internal/cli`'s 9 FAILs are wholly pre-existing; `regression_delta: 0`
4. The divergence scan found 7 real divergences and missed none
5. `claims_re_verified_as_accurate` (6 items) are in fact accurate
6. D2's disposition (record, don't fix) was the right call
7. The CHANGELOG entry is factually accurate against the code
8. The 3-phase close was performed across all four artifacts
9. `new_warnings_or_lints_introduced: 0`

---

## 2. Evidence

### 2.1 AC-BLV-001..009 — verified GREEN, and non-vacuously (claim 1: **UPHELD**)

I ran every AC test verbosely to rule out a silent skip — `TestBinaryLag_DoctorCheckNameSetIsUnchanged` carries a `t.Skipf` on an unavailable baseline blob and would have passed vacuously in a shallow clone. It ran.

```
$ go test -count=1 -v -run 'TestBinaryLag_|TestBuildIdentity_' ./internal/cli/
--- PASS: TestBinaryLag_OneSeamServesBothSurfaces (0.06s)        # AC-BLV-005
--- PASS: TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero (0.00s)  # AC-BLV-003(c)
--- PASS: TestBinaryLag_DoctorCheckNameSetIsUnchanged (0.04s)    # AC-BLV-009  (ran, not skipped)
--- PASS: TestBuildIdentity_VersionDerivationUnchanged (0.03s)   # AC-BLV-004(c)
--- PASS: TestBuildIdentity_IsMonotoneAcrossAnAncestorRelation (0.51s)  # AC-BLV-004(a)(b)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.316s

$ go test -count=1 -v -run 'TestSessionStart_Lag|TestSessionStart_NoLag|TestSessionStart_Blocking' ./internal/hook/
--- PASS: TestSessionStart_LagAdvisoryReachesAdditionalContext (0.07s)     # AC-BLV-001
--- PASS: TestSessionStart_NoLagAdvisoryWhenBinaryMatchesHead (0.06s)      # AC-BLV-002
--- PASS: TestSessionStart_LagAdvisorySerializesUnderAdditionalContext (0.06s)  # AC-BLV-008
--- PASS: TestSessionStart_BlockingComparerDoesNotStallSessionStart (0.31s)     # AC-BLV-006
ok  	github.com/modu-ai/moai-adk/internal/hook	1.100s

$ go test -count=1 -v ./internal/binlag/
--- PASS: TestEvaluate_BinaryAtHeadIsFresh / _AncestorBinaryIsBehind /
          _NonAncestorIsDivergentNotBehind / _NonGitDirectoryIsNotApplicable /
          _MissingCommitMetadataIsNotApplicable / _SubdirectoryStaysApplicable /
          _VersionStringDoesNotDecideTheVerdict / _DispatchesThroughTheComparerSeam
ok  	github.com/modu-ai/moai-adk/internal/binlag	2.954s
```

I additionally read each test against its AC's anti-vacuity clause rather than trusting the name:

- **AC-BLV-008** genuinely unmarshals **by key** (`session_start_binary_lag_test.go:107-117`), so both named mutants — advisory into `Data` (`types.go:394`, confirmed `json:"-"`) and into `SystemMessage` (`types.go:366`, confirmed serialized) — are actually caught. A whole-document substring search would not have.
- **AC-BLV-006**'s stub genuinely **ignores** `ctx` (`<-release`, line 142), and the test genuinely flips `deferredScansAsync` to `true` with restore (lines 138-150). Both are load-bearing: without the flip the inline branch runs with no bounded join at all, and with a cancellation-honouring stub a `context.WithTimeout` wrapper would pass too. The implementation's join is a real caller-side `time.NewTimer` + `select` (`session_start_binary_lag.go:73-82`).
- **AC-BLV-009**'s extraction is `go/ast` over the whole slice literal, keyed on the **entry's name expression** rather than quoted strings (`binary_lag_test.go:112-152`), so the constant-identifier evasion mutant is caught. `lagBaselineSHA = "22f90b1c7"` is a pinned SHA, not a moving ref.
- **AC-BLV-007**'s inversion is reproduced with the descendant carrying the *lower* semver (`binlag_test.go:143-159`).
- **AC-BLV-004** reads the `BUILD_ID` derivation **out of the Makefile** and executes it, rather than restating it — so it judges what the build does. Measured: `Makefile:6 VERSION ?= …--abbrev=0…` (byte-identical to baseline) and `Makefile:19 BUILD_ID := $(shell git describe --tags --dirty …)`.

Mutant record (§E.2) carries per-mutant verbatim RED including the three the CHANGELOG names — `Data`-map routing, `systemMessage` routing (two REDs), and the deadline-free join (`RED   panic: test timed out after 1m0s`).

### 2.2 The @MX commit changed no executable line (claim 2: **UPHELD — mechanically**)

```
$ git show d3454f1e6 -- '*.go' | grep -E "^[+-]" | grep -vE "^(\+\+\+|---)" \
    | grep -vE "^[+-]\s*//" | grep -vE "^[+-]\s*$"
(no output)
```

Every changed Go line in the sync commit is a comment or a blank line. This is the strongest available form of the claim and it holds.

### 2.3 The nine `internal/cli` FAILs and `regression_delta: 0` (claim 3: **UPHELD**)

```
$ go test -count=1 ./internal/cli/...   → EXIT=1
--- FAIL count: 9
TestRunDoctor_WithExport / _WithFix / _Verbose / _AllFlags / _VerboseAndDetail / _ExportMode
TestDoctorCmd_Execution / _ExportFlag / _VerboseExecution
FAIL	github.com/modu-ai/moai-adk/internal/cli	230.018s
(every other internal/cli sub-package: ok)
```

I counted the inherited reds myself rather than accepting the count. **Nine**, and the names match §E.4's stated split (TestRunDoctor_\* 6 + TestDoctorCmd_\* 3) exactly. Failure reason is uniform: `runDoctor error: doctor: 1 check(s) failed`.

Attribution measured independently:

```
$ go run ./cmd/moai doctor
  fail    Agent Emit Embed   no readable binary to judge at …/t326/bin/moai (11 committed artifacts to compare)
  fail    Harness 5-Layer    L1:FAIL L2:FAIL L3:FAIL L4:FAIL L5:FAIL L6:FAIL
  ok      Binary Freshness   development build (no commit metadata)
  Pass 24    Warn 2    Fail 2
$ ls bin/moai → No such file or directory
```

`Binary Freshness` — the only row this SPEC touches — **passes**. The two failing checks are environmental (no built binary in this worktree; harness layers) and unrelated to the lag seam. `regression_delta: 0` holds for the test axis.

### 2.4 The divergence scan (claim 4: **PARTLY UPHELD** — the 7 are real; the scan is incomplete)

All seven D1..D7 independently re-measured and **confirmed real**:

| # | Cited | Measured | Confirmed |
|---|---|---|---|
| D1 | `internal/cli/binary_lag.go` (§4 row 2) | implementation is `internal/binlag/binlag.go` | ✓ (and §4 vs §5 do contradict each other in one document) |
| D2 | binlag.go `:101` / `:111` | `121` = `rev-parse HEAD`, `131` = `merge-base --is-ancestor` | ✓ |
| D3 | session_start.go `:266`/`:277`/`:301`/`:574` | `maps.Copy` at **258, 276**; `json.Marshal(data)` **287**; `HookOutput{Data:` **311**; nothing at 574 | ✓ |
| D4 | append at `:343-346`/`:369` | advisory lands at **479** via `appendAdditionalContext` | ✓ |
| D5 | `doctor.go:140` doctorExitStatus | `func doctorExitStatus` at **142** | ✓ |
| D6 | `session_start.go:622-624` ctx-wrap | `context.WithTimeout(…, driftTimeout)` at **652** | ✓ |
| D7 | `:245-249` checkGroup bundling | `return []checkGroup{` at **254**, rows **255-257** | ✓ |

The `claims_re_verified_as_accurate` list — the half the dispatch flagged as more dangerous — was spot-checked on five of six items and **all hold**:

```
types.go:366  SystemMessage string `json:"systemMessage,omitempty"`     ✓ exact
types.go:394  Data json.RawMessage `json:"-"`                            ✓ exact
main_test.go:45-47  func TestMain → Lock → deferredScansAsync = false    ✓ exact
session_start_parallel_test.go:313-322  origAsync preserved → t.Cleanup  ✓ exact
SystemMessage writes: session_start.go 386-389, 437-440 only; 0 in session_start_binary_lag.go  ✓
Makefile VERSION line byte-identical to 22f90b1c7                        ✓ (test-enforced)
```

**But the scan missed a divergence inside its own declared scope** — see F2 — and its method excludes `plan.md` entirely, where a further set of citations diverges — see F3.

### 2.5 CHANGELOG factual accuracy (claim 7: **UPHELD**)

Spot-checked the entry's load-bearing, falsifiable claims:

```
"internal/hook/session_start.go:479"
  → sed -n '479p':  appendAdditionalContext(out, binaryLagAdvisory(ctx, lagRoot))   ✓ exact
"BUILD_ID … shown on the moai version stamp pill"
  → internal/cli/version.go:41-42  if version.BuildID != "" { stamp = version.BuildID }  ✓
"VERSION ?= line left byte-identical (verified against baseline 22f90b1c7)"          ✓ test-enforced
"run-phase commits c70c6aed9 → 1c042b93a → ee41c967c → 8ed269f46 → bf1a19813"
  → all five exist, in that order, on this branch                                     ✓
"9 PASS 0 FAIL, 13 mutants planted-RED-reverted"                                      ✓ §E.2 carries verbatim RED per mutant
"AC-BLV-009 … no new check name across all three registries"                          ✓ AST test passes
```

The entry is unusually long but I found no factual overstatement in it. Notably it does **not** claim a sync audit was run — §E.4 states plainly `sync_audit_verdict: "NOT RUN"`, which is honest and correct at the time it was written.

### 2.6 3-phase close (claim 8: **UPHELD**)

```
spec.md        status: completed  updated: 2026-08-28  version: 0.4.1
plan.md        status: completed  updated: 2026-08-28  version: 0.4.1
acceptance.md  status: completed  updated: 2026-08-28  version: 0.4.0
progress.md    status: completed  updated: 2026-08-28  version: 0.4.0
```

All four moved. `sync_commit_sha: d3454f1e6` was backfilled in `f9c96c381` as stated — a commit cannot carry its own hash, so the two-commit shape is correct rather than sloppy.

### 2.7 Toolchain gates

```
$ go vet ./internal/binlag/... ./internal/cli/... ./internal/hook/... ./pkg/version/...   → EXIT=0
$ go test ./internal/binlag/... ./internal/hook/... ./pkg/version/...                     → EXIT=0
$ gofmt -l  (binlag, doctor.go, binary_lag_test.go, session_start_binary_lag{,_test}.go, pkg/version)
  (empty — clean)
```

### 2.8 Security read

`gitCompare` shells out twice via `exec.CommandContext` in **argv form** — no shell, so no injection surface even though `BinaryCommit` is an ldflags-injected string and `Dir` comes from `ProjectDir`/`CWD`. All five lenient paths return `StatusNotApplicable`/`StatusDivergent` and `checkBinaryFreshness` maps every non-`Behind` status to `CheckOK` and `Behind` to `CheckWarn` — never `CheckFail` — so `doctorExitStatus` cannot be promoted by this row, which is the actual downstream-availability hazard. The advisory goroutine sends on a **buffered** channel (cap 1) so it cannot leak blocked after abandonment, and it recovers from panics. The comparison receives the handler's `ctx`, so an abandoned `git` subprocess is reaped on handler-context cancellation. No secrets, no writes, no network.

---

## 3. Baseline-attribution

Every measurement in §2 was taken **in this run, against this tree**: worktree `.claude/worktrees/t326`, branch `WT-binary-lag-visibility`, HEAD `bfd09fcfc` (re-read at audit start via `git rev-parse --short HEAD`). No figure is carried over from §E.3, §E.4, the lane verdict, or either plan audit — where a number matched theirs, it matched because I re-derived it.

- Test/vet/race counts: commands quoted verbatim above, exit codes captured, raw output persisted under `.moai/reports/t326/sync-audit-evidence/` (not `/tmp`).
- Line-number remeasurements: `sed -n` / `grep -n` against the working tree at `bfd09fcfc`.
- AC-BLV-009 baseline: `22f90b1c7`, a pinned SHA, read through `git show 22f90b1c7:internal/cli/doctor.go` by the test itself.
- `go test -race` baseline for attribution: full-package run, `internal/hook`, this tree.

---

## 4. Findings

### F1 — [HIGH] [blocking] `binlag.Comparer` is an unsynchronized seam; `go test -race ./internal/hook/` now fails

`internal/binlag/binlag.go:81` · `internal/hook/session_start_binary_lag_test.go:141-150`

```
$ go test -race -count=1 ./internal/hook/     → EXIT=1
WARNING: DATA RACE  (count: 1 in the whole package)
Write at 0x0001076bada0 by goroutine 66:
  …TestSessionStart_BlockingComparerDoesNotStallSessionStart.func3()
      internal/hook/session_start_binary_lag_test.go:146
  testing.(*common).Cleanup.func1()
Previous read at 0x0001076bada0 by goroutine 80:
  …binlag.Evaluate()               internal/binlag/binlag.go:91
  …hook.binaryLagAdvisory.func1()  internal/hook/session_start_binary_lag.go:70
--- FAIL: TestSessionStart_BlockingComparerDoesNotStallSessionStart (0.31s)
    testing.go:1712: race detected during execution of test
FAIL	github.com/modu-ai/moai-adk/internal/hook	29.831s
```

Attribution is unambiguous: **exactly one** data race exists in the entire `internal/hook` package and it is this SPEC's own new test. Every other test in the package is race-clean, so this is a race the SPEC introduced, not one it inherited. `internal/binlag` alone is race-clean (`EXIT=0`).

Mechanism: the test's cleanup writes `binlag.Comparer` while the deliberately-abandoned advisory goroutine is still reading it through `Evaluate`. The abandonment is by design (that is what AC-BLV-006 is testing), but nothing joins the goroutine before the seam is restored.

**Production is not affected** — `grep -rn "Comparer\s*=" --include='*.go' internal pkg | grep -v _test.go` returns only the declaration at `binlag.go:81`, so the var is never written outside tests. This is a test-hygiene defect, not a shipped one.

It is nonetheless blocking, for two reasons. First, `CLAUDE.local.md` §6 prescribes `go test -race` for any code touching goroutines or channels, and this SPEC adds a goroutine — so the gate the project's own rule names was never run, at run phase or at sync. Second, merging it hands the next lane an inherited red on a standard verification command, which is a failure mode this project has been bitten by repeatedly and which costs far more to diagnose later than to fix now.

**Required fix**: in `TestSessionStart_BlockingComparerDoesNotStallSessionStart`, release and **join** the stub goroutine before restoring `binlag.Comparer` (register the restore cleanup so it runs after `close(release)` and after the goroutine has been observed to finish — e.g. a `done` channel the stub closes on exit, waited on in the cleanup). Alternatively route the seam through a mutex-guarded accessor in `binlag`. Then re-run `go test -race ./internal/hook/` and record `EXIT=0`.

### F2 — [MEDIUM] [blocking] The divergence scan missed a divergence inside its own declared scope, and recorded the measurement as a confirmation

`acceptance.md` AC-BLV-009 「판정 근거 위치」 · `progress.md` §E.4 `D7.note`

acceptance.md cites the three registries at `systemChecks(:187)`, `moaiChecks(:195` decl `~ :212` close, items `:196-211)`, `workspaceChecks(:214)`. Measured on this tree:

```
$ grep -n "systemChecks\|moaiChecks\|workspaceChecks" internal/cli/doctor.go
189:	systemChecks := []checkFunc{
197:	moaiChecks := []checkFunc{
220:	workspaceChecks := []checkFunc{
$ sed -n '210,222p' internal/cli/doctor.go   → moaiChecks closes at 218; items span 198-217
```

So the cited addresses are off by 2 / 2 / 6, and the item range `196-211` misses four entries — the very defect AC-BLV-009's own `[HARD]` note warns about ("행 구간으로 추출하면 이 기준이 잡으려는 바로 그 뮤턴트가 창 밖에 놓인다"). This is a divergence of exactly the D5/D7 class and belongs in the list; `divergences_found: 7` should be ≥ 8.

What makes this more than an omission: D7's `note` **measured the correct numbers** — it writes "세 레지스트리 선언(189 systemChecks / 197 moaiChecks / 220 workspaceChecks)도 실재한다" — and then framed that measurement as a confirmation of accuracy rather than as a mismatch against the 187/195/214 the same AC cites. The existence statement is true; the accuracy implication it carries is not. That is the "false accurate" shape the dispatch flagged as the dangerous half, occurring inside the scan's own text.

Mitigating: the AC's *judgment* is unaffected — the AST test reads slice literals by identifier, not by line, so it is immune to these addresses being wrong. The defect is in the record, not the guard.

**Required fix**: add D8 to the divergence list with the measured values above, and reword D7's note so the registry-declaration sentence states the mismatch instead of implying agreement.

### F3 — [MEDIUM] [optional] The scan's method excludes `plan.md`, which carries a further set of divergent citations

`progress.md` §E.4 `spec_body_vs_code_divergence_scan.method`

The stated method covers "spec.md §3·§4·§5·§6, acceptance.md AC-BLV-001..009". `plan.md` is not in it, yet it is a SPEC body artifact that closed to `completed` in the same commit and carries its own code citations — several of which appear nowhere in spec.md or acceptance.md, so nothing covers them:

```
plan.md:25,144  doctor.go:495  checkBinaryFreshness      → measured 518
plan.md:33,148  session_start.go:593 computeDeferredAdvisory → measured 623
plan.md:150     session_start.go:305 AdditionalContext    → measured 315
plan.md:93,152  Makefile:14 RELEASE_BINARY               → measured 25
plan.md:93,152  Makefile:35 version.json                  → measured 66
```

plus plan.md restatements of citations already recorded for the other artifacts (`:266`/`:574`/`:277`/`:301`, `:245-249`, `:622-624`, `doctor.go:140`, `:343-346`/`:369`, `:187`/`:195`/`:212`/`:214`).

This is a scope gap rather than a false statement — the method section is honest about what it read. But the mandate above it reads broader ("남은 본문 주장과 착지한 코드의 어긋남을 능동적으로 훑는다"), and a reader would reasonably take the scan as covering the SPEC's body artifacts. Optional because, as with F2, none of these addresses is load-bearing for any AC judgment.

### F4 — [LOW] [optional] D2's disposition was correct; the reasoning generalizes one step too far

`progress.md` §E.4 `D2`, `disposition`

I affirm the call. Recording rather than reconciling is right: `manager-docs` does not own spec.md body content, and an unverified reconciliation is worse than a recorded Gap — that is a sound principle, not a dodge. The record is also unusually honest, self-attributing D2 to its own commit ("**내가 만든 어긋남**이다") rather than laundering it into "authoring".

One qualification. D2 is the one divergence in the set that is purely mechanical and self-inflicted: this sync commit inserted 20 comment lines above two citations and shifted them 101→121 / 111→131, with zero semantic content on either side. Correcting a line number that the same commit invalidated is arguably not a "body content edit" in the sense the ownership rule protects — it is restoring an address the commit itself broke. The blanket disposition sweeps D2 in with D1 (a real path contradiction, correctly out of scope) when the two are different in kind.

Debt, not defect. No action required if the lead prefers a single manager-spec re-delegation covering D1..D8 together.

### F5 — [LOW] [optional] Two failure counts from two environments are reported as one

`progress.md` §E.4 `verification.test_cli_attribution`

The attribution paragraph reads `go run ./cmd/moai doctor` (which I reproduce as **2 fails**: Agent Emit Embed, Harness 5-Layer) alongside the test failures (whose message is `doctor: 1 check(s) failed`). Both measurements are real and I reproduced both, but they come from different environments — the tests run against their own temp roots — and the record does not say so. A later reader reconciling "2" against "1" will spend time on a discrepancy that is not one.

### F6 — [INFO] Residual risks are accurately stated

The three `residual_risk` entries (Binary Freshness sits inside an already-red suite so only its dedicated test guards it; the 250 ms silent drop can read as "no lag"; `moai version` title line still reads the tag floor) are all correct and I independently confirmed the first two. No action.

---

## 5. Gaps — what I did NOT observe

- **Cross-platform.** All measurement is darwin/arm64. I ran no `GOOS=windows` or `GOOS=linux` build, vet, or test. §E.3's `cross_platform_build: partial` is honest and I did not narrow it.
- **Full repository suite.** Scoped to `./internal/binlag/... ./internal/cli/... ./internal/hook/... ./pkg/version/...` per the lane-local rule. The repository-wide verdict belongs to CI on the integration branch and is **PENDING at report time**. I make no claim about it.
- **`-race` outside `internal/hook` and `internal/binlag`.** I did not run `go test -race ./internal/cli/...`; the 230 s uninstrumented run makes an instrumented one costly, and F1's attribution did not require it. Whether `internal/cli` is race-clean is unmeasured here.
- **Whether the 9 `internal/cli` FAILs reproduce at baseline `22f90b1c7`.** I established attribution by naming the failing checks and confirming `Binary Freshness` passes — not by checking out the baseline and re-running. A baseline run would be the stronger form and I did not do it.
- **Real `moai doctor` from outside a repository.** §E.3 already records this as unobserved (the worktree guard refuses the cwd change). I did not close it either; I judged the same branch through `TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero` and `TestEvaluate_NonGitDirectoryIsNotApplicable`, both of which use `t.TempDir()`.
- **`make build` / a rebuilt binary.** Not run. `bin/moai` does not exist in this worktree, which is why two doctor rows fail here. The BUILD_ID ldflags path is judged only through the Makefile-derivation test, not through an actual built artifact in this tree.
- **`moai mx query` harvest of the new @MX tags.** §E.4 records this as unobserved and I did not close it — I verified the annotations are comments that change no executable line, not that a scanner picks them up.
- **CHANGELOG prose at sentence granularity.** I checked its falsifiable claims (paths, line numbers, commit SHAs, counts). I did not adjudicate every narrative sentence; no mechanical gate exists for that.
- **The sixth `claims_re_verified_as_accurate` item** ("§4 표 행 1·4·5·6 … 전부 실재") was verified only partially — I confirmed `doctor.go` rewiring, `Makefile BUILD_ID`, and both test files exist, but did not walk the §4 table row by row.

---

## 6. Residual-risk

- **F1's fix has a subtle correctness trap.** The obvious repair — moving the `close(release)` cleanup — interacts with `t.Cleanup`'s LIFO ordering. A fix that releases the goroutine but does not *join* it leaves the same race with a narrower window, which will read as fixed on a few runs and reappear intermittently in CI. The fix should be verified by a repeated instrumented run (`go test -race -count=5 -run TestSessionStart_Blocking ./internal/hook/`), not a single pass.
- **The `Binary Freshness` row's only guard is its dedicated test.** Nine `internal/cli` doctor tests are red for unrelated environmental reasons, so a future regression in this row will not show up as a suite-level change — the suite is already red. This is correctly stated in §E.4's residual risk and remains true after my audit.
- **The divergence record is now itself slightly stale.** D2's own reasoning applies recursively: any later commit inserting lines into these files shifts the remaining citations again. The addresses in D1..D7 (and F2's D8) are pinned to `f9c96c381`/`bfd09fcfc` and will drift.
- **My AC-BLV-009 confidence rests on the AST extraction being correct.** I read `checkNamesFromSource` and it looks right (whole slice literal, entry name expression, non-empty guard that fails rather than passing on zero extractions), but I did not mutate the extractor itself to confirm it reports a delta when one exists — the run phase's three recorded mutants did that, and I accepted their record rather than re-planting.
- **`sync_audit_verdict` in §E.4 currently reads `"NOT RUN"`.** That was accurate when written and is now superseded by this report. If it is not updated, a later reader will take the SPEC as closed without an independent quality judgment.

---

## 7. Recommendations

1. **Fix F1 before merge** — join the stub goroutine before restoring `binlag.Comparer`; verify with `go test -race -count=5 -run TestSessionStart_Blocking ./internal/hook/` → EXIT=0, and record the verbatim output. This is the one item I would hold the branch on.
2. **Update §E.4** — add D8 (F2), reword D7's registry note so the measurement reads as a mismatch, and either widen the scan's `method` to `plan.md` or state explicitly that plan.md was out of scope (F3).
3. **Record this audit's verdict** in §E.4's `sync_audit_verdict` field, replacing `"NOT RUN"`.
4. **Optional, lead's call**: a single `manager-spec` re-delegation covering D1..D8 as one batch, rather than per-divergence repair. D1 (the §4-vs-§5 path contradiction) is the only one with semantic content and is the item most worth an author's attention.
5. **No action** on F4/F5/F6 unless the lead wants the two-environment failure counts disambiguated in the record.

---

**Auditor's note on stance**: this card's premise was disproved twice before implementation, so I treated the sync signal's assertions as hypotheses throughout. The record held up better than that prior predicted — the 7 divergences are all real, the `claims_re_verified_as_accurate` list survived spot-checking, the @MX comment-only claim is mechanically true, and the AC tests are non-vacuous in the specific ways their anti-vacuity clauses demand. The two things the record got wrong are both in the same place: it did not look at `plan.md`, and it recorded one measurement it took as a confirmation rather than as the mismatch it was.

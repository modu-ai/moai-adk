---
id: SPEC-TEMPDIR-CLEANUP-RACE-001
title: "Deferred session-start writes must not outlive Handle for a cross-package caller — a deliberate synchronous-scan seam plus a regression guard for the t.TempDir cleanup race"
version: "0.1.3"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: manager-spec (card t352)
priority: P1
phase: "v3.2.0"
module: "internal/hook, internal/cli"
lifecycle: spec-anchored
tags: "test-isolation, tempdir-cleanup, session-start, deferred-scan, race, regression-guard, flake"
tier: S
related_specs:
  - SPEC-CLI-TEST-CWD-ISOLATION-001
  - SPEC-V3R6-HOOK-ASYNC-EXPAND-001
  - SPEC-BINARY-LAG-VISIBILITY-001
---

# SPEC-TEMPDIR-CLEANUP-RACE-001 — deferred writes must not outlive `Handle`

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-28 | manager-spec (card t352) | Initial draft. Scope bound to the one reproduced observation in `.moai/reports/t352/reproduction.md`; the second observation stays an open gap. |
| 0.1.1 | 2026-08-28 | manager-spec (card t352) | Plan-audit repairs D1-D8. Corrected the caller inventory (a production caller exists at `internal/cli/deps.go:221`, and the constructor is non-variadic); recorded the base SHA the diff-shape AC judges against; withdrew the misattributed 1-in-5 CI frequency; restated the fixture rationale in terms of `Handle`'s return; recorded the Tier S budget overrun explicitly (nine ACs and a separate `acceptance.md`, both retained — no criterion folded or deleted; justification in `plan.md` §D.4); stated the CI-runtime headroom constraint; stated why the entry-set comparison generalizes REQ-TCR-001; split the compound REQ-TCR-002. |
| 0.1.2 | 2026-08-28 | manager-spec (card t352) | Plan-audit iteration 2 repairs D9-D10. Removed the unattributed "slowest package" superlative from the §D CI-headroom constraint, replacing it with the repo's own `ci.yml:232` figure cited as an unverified carry; paired AC-TCR-002b(ii)'s emptiness check with a mandatory non-empty positive control so a broken pathspec cannot pass as an unchanged file. |
| 0.1.3 | 2026-08-28 | manager-spec (card t352) | §F factual addition only — no scope, requirement, or design change; observation 1 stays excluded and unreproduced. Recorded that observation 1 is the sole failing test on the latest measured `develop` head (CI run `33173944485`, head `c6aa61346`), that it manifests under the race detector only (`Race Test` failed while `Test (ubuntu-latest)` passed on that head), that the earlier 1-in-5 frequency figure is superseded with no new rate claimed, and — as a Gap — that `-race` on linux/amd64 is the untried lever, since this lane's `-race` run passed on linux/aarch64. |

---

## §A Context

### A.1 Evidence base

The single evidence base for this SPEC is **`.moai/reports/t352/reproduction.md`** (lane t352,
worktree `.claude/worktrees/t352`, base `origin/develop` @ `77b2bcae6`). Every statement below
traces to a measurement recorded there. Its Gaps are carried forward into §F rather than restated
as solved.

### A.2 The established mechanism

`hook.(*sessionStartHandler).Handle` dispatches a durable write into the **caller's** directory
after the join that releases it:

- `internal/hook/session_start.go:240` takes the async branch when `deferredScansAsyncEnabled()`,
  spawning `spawnDeferredAdvisoryScans` and joining it with `deferredScanJoinBound` (250 ms).
- `internal/hook/session_start.go:608-611` — inside that goroutine the advisory is sent into the
  buffered channel **first**, and `runMXColdStartScan(projectDir)` runs only afterwards. The
  ordering is deliberate and documented in the source ("dispatched AFTER the advisory result is
  sent so it never delays advisory keys").
- `internal/hook/session_start.go:1606-1613` — `runMXColdStartScan` writes
  `<projectDir>/.moai/state/mx-index.json`. `mxIndexNeedsRebuild` returns `true` for any fresh
  directory, so the write is always attempted.

The join therefore releases `Handle` while a durable write into the caller's directory is still
ahead of it.

`runMXColdStartScan` is the only durable writer this reproduction established, but REQ-TCR-001 is
deliberately stated over **every** durable write, and the guard matches that scope by construction:
it compares the caller directory's **whole entry set** before and after, so it covers any writer,
not only the MX index. The one other post-`Handle` goroutine known today, `binaryLagAdvisory`,
performs only `git rev-parse` and `git merge-base --is-ancestor`
(`internal/binlag/binlag.go:101,111`) and writes nothing, so it is outside REQ-TCR-001 — a verified
premise, not an assumption.

### A.3 Why the existing protection does not reach the caller

`internal/hook/main_test.go:45-50` flips the package-private `deferredScansAsync` to `false` for
the `internal/hook` test binary, which removes the goroutine entirely. That protection is
package-private and does not cross a test-binary boundary: a test in `internal/cli` runs a
different binary and observes the production value `true`.

**Cross-package caller inventory — two callers, one of them production:**

- `internal/cli/binary_lag_test.go:57` is the single cross-package **test** caller
  (`grep -rn 'hook\.NewSessionStartHandler' --include='*_test.go' internal/ | grep -v '/hook/'`
  returns that line plus the throwaway probe). It passes a `t.TempDir()` as `ProjectDir`, so the
  late write races Go's `t.TempDir` cleanup `RemoveAll`, producing the observed CI failure
  `TempDir RemoveAll cleanup: unlinkat /tmp/TestBinaryLag_OneSeamServesBothSurfaces.../001/.moai/state: directory not empty`.
- `internal/cli/deps.go:221` is a cross-package **production** caller
  (`deps.HookRegistry.Register(hook.NewSessionStartHandler(deps.Config))`). It does not exhibit the
  race — the production hook process exits rather than deleting its project directory — but it
  constrains the fix: `internal/hook/session_start.go:41` declares
  `func NewSessionStartHandler(cfg ConfigProvider) Handler`, which is **non-variadic**, so a seam
  added as a second positional argument would break this line. See the §D constraint.

### A.4 Why it is intermittent, and what is not known about its frequency

The write's lateness scales with scan cost (measured: `nFiles=200` → 11 ms late, `nFiles=2000` →
43 ms, `nFiles=8000` → 223 ms; `nFiles=0` sits on the join-budget boundary and flipped between two
bases). A fast local disk supplies neither a slow cleanup nor a slow scan, which is why 50
iterations of the symptom test passed locally.

**The CI frequency of this observation is unknown.** `.moai/reports/t322/verdict.md:282` records a
single appearance of `TestBinaryLag_OneSeamServesBothSurfaces` ("appeared", count 1). One
appearance is not a rate: nothing measured establishes how often the race bites in CI, and
"intermittent" here is an inference from the mechanism rather than an observed frequency. The
1-in-5 figure that appears in the t322 verdict belongs to `TestGitDiffNameCount_Predicate`
(`t322/verdict.md:281`) — observation 1, which §C excludes — and is not this SPEC's.

---

## §B Requirements (GEARS)

- **REQ-TCR-001 (capability gate).** **Where** a caller requests synchronous deferred scans, the
  session-start handler shall complete every durable write into that caller's `ProjectDir` before
  `Handle` returns.

- **REQ-TCR-002 (unwanted).** Absent an explicit caller request, the session-start handler shall
  not alter its deferred-advisory-scanning behaviour: the scan shall remain asynchronous and shall
  remain joined with the existing bounded deadline.

- **REQ-TCR-003 (event-driven).** **When** a test outside the `internal/hook` package invokes
  `Handle` against a directory whose lifetime that test owns, the test shall request synchronous
  deferred scans.

- **REQ-TCR-004 (ubiquitous, regression guard).** The repository shall carry a guard test that
  fails when a durable write into the caller's directory lands after `Handle` has returned, and
  that guard shall have been observed red on a tree carrying the defect.

- **REQ-TCR-005 (state-driven).** **While** the `internal/hook` test binary runs, the existing
  package-private inline seam shall retain its current behaviour, and the package shall continue to
  report zero goroutine leaks.

- **REQ-TCR-006 (unwanted).** The production default of the existing execution-mode seam
  (`deferredScansAsync`) shall not change, and the existing production call site
  `internal/cli/deps.go:221` shall not require modification.

---

## §C Exclusions

This SPEC deliberately builds nothing beyond the mechanism established in §A.2. Each exclusion
below is stated because it has an active pull toward inclusion.

### Out of Scope — observation 1 (`TestGitDiffNameCount_Predicate` / `internal/graph`)

- The `.git/objects: directory not empty` flake in `internal/graph` is **not** addressed here.
- It was **not reproduced**: 10 iterations across macOS and a Linux aarch64 container recorded zero
  post-command writers, and the whole package is clean under `-race`.
- The card's "one class, two instances" reading is supported for one instance only. This SPEC
  neither claims to fix observation 1 nor asserts that the two observations share a cause.
- Consequence, stated plainly: landing this work leaves observation 1's flake in `develop`.

### Out of Scope — production session-start latency behaviour

- The async deferred path and its 250 ms join bound exist deliberately (the input-lag win,
  documented in the source at `session_start.go:196-215`).
- Any change making the production `Handle` wait for the MX cold-start scan is a regression of that
  design and is rejected. The seam introduced here is opt-in and off by default.

### Out of Scope — a repository-wide `t.TempDir` race sweep

- No sweep for other `t.TempDir` cleanup races is performed. Two observations, one of them
  unreproduced, do not establish a class.
- The blast radius addressed here is the single measured **test** call site,
  `internal/cli/binary_lag_test.go:57`. The production call site `internal/cli/deps.go:221` is in
  scope only as a compile constraint (§D), not as a behaviour change.

### Out of Scope — the throwaway measurement probe

- `internal/cli/zz_t352_probe_test.go` (retained in the lane worktree for the run phase) is a
  measurement instrument, not a deliverable. Whether it survives is a run-phase decision, not a
  requirement here.

---

## §D Constraints

- The change is expected to be small: one seam in `internal/hook`, one call site in
  `internal/cli/binary_lag_test.go`, one guard.
- **The seam is variadic.** The exported constructor is `func NewSessionStartHandler(cfg
  ConfigProvider) Handler` (`internal/hook/session_start.go:41`) — non-variadic. The option MUST be
  introduced as a variadic option parameter so that the production call site
  `internal/cli/deps.go:221` keeps compiling **unchanged**. A diff touching `deps.go` means the
  seam was added in the wrong shape (mechanically checked by AC-TCR-002b; compile-checked by
  AC-TCR-006 and AC-TCR-007).
- Cross-platform: the seam and the guard must build and run on linux, darwin, and windows.
- `internal/cli` carries a `.moai`-residue guard in its `TestMain`
  (`internal/cli/main_test.go:213-245`). The new guard must not trip it and must not rely on it —
  that guard watches the package working directory, a different locus from the caller-owned
  temporary directory this SPEC concerns.
- **CI runtime headroom.** CI runs `go test -race -count=1 ./...` with no `-timeout` flag
  (`.github/workflows/ci.yml:238`), i.e. the 10-minute per-package default. This SPEC measured no
  per-package durations — the reproduction records "`internal/cli`'s full package was not run" as a
  Gap (§F) — so no claim is made here about which package is slowest. What is on record is a repo
  comment: `.github/workflows/ci.yml:232` asserts that `internal/cli`'s `exec.Command` paths
  "dominate race runtime (`internal/cli ~379s/70%`)". That figure is an **unverified carry** — it
  was authored elsewhere and not re-measured in this card — and it is cited only to say that the
  headroom question is worth asking, not to quantify the headroom. The guard's padded fixture plus
  its `Handle` call must add well under whatever headroom the package actually has. The reproduction's whole four-configuration
  probe — the 8000-file row included — completed in 1.61 s wall, so the expected addition is
  seconds, not minutes; AC-TCR-007 records the package's wall-clock so the cost is measured rather
  than assumed.

---

## §E Success criteria

Enumerated as binary-testable acceptance criteria in `acceptance.md`. Every criterion names the
command that decides it, and AC-TCR-002b additionally records the base SHA it judged against.

**Tier S budget overrun, recorded deliberately.** This SPEC carries nine acceptance criteria against
the Tier S ceiling of 8 (`spec-workflow.md:148`) and a separate `acceptance.md` against the Tier S
2-file set (`spec-workflow.md:140`). Both overruns are stated rather than absorbed, and the
justification — why each retained criterion earns its place, and why the implementation scope is
nevertheless genuinely Tier S — is in `plan.md` §D.4. No criterion was deleted or folded to make the
count fit: removing a check to satisfy a budget shrinks the coverage the criteria provide, which is
the wrong direction.

---

## §F Risks (carried from the evidence base)

Carried forward from `.moai/reports/t352/reproduction.md` § Gaps and § Residual risk. These remain
**open**; nothing in this SPEC closes them.

- **CI itself was not re-run.** All reproduction measurements are local. The CI figures in the card
  are carried from `.moai/reports/t322/verdict.md` and were not re-measured.
- **Observation 2's CI frequency is unknown, not low.** `t322/verdict.md:282` records one
  appearance of `TestBinaryLag_OneSeamServesBothSurfaces`; one appearance is not a rate, so the cost
  of leaving the race unfixed is unquantified. The 1-in-5 figure in that same report belongs to
  `TestGitDiffNameCount_Predicate` (`t322/verdict.md:281`) — observation 1, excluded by §C — and is
  not this SPEC's (§A.4).
- **The probe measures the write, not the collision.** It establishes that the write lands after
  `Handle` returns; it does not itself produce the `unlinkat ... directory not empty` string, so the
  final link from mechanism to the observed CI failure remains an inference — well supported, but an
  inference. A guard built on the write, not the collision, inherits that inference.
- **Observation 1's cause is unestablished.** Ruled out: an in-process goroutine, a parallel test in
  the package, and a background writer detectable within 750 ms on macOS or Linux aarch64. Not ruled
  out: an x86-64-specific or runner-specific writer, a writer appearing later than 750 ms, or a
  cleanup interaction that is not a writer at all.

- **The excluded observation is, on the latest measured head, the SOLE failing test on `develop` —
  so landing this SPEC does not turn `develop` green.** §C already states that observation 1's flake
  stays; this is the sharper form: nothing else is red alongside it, so a reader must not infer a
  green `develop` from this card landing. Evidence — CI run `33173944485`, head
  `c6aa613463e6234155f45ce76666e985a42cd80c` (`c6aa61346`) on `develop`, status `completed`,
  conclusion `failure`; log `.moai/state/verify/lead-ci/c6aa61346-failed.log:51-52`:

  ```
  --- FAIL: TestGitDiffNameCount_Predicate (0.14s)
      testing.go:1464: TempDir RemoveAll cleanup: unlinkat /tmp/TestGitDiffNameCount_Predicate2957256880/001/.git/objects: directory not empty
  ```

  `grep -c 'FAIL:'` over that log returns `1` — line 51 is its only failing-test line, and the
  failing package is `internal/graph`. (Carried, not verified here: the lead reports the baseline
  `5e194bba2` carried 15 failing tests of which 14 are now gone. That comparison was not
  re-measured in this card.)

- **Observation 1 manifests under the race detector only — the sharpest discriminant recorded so
  far.** On that same head the `Race Test` job failed while `Test (ubuntu-latest)` **passed**:
  `gh run view 33173944485 --json jobs` reports `failure Race Test` and `success Test
  (ubuntu-latest)`, with every `Build (…)`, `Lint`, and `Integration Tests (…)` job `success`. A
  reproduction attempt that omits `-race` is therefore expected to pass and to prove nothing.

- **The earlier frequency figure is superseded, and no new rate is claimed.** `t322/verdict.md:281`'s
  "1 of 5 post-landing runs" predates this appearance; observation 1 has now been seen on a third,
  later head. Three scattered observations across different heads do not make a rate — computing one
  would be inventing a denominator — so this SPEC records the occurrences and claims no frequency
  for observation 1, exactly as §A.4 already declines to claim one for observation 2.

- **Untried lever for a future observation-1 card: `-race` on linux/amd64.** This lane's Linux
  attempt ran `go test ./internal/graph/... -race -count=5` on **linux/aarch64** and passed, while
  the CI failure is linux/**amd64**. So "run it under the race detector" is demonstrably *not*
  sufficient on its own to reproduce it — the untried combination is the race detector **and** the
  amd64 architecture together. Recorded as an untried lever, not as a hypothesis this SPEC endorses:
  nothing here establishes that architecture is the operative variable rather than runner load,
  filesystem, or something not yet named.
- **The container measured was aarch64; CI is x86-64.** An architecture-specific writer is not
  excluded for observation 1.
- **The graph-probe figures were taken on `e08d5e55c`, not `77b2bcae6`**, and the 22 intervening
  commits were not examined for `internal/graph` fixture changes.
- **`internal/cli`'s full package was not run** during reproduction; only two `-run` selectors
  executed.
- **The mechanism may have further callers.** `Handle`'s deferred writer is fire-and-forget by
  design because the production hook process exits. Any future in-process caller that owns and then
  deletes its project directory inherits the same race. The inventory in §A.3 covers the callers
  present today (one test, one production).
- **The `nFiles=0` row flipped between the two bases**, so the empty-directory case sits on the
  join-budget boundary and is machine- and load-dependent. A guard whose fixture is an empty
  directory would itself be flaky; the fixture must be padded.

---

## §G Cross-references

- `.moai/reports/t352/reproduction.md` — the evidence base (mandatory reading before run phase).
- `.moai/reports/t322/verdict.md` — the origin of the CI figures, not re-measured here; note the
  per-observation attribution at `:281` (observation 1) vs `:282` (observation 2).
- `.moai/reports/t352/plan-audit.md` — the plan audit whose defects D1-D8 version 0.1.1 repairs.
- `internal/hook/session_start.go` — `Handle`, `spawnDeferredAdvisoryScans`, `runMXColdStartScan`,
  `deferredScansAsync`, and the non-variadic constructor at `:41`.
- `internal/hook/main_test.go` — the existing package-private inline seam.
- `internal/cli/binary_lag_test.go` — the cross-package test caller.
- `internal/cli/deps.go` — the cross-package production caller (compile constraint, §D).
- `internal/cli/main_test.go` — the sibling `.moai`-residue guard (related family, different locus).
- `.claude/rules/moai/development/verification-completeness.md` — the observed-failure completion
  axis and the two-cell adoption discipline AC-TCR-004 follows.

# sync-audit — SPEC-INTEGRATION-LOCK-ATOMIC-001 (card t336)

- Tree: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t336`
- Branch `WT-integration-lock-atomic`, HEAD `3c7886d9a`, base `a5f414a8a`
- Auditor: sync-auditor (independent). Every figure below was measured in this run, against this tree.

## Overall verdict: **PASS** (harmonic mean 0.93)

The card's central claim — *the integration-lock record's read-modify-write is now serialized across
OS processes, proven in both directions* — is **supported by evidence I observed myself**. I ran the
GREEN criterion, then disabled the critical section with the documented one-line revert, observed the
attributed double-hold on the first attempt, restored the line, and proved the restoration with an
empty diff and a clean `git status --short`.

### Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim, this run) |
|---|---|---|---|
| Functionality (40%) | 95/100 | PASS | `go test ./internal/kanban/... -count=1` → `ok  github.com/modu-ai/moai-adk/internal/kanban	14.227s` (exit 0). 8 of 9 AC reproduced independently; AC-ILA-007(b) is Windows-runtime and unobservable here. |
| Security (25%) | 90/100 | PASS | `golangci-lint run --timeout=3m ./internal/kanban/...` → `0 issues.` Record mode asserted 0644 by `TestWriteIntegrationLock_UniqueStagingPath` (`--- PASS`); staging name randomized by `os.CreateTemp`; every uncertainty in the clear path resolves toward treat-as-live. |
| Craft (20%) | 92/100 | PASS | `go test -cover ./internal/kanban/... -count=1` → `ok ... 14.560s	coverage: 86.5% of statements` (threshold 85%). |
| Consistency (15%) | 95/100 | PASS | `git diff --stat a5f414a8a -- internal/hook internal/config` → empty; `git diff --stat a5f414a8a -- internal/cli` → empty; `git diff --stat 15453140a -- internal/cli/integration.go` → empty. |

Must-pass firewall: Functionality and Security each clear their thresholds independently. No blocking finding.
Harmonic mean of (95, 90, 92, 95) = 92.95.

---

## The five concentration points

### 1. Is the GREEN real, or manufactured by the interleaving hook? — **Real**

**(a) The hook is unreachable in production.** Declared `var integrationLockMutationTestHook func()`
with no initializer (nil), unexported, and assigned in exactly one place:

```
$ grep -rn "integrationLockMutationTestHook" internal/
internal/kanban/integration_lock.go:83:// integrationLockMutationTestHook is a nil-by-default, TEST-ONLY interleaving
internal/kanban/integration_lock.go:95:var integrationLockMutationTestHook func()
internal/kanban/integration_lock.go:239:		if integrationLockMutationTestHook != nil {
internal/kanban/integration_lock.go:240:			integrationLockMutationTestHook()
internal/kanban/kanban_helper_test.go:135:			integrationLockMutationTestHook = func() {
```

One assignment, in a `_test.go` file, compiled only into the test binary. The call site is
nil-guarded, so the production path is a single nil comparison.

**(b) GREEN still discriminates.** I exercised a lock-free implementation *through the same hook* by
inserting `return fn()` as the first statement of `withIntegrationLockMutation` — the AC-ILA-006
revert — and re-ran the identical criterion:

```
=== RUN   TestIntegrationLockAcquire_SerializedAcrossProcesses
    A: RESULT=acquired REPLACED=none SESSION=lane-a
    B: RESULT=acquired REPLACED=none SESSION=lane-b
    round: successes=2 refusals=0 other=0 sessions_differ=true mid_record_held=true mid_record_stale=false final_holder="lane-a" final_record_stale=false
    successes=2 refusals=0, want exactly 1 and 1 — the record's read-modify-write is NOT serialized across processes.
      attributed_double_hold=true (REPLACED=none on both: true; session ids differ: true; record non-Stale at the second acquire: mid_held=true mid_stale=false; final record non-Stale: true)
      final record: holder="lane-a" pid=50507 pid_source="session-owner"
--- FAIL: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.02s)
```

Restored, then re-ran:

```
    A: RESULT=acquired REPLACED=none SESSION=lane-a
    B: RESULT=held REPLACED=none SESSION=lane-b
    round: successes=1 refusals=1 other=0 sessions_differ=true mid_record_held=false mid_record_stale=false final_holder="lane-a" final_record_stale=false
--- PASS: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.54s)
```

Restoration proof: `git status --short` → no output; `git diff --stat` → no output. The tree is
byte-identical to `3c7886d9a`.

The criterion therefore cannot pass for a reason other than exclusion: with the exclusion removed and
everything else held constant, it goes red on attempt 1, with attribution.

### 2. Is the RED attributed, or merely observed? — **Attributed, and the three parts are asserted**

`successes=2` alone is also produced by a stale takeover and by a same-session re-acquire. All three
discriminators are live in the test body, not merely printed:

- **`REPLACED=none` on both** — the helper reports `REPLACED=<session>` when a takeover occurred
  (`kanban_helper_test.go:161-164`); the RED report asserts the conjunct as `attributed_double_hold`,
  and my run printed `true`. A takeover is separately excluded on the pass side by the `finalStale`
  Fatalf.
- **Distinct session ids** — asserted, not assumed: `integration_lock_cross_test.go:244` fatals when
  `a.out.session == b.out.session`, reading the ids the *children* reported rather than the ids the
  parent passed. My run: `sessions_differ=true`.
- **Winner non-`Stale()`** — the record is read at the moment the second acquire has had its chance
  (`midRecord`, line 220), before A is released, and again at the end (`final`). Both `midStale` and
  `finalStale` were false in my RED run, so the double-hold was live, not two callers reclaiming a
  dead record.

The PID discipline is what makes non-staleness reachable at all: children record `HELPER_OWNER_PID`
(the parent test process, alive for the whole round) rather than `os.Getpid()`. A child recording its
own pid would exit immediately after writing, its record would read STALE, and the second child would
take it over legitimately — `successes=2` under a *correct* lock. That trap is closed by construction
and documented at `kanban_helper_test.go:117-125`.

Alternative causes both excluded. The RED means the race.

### 3. Does the repair cover BOTH mutation paths? — **Yes; release's surface is byte-unchanged**

`git diff 15453140a..HEAD -- internal/kanban/integration_lock.go` shows `ReleaseIntegrationLock`
wrapped in the *same* `withIntegrationLockMutation` call as acquire (`integration_lock.go:259`), with
the read → holder-check → remove sequence moved inside it verbatim. Every sentinel and every format
string is preserved character-for-character:

- `ErrIntegrationLockNotHeld` — returned bare, as before.
- `fmt.Errorf("%w: %s (pid %d) holds it", ErrIntegrationLockForeign, current.holderLabel(), current.PID)` — identical.
- `fmt.Errorf("integration lock: %w", remErr)` on a failed `os.Remove` — identical (only the local variable was renamed `err` → `remErr` to avoid shadowing the named return).
- The `errors.Is(remErr, os.ErrNotExist)` tolerance — identical.

t320's surface is untouched: `git diff --stat 15453140a -- internal/cli/integration.go` → empty, and
`git diff --stat a5f414a8a -- internal/cli` → empty.

Behavioral confirmation: `TestReleaseIntegrationLock_HolderAndForeign` and
`TestReleaseIntegrationLock_EmptyIsReported` both `--- PASS` unmodified.

One honest limit — see F2: no criterion measures release's *cross-process* serialization. It inherits
exclusion from the shared wrapper, and the wrapper's exclusion is what AC-ILA-006 proves, so the
inference is sound; but it is an inference, not a measurement.

### 4. Lifetime separation — **holds; no path lets a mutation-lock event touch the window**

Mechanically traced:

```
$ grep -rn "integrationMutationLockPath\|clearStaleLockAtPath\|withIntegrationLockMutation" internal/
```

- `clearStaleLockAtPath` has exactly two callers: `ClearStaleBoardLock` (board path) and
  `clearWedgedIntegrationMutationLock` (mutation path). **Neither is ever called with
  `integrationLockPath`.** The wedge clear can unlink `integration-mutation.lock` and nothing else.
- No read path consults the mutation artifact to answer who holds the window. `Held()` and `Stale()`
  read only the record; `ReadIntegrationLock` opens only `integrationLockPath`.
- **Timeout:** budget exhaustion returns `ErrIntegrationLockBusy` from
  `acquireIntegrationMutationLock` *before* `fn` runs, so no write and no removal occur — a busy
  acquire cannot silently take over a live window, and a busy release cannot silently drop one.
- **Wedged lock:** the Windows recovery removes the artifact only when its recorded owner is
  positively observed absent, with a re-read immediately before the unlink. On Unix it is a
  documented no-op (flock dies with the process).
- **Filename stem** is deliberately `integration-mutation.lock`, not sharing the record's
  `integration-lock` stem, so a future `.moai/state/integration-lock*` glob cannot sweep both
  lifetimes into one set.

The budget derivation checks out: `boardLockSupportedWriters = 10`, `boardLockCIMutationCost = 33ms`,
`boardLockHeadroom = 5` (`board_store.go:96-117`) → 1.65s, inherited rather than re-guessed.

### 5. Preserved single-threaded semantics — **verified, and the pre-existing file is untouched**

```
$ git diff --stat 15453140a -- internal/kanban/integration_lock_test.go
(no output)
$ git merge-base --is-ancestor 15453140a HEAD; echo $?
0
```

The pin is an ancestor commit SHA, not a moving ref, so the emptiness claim cannot silently falsify
itself. `go test ./internal/kanban/... -run IntegrationLock -count=1 -v` → 16/16 PASS, including
every pre-existing criterion this concern names:

```
--- PASS: TestAcquireIntegrationLock_SameSessionReacquires (0.00s)
--- PASS: TestAcquireIntegrationLock_TakesOverAStaleHolder (0.00s)
--- PASS: TestAcquireIntegrationLock_ForceTakesOverALiveHolder (0.00s)
--- PASS: TestAcquireIntegrationLock_RefusesASecondLiveSession (0.00s)
--- PASS: TestReleaseIntegrationLock_HolderAndForeign (0.00s)
--- PASS: TestReleaseIntegrationLock_EmptyIsReported (0.00s)
--- PASS: TestReadIntegrationLock_CorruptRecordIsNotAFreeWindow (0.00s)
--- PASS: TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale (0.00s)
--- PASS: TestReadIntegrationLock_LegacyRecordWithoutPIDSource (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.802s
```

---

## AC-by-AC, decided on my own evidence

| AC | Verdict | My evidence |
|---|---|---|
| AC-ILA-001 (RED, attributed) | **PASS** | Observed by me, attempt 1, via the AC-ILA-006 revert: `successes=2 refusals=0 other=0`, `attributed_double_hold=true`, `REPLACED=none` on both, `sessions_differ=true`, `mid_stale=false`, `final_record_stale=false`. |
| AC-ILA-002 (GREEN) | **PASS** | `--- PASS (0.54s)`; `A: RESULT=acquired REPLACED=none SESSION=lane-a`, `B: RESULT=held REPLACED=none SESSION=lane-b`, `successes=1 refusals=1 other=0`, `final_holder="lane-a"`. No `RESULT=busy` in any of my three rounds. |
| AC-ILA-003 (positive control) | **PASS** | `--- PASS: TestIntegrationLockAcquire_ConcurrencyPositiveControl (0.02s)`. Refusal is conditional on contention, not on concurrency. |
| AC-ILA-004 (busy ≠ held) | **PASS** | `--- PASS: TestIntegrationLockBusy_IsNotHeld (0.00s)`. I confirmed the `%v`-not-`%w` wrapping at `integration_lock_mutation.go:126`, so neither `IsIntegrationLockHeld` nor `IsBoardLockHeld` answers true for busy. |
| AC-ILA-005 (semantics preserved) | **PASS** | 16/16 PASS above + empty `git diff --stat 15453140a -- internal/kanban/integration_lock_test.go`. |
| AC-ILA-006 (mutation guard) | **PASS** | Both directions run by me, in this tree, this session. Disabled → attributed RED; restored → PASS; restoration proved by empty `git diff --stat` and clean `git status --short`. |
| AC-ILA-007(a) (windows compile) | **PASS** | `GOOS=windows go vet ./internal/kanban/...` → exit 0, **zero output bytes**. `GOOS=windows go vet -tests=true` also exit 0, so the windows-tagged test file type-checks (`deadPIDWin` resolves from the pre-existing `board_lock_clear_windows_test.go`). Compilation only — cited as nothing more. |
| AC-ILA-007(b) (windows runtime) | **UNVERIFIED — cannot be decided here** | This code is `//go:build windows`. No darwin-lane command executes it, and I ran none that claims to. I did not observe the dead-owner clear or the live-owner refusal at runtime. CI's windows job is the only surface that can decide this. |
| AC-ILA-008 (unique staging path) | **PASS** | `grep -n 'path + "\.tmp"' internal/kanban/integration_lock.go` → no output. `--- PASS: TestWriteIntegrationLock_UniqueStagingPath (0.00s)` — no residue, mode 0644. |
| AC-ILA-009 (no child outlives the run) | **PASS** | `pgrep -f "[k]anban\.test" \| wc -l` → `0` before, `0` after `go test ./internal/kanban/... -count=1`. `0`/`0`, not `1`/`1`, so the self-match exclusion worked rather than the measurement being broken. |

**Independent-measurement reconciliation with the lead.** Reproduced, not inherited:
`go test ./internal/kanban/... -count=1` → `ok ... 14.227s` (lead: 13.916s; pre-change baseline
13.758s — all three green, spread is machine noise). `GOOS=windows go vet` → exit 0, zero output.
`grep` for the hook assignment → one hit, `kanban_helper_test.go:135`. `pgrep` → 0 before, 0 after.

**`TestConcurrencyStress` attribution.** It did **not** fail in any of my runs — not in the full
package run, not in the coverage run, not in the disabled-section run. I have nothing to attribute to
card t354 in either direction from this session.

---

## Findings

No blocking finding. All five are optional (style, disclosure precision, or unmeasured-but-inferred).

- **F1 [Low] [optional]** `internal/kanban/integration_lock_cross_test.go:276-297` — AC-ILA-002 states
  that `REPLACED=<session>` on either child FAILS the criterion, but that condition is only evaluated
  inside the *failure* branch (as part of `attributed`). On the passing path only `finalStale` is
  asserted; `a.out.replaced`/`b.out.replaced` are logged, never checked. A hypothetical
  `successes=1 refusals=1` round in which the winner performed a takeover would pass.
  *Reachability:* the round starts from a fresh `t.TempDir()` with no prior record, so no takeover is
  constructible — this is an acceptance-text-vs-assertion gap, not a live hole.
  *Required fix (if taken):* add `if a.out.replaced != "none" || b.out.replaced != "none" { t.Fatalf(...) }`
  before the `successes != 1` check.

- **F2 [Low] [optional]** `internal/kanban/integration_lock.go:259` — no criterion measures
  `ReleaseIntegrationLock`'s cross-process serialization. It is under the same wrapper as acquire, and
  AC-ILA-006 proves the wrapper excludes, so release inherits exclusion by construction; but §D.6's
  "the record layer serializes its own mutations" is *measured* only through acquire.
  *Required fix (if taken):* a second cross-process round driving two concurrent releases through the
  same interleaving hook, or an explicit statement in §D.4 that release's serialization is verified by
  shared-wrapper inference rather than by measurement.

- **F3 [Low] [optional]** `internal/kanban/integration_lock_mutation.go:79` — `ReleaseIntegrationLock`
  gained a side effect it did not have: `withIntegrationLockMutation` runs `os.MkdirAll` on
  `.moai/state/` and creates an inert `integration-mutation.lock` there *before* the holder check, so
  releasing on a tree with no state dir now creates one and leaves a lock artifact, then returns
  `ErrIntegrationLockNotHeld` exactly as before. The author disclosed the mkdir; the artifact residue
  is the unstated half (on Unix, flock artifacts are never unlinked — `board_lock_unix.go` closes the
  fd only). **Judgment: a disclosed trade-off, not a defect** — the sentinel, the message, and the
  return are unchanged, and `.moai/state/` exists in every real project. Untested either way.

- **F4 [Info] [optional]** `internal/kanban/integration_lock_mutation.go:118-122` — the disclosed
  residual "`ErrBoardLockChangedHands` now reachable from the integration mutation path with
  board-naming text" **over-states the risk**. `acquireIntegrationMutationLock` discards `clearErr`
  entirely (`clearErr == nil && report != nil && report.Removed`) and falls through to
  `ErrIntegrationLockBusy` on every other path, so the board-named sentinel and its text
  (`"kanban board lock changed hands between inspection and removal"`, `board_lock.go:36`) can never
  surface through an integration verb. The disclosure is conservative rather than wrong; a reader
  could still act on a risk that does not exist.

- **F5 [Low] [optional]** `internal/kanban/integration_lock_cross_test.go:63` —
  `integrationStallReleaseTimeout = 500ms` couples a MUST-PASS criterion to wall-clock. I verified the
  margin: 500ms / 1.65s = 30.3%, leaving B ~1.15s of budget after A is released, and the derivation
  chain (`board_store.go:96-117`) is real, not asserted. The coupling **fails loudly** — a `RESULT=busy`
  triggers a Fatalf that names the misconfiguration explicitly — rather than silently. Residual: on a
  heavily loaded machine B could still exhaust its budget and produce a spurious MUST-PASS failure.
  *Judgment: a disclosed trade-off, correctly sized.*

---

## Gaps (explicitly NOT observed by me)

- **Windows runtime behavior** — the whole of AC-ILA-007(b). `//go:build windows` means no command I
  ran executed `clearWedgedIntegrationMutationLock`,
  `TestIntegrationMutationLock_DeadOwnerDoesNotWedge`, or
  `TestIntegrationMutationLock_LiveOwnerIsNotCleared`. I verified they *compile* and nothing more. The
  card's Windows claim rests entirely on CI, and the SPEC says so honestly.
- **The busy path end to end** — no criterion drives an integration acquire to budget exhaustion on
  Unix. `TestIntegrationLockBusy_IsNotHeld` constructs the sentinel by hand rather than provoking it.
  The author records this gap in §E.2 M2 and M6; I confirm it stands.
- **Full-suite verdict** — `go test ./...` is prohibited locally in this repository; every run above
  was scoped to `./internal/kanban/...`. CI is the full-suite judge.
- **The 0.86 audit figure and the record-write behavior under real lane load** — not measured; this
  audit is package-scoped and synthetic.

## §E honesty check

Every §E.2/§E.3 claim I sampled carries a named command and verbatim output, and I re-ran the
load-bearing ones rather than inheriting them. Coverage (`86.5%`), lint (`0 issues.`), both boundary
diffs, the windows vet, the hook grep, the `.tmp` grep, and both cleanup readings all reproduce. I
found **no §E claim whose evidence is a summary standing in for an output**. The Gaps sections are
substantive rather than decorative: the CI-judged deferral is labeled as such and not counted as a
local PASS (`ac_pass_count: 9`, `ac_ci_judged_count: 1`), the AC-ILA-009 command-form deviation is
disclosed with its reasoning, and the `TestConcurrencyStress` non-attribution is stated in both
directions. The `run_commit_sha` placeholder was backfilled in `3c7886d9a` per the schema exemption.

## Residual risk

The repair is a coordination signal, not a capability boundary — as its own header says. Two
processes are now serialized inside the record's read-modify-write, but nothing prevents a caller from
writing the record outside the lock, and the unique staging path is defence in depth against exactly
that future caller rather than against anything observed. The Windows wedge recovery inherits the
board clear's documented TOCTOU residual (AP-29) unchanged; it is narrowed by the pre-removal re-read,
not closed. And the whole mechanism remains single-host: two lanes on different machines sharing a
checkout over a network filesystem are outside both the lock's reach and this card's scope.

# SPEC-DRIFT-CACHE-FILL-001 — Acceptance Criteria

Authored against `.claude/rules/moai/development/verification-completeness.md`:
every criterion states **WHEN** it runs, the **INPUT** that turns it red, **WHO
SEES** the failure, and its **continued-firing** answer (§1.3 — how a reader
learns the check stopped firing). Each is adopted as a **RED-now cell** plus a
**green path cell**.

**Mutant-probe scope (claim stated precisely).** A mutant probe is recorded on
every one of the 14 release-blocking criteria in §B. The two regression-guard
criteria in §C carry none, and the reason is stated there: they are measurements
of record, not gates, so there is no verdict for a mutant to satisfy.

**Document-level tree pin** (binds every criterion carrying no pin of its own):
`881aa4bb86878b2f401244819c8ce73edba8d052`
(`git rev-parse HEAD` on `WT-drift-cache-fill`, 2026-09-18, after the base moved
to local `develop` `881aa4bb8`). All five evidence-ledger cells were
**re-executed on this new pin** and reproduce identically — same empty stdout,
same exit 1. The earlier pins `0236646653179d83c17c1919c7fb389d510f067b`
(iterations 1-2) and `71532427dadf89063f78c35ac5c4f33bde4ebf10` (iteration 3)
are retained below only as provenance.

**Counts**: 16 acceptance criteria — 14 release-blocking, 2 regression-guard.
Tier M ceiling is 16.

**Classification key**

- **release-blocking** — RED is a single read-only invocation re-executable on
  this tree (§2.1); the criterion gates the card's close.
- **regression-guard** — the starting observation cannot be taken as a single
  re-executable invocation (a build-plus-clone-plus-env-gated procedure, or a
  machine-sensitive wall-clock measurement). Per the §2.1 undecidable
  disposition these lose release-blocking eligibility and are **not recorded as
  a pass**.

---

## §A Evidence ledger (RED-now observations)

Each entry is one single-invocation read-only command, its verbatim stdout, and
its exit code, taken on the document-level pin. All five were re-executed and
reproduced by plan-audit iteration 2 on pin
`0236646653179d83c17c1919c7fb389d510f067b`, re-executed by the author on
`71532427dadf89063f78c35ac5c4f33bde4ebf10` when the base first moved, and
re-executed again on the current pin
`881aa4bb86878b2f401244819c8ce73edba8d052` after the second base move — output
and exit code unchanged in every cell, on every pin.

### EL-1 — fill implementation symbol absent

```
command: grep -rn "driftCacheFill" internal --include="*.go"
stdout:  (empty)
exit:    1
```

### EL-2 — opt-out config key absent from the local config

```
command: grep -n "drift_cache_fill" .moai/config/sections/workflow.yaml
stdout:  (empty)
exit:    1
```

### EL-3 — child fill entry point absent from the drift CLI

```
command: grep -n "fill-cache" internal/cli/spec_drift.go
stdout:  (empty)
exit:    1
```

### EL-4 — probe asserting mode absent

```
command: grep -n "MOAI_DRIFT_CACHE_PROBE_ASSERT" internal/hook/session_start_drift_cache_probe_test.go
stdout:  (empty)
exit:    1
```

### EL-5 — opt-out has no typed config field

```
command: grep -n "DriftCacheFill" internal/config/types.go
stdout:  (empty)
exit:    1
```

Carried provenance, **not** a RED-now cell and not adopted: card t666 R3
measured 15 of 15 hook runs leaving the cache absent on a `git clone --shared`
copy with 869 SPEC directories. That measurement is from another tree; it
grounds AC-DCF-015 in §C as provenance only.

---

## §B Release-blocking criteria (14)

### AC-DCF-001 — the handler starts exactly one fill on a cache miss
- **Requirement**: REQ-DCF-001
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a handler whose miss path starts nothing — this tree, where no
  fill code exists.
- **WHO SEES**: test exit code; CI on the branch.
- **RED-now**: EL-1 — the symbol does not exist, so the test cannot compile
  against it. Red for absence of the capability, not because of unrelated files.
- **Green path**: M2. `Handle` runs against a seeded cache-miss tree with the
  spawn seam installed; the seam must record exactly 1 attempt.
- **Continued firing**: the same test asserts the seam variable is non-nil-able
  and is reached — deleting the spawn call site turns this test red rather than
  silently reducing the swept set.
- **Mutant probe**: an always-spawning mutant satisfies this; AC-DCF-002 is its
  paired opposite direction and the two are adopted together.

### AC-DCF-002 — no fill is started on a cache hit, and the advisory is rendered
- **Requirement**: REQ-DCF-011
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a handler that spawns unconditionally, or one that stops
  rendering the advisory on a hit.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M4. Cache seeded to the tree's current HEAD; asserts the spawn
  seam recorded **zero** attempts **and** `Data` carries `status_drift_warning`.
- **Continued firing**: the positive half (advisory present) fails if the
  cache-resolve path is removed altogether, so a deleted feature cannot present
  as a pass.
- **Mutant probe**: a never-spawning mutant passes this and fails AC-DCF-001; a
  mutant that drops the whole deferred block fails the advisory half.

### AC-DCF-003 — the detach contract holds, and a non-`moai` executable spawns nothing
- **Requirements**: REQ-DCF-001 (detach), REQ-DCF-004 (basename guard)
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a spawn that inherits the parent's streams, that calls `Wait()`,
  or that executes whatever `os.Executable()` returned.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M2. Three assertions on the constructed command, taken from the
  builder function rather than from a live exec: `Stdin`, `Stdout`, `Stderr` are
  all nil; the spawn path calls `Process.Release()` and never `Wait()`; and with
  a non-`moai` basename the spawn seam records an attempt while no process is
  started — the distinction `internal/statusline/forge_spawn_gate_test.go`
  already makes for the same guard.
- **Continued firing**: the basename case asserts the seam counter **and** the
  absence of a started process, so removing the guard flips the second assertion
  even though the first is unchanged.
- **Mutant probe**: a mutant that never spawns at all satisfies "no process
  started"; the seam-counter assertion is what excludes it.

### AC-DCF-004 — the fill exits at its deadline and leaves no torn cache file
- **Requirements**: REQ-DCF-003, REQ-DCF-005
- **WHEN**: `go test ./internal/cli/...` and `go test ./internal/spec/...`,
  every run.
- **RED input**: a fill entry point with no deadline, driven against a drift seam
  that sleeps past it; and a cache writer that `os.WriteFile`s in place.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-3 — the entry point does not exist.
- **Green path**: M1 + M2. (a) The fill command with a 50 ms deadline against a
  seam that sleeps 5 s returns in under 250 ms. (b) The cache writer is asserted
  atomic **through `internal/atomicfile.Replace`**, not a hand-rolled
  temp-plus-rename: an interrupted write leaves the previous cache file
  byte-identical, and no temporary artefact remains after a successful write.
- **Continued firing**: (b) asserts on the writer function directly, so replacing
  it with a plain `os.WriteFile` turns the test red rather than removing a case.
- **Mutant probe**: a mutant that returns instantly by never computing satisfies
  the elapsed bound — so (a) is paired with a positive control: the same command
  against a fast seam writes a complete, parseable cache.

### AC-DCF-005 — single-flight under concurrency, and no wedge after abnormal exit
- **Requirements**: REQ-DCF-006, REQ-DCF-007, REQ-DCF-008
- **WHEN**: `go test ./internal/spec/...`, every run, **on both platforms** — the
  lock is a build-tagged pair (`lock_unix.go` / `lock_windows.go` shape), so this
  criterion is written against the platform-neutral lock interface and runs
  under both build tags. The CI matrix's `windows/amd64` job is where the
  Windows half is observed; a build-tag split that left the Windows side with no
  test would be an empty sweep (§1.1) and is explicitly not acceptable.
- **RED input**: a fill with no lock — two concurrent fills both increment the
  compute counter; or a pid-file lock that survives `SIGKILL`.
- **WHO SEES**: test exit code; CI on both matrix legs.
- **RED-now**: EL-1.
- **Green path**: M3. (a) Two fills started **concurrently** (goroutines released
  by a common barrier, not sequentially) against one temp project root with a
  compute-counter seam: the counter is exactly 1, and the loser returned before
  the winner finished — the second clause is what separates a non-blocking lock
  from a blocking one. (b) A holder terminated without cleanup is followed by a
  successful acquire.
- **Continued firing**: (b) fails if the lock is removed entirely only because
  (a) is adopted alongside it; the two are a pair and are never split.
- **Mutant probe**: a blocking-lock mutant satisfies counter == 1 and fails the
  "loser returned early" clause; a no-lock mutant satisfies (b) and fails (a).

### AC-DCF-006 — a simultaneous burst produces exactly one spawn
- **Requirements**: REQ-DCF-009, REQ-DCF-006
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: **any non-atomic claim of the suppression record.** Two shapes,
  both of which this criterion must reject: (i) the child writes the record
  (v0.1.0) — N parents all read an absent record and all spawn; (ii) the parent
  writes it after a read (v0.2.0) — a lockless read-check-then-write, where N
  handlers released simultaneously all finish their reads before the first write
  lands, and all spawn. Shape (ii) is the input the v0.3.0 revision was written
  for: against it the criterion is red, and no amount of ordering the write
  earlier makes it green. A third shape (iii) — an in-process serialisation of
  any kind — is red only against the **cross-process** clause (a) below, which
  is the correction v0.4.0 makes.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M3, in three clauses that must all hold.
  - **(a) cross-process burst count.** N ≥ 8 **separate processes**, not
    goroutines, contend for one cache-miss tree; exactly 1 wins the claim and
    writes the record, and exactly 1 spawn is recorded.

    The shape is the in-tree pattern at `internal/cli/gate_lock_cli_test.go`
    (the `TestGateLockHelperSleep` + `gateSleeperCommand` pair): a helper
    `Test…` function that `t.Skip`s unless its env var is set, re-executed as a
    child via `os.Args[0]` with `-test.run=^<helper>$`. The parent launches
    N such children against one temp project dir, released together, and takes
    a census of which ones report winning.

    **Cleanup guarantee** (the card's [HARD] no-background-load constraint):
    the child **bounds itself** — it performs one claim attempt and exits on its
    own — and `-test.timeout=60s` caps it **from outside**. There is no trailing
    `kill`, which the card names as not-cleanup because it is a line the process
    may never reach.
  - **(b) an existing record suppresses.** With the record **pre-created on
    disk** and naming the current HEAD, a handler invocation records zero spawn
    attempts and leaves the record's contents unchanged. This clause asserts the
    suppression behaviour against a pre-existing record; it is sequential by
    construction and therefore says **nothing** about the cross-process
    property — see the mutant note.
  - **(c) single choke point, and it targets the lock.** A grep over the
    package's non-test sources finds exactly **one** `atomicfile.Claim` call
    site on the fill path, and its argument is the suppression record's
    companion `<record-path>.lock` — never the record path itself. Same guard
    shape already accepted for AC-DCF-012.
- **Continued firing**: (a) asserts an exact census over children that must each
  report in, so a claim that stops happening changes the count rather than
  leaving the test quiet; (c) is a structural grep, so a second claim site or a
  retargeted claim turns it red without anyone having to suspect it.

#### Mutant note for the criterion above (AC-DCF-006)

> **"an error value is not evidence of a syscall — `fs.ErrExist` is a sentinel
> any code can wrap."**
>
> — iteration-3 plan-audit, preserved verbatim at the operator's instruction.

The v0.3.0 revision of this criterion claimed clause (b) excluded the
in-process-mutex mutant. **The iteration-3 audit disproved that claim by writing
the counter-mutant**, and the disproof is the most valuable finding on this card.
The reason is structural and worth stating in full, because the trap is general:
**(a) was in-process**, so a mutex defeats it; **(b) is a sequential check
against an existing file**, so there is no race in it at all; and an error value
proves nothing about how it was produced. Neither clause instantiated
cross-process create-versus-create, which is the only place the property lives.
A criterion that does not instantiate the race cannot witness it, however
confidently it is worded.

**Mutant probe** — five mutants, each with the clause that kills it:

| Mutant | Killed by |
|---|---|
| *child-writes-the-record* (v0.1.0): no parent claim at all | **(a)** — N children contend, all of them win, census > 1 |
| *parent plain-write after a read* (v0.2.0): lockless read-check-then-write | **(a)** — N processes all read absent, all write, all win |
| *in-process `sync.Mutex` around the read-check-then-write* | **(a)**, and **only because (a) is now cross-process** — the mutex does not span processes. It is **NOT** killed by (b); the v0.3.0 text claiming otherwise was wrong |
| *`os.Stat`-then-`fs.ErrExist` under a `sync.Mutex`* — the counter-mutant the iteration-3 audit wrote: it returns an error wrapping the sentinel on a hit, so it **passes (b)** (error wraps `fs.ErrExist`, contents unchanged, no spawn) and **passes an in-process (a)** (the mutex serialises goroutines), while two real processes both `Stat`, both miss, both write, both spawn | **(a)** in its cross-process form, and independently **(c)** — its claim is not `atomicfile.Claim`, so the choke-point grep finds no call site |
| *claim the record itself* rather than a companion lock (the NEW-1 / NEW-2 shape) | **(c)** — the guard asserts the claim's argument is `<record-path>.lock`; and AC-DCF-008's reclaim clauses, which a record-as-claim design cannot satisfy without a TOCTOU window |

### AC-DCF-007 — a child that never runs still suppresses the next session
- **Requirement**: REQ-DCF-010
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a record written by the child. Simulated by a spawn that
  "succeeds" and then exits immediately without writing anything — the broken
  binary, the OOM-kill-at-startup, the mislinked executable.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M3. With the child replaced by a no-op that writes nothing, the
  second `Handle` invocation within the TTL must record **zero** further spawn
  attempts, and the record on disk must name the current HEAD.
- **Continued firing**: asserts the record's existence and its `head_sha`, so
  moving the write back to the child turns this red immediately.
- **Mutant probe**: a mutant that suppresses on any record regardless of HEAD
  passes this and fails AC-DCF-008.

### AC-DCF-008 — a HEAD change re-enables the fill; a future timestamp does not suppress
- **Requirements**: REQ-DCF-009, REQ-DCF-010
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a record keyed on time alone; and a record whose `started_at`
  is in the future (clock stepped back, or a clone carrying a foreign stamp),
  which reads as "younger than TTL" forever.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M3. (a) After a record is written for HEAD A, the fixture
  advances HEAD to B inside the TTL; the next `Handle` records exactly one new
  attempt. (b) A record stamped `now + 1h` is treated as expired: the next
  `Handle` records one attempt. (b2) An **unreadable** record — empty,
  truncated, or unparseable — is treated as expired, the single stated
  disposition: the next `Handle` records exactly one attempt and leaves a
  well-formed record behind. (b3) Every read, expiry judgement, removal and
  rewrite of the record happens while the companion lock is held, and the expiry
  judgement is made at removal time rather than carried over from an earlier
  read — asserted by injecting a record replacement between a handler's read and
  its removal and observing that the handler does not destroy the replacement.
  (c) The resolved constants satisfy
  **TTL ≥ child deadline + slack**, asserted over the values
  `internal/config/defaults.go` resolves — not over literals — because a TTL
  shorter than the deadline lets a second session reclaim a record whose child is
  still legitimately computing, reintroducing the burst one TTL later.
- **Continued firing**: both halves assert an exact attempt count, so a
  suppression path that stops being consulted changes the count.
- **Mutant probe**: a mutant that ignores the record entirely passes (a) and (b)
  and fails AC-DCF-006 / AC-DCF-007; a mutant that hardcodes the inequality in
  the test rather than reading the resolved constants is excluded by (c)'s
  requirement to assert over the resolved values.

### AC-DCF-009 — the miss path spends no join budget on drift
- **Requirements**: REQ-DCF-011, REQ-DCF-012, REQ-DCF-013
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: the current in-band compute — a drift seam set to 10 ×
  `deferredScanJoinBound` makes `Handle` pay the full bound.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1. The existing `session_start_parallel_test.go` bound test
  documents today's behaviour (pay the bound, drop the advisory); the asserted
  behaviour has no implementation on this tree.
- **Green path**: M4. With the drift seam at 10 × the bound, `Handle` on a
  cache-miss tree returns with **elapsed < `deferredScanJoinBound` / 5** and its
  `Data` carries no `status_drift_warning`. The threshold is a fixed ratio, not a
  tester's judgment.
- **Continued firing**: paired with AC-DCF-002's positive control, so a mutant
  that deletes the deferred block cannot pass both.
- **Mutant probe**: dropping the whole deferred block satisfies the timing half
  and fails AC-DCF-002.

### AC-DCF-010 — every fill failure path is fail-open, and the success path still spawns
- **Requirement**: REQ-DCF-013
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a handler that propagates a spawn error, or one that returns
  early and skips the remaining advisory keys.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M5. Table-driven over the failure points (unresolvable
  executable, unwritable state dir, record-write error, lock-open error,
  `Start()` error): each case asserts `Handle` returns its normal output with a
  nil error and the other advisory keys intact. A **positive control** row with
  no injected failure asserts one spawn attempt was recorded.
- **Continued firing**: the positive control fails when the spawn path stops
  being reached, so the table cannot decay into all-skipped rows.
- **Mutant probe**: a never-spawning mutant passes every failure row and fails
  the positive control — which is why the control row exists.

### AC-DCF-011 — the opt-out resolves through the config loader and defaults enabled
- **Requirement**: REQ-DCF-014
- **WHEN**: `go test ./internal/config/...` and `go test ./internal/hook/...`,
  every run.
- **RED input**: the key present as YAML text with no typed field — it decodes
  into nothing, the opt-out does not exist, and the feature is permanently on.
  That is the mutant that would actually ship.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-5 (no typed field) and EL-2 (no key in the local config).
- **Green path**: M5. (a) On a project declaring **no** key at all, the loader's
  resolved value is `true`. (b) On a project declaring
  `workflow.drift_cache_fill.enabled: false`, the resolved value is `false` and
  `Handle` records zero spawn attempts. Both assertions read the **resolved
  config value**, never the presence of text in a YAML file.
- **Continued firing**: (a) fails the moment the struct field or its
  `defaults.go` entry is removed, which is the exact decay mode a text-presence
  assertion could not see.
- **Mutant probe**: the dead-key mutant (YAML in both files, no struct field, no
  default) passes any text-presence check and fails (a) and (b) here.

### AC-DCF-012 — the hook test binary starts zero fill children
- **Requirement**: REQ-DCF-015
- **WHEN**: `go test ./internal/hook/...`, every run.
- **RED input**: a spawn point reachable before the async-seam check.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-1.
- **Green path**: M5. The assertion is made against the **spawn-seam counter**,
  which must read 0 across the package's own run — never against a real process.
  `internal/statusline/forge_spawn_gate_test.go` states in its own words why:
  under `go test` the `isSelfInvocable` guard blocks the exec anyway, so the
  counter is the only instrument that can distinguish "gated before the spawn"
  from "spawn blocked by the guard". The seam sits after the suppression gates
  and before the self-invocation guard, preserving that distinction.
- **Continued firing**: the around-the-seam gap is closed **structurally**, not
  by a census. The package routes every fill spawn through exactly **one**
  `exec.Command` call site, and a guard test asserts that choke point — a grep
  over the package's non-test sources finds exactly one `exec.Command`
  constructing the fill child. A second spawn path added later fails that guard
  test rather than slipping past a counter.
- **Mutant probe**: a mutant spawning via a path not routed through the seam is
  invisible to a seam-only assertion — which is what the single-choke-point guard
  test excludes, since such a mutant necessarily introduces a second
  `exec.Command` site.

### AC-DCF-013 — the fill writes a HEAD-matching cache *(headline behaviour)*
- **Requirement**: REQ-DCF-002
- **WHEN**: `go test ./internal/cli/...`, every run.
- **RED input**: this tree — there is no fill entry point, so nothing writes the
  cache after the hook's own goroutine is torn down.
- **WHO SEES**: test exit code; CI.
- **RED-now**: EL-3, a single read-only invocation on the pinned tree.
- **Green path**: M1. The fill entry point is run to completion against a fast
  drift seam on a temp git repository; afterwards `.moai/state/drift-cache.json`
  exists, parses, and its `head_sha` equals the repository's `git rev-parse
  HEAD`. This is the release-blocking carrier of the card's headline behaviour;
  the field measurement in §C is corroboration, not the gate.
- **Continued firing**: the assertion is on cache content keyed to a HEAD the
  test computes itself, so a writer that stops running leaves the file absent
  rather than leaving the test quiet.
- **Mutant probe**: a mutant writing an empty or garbage file satisfies mere
  existence — so the assertion checks `head_sha` equality and parseability, not
  the file's presence.

### AC-DCF-014 — the probe carries a verdict, an empty sweep is not a pass, and its absence is visible
- **Requirement**: REQ-DCF-016
- **WHEN**: split, and stated explicitly because half of it is on demand.
  **Every run** (`go test ./internal/hook/...`): the extracted classifier's unit
  test, and the probe-liveness assertion. **On demand**: the probe itself, under
  `MOAI_DRIFT_CACHE_PROBE_ASSERT` plus the existing `_BIN` / `_SRC` variables.
- **RED input**: today's probe, which by construction never fails — it measures
  and prints, and `t.Skip`s unless its env vars are set, so an ordinary package
  run prints `ok` whether the probe is healthy, broken, or deleted.
- **WHO SEES**: test exit code for the every-run half; the operator for the
  on-demand half.
- **RED-now**: EL-4.
- **Green path**: M6. (a) The classifier returns FAIL for a run set containing
  one `absent`, PASS for an all-`present(head-match)` set, and a distinct
  non-pass verdict for an empty set. (b) The probe prints its swept run count
  beside the verdict and calls `t.Fatalf` on a non-pass.
- **Continued firing** (the §1.3 carrier for this SPEC's on-demand checks): the
  every-run unit test asserts the probe function exists by name and that its
  env-gate reads the expected variables, so deleting or renaming the probe turns
  the ordinary `go test ./internal/hook/...` red instead of silently removing a
  check nobody runs.
- **Mutant probe**: an asserting mode that prints "FAIL" without failing the test
  is the report-not-verdict defect itself — so (a) asserts on the returned
  verdict value and (b) asserts the probe body acts on it.

---

## §C Regression-guard criteria (2)

Neither carries a mutant probe: a mutant probe tests whether a **verdict** can be
satisfied while the requirement is violated, and neither of these produces a
verdict. Both are measurements of record, labelled as such so nobody reads one as
a gate.

### AC-DCF-015 — cache present after a real hook run *(regression-guard)*
- **Requirement**: REQ-DCF-002, REQ-DCF-016
- **WHEN**: on demand, in the probe's asserting mode against a
  `git clone --shared` copy with a real SPEC tree.
- **Classification rationale**: its RED procedure is build-a-binary →
  `git clone --shared` → run an env-gated test with three variables. That is not
  the single read-only invocation §2.1 requires, so it takes the **undecidable
  disposition**: regression-guard, never recorded as a pass, and the card's close
  does not rest on it. AC-DCF-013 is the release-blocking carrier of the same
  behaviour.
- **Provenance**: card t666 R3 measured 15 of 15 runs leaving the cache absent on
  another tree. Recorded as provenance, not adopted as a RED cell.
- **Expected direction** (recorded, not gated): after the change, a second hook
  run on the same HEAD observes `present(head-match)`.
- **Continued firing**: covered by AC-DCF-014's every-run liveness assertion —
  if this probe is deleted or renamed, the ordinary package run goes red.

### AC-DCF-016 — Handle p50 on the miss path *(regression-guard)*
- **Requirement**: REQ-DCF-013
- **WHEN**: on demand, `MOAI_HANDLE_STAGE_PROBE_N` in
  `internal/hook/session_start_stage_probe_test.go`.
- **Classification rationale**: the starting observation is a wall-clock p50 on a
  loaded developer machine (load average 8-15 at the t666 R3 measurement). A
  re-execution measures the machine as much as the code, so it is not a
  deterministic single invocation — same undecidable disposition.
- **Expected direction** (recorded, not gated): miss-path Handle p50 falls from
  ~390 ms toward the warm ~176 ms, because the 251 ms bounded join no longer
  carries drift. Written into `progress.md §E.2` with the machine's load average
  beside it.
- **Continued firing**: `TestSessionStart_StageClock` is always-on and fails if
  the stage clock stops being laps-instrumented, so the instrument's own removal
  is visible.

---

## §D Definition of Done

1. Every release-blocking criterion in §B is green, with its green-path output
   cited.
2. `go test ./internal/hook/... ./internal/spec/... ./internal/cli/...
   ./internal/config/...` passes, and `-race` passes on `internal/hook`. The
   full-suite verdict comes from CI, not from a local `go test ./...`
   (`CLAUDE.local.md` §4).
3. `GOOS=windows GOARCH=amd64 go build ./...` succeeds — the lock pair is
   build-tagged, and a cross-build is the cheapest evidence that the Windows leg
   compiles (AC-DCF-005's Windows half is observed on the CI matrix).
4. `go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak'`
   passes in both its narrow tier and under `MOAI_TEMPLATE_LEAK_STRICT=1`, since
   the M5 template edit crosses
   `.github/workflows/template-neutrality-check.yaml`. The template-side comment
   for the new key carries no SPEC id, no REQ token and no date.
5. `go vet ./...` and `golangci-lint run` report nothing new on the touched
   packages.
6. The opt-out exists as a typed field with its default in
   `internal/config/defaults.go`, and `make build` has been run so the template
   change is embedded.
7. AC-DCF-015 and AC-DCF-016 are recorded in `progress.md §E.2` with their load
   averages, explicitly labelled measurements rather than gates.

🗿 MoAI

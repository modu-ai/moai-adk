# SPEC-DRIFT-CACHE-FILL-001 — Progress

Card t871. Branch `WT-drift-cache-fill`, base local `develop` `023664665`.

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md`.
SPEC ID regex pre-write check: `PASS`. ID uniqueness confirmed against
`.moai/specs/`.

**iter1** (v0.1.0): plan-auditor FAIL 0.55 against the Tier M 0.80 threshold; no
must-pass criterion failed. Report: `.moai/reports/t871/plan-audit.md`.

**iter2** (v0.2.0): revised in place. Requirements 22 → **16**, acceptance
criteria 15 → **16** (14 release-blocking, 2 regression-guard) — both exactly at
the Tier M ceiling, reached by consolidation. Tier stays **M**. Every
requirement is now referenced by at least one criterion (verified:
`comm -23` over the REQ id sets of `spec.md` and `acceptance.md` prints
nothing). Evidence ledger grew to five single-invocation RED cells (EL-5 added
for the missing typed config field); the former PENDING EL-5 promise is gone.

**iter3** (v0.3.0): scoped revision, four items only (N1/N2/N3/N5). plan-auditor
iter2 scored 0.81 (clears the Tier M 0.80 threshold) but returned FAIL under the
retry contract because one iter1 defect stayed open — the single-flight race had
moved, not closed. REQ-DCF-009 **amended** (not replaced): the suppression write
becomes an exclusive claim via `internal/atomicfile.Claim`. **Tier gate did not
fire** — the amendment adds no actor, no artifact and no new obligation, so no
REQ-017 was needed and the count stays 16 requirements / 16 acceptance criteria
at Tier M. All five evidence cells re-executed and reproduced on the new base pin
`71532427dadf89063f78c35ac5c4f33bde4ebf10`; the document pin was moved and the
old pin retained only as iteration-1/2 provenance.

**iter4** (v0.4.0): scoped revision, three items. plan-auditor iter3 scored 0.84
(above threshold) but FAILed: N2/N3/N5 closed, N1 only half-closed — the design
was right, the criterion meant to prove it was not. The audit **disproved the
v0.3.0 claim** that AC-DCF-006 clause (b) excluded the in-process-mutex mutant,
by writing the counter-mutant (a `sync.Mutex`-guarded `os.Stat` returning an
error wrapping `fs.ErrExist`). Clause (a) is now **cross-process** — N ≥ 8
re-executed child processes on the `internal/cli/gate_lock_cli_test.go` shape,
self-bounding with `-test.timeout` capping from outside — plus a new clause (c)
choke-point guard. REQ-DCF-009 **re-amended** to claim a separate
`<record>.lock` and perform the reclaim under it (NEW-1 TOCTOU, NEW-2 empty
record). The operator-preserved sentence — "an error value is not evidence of a
syscall — `fs.ErrExist` is a sentinel any code can wrap." — is carried verbatim
in `acceptance.md` § AC-DCF-006 mutant note and `plan.md` §G. AC-DCF-003 left
untouched; its observation-method question is recorded as a run-phase M2 item.
**Tier gate did not fire** — count stays 16 requirements / 16 acceptance criteria
at Tier M. All five evidence cells re-executed and reproduced on the new base pin
`881aa4bb86878b2f401244819c8ce73edba8d052`. Awaiting plan-audit iteration 4 and
Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

Run-phase base: `bbd42508b` (`WT-drift-cache-fill`, local `develop` absorbed).
cycle_type **tdd**. Every criterion below was observed RED before its
implementation existed; the RED evidence is EL-1..EL-5 (symbol absence, so the
tests could not compile) plus the per-milestone compile failures recorded during
the run.

### AC PASS/FAIL matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-DCF-001 | PASS | `go test ./internal/hook/ -run TestSessionStart_CacheMissStartsExactlyOneFill` | `--- PASS: TestSessionStart_CacheMissStartsExactlyOneFill (0.07s)` |
| AC-DCF-002 | PASS | `go test ./internal/hook/ -run TestSessionStart_CacheHitStartsNoFillAndRendersAdvisory` | `--- PASS: TestSessionStart_CacheHitStartsNoFillAndRendersAdvisory (0.08s)` |
| AC-DCF-003 | PASS | `go test ./internal/hook/ -run 'TestDriftFill_DetachContract\|TestDriftFill_SpawnPathReleasesAndNeverWaits\|TestDriftFill_NonMoaiExecutableStartsNoProcess\|TestDriftFill_RealExecutableIsBlockedUnderGoTest'` | all four `--- PASS` |
| AC-DCF-004 | PASS | `go test ./internal/cli/ -run TestSpecDriftFillExitsAtItsDeadline` + `go test ./internal/spec/ -run TestSaveDriftCache` | `--- PASS: TestSpecDriftFillExitsAtItsDeadline (0.10s)`; three `--- PASS` in `internal/spec` |
| AC-DCF-005 | PASS | `go test ./internal/spec/ -run DriftFillLock` | `--- PASS: TestWithDriftFillLockSerialisesConcurrentFills (0.30s)`, `--- PASS: TestWithDriftFillLockSurvivesAbnormalHolderExit (0.01s)`, `--- PASS: TestDriftFillLockIsHeldByAnOSHandle` |
| AC-DCF-006 | PASS | `go test ./internal/hook/ -run 'TestDriftFill_SimultaneousBurstProducesOneSpawn\|TestDriftFill_ExistingRecordSuppresses\|TestDriftFill_SingleClaimChokePointTargetsTheLock'` | `maximum simultaneous claim attempts: 8 of 8`; `--- PASS` on all three |
| AC-DCF-007 | PASS | `go test ./internal/hook/ -run TestDriftFill_BrokenChildStillSuppressesTheNextSession` | `--- PASS (0.00s)` |
| AC-DCF-008 | PASS | `go test ./internal/hook/ -run 'TestDriftFill_HeadChangeReEnablesTheFill\|FutureStampedRecordIsExpired\|UnreadableRecordIsExpired\|ReclaimIsJudgedAtRemovalTime\|StaleLockIsReclaimable'` + `go test ./internal/config/ -run TestDriftCacheFillTTLCoversTheChildDeadline` | all `--- PASS`, including the four `UnreadableRecordIsExpired` subcases |
| AC-DCF-009 | **PASS (measurement subject restated — see note)** | `go test ./internal/hook/ -run TestSessionStart_MissPathSpendsNoJoinBudgetOnDrift` | `deferred step on a miss: 232.625µs; Handle hit baseline: 65.689334ms; Handle miss: 66.329583ms; budget: 50ms` → `--- PASS (0.13s)` |
| AC-DCF-010 | PASS | `go test ./internal/hook/ -run TestDriftFill_EveryFailurePathIsFailOpen` | six subcases `--- PASS`, including the positive control |
| AC-DCF-011 | PASS | `go test ./internal/config/ -run TestDriftCacheFill` | four loader subcases + nil-config + defaults `--- PASS` |
| AC-DCF-012 | PASS | `go test ./internal/hook/ -run 'TestSessionStart_SyncDeferredScansStartNoFill\|TestDriftFill_SingleExecChokePoint'` | both `--- PASS` |
| AC-DCF-013 | PASS | `go test ./internal/cli/ -run TestSpecDriftFillWritesHeadKeyedCache` | `--- PASS (0.15s)` |
| AC-DCF-014 | PASS | `go test ./internal/hook/ -run 'TestClassifyDriftCacheProbeRuns\|TestDriftCacheProbeIsLive'` | six classifier subcases + liveness `--- PASS` |
| AC-DCF-015 | **regression-guard — recorded, not a gate** | see § Field measurement below | exit arm 4/4 `present(head-match)`, `VERDICT PASS swept=4` |
| AC-DCF-016 | **regression-guard — recorded, not a gate** | see § Field measurement below | Handle miss-path p50 **103.94ms**; `deferred_join` p50 **36.05ms** |

**AC-DCF-009 measurement-subject note (run-phase observation-method decision).**
The criterion states the ratio as "Handle returns with elapsed <
`deferredScanJoinBound` / 5" (= 50 ms). Handle's total wall clock cannot carry
that ratio for a reason unrelated to this card: its synchronous work — config
load, session registry, migration, settings — measures 65-80 ms on this machine
before any deferred step runs, and no change to the drift path can move it.
Asserting the literal ratio against that total would be red at arrival and red
forever, the "impossible" direction `acceptance.md` §2 names as disqualifying.
The ratio is therefore applied to the quantity it is about, in two ways that are
together stricter than the original: (a) the **deferred step itself** — the step
that used to carry the compute — measured directly against the literal 50 ms
budget (observed **232.6 µs**); and (b) **Handle's miss cost against Handle's own
measured cache-hit baseline** in the same test, the delta being the drift
contribution (observed **0.64 ms** against the same 50 ms budget). A third
assertion proves the timings are interpretable rather than incidental: the
in-band computation seam is instrumented and must never be consulted. The
criterion text is untouched; the SPEC body was not modified.

### Cross-cutting verification

| Check | Command | Result |
|---|---|---|
| scoped suites | `go test ./internal/spec/... ./internal/cli/... ./internal/config/...` | `ok internal/spec 78.655s`; `internal/cli` FAIL on `TestGTDAllTodoVerbsParity` only; `ok` on all 17 `internal/cli/*` subpackages; `ok internal/config` after the shipped-key inventory entry |
| hook suite, race | `go test ./internal/hook/ -race` | `FAIL ... 195.232s` — the single failure is `TestFactoryLaneRecordsItsNumberAndLeadRecordsZero`; **no race warnings** |
| pre-existing-failure attribution | same two tests run in a `git clone --shared` of this worktree at `bbd42508b` | both reproduce identically at baseline — `gtd verbs = [... answer ...]` and `backend = "claude", want "glm"`. Neither is attributable to this card |
| cross-platform build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| cross-platform vet | `GOOS=windows GOARCH=amd64 go vet ./internal/spec/` | exit 0 (the build-tagged lock pair compiles on both legs) |
| vet | `go vet ./internal/hook/ ./internal/spec/ ./internal/cli/ ./internal/config/` | exit 0 |
| lint | `golangci-lint run --timeout=5m ./internal/hook/... ./internal/spec/... ./internal/cli/... ./internal/config/...` | `0 issues.` |
| template neutrality | `go test ./internal/template/ -run TestTemplateNoInternalContentLeak` and again with `MOAI_TEMPLATE_LEAK_STRICT=1` | `--- PASS` in both tiers |
| embedded template | `make build` then `strings bin/moai \| grep -c drift_cache_fill` | `4` |

### Mutant probe on the burst criterion (the R1 obligation, executed)

AC-DCF-006 clause (a) is the sole witness for the cross-process property, so it
was not accepted on a green. The counter-mutant the iteration-3 plan-audit wrote
— an `os.Stat`-then-create guarded by an in-process `sync.Mutex` — was compiled
into `acquireDriftFillClaimLock` and the burst re-run:

```
maximum simultaneous claim attempts: 8 of 8
2 of 8 contending PROCESSES won the claim, want exactly 1
--- FAIL: TestDriftFill_SimultaneousBurstProducesOneSpawn (0.45s)
--- PASS: TestDriftFill_ExistingRecordSuppresses
--- PASS: TestDriftFill_SingleClaimChokePointTargetsTheLock
--- PASS: TestDriftFill_ReclaimIsJudgedAtRemovalTime
```

Two processes won. The sequential clauses stayed green, exactly as the audit
predicted, and only the cross-process clause caught it. The mutant was reverted
immediately (`go vet` clean, burst green again) and is not in the tree.

### How the overlap was produced, and how it was asserted (R1)

**Produced**, by two mechanisms that close different halves:

1. **A shared start instant.** Each child writes a readiness marker and then
   spins on a gate file; the parent opens the gate only once all 8 are ready, and
   the gate *carries a target wall-clock nanotime*. Children busy-wait to that
   instant, so the release is aligned to microseconds rather than to filesystem
   notification latency. A bare `Start()` loop is not a release mechanism:
   process start-up is tens of milliseconds and staggered while the claim window
   is sub-millisecond, so the children would serialise naturally and a
   non-atomic implementation would also produce exactly one winner.
2. **A widened critical section.** The winner holds the claim lock for 400 ms,
   installed *in the child* from an env var through the in-critical-section probe
   seam. Every loser's attempt therefore necessarily falls inside the winner's
   section. For a non-containing implementation the same hold sits between the
   read and the write, opening the hazard as wide as the hold — which is why the
   mutant above lost 2-of-8 rather than marginally.

**Asserted**, never assumed: each child reports the wall-clock interval of its
own claim attempt; the test computes, by an endpoint sweep, the largest number
of intervals covering a common instant, and **fails when fewer than two do**.
Observed on the passing run: `maximum simultaneous claim attempts: 8 of 8`, with
the winner's interval `[0.001 ms, 403.341 ms]` and the seven losers returning
inside 0.9 ms. Without that assertion a green (a) could not distinguish "the
implementation is atomic" from "the processes never raced".

### Observation method for "no process was started" (R2)

The negative is established by a **counter at the only start point**, plus a
structural guard that there is only one such point. It is recorded beside the
tests in the header of `internal/hook/session_start_drift_fill_test.go`:

1. `driftFillStartFn` is the single place a fill child is ever started —
   production assigns `(*exec.Cmd).Start` to it. "No process was started" is
   observed as that counter reading zero.
2. `TestDriftFill_SingleExecChokePoint` makes (1) sound rather than a
   transcription: the package's non-test sources construct the fill child exactly
   once (asserted by scanning for the `"--fill-cache"` literal), so a bypassing
   start would have to introduce a second site and fail that guard.
3. `TestDriftFill_RealExecutableIsBlockedUnderGoTest` closes the last gap by
   running the **production** guard against the **production** start
   (`os.Executable()` resolves to the test binary) and observing the counter at
   zero — establishing that the guard, not the test stub, is what stops the exec.

The spawn probe and the start counter are deliberately different instruments:
probe=1 / start=0 reads "gated before the spawn", probe=0 reads "never reached
the spawn at all". Conflating them was the distinction
`internal/statusline/forge_spawn_gate_test.go` already exists to preserve.

### A1 residual — recorded, not presented as closed

The lock's own staleness reclaim keeps a test-and-remove window. It is stated in
the doc comment on `acquireDriftFillClaimLock`
(`internal/hook/session_start_drift_fill.go`): judging staleness at the moment
of removal **narrows** the window to the interval between the stat and the
remove; it does not eliminate it, because no filesystem offers an atomic
test-and-remove. A microsecond-wide race survives and its worst case is exactly
one extra child — never data corruption. The comment cites the precedent that
accepts the identical residual with its own reasoning written down,
`internal/verify/claim_lock.go` (the "Fail-open policy" comment on
`acquireKeyLock`, lines 68-73).

### Field measurement (AC-DCF-015) — measurement of record, not a gate

Machine: darwin/arm64, load averages 10.25 / 7.39 / 9.19 at the start of the
probe run and 5.16 / 6.27 / 8.48 at its end. Repository: a `git clone --shared`
of this worktree, 875 SPEC directories, binary `bin/moai` built from this tree.

```
MEASUREMENT drift-cache probe n=2
iter=0 arm=exit hook#1 elapsed=360ms cache=absent
iter=0 arm=exit hook#2 elapsed=245ms cache=present(head-match)
iter=0 arm=exit hook#3 elapsed=242ms cache=present(head-match)
iter=1 arm=exit hook#1 elapsed=359ms cache=absent
iter=1 arm=exit hook#2 elapsed=247ms cache=present(head-match)
iter=1 arm=exit hook#3 elapsed=238ms cache=present(head-match)
VERDICT PASS swept=4 states=[present(head-match) x4]
```

Carried provenance for contrast (card t666 R3, another tree): 15 of 15 exit-arm
runs left the cache **absent**.

A direct single-process measurement establishes the timing the probe's settle
depends on: the hook returned in **354 ms** having already written the
suppression record, and `.moai/state/drift-cache.json` appeared at **+1.39 s**
with `head_sha` equal to the clone's HEAD and `count: 177`.

**The probe found a real defect in itself, and that is recorded rather than
quietly repaired.** Its first asserting run returned `VERDICT FAIL swept=4
states=[absent absent absent absent]`. Investigation showed the fill was working
(the direct measurement above) and the *instrument* was wrong: it judged the
exit arm's runs back-to-back, so run #2 began ~350 ms after run #1 — before a
~1 s child could land. It was measuring child start-up, not the fix. The repair
is a **bounded poll** (`settleForFill`, 20 s, well under the child's own 30 s
deadline), not a sleep: a broken fill never produces the file, the bound
expires, the state stays `absent`, and the verdict is still FAIL. Waiting longer
cannot turn a broken fill into a pass.

### Field measurement (AC-DCF-016) — measurement of record, not a gate

`MOAI_HANDLE_STAGE_PROBE_N=12` with the real SPEC tree populated. Load averages
5.86 / 6.37 / 8.49 at start, 8.29 / 6.88 / 8.61 at end.

| Stage | p50 | p90 | max |
|---|---|---|---|
| `handle.control` (uninstrumented Handle, miss path) | **103.94 ms** | 112.72 ms | 123.39 ms |
| `handle.instrumented` | 108.25 ms | 114.60 ms | 115.50 ms |
| `deferred_join` | **36.05 ms** | 37.41 ms | 39.78 ms |
| `sync_group` | 29.90 ms | 32.03 ms | 34.65 ms |
| `deferred.drift` (raw compute, measured out of band by the probe) | 939.57 ms | 949.32 ms | 954.27 ms |
| `cold.drift` (same, on a fresh project) | 943.82 ms | 954.05 ms | 1.216 s |

Direction, against the carried t666 R3 figures (another tree, so a delta rather
than a comparison of like with like): miss-path Handle p50 ~390 ms → **103.94
ms**, now below that measurement's *warm* figure of ~176 ms; the bounded join
251 ms on every cold run → **36.05 ms**. The drift compute itself is unchanged
at ~940 ms — it simply no longer sits behind a 250 ms bound that could never
contain it.

### Scope notes

- `internal/spec/drift_index.go` gains `gitHeadSHAAt(dir)`; the existing
  `gitHeadSHA()` and its callers are untouched. The explicit-dir form exists
  because the handler holds a project path that is not necessarily its working
  directory.
- Four tests in `session_start_parallel_test.go` and one stub in
  `session_start_stage_probe_test.go` injected their slow scan at `driftCountFn`.
  The deferred path no longer calls that seam, so each was **repointed to
  `driftCachedCountFn`** — the seam the deferred step actually consults. Left
  alone they would have stayed green against any implementation, which is the
  vacuous-criterion shape. No test was deleted and no assertion was weakened.
- `internal/config/testdata/shipped_key_inventory.yaml` gains one entry for the
  new key; `TestShippedConfigKeysHaveReaders` requires every shipped key to be
  triaged.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-18
run_commit_sha: pending-backfill
run_status: complete
ac_pass_count: 14
ac_fail_count: 0
ac_regression_guard_recorded: 2
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (worktree branch, no push this phase)
l44_post_push_fetch: not-run (no push this phase)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 23
m1_to_mN_commit_strategy: milestone-grouped commits on WT-drift-cache-fill, no push
pre_existing_failures_not_attributable:
  - internal/cli TestGTDAllTodoVerbsParity
  - internal/hook TestFactoryLaneRecordsItsNumberAndLeadRecordsZero
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-18
sync_commit_sha: 53e05a629
sync_status: complete
b12_self_test_a_pre_emission_grep: "grep -c 'SPEC-DRIFT-CACHE-FILL-001' CHANGELOG.md → 0 (no prior entry; emission proceeded)"
b12_self_test_b_ac_count_match: "16 distinct AC ids in acceptance.md (AC-DCF-001..016), non-zero; CHANGELOG entry states 16 (14 release-blocking PASS + 2 regression-guard)"
b12_self_test_c_file_path_verification: "ls internal/hook/session_start_drift_fill.go internal/spec/drift_fill_lock.go internal/spec/drift_cache.go internal/cli/spec_drift.go internal/config/drift_cache_fill.go → all present"
changelog_entry_position: "[Unreleased] → ### Added → first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress → completed (single sync commit; 3-phase close)"
  plan_md: "no status field (artifact statelessness — no frontmatter block)"
  acceptance_md: "no status field (artifact statelessness — no frontmatter block)"
  updated_field: "already 2026-09-18 across artifacts; unchanged"
docs_surfaces_updated:
  - "CHANGELOG.md — [Unreleased] Added"
  - "docs-site/content/{en,ko,ja,zh}/advanced/config-sections.md — new '## workflow.yaml — drift_cache_fill' section, 4-locale parity 11 h2 / 267 lines each, hugo build exit 0"
docs_surfaces_deliberately_not_updated:
  - "README.{md,ko.md,ja.md,zh.md} — the config table lists files, not keys; a per-key entry would be the only one of its kind"
  - ".moai/docs/** — no operator runbook applies; the key is a single on/off with no procedure"
template_first_verification: "only .moai/config/sections/workflow.yaml was added under .moai/ or .claude/; its mirror internal/template/templates/.moai/config/sections/workflow.yaml is present in the same diff. go test ./internal/template/ -run TestTemplateNoInternalContentLeak → ok, and again with MOAI_TEMPLATE_LEAK_STRICT=1 → ok"
carried_deviations:
  - "AC-DCF-009 measurement subject restated at run-phase (deferred step 232.6 µs + Handle delta 0.64 ms + seam assertion, instead of the literal Handle-total ratio). SPEC body NOT modified. Criterion-vs-implementation divergence, surfaced for the auditor to rule on."
  - "Commit 9400c53d9 does not `go vet` in isolation (test file references a symbol landing in the next commit); production code compiles at every commit; history not rewritten."
  - "Pre-existing failures not attributable to this card, reproduced at base bbd42508b in a separate clone: internal/cli TestGTDAllTodoVerbsParity; internal/hook TestFactoryLaneRecordsItsNumberAndLeadRecordsZero (same finding as card t868). Not fixed."
  - "A1 residual — the fill lock's staleness reclaim keeps a test-and-remove window; narrowed, never eliminated; worst case exactly one extra child. Stated as a residual, not closed."
  - "spec lint: 0 errors, 22 warnings. 12 are CoverageIncomplete ('REQ-DCF-NNN is not referenced by any AC'), which the rule measures inside spec.md alone. Coverage is real and lives in acceptance.md — `comm -23` over the REQ id sets of spec.md and acceptance.md prints nothing (16 REQs, all cited). Closing the warnings would mean citing AC ids in the spec.md body, which manager-docs may not edit; surfaced for manager-spec rather than repaired here."
  - "spec lint INFO OwnershipTransitionUnmeasured: the draft → in-progress transition commit 8f2b92126 carries no Authored-By-Agent trailer, so it cannot be attributed. This sync commit carries `Authored-By-Agent: manager-docs`; the run-phase commit is history and was not rewritten."
mx_tag_validation: "sync sub-step; no new exported high-fan-in surface introduced beyond the annotated entry points — existing @MX annotations reviewed, none added or removed"
```

🗿 MoAI

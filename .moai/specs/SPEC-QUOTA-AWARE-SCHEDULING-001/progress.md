# Progress — SPEC-QUOTA-AWARE-SCHEDULING-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L), plus decision-index.md
- plan_complete_at: 2026-10-02
- open_decisions: none — DO-1..DO-12 resolved (oracle; leader verdict on DO-3, DO-7, DO-8; DO-12 `write_backend_at_claim` verified then applied at spec 0.4.0)
- plan_audit_verdict: iteration 2 PASS 0.87 (Tier L threshold 0.85; `.moai/reports/t1347/plan-audit-iter2.md`, audited_sha 7fe1ee49d, local-only). Iteration 1 was FAIL 0.74 (`plan-audit-iter1.md`); the revision at spec 0.5.0 closed D1-D24. Ten MINOR findings (N1-N10) are carried as known plan debt in section F below.
- open_decisions_note: DO-13 (first-exhausted time surfaced in the status block) is PROVISIONAL, escalated to the leader

## §E.2 Run-phase Evidence

_pending run-phase_

### M0 — Baseline goldens (REQ-QAS-016, AC-QAS-015; card t1347)

Commit order witness: goldens commit `c2ae5236af96e8e76fb5c27639c22380bdec5dfa` (parent `3dc73dc103b51d46f1cab6faf43f33e692075a95`) touches only the three golden files; this evidence commit follows it. Measured on tree `bad02c7c047cf9e5a99e069dfa505589c2971a95` (HEAD `3dc73dc10`), branch `WT-quota-aware-scheduling`, toolchain `go version go1.26.8 darwin/arm64`. `moai version` printed `v3.2.0-rc.24 ... gc50da9c2f` (judging build for the slot verbs only; no golden was produced by the installed binary — each was produced by the `go test` toolchain from this tree).

**Pre-change proof (measurement integrity).** `git diff --name-only c50da9c2f..HEAD -- internal` at HEAD `3dc73dc10`, before any generator ran: printed nothing (empty = `internal/` code byte-identical to `c50da9c2f`, the tree the installed build names). `grep -rln "TestQAS_" internal` printed nothing.

**Step 0 — RED-selector record (every anchored selector prints the empty-sweep token on the unmodified tree).** Each row is one invocation of the form `<command>; echo "rc=$?"` with `go -C <card tree>` (the shell cwd was the SPEC directory); `-v -count=1`.

| Package | Selector (anchored, `-run`) | Verbatim empty-sweep output | rc |
|---|---|---|---|
| `./internal/statusline` | `^(TestQAS_AC001_RecordCarriesSuppliedWindowsOnly\|TestQAS_AC002_PreviousSchemaReadsAsNoWindows\|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline\|TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop\|TestQAS_AC004_ExhaustedAtStickyUntilRollover\|TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover\|TestQAS_AC006_FailOpenOnAbsentOrUnreadable\|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.380s [no tests to run]` | 0 |
| `./internal/config` | `^TestQAS_AC007_ConfigDefaultsMirrorTemplate$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/config	0.338s [no tests to run]` | 0 |
| `./internal/cli` (slot `go-test-cli` held) | `^(TestQAS_AC006b_NextLeasesWhenQuotaDataAbsent\|TestQAS_AC008_ClaudeLaneHeldAtThreshold\|TestQAS_AC008b_MCPFactoryNextHeld\|TestQAS_AC009_HoldLineCarriesResetTime\|TestQAS_AC010_WaitLatchReleasesOnlyBelowMarginOrResetOrUnknown\|TestQAS_AC011_NonClaudeBackendsNeverHeld\|TestQAS_AC011b_StageAndCompleteIgnoreQuotaHold\|TestQAS_AC012_StatusQuotaBlockOnlyWhenGateEnabledWithData\|TestQAS_AC013_AcquireWarnsNeverRefuses\|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly\|TestQAS_AC017_SharedPressureEvaluationAndSurfaces\|TestQAS_AC018_LaneInventoryCandidates\|TestQAS_AC019_AutoAndStatusRecommendNonClaudeLanes\|TestQAS_AC020_NoNonClaudeLaneWarnsOnly\|TestQAS_AC021_SteeringChangesNothing\|TestQAS_AC022_GateOffOrPressureOffOutputUnchanged\|TestQAS_AC023b_LaunchersPassTheirBackendToTheClaim)$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	1.223s [no tests to run]` | 0 |
| `./internal/kanban` (slot held) | `^TestQAS_AC023_ClaimRecordsBackend$` | `testing: warning: no tests to run` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/kanban	0.406s [no tests to run]` | 0 |

Classification: all four are the intended RED of a test that does not exist yet (`[no tests to run]` with rc 0 is NOT a pass: verification-completeness §1.1; the swept count is zero). Selector coverage: AC-001..-014, -017..-023 (21 Go deciders; AC-QAS-014 is named in both `./internal/statusline` and `./internal/cli`; AC-QAS-015 and -016 are decided by git commands, not Go tests). The first statusline attempt without `-C` failed `[setup failed]` (cwd was the SPEC directory); the row above is the re-run. Slot lease: `moai slot acquire --resource go-test-cli --max-duration 20m` printed `slot go-test-cli acquired by 2da35a68-1196-4183-b6e5-a50fc9b6d901 until 2026-10-02T07:22:41Z` rc=0; `moai slot release --resource go-test-cli` printed `slot go-test-cli released (was 2da35a68-1196-4183-b6e5-a50fc9b6d901)` rc=0 (released before this report).

**Goldens (measured by a temporary generator test `TestZZQASBaselineGen` per package, deleted before staging; two consecutive runs produced byte-identical files).** `internal/cli` generator run as `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_ROLE MOAI_FACTORY_WORKERS MOAI_FACTORY_CLEAR_POLICY MOAI_FACTORY_AUTO_DISPATCH MOAI_LAUNCH_PROVIDER && go -C <card tree> test -v -run '^TestZZQASBaselineGen$' -count=1 ./internal/cli` (the first attempt under the lane environment failed: `todo add: moai add: refused — lane boundary`, which is the lane environment falsifying the fixture, not a golden); the statusline generator ran WITHOUT a scrub (not measured whether any lane variable could reach it; the golden bytes carry no env-derived value, and the file was byte-identical across two runs).

`shasum -a 256` (run after the second generator run, on the committed files):

```
46db7e4a8922c16a2566d2b23a91ceac7dcb7f212d635773bb7f8fce153f3726  internal/statusline/testdata/qas_baseline_windowless_record.golden.json
d52210a78575b65cbc2a7bfc929cf394610eb31ae65eb357c44b11fcedb7ce09  internal/cli/testdata/qas_baseline_factory_status.golden.json
570dc05140007a68d165c4b077c5496cf5c3d03876fcc790b9301ddbc7f7bc65  internal/cli/testdata/qas_baseline_todo_auto.golden.txt
```

1. `qas_baseline_windowless_record.golden.json` (AC-QAS-002b). The fixed input is `writeContextUsage(<TempDir>, "qas-baseline-session", 4242, MemoryData{Available: true, ContextWindowSize: 200000, TokensUsed: 50000}, handoffStageNone, "Opus 4.8", "high")`; the bytes are the on-disk file the writer produced (MarshalIndent, two-space indent, no trailing newline), with exactly one edit: the `captured_at` value was replaced by the constant `2026-10-02T00:00:00Z` (the writer stamps `time.Now()`, so the value cannot be a fixed byte string otherwise). `writer_pid` is the fixed argument 4242; `schema_version` is the pre-change `2`. Per AC-QAS-002b the later test normalizes only `captured_at`, `writer_pid`, and `schema_version` before comparing, so those three fields carry no identity here. Bytes: keys in order `schema_version, session_id, writer_pid, captured_at, context_window_size, tokens_used, raw_pct, stage, band, model, effort`; `raw_pct` is `25`, `stage` `none`, `band` `standard`.
2. `qas_baseline_factory_status.golden.json` (AC-QAS-012, -022). One file keyed `empty` and `one_card`, each value the exact `moai factory status --run run-cli --json` output (cobra `newFactoryCommand`, `runFactory` helper) re-indented inside the wrapper. Extraction rule (self-checked in the generator, no mismatch): `json.Unmarshal` the file into `map[string]json.RawMessage`, `json.Indent(&buf, raw, "", "  ")`, append one `\n`; the result equals the command's stdout byte for byte. Fixtures: `empty` = `fcFixture(t)` only (no factory database: output `{"run":"run-cli","cards":[],"unavailable":[]}`); `one_card` = `fcFixture`, `fcQueue(store, queued)`, `fcClassify("t1", high, false, serial)`, `sdRegisterLane("lane-1")`, `fcPlace(Card{CardID:"t1", State:leased, LeaseHolder:"lane-1", LeaseExpiresAt:"2026-09-26T10:00:00Z", Stage:run})` under the fixed clock `fcNow` = `2026-09-26T09:00:00Z`. No volatile field exists in this output (no temp path, no wall-clock value), so nothing is normalized. The `one_card` fixture differs from `TestFactoryStatusShowsHolderModePriority` in one respect: one card instead of two, and `LeaseExpiresAt` set so `lease_expired` is `false`.
3. `qas_baseline_todo_auto.golden.txt` (AC-QAS-022). `runAutoCycle` over two queued cards `t1` "card a" and `t2` "card b" (added in that order through `store.Add`, not through a map), `autoOptions{wait: 5m, sessionID: "operator-session-fixture", jev: stub returning "jev: stubbed (qas baseline fixture)", liveness: autoTestLiveness(root, "t1", true, true, nil), sleep: writes <root>/.moai/reports/<id>/evidence.md for card tick, now: unix 0 + tick minutes}` and nil `landed` / `jevRank` (so the ranking stage prints `selection: source=fallback reason=jev-disabled`, `ranked t1 t2`, and the landed-signal non-finding). Both cards complete (`done t1`, `done t2`, two `/clear guidance` blocks). The only normalization: every occurrence of the per-run `t.TempDir()` root in the evidence path is rendered as the literal `<ROOT>`; the later test applies the same `strings.ReplaceAll(out, root, "<ROOT>")`. 28 lines.

**AC-QAS-015 baseline fact (the commit-graph part).** Each B_i is the single commit `c2ae5236af96e8e76fb5c27639c22380bdec5dfa` (all three goldens in one commit, so B_1 = B_2 = B_3); it touches no file under `internal/` that is neither a `_test.go` nor under `testdata/`, so every future implementation commit I follows it. The decider evaluation (`git log --reverse ... -- internal ':(exclude)*_test.go' ':(exclude)*/testdata/*'` and the `git rev-list --count B_i..I` / `I..B_i` pair) is deferred to the close of the run phase: I does not exist yet.

**Gaps.** (a) AC-QAS-015's ancestor test cannot be evaluated until the first implementation commit exists; the recorded fact is only that B_i precedes it by construction. (b) The generator source is not committed (it was a throwaway); the fixtures are reproducible from the recipe above and from the cited existing helpers, and a regenerated file must match the shasum above. (c) The status and `--auto` goldens come from the cobra command and `runAutoCycle` run through the test harness, not from the installed `moai` binary (which would read the real home and queue). (d) `captured_at` in golden 1 is a substituted constant, not a measured value. (e) A stderr line `WARN config sections directory not found, using defaults` appeared during the status generator; it is a log line, not part of any golden. (f) Lane environment: the `internal/cli` generator result under the lane environment was a failure (see above); only the scrubbed run produced goldens. (g) No `-race`, no `go test ./...`, no lint or Windows build was run: no Go source of this repository changed, so neither applies.

**Residual risk.** A later test that rebuilds a fixture even slightly differently (a different lane label, lease time, card text, or root normalization token) will mismatch the golden for a reason that is not a regression; the recipe above is the single source for the fixtures. The `one_card` status fixture was chosen by this milestone; if a later AC needs a card with a decision or contract field populated, it needs its own golden rather than an edit of this one (an edit would not be pre-change).

### M1 — Record: windows, schema 3, throttle, exhausted time (REQ-QAS-001..004; AC-QAS-001..004)

Commit `976b91458e3dea6d4f8eb7d860a1f905d43a04d2` (parent `58c01fb5f`, the M0 evidence commit; the M0 goldens commit `c2ae5236a` precedes it, so AC-QAS-015's ordering holds by construction). The commit also flips `spec.md` frontmatter `status: draft` to `status: in-progress` (the ownership matrix's `draft → in-progress` row; `updated:` was already `2026-10-02`) and carries the paragraph `Authored-By-Agent: manager-develop` immediately before `🗿 MoAI`; verified with `git log -1 --format=%b` (the line is present). Worker: a general-purpose agent in the manager-develop role, cycle_type tdd. Branch `WT-quota-aware-scheduling`, tree toplevel `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1347`, start HEAD `58c01fb5f`, toolchain `go version go1.26.8 darwin/arm64`. Files: `internal/statusline/context_usage.go`, `internal/statusline/builder.go`, `internal/config/defaults.go` (heartbeat 5m and exhaustion 100 constants), `internal/statusline/context_usage_quota_test.go` (new), `spec.md` frontmatter.

Design notes. The `writeContextUsage` signature is pinned by a static assertion in `context_usage_test.go:45`, so it stays and delegates to the new clock-injected `writeContextUsageAt(now, ..., limits)`; the builder calls the latter with `input.RateLimits` (the stdin object itself, never `data.RateLimits`, which the usage provider can fill from the home-level cache fed by a network probe). Plan debt N2 followed: the throttle truncates (`int()`) — a mutant that rounds instead (`int(x+0.5)`) is killed by subtest `a_same_truncated_bucket_within_heartbeat_writes_nothing` (observed: `62.9% shares the truncated bucket 62 with 62.2% and must not rewrite the record`, then reverted). Plan debt N6 applied: when the reset time changes and the window still reads at or above 100, the exhausted time is re-observed from the current capture (subtest `reset_time_change_still_exhausted_reobserves`). Interpretation recorded: once set, the exhausted time is kept while the reset time is unchanged even if the percentage later reads below 100 (design.md §2 read literally; REQ-QAS-004 lists only a reset change and a window leaving the record as drop causes).

**E8 — RED, step 1 (compile; acceptable only as the first step).** `go -C <card tree> test ./internal/statusline -run '^(TestQAS_AC001_RecordCarriesSuppliedWindowsOnly|TestQAS_AC002_PreviousSchemaReadsAsNoWindows|TestQAS_AC002b_WindowlessRecordBytesMatchBaseline|TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop|TestQAS_AC004_ExhaustedAtStickyUntilRollover)$' -count=1 -v` → rc=1, verbatim head:

```
# github.com/modu-ai/moai-adk/internal/statusline [github.com/modu-ai/moai-adk/internal/statusline.test]
internal/statusline/context_usage_quota_test.go:32:2: undefined: writeContextUsageAt
internal/statusline/context_usage_quota_test.go:90:10: rec.FiveHour undefined (type *SessionTelemetryRecord has no field or method FiveHour)
internal/statusline/context_usage_quota_test.go:93:10: rec.SevenDay undefined (type *SessionTelemetryRecord has no field or method SevenDay)
FAIL	github.com/modu-ai/moai-adk/internal/statusline [build failed]
```

**E8 — RED, step 2 (assertion level; EXPECTED_RED per tdd-result-contract).** After adding the record fields, the `writeContextUsageAt` signature, and a stub that ignored the windows (schema constant still 2), the same selector → rc=1; verbatim failure lines:

```
context_usage_quota_test.go:91: five-hour window = <nil>, want 62.5% resetting at 1790000000
context_usage_quota_test.go:94: seven-day window = <nil>, want 41.2% resetting at 1790500000
context_usage_quota_test.go:106: seven-day window = <nil>, want the supplied 41.2%
context_usage_quota_test.go:227: schema_version = 2, want 3
context_usage_quota_test.go:288: five-hour window = <nil>, want 63.0 written (a new truncated bucket)
context_usage_quota_test.go:313: captured_at = "2026-10-02T12:00:00Z", want the refreshed "2026-10-02T12:05:01Z"
context_usage_quota_test.go:369: exhausted time = <nil>, want the first capture "2026-10-02T12:00:00Z"
context_usage_quota_test.go:439: window = <nil>, want no exhausted time at 99%
--- FAIL: TestQAS_AC001_RecordCarriesSuppliedWindowsOnly (0.14s)
--- FAIL: TestQAS_AC002b_WindowlessRecordBytesMatchBaseline (0.00s)
--- FAIL: TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop (0.01s)
--- FAIL: TestQAS_AC004_ExhaustedAtStickyUntilRollover (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.572s
```

(Controls that were already green under the stub, as designed: `AC001/no_rate_limits`, `AC002/schema_1`, `AC002/schema_2`, `AC003/a`, `AC003/d`, `AC003/e`. `TestQAS_AC002_PreviousSchemaReadsAsNoWindows` is therefore a regression guard on the first half of AC-QAS-002; the AC's red is carried by `TestQAS_AC002b` — `schema_version = 2, want 3` — and ledger E6.)

**GREEN.** Same selector plus the three pinned guards, `-run '^(TestQAS_AC001_...|TestQAS_AC002_...|TestQAS_AC002b_...|TestQAS_AC003_...|TestQAS_AC004_...|TestReadsPreviousSchemaRecord|TestWriteContextUsage_ThrottleSkipUnchanged|TestThrottleUnaffectedByModelAndEffort)$' -count=1 -v` → rc=0, `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.874s`, 25 `--- PASS` lines (8 top-level tests, 17 subtests). Top-level PASS lines observed: `TestQAS_AC001_RecordCarriesSuppliedWindowsOnly`, `TestQAS_AC002_PreviousSchemaReadsAsNoWindows`, `TestQAS_AC002b_WindowlessRecordBytesMatchBaseline`, `TestQAS_AC003_ThrottleBucketHeartbeatAndWindowDrop`, `TestQAS_AC004_ExhaustedAtStickyUntilRollover`, `TestReadsPreviousSchemaRecord`, `TestWriteContextUsage_ThrottleSkipUnchanged`, `TestThrottleUnaffectedByModelAndEffort` — the window-less path stays green untouched. Subtests listed: AC001 `both_windows`, `seven_day_only`, `no_rate_limits`; AC002 `schema_1`, `schema_2`; AC003 `a_same_truncated_bucket_within_heartbeat_writes_nothing`, `b_next_bucket_writes`, `b2_changed_reset_time_writes`, `c_unchanged_reading_older_than_heartbeat_rewrites`, `d_window_disappearing_writes_a_record_without_it`, `e_windowless_record_is_not_heartbeat_rewritten`; AC004 `first_seen_at_100_sets_the_time`, `sticky_across_later_writes`, `reset_time_change_drops_it`, `reset_time_change_still_exhausted_reobserves`, `window_no_longer_carried_drops_it`, `first_seen_at_99_carries_no_time`. A wider anchored family (`ContextUsage|Throttle|Telemetry|RecordedModel|ModelAndEffort|ReadsPrevious|StateAnchor|SessionKey|SessionIdentity|TestQAS_`) → rc=0 `ok ... 1.640s`.

Other M1 measurements (this tree, before the M1 commit): `gofmt -l internal/statusline internal/config` listed only `internal/config/slice.go` (untouched by this card, pre-existing); `go vet ./internal/statusline/... ./internal/config/...` rc=0; `GOOS=windows GOARCH=amd64 go build ./internal/statusline/... ./internal/config/...` rc=0 (narrowed to the two changed packages to limit load); `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run ./internal/statusline/... ./internal/config/...` rc=0, `0 issues.`. Machine load average (1 min) at the compile bursts: 21.78, 23.31, 23.05, 22.47, 21.55 (all below the 25 ceiling this task set for the heavy step; `uptime` was run before each burst).

### M2 — Aggregator and configuration (REQ-QAS-005..008; AC-QAS-005, -006 statusline half, -007, -014 statusline half)

Commit `a078b4ecaab9f994f711ac2dbea1e9d7b1610827` (parent `976b91458`), separate from M1. Files: `internal/statusline/quota.go` (new aggregator), `internal/statusline/quota_test.go` (new), `internal/config/loader_quota_gate.go` (new single-block reader), `internal/config/workflow_quota_gate_test.go` (new), `internal/config/types.go` (`QuotaGateConfig`, `WorkflowConfig.QuotaGate`), `internal/config/defaults.go` (unmeasured defaults 90 / 95 / 5 / "30m", compiled skew tolerance 5m, `NewDefaultWorkflowConfig` block), `internal/config/cache.go` (schema 10 to 11 with a comment), `internal/config/testdata/shipped_key_inventory.yaml` (five W/reader rows), `internal/template/templates/.moai/config/sections/workflow.yaml` (block, `enabled: false`, comment says `unmeasured`), `.moai/config/sections/workflow.yaml` (local twin, `enabled: true`). No `internal/cli` change.

Design notes. Aggregator output carries used percentage, reset time, source capture time, and first-exhausted time per window (plan debt N6 / REQ-013); a reset window carries no percentage (never a stale high). Only `.json` record files whose mtime is within `max_age` are parsed (fixtures set mtime = capture time with `os.Chtimes`). Plan debt N5: an extra fixture `freshest_not_min_newer_record_higher` (older 40%, newer 80% fresh) was added; observed mutant results: `win.UsedPercentage > w.win.UsedPercentage` (maximum) killed by `freshest_not_max_five_hour` and `freshest_not_max_seven_day`; `<` (minimum) killed ONLY by the new N5 fixture (`--- FAIL: .../freshest_not_min_newer_record_higher`); `>=` for `>` at the age boundary killed by `age_exactly_max`; skew tolerance widened 100x killed by `future_capture_beyond_tolerance` and `future_record_cannot_win_over_a_real_one_beyond_tolerance`; each mutant was reverted (final `quota.go` re-run green below). The loader seeds `workflowFileWrapper` with `NewDefaultWorkflowConfig()` so an absent key keeps its default while an explicit `release_margin_pct: 0` (in range) is honoured. Plan debt N1: the statusline-side AC-014 sweep globs `quota*.go` (non-test) and asserts at least 1 file; the three-file floor spanning `internal/statusline` and `internal/cli` is evaluated at M5. Observation on the SPEC text: AC-QAS-005 says "seven subtests listed" but names eight (`freshest_not_max_five_hour`, `freshest_not_max_seven_day`, `per_window`, `age_exactly_max`, `age_max_plus_1s`, `future_capture_within_tolerance`, `future_capture_beyond_tolerance`, `reset_window_is_reset`); all eight are implemented, plus five more (13 subtests).

**E8 — RED, step 1 (compile).** statusline: `go -C <card tree> test ./internal/statusline -run '^(TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover|TestQAS_AC006_FailOpenOnAbsentOrUnreadable|TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly)$' -count=1 -v` → rc=1: `internal/statusline/quota_test.go:70:47: undefined: QuotaReading`, `quota_test.go:93:10: undefined: AggregateQuota`. config: `go -C <card tree> test ./internal/config -run '^TestQAS_AC007_ConfigDefaultsMirrorTemplate$' -count=1 -v` → rc=1: `workflow_quota_gate_test.go:81:10: undefined: QuotaGateSettings`, `workflow_quota_gate_test.go:90:43: NewDefaultWorkflowConfig().QuotaGate undefined (type WorkflowConfig has no field or method QuotaGate)`, `workflow_quota_gate_test.go:148:10: undefined: LoadQuotaGate`.

**E8 — RED, step 2 (assertion level; EXPECTED_RED).** With stubs that return zero values (aggregator returning an empty aggregate; loader returning the zero settings; template, local twin, defaults assignment, and cache version untouched), statusline rc=1, verbatim:

```
quota_test.go:222: five_hour: state = "", want "unknown" (reading {State: UsedPercentage:0 ResetsAt:0 CapturedAt:0001-01-01 00:00:00 +0000 UTC ExhaustedAt:0001-01-01 00:00:00 +0000 UTC})
quota_test.go:101: five_hour: state = "", want "fresh" (reading {State: UsedPercentage:0 ResetsAt:0 ...})
--- FAIL: TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly (0.01s)
--- FAIL: TestQAS_AC006_FailOpenOnAbsentOrUnreadable (0.01s)
--- FAIL: TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover (0.03s)
FAIL	github.com/modu-ai/moai-adk/internal/statusline	0.470s
```

config rc=1, verbatim:

```
workflow_quota_gate_test.go:102: DefaultQuotaGate() = {Enabled:false FiveHourHoldPct:0 SevenDayHoldPct:0 ReleaseMarginPct:0 MaxAge:0s}, want {Enabled:false FiveHourHoldPct:90 SevenDayHoldPct:95 ReleaseMarginPct:5 MaxAge:30m0s}
workflow_quota_gate_test.go:117: template carries no quota_gate block
workflow_quota_gate_test.go:141: this repository's local workflow.yaml must enable the quota gate (dogfood twin)
workflow_quota_gate_test.go:213: configCacheSchemaVersion = 10, want > 10 — Workflow gained QuotaGate without a bump
--- FAIL: TestQAS_AC007_ConfigDefaultsMirrorTemplate (0.05s)
FAIL	github.com/modu-ai/moai-adk/internal/config	0.396s
```

**GREEN.** statusline selector → rc=0 `ok  	github.com/modu-ai/moai-adk/internal/statusline	0.435s`; `--- PASS` observed for `TestQAS_AC005_AggregateFreshestBoundariesSkewAndRollover` (13 subtests), `TestQAS_AC006_FailOpenOnAbsentOrUnreadable` (5 subtests: `dir_missing`, `unparseable_record`, `unparseable_record_is_skipped_beside_a_valid_one`, `only_stale_records`, `unparseable_config_uses_default_max_age`), `TestQAS_AC014_AggregatorIsOfflineSpawnFreeAndReadOnly` (swept 1 non-test `quota*.go` file: `quota.go`). config `TestQAS_AC007_ConfigDefaultsMirrorTemplate` → rc=0 `ok ... 0.382s`, 9 `--- PASS` subtests (`defaults_equal_template`, `template_ships_off`, `template_comment_says_unmeasured`, `local_twin_enabled`, `loader_reads_a_configured_block`, `absent_or_unparseable_yields_default`, `out_of_range_yields_default`, `max_age_below_twice_heartbeat_yields_default`, `cache_schema_bumped`). Guard `go test ./internal/config -run '^TestShippedConfigKeysHaveReaders$' -count=1 -v` → rc=0 `ok ... 5.535s`, `--- PASS: TestShippedConfigKeysHaveReaders (5.39s)` with 4 passing subtests and no `quota_gate` line in its dead/unresolved diagnostics. Template guards (`internal/template`, anchored `^(TestTemplateNeutralityAudit|TestTemplateNeutralityAuditC8Preserve|TestTemplateNoInternalContentLeak|TestShippedRetiredModelKeys_ReadsTheEmbeddedTemplate|TestBranchProtectionParity|TestTemplateLearnedWorkflowBlockNeutral)$`) → rc=0 `ok ... 1.292s`, six `--- PASS` lines.

Whole-package runs (1-minute load average 13.47 immediately before the config and statusline whole-package runs, 19.95 before the later lint and coverage runs, both below 25): `go test ./internal/config -count=1` → rc=0 `ok  	github.com/modu-ai/moai-adk/internal/config	9.988s`. `go test ./internal/statusline -count=1` → FIRST run rc=1 with ONE failure, `--- FAIL: TestBuilderSetModeNormalizes (1.45s)`, whose message shows the two compared outputs differing only in live usage figures (`5H: █░░░░░░░░░ 17% (rolling)` versus `5H: ██░░░░░░░░ 21% (rolling)`); the test builds twice through `New(Options{...})`, which uses the real home usage collector, so a moving real quota between the two builds changes the string. Re-run alone: `go test ./internal/statusline -run '^TestBuilderSetModeNormalizes$' -count=1 -v` rc=0 `--- PASS`. Two later whole-package runs (`-cover`, then `-coverprofile`) → rc=0 `ok  	github.com/modu-ai/moai-adk/internal/statusline	32.429s	coverage: 92.2% of statements`. Config `-cover` → rc=0 `coverage: 82.7% of statements` (package figure; per-function figures of the new code: `loader_quota_gate.go` `DefaultQuotaGate` 100.0%, `LoadQuotaGate` 100.0%, `resolveQuotaGate` 100.0%, `defaultQuotaGateMaxAge` 75.0% (the unreachable parse-error branch); statusline `quotaWindowRecord` 100.0%, `sameQuotaWindow` 100.0%, `buildContextUsageRecord` 100.0%, `AggregateQuota` 95.7%, `writeContextUsageAt` 82.6%, `heartbeatDue` 83.3%).

Other M2 measurements: `go vet ./internal/statusline/... ./internal/config/...` rc=0; `GOOS=windows GOARCH=amd64 go build ./internal/statusline/... ./internal/config/...` rc=0 (narrowed to the changed packages); CI-version lint (`golangci-lint@v2.1.6`, same packages) rc=0 `0 issues.`; `make build` run exactly once, at load average 18.67 (below 25), after the template edit and before the M2 commit, rc=0 (last line: `go build -ldflags ... -o bin/moai ./cmd/moai`, binary build id `moai_cp/20260925_122548-1907-g976b91458-dirty`); `git status --short` after it showed no tracked file changed beyond the intended edits (the catalog hashes step printed `catalog.yaml updated successfully (13900 bytes)` without producing a diff).

**Gaps (M1 and M2).** (a) The first whole-package `./internal/statusline` run failed on the live-quota test named above; it is attributed to the real usage collector from the message content and from three later green runs, but it was not re-measured on the base tree (a branch switch is not permitted here), so "pre-existing flake" is an inference. (b) The `internal/config` package coverage figure (82.7%) is the whole-package number measured on this tree only; no pre-change figure was measured, so no delta is claimed. (c) A first attempt to edit a test file through a shell heredoc was refused by the worktree guard (command too complex to verify); the edit was redone with the Edit tool, no measurement was affected. (d) `internal/cli`, `internal/kanban`, and `internal/hook` were not run (no change there; no `go-test-cli` slot taken). (e) AC-QAS-014's three-file floor is not evaluated (N1; M5). (f) No `-race`, no `go test ./...`. (g) The statusline path was exercised through the builder and through the clock-injected writer; no real Claude Code process or real quota was involved.

**Residual risk.** A type-mismatched value in `workflow.quota_gate` (a string where an integer is expected) makes the YAML decode fail, so the single-block reader returns the shipped defaults for the whole block, `enabled` included — the same section-level fail-safe the full loader applies to every workflow key, and in the safe direction (gate off), but a mistyped threshold therefore disables a gate that was enabled rather than only reverting that one key. The exhausted-time stickiness keeps the first-observed time while the reset time is unchanged even if the percentage later reads below 100 (design.md §2 literal reading).

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

### Kickoff gate (plan→run) — autonomous form, `.claude/rules/moai/workflow/auto-semantics.md` §9.1

| Condition (§9.1) | Observed | Evidence |
|---|---|---|
| Independent plan-audit verdict is PASS | PASS, 0.87 against the Tier L threshold 0.85 (iteration 1 was FAIL 0.74; the score rose, no regression stop) | `.moai/reports/t1347/plan-audit-iter2.md` (`verdict: PASS`, `audited_sha: 7fe1ee49d62e67e86ae45e139187a679d67873ca`, read from the file by the lane) |
| Plan phase records audit-ready | `plan_status: audit-ready` | §E.1 above |
| Plan-artifact hashes unchanged since the verdict | equal | `shasum -a 256` of the five plan artifacts re-measured by the lane after the verdict; all five equal the hashes in the iteration 2 report (spec ec919773, plan 6050abc6, acceptance 803e5e69, design bf4d616e, research 99f141e1); `git diff --stat 7fe1ee49d..HEAD` empty; progress.md is not a hash subject |
| No blocker open | none | no BLOCKER or MAJOR finding in iteration 2; DO-1..DO-13 resolved (DO-3 and DO-13 leader-confirmed provisional) |
| Keep-set case (environment-impossible / operator-held / irreversible external-shared) | none applies | nothing is pushed, no PR, no external shared system touched; the operator-held items (push, release) stay with the leader |

```text
decision record: decided_by=claude-code lane-9 orchestrator (factory lane, Kickoff autonomous transition) evidence_refs=.moai/reports/t1347/plan-audit-iter2.md(verdict=PASS score=0.87 audited_sha=7fe1ee49d),.moai/specs/SPEC-QUOTA-AWARE-SCHEDULING-001/progress.md#E.1,commit 7fe1ee49d,sha256 spec=ec919773 plan=6050abc6 acceptance=803e5e69 ladder_path=gate-row plan→run Kickoff (AUTONOMOUS, auto-semantics §9.1)
```

Recorded 2026-10-02T06:59:41Z. The decision board under the moai home (auto-semantics §11) was NOT written: `moai factory decide` records only a human decider (its help text: `--decider ... (F1 accepts only human)`), and this card was dispatched directly by the leader without a factory record. This record lives here and in the card's local evidence files; a reader must treat it as self-attested (auto-semantics §10).

Whether to spend a third and final audit on the ten minor findings was put to the decision oracle (Jev): `proceed_record_debt` (probability 0.86, confidence 0.73). Proceeding keeps the audited hash valid; any edit to a plan artifact would void the verdict and the margin over the threshold is 0.02.

### Known plan debt carried into the run delegation (iteration 2 findings N1-N10, all MINOR)

The implementer may not edit the SPEC body; each item is handed over with the rule that a blocker report goes to the lane if one blocks a milestone.

| ID | Debt | Treatment in run |
|----|------|------------------|
| N1 | AC-014 minimum swept count (3) cannot pass per package before M5, while plan M2/M3 list it | AC-014's count assertion is evaluated at M5 (its own green-path sentence says so); M2/M3 run only the per-file part |
| N2 | plan.md:102 and research.md:19 say "integer-rounded"; REQ-003, design 2 and AC-003 say truncated (`int()`, context_usage.go:272) | follow the REQ/AC: truncate |
| N3 | plan.md:114 reads as if the pressure function takes a caller argument; REQ-017 and AC-017 say it takes none | follow REQ-017: the shared function takes no caller input, the lane gate applies the Claude-caller predicate to its result; AC-017 varies the environment, not an argument |
| N4 | E9's CLI half is a pipeline; AC-022's RED row measures a different directory | add a conforming single-invocation RED row in the M0 evidence if cheap, else record as a gap |
| N5 | AC-005 fixtures are one-directional (a "minimum over fresh" mutant survives); the text-mode status baseline is missing (M0 captures `--json` only) | add a fixture where the newer record is the higher one; capture a text-mode golden in M0 if the SPEC permits, else report the gap |
| N6 | REQ-005 lists only percentage and reset time; REQ-013 also needs capture time and exhausted time; REQ-004 reset-while-at-100% case unspecified | implement REQ-013's fields; on a changed reset time re-observe the exhausted time from the current capture when the window is still at or above 100% |
| N7 | short-form IDs in prose produce 14 ORPHAN lines in the traceability verb (not a coverage gap) | none (doc-only) |
| N8 | AC-018 does not say the fixture registry lacks the `legacy_workers_imported` meta row; release rule for two held windows is unstated | fixture inserts lane rows with plain SQL into a database lacking the marker; treat release as "all held windows released" |
| N9 | the `SaveFactoryRegistry` round-trip of `backend` has no AC | add one subtest alongside AC-023 |
| N10 | AC-016's hunk check reads only the hunk's old-side start line | in the implementation of the check also require the hunk end inside the allowed range |

Also carried from the iteration 2 gaps: the pinned `modernc.org/sqlite v1.57.0` was never exercised; a system `sqlite3` 3.54.0 `-readonly` open of a no-sidecar WAL database failed with `unable to open database file (14)`. At M5 start the implementer measures the real driver with a scratch program outside the tree (a cleanly closed WAL database, `mode=ro`, read one row, attempt a write, record the directory listing before and after) and reports the observed behaviour; if the quiescent case fails, the inventory returns no candidates (fail-open) and the finding goes to the lane.

### Mode evaluation

Input parameters: tier L (REQ 23, AC 23); files affected more than 15 across three packages (`internal/statusline`, `internal/cli`, `internal/kanban`, plus `internal/config` and template mirrors); domains: Go source, tests and goldens, config defaults, template mirror, one rule-adjacent doc line; language mix Go + YAML; concurrency benefit LOW — milestones M0-M6 are ordered (M0 commits the baseline goldens before any implementation commit, M1-M3 build the record, aggregator and gate, M4 the claim write, M5 the steering surfaces that read it). Agent Teams: not requested.

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | non-trivial, spans code, tests, config and template |
| serial | **yes** | coding-heavy, ordered milestones, one writer per tree; the default fallback |
| fanout | no | research is finished; the remaining work is implementation |
| sweep | no | not a uniform mechanical transform |

Decision: serial

Boundary case: the Tier L entry predicate for `manager-lead` (at least 3 milestones and at least 10 files) is met, but a lane session holds standing spawn authority at depth 1 only and agents it spawns are leaf workers that must not spawn further agents, so `manager-lead` (the Agent-carrying coordinator) cannot be used from this lane. The lane drives the milestones itself, one `manager-develop` leaf per milestone, in order; the lane does not write code.

## §J Lane Decision Log (card t1347, lane-9)

Rule (leader dispatch, operator instruction): in-flight decisions are asked of Jev (TypeSafe System One, `jev-1.13.0`) through the lane's dispatch script; question, answer and confidence are recorded here. Confidence < 0.5 or a hard-to-reverse external action goes to the leader. The oracle's answer is advisory evidence; where a fact it was given could be checked in code, the lane checked it first (DO-5: exit status 3 is `factoryNextNoCardExit`, `internal/cli/factory_card.go:181` (comment at `:179-180`), "3, so a supervising launcher can distinguish it from failure").

| ID | Question (short) | Oracle answer | Confidence | Disposition |
|----|------------------|---------------|-----------:|-------------|
| DO-1 | Stamp the hook-observed 429 turn-end into the record? | not_now | 1.00 | applied |
| DO-2 | Distinguish two Claude accounts? | reset_time_only | 0.97 | applied |
| DO-3 | Default thresholds (5h / 7d / margin / max age) | 90 / 95 / 5 / 30m (P 0.54) | 0.32 LOW | leader: accepted as provisional; all four exposed as config keys; "unmeasured defaults"; re-tune after first measurement is a listed follow-up |
| DO-4 | Template default for `workflow.quota_gate.enabled` | false in template, true in local config | 0.99 | applied |
| DO-5 | Exit status for the hold | reuse status 3 (P 0.90) | 0.81 | applied (SPEC default was 4; changed) |
| DO-6 | Gate arm (a), a card already assigned to the lane? | exempt started cards | 0.99 | applied (SPEC default was gate-all; changed) |
| DO-7 | `integration acquire` under pressure | warn_only (P 0.54 vs 0.46) | 0.07 LOW | leader: warn-only, never blocks, FINAL |
| DO-8 | Steer `--auto` / leader beyond a status block? | steer_auto (P 0.66) | 0.32 LOW | leader: REJECTED the lane's status-only provisional; steering IN SCOPE in its minimal form (recommend non-Claude lanes, Claude lanes hold with status 3, warning-only when no non-Claude lane, no forced re-dispatch); SPEC reclassified Tier M to Tier L |
| DO-9 | Stale high reading before reset | treat_unknown (P 0.93) | 0.86 | applied |
| DO-10 | Claude-lane predicate source | launch provider first, kanban backend fallback | 0.96 | applied |
| DO-11 | Record carrier | per-session record | 1.00 | applied |
| DO-12 | Where to read a lane's backend | write_backend_at_claim (P 0.78) | 0.55 | applied after the lane required feasibility verification (U18); verified by manager-spec and re-checked by the lane: `workers.backend TEXT NOT NULL DEFAULT ''` at `internal/homestate/factory.go:27`; claim inserts at `internal/kanban/factory_slots.go:106` and `:329` omit it |

Leader messages: plan commit and the three low-confidence items reported to the leader; verdict received for DO-3 / DO-7 / DO-8 (above). No hard-to-reverse external action has been taken (no push, PR or delete).

Plan artifacts: `400b5f986` (initial), `fdfb2d0c8` (oracle resolutions + leader verdict, Tier L), `ca57dc350` (DO-12). Tier L: REQ 23 of 25, AC 23 of 25.

### Plan-audit iteration 1 (FAIL 0.74) — decisions taken while closing D1-D24

Report `.moai/reports/t1347/plan-audit-iter1.md` is local-only by operator directive (`.gitignore` `.moai/reports/*`, `.moai/docs/audit-artifact-convention.md` Committing section); it is not committed. The leader asked for a commit, the lane declined on that directive, and the leader withdrew the request after reading both sources and recomputing the sha256.

| ID | Question (short) | Oracle answer | Confidence | Disposition |
|----|------------------|---------------|-----------:|-------------|
| D1 | `factory status` quota block vs byte-identical output when pressure is off | block only when the gate is enabled (P 0.94) | 0.88 | applied (REQ-013, REQ-022) |
| D3 | A held wait whose reading ages out mid-wait | release when unknown (P 0.80) | 0.59 | applied (REQ-011, AC-010 `unknown_mid_wait_releases`) |
| D7 | Lane inventory read of the registry | genuinely read-only open (P 0.82) | 0.63 | applied after feasibility was measured by manager-spec on `modernc.org/sqlite v1.57.0` (plan.md D.1: absent file errors without creating anything; `mode=ro` reads a closed WAL database and sees live WAL rows; `immutable=1` rejected because it missed uncheckpointed rows); the lane re-read the existing `mode=ro` pattern the SPEC reuses, `openSQLiteReadOnly` at `internal/discovery/factory_discovery.go:353-362` (fail-open: a WAL database whose recovery needs the write lock contributes no candidates). The lane did not re-run the scratch measurements |
| D12 | Surface the first-exhausted time (otherwise dead data) | surface (P 0.52 vs 0.48) | 0.03 LOW | leader: CONFIRMED surface in the status block (DO-13) |

Revision commits: `0e12b4cfa` (D1-D24 closed, spec 0.5.0), `be14684ab` (drops a suffixed AC identifier that made the commit guard count 24). Lane re-check: tree clean, `moai spec lint --strict` no findings, REQ 23 and AC 23 by grep.

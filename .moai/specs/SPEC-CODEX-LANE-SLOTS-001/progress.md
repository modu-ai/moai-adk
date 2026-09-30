# SPEC-CODEX-LANE-SLOTS-001 — Progress

Status: draft (plan-phase artifacts authored 2026-09-30, card t1378).

## Premise Verification (plan-phase, this tree `d194083fb`)

Verdict: **premise HOLDS — root cause confirmed.** Evidence observed in this run, this worktree:

| # | Claim | Evidence (file:line + observed) |
|---|-------|-------------------------------|
| V1 | Bounded auto scan ignores `maxCanonical`; exact error string | `internal/kanban/factory_slots.go:246-256` — `if maxSlots > 0 { if auto { for candidate := 1; candidate <= maxSlots; candidate++ ... if n == 0 { return claim, fmt.Errorf("factory run %s has no free lane slots in 1..%d", runID, maxSlots) } ... } }`. `maxCanonical` computed at lines 220-235 but used only in the unbounded branch (line 262: `n = maxCanonical + 1`). |
| V2 | Codex join passes the 1-slot bound | `internal/cli/codex_factory.go:140-141` — `kanban.ClaimFactoryLaneWithin(root, entry.LaneLabel, entry.LaneRole, os.Getpid(), os.Getenv(config.EnvMoaiKanbanID), config.DefaultFactoryLeaderLanes, factoryProcessAlive)`. `config.DefaultFactoryLeaderLanes = 1` at `internal/config/defaults.go:648`. |
| V3 | Claude join is unbounded (asymmetry) | `internal/cli/factory.go:761` — `kanban.ClaimFactoryLane(root, label, auto, os.Getpid(), runID, factoryProcessAlive)` (maxSlots=0 → grows via `maxCanonical+1`). |
| V4 | The bound arrived with card t1294 in the rc.16→rc.23 window | Batch log `git log a8a9b9376..d194083fb` contains `5ee35b2f5 fix(factory): bound Codex lane claims to run slots (t1294)`. |
| V5 | Legacy machinery present and tested | `FactoryLegacyRunError` `internal/kanban/factory_slots.go:128-136`, fired at 238-240; `IsLegacyFactoryLabel` / `factoryLabelNumber` at `internal/kanban/bootstrap.go:336-356`; tests `role_naming_m1_test.go:76/109/126`, `factory_slots_test.go:75-139`. |
| V6 | ① Claim path has no wait/retry | `claimFactoryLane` (factory_slots.go:167-285) contains no sleep/retry/deadline; failure at line 255 is immediate. Long-launch delay must be pre-claim: `enterCodexFactory` → `enterFactoryLaneRun` (factory.go:492) → `enterSelectedFactoryRun` (factory.go:417) → codex pre-exec init → `syscall.Exec` (`internal/cli/codex_direct_posix.go:53`). |
| V7 | ② No worktree scan in the launch path | Greps `filepath.Walk|os.ReadDir|filepath.Glob` over `internal/cli/codex_launcher.go`, `codex_init.go`, `session*.go`: no hits. |
| V8 | ③ Hook-trust message is a wiring-pass emit | `ReTrustGuidance` at `internal/codexwiring/codexwiring.go:74`; launcher execs and does not gate on hook approval. |
| V9 | ④ Owner probe is per-run-row | `moai factory runs` = `OpenFactory` + `ClassifyRuns` (`internal/cli/factory_handoff_recover.go:79-101`); the probe is `platformProcessFingerprint(pid)` — per-pid, three build-tagged variants: `process_fingerprint_unix.go:11` (`!windows && !darwin`, `ps` subprocess), `process_fingerprint_darwin.go:11` (darwin, `unix.SysctlKinfoProc` sysctl — no subprocess), `process_fingerprint_windows.go:10` (windows); seam `OwnerClassifier` at `internal/homestate/factory_run_retire.go:44`. (Iteration-1 correction: the earlier citation named only the unix `ps` variant, which is build-excluded on this darwin tree — the measured 3.0s user / 1.86s sys cost cannot be attributed to `ps` alone.) |

SPEC ID self-check: `ID="SPEC-CODEX-LANE-SLOTS-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`. Uniqueness: no `SPEC-CODEX-LANE-SLOTS-*` in `.moai/specs/` (catalog listing observed, 1002 entries).

## Plan-audit Iterations

- **Iteration 1 — FAIL 0.75 vs Tier M 0.80** (report: `.moai/reports/t1378/plan-audit.md`). Fixes applied 2026-09-30 on tree `850aefca9`, no commits: D1 (AC-009/AC-010 added for REQ-007/008/012), D2 (RED-now evidence ledger in acceptance.md — RED-1/RED-2/RED-3/RED-4 executed on the pre-implementation tree at `850aefca9`, scratch probe deleted after capture; regression-guard classification for AC-002/003/004/005/008), D3 (POSIX/Windows exec-split clause at both `syscall.Exec` mentions), D4 (V9, plan §B.4/M3/§E.4 corrected to the three platform variants), D5 (AC-006 package pinned to `internal/cli`), D6 (REQ-012 trigger made operator-configurable with the default in `internal/config/defaults.go`), D9 (`related_specs` dropped from frontmatter; §G body cross-references are the carrier).
- **Iteration 2 — FAIL narrow 0.94, threshold cleared, MP-8 RED re-executability** (report: `.moai/reports/t1378/plan-audit-iter2.md`). Fixes applied 2026-09-30 on tree `4a7a2b0a4`, no commits: D1 (probe committed as `redprobe_t1378_main.go` in this directory with `//go:build ignore` — `go run <file>` file-mode verified by execution; RED-1/RED-2 re-executed at `4a7a2b0a4`, outputs byte-identical to the `850aefca9` originals; RED-4 de-piped to an explicit file list and re-executed; probe documented synthetic-root-only), D2 (Command lines restored on AC-002/003/004/005), D3 (plan §F M2 AC list now includes AC-009), D4/D5 (all `-run` patterns in acceptance.md anchored with `^...$` exact-name regexes to clear the unanchored-run lint warnings).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-01
plan_audit_verdict: PASS (score 1.00, Tier M threshold 0.80; trajectory 0.75 → 0.94 → 1.00 over 3 iterations, no STOP signal)
plan_audit_report: .moai/reports/t1378/plan-audit-iter3.md (full stream: plan-audit.md, plan-audit-iter2.md, plan-audit-iter3.md)
plan_audit_sha: bfdaf564cf8b32755921123780a1485a92bfc980
plan_artifact_hash: 314f8659a13eb14b74678fd343975a8bfdfebea43fca758ac2c9d5694dc21e
recorded_by: lane orchestrator (verdict landed after manager-spec's final fix turn; audit-ready signal derived from the iteration-3 verdict)

## §E.2 Run-phase Evidence

Status: run complete (2026-10-01, manager-develop cycle_type=tdd). Branch: `WT-codex-lane-slots-run`, branched exactly at the card tip `b8f437bae` in the run session's isolated worktree; commits are direct ancestors of the card branch, integrable with `git merge --ff-only WT-codex-lane-slots-run` (the card branch itself stays owned by the lane session — one writer).

### Commits (per-milestone Conventional Commits)

| M | SHA | Subject |
|---|-----|---------|
| M1 | `3d02f66fa` | feat: record declared lane capacity on the run (runs.lane_capacity, schema v5) |
| M2 | `881580165` | feat: capacity-aware growth of the bounded lane claim |
| M3 | `e18f2d193` | feat: per-step pre-exec timing and batched owner probe |
| M4 | _this commit_ | test: bounded-path legacy parity + run-phase evidence |

### Deviations (declared)

- **spec.md frontmatter `status: draft → in-progress` deferred from M1 to this M-final commit** — the ownership-matrix trigger is the first run-phase commit; the flip lands here instead. Declared rather than silently timed; the M1..M3 code commits are unaffected.
- **M1's whole-kanban re-measurement ran with M2's RED tests present in the tree** (written next-milestone, not yet committed) and failed only in those four new tests; the M2-commit-time whole-package run (249.271s, green) and this card's final AC-008 run (231.602s, green) subsume M1's regression coverage. No M1-only whole-package run exists.
- **The REQ-014 measurement briefly migrated the shared live store** (`~/.moai/db/moai-adk-go-1bd3d038/factory/factory.db`, project-keyed across worktree lanes) to schema v5 — an in-place side effect of running the post-M1 binary against it. Restored to v4 the same session (`ALTER TABLE runs DROP COLUMN lane_capacity` + meta 5→4; base-binary listing re-verified). All recorded measurements after that point run on a scratch-rooted copy (`/tmp/t1378-perf-root`), never the live store. Concurrent lanes running pre-v5 binaries were exposed for the minutes between; no failure was reported.

### AC matrix (E1)

Every row: command executed in this run, output verbatim, HEAD attributed. Baseline green at pre-flight: `go test -timeout 30m ./internal/kanban/...` → `ok github.com/modu-ai/moai-adk/internal/kanban 193.031s` at `b8f437bae`.

| AC | Verdict | Command + verbatim evidence (HEAD) |
|----|---------|------------------------------------|
| AC-001 | **PASS** (M2) | RED-first at `b8f437bae`+M1: `factory_slots_capacity_test.go:54: bounded claim on capacity-open run: factory run tm3yoq has no free lane slots in 1..1`; GREEN: `go test ./internal/kanban/ -run '^TestClaimFactoryLaneWithinGrowsCapacityOpen$'` → `--- PASS: TestClaimFactoryLaneWithinGrowsCapacityOpen (4.36s)` (post-M2). Probe flip (M2 green side): `RED-1 claude-shape claim: lane-1` / `RED-1 codex-shape claim error: <nil>` / `RED-1 codex-shape claim label: lane-2` |
| AC-002 | **PASS** (M2) | `TestClaimFactoryLaneWithinBounds` `--- PASS (4.64s)` (legacy bounds pin unmodified) + new subtest `explicit_request_never_grows` `--- PASS (1.10s)`: explicit lane-9 vs 1..1 on a capacity-open run refuses with `outside the allowed slots 1..1` |
| AC-003 | **PASS** (M2) | `TestClaimFactoryLaneWithinExplicitCapacityFull` `--- PASS (3.76s)`; `full_declared_run_refuses` → `has no free lane slots in 1..1`; `the_recorded_count_overrides_the_launcher-side_bound` (bound 3 vs recorded 1 → refuses 1..1) |
| AC-004 | **PASS** (M4) | `go test ./internal/kanban/ -run '^TestClaimFactoryLaneWithinRefusesLiveLegacy$'` → `--- PASS: TestClaimFactoryLaneWithinRefusesLiveLegacy (1.18s)` (bounded path, message names worker-1 / legacyrun1 / `moai factory runs --retire legacyrun1`; registry untouched after refusal) |
| AC-005 | **PASS** (M4) | legacy trio green unmodified: `TestClaimFactoryWorkerRefusesLiveLegacyClaim`, `TestClaimFactoryWorkerDeadLegacyClaimIsStale`, `TestClaimFactoryWorkerIgnoresLegacyClaimOfNoRun`, `TestIsLegacyLeadLabelDetection` — all `--- PASS` in the M2-adjacent selector run; the unbounded trio's intent untouched |
| AC-006 | **PASS** (M1) | RED-first: compile-fail verbatim (`undefined: homestate.LaneCapacityDerived` / `db.RunLaneCapacity undefined` / `undefined: factoryJoinLaneBound`, recorded before the M1 change); GREEN: `go test ./internal/cli/ -run '^TestRecordFactoryRunStartCapacity$'` → `--- PASS: TestRecordFactoryRunStartCapacity (1.73s)`; RED-2 flip: `RED-2 runs table columns: run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,lane_capacity,created_at,updated_at` |
| AC-007 | **PASS** (M3) | RED-first: compile-fail verbatim (`unknown field BatchProbe` / `undefined: ProcessIdentity` / `undefined: BatchProbeProcessIdentity`); GREEN: `--- PASS: TestClassifyRunsBatchesOwnerProbe (0.28s)` (1 batch invocation for 4 rows/3 pids, deduped `[411 412 413]`, both classification arms fed) + platform pin `--- PASS: TestBatchProbeResolvesLiveProcessIdentityDarwin` |
| AC-008 | **PASS** (M4) | `go test -timeout 30m ./internal/kanban/... -count=1` → `ok github.com/modu-ai/moai-adk/internal/kanban 231.602s` (final tree); `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet` on touched packages clean |
| AC-009 | **PASS** (M2) | `TestClaimFactoryLaneWithinSharedPoolDistinctNumbers` `--- PASS (1.40s)` — claude-shape lane-1 + codex-shape bounded lane-2, one registry, distinct numbers; RED (same refusal as AC-001, shared ledger RED-1) |
| AC-010 | **PASS** (M3) | RED-first: compile-fail verbatim (`undefined: factoryLaunchTiming` / `undefined: factoryStepCodexPreInit` …); GREEN: `--- PASS: TestCodexLaneLaunchTimingNamesSteps` (+ `QuietUnderThreshold`, `NilSafe`, `MeasuresElapsed`, all `--- PASS`) — the report carries one line per recorded step naming codex pre-exec init / join gate / active-run resolution / lane claim / exec handoff |

### REQ-014 measurement pair (AC-007 evidence)

Method: the live store was first found to be shared project state (see Deviations), so the pair runs on a scratch-rooted `.backup` copy of the same 10-row store in v4 shape; the after binary's first run absorbs the one-time v4→v5 migration.

```
$ cd /tmp/t1378-perf-root
$ time /tmp/moai-t1378-before factory runs   # built at b8f437bae
0.83s user 0.32s system 79% cpu 1.450 total
$ time /tmp/moai-t1378-before factory runs
0.88s user 0.31s system 91% cpu 1.311 total
$ time /tmp/moai-t1378-after factory runs    # working tree at M3; run 1 absorbs v4→v5 migration
0.88s user 0.35s system 88% cpu 1.403 total
$ time /tmp/moai-t1378-after factory runs
0.92s user 0.33s system 92% cpu 1.353 total
$ time /tmp/moai-t1378-after factory runs
0.94s user 0.36s system 94% cpu 1.379 total
```

Attribution: the pair is flat on this store — the probe bound is pinned mechanically by `TestClassifyRunsBatchesOwnerProbe` (1 invocation per listing), and the listing's wall-time here is dominated by non-probe work (open/schema/boot-proof/broker reads), which is exactly the attribution plan §B.4 stated. The operator's 3.0s-user incident store is larger; its measurement is a follow-up card candidate, not evidence recorded here. An invalid earlier comparison (base binary against the migrated v5 store) was discarded on discovery: the base binary fast-fails with `Unsupported factory schema version "5"` and its apparent speed was the error path.

### M3 split decision

NOT taken — the batched probe landed in-card (seam + three platform variants + instrumented test + measurement), within a single-function change per platform variant on the existing `OwnerClassifier`/`ReconcileOptions` seam.

### Diagnosis findings ①–④ — run-phase resolution

- **① (claim path has no wait/retry)**: CONFIRMED and preserved — the claim still refuses immediately (plan §G forbids retry); the long-launch diagnosis is now measurable: REQ-012's per-step report names the phase that is slow on the next slow codex lane launch. The rc.23 delta attribution needs one slow-launch observation to populate and is explicitly left to that report, not asserted here.
- **② (no worktree scan in the launch path)**: CONFIRMED absent (plan-phase greps stand); this card added no scanning (Out of Scope held).
- **③ (hook-trust is a wiring-pass emit)**: Out of Scope held (spec §F); no launcher gating added.
- **④ (owner probe per run row)**: FIXED by M3 — the per-listing invocation bound is test-pinned; wall-time attribution recorded above.

### Pre-existing red (outside this card, unchanged by it)

`TestCharacterize_AuditPinPrecedenceAndBackendDefault` fails identically on the base tree: the base `defaults.go` already bakes the codex audit pin `{gpt-6.1-sol, high}` (card t1368 / SPEC-MODEL-MATRIX-UPDATE-001) while the characterization test expects the zero value for a bare root. Evidence: `git show b8f437bae:internal/config/defaults.go` carries the pin at line ~1206; `git diff --stat b8f437bae -- internal/config/defaults.go internal/cli/retained_model_surfaces_char_test.go` is empty; observed `retained_model_surfaces_char_test.go:153: codex audit without pin = {Model:gpt-6.1-sol Effort:high}, want the zero value`. Not repaired here (out of scope); the full cli-suite verdict owner remains the CI run on the integration branch, PENDING at report time.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-01
run_commit_sha: e18f2d193  # last code commit (M3); the M4 docs/test commit lands after this signal
run_status: complete
ac_pass_count: 10
ac_fail_count: 0
preserve_list_post_run_count: 5  # TestClaimFactoryLaneWithinBounds, TestClaimFactoryLaneWithinConcurrentOneSlot, role_naming_m1 trio (3) — passing unmodified in intent
l44_pre_commit_fetch: not-run  # session-private worktree; no shared-checkout commit or push from this session
l44_post_push_fetch: not-run  # no push (lane protocol: push is the leader's)
new_warnings_or_lints_introduced: 0  # golangci-lint v2.1.6 (CI pin): 0 issues pre and post on kanban/cli/homestate/config; go vet clean; gofmt clean
cross_platform_build:
  native: pass  # go build ./...
  windows: pass  # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files: 40  # 37 through M3 + legacy-bounded test + progress.md + spec.md frontmatter
m1_to_mN_commit_strategy: per-milestone Conventional Commits M1→M2→M3→M4; spec frontmatter flip deferred from M1 to M4 (declared deviation)
```


## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

### Kickoff record (operator-held gate)

- The card dispatch (leader → lane, 2026-09-30) marked the plan→run Kickoff for OPERATOR DIRECT ANSWER — the keep-set operator form, not the autonomous transition.
- Operator answered via AskUserQuestion in the lane on 2026-10-01: **착수 (proceed to run phase)** — the recommended option; alternatives offered were SPEC 재검토 (revise) and 중단 (abort).
- Gate evidence at ask time: plan-audit iter3 PASS 1.00 (≥ 0.80 Tier M), RED citations 4/4 re-executed and reproduced at `bfdaf564c`, tree-sourced spec lint 0 errors / 0 warnings, artifact hash `314f8659a13eb14b74678fd343975a8bfdfebea43fca758ac2c9d5694dc21e` fixed since the verdict.

### Input parameters

- tier: M · scope: ~9-10 files (internal/kanban claim+tests, internal/cli factory/codex_factory/launcher, internal/homestate store+probe variants, internal/config defaults) · domain count: 4 · file language mix: 100% Go · concurrency benefit: LOW (coding-heavy, dependency-ordered milestones).

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file semantic change, not a typo/one-liner |
| fanout | no | coding-heavy work (Anthropic coding-task parallelism caveat); milestones are dependency-ordered |
| sweep | no | not a mechanical uniform transform; Kickoff just passed but the transform rule is per-milestone semantic |
| agent-team | no | not operator-requested; experimental surface stays unselected |
| serial | **yes** | one manager-develop per milestone, M1→M2→M3→M4 |

Decision: serial

Justification: the four milestones are strictly dependency-ordered (M1 capacity record feeds M2 claim semantics; M3 instrumentation is independent but same-tree; M4 parity verifies the whole), the work is coding-heavy Go across three packages, and a single manager-develop with the Section A-E delegation template carries the full context cheaper than any fan-out. Boundary case: scope estimate ~9-10 files sits at the fanout threshold ±1 — the tie-breaker resolves to the simpler mode (serial), consistent with the coding-heavy override.

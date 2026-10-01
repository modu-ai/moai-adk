# SPEC-CODEX-DEBUG-MODE-001 — Progress

Status: in-progress (run phase executed 2026-10-01 → 2026-10-02, card t1380; plan-phase artifacts authored 2026-10-01). Premise table V1–V12 measured on tree `16424e9b4`; the RED-now observations in `acceptance.md` § Evidence Ledger were re-executed on tree `49a42c2fc` (the lane's artifact commit moved HEAD between the two measurement passes; none of the cited source files changed — anchors below re-verified at `49a42c2fc`).

## Premise Verification (plan-phase, this tree `16424e9b4`, branch `WT-codex-debug-mode`)

Verdict: **all four card premises HOLD; one external-support question closed as honestly unverifiable with a recorded fallback.** Evidence observed in this run, this worktree:

| # | Claim | Evidence (file:line + observed) |
|---|-------|-------------------------------|
| V1 | Codex launcher flag surface today is `--spawn` only | `internal/cli/codex_launcher.go:600` (`var codexCmd = &cobra.Command{`; the card's `:601` anchor names this command), `:641` (`DisableFlagParsing: true` — auditor-verified anchor), `:652` — the ONLY flag registration: `codexCmd.Flags().Bool("spawn", false, ...)` (re-verified at `49a42c2fc`); `:70-72` readout-refusal discipline precedent. Grep for `-d`/`--debug` over the file: zero hits, exit 1 (positive control `codexSpawnReadoutDiag` → `:70/:72/:721`, exit 0). |
| V2 | cc/glm have no `-d` handling — passthrough by structure | `grep '"-d"\|--debug\|debug'` over `internal/cli/cc.go` and `glm.go`: zero hits. Both are `DisableFlagParsing: true` (`cc.go:133`, `glm.go:137`); `runClaudeEntry` (`cc.go:146`) consumes only `--help/-h`, `--spawn` (`:164`), `-p` profile (`:168`), and the `-k/-f` entry shapes (`:175-179`) — leftover args flow into the child invocation, so `-d` reaches `claude` verbatim today. |
| V3 | claude CLI natively supports `-d` | `claude --help` (local, `/Users/goos/.local/bin/claude`): `-d, --debug [filter]  Enable debug mode with optional category` plus `--debug-file <path>`. |
| V4 | glm binary unavailable locally | `command -v glm` → no output. Child-native `-d` support UNVERIFIED for glm — deliberately non-load-bearing (observe-only design). |
| V5 | codex CLI rejects `-d`; no `--debug` flag; `debug` is a subcommand group | codex-cli 0.159.3 at `/Users/goos/.local/bin/codex`. `codex -d` → `error: unexpected argument '-d' found` (verbatim). `codex --help`: no `--debug` option; `debug` listed as a command ("Debugging tools"). `codex debug --help`: subcommands models / app-server / prompt-input only. `codex exec --help` grep for debug/log/RUST: no logging flag. |
| V6 | RUST_LOG: absent from the repo; undocumented for codex locally; literal present in the installed binary | `grep -rn RUST_LOG internal/ pkg/ cmd/` → exit 1 (zero hits). `strings /Users/goos/.local/bin/codex | grep -m3 RUST_LOG` → one concatenated literal hit (`=ErrorRUST_LOGcalled ...` — presence, not wiring proof). Upstream `docs/config.md` fetch → returned non-codex content (inconclusive). Resolution: `RUST_LOG=debug` injection recorded as a best-effort tracing-standard fallback; the launcher-side trace is the primary surface. |
| V7 | Launch-path phases to trace (slot claim, worktree, env injection) — anchors re-verified at `49a42c2fc` | Binary resolve `codex_launcher.go:1002`; timing gate (lane-only) `:1014-1015`; project root `:1023-1027`; init gate `:1033` (`codexInitOfferGate`); local instructions `:1034`; worktree resolve + writer check `:1054-1055` (`resolveCodexWorktreeDir` `:428`); factory entry: join gate `codex_factory.go:135-137` → `enterFactoryLaneRun`, lane claim `:148-151` → `kanban.ClaimFactoryLaneWithin`; env assembly `codexChildEnv` `codex_launcher.go:575-593` (drop-list + `CODEX_HOME` last-wins append) + `codexFactoryEnv` `codex_factory.go:206-227`; anchor lock + exec handoff `:1084` (`endHandoff`), threshold report sites `:1093/:1105`; direct door `codex_direct_posix.go:25-58` (`syscall.Exec` `:53`). |
| V8 | t1378 join reads recorded capacity in factory_slots.go | The card named no directory — the file is `internal/kanban/factory_slots.go` (NOT `internal/cli/`): `ClaimFactoryLaneWithin` `:159`, the recorded-capacity authority read inside the claim transaction `:170` (MX:ANCHOR), the no-recorded-capacity launcher-bound fallback `:280`. Launcher-side bound: `factoryJoinLaneBound` `internal/cli/codex_factory.go:180-186`. |
| V9 | Logging precedents | Level: `MOAI_LOG_LEVEL`/`MOAI_LOG_FORMAT` `internal/config/envkeys.go:40,43`, consumed by the slog default handler `internal/cli/logging.go:43-55` (default `warn`, `:22`). Destination: non-hook subcommands → stderr (`logging.go:78-83`); hook path → lazy-open `.moai/logs/` file sink (`:80`). `.moai/logs/` writer patterns exist for BACKGROUND processes: `internal/config/log.go:22-24` (`config.log`, best-effort append), `internal/codexadapter/diagnostics.go:16` (`.moai/logs/codex-adapter.jsonl`). Interactive launchers print diagnostics to `cmd.ErrOrStderr()` (`codex_launcher.go:1004, 1049, 1093, 1105`). |
| V10 | t1378 timing collector is the composition seam | `internal/cli/factory_launch_timing.go`: nil-safe (`:46-48`, "the cc/glm twins ... pass nil"), records every `begin()` step regardless of threshold (`:58-70`), threshold gates only `reportSlow` (`:87-101`); pre-seam discipline MX:NOTE `:85`; threshold default `internal/config/defaults.go:672` (2s), override key `internal/config/envkeys.go:297` (`MOAI_FACTORY_SLOW_LAUNCH_MS`). Wired lane-only at `codex_launcher.go:1012-1014`. |
| V11 | No dedicated debug env/flag exists anywhere in the launchers | `grep -rn RUST_LOG` (V6, zero) + V1/V2 greps: no `-d`/`--debug`/debug-env surface on any of the three launchers today. |
| V12 | Head/tail scoping precedent for the token scan | `runCodex` `--help` scan breaks at `--` (`codex_launcher.go:663-667`); `splitCodexDashDash` `:669`; `stripSpawnFlag` `:668` — the debug-token scan mirrors this pre-`--` scoping. |

SPEC ID self-check: `ID="SPEC-CODEX-DEBUG-MODE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL` → `PASS`. Uniqueness: `grep -rn SPEC-CODEX-DEBUG .moai/specs/` → zero hits (catalog scan observed this run).

## Plan-audit Iterations

- **Iteration 1 — FAIL 0.81 vs Tier M 0.80 (aggregate cleared the threshold; the MP-8 must-pass firewall forced FAIL on RED-now adoption)** (report: `.moai/reports/t1380/plan-audit.md`). Fixes applied 2026-10-01 on tree `49a42c2fc`, no commits: D1 (all 10 release-blocking ACs now carry RED-now cells with Command + fenced verbatim output + Exit + measured SHA — RED-1 invocation probe `go run ./cmd/moai codex -d status` at exit 1, RED-2/RED-3 greps with exit codes and positive controls per t1133, RED-4..RED-9 anchored test-absence selectors with vacuous-green red-reasons; single-invocation re-executability stated per cell, release-blocking retained), D2 (AC-015 added — pre-seam ordering on both doors, stub-seam-relative assertion, release-blocking; §C mapping REQ-009 → AC-015), D3 (citation convention fixed: `t1378 REQ-012` disambiguated from this SPEC's §B.4 REQ-012 across spec/plan/acceptance), D4 (uniform ledger carrier form: Command / Verbatim / Exit / Measured-at / Positive-control / Red-reason), D5 (chained Verify commands split into separately recordable parts — AC-013 parts 1-2, AC-014 split), D6 (V1/V7 line anchors re-verified and corrected at `49a42c2fc`: `:641` DisableFlagParsing, `:1033` init gate, `:1054-1055` worktree, `:1014-1015` timing wiring). No probe file cited by the ledger (all cells are invocation/grep/selector forms), so the t1378 committed-probe-file rule has no target here.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-01
plan_audit_verdict: PASS (score 1.00, Tier M threshold 0.80; trajectory 0.81 → 1.00 over 2 iterations, no STOP signal; iteration-1 FAIL was the MP-8 RED-now firewall, score-independent)
plan_audit_report: .moai/reports/t1380/plan-audit-iter2.md (full stream: plan-audit.md, plan-audit-iter2.md)
plan_audit_sha: 32caf5820 (verdict re-executed evidence at this tree; pin validity proven via git diff-tree — internal/ byte-identical to the pinned 49a42c2fc)
plan_artifact_hash: 257c6ac68642441dfb6b9218a63b71c078dfa10964baca303e42c6dd70425576
hash_note: computed via moai audit-cache ComputeHash on the SPEC dir AFTER the verdict; spec.md/plan.md/acceptance.md bytes unchanged since 32caf5820 (progress.md is not in the hashed set)
recorded_by: lane orchestrator (verdict landed after manager-spec's final fix turn; audit-ready signal derived from the iteration-2 verdict)

## §E.2 Run-phase Evidence

Run phase executed 2026-10-01 → 2026-10-02 by manager-develop (cycle_type=tdd, serial), branch `WT-codex-debug-mode-run` at the card tip `152adf3bb` (the runtime isolates spawned agents onto a develop-based worktree, so the run branch was recreated at the card tip per the dispatch note; the lane fast-forwards `WT-codex-debug-mode`). Commits, oldest first: M1 `e22be7956` (flag surface + spec.md status flip), M2 `d00210bf8` (trace engine + child-env linkage), M3 `275ebb21a` (cc/glm uniformity), M4 `4e2022868` (composition pins), M5 sweep `2a4950d47` (gofmt + env-name constant), evidence commit = this one. No push, no PR (lane-owned).

### AC matrix (all 15 PASS)

Verification form: one env-scrubbed anchored run of the 16-test debug battery plus the 4 frozen timing tests — `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=1 -timeout 15m -run '^(TestCodexDebugTokenStrippedFromChildArgs|…|TestFactoryLaunchTimingMeasuresElapsed)$' ./internal/cli/` → `ok github.com/modu-ai/moai-adk/internal/cli 21.630s` (verbatim, 2026-10-02, tree `2a4950d47`). Per-AC anchors:

| AC | Test / evidence | Result |
|----|-----------------|--------|
| AC-001 | TestCodexDebugTokenStrippedFromChildArgs (`cli -d` row) | PASS |
| AC-002 | same test, `-w wt-a --debug` row | PASS |
| AC-003 | TestCodexDebugTokenAfterDashDashForwarded (3 rows incl. empty head) | PASS |
| AC-004 | TestCodexDebugTokenRefusedOnReadoutVerb (named constant + exit 1 + 0 launches) | PASS |
| AC-005 | TestCCDebugTokenObserveOnly (child keeps `-d`; trace lines present) | PASS |
| AC-006 | TestGLMDebugTokenObserveOnly | PASS |
| AC-007 | TestThreeRunnerDebugUniformityMatrix (6 combos: linkage + same prefix) | PASS |
| AC-008 | TestCodexDebugTraceStepLines (6 step lines exactly; lane/worktree absent) | PASS |
| AC-009 | TestCodexDebugTraceEnvKeysOnly (12 pinned sentinels; key named, no value) | PASS |
| AC-010 | TestCodexDebugTraceWorktreeEntry (resolved dir + writer-check ok + anchor-lock ok) | PASS |
| AC-011 | TestCodexDebugRustLogInjection (injected when absent / `info` preserved) | PASS |
| AC-012 | TestCodexDebugSupersedesLaunchThreshold (1-hour threshold; dump still prints; label detail) | PASS |
| AC-013 | part 1: TestCodexDebugOffKeepsTimingReportFrozen (fast silent / slow exact report, no debug steps); part 2: the 4 frozen tests, file unmodified (`git diff 152adf3bb --stat -- internal/cli/factory_launch_timing_test.go` → empty) | PASS |
| AC-014 | TestCodexDebugTraceIgnoresLogLevel (MOAI_LOG_LEVEL=error; trace prints) | PASS |
| AC-015 | TestCodexDebugTracePrecedesExecSeam (all prefix lines precede the seam sentinel in one buffer) | PASS |

### Run-phase RED evidence (E8)

- M1 RED (captured pre-implementation, tree `152adf3bb`): all five `TestCodexDebugTokenStrippedFromChildArgs` rows failed at `runCodex(...)`: exit-1 with the generic usage line `unknown verb - usage: moai codex …`; `TestCodexDebugScanNeverConsumesWorktreeValue` and `TestCodexDebugTokenRefusedOnReadoutVerb` failed with `stderr = "unknown verb - …", want … "-w requires an existing worktree name or path" / "-d/--debug applies to the launch verbs only …"`. The post-`--` pin (AC-003) passed pre-implementation — it is a regression-guard (ledger has no RED cell for it).
- M2 RED (captured pre-implementation): `TestCodexDebugTraceStepLines` — all six steps `produced 0 trace lines`; `TestCodexDebugTraceEnvKeysOnly` — `trace does not name the lane key MOAI_KANBAN_LEAD_ADDR`; `TestCodexDebugTraceWorktreeEntry` — `worktree materialization produced 0 trace lines`; `TestCodexDebugRustLogInjection/injected_when_absent` — `child RUST_LOG = [], want … debug`; `TestCodexDebugTracePrecedesExecSeam` — `only 0 trace lines preceded the seam`; `TestCodexDebugTraceIgnoresLogLevel` — all three steps suppressed. The operator-preservation arm (`RUST_LOG=info` kept) passed pre-implementation — regression-guard arm.
- M3 RED (captured pre-implementation): `TestCCDebugTokenObserveOnly`, `TestGLMDebugTokenObserveOnly`, and the four cc/glm matrix legs all failed with `emitted no debug trace lines under the "moai-launcher-debug:" prefix`; observe-only and codex legs passed (already implemented by M2 / pre-existing observe behavior).
- M4/M5: the composition cells landed green-on-arrival — the composition behavior shipped with the M2 engine (whose RED run above shows zero trace lines on the same invocation family, the honest pre-state); they are the integration pins plan §F M4 names. AC-014's pin reds for the same reason as the M2 batch (captured there).

### Frozen-behavior preservation (post-run re-verification)

1. t1378 REQ-012: the four timing tests pass unmodified (multiple runs, exact names; the test file has zero commits on this branch).
2. cc/glm child argv byte-parity: AC-005/006 assert `-d` (and the `-- extra` tail) reach the child verbatim.
3. Operator `RUST_LOG`: AC-011 arm (b) — `info` preserved, no `debug` appended.
4. `MOAI_LOG_LEVEL` axis: AC-014 — the level neither enables nor suppresses the trace.
5. Readout refusals: the `--spawn` refusal is untouched (no diff hunk touches it); the `-d` refusal mirrors it (AC-004).

### Cross-platform, lint, coverage

- Native build: `go build ./...` exit 0 (multiple runs). Windows: `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (`WINDOWS_BUILD_OK`, 2026-10-01 tree `d00210bf8` and re-checked at `2a4950d47`).
- `go vet ./internal/cli/ ./internal/config/` clean; `gofmt -l` clean; `golangci-lint run --timeout=4m ./internal/cli/... ./internal/config/...` → `0 issues.` (v2.1.6 — the CI pin). New warnings/lints introduced: 0.
- Boundary grep (E4): `git diff 152adf3bb..HEAD | grep -c AskUserQuestion` → 0. No interactive prompts, no hook-adjacent changes.
- Coverage (E3, scoped measurement — see Gaps): per-function coverage of the NEW code from the launcher-scoped `-coverprofile` (`go test -coverprofile -run 'TestCodex|TestCC|TestGLM|TestCG|TestFactory|TestWorktree|TestLauncher|TestThreeRunner'`): `beginDebug` 100%, `annotateDetail` 100%, `debugDump` 85.7%, `launcherDebugRequested` 100%, `hasEnvKey` 100%, `codexApplyDebugEnv` 100%, `codexDebugEnvDetail` 100%, `stripCodexDebugFlag` 100%. Touched-file aggregate (6 files, 89 functions): 72.6% function-average under the launcher-scoped subset (subset excludes non-launcher tests that also cover these files, so this is a lower bound, not the package number).
- internal/config: `go test -cover ./internal/config/` → 82.5% (pre-existing package level; this SPEC adds one constant, zero statements). internal/kanban: `go test -run '^TestClaimFactoryLane'` → ok (composition surface untouched).

### Gaps (explicitly unobserved)

1. **The whole-package `./internal/cli` suite does not complete on this machine under current load** — measured twice: this branch (`-timeout 25m -cover` → timeout panic at 1501.7s with 3 failures) and the unmodified baseline `152adf3bb` (`-timeout 30m -cover` → 1801.8s, again over its timeout, with a DIFFERENT flake set: `TestStopChainMemberCostWithinBudget` + the shared language-detection failure; partial coverage 45.9% from the killed run). Load average 69–109 with four other lane test runs observed concurrently. Per the local doctrine (AGENTS.md §4: a full-suite run on a loaded developer machine measures the machine), the repository-wide verdict is **owned by the CI run on the integration branch and is PENDING at report time**.
2. **The package-level coverage percentage is therefore unmeasured on this machine** (both full -cover runs died before completion). The scoped per-function numbers above are the honest substitute; the 85% package target is asserted by CI, not by this report.
3. Flake set observed in full/subset runs (all acquitted by alone-runs and/or the baseline comparison): `TestCodexGoalContinueUntilMet`, `TestCodexTaskBackgroundHandshakeHonorsTaskBound` (non-deterministic alone), `TestCodexHarnessHooksRegisterNoFactoryPeer`, `TestCodexGoalUnmeasuredCapBoundsContinuation`, `TestStopChainMemberCostWithinBudget` — none touches this SPEC's surface.
4. **Pre-existing defect discovered (not this SPEC's scope, route via feedback/card)**: `TestSyncGateLanguageDetectionMatchesScript/kotlin_source` fails deterministically and identically on the unmodified baseline — `languages: Go = [kotlin], script = [kotlin java]` (`codex_sync_gate_test.go:115`). Environment-dependent Go-detector vs script divergence, zero relation to the launcher surface.

### Residual-risk

- The lane relaunch loop (`-f lane -d`) traces its pre-exec steps and dumps once before the first card session; per-card handoffs inside the loop are not individually traced (v1 single-level debug; the card id and worktree already print on the loop's stdout line).
- `RUST_LOG` linkage on the spawn door rides the command-scoped assignment with an absence check against THIS process's env; a tmux server carrying a stale `RUST_LOG` could still win in the pane (best-effort linkage per plan Residual 1).
- The debug-off lane threshold report's step set is asserted frozen by test (AC-013); the threshold start-point remains the init gate as in t1378.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-02
run_commit_sha: 2a4950d47
run_status: complete
run_status_note: code-complete with scoped verification; full-package suite + package coverage delegated to CI (machine-bound, see §E.2 Gaps 1-2)
ac_pass_count: 15
ac_fail_count: 0
preserve_list_post_run_count: 5
preserve_list: [t1378 REQ-012 threshold report + 4 frozen tests unmodified, cc/glm child-argv byte parity, operator RUST_LOG preserved, MOAI_LOG_LEVEL axis untouched, readout refusal discipline mirrored]
l44_pre_commit_fetch: n/a (run phase commits on the run branch only; no fetch performed)
l44_post_push_fetch: n/a (run phase performs no push — lane-owned)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: pass (darwin/arm64, go build ./...)
  windows: pass (GOOS=windows GOARCH=amd64 go build ./...)
total_run_phase_files: 13
m1_to_mN_commit_strategy: per-milestone commits M1..M5 on WT-codex-debug-mode-run (e22be7956, d00210bf8, 275ebb21a, 4e2022868, 2a4950d47) + this evidence commit; lane fast-forwards WT-codex-debug-mode
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-02
sync_commit_sha: pending-backfill-sync
sync_status: complete
changelog_entry_position: "[Unreleased] > Added — first bullet (newest-first within the section; t1378's entry demoted one position)"
b12_self_test_a: "pre-emission grep -c SPEC-CODEX-DEBUG-MODE-001 CHANGELOG.md → 0 (cleared to emit)"
b12_self_test_b: "AC live-count counter over acceptance.md (tier M source) → 15 live / 0 excluded / 0 ambiguous; CHANGELOG entry cites AC-001..015 = 15 (match)"
b12_self_test_c: "all 7 cited implementation paths verified via ls (internal/cli/launcher_debug_trace.go + codex_launcher.go + codex_factory.go + cc.go + glm.go + factory_launch_timing.go + internal/config/envkeys.go)"
mx_compliance: "launcher_debug_trace.go 4x @MX:NOTE; factory_launch_timing.go 3x @MX:ANCHOR (+@MX:REASON) on beginDebug/annotateDetail/debugDump (fan-in >= 3 across the three launchers); existing @MX:SPEC sub-lines preserved; per-file limits respected (anchor 3/3, note 4/10 in launcher_debug_trace.go); comment language en per code_comments"
codemap_refresh: "modules.md internal/cli row 395→396 (root 314→315, launcher_debug_trace.go) + new 현재 부분 갱신 t1380 paragraph (t1378 demoted to 이전 갱신, per t1378/t1383 precedent); provenance.json re-stamped to this tree (3a32a654b)"
docs_site_decision: "no edit — docs-site/content/*/advanced/codex-dual-harness.md enumerates launch-path forms (기본 실행·cli·app·--spawn·-w) for the local-instruction load order, which -d does not falsify (flag, not a path form); the missing -d mention is an omission, recorded as a follow-up candidate, not expanded in this sync"
spec_frontmatter: "status in-progress → completed, updated → 2026-10-02 (frontmatter-only; spec/plan/acceptance body untouched)"
verification_scope: "gofmt -l clean + go build ./internal/cli + go vet ./internal/cli clean after MX edits (this run, tree HEAD 3a32a654b + doc edits); spec lint tree-sourced (see below); full internal/cli suite NOT re-run at sync — machine load 69-109 with parallel lanes, per the run-phase Gaps record; run-phase verification at 3a32a654b stands (15/15 AC, decisive selector ok 20.628s)"
recorded_by: "manager-docs (sync phase, card t1380)"
```

## §F Phase 4 Mode Selection

### Kickoff record (operator-held gate)

- The card dispatch (leader → lane, 2026-10-01) marked the plan→run Kickoff for OPERATOR DIRECT ANSWER — the keep-set operator form, not the autonomous transition.
- Operator answered via AskUserQuestion in the lane on 2026-10-01: **착수 (proceed to run phase)** — the recommended option; alternatives were SPEC 재검토 (override decision-index defaults) and 중단 (abort).
- Gate evidence at ask time: plan-audit iter2 PASS 1.00 (≥ 0.80 Tier M), RED-now 9 cells re-executed by the auditor with pin validity proven, tree-sourced spec lint 0 errors / 0 warnings, plan-artifact hash `257c6ac68642441dfb6b9218a63b71c078dfa10964baca303e42c6dd70425576` fixed since the verdict. decision-index Q2-Q5 defaults proceed (both spellings / RUST_LOG best-effort / single-level v1 / readout refusal) — none gates run-phase entry per the plan-phase record.

### Input parameters

- tier: M · scope: ~8-12 files (internal/cli launchers + internal/config defaults/envkeys + tests) · domain count: 2 · file language mix: 100% Go · concurrency benefit: LOW (coding-heavy, dependency-ordered milestones).

### Mode evaluation

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | no | multi-file semantic feature, not a one-liner |
| fanout | no | coding-heavy work (Anthropic coding-task parallelism caveat); milestones are dependency-ordered |
| sweep | no | semantic new-code work, not a mechanical uniform transform |
| agent-team | no | not operator-requested; experimental surface stays unselected |
| serial | **yes** | one manager-develop per milestone, plan §F order |

Decision: serial

Justification: coding-heavy Go across two packages with strictly ordered milestones (flag surface → trace engine → uniformity → composition → sweep); a single manager-develop with the Section A-E delegation template carries the launcher-path context cheaper than any fan-out. No boundary case: scope sits inside the serial band.

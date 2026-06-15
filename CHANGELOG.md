# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **[SPEC-EVIDENCE-CLAIM-INVARIANT-001](.moai/specs/SPEC-EVIDENCE-CLAIM-INVARIANT-001/spec.md)** — Verification-Claim Integrity Doctrine: "no unobserved-verification-claim" Invariant + Baseline-Integrity Attribution + 5-Section Evidence-Bearing Report Format (Tier S, v0.1.0, run-phase 2026-06-15, sync-phase 2026-06-15). Pure doctrine rule (markdown only, no code/CI/runtime hook): codifies the policy that no actor MUST assert a verification or completion it did not actually observe. Evidence-absent ≠ evidence-of-success. Binds two surfaces — orchestrator self-report (Completion Report + Verification Matrix per `.claude/output-styles/moai/moai.md` §8) and manager-agent completion report (§E self-verification E1-E7 in manager-develop + manager-docs). Baseline-integrity attribution requires command + observed-output pair (not carry-over, not inferred). 5-section report format (§3) enforces separation of Claim / Evidence / Baseline / Gap / Risk. Complementary to runtime detection layer SPEC-STOP-EVIDENCE-GATE-001 (IMP-02/03, advisory gate) — this SPEC codifies policy; that runtime gate detects one shape of violation. IMP-06 final adopted item of fable-ish "Verify, Don't Assume" analysis roadmap. **AC: 7/7 PASS** (AC-ECI-001..007, all MUST, all falsifiable via grep/file-existence) — verification-claim-integrity.md body + template mirror byte-parity, §3.2 C1-C7 concrete verification idioms (ERE grep alternation + literal-pipe escape + synonym hatch + status-literal + colon-only anchor + human-tally -ge 4 + self-application), no unobserved-claim in this SPEC's own sync report. Zero cross-platform build concerns (markdown-only). Zero template-neutrality violations (internal_content_leak_test.go CI guard). plan-auditor 0.84 PASS-WITH-DEBT (Tier S threshold 0.75, testability debt remediated to all 7 AC now executable).

### Added (continued)

- **[SPEC-STOP-EVIDENCE-GATE-001](.moai/specs/SPEC-STOP-EVIDENCE-GATE-001/spec.md)** — Stop-hook Verification-Evidence Completion Gate + Session Ledger Reader (advisory, warn-first) (Tier M, v0.1.1, run-phase 2026-06-15, sync-phase 2026-06-15). **Honest framing (knowingly-dormant scaffold)**: Stop-hook advisory evidence gate (warn-first, fail-open, ≤5s) + session-ledger reader reusing telemetry.LoadBySession + UsageRecord omitempty extension (IsTestPass/IsTestFail/PathKind binary evidence fields). This SPEC is a **scaffold**: wired and correct, but does NOT fire against real telemetry yet because the sole production `UsageRecord` writer `logSkillUsage` hardcodes `Outcome: OutcomeUnknown`, blocking gate activation. The target-defect value (code-session false-success detection) is **blocked on the successor SPEC `SPEC-STOP-EVIDENCE-WRITER-001`** that will set `PathKind=code-change`, `Outcome=success`, and `IsTestPass` at record time. This SPEC delivers the read-side logic, schema, and graceful fallback inference; it does NOT claim to catch false-success claims — that claim becomes falsifiable only when the writer sets the required fields. **11/11 AC PASS** (AC-SEG-001..011, all MUST) — 100% coverage session_ledger.go, 88.9% stop.go gate-lines; hook 81.8%, telemetry 75.9% (no regression); cross-platform build (darwin/windows amd64) exit 0; golangci-lint 0; C-HRA-008 boundary clean.

- **[SPEC-HOOK-DISCIPLINE-WIRING-001](.moai/specs/SPEC-HOOK-DISCIPLINE-WIRING-001/spec.md)** — Discipline Hook Wiring: status-transition-ownership + sync-phase-quality-gate + language-generalization (Tier M, v0.1.0, run-phase 2026-06-15, sync-phase 2026-06-15). Phase-2 realization (pre-announced deferred wiring from the agent-team-rebuild milestone) wires three discipline hook scripts into `settings.json.tmpl`: **status-transition-ownership.sh** (PostToolUse advisory, timeout 5s) guards SPEC frontmatter status transitions per the owner matrix; **sync-phase-quality-gate.sh** (Stop warn-first, timeout 10s) enforces lint+test+coverage-delta+language-neutrality gates. Generalization of the sync-gate from hardcoded Go-only tooling to marker-driven `detect_language()` for 16-language parity (Go/Node/Python/Rust/Java/Kotlin/C#/Ruby/PHP/Elixir/C++/Scala/R/Flutter/Swift, plus graceful skip via `command -v` guard + silent-pass on unrecognized marker). **9/9 AC PASS** — AC-HDW-001..009 (status-transition wiring, language-auto-detect with real-git fixture, sync-gate registration, team-ac-verify deferred, handler coexistence, template-neutrality CI green, local↔template byte-parity via manual mirror, advisory-only no-op default, Go-tool boundedness proof: total==inblk word-boundary match on invocation sites). Cross-platform build (darwin/windows amd64) exit 0, golangci-lint 0, bare-pipe neutrality tests GREEN, C-HRA-008 boundary (AskUserQuestion grep 0). plan-auditor 0.83 PASS-WITH-DEBT (Tier M threshold 0.80, Mode 5 sub-agent). Design decision D2: dormant `exit 2` path gated behind `MOAI_SYNC_GATE_BLOCKING=1` (OFF by default — warn-first), scope discipline (team-ac-verify NOT wired per REQ-HDW-004 exclusion).

### Removed

- **[SPEC-COMPLETION-MARKER-RETIRE-001](.moai/specs/SPEC-COMPLETION-MARKER-RETIRE-001/spec.md)** — Retire completion markers + dormant persistent-mode subsystem (Tier M, v0.1.0, era V3R6, run-phase 2026-06-15, sync-phase 2026-06-15). Removes, as one coupled dead/dormant unit, the MoAI completion markers `<moai>DONE</moai>` / `<moai>COMPLETE</moai>` and the never-activated persistent-mode subsystem. **Go runtime** (`internal/hook`): `stop.go` marker-detection loop + persistent-mode branches removed, `lifecycle/persistent_mode.go` retired whole-file, `compact.go` `readPersistentMode` + Execution-Mode section removed (worktrees section preserved). **Config** (`internal/config`): `CompletionConfig` / `MarkersConfig` structs + defaults removed; `completion:` block dropped from `.moai/config/sections/workflow.yaml` + template mirror; `pkg/models/config.go` `LogCompletionMarkers` removed. **Tests**: 2 whole files retired (`stop_completion_test.go`, `lifecycle/persistent_mode_test.go`) + targeted function/assertion removals across hook + config test files. **Prompt-layer**: 8 live `.claude/` consumers (output-styles `moai`/`einstein`, `SKILL.md`, `loop.md`, `release.md`, `spec-workflow.md`, `release-update-specialist.md`) + 5 template mirrors. **docs-site**: 4-locale (en/ko/ja/zh) content + `docs-site/README.md`. **Design decision Option (a)**: marker emission dropped entirely; `/moai loop` exit signal replaced with the natural-language completion sentence `"All loop completion conditions satisfied; exiting loop."` (the other 4 loop-exit conditions — zero-errors+tests+coverage, max-iterations, memory-pressure, user-interruption — preserved). Run-phase milestones M1-M6 (Go runtime → config → test retirement → output-style+template → docs-site 4-locale → traceability). **AC: 11/11 PASS** — 9 MUST-FIX (AC-CMR-001..008, 011) + 2 SHOULD-FIX (AC-CMR-009 `moai.md` XML-exception removal, AC-CMR-010 traceability); no deferred AC. Cross-platform build (darwin/windows amd64) exit 0; golangci-lint 0; AC-CMR-004 exhaustive zero-residual production grep clean. **Partial supersession** (marker/persistent-mode surface only): SPEC-PERSIST-001 (marker→deactivate contract) + SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 (AC-WSE marker-oracle inversion).

### Fixed

- **[SPEC-HARNESS-EXECUTE-E2E-001](.moai/specs/SPEC-HARNESS-EXECUTE-E2E-001/spec.md)** — Harness execute regression-gate measurementRoot 결함 수정 + e2e 재현 테스트 (Tier M, v0.2.0, run-phase 2026-06-15, sync-phase 2026-06-15). 선행 SPEC-HARNESS-APPLY-EXECUTE-001(completed)이 남긴 보증 갭 `moai harness apply --execute --id <id>` verb가 genuine proposal에서 항상 fail-close하는 **production 결함** 수정. Root cause: `internal/harness/applier.go` `applyWithRegressionGate` 내 regression gate의 `measurementRoot(snapshotDir)` 메서드가 실제 project root 대신 snapshot base dir을 반환하므로 gate가 항상 "packages 없음" 실패를 발생. 수정: `Applier.WithProjectRoot(root)` fluent builder seam 추가 + regression gate가 명시 root를 사용하도록 threading. 신규 테스트 파일 2종: `internal/cli/harness/execute_e2e_test.go` (e2e 재현 — RunExecute → real pipeline → real measurer → apply_outcome 기록), `internal/harness/applier_projectroot_test.go` (WithProjectRoot set/unset unit test). 수정 파일: `internal/harness/applier.go` (+72 net: projectRoot field + WithProjectRoot setter + measurementRoot에서 fallback), `internal/cli/harness/execute.go` (+7: WithProjectRoot 체이닝). **9/9 AC PASS** (AC-E2E-001..009; 8 MUST-PASS gating + 1 SHOULD-PASS coverage non-regression); coverage internal/cli/harness 77.9% (equal) / internal/harness 87.6% (+0.1pp); cross-platform build (darwin/windows amd64) exit 0; golangci-lint 0; C-HRA-008 boundary clean (AskUserQuestion grep 0); FROZEN (pipeline/outcome/observer/regression_gate/harness.yaml) byte-identical diff 0; plan-auditor iter-1 FAIL 0.58 → iter-2 PASS 0.89 (Tier M threshold 0.80; skip-eligible NO, Phase 0.5 재실행 불요). **실제 apply 문제 완전 해소**: production 경로가 project root을 올바르게 측정하여 genuine proposal에서 verdict="kept" apply_outcome telemetry 기록됨.

### Added

- **[SPEC-HARNESS-APPLY-EXECUTE-001](.moai/specs/SPEC-HARNESS-APPLY-EXECUTE-001/spec.md)** — Harness Apply 실행 동사 — opt-in Go execute verb로 Applier.Apply() 프로덕션 caller 활성화 (Tier M, v0.1.0, run-phase 2026-06-15, sync-phase 2026-06-15). Self-Harness 로드맵 P2 "observer/gate activation" 1차: `moai harness apply --execute --id <proposal-id>` 신규 verb가 dormant `Applier.Apply()` pipeline을 처음으로 프로덕션 경로에 배선한다. 실행 흐름: Pipeline(`AutoApply: true`) 생성 → regression gate + outcome observer 배선 → Apply 호출 → 첫 `apply_outcome` telemetry 기록. **정직한 가치 표명 (MUST-PASS)**: markdown-only FROZEN allowlist (현 harness 쓰기 표면)이므로 regression gate는 typical case에서 Δ=0으로 always-pass다. 본 SPEC의 실질 가치는 정확히 두 가지: (1) **telemetry-infrastructure enabler** — apply-outcome 이벤트를 `usage-log.jsonl`에 기록하고 proposal_id로 correlate하여 Phase 5 분석의 입력 기초 제공; (2) **defense-in-depth safety net** — L1~L4 강제(Frozen/Canary/Contradiction/RateLimit)되며 L5는 CLI 레벨 auto-approve, allowlist 확장/applier 버그 시 fire 대기. 신규 파일: `internal/cli/harness/execute.go` (verb factory + RunExecute) + 테스트 3종 (`internal/cli/harness/execute_test.go`, `internal/cli/harness_execute_test.go`, `internal/harness/execute_integration_test.go`). 수정: `internal/cli/harness.go` (apply --execute flag 위임), `internal/cli/harness_route.go` (NewExecuteCmd 등록). `internal/harness/applier.go`는 호출 대상일 뿐 미수정(FROZEN diff 0). 16/16 AC PASS (AC-AEX-001..016); coverage internal/cli/harness/execute.go 86.0% (≥85% DoD), internal/harness 87.5% (no regression); cross-platform build (darwin/windows amd64) exit 0; golangci-lint 0; C-HRA-008 boundary (AskUserQuestion grep 0 matches); FROZEN (applier/pipeline/regression_gate/outcome/observer/canary/lineage/harness.yaml) byte-identical; 마크다운-only 표면에서의 honest framing (회귀 방지 과대광고 금지). plan-auditor 0.91 PASS (Tier M threshold 0.80, skip-eligible Phase 0.5).

- **[SPEC-HARNESS-OUTCOME-CAPTURE-001](.moai/specs/SPEC-HARNESS-OUTCOME-CAPTURE-001/spec.md)** — Harness Apply outcome capture: records each Apply's verdict + project-health delta + proposal_id correlation into usage-log.jsonl (Tier M, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-14). Self-Harness P2 Phase5 enabler: captures apply-outcome events for downstream analysis (failure-signature clustering, canary-effectiveness evaluation). **Honest framing (MUST-PASS)**: this SPEC is **capture + persist ONLY**. The captured delta is typically `Δ=0` for the current markdown-only harness write surface. The record makes the Apply outcome observable (persisted to usage-log.jsonl) and correlable (via proposal_id) but does NOT improve, prevent, or block anything — those consumers (Phase5 analysis/visualization) are downstream and out of scope. New file: `internal/harness/outcome.go` (RecordOutcome method on Observer, delegates to RecordExtendedEvent append machinery). Modified: `internal/harness/types.go` (additive EventTypeApplyOutcome const + 9 omitempty OutcomeVerdict/OutcomeDecision/OutcomeProposalID/OutcomeBaseTests/OutcomeCandTests/OutcomeCandCoverage/OutcomeCandLint/OutcomeRegressed fields on Event struct; LogSchemaVersion "v1" → "v2"), `internal/harness/applier.go` (outcomeObserver field + WithOutcomeObserver setter + recordOutcome seam at both keep/rollback terminal branches of applyWithRegressionGate). New test file: `internal/harness/outcome_test.go` (12 tests: round-trip + omitempty + JSONL append + emit-on-keep + emit-on-rollback + contract-preserved + gate-inactive + nil-observer + record-error-no-flip + lineage correlation + boundary clean + frozen invariants). 12/12 AC PASS (AC-OC-001..012); coverage internal/harness 87.3% package (87.8% across the harness package tree, no regression); cross-platform build (darwin / windows amd64) exit 0; golangci-lint 0; lineage.go 0-diff; C-HRA-008 (AskUserQuestion) boundary verified 0 matches; FROZEN (scorer, tier, auto_apply, safety layers, harness.yaml level=standard) byte-identical; no `lineage_id` field introduced (C11 satisfied). plan-auditor iter-1 0.88 PASS-WITH-DEBT (3 MINOR orch-direct remediated) → iter-2 (Phase 0.5 re-audit) PASS 0.92 (Tier M threshold 0.80; not skip-eligible — gate re-run after remediation). The observer is **passive** — it never feeds back into Apply decisions, never flips verdicts, never alters baseline-store/lineage/regression-gate behavior. **Full integration ready**: ApplyOutcome event + proposal_id correlation key (reused from lineage) enable Phase5 observability without requiring schema changes to lineage itself (DD-3 correlation strategy proved).

- **[SPEC-HARNESS-LOOP-CLOSURE-001](.moai/specs/SPEC-HARNESS-LOOP-CLOSURE-001/spec.md)** — First clean human-gated harness apply + auditable lineage manifest (Tier S, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-14). Proof-of-mechanism SPEC demonstrating the harness apply/lineage/rollback loop closure: when a user approves a harness proposal to modify a `my-harness-*` skill SKILL.md description field, the system atomically (1) creates a pre-apply snapshot under `.moai/harness/learning-history/snapshots/<proposal-id>/`, (2) applies the change to the skill, (3) appends an auditable lineage entry to `.moai/harness/learning-history/manifest.jsonl` with the proposal ID, decision (approved/rejected), applied-surface field, timestamp, and optional reason, and (4) makes the previous state restorable via `RestoreSnapshot()` loading the snapshot by proposal ID — with zero observable difference for rejected proposals (the active harness unchanged). The lineage is **append-only** and round-trips cleanly via `LoadManifest()` — new entries cannot corrupt existing history. Two new files: `internal/harness/lineage.go` (WriteLineageEntry + LoadManifest) and `internal/harness/lineage_test.go` (13 tests covering approved/rejected/pending/backward-compat/byte-identical restore); two modified files: `internal/harness/types.go` (added LineageEntry struct) and `internal/harness/applier.go` (extended Apply() with lineage writes at accept+reject codepaths). 8/8 AC PASS (AC-HLC-001..008); coverage internal/harness 87.2% (no regression); cross-platform build (darwin / windows amd64) exit 0; golangci-lint 0. plan-auditor PASS 0.82 (Tier S threshold 0.75). Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/harness`. **Completes the harness learning subsystem M6+M7 proof (proposal auditable, decision logged, reverse possible), closing the "first clean apply" requirement set.**

- **[SPEC-HARNESS-REGRESSION-GATE-001](.moai/specs/SPEC-HARNESS-REGRESSION-GATE-001/spec.md)** — Harness M2-lite non-regression gate: measurement-infrastructure scaffold + defense-in-depth safety net (Tier M, v0.1.1, run-phase 2026-06-14, sync-phase 2026-06-14). Self-Harness roadmap P1: the non-regression gate ports the M2-lite signal from arXiv 2606.09498v1 into the Apply pipeline. **Honest framing (MUST-PASS)**: the gate measures three project-health dimensions (test pass count, coverage%, lint count) on a baseline and the post-apply candidate, requiring `Δ ≥ 0` for all dimensions. HOWEVER, because the FROZEN allowlist restricts harness proposals to markdown-only (`internal/harness/frozen_guard.go`) fields, modifying a skill SKILL.md description cannot change Go test/coverage/lint counts. **Therefore, the measured delta is typically `Δ=0` for the current write surface.** The gate's genuine value is stated honestly as exactly two things, NOT as "prevents regressions": (1) **measurement-infrastructure scaffold** — establishes the baseline store + delta comparator + measurement collector that Phase5's richer signals reuse; (2) **defense-in-depth safety net** — dormant fire-only if the allowlist is widened or an applier bug writes outside it. New files: `internal/measure/measure.go` (pure `ParseGoTestJSON` / `ParseCoverageFile` / `CountNonEmptyLines` exporteds in a new leaf package, zero import of `lsp` / `harness` / `loop` per C9), `internal/measure/measure_test.go`, `internal/harness/regression_gate.go` (new `MetricTriple`, `ApplyRegressionError`), `internal/harness/regression_gate_test.go`; modified: `internal/harness/applier.go` (seam at DecisionApproved, gate before lineage write, snapshot/rollback wiring), `internal/loop/go_feedback.go` (delegates parsers to `internal/measure`, byte-identical behavior). 13/13 AC PASS (AC-RG-001..013); coverage internal/measure 98.0% / internal/harness 86.1% (no regression); cross-platform build exit 0; golangci-lint 0; C-HRA-008 boundary clean; FROZEN DO-NOT-MODIFY (tier threshold, allowlist, scorer 4-dims, auto_apply gate, 5-layer safety chain) all unchanged; `go list -deps internal/measure` has no `internal/(lsp|gopls|harness|loop)`; C10 honest-framing AC PASS. plan-auditor iter-2 PASS 0.91 (Tier M threshold 0.80) → D1/D2/D3+D5/D6 orch-direct patches; m2-lite scaffold PRESERVE-verified.

### Fixed

- **[SPEC-HARNESS-OUTCOME-ERRJOIN-001](.moai/specs/SPEC-HARNESS-OUTCOME-ERRJOIN-001/spec.md)** — Apply rolled-back branch error propagation (Tier S, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-15). Fixes a correctness gap in the Apply pipeline's error handling when the outcome-record write fails on a regression-blocked rollback: the typed `*ApplyRegressionError` signal and the outcome-record error are now both preserved via `errors.Join(regErr, oerr)` so that `errors.As(err, &*ApplyRegressionError)` remains true (detecting the regression block) while the outcome-record error stays reachable via `errors.Is`/unwrap. **Honest framing (MUST-PASS)**: this fix applies to a DORMANT-SCAFFOLD code path — `Applier.Apply()` has ZERO production callers (test-only invocation). The fix is **forward-looking**: it matters once/if the Go Apply pipeline becomes a production apply path. Stdlib `errors` import added; the kept (success) branch is byte-unchanged. New test `TestApply_Outcome_RolledBack_RecordError` reproduces the scenario; existing regression-gate tests stay GREEN (no regression). 8/8 AC PASS (AC-ERRJOIN-001..008); coverage internal/harness 87.5% (no regression); cross-platform (darwin / windows amd64) build exit 0; golangci-lint 0. Scope disciplined to 2 files (applier.go + applier_test.go); frozen siblings (outcome.go, observer.go, regression_gate.go, measure.go) byte-unchanged. plan-auditor 0.91 PASS (Tier S threshold 0.75, skip-eligible Phase 0.5). Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/harness`.

- **[SPEC-GITSTRATEGY-SAVE-ISOLATION-001](.moai/specs/SPEC-GITSTRATEGY-SAVE-ISOLATION-001/spec.md)** — Restore git_strategy section isolation in `ConfigManager.Save()` (regression from SAVE-WIRING M1) (Tier S, v0.1.0, plan + run-phase 2026-06-13, sync-phase 2026-06-13). Fixes GitHub issue **#1064**: a scoped project-config write (`internal/web` `writeProjectConfig`) mutated the unrelated `git_strategy` section — the `33215af27` SAVE-WIRING M1 write leg unconditionally re-serialized all 6 owned sections, and the loader drops the unmodeled `sentinel:` key into compiled defaults, so the sentinel `git_strategy:\n  sentinel: DO_NOT_TOUCH` expanded into the full default tree on any project-config save. Adds a `gitStrategyDirty` flag to `ConfigManager`: set in `SetSection("git_strategy", …)`, the `git-strategy.yaml` write fires only when the flag is dirty **OR** the file is absent (greenfield create), and the flag resets after a successful `Save()` and on `Reload()`. Two RED reproductions (`TestWriteProjectConfigSectionIsolation` + `TestGoldenPath_ReadWriteRoundTrip`) flip to GREEN; the SAVE-WIRING round-trip + greenfield-create tests stay GREEN (no regression); new `TestConfigManagerSaveDirtyFlagReset` (EC-3) + `TestConfigManagerReloadResetsDirtyFlag` pin the dirty-flag lifecycle. Scope disciplined to `internal/config/manager.go` + its test only (loader / types / validation / defaults / templates unchanged; the 4 production `Save()` callers byte-unchanged). coverage internal/config 78.0% (no regression). plan-auditor PASS-WITH-DEBT **0.86** (Tier S threshold 0.80).

### Security

- **[SPEC-SEC-HARDEN-005](.moai/specs/SPEC-SEC-HARDEN-005/spec.md)** — Security hardening residual containment: shell-aware `${IFS}` word-split closure + update environment-trust allowlist (Tier M, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-15). Fast-follow of SEC-HARDEN-002 §F.1/§F.2 deferred items. **§F.1 (PRIMARY)** — `hasUnquotedShellSeparator` + `Matches` `:*` prefix-rule bypass via `${IFS}`/`$IFS` shell variable insertion (CWE-78 token-bypass, reachability: `:*` allow-rules only). Fixed via NEW direct dependency `mvdan.cc/sh/v3/syntax` (shell-aware parser, not $-blacklist heuristic) — fail-closed on malformed shell → DENY. Modified: `internal/permission/stack.go` hasIFSWordSplit helper + seam at Matches L100,127-136. **§F.2 (PRIMARY)** — moai update env vars (`MOAI_UPDATE_URL`, `MOAI_UPDATE_SOURCE`, `MOAI_RELEASES_DIR`) were read unchecked; off-allowlist scheme/host + URL-shaped `releases_dir` paths could redirect update checks to attacker hosts. Fixed via scheme(`https`)+host allowlist validation (3 env vars, `api.github.com` canonical default) in `internal/cli/deps.go` validateUpdateURL before remote checker construction — fail-closed on off-allowlist. **§F.3 (OPTIONAL, non-gating)** — godoc-only TOCTOU notes on restoreTargetContained/parentChainContained/runMXScan (zero code behavior change). 13/13 AC PASS (mvdan.cc/sh direct require, cross-platform go build exit 0, coverage permission 88.0%→89.5% improved, cli 71.8% baseline-equivalent, C-HRA-008 grep 0 matches, lint 0 NEW issues, dependency mvdan.cc/sh/v3 v3.13.1 pure-Go NFR-SEC5-002 confirmed, legacy command tests regressed after run-gate fix + valid correction).

- **[SPEC-SEC-HARDEN-004](.moai/specs/SPEC-SEC-HARDEN-004/spec.md)** — Security hardening fast-follow: symlinked intermediate directory write escape + symlink-in-root read amplification (Tier S, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-14). Fast-follow of SPEC-SEC-HARDEN-003 §F.3 (two adjacent-class containment fixes deferred from sync-auditor SHOULD-FIX PASS-WITH-DEBT). **F1 parent-chain symlink write escape** — `moai update` restore path leaf-level symlink guard (isSymlinkEntry + filepath.Rel `..` block) did not check parent directories, so a configDir-internal symlinked subdir (`configDir/linkdir → /outside`) could redirect a backup restore to write outside configDir (CWE-22 path-traversal, reachability: pre-existing symlinked dir required). Fixed by adding `filepath.EvalSymlinks(filepath.Dir(targetPath))` resolution + parent re-containment check to the shared helper `restoreTargetContained` — both modern walk (restoreMoaiConfig) + legacy walk (restoreMoaiConfigLegacy) benefit from the single-point fix (AC-SEC4-002 parity preserved). **F2 symlink-in-root read amplification** — MX sidecar file-changed hook used lexical-only containment check (`pathContainedIn`), so root-internal symlinks pointing to secrets outside root passed the lexical gate and ScanFile followed the link (CWE-61 symlink-following amplification, impact: MX-tag-text-only disclosure to in-root sidecar, no out-of-root write). Fixed by resolving the file path via `filepath.EvalSymlinks` before re-checking containment — fail-closed early return + structured logging on symlink-escape detection, async/5s/empty-payload contract preserved. Both fixes are **additive guards** — existing leaf/lexical gates are PRESERVED, parent-chain / post-resolve checks layer on top (AC-SEC4-003, AC-SEC4-006 regression validated). New source files: 0; modified: `internal/cli/update.go` (restoreTargetContained parent-chain seam) + `internal/hook/file_changed.go` (runMXScan scan-target seam) + test files for both (F1 + F2 reproduction tests RED→GREEN; SEC-HARDEN-003 leaf tests stay GREEN). 10/10 AC PASS (AC-SEC4-001..010); cross-platform build (darwin / windows amd64) exit 0; C-HRA-008 boundary (internal/hook specid import) PASS 0 matches; coverage internal/cli 71.7% / internal/hook 81.5% (no regression). plan-auditor PASS 0.91 (Tier S threshold 0.75, skip-eligible Phase 0.5). Deferred: TOCTOU window (check-vs-use race, offline single-process threat model accepted by sync-auditor), §F.1 `${IFS}` shell-aware word-split, §F.2 env-trust follow-ups.

- **[SPEC-SEC-HARDEN-003](.moai/specs/SPEC-SEC-HARDEN-003/spec.md)** — Security hardening fast-follow: non-isolated path containment (MX sidecar + legacy/modern backup restore) (Tier S, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-14). Fast-follow of SPEC-SEC-HARDEN-001/002 §F.3 (`internal/hook` + `internal/cli` uncovered paths). **2 MEDIUM containment fixes + 1 in-scope sibling**: **C-F1** (`internal/hook/file_changed.go:runMXScan`) — MX sidecar scan was non-isolated: a malicious `filePath` or a CWD pointing outside the project root could direct the sidecar write to an arbitrary path (CWE-22 path traversal). Fixed via `resolveProjectRootFromInputOrEnv` (existing canonical seam in `internal/hook/path_resolve.go`) + a new root-relative `pathContainedIn` additive guard; async/5s/empty-payload contract preserved byte-identical. **C-F2** (`internal/cli/update.go:restoreMoaiConfigLegacy`) — legacy backup restore walk did not check for symlinks, allowing a crafted backup archive with symlink entries to write outside `configDir` (CWE-61/CWE-22). Fixed via `os.Lstat`-based symlink rejection + `filepath.Rel`-based traversal guard (private helpers, no `internal/cli/specid` import per scope discipline). **C-F2 in-scope sibling** (`internal/cli/update.go:restoreMoaiConfig`, modern path) — the same symlink class existed in the modern anonymous walk callback; ground-truth verified as in-scope (excluding it would leave the primary code path vulnerable, constituting security theater per REQ-SEC3-006). The `specid` package import count remains 0 (AC-SEC3-007 clean). 10/10 AC PASS (AC-SEC3-001a, 001b, 002, 003, 004a, 004b, 004c, 005, 006, 007); `go test ./internal/hook/... ./internal/cli/...` exit 0; cross-platform build (host darwin / windows amd64) exit 0; `go vet` + golangci-lint 0 NEW issues; `grep -c os.Getenv internal/hook/file_changed.go` = 0 (NFR-SEC3-002 env-trust boundary); `grep -c asyncDeadline internal/hook/file_changed.go` = 5 (NFR-SEC3-001 contract preserved). plan-auditor PASS-WITH-DEBT 0.81→0.86 (Tier S threshold 0.75; iter-2 Phase 0.5 re-audit after D1/D2/D3+lint orch-direct patches). M1 commit `743ad1cc4` (C-F1) / M2 commit `2b9b791a4` (C-F2 + sibling). **Deferred**: §F.1 `${IFS}` shell-aware word-split (mvdan.cc/sh dependency decision pending), §F.2 env-trust — separate follow-up candidate.

- **[SPEC-SEC-HARDEN-002](.moai/specs/SPEC-SEC-HARDEN-002/spec.md)** — Security hardening fast-follow (Tier M, v0.1.0, run-phase 2026-06-14, sync-phase 2026-06-14). Fast-follow of SPEC-SEC-HARDEN-001 §F.3 (`cli` package group uncovered) + D2 (shell redirect operators `>`, `<` deferred). **M1** (internal/cli/specid — NEW leaf package) — unified `ValidateSpecID` sanitizer rejects `..`, `/`, `\`, absolute paths before CLI args reach `filepath.Join` or worktree creation sinks (CWE-22 path-traversal); placed in a NEW leaf package `internal/cli/specid` to break the import cycle where `worktree` needs to call it but cannot import `cli` (both CLI and worktree import specid; cycle resolved). Also exports `ValidateNoTraversal` for polymorphic CLI args (e.g., `moai worktree new`) that may be branch names with `/` (legitimate) — stricter flat-ID checks use `ValidateSpecID`, looser traversal checks use `ValidateNoTraversal`. **M2a** (internal/cli/worktree/new.go — HIGH, CWE-22) — path-traversal guard on `worktree new <SPEC-ID>` before `os.MkdirAll`, rejects traversal sequences (`..`, absolute paths) explicitly. Also includes `5c6085b55` fix-forward allowing legitimate branch names like `fix/something` (Contains `/` but no `..`, no absolute path — accepted by `ValidateNoTraversal`, rejected by strict `ValidateSpecID`). **M2b** (internal/core/git/worktree.go — MEDIUM, CWE-88) — argv option smuggling guard: `--` end-of-options separator before user-derived operands to `git worktree add`, prevents flag injection. **M3** (internal/cli — LOW, CWE-22) — `ValidateSpecID` guard at exactly THREE spec-subcommand boundaries: `spec view <SPEC-ID>` (spec_view.go:45), `spec status <SPEC-ID>` (spec_status.go:73), `spec close <SPEC-ID>` (spec_close.go:107). `spec drift` excluded (no positional SPEC-ID). **M4** (internal/permission/stack.go — MEDIUM) — redirect-operator hardening: `>` and `<` added to `hasUnquotedShellSeparator` `case` statement (line continues `case c == ';' || c == '|' || c == '&' || c == '`' || c == '\n' || c == '>' || c == '<'`), completing the quote-aware shell-separator scan deferred from SPEC-SEC-HARDEN-001 D2. **Two notable incidents** (frame honestly in audit trail): **import-cycle AC contradiction** — AC-M1-004 grep pinned `validateSpecID` definition to `internal/cli/*.go`, but `worktree new` needed to call it from `worktree` package (which CLI imports), violating Go import rules → resolved by leaf-package relocation to `internal/cli/specid` + plan-auditor + run-gate both missed it + `fa48dd8a6` amended AC-M1-004 to match; **M2a polymorphic-arg regression** — initial guard rejected branch-name `/` incorrectly, caught post-push when full test suite ran (not before), fixed forward `5c6085b55` without reverting (main was briefly red). 15/15 AC PASS (AC-SEC2-M1-001..004, M2-001..004, M3-001..003, M4-001..004); coverage permission 90.2% / core/git 88.1%; cross-platform build (darwin / windows amd64) + golangci-lint 0 + C-HRA-008 green. plan-auditor 0.83→0.91 PASS-WITH-DEBT (Tier M threshold 0.80); D3 `${IFS}` word-split remains deferred per §F.1.

### Changed

- **[SPEC-MERGE-METHOD-CONFIG-001](.moai/specs/SPEC-MERGE-METHOD-CONFIG-001/spec.md)** — Configurable sync-phase PR merge method (Tier M, v0.1.0, plan + run-phase 2026-06-13, sync-phase 2026-06-13). Resolves GitHub issue **#1061**: the sync-phase `gh pr merge` method was hardcoded to `--squash` across the delivery workflow + manager-git agent + git-workflow skill. Adds a per-mode `git_strategy.<mode>.merge_method` config field (enum `squash` | `merge` | `rebase`, **default `squash`** to preserve current behavior) — loaded + enum-validated (field-path error, fail-safe empty→default) following the `validateGitConventionConfig` pattern, and wired conceptually to the existing (previously unused) `internal/github` MergeMethod abstraction. The 3 hardcoded `--squash` template/doc sites now honor the config; the squash-default rendered command is byte-equivalent to the prior behavior. The FROZEN `spec-workflow.md` lifecycle table is widened non-destructively from literal "squash" to "configured `merge_method` (default `squash`)" — SSOT + template mirror byte-identical, the squash recommendation + rationale preserved, GATE-2 human-approved. 13/13 AC PASS (AC-MMC-001..013); coverage internal/config 78.1% (no regression); cross-platform (darwin / windows amd64) exit 0; golangci-lint 0; FROZEN mirror parity (`TestRuleTemplateMirrorDrift`) + template neutrality green. plan-auditor PASS-WITH-DEBT **0.82** (Tier M threshold 0.80) → defects D2-D7 patched (`ca4056509`). Built on the #1064 git_strategy Save isolation fix (the merge_method field rides the same git_strategy section through the dirty-flag gate). EX-1 (PRMerger production wiring) + EX-2 (gitflow per-branch-type override) deferred as candidate follow-ups.

- **[SPEC-PREPUSH-SAVE-WIRING-001](.moai/specs/SPEC-PREPUSH-SAVE-WIRING-001/spec.md)** — git_strategy config WRITE/Save leg — completes the PREPUSH dead-config chain READ/WRITE symmetry (Tier S, v0.2.0, plan + run-phase 2026-06-13, sync-phase 2026-06-13). The **terminal SPEC** of the PREPUSH chain (WIRING → MODE → LOADER → SAVE). `ConfigManager.Save()` persisted only 5 sections (user/language/quality/git-convention/llm); git_strategy was loaded (LOADER-WIRING-001 READ leg) but never written back, so `SetSection("git_strategy", …)` → `Save()` silently dropped the edit — the asymmetry LOADER-WIRING-001 **AC-PLW-008** explicitly deferred. This adds a single `saveSection(sectionsDir, "git-strategy.yaml", gitStrategyFileWrapper{GitStrategy: m.config.GitStrategy})` block to `Save()` (manager.go:191), a 1:1 mirror of the git-convention WRITE leg, **reusing** the existing `gitStrategyFileWrapper` (no new type, no validator, no defaults/loader/template change; `SetSection`/`GetSection` `case "git_strategy"` already existed). The round-trip (`SetSection` → `Save` → `Reload`) now preserves non-default git_strategy values (`Mode`, `Team.Hooks.PrePush`). **Latent infrastructure** — pre-stages the future web-console git_strategy editor export seam (a separate SPEC); no production caller mutates `cfg.GitStrategy` then saves yet (honestly framed as completing a documented READ/WRITE asymmetry, not building live dead code). 5/5 AC PASS (AC-PSW-001..005; AC-PSW-005 MUST scope-disciplined to `manager.go` + test only — validator/loader/defaults/types diff 0), coverage internal/config 77.9% (no regression), cross-platform (darwin / windows amd64) exit 0, golangci-lint 0, @MX:ANCHOR Save (fan_in=12) preserved. plan-auditor PASS **0.91** (Tier S threshold 0.75, skip-eligible; D1/D2 MINOR). Closes the PREPUSH dead-config chain — READ/WRITE round-trip now symmetric.

- **[SPEC-WEB-CONSOLE-009](.moai/specs/SPEC-WEB-CONSOLE-009/spec.md)** — git_convention config redesign — drop custom engine, wire auto-detection + max_length, fix flat→nested schema ("honest hybrid") (Tier L, v0.2.0, plan + run-phase M1-M8 2026-06-08, sync-phase 2026-06-08). The second half of the 42-defect `config ↔ code ↔ web-console ↔ runtime` audit (008 took the 17 statusline defects; 009 takes the **25 git_convention defects**) under the same honest-hybrid disposition. **REMOVE** (config theater / dead surface): the entire `custom` convention engine across all 4 lockstep sites (the `models` `oneof` tag, the `internal/config/validation.go` `validGitConventionNames` map + custom-required block, the `internal/web/validate.go` `conventionCanonical` slice, the `internal/cli/profile_setup.go` wizard option) — `custom` was a runtime-unreachable surface (`LoadConvention` could never load it); the `CustomConventionConfig` + `FormattingConfig` structs + `Custom`/`Formatting` fields + `Validation.{Enabled, EnforceOnCommit}` fields (zero consumers); the dead `LoadFromConfig` function (0 production callers, GCR-4) + its tests. The `custom` engine is **removed, not revived** (wiring would be a feature, not a defect fix — audit policy decision 1). **WIRE** (live levers): **Fix A** — `LoadConvention` now honors the full `auto_detection` config (the `enabled` flag gates detection, `sample_size` is forwarded to `Detect()` instead of the hardcoded `100`, `confidence_threshold` gates the detected result, and the configured `fallback` replaces the hardcoded `"conventional-commits"`); **Fix B** — a new `SetMaxLength` setter forwards `validation.max_length` to the loaded `Convention`, so the documented max-length is enforced instead of being silently overridden by the built-in's value. **FIX** (structural drift): **Fix C** — the template `git-convention.yaml` is rewritten from the FLAT schema (which never unmarshaled into the nested struct → silently dropped on every `moai init`) to the trimmed nested shape, `max_length` standardized 72→100; the local YAML aligned; `GitConventionConfig` added to the struct↔YAML symmetry CI guard (added LAST, after the struct trim + template rewrite made them symmetric — HARD-10 sequence gate). **HARD invariants held**: statusline **untouched** (008 scope, git diff 0); the 006 scope-boundary sentinel (`integration_test.go`) **byte-identical** (git_convention is in-scope/writable, but the sentinel-owning file was not modified at all); the **GCR-5 deferred boundary** untouched (`.git_hooks/pre-push`, `hook_install.go` `prePushHookContent`, `git-strategy.yaml hooks.pre_push` — git diff 0; the dormant pre-push engine wiring + the `git_strategy.hooks.pre_push` gap remain a maintainer-decision follow-up SPEC); `validation.enforce_on_push` (the one live gate) + the `isEnforceOnPushEnabled()` read path **preserved**; the built-in `Parse`/`ParseBuiltin` engine + `Detect()` signature **unchanged**; **zero network** (CDN grep 0); **zero Node** (`templ generate` codegen, drift-free). The `LoadConvention` signature-change cascade was controlled to exactly its 1 production caller (`hook_pre_push.go:54`). Web: removed the `custom.pattern` widget, added an `auto_detection.sample_size` number field + a `validation.enforce_on_push` toggle, fixed the empty-option label `"(project default)"`→`"(unchanged)"` + count `"8 fields"`→`"9 fields"`, greyed the detection sub-fields when convention≠auto. 26/26 AC PASS (AC-WC9-001..026); coverage internal/web 72.3% (no regression vs the 006/007/008 baseline) / internal/git/convention 96.4%; cross-platform (darwin / windows amd64) exit 0; golangci-lint 0. plan-auditor Phase 0.5 re-audit **0.87 PASS-WITH-DEBT** (Tier L threshold 0.85, MP-1..4 PASS); the audit-2 D1/D2/D3 + the orchestrator-caught audit-3 D4 were AC-command idiom precision fixes (vacuous / false-fail test patterns — the recurring cohort idiom class), implementation unaffected. **009 of the web-console-v4 cohort** (builds on the 006 Templ tree + 007 nested write-seam; sibling of 008 — together they close the full 42-defect audit).

- **[SPEC-WEB-CONSOLE-008](.moai/specs/SPEC-WEB-CONSOLE-008/spec.md)** — Statusline config redesign — remove config theater, wire live levers, fix drift ("honest hybrid") (Tier M, v0.2.0, plan + run-phase M1-M9 2026-06-07, sync-phase 2026-06-07). Derived from a 42-defect `config ↔ code ↔ web-console ↔ runtime` audit; addresses the **17 statusline defects** under an honest-hybrid disposition (remove dead/lying surface, wire existing levers, fix structural drift — first of a 2-SPEC split, git_convention's 25 defects deferred to **009**). **REMOVE**: the `mode` config surface (template `mode:` YAML + the web `statusline_mode` select + `statuslineModeCanonical` validation) — `mode=full` was a functional lie (`renderFullV3`, the 5-line layout, is never called in production); the **Builder API is preserved** (`StatuslineMode` enum / `NormalizeMode` / `Render(data, mode)` / Builder Config `Mode` field / `SetMode` unchanged — only the YAML+web surface removed, `renderFullV3` **not** resurrected); the dead `refresh_interval` key (zero consumers) and the dead `loadSegmentConfig` function (+ 6 test sites) are removed. **WIRE**: `preset` is now write-effective — a non-custom preset (`full`/`compact`/`minimal`) expands into a full 15-key segments map at save time via a new `internal/statusline/preset.go` SSOT (`PresetToSegments` + `CanonicalSegments`; the `internal/cli` `presetToSegments`/`allStatuslineSegments` became thin delegations), so a preset selection actually changes rendering (the runtime `segments`-wins precedence is **unchanged** — this is write-time materialization only); `defaultStatuslineSegments()` was completed from 11→15 keys. **FIX**: the 3 divergent structs converged — `Mode` dropped from the two private structs (`statuslineFileConfig`, `statuslineData`) while canonical `models.StatuslineConfig` stays `{Preset, Segments, Theme}` (no `Mode` added); the template theme seed `default`→`catppuccin-mocha` (only 2 real themes exist, `default` silently coerced to mocha + was non-round-trippable in the web dropdown); the `builder.go` `ThemeName` doc corrected to the 2 real themes + a `SegmentRepo` intentional-exclusion comment; the web segment checkboxes are now conditionally editable (only when `preset=custom`); and `StatuslineConfig` was added to the struct↔YAML symmetry CI guard (previously uncovered — the drift was invisible to CI). **HARD invariants held**: the 006 scope-boundary sentinel (`integration_test.go:197-205`) is **byte-identical** (git_convention / workflow / harness / git-strategy untouched); git_convention **0-diff** (deferred to 009); **zero network** (CDN grep 0); **zero Node** (`templ generate` codegen, drift-free); the theme-only round-trip characterization gate stays green (preset-absent saves still preserve the existing segments map). 22/22 AC PASS (AC-WC8-001..022); coverage internal/web 72.1% / statusline 83.0% / profile 80.5% (no regression vs 006/007 total-coverage baseline); cross-platform (darwin / windows amd64) exit 0; golangci-lint 0. plan-auditor iter-1 0.84 → 6 orchestrator-direct patches → iter-2 **0.87 PASS-WITH-DEBT** (Tier M threshold 0.80, MP-1..4 PASS). **008 of the web-console-v4 cohort** (builds on the 006 Templ tree + 007 nested write-seam).

- **[SPEC-WEB-CONSOLE-006](.moai/specs/SPEC-WEB-CONSOLE-006/spec.md)** — Web Console HTMX + Templ rendering migration: port-first, behavior-preserving (Tier L, v0.2.0, run-phase M1-M6 2026-06-05, sync-phase 2026-06-05). Migrates the `moai web` console rendering layer from a single ~339-line `internal/web/assets/page.html.tmpl` (`html/template` + a hand-rolled `dict` FuncMap) to **Templ (`github.com/a-h/templ` v0.3.1020) + an HTMX foundation**, porting the 5 fieldsets (Identity / Language / Launch / Statusline / Project) 1:1 with **zero new config section, zero settings.json field, zero observable behavior change**. (M1) Templ scaffolding + build wiring: `a-h/templ` dependency + `//go:generate` + Makefile `templ-generate` target + CI drift-guard (`templ generate` + `git diff --exit-code` in ci.yml + ci-mirror) — a pure-Go codegen step, **zero Node**; generated `*_templ.go` committed as source (bare-clone `go build` preserved; golangci-lint skips them via the `// Code generated by templ - DO NOT EDIT.` header — no `.golangci.yml` needed). (M2) ported `langSelect`/`optSelect`/`icon` helpers to typed Templ components. (M3) ported chrome + 5 fieldsets to a `page(view)` root, swapped the render call, retired `pageTemplate()`/`dict`. (M4) HTMX foundation: self-hosted `htmx.min.js` v2.0.4 via `go:embed` + linked before app.js + form `hx-boost="true"` progressive enhancement — **POST /save stays observably full-page** (no partial-swap; section-scoped `hx-target`/`hx-swap`/fragment `/save` deferred to **007**). (M5/M6) characterization-gate reconciliation + cross-platform verify. **The characterization gate is the spine**: the 7 Class B server-contract test files (`handlers_test.go`, `integration_test.go`, `validate_test.go`, `projectconfig{,_scope,_handler}_test.go`, `server_test.go`) stayed green **byte-unchanged** (= behavior preserved); `internal/web/validate.go` is **byte-unchanged** (`git diff --exit-code` clean); the 12 Class C source-coupled tests (which read the deleted `page.html.tmpl` source / called the retired `pageTemplate()`) were retargeted to the rendered HTTP body (11) or retired (1 — the pure symbol-existence `TestPageTemplateParses`, E.5.8 carve-out), with `grep -c 'readEmbeddedAsset(t,"page.html.tmpl")\|pageTemplate()' internal/web/*_test.go` = 0 (AC-WC6-019); Class A markup tests passed via exact-markup parity (the single relaxation was this SPEC's own additive smoke test for Templ's lowercase `<!doctype html>`). **Zero network** (offline grep for external font/style/script URL in served assets = 0; htmx self-hosted). Vanilla theme/i18n/segment-visibility JS preserved (not forced into HTMX). 20/20 AC PASS (AC-WC6-001a..019), hand-written coverage **90.9%** (= baseline; total 71.6% is `*_templ.go` codegen dilution only), cross-platform build (darwin / windows amd64) exit 0, `-race` clean, golangci-lint 0, templ drift-guard clean. plan-auditor iter-1 0.71 FAIL (caught a wrong characterization-test inventory) → Class C re-classification + 12-test verified inventory → iter-2 0.84 PASS-WITH-DEBT → doc-consistency debt fully resolved. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/web`. **006 of the web-console cohort**; its Templ component tree + HTMX foundation is the **enabler for 007 (S2b — nested config-section editing)**.

- **[SPEC-CC-DOCS-ALIGNMENT-001](.moai/specs/SPEC-CC-DOCS-ALIGNMENT-001/spec.md)** — Claude Code official-docs alignment: 33 rule/doc corrections across 5 doc surfaces (Tier M, v0.2.0, run-phase M1-M5 2026-06-03, sync-phase 2026-06-03). Corrects 33 alignment defects found by auditing 5 official Claude Code docs (workflows / skills / hooks-guide / goal / sub-agents) against the MoAI rule tree. **M1 workflows**: `dynamic-workflows.md` trigger keyword `workflow`→`ultracode` (renamed v2.1.160) + per-prompt-vs-`/effort` disambiguation + new "Manage runs" (`/workflows` TUI) and saved-workflow `args` docs + plan/provider availability + per-run approval note. **M2 skills**: removed the non-existent `type: skill` REQUIRED field from `skill-writing-craft.md` (0/26 actual SKILL.md use it); unified the contradictory description-length limits (80 / 250 / 1024) onto the official 1,536-char combined cap; documented `skillOverrides`, `name`-optional, nested/`--add-dir` discovery, bundled `/run` `/verify` skills, command↔skill merge, and the §13 Level-2 lifecycle wording. **M3 hooks (doc only)**: `if`-field version `v2.1.84`→`v2.1.85` + internal-consistency fix; Stop block-cap (`CLAUDE_CODE_STOP_HOOK_BLOCK_CAP`); 5s-is-MoAI-policy-not-platform-default reframing (LOCAL-ONLY `CLAUDE.local.md` + `internal/hook/CLAUDE.md`); `exec form` troubleshooting; Stop self-gate caveat. **M4 goal**: native `/loop` (time-interval) distinguished from `/moai loop`; auto-mode pairing; desktop/Remote-Control surfaces; `◎ /goal active` indicator; disable-scope precision; evaluator cost note. **M5 sub-agents**: `maxTurns` deprecation removed + phantom `maxContextSize` purged; `SubagentStart` hook event documented; fork subagents (`CLAUDE_CODE_FORK_SUBAGENT`); `claude-code-guide` built-in disambiguation (archived MoAI-custom ≠ live built-in — must not trigger `ARCHIVED_AGENT_REJECTED`); `model` full-ID intentional-divergence note; `delegate` MoAI-extension separation; `name`/`agent_type` semantics; managed-settings scope. Most targets are template-managed (source+mirror double-edit); the 4 neutrality-split files (`CLAUDE.md`, `agent-authoring.md`, `archived-agent-rejection.md`, `agent-common-protocol.md`) had the correction applied to both copies without forcing byte-parity (§25 template-neutrality preserved — `grep -rc 'SPEC-CC-DOCS-ALIGNMENT' internal/template/templates/` == 0). The single Go-code defect (3 missing hook `EventType` constants) was split out and shipped as the sibling SPEC-HOOK-EVENT-REGISTRY-001. 33/33 AC PASS (AC-CDA-001..033, grep-verifiable), `go test ./internal/template/...` green (neutrality + leak + mirror-parity gates), `make build` exit 0. plan-auditor 0.84 PASS-WITH-DEBT (Tier M LEAN threshold 0.80) → 4 minor debts (D1-D4) re-confirmed via grep at run time.

- **[SPEC-GO-DEPS-UPDATE-001](.moai/specs/SPEC-GO-DEPS-UPDATE-001/spec.md)** — Go third-party dependency maintenance update (Tier S, v0.1.1, run-phase 2026-06-03, sync-phase 2026-06-03). Patch/minor bumps of direct deps: `charmbracelet/x/powernap` v0.1.5→v0.1.6, `go-playground/validator/v10` v10.30.2→v10.30.3, `mattn/go-runewidth` v0.0.23→v0.0.24, `golang.org/x/sys` v0.44.0→v0.45.0; plus graph-required `golang.org/x/crypto` v0.51.0→v0.52.0 (indirect). **Zero major-version bumps** (no API-breaking change); `govulncheck ./...` affecting count stays **0** (security floor preserved); cross-platform build (host/windows/linux) green; diff limited to `go.mod` + `go.sum` (no source change). `x/net`/`x/tools`/`x/mod` stayed at tidy-minimal versions — `go mod tidy` reverts unused-indirect pins (correct minimal-tree behavior, sanctioned by acceptance EC-1); `charmbracelet/x/*` + `gorilla/websocket` excluded (parent-pinned / pure transitive via powernap→jsonrpc2, pin-skew avoidance per spec §Out of Scope). 9/9 AC PASS (AC-GDU-002 PASS-WITH-DEBT via the EC-1 tidy carve-out). plan-auditor 0.83 PASS-WITH-DEBT → D1 (grep `\|`→`|`) / D2 (websocket rationale) / D3/D4/D5 patched.

- **[SPEC-WEB-CONSOLE-002](.moai/specs/SPEC-WEB-CONSOLE-002/spec.md)** — `moai web` port default superseded from 8080 to **3041** (Tier S, v0.1.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). `internal/cli/web.go` `--port` flag default updated from `8080` to `3041`; Long help text updated to document the new default. Supersedes the candidate 8080 default introduced by SPEC-WEB-CONSOLE-001 REQ-WC-001. 9/9 AC PASS, coverage internal/web 90.8%. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

### Added

- **[SPEC-PREPUSH-WIRING-001](.moai/specs/SPEC-PREPUSH-WIRING-001/spec.md)** — Wire the dormant pre-push convention engine into the distributed git hook (Tier S, v0.2.0, plan + run 2026-06-08). The deployed pre-push hook (`internal/template/templates/.git_hooks/pre-push`, mirrored byte-identically into the `prePushHookContent` Go constant at `internal/cli/hook_install.go` and the dev-repo root `.git_hooks/pre-push`) previously ran **only** `make ci-local` and never invoked `moai hook pre-push`, leaving the fully-functional convention engine (`runPrePush` / `isEnforceOnPushEnabled`, gated by `enforce_on_push`) unreachable in shipped projects. This SPEC appends a **gated convention-validation block**: after `make ci-local` passes, the block captures git's ref-update stdin once (`REFS="$(cat)"`, before ci-local), translates each ref-update line into commit subjects via a `while read -r ...` loop (`git log --format=%s "$remote_oid".."$local_oid"`, with the new-branch `--not --remotes` and branch-delete `continue` zero-SHA edges handled), and pipes them to `moai hook pre-push` — all wrapped in a `command -v moai` guard so non-moai projects (16-language template neutrality) are unaffected. **Backward-compatible no-op by default**: the engine self-gates on `enforce_on_push`, which the template still ships `false` (NOT flipped — opt-in). Go engine **unchanged** (`internal/cli/hook_pre_push.go` 0-diff); the 3-way hook mirror stays byte-identical (`TestPrePushTemplateMatchesConstant` + root `diff` green) and `TestInstallPrePushHook_FreshRepo` now asserts the `moai hook pre-push` token. **Out of scope** (recorded as deferred): a `git_strategy.hooks.pre_push` (warn/enforce/skip) runtime reader — validation-only with zero runtime consumers; flipping the `enforce_on_push` default. 9/9 AC PASS (AC-PPW-001..009), full suite (`internal/cli` + `internal/template`) green, golangci-lint 0, cross-platform build (darwin / windows amd64) exit 0. plan-auditor iter-1 **PASS-WITH-DEBT 0.84** (Tier S threshold 0.75); D1 BLOCKING (AC-008 `ci-local` grep collided with the line-2 header comment) + D2/D3 patched orchestrator-direct; a run-phase independent-verification finding additionally corrected AC-005's `while read` grep to tolerate the shellcheck-recommended `-r` flag. **GCR-5 follow-up** of the web-console-008/009 git_convention cohort (the "dormant engine" residual).

- **[SPEC-PREPUSH-MODE-WIRING-001](.moai/specs/SPEC-PREPUSH-MODE-WIRING-001/spec.md)** — Wire the dormant `git_strategy.<mode>.hooks.pre_push` severity dial into the pre-push runtime (Tier S, v0.2.0, plan + run + sync 2026-06-08). The **2nd dead-config follow-up** to SPEC-PREPUSH-WIRING-001 (which wired the separate `git_convention.validation.enforce_on_push` gate). The per-mode `hooks.pre_push` string (∈ `skip | warn | enforce`; `HooksConfig.PrePush` at `internal/config/types.go:92`; template default `warn`) was validated for dynamic-token leakage only (`checkStringField` — **no enum gate**) and had **zero runtime consumers**. This SPEC adds the consumer as a **testability-seam pair** in `internal/cli/hook_pre_push.go`: a pure `resolvePrePushAction()` (resolves the precedence chain to a `prePushAction` enum) + a pure `decideExit(action, violationCount) int` (maps to the intended exit code), with `os.Exit(2)` confined to the thin `runPrePush` boundary so the decision is unit-testable **without** a process exit (the pre-existing `TestRunPrePush_WithViolations` is false-named — `readStdinLines` reads `/dev/stdin`, empty under `go test`, so it never reaches the violation path; no subprocess harness existed). **Precedence**: `env(MOAI_ENFORCE_ON_PUSH) > enforce_on_push` (master gate) `> pre_push` (severity dial); a new `MOAI_PRE_PUSH` env override (`EnvPrePushMode`) sets severity but sits **below** the gate (never opens it). **Semantics**: `skip` = no-op (`return nil`); `warn` = validate + print violations + **exit 0** (the new non-blocking middle state); `enforce` = validate + **exit 2** (blocking, the prior behavior). Fail-safe normalization is resolver-side (nil mode profile / unknown value → `enforce`), NOT a `validation.go` enum gate (explicitly out of scope). **Backward-compatible no-op by default**: with the shipped `enforce_on_push: false`, the gate short-circuits and `pre_push` is **never consulted** — SPEC-PREPUSH-WIRING-001 OFF behavior is preserved 100% (pinned by AC-PMW-013 against the existing `TestRunPrePush_EnforcementDisabled_ReturnsNilImmediately`). **Known limitation (recommended 3rd dead-config follow-up)**: the `git_strategy` config **section is not yet loaded from the user's `git-strategy.yaml` at runtime** — `internal/config/loader.go` has no `loadGitStrategySection`, and the only production assignment is the (uncalled) `ConfigManager.SetSection`, so `ActiveModeProfile().Hooks.PrePush` resolves to the compiled default `warn` regardless of the YAML file. The consumer wired here is correct per the SPEC's AC contract (verified via the supported `SetSection` injection path), but **end-to-end YAML editing requires a follow-up** (`SPEC-PREPUSH-LOADER-WIRING-001`) to wire `git_strategy` into the loader chain. 13/13 AC PASS (AC-PMW-001..013), full `internal/cli` + `internal/config` suite green, the pure helpers `resolvePrePushAction` / `parsePrePushAction` / `decideExit` 100% covered, golangci-lint 0, cross-platform build (darwin / windows amd64) exit 0; `validation.go` / `types.go` / `defaults.go` / templates / shell-hook untouched (scope discipline held). plan-auditor iter-1 **0.84 PASS-WITH-DEBT** → iter-2 (run Phase 0.5 re-audit) **0.89 PASS** (Tier S threshold 0.80; monotonic +0.05; Testability 0.62→0.92); D1/D4 ground-truth citation fixes + D2 testability seam (REQ-PMW-002a) + D3 gate-OFF regression pin all resolved & confirmed-real. **2nd dead-config follow-up of the PREPUSH-WIRING line.**

- **[SPEC-PREPUSH-LOADER-WIRING-001](.moai/specs/SPEC-PREPUSH-LOADER-WIRING-001/spec.md)** — Wire the `git_strategy` config section into the loader READ path (Tier S, v0.2.0, plan + run + sync 2026-06-10). The **3rd (terminal) dead-config follow-up** in the PREPUSH chain (PREPUSH-WIRING-001 `enforce_on_push` shell-hook → PREPUSH-MODE-WIRING-001 `pre_push` consumer → this loader PRODUCER). `internal/config/loader.go` `Load()` invoked 14 per-section loaders but had **no `loadGitStrategySection`**, so `cfg.GitStrategy` retained the compiled defaults regardless of the user's `git-strategy.yaml` — the root cause that left SPEC-PREPUSH-MODE-WIRING-001's `resolvePrePushAction()` reading the compiled default `warn` instead of the YAML's value. This SPEC adds the READ path as an exact mirror of `loadGitConventionSection`: a new `gitStrategyFileWrapper` struct (`internal/config/types.go`, top-level key `git_strategy:` wrapping the local-package `GitStrategyConfig`) + a `loadGitStrategySection` method (`internal/config/loader.go`) + the wired `Load()` call (line 59, adjacent to the git-convention loader). The non-strict `yaml.Unmarshal` + defaults-seeded wrapper give partial-override and unknown-key resilience for free. **End-to-end chain completion** (AC-PLW-007): a fixture with `mode: team` + `team.hooks.pre_push: enforce` now yields `cfg.GitStrategy.ActiveModeProfile().Hooks.PrePush == "enforce"` after `Load()` (previously the compiled default `warn`) — the user's `git-strategy.yaml` is finally observable at runtime, closing the PREPUSH dead-config chain end-to-end. **READ-only scope** (user-approved): the WRITE/Save path (`Save()` persisting git-strategy.yaml) is explicitly deferred — `grep -c 'git-strategy.yaml' internal/config/manager.go` stays **0** (AC-PLW-008); the in-memory `SetSection`/`GetSection` git_strategy case, `validation.go`, `defaults.go`, templates, and the 2nd SPEC's `hook_pre_push.go` are untouched. 8/8 AC PASS (AC-PLW-001..008) + 2 edge cases (empty `git_strategy:` block / invalid YAML → `slog.Warn` + keep defaults, no error propagation), `go test ./internal/config/...` green, `loadGitStrategySection` 100% statement coverage, golangci-lint 0, cross-platform build (darwin / windows amd64) exit 0. plan-auditor iter-1 **0.88 PASS** → iter-2 (run Phase 0.5 re-audit) **0.92 PASS** (Tier S threshold 0.80; monotonic +0.04; Completeness + Traceability 1.00; "cleanest grep-AC set in the PREPUSH cohort"); 3 MINOR (D1 loader count 15→14, D2 GEARS label Event-detected→Event-driven, D3 regression over-statement softened) resolved orchestrator-direct & confirmed-real. **Closes the PREPUSH dead-config chain (3/3).**

- **[SPEC-WEB-CONSOLE-007](.moai/specs/SPEC-WEB-CONSOLE-007/spec.md)** — Web Console nested config editing for `quality` + `git_convention` sections, scoped S2b (Tier M, v0.2.0, run-phase M1-M6 2026-06-07, sync-phase 2026-06-07). Extends the `moai web` console — building on the 006 Templ component tree + HTMX foundation — to edit a **curated nested-field set of exactly two project-config sections**: `quality.{test_coverage_target, enforce_quality, tdd_settings.min_coverage_per_commit}` + `git_convention.{auto_detection.confidence_threshold, auto_detection.enabled, custom.pattern}` (plus the retained `convention` scalar from 003). (M1) two reusable Templ widgets — `toggle` (bool checkbox + hidden `__present` companion for EC-1 bool semantics) + `numberField` (`<input type="number">` with min/max/step) — each with a Class A markup-parity test. (M2) **config-validation export seam with ZERO new validator functions** (CRITICAL SCOPE CONSTRAINT): `ValidateQualitySection` / `ValidateGitConventionSection` thin exported wrappers forward verbatim to the existing unexported `validateQualityConfig` / `validateGitConventionConfig` — `grep -cE 'func validate(Workflow|GitStrategy|Harness|Llm)Config'` stays **0**. (M3) explicit dot-path `PostFormValue` nested-form parsing (no reflection path-walker) + view-model echo-back read seam + `fieldsetProject` widget render + 12 i18n keys ×4 locales. (M4) **load-modify-write nested isolation** (HARD-4): writing one nested field copies the whole section struct and mutates only the `*Set`-gated target, leaving sibling nested fields (incl. nested-of-nested `tdd_settings` / `auto_detection`) **byte-identical** — proven by `TestProjectNested*SiblingPreserved`; plus EC-1 (empty=preserve) + bool companion + EC-2 (atomic reject — one invalid field ⇒ no section written, HTTP 400 + per-field `FieldErrors`) + server-canonical out-of-range / custom-required rejects reusing the existing validator messages. (M5/M6) regression guard + 4-locale i18n parity + offline re-verify. **HARD invariants held**: the 006 scope-boundary sentinel (`integration_test.go:197-205`) is **byte-identical** (`git diff` = 0 — no 008-scope creep), `POST /save` stays **hx-boost full-page** (no partial-swap fragment — deferred to 008), **zero network** (CDN grep 0), **zero Node** (`templ generate` codegen only, drift-free), no direct YAML marshal in `internal/web`. **Deferred to SPEC-WEB-CONSOLE-008**: workflow/git-strategy/harness/llm nested editing (REQ-WC-012 boundary lift + new validators + sentinel retarget), partial-swap fragments, dynamic section registry. 20/20 AC PASS (AC-WC7-001..020); coverage internal/web **72.5% total** (= 006 baseline 71.6% **+0.9%**, no regression — the acceptance "90.9%" literal is 006's *hand-written* coverage convention, which `go test -cover` dilutes with the generated `*_templ.go`; new 007 hand-written code is ~90-95% covered). plan-auditor iter-1 (plan) 0.83 → iter-2 (run Phase 0.5 re-audit) **0.84 PASS-WITH-DEBT** (Tier M threshold 0.80); D1 BLOCKING (§F `Exclusions` h2 failed spec-lint `MissingExclusions`) resolved by an `### §F.1 Out of Scope` h3 restructure (orchestrator-direct mechanical patch). Cross-platform build (darwin / windows amd64) exit 0, golangci-lint 0. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/web` / `internal/config`. **S2b of the web-console-v4 cohort** (006 is its enabler).

- **[SPEC-HOOK-EVENT-REGISTRY-001](.moai/specs/SPEC-HOOK-EVENT-REGISTRY-001/spec.md)** — Hook event registry: 3 observe-only Claude Code events added to close doc-ahead-of-code drift (Tier S, v0.2.0, run-phase 2026-06-03, sync-phase 2026-06-03). Adds `PostToolBatch`, `UserPromptExpansion`, `MessageDisplay` as **observe-only** `EventType` constants + `CoverageTable` inventory rows (`Resolution: RETIRE-OBS-ONLY`, `IsActive: false`, `HandlerFile: ""`) in `internal/hook/{types.go,coverage_table.go}` — no `Handle()`, no `settings.json` registration, no blocking behavior, no CLI subcommand. The MoAI doc layer (`.claude/rules/moai/core/hooks-system.md`) already documented these three events; this SPEC closes the Go-code lag only (a sibling doc SPEC handles the documentation-alignment defects). `ValidEventTypes()` 26→29; `CoverageTable` 27→30 rows (the +1 over events is the synthetic `AutoUpdate` COMPOSITE row); `coverage_table.go` header comment + `@MX:ANCHOR` text `26-event`→`29-event` (ANCHOR updated, not deleted). The settings/Go/retired 3-way sync guard (`TestAuditThreeWaySync`) is preserved by appending the 3 names to the **test-local** `deregisteredButLiveEventNames` allowlist — the production-exported `RetiredEventNames` (consumed by `internal/migrate`) is untouched. Cross-package coupled-test assertions updated exhaustively: `internal/hook/types_test.go` (count + membership + `TestCoverageTableLen`), `internal/cli/doctor_hook_test.go` (×3: non-composite 26→29, `RetireObsOnly` 4→7, observability filter 4→7), `internal/cli/hook_e2e_test.go` (×2: `ValidEventTypes` 26→29, `excludedEvents` += 3 — observe-only, NOT `eventToSubcmd`). 28/28 AC PASS (AC-HER-001..009), `go test ./internal/hook/... ./internal/cli/... -count=1` green, `go vet` + golangci-lint `0 issues`, cross-platform build (darwin / windows amd64) exit 0. plan-auditor 0.71 → 0.79 → **0.94 PASS** (iter-3; an empirical full-suite simulation confirmed the coupled surface is complete — no 8th coupled site). Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-WEB-CONSOLE-005](.moai/specs/SPEC-WEB-CONSOLE-005/spec.md)** — Web Console interface i18n (en/ko/ja/zh) + CJK self-host webfont coverage, zero server-contract change (Tier M, v0.2.0, run-phase M1-M4 2026-06-03, sync-phase 2026-06-03). The web-console-v3 cohort **terminator** (S3) — adds client-side interface translation + the CJK webfont coverage that 004 deferred, while preserving the entire server contract (`internal/web/validate.go` byte-unchanged; the interface language is client-only and mutates NO server-submitted field — the cohort's interface≠content-language invariant). (M1) CJK font layer: `pyftsubset` Noto Sans CJK SC (zh) + JP (ja), 3 weights each, subset to EXACTLY the shipped ja/zh dictionary glyph set (measured **279** unique glyphs: 197 CJK ideographs + 44 katakana + 29 hiragana + 9 CJK punctuation — 100% coverage, 0 over-coverage), ~235KB total + OFL-1.1, self-hosted via `go:embed`, activated by `html[lang="ja|zh"]` font-stack override (en/ko stay on the 004 Pretendard subset). (M2) i18n dictionary + wiring: new client-side `internal/web/assets/i18n.js` (4-locale STATIC dictionary, derivative of the design source with the `rv.*` design-review keys stripped) + `data-i18n` attributes on 36 chrome elements (code chips NOT tagged) + go:embed enumeration. (M3) appbar langpick (server-contract gate): a `<select id="uiLangSelect">` interface-language picker in the appbar — **outside `<form>`, NO `name=`** — with `applyI18n()` + `localStorage("moai-console-lang")` persistence + load-time apply + `<head>` FOUC snippet + `<html lang>` update; default `en`. (M4) test reconciliation: inverted the 004 `TestAppbarRendered` S3-exclusion guard (langpick/`data-i18n`/`uiLangSelect` forbidden→expected) + a11y (aria-label, `<html lang>`). **Zero network** (offline grep for `fonts.googleapis.com`/`unpkg`/external font/style/script URL = 0). MUST-PASS verified: POST round-trip byte-identical regardless of interface language (interface≠content), `validate.go` byte-unchanged, offline. plan-auditor 0.84 PASS-WITH-DEBT → D1 (`uiLangSelect` id-collision rename — avoids the live `langSelect` content-language helper) / D2 (glyph-count magic-number → reproducible recipe + shipped-dict binding) / D3 (`data-i18n` ≥25 floor) resolved. 13/13 AC PASS (AC-WC5-001..013), cross-platform build (darwin / windows amd64) exit 0, golangci-lint 0, `go vet` clean. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/web`. **Closes the web-console-v3 cohort** (001 mother / 002 port+validation / 003 flat config / 004 visual restyle / 005 i18n+CJK font).

- **[SPEC-WEB-CONSOLE-004](.moai/specs/SPEC-WEB-CONSOLE-004/spec.md)** — Web Console 모두의AI design-system application: visual restyle, zero server-contract change (Tier M, v0.2.0, run-phase M1-M5 2026-06-03, sync-phase 2026-06-03). Applies the 모두의AI design system to the `moai web` loopback console as a pure visual restyle while preserving the entire server contract (`internal/web/validate.go` byte-unchanged; every `name=` attr, `{{range}}` server-rendered option, and `.FieldErrors` server-side render retained — REQ-WC4-009 MUST-PASS invariant). (M1) token+font: ported the 모두의AI token layer (`:root` + `[data-theme="dark"]` + base typography) into a new embedded `internal/web/assets/console.css`, removed the Google-Fonts `@import`, and self-hosted a Pretendard **woff2 subset** (Latin + brand Hangul, 5 used weights ~100KB total — not the 9MB OTF set) + OFL-1.1 license under `internal/web/assets/fonts/`, extending the `assets.go` go:embed directive; (M2) layout+component port (server-contract gate): restyled `page.html.tmpl` (appbar + pagehead + profilebar + banner + 5 fieldsets + actions) with console.css chrome, extended the langSelect/optSelect define blocks with title/key/desc field-chrome params (structure preserved), added a `BindAddr` view-model field wired to `Server.displayBindAddr()` for the real loopback indicator (REQ-WC4-005), mapped `.BannerKind` ok/error → banner--success/banner--error template-locally (server kind values unchanged, REQ-WC4-010), and dropped the design's non-canonical options (no `es/fr/de`, no `haiku[1m]`, no kebab segment keys; option lists stay `{{range}}`-driven); (M3) dark mode: appbar sun/moon `#themeToggle` + client `localStorage` persistence + FOUC-prevention inline `<head>` init, prefers-reduced-motion guard preserved (no server round-trip); (M4) inline-SVG icon subset (~12 icons via a template `icon` helper, lucide CDN `<script>` removed, lucide ISC acknowledged); (M5) accessibility: non-color error cues (icon+border+text), focus-visible outlines, aria-label on the theme toggle, aria-invalid/aria-describedby on errored fields. **Zero network** achieved (offline regression grep for `fonts.googleapis.com`/`unpkg.com`/external font/style/script URL = 0 matches). plan-auditor 0.89 PASS; D1 (SHOULD-FIX) resolved in run by strengthening the AC-WC4-008 non-canonical-option regression to a structural `segment_[a-z]+-[a-z]` grep (covers all 15 segment keys, not 3 literals). 13/13 AC PASS (AC-WC4-001..013), coverage internal/web 90.9%, cross-platform build (darwin / windows amd64) exit 0, golangci-lint 0, `-race` pass. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/web`. 004 of the web-console-v3 cohort (extends 001/002/003); web interface i18n + appbar language picker + full CJK webfont deferred to S3, nested config to S2b.

- **[SPEC-CCSYNC-DYNWF-001](.moai/specs/SPEC-CCSYNC-DYNWF-001/spec.md)** — Dynamic Workflows doctrine alignment (doc seams) (Tier S, v0.2.0, documentation-only, 8/8 Blocking ACs PASS). Four documentation seams closed between the Dynamic Workflows feature surface and MoAI doctrine: (REQ-1) `dynamic-workflows.md § How a Workflow Runs` states the determinism constraint (script body must not call wall-clock/random; timestamps/randoms injected via args or stamped post-run) enabling resume cache. (REQ-2) `CLAUDE.md § 10` + `moai-domain-research/SKILL.md § Works Well With` cross-reference the `/deep-research <question>` workflow path for multi-source fan-out + cross-check + claim-vote research, carrying 3 facts (WebSearch tool required, higher token cost, AskUserQuestion boundary holds). (REQ-3) `dynamic-workflows.md § When to Use a Dynamic Workflow` provides a routing heuristic with quantitative anchors (dozens-to-hundreds items → workflow; small coordinated peers → Agent Teams; default → sequential subagents; Finding A4 caveat for coding-heavy work). (REQ-4) `dynamic-workflows.md § MoAI Integration Notes` augments the existing ultracode bullet to state `ultracode` resets on a new session and is NOT restored by the `ultrathink.` resume opener (which restores reasoning only); a resumed session must explicitly re-issue `/effort ultracode`. All target files (3 working copies + 3 template mirrors) mirror-parity PASS. No SPEC-ID leak; template neutrality + internal-content leak guards green. Template test suite `go test ./internal/template/...` all PASS (neutrality + leak + catalog-hash-parity regression gates). 8/8 AC PASS (AC-DYNWF-001..008). Cross-platform build exit 0. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-WEB-CONSOLE-001](.moai/specs/SPEC-WEB-CONSOLE-001/spec.md)** — MoAI Web Console: browser-based settings CRUD (Tier M, v0.2.0, run-phase M1 2026-06-03, sync-phase 2026-06-03). New `moai web` subcommand launching a loopback-bound (127.0.0.1) HTTP server that serves a browser settings editor for MoAI profile preferences: READ (`handleIndex`) + WRITE (`handleSave`) handlers, Host-check middleware (DNS-rebinding guard), canonical-value validation (`validatePrefs` reusing `profile.IsValidPermissionMode`), graceful shutdown, and a cross-platform default-browser opener. New package `internal/web` (server.go / browser.go / app.go / handlers.go / validate.go / assets.go + `go:embed` page.html.tmpl / style.css / app.js) + thin `internal/cli/web.go` cobra subcommand registered via `root.go` + `help.go` Launcher row. Supersedes SPEC-V3R3-WEB-001. Mother SPEC of the web-console-v3 cohort (extended by 002 port/validation, 003 project-config parity, 004 design). 12/12 AC PASS + 5 edge cases, coverage internal/web 90.5%, cross-platform build (darwin / windows / linux amd64) exit 0. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-WEB-CONSOLE-003](.moai/specs/SPEC-WEB-CONSOLE-003/spec.md)** — Web Console flat project-config parity: `development_mode` + `git_convention.convention` web + TUI editing (Tier M, v0.2.0, run-phase M1-M5 2026-06-03, sync-phase 2026-06-03). Extends the web console and the TUI profile wizard to read, validate, and persist two FLAT project-config settings that previously had no editor surface: `development_mode` ({ddd, tdd} via exported `models.ValidDevelopmentModes()`) and `git_convention.convention` ({auto, conventional-commits, angular, karma, custom} via new exported `config.IsValidConvention` SSOT). (M1) canonical convention predicate; (M2) `internal/config` project-config validator + read/write seams (config-manager `LoadRaw`→`SetSection`→`Save`, NOT `ProfilePreferences`); (M3) `internal/web` POST /save handler wiring + `<select>` widget select-ification (rejecting out-of-list submissions with per-field `FieldErrors` + HTTP 400, persisted state unchanged); (M4) TUI `profile_setup.go` development_mode + git_convention `huh.Select` parity with 4-locale labels; (M5) section-isolation guard (workflow / harness / git-strategy byte-identical preservation). `llm.mode` / `llm.default_model` NARROWED OUT (no canonical enum to reuse). 12/12 AC PASS (AC-WC3-001a..009 + EC-1..6), coverage internal/web 90.9%. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations. S2a of the web-console-v3 cohort (extends SPEC-WEB-CONSOLE-001/002).

- **[SPEC-GLM-WEBTOOL-ROUTING-001](.moai/specs/SPEC-GLM-WEBTOOL-ROUTING-001/spec.md)** — GLM-backend web tooling routing doctrine + per-tool z.ai MCP registration (Tier M, v0.2.0, run-phase M1-M6 2026-06-03, sync-phase 2026-06-03). Establishes `.claude/rules/moai/core/glm-web-tooling.md` as the single source of truth for how MoAI agents route web search / web fetch / image read under a GLM backend (`moai glm` or `moai cg` GLM teammate panes): the built-in `WebSearch` / `WebFetch` / `Read`-on-image (529-prone z.ai Anthropic-gateway + base64→422 image path) are replaced by `mcp__web_search_prime__webSearchPrime` / `mcp__web_reader__webReader` / `mcp__zai-mcp-server__*` vision tools, with a `moai cg` leader-pane exception. (M1) routing doctrine + zone-registry `CONST-V3R5-040`/`041` HARD clauses; (M2) `internal/cli/glm_tools.go` per-tool registration refactor — `moai glm tools enable [vision|websearch|webreader|all]` now registers the correct server per tool-name (previously `enable webreader` wrongly registered only the npx vision server), with accurate success/disable messages and the Node-version gate scoped to vision only; (M3) dev `.mcp.json` three separate z.ai server entries + `Z_AI_API_KEY`; (M4) concise `glm-web-tooling.md` pointers added to the six cross-link reference points (agent-common-protocol.md, moai-constitution.md, settings-management.md, moai-domain-research SKILL.md, einstein.md, CLAUDE.md §10/§12) — pointer only, no routing-table duplication; (M5) template↔local mirror parity reconciled. 24/24 AC PASS (AC-GWR-001..024); invariants preserved (atomic-write / backup / idempotency / token-mismatch / Node-gate). Template neutrality green (C1/C2/C3/C7). Predecessor: SPEC-GLM-MCP-001. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/cli`.

- **[SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001](.moai/specs/SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001/spec.md)** — Harness activation wiring restoration (Tier M, v0.1.0 → v0.2.0, run-phase M1-M6 2026-06-03, sync-phase 2026-06-03). Restores the orphaned auto-trigger path that connected meta-harness generation (`InjectMarker` / `ScaffoldHarnessDir`) to project activation. (M1) Phase 0.95 mode-selection RED baseline; (M2/M3) new `moai harness install` CLI (`internal/cli/harness/install.go`) + Phase 7 wiring (`internal/cli/harness_route.go`); (M4) `project/meta-harness.md` main.md router + generated-agent self-activation (`internal/harness/layer5.go` `mainMD()`); (M5) Phase-6 smoke gate extending `doctor harness` (`internal/cli/doctor_harness.go`) that fails when a generated agent omits its `skills:` preload (AC-HAW-015, runtime enforcement of REQ-HAW-008); (M6) retrofit note + run-phase audit-ready signal. The generated-harness prefix stays `my-harness-*` (plan-audit D1/D2 boundary fix); the smoke gate preserves L1-L5 `doctor harness` invariants. 17/17 AC PASS (AC-HAW-001..015 with 013b folded into 015 + PROC-1..2), 0 AC fail. Cross-platform build (darwin/windows amd64) exit 0; `go vet ./...` clean. Skill template mirror (`internal/template/templates/.claude/skills/...`) parity maintained. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations in `internal/harness` / `internal/cli/harness`.

- **[SPEC-WEB-CONSOLE-002](.moai/specs/SPEC-WEB-CONSOLE-002/spec.md)** — Web Console validation parity + widget select-ification + TUI model_policy select (Tier S, v0.1.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). Three correctness gaps closed: (1) **Validation parity** — `internal/web/validate.go` `validatePrefs()` now validates `model` (6 canonical values via mirrored wizard list), `effort_level` (5 canonical values), and `model_policy` (wires `template.IsValidModelPolicy()`) — rejecting out-of-list POST /save submissions with per-field `FieldErrors` and HTTP 400, leaving persisted state unchanged; (2) **Widget select-ification** — `internal/web/assets/page.html.tmpl` replaces the three `<input type="text">` widgets for `model`/`effort_level`/`model_policy` with `<select>` dropdowns bound via `newPageView()` option lists; (3) **TUI model_policy parity** — `internal/cli/profile_setup.go` model-settings group gains a `model_policy` `huh.Select` (options: project default / high / medium / low) with 4-locale labels in `internal/cli/profile_setup_translations.go` (en/ko/ja/zh). `model_policy` persists profile-only (`preferences.yaml` via `WritePreferences`); `SyncToProjectConfig` scope unchanged (user/language/statusline only); no new config section introduced. 9/9 AC PASS (AC-WC2-001..009), coverage internal/web 90.8%. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

### Fixed

- **[SPEC-V3R6-HOOK-INPUT-SCHEMA-001](.moai/specs/SPEC-V3R6-HOOK-INPUT-SCHEMA-001/spec.md)** — hook input schema robustness (Tier S, v0.1.0, run-phase 2026-06-03, sync-phase 2026-06-03). Fixes two `internal/hook` stdin-parse defects surfaced from `~/.moai/logs/hook-stderr.log`: (1) **globs schema drift** — `HookInput.Globs` changed `string`→`[]string` so the InstructionsLoaded event's array-form `globs` deserializes natively (was `cannot unmarshal array into …string`, failing `moai hook instructions-loaded` on every rule-glob load); (2) **empty-stdin robustness** — `ReadInput` (`internal/hook/protocol.go`) now treats empty/blank/whitespace-only stdin as a graceful no-op success instead of `ErrHookInvalidInput`, so a PreToolUse(Bash) hook with empty stdin no longer surfaces as `PostToolUseFailure: Bash UnknownFailure`. 4/4 mandatory AC PASS (incl. AC-2 inversion of the 2 prior empty/whitespace `protocol_test.go` cases + AC-4 `moai hook pre-tool </dev/null` exit-0 smoke). Non-empty-malformed JSON still returns `ErrHookInvalidInput` (scope preserved — D2 confirmed the live failure was empty, not truncated). Coverage internal/hook 81.8%. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001](.moai/specs/SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001/spec.md)** — `moai spec drift` false-positive elimination + era/grandfather alignment (Tier M, v0.1.1, run-phase M1-M5 completion 2026-06-03, manager-docs sync-phase 2026-06-03). Aligns the drift detector with the audit engine's era-awareness and terminal-state authority policies, fixing 4 false-positive mechanisms: (①) combined-scope close subject recognition via secondary prefix-grep fallback (REQ-DLC-001/002) + 3-gate word-boundary LSGF-001 guard; (②) stale sync→completed rule correction to modern 4-phase model (REQ-DLC-003/004); (③) terminal-state frontmatter as authoritative (REQ-DLC-005); (④) grandfather/legacy era exemption via ClassifyEra + LoadEraSignalsFromDir read-only consumption (REQ-DLC-006). **Drift count 54→8** (strict decrease per AC-DLC-009); audit.go / era.go preserved byte-identical (PRESERVE verification); 12/12 AC PASS. Genuine-incomplete-close mechanism (⑤) classified as deferred operational follow-up (NOT SCOPE per spec.md §C Out-of-Scope). **Close-subject doctrine amendment** (REQ-DLC-011): lifecycle-sync-gate.md + spec-frontmatter-schema.md mandates individual full-IDs in close-commit subjects, prohibiting combined/abbreviated scope to prevent recurrence. Template mirror obligation satisfied (§25 C2 generalized per coding-standards.md acceptable-content-range). Named exemplars: SPEC-CCSYNC-CLAUDEMD-001 [completed→false via combined-scope fallback], SPEC-CCSYNC-TOOLCAT-001 [completed→false], SPEC-V3R5-STATUSLINE-STDINFIELDS-001 [implemented→false via stale-rule fix], SPEC-V3R3-HARNESS-001 [superseded→false via terminal authority], SPEC-V3R2-SPC-002 [V3R2 era→false via grandfather exemption]. Coverage 87.6% internal/spec. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

### Added

- **[SPEC-V3R5-GIT-STRATEGY-SCHEMA-001](.moai/specs/SPEC-V3R5-GIT-STRATEGY-SCHEMA-001/spec.md)** — git-strategy.yaml ↔ Go struct alignment (Tier M, v0.1.0 → v0.2.0, run-phase completion 2026-06-02, manager-docs sync-phase 2026-06-03). Nested GitStrategyConfig with 6 sub-structs (GitLabConfig / AutomationConfig / BranchCreationConfig / CommitStyleConfig / HooksConfig / ModeProfile) + ActiveModeProfile() accessor + backward-compat FLAT field preservation (Option c, per spec.md §1.4) + validation extension (`checkStringField` nested paths). 5 milestones M0-M5 (M0 struct hierarchy + deprecation comments, M1 roundtrip test + accessor test, M3 defaults, M4 validation + audit registry cleanup, M5 cross-platform verification). 9/9 must-pass ACs PASS. Implementation files (5 modified: types.go / defaults.go / validation.go / audit_loader_completeness_test.go / git_strategy_nested_test.go [NEW]) deliver yaml↔struct parity for ~70 nested keys per template SSOT, enabling Late-Branch skill-body consumers + future Go runtime wire-through without blocking current skill-yaml-direct consumption.

### Removed

- **[SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001](.moai/specs/SPEC-V3R6-MEMORY-CONFIG-CLEANUP-001/spec.md)** — Remove inert MemoryConfig config schema from `workflow.yaml` binding (Tier S, v0.1.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). Gap A: MemoryConfig schema (`config.MemoryConfig` type + `workflow.memory.*` YAML binding) had zero production consumers — editing `.moai/config/sections/workflow.yaml` `memory:` block had no runtime effect. Removal makes the deployed schema honest about what is actually wired. Gap B: deferred non-issue — latent audit index byte-cap checks were never production-wired per `.claude/rules/moai/workflow/moai-memory.md` disclosure; decision recorded in spec.md §B.2 + §E with zero code change (EXCL-MCC-001/002). Gap C: documented per-cwd memory-dir resolution is intentional and aligned with Claude Code's own per-cwd memory model (comment-only additions to `internal/hook/session_end.go` / `internal/hook/handoff/persist.go` + `.moai/docs/memory-dir-resolution-doctrine.md` NEW; doctrine note on L3 `--worktree` Block 0 re-anchoring). 10/10 Blocking AC PASS (AC-MCC-001..009 + full suite regression green). Files modified: 4 config (types.go, defaults.go + 2 config/ test files) + 2 hook (session_end.go, persist.go) + 1 NEW doctrine. Cross-platform (linux/darwin/windows) build exit 0. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-LSPMCP-RETIRE-001](.moai/specs/SPEC-LSPMCP-RETIRE-001/spec.md)** — Retire dormant moai-lsp MCP bridge (Tier S, v0.1.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). `internal/mcp` package (10 files: mcp.go, lsp.go, lsp_test.go, start.go, start_test.go, stop.go, config.go, logging.go, error.go, types.go), `internal/cli/mcp.go` moai mcp lsp subcommand, `.mcp.json.tmpl` template moai-lsp entry, `internal/config/settings_test.go` 2 functions (TestMCPServerFields_LSP + TestMCPServerInheritance_LSP). Run-phase 10/10 AC PASS. Rationale: moai-lsp MCP bridge was a proof-of-concept bridge to Claude Code's LSP integration prior to native LSP support (lsp.go `OpenServerStdin`, `OnRequest`, `OnNotification` handlers are never invoked in production — zero active integration). Native `internal/lsp/` (real LSP suite) supersedes this bridge and is preserved. No user-facing features lost; the bridge was never exposed as a CLI-adjacent product feature (always internal-only prototype). Reference: `.moai/specs/SPEC-LSPMCP-001/spec.md` predecessor (archived, superseded marker). Post-removal, zero moai-lsp / moai mcp lsp string references in docs-site / README / settings template (verified by grep). Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations. 10/10 AC PASS: AC-LSMCP-001 files removed ✓ / AC-LSMCP-002 subcommand removed ✓ / AC-LSMCP-003 settings_test.go functions removed ✓ / AC-LSMCP-004 template entry removed ✓ / AC-LSMCP-005 docs zero moai-lsp refs ✓ / AC-LSMCP-006 internal/lsp preserved ✓ / AC-LSMCP-007 cross-platform build exit 0 ✓ / AC-LSMCP-008 go test ./... exit 0 ✓ / AC-LSMCP-009 golangci-lint 0 NEW issues ✓ / AC-LSMCP-010 successor SPEC-LSPMCP-001 marked superseded ✓.

### Fixed

- **[SPEC-CCSYNC-CLAUDEMD-001](.moai/specs/SPEC-CCSYNC-CLAUDEMD-001/spec.md)** — CLAUDE.md v14.1.0 → v14.2.0 template sync-phase (Tier S, documentation-only sync, 17/17 Blocking ACs PASS). Ports dev-root CLAUDE.md v14.2.0 content (§2 Section 2 agent-routing, §5 Agent Catalog, §11 Error Handling) to template copy, removing 12 archived-agent invocation examples per SPEC-V3R6-AGENT-TEAM-REBUILD-001 consolidation (17→8 agent catalog). Corrects context-window thresholds (1M=50% per context-window-management.md SSOT), Agent Teams version (v2.1.32, not v2.1.50), teammate-spawn capability (subagents cannot spawn, per Anthropic guidance), and dead reference (progressive-disclosure.md → skill-authoring.md). Template neutrality audit PASS (no internal SPEC IDs / REQ / dates / SHAs). Mirror drift resolved for 4 files (settings-management.md, agent-authoring.md, CLAUDE.md template + dev-root). `go test ./internal/template/...` green (neutrality + mirror-drift + internal-content leak guards all PASS). AC-CCSYNC-001/002/003/004/005 (archived-agent sync + version sync) / AC-CCSYNC-006 (neutrality audit) / AC-CCSYNC-007 (no internal tokens) / AC-CCSYNC-008/009/010/011 (context-window threshold SSOT alignment) / AC-CCSYNC-012/013/017 (v2.1.32 Agent Teams version) / AC-CCSYNC-014/015 (teammates spawn + dead ref fixed) / AC-CCSYNC-016 (mirror + build) all PASS.

- **[SPEC-CCSYNC-TOOLCAT-001](.moai/specs/SPEC-CCSYNC-TOOLCAT-001/spec.md)** — Tool catalog audit: MultiEdit deprecated, TodoWrite → Task* migration (Tier S, documentation-only sync, 17/17 Blocking ACs PASS). Removes MultiEdit from 2 agents' `tools:` declarations (manager-spec / manager-develop) plus manager-spec body instructions, manager-develop hook matchers, and the settings.json.tmpl PostToolUse matcher + permissions array, rewords manager-spec body to use parallel Edit/Write instead of MultiEdit, migrates 5 implementation agents' TodoWrite → TaskCreate / TaskUpdate / TaskList / TaskGet (manager-spec / manager-develop / manager-docs / manager-git / builder-harness; plan-auditor and sync-auditor remain read-only, no Task* tools). Updates agent-authoring.md recommendations to Task* family and scrubs MoAI authoring-kit examples (while preserving 3 verbatim official-docs reproductions per exclusion). Adds `internal/template/tool_catalog_audit_test.go` (NEW 220L test) validating MultiEdit absence + retained-agent TodoWrite absence. Mirror parity maintained across 12 edited files; `make build` regenerates the embedded templates (this project embeds via `//go:embed all:templates` directly — there is no generated `embedded.go` file). Template neutrality and internal-content leak guards PASS. AC-CCSYNC-T-001/002/003/004/005/006/007 (tool-catalog sync + agent-authoring updates) / AC-CCSYNC-T-008 (NEW CI guard) / AC-CCSYNC-T-009/010/011 (mirror + build + neutrality) all PASS. Sibling SPEC-CCSYNC-CLAUDEMD-001 committed first; agent-authoring.md line ~212 merged without conflict per sequencing.

- **[SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001](.moai/specs/SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001/spec.md)** — moai spec drift false-positive fix: canonically-closed SPECs now resolve to git-implied completed (Tier S, internal/spec, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). transitions.go close-infix check + drift.go narrow backfill-skip with D5 combined-subject guard; 9/9 AC PASS (AC-DCA-001..008 blocking + AC-DCA-001b directional). drift count 67→51 (16 sub-class-1 false-positives resolved); 4 named exemplars aligned (SPEC-V3R5-GIT-STRATEGY-SCHEMA-001, SPEC-V3R6-CI-FLAKY-STABILIZE-002, SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001, SPEC-V3R6-CI-FLAKY-STABILIZE-001); legacy close conventions + genuine-incomplete-close carved out of scope (SPEC-V3R6-DRIFT-LEGACY-CONVENTION-001 follow-up). Coverage 85.9%, golangci-lint 0. Pre-existing `moai spec audit` (already reporting these as clean) unmodified. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations.

- **[SPEC-V3R6-CI-FLAKY-STABILIZE-002](.moai/specs/SPEC-V3R6-CI-FLAKY-STABILIZE-002/spec.md)** — CI Flaky Test Stabilization — supervisor Watch timeout + quality-gate dotnet timeout (Tier S, v0.1.0 → v0.2.0, run-phase M1-M3 2026-06-01, manager-docs sync-phase 2026-06-01). **2 residual flaky tests stabilized (TEST-ARTIFACT, zero production-code change)**. (FLAKY-1) `internal/lsp/subprocess/supervisor_test.go` `TestSupervisor_NonZeroExit` / `TestSupervisor_NormalExit`: comma-ok receive `ev, ok := <-ch` distinguishes delivered `ExitEvent` (`ok == true`) from context-deadline empty-close (`ok == false`), enabling skip-branch on load-induced timeout (defer to `t.Skip`) instead of spurious FAIL. Empirically reproduced 7/40 under 16× CPU burners + `go build ./...` concurrency; 5s deadline unrelated to test's routing verification intent. Root cause: bare receive `ev := <-ch` returned Go zero-value `ExitEvent{ExitCode:0}` on closed-empty, failing "ExitCode = 0, want 1" under contention even though `supervisor.go` `Watch()` contract documents empty-close when ctx cancelled. Correctness: production caller (`client.go` lines 359-365) ignores the `ExitEvent` value entirely, so timeout tolerance is harmless to production. (FLAKY-2) `internal/hook/quality/gate_test.go` `TestQualityGate_RunsDotnetFormatWhenCSharpStaged`: timeout argument raised from arbitrary 5s (scaffolding, 5.06-5.19s failure duration under 32× `yes` + `GOMAXPROCS=1` + concurrent broad suite) to 60s matching production `DefaultGateConfig.LintTimeout` SLA (line 46, `gate.go`), eliminating artificial deadline pressure. Test verifies routing logic (`.cs` file staged → dotnet executed), not timeout behavior; 60s >> fake-binary cost even under saturation. Sibling `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` (immune, structurally returns early) unchanged per immunity control requirement. **AC verification**: AC-CFS2-001 `TestSupervisor_*` under `-count=30` stress → 0 FAIL lines (timeout instances skip, not fail); AC-CFS2-002 deterministic wrong-code path (`exit 0` stub injection) still FAILs "ExitCode = 0, want 1" (comma-ok guards only timeout, not delivered wrong code); AC-CFS2-003 `-race -count=20` exit 0; AC-CFS2-004 dotnet under `-count=40` CPU saturation → 0 FAIL; AC-CFS2-005 immune sibling untouched; AC-CFS2-006 `git diff --stat` shows ONLY 2 `*_test.go` files, 3 prod files (`supervisor.go`, `client.go`, `gate.go`) byte-identical; AC-CFS2-007 full-suite `go test ./...` or affected packages `-count=5` PASS. **Files modified**: 2 (`supervisor_test.go` comma-ok + skip, `gate_test.go` timeout 5s→60s + LintTimeout 5s→60s). **Production diff**: 0 (supervisor.go, client.go, gate.go unchanged). **Cross-compile**: darwin/amd64 ✓ / darwin/arm64 ✓ / linux/amd64 ✓ / windows/amd64 ✓. **Subagent boundary (C-HRA-008)**: 0 AskUserQuestion invocations in test code. **Scope**: 2 of 3 residual flaky tests deferred by SPEC-V3R6-CI-FLAKY-STABILIZE-001 §C Out-of-Scope (3rd remains out-of-scope for future SPEC-V3R6-CI-FLAKY-STABILIZE-003). **Coverage**: PASS-eligible ACs 1/2/3/4/6 blocking (all PASS) + AC 5/7 informational (both PASS). **Root cause**: both tests bind assertions to arbitrary wall-clock bounds unrelated to what they verify (exit-code correctness ≠ timeout management; routing filter ≠ subprocess perf), making contention-induced timeouts spurious. Fixes are test-only architectural corrections (receive pattern + SLA alignment), not behavior changes to production code.

- **[SPEC-V3R6-CI-FLAKY-STABILIZE-001](.moai/specs/SPEC-V3R6-CI-FLAKY-STABILIZE-001/spec.md)** — CI Flaky Test Stabilization (Tier M, v0.1.1 → v0.2.0, run-phase M1-M3 2026-05-31, manager-docs sync-phase 2026-05-31). **2-issue remediation**: (FLAKY-1) internal/spec git-add race (ubuntu/macos CI race-detector): `performAtomicClose` now passes relative paths via `filepath.Rel(baseDir, ...)` ensuring consistent git resolution regardless of `cmd.Dir` timing; `Close()` normalizes `baseDir` to absolute path via `filepath.Abs` eliminating relative "." default. Verified: `go test -race ./internal/spec/ -count=20` → 0 failures, 0 race findings (local baseline). (FLAKY-2) Windows merge-TUI hang (windows-latest CI 600s timeout): `merge.ConfirmMerge` adds `runtime.GOOS == "windows" && !isatty.IsTerminal(os.Stdin.Fd())` guard returning fail-open error BEFORE `tea.NewProgram().Run()`, preventing Windows non-TTY `ReadConsole` block; caller audit identifies 1 test path (`update_skip_sync_test.go`) reaching `ConfirmMerge` without `--yes` (now skipped on Windows via existing precedent idiom). Reuses vendored `mattn/go-isatty` (no new dependency). **AC verification**: AC-CFS-001..010 (local gates: race test 20x, atomicity non-parallel, filepath inspection, TTY guard, Windows build exit0) all PASS; AC-CFS-011/012 CI-deferred (Windows job timeout + full matrix green) per HONESTY GATE §D.3 — SPEC is `implemented`; `completed` transition deferred until Windows CI confirms (AC-CFS-011/012). **Cross-platform**: darwin `go test -race ./...` green locally (excluding out-of-scope docs-i18n/mirror-drift packages); GOOS=windows GOARCH=amd64 cross-compile exit 0. **Files modified**: 2 (closer.go filepath.Rel+Abs fix; confirm.go TTY guard + 2 test Windows skips). **No new dependencies** (go.mod unchanged). **Subagent boundary (C-HRA-008)**: 0 AskUserQuestion invocations in implementation code.

- **[SPEC-V3R6-DOCS-I18N-PARITY-001](.moai/specs/SPEC-V3R6-DOCS-I18N-PARITY-001/spec.md)** — docs-site 4-locale i18n parity baseline clearance (Tier M, v0.1.2 → v0.2.0, run-phase 2026-05-31, manager-docs sync-phase 2026-05-31). **Baseline 53 errors cleared to 0**: 4 frontmatter (C1) → 0 by adding YAML frontmatter to cost-optimization/prompt-caching.md (4 locales); 26 H1 heading (C2) → 0 by adding `#` H1 to 5 files × 4 locales (advanced/harness-profiles, core-concepts/harness-engineering, db/migration-patterns, getting-started/profile, workflow-commands/moai-design) + 2 draft-true stub files (multi-llm/cg-mode.md, multi-llm/model-policy.md) in en/ja/zh; 23 glossary (C3) → 0 by adding invariant terms (`Claude Code`, `MoAI-ADK`, `moai-adk`, `SPEC-First`) to en/ja/zh where ko source contains them. **Verification**: `DOCS_I18N_STRICT=0 bash scripts/docs-i18n-check.sh` → Errors: 0 (warn-mode, AC-DIP-006); `bash scripts/docs-i18n-check.sh` → exit 0 (strict-mode, AC-DIP-007). **File-path parity intact**: 78 .md per locale, 0 orphan/missing files (AC-DIP-008). **Regression guard (AC-DIP-009)**: 0 forbidden URLs introduced; canonical URL `adk.mo.ai.kr` present; Mermaid non-TD count unchanged ≤ baseline 4. **Files modified**: 39 across all 4 locales (13 × 4 for H1/frontmatter + 8 for glossary terms + 3 doc-only edits). **Blocker count**: 0. **Out-of-scope confirmed**: no Go code, no template source, no new .md files, no docs-site CI flaky work (sibling SPEC-V3R6-CI-FLAKY-STABILIZE-001 separate), no docs-site strict-mode flip (Phase 2 separate decision).

- **[SPEC-V3R6-MAIN-RED-REMEDIATION-001](.moai/specs/SPEC-V3R6-MAIN-RED-REMEDIATION-001/spec.md)** — 4-group main-RED remediation: internal/template test-correction + mirror-drift policy + internal-content-leak sanitization (Tier M, v0.1.1 → v0.2.0, run-phase M1-M4 2026-05-30). **4-milestone summary**: M1 G1 agents-layout 9 stale tests corrected (FLAT `moai/` canonical expectation, core/meta/expert retired subdirs removed, 7 retained agents confirmed); M2 G4 hook-count constant updated (20 → 21, `PreCommit` event addition per LIFECYCLE-SYNC-GATE-001 M3); M3 G3 internal-content leak 30 → 0 (29 prose-substituted + 1 pedagogical allowlist path corrected per CLAUDE.local.md §25); M4 G2 mirror-drift 8 files resolved (7 files byte-parity, 1 file manager-spec.md per-file policy: excluded from byte-parity allowlist, covered by leak test). **Verification**: `go test ./internal/template/...` 13 parent tests RED → GREEN; cross-platform build (linux/amd64, darwin/amd64, darwin/arm64, windows/amd64) exit 0; golangci-lint 0 issues; leak test 0 occurrence; mirror-drift PASS. **Outcome**: internal/template group green restored (13 parent test failures → 0); main CI remains red due to 3 pre-existing flaky failures in other packages (internal/spec `TestClose_FullClose_ProducesCommit` race / windows cli·harness / docs-i18n) that are outside this SPEC's scope and were not introduced by this work.

- **[SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001](.moai/specs/SPEC-V3R6-TEMPLATE-NEUTRALITY-AUDIT-001/spec.md)** — Template Neutrality Audit for 16-Language Distribution (Tier L, v0.1.2 → v0.2.0, plan-auditor iter-2 PASS 0.88 + run-phase M2-M5 complete, sync-phase M6 2026-05-30). **5 C-class sanitizations delivered**: (A) C1 macOS-bias absolute paths (4 files `/Users/` → `$HOME/` in worktree-integration.md / run/context-loading.md / moai-foundation-cc / moai-workflow-loop references, AC-TNA-001 PASS 0 files post-fix); (B) C2 bare-narrative V3R sigils narrowed to 7 files (v0.1.2 rescope: 341 broad→299 ID-embedded deferred to ISOLATION-001 + 7 bare-narrative preserved/generalized, AC-TNA-002 PASS actual=6≤allowlist=6); (C) C4 feedback_/memory.md refs (9 baseline→7 post-fix, 2 GENERALIZE + 7 PRESERVE per doctrinal anchor, AC-TNA-004 PASS); (D) C5 CLAUDE.local.md refs (3 files fixed, 0 post-fix, AC-TNA-005 PASS); (E) C6 PR #N refs (3 files → 0 post-fix, PR #958/#622/#785 generalized, AC-TNA-006 PASS). **Disjoint audit scope** (NEUTRALITY-unique): C1/C2/C4/C5/C6/C8 kept-class detection; C3 (dates) / C7 (commit hash) deferred to ISOLATION-001's internal_content_leak_test.go strict tier. **New artifacts**: `internal/template/template_neutrality_audit_test.go` (422L, 2-pass C2 bare-narrative detection regex + 8 kept-class subtests + C8 GOOS false-positive preserve test + `-run TestTemplateNeutralityAudit` isolated execution); `.github/workflows/template-neutrality-check.yaml` (59L, CI guard trigger on `internal/template/templates/**` path change, audit-script isolation verification REQ-TNA-009/010). **Run-phase M2-M5 summary**: M2 commit `1046c6a3c` (C1=0 AC-TNA-001), M3 commit `c7c7b4e32` plan-phase blocker resolution + manager-develop bare-narrative GENERALIZE `manager-develop-prompt-template.md` both sides mirror-parity (AC-TNA-002 PASS 6≤6), M4 commits M4a/M4b (C4=7≤7 AC-TNA-004 PASS + C5=0 AC-TNA-005 PASS), M5 commits M5a/M5b (C6=0 AC-TNA-006 PASS + audit script + CI gate + C8=3 preserve AC-TNA-011 PASS). **AC coverage**: 8/8 run-phase M2-M5 blocking ACs PASS + 2 deferred (AC-TNA-003/007 to ISOLATION strict tier) + 2 M6 chore (AC-TNA-012/013 sync phase). **Plan-auditor verdict**: iter-2 PASS 0.88 (Tier L 0.85 threshold +0.03). **Blocker resolution history**: M3 blocker "C2 ≤18 unachievable" resolved by manager-spec v0.1.2 narrow to bare-narrative (7 files, 6 PRESERVE + 1 GENERALIZE target). **Migration matrix finalize (M6)**: Post-fix counts verified (C1=0, C2=6, C4=7, C5=0, C6=0, C8=3 PRESERVE) + migration-matrix.md upgraded to v0.2.0 M6 Final + CLAUDE.local.md §2 File Synchronization + §25 Template Internal-Content Isolation cross-reference NEW (acceptable content range guide + CI guard note). **Metadata**: 138-file template scope audited (73 C2 bare-narrative narrowed, 299 ID-embedded deferred, 4 C1 paths, 9 C4 feedback refs, 3 C5 CLAUDE.local, 3 C6 PR refs); 6/8 active kept-classes AC-TNA-001/002/004/005/006/011 PASS; zero post-fix drift on PRESERVE allow-lists; disjoint audit script (no C3/C7 scan) preserves ISOLATION-001 ownership.

- **[SPEC-V3R6-LIFECYCLE-SYNC-GATE-001](.moai/specs/SPEC-V3R6-LIFECYCLE-SYNC-GATE-001/spec.md)** — Atomic 4-Phase SPEC Close + Era Classification Heuristic + Ownership Transition Audit + Pre-Commit Drift Detection (Tier L, v0.1.3 → v0.1.4, run-phase completion 2026-05-30, manager-docs sync-phase 2026-05-30). **5 deliverables**: (A) Era Classification Subsystem (`internal/spec/era.go` H-1..H-6 heuristic engine + 5-bucket classification + grandfather-clause policy + frontmatter `era:` field override), (B) Atomic SPEC Close (`internal/spec/closer.go` / `closer_test.go` + `.moai/logs/lifecycle-close.log` 7-field JSON emission), (C) CLI Subcommands (`moai spec close --{backfill-only,dry-run,force,base-dir,json}` + `moai spec audit --{json,filter-era,include-grandfathered,strict,base-dir}` + help-text + flags), (D) Pre-Commit Hook (`.claude/hooks/moai/handle-pre-commit-spec-status.sh` 140L bash + status mismatch detection + canonical 4-phase-close subject enforcement + template mirror), (E) Rule File & Lint Extension (.claude/rules/moai/workflow/lifecycle-sync-gate.md 387L NEW + spec-frontmatter-schema.md `era:` field amendment + `internal/spec/lint_ownership.go` AC-LSG-004 Authored-By-Agent trailer parsing + AC-LSG-012 lint.skip opt-out). **Run-phase M1-M6 completion**: M1 Go primitives ~1650L / 86.2% cov + M2 CLI subcommands 1206L delta / 92.3% spec_close / 91.9% spec_audit + M3 pre-commit hook 6661B bash + 10512B Go tests / 6 subtests PASS + M4 OwnershipTransitionRule extension +123L lint_ownership.go / +390L tests / AC-LSG-004 WHO-signal + AC-LSG-012 skip opt-out + M5 rule file 387L NEW / 5 h2 sections (era-classification-heuristic / grandfather-clause / frontmatter-era-semantics / status-transition-ownership / worked-example) + M6 no-op regression dogfood (M1/M2 remediation commit b1710fd92 fixing closer.go no-op predicate triple-AND gate + lifecycle-close.log write code absent; 5 SPECs tested × exit 0 no-op × AC-LSG-018/020/022 PASS). **AC coverage**: 22/22 ACs total (15 functional M1-M5 + 5 NFR + 1 dogfood + 1 backfill-only semantics); 14 M1 PASS + 6 M2 PASS + 4 M3 PASS + 2 M4 PASS + 2 M5 PASS + M6 no-op 5 SPECs ≥5 log entries PASS = **21 run-phase ACs PASS + 1 backfill-only CONDITIONAL** per AC-LSG-020 no-op result="success" semantics. **Cross-platform**: linux/amd64 ✓ / darwin/amd64 ✓ / darwin/arm64 ✓ / windows/amd64 ✓. **Subagent boundary (C-HRA-008)**: internal/hook/pre_commit_spec_status_test.go `SubagentBoundary_NoAskUserQuestionReferences` PASS + internal/cli/spec_{close,audit}.go grep-verified 0 AskUserQuestion invocations. **Lint baseline**: golangci-lint 0 NEW issues, `go vet` 0. **Plan-auditor verdict**: iter-3 PASS 0.91 skip-eligible (Tier L 0.80 threshold +0.11). **Blocker resolutions**: D1-D8 amendments documented in HISTORY (v0.1.0→v0.1.1→v0.1.2→v0.1.3). **Sync-phase deliverables (manager-docs)**: spec.md frontmatter (status: in-progress→implemented, version 0.1.3→0.1.4, updated: 2026-05-30), progress.md §E.2 sync_commit_sha + §E.3 run-phase completion + §E.4 §E.5 Mx audit-ready signals, CHANGELOG.md this entry, README.md no-update (feature list stable, docs-site deferred per WIP protection).

- **[SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001](.moai/specs/SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001/spec.md)** — workflow.yaml nested keys (completion/loop_prevention/memory/default_mode/execution_mode/team.*) Go struct 정합 (Tier M, v0.1.0 → v0.2.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). **Full nested struct wire-through**: WorkflowConfig 9 sub-structs (AutoClearConfig / CompletionConfig / LoopPreventionConfig / MemoryConfig / TeamConfig / RoleProfileEntry / TokenBudgetConfig / WorkflowWorktreeConfig + MarkersConfig child) + backward-compat Option (c) accessor methods (WorkflowAutoClearEnabled / WorkflowPlanTokens / WorkflowRunTokens / WorkflowSyncTokens / WorkflowTeamAutoSelection). **5 milestones (M1-M5)**:  M1 `internal/config/types.go` nested struct definition 9 sub-structs + FLAT field rename (AutoClear→AutoClearLegacy, AutoSelection→AutoSelectionLegacy) + deprecated comments; M2 `internal/config/defaults.go` NewDefaultWorkflowConfig 36 nested defaults populated + `workflow_accessors.go` NEW 5 accessors + `loader.go` loadWorkflowSection added to Load() chain; M3 `workflow_nested_test.go` NEW production fixture yaml.Unmarshal + Edge-WSE-002/003/004 edge-case coverage; M4 `internal/cli/team_spawn.go` LoadRoleProfiles migration from ad-hoc string parser to typed `cfg.Workflow.Team.RoleProfiles` + `internal/cli/workflow_lint.go` internal `workflowConfig` type removed uses config.WorkflowConfig; M5 `internal/config/audit_loader_completeness_test.go` "workflow" exception removed from acknowledgedUnloadedSections (full unification completion signal). **8/8 AC PASS**: AC-WSE-001 24 nested path reachability / AC-WSE-002 TeamConfig struct shape + no Patterns field / AC-WSE-003 20 yaml.Unmarshal assertions (implementer.Isolation="worktree" + AutoClear.Enabled=true + TokenBudget.Plan=30000 + 17 more) / AC-WSE-004 FLAT field rename + deprecation comments complete / AC-WSE-005 5 accessor methods all return nested values / AC-WSE-006 LoadRoleProfiles 7-key map + ad-hoc parser removed + typed loader integrated / AC-WSE-007 36 default assertions (AutoClear.Enabled=true, Team.DefaultModel="sonnet", Team.RoleProfiles 7-key + more) / AC-WSE-008 exception removed + TestAuditLoaderCompleteness PASS. **Coverage**: 85%+ internal/config per baseline. **Subagent boundary (C-HRA-008)**: 0 AskUserQuestion invocations in implementation code. **Cross-platform**: linux/amd64 ✓ / darwin/amd64 ✓ / darwin/arm64 ✓ / windows/amd64 ✓. **Plan-auditor verdict**: iter-2 PASS 0.87 (Tier M 0.80 threshold +0.07) post-amendment (D1 Team.DefaultModel oracle + D2 Worktree.AutoCreate template sync). **Files modified (6 + NEW 2)**: types.go / defaults.go / workflow_accessors.go NEW / loader.go / types_test.go / defaults_test.go / workflow_nested_test.go NEW / team_spawn.go / workflow_lint.go / audit_registry.go / audit_loader_completeness_test.go. **Sync-phase deliverables**: spec.md frontmatter (status: in-progress→implemented, version 0.1.0→0.2.0, updated: 2026-06-03), progress.md §E.2 + §E.5 Mx signals, CHANGELOG.md this entry.

### Added

- **[SPEC-AUTONOMY-RUN-GOAL-001](.moai/specs/SPEC-AUTONOMY-RUN-GOAL-001/spec.md)** — Run-phase autonomy wiring: Mode 6 (workflow) + `/goal ac_converge` with GATE-2 preservation (Tier M, v0.1.0, run-phase 2026-06-03, manager-docs sync-phase 2026-06-03). Deliverables: (D1) orchestration-mode-selection.md Mode 6 catalog appended to canonical 5-mode table (trivial/background/agent-team/parallel/sub-agent) + decision tree + capability-gate + anti-patterns (REQ-ARG-001/002/003); (D2) run.md NEW `## Run-phase Autonomy (/goal ac_converge)` self-contained section co-locating GATE-2 human-gate ordering reference + `/goal ac_converge` set with 4 HARD safety conditions (transcript-measurable predicates, max 20 turns, semantic-failure escape, non-substitution guarantee per REQ-ARG-004/005/006/007); (D3) internal/template/gate2_preservation_test.go NEW regression guard asserting GATE-2 marker precedes first `/goal` token (REQ-ARG-008a/008b, Go test RED-GREEN proof); template mirror byte-identical parity (REQ-ARG-011); No named-script Workflow API asserted (REQ-ARG-009). 13/13 AC PASS (AC-ARG-001..012): Mode 6 entry conditions strict (≥30 files AND mechanical AND genuinely parallel; coding-heavy tasks stay Mode 5 per Anthropic Finding A4), GATE-2 mandatory pre-condition for Mode 6 + `/goal` set, score-independent gate per CLAUDE.local.md §19.1 / REQ-ATR-015 doctrine. Cross-platform build (linux/amd64, darwin/amd64, darwin/arm64, windows/amd64) exit 0. Subagent boundary (C-HRA-008) verified 0 AskUserQuestion invocations. Template neutrality audit PASS (no internal SPEC IDs / REQ tokens / dates / commit SHAs). Integration note: run-phase executed in L1 worktree (parallel session race-absorbed disjoint scope); M1-M4 cherry-picked onto diverged main (linear); original worktree backfill commit dropped post-cherry-pick. AC-ARG-004 awk verification corrected (first-GATE-2 guard `if(!g)g=NR` on line 115 < first /goal token on line 117, semantic invariant confirmed by Go test + corrected awk). Sync-phase: spec.md frontmatter status in-progress→implemented + updated: 2026-06-03, progress.md §E.4 sync audit-ready signal added, CHANGELOG entry (this line), no docs-site update (internal doctrine/workflow, not product feature).

- **SPEC-V3R5-INIT-WIZARD-EXPANSION-001 — moai init wizard decision-point expansion** (Tier M, v3.0.0-rc2 target). Expands the moai init wizard to support Phase 1 decision-point candidates (5 candidates) with advanced gating, phase readiness detection, and expanded questions for project mode, harness profile, LSP enablement, quality gates, and design workflow. **Run-phase completion**: M1-M5 implementation (commits 7e36d3697 M1-M5 implementation + 54bef086d progress.md), producing `internal/cli/init.go` flag registration, `internal/cli/wizard/{advanced_gate.go, questions.go, types.go, wizard.go}` wizard pipeline, `internal/core/project/{initializer.go, initializer_expansion.go}` initializer expansion with YAML write paths, tests `internal/cli/init_test.go`, `internal/cli/wizard/expansion_test.go`, `internal/core/project/initializer_expansion_test.go`. **AC verification**: 9/10 ACs PASS (AC-IWE-001..009 blocking PASS + AC-IWE-010 backward-compatibility DEFERRED per acceptance.md §C.2 — condition "byte-identical diff after Quick mode init" requires actual before/after execution which is sync-phase deferred). **Coverage**: wizard package 85.6% new code, core/project 88.9% new code. **Scope**: 13 files total (7 production + 3 test + 3 acceptance artifacts). **Status**: in-progress → implemented transition pending M5 final commit + sync.

- **SPEC-V3R6-PROMPT-CACHE-001 — Prompt Caching Integration** (Tier M, v3.0.0-rc2 target). Enables cost-optimized API calls via Anthropic prompt caching with automatic cache control, telemetry, and user-facing diagnostics. **Run-phase completion**: 5 milestones M1–M5 (commits 0d5680fbc `M1-M2 cache_control injection + cache.yaml` + 6a9cd5172 `M3-M4 PostToolUse telemetry + moai doctor metric` + final sync-phase M5 docs-site); **Key deliverables**: (a) `internal/runtime/cache_control.go` `InjectCacheControl` (system prompt TTL 1h + SPEC body TTL 5m, configurable via cache.yaml, min_cacheable_tokens 2048); (b) `internal/config/cache_config.go` `LoadCacheConfig` (safe-default enabled=false, auto-enable on valid config); (c) `.moai/config/sections/cache.yaml` (session_ttl / spec_ttl / min_cacheable_tokens configuration); (d) `internal/hook/posttooluse_cache.go` + `internal/state/cache_usage_log.go` (JSONL telemetry to .moai/state/cache-usage.jsonl per-turn, single-turn penalty WARN); (e) `internal/cli/doctor_cache.go` (cache hit rate [last 7 days], single-turn ratio warning); (f) `docs-site/content/{en,ko,ja,zh}/cost-optimization/prompt-caching.md` (4-locale break-even rule + mechanism explanation + opt-out guide + moai doctor interpretation). **AC completion**: 10 ACs (9 Blocking PASS + AC-PC-009 Should-fix deferred — 4-locale parity ratio 2.74 > 1.20 threshold; content-present but not yet balanced). **Break-even rule**: 1h cache enabled only when session has 2+ consecutive API requests; single-request sessions incur 2x write premium. **Implementation**: per SPEC plan.md M1–M4, sequenced as cache_control injection → config system → telemetry → moai doctor UI → docs-site user-facing. Session-level cache (1h TTL) + SPEC body cache (5m TTL) independent toggles. **Anthropic docs verification**: WebFetch confirmed prompt caching mechanism (cache-write 2.0x cost, cache-read 0.1x, TTL windows, minimum thresholds, pre-warming strategy) at https://platform.claude.com/docs/en/docs/build-with-claude/prompt-caching. **REQ traceability**: 10 GEARS-format REQs (Ubiquitous 5 + When 3 + Where 2) with 100% bidirectional AC↔REQ mapping. **Files modified**: 5 new (runtime, config, hook, cli, state) + 4 new docs-site locale files. **Status**: in-progress → implemented transition pending M5 final commit.
- **SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 — v2-to-v3 Clean Reinstall paradigm shift + v.2.x FLAT layout restoration** (Tier M, v3.0.0-rc2 target). Replaces the file-level enumeration `moai update` flow with a version-aware **backup → REMOVE → reinstall → MERGE-back → integrity verify** seven-step canonical order for projects detected as v2.x via the new `detectV2Fingerprint` heuristic (3 signals: `.moai/config/sections/system.yaml` `moai.version` field reads as `v2.*` / empty / file-missing → Signal 1; `.agency/` legacy directory presence → Signal 2; any of 43 enumerated `DeprecatedPaths` hits → Signal 3). Reverts the rc1-stage SPEC-V3R6-AGENT-FOLDER-SPLIT-001 `{core,expert,meta}/` subdirectory split (commit `1bd083725`, 2026-05-22) to the canonical v.2.x FLAT `.claude/agents/moai/<agent>.md` layout — verified by `git ls-tree -r 1bd083725^`. Supersedes SPEC-V3R6-AGENT-FOLDER-SPLIT-001. 22 AC-VVCR (17 AC-VVCR + 5 AC-VVCR-LR), 100% bidirectional REQ↔AC traceability (29 REQ-VVCR + 6 REQ-VVCR-LR = 35 REQs), 7 HARD constraints, 7 OOS exclusions. Plan-auditor iter-3 PASS skip-eligible 0.91 (Tier M threshold 0.80 +0.11; trajectory iter-1 0.83 → iter-2 0.89 → iter-3 0.91 monotone). Version bump: `pkg/version/version.go` `Version = "v3.0.0-rc1"` → `Version = "v3.0.0-rc2"` + `.moai/config/sections/system.yaml` `moai.version` / `template_version` synchronized. **Run-phase completion**: 7 milestones M1-M6 (commits 363eff563 `M1 version bump + CHANGELOG + 6 golden files` + 68e3af7b1 `M2 extend DeprecatedPaths 9→43` + e9eb74ae5 `M2a FLAT layout restoration 14 git mv` + 32c01f0eb `M3 v2 detection 552 LOC + 24 tests` + cc53ad421 `M4 clean reinstall orchestration 1215 LOC + 36 tests` + dec24f962 `M5 runUpdate integration + catalog regen` + 6c33a1bf4 `M6 cross-platform + FLAT layout audit`); push range `363eff563..a997f03a2` (9 commits total including L52 case 29 attribution restoration + checkpoint `de2205f2f`). Total LOC delta: ~2475 lines. AC completion: 17/22 PASS, 5 deferred (AC-VVCR-017 telemetry emission + 4 SHOULD-tier follow-up per design.md). L52 case 29 doctrine sustained: 2 occurrences absorbed via non-destructive chore commits (canonical Option A pattern). Cross-platform verified: darwin/amd64, darwin/arm64, linux/amd64, windows/amd64. Subagent boundary (C-HRA-008): grep returns 0 matches on all new code.

### Fixed
- **SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001: Template 내 moai-adk 내부 개발 토큰 영구 격리** (Tier M, Sprint 11 entry, 완료): moai-adk dev-internal-content 토큰 (SPEC ID / REQ token / Audit 인용 / 개발 날짜 / commit sha) 영구 제거 — Option A 44-file scope adopted at run-phase pre-flight (spec.md §A.4 narrow 35-file은 plan-phase approximation). **5개 deliverable**: (a) templates 35→4 narrow scope cleanup (잔존 4 files 모두 의도된 결과 = 1 `.gitignore` bash literal false positive + 1 `output-styles/moai/moai.md` scope-excluded per orchestrator directive [§8 Localization Contract 별도 강화 chore `cbc710721`로 분리] + 2 pedagogical allowlist files [`manager-spec.md` regex walkthrough lines 146/161 + `askuser-protocol.md` Socratic example lines 194/199/204]), (b) CLAUDE.local.md §25 NEW HARD rule + 5-item self-check checklist (96 lines body), (c) Go lint test `TestTemplateNoInternalContentLeak` 신규 (red-green proof + pedagogical allowlist 5 entries 메커니즘 + D-007 short-sha trailing-space + sentence-final punctuation extension), (d) CI workflow extension (docstring + memory feedback cross-reference), (e) maintainer-only audit 통과 (templates 내 97/98/99 dev-only commands + `settings.local.json` + `last-cc-version.json` 부재 확인, 0 forbidden file classes). **AC 결과**: 12 AC 매트릭스 = 8 PASS + 2 PASS-WITH-VARIANCE (AC-TII-001 narrow vs Go regex divergence + AC-TII-007 5 occurrences in scope-excluded output-styles) + 1 PASS-WITH-DEBT (output-styles scope-exclusion 정책) + 1 N/A (AC-TII-010 조건부, M6 audit 발견 0건이므로 trivially PASS). **Run-phase 구조**: 13 REQ-TII (100% GEARS Ubiquitous 4 + Event-driven 4 + State-driven 1 + Where-capability 3 + Unwanted 1) + 12 AC-TII bidirectional traceability + 5 substitution pattern 일관성 (predecessor pass 1/2 호환) + pedagogical allowlist 메커니즘 정착. **L52 case 29 2회차 mitigation 정착**: M1 content + attribution hijack (`d9838995d` race-absorbed → `c5ed59907` 비파괴 attribution chore 복원) + M4.2-redo content hijack 재현 (`69075e8cb` 12 files captured → `100e603d3` 비파괴 attribution chore retrospective 복원). 비파괴 chore 패턴이 multi-session race mitigation canonical doctrine으로 정착. **L67 NEW variant 완전 mitigation**: 1차 spawn M4.1/M4.2 commit-message claim mismatch 후 B15 self-verify discipline 도입 → 2차 spawn 8 commits 모두 commit subject N == 실제 file count 일치. **검증**: bash narrow grep 35→4 (의도된 4 files 모두 분류 가능), Go test narrow `TestTemplateNoInternalContentLeak` 5 occurrences (모두 output-styles scope-excluded), self-check 5-item ✓, lint test red+green proof 적용, CI gate extension 명시. **선례 패턴**: 사용자 명시 정책 (feedback_template_internal_content_isolation.md 2026-05-25) → 본 SPEC으로 HARD rule 정착 → CLAUDE.local.md §25 신규 추가 + lint test enforcement → 향후 SPECs 정책 명시로 팀 동기화 (§21 dev-only commands, §24 harness namespace 계열).

### Fixed
- **SPEC-V3R6-TEST-REFACTOR-001: Go test suite refactor — ATR-001 PROCEED-WITH-DEBT discharge** (Tier M, Sprint 10 cohort 8/8): Discharged 8 architectural-pivot + 7 cascade test failures deferred by SPEC-V3R6-AGENT-TEAM-REBUILD-001 M8 (reported 8 failures; measured 15 total baseline). **Run-phase completion**: 6 milestones M1–M6 (commits 4c0bb8424 `ground-truth` + 5a4fdf96d `internal/template` 11 fixes + d68421012 `internal/skills` 2 fixes + 9f58ed63b `internal/harness` subagent-boundary + bdc707bde `internal/statusline` pre-existing + 6ed1155ea `verification-batch`); **Outcome**: `go test ./...` exit 0 (zero FAIL lines); 14/14 AC-TST-001..014 PASS; 0 PASS-WITH-DEBT debt incurred; **Ground truth verification**: initial measurement 15 failures (spec.md §A.4 enumerated), post-run measurement 0 failures (DDD cycle ANALYZE-PRESERVE-IMPROVE per plan.md §A), zero drift from baseline; **Catalogs & Generated Artifacts**: embedded.go / catalog.yaml / catalog.go regenerated via `make build` (no hand-edits per HARD-3); **Path-specific staging** (L46): each milestone commit staged only files within its milestone scope (M1: 4 SPEC frontmatter; M2: 7 test + 1 mirror sync; M3-M5: 1 each; M6: progress.md only); **L44 verification** (38x): all 6 milestones M1-M6 pre/post fetch returned `0 0` (clean state); M3 absorbed L52 case 29 (b7d1528c8 SPEC-V3R6-TEMPLATE-INTERNAL-ISOLATION-001 plan-phase, scope-disjoint clean FF); **Coverage**: internal/template 84.6%, internal/harness 87.9%, internal/statusline 84.9%, internal/harness/* 86–100%; **Lint**: golangci-lint returned `0 issues` (zero NEW violations); **REQ traceability**: 13 GEARS-format REQs (Ubiquitous 4 + When 3 + While 2 + Where 2 + Unwanted 2) with 100% bidirectional AC↔REQ mapping (no IF/THEN deprecated modality); **Predecessor SSOT preservation** (L48): spec.md / plan.md / acceptance.md / progress.md bodies unchanged; `git diff origin/main -- .moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/ .moai/specs/SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001/ .moai/specs/SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001/` verified empty; **Blocker compliance** (D1–D4): no frontmatter:implemented transition in run-phase (manager-docs owned); no CHANGELOG body edit in run-phase; no scope creep beyond 15-failure enumeration; ATR-001 archive enumeration inline (plan.md); zero deferred debt post-run. **Context**: Tier M test refactor (300–1000 LOC, 11 total files modified, 6-milestone Hybrid Trunk main-direct default per CLAUDE.local.md §23.7–23.9) to discharge 8 failures reported at ATR-001 M8 + 7 additional failures discovered in +7 drift post-ATR-001 sync/Mx cohort completion. Scope: 1 pre-existing (TestRenderPRSegment_Absence, TestRetirementCompletenessAssertion path drift predating ATR-001); 9 architectural-pivot consequences (17→8 agent catalog; retained-agent contract real-ization; hook boundary expansion; embedded catalog refresh); 3 cascade (post-ATR-001 template/rule drift); 2 needs-investigation deferred (TestTemplateMirrorParity && TestSubSkillLOCCeiling, both resolved inline per HARD-1 test-first prefer). [Run-phase commits 4c0bb8424..6ed1155ea = 6 total; pre-fetch 0 0 all; post-push 0 0 final]

### Changed
- **`internal/template/catalog.yaml`: Catalog cleanup — 12 archived agent entries removed, 8 retained agent hashes refreshed, `generated_at` updated to 2026-05-25T08:06:18Z. Entry count 50 → 38. Discharges ATR-001 M8 misreported catalog cleanup debt (commit message claimed `grep = 0` but entries remained until this SPEC). [SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001]**

### Added
- **`internal/spec/catalog_hash_test.go`: Drift lock-in regression guard. `TestCatalogHashParity` asserts every catalog entry's stored sha256 matches its normalized source body hash; `TestCatalogHashParity_MissingTemplates` verifies REQ-CHR-007 fail-loud invariant. Prevents future ORPHAN/HASH-DRIFT regressions from landing silently on main. [SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001]**

### Changed
- **Agent catalog Anthropic 2026 alignment: 17→8 consolidation (SPEC-V3R6-AGENT-TEAM-REBUILD-001)** — Tier L deep audit of MoAI agent catalog revealed structural divergence from Anthropic 2026 published guidance ("Start with 3-5 teammates for most workflows"; "Subagents cannot spawn other subagents"). Audit 3 findings (Anthropic verbatim citations) triggered architectural pivot: consolidate 17 custom agents → 8 retained (7 MoAI-custom + 1 Anthropic built-in `Explore`); archive 12 phantom agents (0 invocations across recent 4-SPEC cohort) to `.moai/backups/agent-archive-2026-05-25/` with migration guide at `.claude/rules/moai/workflow/archived-agent-rejection.md`. 22 AC-ATR-XXX with 100% REQ↔AC traceability (20 REQ-ATR, 100% GEARS notation self-dogfood). Run-phase M1-M8 (8 commits `955299cac..f91acb3a3`): M1 7 retained agent frontmatter refinement + M2 3 workflow router phase-owner declarations + M3 11 archived agents (local + template parity) + M4 3 NEW hook scripts (PostToolUse status-transition + Stop sync-quality-gate + TaskCompleted team-ac-verify, dormant for now) + M5 2 NEW rule files (orchestration-mode-selection.md 5-mode YAML decision tree + archived-agent-rejection.md migration table) + 8 modified rule files + M6 predecessor SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 supersedence verify + AC-ATR-012 reinforcement (manager-develop cycle_type=autofix body 5-section expansion) + M7 CLAUDE.md §4 Agent Catalog rewritten (8-agent table + archive cross-reference) + CLAUDE.local.md §19 GATE-2 + §23 Tier-based PR Routing + `.claude/rules/moai/NOTICE.md` Anthropic 2026 Alignment attribution (6 source URLs) + M8 template parity + catalog.yaml 12-archived purge + 7-item verification batch (Test debt 8 failures PROCEED-WITH-DEBT per user directive — 1 pre-existing path drift + 7 architectural-pivot consequences deferred to follow-up SPEC-V3R6-TEST-REFACTOR-001). **L52 case 22 NEW**: SPEC-V3R6-CATALOG-HASH-REGRESSION-CLEANUP-001 plan-phase `128947eb6` parallel session race-absorbed (scope-disjoint). **Predecessors superseded**: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001 (status: superseded as of plan-phase commit `b957a4d04`; Audit 3 findings invalidated the original 17-phase-fix scope). **Successor queue**: Sprint 10 cohort + TEST-REFACTOR-001 follow-up for test suite modernization.

### Changed

- **SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001** (Tier M, Sprint 10 lane B 1/4): Local agent namespace consolidation — dev-only skill migration (release-update-specialist + github-specialist agents, `.claude/agents/local/` namespace NEW) + template generic refactor (15 CLAUDE.local.md reference removals across 12 template + 11 local files) + generic-patterns-guide authoring (externalized W5). W3-arch scope: 2 agent bodies migrated (~1184 LOC) + 2 thin command wrappers (97/98) updated + 2 dev-only skills deleted. W4 scope: 13 template files leak-fixed + 11 local mirrors (lsp.yaml.tmpl template-only, agent-authoring.md M1-fixed). W5 scope: `.moai/docs/generic-patterns-guide.md` authored (local + template mirror). Plan-auditor 3-iter trajectory: iter-1 FAIL 0.73 → iter-2 FAIL 0.74 (stagnation, D1+D2+D3 partial progress) → iter-3 PASS-WITH-DEBT 0.83 (max-3 contract reached, 6/6 RESOLVED, 2 INFO deferred). 13 REQ-LNC (Ubiquitous + Event-driven `When` + State-driven `While` + Where capability-gate) **100% GEARS notation self-dogfood** + 12 AC-LNC (11 MUST-PASS AC-LNC-001..011 PASS + 1 SOFT AC-LNC-012 deferred). Run-phase 6 commits: M1 `03a568508` (namespace contract docs) + M2 `d4beaa50f` (agent body authoring) + M3 `d9cce5427` (race-absorbed AGENT-TEAM-REBUILD-001 M2, dev-only skill removal + 97/98 rewiring + 2 test updates) + M4 `979bec4eb` (template generic refactor, 23 files actual vs 26 plan due to lsp.yaml.tmpl template-only + agent-authoring.md M1-fixed) + M5 `55b55207f` (generic-patterns-guide.md authoring) + M6 `72a733765` (orchestrator M6 chore: progress.md backfill + agent-common-protocol self-mirror sync drift fix). **L52 case 18 NEW**: AGENT-TEAM-REBUILD-001 M2 `d9cce5427` parallel session race-absorbed (scope-disjoint: workflow router consolidation ↔ dev-only skill migration), overlapping code edits (97/98 thin wrappers, 2 skill deletions, 2 test updates), net effect M3 deliverables present on main attributed to parallel commit, L52 race-absorbed pattern + L46 path-specific discipline 0 cross-attribution leakage. **45 files modified** (M1 9 + M2 2 + M3 6 race-absorbed + M4 23 + M5 3 + M6 2). Predecessor cohort context: Sprint 10 lane B GEARS sweep foundation (FOUNDATION-CORE-GEARS-ALIGN-001 0.87 PASS, PLAN-AUDITOR-GEARS-ALIGN-001 0.913 skip-eligible, SKILL-GEARS-ALIGN-001 0.892 not-skip-eligible) complete; local-namespace follows as Lane B entry 1 consolidating 3 overlapping scopes under single SPEC + single CHANGELOG entry.

- **SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001** (Tier M, Sprint 10 cohort 4/8): `/moai plan` workflow skill bundle GEARS notation alignment. GEARS-first canonical form across 4 local + 4 mirror = 8 files (`.claude/skills/moai/workflows/plan.md` + 3 sub-skills + `internal/template/templates/.claude/skills/moai/workflows/plan.md` + 3 sub-skill mirrors). EARS retained as 6-month backward-compat legacy reference (window expires 2026-11-22 per SPEC-V3R6-GEARS-MIGRATION-001). 13 REQ-WPG-XXX (100% GEARS notation self-dogfood) + 11 AC-WPG-XXX (11/11 PASS). Mirror parity byte-for-byte verified, 0 IF/THEN deprecated modality introduced, LegacyEARSKeyword lint baseline preserved (7 = 7 baseline, zero regression). Plan-auditor 3-iter trajectory iter-1 PASS 0.867 → iter-2 PASS-WITH-DEBT 0.873 → iter-3 PASS-WITH-DEBT 0.870 (max-3 retry contract reached, PROCEED-WITH-DEBT). All 4 residual debt items D_new4/5/6/7 resolved inline during run-phase per plan-auditor PROCEED directive. Run-phase 3 commits (`426adbb64` + `886eb39d6` + `d834f4ac5`).

- **moai-foundation-core SKILL bundle GEARS 우선 정렬** (SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 v0.1.0, run-phase commits `2f1786281` + `31a2e1783` + `a85f7699c` + `dd8a08a59`): 10 local files + 10 template mirrors (SKILL.md + modules cluster A-B + references + catalog.yaml regen) synchronized to GEARS notation primary, EARS labeled legacy reference per 6-month backward-compat window. **Template parity** verified 100% (`diff -r` empty, AC-FCG-005 PASS). **AC matrix**: 9/9 mandatory ACs PASS (8 fully PASS + 1 PASS-WITH-NOTE for AC-FCG-006 embedded.go non-existence; catalog.yaml regen equivalent verified). **GEARS self-dogfood**: all 12 REQ-FCG-XXX in spec.md written in GEARS notation (100% self-test per REQ-FCG-009); zero IF/THEN outside spec-ears-format.md (V3 grep enforcement, AC-FCG-004 PASS); spec-ears-format.md DEPRECATED banner preserved verbatim (AC-FCG-007 PASS). **Predecessor pattern**:  Tier M (5-15 files, ~500 LOC, 0.87 plan-auditor PASS) → 1-pass run-phase success per SKILL-GEARS-ALIGN-001 + PLAN-AUDITOR-GEARS-ALIGN-001 precedent. **Sprint 10 cohort**: entry SPEC #3 of 8 GEARS sweep; foundation-tier alignment unblocks 5 downstream cohort SPECs (WORKFLOW-PLAN, DOCS-SITE-FULL, WORKFLOW-SPEC-EXTRAS, MISC-DOCS, RULES-GO-DOCS, completion target 2026-11-22 before 6-month backward-compat window expiry).

- **plan-auditor GEARS-aware rubric 정렬** (SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001 v0.2.0, run-phase commit `6366d7428`): M1-M3 GEARS-aware rubric body + template mirror parity (2 files: `.claude/agents/meta/plan-auditor.md` + `internal/template/templates/.claude/agents/meta/plan-auditor.md`). **AC verification** (8/8 mandatory AC PASS): AC-PAGA-001 `grep -c "GEARS"` count ≥5 / AC-PAGA-002 combined EARS/GEARS label present / AC-PAGA-003 `shall not` negative form / AC-PAGA-004 IF/THEN deprecation marker in failure-modes / AC-PAGA-005 `#gears-notation` cross-reference link / AC-PAGA-006 template mirror byte-identical (`diff -q` empty) / AC-PAGA-007 self-lint 0 LegacyEARSKeyword + 0 FrontmatterInvalid / AC-PAGA-008 PRESERVE working tree + 4 additions. **Highlights**: (1) MP-2 must-pass label "EARS Format Compliance" → "EARS/GEARS Format Compliance" reflecting unified approach; (2) M3 Score 1.0 anchor expanded to GEARS Five Patterns canonical reference (Ubiquitous / `When` Event-driven / `While` State-driven / `Where` capability-gate reframing former "Optional" / Unwanted as `<subject> shall not <action>` with legacy `If [undesired condition], then [action]` retained & marked `[DEPRECATED — use shall not]`); (3) M2 Adversarial Stance failure-modes list extended with "ACs use IF/THEN syntax without deprecation marker (post-6-month backward-compatibility window)" candidate; (4) generalized `<subject>` substitution (any noun, not just "the system") acknowledged per GEARS policy (88 existing EARS SPECs retain default "The system" for readability); (5) 4-locale docs-site GEARS migration guide (`moai-plan.md#gears-notation`) cross-reference added. **Self-dogfood**: this SPEC's own REQs use GEARS notation (REQ-PAGA-001..009) as self-test exemplar; lint engine confirms 0 LegacyEARSKeyword + 0 FrontmatterInvalid (AC-PAGA-007). **Tier S minimal**: 2 artifacts (spec.md + plan.md inline AC per Tier S LEAN pattern), plan-auditor iter-1 PASS 0.913 skip-eligible ≥0.90 threshold, 9 GEARS-notation REQs, 8 binary AC. **Predecessor SPECs**: SPEC-V3R6-GEARS-MIGRATION-001 (lint engine canonical `LegacyEARSKeyword` rule + 4-locale docs-site baseline) + SPEC-V3R6-SKILL-GEARS-ALIGN-001 (5 authoring guide files + 5 template mirrors GEARS-first) both `status: implemented` (PR #1046 + v0.2.0 merged 2026-05-25). **Backward-compat window**: 6-month grace period active (cohort closure TBD, expiry ~2026-11-22) — plan-auditor GEARS-aware rubric judges both GEARS NEWs and legacy EARS SPECs as PASS-equivalent at Score 1.0 until window closes.

- **Anthropic Best-Practice Audit Tier 3 — Subdirectory CLAUDE.md (F9) + Programmatic DRI Ownership Verification (F13)** (SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 v0.2.0, commits `9d77f890b` + `9d76d72be` + `91adaa53f`): 5개 subdirectory CLAUDE.md (`.claude/`, `internal/{cli,config,hook,spec,template}/CLAUDE.md` 신규 작성, F9 전체) + 2개 lint 파일 extension (F13: `internal/spec/lint_ownership.go` NEW 356L + `internal/spec/lint_ownership_test.go` NEW 444L, `OwnershipTransitionRule` 구현) + 1개 schema document extension (`.claude/rules/moai/development/spec-frontmatter-schema.md` +61L Status Transition Ownership Matrix + Cross-Reference section) + 1개 template mirror byte-identical (internal/template/templates/.claude/rules/moai/development/spec-frontmatter-schema.md) + 1개 progress.md lifecycle tracking = **10 files / ~700-1000 LOC / Tier M standard**. **AC verification** (10 mandatory + 1 AC-FW-004 normative): M1 commit `9d77f890b` (5×CLAUDE.md 신규 32-34 LOC 4-section format) + M2 commit `9d76d72be` (OwnershipTransitionRule 7 canonical transition PASS + 5 violation FAIL + graceful no-op + 4-platform cross-build + coverage 85.3%) + M3 commit `91adaa53f` (schema document + template mirror byte-identical) = 10/10 mandatory ACs PASS + AC-FW-004 content-density 격상 normative PASS. **Cross-platform verification** (4 platforms): linux/amd64 ✓ / darwin/amd64 ✓ / darwin/arm64 ✓ / windows/amd64 ✓. **Quality gate** (zero blockers): golangci-lint 0 NEW issues / go vet 0 / C-HRA-008 subagent boundary PASS (0 AskUserQuestion violations in `internal/spec/`). **Governance impact** (forward-looking): 39 OwnershipTransitionInvalid Warnings emitted on existing SPECs (observation-only severity per REQ-AAT-009 default subset, not blocker — intended governance signal for future SPEC sync-phase audits). **Lessons emitted** (3 NEW proposals): L60 gitignore pre-flight (git check-ignore -v), L61 stale .git/index.lock recovery (user manual rm exception), L62 paste-ready memo verification (git diff HEAD verification before trusting subagent claims). **Race-absorbed** (L52 case 12 NEW): SKILL-GEARS-ALIGN-001 parallel session `1f3a734d8` + `353150294` disjoint scope fast-forward between M1-M2 push, 0 cross-attribution leakage via pre-spawn fetch L4 + path-specific add L46 + pre-commit staging assertion L59 disciplines. **Blocker resolution** (5th blocker NEW): paste-ready memo claim verification revealed M2 prior spawn failure (0 commits, 4-blocker storm) — narrow re-spawn manager-develop M2-only second attempt PASS 1-commit+800LOC (manager-develop-prompt-template.md B1-B12 Known Issues discipline applied post-incident). **Plan-auditor verdict** (2-iter trajectory): iter-1 0.7350 FAIL + D1 H3 sub-heading + D2 3 orphan REQ annotation + D3 title fix → iter-2 PASS 0.8675 (Tier M 0.80 threshold +0.0675 margin). **Tier classification**: M (5 subdirectory CLAUDE.md + 2 lint files + schema doc + template mirror + progress.md = 10 files, Tier M 5-15 file envelope, ~700 LOC within ~300-1000 LOC guidance).

- **moai-workflow-spec SKILL GEARS 우선 정렬** (SPEC-V3R6-SKILL-GEARS-ALIGN-001 v0.2.0, commits `1f3a734d8` + `353150294`): 5 authoring guide files (`.claude/skills/moai-workflow-spec/SKILL.md`, `.claude/skills/moai-workflow-spec/references/reference.md`, `.claude/skills/moai-workflow-spec/references/examples.md`, `.claude/skills/moai-foundation-core/modules/spec-ears-format.md`, `.claude/agents/core/manager-spec.md`) + 5 template mirrors synchronized with GEARS-first guidance. GEARS Five Patterns canonical reference (Ubiquitous / When / While / Where / deprecated If/Then) + unified compound clause + generalized `<subject>` (replaces hardcoded "the system") + If/Then deprecation callout with LegacyEARSKeyword lint behavior + 6-month backward-compat window policy. 13/13 ACs PASS (plan-auditor iter-1 0.892 not skip-eligible; manager-develop 1-pass confirmed all AC). Tier M (markdown-only, 5 files + 5 templates) with self-dogfooding: this SPEC's own REQs written in GEARS notation as canonical examples per REQ-SGA-001..012.

- **SPEC frontmatter drift 추가 정정 — 2개 누락 implemented SPEC 백필** (2026-05-25, chore commit 후속): 어제 batch cleanup `3b7ce18b6` (22 archived + 4 implemented drift 정정)에서 누락된 2개 SPEC을 추가 audit + 백필. (1) **SPEC-V3R5-ATOMIC-WRITE-001** (P0-4 v2.20.0-rc1 release blocker, sync commit `d49a0a7db` 2026-05-21): frontmatter `version: "0.1.0" / status: draft / updated: 2026-05-20` → `version: "0.2.0" / status: implemented / updated: 2026-05-21` 복원 + HISTORY v0.2.0 row 복원 + `progress.md` 신규 작성 (sync 시점에 미포함 — Tier S minimal sync scope 2 files: CHANGELOG.md + spec.md). 실 구현 잔여 검증: `internal/migrate/hook_cleanup.go:124-164` atomicWrite (CreateTemp+Write+Sync+Close+Rename+Chmod + @MX:WARN/@MX:REASON) + `internal/migrate/hook_cleanup_test.go` 613L 회귀 테스트 6종 (TestAtomicWrite_* 5종 + TestCleanupUserSettings 1종) 모두 main에 정상 존재 / 7/8 ACs PASS + 1 CONDITIONAL (AC-AWR-007 coverage EXCL-AWR-003 Tier L deferred). (2) **SPEC-V3R6-HOOK-CONTRACT-FIX-001** (Wave 0 Critical, chore commit `19146bca6` 2026-05-22): frontmatter 동일 패턴 복원 (`v0.2.0 / implemented / 2026-05-22`) + HISTORY v0.2.0 row 복원 + `progress.md` 202L verbatim 복원 (`git show 19146bca6:.moai/specs/.../progress.md` 그대로). 실 구현 잔여 검증: `.claude/settings.json` + `.tmpl` WorktreeCreate/Remove key 부재 (REQ-HCF-003 PASS) / `internal/hook/subagent_stop.go:205 resolveProjectRoot to prefer CLAUDE_PROJECT_DIR` (REQ-HCF-005 PASS) / `internal/hook/.moai/` 부재 (REQ-HCF-006 PASS) / `internal/hook/path_resolve_test.go` TestResolveProjectRoot* 5종 회귀 테스트 정상 / 14/14 in-scope ACs PASS + 2 baseline residual FAIL preserved. 추정 drift 원인: `720a636b5 chore(main-sync): 23 parallel-session + lifecycle cleanup commits (2026-05-22) (#1037)` mass-merge conflict resolution이 spec.md pre-sync 버전을 silent하게 유지 + HOOK-CONTRACT progress.md를 누락. L48 SSOT canary 준수 (양 SPEC body 절대 미수정, frontmatter + HISTORY 표만, progress.md는 신규 작성 또는 verbatim 복원이라 SSOT 외 산출물). 본 정리 자체에 대한 SPEC 작성 없이 chore commit으로 처리 (Hybrid Trunk Tier S admin cleanup, `3b7ce18b6` 패턴 그대로).
- **SPEC ledger 정리 — 26개 stale SPECs 일괄 상태 정정** (2026-05-25, chore commit 후속): 미구현 SPEC 39개를 audit한 결과 26개의 stale entry를 발견하여 일괄 처리. **22개 archived** (status: * → archived): (1) 3개 TEST fixture (AC test 용도) `SPEC-TEST-{AUTH,ORC-LIKE,OVERRIDE}-001`, (2) 1개 old design-only `SPEC-V3R2-EXT-003`, (3) 7개 Wave 3 Tier 2 plan-only (#749 commit으로만 존재, run-phase 진입 없음) `SPEC-{CONTEXT-INJ,MEM-SCOPE,PARALLEL-COOK,SKILL-DESC,STOP-HOOK,TOOL-AUDIT,WT-DOC}-001`, (4) 5개 Wave 4 Tier 3 plan-only (#750) `SPEC-{CACHE-ORDER,CRON-PATTERN,METRICS,NO-HYBRID,RESUME-MSG}-001`, (5) 3개 v2.15 UTIL backlog (#707, ast-grep 미도입) `SPEC-UTIL-{004,005,006}`, (6) 3개 abandoned `SPEC-{LSPMCP,EVO,ASTG-UPGRADE}-001` (stub만 또는 인프라 미구현). **4개 implemented 정정** (status: in-progress → implemented, frontmatter drift): `SPEC-MX-001` (@MX TAG 1465 occurrences in `internal/`), `SPEC-LSP-CORE-002` (`internal/lsp/` 12 sub-dirs), `SPEC-CI-MULTI-LLM-001` (`.github/workflows/claude.yml`), `SPEC-TELEMETRY-001` (`internal/telemetry/` + `internal/cli/telemetry.go`). 결과: 미구현 SPEC 39 → 13 (3분의 1로 축소, 모두 V3R5/V3R6 modern era draft). SPEC body 절대 미수정 (frontmatter status + updated 필드만, L48 SSOT canary). 본 정리 자체에 대한 SPEC 작성 없이 chore commit으로 처리 (Hybrid Trunk Tier S admin cleanup).
- **`spec lint` MissingExclusions 회귀 정리** (SPEC-V3R6-SPEC-LINT-CLEANUP-001, commit `d1558e092`): 8개 sibling SPEC `spec.md` 본문에 H3 sub-heading retroactive 적용. lint count 10 → 2 (8 in-scope 해소). 분류 A (H3 추가) 2건 + 분류 B (기존 H3 하위 list item 추가 / hyphen→space heading 정정) 6건. 잔여 2건은 `ANTHROPIC-AUDIT-TIER3-001` + `HARNESS-NAMESPACE-CLEANUP-001 §5.3` 병렬 세션 drift로 본 SPEC out-of-scope (acceptance.md §D.4 edge case PASS-WITH-NOTE). 0 semantic deletions (PRESERVE invariant). Tier S minimal 1-pass run-phase 성공 + Trust-but-verify 7-item batch 0 critical discrepancies + L52 case 5 real-time race absorbed (HARNESS-NAMESPACE-CLEANUP-001 parallel session 8 NEW items 절대 흡수 0).

## [v3.0.0-rc1] — 2026-05-22: Hooks Contract Cleanup + 9-PR Batch Sync + Hybrid Trunk Config

### Fixed

- **WorktreeCreate / WorktreeRemove hook 등록 해제** (commit `a3239d3de`): Claude Code v2.1.49+ 공식 컨트랙트 (https://code.claude.com/docs/en/hooks)를 직접 확인 후 hook이 **active creator** (Claude Code default git worktree behavior를 **replace**)임을 발견. 우리 observer 의도 등록이 빈 `HookOutput{}` JSON `{}` 출력 → Claude Code가 `{}`를 path로 해석 → `"WorktreeCreate hook returned a path that is not a directory: {}"` 회귀 유발하여 5 agent (manager-develop/expert-frontend/expert-backend/expert-refactoring/researcher) isolation 호출 전부 실패. 정정 11 files +51/-86: `.claude/settings.json` + `.tmpl` WorktreeCreate/WorktreeRemove key entry 제거 (Claude Code default 사용) / `hooks-system.md` (local + template) WorktreeCreate row에 active creator 컨트랙트 + MoAI default 비등록 주석 / `worktree-integration.md` (local + template) §WorktreeCreate and WorktreeRemove Hooks 전면 재작성 (stdin 필드 / stdout plain text path only / exit semantics / 향후 active creator opt-in 가이드) / `docs-site/content/{ko,en,ja,zh}/advanced/hooks-reference.md` 4-locale handler table 정정 (한국어 baseline drift 동시 해결) / `internal/template/settings_test.go` TestSettingsTemplateHookEventCount 22 → 20. Handler 코드 (`internal/hook/worktree_{create,remove}.go`) + CLI subcommand + shell wrapper는 향후 active creator opt-in 인프라로 보존.

### Added

- **Claude Code 신규 hook event 카탈로그 sync** (commit `32fac92e7`, cherry-picked from closed PR #962): UserPromptExpansion (v2.1.90+) + PostToolBatch (v2.1.89+) + mcp_tool hook type 카탈로그 추가. hooks-system.md (local + template) + 4-locale hooks-reference.md (ko/en/ja/zh) 통합. PR #962는 본 commit 직전에 cherry-pick 후 close (5826331f merge commit 우회 + 00f2850c 실제 변경 cherry-pick + en hooks-reference.md baseline drift conflict 합집합 resolve).

### Changed

- **9 OPEN PR 일괄 정리** (admin merge batch): #1042 (v3 blueprint Wave 6) + #1041 (3 Tier M plans: GIT-STRATEGY-SCHEMA + WORKFLOW-SCHEMA-EXTEND + INIT-WIZARD-EXPANSION) + #1040 (config + LSP yaml audit v2 corrections) + #1039 (statusline layout v3 + STATUSLINE-PROFILE-WIZARD + 4-locale docs) + #1003 (§19 AskUserQuestion enforcement canonical cross-ref) + #1002 (dependabot powernap 0.1.4→0.1.5) + #905 (SEMAP M1 Contract schema) + #903 (CI-MULTI-LLM M1 Claude.yml workflow template). 모두 BEHIND state + MERGEABLE 확인 후 `gh pr merge --admin --squash --delete-branch` sequential. #962 conflict는 cherry-pick fallback으로 처리.
- **Hybrid Trunk 1-person OSS config** (commit `cd9eead14`, parallel session): auto-branch/PR + GLM-only LLM review 설정 도입.
- **§23 Local Git Workflows + Hook Setup** (commit `a809e0b98`, parallel session): CLAUDE.local.md에 1-person OSS 워크플로우 doctrine 추가.

### Version

- `v3.0.0-rc1` (major bump from v2.20.0-rc1 doctrine — v3.0 Mega-Sprint W0-W3 누적 흡수). `pkg/version/version.go` default + `.moai/config/sections/system.yaml` `moai.version` / `template_version` 동기화. local git tag `v3.0.0-rc1` (push 안 함, 로컬 release-local + install 검증 목적).

---

## [Unreleased] — v3.0 Mega-Sprint: W0 Claude Refresh + W1 Constitution Dual + W2 Core Slim + W3 Harness Autonomy

### Added

- **[SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001](.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-CLEANUP-001/spec.md)** — Harness Namespace Policy Enforcement & Cleanup: Comprehensive validation + remediation of namespace boundary violations between `moai-harness-*` (builder-only) and `my-harness-*` (user-owned) skill prefixes, addressing CLAUDE.local.md §24 governance layer (Tier S minimal, 7 ACs PASS, backup + regression test coverage). **Scope**: M1 cleanup (6 leaked files from `.claude/agents/harness/` + `moai-harness-{cli-template,patterns}/SKILL.md` deleted with backup at `.moai/backups/harness-namespace-cleanup-2026-05-24T18-53-53Z/`), M2 integration test (`internal/template/embedded_namespace_test.go` 116 LOC, 2 parallel test cases covering template structure invariant + moai-harness-* allowlist enforcement), M3 cross-reference audit (6 SSOT cross-refs verified in `.claude/rules/moai/development/{skill-authoring.md §285-307, agent-authoring.md §13-36}` + 4 moai-adk code locations + 1 template SKILL.md body; all cite §24 without weakening). **Go infrastructure**: `internal/template/embedded_namespace_test.go` introduces `HARNESS_NAMESPACE_LEAK` sentinel via dedicated `TestTemplateMoaiHarnessSkillsAllowlist()` ensuring future compliance at build time. **Pre-existing SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 dependency**: runtime behavior already protected (`internal/cli/update.go` `isMoaiManaged()` safeguard + `isUserOwnedNamespace()` validation + backup pattern + contract-of-evidence). **Run-phase verdict**: 7/7 ACs PASS, plan-auditor iter-1 PASS 0.93 (skip-eligible ≥ 0.75), cross-platform verified, golangci-lint 0 issues. **Attribution note (L46, L52 observed)**: parallel session SPEC-V3R6-MULTI-SESSION-COORD-001 initiated 14-min concurrent run; git status clean, pre-spawn fetch clean, multi-session race scope disjoint verified. **Mx Step C**: SKIP-judge likely (2 Go test files pure declarative, 0 goroutines/complexity≥15/fan_in≥3 per mx-tag-protocol.md §a).

- **[SPEC-V3R6-MULTI-SESSION-COORD-001](.moai/specs/SPEC-V3R6-MULTI-SESSION-COORD-001/spec.md)** — Multi-Session Coordination 4-Layer Race Mitigation Architecture: Session registry + CLI + hook integration + rule extension providing orchestrator-visible active-session tracking and pre-spawn race detection for parallel Claude Code sessions (Tier M, 11 files +2,371 LOC, 93.7% coverage on NEW code). **Package structure**: `internal/session/{registry.go, registry_lock_unix.go, registry_lock_windows.go, registry_test.go, subagent_boundary_test.go}` implement cross-platform session registry with atomic lock-file synchronization for concurrent session state (1,379 NEW LOC, 93.7% coverage, cross-platform tested on linux/amd64 + darwin/amd64 + darwin/arm64 + windows/amd64). `internal/cli/{session.go, session_test.go}` expose `moai session {register,heartbeat,deregister,list,purge}` CLI (474 NEW LOC, 5 verb smoke tests PASS). `internal/hook/session_start.go` (+77 MODIFY) + `session_start_multisession_test.go` (147 NEW) inject 3-step runMultiSessionProtocol into SessionStart hook via helper function (no `.claude/hooks/moai/handle-session-start.sh` modification required — session_id pass-through pre-existed). `.claude/rules/moai/core/agent-common-protocol.md` (+20 LOC) + `.claude/rules/moai/workflow/session-handoff.md` (+3 LOC) + `.claude/output-styles/moai/moai.md` (+5 LOC) extended for source_session_id tracking + Pre-Spawn Sync Check 3-cmd extension + output-style self-check. 3 template mirrors byte-identical. **Architecture highlight (4 layers)**: L1 Go primitive (atomic registry), L2 CLI subcommand (public interface), L3 Hook integration (passive observation of session lifecycle), L4 Pre-spawn assertions (orchestrator-side race detection gate per agent-common-protocol.md §Pre-Spawn Sync Check). **Run-phase verdict**: 16/16 ACs PASS (AC-COORD-001..016 binary verification matrix covering registry operations / CLI 5 verbs / hook integration / rule extensions / cross-platform / pre-spawn commands), plan-auditor iter-2 PASS 0.922 (skip-eligible ≥ 0.90), cross-platform verified, 0 new lint/vet/race-detector issues, C-HRA-008 subagent boundary PASS (0 AskUserQuestion violations). **Known shortfall (user-resolved)**: `internal/session/` package-wide coverage 77.7% (NEW code 93.7%) due to pre-existing SPEC-V3R2-RT-004 checkpoint files; accepted as-is, candidate follow-up SPEC-V3R6-SESSION-LEGACY-COVERAGE-001. **Multi-session empirical evidence**: This SPEC itself documents 3 real race cases in progress.md §C (concurrent HARNESS-PROPOSAL-GEN-001 plan / staging-area race absorbed 14 files during chore / parallel session detection); case-3 staging-area scope drift (20×) catalyzed L4 reinforcement (pre-commit `git diff --cached --name-only` assertion + `git reset` atomic clear). Mx Step C judgment: EVALUATE-PASS (4 NEW Go files candidate for @MX:NOTE + @MX:ANCHOR).

- **[SPEC-V3R6-HARNESS-PROPOSAL-GEN-001](.moai/specs/SPEC-V3R6-HARNESS-PROPOSAL-GEN-001/spec.md)** — V3R4 Self-Evolving Harness Loop Closure: Automatic SPEC Proposal Generator consumes `.moai/harness/learning-history/tier-promotions.jsonl` and emits draft SPEC proposal candidates (Tier M, 14 files +1881 lines, 87.7% coverage). **Package structure** (7 new + 1 modified): `internal/harness/proposalgen/{types.go, reader.go, reader_test.go, mapper.go, mapper_test.go, scaffolder.go, scaffolder_test.go, testdata/}` provide pure-function tier-promotion → ProposalCandidate mapping + on-disk `.moai/proposals/<draft-id>/` scaffolding with spec.md template rendering and proposal.json metadata. `internal/cli/harness/{propose.go, propose_test.go, propose_boundary_test.go}` expose `moai harness propose --dry-run` entry point emitting structured JSON (`{"proposals":[], "reason":"no-actionable-patterns", "malformed_lines":0, "evaluated_patterns":4, "auto_delegate":false}`) + `internal/cli/harness_route.go` registration (+8 lines: 1 import + 7-line `propose` subcommand block). **Graceful no-op contract (REQ-PGN-014)**: current tier-promotions.jsonl data (8 system-event-pattern records, 4 unique pattern_keys) maps to 0 actionable proposals; CLI exits 0 with `reason="no-actionable-patterns"` + `auto_delegate=false` (no orchestrator AskUserQuestion gate). **Subagent boundary (C-HRA-008)**: internal/cli/harness/ package enforces HARD boundary — 0 AskUserQuestion invocations via `TestPropose_NoAskUserQuestion` sentinel. **Run-phase verdict**: 8/8 ACs PASS (AC-PGN-001..008 binary acceptance matrix), plan-auditor iter-1 PASS 0.935 (skip-eligible ≥ 0.90), cross-platform verified, 0 new lint/vet issues. **L52 attribution note**: run-phase deliverables landed on main under commit `24cb6ad4b` whose subject reads `chore(SPEC-V3R6-MULTI-SESSION-COORD-001): progress.md §C plan-auditor signal backfill` due to parallel session indiscriminate staging; the canonical provenance is established via `progress.md §B.2 Multi-Session Attribution Note` and follow-up chore commit `535b5b6ae` (no history rewrite per CLAUDE.local.md §23.5/§23.7).

### Changed

- **[SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001](.moai/specs/SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001/spec.md)** — 정책 codification: manager-spec / manager-develop / manager-docs 3개 매니저 에이전트의 SPEC artifact ownership 경계 명시화 및 상태 전이 소유권 매트릭스 도입 (audit Tier 2 F1+F12 해소, Anthropic best practice #7 DRI ownership at agent-artifact 단위 적용). **Scope**: 3개 agent operational source (.claude/agents/core/{manager-spec,develop,docs}.md) + 3개 template mirror (internal/template/templates/.claude/agents/core/ 동일 경로) + 1개 schema document (.claude/rules/moai/development/spec-frontmatter-schema.md) + progress.md lifecycle tracking = 8파일 총 변경. **Run-phase 7/7 ACs PASS** (AC-ARR-001..007): agent description field updates + Status Transition Ownership Matrix 신규 §추가 + 3개 source/mirror 쌍 byte-identical 동기화 + 스키마 문서 matrix 취소 행 추가 (manager-spec←draft / draft→in-progress @manager-develop / in-progress→implemented @manager-docs / 외 superseded/archived/rejected transitions). **정책 바인딩**: 12-field frontmatter schema 에 canonical Status Transition Ownership Matrix 신규 절 추가 — 어느 agent가 각 상태 전이를 소유하는지 명시. **이 SPEC의 sync-phase는 정책 self-application 첫 사례** — manager-docs가 CHANGELOG + 4개 frontmatter status 만 수정하고 spec.md/plan.md/acceptance.md 본문은 터치 금지 (REQ-ARR-003/006 enforcement). Tier S minimal 1-pass: 4 artifacts (spec/plan/acceptance/progress), 0 design/research, plan-auditor iter-1 PASS 0.875 ≥ 0.75 skip-eligible. **선례**: F1 도출근거 = TMD-001 sync에서 manager-docs가 spec.md body 수정한 archetype incident (2026-05-24); 이 SPEC으로 boundary 폐쇄. **후속**: SIV-001 sync이 신정책 두 번째 적용 케이스 — sync 진행 시 동일 경계 준수 검증 가능.

### Fixed

- **[SPEC-V3R6-SPEC-ID-VALIDATION-001](.moai/specs/SPEC-V3R6-SPEC-ID-VALIDATION-001/spec.md)** — manager-spec SPEC ID regex pre-write self-check protocol + 12-field frontmatter schema fix + test allowlist enrollment (L51 7th protocol implementation). Adds SPEC ID validation against canonical regex literal from `internal/spec/lint.go:573` BEFORE manager-spec body Write call, preventing silent SPEC ID format drift incidents that previously required 3-5 reactive fixup operations. Bundle deliverables: (D1) SPEC ID pre-write validation section + regex literal canonical form + AC sub-ID convention clarification in manager-spec.md (mirror pair); (D2) frontmatter field schema 9→12 canonical update + snake_case rejection table inversion (manager-spec.md mirror pair); (D3) `rule_template_mirror_test.go` allowlist enrollment for manager-spec.md (`TestLateBranchTemplateMirror` now tests manager-spec parity); (D4) regex decomposition print directive. Tier S minimal 1-pass success: 7/7 ACs PASS (AC-SIV-001..007), 3 files ≤ 5 cap, ~80-130 LOC ≤ 300 cap. Run-phase cascade: `catalog.yaml` manager-spec hash regen via canonical `gen-catalog-hashes.go --all`. Plan-auditor PASS-WITH-DEBT 0.815 ≥ 0.75 threshold. **Origin**: L51 lesson promotion from Sprint 7 TMC-001 plan-phase root-cause analysis (5-incident SPEC ID drift chain). **Impact**: Prevents future L32-class incidents via proactive pre-write validation instead of reactive post-lint discovery.

- **[SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001](.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-CASCADE-001/spec.md)** — `internal/template/templates/.claude/skills/moai/workflows/plan/spec-assembly.md` template mirror parity 회복 (TEMPLATE-MIRROR-DRIFT-001 family). Operational source `.claude/skills/moai/workflows/plan/spec-assembly.md` (548줄 / 28,423 bytes)와 mirror byte-for-byte 동등 상태로 복구 (+32줄 / +2,484 bytes — Phase 1.6 Tier Judgment Socratic Question 블록). 발원지: SPEC-V3R5-WORKFLOW-LEAN-001이 Phase 1.6 documentation을 source에 추가했으나 mirror로 전파 누락. 테스트 시그널: `TestLateBranchTemplateMirror/spec-assembly.md` FAIL (`rule_template_mirror_test.go:182` `RULE_TEMPLATE_MIRROR_DRIFT`) → PASS 전환. Tier S minimal 1-pass scope: 1-file mechanical content overwrite. Sprint 2 P4.3 (P4 trio 마지막, P4.1 IVB-001 `d3ed4727d` + P4.2 SARM-001 `5e0dc6a9b` 후속). **5/5 ACs PASS**: AC-TMC-001 (`TestLateBranchTemplateMirror/spec-assembly.md` PASS), AC-TMC-002 (`wc -c mirror` = 28423), AC-TMC-003 (`diff source mirror` = 0), AC-TMC-004 (source 비변경, `git diff 28f783c2a..692f39689 -- source` = 0), AC-TMC-005 (`go vet` 0 + `golangci-lint` `0 issues.`). plan-auditor iter-1 PASS 0.92 (Tier S threshold 0.75, +0.17 margin). Phase 0.5 Plan Audit Gate skip-eligible (0.92 ≥ 0.90 per spec-workflow.md skip policy). Run-phase: M1 commit `5af40acc3` + chore backfill `692f39689` (2 commits, path-specific `git add`, PRESERVE 11 files verbatim). Sibling baseline failures persist per REQ-TMC-006 / L46 attribution discipline — TEMPLATE-MIRROR-DRIFT-001 master SPEC deferred to Sprint 7+ for systematic cleanup of remaining 9+ template mirror drifts.

- **[SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001](.moai/specs/SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001/spec.md)** — Sprint 7 entry SPEC: mechanical 4-mirror cleanup + 1 test registry add + 1 catalog hash regen (6 files total, Tier S minimal). Root cause: 4 `.claude/` source files (spec-workflow.md / agent-common-protocol.md / plan-auditor.md / hooks-system.md) lag behind their `internal/template/templates/.claude/` operational mirrors after intermediate SPECs modified sources without propagating to templates. Run-phase commit `9fe1768e8`: 4 byte-for-byte mirror cp operations (+2654 / +2269 / +2264 / +745 bytes respectively) + `rule_template_mirror_test.go` registry entry add (NEW subtest activation for hooks-system.md per REQ-TMD-005) + `catalog.yaml:160` hash field update (mechanical cascade from A3 plan-auditor.md mirror cp invalidating TestManifestHashFormat — resolved within SPEC per L46 attribution discipline, not PASS-WITH-DEBT). 5/5 ACs PASS (AC-TMD-001..005 binary verification matrix) + TestManifestHashFormat cascade follow-on PASS (50 catalog entries audit). plan-auditor iter-1 PASS 0.94 (skip-eligible ≥ 0.90). L46 attribution: A3 mirror cp direct side-effect (catalog hash invalidation) resolved within SPEC scope (L40 envelope override: declared ≤5 + 1 mechanical follow-up acceptable for Tier S). L44 pre-spawn + pre-commit + post-push fetch all `0 0`. PRESERVE 11 files verbatim. B12 9th self-test PASS.

### Documentation

- **SPEC-V3R6-CLI-AUDIT-001** — moai CLI audit baseline for Sprint 7 FINAL (Tier M research-only)
  - Research deliverable at `.moai/reports/cli-audit/audit-2026-05-23.md` (667 lines, 6 sections)
  - **§1 Subcommand Inventory**: 113 total moai subcommands enumerated (1 root + 23 root-level + 89 nested) with registration paths, flags, and use descriptions
  - **§2 Dead-Command Classification**: 116 classified + 10 dead-suspect candidates identified. Refuted preliminary suspects: `moai hook harness-observe*` family (4 sub-subcommands) — NOT dead, calls documented in `.claude/hooks/moai/handle-harness-observe*.sh`
  - **§3 Integration Map**: `moai init` → `moai update -c` → `moai cc -p <profile>` triad with 10-flag matrix and profile system architecture
  - **§4 Sprint 7 Baseline Scope**: 5-section outline directly consumable by SPEC-V3R6-CLI-INTEGRATION-001 manager-spec (unifications + retirements + gap bridging)
  - **§5 Methodology Appendix**: Grep commands and reproducibility checklist for future audits
  - **6/6 acceptance criteria PASS**: AC-CLA-001 (≥40 inventory entries) PASS / AC-CLA-002 (classification table + dead-suspect coverage) PASS / AC-CLA-003 (integration mapping sections) PASS / AC-CLA-004 (Sprint 7 outline) PASS / AC-CLA-005 (protected-path constraints: 0 `.go`/`.sh`/`.yaml` changes, 0 docs-site modifications) PASS / AC-CLA-006 (metadata completeness: generated-at timestamp, git SHA, 4 artifact frontmatter sync) PASS with AC-CLA-006 minor note (acceptance.md L255-257 awk verification command structural defect in SSOT definition, but deliverable count (6 metadata fields) genuinely PASS)
  - No code, hook script, template, or docs-site content modified (REQ-CLA-005 [Unwanted])
  - **Scope discipline**: 0 Go files, 0 shell scripts, 0 YAML config, 0 template tree modifications; research-only baseline ready for Sprint 7 consumer

### Added

- **SPEC-V3R6-SESSION-HANDOFF-AUTO-001** — SessionEnd hook auto-persist for paste-ready resume messages (Tier S minimal)
  - New package `internal/hook/handoff/` (~1,140 LOC: `persist.go` 383 + `persist_test.go` 823, 85.1% coverage)
  - Hook detects session-handoff pending file and persists 6-block resume message to `~/.claude/projects/{hash}/memory/project_<sprint>_<spec>_<status>.md` with `[SUPERSEDED by ...]` marker for prior entries
  - Cross-platform verified (Windows + Linux 0 errors); C-HRA-008 subagent boundary 0 violations
  - Go-level safeguard for output-style v5.2.0 §6/§8 self-discipline (Trigger 2 SPEC phase complete activation)
  - **Deferred**: AC-SHA-011 path-injection guard (M3 deferred); `resolveMemoryDir` placeholder annotated `@MX:TODO` lines 188-189 of `session_end.go` (follow-up SPEC-V3R6-PROJECT-HASH-RESOLVER-001)

### Changed

- **SPEC-V3R6-LEGACY-CLEANUP-001** — v2.x agency keyword residual cleanup across 31 user-facing files (Tier L, 5-artifact doc-only exemption)
  - **M1 ([ffa65ab15])** — Backup directory `.moai/backups/legacy-cleanup-2026-05-23T103929Z/` (31 files + manifest.json with sha256+bytes per entry) + 4 skills edited (moai-domain-brand-design, moai-domain-copywriting, moai-workflow-gan-loop, moai/workflows/design.md) + 1 rule edited (.claude/rules/moai/design/constitution.md)
  - **M2 ([e517d59e9])** — docs-site ko + en 10 files cleaned (per-locale 5 files: _index.md / gan-loop.md / getting-started.md / migration-guide.md / workflow-commands/moai-design.md). 4-locale parity tracker initialized. Hugo build verified exit 0.
  - **M3 ([42bc8024d])** — docs-site ja + zh 10 files cleaned. Cross-locale parity verified: ko=en=ja=zh=2 retained-reference files (symmetric). Hugo build re-verified exit 0.
  - **M4 ([ccd1fa9cf])** — Root markdown: CLAUDE.md (2→1 agency mention), README.md/ko/ja/zh 4-locale parity edits, CHANGELOG.md preserved untouched per REQ-LCL-007 (append-only historical record). progress.md updated with AC tracker + iter-2 Fix-Forward Log + Known Pre-existing State.
  - **AC results**: 11/11 ACs verified — **9 PASS + 2 PASS-WITH-DEBT** (AC-LCL-003: 16 retained vs ≤5 target due to canonical-locale rationale vs 4-locale verification mismatch; AC-LCL-005: POST_PASS=84 vs PRE=85, marginal i18n-validator timing borderline 31.18s vs 30s budget +4% over). Follow-up SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 (Tier S, budget bump 30s→35s) deferred post-merge.
  - **Scope discipline**: 0 `.go` files modified (AC-LCL-009 PASS), 0 `internal/template/templates/` files modified (AC-LCL-010 PASS), 0 changes in 6 sister SPEC dirs.
  - **5 follow-up SPEC candidates** documented in spec.md §A.6 (LEGACY-CLEANUP-002 template mirror cascade / 003 production Go audit / 004 master design doc / 005 historical SPEC archive) + I18N-VALIDATOR-BUDGET-001 + TEMPLATE-MIRROR-DRIFT-001 (13 pre-existing internal/template test failures).

### Fixed

- **[SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001](.moai/specs/SPEC-V3R6-SKILLS-AUDIT-RUN-MD-001/spec.md)** — Update `internal/template/skills_audit_test.go` solo run.md test entry path to reflect SPEC-V3R4-WORKFLOW-SPLIT-001 Phase 0.5 documentation relocation. Root cause: SPEC-V3R4-WORKFLOW-SPLIT-001 (commit `986418598`, Wave 1) split monolithic `workflows/run.md` into thin router (105 lines) + sub-skill `workflows/run/phase-execution.md` (449 lines) containing all Phase 0.5 documentation. Test `TestSkillsContainPlanAuditGateMarkers` was NOT updated alongside the split — it still asserts 5 patterns (`Phase 0.5: Plan Audit Gate`, `plan-auditor`, `--skip-audit`, `INCONCLUSIVE`, `.moai/reports/plan-audit/`) against the thin router instead of the sub-skill. Tier S minimal scope: 2-line test file edit at `skills_audit_test.go` lines 39-40 (name + filePath). Sprint 2 P4.2. 5/5 ACs PASS: AC-SARM-001 (4/4 sub-tests PASS), AC-SARM-002 (unrelated tests preserved), AC-SARM-003 (full package suite PASS), AC-SARM-004 (byte-identical production templates), AC-SARM-005 (0 new lint/vet findings). plan-auditor iter-3 PASS 0.94 (+0.05 monotonic delta over iter-2 0.89). B12 8th self-test PASS.

- **SPEC-V3R6-LEGACY-CLEANUP-002** — Template mirror cascade for v2.x agency keyword cleanup (Tier S minimal)
  - **M1 ([da5f9906b])** — Mirror 5 surgical agency-keyword cleanup edits from LEGACY-CLEANUP-001 (commits ffa65ab15..19bc873ff) from user-facing `.claude/` paths into `internal/template/templates/.claude/` paths. Restores CLAUDE.local.md §2 [HARD] Template-First Rule compliance: `moai init` now deploys corrected v3.0 agency-free content to new user projects.
  - 5 files mirrored (verbatim `cp -p`): `.claude/rules/moai/design/constitution.md`, `.claude/skills/moai-domain-brand-design/SKILL.md`, `.claude/skills/moai-domain-copywriting/SKILL.md`, `.claude/skills/moai-workflow-gan-loop/SKILL.md`, `.claude/skills/moai/workflows/design.md`.
  - 5/5 ACs binary PASS via SHA-256 verification (all source/target pairs byte-identical).

- **SPEC-V3R6-LEGACY-CLEANUP-003** — production Go Wave→Round terminology cleanup (Tier M per L40 per-SPEC override)
  - **M1 ([80414b8f5])** — Comment-only `Wave` → `Round` renames across 14 production `.go` files per REQ-LCL-001 file-by-file edit map (17 lines across ciwatch/classifier.go, ciwatch/handoff.go, cli/hook.go, cli/pr/watch.go, cli/worktree/{guard,new}.go, config/required_checks.go, harness/types.go, hook/{session_start,spec_status}.go, runtime/budget.go, spec/lint.go, worktree/{doc,state_guard}.go). Per REQ-LCL-001 decision rule, 3 immutable exemptions preserved verbatim: (a) historical SPEC-ID reference `SPEC-V3R3-CI-AUTONOMY-001 Wave 5` at state_guard.go:15 + doc.go:11, (b) file reference `strategy-wave5.md` at state_guard.go:39 (file exists on disk); surrounding comment text rewritten. Scope discipline: 14 production files modified, 0 test files.
  - **M2 ([e432fc276])** — Public API parameter rename in `internal/runtime/persist.go` (REQ-LCL-002): `PersistProgress` + `buildResumeMessage` parameter `waveLabel` → `roundLabel` (5 edits across function signatures + template literal + call site + helper signature + placeholder replacement). Section template literal `- Wave:` → `- Round:` for auto-save message heading. Zero `waveLabel`/`wave_label` residuals remain in persist.go.
  - **M3 ([afb4957f1])** — ResumeMessageFormat default + DefaultFallback constant + budget.go message text (REQ-LCL-003/REQ-LCL-004): `internal/runtime/config.go:28` DefaultFallback `"split_into_waves"` → `"split_into_rounds"`, line 136 ResumeMessageFormat placeholder `{wave_label}` → `{round_label}`, `internal/runtime/budget.go:166` recommendation message `"smaller waves"` → `"smaller rounds"`. Production code residuals post-M3: 0 (excluding budget_test.go forced side-effect scope).
  - **M4 ([9b18493fb])** — Test file forced-side-effect alignment (REQ-LCL-005): `internal/runtime/budget_test.go` 10 edits across test fixtures (lines 24/27/192-193/212/274/382/450/485/491) renaming `"Wave N"` → `"Round N"`, `{wave_label}` → `{round_label}`, `"split_into_waves"` → `"split_into_rounds"`, YAML fixture fallback value. Green-broken-build prevention: line 491 assertion expectation paired with line 485 input rename (both renamed together per §B.5).
  - **8/8 Acceptance Criteria PASS**: AC-LCL-001 (word-boundary grep 0 Wave hits exc. 3 immutable exemptions) / AC-LCL-002 (PersistProgress API 0 waveLabel residuals, 6 roundLabel occurrences) / AC-LCL-003 (ResumeMessageFormat 0 {wave_label}, 1 {round_label}) / AC-LCL-004 (DefaultFallback rename verified, 5+ occurrences) / AC-LCL-005 (budget_test.go 0 Wave fixtures, 5+ Round fixtures) / AC-LCL-006 (go test ./... PASS, zero regression) / AC-LCL-007 (go vet + golangci-lint PASS) / AC-LCL-008 (byte-identical [Unwanted] §A.6 retention: migrate_agency cluster + Copywriter/Designer fields + handle-harness-observe docs + SPEC-V3R3-CI-AUTONOMY-001 historical reference).
  - **3 immutable exemptions preserved per REQ-LCL-001 §a/b/c**: (a) historical SPEC-ID `SPEC-V3R3-CI-AUTONOMY-001 Wave 5` in state_guard.go:15 + doc.go:11 (historical artifact, not subject to rename), (b) file reference `strategy-wave5.md §7` in state_guard.go:39 (file exists on disk; renaming comment would create dead reference), (c) surrounding comment text rewritten.
  - **[Unwanted] §A.6 retention byte-identical**: migrate_agency* cluster (7 files), Copywriter/Designer config fields (internal/config/types.go + defaults.go), frozen.go path constant, handle-harness-observe documentation references (10 occurrences across 6 .go files).
  - **Scope discipline**: 16 production .go files + 1 test file modified (forced side-effect of API rename), 0 changes in other SPEC dirs, 0 docs-site / docs / template modifications (reserved for follow-up LCL-002 cascade).
  - **Run-phase coverage**: 89.5% on internal/runtime package. Zero NEW regressions; pre-existing internal/template test failures (8 baseline failures: TestBackwardCompatibility + TestAgentFrontmatterAudit + TestEmbeddedTemplates_AgentDefinitions + TestLateBranchTemplateMirror + TestAllAgentsInCatalog + TestLoadCatalog + TestRuleTemplateMirrorDrift + TestLoadEmbeddedCatalog_Success + TestSkillsContainPlanAuditGateMarkers + TestRetirementCompletenessAssertion) remain unchanged per L46 attribution rule (pre-existing baseline from sibling SPECs CODE-COMMENTS-EN-001 + HARNESS-RENAME-001 + CORE-SLIM-B-001).

- **[SPEC-V3R6-I18N-VALIDATOR-BUDGET-001](.moai/specs/SPEC-V3R6-I18N-VALIDATOR-BUDGET-001/spec.md)** — i18n-validator TestBudget threshold raised 30s → 35s + function rename to TestBudget_FullRepoScanWithin35Sec (Sprint 2 P4.1, Tier S, 1-file 4-line edit in `scripts/i18n-validator/main_test.go`). Clears SPEC-V3R6-LEGACY-CLEANUP-001 AC-LCL-005 PASS-WITH-DEBT (prior elapsed 31.18s vs 30s budget marginal +4% over; new 35s budget with actual elapsed 3.01s = 11.7x headroom). Run-phase commit `fd58502b7`. plan-auditor iter-1 PASS 0.87 (Tier S 0.80 +0.07). 5/5 ACs PASS (AC-IVB-001 budget value at line 376 / AC-IVB-002 function + JP comment rename at lines 359-360 / AC-IVB-003 renamed test passes 3.01s / AC-IVB-004 elapsed well below 33s warn threshold / AC-IVB-005 no regression in non-budget tests). B12 7th consecutive standing-rule guard self-test PASS.

- **Korean → English code comment translation wave 7 (52 test files)** (SPEC-V3R6-CODE-COMMENTS-EN-001, Tier S Wave 7): Completed Wave 7-2 Test C — 52 Korean test-file comments translated to English across remaining `internal/` scope. Complements Wave 7 partial (commit `b35ca8b96`, 12 test files) for full Wave 7 completion (64 total test files). Scope: `internal/{bodp,brain,coach,cognition,command,completion,conductor,constitution,design,evaluator_leak,github,harness,hook,hooks,lsp,manager,match,mx,permission,plan,promise,render,router,safety,spec,statusline,task,tui}` test files. Committed `ed064a6f2` Wave 7-2. Preserves 5 deliberate Korean comments per EXCL scope decisions (test fixture data, non-source). Multilingual documentation infrastructure (`code_comments: ko` → `en` setting) reinforced. 52 files changed (+52/-52 net). Tier S minimal execution + /moai sync only (no new artifacts, internal test scope only). Wave 7 initiative (W1-W7 cumulative) enables agent reasoning in English-primary codebase while maintaining local documentation in Korean.

- **[SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001](.moai/specs/SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001/spec.md)** — V3R4 harness classifier wiring closure: `/moai harness status` workflow body invokes `moai hook harness-classify` Go subcommand to read usage-log.jsonl, aggregate patterns, classify tier promotions, and write tier-promotions.jsonl (Tier S minimal). Root cause: V3R4 harness learning loop (SPEC-V3R4-HARNESS-001) established 4-tier observability + proposal ladder; wiring was deferred to V3R6 implementation phase. **Implementation (Option A)**: hook subcommand registration in `internal/cli/hook.go` (`harness-classify` use + runHarnessClassify function ~80 LOC + helper readTierThresholds). 3 RED→GREEN tests at new `internal/cli/hook_harness_classify_test.go` (~170 LOC): test no-op when learning disabled (AC-HCW-004), test happy path (4 unique patterns → 4 promotions, AC-HCW-002), test fail-open on corrupt entries (REQ-HCW-003). Sentinel updates: `hook_test.go` / `hook_pre_push_test.go` / `hook_e2e_test.go` subcommand count 35→36 + utility map +1. Workflow body §2.1 status verb step 0 inserted: `moai hook harness-classify 2>&1` invocation with fail-open annotation contract. **6/6 ACs PASS**: AC-HCW-001 (file created, 4 promotion entries), AC-HCW-002 (usage-log 4 unique → promotions 4 unique, parity), AC-HCW-003 (fail-open: corrupt entries silently skipped, summary line emitted), AC-HCW-004 (no-op when learning.enabled=false), AC-HCW-005 (unit tests sentinel updates 0 regression), AC-HCW-006 (golangci-lint + go vet clean). Cross-platform verified (Windows amd64 + Linux x86_64 exit 0). C-HRA-008 subagent boundary 0 AskUserQuestion violations. Scope discipline: 2 new files (`hook_harness_classify_test.go` + internal/harness artifact extensions), 4 modified files (hook.go + 3 test sentinels), 0 changes outside implementation scope. **BC-V3R4-HARNESS-001-CLI-RETIREMENT preserved**: no new `harness` namespace — wiring reuses hook CLI family per contract. plan-auditor iter-1 PASS 0.935 (Tier S 0.75 +0.185 margin, skip-eligible ≥0.90). Run-phase M1 commit `2a3497d59` + chore backfill `103b21695`. B12 10th self-test PASS.

- **`moai update --force` archive contract 정합화 + skip-sync 단락** (SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001, Tier M): `moai update --force` 실행 시 drift backup (`v2.16-drift-<UTC-timestamp>/`) 생성 + archive 덮어쓰기 동작 명시화. `archiveLegacySkills(projectRoot, out, force bool)` 시그니처 확장 (force 인자 추가); `runTemplateSyncWithProgress(cmd) (skipped bool, err)` skip-sync 단락으로 skip-path 버전 일치 시 archive 블록 우회. Template sync 스킵 시 legacy skill archive 블록 미실행 (idempotency 강화, REQ-UAC-004). 신규 테스트 `TestArchiveForce` (4 sub-test) + `TestSkipSyncNoArchive` (2 sub-test) 추가. 기존 호출자 4개 site 시그니처 정합, `runTemplateSyncWithProgress` 2-tuple 수신 정합 4 site. Help text 갱신: "Force update: bypass version-match skip, force backup+merge, and overwrite archive drift (backed up to .moai/archive/skills/v2.16-drift-<UTC-timestamp>/)". AC 9/10 PASS (AC-UAC-008 manual mo.ai.kr reproducer deferred post-merge). 10 files +641/-38. (commit 162581314 + chore 9e0b3db56)

- **`moai update` 반복 경고 노이즈 억제 + `--verbose` flag** (SPEC-V3R6-UPDATE-NOISE-001, Tier S): Reserved filename ack ledger (`.moai/state/reserved-acknowledged.json`) + merge-history ledger (`.moai/cache/merge-history.json`) 도입으로 반복 경고 노이즈 최소화. 동일 파일 구조/hash에 대한 경고 초 제거, 3회 연속 merge fallback 시 "hint: 'moai update -c' to resync templates for <relPath>" 1회 출력 후 침묵 (REQ-UN-002 + REQ-UN-007). `--verbose` flag 추가로 ack-ledger 분기 우회 가능 (모든 경고 재출력). JSON corruption 복구: invalid JSON → silent ledger 재작성 (REQ-UN-011). Ledger 저장: atomic write 패턴 (temp file + rename). 신규 helper 함수 6개 (`loadReservedAckLedger` / `saveReservedAckLedger` / `shouldEmitReservedWarning` / `recordReservedAck` / `loadMergeHistoryLedger` / `recordMergeFallback`). Template `.gitignore` 동기화 (`.moai/state/` + `.moai/cache/` 패턴 추가). AC 12/12 binary PASS. 10 files +895/-20. (commit eb40d0894 + chore debb7af37)

- **`moai update` 진행 라인 깨짐 정정 + TTY/non-TTY 분기** (SPEC-V3R6-UPDATE-PROGRESS-001, Tier S): 22개 legacy `\r  %s` 진행 라인 이미션을 신규 `internal/tui.ProgressLine` API로 통합. ANSI clear prefix `\r\033[2K` (CSI Erase in Line Mode 2) TTY emit, 비-TTY는 plain `\n`-separated 출력으로 분기. `ProgressLine` factory + `ProgressLineHandle` (`Done` / `Fail` / `Update` 메서드) 도입. TTY 감지: `mattn/go-isatty.IsTerminal` 활용. 신규 `internal/tui/progress_line.go` (190 LOC) + `progress_line_test.go` (270 LOC, 8 golden test). Update.go 22 site → 8 진행 라인 영역 collapse (Backup / Validate Templates / Deploy Templates / Restore Settings / Remove paths × 2 / Migrate memory). 기존 메시지 문자 보존 (AC-UPR-006), symbol shift: legacy "-" → success "✓" (skipped not-found 경로는 성공 case로 처리). `symProgress()` 제거 (replacement). Internal/tui coverage 91.5%. AC 9/9 PASS (AC-UPR-008 manual smoke deferred post-merge). 5 files +815/-38. (commit 5fefcd387 + chore fc079abd0)

- **docs-site Hugo theme migration — Hextra → hugo-geekdoc** (4-locale ko/en/ja/zh 유지): Cowork docs-site (`cowork.mo.ai.kr`)와 동일한 hugo-geekdoc v3.0.0 vendored 테마로 전환. Hextra Go modules (`go.mod`/`go.sum`) 제거, `hugo.yaml` → `hugo.toml` 재작성 (4 languages 블록 + `defaultContentLanguageInSubdir = true`). Cowork layouts 14 파일 이식 (`baseof.html`/`site-header.html`/`site-footer.html`/`menu.html`/`foot.html`/`version-badge.html` + 4 shortcodes + custom head/`_markup`). Cowork `moai-brand.css` (62KB) + `moai-design.css` (32KB) + 2 로고 이식 (`moai-cowork-og.png` 제외 — 기존 `og.jpg` 유지). `geekdocMenuBundle: false` (자동 사이드바). 신규 `layouts/shortcodes/callout.html` — Hextra `{{< callout >}}` API → Geekdoc `hint` 자동 매핑으로 기존 콘텐츠 100% 보존. 4 `_index.md` frontmatter migration (`type`/`sidebar`/`toc` 제거) + `guides/_index.md` 4-locale 신규. ADK 브랜딩 정정 (header brand-name `site.Title` 동적 / footer ADK GitHub 링크). Hextra i18n 잔재 (`docs-site/i18n/`) 4 파일 제거. 빌드 결과: 4 locales × 97-105 페이지 × 209 static files, 1.0초. 검증: 4-locale 홈 + sample 페이지 + `moai-brand.css` 모두 HTTP 200, 각 locale title 동적 적용 확인 (`MoAI-ADK 문서`/`Documentation`/`ドキュメント`/`文档`). Migration script: `docs-site/scripts/migrate-frontmatter.sh` (dry-run 지원).

- **SPEC-V3R6-CHANGELOG-CLEANUP-001** — CHANGELOG hallucination cleanup + manager-docs sync-phase B12 standing-rule guard (Tier S minimal)
  - **M1 ([fdd30a94c])** — Line 65 hallucinated SESSION-HANDOFF-AUTO-001 duplicate entry deleted (7-row hallucination catalogue: wrong paths `internal/handoff/{package,atomic_write,parser}.go`, wrong volume "10 files +556/-3", wrong Block 6 "감정", etc.). Line 34-39 correct entry preserved byte-identical (sha256 verified).
  - **M2 ([930eb9420])** — Sibling CHANGELOG AC count reconciliation: HOOK-ASYNC-EXPAND-001 "AC 12/12 PASS" → "AC 8/8 PASS" (matches acceptance.md), HOOK-CWD-LEAK-001 "AC 8/8 PASS" → "AC 7/7 PASS". Counts now match each SPEC's acceptance.md SSOT.
  - **M3 ([87dd61564])** — `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B gained item B12 "Sync-phase CHANGELOG emission discipline" (manager-docs only) — requires Read implementation files before drafting, grep -c duplicate detection pre-flight, acceptance.md AC count verification. First self-test of B12 rule (Tier S gate enforcement).
  - **Scope discipline**: 2 files changed (`CHANGELOG.md` + `manager-develop-prompt-template.md`), 5 sibling SPEC dirs untouched, 0 functional code impact (documentation cleanup only).
  - **5/5 ACs PASS** (AC-CHL-001 ~ AC-CHL-005: binary verification, SSOT alignment, scope discipline, sha256 preservation)

### Added

- **Harness Autonomy — 4-Tier Self-Evolution + 5-Layer Safety + Cold-Start Seeds** (SPEC-V3R5-HARNESS-AUTONOMY-001, W3): 하네스 자율 진화 메커니즘 완성. 7개 신규 패키지: `internal/harness/{capture,router,safety,seeds,throttle,tier}` + root. 18 sentinels (8 HARNESS_FROZEN_* + 10 HARNESS_LEARNING_*) 카탈로그 정의. 10 CLI verbs (route/validate + status/apply/rollback/disable/mute/mute-list/unmute/verify per AC-HRA-009). ≥85% 커버리지 (harness 87.9%, capture 94.9%, router 89.2%, safety 86.5%, seeds 100%, throttle 88.2%, tier 90.0%). 벤치마크: L1 46ns (p99 10ms 대비 우수), L4 1.54µs (p99 100ms 대비 우수). 교차 플랫폼 빌드 PASS (Windows flock split). 메타-분석 결과 SPEC-V3R5-WORKFLOW-OPT-001에서 형식화 (-73% wall-time 검증). 본 SPEC은 그 SPEC의 dogfooding 기준이 됨. PR #1023 plan + PR #1024 run 머지 + sync 완료.

- **Observability hook 3계열 opt-in 마스터 토글 신설** (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 Tier S, commits `18112a7b6..adcb206f2`): `.moai/config/sections/system.yaml` 기존 `hook:` 블록에 NEW `opt_in.enabled: false` 키 추가. 3 observability hook 계열 (`TaskCreated` + `Notification` + `handle-harness-observe-*` secondary wrappers) opt-in 전환 마스터 토글. Default `false` — 매 turn당 hook 호출 ~25-30% 감소. Defense-in-depth: M2 `settings.json.tmpl` conditional render + runtime dispatcher gate (`hookOptInEnabled()` in NEW `internal/hook/hook_opt_in.go`) 양층. **§A.3 3-key cohabitation contract**: 기존 `observability.enabled` (REQ-OBS-005 trace logging, observability.yaml) + `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040 per-event whitelist, system.yaml) 모두 독립 read path 유지 — AC-HOI-007 4-quadrant cohabitation 통합 테스트 + `cohabitation_guard_test.go` static CI guard (5 assertion) 영구 회귀 방어. 7/7 AC PASS. `moai doctor` Hook opt-in 상태 라인 추가 (M3). 3 observability hook 계열 default behavior shift: always-active → opt-in disabled (v3.0 major bump 시그널).

- **Hook 4계열 비동기 확대** (SPEC-V3R6-HOOK-ASYNC-EXPAND-001 Tier M, commits `b00f6afd6..f533b458d`): 4 hook handlers (`FileChanged` + `ConfigChange` + `TaskCreated` + `Notification`) async 전환 (goroutine + `context.WithTimeout 5s` deadline + `sync.WaitGroup`). Observability 조건부 gating: `hook.opt_in.enabled == true` 분기 시에만 async emit (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 R2 partner). NEW `internal/hook/testutil/wait_async.go` helper + benchmark suite (p95 ≤ 100ms per handler). Race detector PASS (5 worker goroutines, concurrent map access 동시성 테스트). Goleak PASS (goroutine 누수 0). AC 8/8 PASS. 16 files +689/-148.

- **Hook cwd leak audit + resolveProjectRoot 일관성** (SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 Tier S, commits `d553d585b..0f4860e17`): 9개 `os.Getwd()` cwd leak sites 정정 (Strategy A `resolveProjectRootFromEnv` vs Strategy B `resolveQualityProjectDir`). Major sites: `internal/hook/{task_created,notification,subagent_start,pre_tool,observability_master}.go` M1/M2 + `internal/quality/gate.go` M3. `ResolveProjectRootFromEnv(homeDir)` 신규 (env `CLAUDE_PROJECT_DIR` fallback `$HOME`) + `resolveQualityProjectDir()` 기존 개선 (`os.Getwd()` 의존 제거). Strategy 적용 규칙: observer (호출자 unknown) → Strategy A, quality gate (caller known) → Strategy B. AC 7/7 PASS. 9 files +178/-145.

- **[SPEC-V3R6-SESSION-LEGACY-COVERAGE-001](.moai/specs/SPEC-V3R6-SESSION-LEGACY-COVERAGE-001/spec.md)** — `internal/session` 패키지 test coverage 보강 (Tier S minimal, run-phase commit `a095bce09`): baseline 77.7% → 85.8% (+8.1%p) 달성, zero 프로덕션 코드 변경 (test-only scope). **4개 test file 변경** (1 NEW + 3 EXTEND): `internal/session/hydrate_test.go` 신규 202L (HydrateForPrompt 0%→~100% + checkpointStatus 0%→~100%, 22 test case), `internal/session/state_test.go` +218L (MarshalJSON 0%→81.8% pointer-receiver round-trip + UnmarshalJSON edge case), `internal/session/store_test.go` +485L (mergePhaseStates Plan/Sync +17.4%p + UTF-8 +6.6%p + RecordBlocker/ResolveBlocker/checkBlockerFiles +multiple% + DetectInFlight 등 7종 보강), `internal/session/checkpoint_test.go` +102L (Plan/Run/Sync Validate enum 11 케이스 모든 status/harness 값). **Coverage improvement**: hydrate 0%→~100% (+100%p NEW module full), state 77.3%→81.8% (+4.5%p), store mergePhaseStates 69.6%→87.0% (+17.4%p), checkpoint validation 66.7%→83.3% (+16.6%p). **Cross-platform verification**: Windows amd64 `go build ./internal/session/...` ✓, Linux x86_64 race detector ✓, Darwin arm64 ✓. **AC verification** (7/7 mandatory PASS): AC-SLCO-001 coverage ≥85.0% PASS (현 85.8%) / AC-SLCO-002 zero production `.go` mutations PASS / AC-SLCO-003 `t.TempDir()` enforcement PASS / AC-SLCO-004 race detector clean PASS / AC-SLCO-005 C-HRA-008 subagent boundary PASS / AC-SLCO-006 Windows cross-build PASS / AC-SLCO-007 OTEL t.Setenv ban PASS. **Quality gates**: golangci-lint 0 NEW / go vet 0 / code review 0. **Race absorption (L52 case 12)**: parallel session SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001 + SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001 concurrent run-phase; pre-spawn fetch `0 0` + L46 L59 disciplines → 0 cross-attribution leakage. **Plan-auditor**: iter-1 PASS 0.945 (Tier S threshold 0.75, +0.195 margin). **Tier classification**: S minimal (4 _test.go files, test-only scope, +807 test lines 0 production). 4 files +807/-0, run-phase 1-pass success. B12 13th CHANGELOG self-test PASS (entry count 1, AC count 7, file paths 5/5 verified).

- **CI baseline drift cleanup — lint 27→0 + TestStatus golden + race fix** (SPEC-V3R6-CI-BASELINE-DRIFT-001 Tier S, commits `fc9ce6822..b159cca36`): Pre-existing baseline (lint 27 + TestStatus golden stale + ConfigManager race) 정리. Lint 27 resolve: errcheck 16 (`os.Remove` + `defer Close` unhandled) / ineffassign 4 / staticcheck 5 / unused 2. TestStatus golden update: v2.17.0 → v3.0.0-rc1 timestamp. ConfigManager race fix: `sync.Mutex` guard on shared `Config` map writes (M3 `internal/quality/gate.go`). AC 8/8 PASS. 5 files +247/-169.

### Changed (Hook opt-in context)

- **3 observability hook 계열 default behavior: always-active → opt-in disabled**: v3.0 major bump 시그널, harness 학습 파이프라인이 telemetry에 의존하는 사용자는 upgrade 후 `system.yaml` `hook.opt_in.enabled: true` 명시 설정 필요. Sibling SPEC `SPEC-V3R6-HOOK-ASYNC-EXPAND-001` (Sprint 2 R2 partner)이 HOI 머지 후 진입하며 `TaskCreated`/`Notification` async stanzas를 `hook.opt_in.enabled == true` 조건부로 래핑.

- **`audit_test.go` baseline 정정** (commit `adcb206f2`): TestAuditRegistrationParity expectedNative 22→20 (M2 conditional rendering note + a3239d3de pre-existing WorktreeCreate/Remove deregistration). TestAuditThreeWaySync 4-way 확장 with `deregisteredButLiveEventNames` allowlist. `TestAuditObservabilityWhitelist` function body는 §A.3 cohabitation contract per UNTOUCHED.

### Tooling (Standing Rules + Output Style)

- **manager-develop-prompt-template Section B — B9/B10/B11 standing rule promotion** (commit `ba85955db`): V3R6 잔여 작업 5+ manager-develop 위임 prompt에 매번 추가됐던 B9/B10/B11 known issues를 `.claude/rules/moai/development/manager-develop-prompt-template.md` Section B에 정식 등재. 향후 prompt 작성 시 자동 포함되어 token 절감 + 일관성 보장.
  - **B9 Git Commit + Push 자체 수행** (Hybrid Trunk 1-person OSS): manager-develop이 main 직진 commit + push 권장. 예외 (parallel race / AC PASS-WITH-DEBT / blocker) 명시. manager-docs에는 적용 안 됨 (sync workflow 자체가 deliverable per [[L15 lesson]]).
  - **B10 Untouched Paths PRESERVE** (Scope Discipline): SPEC plan.md §A.5 PRESERVE list 외 working tree 변경 절대 금지. parallel manager-develop instance 진행 중 특히 주의. runtime-managed files + 무관 SPEC 디렉토리 + research 산출물 보호.
  - **B11 AskUserQuestion 금지** (Subagent Boundary): subagent의 user 직접 interact 금지 (CLAUDE.md §8 + askuser-protocol.md §Orchestrator–Subagent Boundary). Blocker 시 4-옵션 structured blocker report. free-form prose 질문 절대 금지.
  - Byte-identical template mirror 동기화 (`internal/template/templates/.claude/rules/moai/development/`). 2 files +58/-1.

- **output-style moai v5.1.0 → v5.2.0 — Session Handoff template surfaced** (commit `95e3ed247`): `.claude/output-styles/moai/moai.md` §6 "Session Boundary Handoff [HARD]" 5-trigger 추가 (canonical: `.claude/rules/moai/workflow/session-handoff.md` §When To Generate) + §8 Session Handoff [HARD] 6-block format + 5-item pre-emit self-check + auto-memory persistence contract + anti-pattern catalogue. Rationale: session-handoff.md [HARD] rule이 정의되어 있었으나 orchestrator output template에 verbatim format이 없어서 resume message가 self-discipline failure로 skipped 되는 사례 발생. output-style template에 surfacing하여 emit reliability 향상 (no code change required). SESSION-HANDOFF-AUTO-001 (Tier S, line 34/65)의 Go-level safeguard와 함께 dual-layer protection (output-style 자율 + hook 자동 persist). Byte-identical template mirror 동기화. 2 files +150/-4.

## [Unreleased] — v2.20.0-rc1: 11 SPECs complete (RT-002 + RT-003 + RT-006 + CI-FASTTRACK-001 + WORKFLOW-SPLIT-001 + SPC-001 + WF-004 + ORC-002 + ORC-004 + HRN-001 + STATUSLINE-STDINFIELDS-001)

### Added

- **Statusline stdin schema enrichment + 1M handoff threshold tightening** (SPEC-V3R5-STATUSLINE-STDINFIELDS-001): 3 new segment renderers for statusline stdin enrichment. `renderRepoSegment()` exposes GitHub owner/name from `workspace.repo` (v2.1.145+); `renderLongContextSegment()` renders context-warning marker for `exceeds_200k_tokens` (v2.1.139+); `renderHandoffGuideSegment()` displays active handoff thresholds per model class. Internal: `StdinData.ExceedsLongTokens` field mapping, 3 segment predicates (`isRepoEnabled`, `isLongContextEnabled`, `isHandoffGuideEnabled`), task segment renderer integration from SPEC-V3R5-STATUSLINE-V2145-001. Rules updates: CONST-V3R5-022 (1M context threshold tightening from 75% → 50% to align with SSE stall risk envelope), mirror threshold updates in `context-window-management.md` + `session-handoff.md` Trigger #1. Code comments fix: v2.1.122 → v2.1.139 for Effort/Thinking field origin (8 references corrected). 4-locale docs sync: advanced/statusline.md (ko/en/ja/zh) with new segment descriptions maintaining parity within 15%. Coverage: stdinfields_test.go 100% new segments. Tier S, late-branch pattern `feat/SPEC-V3R5-STATUSLINE-STDINFIELDS-001`, squash merge to main.

### Security

- **CWE-732 / CWE-552**: `.claude/settings.local.json` permission hardened to `0o600` (owner-only read/write). Previously `0o644` exposed GLM API tokens (`ANTHROPIC_AUTH_TOKEN`) and other `settings.Env` credentials to other local users on multi-user workstations. Centralized via `secureSettingsMode os.FileMode = 0o600` constant and `writeSettingsSecure` helper in `internal/hook/settings_io.go`; all `session_start.go` / `session_end.go` writers routed through the helper. [b48bd86cb]
- **CWE-214**: `moai cg` tmux env injection now uses the source-file pattern. GLM token (`ANTHROPIC_AUTH_TOKEN`) is written to `~/.moai/run/` temp file (mode `0o600` via `mkstemp` + explicit `chmod 0o600`), injected via `tmux source-file <tmp>`, then unlinked. Token no longer appears in `ps auxe`, `/proc/<pid>/cmdline`, auditd logs, sysmon traces, or crash dumps. Failure returns `ErrTmuxSensitiveInjectFailed` sentinel with **no argv fallback**. `InjectSensitiveEnv` method on `internal/tmux/session.go`; `ensureTmuxGLMEnv` branch in `internal/hook/glm_tmux.go`. [10776c4b8]
- **CWE-345**: `moai update` mandatory checksum verification. `checksums.txt` download is retried 3 times with exponential backoff (2s/4s base delay) via `downloadChecksumWithRetry(checksumsURL, archiveName, maxAttempts, baseDelay)` in `internal/update/checker.go`. On persistent failure, `ErrChecksumUnavailable` sentinel is returned and the update **aborts** — no binary download attempted. Defense-in-depth empty-checksum guard in `downloadAndVerify` (`internal/update/updater.go`). **No `--skip-checksum` opt-out exists**. [ee1335282]
- Cross-cutting verification: 3 GOOS (windows/linux/darwin) builds PASS, race detector PASS on hook/tmux/update, C-HRA-008 subagent boundary grep 0 matches, NEW lint=0 vs 11 pre-existing baseline. Critical NEW function coverage 90.9~96.8%. [b4e7115cb]
- Source SPEC: SPEC-V3R5-SECURITY-CRIT-001 (Tier M, status `implemented` v0.2.0). 11/11 ACs PASS. PR #1032, merge commit `03a2552a2`.
- See [advanced/security-notes](https://adk.mo.ai.kr/ko/advanced/security-notes/) (4 locales: ko/en/ja/zh) for self-audit checklists, threat models, and recovery procedures.

### Added

- **Harness Routing + harness.yaml Go Loader** (SPEC-V3R2-HRN-001): `HarnessConfig` 구조체를 harness.yaml 전체 스키마로 확장. `LoadHarnessConfig()` 검증 강화: FROZEN level enum `{minimal, standard, thorough}`, FROZEN pass_threshold ≥ 0.60 floor, `MOAI_CONFIG_STRICT=1` 스키마 드리프트 감지. 신규 패키지 `internal/harness/router/`: `HarnessRouter.Route(spec, cfg)` Complexity Estimator 기반 레벨 결정 (file_count, domain_count, spec_type, keywords), `EscalationManager.CheckTriggers()` max_escalations 상한 적용, `EffortForLevel()` minimal→medium/standard→high/thorough→xhigh 매핑. CLI: `moai harness route --spec SPEC-XXX [--json]`, `moai harness validate [--path PATH]`. `SPECFrontmatter.HarnessLevel` optional 필드 추가 (REQ-HRN-001-015 spec_override). 4개 sentinel 오류 (`ErrUnknownLevel`, `ErrPassThresholdFloor`, `ErrSchemaDrift`, `ErrEscalationCapExceeded`). `internal/cli/harness_retirement_test.go` CI 가드 업데이트: retired lifecycle 동사 (status/apply/rollback/disable) 차단 + HRN-001 routing 동사 (route/validate) 허용. 88.8% 커버리지 (`internal/harness/router/`). 10 AC (AC-01~10) + 25 tasks 완료. P0 Critical release-blocker.

### Added

- **Sandbox Execution Layer — Bubblewrap (Linux) + Seatbelt (macOS) + Docker (CI)** (SPEC-V3R2-RT-003): 구현 에이전트 tool 호출을 OS-적합 샌드박스 primitive 안에 격리하는 3rd defense-in-depth 레이어 신설. 주요 기능: `Sandbox` 4-값 열거형 (`none|bubblewrap|seatbelt|docker`) + `SandboxBackend` 인터페이스 + `Launcher` 파사드 (`internal/sandbox/`). Linux: `bwrap --unshare-all --die-with-parent` 기반 user-namespace 격리. macOS: `sandbox-exec -p <SBPL-profile>` 기반 Seatbelt 격리 (SBPL profile 결정론적 생성 + 체크섬 안정성). CI: `docker run --rm` 기반 ephemeral container (`CI=1` auto-detect). 기본값: `implementer`/`tester`/`designer` → OS-자동 (seatbelt|bubblewrap), CI=1 → docker; `researcher`/`analyst`/`reviewer`/`architect` → none. 환경변수 스크러빙: `AWS_*` (prefix), `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `NPM_TOKEN`, `GH_TOKEN` 기본 제거 + `sandbox.env_passthrough` opt-in 보존. 16 MiB 출력 트런케이션 + `ErrSandboxOutputTruncated` 반환. `moai doctor sandbox` CLI: backend 가용성 + per-role 해결 결과 + `--profile <role>` 프로파일 덤프. `agent lint` LR-33 규칙: `sandbox: none` without `sandbox.justification` → error. 5 sentinel 오류 (`ErrSandboxBackendUnavailable`, `ErrSandboxProfileInvalid`, `ErrSandboxRequired`, `ErrSandboxOutputTruncated`, `ErrSandboxSetuidDenied`). `internal/config/types.go` 확장: `RoleProfile.Sandbox` 필드 + `SecuritySandbox` 구조체. `security.yaml` 신규 키: `sandbox.required`, `sandbox.network_allowlist`, `sandbox.env_scrub_extra`, `sandbox.docker_image`. Seatbelt p99 ~11ms (50ms 예산 내). 16 AC (AC-01~16) + 52 tasks (T-RT003-01~52) 완료. BC-V3R2-003 (AUTO migration contract 정의). P0 Critical release-blocker. (Breaking: BC-V3R2-003)

### Added

- **8-tier Permission Stack + Bubble Mode** (SPEC-V3R2-RT-002): `internal/permission/` 패키지에 8-소스 permission stack, bubble mode, 5-enum PermissionMode strict 검증, `moai doctor permission` CLI 보강 신설. 주요 기능: `PermissionResolver.Resolve` 8-tier 우선순위 워크 (SrcPolicy > SrcUser > SrcProject > SrcLocal > SrcPlugin > SrcSkill > SrcSession > SrcBuiltin), hook UpdatedInput re-match 단일 재실행 가드, fork depth > 3 → bubble 강등 + SystemMessage 경고, bubble mode parent-unavailable → deny + sentinel, 비대화형 ask → deny fail-closed + `.moai/logs/permission.log` 기록, 동일 tier 충돌 시 specificity-then-fs-order tiebreak (`conflict.go`), legacy `bypassPermissions` action → `acceptEdits` migration + deprecation warning (`migration.go`), strict_mode bypass reject (`spawn.go` `RejectIfStrict`), SrcBuiltin pre-allowlist sync.Once 캐싱 최적화. `moai doctor permission` 플래그 확장: `--all-tiers`, `--mode`, `--fork`, `--format`. `.moai/config/sections/security.yaml` 신규 키: `permission.strict_mode`, `permission.pre_allowlist`, `permission.session_rules`. 16 AC (AC-01~15) + 35 tasks (T-RT002-01~35) 완료. P0 Critical release-blocker.

### Fixed

- **tmux pane leak on SubagentStop (P-H02)** (SPEC-V3R2-RT-006): Critical bug fix for team-mode tmux pane leak on SubagentStop. `subagent_stop.go` now reads `tmuxPaneId` from `~/.claude/teams/{name}/config.json`, invokes `tmux kill-pane` with 500ms per-pane timeout (goroutine + context.WithTimeout), and removes the teammate entry from the registry. Windows path is explicit no-op. (BC-V3R2-018: retire 4 events from settings.json with observability opt-in)

### Added

- **Retire 4 hook events from settings.json with observability opt-in** (SPEC-V3R2-RT-006): `Notification`, `Elicitation`, `ElicitationResult`, `TaskCreated` removed from `settings.json` and `settings.json.tmpl`. Go handlers retained as observability taps: enable via `system.yaml` `hook.observability_events: [notification, ...]`. Pattern A silent return when not opted in. `SystemHookConfig` struct in `internal/config/types.go` with `ObservabilityEvents []string` + `StrictMode bool`.

- **`moai doctor hook` 27-event diagnostic CLI** (SPEC-V3R2-RT-006): New `moai doctor hook` subcommand prints the 27-event coverage table with per-event resolution state (KEEP/UPGRADE/FIX/RETIRE-OBS-ONLY/REMOVE/COMPOSITE). Flags: `--json` (machine-readable), `--trace <event>` (recent log lines from hook.log), `--observability` (filter RETIRE-OBS-ONLY events). `internal/hook/coverage_table.go` is the authoritative 28-entry data structure.

### Changed

- **Worktree MUST Rule for write-heavy role profiles** (SPEC-V3R2-ORC-004): 워크트리 격리를 SHOULD → MUST로 격상. 4개 v3r2 에이전트 frontmatter에 `isolation: worktree` 추가 (expert-backend, expert-frontend, expert-refactoring, researcher). researcher.md body line "when possible" → "mandatory per SPEC-V3R2-ORC-004" (P-A22 self-contradiction 해소). `worktree-integration.md` SHOULD → MUST + 5-에이전트 명시 + Sentinel Key Glossary 신설 (ORC_WORKTREE_MISSING, ORC_WORKTREE_ON_READONLY, ORC_WORKTREE_REQUIRED). `internal/cli/agent_lint.go` LR-05/LR-09 violation message에 sentinel key 추가. `internal/cli/sentinels.go` shared const block (refactor). `moai workflow lint` 신규 CLI — `.moai/config/sections/workflow.yaml` role_profiles 검증 (implementer/tester/designer isolation=worktree 강제). T-ORC004-09 manager-cycle conditional skip (ORC_DEPENDENCY_MISSING — ORC-001 미머지). AC-02 4/5 PASS (manager-cycle deferred — ORC-001 제거됨, `ORC_DEPENDENCY_MISSING` documented path). AC-03/04/05/06/07/08/09/10 모두 PASS. catalog hash 업데이트 (4개 에이전트 수정). v2.20.0-rc1 P1 release-blocker 5건 중 2번째 완료. (PR #982 run + this sync)

- **Agent Common Protocol CI Lint — `moai agent lint`** (SPEC-V3R2-ORC-002): 8-rule lint engine 신설 (`internal/cli/agent_lint.go` ~980 LOC + agent_lint_test.go ~1300 LOC + 9 testdata fixtures). agent-common-protocol.md §User Interaction Boundary [HARD] 규칙을 build-time + CI-time에 정적 검증. Rules: LR-01 literal AskUserQuestion outside fence (error, orchestrator 예외 `isOrchestrator()`), LR-02 `Agent` token in tools CSV (error), LR-03 missing effort (warning [strict: error]), LR-04 dead hook matcher (error), LR-05 missing isolation:worktree (warning [strict: error]), LR-06 --deepthink boilerplate (warning [strict: error]), LR-07 duplicate Skeptical block (error), LR-08 skill-preload drift (warning). Exit codes: 0 clean / 1 error / 2 PARSE_ERROR / 3 IO. `--strict` flag로 LR-03/05/06/08 promotion. `--format=json` schema v1.0 (jq 호환). `--path` flag로 scan root 지정. `.github/workflows/ci.yml`에 `agent lint --strict` step 추가 — PR 머지 차단 게이트 활성화. agent-common-protocol.md §Skeptical Evaluation Stance EVOLVABLE 신설 (6 canonical bullets) + zone-registry.md CONST-V3R2-049 등록. manager-quality.md + sync-auditor.md duplicate Skeptical block 제거 + reference link로 치환. expert-security.md tools CSV에서 dead `Agent` token 제거. `.moai/docs/agent-lint.md` 사용자 문서 + pre-commit hook YAML snippet 신설. 93.5% coverage on agent_lint.go (target 85%). AC-V3R2-ORC-002-01~14 all PASS. v2.20.0-rc1 P0 release-blocker 6건 중 2번째 완료. Unblocks ORC-003 (effort matrix LR-03 promotion to error) + ORC-004 (worktree mandate LR-05 promotion) + MIG-001 (legacy SPEC rewriter validation via lint). (PR #980 run + this sync)

- **Agentless Fixed-Pipeline Classification for Utility Subcommands** (SPEC-V3R2-WF-004): `/moai fix|coverage|mx|codemaps|clean` 5 utility subcommand을 Agentless Pipeline Class로 분류 영구 도입 (Xia et al. 2024 §25 Agentless pattern). 3-phase fixed contract (localize → repair → validate) 강제, LLM dispatcher 금지 (Agent invocation은 phase 내 executor delegation에만 허용). `--mode pipeline` 만 허용, multi-agent subcommand에 `--mode pipeline` 입력 시 `MODE_PIPELINE_ONLY_UTILITY` sentinel (REQ-WF004-014 ↔ WF-003 REQ-WF003-016 shared). Multi-agent subcommand (plan/run/sync/design)에 `--mode <utility>` 입력 시 `MODE_FLAG_IGNORED_FOR_UTILITY` info log (REQ-WF004-011). CI guard `internal/template/agentless_audit_test.go` (REQ-WF004-013) — 5 utility skill에서 LLM-dispatch 패턴 정적 스캔. Subcommand Classification matrix in `.claude/rules/moai/workflow/spec-workflow.md`. Frontmatter dual-schema legacy (`created_at`/`updated_at`/`labels`) 제거 + canonical 12-field 정리 (SPECLINT-DEBT-002 정합). (PR #798 merged 2026-05-09 prior to /moai sync + retrofit sync PR + this lifecycle COMPLETE sync)

- **EARS hierarchical acceptance criteria framework** (SPEC-V3R2-SPC-001): FROZEN-zone CONST-V3R2-001 (`SPEC+EARS format`) amendment 완료. Flat AC (`AC-XXX-NN: Given/When/Then`) → hierarchical tree (`AC-XXX-NN.a`, `.b` 등 inherited Given context + 자체 When/Then) 스키마 영구 도입. Agent-as-a-Judge R1 §9 패턴 채택 (DevAI benchmark 365 sub-requirement 매핑). `internal/spec/ears.go` `Acceptance` 구조체 + `MaxDepth=3` + `GenerateChildID` + `Depth` + `InheritGiven` + `IsLeaf` + `CountLeaves` + `ValidateDepth` + `ExtractRequirementMappings` + `ValidateRequirementMappings` 165 LOC 안정화. Parser auto-wrap (`parser.go:200-227`) 로 185 SPEC 100% flat-corpus backward compat. `moai spec view --shape-trace` CLI 추가. 4/5 CON-002 safety layer PASS (Frozen Guard + Canary 10-SPEC re-parse + Contradiction Detector + Rate Limiter 1/3 used + Human Oversight via maintainer admin merge). v2.20.0-rc1 release-blocker 6 P0 중 1번째 완료. (PR #810 plan + #849 run M2 + #870 run M5 + #925 plan-audit minor defects + this sync)

- **Workflow skills phase-scoped sub-skill split** (SPEC-V3R4-WORKFLOW-SPLIT-001): Bundle F finding 해소. `.claude/skills/moai/workflows/{run,sync,project,plan}.md` 4 monolithic skill (총 4284 LOC) → 4 thin entry router (≤200 LOC each: run 99 / sync 113 / project 110 / plan 144) + 13 phase-scoped sub-skill (≤500 LOC each) 분할. 4-Wave 순차 PR 패턴 (lessons #9 wave-split): Wave 1 PR #973 (run), Wave 2 PR #974 (sync), Wave 3 PR #975 (project), Wave 4 PR #976 (plan self-referential, 3 sub-skill + Path A→B test transition). `internal/template/templates/.claude/skills/moai/workflows/` 동시 mirror 동기화 (dev-only `release-update.md`/`github.md` 제외 100% parity). `.claude/skills/moai/SKILL.md` Intent Router byte-for-byte 무변경 (AC-WFSP-003 frozen invariant). `internal/skills/workflow_split_test.go` 영구 CI fixture 3종 (TestSubSkillLOCCeiling/TestEntryRouterLOCCeiling/TestTemplateMirrorParity) — Wave 4에서 `t.Skip("baseline RED")` 제거하여 AC-WFSP-002 4 entry router 전체 enforce 활성화. EC-1 (clarity-interview borderline) NOT triggered — Wave 4 3 sub-skill outcome (context-discovery 124 / clarity-interview 231 / spec-assembly 500 LOC). v2.20.0-rc1 release-readiness 충족.

- **1-developer CI 3-tier philosophy + paths-filter docs-only fast-path** (SPEC-V3R4-CI-FASTTRACK-001): 사용자 실측 CI wait 5-6분+ → 80% 단축 목표. (축 A) workflow-level optimization 6 task: (T1-T2) ci.yml + codeql.yml paths-filter + skip-marker job 도입 (docs/spec/rule/docs-site markdown PR 즉시 pass); (T3) 5개 영구 RED review workflow 제거 (codex/gemini/glm/llm-panel/claude-code-review.optional) — claude-code-review.yml + review-quality-gate.yml + claude.yml PRESERVE; (T4) 사전 정찰 audit (`private-guard` 비의존성 확인); (T5) lefthook.yml + Makefile `preflight` 로컬 pre-push gate (lint --fast + test -race -short + build); (T6-REVISED) nightly-full-matrix.yml → release-pr-multi-os.yml (일일 cron 대신 release/* branch PR + workflow_dispatch + tag push 트리거, 사용자 2026-05-17 directive 반영); (T7) CLAUDE.local.md §18.7 branch protection doctrine 6→4 항목 정정 + 3-tier philosophy 명문화. (축 B baseline) 사용자가 PRE-PLAN 단계에서 이미 적용한 branch protection rule (ubuntu-latest + lint + build + codeql 4개 required check, macos/windows 제거). 후속: v2.20.0-rc1 release readiness 선결조건 충족. (PR #967 race-merged + #968 plan-revision + #970 run + 이번 sync)

### Fixed

- **AC-V3R2-ORC-002-01 ~ 14 모두 PASS** (SPEC-V3R2-ORC-002): 14 binary acceptance criteria 모두 통과. AC-01 `--help` exit 0 + 4 flags. AC-02 baseline v2.13.2 violations 감지 (9 LR-01 + 4-5 LR-02 + 1 LR-07). AC-03 post-cleanup exit 0. AC-04 + AC-04.a JSON schema v1.0 + version field. AC-04.b pre-commit YAML snippet. AC-05 신규 violation CI red. AC-06 LR-04 dead hook. AC-07 LR-07 duplicate. AC-08/09 LR-03 non-strict warning + strict error. AC-10 fenced-code exempt. AC-11 malformed YAML PARSE_ERROR exit 2. AC-12 canonical Skeptical == 1 in rule file. AC-13 orchestrator carve-out (`isOrchestrator()` 11+ AskUserQuestion 매치에도 0 LR-01). AC-14 tree drift LINT_TREE_DRIFT warning. 7 verification gate all PASS (TDD tests / golangci-lint clean / make build / self-test / `go test ./... -short` / spec lint / template parity). 93.5% coverage on agent_lint.go (target 85%).

- **AC-WF004-001 ~ 016 모두 PASS** (SPEC-V3R2-WF-004): 5 utility subcommand (`fix`/`coverage`/`mx`/`codemaps`/`clean`) 모두 Agentless 3-phase contract 준수. `MODE_PIPELINE_ONLY_UTILITY` + `MODE_FLAG_IGNORED_FOR_UTILITY` sentinel error key 정의. `internal/template/agentless_audit_test.go` CI guard 영구 활성화. Subcommand Classification matrix `.claude/rules/moai/workflow/spec-workflow.md` 등록. v2 `--mode agent` escape hatch 명시적 거부 (§1.2 Non-Goals). Frontmatter canonical schema 정리 (created_at/updated_at/labels legacy 제거).

- **AC-SPC-001-01 ~ 17 + 19 모두 PASS** (SPEC-V3R2-SPC-001): 17 acceptance criteria + AC-019 covering all 18 REQs. AC-SPC-001-01/02/09/14 자체가 hierarchical schema 시연 (parent context inheritance + child leaf 분리). T-SPC001-03 BenchmarkParse365Leaves 6.0ms (<500ms threshold). T-SPC001-04 frontmatter edge case 5 tests. T-SPC001-05 `--shape-trace` audit + 4 tests. T-SPC001-06 spec-workflow.md +37 lines hierarchical schema. T-SPC001-07 SKILL.md +32 lines. T-SPC001-08 zone-registry CONST-V3R2-001 cross-link. T-SPC001-09 MIG-001 §11 handoff note (19 lines). T-SPC001-10 Canary v2 re-parse (`canary-v2-reparse.txt`) — 9/10 SPECs exit 0. T-SPC001-11 3 MX tags (@MX:WARN frozen-zone amendment markers). T-SPC001-12 CON-002 amendment evidence (`con-002-amendment-evidence.md`). plan-auditor MP-1/MP-2 BYPASSED (project convention mismatch — EARS modality block numbering + AC=Gherkin standard). SPC-001 unblocks SPEC-V3R2-SPC-003 (linter) + SPC-V3R2-HRN-002 (Sprint Contract) + SPC-V3R2-HRN-003 (per-leaf scoring) + SPEC-V3R2-MIG-001 (cosmetic AC rewrite).

- **AC-WFSP-001 ~ 008 모두 binary PASS** (SPEC-V3R4-WORKFLOW-SPLIT-001): 4-Wave lifecycle 머지 완료 (PR #973/#974/#975/#976 + sync). AC-WFSP-001 sub-skill LOC ceiling (≤500) — 13 sub-skill 모두 통과 (최대값 spec-assembly.md = 500 LOC exact ceiling, 0 headroom 주의). AC-WFSP-002 entry router LOC ceiling (≤200) — 4 router 통과 (run 99 / sync 113 / project 110 / plan 144). AC-WFSP-003 Intent Router 무변경 — `git diff main -- .claude/skills/moai/SKILL.md` 0 lines (frozen invariant 유지). AC-WFSP-004 slash command regression 0건 — phase trace baseline 보존. AC-WFSP-005 cross-reference integrity — grep `Read workflows/.../` 0 broken refs. AC-WFSP-006 `moai spec lint --strict` ✓ No findings. AC-WFSP-007 docs-site impact analysis 문서화 (grep 0 matches). AC-WFSP-008 template mirror parity 100% (dev-only `release-update.md`/`github.md` 제외). Token-load reduction ~76% (4-workflow aggregate, ~42K → ~10K tokens). Go test 영구 CI fixture 3종 PASS.

- **AC-CIFT-001 ~ 009**: Run-PR #970 머지로 9 acceptance criteria 모두 binary PASS. AC-CIFT-001 docs-only fast-path via paths-filter (skip-marker job match github branch protection canonical names), AC-CIFT-002a/b CodeQL skip-marker pattern + empirical canonical name resolution, AC-CIFT-003 review consolidation 5 delete + 1 preserve (codex-review/gemini-review/glm-review/llm-panel/claude-code-review.optional 삭제), AC-CIFT-004 private-guard audit (grep 0건 cli 호출), AC-CIFT-005 lefthook.yml + Makefile preflight gate, AC-CIFT-006 release-pr-multi-os.yml (3-OS full matrix, release/* branch PR + workflow_dispatch + tag push 트리거), AC-CIFT-007 CLAUDE.local.md §18.7 doctrine sync (4 required checks + 3-tier philosophy), AC-CIFT-008 lessons.md #19 entry 5-section protocol structure (Category/Incorrect/Correct/Why/How-to-apply), AC-CIFT-009 `go test ./...` 0 failures. Workflow count 20 → 16 (delta -4). v2.20.0-rc1 release-readiness 최종 precondition.

### Known Follow-up

- D5/D7/D8/D9 P1 deferred (SHA pinning already applied in run-phase, AC-CIFT-009 orphan reclass, OOS additional items, R6-R9 risks) → optional hot-fix PR or rolled into sync. 
- D10-D15 P2/P3 optional (review-bot env failures, deprecated CI platform, deprecated node version, etc.).
- SPEC-WORKTREE-SKILLS-CLEANUP-001 (가칭): `.claude/skills/moai/{workflows,team}/run.md` stale `[HARD] worktree` references AC-WTD-007 related.

### Housekeeping (2026-05-18 — release-readiness 정합성 정리)

- **SPEC-OPUS47-COMPAT-001 status drift 정정** (`in-progress → completed`, v0.2.0 → v0.3.0): plan-auditor가 2026-04-24 v0.2.0 HISTORY entry에서 "status draft → completed (PR #672/#673 merged)" 명시했으나 frontmatter `status: in-progress` 잔존 drift 정정. 코드 변경 0 (이미 머지된 작업의 메타데이터만). v2.20.0-rc1 release notes에서 "Claude Code v2.1.110/111 + Opus 4.7 호환성 기반 인프라" reference 가능 상태.
- **SPEC-HOOK-002~007 6 묵은 draft archive** (`draft → archived`, v0.1.0 → v0.2.0, 2026-02-04 작성 ~3.5개월 묵음): 본 세션 완료 SPEC들에 의해 대부분 absorbed되어 archive 처리. HOOK-002 Code Quality Automation (~90% overlap with `moai gate` + ORC-002 + ORC-004), HOOK-003 Security & Scanning (~80% with RT-003 Sandbox + ORC-002 + RT-002), HOOK-004 LSP Diagnostics (~70% with LSPMCP-001/LSP-CORE-002 separate track), HOOK-005 Git Operations Manager (~85% with manager-git agent + git_*.go), HOOK-006 Resilience Patterns (~75% with RT-002 + RT-003 + RT-006 timeout), HOOK-007 Session Lifecycle Enhancements (~95% with RT-006 27-Event Coverage). 각 spec.md HISTORY v0.2.0 entry에 supersede 출처 명시. `moai spec lint --strict` ✓ No findings 유지.

### Changed (English)

- **Agent Common Protocol CI Lint — `moai agent lint`** (SPEC-V3R2-ORC-002): 8-rule lint engine introduced (`internal/cli/agent_lint.go` ~980 LOC + agent_lint_test.go ~1300 LOC + 9 testdata fixtures). Statically validates agent-common-protocol.md §User Interaction Boundary [HARD] rules at build-time + CI-time. Rules: LR-01 literal AskUserQuestion outside fence (error, orchestrator carve-out via `isOrchestrator()`), LR-02 `Agent` in tools CSV (error), LR-03 missing effort (warning [strict: error]), LR-04 dead hook matcher (error), LR-05 missing isolation:worktree (warning [strict: error]), LR-06 --deepthink boilerplate (warning [strict: error]), LR-07 duplicate Skeptical block (error), LR-08 skill-preload drift (warning). Exit codes: 0 clean / 1 error / 2 PARSE_ERROR / 3 IO. `--strict` promotes LR-03/05/06/08. `--format=json` schema v1.0 (jq compatible). `--path` flag for scan root. `.github/workflows/ci.yml` `agent lint --strict` step landed — PR merge gate active. agent-common-protocol.md §Skeptical Evaluation Stance EVOLVABLE introduced (6 canonical bullets) + zone-registry.md CONST-V3R2-049 entry. manager-quality.md + sync-auditor.md duplicate Skeptical block removed + replaced with reference link. expert-security.md dead `Agent` token removed from tools CSV. `.moai/docs/agent-lint.md` user documentation + pre-commit hook YAML snippet. 93.5% coverage on agent_lint.go (target 85%). 2nd of 6 P0 release-blockers for v2.20.0-rc1. Unblocks ORC-003 (effort matrix LR-03 promotion) + ORC-004 (worktree mandate LR-05 promotion) + MIG-001 (legacy SPEC rewriter validation). (PR #980 run + this sync)

- **Agentless Fixed-Pipeline Classification for Utility Subcommands** (SPEC-V3R2-WF-004): 5 utility subcommands (`/moai fix|coverage|mx|codemaps|clean`) permanently classified as Agentless Pipeline Class (Xia et al. 2024 §25 pattern). 3-phase fixed contract (localize → repair → validate) enforced; LLM dispatcher prohibited (Agent invocation allowed only as executor delegation within a phase). Sentinel error keys: `MODE_PIPELINE_ONLY_UTILITY` (REQ-WF004-014 ↔ WF-003 REQ-WF003-016 shared) blocks `--mode pipeline` on multi-agent subcommands; `MODE_FLAG_IGNORED_FOR_UTILITY` (REQ-WF004-011) info-logs `--mode <any>` on utility subcommands. CI guard `internal/template/agentless_audit_test.go` (REQ-WF004-013) statically scans 5 utility skills for LLM-dispatch patterns. Subcommand Classification matrix landed in `.claude/rules/moai/workflow/spec-workflow.md`. Frontmatter dual-schema legacy (`created_at`/`updated_at`/`labels`) cleaned up + canonical 12-field schema applied per SPECLINT-DEBT-002. (PR #798 merged 2026-05-09 prior to /moai sync + retrofit sync PR + this lifecycle COMPLETE sync)

- **EARS hierarchical acceptance criteria framework** (SPEC-V3R2-SPC-001): FROZEN-zone CONST-V3R2-001 (`SPEC+EARS format`) amendment landed. Flat AC (`AC-XXX-NN: Given/When/Then`) → hierarchical tree (`AC-XXX-NN.a`, `.b` etc. with inherited Given context + per-leaf When/Then) schema permanently introduced. Adopts Agent-as-a-Judge R1 §9 pattern (DevAI benchmark 365 sub-requirement mapping). `internal/spec/ears.go` `Acceptance` struct + `MaxDepth=3` + `GenerateChildID` + `Depth` + `InheritGiven` + `IsLeaf` + `CountLeaves` + `ValidateDepth` + `ExtractRequirementMappings` + `ValidateRequirementMappings` 165 LOC stabilized. Parser auto-wrap (`parser.go:200-227`) preserves 100% backward compatibility for 185-SPEC flat corpus. `moai spec view --shape-trace` CLI added. 4/5 CON-002 safety layers PASS (Frozen Guard + Canary 10-SPEC re-parse + Contradiction Detector + Rate Limiter 1/3 used + Human Oversight via maintainer admin merge). 1st of 6 P0 release-blockers for v2.20.0-rc1. (PR #810 plan + #849 run M2 + #870 run M5 + #925 plan-audit minor defects + this sync)

- **Workflow skills phase-scoped sub-skill split** (SPEC-V3R4-WORKFLOW-SPLIT-001): Bundle F finding resolved. `.claude/skills/moai/workflows/{run,sync,project,plan}.md` 4 monolithic skills (total 4284 LOC) refactored into 4 thin entry routers (≤200 LOC each: run 99 / sync 113 / project 110 / plan 144) + 13 phase-scoped sub-skills (≤500 LOC each). 4-Wave sequential PR pattern (lessons #9 wave-split): Wave 1 PR #973 (run), Wave 2 PR #974 (sync), Wave 3 PR #975 (project), Wave 4 PR #976 (plan self-referential, 3 sub-skills + Path A→B test transition). `internal/template/templates/.claude/skills/moai/workflows/` mirror sync (100% parity excluding dev-only `release-update.md`/`github.md`). `.claude/skills/moai/SKILL.md` Intent Router byte-for-byte unchanged (AC-WFSP-003 frozen invariant). `internal/skills/workflow_split_test.go` 3 permanent CI fixtures (TestSubSkillLOCCeiling/TestEntryRouterLOCCeiling/TestTemplateMirrorParity) — Wave 4 removed `t.Skip("baseline RED")` to activate AC-WFSP-002 enforcement across all 4 entry routers. EC-1 (clarity-interview borderline) NOT triggered — Wave 4 produced 3 sub-skills (context-discovery 124 / clarity-interview 231 / spec-assembly 500 LOC). v2.20.0-rc1 release-readiness precondition satisfied.

- **1-developer 3-tier CI philosophy + paths-filter docs-only fast-path** (SPEC-V3R4-CI-FASTTRACK-001): Measured user CI wait 5-6min+; goal 80% reduction. (Axis A) workflow-level optimizations (6 tasks): (T1-T2) ci.yml + codeql.yml paths-filter + skip-marker job for docs/spec/rule/docs-site markdown PR instant pass; (T3) 5 permanently RED review workflows deleted (codex/gemini/glm/llm-panel/claude-code-review.optional) — claude-code-review.yml + review-quality-gate.yml + claude.yml PRESERVED; (T4) pre-flight audit (private-guard codex-independence verified); (T5) lefthook.yml + Makefile `preflight` local pre-push gate (lint --fast + test -race -short + build); (T6-REVISED) nightly-full-matrix.yml → release-pr-multi-os.yml (cron schedule replaced with release/* branch PR + workflow_dispatch + tag push per user 2026-05-17 directive); (T7) CLAUDE.local.md §18.7 branch protection doctrine 6→4 items + 3-tier philosophy formalized. (Axis B baseline) user pre-plan branch-protection-rule application (ubuntu-latest + lint + build + codeql 4 required, macos/windows removed). Follow-up: v2.20.0-rc1 release-readiness precondition satisfied. (PR #967 race-merged + #968 plan-revision + #970 run + this sync)

### Fixed (English)

- **AC-V3R2-ORC-002-01 ~ 14 all PASS** (SPEC-V3R2-ORC-002): 14 binary acceptance criteria all PASS. AC-01 `--help` exit 0 + 4 flags. AC-02 baseline v2.13.2 violations detected. AC-03 post-cleanup exit 0. AC-04 + AC-04.a JSON schema v1.0 + version field. AC-04.b pre-commit YAML snippet. AC-05 new violation triggers CI red. AC-06 LR-04 dead hook. AC-07 LR-07 duplicate. AC-08/09 LR-03 non-strict warning + strict error. AC-10 fenced-code exempt. AC-11 malformed YAML PARSE_ERROR exit 2. AC-12 canonical Skeptical == 1 in rule file. AC-13 orchestrator carve-out (`isOrchestrator()`). AC-14 tree drift LINT_TREE_DRIFT warning. 7 verification gates all PASS. 93.5% coverage on agent_lint.go.

- **AC-WF004-001 ~ 016 all PASS** (SPEC-V3R2-WF-004): 5 utility subcommands (`fix`/`coverage`/`mx`/`codemaps`/`clean`) comply with Agentless 3-phase contract. Sentinel error keys `MODE_PIPELINE_ONLY_UTILITY` + `MODE_FLAG_IGNORED_FOR_UTILITY` defined. CI guard `internal/template/agentless_audit_test.go` permanently active. Subcommand Classification matrix registered in `.claude/rules/moai/workflow/spec-workflow.md`. v2 `--mode agent` escape hatch explicitly rejected (§1.2 Non-Goals). Frontmatter canonical schema cleanup.

- **AC-SPC-001-01 ~ 17 + 19 all PASS** (SPEC-V3R2-SPC-001): 17 acceptance criteria + AC-019 covering all 18 REQs. AC-SPC-001-01/02/09/14 self-demonstrate hierarchical schema (parent context inheritance + child leaf separation). T-SPC001-03 BenchmarkParse365Leaves 6.0ms (<500ms threshold). T-SPC001-04 frontmatter edge cases 5 tests. T-SPC001-05 `--shape-trace` audit + 4 tests. T-SPC001-06 spec-workflow.md +37 lines hierarchical schema. T-SPC001-07 SKILL.md +32 lines. T-SPC001-08 zone-registry CONST-V3R2-001 cross-link. T-SPC001-09 MIG-001 §11 handoff note (19 lines). T-SPC001-10 Canary v2 re-parse (9/10 SPECs exit 0). T-SPC001-11 3 MX tags (`@MX:WARN` frozen-zone amendment markers). T-SPC001-12 CON-002 amendment evidence. plan-auditor MP-1/MP-2 BYPASSED (project convention mismatch — EARS modality block numbering + AC=Gherkin standard). SPC-001 unblocks SPEC-V3R2-SPC-003 (linter) + SPC-V3R2-HRN-002 (Sprint Contract) + SPC-V3R2-HRN-003 (per-leaf scoring) + SPEC-V3R2-MIG-001 (cosmetic AC rewrite).

- **AC-WFSP-001 ~ 008 all binary PASS** (SPEC-V3R4-WORKFLOW-SPLIT-001): 4-Wave lifecycle merged (PR #973/#974/#975/#976 + sync). AC-WFSP-001 sub-skill LOC ceiling (≤500) — all 13 sub-skills pass (max spec-assembly.md = 500 LOC exact ceiling, 0 headroom — future additions must trigger sub-split). AC-WFSP-002 entry router LOC ceiling (≤200) — 4 routers pass (run 99 / sync 113 / project 110 / plan 144). AC-WFSP-003 Intent Router unchanged — `git diff main -- .claude/skills/moai/SKILL.md` 0 lines (frozen invariant). AC-WFSP-004 slash command regression 0 — phase trace baseline preserved. AC-WFSP-005 cross-reference integrity — grep `Read workflows/.../` 0 broken refs. AC-WFSP-006 `moai spec lint --strict` ✓ No findings. AC-WFSP-007 docs-site impact analysis documented (grep 0 matches). AC-WFSP-008 template mirror parity 100% (dev-only `release-update.md`/`github.md` excluded). Token-load reduction ~76% (4-workflow aggregate, ~42K → ~10K tokens). Go test 3 permanent CI fixtures PASS.

- **AC-CIFT-001 ~ 009 all binary PASS**: Run-PR #970 merge. AC-CIFT-001 docs-only fast-path via paths-filter (skip-marker job matches github branch-protection canonical names), AC-CIFT-002a/b CodeQL skip-marker pattern + empirical canonical-name resolution, AC-CIFT-003 review consolidation (5 deleted + 1 preserved), AC-CIFT-004 private-guard audit (grep 0 CLI references), AC-CIFT-005 lefthook.yml + Makefile preflight, AC-CIFT-006 release-pr-multi-os.yml (3-OS full matrix, release/* PR + workflow_dispatch + tag push), AC-CIFT-007 CLAUDE.local.md §18.7 doctrine (4 required checks + 3-tier philosophy), AC-CIFT-008 lessons.md #19 entry 5-section protocol, AC-CIFT-009 `go test ./...` pass 0 failures. Workflow count 20 → 16 (delta -4). v2.20.0-rc1 release-readiness final precondition.

### Housekeeping (English, 2026-05-18 — release-readiness consistency)

- **SPEC-OPUS47-COMPAT-001 status drift fix** (`in-progress → completed`, v0.2.0 → v0.3.0): plan-auditor's 2026-04-24 v0.2.0 HISTORY entry stated "status draft → completed (PR #672/#673 merged)" but the frontmatter remained `status: in-progress`. Corrected the drift. Zero code changes (work already merged). Enables v2.20.0-rc1 release notes to cite the SPEC as base infrastructure for "Claude Code v2.1.110/111 + Opus 4.7 compatibility".
- **SPEC-HOOK-002~007 archive of 6 stale drafts** (`draft → archived`, v0.1.0 → v0.2.0, originally created 2026-02-04 ~3.5 months stale): Largely superseded by this session's completed SPECs. HOOK-002 Code Quality Automation (~90% overlap with `moai gate` + ORC-002 + ORC-004), HOOK-003 Security & Scanning (~80% with RT-003 Sandbox + ORC-002 + RT-002), HOOK-004 LSP Diagnostics (~70% with LSPMCP-001/LSP-CORE-002 separate track), HOOK-005 Git Operations Manager (~85% with manager-git agent + git_*.go), HOOK-006 Resilience Patterns (~75% with RT-002 + RT-003 + RT-006 timeout), HOOK-007 Session Lifecycle Enhancements (~95% with RT-006 27-Event Coverage). Each spec.md HISTORY v0.2.0 entry documents supersession sources. `moai spec lint --strict` remains ✓ No findings.

### Known Follow-up (English)

- D5/D7/D8/D9 P1 deferred (SHA pinning already applied in run-phase, AC-CIFT-009 orphan reclass, OOS items, R6-R9 risks).
- D10-D15 P2/P3 optional (review-bot env, deprecated CI, node version, etc.).
- SPEC-WORKTREE-SKILLS-CLEANUP-001 (tentative): stale `[HARD] worktree` refs in `.claude/skills/moai/{workflows,team}/run.md` (AC-WTD-007 related).

## [Unreleased] — v2.20.0-rc1 Doctrine: SPEC-WORKTREE-DOCS-001 Worktree Workflow Harmonization (L1/L2/L3 opt-in)

### Changed

- **doctrine harmonization: worktree workflow rules SHOULD-tier로 격하 + L1/L2/L3 terminology 표준화** (SPEC-WORKTREE-DOCS-001): 2026-05-17 사용자 자율적 정책 (`feedback_worktree_autonomous` 메모리)을 5 rule 파일에 영구 흡수. (a) `spec-workflow.md` Step 2/3 `[HARD] MUST create/reuse worktree` → `[SHOULD] L2 opt-in` (8 HARD → 3, -5건); (b) `CLAUDE.md` §14 4 worktree `[HARD]` bullets → `[SHOULD]` advisory (Claude Code runtime이 L1 isolation 자율 결정 명시); (c) `worktree-integration.md`에 4-row Terminology Glossary 신설 (L1 Claude Code Native / L2 SPEC worktree / L3 Plan worktree / git worktree 구분); (d) `worktree-state-guard.md` Wave 5 primitive dormancy 노트 추가 (orchestrator wiring deferred, manual invocation 보존); (e) `session-handoff.md` Block 0 conditional wording (`--worktree` L3 opt-in 시에만 required 명시). 모든 5 파일에 user policy 2026-05-17 cross-reference 추가. 추가: `scripts/audit-workflow-terminology.sh` (~41 LOC) terminology drift 자동 detect. 메모리 hierarchy: `feedback_worktree_never_use` (2026-05-15 영구 미사용 정책) heading + MEMORY.md 인덱스 supersede marker 부착, `feedback_worktree_autonomous` (pre-existing) compliance verified. 본 SPEC은 documentation-only — Go code 0 LOC 변경, `go test ./...` exit 0. AC-WTD-001~009 all PASS (007 non-blocking, 2 stale `[HARD]` references in `.claude/skills/moai/{workflows,team}/run.md` → follow-up SPEC 후보로 surface). (PR #964 plan + #965 run + 이번 sync, plan-in-main + main-checkout flow 적용, L2/L3 worktree 미사용)

### Fixed

- **AC-WTD-001 ~ 009**: Run-PR #965 머지로 9 acceptance criteria 모두 binary PASS. AC-001 force-worktree [HARD] removal (grep 0건), AC-002 glossary 4-row present, AC-003 audit script exit 0, AC-004 user policy reference ≥4, AC-005 Wave 5 primitive (`internal/cli/worktree/guard.go` 6 functions) intact + dormancy banner, AC-006 memory hierarchy integrity (F1 frontmatter 4/4 + body 4/4, F2 SUPERSEDED ≥1, F3 references ≥1), AC-008 `moai spec lint --strict` 0 errors, AC-009 `go test ./...` exit 0.
- **doctrine consistency v2.20.0-rc1 release-readiness**: 2026-05-15 `feedback_worktree_never_use` (worktree 영구 미사용) 정책이 글로벌 룰의 `[HARD] MUST` mandate와 contradict하던 dormant conflict가 영구 해소. v2.20.0-rc1 release timing에 적합.

### Known Follow-up

- **AC-WTD-007 비차단 findings** (2건): `.claude/skills/moai/workflows/run.md:1001` + `.claude/skills/moai/team/run.md:349`에 `[HARD] All implementation teammates MUST use isolation: "worktree"` 잔존. plan.md R1 mitigation에 따라 본 SPEC scope 외 → 후속 SPEC `SPEC-WORKTREE-SKILLS-CLEANUP-001` (가칭) 후보. v2.20.0-rc1 release timing에 차단 영향 없음.
- **review bot env failures** (3건 review + 1 private-guard `codex: command not found`): branch protection required check 외 → SPEC-V3R4-LLM-REVIEW-CI-001 (가칭) 후속. 본 PR 머지 차단 없음.

### Changed (English)

- **doctrine harmonization: worktree workflow rules downgraded to SHOULD tier + L1/L2/L3 terminology standardized** (SPEC-WORKTREE-DOCS-001): The 2026-05-17 user autonomous policy (`feedback_worktree_autonomous` memory) is permanently absorbed into 5 rule files. (a) `spec-workflow.md` Step 2/3 `[HARD] MUST create/reuse worktree` → `[SHOULD] L2 opt-in` (8 HARD → 3, -5 total); (b) `CLAUDE.md` §14 4 worktree `[HARD]` bullets → `[SHOULD]` advisory (Claude Code runtime decides L1 isolation per-call); (c) `worktree-integration.md` gains a 4-row Terminology Glossary distinguishing L1 (Claude Code Native), L2 (SPEC worktree), L3 (Plan worktree), and git worktree (low-level mechanism); (d) `worktree-state-guard.md` gains a Wave 5 primitive dormancy banner (orchestrator wiring deferred, manual invocation preserved); (e) `session-handoff.md` Block 0 conditional wording (only required when `--worktree` L3 opt-in). All 5 files cross-reference user policy 2026-05-17. New `scripts/audit-workflow-terminology.sh` (~41 LOC) auto-detects future terminology drift. Memory hierarchy: `feedback_worktree_never_use` (2026-05-15 perpetual-no-worktree policy) heading + MEMORY.md index marked `[SUPERSEDED]`, `feedback_worktree_autonomous` (pre-existing) compliance verified. This SPEC is documentation-only — 0 LOC Go code change, `go test ./...` exit 0. AC-WTD-001~009 all binary PASS (007 non-blocking, 2 stale `[HARD]` references in `.claude/skills/moai/{workflows,team}/run.md` surfaced as follow-up SPEC candidates). (PR #964 plan + #965 run + this sync, plan-in-main + main-checkout flow applied, L2/L3 worktree not used)

## [Unreleased] — v2.20.0-rc1 Governance: SPEC-V3R4-HARNESS-NAMESPACE-001 Harness Namespace + Lifecycle Governance closeout

### Changed

- **status drift sweep: 17건 status drift 일괄 해소** (SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002): SPEC-V3R4-LINT-SPECID-GREP-FIX-001 walker word-boundary fix 후 노출된 substring noise drift 17건 정밀 분석 + Category A (5건 frontmatter sync-up) + Category B (12건 per-SPEC mechanism: 3건 sync-commit 적용 + 9건 lint.skip 영구 적용). LSGF-001 → FOLLOWUP-002 lineage로 walker 정밀화 직후 hidden drift 영구 해소. SDF-001 Pattern H precedent 적용 (Wave 1 analysis → Wave 2 apply → Wave 3 verify). `moai spec lint --strict` 0 error / 0 warning binary 달성 (v2.20.0-rc1 release-readiness precondition 충족). (PR #950 plan + #951 run + 이번 sync)

### Fixed

- **CI**: SPEC-V3R4-CI-INFRA-FIX-001 — GitHub Actions CI infrastructure 3 결함 영구 해소: (A) `.github/actions/detect-language/action.yml` SIGPIPE remediation (head -1 → find -print -quit 단일-step 솔루션, awk NR==1 D-1 A2 hotfix replacement); (B) `.github/workflows/spec-status-auto-sync.yml` workflow-level permissions block 추가 (contents: write + issues: write); (C) `.github/workflows/ci.yml` 5 actions/checkout step 모두 `fetch-depth: 0` 명시 + `internal/spec/drift_specid_grep_test.go` 의 `GITHUB_ACTIONS` env skip guard 영구 제거. AC-CIIF-001 SPIRIT 충족 (verification command override 문서화), AC-CIIF-003 CI 검증 완료. v2.20.0-rc1 release-readiness 최종 precondition. (PR #954 plan + #955 run + #956 sync)
- **CI (English)**: SPEC-V3R4-CI-INFRA-FIX-001 — Permanently resolved 3 GitHub Actions CI infrastructure defects: (A) `.github/actions/detect-language/action.yml` SIGPIPE remediation (head -1 → find -print -quit single-step solution, hotfix replacement for failed awk NR==1 D-1 A2 attempt); (B) `.github/workflows/spec-status-auto-sync.yml` workflow-level permissions block (contents: write + issues: write); (C) `.github/workflows/ci.yml` 5 actions/checkout steps explicit `fetch-depth: 0` + permanently removed `GITHUB_ACTIONS` env skip guard from `internal/spec/drift_specid_grep_test.go`. AC-CIIF-001 SPIRIT met (verification command override documented), AC-CIIF-003 CI-verified. Final precondition for v2.20.0-rc1 release-readiness. (PR #954 plan + #955 run + #956 sync)
- **AC-SDF002-X-001**: `moai spec lint --strict` 0 error / 0 warning binary 달성 (v2.20.0-rc1 release-readiness precondition 충족).
- **AC-SDF002-X-002**: V3R4-HARNESS-001/002/003 (LSGF-001 targets) 무영향 — 회귀 0.
- **AC-SDF002-X-003**: 소스 코드 미변경 (metadata-only sweep).

### Known Follow-up

- 9 Category B lint.skip 잔존 (B1/B2/B4/B5/B7/B8/B10/B11/B12) — sync-commit 전략을 적용하면 chain self-drift 유발 가능성으로 lint.skip 영구 적용 결정. 향후 walker 개선 (예: chore-skip semantics 확장) SPEC이 진입하면 일괄 cleanup 가능.
- CI shallow clone (HARNESS001Resolution test skip): SPEC-V3R4-CI-INFRA-FIX-001 (다음 SPEC) `fetch-depth: 0` 영구 fix 예정.
- v2.20.0-rc1 release tagging: CI-INFRA-FIX-001 lifecycle COMPLETE 후 진입 가능.

## [Unreleased] — v2.20.0-rc1 Governance: SPEC-V3R4-HARNESS-NAMESPACE-001 Harness Namespace + Lifecycle Governance closeout

### Changed

- **lint walker (drift.go): SPEC-ID word-boundary 매칭 정밀화 — substring collision 영구 차단** (SPEC-V3R4-LINT-SPECID-GREP-FIX-001): `internal/spec/drift.go` `getGitImpliedStatus` walker가 SPEC-ID를 git log에서 검색할 때 사용하는 `--grep=<SPEC-ID>` substring 매칭을 **단어 경계 (word-boundary) 정밀 매칭**으로 격상. 근본 원인: walker가 `SPEC-V3R4-HARNESS-001` 검색 시 `SPEC-V3R4-HARNESS-NAMESPACE-001` 같이 prefix가 겹치는 다른 SPEC commit (특히 NAMESPACE-001 supersede commit `ea1c10647`)을 walker first match로 채택하여 `V3R4-HARNESS-001/002/003` 3건의 false-positive `StatusGitConsistency` WARNING 유발. 방향 A (git POSIX BRE word-boundary) 대신 Approach B (2-pass post-filter) 채택 — 기존 `ExtractSPECIDs` 정규식 재사용, 외부 의존성 0. TDD 3-Wave (RED test `TestGetGitImpliedStatus_SPECIDWordBoundary` → GREEN implementation → REFACTOR + regression test). (PR #947 plan + #948 run + #이번 sync, plan-in-main + branch-only 정책 적용)

### Fixed

- **AC-LSGF-001: V3R4-HARNESS-001/002/003 3건 false-positive WARNING 영구 해소**: walker word-boundary 필터 main에 영구 적용 후 `moai spec lint --strict` 실행 시 이전 8 warning(LSGF-001 자신 1 + 기존 7) → 7 warning(새로 노출된 drift 7, LSGF-001 자신 0)로 해소. AC-LSGF-001 본질 달성 (HARNESS-001/002/003 false-positive 3 = 0).

### Known Follow-up

- Walker fix가 차단하던 substring noise가 가리던 **real status drift 7건 표면화** — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002 (예정)로 별도 처리. SDF-001 → FOLLOWUP-002 precedent 적용. v2.20.0-rc1 release timing에 영향 없음.
- CI shallow clone (`actions/checkout@v4` default `fetch-depth: 1`)으로 인한 `TestGetGitImpliedStatus_HARNESS001Resolution` CI skip 패턴 — SPEC-V3R4-CI-INFRA-FIX-001 (A+B Bundle)에서 `fetch-depth: 0` 영구 fix.

### Governance

- **SPEC-V3R4-HARNESS-NAMESPACE-001 — Harness namespace + lifecycle governance + PR #908 closeout** (plan PR #944 merged into main `ea1c10647`, run PR #945 merged into main `571258a1e`, sync PR #N this commit, plan-in-main + worktree-ban policy applied): Pure governance SPEC (zero LOC code change) formalizing the V3R4 harness namespace authority across the eight-SPEC family (SPEC-V3R4-HARNESS-{001..008}). Establishes (a) canonical SPEC ID format `^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$` covering SCOPE-before-HARNESS variants (e.g., `SPEC-V3R3-PROJECT-HARNESS-001`), (b) lifecycle independence guarantee preventing one harness SPEC's merge from blocking another (acyclic dependency graph, foundation-only edges), (c) verbatim re-assertion of the `/moai harness` 4-verb stable surface (`status`, `apply`, `rollback`, `disable`) — downstream SPECs MUST NOT extend without a new governance SPEC, (d) `.moai/harness/*` directory hierarchy with reserved file paths and **7-day rolling window retention policy** (REQ-HRN-NS-010 — `usage-log.jsonl` rotates daily, snapshots retained 7 days, frozen-guard-violations.jsonl append-only; purged entries archived to `.moai/research/evolution-log.md`), and (e) explicit closeout of PR #908 (`feat/cmd-harness-slash-wrapper`, OPEN since 2026-05-13) via user-selected **Option 1: rollback-tip-then-close** — branch reset to absorbed tip `452aa638f` (from divergent HEAD `a41d6d139c8c` carrying ad-hoc audit doc) + force-with-lease push + attribution comment + close-with-delete-branch. PR #910 commit `bb80ea0f4` had already absorbed the thin-wrapper code; audit doc commit `a41d6d139` was permanently discarded with attribution to V3R4-HARNESS-001 run-phase. **Lifecycle COMPLETE** (sync PR closes status `implemented → completed`); v2.20.0-rc1 governance authority for downstream SPEC-V3R4-HARNESS-{004..008} namespace compliance now permanently in force. Wave 1 (7 governance verification tasks T-Wave1-001~007) all PASS — SPEC ID regex 0 violations across 7 harness SPECs, dependency graph acyclic (4 V3R4-HARNESS-* lifecycle-independent), `.moai/harness/` canonical subset (main.md + README.md + usage-log.jsonl), `moai spec lint --strict` 0 errors (NAMESPACE-001 draft→implemented transition auto-resolves 4th drift warning post-commit; remaining 3 are pre-existing V3R4-HARNESS-001/002/003 walker-prefix drift, out-of-scope), CLI deprecation grace re-asserted (`internal/cli/harness.go` 433 lines preserved as deprecation stub + `internal/cli/root.go` no `harnessCmd` registration confirmed per BC-V3R4-HARNESS-001-CLI-RETIREMENT). Wave 2 (5 PR #908 closeout tasks T-Wave2-001~005) all PASS — EC-001 escalation path confirmed at HEAD `a41d6d139c8c` (prefix mismatch with `452aa638f*`); harness.md verb-surface diff verified exactly 4 verbs at lines 132/166/196/208 (no drift). Run PR composition: ~3 markdown files modified (CHANGELOG.md + 2 governance audit reports under `.moai/reports/governance/`) + spec.md frontmatter status update. Sync PR will close lifecycle (implemented → completed) with `sync(spec):` prefix per lesson #16 walker filter requirement. AC-HRN-NS-001~008 all binary PASS. Governance authority for downstream SPEC-V3R4-HARNESS-{004..008} namespace compliance now in force. Reference: `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave1-2026-05-16.md` + `-wave2-2026-05-16.md` + plan-auditor PASS 0.86.

### Governance (English)

- **SPEC-V3R4-HARNESS-NAMESPACE-001 — Harness namespace + lifecycle governance + PR #908 closeout** (plan PR #944 merged into main `ea1c10647`, run PR #945 merged into main `571258a1e`, sync PR #N this commit, plan-in-main + worktree-ban policy applied): Pure governance SPEC (zero LOC code change) formalizing the V3R4 harness namespace authority across the eight-SPEC family (SPEC-V3R4-HARNESS-{001..008}). Establishes (a) canonical SPEC ID format `^SPEC-V3R[0-9]+(-[A-Z][A-Z0-9]*)*-HARNESS(-[A-Z][A-Z0-9]*)?-[0-9]{3}$` covering SCOPE-before-HARNESS variants (e.g., `SPEC-V3R3-PROJECT-HARNESS-001`), (b) lifecycle independence guarantee preventing any single harness SPEC's merge from blocking sibling SPECs (acyclic dependency graph, foundation-only edges), (c) verbatim re-assertion of the `/moai harness` 4-verb stable surface (`status`, `apply`, `rollback`, `disable`) — downstream SPECs MUST NOT extend without a new governance SPEC, (d) `.moai/harness/*` directory hierarchy with reserved file paths and **7-day rolling window retention policy** (REQ-HRN-NS-010 — `usage-log.jsonl` rotates daily, snapshots retained 7 days, frozen-guard-violations.jsonl append-only; purged entries archived to `.moai/research/evolution-log.md`), and (e) explicit closeout of PR #908 via user-selected **Option 1: rollback-tip-then-close** — branch reset to absorbed tip `452aa638f` + force-with-lease push + attribution comment + close-with-delete-branch. PR #910 commit `bb80ea0f4` had already absorbed the thin-wrapper code; the divergent audit doc commit `a41d6d139` was permanently discarded with attribution. Wave 1: 7 governance verification tasks all PASS. Wave 2: 5 PR #908 closeout tasks all PASS. All 8 ACs binary-PASS. **Lifecycle COMPLETE** (sync PR closes status `implemented → completed`). Reference: `.moai/reports/governance/SPEC-V3R4-HARNESS-NAMESPACE-001-wave{1,2}-2026-05-16.md` + plan-auditor 0.86 PASS.

## [Unreleased] — SPEC-V3R4-SPECLINT-DEBT-002: plan workflow Pre-Write Frontmatter Checklist를 lint.go canonical (12-field)에 정렬 — dual-schema drift 영구 해소

### Fixed

- **SPEC-V3R4-SPECLINT-DEBT-002 — dual-schema drift 영구 해소** (PR #941 plan + #942 run + sync PR #N, plan-in-main + worktree 미사용 정책 적용): SDF-001 run-phase에서 발생한 hotfix commit `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강`의 근본 원인 (plan workflow body 9-field snake_case vs `internal/spec/lint.go` FrontmatterSchemaRule 12-field canonical 사이 dual-schema drift) 영구 해소. 방향 A 선택 — lint.go 12-field canonical 유지 + plan workflow body 정렬 (197 SPECs 중 51 `created_at:` dual-field + 53 `labels:` dual-field 모두 lint regression 0건, 마이그레이션 불요). TDD 3-Wave 산출: Wave 1 plan.md Pre-Write Frontmatter Checklist 9→12 fields + Rejected aliases 반전 (snake_case → reject) / Wave 2 `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT 신설 (3-layer cross-reference: plan.md + team/plan.md + lint.go) / Wave 3 `.claude/skills/moai/team/plan.md` Case-C 신설 (~75 LOC) + regression fixture (`internal/spec/testdata/frontmatter-schema/{valid-12-field,invalid-snake-case-only}/spec.md`) + Go test 2건 (`TestFrontmatterSchemaRule_Valid12Field` + `TestFrontmatterSchemaRule_SnakeCaseRejected` PASS exactly 3 FrontmatterInvalid). plan-auditor 2회 PASS (iter 1 FAIL 0.62 → iter 2 PASS 0.94, MP 4/4, REQ↔AC bijection 1:1). 11 files / +523/-26 LOC. `moai spec lint --strict` 0 findings + `go test ./...` PASS. **AC-SDBT-002-001~005 모두 verified.** 향후 manager-spec이 plan workflow를 가이드로 새 SPEC 작성 시 snake_case duplicates 추가 중단 + SDF-001 hotfix 시나리오 재발 방지.

### Fixed (English)

- **SPEC-V3R4-SPECLINT-DEBT-002 — Permanent resolution of dual-schema drift** (PR #941 plan + #942 run + sync PR #N, plan-in-main + worktree-ban policy applied): Permanent resolution of the root cause behind SDF-001 run-phase hotfix commit `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강` — dual-schema drift between plan workflow body (9-field snake_case `created_at`/`updated_at`/`labels`/`issue_number`) and `internal/spec/lint.go` FrontmatterSchemaRule (12-field canonical `created`/`updated`/`tags`/`title`/`phase`/`module`/`lifecycle`). Direction A selected — preserve lint.go 12-field canonical + realign plan workflow body (197 SPECs surveyed: 51 with `created_at:` dual-field + 53 with `labels:` dual-field cause 0 lint regressions, no migration required). TDD 3-Wave delivery: Wave 1 plan.md Pre-Write Frontmatter Checklist 9→12 fields + Rejected aliases reversed (snake_case → reject) / Wave 2 new `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT (3-layer cross-reference: plan.md + team/plan.md + lint.go) / Wave 3 new `.claude/skills/moai/team/plan.md` Case-C section (~75 LOC) + regression fixture (`internal/spec/testdata/frontmatter-schema/{valid-12-field,invalid-snake-case-only}/spec.md`) + 2 Go tests (`TestFrontmatterSchemaRule_Valid12Field` + `TestFrontmatterSchemaRule_SnakeCaseRejected` PASS exactly 3 FrontmatterInvalid). plan-auditor: 2 iterations PASS (iter 1 FAIL 0.62 → iter 2 PASS 0.94, all must-pass criteria met, REQ↔AC bijection 1:1). 11 files / +523/-26 LOC. `moai spec lint --strict` 0 findings + `go test ./...` PASS. **AC-SDBT-002-001~005 all verified.** Going forward, manager-spec will no longer add snake_case duplicates when generating new SPEC frontmatter via the plan workflow guide + SDF-001 hotfix scenario non-reproducible.

## [Unreleased] — SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001: 64건 StatusGitConsistency drift 일괄 동기화 + Pattern H 잔여 closeout

### Fixed

- **SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — 64건 StatusGitConsistency drift 일괄 해소** (PR #939 통합, plan-in-main + worktree 미사용 정책 적용): `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (PR #937, main `758341089`) 머지로 `lint.skip` 회피책이 일제히 제거된 직후, 그 mask 아래 잠재해 있던 real status drift 64건이 `moai spec lint --strict --strict-warnings` 에 노출됨. 6-wave 처리(Wave 2 Pattern A 47건 `implemented → planned` downgrade + Wave 3 Pattern B 4건 + Pattern C 6건 + Wave 4 추가 20건 → 총 57건 + 자동 검출 7건 = 64건). Wave 4에서 `internal/spec/lint.go` 에 `terminalStatusEnum` 도입 (TDD M1-M5, `internal/spec/status_terminal_exemption_test.go` +149 LOC) — terminal status (`completed`/`cancelled`/`deprecated`)는 git-implied 추론을 우회하도록 lint rule 보강. Wave 7 sync-phase: Pattern H 잔여 4건 (LSC/LSCSK/SPECLINT-DEBT/SDF-001 자신) frontmatter `status: completed` + `updated: 2026-05-16` 통일하여 `sync(spec):` prefix commit으로 walker filter 우회 → git-implied 'planned' → 'completed' 끌어올림. **AC-SDF-001 binary 0 달성** (`moai spec lint --strict --strict-warnings` 0 ERROR + 0 WARNING). SDF-001 자체 frontmatter status `in-progress → completed`, version `0.1.1 → 0.1.2`.
- **SPEC `status` 컨벤션 추가 정착 (`implemented` → `completed`)**: SPECLINT-DEBT-001 (PR #917)에서 도입한 `implemented` (code merged, sync incomplete) vs `completed` (full plan/run/sync lifecycle closed) 의미 차이를 SDF-001 Pattern H에서 일관 적용. `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (`implemented` → `completed`, version `0.1.2 → 0.1.3`), `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (terminal 유지, version `0.3.0 → 0.3.1` metadata sync), `SPEC-V3R4-SPECLINT-DEBT-001` (terminal 유지, version `0.2.0 → 0.2.1` metadata sync).

### Fixed (English)

- **SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 — Batch resolution of 64 StatusGitConsistency drift warnings** (PR #939 integrated, plan-in-main + worktree-ban policy applied): Immediately after `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (PR #937, main `758341089`) merged and removed all `lint.skip` escape hatches, 64 latent real status drifts were exposed by `moai spec lint --strict --strict-warnings`. Processed in 6 waves (Wave 2 Pattern A 47 `implemented → planned` downgrades + Wave 3 Pattern B 4 + Pattern C 6 + Wave 4 additional 20 = 57 batched + 7 auto-detected = 64 total). Wave 4 introduces `terminalStatusEnum` in `internal/spec/lint.go` (TDD M1-M5, `internal/spec/status_terminal_exemption_test.go` +149 LOC) — terminal statuses (`completed`/`cancelled`/`deprecated`) now bypass git-implied inference. Wave 7 sync-phase: Pattern H residual 4 SPECs (LSC/LSCSK/SPECLINT-DEBT/SDF-001 self) frontmatter unified to `status: completed` + `updated: 2026-05-16`, committed under `sync(spec):` prefix (bypassing walker filter that skips `chore(spec):`) to lift git-implied from 'planned' to 'completed'. **AC-SDF-001 binary 0 achieved** (`moai spec lint --strict --strict-warnings` 0 ERROR + 0 WARNING). SDF-001 self frontmatter status `in-progress → completed`, version `0.1.1 → 0.1.2`.
- **Additional convention reinforcement (`implemented` → `completed`)**: The `implemented` (code merged, sync incomplete) vs `completed` (full plan/run/sync lifecycle closed) distinction codified in SPECLINT-DEBT-001 (PR #917) is now consistently applied via SDF-001 Pattern H. Bumps: `SPEC-V3R4-LINT-SKIP-CLEANUP-001` (`implemented` → `completed`, version `0.1.2 → 0.1.3`), `SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001` (terminal preserved, version `0.3.0 → 0.3.1` metadata sync), `SPEC-V3R4-SPECLINT-DEBT-001` (terminal preserved, version `0.2.0 → 0.2.1` metadata sync).

### Discovered (Follow-up SPEC Candidates)

- **SPECLINT-DEBT-002 후보 (dual-schema drift)**: SDF-001 plan workflow가 사용하는 SPEC frontmatter 7-field canonical (id, version, status, created, updated, author, priority + lifecycle + breaking + bc_id) vs `internal/spec/lint.go` `FrontmatterInvalidRule`의 12-field canonical 사이 schema 불일치 발견. PR #939 CI에서 plan-phase frontmatter가 lint 통과하지 못해 `b2b7f32c7 fix(spec): SDF-001 spec.md frontmatter canonical 7-field 보강` 별도 hotfix 발생. plan workflow / docs / lint rule 3계층의 SSOT 정리가 필요.

## [Unreleased] — SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001: spec-lint bootstrapping bug 해소

### Fixed

- **spec-lint: chore(spec) sweep commit이 StatusGitConsistency WARNING을 유발하던 bootstrapping bug 해소** (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001): `internal/spec/drift.go::getGitImpliedStatus` 에 walker filter 도입. `git log -1` (단일 commit) → `git log -50` (N=50 walker) 변경, `shouldSkipCommitTitle` helper가 `chore(spec):` / `chore(specs):` sweep commit을 건너뛰어 이전 의미 있는 impl/feat/sync commit의 status를 채택. `moai spec lint --strict` 의 7건 StatusGitConsistency WARNING (SPEC-UTIL-001, SPEC-V3R2-CON-001/002/003, SPEC-V3R2-RT-001, SPEC-V3R2-SPC-003, SPEC-V3R4-HARNESS-003) 이 0건으로 해소됨.

### Fixed (English)

- **spec-lint: resolved bootstrapping bug where chore(spec) sweep commits caused StatusGitConsistency WARNINGs** (SPEC-V3R4-LINT-STATUS-CHORE-SKIP-001): Introduced walker filter in `internal/spec/drift.go::getGitImpliedStatus`. Changed `git log -1` (single commit) to `git log -50` (N=50 walker) with new `shouldSkipCommitTitle` helper that skips `chore(spec):`/`chore(specs):` sweep commits and walks back to find the real `impl`/`feat`/`sync` commit. Resolves 7 StatusGitConsistency WARNINGs: SPEC-UTIL-001, SPEC-V3R2-CON-001/002/003, SPEC-V3R2-RT-001, SPEC-V3R2-SPC-003, SPEC-V3R4-HARNESS-003.

## [Unreleased] — Sync: WF-001 + PATTERNS-001 status drift resolution (2026-05-15)

### Changed

- **SPEC-V3R2-WF-001 status: `in-progress` → `completed`**: Implementation Notes(2026-04-25)에 Stage 1 consolidation 48 → 38 skill directories 달성이 명시되어 있으나 frontmatter `status:` 가 동기화되지 않은 drift를 해소. plan-audit (iter 1, 2026-05-15) PASS @ 0.92. spec.md frontmatter HISTORY v1.2.0 entry 추가, 본문/코드 비변경.
- **SPEC-V3R3-PATTERNS-001 status: `in-progress` → `completed`**: revfactory/harness (Apache 2.0) 6 reference docs가 이미 `.claude/rules/moai/{development,quality,workflow}/` 6 rule 파일 + `.claude/rules/moai/NOTICE.md` 로 흡수되었고, `internal/template/templates/` mirror 6/6 byte-identical 확인. plan-audit verdict: closure via metadata sync only. spec.md frontmatter HISTORY v0.2.0 entry 추가, 본문/코드 비변경.

### Changed (English)

- **SPEC-V3R2-WF-001 status: `in-progress` → `completed`**: Implementation Notes (2026-04-25) record full completion of Stage 1 consolidation (48 → 38 skill directories), but frontmatter `status:` was not synced — drift resolved. plan-audit (iter 1, 2026-05-15) PASS @ 0.92. spec.md frontmatter + HISTORY v1.2.0 entry only; body/code unchanged.
- **SPEC-V3R3-PATTERNS-001 status: `in-progress` → `completed`**: 6 reference docs from revfactory/harness (Apache 2.0) already absorbed into `.claude/rules/moai/{development,quality,workflow}/` (6 rule files + `.claude/rules/moai/NOTICE.md`), and `internal/template/templates/` mirror is 6/6 byte-identical. plan-audit verdict: closure via metadata sync only. spec.md frontmatter + HISTORY v0.2.0 entry only; body/code unchanged.

## [Unreleased] — SPEC-V3R4-SPECLINT-DEBT-001: SPEC Lint Debt 일괄 해소

### Fixed

- **SPEC-V3R4-SPECLINT-DEBT-001 — SPEC Lint Debt 일괄 해소 (P0 ERROR 66건 + P1 WARNING 141건 → 0/0)**: V3R4 foundation cleanup. `moai spec lint --strict`가 main `2e27c14f8` 기준으로 보고하던 6개 ERROR 카테고리(FrontmatterInvalid 13, CoverageIncomplete 44, ParseFailure 4, MissingDependency 2, ModalityMalformed 1, MissingExclusions 1, DependencyCycle 1)와 2개 WARNING 카테고리(StatusGitConsistency 141, OrphanBCID 1)를 단일 PR(#917)로 일괄 해소. spec-lint CI workflow가 GREEN으로 전환되어 다른 SPEC들의 정상 머지를 차단하던 risk 제거. SPEC 본문은 비-수정 보장(REQ-SLD-010 self-coverage 적용); 메타데이터/frontmatter/AC reference만 수정. 5 categorical commits 채택 (frontmatter+ID+ParseFailure / deps+cycle / modality+excl / coverage / status sync+bc_id). 90 SPEC frontmatter `status:` 정정 + 51 SPEC `lint.skip`(author-intent preservation, terminal state preservation) + 1 ARCH-007 `bc_id: []`. Wave 1 (T-SLD-001~006) + Wave 2 (T-SLD-007/008) + Wave 3 (T-SLD-009~011) 5-wave 산출. lint-final.md ERROR 66→0 / WARNING 141→0 PASS. plan-auditor 2회 PASS (iter 1 0.92, iter 2 0.88). PR #917 admin squash merged → main `0497f62104`. 다음 V3R4 SPEC들이 lint-baseline-zero 상태에서 진입 가능.

### Fixed (English)

- **SPEC-V3R4-SPECLINT-DEBT-001 — SPEC Lint Debt Batch Cleanup (P0 66 ERRORs + P1 141 WARNINGs → 0/0)**: V3R4 foundation cleanup. Resolves the 6 ERROR categories (FrontmatterInvalid 13, CoverageIncomplete 44, ParseFailure 4, MissingDependency 2, ModalityMalformed 1, MissingExclusions 1, DependencyCycle 1) and 2 WARNING categories (StatusGitConsistency 141, OrphanBCID 1) that `moai spec lint --strict` reported against main `2e27c14f8`. Single PR (#917) batch fix flips spec-lint CI to GREEN, unblocking downstream V3R4 SPEC merges. SPEC body is preservation-guaranteed (REQ-SLD-010 self-coverage); only metadata / frontmatter / AC references modified. Adopted 5 categorical commits (frontmatter+ID+ParseFailure / deps+cycle / modality+excl / coverage / status sync+bc_id). 90 SPEC `status:` corrections + 51 SPEC `lint.skip` directives (author-intent + terminal state preservation) + 1 ARCH-007 `bc_id: []`. Wave 1 (T-SLD-001~006) + Wave 2 (T-SLD-007/008) + Wave 3 (T-SLD-009~011) delivered in 5 waves. lint-final.md ERROR 66→0 / WARNING 141→0 PASS. plan-auditor PASS twice (iter 1 0.92, iter 2 0.88). PR #917 admin squash merged → main `0497f62104`. Future V3R4 SPECs enter on a lint-baseline-zero state.

### Changed

- **SPEC `status` field convention 정착**: `implemented` (code merged, sync incomplete) vs `completed` (full plan/run/sync lifecycle closed) 의미가 lint rule 수준에서 명확화. 본 정착은 자동화 도구(`scripts/spec-status-sync.go` 신규)로 백필되었으며, 향후 SPEC 라이프사이클 종료 시 sync-phase가 일관되게 `completed`로 전환하도록 유도.
- **`lint.skip` mechanism**: SPEC frontmatter `lint:\n  skip:\n    - <category>` 구문이 정당화된 잔존 violation을 카테고리별로 suppress하는 escape-hatch로 공식 도입. `--strict` 모드에서도 lint.skip 처리된 항목은 exit 0을 막지 않음. `internal/spec/lint.go::applylintSkip` 으로 구현.

### Changed (English)

- **SPEC `status` field convention codified**: distinction between `implemented` (code merged but sync incomplete) and `completed` (full plan/run/sync lifecycle closed) is now lint-enforced. Backfilled via a new `scripts/spec-status-sync.go` automation tool. Going forward, sync-phase consistently transitions SPECs to `completed`.
- **`lint.skip` escape-hatch mechanism**: SPEC frontmatter `lint:\n  skip:\n    - <category>` is now an official escape-hatch for justified residual violations on a per-category basis. `--strict` mode honors `lint.skip` and does not block exit 0 for skipped categories. Implemented in `internal/spec/lint.go::applylintSkip`.

## [Unreleased] — SPEC-V3R4-HARNESS-001: Self-Evolving Harness v2 Foundation

### docs (문서)

- Claude Code v2.1.140-142 변경사항 분석 보고서 추가 (`.moai/research/cc-update-20260515.md`) — 41건 분석 중 Tier 1 5건, Tier 2 6건 actionable. 후속 SPEC 후보 2건 식별 (SPEC-V3R4-CC2X-ADOPT-002 `claude agents` flag pass-through, SPEC-V3R4-CC2X-ADOPT-003 hook `terminalSequence` 채택).
- Added Claude Code v2.1.140-142 upstream change report (`.moai/research/cc-update-20260515.md`) — 41 items analyzed, 5 Tier-1 + 6 Tier-2 actionable. Identified 2 follow-up SPEC candidates: SPEC-V3R4-CC2X-ADOPT-002 (`claude agents` flag pass-through) and SPEC-V3R4-CC2X-ADOPT-003 (hook `terminalSequence` adoption).

### Breaking Changes

- **BC-V3R4-HARNESS-001-CLI-RETIREMENT** — `moai harness <verb>` CLI subcommand 경로가 폐기되었습니다. 셸에서 `moai harness status` (또는 `apply`, `rollback`, `disable`)를 호출하면 cobra의 `unknown command "harness" for "moai"` 진단과 함께 non-zero exit code가 반환됩니다. 동일한 기능은 Claude Code 세션 내의 `/moai:harness` 슬래시 커맨드로만 사용할 수 있습니다. 슬래시 커맨드 표면은 V3R3 시절과 동일하게 유지되므로 사용자 머슬 메모리에는 영향이 없습니다.
- 신규 회귀 가드: `internal/cli/harness_retirement_test.go` `TestHarnessRetirement`가 `rootCmd.Commands()`에 `harness` subcommand가 다시 등록되는 PR을 자동으로 차단합니다.

### Breaking Changes (English)

- **BC-V3R4-HARNESS-001-CLI-RETIREMENT** — The `moai harness <verb>` CLI subcommand path is retired. Invoking `moai harness status` (or `apply`, `rollback`, `disable`) from the shell returns cobra's `unknown command "harness" for "moai"` diagnostic with a non-zero exit code. The same functionality is reachable only through the `/moai:harness` slash command inside a Claude Code session. The slash command surface is identical to the V3R3 era, so user muscle memory is unaffected.
- New regression guard: `internal/cli/harness_retirement_test.go` `TestHarnessRetirement` fails the build if any PR re-registers a `harness` subcommand on `rootCmd.Commands()`.

### Added

- **SPEC-V3R4-HARNESS-001 — Unified Self-Evolving Harness Foundation**: V3R4 self-evolving harness 아키텍처의 foundation SPEC. 세 개의 V3R3 SPEC을 단일 V3R4 family로 통합 (`supersedes:` frontmatter): `SPEC-V3R3-HARNESS-001` (meta-skill), `SPEC-V3R3-HARNESS-LEARNING-001` (4-tier learning ladder + 5-Layer Safety), `SPEC-V3R3-PROJECT-HARNESS-001` (16Q 인터뷰 + 통합 wiring). harness 라이프사이클은 슬래시 커맨드 + 스킬 워크플로우 + Claude Code hook 만으로 동작하며 Go 바이너리 호출이 0건입니다. 5-Layer Safety 아키텍처 (`.claude/rules/moai/design/constitution.md` §5) 와 FROZEN zone (§2)은 비트 단위로 보존됩니다 (REQ-HRN-FND-005).
- **`/moai:harness` 슬래시 커맨드 lifecycle (V3R4 contract)**: `.claude/skills/moai/workflows/harness.md` 본문이 `status` / `apply` / `rollback` / `disable` 4개 verb를 모두 file-system 연산으로 구현합니다. Tier-4 적용 시 orchestrator-issued AskUserQuestion 4-option 패턴 (Apply (권장) / Modify / Defer / Reject) 게이트 (REQ-HRN-FND-004, REQ-HRN-FND-015). Tier-4 적용 빈도는 프로젝트당 7일 rolling window 1회 (REQ-HRN-FND-012, 하향 불가 REQ-HRN-FND-018).
- **PostToolUse observer no-op gate (REQ-HRN-FND-009)**: `internal/cli/hook.go` `isHarnessLearningEnabled` 함수가 `.moai/config/sections/harness.yaml`의 `learning.enabled` 필드를 읽어 `false`이면 observer가 완전한 no-op으로 동작합니다. 누락된 config / 파싱 오류 시 fail-open (관측 유지). `internal/cli/hook_harness_observe_test.go` 10건의 table-driven 테스트로 검증.
- **CI 회귀 가드 (REQ-HRN-FND-002)**: `internal/cli/harness_retirement_test.go`가 `rootCmd.Commands()`에 `harness` subcommand 재등록 시도를 차단.
- **moai-harness-learner / moai-meta-harness V3R4 텍스트 주석**: SPEC §10 exclusion #10에 따라 두 skill body는 frontmatter / 동작 변경 없이 V3R4 contract 인용 주석만 추가.

### Added (English)

- **SPEC-V3R4-HARNESS-001 — Unified Self-Evolving Harness Foundation**: Foundation SPEC for the V3R4 self-evolving harness architecture. Consolidates three V3R3 SPECs into a single V3R4 family via `supersedes:` frontmatter: `SPEC-V3R3-HARNESS-001` (meta-skill), `SPEC-V3R3-HARNESS-LEARNING-001` (4-tier learning ladder + 5-Layer Safety), `SPEC-V3R3-PROJECT-HARNESS-001` (16Q interview + integration wiring). The harness lifecycle operates entirely through slash command + skill workflow + Claude Code hooks; zero Go binary invocations remain. The 5-Layer Safety architecture (`.claude/rules/moai/design/constitution.md` §5) and FROZEN zones (§2) are preserved byte-for-byte (REQ-HRN-FND-005).
- **`/moai:harness` slash command lifecycle (V3R4 contract)**: `.claude/skills/moai/workflows/harness.md` body implements all four verbs (`status` / `apply` / `rollback` / `disable`) via file-system operations only. Tier-4 application is gated by an orchestrator-issued AskUserQuestion four-option pattern (Apply (Recommended) / Modify / Defer / Reject) per REQ-HRN-FND-004 and REQ-HRN-FND-015. Tier-4 application is rate-limited to one per project per 7-day rolling window (REQ-HRN-FND-012) with a floor that cannot be lowered by adaptive expansion (REQ-HRN-FND-018).
- **PostToolUse observer no-op gate (REQ-HRN-FND-009)**: `internal/cli/hook.go` `isHarnessLearningEnabled` reads `learning.enabled` from `.moai/config/sections/harness.yaml`; when `false`, the observer becomes a complete no-op (no read, write, or append). Fail-open semantics: missing config or parse error preserves baseline observation. Verified by 10 table-driven cases in `internal/cli/hook_harness_observe_test.go`.
- **CI regression guard (REQ-HRN-FND-002)**: `internal/cli/harness_retirement_test.go` fails the build if any PR re-registers a `harness` subcommand on `rootCmd.Commands()`.
- **moai-harness-learner / moai-meta-harness V3R4 text annotations**: per SPEC §10 exclusion #10, both skill bodies receive text-only annotations reaffirming the V3R4 contract with no frontmatter or behavioral changes.

### Superseded

- `SPEC-V3R3-HARNESS-001` — Meta-Harness Skill core (status transition to `superseded` is performed via the follow-up `manager-git` commit per `.moai/specs/SPEC-V3R4-HARNESS-001/follow-up.md`).
- `SPEC-V3R3-HARNESS-LEARNING-001` — 4-tier learning ladder + 5-Layer Safety + `moai harness` CLI verbs (CLI verb path retired; 4-tier ladder and 5-Layer Safety preserved unchanged).
- `SPEC-V3R3-PROJECT-HARNESS-001` — `/moai project` Phase 5+ socratic interview + 5-Layer integration wiring (runtime behavior preserved; V3R4 SPEC formalizes the contract those layers operate under).

### Downstream (not in scope of this SPEC)

이 foundation SPEC은 자체적으로 self-evolution 메커니즘을 도입하지 않습니다. 다음 7개 downstream SPEC은 이 foundation을 점진적으로 확장합니다 (모두 `.moai/specs/SPEC-V3R4-HARNESS-001/spec.md` §1.3 Non-Goals 에 명시):

- `SPEC-V3R4-HARNESS-002` — Multi-event observer (Stop / SubagentStop / UserPromptSubmit 통합) **[IMPLEMENTED]**
- `SPEC-V3R4-HARNESS-003` — Embedding-cluster pattern detection (frequency-count classifier 대체)
- `SPEC-V3R4-HARNESS-004` — Reflexion 자체-비판 loop (3-iteration cap)
- `SPEC-V3R4-HARNESS-005` — Constitution principle-based scoring
- `SPEC-V3R4-HARNESS-006` — Multi-objective effectiveness measurement + auto-rollback-on-regression
- `SPEC-V3R4-HARNESS-007` — Voyager 스킬 라이브러리 자동 organization (embedding-indexed retrieval)
- `SPEC-V3R4-HARNESS-008` — Cross-project lesson federation (privacy-sensitive, opt-in only)

## [Unreleased] — SPEC-V3R4-HARNESS-002: Multi-Event Observer Expansion

### Added

- **SPEC-V3R4-HARNESS-002 — Multi-Event Observer Expansion (Stop / SubagentStop / UserPromptSubmit)**: V3R4 self-evolving harness의 관찰 표면을 PostToolUse-only baseline에서 4-event 매트릭스(PostToolUse + Stop + SubagentStop + UserPromptSubmit)로 확장합니다. 3개 신규 cobra subcommand (`moai hook harness-observe-{stop, subagent-stop, user-prompt-submit}`) + `EventType` enum 확장 (`session_stop` / `subagent_stop` / `user_prompt`) + `.moai/harness/usage-log.jsonl` schema 확장(additive optional fields, `omitempty`). PII 보안 기본값: Strategy A (SHA-256 hash + length + language heuristic for UserPromptSubmit). `learning.enabled` gate 통합(모든 4개 event hook 공유) → 일괄 no-op 가능. 5-Layer Safety (`.claude/rules/moai/design/constitution.md` §5)와 4-tier ladder (REQ-HRN-FND-011) 보존. 3 wrapper script templates (shell .sh.tmpl) + settings.json.tmpl additive wiring (Strategy WIRE-A). Evaluator iter 1/2 defect fix 포함 (PromptPreview byte boundary + prompt_hash [:16] truncation + prompt_full→prompt_content 필드명 정정). 13/13 AC PASS, coverage 87.9%, MX P1/P2 violations 0. 5-wave delivery (A/A.5/B/C) + Phase 2.75 lint + Phase 2.8a evaluator 3 iterations. PR #914 (#909 plan +  #910/#911 run baseline으로부터). downstream SPEC-V3R4-HARNESS-003 (embedding-cluster classifier) 진입 자격 충족.

### Added (English)

- **SPEC-V3R4-HARNESS-002 — Multi-Event Observer Expansion**: Extends the V3R4 self-evolving harness observation surface from a PostToolUse-only baseline to a 4-event matrix (PostToolUse + Stop + SubagentStop + UserPromptSubmit). Introduces 3 new cobra subcommands (`moai hook harness-observe-{stop, subagent-stop, user-prompt-submit}`) + `EventType` enum extension (`session_stop` / `subagent_stop` / `user_prompt`) + `.moai/harness/usage-log.jsonl` schema expansion (additive optional fields tagged with `omitempty`). PII security default: Strategy A (SHA-256 hash + length + language heuristic for UserPromptSubmit). Unified `learning.enabled` gate across all 4 event hooks enables bulk no-op capability. Preserves 5-Layer Safety (`.claude/rules/moai/design/constitution.md` §5) and 4-tier ladder (REQ-HRN-FND-011). 3 wrapper script templates (shell .sh.tmpl) + settings.json.tmpl additive registration (Strategy WIRE-A). Includes evaluator iter 1/2 defect fixes (PromptPreview byte boundary + prompt_hash [:16] truncation + prompt_full→prompt_content field naming). 13/13 AC PASS, coverage 87.9%, MX P1/P2 violations 0. 5-wave delivery (A/A.5/B/C) + Phase 2.75 lint + Phase 2.8a evaluator 3 iterations. PR #914 (plan #909 + run #910/#911 baseline). Unblocks downstream SPEC-V3R4-HARNESS-003 (embedding-cluster classifier).

## [Unreleased] — SPEC-V3R2-HRN-003: Hierarchical Acceptance Scoring

### Added

- **SPEC-V3R2-HRN-003 — 계층 평가 채점 (Hierarchical Acceptance Scoring)**: sync-auditor의 평면 ScoreCard를 4-dimension × hierarchical sub-criterion 구조로 확장 (#885 / #887 / #889). 4 dimensions (Functionality / Security / Craft / Consistency) FROZEN, 4 rubric anchor levels {0.25, 0.50, 0.75, 1.00} FROZEN, must-pass firewall (Security 필수), 4 evaluator profiles (.md 포맷) loader, sync-auditor 본문 augment (3 augmentation), zone-registry CONST-V3R2-154/155 추가 등록. R1 §9 Agent-as-a-Judge + pattern-library E-1/E-3 ADOPT. (87 commits across 3 PRs, +1933 LOC, coverage 89.5%, MX 10 tags)

### Added (English)

- **SPEC-V3R2-HRN-003 — Hierarchical Acceptance Scoring**: Extended sync-auditor's flat ScoreCard to 4-dimension × hierarchical sub-criterion structure (#885 / #887 / #889). 4 canonical dimensions (Functionality / Security / Craft / Consistency) FROZEN, 4 rubric anchor levels {0.25, 0.50, 0.75, 1.00} FROZEN, must-pass firewall (Security mandatory), 4 evaluator profiles (.md format) loader, sync-auditor body augmentation (3 augmentations), zone-registry CONST-V3R2-154/155 additions. R1 §9 Agent-as-a-Judge + pattern-library E-1/E-3 ADOPTED. (87 commits across 3 PRs, +1933 LOC, coverage 89.5%, MX 10 tags)

## [Unreleased] — SPEC-V3R2-SPC-004: @MX Query Engine — Fan-in + SPEC Association + 16-Language Sweep

### Added

- **SPEC-V3R2-SPC-004**: `internal/mx` 패키지에 `Resolver.Resolve(Query)` API 구현. LSP `find-references` 통합(powernap)을 통해 `LSPFanInCounter`가 정확한 호출자 수를 계산하고, LSP 미사용 시 `TextualFanInCounter`로 graceful fallback.
- **SPEC-V3R2-SPC-004**: `mx.yaml` `danger_categories:` + `test_paths:` 사용자 설정 wire-up. `LoadDangerConfig(projectRoot)` → `DangerCategoryMatcher` → CLI `--danger` 플래그와 연결.
- **SPEC-V3R2-SPC-004**: `.moai/specs/*/spec.md` `module:` frontmatter 로더(`LoadSpecModules`) + 경로 기반 `SpecAssociator`. SPEC ID → 모듈 경로 맵으로 `@MX` 태그를 해당 SPEC에 자동 연결.
- **SPEC-V3R2-SPC-004**: `Resolver.ResolveAnchorCallsites()` additive API. 위치 정보(파일+라인+컬럼+method)를 포함한 `[]Callsite` 반환. LSP 경로와 textual fallback 모두 지원.

### English

- **SPEC-V3R2-SPC-004**: Implemented `Resolver.Resolve(Query)` API in `internal/mx` package. `LSPFanInCounter` computes precise caller counts via LSP `find-references` (powernap integration), falling back to `TextualFanInCounter` when LSP is unavailable.
- **SPEC-V3R2-SPC-004**: Wired `mx.yaml` `danger_categories:` + `test_paths:` user configuration. `LoadDangerConfig(projectRoot)` → `DangerCategoryMatcher` → CLI `--danger` flag.
- **SPEC-V3R2-SPC-004**: `.moai/specs/*/spec.md` `module:` frontmatter loader (`LoadSpecModules`) + path-based `SpecAssociator`. Automatically associates `@MX` tags with their governing SPEC via module path prefix matching.
- **SPEC-V3R2-SPC-004**: Additive `Resolver.ResolveAnchorCallsites()` API returning `[]Callsite` with file/line/column/method. Supports both LSP and textual fallback paths.

## [Unreleased] — SPEC-V3R2-RT-004: Typed Session State + Phase Checkpoint

### Added

- **SPEC-V3R2-RT-004**: 타입이 보장된 세션 상태 관리 시스템 구현. `PhaseState` + `Checkpoint` 인터페이스로 plan/run/sync phase별 상태를 `.moai/state/`에 원자적으로 저장. validator/v10 스키마 검증, cross-platform advisory lock(Unix flock + Windows LockFileEx), blocker 파일 스캔, staleness 검사(`stale_seconds` 설정), in-flight transition 감지, team-mode 체크포인트 병합(bubble-mode). `moai state dump/show-blocker` CLI 서브커맨드, cache-prefix 불변 조건(`HydrateForPrompt`), `retention_days` 기반 artifact 정리, AskUserQuestion 감사 lint. 7 MX 태그(ANCHOR 3, NOTE 2, WARN 2) 적용. AC-01~15 충족.

### English

- **SPEC-V3R2-RT-004**: Typed session state management system. `PhaseState` + `Checkpoint` interface atomically persists plan/run/sync phase state to `.moai/state/`. Features: validator/v10 schema checks, cross-platform advisory locks (Unix flock + Windows LockFileEx), blocker file scanning, staleness TTL (`stale_seconds` config), in-flight transition detection, team-mode checkpoint merge with bubble-mode. Added `moai state dump/show-blocker` CLI subcommands, cache-prefix invariant (`HydrateForPrompt`), `retention_days`-based artifact cleanup, and AskUserQuestion audit lint. 7 MX tags (ANCHOR 3, NOTE 2, WARN 2). AC-01~15 met.

## [Unreleased] — SPEC-V3R4-CATALOG-002: Wave 2 Distribution — Slim init via catalog tier filter

### Changed (BREAKING CHANGE)

- **`moai init` 기본 동작 변경**: catalog manifest 의 `tier == core` 자산만 배포 (20 core skills + 20 core agents + non-catalog 템플릿). Optional packs 9 종 (backend / frontend / mobile / auth / deployment / design / devops / testing / chrome-extension — 합계 17 skills + 7 agents) 및 builder-harness agent 1 개는 기본 배포에서 제외 → 약 38% (25/65 entries) 슬림. 두 가지 opt-out 경로: `moai init --all` flag 또는 `MOAI_DISTRIBUTE_ALL=1` 환경변수 (case-insensitive `"true"` 도 허용). `moai update` 동작은 영향 없음 (full FS 유지). Optional pack 인터랙티브 설치는 SPEC-V3R4-CATALOG-003 (`moai pack add`) 에서 제공 예정. 기존 프로젝트의 update drift sync 는 SPEC-V3R4-CATALOG-004. Builder-harness 자동 부트스트랩은 SPEC-V3R4-CATALOG-005.

### Added

- **SlimFS wrapper** (`internal/template/slim_fs.go`, +235 LOC): `fs.FS` 레벨 tier 필터. `SlimFS(rawFS fs.FS, cat *Catalog) (fs.FS, error)` API. `deployer.go` / `update.go` 미수정 (D7 lock 보존, REQ-005/006). 65 entries 중 25 non-core 자산이 `fs.ErrNotExist` 반환. `testing/fstest` 호환 (REQ-001/002/003/010/011/014/015). Coverage 91.1%.
- **encapsulated slim deployer constructor** (`internal/template/embed_catalog.go`, +59 LOC): `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer(cat, renderer)` 두 export. `embeddedRaw` unexported 유지 (DEFECT-5 encapsulation invariant — `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` zero matches).
- **builder-harness 친절 에러 가드** (`internal/template/slim_guard.go`, +36 LOC): `AssertBuilderHarnessAvailable(projectFS) error` — slim mode 에서 builder-harness 부재 시 `CATALOG_SLIM_HARNESS_MISSING` sentinel + `MOAI_DISTRIBUTE_ALL=1` + `moai init --all` + `SPEC-V3R4-CATALOG-005` 4 substring 안내 (REQ-021). Coverage 100%.
- **audit suite** (`internal/template/catalog_slim_audit_test.go`, +242 LOC): 6 sub-tests — `TestSlimFS_HidesNonCoreEntries` (REQ-014 `CATALOG_SLIM_LEAK`, 25 non-core 검증), `TestSlimFS_PreservesCoreEntries` + EC4 nested (REQ-015 `CATALOG_SLIM_CORE_MISSING`, 40 core + `.claude/skills/moai/workflows/plan.md` nested), `TestSlimFS_PreservesNonCatalogFiles` (REQ-016 `CATALOG_SLIM_OVER_FILTER`, 5 non-catalog paths), `TestSlimFS_WalkDirNoLeak` (REQ-017 `CATALOG_SLIM_WALK_LEAK`, 523 paths visited zero leaks), `TestSlimFS_ReadOnlyInvariant` (REQ-003, reflective struct check + 32-goroutine × 50 iteration race-clean, `CATALOG_SLIM_NOT_READONLY`). 모든 sentinel `t.Errorf` 사용 (CATALOG-001 eval-1 EC3 lesson 흡수).
- **`--all` flag + `shouldDistributeAll(cmd)` helper** (`internal/cli/init.go`, +42/-3): 좁은 env 매칭 (`"1"` exact OR case-insensitive `"true"`). EC2 idempotent (env+flag 동시 set 시 한번만 bypass).
- **slim mode 진입 안내**: `cmd.OutOrStdout()` 로 4 substring 1-line 출력 (REQ-021 notice — `"slim mode"` + `"--all"` + `"MOAI_DISTRIBUTE_ALL=1"` + `"SPEC-V3R4-CATALOG-005"`).

plan-auditor PASS 0.91 (≥ 0.88 stretch). 8 new files (slim_fs.go/_test, embed_catalog.go/_test, slim_guard.go/_test, catalog_slim_audit_test.go, init_slim_branch_test.go) + 1 modified (init.go +42/-3). DEFECT-5 encapsulation gate enforced. Fixes part of #859.

### English

- **BREAKING CHANGE — `moai init` default behavior**: deploys only `tier == core` catalog entries by default (20 core skills + 20 core agents + non-catalog templates). The 9 optional packs (backend / frontend / mobile / auth / deployment / design / devops / testing / chrome-extension — 17 skills + 7 agents total) and the builder-harness agent are no longer deployed by default — approximately 38% (25 / 65 entries) slim. Two opt-out paths: `moai init --all` flag OR `MOAI_DISTRIBUTE_ALL=1` environment variable (also accepts case-insensitive `"true"`). `moai update` behavior is unchanged (always full FS). Interactive optional-pack installation arrives in SPEC-V3R4-CATALOG-003 (`moai pack add`). Update drift sync for existing projects lands in SPEC-V3R4-CATALOG-004. Builder-harness auto-bootstrap lands in SPEC-V3R4-CATALOG-005.

- New `internal/template/slim_fs.go` SlimFS wrapper applies the tier filter at the `fs.FS` layer; `deployer.go` and `update.go` remain untouched (D7 lock). New `internal/template/embed_catalog.go` provides the encapsulated `LoadEmbeddedCatalog()` + `NewSlimDeployerWithRenderer()` entry points while keeping `embeddedRaw` unexported (DEFECT-5 invariant). New `internal/template/slim_guard.go` surfaces a friendly four-substring error when the builder-harness agent is absent. New `internal/template/catalog_slim_audit_test.go` adds 6 audit sub-tests (`CATALOG_SLIM_LEAK` / `CATALOG_SLIM_CORE_MISSING` + EC4 nested / `CATALOG_SLIM_OVER_FILTER` / `CATALOG_SLIM_WALK_LEAK` / `CATALOG_SLIM_NOT_READONLY` reflective + 32-goroutine race-clean) against the real production catalog. All sentinels use `t.Errorf` per the CATALOG-001 EC3 lesson.

- plan-auditor PASS 0.91 (≥ 0.88 stretch target). 8 new files + 1 modified. `git grep 'EmbeddedRaw[A-Za-z]*' internal/cli/` returns zero matches. Fixes part of #859.

## [Unreleased] — Dev Tooling: release-update workflow harness

### Added

- **release-update 워크플로우 하네스** (dev-only): CC 업스트림 릴리스 노트 추적 + moai-adk-go 문서 업데이트 자동화. `.claude/commands/97-release-update.md` (thin-wrapper), `.claude/skills/moai/workflows/release-update.md` (8-Phase 워크플로우), `.moai/state/last-cc-version.json` (상태 파일). `--since`, `--dry`, `--docs-only`, `--report-only`, `--master-spec` 플래그 지원. manager-docs (4개 locale 동기화) + manager-git (PR) 위임. `internal/template/templates/`에 미포함 (dev-only).

### English

- **release-update workflow harness** (dev-only): Automates CC upstream release note tracking and moai-adk-go documentation updates. `.claude/commands/97-release-update.md` (thin wrapper), `.claude/skills/moai/workflows/release-update.md` (8-phase workflow), `.moai/state/last-cc-version.json` (state file). Supports `--since`, `--dry`, `--docs-only`, `--report-only`, `--master-spec` flags. Delegates to manager-docs (4-locale sync) and manager-git (PR). Not included in `internal/template/templates/` (dev-only).

## [Unreleased] — SPEC-V3R4-CATALOG-001: 3-Tier Catalog Manifest (Foundation)

### Added

- **SPEC-V3R4-CATALOG-001**: 3-tier (`core` / `optional-pack:<name>` / `harness-generated`) 카탈로그 매니페스트 도입 — moai-adk-go skill/agent 슬림화 initiative 의 foundation SPEC. `internal/template/catalog.yaml` (37 skills + 28 agents = 65 entries, 9 optional packs, depends_on DAG) + `catalog_loader.go` (typed `LoadCatalog(fs.FS)` API, `LookupSkill`/`LookupAgent` accessors) + `catalog_tier_audit_test.go` (10 sentinel 기반 audit sub-tests: `CATALOG_MANIFEST_ABSENT`, `CATALOG_ENTRY_MISSING`, `CATALOG_ENTRY_ORPHAN`, `CATALOG_TIER_INVALID`, `PACK_DEPENDENCY_CYCLE`, `CATALOG_HASH_INVALID`, `CATALOG_DUPLICATE_ENTRY` 등) + `catalog_hash_norm.go` (LF + trailing-whitespace 정규화 후 sha256) + `scripts/gen-catalog-hashes.go` (offline 헬퍼) + `catalog_doc.md` (schema spec). `embed.go`에 `//go:embed catalog.yaml` directive 추가 (additive). `deployer.go` 미수정 (D7 lock). sync-auditor 독립 평가 PASS 0.82, LoadCatalog coverage 100%. 8 files, +1852/-0 LOC. PR #862 + #863. Wave 2 (Distribution: CATALOG-002+003), Wave 3 (Safety: 004), Wave 4 (Polish: 005+006+007) 진입 자격 충족. Fixes #859.

### English

- **SPEC-V3R4-CATALOG-001**: Introduced 3-tier (`core` / `optional-pack:<name>` / `harness-generated`) catalog manifest as the foundation SPEC of the moai-adk-go skill/agent slim-down initiative. New `internal/template/catalog.yaml` (37 skills + 28 agents = 65 entries, 9 optional packs, acyclic depends_on graph) + `catalog_loader.go` (typed `LoadCatalog(fs.FS)` API with `LookupSkill`/`LookupAgent` accessors) + `catalog_tier_audit_test.go` (10 sentinel-driven audit sub-tests including `CATALOG_MANIFEST_ABSENT`, `CATALOG_ENTRY_MISSING`, `CATALOG_ENTRY_ORPHAN`, `CATALOG_TIER_INVALID`, `PACK_DEPENDENCY_CYCLE`, `CATALOG_HASH_INVALID`, `CATALOG_DUPLICATE_ENTRY`) + `catalog_hash_norm.go` (LF + trailing-whitespace normalization → sha256) + `scripts/gen-catalog-hashes.go` (offline helper) + `catalog_doc.md` (schema spec). Added `//go:embed catalog.yaml` directive in `embed.go` (additive). `deployer.go` untouched (D7 lock). sync-auditor independent review PASS 0.82, LoadCatalog coverage 100%. 8 files, +1852/-0 LOC. PR #862 + #863. Unblocks Wave 2 (Distribution: CATALOG-002+003), Wave 3 (Safety: 004), Wave 4 (Polish: 005+006+007). Fixes #859.

## [Unreleased] — SPEC-V3R2-RT-007: Hardcoded Path Fix + Versioned Migration

### Added

- **SPEC-V3R2-RT-007**: 하드코딩된 경로 제거 및 버전 기반 마이그레이션 도입. `internal/migration/` 신규 패키지(runner, registry, version 추적, JSONL log appender, m001_hardcoded_path 마이그레이션). `internal/runtime/gobin/` 신규 패키지(Detect helper로 GOBIN/GOPATH/$HOME/go/bin 폴백 체인 일원화 — `initializer.go`와 `update.go`의 하드코딩 경로 제거). `moai migration {run,status,rollback}` CLI 3-subcommand 추가. `doctor migration` 헬스체크 통합. `session_start` 훅이 migration runner를 호출하여 세션 시작 시 자동 적용. Cross-platform lock: Unix는 `unix.Flock(LOCK_EX)`, Windows는 `O_EXCL` 파일 mutex(bounded retry 1s)로 분리. 29 files +2068/-667 LOC, CI all-GREEN. PR #846.

### English

- **SPEC-V3R2-RT-007**: Removed hardcoded paths and introduced versioned migration. New `internal/migration/` package (runner, registry, version tracking, JSONL log appender, m001_hardcoded_path migration). New `internal/runtime/gobin/` package (Detect helper unifies GOBIN/GOPATH/$HOME/go/bin fallback chain, eliminating hardcoded paths in `initializer.go` and `update.go`). Added `moai migration {run,status,rollback}` CLI subcommands and `doctor migration` health check. `session_start` hook now invokes the migration runner for automatic application at session start. Cross-platform lock: Unix uses `unix.Flock(LOCK_EX)`, Windows uses `O_EXCL` file mutex (bounded retry, 1s). 29 files +2068/-667 LOC, CI all-green. PR #846.

## [Unreleased] — SPEC-V3R2-ORC-001: Agent Roster Consolidation (22 → 17)

### Added

- **SPEC-V3R2-ORC-001**: 에이전트 카탈로그 22개에서 17개로 통합. manager-cycle → manager-develop(DDD/TDD 통합), agency 에이전트 6개 흡수(copywriter → moai-domain-copywriting skill, designer → moai-domain-brand-design skill), plan-auditor/sync-auditor 독립 유지. ORC-002~005 선행 작업. PR #843.

### English

- **SPEC-V3R2-ORC-001**: Agent catalog consolidation from 22 to 17 agents. manager-cycle → manager-develop (DDD/TDD unified), 6 agency agents absorbed into skills (copywriter → moai-domain-copywriting, designer → moai-domain-brand-design), plan-auditor/sync-auditor retained as independent. Prerequisite for ORC-002~005. PR #843.

## [Unreleased] — SPEC-V3R2-ORC-005: Dynamic Team Generation Formalization

### Added

- **SPEC-V3R2-ORC-005**: Agent Teams 동적 생성 공식화. 정적 team-* 에이전트 파일 삭제, `Agent(subagent_type: "general-purpose")` + workflow.yaml role_profiles로 런타임 생성. Mailbox Protocol v2, file ownership 전략, tmux pane lifecycle 관리. PR #743.

### English

- **SPEC-V3R2-ORC-005**: Dynamic Agent Teams generation formalization. Removed static team-* agent files; spawn via `Agent(subagent_type: "general-purpose")` with workflow.yaml role profiles at runtime. Mailbox Protocol v2, file ownership strategy, tmux pane lifecycle management. PR #743.

## [Unreleased] — SPEC-V3R3-STATUSLINE-FALLBACK-001: Statusline Stdin Fallback

### Added

- **SPEC-V3R3-STATUSLINE-FALLBACK-001**: `moai statusline` Go 바이너리의 stdin JSON empty/partial/null 시 모든 핵심 segment 눐락 문제 해결. cwd guard(삭제된 directory → HOME fallback), model name fallback chain(stdin → `MOAI_LAST_MODEL` env → cache file), project directory 4-level fallback(workspace.project_dir → workspace.current_dir → cwd → os.Getwd) 구현. `model_cache.go` 신규 파일(atomic write, EC-SF-003 silent ignore). AC-SF-001~008 + EC-SF-001~005 충족. MX tags 6개(ANCHOR 3, NOTE 3) 적용. 9 files, +614/-22 LOC, CI 16/16 ALL GREEN.

### English

- **SPEC-V3R3-STATUSLINE-FALLBACK-001**: Fixed `moai statusline` Go binary missing all critical segments when stdin JSON is empty/partial/null. Added cwd guard (deleted directory → HOME fallback), model name fallback chain (stdin → `MOAI_LAST_MODEL` env → cache file), and project directory 4-level fallback. New `model_cache.go` with atomic write. AC-SF-001~008 + EC-SF-001~005 met. 6 MX tags applied. 9 files, +614/-22 LOC, CI 16/16 green.

## [Unreleased] — SPEC-GLM-MCP-001: Z.AI MCP Server Integration

### Added

- **SPEC-GLM-MCP-001**: `moai glm tools enable|disable [vision|websearch|webreader|all]` 서브커맨드 추가. Z.AI 공식 MCP 서버(`zai-mcp-server`)를 사용자(`~/.claude.json`) 또는 프로젝트(`--scope project` → `.mcp.json`) 범위로 등록/해제합니다. Pre-flight로 Node.js >= v22.0.0과 GLM API 키 존재를 검증하고, `~/.claude.json` 수정 시 ISO 타임스탬프 백업과 atomic write로 손실을 방지합니다. 다른 `mcpServers` 항목은 byte-for-byte 보존되며, SPEC-GLM-001의 `DISABLE_BETAS=1` / `DISABLE_PROMPT_CACHING=1` env 정책과 직교(orthogonal)입니다 (REQ-GMC-001~010). 신규 파일: `internal/cli/glm_tools.go` (~600 LOC) + `internal/cli/glm_tools_test.go` (~1,150 LOC, 22 GWT scenarios, coverage ≥85%). `settings-management.md` (template + mirror)에 짧은 등록 노트 추가.

### English

- **SPEC-GLM-MCP-001**: Added `moai glm tools enable|disable [vision|websearch|webreader|all]` subcommand for registering or removing the official Z.AI MCP server (`zai-mcp-server`) at user (`~/.claude.json`) or project (`--scope project` → `.mcp.json`) scope. Pre-flight validates Node.js >= v22.0.0 and the GLM API key. ISO-timestamped backup + atomic write of `~/.claude.json` prevent loss; other `mcpServers` entries are preserved byte-for-byte. Orthogonal to SPEC-GLM-001 env policy (`DISABLE_BETAS=1` / `DISABLE_PROMPT_CACHING=1`). New `internal/cli/glm_tools.go` (~600 LOC) + `internal/cli/glm_tools_test.go` (~1,150 LOC, 22 GWT scenarios, coverage ≥85%). Short registration note added to `settings-management.md` (template + mirror).

## [Unreleased] — SPEC-V3R3-CLI-TUI-001: TUI Unification (M1-M7)

### English

- **SPEC-V3R3-CLI-TUI-001 (7 Milestones)**: Complete TUI integration layer with theme-driven component library, auto-detection, and 14-language i18n support.
  - **M1 (PR #799)**: TUI scaffolding — theme.go (light/dark tokens), Box/ThickBox, Pill components, runeguard helper (35 golden snapshots).
  - **M2 (PR #806)**: 6 TUI primitives — StatusIcon, Spinner, Progress, Stepper, RadioRow, CheckRow, form/table/prompt helpers (74 golden snapshots, reduced-motion support).
  - **M3 (PR #807)**: CLI integration — banner.go DDD migration, terra cotta/보라 hex removal, tui.Theme adoption, 3 Pill integration.
  - **M4 (PR #815, merged #085efe76f)**: 4-command batch — version, doctor, status, update DDD (9 commits, 4-sub-step orchestration, 1M context bypass).
  - **M5 (PR #823)**: Huh wizard + Stepper — huh v0.8.0 custom theming (◆/◇ prefix), ProfileEnv color-depth detection, 20 golden snapshots.
  - **M6 (PR #828)**: R-07 thin wrapper — cc/cg/glm/loop/statusline/help/error TUI rendering, NO_COLOR support, 37 golden snapshots.
  - **M7 (PR #831)**: Auto-detect + i18n — detect.go (MOAI_THEME + TTY + NO_COLOR logic), i18n.go (14-language embed.FS), golden suite (33 errchecks fixed).
  - **Total**: 33 files, 106 golden snapshots, 14-language i18n catalog, 10 @MX:ANCHOR + 8 @MX:NOTE, 17/17 AC complete.

### 한국어

- **SPEC-V3R3-CLI-TUI-001 (7 마일스톤)**: TUI 통합 계층 — 테마 기반 컴포넌트 라이브러리, 자동 감지, 14개 언어 i18n 지원.
  - **M1 (PR #799)**: TUI 골격 — theme.go (라이트/다크 토큰), Box/ThickBox, Pill 컴포넌트, runeguard 헬퍼 (35개 golden).
  - **M2 (PR #806)**: 6개 TUI primitive — StatusIcon, Spinner, Progress, Stepper, RadioRow, CheckRow, form/table/prompt 헬퍼 (74개 golden, reduced-motion).
  - **M3 (PR #807)**: CLI 통합 — banner.go DDD 마이그레이션, 테라코타/보라 hex 제거, tui.Theme 도입, 3개 Pill 통합.
  - **M4 (PR #815, merged #085efe76f)**: 4-command 배치 — version, doctor, status, update DDD (9개 commit, 4-sub-step 오케스트레이션, 1M context bypass).
  - **M5 (PR #823)**: Huh 마법사 + Stepper — huh v0.8.0 custom theming (◆/◇ prefix), ProfileEnv 색상 깊이 감지, 20개 golden.
  - **M6 (PR #828)**: R-07 thin wrapper — cc/cg/glm/loop/statusline/help/error TUI 렌더링, NO_COLOR 지원, 37개 golden.
  - **M7 (PR #831)**: 자동 감지 + i18n — detect.go (MOAI_THEME + TTY + NO_COLOR 로직), i18n.go (14개 언어 embed.FS), golden suite (33개 errcheck 수정).
  - **총계**: 33개 파일, 106개 golden snapshot, 14개 언어 i18n 카탈로그, 10개 @MX:ANCHOR + 8개 @MX:NOTE, 17/17 AC 완료.

## [Unreleased] — SPEC-V3R2-RT-005: Multi-Layer Settings Resolution with Provenance Tags

### English

- **SPEC-V3R2-RT-005 (5 Milestones, M1-M5 with M5 split into 4 sub-waves)**: Multi-layer settings resolution with 8-tier priority chain and Provenance tags.
  - **M1 RED (commit 06a74a401)**: 6 new test files, 41 new test functions establishing baseline (~940 LOC test scaffolding).
  - **M2 GREEN p1 (commit 77571f6f1)**: audit_registry.go with YAMLToStructRegistry + YAMLAuditExceptions (5 MIG-003 exceptions); yaml.v3 strict mode (KnownFields(true)); ConfigTypeError + ConfigAmbiguous + schema_version propagation. ACs: AC-05/08/11/12.
  - **M3 GREEN p2 (commit df8c3c63c)**: encoding/json byte-stable dump (sorted keys); Provenance.MarshalJSON; policy.strict_mode enforcement → PolicyOverrideRejected; YAML alphabetical sort + `# source: <tier>` comments. ACs: AC-01/02/07/09.
  - **M4 GREEN p3 (commit fc6acf70f)**: sync.RWMutex on resolver; tierData diff-cache; (\*resolver).Reload(path) API with tier-isolation; loadSkillTier real impl walks .claude/skills/\*\*/SKILL.md frontmatter; .moai/logs/config.log dedicated logger; ClearSessionTier for SessionEnd hook. ACs: AC-04/13.
  - **M5 Sub-wave A (commit fee736dfd)**: Doctor CLI (`moai doctor config dump|diff|--key`) verified with cmd.OutOrStdout(); loadBuiltinTier reflect-walks Config struct → 60+ keys; IsDefault flag for SrcBuiltin keys; filepath.Abs() normalization across 8 tier loaders. ACs: AC-02/09/10/14.
  - **M5 Sub-wave B (commit f54ab8d7b)**: validator/v10 v10.30.2 dependency; runStructValidation prepended to Validate(); 4 fields tagged (User.Name required, Quality.DevelopmentMode/LLM.PerformanceTier/GitConvention.Convention oneof). AC-05 strengthened.
  - **M5 Sub-wave C (commit 190810092)**: Diff merged-view delta semantics (REQ-051); SettingsResolver interface aligned (Dump io.Writer, Diff returns map without error); BenchmarkResolver_Load + Reload + TestResolver_MemoryFootprint with p50/p95/p99 reporting. ACs: AC-03/16/17/18.
  - **M5 Sub-wave D (commit a8e49371b)**: 7 MX tags inserted (3 ANCHOR + 2 NOTE + 2 WARN); CHANGELOG entry; full verification (go test ./..., go test -race, go vet, make build all clean); progress.md run_status: implementation-complete.
  - **Total**: 22 source/test files, +3759/-195 LOC, 17/18 AC GREEN (AC-15 ConfigSchemaMismatch stub deferred to SPEC-V3R2-EXT-004), 7 MX tags, 1549+ tests passing, race detector clean. Performance budgets verified far under target: Load p99=1011µs (100x margin), Reload p99=749µs (26x margin), HeapAlloc delta=645 KiB (3x margin).
  - **Unblocks downstream SPECs**: SPEC-V3R2-RT-002 (permission stack), RT-003 (sandbox routing), RT-006 (ConfigChange hook), RT-007 (hardcoded path migration), MIG-003 (5 loader additions).

### 한국어

- **SPEC-V3R2-RT-005 (5 마일스톤, M5는 4 sub-wave로 분할)**: 8-tier 우선순위 체인과 Provenance 태그를 가진 다층 설정 해상도.
  - **M1 RED (커밋 06a74a401)**: 6개 신규 테스트 파일, 41개 신규 테스트 함수로 baseline 구축 (~940 LOC 테스트 스캐폴딩).
  - **M2 GREEN p1 (커밋 77571f6f1)**: audit_registry.go (YAMLToStructRegistry + 5개 MIG-003 예외); yaml.v3 strict mode (KnownFields(true)); ConfigTypeError + ConfigAmbiguous + schema_version 전파. AC: AC-05/08/11/12.
  - **M3 GREEN p2 (커밋 df8c3c63c)**: encoding/json 바이트 안정 dump (정렬된 키); Provenance.MarshalJSON; policy.strict_mode 시행 → PolicyOverrideRejected; YAML 알파벳 정렬 + `# source: <tier>` 주석. AC: AC-01/02/07/09.
  - **M4 GREEN p3 (커밋 fc6acf70f)**: resolver에 sync.RWMutex; tierData diff 캐시; (\*resolver).Reload(path) API와 tier 격리; loadSkillTier 실제 구현 (.claude/skills/\*\*/SKILL.md frontmatter 워크); .moai/logs/config.log 전용 로거; SessionEnd hook용 ClearSessionTier. AC: AC-04/13.
  - **M5 Sub-wave A (커밋 fee736dfd)**: Doctor CLI (`moai doctor config dump|diff|--key`) cmd.OutOrStdout() 검증; loadBuiltinTier가 Config 구조체 reflect-walk → 60개 이상 키; SrcBuiltin 키에 IsDefault 플래그; 8개 tier loader 전반 filepath.Abs() 정규화. AC: AC-02/09/10/14.
  - **M5 Sub-wave B (커밋 f54ab8d7b)**: validator/v10 v10.30.2 의존성; runStructValidation을 Validate() 앞에 추가; 4개 필드 태그 (User.Name required, Quality.DevelopmentMode/LLM.PerformanceTier/GitConvention.Convention oneof). AC-05 강화.
  - **M5 Sub-wave C (커밋 190810092)**: Diff 병합 뷰 델타 의미론 (REQ-051); SettingsResolver 인터페이스 정렬 (Dump io.Writer, Diff는 error 없이 map 반환); BenchmarkResolver_Load + Reload + TestResolver_MemoryFootprint와 p50/p95/p99 보고. AC: AC-03/16/17/18.
  - **M5 Sub-wave D (커밋 a8e49371b)**: 7개 MX 태그 삽입 (3 ANCHOR + 2 NOTE + 2 WARN); CHANGELOG 항목; 전체 검증 (go test ./..., go test -race, go vet, make build 모두 클린); progress.md run_status: implementation-complete.
  - **총계**: 22개 source/test 파일, +3759/-195 LOC, 17/18 AC GREEN (AC-15 ConfigSchemaMismatch 스텁은 SPEC-V3R2-EXT-004로 이연), 7개 MX 태그, 1549+ 테스트 통과, race detector 클린. 성능 예산은 목표를 크게 밑돌음: Load p99=1011µs (100배 여유), Reload p99=749µs (26배 여유), HeapAlloc 델타=645 KiB (3배 여유).
  - **다운스트림 SPEC 잠금 해제**: SPEC-V3R2-RT-002 (퍼미션 스택), RT-003 (샌드박스 라우팅), RT-006 (ConfigChange hook), RT-007 (하드코딩 경로 이주), MIG-003 (5개 loader 추가).

## [Unreleased] — SPEC-V3R2-WF-005: Language Rules vs Skills Boundary Codification

### Changed

- SPEC-V3R2-WF-005: Codified that language guidance lives as rules (`.claude/rules/moai/languages/*.md`), not as skills. Added "Language Guidance Lives in Rules, Not Skills" section to `.claude/rules/moai/development/skill-authoring.md`. Removed dead `moai-lang-*` references from ~17 skills/rules. Substituted `moai-essentials-debug`, `moai-quality-testing`, `moai-quality-security`, and `moai-infra-docker` references per REQ-WF005-015. CI guard `lang_boundary_audit_test.go` enforces forward-looking compliance.

## [Unreleased] — SPEC-V3R3-CI-AUTONOMY-001: 8-Tier Autonomous CI/CD + Branch Origin Decision Protocol

### Added

- SPEC-V3R2-WF-004: Agentless fixed-pipeline classification for utility subcommands (`fix`, `coverage`, `mx`, `codemaps`, `clean`). Multi-agent classification preserved for `plan`, `run`, `sync`, `design`. Subcommand × class matrix published in `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification`. CI guard `TestAgentlessUtilityNoLLMControlFlow` enforces no-LLM-dispatch in utility skills. Sentinels `MODE_FLAG_IGNORED_FOR_UTILITY` and `MODE_PIPELINE_ONLY_UTILITY` added.
- SPEC-V3R2-WF-003: Multi-Mode Router (`--mode {autopilot|loop|team|pipeline}` flag) for `/moai run` and `/moai design`. `/moai loop` becomes an alias for `/moai run --mode loop` (REQ-WF003-004). Mode auto-selection by harness level (REQ-WF003-002, 003) with silent downgrade fallback (REQ-WF003-012). New optional `workflow.default_mode` config field (REQ-WF003-014). Sentinels: `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`. Subcommand × mode matrix extended in `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification` (5→8 columns, 9→10 rows incl. `/moai loop`). CI guards: `TestRunDesignSkillsContainModeUnknownSentinel`, `TestRunSkillContainsModeTeamUnavailableSentinel`, `TestLoopAliasCrossReference`.

### English

- **SPEC-V3R3-CI-AUTONOMY-001 (8 Waves, 8 Tiers)**: Establishes autonomous CI/CD pipeline with self-correcting watch+fix loops and pre-merge protocol enforcement. Eliminates manual debug/fix/push cycles for routine PR sweeps.
  - **Wave 1 (T1+T5)**: `make ci-local` 16-language detection + Bash skeleton with parallel test/lint/build (PR #785).
  - **Wave 2 (T2)**: CI watch loop. `internal/ciwatch/` (Classifier via SSoT + WatchState atomic YAML state file + Handoff FailedCheck JSON schema for Wave 3 expert-debug pipe), `internal/cli/pr/` (EmitReadyToMergeReport markdown with `(권장)` first option + EmitFailureHandoff JSON), `internal/cli/pr_watch_cmd.go` (`moai pr watch --abort` SetAbortFlag + `--report` subcommand). POSIX shell polling at `scripts/ci-watch/` (mock-injectable via `MOAI_CIWATCH_GH`) with `lib/_common.sh` + `lib/classify.sh` (yq + grep fallback) + `lib/timeout.sh` (30-min wall-clock guard). `.claude/skills/moai-workflow-ci-watch/` 3-tier Progressive Disclosure skill (SKILL.md 222 lines + modules/). `.claude/rules/moai/workflow/ci-watch-protocol.md` HARD invocation contract with auto-load paths frontmatter. Template-First mirror to `internal/template/templates/.claude/`. Coverage: ciwatch 87.6%, cli/pr 95.0%. Shell tests 9/9 pass; `make ci-local` PASS. (PR #788)
  - **Wave 3 (T3)**: Auto-fix loop. Mechanical (lint/format) vs semantic (test) failure classification, mechanical auto-fix, semantic immediate escalation. (PR #790)
  - **Wave 4**: Auxiliary workflow cleanup. claude-review (org quota) and Release Drafter (stale draft) marked non-blocking; required CI checks SSoT in `.github/required-checks.yml`. (PR #791)
  - **Wave 5 (T6)**: Worktree State Guard. `internal/worktree/` snapshot/verify/restore primitives, `moai worktree snapshot|verify|restore` CLI subcommands, divergence log + suspect flag for `Agent(isolation: "worktree")` regressions. (PR #792)
  - **Wave 6 (T7)**: i18n validator. Standalone Go static analyzer at `scripts/i18n-validator/` with dual-mode oracle (`--all-files` intra-state default + `--diff <git-rev>` temporal/baseline). Magic comment escape (`// i18n:translatable`) supported. (PR #793)
  - **Wave 7 (T8)**: Branch Origin Decision Protocol (BODP). `internal/bodp/` Pure-Go library with 3-signal evaluation (depends_on + co-location + open PR head) and 8-row decision matrix. Three invocation paths (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`). Audit trail at `.moai/branches/decisions/<normalized-branch>.md`. Off-protocol reminder in `moai status` with 4-skip-condition (env var + main/master + audit trail exists + dir absent). (PR #794)
  - **Follow-up #795**: BODP audit trail anchored to git repository root via `git rev-parse --show-toplevel` (overridable in tests). Replaces `os.Getwd()` which leaked audit trail files into the package directory during test execution. CLAUDE.local.md §6 Test Isolation compliance restored for 4 tests (`TestRunNew_Success`, `TestRunNew_AddError`, `TestRunNew_WithTmuxFlag_TmuxAvailable`, `TestRunNew_TmuxNotAvailable_GracefulDegradation`). (PR #795)
  - **AC-CIAUT-020 5-PR sweep replay (manual validation)**: 30-day grace window active **2026-05-09 → 2026-06-08**, anchored to PR #794 (Wave 7 T8 BODP) merge timestamp `2026-05-09T04:48:01Z UTC` (sha `02bed9c14`). PR #795 (BODP cwd leak follow-up fix, merged 05:22Z) is NOT the anchor PR (out of scope of T1-T8 core functionality). SPEC author + 1 reviewer to retrospectively measure manual debug intervention reduction across the next dev cycle (3+ PRs). Target: ≥50% reduction vs pre-SPEC baseline (~3-4 manual debugs/PR). Output: `.moai/reports/post-merge-validation/SPEC-V3R3-CI-AUTONOMY-001.md`.

### Technical Highlights

- **Methodology preservation**: Wave 5/6 surfaced sub-agent context inheritance failures (manager-tdd `worktreePath: {}`) — main-session direct implementation fallback applied per session lesson #12.
- **Template-First mirror**: All `.claude/` artifacts (rules, skills, agents) mirrored to `internal/template/templates/.claude/` per CLAUDE.local.md §2.
- **No release/tag automation**: Verified — none of the 8 PRs trigger GoReleaser auto-tag (T5 explicitly avoids tag mutation).
- **16-language neutrality**: Verified for T1 (make ci-local) and T7 (i18n validator) per CLAUDE.local.md §15.
- **No hardcoded URLs/models**: Verified per CLAUDE.local.md §14.
- **Quality gates**: All Waves pass `make ci-local` (5/5 steps) + `go test -race ./...` + `golangci-lint run`. Coverage targets met: bodp 85.9%, ciwatch 87.6%, cli/pr 95.0%, internal/worktree 87.6%, internal/cli/worktree 82.5%.

### 추가됨

- **SPEC-V3R2-WF-004**: utility 서브커맨드(`fix`, `coverage`, `mx`, `codemaps`, `clean`)에 대한 Agentless 고정 파이프라인 분류. `plan`, `run`, `sync`, `design`은 multi-agent 라우팅 유지. Subcommand × class matrix는 `.claude/rules/moai/workflow/spec-workflow.md#subcommand-classification`에 게시. CI guard `TestAgentlessUtilityNoLLMControlFlow`가 utility skill에서 LLM-dispatch 금지를 강제. Sentinel `MODE_FLAG_IGNORED_FOR_UTILITY`(info-only)와 `MODE_PIPELINE_ONLY_UTILITY`(WF-003과 공유) 추가. (PR #798 — sync retrofit PR)

### 한국어

- **SPEC-V3R3-CI-AUTONOMY-001 (8 Wave, 8 Tier)**: 자율 CI/CD 파이프라인 + self-correcting watch+fix loop + pre-merge 프로토콜 enforcement. 통상적 PR sweep에서 수동 debug/fix/push 사이클 제거 목표.
  - **Wave 1 (T1+T5)**: `make ci-local` 16개 언어 자동 감지 + 병렬 test/lint/build Bash skeleton (PR #785).
  - **Wave 2 (T2)**: CI watch 루프. `internal/ciwatch/` (Classifier SSoT + WatchState atomic YAML state file + Handoff FailedCheck JSON, Wave 3 expert-debug 파이프), `internal/cli/pr/` (EmitReadyToMergeReport markdown `(권장)` first option + EmitFailureHandoff JSON), `internal/cli/pr_watch_cmd.go` (`moai pr watch --abort` SetAbortFlag + `--report` 서브커맨드). POSIX shell polling (`scripts/ci-watch/`, mock-injectable via `MOAI_CIWATCH_GH`) + `lib/_common.sh` + `lib/classify.sh` (yq+grep fallback) + `lib/timeout.sh` (30분 wall-clock guard). `.claude/skills/moai-workflow-ci-watch/` 3-tier Progressive Disclosure (SKILL.md 222 lines + modules/). `.claude/rules/moai/workflow/ci-watch-protocol.md` HARD invocation contract (auto-load paths frontmatter). Template-First mirror (`internal/template/templates/.claude/`). Coverage: ciwatch 87.6%, cli/pr 95.0%. Shell tests 9/9 pass; `make ci-local` PASS. (PR #788)
  - **Wave 3 (T3)**: Auto-fix 루프. mechanical (lint/format) vs semantic (test) 분류, mechanical 자동 수정, semantic 즉시 escalation. (PR #790)
  - **Wave 4**: Auxiliary workflow 정리. claude-review (org quota) / Release Drafter (stale draft) 비차단 마킹, required CI check SSoT (`.github/required-checks.yml`). (PR #791)
  - **Wave 5 (T6)**: Worktree State Guard. `internal/worktree/` snapshot/verify/restore primitive, `moai worktree snapshot|verify|restore` CLI, divergence log + `Agent(isolation: "worktree")` 회귀 감지용 suspect flag. (PR #792)
  - **Wave 6 (T7)**: i18n validator. `scripts/i18n-validator/` Go static analyzer + dual-mode oracle (`--all-files` 기본 + `--diff <git-rev>` temporal). magic comment escape (`// i18n:translatable`). (PR #793)
  - **Wave 7 (T8)**: Branch Origin Decision Protocol (BODP). `internal/bodp/` Pure-Go 라이브러리 — 3시그널 평가(depends_on + co-location + open PR head) + 8행 decision matrix. 3개 invocation path (`/moai plan --branch`, `/moai plan --worktree`, `moai worktree new`). Audit trail (`.moai/branches/decisions/<normalized-branch>.md`). `moai status`에 off-protocol reminder + 4-skip-condition. (PR #794)
  - **Follow-up #795**: BODP audit trail이 git repo root (`git rev-parse --show-toplevel`)에 anchor되도록 수정. test 시 cwd가 패키지 디렉터리가 되어 audit trail이 `internal/cli/worktree/.moai/`에 누수되던 결함 제거. CLAUDE.local.md §6 Test Isolation 위반 4건 해소. (PR #795)
  - **AC-CIAUT-020 5-PR sweep replay (수동 검증)**: 30일 grace window 활성 **2026-05-09 ~ 2026-06-08**, 기준 PR은 #794 (Wave 7 T8 BODP) 머지 시각 `2026-05-09T04:48:01Z UTC` (sha `02bed9c14`). PR #795 (BODP cwd leak follow-up fix, 05:22Z 머지)는 기준 PR이 아님 (T1-T8 핵심 기능 외 fix). SPEC 작성자 + 1 reviewer가 다음 dev cycle (3+ PR) 회고하여 수동 debug 개입 감소율 측정. 목표: SPEC 머지 전 baseline (~3-4회/PR) 대비 50% 이상 감소. 결과물: `.moai/reports/post-merge-validation/SPEC-V3R3-CI-AUTONOMY-001.md`.

### 기술 하이라이트

- **방법론 보존**: Wave 5/6에서 sub-agent context 상속 실패 (manager-tdd `worktreePath: {}`) 발견 → main-session 직접 구현 fallback (lesson #12).
- **Template-First mirror**: 모든 `.claude/` 산출물 `internal/template/templates/.claude/` mirror (CLAUDE.local.md §2).
- **release/tag 자동화 없음**: 8개 PR 어느 것도 GoReleaser auto-tag 트리거 안 함 (T5는 tag mutation 명시적 회피).
- **16개 언어 중립성**: T1 (make ci-local) + T7 (i18n validator) 검증 (CLAUDE.local.md §15).
- **하드코딩 URL/model 없음**: CLAUDE.local.md §14 검증.
- **품질 게이트**: 모든 Wave가 `make ci-local` (5/5 단계) + `go test -race ./...` + `golangci-lint run` 통과. Coverage: bodp 85.9%, ciwatch 87.6%, cli/pr 95.0%, internal/worktree 87.6%, internal/cli/worktree 82.5%.

## [Unreleased] — SPEC-V3R3-RETIRED-AGENT-001: Retired Agent Stub 호환성 수정 + manager-cycle 템플릿 정합화

### Bug Fixes

- **SPEC-V3R3-RETIRED-AGENT-001**: Retired agent stub 호환성 fix + manager-cycle 템플릿 정합화. mo.ai.kr 사이드 프로젝트 2026-05-04 21:14:54 incident (5-layer defect chain → `[ERROR] Path "/Users/.../{}/{}" does not exist`) 차단.
  - 신규 `internal/template/templates/.claude/agents/moai/manager-cycle.md` (unified DDD/TDD implementation agent, SPEC-V3R2-ORC-001 retirement decision 완료).
  - 표준화된 retirement frontmatter: `retired: true`, `retired_replacement`, `retired_param_hint`, `tools: []`, `skills: []`. legacy `status: retired` custom field 제거.
  - SubagentStart hook retired-rejection guard 추가 (block decision JSON + exit code 2, 응답 시간 ≤500ms; 실측 0.056ms — mo.ai.kr 11.4s 대비 9000× 개선).
  - `validateWorktreeReturn` 헬퍼 + WORKTREE_PATH_INVALID sentinel: empty string, literal `{}`, `[object Object]`, `null`, `undefined` 패턴 거부 (5-layer chain Layer 4 차단).
  - manager-cycle workflow lifecycle hook dispatcher (`cycle-pre-implementation` / `cycle-post-implementation` / `cycle-completion`).
  - Documentation 7 references / 6 files substituted (CLAUDE.md, agent-hooks.md, agent-authoring.md, spec-workflow.md, manager-strategy.md, manager-ddd.md).
  - `agent_frontmatter_audit_test.go` CI assertion: retirement standardization 강제 + RETIREMENT_INCOMPLETE_<agent> sentinel.
  - **사용자 action**: `moai update` 실행으로 신규 template 자동 sync. `.moai/specs/`, `.moai/project/` 사용자 데이터 보존.

### Changed

- **SPEC-V3R3-RETIRED-DDD-001**: `manager-ddd` 에이전트 retired stub 표준화. 
  `manager-cycle`(cycle_type=ddd)로 통합. 
  33개 파일 내 `manager-ddd` → `manager-cycle` 치환 완료.
  사용자는 `moai update` 실행 시 자동 반영.

- **SPEC-V3R3-RETIRED-DDD-001**: Standardized `manager-ddd` agent as retired stub.
  Consolidated into `manager-cycle` (cycle_type=ddd).
  33 files updated with `manager-ddd` → `manager-cycle` substitution.
  Users receive changes automatically via `moai update`.

## [Unreleased] — SPEC-V3R3-BRAIN-001: /moai brain 7-phase 아이디에이션 워크플로우

### Added

- **`/moai brain` CLI command** (`internal/cli/brain.go`, 850 LOC): New cobra CLI entry point for ideation workflow. Thin-wrapper pattern delegating to `manager-brain` agent. Implements Phase 1 (Research) through Phase 7 (Export) with argument parsing for `--from-brain <IDEA-ID>` handoff mode and `--instructions-only` flag for prompt extraction. 13 table-driven unit tests (100% coverage), zero race conditions.

- **`manager-brain` agent** (`.claude/agents/manager-brain.md`, 520 LOC): New orchestration agent for 7-phase ideation workflow. Coordinates semantic decomposition (Phase 1), research parallel execution (Phase 2), conceptual design synthesis (Phase 3-4), design handoff package generation (Phase 5-6), and export (Phase 7). Delegates research to domain research skill, design to brand design skill, handoff to design-handoff skill. REQ-BRAIN-001~012 compliance verified per plan-audit iter3.

- **`moai-domain-ideation` skill** (`.claude/skills/moai-domain-ideation/SKILL.md`, 420 LOC): New domain expertise skill for ideation workflow Phase 1 (semantic decomposition). Parses user ideas into structured decomposition candidates with SPEC decomposition pathway matrix (5 pathways: feature, refactor, infra, docs, testing). Output artifact: `proposal.md` (paste-ready for `/moai plan` input).

- **`moai-domain-research` skill** (`.claude/skills/moai-domain-research/SKILL.md`, 380 LOC): Parallel research execution (Phase 2) combining WebSearch + Context7 MCP. Analyzes competitive landscape, market trends, and API/framework maturity for 5-pathway inputs. Output artifact: `research-summary.json` (structured competitive context, token-optimized ≤10K).

- **`moai-domain-design-handoff` skill** (`.claude/skills/moai-domain-design-handoff/SKILL.md`, 360 LOC): Phase 5-6 design handoff package automation. Generates Claude Design-compatible bundle (prompt.md, components.json, design-tokens.yaml, screenshot.md). Prompt is paste-ready without MoAI tokens; components spec enables Path A import in `/moai design`. 8-file worked example (IDEA-EXAMPLE/) demonstrates idempotent handoff at v0.1.0.

- **`IDEA-EXAMPLE/` worked example** (`.moai/brain/IDEA-EXAMPLE/`, 8 files, 2.2 KB): Complete ideation output artifact demonstrating 7-phase workflow on "MoAI Web Dashboard" concept. Files: idea.md (user input), proposal.md (Phase 1 decomposition), research-summary.json (Phase 2), design-brief.md (Phase 3-4), handoff-bundle.tar (Phase 5-6 export), export-log.md (Phase 7). Language-neutral (English comments, Korean example scenario).

- **Workflow patches** (3 files):
  - `project.md` (patch): Added `--from-brain <IDEA-ID>` flag for `/moai plan` Phase 8 auto-triggering
  - `plan.md` (patch): Decomposition parser enhancement (accepts `proposal.md` from Phase 1)
  - `design.md` (patch): Bundle auto-detect for handoff packages from Phase 5-6

- **Test coverage** (`internal/template/commands_audit_test.go`, +42 LOC): Extended `TestBrainCommandThinPattern` validating `/moai brain` thin-wrapper pattern (≤20 LOC body), argument parsing, and phase sequence enforcement (Phase 1→7 ordered gate).

- **`.moai/brain/` directory** (NEW): Reserved namespace for ideation artifacts (ideas/, proposals/, research/, designs/, handoffs/, exports/). Pattern matches `.moai/design/` architecture for design artifacts.

### Technical

- **7-phase orchestration** (manager-brain agent): Research (WebSearch+Context7) → Design (brand-aware synthesis) → Handoff (Claude Design export) → SPEC decomposition. REQ-BRAIN-001~012 traced end-to-end.
- **16-language neutrality**: All skills and examples support 16 canonical languages (go, python, typescript, javascript, rust, java, kotlin, csharp, ruby, php, elixir, cpp, scala, r, flutter, swift). No language-specific hardcoding in template tree.
- **Token optimization**: Research phase ≤10K, Design synthesis ≤8K, Handoff export ≤5K. Total Phase 2-6 budget ≤23K tokens. Enables ideation→SPEC→run pipeline within 250K session budget.
- **MX tag protocol**: 4 ANCHOR tags (@MX:ANCHOR) for ideation flow entry points, 2 WARN tags (@MX:WARN) for handoff export preconditions. Inline NOTEs for phase transitions.
- **Self-bootstrap capability**: `/moai brain "MoAI web dashboard"` → `proposal.md` → `/moai plan --from-brain IDEA-<auto-id>` → SPEC-V3R3-WEB-001 → `/moai run`. Demonstrates orchestrator self-referentiality (brain inspires web SPEC which may improve brain CLI).

### Breaking Changes

- None. New feature does not modify existing APIs or behavior.

### Coverage

- 12 EARS requirements (REQ-BRAIN-001~012) all traced to acceptance scenarios
- 13 unit tests (TestBrain*) + integration pattern validation via plan-audit
- 100% function coverage on `brain.go` CLI entry point
- Go test suite: 100% pass rate with race detection (`-race` flag)

## [Unreleased] — SPEC-V3R2-WF-002: Commands Thin-Wrapper Enforcement (98-github/99-release extraction)

### Added

- **`moai-workflow-github` skill** (`.claude/skills/moai-workflow-github/SKILL.md`, 723 LOC): Dev-only orchestration skill for GitHub issues and pull requests workflow. Extracted from `.claude/commands/98-github.md` (698 LOC body). Includes full GitHub workflow configuration, argument parsing (issues/pr sub-commands), pre-execution context loading, AskUserQuestion fallback for user decisions. Frontmatter declares `user-invocable: false` (dev-only; not surfaced to end users).

- **`moai-workflow-release` skill** (`.claude/skills/moai-workflow-release/SKILL.md`, 958 LOC): Dev-only orchestration skill for Enhanced GitHub Flow release orchestration. Extracted from `.claude/commands/99-release.md` (914 LOC body). Implements 9-phase release workflow (PHASE 0–8): pre-flight checks, quality gates, code review, version selection (AskUserQuestion), bilingual CHANGELOG generation (English-first per CLAUDE.local.md §18), final approval, release branch/tag, GitHub release notes, local environment update. Preserves all 11 metadata keys (`release_target`, `branch`, `tag_format`, `merge_strategy`, `reference_policy`, etc.). Frontmatter declares `user-invocable: false` (dev-only).

- **`TestRootLevelCommandsThinPattern` test** (`internal/template/commands_root_audit_test.go`, 146 LOC): Extends command audit suite to validate thin-wrapper compliance (≤20 LOC body) for root-level dev-only commands `98-github.md` and `99-release.md` in addition to `/moai/*` subcommands. Implements partial migration gate (REQ-WF002-015): verifies that `Skill("<name>")` references resolve to existing skill directories, blocking incomplete extractions. Test methodology: TDD RED→GREEN→REFACTOR via manager-tdd.

- **`TestDevOnlySkillLeak` test** (`internal/template/dev_only_skill_test.go`, 47 LOC): Negative-case validation that dev-only skills `moai-workflow-github` and `moai-workflow-release` do NOT appear in `EmbeddedTemplates()` (i.e., `internal/template/templates/.claude/skills/`). Fails with `DEV_ONLY_SKILL_LEAK` message if accidental template registration occurs. Enforces REQ-WF002-014.

### Changed

- **`.claude/commands/98-github.md`** (698 → 9 LOC total, 1 LOC body): Refactored to thin-wrapper delegating to `Skill("moai-workflow-github")` with `$ARGUMENTS` passthrough. Frontmatter preserved (`description`, `argument-hint`, `type: local`, `allowed-tools: Skill`). Version bumped 2.0.0 → 3.0.0 (extraction semantic change). Behavior preserved: `/98-github issues ...` and `/98-github pr ...` sub-commands now routed to skill implementation.

- **`.claude/commands/99-release.md`** (933 → 21 LOC total, 1 LOC body): Refactored to thin-wrapper delegating to `Skill("moai-workflow-release")` with `$ARGUMENTS` passthrough. Frontmatter preserved including 11 metadata keys and `disable-model-invocation: true` flag. Version bumped 5.0.0 → 6.0.0 (extraction semantic change). Behavior preserved: `/99-release [VERSION] [--hotfix]` now routes logic to skill while maintaining release orchestration semantics.

### Breaking Changes

- **[BC-V3R2-012] Command thin-wrapper extraction**: Internal mechanism change for maintainers — `.claude/commands/98-github.md` and `99-release.md` now delegate to extracted skills instead of inline orchestration logic. **User projects unaffected** (these are dev-only commands not templated into user `.claude/commands/`). Maintainer internal behavior preserved (behavior-preserving extraction).

### Technical

- **TDD methodology** (manager-tdd agent): M4 and M5 implemented via RED→GREEN→REFACTOR cycle. 4 negative test cases verified: leak inject/remove, skill dir rename/restore.
- **Behavior-preserving extraction**: H2 header count parity verified (98: 24→24, 99: 35→35). All GitHub sub-commands (issues/pr flags: --all, --label, --solo, --merge, NUMBER) preserved. All Release phases (PHASE 0–8) preserved with ordering parity check.
- **Dev-only skill containment**: Both skills reside in `.claude/skills/` (local/dev) only. Embedded template tree `internal/template/templates/.claude/skills/` validated empty via `TestDevOnlySkillLeak` (CI gate).
- **Partial migration prevention** (REQ-WF002-015): Commands audit test gate ensures `Skill("<name>")` wrapper dependencies are satisfied before allowing commit (skill dir must exist).
- **Binary size delta**: +193 LOC test code, −1,905 LOC commands + 1,881 LOC skills = net −31 LOC. Binary size impact <50 KiB (CI policy).
- **Coverage**: `go test ./internal/template/...` both new test files PASS. `go test -race ./...` no race conditions detected.

## [Unreleased] — SPEC-V3R2-EXT-001: Typed Memory Taxonomy (4-type enforcement)

### Added

- **statusline (SPEC-CC2122-STATUSLINE-001)**: Claude Code v2.1.122 `effort.level` and `thinking.enabled` indicator
  support in moai statusline. Compact segment rendering (e.g. `e:high·t`), silent omit on absent fields, graceful
  fallback for unknown enum values. 11/11 GWT scenarios PASS, 87.0% coverage, TDD methodology (M2-M6).
- **`internal/hook/memo/taxonomy` sub-package**: 4-type memory enum (`user | feedback | project | reference`) with
  `ParseFile`, `ValidateType`, `DetectStale`, `AggregateWarning`, `AuditFile`, `AuditIndex`, `AuditDuplicates`.
  91.7% test coverage. Source: SPEC-V3R2-EXT-001.
- **SessionStart staleness wrap** (`internal/hook/session_start.go`): Memory files with mtime > 24h are wrapped
  in `<system-reminder>` blocks with a verification caveat. Aggregated single warning when 10+ stale files detected.
- **PostToolUse memory audit** (`internal/hook/post_tool.go`): Non-blocking warnings on Write/Edit of agent-memory
  files: `MEMORY_MISSING_TYPE`, `MEMORY_MISSING_FRONTMATTER`, `MEMORY_BODY_STRUCTURE_MISSING`,
  `MEMORY_EXCLUDED_CATEGORY`, `MEMORY_INDEX_OVERFLOW`, `MEMORY_DUPLICATE`. Static-keyword v1 detection
  (LLM-based detection deferred to v2).
- **`memory:` config section** (`.moai/config/sections/workflow.yaml`): `staleness_threshold_hours: 24`,
  `index_line_cap: 200`, `stale_aggregate_threshold: 10`. Single source of truth for all thresholds.
- **Rule documentation** (`.claude/rules/moai/workflow/moai-memory.md`): 4-type taxonomy section with
  per-type writing guidelines, MEMORY.md 200-line cap explanation, excluded category enumeration.
- **`MOAI_MEMORY_AUDIT=0` rollback flag**: Disables both SessionStart wrap and PostToolUse audit (skip path).

### Technical

- New fixtures: `internal/hook/memo/taxonomy/fixtures/` (11 files for valid/invalid taxonomy permutations)
- Constants centralized in `internal/config/defaults.go` (no hardcoded literals 24/200/10)

## [Unreleased] — SPEC-V3R2-CON-001: FROZEN/EVOLVABLE Zone Registry

### Added

- **`moai constitution list`** CLI command: Browse and filter the zone registry by `--zone frozen|evolvable`,
  `--file <pattern>`, and `--format table|json`. Source: SPEC-V3R2-CON-001.
- **`moai constitution guard`** CLI command: Check a list of changed rule IDs for FROZEN zone violations.
  Returns non-zero exit on any Frozen-zone rule modification. Designed for CI pipelines.
- **Zone Registry** (`.claude/rules/moai/core/zone-registry.md`): Single source of truth for all HARD clauses
  across the MoAI rule tree. 68 entries: 38 Frozen, 30 Evolvable. Template twin included.
- **`internal/constitution` package**: `Zone`, `Rule`, `Registry`, `LoadRegistry`, `ValidateRuleReferences`.
  86.5% test coverage. 200-entry cold load benchmark: ~1.85ms (target <10ms).
- **Doctor constitution check** (`moai doctor`): `Constitution Registry` check validates registry existence,
  Frozen entry count, orphan warnings, and duplicate IDs. Supports `MOAI_CONSTITUTION_STRICT=1` strict mode.
- **Makefile `constitution-check` target**: Runs `moai constitution list --format json` against the live registry.
- **CI `constitution-check` job** (`.github/workflows/ci.yml`): `continue-on-error: true` job verifying
  registry integrity on every push to main and PR.

### Technical

- New package `internal/constitution`: `zone.go`, `rule.go`, `loader.go`, `dangling.go`
- Test fixtures: `internal/constitution/testdata/` (6 fixture files)
- Binary size delta: +33,600 bytes (~33 KiB, limit 50 KiB)
- Integration tests: `internal/cli/constitution_integration_test.go` (build tag: integration)

## [Unreleased] — SPEC-V3R2-WF-001: Skill Consolidation Stage 1 (48 → 38)

### Added

- **Stage 1 skill consolidation**: 48 skills → 38 surviving directories (11 RETIRE + 1 NEW `moai-design-system`).
  Source: SPEC-V3R2-WF-001 v1.1.0.
- **5 merge clusters**: `moai-foundation-thinking` (absorbs philosopher + workflow-thinking triplet),
  `moai-design-system` (NEW, absorbs design-craft + domain-uiux + Pencil portion of design-tools),
  `moai-domain-database` (absorbs platform-database-cloud), `moai-workflow-project` (absorbs templates + docs-generation + jit-docs),
  `moai-foundation-core` (absorbs foundation-context content).
- **11 retired skills** (archived to `.moai/archive/skills/v3.0/`):
  `moai-foundation-context`, `moai-foundation-philosopher`, `moai-workflow-thinking`, `moai-workflow-templates`,
  `moai-workflow-jit-docs`, `moai-domain-uiux`, `moai-design-craft`, `moai-design-tools` (Figma portion),
  `moai-docs-generation`, `moai-platform-database-cloud`, `moai-tool-svg`.
- **Trigger keyword union preservation**: All retired skill triggers migrated to merge target's `triggers:` or `related-skills:` frontmatter.
- **6 REFACTOR skills**: `moai-workflow-testing` (split bundled modules), `moai-domain-backend` (narrow to API matrix),
  `moai-domain-frontend` (router-only), `moai-platform-deployment` (Vercel-only), `moai-platform-auth` (narrower vendor guidance),
  plus 2 UNCLEAR telemetry windows (`moai-framework-electron`, `moai-platform-chrome-extension`).
- **CI fixture tests**: 2 broken-fixture suites at `.moai/specs/SPEC-V3R2-WF-001/fixtures/` validating retirement archive requirements and trigger drop detection.
- **Shared contract** (SPEC-V3R2-MIG-001): `.moai/decisions/skill-rename-map.yaml` artifact schema (v1) for migrator consumption.

### Breaking Changes

- **[BC-V3R2-006] Skill directory deletions**: Users with customized `.claude/skills/` may encounter deleted directories
  during `moai update`. Migrate per `.moai/archive/skills/v3.0/<name>/RETIRED.md` guidance.

### Technical

- Template/local byte-identity maintained across all `.claude/skills/` and `internal/template/templates/.claude/skills/`
  via `diff -rq` CI validation.
- 48→38 progression is Stage 1 of a 2-stage plan; Stage 2 (38→24) deferred to SPEC-V3R3-WF-001.
- Agency-absorbed skills (`moai-domain-copywriting`, `moai-domain-brand-design`) remain FROZEN per design constitution.

## [Unreleased] — SPEC-V3R2-WF-006: Output Styles Alignment

### Added

- **Output style schema validation** (frontmatter fields: `name`, `description`, `keep-coding-instructions`).
  Source: SPEC-V3R2-WF-006 v1.1.0.
- **Loading precedence**: Project-level `settings.json` `outputStyle` > user-level > hardcoded "MoAI" default.
  Documented in `.claude/rules/moai/core/settings-management.md`.
- **Fallback warning**: Unknown style names fall back to "MoAI" and emit `OUTPUT_STYLE_UNKNOWN: <name> not found; falling back to MoAI` to stderr.
- **CI drift check**: `make build` validates template and local output-styles byte-identity; rejects with `OUTPUT_STYLE_DRIFT` on divergence.
- **Schema audit CI** (`internal/template/output_styles_audit_test.go`): Rejects missing/malformed frontmatter with `OUTPUT_STYLE_SCHEMA_ERROR`.

### Technical

- Two styles remain stable: `MoAI` (`keep-coding-instructions: true`) and `Einstein` (`keep-coding-instructions: false`).
- Third style admission gated by schema validation (`OUTPUT_STYLE_UNVERIFIED` block).
- Template/local byte-identity maintained via `diff -rq` CI check (extends existing commands pattern).

## [Unreleased] — SPEC-WF-AUDIT-GATE-001: Plan Audit Gate (grace window 7d)

### Added

- **Plan Audit Gate** (`/moai run` Phase 0.5): Mandatory plan-auditor invocation before every
  `/moai run <SPEC-ID>` call. Prevents unreviewed SPEC artifacts from entering the implementation
  phase. Source: SPEC-WF-AUDIT-GATE-001.
- **Grace window** (7 days): FAIL verdicts emit warnings only until T0 + 7 days, after which
  they block Phase 1. Grace window T0 stored in `.moai/state/audit-gate-merge-at.txt`.
- **`--skip-audit` bypass**: Users can bypass the gate with `--skip-audit` or `MOAI_SKIP_PLAN_AUDIT=1`.
  All bypasses are recorded as BYPASSED verdict in `.moai/reports/plan-audit/` with timestamp, user, and rationale.
- **INCONCLUSIVE fall-back**: Auditor timeout/error/malformed output results in INCONCLUSIVE verdict.
  Never auto-PASS. Presents retry/proceed/abort options to user.
- **24h audit cache**: Repeated `/moai run` calls within 24h with unchanged plan artifacts reuse
  the cached PASS verdict without re-invoking plan-auditor.
- **Daily report persistence**: All gate invocations append to `.moai/reports/plan-audit/<SPEC-ID>-<DATE>.md`.
- **Team mode parity**: Gate applies equally in `--team` mode before any TeamCreate or teammate spawn.

### Technical

- New package `internal/runtime`: `audit_gate.go`, `audit_cache.go`, `audit_report.go`, `clock.go`
- Integration tests: `internal/cli/run_audit_gate_*_test.go`, `team_run_audit_gate_test.go`, `dogfood_self_audit_test.go`
- Workflow skills updated: `workflows/run.md`, `team/run.md`, `plan.md`, `spec-workflow.md`

## [2.13.2] - 2026-04-23

### Summary

Patch release completing the v2.13.0 `/agency` deprecation cycle ahead of schedule and resolving template/local drift across 60+ skill and rule files. The `/agency` redirect wrappers (originally scheduled for removal in the next minor version per REQ-DEPRECATE-003) are now fully removed. Template/local drift that accumulated since v2.13.0 is resolved, restoring HUMAN GATE quality gates, Drift Guard, `effort` field on reasoning-intensive agents, and the v2.13.0 DB Detection pipeline (Phase 4.1a) to local projects.

### Breaking Changes

None. The `/agency` command stubs were already marked DEPRECATED in v2.13.0 with a clear migration path to `/moai design`. Users who did not migrate will now see an "unknown command" error instead of the deprecation redirect.

### Added

- 3 skills synced from template to local: `moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`
- Evolvable blocks (`Common Rationalizations`, `Red Flags`, `Verification`) restored across 41 domain/foundation/workflow/library/platform/ref/tool skills
- v2.13.0 DB Detection (Phase 4.1a) restored in `moai/workflows/project.md` (+257 lines of DB engine/ORM/ODM keyword matrices across 16 supported languages)
- HUMAN GATE quality gates restored: 6 blocks across `plan.md`, `run.md`, `sync.md`
- `Drift Guard` section restored in `workflow-modes.md` (non-blocking scope drift check after DDD/TDD cycles)

### Changed

- `user-invocable: false` applied to `moai-domain-brand-design` and `moai-domain-copywriting` skills (internal `/moai design` pipeline components, no longer appear in `/` slash command menu)
- `effort: high` / `effort: xhigh` field restored on 7 reasoning-intensive agents (sync-auditor, expert-refactoring, expert-security, manager-spec, manager-strategy, plan-auditor, builder-agent) per CLAUDE.md Opus 4.7 HARD rule
- `skill-authoring.md` updated to Opus 4.7 effort level description (xhigh/max require Opus 4.7+)
- `lsp-client.md` template updated with powernap v0.1.3 to v0.1.4 upgrade notes (merged in #679)
- 16-language neutrality table restored in `loop.md`, `references/examples.md`, `references/reference.md` per CLAUDE.local.md Section 22

### Removed

- `/agency` deprecation stubs: 8 command files in `.claude/commands/agency/` and template parallel
- `.claude/rules/agency/constitution.md` redirect stub and template parallel
- `/agency (DEPRECATED)` section in CLAUDE.md and template CLAUDE.md

### Fixed

- Template/local skill drift: 41 skill files synced to pick up evolvable blocks added in recent template updates
- Missing `effort` field on reasoning-intensive agents caused Opus 4.7 to use default reasoning budget instead of configured level
- `moai/workflows/project.md` local version missing v2.13.0 DB Detection caused `/moai project` to skip DB inspection phase

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.13.2] - 2026-04-23 (한국어)

### 요약

v2.13.0에서 예정된 `/agency` deprecation 제거를 앞당겨 완전 제거하고, v2.13.0 이후 누적된 템플릿/로컬 드리프트 60여 개 파일을 해소하는 패치 릴리즈입니다. REQ-DEPRECATE-003에 따라 다음 minor 버전 예정이던 `/agency` redirect 래퍼가 제거되었고, HUMAN GATE 품질 게이트, Drift Guard, 추론 집약 에이전트의 `effort` 필드, v2.13.0 DB Detection 파이프라인(Phase 4.1a)이 로컬 프로젝트에 복구되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음. `/agency` 명령 스텁은 v2.13.0에서 이미 DEPRECATED로 표기되고 `/moai design` 이주 안내가 제공되었습니다. 미이주 사용자는 이제 deprecation 안내 대신 "unknown command" 오류를 보게 됩니다.

### 추가됨 (Added)

- 로컬 부재 스킬 3개 템플릿에서 동기화: `moai-domain-db-docs`, `moai-workflow-design-context`, `moai-workflow-pencil-integration`
- 41개 domain/foundation/workflow/library/platform/ref/tool 스킬에 evolvable blocks (`Common Rationalizations`, `Red Flags`, `Verification`) 복구
- `moai/workflows/project.md`에 v2.13.0 DB Detection (Phase 4.1a) 복구 — 16개 지원 언어의 DB engine/ORM/ODM 키워드 매트릭스 +257줄
- HUMAN GATE 품질 게이트 복구: `plan.md`, `run.md`, `sync.md`에 6개 블록
- `workflow-modes.md`에 `Drift Guard` 섹션 복구 (DDD/TDD 사이클 후 non-blocking scope drift 검사)

### 변경됨 (Changed)

- `moai-domain-brand-design`, `moai-domain-copywriting` 스킬에 `user-invocable: false` 적용 (/moai design 내부 파이프라인 컴포넌트, 슬래시 메뉴 노출 제거)
- 7개 추론 집약 에이전트(sync-auditor, expert-refactoring, expert-security, manager-spec, manager-strategy, plan-auditor, builder-agent)에 `effort: high` / `effort: xhigh` 필드 복구 (CLAUDE.md Opus 4.7 HARD rule 준수)
- `skill-authoring.md` Opus 4.7 effort level 설명 갱신 (xhigh/max는 Opus 4.7+ 요구)
- `lsp-client.md` 템플릿에 powernap v0.1.3 → v0.1.4 업그레이드 노트 반영 (#679)
- CLAUDE.local.md Section 22에 따라 `loop.md`, `references/examples.md`, `references/reference.md`에 16개 언어 중립성 테이블 복구

### 제거됨 (Removed)

- `/agency` deprecation 스텁: `.claude/commands/agency/` 8개 명령 파일 및 템플릿 대응
- `.claude/rules/agency/constitution.md` redirect 스텁 및 템플릿 대응
- CLAUDE.md 및 템플릿 CLAUDE.md의 `/agency (DEPRECATED)` 섹션

### 수정됨 (Fixed)

- 템플릿/로컬 스킬 드리프트: 최근 템플릿 업데이트에서 추가된 evolvable blocks를 반영하도록 41개 스킬 동기화
- 추론 집약 에이전트의 `effort` 필드 누락으로 Opus 4.7이 구성된 레벨 대신 기본 추론 예산을 사용하던 문제 해결
- `moai/workflows/project.md` 로컬 버전에서 v2.13.0 DB Detection이 누락되어 `/moai project`의 DB 검사 단계가 스킵되던 문제 해결

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.13.1] - 2026-04-23

### Summary

Patch release addressing slash command registration for `/moai db` and `/moai design` workflows introduced in v2.13.0. These commands existed as internal skill workflows but lacked the `.claude/commands/moai/*.md` wrapper files, preventing them from appearing in the Claude Code slash command menu.

### Breaking Changes

None.

### Added

- `.claude/commands/moai/db.md` command wrapper (local synced from template)
- `.claude/commands/moai/design.md` command wrapper (new, template + local)
- `.claude/skills/moai/workflows/db.md` local workflow (template sync for files missing locally)

### Changed

- `.claude/skills/moai/SKILL.md` router: `db` and `design` added to Priority 1 subcommand matching, description field, and Workflow Quick Reference blocks
- `.claude/skills/moai/workflows/design.md` synced from template (7.4K → 9.5K, adds Phase 0 pre-flight checks and path B details)
- `CLAUDE.md` Subcommands list: `design, db` inserted
- Template sources (`internal/template/templates/.claude/`) updated in parallel to preserve Template-First rule

### Fixed

- `/moai db` and `/moai design` now appear in Claude Code slash command menu (`moai:db`, `moai:design`)
- First-word subcommand matching in router now correctly routes `db` and `design` to their respective workflows
- Template/local drift between router SKILL.md and workflow files resolved

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.13.1] - 2026-04-23 (한국어)

### 요약

v2.13.0에 도입된 `/moai db`와 `/moai design` 워크플로우의 slash command 등록 누락을 수정하는 패치 릴리즈입니다. 내부 skill 워크플로우로는 존재했으나 `.claude/commands/moai/*.md` 래퍼 파일이 부재하여 Claude Code slash command 메뉴에 노출되지 않던 문제를 해결했습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- `.claude/commands/moai/db.md` command 래퍼 (local, template에서 동기화)
- `.claude/commands/moai/design.md` command 래퍼 (신규, template + local)
- `.claude/skills/moai/workflows/db.md` local 워크플로우 (local 부재분 template 동기화)

### 변경됨 (Changed)

- `.claude/skills/moai/SKILL.md` 라우터: Priority 1 subcommand 매칭, description 필드, Workflow Quick Reference 블록에 `db`와 `design` 추가
- `.claude/skills/moai/workflows/design.md` template에서 동기화 (7.4K → 9.5K, Phase 0 사전 검사 및 path B 세부 추가)
- `CLAUDE.md` Subcommands 목록에 `design, db` 삽입
- Template 원본(`internal/template/templates/.claude/`)도 Template-First 규칙 유지를 위해 동시 업데이트

### 수정됨 (Fixed)

- `/moai db`와 `/moai design`이 Claude Code slash command 메뉴에 정상 노출 (`moai:db`, `moai:design`)
- 라우터의 FIRST WORD subcommand 매칭이 `db`와 `design`을 각 워크플로우로 올바르게 라우팅
- 라우터 SKILL.md와 워크플로우 파일 간 template/local 드리프트 해소

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.13.0] - 2026-04-23

### Summary

Three independent workstreams converged in this release:

1. **SPEC-AGENCY-ABSORB-001 absorption (#682)** — `/agency` command and agents absorbed into the unified `/moai design` hybrid workflow. Brand context promoted to `.moai/project/brand/` as a constitutional constraint.
2. **Design + DB 8-SPEC integrated delivery** — `.moai/design/` folder scaffolding, `/moai db` command family, Pencil MCP integration, PostToolUse DB sync hook, and `/moai project` Phase 4.1a DB detection.
3. **Profile setup wizard hardening (#681)** — 16 review findings applied, silent data coercion of deprecated Claude IDs fixed, `ast_grep_gate` SAST re-enabled, team role_profiles rebalanced for Opus 4.7 / 1M-context models.

Additional highlights: LSP server detection fix restoring 16-language support (#689), `moai glm` settings.local.json pollution fix (#691), `charmbracelet/x/powernap` v0.1.3 → v0.1.4 (#679), and Hextra-based docs-site monorepo integration (#680).

### Breaking Changes

- `/agency` command deprecated — now redirects to `/moai design`. Full removal scheduled per REQ-DEPRECATE-003 (2 minor versions after this release).

### Added

**Design workflow (SPEC-AGENCY-ABSORB-001, SPEC-DESIGN-* series)**
- `/moai design` subcommand — Hybrid design workflow (Claude Design import path + code-based skill path)
- `moai migrate agency` command — Safe migration of .agency/ data to .moai/project/brand/ and .moai/config/sections/design.yaml
- `moai-domain-copywriting` skill — Brand-aligned copywriting with anti-AI-slop enforcement
- `moai-domain-brand-design` skill — Visual design system with hero-first chaining, WCAG 2.1 AA
- `moai-workflow-design-import` skill — Claude Design handoff bundle parser (ZIP/HTML)
- `moai-workflow-gan-loop` skill — Builder-Evaluator iteration with Sprint Contract protocol
- `moai-workflow-design-context` skill (SPEC-DESIGN-ATTACH-001) — `.moai/design/` bare-token auto-loader with priority truncation (`spec > system > research > pencil-plan`) and `ceiling(char/4) * 1.10` token budget enforcement
- `moai-workflow-pencil-integration` skill (SPEC-DESIGN-PENCIL-001) — Pencil MCP batch operation executor with DSL parser (I/M/R), 25-op batch split, layout verification, screenshot archival
- `.moai/design/` folder scaffolding (SPEC-DESIGN-DOCS-001) — README + research/system/spec templates with `_TBD_` markers; SHA-256 based user-edit preservation on `moai update`; reserved filename collision detection (exact + `filepath.Match`); non-empty-dir skip on `moai init`
- `.moai/project/brand/` directory — brand-voice.md, visual-identity.md, target-audience.md templates
- `.moai/config/sections/design.yaml` — Design pipeline configuration (GAN loop, sprint contract, evolution thresholds) + `design_docs` subsection (SPEC-DESIGN-ATTACH-001)
- `.claude/rules/moai/design/constitution.md` v3.3.0 (SPEC-DESIGN-CONST-AMEND-001) — Section 3 expanded to tripartite structure (3.1 Brand Context / 3.2 Design Brief / 3.3 Relationship); FROZEN zone extended to cover each subsection individually

**DB workflow (SPEC-DB-* series)**
- `/moai db` subcommand (SPEC-DB-CMD-001) — Thin wrapper (`commands/moai/db.md` <20 LOC) + router skill (`workflows/db.md` 9 phases) supporting `init`/`refresh`/`verify`/`list`. 16-language migration path mapping table.
- `.moai/project/db/` 7-file template set (SPEC-DB-TEMPLATES-001) — README, schema.md, erd.mmd (Mermaid `erDiagram`), migrations.md, rls-policies.md, queries.md, seed-data.md. `_TBD_` markers for interview-driven customization.
- `.moai/config/sections/db.yaml` (SPEC-DB-TEMPLATES-001) — 8-key structure (5 system + 3 interview): `enabled`, `dir`, `auto_sync`, `migration_patterns` (6 patterns: Prisma/Alembic/Rails/SQL/Supabase/generic), `engine`, `orm`, `multi_tenant`, `migration_tool`. Recursion guard via `.moai/project/db/**` exclusion.
- `moai-domain-db-docs` skill (SPEC-DB-SYNC-001) — Migration file parser facade + schema.md/erd.mmd/migrations.md synchronizer. Preserves user-edited sections and `_TBD_` markers.
- `moai hook db-schema-sync` subcommand (SPEC-DB-SYNC-001) — Go CLI for PostToolUse hook processing with 10s debounce state file, path traversal guard, proposal.json writer, non-blocking error logging.
- `handle-db-schema-change.sh` PostToolUse hook (SPEC-DB-SYNC-001) — Bash wrapper invoking `moai hook db-schema-sync` on Write/Edit events.
- `/moai project` Phase 4.1a DB Detection (SPEC-PROJECT-DB-HINT-001) — Auto-detects DB technology from `.moai/project/tech.md` + 16-language dependency manifests; conditionally surfaces `/moai db init` (Recommended, new project) or `/moai db refresh` (4th option, existing project) in Phase 4.2 Next Steps.

**Profile setup wizard hardening (#681 from origin/main)**
- `normalizeModel(m string) string` helper in `internal/cli/profile_setup.go` — maps deprecated Claude IDs to canonical aliases. Prevents silent loss of saved preferences when an option is removed from the wizard.
- Statusline migration banner (4 languages) — one-time notice when `existingPrefs` differs from normalized value.
- `auto` permission mode option in wizard — Claude Code v2.1.83+ / Sonnet 4.6+ gated option with runtime-failure disclaimer.
- Canonical validation slices + package constants (`defaultStatuslineMode`, `defaultStatuslineTheme`, `defaultPermissionMode`).
- New unit tests: `profile_setup_normalize_test.go` (19+13+7 rows) and `profile_setup_summary_test.go` (4 tests). New helpers at 100% line coverage.

**SPEC documents (this session)**
- SPEC-DESIGN-CONST-AMEND-001 / SPEC-DESIGN-DOCS-001 / SPEC-DESIGN-ATTACH-001 / SPEC-DESIGN-PENCIL-001 — Design workflow family
- SPEC-DB-CMD-001 / SPEC-DB-TEMPLATES-001 / SPEC-DB-SYNC-001 / SPEC-PROJECT-DB-HINT-001 — DB workflow family
- SPEC-DB-SYNC-HARDEN-001 — Hardening follow-up bundling 5 review warnings (file size guard, CheckDebounce atomicity, Windows platform branch, coverage ≥85%, MX tag annotations for 5 exported helpers). plan-auditor iter 1 FAIL → iter 2 PASS.

### Changed

**Design absorption (SPEC-AGENCY-ABSORB-001)**
- Agency Agents catalog reduced from 6 to 2 (copywriter, designer absorbed into skills; planner, builder, evaluator, learner removed per M5)
- `/agency` command redirected to `/moai design` with deprecation warning
- coding-standards.md: removed `Skill("agency")` reference

**Profile setup (#681)**
- `printProfileSummary` signature refactored to `(out io.Writer, t *profileSetupText, prefs *profile.ProfilePreferences, syncedProjectRoot string)` — enables unit testing via `bytes.Buffer` injection; pointer receivers avoid ~800B + 160B value copies.
- Permission mode option ordering: `auto` moved to position 2 for severity gradient.
- `SummarySyncSkipped` phrasing neutralized in all 4 locales.
- PermAuto labels strengthened with runtime-failure disclaimer (en/ko/ja/zh).
- ko/ja `SummaryHeader` — `입력된 값 확인:` → `저장된 설정값:`; `入力された設定値:` → `保存された設定値:`.
- Summary path rendering uses relative paths instead of absolute `filepath.Join(cwd, ...)`.
- `workflow.yaml` role_profiles reassignment: team lead `default_model` → `opus[1m]` (Opus 4.7 + 1M), `architect` → `opus`, `reviewer` → `sonnet` (up from `haiku`).
- `fmt.Fprintf` collapse in `printProfileSummary` (~6 fewer heap allocations per wizard-end).

### Fixed

**LSP server detection restored across 16 languages (#689)**
- Fixed complete LSP server detection failure: corrected `lsp.yaml` template YAML key `binary` → `command` (aligns with `ServerConfig` YAML tags), added `file_extensions` across all 16 languages (restores `detectLanguage()` file-to-server routing).
- Added template compliance tests (`TestTemplate_NoBinaryKey`, `TestTemplate_16LangCommandNonEmpty`, `TestTemplate_16LangFileExtensionsNonEmpty`): parses actual template files to prevent schema drift regression.

**`moai glm` settings.local.json permanent pollution fix (#691)**
- Fixed context window limit error when entering Claude Code after `moai glm`: removed `injectGLMEnvForTeam()` call from `applyGLMMode` to prevent GLM environment variables (`ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `DISABLE_PROMPT_CACHING`, etc.) from being persisted to `settings.local.json`. Since `setGLMEnv()` sets the current process env and `syscall.Exec` inherits it to `claude`, the file write was redundant and caused GLM mode to leak after session end.
- Added main session context-limit warning at `moai glm` startup: informs about full system-prompt retransmission (~30-40K tokens) from `DISABLE_PROMPT_CACHING=1`, Z.AI concurrency limits (paid tier 1-3 in-flight), and GLM model context window sizes. Recommends `moai cg` for the Claude-leader + GLM-teammate combination.
- Narrowed `injectGLMEnvForTeam` scope to `enableTeamMode()` (`moai --team` path) only. Only the `applyGLMMode` caller was removed — function itself retained.
- Added regression tests: `TestApplyGLMMode_NoSettingsLocalPollution`, `TestApplyGLMMode_ProcessEnvIsSet`, `TestGLMCmd_NoSettingsLocalPollution` (inverts the prior `TestGLMCmd_InjectsEnv`).

**Critical Review Findings (this session, post SPEC-DB-SYNC-001 review)**
- **Hook timeout unit bug** (`c6985e2fe`) — `settings.json.tmpl` PostToolUse `handle-db-schema-change.sh` entry had `"timeout": 30000` (8.3 hours). Claude Code hook timeout is in seconds (range 1-600). Corrected to `30`.
- **matchGlob path traversal** (`aa29a9316`) — `migrations/../../../etc/passwd.sql` style paths passed `migrations/**/*.sql` prefix match, enabling read of files outside project root in `proposal.json`. Added `filepath.Clean` + `../` escape rejection guard at `HandleDBSchemaSync` entry. 4-case regression test added.
- **Template-First rule violation** (`8a4022c69`) — SPEC-DESIGN-CONST-AMEND-001 updated the project constitution to v3.3.0 but the template tree copy remained at v3.2.0. New projects created via `moai init` would miss Section 3.2 Design Brief HARD rules. Synchronized template copy byte-for-byte.

**Profile setup (#681)**
- Silent data coercion (Critical) — users with deprecated model IDs in saved preferences no longer silently overwritten by `huh.Select`. Root cause: `huh.Select` binding falls back to cursor-landing when pre-bound value has no matching option. Mitigation: pre-coerce via `normalizeModel` before form binding.
- `ast_grep_gate` SAST re-enabled — `.moai/config/sections/quality.yaml` block restored. Previous removal silently disabled structural scanning.
- Dead-branch fallback removed — `valueOrDefault(prefs.StatuslineMode, "default")` simplified to direct access (post-normalize guaranteed non-empty).

**Lint cleanup** (`76ba50eab`)
- `defer f.Close()` errcheck in `internal/cli/design_folder.go:111` (hashFile) and `internal/hook/dbsync/db_schema_sync.go:312` (logError) wrapped in `defer func() { _ = f.Close() }()`.

**Hardening (SPEC-DB-SYNC-HARDEN-001 — 5 warning resolution)**
- **H1: parseMigrationStub file size guard** — Added package constant `maxMigrationFileSize = 1 << 20` (1 MiB) with `os.Stat` pre-check. Files exceeding the ceiling are now rejected without calling `os.ReadFile`, eliminating memory-pressure from malformed or malicious migration inputs. Return shape extended to `parseMigrationResult{ParsedContent, Truncated}` so callers can distinguish "genuinely empty" from "size-guarded" without either-or ambiguity (REQ-H1-001 ~ REQ-H1-003).
- **H2: CheckDebounce atomicity (signature unchanged)** — Concurrent `CheckDebounce` callers targeting the same `(stateFile, filePath, window)` now provably return `{false, true}` as a multiset (exactly-one-winner semantic, REQ-H2-002). Implementation uses a `stateFile + ".lock"` companion with `O_CREATE|O_EXCL` for mutual exclusion, plus the pre-existing `os.CreateTemp + os.Rename` pattern for torn-write-free state persistence. I/O failure on any step returns the safe default `(true, nil)` and logs to `ErrorLogFile` per REQ-H2-003. **plan.md had suggested `os.Rename` alone would suffice; empirical testing during implementation revealed `os.Rename` prevents torn writes but not decision races, hence the added O_EXCL layer.** AST-level regression guard (`TestCheckDebounce_NoDirectWriteFile`) asserts zero direct `os.WriteFile` + at least one `os.Rename` call, surviving variable renames.
- **H3: settings.json.tmpl Windows platform branch for db-schema-change** — The db-schema-change PostToolUse hook entry was the sole exception to the file's `{{- if eq .Platform "windows"}}...{{- else}}...{{- end}}` branching convention (16 other entries used it consistently). Now aligned. Windows command: `bash "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-db-schema-change.sh"` (bash prefix + forward-slash + bash-style variable). Unix command unchanged. `TestRender_DbSchemaChangeHook_{Unix,Windows,ConsistencyWithOtherEntries}` verify the rendering outputs. SPEC v0.2.0 drafted a conflicting literal (`%CLAUDE_PROJECT_DIR%\.claude\...` with backslashes) that does not expand inside a bash-quoted argument; corrected to v0.2.1 during `/moai run`.
- **H4: internal/hook/dbsync test coverage 79.2% → 85.7%** — 8 boundary test cases added (`empty_file`, `utf8_bom`, `oversized`, `nonexistent`, `trailing_slash`, `double_star_only`, `unicode_path`, `corrupt_state_recovery`). All use `t.TempDir()` isolation; no `t.Setenv("HOME", …)`. Internal test file (`db_schema_sync_internal_test.go`) added for strict `parseMigrationStub` return-shape assertions that require unexported-symbol access.
- **H5: @MX:NOTE godoc for 5 exported helpers** — `BuildProposal`, `MatchesMigrationPattern`, `IsExcluded`, `CheckDebounce` gained godoc blocks with input/output/side-effect three-element contracts (Korean prose per `code_comments: ko`). `HandleDBSchemaSync` retained its pre-existing `@MX:NOTE + @MX:ANCHOR` from commit `aa29a9316`. AC-9 awk-based multiline scan confirms 5/5 coverage.
- **SPEC v0.2.1 amendment** (`38350a698`) — REQ-H3-002 and AC-6 literals corrected to match the file's actual hook entry convention. HISTORY entry records the reason (Git Bash/WSL cannot expand `%CLAUDE_PROJECT_DIR%` inside a bash-quoted argument).

### Removed

- `.claude/agents/agency/` agent definitions: planner, builder, evaluator, learner, copywriter, designer
- `.claude/skills/agency-*` forked skills: agency-copywriting, agency-design-system, agency-evaluation-criteria, agency-client-interview, agency-frontend-patterns
- `.claude/skills/agency/` orchestrator skill
- Fork management via `fork-manifest.yaml` (absorbed into moai-workflow-research)

### Deprecated

- `/agency` subcommands (brief, build, review, learn, evolve, resume, profile) now redirect to equivalent `/moai` subcommands. Scheduled for removal per REQ-DEPRECATE-003 (2 minor versions after this release)

### Migration

- Existing projects with `.agency/` directories can migrate via `moai migrate agency`
- Migration is atomic, reversible (data preserved as `.agency.archived/`), and handles SIGINT/SIGTERM with `--resume` flag
- See SPEC-AGENCY-ABSORB-001 acceptance.md for full behavior

### Testing

- `internal/cli` coverage maintained at 75.3% (wizard sub-package 91.2%, worktree sub-package 84.2%).
- All new profile_setup helpers at 100% line coverage.
- `internal/hook/dbsync` package: **coverage 79.2% → 85.7%** after SPEC-DB-SYNC-HARDEN-001 H4. 13 pre-existing unit tests + H1/H2/H4/H5 additions (8 AC-8 boundary cases, `TestCheckDebounceConcurrency` running at `-race -count=10`, `TestCheckDebounce_NoDirectWriteFile` AST regression guard, oversized-file handler test, readonly-dir safety-default tests).
- `internal/cli` design_folder: 282-line test file covering SHA-256 preservation, glob collision, .DS_Store-only directory handling.
- `go vet ./...`, `go test -race ./... -count=1` (all packages), `golangci-lint run ./internal/...` — all PASS.
- Cross-compile verified for 5 targets: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.13.0] - 2026-04-23 (한국어)

### 요약

세 개의 독립 워크스트림이 이번 릴리즈에서 수렴되었습니다.

1. **SPEC-AGENCY-ABSORB-001 흡수 완료 (#682)** — `/agency` 명령어와 에이전트가 통합 `/moai design` 하이브리드 워크플로우로 흡수되었습니다. 브랜드 컨텍스트가 `.moai/project/brand/`로 승격되어 헌법적 제약으로 작동합니다.
2. **Design + DB 8 SPEC 통합 구현** — `.moai/design/` 폴더 스캐폴딩, `/moai db` 명령어 패밀리, Pencil MCP 통합, PostToolUse DB 동기화 훅, `/moai project` Phase 4.1a DB 감지.
3. **Profile setup wizard 하드닝 (#681)** — 16개 리뷰 지적사항 반영, deprecated Claude ID 무음 데이터 강제 변환 수정, `ast_grep_gate` SAST 재활성화, Opus 4.7 / 1M 컨텍스트 모델에 맞춰 팀 role_profiles 재조정.

추가 하이라이트: LSP 서버 감지 수정으로 16개 언어 지원 복원 (#689), `moai glm` settings.local.json 오염 수정 (#691), `charmbracelet/x/powernap` v0.1.3 → v0.1.4 (#679), Hextra 기반 docs-site 모노레포 통합 (#680).

### 주요 변경 사항 (Breaking Changes)

- `/agency` 명령어 deprecated — `/moai design`으로 리다이렉트. REQ-DEPRECATE-003에 따라 본 릴리즈 2개 마이너 버전 후 완전 제거 예정.

### 추가됨 (Added)

**Design 워크플로우 (SPEC-AGENCY-ABSORB-001, SPEC-DESIGN-* 패밀리)**
- `/moai design` 서브 명령어 — 하이브리드 디자인 워크플로우 (Claude Design 임포트 경로 + 코드 기반 스킬 경로)
- `moai migrate agency` 명령어 — `.agency/` 데이터를 `.moai/project/brand/`와 `.moai/config/sections/design.yaml`로 안전하게 마이그레이션
- `moai-domain-copywriting` 스킬 — anti-AI-slop 적용된 브랜드 정렬 카피라이팅
- `moai-domain-brand-design` 스킬 — hero-first 체이닝 및 WCAG 2.1 AA 준수 시각 디자인 시스템
- `moai-workflow-design-import` 스킬 — Claude Design 핸드오프 번들(ZIP/HTML) 파서
- `moai-workflow-gan-loop` 스킬 — Sprint Contract 프로토콜 기반 Builder-Evaluator 반복 루프
- `moai-workflow-design-context` 스킬 (SPEC-DESIGN-ATTACH-001) — `.moai/design/` 자동 로더, 우선순위 기반 truncation (`spec > system > research > pencil-plan`) 및 토큰 예산 강제
- `moai-workflow-pencil-integration` 스킬 (SPEC-DESIGN-PENCIL-001) — Pencil MCP 배치 연산 실행기 (DSL 파서, 25-op 배치 분할, 레이아웃 검증, 스크린샷 아카이브)
- `.moai/design/` 폴더 스캐폴딩 (SPEC-DESIGN-DOCS-001) — README + research/system/spec 템플릿, `moai update` 시 SHA-256 기반 사용자 수정 보존, 예약 파일명 충돌 감지
- `.moai/project/brand/` 디렉토리 — brand-voice.md, visual-identity.md, target-audience.md 템플릿
- `.moai/config/sections/design.yaml` — 디자인 파이프라인 설정 (GAN loop, sprint contract, evolution 임계치) + `design_docs` 서브섹션
- `.claude/rules/moai/design/constitution.md` v3.3.0 — Section 3 삼분할 구조 확장 (3.1 Brand Context / 3.2 Design Brief / 3.3 Relationship)

**DB 워크플로우 (SPEC-DB-* 패밀리)**
- `/moai db` 서브 명령어 (SPEC-DB-CMD-001) — Thin 래퍼 + 라우터 스킬 (`init`/`refresh`/`verify`/`list` 지원), 16개 언어 마이그레이션 경로 매핑
- `.moai/project/db/` 7-파일 템플릿 세트 — README, schema.md, erd.mmd (Mermaid `erDiagram`), migrations.md, rls-policies.md, queries.md, seed-data.md
- `.moai/config/sections/db.yaml` — 8-키 구조 (5 시스템 + 3 인터뷰) + 6개 마이그레이션 패턴 (Prisma/Alembic/Rails/SQL/Supabase/generic), `.moai/project/db/**` 재귀 가드
- `moai-domain-db-docs` 스킬 — 마이그레이션 파서 facade + schema.md/erd.mmd/migrations.md 동기화
- `moai hook db-schema-sync` 서브 명령어 — PostToolUse 훅 처리 (10초 debounce, path traversal 가드, proposal.json writer)
- `handle-db-schema-change.sh` PostToolUse 훅 — Write/Edit 이벤트 시 `moai hook db-schema-sync` 호출
- `/moai project` Phase 4.1a DB 감지 — `tech.md` + 16-언어 의존성 매니페스트로 DB 기술 자동 감지

**Profile setup wizard 하드닝 (#681)**
- `normalizeModel(m string) string` 헬퍼 — deprecated Claude ID를 정규 별칭으로 매핑하여 저장된 설정 무음 손실 방지
- 4개 언어 statusline 마이그레이션 배너 — `existingPrefs` 정규화 시 1회성 알림
- `auto` 권한 모드 선택지 (Claude Code v2.1.83+ / Sonnet 4.6+ 게이팅, 런타임 실패 경고 포함)
- 정규 검증 슬라이스 + 패키지 상수 (`defaultStatuslineMode`, `defaultStatuslineTheme`, `defaultPermissionMode`)
- 신규 단위 테스트 추가 (`profile_setup_normalize_test.go`, `profile_setup_summary_test.go`, 신규 헬퍼 100% 라인 커버리지)

**SPEC 문서**
- SPEC-DESIGN-CONST-AMEND-001 / SPEC-DESIGN-DOCS-001 / SPEC-DESIGN-ATTACH-001 / SPEC-DESIGN-PENCIL-001 — 디자인 워크플로우 패밀리
- SPEC-DB-CMD-001 / SPEC-DB-TEMPLATES-001 / SPEC-DB-SYNC-001 / SPEC-PROJECT-DB-HINT-001 — DB 워크플로우 패밀리
- SPEC-DB-SYNC-HARDEN-001 — 5개 경고 해결 하드닝 (파일 크기 가드, CheckDebounce 원자성, Windows 분기, 커버리지 ≥85%, MX 태그)

### 변경됨 (Changed)

**Design 흡수 (SPEC-AGENCY-ABSORB-001)**
- Agency 에이전트 카탈로그 6 → 2 축소 (copywriter, designer가 스킬로 흡수; planner, builder, evaluator, learner 제거)
- `/agency` 명령어가 deprecation 경고와 함께 `/moai design`으로 리다이렉트
- coding-standards.md에서 `Skill("agency")` 참조 제거

**Profile setup (#681)**
- `printProfileSummary` 시그니처 리팩토링 — `bytes.Buffer` 주입으로 단위 테스트 가능, 포인터 수신자로 사본 복사 제거
- 권한 모드 선택지 순서 조정 — `auto`를 2번 위치로 이동 (심각도 gradient)
- 4개 locale에서 `SummarySyncSkipped` 표현 중립화
- PermAuto 라벨에 런타임 실패 경고 추가 (en/ko/ja/zh)
- ko/ja `SummaryHeader` — `입력된 값 확인:` → `저장된 설정값:`; `入力された設定値:` → `保存された設定値:`
- Summary 경로 렌더링이 절대 경로 대신 상대 경로 사용
- `workflow.yaml` role_profiles 재조정 — 팀 리더 `default_model` → `opus[1m]`, `architect` → `opus`, `reviewer` → `sonnet` (기존 `haiku`에서 상향)

### 수정됨 (Fixed)

**LSP 서버 감지 16개 언어 복원 (#689)**
- LSP 서버 감지 전체 비활성 문제 수정: `lsp.yaml` 템플릿 YAML 키를 `binary` → `command`로 수정 (`ServerConfig` YAML 태그 일치), 16개 언어 전체에 `file_extensions` 추가 (`detectLanguage()` 파일-서버 라우팅 복원)
- 템플릿 준수 테스트 3종 추가하여 스키마 드리프트 재발 방지

**`moai glm` settings.local.json 영구 오염 수정 (#691)**
- `moai glm` 실행 후 Claude Code 진입 시 context window limit 에러 수정: `applyGLMMode`에서 `injectGLMEnvForTeam()` 호출을 제거하여 `settings.local.json`에 GLM 환경변수가 영구 기록되던 동작 방지
- `moai glm` 시작 시 메인 세션 컨텍스트 한도 경고 메시지 추가 (DISABLE_PROMPT_CACHING 영향, Z.AI 동시성 한도, GLM 컨텍스트 윈도우 안내). Claude 리더 + GLM 팀원 조합은 `moai cg` 권장
- `injectGLMEnvForTeam` 범위를 `enableTeamMode()`(`moai --team` 경로)로만 제한
- 회귀 테스트 3종 추가 (`TestApplyGLMMode_NoSettingsLocalPollution` 등)

**DB 하드닝 (SPEC-DB-SYNC-HARDEN-001)**
- 파일 크기 가드 (1 MiB 제한), CheckDebounce 원자성 (O_EXCL + os.Rename), settings.json.tmpl Windows 분기 정렬, `internal/hook/dbsync` 커버리지 79.2% → 85.7%, 5개 exported helper에 @MX:NOTE godoc 추가

### 제거됨 (Removed)

- `.claude/agents/agency/` 에이전트 정의: planner, builder, evaluator, learner, copywriter, designer
- `.claude/skills/agency-*` 포크된 스킬: agency-copywriting, agency-design-system, agency-evaluation-criteria, agency-client-interview, agency-frontend-patterns
- `.claude/skills/agency/` 오케스트레이터 스킬
- `fork-manifest.yaml` 포크 관리 (moai-workflow-research에 흡수)

### Deprecated

- `/agency` 서브명령어 (brief, build, review, learn, evolve, resume, profile)은 `/moai` 상응 서브명령어로 리다이렉트. REQ-DEPRECATE-003에 따라 2개 마이너 버전 후 제거 예정

### 마이그레이션 (Migration)

- `.agency/` 디렉토리를 가진 기존 프로젝트는 `moai migrate agency`로 마이그레이션 가능
- 마이그레이션은 atomic 및 reversible (데이터는 `.agency.archived/`로 보존), SIGINT/SIGTERM 시 `--resume` 플래그 지원
- 전체 동작은 SPEC-AGENCY-ABSORB-001 acceptance.md 참조

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.12.0] - 2026-04-17

### Summary

Claude Code v2.1.110/111 + Claude Opus 4.7 (`claude-opus-4-7`) compatibility layer (SPEC-OPUS47-COMPAT-001 + post-implementation follow-up). Introduces a 5-tier Effort system, Opus 4.7 prompt philosophy (5 principles documented in `moai-constitution.md`), v2.1.110 runtime handlers (MCP scope duplicate detection, PermissionRequest `updatedInput` re-verification, Windows `CLAUDE_ENV_FILE` injection), and `ApplyEffortPolicy` production wiring that automatically injects effort levels into agent frontmatter during `moai init` / `moai update` while preserving user customisation. Backward compatible with Opus 4.6 / Sonnet 4.6 / Haiku 4.5.

### Breaking Changes

None. `agentModelMap [3]string` signature unchanged (NFR-1). `effort` is opt-in — agents without the field inherit runtime defaults. Existing Opus 4.6 user profiles are preserved (no forced migration).

### Added

- **Effort system** (`internal/profile/preferences.go`, `internal/cli/launcher.go`, `internal/config/envkeys.go`)
  - `ProfilePreferences.EffortLevel string` YAML field (`effort_level`, omitempty)
  - `buildEnvForLaunch(effortLevel, baseEnv)` — injects `CLAUDE_CODE_EFFORT_LEVEL` at `syscall.Exec` time
  - `EnvClaudeCodeEffortLevel` constant

- **Model policy** (`internal/template/model_policy.go`)
  - `ModelIDOpus47 = "claude-opus-4-7"` constant
  - `EffortLevel{Low,Medium,High,XHigh,Max}` constants (5-tier)
  - `agentEffortMap` — explicit overrides for 6 high-reasoning agents (manager-spec/manager-strategy → `xhigh`; plan-auditor/sync-auditor/expert-security/expert-refactoring → `high`)
  - `GetAgentEffort(agentName string) string` exported function
  - `ApplyEffortPolicy(projectRoot, manifestMgr) error` — called by `moai init` and `moai update`; injects effort overrides into agent `.md` frontmatter, preserves any existing `effort:` value (user customisation wins), registers hash changes in the manifest

- **Profile setup UI** (`internal/cli/profile_setup.go`, `internal/cli/profile_setup_translations.go`)
  - `claude-opus-4-7` model option in model selector
  - 5-tier effort selector with localised labels (en/ko/ja/zh)

- **Doctor check** (`internal/cli/doctor.go`)
  - `checkMCPScopeDuplicates` — detects MCP server name collisions between project `.mcp.json` and global `~/.claude/.mcp.json`; warning-level only, `exit 0`

- **Hook: PermissionRequest** (`internal/hook/permission_request.go`)
  - Deny when `ToolInput` contains the `__updated_input_marker__` sentinel (updatedInput re-verification, T-015)

- **Hook: SessionStart** (`internal/hook/session_start.go`)
  - `injectCLAUDEEnvFile` — Windows-only: injects `CLAUDE_ENV_FILE` path into `settings.local.json` when project `.env` exists (T-016); macOS/Linux paths unchanged (R-P1-1 regression verified)

- **Template: settings.json**
  - `disableBypassPermissionsMode: false` field (v2.1.111)

- **Template: harness.yaml**
  - `effort_mapping` section: thorough → `xhigh`, standard → `high`, minimal → `medium`

- **Template: quality.yaml**
  - `session_effort_default: "xhigh"` field

- **Rules documentation**
  - `moai-constitution.md` — new "Opus 4.7 Prompt Philosophy" section with 5 principles (one-turn fully-loaded, Adaptive Thinking, scaffolding removal, explicit fan-out, fewer tool calls)
  - `agent-authoring.md` — new "Bash Tool Timeout Ceiling" section (600,000ms documented)
  - `agent-common-protocol.md` — Bash timeout documentation + parallel fan-out principle
  - `worktree-integration.md` — minimum version table expanded with v2.1.110/111 rows; `Recommended: 2.1.111 or later`
  - `skill-authoring.md` — `effort` field now lists `low/medium/high/xhigh/max` (removed "max is Opus 4.6 only" phrasing)
  - `coding-standards.md` — Claude Code version compatibility table
  - `CLAUDE.md §12` — UltraThink vs Adaptive Thinking vs `--deepthink` (Sequential Thinking MCP) distinction clarified
  - `moai-workflow-thinking/SKILL.md` — Adaptive Thinking redefined; fixed `thinking.budget_tokens` instructions removed (Opus 4.7 rejects fixed budgets with HTTP 400)

- **Local development docs**
  - `CLAUDE.local.md` — `settings.local.json` runtime-management rule + `OTEL_LOG_RAW_API_BODIES` production warning

- **Testing** (post-implementation follow-up, PR #673)
  - Coverage boost in `internal/hook` 77.7% → **85.0%** (12 new table-driven tests)
  - R-P1-1 Handle-level Windows guard regression tests in `internal/hook/session_start_windows_guard_test.go` (verifies macOS/Linux paths do NOT call `injectCLAUDEEnvFile`)
  - `internal/cli` coverage 73.6% → 75.1% (partial; 85% target deferred to future SPEC)

### Changed

- `llm.yaml` template: `claude_models.high` tier updated from `"opus"` to `"claude-opus-4-7"` — ensures `high` policy lands on Opus 4.7 on new projects

### Fixed

- `GetAgentEffort` now has a production caller (`ApplyEffortPolicy`) — resolves W-5 dead-code warning from Phase 2.5 manager-quality review

### Installation & Update

```bash
# Fresh install
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version

# Existing users update
moai update

# Verify version
moai version
```

---

## [2.12.0] - 2026-04-17 (한국어)

### 요약

Claude Code v2.1.110/111 및 Claude Opus 4.7(`claude-opus-4-7`) 호환성 레이어(SPEC-OPUS47-COMPAT-001 + 후속 정리). 5단계 Effort 시스템, Opus 4.7 프롬프트 철학 5원칙(`moai-constitution.md` 명문화), v2.1.110 런타임 핸들러(MCP 중복 감지, PermissionRequest `updatedInput` 재검증, Windows `CLAUDE_ENV_FILE` 주입), 그리고 `moai init`/`moai update` 시점에 agent frontmatter로 effort 값을 자동 주입하되 사용자 커스텀을 보존하는 `ApplyEffortPolicy`를 추가했습니다. Opus 4.6/Sonnet 4.6/Haiku 4.5와 완전 하위 호환.

### 주요 변경 사항 (Breaking Changes)

없음. `agentModelMap [3]string` 시그니처 불변(NFR-1). `effort`는 opt-in으로 미설정 에이전트는 런타임 기본값을 상속. 기존 Opus 4.6 사용자 프로파일은 유지(강제 마이그레이션 없음).

### 추가됨 (Added)

- **Effort 시스템** (`internal/profile/preferences.go`, `internal/cli/launcher.go`, `internal/config/envkeys.go`)
  - `ProfilePreferences.EffortLevel string` YAML 필드(`effort_level`, omitempty)
  - `buildEnvForLaunch(effortLevel, baseEnv)` — `syscall.Exec` 시점에 `CLAUDE_CODE_EFFORT_LEVEL` 주입
  - `EnvClaudeCodeEffortLevel` 상수

- **모델 정책** (`internal/template/model_policy.go`)
  - `ModelIDOpus47 = "claude-opus-4-7"` 상수
  - `EffortLevel{Low,Medium,High,XHigh,Max}` 5단계 상수
  - `agentEffortMap` — 6개 추론 집약 에이전트 명시적 override (manager-spec/strategy → `xhigh`, 나머지 4개 → `high`)
  - `GetAgentEffort(agentName)` 신규 함수
  - `ApplyEffortPolicy(projectRoot, manifestMgr)` — `moai init`/`moai update`가 호출; agent `.md` frontmatter에 effort 주입, 기존 `effort:` 값은 보존(사용자 커스텀 우선), manifest 해시 갱신

- **Profile 설정 UI** (`internal/cli/profile_setup.go`, `profile_setup_translations.go`)
  - `claude-opus-4-7` 모델 선택지
  - 5단계 effort 선택기(한/영/일/중 4개 언어)

- **Doctor 진단** (`internal/cli/doctor.go`)
  - `checkMCPScopeDuplicates` — 프로젝트 `.mcp.json`과 전역 `~/.claude/.mcp.json` 간 서버 이름 충돌 감지; warning 수준(`exit 0`)

- **Hook: PermissionRequest** (`internal/hook/permission_request.go`)
  - `ToolInput`에 `__updated_input_marker__` sentinel 포함 시 deny (updatedInput 재검증, T-015)

- **Hook: SessionStart** (`internal/hook/session_start.go`)
  - `injectCLAUDEEnvFile` — Windows 전용: 프로젝트 `.env` 존재 시 `settings.local.json`에 `CLAUDE_ENV_FILE` 경로 주입 (T-016); macOS/Linux 기존 경로 영향 없음 (R-P1-1 회귀 검증)

- **템플릿 파일**
  - `settings.json.tmpl` — `disableBypassPermissionsMode: false` (v2.1.111)
  - `harness.yaml` — `effort_mapping` 섹션 (thorough→xhigh, standard→high, minimal→medium)
  - `quality.yaml.tmpl` — `session_effort_default: "xhigh"`

- **규칙 문서**
  - `moai-constitution.md` — "Opus 4.7 Prompt Philosophy" 섹션 신설 (5원칙: 1턴 몰빵, Adaptive Thinking, 스캐폴딩 제거, 명시적 fan-out, 툴 호출 감소)
  - `agent-authoring.md` — "Bash Tool Timeout Ceiling" 섹션 신설 (600,000ms 문서화)
  - `agent-common-protocol.md` — Bash timeout + parallel fan-out 원칙
  - `worktree-integration.md` — 최소 버전 표에 v2.1.110/111 추가, 권장 버전 2.1.111+
  - `skill-authoring.md` — `effort` 필드를 `low/medium/high/xhigh/max`로 확장 ("max is Opus 4.6 only" 문구 제거)
  - `coding-standards.md` — Claude Code 버전 호환성 표
  - `CLAUDE.md §12` — UltraThink vs Adaptive Thinking vs `--deepthink`(Sequential Thinking MCP) 구분 명확화
  - `moai-workflow-thinking/SKILL.md` — Adaptive Thinking 재정의, 고정 `thinking.budget_tokens` 지시 제거 (Opus 4.7은 고정 예산 시 HTTP 400 오류)

- **로컬 개발 문서**
  - `CLAUDE.local.md` — `settings.local.json` 런타임 관리 원칙 + `OTEL_LOG_RAW_API_BODIES` 프로덕션 경고

- **테스트** (후속 정리, PR #673)
  - `internal/hook` 커버리지 77.7% → **85.0%** (12개 신규 table-driven 테스트)
  - R-P1-1 Handle 레벨 Windows 가드 회귀 테스트 신설 (`session_start_windows_guard_test.go`); macOS/Linux에서 `injectCLAUDEEnvFile` 미호출 검증
  - `internal/cli` 커버리지 73.6% → 75.1% (부분 개선; 85% 완전 달성은 별도 SPEC으로 연기)

### 변경됨 (Changed)

- `llm.yaml` 템플릿: `claude_models.high` 계층을 `"opus"` → `"claude-opus-4-7"`로 갱신 — 신규 프로젝트에서 `high` 정책이 Opus 4.7을 사용하도록 보장

### 수정됨 (Fixed)

- `GetAgentEffort`에 프로덕션 호출자(`ApplyEffortPolicy`) 추가 — Phase 2.5 manager-quality 리뷰가 지적한 W-5 dead-code 경고 해소

### 설치 및 업데이트 (Installation & Update)

```bash
# 신규 설치
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version

# 기존 사용자 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.11.0] - 2026-04-16

### Summary

First tagged release since v2.10.2, consolidating 32 merged PRs. Headlines: (1) LSP Suite (SPEC-LSP-CORE-002..MULTI-006) — powernap-based multi-language foundation + phase-aware quality gates, (2) Skill Evolution Infrastructure — 5-layer safety + telemetry, (3) 3-perspective security+quality review fixing 16 defects across evolution/telemetry/gopls/astgrep packages, (4) pre-tool quality gate cross-compilation fix (#667), (5) GLM team-mode credential propagation (#640), (6) comprehensive Windows CI compatibility. For detailed entries see v2.10.3, v2.10.4, and the fixes documented below.

### Breaking Changes

None. `lsp.client_impl: gopls_bridge` default maintains existing behavior.

### Fixed (post-v2.10.4)

- **#667 — pre-tool quality gate blocks Bash on macOS cross-compile projects** (PR #668)
  - `gateStep.changedExts` field + `stagedFiles()` helper: skips language-specific lint steps when the staged changeset contains no matching file extensions
  - Applied to `dotnet format` (`.cs` extensions only) — macOS users with Windows-only TFM .NET solutions can now `git commit` non-C# files
  - Defense in depth: `isDotnetRestoreFailure()` detects NuGet restore failure markers (`Restore operation failed`, `NU1202`, `NETSDK1005`, `not supported on this platform`) and logs a warning instead of blocking
  - `GateConfig.DisabledSteps map[string]bool` — per-project step-level disable flag
  - Conservative fallback: git binary missing / outside git repo / empty staging → runs step (preserves existing behavior)

- **#640 — moai glm + --team 401 Unauthorized on tmux teammates** (PR #669)
  - New `internal/hook/glm_tmux.go` with `ensureTmuxGLMEnv()` — SessionStart hook auto-injects GLM credentials into tmux session env when user set up GLM mode outside tmux
  - Propagates 9 GLM vars: `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU}_MODEL`, compatibility flags
  - Guard rails: `TMUX == ""` / `teammateMode != "tmux"` / missing `AUTH_TOKEN` / missing tmux binary → all graceful no-op, never aborts SessionStart
  - UX: `moai glm` non-tmux path now prints actionable 4-step recovery instructions

### Changed

- `TestAsyncRecorder_NonBlockingUnderLoad` (`internal/telemetry/async_recorder_test.go`) now skips on Windows CI due to scheduler granularity causing flaky latency assertions. Non-blocking invariant is still verified on Linux/macOS.
- `TestQualityGate_RunsDotnetFormatWhenCSharpStaged` (`internal/hook/quality/gate_test.go`) skips on Windows because shell-script fake binaries cannot be executed directly by `exec.Command` on Windows.

### Installation & Update

```bash
# Update to the latest development version
moai update

# Verify version
moai version
```

---

## [2.11.0] - 2026-04-16 (한국어)

### 요약

v2.10.2 이후 첫 정식 태그 릴리즈로 32개 PR을 통합합니다. 헤드라인: (1) LSP Suite (SPEC-LSP-CORE-002..MULTI-006) — powernap 기반 다국어 기반 + 단계별 품질 게이트, (2) Skill Evolution Infrastructure — 5계층 안전 + 텔레메트리, (3) 3관점 보안·품질 리뷰 — evolution/telemetry/gopls/astgrep 16 결함 수정, (4) pre-tool 품질 게이트 크로스컴파일 수정(#667), (5) GLM team-mode 자격증명 전파(#640), (6) 포괄적 Windows CI 호환성. 세부 항목은 v2.10.3, v2.10.4 및 아래 수정 섹션 참조.

### 주요 변경 사항 (Breaking Changes)

없음. `lsp.client_impl: gopls_bridge` 기본값으로 기존 동작 유지.

### 수정됨 (Fixed, post-v2.10.4)

- **#667 — macOS 크로스컴파일 .NET 프로젝트에서 pre-tool 품질 게이트가 Bash 차단** (PR #668)
  - `gateStep.changedExts` 필드 + `stagedFiles()` 헬퍼로 staged changeset에 해당 확장자 없으면 언어별 lint 단계 skip
  - `dotnet format`을 `.cs` 확장자로 제한 → Windows-only TFM .NET 솔루션 포함 프로젝트에서 비-C# 파일 `git commit` 가능
  - Defense in depth: `isDotnetRestoreFailure()`가 NuGet 복원 실패 마커 감지 시 경고 후 통과
  - `GateConfig.DisabledSteps` — per-project 단계 명시적 비활성화
  - 보수적 폴백: git 미설치 / git 저장소 밖 / 빈 스테이징 → 기존 동작 유지

- **#640 — moai glm + --team 조합 시 tmux 팀원 401 Unauthorized** (PR #669)
  - `internal/hook/glm_tmux.go` 신규 파일 + `ensureTmuxGLMEnv()`: SessionStart 훅에서 tmux 세션 env 자동 주입
  - GLM 변수 9종 일괄 전파: `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, `ANTHROPIC_DEFAULT_*_MODEL`, 호환 플래그
  - Guard: `TMUX` 없음 / `teammateMode != "tmux"` / `AUTH_TOKEN` 없음 / tmux 바이너리 없음 → graceful no-op
  - UX: `moai glm` 비-tmux 경로에 actionable 4단계 복구 안내 추가

### 변경됨 (Changed)

- `TestAsyncRecorder_NonBlockingUnderLoad`: Windows CI 스케줄러 입도 문제로 latency 기반 검증이 flaky하여 Windows skip. 비블로킹 불변식은 Linux/macOS에서 계속 검증.
- `TestQualityGate_RunsDotnetFormatWhenCSharpStaged`: Windows는 shell-script 기반 fake 바이너리를 `exec.Command`로 실행할 수 없어 skip.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 개발 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.10.4] - 2026-04-15

### Summary

Security + quality hardening across three packages (evolution, telemetry, gopls bridge, astgrep) driven by a 3-perspective review bundle. 16 defects fixed (CRITICAL 6 + IMPORTANT 8 + SUGGESTION 2) with reproduction-first TDD. Includes file-destruction bug fix in `apply.go:80` (MergeEvolvableZones API misuse), two path traversal vulnerabilities, OWASP-aligned binary allowlists for `gopls`/`sg`, async telemetry writer, Windows CI cross-platform compatibility, and strengthened AskUserQuestion-only interaction rules.

### Breaking Changes

None. All changes are bug fixes with no API deletions.

### Added

#### Security & Quality (PR #636 Skill Evolution Infrastructure)

- **CRITICAL 1 & 2 — Path traversal rejection** (`internal/evolution/safety.go`, `learning.go`)
  - `CheckFrozenGuard`: `filepath.IsAbs` + leading `/`/`\` detection + normalized `..` rejection
  - `ApplyProposal`: `filepath.Rel` containment check to prevent projectRoot escape
  - `validateLearningID`: regex `^LEARN-\d{8}-\d{3}$` enforced on `CreateLearning`, `UpdateLearning`, `LoadLearningByID`
  - New sentinel errors: `ErrInvalidLearningID`
- **CRITICAL 3 — MergeEvolvableZones API misuse (file destruction)** (`internal/evolution/apply.go`, `internal/merge/evolvable_zone.go`)
  - New helper: `merge.ReplaceEvolvableZone(content, zoneID, newZoneContent)` for in-place zone substitution
  - Replaces buggy 3-way merge usage at `apply.go:80` that silently destroyed file headers/footers
  - Exported `merge.ErrZoneNotFound`
- **CRITICAL 4 — Frozen Guard coverage** (`internal/evolution/safety.go`)
  - Added `.claude/rules/agency/` prefix + `.agency/fork-manifest.yaml` to frozen set
- **CRITICAL 5 & 6 — Async telemetry + file handle reuse** (`internal/telemetry/async_recorder.go`)
  - `AsyncRecorder`: single writer goroutine, channel-based, drop policy on buffer full
  - Date-keyed file handle cache with `bufio.Writer` (4KB buffer, flush every 16 records)
  - `GetRecorder(projectRoot)` singleton with `sync.Mutex`
  - `ErrRecordDropped` sentinel
- **IMPORTANT — `ensureNewSkillSymlinks` name validation** (`internal/hook/session_start.go`)
  - Rejects `..`, `/`, `\`, null bytes, hidden files as skill entry names
- **IMPORTANT — `UpdateRateLimit` race fix** (`internal/evolution/safety.go`)
  - `rateMu sync.Mutex` serializes Read→mutate→Write sequence

#### gopls Bridge Hardening (PR #660 / Issue #643)

- **F1 — RFC 3986 URI encoding** (`internal/lsp/gopls/uri.go`)
  - New `pathToURI()` helper via `url.URL` for space/unicode/Windows drive paths
- **F2 — Per-URI pending map** (`internal/lsp/gopls/bridge.go`)
  - `pendingMu` + `pendingDiag map[string][]Diagnostic` prevents event drop when processing files sequentially
- **F3 — Prompt shutdown** (`internal/lsp/gopls/bridge.go`)
  - `Close()` closes stdout to unblock `readLoop`'s blocking `Read()`
- **F4 — Timer leak** (`internal/lsp/gopls/bridge.go`)
  - `time.After` → `time.NewTimer` + `defer timer.Stop()`
- **F5 — Binary allowlist** (`internal/lsp/gopls/config.go`)
  - `validateBinary` / `validateArgs` with trusted prefixes and shell-metachar rejection
  - New errors: `ErrUntrustedBinary`, `ErrUnsafeArgs`
  - Invoked on `LoadConfig` + `NewBridge` (defense in depth)

#### astgrep Hardening (PR #661 / Issue #642)

- **F1 — `filterByLang` implementation** (`internal/astgrep/scanner.go`, `internal/cli/astgrep.go`)
  - New `Finding.Language` field populated from `rule.Language` during scan
  - Case-insensitive lang filter; empty Language treated as language-neutral (always included)
- **F2 — Binary allowlist** (`internal/astgrep/scanner.go`)
  - Public `ValidateBinary()` + `trustedBinaryPrefixes()`
  - Allows bare names `sg`/`ast-grep` or trusted absolute prefixes
  - Rejects `..`, shell metachars, untrusted paths
- **F3 — Context cancel defer-safety** (`internal/astgrep/scanner.go`)
  - `runSingleRule` helper per-rule isolation with `defer cancel()`
- **F4 — stderr logging** (`internal/astgrep/scanner.go`)
  - Captured stderr logged via `slog.Debug`; propagated as error when stdout empty
- **F5 — YAML resilience** (`internal/astgrep/rules.go`)
  - `loadFileSkipOnError` now `---`-splits documents for independent parsing; malformed docs no longer drop subsequent rules

#### Documentation (PR #663)

- **CLAUDE.md [HARD] AskUserQuestion-Only Interaction**: All user-facing questions must use AskUserQuestion (no free-form prose questions)
- **Section 8 expansion**:
  - "Socratic Interview via AskUserQuestion" subsection: round design + bias prevention rules
  - Beginner-friendly option design: first option always "(Recommended)" + detailed description
  - "Ambiguity Triggers" section: explicit trigger list for discovery mode
  - `4 questions per call` constraint documented
- Applied to both `internal/template/templates/CLAUDE.md` (Template-First) and local `CLAUDE.md`; `make build` regenerated `internal/template/embedded.go`

### Changed

- `TestAsyncRecorder_NonBlockingUnderLoad`: threshold relaxed from 10ms→100ms with 5% slow-call allowance for Windows CI scheduler variance
- `deduplicateSummaries` (unused helper) commented out in `reflective_write.go`
- `globalRecorderOnce` (unused) removed from `async_recorder.go`
- Template `CLAUDE.md` HARD rules table: new `AskUserQuestion-Only Interaction` entry; `Context-First Discovery` line clarified to reference AskUserQuestion

### Fixed

- **#636** 3-perspective review blockers (8 fixes: CRITICAL 6 + IMPORTANT 2)
- **#643** gopls bridge 5 defects (F1-F5)
- **#642** astgrep bundle 5 defects (F1-F5)
- Windows CI cross-platform path handling:
  - `filepath.IsAbs("/usr/bin/sg")` returns false on Windows → added `strings.HasPrefix(binary, "/")` detection
  - `filepath.Clean` converts `/` to `\` on Windows → use `filepath.ToSlash` for cross-platform prefix comparison
  - Backslash removed from binary shell-metachar list (Windows legitimate paths)
  - `TestBridge_GetDiagnostics{,_Empty}`: hardcoded `/tmp/test/*.go` replaced with `t.TempDir()` + `pathToURI()`
  - `TestCheckFrozenGuard_RejectsPathTraversal/절대_경로_거부`: added leading-`/` detection on Windows
- Lint cleanup: 10 `errcheck`/`staticcheck`/`unused` warnings across telemetry, hook, cli packages

### Closed Issues / PRs

- #636 (merged), #643 → closed by #660, #642 → closed by #661, #663 (merged)
- #662 (backlog) closed — 13/18 commits duplicate with #649, remaining 5 commits pending individual PRs

### 설치 및 업데이트

```bash
go install github.com/modu-ai/moai-adk/cmd/moai@v2.10.4
# 또는
moai update
```

---

## [2.10.4] - 2026-04-15 (한국어)

### 요약

3관점 리뷰 번들(CRITICAL 6 + IMPORTANT 8 + SUGGESTION 2 = 16건)을 재현-먼저 TDD로 일괄 수정: 파일 파괴 버그(apply.go:80 `MergeEvolvableZones` API 오용), path traversal 2건, OWASP 기반 `gopls`/`sg` 바이너리 allowlist, 비동기 텔레메트리 writer, Windows CI 크로스플랫폼 호환성, AskUserQuestion 전용 인터랙션 규칙 강화. 영향 패키지: `internal/evolution`, `internal/telemetry`, `internal/lsp/gopls`, `internal/astgrep`, `internal/merge`.

### Breaking Changes

없음. 모든 변경은 bug fix이며 API 삭제 없음.

### 추가 (요약)

- **보안/품질** (#636): path traversal 거부, 파일 파괴 버그 수정(`ReplaceEvolvableZone`), Agency 헌법 frozen 포함, AsyncRecorder + 파일 핸들 캐시, symlink 이름 검증, rate limit race 수정
- **gopls 브릿지** (#660): RFC 3986 URI 인코딩, per-URI pending map, 즉시 shutdown, timer 누수 제거, binary+args allowlist
- **astgrep** (#661): filterByLang 실제 구현, binary allowlist, context cancel defer, stderr 로깅, YAML 재질러 파싱
- **문서** (#663): `[HARD] AskUserQuestion 전용` + Socratic 인터뷰 라운드 설계 + Ambiguity Triggers

### 수정

- #636 CRITICAL 6건 + IMPORTANT 2건 (all TDD reproduction tests)
- #643 gopls 결함 5건, #642 astgrep 결함 5건
- Windows CI 경로 호환성 일괄 수정 (`filepath.ToSlash` 정규화, 백슬래시 메타문자 제외)
- Lint 경고 10건 정리 (errcheck/staticcheck/unused)

### 종료된 이슈/PR

- #636, #660, #661, #663 (모두 merge)
- #662 (백로그) close — 13/18 중복(#649), 남은 5건 개별 PR 예정

---

## [2.10.3] - 2026-04-14

### Summary

LSP SPEC suite rollout (5 completed SPECs: CORE-002, AGG-003, QGATE-004, LOOP-005, MULTI-006) + critical AC implementations (Feature Flag, `--json`/exit code) + SPEC status enum standardization across 28 SPECs. Statusline fixes: rate_limits skip prevents intermittent disappearance, GLM context-window override prevents 1M gauge mis-display. Quality gate: Flutter project detection, SARIF deterministic ordering, V2 gate tests. Routing: `--team` flag now correctly loads `team/*.md` workflows for plan/run/sync.

### Breaking Changes

None. All changes are additive or bug fixes. `lsp.client_impl: gopls_bridge` default maintains existing behavior.

### Added

#### LSP Infrastructure (SPEC-LSP-CORE-002, AGG-003, QGATE-004, LOOP-005, MULTI-006)

- **SPEC-LSP-CORE-002** — powernap-based multi-language LSP client foundation (6 sprints, 91-94% coverage)
  - `internal/lsp/core/`: Client, Manager, Document, Lifecycle State Machine, Capability Negotiation
  - `internal/lsp/subprocess/`: Supervisor, Launcher with graceful degradation
  - `internal/lsp/transport/`: JSON-RPC 2.0 codec via powernap
  - Manager: Lazy Spawn + Idle Reaper
  - **AC10 Feature Flag**: `lsp.client_impl` config key selects between `gopls_bridge` (SPEC-GOPLS-BRIDGE-001) and `powernap_core` (SPEC-LSP-CORE-002)
- **SPEC-LSP-AGG-003** — Diagnostic Aggregator + TTL Cache
  - `internal/lsp/cache/`: TTL cache with Get/Set/Invalidate (87.9% coverage)
  - `internal/lsp/aggregator/`: Parallel diagnostic collection with circuit breaker (96.8% coverage)
- **SPEC-LSP-QGATE-004** — Phase-aware LSP quality gates (plan/run/sync thresholds)
- **SPEC-LSP-LOOP-005** — Loop/Ralph LSP integration (Go-only)
  - `internal/ralph/`: Severity classification for LSP diagnostics
  - FeedbackChannel for PostTool → LoopController routing
  - Stagnation detection
- **SPEC-LSP-MULTI-006** — 16-language LSP server matrix with install hints + project_markers discovery
  - **AC4**: `moai lsp doctor --json` flag for machine-readable output
  - **AC4**: Non-zero exit code when required servers are missing

#### ast-grep Modernization (SPEC-ASTG-UPGRADE-001)

- `internal/astgrep/`: Full scanner package with SARIF output, rules loader, CLI wrapper
- `moai ast-grep` CLI command with `--json`, `--sarif`, `--text` output formats
- Pre-configured rule sets: security (OWASP/CWE), Go idioms, error handling, resource safety
- `RunAstGrepGateV2` quality gate integrated with Phase-aware LSP gates

#### gopls Bridge (SPEC-GOPLS-BRIDGE-001)

- `internal/lsp/gopls/`: Go-only subprocess bridge as fallback path
- Coexistence: `lsp.client_impl` config toggles between bridge and powernap core

#### Statusline Improvements

- **GLM context window override** (#653): 4-tier priority resolver
  - `MOAI_STATUSLINE_CONTEXT_SIZE` env → `llm.yaml glm.context_windows` map → built-in table → stdin fallback
  - Built-in defaults for 6 GLM models: glm-5.1, glm-5, glm-4.7, glm-4.6, glm-4.5, glm-4.5-air
  - Addresses 1M gauge mis-display in GLM mode (actual limit 128-230K)
- **`glm.context_windows` field** in `.moai/config/sections/llm.yaml` for per-model overrides
- `EnvStatuslineContextSize = "MOAI_STATUSLINE_CONTEXT_SIZE"` constant

#### Flutter Quality Gate Support (#652)

- `resolveDartFlutter()` + `isFlutterProject()` dynamic pubspec.yaml analysis
- Flutter SDK detection → `flutter test` / `flutter analyze`
- Pure Dart CLI projects retain `dart test` / `dart analyze` (no regression)

#### Other

- Simplify workflow Phase 0.05 integration into sync pipeline (SPEC-HOOKWAVE-001)
- Output styles v5.0.0: MoAI+R2D2 merge + Einstein tutor
- Orchestrator Self-Check §24 HARD rule in CLAUDE.local.md
- Delta Marker Detection + Complexity Estimator (SPEC-SDD-001)
- Evaluator Prompt Library 4 profiles (SPEC-EVALLIB-001)
- plan-auditor agent for bias-prevention SPEC reviews

### Changed

- **SPEC status enum standardized** across 28 SPECs (42% of SPEC inventory)
  - `Completed` → `completed`, `Draft` → `draft`, `Planned` → `planned`, `Superseded` → `superseded`
  - `Implemented`/`implemented` → `completed`, `approved` → `planned`
- SPEC-LSP-001 marked `superseded` (superseded by SPEC-LSP-CORE-002)
- `internal/statusline/builder.go` `collectAll()`: skip usageProvider when stdin has `rate_limits` (prevents 5s HTTP timeout → statusline disappearance)
- SARIF `tool.driver.rules` array sorted deterministically (#644 fix)
- Hook event naming: `Stop` → `SubagentStop` across 14 agent definitions
- `internal/hook/quality/gate.go` Dart/Flutter detection: now inspects pubspec.yaml content
- Quality gate V2 (`RunAstGrepGateV2`) gains unit test coverage (#645)

### Fixed

- **#626**: `--team` flag now correctly loads `team/*.md` workflows for `/moai run|plan|sync|fix|mx`. Previously only `review` subcommand honored team routing
- **#644**: SARIF `tool.driver.rules` array had non-deterministic order due to Go map iteration; now `sort.Slice` by rule ID
- **#645**: `RunAstGrepGateV2` quality gate had no unit tests; added table-driven test coverage
- **#646**: Statusline intermittent disappearance caused by blocking Anthropic OAuth API calls (5s timeout on cache miss). Fix: skip usage collector when Claude Code (2.1.80+) already provides `rate_limits` via stdin
- **#652**: Quality gate used `dart test` for Flutter projects, causing commit blocks. Fix: detect Flutter SDK in pubspec.yaml and use `flutter test` + `flutter analyze`
- **#653**: Statusline context gauge showed 1M for GLM models that actually have 128-230K limit. Fix: 4-tier priority resolver with GLM model detection
- `/login`, `/logout` hang in hook system (flat camelCase support + graceful validation)

### Removed

- None

### Security

- ast-grep security rules for OWASP/CWE patterns integrated into quality gate

### Installation & Update

```bash
go install github.com/modu-ai/moai-adk/cmd/moai@v2.10.3
# or
moai update
```

### Pull Requests Merged

- #648 fix(statusline): skip usage collector when stdin rate_limits present
- #649 sync(v2.10.3): LSP suite + AC10/AC4 critical fixes + SPEC enum standardization
- #650 fix(astgrep): SARIF tool.driver.rules 결정적 순서 보장
- #651 test(quality): add unit tests for RunAstGrepGateV2 quality gate
- #654 fix(routing): --team 플래그가 team/*.md 워크플로우를 로드하도록 수정
- #655 fix(statusline): GLM 모드 컨텍스트 게이지 1M 오표시
- #656 fix(quality): Flutter 프로젝트에서 dart test 대신 flutter test 사용

### Closed Issues

#626, #641 (meta), #644, #645, #646, #652, #653

---

## [2.10.3] - 2026-04-14 (한국어)

### 요약

LSP SPEC suite 출시 (완료 SPEC 5개: CORE-002, AGG-003, QGATE-004, LOOP-005, MULTI-006) + 핵심 AC 구현(Feature Flag, `--json`/exit code) + 28개 SPEC status enum 표준화. Statusline 수정: rate_limits skip으로 간헐적 사라짐 방지, GLM context window override로 1M 게이지 오표시 차단. Quality gate: Flutter 프로젝트 감지, SARIF 결정적 정렬, V2 gate 테스트. 라우팅: `--team` 플래그가 plan/run/sync에서도 `team/*.md` 워크플로우를 올바르게 로드.

### Breaking Changes

없음. 모든 변경은 additive 또는 bug fix. `lsp.client_impl: gopls_bridge` 기본값으로 기존 동작 유지.

### 추가

#### LSP 인프라 (SPEC-LSP-CORE-002, AGG-003, QGATE-004, LOOP-005, MULTI-006)

- **SPEC-LSP-CORE-002** — powernap 기반 다국어 LSP 클라이언트 기반 (6 sprints, 91-94% 커버리지)
- **SPEC-LSP-AGG-003** — Diagnostic Aggregator + TTL Cache + Circuit Breaker
- **SPEC-LSP-QGATE-004** — Phase-aware LSP 품질 게이트 (plan/run/sync 단계별 임계값)
- **SPEC-LSP-LOOP-005** — Loop/Ralph LSP 통합 (Go 전용)
- **SPEC-LSP-MULTI-006** — 16개 언어 LSP 서버 매트릭스 + install hints + project_markers
  - **AC4**: `moai lsp doctor --json` 플래그 + missing server 시 exit 1

#### ast-grep 현대화 (SPEC-ASTG-UPGRADE-001)

- `internal/astgrep/` 패키지: scanner, SARIF 출력, rules loader, CLI
- `moai ast-grep` 커맨드 (`--json`, `--sarif`, `--text` 출력)
- Pre-configured 규칙: security (OWASP/CWE), Go idioms, error handling, resource safety

#### GLM statusline context (#653)

- 4-tier priority 리졸버: env → `llm.yaml glm.context_windows` → 내장 테이블 → stdin fallback
- 6개 GLM 모델 내장 기본값 (glm-5.1, glm-5, glm-4.7, glm-4.6, glm-4.5, glm-4.5-air)
- `MOAI_STATUSLINE_CONTEXT_SIZE` 환경변수 지원

#### Flutter 품질 게이트 지원 (#652)

- pubspec.yaml 동적 분석으로 Flutter SDK 감지
- Flutter 프로젝트 → `flutter test` / `flutter analyze`
- 순수 Dart CLI → 기존 `dart test` / `dart analyze` 유지

#### 기타

- Simplify workflow Phase 0.05 sync 통합 (SPEC-HOOKWAVE-001)
- Output styles v5.0.0: MoAI+R2D2 통합 + Einstein 튜터
- Orchestrator Self-Check §24 HARD 규칙
- plan-auditor 에이전트 (SPEC 편향방지 검수)

### 변경

- **28개 SPEC enum 표준화** (전체 SPEC의 42%): 대소문자/비표준 값 소문자 표준으로
- SPEC-LSP-001을 `superseded`로 표시 (SPEC-LSP-CORE-002가 대체)
- statusline `collectAll()`: stdin에 `rate_limits` 있으면 usageProvider 스킵
- SARIF rules 배열 ID 기준 정렬
- Hook 이벤트명: `Stop` → `SubagentStop` (14개 에이전트)

### 수정

- **#626**: `--team` 플래그가 `/moai run|plan|sync|fix|mx`에서 `team/*.md` 로드하도록 수정
- **#644**: SARIF rules 배열 비결정적 순서
- **#645**: `RunAstGrepGateV2` 단위 테스트 공백
- **#646**: Anthropic OAuth API 블로킹(5초 timeout)으로 statusline 간헐 사라짐
- **#652**: Flutter 프로젝트에서 `dart test` 사용으로 commit 차단
- **#653**: GLM 모드 statusline 게이지가 Claude 1M 기준으로 오표시
- `/login`, `/logout` hook hang (flat camelCase 지원)

### 설치 및 업데이트

```bash
go install github.com/modu-ai/moai-adk/cmd/moai@v2.10.3
# 또는
moai update
```

### 머지된 PR

#648, #649, #650, #651, #654, #655, #656

### 종료된 이슈

#626, #641 (meta), #644, #645, #646, #652, #653

---

## [2.10.2] - 2026-04-11

### Summary

Quality gate hardening for multi-language projects: linter config detection prevents false commit blocks. Hook binary discovery expanded for Linux environments. All Korean code comments converted to English.

### Breaking Changes

None

### Added

- `configFiles` field on `gateStep` struct for linter config file detection before execution
- Config file checks for all 15 supported language linters (Go, Node.js, Python, Rust, Java, Kotlin, C#, Ruby, PHP, Swift, Dart, Elixir, Scala, Haskell, Zig)
- `$HOME/.local/bin/moai` fallback path in all 27 hook wrapper templates for Linux installs
- Self-Learning Quality Guard system with ast-grep integration (SPEC-SLQG-001)
- `AstGrepGate` configuration in `GateConfig` for domain-specific pattern scanning
- 4 reproduction tests for configFiles behavior

### Changed

- Extracted 12 hardcoded constants into `internal/config/envkeys.go` and `internal/config/defaults.go`
- Windows platform-aware Go binary path fallback in `update.go`
- `code_comments` setting changed from `ko` to `en` in language.yaml
- Converted ~150 Korean code comments to English across 24 Go source files

### Fixed

- ESLint quality gate blocking git commits in Python-only projects (#619)
- Hook wrapper scripts not finding moai binary at `~/.local/bin/moai` on Linux
- Korean comment in `settings.go` violating English-only code standard

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.10.2] - 2026-04-11 (한국어)

### 요약

다국어 프로젝트를 위한 품질 게이트 강화: 린터 설정 파일 감지로 잘못된 커밋 차단 방지. Linux 환경 hook 바이너리 탐색 확장. 모든 한국어 코드 주석 영어 변환.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- `gateStep` 구조체에 `configFiles` 필드 추가 — 린터 실행 전 설정 파일 존재 여부 확인
- 15개 지원 언어 린터 전체에 설정 파일 검사 적용
- 27개 hook wrapper 템플릿에 `$HOME/.local/bin/moai` 폴백 경로 추가
- Self-Learning Quality Guard 시스템 + ast-grep 통합 (SPEC-SLQG-001)
- `GateConfig`에 `AstGrepGate` 설정 추가
- configFiles 동작 재현 테스트 4건

### 변경됨 (Changed)

- 하드코딩 상수 12건을 `internal/config/` 패키지로 추출
- `update.go`에 Windows 플랫폼별 Go 바이너리 경로 폴백 추가
- `language.yaml`의 `code_comments` 설정을 `ko`에서 `en`으로 변경
- 24개 Go 소스 파일의 한국어 주석 ~150건 영어 변환

### 수정됨 (Fixed)

- Python-only 프로젝트에서 ESLint 품질 게이트가 git commit을 차단하던 버그 (#619)
- Linux에서 `~/.local/bin/moai` 경로의 moai 바이너리를 찾지 못하던 hook wrapper 문제
- `settings.go`의 한국어 주석 — 영어 전용 코드 표준 위반

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.10.1] - 2026-04-09

### Summary

Hooks protocol upgrade: Claude Code hooks specification compliance, new handler behaviors, and expanded input/output field coverage. Fixes UserPromptSubmit hook error and SPEC-ID title duplication.

### Breaking Changes

None

### Added

- `hookEventName` now included in all `hookSpecificOutput` responses (Claude Code protocol compliance)
- `sessionTitle` field in UserPromptSubmit output for Claude Code UI session title integration
- `UpdatedInput`, `UpdatedMCPToolOutput` fields in `HookSpecificOutput` for tool input modification and MCP output replacement
- 16 new `HookInput` fields: `error_type`, `error_message`, `last_assistant_message`, `configuration_source`, `change_type`, `old_cwd`, `new_cwd`, `mcp_tool_name`, `elicitation_request`, `instruction_file_path`, `memory_type`, `load_reason`, `permission_suggestions`, etc.
- `PermissionRequest` event registered in `settings.json` with shell wrapper script
- `PermissionDenied` auto-retry for read-only tools (Read, Grep, Glob, WebFetch, WebSearch, Skill)
- `StopFailure` error_type-based `SystemMessage` responses (rate_limit, authentication_failed, billing_error, max_output_tokens)
- `PostCompact` session memo restoration via `SystemMessage` injection
- `SubagentStart` project context injection via `additionalContext` (project name, type, language, active SPEC)
- `CwdChanged` handler with `CLAUDE_ENV_FILE` support for environment variable persistence
- Matchers added to 6 events: SessionStart (`startup|resume`), PreCompact/PostCompact (`manual|auto`), StopFailure, ConfigChange (`project_settings|local_settings|skills`)
- QualityGate universalized: Go, Node.js (eslint+npm test), Python (ruff+pytest), Rust (clippy+cargo test)

### Fixed

- UserPromptSubmit hook error caused by missing `hookEventName` in `hookSpecificOutput`
- SPEC-ID duplication in session title (e.g., "SPEC-SRS-003: SPEC-SRS-003: Dashboard..." → "SPEC-SRS-003: Dashboard...")
- camelCase field normalization extended with 15 new field mappings

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.10.1] - 2026-04-09 (한국어)

### 요약

Hooks 프로토콜 업그레이드: Claude Code hooks 스펙 준수, 새로운 핸들러 동작, 입출력 필드 확장. UserPromptSubmit hook 오류 및 SPEC-ID 타이틀 중복 수정.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- 모든 `hookSpecificOutput`에 `hookEventName` 포함 (Claude Code 프로토콜 준수)
- UserPromptSubmit 출력에 `sessionTitle` 필드 추가 (Claude Code UI 세션 타이틀 통합)
- `HookSpecificOutput`에 `UpdatedInput`, `UpdatedMCPToolOutput` 필드 추가
- `HookInput`에 16개 새 필드 추가: `error_type`, `error_message`, `last_assistant_message`, `configuration_source`, `change_type`, `old_cwd`, `new_cwd` 등
- `PermissionRequest` 이벤트를 `settings.json`에 등록 + 셸 래퍼 스크립트 생성
- `PermissionDenied` 읽기 전용 도구 자동 재시도 (Read, Grep, Glob, WebFetch, WebSearch, Skill)
- `StopFailure` error_type별 `SystemMessage` 응답 (rate_limit, authentication_failed, billing_error, max_output_tokens)
- `PostCompact` 세션 메모 복원 (`SystemMessage` 주입)
- `SubagentStart` 프로젝트 컨텍스트 자동 주입 (`additionalContext`)
- `CwdChanged` 핸들러에 `CLAUDE_ENV_FILE` 지원 추가
- 6개 이벤트에 matcher 추가: SessionStart, PreCompact/PostCompact, StopFailure, ConfigChange
- QualityGate 범용화: Go, Node.js, Python, Rust 지원

### 수정됨 (Fixed)

- `hookSpecificOutput`에서 `hookEventName` 누락으로 인한 UserPromptSubmit hook 오류 수정
- 세션 타이틀 SPEC-ID 중복 수정 (예: "SPEC-SRS-003: SPEC-SRS-003: Dashboard..." → "SPEC-SRS-003: Dashboard...")
- camelCase 필드 정규화에 15개 새 매핑 추가

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.10.0] - 2026-04-09

### Summary

Major optimization release: 88% reduction in agent definition sizes, 16 language skills converted to path-based rules, Self-Research System, and Claude Code 2.1.97 feature adoption. Template token consumption reduced by ~106K lines while preserving 100% workflow functionality.

### Breaking Changes

None

### Added

- Self-Research System for autonomous experimentation (SPEC-SRS-001~003)
- Claude Code 2.1.97 feature adoption: `permissionMode` strengthening, worktree CWD isolation fix, SubagentStop hook stability (SPEC-CC297-001)
- Background Agent Write restriction rule across 4 governance documents
- Agency section enhanced in README with Mermaid diagrams: /moai vs /agency comparison, GAN Loop sequence, Self-Evolution lifecycle

### Changed

- 20 MoAI agent definitions reduced from ~700 to ~120 lines average (88% reduction) with all workflow steps preserved
- 16 language skills converted to `paths:`-based rules for deterministic auto-loading (57→41 skills)
- `moai-foundation-claude` renamed to `moai-foundation-cc` (official naming convention compliance)
- Common agent protocol extracted to shared rule file (`agent-common-protocol.md`)
- `maxTurns` removed from all 22 agents (deprecated since v2.1.69)
- `manager-strategy`, `manager-quality`: permissionMode aligned to `plan` (read-only)

### Fixed

- `--team` tmux mode: `teammateMode` setting unified to native key (#605)

### Internal

- Template optimization: 314 files changed, -112K lines, +6K lines across agents, skills, and rules

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.10.0] - 2026-04-09 (한국어)

### 요약

대규모 최적화 릴리즈: 에이전트 정의 88% 축소, 16개 언어 스킬을 경로 기반 규칙으로 전환, Self-Research System, Claude Code 2.1.97 기능 채택. 워크플로우 기능 100% 보존하면서 템플릿 토큰 소비 ~106K줄 절감.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- 자율 실험 시스템 Self-Research System (SPEC-SRS-001~003)
- Claude Code 2.1.97 기능 채택: `permissionMode` 강화, worktree CWD 격리 수정, SubagentStop 안정성 (SPEC-CC297-001)
- Background Agent Write 제한 규칙 4개 거버넌스 문서에 추가
- README Agency 섹션 Mermaid 다이어그램 보강: /moai vs /agency 비교, GAN Loop 시퀀스, Self-Evolution 생명주기

### 변경됨 (Changed)

- 20개 MoAI 에이전트 정의: 평균 ~700줄 → ~120줄 (88% 축소), 워크플로우 단계 100% 보존
- 16개 언어 스킬을 `paths:` 기반 규칙으로 전환 (57→41 스킬, 28% 감소)
- `moai-foundation-claude` → `moai-foundation-cc` 이름 변경 (공식 명명 규약 준수)
- 공통 에이전트 프로토콜을 공유 규칙 파일로 추출 (`agent-common-protocol.md`)
- 모든 22개 에이전트에서 `maxTurns` 제거 (v2.1.69 이후 deprecated)
- `manager-strategy`, `manager-quality`: permissionMode를 `plan`으로 정렬 (read-only)

### 수정됨 (Fixed)

- `--team` tmux 모드: `teammateMode` 설정을 네이티브 키로 통일 (#605)

### 내부 변경

- 템플릿 최적화: 314개 파일 변경, -112K줄, +6K줄 (에이전트, 스킬, 규칙 전반)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.9.1] - 2026-04-03

### Summary

Patch release fixing Agency v3.2 template deployment. All Agency files (agents, skills, commands, rules, config) are now correctly included in `moai update` distribution. Also includes CLAUDE.md v14.0.0 with Agency and Harness documentation, and README updates for all 4 languages.

### Breaking Changes

None

### Fixed

- Agency template deployment: 35 Agency files (agents, skills, commands, rules, .agency/ config) were missing from `internal/template/templates/` and not distributed via `moai update`

### Changed

- CLAUDE.md v14.0.0: Added /agency command reference, Agency agent catalog, Harness-Based Quality Routing, and Agency configuration section
- README updated in 4 languages (en, ko, ja, zh) with Agency v3.2 details: GAN Loop, 5 specialized skills, 5-layer safety architecture, knowledge graduation, all 11 subcommands

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.9.1] - 2026-04-03 (한국어)

### 요약

Agency v3.2 템플릿 배포 누락 수정 패치. 모든 Agency 파일(에이전트, 스킬, 커맨드, 규칙, 설정)이 이제 `moai update`를 통해 올바르게 배포됩니다. CLAUDE.md v14.0.0 (Agency/Harness 문서) 및 4개 언어 README 업데이트 포함.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- Agency 템플릿 배포: 35개 Agency 파일(에이전트, 스킬, 커맨드, 규칙, .agency/ 설정)이 `internal/template/templates/`에 누락되어 `moai update`로 배포되지 않던 문제 수정

### 변경됨 (Changed)

- CLAUDE.md v14.0.0: /agency 커맨드 레퍼런스, Agency 에이전트 카탈로그, Harness 품질 라우팅, Agency 설정 섹션 추가
- README 4개 언어(en, ko, ja, zh) Agency v3.2 상세 업데이트: GAN Loop, 5개 전문 스킬, 5계층 안전 아키텍처, 지식 졸업, 전체 11개 서브커맨드

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.9.0] - 2026-04-03

### Summary

Major release: Claude Code v2.1.89-90 compatibility upgrade, AI Agency self-evolution system v3.2 with GAN Loop, harness design integration, and SPEC workflow enhancements. Includes bug fixes for PATH preservation in non-moai directories and 1M context model selection.

### Breaking Changes

- manager-quality model upgraded from haiku to sonnet (increased cost, improved quality)

### Added

- **AI Agency v3.2**: Self-evolving creative production system with GAN Loop (Builder-Evaluator), 5-layer safety architecture, knowledge graduation protocol, and fork management
- **sync-auditor agent**: Independent skeptical quality evaluator with 4-dimension scoring (Functionality/Security/Craft/Consistency)
- **harness.yaml**: 3-level quality depth (minimal/standard/thorough) with auto-detection and escalation
- **constitution.yaml**: Machine-readable project technical constraints
- **evaluator-profiles/**: 4 evaluator profiles (default, strict, lenient, frontend)
- **Agency skill router**: Subcommand pattern (brief, build, review, learn, evolve, resume, profile) with `agency:*` skill routing
- **Complexity Estimator**: Automatic harness level determination based on SPEC complexity
- **Delta Markers**: [EXISTING]/[MODIFY]/[NEW]/[REMOVE] classification for brownfield projects
- **spec-compact.md**: ~30% token savings in Run phase
- **Drift Guard**: Real-time scope drift detection on DDD/TDD cycle completion
- **GitHub workflow skill**: Issue management and PR review automation

### Changed

- Claude Code v2.1.89-90 compatibility: permission-denied hook, updated settings template
- manager-quality: model haiku to sonnet, skeptical evaluation prompts
- manager-spec: What/Why boundary enforcement + Exclusions validation
- SPEC workflow: Harness routing, Phase 2.0 sprint contracts, tasks.md persistence
- Agency refactored from monolithic skill to subcommand-based routing pattern

### Fixed

- PATH preservation in non-moai directories during update (#598, #599)
- 1M context model selection now correctly passes chosen model to Claude Code (#597)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.9.0] - 2026-04-03 (한국어)

### 요약

메이저 릴리즈: Claude Code v2.1.89-90 호환성 업그레이드, AI Agency 자기진화 시스템 v3.2 (GAN Loop 포함), 하네스 설계 통합, SPEC 워크플로우 강화. 비-moai 디렉토리 PATH 유실 및 1M 컨텍스트 모델 선택 버그 수정 포함.

### 주요 변경 사항 (Breaking Changes)

- manager-quality 모델이 haiku에서 sonnet으로 업그레이드 (비용 증가, 품질 향상)

### 추가됨 (Added)

- **AI Agency v3.2**: GAN Loop(Builder-Evaluator), 5계층 안전 아키텍처, 지식 졸업 프로토콜, 포크 관리를 갖춘 자기진화 창작 생산 시스템
- **sync-auditor 에이전트**: 독립적 회의적 품질 평가자. 4차원 평가 (기능성/보안/완성도/일관성)
- **harness.yaml**: 3단계 품질 깊이 (minimal/standard/thorough) 자동 판단 + 실패 시 에스컬레이션
- **constitution.yaml**: 프로젝트 기술 제약 기계 판독 정의
- **evaluator-profiles/**: 4종 평가자 프로필 (default, strict, lenient, frontend)
- **Agency 스킬 라우터**: 서브커맨드 패턴 (brief, build, review, learn, evolve, resume, profile) + `agency:*` 스킬 라우팅
- **Complexity Estimator**: SPEC 복잡도 기반 하네스 레벨 자동 결정
- **Delta Markers**: 브라운필드 프로젝트용 [EXISTING]/[MODIFY]/[NEW]/[REMOVE] 분류
- **spec-compact.md**: Run phase ~30% 토큰 절약
- **Drift Guard**: DDD/TDD 사이클 완료 시 scope drift 실시간 감지
- **GitHub 워크플로우 스킬**: 이슈 관리 및 PR 리뷰 자동화

### 변경됨 (Changed)

- Claude Code v2.1.89-90 호환성: permission-denied 훅, 설정 템플릿 업데이트
- manager-quality: 모델 haiku → sonnet, 회의적 평가 프롬프트
- manager-spec: What/Why 경계 + Exclusions 검증 강화
- SPEC 워크플로우: 하네스 라우팅, Phase 2.0 스프린트 계약, tasks.md 영속화
- Agency가 모놀리식 스킬에서 서브커맨드 기반 라우팅 패턴으로 리팩토링

### 수정됨 (Fixed)

- 비-moai 디렉토리에서 업데이트 시 PATH 유실 방지 (#598, #599)
- 1M 컨텍스트 모델 선택이 Claude Code에 올바르게 전달되도록 수정 (#597)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

## [2.8.4] - 2026-04-01

### Summary

Fixed profile model selection where 1M context models were not being correctly passed to Claude Code, ensuring user's model choice is properly applied.

### Breaking Changes

None

### Added

None

### Changed

None

### Fixed

- Profile: 1M context model selection now correctly passes the chosen model to Claude Code instead of using the default model

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.8.4] - 2026-04-01 (한국어)

### 요약

프로필에서 선택한 1M 컨텍스트 모델이 Claude Code에 올바르게 전달되지 않는 문제를 수정하여 사용자가 선택한 모델이 제대로 적용되도록 했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

없음

### 변경됨 (Changed)

없음

### 수정됨 (Fixed)

- 프로필: 1M 컨텍스트 모델 선택 시 기본 모델 대신 사용자가 선택한 모델이 Claude Code에 올바르게 전달되도록 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.8.3] - 2026-04-01

### Summary

GLM environment variable completeness fix (API_TIMEOUT_MS, DISABLE_NONESSENTIAL_TRAFFIC), `moai cc`/`moai cg` OAuth token preservation fix, statusline [1m] suffix stripping for GLM models, and CI/CD automation improvements with auto-merge on CI pass.

### Breaking Changes

None

### Added

- GLM env: `API_TIMEOUT_MS=3000000` and `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` added to all GLM injection paths (setGLMEnv, injectTmuxSessionEnv, buildGLMEnvVars, injectGLMEnvForTeam, injectGLMEnv)
- Statusline: `[1m]` suffix stripping for non-Claude/GLM model names in display and resolve logic
- CI: Auto-merge workflow that merges PRs when CI checks pass (#595)
- CI: CI/CD AI automation redesign with Claude Code Action integration (SPEC-CICD-001)

### Changed

- CI: All PRs now trigger CI regardless of file changes (#594)
- CI: Claude Code Action `max-turns` increased from 20 to 40
- Statusline: `resolveGLMModelName` now strips `[1m]` before Claude display name matching
- GLM team mode tests updated for new env var expectations (7 → 9 keys)

### Fixed

- GLM: Missing `API_TIMEOUT_MS` and `DISABLE_NONESSENTIAL_TRAFFIC` env vars caused timeouts and unnecessary network traffic on GLM/Z.AI proxy (#590)
- GLM: `[1m]` suffix appearing in statusline model display for GLM models
- GLM: `moai cc` and `moai cg` deleting OAuth token from tmux session env, causing auth failure after mode switch
- CI: PRs from forks and branch patterns not triggering CI workflows (#594)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.8.3] - 2026-04-01 (한국어)

### 요약

GLM 환경변수 누락 수정(API_TIMEOUT_MS, DISABLE_NONESSENTIAL_TRAFFIC), `moai cc`/`moai cg` OAuth 토큰 보존 수정, GLM 모델 상태표시줄 [1m] 접미사 제거, CI 통과 시 자동 머지 등 CI/CD 자동화 개선.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- GLM 환경변수: `API_TIMEOUT_MS=3000000`, `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` 모든 GLM 주입 경로에 추가
- 상태표시줄: 비-Claude/GLM 모델 이름에서 `[1m]` 접미사 자동 제거
- CI: CI 통과 시 PR 자동 머지 워크플로우 추가 (#595)
- CI: Claude Code Action 연동 CI/CD AI 자동화 재설계 (SPEC-CICD-001)

### 변경됨 (Changed)

- CI: 파일 변경 여부와 관계없이 모든 PR에서 CI 실행 (#594)
- CI: Claude Code Action `max-turns` 20→40 증가
- 상태표시줄: `resolveGLMModelName`이 Claude 표시 이름 매칭 전 `[1m]` 제거
- GLM 팀 모드 테스트: 환경변수 키 수 7→9 반영

### 수정됨 (Fixed)

- GLM: `API_TIMEOUT_MS`, `DISABLE_NONESSENTIAL_TRAFFIC` 누락으로 타임아웃 및 불필요한 네트워크 트래픽 발생 문제 수정 (#590)
- GLM: 상태표시줄에 GLM 모델 이름 뒤 `[1m]` 접미사 표시 문제 수정
- GLM: `moai cc`, `moai cg` 실행 시 tmux 세션에서 OAuth 토큰이 삭제되어 인증 실패하던 문제 수정
- CI: 포크 및 브랜치 패턴에서 CI 워크플로우 미실행 문제 수정 (#594)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.8.2] - 2026-03-31

### Summary

GLM compatibility automation with automatic env injection/removal, 1M context model support in profile wizard, and dynamic team generation replacing static agent files.

### Breaking Changes

None

### Added

- GLM compatibility automation: auto DISABLE_BETAS + DISABLE_PROMPT_CACHING injection in GLM mode, auto removal in Claude mode (#581)
- SessionStart hook for GLM env auto-detection and configuration
- 1M context model selection option in profile wizard (#578)
- Dynamic team generation: static team-* agent files replaced with runtime generation from workflow.yaml role profiles
- Code Review Quality Gate workflow with REVIEW.md template (#583)
- web-copy-craft module in design-craft skill v1.2.0
- Team configuration section in workflow.yaml

### Changed

- Profile wizard: removed auto permission mode option
- Statusline: GLM model name resolution from Claude display names
- CI: automerge, dependabot, and codecov workflow improvements
- Template: 6 agent definitions streamlined with reduced token footprint

### Fixed

- PostTool hook: improved GLM environment variable handling
- CI: automerge-action configuration fixes for reliable PR merging

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.8.2] - 2026-03-31 (한국어)

### 요약

GLM 호환성 자동화(환경변수 자동 주입/제거), 프로필 위자드 1M 컨텍스트 모델 지원, 정적 에이전트 파일 대신 동적 팀 생성 도입.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- GLM 호환성 자동화: GLM 모드에서 DISABLE_BETAS + DISABLE_PROMPT_CACHING 자동 주입, Claude 모드에서 자동 제거 (#581)
- SessionStart 훅으로 GLM 환경변수 자동 감지 및 설정
- 프로필 위자드에 1M 컨텍스트 모델 선택 옵션 추가 (#578)
- 동적 팀 생성: 정적 team-* 에이전트 파일을 workflow.yaml role_profiles 기반 런타임 생성으로 교체
- Code Review Quality Gate 워크플로우 및 REVIEW.md 템플릿 추가 (#583)
- design-craft 스킬 v1.2.0 web-copy-craft 모듈 추가
- workflow.yaml에 팀 설정 섹션 추가

### 변경됨 (Changed)

- 프로필 위자드에서 auto 권한 모드 옵션 제거
- Statusline: Claude 표시 이름에서 GLM 모델명 해석 기능
- CI: automerge, dependabot, codecov 워크플로우 개선
- 템플릿: 6개 에이전트 정의 최적화로 토큰 사용량 감소

### 수정됨 (Fixed)

- PostTool 훅: GLM 환경변수 처리 개선
- CI: automerge-action 설정 수정으로 안정적인 PR merge 수행

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.8.1] - 2026-03-30

### Summary

Feedback skill multilingual support and agent harness optimization across 16 agents with 5 new reference skills.

### Breaking Changes

None

### Added

- Feedback skill now writes GitHub issues in user's conversation_language (#573)
- 5 new reference skills: api-patterns, git-workflow, owasp-checklist, react-patterns, testing-pyramid

### Changed

- Agent harness optimization: 16 agent definitions streamlined with reduced token footprint (#576)
- Workflow skills (plan, run, sync) enhanced with improved agent coordination

### Fixed

- None

### Installation & Update

```bash
moai update
moai version
```

---

## [2.8.1] - 2026-03-30 (한국어)

### 요약

Feedback 스킬 다국어 지원 및 16개 에이전트 하네스 최적화, 5개 신규 참조 스킬 추가.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- Feedback 스킬이 conversation_language로 GitHub Issue 작성 (#573)
- 5개 신규 참조 스킬: api-patterns, git-workflow, owasp-checklist, react-patterns, testing-pyramid

### 변경됨 (Changed)

- 에이전트 하네스 최적화: 16개 에이전트 정의 간소화, 토큰 풋프린트 감소 (#576)
- 워크플로우 스킬 (plan, run, sync) 에이전트 조율 개선

### 수정됨 (Fixed)

- 없음

### 설치 및 업데이트

```bash
moai update
moai version
```

---

## [2.8.0] - 2026-03-30

### Summary

Claude Code v2.1.87 harness modernization: hook system completion (18 to 25 events), settings.json schema expansion, CLAUDE.md token optimization, license change to Apache-2.0, and claude-code-action update to v1.0.82.

### Breaking Changes

- License changed from GPL-3.0 to Apache-2.0

### Added

- Elicitation and ElicitationResult hook events for MCP server input interception (v2.1.76+)
- 7 new hook wrapper scripts (StopFailure, PostCompact, InstructionsLoaded, CwdChanged, FileChanged, Elicitation, ElicitationResult)
- 9 new permission tools in settings.json (CronCreate, CronList, CronDelete, EnterWorktree, ExitWorktree, EnterPlanMode, ExitPlanMode, ListMcpResourcesTool, ReadMcpResourceTool)
- `$schema` field for IDE autocomplete and validation
- `effortLevel`, `plansDirectory`, `includeGitInstructions` settings
- `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` security environment variable
- TaskCreated hook event in CLAUDE.md Team Hook Events section

### Changed

- CLAUDE.md optimized from 522 to 356 lines (-32%) for better Claude adherence per CC best practices
- claude-code-action updated to v1.0.82 with pull-requests write permission
- Template skill updates: shadcn/ui v3.0.0 (shadcn/cli v4), frontend design skills (GPT-5.4)
- Claude Code v2.1.86 audit improvements (rule frontmatter, agent names, skill format)
- Background agent kill shortcut updated: Ctrl+F to Ctrl+X Ctrl+K (v2.1.83)

### Fixed

- Profile bypass settings sync to settings.local.json
- hooks-system.md updated with complete 25-event list

### Installation & Update

```bash
moai update
moai version
```

---

## [2.8.0] - 2026-03-30 (한국어)

### 요약

Claude Code v2.1.87 harness 현대화: hook 시스템 완성 (18 → 25 이벤트), settings.json 스키마 확장, CLAUDE.md 토큰 최적화, Apache-2.0 라이선스 변경, claude-code-action v1.0.82 업데이트.

### 주요 변경 사항 (Breaking Changes)

- 라이선스 GPL-3.0에서 Apache-2.0으로 변경

### 추가됨 (Added)

- Elicitation, ElicitationResult hook 이벤트 (MCP 서버 입력 인터셉트, v2.1.76+)
- 7개 신규 hook wrapper script
- settings.json에 9개 permission 도구 추가
- `$schema`, `effortLevel`, `plansDirectory`, `includeGitInstructions` 설정
- `CLAUDE_CODE_SUBPROCESS_ENV_SCRUB` 보안 환경변수

### 변경됨 (Changed)

- CLAUDE.md 522줄 → 356줄 (-32%, CC best practices 준수)
- claude-code-action v1.0.82 업데이트
- 템플릿 스킬: shadcn/ui v3.0.0, 프론트엔드 디자인 스킬
- 백그라운드 에이전트 종료: Ctrl+F → Ctrl+X Ctrl+K (v2.1.83)

### 수정됨 (Fixed)

- Profile bypass 설정 동기화
- hooks-system.md 25개 이벤트 목록

### 설치 및 업데이트

```bash
moai update
moai version
```

---

## [2.7.22] - 2026-03-25

### Summary

Template and profile configuration fixes: sync workflow base branch resolution,
statusline compact mode cleanup in profile UI, and documentation consistency improvements.

### Breaking Changes

None

### Fixed

- fix(template): sync.md github_flow strategy base branch hardcoded to `main` — now uses `{main_branch}` from git-strategy.yaml
- fix(template): manual mode `main_branch` fallback logic clarified for agent interpretation
- fix(statusline): Remove misleading compact mode option from profile setup UI (was showing "2-line" but produced 3-line output identical to default)
- fix(preferences): Update StatuslineMode comment to reflect v3 actual values ("default", "full")
- fix(template): Add missing "compact" to statusline.yaml preset comment

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.22] - 2026-03-25 (한국어)

### 요약

템플릿 및 프로필 설정 정합성 수정: sync 워크플로우 base branch 동적 해결,
statusline compact 모드 프로필 UI 정리, 문서 정합성 개선.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- fix(template): sync.md github_flow 전략에서 base branch가 `main`으로 하드코딩 → `{main_branch}`로 동적 해결
- fix(template): manual 모드 `main_branch` 누락 시 fallback 로직 명시화
- fix(statusline): 프로필 설정에서 거짓 설명의 compact 모드 옵션 제거 ("2줄"이라 표시되었으나 실제 3줄 출력)
- fix(preferences): StatuslineMode 주석을 v3 실제값("default", "full")으로 수정
- fix(template): statusline.yaml preset 주석에 누락된 "compact" 추가

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.21] - 2026-03-23

### Summary

Design-first development workflow integration with interface-design plugin concepts,
dependency updates (glamour 1.0.0, huh 1.0.0, x/text 0.35.0), and statusline/hook fixes.

### Breaking Changes

None

### Added

- Intent-First design craft skill (`moai-design-craft`) with design direction, domain vocabulary, design memory, and post-build critique modules
- Design direction phase in `/moai plan` for UI/UX SPEC workflows
- Design fidelity verification step in `/moai review` workflow
- System memory-aware test execution workflow in `/moai run`

### Changed

- Bump `github.com/charmbracelet/glamour` from 0.10.0 to 1.0.0
- Bump `github.com/charmbracelet/huh` from 0.8.0 to 1.0.0
- Bump `golang.org/x/text` from 0.34.0 to 0.35.0
- Template and skill definition updates for design workflow integration

### Fixed

- Statusline stdin JSON parsing failure and improved `rate_limits` field utilization (#553)
- PostToolUse hook matcher being ignored and `session_id` missing from hook payload (#544)
- Ambiguous 'WorktreeManager' reference in plan.md replaced with `moai worktree new` (#554)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.21] - 2026-03-23 (한국어)

### 요약

interface-design 플러그인 개념을 활용한 디자인 우선 개발 워크플로우 통합,
의존성 업데이트 (glamour 1.0.0, huh 1.0.0, x/text 0.35.0), statusline/hook 버그 수정.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- Intent-First 디자인 크래프트 스킬 (`moai-design-craft`) - 디자인 방향, 도메인 어휘, 디자인 메모리, 빌드 후 비평 모듈
- `/moai plan`에 UI/UX SPEC 디자인 방향 설정 단계 추가
- `/moai review`에 디자인 충실도 검증 단계 추가
- `/moai run`에 시스템 메모리 인식 테스트 실행 워크플로우 추가

### 변경됨 (Changed)

- `github.com/charmbracelet/glamour` 0.10.0 → 1.0.0 업그레이드
- `github.com/charmbracelet/huh` 0.8.0 → 1.0.0 업그레이드
- `golang.org/x/text` 0.34.0 → 0.35.0 업그레이드
- 디자인 워크플로우 통합을 위한 템플릿 및 스킬 정의 업데이트

### 수정됨 (Fixed)

- Statusline stdin JSON 파싱 실패 및 `rate_limits` 필드 활용 개선 (#553)
- PostToolUse 훅 matcher 무시 및 `session_id` 누락 오류 수정 (#544)
- plan.md의 모호한 'WorktreeManager' 참조를 `moai worktree new`로 변경 (#554)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.20] - 2026-03-20

### Summary

Major cleanup release: complete removal of the `moai rank` leaderboard feature, plus integration of Claude Code v2.1.80 new capabilities including 3 new hook events, rate limit statusline support, and effort-based skill optimization.

### Breaking Changes

- `moai rank` command and all subcommands (login, status, logout, sync, exclude, include, register) have been removed
- `internal/rank/` package completely deleted (17 files)
- `Dependencies` struct no longer contains `RankClient`, `RankCredStore`, `RankBrowser` fields

### Added

- Hook events: `PostCompact` (v2.1.76), `InstructionsLoaded` (v2.1.69), `StopFailure` (v2.1.78) with handlers and CLI subcommands
- Statusline: `rate_limits` field parsing from Claude Code v2.1.80 JSON input
- Statusline: Rate limit data from Claude Code prioritized over MoAI API calls for 5H/7D usage bars
- Template: `effort` frontmatter support for skills (thinking: high, philosopher: high, loop: low)
- Template: `worktree` section with `sparse_paths` configuration in workflow.yaml

### Changed

- Hook subcommand count increased from 19 to 22 (3 new events)
- Template skill definitions updated with effort-based optimization

### Fixed

- Removed dead code from `moai rank` feature (~8,900 lines deleted)

### Installation & Update

```bash
moai update
moai version
```

---

## [2.7.20] - 2026-03-20 (한국어)

### 요약

대규모 정리 릴리즈: `moai rank` 리더보드 기능 완전 제거, Claude Code v2.1.80 신기능 통합 (신규 훅 이벤트 3종, rate limit statusline 지원, effort 기반 스킬 최적화).

### 주요 변경 사항 (Breaking Changes)

- `moai rank` 명령어 및 모든 서브커맨드 제거
- `internal/rank/` 패키지 완전 삭제 (17개 파일)
- `Dependencies` 구조체에서 Rank 관련 필드 제거

### 추가됨 (Added)

- 훅 이벤트: `PostCompact`, `InstructionsLoaded`, `StopFailure` 핸들러 및 CLI 서브커맨드
- Statusline: Claude Code v2.1.80 `rate_limits` 필드 파싱 및 우선 사용
- 템플릿: 스킬 `effort` frontmatter 지원, workflow.yaml worktree/sparse_paths 설정

### 변경됨 (Changed)

- 훅 서브커맨드 수 19 → 22

### 수정됨 (Fixed)

- `moai rank` dead code 제거 (~8,900줄)

### 설치 및 업데이트

```bash
moai update
moai version
```

---

## [2.7.16] - 2026-03-18

### Summary

Two improvements: adds model policy selection to the `moai init` and `moai update -c` wizards (allowing users to choose High/Medium/Low tier during project setup), and fixes hardcoded macOS user home paths in generated hook and statusline scripts (replacing init-time absolute paths with runtime `$HOME` environment variable).

### Breaking Changes

None.

### Added

- `moai init` and `moai update -c` wizards now include a model policy selection step (High/Medium/Low) allowing users to configure Claude model tier during project initialization

### Fixed

- Generated hook scripts (`handle-*.sh`) and `status_line.sh` no longer embed the init-time absolute home path (`/Users/username/go/bin/moai`); replaced with runtime `$HOME/go/bin/moai` for portability across users and operating systems

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.16] - 2026-03-18 (한국어)

### 요약

두 가지 개선 사항: `moai init` 및 `moai update -c` 위저드에 모델 정책 선택 기능을 추가(프로젝트 초기화 시 High/Medium/Low 티어 선택 가능)하고, 생성된 훅/statusline 스크립트에 하드코딩된 macOS 사용자 홈 경로를 런타임 `$HOME` 환경변수로 교체하여 이식성을 개선했습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- `moai init` 및 `moai update -c` 위저드에 모델 정책 선택 단계 추가 (High/Medium/Low) — 프로젝트 초기화 시 Claude 모델 티어 직접 설정 가능

### 수정됨 (Fixed)

- 생성된 훅 스크립트(`handle-*.sh`) 및 `status_line.sh`에서 init 시점의 절대 경로(`/Users/username/go/bin/moai`) 제거; `$HOME/go/bin/moai` 런타임 환경변수로 교체하여 사용자/OS 간 이식성 개선

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.15] - 2026-03-16

### Summary

Two worktree-related bug fixes: automatically removes stale `{}` directories created by an unresolved agent memory path template in git worktrees, and reduces log noise by downgrading worktree fallback messages from INFO to DEBUG.

### Breaking Changes

None.

### Fixed

- `session_end` hook now automatically removes bogus `{}` directories at the project root caused by unresolved agent memory path templates in git worktrees
- Worktree fallback log messages (add, list, remove, prune, repair) downgraded from INFO to DEBUG to reduce log noise

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.15] - 2026-03-16 (한국어)

### 요약

워크트리 관련 버그 2건 수정: git 워크트리 환경에서 에이전트 메모리 경로 템플릿 미치환으로 생성되는 `{}` 디렉토리를 자동 정리하고, worktree fallback 로그 레벨을 INFO에서 DEBUG로 낮춰 로그 노이즈를 줄였습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- `session_end` 훅에서 git 워크트리 환경의 에이전트 메모리 경로 미치환으로 생성되는 `{}` 디렉토리를 자동 정리
- worktree fallback 로그 메시지(add, list, remove, prune, repair) 레벨을 INFO에서 DEBUG로 낮춰 로그 노이즈 감소

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.14] - 2026-03-13

### Summary

Fixed a critical bug where `moai update` would silently overwrite user-customized configuration files. The existing 3-way merge engine is now wired into the update flow, preserving user customizations for `.mcp.json`, `.claude/settings.json`, and `.moai/status_line.sh`.

### Breaking Changes

None.

### Fixed

- `moai update` no longer overwrites user-customized `.mcp.json`, `.claude/settings.json`, and `.moai/status_line.sh` — user changes are now preserved via 3-way merge
- `.moai/config/sections/*.yaml` files continue to use the existing merge-based restore (unchanged behavior, now documented)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.14] - 2026-03-13 (한국어)

### 요약

`moai update` 실행 시 사용자 설정 파일을 덮어쓰는 치명적 버그 수정. 기존 3-way merge 엔진을 update 흐름에 연결하여 `.mcp.json`, `.claude/settings.json`, `.moai/status_line.sh`의 사용자 커스터마이징이 보존됩니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- `moai update`가 `.mcp.json`, `.claude/settings.json`, `.moai/status_line.sh` 사용자 설정을 덮어쓰지 않음 — 3-way merge로 사용자 변경 사항 보존
- `.moai/config/sections/*.yaml` 파일은 기존 merge 기반 복원 방식 유지 (동작 변경 없음, 문서화)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.13] - 2026-03-13

### Summary

Documentation improvements: renamed `--ultrathink` flag to `--deepthink` across all agent/skill definitions, updated README with accurate CG mode usage and Multi-LLM sponsor info, and refined statusline FAQ.

### Breaking Changes

None.

### Changed

- Renamed `--ultrathink` flag to `--deepthink` across 52 agent, skill, and rules files
- Updated CG mode documentation: removed unnecessary manual `claude` step from usage instructions

### Fixed

- README: corrected statusline YAML example (`mode` → `preset`)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.13] - 2026-03-13 (한국어)

### 요약

문서 개선: 모든 에이전트/스킬 정의에서 `--ultrathink` 플래그를 `--deepthink`로 일괄 변경, README의 CG 모드 사용법 및 Multi-LLM 스폰서 정보 업데이트, statusline FAQ 정리.

### 주요 변경 사항 (Breaking Changes)

없음.

### 변경됨 (Changed)

- 52개 에이전트, 스킬, 규칙 파일에서 `--ultrathink` 플래그를 `--deepthink`로 일괄 변경
- CG 모드 문서 개선: 사용법에서 불필요한 수동 `claude` 실행 단계 제거

### 수정됨 (Fixed)

- README: statusline YAML 예시 오류 수정 (`mode` → `preset`)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.12] - 2026-03-12

### Summary

Fixed a language config loading bug that caused random language switches, resolved PR base branch misconfiguration when using non-main branches, and improved agent output visibility with structured completion reports.

### Breaking Changes

None

### Fixed

- **Language switching bug** (#511): Config section files (`language.yaml`, `user.yaml`, etc.) are now explicitly loaded via `@import` in 9 agents, preventing random language fallback to Korean
- **PR base branch** (#512): `manager-git` now reads `git-strategy.yaml` to determine the active mode and uses `git_strategy.{mode}.main_branch` as the PR target branch instead of hardcoding `main`

### Added

- **Structured completion reports**: All 9 implementation and manager agents now output a standardized `COMPLETION REPORT` block with files modified, tests run, deliverables, and next steps
- **Progress reporting guidelines**: `fix.md` and `run.md` workflows now include `HARD` rules for agent lifecycle visibility (Dispatch → Progress → Complete templates)
- **Agent Lifecycle Templates**: `moai.md` output-style updated with Agent Dispatch, Agent Progress, Agent Completion, Skill Activation, Parallel Execution Dashboard, and Workflow Progress templates

### Changed

- **project.md Phase 4**: Completion phase now reads generated docs and presents a structured content summary before asking next steps
- **statusline**: Default mode changed from `default` to `compact`
- Internal: ADR-011 passthrough token documentation added; README updated with `/moai github` subcommand

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.12] - 2026-03-12 (한국어)

### 요약

언어 설정이 랜덤하게 전환되는 버그를 수정했고, 비-main 브랜치 사용 시 PR base branch 오설정 문제를 해결했습니다. 에이전트 출력 가시성도 구조화된 완료 보고서로 개선되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- **언어 전환 버그** (#511): 9개 에이전트에서 config 섹션 파일(`language.yaml`, `user.yaml` 등)을 `@import`로 명시적 로딩하여 한국어로 랜덤 폴백되는 현상 해결
- **PR base branch 오설정** (#512): `manager-git`이 `git-strategy.yaml`을 읽어 활성 모드를 확인하고 `git_strategy.{mode}.main_branch`를 PR 타겟 브랜치로 사용 (기존 `main` 하드코딩 제거)

### 추가됨 (Added)

- **구조화된 완료 보고서**: 9개 구현/매니저 에이전트가 파일 수정 내역, 테스트 결과, 산출물, 다음 단계를 포함한 `COMPLETION REPORT` 블록 출력
- **진행 상황 보고 가이드라인**: `fix.md`, `run.md` 워크플로우에 에이전트 생명주기 가시성 HARD 규칙 추가 (Dispatch → Progress → Complete 템플릿)
- **에이전트 생명주기 템플릿**: `moai.md` 출력 스타일에 Agent Dispatch, Agent Progress, Agent Completion, Skill Activation, Parallel Execution Dashboard, Workflow Progress 템플릿 추가

### 변경됨 (Changed)

- **project.md Phase 4**: 완료 단계에서 생성된 문서를 읽어 구조화된 콘텐츠 요약을 표시한 후 다음 단계 질문
- **statusline**: 기본 모드 `default` → `compact`로 변경
- 내부: ADR-011 passthrough token 문서 추가; README에 `/moai github` 서브커맨드 설명 추가

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.11] - 2026-03-11

### Summary

Fixed template path resolution issues by adding `${CLAUDE_SKILL_DIR}` to the passthrough token list, ensuring skill workflows can reference other files using absolute paths. Also added GitHub automation commands and improved statusline compatibility.

### Breaking Changes

None

### Added

- **GitHub Workflow Commands**: New `/moai github` subcommand for GitHub issue and PR automation
- **Template Path Variables**: `${CLAUDE_SKILL_DIR}` now properly passthrough in template renderer

### Changed

- **sync.md workflow**: Added `--pr` flag to skip changelog prompt and auto-open PR
- **SKILL.md**: Updated all workflow references to use `${CLAUDE_SKILL_DIR}` absolute paths
- **github.md**: Converted to local-only command (not deployed to projects)

### Fixed

- **Template renderer**: `${CLAUDE_SKILL_DIR}` passthrough token was missing from whitelist
- **Bare relative paths**: Replaced `team/*.md` and `workflows/*.md` with `${CLAUDE_SKILL_DIR}/` paths
- **Statusline credentials**: Fixed `.credentials.json` (dot-prefix) reading on Linux/WSL2
- **WSL2 paths**: Preserve Windows backslash paths in WSL2 environment

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.11] - 2026-03-11 (한국어)

### 요약

`${CLAUDE_SKILL_DIR}`을 passthrough 토큰 목록에 추가하여 템플릿 경로 해결 문제를 수정했습니다. 또한 GitHub 자동화 커맨드를 추가하고 statusline 호환성을 개선했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **GitHub 워크플로우 커맨드**: GitHub 이슈/PR 자동화를 위한 `/moai github` 서브커맨드
- **템플릿 경로 변수**: 템플릿 렌더러에서 `${CLAUDE_SKILL_DIR}` passthrough 지원

### 변경됨 (Changed)
- **sync.md 워크플로우**: `--pr` 플래그 추가 (changelog 프롬프트 건너뛰고 PR 자동 열기)
- **SKILL.md**: 모든 워크플로우 참조를 `${CLAUDE_SKILL_DIR}` 절대경로로 변경
- **github.md**: 로컬 전용 커맨드로 변경 (프로젝트에 배포되지 않음)

### 수정됨 (Fixed)

- **템플릿 렌더러**: `${CLAUDE_SKILL_DIR}` passthrough 토큰 누락 수정
- **상대 경로**: `team/*.md` 및 `workflows/*.md`를 `${CLAUDE_SKILL_DIR}/` 경로로 변경
- **Statusline 자격증명**: Linux/WSL2에서 `.credentials.json` (점 프리픽스) 읽기 수정
- **WSL2 경로**: WSL2 환경에서 Windows 백슬래시 경로 보존

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.9] - 2026-03-11

### Summary

Statusline compact mode removed and consolidated into default mode. Usage bar display spacing improved. Internal template reference fixes and GLM model mapping updates.

### Breaking Changes

None. Existing `mode: compact` config is silently normalized to `default` for backward compatibility.

### Changed

- **statusline**: Compact mode removed; `mode: compact` in config now silently maps to `default` mode (3-line layout)
- **statusline**: Usage bar display spacing reduced by one space (cleaner output)

### Fixed

- **statusline**: `TestCollectUsage_RetriesAfterCooldownExpires` flaky test fixed by mocking keychain access

### Internal

- Template `@file` reference cleanup in rules/skills/agents
- GLM model mappings updated in config (glm-5, glm-4.7, glm-4.5-air)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.9] - 2026-03-11 (한국어)

### 요약

statusline compact 모드를 제거하고 default 모드로 통합했습니다. 사용량 바 표시 공백이 개선되었으며, 내부 템플릿 참조 오류 수정 및 GLM 모델 매핑이 업데이트되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음. 기존 설정의 `mode: compact`는 하위 호환성을 위해 `default` 모드로 자동 변환됩니다.

### 변경됨 (Changed)

- **statusline**: compact 모드 제거; 설정의 `mode: compact`는 `default` 모드(3줄 레이아웃)로 자동 변환
- **statusline**: 사용량 바 표시 공백 1칸 감소 (더 깔끔한 출력)

### 수정됨 (Fixed)

- **statusline**: `TestCollectUsage_RetriesAfterCooldownExpires` 불안정 테스트 수정 - 키체인 접근을 모킹으로 대체

### 내부 변경

- 템플릿 rules/skills/agents의 잘못된 `@file` 참조 정리
- config GLM 모델 매핑 업데이트 (glm-5, glm-4.7, glm-4.5-air)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---


## [2.7.8] - 2026-03-11

### Summary

Statusline v3 with RGB gradient, 5H/7D usage monitoring, and new multi-line layout. Worktree paths migrated to global `~/.moai/worktrees/`. Fixes 5H/7D usage always showing 0% (missing OAuth beta header), hook input capture, and GLM model override support.

### Breaking Changes

None. All changes are backward compatible.

### Added

- **Statusline v3**: RGB gradient coloring, 5H/7D API usage monitoring bars, redesigned multi-line layout (compact/default/full modes) (SPEC-SLV3-001)
- **Usage monitoring**: Real-time 5-hour and 7-day API utilization bars in statusline
- **Worktree global paths**: `moai worktree new` now creates worktrees at `~/.moai/worktrees/{ProjectName}/{branch}/` instead of inside the project (SPEC-WORKTREE-001)
- **Worktree automation workflow**: Automated worktree lifecycle management with tmux integration (SPEC-WORKTREE-002)
- **MX Tag Auto-Validation System**: Automatic @MX tag validation on file edits (SPEC-MX-002)
- **GLM model override**: Individual model name override support in GLM configuration

### Changed

- **Statusline themes**: Renamed from "Default/Catppuccin Mocha/Catppuccin Latte" to **"MoAI Dark/MoAI Light"** — two themes that match the statusline design language
- **Statusline wizard**: Removed segment preset selection UI from `moai init`/`moai update` wizard — configure segments directly in `.moai/config/sections/statusline.yaml`
- **Hook input normalization**: Unified hook stdin JSON parsing with robust normalization layer for cross-environment compatibility (#467, #474)

### Removed

- **`/moai context` command**: Git-based context retrieval workflow removed — context is now managed through Claude Code's built-in auto-memory and structured commit messages

### Fixed

- **5H/7D usage always showing 0%**: Messages API requires `anthropic-beta: oauth-2025-04-20` header when using OAuth tokens — previously missing, causing 401 errors and empty usage bars
- **Hook input capture**: Fixed hook stdin JSON not being properly captured in some environments (#467)
- **PATH capture in hooks**: Fixed PATH environment variable not being captured correctly (#474)
- **Windows test**: Fixed git path separator mismatch in `cleanupMoaiWorktrees` tests

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.8] - 2026-03-11 (한국어)

### 요약

RGB 그라디언트와 5H/7D 사용량 모니터링을 포함한 Statusline v3 출시. Worktree 경로를 글로벌 `~/.moai/worktrees/`로 이동. 5H/7D 사용량이 항상 0%로 표시되던 문제(OAuth beta 헤더 누락), 훅 입력 캡처, GLM 모델 오버라이드 지원 수정.

### 주요 변경 사항 (Breaking Changes)

없음. 모든 변경사항은 하위 호환성을 유지한다.

### 추가됨 (Added)

- **Statusline v3**: RGB 그라디언트 색상, 5H/7D API 사용량 모니터링 바, 다중 라인 레이아웃 재설계 (compact/default/full 모드) (SPEC-SLV3-001)
- **사용량 모니터링**: statusline에 5시간/7일 API 사용률 바 표시
- **Worktree 글로벌 경로**: `moai worktree new`가 프로젝트 내부 대신 `~/.moai/worktrees/{ProjectName}/{branch}/`에 worktree 생성 (SPEC-WORKTREE-001)
- **Worktree 자동화 워크플로우**: tmux 통합을 포함한 worktree 생명주기 자동화 (SPEC-WORKTREE-002)
- **MX 태그 자동 검증 시스템**: 파일 편집 시 @MX 태그 자동 검증 (SPEC-MX-002)
- **GLM 모델 오버라이드**: GLM 설정에서 개별 모델명 오버라이드 지원

### 변경됨 (Changed)

- **Statusline 테마**: "Default/Catppuccin Mocha/Catppuccin Latte"에서 **"MoAI Dark/MoAI Light"**로 이름 변경 — statusline 디자인 언어에 맞는 두 가지 테마
- **Statusline 위자드**: `moai init`/`moai update` 위자드에서 세그먼트 프리셋 선택 UI 제거 — 세그먼트는 `.moai/config/sections/statusline.yaml`에서 직접 설정
- **훅 입력 정규화**: 크로스 환경 호환성을 위한 통합 훅 stdin JSON 파싱 및 정규화 레이어 (#467, #474)

### 삭제됨 (Removed)

- **`/moai context` 명령어**: Git 기반 컨텍스트 조회 워크플로우 삭제 — 컨텍스트는 이제 Claude Code의 내장 auto-memory와 구조화된 커밋 메시지로 관리

### 수정됨 (Fixed)

- **5H/7D 사용량 항상 0% 표시**: Messages API에 OAuth 토큰 사용 시 `anthropic-beta: oauth-2025-04-20` 헤더가 필수인데 누락되어 401 에러 발생 → 사용량 바 항상 빈 상태
- **훅 입력 캡처**: 일부 환경에서 훅 stdin JSON이 올바르게 캡처되지 않는 문제 수정 (#467)
- **훅에서 PATH 캡처**: PATH 환경변수가 올바르게 캡처되지 않는 문제 수정 (#474)
- **Windows 테스트**: `cleanupMoaiWorktrees` 테스트에서 git 경로 구분자 불일치 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.3] - 2026-03-04

### Summary

This release fixes a bug where the `-p`/`--profile` flag without a value was silently ignored and passed through to Claude, causing a confusing "Input must be provided" error. The fix returns a clear error message with usage guidance.

### Breaking Changes

None.

### Added

- `TestParseProfileFlag`: 11 table-driven test cases for profile flag parsing edge cases.

### Changed

None.

### Fixed

- **`-p`/`--profile` flag without value**: Previously, omitting a value after `-p` or `--profile` caused Claude to interpret it as `--print` mode, resulting in a confusing "Input must be provided" error. Now returns a clear error message with usage guidance. Affects `cc`, `cg`, and `glm` commands. (#462)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.3] - 2026-03-04 (한국어)

### 요약

`-p`/`--profile` 플래그 뒤에 값을 지정하지 않으면 조용히 무시되어 Claude에 전달되던 버그를 수정한다. 이로 인해 Claude가 `--print` 모드로 해석해 혼란스러운 "Input must be provided" 에러가 발생했다. 이제 명확한 사용법 안내와 함께 에러 메시지를 반환한다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- `TestParseProfileFlag`: 프로필 플래그 파싱 엣지 케이스를 위한 11개 테이블 드리븐 테스트 케이스.

### 변경됨 (Changed)

없음.

### 수정됨 (Fixed)

- **`-p`/`--profile` 플래그 값 누락 버그**: `-p` 또는 `--profile` 뒤에 값이 없을 때 에러 없이 무시되어 Claude의 `--print` 모드로 해석되던 문제를 수정한다. 이제 사용법 안내를 포함한 에러 메시지를 반환한다. `cc`, `cg`, `glm` 커맨드에 영향. (#462)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.2] - 2026-03-03

### Summary

This release introduces three Boris Cherny-inspired workflow improvements — Lessons Protocol, Re-planning Gate, and Pre-submission Self-Review — that help AI agents learn from mistakes and avoid repetitive loops. The deprecated `expert-chrome-extension` agent (now built into Claude Code) has been removed.

### Breaking Changes

None.

### Added

- **Lessons Protocol**: AI agents now capture user corrections automatically in `~/.claude/projects/{project-hash}/memory/lessons.md`. Each session scans relevant lessons before starting domain-specific tasks.
- **Re-planning Gate**: When an implementation loop stalls for 3+ iterations without acceptance-criteria progress, MoAI automatically triggers a replanning phase instead of continuing futile attempts.
- **Pre-submission Self-Review**: Before submitting work, agents run a self-review gate that checks code quality, test coverage, and annotation completeness. Skipped for changes explicitly approved in the SPEC annotation cycle.

### Changed

- Template rule files updated to include Boris Cherny best practices across `moai-constitution.md`, `spec-workflow.md`, `workflow-modes.md`, and `run.md`.

### Fixed

- Agent count corrected to 27 in all README variants (en/ko/ja/zh). The count was 28 due to a stale reference after the `expert-chrome-extension` removal.

### Removed

- `expert-chrome-extension` agent removed. This agent is now provided as a built-in agent by Claude Code and no longer needs a custom definition in MoAI-ADK.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.2] - 2026-03-03 (한국어)

### 요약

Boris Cherny의 베스트 프랙티스에서 영감을 받은 세 가지 워크플로우 개선 사항을 도입합니다 — Lessons Protocol(교훈 프로토콜), Re-planning Gate(재기획 게이트), Pre-submission Self-Review(제출 전 자체 검토). AI 에이전트가 실수로부터 학습하고 반복 루프를 방지하는 데 도움을 줍니다. 더 이상 필요 없는 `expert-chrome-extension` 에이전트(현재 Claude Code에 내장)가 제거되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **Lessons Protocol**: AI 에이전트가 사용자 수정 사항을 `~/.claude/projects/{project-hash}/memory/lessons.md`에 자동으로 캡처합니다. 각 세션은 도메인별 작업을 시작하기 전에 관련 교훈을 스캔합니다.
- **Re-planning Gate**: 구현 루프가 수락 기준 진행 없이 3회 이상 반복될 경우, MoAI가 무익한 시도를 계속하는 대신 자동으로 재기획 단계를 트리거합니다.
- **Pre-submission Self-Review**: 작업 제출 전, 에이전트가 코드 품질, 테스트 커버리지, 어노테이션 완성도를 점검하는 자체 검토 게이트를 실행합니다. SPEC 어노테이션 사이클에서 명시적으로 승인된 변경 사항은 건너뜁니다.

### 변경됨 (Changed)

- `moai-constitution.md`, `spec-workflow.md`, `workflow-modes.md`, `run.md`에 Boris Cherny 베스트 프랙티스를 포함하도록 템플릿 규칙 파일이 업데이트되었습니다.

### 수정됨 (Fixed)

- 모든 README 변형(en/ko/ja/zh)에서 에이전트 수가 27개로 수정되었습니다. `expert-chrome-extension` 제거 후 오래된 참조로 인해 28개로 표시되었습니다.

### 제거됨 (Removed)

- `expert-chrome-extension` 에이전트 제거. 이 에이전트는 이제 Claude Code에서 기본 제공하며 MoAI-ADK에 커스텀 정의가 더 이상 필요하지 않습니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.1] - 2026-03-03

### Summary

This release enhances the init wizard with locale-awareness and GitLab support, introduces the Execution Mode Selection Gate for cc/glm/cg mode selection, and wires up previously unimplemented packages into the execution flow. CG mode is now the default team mode.

### Breaking Changes

None.

### Added

- **Locale-aware wizard**: The init wizard now automatically detects `conversation_language` from `language.yaml` and displays UI in the user's configured language.
- **GitLab support in wizard**: Added `gitlab_username` and `gitlab_token` fields to the init wizard; stored in `user.yaml`.
- **Pre-fill wizard from existing config**: GitHub/GitLab usernames are pre-filled from existing `user.yaml` values when reconfiguring.
- **gh CLI auth detection**: When `gh auth status` passes, the `github_token` question is automatically skipped.
- **Execution Mode Selection Gate**: When transitioning from Plan to Run phase, MoAI detects the current execution environment (cc/glm/cg) and presents a mode selection UI before implementation begins.
- **Wired packages into execution flow**: `loop`, `hook/post_tool`, `hook/session_end`, `lsp/hook/fallback`, `tmux/detector`, `tmux/session` packages are now fully integrated into the execution flow.

### Changed

- **CG mode as default team mode**: `team_mode` in workflow config now defaults to `cg` (Claude Leader + GLM Workers) for optimal cost-quality balance.
- **Wizard API refactored**: `RunWithDefaults` now accepts a `locale` parameter; new `RunWithLocale` function introduced.
- **Template skill updates**: Workflow skills (plan, run, sync, context, moai) updated with Execution Mode Selection Gate integration.

### Fixed

None.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.1] - 2026-03-03 (한국어)

### 요약

이번 릴리즈에서는 init 위저드에 로케일 인식 기능과 GitLab 지원이 추가되었으며, cc/glm/cg 모드 선택을 위한 실행 모드 선택 게이트가 도입되었습니다. 미구현 패키지들이 실행 흐름에 통합되었으며, CG 모드가 기본 팀 모드로 설정되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **로케일 인식 위저드**: init 위저드가 `language.yaml`의 `conversation_language`를 자동 감지하여 해당 언어로 UI를 표시합니다.
- **위저드 GitLab 지원**: init 위저드에 `gitlab_username` 및 `gitlab_token` 필드 추가; `user.yaml`에 저장됩니다.
- **기존 설정 값 사전 입력**: 재구성 시 기존 `user.yaml`의 GitHub/GitLab 사용자명을 기본값으로 미리 채워줍니다.
- **gh CLI 인증 자동 감지**: `gh auth status`가 통과된 경우 `github_token` 질문이 자동으로 스킵됩니다.
- **실행 모드 선택 게이트**: Plan에서 Run 단계로 전환 시, 현재 실행 환경(cc/glm/cg)을 감지하고 구현 시작 전 모드 선택 UI를 제공합니다.
- **패키지 실행 흐름 통합**: `loop`, `hook/post_tool`, `hook/session_end`, `lsp/hook/fallback`, `tmux/detector`, `tmux/session` 패키지가 실행 흐름에 완전히 통합되었습니다.

### 변경됨 (Changed)

- **CG 모드 기본 팀 모드**: 워크플로우 설정의 `team_mode`가 최적의 비용-품질 균형을 위해 `cg`(Claude 리더 + GLM 워커)로 기본 설정되었습니다.
- **위저드 API 리팩토링**: `RunWithDefaults`가 `locale` 파라미터를 받도록 변경; 새로운 `RunWithLocale` 함수가 도입되었습니다.
- **템플릿 스킬 업데이트**: 실행 모드 선택 게이트 통합을 위해 워크플로우 스킬(plan, run, sync, context, moai)이 업데이트되었습니다.

### 수정됨 (Fixed)

없음.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.7.0] - 2026-03-02

### Summary

This release introduces true login-based profile management for cc/cg/glm commands, harness workflow skill improvements, and status line enhancements. Contributor: [@yejune](https://github.com/yejune) (PR [#431](https://github.com/modu-ai/moai-adk/pull/431)).

### Breaking Changes

None.

### Added

- **Profile-based login for cc/cg/glm commands**: `moai cc`, `moai cg`, and `moai glm` are now true login commands with persistent profile support. Profiles stored at `~/.moai/claude-profiles/`.
- **Profile setup wizard**: Interactive wizard for initial profile configuration when using cc/cg/glm commands for the first time.
- **Harness workflow improvements**: Updated workflow skills including enhanced batch mode decision logic, simplified run workflow, and streamlined team agents configuration.
- **Status line enhancements**: Additional functionality for the status line template.

### Changed

- **Profile relocation**: Profile files moved to `~/.moai/claude-profiles/` for better organization.
- **Template refinements**: Simplified simplify/batch pass integration into review workflow, streamlined team agents and skills format.
- **CI dependency updates**: claude-code-action 1.0.63→1.0.64, actions/upload-artifact 6→7.

### Fixed

- **Security and performance issues** in profile and launch commands.
- **errcheck lint violations** in profile and glm commands.
- **Accidentally committed temp files** removed from repository; `temp/` added to `.gitignore`.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.7.0] - 2026-03-02 (한국어)

### 요약

이번 릴리즈에서는 cc/cg/glm 명령어에 로그인 기반 프로파일 관리 기능이 추가되었습니다. harness 워크플로우 스킬 개선과 상태줄 기능 향상도 포함됩니다. 기여자: [@yejune](https://github.com/yejune) (PR [#431](https://github.com/modu-ai/moai-adk/pull/431)).

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **cc/cg/glm 명령어 로그인 기반 프로파일 지원**: `moai cc`, `moai cg`, `moai glm`이 이제 영구 프로파일을 지원하는 진정한 로그인 명령어로 동작합니다. 프로파일은 `~/.moai/claude-profiles/`에 저장됩니다.
- **프로파일 설정 위자드**: cc/cg/glm 명령어 최초 실행 시 대화형 프로파일 설정 위자드 제공.
- **Harness 워크플로우 개선**: 배치 모드 결정 로직 강화, run 워크플로우 단순화, team agents 설정 정리.
- **상태줄 기능 향상**: 상태줄 템플릿에 추가 기능 구현.

### 변경됨 (Changed)

- **프로파일 위치 변경**: 프로파일 파일이 `~/.moai/claude-profiles/`로 이동하여 구성이 개선되었습니다.
- **템플릿 개선**: review 워크플로우에 simplify/batch 통합, team agents 및 skills 형식 정리.
- **CI 의존성 업데이트**: claude-code-action 1.0.63→1.0.64, actions/upload-artifact 6→7.

### 수정됨 (Fixed)

- **프로파일 및 launch 명령어의 보안/성능 문제** 수정.
- **errcheck lint 위반** 수정 (profile, glm 명령어).
- **실수로 커밋된 temp 파일** 제거 및 `temp/`를 `.gitignore`에 추가.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.15] - 2026-02-28

### Summary

This release includes memory architecture migration to Claude Code's native auto-memory, builder-skill path enforcement, and multiple bug fixes for SessionEnd hooks and GLM compatibility.

### Breaking Changes

None.

### Changed

- **Memory-to-auto-memory migration**: Replaced custom `.moai/memory/` injection system with Claude Code's native auto-memory at `~/.claude/projects/`. Removed `InjectMemoryToPrompt` and related code.
- Rules loading optimization, @MX tags English standardization, worktree isolation rules, skill definition updates.

### Fixed

- **builder-skill path enforcement** (#444): Skill creation now enforces correct skill path structure and uses `my-` prefix for user-created skills to prevent conflicts with built-in skills.
- **State migration cleanup**: Completed memory-to-state migration with proper cleanup of deprecated state files and update logic.
- **SessionEnd hook improvements**: Added `context.Context` with timeout to `clearTmuxSessionEnv`, CWD fallback for `ProjectDir`, cleanup failure summary logging, and timeout warning for tmux cleanup.
- **GLM model compatibility**: Replaced deprecated `glm-4.7-flashx` with `glm-4.5-air` for coding plan compatibility.
- **GLM settings cleanup**: Added GLM environment variable cleanup to SessionEnd handler to prevent stale variables.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.15] - 2026-02-28 (한국어)

### 요약

이번 릴리스는 Claude Code 네이티브 자동 메모리로의 메모리 아키텍처 마이그레이션, builder-skill 경로 적용, SessionEnd 훅 및 GLM 호환성 관련 다수의 버그 수정을 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 변경됨 (Changed)

- **메모리-자동메모리 마이그레이션**: 커스텀 `.moai/memory/` 주입 시스템을 Claude Code의 네이티브 자동 메모리(`~/.claude/projects/`)로 교체했습니다. `InjectMemoryToPrompt` 및 관련 코드가 제거되었습니다.
- 규칙 로딩 최적화, @MX 태그 영어 표준화, 워크트리 격리 규칙, 스킬 정의 업데이트.

### 수정됨 (Fixed)

- **builder-skill 경로 적용** (#444): 스킬 생성 시 올바른 스킬 경로 구조를 적용하고, 내장 스킬과의 충돌을 방지하기 위해 사용자 생성 스킬에 `my-` 접두사를 사용합니다.
- **상태 마이그레이션 정리**: 메모리-스테이트 마이그레이션을 완료하고 더 이상 사용되지 않는 상태 파일 및 업데이트 로직을 정리했습니다.
- **SessionEnd 훅 개선**: `clearTmuxSessionEnv`에 타임아웃 `context.Context` 추가, `ProjectDir` CWD 폴백, 정리 실패 요약 로깅, tmux 정리 타임아웃 경고가 추가되었습니다.
- **GLM 모델 호환성**: 더 이상 사용되지 않는 `glm-4.7-flashx`를 코딩 플랜 호환을 위해 `glm-4.5-air`로 교체했습니다.
- **GLM 설정 정리**: stale 변수 방지를 위해 SessionEnd 핸들러에 GLM 환경 변수 정리가 추가되었습니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.12] - 2026-02-27

### Summary

This patch release fixes an OAuth token loss bug introduced in earlier releases.
When `moai glm` injected GLM API credentials into `settings.local.json`, any
pre-existing `ANTHROPIC_AUTH_TOKEN` (stored by Claude Code's `/login` OAuth flow)
was silently overwritten. When `moai cc` then cleaned up GLM env vars, it deleted
the token entirely — causing Claude Code to prompt for `/login` on every session
restart.

The fix uses a backup/restore pattern: `moai glm` backs up the existing OAuth token
as `MOAI_BACKUP_AUTH_TOKEN` before overwriting, and `moai cc` restores it when
removing the GLM key.

### Breaking Changes

None.

### Fixed

- **OAuth token preserved across moai glm/cc round-trip**: `injectGLMEnvForTeam`
  and `injectGLMEnv` now back up any pre-existing `ANTHROPIC_AUTH_TOKEN` as
  `MOAI_BACKUP_AUTH_TOKEN` before injecting the GLM key. `removeGLMEnv` restores
  the backed-up token when cleaning up GLM env vars, eliminating the `/login` on
  restart regression for OAuth users.

### Tests

- Added `oauth_token_preservation_test.go` with 4 test cases covering:
  the full GLM→CC round-trip, individual inject/remove behavior, and the
  no-prior-token case.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.11] - 2026-02-27

### Summary

This patch release restores the pre-v2.6 unconditional GLM token cleanup in the SessionEnd hook, permanently fixing the `/login` prompt regression. Previously, cleanup was conditional on finding an active team session; this meant GLM environment variables could persist in tmux after a team session ended abnormally. The cleanup is now unconditional and safe: it's a no-op when not in tmux or when the variables don't exist.

### Breaking Changes

None.

### Fixed

- **Unconditional GLM token cleanup on SessionEnd**: Restored pre-v2.6 behavior where `clearTmuxSessionEnv()` is called unconditionally at every session end. The previous conditional logic (only cleaning up when a matching team session was found) allowed stale GLM env vars to persist in tmux, causing `/login` prompts on the next session.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.11] - 2026-02-27 (한국어)

### 요약

이번 패치 릴리스는 SessionEnd 훅에서 pre-v2.6의 무조건적인 GLM 토큰 정리를 복원하여 `/login` 프롬프트 회귀 문제를 영구적으로 수정합니다. 이전에는 정리가 활성 팀 세션을 찾은 경우에만 조건부로 실행되어, 팀 세션이 비정상 종료될 경우 GLM 환경 변수가 tmux에 남아있을 수 있었습니다. 이제 정리는 무조건적으로 실행되며 안전합니다: tmux 환경이 아니거나 변수가 없는 경우 no-op으로 처리됩니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **SessionEnd 시 무조건적인 GLM 토큰 정리**: 매 세션 종료 시 `clearTmuxSessionEnv()`가 무조건 호출되는 pre-v2.6 동작 복원. 이전의 조건부 로직(매칭되는 팀 세션이 발견된 경우에만 정리)은 비정상 종료 시 stale GLM env 변수를 tmux에 남겨 다음 세션에서 `/login` 프롬프트를 유발했음.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.10] - 2026-02-26

### Summary

This release fixes the root cause of `/login` prompts appearing on every session restart. The team skill cleanup instructions were manually overwriting `~/.claude/settings.local.json` using the AI's Write tool, which silently erased `ANTHROPIC_AUTH_TOKEN` and other settings. All cleanup now delegates to `moai cc` which safely handles JSON merging. Also includes worktree-isolated parallel processing and local CI mirror for the GitHub workflow.

### Breaking Changes

None.

### Added

- **Worktree-isolated parallel processing**: GitHub workflow now uses worktree isolation for parallel issue fixing, preventing file conflicts during concurrent work.
- **Local CI mirror**: CI validation runs locally in worktrees before pushing, reducing failed remote CI runs.

### Fixed

- **Persistent /login prompts (root cause)**: Team skill cleanup in `plan.md`, `run.md`, `debug.md`, and `review.md` instructed MoAI to manually `Read`/`Write` `~/.claude/settings.local.json`. The `Write` tool overwrites the entire file, destroying `ANTHROPIC_AUTH_TOKEN` and forcing re-login on every session. All four skills now delegate cleanup to `moai cc` which handles JSON merging correctly.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.10] - 2026-02-26 (한국어)

### 요약

이번 릴리스는 매 세션 재시작 시 `/login`이 나타나는 문제의 근본 원인을 수정합니다. 팀 스킬의 cleanup 지시가 AI의 `Write` 도구로 `~/.claude/settings.local.json`을 직접 덮어쓰며 `ANTHROPIC_AUTH_TOKEN` 및 기타 설정을 삭제했습니다. 이제 모든 cleanup은 JSON 병합을 안전하게 처리하는 `moai cc`에 위임합니다. GitHub 워크플로우용 워크트리 격리 병렬 처리 및 로컬 CI 미러도 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **워크트리 격리 병렬 처리**: GitHub 워크플로우가 병렬 이슈 수정 시 워크트리 격리를 사용하여 동시 작업 중 파일 충돌 방지.
- **로컬 CI 미러**: 푸시 전 워크트리에서 CI 검증을 로컬로 실행하여 원격 CI 실패 감소.

### 수정됨 (Fixed)

- **지속적인 /login 프롬프트 (근본 원인)**: `plan.md`, `run.md`, `debug.md`, `review.md`의 팀 스킬 cleanup이 MoAI에게 `~/.claude/settings.local.json`을 직접 `Read`/`Write`하도록 지시했음. `Write` 도구는 전체 파일을 덮어써 `ANTHROPIC_AUTH_TOKEN`을 삭제하여 매 세션마다 재로그인을 강제함. 4개 스킬 모두 JSON 병합을 올바르게 처리하는 `moai cc`에 cleanup을 위임하도록 수정.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.8] - 2026-02-26

### Summary

This release introduces the auto-memory system for persistent session context, fixes ANTHROPIC_AUTH_TOKEN preservation across all GLM cleanup paths, resolves cross-platform test failures, and removes deprecated WorktreeCreate/WorktreeRemove hooks from settings templates.

### Breaking Changes

None.

### Added

- **Auto-memory system**: `SessionStart` hook now automatically discovers and injects the `.claude/projects/` memory directory path into the SystemMessage, giving Claude Code persistent context across sessions.
- **Memory configuration**: New `.moai/config/sections/memory.yaml` and `.moai/config/sections/context.yaml` configuration files for controlling memory and context behavior.
- **Template memory directory**: Added `.moai/memory/.gitkeep` to template output so projects initialize with a proper memory storage location.

### Changed

- **Template settings cleanup**: Removed deprecated `WorktreeCreate` and `WorktreeRemove` hook definitions from `settings.json.tmpl`.

### Fixed

- **ANTHROPIC_AUTH_TOKEN preservation**: `moai cc` now preserves `ANTHROPIC_AUTH_TOKEN` unless it matches the stored GLM key, preventing accidental deletion of the user's permanent API credential.
- **Session-end GLM cleanup**: `cleanupGLMSettings()` now excludes `ANTHROPIC_AUTH_TOKEN` from removal during session-end cleanup.
- **Registry SystemMessage merging**: `internal/hook/registry.go` now appends SystemMessage content from multiple hooks instead of overwriting the previous value.
- **Windows CI test failures**: Fixed cross-platform test failures with `context.TODO()` and platform-specific path handling.
- **Test race conditions**: Removed `t.Parallel()` from tests that mutate global `deps` variable.
- **Test browser isolation**: Injected mock browser to prevent real browser opening during OAuth flow tests.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.8] - 2026-02-26 (한국어)

### 요약

이번 릴리스는 영구 세션 컨텍스트를 위한 자동 메모리 시스템 도입, 모든 GLM 정리 경로에서의 ANTHROPIC_AUTH_TOKEN 보존 수정, 크로스 플랫폼 테스트 실패 해결, 더 이상 사용되지 않는 WorktreeCreate/WorktreeRemove 훅의 설정 템플릿 제거를 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **자동 메모리 시스템**: `SessionStart` 훅이 이제 `.claude/projects/` 메모리 디렉토리 경로를 자동으로 감지하여 SystemMessage에 주입, 세션 간 영구적인 컨텍스트를 제공합니다.
- **메모리 설정**: 메모리 및 컨텍스트 동작을 제어하는 새 설정 파일 `.moai/config/sections/memory.yaml`과 `.moai/config/sections/context.yaml` 추가.
- **템플릿 메모리 디렉토리**: 프로젝트 초기화 시 적절한 메모리 저장 위치를 갖도록 템플릿 출력에 `.moai/memory/.gitkeep` 추가.

### 변경됨 (Changed)

- **템플릿 설정 정리**: `settings.json.tmpl`에서 더 이상 사용되지 않는 `WorktreeCreate`/`WorktreeRemove` 훅 정의 제거.

### 수정됨 (Fixed)

- **ANTHROPIC_AUTH_TOKEN 보존**: `moai cc`가 이제 저장된 GLM 키와 일치하지 않는 한 `ANTHROPIC_AUTH_TOKEN`을 보존하여 사용자의 영구 API 자격증명이 실수로 삭제되는 것을 방지.
- **세션 종료 GLM 정리**: `cleanupGLMSettings()`가 세션 종료 정리 시 `ANTHROPIC_AUTH_TOKEN`을 제거 대상에서 제외.
- **레지스트리 SystemMessage 병합**: `internal/hook/registry.go`가 이제 여러 훅의 SystemMessage 내용을 덮어쓰지 않고 이어붙임.
- **Windows CI 테스트 실패**: `context.TODO()` 및 플랫폼별 경로 처리로 크로스 플랫폼 테스트 실패 수정.
- **테스트 Race Condition**: 전역 `deps` 변수를 변경하는 테스트에서 `t.Parallel()` 제거.
- **테스트 브라우저 격리**: OAuth 플로우 테스트에서 실제 브라우저가 열리지 않도록 mock 브라우저 주입.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.5] - 2026-02-26

### Summary

Hotfix release correcting a configuration management regression introduced in v2.6.4. The v2.6.4 release inadvertently preserved `ANTHROPIC_AUTH_TOKEN` in project `settings.local.json` files after GLM mode cleanup, causing GLM API keys to be routed to Anthropic's API endpoints (resulting in 401 authentication errors and persistent `/login` prompts). This release restores the correct behavior where GLM credentials are managed solely through `~/.moai/.env.glm` and properly removed from project settings during `moai cc` or session end.

### Breaking Changes

None.

### Fixed

- **Configuration management regression**: Reverted v2.6.4's preservation of `ANTHROPIC_AUTH_TOKEN` in `settings.local.json`. The token is now correctly removed by `removeGLMEnv()` (cc.go), `clearTmuxSessionEnv()` (glm.go), and `cleanupGLMSettings()` (session_end.go) during cleanup operations.
- **Root cause**: v2.6.4 attempted to preserve user's permanent API credentials, but this caused GLM keys (stored in `~/.moai/.env.glm`) to persist in project settings and be routed to Anthropic's API instead of GLM endpoints.
- **Empty file handling**: Added graceful handling of empty `settings.local.json` files (0-byte files) to prevent JSON parsing errors.
- **Home directory detection**: Fixed `findProjectRoot()` to skip `~/.moai/` (global cache) when searching for project root.

### Changed

- **Agent Teams display mode**: Changed `CLAUDE_CODE_TEAMMATE_DISPLAY` from `"tmux"` to `"auto"` for better fallback support. Claude Code now automatically detects tmux availability and falls back to in-process display when tmux is not installed.
- **Test coverage**: Increased test coverage for GLM-related functions and edge cases.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.5] - 2026-02-26 (한국어)

### 요약

v2.6.4에서 도입된 설정 관리 회귀를 수정하는 핫픽스 릴리스. v2.6.4 릴리스는 GLM 모드 정리 후 프로젝트 `settings.local.json`에 `ANTHROPIC_AUTH_TOKEN`을 실수로 보존하여, GLM API 키가 Anthropic API 엔드포인트로 라우팅되는 문제(401 인증 오류 및 지속적인 `/login` 프롬프트)를 일으켰음. 이 릴리스는 GLM 자격증명이 `~/.moai/.env.glm`를 통해 exclusive하게 관리되며 `moai cc` 또는 세션 종료 시 프로젝트 설정에서 올바르게 제거되는 올바른 동작을 복원.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **설정 관리 회귀**: v2.6.4의 `settings.local.json`에서 `ANTHROPIC_AUTH_TOKEN` 보존 동작을 되돌림. 토큰이 이제 `removeGLMEnv()` (cc.go), `clearTmuxSessionEnv()` (glm.go), `cleanupGLMSettings()` (session_end.go)에 의해 정리 작업 중 올바르게 제거됨.
- **근본 원인**: v2.6.4는 사용자의 영구 API 자격증명을 보존하려 했으나, 이로 인해 `~/.moai/.env.glm`에 저장된 GLM 키가 프로젝트 설정에 남아 GLM 엔드포인트가 아닌 Anthropic API로 라우팅되는 문제 발생.
- **빈 파일 처리**: 빈 `settings.local.json` 파일(0바이트)을 우아하게 처리하여 JSON 파싱 오류 방지.
- **홈 디렉토리 감지**: 프로젝트 루트 검색 시 `~/.moai/`(전역 캐시)를 건너뛰도록 수정.

### 변경됨 (Changed)

- **Agent Teams 표시 모드**: `CLAUDE_CODE_TEAMMATE_DISPLAY`를 `"tmux"`에서 `"auto"`로 변경하여 더 나은 폴백 지원. Claude Code가 tmux 사용 가능성을 자동 감지하고 tmux가 설치되지 않은 경우 in-process 표시로 폴백.
- **테스트 커버리지**: GLM 관련 함수 및 엣지 케이스에 대한 테스트 커버리지 증대.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.4] - 2026-02-26

### Summary

Bug fix release completing the login persistence fix by addressing remaining CLI paths that were removing API credentials. This release ensures `ANTHROPIC_AUTH_TOKEN` is preserved across all code paths (hooks and CLI commands).

### Breaking Changes

None.

### Fixed

- **CLI authentication persistence**: Fixed `removeGLMEnv()` in `cc.go` and `clearTmuxSessionEnv()` in `glm.go` incorrectly removing `ANTHROPIC_AUTH_TOKEN`. These CLI functions were not updated in the previous v2.6.3 fix (which only addressed the hook path), causing `/login` prompts to persist after `moai cc` or `moai cg` commands. Now all code paths (hooks and CLI) consistently preserve the authentication token.
- **Test coverage**: Updated all related tests to expect token preservation behavior.

### Changed

- **Workflow documentation**: Standardized flag aliases across workflow skill files for consistency (`--solo` and `--team` flags now uniformly available).

### Installation & Update

\`\`\`bash
# Update to the latest version
moai update

# Verify version
moai version
\`\`\`

---

## [2.6.4] - 2026-02-26 (한국어)

### 요약

로그인 지속성 수정을 완료하는 버그 수정 릴리스입니다. API 자격증명을 삭제하던 나머지 CLI 경로를 수정하여 모든 코드 경로(hooks 및 CLI 명령)에서 인증 토큰이 보존되도록 완성했습니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **CLI 인증 지속성**: `cc.go`의 `removeGLMEnv()`와 `glm.go`의 `clearTmuxSessionEnv()`가 `ANTHROPIC_AUTH_TOKEN`을 잘못 삭제하는 문제 수정. 이 CLI 함수들은 이전 v2.6.3 수정(hook 경로만 수정됨)에서 업데이트되지 않아 `moai cc` 또는 `moai cg` 명령 후 `/login` 프롬프트가 지속되었음. 이제 모든 코드 경로(hooks 및 CLI)가 인증 토큰을 일관되게 보존.
- **테스트 커버리지**: 관련 모든 테스트를 토큰 보존 동작을 기대하도록 업데이트.

### 변경됨 (Changed)

- **워크플로우 문서화**: 워크플로우 스킬 파일 전체에서 플래그 별칭 표준화로 일관성 개선 (`--solo` 및 `--team` 플래그가 이제 모든 워크플로우에서 사용 가능).

### 설치 및 업데이트 (Installation & Update)

\`\`\`bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
\`\`\`

---

## [2.6.3] - 2026-02-26

### Summary

Bug fix release resolving session authentication persistence issue. Restores login state between Claude Code restarts by preserving API credentials during session cleanup.

### Breaking Changes

None.

### Fixed

- **Session authentication persistence**: Fixed `cleanupGLMSettings()` and `clearTmuxSessionEnv()` in `session_end.go` incorrectly removing `ANTHROPIC_AUTH_TOKEN` on session end. This token is the user's permanent API credential (e.g., GLM API key), not a temporary team-mode token. Removing it forced `/login` on every Claude Code restart. Now only `ANTHROPIC_BASE_URL` and model overrides are cleaned up to reset GLM team mode while preserving authentication.
- **Agent memory cleanup**: Added `.claude/agent-memory/` to `.gitignore` as ephemeral agent state that should not be committed to version control.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.3] - 2026-02-26 (한국어)

### 요약

세션 인증 지속성 문제를 해결하는 버그 수정 릴리즈입니다. 세션 정리 중 API 자격증명을 보존하여 Claude Code 재시작 간 로그인 상태를 유지합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **세션 인증 지속성**: `session_end.go`의 `cleanupGLMSettings()`와 `clearTmuxSessionEnv()`가 세션 종료 시 `ANTHROPIC_AUTH_TOKEN`을 잘못 삭제하는 문제 수정. 이 토큰은 임시 팀 모드 토큰이 아니라 사용자의 영구 API 자격증명(예: GLM API 키)임. 삭제 시 Claude Code 재시작마다 `/login`이 필요했음. 이제 `ANTHROPIC_BASE_URL`과 모델 오버라이드만 정리하여 GLM 팀 모드만 리셋하고 인증은 보존.
- **에이전트 메모리 정리**: 버전 관리에 커밋되지 않아야 할 임시 에이전트 상태인 `.claude/agent-memory/`를 `.gitignore`에 추가.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.2-teddy] - 2026-02-26

### Summary

**Teddy Edition** - Bug fix release with documentation optimization and Windows CI stability improvements. Resolves token efficiency issue and prevents CI test timeouts.

### Breaking Changes

None.

### Fixed

- **Token optimization**: Removed `@` direct references to `agent-authoring.md` from CLAUDE.md and SKILL.md to honor `paths` frontmatter restriction. Saves ~2,000–2,800 tokens per session for users not creating agents. (#427)
- **Windows CI stability**: Skip Non-TTY interactive tests on Windows to prevent indefinite hangs in `huh.Form.Run()`. Test timeouts resolved across 7 affected tests. (#428)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.2-teddy] - 2026-02-26 (한국어)

### 요약

**Teddy Edition** - 문서 최적화와 Windows CI 안정성 개선이 포함된 버그 수정 릴리즈입니다. 토큰 효율성 문제와 CI 테스트 타임아웃을 해결합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **토큰 최적화**: `agent-authoring.md`에 대한 `@` 직접 참조를 CLAUDE.md와 SKILL.md에서 제거하여 `paths` frontmatter 제약을 준수하도록 수정. 에이전트 생성을 하지 않는 사용자의 세션당 ~2,000–2,800 토큰 절약. (#427)
- **Windows CI 안정성**: Windows에서 `huh.Form.Run()`의 무한 대기를 방지하기 위해 Non-TTY 인터랙티브 테스트 스킵. 7개 영향받은 테스트의 타임아웃 문제 해결. (#428)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.1-teddy] - 2026-02-26

### Summary

**Teddy Edition** - Special pre-release with code formatting improvements and build consistency updates. This edition celebrates community collaboration and contributions to the MoAI project.

### Breaking Changes

None.

### Fixed

- Code formatting consistency improvements across codebase

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.1-teddy] - 2026-02-26 (한국어)

### 요약

**Teddy Edition** - 코드 포맷팅 개선과 빌드 일관성 업데이트가 포함된 특별 pre-release입니다. 커뮤니티 협력과 MoAI 프로젝트에 대한 기여를 기립합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- 코드베이스 전체의 코드 포맷팅 일관성 개선

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.6.0] - 2026-02-25

### Summary

This feature release introduces MX pre-scan for intelligent code context analysis and Context Memory system for cross-session learning. Also adds automatic GLM settings cleanup when using Agent Teams, improved Windows internationalization support with Korean path workarounds, and safer SessionEnd hook behavior.

### Breaking Changes

None.

### Added

- **MX Pre-scan system**: New workflow that performs intelligent codebase analysis before SPEC creation. Scans codebase for high-value targets (high fan_in functions, complex code, danger zones) and produces context-aware analysis to guide SPEC development.
- **Context Memory system**: New `/moai context` command for managing persistent memory across sessions. View, search, add, and remove context memories that carry over between Claude Code sessions for improved continuity.
- **Windows Korean path workaround**: Added special handling for Windows systems with Korean locale to work around Go's filepath.Join bug with absolute paths on non-English Windows versions.

### Changed

- **Team workflow GLM cleanup**: All team workflows (plan, run, debug, review) now automatically clean up GLM environment variables from `~/.claude/settings.local.json` before TeamDelete, ensuring main session returns to Claude models after team work completes.
- **SessionEnd hook safety**: Removed unsafe tmux session cleanup that could accidentally kill user tmux sessions. CG mode teammates are properly managed through `~/.claude/teams/` directory structure.
- **Manager-git agent**: Enhanced with additional task management capabilities for better team coordination.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.6.0] - 2026-02-25 (한국어)

### 요약

이 기능 릴리즈는 지능형 코드 컨텍스트 분석을 위한 MX 사전 스캔과 세션 간 학습을 위한 컨텍스트 메모리 시스템을 도입합니다. 또한 Agent Teams 사용 시 자동 GLM 설정 정리, 한국어 경로 해결을 통한 개선된 Windows 국제화 지원, 그리고 더 안전한 SessionEnd 후크 동작을 추가합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **MX 사전 스캔 시스템**: SPEC 생성 전 지능형 코드베이스 분석을 수행하는 새로운 워크플로우. 높은 fan_in 함수, 복잡한 코드, 위험 영역 등 높은 가치 대상을 스캔하고 컨텍스트 인식 분석을 생성하여 SPEC 개발을 안내합니다.
- **컨텍스트 메모리 시스템**: 세션 간 지속 메모리 관리를 위한 새로운 `/moai context` 명령어. Claude Code 세션 간에 지속되는 컨텍스트 메모리를 보고, 검색, 추가, 제거하여 개선된 연속성을 제공합니다.
- **Windows 한국어 경로 해결책**: 비영어 Windows 버전에서 Go의 filepath.Join 버그를 우회하기 위한 한국어 로케일 Windows 시스템을 위한 특수 처리 추가.

### 변경됨 (Changed)

- **팀 워크플로우 GLM 정리**: 모든 팀 워크플로우(plan, run, debug, review)가 TeamDelete 전에 `~/.claude/settings.local.json`에서 GLM 환경 변수를 자동으로 정리하여 팀 작업 완료 후 메인 세션이 Claude 모델로 돌아오도록 합니다.
- **SessionEnd 후크 안전성**: 사용자 tmux 세션을 실수로 종료할 수 있는 안전하지 않은 tmux 세션 정리를 제거했습니다. CG 모드 팀메이트는 `~/.claude/teams/` 디렉토리 구조를 통해 적절히 관리됩니다.
- **manager-git 에이전트**: 향상된 팀 코디네이션을 위한 추가적인 작업 관리 기능.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.5.3] - 2026-02-25

### Summary

This patch release improves the E2E test workflow by always presenting tool selection to users with intelligent recommendations, rather than auto-selecting when only one tool is available. Also includes a Windows console encoding fix for proper Unicode rendering.

### Breaking Changes

None.

### Changed

- **E2E tool selection**: The `/moai:e2e` command now always prompts users to select between Playwright CLI, Agent Browser, and Claude in Chrome. Auto-selection logic has been converted to recommendation logic that marks the best option with "(Recommended)" based on task characteristics. Each option now displays installation status (installed/not installed) to help users make informed choices.

### Fixed

- **Windows PowerShell Unicode rendering**: Fixed mojibake (garbled characters like `Γò¡ΓöÇΓò«`) when displaying box-drawing borders in PowerShell on Windows. The CLI now calls `SetConsoleOutputCP(65001)` on startup to enable UTF-8 console mode, allowing proper rendering of Unicode characters (`╔═╗║╚╝`) and symbols (`✓✗⚠`).

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.5.3] - 2026-02-25 (한국어)

### 요약

이 패치 릴리즈는 E2E 테스트 워크플로우를 개선하여 도구가 하나만 감지되어도 사용자에게 항상 선택을 제시하고 지능적인 추천을 제공합니다. 또한 Windows 콘솔 인코딩 수정을 포함하여 유니코드 문자가 올바르게 렌더링되도록 합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 변경됨 (Changed)

- **E2E 도구 선택**: `/moai:e2e` 명령이 이제 Playwright CLI, Agent Browser, Claude in Chrome 중 선택을 항상 사용자에게 묻습니다. 자동 선택 로직은 추천 로직으로 변환되어 작업 특성에 따라 최적 옵션에 "(Recommended)" 표시를 합니다. 각 옵션은 설치 상태(설치됨/설치안됨)를 표시하여 사용자가 정보에 입각한 선택을 할 수 있습니다.

### 수정됨 (Fixed)

- **Windows PowerShell 유니코드 렌더링**: Windows PowerShell에서 박스 그리기 테두리 문자가 mojibake(`Γò¡ΓöÇΓò«` 같은 깨진 문자)로 표시되던 문제를 수정했습니다. CLI 시작 시 `SetConsoleOutputCP(65001)`을 호출하여 UTF-8 콘솔 모드를 활성화하므로 유니코드 문자(`╔═╗║╚╝`)와 기호(`✓✗⚠`)가 올바르게 렌더링됩니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

### Changed

- **Team Agent Roster Refactored**: Replaced 8 specialized team agents with a streamlined 5-agent model
  - **Removed**: `team-researcher`, `team-analyst`, `team-architect`, `team-backend-dev`, `team-frontend-dev`, `team-quality`
  - **Added**: `team-reader` (plan-phase read-only), `team-coder` (implementation), `team-validator` (quality validation)
  - **Updated**: `team-tester`, `team-designer` now use worktree isolation and background mode
  - Roles are assigned dynamically via Task() spawn prompt rather than by agent name
  - Added `team-protocol.md` shared protocol reference for all team agents

- **Skills Improvement (moai Skills 개선 계획)**:
  - Fixed `triggers.keywords` YAML block array → inline array format in 12 skills (P1 Critical)
    - Affected: `moai-foundation-*` (6), `moai-domain-*` (4), `moai-workflow-spec`, `moai-workflow-thinking`
  - Split `moai-workflow-spec/SKILL.md` 561→449 lines, extracted migration guide to `reference/migration-guide.md` (P2)
  - Added `disable-model-invocation: true` to `99-release.md` to prevent accidental production releases (P3)

- **New Template**: Added `github.md` command to template distribution

---

## [2.5.2] - 2026-02-23

### Summary

This patch release fixes a critical issue where SessionEnd hook was incorrectly killing user's tmux sessions instead of only cleaning up CG mode teammate sessions.

### Breaking Changes

None.

### Fixed

- **SessionEnd tmux cleanup**: Removed the unsafe `cleanupOrphanedTmuxSessions()` function that was killing ALL non-current, non-attached tmux sessions regardless of whether they were CG mode teammate sessions or user sessions. The root cause was that CG mode teammates are created as tmux **panes** within the current session (not separate sessions), so `tmux kill-session` could never clean them up. Real teammate cleanup is properly handled by `cleanupCurrentSessionTeam()` and `garbageCollectStaleTeams()` which operate on the `~/.claude/teams/` directory structure.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.5.2] - 2026-02-23 (한국어)

### 요약

이 패치 릴리즈는 SessionEnd 훅이 CG mode teammate 세션이 아닌 사용자의 tmux 세션을 잘못 종료하던 치명적인 문제를 수정합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **SessionEnd tmux 정리**: CG mode teammate 세션인지 여부를 구분하지 않고 **모든** non-current, non-attached tmux 세션을 kill하던 잘못된 `cleanupOrphanedTmuxSessions()` 함수를 제거했습니다. 근본 원인은 CG mode의 teammate들이 별도 tmux 세션이 아니라 현재 세션 내의 **pane**으로 생성되기 때문에 `tmux kill-session`로는 절대 정리할 수 없었습니다. 실제 teammate 정리는 `~/.claude/teams/` 디렉토리 구조를 기반으로 작동하는 `cleanupCurrentSessionTeam()`과 `garbageCollectStaleTeams()`가 올바르게 처리합니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.5.1] - 2026-02-23

### Summary

Bug fix release preventing SessionEnd hook from accidentally terminating the user's tmux session when cleaning up orphaned teammate sessions.

### Breaking Changes

None.

### Fixed

- **SessionEnd tmux cleanup**: SessionEnd hook now correctly identifies and protects the user's actual tmux session from being killed during orphaned session cleanup. Added `getCurrentTmuxSession()` helper function that uses `tmux display-message -p '#S'` to get the current session name. The cleanup logic now explicitly skips the current session before checking for "(attached)" status, preventing accidental termination of the user's tmux session when Claude Code session ends.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.5.1] - 2026-02-23 (한국어)

### 요약

SessionEnd 훅이 orphaned teammate 세션을 정리할 때 사용자의 tmux 세션을 실수로 종료하는 문제를 수정하는 버그 수정 릴리즈입니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **SessionEnd tmux 정리**: SessionEnd 훅이 이제 orphaned 세션 정리 중 사용자의 실제 tmux 세션을 올바르게 보호합니다. `getCurrentTmuxSession()` 헬퍼 함수를 추가하여 `tmux display-message -p '#S'`로 현재 세션 이름을 가져옵니다. 정리 로직이 "(attached)" 상태를 확인하기 전에 현재 세션을 명시적으로 건너뛰도록 수정되어, Claude Code 세션 종료 시 사용자의 tmux 세션이 실수로 종료되는 문제를 방지합니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.5.0] - 2026-02-23

### Summary

Production-ready release for Agent Teams integration with Claude Code v2.1.30-45, featuring comprehensive quality hooks, persistent memory, CG Mode for cost-effective development, and enhanced workflow methodology including the Research-Plan-Annotate cycle from Boris Tane's development best practices.

### Added

- **Research-Plan-Annotate Cycle**: Implemented Boris Tane's development workflow with `research.md` artifact generation, 1-6 iteration annotation cycle before implementation, and implementation guards preventing premature code writing during planning phases. Deep reading patterns ("IN DEPTH", "IN GREAT DETAIL") and reference implementation search integrated across plan, moai, and team workflows.
- **Agent Teams Quality Hooks**: TeammateIdle hook now enforces LSP quality gates when diagnostics baseline exists. TaskCompleted hook verifies SPEC documents exist when task references SPEC-XXX patterns. All validation uses graceful degradation.
- **Agent Persistent Memory**: All 28 agent templates now have consistent `memory` frontmatter. Manager/Expert/Team agents use `project` scope; Builder agents use `user` scope for cross-project learning.
- **Settings Enhancements**: `spinnerTipsOverride` with 8 MoAI-specific workflow tips added to settings.json template (Claude Code v2.1.45).
- **Task Metrics Logging**: PostToolUse hook now captures Task tool metrics (tokens, tool uses, duration) to `.moai/logs/task-metrics.jsonl` for session analytics.
- **MCP OAuth Support**: Added `.moai/docs/MCP_OAUTH_SETUP.md` guide for configuring OAuth credentials for MCP servers (Slack, GitHub, Sentry).
- **Troubleshooting Guide**: Added troubleshooting section to CLAUDE.md covering `/debug` command usage, common Agent Teams issues, and PDF pagination tips.
- **Test Coverage**: Comprehensive test suite added across all packages to meet 85%+ coverage threshold. Key packages: internal/hook (3 subpackages), internal/shell, internal/template, internal/rank, internal/github, internal/merge, internal/update, pkg/models, internal/ui, internal/core/git, internal/core/project, internal/hook/agents, internal/hook/lifecycle. internal/cli improved from 60.6% → 73.3% (OAuth browser-flow functions excluded from automated testing).
- **Binary TDD/DDD Methodology**: Removed hybrid mode, implemented clean binary selection between TDD (default for new code) and DDD (for legacy refactoring). Simplified development mode selection and documentation.
- **GLM Team Flag**: Added `moai glm --team` flag for Agent Teams parallel execution in GLM workflow.
- **CG Mode (Claude + GLM)**: Implemented `/moai --team` workflow with Leader (Claude, current tmux pane) + Teammates (GLM, new tmux panes) architecture for cost-effective development. Uses tmux session-level env isolation (`CLAUDE_CODE_TEAMMATE_DISPLAY=tmux`) so teammates inherit GLM API env vars while leader stays on Claude. 60-70% cost reduction for implementation-heavy tasks.
- **Go 1.26 Upgrade**: Integrated Green Tea GC with 10-40% memory improvement, goroutine leak profiler, and modernization utilities.
- **Agent Documentation**: Corrected expert agent count documentation (8 to 9 agents), added per-agent model assignment tables by tier, fixed team agent model values.

### Fixed

- **SubagentStop Hook**: `moai hook subagent-stop` was not registered as a CLI subcommand, causing silent failures. Now properly registered (Claude Code v2.1.33).
- **SessionEnd Cleanup**: SessionEnd hook now automatically removes orphaned team directories and tmux sessions from interrupted Agent Teams workflows.
- **Settings Format**: Changed `spinnerTipsOverride` from array to object format for consistency.
- **Lint Quality**: Replaced all `WriteString(fmt.Sprintf())` patterns with `fmt.Fprintf()` for improved code quality and performance.
- **Unused Settings Fields**: Removed unused `spinnerTipsEnabled`, `spinnerTipsOverride` (reverted), `enabledPlugins`, `extraKnownMarketplaces` from template to reduce configuration bloat.
- **Model Inheritance**: Removed `inherit` model option, fixed `team.enabled` default setting.

### Changed

- **HookInput**: Added `TeamName`, `TeammateName`, `TaskID`, `TaskSubject`, `TaskDescription` fields for Agent Teams event handling (Claude Code v2.1.33).
- **Development Methodology**: Binary methodology selection replacing hybrid mode for clearer workflow adoption and documentation.

---

## [2.4.7] - 2026-02-18

### Summary

This patch release fully resolves the `moai init` / `moai update -c` interactive wizard scrolling regression introduced by charmbracelet/huh v0.8.x. Three compounding bugs were eliminated: shared viewport state across questions, incorrect height calculation, and `OptionsFunc` forcing a non-zero height that caused `updateViewportHeight()` to reset `YOffset` on every keypress — scrolling the selected item to the top and hiding options above the cursor.

### Breaking Changes

None.

### Fixed

- **Wizard scroll regression** (`moai init`, `moai update -c`): Pressing the down arrow no longer scrolls the option list — only the cursor highlight moves. Root cause was `OptionsFunc()` forcing `s.height = defaultHeight(10)` in huh v0.8.x, which triggered `updateViewportHeight()` to reset `viewport.YOffset = s.selected` on every `Update()` call. Fixed by replacing `OptionsFunc` with static `Options()` and removing any explicit `Height()` call, keeping `s.height = 0` so the auto-size branch sizes the viewport to exactly the number of options and never resets `YOffset`.
- **Wizard shared viewport**: Each wizard question now runs as its own independent `huh.Form` instead of sharing a single form with multiple groups. This eliminates cross-question viewport state pollution in huh v0.8.x.
- **Wizard height calculation**: `Select.Height()` was previously set incorrectly (options count only), ignoring title and description overhead. The explicit height call has been removed entirely in favour of huh's auto-sizing.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.7] - 2026-02-18 (한국어)

### 요약

이번 패치 릴리즈는 charmbracelet/huh v0.8.x에서 발생하는 `moai init` / `moai update -c` 인터랙티브 위저드 스크롤 버그를 완전히 수정합니다. 세 가지 복합적인 버그를 제거했습니다: 질문 간 뷰포트 상태 공유, 잘못된 높이 계산, 그리고 `OptionsFunc`가 강제로 `s.height` 값을 설정해 매번 `updateViewportHeight()`가 `YOffset`을 리셋하여 선택 항목이 항상 맨 위로 스크롤되던 문제입니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 수정됨 (Fixed)

- **위저드 스크롤 버그** (`moai init`, `moai update -c`): 이제 아래 화살표를 누르면 옵션 목록이 스크롤되지 않고 커서 하이라이트만 이동합니다. 근본 원인은 huh v0.8.x에서 `OptionsFunc()`가 `s.height = defaultHeight(10)`으로 강제 설정하여 `Update()` 호출마다 `viewport.YOffset = s.selected`가 리셋되었기 때문입니다. `OptionsFunc` 대신 정적 `Options()`를 사용하고 `Height()` 호출을 제거하여 `s.height = 0`을 유지함으로써 뷰포트가 정확히 옵션 개수만큼 자동 크기 조정되어 `YOffset`이 강제 초기화되지 않도록 수정했습니다.
- **위저드 공유 뷰포트**: 이제 각 위저드 질문이 단일 폼 내 여러 그룹을 공유하는 대신 독립적인 `huh.Form`으로 실행됩니다. huh v0.8.x에서의 질문 간 뷰포트 상태 오염이 제거됩니다.
- **위저드 높이 계산**: `Select.Height()`가 제목 및 설명 오버헤드를 무시하고 옵션 수만으로 잘못 계산되던 문제를 수정했습니다. 이제 명시적 `Height()` 호출을 제거하고 huh 자동 크기 조정을 활용합니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.6] - 2026-02-18

### Summary

This patch release fixes the `moai init` / `moai update -c` interactive wizard where the **max teammates** and **default model** select fields displayed only one option and became unresponsive after selection, caused by two compounding bugs in charmbracelet/huh v0.8.x. Also includes a major update to the MoAI template system with new lifecycle hooks, updated agent definitions, and new skill modules.

### Breaking Changes

None.

### Added

- **New lifecycle hooks**: `Notification`, `PermissionRequest`, `PostToolUseFailure`, `SubagentStart`, `SubagentStop`, `TaskCompleted`, `TeammateIdle`, `UserPromptSubmit` hooks for comprehensive Claude Code event coverage.
- **New skill modules**: `moai-foundation-thinking` skill with critical-evaluation, diverge-converge, and deep-questioning modules; `token-optimization` module for context management; `design-system-tokens` module for UI/UX.
- **Git convention config**: New `git-convention.yaml` section for configurable commit message standards.

### Changed

- **Agent definitions updated**: All 19 MoAI agents (manager, expert, builder, team) updated with latest capabilities and descriptions.
- **Workflow commands updated**: `github`, `loop`, `team-plan`, `team-review`, `team-run` workflows updated with Agent Teams improvements.
- **MCP integration rules updated**: Improved context7, pencil, and claude-in-chrome integration documentation.
- **Removed deprecated skill files**: Removed `moai-tool-ast-grep` rule files, `moai-workflow-testing` examples, and `moai-foundation-quality` scripts that are no longer bundled as static files.

### Fixed

- **Wizard viewport freeze (Height(0))**: The select field height was set to `0`, which in huh v0.8.x means a viewport of zero lines. Only the currently-selected item was rendered and the list could not scroll. Fixed to `Height(max(len(options), 3))` so all options are always fully visible.
- **Wizard YOffset scroll bug (max teammates)**: When the default option is at index N, huh v0.8.x unconditionally sets `viewport.YOffset = N`, hiding all options above it. Fixed by reordering `max_teammates` options descending (10 → 2) so the default ("10") is always at index 0.
- **Wizard YOffset scroll bug (default model)**: Same YOffset issue affected the `default_model` field. Fixed by reordering options so "Sonnet (Balanced)" (the default) appears first.
- **MaxTeammates comment**: Type comment incorrectly stated "2-5"; corrected to "2-10".
- **Korean translation typos**: Size-description labels in the `max_teammates` Korean translation were missing "모" (e.g., "소규 팀" → "소규모 팀", "중대규 팀" → "중대규모 팀").

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.6] - 2026-02-18 (한국어)

### 요약

이번 패치 릴리즈는 `moai init` / `moai update -c` 위자드에서 **최대 팀원 수**와 **기본 모델** 선택 필드가 하나의 항목만 표시되고 선택 후 멈추던 버그를 수정합니다. charmbracelet/huh v0.8.x에서 발생한 두 가지 버그가 복합적으로 작용한 것이었습니다. 또한 새 라이프사이클 훅, 업데이트된 에이전트 정의, 새 스킬 모듈을 포함한 MoAI 템플릿 시스템 대규모 업데이트도 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **새 라이프사이클 훅**: Claude Code 이벤트 완전 커버를 위해 `Notification`, `PermissionRequest`, `PostToolUseFailure`, `SubagentStart`, `SubagentStop`, `TaskCompleted`, `TeammateIdle`, `UserPromptSubmit` 훅 추가.
- **새 스킬 모듈**: 비판적 평가·발산-수렴·심층 질문 모듈이 포함된 `moai-foundation-thinking` 스킬; 컨텍스트 관리를 위한 `token-optimization` 모듈; UI/UX용 `design-system-tokens` 모듈.
- **Git 컨벤션 설정**: 커밋 메시지 기준 설정을 위한 새 `git-convention.yaml` 섹션 추가.

### 변경됨 (Changed)

- **에이전트 정의 업데이트**: 모든 19개 MoAI 에이전트(manager, expert, builder, team)가 최신 기능과 설명으로 업데이트.
- **워크플로우 커맨드 업데이트**: `github`, `loop`, `team-plan`, `team-review`, `team-run` 워크플로우가 Agent Teams 개선사항을 반영하여 업데이트.
- **MCP 통합 규칙 업데이트**: context7, pencil, claude-in-chrome 통합 문서 개선.
- **더 이상 사용되지 않는 스킬 파일 제거**: 더 이상 정적 파일로 번들되지 않는 `moai-tool-ast-grep` 규칙 파일, `moai-workflow-testing` 예제, `moai-foundation-quality` 스크립트 제거.

### 수정됨 (Fixed)

- **위자드 뷰포트 멈춤 (Height(0))**: 선택 필드의 높이가 `0`으로 설정되어 huh v0.8.x에서 뷰포트 크기가 0줄이 되던 문제를 수정했습니다. 현재 선택된 항목만 렌더링되고 스크롤이 불가능했습니다. `Height(max(옵션 수, 3))`으로 변경하여 항상 모든 옵션이 표시됩니다.
- **위자드 YOffset 스크롤 버그 (최대 팀원 수)**: 기본값 옵션이 인덱스 N에 있을 때 huh v0.8.x가 `viewport.YOffset = N`으로 설정하여 그 위의 옵션이 모두 숨겨지던 문제를 수정했습니다. `max_teammates` 옵션을 내림차순(10 → 2)으로 재정렬하여 기본값("10")이 항상 인덱스 0에 위치하도록 했습니다.
- **위자드 YOffset 스크롤 버그 (기본 모델)**: `default_model` 필드에서도 동일한 YOffset 문제가 발생했습니다. "Sonnet (균형)" (기본값)이 첫 번째로 오도록 옵션 순서를 변경했습니다.
- **MaxTeammates 주석 오류**: 타입 주석에 "2-5"로 잘못 표기되어 있던 것을 "2-10"으로 수정했습니다.
- **한국어 번역 오타**: `max_teammates`의 한국어 번역에서 크기 설명 레이블에 "모"가 누락되어 있었습니다 (예: "소규 팀" → "소규모 팀", "중대규 팀" → "중대규모 팀").

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.5] - 2026-02-17

### Summary

This patch release fixes two data-loss bugs in `moai update`: user-added `.gitignore` patterns are now preserved via EntryMerge after template sync, and user-created custom config sections (not present in the template) are correctly restored instead of being silently dropped.

### Breaking Changes

None.

### Added

- **`.gitignore` EntryMerge**: User-specific patterns are automatically detected after template deploy and appended under a `# User Custom Patterns` section, preventing data loss on `moai update`.

### Fixed

- **`.gitignore` overwritten**: `moai update` previously overwrote the entire `.gitignore` with the template version, discarding user-added patterns. Now user patterns are preserved.
- **Custom config sections dropped**: User-created files in `.moai/config/sections/` (e.g., `my-custom.yaml`) were silently dropped when the template did not include them. They are now restored after template sync.

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.5] - 2026-02-17 (한국어)

### 요약

이번 패치 릴리즈는 `moai update` 실행 시 데이터가 손실되던 두 가지 버그를 수정합니다. 사용자가 추가한 `.gitignore` 패턴이 EntryMerge를 통해 보존되며, 사용자가 생성한 커스텀 config 섹션도 템플릿에 없더라도 올바르게 복원됩니다.

### 주요 변경 사항 (Breaking Changes)

없음.

### 추가됨 (Added)

- **`.gitignore` EntryMerge**: 템플릿 배포 후 사용자가 추가한 패턴을 자동으로 감지하여 `# User Custom Patterns` 섹션 아래에 추가합니다. `moai update` 시 데이터 손실이 방지됩니다.

### 수정됨 (Fixed)

- **`.gitignore` 덮어쓰기 문제**: `moai update`가 사용자가 추가한 패턴을 포함한 `.gitignore` 전체를 템플릿 버전으로 덮어쓰던 문제를 수정했습니다.
- **커스텀 config 섹션 손실 문제**: `.moai/config/sections/`에 사용자가 직접 생성한 파일(예: `my-custom.yaml`)이 새 템플릿에 해당 파일이 없으면 조용히 사라지던 문제를 수정했습니다. 이제 템플릿 배포 후에도 올바르게 복원됩니다.

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.4] - 2026-02-17

### Summary

GitHub Workflow integration release with SPEC-GITHUB-WORKFLOW (Milestones 1-6) enabling automated PR review, issue management, and SPEC linking. Includes project-scale-aware DDD test strategies, wizard viewport fix, and multiple bug fixes for worktree, GLM, and install.sh.

### Breaking Changes

None

### Added

- **GitHub Workflow Integration**: Full SPEC-GITHUB-WORKFLOW implementation (Milestones 1-6) with automated PR review, issue closing, SPEC linking, and concurrent GitHub operations (`internal/github/`)
- **GitHub CLI Command**: New `moai github` command for managing issues and PRs from the CLI (`internal/cli/github.go`)
- **Project-Scale-Aware DDD Test Strategy**: DDD agent now adapts test strategies based on project size (small/medium/large/xlarge) for optimized coverage
- **Tmux Session Management**: New `internal/tmux/` package for robust Agent Teams tmux session detection and management
- **Workflow Orchestrator**: New `internal/workflow/worktree_orchestrator.go` for improved worktree workflow coordination
- **i18n Templates**: Internationalization support for template messages (`internal/i18n/templates.go`)
- **Branch Detector**: New `internal/git/branch_detector.go` for reliable branch detection

### Changed

- **Output Style**: MoAI output style templates updated to English-first format
- **GLM Status Line**: `status_line.sh` now loads GLM environment variables for Agent Teams tmux mode
- **Manager-DDD Agent**: Updated with project-scale-aware test strategy documentation

### Fixed

- **Wizard Viewport**: Reordered language options to fix huh library viewport YOffset rendering bug in `moai update -c` wizard
- **GitHub Quality**: Resolved 11 code quality suggestions from PR #390 code review
- **Worktree SPEC-ID**: Fixed SPEC-ID matching and path standardization inconsistencies in worktree commands
- **GLM tmux Mode**: Fixed GLM environment variables not loading in status_line.sh for Agent Teams tmux mode
- **Install Script**: Removed unsafe sed-based JSON editing from install.sh for improved security

### Dependencies

- Bumped `github.com/charmbracelet/bubbles` to v1.0.0
- Bumped `golang.org/x/text` (minor/patch)
- Bumped `golang.org/x/net` to v0.38.0

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.4] - 2026-02-17 (한국어)

### 요약

GitHub 워크플로우 통합 릴리즈로, SPEC-GITHUB-WORKFLOW (Milestone 1-6)를 통해 자동화된 PR 리뷰, 이슈 관리, SPEC 연결 기능을 제공합니다. 프로젝트 규모 인식 DDD 테스트 전략, wizard viewport 버그 수정, worktree/GLM/install.sh 다수 버그 픽스를 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **GitHub 워크플로우 통합**: SPEC-GITHUB-WORKFLOW 전체 구현 (Milestone 1-6) — 자동 PR 리뷰, 이슈 종료, SPEC 연결, 동시 GitHub 작업 (`internal/github/`)
- **GitHub CLI 명령어**: CLI에서 이슈 및 PR을 관리하는 새로운 `moai github` 명령어 (`internal/cli/github.go`)
- **프로젝트 규모 인식 DDD 테스트 전략**: DDD 에이전트가 프로젝트 크기(small/medium/large/xlarge)에 따라 테스트 전략을 최적화
- **Tmux 세션 관리**: Agent Teams tmux 세션 감지 및 관리를 위한 새로운 `internal/tmux/` 패키지
- **워크플로우 오케스트레이터**: 개선된 worktree 워크플로우 조율을 위한 `internal/workflow/worktree_orchestrator.go`
- **i18n 템플릿**: 템플릿 메시지 다국어 지원 (`internal/i18n/templates.go`)
- **브랜치 감지기**: 안정적인 브랜치 감지를 위한 `internal/git/branch_detector.go`

### 변경됨 (Changed)

- **출력 스타일**: MoAI 출력 스타일 템플릿을 영문 우선 형식으로 업데이트
- **GLM 상태 라인**: `status_line.sh`에서 Agent Teams tmux 모드용 GLM 환경 변수 로드 지원 추가
- **Manager-DDD 에이전트**: 프로젝트 규모 인식 테스트 전략 문서 업데이트

### 수정됨 (Fixed)

- **Wizard Viewport**: `moai update -c` 위저드에서 huh 라이브러리 viewport YOffset 렌더링 버그 수정 (언어 옵션 순서 변경)
- **GitHub 코드 품질**: PR #390 코드 리뷰의 코드 품질 제안 11개 해결
- **Worktree SPEC-ID**: worktree 명령어에서 SPEC-ID 매칭 및 경로 표준화 불일치 수정
- **GLM tmux 모드**: Agent Teams tmux 모드에서 status_line.sh GLM 환경 변수 미로드 버그 수정
- **설치 스크립트**: install.sh에서 보안 취약한 sed 기반 JSON 편집 방식 제거

### 의존성 업데이트

- `github.com/charmbracelet/bubbles` v1.0.0으로 업데이트
- `golang.org/x/text` 마이너/패치 업데이트
- `golang.org/x/net` v0.38.0으로 업데이트

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.3] - 2026-02-16

### Summary

TUI modernization release with lipgloss rounded box design across all CLI commands and SKILL.md routing fix. This update replaces plain text output and `======` separators with modern styled cards for consistent user experience.

### Breaking Changes

None

### Added

- **Shared Rendering Primitives**: New `internal/cli/render.go` with reusable lipgloss functions (`renderCard`, `renderSuccessCard`, `renderInfoCard`, `renderStatusLine`, `renderSummaryLine`, `renderKeyValue`)
- **Worktree Rendering**: New `internal/cli/worktree/render.go` for worktree-specific styled output
- **Styled Console Reporter**: Added `StyledConsoleReporter` in `internal/core/project/reporter.go` for colored project detection output

### Changed

- **CLI Output**: Modernized all CLI commands with rounded box design:
  - `moai status`: Replaced `======` separator with styled card showing project info
  - `moai doctor`: Replaced `======` separator with styled card + colored status icons
  - `moai version`: Added styled card for version display
  - `moai init`: Added styled success card for completion output
  - `moai cc` / `moai glm`: Replaced emoji with styled success cards
  - `moai hook list`: Replaced `======` separator with styled card
  - `moai rank`: All 7 subcommands now use styled cards
  - `moai worktree`: All subcommands (list, status, new, done, clean, sync) use styled output
- **SelectHeight**: Added chrome padding compensation (2 lines) to prevent UI cutoff in huh selectors

### Fixed

- **SKILL.md Routing**: Prevented `$ARGUMENTS` inline expansion from contaminating intent router logic by isolating to dedicated "Raw User Input" section with [HARD] enforcement on Priority 1 subcommand matching

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.3] - 2026-02-16 (한국어)

### 요약

모든 CLI 명령어에 lipgloss 라운드 박스 디자인을 적용하고 SKILL.md 라우팅 버그를 수정한 TUI 현대화 릴리스입니다. 이 업데이트는 plain text 출력과 `======` 구분선을 모던한 스타일 카드로 대체하여 일관된 사용자 경험을 제공합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **공유 렌더링 기능**: `internal/cli/render.go`에 재사용 가능한 lipgloss 함수 추가 (`renderCard`, `renderSuccessCard`, `renderInfoCard`, `renderStatusLine`, `renderSummaryLine`, `renderKeyValue`)
- **Worktree 렌더링**: worktree 전용 스타일 출력을 위한 `internal/cli/worktree/render.go` 추가
- **스타일 콘솔 리포터**: 프로젝트 감지 출력에 색상을 적용한 `StyledConsoleReporter` 추가

### 변경됨 (Changed)

- **CLI 출력**: 모든 CLI 명령어 라운드 박스 디자인으로 현대화:
  - `moai status`: `======` 구분선을 프로젝트 정보 스타일 카드로 대체
  - `moai doctor`: `======` 구분선을 색상 상태 아이콘이 있는 스타일 카드로 대체
  - `moai version`: 버전 표시용 스타일 카드 추가
  - `moai init`: 완료 출력용 스타일 성공 카드 추가
  - `moai cc` / `moai glm`: 이모지를 스타일 성공 카드로 대체
  - `moai hook list`: `======` 구분선을 스타일 카드로 대체
  - `moai rank`: 7개 서브커맨드 모두 스타일 카드 사용
  - `moai worktree`: 모든 서브커맨드 (list, status, new, done, clean, sync) 스타일 출력 사용
- **SelectHeight**: huh 셀렉터에서 UI 잘림 방지를 위한 chrome 패딩 보정 (2줄) 추가

### 수정됨 (Fixed)

- **SKILL.md 라우팅**: `$ARGUMENTS` 인라인 확장이 인텐트 라우터 로직을 오염시키는 문제 수정 - 전용 "Raw User Input" 섹션으로 격리하고 Priority 1 서브커맨드 매칭에 [HARD] 강제 적용

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.2] - 2026-02-15

### Summary

Visual polish release with adaptive terminal colors and modern Unicode symbols. This update enhances CLI output readability by migrating to `lipgloss.AdaptiveColor` for automatic light/dark theme support and replacing ASCII status indicators with professional Unicode glyphs across all interactive components.

### Breaking Changes

None

### Changed

- **Adaptive Colors**: Migrated all hardcoded hex colors to `lipgloss.AdaptiveColor` for automatic light/dark terminal support across banner, doctor, update, and merge confirm components
- **Status Icons**: Replaced ASCII status text (`[OK]`, `[WARN]`, `[FAIL]`) with colored Unicode icons (✓, ⚠, ✗) in `moai doctor` output
- **CLI Output Styles**: Extracted shared lipgloss style variables (`cliSuccess`, `cliError`, `cliMuted`, `cliPrimary`, `cliBorder`) and symbol functions for consistent themed output in `moai update`
- **Select Height**: Added dynamic height auto-sizing for Select and MultiSelect fields based on option count (max 10) to prevent terminal overflow
- **Selector Prefixes**: Updated selector cursor from default to `▸` and checkbox prefixes from `[x]`/`[ ]` to `◆`/`◇` in both UI and wizard themes
- **Rounded Borders**: Changed merge analysis table corners from sharp (`┌┐└┘`) to rounded (`╭╮╰╯`) style
- **Model Policy Box**: Replaced manual ASCII box drawing with lipgloss `RoundedBorder` styled box in `moai update` output

### Fixed

- **Unused Constant**: Removed orphaned `claudeTerraCotta` constant from banner.go after AdaptiveColor migration

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

**Migrating from Python Version (v1.x)**:
1. Uninstall Python version: `uv tool uninstall moai-adk`
2. Install Go Edition: `curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash`
3. Update project templates: `moai init`

---

## [2.4.2] - 2026-02-15 (한국어)

### 요약

적응형 터미널 색상과 현대적인 유니코드 기호를 사용한 비주얼 개선 릴리스입니다. 이 업데이트는 자동 라이트/다크 테마 지원을 위해 `lipgloss.AdaptiveColor`로 마이그레이션하고 모든 대화형 컴포넌트에서 ASCII 상태 표시기를 전문적인 유니코드 글리프로 교체하여 CLI 출력 가독성을 향상시킵니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 변경됨 (Changed)

- **적응형 색상**: 배너, doctor, update, merge confirm 컴포넌트 전체에서 자동 라이트/다크 터미널 지원을 위해 모든 하드코딩된 hex 색상을 `lipgloss.AdaptiveColor`로 마이그레이션
- **상태 아이콘**: `moai doctor` 출력에서 ASCII 상태 텍스트 (`[OK]`, `[WARN]`, `[FAIL]`)를 컬러 유니코드 아이콘 (✓, ⚠, ✗)으로 교체
- **CLI 출력 스타일**: `moai update`에서 일관된 테마 출력을 위해 공유 lipgloss 스타일 변수 (`cliSuccess`, `cliError`, `cliMuted`, `cliPrimary`, `cliBorder`) 및 심볼 함수 추출
- **선택 높이**: 터미널 오버플로우를 방지하기 위해 옵션 수에 따라 Select 및 MultiSelect 필드의 동적 높이 자동 조정 추가 (최대 10)
- **선택자 접두사**: UI 및 마법사 테마 모두에서 선택자 커서를 기본값에서 `▸`로, 체크박스 접두사를 `[x]`/`[ ]`에서 `◆`/`◇`로 업데이트
- **둥근 테두리**: 병합 분석 테이블 모서리를 예리한 스타일 (`┌┐└┘`)에서 둥근 스타일 (`╭╮╰╯`)로 변경
- **모델 정책 박스**: `moai update` 출력에서 수동 ASCII 박스 그리기를 lipgloss `RoundedBorder` 스타일 박스로 교체

### 수정됨 (Fixed)

- **사용하지 않는 상수**: AdaptiveColor 마이그레이션 후 banner.go에서 고아가 된 `claudeTerraCotta` 상수 제거

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

**Python 버전(v1.x)에서 마이그레이션**:
1. Python 버전 제거: `uv tool uninstall moai-adk`
2. Go 에디션 설치: `curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash`
3. 프로젝트 템플릿 업데이트: `moai init`

---

## [2.4.1] - 2026-02-15

### Summary

Maintenance release modernizing the TUI system with charmbracelet/huh form components and glamour markdown renderer (SPEC-UI-003). This update replaces custom Bubble Tea components with industry-standard Charmbracelet libraries, providing modern visual styling, accessibility support, and enhanced user experience.

### Breaking Changes

None

### Added

- **Modern Form System** (SPEC-UI-003): Replaced custom form components with charmbracelet/huh v0.8.0
  - Select/MultiSelect: Modern visual styling with rounded borders and proper keyboard navigation
  - Input: Bordered containers with placeholder text and validation callbacks
  - Confirm: Clear visual distinction for yes/no prompts
  - Form+Groups: Page-based progression for multi-step wizards
  - MoAI custom theme: Brand colors (#DA7756 primary, #7C3AED secondary) with adaptive light/dark support
  - Accessibility: NoColor mode for headless/CI environments
- **Markdown Rendering**: charmbracelet/glamour v0.10.0 for terminal documentation display
  - Auto light/dark theme detection
  - Syntax highlighting for code blocks
  - Professional layout for help text
- **Layout Enhancement**: Lipgloss advanced features (RoundedBorder, responsive width, terminal detection)
- **Progress Components**: Animated Bubbles spinner and progress bar with MoAI theme colors

### Changed

- **Form Components**: Rewrote selector.go, checkbox.go, prompt.go, progress.go with huh/Bubbles integration
- **Wizard System**: Rebuilt wizard.go using huh.Form + huh.Group for multi-step forms
- **Test Architecture**: Extracted pure functions (buildSelectField, buildMultiSelectField, buildInputField, buildConfirmField) for 100% test coverage
- **File Cleanup**: Removed runner.go and runner_test.go (replaced by huh Form.Run pattern)

### Fixed

- **Lint QF1001**: Applied De Morgan's law to internal/rank/transcript.go:218 for staticcheck compliance
- **Windows Test Compatibility**: Added USERPROFILE environment variable to 5 test functions in internal/rank/transcript_test.go

### Testing

- UI package: 73.5% coverage (220 tests passing with -race flag)
- Wizard package: 80.5% coverage
- Pure functions: 100% coverage
- Coverage gap: Structural limitation from TTY-dependent wrappers (huh.Form.Run, tea.Program.Run)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.1] - 2026-02-15 (한국어)

### 요약

charmbracelet/huh 폼 컴포넌트와 glamour 마크다운 렌더러를 사용하여 TUI 시스템을 현대화하는 유지보수 릴리스입니다 (SPEC-UI-003). 이 업데이트는 커스텀 Bubble Tea 컴포넌트를 업계 표준 Charmbracelet 라이브러리로 교체하여 현대적인 비주얼 스타일링, 접근성 지원 및 향상된 사용자 경험을 제공합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **현대적인 폼 시스템** (SPEC-UI-003): 커스텀 폼 컴포넌트를 charmbracelet/huh v0.8.0으로 교체
  - Select/MultiSelect: 둥근 테두리와 적절한 키보드 탐색을 갖춘 현대적인 비주얼 스타일링
  - Input: 플레이스홀더 텍스트 및 유효성 검사 콜백이 있는 테두리 컨테이너
  - Confirm: 예/아니오 프롬프트의 명확한 시각적 구분
  - Form+Groups: 다단계 마법사를 위한 페이지 기반 진행
  - MoAI 커스텀 테마: 적응형 라이트/다크 지원이 포함된 브랜드 색상 (#DA7756 기본, #7C3AED 보조)
  - 접근성: 헤드리스/CI 환경을 위한 NoColor 모드
- **마크다운 렌더링**: 터미널 문서 표시를 위한 charmbracelet/glamour v0.10.0
  - 자동 라이트/다크 테마 감지
  - 코드 블록 구문 강조
  - 도움말 텍스트의 전문적인 레이아웃
- **레이아웃 개선**: Lipgloss 고급 기능 (RoundedBorder, 반응형 너비, 터미널 감지)
- **진행 컴포넌트**: MoAI 테마 색상이 적용된 애니메이션 Bubbles 스피너 및 진행 표시줄

### 변경됨 (Changed)

- **폼 컴포넌트**: huh/Bubbles 통합으로 selector.go, checkbox.go, prompt.go, progress.go 재작성
- **마법사 시스템**: 다단계 폼을 위해 huh.Form + huh.Group을 사용하여 wizard.go 재구축
- **테스트 아키텍처**: 100% 테스트 커버리지를 위해 순수 함수 추출 (buildSelectField, buildMultiSelectField, buildInputField, buildConfirmField)
- **파일 정리**: runner.go 및 runner_test.go 제거 (huh Form.Run 패턴으로 대체)

### 수정됨 (Fixed)

- **Lint QF1001**: staticcheck 준수를 위해 internal/rank/transcript.go:218에 De Morgan의 법칙 적용
- **Windows 테스트 호환성**: internal/rank/transcript_test.go의 5개 테스트 함수에 USERPROFILE 환경 변수 추가

### 테스트 (Testing)

- UI 패키지: 73.5% 커버리지 (-race 플래그로 220개 테스트 통과)
- 마법사 패키지: 80.5% 커버리지
- 순수 함수: 100% 커버리지
- 커버리지 갭: TTY 의존 래퍼의 구조적 제한 (huh.Form.Run, tea.Program.Run)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.4.0] - 2026-02-14

### Summary

Feature release adding statusline segment configuration (SPEC-STATUSLINE-001) and enhanced workflow documentation. Users can now customize which statusline segments are displayed via the init/update wizard or YAML configuration.

### Breaking Changes

None

### Added

- **Statusline Segment Configuration** (SPEC-STATUSLINE-001): Configurable statusline display with preset and per-segment controls
  - 4 presets: Full (all 8 segments), Compact (model, context, git status, branch), Minimal (model, context only), Custom (pick individual segments)
  - New `statusline.yaml` configuration file at `.moai/config/sections/statusline.yaml`
  - Wizard questions for statusline preset and individual segment selection during `moai init` / `moai update -c`
  - i18n translations for Korean, Japanese, and Chinese
  - Renderer-level segment filtering with backward compatibility (nil config = all enabled)
- **Extended Run Workflow Quality Checks**: Code complexity analysis, dead code detection, side effect analysis
- **Post-Implementation Review Phase** (Phase 2.7): Multi-dimensional review iteration for run workflow
- **Deployment Readiness Check** (Phase 0 in sync workflow): Test verification, migration detection, backward compatibility
- **UX Review Perspective**: Added 4th review dimension to team-review workflow

### Changed

- **Development Mode**: Changed default `development_mode` from `ddd` to `hybrid` in quality.yaml
- **Model Policy Application**: `ApplyModelPolicy` now applies for all policy values including "high"
- **Agent Teams Documentation**: Added token cost awareness, team workflow references, and known limitations to spec-workflow.md
- **Workflow Skills**: Updated run.md (v2.1.0), sync.md (v3.1.0), team-review.md (v1.1.0)
- **Permission Mode**: Changed default `permissions.defaultMode` from `default` to `acceptEdits` for smoother agent workflows

### Fixed

- **Model Policy Skip Bug**: Fixed `moai init` and `moai update` skipping model policy application when policy was "high"

### Testing

- Statusline: 281 lines covering segment config loading, filtering, preset mapping, and renderer behavior
- Wizard: 356 lines covering question generation, conditional visibility, and answer saving
- Update command: 261 lines covering preset-to-segments conversion and config file writing

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.4.0] - 2026-02-14 (한국어)

### 요약

상태표시줄 세그먼트 설정(SPEC-STATUSLINE-001)과 향상된 워크플로우 문서화를 포함하는 기능 릴리스입니다. init/update 마법사 또는 YAML 설정을 통해 표시할 상태표시줄 세그먼트를 사용자 정의할 수 있습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **상태표시줄 세그먼트 설정** (SPEC-STATUSLINE-001): 프리셋 및 개별 세그먼트 제어를 통한 상태표시줄 표시 설정
  - 4가지 프리셋: Full (8개 전체), Compact (모델, 컨텍스트, git 상태, 브랜치), Minimal (모델, 컨텍스트만), Custom (개별 선택)
  - `.moai/config/sections/statusline.yaml` 설정 파일 추가
  - `moai init` / `moai update -c` 실행 시 상태표시줄 프리셋 및 개별 세그먼트 선택 마법사
  - 한국어, 일본어, 중국어 i18n 번역 지원
  - 하위 호환성을 유지하는 렌더러 수준 세그먼트 필터링 (nil 설정 = 전체 활성화)
- **확장된 Run 워크플로우 품질 검사**: 코드 복잡도 분석, 데드 코드 감지, 부작용 분석
- **구현 후 리뷰 단계** (Phase 2.7): run 워크플로우의 다차원 리뷰 반복
- **배포 준비 검사** (sync 워크플로우 Phase 0): 테스트 검증, 마이그레이션 감지, 하위 호환성
- **UX 리뷰 관점**: team-review 워크플로우에 4번째 리뷰 차원 추가

### 변경됨 (Changed)

- **개발 모드**: quality.yaml의 기본 `development_mode`가 `ddd`에서 `hybrid`로 변경
- **모델 정책 적용**: `ApplyModelPolicy`가 "high" 포함 모든 정책 값에 대해 적용
- **에이전트 팀 문서화**: spec-workflow.md에 토큰 비용 인식, 팀 워크플로우 참조, 알려진 제한 사항 추가
- **워크플로우 스킬**: run.md (v2.1.0), sync.md (v3.1.0), team-review.md (v1.1.0) 업데이트
- **권한 모드**: 기본 `permissions.defaultMode`를 `default`에서 `acceptEdits`로 변경하여 에이전트 워크플로우 원활화

### 수정됨 (Fixed)

- **모델 정책 건너뛰기 버그**: 정책이 "high"일 때 `moai init`과 `moai update`에서 모델 정책 적용을 건너뛰는 문제 수정

### 테스트 (Testing)

- 상태표시줄: 세그먼트 설정 로딩, 필터링, 프리셋 매핑, 렌더러 동작 커버 281줄
- 마법사: 질문 생성, 조건부 표시, 응답 저장 커버 356줄
- 업데이트 명령: 프리셋→세그먼트 변환, 설정 파일 쓰기 커버 261줄

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.3.1] - 2026-02-12

### Summary

Patch release merging three feature PRs and a security fix. Adds Claude Code CLI transcript path support, GitLab/self-hosted GitLab instance support, git commit message convention validation system, and sessionID path traversal prevention.

### Breaking Changes

None

### Added

- **GitLab Support**: Self-hosted GitLab instance support in `moai init` wizard (#375)
  - Git strategy configuration for GitHub, GitLab, and self-hosted GitLab
  - New `git-strategy.yaml` configuration template
- **Commit Convention Validation**: Full commit message convention validation system (#374)
  - Support for conventional-commits, angular, karma, and custom conventions
  - Auto-detection from repository commit history
  - Pre-push hook handler (`moai hook pre-push`) for enforcement
  - `git-convention.yaml` template config with documentation
- **Claude Code CLI Transcript Paths**: Support for new Claude Code CLI transcript locations (#371)
  - Priority search: `~/.claude/projects/*/*.jsonl` > `~/.claude/transcripts/*.jsonl` > Desktop paths
  - Deduplication logic for multi-source transcript discovery

### Fixed

- **Security: Path Traversal Prevention**: Added `isValidSessionID()` validation to `FindTranscriptForSession()`
  - Rejects path traversal characters (`../`, `/`, `\`)
  - Enforces alphanumeric + hyphen/underscore whitelist
  - Max length validation (128 characters)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.3.1] - 2026-02-12 (한국어)

### 요약

3개의 기능 PR 머지와 보안 수정을 포함하는 패치 릴리스입니다. Claude Code CLI 트랜스크립트 경로 지원, GitLab/자체 호스팅 GitLab 인스턴스 지원, git 커밋 메시지 컨벤션 검증 시스템, sessionID 경로 탐색 방지 기능을 추가합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **GitLab 지원**: `moai init` 마법사에서 자체 호스팅 GitLab 인스턴스 지원 (#375)
  - GitHub, GitLab, 자체 호스팅 GitLab용 Git 전략 설정
  - 새로운 `git-strategy.yaml` 설정 템플릿
- **커밋 컨벤션 검증**: 완전한 커밋 메시지 컨벤션 검증 시스템 (#374)
  - conventional-commits, angular, karma, 커스텀 컨벤션 지원
  - 저장소 커밋 히스토리 자동 감지
  - 푸시 전 훅 핸들러 (`moai hook pre-push`) 적용
  - `git-convention.yaml` 템플릿 설정 및 문서
- **Claude Code CLI 트랜스크립트 경로**: 새로운 Claude Code CLI 트랜스크립트 위치 지원 (#371)
  - 우선순위 검색: `~/.claude/projects/*/*.jsonl` > `~/.claude/transcripts/*.jsonl` > Desktop 경로
  - 멀티 소스 트랜스크립트 중복 제거 로직

### 수정됨 (Fixed)

- **보안: 경로 탐색 방지**: `FindTranscriptForSession()`에 `isValidSessionID()` 검증 추가
  - 경로 탐색 문자 거부 (`../`, `/`, `\`)
  - 영숫자 + 하이픈/밑줄 화이트리스트 적용
  - 최대 길이 검증 (128자)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.3.0] - 2026-02-12

### Summary

Feature release introducing Model Policy internationalization with pricing tiers, critical --force flag bug fix, and comprehensive README documentation updates. This release adds full i18n support (Korean/Japanese/Chinese) for the model policy wizard with clear pricing tier indicators ($200/$100/$20 plans), fixes a version check bypass bug, and includes template cleanup.

### Breaking Changes

None

### Added

- **Model Policy i18n**: Full internationalization support for model policy wizard
  - Korean (ko), Japanese (ja), Chinese (zh) translations added
  - Pricing tier indicators: High (Max $200/mo), Medium (Max $100/mo), Low (Plus $20/mo)
  - Clear model descriptions: High uses opus, Medium mixes opus/sonnet/haiku, Low uses sonnet/haiku only
- **Model Policy Notice**: User-friendly configuration guide displayed after `moai update`
  - Shows pricing tiers and agent model assignments
  - Guides users to run `moai update -c` for reconfiguration

### Changed

- **README Updates**: Synchronized architecture and statistics across all 4 language versions (EN, KO, JA, ZH)
  - Added Model Policy section with comparison table
  - Updated agent counts and configuration examples
  - Replaced v2.2.3 warning notice with official documentation links

### Fixed

- **Critical: --force Flag Bug**: Fixed `--force` flag not bypassing version check in `moai update`
  - `moai update --force` now correctly forces template sync even when versions match
  - Applied to both `runTemplateSyncWithReporter` and `runTemplateSyncWithProgress`
- **Template Cleanup**: Removed outdated `workflow.yaml.tmpl` that was overwriting team settings
- **Hardcoded Paths**: Removed hardcoded `moai-adk-go` references and developer-specific paths
- **CI: Lint errcheck**: Fixed unchecked `os.Remove` return value in `sync_state_test.go`
- **CI: Windows test**: Fixed platform-specific binary name (`moai.exe`) in `updater_test.go`

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.3.0] - 2026-02-12 (한국어)

### 요약

요금제 계층이 포함된 모델 정책 국제화, 중요한 --force 플래그 버그 수정, 포괄적인 README 문서 업데이트를 도입하는 기능 릴리스입니다. 이 릴리스는 명확한 요금제 계층 표시기($200/$100/$20 플랜)가 포함된 모델 정책 마법사의 완전한 i18n 지원(한국어/일본어/중국어)을 추가하고, 버전 확인 우회 버그를 수정하며, 템플릿 정리를 포함합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **모델 정책 i18n**: 모델 정책 마법사의 완전한 국제화 지원
  - 한국어(ko), 일본어(ja), 중국어(zh) 번역 추가
  - 요금제 계층 표시기: High (Max $200/mo), Medium (Max $100/mo), Low (Plus $20/mo)
  - 명확한 모델 설명: High는 opus 사용, Medium은 opus/sonnet/haiku 혼합, Low는 sonnet/haiku만 사용
- **모델 정책 안내**: `moai update` 후 표시되는 사용자 친화적 설정 가이드
  - 요금제 계층 및 에이전트 모델 할당 표시
  - `moai update -c`를 실행하여 재구성하도록 사용자 안내

### 변경됨 (Changed)

- **README 업데이트**: 4개 언어 버전(EN, KO, JA, ZH) 전체에 걸친 아키텍처 및 통계 동기화
  - 비교 표가 포함된 모델 정책 섹션 추가
  - 에이전트 수 및 설정 예제 업데이트
  - v2.2.3 경고 알림을 공식 문서 링크로 교체

### 수정됨 (Fixed)

- **중요: --force 플래그 버그**: `moai update`에서 `--force` 플래그가 버전 확인을 우회하지 않던 버그 수정
  - `moai update --force`가 이제 버전이 일치해도 템플릿 동기화를 강제로 수행
  - `runTemplateSyncWithReporter` 및 `runTemplateSyncWithProgress` 모두에 적용
- **템플릿 정리**: 팀 설정을 덮어쓰던 오래된 `workflow.yaml.tmpl` 제거
- **하드코딩된 경로**: 하드코딩된 `moai-adk-go` 참조 및 개발자별 경로 제거
- **CI: Lint errcheck**: `sync_state_test.go`에서 `os.Remove` 반환값 미확인 수정
- **CI: Windows 테스트**: `updater_test.go`에서 플랫폼별 바이너리 이름(`moai.exe`) 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.8] - 2026-02-11

### Summary

Patch release upgrading default GLM models from glm-4.7 to glm-5 for enhanced AI capabilities. This update provides improved response quality and advanced features for Sonnet and Opus model tiers while maintaining the optimized glm-4.7-flashx for Haiku.

### Breaking Changes

None

### Changed

- **GLM Model Defaults**: Upgraded Sonnet and Opus models from glm-4.7 to glm-5
  - Sonnet: glm-4.7 → glm-5 (enhanced model performance and capabilities)
  - Opus: glm-4.7 → glm-5 (enhanced model performance and capabilities)
  - Haiku: glm-4.7-flashx (unchanged, optimized for speed and cost)
  - Applies only to default fallback values when GLM config is unavailable
  - Users with custom GLM configurations unaffected

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.8] - 2026-02-11 (한국어)

### 요약

glm-4.7에서 glm-5로 기본 GLM 모델을 업그레이드하여 향상된 AI 기능을 제공하는 패치 릴리스입니다. 이 업데이트는 Sonnet 및 Opus 모델 계층에 대한 향상된 응답 품질과 고급 기능을 제공하며, Haiku는 최적화된 glm-4.7-flashx를 유지합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 변경됨 (Changed)

- **GLM 모델 기본값**: Sonnet 및 Opus 모델을 glm-4.7에서 glm-5로 업그레이드
  - Sonnet: glm-4.7 → glm-5 (향상된 모델 성능 및 기능)
  - Opus: glm-4.7 → glm-5 (향상된 모델 성능 및 기능)
  - Haiku: glm-4.7-flashx (변경 없음, 속도 및 비용 최적화)
  - GLM 설정이 없을 때 기본 fallback 값에만 적용됨
  - 사용자 정의 GLM 설정이 있는 경우 영향 없음

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.7] - 2026-02-11

### Summary

Feature release adding worktree shell navigation, device tracking for MoAI Rank sync, and configuration cleanup. Introduces `moai worktree go` command for seamless shell directory switching, device-aware session management, and removes legacy LLM/pricing configs moved to Context7 MCP.

### Breaking Changes

- **Configuration Cleanup**: Removed llm.yaml and pricing.yaml configs (moved to Context7 MCP server)
  - Users must run `moai update` to sync template changes
  - Existing projects continue working with Context7 MCP integration

### Added

- **Worktree Shell Navigation**: New `moai worktree go <branch>` command for shell directory switching
  - Usage: `cd $(moai wt go my-branch)` for instant navigation to worktree directory
  - Comprehensive test coverage (+707 lines in worktree/subcommands_test.go)
  - Mock extension framework for testability
- **MoAI Rank Device Tracking**: Multi-device sync awareness for session management
  - New device.go module for device identification and tracking
  - Device-scoped session persistence and sync state
  - Comprehensive test coverage (device_test.go)
- **MoAI Rank Sync State**: Session state management for cross-device synchronization
  - New sync_state.go module for sync coordination
  - State persistence and recovery mechanisms
  - Comprehensive test coverage (sync_state_test.go)
- **Update Command Enhancements**: Major refactoring of update.go (+315 lines)
  - Enhanced error handling and retry logic
  - Improved update verification and rollback
  - Better progress reporting

### Changed

- **Worktree Configuration**: Improved config management and error handling
- **Rank Client**: Refactored with enhanced error handling and retry logic
- **Template Structure**: Consolidated configuration templates for better maintainability
- **Project Initializer**: Updated for new configuration layout

### Removed

- **Legacy Configuration Files**:
  - .moai/config/sections/llm.yaml (moved to Context7 MCP)
  - .moai/config/sections/pricing.yaml (moved to Context7 MCP)
  - internal/template/templates/.moai/config/config.yaml.tmpl (consolidated)
  - internal/template/templates/.moai/config/sections/llm.yaml (moved to Context7 MCP)
  - internal/template/templates/.moai/config/sections/pricing.yaml (moved to Context7 MCP)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.7] - 2026-02-11 (한국어)

### 요약

워크트리 셸 내비게이션, MoAI Rank 동기화를 위한 디바이스 추적, 설정 정리 기능이 추가된 기능 릴리스입니다. 원활한 셸 디렉토리 전환을 위한 `moai worktree go` 명령어를 도입하고, 디바이스 인식 세션 관리를 제공하며, Context7 MCP로 이동한 레거시 LLM/pricing 설정을 제거했습니다.

### 주요 변경 사항 (Breaking Changes)

- **설정 정리**: llm.yaml 및 pricing.yaml 설정 제거 (Context7 MCP 서버로 이동)
  - 사용자는 `moai update`를 실행하여 템플릿 변경 사항을 동기화해야 함
  - 기존 프로젝트는 Context7 MCP 통합으로 계속 작동

### 추가됨 (Added)

- **워크트리 셸 내비게이션**: 셸 디렉토리 전환을 위한 새로운 `moai worktree go <branch>` 명령어
  - 사용법: `cd $(moai wt go my-branch)`로 워크트리 디렉토리로 즉시 이동
  - 포괄적인 테스트 커버리지 (worktree/subcommands_test.go에 +707줄)
  - 테스트 가능성을 위한 Mock 확장 프레임워크
- **MoAI Rank 디바이스 추적**: 세션 관리를 위한 다중 디바이스 동기화 인식
  - 디바이스 식별 및 추적을 위한 새로운 device.go 모듈
  - 디바이스 범위 세션 지속성 및 동기화 상태
  - 포괄적인 테스트 커버리지 (device_test.go)
- **MoAI Rank 동기화 상태**: 크로스 디바이스 동기화를 위한 세션 상태 관리
  - 동기화 조정을 위한 새로운 sync_state.go 모듈
  - 상태 지속성 및 복구 메커니즘
  - 포괄적인 테스트 커버리지 (sync_state_test.go)
- **Update 명령어 개선**: update.go의 주요 리팩토링 (+315줄)
  - 향상된 오류 처리 및 재시도 로직
  - 개선된 업데이트 검증 및 롤백
  - 더 나은 진행 상황 보고

### 변경됨 (Changed)

- **워크트리 설정**: 개선된 설정 관리 및 오류 처리
- **Rank 클라이언트**: 향상된 오류 처리 및 재시도 로직으로 리팩토링
- **템플릿 구조**: 더 나은 유지 관리를 위한 설정 템플릿 통합
- **프로젝트 초기화**: 새로운 설정 레이아웃에 맞게 업데이트

### 제거됨 (Removed)

- **레거시 설정 파일**:
  - .moai/config/sections/llm.yaml (Context7 MCP로 이동)
  - .moai/config/sections/pricing.yaml (Context7 MCP로 이동)
  - internal/template/templates/.moai/config/config.yaml.tmpl (통합됨)
  - internal/template/templates/.moai/config/sections/llm.yaml (Context7 MCP로 이동)
  - internal/template/templates/.moai/config/sections/pricing.yaml (Context7 MCP로 이동)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.6] - 2026-02-10

### Summary

Feature-rich release introducing comprehensive Agent Teams integration and Sequential Thinking MCP. Added 28 skill assignments across 15 agents for enhanced domain expertise, and introduced the moai-foundation-thinking skill for deep analysis workflows.

### Breaking Changes

None

### Added

- **Agent Teams Integration**: Comprehensive team mode support with TeamCreate, SendMessage, and TeamDelete APIs
- **Team Coordination Rules**: Parallel execution patterns, quality gates (TeammateIdle, TaskCompleted hooks)
- **Team Workflow References**: Prerequisites, fallback strategies, and team mode methodology documentation
- **Sequential Thinking Skill**: moai-foundation-thinking with critical evaluation, deep questioning, and diverge-converge modules
- **Agent Skills Enhancement**: 28 skill assignments across 15 agents (expert-backend, expert-frontend, manager-spec, team agents)
- **Hook System**: TeammateIdle and TaskCompleted hook configurations for team quality validation
- **FAQ Documentation**: Added FAQ section for statusline and external import warnings
- **Update Fix**: Clean up global hooks directory during `moai update` to prevent stale hooks

### Changed

- **CLAUDE.md**: Updated sections 5/8/14/15 with team patterns, coordination, and file ownership
- **Agent Authoring**: Enhanced with team invocation patterns and MCP references
- **Workflow Modes**: Added Team Mode Methodology section for DDD/TDD/Hybrid in team context

### Removed

- **Review Workflow**: Removed redundant `/moai review` workflow (functionality covered by `run` quality gates)
- **team-review Pattern**: Removed from workflow.yaml configuration

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.6] - 2026-02-10 (한국어)

### 요약

포괄적인 Agent Teams 통합과 Sequential Thinking MCP를 도입한 기능 중심 릴리스입니다. 15개 에이전트에 28개 스킬 할당을 추가하여 도메인 전문성을 강화했으며, 심층 분석 워크플로우를 위한 moai-foundation-thinking 스킬을 도입했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **Agent Teams 통합**: TeamCreate, SendMessage, TeamDelete API를 포함한 포괄적인 팀 모드 지원
- **팀 조정 규칙**: 병렬 실행 패턴, 품질 게이트 (TeammateIdle, TaskCompleted 훅)
- **팀 워크플로우 참조**: 전제 조건, 폴백 전략, 팀 모드 방법론 문서
- **Sequential Thinking 스킬**: 비판적 평가, 심층 질문, 발산-수렴 모듈을 포함한 moai-foundation-thinking
- **에이전트 스킬 강화**: 15개 에이전트에 28개 스킬 할당 (expert-backend, expert-frontend, manager-spec, team 에이전트)
- **Hook 시스템**: 팀 품질 검증을 위한 TeammateIdle 및 TaskCompleted 훅 구성
- **FAQ 문서**: 상태바 및 외부 임포트 경고에 대한 FAQ 섹션 추가
- **Update 수정**: `moai update` 중 전역 훅 디렉토리 정리로 오래된 훅 방지

### 변경됨 (Changed)

- **CLAUDE.md**: 팀 패턴, 조정, 파일 소유권으로 섹션 5/8/14/15 업데이트
- **에이전트 작성**: 팀 호출 패턴 및 MCP 참조로 강화
- **워크플로우 모드**: 팀 컨텍스트에서 DDD/TDD/Hybrid를 위한 Team Mode Methodology 섹션 추가

### 제거됨 (Removed)

- **Review 워크플로우**: 중복된 `/moai review` 워크플로우 제거 (기능은 `run` 품질 게이트에서 처리)
- **team-review 패턴**: workflow.yaml 구성에서 제거

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.5] - 2026-02-10

### Summary

Security enhancement release adding comprehensive binary format validation to the `moai update` command. Building on the v2.2.4 extraction fix, this release adds magic byte detection for all supported executable formats (Mach-O, ELF, PE) to prevent archive files from being mistakenly installed as executables. Includes extensive test coverage with 7 new validation test cases.

### Breaking Changes

None

### Added

- **Binary Format Validation**: Added `validateBinaryFormat()` function with magic byte detection for Mach-O (macOS), ELF (Linux), and PE (Windows) executable formats
- **Archive Rejection**: Automatic detection and rejection of gzip/zip archives with clear error messages and recovery instructions
- **Comprehensive Test Coverage**: Added 7 new test cases covering valid executables, archive rejection, and corrupted file handling

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.5] - 2026-02-10 (한국어)

### 요약

`moai update` 명령어에 포괄적인 바이너리 형식 검증 기능을 추가한 보안 개선 릴리스입니다. v2.2.4의 추출 수정을 기반으로, 지원되는 모든 실행 파일 형식(Mach-O, ELF, PE)에 대한 매직 바이트 감지를 추가하여 아카이브 파일이 실행 파일로 잘못 설치되는 것을 방지합니다. 7개의 새로운 검증 테스트 케이스로 광범위한 테스트 커버리지를 제공합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **바이너리 형식 검증**: Mach-O(macOS), ELF(Linux), PE(Windows) 실행 파일 형식에 대한 매직 바이트 감지 기능을 갖춘 `validateBinaryFormat()` 함수 추가
- **아카이브 거부**: gzip/zip 아카이브 자동 감지 및 명확한 오류 메시지와 복구 지침과 함께 거부
- **포괄적인 테스트 커버리지**: 유효한 실행 파일, 아카이브 거부, 손상된 파일 처리를 다루는 7개의 새로운 테스트 케이스 추가

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.4] - 2026-02-10

### Summary

Critical patch release fixing a major bug in the `moai update` command that prevented binary updates from working correctly. The updater was saving compressed archive files as executables instead of extracting the actual binary, causing "exec format error" when running moai after update. This release adds proper archive extraction logic for both tar.gz and zip formats.

### Breaking Changes

None

### Fixed

- **Critical: Binary Update Extraction**: Fixed `moai update` command that was saving tar.gz/zip archives as executables instead of extracting the moai binary, causing "exec format error" on all platforms after update
- **Windows Help Flag**: Added `/? goto show_help` support to install.bat for Windows CMD help convention
- **CI Workflow**: Resolved test-install.yml workflow file issue by properly splitting shell matrix configuration

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.4] - 2026-02-10 (한국어)

### 요약

`moai update` 명령어의 바이너리 업데이트 기능이 작동하지 않던 주요 버그를 수정한 긴급 패치 릴리스입니다. 업데이터가 압축 아카이브 파일을 실행 파일로 저장하여 업데이트 후 moai 실행 시 "exec format error"가 발생하던 문제를 해결했습니다. 이번 릴리스에서 tar.gz 및 zip 형식 모두에 대한 적절한 아카이브 추출 로직을 추가했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- **중요: 바이너리 업데이트 추출**: `moai update` 명령어가 tar.gz/zip 아카이브를 실행 파일로 저장하는 대신 moai 바이너리를 추출하지 않아 모든 플랫폼에서 업데이트 후 "exec format error"가 발생하던 문제 수정
- **Windows 도움말 플래그**: install.bat에 Windows CMD 도움말 규칙을 위한 `/? goto show_help` 지원 추가
- **CI 워크플로우**: shell 매트릭스 구성을 적절히 분리하여 test-install.yml 워크플로우 파일 이슈 해결

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.3] - 2026-02-10

### Summary

Patch release with beginner-friendly project workflow improvements, Windows support enhancements, and critical bug fixes. Introduces smart questions during project setup for better user guidance and fixes git command errors in non-git directories.

### Breaking Changes

None

### Added

- **Project Doc Auto-detection**: Automatically detects existing project documentation (product.md, architecture.md, tech.md) and creates smart questions to gather missing information
- **Beginner-friendly Smart Questions**: Interactive questions during `/moai project` to guide users through requirement analysis, with options to skip or provide natural language input
- **Two Large Skill Modules**: design-system-tokens.md (441 lines) and token-optimization.md (708 lines) previously excluded from git by overly broad `.gitignore` pattern

### Changed

- **Windows Support Scope**: Limited to WSL (recommended) and PowerShell 7.x+, with explicit requirement for Git for Windows installation
- **CI Dependencies**: Upgraded actions/upload-artifact to v6 and github/codeql-action to v4

### Fixed

- **SKILL.md Git Commands**: Fixed `!git branch --show-current` errors in non-git directories by adding `|| true` to pre-execution context commands
- **Legacy Hooks Cleanup**: `moai update` now properly removes old hook files that are no longer managed by MoAI-ADK
- **manager-spec Post-Edit Verification**: Added post-edit verification to ensure SPEC document was successfully written before proceeding
- **PowerShell Architecture Detection**: Added multi-layer fallback for architecture detection on Windows (environment variables, registry, WMI)
- **Cross-Platform Test Compatibility**: Fixed CI test failures on Windows for unknown shell type detection and platform-specific tests

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.3] - 2026-02-10 (한국어)

### 요약

초보자 친화적인 프로젝트 워크플로우 개선, Windows 지원 강화 및 중요한 버그 수정을 포함한 패치 릴리스입니다. 프로젝트 설정 중 스마트 질문을 도입하여 사용자 안내를 개선하고 git이 아닌 디렉토리에서 git 명령 오류를 수정했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **프로젝트 문서 자동 감지**: 기존 프로젝트 문서(product.md, architecture.md, tech.md)를 자동으로 감지하고 누락된 정보를 수집하기 위한 스마트 질문 생성
- **초보자 친화적 스마트 질문**: `/moai project` 실행 시 요구사항 분석을 안내하는 대화형 질문 제공, 건너뛰기 또는 자연어 입력 옵션 포함
- **2개의 대형 스킬 모듈**: 과도하게 넓은 `.gitignore` 패턴으로 git에서 제외되었던 design-system-tokens.md (441줄)와 token-optimization.md (708줄) 추가

### 변경됨 (Changed)

- **Windows 지원 범위**: WSL(권장) 및 PowerShell 7.x+로 제한하고, Git for Windows 설치 명시적 요구
- **CI 종속성**: actions/upload-artifact를 v6로, github/codeql-action을 v4로 업그레이드

### 수정됨 (Fixed)

- **SKILL.md Git 명령**: git이 아닌 디렉토리에서 `!git branch --show-current` 오류를 pre-execution context 명령에 `|| true` 추가로 수정
- **레거시 훅 정리**: `moai update`가 이제 MoAI-ADK에서 더 이상 관리하지 않는 오래된 훅 파일을 올바르게 제거
- **manager-spec 편집 후 검증**: SPEC 문서가 성공적으로 작성되었는지 확인하는 편집 후 검증 추가
- **PowerShell 아키텍처 감지**: Windows에서 아키텍처 감지를 위한 다층 fallback 추가 (환경 변수, 레지스트리, WMI)
- **크로스 플랫폼 테스트 호환성**: 알 수 없는 셸 유형 감지 및 플랫폼별 테스트에 대한 Windows CI 테스트 실패 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.2] - 2026-02-09

### Summary

Feature release adding persistent cross-session memory to agents, improving agent effectiveness through accumulated learnings. Expert agents now remember debugging patterns, API conventions, and component structures across sessions.

### Breaking Changes

None

### Added

- **Agent Memory System**: Added `memory` field to 10 agents for persistent cross-session learning
  - `expert-debug`: User-scoped memory for cross-project debugging patterns
  - `expert-backend`: Project-scoped memory for API/architecture patterns
  - `expert-frontend`: Project-scoped memory for component/style patterns
  - `manager-ddd`: Project-scoped memory for refactoring history
  - `manager-quality`: Project-scoped memory for quality gate results
  - `builder-skill`, `builder-agent`, `builder-plugin`: User-scoped memory for authoring patterns
- **Memory Scope Optimization**: Changed `team-researcher` and `team-designer` from project to user scope for cross-project pattern reuse

### Changed

- **Version Handling**: Enhanced version test coverage with comprehensive edge case handling

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.2] - 2026-02-09 (한국어)

### 요약

에이전트에 지속적인 세션 간 메모리를 추가하여 축적된 학습을 통해 에이전트 효율성을 개선하는 기능 릴리즈. 전문가 에이전트는 이제 디버깅 패턴, API 규칙, 컴포넌트 구조를 세션 간에 기억합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **에이전트 메모리 시스템**: 10개 에이전트에 세션 간 지속 학습을 위한 `memory` 필드 추가
  - `expert-debug`: 프로젝트 간 디버깅 패턴용 user 스코프 메모리
  - `expert-backend`: API/아키텍처 패턴용 project 스코프 메모리
  - `expert-frontend`: 컴포넌트/스타일 패턴용 project 스코프 메모리
  - `manager-ddd`: 리팩토링 이력용 project 스코프 메모리
  - `manager-quality`: 품질 게이트 결과용 project 스코프 메모리
  - `builder-skill`, `builder-agent`, `builder-plugin`: 작성 패턴용 user 스코프 메모리
- **메모리 스코프 최적화**: 프로젝트 간 패턴 재사용을 위해 `team-researcher`와 `team-designer`를 project에서 user 스코프로 변경

### 변경됨 (Changed)

- **버전 처리**: 포괄적인 엣지 케이스 처리로 버전 테스트 커버리지 향상

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.1] - 2026-02-09

### Summary

Critical patch release fixing checksum verification bug that prevented automatic binary updates. Users on v2.2.0 should update to v2.2.1 to restore automatic update functionality.

### Breaking Changes

None

### Fixed

- **Checksum Verification**: Fixed critical bug where checksums.txt URL was used as checksum value instead of downloading and parsing the file
- **Update Functionality**: Automatic binary updates now work correctly with proper SHA256 checksum verification
- **Graceful Degradation**: Update continues without checksum if checksums.txt download fails (with warning)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.1] - 2026-02-09 (한국어)

### 요약

자동 바이너리 업데이트를 방해하는 체크섬 검증 버그를 수정하는 치명적인 패치 릴리즈. v2.2.0 사용자는 자동 업데이트 기능을 복원하기 위해 v2.2.1로 업데이트해야 합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- **체크섬 검증**: 체크섬 파일을 다운로드하고 파싱하는 대신 URL을 체크섬 값으로 사용하는 치명적인 버그 수정
- **업데이트 기능**: 적절한 SHA256 체크섬 검증으로 자동 바이너리 업데이트가 정상 작동
- **우아한 저하**: checksums.txt 다운로드 실패 시 체크섬 없이 업데이트 계속 (경고 포함)

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.2.0] - 2026-02-09

### Summary

Major refactoring release consolidating skills from 64 to 52 for improved token efficiency and standardized architecture. Includes comprehensive agent skill injection optimization to ensure all skills are properly mapped to their target agents based on trigger configuration.

### Breaking Changes

None

### Added

- **Skill Consolidation**: Reduced skill count from 64 to 52 through merging related domain skills (moai-platform-vercel, moai-platform-railway, moai-platform-supabase, moai-platform-neon merged into moai-platform-deployment and moai-platform-database-cloud)
- **Design Tools Integration**: New moai-design-tools skill providing unified guidance for Figma MCP, Pencil MCP, and design-to-code workflows
- **Enhanced Skill Triggers**: All skills now have proper `triggers.agents` mapping for automatic skill loading

### Changed

- **Agent Skill Injection**: Optimized skill injection across 10 agents (expert-backend, expert-frontend, expert-devops, manager-spec, manager-ddd, manager-quality, builder-agent, builder-skill, builder-plugin, expert-chrome-extension)
- **Foundation Skills**: Added moai-foundation-philosopher to expert agents for strategic analysis capabilities
- **Platform Skills**: Added platform-specific skills (auth, deployment, database-cloud) to relevant agents
- **Workflow Skills**: Added moai-workflow-jit-docs and moai-workflow-worktree to appropriate agents
- **Installer Title**: Updated to "MoAI's Agentic Development Kit" for better branding

### Fixed

- **Config File Restoration**: Fixed issue where parent directories weren't created when restoring config files during `moai update`
- **Skill Name Standardization**: Standardized all skill name fields to match directory names for consistency
- **Test Isolation**: Added `MOAI_TEST_MODE` environment variable to prevent tests from modifying actual project settings files
- **Platform Support**: Enhanced transcript parsing to support macOS, Linux, and Windows platforms with platform-specific Claude configuration directories

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.2.0] - 2026-02-09 (한국어)

### 요약

스킬 통합(64→52개)을 통한 토큰 효율 개선과 표준화된 아키텍처를 위한 대규모 리팩토링 릴리즈. 모든 스킬이 트리거 설정에 따라 대상 에이전트에 제대로 매핑되도록 에이전트 스킬 주입을 최적화했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- **스킬 통합**: 관련 도메인 스킬을 병합하여 스킬 개수를 64개에서 52개로 감소 (moai-platform-vercel, moai-platform-railway, moai-platform-supabase, moai-platform-neon을 moai-platform-deployment와 moai-platform-database-cloud로 통합)
- **디자인 도구 통합**: Figma MCP, Pencil MCP, 디자인-투-코드 워크플로우를 위한 통합 가이드를 제공하는 새로운 moai-design-tools 스킬
- **향상된 스킬 트리거**: 자동 스킬 로딩을 위해 모든 스킬에 적절한 `triggers.agents` 매핑 추가

### 변경됨 (Changed)

- **에이전트 스킬 주입**: 10개 에이전트(expert-backend, expert-frontend, expert-devops, manager-spec, manager-ddd, manager-quality, builder-agent, builder-skill, builder-plugin, expert-chrome-extension)의 스킬 주입 최적화
- **파운데이션 스킬**: 전략적 분석 기능을 위해 expert 에이전트에 moai-foundation-philosopher 추가
- **플랫폼 스킬**: 관련 에이전트에 플랫폼별 스킬(auth, deployment, database-cloud) 추가
- **워크플로우 스킬**: 적절한 에이전트에 moai-workflow-jit-docs와 moai-workflow-worktree 추가
- **설치 관리자 제목**: 브랜딩 개선을 위해 "MoAI's Agentic Development Kit"로 업데이트

### 수정됨 (Fixed)

- **설정 파일 복원**: `moai update` 중 설정 파일 복원 시 상위 디렉토리가 생성되지 않던 문제 수정
- **스킬 이름 표준화**: 일관성을 위해 모든 스킬 이름 필드를 디렉토리 이름과 일치하도록 표준화
- **테스트 격리**: 테스트에서 실제 프로젝트 설정 파일이 수정되지 않도록 `MOAI_TEST_MODE` 환경 변수 추가
- **플랫폼 지원**: macOS, Linux, Windows 플랫폼을 지원하도록 트랜스크립트 파싱 개선 및 플랫폼별 Claude 설정 디렉토리 지원 추가

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.1.2] - 2026-02-09

### Summary

Hotfix release addressing UI/UX improvements and token optimization for Agent Teams. Resolves .tmpl file display in merge list, JSON logging noise during initialization, and reduces token consumption by 30-45K tokens per team execution through skill injection optimization.

### Breaking Changes

None

### Fixed

- **Template Display**: Fixed .tmpl files appearing in merge confirmation list during `moai init` and `moai update` — deployer now strips .tmpl suffix before returning file paths
- **JSON Logging**: Removed JSON-formatted log output during CLI commands by replacing `slog.Default()` with discard handler in `internal/cli/deps.go`
- **Token Optimization**: Removed `moai-foundation-core` from all 8 team agent skill injections, reducing redundant file loading by 30-45K tokens per team execution

### Changed

- **Agent Skills Injection**: Team agents now load only domain-specific skills instead of foundation skills, following single-responsibility principle
- **Logging Strategy**: CLI commands now use no-op logger to eliminate structured log noise in user-facing output

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

**For users on v2.1.0 experiencing "No Go binary available" error**:

The v2.1.1 hotfix resolved the binary download issue. If you're still on v2.1.0, use the official install script to upgrade:

```bash
# Reinstall to latest version (recommended)
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version
```

---

## [2.1.2] - 2026-02-09 (한국어)

### 요약

Agent Teams의 UI/UX 개선 및 토큰 최적화를 처리하는 핫픽스 릴리스입니다. 병합 목록의 .tmpl 파일 표시, 초기화 중 JSON 로깅 노이즈를 해결하고, skill injection 최적화를 통해 팀 실행당 30-45K 토큰 소비를 줄였습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- **템플릿 표시**: `moai init` 및 `moai update` 시 병합 확인 목록에 .tmpl 파일이 표시되는 문제 수정 — deployer가 파일 경로 반환 전에 .tmpl suffix를 제거하도록 수정
- **JSON 로깅**: `internal/cli/deps.go`에서 `slog.Default()`를 discard handler로 교체하여 CLI 명령어 실행 시 JSON 형식 로그 출력 제거
- **토큰 최적화**: 8개 team agent의 skill injection에서 `moai-foundation-core` 제거, 팀 실행당 중복 파일 로딩을 30-45K 토큰 감소

### 변경됨 (Changed)

- **Agent Skill Injection**: Team agent가 foundation skill 대신 도메인별 skill만 로드하도록 변경, 단일 책임 원칙 준수
- **로깅 전략**: CLI 명령어가 no-op logger를 사용하여 사용자 대면 출력에서 구조화된 로그 노이즈 제거

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

**v2.1.0에서 "No Go binary available" 오류가 발생하는 사용자**:

v2.1.1 핫픽스에서 바이너리 다운로드 문제가 해결되었습니다. 여전히 v2.1.0을 사용 중이라면 공식 설치 스크립트를 사용하여 업그레이드하세요:

```bash
# 최신 버전으로 재설치 (권장)
curl -sSL https://raw.githubusercontent.com/modu-ai/moai-adk/main/install.sh | bash
moai version
```

---

## [2.1.1] - 2026-02-09

### Summary

Critical hotfix resolving binary download failure in `moai update`. Version prefix mismatch between GoReleaser archive naming and update checker caused "No Go binary available" error for all platforms.

### Breaking Changes

None

### Fixed

- **Binary Download**: Fixed archive name mismatch in update checker - GoReleaser strips "v" prefix from version tags, but checker was using full tag name (e.g., "v2.1.0"), causing download to fail
- **Update Logic**: Added version prefix stripping logic to handle both "v" and "go-v" tag prefixes, ensuring correct archive URL construction

### Installation & Update

\`\`\`bash
# Update to the latest version
moai update

# Verify version
moai version
\`\`\`

**Note**: If `moai update` still fails with v2.1.0, manually download v2.1.1:

\`\`\`bash
# macOS arm64 (Apple Silicon)
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_darwin_arm64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/

# macOS amd64 (Intel)
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_darwin_amd64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/

# Linux amd64
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_linux_amd64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/
\`\`\`

---

## [2.1.1] - 2026-02-09 (한국어)

### 요약

`moai update`에서 바이너리 다운로드 실패를 해결하는 긴급 핫픽스입니다. GoReleaser 아카이브 네이밍과 업데이트 체커 간의 버전 접두사 불일치로 인해 모든 플랫폼에서 "No Go binary available" 오류가 발생했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- **바이너리 다운로드**: 업데이트 체커의 아카이브 이름 불일치 수정 - GoReleaser가 버전 태그에서 "v" 접두사를 제거하지만 체커는 전체 태그 이름(예: "v2.1.0")을 사용하여 다운로드 실패
- **업데이트 로직**: "v"와 "go-v" 태그 접두사를 모두 처리하는 버전 접두사 제거 로직 추가, 올바른 아카이브 URL 구성 보장

### 설치 및 업데이트 (Installation & Update)

\`\`\`bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
\`\`\`

**참고**: v2.1.0에서 `moai update`가 여전히 실패하면 v2.1.1을 수동으로 다운로드하세요:

\`\`\`bash
# macOS arm64 (Apple Silicon)
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_darwin_arm64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/

# macOS amd64 (Intel)
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_darwin_amd64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/

# Linux amd64
curl -L "https://github.com/modu-ai/moai-adk/releases/download/v2.1.1/moai-adk_2.1.1_linux_amd64.tar.gz" | tar -xz && sudo mv moai /usr/local/bin/
\`\`\`

---

## [2.1.0] - 2026-02-09

### Summary

Major update introducing SessionEnd hook support, Agent Teams enabled by default, and critical template system improvements. This release fixes cross-platform test failures and enhances the workflow execution system with intelligent mode selection.

### Breaking Changes

- `--auto` flag removed from workflow execution (auto-selection now default behavior)

### Added

- **SessionEnd Hook**: New `.claude/hooks/moai/handle-session-end.sh` wrapper for Claude Code session lifecycle management
- **Agent Hook System**: Dedicated agent-specific hook configuration in agent frontmatter with PreToolUse, PostToolUse, and SubagentStop support
- **Session Management**: Automatic session cleanup and state persistence through SessionEnd event handling

### Changed

- **Agent Teams Default**: Teams mode now enabled by default with complexity-based auto-selection (3+ domains, 10+ files, or score 7+)
- **Workflow Mode Selection**: Simplified execution mode logic — auto-selection analyzes task complexity to choose between team and sub-agent modes
- **Parallel Execution**: Enhanced efficiency with Agent Teams as primary execution mode for complex workflows

### Fixed

- **Cross-Platform Tests**: Resolved Windows path escaping, macOS Unicode NFD/NFC normalization, and non-git directory detection errors
- **Windows CI**: Fixed path separator issues, permission tests, and filesystem compatibility across Windows, macOS, and Linux
- **Template Filter**: `moai update` now correctly processes `.tmpl` files using rendered target paths instead of template paths
- **JSON Logging**: Merge confirmation now uses structured output, fixing JSON formatting issues during `moai update`
- **Config Cleanup**: Full configuration backup (including sections/) ensures complete v2.x-to-v2.x migration restore capability
- **Test Imports**: Removed unused `runtime` imports from shell and template test files

### Removed

- **Deprecated Flag**: `--auto` flag (auto-selection now default)
- **builder-command.md**: Removed 1,208-line agent definition in favor of skill-based command creation approach
- **Verbose Docs**: Cleaned up redundant documentation in hooks-system.md and workflow skills
- **Settings Bloat**: Removed unused settings from settings.json template

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.1.0] - 2026-02-09 (한국어)

### 요약

SessionEnd 훅 지원, Agent Teams 기본 활성화, 템플릿 시스템 개선을 포함한 주요 업데이트입니다. 크로스 플랫폼 테스트 실패를 수정하고 지능형 모드 선택으로 워크플로우 실행 시스템을 강화했습니다.

### 주요 변경 사항 (Breaking Changes)

- `--auto` 플래그 제거 (자동 선택이 이제 기본 동작)

### 추가됨 (Added)

- **SessionEnd Hook**: Claude Code 세션 생명주기 관리를 위한 `.claude/hooks/moai/handle-session-end.sh` 래퍼
- **Agent Hook System**: 에이전트별 훅 설정 지원 (PreToolUse, PostToolUse, SubagentStop)
- **세션 관리**: SessionEnd 이벤트를 통한 자동 세션 정리 및 상태 지속성

### 변경됨 (Changed)

- **Agent Teams 기본 활성화**: 복잡도 기반 자동 선택으로 Teams 모드가 기본값 (3개 이상 도메인, 10개 이상 파일, 또는 점수 7 이상)
- **워크플로우 모드 선택**: 실행 모드 로직 단순화 — 작업 복잡도를 분석하여 팀 모드와 서브 에이전트 모드 중 선택
- **병렬 실행 강화**: Agent Teams를 복잡한 워크플로우의 주요 실행 모드로 사용하여 효율성 향상

### 수정됨 (Fixed)

- **크로스 플랫폼 테스트**: Windows 경로 이스케이핑, macOS Unicode NFD/NFC 정규화, non-git 디렉토리 감지 오류 해결
- **Windows CI**: 경로 구분자 문제, 권한 테스트, Windows/macOS/Linux 파일시스템 호환성 수정
- **템플릿 필터**: `moai update`가 템플릿 경로 대신 렌더링된 대상 경로를 사용하여 `.tmpl` 파일을 올바르게 처리
- **JSON 로깅**: 병합 확인이 구조화된 출력을 사용하여 `moai update` 중 JSON 형식 문제 해결
- **설정 정리**: sections/를 포함한 전체 설정 백업으로 완전한 v2.x-to-v2.x 마이그레이션 복원 보장
- **테스트 import**: shell 및 template 테스트 파일에서 사용하지 않는 `runtime` import 제거

### 제거됨 (Removed)

- **더 이상 사용되지 않는 플래그**: `--auto` 플래그 (자동 선택이 기본값)
- **builder-command.md**: 1,208줄 에이전트 정의를 스킬 기반 명령 생성 방식으로 대체
- **장황한 문서**: hooks-system.md 및 워크플로우 스킬에서 중복 문서 정리
- **불필요한 설정**: settings.json 템플릿에서 사용되지 않는 설정 제거

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.5] - 2026-02-08

### Summary

Add git installation check to `moai init`, remove TUI experimental feature, and add v1-to-v2 migration cleanup utility.

### Breaking Changes

- Removed TUI (Terminal UI) experimental feature from `moai init` — `--tui` flag no longer available, `internal/cli/tui/` package deleted
- TUI will be redeveloped in future releases with improved architecture

### Added

- Git installation check in `moai init` with OS-specific installation guidance (macOS, Windows, Linux)
- `GitInstallHint()` function providing platform-specific git installation instructions
- `cleanMoaiManagedPaths()` utility for v1-to-v2 migration path cleanup
- Test coverage for git installation hints (`TestGitInstallHint`, `TestCheckGit_DetailWhenMissing`)

### Removed

- TUI (Terminal UI) experimental feature — 6 files deleted from `internal/cli/tui/` package (~1600 lines)
- `--tui` flag from `moai init` command
- `RunInitWizardTUI()` and `RunInitWithTUI()` functions
- Bubble Tea dependency from init command (CLI wizard remains intact)

### Changed

- `moai init` now shows non-fatal warning when git is not installed instead of silently continuing
- Git check runs after binary update step, before flag parsing

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.5] - 2026-02-08 (한국어)

### 요약

`moai init`에 git 설치 확인 기능을 추가하고, TUI 실험 기능을 제거하며, v1-to-v2 마이그레이션 정리 유틸리티를 추가했습니다.

### 주요 변경 사항 (Breaking Changes)

- TUI (Terminal UI) 실험 기능 제거 — `--tui` 플래그 더 이상 사용 불가, `internal/cli/tui/` 패키지 삭제
- TUI는 향후 개선된 아키텍처로 재개발될 예정

### 추가

- `moai init`에 OS별 설치 안내가 포함된 git 설치 확인 기능 추가 (macOS, Windows, Linux)
- 플랫폼별 git 설치 지침을 제공하는 `GitInstallHint()` 함수 추가
- v1-to-v2 마이그레이션 경로 정리를 위한 `cleanMoaiManagedPaths()` 유틸리티 추가
- git 설치 힌트 테스트 커버리지 추가 (`TestGitInstallHint`, `TestCheckGit_DetailWhenMissing`)

### 제거

- TUI (Terminal UI) 실험 기능 — `internal/cli/tui/` 패키지에서 6개 파일 삭제 (~1600줄)
- `moai init` 명령에서 `--tui` 플래그 제거
- `RunInitWizardTUI()`와 `RunInitWithTUI()` 함수 제거
- init 명령에서 Bubble Tea 의존성 제거 (CLI wizard는 유지)

### 변경

- git이 설치되지 않은 경우 `moai init`이 치명적 오류 대신 경고 메시지 표시
- git 확인은 바이너리 업데이트 단계 후, 플래그 파싱 전에 실행

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.4] - 2026-02-08

### Summary

Fix version persistence in `moai update` and `moai init`, and exclude hook files from merge confirmation UI. Official documentation link added to all README files.

### Breaking Changes

None

### Fixed

- Template version not persisted after `moai update` — `WithVersion()` was missing from `TemplateContext` creation in both `update.go` and `initializer.go`, causing `config.yaml` to render with empty version fields
- Status line showing stale version (`v1.14.0`) and perpetual update indicator because `moai.version` was empty in config
- `.claude/hooks/moai/*` files incorrectly appearing in merge confirmation UI during `moai update` — added `hooks` to `isMoaiManaged()` filter

### Added

- Official documentation link (https://adk.mo.ai.kr) to all README files (EN, KO, JA, ZH)
- Test cases for hooks path in `TestIsMoaiManaged` (3 new cases)

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.4] - 2026-02-08 (한국어)

### 요약

`moai update`와 `moai init`에서 템플릿 버전이 저장되지 않던 버그를 수정하고, 훅 파일이 병합 확인 UI에 노출되던 문제를 해결했습니다. 모든 README에 공식 문서 링크를 추가했습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 수정됨 (Fixed)

- `moai update` 후 템플릿 버전이 저장되지 않는 버그 — `update.go`와 `initializer.go`에서 `TemplateContext` 생성 시 `WithVersion()`이 누락되어 `config.yaml`의 버전 필드가 빈 문자열로 렌더링됨
- 상태 표시줄에 이전 버전(`v1.14.0`)이 표시되고 업데이트 표시가 계속 나타나는 문제 — config의 `moai.version`이 비어있었기 때문
- `moai update` 중 `.claude/hooks/moai/*` 파일이 병합 확인 UI에 잘못 표시되는 문제 — `isMoaiManaged()` 필터에 `hooks` 추가

### 추가됨 (Added)

- 모든 README(EN, KO, JA, ZH)에 공식 문서 링크(https://adk.mo.ai.kr) 추가
- `TestIsMoaiManaged`에 hooks 경로 테스트 케이스 3개 추가

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.3] - 2026-02-07

### Summary

Binary-first self-update and configuration improvements. The `moai update` command now updates the binary before syncing templates, ensuring the latest template engine processes files. Agent hook definitions and settings schema have been corrected.

### Breaking Changes

None

### Added

- Binary self-update step in `moai update` and `moai init` commands with re-exec pattern
- 3-layer loop prevention for binary update: env var guard, dev build detection, version comparison
- `--templates-only` flag for skipping binary update during re-exec
- `plansDirectory` setting in settings.json for Claude Code plan storage

### Changed

- `moai update` now performs binary update before template sync
- Agent hook definitions converted from object to array format for SubagentStop events
- Removed Homebrew tap from GoReleaser configuration

### Fixed

- Invalid schema fields removed from settings.json template
- Missing configuration fields added to settings.json template
- SubagentStop hooks in 8 agent definitions corrected to valid array format

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.3] - 2026-02-07 (한국어)

### 요약

바이너리 우선 자체 업데이트 및 설정 개선. `moai update` 명령어가 이제 템플릿 동기화 전에 바이너리를 먼저 업데이트하여 최신 템플릿 엔진이 파일을 처리하도록 보장합니다. 에이전트 훅 정의와 설정 스키마가 수정되었습니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- `moai update` 및 `moai init` 명령어에 re-exec 패턴을 활용한 바이너리 자체 업데이트 단계 추가
- 바이너리 업데이트를 위한 3중 루프 방지: 환경변수 가드, 개발 빌드 감지, 버전 비교
- re-exec 시 바이너리 업데이트 건너뛰기를 위한 `--templates-only` 플래그
- Claude Code 계획 문서 저장을 위한 settings.json에 `plansDirectory` 설정 추가

### 변경됨 (Changed)

- `moai update`가 이제 템플릿 동기화 전에 바이너리 업데이트를 수행
- SubagentStop 이벤트의 에이전트 훅 정의를 객체에서 배열 형식으로 변환
- GoReleaser 설정에서 Homebrew tap 제거

### 수정됨 (Fixed)

- settings.json 템플릿에서 잘못된 스키마 필드 제거
- settings.json 템플릿에 누락된 설정 필드 추가
- 8개 에이전트 정의의 SubagentStop 훅을 유효한 배열 형식으로 수정

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.2] - 2026-02-07

### Summary

Template system refactoring and cross-platform compatibility improvements. This patch release migrates settings.json generation from runtime-based to template-based approach, improves PATH handling, and fixes Windows CI test failures.

### Breaking Changes

None

### Added

- Template-based configuration files: settings.json.tmpl, .mcp.json.tmpl, handle-session-end.sh.tmpl
- SmartPATH and Platform fields in TemplateContext for better cross-platform support

### Changed

- Migrated settings.json generation from runtime JSON builder to template-based rendering
- Simplified SettingsGenerator by removing complex JSON construction logic
- Removed settings.json merge logic from update command (now handled by template deployment)
- Enhanced template rendering with SmartPATH and Platform context

### Fixed

- Resolved cross-platform test failures on Windows CI
- Restored .moai/project, specs, and config directories deleted in v2.0.0 cleanup
- Fixed PowerShell `$IsWindows` read-only variable conflict

### Technical Details

**Template System Improvements:**
- Centralized configuration in templates for single source of truth
- Better cross-platform PATH handling via SmartPATH
- Consistent template rendering across init and update commands
- Reduced maintenance overhead with template-based approach

**Test Coverage:**
- All 30 packages pass race detection tests
- Zero linting issues
- Enhanced test coverage for template rendering

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.2] - 2026-02-07 (한국어)

### 요약

템플릿 시스템 리팩토링 및 크로스 플랫폼 호환성 개선. 이 패치 릴리스는 settings.json 생성을 런타임 기반에서 템플릿 기반 접근 방식으로 마이그레이션하고, PATH 처리를 개선하며, Windows CI 테스트 실패를 수정합니다.

### 주요 변경 사항 (Breaking Changes)

없음

### 추가됨 (Added)

- 템플릿 기반 구성 파일: settings.json.tmpl, .mcp.json.tmpl, handle-session-end.sh.tmpl
- 더 나은 크로스 플랫폼 지원을 위한 TemplateContext의 SmartPATH 및 Platform 필드

### 변경됨 (Changed)

- settings.json 생성을 런타임 JSON 빌더에서 템플릿 기반 렌더링으로 마이그레이션
- 복잡한 JSON 구성 로직을 제거하여 SettingsGenerator 단순화
- update 명령에서 settings.json 병합 로직 제거 (이제 템플릿 배포로 처리)
- SmartPATH 및 Platform 컨텍스트로 템플릿 렌더링 강화

### 수정됨 (Fixed)

- Windows CI에서 크로스 플랫폼 테스트 실패 해결
- v2.0.0 정리 시 삭제된 .moai/project, specs, config 디렉토리 복원
- PowerShell `$IsWindows` 읽기 전용 변수 충돌 수정

### 기술 세부 사항

**템플릿 시스템 개선:**
- 단일 소스로서의 템플릿에 구성 중앙화
- SmartPATH를 통한 더 나은 크로스 플랫폼 PATH 처리
- init 및 update 명령에서 일관된 템플릿 렌더링
- 템플릿 기반 접근 방식으로 유지 관리 오버헤드 감소

**테스트 커버리지:**
- 30개 패키지 모두 race detection 테스트 통과
- linting 문제 0개
- 템플릿 렌더링에 대한 향상된 테스트 커버리지

### 설치 및 업데이트 (Installation & Update)

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.1] - 2026-02-07

### 요약

Windows 설치 스크립트 버그 수정 및 릴리즈 워크플로우 개선

### 주요 변경 사항

없음

### 수정됨

- Windows PowerShell 6+ 환경에서 `$IsWindows` 읽기 전용 변수 충돌 해결
- `moai update` 실행 시 불필요한 JSON 로그 출력 제거 (merge confirmation)

### 변경됨

- 릴리즈 노트 이중언어 형식을 영어 우선으로 변경 (이전: 한국어 우선)
- CI/CD 워크플로우에 OAuth 토큰 설정 추가

### 설치 및 업데이트

```bash
# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

---

## [2.0.1] - 2026-02-07 (English)

### Summary

Windows installer bugfix and release workflow improvements

### Breaking Changes

None

### Fixed

- Resolved PowerShell `$IsWindows` read-only variable conflict in Windows installer (PowerShell 6+)
- Removed unwanted JSON log output during `moai update` (merge confirmation)

### Changed

- Updated release notes bilingual format to English-first (previously Korean-first)
- Added OAuth token configuration to CI/CD workflows

### Installation & Update

```bash
# Update to the latest version
moai update

# Verify version
moai version
```

---

## [2.0.0] - 2026-02-06

### Summary

**Major Release: MoAI-ADK Go Edition**

This is the first official release of MoAI-ADK Go Edition, a complete rewrite of the Python-based MoAI-ADK in Go. This release delivers significantly improved performance, easier installation, and enhanced features while maintaining full compatibility with Claude Code workflows.

### Latest Updates (2026-02-06)

**Template Synchronization:**
- Synchronized 17 agent definition files with updated skill frontmatter
- Updated workflow skills (SKILL.md v2.0.0, moai.md) with team mode support
- Updated workflow-modes.md with Hybrid methodology as default
- Synchronized workflow.yaml and status_line.sh templates
- Updated CLAUDE.md to v12.0.0 with Agent Teams documentation

**Agent Hooks System:**
- Added agent-specific hooks for workflow enforcement
- Implemented `SubagentStop` event type for agent completion hooks
- Created `handle-agent-hook.sh` wrapper script for agent hooks
- Added factory pattern for agent-specific handlers in `internal/hook/agents/`
- Implemented hook actions for DDD workflow (ddd-pre-transformation, ddd-post-transformation, ddd-completion)
- Implemented hook actions for TDD workflow (tdd-pre-implementation, tdd-post-implementation, tdd-completion)
- Added validation/verification hooks for expert agents (backend, frontend, testing, debug, devops)
- Added completion hooks for manager agents (quality, spec, docs)
- Updated hooks-system.md documentation with agent hooks reference
- Synchronized agent hook configuration to all template locations

**Code Quality Improvements:**
- Fixed missing error checks in init_tui.go (added nolint comments for informational messages)
- Fixed missing error checks in init.go (added nolint comment for informational message)
- Simplified character validation logic in wizard_tui.go using De Morgan's law
- All 26 packages pass race detection tests
- Zero linting issues after fixes
- Fixed `.tmpl` file display in `moai update` (now shows rendered target paths)
- Fixed `permissions.allow` format (array instead of string per Claude Code IAM docs)

**Language Configuration:**
- Default conversation language set to Korean (ko) for improved user experience

**Additional Updates (Post v2.0.0 Tag):**
- **Documentation Restructuring**:
  - Made English the default README, moved Korean to README.ko.md (2e28f54f)
  - Maintained multilingual support (EN, JA, ZH, KO)
- **CI/CD Enhancements**:
  - Switched claude-code-action to GLM API Key (unofficial) (29d353ca)
  - Added open-source AI automation infrastructure (ffcaa6a2)
  - Improved CI/CD workflows with CodeQL, community automation
- **Project Organization**:
  - Untracked .moai local config, keeping only project/ and status_line.sh (8153bb19)
  - Cleaned up 38,895 lines of stale SPEC/project files
- **GitHub Flow Integration**:
  - Added /moai cpr command for issue-to-PR automation (081e5b7a)
  - Switched to GitHub Flow branch protection with feature/hotfix patterns (61f54378)
  - Made git delivery strategy-aware instead of GitHub Flow only (3fdec7aa)
- **Agent Teams Infrastructure** (a95e2a8d):
  - Added 8 team agents: team-researcher, team-analyst, team-architect, team-designer, team-backend-dev, team-frontend-dev, team-tester, team-quality
  - Created team workflow skills: team-plan, team-run, team-debug, team-review, team-sync
  - Implemented dual-mode execution (sub-agent vs Agent Teams)
  - Added complexity-based automatic mode selection
- **Settings Migration** (d01d16b8):
  - Migrated env, permissions, and teammateMode from global to project-level settings
  - Smart PATH capture instead of removing env.PATH (233f8907, 76500f84)
  - Added required type field to statusLine configuration (ad40b799)
- **Code Quality**:
  - Improved StatusLine version display format with config fallback (9a8183cc)
  - Fixed CI builds for Go 1.25 compatibility with golangci-lint (c72f4516, 542e146b, c58a61f7)
- **Community Infrastructure**:
  - Added CONTRIBUTING.md (KO/EN), SECURITY.md, LICENSE
  - GitHub issue/PR templates, dependabot, labeler, CodeQL

### Breaking Changes

- **Installation Method**: Changed from `uv tool install moai-adk` to single binary installation
- **Hook System**: Migrated from Python hooks to shell script wrappers
- **Configuration**: Updated configuration file structure and locations
- **Update Mechanism**: New automatic update system with GitHub releases integration

### Added

- **Go Edition Core**: Complete rewrite in Go for better performance and easier distribution
- **Multi-platform Binary Support**: Pre-built binaries for macOS (ARM64/Intel), Linux (ARM64/AMD64), Windows (AMD64)
- **Embedded Template System**: Templates now embedded using `go:embed` for faster startup
- **Web-based Installation UI**: Modern web interface for installation instructions
- **Korean Documentation**: Full Korean language documentation and migration guide
- **Go-specific Release Command**: `/moai:99-release` for automated release workflow
- **Transcript Parsing**: Support for Claude Code transcript analysis with MoAI Rank
- **LSP Quality Gates**: Integrated LSP diagnostics for quality validation
- **Security Scanner**: Hook-based security scanning for code changes
- **i18n Support**: Multi-language support in CLI commands
- **Agent Teams v2.0** (Experimental): Dual-mode execution engine with sub-agent and Agent Teams support
  - 5 team agents: researcher, backend-dev, frontend-dev, tester, quality
  - Team workflow skill with plan/run orchestration
  - `--team`, `--solo`, `--auto` execution mode flags
  - Complexity-based automatic mode selection
  - File ownership strategy for write conflict prevention
  - Workflow configuration (`workflow.yaml`) with team patterns
- **Hook Auto-Update**: Automatic update checking via session hooks
- **Update Cache**: Caching layer for update checks to reduce API calls
- **Agent Hooks System**: Agent-specific hooks for workflow enforcement
  - SubagentStop event type for agent lifecycle hooks
  - handle-agent-hook.sh wrapper script for consistent interface
  - Factory pattern for agent-specific handlers
  - DDD workflow hooks (pre/post-transformation, completion)
  - TDD workflow hooks (pre/post-implementation, completion)
  - Expert agent validation/verification hooks
  - Manager agent completion hooks

### Changed

- **Performance**: 10x faster startup time compared to Python version
- **Memory Usage**: Reduced memory footprint with Go runtime
- **Update System**: New update mechanism with GitHub releases integration
- **Template Deployment**: Automatic template deployment during initialization
- **Configuration Management**: Enhanced configuration with better validation
- **Development Methodology**: Hybrid (TDD+DDD) is now the default for new projects; DDD reserved for brownfield/legacy
- **CLI Update Command**: Refactored with extracted dependency management (`deps.go`)
- **StatusLine**: Improved version display and rendering with expanded test coverage
- **CLAUDE.md**: Updated to v12.0.0 with Agent Teams section (Section 15)
- **SKILL.md**: Updated to v2.0.0 with team mode support and execution mode selection

### Fixed

- **GitHub Issue #323**: Fixed PowerShell `irm | iex` installation failure
  - Wrapped install.ps1 script in `& { ... } @args` scriptblock for piping compatibility
  - Added ARM64 platform detection via ProcessArchitecture
  - Changed install location from `$env:USERPROFILE` to `$env:LOCALAPPDATA\Programs\moai`
  - Added SHA-256 checksum verification
- **GitHub Issue #324**: Fixed Linux/WSL2 installation 404 download error
  - Updated download URL to match goreleaser archive naming (`moai-adk_go-vX.Y.Z_OS_ARCH.tar.gz`)
  - Added tar.gz extraction step
  - Added SHA-256 checksum verification
  - Added WSL environment detection
- Windows CMD installation script improvements
  - Added ARM64 platform detection
  - Updated download URL to match goreleaser naming
  - Added extraction via PowerShell Expand-Archive
  - Fixed install location to `%LOCALAPPDATA%\Programs\moai`
- goreleaser configuration fixes
  - Fixed module path from `moai-adk-go` to `moai-adk` in ldflags
  - Fixed release target repository from `moai-adk-go` to `moai-adk`
- Windows hook execution improvements
  - Changed from `cmd.exe /c` to `bash` command (uses Git for Windows)
  - Ensures consistent hook execution across all platforms
- Cross-platform path construction
  - Replaced string concatenation with `filepath.Join()` in shell detection
  - Fixed path handling for PowerShell profile detection
- Update checker enhancements
  - Added `go-v` prefix support for version comparison
  - Updated archive naming to match goreleaser conventions
- StatusLine configuration
  - Changed from absolute path to relative path for better portability
  - Addresses GitHub Issue #7925 (StatusLine doesn't expand environment variables)
- Go bin path detection on Windows
  - Added fallback paths for Go installation directory detection
  - Checks `%PROGRAMFILES%\Go\bin` and `C:\Go\bin`
- Template synchronization issues in development builds
- Browser opening during automated tests
- Hook JSON output schema compliance
- API URL routing to correct repository

### Installation & Update

```bash
# Install MoAI-ADK Go Edition (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash

# Or download binary directly from GitHub Releases
# Visit: https://github.com/modu-ai/moai-adk/releases/tag/go-v2.0.0

# Update to the latest version
moai update

# Verify version
moai version
```

### Migration from Python Version

Users migrating from Python MoAI-ADK v1.x should:

1. Uninstall Python version: `uv tool uninstall moai-adk`
2. Install Go Edition using binary installation
3. Run `moai init` to update project templates

See [MIGRATION.ko.md](MIGRATION.ko.md) for detailed migration guide.

---

## [2.0.0] - 2026-02-06

### 요약

**메이저 릴리스: MoAI-ADK Go 에디션**

Python 기반 MoAI-ADK를 Go로 완전히 재작성한 첫 번째 공식 릴리스입니다. 성능이 크게 향상되고 설치가 간편해지며 기능이 향상되었습니다.

### 최신 업데이트 (2026-02-06)

**템플릿 동기화:**
- 업데이트된 스킬 프론트매터로 17개 에이전트 정의 파일 동기화
- 팀 모드 지원이 포함된 워크플로우 스킬 (SKILL.md v2.0.0, moai.md) 업데이트
- Hybrid 방법론을 기본값으로 사용하는 workflow-modes.md 업데이트
- workflow.yaml 및 status_line.sh 템플릿 동기화
- Agent Teams 문서가 포함된 CLAUDE.md v12.0.0 업데이트

**코드 품질 개선:**
- init_tui.go에서 누락된 오류 검사 수정 (정보 메시지에 nolint 주석 추가)
- init.go에서 누락된 오류 검사 수정 (정보 메시지에 nolint 주석 추가)
- 드 모르간 법칙을 사용한 wizard_tui.go의 문자 검증 로직 단순화
- 26개 패키지 모두 race detection 테스트 통과
- 수정 후 linting 문제 0개
- `moai update`에서 `.tmpl` 파일 표시 수정 (이제 렌더링된 대상 경로 표시)
- `permissions.allow` 형식 수정 (Claude Code IAM 문서에 따라 문자열 대신 배열 사용)

**언어 설정:**
- 개선된 사용자 경험을 위해 기본 대화 언어를 한국어(ko)로 설정

**추가 업데이트 (v2.0.0 태그 이후):**
- **문서 재구성**:
  - 영문 README를 기본으로 설정, 한국어를 README.ko.md로 이동 (2e28f54f)
  - 다국어 지원 유지 (EN, JA, ZH, KO)
- **CI/CD 개선**:
  - claude-code-action을 GLM API Key로 전환 (비공식) (29d353ca)
  - 오픈소스 AI 자동화 인프라 추가 (ffcaa6a2)
  - CodeQL, 커뮤니티 자동화를 포함한 CI/CD 워크플로우 개선
- **프로젝트 정리**:
  - .moai 로컬 설정 untrack, project/ 및 status_line.sh만 유지 (8153bb19)
  - 오래된 SPEC/project 파일 38,895줄 정리
- **GitHub Flow 통합**:
  - issue-to-PR 자동화를 위한 /moai cpr 명령어 추가 (081e5b7a)
  - feature/hotfix 패턴을 사용한 GitHub Flow 브랜치 보호 전환 (61f54378)
  - GitHub Flow만이 아닌 전략 인식 git 전달 방식으로 변경 (3fdec7aa)
- **에이전트 팀 인프라** (a95e2a8d):
  - 8개 팀 에이전트 추가: team-researcher, team-analyst, team-architect, team-designer, team-backend-dev, team-frontend-dev, team-tester, team-quality
  - 팀 워크플로우 스킬 생성: team-plan, team-run, team-debug, team-review, team-sync
  - 이중 모드 실행 구현 (sub-agent vs Agent Teams)
  - 복잡도 기반 자동 모드 선택 추가
- **설정 마이그레이션** (d01d16b8):
  - env, permissions, teammateMode를 global에서 project-level로 마이그레이션
  - env.PATH 제거 대신 Smart PATH 캡처 (233f8907, 76500f84)
  - statusLine 구성에 필수 type 필드 추가 (ad40b799)
- **코드 품질**:
  - config fallback을 사용한 StatusLine 버전 표시 형식 개선 (9a8183cc)
  - golangci-lint와 Go 1.25 호환성을 위한 CI 빌드 수정 (c72f4516, 542e146b, c58a61f7)
- **커뮤니티 인프라**:
  - CONTRIBUTING.md (KO/EN), SECURITY.md, LICENSE 추가
  - GitHub 이슈/PR 템플릿, dependabot, labeler, CodeQL

**에이전트 훅 시스템:**
- 워크플로우 강제를 위한 에이전트별 훅 추가
- 에이전트 완료 훅을 위한 `SubagentStop` 이벤트 타입 구현
- 에이전트 훅을 위한 `handle-agent-hook.sh` 래퍼 스크립트 생성
- `internal/hook/agents/`의 에이전트별 핸들러를 위한 팩토리 패턴 추가
- DDD 워크플로우 훅 구현 (ddd-pre-transformation, ddd-post-transformation, ddd-completion)
- TDD 워크플로우 훅 구현 (tdd-pre-implementation, tdd-post-implementation, tdd-completion)
- 전문가 에이전트를 위한 검증/확인 훅 추가 (backend, frontend, testing, debug, devops)
- 관리자 에이전트를 위한 완료 훅 추가 (quality, spec, docs)
- 에이전트 훅 참조가 포함된 hooks-system.md 문서 업데이트
- 모든 템플릿 위치에 에이전트 훅 구성 동기화

### Breaking Changes

- **설치 방법**: `uv tool install moai-adk`에서 단일 바이너리 설치로 변경
- **훅 시스템**: Python 훅에서 셸 스크립트 래퍼로 마이그레이션
- **설정**: 설정 파일 구조 및 위치 업데이트
- **업데이트 메커니즘**: GitHub 릴리스 통합 새 업데이트 시스템

### 추가됨

- **Go 에디션 코어**: 더 나은 성능과 배포를 위한 Go로 완전 재작성
- **멀티 플랫폼 바이너리 지원**: macOS (ARM64/Intel), Linux (ARM64/AMD64), Windows (AMD64)용 미리 빌드된 바이너리
- **임베디드 템플릿 시스템**: `go:embed`를 사용한 더 빠른 시작을 위한 템플릿 임베딩
- **웹 기반 설치 UI**: 설치 안내를 위한 현대적 웹 인터페이스
- **한국어 문서**: 완전한 한국어 문서 및 마이그레이션 가이드
- **Go 전용 릴리스 명령**: 자동화된 릴리스 워크플로우를 위한 `/moai:99-release`
- **트랜스크립트 파싱**: MoAI Rank를 위한 Claude Code 트랜스크립트 분석 지원
- **LSP 품질 게이트**: 품질 검증을 위한 통합 LSP 진단
- **보안 스캐너**: 코드 변경을 위한 훅 기반 보안 스캐닝
- **i18n 지원**: CLI 명령어의 다국어 지원
- **에이전트 훅 시스템**: 워크플로우 강제를 위한 에이전트별 훅
  - 에이전트 수명주기 훅을 위한 SubagentStop 이벤트 타입
  - 일관된 인터페이스를 위한 handle-agent-hook.sh 래퍼 스크립트
  - 에이전트별 핸들러를 위한 팩토리 패턴
  - DDD 워크플로우 훅 (pre/post-transformation, completion)
  - TDD 워크플로우 훅 (pre/post-implementation, completion)
  - 전문가 에이전트 검증/확인 훅
  - 관리자 에이전트 완료 훅

### 변경됨

- **성능**: Python 버전 대비 10배 더 빠른 시작 시간
- **메모리 사용량**: Go 런타임으로 감소된 메모리 사용량
- **업데이트 시스템**: GitHub 릴리스 통합 새 업데이트 메커니즘
- **템플릿 배포**: 초기화 중 자동 템플릿 배포
- **설정 관리**: 향상된 검증을 통한 개선된 설정

### 수정됨

- **GitHub Issue #323**: PowerShell `irm | iex` 설치 실패 수정
  - 파이핑 호환성을 위해 install.ps1 스크립트를 `& { ... } @args` 스크립트블록으로 래핑
  - ProcessArchitecture를 통한 ARM64 플랫폼 감지 추가
  - 설치 위치를 `$env:USERPROFILE`에서 `$env:LOCALAPPDATA\Programs\moai`로 변경
  - SHA-256 체크섬 검증 추가
- **GitHub Issue #324**: Linux/WSL2 설치 404 다운로드 오류 수정
  - goreleaser 아카이브 명명 규칙에 맞게 다운로드 URL 업데이트 (`moai-adk_go-vX.Y.Z_OS_ARCH.tar.gz`)
  - tar.gz 압축 해제 단계 추가
  - SHA-256 체크섬 검증 추가
  - WSL 환경 감지 추가
- Windows CMD 설치 스크립트 개선
  - ARM64 플랫폼 감지 추가
  - goreleaser 명명 규칙에 맞게 다운로드 URL 업데이트
  - PowerShell Expand-Archive를 통한 압축 해제 추가
  - 설치 위치를 `%LOCALAPPDATA%\Programs\moai`로 수정
- goreleaser 설정 수정
  - ldflags의 모듈 경로를 `moai-adk-go`에서 `moai-adk`로 수정
  - 릴리스 대상 저장소를 `moai-adk-go`에서 `moai-adk`로 수정
- Windows 훅 실행 개선
  - `cmd.exe /c`에서 `bash` 명령으로 변경 (Git for Windows 사용)
  - 모든 플랫폼에서 일관된 훅 실행 보장
- 크로스 플랫폼 경로 구성
  - 셸 감지에서 문자열 연결을 `filepath.Join()`으로 교체
  - PowerShell 프로필 감지를 위한 경로 처리 수정
- 업데이트 검사기 개선
  - 버전 비교를 위한 `go-v` 접두사 지원 추가
  - goreleaser 규칙에 맞게 아카이브 명명 업데이트
- StatusLine 설정
  - 이식성 향상을 위해 절대 경로에서 상대 경로로 변경
  - GitHub Issue #7925 해결 (StatusLine이 환경 변수를 확장하지 않음)
- Windows에서 Go bin 경로 감지
  - Go 설치 디렉터리 감지를 위한 대체 경로 추가
  - `%PROGRAMFILES%\Go\bin` 및 `C:\Go\bin` 확인
- 개발 빌드에서의 템플릿 동기화 문제
- 자동화된 테스트 중 브라우저 열림 문제
- 훅 JSON 출력 스키마 준수
- 올바른 저장소로의 API URL 라우팅

### 설치 및 업데이트

```bash
# MoAI-ADK Go 에디션 설치 (macOS/Linux)
curl -fsSL https://raw.githubusercontent.com/modu-ai/moai-adk/moai-go-v2/install.sh | bash

# 또는 GitHub 릴리스에서 바이너리 직접 다운로드
# 방문: https://github.com/modu-ai/moai-adk/releases/tag/go-v2.0.0

# 최신 버전으로 업데이트
moai update

# 버전 확인
moai version
```

### Python 버전에서 마이그레이션

Python MoAI-ADK v1.x에서 마이그레이션하는 사용자는:

1. Python 버전 제거: `uv tool uninstall moai-adk`
2. 바이너리 설치로 Go 에디션 설치
3. `moai init` 실행으로 프로젝트 템플릿 업데이트

자세한 마이그레이션 가이드는 [MIGRATION.ko.md](MIGRATION.ko.md)를 참조하세요.

---

## Release History

For previous releases, see [GitHub Releases](https://github.com/modu-ai/moai-adk/releases).

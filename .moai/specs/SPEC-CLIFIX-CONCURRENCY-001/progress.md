# SPEC-CLIFIX-CONCURRENCY-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §3 clusters 2/5 + §5 P2. Status: draft. Pending plan-audit. Sequenced after SPEC-CLIFIX-CONTRACT-001.
- 2026-07-11: iter-1 plan-audit FAIL (aggregate 0.77 < 0.80 Tier M threshold). Re-grounded against post-CRITICAL-001/CONTRACT-001 HEAD — all anchors re-verified by grep. 7 fixes applied (D1-D7). REQ/AC reduced 7/7→6/6: dropped REQ-001-005 (Upsert tier-transition already fixed at filestore.go:96 + read-side cascade/dedupe covers it). Pending iter-2 audit.
- 2026-07-11: D1 micro-fix (v0.2.1) — iter-2 plan-audit PASS 0.89 with single D1 debt (atomic-writer inventory under-count). Corrected inventory 4→6 sites (added atomicWrite preference/filestore.go:473 + glm.go:689-704 inline). Option (b) for atomicWrite (package-internal exception). AC-004 grep widened to include `func atomicWrite`. No other REQ/AC touched.

## §E.2 Run-phase Evidence

### M1 — Re-route writers + removeGLMEnv key-set (REQ-CONC-001-001 + REQ-CONC-001-002)

**AC PASS/FAIL matrix (M1 scope = AC-001 + AC-002):**

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CONC-001-001 | PASS | `go test -race ./internal/cli/ -run 'SettingsLocalConcurrent' -count=1 -v` | `PASS — TestSettingsLocalConcurrentWrites (0.12s)`; grep: 0 `writeSettingsMap` callers outside settings.go (all 5 re-routed) |
| AC-CONC-001-002 | PASS | `go test ./internal/cli/ -run 'RemoveGLMEnvComplete' -count=1 -v` && `grep -n 'EnvClaudeCodeAutoCompactWindow' internal/cli/launcher.go` | `PASS — TestRemoveGLMEnvComplete`; `launcher.go:292: delete(env, config.EnvClaudeCodeAutoCompactWindow)` |

**RED evidence (SPEC §D.5 — tests must fail on pre-fix commit):**
- `TestSettingsLocalConcurrentWrites` FAILED pre-M1: `parse settings.local.json: invalid character 'P' after top-level value` — concurrent `writeSettingsMap` (plain os.WriteFile) truncated the file mid-read by another goroutine.
- `TestRemoveGLMEnvComplete` FAILED pre-M1: `CLAUDE_CODE_AUTO_COMPACT_WINDOW was NOT removed by removeGLMEnv` — key missing from the delete set at launcher.go:268-281.

**Changes applied (GREEN):**
1. `internal/cli/glm.go`: 3 callers re-routed through `mutateSettingsLocal` — `ensureSettingsLocalJSON`, `injectGLMEnvForTeam`, `injectGLMEnv`.
2. `internal/cli/launcher.go`: 2 callers re-routed through `mutateSettingsLocal` — `removeGLMEnv`, `syncPermissionModeToSettingsLocal`; `removeGLMEnv` delete-set gains `config.EnvClaudeCodeAutoCompactWindow`.
3. `internal/cli/settings.go`: deleted dead `writeSettingsMap` (golangci-lint `unused` after all 5 callers re-routed).
4. `internal/cli/clifix_concurrency_repro_test.go`: new RED→GREEN reproduction tests.

**Cross-platform build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
**Lint:** `golangci-lint run --timeout=2m` → 0 issues (pre-flight baseline was also 0; the transient `writeSettingsMap unused` finding was resolved by deleting the dead function).
**Coverage:** `go test -cover ./internal/cli/` → 72.7% of statements (package-level).
**Race suite:** `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` → all 10 packages PASS.

### M2 — `~/.claude.json` RMW guard (REQ-CONC-001-003 / AC-CONC-001-003)

**AC PASS/FAIL matrix (M2 scope = AC-003):**

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CONC-001-003 | PASS | `go test -race ./internal/cli/ -run 'ClaudeJSONGuard' -count=1 -v` | `PASS` — 3 tests green: `TestClaudeJSONGuard_ConcurrentForceNoLostUpdate` (6.9s), `TestClaudeJSONGuard_ConcurrentIdempotentNoLostUpdate` (7.0s), `TestClaudeJSONGuard_ExternalWriteDetectedByCompareRetry` (0.02s) |

**RED evidence (SPEC §D.5 — tests must fail on pre-fix commit):**
- `TestClaudeJSONGuard_ConcurrentForceNoLostUpdate` FAILED pre-M2: `iter 0: MCP server "zai-mcp-server" lost (concurrent RMW lost update)` — three goroutines (vision/websearch/webreader) read the same empty state and the later `os.Rename` overwrote the earlier writer's entries.
- `TestClaudeJSONGuard_ConcurrentIdempotentNoLostUpdate` FAILED pre-M2: identical root cause via the idempotent path.
- Confirmed at HEAD `37b535820` (before the M2 edit): both tests fail on iter 0.

**Changes applied (GREEN):**
1. `internal/cli/glm_tools.go` — new `mutateClaudeJSONAtomic(configPath, apply)` seam: a guarded RMW that combines (a) advisory flock on a sibling `.lock` file (reuses the existing `lockFile`/`unlockFile` family — same convention as `mutateSettingsLocal`, no second lock convention) to serialize cooperating writers, and (b) a content compare-and-retry inside the lock to detect non-cooperating writers (e.g., a live Claude Code process that does not respect the `.lock` file). Marshal stays OUTSIDE the lock (prep phase); inside the lock the guard only re-reads, compares, conditionally re-applies, and publishes via temp+rename. Fail-open with stderr warning on unsupported FS or exhausted retries (claudeJSONGuardMaxRetries = 3). Bounded critical section so a large `~/.claude.json` does not block a live CC process.
2. `internal/cli/glm_tools.go` — extracted `writeClaudeJSONBytes(configPath, data)` (bytes-level temp+rename) from `writeClaudeJSONAtomic`; the guard writes pre-marshaled bytes inside the lock without re-running the marshal. `writeClaudeJSONAtomic` delegates here (behavior-preserving — its signature is unchanged; still used by `disableMCPServerForTool` which is NOT in M2 scope).
3. `internal/cli/glm_tools.go` — `runEnableMCPServerForTool` and `enableMCPServerIdempotentForTool` re-routed through `mutateClaudeJSONAtomic`. The apply closures are re-callable: token-mismatch check + idempotency check + mcpServers mutation are re-evaluated against the fresh in-lock state on each retry. The idempotent `skipped` flag reflects the LAST apply invocation so it stays correct even when a concurrent writer achieves the desired state mid-guard.
4. `internal/cli/glm_tools.go` — `readClaudeJSONRaw` + `parseClaudeJSON` helpers (raw bytes for the compare; parsed root for the apply). `claudeJSONGuardPreLockHook` test injection point (nil in production).
5. `internal/cli/clifix_concurrency_m2_test.go` — 3 RED→GREEN tests (2 concurrent stress + 1 deterministic compare-retry).

**Design choice (M3-forward note):** the guard is implemented as a `mutateClaudeJSONAtomic(path, apply)` seam mirroring `mutateSettingsLocal`, NOT by mutating `writeClaudeJSONAtomic`'s signature. M3's atomic-writer consolidation can build on the extracted `writeClaudeJSONBytes`. `disableMCPServerForTool` is NOT routed through the guard in M2 (SPEC REQ-CONC-001-003 names only the two enable sites); it remains a known gap that could use the same seam in a follow-up.

**Cross-platform build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (lockFile/unlockFile already build-tagged unix/windows by CRITICAL-001).
**Lint:** `golangci-lint run --timeout=3m` → 0 issues (baseline was also 0).
**Coverage:** `go test -cover ./internal/cli/` → 72.8% of statements (package-level; consistent with the pre-M2 baseline — the new guard code is exercised by 3 new tests + 50+ existing glm_tools tests that now route through the guard).
**Vet:** `go vet ./internal/cli/` exit 0; `gofmt` clean.
**Race suite (full, on worktree at main HEAD):** `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` → all 10 packages PASS. NOTE: an initial full-suite run surfaced a data race on the package-global `claudeJSONGuardPreLockHook` (the deterministic test wrote it while concurrent tests read it — both were `t.Parallel()`). Fixed by making `TestClaudeJSONGuard_ExternalWriteDetectedByCompareRetry` serial (NOT `t.Parallel()`) so it completes before any parallel test reads the hook. The cascade of other test failures in that initial run was collateral (race-detector tainting), not separate races — confirmed by the clean full-suite re-run.

### M3 — Atomic-writer consolidation (REQ-CONC-001-004 / AC-CONC-001-004)

**AC PASS/FAIL matrix (M3 scope = AC-004):**

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CONC-001-004 | PASS | `grep -rn 'func writeFileAtomic\|func atomicWriteJSON\|func writeClaudeJSONAtomic\|func atomicWrite' internal/cli --include='*.go' \| grep -v _test.go` | Exactly 2 matches: `internal/cli/settings.go:119:func writeFileAtomic` (consolidated) + `internal/cli/preference/filestore.go:473:func atomicWrite` (option (b) exception) |

**Union-of-guarantees characterization contract (DDD PRESERVE):**

| Former site | fsync? | Perm | MkdirAll? | Chmod? | Rename? |
|---|---|---|---|---|---|
| `writeFileAtomic` (settings.go) | NO | param (0600) | NO → gained | YES | YES |
| `atomicWriteJSON` (update_noise.go) | NO | CreateTemp default (0600) | NO → gained | NO → gained via perm param | YES |
| `writeClaudeJSONAtomic` (glm_tools.go) | NO | CreateTemp default (0600) | YES | NO → gained via perm param | YES |
| `writeClaudeJSONBytes` (glm_tools.go, M2 seam) | NO | CreateTemp default (0600) | YES | NO → gained via perm param | YES |
| inline harness_mute.go | NO | 0o644 (os.WriteFile) | YES | NO → gained via perm param | YES |
| inline glm.go saveLLMSection | NO | CreateTemp default (0600) | NO → gained | NO → gained via perm param | YES |

Union contract: perm param + `tmp.Chmod(perm)` (supports both 0600 credential and 0644 config), `os.MkdirAll(dir, 0o755)` (safe superset), `os.Rename` atomicity, NO fsync (none of the former callers relied on it). The consolidated `writeFileAtomic` takes the UNION of all guarantees.

**Characterization evidence (RED-anchored baseline → GREEN post-consolidation):**

8 characterization tests in `internal/cli/clifix_concurrency_m3_test.go`:
- Pre-consolidation (2e4513f90): 7 PASS + 1 FAIL (`TestCharacterize_WriteFileAtomic_CreatesParentDir` — MkdirAll absent pre-consolidation, confirming RED for the one superset behavior).
- Post-consolidation: all 8 PASS (GREEN = behavior preserved + superset added).

**Changes applied (GREEN):**
1. `internal/cli/settings.go` — `writeFileAtomic` enhanced with `os.MkdirAll(dir, 0o755)` (safe superset — matches writeClaudeJSONBytes + preference/atomicWrite). Added `@MX:ANCHOR` tag (consolidated helper, high fan_in).
2. `internal/cli/glm_tools.go` — `writeClaudeJSONAtomic` DELETED (sole caller `disableMCPServerForTool` inlines the marshal + calls `writeClaudeJSONBytes`); `writeClaudeJSONBytes` becomes a thin delegating wrapper to `writeFileAtomic(configPath, data, 0o600)` (M2 seam PRESERVED per task constraint).
3. `internal/cli/update_noise.go` — `atomicWriteJSON` DELETED (sole caller `saveMergeHistoryLedger` inlines the JSON encode via `bytes.Buffer` + `json.Encoder` to preserve trailing-newline convention + calls `writeFileAtomic(path, bytes, 0o600)`). Removed `defs` import (no longer needed).
4. `internal/cli/harness_mute.go` — inline tmp+rename block replaced with `writeFileAtomic(path, out, 0o644)` (perm 0644 preserved for non-credential config).
5. `internal/cli/glm.go` — `saveLLMSection` inline tmp+rename block replaced with `writeFileAtomic(path, data, 0o600)`.
6. `internal/cli/glm_tools_test.go` — `TestWriteClaudeJSONAtomic_BadDir` renamed to `TestWriteClaudeJSONBytes_BadDir` (function deleted; test now exercises the consolidated path via writeClaudeJSONBytes).
7. `internal/cli/target_coverage_test.go` — `TestSaveLLMSection_NonexistentDirFails` updated: error-message assertion loosened (failure point shifted from CreateTemp to MkdirAll; behavior contract — "fails for unwritable path" — preserved).

**Cross-platform build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (consolidated helper uses stdlib os.CreateTemp/os.Rename, no new syscall).
**Lint:** `golangci-lint run --timeout=3m` → 0 issues (baseline was also 0).
**Coverage:** `go test -cover ./internal/cli/` → 72.9% of statements (M2 was 72.8%; slight increase from new characterization tests + MkdirAll coverage).
**Vet:** `go vet ./internal/cli/...` exit 0; `gofmt` clean.
**Race suite (full):** `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` → all 10 packages PASS.

### M4 — Preference store crash-consistency + TOCTOU hardening (REQ-CONC-001-005 + REQ-CONC-001-006 / AC-005 + AC-006)

**AC PASS/FAIL matrix (M4 scope = AC-005 + AC-006):**

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CONC-001-005 | PASS | `go test -race ./internal/cli/preference/ -run 'DecayCrash\|ScanDueRace\|ToggleRace' -count=1 -v` | PASS — 3 tests green: `TestDecayCrash_ReconcilesDuplicateAfterInterruptedScan` (0.00s), `TestScanDueRace_ConcurrentMarkScannedStaysValid` (0.50s), `TestToggleRace_ConcurrentFlipsPreserveParity` (0.09s, even+odd subtests) |
| AC-CONC-001-006 | PASS | `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` | PASS — all 10 packages green (internal/cli 29.3s, preference 2.5s, + 8 others) |

**RED evidence (SPEC §D.5 — tests must fail on pre-fix commit c31db9e2b):**
- `TestDecayCrash_ReconcilesDuplicateAfterInterruptedScan` FAILED pre-M4 (runtime): `stale recall duplicate survived scan: entry test-domain/crash-key still in recall (post=[...]); archival also holds it → on-disk duplicate persists`. Pre-M4 DecayScan has no reconcile step, so a crash-induced recall+archival duplicate survives the scan.
- `TestToggleRace_ConcurrentFlipsPreserveParity` FAILED pre-M4 (compile): `undefined: TogglePersonalization`. The locked flip seam did not exist pre-M4; the read-then-flip TOCTOU was unhardened.
- `TestScanDueRace_ConcurrentMarkScannedStaysValid` PASSED pre-M4 (os.WriteFile small-write atomicity held at this concurrency level) — the atomic-write fix is still the correct hardening (closes the truncate-then-write window a concurrent reader could observe), but it is not a deterministic RED.

**Changes applied (GREEN):**

1. `internal/cli/preference/decay.go` — **DecayScan crash reconciliation (option b reconcile-at-scan-start).** New `reconcileRecallArchival(recall)` helper called at the top of DecayScan (after loadRecall, before the processing loop): drops any recall entry whose `(domain, decisionKey)` already exists in archival. A prior scan that crashed between `writeArchivalEntry` (decay.go:169) and `writeRecall` (decay.go:185) leaves an entry in BOTH tiers; the normal decay path only removes EXPIRED entries, so a non-expired stale recall copy would survive indefinitely. The reconcile step makes the store self-heal on every scan. Design choice (option b) over option (a transactional): a cross-file WAL/2PC is heavy for a preference store; reconcile-at-scan is idempotent and runs at the natural daily-maintenance boundary. Read paths (Query `seen`-map at filestore.go:147, Get cascade) already dedupe in-memory between scans.
2. `internal/cli/preference/decay.go` — **MarkScanned atomic write.** `MarkScanned` switched from `os.WriteFile` (O_TRUNC-then-write — a concurrent ScanDue reader could observe a 0-byte file mid-write) to the existing package-level `atomicWrite` helper (filestore.go:473, temp-in-same-dir + os.Rename). The stamp is now never observed in a half-written state. (Perm side-effect: 0644 → 0600 via CreateTemp; tightening, acceptable for a non-credential state file.)
3. `internal/cli/preference/toggle.go` — **Toggle TOCTOU hardening via O_EXCL sidecar lock.** New `TogglePersonalization(projectRoot, now) (bool, error)` seam: acquires an exclusive O_EXCL sidecar lock (`<root>/.moai/state/session-preference-disabled.toggle-lock`), reads `IsPersonalizationDisabled`, flips (Enable/Disable), releases. Lock primitive choice (SPEC option (a) "stdlib flock-equivalent directly in preference"): `os.OpenFile` with `O_CREATE|O_EXCL` is portable stdlib — no `syscall.Flock` (which would need `//go:build` build tags), no import of `internal/cli`'s `lockFile`/`unlockFile` (cli→preference←cli import cycle), no new build-tagged files. Stale-lock recovery: a lock older than `toggleLockStaleTimeout` (10s) is removed best-effort so a crashed holder does not permanently block toggles. `runToggle` re-routed through `TogglePersonalization` (the inline read-then-flip removed).
4. `internal/cli/preference/m4_crash_repro_test.go` — new: `TestDecayCrash_*` (runtime RED pre-M4) + `TestScanDueRace_*`.
5. `internal/cli/preference/m4_toggle_race_test.go` — new: `TestToggleRace_*` (compile RED pre-M4 — references `TogglePersonalization`).

**Cross-platform build:** `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (reconcile uses only stdlib map/slice; MarkScanned delegates to existing atomicWrite; toggle lock uses portable os.OpenFile O_EXCL — no new syscall, no build tags).
**Lint:** `golangci-lint run --timeout=3m` → 0 issues (baseline was also 0; two transient staticcheck findings in the test file — unused `due` var + empty `if` branch — fixed before commit).
**Coverage:** `go test -cover ./internal/cli/preference/` → 85.8% of statements (baseline was 86.0%; the 0.2% dip is from the new TogglePersonalization + acquireToggleLock + reconcileRecallArchival code exercised by the new tests, with lock backoff/retry edges uncovered — above the 85% threshold).
**Vet:** `go vet ./internal/cli/preference/...` exit 0; `gofmt` clean.
**Subagent boundary (C-HRA-008):** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/ | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*//'` → 0 actual calls (all matches are comments/docstrings/testdata/agentlint detector); preference package: 0 matches.
**Race suite (full, on main checkout HEAD):** `go test -race ./internal/cli/... ./internal/cli/preference/... -count=1` → all 10 packages PASS.
**Full repo test suite:** `go test ./...` → exit 0, 0 FAIL (no cascading failures per CLAUDE.local.md §6 HARD rule).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-11
run_commit_sha: 83e139bb0f53bc7fc92e9ca3f163cf4c36acd42e
run_status: complete
ac_pass_count: 6
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "0 0 (origin/main...HEAD synced at pre-flight)"
l44_post_push_fetch: pending-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 5 (decay.go, toggle.go, m4_crash_repro_test.go, m4_toggle_race_test.go, progress.md)
m1_to_mN_commit_strategy: per-milestone (M1 c31db9e2b-precursor, M2, M3 already landed; M4 this commit)
```

**Milestone closure:** M1 (5 writers re-routed + removeGLMEnv key) ✅, M2 (~/.claude.json RMW guard) ✅, M3 (atomic-writer consolidation, 2 helpers remain per option-b) ✅, M4 (preference crash-consistency + TOCTOU) ✅. All 4 milestones landed; run-phase is complete. Orchestrator proceeds to sync-phase.

**AC coverage (run-phase, all 6 AC):**
- AC-001 (M1): PASS — `TestSettingsLocalConcurrentWrites`; 0 `writeSettingsMap` callers outside settings.go.
- AC-002 (M1): PASS — `TestRemoveGLMEnvComplete`; `launcher.go` delete-set includes `EnvClaudeCodeAutoCompactWindow`.
- AC-003 (M2): PASS — 3 `TestClaudeJSONGuard_*` tests green.
- AC-004 (M3): PASS — exactly 2 atomic-writer helpers remain (consolidated `writeFileAtomic` + option-b `atomicWrite`).
- AC-005 (M4): PASS — `TestDecayCrash_*` + `TestScanDueRace_*` + `TestToggleRace_*` green under `-race`.
- AC-006 (M4): PASS — full `go test -race ./internal/cli/... ./internal/cli/preference/...` green.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-11
sync_commit_sha: pending-backfill-sync
sync_status: complete
mx_complete_at: 2026-07-11
mx_commit_sha: <NA>
mx_status: <NA>
b12_changelog_entry_position: appended to CHANGELOG.md [Unreleased] section
b12_self_tests:
  - pre-emission grep: `grep -c 'SPEC-CLIFIX-CONCURRENCY-001' CHANGELOG.md` → 0 (no duplicates)
  - ac_count_match: `grep -cE '^\| \*\*AC-[A-Z]+-[0-9]+\*\*' .moai/specs/SPEC-CLIFIX-CONCURRENCY-001/acceptance.md` → 6 entries (6 AC in SSOT acceptance.md)
  - file_path_verification: 4 test files verified via ls (all exist)
frontmatter_status_transitions:
  spec.md: status in-progress → completed, version 0.2.1 → 1.0.0, updated 2026-07-11
  plan.md: status in-progress → implemented → completed (atomic with sync commit)
  acceptance.md: status in-progress → implemented → completed (atomic with sync commit)
  progress.md: §E.4 populated (sync_commit_sha placeholder to be backfilled)
canary_compliance_check:
  - mode-5 sequential delegation: observed per progress.md §F (Mode Selection)
  - TDD RED→GREEN: all 4 milestones have reproduction tests (M1-M4)
  - cross-platform build: darwin+windows pass (verified)
  - @MX tag validation: grep confirmed in settings.go, glm_tools.go, update_noise.go
```


## §F Phase 0.95 Mode Selection

**Input parameters:**
- tier: M
- scope (file count): ~12-15 (settings.go, glm.go, launcher.go, glm_tools.go, update_noise.go, harness_mute.go, preference/{filestore,decay,toggle}.go + new race/repro tests)
- domain count: 1-2 (internal/cli concurrency + atomic-write; single Go module)
- file language mix: 100% Go (implementation + tests)
- concurrency benefit: LOW (coding-heavy, sequential dependency between milestones M1→M4)
- Agent Teams prereqs: NOT MET — workflow.team.enabled=false (Sonnet 5/Opus 4.8 default-off); CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS unset; harness level not `thorough`

**Mode evaluation:**
| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | non-trivial concurrency hardening across 6+ files |
| 2 background | no | write tasks (implementation); background writes raise main-session prompts |
| 3 agent-team | no | team.enabled=false — capability gate fails |
| 4 parallel | no | coding-heavy Go work; sequential milestone dependency (M1 seam reuse → M2 → M3 → M4); Anthropic coding-task parallelism caveat |
| 5 sub-agent | **yes** | coding-heavy, sequential per-milestone delegation; default fallback |
| 6 workflow | no | <30 files, semantic concurrency edits (not mechanical uniform transform) |

**Decision:** sub-agent (Mode 5, sequential manager-develop per milestone)

**Justification:** This is coding-heavy Go concurrency work with tight sequential dependency (M1 re-routes writers through the existing CRITICAL-001 seam; M2/M3/M4 build on that). Per Anthropic's coding-task parallelism caveat, sequential sub-agent delegation is the safe default. Agent Teams (Mode 3) is unavailable (team.enabled=false), and the scope is neither ≥30-file mechanical (Mode 6) nor research-parallel (Mode 4). Mode 5 with cycle_type=tdd (quality.yaml `development_mode: tdd`, AG-01 backward-compat) drives RED reproduction tests (race-detector failures) → GREEN fixes per milestone. Implementation Kickoff Approval obtained (plan→run HUMAN GATE, score-independent per REQ-ATR-015).

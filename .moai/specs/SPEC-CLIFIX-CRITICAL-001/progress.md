# SPEC-CLIFIX-CRITICAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- 2026-07-10: plan-phase artifact set authored (spec.md / plan.md / acceptance.md) by manager-spec from CLI audit 2026-07-10 §1 + §5 P0. Status: draft. Pending plan-audit.

## §E.2 Run-phase Evidence

### M1 — Reproduction batch (RED confirmed 2026-07-10)

8 failing repro tests written before fixes (REQ-CRIT-001-009). Verbatim output
persisted at `.moai/state/verify/327c3427/m1-red-{cli,harness}.log`.

| Defect | Test | M1 result | Evidence |
|--------|------|-----------|----------|
| a | TestSettingsLocalPreserve_Repro | FAIL | hooks/outputStyle/model wiped by struct RMW |
| b | TestClaimTaskAppend_Repro | FAIL | ledger head overwritten (O_RDWR write at offset 0) |
| c | TestHarnessMutePreserve_Repro | FAIL | agentic_loop/team wiped by minimal-struct YAML marshal |
| d | TestRemoveHarnessBoundary_Repro | FAIL | release-update specialist + skill dir deleted by bare prefix match |
| e | TestUpdateLock_ContendedFailsFast_Repro | PASS (primitive) | lock primitive works; defect = zero prod callers (grep-verified pre-fix) |
| f | TestMigrateRollbackPreexisting_Repro | FAIL | pre-existing precious.txt removed by unconditional rollback |
| g | TestMigrateSymlinkSkip_Repro | FAIL | out-of-tree symlink target copied (os.Stat follows link) |
| h | TestTierPromotionHighWater_Repro | FAIL | before=1 after=2 (duplicate promotion appended) |

### M2 — Config/ledger integrity (GREEN)

- a: all 6 SettingsLocal write-back paths opened to map[string]any (mutateSettingsLocal
  seam + ensureSettingsLocalJSON + injectGLMEnvForTeam + injectGLMEnv + removeGLMEnv +
  syncPermissionModeToSettingsLocal — D1-INJECTION listed 5; found a 6th).
- b: ClaimTask opens O_APPEND|O_WRONLY (was O_RDWR).
- c: saveWorkflowMuteConfig uses yaml.v3 Node API (setYAMLNodeSequence).
- Commit: 9e229b81d.

### M3 — Destructive-path fixes (GREEN)

- d: harnessArtifactBelongsTo with longest-match disambiguation (all 4 sites).
- e: acquireUpdateLock wired into runUpdate (line 208) + MkdirAll robustness.
- f: transactionLog.restoredFiles snapshot; copyDir/copyFile check pre-existence.
- g: copyFile uses os.Lstat (was os.Stat).
- Commit: 60cd78157.

### M4 — Growth fix + closure (GREEN)

- h: readPromotionHighWater builds per-pattern latest-tier map; classifyHarnessPatterns
  skips WritePromotion when tier unchanged.
- Commit: (this M4 commit).

### §E.2 AC Matrix (all 11 rows PASS)

| AC | Status | Verification command | Result |
|----|--------|---------------------|--------|
| AC-CRIT-001-001 | PASS | `go test ./internal/cli/ -run SettingsLocalPreserve -count=1` | ok (exit 0) |
| AC-CRIT-001-001b | PASS | `grep -n SettingsLocal internal/cli/{glm,launcher,settings}.go \| grep MarshalIndent` | 0 matches — no closed-struct marshal on any write-back path |
| AC-CRIT-001-002 | PASS | `go test ./internal/cli/ -run ClaimTaskAppend -count=1` | ok (exit 0) |
| AC-CRIT-001-002b | PASS | `grep -n O_APPEND internal/cli/team_spawn.go` | line 324 (ClaimTask) + line 292 (AppendTask) |
| AC-CRIT-001-003 | PASS | `go test ./internal/cli/ -run HarnessMutePreserve -count=1` | ok (exit 0) |
| AC-CRIT-001-004 | PASS | `go test ./internal/cli/harness/ -run RemoveHarnessBoundary -count=1` | ok (exit 0) |
| AC-CRIT-001-005 | PASS | `go test ./internal/cli/ -run UpdateLock -count=1` + `grep acquireUpdateLock update.go` | ok + line 208 |
| AC-CRIT-001-006 | PASS | `go test ./internal/cli/ -run MigrateRollbackPreexisting -count=1` | ok (exit 0) |
| AC-CRIT-001-007 | PASS | `go test ./internal/cli/ -run MigrateSymlinkSkip -count=1` + `grep Lstat migrate_agency.go` | ok + line 488 |
| AC-CRIT-001-008 | PASS | `go test ./internal/cli/ -run TierPromotionHighWater -count=1` | ok (exit 0) |
| AC-CRIT-001-009 | PASS | `go test ./internal/cli/... -count=1` | 9/9 packages ok (exit 0) |

Verbatim evidence persisted at `.moai/state/verify/327c3427/m4-final-{suite,lint}.log`
and `m4-windows-build.log`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-07-10"
run_commit_sha: "a1c38328e"
run_status: complete
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: true
l44_post_push_fetch: true
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
total_run_phase_files: 11
m1_to_mN_commit_strategy: "M-separated (5703512c8 M1, daba6b7ab M2, 0bace43be M3, a1c38328e M4) — rebased onto origin/main 713a36909 then fast-forward pushed"
```

- E1 AC matrix: 11/11 PASS (see §E.2 table above).
- E2 Cross-platform build: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- E3 Coverage: internal/cli 71.6%, internal/cli/harness 74.7% — not lower than baseline (tests only added).
- E4 Boundary grep: 0 NEW AskUserQuestion invocations in modified files (32 pre-existing doc strings across internal/cli, baseline — none are calls).
- E5 Lint: `golangci-lint run ./internal/cli/...` → 0 issues (baseline was 0; 0 new).
- E6 Race: `go test -race ./internal/cli/ -run 'ClaimTask|UpdateLock' -count=1` → exit 0.


## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-07-10"
sync_commit_sha: "pending-backfill"   # D3 SHA placeholder — backfilled in follow-up commit
run_commit_sha: "a1c38328e"           # M4 closure (substantive final run commit)
sync_status: complete
ac_count_verified: 11                 # 9 numbered ACs verified via 11 matrix rows (incl. 001b/002b grep)
ac_fail_count: 0
close_subject_full_id: "SPEC-CLIFIX-CRITICAL-001"
three_phase_close_infix: "3-phase close"
orchestrator_direct_sync: true        # manager-docs delegation aborted at context-window limit during B12 impl-file read; orchestrator-direct per CLAUDE.local.md §13 GLM orchestrator-direct sync/Mx fallback
changelog_dup_check_pre: "grep -c SPEC-CLIFIX-CRITICAL-001 CHANGELOG.md = 0"
mx_tag_scan: "0 unresolved @MX:TODO (all [AUTO] annotation tags are pre-existing, not added by this SPEC)"
ownership_note: "frontmatter in-progress→completed performed orchestrator-direct. OwnershipTransitionRule MAY emit an advisory finding (canonical owner = manager-docs); documented here as the context-limit fallback."
cross_platform_build:
  darwin_amd64: pass
  windows_amd64: pass
race_check: "go test -race ./internal/cli/ -run 'ClaimTask|UpdateLock' -count=1 → exit 0"
lint_baseline_preserved: true         # golangci-lint 0 issues pre and post
```

- S1 Frontmatter transition: spec.md `status: in-progress → completed` (this sync commit).
- S2 §E.4 sync_commit_sha: populated via D3 backfill pattern (placeholder in this commit; real SHA in the follow-up backfill commit).
- S3 CHANGELOG: [Unreleased] Fixed entry added; `grep -c` was 0 pre-sync; AC count derived from acceptance.md SSOT (11 verification rows).
- S4 MX tag scan: 0 unresolved `@MX:TODO` in the 11 run-phase files (repro tests carry `SPEC-CLIFIX-CRITICAL-001 REQ-CRIT-001-00X` annotations; all `[AUTO]` tags pre-existing).
- S5 close-subject full-ID: `chore(SPEC-CLIFIX-CRITICAL-001): sync-phase artifacts + 3-phase close` (full individual SPEC-ID + `3-phase close` infix).
- S6 push state: see backfill commit (sync + backfill pushed together).

## §F Phase 0.95 Mode Selection

- 2026-07-10: Phase 0.5 plan-audit verdict — PASS 0.88 (Tier M threshold 0.80 met; skip-eligible 0.90 not met). Report: `.moai/reports/plan-audit/SPEC-CLIFIX-CRITICAL-001-review-1.md`. 8 anchors re-verified CONFIRMED against live tree. Defects D1 (SHOULD-FIX, plan prose undercount of defect-(a) write-back surface) + D2 (MINOR, BDD ACs) — non-BLOCKING. Implementation Kickoff Approval: OBTAINED (user selected "PASS 수용, run 진입"; D1 5-paths enumeration injected into manager-develop delegation prompt to zero out under-scope risk).

- Input parameters:
  - tier: M
  - scope: ~10-15 files (8 source anchors + repro tests; defect (a) spans settings.go + glm.go + launcher.go)
  - domain count: 1 (internal/cli, single Go domain)
  - language mix: 100% Go
  - concurrency benefit: LOW (coding-heavy)
  - Agent Teams prereqs: not satisfied (team.enabled default false + coding-heavy)

- Mode evaluation:
  - Mode 1 trivial: not selected (8 multi-file defects)
  - Mode 2 background: not selected (write work, synchronous verification required)
  - Mode 3 agent-team: not selected (team prereqs unmet; coding-heavy → Mode 4/5 preferred)
  - Mode 4 parallel: not selected (single domain, coding-heavy — Anthropic coding-task parallelism caveat)
  - Mode 5 sub-agent: **selected** (default fallback; sequential per-milestone delegation)
  - Mode 6 workflow: not selected (8 defects each require a distinct fix mechanism — not a single uniform mechanical transform)

- Decision: Mode 5 (sub-agent, sequential)
- Justification: 8 Critical defects concentrated in a single Go domain (internal/cli) each requiring a distinct fix mechanism (map round-trip / O_APPEND / yaml Node API / prefix boundary / lock wiring / snapshot-restore / Lstat / high-water mark). Per Anthropic's coding-task parallelism caveat, sequential sub-agent delegation is the safe default for coding-heavy work. Mode 6 (workflow) is excluded — it admits only a single uniform mechanical transform with no inter-file dependency, which the 8 distinct fix mechanisms do not satisfy.

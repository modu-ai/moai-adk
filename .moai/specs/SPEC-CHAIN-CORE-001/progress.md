# SPEC-CHAIN-CORE-001 — Progress

> **Lifecycle**: plan → run → sync (3-phase)
> **Status**: draft (plan-phase)

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-CHAIN-CORE-001
plan_status: audit-ready
plan_complete_at: 2026-08-13
tier: L
artifact_count: 5
milestone_count: 7
req_count: 21
ac_count: 24
```

**Frontmatter schema check**: PASS (12 canonical fields present)
**SPEC ID regex check**: PASS (`SPEC-CHAIN-CORE-001` matches specIDPattern)
**ID uniqueness**: PASS (0 existing `SPEC-CHAIN-*` SPECs)
**Out of Scope rule**: PASS (7 `### Out of Scope — <topic>` H3 sub-headings with `-` bullets)
**GEARS notation**: PASS (20 REQs in Ubiquitous / When / While / Where / shall-not patterns)

---

## §E.2 Run-phase Evidence

### AC Binary PASS/FAIL Matrix

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-CHAIN-001 (MUST) | PASS | `go doc .../chain.WorktreeNode` | 13 fields: NodeID, ParentNodeID, Depth, OriginChain, WorktreePath, SessionID, SpecID, Milestone, EnteredAt, ExitedAt, LastCompletedMilestone, ResumeTarget, ResumeCommand |
| AC-CHAIN-002 (MUST) | PASS | `go test ./internal/chain/ -run TestAppendDoesNotOverwrite -v` | `--- PASS: TestAppendDoesNotOverwrite (0.01s)` |
| AC-CHAIN-003 (MUST) | PASS | `go test ./internal/chain/ -run TestCorruptLineTolerance -v` | `--- PASS: TestCorruptLineTolerance (0.01s)` + WARN corrupt line logged |
| AC-CHAIN-004 (MUST) | PASS | `go test ./internal/chain/ -run TestCWDCollisionResolution -v` | `--- PASS: TestCWDCollisionResolution (0.01s)` |
| AC-CHAIN-005 (MUST) | PASS | `go test ./internal/chain/ -run TestSpawnBoundaryNodeCreation -v` | `--- PASS: TestSpawnBoundaryNodeCreation (0.01s)` |
| AC-CHAIN-006 (MUST) | PASS | `go doc .../config.EnvChainNodeID` | `EnvChainNodeID = "MOAI_CHAIN_NODE_ID"` |
| AC-CHAIN-007 | PASS | `go test ./internal/cli/ -run TestChainStatus -v` | `--- PASS: TestChainStatus (0.00s)` |
| AC-CHAIN-008 | PASS | `go test ./internal/cli/ -run TestChainLineage -v` | `--- PASS: TestChainLineage (0.00s)` |
| AC-CHAIN-009 | PASS | `go test ./internal/cli/ -run TestChainBack -v` | `--- PASS: TestChainBack (0.00s)` |
| AC-CHAIN-010 | PASS | `go test ./internal/cli/ -run TestChainList -v` | `--- PASS: TestChainList (0.00s)` |
| AC-CHAIN-011 | PASS | `go test ./internal/chain/ -run TestChainPruneAgeThreshold -v` | `--- PASS: TestChainPruneAgeThreshold (0.01s)` |
| AC-CHAIN-012 | PASS | `go test ./internal/hook/ -run TestChainEventHook -v` | `--- PASS: TestChainEventHook (0.00s)` |
| AC-CHAIN-013 | PASS | `go test ./internal/hook/ -run TestSessionStartLineageBanner -v` | `--- PASS: TestSessionStartLineageBanner (0.00s)` |
| AC-CHAIN-014 (MUST) | PASS | `go test ./internal/hook/ -run TestChainBannerTimeout -v` | `--- PASS: TestChainBannerTimeout (0.00s)` |
| AC-CHAIN-015 (MUST) | PASS | `grep -cE 'ChainNodeID\|OriginChain\|...' registry.go` | `0` (Entry frozen) |
| AC-CHAIN-016 | PASS | `go test ./internal/cli/ -run TestClassifyStaleness -v` | `--- PASS: TestClassifyStaleness (0.00s)` |
| AC-CHAIN-017 | PASS | `go test ./internal/cli/ -run TestChainRemoteCWD -v` | `--- PASS: TestChainRemoteCWD (0.00s)` |
| AC-CHAIN-018 (MUST) | PASS | `grep -rn 'MOAI_KANBAN\|...' internal/chain/ internal/cli/chain.go ...` | `PASS: 0 matches` |
| AC-CHAIN-019 (MUST) | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/chain/ internal/cli/chain.go ...` | `PASS: 0 matches` |
| AC-CHAIN-020 | PASS | template neutrality grep | `PASS: no SPEC IDs in template` |
| AC-CHAIN-021 (MUST) | PASS | `go test ./internal/chain/ -run TestStoreIsolation -v` | `--- PASS: TestStoreIsolation (0.01s)` |
| AC-CHAIN-022 (MUST) | PASS | `go test ./internal/chain/... -count=1` | `ok github.com/modu-ai/moai-adk/internal/chain 0.434s` |
| AC-CHAIN-023 (MUST) | PASS | `go test ./internal/chain/ -run TestSessionIDBackfill -v` | `--- PASS: TestSessionIDBackfill (0.01s)` |
| AC-CHAIN-024 (MUST) | PASS | `sed -n '/type Entry struct/,/^}/p' registry.go \| grep -cE 'ChainNodeID\|...'` | `0` |

**Summary**: 24/24 AC PASS (14 MUST-PASS + 6 SHOULD-PASS + 4 NICE-TO-HAVE)

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-13
run_commit_sha: bd0a34f37
run_status: audit-ready
ac_pass_count: 24
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a
l44_post_push_fetch: n/a
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: PASS
  windows_amd64: PASS
total_run_phase_files: 17
m1_to_mN_commit_strategy: per-milestone conventional commits (M0-M6, 7 commits)
```

**Coverage**: `go test -cover ./internal/chain/` → `87.4% of statements` (≥85% threshold)
**Vet**: `go vet ./internal/chain/... ./internal/cli/... ./internal/hook/...` → clean
**Build**: `go build ./...` → exit 0
**Cross-platform**: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
**Entry freeze (AC-CHAIN-024)**: 0 chain-specific fields on Entry struct
**Subagent boundary (AC-CHAIN-018/019)**: 0 grep matches for kanban/factory/lead + AskUserQuestion
**Template neutrality (AC-CHAIN-020)**: 0 SPEC IDs, 0 macOS-bias paths in template chain artifacts
**Pre-existing CLI failures (not chain-related)**: TestDoctorGolden_{Light,Dark,NoColor} — 3 pre-existing golden-file failures unrelated to chain code

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this section>_

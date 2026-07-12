# SPEC-HARNESS-EVOLVE-002 — Progress

> Lifecycle tracking document. The `§E.*` namespace is parser-load-bearing
> (`internal/spec/era.go` `ClassifyEra` matches the literal `§E.2` / `§E.3` /
> `§E.4` heading tokens + the `sync_commit_sha` field — see
> `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md
> Section Map). Renaming any `§E.N` heading would silently break era
> classification. This file is authored by manager-spec at plan-phase
> (§E.1 only); §E.2-§E.4 are placeholder headings left for run-phase /
> sync-phase owners.

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at:
plan_artifact_count: 6
plan_tier: L
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - design.md
  - research.md
  - progress.md
plan_req_count: 36
plan_ac_count: 53
plan_needs_clarification_count: 3
plan_needs_clarification_items:
  - H-1-claude-local-md-section-marker-naming
  - H-2-mergeSectionBased-recognition-mechanism
  - H-3-debug-cli-verb-scope
plan_commit_subject: "feat(SPEC-HARNESS-EVOLVE-002): plan-phase artifacts (L, 5 artifacts)"
plan_depends_on: SPEC-HARNESS-EVOLVE-001
plan_depends_on_status: completed
plan_era: V3R6
# plan-audit: iter-1 PASS 0.86 (D1-D7 amended) → iter-2 PASS 0.91 (audit-ready).
plan_audit_final_verdict: PASS
plan_audit_final_score: 0.91
plan_audit_final_iter: 2
```

## §E.2 Run-phase Evidence

### M1 — Typed Managed-Block Writer foundation

**Deliverables shipped:**
- `internal/harness/curator/writer.go` — `WriteManagedBlock(path, blockType, content)` + `BlockType` enum (`BlockTypeLearnedWorkflow`, `BlockTypeHarnessGenerated`) + `BlockContent`/`Bullet` types + per-`BlockType` marker registry (generalizes `layer3.go markerBlockPattern`).
- `internal/harness/curator/crud.go` — `AddBullet`/`UpdateBullet`/`DeleteBullet` per-bullet CRUD (REQ-HEV2-007).
- `internal/harness/curator/{writer,crud,subagent_boundary}_test.go` — table-driven tests with `t.TempDir()` isolation.
- `internal/harness/layer3.go` — refactored `InjectMarker` into a thin wrapper over `curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)` (REQ-HEV2-002, AP-HEV2-004). `markerBlockPattern` + `buildMarkerBlock` removed; replaced by `buildHarnessBody` (returns `startAttrs` + `body` byte-identical to legacy output).

**AC PASS/FAIL matrix (M1-scoped ACs):**

| AC | Status | Verification |
|----|--------|-------------|
| AC-HEV2-001 (WriteManagedBlock signature) | PASS | `TestWriteManagedBlock_Signature` — `go test ./internal/harness/curator/...` ok |
| AC-HEV2-002 (atomic replace) | PASS | `TestWriteManagedBlock_AtomicReplace` — 1 heading, 1 start marker, 1 end marker after replace |
| AC-HEV2-003 (BlockTypeHarnessGenerated enum + D1 byte-identical) | PASS | `TestBlockTypeEnum_IncludesHarnessGenerated` — byte-exact golden comparison + existing `layer3_test.go` (7 tests) pass unchanged |
| AC-HEV2-004 (idempotent zero-byte-diff) | PASS | `TestWriteManagedBlock_Idempotent_ZeroByteDiff` |
| AC-HEV2-005 (byte preservation) | PASS | `TestWriteManagedBlock_PreBlockPostBlockPreserved` |
| AC-HEV2-006 (append-mode newline convention) | PASS | `TestWriteManagedBlock_AppendMode_NewlineConvention` (2 subtests) |
| AC-HEV2-009 (AddBullet single-line) | PASS | `TestAddBullet_SingleLine_RewritesOnlyTarget` |
| AC-HEV2-010 (UpdateBullet by ledger_key) | PASS | `TestUpdateBullet_ByLedgerKey` |
| AC-HEV2-011 (DeleteBullet preserves others) | PASS | `TestDeleteBullet_ByLedgerKey_PreservesOthers` |
| AC-HEV2-044 (subagent boundary) | PASS | `TestWriter_NoAskUserQuestion` + grep `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/curator/ \| grep -v _test.go \| grep -v //` → 0 matches |
| AC-HEV2-047 (reachability: WriteManagedBlock exists) | PASS | `grep -rn "func WriteManagedBlock" internal/harness/curator/` → ≥1 |
| AC-HEV2-048 (reachability: layer3 calls curator.WriteManagedBlock) | PASS | `grep -n "curator.WriteManagedBlock(" internal/harness/layer3.go` → 1 match (line 33) |

**Test output:** `go test -v ./internal/harness/curator/...` — 23 tests PASS (0 fail). `go test ./internal/harness/...` — all harness sub-packages ok.

**Coverage:** `go test -cover ./internal/harness/curator/...` → 92.4% statement coverage (target ≥85% per QG1).

**D1 load-bearing verification:** the production caller `internal/cli/harness/install.go:85` (`harness.InjectMarker`) delegates through the refactored `InjectMarker` → `curator.WriteManagedBlock(path, BlockTypeHarnessGenerated, ...)`. The byte-identical golden test (`TestBlockTypeEnum_IncludesHarnessGenerated`) verifies the HarnessGenerated block format is preserved exactly. The existing `layer3_test.go` suite (7 tests covering fresh-file, different-content-idempotent, same-spec-id-idempotent, append-without-trailing-newline, empty-path, empty-specID, file-not-found) passes unchanged — confirming the production path is preserved.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at:
run_commit_sha: pending-backfill-m1
run_status: m1-complete-m2-m7-pending
# M1 (Typed Managed-Block Writer foundation) is complete: curator/writer.go +
# curator/crud.go + layer3.go InjectMarker generalization. All M1-scoped ACs
# PASS (AC-HEV2-001..006, 009..011, 044, 047, 048). Coverage 92.4%.
# Remaining milestones M2-M7 are NOT started — run_status is NOT audit-ready.
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 0  # no PRESERVE-list files modified
l44_pre_commit_fetch: "0	0 (origin/main...HEAD at commit time)"
l44_post_push_fetch: pending-push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  go_build_all: exit_0
  go_build_windows_amd64: exit_0
total_run_phase_files: 6  # 3 new curator src + 3 new curator test + 1 modified layer3.go
m1_to_mN_commit_strategy: per-milestone
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

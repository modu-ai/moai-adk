# Progress — SPEC-UPDATE-REINSTALL-LOOP-001

## §E.1 Plan-phase Audit-Ready Signal

_Plan-phase artifacts authored (spec.md + plan.md + acceptance.md + research.md + this progress.md).
`plan_status` is intentionally unset — the orchestrator sets the audit-ready signal after plan-audit._

- artifacts: spec.md, plan.md, acceptance.md, research.md, progress.md
- open clarifications: resolved 2026-07-17 (model-pin=merge-preserve; R2=split to SPEC-UPDATE-PREFLIGHT-SAFETY-001; version-format=informational; preservation AC=all sections/*.yaml)
- plan_status: _<pending plan-audit — orchestrator-owned>_

## §E.2 Run-phase Evidence

cycle_type: tdd (RED-GREEN-REFACTOR). Milestones: M2 (loop-break, commit 94eb1b58a) executed first as the Critical hotfix; M1 (config-preservation) second. Scope R1+R3 only — R2 (AC-RIL-008/009) SPLIT to SPEC-UPDATE-PREFLIGHT-SAFETY-001 (no DoD weight).

### AC PASS/FAIL matrix (excludes [SPLIT →] AC-RIL-008/009)

| AC | REQ | Sev | Status | Verification command | Actual output |
|----|-----|-----|--------|----------------------|---------------|
| AC-RIL-001 | REQ-RIL-001 | Critical | PASS | `go test -run TestDeprecatedPaths_NoTemplateCollision ./internal/cli/` | PASS — `.claude/rules/moai/design` removed from DeprecatedPaths; intersection empty |
| AC-RIL-002 | REQ-RIL-003 | Critical | PASS | `go test -run 'TestDeprecatedPaths_NoTemplateCollision\|TestDeprecatedPaths_CollisionGuardDetectsReinsertion' ./internal/cli/` | PASS (guard) + PASS (negative-path: re-inserted colliding entry detected) |
| AC-RIL-003 | REQ-RIL-002 | Critical | PASS | `go test -run 'TestUpdate_ZeroNetChange_DesignDirNoLongerTriggersLoop\|TestRunCleanReinstall_ZeroRemovalOnDesignOnlyV3' ./internal/cli/` | PASS — v3 design-dir-only project resolves IsV2=false; clean-reinstall removes 0 |
| AC-RIL-004 | REQ-RIL-004 | High | PASS | `go test ./internal/defs/...` | PASS — TestDeprecatedPathsTotalCount(40) + CategorySplit(9/28/3) + CategoryBExpectedEntries updated atomically |
| AC-RIL-005 | REQ-RIL-007 | High | PASS | `go test -run TestCleanReinstall_SettingsJSONUserKeysPreserved ./internal/cli/` | PASS — effortLevel=high + user permission survive force-deploy clobber |
| AC-RIL-006 | REQ-RIL-008 | High | PASS | `go test -run TestCleanReinstall_AllSectionsYAMLPreserved ./internal/cli/` | PASS — user.yaml name, language.yaml conversation_language, design.yaml default_framework all survive |
| AC-RIL-007 | REQ-RIL-009 | High | PASS | `go test -run TestCleanReinstall_MatchesNormalPathProtection ./internal/cli/` | PASS — same backup.RestoreMoaiConfig + updatemerge.MergeUserFiles as normal path; user keys survive an active clobber |
| AC-RIL-010 | REQ-RIL-010 | Medium | PASS | `go test -run TestCleanReinstall_ModelPinDoesNotDowngrade ./internal/cli/` | PASS — user model=opus survives; template sonnet pin does not downgrade |
| AC-RIL-008 | REQ-RIL-005 | Medium | DEFERRED | — | [SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001] no DoD weight |
| AC-RIL-009 | REQ-RIL-006 | Medium | DEFERRED | — | [SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001] no DoD weight |

### Invariants

| Invariant | Status | Evidence |
|-----------|--------|----------|
| Existing test suite unbroken | PASS | `go test ./...` — only pre-existing `internal/config` TestAuditLoaderCompleteness (delegation loader) FAILs; confirmed pre-existing on clean tree (stash-verified) |
| No template edit (model-pin=merge-preserve) | PASS | no `internal/template/templates/` change → no `make build` needed |
| Cross-platform build | PASS | `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Subagent boundary (no AskUserQuestion) | PASS | grep on modified files → 0 matches |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-17
run_commit_sha: 59f52dc8e   # M2=94eb1b58a (loop-break); M1=59f52dc8e config-preservation
run_status: audit-ready
ac_pass_count: 8        # AC-RIL-001..007 + 010
ac_fail_count: 0
ac_deferred_count: 2    # AC-RIL-008/009 SPLIT → SPEC-UPDATE-PREFLIGHT-SAFETY-001
preserve_list_post_run_count: 0        # no PRESERVE-scope files modified outside SPEC envelope
l44_pre_commit_fetch: "0 0 (synced with origin/main)"
l44_post_push_fetch: pending-backfill
new_warnings_or_lints_introduced: 0    # golangci-lint ./internal/defs/... ./internal/cli/ → 0 issues (baseline 6 errcheck in internal/cli/update/merge untouched)
cross_platform_build:
  linux_darwin_native: pass
  windows_amd64: pass
total_run_phase_files: 6   # dirs.go, dirs_test.go, v2_detection.go, update_clean_install.go + 2 new test files
m1_to_mN_commit_strategy: "M2 first (Critical hotfix, 94eb1b58a) → M1 (config-preservation); 2 commits, single push"
coverage_note: "internal/cli package 74.4% (large pre-existing baseline, not regressed — new code added with tests); internal/defs dirs.go data-only (no statements)"
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs>_

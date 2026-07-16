# Progress — SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-16
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (3-file Tier M set) + progress.md
- REQ count: 8 (REQ-DFS-001 .. REQ-DFS-008)
- AC count: 9 (AC-DFS-001 .. AC-DFS-009)
- SPEC ID self-check: `^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$` → PASS
- plan-phase evidence: embedded template = 27 moai-* skills; static allowlist = 22; drift = exactly 10 unknown (matches issue #1088 report)

## §E.2 Run-phase Evidence

cycle_type=tdd (RED → GREEN → REFACTOR). Both RED reproduction subsets demonstrably failed on the pre-fix tree, then passed post-fix.

| AC | REQ(s) | Defect | Test name | Actual Output | Status |
|----|--------|--------|-----------|---------------|--------|
| AC-DFS-001 | REQ-DFS-002, 004 | A | `TestRunHarnessCheck_TelemetryOnlyReportsOK` (+`_TelemetryDirsOnlyReportsOK`) | RED pre-fix: `got fail: L1:PASS L2:FAIL … L5:FAIL`; post-fix CheckOK "contains only runtime telemetry (no harness configured)", no L5 battery | PASS |
| AC-DFS-002 | REQ-DFS-003 | A | `TestRunHarnessCheck_FullBaselineWithTelemetryRunsBattery` | full L1-L6 battery runs (message carries `L1: … L6:`), not short-circuited; CheckOK | PASS |
| AC-DFS-003 | REQ-DFS-003 | A | `TestRunHarnessCheck_PartialBaselineWithTelemetryStillFails` | partial baseline → CheckFail with `L5:FAIL` (not masked) | PASS |
| AC-DFS-004 | REQ-DFS-005, 006 | B | `TestCheckSkillsAllowlist_TemplateFreshZeroUnknown` | RED pre-fix: `warn: 10 unknown moai- skill(s) detected` (byte-match #1088); post-fix CheckOK, 0 unknown | PASS |
| AC-DFS-005 | REQ-DFS-001, 005 | B | `TestClassifySkill_EmbeddedAntiDrift` | RED pre-fix: 10 embedded skills classified WARN; post-fix every embedded moai-* skill non-WARN | PASS |
| AC-DFS-006 | REQ-DFS-007 | B | `TestCheckSkillsAllowlist_GenuineUnknownWarns` | bogus `moai-nonexistent-xyz` → CheckWarn, `1 unknown moai- skill` | PASS |
| AC-DFS-007 | REQ-DFS-002, 005 | A+B | `TestDoctor_Current_{Light,Dark}`, `TestDoctor_NoColor` (golden) | golden runs against a bare cwd (no `.claude/skills/` / `.moai/harness/`) so neither defect surfaces; golden output UNCHANGED and passing — dedicated reproduction tests carry the defect coverage (plan §G / D4) | PASS |
| AC-DFS-008 | REQ-DFS-008 | preserve | `-run Preserve\|Namespace\|CleanInstall` suites | all green; `git diff --name-only origin/main` excludes update.go / update_clean_install.go / applier.go | PASS |
| AC-DFS-009 | (cross) | all | `go test ./...` + cross-platform build | changed pkgs (internal/cli, internal/template) green; `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0; 2 pre-existing/flaky failures OUTSIDE scope (see Residual-risk) | PASS-WITH-DEBT |

Per-function coverage (changed funcs): `runHarnessCheck` 98.2%, `harnessConfigured` 85.7%, `knownCoreSkills` 87.5%, `classifySkill` 90.0%, `checkSkillsAllowlist` 80.8%. Package: internal/cli 74.4% (pre-existing baseline, large package), internal/template 85.3%.

@MX tags added (plan §G): 2 ANCHOR+REASON (`doctor_harness.go` telemetry-exclusion gate; `doctor_skills.go` classifySkill manifest-derived) + 3 NOTE (`harnessConfigured` helper; `doctor_skills.go` SSOT var block; `doctor.go` remediation hint) + 1 bonus NOTE (`skills_manifest.go` SSOT source). Per-file limits respected.

Stale-assertion correction (SPEC-required): `doctor_hns_test.go` `TestClassifySkill_HNS` pinned `moai-harness-learner`→WARN, which encoded the exact #1088 drift; `moai-harness-learner` is an embedded skill so AC-DFS-005 requires non-WARN → assertion updated to PASS.

Files changed: `internal/cli/doctor_harness.go`, `internal/cli/doctor_skills.go`, `internal/cli/doctor.go`, `internal/cli/doctor_harness_test.go`, `internal/cli/doctor_skills_test.go`, `internal/cli/doctor_hns_test.go` (modified) + `internal/template/skills_manifest.go`, `internal/template/skills_manifest_test.go` (new).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-16
run_commit_sha: d6405b3953b325fef7e5722e2abf26d5e6e6ede6
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
ac_pass_with_debt_count: 1   # AC-DFS-009 (pre-existing/flaky failures outside scope)
preserve_list_post_run_count: 0   # update.go / update_clean_install.go / applier.go un-diffed
l44_pre_commit_fetch: n/a (worktree, orchestrator handles landing)
l44_post_push_fetch: n/a (do-not-push per spawn contract)
new_warnings_or_lints_introduced: 0   # go vet clean; golangci-lint findings only in pre-existing merge_test.go
cross_platform_build:
  darwin_build: pass
  windows_amd64_build: pass
total_run_phase_files: 8   # 6 modified + 2 new
m1_to_mN_commit_strategy: single run-phase commit (Tier M, all M1-M5 milestones)
red_reproduction_confirmed: true   # AC-DFS-001/004/005 failed pre-fix, pass post-fix
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_

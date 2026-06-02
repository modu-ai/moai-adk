---
id: SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001
artifact: acceptance.md
version: "0.1.0"
created: 2026-05-22
updated: 2026-06-02
---

# Acceptance Criteria — SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001

## Overview

8 binary ACs each mapped 1-to-1 with REQ-WSE-001..008. Every AC includes an explicit verification command and a binary PASS/FAIL outcome. No subjective criteria.

## REQ↔AC Traceability (100%)

| REQ | AC | Coverage |
|-----|----|----------|
| REQ-WSE-001 (nested struct fields) | AC-WSE-001 | ✅ |
| REQ-WSE-002 (TeamConfig sub-struct) | AC-WSE-002 | ✅ |
| REQ-WSE-003 (yaml.Unmarshal correctness) | AC-WSE-003 | ✅ |
| REQ-WSE-004 (FLAT field rename + deprecation) | AC-WSE-004 | ✅ |
| REQ-WSE-005 (accessor methods) | AC-WSE-005 | ✅ |
| REQ-WSE-006 (LoadRoleProfiles migration) | AC-WSE-006 | ✅ |
| REQ-WSE-007 (NewDefaultWorkflowConfig defaults) | AC-WSE-007 | ✅ |
| REQ-WSE-008 (audit_loader_completeness_test.go cleanup) | AC-WSE-008 | ✅ |

---

## AC-WSE-001 — Nested struct field reachability (REQ-WSE-001)

**Given** the new `WorkflowConfig` type definition in `internal/config/types.go`
**When** Go compilation succeeds and reflection enumerates the struct
**Then** the following dotted-path field access expressions SHALL all compile and resolve to non-zero-value types (zero-value content acceptable):
- `cfg.Workflow.AutoClear.Enabled` (bool)
- `cfg.Workflow.AutoClear.AfterPlan` (bool)
- `cfg.Workflow.AutoClear.AfterRun` (bool)
- `cfg.Workflow.AutoClear.TokenThreshold` (int)
- `cfg.Workflow.Completion.DetectInOutput` (bool)
- `cfg.Workflow.Completion.Markers.Complete` (string)
- `cfg.Workflow.Completion.Markers.Done` (string)
- `cfg.Workflow.DefaultMode` (string)
- `cfg.Workflow.ExecutionMode` (string)
- `cfg.Workflow.LoopPrevention.FailurePatternDetection` (bool)
- `cfg.Workflow.LoopPrevention.MaxIterations` (int)
- `cfg.Workflow.LoopPrevention.MaxRetriesPerOperation` (int)
- `cfg.Workflow.Memory.AuditEnabled` (bool)
- `cfg.Workflow.Memory.IndexLineCap` (int)
- `cfg.Workflow.Memory.StaleAggregateThreshold` (int)
- `cfg.Workflow.Memory.StalenessThresholdHours` (int)
- `cfg.Workflow.TokenBudget.Plan` (int)
- `cfg.Workflow.TokenBudget.Run` (int)
- `cfg.Workflow.TokenBudget.Sync` (int)
- `cfg.Workflow.Worktree.AutoCleanup` (bool)
- `cfg.Workflow.Worktree.AutoCreate` (bool)
- `cfg.Workflow.Worktree.AutoMerge` (bool)
- `cfg.Workflow.Worktree.SessionNamePattern` (string)
- `cfg.Workflow.Worktree.TmuxPreferred` (bool)

**Verification command**:
```bash
go build ./internal/config/... && go test -run TestWorkflowConfigNestedFieldReachability ./internal/config/...
```
Expected output: `PASS` (exit 0). The test uses `reflect.TypeOf(WorkflowConfig{}).FieldByName(...)` to assert each nested path exists.

**Outcome**: PASS = all 24 paths reachable; FAIL = any path missing.

---

## AC-WSE-002 — TeamConfig sub-struct presence (REQ-WSE-002)

**Given** the new `WorkflowConfig` type definition
**When** Go compilation succeeds and reflection enumerates `WorkflowConfig.Team`
**Then** `WorkflowConfig.Team` field type SHALL be `TeamConfig` containing exactly the following exported fields:
- `AutoSelection TeamAutoSelectionConfig`
- `Enabled bool`
- `MaxTeammates int`
- `DefaultModel string`
- `DelegateMode bool`
- `RequirePlanApproval bool`
- `RoleProfileKeys []string`
- `RoleProfiles map[string]RoleProfileEntry`

And `RoleProfileEntry` SHALL contain exactly:
- `Description string`
- `Isolation string`
- `Mode string`
- `Model string`

And the field `Patterns` SHALL NOT be present (EXCL-WSE-004).

**Verification command**:
```bash
go test -run TestTeamConfigStructShape ./internal/config/...
```
Expected output: `PASS`. The test uses `reflect.VisibleFields(reflect.TypeOf(TeamConfig{}))` to assert field set equality.

**Outcome**: PASS = exact field set match; FAIL = any field missing, extra (including `Patterns`), or wrong type.

---

## AC-WSE-003 — yaml.Unmarshal correctness (REQ-WSE-003)

**Given** the production `internal/template/templates/.moai/config/sections/workflow.yaml` template SSOT (or `.moai/config/sections/workflow.yaml` from the local project as proxy)
**When** the yaml content is loaded via the standard `config.LoadAll(projectRoot)` path
**Then** the resulting `*Config` value SHALL satisfy:
- `cfg.Workflow.Team.RoleProfiles["implementer"].Isolation == "worktree"`
- `cfg.Workflow.Team.RoleProfiles["implementer"].Mode == "acceptEdits"`
- `cfg.Workflow.Team.RoleProfiles["tester"].Isolation == "worktree"`
- `cfg.Workflow.Team.RoleProfiles["reviewer"].Isolation == "none"`
- `cfg.Workflow.Team.RoleProfiles["reviewer"].Mode == "plan"`
- `len(cfg.Workflow.Team.RoleProfiles) == 7`
- `cfg.Workflow.AutoClear.Enabled == true`
- `cfg.Workflow.AutoClear.TokenThreshold == 150000`
- `cfg.Workflow.Completion.Markers.Done == "<moai>DONE</moai>"`
- `cfg.Workflow.Completion.Markers.Complete == "<moai>COMPLETE</moai>"`
- `cfg.Workflow.TokenBudget.Plan == 30000`
- `cfg.Workflow.TokenBudget.Run == 180000`
- `cfg.Workflow.TokenBudget.Sync == 40000`
- `cfg.Workflow.Team.AutoSelection.MinDomainsForTeam == 3`
- `cfg.Workflow.Team.AutoSelection.MinFilesForTeam == 10`
- `cfg.Workflow.Team.AutoSelection.MinComplexityScore == 7`
- `cfg.Workflow.LoopPrevention.MaxIterations == 100`
- `cfg.Workflow.Memory.IndexLineCap == 200`
- `cfg.Workflow.Worktree.SessionNamePattern == "moai-{ProjectName}-{SPEC-ID}"`
- `cfg.Workflow.Worktree.TmuxPreferred == true`

**Verification command**:
```bash
go test -run TestWorkflowYAMLUnmarshalProductionFixture ./internal/config/...
```
Expected output: `PASS`. The test loads the embedded template yaml (or in-tree `.moai/config/sections/workflow.yaml` if newer) and asserts each value.

**Outcome**: PASS = all 20 assertions satisfied; FAIL = any value mismatch (signals yaml-path bug or default drift).

---

## AC-WSE-004 — FLAT field rename + deprecation (REQ-WSE-004)

**Given** the updated `WorkflowConfig` type
**When** `go doc` is invoked on the struct fields
**Then**:
- `WorkflowConfig` SHALL contain a field named `AutoClearLegacy` (not `AutoClear` as bool) with a doc comment starting with `// Deprecated: use AutoClear.Enabled`.
- `WorkflowConfig` SHALL contain a field named `AutoSelectionLegacy` (not `AutoSelection`) with a doc comment starting with `// Deprecated: use Team.AutoSelection`.
- `WorkflowConfig.PlanTokens`, `RunTokens`, `SyncTokens` SHALL retain their identifiers and each have a doc comment starting with `// Deprecated: use TokenBudget.Plan` (and analogous for Run/Sync).
- `grep -n "AutoClear bool" internal/config/types.go` SHALL match zero lines (rename completed).
- `grep -n "AutoSelection TeamAutoSelectionConfig" internal/config/types.go | grep -v "Team.AutoSelection"` SHALL match zero lines (FLAT struct member rename completed; the new nested `Team.AutoSelection TeamAutoSelectionConfig` remains).

**Verification command**:
```bash
go test -run TestWorkflowConfigLegacyFieldsRenamed ./internal/config/... \
  && grep -c "AutoClear bool" internal/config/types.go | grep -q "^0$" \
  && grep -c "^	AutoSelection TeamAutoSelectionConfig" internal/config/types.go | grep -q "^0$"
```
Expected output: `PASS` from go test + both grep return zero.

**Outcome**: PASS = renames applied and deprecation comments present; FAIL = any FLAT identifier still present.

---

## AC-WSE-005 — Accessor method behavior (REQ-WSE-005)

**Given** a populated `*Config` value
**When** the following accessor methods are invoked
**Then** each SHALL return the value from the nested location:
- `(*Config).WorkflowAutoClearEnabled() bool` returns `c.Workflow.AutoClear.Enabled`
- `(*Config).WorkflowPlanTokens() int` returns `c.Workflow.TokenBudget.Plan`
- `(*Config).WorkflowRunTokens() int` returns `c.Workflow.TokenBudget.Run`
- `(*Config).WorkflowSyncTokens() int` returns `c.Workflow.TokenBudget.Sync`
- `(*Config).WorkflowTeamAutoSelection() TeamAutoSelectionConfig` returns `c.Workflow.Team.AutoSelection`

Table-driven test asserts 5 cases each with a distinct nested value and verifies the accessor returns exactly that value.

**Verification command**:
```bash
go test -run TestWorkflowAccessorMethods -v ./internal/config/...
```
Expected output: `PASS` with all 5 sub-test cases visible (`--- PASS: TestWorkflowAccessorMethods/AutoClearEnabled`, etc.).

**Outcome**: PASS = all 5 accessors return nested value; FAIL = any accessor returns wrong value or fails to compile.

---

## AC-WSE-006 — LoadRoleProfiles migration (REQ-WSE-006)

**Given** the updated `internal/cli/team_spawn.go`
**When** `LoadRoleProfiles(workflowPath)` is invoked with a path to a valid `workflow.yaml` containing 7 role_profiles entries
**Then**:
- The returned `map[string]RoleProfile` SHALL contain exactly 7 entries (`analyst`, `architect`, `designer`, `implementer`, `researcher`, `reviewer`, `tester`).
- Each entry's `Mode`, `Model`, `Isolation`, `Description` fields SHALL match the yaml values.
- `WriteHeavy` boolean SHALL be `true` for `implementer`, `tester`, `designer` and `false` for the other 4.
- The function body SHALL NOT contain the string-parsing block (lines previously at 421-475 in the unmigrated version: `for _, line := range lines`, `strings.HasPrefix(trimmed, "role_profiles:")`, `indent == 2`, `indent == 4`, etc.).
- The function body SHALL invoke `config.LoadAll(filepath.Dir(filepath.Dir(workflowPath)))` (or equivalent canonical loader) to obtain the typed struct.

**Verification commands**:
```bash
# Behavior preservation
go test -run TestLoadRoleProfiles -v ./internal/cli/...

# Migration completeness
grep -c "strings.HasPrefix(trimmed, \"role_profiles:\")" internal/cli/team_spawn.go | grep -q "^0$"
grep -c "indent := len(line) - len(strings.TrimLeft(line, \" \"))" internal/cli/team_spawn.go | grep -q "^0$"

# Loader integration
grep -q "config.LoadAll\|cfg\.Workflow\.Team\.RoleProfiles" internal/cli/team_spawn.go
```
Expected output: `PASS` from test + first 2 greps return zero + 3rd grep finds match.

**Outcome**: PASS = test green AND ad-hoc parser removed AND typed loader integrated; FAIL = any of the 3 conditions unmet.

---

## AC-WSE-007 — NewDefaultWorkflowConfig defaults (REQ-WSE-007)

**Given** the updated `NewDefaultWorkflowConfig()` function
**When** the function is invoked
**Then** the returned `WorkflowConfig` value SHALL satisfy ALL of the following assertions:
- `cfg.AutoClear.Enabled == true`
- `cfg.AutoClear.AfterPlan == true`
- `cfg.AutoClear.AfterRun == false`
- `cfg.AutoClear.TokenThreshold == 150000`
- `cfg.Completion.DetectInOutput == true`
- `cfg.Completion.Markers.Done == "<moai>DONE</moai>"`
- `cfg.Completion.Markers.Complete == "<moai>COMPLETE</moai>"`
- `cfg.LoopPrevention.FailurePatternDetection == true`
- `cfg.LoopPrevention.MaxIterations == 100`
- `cfg.LoopPrevention.MaxRetriesPerOperation == 3`
- `cfg.Memory.AuditEnabled == true`
- `cfg.Memory.IndexLineCap == 200`
- `cfg.Memory.StaleAggregateThreshold == 10`
- `cfg.Memory.StalenessThresholdHours == 24`
- `cfg.TokenBudget.Plan == 30000`
- `cfg.TokenBudget.Run == 180000`
- `cfg.TokenBudget.Sync == 40000`
- `cfg.Team.Enabled == true`
- `cfg.Team.MaxTeammates == 10`
- `cfg.Team.DefaultModel == "sonnet"`
- `cfg.Team.DelegateMode == true`
- `cfg.Team.RequirePlanApproval == true`
- `cfg.Team.AutoSelection.MinDomainsForTeam == 3`
- `cfg.Team.AutoSelection.MinFilesForTeam == 10`
- `cfg.Team.AutoSelection.MinComplexityScore == 7`
- `len(cfg.Team.RoleProfileKeys) == 3` with values `["implementer", "tester", "reviewer"]`
- `len(cfg.Team.RoleProfiles) == 7` (per AC-WO-009 contract)
- `cfg.Team.RoleProfiles["implementer"].Isolation == "worktree"`
- `cfg.Team.RoleProfiles["implementer"].Mode == "acceptEdits"`
- `cfg.Team.RoleProfiles["implementer"].Model == "sonnet"`
- `cfg.Team.RoleProfiles["researcher"].Model == "haiku"`
- `cfg.Worktree.AutoCleanup == true`
- `cfg.Worktree.AutoCreate == false`
- `cfg.Worktree.AutoMerge == true`
- `cfg.Worktree.SessionNamePattern == "moai-{ProjectName}-{SPEC-ID}"`
- `cfg.Worktree.TmuxPreferred == true`

**Verification command**:
```bash
go test -run TestNewDefaultWorkflowConfigNestedDefaults -v ./internal/config/...
```
Expected output: `PASS` with all 36 assertions in the table-driven test passing.

**Outcome**: PASS = all defaults match `workflow.yaml` (template SSOT); FAIL = any default mismatch (signals template drift).

---

## AC-WSE-008 — audit_loader_completeness exception cleanup (REQ-WSE-008)

**Given** the updated `internal/config/audit_loader_completeness_test.go`
**When** the file is inspected and the test is executed
**Then**:
- `grep -n "\"workflow\"" internal/config/audit_loader_completeness_test.go` SHALL match zero lines containing `"workflow"` inside the `acknowledgedUnloadedSections` slice (the entry at the pre-migration line 27 is REMOVED).
- The `TestAuditLoaderCompleteness` test SHALL PASS with the workflow section now counted under `loaded` (via `Loader.Load()` Config.Workflow field populated by yaml.Unmarshal — i.e., workflow has a covering loader because Config.Workflow is populated as part of the main config load).

**Verification commands**:
```bash
# Exception removed
! grep -n '"workflow"' internal/config/audit_loader_completeness_test.go | grep -q "acknowledged"

# Test passes with workflow counted as loaded
go test -run TestAuditLoaderCompleteness -v ./internal/config/...
```
Expected: first grep returns exit 1 (no match) due to leading `!` negation — overall command exit 0. Second test PASS.

**Outcome**: PASS = exception removed AND test green; FAIL = exception still present OR test red.

---

## Edge Case Coverage

Edge cases (Edge-WSE-001..004 in spec.md §5) MUST each have at least one assertion in `workflow_nested_test.go` or `defaults_test.go`. Specific mapping:

- Edge-WSE-001 (empty role_profiles map): asserted in `TestLoadRoleProfiles_EmptyMap` (existing test extension)
- Edge-WSE-002 (role_profile_keys not subset): asserted in `TestWorkflowConfigInconsistentRoleProfileKeys` (new test, observation only, no error raised per EXCL-WSE-005)
- Edge-WSE-003 (omitted token_budget block): asserted in `TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults`
- Edge-WSE-004 (legacy FLAT yaml `auto_clear: true` scalar against new struct field): asserted in `TestWorkflowYAMLUnmarshal_LegacyFlatYamlTypeMismatch_BehaviorDocumented` (covers either flexible-unmarshal accept OR yaml warning emit — implementation choice documented in test)

---

## Definition of Done

The SPEC is complete when ALL of the following hold simultaneously:

1. AC-WSE-001 through AC-WSE-008 each PASS via the documented verification commands.
2. `go build ./...` exits 0.
3. `GOOS=windows GOARCH=amd64 go build ./...` exits 0 (cross-platform).
4. `go test ./internal/config/... ./internal/cli/...` exits 0.
5. `go test -race ./internal/config/... ./internal/cli/...` exits 0.
6. Coverage of `internal/config` package ≥ 85% (existing baseline preserved).
7. `golangci-lint run --timeout=2m` introduces 0 NEW issues (pre-existing baseline acceptable).
8. spec-lint emits 0 NEW errors against this SPEC (`### N.M Out of Scope` h3 heading present).
9. C-HRA-008 subagent boundary: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/ internal/cli/team_spawn.go internal/cli/workflow_lint.go | grep -v _test | grep -v "^[^:]*:[0-9]*:[ \t]*//"` matches zero.
10. Edge cases Edge-WSE-001..004 each have at least one assertion in test code.
11. SPEC frontmatter `status` advanced from `draft` to `implemented` and version bumped to `0.2.0`.

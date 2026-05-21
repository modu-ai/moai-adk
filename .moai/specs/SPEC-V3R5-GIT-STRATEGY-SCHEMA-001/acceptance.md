---
id: SPEC-V3R5-GIT-STRATEGY-SCHEMA-001
title: "git-strategy.yaml ↔ Go struct 정합 — Acceptance Criteria"
version: "0.1.0"
created: 2026-05-22
updated: 2026-05-22
---

# Acceptance Criteria — SPEC-V3R5-GIT-STRATEGY-SCHEMA-001

All ACs below are **binary PASS/FAIL** with explicit verification commands. AC-GSS-001..006 are tied to specific REQ-GSS-XXX (must-pass). AC-GSS-007..009 verify cross-cutting concerns. Tier M threshold per `.claude/rules/moai/workflow/spec-workflow.md` (≥ 80% AC pass rate, ALL must-pass = PASS).

---

## AC-GSS-001 — Struct Hierarchy Defined (REQ-GSS-001, REQ-GSS-002)

**Status**: must-pass

**Given** the existing `internal/config/types.go` containing FLAT `GitStrategyConfig` struct (lines 41-48),
**When** the SPEC implementation lands,
**Then** the file shall define a `GitStrategyConfig` struct containing:
- Top-level: `Mode string` with tag `` `yaml:"mode"` ``, `Provider string` with tag `` `yaml:"provider"` ``, `GitHubUsername string` with tag `` `yaml:"github_username"` ``, `GitLab GitLabConfig` with tag `` `yaml:"gitlab"` ``
- Mode profiles: `Manual ModeProfile` with tag `` `yaml:"manual"` ``, `Personal ModeProfile` with tag `` `yaml:"personal"` ``, `Team ModeProfile` with tag `` `yaml:"team"` ``
- Deprecated (preserved): `AutoBranch bool` with tag `` `yaml:"auto_branch"` ``, `BranchPrefix string`, `CommitStyle string`, `WorktreeRoot string`, `GitLabInstanceURL string`
- Each new sub-struct (`ModeProfile`, `AutomationConfig`, `BranchCreationConfig`, `CommitStyleConfig`, `HooksConfig`, `GitLabConfig`) exists with the fields enumerated in REQ-GSS-002.

**Verification (binary)**:
```bash
go doc internal/config GitStrategyConfig 2>&1 | grep -E "Mode|Provider|GitHubUsername|GitLab|Manual|Personal|Team|AutoBranch|BranchPrefix|CommitStyle|WorktreeRoot" | wc -l
# Expected output: 12 (or more) lines containing the listed identifiers

go doc internal/config ModeProfile 2>&1 | grep -E "Automation|BranchCreation|CommitStyle|Hooks|Workflow|Environment|GitHubIntegration|PushToRemote|AutoCheckpoint|BranchPrefix|MainBranch|DraftPR|RequiredReviews|BranchProtection" | wc -l
# Expected output: 14 (or more) lines

go doc internal/config GitLabConfig 2>&1 | grep -E "InstanceURL" | wc -l
# Expected: 1 line
```

**PASS** if all three commands produce expected line counts.

---

## AC-GSS-002 — yaml.Unmarshal Nested Roundtrip (REQ-GSS-003)

**Status**: must-pass

**Given** a yaml fixture matching the current production `.moai/config/sections/git-strategy.yaml` structure (3 modes fully populated, ~70 keys),
**When** the test `TestGitStrategyConfig_NestedRoundTrip` in `internal/config/git_strategy_nested_test.go` runs,
**Then** all of the following assertions shall pass:
- `cfg.Mode == "team"` (top-level mode)
- `cfg.Provider == "github"` (top-level provider)
- `cfg.GitLab.InstanceURL == ""` (top-level nested empty)
- `cfg.Team.BranchCreation.AutoEnabled == false` (Late-Branch default)
- `cfg.Team.BranchCreation.PromptAlways == true`
- `cfg.Team.Automation.AutoBranch == false`
- `cfg.Team.Automation.AutoPush == true`
- `cfg.Team.CommitStyle.ScopeRequired == true`
- `cfg.Manual.Environment == "local"`
- `cfg.Personal.BranchPrefix == "feature/SPEC-"`

**Verification (binary)**:
```bash
go test -run TestGitStrategyConfig_NestedRoundTrip ./internal/config/...
# Expected: PASS
```

**PASS** if exit code 0 and no `FAIL` in output.

---

## AC-GSS-003 — ActiveModeProfile Accessor (REQ-GSS-004)

**Status**: must-pass

**Given** a `GitStrategyConfig` value with `Mode` set to one of `"manual"`, `"personal"`, `"team"`, or empty/invalid,
**When** the test `TestGitStrategyConfig_ActiveModeProfile` in `internal/config/git_strategy_nested_test.go` runs covering all four cases,
**Then**:
- `Mode == "manual"`: returns `(&cfg.Manual, true)`
- `Mode == "personal"`: returns `(&cfg.Personal, true)`
- `Mode == "team"`: returns `(&cfg.Team, true)`
- `Mode == ""`: returns `(nil, false)`
- `Mode == "unknown"`: returns `(nil, false)`

**Verification (binary)**:
```bash
go test -run TestGitStrategyConfig_ActiveModeProfile ./internal/config/...
# Expected: PASS

# Sub-test count check:
go test -v -run TestGitStrategyConfig_ActiveModeProfile ./internal/config/... 2>&1 | grep -c "^=== RUN" 
# Expected: 6 (1 parent + 5 sub-tests)
```

**PASS** if both commands succeed.

---

## AC-GSS-004 — Backward-Compat FLAT Fields Preserved (REQ-GSS-005)

**Status**: must-pass

**Given** existing `internal/config/types_test.go` test cases asserting FLAT field values (`cfg.GitStrategy.AutoBranch`, `cfg.GitStrategy.BranchPrefix`, `cfg.GitStrategy.CommitStyle`, line 52-86),
**When** the SPEC implementation lands,
**Then**:
- Existing FLAT field tests in `types_test.go` shall continue to pass without modification (backward-compat preserved).
- `go doc` for `GitStrategyConfig.AutoBranch` shall contain a `// Deprecated:` comment line referencing this SPEC.
- `go doc` for `GitStrategyConfig.WorktreeRoot` shall contain a `// Deprecated:` comment line.

**Verification (binary)**:
```bash
go test -run TestConfig ./internal/config/... 2>&1 | grep -E "PASS|FAIL"
# Expected: PASS (no FAIL)

grep -n "Deprecated:" internal/config/types.go | grep -i "GitStrategy\|AutoBranch\|WorktreeRoot\|FLAT" | wc -l
# Expected: ≥ 2 (at minimum AutoBranch and WorktreeRoot deprecation comments)
```

**PASS** if both checks succeed.

---

## AC-GSS-005 — Default Values Match Template (REQ-GSS-006)

**Status**: must-pass

**Given** `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` containing canonical defaults for each ModeProfile,
**When** `NewDefaultGitStrategyConfig()` is called,
**Then** the returned value shall satisfy at minimum:
- `cfg.Team.Automation.AutoPush == true`
- `cfg.Team.Automation.AutoBranch == false`
- `cfg.Team.CommitStyle.ScopeRequired == true`
- `cfg.Team.BranchProtection == true`
- `cfg.Team.DraftPR == true`
- `cfg.Team.RequiredReviews == 1`
- `cfg.Manual.Environment == "local"`
- `cfg.Manual.PushToRemote == false`
- `cfg.Manual.GitHubIntegration == false`
- `cfg.Manual.AutoCheckpoint == "disabled"`
- `cfg.Personal.BranchPrefix == "feature/SPEC-"`
- `cfg.Personal.PushToRemote == true`

**Verification (binary)**:
```bash
go test -run TestNewDefaultGitStrategyConfig ./internal/config/...
# Expected: PASS

go test -v -run TestNewDefaultGitStrategyConfig ./internal/config/... 2>&1 | grep -c "PASS:"
# Expected: ≥ 1 (parent test passes)
```

**PASS** if both succeed.

---

## AC-GSS-006 — Extended Validation Coverage (REQ-GSS-007)

**Status**: must-pass

**Given** `internal/config/validation.go` containing FLAT-field validation (lines 216-217),
**When** the SPEC implementation lands,
**Then** the file shall contain `checkStringField` invocations for at minimum:
- `git_strategy.team.workflow`
- `git_strategy.personal.workflow`
- `git_strategy.manual.workflow`
- `git_strategy.team.environment`
- `git_strategy.team.commit_style.format`
- `git_strategy.team.hooks.pre_commit`
- `git_strategy.team.hooks.pre_push`
- `git_strategy.team.hooks.commit_msg`
- `git_strategy.team.branch_prefix`
- Pre-existing `git_strategy.branch_prefix` and `git_strategy.commit_style` lines shall remain (backward-compat).

**Verification (binary)**:
```bash
grep -c 'checkStringField("git_strategy\.' internal/config/validation.go
# Expected: ≥ 11 (2 pre-existing FLAT + ≥ 9 new nested)

grep -E 'git_strategy\.(team|personal|manual)\.(workflow|environment|commit_style|hooks|branch_prefix)' internal/config/validation.go | wc -l
# Expected: ≥ 9
```

**PASS** if both grep counts meet thresholds.

---

## AC-GSS-007 — Audit Registry Updated (REQ-GSS-008)

**Status**: must-pass

**Given** `internal/config/audit_loader_completeness_test.go` line 19 containing `"git-strategy",   // out-of-scope: loaded via git-strategy.yaml.tmpl template rendering path`,
**When** the SPEC implementation lands,
**Then**:
- The line shall be removed from `acknowledgedUnloadedSections` slice.
- `internal/config/audit_loader_completeness_test.go` shall continue to compile and pass.

**Verification (binary)**:
```bash
grep -c '"git-strategy"' internal/config/audit_loader_completeness_test.go
# Expected: 0 (entry removed from acknowledgedUnloadedSections; the file may still reference git-strategy in comments)

# Confirm via context (entry inside slice, not comment):
awk '/acknowledgedUnloadedSections = \[\]string/,/^}/' internal/config/audit_loader_completeness_test.go | grep -c '"git-strategy"'
# Expected: 0

go test -run TestAuditLoaderCompleteness ./internal/config/...
# Expected: PASS
```

**PASS** if both grep counts are 0 and test passes.

---

## AC-GSS-008 — Cross-Platform Build PASS

**Status**: must-pass

**Given** the SPEC implementation lands,
**When** cross-platform build verification runs,
**Then**:
- `go build ./...` succeeds (default GOOS).
- `GOOS=windows GOARCH=amd64 go build ./...` succeeds.
- `GOOS=linux GOARCH=amd64 go build ./...` succeeds.

**Verification (binary)**:
```bash
go build ./... && echo "PASS-darwin"
GOOS=windows GOARCH=amd64 go build ./... && echo "PASS-windows"
GOOS=linux GOARCH=amd64 go build ./... && echo "PASS-linux"
# Expected: three "PASS-*" lines printed; exit code 0 for all
```

**PASS** if all three echo outputs appear.

---

## AC-GSS-009 — Subagent Boundary + Lint Clean

**Status**: must-pass

**Given** the SPEC implementation lands,
**When** boundary and lint checks run,
**Then**:
- C-HRA-008 grep yields 0 matches in modified files (no AskUserQuestion in config/ which is subagent territory).
- `golangci-lint run --timeout=2m` produces 0 NEW issues compared to pre-implementation baseline (baseline captured in plan-phase M0).
- `go test -race ./internal/config/...` PASS.

**Verification (binary)**:
```bash
# C-HRA-008 boundary check
grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/ 2>/dev/null | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
# Expected: (empty output)

# Race detector
go test -race ./internal/config/...
# Expected: PASS

# Lint NEW = 0 (baseline comparison)
golangci-lint run --timeout=2m ./internal/config/... 2>&1 | grep -E "^internal/config" | wc -l
# Compare to baseline captured in plan-phase progress.md M0; NEW count = current - baseline ≤ 0
```

**PASS** if all three checks succeed (grep 0 matches, race PASS, lint NEW ≤ 0).

---

## Coverage Summary (REQ ↔ AC Traceability)

| REQ | AC | Must-pass | Tier M threshold |
|-----|----|-----------| -----------------|
| REQ-GSS-001 (struct hierarchy) | AC-GSS-001 | ✅ | included |
| REQ-GSS-002 (ModeProfile sub-structs) | AC-GSS-001 | ✅ | included |
| REQ-GSS-003 (yaml roundtrip) | AC-GSS-002 | ✅ | included |
| REQ-GSS-004 (ActiveModeProfile) | AC-GSS-003 | ✅ | included |
| REQ-GSS-005 (FLAT backward-compat) | AC-GSS-004 | ✅ | included |
| REQ-GSS-006 (default values) | AC-GSS-005 | ✅ | included |
| REQ-GSS-007 (validation extension) | AC-GSS-006 | ✅ | included |
| REQ-GSS-008 (audit registry) | AC-GSS-007 | ✅ | included |
| (cross-cutting: build) | AC-GSS-008 | ✅ | included |
| (cross-cutting: lint/race/boundary) | AC-GSS-009 | ✅ | included |

100% REQ → AC traceability. 8 REQs ↔ 9 ACs (REQ-001 + REQ-002 share AC-001 by design, both verified via single `go doc` check).

---

End of acceptance.md.

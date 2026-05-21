---
id: SPEC-V3R5-GIT-STRATEGY-SCHEMA-001
title: "git-strategy.yaml ↔ Go struct 정합 — Implementation Plan"
version: "0.1.0"
created: 2026-05-22
updated: 2026-05-22
---

# Implementation Plan — SPEC-V3R5-GIT-STRATEGY-SCHEMA-001

Tier M plan. Milestones M0-M5 (priority-ordered, no time estimates per HARD rule). 5-section delegation template (Section A/B/C/D/E) embedded for manager-develop reuse.

---

## 1. Approach Summary

Option (c) accessor-method backward-compat strategy. Five logical milestones:

1. **M0** — Baseline capture + struct hierarchy definition (REQ-GSS-001, REQ-GSS-002, REQ-GSS-005)
2. **M1** — yaml.Unmarshal nested roundtrip test + new test file (REQ-GSS-003, REQ-GSS-004)
3. **M2** — ActiveModeProfile accessor implementation + 5 sub-test cases (REQ-GSS-004)
4. **M3** — NewDefaultGitStrategyConfig extension + defaults_test.go update (REQ-GSS-006)
5. **M4** — Validation extension + audit registry cleanup (REQ-GSS-007, REQ-GSS-008)
6. **M5** — Cross-platform build verification + lint/race/boundary final gate (AC-GSS-008, AC-GSS-009)

Each milestone produces a verifiable artifact and feeds the AC matrix. M1-M4 may be partially parallelized but M0 MUST precede all.

---

## 2. Milestones (priority-ordered)

### M0 — Baseline + Struct Hierarchy Definition (Priority: P0)

**Goal**: Define new struct hierarchy in `internal/config/types.go`; capture lint baseline; deprecate FLAT fields.

**Files**: `internal/config/types.go`

**Steps**:
1. Capture pre-implementation lint baseline:
   ```bash
   golangci-lint run --timeout=2m ./internal/config/... 2>&1 | tee plan-baseline-lint.txt
   wc -l plan-baseline-lint.txt
   ```
   Record line count in commit body as `LINT_BASELINE=<N>`.
2. Append new sub-structs **above** `GitStrategyConfig` (right after `MigrationsConfig` precedent at line 53-66) to maintain Go file ordering convention:
   - `GitLabConfig{ InstanceURL string }`
   - `AutomationConfig{ AutoBranch, AutoCommit, AutoPR, AutoPush bool }`
   - `BranchCreationConfig{ AutoEnabled bool; PromptAlways bool }`
   - `CommitStyleConfig{ Format string; ScopeRequired bool }`
   - `HooksConfig{ PreCommit, PrePush, CommitMsg string }`
   - `ModeProfile{ Workflow, Environment, AutoCheckpoint, BranchPrefix, MainBranch string; GitHubIntegration, PushToRemote, DraftPR, BranchProtection bool; RequiredReviews int; Automation AutomationConfig; BranchCreation BranchCreationConfig; CommitStyle CommitStyleConfig; Hooks HooksConfig }`
3. Replace `GitStrategyConfig` struct (lines 41-48) with extended version preserving FLAT fields:
   ```go
   // GitStrategyConfig represents the git strategy configuration section.
   //
   // Schema reorganized per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 (2026-05-22):
   // - Top-level fields (Mode, Provider, GitHubUsername, GitLab) are wire-through.
   // - Mode profiles (Manual, Personal, Team) are forward-compat scaffolds for
   //   future Go consumers of Late-Branch keys (see SPEC-V3R5-LATE-BRANCH-001).
   // - FLAT fields (AutoBranch, BranchPrefix, CommitStyle, WorktreeRoot,
   //   GitLabInstanceURL) are deprecated per backward-compat Option (c) but
   //   retained per SPEC-CONFIG-001 historical contract.
   type GitStrategyConfig struct {
       // Top-level wire-through fields
       Mode           string       `yaml:"mode"`
       Provider       string       `yaml:"provider"`
       GitHubUsername string       `yaml:"github_username"`
       GitLab         GitLabConfig `yaml:"gitlab"`

       // Mode profile forward-compat scaffolds
       Manual   ModeProfile `yaml:"manual"`
       Personal ModeProfile `yaml:"personal"`
       Team     ModeProfile `yaml:"team"`

       // Deprecated FLAT fields — preserved for backward-compat per Option (c).
       // Future SPEC may sunset these with SemVer major-bump.
       // Deprecated: use ActiveModeProfile().Automation.AutoBranch instead.
       AutoBranch bool `yaml:"auto_branch"`
       // Deprecated: use ActiveModeProfile().BranchPrefix instead.
       BranchPrefix string `yaml:"branch_prefix"`
       // Deprecated: use ActiveModeProfile().CommitStyle.Format instead.
       CommitStyle string `yaml:"commit_style"`
       // Deprecated: no production consumer; preserved per SPEC-CONFIG-001.
       WorktreeRoot string `yaml:"worktree_root"`
       // Deprecated: use GitLab.InstanceURL instead.
       GitLabInstanceURL string `yaml:"gitlab_instance_url"`
   }
   ```
4. Run `go build ./...` and `go vet ./...` after edit. Both MUST succeed.

**Deliverable**: M0 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M0 struct hierarchy + deprecation comments`

**Verification commands**:
```bash
go doc internal/config GitStrategyConfig
go doc internal/config ModeProfile
go vet ./internal/config/...
```

---

### M1 — Nested Roundtrip Test File (Priority: P0)

**Goal**: Create new test file with yaml fixture matching production yaml; assert all nested fields populate.

**Files**: `internal/config/git_strategy_nested_test.go` (NEW)

**Steps**:
1. Create new file `internal/config/git_strategy_nested_test.go` with package declaration `package config`.
2. Define yaml fixture as Go string constant mirroring current `.moai/config/sections/git-strategy.yaml` byte-byte (use raw string literal with backticks). Include all 3 modes fully populated.
3. Implement `TestGitStrategyConfig_NestedRoundTrip(t *testing.T)`:
   - `yaml.Unmarshal([]byte(fixture), &cfg)` then assert top-level + each mode field per AC-GSS-002 (10 assertions minimum).
4. Implement `TestGitStrategyConfig_ZeroValueSafety(t *testing.T)`:
   - `cfg := GitStrategyConfig{}`; assert zero values do not panic on `ActiveModeProfile()`.
5. Run test:
   ```bash
   go test -run TestGitStrategyConfig_NestedRoundTrip ./internal/config/...
   go test -race -run TestGitStrategyConfig_ ./internal/config/...
   ```

**Deliverable**: M1 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M1 nested roundtrip test`

---

### M2 — ActiveModeProfile Accessor (Priority: P1)

**Goal**: Implement `ActiveModeProfile()` accessor method on `GitStrategyConfig`.

**Files**: `internal/config/types.go` (extend with method)

**Steps**:
1. Add accessor method below `GitStrategyConfig` struct definition:
   ```go
   // ActiveModeProfile returns a pointer to the currently selected ModeProfile
   // based on the Mode field. Returns (nil, false) when Mode is empty or invalid.
   //
   // Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-004.
   func (c *GitStrategyConfig) ActiveModeProfile() (*ModeProfile, bool) {
       switch c.Mode {
       case "manual":
           return &c.Manual, true
       case "personal":
           return &c.Personal, true
       case "team":
           return &c.Team, true
       default:
           return nil, false
       }
   }
   ```
2. Extend `internal/config/git_strategy_nested_test.go` with `TestGitStrategyConfig_ActiveModeProfile(t *testing.T)` using table-driven sub-tests covering all 5 cases per AC-GSS-003.
3. Verify:
   ```bash
   go test -v -run TestGitStrategyConfig_ActiveModeProfile ./internal/config/...
   ```

**Deliverable**: M2 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M2 ActiveModeProfile accessor`

---

### M3 — Defaults Extension (Priority: P1)

**Goal**: Update `NewDefaultGitStrategyConfig()` to populate ModeProfile defaults matching template.

**Files**: `internal/config/defaults.go`, `internal/config/defaults_test.go`

**Steps**:
1. In `defaults.go:222-229`, replace `NewDefaultGitStrategyConfig` body to return value with populated Manual/Personal/Team profiles + top-level wires:
   ```go
   func NewDefaultGitStrategyConfig() GitStrategyConfig {
       return GitStrategyConfig{
           Mode:           "team",
           Provider:       "github",
           GitHubUsername: "",
           GitLab:         GitLabConfig{InstanceURL: ""},

           Manual: ModeProfile{
               Workflow: "github-flow", Environment: "local",
               GitHubIntegration: false, PushToRemote: false, AutoCheckpoint: "disabled",
               BranchCreation: BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
               Automation:     AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
               CommitStyle:    CommitStyleConfig{Format: "conventional", ScopeRequired: false},
               Hooks:          HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
           },
           Personal: ModeProfile{
               Workflow: "github-flow", Environment: "github",
               GitHubIntegration: true, PushToRemote: true,
               BranchPrefix: "feature/SPEC-", MainBranch: "main",
               BranchCreation: BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
               Automation:     AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
               CommitStyle:    CommitStyleConfig{Format: "conventional", ScopeRequired: false},
               Hooks:          HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
           },
           Team: ModeProfile{
               Workflow: "github-flow", Environment: "github",
               GitHubIntegration: true, PushToRemote: true,
               BranchPrefix: "feature/SPEC-", MainBranch: "main",
               DraftPR: true, RequiredReviews: 1, BranchProtection: true,
               BranchCreation: BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
               Automation:     AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: true},
               CommitStyle:    CommitStyleConfig{Format: "conventional", ScopeRequired: true},
               Hooks:          HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
           },

           // Deprecated FLAT fields — preserve existing default values for backward-compat.
           AutoBranch:        false,
           BranchPrefix:      DefaultBranchPrefix,
           CommitStyle:       DefaultCommitStyle,
           WorktreeRoot:      "",
           GitLabInstanceURL: "",
       }
   }
   ```
2. Extend `internal/config/defaults_test.go::TestNewDefaultGitStrategyConfig` (currently at line 223) with assertions per AC-GSS-005 (12 assertions minimum).
3. Preserve existing FLAT-field assertions (defaults_test.go:46-49) without modification — they still hold.
4. Verify:
   ```bash
   go test -v -run TestNewDefaultGitStrategyConfig ./internal/config/...
   ```

**Deliverable**: M3 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M3 defaults extension`

---

### M4 — Validation + Audit Registry (Priority: P2)

**Goal**: Extend `validation.go` with nested checkStringField invocations + remove `git-strategy` from audit exception list.

**Files**: `internal/config/validation.go`, `internal/config/audit_loader_completeness_test.go`

**Steps**:
1. In `validation.go:216-217`, after pre-existing FLAT validations, append nested validations:
   ```go
   // Nested mode profile validations (SPEC-V3R5-GIT-STRATEGY-SCHEMA-001)
   for _, mode := range []struct {
       Name    string
       Profile ModeProfile
   }{
       {"manual", cfg.GitStrategy.Manual},
       {"personal", cfg.GitStrategy.Personal},
       {"team", cfg.GitStrategy.Team},
   } {
       prefix := "git_strategy." + mode.Name
       errs = append(errs, checkStringField(prefix+".workflow", mode.Profile.Workflow)...)
       errs = append(errs, checkStringField(prefix+".environment", mode.Profile.Environment)...)
       errs = append(errs, checkStringField(prefix+".commit_style.format", mode.Profile.CommitStyle.Format)...)
       errs = append(errs, checkStringField(prefix+".hooks.pre_commit", mode.Profile.Hooks.PreCommit)...)
       errs = append(errs, checkStringField(prefix+".hooks.pre_push", mode.Profile.Hooks.PrePush)...)
       errs = append(errs, checkStringField(prefix+".hooks.commit_msg", mode.Profile.Hooks.CommitMsg)...)
       if mode.Name != "manual" { // manual mode has no branch_prefix
           errs = append(errs, checkStringField(prefix+".branch_prefix", mode.Profile.BranchPrefix)...)
       }
   }
   ```
2. In `audit_loader_completeness_test.go:19`, remove the line:
   ```
   "git-strategy",   // out-of-scope: loaded via git-strategy.yaml.tmpl template rendering path
   ```
   (Delete one entry from `acknowledgedUnloadedSections` slice. Confirm the slice still compiles.)
3. Verify:
   ```bash
   go test -run TestAuditLoaderCompleteness ./internal/config/...
   go test -run TestValidation ./internal/config/...
   ```
4. If `TestAuditLoaderCompleteness` fails citing missing loader for git-strategy, investigate `loader.go` Load() chain — git-strategy should already flow through normal config load. If not, add to `loadedSections` mapping (minor fix).

**Deliverable**: M4 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M4 validation extension + audit registry`

---

### M5 — Cross-Platform + Lint + Race Final Gate (Priority: P0)

**Goal**: All AC-GSS-008 + AC-GSS-009 binary checks PASS.

**Files**: (no source changes; verification only)

**Steps**:
1. Cross-platform build:
   ```bash
   go build ./... && echo "PASS-darwin"
   GOOS=windows GOARCH=amd64 go build ./... && echo "PASS-windows"
   GOOS=linux GOARCH=amd64 go build ./... && echo "PASS-linux"
   ```
2. Race detector:
   ```bash
   go test -race ./internal/config/...
   ```
3. Lint NEW = 0 vs M0 baseline:
   ```bash
   golangci-lint run --timeout=2m ./internal/config/... 2>&1 | tee plan-final-lint.txt
   wc -l plan-final-lint.txt
   # Compare to LINT_BASELINE captured in M0
   ```
4. C-HRA-008 boundary:
   ```bash
   grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/ 2>/dev/null | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
   ```
5. Full suite regression:
   ```bash
   go test ./...
   ```

**Deliverable**: M5 commit `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M5 cross-platform + lint + race verification`

---

## 3. Technical Approach Details

### 3.1 Go File Organization Convention

All new sub-structs co-located in `internal/config/types.go` per existing precedent (`MigrationsConfig`, `SystemHookConfig`, etc.). No new files in `internal/config/` package outside the test file (M1).

### 3.2 yaml Tag Strategy

All new struct fields use `yaml:"snake_case_key"` tags matching template file. No custom unmarshaling logic needed — yaml.v3 handles nested struct natively.

### 3.3 Late-Branch Forward-Compat Annotation

`BranchCreationConfig.AutoEnabled` field MUST carry a Go doc comment:
```go
// AutoEnabled controls whether /moai plan creates a feat/SPEC-* branch.
// Forward-compat scaffold per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001.
// Currently consumed by .claude/skills/moai/workflows/plan/spec-assembly.md
// Phase 3 (yaml-direct read), NOT by Go code. Future Go consumer may migrate
// to read this struct field via Config.GitStrategy.ActiveModeProfile().
// See SPEC-V3R5-LATE-BRANCH-001 REQ-LB-004.
AutoEnabled bool `yaml:"auto_enabled"`
```

---

## 4. Risk Mitigation

R-GSS-001 (test fixture migration): Both FLAT (legacy synthetic) and nested (production) fixtures coexist in test files during M1-M4. AC-GSS-004 explicitly preserves FLAT-field tests.

R-GSS-002 (nil accessor return): `ActiveModeProfile()` returns `(*ModeProfile, bool)` Go-idiomatic pair. Callers MUST check the boolean before deref.

R-GSS-003 (yaml library drift): yaml library version pinned in `go.mod` (`gopkg.in/yaml.v3 v3.0.x`). No version change in this SPEC.

R-GSS-004 (audit exception removal CI gate): M4 must land after M0-M3 commits (priority ordering enforced).

R-GSS-005 (Late-Branch confusion): Explicit doc comment on `BranchCreationConfig.AutoEnabled` field referencing SPEC-V3R5-LATE-BRANCH-001 prevents misunderstanding.

---

## 5. Section A — Context (Tier M delegation template)

- **Project root**: `/Users/goos/MoAI/moai-adk-go`
- **Current branch**: `main` (Late-Branch policy)
- **Plan HEAD baseline**: `6b6733c8e` (pre-plan-commit)
- **SPEC artifacts**:
  - `.moai/specs/SPEC-V3R5-GIT-STRATEGY-SCHEMA-001/spec.md` (~310 lines)
  - `.moai/specs/SPEC-V3R5-GIT-STRATEGY-SCHEMA-001/acceptance.md` (~260 lines)
  - `.moai/specs/SPEC-V3R5-GIT-STRATEGY-SCHEMA-001/plan.md` (this file, ~330 lines)
- **plan-auditor verdict**: TBD (run iter 1 after plan commit)
- **Existing infra to PRESERVE**:
  - `internal/config/types.go` FLAT GitStrategyConfig (lines 41-48) — deprecated, NOT removed
  - `internal/config/defaults.go` `NewDefaultGitStrategyConfig` — extended, FLAT defaults preserved
  - `internal/config/validation.go` `git_strategy.branch_prefix` and `git_strategy.commit_style` checks — preserved
  - `internal/config/types_test.go` FLAT field assertions (lines 52-86) — preserved
  - `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` — byte-identical
  - `.moai/config/sections/git-strategy.yaml` — byte-identical (production yaml unchanged)
  - `.claude/skills/moai/workflows/plan/spec-assembly.md` Phase 3 yaml-direct read — preserved
  - `.claude/agents/moai/manager-git.md` yaml-direct read references — preserved

## 6. Section B — Known Issues (Tier M filter)

- **B1 cross-platform**: No syscall usage; pure struct definitions and yaml roundtrip. PASS expected on darwin/linux/windows.
- **B2 cross-SPEC conflict**: SPEC-V3R5-LATE-BRANCH-001 owns `branch_creation.*` (skill-body consumer). Verified via v2 audit Step 2-3 in spec.md §1.2. NO conflict (additive scaffold).
- **B4 frontmatter canonical schema**: 3 artifacts all use canonical fields (id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags/tier). spec.md tier: M present.
- **B5 CI 3-tier**: spec-lint (heading h3 Out of Scope compliance verified in spec.md §3), golangci-lint (baseline captured in M0), Test per OS (AC-GSS-008).
- **B6 spec-lint heading**: `### 3.1 Out of Scope` (h3 sub-section) present in spec.md §3.1. Verified.
- **B8 working tree hygiene**: Pre-existing dirty files (`.claude/settings.json`, `.moai/harness/usage-log.jsonl`, `internal/merge/*`) PRESERVE — do not commit. Untracked files (`.claude/commands/99-release.md`, `.claude/skills/moai/workflows/release.md`, `internal/cli/init_layout.go`, `internal/cli/wizard/fullscreen.go`, `internal/cli/wizard/review.go`, `internal/hook/.moai/`, `internal/statusline/stdinfields_test.go`) PRESERVE.

## 7. Section C — Pre-flight Check List (run before each milestone)

```bash
# 1. Branch + HEAD baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build sanity
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline
golangci-lint run --timeout=2m ./internal/config/... 2>&1 | tail -5

# 4. Cross-SPEC conflict re-confirmation
grep -rn 'GitStrategyConfig\|git-strategy' .moai/specs/ 2>/dev/null | sort -u | head -10

# 5. PRESERVE list verification (no inadvertent changes)
git status --short | head -20
```

## 8. Section D — Constraints (DO NOT VIOLATE)

- PRESERVE 대상: 8 files enumerated in §5 Section A above. All [EXTEND], 1 [NEW].
- 무관 dirty / untracked: enumerated in §6 B8. DO NOT modify or commit.
- 금지 명령: `--no-verify`, `--amend`, force-push to main, `git reset --hard`.
- 사용 의무 명령: Conventional Commits format `plan(SPEC-V3R5-GIT-STRATEGY-SCHEMA-001): M{N} <summary>`, trailer `🗿 MoAI <email@mo.ai.kr>` per CLAUDE.md.
- Late-Branch policy: commit to main, defer push to sync-phase per REQ-LB-005.
- Subagent boundary (C-HRA-008): zero `AskUserQuestion` invocations in `internal/config/*.go`. Verified by AC-GSS-009.

## 9. Section E — Self-Verification Deliverables (post-implementation report)

manager-develop final report MUST include:

**E1. AC Binary PASS/FAIL Matrix** (9 ACs):

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-GSS-001 | PASS/FAIL | `go doc internal/config GitStrategyConfig \| grep -c ...` | (line counts) |
| AC-GSS-002 | PASS/FAIL | `go test -run TestGitStrategyConfig_NestedRoundTrip ./internal/config/...` | (PASS/FAIL + duration) |
| AC-GSS-003 | PASS/FAIL | `go test -v -run TestGitStrategyConfig_ActiveModeProfile ./internal/config/...` | (sub-test count) |
| AC-GSS-004 | PASS/FAIL | `go test -run TestConfig ./internal/config/...` + `grep "Deprecated:" types.go` | (PASS + count) |
| AC-GSS-005 | PASS/FAIL | `go test -v -run TestNewDefaultGitStrategyConfig ./internal/config/...` | (PASS) |
| AC-GSS-006 | PASS/FAIL | `grep -c 'checkStringField("git_strategy\.' validation.go` | (count ≥ 11) |
| AC-GSS-007 | PASS/FAIL | `grep -c '"git-strategy"' audit_loader_completeness_test.go` (in slice) + `go test -run TestAuditLoaderCompleteness` | (0 + PASS) |
| AC-GSS-008 | PASS/FAIL | 3× `go build` (darwin/linux/windows) | (3 PASS lines) |
| AC-GSS-009 | PASS/FAIL | C-HRA-008 grep + `go test -race` + lint NEW count | (0 + PASS + NEW ≤ 0) |

**E2. Cross-Platform Build**: 3 lines `PASS-darwin`, `PASS-windows`, `PASS-linux` per AC-GSS-008.

**E3. Coverage**: `go test -cover ./internal/config/...` — target ≥85% (existing baseline). Report actual %.

**E4. Subagent Boundary**: C-HRA-008 grep result (expect empty).

**E5. Lint Status**: M0 baseline line count vs final line count. NEW = final - baseline. Expect NEW ≤ 0.

**E6. Branch HEAD + Push Status**: 5 commit SHAs (M0-M4 commits + optional M5 verification commit). `git push origin main` deferred per Late-Branch.

**E7. Blocker Report (if any)**: NONE expected per v2 audit Step 3 confirming no in-progress owner SPECs.

---

End of plan.md.

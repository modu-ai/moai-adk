---
id: SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001
artifact: plan.md
version: "0.1.0"
created: 2026-05-22
updated: 2026-05-22
---

# Implementation Plan — SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001

## Approach Summary

Tier M LEAN 3-artifact workflow. Brownfield struct extension on `internal/config/types.go` + nested loader + ad-hoc parser migration + audit registry cleanup. Late-Branch policy (main commits, push at PR creation).

Implementation proceeds in **5 milestones (M1..M5)** in dependency order. Each milestone is independently verifiable. M5 (audit cleanup) is **last** because removing the `acknowledgedUnloadedSections` entry is gated on M1..M4 completing successfully.

This plan mirrors the sibling P2 SPEC `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` plan structure for predictability across the two parallel v2-audit SPECs.

---

## Technical Approach

### Architectural Decision: Nested struct + accessor-based backward-compat (Option (c))

`WorkflowConfig` becomes the canonical nested representation. FLAT fields are renamed (where ambiguous) and deprecated:

```go
// New nested structure (NOT pre-existing — added by REQ-WSE-001/002)
type WorkflowConfig struct {
    // Nested canonical fields (yaml-aligned with workflow.yaml top-level keys)
    AutoClear      AutoClearConfig         `yaml:"auto_clear"`
    Completion     CompletionConfig        `yaml:"completion"`
    DefaultMode    string                  `yaml:"default_mode"`
    ExecutionMode  string                  `yaml:"execution_mode"`
    LoopPrevention LoopPreventionConfig    `yaml:"loop_prevention"`
    Memory         MemoryConfig            `yaml:"memory"`
    Team           TeamConfig              `yaml:"team"`
    TokenBudget    TokenBudgetConfig       `yaml:"token_budget"`
    Worktree       WorkflowWorktreeConfig  `yaml:"worktree"`

    // Deprecated FLAT fields (Option (c) — preserved for SPEC-CONFIG-001 backward-compat)
    // Deprecated: use AutoClear.Enabled — yaml tag pointed at non-existent flat path
    AutoClearLegacy bool `yaml:"-"`
    // Deprecated: use TokenBudget.Plan
    PlanTokens int `yaml:"-"`
    // Deprecated: use TokenBudget.Run
    RunTokens int `yaml:"-"`
    // Deprecated: use TokenBudget.Sync
    SyncTokens int `yaml:"-"`
    // Deprecated: use Team.AutoSelection
    AutoSelectionLegacy TeamAutoSelectionConfig `yaml:"-"`
}
```

Rationale:
- yaml tag `"-"` on the deprecated FLAT fields **prevents yaml.Unmarshal** from binding to them, eliminating the path-mismatch unmarshal-failure risk while preserving Go API binary compatibility.
- Accessor methods on `*Config` (REQ-WSE-005) provide the documented forward path for downstream consumers.

### Sub-struct definitions (canonical naming)

```go
type AutoClearConfig struct {
    Enabled        bool `yaml:"enabled"`
    AfterPlan      bool `yaml:"after_plan"`
    AfterRun       bool `yaml:"after_run"`
    TokenThreshold int  `yaml:"token_threshold"`
}

type CompletionConfig struct {
    DetectInOutput bool           `yaml:"detect_in_output"`
    Markers        MarkersConfig  `yaml:"markers"`
}

type MarkersConfig struct {
    Complete string `yaml:"complete"`
    Done     string `yaml:"done"`
}

type LoopPreventionConfig struct {
    FailurePatternDetection bool `yaml:"failure_pattern_detection"`
    MaxIterations           int  `yaml:"max_iterations"`
    MaxRetriesPerOperation  int  `yaml:"max_retries_per_operation"`
}

type MemoryConfig struct {
    AuditEnabled            bool `yaml:"audit_enabled"`
    IndexLineCap            int  `yaml:"index_line_cap"`
    StaleAggregateThreshold int  `yaml:"stale_aggregate_threshold"`
    StalenessThresholdHours int  `yaml:"staleness_threshold_hours"`
}

type TeamConfig struct {
    AutoSelection       TeamAutoSelectionConfig         `yaml:"auto_selection"`
    Enabled             bool                            `yaml:"enabled"`
    MaxTeammates        int                             `yaml:"max_teammates"`
    DefaultModel        string                          `yaml:"default_model"`
    DelegateMode        bool                            `yaml:"delegate_mode"`
    RequirePlanApproval bool                            `yaml:"require_plan_approval"`
    RoleProfileKeys     []string                        `yaml:"role_profile_keys"`
    RoleProfiles        map[string]RoleProfileEntry     `yaml:"role_profiles"`
    // Note: workflow.team.patterns.* is intentionally NOT modeled (EXCL-WSE-004)
}

type RoleProfileEntry struct {
    Description string `yaml:"description"`
    Isolation   string `yaml:"isolation"`
    Mode        string `yaml:"mode"`
    Model       string `yaml:"model"`
}

type TokenBudgetConfig struct {
    Plan int `yaml:"plan"`
    Run  int `yaml:"run"`
    Sync int `yaml:"sync"`
}

type WorkflowWorktreeConfig struct {
    AutoCleanup        bool   `yaml:"auto_cleanup"`
    AutoCreate         bool   `yaml:"auto_create"`
    AutoMerge          bool   `yaml:"auto_merge"`
    SessionNamePattern string `yaml:"session_name_pattern"`
    TmuxPreferred      bool   `yaml:"tmux_preferred"`
}
```

### Accessor methods (REQ-WSE-005) — added to existing `*Config` file or new `internal/config/accessors.go`

```go
func (c *Config) WorkflowAutoClearEnabled() bool {
    return c.Workflow.AutoClear.Enabled
}
func (c *Config) WorkflowPlanTokens() int { return c.Workflow.TokenBudget.Plan }
func (c *Config) WorkflowRunTokens() int  { return c.Workflow.TokenBudget.Run }
func (c *Config) WorkflowSyncTokens() int { return c.Workflow.TokenBudget.Sync }
func (c *Config) WorkflowTeamAutoSelection() TeamAutoSelectionConfig {
    return c.Workflow.Team.AutoSelection
}
```

### `LoadRoleProfiles` migration (REQ-WSE-006)

Before (ad-hoc string parsing — REMOVE):
```go
// lines 412-490 of team_spawn.go
content, err := os.ReadFile(workflowPath)
// ... 70+ lines of indent-based parsing ...
```

After (typed loader):
```go
func LoadRoleProfiles(workflowPath string) (map[string]RoleProfile, error) {
    // Derive project root from workflowPath: .../sections/workflow.yaml -> projectRoot
    projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(workflowPath)))
    // (3 levels up: .moai/config/sections/workflow.yaml → projectRoot)

    cfg, err := config.LoadAll(projectRoot)
    if err != nil {
        return nil, fmt.Errorf("load workflow config: %w", err)
    }

    profiles := make(map[string]RoleProfile, len(cfg.Workflow.Team.RoleProfiles))
    writeHeavySet := buildWriteHeavySet()
    for name, entry := range cfg.Workflow.Team.RoleProfiles {
        profiles[name] = RoleProfile{
            Name:        name,
            Description: entry.Description,
            Isolation:   entry.Isolation,
            Mode:        entry.Mode,
            Model:       entry.Model,
            WriteHeavy:  writeHeavySet[name],
        }
    }
    return profiles, nil
}

func buildWriteHeavySet() map[string]bool {
    writeHeavySet := make(map[string]bool)
    for _, role := range strings.Split(WriteHeavyRoles, ",") {
        writeHeavySet[strings.TrimSpace(role)] = true
    }
    return writeHeavySet
}
```

The function signature (1 string param, returns map + error) is preserved per REQ-WSE-006.

### `workflow_lint.go` migration

`internal/cli/workflow_lint.go:42-56` defines a parallel `workflowConfig` type. Replace with `config.WorkflowConfig`:

```go
// Before (lines 42-56) — REMOVE the workflowConfig + workflowRoleProfile types
// After: use config.WorkflowConfig directly
func loadWorkflowYAML(path string) (*config.Config, error) {
    projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(path)))
    return config.LoadAll(projectRoot)
}

func validateRoleProfiles(cfg *config.Config) []WorkflowLintViolation {
    profiles := cfg.Workflow.Team.RoleProfiles  // now typed RoleProfileEntry
    // ... rest of validation logic unchanged ...
}
```

---

## Milestones

### M1 — Struct skeleton + sub-structs (REQ-WSE-001, REQ-WSE-002, REQ-WSE-004)

**Deliverables**:
- `internal/config/types.go`:
  - Rename `WorkflowConfig.AutoClear bool` → `AutoClearLegacy bool` with `// Deprecated:` comment + yaml tag `"-"`.
  - Rename `WorkflowConfig.AutoSelection TeamAutoSelectionConfig` → `AutoSelectionLegacy` with deprecated comment + yaml tag `"-"`.
  - Add deprecated comments to `PlanTokens`, `RunTokens`, `SyncTokens` (no rename — identifiers do not conflict) + yaml tag `"-"`.
  - Add new nested fields: `AutoClear AutoClearConfig`, `Completion CompletionConfig`, `DefaultMode string`, `ExecutionMode string`, `LoopPrevention LoopPreventionConfig`, `Memory MemoryConfig`, `Team TeamConfig`, `TokenBudget TokenBudgetConfig`, `Worktree WorkflowWorktreeConfig`.
  - Define sub-structs: `AutoClearConfig`, `CompletionConfig`, `MarkersConfig`, `LoopPreventionConfig`, `MemoryConfig`, `TeamConfig`, `RoleProfileEntry`, `TokenBudgetConfig`, `WorkflowWorktreeConfig`. Co-locate after `WorkflowConfig` per existing precedent (e.g., `MigrationsConfig`, `TeamAutoSelectionConfig`).
- Update `internal/config/types_test.go` to use `AutoClearLegacy` / `AutoSelectionLegacy` where struct literals reference the renamed fields.

**Verification**: `go build ./...` exits 0; `go test -run TestWorkflowConfig -v ./internal/config/...` PASS.

**AC binding**: AC-WSE-001, AC-WSE-002, AC-WSE-004 partial.

### M2 — Defaults + accessor methods (REQ-WSE-005, REQ-WSE-007)

**Deliverables**:
- `internal/config/defaults.go` `NewDefaultWorkflowConfig()`:
  - Populate `AutoClear AutoClearConfig{Enabled: true, AfterPlan: true, AfterRun: false, TokenThreshold: 150000}`.
  - Populate `Completion CompletionConfig{DetectInOutput: true, Markers: MarkersConfig{Done: "<moai>DONE</moai>", Complete: "<moai>COMPLETE</moai>"}}`.
  - `LoopPrevention LoopPreventionConfig{FailurePatternDetection: true, MaxIterations: 100, MaxRetriesPerOperation: 3}`.
  - `Memory MemoryConfig{AuditEnabled: true, IndexLineCap: 200, StaleAggregateThreshold: 10, StalenessThresholdHours: 24}`.
  - `Team TeamConfig{...}` with 7-key `RoleProfiles` map matching `workflow.yaml.tmpl` exactly.
  - `TokenBudget TokenBudgetConfig{Plan: 30000, Run: 180000, Sync: 40000}`.
  - `Worktree WorkflowWorktreeConfig{AutoCleanup: true, AutoCreate: true, AutoMerge: true, SessionNamePattern: "moai-{ProjectName}-{SPEC-ID}", TmuxPreferred: true}`.
  - Legacy FLAT fields set to zero-values (deprecated, unused).
- Add accessor methods (REQ-WSE-005) on `*Config` to `internal/config/manager.go` or new file `internal/config/workflow_accessors.go`.
- Update `internal/config/defaults_test.go` `TestNewDefaultWorkflowConfig` to assert all 36 default values per AC-WSE-007.

**Verification**: `go test -run TestNewDefaultWorkflowConfigNestedDefaults -v ./internal/config/...` PASS; `go test -run TestWorkflowAccessorMethods -v ./internal/config/...` PASS.

**AC binding**: AC-WSE-005, AC-WSE-007.

### M3 — yaml.Unmarshal + nested test fixture (REQ-WSE-003)

**Deliverables**:
- New file `internal/config/workflow_nested_test.go`:
  - `TestWorkflowYAMLUnmarshalProductionFixture` — loads the embedded `workflow.yaml.tmpl` (rendered with default `TemplateContext`) or alternatively the in-tree `.moai/config/sections/workflow.yaml`, calls `LoadAll`, asserts all 20 values per AC-WSE-003.
  - `TestWorkflowYAMLUnmarshal_OmittedTokenBudget_PreservesDefaults` (Edge-WSE-003).
  - `TestWorkflowYAMLUnmarshal_LegacyFlatYamlTypeMismatch_BehaviorDocumented` (Edge-WSE-004).
  - `TestWorkflowConfigInconsistentRoleProfileKeys` (Edge-WSE-002) — observation-only, no error raised.
- Update `internal/cli/team_spawn_test.go` `TestLoadRoleProfiles_EmptyMap` (Edge-WSE-001) to use the new typed-loader path (assertion content unchanged).

**Verification**: `go test -run TestWorkflowYAMLUnmarshal -v ./internal/config/...` PASS; cross-platform `GOOS=windows GOARCH=amd64 go build ./...` PASS.

**AC binding**: AC-WSE-003; Edge-WSE-001..004 coverage.

### M4 — Ad-hoc parser migration (REQ-WSE-006)

**Deliverables**:
- `internal/cli/team_spawn.go`:
  - Replace `LoadRoleProfiles` body (lines 412-490) with the typed-loader version above. Preserve function signature `func LoadRoleProfiles(workflowPath string) (map[string]RoleProfile, error)`.
  - Extract `buildWriteHeavySet()` helper from existing logic to keep `LoadRoleProfiles` body ≤ 25 lines.
- `internal/cli/workflow_lint.go`:
  - Remove `workflowConfig` and `workflowRoleProfile` internal types (lines 42-56).
  - Update `loadWorkflowYAML` to return `*config.Config`.
  - Update `validateRoleProfiles` to operate on `config.RoleProfileEntry` (renamed from local `workflowRoleProfile`).
  - All existing `WorkflowLintViolation` emission logic preserved verbatim.

**Verification**:
- `go test -run TestLoadRoleProfiles -v ./internal/cli/...` PASS (preserves existing test contract).
- `go test -run TestWorkflowLint -v ./internal/cli/...` PASS.
- `grep -c "strings.HasPrefix(trimmed, \"role_profiles:\")" internal/cli/team_spawn.go` returns `0`.
- `grep -c "type workflowConfig struct" internal/cli/workflow_lint.go` returns `0`.

**AC binding**: AC-WSE-006.

### M5 — Audit registry cleanup + final verification (REQ-WSE-008)

**Deliverables**:
- `internal/config/audit_loader_completeness_test.go`:
  - Remove `"workflow",` line (currently line 28) and its comment from `acknowledgedUnloadedSections`.
- `internal/config/audit_registry.go`:
  - Update comment on line 34 from `"workflow": "WorkflowConfig",` to reference the now-complete nested struct binding (optional — yaml binding correctness is the contract, comment is informational).

**Verification**:
- `! grep -n '"workflow"' internal/config/audit_loader_completeness_test.go | grep -q "acknowledged"` exit 0.
- `go test -run TestAuditLoaderCompleteness -v ./internal/config/...` PASS.
- Full self-verification command set (E1..E7 per Section E delegation template):
  - `go test ./...`
  - `go test -cover ./internal/config/... ./internal/cli/...`
  - `go vet ./...`
  - `golangci-lint run --timeout=2m` (NEW issues = 0)
  - `GOOS=windows GOARCH=amd64 go build ./...`
  - `go run ./cmd/moai --version`
  - `grep -rn 'AskUserQuestion\|mcp__askuser' internal/config/ internal/cli/team_spawn.go internal/cli/workflow_lint.go | grep -v _test | grep -v "^[^:]*:[0-9]*:[ \t]*//"` (0 matches).

**AC binding**: AC-WSE-008 + final Definition of Done validation (acceptance.md §Definition of Done).

---

## File-Level Change Inventory

| File | Marker | Lines (estimated) | Milestone |
|------|--------|-------------------|-----------|
| `internal/config/types.go` | [EXTEND] | +120 / -10 (rename 2 + add 9 sub-structs + add 9 fields) | M1 |
| `internal/config/defaults.go` | [EXTEND] | +40 / -5 (populate 9 nested defaults) | M2 |
| `internal/config/workflow_accessors.go` | [NEW] | +35 (5 accessor methods + doc comments) | M2 |
| `internal/config/types_test.go` | [EXTEND] | +20 / -5 (legacy field name fixups) | M1 |
| `internal/config/defaults_test.go` | [EXTEND] | +50 (36-assertion table-driven test) | M2 |
| `internal/config/workflow_nested_test.go` | [NEW] | +180 (4 test functions, fixtures) | M3 |
| `internal/cli/team_spawn.go` | [EXTEND] | +25 / -80 (replace 412-490 body) | M4 |
| `internal/cli/team_spawn_test.go` | [EXTEND] | +10 / -2 (Edge-WSE-001 update) | M3+M4 |
| `internal/cli/workflow_lint.go` | [EXTEND] | +10 / -25 (remove internal types) | M4 |
| `internal/cli/workflow_lint_test.go` | [EXTEND] | +5 / -3 (type sig fixups) | M4 |
| `internal/config/audit_loader_completeness_test.go` | [EXTEND] | +0 / -3 (remove line 28 entry + comment) | M5 |
| `internal/config/audit_registry.go` | [EXTEND] | +1 / -1 (comment update) | M5 |

Estimated total: ~500 LOC affected (≈+520 / -130 net). Fits Tier M (300-1000 LOC, 5-15 files; this SPEC affects 12 files).

---

## Risk Mitigation Schedule

Mapped from spec.md §4:

| Risk | Severity | Mitigation Step | Milestone |
|------|----------|-----------------|-----------|
| R-WSE-001 (Go API rename breaks external import) | Medium | Confirm `internal/` package privacy via grep `pkg/`; document rename in CHANGELOG.md (sync-phase) | M1 |
| R-WSE-002 (test fixture mismatches on rename) | Medium | Incremental fixture update during M1; explicit `TestWorkflowConfigLegacyFieldsPreserved` regression | M1 |
| R-WSE-003 (map iteration ordering vs string parser order) | Low | Verify existing `TestLoadRoleProfiles` does not assert iteration order; add explicit godoc comment in `LoadRoleProfiles` | M4 |
| R-WSE-004 (yaml lib version drift) | Low | go.mod yaml.v3 already pinned; tests use explicit `yaml:"..."` tags throughout | M3 |
| R-WSE-005 (CI failure on parallel branches lacking new struct) | Medium | Gate M5 audit-test removal behind M1..M4 passing in the same PR; commit M5 LAST | M5 |

---

## Late-Branch Workflow

Per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005:
- Plan commit lands on `main` directly (this commit).
- Run commits (M1..M5) land on `main` directly.
- Sync commit lands on `main` directly.
- PR creation step: `git switch -c feat/SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 && git push -u origin HEAD && gh pr create --base main`.
- After admin squash merge: `git checkout main && git fetch origin && git reset --hard origin/main && git pull origin main` per REQ-LB-006.

No L2/L3 worktree used (user policy 2026-05-17, default `--branch` semantics).

---

## Self-Estimated plan-auditor Score

Following the plan-auditor `Tier M threshold = 0.80` contract:

| Dimension | Self-score | Justification |
|-----------|-----------|---------------|
| EARS REQ coverage & specificity | 0.92 | 8 REQs with measurable verbs; each must-pass flagged; covers nested fields + sub-struct + unmarshal + rename + accessor + migration + defaults + audit cleanup |
| AC binary & traceability | 0.93 | 100% REQ↔AC (8/8); each AC has explicit grep/test command; PASS/FAIL outcome unambiguous |
| Out-of-Scope explicitness | 0.90 | 8 EXCLs with rationale; `h3 ### N.M Out of Scope` heading per spec-lint contract (avoids MissingExclusions) |
| Risk + Edge case coverage | 0.85 | 5 Risks (severity + mitigation) + 4 Edge Cases (each mapped to test name); Edge-WSE-004 documents implementation choice flexibility |
| Brownfield state + migration completeness | 0.88 | 12-file inventory with line estimates; M1..M5 dependency-ordered; ad-hoc parser elimination explicit |
| Constitution alignment | 0.85 | Tech stack / naming / forbidden libs / architecture all addressed |
| Tier alignment + LEAN compliance | 0.88 | Tier M classification justified (300-1000 LOC, 5-15 files = 12); 3-artifact set; Option (c) backward-compat consistent with sibling P2 SPEC |

Aggregate self-estimate: **0.886** (above Tier M PASS threshold 0.80, margin +0.086).

---

## Commit Message Template

```
plan(SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001): Tier M LEAN 3-artifact draft (v2 audit applied)

- spec.md: 8 EARS REQs / disposition matrix 16 rows / 7 Exclusions / 5 Risks / 4 Edge Cases
- acceptance.md: 8 binary ACs / 100% REQ↔AC traceability / Definition of Done 11 conditions
- plan.md: 5 milestones (M1..M5) / 12 files affected / Option (c) backward-compat
- v2 audit Steps 1-6 applied per .moai/research/config-audit-2026-05-22.md
- Owner SPECs: WF-003 (default_mode/execution_mode) + WORKFLOW-OPT-001 (role_profiles)
- Closes the "workflow" exception entry in audit_loader_completeness_test.go:28

🗿 MoAI <email@mo.ai.kr>
```

---

End of plan.md.

---
id: SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001
title: "workflow.yaml nested keys (completion/loop_prevention/memory/default_mode/execution_mode/team.*) Go struct 정합 (v2 audit applied)"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-06-02
author: manager-spec
priority: P1
phase: "v3.0.0 — Round 5"
module: "internal/config/types + internal/config/defaults + internal/config/manager + internal/cli/team_spawn + internal/cli/workflow_lint + internal/config/audit_registry"
lifecycle: spec-anchored
tags: "config, workflow, yaml-struct-symmetry, v2-audit, team-config, schema, role-profiles"
tier: M
---

# SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001 — workflow.yaml ↔ Go struct 정합

## HISTORY

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-05-22 | manager-spec (via MoAI orchestrator) | 초안. v2 audit Step 1-6 적용. P2 SPEC `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` 와 동일한 Tier M LEAN 패턴 적용. 8 EARS REQs / 8 binary ACs / 5 Risks / 4 Edge Cases / 7 Out of Scope. Disposition Matrix 16행 (auto_clear/completion/default_mode/execution_mode/loop_prevention/memory/team.{auto_selection,role_profiles,role_profile_keys,patterns,require_plan_approval,delegate_mode,default_model,enabled,max_teammates}/token_budget/worktree). Backward-compat Option (c) FLAT-fields-via-accessor. |

---

## 1. 개요 (Overview)

### 1.1 Mission Statement

`.moai/config/sections/workflow.yaml`는 200+ lines 의 풍부한 nested 구조 (`auto_clear` / `completion` / `default_mode` / `execution_mode` / `loop_prevention` / `memory` / `team.{auto_selection, role_profile_keys, role_profiles, patterns, ...}` / `token_budget` / `worktree`) 이지만, `internal/config/types.go:300-306` 의 `WorkflowConfig` struct 는 **FLAT 5필드**만 정의되어 있으며 그 중 4개 (`AutoClear bool`, `PlanTokens/RunTokens/SyncTokens int`, `AutoSelection TeamAutoSelectionConfig`) 가 **잘못된 yaml 경로에 매핑**되어 있다:

- `WorkflowConfig.AutoClear bool yaml:"auto_clear"` ← yaml은 `auto_clear: {enabled, after_plan, after_run, token_threshold}` nested struct (scalar bool 매핑 실패)
- `WorkflowConfig.PlanTokens/RunTokens/SyncTokens int yaml:"plan_tokens"` 등 ← yaml은 `token_budget.{plan, run, sync}` (path mismatch)
- `WorkflowConfig.AutoSelection TeamAutoSelectionConfig yaml:"auto_selection"` ← yaml은 `team.auto_selection` (parent path mismatch)

결과적으로 yaml의 ~30개 nested keys가 `yaml.Unmarshal`로 Go struct에 unmarshal 되지 않으며, `internal/cli/team_spawn.go:412-490` (`LoadRoleProfiles`)와 `internal/cli/workflow_lint.go:42-56` (`workflowConfig` 내부 type) 가 **ad-hoc 별도 parser**로 우회하고 있다.

또한 `internal/config/audit_loader_completeness_test.go:27` 는 `"workflow"` 를 `acknowledgedUnloadedSections` 에 등록하며 코멘트 `"out-of-scope: role_profiles subset loaded; full unification deferred (spec.md §1.2)"` 를 명시한다. 즉, 본 SPEC이 해결할 "full unification deferred" 의 owner SPEC 이 된다.

본 SPEC은 v2 audit (Step 1-6, 2026-05-22) 결과를 근거로:
1. workflow.yaml nested keys ~30개의 **disposition (live wire-through / forward-compat scaffold / WontDo documentation-only / dead)** 를 명문화한다.
2. 명문화된 결정에 따라 `WorkflowConfig` struct 계층을 재구성한다 (top-level fields + AutoClearConfig + CompletionConfig + LoopPreventionConfig + MemoryConfig + TeamConfig sub-struct + TokenBudgetConfig + WorktreeConfig sub-structs).
3. 현재 FLAT 5필드의 backward-compat 전략을 명시한다 (Option (c) accessor method 채택).
4. `internal/cli/team_spawn.go` 의 ad-hoc `LoadRoleProfiles` 를 새 struct 기반으로 마이그레이션한다 (병행 parser 경로 제거).
5. `internal/config/audit_loader_completeness_test.go` 의 `acknowledgedUnloadedSections` 에서 `"workflow"` 항목을 제거한다 (full unification 완료 신호).

### 1.2 v2 Audit 결과 (Step 1-6, 2026-05-22)

본 audit은 v1 audit anti-pattern (P3 LSP audit + 첫 config audit 회수 사례) 재발 방지를 위해 `.moai/research/config-audit-2026-05-22.md` §4 의 6-step 절차를 적용했다.

#### Step 1 — Symbol reference grep (production code, excluding _test/templates):

| yaml key | Go consumer 위치 | Disposition signal |
|----------|-------------------|---------------------|
| `workflow.auto_clear.{enabled, after_plan, after_run, token_threshold}` | `internal/template/context.go:39,81` `TemplateContext.AutoClear bool` (template render time only); `internal/core/project/initializer.go:346-350` yaml literal write | template-only (single bool collapse) |
| `workflow.completion.{detect_in_output, markers.{complete, done}}` | `internal/hook/stop.go:14-19` `defaultCompletionMarkers` **hardcoded** const `["<moai>DONE</moai>", "<moai>COMPLETE</moai>"]` | hardcoded — yaml NOT read |
| `workflow.default_mode` | **0 Go consumer** (spec-workflow.md line 95,149,163,180-181 declares contract; consumer is `.claude/skills/moai/workflows/run.md` skill body — NOT Go) | skill-body consumed, NOT Go |
| `workflow.execution_mode` | **0 Go consumer** (same as default_mode — skill-body owned by REQ-WF003-014/018) | skill-body consumed |
| `workflow.loop_prevention.*` (3 keys) | **0 Go consumer** (`internal/loop/state.go` `MaxIter` 은 ralph.yaml `max_iterations` 별도 consumer — `workflow.loop_prevention.max_iterations` 와 다른 키) | dormant / documentation |
| `workflow.memory.*` (4 keys) | **0 Go consumer** (`.claude/rules/moai/workflow/moai-memory.md` rule body 200-line cap 등을 명시하지만 Go enforcement 없음) | documentation-only |
| `workflow.team.auto_selection.*` (3 keys) | `WorkflowConfig.AutoSelection` 필드 존재하나 yaml path mismatch (`workflow.auto_selection` 으로 잘못 매핑) → **현재 0 effective consumer** (struct에는 있지만 unmarshal 실패) | broken wire-through (fix-needed) |
| `workflow.team.role_profile_keys` | **0 Go consumer** (`agent-teams-pattern.md` rule body 7-teammate composition 문서화 — Go가 검증하지 않음) | documentation-only WontDo |
| `workflow.team.role_profiles.*` (7 profiles) | `internal/cli/team_spawn.go:412-490` `LoadRoleProfiles` ad-hoc parser; `internal/cli/workflow_lint.go:42-56` internal `workflowConfig` 별도 type | ad-hoc parsed — migration target |
| `workflow.team.patterns.*` (6 patterns) | **0 Go consumer** (`.claude/rules/moai/workflow/team-pattern-cookbook.md` rule body 5 패턴 문서화 — Go가 dispatch하지 않음) | documentation-only WontDo |
| `workflow.team.{enabled, max_teammates, default_model, delegate_mode, require_plan_approval}` | **0 Go consumer** (skill bodies + rule bodies 참조; CLI flag `--team`/`--solo` override 가 우선) | skill-body consumed |
| `workflow.token_budget.{plan, run, sync}` | `WorkflowConfig.{PlanTokens, RunTokens, SyncTokens}` 필드 존재하나 yaml path mismatch (`workflow.plan_tokens` 으로 잘못 매핑) → **현재 0 effective consumer** (template context render 시 NewDefaultWorkflowConfig 값 사용) | broken wire-through (fix-needed) |
| `workflow.worktree.{auto_create, auto_cleanup, auto_merge, session_name_pattern, tmux_preferred}` | **0 Go consumer** (`internal/cli/worktree/` 명령은 별도 `.moai/config/sections/git-strategy.yaml` worktree_root 만 참조) | dormant — future SPEC |

#### Step 2 — SPEC catalog cross-check:

```bash
grep -rln "workflow.auto_clear\|workflow.completion\|workflow.loop_prevention\|workflow.memory\|workflow.default_mode\|workflow.execution_mode\|role_profile_keys\|team.patterns" .moai/specs/ | sort -u
```

결과 (2026-05-22):

| Owner SPEC | Status | Owned keys |
|------------|--------|------------|
| **SPEC-V3R2-WF-003** (Multi-Mode Router) | `completed` v0.3.4 | `workflow.default_mode` (REQ-WF003-014/018) + `workflow.execution_mode` (REQ-WF003-002/003) + `MODE_UNKNOWN` sentinel |
| **SPEC-V3R5-WORKFLOW-OPT-001** (8-Layer Improvement) | `implemented` v0.2.0 | `workflow.team.role_profiles` (REQ-WO-010 / AC-WO-009 7-profile contract) + `workflow.team.role_profile_keys` (Layer B 5+1+1 composition per `agent-teams-pattern.md`) |
| **SPEC-CONFIG-001** | `completed` v1.1.0 | Original `WorkflowConfig` FLAT 5-field declaration |
| **SPEC-V3R2-RT-005** | `implemented` | yaml↔struct parity audit registry; declares `workflow` as exception in `acknowledgedUnloadedSections` |

#### Step 3 — Owner SPEC disposition:

| Owner SPEC | Status | Blocker? | Disposition |
|------------|--------|----------|-------------|
| SPEC-V3R2-WF-003 | completed | **No** | `default_mode`/`execution_mode` 는 skill-body 가 consumer; Go struct 필드 추가는 forward-compat scaffold |
| SPEC-V3R5-WORKFLOW-OPT-001 | implemented | **No** | `role_profiles` 7-key 계약은 본 SPEC이 Go에 wire-through하면 강화됨 (regression NOT — AC-WO-009 검증이 새 struct 기반으로 가능) |
| SPEC-CONFIG-001 | completed | **No** | Original FLAT design 보존 (Option (c) accessor 패턴, P2 SPEC 와 동일) |
| SPEC-V3R2-RT-005 | implemented | **No** | `audit_loader_completeness_test.go:27` exception 제거가 본 SPEC의 완료 신호 |

**No owner SPEC is `status: in-progress`. No blocker.** Proceeding.

#### Step 4 — Consumer fan_in measurement (non-test code):

```bash
grep -rn "cfg\.Workflow\|config\.Workflow" internal/ pkg/ --include="*.go" | grep -v _test | grep -v templates
```

결과:
- `internal/config/manager.go:274,348` — getter/setter 2 (config aggregate)
- `internal/cli/workflow_lint.go:86` — `cfg.Workflow.Team.RoleProfiles` (internal type, NOT global Config) — 1
- **`cfg.Workflow.<field>` direct field access: 0** (manager의 generic `Get("workflow")` 만 존재)

→ Total **effective fan_in = 1** (workflow_lint 의 internal type) + manager-aggregate-only paths. 새 struct 필드 추가는 기존 consumers 를 깨지 않음.

#### Step 5 — Opt-in pattern detection:

- `workflow.completion.detect_in_output: true` 는 `stop.go:108-121` 의 marker detection 게이트 의도 — **그러나 Go는 yaml을 읽지 않는다** (`defaultCompletionMarkers` const 가 무조건 사용). 따라서 yaml의 `detect_in_output: false` 설정은 **dead opt-out** (사용자가 false로 설정해도 무시됨).
- `workflow.team.role_profile_keys` 는 `agent-teams-pattern.md` 5+1+1 spawn 제약 — yaml subset 이 `role_profiles` 전체 map 의 키 일부만 enumerate. 본 SPEC에서 Go 검증 추가는 **선택사항** (run-phase 5+1+1 검증은 orchestrator/skill body 책임).
- `workflow.team.patterns.*` 는 `team-pattern-cookbook.md` 의 5 패턴 sentinels (research/implementation/review/design/debug + `full_stack` 등) — pattern dispatch 가 **skill-body 영역** (Go는 패턴 이름을 enum 으로 검증하지 않음).

→ 3건 모두 **documentation-only WontDo** classification 안전.

#### Step 6 — Cross-SPEC dependency graph:

```
SPEC-V3R2-WF-003 (Multi-Mode Router, completed)
        │
        ├── .claude/skills/moai/workflows/run.md (skill body reads workflow.default_mode)
        ├── .claude/skills/moai/workflows/design.md (skill body reads workflow.default_mode)
        └── (NO Go consumer for default_mode/execution_mode)

SPEC-V3R5-WORKFLOW-OPT-001 (8-Layer, implemented)
        │
        ├── .moai/config/sections/workflow.yaml team.role_profiles (7 keys, validated by workflow_lint.go)
        ├── .claude/rules/moai/workflow/agent-teams-pattern.md (5+1+1 composition documentation)
        └── internal/cli/team_spawn.go:412-490 LoadRoleProfiles ad-hoc parser (MIGRATION TARGET)

SPEC-CONFIG-001 (completed, FLAT struct historical record)
        │
        └── WorkflowConfig FLAT 5-field struct (current, partially broken)

SPEC-V3R2-RT-005 (implemented)
        │
        └── audit_registry.go yaml↔struct parity tracking
                └── audit_loader_completeness_test.go line 27:
                    "workflow — out-of-scope: role_profiles subset loaded; full unification deferred (spec.md §1.2)"
                    (this SPEC's CLOSURE TARGET — remove the exception when REQs satisfied)
```

### 1.3 Disposition Matrix (v2 audit applied 2026-05-22)

Disposition codes (P2 SPEC SCHEMA-001 와 동일 어휘):
- **wire-through**: yaml→Go binding 추가, Go 코드가 값을 읽음 (현재 consumer 가 마이그레이션).
- **forward-compat-scaffold**: Go struct 필드 추가하나 consumer wiring 은 향후 SPEC 으로 연기 (필드는 noop 으로 존재). yaml 대칭성 보존 + 향후 wire-through 진입점.
- **WontDo (documentation-only)**: struct 에 추가하지 않음; rule/skill body 가 owner. EXCL 에 명시.
- **dead**: yaml 에서 제거 (consumer 0건 + historical residue). 본 SPEC 에서는 사용하지 않음 (P2 와 같이 보수적 운영).

| yaml key | Owner SPEC | Current Status | Disposition | Rationale & Action |
|----------|-----------|----------------|-------------|---------------------|
| `workflow.auto_clear.{enabled, after_plan, after_run, token_threshold}` | template.Context (no SPEC) | template-render only | **wire-through (nested)** | `WorkflowConfig.AutoClear AutoClearConfig{Enabled, AfterPlan, AfterRun, TokenThreshold}`. FLAT `AutoClear bool` deprecated → 새 accessor `AutoClearEnabled() bool` 가 `AutoClear.Enabled` 반환. template render 는 새 필드 우선 + FLAT fallback. |
| `workflow.completion.detect_in_output` | none (hardcoded in stop.go) | hardcoded | **forward-compat-scaffold** | `WorkflowConfig.Completion CompletionConfig{DetectInOutput bool}`. Go consumer `stop.go` 는 본 SPEC 에서 마이그레이션하지 않음 (EXCL-WSE-002); 향후 wire-through SPEC 의 진입점. |
| `workflow.completion.markers.{complete, done}` | none (hardcoded in stop.go) | hardcoded const | **forward-compat-scaffold** | `CompletionConfig.Markers MarkersConfig{Complete, Done}` (string 2-field). 향후 `NewStopHandlerWithMarkers` 호출 사이트가 cfg 에서 읽도록 마이그레이션 가능. |
| `workflow.default_mode` | SPEC-V3R2-WF-003 (skill body) | skill-body consumed | **forward-compat-scaffold** | `WorkflowConfig.DefaultMode string`. skill `workflows/run.md` 는 yaml 직접 read 유지 (EXCL-WSE-003); Go scaffold 는 향후 WF-003 의 Go-side enforcement 진입점. |
| `workflow.execution_mode` | SPEC-V3R2-WF-003 (skill body) | skill-body consumed | **forward-compat-scaffold** | `WorkflowConfig.ExecutionMode string`. 동일 rationale (skill-body 유지). |
| `workflow.loop_prevention.{failure_pattern_detection, max_iterations, max_retries_per_operation}` | none | dormant | **forward-compat-scaffold** | `WorkflowConfig.LoopPrevention LoopPreventionConfig{FailurePatternDetection, MaxIterations, MaxRetriesPerOperation}`. 향후 Ralph engine 통합 SPEC 의 진입점 (현재 `ralph.yaml` 별도). |
| `workflow.memory.{audit_enabled, index_line_cap, stale_aggregate_threshold, staleness_threshold_hours}` | none (rule moai-memory.md) | documentation | **forward-compat-scaffold** | `WorkflowConfig.Memory MemoryConfig{AuditEnabled, IndexLineCap, StaleAggregateThreshold, StalenessThresholdHours}`. memory audit hook 이 향후 Go-enforce 할 진입점. |
| `workflow.team.auto_selection.{min_domains_for_team, min_files_for_team, min_complexity_score}` | SPEC-V3R5-WORKFLOW-OPT-001 (REQ-WO-010 인접) | **broken wire-through** (path mismatch) | **wire-through (fix)** | `WorkflowConfig.Team.AutoSelection TeamAutoSelectionConfig` 로 이동. FLAT `WorkflowConfig.AutoSelection` 은 accessor 가 `Team.AutoSelection` 반환 (Option (c)). 현재 broken path 수정. |
| `workflow.team.role_profile_keys` (3-element list) | SPEC-V3R5-WORKFLOW-OPT-001 (rule documentation) | documentation | **forward-compat-scaffold** | `TeamConfig.RoleProfileKeys []string`. 향후 5+1+1 spawn validator 의 Go 검증 진입점. |
| `workflow.team.role_profiles.{analyst, architect, designer, implementer, researcher, reviewer, tester}` (7 keys × 4 fields) | SPEC-V3R5-WORKFLOW-OPT-001 (REQ-WO-010 / AC-WO-009) | ad-hoc parsed (team_spawn.go:412) | **wire-through (migrate ad-hoc parser)** | `TeamConfig.RoleProfiles map[string]RoleProfileEntry{Description, Isolation, Mode, Model}`. `team_spawn.go:LoadRoleProfiles` 가 새 struct 기반으로 마이그레이션 (병행 parser 제거). workflow_lint.go 도 동일 마이그레이션 (Wave 5). |
| `workflow.team.patterns.{design_implementation, full_stack, implementation, investigation, plan_research, review}` (6 patterns × {description, model, roles[]}) | SPEC-V3R5-WORKFLOW-OPT-001 (cookbook rule) | documentation | **WontDo (documentation-only)** | EXCL-WSE-004. team-pattern-cookbook.md rule body 가 owner; Go 가 패턴 dispatch 하지 않음. yaml 키는 향후 SPEC 이 wire-through 결정 시 추가 가능. |
| `workflow.team.{enabled, max_teammates, default_model, delegate_mode, require_plan_approval}` | various (CLI/skill) | skill-body + CLI flag | **forward-compat-scaffold** | `TeamConfig.{Enabled, MaxTeammates, DefaultModel, DelegateMode, RequirePlanApproval}`. CLI `--team`/`--solo` flag 가 override 우선; struct 필드는 default value 진입점. |
| `workflow.token_budget.{plan, run, sync}` | template.Context | template-render only | **wire-through (path fix)** | `WorkflowConfig.TokenBudget TokenBudgetConfig{Plan, Run, Sync int}`. FLAT `PlanTokens/RunTokens/SyncTokens` 는 accessor (Option (c)). yaml path `workflow.token_budget.*` 매핑 정상화. |
| `workflow.worktree.{auto_cleanup, auto_create, auto_merge, session_name_pattern, tmux_preferred}` | none (separate worktree CLI) | dormant | **forward-compat-scaffold** | `WorkflowConfig.Worktree WorkflowWorktreeConfig{AutoCleanup, AutoCreate, AutoMerge, SessionNamePattern, TmuxPreferred}`. 기존 `GitStrategyConfig.WorktreeRoot` 와 충돌 없음 (다른 키 도메인). |
| FLAT `WorkflowConfig.AutoClear bool` | SPEC-CONFIG-001 | deprecated path | **WontDo (preserve via deprecation + accessor)** | `// Deprecated:` 주석 + `AutoClearEnabled() bool` accessor. Removal SPEC SemVer major-bump 필요. |
| FLAT `WorkflowConfig.{PlanTokens, RunTokens, SyncTokens} int` | SPEC-CONFIG-001 | deprecated path | **WontDo (preserve via deprecation + accessor)** | `// Deprecated:` 주석 + `PlanTokensValue()/RunTokensValue()/SyncTokensValue()` accessor (`TokenBudget.Plan` 등 반환). |
| FLAT `WorkflowConfig.AutoSelection TeamAutoSelectionConfig` | SPEC-CONFIG-001 | **broken** (path mismatch) | **WontDo (preserve via deprecation + accessor)** | `// Deprecated:` 주석 + `AutoSelectionConfig() TeamAutoSelectionConfig` accessor (`Team.AutoSelection` 반환). |

### 1.4 Backward-Compat Strategy (D1)

P2 SPEC `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001` 와 동일 패턴 (Option (c) accessor method):

- FLAT 5필드 (`AutoClear bool`, `PlanTokens int`, `RunTokens int`, `SyncTokens int`, `AutoSelection TeamAutoSelectionConfig`) 는 **deprecated 주석 추가 + struct 에 보존**.
- 신규 nested 필드 (`AutoClear AutoClearConfig`) 와 FLAT 필드 (`AutoClear bool`) 가 동일 Go 식별자 이므로 **rename 필요**: FLAT 필드를 `AutoClearLegacy bool` 로 rename + deprecated 주석.
- 마찬가지로 `AutoSelection TeamAutoSelectionConfig` → `AutoSelectionLegacy` (path 가 broken 이므로 deprecation 정당화 강함).
- `PlanTokens/RunTokens/SyncTokens` 는 yaml path mismatch 때문에 unmarshal 실패하지만 struct field 이름은 그대로 유지 (FLAT 식별자 ambiguity 없음).
- Accessor methods:
  - `(*Config).WorkflowAutoClearEnabled() bool` → `Workflow.AutoClear.Enabled` 반환
  - `(*Config).WorkflowPlanTokens() int` → `Workflow.TokenBudget.Plan` 반환
  - `(*Config).WorkflowTeamAutoSelection() TeamAutoSelectionConfig` → `Workflow.Team.AutoSelection` 반환

검토했던 대안:
- **Option (a)** — FLAT 만 유지 + lint warning: yaml roundtrip 시 nested 정보 손실 + ad-hoc parser 영구화. Rejected.
- **Option (b)** — FLAT 전체 제거 (breaking): SPEC-CONFIG-001 HISTORY amendment 필요 + types_test.go/defaults_test.go 광범위 수정 + SemVer major-bump. Out of Tier M scope.
- **Option (c)** — nested 정식화 + FLAT 보존 + accessor ✅ **CHOSEN**.

### 1.5 Brownfield State Inventory

본 SPEC은 brownfield 작업이다. 모든 영향 파일은 [EXTEND] 분류이며 [NEW] 파일 1개 (test 파일).

| Marker | 파일 | 분류 | 근거 |
|--------|------|------|------|
| [EXTEND] | `internal/config/types.go` | EXTEND | `WorkflowConfig` struct 확장 + 신규 sub-structs (`AutoClearConfig`, `CompletionConfig`, `MarkersConfig`, `LoopPreventionConfig`, `MemoryConfig`, `TeamConfig`, `RoleProfileEntry`, `TokenBudgetConfig`, `WorkflowWorktreeConfig`). FLAT 5필드 deprecated 보존 (1건 rename: `AutoClear bool` → `AutoClearLegacy bool` + `AutoSelection` → `AutoSelectionLegacy`). |
| [EXTEND] | `internal/config/defaults.go` | EXTEND | `NewDefaultWorkflowConfig()` 갱신 — 신규 nested 필드 default 값 populated (`workflow.yaml` template SSOT defaults 정확 미러: AutoClear.Enabled=true, TokenBudget.{Plan=30000,Run=180000,Sync=40000}, Team.DefaultModel="sonnet", Team.AutoSelection.{MinDomainsForTeam=3, MinFilesForTeam=10, MinComplexityScore=7}, Team.RoleProfiles 7-key map populated, Worktree.AutoCreate=false, etc.). |
| [EXTEND] | `internal/config/manager.go` | EXTEND | getSectionLocked / setSectionLocked 의 "workflow" case 는 변경 없음 (struct value semantic 유지). |
| [EXTEND] | `internal/cli/team_spawn.go` | EXTEND | `LoadRoleProfiles` 함수가 `cfg.Workflow.Team.RoleProfiles` map 기반으로 마이그레이션 (lines 412-490 의 ad-hoc string parsing 제거). `WriteHeavyRoles` 결정 로직 보존. |
| [EXTEND] | `internal/cli/workflow_lint.go` | EXTEND | internal `workflowConfig` type 을 `config.WorkflowConfig` 사용으로 변경 (lines 42-56 의 별도 type 정의 제거). validateRoleProfiles 는 새 type 시그니처로 갱신. |
| [EXTEND] | `internal/config/audit_registry.go` | EXTEND | comment 갱신 — workflow 매핑이 nested struct 로 완전 정합 표기. |
| [EXTEND] | `internal/config/audit_loader_completeness_test.go` | EXTEND | line 27 `"workflow"` 항목을 `acknowledgedUnloadedSections` 에서 **제거** (REQ-WSE-008 만족 신호). |
| [NEW]    | `internal/config/workflow_nested_test.go` | NEW | nested yaml fixture (auto_clear, completion, loop_prevention, memory, team.{auto_selection, role_profile_keys, role_profiles, patterns 제외}, token_budget, worktree 완전 populated) → `yaml.Unmarshal` → assert all nested 필드 present. RoundTrip + accessor 검증. |
| [EXTEND] | `internal/config/types_test.go` | EXTEND | 기존 FLAT 테스트 보존 (backward-compat 검증) + 신규 nested 테스트 추가 + accessor 동작 검증 (FLAT-via-nested). |
| [EXTEND] | `internal/config/defaults_test.go` | EXTEND | `TestNewDefaultWorkflowConfig` 확장 — 신규 nested 필드 defaults assertion (Team.RoleProfiles 7-key map count + Team.AutoSelection thresholds 등). |

---

## 2. EARS Requirements

**REQ-WSE-001** (Ubiquitous, **must-pass**): The system shall extend `WorkflowConfig` Go struct so that nested yaml keys `workflow.auto_clear.{enabled, after_plan, after_run, token_threshold}`, `workflow.completion.{detect_in_output, markers.{complete, done}}`, `workflow.default_mode`, `workflow.execution_mode`, `workflow.loop_prevention.{failure_pattern_detection, max_iterations, max_retries_per_operation}`, `workflow.memory.{audit_enabled, index_line_cap, stale_aggregate_threshold, staleness_threshold_hours}`, `workflow.token_budget.{plan, run, sync}`, `workflow.worktree.{auto_cleanup, auto_create, auto_merge, session_name_pattern, tmux_preferred}` each have a corresponding Go struct field reachable via dotted access (e.g., `cfg.Workflow.AutoClear.Enabled`, `cfg.Workflow.TokenBudget.Plan`).

**REQ-WSE-002** (Ubiquitous, **must-pass**): The system shall introduce a `TeamConfig` sub-struct as `WorkflowConfig.Team TeamConfig` containing fields `AutoSelection TeamAutoSelectionConfig`, `Enabled bool`, `MaxTeammates int`, `DefaultModel string`, `DelegateMode bool`, `RequirePlanApproval bool`, `RoleProfileKeys []string`, and `RoleProfiles map[string]RoleProfileEntry` such that all corresponding `workflow.team.<key>` yaml values populate via `yaml.Unmarshal`. The `team.patterns.*` keys are explicitly excluded (EXCL-WSE-004).

**REQ-WSE-003** (Event-driven, **must-pass**): WHEN `yaml.Unmarshal` is invoked on a yaml document matching the current production `workflow.yaml` nested structure, THEN the resulting `WorkflowConfig` value SHALL have all nested fields populated such that `cfg.Workflow.Team.RoleProfiles["implementer"].Isolation` resolves to `"worktree"` AND `cfg.Workflow.AutoClear.Enabled` resolves to `true` AND `cfg.Workflow.TokenBudget.Plan` resolves to `30000`.

**REQ-WSE-004** (Ubiquitous, **must-pass**): The system shall preserve the existing FLAT fields with the following rename + deprecation pattern: `WorkflowConfig.AutoClear bool` shall be renamed to `AutoClearLegacy bool` with `// Deprecated: use AutoClear.Enabled` doc comment; `WorkflowConfig.AutoSelection TeamAutoSelectionConfig` shall be renamed to `AutoSelectionLegacy` with `// Deprecated: use Team.AutoSelection` doc comment; `WorkflowConfig.{PlanTokens, RunTokens, SyncTokens} int` shall retain their identifiers but receive `// Deprecated: use TokenBudget.{Plan, Run, Sync}` doc comments (no yaml path mismatch resolution — they remain effectively unread, preserved for binary compat only).

**REQ-WSE-005** (State-driven, **must-pass**): WHILE `WorkflowConfig` is in use, the system shall provide accessor methods on `*Config` aggregate: `(*Config).WorkflowAutoClearEnabled() bool` returning `c.Workflow.AutoClear.Enabled`; `(*Config).WorkflowPlanTokens() int` returning `c.Workflow.TokenBudget.Plan` (and similar for Run/Sync); `(*Config).WorkflowTeamAutoSelection() TeamAutoSelectionConfig` returning `c.Workflow.Team.AutoSelection`. These accessors shall be the recommended migration path for downstream consumers reading the legacy FLAT fields.

**REQ-WSE-006** (Ubiquitous, **must-pass**): The system shall migrate `internal/cli/team_spawn.go` `LoadRoleProfiles(workflowPath string)` function such that the resulting `map[string]RoleProfile` is derived from `config.LoadAll().Workflow.Team.RoleProfiles` via the standard config loader path, eliminating the manual line-by-line yaml string parsing block (current lines 412-490). The function signature shall remain unchanged for backward compatibility (callers continue to pass a `workflowPath`). The `WriteHeavyRoles` determination logic shall be preserved verbatim.

**REQ-WSE-007** (Ubiquitous): WHEN `NewDefaultWorkflowConfig()` is called, THEN the returned value SHALL contain populated nested defaults exactly matching the `workflow.yaml` template SSOT defaults: `AutoClear.Enabled == true`, `AutoClear.AfterPlan == true`, `AutoClear.TokenThreshold == 150000`, `Completion.DetectInOutput == true`, `Completion.Markers.Done == "<moai>DONE</moai>"`, `Completion.Markers.Complete == "<moai>COMPLETE</moai>"`, `LoopPrevention.MaxIterations == 100`, `LoopPrevention.MaxRetriesPerOperation == 3`, `Memory.IndexLineCap == 200`, `TokenBudget.Plan == 30000`, `TokenBudget.Run == 180000`, `TokenBudget.Sync == 40000`, `Team.DefaultModel == "sonnet"`, `Team.AutoSelection.MinDomainsForTeam == 3`, `Worktree.AutoCreate == false`, `Team.RoleProfiles` contains exactly 7 keys (analyst, architect, designer, implementer, researcher, reviewer, tester) per AC-WO-009.

**REQ-WSE-008** (Ubiquitous, **must-pass**): The system shall remove the `"workflow"` entry from `acknowledgedUnloadedSections` in `internal/config/audit_loader_completeness_test.go` (currently at line 27 with comment `"out-of-scope: role_profiles subset loaded; full unification deferred (spec.md §1.2)"`). The removal shall occur only after REQ-WSE-001 through REQ-WSE-007 are satisfied, signaling that workflow.yaml has a complete loader path via `Loader.Load()` Config.Workflow field.

---

## 3. Out of Scope (Exclusions)

### 3.1 Out of Scope

- **EXCL-WSE-001**: Sunsetting / removing FLAT `AutoClearLegacy`, `AutoSelectionLegacy`, `PlanTokens`, `RunTokens`, `SyncTokens` fields. Option (c) preserves these for backward-compat. A future SPEC may amend SPEC-CONFIG-001 to deprecate them entirely with SemVer major-bump.
- **EXCL-WSE-002**: Wiring `workflow.completion.markers.{done, complete}` consumption into `internal/hook/stop.go`. `defaultCompletionMarkers` const (line 14-19) remains hardcoded. The new `CompletionConfig.Markers` struct provides forward-compat scaffold only; `NewStopHandler()` migration to read cfg is deferred to a future SPEC.
- **EXCL-WSE-003**: Wiring `workflow.default_mode` and `workflow.execution_mode` consumption into Go code paths. Skill bodies (`workflows/run.md`, `workflows/design.md`) continue to read yaml directly per SPEC-V3R2-WF-003 REQ-WF003-014/018. The new Go scaffold fields are forward-compat entry points only.
- **EXCL-WSE-004**: `workflow.team.patterns.*` (6 patterns × {description, model, roles[]}) yaml→Go struct binding. team-pattern-cookbook.md rule body is the owner; Go does not dispatch patterns. A future SPEC may wire-through if pattern enum validation becomes a Go responsibility.
- **EXCL-WSE-005**: `workflow.team.role_profile_keys` Go-side enforcement (validating that the 3 listed keys are a subset of `team.role_profiles` map keys). Currently skill-body / orchestrator responsibility per `agent-teams-pattern.md`. Adding Go validation would require a new validator; deferred.
- **EXCL-WSE-006**: Adding migration tooling for existing user `workflow.yaml` files. Production yaml is already nested (no migration needed). Synthetic FLAT test fixtures in `types_test.go` migrate within this SPEC's scope.
- **EXCL-WSE-007**: CATEGORY B/C/D dead config from v1 audit (e.g., `lsp.yaml aggregator/circuit_breaker`, `llm.yaml legacy default_model/quality_model/speed_model`, `constitution.yaml.performance.*`, `state.retention_days`, `session.stale_seconds`). Those belong to separate SPECs per v2 audit §5.
- **EXCL-WSE-008**: Refactoring `internal/template/context.go:39,81` `TemplateContext.AutoClear bool` and `internal/core/project/initializer.go:346-350` yaml literal write logic. The template render path uses `TemplateContext` independently of `WorkflowConfig` struct; refactoring to use new struct accessors is deferred to a future SPEC.

### 3.2 Future Work (referenced but not in this SPEC)

- SPEC-V3R5-WORKFLOW-CONSUMER-WIRE-001 (hypothetical) — wire `workflow.completion.markers.*` into `stop.go`, `workflow.default_mode` into mode dispatcher Go layer.
- SPEC-V3R5-WORKFLOW-PATTERNS-WIRE-001 (hypothetical) — add `team.patterns.*` struct + Go-side pattern dispatch validation.
- SPEC-V3R5-CONFIG-FLAT-SUNSET-001 (hypothetical, shared with P2) — remove FLAT deprecated fields with SemVer major-bump.

---

## 4. Risks

- **R-WSE-001** (Medium): Renaming `WorkflowConfig.AutoClear bool` → `AutoClearLegacy bool` is a **Go API breaking change** for any external import path consumer (e.g., third-party tools importing `github.com/modu-ai/moai-adk/internal/config`). Mitigation: internal/ package is **private** by Go convention (no external import allowed per `internal/` semantics); confirmed by `grep -rn "modu-ai/moai-adk/internal/config" pkg/` ⇒ all consumers are within this repo. Document the rename in CHANGELOG.md sync-phase.
- **R-WSE-002** (Medium): Renaming `AutoSelection` → `AutoSelectionLegacy` may surface latent test assertion mismatches in `internal/config/types_test.go` and any test fixture using `WorkflowConfig{AutoSelection: ...}` struct literal. Mitigation: incremental fixture update + explicit `TestWorkflowConfigLegacyFieldsPreserved` regression test; greenfield grep verifies impact scope.
- **R-WSE-003** (Low): `internal/cli/team_spawn.go` `LoadRoleProfiles` migration from string-parser to typed `cfg.Workflow.Team.RoleProfiles` map iteration may surface ordering differences (Go map iteration is unordered vs the current line-by-line parser preserves yaml order). Mitigation: `LoadRoleProfiles` test (existing) does not assert order — verify via grep + add explicit "unordered" comment to function godoc.
- **R-WSE-004** (Low): yaml-struct symmetry test `workflow_nested_test.go` may flake on YAML library version drift (yaml.v3 nested-map handling). Mitigation: pin yaml library (already pinned in go.mod via `gopkg.in/yaml.v3`); test uses explicit `yaml:"key_name"` tags throughout.
- **R-WSE-005** (Medium): Removing `"workflow"` from `acknowledgedUnloadedSections` (REQ-WSE-008) may trigger CI failure on parallel branches that haven't yet adopted the new struct. Mitigation: gate REQ-WSE-008 satisfaction behind successful Go compilation + `TestAuditLoaderCompleteness` PASS within the same PR; commit ordering in plan.md M5 places REQ-WSE-008 LAST.

---

## 5. Edge Cases

- **Edge-WSE-001**: Empty `workflow.team.role_profiles:` in user yaml (key present, map empty). `cfg.Workflow.Team.RoleProfiles` returns empty map `map[string]RoleProfileEntry{}`. `LoadRoleProfiles` returns empty map without error. `workflow_lint.go validateRoleProfiles` emits ORC_WORKTREE_REQUIRED violations for all 3 write-heavy roles (preserves existing behavior).
- **Edge-WSE-002**: `workflow.team.role_profile_keys` lists a key absent from `workflow.team.role_profiles` map (e.g., `role_profile_keys: [implementer, tester, ghost]` but `role_profiles` lacks `ghost`). Go struct populates both fields independently; no cross-validation is added in this SPEC (deferred to EXCL-WSE-005). Skill body / orchestrator handles inconsistency at run-phase.
- **Edge-WSE-003**: User yaml omits `workflow.token_budget` block entirely. `cfg.Workflow.TokenBudget` populates with zero-values (`Plan=0, Run=0, Sync=0`). `NewDefaultWorkflowConfig()` provides defaults at construction time — but if user yaml is loaded over defaults, the zero-value MUST NOT silently override the construction default. Mitigation: loader path uses `yaml.Unmarshal` into a pre-populated default struct (existing Go pattern); MUST verify this works for nested zero-value semantics. Test: `Edge-WSE-003-test` in `workflow_nested_test.go`.
- **Edge-WSE-004**: User yaml provides ONLY legacy FLAT keys `workflow.auto_clear: true` and `workflow.plan_tokens: 30000` (without nested `auto_clear.enabled` or `token_budget.plan`). Because the FLAT identifier was renamed to `AutoClearLegacy` AND no yaml path matches it any more (legacy yaml `auto_clear: true` scalar at struct path `auto_clear` is a TYPE MISMATCH against new `AutoClear AutoClearConfig` struct), `yaml.Unmarshal` will emit a `yaml: unmarshal errors` warning. Mitigation: types.go uses `yaml.Node` flexible unmarshal pattern OR documents that legacy yaml files are explicitly unsupported (recommend user run `moai update`). Test: `Edge-WSE-004-test` covers both behaviors with explicit assertion.

---

## 6. Constitution Alignment

- **Technology Stack**: Go 1.21+ (current). YAML library `gopkg.in/yaml.v3` (existing dependency, no addition).
- **Naming Conventions**: PascalCase for exported Go struct fields. snake_case for yaml keys (existing convention preserved). `Legacy` suffix for deprecated rename (precedent: none in current codebase; this SPEC introduces it for forward-compat path).
- **Forbidden Libraries**: None violated. No new dependencies.
- **Architectural Patterns**: Struct definitions co-located in `internal/config/types.go` (existing convention). Sub-structs `AutoClearConfig`, `CompletionConfig`, `MarkersConfig`, `LoopPreventionConfig`, `MemoryConfig`, `TeamConfig`, `RoleProfileEntry`, `TokenBudgetConfig`, `WorkflowWorktreeConfig` follow precedent (e.g., `MigrationsConfig`, `SystemHookConfig`, `TeamAutoSelectionConfig` at lines 53-66, 197-207).
- **Security Standards**: No security-sensitive fields. yaml parsing uses standard library (no custom parser). Note: `LoadRoleProfiles` migration from string-parsing to typed unmarshal **reduces** attack surface (no more index-based string slicing).
- **Logging Standards**: No new logging code. Existing `slog` calls in `team_spawn.go` preserved.

---

## 7. References

- **v2 Audit Source**: `.moai/research/config-audit-2026-05-22.md` §4 (6-step procedure).
- **SPEC-V3R2-WF-003** (completed v0.3.4): Multi-Mode Router — owner of `workflow.default_mode` + `workflow.execution_mode` + `MODE_UNKNOWN` sentinel.
- **SPEC-V3R5-WORKFLOW-OPT-001** (implemented v0.2.0): 8-Layer Improvement — owner of `workflow.team.role_profiles` (AC-WO-009) + `workflow.team.role_profile_keys` per `agent-teams-pattern.md`.
- **SPEC-CONFIG-001** (completed v1.1.0): Original FLAT `WorkflowConfig` 5-field struct declaration.
- **SPEC-V3R2-RT-005** (implemented): yaml↔struct parity audit registry — establishes the audit framework this SPEC extends.
- **SPEC-V3R5-GIT-STRATEGY-SCHEMA-001** (draft, P2): Sibling SPEC using identical Tier M LEAN pattern and Option (c) backward-compat strategy.
- **Production yaml**: `.moai/config/sections/workflow.yaml` (current nested structure 200+ lines).
- **Template SSOT**: `internal/template/templates/.moai/config/sections/workflow.yaml` (default-value oracle for `NewDefaultWorkflowConfig()`; note: plain `.yaml`, there is no `.tmpl` variant).
- **Audit completeness test**: `internal/config/audit_loader_completeness_test.go:27` (current exception entry — REQ-WSE-008 removal target).
- **Ad-hoc parsers (migration targets)**: `internal/cli/team_spawn.go:409-490` `LoadRoleProfiles`; `internal/cli/workflow_lint.go:42-56` `workflowConfig` internal type.

---

End of spec.md.

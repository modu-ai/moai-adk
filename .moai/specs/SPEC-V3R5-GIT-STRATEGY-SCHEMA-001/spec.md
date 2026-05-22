---
id: SPEC-V3R5-GIT-STRATEGY-SCHEMA-001
title: "git-strategy.yaml nested mode-based 키들의 Go struct 정합 (v2 audit applied)"
version: "0.1.0"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0 — Round 5"
module: "internal/config/types + internal/config/defaults + internal/config/validation + internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl"
lifecycle: spec-anchored
tags: "config, git-strategy, yaml-struct-symmetry, v2-audit, late-branch, schema"
tier: M
---

# SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 — git-strategy.yaml ↔ Go struct 정합

## HISTORY

| Version | Date | Author | Notes |
|---------|------|--------|-------|
| 0.1.0 | 2026-05-22 | manager-spec (via MoAI orchestrator) | 초안. v2 audit Step 1-6 적용 결과 첫번째 Tier M draft. 8 EARS REQs / 9 binary ACs / 5 Risks / 5 Edge Cases / 6 Out of Scope 분류 항목. Disposition Matrix 12행 (mode/provider/github_username + gitlab + manual/personal/team 3 mode profiles × {automation, branch_creation, commit_style, hooks, environment, github_integration, push_to_remote, workflow, auto_checkpoint, branch_prefix, main_branch, draft_pr, required_reviews, branch_protection}). Backward-compat Option (c) accessor-method 채택 (D1). |

---

## 1. 개요 (Overview)

### 1.1 Mission Statement

`.moai/config/sections/git-strategy.yaml`는 mode-based nested 구조 (~70 keys across manual/personal/team profiles + top-level mode/provider/github_username + gitlab.instance_url)이지만, `internal/config/types.go:41-48` `GitStrategyConfig` struct는 **FLAT 6필드** (AutoBranch, BranchPrefix, CommitStyle, WorktreeRoot, Provider, GitLabInstanceURL)만 정의되어 있다. 결과적으로 nested yaml 키 ~70개가 `yaml.Unmarshal`로 Go struct에 unmarshal 되지 않는다.

본 SPEC은 v2 audit (Step 1-6, 2026-05-22) 결과를 근거로:
1. nested yaml 키 ~70개의 **disposition (live wire-through / forward-compat scaffold / WontDo / dead)** 을 명문화한다.
2. 명문화된 결정에 따라 `GitStrategyConfig` struct 계층을 재구성한다 (top-level + ModeProfile sub-struct + Automation/BranchCreation/CommitStyle/Hooks sub-structs).
3. 현재 FLAT 6필드의 backward-compat 전략을 명시한다 (Option (c) accessor method 채택).
4. yaml audit registry / template mirror / 신규 tests 갱신을 통해 향후 schema drift를 예방한다.

### 1.2 v2 Audit 결과 (Step 1-6, 2026-05-22)

본 audit은 v1 audit anti-pattern (P3 LSP audit) 재발 방지를 위해 6-step 절차를 적용했다.

#### Step 1 — Symbol reference grep (production code, excluding _test/templates):

| Field | Production consumers | Disposition signal |
|-------|----------------------|--------------------|
| `GitStrategyConfig.AutoBranch` | **0** | dead field |
| `GitStrategyConfig.BranchPrefix` | 1 (`validation.go:216` checkStringField only) | string-validation only |
| `GitStrategyConfig.CommitStyle` | 1 (`validation.go:217` checkStringField only) | string-validation only |
| `GitStrategyConfig.WorktreeRoot` | **0** | dead field |
| `GitStrategyConfig.Provider` | **0** | dead field (consumed via template.Context separate path) |
| `GitStrategyConfig.GitLabInstanceURL` | **0** in struct path | dead field (consumed via template.Context separate path) |

#### Step 2 — SPEC catalog cross-check:

Owner SPECs:
- **SPEC-CONFIG-001** (status: completed) — declares the original FLAT 4-field struct (AutoBranch/BranchPrefix/CommitStyle/WorktreeRoot, spec.md:414).
- **SPEC-GIT-001** (status: completed) — references SPEC-CONFIG-001 GitStrategyConfig.
- **SPEC-V3R5-LATE-BRANCH-001** (status: completed) — declares `team.branch_creation.auto_enabled` and `prompt_always` as **skill-body consumed** (REQ-LB-004: "the skill shall read `team.branch_creation.auto_enabled` from `git-strategy.yaml`"), NOT Go struct consumed. Verified by `grep "git-strategy" .claude/skills/moai/workflows/plan/spec-assembly.md` (line 302 reads yaml directly).

#### Step 3 — Owner SPEC disposition:

| Owner SPEC | Status | Blocker? | Disposition |
|------------|--------|----------|-------------|
| SPEC-CONFIG-001 | completed | No | Original FLAT design preserved by historical record; struct evolution allowed |
| SPEC-GIT-001 | completed | No | References SPEC-CONFIG-001; no in-progress |
| SPEC-V3R5-LATE-BRANCH-001 | completed | No | Nested keys consumed by skill body, not Go code; Go binding addition is additive |

No owner SPEC is `status: in-progress`. No blocker. Proceeding.

#### Step 4 — Consumer fan_in vs hardcoded fallback:

- `internal/config/defaults.go:225-226` hardcodes `DefaultBranchPrefix = "moai/"` and `DefaultCommitStyle = "conventional"` — these are FLAT fallback values.
- yaml `personal.branch_prefix` and `team.branch_prefix` both default to `"feature/SPEC-"` in template.
- Mismatch: hardcoded default (`"moai/"`) ≠ yaml default (`"feature/SPEC-"`). Wire-through would surface this.

- `internal/cli/update.go:2097-2148` mutates yaml via raw `map[string]any` paths — only `mode`, `provider`, `gitlab.instance_url` are touched. nested manual/personal/team are template-only.
- `internal/core/project/initializer.go:356-365` writes minimal yaml content (`mode`, `provider`, `github_username`, `gitlab.instance_url`) and relies on `git-strategy.yaml.tmpl` for the nested mode profiles via template rendering.

#### Step 5 — Opt-in pattern detection (Late-branch):

`team.branch_creation.auto_enabled: false` + `prompt_always: true` form the **Late-branch opt-in default pair** per SPEC-V3R5-LATE-BRANCH-001 (D1 Option a). The pair is consumed by `.claude/skills/moai/workflows/plan/spec-assembly.md` Phase 3 (line 302) which reads the yaml directly. Go code has ZERO consumers of these keys (verified by `grep -rn 'auto_enabled\|prompt_always\|Late-Branch' internal pkg`).

Implication: **wire-through is desirable but not currently blocking any production workflow**. The skill body already consumes the yaml correctly.

#### Step 6 — Cross-SPEC dependency graph:

```
SPEC-V3R5-LATE-BRANCH-001 (skill-body yaml consumer)
        │
        ├── .claude/skills/moai/workflows/plan/spec-assembly.md (reads yaml @ line 302)
        ├── .claude/agents/moai/manager-git.md (reads yaml @ line 43, 108, 114)
        └── (NO Go code consumer)

SPEC-CONFIG-001 (completed, FLAT struct historical record)
        │
        └── GitStrategyConfig FLAT 6-field struct (current)

SPEC-V3R2-RT-005 (completed)
        │
        └── audit_registry.go yaml↔struct parity tracking
                └── audit_loader_completeness_test.go line 19:
                    "git-strategy — out-of-scope: loaded via git-strategy.yaml.tmpl template rendering path"
                    (recognized tech debt, documented exception)
```

### 1.3 Disposition Matrix (v2 audit applied 2026-05-22)

Disposition codes:
- **wire-through**: yaml→Go binding added, Go code reads the value (current consumer migrates).
- **forward-compat-scaffold**: Go struct field added but consumer wiring deferred (future SPEC adds reader). Justified when adding binding now preserves yaml symmetry without breaking workflows.
- **WontDo**: not added to struct; remains skill-body / template-only consumed. Documented in EXCL.
- **dead**: remove from yaml (no consumer, historical residue).

| Yaml key | Owner SPEC | Current Status | Disposition | Rationale & Action |
|----------|-----------|----------------|-------------|---------------------|
| `git_strategy.mode` | template.Context (no SPEC) | template-context consumed | **wire-through** | Top-level mode selector. Add `Mode string` to struct. ActiveModeProfile() accessor uses it. |
| `git_strategy.provider` | template.Context | template-context consumed | **wire-through** | Already FLAT-struct-defined; preserved. |
| `git_strategy.github_username` | template.Context | template-context consumed | **wire-through** | Add `GitHubUsername string` to struct. |
| `git_strategy.gitlab.instance_url` | template.Context | template-context consumed | **wire-through (sub-struct)** | Migrate FLAT `GitLabInstanceURL` to nested `GitLab.InstanceURL`. Preserve FLAT field via accessor (Option c). |
| `git_strategy.{manual,personal,team}.workflow` | none (skill body reads) | yaml-only | **forward-compat-scaffold** | `ModeProfile.Workflow string`. Future Late-Branch-002 may consume. |
| `git_strategy.{manual,personal,team}.environment` | none | yaml-only | **forward-compat-scaffold** | `ModeProfile.Environment string`. |
| `git_strategy.{manual,personal,team}.github_integration` | none | yaml-only | **forward-compat-scaffold** | `ModeProfile.GitHubIntegration bool`. |
| `git_strategy.{manual,personal,team}.push_to_remote` | none | yaml-only | **forward-compat-scaffold** | `ModeProfile.PushToRemote bool`. |
| `git_strategy.{manual,personal,team}.automation.{auto_branch,auto_commit,auto_pr,auto_push}` | LATE-BRANCH-001 (skill body) | yaml-consumed by skill | **forward-compat-scaffold** | `AutomationConfig{AutoBranch, AutoCommit, AutoPR, AutoPush}`. Future Go consumer migration deferred. |
| `git_strategy.{manual,personal,team}.branch_creation.{auto_enabled,prompt_always}` | LATE-BRANCH-001 (REQ-LB-004 skill body) | yaml-consumed by skill | **forward-compat-scaffold** | `BranchCreationConfig{AutoEnabled, PromptAlways}`. Skill continues to read yaml directly until future Go wire-through SPEC. |
| `git_strategy.{manual,personal,team}.commit_style.{format,scope_required}` | none | yaml-only | **forward-compat-scaffold** | `CommitStyleConfig{Format, ScopeRequired}`. |
| `git_strategy.{manual,personal,team}.hooks.{pre_commit,pre_push,commit_msg}` | none | yaml-only | **forward-compat-scaffold** | `HooksConfig{PreCommit, PrePush, CommitMsg}`. |
| `git_strategy.manual.auto_checkpoint` | none | manual-only key | **forward-compat-scaffold** | `ModeProfile.AutoCheckpoint string` (manual mode only; nil for personal/team). |
| `git_strategy.{personal,team}.branch_prefix` | none (validation.go for FLAT field only) | yaml-only for nested | **forward-compat-scaffold** | `ModeProfile.BranchPrefix string` (manual mode lacks this key). |
| `git_strategy.{personal,team}.main_branch` | none | yaml-only | **forward-compat-scaffold** | `ModeProfile.MainBranch string`. |
| `git_strategy.team.{draft_pr,required_reviews,branch_protection}` | none | team-only keys | **forward-compat-scaffold** | `ModeProfile.DraftPR bool / RequiredReviews int / BranchProtection bool`. |
| `GitStrategyConfig.AutoBranch` (FLAT) | none | dead field | **WontDo (preserve via deprecation)** | Cannot remove (defaults_test.go + types_test.go reference). Mark `// Deprecated:` comment + accessor delegates to ActiveModeProfile().Automation.AutoBranch. |
| `GitStrategyConfig.WorktreeRoot` (FLAT) | none | dead field | **dead (preserve for SPEC-CONFIG-001 backward-compat)** | Not in production yaml. Preserve struct field with deprecation comment. Future removal SPEC may sunset. |

### 1.4 Backward-Compat Strategy (D1)

Three options considered:

- **Option (a)** — Keep FLAT fields as deprecated, lint warning on use.
  - Pros: zero breaking change.
  - Cons: yaml roundtrip still drops nested data unless struct extended; FLAT and nested coexist confusingly.
- **Option (b)** — Remove FLAT fields entirely (breaking change).
  - Pros: clean schema.
  - Cons: requires SPEC HISTORY amendment to SPEC-CONFIG-001 + SPEC-GIT-001; breaks `types_test.go` and `defaults_test.go`; SemVer major-bump.
- **Option (c)** — Compute FLAT field values from mode-based nested via accessor method ✅ **CHOSEN**.
  - Pros: backward-compat preserved; nested becomes canonical source; FLAT fields become read-only computed views.
  - Cons: adds accessor method complexity; existing test fixtures need migration to nested yaml.
  - Decision rationale: Option (c) satisfies REQ-V3R2-RT-005 yaml↔struct parity AND preserves SPEC-CONFIG-001 historical contract. Migration path is incremental.

### 1.5 Brownfield State Inventory

본 SPEC은 brownfield 작업이다. 모든 영향 파일은 [EXTEND] 분류이며 [NEW] 파일 1개 (test 파일).

| Marker | 파일 | 분류 | 근거 |
|--------|------|------|------|
| [EXTEND] | `internal/config/types.go` | EXTEND | `GitStrategyConfig` struct 확장 + 신규 sub-structs (`ModeProfile`, `AutomationConfig`, `BranchCreationConfig`, `CommitStyleConfig`, `HooksConfig`, `GitLabConfig`). FLAT 6필드는 deprecated 주석으로 보존. |
| [EXTEND] | `internal/config/defaults.go` | EXTEND | `NewDefaultGitStrategyConfig()` 갱신 — top-level mode/provider + 3 ModeProfile defaults (각 모드별 yaml template 기본값 정확 미러). |
| [EXTEND] | `internal/config/validation.go` | EXTEND | `checkStringField` 호출 확장 — `git_strategy.{personal,team}.branch_prefix`, `{manual,personal,team}.workflow`, hooks.* 등 nested 키 validation. |
| [EXTEND] | `internal/config/audit_registry.go` | EXTEND | yamlToStructRegistry 코멘트 갱신 (FLAT→nested 표기). audit_loader_completeness_test.go acknowledgedUnloadedSections에서 `git-strategy` 제거. |
| [EXTEND] | `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` | EXTEND | byte-byte identical 유지 (이미 nested 구조). 변경 없음 — 다만 본 SPEC이 이 파일을 schema 정의의 SSOT로 명시. |
| [NEW]    | `internal/config/git_strategy_nested_test.go` | NEW | nested yaml fixture (3 modes 완전 populated) → `yaml.Unmarshal` → assert all nested 필드 present. RoundTrip test. |
| [EXTEND] | `internal/config/types_test.go` | EXTEND | 기존 FLAT 테스트 보존 (backward-compat 검증) + 신규 nested 테스트 추가. |
| [EXTEND] | `internal/config/defaults_test.go` | EXTEND | `TestNewDefaultGitStrategyConfig` 확장 — 3 ModeProfile defaults assertion. |

---

## 2. EARS Requirements

**REQ-GSS-001** (Ubiquitous, **must-pass**): The system shall provide a `GitStrategyConfig` Go struct that contains top-level fields (`Mode string`, `Provider string`, `GitHubUsername string`, `GitLab GitLabConfig`) and three mode profile fields (`Manual ModeProfile`, `Personal ModeProfile`, `Team ModeProfile`) such that every key currently present in `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl` has a corresponding Go struct field reachable via dotted path.

**REQ-GSS-002** (Ubiquitous, **must-pass**): The `ModeProfile` struct shall contain sub-struct fields (`Automation AutomationConfig`, `BranchCreation BranchCreationConfig`, `CommitStyle CommitStyleConfig`, `Hooks HooksConfig`) plus scalar fields (`Workflow string`, `Environment string`, `GitHubIntegration bool`, `PushToRemote bool`) plus mode-conditional optional fields (`AutoCheckpoint string`, `BranchPrefix string`, `MainBranch string`, `DraftPR bool`, `RequiredReviews int`, `BranchProtection bool`).

**REQ-GSS-003** (Event-driven, **must-pass**): WHEN `yaml.Unmarshal` is invoked on a yaml document matching the current production `git-strategy.yaml` nested structure, THEN the resulting `GitStrategyConfig` value SHALL have all nested fields populated such that `cfg.Team.BranchCreation.AutoEnabled` resolves to the value of `git_strategy.team.branch_creation.auto_enabled` from the yaml input.

**REQ-GSS-004** (State-driven, **must-pass**): WHILE the `GitStrategyConfig.Mode` field holds one of `"manual"`, `"personal"`, or `"team"`, the system shall provide an `ActiveModeProfile()` accessor method on `GitStrategyConfig` that returns the corresponding `*ModeProfile` pointer (or `nil` if mode is empty/invalid). The accessor shall not modify any field.

**REQ-GSS-005** (Ubiquitous, **must-pass**): The system shall preserve the existing FLAT fields `AutoBranch bool`, `BranchPrefix string`, `CommitStyle string`, `WorktreeRoot string` on `GitStrategyConfig` with `// Deprecated:` Go doc comments per Option (c). The fields shall NOT be removed in this SPEC.

**REQ-GSS-006** (Event-driven): WHEN `NewDefaultGitStrategyConfig()` is called, THEN the returned value SHALL contain three populated `ModeProfile` instances whose field values exactly match the template defaults in `git-strategy.yaml.tmpl` (e.g., `Team.Automation.AutoPush == true`, `Manual.Automation.AutoPush == false`).

**REQ-GSS-007** (Ubiquitous): The system shall extend `internal/config/validation.go` to invoke `checkStringField` on at minimum: `git_strategy.{manual,personal,team}.workflow`, `git_strategy.{manual,personal,team}.environment`, `git_strategy.{manual,personal,team}.commit_style.format`, `git_strategy.{manual,personal,team}.hooks.{pre_commit,pre_push,commit_msg}`, and `git_strategy.{personal,team}.branch_prefix`. Pre-existing FLAT field validations (`git_strategy.branch_prefix`, `git_strategy.commit_style`) shall remain.

**REQ-GSS-008** (Ubiquitous): The system shall update `internal/config/audit_loader_completeness_test.go` to remove `"git-strategy"` from `acknowledgedUnloadedSections` once REQ-GSS-001 through REQ-GSS-006 are satisfied, since git-strategy will then have a complete struct loader path via `Loader.Load()` Config.GitStrategy field, not only via template rendering.

---

## 3. Out of Scope (Exclusions)

### 3.1 Out of Scope

- **EXCL-GSS-001**: Sunsetting / removing FLAT `AutoBranch`/`BranchPrefix`/`CommitStyle`/`WorktreeRoot` fields. Option (c) preserves these for backward-compat. A future SPEC may amend SPEC-CONFIG-001 to deprecate them entirely with SemVer major-bump.
- **EXCL-GSS-002**: Wiring `team.branch_creation.auto_enabled` consumption into Go code paths. Skill body (`spec-assembly.md` line 302) continues to read yaml directly. Forward-compat scaffold provides the struct field for future Go consumers without breaking the current skill-body workflow.
- **EXCL-GSS-003**: Adding migration tooling for existing user `git-strategy.yaml` files. Production yaml is already nested (no migration needed). Synthetic FLAT test fixtures in `types_test.go` migrate within this SPEC's scope.
- **EXCL-GSS-004**: Modifying `internal/cli/update.go` raw `map[string]any` mutation path (lines 2097-2148). That code path mutates only top-level keys (`mode`, `provider`, `github_username`, `gitlab.instance_url`) which are already covered. Refactoring to use the new struct accessors is deferred to a future SPEC.
- **EXCL-GSS-005**: Refactoring `internal/core/project/initializer.go:356-365` minimal-yaml-write logic. The yaml-write path uses `fmt.Sprintf` with template variables and is decoupled from struct definitions. Refactoring is deferred.
- **EXCL-GSS-006**: CATEGORY B/C/D dead config from v1 audit. Those belong to separate SPECs (per Section A constraints).

### 3.2 Future Work (referenced but not in this SPEC)

- SPEC-V3R5-GIT-STRATEGY-WIRE-001 (hypothetical) — wire Late-Branch keys into Go runtime consumers.
- SPEC-V3R5-CONFIG-FLAT-SUNSET-001 (hypothetical) — remove FLAT deprecated fields with SemVer major-bump.

---

## 4. Risks

- **R-GSS-001** (Medium): Test fixture migration from FLAT to nested yaml may surface latent assertion mismatches in `types_test.go` / `defaults_test.go`. Mitigation: incremental fixture update with both FLAT and nested assertions during transition; explicit roundtrip golden file.
- **R-GSS-002** (Low): `ActiveModeProfile()` accessor returning `nil` on invalid mode could trigger nil deref in future callers. Mitigation: accessor returns sentinel default ModeProfile (zero value) AND a boolean `ok` per Go idiom; tests cover empty/invalid mode cases (Edge-002, Edge-003).
- **R-GSS-003** (Low): yaml-struct symmetry test `git_strategy_nested_test.go` may flake on YAML library version drift (yaml.v3 vs yaml.v2 field-tag handling). Mitigation: pin yaml library version (already pinned in go.mod); test uses explicit `yaml:"key_name"` tags.
- **R-GSS-004** (Low): Removing `"git-strategy"` from `acknowledgedUnloadedSections` (REQ-GSS-008) may trigger CI failure on parallel branches that haven't yet adopted the new struct. Mitigation: gate REQ-GSS-008 behind successful merge of REQ-GSS-001 through REQ-GSS-007; commit ordering in plan.md M5.
- **R-GSS-005** (Medium): SPEC-V3R5-LATE-BRANCH-001 declares `team.branch_creation.auto_enabled` as skill-body consumed (REQ-LB-004). Adding a Go struct field with the same name may cause downstream developers to assume Go reads it. Mitigation: explicit `// Forward-compat scaffold` Go doc comment on the field referencing this SPEC and SPEC-V3R5-LATE-BRANCH-001 to prevent confusion.

---

## 5. Edge Cases

- **Edge-GSS-001**: Empty `mode:` field in yaml (`git_strategy.mode: ""`). `ActiveModeProfile()` returns `(nil, false)`. Existing `update.go` path that mutates `mode` value is unaffected.
- **Edge-GSS-002**: Invalid `mode:` value (e.g., `git_strategy.mode: "unknown"`). `ActiveModeProfile()` returns `(nil, false)`. Validation layer logs warning but does not fail (mode validity check is informational, not gate).
- **Edge-GSS-003**: gitlab provider with manual mode (`provider: gitlab`, `mode: manual`). `gitlab.instance_url` is read but manual mode's `github_integration: false` is also valid. `ActiveModeProfile()` returns `Manual` ModeProfile. No contradiction since gitlab.* is top-level, mode-independent.
- **Edge-GSS-004**: Existing user yaml file (pre-this-SPEC) missing nested keys (e.g., only top-level `mode/provider/github_username`). `yaml.Unmarshal` populates zero values for missing nested fields. `NewDefaultGitStrategyConfig()` is NOT auto-applied to user files — defaults are construction-time only. Skill body Phase 3 (line 302) handles missing keys with sensible defaults.
- **Edge-GSS-005**: `personal.branch_prefix` value `"moai/"` (legacy FLAT default) vs `"feature/SPEC-"` (current template default). Both are valid string values; validation only checks for non-empty. ActiveModeProfile()-based FLAT accessor returns the nested value.

---

## 6. Constitution Alignment

- **Technology Stack**: Go 1.21+ (current). YAML library `gopkg.in/yaml.v3` (existing dependency, no addition).
- **Naming Conventions**: PascalCase for exported Go struct fields. snake_case for yaml keys (existing convention preserved).
- **Forbidden Libraries**: None violated. No new dependencies.
- **Architectural Patterns**: Struct definition resides in `internal/config/types.go` (existing convention). Sub-structs co-located in the same file (consistent with `MigrationsConfig`, `SystemHookConfig` precedent at line 53-66).
- **Security Standards**: No security-sensitive fields. yaml parsing uses standard library (no custom parser).
- **Logging Standards**: Validation warnings use existing `checkStringField` infrastructure (no new logging code).

---

## 7. References

- **v2 Audit Source**: `.moai/research/config-audit-2026-05-22.md` §4 (6-step procedure).
- **SPEC-V3R5-LATE-BRANCH-001** (completed): Late-Branch yaml keys ownership — `team.branch_creation.{auto_enabled,prompt_always}` consumer documentation.
- **SPEC-CONFIG-001** (completed): Original FLAT `GitStrategyConfig` struct declaration (spec.md:414).
- **SPEC-GIT-001** (completed): References SPEC-CONFIG-001 GitStrategyConfig (spec.md:425).
- **SPEC-V3R2-RT-005** (implemented): yaml↔struct parity audit registry — establishes the audit framework this SPEC extends.
- **Production yaml**: `.moai/config/sections/git-strategy.yaml` (current nested structure ~70 keys).
- **Template SSOT**: `internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl`.
- **Audit completeness test**: `internal/config/audit_loader_completeness_test.go:19` (current exception entry).

---

End of spec.md.

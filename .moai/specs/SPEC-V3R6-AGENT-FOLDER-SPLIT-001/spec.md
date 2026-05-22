---
id: SPEC-V3R6-AGENT-FOLDER-SPLIT-001
title: "Agent Folder Split (.claude/agents/moai/ → core/ + expert/ + meta/)"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: ".claude/agents, internal/template/templates/.claude/agents, internal/template/catalog.yaml"
lifecycle: spec-anchored
tags: "agent, folder-restructure, foundation, wave-2, v3.0.0, template-first"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-HARNESS-RENAME-001]
related_specs: [SPEC-V3R6-META-HARNESS-PATH-001, SPEC-V3R6-AGENT-SLIM-001]
---

# SPEC-V3R6-AGENT-FOLDER-SPLIT-001: Agent Folder Split

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial draft — Wave 2 결정에 따라 `.claude/agents/moai/` 19 agents를 `core/` + `expert/` + `meta/` 3개 폴더로 분리. Template-First Rule 동시 적용. `harness/` 4 agents는 PRESERVE (HARNESS-RENAME-001 결과). Tier M. |
| 0.1.1 | 2026-05-22 | manager-spec | iter 2 scope reduction + D1-D8 fix per plan-auditor (iter 1 REVISE 0.69 → iter 2 self-estimate target ≥0.82). **Scope reduction**: 9 Go files OUT OF SCOPE (frozen_guard.go family + pre_tool.go SentinelHarnessFrozenAgent + walker tests TestAllAgentsInCatalog/TestAgentFrontmatterAudit + safety preservation/pipeline tests + meta_invocation_test). 별도 SPEC 후보 `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) 식별. D1: 22 occurrences in 19 files enumerated. D2: in-scope Go = 41 occurrences in 21 files (out-of-scope = 42 occurrences in 9 files). D6: 3-commit strategy explicit. D7: §10 pre-flight baseline 구체적 숫자. D8: AC-AFS-006 hedge word removed. New AC-AFS-012: out-of-scope file PRESERVE verification. catalog.yaml 19 entries (not 18 — builder-harness.md line 341 nested) confirmed. |
| 0.2.0 | 2026-05-22 | manager-develop | run-phase COMPLETE. 3 commits on main (1bd083725 M1+M2 / fdf325eb6 M3 catalog hash / a912e7d5a extended walker coverage). 38 git mv + 82 Edits in 21 in-scope Go files + 6 additional walker callers (initializer/validator/model_policy + 3 test files) discovered during M4 verification. 9 out-of-scope files byte-identical PRESERVE (AC-AFS-012 PASS, 42 occurrences). 12/12 ACs PASS (AC-AFS-008 partial — only meta/plan-auditor.md inherits pre-existing template-mirror drift, 3 other folders zero-drift). Test classification: 3 baseline + 1 inherited statusline baseline + 3 out-of-scope deferred residual (TestAllAgentsInCatalog / TestAgentFrontmatterAudit / TestRetirementCompletenessAssertion) + 0 NEW regression. Cross-platform build PASS. Follow-up SPEC: SPEC-V3R6-FROZEN-PREFIX-REALIGN-001 (가칭) for walker logic + retirement-replacement-path supersession. |

## 1. Goal

`.claude/agents/moai/` 단일 폴더에 누적된 19개 agent definition을 도메인 기반 3개 폴더로 분리하여 navigability와 분류 가능성을 확보한다. 동일한 폴더 구조를 `internal/template/templates/.claude/agents/`에 Template-First Rule로 적용한다. `.claude/agents/harness/` 4 agents (HARNESS-RENAME-001 결과)는 그대로 보존한다.

**최종 폴더 구조** (4 폴더):

```
.claude/agents/
  ├── core/      (8 agents)  — manager-* workflow operational
  ├── expert/    (6 agents)  — expert-* domain specialists
  ├── meta/      (5 agents)  — builder/evaluator/auditor/guide/researcher
  └── harness/   (4 agents)  — moai-harness-*-specialist (PRESERVE, from HARNESS-RENAME-001)
```

## 2. Why

### 2.1 현재 문제 (Status Quo)

- `.claude/agents/moai/`에 19개 agent가 평면 적재되어 있음
- agent 분류 (manager / expert / builder / evaluator / auditor / researcher / guide)가 폴더 구조로 표현되지 않음
- Wave 3 (SPEC-V3R6-AGENT-SLIM-001)에서 19→11 slim-down 진행 시 사전에 도메인 경계 명확화가 필요
- 새 agent 추가 시 분류 위치 불명확 (현재는 모두 `moai/`에 추가)

### 2.2 v3.0.0 redesign blueprint § Wave 2 결정

`.moai/research/v3-redesign-blueprint-2026-05-22.md`:

> SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (M — `.claude/agents/moai/` → `core/` + `expert/` + `meta/`)

Wave 2 sequential SPEC sequence:
1. ~~SPEC-V3R6-HARNESS-RENAME-001~~ (DONE — PR #1043 merged 2026-05-22)
2. **SPEC-V3R6-AGENT-FOLDER-SPLIT-001** (THIS SPEC)
3. SPEC-V3R6-META-HARNESS-PATH-001 (다음)

### 2.3 Template-First Rule (CLAUDE.local.md §2)

[ZONE:Frozen] [HARD] `.claude/agents/` 변경은 반드시 `internal/template/templates/.claude/agents/`에도 동일 적용되어야 한다. 사용자가 명시 강조: "템플릿에도 동일하게 적용해야한다".

### 2.4 iter 2 Scope Reduction Rationale (NEW)

iter 1 plan-auditor (REVISE 0.69) 결과 5 BLOCKING defects (D1-D5) 중 D3/D4/D5는 production security code (`frozen_guard.go` + `pre_tool.go SentinelHarnessFrozenAgent` table + walker tests) 영향. 본 SPEC scope에서 처리 시:

1. **Tier L 승격 위험**: 9 추가 Go files 영향 + Tier M 한계 (5-15 files) 초과
2. **Cross-SPEC tension**: FROZEN prefix 정책 (`.claude/agents/moai/` literal로 hardcoded in `pre_tool.go:720` SentinelHarnessFrozenAgent table) → 단순 literal rename이 user/system distinction 손상 가능
3. **Walker semantic conflict**: `TestAllAgentsInCatalog` (`fs.WalkDir(fsys, ".claude/agents/moai", ...)`) + `TestAgentFrontmatterAudit` (3 WalkDir calls) — walker logic 정정은 별도 SPEC scope로 분리하여 정확성 확보

**사용자 결정 (AskUserQuestion 2026-05-22)**: scope reduction + 별도 후속 SPEC `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) 분리. 본 SPEC implemented 이후 별도 plan-phase.

## 3. Agent Classification

19개 agent를 다음 3개 폴더로 분류한다. PRESERVE된 `harness/` 4개와 합쳐 총 23개.

### 3.1 core/ — 8 agents (manager-* workflow operational)

| Agent | Current Path | New Path |
|---|---|---|
| manager-brain | `.claude/agents/moai/manager-brain.md` | `.claude/agents/core/manager-brain.md` |
| manager-develop | `.claude/agents/moai/manager-develop.md` | `.claude/agents/core/manager-develop.md` |
| manager-docs | `.claude/agents/moai/manager-docs.md` | `.claude/agents/core/manager-docs.md` |
| manager-git | `.claude/agents/moai/manager-git.md` | `.claude/agents/core/manager-git.md` |
| manager-project | `.claude/agents/moai/manager-project.md` | `.claude/agents/core/manager-project.md` |
| manager-quality | `.claude/agents/moai/manager-quality.md` | `.claude/agents/core/manager-quality.md` |
| manager-spec | `.claude/agents/moai/manager-spec.md` | `.claude/agents/core/manager-spec.md` |
| manager-strategy | `.claude/agents/moai/manager-strategy.md` | `.claude/agents/core/manager-strategy.md` |

분류 근거: MoAI workflow의 핵심 운영 agent. `/moai plan|run|sync|project|quality|...` 명령의 1차 위임 대상.

### 3.2 expert/ — 6 agents (domain specialists)

| Agent | Current Path | New Path |
|---|---|---|
| expert-backend | `.claude/agents/moai/expert-backend.md` | `.claude/agents/expert/expert-backend.md` |
| expert-devops | `.claude/agents/moai/expert-devops.md` | `.claude/agents/expert/expert-devops.md` |
| expert-frontend | `.claude/agents/moai/expert-frontend.md` | `.claude/agents/expert/expert-frontend.md` |
| expert-performance | `.claude/agents/moai/expert-performance.md` | `.claude/agents/expert/expert-performance.md` |
| expert-refactoring | `.claude/agents/moai/expert-refactoring.md` | `.claude/agents/expert/expert-refactoring.md` |
| expert-security | `.claude/agents/moai/expert-security.md` | `.claude/agents/expert/expert-security.md` |

분류 근거: 특정 도메인 전문성 (백엔드/프론트엔드/보안/DevOps/성능/리팩토링). 모두 `expert-` prefix 통일.

### 3.3 meta/ — 5 agents (builders, evaluators, auditors, guides, researchers)

| Agent | Current Path | New Path |
|---|---|---|
| builder-harness | `.claude/agents/moai/builder-harness.md` | `.claude/agents/meta/builder-harness.md` |
| claude-code-guide | `.claude/agents/moai/claude-code-guide.md` | `.claude/agents/meta/claude-code-guide.md` |
| evaluator-active | `.claude/agents/moai/evaluator-active.md` | `.claude/agents/meta/evaluator-active.md` |
| plan-auditor | `.claude/agents/moai/plan-auditor.md` | `.claude/agents/meta/plan-auditor.md` |
| researcher | `.claude/agents/moai/researcher.md` | `.claude/agents/meta/researcher.md` |

분류 근거: MoAI 자체를 빌드/평가/감사/안내/조사하는 메타-수준 agent. workflow operational (`core/`)이나 domain expert (`expert/`)에 속하지 않음.

### 3.4 harness/ — 4 agents (PRESERVE)

HARNESS-RENAME-001 결과로 이미 `.claude/agents/harness/`에 위치. 변경 없음.

| Agent | Path |
|---|---|
| moai-harness-cli-template-specialist | `.claude/agents/harness/cli-template-specialist.md` |
| moai-harness-hook-ci-specialist | `.claude/agents/harness/hook-ci-specialist.md` |
| moai-harness-quality-specialist | `.claude/agents/harness/quality-specialist.md` |
| moai-harness-workflow-specialist | `.claude/agents/harness/workflow-specialist.md` |

## 4. EARS Requirements

### REQ-AFS-001 (Ubiquitous, Folder Creation)

[ZONE:Evolvable] [HARD] `/moai run` 실행 시, 시스템은 `.claude/agents/{core,expert,meta}/` 3개 폴더와 `internal/template/templates/.claude/agents/{core,expert,meta}/` 3개 폴더를 생성하고, 19개 agent file을 §3 Classification matrix에 따라 정확히 배치해야 한다.

### REQ-AFS-002 (Event-driven, Git mv preservation)

WHEN agent file을 이동할 때, 시스템은 `git mv` 명령을 사용하여 file history continuity를 보존해야 한다. `Write + rm` 대신 rename 등록이 의무이다.

### REQ-AFS-003 (Ubiquitous, Frontmatter preservation)

[ZONE:Evolvable] [HARD] agent file의 YAML frontmatter는 byte-identical로 보존되어야 한다. `name:` field 값, `hooks:` config, `skills:` array, `model:`, `effort:`, `permissionMode:`, `memory:` 등 모든 필드를 그대로 유지한다. 폴더만 변경.

### REQ-AFS-004 (Event-driven, In-scope cross-reference update)

WHEN agent file이 새 경로로 이동된 후, 시스템은 다음 in-scope source에서 옛 경로 `.claude/agents/moai/<name>.md`를 새 경로 `.claude/agents/{folder}/<name>.md`로 갱신해야 한다 (out-of-scope 9 Go files는 PRESERVE — §6.2.2 참조):

**In-scope active rule/skill files (22 occurrences in 19 unique files)**:

| File (local) | Occurrence count | Lines |
|---|---|---|
| `.claude/rules/moai/development/agent-authoring.md` | 1 | 15 |
| `.claude/rules/moai/development/model-policy.md` | 2 | 45, 46 |
| `.claude/rules/moai/workflow/spec-workflow.md` | 2 | 45, 137 |
| `.claude/rules/moai/workflow/team-protocol.md` | 1 | 114 |
| `.claude/skills/moai-harness-learner/SKILL.md` | 1 | 151 |
| `.claude/skills/moai-meta-harness/SKILL.md` | 1 | 352 |
| `.claude/skills/moai/workflows/harness.md` | 1 | 175 |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | 1 | 308 |
| `.claude/skills/moai/workflows/project/meta-harness.md` | 1 | 227 |
| `.claude/skills/moai/workflows/release.md` | 1 | 878 (local-only per CLAUDE.local.md §21, no template mirror) |
| **Local subtotal** | **12** | **10 files** |
| Template mirrors of all above (excluding release.md — dev-only) | 10 | 9 files |
| **Grand total** | **22** | **19 unique files** |

**In-scope Go code files (41 occurrences in 21 files)**: §6.1.2 enumeration 참조.

### REQ-AFS-005 (Ubiquitous, Template-First Rule)

[ZONE:Frozen] [HARD] `.claude/agents/` 변경과 `internal/template/templates/.claude/agents/` 변경은 동일 commit 또는 동일 SPEC 내에 함께 적용되어야 한다. drift 발생 시 `TestRuleTemplateMirrorDrift` 또는 동등 CI guard test가 fail해야 한다.

### REQ-AFS-006 (Ubiquitous, Backward-compat aliasing prohibition)

[ZONE:Evolvable] [HARD] 옛 경로 `.claude/agents/moai/<name>.md`에 대한 symlink, alias, 또는 stub file을 생성하는 것은 금지된다. 이전 buyer는 cross-ref 갱신 누락을 추후 발견하기 어렵게 만든다. 옛 폴더 `.claude/agents/moai/`는 완전히 제거되어야 한다.

### REQ-AFS-007 (Event-driven, catalog.yaml hash regeneration)

WHEN agent file 경로가 변경된 후, `internal/template/catalog.yaml`의 19개 agent entry의 `path:` 필드는 새 경로로 갱신되어야 하며, `make build`가 자동으로 hash regeneration을 수행해야 한다 (CATALOG-SSOT-001 mechanism). **수정**: iter 1은 "18 entries"라 기재되어 있었으나 실제 19개 (builder-harness.md line 341 nested section 포함).

### REQ-AFS-008 (Ubiquitous, harness/ preservation)

[ZONE:Frozen] [HARD] `.claude/agents/harness/`와 `internal/template/templates/.claude/agents/harness/` 4 agent file은 변경되지 않아야 한다 (cli-template-specialist.md, hook-ci-specialist.md, quality-specialist.md, workflow-specialist.md). 본 SPEC scope 외부.

### REQ-AFS-009 (Optional, name field unchanged)

WHERE agent의 `name:` frontmatter field 값을 변경하지 말아야 하는 곳에서, 시스템은 file 위치만 이동하고 `name:` 값을 그대로 유지해야 한다. 예: `manager-spec`은 폴더가 `moai/`에서 `core/`로 바뀌어도 `name: manager-spec` 그대로.

### REQ-AFS-010 (Ubiquitous, Out-of-scope file PRESERVE) — NEW iter 2

[ZONE:Frozen] [HARD] 시스템은 다음 9개 Go files를 변경하지 말아야 한다 (out-of-scope, follow-up SPEC `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` 후보):

| File | Reason for OUT OF SCOPE |
|---|---|
| `internal/harness/frozen_guard.go` | FROZEN prefix detection logic — production security gate |
| `internal/harness/safety/frozen_guard.go` | FROZEN prefix detection (safety subpackage mirror) |
| `internal/harness/safety/frozen_guard_test.go` | FROZEN prefix expected-array test fixtures (8 occurrences) |
| `internal/harness/safety_preservation_test.go` | FROZEN prefix preservation contract test |
| `internal/harness/safety/pipeline_test.go` | FROZEN prefix detection pipeline test |
| `internal/harness/meta_invocation_test.go` | FROZEN prefix forbidden-invocation test (4 occurrences) |
| `internal/hook/pre_tool.go` | SentinelHarnessFrozenAgent table entry (line 720) + comment (line 35) |
| `internal/template/catalog_tier_audit_test.go` | TestAllAgentsInCatalog hardcoded WalkDir on `.claude/agents/moai` (line 178) |
| `internal/template/agent_frontmatter_audit_test.go` | TestAgentFrontmatterAudit 3 hardcoded WalkDir on `.claude/agents/moai` (lines 72, 185, 485) — 20 occurrences |

이 9 files는 run-phase에서 byte-identical로 보존되어야 한다. 결과적으로 일부 test (TestAllAgentsInCatalog, TestAgentFrontmatterAudit, FROZEN guard tests)는 본 SPEC implemented 후 FAIL 상태로 전환될 수 있으며, 이는 baseline residual 3 FAIL과 별개로 "out-of-scope deferred residual"로 분류된다. AC-AFS-012로 verify.

### REQ-AFS-011 (Ubiquitous, dirty working tree preservation)

[ZONE:Evolvable] [HARD] 시스템은 현재 dirty working tree (modified files + untracked files)를 변경하거나 제거하지 말아야 한다. PRESERVE list:
- 4 modified files: `.moai/harness/usage-log.jsonl`, `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html`
- 4 untracked SPEC dirs: `SPEC-V3R5-GIT-STRATEGY-SCHEMA-001`, `SPEC-V3R5-INIT-WIZARD-EXPANSION-001`, `SPEC-V3R5-STATUSLINE-PROFILE-WIZARD-001`, `SPEC-V3R5-WORKFLOW-SCHEMA-EXTEND-001`
- 2 untracked research files
- docs-site/data, docs-site/scripts, internal/hook/.moai/

## 5. Constraints

### 5.1 Technical Constraints

- Git `mv` 사용 의무 (file history 보존)
- agent `name:` frontmatter 필드 변경 금지 (역호환성)
- `harness/` 폴더는 PRESERVE (HARNESS-RENAME-001 산출물)
- 9 out-of-scope Go files PRESERVE (REQ-AFS-010)
- Cross-platform build PASS 의무: `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` 둘 다 exit 0
- spec-lint NEW regression 0건 (baseline 잔여 22 NEW preserve)
- Conventional Commits + `🗿 MoAI <email@mo.ai.kr>` trailer 의무

### 5.2 Cross-SPEC Constraints

- depends_on: SPEC-V3R6-HARNESS-RENAME-001 (`harness/` 4 agent 분리 선행 완료)
- 후행 SPEC-V3R6-META-HARNESS-PATH-001은 본 SPEC implemented 이후 진행
- Wave 3 SPEC-V3R6-AGENT-SLIM-001은 본 SPEC implemented 이후 진행 (19→11 slim-down은 새 폴더 구조 기준)
- **NEW**: Follow-up `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) — 본 SPEC implemented 후 별도 plan-phase. 9 out-of-scope Go files의 FROZEN prefix detection 및 walker logic 정정 담당. REQ-AFS-010 PRESERVE를 supersede할 예정.

### 5.3 PRESERVE Constraints

- `.claude/agents/harness/` 4 files (HARNESS-RENAME-001 결과)
- 9 out-of-scope Go files (REQ-AFS-010 명시 list)
- §4.11 enumeration의 dirty working tree files
- Active SPEC frontmatter historical artifact integrity (다른 SPEC의 `module:` 필드에 `.claude/agents/moai/...` 문자열이 있어도 historical record로 보존; 본 SPEC scope는 active rule + skill + 21 in-scope Go files만)

### 5.4 Forbidden Operations

- `git reset --hard` (대신 `--keep`)
- `--no-verify` flag on commit
- `--amend` (대신 새 commit)
- force-push to main
- backward-compat symlink 또는 alias file 생성
- agent `name:` field 값 변경
- 9 out-of-scope Go files 수정

## 6. Scope and Out of Scope

### 6.1 In Scope

#### 6.1.1 File movements (38 git mv)

- **Local**: 19 agent files from `.claude/agents/moai/` → `.claude/agents/{core,expert,meta}/`
- **Template mirror**: 19 agent files from `internal/template/templates/.claude/agents/moai/` → `internal/template/templates/.claude/agents/{core,expert,meta}/`

#### 6.1.2 Cross-reference updates

**Active rule + skill files (22 occurrences in 19 unique files)**: REQ-AFS-004 표 참조.

**catalog.yaml**: 19 `path:` field updates (template path prefix changes — iter 1은 18로 잘못 기재).

**In-scope Go code (41 occurrences in 21 files)**:

| File | Occurrence count |
|---|---|
| `internal/cli/agent_lint.go` | 2 |
| `internal/cli/plan_audit_d7_d8_test.go` | 1 |
| `internal/cli/update.go` | 1 |
| `internal/cli/update_preserve_my_harness_test.go` | 1 |
| `internal/cli/update_safety_test.go` | 1 |
| `internal/core/project/initializer_test.go` | 1 |
| `internal/hook/agent_start_test.go` | 2 |
| `internal/hook/subagent_start.go` | 1 |
| `internal/manifest/types_test.go` | 1 |
| `internal/merge/confirm_coverage_test.go` | 1 |
| `internal/research/safety/frozen.go` | 1 |
| `internal/research/safety/frozen_test.go` | 3 |
| `internal/template/builder_skill_path_test.go` | 1 |
| `internal/template/deployer_bench_test.go` | 1 |
| `internal/template/deployer_mode_test.go` | 8 |
| `internal/template/deployer_test.go` | 4 |
| `internal/template/embed.go` | 2 |
| `internal/template/manager_develop_present_test.go` | 3 |
| `internal/template/rule_template_mirror_test.go` | 3 |
| `internal/template/slim_guard.go` | 1 |
| `internal/template/slim_guard_test.go` | 2 |
| **Total** | **41 occurrences in 21 files** |

#### 6.1.3 Folder removal

- `.claude/agents/moai/` 빈 폴더 제거 (`git mv` 후 git이 자동으로 빈 디렉토리 추적 안함)
- `internal/template/templates/.claude/agents/moai/` 빈 폴더 제거

#### 6.1.4 Cascade test count adjustments

- 본 SPEC은 catalog.yaml entry 개수 (19 agent entries)는 그대로 유지 (path만 변경)
- `TestAllAgentsInCatalog` (현재 expected 19개, on disk 19개)는 PRESERVE — out-of-scope (REQ-AFS-010), 본 SPEC implemented 후 walker가 `.claude/agents/moai`에서 file을 찾지 못해 FAIL 전환 → follow-up SPEC `FROZEN-PREFIX-REALIGN-001` scope

### 6.2 Out of Scope

#### 6.2.1 Out of Scope Items

- **agent `name:` field 값 변경**: `name: manager-spec`는 그대로 유지, 폴더만 변경
- **agent slim-down** (19→11): SPEC-V3R6-AGENT-SLIM-001 (Wave 3)
- **`harness/` 4 agents 재분류**: HARNESS-RENAME-001 산출물 PRESERVE
- **`/moai sync` PR 생성**: 본 SPEC run-phase는 SPEC implemented 까지. PR 별도 sync-phase.
- **CLAUDE.md / CLAUDE.local.md agent catalog 문서 prose 수정**: Wave 6 SPEC-V3R6-RULES-COMPLIANCE-001 또는 별도 doc-only SPEC. 본 SPEC은 active rule/skill/code path만 갱신.
- **Historical SPEC artifact `module:` field 갱신**: 다른 SPEC의 `module:` 또는 본문에 옛 경로 문자열이 있어도 historical record로 보존 (예: SPEC-V3R5-LINT-CLEAN-001, SPEC-AGENT-002 등)
- **Memory file 갱신**: `.claude/agent-memory/`, `~/.claude/projects/.../memory/` 옛 경로 문자열 보존 (historical)
- **Research file 갱신**: `.moai/research/architecture-audit-2026-05-18.md` 등 옛 경로 문자열 보존
- **docs-site Hugo content 갱신**: 별도 docs-only SPEC

#### 6.2.2 Out-of-scope Go files (REQ-AFS-010 PRESERVE list) — NEW iter 2

9 Go files PRESERVE (byte-identical), follow-up SPEC `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) scope:

| File | Lines | Rationale |
|---|---|---|
| `internal/harness/frozen_guard.go` | 1 | FROZEN prefix detection logic — production security gate. Literal `.claude/agents/moai/` is the user/system distinction identifier. |
| `internal/harness/safety/frozen_guard.go` | 1 | safety subpackage mirror of frozen_guard |
| `internal/harness/safety/frozen_guard_test.go` | 8 | FROZEN prefix expected-array test fixtures |
| `internal/harness/safety_preservation_test.go` | 1 | FROZEN prefix preservation contract |
| `internal/harness/safety/pipeline_test.go` | 1 | FROZEN prefix detection pipeline test |
| `internal/harness/meta_invocation_test.go` | 4 | FROZEN prefix forbidden-invocation test |
| `internal/hook/pre_tool.go` | 2 (line 35 comment + line 720 `SentinelHarnessFrozenAgent` table entry) | hardcoded sentinel mapping `.claude/agents/moai/` → `SentinelHarnessFrozenAgent` — semantic supersession only |
| `internal/template/catalog_tier_audit_test.go` | 4 (line 178 `fs.WalkDir(fsys, ".claude/agents/moai", ...)`) | TestAllAgentsInCatalog walker logic |
| `internal/template/agent_frontmatter_audit_test.go` | 20 (3 WalkDir + per-agent fixture references) | TestAgentFrontmatterAudit walker logic + per-agent retired/replacement fixtures |

**OUT-of-scope deferred residual** (post-implementation FAIL state acceptable, follow-up SPEC가 정정):
- `TestAllAgentsInCatalog` (WalkDir returns empty → expected 19 vs 0 FAIL)
- `TestAgentFrontmatterAudit` (WalkDir + agent file existence checks FAIL)
- `TestFrozenPrefixDetection` 류 frozen_guard_test (hardcoded prefix expectations 정상 동작 유지 — `.claude/agents/moai/` literal는 FROZEN prefix detector input으로 사용, 실제 file은 없어도 detector 로직 자체는 PASS)

#### 6.2.3 Deferred Risks

- 본 SPEC implemented 후 다른 contributor의 PR가 옛 경로로 add file 시도 가능 → run-phase에서 CI guard test로 차단
- `.claude/settings.json` 또는 `.claude/settings.local.json`에 옛 경로 hard-coded references → 발견 시 별도 SPEC 또는 본 SPEC fix-up commit

## 7. Risks

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| R-AFS-001: Go test code의 fixture path 갱신 누락 (21 in-scope files) | M | H | §6.1.2 enumeration grep으로 모든 `.claude/agents/moai/` 문자열 search 후 fixup. acceptance.md AC-AFS-006로 verification. |
| R-AFS-002: catalog.yaml hash mismatch (Makefile gen-catalog-hashes.go --all 누락) | L | M | CATALOG-SSOT-001 mechanism (make build → gen-catalog-hashes.go --all 자동 실행)가 이미 작동. `TestManifestHashFormat` baseline guard. |
| R-AFS-003: 19개 file × 2 (local + template) = 38 git mv 부분 실패 | L | H | acceptance AC-AFS-001 (final tree structure 검증)와 AC-AFS-002 (file count) 둘 다 PASS 필요. 3-commit strategy (M1+M2 / M3 / M4 — §1.2) 적용으로 atomic 단위 분할. |
| R-AFS-004: `name:` frontmatter 누락된 file을 다른 곳에서 invoke 시 fail | L | H | REQ-AFS-003에서 frontmatter byte-identical 보존 의무. AC-AFS-005에서 sample frontmatter diff verify. |
| R-AFS-005: 본 SPEC implemented 후 다른 SPEC PR가 옛 경로로 add file 시도 | M | M | CI guard test 추가 권장 (run-phase 결정). `TestNoAgentMoaiPathRemnant` 등 grep-based guard. |
| R-AFS-006 (NEW iter 2): 9 out-of-scope Go files 실수로 수정 시 cross-SPEC tension | M | H | AC-AFS-012 (git diff origin/main check) 의무. M2 phase에서 grep filtering: `grep -rln '\.claude/agents/moai/' internal/ \| grep -v -E '(frozen_guard\|safety_preservation\|safety/pipeline\|meta_invocation\|pre_tool\|catalog_tier_audit\|agent_frontmatter_audit)'` 로 in-scope만 추출. |
| R-AFS-007 (NEW iter 2): `TestAllAgentsInCatalog` + `TestAgentFrontmatterAudit` post-implementation FAIL 상태가 baseline 3 FAIL과 혼선 | L | M | progress.md에 explicit "out-of-scope deferred residual" 분류 + follow-up SPEC `FROZEN-PREFIX-REALIGN-001` (가칭) 명시. Baseline 3 FAIL (`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers`) + Out-of-scope deferred FAIL (TestAllAgentsInCatalog, TestAgentFrontmatterAudit, 일부 frozen_guard_test) 두 카테고리 명확 분리. |

## 8. Dependencies

### 8.1 Explicit Dependencies

- **SPEC-V3R6-HARNESS-RENAME-001** (PR #1043 merged 2026-05-22): `harness/` 4 agents 분리 완료 — 본 SPEC scope의 PRESERVE 영역 정의.
- **SPEC-V3R6-CATALOG-SSOT-001** (run-COMPLETE 2026-05-22): `make build`가 `gen-catalog-hashes.go --all`을 자동 실행하는 self-healing manifest gate가 이미 작동.

### 8.2 Cross-cutting

- Template-First Rule (CLAUDE.local.md §2): 본 SPEC 적용 의무
- 1인 OSS Hybrid Trunk policy (commit cd9eead14, CLAUDE.local.md §23): `auto_branch: true`, `auto_pr: true`, `branch_prefix: feat/SPEC-`

### 8.3 Follow-up SPEC (NEW iter 2)

- **SPEC-V3R6-FROZEN-PREFIX-REALIGN-001** (가칭, 본 SPEC implemented 후 plan-phase): 9 out-of-scope Go files (REQ-AFS-010 list)의 FROZEN prefix detection 정책 및 walker logic 정정. 본 SPEC implemented 후 발생하는 "out-of-scope deferred FAIL" 상태 (TestAllAgentsInCatalog + TestAgentFrontmatterAudit)를 정정. Scope: ~17 LOC change in 6-7 files + walker logic restructuring.

## 9. Success Metrics

### 9.1 Quantitative

- 19개 agent file이 §3 classification matrix대로 정확한 폴더에 위치 (local + template = 38)
- `.claude/agents/moai/` 폴더 완전 제거 (local + template)
- `internal/template/catalog.yaml` 19개 entry 모두 새 경로 reflect (iter 1 "18"은 오기 — builder-harness.md line 341 nested section 포함하면 19)
- `git mv` 38 rename 등록 (19 local + 19 template)
- `go build ./...` exit 0, `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- `go test ./...`: 본 SPEC scope NEW regression 0건 (baseline FAIL 3건 + out-of-scope deferred FAIL TestAllAgentsInCatalog + TestAgentFrontmatterAudit 보존)
- spec-lint NEW regression 0건
- 9 out-of-scope Go files byte-identical PRESERVE (AC-AFS-012)

### 9.2 Qualitative

- 새 agent 추가 시 분류 위치 자명함 (manager → core/, expert → expert/, builder/evaluator/auditor/guide/researcher → meta/)
- Wave 3 SPEC-V3R6-AGENT-SLIM-001 (19→11) 진행 시 폴더 단위 평가 가능
- Follow-up SPEC `FROZEN-PREFIX-REALIGN-001`로 frozen_guard 정책 cleanup 경로 확보

## 10. Pre-flight Baseline (2026-05-22)

본 SPEC plan-phase iter 2 작성 시점 (precise grep numbers):

### 10.1 File counts

- Branch: `main`, HEAD `a809e0b98`
- Local `.claude/agents/moai/` = **19 .md files** (builder-harness, claude-code-guide, evaluator-active, expert-backend, expert-devops, expert-frontend, expert-performance, expert-refactoring, expert-security, manager-brain, manager-develop, manager-docs, manager-git, manager-project, manager-quality, manager-spec, manager-strategy, plan-auditor, researcher)
- Local `.claude/agents/harness/` = **4 .md files** (cli-template-specialist, hook-ci-specialist, quality-specialist, workflow-specialist)
- Template mirror identical: **19 + 4 = 23 .md files**

### 10.2 Cross-reference counts (precise grep)

- **catalog.yaml**: 19 `path: templates/.claude/agents/moai/` entries (lines 129, 154, 159, 164, 169, 174, 179, 184, 189, 194, 199, 204, 209, 235, 246, 303, 308, 328, 341) + 4 harness entries (lines 134, 139, 144, 149) = 23 total agent path entries (iter 1 "18" 오기, 실제 19개)
- **Active rule + skill direct refs**: **22 occurrences in 19 unique files** (12 local + 10 template) — REQ-AFS-004 표 참조
- **In-scope Go code refs**: **41 occurrences in 21 files** — §6.1.2 표 참조
- **Out-of-scope Go code refs**: **42 occurrences in 9 files** (REQ-AFS-010 PRESERVE list)
- **Total `.claude/agents/moai/` occurrences in internal/**: 122 (19 catalog + 20 template mirror + 41 in-scope Go + 42 out-of-scope Go = 122; verified by `grep -rn '\.claude/agents/moai/' internal/ | wc -l`)

### 10.3 Edit count summary

- **38 git mv operations** (19 local + 19 template)
- **22 Edit operations** in 19 active rule/skill files
- **41 Edit operations** in 21 in-scope Go files
- **19 Edit operations** in catalog.yaml (path field replacements)
- **0 Edit operations** in 9 out-of-scope Go files (PRESERVE)
- **Total: 38 git mv + ~82 Edit operations across ~41 unique files**

### 10.4 Baseline test failures

Pre-existing baseline 3 FAIL (`project_v3r6_template_mirror_drift_audit_2026_05_22` 기록): `TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers`. 본 SPEC run-phase는 이들을 fail-state로 그대로 보존.

**NEW post-implementation FAIL (out-of-scope deferred, follow-up SPEC scope)**:
- `TestAllAgentsInCatalog` (catalog_tier_audit_test.go:174-211) — WalkDir on `.claude/agents/moai` returns 0 files → `len(diskAgents) != 19` FAIL
- `TestAgentFrontmatterAudit` (agent_frontmatter_audit_test.go) — 3 WalkDir calls fail similarly + per-agent fs.Stat checks

이 두 FAIL은 acceptable "out-of-scope deferred residual" 카테고리이며, `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) follow-up SPEC가 정정 담당.

## 11. References

- v3.0.0 redesign blueprint: `.moai/research/v3-redesign-blueprint-2026-05-22.md` § Wave 2
- HARNESS-RENAME-001: `.moai/specs/SPEC-V3R6-HARNESS-RENAME-001/spec.md` (선례 — folder rename + Template-First + cross-ref pattern)
- CATALOG-SSOT-001: `.moai/specs/SPEC-V3R6-CATALOG-SSOT-001/` (Makefile gen-catalog-hashes.go --all 자동 실행 mechanism)
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Agent authoring rules: `.claude/rules/moai/development/agent-authoring.md`
- CLAUDE.local.md §2 (Template-First Rule), §21 (Dev-Only Commands Isolation — workflows/release.md), §23 (1인 OSS Hybrid Trunk)

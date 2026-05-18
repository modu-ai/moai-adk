---
id: SPEC-V3R4-WORKFLOW-SPLIT-001
title: "Workflow Skills Phase-Scoped Sub-Skill Split (Bundle F)"
version: "0.2.0"
status: completed
created: 2026-05-17
updated: 2026-05-18
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
tags: "workflow, split, sub-skill, progressive-disclosure, refactor, bundle-f"
---

# SPEC-V3R4-WORKFLOW-SPLIT-001 — Workflow Skills Phase-Scoped Sub-Skill Split

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-17 | manager-spec | 초기 plan-phase SPEC 작성 — Bundle F finding 대응, 4 monolithic workflow → 13 sub-skill + 4 entry router 분할 설계 |
| 0.1.1 | 2026-05-17 | orchestrator | plan-auditor R1 PASS (0.9125) 반영 — plan.md Wave 0에 T0.5 (slash command regression baseline trace 캡처 메커니즘) 추가하여 M1 mitigation 확정. status `draft → planned` 전환 |
| 0.2.0 | 2026-05-18 | orchestrator | Wave 1 (PR #973 run.md) + Wave 2 (PR #974 sync.md) + Wave 3 (PR #975 project.md) + Wave 4 (PR #976 plan.md self-referential) 4-Wave lifecycle 완주. AC-WFSP-001~008 모두 binary PASS. 4 entry router 모두 ≤200 LOC (run 99/sync 113/project 110/plan 144), 13 sub-skill 모두 ≤500 LOC. EC-1 (clarity-interview borderline) NOT triggered — 3 sub-skill outcome. workflow_split_test.go:133 t.Skip() 제거로 Path A→B 전환 완료 (AC-WFSP-002 4 router 전체 enforce). Template mirror 100% parity (dev-only exclusion 제외). SKILL.md byte-for-byte 무변경. Token-load reduction ~76% 달성 (4-workflow aggregate). status `in-progress → completed`. v2.20.0-rc1 release tag 발행 가능 상태 |

---

## Overview

Workflow Architecture Audit (2026-05-16, `.moai/research/workflow-audit-2026-05-16.md`) Bundle F finding을 해소한다. `.claude/skills/moai/workflows/` 의 4개 monolithic workflow skill이 progressive-disclosure level-2 token budget (~5000 tokens ≈ ~500 LOC)을 크게 초과한 상태로, agent context window를 과도하게 점유한다.

main HEAD `7a118e6b2` 기준 verified LOC:

| 파일 | LOC |
|------|-----|
| `.claude/skills/moai/workflows/run.md` | 1073 |
| `.claude/skills/moai/workflows/sync.md` | 1203 |
| `.claude/skills/moai/workflows/project.md` | 1076 |
| `.claude/skills/moai/workflows/plan.md` | 932 |
| **Total** | **4284** |

본 SPEC은 사용자 결정 사항 (4 AskUserQuestion rounds) 에 따라 다음 분할 전략을 적용한다:

- **Phase-scoped sub-skill 분리**: H2/H3 자연 경계를 따라 13개 sub-skill 생성
- **Entry router 보존**: 4개 원본 workflow 파일은 ≤200 LOC thin router로 축소 (backward compat)
- **4-Wave 순차 PR**: Wave 1=run, Wave 2=sync, Wave 3=project, Wave 4=plan (lessons #9 wave-split 적용)
- **Template mirror 동시 진행**: Wave별로 `internal/template/templates/.claude/skills/moai/workflows/` 동일 구조 동기화
- **SKILL.md Intent Router 무변경**: subcommand 매칭 surface는 entry skill에 의존하므로 변경 불필요
- **Harness level: thorough**: plan-auditor PASS + evaluator-active cross-validation 필수

---

## Goals

1. **Token-load reduction**: 각 workflow의 active context load를 ~10K tokens (run.md 기준) → ≤2500 tokens로 약 75% 감소
2. **Focused agent context**: phase별 sub-skill을 on-demand 로드하여 task-specific context 최적화 (level-2 → level-3 disclosure 활용도 증가)
3. **Maintenance scalability**: 새 phase 추가/수정 시 sub-skill 단위 격리 편집 가능 (현재는 1000+ LOC 단일 파일 수정 부담)
4. **Backward compatibility preservation**: 4개 slash command (`/moai plan|run|sync|project`)와 SKILL.md Intent Router는 완전히 무변경 (외부 reference 깨짐 0건)
5. **Template parity**: `moai init` 사용자 프로젝트에도 동일 sub-skill 구조 배포 (template mirror)

---

## Non-Goals

### Out of Scope (Exclusions — MIN 1 required)

- **SKILL.md Intent Router 수정 금지** — subcommand → entry skill 매칭 로직은 그대로. Router는 `"plan"|"run"|"sync"|"project"` 키만 매칭, sub-skill 직접 라우팅 없음
- **Slash command surface 변경 금지** — `.claude/commands/*.md` 11개 entry는 전혀 건드리지 않음. `/moai plan SPEC-X` 같은 invocation 패턴 100% 보존
- **Workflow execution semantics 변경 금지** — 각 phase의 실행 순서, 산출물, AskUserQuestion 호출 패턴, plan-auditor delegation, evaluator-active gate는 모두 그대로 유지. 순수 콘텐츠 재배치 only
- **Hot patch 금지** — Wave 1-4 모두 plan → run → sync 정상 lifecycle 경유. main 직접 patch 또는 worktree 임시 우회 금지
- **F-015 docs drift (UserPromptExpansion/PostToolBatch/mcp_tool)** — 본 SPEC 범위 밖, 별도 chore PR로 처리 (scenarios.md "Out of Scope" 명시)
- **SPEC-V3R4-LLM-REVIEW-CI-001** — CI runner LLM CLI missing 이슈는 별도 SPEC 후속, 본 SPEC과 무관
- **docs-site 4-locale sync (조건부)** — 사전 grep 결과 `docs-site/` 에 workflow skill 직접 reference 없음. 만약 run-phase에서 발견 시 `SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP` 별도 SPEC 생성

---

## EARS Requirements

### REQ-WFSP-001 — Sub-skill Boundary Discipline

**Ubiquitous**: 시스템은 각 sub-skill 파일이 다음 제약을 만족하도록 **shall** 강제한다:

- **REQ-WFSP-001a**: 각 sub-skill 파일의 LOC는 500 이하 **shall** 유지 (level-2 token budget 안전 마진)
- **REQ-WFSP-001b**: 각 sub-skill 파일은 valid YAML frontmatter를 **shall** 보유 (`name`, `description`, `metadata` 최소)
- **REQ-WFSP-001c**: 각 sub-skill의 `name` 필드는 `moai-workflow-{name}-{sub}` 패턴을 **shall** 따른다 (예: `moai-workflow-run-context-loading`)
- **REQ-WFSP-001d**: 각 sub-skill의 frontmatter는 `user-invocable: false`를 **shall** 명시 (사용자 직접 호출 차단)
- **REQ-WFSP-001e**: Phase 콘텐츠는 자연 H2/H3 경계를 따라 **shall** 분할 (artificial split 금지)

### REQ-WFSP-002 — Entry Router Skill

**Ubiquitous**: 시스템은 4개 entry router skill (`workflows/run.md`, `workflows/sync.md`, `workflows/project.md`, `workflows/plan.md`)이 다음 제약을 만족하도록 **shall** 강제한다:

- **REQ-WFSP-002a**: 각 entry router의 LOC는 200 이하 **shall** 유지
- **REQ-WFSP-002b**: Entry router는 frontmatter `user-invocable: true` (또는 미명시 — 기본값)를 **shall** 보유
- **REQ-WFSP-002c**: Entry router body는 sub-skill 라우팅 테이블 (phase → sub-skill path) 을 **shall** 포함
- **REQ-WFSP-002d**: Entry router는 phase 실행 순서를 **shall** 보존하며 각 phase에서 해당 sub-skill을 `Read`로 로드하는 패턴 사용
- **REQ-WFSP-002e**: **When** 사용자가 slash command를 호출하면, 시스템은 entry router를 첫 로드하고 phase progression 중 on-demand로 sub-skill을 **shall** 로드

### REQ-WFSP-003 — Backward Compatibility

**Ubiquitous**: 시스템은 다음 backward compatibility 제약을 **shall** 강제한다:

- **REQ-WFSP-003a**: `.claude/skills/moai/SKILL.md` Intent Router는 byte-for-byte 무변경 **shall** 유지
- **REQ-WFSP-003b**: 4개 slash command (`.claude/commands/01-plan.md`, `02-run.md`, `03-sync.md`, `04-project.md` 등) entry는 **shall** 무변경
- **REQ-WFSP-003c**: **When** `/moai plan|run|sync|project` 호출 시 phase execution trace는 split 전/후 동일 **shall** 보장 (regression test로 검증)
- **REQ-WFSP-003d**: **If** MEMORY entries 또는 lessons 내 cross-reference가 `workflows/{name}.md` 를 가리키면, **then** 시스템은 entry router 경로를 그대로 보존하여 reference 깨짐을 **shall** 방지
- **REQ-WFSP-003e**: docs-site `content/` 하위에 workflow skill 직접 reference가 없음을 plan-phase에서 확인 (사전 grep 결과 0건). **If** run-phase에서 발견되면, **then** Wave별 PR에 docs-site 4-locale sync 작업 **shall** 추가 또는 별도 follow-up SPEC 생성

### REQ-WFSP-004 — Template Synchronization

**Ubiquitous**: 시스템은 template mirror가 local copy와 1:1 동기화되도록 **shall** 강제한다:

- **REQ-WFSP-004a**: 각 Wave PR은 `internal/template/templates/.claude/skills/moai/workflows/` 하위에 동일한 sub-skill + entry router 구조를 **shall** 포함
- **REQ-WFSP-004b**: 각 Wave PR 종료 시 `make build` **shall** 실행하여 `internal/template/embedded.go` 재생성
- **REQ-WFSP-004c**: **When** `moai init <new-project>` 실행 시 사용자 프로젝트에 동일한 sub-skill 구조가 **shall** 배포
- **REQ-WFSP-004d**: Audit test는 `find internal/template/templates/.claude/skills/moai/workflows -name '*.md'` 결과와 local 동일 디렉토리 결과를 **shall** 비교하여 parity 검증

### REQ-WFSP-005 — Validation & Verification

**Ubiquitous**: 시스템은 다음 validation 제약을 **shall** 강제한다:

- **REQ-WFSP-005a**: `moai spec lint --strict` **shall** clean (`✓ No findings`)
- **REQ-WFSP-005b**: Cross-reference link checker (Go test or shell script) **shall** 모든 sub-skill 간 reference 경로를 검증, 깨진 link 0건
- **REQ-WFSP-005c**: Slash command regression test: `/moai plan SPEC-DUMMY`, `/moai run SPEC-DUMMY`, `/moai sync SPEC-DUMMY`, `/moai project` 각 1회 dry-run 실행, phase execution log를 split 전과 diff 비교 **shall** 통과
- **REQ-WFSP-005d**: LOC ceiling 자동 검증: Go test가 모든 sub-skill ≤500 LOC, entry router ≤200 LOC를 **shall** assertion
- **REQ-WFSP-005e**: golangci-lint **shall** clean (template embedded.go 재생성 부수 효과 확인)

---

## Affected Files

### MODIFY (4 files)

- `.claude/skills/moai/workflows/run.md` (1073 → ≤200 LOC, entry router 전환)
- `.claude/skills/moai/workflows/sync.md` (1203 → ≤200 LOC, entry router 전환)
- `.claude/skills/moai/workflows/project.md` (1076 → ≤200 LOC, entry router 전환)
- `.claude/skills/moai/workflows/plan.md` (932 → ≤200 LOC, entry router 전환)

### NEW (13 sub-skill files)

#### Wave 1 (run.md split)
- `.claude/skills/moai/workflows/run/context-loading.md` (~210 LOC)
- `.claude/skills/moai/workflows/run/phase-execution.md` (~500 LOC, splittable if exceeds)
- `.claude/skills/moai/workflows/run/mode-orchestration.md` (~120 LOC)

#### Wave 2 (sync.md split)
- `.claude/skills/moai/workflows/sync/quality-gates.md` (~500 LOC, splittable if exceeds)
- `.claude/skills/moai/workflows/sync/doc-execution.md` (~205 LOC)
- `.claude/skills/moai/workflows/sync/delivery.md` (~460 LOC)

#### Wave 3 (project.md split)
- `.claude/skills/moai/workflows/project/mode-detection.md` (~200 LOC)
- `.claude/skills/moai/workflows/project/codebase-analysis.md` (~115 LOC)
- `.claude/skills/moai/workflows/project/doc-generation.md` (~470 LOC)
- `.claude/skills/moai/workflows/project/meta-harness.md` (~300 LOC)

#### Wave 4 (plan.md split)
- `.claude/skills/moai/workflows/plan/context-discovery.md` (~140 LOC)
- `.claude/skills/moai/workflows/plan/clarity-interview.md` (~475 LOC)
- `.claude/skills/moai/workflows/plan/spec-assembly.md` (~315 LOC)

### NEW (template mirror — 17 files)

`internal/template/templates/.claude/skills/moai/workflows/` 하위 1:1 mirror:
- 4 entry router (run.md/sync.md/project.md/plan.md)
- 13 sub-skill (위와 동일 구조)

### NEW (audit/test files)

- `internal/skills/workflow_split_test.go` (LOC ceiling assertion + cross-ref validation)
- `scripts/audit-workflow-split.sh` (CI에서 호출 가능한 audit 스크립트)

### MODIFY (build artifact, auto-regenerated)

- `internal/template/embedded.go` (4 `make build` 실행으로 자동 재생성)

### Estimated total file count

- 4 modify + 13 new sub-skill + 17 template mirror + 2 test/script + 1 auto = **37 files**

---

## docs-site Impact Analysis

Pre-write grep 결과 (`grep -rln "workflows/run\|workflows/sync\|workflows/project\|workflows/plan" docs-site/`):

```
(empty result — 0 matches)
```

**결론**: docs-site 4-locale (ko/en/ja/zh) content 어디에도 `.claude/skills/moai/workflows/*.md` 직접 reference 없음. 따라서 본 SPEC은 §17 docs-site 4-locale sync 의무에서 면제.

**Run-phase 재검증 의무**: 각 Wave PR 종료 직전에 동일 grep 재실행. 만약 신규 reference가 v3.0.0 release-prep 중 추가되었으면 `SPEC-V3R4-WORKFLOW-SPLIT-001-DOCS-FOLLOWUP` 별도 SPEC 생성하여 처리 (scenarios.md "Out of Scope" 참조).

---

## Delta Markers

본 SPEC은 brownfield 작업 (기존 4 workflow 파일 분할). 다음 marker 적용:

- `[MODIFY]` — 4 workflow entry files (1000+ LOC monolithic → ≤200 LOC router)
- `[NEW]` — 13 sub-skill files (phase-scoped 콘텐츠 추출)
- `[NEW]` — 17 template mirror files (1:1 동기화)
- `[NEW]` — 2 audit/test files (regression 방지)
- `[AUTO]` — 1 build artifact (`embedded.go` regenerated by `make build`)

---

## References

- **Audit source**: `.moai/research/workflow-audit-2026-05-16.md` Bundle F finding
- **Frontmatter schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- **Progressive disclosure rule**: `.claude/rules/moai/development/skill-authoring.md` § Progressive Disclosure
- **Skill body budget**: Level-2 ~5000 tokens (~500 LOC narrative)
- **Wave split lesson**: lessons #9 (wave-split for SPECs with 30+ tasks)
- **BODP decision**: ¬a ¬b ¬c → main @ origin/main (signals pre-evaluated by orchestrator)
- **Template-First rule**: CLAUDE.local.md §2 Template-First Rule
- **Harness level**: `.moai/config/sections/harness.yaml` → thorough triggers plan-auditor PASS + evaluator-active

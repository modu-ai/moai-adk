---
id: SPEC-V3R6-DOCS-V3-README-001
title: "README v3 rewrite — reconcile en/ko against docs-truth canonical facts"
version: "0.2.1"
status: completed
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "README.md, README.ko.md"
lifecycle: spec-anchored
tags: "docs, readme, docs-truth, en-ko-sync, docs-v3-cohort"
era: V3R6
depends_on: [SPEC-V3R6-DOCS-CODEMAPS-V3-001]
related_specs: [SPEC-V3R6-DOCS-CODEMAPS-V3-001, SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001, SPEC-V3R6-WEB-CONSOLE-008]
---

# SPEC-V3R6-DOCS-V3-README-001 — README v3 rewrite

## §A. Overview

이 SPEC은 Sprint 14 Docs-v3 코호트의 2번째 SPEC으로, repo-root의 두 README 파일(`README.md` 1370L + `README.ko.md` 1418L)을 직전 종료된 `SPEC-V3R6-DOCS-CODEMAPS-V3-001`이 작성한 canonical facts checklist `.moai/project/codemaps/docs-truth.md` 기반으로 재작성한다.

**Hard scope boundary**:
- INCLUDE: `README.md` + `README.ko.md` (en/ko 2-locale 동기화)
- EXCLUDE: docs-site 4-locale content (`docs-site/content/{en,ko,ja,zh}/`) — 별도 DOCSITE-001 scope
- EXCLUDE: Go 코드 변경 (documentation-only SPEC, Go change LOC = 0 예상)

**Tier 제안**: **Tier M (standard)** — README 2개 파일(총 2788L)의 다축 사실 조정(reconciliation). Tier S(단일 파일/minor edit)보다 크고, Tier L(다컴포넌트 구조 변경/신파일 다수)보다는 작다.

## §B. Context

### §B.1 문제 배경

README en/ko의 사실표(facts table)가 codebase 실제 상태와 다수 drift 발생. 직전 CODEMAPS-V3-001이 codebase 정합 사실체크리스트 `docs-truth.md`를 작성하여 5개 사실 축(agent catalog, SPEC status enum, frontmatter schema, `/moai` command set, GLM tier-models)을 단일 탐색 보조 도구로 정리 완료 (commit 4a6f4b4d3에서 종료).

본 SPEC은 docs-truth.md가 가리키는 **1차 소스(primary source)**를 기준선으로 README의 stale claim들을 수정한다. docs-truth.md 자체가 새 SSOT가 아니라 navigation aid임에 주의 — 모든 사실 단언은 docs-truth.md가 인용하는 코드/디렉터리 1차 소스로 소급 검증된다.

### §B.2 Primary sources (re-verified at commit 4a6f4b4d3)

모든 1차 소스는 본 plan-phase에서 현 tree 기준으로 재검증 완료:

| Axis | docs-truth §  | Primary source | Re-verify result at 4a6f4b4d3 |
|------|---------------|----------------|-------------------------------|
| §1 Agent catalog | docs-truth §1 | `ls .claude/agents/moai/*.md` (= 7) + CLAUDE.md §4 + `archived-agent-rejection.md` | PASS — 7 MoAI-custom files 확인 |
| §2 SPEC status enum | docs-truth §2 | `internal/spec/status.go` `ValidStatuses` (lines 13-22) | PASS — 8 lowercase values grep count = 8 |
| §3 Frontmatter schema | docs-truth §3 | `internal/spec/lint.go` required slice (lines 590-601) | PASS — exactly 12 entries |
| §4.2 `/moai` command set | docs-truth §4.2 | `ls .claude/commands/moai/*.md` (= 17) | PASS — 17 files 확인 |
| §5 GLM tier-models | docs-truth §5 | `internal/config/defaults.go` lines 42-57 (`DefaultGLM*` 상수 블록) | PASS — `DefaultGLMHigh = "glm-5.2[1m]"` 확인 |
| §6 statusline preset retire | — | SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 outcome (completed) | PASS — README에 이미 반영 (L1282/L1351) |
| §7 statusline redesign | — | SPEC-V3R6-WEB-CONSOLE-008 honest-config outcome (completed) | PASS — README에 이미 반영 (statusline v3 multi-line) |

**재검증 결과: 모든 1차 소스가 docs-truth 기준선과 일치.** Drift는 1차 소스가 아니라 README의 stale claim 쪽에만 존재. blocker 없음.

## §C. Drift Inventory

아래 표는 README의 현 상태와 docs-truth 기준 correct value의 차이. "Current claim"은 README에 실제로 적힌 내용이고, "docs-truth correct value"는 1차 소스가 확인한 값이다.

### §C.1 §1 Agent catalog drift

**README.md L288-300 "Agent Categories"**:

| Field | README current claim | docs-truth correct value |
|-------|----------------------|--------------------------|
| Total agent count | "27 agents" (L298) | **8 retained agents** (7 MoAI-custom + 1 `Explore`) |
| Manager category | "8: spec, ddd, tdd, docs, quality, project, strategy, git" (L292) | **4: manager-spec, manager-develop, manager-docs, manager-git** |
| Expert category | "8: backend, frontend, security, devops, performance, debug, testing, refactoring" (L293) | **ARCHIVED** (12 agents 전부 archived, 카테고리 자체 제거) |
| Builder category | "3: agent, skill, plugin" (L294) | **1: builder-harness** |
| Evaluator category | "2: sync-auditor, plan-auditor" (L295) | **2: plan-auditor, sync-auditor** (순서만 정정) |
| Design System row | "4 (+ evaluator)" (L296) | 제거 (design system은 skill 카테고리로, agent 아님) |

**README.md L335-360 "Agent Model Assignment by Tier" (tier-mapping table)**: 10 archived agent 이름이 ACTIVE model-policy 행으로 등장 — `manager-strategy`, `manager-project`, `manager-quality`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-debug`, `expert-refactoring`, `expert-devops`, `expert-performance`. 전부 `.claude/rules/moai/workflow/archived-agent-rejection.md` archived list에 해당. 정정: retained 8종에 대한 tier-mapping 행만 유지 (manager-spec/develop/docs/git, plan-auditor, sync-auditor, builder-harness + `Explore`는 별도 Anthropic built-in 노트), archived 이름은 최대 1줄 "see archived-agent-rejection.md migration table" 참조로 대체 (active tier row 유지 금지).

**README.ko.md L334 "에이전트 카테고리"**: 동일 drift. 추가로 ko L342 카테고리 표의 "Agency | 6" 행(크리에이티브 프로덕션 파이프라인 — design system의 ko-locale 명칭으로, retained 8종에 해당하지 않는 stale 카테고리 행)이 en L296 "Design System | 4 (+evaluator)" 행의 ko 대응이나, §C.1 상단 표에 별도 열거 누락 (비대칭 열거 — Dn-1 정정으로 AC-2(b)에 콘텐츠 앵커 grep 추가). 5개 stale agent-count surface (D1 정정):
- L40 "24개 전문 AI 에이전트"
- L110 "26개 전문 AI 에이전트 + 47개 스킬"
- L308 "24개 전문 에이전트에게 작업을 위임" (참고: "AI" 누락 variant)
- L334 카테고리 표 자체
- L372 "24개 에이전트에 최적의 AI 모델을 할당" (또 다른 suffix variant)

**README.ko.md L380-410 "티어별 에이전트 모델 할당" (tier-mapping table)**: en L335-360과 동일 drift + `expert-testing` 1종 추가 (총 11 archived). 정정 방침은 en과 동일.

### §C.2 §4.2 `/moai` command set drift (47 → 17)

**README.md L302-321 "47 Skills (Progressive Disclosure)"**: skill 카테고리 분류표로 47개를 나열 (L306-320의 13-row count table 포함). 본 축은 `/moai` slash command set (17개)가 아니라 skill 카탈로그 count이므로, 본 SPEC의 17 정정은 **"47 Skills" 헤더와 관련 count 표 전체를 `/moai` 17-command set 라스팅으로 치환**하는 방향이다. (skill catalog count는 별도 audit이 필요하며 본 SPEC 범위를 벗어남 — 본 SPEC은 헤더 라벨의 오도성(misleading)을 제거하고 17개 `/moai` command set을 명시하는 것만 수행. 기존 13-row skill-category count 표는 stale 상태로 방치하지 않고 간결한 17-command `/moai` 라스팅으로 대체 — plan.md §F.3 decision iii.)

**README.ko.md L348 "47개 스킬" 헤더 + L350-368 count 표**: en과 거의 동일한 구조. L348 `### 47개 스킬 (프로그레시브 디스클로저)` 헤더 + L350-368의 13-row skill-category count table (Foundation 6 / Workflow 12 / Domain 4 / Format 1 / Platform 4 / Library 3 / Reference 5 / Tool 2 / Design 2 / Framework 1 / Legacy 5 / Docs 1 / Language Rules 16 — en L306-320의 near-mirror + ko-only "Legacy v2.x (retired)" 행). 추가로 L40 "52개 스킬", L110 "47개 스킬"로 내부 모순. en/ko 동기화 시一并 정정.

### §C.3 §5 GLM tier-models drift

**README.md L676, L682**:

| Field | README current claim | docs-truth correct value |
|-------|----------------------|--------------------------|
| Opus-tier GLM model | `GLM-5.1` (L682) | **`glm-5.2[1m]`** (1M context 활성화) |
| Models summary line | "GLM-5.1, GLM-4.7, GLM-4.5-Air, and free models" (L676) | "glm-5.2[1m], glm-4.7, glm-4.5-air, and free models" |
| Sonnet-tier | `GLM-4.7` (L683) | `glm-4.7` (PASS, 그대로) |
| Haiku-tier | `GLM-4.5-Air` (L684) | `glm-4.5-air` (PASS, 그대로, case만 표준화) |

**README.ko.md L722, L728**: 동일 drift (Opus `GLM-5.1` → `glm-5.2[1m]`).

### §C.4 §2 SPEC status enum — 정보성 검증

README는 SPEC status enum을 직접 나열하지 않음. 따라서 이 축은 "추가(add)"가 아니라 "README의 SPEC 워크플로 섹션이 8-value enum을 암시적으로라도 반영하는가" 검증만 수행. 현재 README L78-79, L445, L635 등은 `SPEC-XXX` placeholder만 언급하고 status value 자체는 나열 안 함 → **drift 없음, info-only axis**.

### §C.5 §3 Frontmatter schema — 정보성 검증

README에 frontmatter YAML 예제 없음 → drift 없음, info-only axis. AC는 "README가 spec.md 예제를 추가할 경우 12-field schema를 반영" 형태의 forward-looking check로만 둠.

### §C.6 §6/§7 statusline — 이미 반영됨

README.md L1282, README.ko.md L1351에 `preset` retire가 이미 명시. README.md L1242-1282에 statusline v3 multi-line도 반영. → **drift 없음**. 본 SPEC의 AC는 "이미 반영된 상태가 유지되는가" 보존 검증만 수행.

### §C.7 Drift count 요약

| 축 | Drift 항목 수 | 조치 유형 |
|----|---------------|-----------|
| §1 Agent catalog (categories) | 6 (en) + 5 (ko stale count surfaces: L40/L110/L308/L334/L372) | rewrite |
| §1 Agent catalog (tier-mapping tables) | 10 archived names (en L335-360) + 11 archived names (ko L380-410, `expert-testing` 추가) | rewrite |
| §4.2 `/moai` command set | 1 헤더 + 13-row count 표 (en L302-321), ko L348 헤더 + L350-368 13-row count 표 + L40 "52개" + L110 "47개" | rewrite |
| §5 GLM tier-models | 2 (en L676/L682) + 2 (ko L722/L728) | rewrite |
| §2 status enum | 0 | info-only verify |
| §3 frontmatter | 0 | info-only verify |
| §6 preset retire | 0 (이미 반영) | preservation check |
| §7 statusline redesign | 0 (이미 반영) | preservation check |
| **총 drift 항목** | **17 + tier-table 2축 + ko L308/L372 2 surface** | (rewrite 15+ / info-only verify 4 축) |

## §D. Requirements (GEARS)

### REQ-README-001 — Agent catalog 정합 (docs-truth §1)

The README files (`README.md` §"Agent Categories" 및 `README.ko.md` §"에이전트 카테고리") **shall** 기재하기를 정확히 **8 retained agents** (7 MoAI-custom + 1 Anthropic built-in `Explore`)를, archived agent 이름(`manager-strategy`, `manager-quality`, `manager-project`, `manager-brain`, `claude-code-guide`, `researcher`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`, `expert-testing`, `expert-debug` — `.claude/rules/moai/workflow/archived-agent-rejection.md` canonical list)은 **어떤 활성 카테고리나 활성 tier-mapping 행에도 나열하지 않고**, archived-name migration은 `archived-agent-rejection.md`로의 참조로만 처리할 것.

**Tier-mapping tables explicit scope**: 본 REQ는 "Agent Categories" 섹션뿐 아니라 **"Agent Model Assignment by Tier" / "티어별 에이전트 모델 할당" tier-mapping table** (README.md L335-360, README.ko.md L380-410)에도 동일하게 적용된다 — archived agent 이름이 active model-policy 행으로 등장하는 것은 금지. Retained 8종에 대한 tier-mapping 행만 유지하며, archived 이름은 최대 1줄 "see archived-agent-rejection.md migration table" 참조로 대체한다.

**Source**: docs-truth §1 (1차 소스: `ls .claude/agents/moai/*.md` = 7 + CLAUDE.md §4 + `archived-agent-rejection.md` canonical archived list).

**Corrects**: README.md L288-300 "27 agents" stale claim + L335-360 tier-mapping table 내 10 archived active 행; README.ko.md L334 카테고리 표 + L380-410 tier-mapping table 내 11 archived active 행 + L40/L110/L308/L372 stale count (총 5 ko stale agent-count surfaces).

### REQ-README-002 — `/moai` command set 정합 (docs-truth §4.2)

The README files **shall** 기재하기를 `.claude/commands/moai/` 디렉터리에 **17개 `/moai` slash command**가 존재함을 (brain/clean/codemaps/coverage/design/e2e/feedback/fix/gate/harness/loop/mx/plan/project/review/run/sync), 그리고 **"47 Skills" 헤더 라벨을 제거**할 것 (해당 숫자는 `/moai` command set이 아니라 skill catalog count로서 본 SPEC 범위를 벗어나므로, 오도성 제거가 본 SPEC의 목적).

**Source**: docs-truth §4.2 (1차 소스: `ls -1 .claude/commands/moai/*.md | wc -l` → 17).

**Corrects**: README.md L302 "### 47 Skills (Progressive Disclosure)" 헤더 + L306-320 count 표; README.ko.md L40 "52개 스킬" + L110 "47개 스킬" 내부 모순.

### REQ-README-003 — GLM tier-model 정합 (docs-truth §5)

The README files (`README.md` §"GLM" 표 및 `README.ko.md` 대응 섹션) **shall** 기재하기를 Opus-tier GLM model을 **`glm-5.2[1m]`**로 (1M context 활성화), Medium/Sonnet-tier를 `glm-4.7`로, Low/Haiku-tier를 `glm-4.5-air`로, 그리고 `[1m]` suffix가 Claude Code의 1M context mode를 활성화함을 부가 설명에 포함할 것.

**Source**: docs-truth §5 (1차 소스: `internal/config/defaults.go` lines 42-57, `DefaultGLMHigh = "glm-5.2[1m]"`).

**Corrects**: README.md L676 "GLM-5.1" + L682 Opus-tier `GLM-5.1`; README.ko.md L722 + L728 동일.

### REQ-README-004 — en/ko 동기화

**Where** 두 README 파일 모두 본 SPEC의 변경사항을 적용받는 범위 내에서, the README files **shall** 서로 동일한 사실 표(table) 구조와 동일한 사실 단언을 유지할 것 (en/ko locale 번역 차이는 허용하나, count/이름/모델명 등 사실값은 양쪽이 정확히 일치).

**Source**: `.moai/config/sections/language.yaml` (en/ko 동등 대우 원칙) + 본 SPEC의 §C drift inventory.

### REQ-README-005 — statusline 정합 보존 (preservation, docs-truth §6/§7)

**While** statusline v3 multi-line layout 및 `preset` retire가 이미 README에 반영된 상태이면 (README.md L1242-1282, README.ko.md L1309-1351), the README files **shall not** 본 SPEC 작업 중 해당 섹션을 stale 상태로 회귀시키거나 삭제하지 않을 것 (preservation requirement).

**Source**: SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001 outcome + SPEC-V3R6-WEB-CONSOLE-008 honest-config outcome.

### REQ-README-006 — 범위 경계 준수

The README files **shall not** docs-site 4-locale content (`docs-site/content/{en,ko,ja,zh}/`) 또는 Go 코드를 수정하지 않을 것. 본 SPEC은 documentation-only이며, Go change LOC는 0이어야 한다.

### Out of Scope — What NOT to Build

- **EXCLUDE**: docs-site 4-locale content 재작성 (`docs-site/content/{en,ko,ja,zh}/`) — 이것은 별도 DOCSITE-001 scope.
- **EXCLUDE**: Go 코드 변경 (어떤 `.go` 파일도 수정하지 않음).
- **EXCLUDE**: skill catalog count의 정확한 값 산출 ("47 Skills" 헤더를 제거하되, 정확한 skill 총수를 새로 산출하여 README에 추가하지는 않음 — 그것은 별도 skill-audit SPEC scope).
- **EXCLUDE**: CLAUDE.md 변경 (CLAUDE.md는 이미 직전 CODEMAPS-V3-001에서 정합 완료).
- **EXCLUDE**: spec status enum 8-value를 README에 명시적으로 추가 (현재 README가 암시적으로만 언급하므로, 본 SPEC은 drift 제거에만 집중).
- **EXCLUDE**: archived agent 12종의 이름을 README에 명시 (migration table은 `archived-agent-rejection.md`로의 참조로 처리).

## §F. Non-Functional Constraints

- **Doc-only / no Go change**: 본 SPEC은 documentation-only이며 Go change LOC = 0. Repo-root `README.md` / `README.ko.md`는 project-owned asset이지 template-distributed asset이 아님 (`find internal/template/templates -iname "README*"` → 0 results 확인됨). 따라서 `internal/template/templates/` template-neutrality CI guard scope는 본 SPEC과 무관하며, 해당 framing은 적용하지 않는다.
- **i18n**: en/ko 양쪽 동시 수정. 번역 품질은 ko가 en의 의미를 정확히 반영.
- **Backward compat**: README의 기존 섹션 구조(Quick Start / Features / Architecture 등)는 유지. 사실값만 정정.
- **Token efficiency**: README 길이를 증가시키지 않는 방향 (stale 섹션 제거로 감소 허용).

## §G. Acceptance Criteria Cross-Reference

| REQ | AC | Severity | Verification |
|-----|----|----------|--------------|
| REQ-README-001 (agent catalog — categories + tier tables) | AC-1 (en categories + tier table), AC-2 (ko categories + tier table + 5 stale count surfaces) | MUST | grep "27 agents" 부재 + grep "8 retained" 존재 + archived name 부재 (categories) + tier-mapping line-range grep (en L335-360, ko L380-410) |
| REQ-README-002 (command set) | AC-3 (en + ko 양쪽 "47 Skills" / "47개 스킬" 헤더 부재) | MUST | grep "47 Skills\|47개 스킬" 부재 (en+ko) + 17-command 명시 존재 |
| REQ-README-003 (GLM tier) | AC-4 (en), AC-5 (ko) | MUST | grep "glm-5.2\[1m\]" 존재 + grep "GLM-5.1" 부재 (GLM context) |
| REQ-README-004 (en/ko sync) | AC-6 | MUST | 양쪽 사실값 일치 diff 검증 |
| REQ-README-005 (statusline 보존) | AC-7 | MUST | grep "preset" retire 문구 보존 |
| REQ-README-006 (scope boundary) | AC-8 | MUST | `git diff --stat`에 `.go` 파일 부재 |

상세 Given-When-Then 시나리오는 `acceptance.md`에 정의.

## §H. Cross-References

- `.moai/project/codemaps/docs-truth.md` — canonical facts checklist (직전 CODEMAPS-V3-001 작성)
- `.moai/specs/SPEC-V3R6-DOCS-CODEMAPS-V3-001/` — 선행 SPEC (origin/main 4a6f4b4d3에서 종료)
- `.claude/rules/moai/workflow/archived-agent-rejection.md` — archived agent migration table
- `internal/spec/status.go` — §2 1차 소스
- `internal/spec/lint.go` — §3 1차 소스
- `internal/config/defaults.go` — §5 1차 소스
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter schema SSOT

---

## HISTORY

- 2026-06-17: plan-phase artifacts authored (spec.md + plan.md + acceptance.md + progress.md). Tier M 제안. docs-truth.md 5개 축 + statusline 2개 축 전수 1차 소스 재검증 PASS. drift inventory 17항 (rewrite 13 + info-only 4축).
- 2026-06-17 (iter-2, v0.2.0): plan-auditor PASS-WITH-DEBT 0.83 → 3 defects + 1 optional + neutrality over-specification 정정. D1 (ko stale agent-count L308/L372 2면 추가 + AC-2 regex broadening), D2 (tier-mapping table archived-name active 행 — REQ-README-001 scope 확장으로 처리, REQ-007 신규 발행 안 함), D3 (ko L348 "47개 스킬" 헤더 + L350-368 count 표 실존 인정 + AC-3 양 locale 커버), D4 (plan §F.3 13-row count 표 disposition 명시 — 17-command 라스팅으로 치환), neutrality (repo-root README는 project-owned not template-distributed — template-neutrality framing 경량화).
- 2026-06-17 (iter-3, v0.2.1): plan-auditor iter-2에서 추가 지적된 2 MINOR defect surgical patch. Dn-1 (Testability — stale category-row drift 미커버): AC-1(b2) 추가 (en L296 "Design System** | 4 (+ evaluator)" 콘텐츠 앵커 grep — "(+ evaluator)"로 L318 skill-category 행과 구분), AC-2(b2) 추가 (ko L342 "**Agency** | 6" 콘텐츠 앵커 grep), spec §C.1 ko "Agency | 6" 행 비대칭 열거 정정. Dn-2 (Edge-1b ko fallback regex typo): Edge-1b ko branch `할당` → `(할당|배정)` broadening (README.ko.md L382 실제 헤더 "티어별 에이전트 모델 배정" 반영, fallback no-match 상실 방지). scope 외 미변경 (REQs, plan.md milestones, design.md, research.md untouched).

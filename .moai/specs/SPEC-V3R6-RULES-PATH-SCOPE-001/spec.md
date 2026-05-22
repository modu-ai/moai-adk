---
id: SPEC-V3R6-RULES-PATH-SCOPE-001
title: "Always-Loaded Rule 9→5 — Path-Scoped 전환 4건 (zone-registry / design constitution / manager-develop-prompt-template / agent-teams-pattern)"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "rules, path-scope, token-economy, wave-1, v3r6, always-loaded-reduction"
tier: M
depends_on: [SPEC-V3R6-HOOK-CONTRACT-FIX-001, SPEC-V3R6-DOCS-USER-DRIFT-001]
related_specs: [SPEC-V3R6-RULES-COMPRESS-001, SPEC-V3R6-SKILL-CONSOLIDATE-001, SPEC-V3R6-SKILL-COMPRESS-001]
---

# SPEC-V3R6-RULES-PATH-SCOPE-001 — Always-Loaded Rule 9→5 Path-Scoped 전환

## HISTORY

- v0.2.0 (2026-05-22): run-phase COMPLETE. M1-M5 모두 1-pass. 18/18 ACs PASS (`progress.md` 참조). manager-develop cycle_type=ddd orchestrator-direct (Agent worktree hook regression precedent per V3R6 25ee73039). 8 files modified (4 local rule frontmatter prepend + 4 template mirror byte-identical sync). Doctor simulation report `.moai/reports/rules-path-scope-simulation-2026-05-22.md` 작성 (5 scenarios, trigger miss 0, spurious load 0). Pre-existing baseline test failures preserved (zone-registry 4 diff lines + manager-develop-prompt-template +1316B → post-SPEC drift modulo, §1.4.1 documented, AC-RPS-007 Group B PASS). Cross-platform build PASS (linux+windows exit 0). C-HRA-008 0 matches (baseline). Always-loaded word count saving = -7,831 words (target: -7,500). 운영 발견: SPEC plan-phase artifacts (spec/plan/acceptance .md) 가 working tree 에서 사라졌다 git HEAD 421e3672d 로부터 복원 필요 (parallel session 또는 hook side-effect 원인 미상, 별도 lesson 후보). status: draft → implemented, version: 0.1.0 → 0.2.0.
- v0.1.0 (2026-05-22): 초기 draft. Wave 1 Lane A 첫 SPEC. v3.0 환골탈태 (`.moai/research/v3.0-design-2026-05-22.md` §Layer 1 라인 165-184) 첫 번째 토큰 감축 SPEC. Wave 0 두 SPEC (`SPEC-V3R6-HOOK-CONTRACT-FIX-001` commit `48340e726` MERGED + `SPEC-V3R6-DOCS-USER-DRIFT-001` commit `d386cca0e` MERGED) 머지 직후 진입. 사용자 §8 결정 4건 반영 (6/15 이전 완료 / Wave 0+1+2 토큰 감축 우선 / default profile / GLM 보수). 본 SPEC은 4 rule 파일에 `paths:` frontmatter 추가만으로 always-loaded 부담을 -23.5K tokens (~62%) 감축. **Go 코드 변경 0건** — Claude Code 런타임이 frontmatter `paths:` 를 이미 직접 해석 (선례: skill-authoring, agent-patterns, spec-frontmatter-schema 등 14+ 기존 사례). `internal/rules/loader.go` 신규 작성 불필요 — 위임 prompt §Pre-flight 결과 ("internal/rules/ MISSING") 가 design.md §Layer 1 "가칭" 표현을 확정 정정.

---

## 1. Overview

### 1.1 Goal

MoAI-ADK 의 always-loaded rule 부담을 9개 (~40,500w / ~121K tokens) 에서 5개 (~15,000w / ~45K tokens) 로 감축. 4개 rule (`zone-registry.md` / `design/constitution.md` / `development/manager-develop-prompt-template.md` / `workflow/agent-teams-pattern.md`) 에 `paths:` frontmatter 추가하여 **path-scoped 조건부 로드** 로 전환. 모든 session 시작 시 자동 로드되던 7,831 단어 (~23.5K tokens) 가 path 매치 시에만 로드되어 token economy 개선.

### 1.2 Motivation

- **현재 부담**: session 시작 시 `.claude/rules/moai/**/*.md` 중 frontmatter 없는 파일 9개가 무조건 로드 (CLAUDE.md auto-discovery 메커니즘). 9개 합산 ~40,500w / ~121K tokens (전체 Opus 4.7 1M 윈도 대비 12.1%, 200K 윈도 대비 60%).
- **불필요한 비용**: `zone-registry.md` (3,453w / ~10.4K) 는 HARD 조항 SSOT 으로서 `.claude/rules` 또는 `.moai/specs` 수정 시에만 필요. 일반 코드 작업 (Go/Python/TypeScript) session 에는 불필요. `design/constitution.md` (2,472w / ~7.4K) 는 GAN Loop / brand design pipeline 활성 시에만 필요. `manager-develop-prompt-template.md` (1,136w / ~3.4K) 는 run-phase manager-develop 위임 시에만 필요. `agent-teams-pattern.md` (770w / ~2.3K) 는 team mode 활성 시에만 필요.
- **선례 검증**: 14+ 규칙이 이미 `paths:` frontmatter 로 path-scoped 로드 중 (검증된 인프라). 예: `skill-authoring.md` (`paths: "**/.claude/skills/**"`), `agent-patterns.md` (`paths: ".claude/agents/**/*.md,.claude/rules/moai/development/agent-authoring.md"`), `spec-frontmatter-schema.md` (`paths: "**/*.md,.moai/specs/**/*.md"`). Claude Code 런타임이 frontmatter `paths:` 를 이미 직접 해석 — 신규 loader 코드 작성 불필요.
- **v3.0 청사진 §Layer 1**: 9 → 5 always 전환 + 4 path-scoped + (별도 SPEC `SPEC-V3R6-RULES-COMPRESS-001`) 3 압축으로 **-25,500w 절감 (40,500 → 15,000, 62% off)** 목표. 본 SPEC은 그 중 path-scoped 4건 (-23,500w / -23.5K tokens) 담당.

### 1.3 Scope (이번 SPEC이 다루는 것)

다음 4 파일에 `paths:` frontmatter (+ 보조 `description:`) 추가:

| Rule 파일 | 현재 부담 | 신규 `paths:` 글롭 | 발동 조건 |
|---|---|---|---|
| `.claude/rules/moai/core/zone-registry.md` | 3,453w / ~10.4K | `.claude/**,.moai/specs/**,.claude/rules/**` | rules / agents / skills / SPEC 디렉토리 수정 시 |
| `.claude/rules/moai/design/constitution.md` | 2,472w / ~7.4K | `.moai/design/**,.moai/specs/SPEC-*-DESIGN-*/**,.moai/project/brand/**,.claude/skills/moai/**/design*.md,.claude/skills/moai/**/brand*.md` | design pipeline / brand 작업 시 |
| `.claude/rules/moai/development/manager-develop-prompt-template.md` | 1,136w / ~3.4K | `.moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md` | run-phase 또는 SPEC 위임 작성 시 |
| `.claude/rules/moai/workflow/agent-teams-pattern.md` | 770w / ~2.3K | `.moai/config/sections/workflow.yaml,.claude/agents/moai/manager-strategy.md,.claude/skills/moai/team/**` | team mode 설정 / 활성 시 |
| **합계 감축** | **7,831w / ~23.5K** | | |

추가 의무 작업:
- 4 rule 파일 각각 `paths:` 외에 `description:` 1줄 frontmatter (other 14 규칙 표준 형식 일치)
- Trigger miss 위험 완화: 본 SPEC body 에 §6 위험 평가 항목 + run-phase implementation 단계에서 doctor 시뮬레이션 의무
- v3.0 design.md §Layer 1 표 값 (recommendation `path-scoped`) → 실제 frontmatter 값 일치 검증
- Template 동기화 의무 (CLAUDE.local.md §2 [HARD] Template-First Rule): `internal/template/templates/.claude/rules/moai/{core,design,development,workflow}/` 4 mirror 파일 동시 갱신

### 1.4 Pre-flight 검증 결과 (사실 검증 완료)

| 항목 | 사실 | 검증 명령 |
|---|---|---|
| 4 target rule 존재 | 4 파일 모두 존재 (zone-registry 35,969B / design constitution 18,757B / manager-develop-prompt 8,250B / agent-teams-pattern 6,119B) | `ls -la .claude/rules/moai/core/zone-registry.md .claude/rules/moai/design/constitution.md .claude/rules/moai/development/manager-develop-prompt-template.md .claude/rules/moai/workflow/agent-teams-pattern.md` |
| 4 target 모두 zero-frontmatter | 모두 `# <Title>` 으로 시작 (YAML frontmatter 부재) | `head -1` × 4 |
| paths frontmatter 선례 14+ 건 | `agent-hooks.md` / `moai-constitution.md` / `settings-management.md` / `boundary-verification.md` / `hooks-system.md` / `spec-frontmatter-schema.md` / `skill-authoring.md` / `karpathy-quickref.md` / `skill-ab-testing.md` / `agent-patterns.md` / `skill-writing-craft.md` / `model-policy.md` / `agent-authoring.md` 등 | `grep -rn '^paths:' .claude/rules/moai/` |
| internal/rules/ 부재 | `internal/rules/` 디렉토리 미존재. `internal/config/loader.go` 만 존재 (config YAML 로더, rule 로더 아님) | `ls internal/rules/ 2>/dev/null \|\| echo MISSING` |
| Claude Code 런타임 paths 해석 | 14+ 기존 path-scoped 규칙이 이미 정상 동작 중 — 런타임 메커니즘 검증됨 | 위 선례 14+ 건 + `agent-hooks.md` `paths: "**/.claude/agents/**,**/.claude/hooks/**"` 실제 작동 |
| 단어 카운트 baseline | zone-registry 3,453 / design constitution 2,472 / manager-develop-prompt-template 1,136 / agent-teams-pattern 770 | `wc -w` × 4 |
| session-handoff.md 미포함 사유 | 본문 라인 5 footnote: "intentionally always-loaded ... Trigger #3 (user explicit session-end request) can fire from any session context, including those without SPEC files. The ~1,400-token cost is justified by the protocol's cross-cutting applicability." | `head -10 .claude/rules/moai/workflow/session-handoff.md` |
| Out of scope 5 rules | `agent-common-protocol.md` (1,927w) / `session-handoff.md` (1,927w) / `context-window-management.md` (712w) / `verification-batch-pattern.md` (764w) / `NOTICE.md` (349w) | `wc -w` 표 |
| Template 동기화 의무 확인 | `internal/template/templates/.claude/rules/moai/workflow/session-handoff.md` 존재 (template mirror 운영 중) | `find internal/template/templates/.claude/rules -name '*.md' \| head` |
| **Pre-existing template drift baseline (Critical)** | **2/4 target files drift 확인** (2026-05-22 측정): zone-registry.md = 4 diff lines (byte-equal 35,969B but content differ), manager-develop-prompt-template.md = 20 diff lines (local 8,250B vs template 6,934B, +1,316B Tier S Applicability section); 2 files byte-identical (design/constitution.md 18,757B, agent-teams-pattern.md 6,119B). 본 SPEC 는 frontmatter prepend 외 drift 자체는 out of scope — §1.4.1 참조 | `for sub in core/zone-registry.md design/constitution.md development/manager-develop-prompt-template.md workflow/agent-teams-pattern.md; do diff -q ".claude/rules/moai/$sub" "internal/template/templates/.claude/rules/moai/$sub"; done` |

### 1.4.1 Pre-existing Template Mirror Drift Baseline (BLOCKING B1 mitigation)

**발견 시점**: plan-auditor iter 1 BLOCKING 지적 (2026-05-22). pre-existing drift 가 본 SPEC AC-RPS-007 byte-identical sync 의무와 직접 충돌.

**Drift 정량화** (Pre-flight `diff` 실측 2026-05-22):

| Rule | Local | Template | Drift 종류 | 본 SPEC scope |
|---|---|---|---|---|
| `core/zone-registry.md` | 35,969B | 35,969B | content differ (4 diff lines) — Operational threshold 표 비대칭 추정 (이전 V3R6 정정 단방향) | **drift modulo** — frontmatter prepend 만 적용, baseline drift 는 별도 `SPEC-V3R6-TEMPLATE-MIRROR-DRIFT-001` (V3R6 lesson #25) 흡수 |
| `design/constitution.md` | 18,757B | 18,757B | byte-identical | drift 없음 — AC-RPS-007 strict byte-identical 검증 |
| `development/manager-develop-prompt-template.md` | 8,250B | 6,934B | template lacks ~1,316B (Tier S Applicability section, SPEC-V3R5-WORKFLOW-LEAN-001 후속 추가분) | **drift modulo** — frontmatter prepend 만 적용, baseline drift 는 별도 sync SPEC 에서 흡수 |
| `workflow/agent-teams-pattern.md` | 6,119B | 6,119B | byte-identical | drift 없음 — AC-RPS-007 strict byte-identical 검증 |

**본 SPEC 변경분 vs baseline drift 분리 원칙**:
- AC-RPS-007 byte-identical sync 의무는 **frontmatter prepend 후 baseline drift modulo** 로 해석.
- 2 byte-identical file (design constitution, agent-teams-pattern) 은 AC-RPS-007 strict byte-identical 검증 (post-SPEC diff empty).
- 2 drift file (zone-registry, manager-develop-prompt-template) 는 AC-RPS-007 conditional 검증 — frontmatter prepend 후 `diff` 결과가 pre-SPEC baseline drift 와 동일 (drift 가 본 SPEC 으로 인해 확대되지 않음) → PASS.

**구현 의무**: run-phase manager-develop 는 frontmatter prepend 전 본 §1.4.1 표를 progress.md 에 baseline snapshot 으로 기록. AC-RPS-007 verification 결과 보고 시 drift modulo 명시 (예: "zone-registry.md: 4 diff lines preserved from baseline, no new drift introduced").

### 1.5 Constitution 정렬

- **CLAUDE.md §1 HARD Rules**: "Multi-File Decomposition: Split work when modifying 3+ files" 정렬 — 본 SPEC는 4 rule + 4 template mirror = 8 파일 수정으로 plan.md §3 Implementation Order 에서 명시적 분할.
- **CLAUDE.local.md §2 [HARD] Template-First Rule**: 4 rule 변경은 `internal/template/templates/.claude/rules/moai/` 4 mirror 파일에도 동시 적용 의무. `make build` 후 임베디드 갱신.
- **CLAUDE.local.md §15 [HARD] 16-language 중립성**: 본 SPEC은 영어 instruction file (`.md`) 만 수정, 16개 언어 중 어떤 것도 우선시하지 않음 — 위반 없음.
- **`.claude/rules/moai/development/coding-standards.md` §Paths Frontmatter**: 본 SPEC가 정확히 이 표준의 활용 사례. 표준 본문 인용: "Use paths frontmatter for conditional rule loading: ... This ensures rules load only when working with matching files."
- **`.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT**: 본 SPEC 자체 spec.md 는 12-field canonical (created/updated/tags) 준수. **rule 파일의 frontmatter** 는 SPEC 의 frontmatter 와 다른 schema (paths + description) — 무관.

---

## 2. User Stories

### US-RPS-001: 일반 Go 작업 session 에서 design constitution 불필요 로드 회피

> **As a** MoAI-ADK 사용자 (Go 코드 작업 중)
> **I want** session 시작 시 design/constitution.md (2,472w / ~7.4K tokens) 가 자동 로드되지 않기를
> **So that** GAN Loop / brand design 작업이 없는 일반 Go session 에서 ~7.4K tokens 를 절약하고 200K 컨텍스트 윈도에서 더 많은 코드 변경을 수용한다

**Acceptance**: `design/constitution.md` 첫 줄이 `---` YAML frontmatter 시작, 본문 라인에 `paths: ".moai/design/**,.moai/specs/SPEC-*-DESIGN-*/**,.moai/project/brand/**,.claude/skills/moai/**/design*.md,.claude/skills/moai/**/brand*.md"` 정확히 일치.

### US-RPS-002: 일반 SPEC 작성 session 에서 zone-registry 불필요 로드 회피

> **As a** MoAI-ADK 사용자 (앱 코드 또는 README 작업 중)
> **I want** `zone-registry.md` (3,453w / ~10.4K tokens) 가 `.claude/rules/` / agents / skills / `.moai/specs/` 디렉토리 외부 작업 session 에서 자동 로드되지 않기를
> **So that** HARD 조항 SSOT 가 필요한 작업 (rule/agent/skill/SPEC 수정) 에만 로드되어 일반 코드 작업 token budget 을 보호한다

**Acceptance**: `zone-registry.md` frontmatter 에 `paths: ".claude/**,.moai/specs/**,.claude/rules/**"` 존재.

### US-RPS-003: team mode 미사용 session 에서 agent-teams-pattern 불필요 로드 회피

> **As a** MoAI-ADK 사용자 (solo mode 작업 중)
> **I want** `agent-teams-pattern.md` (770w / ~2.3K tokens) 가 team 활성 작업 외에서 로드되지 않기를
> **So that** `workflow.yaml`/`manager-strategy.md`/`team/*.md` 수정 시에만 발동되어 단순 solo session 의 token 부담을 줄인다

**Acceptance**: `agent-teams-pattern.md` frontmatter 에 `paths: ".moai/config/sections/workflow.yaml,.claude/agents/moai/manager-strategy.md,.claude/skills/moai/team/**"` 존재.

### US-RPS-004: SPEC 위임 시 manager-develop-prompt-template 자동 로드 유지

> **As a** MoAI-ADK orchestrator
> **I want** `manager-develop-prompt-template.md` (1,136w / ~3.4K tokens) 가 SPEC 위임 prompt 작성 시 자동 로드되지만 다른 session 에는 로드되지 않기를
> **So that** run-phase Tier M/L 위임 5-section template 의무를 만족하면서 일반 session token 부담을 회피한다

**Acceptance**: `manager-develop-prompt-template.md` frontmatter 에 `paths: ".moai/specs/**,.claude/agents/moai/manager-develop.md,.claude/skills/moai/workflows/run.md"` 존재.

### US-RPS-005: trigger miss 시 사용자가 즉시 감지

> **As a** MoAI-ADK 사용자 또는 메인테이너
> **I want** path glob 이 너무 좁아 필요한 rule 이 로드되지 않는 경우 (trigger miss) 를 doctor 명령으로 즉시 감지하기를
> **So that** silent failure (위반 미감지) 대신 명시적 보고로 path glob 보정이 가능하다

**Acceptance**: plan.md §3 Implementation Order 의 M3 doctor 검증 단계에서 4 rule 의 path glob 시뮬레이션 결과가 명시. run-phase 종료 후 보고 의무.

---

## 3. EARS Requirements

### 3.1 Functional Requirements

**REQ-RPS-001 (Ubiquitous — zone-registry path-scoped)**:
The MoAI-ADK rule loading system **shall** load `.claude/rules/moai/core/zone-registry.md` only when the active session involves files matching `.claude/**` OR `.moai/specs/**` OR `.claude/rules/**`.

**REQ-RPS-002 (Ubiquitous — design constitution path-scoped)**:
The MoAI-ADK rule loading system **shall** load `.claude/rules/moai/design/constitution.md` only when the active session involves files matching `.moai/design/**` OR `.moai/specs/SPEC-*-DESIGN-*/**` OR `.moai/project/brand/**` OR `.claude/skills/moai/**/design*.md` OR `.claude/skills/moai/**/brand*.md`.

**REQ-RPS-003 (Ubiquitous — manager-develop-prompt-template path-scoped)**:
The MoAI-ADK rule loading system **shall** load `.claude/rules/moai/development/manager-develop-prompt-template.md` only when the active session involves files matching `.moai/specs/**` OR `.claude/agents/moai/manager-develop.md` OR `.claude/skills/moai/workflows/run.md`.

**REQ-RPS-004 (Ubiquitous — agent-teams-pattern path-scoped)**:
The MoAI-ADK rule loading system **shall** load `.claude/rules/moai/workflow/agent-teams-pattern.md` only when the active session involves files matching `.moai/config/sections/workflow.yaml` OR `.claude/agents/moai/manager-strategy.md` OR `.claude/skills/moai/team/**`.

**REQ-RPS-005 (Event-Driven — frontmatter 추가)**:
**WHEN** the implementer applies path-scoped transformations, **THE** SPEC implementation **shall** prepend a YAML frontmatter block containing `description:` (one-line) and `paths:` (CSV string per `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter) to each of the 4 target rule files, while preserving the existing body content byte-for-byte (except prepending).

**REQ-RPS-006 (Event-Driven — template mirror sync)**:
**WHEN** any of the 4 target rule files is modified, **THE** SPEC implementation **shall** apply byte-identical changes to the corresponding mirror file at `internal/template/templates/.claude/rules/moai/<sub-path>/<filename>` to satisfy CLAUDE.local.md §2 [HARD] Template-First Rule.

**REQ-RPS-007 (State-Driven — Out of scope keep-always)**:
**WHILE** the 5 rules listed in §1.3 Out of scope (`agent-common-protocol.md`, `session-handoff.md`, `context-window-management.md`, `verification-batch-pattern.md`, `NOTICE.md`) remain in scope as keep-always rules, **THE** SPEC implementation **shall not** add `paths:` frontmatter to these 5 files.

**REQ-RPS-008 (Optional — coding-standards 인용)**:
**WHERE** `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter section exists, **THE** rule frontmatter format **shall** use CSV string syntax (`paths: "a,b,c"`) matching the canonical example in coding-standards.md.

**REQ-RPS-009 (Optional — description 표준화)**:
**WHERE** existing path-scoped rules (e.g., `spec-frontmatter-schema.md`, `agent-patterns.md`) include a `description:` frontmatter line, **THE** 4 newly-converted rule files **should** include an analogous one-line `description:` field for consistency with the established convention.

### 3.2 Non-Functional Requirements

**REQ-RPS-NF-010 (Token economy)**:
The cumulative always-loaded rule word count **shall** decrease by at least 7,500 words (from current baseline ~40,500w including 9 always-loaded rules to ~33,000w with 4 rules path-scoped). Measurement: `wc -w` sum over remaining 5 always-loaded rules + their `description:` overhead, compared to current 9-rule sum.

**REQ-RPS-NF-011 (No Go code change)**:
The SPEC implementation **shall not** introduce any new Go source file, modify `internal/config/loader.go`, or create any rule-loading mechanism. The Claude Code runtime's existing `paths:` frontmatter handler (verified by 14+ pre-existing path-scoped rules) is the sole loading mechanism.

**REQ-RPS-NF-012 (Trigger miss detection)**:
The SPEC implementation **shall** include a doctor-level verification step in run-phase (plan.md §3 M3) that simulates rule loading for representative session types (Go-only / SPEC-only / design / team / general) and reports which of the 4 path-scoped rules would load in each scenario. Output stored at `.moai/reports/rules-path-scope-simulation-<DATE>.md` (local artifact, gitignored).

**REQ-RPS-NF-013 (Backward compat)**:
The SPEC implementation **shall** preserve all current rule loading behavior for the 5 keep-always rules (REQ-RPS-007). No existing rule loads slower, fails to load, or duplicate-loads as a side effect.

**REQ-RPS-NF-014 (Body byte preservation)**:
For each of the 4 target rule files, the SPEC implementation **shall** ensure that the byte sequence following the new closing `---` frontmatter delimiter is byte-identical to the original file's full content (no body edits, no whitespace normalization, no line ending changes). Verification: `tail -n +<frontmatter-end+1>` output diff against original is empty.

### 3.3 Out of Scope

#### 3.3.1 Out of Scope (다른 SPEC 또는 향후 작업)

본 SPEC가 명시적으로 다루지 않는 것 — 별도 SPEC 또는 향후 시점에서 처리:

| 항목 | 처리 위치 |
|---|---|
| **5 keep-always rules 압축** (`agent-common-protocol.md` 1,927w + `session-handoff.md` 1,927w + `context-window-management.md` 712w + `verification-batch-pattern.md` 764w + `NOTICE.md` 349w) | 별도 SPEC `SPEC-V3R6-RULES-COMPRESS-001` (Wave 1 Lane A, Tier S) |
| **Skill 통폐합** (3 그룹 통합: ci-watch+ci-autofix→ci-loop / design-import+design-context→design / harness × 3→harness-patterns) | 별도 SPEC `SPEC-V3R6-SKILL-CONSOLIDATE-001` (Wave 1 Lane A, Tier M) |
| **Skill body 압축** (상위 5개 무거운 workflow skill: testing / spec / project / design-handoff / meta-harness 평균 -30% 압축) | 별도 SPEC `SPEC-V3R6-SKILL-COMPRESS-001` (Wave 1 Lane A, Tier M) |
| **Agent 모델 명시 라우팅** (19 agents `inherit` → 명시적 opus/sonnet/haiku 분배) | Wave 2 SPEC `SPEC-V3R6-AGENT-MODEL-ROUTING-001` (가칭) |
| **Hook 효율화** (Observe Hook opt-in, async 확대) | Wave 2 SPEC `SPEC-V3R6-HOOK-EFFICIENCY-001` (가칭) |
| **Documentation Drift Prevention** (docs-site refresh CI) | Wave 5 SPEC `SPEC-V3R6-DOCS-DRIFT-PREVENT-001` (가칭) |

#### 3.3.2 Out of Scope (본 SPEC가 절대 건드리지 않는 영역)

| 항목 | 사유 |
|---|---|
| **`session-handoff.md` path-scoped 화** | 본문 line 5 footnote 명시: "intentionally always-loaded ... Trigger #3 (user explicit session-end request) can fire from any session context, including those without SPEC files. The ~1,400-token cost is justified by the protocol's cross-cutting applicability." — design.md §Layer 1 "압축 대상" 분류, 본 SPEC 범위 외 (RULES-COMPRESS-001 담당) |
| **`agent-common-protocol.md` path-scoped 화** | design.md §Layer 1 "keep always (모든 agent 호출에 필요)" 명시. 모든 agent 호출 의무 reference — path-scoped 화 시 trigger miss High 위험 |
| **`NOTICE.md` path-scoped 화** | 라이선스 의무 — Apache 2.0 attribution 항상 가시화 의무 |
| **Go 코드 변경** | Claude Code 런타임 기존 메커니즘 (frontmatter `paths:`) 사용. `internal/rules/loader.go` 신규 작성 불필요 (위임 prompt §Pre-flight 결과 확정) |
| **Rule body 내용 수정** | 본 SPEC는 frontmatter prepend 만 수행. body byte preservation 의무 (REQ-RPS-NF-014) |
| **9 → 5 외 추가 rule path-scoped 화** | design.md §Layer 1 표 4건만 path-scoped 대상으로 명시. 추가 후보는 향후 분석 후 별도 SPEC |
| **doctor 명령 CLI 신규 sub-command** | REQ-RPS-NF-012 의 "doctor-level verification" 는 manual 시뮬레이션 (Bash + grep 조합) 으로 충분. 신규 `moai doctor --simulate-paths` 추가는 본 SPEC 범위 외 |

---

## 4. Stakeholders

| 역할 | 이름 | 책임 |
|---|---|---|
| **사용자 (메인테이너)** | GOOS Kim | v3.0 환골탈태 §8 4 결정 (6/15 이전 / Wave 0+1+2 토큰 감축 우선 / default profile / GLM 보수) 발의자. 머지 승인 |
| **manager-spec** | (본 위임) | spec.md / plan.md / acceptance.md 작성. plan-auditor 별도 dispatch |
| **plan-auditor** | (orchestrator 별도 dispatch) | Tier M PASS threshold 0.80 검증 |
| **manager-develop** | run-phase 위임 | M1 (rule frontmatter 4건) → M2 (template mirror 4건) → M3 (doctor 시뮬레이션) → M4 (회귀 테스트) |
| **manager-git** | sync-phase | Wave 1 PR 생성. Hybrid Trunk Tier M = feat branch + PR 의무 (Tier S main 직진 정책 무관) |
| **claude-code-guide** | 검증 보조 (선택) | Claude Code 런타임 `paths:` 해석 메커니즘 회귀 조사 (필요 시) |

---

## 5. Constraints

### 5.1 Technical Constraints

- **Frontmatter 형식**: CSV string (`paths: "a,b,c"`). YAML array (`paths: [a, b, c]`) 사용 금지 (`.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter 표준 위반).
- **Body 보존**: frontmatter prepend 만 허용. 기존 body 1 byte 도 변경 금지 (REQ-RPS-NF-014).
- **Template-First Rule**: local + template mirror 2 곳 동시 갱신 (CLAUDE.local.md §2 HARD).
- **No Go code**: `internal/rules/`, `internal/loader/` 디렉토리 신규 작성 금지 (REQ-RPS-NF-011).
- **glob 형식**: 기존 14+ 선례 표준 `**/.claude/skills/**` 또는 `.claude/agents/**/*.md` 패턴 따름. 절대 경로 금지.

### 5.2 Business Constraints

- **v3.0 6/15 deadline**: 사용자 §8 결정 1. Wave 1 본 SPEC 는 deadline 3주 이내 진입 (현재 5/22, 잔여 ~24일).
- **Wave 순서 의무**: Wave 0 → Wave 1 → Wave 2. Wave 0 (HOOK-CONTRACT-FIX-001 + DOCS-USER-DRIFT-001) MERGED 직후 진입 — 위반 없음.
- **GLM 보수 라우팅** (사용자 §8 결정 4): 본 SPEC 는 manager-develop 위임만 사용, GLM teammate 미사용 → 영향 없음.

### 5.3 Quality Constraints

- **TRUST 5 (CLAUDE.md §6)**: Tested (M3 doctor 시뮬레이션) / Readable (frontmatter convention 일치) / Unified (4 rule 동일 형식) / Secured (변경 영역 무관) / Trackable (Conventional Commits + 🗿 MoAI trailer).
- **Tier M plan-auditor PASS threshold**: 0.80 (spec-workflow.md § Tier M 정의).
- **Pre-existing CI baseline**: 본 SPEC 변경이 새 CI 결함 도입 금지. `internal/template/agentless_audit_test.go` + `internal/template/rule_template_mirror_test.go` (가칭) 모두 PASS 유지.

---

## 6. Risks

### R-RPS-001: Trigger miss → 필요한 rule 미로드 (Medium / High)

**Risk**: path glob 이 너무 좁아 rule 이 필요한 session 에 미로드. 예) `manager-develop-prompt-template.md` 의 paths 가 `.moai/specs/**` 만 포함 시, orchestrator 가 SPEC 외부 (예: TodoList 만 사용한 manager-develop 위임) 에서 5-section template 의무 망각.

**Likelihood**: Medium — path glob 설계가 보수적이면 (광범위) 위험 낮음. 너무 좁으면 trigger miss.
**Impact**: High — 위반 사항 (5-section template 누락, AskUserQuestion 잘못 호출 등) 이 silent 통과.
**Mitigation**:
- 각 rule 의 paths glob 을 design.md §Layer 1 표 권장값 + 1-2 추가 안전 글롭 (예: `manager-develop.md` agent 파일 자체 포함) 으로 확장
- run-phase M3 doctor 시뮬레이션 의무 (REQ-RPS-NF-012): 5 session 시나리오 (Go-only / SPEC-only / design / team / general) 별 rule 로드 매트릭스 생성. 위반 발견 시 paths glob 보정 후 재시뮬레이션.
- Acceptance.md AC-RPS-010 ↔ AC-RPS-013 으로 4 rule 각 trigger 조건 검증.

### R-RPS-002: Claude Code 런타임 `paths:` 해석 회귀 (Low / High)

**Risk**: 14+ 기존 path-scoped 규칙은 정상 동작 중이나, Claude Code 미래 버전이 `paths:` 해석을 변경 시 본 SPEC 4 rule 도 동시 실패.

**Likelihood**: Low — Claude Code 의 frontmatter `paths:` 는 안정 API (CLAUDE.md §13 Progressive Disclosure 와 함께 운영).
**Impact**: High — 9 → 5 감축이 무력화되어 ~23.5K tokens 회귀.
**Mitigation**:
- 본 SPEC 가 신규 메커니즘 도입 ZERO — 기존 14+ 사례와 동일 인프라 사용 (REQ-RPS-NF-011).
- Claude Code 런타임 회귀 발생 시, 14+ 규칙 + 본 4 rule 모두 동시 영향 → 전사적 정정 SPEC 필요. 본 SPEC 단독 책임 아님.
- 회귀 조사 도구: `claude-code-guide` agent (`.claude/rules/moai/development/agent-authoring.md`).

### R-RPS-003: Template mirror drift (Medium / Medium)

**Risk**: local `.claude/rules/moai/<sub>/*.md` 만 갱신 + template mirror `internal/template/templates/.claude/rules/moai/<sub>/*.md` 누락 시 → `moai update` 사용자가 옛 rule 받음.

**Likelihood**: Medium — Wave 1 Lane A 다른 SPEC 또는 V3R6 sync 작업에서 발생 선례 있음 (`project_v3r6_template_mirror_drift_audit_2026_05_22` memory: 10 .claude drift + 15 .moai/config drift 발견 패턴).
**Impact**: Medium — 사용자 프로젝트에 옛 rule 반영, plan-auditor `TestRuleTemplateMirrorDrift` 가 BLOCKING ERROR.
**Mitigation**:
- REQ-RPS-006 (template mirror sync 의무) HARD 강제.
- run-phase M2 단계에서 `diff` 비교 의무 (acceptance.md AC-RPS-015).
- `make build` 후 임베디드 갱신 검증.

### R-RPS-004: Frontmatter parse 실패 → rule 전체 미로드 (Low / High)

**Risk**: YAML frontmatter 문법 오류 (예: 따옴표 escape 누락, indent 불일치) → Claude Code 가 rule 자체를 무시 → rule 부재 효과.

**Likelihood**: Low — 14+ 선례 frontmatter 형식이 CSV 단순 형식, 복잡한 escape 불필요.
**Impact**: High — rule 부재 시 HARD 조항 SSOT (zone-registry) 미가시 → 회귀 가능.
**Mitigation**:
- 4 rule frontmatter 작성 시 가장 단순한 선례 (`agent-hooks.md` `paths: "**/.claude/agents/**,**/.claude/hooks/**"`) 형식 모방.
- acceptance.md AC-RPS-009 에서 4 rule 의 YAML parse 검증 (`yq` 또는 Python `yaml.safe_load`).
- M3 doctor 시뮬레이션 시 4 rule 모두 정상 로드 확인.

### R-RPS-005: Always-loaded 부담 잘못 측정 → 절감 효과 미달 (Low / Low)

**Risk**: 본 SPEC가 -23.5K tokens 절감 주장이나, 실제 Claude Code 의 tokenizer 계산 시 7,831w → tokens 변환 비율 차이로 절감액이 ±15% 오차.

**Likelihood**: Low — token estimation 은 일반적으로 ~3 tokens per word (OpenAI/Anthropic) 안정 값.
**Impact**: Low — 절감 자체는 유지, 절감 절댓값만 ±15% 변동.
**Mitigation**:
- spec.md §1.3 표에서 "단어 수" + "추정 tokens (~3w/token)" 명시.
- acceptance.md 는 절감 절댓값 (-23.5K) 보다는 **단어 수 감소 (REQ-RPS-NF-010)** 를 binary AC 로 측정.
- run-phase 보고에 실측 token 측정 별도 보고 (선택).

---

## 7. References

- `.moai/research/v3.0-design-2026-05-22.md` §Layer 1 (라인 165-184) — 본 SPEC 청사진 직접 출처
- `.moai/research/v3.0-design-2026-05-22.md` §Wave 1 (라인 358-365) — Wave 카탈로그
- `.moai/research/v3.0-design-2026-05-22.md` §7 (라인 439) — "Rule path-scoped 화 시 trigger miss" 위험 평가
- `.moai/research/v3.0-design-2026-05-22.md` §8 (라인 449-456) — 사용자 4 결정
- `.claude/rules/moai/development/coding-standards.md` § Paths Frontmatter — CSV string 표준
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — SPEC frontmatter 12-field SSOT (본 spec.md 자체에 적용)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier M 정의 + 3 artifacts + 0.80 threshold
- `.moai/specs/SPEC-V3R6-DOCS-USER-DRIFT-001/spec.md` — Wave 0 두 번째 SPEC 패턴 reference
- `.moai/specs/SPEC-V3R6-HOOK-CONTRACT-FIX-001/spec.md` — Wave 0 첫 SPEC 패턴 reference
- CLAUDE.md §1 HARD Rules — Multi-File Decomposition (3+ files)
- CLAUDE.local.md §2 [HARD] Template-First Rule — local + template mirror 동시 갱신
- 선례 14+ paths frontmatter 사례 — `agent-hooks.md` / `moai-constitution.md` / `settings-management.md` / `boundary-verification.md` / `hooks-system.md` / `spec-frontmatter-schema.md` / `skill-authoring.md` / `karpathy-quickref.md` / `skill-ab-testing.md` / `agent-patterns.md` / `skill-writing-craft.md` / `model-policy.md` / `agent-authoring.md`

---
id: SPEC-V3R6-AGENT-MODEL-ROUTING-001
title: "Agent 23개 모델 명시 라우팅 (opus 7 / sonnet 13 / haiku 3)"
version: "0.2.1"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents"
lifecycle: spec-anchored
tags: "agent, model-routing, opus, sonnet, haiku, cost-optimization, sprint-2, v3.0"
tier: L
depends_on: [SPEC-V3R6-RULES-PATH-SCOPE-001]
related_specs: [SPEC-V3R6-PROMPT-CACHE-001, SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-BACKEND-ROUTING-001]
---

# SPEC-V3R6-AGENT-MODEL-ROUTING-001 — Agent 23개 모델 명시 라우팅

## HISTORY

- v0.2.1 (2026-05-23): orchestrator-direct minor fix-forward — REQ-AMR-NF-009 + AC-AMR-009 + plan.md §D.5 정정. CLAUDE.local.md §24.2 namespace policy 충돌 해소: `internal/template/templates/.claude/agents/harness/` directory는 §24.2에 의해 **존재 자체 금지**이므로 template mirror 의무에서 harness/ 4 agents 제외. 23 agents 모두 routing 적용 유지 (REQ-AMR-001..005), 단 mirror sync는 19 non-harness pairs (core 8 + expert 6 + meta 5)만 의무. AC-AMR-009 anti-leak guard 추가 (harness/ template directory 존재 시 FAIL). 본 정정은 plan-auditor iter 2 PASS 0.871 post-audit codebase-state blindspot 발견에 따른 fix-forward (≤5 edits, scope-clarification only, score-affecting REQ wording 외 변경 없음, manager-spec 재위임 불요). plan-auditor codebase state blindspot lesson candidate (memory entry [[feedback-plan-auditor-codebase-state-blindspot]] 재현).

- v0.2.0 (2026-05-23): plan-auditor iter 1 REVISE 0.633 (Tier M 0.80 미만 -0.167) 4 BLOCKING 적용. B1 foundational inventory 정정: 19 agents → 23 agents (4 subdirectories: core 8 / expert 6 / harness 4 / meta 5). B2 AC 검증 명령 4-subdirectory glob으로 재작성. B3 `module:` 경로 정정 (`.claude/agents/moai` → `.claude/agents`). B4 cost 산수 + opus 카운트 정합: 7 opus (manager-develop, manager-spec, manager-strategy, expert-security, **expert-refactoring** — constitution-aligned, plan-auditor, evaluator-active) + 13 sonnet (8 harness/builder/expert/manager + 4 new harness specialists + 1 claude-code-guide) + 3 haiku (manager-docs, manager-git, researcher + batch_api). Opus 호출 빈도 23/23 baseline → 7/23 = 70% off. 메인 워크플로우 `/moai run` 5-agent 평균 input 단가 $4.20 → $2.20/MTok 유지 (5-agent set membership 불변). 5 SHOULD-FIX 적용: S1 sprint-round-naming.md SSOT (Wave → Round retired, Sprint = multi-SPEC). S2 expert-refactoring opus carve-out constitution 본문 직접 인용. S3 Tier M → Tier L 재분류 (50+ files affected, 5 artifacts). S4 batch_api key resilience (3-key accept). S5 baseline state line 정확화 (main HEAD `0abcda296` + Sprint 1 Lane A status).

- v0.1.0 (2026-05-23): 초기 draft (Tier M, 19-agent 추정). plan-auditor iter 1 REVISE 0.633 — 4 BLOCKING + 5 SHOULD-FIX. 본 v0.2.0에서 폐기 후 재작성.

---

## 1. Background

### 1.1 Project Baseline (정확 상태)

- **Main HEAD**: `0abcda296` (2026-05-23, V3R6 UPDATE Audit 3 SPECs plan-complete)
- **Sprint 1 Lane A 진행 상황**: 4 SPECs 중 1개 implementation 머지 완료 (RULES-PATH-SCOPE-001, commit `7ed4c841c`); 잔여 3개 SPECs (RULES-COMPRESS-001 / SKILL-CONSOLIDATE-001 / SKILL-COMPRESS-001) plan-complete + run-pending
- **현 Sprint 위치**: AGENT-MODEL-ROUTING-001은 **Sprint 2** Tier L 첫 SPEC (parallel SPECs: PROMPT-CACHE-001 / HOOK-OBSERVE-OPT-IN-001 / HOOK-ASYNC-EXPAND-001)
- **Run-phase entry precondition**: Sprint 1 Lane A 잔여 3개 SPECs 머지 완료 후. Plan-phase 자체는 Sprint 1 잔여 SPECs와 병렬 가능 (markdown only).

### 1.2 Actual Agent Inventory (검증 명령 `find .claude/agents -name "*.md" -type f`)

**총 23 agents in 4 subdirectories** (`.claude/agents/{core, expert, harness, meta}/`):

```
.claude/agents/
├── core/ (8 agents)
│   ├── manager-brain.md          (currently model: inherit)
│   ├── manager-develop.md        (currently model: inherit)
│   ├── manager-docs.md           (currently model: haiku)
│   ├── manager-git.md            (currently model: haiku)
│   ├── manager-project.md        (currently model: inherit)
│   ├── manager-quality.md        (currently model: inherit)
│   ├── manager-spec.md           (currently model: inherit)
│   └── manager-strategy.md       (currently model: inherit)
├── expert/ (6 agents)
│   ├── expert-backend.md         (currently model: inherit)
│   ├── expert-devops.md          (currently model: inherit)
│   ├── expert-frontend.md        (currently model: inherit)
│   ├── expert-performance.md     (currently model: inherit)
│   ├── expert-refactoring.md     (currently model: inherit)
│   └── expert-security.md        (currently model: inherit)
├── harness/ (4 agents)
│   ├── cli-template-specialist.md       (currently model: inherit)
│   ├── hook-ci-specialist.md            (currently model: inherit)
│   ├── quality-specialist.md            (currently model: inherit)
│   └── workflow-specialist.md           (currently model: inherit)
└── meta/ (5 agents)
    ├── builder-harness.md        (currently model: inherit)
    ├── claude-code-guide.md      (currently model: inherit)
    ├── evaluator-active.md       (currently model: inherit)
    ├── plan-auditor.md           (currently model: inherit)
    └── researcher.md             (currently model: inherit)

Total: 23 agents (21 inherit + 2 haiku + 0 sonnet + 0 opus)
```

**Pre-flight 검증 명령**:

```bash
find .claude/agents -name "*.md" -type f | wc -l                                      # → 23
grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md | wc -l       # → 21
grep -l 'model: haiku'   .claude/agents/{core,expert,harness,meta}/*.md | wc -l       # → 2
```

### 1.3 Goal

23 agents in `.claude/agents/{core,expert,harness,meta}/*.md` 모두의 YAML frontmatter `model:` 필드를 명시적 3-tier 분배로 전환. 현재 21/23 = `model: inherit` (실질 Opus 4.7 default) → 23/23 명시: **opus 7 / sonnet 13 / haiku 3** (researcher는 batch_api opt-in 동반). 본 SPEC는 정적 frontmatter만 변경 (Go 코드 변경 0건).

### 1.4 Motivation — 비용 산술 (재산정)

**현재 부담**:
- 23 agents 중 21개 = `model: inherit` → 세션 기본 모델 추종 (실질적으로 Opus 4.7 1M).
- Opus 4.7 단가: $5 in / $25 out per MTok. 새 토크나이저 +35% 실효 비용 (`.moai/research/v3.0-design-2026-05-22.md` §1.3).
- 메인 `/moai run` 5-agent (manager-develop + expert-backend + manager-quality + manager-git + manager-docs) 평균 input 단가: **~$4.20/MTok**.

**v3.0 명시 분배 후**:
- Opus 호출 빈도: 23/23 baseline (모두 inherit/Opus) → **7/23 (30%)** = **70% off Opus calls**.
- 5-agent set membership 불변 (manager-develop opus + expert-backend sonnet + manager-quality sonnet + manager-git haiku + manager-docs haiku) → 평균 input 단가 **~$2.20/MTok = 48% off**.
- researcher Batch API: $5 → $0.50/MTok input (90% off, 비동기 워크로드).

**v3.0 청사진 §Layer 3 정합**: 19-agent 표는 청사진 작성 시점 인벤토리 추정 (`.moai/research/v3.0-design-2026-05-22.md` 라인 213-241). 본 SPEC는 실측 23-agent 인벤토리로 정정 + 4 harness specialists (cli-template, hook-ci, quality, workflow) 추가 분류 + expert-refactoring constitution-aligned opus 재배치.

### 1.5 Constitution Alignment — expert-refactoring Opus Carve-out

`.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy (라인 55) 직접 인용:

> "Effort level selection: reasoning-intensive agents (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, **expert-refactoring**) → `effort: xhigh` or `high`; implementation agents (expert-backend, expert-frontend, builder-*) → `effort: high` (default for Opus 4.7); speed-critical agents (manager-git, Explore) → `effort: medium`"

본 SPEC는 constitution을 정렬 우선 권위로 인정:
- **expert-refactoring → opus** (constitution reasoning-intensive list verbatim, AST 변환은 추론 깊이 요구).
- Iter 1 risk R-AMR-006 (constitution vs design.md 불일치)은 본 v0.2.0에서 **해소** — risk가 아니라 design choice (constitution 본문 우선).

### 1.6 Sprint-Round-Milestone Naming (SSOT 정렬)

`.claude/rules/moai/development/sprint-round-naming.md` SSOT v2.0.0 정렬:

| Term | Meaning | 본 SPEC 적용 |
|------|---------|--------------|
| **Sprint** | Multi-SPEC time-unit grouping | 본 SPEC = Sprint 2 첫 SPEC (4 SPECs: AGENT-MODEL-ROUTING + PROMPT-CACHE + HOOK-OBSERVE-OPT-IN + HOOK-ASYNC-EXPAND) |
| **Round** | Single-SPEC internal phase (SSE stall mitigation, ≥30 tasks) | 본 SPEC = Tier L M1~M6 = 6 milestones, no Round split required |
| **Milestone** | M1~M6 ordered work steps | 본 SPEC M1 (baseline) → M2 (opus) → M3 (sonnet) → M4 (haiku + batch_api) → M5 (template mirror) → M6 (docs-site 4-locale + chore) |

design 문서 § 5에서 인용 시: "Wave 2 표" (design doc legacy 명명 "Wave 2"; per `.claude/rules/moai/development/sprint-round-naming.md` SSOT v2.0.0 = **Sprint 2**).

---

## 2. User Stories

### US-AMR-001: 메인 `/moai run` 워크플로우 비용 48% 절감

> **As a** MoAI-ADK 사용자 (`/moai run` 일상 사용)
> **I want** `/moai run` 워크플로우의 5-agent 평균 input 단가가 $4.20/MTok → $2.20/MTok 으로 절감되기를
> **So that** Pro $20 Agent SDK 크레딧으로 가능한 turn 수가 ~18턴 → ~34턴 (88% 증가) 으로 늘어난다

**Acceptance**: 23 agents 중 7개 `model: opus`, 13개 `model: sonnet`, 3개 `model: haiku`, `model: inherit` 카운트 = 0.

### US-AMR-002: 품질 검증 작업 sonnet 강등으로 비용 절감 (regression bound ±5%)

> **As a** MoAI-ADK 사용자
> **I want** `manager-quality` / `expert-backend` / `expert-frontend` / 4 harness specialists 등 sonnet 강등 가능 agent가 sonnet으로 작동하되 응답 품질이 baseline ±5% 이내로 유지되기를
> **So that** 비용 절감 (Opus → Sonnet -40%) 효과가 품질 저하 없이 실현된다

**Acceptance**: REQ-AMR-006 baseline 측정 + REQ-AMR-007 regression bound ±5% 정의. AC-AMR-006 binary 검증.

### US-AMR-003: researcher Batch API 활성으로 자체 연구 비용 90% 절감

> **As a** MoAI-ADK 사용자
> **I want** `researcher` agent가 Haiku 4.5 + Batch API 조합으로 작동하기를
> **So that** 비동기 OK 인 대량 처리 작업이 Opus 동기 호출 대비 ~90% 비용 절감 ($5 → $0.50/MTok input)

**Acceptance**: `researcher.md` frontmatter에 `model: haiku` + (`batch_api: true` OR `use_batch_api: true` OR `invocation_mode: batch`) 중 1개 명시.

### US-AMR-004: Template mirror 동기화로 사용자 프로젝트 일관성

> **As a** MoAI-ADK 사용자 (신규 `moai init` 또는 `moai update -t`)
> **I want** 새로 init/update한 프로젝트의 23 agents가 본 SPEC 분류와 동일하게 작동하기를
> **So that** v3.0 cost-optimization 효과가 사용자 프로젝트에도 즉시 적용된다

**Acceptance**: `internal/template/templates/.claude/agents/{core,expert,harness,meta}/*.md` 23 파일 모두 `model:` 명시, local과 byte-identical.

### US-AMR-005: docs-site 4-locale catalog 일관성

> **As a** MoAI-ADK 사용자 (docs-site `adk.mo.ai.kr` 참고)
> **I want** docs-site agent catalog 4-locale 페이지에 23 agents의 신규 모델 분류가 반영되기를
> **So that** 사용자 가시 catalog가 실제 frontmatter와 일치 (drift 회피)

**Acceptance**: `docs-site/content/{en,ko,ja,zh}/` agent catalog 페이지에 opus/sonnet/haiku 분류 명시, parity ratio ≤ 1.20.

---

## 3. EARS Requirements (GEARS notation)

### 3.1 Functional Requirements

**REQ-AMR-001 (Ubiquitous — 23/23 명시)**:
The MoAI-ADK orchestrator **shall** classify each of the 23 agents in `.claude/agents/{core,expert,harness,meta}/*.md` into exactly one of `{opus, sonnet, haiku}` model tiers via the YAML frontmatter `model:` field, with zero occurrences of `model: inherit`.

**REQ-AMR-002 (Where — opus tier 7 agents)**:
**Where** an agent is classified into the opus tier (7 agents: `manager-develop`, `manager-spec`, `manager-strategy`, `expert-security`, `expert-refactoring`, `plan-auditor`, `evaluator-active`), the SPEC implementation **shall** set `model: opus` in the agent's YAML frontmatter.

**REQ-AMR-003 (Where — sonnet tier 13 agents)**:
**Where** an agent is classified into the sonnet tier (13 agents: `manager-brain`, `manager-project`, `manager-quality`, `expert-backend`, `expert-devops`, `expert-frontend`, `expert-performance`, `builder-harness`, `claude-code-guide`, `cli-template-specialist`, `hook-ci-specialist`, `quality-specialist`, `workflow-specialist`), the SPEC implementation **shall** set `model: sonnet` in the agent's YAML frontmatter.

**REQ-AMR-004 (Where — haiku tier 3 agents)**:
**Where** an agent is classified into the haiku tier (3 agents: `manager-docs`, `manager-git`, `researcher`), the SPEC implementation **shall** set `model: haiku` in the agent's YAML frontmatter.

**REQ-AMR-005 (When + Where — researcher Batch API opt-in)**:
**When** the `researcher` agent's frontmatter is updated under REQ-AMR-004 **AND** **where** the Claude Code runtime accepts a Batch API opt-in key, the SPEC implementation **shall** additionally set exactly one of the following keys: `batch_api: true` OR `use_batch_api: true` OR `invocation_mode: batch`. M1 verifies the canonical key against the Claude Code SDK official documentation before frontmatter editing.

**REQ-AMR-006 (Ubiquitous — baseline measurement)**:
The MoAI-ADK orchestrator **shall** record an A/B baseline at `.moai/state/agent-model-baseline.jsonl` BEFORE applying model changes, containing at least 23 entries (one per agent) with fields `{agent_name, baseline_input_tokens_avg, baseline_output_tokens_avg, baseline_quality_score, measurement_window_start, measurement_window_end}`. This baseline enables regression detection per REQ-AMR-007.

**REQ-AMR-007 (When + Where — quality regression bound)**:
**When** a sonnet-tier or haiku-tier agent (formerly inherit/Opus) is invoked under typical workflow conditions **AND** **where** the task category is `analysis | code_review | refactoring | documentation`, the post-change quality score (measured on a defined eval set) **shall not** regress beyond ±5% relative to the REQ-AMR-006 baseline. Out-of-bound regression triggers escalation per plan.md §M4 risk mitigation.

**REQ-AMR-008 (Ubiquitous — inherit count zero)**:
The MoAI-ADK orchestrator **shall** maintain `model: inherit` count = 0 across all 23 agents post-migration, verified by `grep -l 'model: inherit' .claude/agents/{core,expert,harness,meta}/*.md | wc -l` returning `0`.

### 3.2 Non-Functional Requirements

**REQ-AMR-NF-009 (Template mirror sync, harness/ exclusion)**:
**When** any agent file in `.claude/agents/{core,expert,meta}/` is modified under REQ-AMR-001..005, the SPEC implementation **shall** apply byte-identical changes to the corresponding mirror file at `internal/template/templates/.claude/agents/{core,expert,meta}/<filename>.md` to satisfy CLAUDE.local.md §2 [HARD] Template-First Rule. Verification: `diff -q .claude/agents/<sub>/<agent>.md internal/template/templates/.claude/agents/<sub>/<agent>.md` returns 0 mismatches for **19 non-harness pairs** (core 8 + expert 6 + meta 5 = 19).

**Exception (CLAUDE.local.md §24.2 Namespace Policy — harness/ user-owned)**: harness/ 4 agents (`cli-template-specialist`, `hook-ci-specialist`, `quality-specialist`, `workflow-specialist`) are **user-owned namespace** per CLAUDE.local.md §24.2. Their `model:` frontmatter is modified in `.claude/agents/harness/*.md` only — `internal/template/templates/.claude/agents/harness/` directory **MUST NOT exist** (per §24.2 contract). Template mirror sync verification SKIPS the 4 harness pairs. REQ-AMR-001..005 still apply (23 agents total receive routing); only the mirror sync obligation excludes harness/.

**REQ-AMR-NF-010 (Cost reduction target)**:
The cumulative Opus invocation frequency across the 23 agents **shall** decrease from 23/23 (current effective state via `inherit`) to 7/23 (70% reduction in Opus calls). Measurement: post-migration grep count of `model: opus` returns exactly `7`.

**REQ-AMR-NF-011 (No Go code change)**:
The SPEC implementation **shall not** introduce any new Go source file, modify `internal/agent/` (if exists), or create any agent-loading mechanism. The Claude Code runtime's existing `model:` frontmatter handler (verified by current operation of `manager-git` + `manager-docs` with `model: haiku`) is the sole loading mechanism.

**REQ-AMR-NF-012 (Backward compat — env-var override)**:
The SPEC implementation **shall** preserve all current runtime model override mechanisms (e.g., `CLAUDE_MODEL` env var, `moai cc -p <profile>` profile-specific settings, `effortLevel` from settings.json). Static frontmatter `model:` is the default; env-var override remains the runtime-level escape hatch.

**REQ-AMR-NF-013 (docs-site parity)**:
**When** the 23 agent files are migrated, the docs-site agent catalog 4-locale pages (`docs-site/content/{en,ko,ja,zh}/`) **shall** reflect the new opus/sonnet/haiku classification with parity ratio ≤ 1.20 across locales (per `.moai/research/v3.0-design-2026-05-22.md` §6 KPI § docs-site i18n parity convention).

### 3.3 Out of Scope

### Out of Scope: model retirement lifecycle

v3.0 sunset planning (deprecation of older Anthropic model identifiers, e.g., `opus-4.6`, `sonnet-4.5`) is deferred to Sprint 5 SPEC-V3R6-RELEASE-V3-001. Do not include `sunset.yaml` changes in this SPEC. This SPEC only sets per-agent model tier; the underlying model version (Opus 4.7 / Sonnet 4.6 / Haiku 4.5) resolution is handled by Claude Code runtime configuration.

### Out of Scope: runtime model override

Existing env-var override mechanisms (e.g., `CLAUDE_MODEL` environment variable, `effortLevel` in `.claude/settings.json`, `moai cc -p <profile>` profile-specific model settings) remain untouched. This SPEC only changes static frontmatter defaults. Users retain the ability to override model selection at runtime via existing mechanisms — no policy change.

### Out of Scope: backend routing policy

GLM / Claude / Haiku backend selection (Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001) is sibling work and is **not** part of this SPEC. This SPEC only sets per-agent model tier (Opus / Sonnet / Haiku within Anthropic Claude family), not the underlying provider (Anthropic Claude vs Z.AI GLM vs others). Cross-SPEC composition (e.g., GLM + sonnet tier) is design.md §Layer 6 future work.

### Out of Scope: prompt caching policy

Prompt caching ephemeral configuration (1h cache_write / 1h hit break-even per model tier) is deferred to Sprint 2 SPEC-V3R6-PROMPT-CACHE-001 (sibling SPEC). Cross-SPEC R-AMR-004 documents the sync-phase ordering: AMR merge first → PROMPT-CACHE merge second.

---

## 4. Stakeholders

| 역할 | 이름 | 책임 |
|---|---|---|
| **사용자 (메인테이너)** | GOOS행님 | v3.0 환골탈태 §8 4 결정 발의자. 머지 승인 |
| **manager-spec** | (본 SPEC 위임) | spec/plan/acceptance/design/research 5 artifacts 작성 |
| **plan-auditor** | (orchestrator dispatch) | Tier L PASS threshold 0.85 검증 |
| **manager-develop** | run-phase 위임 | M1 (baseline) → M2 (opus 7) → M3 (sonnet 13) → M4 (haiku 3 + researcher batch_api) → M5 (template mirror) → M6 (docs-site 4-locale) |
| **manager-git** | sync-phase | Sprint 2 PR 생성. Hybrid Trunk Tier L = feat branch + PR 의무 |
| **manager-quality** | post-run validation | REQ-AMR-007 regression bound ±5% 검증 |
| **claude-code-guide** | M1 보조 | Claude Code 런타임 `model:` / `batch_api:` 해석 메커니즘 회귀 조사 |
| **expert-security** | M2 검토 (선택) | opus tier 7개 중 expert-security 자체 분류 정합 |

---

## 5. Constraints

### 5.1 Technical Constraints

- **Frontmatter 형식**: YAML 단일 값 (`model: opus` / `model: sonnet` / `model: haiku`). 따옴표 불필요 (선례: `manager-git.md` / `manager-docs.md` 형식 일치).
- **Body 보존**: frontmatter `model:` 필드만 변경. 기존 body 1 byte도 변경 금지.
- **Template-First Rule**: local + template mirror 2곳 동시 갱신 (CLAUDE.local.md §2 HARD). 4 subdirectory mirror (`internal/template/templates/.claude/agents/{core,expert,harness,meta}/`).
- **No Go code**: `internal/agent/`, `internal/loader/` 디렉토리 신규 작성 금지 (REQ-AMR-NF-011).
- **researcher Batch API key 명**: M1 단계에서 Claude Code 공식 SDK 문서 또는 claude-code-guide agent 회귀 조사로 canonical key 확정. 후보 3: `batch_api: true` / `use_batch_api: true` / `invocation_mode: batch`.

### 5.2 Business Constraints

- **v3.0 6/15 deadline**: 사용자 §8 결정 1. Sprint 2 본 SPEC는 deadline 3주 이내 진입 (현재 5/23, 잔여 ~23일).
- **Sprint 순서 의무**: Sprint 0 → Sprint 1 → Sprint 2. Sprint 1 Lane A 잔여 3개 SPECs (RULES-COMPRESS / SKILL-CONSOLIDATE / SKILL-COMPRESS) 머지 후 본 SPEC run-phase 진입 권장. Plan-phase 자체는 Sprint 1과 병렬 가능.
- **Cross-SPEC R-AMR-004 (PROMPT-CACHE-001)**: 본 SPEC + PROMPT-CACHE-001 sync-phase 분리 의무. AMR 먼저 머지 (model 명시 → cache_write 손익분기 모델별 다름).

### 5.3 Quality Constraints

- **TRUST 5**: Tested (M4 manager-quality regression ±5% 검증) / Readable (frontmatter convention 일치) / Unified (23 agents 동일 형식) / Secured (변경 영역 무관) / Trackable (Conventional Commits + 🗿 MoAI trailer).
- **Tier L plan-auditor PASS threshold**: 0.85 (spec-workflow.md § Tier L 정의).
- **Pre-existing CI baseline**: 본 SPEC 변경이 새 CI 결함 도입 금지. `go build ./...` + `go vet ./...` + `moai doctor` exit 0 유지.

---

## 6. Risks

### R-AMR-001: Sonnet 강등으로 manager-quality / expert-backend / 4 harness specialists 등 품질 ±5% 이상 저하 (Medium / Medium)

**Risk**: 13 sonnet tier agents 중 일부가 sonnet 강등 후 baseline 대비 -5% 초과 품질 저하. 특히 복잡한 OWASP 외 보안 패턴 분석 또는 다층 의존성 refactoring에서 Sonnet 추론 깊이 부족 표면화 가능.

**Likelihood**: Medium — Sonnet 4.6은 Claude 3.7 Sonnet 대비 +30% 추론 능력 향상 (Anthropic 공식). 그러나 Opus 4.7 새 토크나이저 +35% 충격과 결합 시 동등 성능 보장 불확실.
**Impact**: Medium — 품질 저하 시 사용자가 즉시 표면화 가능. 본 SPEC frontmatter 변경은 즉시 revert 가능 (1 commit).
**Mitigation**:
- REQ-AMR-006 baseline 측정 의무 (23 agents × 5 task category × N samples).
- REQ-AMR-007 regression bound ±5% — 위반 발견 시 plan.md §M4 escalation: 해당 agent만 opus로 revert.
- post-run manager-quality 위임에서 baseline JSONL과 post-change actual 비교 의무.

### R-AMR-002: researcher Batch API 활성 시 latency ↑ (응답 지연) (Low / Medium)

**Risk**: Batch API는 본질적으로 비동기 — submission → completion까지 수 분 ~ 수 시간 소요 가능. Caller가 동기 응답 기대 시 timeout.

**Likelihood**: Low — researcher agent는 본래 비동기 OK (워크트리 격리 실험 / 자체 연구).
**Impact**: Medium — Caller 동기 의존 시 workflow 차단. 단 본 SPEC는 caller 패턴 변경 0건.
**Mitigation**:
- REQ-AMR-005 specifies opt-in key only — caller pattern 변경 없음.
- M4 단계에서 researcher caller 패턴 검증 (1개 sample 실제 호출 + 응답 시간 측정).
- 위반 발견 시 batch_api opt-in revert.

### R-AMR-003: `model: inherit` 제거 시 일부 agent fallback chain 의존 → 호출 실패 (Medium / High)

**Risk**: 21/23 = `model: inherit` 현재 상태에서 일부 agent가 명시적 fallback chain (Opus 거부 → Sonnet fallback)에 의존 중일 가능성. 명시 model 지정 시 fallback chain 우회 → 호출 실패.

**Likelihood**: Medium — `manager-git` + `manager-docs`는 이미 `model: haiku` 명시 + 정상 동작 (선례). 나머지 21 agents fallback 의존 여부 미검증.
**Impact**: High — 호출 실패 시 해당 agent 작동 불능. 본 SPEC 머지 직후 표면화 가능.
**Mitigation**:
- pre-change inventory: M1 baseline 측정 시 23 agents 모두 정상 호출 가능 dry-run.
- canary first: opus 7 → sonnet 13 → haiku 3 milestone 분할 적용.
- M2-M4 단계별 즉시 검증 의무.

### R-AMR-004: Cross-Sprint conflict with SPEC-V3R6-PROMPT-CACHE-001 (Medium / Medium)

**Risk**: 본 SPEC + PROMPT-CACHE-001 sync-phase 중첩 시 model + cache 조합 미검증. Sonnet의 1h cache_write 손익분기 (1h cache_write +100% / 1h hit -90%)가 Opus와 다름.

**Likelihood**: Medium — Sprint 2 동기 SPECs 4건. plan-phase는 4건 모두 병렬, run-phase는 AMR 먼저 머지 권장.
**Impact**: Medium — A/B regression detection (REQ-AMR-007)이 cache 효과와 혼합.
**Mitigation**:
- 본 SPEC AMR 먼저 머지 의무 명시 (REQ-AMR-006 baseline = pre-cache state).
- PROMPT-CACHE-001 SPEC body에 "AMR 머지 후 진입" pre-condition 명시.
- depends_on / related_specs frontmatter 양방향 연결.

### R-AMR-005: docs-site 4-locale mirror drift (Low / Low)

**Risk**: `docs-site/content/{en,ko,ja,zh}/` agent catalog 4건 동시 갱신 시 locale 별 drift 가능.

**Likelihood**: Low — Sprint 1 RULES-PATH-SCOPE-001 + Sprint 6 GEARS-MIGRATION-001에서 4-locale 동기화 패턴 확립.
**Impact**: Low — docs-site drift만, runtime 무관.
**Mitigation**:
- AC-AMR-013 parity ratio ≤ 1.20 검증.
- M6 단계에서 4-locale 동시 갱신 의무.
- post-run `docs-i18n-check` baseline 확인.

---

## 7. References

- `.moai/research/v3.0-design-2026-05-22.md` §Layer 3 (라인 213-241) — 19-agent 표 verbatim (본 SPEC v0.2.0에서 실측 23-agent로 정정)
- `.moai/research/v3.0-design-2026-05-22.md` §Wave 2 (라인 370-374) — Sprint 2 (design doc legacy "Wave 2") SPEC 4건 카탈로그
- `.moai/research/v3.0-design-2026-05-22.md` §6.1 KPI (라인 412-414) — 단일 `/moai run` 턴 평균 비용 $1.10 → ≤ $0.45 목표
- `.moai/research/v3.0-design-2026-05-22.md` §1.3 (라인 56-60) — 모델 단가표 (Opus 4.7 / Sonnet 4.6 / Haiku 4.5)
- `.moai/research/v3.0-design-2026-05-22.md` §1.4 (라인 64-73) — 가격 레버 (Batch API 50% off, Haiku 강등 -80%, Sonnet 강등 -40%)
- `.moai/research/moai-adk-current-state-2026-05-22.md` §3 (라인 134-178) — agent 현재 상태 (19-agent 추정, 본 SPEC에서 23-agent로 정정)
- `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy 라인 55 — **expert-refactoring opus carve-out verbatim**
- `.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT — 12-field canonical (본 spec.md 자체 적용)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier L 정의 + 5 artifacts + 0.85 threshold
- `.claude/rules/moai/development/sprint-round-naming.md` SSOT v2.0.0 — Sprint = multi-SPEC time-unit; Round = within-SPEC SSE-stall mitigation; Milestone = M1~M6 lifecycle. Wave terminology retired (AP-SRN-004).
- `.moai/specs/SPEC-V3R6-RULES-PATH-SCOPE-001/spec.md` — Sprint 1 첫 SPEC 패턴 reference (frontmatter, Out of Scope h3, REQ/AC traceability)
- `.moai/specs/SPEC-V3R6-GEARS-MIGRATION-001/spec.md` — GEARS notation 선례
- CLAUDE.md §1 HARD Rules — Multi-File Decomposition (3+ files; 본 SPEC 50+ files affected → Tier L)
- CLAUDE.local.md §2 [HARD] Template-First Rule — local + template mirror 동시 갱신
- CLAUDE.local.md §15 [HARD] 16-language 중립성
- Anthropic Batch API Documentation — `https://docs.anthropic.com/en/api/messages-batches` (Batch API 50% off in+out, async submission, up to 24h completion)
- Claude Code agent frontmatter spec — `model:` field 운영 사례 (manager-git / manager-docs 현재 운영)

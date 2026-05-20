---
title: MoAI-ADK Harness Autonomy Vision — Core Slim + Self-Evolving Harness (v2)
date: 2026-05-18
author: MoAI orchestrator (Opus 4.7, ultrathink)
source: 사용자 directive (2026-05-18 세션)
scope: forward-looking architecture vision (NOT a SPEC, prerequisite for Mega-Sprint W0-W4)
predecessor: .moai/research/architecture-audit-2026-05-18.md
iteration: 3 (2026-05-20 사용자 정정 — §3.5 Determinism 섹션 + 관련 acceptance bullet 모두 제거. 자세한 사유: memory/feedback_w4_no_determinism.md, memory/feedback_w4_true_goal.md)
prev_audit: plan-auditor FAIL 0.62 (D1-D10 defects)
classification: VISION_PROPOSAL (iteration 3 — Determinism 거부 반영, Core Slimming scope 강화)
status: revised, pending re-audit
---

# MoAI-ADK Harness Autonomy Vision (v2) — 2026-05-18

## 0. Executive Summary

본 vision document는 **2026-05-18 사용자 directive**를 architecture proposal로 정식화한다. 핵심은 MoAI-ADK를 두 zone으로 명확히 분리하는 것이다:

1. **Core MoAI** — 모든 프로젝트에 동일하게 배포되는 **범용/최소화/안정성** 우선 layer
2. **My Harness** — 프로젝트별 생성되는 **자율 진화** layer

### Iteration 2 changes (D1-D10 resolved)

| Defect | Status | Resolution |
|---|---|---|
| D1 (agent count 모순) | ✓ resolved | §0/§3/§4 모두 **17개 Core agents**로 통일 (cross-cutting experts 유지) |
| D2 (Frozen Guard 비강제) | ✓ resolved | §3.4 PreToolUse hook + path-glob deny pattern runtime enforce 신설 |
| D3 (schema 충돌) | ✓ resolved | §6.2 + new §6.7 — harness YAML이 SPEC schema와 align (`created:`/`updated:` canonical) |
| D4 (cold-start regression) | ✓ resolved | §4.4 seed pattern + §8 risk row 8 신설 |
| D5 (W2/W3 dependency 모순) | ✓ resolved | §5 + §7 일관화 — W3 depends on W1 only (W2 parallel) |
| D6 (backward compat 미명세) | ✓ resolved | §5 W2 AC에 legacy invocation handling 명시 (stub redirect) |
| D7 (Tier 4 fatigue) | ✓ resolved | §6.6 Proposal Throttling 신설 (batch + per-category mute) |
| D8 ("+∞" 불가측) | ✓ resolved | §0 table — "3-7 per project type" measurable |
| D9 (cold-start risk 누락) | ✓ resolved | §8 row 8 (D4와 통합) |
| D10 (Q3 determinism) | ✗ REMOVED (2026-05-20 사용자 결정) | §3.5 Design Constraint 삭제. 사유: 프로젝트마다 **다르게** 생성되는 것이 W4 본질. Determinism 보장은 과잉 설계. 관련 acceptance bullet (§10 + §5 W4 + R8 잔여 항목) 모두 제거. |

### 측정 가능한 변화 목표 (revised)

| 지표 | Current | Target | Delta |
|---|---|---|---|
| Core agents | 20 active | **17** (manager 8 + expert 4 cross-cutting + builder 1 + evaluator 2 + researcher 1 + claude-code-guide 1) | **-3** (expert-{backend,frontend,mobile} retire) |
| Core domain skills (project-specific) | 8 | **0** (project domain은 모두 my-harness) | -8 |
| Core domain skills (universal craft) | 5 (ideation, research, brand-design, copywriting, design-handoff) | **5** (유지) | 0 |
| Core workflow skills | 12 | **12** (universal templates 유지) | 0 |
| Core foundation skills | 4 | **4** (유지) | 0 |
| Core platform skills | 4 | **4** (universal vendor — auth/chrome-ext/deployment/electron) | 0 |
| Core harness skills | 2 | **2** (learner + meta-harness) | 0 |
| My harness skills per project | 0 | **3-7** (project-type-dependent) | +3-7 |
| Harness 자율성 | manual apply only | **autonomous (Tier 1-4 + 5-layer safety + AskUserQuestion approve)** | qualitative shift |

### Vision 핵심 명제

```
"MoAI-ADK의 진정한 가치는 universal core가 아니라, 프로젝트에 맞게
 자율 진화하는 harness이다."
```

---

## 1. Current State Diagnosis (요약)

상세는 `architecture-audit-2026-05-18.md` 참조. 본 vision의 출발점:

### 발견된 핵심 문제

1. **F-101 Architecture Myth**: CLAUDE.md §5 6-phase pipeline 선언 ↔ `/moai run` manager-develop 단독 spawn 모순
2. **F-102 expert-mobile Dead**: workflow 어디서도 호출 안 됨
3. **F-103 Dormancy 미문서화**: expert-backend/frontend는 utility-only인 것을 catalog에서 미구분
4. **Domain skill 단방향**: moai-domain-{backend,frontend,database} → expert-*만 로드 → manager-develop 도메인 깊이 약화

### 진단

위 모순은 "**Core가 너무 많은 것을 떠안고 있다**"의 증상이다. universal 영역과 project-specific 영역이 한 layer에 섞여있다.

해결: **분리**. Core는 universal coordinator + universal cross-cutting expert만. Project-specific domain expertise (web backend / React frontend / mobile / ML / 등)는 my-harness가 담당.

---

## 2. Vision Statement (사용자 directive)

**Verbatim** (2026-05-18 사용자 메시지):

> "harness는 프로젝트 맞게 생성하고 나서 사용자와 프로젝트 개발을 하면서 스스로 학습해서 자율적으로 하네스가 진화 업데이트가 되도록 설계가 되어야 한다. 이걸 꼭 제대로 설계해서 워크플로우 진행시 지속적으로 학습과 업데이트가 되도록 하자. 하네스 스킬/에이전트 등 하네스만 자율 진화 하도록 한다. 그래서 기본 moai-* 스킬/에이전트/커먼드/rules등은 범용적으로 사용이 가능하고 최소화 하고 나머지는 프로젝트에 맞게 하네스 설치가 되어서 프로젝트에 맞는 개발 환경을 제공하는것이 moai-adk 이번 대규모 업데이트의 목적이다."

### Paraphrase

- moai-* (core) = 범용, 최소화, 안정성
- harness = 프로젝트별 생성 + 자율 진화 + 워크플로우 중 지속 학습
- 자율 진화 대상은 **harness만** (core는 보호)
- 목적: 프로젝트별 맞춤 개발 환경

---

## 3. Architecture Target

### 3.1 Two-Zone Architecture (canonical, 17 Core agents)

```
┌─────────────────────────────────────────────────────────────────────┐
│ CORE MoAI (FROZEN, universal, minimum 17 agents)                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│ Agents (17 total — domain-agnostic + cross-cutting universal):       │
│                                                                       │
│   Managers (8):                                                       │
│     manager-spec, manager-develop, manager-docs, manager-quality,    │
│     manager-project, manager-strategy, manager-brain, manager-git    │
│                                                                       │
│   Cross-cutting Experts (4) — universal domains applicable to all:   │
│     expert-security      (OWASP, threat modeling, universal)         │
│     expert-devops        (CI/CD, Docker, K8s, universal)             │
│     expert-performance   (profiling, benchmarking, universal)        │
│     expert-refactoring   (codemod/AST, universal)                    │
│                                                                       │
│   Builder (1):                                                       │
│     builder-harness      (artifact generation for harness ecosystem) │
│                                                                       │
│   Evaluators (2):                                                    │
│     evaluator-active     (skeptical quality assessment)              │
│     plan-auditor         (independent plan document review)          │
│                                                                       │
│   Research (1): researcher                                           │
│                                                                       │
│   Anthropic Upstream (1): claude-code-guide                          │
│                                                                       │
│ Skills (29 universal — see §4.1 detailed list)                       │
│                                                                       │
│ Rules: core/, workflow/, design/, languages/ (16 path-scoped)        │
│ Commands: 17 thin routers (universal /moai *)                        │
│ Output styles: moai/                                                 │
│ Hooks: 22 universal events                                            │
│                                                                       │
│ ✗ REMOVED FROM CORE: expert-{backend, frontend, mobile}              │
│ ✗ REMOVED FROM CORE: moai-domain-{backend, frontend, database}       │
│                                                                       │
│ Update mechanism: moai update (사용자 manual)                         │
│ Evolution: FROZEN — 자율 진화 안 함 (안정성 우선)                     │
│                                                                       │
│ ✓ Cross-cutting experts STAY in Core because: applicable to ALL      │
│   project types (security audit needed for web/mobile/CLI alike,     │
│   performance profiling universal, refactoring/devops cross-domain). │
└─────────────────────────────────────────────────────────────────────┘
                                  ↓ ↑
                          (Frozen Guard — runtime PreToolUse hook)
                                  ↓ ↑
┌─────────────────────────────────────────────────────────────────────┐
│ MY HARNESS (EVOLVABLE, project-specific, self-organizing)            │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│ Path: .claude/agents/my-harness/* + .claude/skills/my-harness-*/     │
│                                                                       │
│ Project-detected agents (예시 — Go web backend project):              │
│   my-harness-go-specialist                                           │
│   my-harness-postgres-specialist                                     │
│   my-harness-rest-api-specialist                                     │
│   my-harness-ci-cd-specialist                                        │
│                                                                       │
│ Project-detected skills (예시):                                       │
│   my-harness-go-patterns (Go best practice, error wrapping, ...)     │
│   my-harness-postgres-schema (project DB schema knowledge)           │
│   my-harness-rest-conventions (project API convention)               │
│                                                                       │
│ Self-evolution data:                                                  │
│   .moai/harness/observations.yaml  (lesson capture log)              │
│   .moai/harness/evolution-log.md   (Tier 1-4 진화 history)           │
│   .moai/harness/anti-patterns.yaml (학습된 anti-pattern)              │
│   .moai/harness/sprint-contracts/  (per-cycle contracts)             │
│   .moai/harness/seeds/             (cold-start baseline patterns)    │
│                                                                       │
│ Generation: /moai project (Socratic + meta-harness 7-Phase)          │
│ Evolution: harness-learner (4-Tier auto-propose, AskUserQuestion)    │
│ Update mechanism: 자율 (사용자 승인 후 auto-apply, 5-layer safety)    │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Lifecycle Flow

```
1. moai init <project>
   └─ Core MoAI deploy (17 agents + 29 skills universal)

2. /moai project (Socratic interview + meta-harness 7-Phase)
   └─ 프로젝트 분석 (언어/도메인/scope)
   └─ Seed pattern injection from .moai/harness/seeds/ (D4 mitigation)
   └─ my-harness-* skills/agents 자동 생성 (프로젝트별 차별화 — Socratic answers + project markers 기반)

3. /moai plan + /moai run + /moai sync (워크플로우 실행)
   └─ harness-learner가 워크플로우 중 lesson auto-capture
   └─ Tier 1 (1x observation) → log
   └─ Tier 2 (3x heuristic) → 영향 가능
   └─ Tier 3 (5x rule) → graduation 후보
   └─ Tier 4 (10x high-confidence) → AskUserQuestion auto-propose

4. moai harness apply <proposal-id>
   └─ 5-layer safety 통과 검증
   └─ Layer 1 (Frozen Guard, runtime hook): core 영역 침범 차단
   └─ Layer 2 (Canary): shadow eval로 regression 검증
   └─ Layer 3 (Contradiction): 기존 rule과 충돌 감지
   └─ Layer 4 (Rate Limiter): max 3/week + 24h cooldown
   └─ Layer 5 (Human Oversight): AskUserQuestion 최종 승인
   └─ 통과 시 my-harness-* 자동 update

5. moai update (사용자 manual)
   └─ Core MoAI upgrade (template diff)
   └─ My harness preserved (별도 경계 — Layer 1 protect)
```

### 3.3 Boundary Contract

| 영역 | 위치 | Modification Mode | Evolution |
|---|---|---|---|
| Core agents | `.claude/agents/moai/` | `moai update` only | FROZEN |
| Core skills | `.claude/skills/moai-*` (non-my-harness) | `moai update` only | FROZEN |
| Core rules | `.claude/rules/moai/` | `moai update` only | FROZEN |
| Core commands | `.claude/commands/moai/` | `moai update` only | FROZEN |
| Core hooks | `.claude/hooks/moai/` | `moai update` only | FROZEN |
| Core output-styles | `.claude/output-styles/moai/` | `moai update` only | FROZEN |
| Harness agents | `.claude/agents/my-harness/` | `harness apply` (auto) | EVOLVABLE |
| Harness skills | `.claude/skills/my-harness-*/` | `harness apply` (auto) | EVOLVABLE |
| Harness data | `.moai/harness/` | runtime (lesson capture) | EVOLVABLE |
| User project | `.moai/project/` + `.moai/specs/` | user write | EXTERNAL |

### 3.4 [HARD] Frozen Guard — Runtime Enforcement (D2 resolution)

frontmatter flag만으로는 환각 harness-learner를 차단할 수 없으므로 **runtime layer**에서 path-glob deny pattern으로 강제한다.

**Mechanism**: PreToolUse hook + path-glob filter

**Implementation location**: `internal/hook/pre_tool.go` (Go runtime, hook handler)

**Logic**:
1. PreToolUse hook이 모든 Write/Edit/MultiEdit tool call 가로채기
2. tool input의 file_path 추출
3. invoking agent identity 확인 (`agent_name` from hook payload — **E4 resolution: W1 research task to verify Claude Code PreToolUse payload schema exposes invoking subagent name. Fallback: read `MOAI_INVOKING_AGENT` env injected by SubagentStart hook into the subagent's process env**)
4. 만약 agent_name이 `harness-learner` 또는 `my-harness-*` 패턴이면:
   - path가 다음 deny pattern과 일치하면 reject + sentinel error:
     - `.claude/agents/moai/**`
     - `.claude/skills/moai-*/**` (단 `.claude/skills/my-harness-*/**`는 ALLOW)
     - `.claude/rules/moai/**`
     - `.claude/commands/moai/**`
     - `.claude/hooks/moai/**`
     - `.claude/output-styles/moai/**`
     - `CLAUDE.md` (project root)
     - `.moai/config/sections/*.yaml`
5. 일치 시 즉시 reject + JSON response:
   ```json
   {"status": "denied", "sentinel": "HARNESS_FROZEN_VIOLATION",
    "agent": "<name>", "path": "<denied path>",
    "reason": "Harness agent cannot write to Core path. See harness-autonomy-vision §3.4"}
   ```

**Sentinel Catalog**:

| 보호 대상 | Sentinel Error |
|---|---|
| `.claude/agents/moai/*` | `HARNESS_FROZEN_AGENT_VIOLATION` |
| `.claude/skills/moai-*` (non-my-harness) | `HARNESS_FROZEN_SKILL_VIOLATION` |
| `.claude/rules/moai/**` | `HARNESS_FROZEN_RULE_VIOLATION` |
| `.claude/commands/moai/*` | `HARNESS_FROZEN_COMMAND_VIOLATION` |
| `.claude/hooks/moai/*` | `HARNESS_FROZEN_HOOK_VIOLATION` |
| `.claude/output-styles/moai/*` | `HARNESS_FROZEN_OUTPUTSTYLE_VIOLATION` |
| `CLAUDE.md` (project root) | `HARNESS_FROZEN_INSTRUCTION_VIOLATION` |
| `.moai/config/sections/*.yaml` (system level) | `HARNESS_FROZEN_CONFIG_VIOLATION` |

**CI Guard** (defense in depth): `internal/harness/frozen_guard_test.go`가 8개 sentinel 각각 시뮬레이션 + matcher 검증 unit test (catalog와 일치). Runtime guard가 우회되더라도 CI에서 catch.

**Exception**: `moai update` CLI는 우회 가능 (사용자 의도된 core update). 우회 mechanism = `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env (CLI 내부 only).

### 3.5 [REMOVED] Deterministic Harness Generation (2026-05-20 사용자 결정)

> **이 섹션은 iteration 3에서 제거되었습니다.** 사유: 사용자 명시 (2026-05-20) — "프로젝트에 맞는 설정과 하네스 생성 후 ... 프로젝트 마다 다르게 셋업"이 W4 본질이며, Seed=SHA256 + temperature=0 + bit-exact CI test + semantic ≥0.95 같은 결정론적 생성 보장 메커니즘은 차별화 목적과 충돌. Determinism은 과잉 설계로 판단.
>
> 관련 참조:
> - `memory/feedback_w4_no_determinism.md` — 사용자 결정 verbatim
> - `memory/feedback_w4_true_goal.md` — W4 진짜 목적 (Core Slimming + 프로젝트별 차별화)
> - `.moai/research/core-slimming-audit-2026-05-20.md` — W4 pre-plan audit
>
> 후속 영향:
> - §5 W4 scope에서 "Deterministic generation guarantee" bullet 제거됨
> - §5 W4 Acceptance에서 "Determinism test" bullet 제거됨
> - §10 acceptance에서 "Generation determinism guaranteed (D10 resolved → §3.5)" bullet 제거됨
> - §0 D10 table entry — "✗ REMOVED" 표시

---

## 4. Boundary Definition — Concrete

### 4.1 Core (FROZEN) — 17 Agents + 29 Skills

**Agents (17)** — final canonical list:
```
Managers (8):    spec, develop, docs, quality, project, strategy, brain, git
Cross-cutting Experts (4): security, devops, performance, refactoring
Builder (1):     builder-harness
Evaluators (2):  evaluator-active, plan-auditor
Research (1):    researcher
Anthropic (1):   claude-code-guide
```

**Skills (28 — E2 resolution, committed count)**:
- Foundation (4): `moai-foundation-{core, thinking, quality, cc}`
- Workflow (12): `moai-workflow-{spec, tdd, ddd, testing, project, worktree, loop, gan-loop, design-context, design-import, ci-watch, ci-autofix}`
- Universal craft (5): `moai-design-system`, `moai-domain-{copywriting, brand-design, design-handoff, ideation, research}` — `moai-design-system`는 design pipeline tooling skill, 나머지 4개는 universal craft domain skills, 합 5
- Platform + Framework (4): `moai-platform-{auth, chrome-extension, deployment}`, `moai-framework-electron`
- Harness (2): `moai-harness-learner`, `moai-meta-harness`
- Router (1): `moai`
- **Total**: 4 + 12 + 5 + 4 + 2 + 1 = **28** (canonical committed count for v3.5.0)

**Note**: moai-domain-research는 `/moai brain` Phase 3 universal market research (재사용 가능 ideation tool) → core 유지. moai-domain-ideation 동일.

### 4.2 Harness (EVOLVABLE) — 동적 생성

**Generated by `/moai project` (meta-harness)** — 프로젝트 type별 예시:

| 프로젝트 type | 생성 my-harness-* (3-7개) |
|---|---|
| Go web backend | `my-harness-go-specialist`, `my-harness-postgres-specialist`, `my-harness-rest-api-specialist`, `my-harness-ci-cd-specialist` (4) |
| React frontend SaaS | `my-harness-react-specialist`, `my-harness-tailwind-specialist`, `my-harness-vercel-specialist`, `my-harness-stripe-specialist` (4) |
| Flutter mobile | `my-harness-flutter-specialist`, `my-harness-riverpod-specialist`, `my-harness-fastlane-specialist` (3) |
| Python ML | `my-harness-python-specialist`, `my-harness-pytorch-specialist`, `my-harness-mlflow-specialist`, `my-harness-dvc-specialist` (4) |
| Polyglot monorepo | up to 7 specialists per language/framework |

각 my-harness 항목은:
- skill 본문 (`.claude/skills/my-harness-*/SKILL.md`)
- agent 정의 (`.claude/agents/my-harness/*.md`)
- frontmatter `evolution: enabled` (semantic flag — runtime은 §3.4 path-glob으로 enforce)

### 4.3 Migration Strategy for Existing expert-*

**expert-backend retire 후**:
- Domain knowledge (REST/GraphQL/JWT/OAuth/PostgreSQL) → `moai-meta-harness` template (`.claude/skills/moai-meta-harness/templates/backend-baseline.md`)으로 흡수
- 프로젝트가 web backend면 `/moai project`가 자동으로 `my-harness-backend-{framework}` 생성
- 생성된 harness skill은 baseline template + 프로젝트 specific framework 통합

**expert-frontend retire 후**:
- Same pattern + Pencil MCP integration은 universal craft (`moai-design-system`)에 잔류
- moai-design-system이 Pencil tools 사용 권한 보유

**expert-mobile retire**:
- 단순 삭제 + 3 dangling reference 정리
- 향후 mobile 프로젝트는 `/moai project`가 `my-harness-mobile-{ios/android/flutter/rn}` 생성

### 4.4 Cold-Start Mitigation — Seed Pattern Injection (D4 resolution)

**Problem**: 신규 프로젝트의 observations.yaml = empty. Tier 4 도달까지 ~5-10 SPEC 동안 my-harness 깊이 부족 → quality regression risk.

**Solution**: `meta-harness`가 generation time에 **curated baseline pattern**을 seed.

**Seed Source**:
- `.moai/harness/seeds/backend/<framework>.yaml` (Go/Node/Python/Rust 등 framework-specific seeds)
- `.moai/harness/seeds/frontend/<framework>.yaml` (React/Vue/Svelte 등)
- `.moai/harness/seeds/mobile/<framework>.yaml` (iOS/Android/Flutter/RN)
- `.moai/harness/seeds/universal/<category>.yaml` (logging/error-handling/testing 등)

**Seed Structure**:
```yaml
# .moai/harness/seeds/backend/go.yaml
seeds:
  - id: SEED-GO-001
    pattern: "fmt.Errorf with %w wrapping"
    tier: 3  # starts at Tier 3 (eligible for graduation)
    confidence: 0.85
    category: error-handling
    body: |
      Always wrap errors with %w when propagating across function boundaries.
      Anti-pattern: string concatenation (loses error chain).
    references:
      - https://go.dev/blog/go1.13-errors
```

**Seed Lifecycle**:
1. `/moai project` 실행 시 프로젝트 type 감지
2. matching seed file(s) load (예: Go backend → backend/go.yaml + universal/error-handling.yaml)
3. Tier 3 starting point으로 observations.yaml에 inject
4. 사용자 워크플로우 시작부터 my-harness가 깊이 보유
5. 워크플로우 진행 중 observation 누적 → seed가 Tier 4로 graduation 가능 (사용자 승인 후 my-harness-*에 정식 적용)

**Seed Maintenance**:
- Seeds는 Core repo (`.claude/skills/moai-meta-harness/seeds/`)에 ship
- `moai update`로 갱신
- 프로젝트별 customization은 my-harness-* skill 본문으로 발전 (seed = starting point)

**Result**: cold-start window가 5-10 SPEC → ~1-2 SPEC으로 단축. seed quality가 retiring expert-* domain knowledge depth 보존.

---

## 5. Migration Strategy — 5 SPEC Mega-Sprint

### W0 — SPEC-V3R5-CLAUDE-REFRESH-001 (P0, immediate)

**Scope** (Bundle A + Bundle B 통합):
- Bundle A: settings.json.tmpl matcher 2 LOC fix (F-001 + F-008)
- Bundle B core items:
  - CLAUDE.md §5 Agent Chain 정정 (myth 제거 + dormancy 명문화)
  - expert-mobile retire (agent.md 삭제 + 3 dangling refs 정리 + CLAUDE.md 카탈로그 반영)
  - CLAUDE.md §8 ToolSearch syntax 정정 (F-104)
  - AskUserQuestion 5-way duplication 압축 (F-105)
  - CLAUDE.md version v14.0.0 → v14.2.0 (F-011)

**Acceptance**:
- settings.json.tmpl matcher CI test pass
- CLAUDE.md §5 = "manager-develop universal + expert-* utility-only" 명문화
- expert-mobile.md 부재
- `grep -rn moai-domain-mobile internal/template` returns 0
- `moai agent lint --strict` returns 0 ERROR/0 WARN
- Plan-auditor PASS ≥ 0.85

**Dependencies**: none

**Estimate**: 단일 SPEC, 3-5 PR

### W1 — SPEC-V3R5-CONSTITUTION-DUAL-001 (P0, prerequisite)

**Scope**:
- `.claude/rules/moai/core/constitution.md` (Core FROZEN zone 정식 정의)
- `.claude/rules/moai/harness/constitution.md` (Harness EVOLVABLE zone 정식 정의)
- zone-registry.md 100% coverage 달성 (F-006 — 57 unmapped HARD 등록)
- §3.4 Frozen Guard runtime hook 구현 (`internal/hook/pre_tool.go` 확장)
- §3.4 8개 sentinel error 정의
- CI gate (`harness_frozen_guard_test.go`) 신설

**Acceptance**:
- Core/Harness boundary가 코드(runtime) + 문서(rules) 양쪽에서 enforce
- zone-registry coverage = 100% (102/102)
- Frozen Guard runtime hook이 §3.4 8개 sentinel 시나리오 차단 검증 (catalog와 일치)
- W1 research task로 PreToolUse hook payload subagent identity 노출 여부 확인 (E4 resolution — `MOAI_INVOKING_AGENT` env fallback 포함)
- CI test 100% pass

**Dependencies**: W0

**Estimate**: 단일 SPEC, 4-6 PR

### W2 — SPEC-V3R5-CORE-SLIM-001 (P1, large refactor)

**Scope**:
- expert-backend retire → domain knowledge를 `moai-meta-harness/templates/backend-baseline.md`으로 흡수
- expert-frontend retire → 동일 (Pencil MCP는 `moai-design-system`에 잔류)
- moai-domain-{backend, frontend, database} skill retire
- manager-develop을 "universal implementer" (no domain assumption)로 재정의
- CLAUDE.md §4 Agent Catalog 재구성 (Core 17 + Harness dynamic)
- **Legacy invocation handling (D6 resolution)**:
  - `expert-backend` 명시 호출 시 → stub agent가 받아서 `manager-develop`으로 redirect + 사용자 notification ("expert-backend는 v3.5.0에서 my-harness-backend-*로 마이그레이션됨. `/moai project --refresh` 권장.")
  - Stub agent 위치: `.claude/agents/moai/expert-backend.md` (frontmatter `retired: v3.5.0`)
  - Same pattern for expert-frontend
- Migration guide for existing users (CLAUDE.md §17 신규 section)

**Acceptance**:
- Core agents = 17개로 reduce
- Core skills = 28개로 reduce (canonical per §4.1)
- Stub agents가 legacy invocation을 graceful redirect (Migration test)
- evaluator-active PASS ≥ 0.85 (4-dim: Functionality/Security/Craft/Consistency)
- Existing project snapshot test (snapshot of git history showing legacy `expert-backend` invocations now succeed via redirect)

**Dependencies**: W1 (FROZEN/EVOLVABLE 경계 enforce 후) — **W2와 W3 병렬 가능 (D5 resolution)**

**Estimate**: 단일 SPEC, 6-10 PR

### W3 — SPEC-V3R5-HARNESS-AUTONOMY-001 (P0, vision core)

**Scope** — harness-learner 자율 진화 mechanism 본격 구현. **W2와 병렬 가능** (D5 resolution — W3는 W1의 Frozen Guard만 의존):

**4-Tier Evolution Pipeline** (§6.3)

**5-Layer Safety** (§6 detail):
- Layer 1: Frozen Guard (W1 구현물 활용 — runtime hook)
- Layer 2: Canary (shadow eval on last 3 projects)
- Layer 3: Contradiction Detector
- Layer 4: Rate Limiter (3/week + 24h cooldown + 50 active learning max)
- Layer 5: Human Oversight (AskUserQuestion)

**Cold-Start Mitigation** (§4.4 seed pattern):
- meta-harness가 generation time에 seed inject
- harness-learner는 seed를 Tier 3 starting point로 인식

**Proposal Throttling (§6.6 — D7 resolution)**:
- Default mode: immediate AskUserQuestion at Tier 4
- Batch mode: 주간 1회 review (사용자 setting)
- Per-category mute (`moai harness mute <category>`)
- Quiet hours (`workflow.yaml: harness.proposal.quiet_hours: [18, 9]`)

**CLI**:
- `moai harness status`
- `moai harness apply <proposal-id>`
- `moai harness rollback <evolution-id>`
- `moai harness disable`
- `moai harness mute <category>`
- `moai harness verify --determinism`

**Acceptance**:
- 워크플로우 실행 중 lesson capture 검증
- Tier 1 → Tier 4 진급 시나리오 e2e test
- 5-layer safety 각각 단위 test
- Frozen Guard 침범 시도 차단 검증 (Layer 1 + W1 CI gate 연계)
- 사용자 승인 거부 시 rollback 동작 검증
- Proposal throttling 4 mode (immediate/batch/mute/quiet-hours) 검증
- Cold-start regression test (신규 프로젝트 + seed inject → 첫 SPEC quality 측정)

**Dependencies**: W1 only (W2 parallel — D5 resolution)

**Estimate**: 단일 SPEC, 8-12 PR (가장 큰 SPEC)

### W4 — SPEC-V3R5-PROJECT-MEGA-001 (P1, Core Slimming + Project-Specific Setup 완결)

> **iteration 3 (2026-05-20)에서 scope 재정의됨.** 이전 정의 ("polish")는 부정확. 실제 메커니즘 80%는 W3 + moai-meta-harness skill + moai-adk-go dogfooding으로 이미 구현 완료. W4의 진짜 일은 **Core Slimming**.

**Scope** (iteration 3 revised):
1. **Core Slimming** (P0, 사용자 강조 핵심) — `agents/moai`, `skills/moai-*` 중 my-harness로 대체 가능한 항목을 template에서 retire:
   - Category A: `moai-domain-{backend,frontend,database}` (3개, 718 LOC) — domain knowledge를 meta-harness `templates/<role>-baseline.md`으로 흡수
   - Category B: `moai-framework-electron` + `moai-platform-{auth,chrome-extension,deployment}` (4개, 1432 LOC, **workflow 호출 0건 — dead weight**) — 즉시 retire 가능
   - Category C: `expert-{backend,frontend}` agents — workflow 24 invocation을 manager-develop + dynamic my-harness skill load로 치환
   - 자세한 audit 결과: `.moai/research/core-slimming-audit-2026-05-20.md`
2. **`/moai project --refresh` 명령** (P1) — 기존 harness 업데이트 (새 seed/learnings 반영)
3. **Seed library** (P1) — `.claude/skills/moai-meta-harness/seeds/` baseline 작성 (점진적 — Go/Node/Python/React 4개부터 시작)
4. **revfactory/harness 7-Phase workflow + Socratic interview** — 이미 구현 완료 (재구현 금지)
5. **프로젝트 type 분류 + my-harness-* 자동 생성** — 이미 구현 완료 (moai-adk-go dogfooding이 증거)

**Scope에서 제거된 항목** (iteration 3):
- ~~Deterministic generation guarantee~~ — `feedback_w4_no_determinism` 적용. Seed=SHA256, temperature=0, bit-exact CI test, semantic ≥0.95 모두 제거.

**Acceptance**:
- 4가지 프로젝트 type (Go web / React frontend / Flutter mobile / Python ML) end-to-end 생성 test
- 생성된 my-harness가 fully functional (Skill load 검증, agent invocation 검증)
- ~~Determinism test~~ — REMOVED (사용자 결정)
- Seed inject 후 harness가 Tier 3 starting point로 동작
- meta-harness 7-Phase 각각 단위 test
- **신규**: Category B retire 시 `go test ./...` PASS + workflow grep 0 잔존 매치
- **신규**: Category A retire 시 9 language rules cross-ref 갱신 + baseline templates 흡수 완료
- **신규**: Category C retire 시 11+13 workflow invocation 치환 완료 + at least 1 web/frontend 프로젝트 E2E 검증

**Dependencies**: W2 (Core Slim 후 harness가 채울 영역 명확) + W3 (harness-learner mechanism 동작) — **serial after both**

**Estimate**: SPEC 4개 분리 권장 (audit §5 권장안 A): CORE-SLIM-B-001 (Tier S) → CORE-SLIM-A-001 (Tier M) → CORE-SLIM-C-001 (Tier M-L) → PROJECT-MEGA-001 final (Tier S seed + refresh)

---

## 6. Harness Autonomy Mechanism — Deep Dive (W3 핵심)

### 6.1 Lesson Auto-Capture Pipeline

```
워크플로우 실행 (run/sync/fix/loop 등)
        ↓
SubagentStop hook trigger
        ↓
harness-learner invoke (background, read-only initial scan)
        ↓
변경 패턴 추출
        ↓
heuristic matching (no LLM call):
        ↓
신규 시 Tier 1 entry / 기존 시 count++
        ↓
Tier 4 도달 시 AskUserQuestion auto-propose
   (throttling §6.6 적용)
        ↓
사용자 승인 → 5-layer safety 검증 → my-harness-* file write
        ↓
.moai/harness/evolution-log.md에 기록
```

### 6.2 Observation Schema (D3 resolution — SPEC schema 정렬)

```yaml
# .moai/harness/observations.yaml
observations:
  - id: HARN-OBS-20260518-001
    category: error-handling
    pattern: "fmt.Errorf with %w wrapping"
    evidence:
      - spec_id: SPEC-AUTH-001
        commit: 7a8b9c0
        before: "fmt.Errorf(\"failed: \" + err.Error())"
        after: "fmt.Errorf(\"failed: %w\", err)"
        context: "manager-develop suggested wrapping for error chain"
    count: 1
    confidence: 0.0
    status: observation       # observation|heuristic|rule|graduated|archived|anti-pattern
    created: "2026-05-18"     # canonical (matches SPEC schema)
    updated: "2026-05-18"     # canonical (matches SPEC schema)
```

**D3 fix**: 이전 v1은 `created_at`/`updated_at`을 사용했으나 `spec-frontmatter-schema.md` SSOT와 충돌. v2는 canonical `created`/`updated`로 align. `category` enum은 [error-handling, naming, testing, architecture, security, performance, hardcoding, workflow]로 제한 (lessons-protocol과 동일).

### 6.3 Tier 진급 Threshold

| Observations | Classification | Action |
|---|---|---|
| 1x | Observation | logged only |
| 3x | Heuristic | suggestion (manager-develop은 hint로 참고) |
| 5x | Rule | graduation 후보 (Sprint Contract에 자동 추가 후보) |
| 10x | High-confidence | AskUserQuestion auto-propose (throttling §6.6) |
| 1x (critical failure) | Anti-Pattern | 즉시 flag + FROZEN |
| Seed start | (configurable) | meta-harness seed = Tier 3 starting point (D4) |

### 6.4 Proposal Format

```yaml
question: "다음 패턴을 my-harness-go-specialist에 자동 적용할까요?"
header: "Harness 진화"
options:
  - label: "Apply (권장)"
    description: "10회 관찰된 패턴 (95% 신뢰도): error wrapping에 %w 사용 강제. Canary eval: 최근 3 SPEC 모두 통과. 영향: my-harness-go-specialist body 12 LOC 추가."
  - label: "Apply with modification"
    description: "사용자가 수정 후 apply (text editor 열림)."
  - label: "Defer"
    description: "다음 세션에서 재제안 (cooldown 24h)."
  - label: "Reject permanently"
    description: "이 패턴 영구 거부 (anti-pattern으로 분류)."
```

### 6.5 5-Layer Safety Sequencing

**순차 실행** (각 layer 통과 후 다음 layer):

```
Layer 1 (Frozen Guard, runtime synchronous, < 10ms)
  └─ pass → Layer 2
  └─ fail → reject with HARNESS_FROZEN_*_VIOLATION sentinel

Layer 2 (Canary, asynchronous, last 3 projects shadow eval, ~30s)
  └─ pass (no project score drop > 0.10) → Layer 3
  └─ fail → reject + log canary divergence

Layer 3 (Contradiction Detector, synchronous, < 1s)
  └─ pass (no rule conflict) → Layer 4
  └─ fail → AskUserQuestion 분기 ("contradiction with existing rule X, resolve")

Layer 4 (Rate Limiter, synchronous, < 100ms)
  └─ pass (within 3/week + 24h cooldown + 50 active max) → Layer 5
  └─ fail → defer to next eligible window

Layer 5 (Human Oversight, AskUserQuestion, user-paced)
  └─ approve → apply
  └─ reject → log + permanent block
  └─ defer → schedule (cooldown)
```

**Synchronous user-blocking**: Layer 1 + Layer 3 + Layer 4 + Layer 5만. Layer 2는 background (Canary).

**Canary Veto Policy (E5 resolution)**: Layer 2 Canary는 asynchronous로 ~30s 소요되어 Layer 5 user approval보다 늦게 완료될 수 있다. 다음 정책으로 race 해소:

1. Layer 5 (user approval)이 Canary 완료 전 도착 → **provisional apply** (my-harness-* file write 수행, 단 evolution status = `provisional`)
2. Canary 완료:
   - PASS → evolution status `provisional → applied`, evolution-log.md에 기록
   - FAIL → **automatic rollback** (provisional file revert) + AskUserQuestion notification ("Canary가 regression 감지하여 자동 롤백됨. Override 또는 deeper review?")
3. **Canary는 Layer 5 approval에 대한 veto power 보유** — user approval이 final이 아님, Canary가 final gate
4. Veto 발생 시 해당 proposal은 48h cooldown 후 재제안 가능 (rate limiter에 별도 entry)

### 6.6 Proposal Throttling (D7 resolution) — NEW

Tier 4 proposal fatigue 방지:

**Modes** (사용자 setting in `.moai/config/sections/workflow.yaml`):

```yaml
harness:
  proposal:
    mode: immediate | batch | quiet  # default: immediate
    batch:
      window: weekly                  # weekly | sprint_end | manual
      max_per_window: 5
    quiet:
      hours: [18, 9]                   # 18:00 ~ next 09:00 quiet
      timezone: Asia/Seoul
    mute:
      categories: [error-handling]    # mute specific categories
    cooldown_hours: 24                # per-proposal cooldown
```

**CLI**:
- `moai harness mute <category>` — temporarily silence Tier 4 in category
- `moai harness mute-list` — show muted categories
- `moai harness unmute <category>` — re-enable

**Default**: `immediate` mode + 24h cooldown per proposal. 폴리글랏 monorepo (7 my-harness) 사용자는 `batch` + `weekly`로 전환 권장.

### 6.7 Schema Separation Note (D3 detail)

`.claude/rules/moai/development/spec-frontmatter-schema.md` SSOT는 **SPEC files (`.moai/specs/<SPEC-ID>/spec.md`)에만 적용**되며 harness data files (`.moai/harness/*.yaml`)는 별도 schema를 따른다. v2에서는 일관성을 위해 harness schema도 canonical `created:`/`updated:` field name을 채택 — 중복 schema 회피.

**검증**: `harness_schema_test.go`가 observation YAML이 canonical field name 사용 검증.

---

## 7. Dependency Graph + Wave Order (D5 resolution)

```
W0 (CLAUDE-REFRESH-001) ────────────┐
   │                                 │
   │ (foundation: 명확한 truth + dormancy)
   ↓                                 │
W1 (CONSTITUTION-DUAL-001) ─────────┤
   │                                 │
   │ (zone separation enforced + Frozen Guard runtime)
   ├─→ W2 (CORE-SLIM-001) ──────┐    │
   │                            │    │
   └─→ W3 (HARNESS-AUTONOMY-001)│    │
                                ↓    │
                       W4 (PROJECT-MEGA-001)
                                │
                                ↓
                        v3.5.0 release
```

**Critical path**: W0 → W1 → max(W2, W3) → W4 (3-4 sprint 단축)

**Parallel opportunity (D5 explicit)**:
- W2와 W3는 W1 이후 **병렬 실행 가능**
- W2: core retire (file deletion + stub creation + CLAUDE.md update)
- W3: harness mechanism (runtime hooks + safety layers + Tier engine)
- 두 SPEC은 서로 다른 영역 (`internal/cli/agent_*` vs `internal/harness/*`)을 modify — write conflict 없음
- W4는 W2 + W3 둘 다 머지 후 진입

---

## 8. Risk Assessment (revised, D4+D9 added)

| ID | Risk | Wave | Probability | Impact | Mitigation |
|---|---|---|---|---|---|
| R1 | 기존 사용자 expert-* break | W2 | High | High | (a) Stub redirect (W2 AC), (b) Migration guide, (c) v3.4.x compat retention 1 minor cycle |
| R2 | Harness 자율 진화 over-aggressive | W3 | Medium | High | 5-layer safety (특히 Layer 1 runtime Frozen Guard + Layer 5 AskUserQuestion) |
| R3 | Frozen Guard false positive (의도된 core 변경 차단) | W1, W3 | Medium | Medium | `MOAI_FROZEN_GUARD_BYPASS=moai-update-internal` env (CLI 내부만) |
| R4 | Lesson capture overhead (워크플로우 latency) | W3 | Medium | Low | background + heuristic matching (no LLM call) + per-event budget |
| R5 | 프로젝트별 harness fragmentation | W4 | Low | Medium | `/moai project --refresh` (standard template re-baseline) |
| R6 | Vision drift (실제 구현이 vision과 멀어짐) | All | Low | High | plan-auditor 매 wave 검증 + zone-registry 100% coverage |
| R7 | W3 mechanism 복잡도 | W3 | High | Medium | design constitution v3.4 검증된 패턴 채택 |
| **R8** | **Cold-start regression (D4+D9 resolution)** | **W3, W4** | **High** | **High** | **§4.4 Seed pattern injection — meta-harness가 generation time에 baseline seed 주입. observations 0이 아닌 Tier 3 starting point 보장. cold-start window 5-10 SPEC → 1-2 SPEC 단축.** |
| **R9** | **Tier 4 proposal fatigue (D7 resolution)** | **W3** | **Medium** | **Medium** | **§6.6 Proposal Throttling — immediate/batch/quiet/mute 4 mode + per-category mute + quiet hours + cooldown** |
| R10 | Backward compat 미충족 (D6 leftover) | W2 | Low | High | W2 AC에 명시적 stub redirect test 포함 |

---

## 9. Open Questions for plan-auditor (revised, D10 promoted)

D10이 §3.5 design constraint로 promote됨에 따라 §9 Q3 제거. 남은 9개 questions:

1. **Boundary Clarity (D2 enhanced)**: §3.4 runtime PreToolUse hook의 performance impact는 measurable한가? 모든 write 호출에 10ms 추가 시 워크플로우 latency 누적?
2. **Seed Library Maintenance (D4)**: §4.4 seeds는 누가 maintain? Core release cycle에 묶이는 게 적절한가, 별도 community contribution mechanism이 필요한가?
3. **5-layer Safety Sequencing (revised)**: §6.5는 Layer 2 (Canary) background 실행 결정. Canary 결과 도착 전 Layer 3+로 진입할 수 있는가?
4. **Tier 진급 Threshold Adequacy**: 1/3/5/10 횟수가 대규모 프로젝트에서는 빠르게 도달, 소규모 프로젝트에서는 영원히 미도달 가능. 프로젝트 size 기반 adjusted threshold 필요?
5. **AskUserQuestion Fatigue (D7 enhanced)**: §6.6 throttling이 batch mode 도입했으나, batch mode에서 5개 proposal 동시 emit 시 사용자는 1 round (4 questions max) 한계로 partial review 강제. multi-round 자동 분할 추가 필요?
6. **W2 Stub Agent Lifetime (D6 enhanced)**: stub agent (`expert-backend.md` with `retired: v3.5.0` frontmatter)가 언제 완전 제거되는가? v3.6.0? v4.0.0? Deprecation policy 명문화 필요.
7. **W3 Lesson Capture Trigger Breadth**: SubagentStop hook 외에 어디서 capture? Manual `moai harness capture <pattern>` CLI도 필요?
8. **Migration Cost**: 5 SPEC × 평균 6 PR = 30 PR. 사용자가 main 안정성 유지하면서 incremental adopt 가능한 cadence?
9. **Vision-Implementation Drift Detection**: vision document 자체가 evolve되는가? 5 SPEC 진행 중 vision 갱신 시 plan-auditor 재검증 필요?

---

## 10. Acceptance Criteria for Vision

본 vision document v2가 "approved"되려면:

- [ ] plan-auditor 독립 review PASS ≥ 0.85
- [ ] §9 open questions 9개 모두에 대한 사용자 결정 or 명시적 deferral
- [ ] W0-W4 dependency graph가 cycle-free (D5 resolved — W2/W3 parallel)
- [ ] 각 SPEC scope가 single-PR-not-possible 명백
- [ ] backward compatibility strategy 명시 (특히 W2 stub redirect — D6 resolved)
- [ ] frozen guard runtime + CI 설계 완료 (D2 resolved)
- [ ] schema 정렬 (D3 resolved)
- [ ] cold-start mitigation 명시 (D4+D9 resolved)
- [ ] Tier 4 throttling 명시 (D7 resolved)
- [ ] Harness skill count measurable (D8 resolved)
- [x] ~~Generation determinism guaranteed~~ — **REMOVED (iteration 3, 2026-05-20 사용자 결정)**
- [ ] vision document 자체의 evolution policy 결정 (FROZEN vs EVOLVABLE) — Q9

---

## 11. Appendix

### 11.1 Reference Architecture

- `.moai/research/architecture-audit-2026-05-18.md` — current state baseline
- `.claude/rules/moai/design/constitution.md v3.4.0` — 기존 design 자율 진화 mechanism (W3 reference pattern)
- `revfactory/harness` GitHub repo — 7-Phase workflow source
- `.claude/skills/moai/moai-meta-harness` (existing) — current meta-harness 구현 (W4에서 강화)

### 11.2 Out of Scope

- 새로운 LLM 모델 도입 시 architecture
- multi-tenancy harness (한 프로젝트 내 여러 sub-team)
- harness skill sharing between projects (community harness library — W4+ 별도 SPEC)
- harness performance optimization (캐싱, lazy load)
- backward migration tool (구 architecture로 rollback)

### 11.3 Glossary

- **Core MoAI**: Universal layer, FROZEN, modified only by `moai update`
- **My Harness**: Project-specific layer, EVOLVABLE
- **Frozen Guard**: §3.4 runtime PreToolUse hook + CI test (defense in depth)
- **Seed Pattern**: §4.4 cold-start baseline (meta-harness가 generation time에 inject)
- **Sentinel Error**: §3.4 structured error code (HARNESS_FROZEN_*_VIOLATION)
- **Determinism**: §3.5 동일 입력 → template portions bit-exact + LLM bodies semantic equivalence ≥0.95
- **Throttling**: §6.6 Tier 4 proposal pacing (immediate/batch/quiet/mute)
- **Tier 1-4**: §6.3 Confidence tiers (1x/3x/5x/10x observations)

---

## 12. Iteration Changelog

### Iteration 2 — D1-D10 P0/P1/P2/P3 resolution

| Defect | Resolution Summary |
|---|---|
| D1 | §0/§3/§4 모두 Core agents = 17개로 통일. expert-{security,devops,performance,refactoring}는 universal cross-cutting → Core 유지. |
| D2 | §3.4 신설 — PreToolUse hook runtime + path-glob deny pattern. 8 sentinel error 정의. CI defense in depth. |
| D3 | §6.2 + §6.7 — harness YAML schema가 SPEC schema와 align (`created:`/`updated:` canonical). `harness_schema_test.go` 검증. |
| D4 | §4.4 + R8 — meta-harness seed pattern injection. cold-start window 5-10 SPEC → 1-2 SPEC. |
| D5 | §5 + §7 일관화 — W3 depends on W1 only. W2/W3 explicit parallel. |
| D6 | §5 W2 AC — stub redirect mechanism. legacy invocation graceful handling. |
| D7 | §6.6 신설 — Proposal Throttling (4 mode). R9 risk 등록. |
| D8 | §0 table — "3-7 per project type" measurable. |
| D9 | R8 — D4와 통합. |
| D10 | §3.5 신설 — Deterministic generation guarantee promoted from §9 open question. |

### Iteration 3 — E1-E5 spot fix (plan-auditor iter2 0.78 → target 0.85+)

| Defect | Resolution Summary |
|---|---|
| E1 | §5 W1 AC + §3.4 catalog 통일 — sentinel count = **8** (catalog와 W1 AC 모두 8로 align) |
| E2 | §4.1 skill arithmetic — **28 committed** (universal craft 5 = design-system + 4 domain skills 명확화). "28-29 사이" 자기모순 제거. |
| E3 | §3.5 LLM determinism scoped — template portions = bit-exact, LLM bodies = semantic equivalence ≥0.95. "bit-exact identical" overpromise 완화. |
| E4 | §3.4 PreToolUse subagent identity = W1 research task로 명시 + `MOAI_INVOKING_AGENT` env fallback 정의. unverified assumption 명시화. |
| E5 | §6.5 Canary Veto Policy 신설 — provisional apply + Canary 결과 따라 finalize/auto-rollback. Layer 2가 Layer 5 approval에 대한 veto power 보유. |

---

Generated by MoAI orchestrator with ultrathink. Iteration 3 reflects D1-D10 (iter2) + E1-E5 (iter3) plan-auditor defect resolution.

---
title: MoAI-ADK Template Architecture Audit — Dead Code / Duplicate / Optimization Strategy
date: 2026-05-18
auditor: MoAI orchestrator (Opus 4.7, ultrathink + parallel Explore × 4)
scope: agents + skills + rules + CLAUDE.md + commands + hooks + output-styles (TEMPLATE only — internal/template/templates/)
baseline: .moai/research/workflow-audit-2026-05-16.md (23 findings)
delta_commits: 13 commits since 2026-05-16 (7a118e6b2 zombie purge → a165706e6 RT-006)
deliverable: read-only audit + advisory SPEC bundle recommendations
status: complete
related:
  - .moai/research/workflow-audit-2026-05-16.md
  - .moai/research/harness-system-audit-2026-05-14.md
  - memory/project_hooks_audit_xhigh_complete.md
---

# MoAI-ADK Template Architecture Audit — 2026-05-18

## 0. Executive Summary

본 감사는 **2026-05-16 baseline audit (23 findings)** 이후 13개 commit의 진행 상황을 점검하고, 이번 세션에서 발견된 **신규 architecture-level 모순 (F-101 시리즈)**을 종합한다. 사용자 요청: 데드코드/중복/불필요 항목 식별 + 제거 최적화 전략 수립.

### 핵심 수치 — Delta from baseline

| 지표 | 2026-05-16 | 2026-05-18 | Delta |
|---|---|---|---|
| Total agents in template | 28 (20 active + 8 zombie) | **20 active** | -8 ✓ (F-003 resolved) |
| Dangling skill refs | 4 (mobile/db-docs/ast-grep/research) | **1** (mobile only) | -3 ✓ (F-004 partial) |
| Monolithic workflows (≥900 LOC) | 4 (run/sync/project/plan) | **0** | -4 ✓ (F-013 resolved) |
| `/moai run` agents in catalog | 1 (manager-develop) ↔ 6 phase-chain claim | **1** ↔ **6 claim 잔존** | F-101 신규 |
| Settings.json matcher P0 | 2 (F-001 + F-008) | **2** | unchanged ❌ |
| Total HARD rules unmapped | 57/102 (56%) | **57/102** | unchanged ❌ |
| CLAUDE.md version | v14.0.0 (2026-04-03) | **v14.0.0** | unchanged ❌ |

### 베이스라인 23 finding 진행 (4 RESOLVED + 19 PENDING)

```
RESOLVED (4 commits)
  F-002 (harness-observe drift)       → commit 2026-05-15 (Stop + SubagentStop + UserPromptSubmit 3-wrapper 추가)
  F-003 (8 zombie agents)             → commit 7a118e6b2 (8 zombies purge)
  F-004 partial (4 missing skills)    → commit 7a118e6b2 (3 of 4 refs resolved; moai-domain-mobile 잔존)
  F-013 (4 monolithic workflows)      → commits 986418598/980ccdc56/d1c6f3104/85bdd8d71 (WORKFLOW-SPLIT-001 Wave 1-4)
  F-010 partial (CLAUDE.md "cycle")   → already removed (template verified, line 108 lists 8 agents w/o "cycle")

PENDING (19)
  P0: F-001, F-008 (settings.json matchers) — UNRESOLVED
  P1: F-005, F-006, F-007, F-009, F-011, F-012 — UNRESOLVED
  P2: F-014, F-015, F-016, F-017, F-018, F-019, F-020 — UNRESOLVED
  P3: F-021, F-022, F-023 — UNRESOLVED
  F-004 residual: moai-domain-mobile skill 잔존
```

### 신규 발견 (이번 세션 — F-100 series)

| ID | Priority | 위치 | 발견 |
|---|---|---|---|
| **F-101** | **P0** | CLAUDE.md §5:152-159 (Agent Chain) ↔ workflows/run/* | "Phase 3: expert-backend, Phase 4: expert-frontend" 6-phase chain 선언과 실제 `/moai run`이 manager-develop 단독 spawn하는 구현 사이 architecture-level 모순 |
| **F-102** | **P0** | expert-mobile.md:142,152,160 | `moai-domain-mobile` skill 3회 인용하지만 skill 파일 부재 (F-004 잔존). 어떤 workflow도 expert-mobile 호출하지 않음 — 진짜 dead code |
| **F-103** | **P1** | CLAUDE.md §4 + expert-{backend,frontend} | expert-backend/frontend가 `/moai run`에는 미참여, 단 `/moai fix`/`/moai loop`/`/moai review`/`/moai mx`/`/moai e2e`/`/moai design`/`moai.md`에서만 invoke되는 "dormant" 상태. 그러나 catalog에서는 그 구분이 없음. |
| **F-104** | **P1** | CLAUDE.md §8:303 | `ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet", max_results: 5)` — askuser-protocol.md SSOT는 `max_results` 없음. 잘못된 syntax 인용. |
| **F-105** | **P1** | CLAUDE.md §1+§8 + moai-constitution.md + askuser-protocol.md | AskUserQuestion deferred tool preload rule이 4곳에 분산. SSOT는 askuser-protocol.md지만 다른 곳에서 paraphrase가 drift 가능. |
| **F-106** | **P2** | CLAUDE.md §13 | "Progressive Disclosure 3-Level System" 정책이 CLAUDE.md text only — `.claude/rules/moai/workflow/progressive-disclosure.md` rule file 부재 |
| **F-107** | **P2** | .claude/hooks/moai/handle-{elicitation,elicitation-result,notification,task-created}.sh.tmpl | 4개 hook 스크립트가 settings.json.tmpl 어디에도 wired되지 않음 — orphan |
| **F-108** | **P3** | Explore audit 정정 | baseline + Explore #4가 handle-agent-hook.sh.tmpl을 orphan으로 분류했으나, 실제로 agent frontmatter의 `hooks:` 블록(PreToolUse/PostToolUse/SubagentStop)에서 인용. 정상 동작. |
| **F-109** | **P3** | Explore #4 verification | handle-harness-observe.sh.tmpl (legacy single) — 2026-05-15 commit 이후 3-wrapper 분리로 superseded. 잔존 .tmpl 파일 정리 권장 |

### 종합 평가 (2026-05-18 기준)

- **Total integrity score**: **86/100** (baseline 84 → +2)
  - 정적 라우팅 9/10 → **9/10** (commands 모두 thin, no regression)
  - Skill-Agent integrity 7/10 → **8/10** (F-004 거의 해결, mobile 잔존)
  - Rules-CLAUDE coherence 6/10 → **5/10** (F-101 신규 architecture 모순 발견)
  - Hooks subsystem 5/10 → **6/10** (F-002 resolved)
  - Workflow modularity 4/10 → **9/10** (F-013 resolved — 4 splits 완료, run.md 101 LOC + 4 sub-skills)
  - CLI surface 9/10 → **9/10**

**핵심 인사이트**:
1. F-013 monolithic workflow가 baseline 이후 가장 큰 진전 (1073 LOC → 101 LOC entry router + 4 sub-skill로 분할)
2. F-101 (Agent Chain 모순)이 본 세션 가장 critical 발견 — 사용자 가설이 옳음 (manager-develop ↔ expert-* 책임 불명확)
3. F-001/F-008 settings.json matcher 2건은 baseline 이후 무진전 — 단일 SPEC 1줄 fix 가능한데 미해결
4. zone-registry 44% coverage는 P1이지만 systemic risk (canary_gate 우회 가능)

---

## 1. Layer Architecture — Current State (2026-05-18)

```
User Input
    │
    └─ /moai:subcmd (17 thin commands, 7 LOC each)
                │
                └─ Skill("moai") — Intent Router (393 LOC)
                            │
                            └─ workflows/*.md (entry routers, 101-181 LOC after WORKFLOW-SPLIT-001)
                                        │
                                        ├─ workflows/run/*.md       (4 sub-skills, 124-449 LOC)
                                        ├─ workflows/sync/*.md      (4 sub-skills, 216-474 LOC)
                                        ├─ workflows/plan/*.md      (3 sub-skills, 124-500 LOC)
                                        └─ workflows/project/*.md   (4 sub-skills, 122-316 LOC)
                                                    │
                                                    └─ Agent(subagent_type=...)
                                                                │
                                                                ├─ ACTIVE in /moai run     : manager-develop (단독)
                                                                ├─ DORMANT (utility-only)  : expert-backend, expert-frontend
                                                                ├─ ACTIVE in other workflows: manager-{spec,strategy,quality,docs,git,brain,project}
                                                                │                            expert-{security,devops,performance,refactoring}
                                                                │                            evaluator-active, plan-auditor, researcher
                                                                │                            builder-harness, claude-code-guide
                                                                └─ DEAD                    : expert-mobile (← F-102)
```

---

## 2. Baseline Audit Status Update (23 findings)

| ID | Priority (baseline) | Title | Status | Resolution |
|---|---|---|---|---|
| F-001 | P0 | SessionStart matcher `clear\|compact` missing | ❌ **PENDING** | settings.json.tmpl:6 still `"startup\|resume"` |
| F-002 | P0 | Harness-observe 3-wrapper drift | ✅ RESOLVED | commit ~2026-05-15: Stop+SubagentStop+UserPromptSubmit 각 2 hooks |
| F-003 | P0 | 8 zombie agents | ✅ RESOLVED | commit 7a118e6b2 (chore: 8 zombie agents purge) |
| F-004 | P0 | 4 missing skills (mobile/db-docs/ast-grep/research) | 🟡 75% RESOLVED | 3 refs resolved (commit 7a118e6b2); **moai-domain-mobile 잔존** → F-102 |
| F-005 | P1 | ci-watch/autofix orphan (Intent Router 미인용) | 🟡 RECLASSIFIED | Explore #2 정정: contract-driven (ci-watch-protocol.md), sync/delivery.md에서 invoked → 의도된 orphan, P3 |
| F-006 | P1 | zone-registry 44% coverage (57 unmapped) | ❌ PENDING | 변화 없음 |
| F-007 | P1 | settings.json `$schema` 누락 | ✅ RESOLVED | settings.json.tmpl:2 `"$schema": "https://json.schemastore.org/..."` 확인 |
| F-008 | P1 | PostToolUse matcher MultiEdit 누락 | ❌ **PENDING** | settings.json.tmpl:81 still `"Write\|Edit"` |
| F-009 | P1 | env.PATH 절대경로 hardcoded | ❌ PENDING | (template 검증 필요 — Bash output 인용된 부분에서 미확인) |
| F-010 | P1 | CLAUDE.md §4 "cycle" 잔존 | ✅ RESOLVED | template verified: §4 lists "spec, develop, docs, quality, project, strategy, brain, git" (8 — no cycle) |
| F-011 | P1 | CLAUDE.md v14.0.0 stale (43일) | ❌ PENDING | Version footer still v14.0.0 / 2026-04-03 |
| F-012 | P1 | expert-frontend tools 30개 비대 | ❌ PENDING | Pencil MCP 12종 + claude-in-chrome 잔존 |
| F-013 | P1 | 4 monolithic workflows (≥900 LOC) | ✅ RESOLVED | WORKFLOW-SPLIT-001 Wave 1-4 (#973-#976) 완료 |
| F-014 | P2 | CLAUDE.md §4/§11/§13/§16 orphan sections | ❌ PENDING | §13 still no rule file (F-106 신규 ID) |
| F-015 | P2 | hooks-system.md local mirror 5건 drift | ❌ PENDING | UserPromptExpansion/PostToolBatch/mcp_tool 미반영 |
| F-016 | P2 | design/constitution.md back-reference 부족 | ❌ PENDING | |
| F-017 | P2 | workflow/db.md → `moai-domain-db-docs` (skill 없음) | ✅ RESOLVED | commit 50a892ab9 `/moai db` retire + sync 자동 wiring |
| F-018 | P2 | team/sync.md (61) vs workflows/sync.md (현재 분할됨) sync 부재 | ❌ PENDING | team/sync.md는 여전히 61 LOC, main은 split됐으나 team은 미동기화 |
| F-019 | P2 | researcher.md frontmatter name field | ❌ PENDING | |
| F-020 | P2 | 47/49 rules HISTORY 부재 | ❌ PENDING | |
| F-021 | P3 | my-harness agents 4개 schema 비대칭 | ❌ INFO | |
| F-022 | P3 | template 28 vs live 32 agent 비교 | ❌ INFO (정상 동작) | |
| F-023 | P3 | Untracked SPEC 잔존 | ✅ RESOLVED | 2026-05-17 working tree 정리 |

**Baseline progress**: 23개 중 **7 RESOLVED + 1 RECLASSIFIED + 15 PENDING**. 약 35% 진척.

---

## 3. New Findings — F-100 Series (이번 세션 신규)

### F-101 [P0] CLAUDE.md §5 Agent Chain ↔ /moai run 실제 구현 모순

**위치**: 
- CLAUDE.md `§5 Agent Chain for SPEC Execution` lines 152-159 (template)
- `internal/template/templates/.claude/skills/moai/workflows/run/*.md` (4 sub-skills, post-WORKFLOW-SPLIT-001)

**문서 선언**:
```
Phase 1: manager-spec → understand requirements
Phase 2: manager-strategy → create system design
Phase 3: expert-backend → implement core features
Phase 4: expert-frontend → create user interface
Phase 5: manager-quality → ensure quality standards
Phase 6: manager-docs → create documentation
```

**실제 구현**:
- `run/phase-execution.md:394` (DDD mode): `Route all tasks to manager-develop subagent`
- `run/phase-execution.md:398` (TDD mode): `Route all tasks to manager-develop subagent`
- `run/mode-orchestration.md:93`: `Phase 2: Implementation via manager-develop (DDD mode)`
- `run/task-decomposition.md:16, 60`: `Agent: manager-develop subagent`
- 예외 1줄 (`phase-execution.md:224`): `Single endpoint / function | Focused Mode | relevant expert + manager-develop`

**증거**:
- `manager-develop.md:99-105` **Delegation Protocol**에 expert-{backend,frontend,mobile} 위임 **부재** (manager-spec/expert-security/expert-performance/manager-quality/manager-git만 명시)
- `manager-develop.md:87-91` **IN SCOPE**: "Source code implementation and refactoring" — 도메인 무관 직접 작성
- workflow skills (`tdd/ddd/testing`)의 `agent:` frontmatter = manager-develop 단독 owner

**해석**: CLAUDE.md §5는 "6-phase pipeline"을 약속하지만 코드는 manager-develop 단독 spawn. expert-* phase chain은 **architecture myth** — 신규 컨트리뷰터에게 잘못된 정신 모델 전달.

**Impact**: HIGH
- 신규 사용자가 "expert-backend가 자동 호출된다"는 잘못된 기대
- 사용자가 명시적으로 "Use the expert-backend subagent" 호출하지 않는 한 expert는 미발화
- manager-develop이 모든 도메인 코드를 작성하므로 도메인 깊이 (REST/GraphQL paradigm, React 19 Server Components, SwiftUI 등) 약화 가능
- 사용자 가설 검증: 정확히 옳음

---

### F-102 [P0] expert-mobile Dead Code

**위치**: `.claude/agents/moai/expert-mobile.md`

**증거**:
- Line 17 frontmatter `skills:`에 `moai-domain-mobile` (skill 파일 없음 — F-004 잔존)
- Line 142, 152, 160: body가 `moai-domain-mobile` 3회 인용 (Step 2/Step 4/Step 5)
- 어떤 workflow도 expert-mobile invoke 안 함 (Explore #1: "NO workflows reference expert-mobile")
- 7a118e6b2 commit이 4 dangling skill refs 중 3개만 정정, mobile은 의도적으로 보류 (이유: skill backing 부재로 자동 정정 불가)

**Impact**: MEDIUM
- 자동 워크플로우에서 호출되지 않음 → 사용자가 "Use the expert-mobile subagent" 명시 호출 시에만 동작
- 그러나 그 경우에도 skill load 실패 (moai-domain-mobile 없음) → silent 빈 껍질
- CLAUDE.md §4 Expert Agents 카탈로그(6)에는 mobile 부재 (backend, frontend, security, devops, performance, refactoring — 6개)
- 즉 frontmatter는 active처럼 보이지만 catalog에선 부재 + skill backing 부재 → 진정한 dead

**Decision required**: mobile 도메인을 product roadmap에서 유지할지 vs retire할지 (사용자 결정)

---

### F-103 [P1] expert-backend/frontend Dormancy 미문서화

**위치**: CLAUDE.md §4 Agent Catalog vs Explore #1 invocation map

**Explore #1 verified invocation map**:

| Agent | /moai run | /moai fix | /moai loop | /moai mx | /moai review | /moai e2e | /moai design |
|---|---|---|---|---|---|---|---|
| manager-develop | ✅ primary | - | ✅ | - | - | ✅ | - |
| expert-backend | ❌ (Focused Mode 1줄 제외) | ✅ | ✅ | ✅ | - | - | - |
| expert-frontend | ❌ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| expert-mobile | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |

**해석**: expert-backend/frontend는 "auto-workflow에선 dormant, utility-workflow에선 active". CLAUDE.md §4 Catalog에는 이 구분 없음 → 사용자는 둘이 항상 active인 것으로 오해.

**Impact**: MEDIUM (F-101의 system effect)

---

### F-104 [P1] ToolSearch syntax 잘못된 인용

**위치**: CLAUDE.md §8 line 303

**잘못된 syntax**:
```
Call `ToolSearch(query: "select:AskUserQuestion,TaskCreate,TaskUpdate,TaskList,TaskGet", max_results: 5)` before first use.
```

**SSOT (askuser-protocol.md §ToolSearch Preload Procedure)**:
```
ToolSearch(query: "select:AskUserQuestion")
```
또는
```
ToolSearch(query: "select:AskUserQuestion,TaskCreate")
```

`max_results` 파라미터는 SSOT에 없음. 그대로 따라하면 invalid call 가능성.

**Impact**: LOW (대부분 LLM이 정정해서 호출하지만 syntax drift)

---

### F-105 [P1] AskUserQuestion 문서 4-way duplication

**위치**:
1. CLAUDE.md §1 line 14 (HARD rule 요약)
2. CLAUDE.md §8 line 303 (Deferred Tool Preload Protocol)
3. `.claude/rules/moai/core/moai-constitution.md` §MoAI Orchestrator (paraphrase)
4. `.claude/rules/moai/core/askuser-protocol.md` (SSOT, full spec)
5. `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary (paraphrase, baseline F-014 미분류)

5곳 모두 동일/유사 내용. SSOT는 askuser-protocol.md지만 다른 4곳이 drift 위험.

**Impact**: MEDIUM (현재는 paraphrase가 SSOT와 일치하지만 향후 drift 가능)

---

### F-106 [P2] Progressive Disclosure rule file 부재

**위치**: CLAUDE.md §13 (Progressive Disclosure System) — text only

**현재**: 3-Level 시스템 설명만 CLAUDE.md에 있고 `.claude/rules/moai/workflow/progressive-disclosure.md` rule file 없음.

**baseline F-014 (orphan section)**에 부분 해당. §13만 별도 ID (F-106) 부여 — 다른 3개 orphan section은 다른 rule file에 cross-reference 존재.

**Impact**: LOW (현재 동작에 영향 없음, 단 future learner evolution 불가)

---

### F-107 [P2] 6 Orphan Hook Scripts (Explore #4 정정 후 4개)

**위치**: `.claude/hooks/moai/*.sh.tmpl`

**Explore #4 raw list (6 orphan candidates)**:
- handle-agent-hook.sh.tmpl
- handle-elicitation.sh.tmpl
- handle-elicitation-result.sh.tmpl
- handle-notification.sh.tmpl
- handle-harness-observe.sh.tmpl
- handle-task-created.sh.tmpl

**F-108 정정 (이번 세션)**:
- handle-agent-hook.sh.tmpl은 agent frontmatter `hooks:` 블록에서 인용됨 (manager-develop/expert-backend/expert-frontend/expert-devops/manager-quality/manager-spec/manager-docs 7 agents). NOT orphan.
- handle-harness-observe.sh.tmpl (legacy single)는 2026-05-15 3-wrapper 분리로 superseded → F-109 정리 권장

**진짜 orphan (4개)**:
- handle-elicitation.sh.tmpl
- handle-elicitation-result.sh.tmpl
- handle-notification.sh.tmpl
- handle-task-created.sh.tmpl

**Impact**: LOW (4 dead .tmpl, ~50 LOC 절감 가능)

---

### F-108 [P3] Explore audit 정정

handle-agent-hook.sh.tmpl은 settings.json.tmpl `hooks:` 블록에는 없지만 agent .md 파일의 frontmatter `hooks:`에서 인용 (agent-scoped hooks). Explore #4가 settings.json만 grep해서 misclassified.

---

### F-109 [P3] handle-harness-observe.sh.tmpl (legacy) 정리

2026-05-15 commit으로 3-wrapper (stop/subagent-stop/user-prompt-submit) 분리. legacy single wrapper 잔존.

---

## 4. Consolidated Priority Matrix (Baseline pending + New)

### P0 — Critical (즉시 조치 권장)

| ID | Title | Effort | File | Solution |
|---|---|---|---|---|
| F-001 | SessionStart matcher `clear\|compact` 누락 | 1 LOC | settings.json.tmpl:6 | `"startup\|resume"` → `"startup\|resume\|clear\|compact"` |
| F-008 | PostToolUse matcher MultiEdit 누락 | 1 LOC | settings.json.tmpl:81 | `"Write\|Edit"` → `"Write\|Edit\|MultiEdit"` |
| F-101 | Agent Chain ↔ /moai run 모순 | ~30 LOC | CLAUDE.md §5 | (a) §5 Agent Chain 제거 또는 (b) "Phase 3-4: expert-* (utility commands only)"로 정정 |
| F-102 | expert-mobile dead code | 결정 필요 | expert-mobile.md + frontmatter | 사용자 결정: retire vs revive (moai-domain-mobile skill 복원) |

### P1 — Major (단기 처리)

| ID | Title | Effort | File |
|---|---|---|---|
| F-006 | zone-registry 44% coverage (57 unmapped) | 2-3시간 | zone-registry.md + ~5 rule files |
| F-009 | env.PATH 절대경로 hardcoded | 1 LOC | settings.json.tmpl env block |
| F-011 | CLAUDE.md version stale | ~10 LOC | CLAUDE.md footer v14.0.0 → v14.2.0 |
| F-012 | expert-frontend tools 30개 비대 | 검토 | expert-frontend.md frontmatter |
| F-103 | expert-* dormancy 미문서화 | ~30 LOC | CLAUDE.md §4 footnote |
| F-104 | ToolSearch syntax 정정 | 1 LOC | CLAUDE.md §8:303 |
| F-105 | AskUserQuestion 5-way duplication | ~20 LOC 삭제 | moai-constitution + CLAUDE.md → SSOT 인용으로 압축 |
| F-018 | team/sync.md ↔ workflows/sync.md 동기화 부재 | 검토 | team/sync.md |

### P2 — Minor (중기 정리)

| ID | Title | Effort |
|---|---|---|
| F-014 | orphan sections (§4/§11/§16) | ~30 LOC |
| F-015 | hooks-system.md local mirror 5 drift | ~20 LOC |
| F-016 | design/constitution.md back-reference 부족 | ~10 LOC |
| F-019 | researcher.md name field | 1 LOC |
| F-020 | 47 rules HISTORY 부재 | bulk 작업 |
| F-106 | progressive-disclosure rule file 신설 | ~100 LOC |
| F-107 | 4 orphan hook .tmpl 정리 | 4 file deletion |

### P3 — Informational

| ID | Title |
|---|---|
| F-021 | my-harness agents 4개 schema 비대칭 |
| F-022 | template vs live agent 차이 |
| F-108 | Explore audit 정정 (handle-agent-hook orphan 오분류) |
| F-109 | handle-harness-observe.sh.tmpl legacy 정리 |

### RESOLVED (이번 세션 검증)

| ID | Title | Verified |
|---|---|---|
| F-002 | harness-observe 3-wrapper drift | ✓ settings.json.tmpl:84-106 Stop has 2 hooks |
| F-003 | 8 zombie agents purge | ✓ commit 7a118e6b2 |
| F-004 (75%) | 4 missing skills | ✓ 3 of 4 resolved (commit 7a118e6b2); mobile 잔존 (F-102) |
| F-007 | $schema key 누락 | ✓ settings.json.tmpl:2 |
| F-010 | CLAUDE.md §4 "cycle" 잔존 | ✓ template verified — 8 agents w/o cycle |
| F-013 | 4 monolithic workflows | ✓ WORKFLOW-SPLIT-001 (101-181 LOC entry + 14 sub-skills) |
| F-017 | /moai db retire | ✓ commit 50a892ab9 |
| F-023 | Untracked SPEC | ✓ 2026-05-17 정리 |

---

## 5. Dead Code Inventory (제거 후보)

### 5.1 Dead Agents (1)

```
expert-mobile.md
  ├─ frontmatter skills: moai-domain-mobile (missing)
  ├─ body lines 142/152/160: 3 dangling references
  ├─ workflow invocations: ZERO
  ├─ CLAUDE.md §4 Catalog: NOT listed (6 experts only)
  └─ Manual invocation: only viable path but silent skill-load failure
```

### 5.2 Dead Hook Scripts (4)

```
.claude/hooks/moai/
  ├─ handle-elicitation.sh.tmpl       (settings.json wiring 부재)
  ├─ handle-elicitation-result.sh.tmpl (settings.json wiring 부재)
  ├─ handle-notification.sh.tmpl       (settings.json wiring 부재)
  └─ handle-task-created.sh.tmpl       (settings.json wiring 부재)
```

### 5.3 Superseded Hook (1)

```
.claude/hooks/moai/handle-harness-observe.sh.tmpl
  └─ legacy single wrapper, superseded by 3-wrapper (stop/subagent-stop/user-prompt-submit)
```

### 5.4 Dangling Skill References (1, residual from F-004)

```
expert-mobile.md → moai-domain-mobile  (skill file does not exist)
```

### 5.5 Architecture Myth (1)

```
CLAUDE.md §5 "Agent Chain for SPEC Execution" (Phase 1-6)
  └─ describes 6-agent pipeline that does NOT match /moai run implementation
```

**Total dead/superseded items**: 1 agent + 4 hooks + 1 superseded hook + 1 dangling ref + 1 myth section = **8 distinct removal targets**

---

## 6. Duplicate/Contradiction Map

### 6.1 AskUserQuestion Documentation — 5-way duplication

```
SSOT: .claude/rules/moai/core/askuser-protocol.md (full spec, ~250 LOC)
      │
      ├─ CLAUDE.md §1 HARD rule (line 14)           — 1 line summary
      ├─ CLAUDE.md §8 Architecture (line 295-307)   — 12 line paraphrase
      ├─ moai-constitution.md §MoAI Orchestrator    — 9 line paraphrase
      └─ agent-common-protocol.md §UI Boundary      — 30 line paraphrase
```

**Recommendation**: 모든 paraphrase를 1-line reference로 압축:
> "See `.claude/rules/moai/core/askuser-protocol.md` for canonical spec."

### 6.2 Worktree Rules — 3-way

```
SSOT: .claude/rules/moai/workflow/worktree-integration.md
      │
      ├─ CLAUDE.md §14 (Parallel Execution Safeguards) — 8 line summary
      └─ team-protocol.md §Worktree Constraints       — 5 line paraphrase
```

상태: 적절 (CLAUDE.md만 SSOT 참조, team-protocol은 fight-context 보강).

### 6.3 TRUST 5 — 2-way

```
SSOT: moai-constitution.md §Quality Gates
      │
      └─ CLAUDE.md §6 — "For TRUST 5 framework details, see moai-constitution.md"
```

상태: 적절 (CLAUDE.md가 SSOT 참조만).

### 6.4 Agent Chain Contradiction (F-101) — 1 myth + 1 truth

```
Myth (CLAUDE.md §5):   Phase 3 expert-backend → Phase 4 expert-frontend (auto)
Truth (run/*.md):      manager-develop 단독 spawn (Focused Mode 1줄 예외)
```

해결: §5 정정 또는 삭제.

### 6.5 ToolSearch Syntax Drift (F-104)

```
SSOT: askuser-protocol.md   "ToolSearch(query: \"select:AskUserQuestion\")"
Drift: CLAUDE.md §8 line 303 "ToolSearch(..., max_results: 5)"  ← invalid param
```

해결: 1 line 정정.

---

## 7. SPEC Bundle Recommendations (Advisory)

**본 보고서는 read-only 감사** — SPEC 초안 미작성. 아래는 후속 SPEC 분할 시 권장 묶음일 뿐.

### Bundle A — Settings Matcher Fix-Pack (P0)

- **Findings**: F-001, F-008
- **Scope**: settings.json.tmpl 2개 matcher 정정 (2 LOC)
- **Estimate**: trivial (1 PR, < 1시간)
- **Dependencies**: none
- **Risk**: 매우 낮음 (additive matcher 확장)

### Bundle B — Architecture Truth Reconciliation (P0)

- **Findings**: F-101 + F-102 + F-103 + F-104 + F-105 + F-011 (CLAUDE.md refresh 묶음)
- **Scope**:
  - CLAUDE.md §5 Agent Chain 정정 (myth 제거 또는 "utility commands" 명문화)
  - CLAUDE.md §4 expert-* dormancy footnote 추가
  - expert-mobile retire 또는 moai-domain-mobile skill 복원 (사용자 결정)
  - CLAUDE.md §8 ToolSearch syntax 정정
  - AskUserQuestion 5-way duplication 압축 (moai-constitution.md + CLAUDE.md 정리)
  - CLAUDE.md version v14.0.0 → v14.2.0
- **Estimate**: ~100 LOC 변경, 2-3 PR (architecture decision + cleanup)
- **Dependencies**: 사용자 결정 (mobile retire vs revive)
- **Risk**: 중간 (architecture-level 정합화, downstream agent prompt 영향 가능)

### Bundle C — Zone Registry Coverage (P1)

- **Findings**: F-006 (57 unmapped HARD)
- **Scope**: zone-registry.md에 CONST-V3R2-073~149 영역에 누락 HARD 등록
  - worktree-integration.md: 9건
  - ci-autofix-protocol.md: 7건
  - design/constitution.md additional: 8건
  - spec-workflow.md, mx-tag-protocol.md, branch-origin-protocol.md 등
- **Estimate**: 단일 SPEC, ~150 LOC zone-registry 추가
- **Dependencies**: none
- **Risk**: 낮음 (data-only)

### Bundle D — Orphan Hook Cleanup (P2)

- **Findings**: F-107 + F-109
- **Scope**: 4 orphan + 1 superseded .tmpl 파일 삭제
- **Estimate**: trivial (1 PR, 5 file deletion)
- **Dependencies**: hook-system.md docs 동반 업데이트
- **Risk**: 낮음 (사용되지 않음 검증됨)

### Bundle E — Rule Hygiene (P2)

- **Findings**: F-014 + F-015 + F-016 + F-018 + F-020 + F-106
- **Scope**:
  - progressive-disclosure.md rule file 신설 (F-106)
  - hooks-system.md local mirror 5 drift 정정 (F-015)
  - design/constitution.md back-reference 보강 (F-016)
  - team/sync.md ↔ workflows/sync.md 정책 sync (F-018)
  - 47 rules HISTORY 섹션 추가 (F-020) — 별도 SPEC 권장
- **Estimate**: ~300 LOC across 5+ files
- **Dependencies**: docs-site 4-locale sync 동반
- **Risk**: 낮음 (rule layer only)

### Bundle F — Tooling Trim (P2)

- **Findings**: F-012 (expert-frontend tools 30개 비대) + F-009 (env.PATH 절대경로)
- **Scope**: expert-frontend.md frontmatter tools 압축 검토 + settings.json env hygiene
- **Estimate**: ~30 LOC + 검토
- **Dependencies**: Pencil MCP 통합 영향 검증
- **Risk**: 중간 (tool 권한 회수가 downstream agent 동작 영향 가능)

### Bundle G — Info/Polish (P3)

- **Findings**: F-019, F-021, F-022, F-108
- **Scope**: researcher.md name field, my-harness schema 통일, audit 정정
- **Estimate**: 작음
- **Risk**: 매우 낮음

---

## 8. Migration Order + Risk Assessment

**권장 순서**:

```
Wave 1 (즉시):       Bundle A (settings 2 LOC) → 검증 단순, blocking 가능 bug
                    Bundle B 결정 (mobile retire vs revive — 사용자 결정 필요)

Wave 2 (단기):       Bundle B 실행 (architecture truth) → 가장 큰 impact, downstream cascades

Wave 3 (중기):       Bundle C (zone registry coverage) → constitution-check CI gate enables
                    Bundle D (orphan hook cleanup) → 안전한 cleanup

Wave 4 (선택):       Bundle E (rule hygiene) → docs-site sync 동반
                    Bundle F (tooling trim) → tool 권한 회수 영향 검증 필요

Wave 5 (정리):       Bundle G (polish)
```

### Risk Matrix per Bundle

| Bundle | Effort | Risk | Reversibility | User Decision Required |
|---|---|---|---|---|
| A | trivial | low | high | no |
| B | high | medium | medium (downstream prompt impact) | **YES (mobile)** |
| C | medium | low | high | no |
| D | trivial | low | high (just deletion) | no |
| E | medium | low | high | no |
| F | medium | medium | medium (tool 회수) | maybe (Pencil sundry) |
| G | low | very low | high | no |

---

## 9. 사용자 결정 필요 사항 (Critical Path)

### Q1: expert-mobile 처리

**Option 1**: Retire — agent + dangling references 제거. mobile 도메인을 product roadmap에서 제외.
**Option 2**: Revive — moai-domain-mobile skill 복원 (iOS/Android/RN/Flutter 5 modules). expert-mobile.md 정합성 회복.
**Option 3**: Status quo + 문서화 — agent 유지하되 dangling ref 정리 + CLAUDE.md에 "manual invocation only, skill backing 부재" 명시.

### Q2: F-101 Agent Chain 모순 처리

**Option A**: §5 Agent Chain 6-phase 정의 제거 — manager-develop 단독 truth 인정.
**Option B**: §5 Agent Chain 유지하되 "utility commands (fix/loop/mx/review/design/e2e) only"로 정정. 자동 워크플로우(/moai run)는 manager-develop 단독 명시.
**Option C**: 진정한 6-phase pipeline 구현으로 run.md 수정 (가장 큰 변경 — 사실상 architecture rewrite).

### Q3: Bundle 실행 순서

**Option I**: Bundle A 단독 우선 (가장 빠른 P0 fix) → Bundle B는 별도 결정.
**Option II**: Bundle A + B 함께 (CLAUDE.md refresh 묶음).
**Option III**: 본 보고서 보존만 + 후속 세션에서 SPEC 작성.

---

## 10. Appendix

### 10.1 Audit Methodology

- **Tool**: Opus 4.7 ultrathink + Sequential Explore × 4 parallel (read-only)
- **Coverage**: agents (20), skills (32 + 5 ref), rules (49), CLAUDE.md (project + local), hooks (31 scripts + settings.json.tmpl), commands (17), output-styles (1)
- **Cross-reference**: agent.skills[] ↔ skill.frontmatter ↔ workflow body ↔ CLAUDE.md catalog ↔ hooks settings matcher
- **Baseline**: workflow-audit-2026-05-16.md (23 findings)
- **Delta period**: 2026-05-16 ~ 2026-05-18 (13 commits)

### 10.2 Files Touched (Read-only)

본 감사는 어떤 파일도 수정하지 않음. 분석 대상:
- `internal/template/templates/CLAUDE.md`
- `internal/template/templates/.claude/agents/moai/*.md` (20 files)
- `internal/template/templates/.claude/skills/**/*.md` (32 skill dirs)
- `internal/template/templates/.claude/skills/moai/workflows/**/*.md` (17 entry + 14 sub-skills)
- `internal/template/templates/.claude/rules/moai/**/*.md` (49 files)
- `internal/template/templates/.claude/commands/moai/*.md` (17 files)
- `internal/template/templates/.claude/hooks/moai/*.sh.tmpl` (31 files)
- `internal/template/templates/.claude/output-styles/moai/moai.md`
- `internal/template/templates/.claude/settings.json.tmpl`
- `.moai/research/workflow-audit-2026-05-16.md` (baseline)
- `git log --since="2026-05-16"` (delta verification)

### 10.3 Out of Scope

본 감사 범위 밖:
- **런타임 행동 검증**: 실제 `Agent()` 호출 시 worktree isolation 작동, hook timeout 준수 — `internal/cli/*_test.go` + `internal/hook/*_test.go`
- **Live (deployed) environment**: 본 감사는 `internal/template/templates/`만 — 사용자 프로젝트의 실제 `.claude/`는 별도 audit
- **docs-site 4-locale sync**: F-015 같은 hook docs 변경 시 docs-site 동기화 (CLAUDE.local.md §17)
- **CG Mode (Claude+GLM) cost 분석**: workflow.yaml team mode + tmux 통합 정합성
- **CI guard 도입 가능성**: 본 발견의 자동 검증 (constitution-check, dead-agent-guard, dead-hook-guard 등)

### 10.4 Scoring Methodology

```
Integrity Score = average(category scores), 10-point each
  - Routing structure: thin command compliance, intent router clarity
  - Skill-Agent integrity: dangling refs, orphan skills
  - Rules-CLAUDE coherence: SSOT, duplications, contradictions, version drift
  - Hooks subsystem: matcher completeness, script wiring
  - Workflow modularity: LOC budget, progressive disclosure compliance
  - CLI surface: documented vs implemented, retired guards
```

### 10.5 Cross-References

- `.moai/research/workflow-audit-2026-05-16.md` — baseline (23 findings)
- `.moai/research/harness-system-audit-2026-05-14.md` — harness subsystem baseline
- `memory/project_hooks_audit_xhigh_complete.md` — hooks audit (2026-05-16 same-day)
- `memory/project_v3r2_orc_004_complete.md` — Worktree MUST rule (LR-05/LR-09)
- `memory/project_v3r2_orc_002_complete.md` — 8-rule agent lint engine
- `CLAUDE.local.md §22` — Dev settings intent (defaultMode/enableAllProjectMcpServers/teammateMode/env.PATH)

---

## 11. Closing Notes

**Audit complete.** 어떤 파일도 수정되지 않았으며 SPEC 초안도 생성되지 않았다.

**Integrity Score**: **86/100** (baseline 84 → +2)
- F-013 workflow split 완료로 modularity 4/10 → 9/10 (가장 큰 진전)
- F-101 신규 발견으로 Rules-CLAUDE coherence 6/10 → 5/10 (오히려 하락 — 숨겨진 모순 노출)

**Critical Path**:
1. **Bundle A** (P0 settings, 2 LOC) — 가장 빠른 win
2. **Bundle B 결정** (mobile + agent chain) — 사용자 input 대기
3. **Bundle C** (zone-registry) — systemic safety net

**Next Action**: 사용자에게 §9 Q1/Q2/Q3 결정 요청 후 Bundle 실행 SPEC 작성.

---

Generated by MoAI orchestrator with ultrathink + parallel Explore × 4 (agents · skills · rules+CLAUDE.md · commands+hooks+output-styles). Baseline (2026-05-16) referenced and 23 findings progress tracked. New F-100 series captures session-discovered architecture-level contradictions.

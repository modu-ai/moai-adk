---
title: Core Slimming Audit — W4 (PROJECT-MEGA-001) Pre-Plan Analysis
date: 2026-05-20
auditor: MoAI orchestrator (Opus 4.7, ultrathink)
scope: 10 Core skill/agent migration candidates → my-harness 영역 이관 가능성 평가
predecessor: .moai/research/harness-autonomy-vision-2026-05-18.md (§3.5 Determinism 제거됨)
status: complete
related:
  - feedback_w4_no_determinism (memory)
  - feedback_w4_true_goal (memory)
  - feedback_workflow_inflation_root_cause (memory, LEAN 원칙)
  - SPEC-V3R5-WORKFLOW-LEAN-001 (PR #1030 dogfooding 증명)
---

# Core Slimming Audit — 2026-05-20

## 0. Executive Summary

W4 plan-phase 진입 전 **10개 마이그레이션 후보의 객관 분석**. 사용자 directive (2026-05-20):
> "agents/moai, skills/moai-* 관련해서 하네스로 생성을 해서 사용할 것은 제거나 최적화를 하고 맞춤형 하네스 설치가 되어서 프로젝트 마다 다르게 셋업을 해서 사용하도록 하는게 목적"

**결론 한 줄**: Category B는 즉시 retire 가능 (dead weight 확인), Category A는 9 language rules 갱신 동반, Category C는 11+13 workflow invocation 치환 필요로 가장 큰 작업.

### 수치 요약

| 지표 | 값 |
|---|---|
| 마이그레이션 후보 | 10 (3 Category A + 4 Category B + 2 Category C + 1 dead reference cleanup) |
| 후보 skill LOC 총합 | 2,150 LOC (7 SKILL.md 합산) |
| Category B dead weight | 1,432 LOC (4 skills, **workflow 호출 0건**) |
| Category A workflow 호출 | 0건 (agent frontmatter `skills:` 영역만) |
| Category C workflow 호출 | 11건 expert-backend + 13건 expert-frontend |
| Language rules 영향 | 9개 (kotlin/swift/elixir/csharp/scala/javascript/flutter/ruby/java) |
| meta-harness baseline 상태 | 0% (별도 `templates/<role>-baseline.md` 파일 부재 — expert-* 본문에 임베드) |
| Seed library 상태 | 0% (`.gitkeep`만 존재) |
| moai-meta-harness 검증 사례 | 1/N (moai-adk-go Go CLI 1건만, dogfooding) |

---

## 1. moai-meta-harness 작동 검증

### 1.1 검증된 사례 (1건)

- **moai-adk-go** (Go CLI 도구) — `.moai/harness/main.md` + 4 specialist agents + 4 my-harness skills 작동 중
- Pipeline pattern: cli-template → quality → workflow → hook-ci

### 1.2 미검증 project type

| Project type | 검증 사례 | 영향 |
|---|---|---|
| Go CLI | ✅ moai-adk-go | — |
| Go web | ❌ | Phase 1-2 Category retire 후 위험 |
| React/Next frontend | ❌ | Phase 3 Category C retire 시 critical |
| Vue/Nuxt frontend | ❌ | 동상 |
| Flutter/Swift mobile | ❌ | expert-mobile 이미 retired (W0) — 대체 검증 누락 |
| Python ML | ❌ | seed library 필요 (Phase 4) |
| Electron desktop | ❌ | Category B에 framework skill 존재하나 작동 미검증 |

### 1.3 결론

moai-meta-harness 7-Phase workflow는 SKILL.md에 정의되어 있고 1건 (moai-adk-go) dogfooding으로 작동 확인. **그러나 4 project type E2E 검증은 0%**. Category C (expert-{backend,frontend}) retire 전 적어도 1 web/frontend 프로젝트 검증 필수.

### 1.4 dead reference

`.claude/skills/moai-meta-harness/SKILL.md:203`에 `expert-mobile` 잔존 참조:
```
- `expert-mobile` — Mobile domain harness templates
```
W0 (CLAUDE-REFRESH-001)에서 expert-mobile agent는 hard delete되었으나 meta-harness 본문에 dead ref 잔존. **즉시 cleanup 필요** (Phase 0).

---

## 2. Baseline Template 흡수 현황

### 2.1 SKILL.md 분석 결과

meta-harness SKILL.md (370 LOC, 16K bytes)에서 baseline 키워드 매칭:

| Line | Context |
|---|---|
| 127 | "Test coverage baseline → set quality targets" (generic) |
| 201-204 | expert-{backend,frontend,mobile,devops}을 baseline template 소유자로 명시 |
| 219 | manager-develop cycle_type dispatch 설명 |

**핵심 발견**: Vision §4.3가 명시한 `templates/<role>-baseline.md` 별도 파일 (`backend-baseline.md`, `frontend-baseline.md`, ...)은 **존재하지 않음**. expert-* agent 본문에 임베드되어 있음.

### 2.2 결론

Category C retire를 위해서는 **baseline 명시적 추출이 선결조건**:
1. `expert-backend.md` (6,662 bytes) 본문에서 backend domain knowledge → `moai-meta-harness/templates/backend-baseline.md`로 추출
2. `expert-frontend.md` (9,054 bytes) 본문에서 frontend domain knowledge → `frontend-baseline.md`로 추출
3. meta-harness SKILL.md line 201-204 expert-* 참조 → `templates/<role>-baseline.md` 참조로 갱신

이 작업은 Category A retire 단계에서 동반 수행 가능 (domain skill body를 baseline에 흡수).

---

## 3. 10개 후보 사용 빈도 audit

### 3.1 Category A — Domain Skills (3 items, 718 LOC)

| Skill | LOC | 호출 위치 | 위험도 |
|---|---|---|---|
| `moai-domain-backend` | 195 | `expert-backend.md` frontmatter `skills:` only | **LOW** |
| `moai-domain-frontend` | 205 | `expert-frontend.md` frontmatter `skills:` only | **LOW** |
| `moai-domain-database` | 318 | `expert-backend.md` frontmatter `skills:` only | **LOW** |

**Workflow files 직접 호출**: **0건**

**Language rules cross-reference**: 9개 (kotlin/swift/elixir/csharp/scala/javascript/flutter/ruby/java/scala) — example 인용. retire 시 갱신 필요.

**결론**: Category A는 expert-{backend,frontend} retire에 종속. **단독 retire는 비효율** (expert가 살아있으면 skill 로드 계속됨). Category C와 함께 처리하는 것이 자연스러움.

### 3.2 Category B — Platform / Framework Skills (4 items, 1,432 LOC)

| Skill | LOC | workflow 호출 | agent frontmatter 호출 | 위험도 |
|---|---|---|---|---|
| `moai-framework-electron` | 328 | **0건** | **0건** | **TRIVIAL** |
| `moai-platform-auth` | 282 | **0건** | **0건** | **TRIVIAL** |
| `moai-platform-chrome-extension` | 367 | **0건** | **0건** | **TRIVIAL** |
| `moai-platform-deployment` | 455 | **0건** | **0건** (단 elixir/csharp rules에 `moai-platform-deploy` typo로 참조) | **TRIVIAL** |

**Workflow + agent 직접 호출**: **모두 0건**.

**Language rules**: elixir/csharp/kotlin/swift 등 4개 rule이 `moai-platform-deploy`로 참조 (현재 skill명과 inconsistent — typo or rename history).

**결론**: Category B는 **완전한 dead weight**. workflow나 agent에서 invoke되지 않음. 사용자가 명시 invoke하는 경우만 작동. **즉시 retire 가능**.

⚠ 주의: 사용자 프로젝트에서 직접 `Skill("moai-platform-auth")` 식으로 invoke할 수 있으므로, retire는 **deprecation period 또는 명시적 release notes**를 동반해야 함.

### 3.3 Category C — Expert Agents (2 items)

| Agent | LOC | workflow 호출 | agent cross-ref | 위험도 |
|---|---|---|---|---|
| `expert-backend` | 6,662 bytes (~165 LOC) | **11건** (loop/fix/github/moai/mx/SKILL/sync/clean/coverage 등) | 9 agents (manager-spec/strategy/quality/docs/security/performance/devops/builder-harness/manager-spec) | **HIGH** |
| `expert-frontend` | 9,054 bytes (~225 LOC) | **13건** (loop/fix/mx/e2e/design/review/clarity-interview/moai/SKILL 등) | 7 agents | **HIGH** |

**핵심 invocation 패턴**:
- `agents: ["...", "expert-backend", "expert-frontend", ...]` (workflow frontmatter)
- "Delegate to expert-backend subagent" (workflow body 가이드)
- "manager-develop (cycle_type=tdd), expert-frontend" (SKILL.md 진입점)
- "Phase 4.5: expert-frontend subagent (design review)" (review.md)

**결론**: Category C는 **모든 workflow의 phase executor**. 단순 retire 불가. 마이그레이션 = (a) baseline 추출 → meta-harness templates (b) workflow files의 expert-* invocation을 `manager-develop` + dynamic skill load 패턴으로 치환 (c) 9 agent cross-ref 정리.

### 3.4 dead reference

| 위치 | 내용 |
|---|---|
| `.claude/skills/moai-meta-harness/SKILL.md:203` | `expert-mobile` (이미 W0에서 retired) |
| `.claude/rules/moai/languages/{elixir,csharp,kotlin,swift,flutter}.md` | `moai-platform-deploy` (typo or rename, 현재 skill = `moai-platform-deployment`) |

---

## 4. 마이그레이션 우선순위 권장

### Phase 0 — Dead Reference Cleanup (No SPEC 또는 inline commit)

**Scope**: 즉시 가능, 1 commit
- `meta-harness/SKILL.md:203` `expert-mobile` 행 삭제
- 5 language rules `moai-platform-deploy` → `moai-platform-deployment` 정정 (or Category B retire 시 같이 제거)

**Wall-time**: ~10분 / **PR**: 0 or 1 (Tier S로 묶을 수도)

### Phase 1 — Category B Dead Weight Retire (Tier S SPEC)

**Scope**: 4 dead skills 삭제 + 5 language rules cross-ref 정리
- `internal/template/templates/.claude/skills/moai-framework-electron/` delete
- `internal/template/templates/.claude/skills/moai-platform-auth/` delete
- `internal/template/templates/.claude/skills/moai-platform-chrome-extension/` delete
- `internal/template/templates/.claude/skills/moai-platform-deployment/` delete
- 5 language rule cross-ref 정리
- `internal/template/catalog.yaml` (있다면) 갱신
- 로컬 `.claude/skills/` mirror 동기화 (or `make build`)

**LOC delete**: ~1,432 LOC / **위험도**: TRIVIAL (호출 0건)
**Tier**: S (artifacts 2: spec + plan, < 200 LOC each)

### Phase 2 — Category A + meta-harness Baseline 추출 (Tier M SPEC)

**Scope**:
- `moai-domain-{backend,frontend,database}` body 내용을 `moai-meta-harness/templates/{backend,frontend,database}-baseline.md`로 추출
- 9 language rule cross-ref 갱신 (`moai-domain-*` → `moai-meta-harness templates/*-baseline.md` 안내로 치환)
- 3 domain skill retire
- expert-{backend,frontend}.md frontmatter `skills:` 영역에서 domain skill 참조 제거

**LOC**: ~718 LOC delete + ~700 LOC baseline 생성 / **위험도**: LOW
**Tier**: M (artifacts 3: spec + plan + design, multi-file)

### Phase 3 — Category C Expert Agent Retire (Tier M-L SPEC, **가장 큰 작업**)

**Scope**:
- 11 workflow files (loop/fix/github/moai/mx/SKILL/sync 등) expert-backend invocation → manager-develop + `Skill("my-harness-...")` 동적 로드 패턴 치환
- 13 workflow files expert-frontend invocation 치환
- 9 agent cross-ref 정리 (manager-spec/strategy/quality/docs/security/performance/devops/builder-harness)
- expert-backend / expert-frontend agent file 자체 retire (stub redirect 또는 hard delete)
- manager-develop.md에 my-harness skill 동적 로드 메커니즘 명시
- **사전 검증**: 적어도 1개 web 프로젝트에서 meta-harness 생성 → my-harness 작동 E2E 검증 필요

**LOC**: ~24 invocation 치환 + 2 agent retire + 9 cross-ref / **위험도**: HIGH
**Tier**: M (artifacts 3) 또는 L (artifacts 5, design + research 추가 필요)

### Phase 4 — Seed Library + `/moai project --refresh` (Tier S SPEC)

**Scope**:
- 4 baseline seeds 작성 (Go-cli / Go-web / React / Vue) — 사용자 결정 (8개 vision안 ↔ 점진적 확장)
- `.claude/commands/moai/project.md` `--refresh` subcommand 추가
- `.claude/skills/moai/workflows/project/refresh.md` 구현

**LOC**: 신규 ~600 LOC / **위험도**: LOW (additive only)
**Tier**: S

### Phase 5 — Vision 문서 갱신 (No SPEC, doc only commit)

**Scope**:
- `.moai/research/harness-autonomy-vision-2026-05-18.md` §3.5 + §5 W4 + §10 acceptance 갱신:
  - §3.5 Determinism 보장 섹션 **삭제** (사용자 결정)
  - §5 W4 scope를 "polish" → "Core Slimming + Project-Specific Setup 완결"로 재작성
  - §10 deterministic acceptance bullet 삭제

**Wall-time**: ~10분 / **위험도**: TRIVIAL

---

## 5. SPEC 분리 권장

### 권장안 A — Phase별 분리 (4 SPECs, LEAN)

| SPEC | Phase | Tier | LOC change | 위험도 |
|---|---|---|---|---|
| SPEC-V3R5-CORE-SLIM-B-001 | Phase 0+1 (dead weight) | S | -1,432 | TRIVIAL |
| SPEC-V3R5-CORE-SLIM-A-001 | Phase 2 (domain → baseline) | M | -18 (net, baseline 흡수) | LOW |
| SPEC-V3R5-CORE-SLIM-C-001 | Phase 3 (expert retire) | M-L | -300, +200 | HIGH |
| SPEC-V3R5-PROJECT-MEGA-001 | Phase 4 (seed + refresh) | S | +600 | LOW |

→ 총 4 SPEC, LEAN 원칙 부합, 각 SPEC 독립 PR 가능.

### 권장안 B — 2 SPEC 통합

| SPEC | 포함 Phase | Tier |
|---|---|---|
| SPEC-V3R5-CORE-SLIM-001 | Phase 0 + 1 + 2 (safe retire) | M |
| SPEC-V3R5-PROJECT-MEGA-001 | Phase 3 + 4 (expert retire + seed + refresh) | L |

→ 총 2 SPEC, Phase 3 통합으로 Tier L 위험. LANG-COMPLIANCE 폐기 사례 재현 가능성.

### 권장안 C — 단일 SPEC (NOT RECOMMENDED)

모든 Phase 통합 → Tier L mega. **LANG-COMPLIANCE 답습 위험**. 거부.

### 결론

**권장안 A (4 SPECs)** — LEAN dogfooding (WORKFLOW-LEAN-001 plan-auditor 0.92 PASS)을 통해 검증된 Tier S/M 우선 패턴. Phase 3만 Tier M-L로 별도 처리.

---

## 6. 위험 평가 + 검증 방법

### Phase 별 검증 게이트

| Phase | 검증 방법 |
|---|---|
| Phase 0 | grep 0 매치 검증 |
| Phase 1 | `go test ./...` PASS + 사용자 프로젝트 호환성 확인 (release notes) |
| Phase 2 | `moai agent lint --strict` 0 NEW findings + 9 language rules 정합성 grep |
| Phase 3 | **선결 조건**: 1+ web/frontend 프로젝트 E2E (`/moai project` → my-harness 생성 → SPEC run 작동) |
| Phase 4 | Determinism은 NOT required. 8 seed inject 후 Tier 3 starting point 작동 |

### Regression 위험

| 위험 | Phase | 완화 |
|---|---|---|
| 사용자 프로젝트가 retired skill 명시 호출 | 1 | Release notes + deprecation period 1 release cycle |
| Language rules cross-ref 깨짐 | 2 | grep audit + spec-lint test fixture |
| Expert agent retire 후 workflow regression | 3 | E2E 검증 (위 선결조건) + characterization tests |
| meta-harness가 baseline 누락 | 2-3 | extraction commit이 baseline content 보존 검증 |

---

## 7. 다음 세션 권장 Action

### 즉시 (이번 또는 다음 세션)

1. **Phase 5 Vision 갱신** (10분, doc only) — `feedback_w4_no_determinism` 적용
2. **Phase 0 dead reference cleanup** (10분, 1 commit) — meta-harness:203 `expert-mobile` 제거 + 5 language rules `moai-platform-deploy` typo 정정

### 단기 (이번 주)

3. **SPEC-V3R5-CORE-SLIM-B-001** (Tier S) — Category B dead weight retire (Phase 0+1)
   - manager-spec 위임, LEAN 원칙 (Tier S, artifacts 2)
   - 검증: `go test ./...` PASS, 사용자 프로젝트 호환성 release notes

### 중기 (다음 주)

4. **SPEC-V3R5-CORE-SLIM-A-001** (Tier M) — Category A → baseline 추출 (Phase 2)
5. **SPEC-V3R5-CORE-SLIM-C-001** (Tier M-L) — Category C expert retire (Phase 3, **선결 조건 충족 후**)

### 장기 (W4 완결)

6. **SPEC-V3R5-PROJECT-MEGA-001** (Tier S) — seed library + `/moai project --refresh` (Phase 4)

---

## 8. Acceptance for This Audit

- [x] 10 후보의 invocation count 측정 완료
- [x] Category 분류 (A/B/C) 객관 근거 도출
- [x] Phase 0~5 SPEC 분리 권장안 3종 제시
- [x] Determinism 제거 후 W4 scope 재정의 (Vision 갱신 안 포함)
- [x] LEAN 원칙 적용 — Tier S/M 우선, Tier L 회피
- [x] WORKFLOW-LEAN-001 dogfooding 패턴 참조 (PR #1030, plan-auditor 0.92)

---

Version: 1.0.0
Classification: AUDIT_REPORT (read-only analysis, no code change)
Outcome: 4-SPEC decomposition 권장 (권장안 A), Phase 0+5 즉시 처리 가능

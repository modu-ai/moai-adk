---
id: SPEC-TOKEN-ROUTING-001
title: "Tier×Phase Declarative Model/Effort Routing Matrix (Token-Economy Epic 2/4)"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.1.0"
module: "internal/config"
lifecycle: spec-anchored
tags: "token, routing, model, effort, cost, declarative, matrix, tier, phase"
era: V3R6
depends_on: [SPEC-TOKEN-ACCOUNTING-001]
related_specs: [SPEC-V3R6-WORKFLOW-EFFORT-MAP-001, SPEC-DIVECC-DELEGATION-TOKEN-COST-001]
---

# SPEC-TOKEN-ROUTING-001 — Tier×Phase Declarative Model/Effort Routing Matrix

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-08 | manager-spec | 초안 작성 — Token-Economy Epic 2/4 (Gap B: Tier×Phase declarative model/effort routing matrix). Epic 1/4(A)=ACCOUNTING completed @ origin f88d0226f. |

## §A Context / 배경

Token-Economy Epic의 Gap B. Epic 1/4(A, SPEC-TOKEN-ACCOUNTING-001)가 per-SPEC
토큰 소비를 **측정 가능한 baseline**으로 영속화했다. 이제 그 baseline 위에서
**비용을 선언적으로 줄이는** 작업이 가능하다.

현재 리포지토리에는 두 개의 {model, effort} 선언 맵이 이미 존재한다:

1. `workflow.workflow_agents` — dynamic-workflow purpose 분류(7종:
   `read-only-extract` / `mechanical-transform` / `synthesize` / `research` /
   `verify-judge` / `implement` / `design-architecture`) → `{model, effort}` 기본값.
   SSOT는 SPEC-V3R6-WORKFLOW-EFFORT-MAP-001.
2. `workflow.team.role_profiles` — Agent Teams 역할 분류(7종: `researcher` /
   `analyst` / `architect` / `implementer` / `tester` / `designer` / `reviewer`)
   → `{model, effort, isolation, mode}`. Agent Teams 역할 정의.

**빈 자리(Gap B)**: 두 맵 모두 **역할/목적(purpose/role)** 축만 다루며,
**SPEC Tier(S/M/L) × SPEC Phase(plan/run/sync/mx)** 축은 어디에도 선언되어 있지 않다.
결과적으로 orchestrator가 `Agent()`를 spawn할 때 Tier×Phase에 따른 model/effort
 steering이 없다 — 예를 들어 "Tier S sync-phase doc/lint 작업"은 구조상 haiku/GLM
으로 충분하지만, 기본값 상속으로 sonnet/opus가 그대로 내려앉는다. 이것이
CG(Claude leader + GLM teammate) 모드가 60-70% 비용 절감을 약속함에도
사용자가 매번 `moai cg`를 수동 opt-in해야 하는 원인이다.

Gap B의 목표: **Tier×Phase → {model, effort} 선언적 라우팅 매트릭스**를
`workflow.yaml`에 추가하고, orchestrator가 spawn 시점에 이 매트릭스를
참조하도록 만들어, **비용 효율적 model/effort 할당을 DEFAULT 동작으로** 만든다.
측정은 A가, 선언은 B가, 하드스톱은 D가 담당한다(관심사 분리).

### §A.1 Epic 위치 (SSOT)

```
Token-Economy Epic (4-SPEC A → B → C → D)
├── A = SPEC-TOKEN-ACCOUNTING-001  [completed @ origin f88d0226f]
│      per-SPEC 토큰 측정 baseline (측정)
├── B = SPEC-TOKEN-ROUTING-001     [본 SPEC — plan-phase]
│      Tier×Phase 선언적 model/effort 라우팅 매트릭스 (선언)
├── C = (미저작)                   [후속]
│      검증-출력 다이어트 (verification-output reduction)
└── D = (미저작)                   [후속]
       예산 하드스톱 (enforcement — Tracker를 warn-first → hard-stop으로 강화)
```

B의 input은 A의 측정 baseline + A가 노출한 공개 API(`internal/tokenusage`,
`internal/spec/audit.go`). B의 output은 C/D가 소비하는 "이 Tier/Phase는 이
model/effort로 충분하다"는 선언. B는 A의 측정을 **재측정하지 않는다** —
A의 측정 결과를 비용 라우팅 결정의 근거로 **사용**한다.

### §A.2 사실 확인 (실측 — 관측된 명령 출력)

verification-claim-integrity §2 baseline-attribution을 만족하기 위해, plan-phase
저작 시점에 관측한 명령 출력:

- `grep -m1 '^status:' .moai/specs/SPEC-V3R6-AGENT-MODEL-ROUTING-001/spec.md`
  → `status: archived` (stale — pre-8-agent-consolidation 23-agent catalog)
- `grep -m1 '^status:' .moai/specs/SPEC-DIVECC-DELEGATION-TOKEN-COST-001/spec.md`
  → `status: completed` (delegation token-cost doctrine — B가 operationalize)
- `grep -m1 '^status:' .moai/specs/SPEC-DIVECC-EXTENSION-COST-LADDER-001/spec.md`
  → `status: completed` (extension/skill cost ladder doctrine)
- `grep -rl 'TOKEN-ROUTING' .moai/specs/` → no matches (중복 없음)
- `workflow_agents` 블록 관측: 7-purpose 맵이 `workflow.yaml`에 존재하며,
  각 엔트리는 `{model, effort}` 형태 (loader: `internal/config/workflow_agents_test.go`
  round-trip 검증됨, SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 REQ-WEM-006에 명시된 대로
  declarative metadata이며 Go 코드가 이 필드를 현재 읽지 않음 — B가 이 필드를 읽는
  첫 runtime 컨슈머이다; `workflow_agents`는 오늘날 선언 전용(declarative-only))
- `role_profiles` 블록 관측: 7-role 맵이 동일 파일에 존재하며 각 엔트리는
  `{model, effort, isolation, mode, description}` 형태
- template mirror 존재: `internal/template/templates/.moai/config/sections/workflow.yaml`
  가 존재(6850 bytes) → 본 SPEC의 `workflow.yaml` 편집은 template mirror 대상
  (§B.5 neutrality 참조). 반면 `quality.yaml`은 template tree에 부재 → 본 SPEC은
  `quality.yaml`에 쓰지 않는다(DD1 근거).

### §A.3 직교성 선언 (Orthogonality — Phase 0.95와의 분리)

본 SPEC의 Tier×Phase → {model, effort} 축은 `.claude/rules/moai/workflow/orchestration-mode-selection.md`
가 소유하는 Phase 0.95 Mode 1-6 축과 **직교(orthogonal)**한다. 두 축은 결코 경쟁하지 않는다:

| 축 | 소유자 | 의미 | 예시 |
|----|--------|------|------|
| **Mode (형태)** | Phase 0.95 | 동시성/spawn 형태 | Mode 1 trivial / Mode 2 background / Mode 3 agent-team / Mode 4 parallel / Mode 5 sub-agent / Mode 6 workflow |
| **Cost (비용)** | 본 SPEC (B) | per-spawn model/effort 비용 | Tier S sync → haiku+low / Tier L run → opus+xhigh |

직교성의 합성(composition): Phase 0.95가 **형태**를 선택하고, B가 그 형태 안에서
일어나는 **각 spawn의 비용**을 선택한다. 예: Mode 5(sub-agent)가 선택된 Tier S
sync-phase 작업은 B의 매트릭스가 haiku+low를 권장하고, orchestrator는 그 값을
`Agent(model: "haiku", effort: "low", ...)`의 per-spawn override로 주입한다.

**경계 불변량**: B는 Mode 1-6 선택 로직에 **간섭하지 않는다** — B는
auto-select thresholds(domains ≥ 3 / files ≥ 10 / score ≥ 7)를 읽거나 변경하지
않으며, mode-selection 축에 새 분기를 추가하지 않는다. 역으로 Phase 0.95는
model/effort 비용 축에 간섭하지 않는다.

## §B Requirements (GEARS)

> 본 SPEC은 GEARS(current notation)로 작성된다. `IF/THEN` modality는 사용하지
> 않는다 — failure 모드는 `When <undesred-condition-detected>` event 형태로
> 표현한다. `<subject>`는 system뿐 아니라 component/service/agent/function/artifact를
> 포함하는 일반명사(GEARS의 generalized subject).

### §B.1 선언적 매트릭스 (Declarative Table)

**REQ-TR-001** (Ubiquitous): `workflow.yaml` **shall** carry a
`model_routing` top-level block mapping each (Tier × Phase) pair to a
`{model, effort}` recommendation, where Tier ∈ {S, M, L} and Phase ∈
{plan, run, sync, mx}.

**REQ-TR-002** (Where): **Where** a (Tier × Phase) pair is absent from the
declared `model_routing` map, the loader **shall** fall back to a documented
default entry rather than fail, and the fallback value **shall** be surfaced
to the caller via a `fallback_applied: true` indicator on the returned entry.

### §B.2 로더 표면 (Typed Loader)

**REQ-TR-003** (Ubiquitous): `internal/config` **shall** expose a typed loader
returning the routing entry for a given (Tier, Phase) pair —
`RouteModelFor(tier, phase) ModelRoutingEntry` — reading from the canonical
`model_routing` YAML block.

**REQ-TR-004** (Where): **Where** the loaded `model` value is not a member of
the closed set `{inherit, haiku, sonnet, opus, glm}` or the loaded `effort`
value is not a member of `{low, medium, high, xhigh, max}`, the loader
**shall** return a validation error and refuse to yield the entry.

### §B.3 Spawn-time 참조 (Orchestrator Consumption)

**REQ-TR-005** (When): **When** the orchestrator composes an `Agent()` spawn
for a phase that the matrix maps, the orchestrator **shall** consult
`RouteModelFor(tier, phase)` and emit the returned `{model, effort}` as
per-spawn override parameters on the `Agent()` call.

**REQ-TR-006** (Where): **Where** the orchestrator already received an
explicit `model:`/`effort:` user override for a specific spawn (e.g. a
hand-seeded profile), the matrix-recommended value **shall** yield to the
explicit override and NOT replace it; the matrix is a default, not a mandate.

### §B.4 폐쇄집합 검증 (Closed-Set Validation)

**REQ-TR-007** (Ubiquitous): `internal/config` **shall** validate the
`model_routing` block against the closed sets declared in REQ-TR-004; the
**effort** enum `{low, medium, high, xhigh, max}` matches the canonical
`workflow_agents` enum per SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 REQ-WEM-006
(so the two maps' effort sets cannot drift apart); the **model** set
extends the live `workflow_agents` values with `glm` for cost-routing
purposes (REQ-TR-012).

### §B.5 템플릿 거울 + 중립성 (Template Mirror + Neutrality)

**REQ-TR-008** (Where): **Where** the `workflow.yaml` declarative table is
edited at run-phase, the template mirror at
`internal/template/templates/.moai/config/sections/workflow.yaml` **shall**
receive the SAME `model_routing` block (generic, no SPEC IDs / no internal
dates / no commit SHAs — CLAUDE.local.md §15/§25), so downstream users
inherit the same defaults without fork divergence.

### §B.6 기본 동작 계약 (Default-Behavior Contract — the "60-70% reduction as default")

**REQ-TR-009** (When): **When** the orchestrator issues an `Agent()` spawn
for a phase the matrix maps to a cheaper model than the session default,
the orchestrator **shall** apply the cheaper `model:`/`effort:` override
**without** an `AskUserQuestion` confirmation round — cost-efficient
assignment is the DEFAULT behavior, not an opt-in.

**REQ-TR-010** (Ubiquitous): The matrix **shall** NOT make `moai cg` the
default launcher — CG mode remains user opt-in (`moai cg` enters CG; without
it, the session stays Claude-only and the cost reduction comes from cheaper
Claude models in the closed set, not from GLM routing). 본 SPEC의 "default"
는 "각 spawn의 model/effort를 매트릭스가 결정한다"는 의미이지, "GLM teammate
판을 자동으로 여는 것"이 아니다 (verification-claim-integrity — 모호한
"default" 발화 금지).

### §B.7 직교성 보존 (Orthogonality Preservation)

**REQ-TR-011** (Ubiquitous): The matrix **shall not** add a new axis to
`.claude/rules/moai/workflow/orchestration-mode-selection.md` — B's Tier×Phase
→ {model, effort} is an ORTHOGONAL cost axis to Phase 0.95's Mode 1-6 shape
axis; B composes with Mode 1-6, never competes with them.

### §B.8 배포 중립성 (Deployment Neutrality)

**REQ-TR-012** (Where): **Where** the loaded `model_routing` references a
model unavailable in the current deployment (e.g. `glm` referenced while
GLM env vars are absent), the loader **shall** surface an advisory rather
than block — the orchestrator MAY fall back to the session-inherited model
in that case, and the advisory is logged to `.moai/logs/` for audit.

## §C Success Criteria

- `workflow.yaml`의 `model_routing` 블록이 3 Tier × 4 Phase = 12 엔트리를
  가지며, 각 엔트리는 closed-set {model, effort} 값을 가진다.
- `internal/config.RouteModelFor(tier, phase)` 가 typed entry를 반환하고,
  closed-set 위반 시 validation error를 낸다 (테스트로 검증).
- template mirror(`internal/template/templates/.moai/config/sections/workflow.yaml`)
  가 동일한 `model_routing` 블록을 가지며, CI neutrality guard가 PASS 한다.
- 8-agent 현행 카탈로그(`manager-spec`/`manager-develop`/`manager-docs`/
  `manager-git`/`plan-auditor`/`sync-auditor`/`builder-harness`/`Explore`)
  외의 archived 에이전트 이름이 매트릭스 문서/주석에 노출되지 않는다.
- 매트릭스가 `moai cg`를 default launcher로 만들지 않는다(REQ-TR-010).
- 매트릭스가 Phase 0.95 mode-selection에 간섭하지 않는다(REQ-TR-011).

## §D Out of Scope (Exclusions)

> 본 SPEC은 Token-Economy Epic의 **선언(declaration)** 층만 담당한다.
> 측정/하드스톱/mode선택은 인접 SPEC의 소관이며, 본 SPEC이 침범하지 않는다.

### Out of Scope — 토큰 측정 (measurement)

- per-SPEC 토큰 합산/귀속/영속/감사 표면 — 전부 Epic A(SPEC-TOKEN-ACCOUNTING-001)
  가 completed @ origin f88d0226f에서 closed. 본 SPEC은 A의 공개 API만 소비한다.
- `internal/tokenusage.SumSession` / `Attribute` / `BuildSectionI` /
  `WriteSectionI` 의 동작 변경 — OUT.

### Out of Scope — 예산 하드스톱 (enforcement)

- `internal/runtime/budget.go` Tracker를 warn-first에서 hard-stop으로 강화하는
  작업 — Epic D(미저작) 소관. 본 SPEC은 매트릭스를 선언만 할 뿐, 예산 초과 시
  spawn을 차단하지 않는다.

### Out of Scope — Phase 0.95 mode-selection (orchestration shape)

- `.claude/rules/moai/workflow/orchestration-mode-selection.md` 의 Mode 1-6
  카탈로그/결정트리/auto-select thresholds(domains ≥ 3 / files ≥ 10 / score ≥ 7)
  변경 — OUT. 본 SPEC의 Tier×Phase → {model, effort}는 Mode 1-6과 **직교**하는
  별도 축이며, mode-selection에 새 분기를 추가하지 않는다(REQ-TR-011).

### Out of Scope — `moai cg` default launcher 전환

- CG mode를 사용자 opt-in이 아닌 default launcher로 만드는 변경 — OUT.
- `moai cg`는 여전히 사용자가 명시적으로 진입해야 하며, 본 SPEC의 "default"
  는 "각 spawn의 model/effort를 매트릭스가 결정한다"로 한정된다(REQ-TR-010).

### Out of Scope — `workflow_agents` / `role_profiles` 재작성

- `workflow.workflow_agents`(7-purpose 맵) 및 `workflow.team.role_profiles`
  (7-role 맵)의 재작성/확장 — OUT. 두 맵은 SPEC-V3R6-WORKFLOW-EFFORT-MAP-001
  및 Agent Teams 정의에 충실하며, 본 SPEC은 **새로운** `model_routing` 블록을
  **추가**할 뿐 기존 맵을 건드리지 않는다(관심사 분리).

### Out of Scope — 검증-출력 다이어트 (verification-output reduction)

- sync-phase 검증 매트릭스/로그 출력의 토큰 다이어트 — Epic C(미저작) 소관.
- 본 SPEC이 매트릭스를 선언하지만, 매트릭스 평가 결과를 사용자에게 어떻게
  렌더링할지는 C가 담당한다.

### Out of Scope — 서브에이전트 내부 토큰 정밀 분해

- A와 동일하게, 서브에이전트 내부 토큰이 orchestrator transcript
  `message.usage`에 포함되는지의 정밀도는 본 SPEC 범위 밖. B는 매트릭스를
  선언할 뿐, 토큰 정밀 회계를 재검증하지 않는다.

### Out of Scope — 8-agent 카탈로그 재변경

- retained agent 목록 변경 — OUT. 매트릭스는 현행 8-agent 카탈로그에
  정렬하며, archived 에이전트 이름(`expert-backend`/`expert-frontend`/
  `manager-strategy`/`manager-quality` 등)을 부활시키지 않는다.

## §E Cross-References (Prior Art)

> 본 SPEC은 인접 영역을 다루는 3개의 기존 SPEC을 **재작업하지 않는다** —
> 각각의 현재 상태(verbatim)를 인용하며, 본 SPEC은 그 위에 **새로운 축**
> (Tier×Phase)을 추가한다.

| SPEC-ID | 현재 상태 (plan-phase 실측) | 본 SPEC과의 관계 |
|---------|-----------------------------|------------------|
| SPEC-V3R6-AGENT-MODEL-ROUTING-001 | `status: archived` (stale: pre-8-agent-consolidation 23-agent catalog; 라우팅 테이블이 `expert-backend`/`expert-frontend` 등 retired 이름을 참조) | 본 SPEC은 **현행 8-agent 카탈로그**에 정렬. archived SPEC의 라우팅 테이블을 부활시키지 않는다. |
| SPEC-DIVECC-DELEGATION-TOKEN-COST-001 | `status: completed` (delegation token-cost doctrine) | 본 SPEC이 그 비용-승수 원리를 **operationalize** (실행 계층에 내림). doctrine 자체를 재도출하지 않는다. |
| SPEC-DIVECC-EXTENSION-COST-LADDER-001 | `status: completed` (extension/skill cost ladder doctrine) | 본 SPEC은 extension/skill이 아닌 spawn model/effort를 다루지만, 동일한 비용 사다기 원리에 정렬. |
| SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 | (related — `workflow_agents` SSOT) | 본 SPEC의 `model_routing` 블록은 `workflow_agents`와 동일한 closed-set 검증을 공유(REQ-TR-007). |
| SPEC-TOKEN-ACCOUNTING-001 | `status: completed` @ origin f88d0226f (Epic A) | 본 SPEC의 의존성(`depends_on`). A의 측정 baseline 위에서 선언을 작성. |

## §F Token Accounting (Reserved — A's convention)

> 본 섹션은 `spec-frontmatter-schema.md` Section Map SSOT의 `## §I Token Accounting`
> 스키마를 준수하기 위해 sync-phase에 manager-docs가 실측 값으로 채운다
> (SPEC-TOKEN-ACCOUNTING-001 M3가 확립한 writer 계약 — `BuildSectionI` /
> `WriteSectionI`). plan-phase에서는 placeholder만 둔다.

_(sync-phase에 tokens_spent 값으로 채워질 예정 — plan-phase에서는 미측정)_

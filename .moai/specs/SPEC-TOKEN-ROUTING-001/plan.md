# Implementation Plan — SPEC-TOKEN-ROUTING-001

> plan.md는 파생 실행 계획이다. WHAT/WHY의 SSOT는 spec.md. 본 문서는 HOW의
> 골격(마일스톤/제약/리스크)이며 함수명·시그니처 등 세부 설계는 run-phase 소관.
> 본 SPEC은 **plan-phase 전용** 산출물이며, Implementation Kickoff Approval
> (human gate) 이전에 run-phase에 진입하지 않는다.

## §A Context

Token-Economy Epic 2/4. A가 측정했다면, B는 **선언**한다: Tier×Phase →
{model, effort} 매트릭스를 `workflow.yaml`에 추가하고, orchestrator가 spawn
시점에 참조하도록 만든다. 결과적으로 "Tier S sync-phase doc/lint 작업은
haiku+low로 충분하다" 같은 비용 효율적 할당이 DEFAULT 동작이 된다.

전면 신규 구축이 아닌 **확장(EXTEND)** 전략: 두 개의 기존 {model, effort}
맵(`workflow_agents` / `role_profiles`)이 이미 `workflow.yaml`에 존재하며,
본 SPEC은 세 번째 맵(`model_routing`)을 **같은 파일, 같은 로더 패턴,
같은 closed-set 검증**으로 추가한다.

Tier: **M** (표준). 근거는 §D. 8-agent 현행 카탈로그 정렬 + template mirror +
loader 신규 코드를 동시에 다루나, 다중 도메인에 걸치지는 않는다.

## §B Known Design Challenges (spec.md §A 대응 해소 — 5 Design Decisions)

### B-location (DD1) — 선언적 매트릭스의 위치

세 후보를 평가한다:

| 옵션 | 위치 | 근거 | 판정 |
|------|------|------|------|
| (i) `quality.yaml` | harness level 섹션에 추가 | "비용 = 품질" 연상 | **REJECT** — `quality.yaml`은 template tree에 부재(§A.2 실측). 여기에 두면 template mirror 불가 → fork divergence. |
| (ii) `role_profiles` 확장 | tier/phase 축을 role 맵에 용여 | 한 맵에 모든 축을 결합 | **REJECT** — 역할(role)과 Tier×Phase는 직교 축. 한 맵에 용여하면 읽기/유지보수 비용이 폭발. |
| **(iii→채택) `workflow.yaml` 최상위 `model_routing` 블록** | `workflow_agents`/`role_profiles`와 형제 | 같은 파일, 같은 로더 패턴, 같은 closed-set | **채택** |

**채택: (iii)**. 이유:
- `workflow.yaml`은 이미 두 개의 {model, effort} 맵을 SSOT로 소유하며,
  같은 파일의 같은 로더 패턴을 따르는 세 번째 맵은 **단일 진실 공급원 원칙**
  (single-source-of-truth)을 만족한다.
- `workflow.yaml`은 template tree에 **존재**(`internal/template/templates/.moai/config/sections/workflow.yaml`
  6850 bytes, §A.2 실측) → template mirror 가능 → downstream 사용자에게
  동일한 기본값이 배포된다(REQ-TR-008).
- 새 최상위 블록(`model_routing`)은 Tier×Phase 매트릭스를 자체 주소 가능한
  SSOT로 유지한다 — `RouteModelFor(tier, phase)` 가 이 블록만 읽고,
  7-purpose 맵이나 7-role 맵과 엉키지 않는다.

### B-mechanism (DD2) — Orchestrator spawn-time 참조 메커니즘

세 옵션을 평가한다:

| 옵션 | 방식 | 검증 가능성 | 판정 |
|------|------|-------------|------|
| (a) agent-body reference | `manager-spec`/`manager-develop`/`manager-docs` 본문이 매트릭스를 path로 참조, per-spawn `model:`/`effort:`를 산출 | 약(prose-dependent) | 단독 사용 REJECT |
| (b) runtime loader | `internal/config.RouteModelFor(tier, phase)` 가 typed entry 반환; hook/spawn 코드가 읽음 | 강(typed + tested) | **채택(주)** |
| (c) pure declarative + docs | orchestrator가 YAML을 prose로 읽음 | 매약(검증 없음) | REJECT |

**채택: (b) + (a) hybrid**. 이유:
- Loader가 typed, tested 표면 — closed-set 검증(REQ-TR-007)과 fallback(REQ-TR-002)을
  코드로 강제한다. prose-only는 verification-claim-integrity 위험(매트릭스가
  실제 동작과 drift해도 detection 불가).
- Agent body reference(a)는 **보조**: spawn 프롬프트를 조립하는 코드 경로가
  loader를 호출하도록 짧은 call-site 포인터를 갖는다. agent 본문을 전면
  rewrite하지 않는다.
- Pure declarative(c)는 매트릭스 drift를 기계적으로 잡을 수 없어 REJECT.
- **scope 제한**: run-phase는 loader + call-site reference를 추가하되,
  agent 본문의 behavior를 변경하지 않는다(agent 본문에 "매트릭스를 읽어라"
  지시문 추가 OUT — 그것은 agent body rewrite이며 Tier를 넘음).

### B-orthogonality (DD3) — Phase 0.95와의 관계

본 SPEC의 Tier×Phase → {model, effort}는 `.claude/rules/moai/workflow/orchestration-mode-selection.md`
가 소유하는 Phase 0.95 Mode 1-6과 **직교**한다:

- **Mode 축**(Phase 0.95): 동시성/spawn **형태** (trivial/background/agent-team/parallel/sub-agent/workflow).
  auto-select thresholds: domains ≥ 3 / files ≥ 10 / score ≥ 7.
- **Cost 축**(본 SPEC): per-spawn **비용** ({model, effort}).
  매트릭스: Tier × Phase.

**합성(composition)**: Phase 0.95가 형태를, B가 그 형태 안의 각 spawn 비용을
고른다. 예: Mode 5(sub-agent)로 Tier S sync-phase 작업이 선택되면, B의 매트릭스가
haiku+low를 권장하고 orchestrator는 `Agent(model: "haiku", effort: "low")`로
spawn. 두 축은 결합할 수 있지만, 결코 경쟁하지 않는다.

**경계 불변량(REQ-TR-011)**: B는 mode-selection thresholds를 읽지/변경하지
않으며, Mode 1-6 카탈로그에 새 분기를 추가하지 않는다. 역으로 Phase 0.95는
model/effort 비용 축에 간섭하지 않는다. 이 직교성을 design.md에 명시하여
scope overlap을 방지한다.

### B-default (DD4) — "CG 60-70% reduction as default behavior" 정확한 범위

"Default"의 의미를 세 후보로 좁혀 정의한다:

| 해석 | 의미 | 판정 |
|------|------|------|
| (a) auto-route low-stakes phases | Tier S sync-phase doc/lint → haiku/GLM **자동** (AskUserQuestion 없음) | **채택** |
| (b) make `moai cg` default launcher | CG mode를 사용자 opt-in이 아닌 기본 진입으로 | **REJECT** |
| (c) both | (a) + (b) 동시 | **REJECT** (b 때문에) |

**채택: (a) only** (REQ-TR-009, REQ-TR-010). 이유:
- (a)는 매트릭스가 결정하면 orchestrator가 묻지 않고 적용하는 **기본 동작**이다.
  이것이 "60-70% 절감을 default로"의 구체적 메커니즘이다.
- (b)는 **overclaim 위험**: `moai cg`는 tmux/GLM env/teammateMode를 전제로
  하는 사용자 opt-in 진입점이다. 이것을 default로 만들려면 사용자가 tmux
  세션에 있지 않을 때 강제로 진입하게 되어, 자율성 침식(askuser-protocol.md
  §3 cold-start 공개 의무 참조) + 환경 의존성 폭발. 본 SPEC은 (b)를 명시적으로
  배제한다(REQ-TR-010).
- (a)만으로도 "Tier S sync → haiku+low 자동 적용"이 CG mode 없이 Claude-only
  환경에서 비용 절감을 가져온다(haiku는 60-70% 저렴). GLM routing은
  `moai cg` 진입 시에만 매트릭스의 `glm` 값이 유효하다(REQ-TR-012 unavailable
  model advisory).
- **verification-claim-integrity 준수**: 모호한 "default" 발언을 (a)의 정확한
  메커니즘으로 대체한다. AC는 이 계약을 관측 가능하게 검증(§D.1 Scenario 3).

### B-tier (DD5) — 본 SPEC 자체의 Tier 분류

**추정: Tier M**. 근거:

- run-phase가 touch할 파일:
  1. `internal/config/types.go` — `ModelRoutingConfig` 타입 + 신규 `ModelRoutingEntry`
     구조체 추가 (**기존 `WorkflowAgentEntry` 재사용 불가** — REQ-TR-002가 `fallback_applied`
     필드를 mandate하는데 `WorkflowAgentEntry`(`internal/config/types.go` L362-L365)는
     `{Model, Effort}`만 보유. 제안 shape: `{Model string, Effort string, FallbackApplied bool}`.
     재사용 후보는 run-phase probe 전후로 "open question"으로 남아 있었으나, D2 audit
     에서 구조적 불가능함이 확인되어 plan-phase에서 확정 차단 — run-phase 결정 아님)
  2. `internal/config/model_routing.go` (신규) — typed loader `RouteModelFor(tier, phase)` + validation + fallback
  3. `.moai/config/sections/workflow.yaml` — `model_routing` 블록 추가(12 엔트리: 3 Tier × 4 Phase)
  4. `internal/template/templates/.moai/config/sections/workflow.yaml` — template mirror(neutrality)
  5. `internal/config/model_routing_test.go` (신규) — round-trip + closed-set + fallback 테스트
  6. (보조) orchestrator spawn call-site 경로 — loader 호출 포인터(최소 patch)
- 6 files = Tier M 영역(3-5 files = S, 6-9 = M, 10+ = L).
- loader 코드는 신규 but bounded(~150-250 LOC + 테스트). 복잡한 알고리즘 없음 —
  YAML → map read + closed-set 검증 + fallback 상수.
- 단일 도메인(`internal/config` + `template/templates/.moai/`) + template mirror. 다중 도메인 아님.
- Tier L이 아닌 이유: runtime/agent 본문을 다수 변경하지 않음. Tier S가 아닌
  이유: pure declarative + docs가 아닌 typed loader 코드가 신규 작성됨.

workflow-specialist / orchestrator가 Kickoff에서 최종 확인.

### B-noleak — 템플릿 중립성 (CLAUDE.local.md §15/§25)

- `workflow.yaml` 은 template tree에 존재(§A.2 실측) → `model_routing` 블록은
  template mirror 대상. mirror 내용은 generic(SPEC ID / 내부 날짜 / commit SHA
  전부 제외)해야 한다(REQ-TR-008).
- `model_routing` 블록의 주석에 SPEC-ID(`SPEC-TOKEN-ROUTING-001`)를 적으면
  안 된다 — `internal_content_leak_test.go` + `template-neutrality-check.yaml`
  CI guard가 fail. 주석은 generic("Tier×Phase → {model, effort} routing defaults")
  로만.
- 본 SPEC이 touch하는 `internal/config/**` Go 코드는 template tree 밖
  (runtime/dev 전용) → A와 동일하게 template neutrality guard 대상 밖.

### B-existing — 중복 금지 (재구축이 아닌 확장)

- `workflow_agents`(7-purpose) 맵 — SPEC-V3R6-WORKFLOW-EFFORT-MAP-001 SSOT.
  본 SPEC이 재작성하지 않음.
- `role_profiles`(7-role) 맵 — Agent Teams 정의 SSOT. 본 SPEC이 재작성하지 않음.
- `internal/config/token_budget_guard.go` — 항상-로드 75K 정적 회귀 가드
  (SPEC-TOKEN-EFFICIENCY-001). 본 SPEC과 별개 — 런타임 per-spawn 라우팅이지
  always-loaded 컨텍스트 표면 예산이 아님.
- `internal/tokenusage/**` — A의 측정 패키지. 본 SPEC이 소비(`RouteModelFor`
  가 A의 측정을 읽지는 않지만, 설계 결정이 A의 baseline 위에 있음을 문서화).
- `internal/runtime/budget.go` Tracker — warn-first 정책. 본 SPEC이 hard-stop으로
  강화하지 않음(D의 소관, out-of-scope).

### B-frontmatter — 12-field canonical schema

본 SPEC의 `spec.md` frontmatter는 12개 canonical 필드 + `era: V3R6` + 2개
optional 필드(`depends_on`/`related_specs`)로 구성. `status: draft` 로 시작,
`draft → in-progress` 전이는 manager-develop 소관, `in-progress → implemented
→ completed` 전이는 manager-docs 소관. 본 SPEC(manager-spec)은 초기 `draft`
만 emit. snake_case alias 거부(`created_at`/`labels` 등 사용 금지).

### B-scope-discipline — 병렬 세션 미커밋 변경 무접촉

- 현재 working tree는 병렬 세션 미커밋 변경을 포함(`system.yaml` / `statusline/*`
  / `version.go` / untracked hooks/specs). 본 SPEC의 산출물은 오직
  `.moai/specs/SPEC-TOKEN-ROUTING-001/` pathspec 내에 생성.
- 커밋 시 `git add .moai/specs/SPEC-TOKEN-ROUTING-001/ && git commit`
  (specific-path only) — `git add -A` 절대 금지(feedback_shared_checkout_concurrent_commit_race).
- run-phase 진입 후에도 동일 규율: `workflow.yaml` 편집 커밋은 본 SPEC 진행과
  관련 없는 병렬 세션 파일을 포함하지 않는다.

## §C Pre-flight

- [x] Epic A(SPEC-TOKEN-ACCOUNTING-001) `status: completed` 실측(origin f88d0226f)
- [x] 3 prior-art SPEC 상태 실측(archived/completed/completed)
- [x] `workflow.yaml` template mirror 존재 실측(6850 bytes)
- [x] `quality.yaml` template tree 부재 실측 → DD1 근거 확정
- [x] `workflow_agents_test.go` round-trip 패턴 확인 → loader 구현 패턴 확정
- [x] `workflow_accessors.go` backward-compat accessor 패턴 확인
- [x] 중복 SPEC 부재 확인(`grep TOKEN-ROUTING .moai/specs/` → no matches)
- [x] 8-agent 현행 카탈로그 확인(CLAUDE.md §4)
- [ ] plan-auditor Phase 0.5 PASS (대기)
- [ ] Implementation Kickoff Approval (대기)

## §D Constraints & Tier 근거

**Tier: M (표준)**

근거:
- 산출물 범위: typed loader 코드(~150-250 LOC) + 선언적 YAML(12엔트리) +
  template mirror + round-trip 테스트 + 짧은 call-site 포인터. 다중 도메인
  산술이 아닌 단일 도메인(`internal/config` + `template/templates/.moai/`).
- 리스크: 중간. loader validation drift(`workflow_agents` closed-set과 어긋나면
  REQ-TR-007 위반)를 테스트로 잡는다. template neutrality 위반을 CI guard로 잡는다.
- 직교성(REQ-TR-011): Phase 0.95에 간섭하지 않는다고 AC로 검증.
- default-behavior 계약(REQ-TR-009/010): "default"의 정확한 범위를 AC로 검증.
- Tier L이 아닌 이유: runtime/agent 본문 다수 변경 없음.
- Tier S가 아닌 이유: pure declarative가 아닌 typed loader 코드 신규 작성.

**제약**:
- 병렬 세션 공유 체크아웃 — pathspec-only commit.
- template neutrality — `model_routing` 주석에 SPEC-ID 미노출.
- closed-set 일치 — `workflow_agents`와 동일 집합 사용(REQ-TR-007).
- era: V3R6 — 3-phase close 계약(plan→run→sync; `completed`는 sync 커밋에 ride).

## §E Self-Verification (frontmatter schema)

- [x] `id: SPEC-TOKEN-ROUTING-001` — canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` PASS
- [x] 12 canonical 필드 모두 존재(id/title/version/status/created/updated/author/priority/phase/module/lifecycle/tags)
- [x] `status: draft` — 초기치(manager-spec 소관)
- [x] `created`/`updated` — `created_at`/`updated_at` alias 사용 안 함
- [x] `tags` — comma-separated quoted string, `labels` YAML array 아님
- [x] `version: "0.1.0"` — quoted string, unquoted float 아님
- [x] `priority: P1` — P-prefixed uppercase
- [x] `lifecycle: spec-anchored` — enum in {spec-anchored|spec-lite|exploratory}
- [x] `module: "internal/config"` — non-empty
- [x] `era: V3R6` + `depends_on`/`related_specs` optional 필드 유효
- [x] Out of Scope 섹션 `### Out of Scope — <topic>` H3 + `-` bullet 충족(OutOfScopeRule)
- [x] GEARS notation — `IF/THEN` 미사용, `When <event>` 형태 사용
- [x] 12 REQ + 12 AC 매핑(REQ-TR-001..012 ↔ AC-TR-001..012)

## §F Milestones (우선순위 기반, 시간 추정 없음)

> Milestone은 priority-based이며 time estimate를 포함하지 않는다
> (agent-common-protocol.md §Time Estimation). run-phase 진입 순서.

### M1 — 선언적 매트릭스 + template mirror (priority: High)

- `workflow.yaml` 최상위 `model_routing` 블록 추가 — 3 Tier × 4 Phase = 12 엔트리
- template mirror(`internal/template/templates/.moai/config/sections/workflow.yaml`) 동일 블록 추가(neutrality)
- AC-TR-001, AC-TR-002, AC-TR-008, AC-TR-011

### M2 — typed loader + validation (priority: High)

- `internal/config/types.go` — `ModelRoutingConfig` 타입 + 신규 `ModelRoutingEntry`
  (REQ-TR-002 `fallback_applied` 필드 mandate로 `WorkflowAgentEntry` 재사용 불가 —
  제안 shape `{Model string, Effort string, FallbackApplied bool}`; 상세 근거는 §B DD5 file-map)
- `internal/config/model_routing.go` (신규) — `RouteModelFor(tier, phase)` + closed-set 검증 + fallback
- AC-TR-003, AC-TR-004, AC-TR-007

### M3 — orchestrator call-site 참조 + default-behavior 계약 (priority: High)

- orchestrator spawn call-site이 `RouteModelFor` 를 호출해 per-spawn `model:`/`effort:` override 주입
- explicit user override 시 matrix 양보(REQ-TR-006)
- AskUserQuestion 없이 자동 적용(REQ-TR-009)
- `moai cg` default launcher 전환 배제(REQ-TR-010)
- AC-TR-005, AC-TR-006, AC-TR-009, AC-TR-010

### M4 — 회귀 테스트 + deployment neutrality + sync 준비 (priority: Medium)

- unavailable model advisory(REQ-TR-012) — `glm` referenced but GLM env 부재 시 advisory + session-inherited fallback
- `internal/config/model_routing_test.go` — round-trip / closed-set / fallback / shared-UUID advisory 4 subcase
- template neutrality CI guard PASS
- 8-agent catalog 정렬 grep(archived 이름 0 노출)
- AC-TR-012, §D.3 Quality Gate, §D.4 DoD

## §G Anti-Patterns (회피)

- **AP-1 `moai cg` default launcher overclaim**: "60-70% 절감 default"를
  GLM teammate 판 자동 오픈으로 해석 → REQ-TR-010 위반. default는 (a)로 한정.
- **AP-2 mode-selection 간섭**: 매트릭스가 Mode 1-6 카탈로그나 auto-select
  thresholds를 읽거나 변경 → REQ-TR-011 위반. 직교성 위반.
- **AP-3 prose-only 매트릭스**: loader 없이 orchestrator가 YAML을 prose로 읽음
  → 검증 불가, drift 탐지 불가. DD2 REJECT 사유.
- **AP-4 template neutrality 위반**: `model_routing` 주석에 SPEC-ID/내부날짜/SHA
  노출 → CI guard fail + downstream leak.
- **AP-5 closed-set drift**: `workflow_agents`의 {model, effort} 집합과
  `model_routing`의 집합이 어긋남 → REQ-TR-007 위반.
- **AP-6 archived agent 이름 부활**: `expert-backend`/`expert-frontend`/
  `manager-strategy` 등을 매트릭스에서 참조 → archived-agent-rejection.md 위반.
- **AP-7 pathspec 위반**: `git add -A` 사용 → 병렬 세션 미커밋 변경 흡수
  (feedback_shared_checkout_concurrent_commit_race).
- **AP-8 사용자 override 무시**: 매트릭스가 explicit user override를 밀어냄
  → REQ-TR-006 위반. matrix는 default, mandate 아님.

## §H Cross-References

- **Epic A (completed)**: `.moai/specs/SPEC-TOKEN-ACCOUNTING-001/spec.md` —
  측정 baseline. 본 SPEC의 `depends_on`.
- **Phase 0.95**: `.claude/rules/moai/workflow/orchestration-mode-selection.md`
  §A Mode Catalog / §B Decision Tree — 본 SPEC과 직교.
- **workflow_agents SSOT**: `.moai/specs/SPEC-V3R6-WORKFLOW-EFFORT-MAP-001/`
  (related_specs) — closed-set 원천.
- **delegation cost doctrine**: `.moai/specs/SPEC-DIVECC-DELEGATION-TOKEN-COST-001/`
  (completed) — 본 SPEC이 operationalize.
- **frontmatter schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- **neutrality guards**: `internal/template/internal_content_leak_test.go`,
  `.github/workflows/template-neutrality-check.yaml`
- **archived agent rejection**: `.claude/rules/moai/workflow/archived-agent-rejection.md`
- **loader 패턴 참조**: `internal/config/workflow_accessors.go`,
  `internal/config/workflow_agents_test.go` (round-trip 패턴)

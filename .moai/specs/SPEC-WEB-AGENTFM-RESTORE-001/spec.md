---
id: SPEC-WEB-AGENTFM-RESTORE-001
title: "moai web 서브 에이전트 설정 표면 복원 — agentfm 오버라이드 · llm.profile · 저장 경로 · UI 노출"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/web, internal/settings/agentfm, internal/config, internal/template, internal/template/templates/.moai/config/sections"
lifecycle: spec-anchored
tags: "web-console, agentfm, agent-overrides, llm-profile, restore, i18n, t1246-reversal"
tier: M
related_specs: [SPEC-AGENT-MODEL-INHERIT-001, SPEC-AGENT-TIER-001, SPEC-WEB-CONSOLE-011, SPEC-MODEL-PROFILE-MATRIX-002]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-02 | manager-spec | 최초 작성 (카드 t1411, Class C, 운영자 지시 2026-10-02). 삭제 커밋 `384eb3460`(card t1246, SPEC-AGENT-MODEL-INHERIT-001 M2/REQ-AMI-011)의 역방향 복원 SPEC. 블루프린트: `.moai/reports/t1411/removed-commit-384eb3460.diff` (7,922행). |

## §A 배경과 복원 대상

2026-09-27 커밋 `384eb3460`은 SPEC-AGENT-MODEL-INHERIT-001 REQ-AMI-011(서브에이전트가 메인 세션의 model/effort를 상속한다)에 따라 `moai web` 콘솔에서 에이전트 설정 탭과 저장 API를 삭제했다. 운영자는 2026-10-02 카드 t1411로 이 설정 표면(스키마·저장 경로·UI 노출)의 복원을 지시했다. 상속은 기본 동작으로 유지하되, 에이전트별 오버라이드를 다시 저장·표시하는 콘솔 표면을 되살린다.

블루프린트는 삭제 커밋의 전체 diff(`.moai/reports/t1411/removed-commit-384eb3460.diff`)이며, "revert가 아니라 re-port"가 원칙이다(§C 결정 4). 삭제 이후 5일간의 드리프트 — SPEC-AGENT-TIER-001 착지(workflow 패널 안 agent-tier 서브섹션, `agenttierpanel.go` / `jevkey.go` 배치 패턴), 모델 매트릭스 갱신(t1368), i18n 거버넌스 강화 — 위에서 현재 트리의 정형 형태로 재작성한다.

측정으로 확인된 현재 트리 상태:

- 콘솔 쪽 부재 락: `internal/web/agent_settings_removed_test.go`(렌더 마커 5종 부재 + 위조 POST 무효), `internal/web/agent_settings_test.go`(role_profiles·workflow_agents 무시 — 복원과 무관, 유지).
- 설정·템플릿 쪽 부재: `config.LLMConfig`에 `Profile`/`AgentOverrides` 필드 없음, `internal/template/profile_matrix.go` 소멸(삭제 커밋 `3fa8bd2ab`, AMI-001 M5), `internal/settings/agentfm` 패키지 소멸, 템플릿 `llm.yaml`에서 키 삭제 + 은퇴 주석 블록. 같은 커밋 `3fa8bd2ab`가 `internal/template/glm_effort_overlay.go`의 `ResolveGLMReasoningForModel`도 소멸시켰다(측정: HEAD에서 정의 0건, 본문은 제거 주석만 생존 — glm_effort_overlay.go:195) — **재포트 대상**(plan §A.1/§D.4, M1).
- 생존 기계: `internal/harness/v4manifest` closed-set 상수(`ModelInherit`/`ModelHaiku`/`EffortLow`..`EffortMax` — 실측 `schema.go:44-63`), `template.IsGLMBackend`, 키 스크립 기계 `template.ShippedRetiredModelKeys`(템플릿이 여전히 수재하는 키는 strip 제외 — 자동 해장 seam), i18n `hint.effort.go_unbound`/`hint.effort.haiku_na`(en·ko 생존 실측).
- 현재 에이전트 로스터(`.claude/agents/moai/`, 12종): manager-todo가 있고 mission-governor는 없다 — 삭제당시 agentdesc 키 집합과 다르다(re-port 근거).

## §B 요구사항 (GEARS)

### B.1 UI 노출

- **REQ-AFR-001** (Ubiquitous): The `moai web` console shall render an agent-settings sub-section inside the llm (3rd Party LLM) panel that lists every parseable agent definition under `.claude/agents/moai/`, each row carrying the agent name, the profile-matrix-resolved model/effort, and a select pair wired as `agentfm.<agent>.model` / `agentfm.<agent>.effort`, with harness-directory rows scanned but not rendered.
- **REQ-AFR-009** (State-driven): **While** a row's resolved model is `haiku`, the console shall render that row's effort select disabled with the `hint.effort.haiku_na` note, and the client wiring shall re-apply the lock on model change and on tier repopulation.
- **REQ-AFR-010** (Capability-gate): **Where** the live llm.yaml marks a GLM backend (`team_mode: glm`), the console shall additionally render each row's z.ai reasoning state, re-porting and reusing `template.ResolveGLMReasoningForModel`(원천 `3fa8bd2ab^`, plan §A.1) — no second derivation — and shall render nothing extra under a Claude backend.
- **REQ-AFR-011** (Ubiquitous): The sub-section's strings shall resolve through 4-locale i18n — new `agentfm.*` / `sec.agentfm.*` / `fieldDesc.agentfm.*` keys plus per-agent `agentdesc.<name>` ko/ja/zh entries whose en baseline is the agent frontmatter description (`data-i18n-baseline`) — and the `agentdesc.` prefix shall re-enter `i18nEnExemptPrefixes` with a recorded justification.
- **REQ-AFR-013** (Ubiquitous): The rendered sub-section shall be test-sliceable through one marker attribute (`data-section="agent-overrides"`), and the llm panel's field count shall count the rendered agent rows, so the 13-tab settings contract (`tab_layout_test.go` `wantTabOrder`, `primary_surface_test.go` `>13<`) stays byte-for-byte untouched.

### B.2 저장 경로

- **REQ-AFR-003** (Event-driven): **When** the operator submits the performance-tier selector (wire field `performance_tier`, closed set {max, medium, low} plus the client-only pseudo-state `custom`), the console shall persist a non-custom selection to `llm.profile`, and shall treat `custom` and an empty submission as preserve.
- **REQ-AFR-004** (Event-driven): **When** the operator submits per-agent edits, the console shall persist to `llm.agent_overrides` exactly the submitted matrix-member agents whose (model, effort) differs from the profile default under the target tier, shall clear submitted overrides that equal the profile default, shall skip non-matrix agents, shall backfill an unsubmitted effort with the resolved value, and shall leave llm.yaml byte-identical when nothing changes.
- **REQ-AFR-005** (Ubiquitous): The console shall not write agent frontmatter — `.claude/agents/**/*.md` is a read-only scan surface, and every save path shall leave it byte-identical (SPEC-MODEL-PROFILE-MATRIX-001 REQ-MPM-040 계약 승계).
- **REQ-AFR-006** (Ubiquitous): Validation closed sets shall be model ∈ {inherit, haiku, sonnet, opus, fable} and effort ∈ {low, medium, high, xhigh, max}, reusing the surviving v4manifest closed-set constants (`inherit`/`haiku`/`sonnet`/`opus` — `schema.go:62-73`) plus the config-level `Fable` constant (`internal/config/types.go:367` — v4manifest에는 fable 상수가 없다); the out-of-set rejection shall join the existing atomic-reject flow.
- **REQ-AFR-007** (Event-detected): **When** any validation or persistence error exists in a save request, the console shall reject atomically — llm.yaml, agent frontmatter, and all section files remain byte-identical — and shall re-render with per-field errors.
- **REQ-AFR-012** (Ubiquitous): The restored save path shall expose injectable seams (list / parse / persist) so tests substitute each step without touching default wiring, matching the console's existing seam convention (`glmcredSave` / `jevcredSave` pattern).

### B.3 설정 스키마

- **REQ-AFR-002** (Ubiquitous): Sub-agent model/effort inheritance shall remain the default behavior — an agent without an `llm.agent_overrides` entry resolves to the session model/effort — and this SPEC shall change no spawn-path behavior.
- **REQ-AFR-008** (Ubiquitous): The shipped template `llm.yaml` shall carry the `llm.profile` and `llm.agent_overrides` keys again with an updated comment block, and `moai update` shall not strip a retired key while the embedded template still ships it (`template.ShippedRetiredModelKeys` seam), so `llm.performance_tier`, `llm.harness_agents`, and the `workflow.yaml` retired keys remain retired-and-stripped.

### B.4 상위 SPEC 수정 관계

- **REQ-AFR-014** (Ubiquitous): The restoration shall be recorded as a narrow amendment of SPEC-AGENT-MODEL-INHERIT-001 — REQ-AMI-011 gains the operator-directed console-surface exception, REQ-AMI-013 narrows to the still-retired key set (`performance_tier`, `harness_agents`, workflow keys) — following the repo's completed-SPEC amendment precedent (SPEC-JEV-CORE-001 v0.2.0/v0.3.0 형식: HISTORY 행, 요구사항 삭제 없음, id 재번호 없음, `status`·`sync_commit_sha` 불변).

REQ 총수 14 ≤ Tier M 상한 16.

## §C 설계 결정 (운영자 지시 5항목의 명시적 처분)

| # | 질문 | 처분 | 요지 |
|---|------|------|------|
| 1 | REQ-AMI-011 처리 | **형식적 좁은 수정안(narrow amendment)** — Q1 운영자 비준 2026-10-02 (decision-index) | 상속은 기본값으로 유지, 콘솔 표면만 예외화. 단면별 처분표는 plan.md §D.1. 정합성 필연으로 REQ-AMI-013(템플릿 키)·REQ-AMI-014(스크립)도 함께 정리한다 — 복원 저장 경로가 쓰는 키를 update가 벗겨내면 표면이 자기모순이 되기 때문(`ShippedRetiredModelKeys` seam 실측: 템플릿 수재 키는 자동 strip 제외). |
| 2 | 같은 커밋이 지운 rosterguard/registry.go·v4manifest 표시 헬퍼 | **OUT** | 블루프린트에서 웹 저장 경로가 v4manifest에서 가져오는 것은 closed-set 상수뿐이고(전부 생존 실측), `ModelColor`/`Tier` 표시 헬퍼는 현재 티어 기계와 중복된다. rosterguard 행은 콘솔 저장 경로와 무관. 단, 복원 파일에 대응하는 rosterguard/numeral 등록은 M5에서 재검토한다(고아 방지). |
| 3 | 착지된 티어 패널과의 공존 | **llm 패널 서브섹션 + 기계 공유** — Q3 운영자 확정 2026-10-02 (decision-index) | `agenttierpanel.go`/`jevkey.go`의 배치 패턴(마커 상수 + `.subsection` 블록 + 커스텀 parse/persist)을 공유한다. 배치는 workflow 패널이 아니라 **llm(3rd Party LLM) 패널** — 저장 대상이 llm.yaml이므로 파일 정합성이 있다. 티어 차트(workflow 패널)는 '왜 티어가 있는가'의 표시 근거, 본 서브섹션은 에이전트별 제어 — 상호 참조하는 두 표면. 탭 14개 복원안은 기각(탭 계약·i18n·chrome 락 전면 재협상 비용). |
| 4 | revert가 아니라 re-port | **채택 (구체 차이 명시)** | 로스터 차이(mission-governor→manager-todo), 매트릭스 셀은 9-27 스냅숏이 아니라 현재 모델 매트릭스(t1368 착지본)에서 재유도, `llm.performance_tier` 레거시 alias 미부활(마지막 소비자 소멸), 프론트매터 쓰기 없음(REQ-MPM-040 승계), 14번째 탭 아님. 포트 선례: t1385/t1388. |
| 5 | 락 테스트 반전 (observational-RED) | **열거 처분표 (표면-결합 집합)** | plan.md §D.5에 표면-결합 테스트 파일 전부(실측 스캔 기준 28파일)를 파일별 invert/replace/keep-and-why 행으로 **명시 열거**한다 — 전수 서술이 아니라 열거이며, 신규 파일은 run-phase 착수 시 재판정한다. 현재 부재 락이 TDD의 RED 기선이다. |

## §D 비기능 제약

- 방법론: TDD(quality.yaml `constitution.development_mode: tdd`) — 각 마일스톤 RED→GREEN→REFACTOR, 최소 커버리지 커밋당 80%.
- 원자성: 저장 실패 시 모든 영속 표면 byte-identical (기존 atomic-reject 흐름 재사용).
- i18n: 신규 사용자 노출 문자열 4-locale(ko/en/ja/zh) 필수, i18n 거버넌스·미번역 허용목록 테스트 통과.
- TRUST 5 전 영역 + LSP 게이트(quality.yaml `lsp_quality_gates`: plan 기준선 필수, run 오류 0).
- 탭 계약: 13탭 유지 — `wantTabOrder`·`>13<` 어설션 무변경 통과.

## §E 관련 SPEC

- 수정 대상: SPEC-AGENT-MODEL-INHERIT-001 (completed → 좁은 수정안, REQ-AFR-014)
- 기계 공유: SPEC-AGENT-TIER-001 (agenttierpanel 배치 패턴), SPEC-WEB-CONSOLE-011 (원래 agentfm 표면의 모 SPEC)
- 셀 데이터 계보: SPEC-MODEL-PROFILE-MATRIX-002 (33셀 매트릭스 원형; superseded — 현재 계보: SPEC-MODEL-MATRIX-UPDATE-001), SPEC-MODEL-MATRIX-UPDATE-001 (t1368 현재 모델 매트릭스)

## Out of Scope

### Out of Scope — 런타임 소비(spawn 경로)

- 복원된 `llm.agent_overrides`를 런처/스폰 경로가 읽어 실제 모델·effort를 고정하는 소비 계약. 현재 트리에 그 소비자가 존재하지 않음(실측: `ResolveAgentModelEffort` 소멸 커밋 `3fa8bd2ab`). 본 SPEC은 콘솔이 읽고·쓰는 저장 표면까지이며, 스폰 시점 강제는 별도 후속 SPEC이다. decision-index Q2 운영자 결정(2026-10-02): 스폰-소비 계약 SPEC의 후속 카드 발행 완료 — 이 공백은 무기한이 아니라 그 후속 카드(전제: 본 SPEC 착지)로 경계 진다.
- `moai model profile` CLI 부활.

### Out of Scope — t1246 커밋의 비(非)콘솔 삭제분

- `internal/harness/rosterguard/registry.go` 행 복원(삭제된 5개 사이트·1 exemption) — 콘솔 저장 경로가 임포트하지 않음(§C 결정 2). 단 복원 파일 자체의 가드 등록은 M5 범위.
- `internal/harness/v4manifest` Tier/ModelColor 표시 헬퍼 복원 — 현재 티어 기계와 중복.
- `internal/web/webux_followup_test.go`(-402)·`webux_haiku_effort_test.go`(-214) 통째 재생성 — 해당 AC는 새 형태의 테스트로 흡수.

### Out of Scope — 프론트매터 쓰기·레거시 alias·문서

- 에이전트 `.md` frontmatter `model:`/`effort:` 쓰기 (REQ-AFR-005 — 읽기 전용 스캔만 복원, `internal/settings/agentfm`의 `Patch` 쓰기 레이어는 부활하지 않는다).
- `llm.performance_tier` 레거시 alias 키와 `llm.harness_agents` — 마지막 소비자가 소멸한 죽은 키를 되살리지 않는다 (단순성 사다리).
- docs-site 페이지(삭제당시 i18n 안내 문구 포함)와 `model-policy.md` 등 규칙 문서 개정 — run-phase 코드 변화 확정 후 sync-phase에서 REQ-AMD-001 계열과 정산.
- `workflow.workflow_agents`·`model_routing*` 키 — Agent Teams 실험 재허용과 무관하게 유지.
